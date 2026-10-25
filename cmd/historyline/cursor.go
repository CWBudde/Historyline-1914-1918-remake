package main

import (
	"image"
	"image/color"
	"math"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/gesture"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/minimap"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// The cursor of -rules dos: the human side's cursor cell (g.hover), moved by
// the mouse and by the arrows of key set A, and the gestures of
// internal/gesture on it. Holding fire (left button, Space) shows the cancel
// plate; a drag of a cell or an arrow picks the action placed there and
// releasing runs it; a plain click selects or confirms as before; right-click
// opens the ring of every possible action. While a move or an attack is aimed,
// every cell out of reach is darkened and the plate with the order follows
// the cursor over the cells in reach. The map gesture opens the strategic map
// in the side's panel, the info gesture the unit or game info (overlay.go)
// and the open gesture the contents (contents.go). The OS pointer is hidden
// over the map cells: the ring is the pointer there.

// Cursor tuning: the arrow repeat (as the view keys had) and the share of
// brightness an out-of-reach cell keeps.
const (
	repeatDelay = 18 // frames an arrow is held before the cursor repeats
	repeatEvery = 6
	darkKeep    = 0.62
)

// panelMode is what a panel shows.
type panelMode int

const (
	panelMap       panelMode = iota // the side's map
	panelStrategic                  // the whole map at a glance
	panelUnitInfo                   // the unit info of a type (overlay.go)
	panelGameInfo                   // the game info over the darkened map
	panelContents                   // a container's slots (contents.go)
)

// newTracker is the gesture tracker of the human side's panel; layoutSplit
// sets its threshold to the cell scale.
func newTracker() *gesture.Tracker {
	return gesture.NewTracker(gesture.Config{RepeatDelay: repeatDelay, RepeatEvery: repeatEvery})
}

// gestureThreshold is the drag that picks: the column pitch between two
// neighbouring cells (HalfPitch) on screen, a little less than a cell.
func gestureThreshold(scale float64) float64 { return hexgrid.HalfPitch * scale }

// picking is the order being aimed: Move while the selected unit's
// destinations show, Attack while its targets do, else Cancel.
func (d *dosControl) picking() gesture.Action {
	switch {
	case !d.selected:
	case len(d.moves) > 0:
		return gesture.Move
	case len(d.targets) > 0:
		return gesture.Attack
	}
	return gesture.Cancel
}

// reach is the cells that keep their colour while an order is aimed: the
// destinations or targets and the unit's own cell; nil while none is aimed.
func (d *dosControl) reach() map[hexgrid.Coord]bool {
	if d.picking() == gesture.Cancel {
		return nil
	}
	out := map[hexgrid.Coord]bool{d.from: true}
	for _, set := range []map[hexgrid.Coord]bool{d.moves, d.targets} {
		for c, ok := range set {
			if ok {
				out[c] = true
			}
		}
	}
	return out
}

// onTarget reports whether c is a destination or target of the aimed order.
func (d *dosControl) onTarget(c hexgrid.Coord) bool { return d.moves[c] || d.targets[c] }

// ownContainer reports whether c holds an own building or carrier, whose
// contents the left gesture opens.
func (d *dosControl) ownContainer(c hexgrid.Coord) bool {
	v := d.own()
	if b, ok := v.Buildings[c]; ok && b.Owner == d.s.Faction(d.side) {
		return true
	}
	u, ok := v.Units[c]
	return ok && u.Side == d.side && u.Carrier
}

// can reports whether action a is possible on c. The map and the info always
// are; orders wait while the resolution plays and after the end, open needs
// an own container with units or something to build, and move and attack
// need an own unit that may act and has somewhere to go or shoot.
func (d *dosControl) can(a gesture.Action, c hexgrid.Coord) bool {
	switch a {
	case gesture.Map, gesture.Info:
		return true
	case gesture.Cancel:
		return true
	}
	if d.over || d.busy() {
		return false
	}
	switch a {
	case gesture.ModeChange:
		return true
	case gesture.Open:
		hold, units, ok := d.s.Contents(d.side, c)
		return ok && (len(units) > 0 || len(d.s.Buildable(d.side, hold)) > 0)
	case gesture.Move, gesture.Attack:
		u, _, ok := shipHead(d.own().Units, c)
		if !ok || u.Side != d.side || d.s.Ready(d.side) || d.s.CanOrder(d.side, u.Index) != nil {
			return false
		}
		if a == gesture.Attack {
			return len(d.s.Targets(d.side, u.Index)) > 0
		}
		dests, err := d.s.Destinations(d.side, u.Index)
		return err == nil && len(dests) > 0
	default:
	}
	return false
}

// placeKey is what the placement on the cursor cell depends on.
type placeKey struct {
	at, from          hexgrid.Coord
	version           int
	pick              gesture.Action
	ready, busy, over bool
}

// cursorPlacement is the placement on the cursor cell, computed again only
// when the cell or the session changed.
func (g *game) cursorPlacement() gesture.Placement {
	s, d := g.split, g.dos
	k := placeKey{at: g.hover, from: d.from, version: d.s.Version(), pick: d.picking(), ready: d.s.Ready(d.side), busy: d.busy(), over: d.over}
	if !s.placed || s.placeKey != k {
		s.placement, s.placeKey, s.placed = d.placement(g.hover), k, true
	}
	return s.placement
}

// placement is the gesture placement on c: the map's (attack above in the
// action phase, open on an own container) restricted to what is possible.
// While an order is aimed every direction cancels: a click runs it, a drag
// drops it.
func (d *dosControl) placement(c hexgrid.Coord) gesture.Placement {
	if d.picking() != gesture.Cancel {
		return gesture.Placement{}
	}
	attack := d.s.Phase(d.side) == dosrules.PhaseAction
	return gesture.ForMap(attack, d.ownContainer(c)).Restrict(func(a gesture.Action) bool { return d.can(a, c) })
}

// order runs the session's part of a gesture on c: move and attack select the
// unit there to aim its order, mode change confirms (or opens the orders
// again, as F1), open opens the contents. A gesture during a walk only
// finishes the walk.
func (d *dosControl) order(a gesture.Action, c hexgrid.Coord) {
	if d.SkipWalk() {
		return
	}
	d.Message = ""
	d.fx.Dismiss()
	d.menu = nil
	switch a {
	case gesture.Move, gesture.Attack:
		d.selectAt(c)
	case gesture.ModeChange:
		d.EndTurn()
	case gesture.Open:
		d.clear()
		d.openContents(c)
	default:
	}
}

// setCursor puts the cursor on c.
func (g *game) setCursor(c hexgrid.Coord) { g.hover, g.hoverOnMap = c, true }

// keepInView is the resting place nearest to a cols × rows view at pair, row
// that shows c, on the map.
func keepInView(g hexgrid.Grid, cols, rows, pair, row int, c hexgrid.Coord) (int, int) {
	k := 2*c.Col + c.Line&1 // c's column in the sequence of even and odd columns
	i := c.Line >> 1
	pair = min(max(pair, -floorDiv(cols-1-k, 2)), floorDiv(k, 2))
	row = min(max(row, i-rows+1), i)
	return clampView(g, cols, rows, pair, row)
}

// showCursor moves the human side's camera as little as it takes to show the
// cursor.
func (g *game) showCursor() {
	s, side := g.split, g.dos.side
	c := &s.cams[side]
	pair, row := keepInView(g.playGrid(), s.screen.Cols, s.screen.Rows, c.pair, c.row, g.hover)
	c.moveTo(pair, row, g.dos.reduced)
	g.syncSplitCam()
}

// gestureInput is this frame's mouse and key set A for the human side.
func (g *game) gestureInput() gesture.Input {
	mx, my := ebiten.CursorPosition()
	p := gesture.Pointer{
		X: float64(mx), Y: float64(my), Active: g.pointerActive(mx, my),
		Down:         ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft),
		RightPressed: inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight),
	}
	return gesture.Input{Pointer: p, Keys: gesture.KeySetA.Read(ebiten.IsKeyPressed)}
}

