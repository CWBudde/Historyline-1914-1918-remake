package mapcompose

import (
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// The harbour group: the warehouse (harbourN) with the customs house
// (harbourW) at its SW and the quay yard (harbourS) at its S.
var (
	harbourNAt = hexgrid.Coord{Line: 2, Col: 1}
	harbourWAt = hexgrid.Coord{Line: 3, Col: 0}
	harbourSAt = hexgrid.Coord{Line: 4, Col: 1}
)

func harbourLevel(season string) (hexgrid.Grid, func(hexgrid.Coord) Hex) {
	return hexgrid.Grid{Width: 3, Lines: 7}, func(c hexgrid.Coord) Hex {
		switch c {
		case harbourNAt:
			return Hex{Terrain: "house" + season, Infra: []string{"harbourN"}}
		case harbourWAt:
			return Hex{Terrain: "house" + season, Infra: []string{"harbourW"}}
		case harbourSAt:
			return Hex{Terrain: "grass" + season, Infra: []string{"road36", "harbourS"}}
		}
		return Hex{Terrain: "grass" + season}
	}
}

// Each harbour hex draws the open ground of its season, its paving and its
// building or cargo with the cast shadow, in the season of its hex; the quay
// yard keeps its road beneath the paving.
func TestHarbourDrawsBuildingsAndYard(t *testing.T) {
	catalog := map[string]bool{"tile/road36": true, "tile/road36_w": true}
	for _, piece := range []string{"harbourN", "harbourW", "harbourS"} {
		for _, s := range []string{"", "_w"} {
			catalog[TileID(piece+s)] = true
			catalog[StandingID(piece+s)] = true
			catalog[ShadowID(StandingID(piece+s))] = true
		}
	}
	has := func(id string) bool { return catalog[id] }
	for _, season := range []string{"", "_w"} {
		g, at := harbourLevel(season)
		s := Compose(g, at, has)
		for c, want := range map[hexgrid.Coord]struct {
			overlays []string
			piece    string
		}{
			harbourNAt: {[]string{"tile/harbourN" + season}, "harbourN" + season},
			harbourWAt: {[]string{"tile/harbourW" + season}, "harbourW" + season},
			harbourSAt: {[]string{"tile/road36" + season, "tile/harbourS" + season}, "harbourS" + season},
		} {
			cell := s.Cells[c.Line][c.Col]
			if cell.Background != "tile/grass"+season {
				t.Errorf("%q %v: background %s", season, c, cell.Background)
			}
			if !slices.Equal(cell.Overlays, want.overlays) {
				t.Errorf("%q %v: overlays %v, want %v", season, c, cell.Overlays, want.overlays)
			}
			if standing := StandingID(want.piece); !slices.Equal(cell.Standing, []string{standing}) ||
				!slices.Equal(cell.Shadows, []string{ShadowID(standing)}) {
				t.Errorf("%q %v: standing %v, shadows %v", season, c, cell.Standing, cell.Shadows)
			}
		}
	}
}

// Without the harbour art the buildings stay houses and the yard its road.
func TestHarbourWithoutArtKeepsHousesAndRoad(t *testing.T) {
	g, at := harbourLevel("")
	s := Compose(g, at, func(id string) bool { return id == "tile/house" || id == "tile/road36" })
	for _, c := range []hexgrid.Coord{harbourNAt, harbourWAt} {
		if cell := s.Cells[c.Line][c.Col]; cell.Background != "tile/house" || len(cell.Standing) != 0 {
			t.Errorf("%v: %+v", c, cell)
		}
	}
	if cell := s.Cells[harbourSAt.Line][harbourSAt.Col]; cell.Background != "tile/grass" || len(cell.Standing) != 0 ||
		!slices.Equal(cell.Overlays, []string{"tile/road36", "tile/harbourS"}) {
		t.Errorf("yard: %+v", cell)
	}
}
