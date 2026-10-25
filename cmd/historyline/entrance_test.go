package main

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/gesture"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// TestBuildingCellsActAsEntrance checks that every cell of a building does
// what its entrance does on PULSE's opening: a click opens the own HQ's
// contents, the open gesture is offered, and a unit aimed at the entrance
// enters on a click on another cell, which keeps its colour while aimed.
func TestBuildingCellsActAsEntrance(t *testing.T) {
	d := newDosTest(t, "monthly/1914-08:cp", passPlanner{})
	hq := ownHQ(t, d)
	cells := d.s.Footprint(hq)
	if len(cells) < 2 || cells[0] != hq {
		t.Fatalf("HQ footprint %v", cells)
	}
	part := cells[len(cells)-1]
	if e, ok := d.s.Entrance(part); !ok || e != hq {
		t.Fatalf("Entrance(%v) = %v, %v, want %v", part, e, ok, hq)
	}
	if !d.ownContainer(part) || d.can(gesture.Open, part) != d.can(gesture.Open, hq) {
		t.Errorf("open on %v: container %v, can %v, entrance can %v", part, d.ownContainer(part), d.can(gesture.Open, part), d.can(gesture.Open, hq))
	}
	d.Click(part)
	if d.can(gesture.Open, hq) && (d.box == nil || d.box.at != hq) {
		t.Errorf("a click on %v opened %+v, want the HQ at %v", part, d.box, hq)
	}
	d.Clear()

	var unit int
	from := hexgrid.Coord{Line: -1}
	for _, at := range ownUnits(d) {
		d.Click(at)
		if d.selected && d.moves[hq] {
			from, unit = at, d.unit
			break
		}
		d.Clear()
	}
	if from.Line < 0 {
		t.Fatal("no unit reaches the HQ")
	}
	if d.moves[part] || !d.onTarget(part) || !d.reach()[part] {
		t.Fatalf("aimed at the HQ: %v a destination %v, on target %v, in reach %v", part, d.moves[part], d.onTarget(part), d.reach()[part])
	}
	d.Click(part)
	d.SkipWalk()
	if u := d.s.Unit(unit); u.At != hq {
		t.Errorf("the unit from %v is at %v (held %v), want in the HQ at %v", from, u.At, u.Held, hq)
	}
}
