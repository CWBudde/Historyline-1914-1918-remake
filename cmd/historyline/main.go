// Command historyline is the Go port of the Historyline 1914-1918 remake.
//
// For now it shows a shipped level as flat-colour placeholder hexes with scrolling,
// zoom and hex picking (PLAN.md 2.4). Real graphics follow with the runtime asset
// packs (2.3).
package main

import (
	"flag"
	"fmt"
	"hash/fnv"
	"image/color"
	"log"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/ai"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dospanel"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/minimap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/saves"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scores"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

const (
	startWidth  = 1800
	startHeight = 1000
	scrollSpeed = 12 // default map scroll speed, screen pixels per frame
)

// Pixel outline of a hex around its centre (flat top, 96×80).
var hexOutline = [6][2]float32{{-48, 0}, {-25, -40}, {25, -40}, {48, 0}, {25, 40}, {-25, 40}}

const (
	minNumberZoom = 0.6 // below this the hex numbers would overlap
)

type game struct {
	menuMX, menuMY int       // last cursor position seen by the menu
	menuFocus      int       // focused menu row, menuNone when unset
	menuPress      menuPress // row the mouse button went down on
	lvl            *level.Level
	state          *state.State
	player         state.Player // whose fog is shown
	reveal         bool         // dev: ignore fog
	hexNumbers     bool         // dev: print line,col on every hex
	grid           hexgrid.Grid
	cam            *camera.Camera
	mini           minimap.Minimap
	miniMap        miniMap // the minimap's raster, picture and press
	hover          hexgrid.Coord
	hoverOnMap     bool
	view           view // device and logical screen sizes, UI scale
	w, h           int  // logical HUD size (view.LogW, view.LogH)
	hud            hudLayout
	hudImg         *ebiten.Image // logical HUD canvas, magnified onto the screen
	terrainFill    map[string]color.RGBA
	ctl            *control
	dos            *dosControl // set on the original rules (-rules dos); ctl and state are nil then
	art            *artPacks   // unit sprite packs (-units); nil draws placeholder dots
	screens        *screenArt  // the screen catalog (-screens): portraits, badges, effects; nil shows plain ones
	finale         struct {    // the finale's art and the catalog it was cut from (taking.go)
		from *screenArt
		art  *dospanel.FinaleArt
	}
	scores     *scores.Store      // the best scores; nil keeps none
	artErr     string             // last sprite failure, shown in the status line
	terrain    *terrainArt        // composed terrain tiles (-terrain); nil draws coloured hexes
	terrainErr string             // last terrain tile failure, shown in the status line
	hdUnits    <-chan *artPacks   // full-quality unit art while it loads (-units-hd, without the menus)
	hdTerrain  <-chan *terrainArt // full-quality terrain while it loads (-terrain-hd)
	records    []int              // the terrain records as the sappers' work played left them, nil before (trench.go)
	scroll     float64            // map scroll speed, screen pixels per frame

	split    *split // the split screen of -rules dos, nil otherwise
	drawing  bool   // a dos panel is being drawn: drawSide's map
	drawSide int
	tileTint map[string]color.RGBA // average colour of a terrain tile, by asset ID
	orig     *origArt              // the original tiles and units (F8, origtiles.go); nil until first asked for
}

func newGame(lvl *level.Level, st *state.State, player state.Player, reveal bool) *game {
	gm := newMapGame(lvl, reveal)
	gm.state, gm.player, gm.ctl = st, player, &control{st: st, player: player}
	return gm
}

// newMapGame is the map view of lvl without a rules engine.
func newMapGame(lvl *level.Level, reveal bool) *game {
	g := lvl.Grid()
	ww, wh := g.Bounds()
	world := camera.Size{W: ww, H: wh}
	gm := &game{
		lvl: lvl, reveal: reveal, grid: g,
		cam:         camera.New(world, camera.Size{W: startWidth, H: startHeight}),
		terrainFill: map[string]color.RGBA{},
		scroll:      scrollSpeed,
		menuPress:   newMenuPress(),
	}
	gm.applyLayout(startWidth, startHeight, 1)
	return gm
}

// applyLayout adapts the camera, HUD and minimap to a window of outW × outH
// device-independent pixels at display scale dpi. The camera works in device
// pixels with the UI scale as its display factor, so the map keeps its apparent
// size and the view centre stays put.
func (g *game) applyLayout(outW, outH int, dpi float64) {
	v := newView(outW, outH, dpi)
	if v == g.view {
		return
	}
	g.view = v
	g.w, g.h = v.LogW, v.LogH
	g.hud = layoutHUD(v.LogW, v.LogH)
	g.cam.Resize(camera.Size{W: float64(v.PhysW), H: float64(v.PhysH)})
	g.cam.SetScale(float64(v.UI))
	g.layoutMinimap()
	g.layoutSplit()
}

// cursor returns the mouse position in device pixels (map picking) and in logical
// HUD pixels (menus, minimap).
func (g *game) cursor() (px, py, hx, hy int) {
	px, py = ebiten.CursorPosition()
	hx, hy = g.view.toHUD(px, py)
	return px, py, hx, hy
}

// fogOf is what the shown player knows about c; everything is visible in reveal mode.
func (g *game) fogOf(c hexgrid.Coord) state.Fog {
	if g.reveal || g.dos != nil {
		return state.Visible
	}
	return g.state.Fog(g.player, c)
}

// fogShade applies what the player knows to a terrain colour. Unseen hexes show
// nothing of the terrain and remembered ones are dimmed and grey.
func fogShade(f state.Fog, c color.RGBA) color.RGBA {
	switch f {
	case state.Visible:
		return c
	case state.Seen:
		grey := uint8((int(c.R) + int(c.G) + int(c.B)) / 3 / 2)
		return color.RGBA{R: grey, G: grey, B: grey + 8, A: 255}
	}
	return color.RGBA{R: 8, G: 8, B: 10, A: 255}
}

// walking reports whether the unit standing on at is still being animated
// there. A building or carrier the unit walks into stays drawn.
func (g *game) walking(at hexgrid.Coord) bool {
	if g.dos != nil {
		return g.dosWalking(at)
	}
	w := g.drv().ui().walk
	return !w.Done() && w.path[len(w.path)-1] == at && g.state.At(at).Unit == w.unit
}

// showsUnit reports whether live units on a hex of the given fog state are drawn.
func showsUnit(f state.Fog) bool { return f == state.Visible }

// unitShown reports whether a unit is drawn on c and whether it is only the
// player's last-seen snapshot (it may have gone since). Live units show only on
// Visible hexes; remembered own units are skipped because the player always
// knows where those are.
func (g *game) unitShown(c hexgrid.Coord) (shown, remembered bool) {
	if g.dos != nil {
		_, ok := g.dos.unitsOf(g.shownSide(), g.reveal)[c]
		return ok && !g.walking(c), false
	}
	f := g.fogOf(c)
	if showsUnit(f) {
		return g.state.At(c).Unit != nil && !g.walking(c), false
	}
	if f != state.Seen {
		return false, false
	}
	m, ok := g.state.Remembered(g.player, c)
	return ok && m.Unit != nil && m.Unit.Owner != g.player, true
}

// drawUnit draws the unit shown on c at screen point (sx, sy): the live unit, or
// the player's last-seen snapshot greyed. Without unit packs it is a dot.
func (g *game) drawUnit(dst *ebiten.Image, c hexgrid.Coord, sx, sy float64, remembered bool) {
	if g.dos != nil {
		g.drawDosUnit(dst, c, sx, sy)
		return
	}
	col := color.RGBA{230, 230, 230, 255}
	if remembered {
		col = rememberedUnit
		if m, _ := g.state.Remembered(g.player, c); m.Unit != nil &&
			g.drawUnitSprite(dst, c, sx, sy, m.Unit.Type, m.Unit.Owner, false, true, unitsprite.N) {
			return
		}
	} else if u := g.state.At(c).Unit; g.drawUnitSprite(dst, c, sx, sy, u.Type, u.Owner, unitSpent(u), false, g.drv().ui().Facing(u)) {
		return
	}
	vector.FillCircle(dst, float32(sx), float32(sy), float32(14*g.cam.Effective()), col, true)
}

// ownerShown is the building owner the player knows for c: live when Visible,
// last seen when Seen, None when Unseen or not a building.
func (g *game) ownerShown(c hexgrid.Coord) state.Player {
	if g.dos != nil {
		return g.dos.buildingsOf(g.shownSide())[c].Owner // None when no building
	}
	if g.fogOf(c) == state.Visible {
		if h := g.state.At(c); h.Container != nil {
			return h.Owner
		}
		return state.None
	}
	m, _ := g.state.Remembered(g.player, c)
	return m.Owner
}

// ownerColour marks the two sides' buildings and unit plates in their
// minimap colours, as the original does: the German side blue, the Entente
// yellow.
func ownerColour(p state.Player) color.RGBA {
	return sideShade(p).Lit
}

// rememberedUnit is the colour of a last-seen unit (main view and minimap).
var rememberedUnit = color.RGBA{120, 120, 128, 255}

func (g *game) fill(terrain string) color.RGBA {
	if c, ok := g.terrainFill[terrain]; ok {
		return c
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(terrain))
	v := hash.Sum32()
	c := color.RGBA{R: 70 + uint8(v>>16)%120, G: 90 + uint8(v>>8)%120, B: 60 + uint8(v)%100, A: 255}
	g.terrainFill[terrain] = c
	return c
}

func (g *game) Update() error {
	if u, ok := arrived(&g.hdUnits); ok {
		g.art.release()
		g.art = u
	}
	if t, ok := arrived(&g.hdTerrain); ok {
		g.terrain.release()
		g.terrain = t
		if g.records != nil {
			g.reworkTerrain()
		}
	}
	if g.split == nil {
		g.scrollKeys()
	}
	anyInput := len(inpututil.AppendJustPressedKeys(nil)) > 0 || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if ui := g.drv().ui(); ui.report.Active() && anyInput {
		ui.DismissReport()
	}
	if g.dos != nil && g.dos.scene() && anyInput {
		g.dos.SkipScene() // as the battle panel, any key or click skips a scene
	}
	g.drv().Tick()
	if g.dos != nil && g.dos.rework {
		g.dos.rework = false
		g.reworkTerrain()
	}
	if g.split != nil {
		g.updateSplit()
	}
	pressed := inpututil.IsKeyJustPressed
	if g.split != nil {
		// Space is the fire key of the cursor (key set A).
		pressed = func(k ebiten.Key) bool { return k != ebiten.KeySpace && inpututil.IsKeyJustPressed(k) }
	}
	for _, a := range actionsFor(pressed) {
		if !g.applyAction(a) {
			return ebiten.Termination
		}
	}
	if g.split != nil {
		// Under -rules dos the wheel steps the zoom levels and the cursor
		// takes the mouse and the keys; the menu keeps the keys while it is
		// open.
		menu := g.dos.MenuOpen()
		if _, wheel := ebiten.Wheel(); !menu {
			mx, my := ebiten.CursorPosition()
			g.wheelSplit(wheel, mx, my)
		}
		g.updateMenu()
		g.updateCursor(menu || g.dos.MenuOpen())
		return nil
	}
	showPointer(true)
	mx, my, hx, hy := g.cursor()
	// The press rule sees every frame, so a press that starts under a menu or
	// on the map never drives the minimap.
	onMini := g.mini.Contains(float64(hx), float64(hy))
	drive := g.miniMap.press.Update(ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft), onMini)
	if g.updateMenu() {
		return nil
	}
	if _, wheel := ebiten.Wheel(); wheel != 0 {
		factor := 1.1
		if wheel < 0 {
			factor = 1 / 1.1
		}
		g.cam.ZoomAt(factor, float64(mx), float64(my))
	}
	if drive {
		g.cam.CenterOn(g.mini.ToWorld(g.mini.Clamp(float64(hx), float64(hy))))
	}
	if onMini || g.miniMap.press.Captured() {
		g.hoverOnMap = false
		return nil
	}
	wx, wy := g.cam.ScreenToWorld(float64(mx), float64(my))
	g.hover, g.hoverOnMap = g.grid.PixelToHex(wx, wy)
	if g.hoverOnMap && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.drv().Click(g.hover)
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		g.drv().Clear()
	}
	return nil
}

