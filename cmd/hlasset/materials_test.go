package main

import (
	"fmt"
	"strings"
	"testing"
)

func materialIndex(t *testing.T) []indexEntry {
	t.Helper()
	var es []indexEntry
	for _, e := range retainedIndex(t) {
		if !strings.HasPrefix(e.Pack, "terrain-materials-") {
			es = append(es, e)
		}
	}
	return es
}

func withMaterials(es []indexEntry) []indexEntry {
	for _, d := range []int{2, 4} {
		for _, id := range []string{"material/grass/summer", "material/hills/summer"} {
			es = append(es, indexEntry{ID: id, Pack: fmt.Sprintf("terrain-materials-%dx", d), Density: d,
				Path: fmt.Sprintf("%s-%dx.png", id, d), SHA256: strings.Repeat("0", 64), Width: 568 * d, Height: 400 * d,
				Anchor: &[2]float64{0, 0}, Layer: "ground"})
		}
	}
	return es
}

func TestTerrainPlanMaterialsOptional(t *testing.T) {
	if _, err := planTerrain(materialIndex(t)); err != nil {
		t.Fatalf("legacy index: %v", err)
	}
	packs, err := planTerrain(withMaterials(materialIndex(t)))
	if err != nil {
		t.Fatal(err)
	}
	found := map[int]bool{}
	for _, p := range packs {
		if p.Group != "materials" {
			continue
		}
		found[p.Density] = true
		if len(p.Manifest.Entries) != 2 {
			t.Fatalf("%s has %d entries", p.Manifest.ID, len(p.Manifest.Entries))
		}
		for _, e := range p.Manifest.Entries {
			if e.ID != "material/grass/summer" && e.ID != "material/hills/summer" {
				t.Errorf("unexpected material %s", e.ID)
			}
			if e.Width != 568*p.Density || e.Height != 400*p.Density || e.Anchor != [2]float64{} {
				t.Errorf("bad geometry: %+v", e)
			}
		}
	}
	if len(found) != 2 || !found[2] || !found[4] {
		t.Fatalf("material densities = %v, want 2 and 4", found)
	}
}

func TestTerrainPlanMaterialsRejectsMalformed(t *testing.T) {
	cases := []struct {
		name   string
		damage func([]indexEntry) []indexEntry
	}{
		{"partial both densities", func(es []indexEntry) []indexEntry { n := len(es); return append(es[:n-4], es[n-4], es[n-2]) }},
		{"partial one density", func(es []indexEntry) []indexEntry { return es[:len(es)-1] }},
		{"one density only", func(es []indexEntry) []indexEntry { return es[:len(es)-2] }},
		{"unknown ID", func(es []indexEntry) []indexEntry { es[len(es)-1].ID = "material/water/summer"; return es }},
		{"duplicate ID", func(es []indexEntry) []indexEntry { es[len(es)-1].ID = "material/grass/summer"; return es }},
		{"wrong density", func(es []indexEntry) []indexEntry { es[len(es)-1].Density = 2; return es }},
		{"wrong width", func(es []indexEntry) []indexEntry { es[len(es)-1].Width -= 4; return es }},
		{"wrong height", func(es []indexEntry) []indexEntry { es[len(es)-1].Height -= 4; return es }},
		{"paired wrong geometry", func(es []indexEntry) []indexEntry {
			for i := range es {
				if strings.HasPrefix(es[i].Pack, "terrain-materials-") {
					es[i].Width = 567 * es[i].Density
					es[i].Height = 399 * es[i].Density
				}
			}
			return es
		}},
		{"paired wrong anchor", func(es []indexEntry) []indexEntry {
			for i := range es {
				if strings.HasPrefix(es[i].Pack, "terrain-materials-") {
					es[i].Anchor = &[2]float64{1, 1}
				}
			}
			return es
		}},
		{"wrong anchor", func(es []indexEntry) []indexEntry { es[len(es)-1].Anchor = &[2]float64{1, 0}; return es }},
		{"wrong layer", func(es []indexEntry) []indexEntry { es[len(es)-1].Layer = "standing"; return es }},
		{"unsupported pack density", func(es []indexEntry) []indexEntry { es[len(es)-1].Pack = "terrain-materials-3x"; return es }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := planTerrain(tc.damage(withMaterials(materialIndex(t)))); err == nil {
				t.Fatal("malformed materials accepted")
			}
		})
	}
}
