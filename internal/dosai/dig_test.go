package dosai

import (
	"reflect"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// grassLevel is a 21-line level of grass with the given units.
func grassLevel(units map[hexgrid.Coord]string) *level.Level {
	lv := &level.Level{MapName: "grass", Player: "Central Powers"}
	for l := range 21 {
		line := make([]level.Hex, 8-l%2)
		for c := range line {
			line[c] = level.Hex{Terrain: "grass"}
		}
		lv.Lines = append(lv.Lines, line)
	}
	for c, id := range units {
		lv.Lines[c.Line][c.Col].Unit = id
	}
	return lv
}

// digs plans side 1's action phase (side 1 starts in it) and returns its
// work orders.
func digs(t *testing.T, s *dosplay.Session) []dosplay.Event {
	t.Helper()
	if s.Phase(1) != dosrules.PhaseAction {
		t.Fatal("side 1 is not in its action phase")
	}
	evs, err := (&Planner{}).Plan(s, 1)
	if err != nil {
		t.Fatal(err)
	}
	var out []dosplay.Event
	for _, ev := range evs {
		if ev.Kind == dosplay.Tasked {
			out = append(out, ev)
		}
	}
	return out
}

// worked resolves the phase and returns the sappers' work events.
func worked(t *testing.T, s *dosplay.Session) []dosplay.Event {
	t.Helper()
	s.Confirm(0)
	evs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	var out []dosplay.Event
	for _, ev := range evs {
		if ev.Kind == dosplay.Dug {
			out = append(out, ev)
		}
	}
	return out
}

// A sapper near an enemy digs a trench on a neighbour towards it.
func TestSapperDigs(t *testing.T) {
	from, enemy := hexgrid.Coord{Line: 4, Col: 3}, hexgrid.Coord{Line: 12, Col: 3}
	s := session(t, grassLevel(map[hexgrid.Coord]string{from: "EC_sappers", enemy: "CP_infantry"}))
	sapper := s.View(1).Units[from].Index
	tasked := digs(t, s)
	if len(tasked) != 1 || tasked[0].Unit != sapper || tasked[0].From != from {
		t.Fatalf("work orders %+v", tasked)
	}
	to := tasked[0].To
	if hexgrid.Distance(from, to) != 1 || hexgrid.Distance(to, enemy) >= hexgrid.Distance(from, enemy) {
		t.Errorf("the sapper works on %v, not towards the enemy at %v", to, enemy)
	}
	done := worked(t, s)
	if len(done) != 1 || done[0].Text != "Trench dug." || done[0].To != to {
		t.Fatalf("work %+v", done)
	}
	re := s.Reworked()
	if len(re) != len(done[0].Hexes) {
		t.Errorf("reworked %v, dug %v", re, done[0].Hexes)
	}
	for _, c := range done[0].Hexes {
		if _, ok := re[c]; !ok {
			t.Errorf("dug hex %v not reworked", c)
		}
	}
}

// A sapper with no enemy near and no own HQ near gives no order, and one
// with only trenches around it fills none of them.
func TestSapperIdle(t *testing.T) {
	from := hexgrid.Coord{Line: 1, Col: 3}
	if d := hexgrid.Distance(from, hexgrid.Coord{Line: 19, Col: 3}); d <= digRadius {
		t.Fatalf("distance %d", d)
	}
	s := session(t, grassLevel(map[hexgrid.Coord]string{from: "EC_sappers", {Line: 19, Col: 3}: "CP_infantry"}))
	if tasked := digs(t, s); len(tasked) != 0 {
		t.Errorf("a sapper far from the enemy works: %+v", tasked)
	}

	from = hexgrid.Coord{Line: 6, Col: 3}
	s = session(t, grassLevel(map[hexgrid.Coord]string{from: "EC_sappers", {Line: 12, Col: 3}: "CP_infantry"}))
	for _, c := range hexgrid.Around(from) {
		if cell, ok := s.Board.Cell(c); ok {
			for side := range 2 {
				s.Game.Maps[side][2*cell] = 0xA7
			}
		}
	}
	if tasked := digs(t, s); len(tasked) != 0 {
		t.Errorf("a sapper among trenches fills one: %+v", tasked)
	}
}

// Two sappers side by side never fill what the other digs in the same
// phase.
func TestSappersTogether(t *testing.T) {
	a, b, enemy := hexgrid.Coord{Line: 4, Col: 3}, hexgrid.Coord{Line: 4, Col: 4}, hexgrid.Coord{Line: 10, Col: 4}
	s := session(t, grassLevel(map[hexgrid.Coord]string{a: "EC_sappers", b: "EC_sappers", enemy: "CP_infantry"}))
	tasked := digs(t, s)
	if len(tasked) != 2 {
		t.Fatalf("work orders %+v", tasked)
	}
	for _, ev := range worked(t, s) {
		if ev.Text != "Trench dug." {
			t.Errorf("work %+v", ev)
		}
	}
}

// TestPlayOutDigs: on 1914-11 (the computer side's sapper on the map) the
// planners dig trenches and never fill one, the same way every time.
func TestPlayOutDigs(t *testing.T) {
	lv, err := scenario.ShippedLevel("1914-11", scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	_, a := play(t, session(t, lv), [2]*Planner{{}, {}})
	dug := slices.DeleteFunc(slices.Clone(a), func(ev dosplay.Event) bool { return ev.Kind != dosplay.Dug })
	if len(dug) == 0 {
		t.Error("no trench dug")
	}
	for _, ev := range dug {
		if ev.Text != "Trench dug." {
			t.Errorf("side %d: %s at %v", ev.Side, ev.Text, ev.To)
		}
	}
	t.Logf("%d trenches dug", len(dug))
	_, b := play(t, session(t, lv), [2]*Planner{{}, {}})
	if !reflect.DeepEqual(a, b) {
		t.Error("two playouts gave different events")
	}
}
