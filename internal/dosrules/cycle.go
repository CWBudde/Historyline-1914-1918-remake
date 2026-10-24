package dosrules

import (
	"errors"
	"fmt"
	"maps"
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
// the players, the turn counter (32e2:4cab), the move limit (the option
// 1653:253c: 4, 8 or 16 orders per side and phase; 0 none), the option
// "with concealed inventories", which counts for the score, the battle's end
// once a resolution has decided it (nil while it is open), each side's energy
// account and the scenario's buildings by container record (Energy).
type Game struct {
	Profile   Profile
	Width     int
	Height    int
	Maps      [2][]int
	Units     map[int]Live
	Holds     map[int]Hold
	Players   [2]Player
	Turn      int
	Limit     int
	Concealed bool
	Ended     *End
	Energy    [2]int
	Buildings map[int]Building
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
//
// The container state follows from holds, as an arrival leaves it:
//   - A unit in a slot of a record is held (+4 0x4000), its position on both
//     sides is the record's first cell and no map shows it; its Cell is read
//     only to find a transporter's own record. A building record (HoldFlag40 and up) with a unit in a side's
//     slots shows InBuilding on that side's map at its first cell where no
//     unit stands.
//   - A transporter (+4 0x1000) gets as +0x0a the lowest record of flag
//     0x20 whose first cell is its own (2*Cell+1) that no transporter of a
//     lower index took; without one NewGame fails.
//   - A construction unit (Profile.Builders) starts with 2 charges in +0x0a.
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
	recs := slices.Sorted(maps.Keys(holds))
	in := map[int]int{} // held unit -> its record
	for _, r := range recs {
		h := holds[r]
		if r < 0 || r >= HoldRecords {
			return nil, fmt.Errorf("dosrules: container record %d", r)
		}
		for _, u := range h.Slots {
			if u > SlotFree {
				continue
			}
			if _, ok := units[u]; !ok {
				return nil, fmt.Errorf("dosrules: container record %d holds unit %d, which is not placed", r, u)
			}
			if o, ok := in[u]; ok && o != r {
				return nil, fmt.Errorf("dosrules: unit %d is held in records %d and %d", u, o, r)
			}
			in[u] = r
		}
		g.Holds[r] = h
	}
	taken := map[int]bool{}
	for _, i := range slices.Sorted(maps.Keys(units)) {
		u := units[i]
		r, held := in[i]
		if i < 0 || i > SlotFree || !held && (u.Cell < 0 || u.Cell >= width*height) || u.Type < 0 || u.Type >= len(t.Types) ||
			u.Owner < 0 || u.Owner > 2 {
			return nil, fmt.Errorf("dosrules: unit %d: %+v", i, u)
		}
		typ := t.Types[u.Type]
		size := u.Size
		if size == 0 {
			size = p.Tables.Types[u.Type].Group
		}
		l := Live{Budget: typ.Budget, Exp: u.Exp, Size: size, Alt: u.Alt, Flags4: typ.Class | [3]int{0, 1, 2}[u.Owner],
			Flags6: typ.Flags, Type: u.Type, Hold: 0xFF}
		if p.builder(u.Type) {
			l.Hold = builderCharges
		}
		if l.Flags4&0x1000 != 0 {
			l.Hold = -1
			for _, k := range recs {
				if h := holds[k]; !taken[k] && h.Flags&0x20 != 0 && h.Cells[0] == 2*u.Cell+1 {
					l.Hold, taken[k] = k, true
					break
				}
			}
			if l.Hold < 0 {
				return nil, fmt.Errorf("dosrules: transporter %d at cell %d has no container record", i, u.Cell)
			}
		}
		for k := range 2 {
			if held {
				l.Pos[k] = holds[r].Cells[0]
				continue
			}
			l.Pos[k] = 2*u.Cell + 1
			g.Maps[k][2*u.Cell+1] = i
		}
		if held {
			l.Flags4 |= 0x4000
		}
		g.Units[i] = l
	}
	for _, r := range recs {
		h := holds[r]
		at := h.Cells[0] + 1
		if r < HoldFlag40 || at < 1 || at >= 2*width*height {
			continue
		}
		for k := range 2 {
			if slices.ContainsFunc(h.Slots[7*k:7*k+7], func(u int) bool { return u <= SlotFree }) && g.Maps[k][at] == 0xFF {
				g.Maps[k][at] = InBuilding
			}
		}
	}
	g.markKinds()
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
	CmdRepair CommandKind = "repair" // the unit in Slot of building Hold
	CmdBuild  CommandKind = "build"  // a unit of Type in building Hold
)

// Command is an order of Side.
type Command struct {
	Kind CommandKind `json:"kind"`
	Side int         `json:"side"`
	Unit int         `json:"unit,omitempty"`
	Hold int         `json:"hold,omitempty"`
	Slot int         `json:"slot,omitempty"`
	Cell int         `json:"cell"`
	Type int         `json:"type,omitempty"`
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

// ErrOver is returned by Resolve once the battle has ended.
var ErrOver = errors.New("dosrules: the battle is over")

// EventKind names an event.
type EventKind string

// The events of the cycle.
const (
	EvMoved    EventKind = "moved"   // a move or an exit ran: Route
	EvRefused  EventKind = "refused" // the order code or the move start refused it: Messages
	EvAttack   EventKind = "attack"  // an attack order entered the list
	EvReady    EventKind = "ready"
	EvUnready  EventKind = "unready"
	EvBattle   EventKind = "battle"   // an attack order resolved: Outcome
	EvSkipped  EventKind = "skipped"  // an attack order whose cell held no unit
	EvPhase    EventKind = "phase"    // the phase change: Phase, Turn
	EvEnd      EventKind = "end"      // the battle ended: Side is the winner (-1 a draw), End
	EvRepaired EventKind = "repaired" // a unit repaired: Energy left
	EvBuilt    EventKind = "built"    // a unit built: Unit, the building's Cell, Energy left
	EvCredit   EventKind = "credit"   // a side credited after the phase change: Energy
	EvCaptured EventKind = "captured" // Side took building record Hold at Cell with Unit (capture.go)
	EvWork     EventKind = "work"     // a sapper's work order on Cell was taken (sapper.go): Messages
	EvDug      EventKind = "dug"      // a sapper dug a trench at Cell: Cells took trench records (sapper.go)
	EvFilled   EventKind = "filled"   // a sapper filled a trench at Cell: Cells took TerrainFilled (sapper.go)
)

// Event is one thing that happened, in order.
type Event struct {
	Kind     EventKind `json:"kind"`
	Side     int       `json:"side"`
	Unit     int       `json:"unit"`
	Cell     int       `json:"cell"`
	Route    []int     `json:"route,omitempty"`
	Outcome  *Outcome  `json:"outcome,omitempty"`
	End      *End      `json:"end,omitempty"`
	Messages []Message `json:"messages,omitempty"`
	Phase    [2]int    `json:"phase"`
	Turn     int       `json:"turn"`
	Energy   int       `json:"energy,omitempty"`
	Hold     int       `json:"hold,omitempty"`
	Cells    []int     `json:"cells,omitempty"`
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
//     message 0x45) that is a target of the unit (Target). Assumed, not
//     verified: as for a move, the attacker must be on the side's map, not
//     held in a container and the occupant of its cell. A sapper's order
//     needs a neighbouring cell it can dig or fill instead (sapper.go).
//   - A repair or a build needs an own building, the action phase and the
//     energy (energy.go).
//
// The move order code refuses a destination itself (EvRefused), as it does
// in the game.
func (g *Game) Check(c Command) error {
	if c.Side != 0 && c.Side != 1 {
		return fmt.Errorf("dosrules: side %d", c.Side)
	}
	p := &g.Players[c.Side]
	if g.Ended != nil {
		return refuse(0, "the battle is over")
	}
	if p.Ready {
		return refuse(0, "the side has confirmed its orders")
	}
	if c.Kind == CmdRepair || c.Kind == CmdBuild {
		return g.checkBuilding(c)
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
	case u.Flags4&0xC0 == 0x40:
		// A two-cell ship's rear (+4 0x40 without the head's 0x80) moves and
		// fights with its head.
		return refuse(0, "the ship's rear follows its head")
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
		if u.Flags4&0x4000 != 0 || g.occupant(c.Side, u.Cell(c.Side)) != idx {
			return refuse(0, "the unit is not on the map")
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
	if c.Kind == CmdAttack && u.Sapper() {
		if !g.trenchTarget(c.Side, u, c.Cell) {
			return refuse(MsgNoTarget, "no trench work there")
		}
	} else if c.Kind == CmdAttack {
		t := g.occupant(c.Side, c.Cell)
		if v, ok := g.Units[t]; t > SlotFree || !ok || !g.Target(c.Side, u, v) {
			return refuse(MsgNoTarget, "no target in the cell")
		}
	}
	return nil
}

// Target reports whether the unit a of side may attack the unit v where the
// side's map shows them. This is an assumed rule, not a verified one: v is
// not the side's, a's target flags cover a class of v's live word, and v lies
// within the range of a's weapon against that class (the air range against
// aircraft, the surface range otherwise; a range byte is one more than the
// distance it reaches, so 1 reaches nothing).
func (g *Game) Target(side int, a, v Live) bool {
	if Owns(v.Flags4, side) {
		return false
	}
	t := g.Profile.Tables.Types[a.Type]
	if t.TargetFlags&v.Flags4&classBits == 0 {
		return false
	}
	_, rng := weapon(t, v.Flags4)
	d := g.Profile.LayoutDistance(g.Width, a.Cell(side), v.Cell(side))
	return d >= 1 && int64(d) < rng
}

// WeaponAgainst is the weapon of type typ against a unit with the live word
// flags4: the air power and range against aircraft, the water power and
// the surface range against ships, the ground power and the surface range
// otherwise (a range is encoded, one more than the distance it reaches).
func (p Profile) WeaponAgainst(typ, flags4 int) (power, rng int) {
	pw, r := weapon(p.Tables.Types[typ], flags4)
	return int(pw), int(r)
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
//     unit to the list (message 0x49). A sapper's order sets the moved flag,
//     +0x11 and +0x13 = 0x11 and stays off the list (message 0x6e,
//     EvWork).
//
// A repair or a build is carried out at once (energy.go).
//
// Nothing withdraws an order once given: the cancel (10cc:0001) only clears
// the reach and selection marks and undoes an exit in progress.
func (g *Game) Apply(c Command) ([]Event, error) {
	if err := g.Check(c); err != nil {
		return nil, err
	}
	switch c.Kind {
	case CmdRepair:
		return g.repair(c), nil
	case CmdBuild:
		return g.build(c), nil
	}
	if c.Kind == CmdAttack && g.Units[c.Unit].Sapper() {
		u := g.Units[c.Unit]
		u.Flags4 |= 0x200
		u.Target = 2 * c.Cell
		u.Action[0] = actionTrench
		g.Units[c.Unit] = u
		return []Event{{Kind: EvWork, Side: c.Side, Unit: c.Unit, Cell: c.Cell,
			Messages: []Message{{MsgTrench, c.Side}}}}, nil
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

// Movement is the movement query of unit for side on the game's state: the
// side's map with its units, the units held in containers, the budget of
// live byte 0 and the flag mask the player's order code passes (0xffff for
// live +6 bit 1, else 6 for a side-1 unit and 3 for others).
func (g *Game) Movement(side, unit int) *Movement {
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

// ExitMovement is the movement query of the exit from slot of side's part
// of container record hold, as the exit runs its reach search (Exit): the
// unit in the slot at the record's first cell (one less for a transporter,
// flag 0x20), budget 8 (4 for live +6 bit 0x10) and the flag mask 3 (6 for
// live +4 bit 1, with 0x18 for live +6 bit 4). The unit is not on the
// side's map yet; its start cell is among the reached cells, and the order
// code refuses it as a destination.
func (g *Game) ExitMovement(side, hold, slot int) (*Movement, error) {
	h, ok := g.Holds[hold]
	if side != 0 && side != 1 || !ok || slot < 0 || slot >= 7 {
		return nil, fmt.Errorf("dosrules: no slot %d of side %d in container record %d", slot, side, hold)
	}
	u := h.Slots[7*side+slot]
	if u > SlotFree {
		return nil, refuse(0, "the slot is empty")
	}
	m, _, _, err := exitQuery(g.Movement(side, u), h.Bytes(), slot, func(i int) []byte { return g.Units[i].Bytes() })
	return m, err
}

// HoldAt is the container record of a cell of side's map: the occupant's
// +0x0a where it is a transporter (+4 0x1000), else the building record of
// the cell's terrain flag as the end-cell check finds it (0x400 and 0x100
// by the cells the records list, 0x40 by flag bit 1). It reports false for
// a cell without one.
func (g *Game) HoldAt(side, cell int) (int, bool) {
	t, err := g.Profile.moveTables()
	if err != nil || side != 0 && side != 1 || cell < 0 || cell >= g.Width*g.Height {
		return 0, false
	}
	if u, ok := g.Units[g.occupant(side, cell)]; ok && u.Flags4&0x1000 != 0 {
		return u.Hold, u.Hold >= 0 && u.Hold < HoldRecords
	}
	tr := g.Maps[side][2*cell]
	if tr < 0 || tr >= len(t.Terrains) {
		return 0, false
	}
	find := func(base int) (int, bool) {
		for i := range holdTable {
			h, ok := g.Holds[base+i]
			if !ok {
				h = EmptyHold()
			}
			if slices.Contains(h.Cells[:], 2*cell) {
				return base + i, true
			}
		}
		return 0, false
	}
	switch f := t.Terrains[tr].Flags; {
	case f&0x400 != 0:
		return find(HoldFlag400)
	case f&0x100 != 0:
		return find(HoldFlag100)
	case f&0x40 != 0:
		return HoldFlag40 + f&1, true
	}
	return 0, false
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
	s := moveStateOn(t, g.Movement(c.Side, unit), live,
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
// unit taken off that map with bit 4 of its +0x14 set (02dc:2083). The
// action pass then hands the buildings the movement side's units entered to
// that side (capture, EvCaptured) and carries out the sappers' work on the
// action side's map (trench, EvDug and EvFilled). Then the phase change (ChangePhase)
// brings both maps into line, marks the destroyed units 0x8000, swaps the
// phases, clears both lists and the confirmations and counts the turn. It
// also clears every unit's +4 bits 0x0f00 (moved among them), its order cell
// +0x11 and its action bytes +0x13 to +0x16 (02dc:1cf0, 1ebc), except that a
// unit that must set up (type +0x0f bit 0: medium, heavy and train artillery)
// and arrived on the map this phase (+0x14 bit 0), not destroyed and not held,
// keeps the moved flag: it gives no order in the next phase (02dc:1cf9-1d60).
// Each side is then credited its buildings' energy (EvCredit).
// A destroyed attacker cannot have a later order: it is listed once, and
// only return fire in its own battle destroys it. When the battle has ended
// (End), an EvEnd event follows the phase change and no further command or
// resolution is taken.
func (g *Game) Resolve() ([]Event, error) {
	if g.Ended != nil {
		return nil, ErrOver
	}
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
	caps, hq := g.work(a)
	evs = append(evs, caps...)
	ev, err := g.changePhase(a)
	evs = append(evs, ev...)
	if err != nil {
		return evs, err
	}
	evs = append(evs, g.credit(ev[0])...)
	if e := g.end(hq); e.Over {
		g.Ended = &e
		evs = append(evs, Event{Kind: EvEnd, Side: e.Winner, End: &e, Phase: ev[0].Phase, Turn: g.Turn})
	}
	return evs, nil
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
	for i, u := range r.Units {
		// A ship half whose other half that map does not hold: it is one unit.
		p := i - 1
		if u.Flags4&0x80 != 0 {
			p = i + 1
		}
		if _, seen := r.Units[p]; u.Flags4&0x40 == 0 || seen {
			continue
		}
		if l, ok := g.Units[p]; ok && l.Flags4&0x40 != 0 {
			l.Size, l.Alt, l.Exp = u.Size, u.Alt, u.Exp
			g.Units[p] = l
			if slices.Contains(r.Gone, i) {
				r.Gone = append(r.Gone, p)
			}
		}
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
		l.Flags4, l.Size, l.Pos, l.Facing = d.Flags4&0xF0FF, d.Size, d.Pos, d.Facing
		if l.Flags4&0xC000 == 0 && l.Action[1]&1 != 0 && g.Profile.Tables.Types[l.Type].setsUp() {
			l.Flags4 |= 0x200 // 02dc:1cf9-1d60
		}
		l.Target, l.Action[0], l.Action[1], l.Action[2], l.Action[3] = 0, 0, 0, 0, 0
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
