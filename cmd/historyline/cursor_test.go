package main

import (
	"image"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/gesture"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/minimap"
)

// The tests below drive the cursor of -rules dos as the player does: frames of
// mouse and key set A through the gesture tracker into the session.

// newCursorTest is the split game of level1_cp against a passing side, with
// an own unit in the human panel that can move.
func newCursorTest(t *testing.T) (*game, *dosControl, hexgrid.Coord) {
	t.Helper()
	g, d := newSplitTest(t, passPlanner{}, 1800, 1000)
	at, ok := g.orderableInView()
	if !ok {
		t.Fatal("no own unit in view that can move")
	}
	return g, d, at
}

// Frame builders: the mouse at a screen point, held or not; a right click;
// keys held with the pointer elsewhere.
func mouseAt(x, y float64, down bool) gesture.Input {
	return gesture.Input{Pointer: gesture.Pointer{X: x, Y: y, Active: true, Down: down}}
}

func rightAt(x, y float64) gesture.Input {
	return gesture.Input{Pointer: gesture.Pointer{X: x, Y: y, Active: true, RightPressed: true}}
}

func held(k gesture.Keys) gesture.Input { return gesture.Input{Keys: k} }

// screenOf is the centre of c on the human panel in device pixels.
func screenOf(g *game, c hexgrid.Coord) (float64, float64) {
	return g.cam.WorldToScreen(hexgrid.Center(c))
}

// clickAt is a mouse click on c.
func clickAt(g *game, c hexgrid.Coord) []gesture.Input {
	x, y := screenOf(g, c)
	return []gesture.Input{mouseAt(x, y, false), mouseAt(x, y, true), mouseAt(x, y, false)}
}

// dragFrom presses on c, drags by (dx, dy) cells' thresholds and releases.
func dragFrom(g *game, c hexgrid.Coord, dx, dy float64) []gesture.Input {
	x, y := screenOf(g, c)
	th := g.split.track.Threshold * 1.2
	return []gesture.Input{mouseAt(x, y, false), mouseAt(x, y, true), mouseAt(x+dx*th, y+dy*th, true), mouseAt(x+dx*th, y+dy*th, false)}
}

// feed plays frames through the human side's tracker.
func feed(g *game, ins []gesture.Input) {
	for _, in := range ins {
		g.feedGesture(in, false)
	}
}

// ownHQ is the human side's HQ entrance.
func ownHQ(t *testing.T, d *dosControl) hexgrid.Coord {
	t.Helper()
	for c, b := range d.own().Buildings {
		if b.HQ && b.Owner == d.s.Faction(d.side) {
			return c
		}
	}
	t.Fatal("no own HQ")
	return hexgrid.Coord{}
}

// destinationInView is a destination of the aimed move that the human panel
// shows.
func destinationInView(t *testing.T, g *game, d *dosControl) hexgrid.Coord {
	t.Helper()
	for _, c := range dosplay.Coords(d.moves) {
		if viewShows(g.splitView(d.side), g.grid, c) {
			return c
		}
	}
	t.Fatal("no destination in view")
	return hexgrid.Coord{}
}

