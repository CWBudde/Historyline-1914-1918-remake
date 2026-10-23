package main

import (
	"image"
	"image/color"
	"maps"
	"math"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/minimap"
)

func TestViewRanges(t *testing.T) {
	for _, tc := range []struct {
		name         string
		g            hexgrid.Grid
		cols, rows   int
		plo, phi     int
		rlo, rhi     int
		wantCentered bool
	}{
		// 15 columns (8 even, 7 odd), 12 rows of even lines.
		{"PULSE-like, 7 columns", hexgrid.Grid{Width: 8, Lines: 24}, 7, 7, 0, 4, 0, 5, false},
		{"8 columns may end beyond the map", hexgrid.Grid{Width: 8, Lines: 24}, 8, 7, 0, 4, 0, 5, false},
		{"exact fit", hexgrid.Grid{Width: 4, Lines: 14}, 7, 7, 0, 0, 0, 0, false},
		{"odd line count", hexgrid.Grid{Width: 8, Lines: 23}, 7, 7, 0, 4, 0, 5, false},
		{"map narrower and shorter than the block", hexgrid.Grid{Width: 3, Lines: 6}, 13, 7, -2, -2, -2, -2, true},
	} {
		plo, phi := pairRange(tc.g, tc.cols)
		rlo, rhi := rowRange(tc.g, tc.rows)
		if plo != tc.plo || phi != tc.phi || rlo != tc.rlo || rhi != tc.rhi {
			t.Errorf("%s: pairs %d..%d rows %d..%d, want %d..%d and %d..%d", tc.name, plo, phi, rlo, rhi, tc.plo, tc.phi, tc.rlo, tc.rhi)
		}
		if tc.wantCentered {
			continue
		}
		// At the far end the last (even) column and row are on the map, one further
		// they are not.
		even := (tc.cols - 1) &^ 1
		last := dosscreen.View{Cols: tc.cols, Rows: tc.rows, Pair: phi, Row: rhi}
		if c := last.Cell(even, 0); !tc.g.Contains(c) || c.Col != tc.g.Width-1 {
			t.Errorf("%s: last column %v off the map", tc.name, c)
		}
		if c := last.Cell(0, tc.rows-1); !tc.g.Contains(c) {
			t.Errorf("%s: last row %v off the map", tc.name, c)
		}
		beyond := dosscreen.View{Cols: tc.cols, Rows: tc.rows, Pair: phi + 1, Row: rhi + 1}
		if tc.g.Contains(beyond.Cell(even, 0)) || tc.g.Contains(beyond.Cell(0, tc.rows-1)) {
			t.Errorf("%s: the range ends early", tc.name)
		}
	}
}

func TestClampAndCentreView(t *testing.T) {
	g := hexgrid.Grid{Width: 8, Lines: 24}
	for _, tc := range []struct{ pair, row, wp, wr int }{
		{-3, -9, 0, 0}, {2, 3, 2, 3}, {9, 99, 4, 5},
	} {
		if p, r := clampView(g, 7, 7, tc.pair, tc.row); p != tc.wp || r != tc.wr {
			t.Errorf("clamp %d,%d = %d,%d, want %d,%d", tc.pair, tc.row, p, r, tc.wp, tc.wr)
		}
	}
	for _, size := range [][2]int{{7, 7}, {8, 6}, {9, 7}, {12, 9}} {
		cols, rows := size[0], size[1]
		for l := range g.Lines {
			for c := range g.Width - l&1 {
				at := hexgrid.Coord{Line: l, Col: c}
				pair, row := centreView(g, cols, rows, at)
				v := dosscreen.View{Cols: cols, Rows: rows, Pair: pair, Row: row}
				if !viewShows(v, g, at) {
					t.Fatalf("%dx%d centred on %v at %d,%d does not show it", cols, rows, at, pair, row)
				}
				if p, r := clampView(g, cols, rows, pair, row); p != pair || r != row {
					t.Fatalf("%dx%d centred on %v leaves the map: %d,%d", cols, rows, at, pair, row)
				}
				// Away from the edges the hex sits in the middle column (pair) and row.
				k := 2*at.Col + at.Line&1
				plo, phi := pairRange(g, cols)
				rlo, rhi := rowRange(g, rows)
				if j := k - 2*pair; pair > plo && pair < phi && (j < (cols-1)/2-1 || j > (cols-1)/2+1) {
					t.Errorf("%dx%d: %v in column %d of %d", cols, rows, at, j, cols)
				}
				if i := l>>1 - row; row > rlo && row < rhi && i != (rows-1)/2 {
					t.Errorf("%dx%d: %v in row %d of %d", cols, rows, at, i, rows)
				}
			}
		}
	}
}