// scrollKeys scrolls the single map with the arrow keys and WASD.
func (g *game) scrollKeys() {
	dx, dy := 0.0, 0.0
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		dx -= g.scroll
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		dx += g.scroll
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		dy -= g.scroll
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		dy += g.scroll
	}
	g.cam.Scroll(dx*float64(g.view.UI), dy*float64(g.view.UI))
}

// menuRow is the HUD rectangle of menu row i: beside the selected unit for
// the action menu, at the bottom left for the weapon and contents menus.
func (g *game) menuRow(i int) (x, y, w, h int) {
	ui := g.drv().ui()
	if ui.menuKind != menuActions {
		r := g.hud.menuRowRect(i, len(ui.menu))
		return r.X, r.Y, r.W, r.H
	}
	r := g.actionMenu()
	return r.X, r.Y + i*menuRowHeight, r.W, menuRowHeight
}

// actionMenu is the HUD rectangle of the action menu beside the selected unit.
func (g *game) actionMenu() rect {
	ui := g.drv().ui()
	mw := 0
	for i, o := range ui.menu {
		mw = max(mw, textWidth(menuLabel(i, o))+8)
	}
	px, py := g.cam.WorldToScreen(hexgrid.Center(ui.from))
	cx, cy := g.view.toHUD(int(px), int(py))
	half := int(float64(hexgrid.HexWidth) / 2 * g.cam.Effective() / float64(max(g.view.UI, 1)))
	return actionMenuRect(cx, cy, half, mw, len(ui.menu), g.w, g.h)
}