func TestDosPlacement(t *testing.T) {
	g, d, at := newCursorTest(t)
	empty := hexgrid.Coord{Line: -1}
	for _, c := range g.splitView(d.side).Cells() {
		if _, unit := d.own().Units[c]; !unit && g.grid.Contains(c) && !d.ownContainer(c) {
			empty = c
			break
		}
	}
	type want struct {
		dir gesture.Dir
		act gesture.Action
		ok  bool
	}
	tests := []struct {
		name  string
		setup func()
		at    hexgrid.Coord
		want  []want
	}{
		{"own unit", nil, at, []want{
			{gesture.Centre, gesture.Cancel, true}, {gesture.Up, gesture.Move, true}, {gesture.Left, gesture.ModeChange, true},
			{gesture.Right, gesture.Map, true}, {gesture.Down, gesture.Info, true},
		}},
		{"empty hex", nil, empty, []want{{gesture.Up, gesture.Move, false}, {gesture.Left, gesture.ModeChange, true}, {gesture.Right, gesture.Map, true}}},
		{"own HQ", nil, ownHQ(t, d), []want{{gesture.Left, gesture.Open, len(d.own().Buildings[ownHQ(t, d)].Units) > 0}, {gesture.Down, gesture.Info, true}}},
		{"confirmed", func() { d.EndTurn(); d.EndTurn() }, at, []want{{gesture.Up, gesture.Move, false}, {gesture.Left, gesture.ModeChange, true}}},
		{"aiming", func() { d.Click(at) }, at, []want{{gesture.Up, gesture.Cancel, true}, {gesture.Right, gesture.Cancel, true}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d.clear()
			if d.s.Ready(d.side) {
				d.s.Unconfirm(d.side)
			}
			if tt.setup != nil {
				tt.setup()
			}
			p := d.placement(tt.at)
			for _, w := range tt.want {
				if a, ok := p.At(w.dir); a != w.act || ok != w.ok {
					t.Errorf("%v: %v %v, want %v %v", w.dir, a, ok, w.act, w.ok)
				}
			}
		})
	}
	// The side in the action phase attacks above.
	other := newDosControl(d.s, 1, passPlanner{}, false)
	for c, u := range other.own().Units {
		if u.Side == 1 {
			if a, _ := other.placement(c).At(gesture.Up); a != gesture.Attack {
				t.Errorf("action phase: up is %v", a)
			}
			break
		}
	}
}

