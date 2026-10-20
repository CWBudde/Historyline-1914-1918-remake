package dosrules

import (
	"bytes"
	"errors"
	"slices"
)

// The typed view of the movement, container and visibility rules for the
// command cycle: the record-level functions run as executed, and their
// records are read into Live and Hold values.

// ErrNoMovement is returned where the profile's build has no movement tables.
var ErrNoMovement = errors.New("dosrules: the profile has no movement tables")

// Live is a 26-byte live unit record (Build.Units in the demo).
type Live struct {
	Budget int // +0: the movement budget, from type byte 0
	Exp    int // +1: experience, 0-6
	Size   int // +2: group size; 0 marks the unit destroyed
	Alt    int // +3: the alternate strength
	Flags4 int // +4 word: side bits 1/2 (Owns), the type's class word +0x0d, 0x4000 in a container, 0x8000 no unit
	Flags6 int // +6 word: the type's word +0x11
	Type   int // +8: the type record
	Serial int // +9: the per-side serial number
	// Hold is +0x0a: a transporter's container record, a construction
	// unit's builder charges, 0xff otherwise.
	Hold int
	// Pos is +0x0b + 2*side: the map offset 2*cell+1 of the occupant byte
	// where each side sees the unit (Cell).
	Pos    [2]int
	Facing [2]int // +0x0f + side: the direction each side sees it face
	Target int    // +0x11 word: the order's destination 2*cell, then the start's position
	Action [7]int // +0x13 to +0x19: the action bytes
}

// DecodeLive reads a live record.
func DecodeLive(b []byte) Live {
	u := Live{Budget: int(b[0]), Exp: int(b[1]), Size: int(b[2]), Alt: int(b[3]),
		Flags4: word(b, 4), Flags6: word(b, 6), Type: int(b[8]), Serial: int(b[9]), Hold: int(b[0x0A]),
		Target: word(b, 0x11)}
	for k := range 2 {
		u.Pos[k], u.Facing[k] = word(b, 0x0B+2*k), int(b[0x0F+k])
	}
	for i := range u.Action {
		u.Action[i] = int(b[0x13+i])
	}
	return u
}

// Bytes is the live record.
func (u Live) Bytes() []byte {
	b := make([]byte, LiveLen)
	b[0], b[1], b[2], b[3] = byte(u.Budget), byte(u.Exp), byte(u.Size), byte(u.Alt)
	putWord(b, 4, u.Flags4)
	putWord(b, 6, u.Flags6)
	b[8], b[9], b[0x0A] = byte(u.Type), byte(u.Serial), byte(u.Hold)
	for k := range 2 {
		putWord(b, 0x0B+2*k, u.Pos[k])
		b[0x0F+k] = byte(u.Facing[k])
	}
	putWord(b, 0x11, u.Target)
	for i, v := range u.Action {
		b[0x13+i] = byte(v)
	}
	return b
}

// Cell is the cell where side sees the unit.
func (u Live) Cell(side int) int { return (u.Pos[side] - 1) / 2 }

// Destroyed reports whether the record holds no unit (+4 bit 0x8000) or a
// destroyed one (size 0).
func (u Live) Destroyed() bool { return u.Flags4&0x8000 != 0 || u.Size == 0 }

// DecodeHold reads a container record (Hold.Bytes). It drops the bytes Hold
// does not name, +0x16 to +0x18 and +0x1a.
func DecodeHold(b []byte) Hold {
	var h Hold
	for i := range h.Slots {
		h.Slots[i] = int(b[i])
	}
	for k := range h.Cells {
		h.Cells[k] = word(b, 0x0E+2*k)
	}
	h.Flags, h.Occupant = word(b, 0x19), int(b[0x1B])
	return h
}

// Message is a status message the rules show to a side.
type Message struct{ Code, Side int }

func messages(ms [][2]int) []Message {
	out := make([]Message, len(ms))
	for i, m := range ms {
		out[i] = Message{m[0], m[1]}
	}
	return out
}

func lives(recs map[int][]byte) map[int]Live {
	out := make(map[int]Live, len(recs))
	for i, b := range recs {
		out[i] = DecodeLive(b)
	}
	return out
}

// changed are the container records that differ from before.
func changed(before, after map[int][]byte) map[int]Hold {
	out := map[int]Hold{}
	for i, b := range after {
		if !bytes.Equal(before[i], b) {
			out[i] = DecodeHold(b)
		}
	}
	return out
}

// occupants are the occupied cells of a map of two bytes per cell.
func occupants(m []byte) map[int]int {
	out := map[int]int{}
	for c := range len(m) / 2 {
		if v := m[2*c+1]; v != 0xFF {
			out[c] = int(v)
		}
	}
	return out
}

