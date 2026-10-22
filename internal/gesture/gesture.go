// Package gesture is the cursor and order input of the original's battle screen
// (`-rules dos`): the action placed on each gesture direction, a per-frame
// recognizer for the mouse and a key set, the remake's right-click ring and
// placeholder drawing of the cursor ring and the action plates.
//
// Gestures go four ways from the cursor cell (Dir). While fire is held the
// cursor becomes a plate that shows cancel; pushing a direction picks the
// action placed there and releasing runs it. On the map, up is move (attack in
// the attack mode), left is mode change (open on an own container), right is
// the strategic map and down is info; in a container's contents up is move
// out, left produce, right cancel and down repair.
//
// The Tracker turns pointer and key state into events. A mouse press and
// release that never dragged as far as the threshold (one cell pitch on
// screen) is a Click, which the caller treats as select or confirm; a drag
// picks by its dominant axis, and dragging back inside the threshold returns
// to cancel. With the keys, an arrow moves the cursor one cell (see Neighbor)
// and, while fire is held, picks. Right-click opens a Ring of every available
// action around the cursor.
//
// Positions are device pixels; drawing takes a centre and a scale in device
// pixels per world pixel, so a plate is the hexgrid cell polygon at that scale.
// The package does not know the game: the caller decides which actions are
// possible (Placement.Restrict) and what Click, Run and Step do.
package gesture

import (
	"fmt"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Action is an order on an action plate.
type Action int

// The actions of the map and of a container's contents. Cancel is the zero
// value, so an unset Placement cancels everywhere.
const (
	Cancel Action = iota
	Move
	Attack
	ModeChange
	Open
	Map
	Info
	MoveOut
	Produce
	Repair
)

var actionNames = [...]string{"cancel", "move", "attack", "mode change", "open", "map", "info", "move out", "produce", "repair"}

func (a Action) String() string {
	if a < 0 || int(a) >= len(actionNames) {
		return fmt.Sprintf("Action(%d)", int(a))
	}
	return actionNames[a]
}

// Dir is a gesture direction from the cursor cell; Centre is no direction.
type Dir int

// Gesture directions, in screen terms (up is towards the top of the screen).
const (
	Centre Dir = iota
	Up
	Left
	Right
	Down
)

var dirNames = [...]string{"centre", "up", "left", "right", "down"}

func (d Dir) String() string {
	if d < 0 || int(d) >= len(dirNames) {
		return fmt.Sprintf("Dir(%d)", int(d))
	}
	return dirNames[d]
}

// Placement is the action on each direction for one cursor context, and
// whether it is possible there. The zero Placement is cancel on every
// direction.
type Placement struct {
	acts [5]Action
	off  [5]bool
}

// Place returns a placement with every action enabled.
func Place(centre, up, left, right, down Action) Placement {
	return Placement{acts: [5]Action{centre, up, left, right, down}}
}

// ForMap is the placement on the map: move or attack above, mode change or
// open (on an own container) to the left, map to the right, info below.
func ForMap(attackMode, ownContainer bool) Placement {
	upAct, leftAct := Move, ModeChange
	if attackMode {
		upAct = Attack
	}
	if ownContainer {
		leftAct = Open
	}
	return Place(Cancel, upAct, leftAct, Map, Info)
}

// ForContents is the placement in a container's contents: move out above,
// produce to the left, cancel to the right, repair below.
func ForContents() Placement { return Place(Cancel, MoveOut, Produce, Cancel, Repair) }

// At returns the action on d and whether it is possible.
func (p Placement) At(d Dir) (Action, bool) {
	if d < Centre || d > Down {
		return Cancel, true
	}
	return p.acts[d], !p.off[d]
}

// Restrict disables every action ok rejects; cancel always stays possible.
// A disabled action still shows on its plate, greyed, and releasing on it
// cancels.
func (p Placement) Restrict(ok func(Action) bool) Placement {
	for d, a := range p.acts {
		p.off[d] = p.off[d] || (a != Cancel && !ok(a))
	}
	return p
}

// RingActions returns the possible actions other than cancel clockwise from
// the top (up, right, down, left), the order of a Ring, so a full ring puts
// each action on its gesture direction.
func (p Placement) RingActions() []Action {
	var out []Action
	for _, d := range [...]Dir{Up, Right, Down, Left} {
		if a, ok := p.At(d); ok && a != Cancel {
			out = append(out, a)
		}
	}
	return out
}

// Neighbor returns the cell one key step from c. Up and down go to the cell
// straight above and below (two lines away). There is no horizontal neighbour
// on flat-top hexes, so left and right go to the upper diagonal from an even
// line and to the lower one from an odd line: the cursor zigzags along a row
// instead of drifting up or down, and left undoes right.
func Neighbor(c hexgrid.Coord, d Dir) hexgrid.Coord {
	n := hexgrid.Around(c) // NE, SE, S, SW, NW, N
	odd := c.Line&1 == 1
	switch d {
	case Up:
		return n[5]
	case Down:
		return n[2]
	case Left:
		if odd {
			return n[3]
		}
		return n[4]
	case Right:
		if odd {
			return n[1]
		}
		return n[0]
	}
	return c
}
