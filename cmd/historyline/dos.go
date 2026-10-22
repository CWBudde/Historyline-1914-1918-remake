package main

import (
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosai"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// The original order cycle (-rules dos): both sides give their orders at the
// same time. The side in the movement phase moves its units (each move runs
// at once on its own map), the side in the action phase queues attack orders;
// both confirm (F1). The resolution fights the action side's orders, brings
// both maps into line and swaps the phases. A side sees the other side's
// units where they stood at the last phase change.

// dosPlanner gives the orders of a computer side for the phase it is in and
// confirms them.
type dosPlanner interface {
	Plan(s *dosplay.Session, side int) ([]dosplay.Event, error)
}

// passPlanner confirms without giving any order.
type passPlanner struct{}

func (passPlanner) Plan(s *dosplay.Session, side int) ([]dosplay.Event, error) {
	return s.Confirm(side), nil
}

// computerPlanner is the computer side of -opponent ai.
func computerPlanner() dosPlanner { return &dosai.Planner{} }

// dosControl turns clicks into orders of the human side of a session and
// plays the resolution back. Every order goes through the session, so a
// refused one comes back as a message.
type dosControl struct {
	presentation

	s       *dosplay.Session
	side    int        // the human side
	planner dosPlanner // the other side
	reveal  bool       // dev: the frozen view is the true view

	unit       int                    // live record of the selected unit, -1 none
	hold, slot int                    // container record and slot of a unit picked to leave it, hold -1 otherwise
	moves      map[hexgrid.Coord]bool // blue hexes of the selected unit
	stored     []dosplay.UnitView     // contents menu rows
	storedHold int                    // their container record

	queue       []dosplay.Event // resolution events still to play
	frozen      *dosplay.View   // the side's view from before the resolution while its battles play
	otherFrozen *dosplay.View   // the other side's view, frozen alike
	banner      string          // phase banner, shown once the revealed moves have played
	over        bool            // the end of the battle has played

	views  [2]sideCache      // each side's map, by side
	follow [2]*hexgrid.Coord // where each side's panel should look next; the front end takes it
}

// sideCache is a side's view at a session version and the units drawn on its
// panel for a shownKey.
type sideCache struct {
	view    *dosplay.View
	ver     int
	shown   map[hexgrid.Coord]dosplay.UnitView
	shownAt shownKey
}

type shownKey struct {
	version, pending int
	reveal           bool
}

var _ driver = (*dosControl)(nil)

func newDosControl(s *dosplay.Session, side int, planner dosPlanner, reveal bool) *dosControl {
	if planner == nil {
		planner = passPlanner{}
	}
	d := &dosControl{s: s, side: side, planner: planner, reveal: reveal}
	d.clear()
	return d
}

func (d *dosControl) clear() {
	d.clearSelection()
	d.unit, d.hold, d.slot, d.moves, d.stored = -1, -1, -1, nil, nil
}

// Clear drops the selection (right click).
func (d *dosControl) Clear() { d.clear() }

// Cancel finishes the walk in progress, else closes the battle panel, else
// drops the selection (Esc).
func (d *dosControl) Cancel() {
	switch {
	case d.SkipWalk():
	case d.report.Active():
		d.DismissReport()
	default:
		d.clear()
	}
}

// own is the human side's view, recomputed when the session changes.
func (d *dosControl) own() *dosplay.View { return d.viewOf(d.side) }

// viewOf is side's view, recomputed when the session changes.
func (d *dosControl) viewOf(side int) *dosplay.View {
	c := &d.views[side]
	if c.view == nil || c.ver != d.s.Version() {
		v := d.s.View(side)
		c.view, c.ver = &v, d.s.Version()
	}
	return c.view
}

// frozenOf is side's frozen view while a resolution plays, else nil.
func (d *dosControl) frozenOf(side int) *dosplay.View {
	if side == d.side {
		return d.frozen
	}
	return d.otherFrozen
}

// unfreeze drops both frozen views.
func (d *dosControl) unfreeze() { d.frozen, d.otherFrozen = nil, nil }

// lookAt asks side's panel to look at c.
func (d *dosControl) lookAt(side int, c hexgrid.Coord) {
	if side == 0 || side == 1 {
		d.follow[side] = &c
	}
}

// takeFollow returns and clears where side's panel should look next.
func (d *dosControl) takeFollow(side int) (hexgrid.Coord, bool) {
	c := d.follow[side]
	d.follow[side] = nil
	if c == nil {
		return hexgrid.Coord{}, false
	}
	return *c, true
}

// planFocus is where a computer side acted in its orders: the goal of its last
// move, else the attacker of its last attack order.
func planFocus(evs []dosplay.Event) (hexgrid.Coord, bool) {
	for i := len(evs) - 1; i >= 0; i-- {
		if ev := evs[i]; ev.Kind == dosplay.Moved && len(ev.Path) > 0 {
			return ev.Path[len(ev.Path)-1], true
		}
	}
	for i := len(evs) - 1; i >= 0; i-- {
		if ev := evs[i]; ev.Kind == dosplay.Ordered {
			return ev.From, true
		}
	}
	return hexgrid.Coord{}, false
}

// busy reports whether a resolution is being played back.
func (d *dosControl) busy() bool { return len(d.queue) > 0 || d.frozen != nil || d.banner != "" }

// phaseName names a phase.
func phaseName(p int) string {
	if p == dosrules.PhaseAction {
		return "action"
	}
	return "movement"
}

// reason is the text of a refused order, as a clause.
func reason(err error) string {
	text := err.Error()
	var oe *dosplay.OrderError
	if errors.As(err, &oe) {
		text = oe.Text
	}
	text = strings.TrimSuffix(text, ".")
	if text != "" {
		text = strings.ToLower(text[:1]) + text[1:]
	}
	return text
}

// Click handles a left click on a map hex: select an own unit, building or
// carrier, move to a blue hex or order an attack on a red one. A click during
// a walk only finishes the walk.
func (d *dosControl) Click(at hexgrid.Coord) {
	if d.SkipWalk() {
		return
	}
	d.Message = ""
	d.fx.Dismiss()
	switch {
	case d.over:
		d.Message = d.endText()
		return
	case d.busy():
		d.Message = "the orders are being carried out"
		return
	case d.s.Ready(d.side):
		d.Message = "orders confirmed: F1 opens them again"
		return
	}
	d.menu = nil
	switch {
	case d.selected && d.moves[at]:
		d.move(at)
	case d.selected && d.targets[at]:
		d.attack(at)
	default:
		d.selectAt(at)
	}
}

// selectAt selects the own unit on at, or opens the contents of the own
// building there.
func (d *dosControl) selectAt(at hexgrid.Coord) {
	d.clear()
	v := d.own()
	if u, ok := v.Units[at]; ok && u.Side == d.side {
		d.selectUnit(at, u)
		return
	}
	if b, ok := v.Buildings[at]; ok && b.Owner == d.s.Faction(d.side) {
		d.openContents(at)
	}
}

// selectUnit selects u on at with its blue hexes in the movement phase or its
// red hexes in the action phase. A carrier also lists its contents.
func (d *dosControl) selectUnit(at hexgrid.Coord, u dosplay.UnitView) {
	err := d.s.CanOrder(d.side, u.Index)
	if err == nil {
		d.selected, d.from, d.unit = true, at, u.Index
		if d.s.Phase(d.side) == dosrules.PhaseAction {
			d.targets = d.s.Targets(d.side, u.Index)
			d.Message = u.Name + ": pick a red hex to attack"
			if len(d.targets) == 0 {
				d.Message = u.Name + ": no target in range"
			}
		} else {
			d.moves, err = d.s.Destinations(d.side, u.Index)
			d.Message = u.Name + ": pick a blue hex to move to"
			if err == nil && len(d.moves) == 0 {
				d.Message = u.Name + ": nowhere to go"
			}
		}
	}
	if err != nil {
		d.Message = u.Name + ": " + reason(err)
	}
	if u.Carrier && d.s.Phase(d.side) == dosrules.PhaseMovement {
		d.openContents(at)
	}
}

// openContents lists the units of the own building or carrier on at; a row
// is usable when its unit may leave now.
func (d *dosControl) openContents(at hexgrid.Coord) {
	hold, units, ok := d.s.Contents(d.side, at)
	if !ok {
		return
	}
	if len(units) == 0 {
		if !d.selected {
			d.Message = "nothing stored here"
		}
		return
	}
	d.menuKind, d.menuTarget, d.storedHold, d.stored, d.menu = menuContents, at, hold, units, nil
	for i, u := range units {
		o := WeaponOption{Slot: i, Usable: true, Label: fmt.Sprintf("%s (size %d/%d)", u.Name, u.Size, u.Full)}
		if why := d.cannotLeave(hold, u); why != "" {
			o.Label, o.Usable = o.Label+": "+why, false
		}
		d.menu = append(d.menu, o)
	}
	if !d.selected {
		d.Message = fmt.Sprintf("contents: %d units", len(units))
	}
}

// cannotLeave is why the held unit u cannot leave container record hold now,
// empty when it can.
func (d *dosControl) cannotLeave(hold int, u dosplay.UnitView) string {
	if d.s.Phase(d.side) != dosrules.PhaseMovement {
		return "leaves in the movement phase"
	}
	dests, err := d.s.ExitDestinations(d.side, hold, u.Slot)
	switch {
	case err != nil:
		return reason(err)
	case len(dests) == 0:
		return "nowhere to go"
	}
	return ""
}

// ChooseMenu picks contents row i: its unit leaves, to a blue hex.
func (d *dosControl) ChooseMenu(i int) {
	if d.menuKind != menuContents || i < 0 || i >= len(d.menu) || i >= len(d.stored) {
		return
	}
	if o := d.menu[i]; !o.Usable {
		d.Message = o.Label
		return
	}
	u, at, hold := d.stored[i], d.menuTarget, d.storedHold
	dests, err := d.s.ExitDestinations(d.side, hold, u.Slot)
	if err != nil {
		d.Message = u.Name + ": " + reason(err)
		return
	}
	d.clear()
	d.selected, d.from, d.unit, d.hold, d.slot, d.moves = true, at, u.Index, hold, u.Slot, dests
	d.targets = map[hexgrid.Coord]bool{}
	d.Message = u.Name + " leaves: pick a blue hex"
}

// move orders the selected unit to to: a move on the map, or the exit from
// its container. It runs at once and walks.
func (d *dosControl) move(to hexgrid.Coord) {
	var evs []dosplay.Event
	var err error
	if d.hold >= 0 {
		evs, err = d.s.Exit(d.side, d.hold, d.slot, to)
	} else {
		evs, err = d.s.Move(d.side, d.unit, to)
	}
	d.clear()
	if err != nil {
		d.Message = reason(err)
		return
	}
	for _, ev := range evs {
		switch ev.Kind {
		case dosplay.Moved:
			d.walkAlong(ev)
			d.Message = d.s.Unit(ev.Unit).Name + " moved"
		case dosplay.Refused:
			d.Message = ev.Text
		default:
		}
	}
}

// walkAlong walks the unit of a move event along its path and turns it the
// way it went.
func (d *dosControl) walkAlong(ev dosplay.Event) {
	d.walk = newWalkMotion(ev.Path, ev.Unit, d.reduced)
	if n := len(ev.Path); n >= 2 {
		d.turn(ev.Unit, ev.Path[n-2], ev.Path[n-1])
	}
}

// attack queues the attack of the selected unit on to.
func (d *dosControl) attack(to hexgrid.Coord) {
	evs, err := d.s.Attack(d.side, d.unit, to)
	d.clear()
	if err != nil {
		d.Message = reason(err)
		return
	}
	for _, ev := range evs {
		if ev.Kind == dosplay.Ordered {
			d.turn(ev.Unit, ev.From, ev.To)
			d.Message = fmt.Sprintf("%s (order %d)", strings.TrimSuffix(ev.Text, "."), len(d.s.Orders(d.side)))
		}
	}
}

// EndTurn (F1) confirms the side's orders, or opens them again while the
// other side has not confirmed yet.
func (d *dosControl) EndTurn() {
	d.clear()
	switch {
	case d.over:
		d.Message = d.endText()
	case d.busy():
		d.Message = "the orders are being carried out"
	case !d.s.Ready(d.side):
		d.s.Confirm(d.side)
		d.Message = "orders confirmed - waiting for the opponent"
	default:
		d.s.Unconfirm(d.side)
		d.Message = "orders open again"
	}
}

// Tick advances the presentation by one frame. The computer side plans as
// soon as its phase begins; once both sides have confirmed the orders are
// resolved and the events played back, one per step.
func (d *dosControl) Tick() {
	d.tick()
	switch {
	case d.over:
	case d.busy():
		d.step()
	case d.s.Ended() != nil:
		d.over = true
		d.Message = d.endText()
	default:
		if opp := 1 - d.side; !d.s.Ready(opp) {
			d.plan(opp)
		}
		if d.s.Ready(0) && d.s.Ready(1) {
			d.resolve()
		}
	}
}

// plan lets the computer give side's orders. A side the planner leaves
// unconfirmed (after an error) is confirmed, so the battle goes on.
func (d *dosControl) plan(side int) {
	evs, err := d.planner.Plan(d.s, side)
	if err != nil {
		d.Message = "computer side: " + err.Error()
	}
	if c, ok := planFocus(evs); ok {
		d.lookAt(side, c)
	}
	if !d.s.Ready(side) {
		d.s.Confirm(side)
	}
}

// resolve freezes both sides' views and resolves the orders; the events wait
// in the queue.
func (d *dosControl) resolve() {
	d.clear()
	v, o := d.s.View(d.side), d.s.View(1-d.side)
	if d.reveal {
		v, o = d.s.TrueView(), d.s.TrueView()
	}
	evs, err := d.s.Resolve()
	if err != nil {
		d.Message = "resolution: " + err.Error()
	}
	d.frozen, d.otherFrozen, d.queue = &v, &o, evs
}

// step plays the next event once the last one has played out on screen
// (walk finished, battle panel closed).
func (d *dosControl) step() {
	if !d.walk.Done() || d.report.Active() {
		return
	}
	if d.banner != "" && (len(d.queue) == 0 || d.queue[0].Kind != dosplay.Revealed) {
		if !d.endQueued() {
			d.fx.Show(d.banner, roundFrames)
		}
		d.banner = ""
	}
	if len(d.queue) == 0 {
		d.unfreeze()
		return
	}
	ev := d.queue[0]
	d.queue = d.queue[1:]
	switch ev.Kind {
	case dosplay.Battle:
		d.playBattle(ev)
	case dosplay.Skipped:
		d.Message = fmt.Sprintf("an attack order on hex %d,%d found no unit", ev.To.Line, ev.To.Col)
	case dosplay.PhaseChanged:
		d.unfreeze()
		d.banner = fmt.Sprintf("TURN %d - %v: %s phase", ev.Turn+1, d.s.Faction(d.side), phaseName(ev.Phase[d.side]))
	case dosplay.Revealed:
		if ev.Side != d.side {
			d.walkAlong(ev)
			if len(ev.Path) > 0 {
				d.lookAt(ev.Side, ev.Path[len(ev.Path)-1])
			}
		}
	case dosplay.Ended:
		d.over = true
		d.Message = d.endText()
		d.fx.Show(d.endText(), victoryFrames)
	default:
	}
}

// endQueued reports whether the end of the battle waits in the queue.
func (d *dosControl) endQueued() bool {
	for _, ev := range d.queue {
		if ev.Kind == dosplay.Ended {
			return true
		}
	}
	return false
}

// playBattle shows a fought attack order: the battle panel, hit marks at the
// fighters' hexes in the frozen view, facings, and the destroyed units leave
// the frozen view.
func (d *dosControl) playBattle(ev dosplay.Event) {
	b := ev.Battle
	if b == nil {
		return
	}
	atk, def := b.Attacker, b.Defender
	from, to := d.frozenAt(atk.Unit.Index, ev.From), d.frozenAt(def.Unit.Index, ev.To)
	reply := ""
	if b.ReturnFire {
		reply = "fires back"
	}
	d.report = &report{frames: d.reportTime(), Attacker: fighterSide(atk, "attacks"), Defender: fighterSide(def, reply)}
	d.fx.AddHits(from, to, hit(def), def.Destroyed, hit(atk), atk.Destroyed)
	d.turn(atk.Unit.Index, from, to)
	d.turn(def.Unit.Index, to, from)
	other := 1 - d.side
	switch other {
	case atk.Unit.Side:
		d.lookAt(other, from)
	case def.Unit.Side:
		d.lookAt(other, to)
	}
	d.Message = fmt.Sprintf("%s attacked %s: %d -> %d", atk.Unit.Name, def.Unit.Name, def.Before, def.After)
	for _, f := range []dosplay.Fighter{def, atk} {
		if f.Destroyed {
			d.Message += "; " + f.Unit.Name + " destroyed"
			d.drop(f.Unit.Index)
		}
	}
}

// hit reports whether a fighter lost strength.
func hit(f dosplay.Fighter) bool { return f.After < f.Before || f.AltAfter < f.AltBefore }

// fighterSide is a fighter in the battle panel; action is what it did.
func fighterSide(f dosplay.Fighter, action string) side {
	if action != "" && f.ExpAfter != f.ExpBefore {
		action += fmt.Sprintf(", exp %d -> %d", f.ExpBefore, f.ExpAfter)
	}
	return side{Name: f.Unit.Name, Weapon: action, Rank: f.ExpBefore, Before: f.Before, After: f.After, Max: f.Unit.Full}
}

// frozenAt is the hex unit stands on in the frozen view, else at.
func (d *dosControl) frozenAt(unit int, at hexgrid.Coord) hexgrid.Coord {
	if d.frozen != nil {
		for p, u := range d.frozen.Units {
			if u.Index == unit {
				return p
			}
		}
	}
	return at
}

// drop takes a destroyed unit out of both frozen views.
func (d *dosControl) drop(unit int) {
	delete(d.facing, unit)
	for _, v := range []*dosplay.View{d.frozen, d.otherFrozen} {
		if v != nil {
			maps.DeleteFunc(v.Units, func(_ hexgrid.Coord, u dosplay.UnitView) bool { return u.Index == unit })
		}
	}
}

// endText says how the battle ended, from the human side's view.
func (d *dosControl) endText() string {
	e := d.s.Ended()
	if e == nil {
		return ""
	}
	switch {
	case e.Winner != 0 && e.Winner != 1:
		return "The battle ends in a draw"
	case e.Winner == d.side:
		return fmt.Sprintf("VICTORY - %s (score %d)", d.endReason(), e.Score)
	}
	return "DEFEAT - " + d.endReason()
}

// endReason says who won and how, "" while the battle is open or after a draw.
func (d *dosControl) endReason() string {
	e := d.s.Ended()
	if e == nil || e.Winner != 0 && e.Winner != 1 {
		return ""
	}
	how := "destroyed every enemy unit"
	if e.Won[e.Winner] == dosrules.EndHQ {
		how = "hold the enemy HQ"
	}
	return fmt.Sprintf("%v %s", d.s.Faction(e.Winner), how)
}

// units are the units drawn on the human side's panel (see unitsOf).
func (d *dosControl) units(reveal bool) map[hexgrid.Coord]dosplay.UnitView {
	return d.unitsOf(d.side, reveal)
}

// unitsOf are the units drawn on side's panel, by hex: its frozen view while
// a resolution plays, else its view (the true view with reveal) with the other
// side's moves still to be revealed at their start.
func (d *dosControl) unitsOf(side int, reveal bool) map[hexgrid.Coord]dosplay.UnitView {
	if f := d.frozenOf(side); f != nil {
		return f.Units
	}
	pending := 0
	for _, ev := range d.queue {
		if ev.Kind == dosplay.Revealed && ev.Side != side {
			pending++
		}
	}
	c := &d.views[side]
	key := shownKey{version: d.s.Version(), pending: pending, reveal: reveal}
	if c.shown != nil && c.shownAt == key {
		return c.shown
	}
	v := d.viewOf(side)
	if reveal {
		tv := d.s.TrueView()
		v = &tv
	}
	units := v.Units
	if pending > 0 {
		units = maps.Clone(units)
		for _, ev := range d.queue {
			if ev.Kind != dosplay.Revealed || ev.Side == side || len(ev.Path) == 0 {
				continue
			}
			u := d.s.Unit(ev.Unit)
			for p, w := range units {
				if w.Index == ev.Unit {
					u = w
					delete(units, p)
				}
			}
			if _, inside := v.Buildings[ev.Path[0]]; !inside {
				u.At = ev.Path[0]
				units[u.At] = u
			}
		}
	}
	c.shown, c.shownAt = units, key
	return units
}

// buildings are the buildings drawn on the human side's panel, by entrance.
func (d *dosControl) buildings() map[hexgrid.Coord]dosplay.BuildingView {
	return d.buildingsOf(d.side)
}

// buildingsOf are the buildings drawn on side's panel, by entrance.
func (d *dosControl) buildingsOf(side int) map[hexgrid.Coord]dosplay.BuildingView {
	if f := d.frozenOf(side); f != nil {
		return f.Buildings
	}
	return d.viewOf(side).Buildings
}
