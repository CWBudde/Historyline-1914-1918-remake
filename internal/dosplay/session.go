// Package dosplay is a game session on the original rules: a level bound to
// the original records (internal/dosbind), the command cycle of
// internal/dosrules on it, and the queries, commands and events in the terms
// a front end or a computer side works with (hexes, unit views, routes,
// battles).
//
// The original has no line of sight. A side sees its own map
// (dosrules.Game.Maps): the other side's units stand where they stood at the
// last phase change, a move changes only the mover's map, and the phase
// change makes both maps alike. So every hex of the level is visible, and a
// side's view is the occupants of its map.
//
// Every command and the resolution change the state at once and return the
// events in the order they happened; a front end that animates them keeps
// what it needs from the events (or from a View taken before) and plays
// them back. Lists are in a fixed order, so equal command sequences give
// equal events.
package dosplay

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// Options are the game options of a session: the move limit (4, 8 or 16
// orders per side and phase, 0 none), which sides the computer plays (the
// limit holds for human sides only) and the rules profile (nil:
// dosrules.Full; the board applies the level's season).
type Options struct {
	Limit    int
	Computer [2]bool
	Profile  *dosrules.Profile
}

// Session is a battle on the original rules. Game and Board are exposed for
// tools and tests; change the state only through the session's commands.
type Session struct {
	Game  *dosrules.Game
	Board *dosbind.Board
	Level *level.Level

	set     *data.Set
	version int
	// routes are the moves of each side since the last phase change
	// (Moved events), revealed to the other side after it.
	routes [2][]Event
	ids    map[int]string // unit ids of built units
}

// New binds a level and starts a battle on it: side 0 (the Central Powers)
// in the movement phase.
func New(lvl *level.Level, set *data.Set, o Options) (*Session, error) {
	b, err := dosbind.Bind(lvl, set)
	if err != nil {
		return nil, err
	}
	return NewBoard(b, lvl, set, o)
}

// NewBoard starts a battle on a bound board, such as an original map
// (dosbind.FromOriginal); lvl is the level the front end draws.
func NewBoard(b *dosbind.Board, lvl *level.Level, set *data.Set, o Options) (*Session, error) {
	if err := fits(b, lvl); err != nil {
		return nil, err
	}
	p := dosrules.Full
	if o.Profile != nil {
		p = *o.Profile
	}
	g, err := b.Game(p)
	if err != nil {
		return nil, fmt.Errorf("dosplay: %w", err)
	}
	g.Limit = o.Limit
	for k := range 2 {
		g.Players[k].Computer = o.Computer[k]
	}
	return &Session{Game: g, Board: b, Level: lvl, set: set, ids: map[int]string{}}, nil
}

// Load resumes a battle saved by Save on its level. The routes of moves made
// before the save are not part of it, so the next resolution reveals only
// the moves made since.
func Load(lvl *level.Level, set *data.Set, saved []byte) (*Session, error) {
	b, err := dosbind.Bind(lvl, set)
	if err != nil {
		return nil, err
	}
	return LoadBoard(b, lvl, set, saved)
}

// LoadBoard resumes a battle saved by Save on its bound board.
func LoadBoard(b *dosbind.Board, lvl *level.Level, set *data.Set, saved []byte) (*Session, error) {
	if err := fits(b, lvl); err != nil {
		return nil, err
	}
	g, err := dosrules.LoadGame(saved)
	if err != nil {
		return nil, err
	}
	if g.Width != b.Width || g.Height != b.Height || g.Profile.Winter() != b.Winter {
		return nil, fmt.Errorf("dosplay: a saved %d×%d battle (winter %v) on a %d×%d level (winter %v)",
			g.Width, g.Height, g.Profile.Winter(), b.Width, b.Height, b.Winter)
	}
	return &Session{Game: g, Board: b, Level: lvl, set: set, ids: map[int]string{}}, nil
}

