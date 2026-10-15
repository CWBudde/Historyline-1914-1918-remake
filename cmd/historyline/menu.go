package main

// menuNone means no row has keyboard or hover focus.
const menuNone = -1

// menuStep moves the focus one row in dir (+1 or -1), wrapping at both ends and
// skipping unusable rows. With no usable row the focus is left unchanged. From
// menuNone, +1 lands on the first usable row and -1 on the last.
func menuStep(focus, dir int, rows []WeaponOption) int {
	n := len(rows)
	if n == 0 || (dir != 1 && dir != -1) {
		return focus
	}
	i := focus
	if i < 0 || i >= n {
		if dir > 0 {
			i = -1
		} else {
			i = n
		}
	}
	for range n {
		i = ((i+dir)%n + n) % n
		if rows[i].Usable {
			return i
		}
	}
	return focus
}

// menuFocusValid returns focus if it names a row, else menuNone.
func menuFocusValid(focus int, rows []WeaponOption) int {
	if focus < 0 || focus >= len(rows) {
		return menuNone
	}
	return focus
}

// menuHit returns the row under (mx, my) given a row layout function, or menuNone.
func menuHit(mx, my, n int, row func(i int) (x, y, w, h int)) int {
	for i := range n {
		x, y, w, h := row(i)
		if mx >= x && mx < x+w && my >= y && my < y+h {
			return i
		}
	}
	return menuNone
}

// menuLabel is the text of row i: its number key, the label and, for a disabled
// row, a leading marker (the label already carries the reason).
func menuLabel(i int, o WeaponOption) string {
	s := "  " + o.Label
	if i < 9 {
		s = string(rune('1'+i)) + " " + o.Label
	}
	if !o.Usable {
		s = "x " + s
	}
	return s
}