// updateMenu handles the open menu (arrow keys, Enter or Space, number keys, or
// a press and release on a row) and reports whether it consumed this frame's
// input.
func (g *game) updateMenu() bool {
	ui := g.drv().ui()
	if !ui.MenuOpen() {
		g.menuFocus = menuNone
		g.menuPress = newMenuPress()
		return false
	}
	g.menuFocus = menuFocusValid(g.menuFocus, ui.menu)
	_, _, mx, my := g.cursor()
	if mx != g.menuMX || my != g.menuMY {
		g.menuMX, g.menuMY = mx, my
		if h := menuHit(mx, my, len(ui.menu), g.menuRow); h != menuNone {
			g.menuFocus = h
		}
	}
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowDown):
		g.menuFocus = menuStep(g.menuFocus, 1, ui.menu)
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowUp):
		g.menuFocus = menuStep(g.menuFocus, -1, ui.menu)
	case inpututil.IsKeyJustPressed(ebiten.KeyEnter), inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter),
		inpututil.IsKeyJustPressed(ebiten.KeySpace):
		if g.menuFocus != menuNone {
			g.drv().ChooseMenu(g.menuFocus)
			return true
		}
	}
	for i, k := range []ebiten.Key{ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4, ebiten.Key5, ebiten.Key6, ebiten.Key7, ebiten.Key8, ebiten.Key9} {
		if inpututil.IsKeyJustPressed(k) && i < len(ui.menu) {
			g.drv().ChooseMenu(i)
			return true
		}
	}
	row, used := g.menuPress.update(inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft),
		inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft),
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft),
		menuHit(mx, my, len(ui.menu), g.menuRow))
	if row != menuNone {
		g.drv().ChooseMenu(row)
	}
	return used
}