// fits checks that lvl has the hex layout of board b.
func fits(b *dosbind.Board, lvl *level.Level) error {
	if g := lvl.Grid(); 2*g.Width != b.Width || g.Lines != 2*b.Height-1 {
		return fmt.Errorf("dosplay: a level of %d lines of %d hexes for a %d×%d board", g.Lines, g.Width, b.Width, b.Height)
	}
	return nil
}

// Save is the battle's state (dosrules.Game.Save).
func (s *Session) Save() ([]byte, error) { return s.Game.Save() }

// Faction is the remake player of a side (2: neutral).
func (s *Session) Faction(side int) game.Player {
	if side < 0 || side > 2 {
		return game.None
	}
	return dosbind.Factions[side]
}

// SideOf is the side of a remake player, 2 for none.
func (s *Session) SideOf(p game.Player) int { return dosbind.Side(p) }

// Phase is a side's phase: dosrules.PhaseMovement or PhaseAction.
func (s *Session) Phase(side int) int { return s.Game.Players[side].Phase }

// Ready reports whether a side has confirmed its orders.
func (s *Session) Ready(side int) bool { return s.Game.Players[side].Ready }

// Turn is the turn counter; it counts up when side 0's movement phase begins.
func (s *Session) Turn() int { return s.Game.Turn }

// Ended is the battle's end, nil while it is open.
func (s *Session) Ended() *dosrules.End { return s.Game.Ended }

// OrdersLeft is the number of orders the move limit still allows a side in
// this phase, -1 when there is no limit (none set, or a computer side).
func (s *Session) OrdersLeft(side int) int {
	p := s.Game.Players[side]
	if s.Game.Limit <= 0 || p.Computer {
		return -1
	}
	return max(0, s.Game.Limit-len(p.Orders))
}

// Energy is a side's energy account.
func (s *Session) Energy(side int) int { return s.Game.Energy[side] }

// Version counts the state changes; a caller may cache what it derives from
// the state until it changes.
func (s *Session) Version() int { return s.version }

// OrderError is a command the rules refuse, in our wording; it unwraps to
// the *dosrules.Refusal.
type OrderError struct {
	Text    string
	Refusal *dosrules.Refusal
}

func (e *OrderError) Error() string { return e.Text }

func (e *OrderError) Unwrap() error { return e.Refusal }

// wrap gives a refusal its text; other errors pass.
func wrap(err error) error {
	var r *dosrules.Refusal
	if !errors.As(err, &r) {
		return err
	}
	text := MessageText(r.Code)
	if r.Code == 0 && r.Reason != "" {
		text = strings.ToUpper(r.Reason[:1]) + r.Reason[1:] + "."
	}
	return &OrderError{Text: text, Refusal: r}
}

func refusal(reason string) error { return wrap(&dosrules.Refusal{Reason: reason}) }

// cell is the original cell of a hex; an error for a hex off the level.
func (s *Session) cell(c hexgrid.Coord) (int, error) {
	cell, ok := s.Board.Cell(c)
	if !ok {
		return 0, fmt.Errorf("dosplay: hex %v is not on the level", c)
	}
	return cell, nil
}

// coord is the hex of a cell, bound or not.
func (s *Session) coord(cell int) hexgrid.Coord {
	c, _ := s.Board.Coord(cell)
	return c
}

// bound reports whether a cell is a hex of the level.
func (s *Session) bound(cell int) bool {
	_, ok := s.Board.Coord(cell)
	return ok
}

// container is the record and slot of side's part of the container records
// that hold unit, false if none does.
func (s *Session) container(side, unit int) (hold, slot int, ok bool) {
	for _, i := range slices.Sorted(maps.Keys(s.Game.Holds)) {
		h := s.Game.Holds[i]
		for k := range 7 {
			if h.Slots[7*side+k] == unit {
				return i, k, true
			}
		}
	}
	return 0, 0, false
}

// startCell is the cell an exit from a container record starts on: its first
// cell, the transporter's for a transporter's record.
func startCell(h dosrules.Hold) int {
	start := h.Cells[0]
	if h.Flags&0x20 != 0 {
		start--
	}
	return start / 2
}

