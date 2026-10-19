package dosoracle

import "fmt"

// Container records (Move.Containers, 0x1c bytes each) hold the units a
// transporter carries and the units in a building. Transporters index them
// by live +0x0a; the building records follow the transporters' in three
// tables, which the end-cell check picks by terrain flag: 32e2:38b4 (two
// records, chosen by flag bit 1, for flag 0x40), 32e2:38ec (flag 0x100) and
// 32e2:3a04 (flag 0x400), ten records each, found by cell (0e40:1728).
const (
	holdRecords = 92
	holdFlag40  = 70 // (0x38b4 - 0x310c) / 0x1c
	holdFlag100 = 72
	holdFlag400 = 82
	holdTable   = 10   // records 0e40:1728 searches
	slotFree    = 0xF0 // a slot or map occupant above it holds no unit
	inBuilding  = 0xF4 // the map occupant of a building cell with units in it
)

// Hold is a container record: seven slot bytes per side (unit indices,
// above 0xf0 free) at +7*side, the cells it covers as map offsets (four
// words at +0x0e; a transporter's first is its position 2*cell+1, marked by
// flag 0x20), the flag word +0x19 (side bits 1 and 2 as in live +4, 2 also a
// neutral building; bits 0xfc0 the classes a transporter takes, matched
// against live +6) and +0x1b, the occupant the map gets back where a unit
// leaves it.
type Hold struct {
	Slots    [14]int `json:"slots"`
	Cells    [4]int  `json:"cells"`
	Flags    int     `json:"flags"`
	Occupant int     `json:"occupant"`
}

// EmptyHold is a record with free slots and zeros otherwise, as every
// record not given in Movement.Holds is set up.
func EmptyHold() Hold {
	var h Hold
	for i := range h.Slots {
		h.Slots[i] = 0xFF
	}
	return h
}

func (h Hold) bytes() []byte {
	b := make([]byte, containerLen)
	for i, v := range h.Slots {
		b[i] = byte(v)
	}
	for k, c := range h.Cells {
		putWord(b, 0x0E+2*k, c)
	}
	putWord(b, 0x19, h.Flags)
	b[0x1B] = byte(h.Occupant)
	return b
}

// holdBytes are all container records of a movement.
func holdBytes(g *Movement) map[int][]byte { return holdsOf(g.Holds) }

// holdsOf are all container records, EmptyHold where holds has none.
func holdsOf(holds map[int]Hold) map[int][]byte {
	out := make(map[int][]byte, holdRecords)
	for i := range holdRecords {
		h, ok := holds[i]
		if !ok {
			h = EmptyHold()
		}
		out[i] = h.bytes()
	}
	return out
}

// Action is what the end-cell check leaves in the action record 32e2:1805
// for the arrival 10cc:0ae2.
type Action struct {
	Kind     int // +0x0c: 3 a plain move, 2 into a transporter, 4 onto a foreign building, 5 into an own building
	Occupant int // +1: the unit on the cell, as the map holds it (a ship's rear)
	Slot     int // +4: the free slot taken
	Hold     int // the record +5 points at
	Index    int // +9: the transporter's live +0x0a, or the building's index in its table (-1: none)
}

// EndCheck is 10cc:0f74 for the moving unit: 0 if it may end on the cell,
// else the code it shows, and the action it leaves. On a unit's cell (a
// two-cell ship's rear stands for its head) it gives 0x37 for another
// side's unit and 0x39 for an own unit that carries nothing. An own
// transporter takes the unit into a free slot of the moving side (action 2)
// unless the unit's live +6 shares none of the classes in the record's
// +0x19 & 0xfc0 or the transporter has moved (+4 bit 0x200: 0x3a), or the
// side's seven slots are taken or the weights exceed the transporter type's
// capacity (0x3b). The weights are the unit type's, the cargo's of a
// transporter that moves (which also needs a free slot per cargo unit plus
// one) and the side's cargo already in the record. A building cell of
// terrain flag 0x400, 0x100 or 0x40 is action 4 if the building's +0x19 is
// another side's or neutral (2), else action 5 into a free slot (0x3b if
// none, or too few for a moving transporter's cargo); the terrain is not
// checked then. Otherwise it gives 0x40 for terrain the unit's mask does
// not enter or, for an air unit, terrain flag 0x4000.
func (t *MoveTables) EndCheck(g *Movement, cell int) (int, Action, error) {
	return newMoveState(t, g).endCheck(cell)
}

