package dosrules

import (
	"errors"
	"fmt"
	"slices"
)

// The command cycle. Both sides plan at the same time: the side in the
// movement phase moves its units, one order at a time and at once, on its own
// map; the side in the action phase gives attack orders, which wait in its
// order list. When both sides have confirmed, the order handler resolves the
// action side's list and the phase change brings the two maps into line and
// swaps the phases. Game holds that state; every command and the resolution
// change it at once and return the events in the order they happened, so no
// state waits on an animation.

// The phases (player record +0x16).
const (
	PhaseMovement = 1
	PhaseAction   = 2
)

// Status messages of the cycle.
const (
	MsgNoTarget    = 0x45 // the attack order: no target in the cell
	MsgAttack      = 0x49 // the attack order taken
	MsgLimit       = 0x53 // the move limit reached
	MsgReadyUndone = 0x34 // the fire button cancels the confirmation
)

// Player is a side's part of the cycle: its player record (32e2:4d88 +
// 0x31·side in the demo) and its order list.
type Player struct {
	Phase    int  // +0x16: PhaseMovement or PhaseAction
	Ready    bool // +0x17 = 5: the side has confirmed its orders
	Computer bool // 32e2:4d29/4d40 bit 2; the move limit holds for human sides only
	// Orders is the side's order list (32e2:4d3a + 0x17·side): the units of
	// its moves, exits and attacks in the order given. Its length is the
	// count (4d3e) the move limit compares. Only the action side's list is
	// resolved.
	Orders []int
}

// Game is the state of a battle the command cycle works on: both sides'
// maps (two numbers per cell, the terrain record and the occupant, 0xff
// empty), the live records of the placed units and the container records,
// the players, the turn counter (32e2:4cab) and the move limit (the option
// 1653:253c: 4, 8 or 16 orders per side and phase; 0 none).
type Game struct {
	Profile Profile
	Width   int
	Height  int
	Maps    [2][]int
	Units   map[int]Live
	Holds   map[int]Hold
	Players [2]Player
	Turn    int
	Limit   int
}

// Placement is a unit as a scenario places it: its type, owner (0 or 1, 2
// neutral), cell and strength.
type Placement struct {
	Type, Owner, Cell int
	Size, Alt, Exp    int // Size 0: the type's full group
}

// NewGame sets up a battle on terrain (a terrain record per cell): both maps
// show every unit, side 0 starts in the movement phase and side 1 in the
// action phase (0661:0bb2). A unit's live record gets what the constructor
// (159e:0006) is known to copy: the budget (type byte 0), the class word
// (+0x0d) with the owner's bit into +4 and the flags word (+0x11) into +6;
// its hold is 0xff.
func NewGame(p Profile, width, height int, terrain []int, units map[int]Placement, holds map[int]Hold) (*Game, error) {
	t, err := p.moveTables()
	if err != nil {
		return nil, err
	}
	if len(terrain) != width*height {
		return nil, fmt.Errorf("dosrules: %d terrain records for %d cells", len(terrain), width*height)
	}
	g := &Game{Profile: p, Width: width, Height: height, Units: map[int]Live{}, Holds: map[int]Hold{}}
	for k := range 2 {
		g.Maps[k] = make([]int, 2*width*height)
		for c, tr := range terrain {
			g.Maps[k][2*c], g.Maps[k][2*c+1] = tr, 0xFF
		}
	}
	for i, u := range units {
		if i > SlotFree || u.Cell < 0 || u.Cell >= width*height || u.Type < 0 || u.Type >= len(t.Types) {
			return nil, fmt.Errorf("dosrules: unit %d: %+v", i, u)
		}
		typ := t.Types[u.Type]
		size := u.Size
		if size == 0 {
			size = p.Tables.Types[u.Type].Group
		}
		l := Live{Budget: typ.Budget, Exp: u.Exp, Size: size, Alt: u.Alt, Flags4: typ.Class | [3]int{0, 1, 2}[u.Owner],
			Flags6: typ.Flags, Type: u.Type, Hold: 0xFF}
		for k := range 2 {
			l.Pos[k] = 2*u.Cell + 1
			g.Maps[k][2*u.Cell+1] = i
		}
		g.Units[i] = l
	}
	for i, h := range holds {
		g.Holds[i] = h
	}
	g.Players[0].Phase, g.Players[1].Phase = PhaseMovement, PhaseAction
	return g, nil
}

