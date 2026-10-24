package main

import (
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

// presentation is the UI state every rules engine's controller shares: the
// selection and its attack overlay, the open menu, the walk, the battle panel,
// effects, unit facings and the message line. It holds no rules state; units
// are identified by whatever key the engine uses (any comparable value).
type presentation struct {
	selected bool
	from     hexgrid.Coord
	targets  map[hexgrid.Coord]bool

	menu       []WeaponOption // open menu, nil when closed
	menuKind   menuKind       // which menu menu is
	menuTarget hexgrid.Coord

	walk   *walk   // movement being animated; input waits until it is done
	report *report // battle panel of the last shot
	fx     effects // hit marks and banners

	reduced bool // reduced motion: instant walk, short static effects

	facing map[any]unitsprite.Facing // where each unit looks, keyed by the engine's unit

	Message string // result of the last action, shown to the player
}

// driver is what the game loop needs from a rules engine's controller: input
// in, presentation state out.
type driver interface {
	ui() *presentation
	Click(at hexgrid.Coord)
	ChooseMenu(i int)
	EndTurn() // F1
	Cancel()  // Esc: skip the walk, else clear the selection
	Clear()   // right click
	Tick()    // one frame: walk, battle panel, effects and the computer side
	SetReducedMotion(on bool)
}

var _ driver = (*control)(nil)

// drv is the controller of the running engine.
func (g *game) drv() driver {
	if g.dos != nil {
		return g.dos
	}
	return g.ctl
}

func (p *presentation) ui() *presentation { return p }

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

// DismissReport hides the battle panel (map click or key press).
func (p *presentation) DismissReport() { p.report.Dismiss() }

func (p *presentation) reportTime() int {
	if p.reduced {
		return reportFramesReduced
	}
	return reportFrames
}

// MenuOpen reports whether a menu waits for a choice.
func (p *presentation) MenuOpen() bool { return len(p.menu) > 0 }

// clearSelection drops the selection, its attack overlay and the menu.
func (p *presentation) clearSelection() { p.selected, p.targets, p.menu = false, nil, nil }

// tick advances the walk, the battle panel and the effects by one frame.
func (p *presentation) tick() {
	if p.walk != nil {
		p.walk.Tick()
	}
	p.report.Tick()
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
