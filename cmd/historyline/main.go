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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dospanel"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/faction"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/llmplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/minimap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/saves"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scores"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
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
	lvl         *level.Level
	reveal      bool // dev: ignore fog
	hexNumbers  bool // dev: print line,col on every hex
	grid        hexgrid.Grid
	cam         *camera.Camera
	mini        minimap.Minimap
	miniMap     miniMap // the minimap's raster, picture and press
	hover       hexgrid.Coord
	hoverOnMap  bool
	view        view // device and logical screen sizes, UI scale
	w, h        int  // logical HUD size (view.LogW, view.LogH)
	hud         hudLayout
	hudImg      *ebiten.Image // logical HUD canvas, magnified onto the screen
	terrainFill map[string]color.RGBA
	dos         *dosControl // the original rules' controller
	gasTick     int         // frames drawn, for the drift of the gas cloud (gas.go)
	art         *artPacks   // unit sprite packs; nil only in tests, which draw no units
	screens     *screenArt  // the screen catalog (-screens): portraits, badges, effects; nil shows plain ones
	finale      struct {    // the finale's art and the catalog it was cut from (taking.go)
		from *screenArt
		art  *dospanel.FinaleArt
	}
	surrender struct { // the surrender scene's art and its catalog (taking.go)
		from *screenArt
		art  *dospanel.SurrenderArt
	}
	scores         *scores.Store          // the best scores; nil keeps none
	artErr         string                 // last sprite failure, shown in the status line
	terrain        *terrainArt            // composed terrain tiles (-terrain); nil draws coloured hexes
	terrainErr     string                 // last terrain tile failure, shown in the status line
	records        []int                  // the terrain records as the sappers' work played left them, nil before (trench.go)
	backdrop       *backdrop              // the painted ground drawn in place of the terrain (-backdrop); nil draws the terrain
	backdropHidden *backdrop              // the backdrop while F8 shows another map design (design.go); nil when it is shown or there is none
	backdropView   *ebiten.Image          // the view of the backdrop the seen hexes are greyed from
	reworkedHex    map[hexgrid.Coord]bool // the hexes the sappers changed (trench.go)
	scroll         float64                // map scroll speed, screen pixels per frame

	split    *split // the split screen; nil before the first layout and in map-only captures
	mapOnly  bool   // draw the whole map without the split screen (-capture-map)
	drawing  bool   // a dos panel is being drawn: drawSide's map
	drawSide int
	tileTint map[string]color.RGBA // average colour of a terrain tile, by asset ID
	orig     *origArt              // the original tiles and units (F8, origtiles.go); nil until first asked for
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

// fogOf is what the shown player knows about c: under the original rules
// every hex is visible, and the units decide what is shown.
func (g *game) fogOf(hexgrid.Coord) fogState { return fogVisible }

// fogShade applies what the player knows to a terrain colour. Unseen hexes show
// nothing of the terrain and remembered ones are dimmed and grey.
func fogShade(f fogState, c color.RGBA) color.RGBA {
	switch f {
	case fogVisible:
		return c
	case fogSeen:
		grey := uint8((int(c.R) + int(c.G) + int(c.B)) / 3 / 2)
		return color.RGBA{R: grey, G: grey, B: grey + 8, A: 255}
	}
	return color.RGBA{R: 8, G: 8, B: 10, A: 255}
}

// walking reports whether the unit standing on at is still being animated
// there. A building or carrier the unit walks into stays drawn.
func (g *game) walking(at hexgrid.Coord) bool {
	return g.dos != nil && g.dosWalking(at)
}

// unitShown reports whether a unit is drawn on c: one the shown side sees
// that is not walking.
func (g *game) unitShown(c hexgrid.Coord) bool {
	if g.dos == nil {
		return false
	}
	_, ok := g.dos.unitsOf(g.shownSide(), g.reveal)[c]
	return ok && !g.walking(c)
}

// ownerShown is the building owner the shown side knows for c, None when
// not a building.
func (g *game) ownerShown(c hexgrid.Coord) faction.Player {
	if g.dos == nil {
		return faction.None
	}
	return g.dos.buildingsOf(g.shownSide())[c].Owner // None when no building
}

// ownerColour marks the two sides' buildings and unit plates in their
// minimap colours, as the original does: the German side blue, the Entente
// yellow.
func ownerColour(p faction.Player) color.RGBA {
	return sideShade(p).Lit
}

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

// advance moves the game on a frame without reading input: the driver's
// tick, a terrain rework it asked for, the split screen's panels. Update
// and the recorder (-record) both advance by it.
func (g *game) advance() {
	g.drv().Tick()
	if g.dos != nil && g.dos.rework {
		g.dos.rework = false
		g.reworkTerrain()
	}
	if g.split != nil {
		g.updateSplit()
	}
}