func (s *moveState) endCheck(cell int) (int, Action, error) {
	t, side := s.t, s.g.Side
	mover := s.rec(s.g.Unit)
	a := Action{Kind: 3}
	if occ := int(s.maps[0][2*cell+1]); occ <= slotFree {
		a.Occupant = occ
		o := occ
		if f := s.flags(o); f&0x40 != 0 && f&0x80 == 0 {
			o-- // the rear of a two-cell ship
		}
		of := s.flags(o)
		switch {
		case !own(of, side):
			return 0x37, a, nil
		case of&0x1000 == 0:
			return 0x39, a, nil
		}
		a.Index = int(int8(s.rec(o)[0x0A]))
		a.Hold = a.Index
		h, err := s.hold(a.Hold)
		if err != nil {
			return 0, a, err
		}
		if word(mover, 6)&word(h, 0x19)&0xFC0 == 0 || of&0x200 != 0 {
			return 0x3A, a, nil
		}
		slot := freeSlot(h, side)
		if slot == 7 {
			return 0x3B, a, nil
		}
		fits := true
		w := t.Types[mover[8]].Weight
		if word(mover, 4)&0x1000 != 0 {
			cargo, err := s.hold(int(int8(mover[0x0A])))
			if err != nil {
				return 0, a, err
			}
			w += s.cargoWeight(cargo)
			if slotCount(h, side, false) < slotCount(cargo, side, true)+1 {
				fits = false
			}
		}
		w += s.cargoWeight(h)
		if w > t.Types[s.rec(o)[8]].Capacity || !fits {
			return 0x3B, a, nil
		}
		a.Slot, a.Kind = slot, 2
		return 0, a, nil
	}

	tr := t.Terrains[s.maps[0][2*cell]]
	base, idx := -1, 0
	switch {
	case tr.Flags&0x400 != 0:
		base, idx = holdFlag400, s.findHold(holdFlag400, 2*cell)
	case tr.Flags&0x100 != 0:
		base, idx = holdFlag100, s.findHold(holdFlag100, 2*cell)
	case tr.Flags&0x40 != 0:
		base, idx = holdFlag40, tr.Flags&1
	}
	if base >= 0 {
		// A cell no record lists (-1) reads the record before the table.
		a.Index, a.Hold = idx, base+idx
		h, _ := s.hold(a.Hold)
		slot := freeSlot(h, side)
		if hf := word(h, 0x19); !own(hf, side) || hf&2 != 0 {
			a.Kind = 4
			return 0, a, nil
		}
		if word(mover, 4)&0x1000 != 0 {
			cargo, err := s.hold(int(int8(mover[0x0A])))
			if err != nil {
				return 0, a, err
			}
			if slotCount(h, side, false) < slotCount(cargo, side, true)+1 {
				slot = 7
			}
		}
		if slot == 7 {
			return 0x3B, a, nil
		}
		a.Slot, a.Kind = slot, 5
		return 0, a, nil
	}
	if t.Types[mover[8]].Mask&tr.Mask == 0 || word(mover, 4)&0x10 != 0 && tr.Flags&0x4000 != 0 {
		return 0x40, a, nil
	}
	return 0, a, nil
}

// hold is a container record by index; one outside the records the runner
// sets up is undefined.
func (s *moveState) hold(idx int) ([]byte, error) {
	h, ok := s.holds[idx]
	if !ok {
		return nil, ErrUndefined
	}
	return h, nil
}

// findHold is 0e40:1728: the record of the table at base listing the map
// offset among its four cells, or -1.
func (s *moveState) findHold(base, at int) int {
	for i := range holdTable {
		for k := range 4 {
			if word(s.holds[base+i], 0x0E+2*k) == at {
				return i
			}
		}
	}
	return -1
}

// freeSlot is the side's first free slot, 7 if none.
func freeSlot(h []byte, side int) int {
	for i := range 7 {
		if h[7*side+i] > slotFree {
			return i
		}
	}
	return 7
}

