package llmplay

import (
	"fmt"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// Replay is a recorded game opened for replay: the transcript and a fresh
// session on the header's level in the state before the first resolution.
type Replay struct {
	Dir     string
	Header  Header
	Lines   []Line
	Session *dosplay.Session
}

// OpenReplay reads the transcript in dir and builds the fresh session. The
// session's state must match the first line's StateBefore; a transcript
// without lines opens as is.
func OpenReplay(dir string) (*Replay, error) {
	h, lines, err := ReadTranscript(dir)
	if err != nil {
		return nil, err
	}
	month, mode, err := ParseLevelRef(h.Level)
	if err != nil {
		return nil, err
	}
	lvl, err := ShipLevel(month, mode)
	if err != nil {
		return nil, err
	}
	s, err := NewSession(lvl, h.LLMSide)
	if err != nil {
		return nil, err
	}
	if len(lines) > 0 {
		got, err := StateHash(s)
		if err != nil {
			return nil, err
		}
		if got != lines[0].StateBefore {
			return nil, fmt.Errorf("resolution %d: opening state hash %s, transcript has %s", lines[0].Resolution, got, lines[0].StateBefore)
		}
	}
	return &Replay{Dir: dir, Header: h, Lines: lines, Session: s}, nil
}

// SideCommands are the orders side gave in l, in recorded order: every
// command of side but the cycle commands confirm, unconfirm and resolve.
// A surrender is an order and stays.
func SideCommands(l Line, side int) []dosrules.Command {
	var out []dosrules.Command
	for _, c := range l.Commands {
		switch c.Kind {
		case dosplay.CmdConfirm, dosplay.CmdUnconfirm, dosplay.CmdResolve:
			continue
		}
		if c.Side == side {
			out = append(out, c)
		}
	}
	return out
}

// CheckResolution compares the session's state with the state l recorded
// after its resolution.
func CheckResolution(s *dosplay.Session, l Line) error {
	got, err := StateHash(s)
	if err != nil {
		return err
	}
	if got != l.StateAfter {
		return fmt.Errorf("resolution %d: state hash %s, transcript has %s", l.Resolution, got, l.StateAfter)
	}
	return nil
}
