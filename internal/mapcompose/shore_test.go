package mapcompose

import (
	"fmt"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestEveryShoreLandEdgeSupportsAnOutlet(t *testing.T) {
	// Hand-checked fully land edges of all nine corner profiles. The fourth
	// land corner of lakeside23 supplies three possible mouths, not two.
	for _, tt := range []struct {
		shore string
		ports []int
	}{
		{"lakeside12", []int{1, 2}},
		{"lakeside16", []int{1, 6}},
		{"lakeside23", []int{2, 3, 4}},
		{"lakeside34", []int{3, 4}},
		{"lakeside56", []int{5, 6}},
		{"lakeside123", []int{1, 2, 3}},
		{"lakeside126", []int{1, 2, 6}},
		{"lakeside156", []int{1, 5, 6}},
		{"lakeside3456", []int{3, 4, 5, 6}},
	} {
		for _, port := range tt.ports {
			want := fmt.Sprintf("tile/%s-outlet%d", tt.shore, port)
			t.Run(want, func(t *testing.T) {
				g := hexgrid.Grid{Width: 5, Lines: 7}
				shore := hexgrid.Coord{Line: 3, Col: 2}
				neighbor := hexgrid.Around(shore)[port-1]
				facingStreams := [6]string{"stream14", "stream25", "stream36", "stream14", "stream25", "stream36"}
				at := func(c hexgrid.Coord) Hex {
					if c == shore {
						return Hex{Terrain: tt.shore}
					}
					if c == neighbor {
						return Hex{Terrain: "grass", Infra: []string{facingStreams[port-1]}}
					}
					return Hex{Terrain: "grass"}
				}
				s := Compose(g, at, func(id string) bool { return id == want })
				if got := s.Cells[shore.Line][shore.Col].Background; got != want {
					t.Errorf("background %s, want %s", got, want)
				}
			})
		}
	}
}

func TestShoreOutletSelection(t *testing.T) {
	for _, tt := range []struct {
		name      string
		at        hexgrid.Coord
		shore     string
		port      int
		infra     string
		asset     string
		available bool
		want      string
	}{
		{"PULSE shoreline", hexgrid.Coord{Line: 2, Col: 2}, "lakeside23", 4, "stream14", "tile/lakeside23-outlet4", true, "tile/lakeside23-outlet4"},
		{"same tile on odd row", hexgrid.Coord{Line: 3, Col: 1}, "lakeside23", 4, "stream14", "tile/lakeside23-outlet4", true, "tile/lakeside23-outlet4"},
		{"different shoreline and direction", hexgrid.Coord{Line: 3, Col: 2}, "lakeside16", 1, "stream14", "tile/lakeside16-outlet1", true, "tile/lakeside16-outlet1"},
		{"broad shoreline", hexgrid.Coord{Line: 2, Col: 1}, "lakeside3456", 4, "stream14", "tile/lakeside3456-outlet4", true, "tile/lakeside3456-outlet4"},
		{"stream points elsewhere", hexgrid.Coord{Line: 2, Col: 2}, "lakeside23", 4, "stream25", "tile/lakeside23-outlet4", true, "tile/lakeside23"},
		{"road is not a stream", hexgrid.Coord{Line: 2, Col: 2}, "lakeside23", 4, "road14", "tile/lakeside23-outlet4", true, "tile/lakeside23"},
		{"water edge needs no mouth", hexgrid.Coord{Line: 2, Col: 2}, "lakeside23", 6, "stream36", "tile/lakeside23-outlet6", true, "tile/lakeside23"},
		{"mixed edge is not a land outlet", hexgrid.Coord{Line: 2, Col: 2}, "lakeside23", 1, "stream14", "tile/lakeside23-outlet1", true, "tile/lakeside23"},
		{"missing export keeps base", hexgrid.Coord{Line: 2, Col: 2}, "lakeside23", 4, "stream14", "tile/lakeside23-outlet4", false, "tile/lakeside23"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := hexgrid.Grid{Width: 5, Lines: 7}
			neighbor := hexgrid.Around(tt.at)[tt.port-1]
			at := func(c hexgrid.Coord) Hex {
				if c == tt.at {
					return Hex{Terrain: tt.shore}
				}
				if c == neighbor {
					return Hex{Terrain: "grass", Infra: []string{tt.infra}}
				}
				return Hex{Terrain: "grass"}
			}
			s := Compose(g, at, func(id string) bool { return tt.available && id == tt.asset })
			if got := s.Cells[tt.at.Line][tt.at.Col].Background; got != tt.want {
				t.Errorf("background %s, want %s", got, tt.want)
			}
		})
	}
}

func TestShoreWithMultipleFacingStreamsKeepsBase(t *testing.T) {
	g := hexgrid.Grid{Width: 5, Lines: 7}
	shore := hexgrid.Coord{Line: 2, Col: 2}
	neighbors := hexgrid.Around(shore)
	at := func(c hexgrid.Coord) Hex {
		switch c {
		case shore:
			return Hex{Terrain: "lakeside23"}
		case neighbors[1]: // SE, facing NW port 5.
			return Hex{Terrain: "grass", Infra: []string{"stream25"}}
		case neighbors[3]: // SW, facing NE port 1.
			return Hex{Terrain: "grass", Infra: []string{"stream14", "bridge25"}}
		default:
			return Hex{Terrain: "grass"}
		}
	}
	s := Compose(g, at, func(id string) bool {
		return id == "tile/lakeside23-outlet2" || id == "tile/lakeside23-outlet4"
	})
	if got := s.Cells[shore.Line][shore.Col].Background; got != "tile/lakeside23" {
		t.Errorf("multiple mouths silently selected %s", got)
	}
}
