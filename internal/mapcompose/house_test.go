package mapcompose

import (
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// A house hex draws tile/house or one of the numbered variants of its season
// the catalog has; a catalog without variants keeps tile/house.
func TestComposeHouseVariants(t *testing.T) {
	g, at := grid("house house house house", "house house house", "house_w house_w house_w house_w")
	catalog := []string{"tile/house", "tile/house2", "tile/house3", "tile/house5", "tile/house_w", "tile/house2_w"}
	has := func(id string) bool { return slices.Contains(catalog, id) }
	s := Compose(g, at, has)
	seen := map[string]bool{}
	for line := range s.Cells {
		for _, c := range s.Cells[line] {
			want := []string{"tile/house", "tile/house2", "tile/house3"}
			if line == 2 {
				want = []string{"tile/house_w", "tile/house2_w"}
			}
			if !slices.Contains(want, c.Background) {
				t.Errorf("line %d: %s, want one of %v", line, c.Background, want)
			}
			seen[c.Background] = true
		}
	}
	if len(seen) < 4 {
		t.Errorf("the houses use only %v", seen)
	}
	if old := Compose(g, at, none); old.Cells[0][0].Background != "tile/house" || old.Cells[2][0].Background != "tile/house_w" {
		t.Fatalf("a catalog without variants: %s, %s", old.Cells[0][0].Background, old.Cells[2][0].Background)
	}
}

// Neighbouring houses draw different tiles when the catalog has enough.
func TestComposeHousesDiffer(t *testing.T) {
	g, at := grid("house house house house", "house house house", "house house house house", "house house house")
	catalog := []string{"tile/house", "tile/house2", "tile/house3", "tile/house4", "tile/house5", "tile/house6"}
	s := Compose(g, at, func(id string) bool { return slices.Contains(catalog, id) })
	for l := range s.Cells {
		for col, c := range s.Cells[l] {
			for _, n := range g.Neighbors(hexgrid.Coord{Line: l, Col: col}) {
				if s.Cells[n.Line][n.Col].Background == c.Background {
					t.Errorf("%d,%d and its neighbour %v both draw %s", l, col, n, c.Background)
				}
			}
		}
	}
}
