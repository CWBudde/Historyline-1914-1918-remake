package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	runtimeassets "github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/presentation"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

type assets struct {
	animations []presentation.Animation
	images     map[string]image.Image
	hashes     map[string]string
	family     string
}

// indexEntry is one runtime file of assets/index.json; Path is relative to assets/.
type indexEntry struct {
	ID      string     `json:"id"`
	Pack    string     `json:"pack"`
	Path    string     `json:"path"`
	SHA     string     `json:"sha256"`
	Width   int        `json:"width"`
	Height  int        `json:"height"`
	Density int        `json:"density"`
	Anchor  [2]float64 `json:"logical_anchor"`
	Art     string     `json:"art"`
}

func loadAssets(repo string) (*assets, error) {
	return loadFamilyAssets(repo, "infantry")
}

// loadFamilyAssets reads the interface frames, one unit family and the grass
// tile from the runtime assets of repo, each checked against assets/index.json.
// Family slugs come from the closed unitsprite table, never from paths.
func loadFamilyAssets(repo, family string) (*assets, error) {
	if !slices.Contains(unitsprite.Families(), family) {
		return nil, fmt.Errorf("unknown unit family %q", family)
	}
	r, err := os.OpenRoot(repo)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	a := &assets{images: make(map[string]image.Image), hashes: make(map[string]string), family: family}
	read := func(path, want string) ([]byte, error) {
		if !filepath.IsLocal(path) {
			return nil, fmt.Errorf("nonlocal asset %q", path)
		}
		for _, part := range strings.Split(filepath.ToSlash(path), "/") {
			if strings.EqualFold(part, "legacy") || strings.EqualFold(part, "original") {
				return nil, fmt.Errorf("reference bytes forbidden: %s", path)
			}
		}
		b, err := r.ReadFile(path)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(b)
		actual := hex.EncodeToString(sum[:])
		if want != "" && want != actual {
			return nil, fmt.Errorf("hash mismatch: %s", path)
		}
		a.hashes[filepath.ToSlash(path)] = actual
		return b, nil
	}
	const base = "assets/"
	b, err := read(base+"index.json", "")
	if err != nil {
		return nil, err
	}
	var index struct {
		Schema int          `json:"schema_version"`
		Assets []indexEntry `json:"assets"`
	}
	if err := json.Unmarshal(b, &index); err != nil {
		return nil, err
	}
	if index.Schema != 1 {
		return nil, fmt.Errorf("unsupported asset index")
	}
	// Interface frames at 1x plus the animation metadata, unit sprites and tiles
	// at 2x; the index never lists a path twice for one key.
	entries := make(map[string]indexEntry)
	for _, f := range index.Assets {
		if !filepath.IsLocal(f.Path) || len(f.SHA) != 64 {
			return nil, fmt.Errorf("invalid asset path %q", f.Path)
		}
		var key string
		switch {
		case f.Pack == "interface-v1-1x", f.Art == "interface" && (f.ID == "animations" || f.ID == "bindings"):
			key = f.ID
		case f.Pack == fmt.Sprintf("unit-%s-2x", family), f.Pack == "terrain-backgrounds-2x":
			key = f.ID
		default:
			continue
		}
		if _, ok := entries[key]; ok {
			return nil, fmt.Errorf("duplicate asset ID %q", key)
		}
		entries[key] = f
	}
	animationEntry, ok := entries["animations"]
	if !ok || path.Base(animationEntry.Path) != "animations.json" {
		return nil, fmt.Errorf("missing pinned animations")
	}
	b, err = read(base+animationEntry.Path, animationEntry.SHA)
	if err != nil {
		return nil, err
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(&a.animations); err != nil {
		return nil, err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("animation JSON must contain exactly one value")
	}
	if _, err := presentation.New(a.animations, nil, false); err != nil {
		return nil, err
	}
	bindingsEntry, ok := entries["bindings"]
	if !ok || path.Base(bindingsEntry.Path) != "bindings.json" {
		return nil, fmt.Errorf("missing pinned bindings")
	}
	b, err = read(base+bindingsEntry.Path, bindingsEntry.SHA)
	if err != nil {
		return nil, err
	}
	var bindings struct {
		Motion struct {
			Facings []string `json:"facings"`
		} `json:"motion"`
		Map struct {
			Center [2]int `json:"tile_center"`
			Size   [2]int `json:"tile_size"`
		} `json:"map_registration"`
	}
	if err := json.Unmarshal(b, &bindings); err != nil {
		return nil, err
	}
	if strings.Join(bindings.Motion.Facings, ",") != "n,nw,sw,s,se,ne" || bindings.Map.Center != [2]int{48, 40} || bindings.Map.Size != [2]int{96, 80} {
		return nil, fmt.Errorf("unsupported candidate geometry")
	}
	wanted := map[string]bool{"map/selected": true, "map/path": true}
	for _, anim := range a.animations {
		for _, frame := range append(append([]string{}, anim.Frames...), anim.ReducedFrames...) {
			f, ok := entries[frame]
			if !ok || f.Anchor != [2]float64{float64(anim.Anchor[0]), float64(anim.Anchor[1])} {
				return nil, fmt.Errorf("missing frame or anchor mismatch: %s", frame)
			}
			wanted[frame] = true
		}
	}
	decode := func(id string, f indexEntry, w, h int) error {
		b, err := read(base+f.Path, f.SHA)
		if err != nil {
			return err
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(b))
		if err != nil {
			return err
		}
		if cfg.Width != w || cfg.Height != h || w <= 0 || h <= 0 || w > 2048 || h > 2048 {
			return fmt.Errorf("bad dimensions for %s: %dx%d", id, cfg.Width, cfg.Height)
		}
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			return err
		}
		a.images[id] = im
		return nil
	}
	// The shared runtime loader snapshots and caches the selected interface PNGs.
	// Candidate use is explicit; this diagnostic does not promote the art.
	manifest := runtimeassets.Manifest{Schema: 1, ID: "effects-interface-v1", Status: "candidate"}
	for id := range wanted {
		f, ok := entries[id]
		if !ok {
			return nil, fmt.Errorf("missing asset %s", id)
		}
		manifest.Entries = append(manifest.Entries, runtimeassets.Entry{
			ID: id, Path: base + f.Path, SHA256: f.SHA, Width: f.Width, Height: f.Height, Density: 1, Anchor: f.Anchor,
		})
	}
	sort.Slice(manifest.Entries, func(i, j int) bool { return manifest.Entries[i].ID < manifest.Entries[j].ID })
	pack, err := runtimeassets.Load(r.FS(), manifest, runtimeassets.Options{AllowCandidates: true})
	if err != nil {
		return nil, err
	}
	for _, entry := range manifest.Entries {
		im, err := pack.Image(entry.ID, entry.Density)
		if err != nil {
			return nil, err
		}
		a.images[entry.ID] = im
		a.hashes[entry.Path] = entry.SHA256
	}
	for _, facing := range bindings.Motion.Facings {
		face, ok := unitsprite.ParseFacing(facing)
		if !ok {
			return nil, fmt.Errorf("unknown facing %q", facing)
		}
		id := unitsprite.ID(family, unitsprite.CP, face, unitsprite.Normal)
		f, ok := entries[id]
		if !ok {
			return nil, fmt.Errorf("missing unit sprite %s", id)
		}
		if err := decode("unit/"+facing, f, 192, 192); err != nil {
			return nil, err
		}
	}
	grass, ok := entries["tile/grass"]
	if !ok {
		return nil, fmt.Errorf("missing pinned grass fixture")
	}
	if err := decode("terrain", grass, 192, 160); err != nil {
		return nil, err
	}
	return a, nil
}