// CommandKind names an order.
type CommandKind string

// The orders a side gives.
const (
	CmdMove   CommandKind = "move"   // Unit to Cell
	CmdExit   CommandKind = "exit"   // the unit in Slot of Hold to Cell
	CmdAttack CommandKind = "attack" // Unit against Cell
)

// Command is an order of Side.
type Command struct {
	Kind CommandKind `json:"kind"`
	Side int         `json:"side"`
	Unit int         `json:"unit,omitempty"`
	Hold int         `json:"hold,omitempty"`
	Slot int         `json:"slot,omitempty"`
	Cell int         `json:"cell"`
}

// Refusal is a command the cycle does not take. Code is the status message
// the original shows, 0 where it shows none (the menu does not offer it).
type Refusal struct {
	Code   int
	Reason string
}

func (r *Refusal) Error() string {
	if r.Code != 0 {
		return fmt.Sprintf("dosrules: %s (message %#x)", r.Reason, r.Code)
	}
	return "dosrules: " + r.Reason
}

func refuse(code int, reason string) error { return &Refusal{code, reason} }

// ErrNotReady is returned by Resolve while a side has not confirmed.
var ErrNotReady = errors.New("dosrules: both sides must confirm before the orders resolve")

// EventKind names an event.
type EventKind string

// The events of the cycle.
const (
	EvMoved   EventKind = "moved"   // a move or an exit ran: Route
	EvRefused EventKind = "refused" // the order code or the move start refused it: Messages
	EvAttack  EventKind = "attack"  // an attack order entered the list
	EvReady   EventKind = "ready"
	EvUnready EventKind = "unready"
	EvBattle  EventKind = "battle"  // an attack order resolved: Outcome
	EvSkipped EventKind = "skipped" // an attack order whose cell held no unit
	EvPhase   EventKind = "phase"   // the phase change: Phase, Turn
)

// Event is one thing that happened, in order.
type Event struct {
	Kind     EventKind `json:"kind"`
	Side     int       `json:"side"`
	Unit     int       `json:"unit"`
	Cell     int       `json:"cell"`
	Route    []int     `json:"route,omitempty"`
	Outcome  *Outcome  `json:"outcome,omitempty"`
	Messages []Message `json:"messages,omitempty"`
	Phase    [2]int    `json:"phase"`
	Turn     int       `json:"turn"`
}

// unitOf is the unit a command orders.
func (g *Game) unitOf(c Command) (int, error) {
	if c.Kind != CmdExit {
		return c.Unit, nil
	}
	h, ok := g.Holds[c.Hold]
	if !ok || c.Slot < 0 || c.Slot >= 7 {
		return 0, refuse(0, "no such slot")
	}
	u := h.Slots[7*c.Side+c.Slot]
	if u > SlotFree {
		return 0, refuse(0, "the slot is empty")
	}
	return u, nil
}

// Check is the validation of a command, without changing the state.
//   - The side must not have confirmed (0661:3b81: a confirmed side's input
//     only cancels the confirmation).
//   - The unit must be one of the side's (0ca0:020c), not neutral (live +4
//     bit 2) and without the moved flag (0x200): one order per unit and
//     phase (0ca0:0f59).
//   - Moves and exits belong to the movement phase, attacks to the action
//     phase (0ca0:0f59: the move symbol in phase 1, the fist in phase 2).
//   - A human side's list must be shorter than the move limit (0ca0:1757,
//     0661:3349, message 0x53); a computer side has none.
//   - An attack needs a unit in the target cell of the side's map (0ad1:000c,
//     message 0x45) that the side does not own.
//
// The move order code refuses a destination itself (EvRefused), as it does
// in the game.
func (g *Game) Check(c Command) error {
	if c.Side != 0 && c.Side != 1 {
		return fmt.Errorf("dosrules: side %d", c.Side)
	}
	p := &g.Players[c.Side]
	if p.Ready {
		return refuse(0, "the side has confirmed its orders")
	}
	idx, err := g.unitOf(c)
	if err != nil {
		return err
	}
	u, ok := g.Units[idx]
	switch {
	case !ok:
		return refuse(0, "no such unit")
	case u.Flags4&2 != 0 || !Owns(u.Flags4, c.Side):
		return refuse(0, "not the side's unit")
	case u.Flags4&0x200 != 0:
		return refuse(0, "the unit has its order")
	}
	switch c.Kind {
	case CmdMove, CmdExit:
		if p.Phase != PhaseMovement {
			return refuse(0, "moves are given in the movement phase")
		}
		if c.Kind == CmdMove && (u.Flags4&0x4000 != 0 || g.occupant(c.Side, u.Cell(c.Side)) != idx) {
			return refuse(0, "the unit is not on the map")
		}
	case CmdAttack:
		if p.Phase != PhaseAction {
			return refuse(0, "attacks are given in the action phase")
		}
	default:
		return fmt.Errorf("dosrules: command %q", c.Kind)
	}
	if c.Cell < 0 || c.Cell >= g.Width*g.Height {
		return fmt.Errorf("dosrules: cell %d", c.Cell)
	}
	if g.Limit > 0 && !p.Computer && len(p.Orders) >= g.Limit {
		return refuse(MsgLimit, "the move limit is reached")
	}
	if c.Kind == CmdAttack {
		t := g.occupant(c.Side, c.Cell)
		if v, ok := g.Units[t]; t > SlotFree || !ok || Owns(v.Flags4, c.Side) {
			return refuse(MsgNoTarget, "no target in the cell")
		}
	}
	return nil
}

