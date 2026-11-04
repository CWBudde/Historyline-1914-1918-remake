package main

import (
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/llmplay"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// A recorded llmplay game (-replay DIR) plays in the split battle screen:
// both sides carry out the transcript's orders the way the computer side
// carries out its own (plan.go), the recorded side first, then the other,
// then the resolution, whose state must match the recorded one. Nothing
// but quitting, fullscreen and the scenes takes input.

// transcriptPlanner gives both sides' orders of a recorded game, a line of
// the transcript per resolution. It plans in no time, so a replay takes the
// same frames every time.
type transcriptPlanner struct {
	lines []llmplay.Line
	at    int   // the line of the resolution being replayed
	err   error // the first mismatch or failed order; the replay stops
}

var (
	_ dosPlanner     = (*transcriptPlanner)(nil)
	_ commandPlanner = (*transcriptPlanner)(nil)
)

// Commands are side's orders of the current line, none past the last.
func (p *transcriptPlanner) Commands(_ *dosplay.Session, side int) ([]dosrules.Command, error) {
	if p.at >= len(p.lines) {
		return nil, nil
	}
	return llmplay.SideCommands(p.lines[p.at], side), nil
}

// Plan replays side's orders of the current line at once and confirms.
func (p *transcriptPlanner) Plan(s *dosplay.Session, side int) ([]dosplay.Event, error) {
	cmds, err := p.Commands(s, side)
	if err != nil {
		return nil, err
	}
	evs, err := s.Replay(cmds)
	if err != nil {
		return evs, err
	}
	return append(evs, s.Confirm(side)...), nil
}

// result is side's orders as a finished planning.
func (p *transcriptPlanner) result(side int) planResult {
	cmds, err := p.Commands(nil, side)
	return planResult{cmds, err}
}

// stopped reports whether the replay gives no more orders: a mismatch, or
// every line replayed.
func (p *transcriptPlanner) stopped() bool { return p.err != nil || p.at >= len(p.lines) }

// startReplay plays both sides of the session from lines.
func (d *dosControl) startReplay(lines []llmplay.Line) {
	p := &transcriptPlanner{lines: lines}
	d.replay, d.self, d.planner = p, p, p
}

// replaying reports whether a replay still gives orders.
func (d *dosControl) replaying() bool { return d.replay != nil && !d.replay.stopped() }

// checkReplay compares the state after a resolution with the recorded line
// and moves on to the next; a mismatch stops the replay.
func (d *dosControl) checkReplay() {
	p := d.replay
	if p == nil || p.stopped() {
		return
	}
	if err := llmplay.CheckResolution(d.s, p.lines[p.at]); err != nil {
		d.stopReplay(err)
		return
	}
	p.at++
}

// stopReplay stops the replay on err and shows it.
func (d *dosControl) stopReplay(err error) {
	if d.replay.err == nil {
		d.replay.err = err
	}
	d.Message = "replay: " + d.replay.err.Error()
}

// replayErr is why the replay stopped early, nil while it matches the
// transcript (and outside a replay).
func (d *dosControl) replayErr() error {
	if d.replay == nil {
		return nil
	}
	return d.replay.err
}

// replayDone reports whether a replay has played out: the battle's end
// shown, or the replay stopped (a mismatch, a transcript ending before the
// battle does) and its last resolution played.
func (d *dosControl) replayDone() bool {
	if d.replay == nil {
		return false
	}
	if d.over {
		return true
	}
	return d.replay.stopped() && d.act == nil && !d.busy() && d.s.Ended() == nil
}

// replayGame is the game on a recorded llmplay game: the recorded side's
// view on its panel, both sides played from the transcript.
func replayGame(rp *llmplay.Replay, reveal bool) *game {
	g := newDosGame(rp.Session.Level, rp.Session, rp.Header.LLMSide, nil, reveal)
	g.dos.startReplay(rp.Lines)
	return g
}

// updateReplay takes a replay's input: Q or Esc quits, F11 and F9 work as
// in play, nothing else reaches the game. It reports false to quit.
func (g *game) updateReplay() bool {
	for _, a := range actionsFor(inpututil.IsKeyJustPressed) {
		switch a {
		case actQuit, actCancel:
			return false
		case actFullscreen:
			ebiten.SetFullscreen(!ebiten.IsFullscreen())
		case actScenes:
			g.dos.ToggleScenes()
		default:
		}
	}
	showPointer(true)
	return true
}