// command is a checked command of side for unit in the phase the side is in;
// cell is where the order goes.
func (s *Session) command(side, unit, cell int) (dosrules.Command, error) {
	if side != 0 && side != 1 {
		return dosrules.Command{}, fmt.Errorf("dosplay: side %d", side)
	}
	u, ok := s.Game.Units[unit]
	if !ok {
		return dosrules.Command{}, refusal("no such unit")
	}
	c := dosrules.Command{Kind: dosrules.CmdMove, Side: side, Unit: unit, Cell: cell}
	switch {
	case s.Phase(side) == dosrules.PhaseAction:
		c.Kind = dosrules.CmdAttack
	case u.Flags4&0x4000 != 0:
		hold, slot, ok := s.container(side, unit)
		if !ok {
			return c, refusal("the unit is not the side's to order")
		}
		c.Kind, c.Hold, c.Slot = dosrules.CmdExit, hold, slot
	}
	return c, nil
}

// CanOrder reports whether unit may get an order in side's current phase: a
// move or an exit in the movement phase, an attack in the action phase.
func (s *Session) CanOrder(side, unit int) error {
	u, ok := s.Game.Units[unit]
	if !ok {
		return refusal("no such unit")
	}
	c, err := s.command(side, unit, u.Pos[side&1]/2)
	if err != nil {
		return err
	}
	err = s.Game.Check(c)
	// The unit's own cell holds no target; every other check passed.
	if r := (*dosrules.Refusal)(nil); c.Kind == dosrules.CmdAttack && errors.As(err, &r) && r.Code == dosrules.MsgNoTarget {
		return nil
	}
	return wrap(err)
}

// Destinations are the hexes unit of side may be ordered to in the movement
// phase.
func (s *Session) Destinations(side, unit int) (map[hexgrid.Coord]bool, error) {
	if err := s.canMove(side, unit); err != nil {
		return nil, err
	}
	cells, err := s.Game.Profile.Destinations(s.Game.Movement(side, unit))
	if err != nil {
		return nil, err
	}
	return s.hexes(cells, s.Game.Units[unit].Cell(side)), nil
}

// canMove is CanOrder for a move of a unit on the map.
func (s *Session) canMove(side, unit int) error {
	if err := s.CanOrder(side, unit); err != nil {
		return err
	}
	if s.Phase(side) != dosrules.PhaseMovement || s.Game.Units[unit].Flags4&0x4000 != 0 {
		return refusal("the unit does not move from the map now")
	}
	return nil
}

// hexes are the bound cells of cells but skip, as hexes.
func (s *Session) hexes(cells []int, skip int) map[hexgrid.Coord]bool {
	out := map[hexgrid.Coord]bool{}
	for _, c := range cells {
		if p, ok := s.Board.Coord(c); ok && c != skip {
			out[p] = true
		}
	}
	return out
}

// Route is the route the move of unit to a hex takes, start first; nil
// where there is none.
func (s *Session) Route(side, unit int, to hexgrid.Coord) ([]hexgrid.Coord, error) {
	if err := s.canMove(side, unit); err != nil {
		return nil, err
	}
	cell, err := s.cell(to)
	if err != nil {
		return nil, err
	}
	cells, err := s.Game.Profile.Route(s.Game.Movement(side, unit), cell)
	if err != nil || cells == nil {
		return nil, err
	}
	out := make([]hexgrid.Coord, len(cells))
	for i, c := range cells {
		out[i] = s.coord(c)
	}
	return out, nil
}