func TestViewShows(t *testing.T) {
	g := hexgrid.Grid{Width: 8, Lines: 23}
	v := dosscreen.View{Cols: 7, Rows: 7, Pair: 1, Row: 2}
	for _, tc := range []struct {
		at   hexgrid.Coord
		want bool
	}{
		{hexgrid.Coord{Line: 4, Col: 1}, true},   // first column, first cell
		{hexgrid.Coord{Line: 5, Col: 1}, true},   // second column (odd lines)
		{hexgrid.Coord{Line: 4, Col: 4}, true},   // seventh column
		{hexgrid.Coord{Line: 5, Col: 4}, false},  // eighth column
		{hexgrid.Coord{Line: 3, Col: 2}, false},  // above
		{hexgrid.Coord{Line: 17, Col: 2}, true},  // last row, odd column
		{hexgrid.Coord{Line: 18, Col: 2}, false}, // below
		{hexgrid.Coord{Line: 4, Col: 0}, false},  // left
	} {
		if got := viewShows(v, g, tc.at); got != tc.want {
			t.Errorf("%v: %v, want %v", tc.at, got, tc.want)
		}
	}
	// Off the map: the odd line beyond the last is steel.
	if viewShows(dosscreen.View{Cols: 7, Rows: 7, Pair: 0, Row: 5}, g, hexgrid.Coord{Line: 23, Col: 0}) {
		t.Error("an off-map cell counts as shown")
	}
}

func TestPanelCamSlide(t *testing.T) {
	var c panelCam
	c.moveTo(2, 3, false)
	if c.left != 0 || c.slideX != 0 || c.pair != 2 || c.row != 3 {
		t.Fatalf("the first placement slid: %+v", c)
	}
	c.moveTo(3, 2, false)
	if c.left != slideFrames || c.slideX != hexgrid.ColPitch || c.slideY != -hexgrid.HexHeight {
		t.Fatalf("slide %+v", c)
	}
	// The drawn origin (rest minus slide) starts where the camera was and moves
	// steadily to the new rest.
	last := math.Inf(1)
	for range slideFrames {
		c.tick()
		if c.slideX >= last || c.slideX < 0 {
			t.Fatalf("slide x %v after %v", c.slideX, last)
		}
		last = c.slideX
	}
	if c.left != 0 || c.slideX != 0 || c.slideY != 0 {
		t.Errorf("slide not done: %+v", c)
	}
	// A move during a slide goes on from where the panel is drawn.
	c.moveTo(4, 2, false)
	c.tick()
	c.moveTo(5, 2, false)
	if want := hexgrid.ColPitch*7.0/8 + hexgrid.ColPitch; math.Abs(c.slideX-want) > 1e-9 {
		t.Errorf("chained slide %v, want %v", c.slideX, want)
	}
	c.moveTo(1, 1, true)
	if c.left != 0 || c.slideX != 0 || c.slideY != 0 || c.pair != 1 || c.row != 1 {
		t.Errorf("reduced motion slid: %+v", c)
	}
}

func TestPanelCameraMatchesView(t *testing.T) {
	s := dosscreen.Layout(1920, 1080, 1)
	world := camera.Size{W: 1200, H: 2000}
	screen := camera.Size{W: 1920, H: 1080}
	for side := range 2 {
		v := s.View(side, 1, 2)
		cam := panelCamera(v, 0, 0, world, screen, 2)
		if math.Abs(cam.Effective()-v.Scale) > 1e-12 || cam.Scale != 2 {
			t.Fatalf("effective %v scale %v, view scale %v", cam.Effective(), cam.Scale, v.Scale)
		}
		for _, p := range [][2]float64{{0, 0}, {142, 80}, {537.5, 913}} {
			gx, gy := cam.WorldToScreen(p[0], p[1])
			wx, wy := v.WorldToScreen(p[0], p[1])
			if math.Abs(gx-wx) > 1e-9 || math.Abs(gy-wy) > 1e-9 {
				t.Errorf("side %d: %v at %v,%v, view %v,%v", side, p, gx, gy, wx, wy)
			}
		}
		slid := panelCamera(v, 142, -80, world, screen, 2)
		gx, gy := slid.WorldToScreen(0, 0)
		wx, wy := v.WorldToScreen(0, 0)
		if math.Abs(gx-(wx+142*v.Scale)) > 1e-9 || math.Abs(gy-(wy-80*v.Scale)) > 1e-9 {
			t.Errorf("side %d: slid origin at %v,%v, rest %v,%v", side, gx, gy, wx, wy)
		}
	}
}