func (g *game) drawMenu(dst *ebiten.Image) {
	ui := g.drv().ui()
	for i, o := range ui.menu {
		x, y, w, h := g.menuRow(i)
		bg := color.RGBA{30, 60, 30, 230}
		if !o.Usable {
			bg = color.RGBA{50, 40, 40, 200}
		}
		if i == g.menuFocus {
			bg = color.RGBA{70, 110, 60, 245}
			if !o.Usable {
				bg = color.RGBA{100, 50, 50, 245}
			}
		}
		pressed := i == g.menuPress.Row() && menuHit(g.menuMX, g.menuMY, len(ui.menu), g.menuRow) == i
		if pressed {
			bg = color.RGBA{bg.R / 2, bg.G / 2, bg.B / 2, 250}
		}
		vector.FillRect(dst, float32(x), float32(y), float32(w), float32(h), bg, false)
		if i == g.menuFocus {
			vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 2, color.RGBA{240, 220, 120, 255}, false)
		}
		dx := 0
		if pressed {
			dx = 1 // the label sinks with the row
		}
		ebitenutil.DebugPrintAt(dst, menuLabel(i, o), x+4+dx, y+1+dx)
	}
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 24, 255})
	vx, vy, vw, vh := g.cam.Visible()
	if g.split != nil {
		g.drawSplit(screen)
		g.drawWells(screen)
		g.drawFinale(screen)
	} else {
		g.drawMapLayer(screen, vx, vy, vw, vh, true)
	}

	// Everything below is HUD, drawn in logical pixels and magnified by the UI scale.
	if g.hudImg == nil || g.hudImg.Bounds().Dx() != g.w || g.hudImg.Bounds().Dy() != g.h {
		g.hudImg = ebiten.NewImage(g.w, g.h)
	}
	hud := g.hudImg
	hud.Clear()
	switch {
	case g.hexNumbers && g.split != nil:
		g.drawSplitHexNumbers(hud)
	case g.hexNumbers && g.cam.Zoom >= minNumberZoom:
		g.drawHexNumbers(hud, vx, vy, vw, vh)
	}
	g.drawOrderNumbers(hud)
	if g.split == nil {
		g.drawMinimap(hud)
	}
	g.drawMenu(hud)
	g.drawReport(hud)
	g.drawBanner(hud)
	g.drawInfo(hud)
	if dev := g.devIndicators(); dev != "" {
		ebitenutil.DebugPrintAt(hud, dev, g.hud.DevRight-textWidth(dev), 0)
	}
	g.drawStatus(hud)
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(g.view.UI), float64(g.view.UI))
	screen.DrawImage(hud, &op)
}

