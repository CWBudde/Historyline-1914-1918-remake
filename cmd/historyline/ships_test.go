package main

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
	"github.com/hajimehoshi/ebiten/v2"
)

// shipMatch is a dos match on original map 06, whose fleets hold twelve
// two-cell ships.
func shipMatch(t *testing.T) *game {
	t.Helper()
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	lvl, b, err := loadMap(t.TempDir(), "original/06", set)
	if err != nil {
		t.Fatal(err)
	}
	player, err := state.PlayerByName(lvl.Player)
	if err != nil {
		t.Fatal(err)
	}
	gm, err := newMatch(matchConfig{lvl: lvl, board: b, set: set, player: player, rules: "dos", opponent: "pass"}, nil)
	if err != nil || gm.dos == nil {
		t.Fatalf("dos match: %v", err)
	}
	return gm
}

// A ship faces from its rear toward its head; at the start side 0's rears
// lie north of their heads and side 1's south.
func TestShipFacing(t *testing.T) {
	gm := shipMatch(t)
	ships := 0
	for at, u := range gm.dos.s.TrueView().Units {
		head, rear, ok := shipHalves(u, at)
		if !ok || u.Rear == nil {
			continue
		}
		ships++
		hx, hy := hexgrid.Center(head)
		rx, ry := hexgrid.Center(rear)
		want := map[int]unitsprite.Facing{0: unitsprite.S, 1: unitsprite.N}[u.Side]
		if f := shipFacing(hx, hy, rx, ry); f != want {
			t.Errorf("side %d ship at %v, rear %v: facing %v, want %v", u.Side, head, rear, f, want)
		}
	}
	if ships != 12 {
		t.Errorf("%d ships, want 12", ships)
	}
}

// Each ship is drawn from exactly one of its hexes, the lower one while it is
// in sight.
func TestShipDrawnOnce(t *testing.T) {
	gm := shipMatch(t)
	dst := ebiten.NewImage(int(gm.cam.View.W), int(gm.cam.View.H))
	defer dst.Deallocate()
	for at, u := range gm.dos.s.TrueView().Units {
		head, rear, ok := shipHalves(u, at)
		if !ok || u.Rear == nil {
			continue
		}
		hx, hy := hexgrid.Center(head)
		gm.cam.CenterOn(hx, hy)
		from := []bool{gm.shipDrawnFrom(dst, head, head, rear), gm.shipDrawnFrom(dst, rear, head, rear)}
		lower := rear.Line > head.Line
		if from[0] == from[1] || from[1] != lower {
			t.Errorf("ship at %v, rear %v: drawn from head %v, from rear %v", head, rear, from[0], from[1])
		}
		if gm.shipDrawnFrom(dst, hexgrid.Coord{Line: head.Line, Col: head.Col + 1}, head, rear) {
			t.Errorf("ship at %v drawn from a neighbour", head)
		}
	}
}

// Selecting either half selects the ship; its move walks the rear one step
// behind, and the walking ship hides both its hexes.
func TestShipSelectAndWalk(t *testing.T) {
	gm := shipMatch(t)
	d := gm.dos
	for at, u := range d.own().Units {
		if u.Head == nil || u.Side != d.side {
			continue
		}
		head := *u.Head
		d.selectAt(at)
		if !d.selected || d.unit != u.Index-1 || d.from != head {
			t.Fatalf("selecting the rear on %v: selected %v, unit %d at %v; want %d at %v (%s)", at, d.selected, d.unit, d.from, u.Index-1, head, d.Message)
		}
		if len(d.moves) == 0 {
			continue
		}
		to := dosplay.Coords(d.moves)[0]
		evs, err := d.s.Move(d.side, d.unit, to)
		if err != nil || len(evs) == 0 || evs[0].Kind != dosplay.Moved {
			t.Fatalf("move to %v: %+v, %v", to, evs, err)
		}
		d.reduced = false
		d.walkAlong(evs[0])
		w := d.walk
		if w == nil || len(w.tail) != len(w.path) || w.tail[0] != at {
			t.Fatalf("walk %+v from the rear at %v", w, at)
		}
		if x, y := w.TailPos(); [2]float64{x, y} != [2]float64(hexCentre(at)) {
			t.Errorf("the rear starts at %.0f,%.0f, not on %v", x, y, at)
		}
		for _, c := range []hexgrid.Coord{w.path[len(w.path)-1], w.tail[len(w.tail)-1]} {
			if !gm.dosWalking(c) {
				t.Errorf("the walking ship shows on %v", c)
			}
		}
		return
	}
	t.Fatal("no ship of the human side could move")
}

func hexCentre(c hexgrid.Coord) [2]float64 {
	x, y := hexgrid.Center(c)
	return [2]float64{x, y}
}

// A ship is one unit: its head carries the markers, its rear none.
func TestShipMarkedOnce(t *testing.T) {
	gm := shipMatch(t)
	units := gm.dos.unitsOf(gm.shownSide(), gm.reveal)
	var hexes []hexgrid.Coord
	for c := range units {
		hexes = append(hexes, c)
	}
	marked := map[hexgrid.Coord]bool{}
	for _, m := range gm.unitMarks(hexes) {
		marked[m.at] = true
	}
	if len(marked) == 0 {
		t.Fatal("no markers")
	}
	ships := 0
	for c, u := range units {
		if u.Rear != nil {
			ships++
			if !marked[c] || marked[*u.Rear] {
				t.Errorf("ship at %v: head marked %v, rear at %v marked %v", c, marked[c], *u.Rear, marked[*u.Rear])
			}
		}
	}
	if ships == 0 {
		t.Error("no ships in the shown view")
	}
}