// newSplitTest is the game of a dos session against planner in a window of
// w × h.
func newSplitTest(t *testing.T, planner dosPlanner, w, h int) (*game, *dosControl) {
	t.Helper()
	d := newDosTest(t, "level1_cp", planner)
	g := newMapGame(d.s.Level, false)
	g.dos = d
	g.applyLayout(w, h, 1)
	g.layoutSplit()
	return g, d
}

func TestSplitLayoutAndPicking(t *testing.T) {
	g, d := newSplitTest(t, passPlanner{}, 1800, 1000)
	s := g.split
	if s == nil || s.screen.Cols < dosscreen.MinCols || s.screen.Rows < dosscreen.MinRows {
		t.Fatalf("split layout %+v", s)
	}
	// Each panel opens on its side's HQ.
	for side := range 2 {
		hq := hexgrid.Coord{Line: -1}
		for c, b := range d.buildingsOf(side) {
			if b.HQ && b.Owner == d.s.Faction(side) {
				hq = c
			}
		}
		if !viewShows(g.splitView(side), g.grid, hq) {
			t.Errorf("panel %d does not show its HQ %v", side, hq)
		}
	}
	// g.cam is the human panel's camera and rests on its view.
	v := g.splitView(d.side)
	ox, oy := v.Origin()
	bx, by := v.Offset()
	if sx, sy := g.cam.WorldToScreen(ox, oy); sx != float64(bx) || sy != float64(by) {
		t.Errorf("block corner at %v,%v, view offset %d,%d", sx, sy, bx, by)
	}
	for _, c := range v.Cells() {
		if !g.grid.Contains(c) {
			continue
		}
		sx, sy := v.WorldToScreen(hexgrid.Center(c))
		if got, ok := g.pickSplit(int(sx), int(sy)); !ok || got != c {
			t.Fatalf("pick at the centre of %v: %v %v", c, got, ok)
		}
		// The same spot on the other panel is not the human side's.
		other := g.split.screen.Panels[1-d.side]
		if _, ok := g.pickSplit(int(sx)-v.Panel.X+other.X, int(sy)); ok {
			t.Fatalf("the other panel takes a click at %v", c)
		}
	}
	// Steel around the block and the status bar pick nothing.
	if _, ok := g.pickSplit(v.Panel.X, v.Panel.Y); ok {
		t.Error("the panel corner picks a hex")
	}
	st := g.split.screen.Status[d.side]
	if _, ok := g.pickSplit(st.X+st.W/2, st.Y+st.H/2); ok {
		t.Error("the status bar picks a hex")
	}
	// A new window keeps the cameras on the map.
	g.applyLayout(1024, 768, 1)
	for side := range 2 {
		c := g.split.cams[side]
		if p, r := clampView(g.grid, g.split.screen.Cols, g.split.screen.Rows, c.pair, c.row); p != c.pair || r != c.row || c.left != 0 {
			t.Errorf("panel %d after a resize: %+v", side, c)
		}
	}
}

func TestSplitStepView(t *testing.T) {
	g, d := newSplitTest(t, passPlanner{}, 1800, 1000)
	c := &g.split.cams[d.side]
	c.moveTo(0, 0, true)
	g.stepView(d.side, 1, 1)
	if c.pair != 1 || c.row != 1 || c.left != slideFrames {
		t.Errorf("step: %+v", *c)
	}
	g.stepView(d.side, -5, -5)
	if c.pair != 0 || c.row != 0 {
		t.Errorf("step off the map: %+v", *c)
	}
	g.applyAction(actViewBottom)
	if _, hi := rowRange(g.grid, g.split.screen.Rows); c.row != hi {
		t.Errorf("PageDown: row %d, want %d", c.row, hi)
	}
}

func TestPlanFocus(t *testing.T) {
	a, b, c := hexgrid.Coord{Line: 1, Col: 1}, hexgrid.Coord{Line: 2, Col: 2}, hexgrid.Coord{Line: 3, Col: 3}
	for _, tc := range []struct {
		name string
		evs  []dosplay.Event
		want hexgrid.Coord
		ok   bool
	}{
		{"nothing", []dosplay.Event{{Kind: dosplay.Confirmed}}, hexgrid.Coord{}, false},
		{"last move", []dosplay.Event{{Kind: dosplay.Moved, Path: []hexgrid.Coord{a, b}}, {Kind: dosplay.Moved, Path: []hexgrid.Coord{b, c}}, {Kind: dosplay.Confirmed}}, c, true},
		{"an attack order", []dosplay.Event{{Kind: dosplay.Ordered, From: a, To: b}, {Kind: dosplay.Ordered, From: c, To: b}}, c, true},
	} {
		if got, ok := planFocus(tc.evs); got != tc.want || ok != tc.ok {
			t.Errorf("%s: %v %v, want %v %v", tc.name, got, ok, tc.want, tc.ok)
		}
	}
}

