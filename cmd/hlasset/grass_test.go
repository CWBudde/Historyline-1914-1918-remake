package main

import (
	"strings"
	"testing"
)

func TestTerrainPlanIncludesGrassVariations(t *testing.T) {
	packs, err := planTerrain(terrainCoverage, unitRecipes)
	if err != nil {
		t.Fatal(err)
	}
	counts := map[int]int{}
	for _, p := range packs {
		for _, e := range p.Manifest.Entries {
			if strings.HasPrefix(e.ID, "tile/grass-detail-") {
				if p.Group != "backgrounds" || e.Anchor != [2]float64{48, 40} || e.Width != 96*p.Density || e.Height != 80*p.Density {
					t.Errorf("grass detail has wrong terrain registration: %+v", e)
				}
				counts[p.Density]++
			}
		}
	}
	if counts[2] != 16 || counts[4] != 16 {
		t.Fatalf("terrain catalog has grass details %v, want sixteen per density", counts)
	}
}
