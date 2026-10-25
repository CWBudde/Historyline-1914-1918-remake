package dosplay

import (
	"fmt"
	"maps"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// A session logs every command that changes its state: the orders as the
// rules engine takes them (moves, exits, attacks, repairs, builds) and the
// confirmations, withdrawals and resolutions under the kinds below. A
// computer side may plan on a Clone off the main thread while the other side
// gives orders on the session; the clone's commands since a Mark then
// Replay onto the session with the events the direct calls give.

// The commands of the log beyond the orders of dosrules.
const (
	CmdConfirm   dosrules.CommandKind = "confirm"   // Confirm(Side)
	CmdUnconfirm dosrules.CommandKind = "unconfirm" // Unconfirm(Side), logged when it withdrew a confirmation
	CmdResolve   dosrules.CommandKind = "resolve"   // Resolve()
)

// Clone is an independent copy of the session: the same state, profile,
// routes to reveal, version and log. It shares only what no command
// changes (the board, the level and the data set), so the clone may run on
// another goroutine while the session is played on.
func (s *Session) Clone() (*Session, error) {
	raw, err := s.Game.Save()
	if err != nil {
		return nil, err
	}
	g, err := dosrules.LoadGame(raw)
	if err != nil {
		return nil, err
	}
	// A save names the profile by its tables; keep the session's own.
	g.Profile = s.Game.Profile
	c := &Session{
		Game: g, Board: s.Board, Level: s.Level, set: s.set, version: s.version, ids: maps.Clone(s.ids),
		log: slices.Clone(s.log),
	}
	for k := range s.routes {
		c.routes[k] = slices.Clone(s.routes[k])
	}
	return c, nil
}

// Mark is the position in the log the next command takes.
func (s *Session) Mark() int { return len(s.log) }

// CommandsSince are the commands logged since mark (Mark), oldest first;
// nil for none or a mark the log does not have.
func (s *Session) CommandsSince(mark int) []dosrules.Command {
	if mark < 0 || mark >= len(s.log) {
		return nil
	}
	return slices.Clone(s.log[mark:])
}

// Replay carries out logged commands in order through the session's own
// calls (Move, Exit, Attack, Repair, Build, Confirm, Unconfirm, Resolve)
// and returns their events. It stops at the first command the session
// refuses and returns the events so far with the error.
func (s *Session) Replay(cmds []dosrules.Command) ([]Event, error) {
	var out []Event
	for i, c := range cmds {
		evs, err := s.replay(c)
		out = append(out, evs...)
		if err != nil {
			return out, fmt.Errorf("dosplay: replay of command %d (%s): %w", i, c.Kind, err)
		}
	}
	return out, nil
}

// replay carries out one logged command.
func (s *Session) replay(c dosrules.Command) ([]Event, error) {
	switch c.Kind {
	case dosrules.CmdRepair:
		return s.Repair(c.Side, c.Hold, c.Slot)
	case dosrules.CmdBuild:
		return s.Build(c.Side, c.Hold, c.Type)
	case CmdResolve:
		return s.Resolve()
	}
	if c.Side != 0 && c.Side != 1 {
		return nil, fmt.Errorf("dosplay: side %d", c.Side)
	}
	switch c.Kind {
	case CmdConfirm:
		return s.Confirm(c.Side), nil
	case CmdUnconfirm:
		return s.Unconfirm(c.Side), nil
	}
	to, ok := s.Board.Coord(c.Cell)
	if !ok {
		return nil, fmt.Errorf("dosplay: cell %d is not on the level", c.Cell)
	}
	switch c.Kind {
	case dosrules.CmdMove:
		return s.Move(c.Side, c.Unit, to)
	case dosrules.CmdExit:
		return s.Exit(c.Side, c.Hold, c.Slot, to)
	case dosrules.CmdAttack:
		return s.Attack(c.Side, c.Unit, to)
	}
	return nil, fmt.Errorf("dosplay: unknown command %q", c.Kind)
}