func (g *game) Update() error {
	g.gasTick++
	if g.split == nil {
		g.scrollKeys()
	}
	anyInput := len(inpututil.AppendJustPressedKeys(nil)) > 0 || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if g.dos != nil && g.dos.scene() && anyInput {
		g.dos.SkipScene() // any key or click skips a scene
	}
	g.advance()
	if g.dos != nil && g.dos.replay != nil {
		if !g.updateReplay() {
			return ebiten.Termination
		}
		return nil
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
		// The wheel steps the zoom levels and the cursor takes the mouse
		// and the keys.
		_, wheel := ebiten.Wheel()
		mx, my := ebiten.CursorPosition()
		g.wheelSplit(wheel, mx, my)
		g.updateCursor()
		return nil
	}
	showPointer(true)
	mx, my, hx, hy := g.cursor()
	// The press rule sees every frame, so a press that starts under a menu or
	// on the map never drives the minimap.
	onMini := g.mini.Contains(float64(hx), float64(hy))
	drive := g.miniMap.press.Update(ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft), onMini)
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

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(mapGround)
	vx, vy, vw, vh := g.cam.Visible()
	if g.split != nil {
		g.drawSplit(screen)
		g.drawWells(screen)
		g.drawFinale(screen)
		g.drawSurrender(screen)
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
	case g.backdrop != nil:
		g.drawBackdrop(dst, vx, vy, vw, vh, inView)
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
		if owner := g.ownerShown(coord); owner != faction.None {
			sx, sy := g.cam.WorldToScreen(hexgrid.Center(coord))
			e := float32(g.cam.Effective())
			vector.FillRect(dst, float32(sx)-20*e, float32(sy)-34*e, 40*e, 6*e, fogShade(g.fogOf(coord), ownerColour(owner)), false)
		}
	}
	for _, coord := range inView {
		if g.dos != nil && g.unitShown(coord) {
			sx, sy := g.cam.WorldToScreen(hexgrid.Center(coord))
			g.drawDosUnit(dst, coord, sx, sy)
		}
	}
	g.drawUnitMarks(dst, inView)
	g.drawGas(dst, inView)
	g.drawTunnels(dst)
	if w := g.walkShown(); !w.Done() && g.fogOf(w.Hex()) == fogVisible {
		wx, wy := w.Pos()
		sx, sy := g.cam.WorldToScreen(wx, wy)
		face, _ := unitsprite.Toward(w.Heading())
		g.drawWalker(dst, w, w.Hex(), sx, sy, face)
	}
	if human && g.dos != nil {
		g.darken(dst, inView) // the cursor is drawn over the frame (drawCursor)
		g.drawOrders(dst)
	}
	// Hit marks show on both sides' maps: both see the battle.
	g.drawMarks(dst)
}

// walkShown is the walk drawn on the map being drawn: the shown side's own.
func (g *game) walkShown() *walk {
	if g.dos == nil {
		return nil
	}
	return g.dos.walkOf(g.shownSide())
}

// drawStatus prints the status and help text at the top left of the HUD.
func (g *game) drawStatus(dst *ebiten.Image) { ebitenutil.DebugPrint(dst, g.statusText()) }

// statusText is the status text: the status wells carry the state and the
// message, and only art failures stay here.
func (g *game) statusText() string { return g.splitErrors() }

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
// resolution; the HUD is magnified separately (see view). The split screen
// keeps the window large enough for its smallest panels.
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

// loadLevel loads name: a setup of a level folder below dir ("1914-08:ec",
// see loadScenario), or a level file of the legacy format by its path
// ("art/reference/pulse/level.json").
func loadLevel(dir, name string) (*level.Level, error) {
	if scenarioName(name) {
		return loadScenario(dir, name)
	}
	if !strings.HasSuffix(name, ".json") {
		return nil, fmt.Errorf("level %q: not a level folder's setup (1914-08:cp), original/NN or a .json file", name)
	}
	return level.Load(name)
}

// matchConfig is everything a battle is started from: by the menu, or directly
// by -level or -player.
type matchConfig struct {
	lvl        *level.Level
	board      *dosbind.Board // an original map's exact binding (loadMap); nil binds lvl
	set        *data.Set
	player     faction.Player // the human side, whose view is shown
	opponent   string         // ai or pass
	limit      int            // dos: orders per phase, 0 for no limit
	reveal     bool
	hexNumbers bool
	reduced    bool
	scroll     float64         // map scroll speed, screen pixels per frame
	units      *artPacks       // shared by every battle; nil only in tests
	screens    *screenArt      // the screen catalog for the portraits; nil shows blank photos
	scores     *scores.Store   // the best scores, opened once; nil keeps none
	saves      *saves.Store    // the single saved battle; nil keeps none
	saved      []byte          // a saved battle to resume (dos only); nil starts fresh
	terrain    string          // terrain catalog directory, composed per level; "" draws coloured hexes
	backdrop   *backdropImage  // the painted ground of every level (-backdrop); nil takes the level's own (levelBackdrop)
	noBackdrop bool            // draw the terrain even where the level has a backdrop (-backdrop off)
	levels     string          // the directory of the level files (-levels)
	replay     *llmplay.Replay // a recorded llmplay game played on both sides (-replay); nil plays a battle
}

