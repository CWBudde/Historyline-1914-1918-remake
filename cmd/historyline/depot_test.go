package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// TestDosCaptureDepot walks a foot unit onto PULSE's neutral depot by
// clicks: after the resolution the depot and the unit found in it are the
// player's, and the playback reports it.
func TestDosCaptureDepot(t *testing.T) {
	d := newDosTest(t, "level1_cp", passPlanner{})
	depot := hexgrid.Coord{Line: 18, Col: 5}
	if b, ok := d.own().Buildings[depot]; !ok || b.Kind != dosplay.Depot || b.Owner == d.s.Faction(d.side) {
		t.Fatalf("no foreign depot at %v: %+v", depot, b)
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
			return
		}
	}
	t.Fatal("the depot was not captured")
}
