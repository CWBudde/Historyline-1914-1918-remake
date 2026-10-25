package main

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// Removing a campaign context from the retained index must fail even when
// all default terrain IDs still have artwork. Boundary transitions must also have an export.
func TestCampaignContextExportsPresent(t *testing.T) {
	packs, err := planTerrain(retainedIndex(t))
	if err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	for _, p := range packs {
		for _, e := range p.Manifest.Entries {
			have[e.ID] = true
		}
	}
	shipped, err := scenario.Shipped(scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	for _, sh := range shipped {
		name, l := sh.Name, sh.Level
		s := mapcompose.Compose(l.Grid(), func(c hexgrid.Coord) mapcompose.Hex {
			h := l.Lines[c.Line][c.Col]
			return mapcompose.Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
		}, func(id string) bool { return have[id] })
		missing := 0
		for _, g := range s.Gaps {
			if g.Reason == mapcompose.Missing || g.Reason == mapcompose.Undefined {
				missing++
			}
		}
		if missing != 0 {
			t.Errorf("%s: %d terrain contexts lack exports or rules", name, missing)
		}
	}
}
