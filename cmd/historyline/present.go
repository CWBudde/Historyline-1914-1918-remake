package main

import (
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

// presentation is the controller's UI state: the selection and its attack
// overlay, the walk, effects, unit facings and the message line. It holds no
// rules state; units are identified by whatever key the controller uses (any
// comparable value).
type presentation struct {
	selected bool
	from     hexgrid.Coord
	targets  map[hexgrid.Coord]bool

	walk *walk   // movement being animated; input waits until it is done
	fx   effects // hit marks and banners

	reduced bool // reduced motion: instant walk, short static effects

	facing map[any]unitsprite.Facing // where each unit looks, keyed by the controller's unit

	Message string // result of the last action, shown to the player
}

// SetReducedMotion switches reduced motion on or off. Turning it on ends any
// walk and banner in progress.
func (p *presentation) SetReducedMotion(on bool) {
	p.reduced, p.fx.Reduced = on, on
	if on {
		p.walk = nil
		p.fx.Dismiss()
	}
}

// SkipWalk finishes the walk in progress at once and reports whether there
// was one. The rules engine has already placed the unit, so only the drawing
// changes.
func (p *presentation) SkipWalk() bool {
	if p.walk.Done() {
		return false
	}
	p.walk = nil
	return true
}

// clearSelection drops the selection and its attack overlay.
func (p *presentation) clearSelection() { p.selected, p.targets = false, nil }

// tick advances the walk and the effects by one frame.
func (p *presentation) tick() {
	if p.walk != nil {
		p.walk.Tick()
	}
	p.fx.Tick()
}

// Facing is the authored direction u is drawn in; units start facing north.
// It is presentation state and never reaches the rules.
func (p *presentation) Facing(u any) unitsprite.Facing { return p.facing[u] }

// turn makes u face from hex from towards hex to.
func (p *presentation) turn(u any, from, to hexgrid.Coord) {
	f, ok := toward(from, to)
	if !ok {
		return
	}
	if p.facing == nil {
		p.facing = map[any]unitsprite.Facing{}
	}
	p.facing[u] = f
}

// toward is the facing from hex from towards hex to, false for the same hex.
func toward(from, to hexgrid.Coord) (unitsprite.Facing, bool) {
	x0, y0 := hexgrid.Center(from)
	x1, y1 := hexgrid.Center(to)
	return unitsprite.Toward(x1-x0, y1-y0)
}