// pointerActive reports whether the tracker takes the mouse at (mx, my): over
// the human side's panel away from a menu row, anywhere while a gesture is
// held or the ring or a view other than the map is open.
func (g *game) pointerActive(mx, my int) bool {
	s := g.split
	if _, ok := s.track.Ring(); ok || s.mode[g.dos.side] != panelMap {
		return true
	}
	if _, _, held := s.track.Plate(); held {
		return true
	}
	p := s.screen.Panels[g.dos.side]
	return mx >= p.X && my >= p.Y && mx < p.Right() && my < p.Bottom() && !g.overMenu(mx, my)
}

// overMenu reports whether device pixel (mx, my) is on a row of the open menu.
func (g *game) overMenu(mx, my int) bool {
	ui := g.drv().ui()
	if !ui.MenuOpen() {
		return false
	}
	hx, hy := g.view.toHUD(mx, my)
	return menuHit(hx, hy, len(ui.menu), g.menuRow) != menuNone
}

// hidePointer reports whether the OS pointer is hidden at (mx, my): only under
// -rules dos while the battle runs, over a cell of the human side's map (not
// on a menu row, the ring or another view) or while the mouse holds a
// gesture.
func (g *game) hidePointer(mx, my int, down bool) bool {
	if g.split == nil || g.dos == nil || g.dos.over {
		return false
	}
	s := g.split
	if s.mode[g.dos.side] != panelMap {
		return false
	}
	if _, ok := s.track.Ring(); ok {
		return false
	}
	if _, _, held := s.track.Plate(); held && down {
		return true
	}
	_, onCell := g.pickSplit(mx, my)
	return onCell && !g.overMenu(mx, my)
}

