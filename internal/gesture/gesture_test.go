package gesture

import (
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestPlacementTables(t *testing.T) {
	tests := []struct {
		name string
		p    Placement
		want [5]Action // Centre, Up, Left, Right, Down
	}{
		{"map, movement mode", ForMap(false, false), [5]Action{Cancel, Move, ModeChange, Map, Info}},
		{"map, attack mode", ForMap(true, false), [5]Action{Cancel, Attack, ModeChange, Map, Info}},
		{"map, own container", ForMap(false, true), [5]Action{Cancel, Move, Open, Map, Info}},
		{"map, attack mode, own container", ForMap(true, true), [5]Action{Cancel, Attack, Open, Map, Info}},
		{"contents", ForContents(), [5]Action{Cancel, MoveOut, Produce, Cancel, Repair}},
	}
	for _, tt := range tests {
		for d := Centre; d <= Down; d++ {
			a, ok := tt.p.At(d)
			if a != tt.want[d] || !ok {
				t.Errorf("%s: At(%v) = %v, %v; want %v, true", tt.name, d, a, ok, tt.want[d])
			}
		}
	}
}

func TestPlacementRestrict(t *testing.T) {
	p := ForMap(false, false).Restrict(func(a Action) bool { return a != Move && a != Cancel })
	for d, want := range map[Dir]bool{Centre: true, Up: false, Left: true, Right: true, Down: true} {
		if _, ok := p.At(d); ok != want {
			t.Errorf("At(%v) enabled = %v, want %v (cancel always stays enabled)", d, ok, want)
		}
	}
	if a, _ := p.At(Up); a != Move {
		t.Errorf("a disabled slot keeps its action, got %v", a)
	}
	if got, want := p.RingActions(), []Action{Map, Info, ModeChange}; !slices.Equal(got, want) {
		t.Errorf("RingActions = %v, want %v", got, want)
	}
	if got, want := ForMap(true, true).RingActions(), []Action{Attack, Map, Info, Open}; !slices.Equal(got, want) {
		t.Errorf("RingActions = %v, want %v (clockwise from the top, no cancel)", got, want)
	}
	if got, want := ForContents().RingActions(), []Action{MoveOut, Repair, Produce}; !slices.Equal(got, want) {
		t.Errorf("contents RingActions = %v, want %v", got, want)
	}
}

func TestZeroPlacementIsAllCancel(t *testing.T) {
	var p Placement
	for d := Centre; d <= Down; d++ {
		if a, ok := p.At(d); a != Cancel || !ok {
			t.Errorf("zero Placement At(%v) = %v, %v", d, a, ok)
		}
	}
}

func TestNames(t *testing.T) {
	seen := map[string]bool{}
	for a := Cancel; a <= Repair; a++ {
		s := a.String()
		if s == "" || seen[s] {
			t.Errorf("action %d has name %q", a, s)
		}
		seen[s] = true
	}
	if Action(99).String() != "Action(99)" || Dir(9).String() != "Dir(9)" || Kind(42).String() != "Kind(42)" {
		t.Error("out-of-range names")
	}
}

func TestNeighbor(t *testing.T) {
	tests := []struct {
		from hexgrid.Coord
		d    Dir
		want hexgrid.Coord
	}{
		{hexgrid.Coord{Line: 4, Col: 3}, Up, hexgrid.Coord{Line: 2, Col: 3}},
		{hexgrid.Coord{Line: 4, Col: 3}, Down, hexgrid.Coord{Line: 6, Col: 3}},
		{hexgrid.Coord{Line: 4, Col: 3}, Left, hexgrid.Coord{Line: 3, Col: 2}},  // even line: NW
		{hexgrid.Coord{Line: 4, Col: 3}, Right, hexgrid.Coord{Line: 3, Col: 3}}, // even line: NE
		{hexgrid.Coord{Line: 3, Col: 2}, Left, hexgrid.Coord{Line: 4, Col: 2}},  // odd line: SW
		{hexgrid.Coord{Line: 3, Col: 2}, Right, hexgrid.Coord{Line: 4, Col: 3}}, // odd line: SE
		{hexgrid.Coord{Line: 4, Col: 3}, Centre, hexgrid.Coord{Line: 4, Col: 3}},
	}
	for _, tt := range tests {
		if got := Neighbor(tt.from, tt.d); got != tt.want {
			t.Errorf("Neighbor(%v, %v) = %v, want %v", tt.from, tt.d, got, tt.want)
		}
	}
	// Every step is a hex neighbour, opposite steps undo each other, and
	// horizontal steps never drift more than one line from the start.
	opposite := map[Dir]Dir{Up: Down, Down: Up, Left: Right, Right: Left}
	for line := 2; line < 6; line++ {
		c := hexgrid.Coord{Line: line, Col: 5}
		for d, o := range opposite {
			n := Neighbor(c, d)
			if around := hexgrid.Around(c); !slices.Contains(around[:], n) {
				t.Errorf("Neighbor(%v, %v) = %v is not adjacent", c, d, n)
			}
			if back := Neighbor(n, o); back != c {
				t.Errorf("Neighbor(Neighbor(%v, %v), %v) = %v", c, d, o, back)
			}
		}
		walk := c
		for range 6 {
			walk = Neighbor(walk, Right)
			if walk.Line < c.Line-1 || walk.Line > c.Line+1 {
				t.Errorf("walking right from %v left the row at %v", c, walk)
			}
		}
		x0, _ := hexgrid.Center(c)
		if x, _ := hexgrid.Center(walk); x != x0+3*hexgrid.ColPitch {
			t.Errorf("six steps right from %v moved to x %v, want %v", c, x, x0+3*hexgrid.ColPitch)
		}
	}
}
