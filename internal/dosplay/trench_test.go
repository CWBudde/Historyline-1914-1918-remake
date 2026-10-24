package dosplay

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// passUntilAction resolves empty phases until side is in its action phase.
func passUntilAction(t *testing.T, s *Session, side int) {
	t.Helper()
	for s.Phase(side) != 2 {
		s.Confirm(0)
		s.Confirm(1)
		if _, err := s.Resolve(); err != nil {
			t.Fatal(err)
		}
	}
}

// work gives sapper of side its order on to and resolves the phase; it
// returns the work's event.
func work(t *testing.T, s *Session, side, sapper int, to hexgrid.Coord) Event {
	t.Helper()
	passUntilAction(t, s, side)
	from := s.Unit(sapper).At
	if !s.Targets(side, sapper)[to] {
		t.Fatalf("%v is no target of the sapper at %v", to, from)
	}
	evs, err := s.Attack(side, sapper, to)
	if err != nil || len(evs) != 1 || evs[0].Kind != Tasked || evs[0].From != from || evs[0].To != to || evs[0].Text == "" {
		t.Fatalf("the work order: %+v, %v", evs, err)
	}
	if len(s.Orders(side)) != 0 {
		t.Errorf("the work entered the attack list: %+v", s.Orders(side))
	}
	if u := s.View(side).Units[from]; u.Target == nil || *u.Target != to || !u.Moved {
		t.Errorf("the sapper's view: %+v", u)
	}
	s.Confirm(0)
	s.Confirm(1)
	evs, err = s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	var done []Event
	for _, ev := range evs {
		switch ev.Kind {
		case Battle, Skipped:
			if ev.Unit == sapper {
				t.Errorf("the sapper fought: %+v", ev)
			}
		case Dug:
			done = append(done, ev)
		}
	}
	if len(done) != 1 || done[0].Unit != sapper || done[0].To != to || !slices.Contains(done[0].Hexes, to) {
		t.Fatalf("work events %+v", done)
	}
	return done[0]
}

// sapperMap is a session on the first original map with a sapper of a
// side on the map, the sapper and its side.
func sapperMap(t *testing.T) (s *Session, sapper, side int) {
	t.Helper()
	for n := range 73 {
		s, _ = original(t, n, Options{})
		for side := range 2 {
			for _, i := range own(s, side) {
				if s.Sapper(i) && !s.Unit(i).Held {
					return s, i, side
				}
			}
		}
	}
	t.Fatal("no original map has a sapper on the map")
	return nil, 0, 0
}

func TestSapperTrench(t *testing.T) {
	s, sapper, side := sapperMap(t)
	lv := s.Level
	passUntilAction(t, s, side)
	var to hexgrid.Coord
	found := false
	for _, c := range Coords(s.Targets(side, sapper)) {
		if _, taken := s.View(side).Units[c]; !taken && s.Level.Lines[c.Line][c.Col].Terrain == "grass" {
			to, found = c, true
			break
		}
	}
	if !found {
		t.Fatal("no free grass hex next to the sapper")
	}

	ev := work(t, s, side, sapper, to)
	re := s.Reworked()
	if ev.Text != "Trench dug." || len(re) != len(ev.Hexes) {
		t.Errorf("dig %+v, reworked %v", ev, re)
	}
	for _, c := range ev.Hexes {
		h, ok := re[c]
		if !ok || h.Terrain != "grass" || !slices.ContainsFunc(h.Infra.IDs, func(id string) bool { return strings.HasPrefix(id, "trench") }) {
			t.Errorf("dug hex %v: %+v", c, h)
		}
	}
	if id := re[to].Infra.IDs; len(ev.Hexes) == 3 && (len(id) != 1 || len(id[0]) != len("trench")+2) {
		t.Errorf("the middle of the line links to both ends: %v", id)
	}

	// A resumed battle shows the trench.
	saved, err := s.Save()
	if err != nil {
		t.Fatal(err)
	}
	r, err := LoadBoard(s.Board, lv, dataSet(t), saved)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r.Reworked(), re) {
		t.Errorf("after Load: %v, want %v", r.Reworked(), re)
	}

	// Filling it leaves grass.
	ev = work(t, s, side, sapper, to)
	if ev.Text != "Trench filled." {
		t.Errorf("fill %+v", ev)
	}
	for _, c := range ev.Hexes {
		if h := s.Reworked()[c]; h.Terrain != "grass" || len(h.Infra.IDs) != 0 {
			t.Errorf("filled hex %v: %+v", c, h)
		}
	}
	if lv.Lines[to.Line][to.Col].Terrain != "grass" {
		t.Error("the work changed the level")
	}
}
