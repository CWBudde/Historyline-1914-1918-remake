package mapcompose

import (
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// A runway is drawn instead of its road, with the aprons of the hangars at its
// SW and NW; a hangar hex draws open ground, its apron and the hangar, each in
// the season of its hex.
func TestAirfieldDrawsRunwayAndHangars(t *testing.T) {
	catalog := map[string]bool{}
	for _, id := range []string{
		"tile/runway36", "tile/runway36_w", "tile/hangar", "tile/hangar_w", "tile/apron-sw", "tile/apron-sw_w",
		"tile/apron-nw", "tile/apron-nw_w", "standing/hangar", "standing/hangar_w",
	} {
		catalog[id] = true
	}
	has := func(id string) bool { return catalog[id] }
	// The runway runs down column 1 of the even lines; the hangar at line 3,
	// column 0 lies SW of the runway hex at line 2 and NW of the one at line 4.
	runwayAt := []hexgrid.Coord{{Line: 2, Col: 1}, {Line: 4, Col: 1}, {Line: 6, Col: 1}}
	hangarAt := hexgrid.Coord{Line: 3, Col: 0}
	for _, season := range []string{"", "_w"} {
		g := hexgrid.Grid{Width: 3, Lines: 8}
		at := func(c hexgrid.Coord) Hex {
			switch {
			case slices.Contains(runwayAt, c):
				return Hex{Terrain: "grass" + season, Infra: []string{"road36", "runway36"}}
			case c == hangarAt:
				return Hex{Terrain: "house" + season, Infra: []string{"hangar"}}
			}
			return Hex{Terrain: "grass" + season}
		}
		s := Compose(g, at, has)
		for c, want := range map[hexgrid.Coord][]string{
			runwayAt[0]: {"tile/runway36" + season, "tile/apron-sw" + season},
			runwayAt[1]: {"tile/runway36" + season, "tile/apron-nw" + season},
			runwayAt[2]: {"tile/runway36" + season},
			hangarAt:    {"tile/hangar" + season},
		} {
			if got := s.Cells[c.Line][c.Col].Overlays; !slices.Equal(got, want) {
				t.Errorf("%q %v: overlays %v, want %v", season, c, got, want)
			}
		}
		cell := s.Cells[hangarAt.Line][hangarAt.Col]
		if cell.Background != "tile/grass"+season || !slices.Equal(cell.Standing, []string{"standing/hangar" + season}) {
			t.Errorf("%q hangar: background %s, standing %v", season, cell.Background, cell.Standing)
		}
	}
}

// Without the airfield art a runway stays its road and a hangar its house.
func TestAirfieldWithoutArtKeepsRoadAndHouse(t *testing.T) {
	g := hexgrid.Grid{Width: 3, Lines: 5}
	at := func(c hexgrid.Coord) Hex {
		switch c {
		case hexgrid.Coord{Line: 2, Col: 1}:
			return Hex{Terrain: "grass", Infra: []string{"road36", "runway36"}}
		case hexgrid.Coord{Line: 3, Col: 0}:
			return Hex{Terrain: "house", Infra: []string{"hangar"}}
		}
		return Hex{Terrain: "grass"}
	}
	s := Compose(g, at, none)
	if got := s.Cells[2][1].Overlays; !slices.Equal(got, []string{"tile/road36", "tile/runway36"}) {
		t.Errorf("runway overlays %v", got)
	}
	if c := s.Cells[3][0]; c.Background != "tile/house" || len(c.Standing) != 0 {
		t.Errorf("hangar cell %+v", c)
	}
}