// occupant is the occupant byte of a cell of side's map.
func (g *Game) occupant(side, cell int) int {
	if cell < 0 || cell >= g.Width*g.Height {
		return 0xFF
	}
	return g.Maps[side][2*cell+1]
}

// Apply checks a command and carries it out.
//   - A move or an exit runs at once, as in the game, where the unit moves
//     during the phase while no other move may start (4c9e bit 0x80): the
//     order code, the move start and the steps on the side's map
//     (MoveTables.Execute, Exit). The move start appends the unit to the
//     side's list.
//   - An attack order (0ad1:000c, the attack kind) sets the moved flag (live
//     +4 0x200), +0x11 to the target, +0x13 = 9 and +0x14 = 2, and appends the
//     unit to the list (message 0x49).
//
// Nothing withdraws an order once given: the cancel (10cc:0001) only clears
// the reach and selection marks and undoes an exit in progress.
func (g *Game) Apply(c Command) ([]Event, error) {
	if err := g.Check(c); err != nil {
		return nil, err
	}
	if c.Kind == CmdAttack {
		u := g.Units[c.Unit]
		u.Flags4 |= 0x200
		u.Target = 2 * c.Cell
		u.Action[0], u.Action[1] = 9, 2
		g.Units[c.Unit] = u
		p := &g.Players[c.Side]
		p.Orders = append(p.Orders, c.Unit)
		return []Event{{Kind: EvAttack, Side: c.Side, Unit: c.Unit, Cell: c.Cell,
			Messages: []Message{{MsgAttack, c.Side}}}}, nil
	}
	return g.move(c)
}

// movement is the movement query of unit for side on the game's state: the
// side's map with its units, the units held in containers, the budget of
// live byte 0 and the flag mask the player's order code passes (0xffff for
// live +6 bit 1, else 6 for a side-1 unit and 3 for others).
func (g *Game) movement(side, unit int) *Movement {
	m := &Movement{Width: g.Width, Height: g.Height, Terrain: make([]int, g.Width*g.Height),
		Units: map[int]Piece{}, Held: map[int]Piece{}, Holds: g.Holds, Unit: unit, Side: side}
	piece := func(u Live, cell int) Piece {
		return Piece{Type: u.Type, Cell: cell, Flags4: u.Flags4, Flags6: u.Flags6, Hold: u.Hold}
	}
	for c := range m.Terrain {
		m.Terrain[c] = g.Maps[side][2*c]
		if occ := g.Maps[side][2*c+1]; occ <= SlotFree {
			if u, ok := g.Units[occ]; ok {
				m.Units[occ] = piece(u, c)
			}
		}
	}
	for i, u := range g.Units {
		if u.Flags4&0x4000 != 0 {
			m.Held[i] = piece(u, u.Cell(side))
		}
	}
	u := g.Units[unit]
	m.Budget, m.Block = u.Budget, 3
	switch {
	case u.Flags6&1 != 0:
		m.Block = 0xFFFF
	case u.Flags4&1 != 0:
		m.Block = 6
	}
	return m
}

