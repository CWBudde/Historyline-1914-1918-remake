package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dospanel"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/gesture"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// The tests below open a container's contents in the human side's panel on
// level2_cp (Civil), whose sides start with factories and depots holding
// units, and drive the slot cursor as the player does.

// newContentsTest is the split game of Civil against a passing side with the
// human panel looking at its factory, and the factory's entrance.
func newContentsTest(t *testing.T) (*game, *dosControl, hexgrid.Coord) {
	t.Helper()
	d := newDosTest(t, "level2_cp", passPlanner{})
	g := newMapGame(d.s.Level, false)
	g.dos = d
	g.applyLayout(1920, 1080, 1)
	g.layoutSplit()
	d.SetReducedMotion(true) // no slides: the cells stay where the view rests
	f := ownKind(t, d, dosplay.Factory)
	g.lookAt(d.side, f)
	g.syncSplitCam()
	return g, d, f
}

// ownKind is the human side's first building of kind k.
func ownKind(t *testing.T, d *dosControl, k dosplay.BuildingKind) hexgrid.Coord {
	t.Helper()
	bs := d.own().Buildings
	for _, c := range dosplay.Coords(toSet(bs)) {
		if b := bs[c]; b.Kind == k && b.Owner == d.s.Faction(d.side) {
			return c
		}
	}
	t.Fatalf("no own %v", k)
	return hexgrid.Coord{}
}

// fireDir holds fire, pushes dir and releases.
func fireDir(dir gesture.Dir) []gesture.Input {
	k := gesture.Keys{Fire: true}
	switch dir {
	case gesture.Up:
		k.Up = true
	case gesture.Left:
		k.Left = true
	case gesture.Right:
		k.Right = true
	case gesture.Down:
		k.Down = true
	default:
	}
	return []gesture.Input{held(gesture.Keys{Fire: true}), held(k), held(gesture.Keys{})}
}

// fire is a press and release of fire alone.
func fire() []gesture.Input {
	return []gesture.Input{held(gesture.Keys{Fire: true}), held(gesture.Keys{})}
}

// step is an arrow pressed and released.
func step(k gesture.Keys) []gesture.Input { return []gesture.Input{held(k), held(gesture.Keys{})} }

// openFactory opens the factory's contents with the open gesture.
func openFactory(t *testing.T, g *game, d *dosControl, f hexgrid.Coord) {
	t.Helper()
	feed(g, dragFrom(g, f, -1, 0))
	if g.split.mode[d.side] != panelContents || d.box == nil || d.box.at != f {
		t.Fatalf("contents did not open: mode %v box %+v (%q)", g.split.mode[d.side], d.box, d.Message)
	}
}

func TestContentsCursorOpen(t *testing.T) {
	g, d, f := newContentsTest(t)
	openFactory(t, g, d, f)
	slots := d.boxSlots()
	if len(slots) != dospanel.SlotCount || slots[d.box.sel].Unit == nil {
		t.Fatalf("%d slots, the cursor on slot %d (%+v)", len(slots), d.box.sel, slots[d.box.sel])
	}
	c := g.contentsView()
	if c.Name != "FACTORY" || c.Total != d.s.Energy(d.side) || c.Stored != d.s.Game.Buildings[d.box.hold].Credits || c.Selected != d.box.sel {
		t.Errorf("contents %q stored %d total %d selected %d", c.Name, c.Stored, c.Total, c.Selected)
	}
	n := 0
	for i, s := range c.Slots {
		if s.Occupied != (slots[i].Unit != nil) {
			t.Errorf("slot %d occupied %v", i, s.Occupied)
		}
		if s.Occupied {
			n++
		}
	}
	if n == 0 || len(c.Details) == 0 || !strings.Contains(c.Details[0], "SIZE") {
		t.Errorf("%d units, details %q", n, c.Details)
	}
	// An own factory without units opens too, so it can produce; an empty HQ
	// does not.
	if !d.can(gesture.Open, f) {
		t.Error("the factory does not open")
	}
	if hq := ownHQ(t, d); len(d.own().Buildings[hq].Units) == 0 && d.can(gesture.Open, hq) {
		t.Error("an empty HQ opens")
	}
	// The map's HUD menu stays closed.
	if d.MenuOpen() {
		t.Error("the contents opened the HUD menu")
	}
}

