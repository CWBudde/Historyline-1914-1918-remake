package main

import (
	"fmt"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

const unitLogical = 96 // logical width and height of every unit sprite

// unitPack is one family pack at one density.
type unitPack struct {
	Family   string
	Density  int
	Manifest assets.Manifest
}

// parsePackID splits "<prefix><name>-<d>x" into name and density.
func parsePackID(id, prefix string) (string, int, bool) {
	rest, ok := strings.CutPrefix(id, prefix)
	i := strings.LastIndex(rest, "-")
	if !ok || i <= 0 || !strings.HasSuffix(rest, "x") {
		return "", 0, false
	}
	d, err := strconv.Atoi(strings.TrimSuffix(rest[i+1:], "x"))
	if err != nil || (d != 1 && d != 2 && d != 4) {
		return "", 0, false
	}
	return rest[:i], d, true
}

// planUnits returns the unit packs of the index. Every family has every palette ×
// facing × state at each of its densities, and all families share one density set.
func planUnits(entries []indexEntry) ([]unitPack, error) {
	var packs []unitPack
	densities := map[string][]int{}
	for _, m := range indexPacks(entries, "unit-") {
		family, d, ok := parsePackID(m.ID, "unit-")
		if !ok || !slices.Contains(unitsprite.Families(), family) {
			return nil, fmt.Errorf("pack %q: no sprite family of that name", m.ID)
		}
		want := map[string]bool{}
		for _, pal := range unitsprite.Palettes {
			for f := range unitsprite.Facings {
				for _, st := range unitsprite.States {
					want[unitsprite.ID(family, pal, f, st)] = true
				}
			}
		}
		seen := map[string]bool{}
		for _, e := range m.Entries {
			if !want[e.ID] || seen[e.ID] {
				return nil, fmt.Errorf("%s: unknown or duplicate variant %s", m.ID, e.ID)
			}
			seen[e.ID] = true
			if e.Density != d || e.Width != unitLogical*d || e.Height != unitLogical*d {
				return nil, fmt.Errorf("%s: %s is %dx%d at %dx, not %d logical pixels", m.ID, e.ID, e.Width, e.Height, e.Density, unitLogical)
			}
		}
		if len(seen) != len(want) {
			return nil, fmt.Errorf("%s: %d sprite variants, want %d", m.ID, len(seen), len(want))
		}
		densities[family] = append(densities[family], d)
		packs = append(packs, unitPack{Family: family, Density: d, Manifest: m})
	}
	var first []int
	for i, f := range unitsprite.Families() {
		ds := densities[f]
		if len(ds) == 0 {
			return nil, fmt.Errorf("index has no pack for unit family %s", f)
		}
		slices.Sort(ds)
		if i == 0 {
			first = ds
		} else if !slices.Equal(ds, first) {
			return nil, fmt.Errorf("unit family %s has densities %v, others %v", f, ds, first)
		}
	}
	return packs, nil
}

// buildUnits builds the whole unit catalog: out/<d>x/<family>/ holds one pack per
// family and density, so each stays inside the loader's decoded-image budget (all
// 4x units in one pack would not). Every pack is loaded and fully validated before
// out is created; out must be new.
func buildUnits(assetsDir, out string) ([]unitPack, error) {
	entries, err := readIndex(assetsDir)
	if err != nil {
		return nil, err
	}
	packs, err := planUnits(entries)
	if err != nil {
		return nil, err
	}
	manifests := make([]assets.Manifest, len(packs))
	for i, p := range packs {
		manifests[i] = p.Manifest
	}
	written, err := writePacks(assetsDir, out, manifests, func(i int) string {
		return path.Join(fmt.Sprintf("%dx", packs[i].Density), packs[i].Family)
	})
	if err != nil {
		return nil, err
	}
	for i := range packs {
		packs[i].Manifest = written[i]
	}
	return packs, nil
}