// drawMapLayer draws the map in the world rectangle (vx, vy, vw, vh) through
// g.cam on dst, in device pixels. The selection, orders, walker and hover
// outline belong to the human side's map only (human). The edge of an original
// map is not drawn (playGrid), so its standing items cannot reach into view.
func (g *game) drawMapLayer(dst *ebiten.Image, vx, vy, vw, vh float64, human bool) {
	// A hex just outside the view counts while its standing items may reach in.
	const mx, my = hexgrid.HexWidth/2 + mapcompose.OverhangX, hexgrid.HexHeight/2 + mapcompose.OverhangY
	play := g.playGrid()
	var inView []hexgrid.Coord
	for l, line := range g.lvl.Lines {
		for c := range line {
			coord := hexgrid.Coord{Line: l, Col: c}
			cx, cy := hexgrid.Center(coord)
			if !play.Contains(coord) || cx+mx < vx || cx-mx > vx+vw || cy+my < vy || cy-my > vy+vh {
				continue
			}
			inView = append(inView, coord)
		}
	}
	// Terrain, then owner marks, then units and their markers, each in its own pass so that a later
	// hex never covers what an earlier one drew beyond its outline.
	switch {
	case g.drawOrigTiles(dst, inView):
	case g.terrain != nil:
		g.drawTerrain(dst, inView)
	default:
		for _, coord := range inView {
			sx, sy := g.cam.WorldToScreen(hexgrid.Center(coord))
			hx := g.lvl.Lines[coord.Line][coord.Col]
			g.drawHex(dst, float32(sx), float32(sy), fogShade(g.fogOf(coord), g.fill(hx.Terrain)), false)
		}
	}
	for _, coord := range inView {
		if owner := g.ownerShown(coord); owner != state.None {
			sx, sy := g.cam.WorldToScreen(hexgrid.Center(coord))
			e := float32(g.cam.Effective())
			vector.FillRect(dst, float32(sx)-20*e, float32(sy)-34*e, 40*e, 6*e, fogShade(g.fogOf(coord), ownerColour(owner)), false)
		}
	}
	for _, coord := range inView {
		if shown, remembered := g.unitShown(coord); shown {
			sx, sy := g.cam.WorldToScreen(hexgrid.Center(coord))
			g.drawUnit(dst, coord, sx, sy, remembered)
		}
	}
	g.drawUnitMarks(dst, inView)
	if w := g.walkShown(human); !w.Done() && g.fogOf(w.Hex()) == state.Visible {
		wx, wy := w.Pos()
		sx, sy := g.cam.WorldToScreen(wx, wy)
		face, _ := unitsprite.Toward(w.Heading())
		if !g.drawWalker(dst, w, w.Hex(), sx, sy, face) {
			vector.FillCircle(dst, float32(sx), float32(sy), float32(14*g.cam.Effective()), color.RGBA{230, 230, 230, 255}, true)
		}
	}
	if human {
		if g.dos != nil {
			g.darken(dst, inView) // the cursor is drawn over the frame (drawCursor)
		} else {
			g.drawOverlays(dst)
		}
		g.drawOrders(dst)
		if g.hoverOnMap && g.dos == nil {
			cx, cy := hexgrid.Center(g.hover)
			sx, sy := g.cam.WorldToScreen(cx, cy)
			g.drawHex(dst, float32(sx), float32(sy), color.RGBA{}, true)
		}
	}
	// Hit marks show on both sides' maps: both see the battle.
	g.drawMarks(dst)
}

// walkShown is the walk drawn on the map being drawn: under -rules dos the
// shown side's own, else the walk on the human side's map only.
func (g *game) walkShown(human bool) *walk {
	switch {
	case g.dos != nil:
		return g.dos.walkOf(g.shownSide())
	case human:
		return g.drv().ui().walk
	}
	return nil
}

// drawStatus prints the status and help text at the top left of the HUD.
func (g *game) drawStatus(dst *ebiten.Image) { ebitenutil.DebugPrint(dst, g.statusText()) }

// statusText is the status and help text. It names nothing of the hovered hex:
// the info box shows that, fog-safe. Under -rules dos the status wells carry the
// state and the message, and only art failures stay here.
func (g *game) statusText() string {
	if g.dos != nil {
		return g.splitErrors()
	}
	msg := fmt.Sprintf("%s  round %d, %v to move  zoom %.2f (%dx assets)\narrows/WASD scroll, wheel zoom, click unit for its menu or click a blue/red hex (1-9 picks menu row; click own building for contents; click skips a walk), F1 end turn, F11 fullscreen, Esc/right-click cancel",
		g.lvl.MapName, g.state.Round, g.state.Turn, g.cam.Zoom, g.drawnDensity())
	if g.state.Winner != state.None {
		msg += fmt.Sprintf("\n%v has won", g.state.Winner)
	}
	if m := g.drv().ui().Message; m != "" {
		msg += "\n" + m
	}
	if g.artErr != "" {
		msg += "\nunit sprites: " + g.artErr
	}
	if g.terrainErr != "" {
		msg += "\nterrain tiles: " + g.terrainErr
	}
	return msg
}

// drawnDensity is the asset density the map is drawn with: the terrain density
// once terrain packs are loaded, else the sprite density once unit packs are
// (no 1x unit art exists), the requested density otherwise.
func (g *game) drawnDensity() int {
	if g.terrain != nil {
		return terrainDensity(g.cam.Effective())
	}
	if g.art != nil {
		return spriteDensity(g.cam.Effective())
	}
	return assetDensity(g.cam.Effective())
}

