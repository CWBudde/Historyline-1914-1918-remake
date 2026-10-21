package mapcompose

import (
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestStandingItems(t *testing.T) {
	g, at := grid(
		"forest grass mountS",
		"house forest",
	)
	at2 := func(c hexgrid.Coord) Hex {
		h := at(c)
		if c == (hexgrid.Coord{Line: 1, Col: 0}) {
			h.Infra = []string{"bridge25"}
		}
		return h
	}
	have := map[string]bool{"standing/forest": true, "standing/house": true, "standing/bridge25": true, "standing/mountain": true}
	s := Compose(g, at2, func(id string) bool { return have[id] })
	want := []Standing{
		{hexgrid.Coord{Line: 0, Col: 0}, "standing/forest"},
		// A stray massif piece shows the mountain's standing item.
		{hexgrid.Coord{Line: 0, Col: 2}, "standing/mountain"},
		{hexgrid.Coord{Line: 1, Col: 0}, "standing/house"},
		{hexgrid.Coord{Line: 1, Col: 0}, "standing/bridge25"},
		{hexgrid.Coord{Line: 1, Col: 1}, "standing/forest"},
	}
	if got := s.Standing(); !slices.Equal(got, want) {
		t.Errorf("standing = %v\nwant %v", got, want)
	}
	if c := s.Cells[0][1]; c.Standing != nil {
		t.Errorf("grass has standing %v", c.Standing)
	}
}

func TestStandingFits(t *testing.T) {
	for _, c := range []struct {
		w, h   float64
		anchor [2]float64
		ok     bool
	}{
		{96, 80, [2]float64{48, 40}, true},
		{192, 160, [2]float64{96, 80}, true},
		{96, 120, [2]float64{48, 80}, true}, // a crown overhanging only upwards
		{193, 160, [2]float64{96, 80}, false},
		{192, 160, [2]float64{97, 80}, false},
		{96, 130, [2]float64{48, 121}, false},
	} {
		if got := StandingFits(c.w, c.h, c.anchor); got != c.ok {
			t.Errorf("%vx%v anchor %v: %v, want %v", c.w, c.h, c.anchor, got, c.ok)
		}
	}
}
