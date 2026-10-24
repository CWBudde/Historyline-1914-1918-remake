package main

import (
	"fmt"
	"image"
	"slices"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dospanel"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/gesture"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/pixfont"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

// The contents of -rules dos: the open gesture (or a click) on an own building
// or carrier opens its contents in the human side's panel: the container's
// plate with its energy, the side's seven slots and a cursor on them. The
// arrows, a drag or the mouse move the cursor. Holding fire shows the plate of
// the contents' gestures on the slot: move out above, produce to the left,
// cancel to the right, repair below. A plain fire moves the unit out. Produce
// lists the types the factory builds with their prices in the details plate:
// up and down pick (the mouse picks the row under it, the wheel scrolls), fire
// builds, cancel and Esc go back to the slots.

// buildRows is how many types the production list shows below its heading.
const buildRows = 5

// contents is the open contents of an own container.
type contents struct {
	at    hexgrid.Coord // the container's hex
	hold  int           // its container record
	sel   int           // the slot under the cursor
	build bool          // the production list is open
	pick  int           // its picked row
	top   int           // its first row shown
	cache boxCache
}

// boxCache is the placement and the details of the open contents at a key.
type boxCache struct {
	key     boxKey
	ok      bool
	place   gesture.Placement
	details []string
}

// boxKey is what the placement and the details depend on.
type boxKey struct {
	version, energy, sel, pick, top int
	build, ready, over, busy        bool
}

// openContents opens the contents of the own building or carrier on at with
// the cursor on its first unit. A container with no unit and nothing to
// build stays closed.
func (d *dosControl) openContents(at hexgrid.Coord) {
	hold, units, ok := d.s.Contents(d.side, at)
	if !ok {
		return
	}
	if len(units) == 0 && len(d.s.Buildable(d.side, hold)) == 0 {
		d.Message = "nothing stored here"
		return
	}
	d.box = &contents{at: at, hold: hold}
	if len(units) > 0 {
		d.box.sel = units[0].Slot
	}
	d.Message = d.boxName() + ": " + count(len(units), "unit")
}

// boxSlots are the side's seven slots of the open container.
func (d *dosControl) boxSlots() []dosplay.SlotView {
	if d.box == nil {
		return nil
	}
	_, slots, ok := d.s.ContentSlots(d.side, d.box.at)
	if !ok {
		return nil
	}
	return slots
}

// boxUnit is the unit in the slot under the contents cursor.
func (d *dosControl) boxUnit() (dosplay.UnitView, bool) {
	slots := d.boxSlots()
	if d.box == nil || d.box.sel < 0 || d.box.sel >= len(slots) || slots[d.box.sel].Unit == nil {
		return dosplay.UnitView{}, false
	}
	return *slots[d.box.sel].Unit, true
}

// boxName names the open container: a building by its kind, a carrier by its
// name.
func (d *dosControl) boxName() string {
	if b, ok := d.own().Buildings[d.box.at]; ok {
		switch b.Kind {
		case dosplay.HQ:
			return "headquarters"
		case dosplay.Depot:
			return "depot"
		case dosplay.Factory:
		}
		return "factory"
	}
	if u, ok := d.own().Units[d.box.at]; ok {
		return u.Name
	}
	return "contents"
}

// inBuilding reports whether the open container is a building, which
// repairs.
func (d *dosControl) inBuilding() bool { return d.box.hold >= dosrules.HoldFlag40 }

// repairCost is the energy a repair of a unit of type typ costs.
func (d *dosControl) repairCost(typ int) int {
	p := d.s.Game.Profile
	if slices.Contains(p.Builders, typ) {
		return p.BuilderRepair
	}
	return p.Repair
}

// buildable are the types the open container builds.
func (d *dosControl) buildable() []dosplay.BuildOption { return d.s.Buildable(d.side, d.box.hold) }

// canProduce reports whether the open container builds anything and, if so,
// why it cannot build any of its types now (nil: it can).
func (d *dosControl) canProduce() (builds bool, err error) {
	opts := d.buildable()
	for _, o := range opts {
		if err = d.s.CanBuild(d.side, d.box.hold, o.Type); err == nil {
			return true, nil
		}
	}
	return len(opts) > 0, err
}

// canBox reports whether contents action a is possible on the slot under the
// cursor. Only cancel is while the orders are confirmed or carried out.
func (d *dosControl) canBox(a gesture.Action) bool {
	if d.over || d.busy() || d.s.Ready(d.side) {
		return a == gesture.Cancel
	}
	u, unit := d.boxUnit()
	switch a {
	case gesture.Cancel:
		return true
	case gesture.MoveOut:
		return unit && d.cannotLeave(d.box.hold, u) == ""
	case gesture.Repair:
		return unit && d.s.CanRepair(d.side, d.box.hold, u.Slot) == nil
	case gesture.Produce:
		builds, err := d.canProduce()
		return builds && err == nil
	default:
	}
	return false
}

// refreshBox computes the placement and the details of the open contents
// again when what they depend on changed.
func (d *dosControl) refreshBox() {
	b := d.box
	k := boxKey{version: d.s.Version(), energy: d.s.Energy(d.side), sel: b.sel, pick: b.pick, top: b.top, build: b.build,
		ready: d.s.Ready(d.side), over: d.over, busy: d.busy()}
	if b.cache.ok && b.cache.key == k {
		return
	}
	place := gesture.Placement{}
	details := d.buildDetails()
	if !b.build {
		place = gesture.ForContents().Restrict(d.canBox)
		details = d.slotDetails()
	}
	b.cache = boxCache{key: k, ok: true, place: place, details: details}
}

// contentsPlacement is the gesture placement on the contents cursor: the
// contents' gestures restricted to what is possible, cancel everywhere in
// the production list.
func (d *dosControl) contentsPlacement() gesture.Placement {
	if d.box == nil {
		return gesture.Placement{}
	}
	d.refreshBox()
	return d.box.cache.place
}

// boxDetails are the lines of the details plate.
func (d *dosControl) boxDetails() []string {
	d.refreshBox()
	return d.box.cache.details
}

// slotDetails are the details of the slot under the cursor: the unit's group
// size and whether it can move out and be repaired, and what the container
// builds, with the costs.
func (d *dosControl) slotDetails() []string {
	var out []string
	if u, ok := d.boxUnit(); ok {
		out = append(out, fmt.Sprintf("size %d/%d", u.Size, u.Full))
		if why := d.cannotLeave(d.box.hold, u); why != "" {
			out = append(out, "move out: "+why)
		} else {
			out = append(out, "move out")
		}
		if d.inBuilding() {
			cost := d.repairCost(u.Type)
			if err := d.s.CanRepair(d.side, d.box.hold, u.Slot); err != nil {
				out = append(out, fmt.Sprintf("repair %d: %s", cost, reason(err)))
			} else {
				out = append(out, fmt.Sprintf("repair: %d energy", cost))
			}
		}
	} else {
		out = append(out, "empty slot")
	}
	if builds, err := d.canProduce(); builds {
		if err != nil {
			out = append(out, "produce: "+reason(err))
		} else {
			out = append(out, "produce: "+count(len(d.buildable()), "type"))
		}
	}
	return out
}

// buildTop is the first row the production list shows of n with row pick
// picked when it showed from row top before: the rows move only as far as
// the pick needs, so a pick under the pointer never scrolls the list.
func buildTop(top, pick, n int) int {
	top = max(min(top, pick), pick-buildRows+1)
	return min(max(top, 0), max(n-buildRows, 0))
}

// setPick picks row i of the production list's n rows and scrolls it into
// view.
func (b *contents) setPick(i, n int) {
	b.pick = min(max(i, 0), n-1)
	b.top = buildTop(b.top, b.pick, n)
}

// buildDetails is the production list: the energy at hand, then the types
// and their prices around the picked one.
func (d *dosControl) buildDetails() []string {
	opts := d.buildable()
	out := []string{fmt.Sprintf("produce - energy %d", d.s.Energy(d.side))}
	first := d.box.top
	for _, o := range opts[first:min(first+buildRows, len(opts))] {
		out = append(out, fmt.Sprintf("%s  %d", o.Name, o.Cost))
	}
	return out
}

// boxStep moves the contents cursor one slot towards dir, or the production
// list's pick up or down.
func (d *dosControl) boxStep(dir gesture.Dir) {
	b := d.box
	if b.build {
		switch dir {
		case gesture.Up:
			b.setPick(b.pick-1, len(d.buildable()))
		case gesture.Down:
			b.setPick(b.pick+1, len(d.buildable()))
		default:
		}
		return
	}
	if j, ok := dospanel.SlotNeighbor(b.sel, panelDir(dir)); ok {
		b.sel = j
	}
}

// panelDir is the dospanel direction of a gesture direction.
func panelDir(dir gesture.Dir) dospanel.Dir {
	switch dir {
	case gesture.Up:
		return dospanel.Up
	case gesture.Left:
		return dospanel.Left
	case gesture.Right:
		return dospanel.Right
	case gesture.Down:
		return dospanel.Down
	default:
	}
	return dospanel.Centre
}

// runBox runs contents action a on the slot under the cursor.
func (d *dosControl) runBox(a gesture.Action) {
	switch a {
	case gesture.MoveOut:
		d.moveOut()
	case gesture.Produce:
		d.openBuild()
	case gesture.Repair:
		d.repair()
	case gesture.Cancel:
		d.box = nil
	default:
	}
}

// boxFire is fire alone on the contents: build the picked type in the
// production list, else move the unit under the cursor out.
func (d *dosControl) boxFire() {
	switch _, unit := d.boxUnit(); {
	case d.box.build:
		d.produce()
	case unit:
		d.moveOut()
	}
}

// moveOut closes the contents and aims the exit of the unit under the
// cursor, or says why it cannot leave.
func (d *dosControl) moveOut() {
	u, ok := d.boxUnit()
	if !ok {
		return
	}
	if why := d.cannotLeave(d.box.hold, u); why != "" {
		d.Message = u.Name + ": " + why
		return
	}
	d.aimExit(d.box.at, d.box.hold, u)
}

// openBuild opens the production list on the first type the side can pay.
func (d *dosControl) openBuild() {
	b := d.box
	opts := d.buildable()
	b.build, b.top = true, 0
	b.setPick(0, len(opts))
	for i, o := range opts {
		if d.s.CanBuild(d.side, b.hold, o.Type) == nil {
			b.setPick(i, len(opts))
			break
		}
	}
	d.Message = "produce: pick a type"
}

// produce builds the picked type; the cursor goes to the new unit's slot.
func (d *dosControl) produce() {
	b := d.box
	opts := d.buildable()
	if b.pick < 0 || b.pick >= len(opts) {
		return
	}
	o := opts[b.pick]
	evs, err := d.s.Build(d.side, b.hold, o.Type)
	if err != nil {
		d.Message = o.Name + ": " + reason(err)
		return
	}
	b.build = false
	for _, ev := range evs {
		if ev.Kind == dosplay.Built {
			b.sel = d.s.Unit(ev.Unit).Slot
			d.Message = fmt.Sprintf("%s built - %d energy left", o.Name, ev.Energy)
		}
	}
}

// repair repairs the unit under the cursor.
func (d *dosControl) repair() {
	u, ok := d.boxUnit()
	if !ok {
		return
	}
	if _, err := d.s.Repair(d.side, d.box.hold, u.Slot); err != nil {
		d.Message = u.Name + ": " + reason(err)
		return
	}
	d.Message = fmt.Sprintf("%s repaired - %d energy left", u.Name, d.s.Energy(d.side))
}

// syncContents shows the contents in the human side's panel while the
// session has a container open, and the map again once it closed.
func (g *game) syncContents() {
	s, d := g.split, g.dos
	m := &s.mode[d.side]
	switch {
	case d.box != nil && *m != panelContents:
		*m = panelContents
		s.track.Reset()
	case d.box == nil && *m == panelContents:
		*m = panelMap
	}
}

// contentsLayout is the contents' layout in the human side's panel.
func (g *game) contentsLayout() dospanel.ContentsRects {
	side := g.dos.side
	return dospanel.ContentsLayout(g.panelRect(side), g.plateScale(side, contentsFit))
}

// detailRow is the device rectangle of line i of the details plate.
func detailRow(l dospanel.ContentsRects, i int) image.Rectangle {
	k := l.Scale
	cy := float64(l.Details.Min.Y) + (6+7*float64(i))*k
	return image.Rect(l.Details.Min.X+int(2*k), int(cy-3.5*k), l.Details.Max.X-int(2*k), int(cy+3.5*k))
}

// buildRowAt is the production list's row under device pixel (x, y).
func (g *game) buildRowAt(x, y int) (int, bool) {
	l, n := g.contentsLayout(), len(g.dos.buildable())
	first := g.dos.box.top
	for i := range min(buildRows, n-first) {
		if image.Pt(x, y).In(detailRow(l, i+1)) {
			return first + i, true
		}
	}
	return 0, false
}

// contentsGesture handles the human side's input while the contents are
// open: the cursor follows the arrows and the mouse over the slots (or the
// list's rows), fire alone moves out (builds in the list), a gesture runs
// its action; the cancel direction, a right click or a press off the plate
// closes them, or goes back from the list.
func (g *game) contentsGesture(ev gesture.Event) {
	s, d := g.split, g.dos
	b := d.box
	if b == nil {
		return
	}
	at := func() (int, bool) {
		if b.build {
			return g.buildRowAt(int(ev.X), int(ev.Y))
		}
		return dospanel.SlotAt(g.contentsLayout(), int(ev.X), int(ev.Y))
	}
	set := func(i int) {
		if b.build {
			b.setPick(i, len(d.buildable()))
		} else {
			b.sel = i
		}
	}
	back := func() {
		if b.build {
			b.build = false
			return
		}
		d.box = nil
	}
	switch ev.Kind {
	case gesture.Hover:
		if i, ok := at(); ok {
			set(i)
		}
	case gesture.Step:
		d.boxStep(ev.Dir)
	case gesture.Press:
		s.dir = gesture.Centre
		if !ev.Mouse {
			break
		}
		if i, ok := at(); ok {
			set(i)
			s.track.Placement = d.contentsPlacement()
		} else if !image.Pt(int(ev.X), int(ev.Y)).In(g.contentsLayout().Plate) {
			s.track.Reset()
			back()
		}
	case gesture.Pick:
		s.dir = ev.Dir
	case gesture.Click:
		if _, ok := at(); ok || !ev.Mouse {
			d.boxFire()
		}
	case gesture.Run:
		d.runBox(ev.Action)
	case gesture.Cancelled:
		if a, _ := gesture.ForContents().At(s.dir); s.dir != gesture.Centre && a == gesture.Cancel || b.build {
			back()
		}
	case gesture.RingOpen:
		s.track.Reset()
		back()
	}
	g.syncContents()
}

// contentsView is what the contents plate shows: the container's name, the
// energy its building pays each round and the side's account, the slots and
// the details, cut to the details plate.
func (g *game) contentsView() dospanel.Contents {
	d := g.dos
	b := d.box
	c := dospanel.Contents{Name: strings.ToUpper(d.boxName()), Stored: d.s.Game.Buildings[b.hold].Credits,
		Total: d.s.Energy(d.side), Selected: b.sel}
	for i, sl := range d.boxSlots() {
		if u := sl.Unit; u != nil && i < dospanel.SlotCount {
			c.Slots[i] = dospanel.Slot{Occupied: true, Name: strings.ToUpper(u.Name), Size: u.Size, Full: u.Full,
				Sprite: g.panelUnit(*u, unitsprite.N), Portrait: g.portrait(u.ID)}
		}
	}
	l := g.contentsLayout()
	k := textScale(l.Scale)
	room := l.Details.Dx() - int(12*l.Scale)
	for _, line := range d.boxDetails() {
		c.Details = append(c.Details, pixfont.Small.Fit(strings.ToUpper(line), k, room))
	}
	return c
}

// drawContents draws the open contents in panel r: the plate, the picked row
// of the production list and, while fire is held, the gesture plate on the
// slot under the cursor.
func (g *game) drawContents(dst *ebiten.Image, r image.Rectangle) {
	d := g.dos
	if d.box == nil {
		return
	}
	l := g.contentsLayout()
	dospanel.DrawContents(dst, r, l.Scale, g.contentsView(), panelText{textScale(l.Scale)})
	if d.box.build {
		row := detailRow(l, d.box.pick-d.box.top+1)
		vector.StrokeRect(dst, float32(row.Min.X), float32(row.Min.Y), float32(row.Dx()), float32(row.Dy()),
			float32(max(l.Scale/2, 1)), dosscreen.NumberColour, false)
	}
}

// drawContentsPlate draws the gesture plate on the slot under the contents
// cursor while fire is held.
func (g *game) drawContentsPlate(dst *ebiten.Image) {
	s, d := g.split, g.dos
	a, ok, held := s.track.Plate()
	if d.box == nil || d.box.build || !held || d.box.sel < 0 || d.box.sel >= dospanel.SlotCount {
		return
	}
	r := g.contentsLayout().Slots[d.box.sel]
	act, st, _ := cursorPlate(gesture.Cancel, false, held, s.dir == gesture.Centre, a, ok)
	cx, cy := float64(r.Min.X+r.Max.X)/2, float64(r.Min.Y+r.Max.Y)/2
	gesture.DrawPlate(dst, cx, cy, float64(r.Dx())/hexgrid.HexWidth, act, st)
}