func TestSplitFollowsTheComputer(t *testing.T) {
	g, d := newSplitTest(t, approachPlanner{}, 1800, 1000)
	d.Tick()
	finish(t, d) // the computer is in its movement phase now and has moved
	moved := d.follow[1]
	if moved == nil {
		t.Fatal("the computer's moves ask its panel to look nowhere")
	}
	want := *moved
	g.split.cams[1].moveTo(0, 0, true)
	g.followSplit()
	if !viewShows(g.splitView(1), g.grid, want) || d.follow[1] != nil {
		t.Errorf("the computer's panel does not show %v", want)
	}
	if want != g.startHex(1) && g.split.cams[1].left == 0 && (g.split.cams[1].pair != 0 || g.split.cams[1].row != 0) {
		t.Error("the computer's panel jumped instead of sliding")
	}
	// The resolution freezes the computer's view; its moves are not walked
	// again on the human panel.
	confirmOrders(t, d)
	d.Tick()
	if d.otherFrozen == nil || !maps.Equal(d.unitsOf(1, false), d.otherFrozen.Units) {
		t.Fatal("the computer's view is not frozen during the resolution")
	}
	for range 10000 {
		if !d.busy() {
			break
		}
		if !d.walk.Done() {
			t.Fatalf("the computer's move %v walks again on the human panel", d.walk.path)
		}
		d.DismissReport()
		d.Tick()
	}
	if d.otherFrozen != nil || d.frozen != nil {
		t.Error("a view stays frozen")
	}
}

func TestSplitOtherPanelUnits(t *testing.T) {
	g, d := newSplitTest(t, passPlanner{}, 1800, 1000)
	if !maps.Equal(d.unitsOf(1, false), d.s.View(1).Units) || !maps.Equal(toSet(d.buildingsOf(1)), toSet(d.s.View(1).Buildings)) {
		t.Error("the other panel does not show the other side's view")
	}
	// Drawing the other panel draws its units; outside the panel pass the human
	// side's map is the one shown.
	g.drawing, g.drawSide = true, 1
	for c := range d.s.View(1).Units {
		if shown, _ := g.unitShown(c); !shown {
			t.Errorf("unit on %v not drawn on the other panel", c)
		}
	}
	g.drawing = false
	if g.shownSide() != d.side {
		t.Error("the human side's map is not the default")
	}
}

func TestSplitFollowsTheWalker(t *testing.T) {
	g, d := newSplitTest(t, passPlanner{}, 1800, 1000)
	far := hexgrid.Coord{Line: g.grid.Lines - 1, Col: 0}
	g.split.cams[d.side].moveTo(0, 0, true)
	d.walk = newWalkMotion([]hexgrid.Coord{far, {Line: far.Line - 2, Col: 0}}, 0, false)
	g.followSplit()
	if !viewShows(g.splitView(d.side), g.grid, d.walk.Hex()) {
		t.Error("the human panel does not follow the walker")
	}
	// A selection in view does not move the camera; scrolling away from it
	// does not snap back.
	d.walk = nil
	cam := g.split.cams[d.side]
	d.selected, d.from = true, g.splitView(d.side).Cell(2, 2)
	g.followSplit()
	g.stepView(d.side, 0, -9)
	g.followSplit()
	if g.split.cams[d.side].row == cam.row {
		t.Errorf("the camera snapped back to the selection")
	}
}

func TestWorldBounds(t *testing.T) {
	g := newMapGame(testLevel(t), false)
	w, h := g.grid.Bounds()
	if g.cam.World != (camera.Size{W: w, H: h}) {
		t.Errorf("world %v, grid bounds %vx%v", g.cam.World, w, h)
	}
}

func TestMinimapScene(t *testing.T) {
	c := newTestControl(t)
	g := newGame(testLevel(t), c.st, c.player, true)
	fills, units := g.minimapScene()
	n := 0
	for _, line := range g.lvl.Lines {
		n += len(line)
	}
	if len(fills) != n {
		t.Fatalf("%d fills for %d hexes", len(fills), n)
	}
	sides := map[minimap.Shade]int{}
	for _, u := range units {
		sides[u.Shade]++
	}
	if sides[minimap.Blue] == 0 || sides[minimap.Yellow] == 0 || len(sides) != 2 {
		t.Errorf("unit shades %v", sides)
	}
	if sideShade(c.player) != minimap.Blue {
		t.Error("the Central Powers are not blue")
	}
	if d := dimShade(minimap.Blue); d.Lit.R >= minimap.Blue.Lit.R {
		t.Error("remembered units are not dimmed")
	}
}

