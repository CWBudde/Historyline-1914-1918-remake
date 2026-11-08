package llmplay

import (
	"errors"
	"fmt"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosai"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// OwnCommands are the commands side gave in l, its confirm included, in
// recorded order: what side has to replay before the opponent plans.
func OwnCommands(l Line, side int) []dosrules.Command {
	var out []dosrules.Command
	for _, c := range l.Commands {
		if c.Kind != dosplay.CmdResolve && c.Side == side {
			out = append(out, c)
		}
	}
	return out
}

// ErrOrdersEnd is returned by OrdersPlayer once its recording has no
// resolution left.
var ErrOrdersEnd = errors.New("llmplay: the recorded orders have ended")

// OrdersPlayer plays a recorded side's orders back, one recorded resolution
// per phase, against whatever the opponent does now. It is how a game whose
// orders came from outside (an LLM, a file) is carried to a changed engine
// or computer side: the first order the session refuses ends it with an
// error naming the resolution.
type OrdersPlayer struct {
	Lines []Line
	next  int
	text  string
}

// PlayPhase implements PhasePlayer.
func (p *OrdersPlayer) PlayPhase(s *dosplay.Session, side int) ([]dosplay.Event, error) {
	if s.Ended() != nil || s.Ready(side) {
		return nil, nil
	}
	if p.next >= len(p.Lines) {
		return nil, ErrOrdersEnd
	}
	l := p.Lines[p.next]
	p.next++
	p.text = l.OrdersText
	evs, err := s.Replay(OwnCommands(l, side))
	if err != nil {
		return evs, fmt.Errorf("resolution %d: %w", l.Resolution, err)
	}
	if s.Ended() == nil && !s.Ready(side) {
		evs = append(evs, s.Confirm(side)...)
	}
	return evs, nil
}

// OrdersText is the recorded orders text of the resolution just played.
func (p *OrdersPlayer) OrdersText() string { return p.text }

// LiveReplay replays the recorded side's orders of the transcript at path
// (a game dir or a transcript file) with today's computer side planning the
// opponent, and checks every state hash. Where VerifyReplay proves the
// rules still give the recorded game, LiveReplay proves the recording is
// still a game against the current dosai: a planner change that alters any
// opponent order fails it at that resolution. It returns the end and the
// number of resolutions played.
func LiveReplay(path string) (*dosrules.End, int, error) {
	rp, err := OpenReplay(path)
	if err != nil {
		return nil, 0, err
	}
	s, side := rp.Session, rp.Header.LLMSide
	planner := &dosai.Planner{}
	for i, l := range rp.Lines {
		if s.Ended() != nil {
			return nil, i, fmt.Errorf("resolution %d: the game ended at resolution %d", l.Resolution, i)
		}
		if _, err := s.Replay(OwnCommands(l, side)); err != nil {
			return nil, i, fmt.Errorf("resolution %d: %w", l.Resolution, err)
		}
		if _, err := planner.Plan(s, 1-side); err != nil {
			return nil, i, fmt.Errorf("resolution %d: computer side: %w", l.Resolution, err)
		}
		if s.Ended() == nil {
			if _, err := s.Resolve(); err != nil {
				return nil, i, fmt.Errorf("resolution %d: %w", l.Resolution, err)
			}
		}
		if err := CheckResolution(s, l); err != nil {
			return nil, i, fmt.Errorf("%w (the computer side or the rules changed since the recording)", err)
		}
	}
	return s.Ended(), len(rp.Lines), nil
}