// devIndicators names the active development flags so a screenshot can never be
// mistaken for normal play.
func (g *game) devIndicators() string {
	var on []string
	if g.reveal {
		on = append(on, "REVEAL")
	}
	if g.hexNumbers {
		on = append(on, "HEXNUMBERS")
	}
	if g.art != nil && g.art.candidate || g.terrain != nil && g.terrain.candidate {
		on = append(on, "CANDIDATE-ART")
	}
	if g.terrain != nil {
		if n := len(g.terrain.scene.Gaps); n > 0 {
			on = append(on, fmt.Sprintf("CONTEXT-GAP %d", n))
		}
		if n := len(g.terrain.scene.Issues); n > 0 {
			on = append(on, fmt.Sprintf("COMPOUND-ISSUE %d", n))
		}
	}
	if len(on) == 0 {
		return ""
	}
	return "DEV: " + strings.Join(on, " ")
}

// drawHexNumbers prints "line,col" under each visible hex centre (dev flag).
func (g *game) drawHexNumbers(dst *ebiten.Image, vx, vy, vw, vh float64) {
	for l, line := range g.lvl.Lines {
		for c := range line {
			cx, cy := hexgrid.Center(hexgrid.Coord{Line: l, Col: c})
			if cx+48 < vx || cx-48 > vx+vw || cy+40 < vy || cy-40 > vy+vh {
				continue
			}
			sx, sy := g.cam.WorldToScreen(cx, cy)
			hx, hy := g.view.toHUD(int(sx), int(sy+22*g.cam.Effective()))
			ebitenutil.DebugPrintAt(dst, fmt.Sprintf("%d,%d", l, c), hx-14, hy-8)
		}
	}
}

// drawOverlays tints the hexes the selected unit may reach (blue) or attack
// (red) and outlines the selection, all taken from the rules engine's queries
// (the remake rules; -rules dos darkens what is out of reach instead).
func (g *game) drawOverlays(dst *ebiten.Image) {
	ui := g.drv().ui()
	if !ui.selected {
		return
	}
	tint := func(at hexgrid.Coord, col color.RGBA) {
		cx, cy := hexgrid.Center(at)
		sx, sy := g.cam.WorldToScreen(cx, cy)
		g.drawHex(dst, float32(sx), float32(sy), col, false)
	}
	for at := range g.ctl.moves {
		tint(at, color.RGBA{60, 120, 255, 90})
	}
	for at := range ui.targets {
		tint(at, color.RGBA{255, 50, 40, 130})
	}
	cx, cy := hexgrid.Center(ui.from)
	sx, sy := g.cam.WorldToScreen(cx, cy)
	g.drawHex(dst, float32(sx), float32(sy), color.RGBA{}, true)
}

func (g *game) drawHex(dst *ebiten.Image, cx, cy float32, fill color.RGBA, outline bool) {
	z, ui := float32(g.cam.Effective()), float32(g.view.UI)
	var p vector.Path
	for i, v := range hexOutline {
		x, y := cx+v[0]*z, cy+v[1]*z
		if i == 0 {
			p.MoveTo(x, y)
		} else {
			p.LineTo(x, y)
		}
	}
	p.Close()
	if outline {
		vector.StrokePath(dst, &p, &vector.StrokeOptions{Width: 3 * ui}, pathOptions(color.RGBA{255, 240, 80, 255}))
		return
	}
	vector.FillPath(dst, &p, &vector.FillOptions{}, pathOptions(fill))
	vector.StrokePath(dst, &p, &vector.StrokeOptions{Width: ui}, pathOptions(color.RGBA{0, 0, 0, 90}))
}

func pathOptions(c color.RGBA) *vector.DrawPathOptions {
	var o vector.DrawPathOptions
	o.AntiAlias = true
	o.ColorScale.ScaleWithColor(c)
	return &o
}

// Layout returns the screen in device pixels, so the map is drawn at full
// resolution; the HUD is magnified separately (see view). The split screen of
// -rules dos keeps the window large enough for its smallest panels.
func (g *game) Layout(outsideW, outsideH int) (int, int) {
	dpi := ebiten.Monitor().DeviceScaleFactor()
	g.applyLayout(outsideW, outsideH, dpi)
	if g.split != nil {
		limitWindow(splitMinWindow(dpi))
	} else {
		limitWindow(0, 0)
	}
	return g.view.PhysW, g.view.PhysH
}

// loadLevel reads name from dir, falling back to the embedded shipped levels when
// the file is missing (always the case in the browser).
func loadLevel(dir, name string) (*level.Level, error) {
	lvl, err := level.Load(filepath.Join(dir, name+".json"))
	if err == nil {
		return lvl, nil
	}
	raw, embErr := levels.FS.ReadFile(name + ".json")
	if embErr != nil {
		return nil, err
	}
	return level.Parse(raw)
}

