package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// TestDosCaptureDepot walks a foot unit onto PULSE's neutral depot by
// clicks: after the resolution the depot and the unit found in it are the
// player's, the playback reports it, and both panels mark the depot as the
// player's.
func TestDosCaptureDepot(t *testing.T) {
	d := newDosTest(t, "monthly/1914-08-pulse:cp", passPlanner{})
	g := newMapGame(d.s.Level, false)
	g.dos = d
	depot := hexgrid.Coord{Line: 18, Col: 5}
	if b, ok := d.own().Buildings[depot]; !ok || b.Kind != dosplay.Depot || b.Owner == d.s.Faction(d.side) {
		t.Fatalf("no foreign depot at %v: %+v", depot, b)
	}
	before := panelOwners(g, depot)
	if before[0] == d.s.Faction(d.side) || before[0] != before[1] {
		t.Fatalf("the panels mark the depot as %v before the capture", before)
	}
	unit := -1
	for _, at := range ownUnits(d) {
		if u := d.own().Units[at]; u.Name == "Infantry" && (unit < 0 || hexgrid.Distance(at, depot) < hexgrid.Distance(d.s.Unit(unit).At, depot)) {
			unit = u.Index
		}
	}
	if unit < 0 {
		t.Fatal("no infantry")
	}
	for round := 0; round < 12; round++ {
		if d.s.Phase(d.side) == 1 {
			d.Click(d.s.Unit(unit).At)
			if d.selected && len(d.moves) > 0 {
				to := slices.MinFunc(dosplay.Coords(d.moves), func(a, b hexgrid.Coord) int {
					return hexgrid.Distance(a, depot) - hexgrid.Distance(b, depot)
				})
				d.Click(to)
				d.SkipWalk()
			} else {
				d.Clear()
			}
		}
		for _, p := range finish(t, d) {
			if p.ev.Kind != dosplay.Captured {
				continue
			}
			if p.ev.Side != d.side || p.ev.To != depot || !strings.HasPrefix(p.ev.Text, "Depot captured, 1 unit found") {
				t.Fatalf("capture event %+v", p.ev)
			}
			b := d.own().Buildings[depot]
			if b.Owner != d.s.Faction(d.side) || len(b.Units) != 2 {
				t.Fatalf("depot after the capture: %+v", b)
			}
			for _, u := range b.Units {
				if u.Side != d.side {
					t.Errorf("unit %s in the depot is side %d's", u.Name, u.Side)
				}
			}
			d.captured(p.ev)
			if d.Message != p.ev.Text {
				t.Errorf("message %q, want %q", d.Message, p.ev.Text)
			}
			if got, want := panelOwners(g, depot), d.s.Faction(d.side); got != [2]state.Player{want, want} {
				t.Errorf("the panels mark the captured depot as %v, want %v", got, want)
			}
			return
		}
	}
	t.Fatal("the depot was not captured")
}

// panelOwners is the owner each side's panel marks at c.
func panelOwners(g *game, c hexgrid.Coord) (out [2]state.Player) {
	defer func() { g.drawing = false }()
	for side := range out {
		g.drawing, g.drawSide = true, side
		out[side] = g.ownerShown(c)
	}
	return out
}
