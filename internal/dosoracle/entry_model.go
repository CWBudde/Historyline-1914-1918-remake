package dosoracle

import "slices"

// clone copies a Desk deeply.
func (d Desk) clone() Desk {
	c := d
	c.Live, c.Holds = map[int][]byte{}, map[int][]byte{}
	for i, b := range d.Live {
		c.Live[i] = slices.Clone(b)
	}
	for i, b := range d.Holds {
		c.Holds[i] = slices.Clone(b)
	}
	c.Maps = [2][]byte{slices.Clone(d.Maps[0]), slices.Clone(d.Maps[1])}
	c.Bits = slices.Clone(d.Bits)
	c.Record = slices.Clone(d.Record)
	c.Saved = slices.Clone(d.Saved)
	c.List = slices.Clone(d.List)
	c.Messages = slices.Clone(d.Messages)
	return c
}

// bitIndex is a cell's byte in the bit array: 64 per row.
func bitIndex(cell, width int) int { return cell/width*64 + cell%width }

// setFlag transcribes 0e40:1ddd with set = 1: the flag (0x100 or 0x200) in
// the unit's live +4, and in the bit array at the cell where side sees it
// the side's bit for that flag (0x40/0x80, or 0x10/0x20 for side 1).
func (d *Desk) setFlag(unit, flag, side, width int) {
	live := d.Live[unit]
	putWord(live, 4, word(live, 4)|flag)
	bit := 0
	switch {
	case flag&0x100 != 0:
		bit = 0x40 << side
	case flag&0x200 != 0:
		bit = 0x10 << side
	}
	cell := (word(live, 0x0B+2*side) - 1) / 2
	d.Bits[bitIndex(cell, width)] |= byte(bit)
}

// EnterAttack transcribes the attack order 0ad1:000c on the state before it.
//
// The kind of order comes from the unit's live +6: 0x4000 a depot builder,
// else 0x2000 a sapper, else an attack, which needs a unit byte (at most
// 0xf0) in the cursor's cell of the side's map. Every kind needs the side's
// target bit in the cell's bit-array byte (0x04, 0x08 for side 1); without
// it the order is refused with message 0x45. An accepted order sets the
// moved flag (setFlag 0x200) and live +0x11 to the cursor's map offset:
//   - a depot builder on depot terrain (the cell's terrain byte equals
//     Cycle.Site) gets +0x13 = 0x0d unless 10 depots are under way (message
//     0x47, refused); on other terrain +0x13 = 0x0b; message 0x46;
//   - a sapper gets +0x13 = 0x11, message 0x6e;
//   - an attack gets +0x14 = 2 and +0x13 = 9, is appended to the side's
//     order list, message 0x49.
//
// Only attacks enter the list, and the routine does not compare the count
// with the move limit. The message is shown in any case, with animation
// flag 0x08 (0x10 for side 1); an accepted order also sets flag 0x100 and
// Cycle.Wait to 0. AX is 1 when accepted.
func EnterAttack(before Desk, side, width int) Desk {
	d := before.clone()
	unit, off := word(d.Record, orderUnit), word(d.Record, 0)
	live := d.Live[unit]
	kind, msg := 0, 0
	switch f6 := word(live, 6); {
	case f6&0x4000 != 0:
		kind = 0x0C
	case f6&0x2000 != 0:
		kind = 0x10
	case d.Maps[0][off+1] <= 0xF0:
		kind = 9
	default:
		msg = 0x45
	}
	if d.Bits[bitIndex(off>>1, width)]&byte(0x55<<side)&0x0C == 0 {
		kind, msg = 0, 0x45
	}
	accept := func(action, code int) {
		d.setFlag(unit, 0x200, side, width)
		putWord(live, 0x11, off)
		live[0x13] = byte(action)
		msg = code
		d.Result = 1
	}
	d.Result = 0
	switch kind {
	case 0x0C:
		switch {
		case int(d.Maps[0][off]) != d.Site:
			accept(0x0B, 0x46)
		case d.Depots >= 10:
			msg = 0x47
		default:
			accept(0x0D, 0x46)
		}
	case 0x10:
		accept(0x11, 0x6E)
	case 9:
		live[0x14] = 2
		accept(9, 0x49)
		d.List = append(d.List, unit&0xFF)
	}
	d.Messages = append(d.Messages, [2]int{msg, side})
	d.Anim |= 0x08 << side
	if d.Result == 1 {
		d.Anim |= 0x100
		d.Wait = 0
	}
	return d
}