// Costs are the entry costs of the level's hexes for unit of side on the
// terrain alone, as the reach search counts them without any unit on the
// map; a hex the unit cannot enter is missing. A computer side plans routes
// beyond one move with them. The unit may be held.
func (s *Session) Costs(side, unit int) (map[hexgrid.Coord]int, error) {
	if side != 0 && side != 1 {
		return nil, fmt.Errorf("dosplay: side %d", side)
	}
	u, ok := s.Game.Units[unit]
	if !ok || u.Destroyed() {
		return nil, refusal("no such unit")
	}
	t := s.Game.Profile.Tables.Move
	if t == nil {
		return nil, dosrules.ErrNoMovement
	}
	m := s.Game.Movement(side, unit)
	p := m.PieceAt(unit)
	m.Units = map[int]dosrules.Piece{}
	if u.Flags4&0x4000 == 0 {
		m.Units[unit] = p
	}
	buf := t.Costs(m)
	out := map[hexgrid.Coord]int{}
	for c := range s.Game.Width * s.Game.Height {
		// A cost is stored negated; -80 marks a closed cell.
		if v := int(buf[c/s.Game.Width*64+c%s.Game.Width]); v < 0 && v > -80 {
			if p, ok := s.Board.Coord(c); ok {
				out[p] = -v
			}
		}
	}
	return out, nil
}

// ExitDestinations are the hexes the unit in slot of side's part of
// container record hold may leave it for.
func (s *Session) ExitDestinations(side, hold, slot int) (map[hexgrid.Coord]bool, error) {
	if err := s.canExit(side, hold, slot); err != nil {
		return nil, err
	}
	m, err := s.Game.ExitMovement(side, hold, slot)
	if err != nil {
		return nil, wrap(err)
	}
	cells, err := s.Game.Profile.Destinations(m)
	if err != nil {
		return nil, err
	}
	return s.hexes(cells, startCell(s.Game.Holds[hold])), nil
}

// canExit checks an exit without its destination.
func (s *Session) canExit(side, hold, slot int) error {
	h, ok := s.Game.Holds[hold]
	if !ok {
		return refusal("no such container")
	}
	return wrap(s.Game.Check(dosrules.Command{Kind: dosrules.CmdExit, Side: side, Hold: hold, Slot: slot, Cell: startCell(h)}))
}

// Targets are the hexes of side's map unit may attack now.
func (s *Session) Targets(side, unit int) map[hexgrid.Coord]bool {
	out := map[hexgrid.Coord]bool{}
	if side != 0 && side != 1 || s.Phase(side) != dosrules.PhaseAction || s.CanOrder(side, unit) != nil {
		return out
	}
	m := s.Game.Maps[side]
	for c := range s.Game.Width * s.Game.Height {
		if m[2*c+1] > dosrules.SlotFree || !s.bound(c) {
			continue
		}
		if s.Game.Check(dosrules.Command{Kind: dosrules.CmdAttack, Side: side, Unit: unit, Cell: c}) == nil {
			out[s.coord(c)] = true
		}
	}
	return out
}

// Battlefield is side's map as the order handler fights on it.
func (s *Session) Battlefield(side int) (*dosrules.Battlefield, error) {
	return s.Game.Battlefield(side)
}

// apply runs a command and turns its events.
func (s *Session) apply(c dosrules.Command, start int) ([]Event, error) {
	raw, err := s.Game.Apply(c)
	if err != nil {
		return nil, wrap(err)
	}
	s.version++
	evs := make([]Event, 0, len(raw))
	for _, r := range raw {
		var ev Event
		switch r.Kind {
		case dosrules.EvMoved, dosrules.EvRefused:
			ev = s.moved(r, start)
		default:
			ev = s.event(r, nil)
		}
		if ev.Kind == Moved {
			s.routes[c.Side] = append(s.routes[c.Side], ev)
		}
		evs = append(evs, ev)
	}
	return evs, nil
}

// Move orders unit of side to a hex; the move runs at once. A move the
// order code refuses gives a Refused event.
func (s *Session) Move(side, unit int, to hexgrid.Coord) ([]Event, error) {
	cell, err := s.cell(to)
	if err != nil {
		return nil, err
	}
	if err := s.canMove(side, unit); err != nil {
		return nil, err
	}
	return s.apply(dosrules.Command{Kind: dosrules.CmdMove, Side: side, Unit: unit, Cell: cell}, s.Game.Units[unit].Cell(side))
}