func TestDosGestures(t *testing.T) {
	tests := []struct {
		name  string
		input func(g *game, d *dosControl, at hexgrid.Coord) []gesture.Input
		check func(t *testing.T, g *game, d *dosControl, at hexgrid.Coord)
	}{
		{"hover moves the cursor", func(g *game, _ *dosControl, at hexgrid.Coord) []gesture.Input {
			x, y := screenOf(g, at)
			return []gesture.Input{mouseAt(0, 0, false), mouseAt(x, y, false)}
		}, func(t *testing.T, g *game, _ *dosControl, at hexgrid.Coord) {
			if g.hover != at || !g.hoverOnMap {
				t.Errorf("cursor %v %v, want %v", g.hover, g.hoverOnMap, at)
			}
		}},
		{"a click aims the unit's move", func(g *game, _ *dosControl, at hexgrid.Coord) []gesture.Input { return clickAt(g, at) }, func(t *testing.T, _ *game, d *dosControl, at hexgrid.Coord) {
			if !d.selected || d.from != at || d.picking() != gesture.Move || len(d.moves) == 0 {
				t.Errorf("selected %v from %v picking %v, %d moves", d.selected, d.from, d.picking(), len(d.moves))
			}
		}},
		{"dragging up aims the move", func(g *game, _ *dosControl, at hexgrid.Coord) []gesture.Input {
			return dragFrom(g, at, 0, -1)
		}, func(t *testing.T, _ *game, d *dosControl, at hexgrid.Coord) {
			if d.picking() != gesture.Move || d.from != at {
				t.Errorf("picking %v from %v", d.picking(), d.from)
			}
		}},
		{"dragging back cancels", func(g *game, _ *dosControl, at hexgrid.Coord) []gesture.Input {
			x, y := screenOf(g, at)
			th := g.split.track.Threshold * 1.2
			return []gesture.Input{mouseAt(x, y, false), mouseAt(x, y, true), mouseAt(x, y-th, true), mouseAt(x, y, true), mouseAt(x, y, false)}
		}, func(t *testing.T, _ *game, d *dosControl, _ hexgrid.Coord) {
			if d.selected {
				t.Error("a cancelled gesture selected the unit")
			}
		}},
		{"dragging left asks for the mode change", func(g *game, _ *dosControl, at hexgrid.Coord) []gesture.Input {
			return dragFrom(g, at, -1, 0)
		}, func(t *testing.T, _ *game, d *dosControl, _ hexgrid.Coord) {
			if !d.prompt || d.s.Ready(d.side) {
				t.Errorf("no prompt: %v, ready %v, %q", d.prompt, d.s.Ready(d.side), d.Message)
			}
		}},
		{"dragging right opens the strategic map", func(g *game, _ *dosControl, at hexgrid.Coord) []gesture.Input {
			return dragFrom(g, at, 1, 0)
		}, func(t *testing.T, g *game, d *dosControl, _ hexgrid.Coord) {
			if g.split.mode[d.side] != panelStrategic || g.split.mode[1-d.side] != panelMap {
				t.Errorf("panel modes %v", g.split.mode)
			}
		}},
		{"dragging down on the own HQ shows the game info", func(g *game, d *dosControl, _ hexgrid.Coord) []gesture.Input {
			return dragFrom(g, ownHQ(t, d), 0, 1)
		}, func(t *testing.T, g *game, d *dosControl, _ hexgrid.Coord) {
			if g.split.mode[d.side] != panelGameInfo {
				t.Errorf("panel mode %v", g.split.mode[d.side])
			}
		}},
		{"dragging down on a unit shows its type", func(g *game, _ *dosControl, at hexgrid.Coord) []gesture.Input {
			return dragFrom(g, at, 0, 1)
		}, func(t *testing.T, g *game, d *dosControl, at hexgrid.Coord) {
			if g.split.mode[d.side] != panelUnitInfo || g.split.infoUnit.Index != d.own().Units[at].Index {
				t.Errorf("panel mode %v, unit %+v", g.split.mode[d.side], g.split.infoUnit)
			}
		}},
		{"arrows step the cursor", func(g *game, _ *dosControl, at hexgrid.Coord) []gesture.Input {
			g.setCursor(at)
			return []gesture.Input{held(gesture.Keys{Right: true}), held(gesture.Keys{})}
		}, func(t *testing.T, g *game, _ *dosControl, at hexgrid.Coord) {
			if want := gesture.Neighbor(at, gesture.Right); g.hover != want {
				t.Errorf("cursor %v, want %v", g.hover, want)
			}
		}},
		{"space and up aim the move", func(g *game, _ *dosControl, at hexgrid.Coord) []gesture.Input {
			g.setCursor(at)
			return []gesture.Input{held(gesture.Keys{Fire: true}), held(gesture.Keys{Fire: true, Up: true}), held(gesture.Keys{})}
		}, func(t *testing.T, _ *game, d *dosControl, at hexgrid.Coord) {
			if d.picking() != gesture.Move || d.from != at {
				t.Errorf("picking %v from %v", d.picking(), d.from)
			}
		}},
		{"space alone clicks", func(g *game, _ *dosControl, at hexgrid.Coord) []gesture.Input {
			g.setCursor(at)
			return []gesture.Input{held(gesture.Keys{Fire: true}), held(gesture.Keys{})}
		}, func(t *testing.T, _ *game, d *dosControl, at hexgrid.Coord) {
			if !d.selected || d.from != at {
				t.Errorf("selected %v from %v", d.selected, d.from)
			}
		}},
		{"the ring runs the strategic map", func(g *game, _ *dosControl, at hexgrid.Coord) []gesture.Input {
			x, y := screenOf(g, at)
			feed(g, []gesture.Input{mouseAt(x, y, false), rightAt(x, y)})
			r, ok := g.split.track.Ring()
			i := slices.Index(r.Actions, gesture.Map)
			if !ok || i < 0 || !slices.Contains(r.Actions, gesture.Move) {
				t.Fatalf("ring %v %v", ok, r.Actions)
			}
			px, py := r.Pos(i)
			return []gesture.Input{mouseAt(px, py, false), mouseAt(px, py, true), mouseAt(px, py, false)}
		}, func(t *testing.T, g *game, d *dosControl, _ hexgrid.Coord) {
			if g.split.mode[d.side] != panelStrategic {
				t.Error("the ring's map did not open the strategic map")
			}
		}},
		{"a right click drops the aim", func(g *game, _ *dosControl, at hexgrid.Coord) []gesture.Input {
			x, y := screenOf(g, at)
			return append(clickAt(g, at), rightAt(x, y))
		}, func(t *testing.T, g *game, d *dosControl, _ hexgrid.Coord) {
			if _, open := g.split.track.Ring(); d.selected || open {
				t.Errorf("selected %v ring %v", d.selected, open)
			}
		}},
		{"a click on a lit hex moves", func(g *game, d *dosControl, at hexgrid.Coord) []gesture.Input {
			feed(g, clickAt(g, at))
			return clickAt(g, destinationInView(t, g, d))
		}, func(t *testing.T, _ *game, d *dosControl, at hexgrid.Coord) {
			if d.walk.Done() || d.selected {
				t.Errorf("no walk: %q", d.Message)
			}
		}},
		{"a drag while aiming drops the aim", func(g *game, _ *dosControl, at hexgrid.Coord) []gesture.Input {
			feed(g, clickAt(g, at))
			return dragFrom(g, at, 0, -1)
		}, func(t *testing.T, _ *game, d *dosControl, _ hexgrid.Coord) {
			if d.selected {
				t.Error("still aiming")
			}
		}},
		{"keys wait while a menu is open", func(g *game, d *dosControl, at hexgrid.Coord) []gesture.Input {
			g.setCursor(at)
			d.menu = []WeaponOption{{Label: "row"}}
			g.feedGesture(held(gesture.Keys{Fire: true}), true)
			d.menu = nil
			return []gesture.Input{held(gesture.Keys{Fire: true}), held(gesture.Keys{})}
		}, func(t *testing.T, _ *game, d *dosControl, _ hexgrid.Coord) {
			if d.selected {
				t.Error("the fire that chose a menu row also clicked the map")
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, d, at := newCursorTest(t)
			feed(g, tt.input(g, d, at))
			tt.check(t, g, d, at)
		})
	}
}

func TestStrategicMap(t *testing.T) {
	g, d, at := newCursorTest(t)
	feed(g, dragFrom(g, at, 1, 0))
	if g.split.mode[d.side] != panelStrategic {
		t.Fatal("the strategic map did not open")
	}
	// A press on a far cell of the strategic map looks there and closes it.
	far := hexgrid.Coord{Line: g.grid.Lines - 2, Col: g.grid.Width - 2}
	lay := g.strategicLayout(d.side)
	b := lay.Raster.Block(far).Add(lay.Map.Min)
	p, r, ui := g.split.screen.Panels[d.side], g.strategicRect(d.side), float64(g.view.UI)
	x, y := float64(r.X)+(float64(b.Min.X+b.Max.X)/2)*ui, float64(r.Y)+(float64(b.Min.Y+b.Max.Y)/2)*ui
	if c, ok := g.strategicPick(d.side, x, y); !ok || c != far {
		t.Fatalf("strategic pick %v %v, want %v", c, ok, far)
	}
	feed(g, []gesture.Input{mouseAt(x, y, false), mouseAt(x, y, true), mouseAt(x, y, false)})
	pair, row := centreView(g.grid, g.split.screen.Cols, g.split.screen.Rows, far)
	if c := g.split.cams[d.side]; g.split.mode[d.side] != panelMap || c.pair != pair || c.row != row || g.hover != far {
		t.Errorf("after the pick: mode %v camera %d,%d (want %d,%d) cursor %v", g.split.mode[d.side], c.pair, c.row, pair, row, g.hover)
	}
	if d.selected {
		t.Error("the release after the pick clicked the map")
	}
	// A press off the map, Esc and a right click close it without looking.
	for name, close := range map[string]func(){
		"press off the map": func() { feed(g, []gesture.Input{mouseAt(float64(p.X+1), float64(p.Y+1), true), mouseAt(0, 0, false)}) },
		"Esc":               func() { g.applyAction(actCancel) },
		"right click":       func() { feed(g, []gesture.Input{rightAt(x, y), mouseAt(x, y, false)}) },
	} {
		g.openStrategic(d.side)
		before := g.split.cams[d.side]
		close()
		if g.split.mode[d.side] != panelMap || g.split.cams[d.side].pair != before.pair || g.split.cams[d.side].row != before.row {
			t.Errorf("%s: mode %v camera %+v", name, g.split.mode[d.side], g.split.cams[d.side])
		}
	}
	// The scene shows the side's units and HQ in the side colours.
	sc := g.strategicScene(d.side)
	if sc.View == nil || sc.Terrain == nil {
		t.Fatal("scene without view or terrain")
	}
	if u := d.own().Units[at]; !slices.Contains(sc.Units, minimap.Unit{At: at, Shade: sideShade(u.Owner)}) || sideShade(u.Owner) != minimap.Blue {
		t.Errorf("own unit %v not a blue dot", at)
	}
	hq := ownHQ(t, d)
	if !slices.Contains(sc.Buildings, minimap.Building{At: hq, Kind: minimap.HQ, Shade: minimap.Blue}) {
		t.Errorf("own HQ %v missing from %v", hq, sc.Buildings)
	}
	if r := lay.ViewRect(sc.View); r.Empty() || !r.In(lay.Map) {
		t.Errorf("view rectangle %v in %v", r, lay.Map)
	}
}

func TestDosReachAndPlate(t *testing.T) {
	g, d, at := newCursorTest(t)
	if d.reach() != nil {
		t.Error("cells are dark without an aim")
	}
	feed(g, clickAt(g, at))
	r := d.reach()
	if !r[at] || len(r) != len(d.moves)+1 {
		t.Errorf("reach %d cells, %d moves", len(r), len(d.moves))
	}
	for c := range d.moves {
		if !r[c] {
			t.Errorf("destination %v dark", c)
		}
	}
	tests := []struct {
		name                       string
		pick                       gesture.Action
		onTarget, held, centre, ok bool
		act                        gesture.Action
		wantAct                    gesture.Action
		wantState                  gesture.State
		wantPlate                  bool
	}{
		{"idle", gesture.Cancel, false, false, true, true, gesture.Cancel, gesture.Cancel, gesture.Normal, false},
		{"held", gesture.Cancel, false, true, true, true, gesture.Cancel, gesture.Cancel, gesture.Normal, true},
		{"held on move", gesture.Cancel, false, true, false, true, gesture.Move, gesture.Move, gesture.Normal, true},
		{"held on a refused move", gesture.Cancel, false, true, false, false, gesture.Move, gesture.Move, gesture.Disabled, true},
		{"aiming over a target", gesture.Move, true, false, true, true, gesture.Cancel, gesture.Move, gesture.Normal, true},
		{"aiming elsewhere", gesture.Move, false, false, true, true, gesture.Cancel, gesture.Cancel, gesture.Normal, false},
		{"aim held on a target", gesture.Attack, true, true, true, true, gesture.Cancel, gesture.Attack, gesture.Pressed, true},
		{"aim held elsewhere", gesture.Attack, false, true, true, true, gesture.Cancel, gesture.Attack, gesture.Disabled, true},
		{"aim dragged off", gesture.Move, true, true, false, true, gesture.Cancel, gesture.Cancel, gesture.Normal, true},
	}
	for _, tt := range tests {
		a, st, plate := cursorPlate(tt.pick, tt.onTarget, tt.held, tt.centre, tt.act, tt.ok)
		if plate != tt.wantPlate || plate && (a != tt.wantAct || st != tt.wantState) {
			t.Errorf("%s: %v %v %v, want %v %v %v", tt.name, a, st, plate, tt.wantAct, tt.wantState, tt.wantPlate)
		}
	}
}

func TestKeepInView(t *testing.T) {
	g := hexgrid.Grid{Width: 20, Lines: 40}
	tests := []struct {
		pair, row int
		c         hexgrid.Coord
		wp, wr    int
	}{
		{3, 4, hexgrid.Coord{Line: 10, Col: 5}, 3, 4},    // inside: stays
		{3, 4, hexgrid.Coord{Line: 10, Col: 2}, 2, 4},    // one column pair left
		{3, 4, hexgrid.Coord{Line: 6, Col: 5}, 3, 3},     // one row up
		{3, 4, hexgrid.Coord{Line: 21, Col: 7}, 5, 5},    // right and down
		{0, 0, hexgrid.Coord{Line: 0, Col: 0}, 0, 0},     // the corner
		{3, 4, hexgrid.Coord{Line: 39, Col: 19}, 16, 14}, // clamped to the map
	}
	for _, tt := range tests {
		p, r := keepInView(g, 7, 6, tt.pair, tt.row, tt.c)
		v := tt.c
		if p != tt.wp || r != tt.wr {
			t.Errorf("view %d,%d with %v: %d,%d, want %d,%d", tt.pair, tt.row, v, p, r, tt.wp, tt.wr)
		}
	}
}

func TestHidePointer(t *testing.T) {
	g, d, at := newCursorTest(t)
	x, y := screenOf(g, at)
	other := g.split.screen.Panels[1-d.side]
	st := g.split.screen.Status[d.side]
	tests := []struct {
		name  string
		setup func()
		x, y  float64
		down  bool
		want  bool
	}{
		{"over a cell", nil, x, y, false, true},
		{"over the other panel", nil, float64(other.X + other.W/2), float64(other.Y + other.H/2), false, false},
		{"over the status bar", nil, float64(st.X + st.W/2), float64(st.Y + st.H/2), false, false},
		{"over the strategic map", func() { g.openStrategic(d.side) }, x, y, false, false},
		{"over a menu row", func() {
			d.menu = []WeaponOption{{Label: "row", Usable: true}}
			d.menuKind = menuContents
		}, 0, 0, false, false},
		{"after the battle", func() { d.over = true }, x, y, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g.split.mode, d.menu, d.over = [2]panelMode{}, nil, false
			if tt.setup != nil {
				tt.setup()
			}
			px, py := tt.x, tt.y
			if d.menu != nil {
				rx, ry, _, _ := g.menuRow(0)
				px, py = float64(rx*g.view.UI+2), float64(ry*g.view.UI+2)
			}
			if got := g.hidePointer(int(px), int(py), tt.down); got != tt.want {
				t.Errorf("hidden %v, want %v", got, tt.want)
			}
		})
	}
	classic := newMapGame(g.lvl, false)
	if classic.hidePointer(int(x), int(y), false) {
		t.Error("Classic hides the pointer")
	}
}

