package dosplay

import (
	"reflect"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// construct gives construction unit u of side its order on to and resolves
// the phase; it returns the work's event.
func construct(t *testing.T, s *Session, side, u int, to hexgrid.Coord) Event {
	t.Helper()
	passUntilAction(t, s, side)
	from := s.Unit(u).At
	if !s.Targets(side, u)[to] {
		t.Fatalf("%v is no site of the construction unit at %v", to, from)
	}
	evs, err := s.Attack(side, u, to)
	if err != nil || len(evs) != 1 || evs[0].Kind != Tasked || evs[0].Text != "Depot construction ordered." {
		t.Fatalf("the order: %+v, %v", evs, err)
	}
	if len(s.Orders(side)) != 0 {
		t.Errorf("the order entered the attack list: %+v", s.Orders(side))
	}
	s.Confirm(0)
	s.Confirm(1)
	evs, err = s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	var done []Event
	for _, ev := range evs {
		if ev.Kind == Constructed {
			done = append(done, ev)
		}
	}
	if len(done) != 1 || done[0].Unit != u || done[0].To != to || done[0].From != from || len(done[0].Hexes) != 4 {
		t.Fatalf("work events %+v", done)
	}
	return done[0]
}

// A construction unit builds a depot in two halves; the session draws it,
// shows its record as an own building and keeps it over a save.
func TestBuilderDepot(t *testing.T) {
	from, enemy := hexgrid.Coord{Line: 4, Col: 3}, hexgrid.Coord{Line: 9, Col: 3}
	s := session(t, grassLevel(map[hexgrid.Coord]string{from: "CP_construction", enemy: "EC_infantry"}), Options{})
	u := s.View(0).Units[from].Index
	if !s.Builder(u) || s.Sapper(u) {
		t.Fatalf("unit %d: builder %v, sapper %v", u, s.Builder(u), s.Sapper(u))
	}
	passUntilAction(t, s, 0)
	if err := s.CanConstruct(0, u); err != nil {
		t.Fatalf("CanBuild: %v", err)
	}
	sites := Coords(s.Targets(0, u))
	if len(sites) != 10 {
		t.Fatalf("sites %v, want ten", sites)
	}
	to := sites[0]
	lv := s.Level

	ev := construct(t, s, 0, u, to)
	if ev.Text != "Half a depot built." || ev.Hexes[0] != to {
		t.Errorf("first half %+v", ev)
	}
	re := s.Reworked()
	for k, c := range ev.Hexes {
		if h, ok := re[c]; !ok || h.Terrain != "cmp"+string("NESW"[k]) {
			t.Errorf("half depot hex %v: %+v", c, re[c])
		}
	}
	if _, ok := s.View(0).Buildings[to]; ok {
		t.Error("half a depot is a building")
	}

	ev = construct(t, s, 0, u, to)
	if ev.Text != "Depot built." || ev.Hold != dosrules.HoldFlag100 {
		t.Errorf("completion %+v", ev)
	}
	b, ok := s.View(0).Buildings[to]
	if !ok || b.Kind != Depot || b.Owner != s.Faction(0) || b.Hold != dosrules.HoldFlag100 {
		t.Errorf("building %+v, %v", b, ok)
	}
	if fp := s.Footprint(to); len(fp) != 4 || fp[0] != to {
		t.Errorf("footprint %v", fp)
	}
	for _, c := range ev.Hexes {
		if e, ok := s.Entrance(c); !ok || e != to {
			t.Errorf("entrance of %v: %v, %v", c, e, ok)
		}
	}
	if !reflect.DeepEqual(s.Reworked(), re) {
		t.Errorf("the complete depot draws as %v, the half one as %v", s.Reworked(), re)
	}
	if lv.Lines[to.Line][to.Col].Terrain != "grass" {
		t.Error("the work changed the level")
	}

	saved, err := s.Save()
	if err != nil {
		t.Fatal(err)
	}
	r, err := LoadBoard(s.Board, lv, dataSet(t), saved)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r.Reworked(), re) || !reflect.DeepEqual(r.View(0).Buildings[to], b) {
		t.Errorf("after Load: %v, %+v", r.Reworked(), r.View(0).Buildings[to])
	}
	passUntilAction(t, s, 0)
	if err := s.CanConstruct(0, u); len(s.Targets(0, u)) != 0 || err == nil || err.Error() != "The unit has nothing left to build with." {
		t.Errorf("after two charges: targets %v, %v", s.Targets(0, u), err)
	}
}
