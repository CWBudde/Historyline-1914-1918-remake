package main

import (
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/gesture"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// The split battle screen of -rules dos: one map panel per side (left side 0,
// right side 1), each with its own camera that rests on whole column pairs and
// rows, inside the steel frame of internal/dosscreen, with a status well under
// each panel. Only the human side's panel takes the mouse.

// slideFrames is how long a panel camera slides to a new resting place.
const slideFrames = 8

// panelCam is one panel's camera: the first column pair and row it rests on
// and the world offset still to slide; the panel is drawn at the resting origin
// minus the slide.
type panelCam struct {
	pair, row      int
	slideX, slideY float64
	left           int // frames of the slide left
	set            bool
}

// moveTo rests the camera on pair, row. A change slides there over slideFrames
// unless reduced (reduced motion, a new layout), from wherever the panel is
// drawn now.
func (c *panelCam) moveTo(pair, row int, reduced bool) {
	if c.set && !reduced && (pair != c.pair || row != c.row) {
		c.slideX += float64((pair - c.pair) * hexgrid.ColPitch)
		c.slideY += float64((row - c.row) * hexgrid.HexHeight)
		c.left = slideFrames
	}
	if reduced {
		c.slideX, c.slideY, c.left = 0, 0, 0
	}
	c.pair, c.row, c.set = pair, row, true
}

// tick advances the slide by one frame.
func (c *panelCam) tick() {
	if c.left <= 0 {
		return
	}
	c.slideX -= c.slideX / float64(c.left)
	c.slideY -= c.slideY / float64(c.left)
	c.left--
	if c.left == 0 {
		c.slideX, c.slideY = 0, 0
	}
}

// floorDiv is a/b rounded down.
func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// pairRange is the range of first column pairs of a panel cols columns wide
// that keeps the block on the map; a map narrower than the block gets the one
// pair that centres it. The map has 2·Width-1 columns (even and odd lines) and
// a block starts on an even one, so a block of an even number of columns may
// end one column beyond the map (steel) to reach the last.
func pairRange(g hexgrid.Grid, cols int) (lo, hi int) {
	span := 2*g.Width - 1 - cols
	if span >= 0 {
		return 0, (span + 1) / 2
	}
	c := floorDiv(span, 4)
	return c, c
}

// rowRange is the range of first rows of a panel of rows cells per column that
// keeps the block on the map, centred on a map shorter than the block. A row is
// two lines; the even lines decide the number of rows.
func rowRange(g hexgrid.Grid, rows int) (lo, hi int) {
	span := (g.Lines+1)/2 - rows
	if span >= 0 {
		return 0, span
	}
	c := floorDiv(span, 2)
	return c, c
}

// clampView keeps a view of cols × rows cells starting at pair, row on the map.
func clampView(g hexgrid.Grid, cols, rows, pair, row int) (int, int) {
	plo, phi := pairRange(g, cols)
	rlo, rhi := rowRange(g, rows)
	return min(max(pair, plo), phi), min(max(row, rlo), rhi)
}

// centreView is the clamped resting place of a cols × rows view with hex c as
// near its middle as whole column pairs and rows allow.
func centreView(g hexgrid.Grid, cols, rows int, c hexgrid.Coord) (pair, row int) {
	k := 2*c.Col + c.Line&1 // c's column in the sequence of even and odd columns
	return clampView(g, cols, rows, floorDiv(k-(cols-1)/2, 2), c.Line>>1-(rows-1)/2)
}

// viewShows reports whether c is an on-map cell of v: an opening of its frame.
func viewShows(v dosscreen.View, g hexgrid.Grid, c hexgrid.Coord) bool {
	j := 2*(c.Col-v.Pair) + c.Line&1
	i := c.Line>>1 - v.Row
	return j >= 0 && j < v.Cols && i >= 0 && i < v.Rows && g.Contains(c)
}

// panelCamera is a camera that draws in screen pixels exactly as v does
// (dosscreen.View.WorldToScreen), shifted by the slide still to go, so the map
// helpers that draw through a camera.Camera draw into the panel. View is the
// screen, so Visible is the world under the whole screen.
func panelCamera(v dosscreen.View, slideX, slideY float64, world, screen camera.Size, ui float64) camera.Camera {
	ox, oy := v.Origin()
	bx, by := v.Offset()
	ui = max(ui, 1)
	return camera.Camera{
		World: world, View: screen, Zoom: v.Scale / ui, Scale: ui,
		Origin: camera.Point{X: ox - slideX - float64(bx)/v.Scale, Y: oy - slideY - float64(by)/v.Scale},
	}
}

// split is the state of the split screen.
type split struct {
	screen dosscreen.Screen
	cams   [2]panelCam
	frame  dosscreen.Frame
	sel    hexgrid.Coord // the selection last followed
	selOn  bool

	track    *gesture.Tracker // the human side's gestures (cursor.go)
	dir      gesture.Dir      // the direction its hold picks
	mode     [2]panelMode     // what each panel shows
	strat    [2]strategicPic  // each panel's strategic map
	grey     [2]*ebiten.Image // each panel drawn offscreen for the greyscale (phase.go)
	infoUnit dosplay.UnitView // the unit whose type the unit info shows

	placement gesture.Placement // on the cursor cell (cursorPlacement)
	placeKey  placeKey
	placed    bool

	edgeOn   bool    // the pointer rests on an edge (edgeScroll)
	edgeWait int     // frames until its next step
	zoom     int     // the zoom level asked for (dosscreen.Zooms), kept over resizes
	wheel    float64 // wheel turns not yet stepped (touchpads turn in fractions)
	lastMsg  string  // the human side.s message, aged by tickWells
	msgAge   int
}

// playGrid is the part of the level the original rules play on: the bound
// board without its edge row and column (dosbind.Board.Playable), else the
// whole level. The split screen's cameras, frame, cursor and strategic map stay
// on it, so the edge is never shown.
func (g *game) playGrid() hexgrid.Grid {
	if g.dos != nil && g.dos.s != nil && g.dos.s.Board != nil {
		return g.dos.s.Board.Playable()
	}
	return g.grid
}

// layoutSplit lays the split screen out for the current window and keeps each
// camera on the map; a camera not placed yet starts on its side's HQ.
func (g *game) layoutSplit() {
	if g.dos == nil {
		return
	}
	if g.split == nil {
		g.split = &split{track: newTracker(), zoom: dosscreen.DefaultZoom}
	}
	s := g.split
	s.screen = dosscreen.Layout(g.view.PhysW, g.view.PhysH, float64(g.view.UI)).Zoomed(s.zoom)
	s.track.Threshold = gestureThreshold(s.screen.CellScale)
	for side := range 2 {
		c := &s.cams[side]
		pair, row := c.pair, c.row
		if !c.set {
			pair, row = centreView(g.playGrid(), s.screen.Cols, s.screen.Rows, g.startHex(side))
		}
		pair, row = clampView(g.playGrid(), s.screen.Cols, s.screen.Rows, pair, row)
		c.moveTo(pair, row, true)
	}
	if !g.hoverOnMap {
		g.setCursor(g.startHex(g.dos.side))
	}
	g.syncSplitCam()
}

// startHex is where side's panel opens: its HQ, else its first unit, else the
// middle of the map.
func (g *game) startHex(side int) hexgrid.Coord {
	d := g.dos
	bs := d.buildingsOf(side)
	for _, c := range dosplay.Coords(toSet(bs)) {
		if b := bs[c]; b.HQ && b.Owner == d.s.Faction(side) {
			return c
		}
	}
	us := d.unitsOf(side, false)
	for _, c := range dosplay.Coords(toSet(us)) {
		if us[c].Side == side {
			return c
		}
	}
	return hexgrid.Coord{Line: g.playGrid().Lines / 2, Col: g.playGrid().Width / 2}
}

// toSet is the set of a map's hexes.
func toSet[V any](m map[hexgrid.Coord]V) map[hexgrid.Coord]bool {
	out := make(map[hexgrid.Coord]bool, len(m))
	for k := range m {
		out[k] = true
	}
	return out
}

// splitView is side's panel view at its resting place.
func (g *game) splitView(side int) dosscreen.View {
	c := g.split.cams[side]
	return g.split.screen.View(side, c.pair, c.row)
}

// splitCamera is side's panel camera as drawn now (with its slide).
func (g *game) splitCamera(side int) camera.Camera {
	c := g.split.cams[side]
	return panelCamera(g.splitView(side), c.slideX, c.slideY, g.cam.World,
		camera.Size{W: float64(g.view.PhysW), H: float64(g.view.PhysH)}, float64(g.view.UI))
}

// syncSplitCam makes g.cam the human side's panel camera, so picking, the order
// numbers and the info box work through it.
func (g *game) syncSplitCam() { *g.cam = g.splitCamera(g.dos.side) }

// lookAt rests side's camera with c near the middle.
func (g *game) lookAt(side int, c hexgrid.Coord) {
	s := g.split
	pair, row := centreView(g.playGrid(), s.screen.Cols, s.screen.Rows, c)
	s.cams[side].moveTo(pair, row, g.dos.reduced)
}

// stepView moves side's camera by whole column pairs and rows, on the map.
func (g *game) stepView(side, dp, dr int) {
	s := g.split
	c := s.cams[side]
	pair, row := clampView(g.playGrid(), s.screen.Cols, s.screen.Rows, c.pair+dp, c.row+dr)
	s.cams[side].moveTo(pair, row, g.dos.reduced)
}

// updateSplit runs one frame of the cameras: each panel looks where the
// session asks (the computer's moves and battles), the human side's panel
// follows its walker and a new selection when they are out of view (and its
// cursor, see showCursor), and the slides advance.
func (g *game) updateSplit() {
	s := g.split
	if g.dos.locked() {
		s.mode = [2]panelMode{} // the phase change closes the strategic maps and plates
		g.dos.box = nil
	}
	g.syncContents()
	g.followSplit()
	g.tickWells()
	for side := range 2 {
		s.cams[side].tick()
	}
	g.syncSplitCam()
}

// followSplit applies the session's look requests and the human side's
// follow rules.
func (g *game) followSplit() {
	d, s := g.dos, g.split
	for side := range 2 {
		if c, ok := d.takeFollow(side); ok {
			g.lookAt(side, c)
		}
	}
	v := g.splitView(d.side)
	if w := d.walk; !w.Done() {
		if at := w.Hex(); !viewShows(v, g.playGrid(), at) {
			g.lookAt(d.side, at)
		}
	} else if d.selected && (!s.selOn || s.sel != d.from) && !viewShows(v, g.playGrid(), d.from) {
		g.lookAt(d.side, d.from)
	}
	s.sel, s.selOn = d.from, d.selected
}

// pickSplit is the hex under device pixel (mx, my): only on the human side's
// panel, and only through an opening of its frame.
func (g *game) pickSplit(mx, my int) (hexgrid.Coord, bool) {
	v := g.splitView(g.dos.side)
	p := v.Panel
	if mx < p.X || my < p.Y || mx >= p.Right() || my >= p.Bottom() {
		return hexgrid.Coord{}, false
	}
	c, ok := g.playGrid().PixelToHex(g.cam.ScreenToWorld(float64(mx), float64(my)))
	return c, ok && viewShows(v, g.playGrid(), c)
}

// drawSplit draws both panels' maps (or strategic maps), each through its own
// camera into its panel and in the look the phase change asks for (phase.go),
// the steel frame over them (resting on the map's first cells where a plate or
// the battle scene covers the map) and the human side's cursor.
func (g *game) drawSplit(screen *ebiten.Image) {
	s := g.split
	human := *g.cam
	defer func() { *g.cam, g.drawing = human, false }()
	var views [2]dosscreen.View
	looks := g.dos.panelLooks()
	for side := range 2 {
		views[side] = g.splitView(side)
		g.drawPanelLook(screen, side, looks[side], func(dst *ebiten.Image, dx, dy float64) {
			g.drawPanel(dst, side, views[side].Panel, dx, dy)
		})
	}
	*g.cam, g.drawing = human, false
	for side := range 2 {
		if s.mode[side] != panelMap && s.mode[side] != panelGameInfo || looks[side] == lookBattle || looks[side] == lookCaptured {
			views[side] = g.plateView(side)
		}
	}
	s.frame.Draw(screen, s.screen, views, g.playGrid())
	if g.dos.prompt || g.dos.stage == stageBlack {
		g.cursorPlates(screen)
		return
	}
	if !g.dos.sequence() || g.dos.stage == stageNewMode {
		g.drawCursor(screen)
	}
}

// plateView is the view whose frame side's panel has while a plate covers
// its map: one resting on the map's first cells, so the frame does not close
// in where the map ends (unless the map is smaller than the view).
func (g *game) plateView(side int) dosscreen.View {
	s := g.split
	pair, row := clampView(g.playGrid(), s.screen.Cols, s.screen.Rows, 0, 0)
	return s.screen.View(side, pair, row)
}

// drawPanel draws side's map (or the view open there) in panel p onto dst,
// whose origin lies (dx, dy) device pixels from the screen's. The phase change
// closes the views (updateSplit), so they are drawn unshifted.
func (g *game) drawPanel(dst *ebiten.Image, side int, p dosscreen.Rect, dx, dy float64) {
	mode := g.split.mode[side]
	if dx != 0 || dy != 0 {
		mode = panelMap
	}
	switch mode {
	case panelStrategic:
		g.drawStrategic(dst, side)
		return
	case panelUnitInfo, panelContents:
		g.drawOverlay(dst, side, g.panelRect(side))
		return
	case panelMap, panelGameInfo:
	}
	*g.cam = g.splitCamera(side)
	g.drawing, g.drawSide = true, side
	x0, y0 := g.cam.ScreenToWorld(float64(p.X), float64(p.Y))
	x1, y1 := g.cam.ScreenToWorld(float64(p.Right()), float64(p.Bottom()))
	e := g.cam.Effective()
	g.cam.Origin.X -= dx / e
	g.cam.Origin.Y -= dy / e
	r := image.Rect(p.X, p.Y, p.Right(), p.Bottom()).Add(image.Pt(int(dx), int(dy)))
	g.drawMapLayer(dst.SubImage(r).(*ebiten.Image), x0, y0, x1-x0, y1-y0, side == g.dos.side)
	if mode == panelGameInfo {
		g.drawOverlay(dst, side, g.panelRect(side))
	}
}

// shownSide is the side whose map is drawn: the panel being drawn, else the
// human side.
func (g *game) shownSide() int {
	if g.drawing {
		return g.drawSide
	}
	return g.dos.side
}

// drawSplitHexNumbers prints "line,col" in every opening of both panels (dev
// flag), on the HUD.
func (g *game) drawSplitHexNumbers(dst *ebiten.Image) {
	for side := range 2 {
		v := g.splitView(side)
		cam := g.splitCamera(side)
		for _, c := range v.Cells() {
			if !g.playGrid().Contains(c) {
				continue
			}
			sx, sy := cam.WorldToScreen(hexgrid.Center(c))
			hx, hy := g.view.toHUD(int(sx), int(sy+22*cam.Effective()))
			ebitenutil.DebugPrintAt(dst, fmt.Sprintf("%d,%d", c.Line, c.Col), hx-14, hy-8)
		}
	}
}

// splitErrors are the art failures, shown at the top left under -rules dos.
func (g *game) splitErrors() string {
	var out []string
	if g.artErr != "" {
		out = append(out, "unit sprites: "+g.artErr)
	}
	if g.terrainErr != "" {
		out = append(out, "terrain tiles: "+g.terrainErr)
	}
	return strings.Join(out, "\n")
}

// windowLimit is the window size limit set last, in device-independent pixels.
var windowLimit = [2]int{-1, -1}

// limitWindow sets the window's smallest size (w, h ≤ 0: none) if it changed.
func limitWindow(w, h int) {
	if w <= 0 || h <= 0 {
		w, h = -1, -1
	}
	if windowLimit != [2]int{w, h} {
		windowLimit = [2]int{w, h}
		ebiten.SetWindowSizeLimits(w, h, -1, -1)
	}
}

// splitMinWindow is the smallest window of the split screen in
// device-independent pixels at display scale dpi.
func splitMinWindow(dpi float64) (w, h int) {
	dpi = max(dpi, 1e-3)
	pw, ph := dosscreen.MinWindow(float64(max(int(dpi+0.5), 1)))
	return int(math.Ceil(float64(pw) / dpi)), int(math.Ceil(float64(ph) / dpi))
}

// zoomSplit steps both panels' zoom level by step (the mouse wheel), keeping
// the human side's cursor cell, or else the middle of each view, in the
// middle; the levels end where the window has no more (dosscreen.ZoomLevels).
func (g *game) zoomSplit(step int) {
	s := g.split
	base := dosscreen.Layout(g.view.PhysW, g.view.PhysH, float64(g.view.UI))
	z := base.ClampZoom(s.screen.Zoom + step)
	if z == s.screen.Zoom {
		return
	}
	var keep [2]hexgrid.Coord
	for side := range 2 {
		v := g.splitView(side)
		keep[side] = v.Cell(v.Cols/2, v.Rows/2)
	}
	if g.hoverOnMap {
		keep[g.dos.side] = g.hover
	}
	s.zoom = z
	s.screen = base.Zoomed(z)
	s.track.Threshold = gestureThreshold(s.screen.CellScale)
	for side := range 2 {
		pair, row := centreView(g.playGrid(), s.screen.Cols, s.screen.Rows, keep[side])
		s.cams[side].moveTo(pair, row, true)
		s.strat[side] = strategicPic{}
	}
	g.syncSplitCam()
}

// wheelSplit adds a wheel turn of dy with the pointer at (mx, my): over the
// human side's half of the window each whole turn steps the zoom, up zooming
// in.
func (g *game) wheelSplit(dy float64, mx, my int) {
	s := g.split
	if !g.edgeZone(mx, my) || s.mode[g.dos.side] != panelMap || g.dos.locked() {
		s.wheel = 0
		return
	}
	s.wheel += dy
	for s.wheel >= 1 {
		s.wheel--
		g.zoomSplit(1)
	}
	for s.wheel <= -1 {
		s.wheel++
		g.zoomSplit(-1)
	}
}
