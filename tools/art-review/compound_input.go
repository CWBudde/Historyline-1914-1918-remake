package artreview

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type compoundSceneInput struct {
	ID      string             `json:"id"`
	Artwork connectionMaterial `json:"artwork"`
	Inset   float64            `json:"inset"`
	Framing string             `json:"framing,omitempty"`
	// Shift moves reference-layout artwork by world pixels after it is scaled
	// to fill the scene box, so its entry street lies on the entry port.
	Shift []float64 `json:"shift,omitempty"`
	// Zoom enlarges filled artwork beyond the box (1..1.1), which gives a
	// shift more slack.
	Zoom float64 `json:"zoom,omitempty"`
	// EntryRoad pins the connected-road manifest whose straight road through
	// the entry port the entry street blends into at the gate.
	EntryRoad *connectionMaterial `json:"entry_road,omitempty"`
	// EdgeFeather (world px; HQ, depot, ruined depot and factory with an entry
	// road) fades the artwork out along an irregular line towards the outer
	// boundary, beside the entry corridor.
	EdgeFeather float64 `json:"edge_feather,omitempty"`
	// EdgeKeep lists polygons (world px of the scene box) inside which the edge
	// feather keeps the artwork exactly: walls, towers, fences and buildings
	// that reach the boundary keep their straight edge.
	EdgeKeep [][]connectionPoint `json:"edge_keep,omitempty"`
}

// fit is how the scene's artwork maps onto its box: reference-layout scenes
// (whose streets and fences reach the box edge) fill it.
func (s compoundSceneInput) fit() compoundFit {
	f := compoundFit{Fill: s.Framing == "reference-layout", Zoom: s.Zoom}
	if len(s.Shift) == 2 {
		f.Shift = connectionPoint{s.Shift[0], s.Shift[1]}
	}
	return f
}

type compoundManifest struct {
	WaterProfile  string                        `json:"water_profile,omitempty"`
	SchemaVersion int                           `json:"schema_version"`
	Grounds       map[string]connectionMaterial `json:"grounds"`
	Scenes        []compoundSceneInput          `json:"scenes"`
}
type compoundSource struct {
	Ref   connectionMaterial
	Bytes []byte
	Image *image.NRGBA
}
type compoundInputs struct {
	Manifest compoundManifest
	Bytes    []byte
	Hash     string
	Grounds  map[string]compoundSource
	Artwork  map[string]compoundSource
	// EntryRoads holds the pinned connected-road inputs of scenes with entry_road.
	EntryRoads map[string]connectionInputs
}