// holdAt is the container record a far pointer of the order record points
// to, or -1.
func (d *Desk) holdAt(off, seg int) int {
	if seg != d.HoldSeg || off < d.HoldBase || (off-d.HoldBase)%containerLen != 0 {
		return -1
	}
	return (off - d.HoldBase) / containerLen
}

// Cancel transcribes 10cc:0001 on the state before it, with the routine's
// mode argument.
//
// It clears the side's reach bits (0x01, 0x02 for side 1) everywhere. With
// mode 0x10 it also clears the side's selection bits (0x40/0x80), the
// unit's live +4 bit 0x100 (and its other half's, for a two-cell ship),
// sets the menu state (+0x17 = 0, +0x18 = 0x19, +0x1b = 5), animation flag
// 0x100 and Cycle.Wait = 1.
//
// When the order record shows an exit in progress (+0x19 = 3), mode 0x10
// first puts the container back on the side's map: 0xf4 in the cell where
// the unit stands, the unit's position word set to the record's first cell
// (+0x0e), and, for a transporter (+0x19 bit 0x20), the transporter's index
// (+0x1b) on that cell. The order record is then restored from the saved
// copy (0ca0:044c), and with mode 0x10 the unit goes back into the slot the
// restored record names (its +0x24 record, side's half, +0x0e). Finally the
// menu returns to the inventory (+0x17 = 2, +0x18 = 9, +0x19 = 0, +0x1b = 5,
// +0x0c/+0x0e from +0x04/+0x06).
func Cancel(before Desk, side, mode int) Desk {
	d := before.clone()
	rec := d.Record
	unit := word(rec, orderUnit)
	live := d.Live[unit]
	clear := func(mask byte) {
		for i := range d.Bits {
			d.Bits[i] &^= mask
		}
	}
	clear(byte(1 << side))
	if mode&0x10 != 0 {
		clear(byte(0x40 << side))
		putWord(live, 4, word(live, 4)&^0x100)
		if f4 := word(live, 4); f4&0x40 != 0 {
			other := unit - 1
			if f4&0x80 != 0 {
				other = unit + 1
			}
			o := d.Live[other]
			putWord(o, 4, word(o, 4)&^0x100)
		}
		rec[0x17], rec[0x18], rec[0x1b] = 0, 0x19, 5
		d.Anim |= 0x100
		d.Wait = 1
	}
	if rec[0x19] != 3 {
		return d
	}
	if mode&0x10 != 0 {
		d.Maps[0][word(live, 0x0B+2*side)] = 0xF4
		h := d.Holds[d.holdAt(word(rec, orderHold), word(rec, orderHold+2))]
		putWord(live, 0x0B+2*side, word(h, 0x0E))
		if word(rec, orderHold)|word(rec, orderHold+2) != 0 && word(h, 0x19)&0x20 != 0 {
			d.Maps[0][word(live, 0x0B+2*side)] = h[0x1B]
		}
	}
	copy(rec, d.Saved)
	if mode&0x10 != 0 {
		h := d.Holds[d.holdAt(word(rec, orderHold), word(rec, orderHold+2))]
		h[7*side+word(rec, orderSlot)] = byte(unit)
	}
	rec[0x17], rec[0x18], rec[0x19], rec[0x1b] = 2, 9, 0, 5
	putWord(rec, 0x0C, word(rec, 4))
	putWord(rec, 0x0E, word(rec, 6))
	return d
}
