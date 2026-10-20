package dosrules

import "fmt"

// moveState is the memory the move routines read and change.
type moveState struct {
	t        *MoveTables
	g        *Movement
	live     map[int][]byte
	maps     [2][]byte // the side's map and the other side's map
	bits     []byte
	route    []int // cells from the goal back to the start, as the path search stores them
	count    int
	orders   []int
	saved    [2]int         // 32e2:1807 and 1808
	holds    map[int][]byte // the container records
	action   Action         // the action record the order code's end-cell check left
	messages [][2]int
}

// newMoveState is the memory RunMove sets up for a movement.
func newMoveState(t *MoveTables, g *Movement) *moveState {
	s := &moveState{t: t, g: g, live: LiveRecords(g), holds: HoldBytes(g)}
	s.maps[0], s.maps[1] = MovementMaps(g)
	return s
}

// rec is a live record; one never placed reads as zeros, as in the game's
// array.
func (s *moveState) rec(idx int) []byte {
	if s.live[idx] == nil {
		s.live[idx] = make([]byte, LiveLen)
	}
	return s.live[idx]
}

func (s *moveState) pos(idx int) int { return word(s.rec(idx), 0x0B+2*s.g.Side) }

func (s *moveState) setPos(idx, v int) { putWord(s.rec(idx), 0x0B+2*s.g.Side, v) }

func (s *moveState) flags(idx int) int { return word(s.rec(idx), 4) }

func (s *moveState) setFlags(idx, v int) { putWord(s.rec(idx), 4, v) }

func (s *moveState) bitIndex(cell int) int { return cell/s.g.Width*64 + cell%s.g.Width }

// Execute is a move order as RunMove executes it: Reach, then the order code
// 10cc:02dd, the move start 10cc:05c3 and the step 10cc:071e until it
// arrives (10cc:0ae2). It returns ErrUndefined where Path or EndCheck do.
func (t *MoveTables) Execute(g *Movement, o MoveOrder) (StepRun, error) {
	s := newMoveState(t, g)
	buf, bits, _ := t.Reach(g)
	for _, c := range o.Stale {
		bits[s.bitIndex(c)] |= byte(1 + g.Side)
	}
	s.bits = bits
	return s.drive(buf, o, -1)
}

// drive runs the order code, the start and the steps on the reach search's
// buffer and bits; from is the container record an exiting unit leaves
// (-1: none).
func (s *moveState) drive(buf []int8, o MoveOrder, from int) (StepRun, error) {
	g := s.g
	var r StepRun
	var err error
	if r.Order, err = s.order(buf, o.Goal); err != nil {
		return r, err
	}
	dest := 2 * o.Goal
	if o.Retarget >= 0 {
		dest = 2 * o.Retarget
	}
	if r.Order&1 != 0 {
		r.Start = s.start(dest, from)
	}
	for r.Start&1 != 0 {
		entry := s.count
		if err := s.step(); err != nil {
			return r, err
		}
		r.Cells = append(r.Cells, (s.pos(g.Unit)-1)/2)
		if entry == 0 {
			break
		}
		if len(r.Cells) == MaxSteps {
			return r, fmt.Errorf("step: no arrival")
		}
	}
	r.Live = map[int][]byte{}
	for idx := range LiveRecords(g) {
		r.Live[idx] = s.live[idx]
	}
	r.Messages, r.Maps, r.Bits, r.Orders, r.Count, r.Holds =
		s.messages, s.maps, s.bits, s.orders, s.count, s.holds
	r.Container = s.holds[0]
	return r, nil
}

// order is 10cc:02dd: it refuses (0x12) a destination without the side's
// reach bit, the unit's own cell, terrain flag 0x4000 for an air unit and a
// non-zero end code, which it shows. Otherwise it searches the path (which
// leaves the bits as it found them); without one it shows 0x48 and refuses.
// With one it clears the side's reach bits everywhere, sets them on the
// route's cells and stores the destination in live +0x11. A two-cell ship
// adds 8.
func (s *moveState) order(buf []int8, goal int) (int, error) {
	g, t := s.g, s.t
	u := g.Unit
	reach := 1
	if g.Side != 0 {
		reach = 2
	}
	res := 1
	if int(s.bits[s.bitIndex(goal)])&reach == 0 {
		res = 0x12
	}
	if s.pos(u) == 2*goal+1 {
		res = 0x12
	}
	if s.flags(u)&0x10 != 0 && t.Terrains[g.Terrain[goal]].Flags&0x4000 != 0 {
		res = 0x12
	}
	if res&1 != 0 {
		code, a, err := s.endCheck(goal)
		if err != nil {
			return 0, err
		}
		s.action = a
		if code != 0 {
			s.messages = append(s.messages, [2]int{code, g.Side})
			res = 0x12
		}
	}
	if s.flags(u)&0x40 != 0 {
		res |= 8
	}
	if res&1 == 0 {
		return res, nil
	}
	route, err := t.Path(g, buf, s.bits, goal)
	if err != nil {
		return 0, err
	}
	// The path search clears the side's visit bit everywhere when it
	// returns (0b06:000f), an exit's selection bit included.
	for i := range s.bits {
		s.bits[i] &^= byte(0x40 << g.Side)
	}
	if route == nil {
		s.messages = append(s.messages, [2]int{0x48, g.Side})
		return 0x12, nil
	}
	for i := range s.bits {
		s.bits[i] &^= byte(reach)
	}
	for _, c := range route {
		s.bits[s.bitIndex(c)] |= byte(reach)
	}
	s.route, s.count = route, len(route)
	putWord(s.rec(u), 0x11, 2*goal)
	return res, nil
}