func TestContentsCursorMoves(t *testing.T) {
	g, d, f := newContentsTest(t)
	openFactory(t, g, d, f)
	d.box.sel = 2
	feed(g, step(gesture.Keys{Right: true}))
	if d.box.sel != 3 {
		t.Errorf("right: slot %d, want 3", d.box.sel)
	}
	feed(g, step(gesture.Keys{Down: true}))
	if d.box.sel != 2 {
		t.Errorf("down from the upper row: slot %d, want 2", d.box.sel)
	}
	feed(g, step(gesture.Keys{Up: true}))
	if d.box.sel != 3 {
		t.Errorf("up: slot %d, want 3", d.box.sel)
	}
	d.box.sel = 0
	feed(g, step(gesture.Keys{Left: true}))
	feed(g, step(gesture.Keys{Down: true}))
	if d.box.sel != 0 {
		t.Errorf("past the edge: slot %d", d.box.sel)
	}
	// The mouse hovers and presses slots.
	l := g.contentsLayout()
	r := l.Slots[5]
	x, y := float64(r.Min.X+r.Max.X)/2, float64(r.Min.Y+r.Max.Y)/2
	feed(g, []gesture.Input{mouseAt(1, 1, false), mouseAt(x, y, false)})
	if d.box.sel != 5 {
		t.Errorf("hover: slot %d, want 5", d.box.sel)
	}
	r = l.Slots[6]
	x, y = float64(r.Min.X+r.Max.X)/2, float64(r.Min.Y+r.Max.Y)/2
	feed(g, []gesture.Input{mouseAt(x, y, true)})
	if d.box.sel != 6 {
		t.Errorf("press: slot %d, want 6", d.box.sel)
	}
	feed(g, []gesture.Input{mouseAt(x, y, false)})
	if g.split.mode[d.side] != panelContents {
		t.Errorf("a click on an empty slot closed the contents: %v", g.split.mode[d.side])
	}
	// The cancel direction closes.
	feed(g, fireDir(gesture.Right))
	if g.split.mode[d.side] != panelMap || d.box != nil {
		t.Errorf("cancel: mode %v box %+v", g.split.mode[d.side], d.box)
	}
}

func TestContentsCursorRestrict(t *testing.T) {
	g, d, f := newContentsTest(t)
	openFactory(t, g, d, f)
	unit := d.box.sel
	empty := slices.IndexFunc(d.boxSlots(), func(s dosplay.SlotView) bool { return s.Unit == nil })
	if empty < 0 {
		t.Fatal("no empty slot")
	}
	type want struct{ moveOut, produce, repair bool }
	check := func(name string, slot int, w want) {
		t.Helper()
		d.box.sel = slot
		p := d.contentsPlacement()
		for _, c := range []struct {
			dir gesture.Dir
			act gesture.Action
			ok  bool
		}{{gesture.Up, gesture.MoveOut, w.moveOut}, {gesture.Left, gesture.Produce, w.produce}, {gesture.Down, gesture.Repair, w.repair}, {gesture.Right, gesture.Cancel, true}} {
			if a, ok := p.At(c.dir); a != c.act || ok != c.ok {
				t.Errorf("%s: %v is %v %v, want %v %v", name, c.dir, a, ok, c.act, c.ok)
			}
		}
	}
	u := d.boxSlots()[unit].Unit
	leaves := d.cannotLeave(d.box.hold, *u) == ""
	check("no energy, a unit", unit, want{moveOut: leaves})
	check("no energy, empty", empty, want{})
	d.s.Game.Energy[d.side] = 500
	check("energy, a unit", unit, want{moveOut: leaves, produce: true, repair: true})
	check("energy, empty", empty, want{produce: true})
	// The details name the actions with their costs.
	d.box.sel = unit
	details := strings.Join(g.contentsView().Details, "\n")
	for _, w := range []string{"REPAIR", "15", "PRODUCE"} {
		if !strings.Contains(details, w) {
			t.Errorf("details lack %q: %q", w, details)
		}
	}
	// In the action phase nothing leaves, is built or repaired.
	finish(t, d)
	g.split.mode[d.side] = panelMap
	d.openContents(f)
	check("action phase", unit, want{produce: false, repair: false})
}

