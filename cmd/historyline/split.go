package main

import (
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
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

// keyRepeat reports whether a key held for d frames steps once more: on the
// first frame, then every few frames after a delay.
func keyRepeat(d int) bool {
	const delay, every = 18, 6
	return d == 1 || d > delay && (d-delay)%every == 0
}

// fitWell is the text of a status well that holds chars characters: the state
// and the message when both fit, else the message alone, shortened if needed.
func fitWell(state, msg string, chars int) string {
	chars = max(chars, 4)
	short := func(s string) string {
		if len(s) <= chars {
			return s
		}
		return s[:chars-3] + "..."
	}
	if msg == "" {
		return short(state)
	}
	if full := state + " - " + msg; len(full) <= chars {
		return full
	}
	return short(msg)
}

// split is the state of the split screen.
type split struct {
	screen dosscreen.Screen
	cams   [2]panelCam
	frame  dosscreen.Frame
	sel    hexgrid.Coord // the selection last followed
	selOn  bool
}

// layoutSplit lays the split screen out for the current window and keeps each
// camera on the map; a camera not placed yet starts on its side's HQ.
func (g *game) layoutSplit() {
	if g.dos == nil {
		return
	}
	if g.split == nil {
		g.split = &split{}
	}
	s := g.split
	s.screen = dosscreen.Layout(g.view.PhysW, g.view.PhysH, float64(g.view.UI))
	for side := range 2 {
		c := &s.cams[side]
		pair, row := c.pair, c.row
		if !c.set {
			pair, row = centreView(g.grid, s.screen.Cols, s.screen.Rows, g.startHex(side))
		}
		pair, row = clampView(g.grid, s.screen.Cols, s.screen.Rows, pair, row)
		c.moveTo(pair, row, true)
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
	return hexgrid.Coord{Line: g.grid.Lines / 2, Col: g.grid.Width / 2}
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
	pair, row := centreView(g.grid, s.screen.Cols, s.screen.Rows, c)
	s.cams[side].moveTo(pair, row, g.dos.reduced)
}

// stepView moves side's camera by whole column pairs and rows, on the map.
func (g *game) stepView(side, dp, dr int) {
	s := g.split
	c := s.cams[side]
	pair, row := clampView(g.grid, s.screen.Cols, s.screen.Rows, c.pair+dp, c.row+dr)
	s.cams[side].moveTo(pair, row, g.dos.reduced)
}

// updateSplit runs one frame of the cameras: arrow keys step the human side's
// panel, each panel looks where the session asks (the computer's moves and
// battles), the human side's panel follows its walker and a new selection when
// they are out of view, and the slides advance.
func (g *game) updateSplit() {
	d, s := g.dos, g.split
	dp, dr := 0, 0
	for _, k := range []struct {
		key    ebiten.Key
		dp, dr int
	}{{ebiten.KeyArrowLeft, -1, 0}, {ebiten.KeyArrowRight, 1, 0}, {ebiten.KeyArrowUp, 0, -1}, {ebiten.KeyArrowDown, 0, 1}} {
		if keyRepeat(inpututil.KeyPressDuration(k.key)) {
			dp, dr = dp+k.dp, dr+k.dr
		}
	}
	if (dp != 0 || dr != 0) && !d.MenuOpen() {
		g.stepView(d.side, dp, dr)
	}
	g.followSplit()
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
		if at := w.Hex(); !viewShows(v, g.grid, at) {
			g.lookAt(d.side, at)
		}
	} else if d.selected && (!s.selOn || s.sel != d.from) && !viewShows(v, g.grid, d.from) {
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
	c, ok := g.grid.PixelToHex(g.cam.ScreenToWorld(float64(mx), float64(my)))
	return c, ok && viewShows(v, g.grid, c)
}

// drawSplit draws both panels' maps, each through its own camera into its
// panel, and the steel frame over them.
func (g *game) drawSplit(screen *ebiten.Image) {
	s := g.split
	human := *g.cam
	defer func() { *g.cam, g.drawing = human, false }()
	var views [2]dosscreen.View
	for side := range 2 {
		views[side] = g.splitView(side)
		p := views[side].Panel
		*g.cam = g.splitCamera(side)
		g.drawing, g.drawSide = true, side
		dst := screen.SubImage(image.Rect(p.X, p.Y, p.Right(), p.Bottom())).(*ebiten.Image)
		x0, y0 := g.cam.ScreenToWorld(float64(p.X), float64(p.Y))
		x1, y1 := g.cam.ScreenToWorld(float64(p.Right()), float64(p.Bottom()))
		g.drawMapLayer(dst, x0, y0, x1-x0, y1-y0, side == g.dos.side)
	}
	*g.cam, g.drawing = human, false
	s.frame.Draw(screen, s.screen, views, g.grid)
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
			if !g.grid.Contains(c) {
				continue
			}
			sx, sy := cam.WorldToScreen(hexgrid.Center(c))
			hx, hy := g.view.toHUD(int(sx), int(sy+22*cam.Effective()))
			ebitenutil.DebugPrintAt(dst, fmt.Sprintf("%d,%d", c.Line, c.Col), hx-14, hy-8)
		}
	}
}

// wellText is side's status well: the turn, its phase and its orders or
// readiness, and for the human side the last message.
func (g *game) wellText(side int) (state, msg string) {
	d, s := g.dos, g.dos.s
	state = fmt.Sprintf("TURN %d  %v: %s PHASE", s.Turn()+1, s.Faction(side), phaseName(s.Phase(side)))
	switch {
	case s.Ready(side):
		state += ", READY"
	case side == d.side && s.Game.Limit > 0:
		state += fmt.Sprintf(", %d OF %d ORDERS", len(s.Game.Players[side].Orders), s.Game.Limit)
	case side == d.side:
		state += ", " + count(len(s.Game.Players[side].Orders), "order")
	}
	if side == d.side {
		msg = d.Message
	}
	return strings.ToUpper(state), strings.ToUpper(msg)
}

// drawWells prints each side's status line centred in its well (HUD).
func (g *game) drawWells(dst *ebiten.Image) {
	ui := max(g.view.UI, 1)
	for side := range 2 {
		r := g.split.screen.Status[side]
		x, y, w, h := r.X/ui, r.Y/ui, r.W/ui, r.H/ui
		state, msg := g.wellText(side)
		text := fitWell(state, msg, (w-8)/glyphW)
		ebitenutil.DebugPrintAt(dst, text, x+(w-textWidth(text))/2, y+(h-16)/2)
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
