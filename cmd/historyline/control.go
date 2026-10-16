package main

import (
	"fmt"
	"math/rand/v2"

	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

// control turns hex clicks into rules commands for one player. It holds only the
// selection and its overlays; all game state stays in state.State, and every
// action goes through the command API so rejected actions come back as messages.
type control struct {
	st     *state.State
	player state.Player

	selected bool
	from     hexgrid.Coord
	unit     *state.Unit
	moves    map[hexgrid.Coord]state.Route
	targets  map[hexgrid.Coord]bool

	menu       []WeaponOption // open weapon or content menu, nil when closed
	menuTarget hexgrid.Coord
	stored     []*Unit // set while the menu lists container contents; row i is stored[i]

	walk   *walk   // movement being animated; input waits until it is done
	report *report // battle panel of the last shot
	fx     effects // hit marks and banners

	reduced bool // reduced motion: instant walk, short static effects

	facing map[*state.Unit]unitsprite.Facing // presentation only: where each unit looks

	Message string // result of the last action, shown to the player
}

// roll is random.randint(-5, 5) of the Python reference.
func roll() int { return rand.IntN(11) - 5 }

// SetReducedMotion switches reduced motion on or off. Turning it on ends any
// walk and banner in progress.
func (c *control) SetReducedMotion(on bool) {
	c.reduced, c.fx.Reduced = on, on
	if on {
		c.walk = nil
		c.fx.Dismiss()
	}
}

// DismissReport hides the battle panel (map click or key press).
func (c *control) DismissReport() { c.report.Dismiss() }

func (c *control) reportTime() int {
	if c.reduced {
		return reportFramesReduced
	}
	return reportFrames
}

func (c *control) clear() {
	c.selected, c.unit, c.moves, c.targets, c.menu, c.stored = false, nil, nil, nil, nil, nil
}

// Facing is the authored direction u is drawn in; units start facing north.
// It is presentation state and never reaches the rules.
func (c *control) Facing(u *state.Unit) unitsprite.Facing { return c.facing[u] }

// turn makes u face from hex from towards hex to.
func (c *control) turn(u *state.Unit, from, to hexgrid.Coord) {
	x0, y0 := hexgrid.Center(from)
	x1, y1 := hexgrid.Center(to)
	f, ok := unitsprite.Toward(x1-x0, y1-y0)
	if !ok {
		return
	}
	if c.facing == nil {
		c.facing = map[*state.Unit]unitsprite.Facing{}
	}
	c.facing[u] = f
}

// mine reports whether it is the controlled player's turn.
func (c *control) mine() bool { return c.st.Turn == c.player && c.st.Winner == state.None }

// Select picks the own map unit at at and computes its overlays.
func (c *control) Select(at hexgrid.Coord) {
	c.clear()
	h := c.st.At(at)
	if h == nil || h.Unit == nil || h.Unit.Owner != c.player || c.st.Fog(c.player, at) != state.Visible {
		return
	}
	c.selected, c.from, c.unit = true, at, h.Unit
	c.moves = c.st.Movable(at, h.Unit)
	c.targets = map[hexgrid.Coord]bool{}
	if !c.st.Rules.OneActionPerTurn || !h.Unit.Moved {
		for _, t := range c.st.Targets(at) {
			c.targets[t] = true
		}
	}
}

// Click handles a left click on a map hex: select, move, attack or deselect.
func (c *control) Click(at hexgrid.Coord) {
	c.Message = ""
	c.fx.Dismiss()
	c.menu, c.stored = nil, nil
	if !c.walk.Done() {
		return
	}
	if !c.mine() {
		if c.st.Winner != state.None {
			c.Message = fmt.Sprintf("%v has won", c.st.Winner)
		}
		return
	}
	switch {
	case c.selected && c.targets[at]:
		c.OpenWeaponMenu(at)
	case c.selected && c.moves[at].Path != nil:
		c.move(at)
	case c.selected && at == c.from && c.contents(at) != nil:
		c.OpenContents(at)
	default:
		c.Select(at)
		if !c.selected && c.contents(at) != nil {
			c.OpenContents(at)
		}
	}
}

// contents returns the container of an own, visible building or carrier at at.
func (c *control) contents(at hexgrid.Coord) *state.Container {
	h := c.st.At(at)
	if h == nil || c.st.Fog(c.player, at) != state.Visible {
		return nil
	}
	switch {
	case h.Unit != nil:
		if h.Unit.Owner == c.player {
			return h.Unit.Content
		}
	case h.Container != nil && h.Owner == c.player:
		return h.Container
	}
	return nil
}

// Unit is the rules unit, so the stored list reads naturally in this file.
type Unit = state.Unit

// OpenContents lists the units stored at at (the Python 2 × 9 grid, flattened)
// with the weight used; a row is usable when the stored unit may leave now.
func (c *control) OpenContents(at hexgrid.Coord) {
	box := c.contents(at)
	if box == nil {
		return
	}
	c.menuTarget, c.stored, c.menu = at, box.Units(), nil
	if len(c.stored) == 0 {
		c.Message = "nothing stored here"
		c.stored = nil
		return
	}
	c.Message = fmt.Sprintf("contents: weight %d of %d", box.Used(), box.Capacity)
	for i, u := range c.stored {
		o := WeaponOption{Slot: i, Usable: true, Label: fmt.Sprintf("%s (size %d)", u.Type, u.Size)}
		if len(c.st.Movable(at, u)) == 0 {
			o.Label, o.Usable = o.Label+": cannot leave now", false
		}
		c.menu = append(c.menu, o)
	}
}

// ChooseMenu answers menu row i for whichever menu is open.
func (c *control) ChooseMenu(i int) {
	if c.stored == nil {
		c.ChooseWeapon(i)
		return
	}
	if i < 0 || i >= len(c.menu) {
		return
	}
	if o := c.menu[i]; !o.Usable {
		c.Message = o.Label
		return
	}
	at, u := c.menuTarget, c.stored[i]
	c.clear()
	c.selected, c.from, c.unit = true, at, u
	c.moves, c.targets = c.st.Movable(at, u), map[hexgrid.Coord]bool{}
	c.Message = fmt.Sprintf("%s leaves: pick a hex", u.Type)
}

func (c *control) move(to hexgrid.Coord) {
	dst := c.st.At(to)
	before := dst.Owner
	path, err := c.st.Move(c.from, c.unit, to)
	if err != nil {
		c.Message = err.Error()
		return
	}
	c.walk = newWalkMotion(path, c.reduced)
	if n := len(path); n >= 2 {
		c.turn(c.unit, path[n-2], path[n-1])
	}
	c.Message = fmt.Sprintf("%s moved", c.unit.Type)
	if dst.Container != nil && dst.Owner != before && c.st.Winner == state.None {
		// The engine flips the depot owner when a unit enters it; report that.
		c.Message = fmt.Sprintf("%s is capturing the depot", c.unit.Type)
		c.fx.Capture(c.unit.Type)
	}
	if c.st.Winner != state.None {
		c.Message = fmt.Sprintf("%v captured the HQ and wins", c.st.Winner)
		c.fx.Victory(c.st.Winner)
	}
	c.Select(to)
}

// WeaponOption is one row of the weapon menu: the slot, what the player should
// see and whether the rules would accept a shot at the chosen target.
type WeaponOption struct {
	Slot   int
	Label  string
	Usable bool
}

// weaponOptions lists the unit's weapons against the target at to, each with the
// reason it cannot fire when it cannot (range, ammunition, target kind).
func (c *control) weaponOptions(to hexgrid.Coord) []WeaponOption {
	var out []WeaponOption
	th := c.st.At(to)
	if c.unit == nil || th == nil || th.Unit == nil {
		return nil
	}
	steps := hexgrid.Distance(c.from, to)
	for slot, w := range c.unit.Weapons {
		if w == nil {
			continue
		}
		o := WeaponOption{Slot: slot, Usable: true, Label: w.Name}
		if n := c.unit.Ammo[slot]; n != nil {
			o.Label += fmt.Sprintf(" (ammo %d)", *n)
		}
		switch {
		case c.st.Rules.AmmoPerShot && c.unit.Ammo[slot] != nil && *c.unit.Ammo[slot] == 0:
			o.Label, o.Usable = o.Label+": no ammo", false
		case steps < max(w.RangeMin, 1) || steps > w.RangeMax:
			o.Label, o.Usable = o.Label+fmt.Sprintf(": range %d-%d", w.RangeMin, w.RangeMax), false
		case !c.st.CanFire(c.unit, slot, th.Unit):
			o.Label, o.Usable = o.Label+": cannot hit this target", false
		}
		out = append(out, o)
	}
	return out
}

// OpenWeaponMenu starts the weapon choice for an attack on to.
func (c *control) OpenWeaponMenu(to hexgrid.Coord) {
	c.menuTarget, c.menu = to, c.weaponOptions(to)
}

// MenuOpen reports whether the weapon menu waits for a choice.
func (c *control) MenuOpen() bool { return len(c.menu) > 0 }

// ChooseWeapon fires menu row i and closes the menu; an unusable row only
// reports why it is unavailable and keeps the menu open.
func (c *control) ChooseWeapon(i int) {
	if i < 0 || i >= len(c.menu) {
		return
	}
	o := c.menu[i]
	if !o.Usable {
		c.Message = o.Label
		return
	}
	c.menu = nil
	c.attack(c.menuTarget, o.Slot)
}

func (c *control) attack(to hexgrid.Coord, slot int) {
	atk, def := c.st.At(c.from).Unit, c.st.At(to).Unit
	rp := &report{frames: c.reportTime(),
		Attacker: side{Name: atk.Type, Weapon: atk.Weapons[slot].Name, Rank: atk.Experience, Before: atk.Size, Max: atk.MaxSize},
		Defender: side{Name: def.Type, Rank: def.Experience, Before: def.Size, Max: def.MaxSize}}
	b, err := c.st.Attack(c.from, to, slot, roll)
	if err != nil {
		c.Message = err.Error()
		return
	}
	rp.Attacker.After, rp.Defender.After = atk.Size, def.Size
	if b.Counter >= 0 && def.Weapons[b.Counter] != nil {
		rp.Defender.Weapon = def.Weapons[b.Counter].Name
	}
	c.report = rp
	c.fx.AddBattle(c.from, to, b)
	c.turn(atk, c.from, to)
	if b.DefenderDestroyed {
		delete(c.facing, def)
	} else {
		c.turn(def, to, c.from)
	}
	if b.AttackerDestroyed {
		delete(c.facing, atk)
	}
	c.Message = fmt.Sprintf("hit for %d, took %d back", b.Damage, b.CounterDamage)
	switch {
	case b.DefenderDestroyed:
		c.Message += "; target destroyed"
	case b.AttackerDestroyed:
		c.Message += "; attacker destroyed"
	}
	c.clear()
	if !b.AttackerDestroyed {
		c.Select(c.from)
	}
}

// EndTurn ends the player's turn. With no AI connected yet the opponent passes at
// once, so the game stays playable for a single human.
func (c *control) EndTurn() {
	c.clear()
	c.walk = nil
	if !c.mine() {
		return
	}
	if err := c.st.EndTurn(); err != nil {
		c.Message = err.Error()
		return
	}
	for c.st.Turn != c.player && c.st.Winner == state.None {
		if err := c.st.EndTurn(); err != nil {
			c.Message = err.Error()
			return
		}
	}
	c.Message = fmt.Sprintf("round %d (the opponent passed: no AI yet)", c.st.Round)
	c.fx.Round(c.st.Round, c.st.Turn)
	if c.st.Winner != state.None {
		c.fx.Victory(c.st.Winner)
	}
}