// pointerHidden is the OS pointer mode set last.
var pointerHidden bool

// showPointer shows or hides the OS pointer if that changed.
func showPointer(visible bool) {
	if pointerHidden != visible {
		return
	}
	pointerHidden = !visible
	if visible {
		ebiten.SetCursorMode(ebiten.CursorModeVisible)
	} else {
		ebiten.SetCursorMode(ebiten.CursorModeHidden)
	}
}

// updateCursor runs one frame of the human side's input; menu tells that a
// menu took the keys this frame (it was open before or is open now).
func (g *game) updateCursor(menu bool) {
	g.feedGesture(g.gestureInput(), menu)
	mx, my := ebiten.CursorPosition()
	g.edgeScroll(mx, my, g.edgeActive(menu))
	showPointer(!g.hidePointer(mx, my, ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)))
}

// feedGesture feeds one frame to the tracker and runs what it recognized.
// While a menu has the keys the arrows are withheld and a fire press is
// dropped, so the Space that chose a row does not click the map.
func (g *game) feedGesture(in gesture.Input, menu bool) {
	t := g.split.track
	if menu {
		in.Keys = gesture.Keys{Fire: in.Keys.Fire}
	}
	t.Placement = g.panelPlacement()
	for _, ev := range t.Update(in) {
		if menu && !ev.Mouse {
			if ev.Kind == gesture.Press {
				t.Reset()
			}
			continue
		}
		g.onGesture(ev)
		g.syncContents()
		t.Placement = g.panelPlacement()
	}
}

// panelPlacement is the gesture placement of the human side's panel: on the
// cursor cell of the map, on the contents cursor, cancel on every other view.
func (g *game) panelPlacement() gesture.Placement {
	switch g.split.mode[g.dos.side] {
	case panelMap:
		return g.cursorPlacement()
	case panelContents:
		return g.dos.contentsPlacement()
	default:
	}
	return gesture.Placement{}
}

// onGesture runs one recognized event on the human side's panel.
func (g *game) onGesture(ev gesture.Event) {
	s, d := g.split, g.dos
	t := s.track
	if d.locked() {
		g.lockedGesture(ev)
		return
	}
	switch s.mode[d.side] {
	case panelStrategic:
		g.strategicGesture(ev)
		return
	case panelUnitInfo, panelGameInfo:
		g.infoGesture(ev)
		return
	case panelContents:
		g.contentsGesture(ev)
		return
	case panelMap:
	}
	switch ev.Kind {
	case gesture.Hover:
		if c, ok := g.pickSplit(int(ev.X), int(ev.Y)); ok {
			g.setCursor(c)
		}
	case gesture.Step:
		if c := gesture.Neighbor(g.hover, ev.Dir); g.playGrid().Contains(c) {
			g.setCursor(c)
			g.showCursor()
		}
	case gesture.Press:
		s.dir = gesture.Centre
		if ev.Mouse {
			c, ok := g.pickSplit(int(ev.X), int(ev.Y))
			if !ok {
				t.Reset() // a press on the steel holds nothing
				return
			}
			g.setCursor(c)
			t.Placement = d.placement(c)
		}
	case gesture.Pick:
		s.dir = ev.Dir
	case gesture.Click:
		if g.hoverOnMap {
			d.Click(g.hover)
		}
	case gesture.Run:
		g.runGesture(ev.Action, g.hover)
	case gesture.Cancelled:
		if d.picking() != gesture.Cancel {
			d.Clear()
		}
	case gesture.RingOpen:
		g.openRing(ev)
	}
}

