package main

import (
	"testing"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosai"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// tickUntil ticks d until done holds, giving the planning goroutine time.
func tickUntil(t *testing.T, d *dosControl, done func() bool) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for !done() {
		if time.Now().After(deadline) {
			t.Fatal("timed out")
		}
		d.Tick()
		time.Sleep(time.Millisecond)
	}
}

func TestPlanOffTheFrame(t *testing.T) {
	d := newDosTest(t, "level1_cp", computerPlanner())
	same := newDosTest(t, "level1_cp", computerPlanner())
	opp := 1 - d.side
	for turn := range 4 {
		// The planning shows for planMinFrames at least with a rising bar,
		// while the human side may still give orders.
		d.Tick()
		if !d.planning() || d.s.Ready(opp) {
			t.Fatalf("turn %d: planning %v, ready %v after the first frame", turn, d.planning(), d.s.Ready(opp))
		}
		frames, last := 1, 0.0
		tickUntil(t, d, func() bool {
			if p := d.planProgress(); d.planning() {
				if p < last {
					t.Fatalf("progress fell from %v to %v", last, p)
				}
				last = p
				frames++
			}
			return d.s.Ready(opp)
		})
		if frames < planMinFrames || last != 1 {
			t.Errorf("turn %d: planning shown %d frames, last progress %v", turn, frames, last)
		}
		// It gives the same orders as planning in the frame.
		if _, err := (&dosai.Planner{}).Plan(same.s, opp); err != nil {
			t.Fatal(err)
		}
		a, err := d.s.Save()
		if err != nil {
			t.Fatal(err)
		}
		b, err := same.s.Save()
		if err != nil {
			t.Fatal(err)
		}
		if string(a) != string(b) {
			t.Fatalf("turn %d: planning off the frame differs from planning in it", turn)
		}
		// Both confirm; the resolution plays and the next phase begins.
		d.s.Confirm(d.side)
		same.s.Confirm(d.side)
		if _, err := same.s.Resolve(); err != nil {
			t.Fatal(err)
		}
		d.SetReducedMotion(true)
		tickUntil(t, d, func() bool { return !d.busy() && !d.s.Ready(d.side) })
		d.SetReducedMotion(false)
		if d.s.Ended() != nil {
			break
		}
	}
	if d.s.Phase(opp) != dosrules.PhaseAction && d.s.Phase(opp) != dosrules.PhaseMovement {
		t.Fatal("no phase")
	}
}

func TestPlanInTheFrameWithoutCommands(t *testing.T) {
	d := newDosTest(t, "level1_cp", passPlanner{})
	d.Tick()
	if d.planning() || !d.s.Ready(1-d.side) {
		t.Errorf("pass planner: planning %v, ready %v", d.planning(), d.s.Ready(1-d.side))
	}
}