// matchConfig is everything a battle is started from: by the menu, or directly
// by -level or -player.
type matchConfig struct {
	lvl        *level.Level
	board      *dosbind.Board // an original map's exact binding (loadMap); nil binds lvl
	set        *data.Set
	player     state.Player // the human side, whose view is shown
	rules      string       // dos (default, original) or standard (the remake rules)
	opponent   string       // ai or pass
	limit      int          // dos: orders per phase, 0 for no limit
	reveal     bool
	hexNumbers bool
	reduced    bool
	scroll     float64       // map scroll speed, screen pixels per frame
	units      *artPacks     // shared by every battle; nil draws placeholder dots
	screens    *screenArt    // the screen catalog for the portraits; nil shows blank photos
	scores     *scores.Store // the best scores, opened once; nil keeps none
	saves      *saves.Store  // the single saved battle; nil keeps none
	saved      []byte        // a saved battle to resume (dos only); nil starts fresh
	terrain    string        // terrain catalog directory, composed per level; "" draws coloured hexes
	terrainHD  string        // full-quality terrain catalog loaded in the background of a battle; "" for none
	candidates bool          // allow candidate terrain packs
}

// newMatch starts the battle of c, counting the art it fetches in m.
func newMatch(c matchConfig, m *meter) (*game, error) {
	var gm *game
	switch c.rules {
	case "dos":
		o := dosplay.Options{Limit: c.limit}
		var planner dosPlanner = passPlanner{}
		if c.opponent == "ai" {
			o.Computer[1-dosbind.Side(c.player)] = true
			planner = computerPlanner()
		}
		var s *dosplay.Session
		var err error
		switch {
		case c.saved != nil && c.board != nil:
			s, err = dosplay.LoadBoard(c.board, c.lvl, c.set, c.saved)
		case c.saved != nil:
			s, err = dosplay.Load(c.lvl, c.set, c.saved)
		case c.board != nil:
			s, err = dosplay.NewBoard(c.board, c.lvl, c.set, o)
		default:
			s, err = dosplay.New(c.lvl, c.set, o)
		}
		if err != nil {
			return nil, err
		}
		if c.saved != nil {
			// The save carries Computer flags and the limit; the planner
			// follows the other side's flag.
			planner = passPlanner{}
			if s.Game.Players[1-s.SideOf(c.player)].Computer {
				planner = computerPlanner()
			}
		}
		gm = newDosGame(c.lvl, s, s.SideOf(c.player), planner, c.reveal)
	case "standard":
		st, err := state.New(c.lvl, c.set)
		if err != nil {
			return nil, err
		}
		gm = newGame(c.lvl, st, c.player, c.reveal)
		if c.opponent == "ai" {
			gm.ctl.ai = &ai.Planner{}
		}
	default:
		return nil, fmt.Errorf("unknown -rules %q (want dos or standard)", c.rules)
	}
	gm.hexNumbers = c.hexNumbers
	gm.drv().SetReducedMotion(c.reduced)
	if c.scroll > 0 {
		gm.scroll = c.scroll
	}
	gm.art, gm.screens, gm.scores = c.units, c.screens, c.scores
	if c.terrain != "" {
		var err error
		if gm.terrain, err = loadTerrainArt(c.terrain, c.candidates, c.lvl, m); err != nil {
			return nil, err
		}
		if c.terrainHD != "" {
			gm.hdTerrain = background("full-quality terrain", func() (*terrainArt, error) {
				return loadTerrainArt(c.terrainHD, c.candidates, c.lvl, &meter{pause: pagePauser()})
			})
		}
	}
	return gm, nil
}