// slotCount is 10cc:17cb: the side's taken slots, or its free ones.
func slotCount(h []byte, side int, taken bool) int {
	n := 0
	for i := range 7 {
		if (h[7*side+i] <= slotFree) == taken {
			n++
		}
	}
	return n
}

// cargoWeight is 1307:15fd: the weights of the moving side's units in the
// record.
func (s *moveState) cargoWeight(h []byte) int {
	w := 0
	for i := range 7 {
		if u := h[7*s.g.Side+i]; u <= slotFree {
			w += s.t.Types[s.rec(int(u))[8]].Weight
		}
	}
	return w
}

// transfer is 10cc:1840: each unit in the moving side's slots of src moves
// to the first free slot of dst, if there is one, and takes dst's first
// cell as its position.
func (s *moveState) transfer(src, dst []byte) {
	side := s.g.Side
	for i := range 7 {
		u := src[7*side+i]
		if u > slotFree {
			continue
		}
		for k := range 7 {
			if dst[7*side+k] > slotFree {
				dst[7*side+k] = u
				s.setPos(int(u), word(dst, 0x0E))
				src[7*side+i] = 0xFF
				break
			}
		}
	}
}

// ExitOrder is a unit leaving a container: the moving side's slot of record
// Hold, then a move order to Goal.
type ExitOrder struct {
	Hold int `json:"hold"`
	Slot int `json:"slot"`
	MoveOrder
}

// ExitRun is the executed outcome of an exit: the unit, the cost buffer the
// exit's reach search left, and the move order.
type ExitRun struct {
	Unit      int
	Remaining []int8
	StepRun
}

// Exit is the exit branch of the main loop (0661:2f0a-30fe), then the move
// order as Execute runs it. The unit is taken from the slot (which becomes
// free) and placed at the record's first cell, one less for a transporter
// (flag 0x20); its live +4 gains 0x100, the side's map shows it there and
// the cell gets the side's selection bit (0x40, 0x80 for side 1). The reach
// search runs from there with budget 4 for live +6 bit 0x10 and 8 otherwise,
// whatever live byte 0 holds, and the flag mask 3 (6 for live +4 bit 1),
// with 0x18 for live +6 bit 4. The move start then takes the record's +0x1b
// as the occupant the cell gets back. Undefined: an empty slot and an odd
// start offset (the map write would hit a terrain byte).
func (t *MoveTables) Exit(g *Movement, o ExitOrder) (ExitRun, error) {
	side := g.Side
	s := newMoveState(t, g)
	h, err := s.hold(o.Hold)
	if err != nil {
		return ExitRun{}, err
	}
	u := int(h[7*side+o.Slot])
	start := word(h, 0x0E)
	if word(h, 0x19)&0x20 != 0 {
		start--
	}
	if u > slotFree || start&1 != 0 || start/2 >= g.Width*g.Height {
		return ExitRun{}, ErrUndefined
	}
	if u != g.Unit {
		return ExitRun{}, fmt.Errorf("exit: slot %d holds unit %d, not Movement.Unit %d", o.Slot, u, g.Unit)
	}
	h[7*side+o.Slot] = 0xFF
	s.setPos(u, start+1)

	moved := *g
	moved.Budget, moved.Block = 8, 3
	l := s.rec(u)
	if word(l, 6)&0x10 != 0 {
		moved.Budget = 4
	}
	if word(l, 4)&1 != 0 {
		moved.Block = 6
	}
	if word(l, 6)&4 != 0 {
		moved.Block |= 0x18
	}
	p := moved.piece(u)
	p.Cell = start / 2
	moved.Held = map[int]Piece{u: p}
	for idx, q := range g.Held {
		if idx != u {
			moved.Held[idx] = q
		}
	}
	units := map[int]Piece{}
	for idx, q := range g.Units {
		if idx != u {
			units[idx] = q
		}
	}
	moved.Units = units
	s.g = &moved

	buf, bits, _ := t.Reach(&moved)
	s.maps[0][start+1] = byte(u)
	s.setFlags(u, s.flags(u)|0x100)
	bits[s.bitIndex(start/2)] |= byte(0x40 << side)
	s.bits = bits
	r := ExitRun{Unit: u, Remaining: buf}
	r.StepRun, err = s.drive(buf, o.MoveOrder, o.Hold)
	return r, err
}