func TestAverageColour(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{200, 100, 0, 255})
	img.Set(1, 0, color.RGBA{0, 100, 200, 255})
	if got := averageColour(img); got != (color.RGBA{100, 100, 100, 255}) {
		t.Errorf("average %v", got)
	}
	if got := averageColour(image.NewRGBA(image.Rect(0, 0, 2, 2))); got.A != 0 {
		t.Errorf("transparent image averages to %v", got)
	}
}

func TestSplitCapturePlan(t *testing.T) {
	seen := map[string]bool{}
	for _, f := range splitFramePlan() {
		if seen[f.filename()] {
			t.Errorf("%s twice", f.filename())
		}
		seen[f.filename()] = true
	}
	for _, name := range []string{"pulse-split-1920x1080.png", "pulse-split-revealed-1920x1080.png", "pulse-split-1024x768.png"} {
		if !seen[name] {
			t.Errorf("missing %s", name)
		}
	}
}

func TestSplitZoom(t *testing.T) {
	g, d := newSplitTest(t, passPlanner{}, 1920, 1080)
	s := g.split
	levels := s.screen.ZoomLevels()
	if len(levels) < 3 || s.screen.Zoom != dosscreen.DefaultZoom {
		t.Fatalf("zoom %d of levels %v", s.screen.Zoom, levels)
	}
	at := g.startHex(d.side)
	g.setCursor(at)
	fitted := s.screen.CellScale
	// The wheel over the human half steps one level per whole turn, touchpad
	// fractions add up, and the other half does nothing.
	other := s.screen.Panels[1-d.side]
	g.wheelSplit(1, other.X+other.W/2, other.Y+other.H/2)
	if s.screen.Zoom != dosscreen.DefaultZoom {
		t.Fatalf("the other half zoomed to %d", s.screen.Zoom)
	}
	p := s.screen.Panels[d.side]
	mx, my := p.X+p.W/2, p.Y+p.H/2
	g.wheelSplit(0.5, mx, my)
	if s.screen.Zoom != dosscreen.DefaultZoom {
		t.Fatal("half a turn zoomed")
	}
	g.wheelSplit(0.5, mx, my)
	if s.screen.Zoom != dosscreen.DefaultZoom+1 || s.screen.CellScale <= fitted {
		t.Fatalf("zoom %d scale %v after a turn in", s.screen.Zoom, s.screen.CellScale)
	}
	// Every level keeps the cursor cell in view, the cameras on the map and the
	// human camera on its view; the ends hold.
	for range 10 {
		g.wheelSplit(1, mx, my)
	}
	if hi := levels[len(levels)-1]; s.screen.Zoom != hi {
		t.Errorf("zoomed in to %d, want the last level %d", s.screen.Zoom, hi)
	}
	for range 20 {
		g.wheelSplit(-1, mx, my)
		for side := range 2 {
			c := s.cams[side]
			if pp, rr := clampView(g.grid, s.screen.Cols, s.screen.Rows, c.pair, c.row); pp != c.pair || rr != c.row {
				t.Fatalf("zoom %d: camera %d off the map at %d,%d", s.screen.Zoom, side, c.pair, c.row)
			}
		}
		if !viewShows(g.splitView(d.side), g.grid, at) {
			t.Fatalf("zoom %d hides the cursor cell %v", s.screen.Zoom, at)
		}
		sx, sy := g.splitView(d.side).WorldToScreen(hexgrid.Center(at))
		if c, ok := g.pickSplit(int(sx), int(sy)); !ok || c != at {
			t.Fatalf("zoom %d: the camera does not pick %v at its centre", s.screen.Zoom, at)
		}
	}
	if lo := levels[0]; s.screen.Zoom != lo || s.screen.CellScale >= fitted {
		t.Errorf("zoomed out to %d (scale %v), want level %d", s.screen.Zoom, s.screen.CellScale, lo)
	}
	// A new window keeps the level asked for where it is usable.
	g.wheelSplit(1, mx, my)
	want := s.screen.Zoom
	g.applyLayout(1280, 720, 1)
	g.layoutSplit()
	if s.screen.Zoom != want {
		t.Errorf("resize moved the zoom from %d to %d", want, s.screen.Zoom)
	}
}