// lockedGesture handles the human side's input in the prompt and while the
// phase change plays: fire (a click or a run) goes back from the prompt; the
// cursor stays and no ring opens. A press skips a battle in Update.
func (g *game) lockedGesture(ev gesture.Event) {
	switch ev.Kind {
	case gesture.Click, gesture.Run:
		if g.dos.prompt {
			g.dos.closePrompt()
		}
	case gesture.RingOpen:
		g.split.track.Reset()
	default:
	}
}

// runGesture runs action a on c: the map and info on the panel, the orders
// through the session.
func (g *game) runGesture(a gesture.Action, c hexgrid.Coord) {
	switch a {
	case gesture.Map:
		g.openStrategic(g.dos.side)
	case gesture.Info:
		g.openInfo(c)
	default:
		g.dos.order(a, c)
	}
}

// openRing opens the ring of the actions possible on the cell under a right
// click. A right click while an order is aimed drops the aim instead.
func (g *game) openRing(ev gesture.Event) {
	t, d := g.split.track, g.dos
	if d.picking() != gesture.Cancel {
		d.Clear()
		t.Reset()
		return
	}
	c, ok := g.pickSplit(int(ev.X), int(ev.Y))
	if !ok {
		t.Reset()
		return
	}
	g.setCursor(c)
	p := d.placement(c)
	t.Placement = p
	x, y := g.cam.WorldToScreen(hexgrid.Center(c))
	t.SetRing(fitRing(gesture.NewRing(x, y, g.cam.Effective(), p.RingActions()), g.split.screen.Panels[d.side]))
}

// fitRing moves ring r as little as it takes to keep all its plates inside
// the panel p, where they fit; a ring on an edge cell leaves the cell.
func fitRing(r gesture.Ring, p dosscreen.Rect) gesture.Ring {
	hw, hh := hexgrid.HexWidth/2*r.Scale, hexgrid.HexHeight/2*r.Scale
	x0, y0, x1, y1 := r.X-hw, r.Y-hh, r.X+hw, r.Y+hh
	for i := range r.Actions {
		x, y := r.Pos(i)
		x0, y0, x1, y1 = min(x0, x-hw), min(y0, y-hh), max(x1, x+hw), max(y1, y+hh)
	}
	r.X += shiftInto(x0, x1, float64(p.X), float64(p.Right()))
	r.Y += shiftInto(y0, y1, float64(p.Y), float64(p.Bottom()))
	return r
}

// shiftInto is the shift that moves the span lo..hi into a..b; a span wider
// than a..b keeps its start at a.
func shiftInto(lo, hi, a, b float64) float64 {
	switch {
	case lo < a:
		return a - lo
	case hi > b:
		return max(b-hi, a-lo)
	}
	return 0
}

// closeGestures closes the view of the human side's panel (the production
// list goes back to the slots), the ring or a held gesture (Esc) and reports
// whether there was one.
func (g *game) closeGestures() bool {
	s, d := g.split, g.dos
	if s.mode[d.side] != panelMap {
		s.track.Reset()
		if b := d.box; b != nil && b.build {
			b.build = false
			return true
		}
		s.mode[d.side], d.box = panelMap, nil
		return true
	}
	_, ring := s.track.Ring()
	_, _, held := s.track.Plate()
	if ring || held {
		s.track.Reset()
		return true
	}
	return false
}

// cursorPlate is what the cursor shows: a plate (its action and state) or the
// plain ring (plate false). pick is the aimed order, onTarget whether the
// cursor is on a cell it reaches, held whether fire is held, centre whether
// the hold picks no direction, and a, ok the tracker's plate. While an order
// is aimed its plate follows the cursor over the cells in reach (pressed
// while fire is held, disabled elsewhere while held); otherwise fire turns
// the ring into the plate of the picked action, cancel first.
func cursorPlate(pick gesture.Action, onTarget, held, centre bool, a gesture.Action, ok bool) (gesture.Action, gesture.State, bool) {
	switch {
	case pick != gesture.Cancel && held && centre && onTarget:
		return pick, gesture.Pressed, true
	case pick != gesture.Cancel && held && centre:
		return pick, gesture.Disabled, true
	case pick != gesture.Cancel && !held:
		return pick, gesture.Normal, onTarget
	case !held:
		return gesture.Cancel, gesture.Normal, false
	case !ok:
		return a, gesture.Disabled, true
	}
	return a, gesture.Normal, true
}