func TestSplitStrategicLayoutSize(t *testing.T) {
	g, d, _ := newCursorTest(t)
	p, r, ui := g.split.screen.Panels[d.side], g.strategicRect(d.side), g.view.UI
	// Inside the frame's teeth: half a cell from every edge of the panel.
	sc := g.split.screen.CellScale
	if r.X-p.X < int(hexgrid.HexWidth/2*sc) || r.Y-p.Y < int(hexgrid.HexHeight/2*sc) ||
		p.Right()-r.Right() < int(hexgrid.HexWidth/2*sc) || p.Bottom()-r.Bottom() < int(hexgrid.HexHeight/2*sc) || r.W <= 0 || r.H <= 0 {
		t.Errorf("strategic rect %+v in panel %+v", r, p)
	}
	if got, want := g.strategicLayout(d.side).Size, image.Pt(r.W/ui, r.H/ui); got != want {
		t.Errorf("strategic map size %v, want %v", got, want)
	}
}

func TestFitRing(t *testing.T) {
	p := dosscreen.Rect{X: 100, Y: 50, W: 800, H: 900}
	acts := []gesture.Action{gesture.Move, gesture.Map, gesture.Info, gesture.ModeChange}
	inside := func(r gesture.Ring) bool {
		for i := -1; i < len(r.Actions); i++ {
			x, y := r.X, r.Y
			if i >= 0 {
				x, y = r.Pos(i)
			}
			hw, hh := hexgrid.HexWidth/2*r.Scale, hexgrid.HexHeight/2*r.Scale
			if x-hw < float64(p.X)-1e-9 || y-hh < float64(p.Y)-1e-9 || x+hw > float64(p.Right())+1e-9 || y+hh > float64(p.Bottom())+1e-9 {
				return false
			}
		}
		return true
	}
	for _, at := range [][2]float64{{500, 500}, {110, 500}, {890, 60}, {500, 940}} {
		r := fitRing(gesture.NewRing(at[0], at[1], 1.5, acts), p)
		if !inside(r) {
			t.Errorf("ring at %v leaves the panel: %v,%v", at, r.X, r.Y)
		}
		if at == [2]float64{500, 500} && (r.X != 500 || r.Y != 500) {
			t.Errorf("a ring that fits moved to %v,%v", r.X, r.Y)
		}
	}
}

