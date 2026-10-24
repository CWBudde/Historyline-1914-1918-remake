package dosplay

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// checkShips checks that each ship's head and rear in v point at each other
// on adjacent hexes, and returns the number of ships.
func checkShips(t *testing.T, name string, v View) int {
	t.Helper()
	ships := 0
	for c, u := range v.Units {
		if u.Rear != nil {
			ships++
			r, ok := v.Units[*u.Rear]
			if !ok || r.Head == nil || *r.Head != c || r.Index != u.Index+1 || r.ID != u.ID || u.ID == "" || hexgrid.Distance(c, *u.Rear) != 1 {
				t.Errorf("%s: head %d at %v, rear %+v at %v", name, u.Index, c, r, *u.Rear)
			}
		}
		if u.Head != nil {
			if h, ok := v.Units[*u.Head]; !ok || h.Rear == nil || *h.Rear != c {
				t.Errorf("%s: rear %d at %v has no head at %v", name, u.Index, c, *u.Head)
			}
		}
	}
	return ships
}

// Map 06 holds twelve two-cell ships: every view links their halves, and the
// link follows a ship that moves.
func TestShipHalvesAreLinked(t *testing.T) {
	s, _ := original(t, 6, Options{})
	for _, tt := range []struct {
		name string
		v    View
	}{{"side 0", s.View(0)}, {"side 1", s.View(1)}, {"true view", s.TrueView()}} {
		if n := checkShips(t, tt.name, tt.v); n != 12 {
			t.Errorf("%s: %d ships, want 12", tt.name, n)
		}
	}
	moved := false
	for _, u := range s.View(0).Units {
		if u.Rear == nil || u.Side != 0 {
			continue
		}
		d, err := s.Destinations(0, u.Index)
		if err != nil || len(d) == 0 {
			continue
		}
		evs, err := s.Move(0, u.Index, Coords(d)[0])
		if err != nil {
			t.Fatal(err)
		}
		path, rear := evs[0].Path, evs[0].Rear
		if len(rear) != len(path) || rear[0] != *u.Rear || hexgrid.Distance(rear[0], path[0]) != 1 {
			t.Errorf("ship %d moved along %v, its rear along %v from %v", u.Index, path, rear, *u.Rear)
		}
		v := s.View(0)
		checkShips(t, "after the move", v)
		head := v.Units[path[len(path)-1]]
		if head.Index != u.Index || head.Rear == nil || *head.Rear != path[len(path)-2] {
			t.Errorf("ship %d moved along %v: head %+v", u.Index, path, head)
		}
		moved = true
		break
	}
	if !moved {
		t.Fatal("no ship of side 0 could move")
	}
}