// newMatch starts the battle of c.
func newMatch(c matchConfig) (*game, error) {
	gm, err := newDosMatch(c)
	if err != nil {
		return nil, err
	}
	gm.hexNumbers = c.hexNumbers
	gm.drv().SetReducedMotion(c.reduced)
	if c.scroll > 0 {
		gm.scroll = c.scroll
	}
	gm.art, gm.screens, gm.scores = c.units, c.screens, c.scores
	bd := c.backdrop
	if bd == nil && !c.noBackdrop && c.lvl.Extra != nil && c.lvl.Extra.Backdrop != "" {
		if bd, err = levelBackdrop(c.levels, c.lvl.Extra.Backdrop); err != nil {
			return nil, err
		}
	}
	if bd != nil {
		if gm.backdrop, err = bd.on(c.lvl.Grid()); err != nil {
			return nil, err
		}
	}
	if c.terrain != "" {
		if gm.terrain, err = loadTerrainArt(c.terrain, c.lvl); err != nil {
			return nil, err
		}
	}
	return gm, nil
}

// newDosMatch is the battle of c under the original rules, without its art.
func newDosMatch(c matchConfig) (*game, error) {
	if c.replay != nil {
		return replayGame(c.replay, c.reveal), nil
	}
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
	return newDosGame(c.lvl, s, s.SideOf(c.player), planner, c.reveal), nil
}