func TestContentsCursorProduce(t *testing.T) {
	g, d, f := newContentsTest(t)
	openFactory(t, g, d, f)
	d.s.Game.Energy[d.side] = 500
	_, units, _ := d.s.Contents(d.side, f)
	feed(g, fireDir(gesture.Left))
	if d.box == nil || !d.box.build {
		t.Fatalf("no production list: %+v (%q)", d.box, d.Message)
	}
	opts := d.s.Buildable(d.side, d.box.hold)
	details := strings.Join(g.contentsView().Details, "\n")
	if !strings.Contains(details, strings.ToUpper(opts[0].Name)) {
		t.Errorf("the list lacks %s: %q", opts[0].Name, details)
	}
	// Cancel goes back to the slots, the left gesture opens the list again.
	g.applyAction(actCancel)
	if d.box == nil || d.box.build || g.split.mode[d.side] != panelContents {
		t.Fatalf("Esc in the list: %+v mode %v", d.box, g.split.mode[d.side])
	}
	feed(g, fireDir(gesture.Left))
	feed(g, step(gesture.Keys{Down: true}))
	if d.box.pick != 1 {
		t.Errorf("down picks row %d", d.box.pick)
	}
	feed(g, fire())
	_, after, _ := d.s.Contents(d.side, f)
	if len(after) != len(units)+1 || d.s.Energy(d.side) != 500-opts[1].Cost {
		t.Fatalf("build: %d units (was %d), energy %d (%q)", len(after), len(units), d.s.Energy(d.side), d.Message)
	}
	if d.box == nil || d.box.build {
		t.Errorf("after the build: %+v", d.box)
	}
	if s := d.boxSlots()[d.box.sel]; s.Unit == nil || s.Unit.Type != opts[1].Type {
		t.Errorf("the cursor is not on the new unit: %+v", s)
	}
	if !strings.Contains(strings.ToLower(d.Message), "built") {
		t.Errorf("message %q", d.Message)
	}
}

func TestContentsCursorRepair(t *testing.T) {
	g, d, f := newContentsTest(t)
	openFactory(t, g, d, f)
	d.s.Game.Energy[d.side] = 100
	idx := d.boxSlots()[d.box.sel].Unit.Index
	feed(g, fireDir(gesture.Down))
	if d.s.Energy(d.side) != 85 || !d.s.Unit(idx).Moved {
		t.Errorf("repair: energy %d, moved %v (%q)", d.s.Energy(d.side), d.s.Unit(idx).Moved, d.Message)
	}
	if !strings.Contains(strings.ToLower(d.Message), "repaired") {
		t.Errorf("message %q", d.Message)
	}
	// Once repaired the unit has its order: no second repair.
	if _, ok := d.contentsPlacement().At(gesture.Down); ok {
		t.Error("a second repair is possible")
	}
	if g.split.mode[d.side] != panelContents {
		t.Error("the repair closed the contents")
	}
}

func TestContentsCursorMoveOut(t *testing.T) {
	g, d, f := newContentsTest(t)
	openFactory(t, g, d, f)
	slot := -1
	for _, s := range d.boxSlots() {
		if s.Unit != nil && d.cannotLeave(d.box.hold, *s.Unit) == "" {
			slot = s.Slot
			break
		}
	}
	if slot < 0 {
		t.Fatal("no unit can leave the factory")
	}
	d.box.sel = slot
	idx := d.boxSlots()[slot].Unit.Index
	feed(g, fireDir(gesture.Up))
	if g.split.mode[d.side] != panelMap || d.box != nil || !d.selected || d.hold < 0 || d.unit != idx || len(d.moves) == 0 {
		t.Fatalf("move out: mode %v selected %v hold %d unit %d, %d moves (%q)", g.split.mode[d.side], d.selected, d.hold, d.unit, len(d.moves), d.Message)
	}
	to := dosplay.Coords(d.moves)[0]
	d.Click(to)
	if v := d.own().Units[to]; v.Index != idx || v.Held {
		t.Errorf("after the exit %v shows %+v (%q)", to, v, d.Message)
	}
}

func TestContentsCursorDraw(t *testing.T) {
	g, d, f := newContentsTest(t)
	openFactory(t, g, d, f)
	img := ebiten.NewImage(1920, 1080)
	g.drawSplit(img)
	feed(g, []gesture.Input{held(gesture.Keys{Fire: true}), held(gesture.Keys{Fire: true, Down: true})})
	g.drawSplit(img) // the plate on the slot
	feed(g, []gesture.Input{held(gesture.Keys{})})
	d.s.Game.Energy[d.side] = 500
	feed(g, fireDir(gesture.Left))
	g.drawSplit(img)
	g.drawWells(img)
}