// encoding/json accepts case aliases. Inspect exact schema keys independently,
// after the shared duplicate-key/exactly-one-document validation.
func decodeCompoundManifest(b []byte, m *compoundManifest) error {
	if err := decodeConnectionJSON(b, m); err != nil {
		return err
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(b, &root); err != nil {
		return err
	}
	keys := func(object map[string]json.RawMessage, allowed ...string) error {
		for k := range object {
			ok := false
			for _, a := range allowed {
				if k == a {
					ok = true
				}
			}
			if !ok {
				return fmt.Errorf("noncanonical compound field %q", k)
			}
		}
		return nil
	}
	if root == nil {
		return errors.New("compound manifest requires an object")
	}
	if err := keys(root, "schema_version", "grounds", "scenes", "water_profile"); err != nil {
		return err
	}
	var grounds map[string]json.RawMessage
	if err := json.Unmarshal(root["grounds"], &grounds); err != nil {
		return err
	}
	var scenes []map[string]json.RawMessage
	if err := json.Unmarshal(root["scenes"], &scenes); err != nil {
		return err
	}
	checkRef := func(b json.RawMessage) error {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(b, &obj); err != nil {
			return err
		}
		if len(obj) != 2 {
			return errors.New("source requires source and source_sha256")
		}
		return keys(obj, "source", "source_sha256")
	}
	for season, b := range grounds {
		if season != "summer" && season != "winter" && season != "shared" {
			return fmt.Errorf("unknown ground season %q", season)
		}
		if err := checkRef(b); err != nil {
			return err
		}
	}
	for _, s := range scenes {
		if err := keys(s, "id", "artwork", "inset", "framing", "shift", "zoom", "entry_road", "edge_feather", "edge_keep"); err != nil {
			return err
		}
		if road, ok := s["entry_road"]; ok {
			if err := checkRef(road); err != nil {
				return err
			}
		}
		if field, ok := s["framing"]; ok && bytes.Equal(bytes.TrimSpace(field), []byte("null")) {
			return errors.New("compound framing must be a string")
		}
		if err := checkRef(s["artwork"]); err != nil {
			return err
		}
	}
	return nil
}

func compoundForbidden(path string) bool {
	for _, s := range strings.Split(filepath.ToSlash(path), "/") {
		s = strings.ToLower(s)
		if s == "original" || s == "pilot" {
			return true
		}
	}
	return false
}

func compoundSafeRelative(path string) error {
	if path == "" || filepath.IsAbs(path) || filepath.VolumeName(path) != "" || strings.Contains(path, "\\") || filepath.ToSlash(filepath.Clean(path)) != path || compoundForbidden(path) {
		return errors.New("compound sources require safe canonical relative paths excluding original/pilot")
	}
	for _, s := range strings.Split(path, "/") {
		if s == ".." || s == "." || s == "" {
			return errors.New("compound source traversal is excluded")
		}
	}
	return nil
}

func compoundReadSource(base string, ref connectionMaterial, ground bool) (compoundSource, error) {
	out := compoundSource{Ref: ref}
	if err := compoundSafeRelative(ref.Source); err != nil {
		return out, err
	}
	ref, _, err := seasonRef(base, ref)
	if err != nil {
		return out, err
	}
	out.Ref = ref
	want, err := hex.DecodeString(ref.SourceSHA256)
	if err != nil || len(want) != 32 || strings.ToLower(ref.SourceSHA256) != ref.SourceSHA256 {
		return out, errors.New("compound source requires canonical lowercase SHA256")
	}
	lexical := filepath.Join(base, ref.Source)
	resolved, err := filepath.EvalSymlinks(lexical)
	if err != nil {
		return out, err
	}
	absolute, err := filepath.Abs(resolved)
	if err != nil {
		return out, err
	}
	if compoundForbidden(lexical) || compoundForbidden(absolute) {
		return out, errors.New("resolved original/pilot source excluded")
	}
	baseAbs, err := filepath.Abs(base)
	if err != nil {
		return out, err
	}
	baseResolved, err := filepath.EvalSymlinks(baseAbs)
	if err != nil {
		return out, err
	}
	rel, err := filepath.Rel(baseResolved, absolute)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return out, errors.New("compound symlink escapes the source root")
	}
	out.Bytes, err = os.ReadFile(absolute)
	if err != nil {
		return out, err
	}
	sum := sha256.Sum256(out.Bytes)
	if !bytes.Equal(want, sum[:]) {
		return out, errors.New("compound source SHA256 mismatch")
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(out.Bytes))
	if err != nil {
		return out, err
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > maxInputPixels {
		return out, errors.New("compound PNG dimensions invalid or exceed 64 megapixels")
	}
	if ground && (cfg.Width < 384 || cfg.Height < 320) {
		return out, errors.New("compound ground requires at least 384x320")
	}
	if !ground && (len(out.Bytes) < 26 || out.Bytes[25] != 6) {
		return out, errors.New("compound artwork must be RGBA PNG")
	}
	decoded, err := png.Decode(bytes.NewReader(out.Bytes))
	if err != nil {
		return out, err
	}
	out.Image = image.NewNRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
	draw.Draw(out.Image, out.Image.Bounds(), decoded, decoded.Bounds().Min, draw.Src)
	opaque, transparent := 0, 0
	for y := range cfg.Height {
		for x := range cfg.Width {
			a := out.Image.NRGBAAt(x, y).A
			if ground && a != 255 {
				return out, errors.New("compound ground must be opaque")
			}
			if a >= 250 {
				opaque++
			}
			if a == 0 {
				transparent++
			}
		}
	}
	if !ground && (opaque < 64 || transparent == 0) {
		return out, errors.New("compound artwork requires substantial opaque content (64 pixels) and transparent pixels")
	}
	return out, nil
}

func loadCompoundInputs(path string) (compoundInputs, error) {
	in := compoundInputs{Grounds: map[string]compoundSource{}, Artwork: map[string]compoundSource{}, EntryRoads: map[string]connectionInputs{}}
	var err error
	in.Bytes, err = os.ReadFile(path)
	if err != nil {
		return in, err
	}
	if err = decodeCompoundManifest(in.Bytes, &in.Manifest); err != nil {
		return in, err
	}
	m := in.Manifest
	if m.SchemaVersion != 1 || len(m.Scenes) < 1 || len(m.Scenes) > 12 {
		return in, errors.New("compound requires schema_version 1 and one to twelve scenes")
	}
	if m.WaterProfile != "" && (m.WaterProfile != authoredBankProfile || len(m.Scenes) != 1 || m.Scenes[0].ID != "water-stones") {
		return in, errors.New("water_profile requires authored-rocks-v1 and water-stones only")
	}
	used := map[string]bool{}
	seen := map[string]bool{}
	for _, s := range m.Scenes {
		l, e := compoundLayoutFor(s.ID)
		if e != nil {
			return in, e
		}
		if e = compoundValidateFraming(l, s.Framing); e != nil {
			return in, fmt.Errorf("%s: %w", s.ID, e)
		}
		if seen[s.ID] {
			return in, errors.New("duplicate compound scene")
		}
		seen[s.ID] = true
		used[l.Season] = true
		if math.IsNaN(s.Inset) || math.IsInf(s.Inset, 0) || s.Inset < .5 || s.Inset > 1 {
			return in, errors.New("compound inset must be finite in [0.5,1]")
		}
		if s.Shift != nil {
			if s.Framing != "reference-layout" || len(s.Shift) != 2 || !(math.Abs(s.Shift[0]) <= 8) || !(math.Abs(s.Shift[1]) <= 8) {
				return in, fmt.Errorf("%s: shift requires reference-layout framing and two world-pixel offsets within ±8", s.ID)
			}
		}
		if s.Zoom != 0 && (s.Framing != "reference-layout" || !(s.Zoom >= 1 && s.Zoom <= 1.1)) {
			return in, fmt.Errorf("%s: zoom requires reference-layout framing and lies in [1,1.1]", s.ID)
		}
		if s.EntryRoad != nil {
			if s.Framing != "reference-layout" || l.EntryCell == "" {
				return in, fmt.Errorf("%s: entry_road requires reference-layout framing and an entry cell", s.ID)
			}
			if s.EntryRoad.Source == "" || filepath.IsAbs(s.EntryRoad.Source) || compoundForbidden(s.EntryRoad.Source) {
				return in, fmt.Errorf("%s: entry_road source must be relative", s.ID)
			}
			road, e := loadPinnedMixedConnection(filepath.Join(sourceBase(path), s.EntryRoad.Source), s.EntryRoad.SourceSHA256)
			if e != nil {
				return in, fmt.Errorf("%s: entry_road: %w", s.ID, e)
			}
			if _, _, e := compoundEntryRoad(road, l); e != nil {
				return in, fmt.Errorf("%s: entry_road: %w", s.ID, e)
			}
			in.EntryRoads[s.ID] = road
		}
		if e = compoundValidateFeather(l, s.Framing, s.EdgeFeather, s.EntryRoad != nil, s.EdgeKeep); e != nil {
			return in, fmt.Errorf("%s: %w", s.ID, e)
		}
		src, e := compoundReadSource(sourceBase(path), s.Artwork, false)
		if e != nil {
			return in, fmt.Errorf("%s: %w", s.ID, e)
		}
		in.Artwork[s.ID] = src
	}
	if len(used) != len(m.Grounds) {
		return in, errors.New("grounds must match exactly the used seasons")
	}
	for _, season := range []string{"summer", "winter", "shared"} {
		if !used[season] {
			continue
		}
		ref, ok := m.Grounds[season]
		if !ok {
			return in, errors.New("missing seasonal ground")
		}
		src, e := compoundReadSource(sourceBase(path), ref, true)
		if e != nil {
			return in, e
		}
		in.Grounds[season] = src
	}
	sum := sha256.Sum256(in.Bytes)
	in.Hash = hex.EncodeToString(sum[:])
	return in, nil
}
