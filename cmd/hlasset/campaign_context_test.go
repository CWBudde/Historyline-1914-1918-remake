package main

import (
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// Removing a campaign context from the retained recipes must fail even when
// all default terrain IDs still have artwork. Boundary transitions must also have an export.
func TestCampaignContextExportsPresent(t *testing.T) {
	root := filepath.Join("..", "..")
	packs, err := planTerrain(filepath.Join(root, "docs/art/runtime-tile-coverage.csv"), filepath.Join(root, "art/recipes"))
	if err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	for _, p := range packs {
		for _, e := range p.Manifest.Entries {
			have[e.ID] = true
		}
	}
	names, err := fs.Glob(levels.FS, "*.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		b, err := fs.ReadFile(levels.FS, name)
		if err != nil {
			t.Fatal(err)
		}
		l, err := level.Parse(b)
		if err != nil {
			t.Fatal(err)
		}
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
