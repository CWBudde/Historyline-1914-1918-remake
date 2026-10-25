package main

import (
	"maps"
	"testing"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosai"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
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
	d := newDosTest(t, "monthly/1914-08:cp", computerPlanner())
	same := newDosTest(t, "monthly/1914-08:cp", computerPlanner())
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
	d := newDosTest(t, "monthly/1914-08:cp", passPlanner{})
	d.Tick()
	if d.planning() || !d.s.Ready(1-d.side) {
		t.Errorf("pass planner: planning %v, ready %v", d.planning(), d.s.Ready(1-d.side))
	}
}

// TestComputerMovesWalk: after planning, the computer carries out its orders
// one by one, each move walking on its own panel in view of its camera, and
// confirms only after the last. Its units turn on its own map only until the
// phase change reveals their moves.
func TestComputerMovesWalk(t *testing.T) {
	g, d := newSplitTest(t, computerPlanner(), 1800, 1000)
	opp := 1 - d.side
	facings := func(side int) map[int]unitsprite.Facing {
		out := map[int]unitsprite.Facing{}
		for _, u := range d.s.TrueView().Units {
			out[u.Index] = d.FacingOn(side, u.Index)
		}
		return out
	}
	walks := 0
	for turn := 0; turn < 4 && walks == 0; turn++ {
		tickUntil(t, d, func() bool { return d.acting() || d.s.Ready(opp) })
		before := facings(d.side)
		var last *walk
		for d.acting() {
			if d.s.Ready(opp) {
				t.Fatalf("turn %d: confirmed while carrying out its orders", turn)
			}
			if w := g.wellOf(opp); w.text != "CARRYING OUT ITS ORDERS" {
				t.Fatalf("turn %d: well %q while acting", turn, w.text)
			}
			if w := d.walkOf(opp); !w.Done() {
				if w != last {
					walks, last = walks+1, w
				}
				end := w.path[len(w.path)-1]
				_, standing := d.unitsOf(opp, false)[end] // not inside a building
				g.drawing, g.drawSide = true, opp
				if standing && !g.dosWalking(end) {
					t.Errorf("turn %d: the walker's unit stands on %v before it arrives", turn, end)
				}
				g.drawSide = d.side
				if g.dosWalking(end) {
					t.Errorf("turn %d: the computer's walk shows on the human's map", turn)
				}
				g.drawing = false
				g.updateSplit()
				if !viewShows(g.splitView(opp), g.playGrid(), w.Hex()) {
					t.Errorf("turn %d: walker on %v out of its panel's view", turn, w.Hex())
				}
			}
			d.Tick()
		}
		if !d.s.Ready(opp) {
			t.Fatalf("turn %d: not confirmed after its orders", turn)
		}
		if after := facings(d.side); !maps.Equal(before, after) {
			t.Errorf("turn %d: the computer's moves turned units on the human's map", turn)
		}
		d.s.Confirm(d.side)
		d.SetReducedMotion(true)
		tickUntil(t, d, func() bool { return !d.busy() && !d.s.Ready(d.side) })
		d.SetReducedMotion(false)
		if last != nil {
			if u, ok := last.unit.(int); ok && d.FacingOn(d.side, u) != d.FacingOn(opp, u) {
				t.Errorf("turn %d: after the phase change unit %d faces %v on the human's map, %v on the computer's", turn, u, d.FacingOn(d.side, u), d.FacingOn(opp, u))
			}
		}
		if d.s.Ended() != nil {
			break
		}
	}
	if walks == 0 {
		t.Fatal("no move of the computer walked")
	}
}

// TestComputerActsAtOnceReduced: with reduced motion the computer's orders
// all land in the frame its planning comes in.
func TestComputerActsAtOnceReduced(t *testing.T) {
	d := newDosTest(t, "monthly/1914-08:cp", computerPlanner())
	d.SetReducedMotion(true)
	tickUntil(t, d, func() bool { return !d.planning() && (d.acting() || d.s.Ready(1-d.side)) })
	if d.acting() || !d.s.Ready(1-d.side) {
		t.Errorf("acting %v, ready %v", d.acting(), d.s.Ready(1-d.side))
	}
}