// drawCursor draws the human side's cursor over its panel: the ring, the
// plate, or the open ring menu; in the contents the plate on the slot.
func (g *game) drawCursor(dst *ebiten.Image) {
	s, d := g.split, g.dos
	if s.mode[d.side] == panelContents {
		g.drawContentsPlate(dst)
	}
	if s.mode[d.side] != panelMap {
		return
	}
	t := s.track
	if r, ok := t.Ring(); ok {
		gesture.DrawRingMenu(dst, r, t.RingHover())
		return
	}
	if !g.hoverOnMap || !g.panelShows(d.side, g.hover) {
		return
	}
	if s.frame.Slim {
		// No steel covers the panel's edge: the cursor stays inside the opening.
		o := s.screen.SlimOpening(d.side)
		dst = dst.SubImage(image.Rect(o.X, o.Y, o.Right(), o.Bottom())).(*ebiten.Image)
	}
	cam := g.splitCamera(d.side)
	x, y := cam.WorldToScreen(hexgrid.Center(g.hover))
	a, ok, held := t.Plate()
	act, st, plate := cursorPlate(d.picking(), d.onTarget(g.hover), held, s.dir == gesture.Centre, a, ok)
	if plate {
		gesture.DrawPlate(dst, x, y, cam.Effective(), act, st)
	} else {
		gesture.DrawRing(dst, x, y, cam.Effective())
	}
}

// darken dims every cell of cells outside the aimed order's reach (terrain
// and units) to darkKeep of its brightness, in one pass so shared edges are
// not dimmed twice.
func (g *game) darken(dst *ebiten.Image, cells []hexgrid.Coord) {
	reach := g.dos.reach()
	if reach == nil {
		return
	}
	var p vector.Path
	for _, c := range cells {
		if reach[c] {
			continue
		}
		for i, v := range hexgrid.Corners(c) {
			x, y := g.cam.WorldToScreen(v[0], v[1])
			if i == 0 {
				p.MoveTo(float32(x), float32(y))
			} else {
				p.LineTo(float32(x), float32(y))
			}
		}
		p.Close()
	}
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.Scale(0, 0, 0, 1-darkKeep)
	vector.FillPath(dst, &p, nil, op)
}

// strategicPic is a panel's strategic map: its layout and its picture, redrawn
// when what it shows changed.
type strategicPic struct {
	lay   minimap.Strategic
	img   *ebiten.Image
	key   strategicKey
	plate color.RGBA // the plate colour, which fills the rest of the panel
}

// strategicKey is what a strategic picture was drawn from.
type strategicKey struct {
	size           image.Point
	version, queue int
	frozen, reveal bool
	pair, row      int
	terrain        *terrainArt
	drawn          bool
}

// openStrategic opens the strategic map in side's panel.
func (g *game) openStrategic(side int) {
	g.split.mode[side] = panelStrategic
	g.split.track.Reset()
	if side == g.dos.side {
		g.dos.Message = "strategic map: pick a place to look at"
	}
}

// strategicGesture handles the human side's input while its strategic map is
// open: a press on the map looks there and puts the cursor there, any other
// press, fire or right click closes it.
func (g *game) strategicGesture(ev gesture.Event) {
	s, side := g.split, g.dos.side
	switch ev.Kind {
	case gesture.Press, gesture.RingOpen:
		if ev.Kind == gesture.Press && ev.Mouse {
			if c, ok := g.strategicPick(side, ev.X, ev.Y); ok {
				g.lookAt(side, c)
				g.setCursor(c)
			}
		}
		s.mode[side] = panelMap
		s.track.Reset()
		g.dos.Message = ""
	default:
	}
}

// strategicRect is the part of side's panel the strategic map is laid out in:
// inside the frame's teeth, half a cell from every edge.
func (g *game) strategicRect(side int) dosscreen.Rect {
	p, sc := g.split.screen.Panels[side], g.split.screen.CellScale
	ix, iy := int(math.Ceil(hexgrid.HexWidth/2*sc)), int(math.Ceil(hexgrid.HexHeight/2*sc))
	return dosscreen.Rect{X: p.X + ix, Y: p.Y + iy, W: max(p.W-2*ix, 0), H: max(p.H-2*iy, 0)}
}

