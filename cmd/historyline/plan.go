package main

import (
	"sync/atomic"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosai"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// The computer plans off the frame: the planner works on a copy of the
// session in a goroutine, reporting its progress, while the human side keeps
// giving orders on the real one; its commands are then replayed on the real
// session on the game's goroutine. A planner without commands (passPlanner,
// the test planners) plans in the frame as before, and so does every
// planner when the copy or the replay fails.

// planMinFrames is how long the planning shows at least, so the progress bar
// is seen to run.
const planMinFrames = 45

// commandPlanner plans side's orders on a session and returns them as
// commands to replay (dosai.Planner.Commands).
type commandPlanner interface {
	Commands(s *dosplay.Session, side int) ([]dosrules.Command, error)
}

// planRun is one planning off the frame.
type planRun struct {
	side        int
	done, total atomic.Int64
	result      chan planResult
	frames      int // frames shown so far
}

// planResult is what a planning run gave.
type planResult struct {
	cmds []dosrules.Command
	err  error
}

// planning reports whether the computer is planning off the frame.
func (d *dosControl) planning() bool { return d.run != nil }

// planProgress is how far the planning got, from 0 to 1.
func (d *dosControl) planProgress() float64 {
	if d.run == nil {
		return 0
	}
	done, total := d.run.done.Load(), d.run.total.Load()
	if total <= 0 {
		return 0
	}
	return min(float64(done)/float64(total), 1)
}

// planStep runs one frame of the computer side's planning: it starts a run,
// or takes a finished one in, or plans in the frame for a planner that
// cannot plan off it.
func (d *dosControl) planStep(side int) {
	if d.run == nil {
		if !d.startPlan(side) {
			d.plan(side)
		}
		return
	}
	r := d.run
	r.frames++
	if r.frames < d.planFrames() {
		return
	}
	select {
	case res := <-r.result:
		d.run = nil
		d.finishPlan(side, res)
	default:
	}
}

// planFrames is how long a planning shows at least.
func (d *dosControl) planFrames() int {
	if d.reduced {
		return 0
	}
	return planMinFrames
}

// startPlan starts planning side's orders on a copy of the session and
// reports whether it did.
func (d *dosControl) startPlan(side int) bool {
	cp, ok := d.planner.(commandPlanner)
	if !ok {
		return false
	}
	clone, err := d.s.Clone()
	if err != nil {
		return false
	}
	r := &planRun{side: side, result: make(chan planResult, 1)}
	if p, ok := cp.(*dosai.Planner); ok {
		own := *p
		pause := pagePauser()
		own.Progress = func(done, total int) {
			r.done.Store(int64(done))
			r.total.Store(int64(total))
			if pause != nil {
				pause()
			}
		}
		cp = &own
	}
	d.run = r
	go func() {
		cmds, err := cp.Commands(clone, side)
		r.total.Store(1) // done, also for a phase with nothing to plan
		r.done.Store(1)
		r.result <- planResult{cmds, err}
	}()
	return true
}

// finishPlan replays a finished run's commands on the session; a failed run
// or replay plans in the frame instead, on the session as it is now.
func (d *dosControl) finishPlan(side int, res planResult) {
	if res.err == nil {
		evs, err := d.s.Replay(res.cmds)
		if err == nil {
			if c, ok := planFocus(evs); ok {
				d.lookAt(side, c)
			}
			if !d.s.Ready(side) {
				d.s.Confirm(side)
			}
			return
		}
		res.err = err
	}
	d.Message = "computer side: " + res.err.Error()
	if !d.s.Ready(side) {
		d.plan(side)
	}
}