func main() {
	name := flag.String("level", "1914-08:cp", "a level folder's setup (1914-08:ec, 1918-11:2p), original/NN for a map of the local original files, or the path of a level file of the legacy format (art/reference/pulse/level.json); skips the menu")
	dir := flag.String("levels", "levels", "directory with the level folders")
	side := flag.String("player", "", "side the player plays and whose view is shown: central, entente (default: the level's side); skips the menu")
	reveal := flag.Bool("reveal", false, "dev: show the whole map, ignoring fog of war (dos: every unit where its own side sees it)")
	hexNumbers := flag.Bool("hexnumbers", false, "dev: print line,col on every hex")
	reduced := flag.Bool("reduced-motion", false, "walk instantly and show short static effects (toggle in game with M)")
	terrain := flag.String("terrain", "", "terrain tile catalog built by hl assets build-terrain (default: build/packs/terrain when present, else coloured hexes)")
	screens := flag.String("screens", "", "screen catalog built by hl assets build-screens (default: build/packs/screens when present, else plain menu panels)")
	opponent := flag.String("opponent", "ai", "who plays the other side: ai (computer opponent) or pass (passes every turn)")
	capture := flag.String("capture", "", "dev: render review frames of the battle screen to a directory under build/ and exit (requires the terrain pack)")
	captureMap := flag.Bool("capture-map", false, "with -capture: render the nine 720p/1080p/4K whole-map frames of the PULSE study (hl review gallery -captures) instead")
	captureLook := flag.Bool("capture-look", false, "with -capture: render whole-map revealed/fog views at exact 0.5x, 1x and 1.6x effective zoom at 1080p/4K")
	backdropPath := flag.String("backdrop", "", "a PNG or JPEG painting of the whole level's ground, drawn in place of the terrain tiles of every level, without a hex grid (must match the level's world aspect); off draws the tiles where a level has its own painting")
	snapshotPath := flag.String("snapshot", "", "start in a saved dos battle: the save slot file (save.json in the config directory, naming its level and side) or a bare dos save such as an llmplay game's save.json (then -level names the level; -player overrides the side)")
	replayDir := flag.String("replay", "", "a recorded llmplay game directory (build/review/llmplay/<game>): watch it in the battle screen, both sides played from its transcript (the level and side come from the transcript)")
	record := recordConfig{}
	flag.StringVar(&record.out, "record", "", "with -replay: record the replay as an MP4 video of the battle screen under build/ (needs ffmpeg) and exit")
	recordSize := flag.String("record-size", "1920x1080", "with -record: the video's size, WxH, both even")
	flag.IntVar(&record.fps, "record-fps", 30, "with -record: video frames per second")
	flag.IntVar(&record.speed, "record-speed", 2, "with -record: game ticks per video frame (the game runs 60 ticks a second: 2 at 30 fps is real time)")
	flag.DurationVar(&record.tail, "record-tail", 3*time.Second, "with -record: video after the replay is done")
	flag.IntVar(&record.crf, "record-crf", 23, "with -record: x264 quality (higher is smaller)")
	flag.IntVar(&record.maxFrames, "record-max-frames", 0, "with -record: stop after this many video frames (0: record to the end)")
	flag.IntVar(&record.zoom, "record-zoom", -2, "with -record: zoom levels from the fitted zoom, as many mouse wheel steps; negative zooms out, kept within the levels the screen has")
	flag.Parse()
	if (*captureLook || *captureMap) && *capture == "" {
		log.Fatal("-capture-look and -capture-map require -capture")
	}
	direct := *capture != "" || *replayDir != "" || *snapshotPath != ""
	given := map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		given[f.Name] = true
		direct = direct || f.Name == "level" || f.Name == "player"
		if *replayDir != "" && (f.Name == "level" || f.Name == "player" || f.Name == "opponent") {
			log.Fatalf("-replay takes the level and the sides from the transcript: drop -%s", f.Name)
		}
	})

	// The painted art is the default: the packs just packs builds, where they are.
	for name, path := range map[string]*string{"terrain": terrain, "screens": screens} {
		if !given[name] {
			*path = defaultPack(name)
		}
	}

	if err := record.check(*replayDir, *capture, *reduced, given); err != nil {
		log.Fatal(err)
	}
	if record.out != "" {
		var err error
		if record.w, record.h, err = parseRecordSize(*recordSize); err != nil {
			log.Fatal(err)
		}
	}

	set, err := data.Load()
	if err != nil {
		log.Fatal(err)
	}
	if *opponent != "ai" && *opponent != "pass" {
		log.Fatalf("unknown -opponent %q (want ai or pass)", *opponent)
	}
	c := matchConfig{
		set: set, opponent: *opponent, reveal: *reveal, hexNumbers: *hexNumbers,
		reduced: *reduced, terrain: *terrain, scores: scores.Open(), saves: saves.Open(), levels: *dir,
	}
	// The unit art is always there: a battle draws no unit without it.
	if c.units, err = loadUnitArt(filepath.Join(packsDir, "units")); err != nil {
		log.Fatalf("unit art: %v (build it with just packs)", err)
	}
	ebiten.SetWindowSize(startWidth, startHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Historyline 1914-1918")
	switch *backdropPath {
	case "":
	case "off":
		c.noBackdrop = true
	default:
		if c.backdrop, err = loadBackdrop(*backdropPath); err != nil {
			log.Fatal(err)
		}
	}
	if *screens != "" {
		if c.screens, err = loadScreenArt(*screens); err != nil {
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
		a := newApp(c, campaign, c.screens)
		if err := ebiten.RunGame(a); err != nil {
			log.Fatal(err)
		}
		return
	}

	var snap snapshot
	if *replayDir != "" && *snapshotPath != "" {
		log.Fatal("-replay and -snapshot do not combine")
	}
	if *replayDir != "" {
		if c.replay, err = llmplay.OpenReplay(*replayDir); err != nil {
			log.Fatal(err)
		}
		if c.lvl, err = replayLevel(c.replay); err != nil {
			log.Fatal(err)
		}
	} else {
		level := *name
		if *snapshotPath != "" {
			if snap, err = startSnapshot(*snapshotPath, given["level"], *name); err != nil {
				log.Fatal(err)
			}
			if snap.level != "" {
				level = snap.level
			}
		}
		if c.lvl, c.board, err = loadMap(*dir, level, set); err != nil {
			log.Fatal(err)
		}
	}
	if c.player, err = faction.ByName(c.lvl.Player); err != nil {
		log.Fatal(err)
	}
	switch *side {
	case "":
	case "central":
		c.player = faction.CentralPowers
	case "entente":
		c.player = faction.Entente
	default:
		log.Fatalf("unknown -player %q (want central or entente)", *side)
	}
	if snap.data != nil {
		c.saved = snap.data
		if snap.side >= 0 && *side == "" {
			c.player = dosbind.Factions[snap.side]
		}
	}
	gm, err := newMatch(c)
	if err != nil {
		log.Fatal(err)
	}
	if *capture != "" {
		if gm.terrain == nil {
			log.Fatal("-capture requires the terrain pack")
		}
		if *captureMap || *captureLook {
			gm.mapOnly, gm.split = true, nil // the whole map, without the battle screen
		}
		if err := writeReviewFrames(gm, *capture, *captureLook); err != nil {
			log.Fatal(err)
		}
		return
	}
	if record.out != "" {
		if err := recordReplay(gm, record, os.Stdout); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := ebiten.RunGame(gm); err != nil {
		log.Fatal(err)
	}
}
