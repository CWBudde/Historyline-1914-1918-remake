package scenario

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// authoredMonths are the hand-authored months; the others come from the
// original maps (cmd/hlcampaign).
var authoredMonths = []string{"1914-09", "1914-11", "1915-01", "1918-10"}

// authored are the hand-authored months' folders.
func authored(t *testing.T) []string {
	t.Helper()
	var out []string
	for _, m := range authoredMonths {
		out = append(out, filepath.Join(levelsDir, m))
	}
	return out
}

// Roads, paths, rails and streams are continuous: every port of a hex meets
// the opposite port of the same kind on its neighbour, or leaves the map, or
// runs into a building. A path may end on a road's port, as the original maps
// join them, a stream in open water and a harbour yard's road in its
// warehouse.
func TestAuthoredLinesAreContinuous(t *testing.T) {
	for _, path := range authored(t) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			if filepath.Base(path) == "1914-09" {
				t.Skip("Marne predates this check: three path ports and a road port dangle")
			}
			f, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			g := f.Grid()
			at := func(c hexgrid.Coord) []string { return f.Map[c.Line][c.Col].Infra.IDs }
			for l, line := range f.Map {
				for col, h := range line {
					c := hexgrid.Coord{Line: l, Col: col}
					for _, id := range h.Infra.IDs {
						for _, s := range strands(id) {
							for _, p := range s.ports {
								n := hexgrid.Around(c)[p-1]
								if !g.Contains(n) || building(f.Map[n.Line][n.Col].Terrain) {
									continue
								}
								if nh := f.Map[n.Line][n.Col]; s.kind == "stream" && strings.HasPrefix(nh.Terrain, "water") && len(nh.Infra.IDs) == 0 ||
									s.kind == "road" && slices.Contains(nh.Infra.IDs, "harbourN") {
									continue
								}
								want := (p+2)%6 + 1
								found := false
								for _, nid := range at(n) {
									for _, ns := range strands(nid) {
										if joins(s.kind, ns.kind) && slices.Contains(ns.ports, want) {
											found = true
										}
									}
								}
								if !found {
									t.Errorf("%s at line %d col %d: port %d has no %s port %d at line %d col %d",
										id, l+1, col+1, p, s.kind, want, n.Line+1, n.Col+1)
								}
							}
						}
					}
				}
			}
		})
	}
}

// joins reports whether a strand of kind a continues into one of kind b: the
// same kind, or a path and a road.
func joins(a, b string) bool {
	return a == b || a == "path" && b == "road" || a == "road" && b == "path"
}

// strand is one road, path, rail or stream through a hex.
type strand struct {
	kind  string
	ports []int
}

// crossings are the level crossings' strands (art/terrain/cross/prompt.md).
var crossings = map[string][]strand{
	"cross1": {{"path", []int{1, 4}}, {"rail", []int{3, 6}}},
	"cross2": {{"road", []int{1, 4}}, {"rail", []int{3, 6}}},
	"cross3": {{"road", []int{2, 5}}, {"rail", []int{3, 6}}},
	"cross4": {{"path", []int{3, 6}}, {"rail", []int{1, 4}}},
}

// strands are the roads, paths, rails and streams an infrastructure id
// carries: a level crossing carries a rail and a road or path.
func strands(id string) []strand {
	if c, ok := crossings[id]; ok {
		return c
	}
	for _, k := range []string{"road", "path", "rail", "stream"} {
		if rest, found := strings.CutPrefix(id, k); found && rest != "" {
			var ports []int
			for _, r := range rest {
				ports = append(ports, int(r-'0'))
			}
			return []strand{{k, ports}}
		}
	}
	return nil
}

// Every building of a hand-authored month can be reached on foot from every
// other: no stream, building or terrain cuts the map in two.
func TestAuthoredBuildingsConnect(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range authored(t) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			f, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			g := f.Grid()
			walkable := func(c hexgrid.Coord) bool {
				h := f.Map[c.Line][c.Col]
				terr, ok := set.Terrain[h.Terrain]
				if !ok {
					t.Fatalf("terrain %q unknown", h.Terrain)
				}
				// As the rules decide it for units on foot (game.stepCost).
				return data.FixedModifiers(terr, h.Infra.IDs).Movement != nil
			}
			for _, mode := range []Mode{CP, EC} {
				bs := f.Setup(mode).Buildings
				if len(bs) < 2 {
					t.Fatalf("%s: %d buildings", mode, len(bs))
				}
				seen := map[hexgrid.Coord]bool{bs[0].At: true}
				queue := []hexgrid.Coord{bs[0].At}
				for len(queue) > 0 {
					c := queue[0]
					queue = queue[1:]
					for _, n := range g.Neighbors(c) {
						if seen[n] {
							continue
						}
						// Buildings are targets, not thoroughfares.
						switch {
						case building(f.Map[n.Line][n.Col].Terrain):
							seen[n] = true
						case walkable(n):
							seen[n] = true
							queue = append(queue, n)
						}
					}
				}
				for _, b := range bs {
					if !seen[b.At] {
						t.Errorf("%s: building at line %d col %d cannot be reached from line %d col %d",
							mode, b.At.Line+1, b.At.Col+1, bs[0].At.Line+1, bs[0].At.Col+1)
					}
				}
			}
		})
	}
}
