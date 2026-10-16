package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

const (
	maxRecipe   = 16 << 20
	unitLogical = 96 // logical width and height of every unit sprite
)

// unitPack is one built family pack at one density.
type unitPack struct {
	Family   string
	Density  int
	Manifest assets.Manifest
	source   string // recipe ID: the export directory under -exports
}

// importUnits reads one unit recipe (art/recipes/unit-*-v1.json) and returns a
// candidate manifest per density, with paths relative to the family's export
// directory. Every palette × facing × state must be present at every density, and
// the recipe's class IDs must match the shared family table.
func importUnits(b []byte) ([]unitPack, error) {
	if len(b) > maxRecipe {
		return nil, fmt.Errorf("unit recipe too large")
	}
	type output struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
		Size   struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"dimensions"`
	}
	var recipe struct {
		ID     string `json:"id"`
		Report *struct {
			Schema   int      `json:"schema_version"`
			ClassIDs []string `json:"class_ids"`
			Frames   []struct {
				Index  int        `json:"facing_index"`
				Label  string     `json:"facing_label"`
				Anchor [2]float64 `json:"logical_anchor"`
			} `json:"frames"`
			Variants []struct {
				Index   int      `json:"facing_index"`
				Label   string   `json:"facing_label"`
				Palette string   `json:"palette"`
				State   string   `json:"state"`
				Outputs []output `json:"outputs"`
			} `json:"variants"`
		} `json:"bundle_report"`
	}
	if err := json.Unmarshal(b, &recipe); err != nil {
		return nil, err
	}
	family, ok := unitsprite.SlugOfRecipe(recipe.ID)
	if !ok || len(unitsprite.ClassIDs(family)) == 0 {
		return nil, fmt.Errorf("recipe %q: no sprite family of that name", recipe.ID)
	}
	rep := recipe.Report
	if rep == nil || rep.Schema != 1 {
		return nil, fmt.Errorf("%s: requires a schema-1 bundle report", recipe.ID)
	}
	classes := slices.Clone(rep.ClassIDs)
	slices.Sort(classes)
	if want := unitsprite.ClassIDs(family); !slices.Equal(classes, want) {
		return nil, fmt.Errorf("%s: class IDs %v, family table has %v", recipe.ID, classes, want)
	}
	anchors := map[unitsprite.Facing][2]float64{}
	for _, f := range rep.Frames {
		face, ok := unitsprite.ParseFacing(f.Label)
		if !ok || int(face) != f.Index {
			return nil, fmt.Errorf("%s: frame facing %q/%d", recipe.ID, f.Label, f.Index)
		}
		anchors[face] = f.Anchor
	}
	byDensity := map[int][]assets.Entry{}
	seen := map[string]bool{}
	for _, v := range rep.Variants {
		face, ok := unitsprite.ParseFacing(v.Label)
		if !ok || int(face) != v.Index {
			return nil, fmt.Errorf("%s: unknown facing %q/%d", recipe.ID, v.Label, v.Index)
		}
		pal := unitsprite.Palette(v.Palette)
		if !slices.Contains(unitsprite.Palettes, pal) {
			return nil, fmt.Errorf("%s: unknown palette %q", recipe.ID, v.Palette)
		}
		st := unitsprite.State(v.State)
		if !slices.Contains(unitsprite.States, st) {
			return nil, fmt.Errorf("%s: unknown state %q", recipe.ID, v.State)
		}
		anchor, ok := anchors[face]
		if !ok {
			return nil, fmt.Errorf("%s: no frame registration for facing %v", recipe.ID, face)
		}
		id := unitsprite.ID(family, pal, face, st)
		if seen[id] {
			return nil, fmt.Errorf("%s: duplicate variant %s", recipe.ID, id)
		}
		seen[id] = true
		for _, o := range v.Outputs {
			w, h := o.Size.Width, o.Size.Height
			d := w / unitLogical
			if w != h || w%unitLogical != 0 || (d != 1 && d != 2 && d != 4) {
				return nil, fmt.Errorf("%s: %s is %dx%d, not %d logical pixels at 1x/2x/4x", recipe.ID, o.Path, w, h, unitLogical)
			}
			if !strings.HasSuffix(o.Path, fmt.Sprintf("-%dx.png", d)) {
				return nil, fmt.Errorf("%s: %s does not name density %dx", recipe.ID, o.Path, d)
			}
			byDensity[d] = append(byDensity[d], assets.Entry{ID: id, Path: o.Path, SHA256: o.SHA256, Width: w, Height: h, Density: d, Anchor: anchor})
		}
	}
	if want := len(unitsprite.Palettes) * int(unitsprite.Facings) * len(unitsprite.States); len(seen) != want {
		return nil, fmt.Errorf("%s: %d sprite variants, want %d", recipe.ID, len(seen), want)
	}
	var packs []unitPack
	for _, d := range []int{1, 2, 4} {
		entries := byDensity[d]
		if len(entries) == 0 {
			continue
		}
		if len(entries) != len(seen) {
			return nil, fmt.Errorf("%s: %d of %d variants have density %dx", recipe.ID, len(entries), len(seen), d)
		}
		packs = append(packs, unitPack{
			Family: family, Density: d, source: recipe.ID,
			Manifest: assets.Manifest{Schema: 1, ID: fmt.Sprintf("unit-%s-%dx", family, d), Status: "candidate", Entries: entries},
		})
	}
	return packs, nil
}

// buildUnits builds the whole unit catalog: out/<d>x/<family>/ holds one pack per
// family and density, so each stays inside the loader's decoded-image budget (all
// 4x units in one pack would not). Every pack is loaded and fully validated before
// out is created; out must be new.
func buildUnits(recipes, exports, out string) ([]unitPack, error) {
	paths, err := filepath.Glob(filepath.Join(recipes, "unit-*-v1.json"))
	if err != nil {
		return nil, err
	}
	var packs []unitPack
	families := map[string]bool{}
	for _, p := range paths {
		b, err := readFileBounded(p, maxRecipe)
		if err != nil {
			return nil, err
		}
		ps, err := importUnits(b)
		if err != nil {
			return nil, err
		}
		if len(ps) > 0 {
			families[ps[0].Family] = true
		}
		packs = append(packs, ps...)
	}
	for _, f := range unitsprite.Families() {
		if !families[f] {
			return nil, fmt.Errorf("no recipe for unit family %s in %s", f, recipes)
		}
	}
	// Validate everything first; packs are dropped after validation so only one
	// decoded pack is held at a time.
	for _, p := range packs {
		if err := withUnitPack(exports, p, func(pack *assets.Pack) error { return pack.Validate() }); err != nil {
			return nil, err
		}
	}
	o, err := newOutput(out)
	if err != nil {
		return nil, err
	}
	defer o.close()
	for i := range packs {
		p := &packs[i]
		dir := path.Join(fmt.Sprintf("%dx", p.Density), p.Family)
		if err := withUnitPack(exports, *p, func(pack *assets.Pack) error {
			m, err := o.writePack(dir, pack, p.Manifest)
			p.Manifest = m
			return err
		}); err != nil {
			return nil, err
		}
	}
	o.done()
	return packs, nil
}

// withUnitPack loads p from its export directory and hands it to use.
func withUnitPack(exports string, p unitPack, use func(*assets.Pack) error) error {
	r, err := os.OpenRoot(filepath.Join(exports, p.source))
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()
	pack, err := assets.Load(r.FS(), p.Manifest, assets.Options{AllowCandidates: true})
	if err != nil {
		return fmt.Errorf("%s: %w", p.Manifest.ID, err)
	}
	if err := use(pack); err != nil {
		return fmt.Errorf("%s: %w", p.Manifest.ID, err)
	}
	return nil
}

func readFileBounded(name string, limit int64) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("%s: larger than %d bytes", name, limit)
	}
	return b, nil
}