// Moved is the outcome of a move order or an exit.
type Moved struct {
	// Refused: the order code or the move start refused the order; the
	// unit stays.
	Refused bool
	TwoCell bool // the mover is a ship two cells long
	// Route is the mover's cell, as its side sees it, after each step.
	Route    []int
	Messages []Message
	Units    map[int]Live // the record of every placed unit afterwards
	Holds    map[int]Hold // the container records that changed
	// Occupants are the occupied cells of the mover's side's map and of
	// the other side's map afterwards.
	Occupants [2]map[int]int
	// Remaining is, for an exit, the budget left on every cell its reach
	// search reached.
	Remaining map[int]int
}

// Season is the profile with the season's terrain records; in winter
// without a winter table its movement methods return ErrNoMovement.
func (p Profile) Season(winter bool) Profile {
	if !winter || p.Tables == nil || p.Tables.Move == nil {
		return p
	}
	t := *p.Tables
	t.Move = t.Move.Season(true)
	p.Tables = &t
	return p
}

func (p Profile) moveTables() (*MoveTables, error) {
	if p.Tables == nil || p.Tables.Move == nil {
		return nil, ErrNoMovement
	}
	return p.Tables.Move, nil
}

func movedOf(g *Movement, r StepRun) Moved {
	m := Moved{Refused: r.Order&1 == 0 || r.Start&1 == 0, TwoCell: r.Order&8 != 0,
		Route: append([]int{}, r.Cells...), Messages: messages(r.Messages), Units: lives(r.Live),
		Holds: changed(HoldBytes(g), r.Holds)}
	for k := range 2 {
		m.Occupants[k] = occupants(r.Maps[k])
	}
	return m
}

// Destinations are the cells the moving unit may be ordered to: the reach
// search, then the end-cell check on every reached cell (0b06:1383).
func (p Profile) Destinations(g *Movement) ([]int, error) {
	t, err := p.moveTables()
	if err != nil {
		return nil, err
	}
	_, bits, _ := t.Reach(g)
	return t.Destinations(g, bits)
}

// Route is the route the path search gives from the unit's cell to goal,
// start first; nil where the search finds none.
func (p Profile) Route(g *Movement, goal int) ([]int, error) {
	t, err := p.moveTables()
	if err != nil {
		return nil, err
	}
	buf, bits, _ := t.Reach(g)
	path, err := t.Path(g, buf, bits, goal)
	if err != nil || path == nil {
		return nil, err
	}
	slices.Reverse(path)
	return path, nil
}

// EndCheck is the end-cell check of the moving unit on a cell: 0 if it may
// end there, else the code it shows, and the action an arrival there takes.
func (p Profile) EndCheck(g *Movement, cell int) (int, Action, error) {
	t, err := p.moveTables()
	if err != nil {
		return 0, Action{}, err
	}
	return t.EndCheck(g, cell)
}

// Move executes a move order: the order code, the move start and the steps
// until the unit arrives, loading into a transporter or a building where
// the destination is one.
func (p Profile) Move(g *Movement, o MoveOrder) (Moved, error) {
	t, err := p.moveTables()
	if err != nil {
		return Moved{}, err
	}
	r, err := t.Execute(g, o)
	if err != nil {
		return Moved{}, err
	}
	return movedOf(g, r), nil
}

// Exit takes a unit out of a container and executes its move order.
func (p Profile) Exit(g *Movement, o ExitOrder) (Moved, error) {
	t, err := p.moveTables()
	if err != nil {
		return Moved{}, err
	}
	r, err := t.Exit(g, o)
	if err != nil {
		return Moved{}, err
	}
	m := movedOf(g, r.StepRun)
	m.Remaining = map[int]int{}
	for c := range g.Width * g.Height {
		if v := r.Remaining[c/g.Width*64+c%g.Width]; v >= 0 {
			m.Remaining[c] = int(v)
		}
	}
	return m, nil
}

// Phased is the state the phase change leaves.
type Phased struct {
	Units     map[int]Live   // the records of PhaseChange.Units
	Holds     map[int]Hold   // the container records that changed
	Occupants [2]map[int]int // the occupied cells of side 0's and side 1's map (identical)
	Phase     [2]int         // the sides' phase bytes
	Count     int            // the turn counter
	Messages  []Message
}

// ChangePhase is the phase change at the end of a side's orders: both
// sides' views are brought into line and the phases swap.
func (p Profile) ChangePhase(g *PhaseChange) (Phased, error) {
	r, err := ChangePhase(g)
	if err != nil {
		return Phased{}, err
	}
	out := Phased{Units: map[int]Live{}, Holds: changed(ContainerRecords(g.Holds), r.Holds),
		Count: r.Count, Messages: messages(r.Messages)}
	for i := range g.Units {
		out.Units[i] = DecodeLive(r.Live[i])
	}
	for k := range 2 {
		out.Occupants[k] = occupants(r.Maps[k])
		out.Phase[k] = int(r.Players[k][0x16])
	}
	return out, nil
}