// strategicLayout is side's strategic map laid out in its strategicRect at the
// HUD scale: the picture is magnified by the UI scale, so the frame and the
// plates keep their pixel size.
func (g *game) strategicLayout(side int) minimap.Strategic {
	pic := &g.split.strat[side]
	r, ui := g.strategicRect(side), max(g.view.UI, 1)
	size := image.Pt(r.W/ui, r.H/ui)
	if pic.lay.Size != size || pic.lay.Raster.Grid != g.playGrid() {
		pic.lay = minimap.NewStrategic(g.playGrid(), size)
	}
	return pic.lay
}

// strategicPick is the cell under device pixel (x, y) on side's strategic map.
func (g *game) strategicPick(side int, x, y float64) (hexgrid.Coord, bool) {
	r, ui := g.strategicRect(side), float64(max(g.view.UI, 1))
	wx, wy, ok := g.strategicLayout(side).ToWorld((x-float64(r.X))/ui, (y-float64(r.Y))/ui)
	if !ok {
		return hexgrid.Coord{}, false
	}
	return g.playGrid().PixelToHex(wx, wy)
}

// viewCamera is a camera whose visible world is side's opening block (in the
// slim frame its opening), for the strategic map's view rectangle.
func (g *game) viewCamera(side int) camera.Camera {
	v := g.splitView(side)
	if g.split.frame.Slim {
		o := g.split.screen.SlimOpening(side)
		x0, y0 := v.ScreenToWorld(float64(o.X), float64(o.Y))
		x1, y1 := v.ScreenToWorld(float64(o.Right()), float64(o.Bottom()))
		return camera.Camera{World: g.cam.World, View: camera.Size{W: x1 - x0, H: y1 - y0}, Zoom: 1, Origin: camera.Point{X: x0, Y: y0}}
	}
	ox, oy := v.Origin()
	bw, bh := dosscreen.BlockSize(v.Cols, v.Rows)
	return camera.Camera{World: g.cam.World, View: camera.Size{W: bw, H: bh}, Zoom: 1, Origin: camera.Point{X: ox, Y: oy}}
}

// strategicScene is what side's strategic map shows: the terrain colours, the
// units its map shows as dots and its HQs and depots as plates, in the side
// colours of the minimap (the Central Powers blue), and its panel's view.
func (g *game) strategicScene(side int) minimap.Scene {
	d := g.dos
	cam := g.viewCamera(side)
	sc := minimap.Scene{View: &cam, Terrain: func(c hexgrid.Coord) color.RGBA {
		if !g.playGrid().Contains(c) {
			return color.RGBA{}
		}
		return g.hexColour(c)
	}}
	us := d.unitsOf(side, g.reveal)
	for _, c := range dosplay.Coords(toSet(us)) {
		sc.Units = append(sc.Units, minimap.Unit{At: c, Shade: sideShade(us[c].Owner)})
	}
	bs := d.buildingsOf(side)
	for _, c := range dosplay.Coords(toSet(bs)) {
		b := bs[c]
		kind := minimap.Depot
		switch {
		case b.HQ:
			kind = minimap.HQ
		case b.Kind != dosplay.Depot:
			continue
		}
		sc.Buildings = append(sc.Buildings, minimap.Building{At: c, Kind: kind, Shade: sideShade(b.Owner)})
	}
	return sc
}

// drawStrategic draws side's strategic map into its panel: the plate over the
// whole panel and the picture in strategicRect, magnified by the UI scale
// with nearest filtering; the frame goes over it.
func (g *game) drawStrategic(dst *ebiten.Image, side int) {
	s, d := g.split, g.dos
	pic := &s.strat[side]
	lay := g.strategicLayout(side)
	if lay.Size.X <= 0 || lay.Size.Y <= 0 {
		return
	}
	cam := s.cams[side]
	key := strategicKey{
		size: lay.Size, version: d.s.Version(), queue: len(d.queue), frozen: d.frozenOf(side) != nil,
		reveal: g.reveal, pair: cam.pair, row: cam.row, terrain: g.terrain, drawn: true,
	}
	if pic.img == nil || pic.img.Bounds().Size() != lay.Size {
		if pic.img != nil {
			pic.img.Deallocate()
		}
		pic.img, pic.key = ebiten.NewImage(lay.Size.X, lay.Size.Y), strategicKey{}
	}
	if pic.key != key {
		rgba := lay.Render(g.strategicScene(side))
		pic.img.WritePixels(rgba.Pix)
		pic.key, pic.plate = key, rgba.RGBAAt(0, 0)
	}
	p, r, ui := s.screen.Panels[side], g.strategicRect(side), float64(max(g.view.UI, 1))
	vector.FillRect(dst, float32(p.X), float32(p.Y), float32(p.W), float32(p.H), pic.plate, false)
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(ui, ui)
	op.GeoM.Translate(float64(r.X), float64(r.Y))
	op.Filter = ebiten.FilterNearest
	dst.DrawImage(pic.img, &op)
}

