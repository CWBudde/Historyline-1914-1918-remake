package main

import (
	"fmt"
	"math/rand/v2"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/ai"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// control turns hex clicks into rules commands for one player. It holds only the
// selection and its overlays; all game state stays in state.State, and every
// action goes through the command API so rejected actions come back as messages.
// The engine-neutral UI state lives in the embedded presentation.
type control struct {
	presentation

	st     *state.State
	player state.Player

	unit   *state.Unit
	moves  map[hexgrid.Coord]state.Route
	stored []*Unit // set while the menu lists container contents; row i is stored[i]

	ai                 *ai.Planner // computer opponent; nil passes every turn
	aiMoves, aiAttacks int         // orders of the computer's current turn
}

// roll is random.randint(-5, 5) of the Python reference.
func roll() int { return rand.IntN(11) - 5 }

func (c *control) clear() {
	c.clearSelection()
	c.unit, c.moves, c.stored = nil, nil, nil
}

// Clear drops the selection (right click).
func (c *control) Clear() { c.clear() }

// Cancel finishes the walk in progress, or else drops the selection (Esc).
func (c *control) Cancel() {
	if !c.SkipWalk() {
		c.clear()
	}
}

// Tick advances the presentation by one frame and lets the computer side play
// its next order.
func (c *control) Tick() {
	c.tick()
	c.StepAI()
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
	if !h.Unit.Moved && !h.Unit.Attacked {
		for _, t := range c.st.Targets(at) {
			c.targets[t] = true
		}
	}
}

// Click handles a left click on a map hex: select, move, attack or deselect.
// Selecting an own unit opens its action menu; a click on one of its blue or
// red hexes acts at once. A click during a walk only finishes the walk.
func (c *control) Click(at hexgrid.Coord) {
	if c.SkipWalk() {
		return
	}
	c.Message = ""
	c.fx.Dismiss()
	c.menu, c.stored = nil, nil
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
		switch {
		case c.selected:
			c.openActions()
		case c.contents(at) != nil:
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
	c.menuKind, c.menuTarget, c.stored, c.menu = menuContents, at, box.Units(), nil
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
	switch c.menuKind {
	case menuWeapons:
		c.ChooseWeapon(i)
		return
	case menuActions:
		c.chooseAction(i)
		return
	case menuContents:
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
	c.openActions()
}

func (c *control) move(to hexgrid.Coord) {
	dst := c.st.At(to)
	before := dst.Owner
	path, err := c.st.Move(c.from, c.unit, to)
	if err != nil {
		c.Message = err.Error()
		return
	}
	c.walk = newWalkMotion(path, c.unit, c.reduced)
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
		case c.unit.Ammo[slot] != nil && *c.unit.Ammo[slot] == 0:
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
	c.menuKind, c.menuTarget, c.menu = menuWeapons, to, c.weaponOptions(to)
}

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
	b, err := c.fire(c.from, to, slot, false)
	if err != nil {
		c.Message = err.Error()
		return
	}
	c.Message = battleText(b, "target", "attacker")
	c.clear()
	if !b.AttackerDestroyed {
		c.Select(c.from)
	}
}

// fire resolves a shot of the map unit at from at to and shows it: the battle
// panel, hit marks and facings. hidden means the player cannot see the shooter,
// so the panel tells nothing about it.
func (c *control) fire(from, to hexgrid.Coord, slot int, hidden bool) (state.Battle, error) {
	atk, def := c.st.At(from).Unit, c.st.At(to).Unit
	rp := &report{
		frames:   c.reportTime(),
		Attacker: side{Name: atk.Type, Weapon: atk.Weapons[slot].Name, Rank: atk.Experience, Before: atk.Size, Max: atk.MaxSize},
		Defender: side{Name: def.Type, Rank: def.Experience, Before: def.Size, Max: def.MaxSize},
	}
	b, err := c.st.Attack(from, to, slot, roll)
	if err != nil {
		return b, err
	}
	rp.Attacker.After, rp.Defender.After = atk.Size, def.Size
	if b.Counter >= 0 && def.Weapons[b.Counter] != nil {
		rp.Defender.Weapon = def.Weapons[b.Counter].Name
	}
	if hidden {
		rp.Attacker = side{Name: "unknown", Hidden: true}
	}
	c.report = rp
	c.fx.AddBattle(from, to, b)
	c.turn(atk, from, to)
	if b.DefenderDestroyed {
		delete(c.facing, def)
	} else {
		c.turn(def, to, from)
	}
	if b.AttackerDestroyed {
		delete(c.facing, atk)
	}
	return b, nil
}

// battleText sums a battle up, naming the two units as given.
func battleText(b state.Battle, defender, attacker string) string {
	msg := fmt.Sprintf("hit for %d, took %d back", b.Damage, b.CounterDamage)
	switch {
	case b.DefenderDestroyed:
		msg += "; " + defender + " destroyed"
	case b.AttackerDestroyed:
		msg += "; " + attacker + " destroyed"
	}
	return msg
}

// EndTurn ends the player's turn. The opponent then plays: the computer one
// order at a time through StepAI, or, without one, it passes at once.
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
	c.aiMoves, c.aiAttacks = 0, 0
	if c.ai == nil {
		c.pass()
		return
	}
	c.Message = "enemy turn"
}

// pass ends the opponent's turns until it is the player's again.
func (c *control) pass() {
	for c.st.Turn != c.player && c.st.Winner == state.None {
		if err := c.st.EndTurn(); err != nil {
			c.Message = err.Error()
			return
		}
	}
	c.Message = fmt.Sprintf("round %d (the opponent passed: no AI)", c.st.Round)
	c.newRound()
}

// newRound shows the banner of the round the player now moves in.
func (c *control) newRound() {
	c.fx.Round(c.st.Round, c.st.Turn)
	if c.st.Winner != state.None {
		c.fx.Victory(c.st.Winner)
	}
}

// AITurn reports whether the opponent is playing its turn.
func (c *control) AITurn() bool {
	return c.st.Turn != c.player && c.st.Winner == state.None
}

// StepAI plays the opponent's next order once the last one has played out on
// screen (walk finished, battle panel closed). When the computer has nothing
// left to do the turn passes back to the player.
func (c *control) StepAI() {
	if !c.AITurn() || !c.walk.Done() || c.report.Active() {
		return
	}
	if c.ai == nil {
		c.pass()
		return
	}
	o, ok := c.ai.Next(c.st)
	if !ok {
		if err := c.st.EndTurn(); err != nil {
			c.Message = err.Error()
			return
		}
		c.Message = fmt.Sprintf("round %d: the enemy made %s and %s", c.st.Round, count(c.aiMoves, "move"), count(c.aiAttacks, "attack"))
		c.newRound()
		return
	}
	if o.Attack {
		c.aiFire(o)
	} else {
		c.aiMove(o)
	}
}

// count is "n thing(s)".
func count(n int, thing string) string {
	if n == 1 {
		return "1 " + thing
	}
	return fmt.Sprintf("%d %ss", n, thing)
}

// seen reports whether the player can see the hex at at now.
func (c *control) seen(at hexgrid.Coord) bool { return c.st.Fog(c.player, at) == state.Visible }

// aiMove executes a computer move. The walk is shown only when the player can
// see part of the path; the map view draws it only on visible hexes.
func (c *control) aiMove(o ai.Order) {
	dst := c.st.At(o.To)
	before := dst.Owner
	path, err := c.st.Move(o.From, o.Unit, o.To)
	if err != nil {
		return // the planner has marked the unit; it is skipped
	}
	c.aiMoves++
	if n := len(path); n >= 2 {
		c.turn(o.Unit, path[n-2], path[n-1])
	}
	if slices.ContainsFunc(path, c.seen) {
		c.walk = newWalkMotion(path, o.Unit, c.reduced)
	}
	switch {
	case c.st.Winner != state.None:
		c.Message = fmt.Sprintf("%v captured the HQ and wins", c.st.Winner)
		c.fx.Victory(c.st.Winner)
	case dst.Container != nil && dst.Owner != before && before == c.player:
		c.Message = fmt.Sprintf("the enemy took your building at line %d col %d", o.To.Line, o.To.Col)
	}
}

// aiFire executes a computer attack and shows it like the player's own.
func (c *control) aiFire(o ai.Order) {
	b, err := c.fire(o.From, o.To, o.Slot, !c.seen(o.From))
	if err != nil {
		return
	}
	c.aiAttacks++
	c.Message = "enemy attack: " + battleText(b, "your unit", "attacker")
}