func bytesOf(v []int) []byte {
	b := make([]byte, len(v))
	for i, x := range v {
		b[i] = byte(x)
	}
	return b
}

// move runs a checked move or exit on the game's records and keeps what it
// leaves: the live records, the side's map, the container records and the
// order list.
func (g *Game) move(c Command) ([]Event, error) {
	t, err := g.Profile.moveTables()
	if err != nil {
		return nil, err
	}
	unit, err := g.unitOf(c)
	if err != nil {
		return nil, err
	}
	live := map[int][]byte{}
	for i, u := range g.Units {
		live[i] = u.Bytes()
	}
	s := moveStateOn(t, g.movement(c.Side, unit), live,
		[2][]byte{bytesOf(g.Maps[c.Side]), bytesOf(g.Maps[1-c.Side])}, ContainerRecords(g.Holds))
	o := MoveOrder{Goal: c.Cell, Retarget: -1}
	var r StepRun
	if c.Kind == CmdExit {
		var x ExitRun
		x, err = t.exit(s, ExitOrder{Hold: c.Hold, Slot: c.Slot, MoveOrder: o})
		r = x.StepRun
	} else {
		r, err = t.execute(s, o)
	}
	if err != nil {
		return nil, err
	}
	for i := range g.Units {
		g.Units[i] = DecodeLive(s.live[i])
	}
	for i, b := range s.maps[0] {
		g.Maps[c.Side][i] = int(b)
	}
	empty := EmptyHold()
	for i, b := range s.holds {
		if _, had := g.Holds[i]; had || DecodeHold(b) != empty {
			h := DecodeHold(b)
			g.Holds[i] = h
		}
	}
	p := &g.Players[c.Side]
	p.Orders = append(p.Orders, s.orders...)
	ev := Event{Kind: EvMoved, Side: c.Side, Unit: unit, Cell: c.Cell, Route: r.Cells, Messages: messages(r.Messages)}
	if r.Order&1 == 0 || r.Start&1 == 0 {
		ev.Kind = EvRefused
	}
	return []Event{ev}, nil
}

// Confirm is the double-arrow symbol: the side's orders are complete
// (+0x17 = 5, 0ca0:16b6).
func (g *Game) Confirm(side int) []Event {
	g.Players[side].Ready = true
	return []Event{{Kind: EvReady, Side: side}}
}

// Unconfirm is the fire button on a confirmed side (0661:3b81): the side may
// give orders again (message 0x34).
func (g *Game) Unconfirm(side int) []Event {
	if !g.Players[side].Ready {
		return nil
	}
	g.Players[side].Ready = false
	return []Event{{Kind: EvUnready, Side: side, Messages: []Message{{MsgReadyUndone, side}}}}
}

// Action is the side in the action phase.
func (g *Game) Action() (int, error) {
	switch {
	case g.Players[0].Phase == PhaseAction && g.Players[1].Phase == PhaseMovement:
		return 0, nil
	case g.Players[1].Phase == PhaseAction && g.Players[0].Phase == PhaseMovement:
		return 1, nil
	}
	return 0, fmt.Errorf("dosrules: phases %d/%d", g.Players[0].Phase, g.Players[1].Phase)
}

// Battlefield is the action side's map as the order handler fights on it:
// the units its map shows and the combat class of every cell's terrain
// record (byte 5).
func (g *Game) Battlefield(side int) (*Battlefield, error) {
	t, err := g.Profile.moveTables()
	if err != nil {
		return nil, err
	}
	mp := g.Maps[side]
	b := &Battlefield{Width: g.Width, Height: g.Height, Units: map[int]Unit{}, Terrain: map[int]int{}, MapByte0: mp[0]}
	for c := range g.Width * g.Height {
		if tr := mp[2*c]; tr < len(t.Terrains) {
			b.Terrain[c] = t.Terrains[tr].Class
		}
		if occ := mp[2*c+1]; occ <= SlotFree {
			if u, ok := g.Units[occ]; ok {
				b.Units[occ] = Unit{Type: u.Type, Cell: c, Size: u.Size, Alt: u.Alt, Exp: u.Exp, Flags4: u.Flags4, Flags6: u.Flags6}
			}
		}
	}
	return b, nil
}

