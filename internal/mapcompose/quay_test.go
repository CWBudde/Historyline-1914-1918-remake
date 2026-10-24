package mapcompose

import (
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// A road bridge on water (the levels' stand-in for a quay) draws the quay on
// its axis instead of the bridge and its road; on land it stays a bridge.
func TestRoadBridgeOnWaterDrawsAQuay(t *testing.T) {
	has := func(id string) bool { return id == "tile/quay36" || id == "standing/rBridge36" }
	for _, tt := range []struct {
		terrain  string
		infra    []string
		overlays []string
		standing bool
	}{
		{"water", []string{"rBridge36", "road36"}, []string{"tile/quay36"}, false},
		{"lakeside12", []string{"rBridge36", "road36", "road14"}, []string{"tile/quay36", "tile/road14"}, false},
		{"grass", []string{"stream25", "rBridge36", "road36"}, []string{"tile/stream25", "tile/rBridge36", "tile/road36"}, true},
		{"water", []string{"rBridge25", "road25"}, []string{"tile/rBridge25", "tile/road25"}, false},
	} {
		g, at := grid("grass grass grass", "grass grass", "grass water grass", "grass grass", "grass grass grass")
		c := hexgrid.Coord{Line: 2, Col: 1}
		hex := func(x hexgrid.Coord) Hex {
			if x == c {
				return Hex{Terrain: tt.terrain, Infra: tt.infra}
			}
			return at(x)
		}
		s := Compose(g, hex, has)
		cell := s.Cells[c.Line][c.Col]
		if !slices.Equal(cell.Overlays, tt.overlays) || slices.Contains(cell.Standing, "standing/rBridge36") != tt.standing {
			t.Errorf("%s %v: overlays %v, standing %v", tt.terrain, tt.infra, cell.Overlays, cell.Standing)
		}
	}
}