func main() {
	name := flag.String("level", "level1_cp", "shipped level to load from levels/ (name without .json), or original/NN for an imported original map (its campaign level); skips the menu")
	dir := flag.String("levels", "levels", "directory with the level files")
	side := flag.String("player", "", "side the player plays and whose view is shown: central, entente (default: the level's side); skips the menu")
	rules := flag.String("rules", "dos", "rule set: dos (default: the original order cycle) or standard (the remake rules)")
	reveal := flag.Bool("reveal", false, "dev: show the whole map, ignoring fog of war (dos: every unit where its own side sees it)")
	hexNumbers := flag.Bool("hexnumbers", false, "dev: print line,col on every hex")
	reduced := flag.Bool("reduced-motion", false, "walk instantly and show short static effects (toggle in game with M)")
	units := flag.String("units", "", "unit sprite catalog built by hlasset build-units (default: placeholder dots)")
	terrain := flag.String("terrain", "", "terrain tile catalog built by hlasset build-terrain (default: coloured hexes)")
	screens := flag.String("screens", "", "screen catalog built by hlasset build-screens (default: plain menu panels)")
	opponent := flag.String("opponent", "ai", "who plays the other side: ai (computer opponent) or pass (passes every turn)")
	unitsHD := flag.String("units-hd", "", "web: full-quality unit catalog, loaded in the background when the -units catalog downloaded fast")
	terrainHD := flag.String("terrain-hd", "", "web: full-quality terrain catalog, loaded in the background of a battle when the -units catalog downloaded fast")
	screensHD := flag.String("screens-hd", "", "web: full-quality screen catalog, loaded in the background when the -units catalog downloaded fast")
	candidates := flag.Bool("allow-candidates", false, "allow unit, terrain and screen packs that are review candidates, not approved art")
	capture := flag.String("capture", "", "dev: render nine 720p/1080p/4K review frames to a directory under build/ and exit (requires -units and -terrain)")
	captureLook := flag.Bool("capture-look", false, "with -capture: render revealed/fog views at exact 0.5x, 1x and 1.6x effective zoom at 1080p/4K")
	flag.Parse()
	if *captureLook && *capture == "" {
		log.Fatal("-capture-look requires -capture")
	}
	direct := *capture != ""
	flag.Visit(func(f *flag.Flag) { direct = direct || f.Name == "level" || f.Name == "player" })

	set, err := data.Load()
	if err != nil {
		log.Fatal(err)
	}
	if *opponent != "ai" && *opponent != "pass" {
		log.Fatalf("unknown -opponent %q (want ai or pass)", *opponent)
	}
	if *rules != "dos" && *rules != "standard" {
		log.Fatalf("unknown -rules %q (want dos or standard)", *rules)
	}
	c := matchConfig{set: set, rules: *rules, opponent: *opponent, reveal: *reveal, hexNumbers: *hexNumbers,
		reduced: *reduced, terrain: *terrain, candidates: *candidates, scores: scores.Open(), saves: saves.Open()}
	// The web page shows how much of the art has arrived until the game starts.
	boot := &meter{fetched: pageFetched}
	if *units != "" {
		if c.units, err = loadUnitArt(*units, *candidates, boot); err != nil {
			log.Fatal(err)
		}
	}
	// Only a fast connection loads the full-quality art as well.
	hd := hdDirs{units: *unitsHD, screens: *screensHD, terrain: *terrainHD}
	if !boot.fast() {
		hd = hdDirs{}
	}
	c.terrainHD = hd.terrain
	ebiten.SetWindowSize(startWidth, startHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Historyline 1914-1918")
	if *screens != "" {
		if c.screens, err = loadScreenArt(*screens, *candidates, boot); err != nil {
			log.Fatal(err)
		}
		dospanel.SetSurfaces(c.screens.surfaces())
		dosscreen.SetArt(c.screens.frameArt())
	}

	if !direct {
		campaign, err := loadCampaign(*dir, set)
		if err != nil {
			log.Fatal(err)
		}
		pageReady()
		a := newApp(c, campaign, c.screens)
		a.hd = loadHDArt(hd, *candidates)
		if err := ebiten.RunGame(a); err != nil {
			log.Fatal(err)
		}
		return
	}

	if c.lvl, c.board, err = loadMap(*dir, *name, set); err != nil {
		log.Fatal(err)
	}
	if c.player, err = state.PlayerByName(c.lvl.Player); err != nil {
		log.Fatal(err)
	}
	switch *side {
	case "":
	case "central":
		c.player = state.CentralPowers
	case "entente":
		c.player = state.Entente
	default:
		log.Fatalf("unknown -player %q (want central or entente)", *side)
	}
	gm, err := newMatch(c, boot)
	if err != nil {
		log.Fatal(err)
	}
	if *capture != "" {
		if gm.art == nil || gm.terrain == nil {
			log.Fatal("-capture requires both -units and -terrain")
		}
		if err := writeReviewFrames(gm, *capture, *captureLook); err != nil {
			log.Fatal(err)
		}
		return
	}
	if hd.units != "" {
		gm.hdUnits = background("full-quality units", func() (*artPacks, error) {
			return loadUnitArt(hd.units, *candidates, &meter{pause: pagePauser()})
		})
	}
	pageReady()
	if err := ebiten.RunGame(gm); err != nil {
		log.Fatal(err)
	}
}

// drawReport shows the battle panel at the top centre while it is active.
func (g *game) drawReport(dst *ebiten.Image) {
	r := g.drv().ui().report
	if !r.Active() {
		return
	}
	p := g.hud.Report

	vector.FillRect(dst, float32(p.X), float32(p.Y), float32(p.W), float32(p.H), color.RGBA{20, 20, 40, 235}, false)
	for i, l := range r.Lines() {
		ebitenutil.DebugPrintAt(dst, l, p.X+6, p.Y+4+i*32)
	}
}
