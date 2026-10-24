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
// session on the game's goroutine one by one, each move walking on the
// computer's panel, and the side confirms after the last. A planner without
// commands (passPlanner, the test planners) plans in the frame as before, and
// so does every planner when the copy or the replay fails.

// planMinFrames is how long the planning shows at least, so the progress bar
// is seen to run.
const planMinFrames = 45

// actPauseFrames is how long the computer's panel rests before each move:
// its camera slides to the unit, and the moves read one by one.
const actPauseFrames = 12

// computerTurn is the computer side carrying out its planned commands.
type computerTurn struct {
	side int
	cmds []dosrules.Command // still to replay
	walk *walk              // the move walking on the side's panel
	wait int                // frames of rest before walk starts
}

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

// acting reports whether the computer is carrying out its planned commands.
func (d *dosControl) acting() bool { return d.act != nil }

// walkOf is the walk on side's panel: the human side's own move or the
// computer's move being carried out.
func (d *dosControl) walkOf(side int) *walk {
	if side == d.side {
		return d.walk
	}
	if d.act != nil {
		return d.act.walk
	}
	return nil
}

// planStep runs one frame of the computer side's planning: it carries out
// planned commands, starts a run, or takes a finished one in, or plans in the
// frame for a planner that cannot plan off it.
func (d *dosControl) planStep(side int) {
	if d.act != nil {
		d.actStep()
		return
	}
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

// finishPlan hands a finished run's commands to be carried out (actStep); a
// failed run plans in the frame instead, on the session as it is now.
func (d *dosControl) finishPlan(side int, res planResult) {
	if res.err == nil {
		d.act = &computerTurn{side: side, cmds: res.cmds}
		d.actStep()
		return
	}
	d.planFailed(side, res.err)
}

// planFailed reports a failed planning or replay and plans in the frame on
// the session as it is now.
func (d *dosControl) planFailed(side int, err error) {
	d.Message = "computer side: " + err.Error()
	if !d.s.Ready(side) {
		d.plan(side)
	}
}

// actStep runs one frame of the computer carrying out its commands: the move
// walking on its panel plays on, else the next commands replay on the session
// up to the next move, which rests a moment and then walks. After the last
// the side confirms. With reduced motion every command replays at once.
func (d *dosControl) actStep() {
	a := d.act
	if d.reduced {
		a.walk, a.wait = nil, 0
	}
	switch {
	case a.wait > 0:
		a.wait--
		return
	case !a.walk.Done():
		a.walk.Tick()
		return
	}
	for len(a.cmds) > 0 {
		evs, err := d.s.Replay(a.cmds[:1])
		a.cmds = a.cmds[1:]
		if err != nil {
			d.act = nil
			d.planFailed(a.side, err)
			return
		}
		for _, ev := range evs {
			if ev.Kind != dosplay.Moved {
				continue
			}
			d.face(a.side, ev)
			if w := newWalkMotion(ev.Path, ev.Unit, d.reduced); w != nil {
				if len(ev.Rear) == len(ev.Path) {
					w.tail = ev.Rear
				}
				a.walk, a.wait = w, actPauseFrames
			}
		}
		if a.walk != nil && !a.walk.Done() {
			return
		}
	}
	d.act = nil
	if !d.s.Ready(a.side) {
		d.s.Confirm(a.side)
	}
}