// Exit takes the unit in slot of side's part of container record hold out
// to a hex; the move runs at once.
func (s *Session) Exit(side, hold, slot int, to hexgrid.Coord) ([]Event, error) {
	cell, err := s.cell(to)
	if err != nil {
		return nil, err
	}
	h, ok := s.Game.Holds[hold]
	if !ok {
		return nil, refusal("no such container")
	}
	if cell == startCell(h) {
		return nil, refusal("the unit must leave the container")
	}
	return s.apply(dosrules.Command{Kind: dosrules.CmdExit, Side: side, Hold: hold, Slot: slot, Cell: cell}, startCell(h))
}

// Attack gives unit of side the attack order on a hex; it waits in the
// side's order list until the resolution.
func (s *Session) Attack(side, unit int, target hexgrid.Coord) ([]Event, error) {
	cell, err := s.cell(target)
	if err != nil {
		return nil, err
	}
	return s.apply(dosrules.Command{Kind: dosrules.CmdAttack, Side: side, Unit: unit, Cell: cell}, -1)
}

// Repair repairs the unit in slot of side's part of building record hold.
func (s *Session) Repair(side, hold, slot int) ([]Event, error) {
	return s.apply(dosrules.Command{Kind: dosrules.CmdRepair, Side: side, Hold: hold, Slot: slot}, -1)
}

// Build builds a unit of UNIT.DAT type typ in building record hold.
func (s *Session) Build(side, hold, typ int) ([]Event, error) {
	return s.apply(dosrules.Command{Kind: dosrules.CmdBuild, Side: side, Hold: hold, Type: typ}, -1)
}

// Confirm completes side's orders.
func (s *Session) Confirm(side int) []Event {
	s.version++
	return s.events(s.Game.Confirm(side))
}

// Unconfirm lets side give orders again.
func (s *Session) Unconfirm(side int) []Event {
	raw := s.Game.Unconfirm(side)
	if len(raw) > 0 {
		s.version++
	}
	return s.events(raw)
}

func (s *Session) events(raw []dosrules.Event) []Event {
	out := make([]Event, 0, len(raw))
	for _, r := range raw {
		out = append(out, s.event(r, nil))
	}
	return out
}

// Resolve resolves the action side's attack orders once both sides have
// confirmed, then changes the phase. The battles are fought on the action
// side's map as it stood before; each Battle event carries the fighters as
// they were then. After the PhaseChanged event a Revealed event follows for
// every move the movement side made this phase whose unit survived, in the
// order made.
func (s *Session) Resolve() ([]Event, error) {
	a, err := s.Game.Action()
	if err != nil {
		return nil, err
	}
	b := &battles{side: a, views: map[int]UnitView{}, state: map[int][3]int{}}
	for i, u := range s.Game.Units {
		b.views[i] = s.unitView(i, a)
		b.state[i] = [3]int{u.Size, u.Alt, u.Exp}
	}
	raw, err := s.Game.Resolve()
	if len(raw) > 0 {
		s.version++
	}
	var out []Event
	for _, r := range raw {
		out = append(out, s.event(r, b))
		if r.Kind != dosrules.EvPhase {
			continue
		}
		for _, mv := range s.routes[1-a] {
			if s.Game.Units[mv.Unit].Destroyed() {
				continue
			}
			mv.Kind, mv.Text = Revealed, ""
			out = append(out, mv)
		}
		s.routes = [2][]Event{}
	}
	return out, err
}

// Orders are side's queued attacks in list order, as Ordered events.
func (s *Session) Orders(side int) []Event {
	if side != 0 && side != 1 || s.Phase(side) != dosrules.PhaseAction {
		return nil
	}
	var out []Event
	for _, u := range s.Game.Players[side].Orders {
		l := s.Game.Units[u]
		out = append(out, Event{Kind: Ordered, Side: side, Unit: u, From: s.coord(l.Cell(side)), To: s.coord(l.Target / 2)})
	}
	return out
}