// orderableInView is the first own unit the human side's panel shows that can
// take its phase's order (move or attack).
func (g *game) orderableInView() (hexgrid.Coord, bool) {
	d := g.dos
	for _, c := range dosplay.Coords(toSet(d.own().Units)) {
		if a, ok := d.placement(c).At(gesture.Up); ok && a != gesture.Cancel && g.panelShows(d.side, c) {
			return c, true
		}
	}
	return hexgrid.Coord{}, false
}

// Edge scroll: the pointer resting on the outer half of the human panel's
// outermost cells, or past them on the steel, the status bar or the window
// edge on the panel's side, steps the panel one column pair or row that way,
// first after edgeFirst frames and then every edgeEvery frames once the slide
// has ended. The cursor stays on the cell under the pointer.
const (
	edgeFirst = 12
	edgeEvery = 8
)

// edgeDir is the way a pointer at device pixel (mx, my) pushes view v: -1, 0
// or +1 per axis (column pairs, rows).
func edgeDir(v dosscreen.View, mx, my float64) (dp, dr int) {
	wx, wy := v.ScreenToWorld(mx, my)
	ox, oy := v.Origin()
	rx, ry := wx-ox, wy-oy
	bw, bh := dosscreen.BlockSize(v.Cols, v.Rows)
	const bx, by = hexgrid.HexWidth / 2, hexgrid.HexHeight / 2
	switch {
	case rx < bx:
		dp = -1
	case rx >= bw-bx:
		dp = 1
	}
	switch {
	case ry < by:
		dr = -1
	case ry >= bh-by:
		dr = 1
	}
	return dp, dr
}

// edgeZone reports whether device pixel (mx, my) lies where the pointer
// scrolls the human panel: inside the window, from the panel's outer window
// edge to the middle of the divider, at any height.
func (g *game) edgeZone(mx, my int) bool {
	s := g.split
	if mx < 0 || my < 0 || mx >= s.screen.W || my >= s.screen.H {
		return false
	}
	mid := s.screen.Divider.X + s.screen.Divider.W/2
	if g.dos.side == 0 {
		return mx < mid
	}
	return mx >= mid
}

// edgeScroll runs one frame of the edge scroll for a pointer at (mx, my);
// active tells that the pointer may scroll now (focused window, map mode, no
// menu, ring or held gesture).
func (g *game) edgeScroll(mx, my int, active bool) {
	s, side := g.split, g.dos.side
	dp, dr := 0, 0
	if active && g.edgeZone(mx, my) {
		dp, dr = edgeDir(g.splitView(side), float64(mx), float64(my))
	}
	if dp == 0 && dr == 0 {
		s.edgeOn = false
		return
	}
	if !s.edgeOn {
		s.edgeOn, s.edgeWait = true, edgeFirst
	}
	if s.edgeWait > 0 {
		s.edgeWait--
		return
	}
	if s.cams[side].left > 0 {
		return
	}
	s.edgeWait = edgeEvery
	g.stepView(side, dp, dr)
	g.syncSplitCam()
	if c, ok := g.pickSplit(mx, my); ok {
		g.setCursor(c)
	}
}

// edgeActive reports whether the pointer may edge-scroll the human panel now.
func (g *game) edgeActive(menu bool) bool {
	s := g.split
	if menu || g.dos.over || g.dos.locked() || !ebiten.IsFocused() || s.mode[g.dos.side] != panelMap {
		return false
	}
	if _, ok := s.track.Ring(); ok {
		return false
	}
	_, _, held := s.track.Plate()
	return !held
}
