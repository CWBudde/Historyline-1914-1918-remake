package main

import (
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// menuKind says which menu control.menu holds.
type menuKind int

const (
	menuWeapons  menuKind = iota // weapon choice for an attack, bottom left
	menuContents                 // units stored in a building or carrier, bottom left
	menuActions                  // what the selected unit should do, beside it
)

// command is a row of the action menu (Python's ActionMenu buttons).
type command int

const (
	cmdMove command = iota
	cmdAttack
	cmdContents
	cmdExit
)

var commandNames = [...]string{cmdMove: "Move", cmdAttack: "Attack", cmdContents: "Contents", cmdExit: "Exit"}

// actions lists the action menu of the selected unit, each row disabled with
// the reason when the rules refuse it. Attack is listed only for an armed
// map unit, Contents only for a carrier, and a unit inside a building
// or carrier gets Move and Exit.
func (c *control) actions() []WeaponOption {
	u := c.unit
	if !c.selected || u == nil {
		return nil
	}
	inside := c.st.At(c.from).Unit != u
	var out []WeaponOption
	add := func(cmd command, why string) {
		o := WeaponOption{Slot: int(cmd), Label: commandNames[cmd], Usable: why == ""}
		if why != "" {
			o.Label += ": " + why
		}
		out = append(out, o)
	}

	var why string
	switch {
	case u.Moved:
		why = "already moved"
	case u.Attacked:
		why = "already attacked"
	case !inside && u.Def.Speed == 0:
		why = "cannot move"
	case len(c.moves) == 0:
		why = "nowhere to go"
	}
	add(cmdMove, why)

	if !inside && slices.ContainsFunc(u.Weapons[:], func(w *data.Weapon) bool { return w != nil }) {
		why = ""
		switch {
		case u.Moved:
			why = "already moved"
		case u.Attacked:
			why = "already attacked"
		case len(c.targets) == 0:
			why = "no target in range"
		}
		add(cmdAttack, why)
	}

	if !inside && u.Content != nil {
		why = ""
		if len(u.Content.Units()) == 0 {
			why = "empty"
		}
		add(cmdContents, why)
	}
	add(cmdExit, "")
	return out
}

// openActions opens the action menu of the selected unit.
func (c *control) openActions() {
	c.menuKind, c.menu, c.stored = menuActions, c.actions(), nil
}

// chooseAction answers action menu row i: Move and Attack narrow the overlays
// to that kind, Contents lists the carrier, Exit drops the selection. A
// disabled row only says why and keeps the menu open.
func (c *control) chooseAction(i int) {
	if i < 0 || i >= len(c.menu) {
		return
	}
	o := c.menu[i]
	if !o.Usable {
		c.Message = o.Label
		return
	}
	c.menu = nil
	switch command(o.Slot) {
	case cmdMove:
		c.targets = map[hexgrid.Coord]bool{}
		c.Message = c.unit.Type + ": pick a blue hex to move to"
	case cmdAttack:
		c.moves = nil
		c.Message = c.unit.Type + ": pick a red hex to attack"
	case cmdContents:
		c.OpenContents(c.from)
	case cmdExit:
		c.clear()
	}
}

// actionMenuGap is the space between a unit's hex and its action menu.
const actionMenuGap = 6

// actionMenuRect places an action menu of n rows, mw pixels wide, beside a unit
// whose hex is centred on (cx, cy) and half wide on screen: right of it as in
// Python, else left of it, and always inside the w x h screen margins.
func actionMenuRect(cx, cy, half, mw, n, w, h int) rect {
	mh := n * menuRowHeight
	x := cx + half + actionMenuGap
	if x+mw > w-hudMargin {
		if left := cx - half - actionMenuGap - mw; left >= hudMargin {
			x = left
		} else {
			x = w - hudMargin - mw
		}
	}
	x = max(x, hudMargin)
	y := min(cy-mh/2, h-hudMargin-mh)
	y = max(y, hudMargin)
	return rect{x, y, mw, mh}
}

// menuPress is the pressed state of a menu row: a row is chosen when the mouse
// button goes up over the row it went down on, so a press can be taken back by
// dragging off.
type menuPress struct{ row int }

func newMenuPress() menuPress { return menuPress{row: menuNone} }

// Row is the pressed row, menuNone when none is.
func (p *menuPress) Row() int { return p.row }

// update takes this frame's button state (just down, just up, held) and the
// row under the cursor. It returns the chosen row, or menuNone, and whether
// the menu used the mouse this frame.
func (p *menuPress) update(down, up, held bool, hit int) (chosen int, used bool) {
	switch {
	case down && hit != menuNone:
		p.row = hit
		return menuNone, true
	case p.row == menuNone:
		return menuNone, false
	case up:
		r := p.row
		p.row = menuNone
		if hit == r {
			return r, true
		}
		return menuNone, true
	case !held:
		p.row = menuNone
		return menuNone, false
	}
	return menuNone, true
}