// start is 10cc:05c3: it refuses (0x12) an order whose destination is no
// longer the one in live +0x11. Otherwise it appends the unit to the side's
// order list, stores the unit's cell in live +0x11, skips the route's start
// and the first count step (count - 2) and clears both saved occupants; an
// exit order (from >= 0) takes the first from the record's +0x1b.
func (s *moveState) start(dest, from int) int {
	u := s.g.Unit
	res := 1
	if word(s.rec(u), 0x11) != dest {
		res = 0x12
	}
	if s.flags(u)&0x40 != 0 {
		res |= 8
	}
	if res&1 == 0 {
		return res
	}
	s.orders = append(s.orders, u)
	putWord(s.rec(u), 0x11, s.pos(u))
	s.count -= 2
	s.saved = [2]int{0xFF, 0xFF}
	if from >= 0 {
		s.saved[0] = int(s.holds[from][0x1B])
	}
	return res
}

// step is 10cc:071e: the unit enters route[count] on its side's map only. The
// cell it leaves gets back the occupant saved when it entered (0xff at the
// start) and the occupant of the cell it enters is saved. A two-cell ship's
// rear (the next record) takes the head's old cell and facing, its own old
// cell getting back the second saved occupant. The facing +0x0f+side is the
// direction of the step (facing). A transporter moves its cargo (cargo). With
// count 0 the unit has arrived (arrive); otherwise count drops by one.
func (s *moveState) step() error {
	g := s.g
	u, side := g.Unit, g.Side
	mp := s.maps[0]
	next := 2 * s.route[s.count]
	if s.flags(u)&0x40 != 0 {
		rear := u + 1
		if s.pos(rear) == 0 {
			// A head without its rear record (the game places both): the
			// arrival marks the byte before the bit array.
			return ErrUndefined
		}
		old := s.pos(u)
		mp[s.pos(rear)] = byte(s.saved[1])
		s.saved[1] = s.saved[0]
		s.saved[0] = int(mp[next+1])
		mp[next+1] = byte(u)
		s.rec(u)[0x0F+side] = byte(facing(old-1, next, g.Width))
		s.setPos(u, next+1)
		mp[old] = byte(u + 1)
		s.rec(rear)[0x0F+side] = s.rec(u)[0x0F+side]
		s.setPos(rear, old)
	} else {
		old := s.pos(u)
		mp[old] = byte(s.saved[0])
		s.saved[0] = int(mp[next+1])
		mp[next+1] = byte(u)
		s.rec(u)[0x0F+side] = byte(facing(old-1, next, g.Width))
		s.setPos(u, next+1)
	}
	if s.flags(u)&0x1000 != 0 {
		s.cargo(u)
	}
	if s.count > 0 {
		s.count--
		return nil
	}
	return s.arrive()
}