// Resolve is the order handler once both sides have confirmed (F1, or two
// computer sides, 0661:3cfb). The action side's list resolves on its map, in
// list order (ResolveOrders): each attack against the cell in the unit's
// +0x11, with the strengths and experience it leaves, and every destroyed
// unit taken off that map with bit 4 of its +0x14 set (02dc:2083). Then the
// phase change (ChangePhase) brings both maps into line, marks the
// destroyed units 0x8000, swaps the phases, clears both lists and the
// confirmations and counts the turn. It also clears every unit's moved flag.
// A destroyed attacker cannot have a later order: it is listed once, and
// only return fire in its own battle destroys it.
func (g *Game) Resolve() ([]Event, error) {
	if !g.Players[0].Ready || !g.Players[1].Ready {
		return nil, ErrNotReady
	}
	a, err := g.Action()
	if err != nil {
		return nil, err
	}
	evs, err := g.attack(a)
	if err != nil {
		return evs, err
	}
	ev, err := g.changePhase(a)
	return append(evs, ev...), err
}

// attack resolves side a's list on its map.
func (g *Game) attack(a int) ([]Event, error) {
	b, err := g.Battlefield(a)
	if err != nil {
		return nil, err
	}
	var orders []Order
	for _, u := range g.Players[a].Orders {
		orders = append(orders, Order{u, g.Units[u].Target / 2})
	}
	r, err := g.Profile.ResolveOrders(b, orders)
	if err != nil {
		return nil, err
	}
	var evs []Event
	for i, o := range r.Orders {
		ev := Event{Kind: EvSkipped, Side: a, Unit: orders[i].Attacker, Cell: orders[i].Target}
		if o.Battle {
			ev.Kind, ev.Outcome = EvBattle, &r.Orders[i]
		}
		evs = append(evs, ev)
	}
	for i, u := range r.Units {
		l := g.Units[i]
		l.Size, l.Alt, l.Exp = u.Size, u.Alt, u.Exp
		g.Units[i] = l
	}
	for _, i := range r.Gone {
		l := g.Units[i]
		l.Action[1] |= 4
		g.Units[i] = l
		for c := range g.Width * g.Height {
			if g.Maps[a][2*c+1] == i {
				g.Maps[a][2*c+1] = 0xFF
			}
		}
	}
	return evs, nil
}

// changePhase is the phase change after side a's attacks.
func (g *Game) changePhase(a int) ([]Event, error) {
	pc := &PhaseChange{Width: g.Width, Height: g.Height, Maps: g.Maps, Units: map[int]Sighted{}, Holds: g.Holds,
		Attacker: a, Phase: [2]int{g.Players[0].Phase, g.Players[1].Phase}, Count: g.Turn}
	for i, u := range g.Units {
		pc.Units[i] = Sighted{Flags4: u.Flags4, Size: u.Size, Pos: u.Pos, Facing: u.Facing, Hold: u.Hold}
	}
	pr, err := ChangePhase(pc)
	if err != nil {
		return nil, err
	}
	for i := range g.Units {
		d := DecodeLive(pr.Live[i])
		l := g.Units[i]
		l.Flags4, l.Size, l.Pos, l.Facing = d.Flags4&^0x200, d.Size, d.Pos, d.Facing
		g.Units[i] = l
	}
	for k := range 2 {
		g.Maps[k] = make([]int, len(pr.Maps[k]))
		for i, v := range pr.Maps[k] {
			g.Maps[k][i] = int(v)
		}
	}
	for i, h := range changed(ContainerRecords(g.Holds), pr.Holds) {
		g.Holds[i] = h
	}
	for k := range 2 {
		g.Players[k].Phase = int(pr.Players[k][0x16])
		g.Players[k].Ready = false
		g.Players[k].Orders = nil
	}
	g.Turn = pr.Count
	return []Event{{Kind: EvPhase, Side: a, Messages: messages(pr.Messages),
		Phase: [2]int{g.Players[0].Phase, g.Players[1].Phase}, Turn: g.Turn}}, nil
}

// Destroyed are the units the last resolutions destroyed and the phase
// change marked, ascending.
func (g *Game) Destroyed() []int {
	var out []int
	for i, u := range g.Units {
		if u.Destroyed() {
			out = append(out, i)
		}
	}
	slices.Sort(out)
	return out
}
