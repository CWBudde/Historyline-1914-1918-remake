package dosrules

import "errors"

// ErrUndefined is the case where the original reads an uninitialised stack
// byte: a defender at the map edge, not adjacent to its attacker, whose
// off-map neighbour reads map byte 0 as the attacker's index.
var ErrUndefined = errors.New("undefined: direction from an uninitialised stack byte")

// Neighbour steps (column, row) in the DOS order 0 up, 1-2 right, 3 down,
// 4-5 left, for even and odd columns; odd columns sit half a cell lower.
var steps = [2][6][2]int{
	{{0, -1}, {1, -1}, {1, 0}, {0, 1}, {-1, 0}, {-1, -1}},
	{{0, -1}, {1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}},
}

// neighbours returns the six neighbours of a cell (0b06:0157), -1 off the map.
// Width and height are at least 2.
func neighbours(cell, width, height int) [6]int {
	x, y := cell%width, cell/width
	var n [6]int
	for k, s := range steps[x&1] {
		nx, ny := x+s[0], y+s[1]
		n[k] = -1
		if nx >= 0 && nx < width && ny >= 0 && ny < height {
			n[k] = ny*width + nx
		}
	}
	return n
}

// friendly is the side test (0ca6:020a) on two live flag words: side bits 1
// and 2 pair with themselves, a unit without either pairs with non-side-1.
func friendly(unit, own int) bool {
	switch {
	case unit&1 != 0:
		return own&1 != 0
	case unit&2 != 0:
		return own&2 != 0
	}
	return own&1 == 0
}

// Pincer weights by the step from the attacker's direction around the defender.
var pincerWeights = [6]int{0, 25, 50, 100, 50, 25}

const (
	blockStep = 80
	blockCap  = 150
)

// supporters marks the neighbours of anchor that would support against
// opponent and returns the direction in which opponent stands (-1 if not).
// A supporter's type has one of anchor's live target classes (or anchor's
// live +6 bit 2) in its target flags, is not long range, and stands on the
// opponent's side.
func (p Profile) supporters(b *Battlefield, anchor, opponent int) (slots [6]bool, dir int) {
	u := b.Units[anchor]
	classes, bit2 := u.Flags4&classBits, u.Flags6&2
	dir = -1
	for k, c := range neighbours(u.Cell, b.Width, b.Height) {
		if c < 0 {
			continue
		}
		idx, ok := b.at(c)
		if ok {
			n := b.Units[idx]
			t := p.Tables.Types[n.Type].TargetFlags
			slots[k] = (classes&t != 0 || bit2&t != 0) && t&longRange == 0 &&
				friendly(n.Flags4, b.Units[opponent].Flags4)
			if idx == opponent {
				dir = k
			}
		}
	}
	return slots, dir
}

// Support rates the neighbours of a battle (02dc:262a): units on the
// attacker's side around the defender add the pincer percentage, units on
// the defender's side beside the defender's direction around the attacker
// add the blocking percentage. Only adjacent battles have support. The
// pincer is a byte (five supporters give 250, the 300 clamp is dead code);
// blocking adds 80 per unit and stops at 150.
func (p Profile) Support(b *Battlefield, attacker, defender int) (pincer, blocking int, err error) {
	adjacent := false
	for _, c := range neighbours(b.Units[defender].Cell, b.Width, b.Height) {
		// an off-map neighbour is read at map byte 0
		v := b.MapByte0
		if c >= 0 {
			v = 0xFF
			if idx, ok := b.at(c); ok {
				v = idx
			}
		}
		adjacent = adjacent || v == attacker
	}
	if !adjacent {
		return 0, 0, nil
	}
	slots, dir := p.supporters(b, defender, attacker)
	if dir < 0 {
		return 0, 0, ErrUndefined
	}
	for k := 1; k < 6; k++ {
		if slots[(dir+k)%6] {
			pincer = (pincer + pincerWeights[k]) & 0xFF
		}
	}
	slots, dir = p.supporters(b, attacker, defender)
	if dir < 0 {
		return 0, 0, ErrUndefined
	}
	for _, k := range [2]int{1, 5} {
		if slots[(dir+k)%6] {
			blocking += blockStep
		}
	}
	return pincer, min(blocking, blockCap), nil
}