func TestEdgeDir(t *testing.T) {
	v := dosscreen.NewView(dosscreen.Rect{X: 0, Y: 0, W: 700, H: 700}, 1, 0, 0)
	bx, by := v.Offset()
	bw, bh := dosscreen.BlockSize(v.Cols, v.Rows)
	x0, y0, x1, y1 := float64(bx), float64(by), float64(bx)+bw, float64(by)+bh
	mx, my := (x0+x1)/2, (y0+y1)/2
	tests := []struct {
		name   string
		x, y   float64
		dp, dr int
	}{
		{"middle", mx, my, 0, 0},
		{"left edge cell", x0 + 10, my, -1, 0},
		{"inner half of the left cells", x0 + hexgrid.HexWidth/2 + 2, my, 0, 0},
		{"right edge cell", x1 - 10, my, 1, 0},
		{"top", mx, y0 + 5, 0, -1},
		{"bottom", mx, y1 - 5, 0, 1},
		{"past the block on the steel", x1 + 5, y1 + 5, 1, 1},
		{"top-left corner", x0 - 5, y0 - 5, -1, -1},
	}
	for _, tt := range tests {
		if dp, dr := edgeDir(v, tt.x, tt.y); dp != tt.dp || dr != tt.dr {
			t.Errorf("%s: %d,%d, want %d,%d", tt.name, dp, dr, tt.dp, tt.dr)
		}
	}
}