// arrive is 10cc:0ae2 for the action the end-cell check left, then the end
// of 071e: it shows the arrival's code (0x38) and clears live +4 bit 0x100,
// for the ship partner too.
func (s *moveState) arrive() error {
	u, side := s.g.Unit, s.g.Side
	a := s.action
	l := s.rec(u)
	switch a.Kind {
	case 3:
		// The plain move: live +4 loses 0x4100 and gains the moved bit
		// 0x200, which 0e40:1ddd also sets in the bit array (0x10, 0x20 for
		// side 1) at the unit's cell and at its ship partner's. +0x13 takes
		// the action, +0x14 gains 1 and +0x15 to +0x19 are cleared.
		s.setFlags(u, s.flags(u)&0xBEFF|0x200)
		s.setMoved(u, 0x200)
		l[0x13] = 3
		if s.flags(u)&0x1000 != 0 {
			s.cargo(u)
		}
		l[0x14] |= 1
		l[0x15], l[0x19], l[0x17], l[0x18], l[0x16] = 0, 0, 0, 0, 0
	case 2:
		// Into a transporter: the unit takes the free slot, the map shows
		// the transporter again, live +4 loses 0x100 and gains 0x4200 (the
		// bit array gets no moved bit) and a transporter hands its cargo
		// over. +0x19 holds the slot, +0x15 the record, +0x16 the
		// transporter.
		h := s.holds[a.Hold]
		h[7*side+a.Slot] = byte(u)
		s.maps[0][s.pos(u)] = byte(a.Occupant)
		s.setFlags(u, s.flags(u)&0xFEFF|0x4200)
		if s.flags(u)&0x1000 != 0 {
			s.transfer(s.holds[int(int8(l[0x0A]))], h)
		}
		l[0x13], l[0x19], l[0x15], l[0x16] = 2, byte(a.Slot), byte(a.Index), byte(a.Occupant)
		l[0x14] |= 1
	case 4:
		// Onto a foreign or neutral building: a plain arrival on the map;
		// +0x17 takes the building's flag word, +0x15 its index.
		s.setFlags(u, s.flags(u)&0xBEFF|0x200)
		s.setMoved(u, 0x200)
		l[0x13] = 4
		l[0x14] |= 1
		if s.flags(u)&0x1000 != 0 {
			s.cargo(u)
		}
		putWord(l, 0x17, word(s.holds[a.Hold], 0x19))
		l[0x15], l[0x16] = byte(a.Index), 0
	case 5:
		// Into an own building: the unit takes the free slot, a
		// transporter hands its cargo over, the map shows 0xf4 and the
		// unit's position becomes the building's first cell (an even map
		// offset); live +4 loses 0x100 and gains 0x4200.
		h := s.holds[a.Hold]
		l[0x13] = 5
		l[0x14] |= 1
		l[0x19], l[0x15] = byte(a.Slot), byte(a.Index)
		putWord(l, 0x17, word(h, 0x19))
		h[7*side+a.Slot] = byte(u)
		if s.flags(u)&0x1000 != 0 {
			s.transfer(s.holds[int(int8(l[0x0A]))], h)
		}
		s.maps[0][s.pos(u)] = InBuilding
		s.setPos(u, word(h, 0x0E))
		s.setFlags(u, s.flags(u)&0xFEFF|0x4200)
	default:
		return fmt.Errorf("arrival: action %d is not set by the end-cell check", a.Kind)
	}
	s.messages = append(s.messages, [2]int{0x38, side})
	s.setFlags(u, s.flags(u)&^0x100)
	if p, ok := s.partner(u); ok {
		s.setFlags(p, s.flags(p)&^0x100)
	}
	return nil
}

// partner is the other record of a two-cell ship: the next one from the
// head (+4 bit 0x80), the previous one from the rear.
func (s *moveState) partner(u int) (int, bool) {
	f := s.flags(u)
	switch {
	case f&0x40 == 0:
		return 0, false
	case f&0x80 != 0:
		return u + 1, true
	}
	return u - 1, true
}

// setMoved is 0e40:1ddd setting flag 0x200: live +4 and the side's moved bit
// at the unit's cell, for both records of a ship.
func (s *moveState) setMoved(u, flag int) {
	bit := byte(0x10)
	if s.g.Side&1 != 0 {
		bit = 0x20
	}
	mark := func(idx int) {
		s.setFlags(idx, s.flags(idx)|flag)
		s.bits[s.bitIndex((s.pos(idx)-1)/2)] |= bit
	}
	mark(u)
	if p, ok := s.partner(u); ok {
		mark(p)
	}
}

// cargo is 159e:0805: the units in the side's seven slots of the unit's
// container record (live +0x0a) and the record's first cell take the unit's
// position.
func (s *moveState) cargo(u int) {
	c := s.holds[int(int8(s.rec(u)[0x0A]))]
	p := s.pos(u)
	for k := range 7 {
		if occ := int(c[7*s.g.Side+k]); occ <= SlotFree {
			s.setPos(occ, p)
		}
	}
	putWord(c, 0x0E, p)
}

// facing is 10cc:143e: the direction (neighbours64 order) from map offset
// from to map offset to (2*cell each).
func facing(from, to, width int) int {
	fc, fr := from/2%width, from/2/width
	tc, tr := to/2%width, to/2/width
	switch {
	case tc == fc:
		if tr < fr {
			return 0
		}
		return 3
	case fc&1 != 0:
		switch {
		case tc < fc && tr > fr:
			return 4
		case tc < fc:
			return 5
		case tr > fr:
			return 2
		}
		return 1
	}
	switch {
	case tc < fc && tr < fr:
		return 5
	case tc < fc:
		return 4
	case tr < fr:
		return 1
	}
	return 2
}
