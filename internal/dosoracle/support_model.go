package dosoracle

import "errors"

// ErrUndefined is raised where the original reads an uninitialised stack byte.
var ErrUndefined = errors.New("undefined: direction from an uninitialised stack byte")

var supportWeights = [6]int{0, 25, 50, 100, 50, 25}

const longRange = 0x40

// Unit is a live unit as the support routine reads it.
type Unit struct{ Flags4, Flags6, Type, Cell int }

// Ring returns the six neighbours of a cell in the DOS order (0 up, 1..2
// right side, 3 down, 4..5 left); -1 is off the map. Odd columns sit half a
// cell lower.
func Ring(cell, width, height int) [6]int {
	x, y := cell%width, cell/width
	n := [6]int{}
	if x&1 != 0 {
		n[0], n[3] = cell-width, cell+width
		n[1], n[5] = cell+1, cell-1
		n[4], n[2] = cell+width-1, cell+width+1
		if y <= 0 {
			n[0] = -1
		} else if height-1 <= y {
			n[2], n[3], n[4] = -1, -1, -1
		}
		if width-1 <= x {
			n[1], n[2] = -1, -1
		}
	} else {
		n[0], n[3] = cell-width, cell+width
		n[2], n[4] = cell+1, cell-1
		n[1], n[5] = cell-width+1, cell-width-1
		if y <= 0 {
			n[0], n[5], n[1] = -1, -1, -1
		} else if height-1 <= y {
			n[3] = -1
		}
		if x <= 0 {
			n[4], n[5] = -1, -1
		} else if width-1 <= x {
			n[1], n[2] = -1, -1
		}
	}
	return n
}

// friendly is 0ca6:020a: units of side bit 1 or 2 pair with the same bit,
// others pair with non-side-1.
func friendly(neighbour, own int) bool {
	switch {
	case neighbour&1 != 0:
		return own&1 != 0
	case neighbour&2 != 0:
		return own&2 != 0
	}
	return own&1 == 0
}

// Map is the synthetic battlefield of the support routine: Cells holds one
// unit index per cell (0xFF empty) and Terrain0 is map byte 0.
type Map struct {
	Width, Height int
	Cells         []int
	Terrain0      int
	Units         map[int]Unit
	Types         map[int]int // type index -> flag word at +0xf
}

func (m *Map) scan(anchor, reference int) (slots [6]bool, direction int, err error) {
	b := m.Units[anchor]
	classes, bit2 := b.Flags4&0x3C, b.Flags6&2
	direction = -1
	for k, c := range Ring(b.Cell, m.Width, m.Height) {
		if c < 0 {
			continue
		}
		u := m.Cells[c]
		if u <= 0xF0 {
			un, ok := m.Units[u]
			if !ok {
				return slots, -1, errors.New("cell refers to an unknown unit")
			}
			t := m.Types[un.Type]
			if (classes&t != 0 || bit2&t != 0) && t&longRange == 0 &&
				friendly(un.Flags4, m.Units[reference].Flags4) {
				slots[k] = true
			}
		}
		if u == reference {
			direction = k
		}
	}
	return slots, direction, nil
}

// Support transcribes the pincer and blocking calculation 02dc:262a and
// returns the two bytes at 32f7:1854/1855.
func Support(m *Map, attacker, defender int) (pincer, blocking int, err error) {
	differing := 0
	for _, c := range Ring(m.Units[defender].Cell, m.Width, m.Height) {
		// off-map neighbours were read at map byte 0, valid ones at their unit byte
		v := m.Terrain0
		if c >= 0 {
			v = m.Cells[c]
		}
		if v != attacker {
			differing++
		}
	}
	if differing == 6 {
		return 0, 0, nil
	}
	slots, dir, err := m.scan(defender, attacker)
	if err != nil {
		return 0, 0, err
	}
	if dir < 0 {
		return 0, 0, ErrUndefined
	}
	for k := 1; k < 6; k++ {
		if slots[(dir+k)%6] {
			pincer = (pincer + supportWeights[k]) & 0xFF // a byte: the 300 clamp is dead code
		}
	}
	slots, dir, err = m.scan(attacker, defender)
	if err != nil {
		return 0, 0, err
	}
	if dir < 0 {
		return 0, 0, ErrUndefined
	}
	for _, k := range [2]int{1, 5} {
		if slots[(dir+k)%6] {
			blocking += 80
		}
	}
	return pincer, min(blocking, 150), nil
}