func TestEdgeScroll(t *testing.T) {
	g, d, _ := newCursorTest(t)
	s := g.split
	run := func(x, y, frames int, active bool) {
		for range frames {
			g.edgeScroll(x, y, active)
			g.updateSplit()
		}
	}
	v := g.splitView(d.side)
	bx, by := v.Offset()
	bw, bh := dosscreen.BlockSize(v.Cols, v.Rows)
	right := bx + int(bw*v.Scale) - 2
	midY := by + int(bh*v.Scale)/2
	start := s.cams[d.side]

	// Resting in the middle, or inactive on the edge, never scrolls.
	run(bx+int(bw*v.Scale)/2, midY, 100, true)
	run(right, midY, 100, false)
	if got := s.cams[d.side]; got.pair != start.pair || got.row != start.row {
		t.Fatalf("scrolled to %d,%d without an edge", got.pair, got.row)
	}
	// The other panel's side of the window does not scroll the human panel.
	other := s.screen.Panels[1-d.side]
	run(other.X+other.W/2, midY, 100, true)
	if got := s.cams[d.side]; got.pair != start.pair {
		t.Fatalf("the other panel scrolled the human panel to %d", got.pair)
	}

	// Not before edgeFirst frames, then step by step up to the map's edge.
	run(right, midY, edgeFirst, true)
	if got := s.cams[d.side].pair; got != start.pair {
		t.Fatalf("stepped after %d frames", edgeFirst)
	}
	_, hi := pairRange(g.grid, s.screen.Cols)
	run(right, midY, 2, true)
	if got := s.cams[d.side].pair; got != start.pair+1 {
		t.Fatalf("pair %d after the first step, want %d", got, start.pair+1)
	}
	if c, ok := g.pickSplit(right, midY); !ok || g.hover != c {
		t.Errorf("cursor %v after a step, under the pointer %v (%v)", g.hover, c, ok)
	}
	run(right, midY, 400, true)
	if got := s.cams[d.side].pair; got != hi {
		t.Fatalf("pair %d at the right edge, want the map's last %d", got, hi)
	}
	// The cursor is the cell under the pointer (when it is a map cell; the
	// odd lines end a cell early at the right edge).
	if c, ok := g.pickSplit(right, midY); ok && (!g.hoverOnMap || g.hover != c) {
		t.Errorf("cursor %v (%v), under the pointer %v (%v)", g.hover, g.hoverOnMap, c, ok)
	}
	// Up and left back to the corner.
	run(bx+2, by+2, 800, true)
	if got := s.cams[d.side]; got.pair != 0 || got.row != 0 {
		t.Errorf("corner at %d,%d, want 0,0", got.pair, got.row)
	}
}
