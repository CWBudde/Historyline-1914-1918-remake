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
	"path/filepath"
	"sort"
	"strings"

	runtimeassets "github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/presentation"
)

type assets struct {
	animations []presentation.Animation
	images     map[string]image.Image
	hashes     map[string]string
	family     string
}

// Slugs identify delivered packs; IDs identify compatibility bindings in them.
// This closed list prevents untrusted paths from being interpolated below.
var families = []struct{ slug, id string }{
	{"infantry", "infantry"}, {"elite-infantry", "eliteInfantry"},
	{"cavalry", "cavalry"}, {"sappers", "sappers"},
	{"light-artillery", "lightArtillery"}, {"medium-artillery", "mediumArtillery"},
	{"heavy-artillery", "heavyArtillery"}, {"supply-car", "supplyCar"},
	{"bunker", "bunker"}, {"armored-car", "armoredCar"}, {"charron", "charron"},
	{"patrol-boat", "patrolBoat"}, {"transport-train", "transportTrain"},
	{"armored-train", "armoredTrain"}, {"train-artillery", "trainArtillery"},
}

type fileEntry struct {
	ID      string     `json:"id"`
	Path    string     `json:"path"`
	SHA     string     `json:"sha256"`
	Width   int        `json:"width"`
	Height  int        `json:"height"`
	Density int        `json:"density"`
	Anchor  [2]float64 `json:"logical_anchor"`
}

func loadAssets(repo string) (*assets, error) {
	return loadFamilyAssets(repo, "infantry")
}

func loadFamilyAssets(repo, family string) (*assets, error) {
	unitID := ""
	for _, f := range families {
		if f.slug == family {
			unitID = f.id
			break
		}
	}
	if unitID == "" {
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
	base := "art/exports/interface-v1/"
	b, err := read(base+"report.json", "")
	if err != nil {
		return nil, err
	}
	var report struct {
		Schema  int         `json:"schema_version"`
		Profile string      `json:"profile"`
		Files   []fileEntry `json:"files"`
	}
	if err := json.Unmarshal(b, &report); err != nil {
		return nil, err
	}
	if report.Schema != 1 || report.Profile != "interface-v1" {
		return nil, fmt.Errorf("unsupported interface report")
	}
	entries := make(map[string]fileEntry)
	paths := make(map[string]bool)
	for _, f := range report.Files {
		if !filepath.IsLocal(f.Path) || len(f.SHA) != 64 || paths[f.Path] {
			return nil, fmt.Errorf("invalid or duplicate interface path %q", f.Path)
		}
		paths[f.Path] = true
		if f.Density != 1 && f.Path != "animations.json" && f.Path != "bindings.json" {
			continue
		}
		if _, ok := entries[f.ID]; ok {
			return nil, fmt.Errorf("duplicate asset ID %q", f.ID)
		}
		entries[f.ID] = f
	}
	animationEntry, ok := entries["animations"]
	if !ok || animationEntry.Path != "animations.json" {
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
	if !ok || bindingsEntry.Path != "bindings.json" {
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
	decode := func(id, path, hash string, w, h int) error {
		b, err := read(path, hash)
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
	// Existing unit bundles have no exported-output hash report. Their dimensions
	// are validated and exact bytes are fingerprinted in the smoke-test evidence.
	for _, facing := range bindings.Motion.Facings {
		if err := decode("unit/"+facing, "art/exports/unit-"+family+"-v1/sprites/"+unitID+"/cp/"+facing+"-normal-2x.png", "", 192, 192); err != nil {
			return nil, err
		}
	}
	b, err = read("art/unit-"+family+"-v1.json", "")
	if err != nil {
		return nil, err
	}
	var unit struct {
		Family  string `json:"family"`
		Grounds map[string]map[string]struct {
			Source string `json:"source"`
			SHA    string `json:"source_sha256"`
		} `json:"grounds"`
	}
	if err := json.Unmarshal(b, &unit); err != nil {
		return nil, err
	}
	if unit.Family != unitID {
		return nil, fmt.Errorf("unit family manifest disagrees with selected binding")
	}
	grass, ok := unit.Grounds["grass"]["2x"]
	if !ok || len(grass.SHA) != 64 {
		return nil, fmt.Errorf("missing pinned grass fixture")
	}
	if err := decode("terrain", "art/"+grass.Source, grass.SHA, 192, 160); err != nil {
		return nil, err
	}
	return a, nil
}
