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

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/minimap"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

const (
	startWidth  = 1800
	startHeight = 1000
	scrollSpeed = 12 // screen pixels per frame
)

// Pixel outline of a hex around its centre (flat top, 96×80).
var hexOutline = [6][2]float32{{-48, 0}, {-25, -40}, {25, -40}, {48, 0}, {25, 40}, {-25, 40}}

const (
	minimapMax    = 280 // longest minimap edge in screen pixels
	minimapMargin = 12
	minNumberZoom = 0.6 // below this the hex numbers would overlap
)

type game struct {
	menuMX, menuMY int // last cursor position seen by the menu
	menuFocus      int // focused action-menu row, menuNone when unset
	lvl            *level.Level
	state          *state.State
	player         state.Player // whose fog is shown
	reveal         bool         // dev: ignore fog
	hexNumbers     bool         // dev: print line,col on every hex
	grid           hexgrid.Grid
	cam            *camera.Camera
	mini           minimap.Minimap
	hover          hexgrid.Coord
	hoverOnMap     bool
	w, h           int
	terrainFill    map[string]color.RGBA
	ctl            *control
}

func newGame(lvl *level.Level, st *state.State, player state.Player, reveal bool) *game {
	g := lvl.Grid()
	world := camera.Size{
		W: float64(g.Width*hexgrid.ColPitch + hexgrid.HalfPitch),
		H: float64((g.Lines-1)*hexgrid.LinePitch + hexgrid.HexHeight),
	}
	gm := &game{
		lvl: lvl, state: st, player: player, reveal: reveal, grid: g,
		cam: camera.New(world, camera.Size{W: startWidth, H: startHeight}),
		w:   startWidth, h: startHeight,
		terrainFill: map[string]color.RGBA{},
		ctl:         &control{st: st, player: player},
	}
	gm.layoutMinimap()
	return gm
}

func (g *game) layoutMinimap() {
	g.mini = minimap.Fit(g.cam.World, camera.Size{W: minimapMax, H: minimapMax},
		camera.Size{W: float64(g.w), H: float64(g.h)}, minimapMargin)
}

// fogOf is what the shown player knows about c; everything is visible in reveal mode.
func (g *game) fogOf(c hexgrid.Coord) state.Fog {
	if g.reveal {
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

// walking reports whether the unit standing on at is still being animated there.
func (g *game) walking(at hexgrid.Coord) bool {
	w := g.ctl.walk
	return !w.Done() && w.path[len(w.path)-1] == at
}

// showsUnit reports whether units on a hex of the given fog state are drawn.
func showsUnit(f state.Fog) bool { return f == state.Visible }

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
	dx, dy := 0.0, 0.0
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		dx -= scrollSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		dx += scrollSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		dy -= scrollSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		dy += scrollSpeed
	}
	g.cam.Scroll(dx, dy)
	if g.ctl.walk != nil {
		g.ctl.walk.Tick()
	}
	g.ctl.report.Tick()
	for _, a := range actionsFor(inpututil.IsKeyJustPressed) {
		if !g.applyAction(a) {
			return ebiten.Termination
		}
	}
	if g.updateMenu() {
		return nil
	}

	mx, my := ebiten.CursorPosition()
	if _, wheel := ebiten.Wheel(); wheel != 0 {
		factor := 1.1
		if wheel < 0 {
			factor = 1 / 1.1
		}
		g.cam.ZoomAt(factor, float64(mx), float64(my))
	}
	if g.mini.Contains(float64(mx), float64(my)) {
		g.hoverOnMap = false
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.cam.CenterOn(g.mini.ToWorld(float64(mx), float64(my)))
		}
		return nil
	}
	wx, wy := g.cam.ScreenToWorld(float64(mx), float64(my))
	g.hover, g.hoverOnMap = g.grid.PixelToHex(wx, wy)
	if g.hoverOnMap && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.ctl.Click(g.hover)
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		g.ctl.clear()
	}
	return nil
}

const menuRowH, menuW = 18, 340

// menuRow is the screen rectangle of weapon menu row i (bottom left).
func (g *game) menuRow(i int) (x, y, w, h int) {
	return 8, g.h - 8 - (len(g.ctl.menu)-i)*menuRowH, menuW, menuRowH
}

// updateMenu handles the open weapon menu (number keys or a click on a row) and
// reports whether it consumed this frame's input.
func (g *game) updateMenu() bool {
	if !g.ctl.MenuOpen() {
		g.menuFocus = menuNone
		return false
	}
	g.menuFocus = menuFocusValid(g.menuFocus, g.ctl.menu)
	mx, my := ebiten.CursorPosition()
	if mx != g.menuMX || my != g.menuMY {
		g.menuMX, g.menuMY = mx, my
		if h := menuHit(mx, my, len(g.ctl.menu), g.menuRow); h != menuNone {
			g.menuFocus = h
		}
	}
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowDown):
		g.menuFocus = menuStep(g.menuFocus, 1, g.ctl.menu)
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowUp):
		g.menuFocus = menuStep(g.menuFocus, -1, g.ctl.menu)
	case inpututil.IsKeyJustPressed(ebiten.KeyEnter), inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter):
		if g.menuFocus != menuNone {
			g.ctl.ChooseMenu(g.menuFocus)
			return true
		}
	}
	for i, k := range []ebiten.Key{ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4, ebiten.Key5, ebiten.Key6, ebiten.Key7, ebiten.Key8, ebiten.Key9} {
		if inpututil.IsKeyJustPressed(k) && i < len(g.ctl.menu) {
			g.ctl.ChooseMenu(i)
			return true
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if h := menuHit(mx, my, len(g.ctl.menu), g.menuRow); h != menuNone {
			g.ctl.ChooseMenu(h)
			return true
		}
	}
	return false
}

func (g *game) drawMenu(dst *ebiten.Image) {
	for i, o := range g.ctl.menu {
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
		vector.FillRect(dst, float32(x), float32(y), float32(w), float32(h), bg, false)
		if i == g.menuFocus {
			vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 2, color.RGBA{240, 220, 120, 255}, false)
		}
		ebitenutil.DebugPrintAt(dst, menuLabel(i, o), x+4, y+1)
	}
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 24, 255})
	vx, vy, vw, vh := g.cam.Visible()
	for l, line := range g.lvl.Lines {
		for c, hx := range line {
			coord := hexgrid.Coord{Line: l, Col: c}
			cx, cy := hexgrid.Center(coord)
			if cx+48 < vx || cx-48 > vx+vw || cy+40 < vy || cy-40 > vy+vh {
				continue
			}
			sx, sy := g.cam.WorldToScreen(cx, cy)
			fog := g.fogOf(coord)
			g.drawHex(screen, float32(sx), float32(sy), fogShade(fog, g.fill(hx.Terrain)), false)
			if g.state.At(coord).Unit != nil && showsUnit(fog) && !g.walking(coord) {
				vector.FillCircle(screen, float32(sx), float32(sy), float32(14*g.cam.Zoom), color.RGBA{230, 230, 230, 255}, true)
			}
		}
	}
	if w := g.ctl.walk; !w.Done() {
		wx, wy := w.Pos()
		sx, sy := g.cam.WorldToScreen(wx, wy)
		vector.FillCircle(screen, float32(sx), float32(sy), float32(14*g.cam.Zoom), color.RGBA{230, 230, 230, 255}, true)
	}
	if g.hexNumbers && g.cam.Zoom >= minNumberZoom {
		g.drawHexNumbers(screen, vx, vy, vw, vh)
	}
	g.drawOverlays(screen)
	if g.hoverOnMap {
		cx, cy := hexgrid.Center(g.hover)
		sx, sy := g.cam.WorldToScreen(cx, cy)
		g.drawHex(screen, float32(sx), float32(sy), color.RGBA{}, true)
	}
	g.drawMinimap(screen)
	g.drawMenu(screen)
	g.drawReport(screen)
	if dev := g.devIndicators(); dev != "" {
		ebitenutil.DebugPrintAt(screen, dev, g.w-8*len(dev)-8, 0)
	}
	msg := fmt.Sprintf("%s  round %d, %v to move  zoom %.2f\narrows/WASD scroll, wheel zoom, click unit then hex (1-9 picks menu row; click own building or selected carrier for contents), F1 end turn, Esc/right-click cancel",
		g.lvl.MapName, g.state.Round, g.state.Turn, g.cam.Zoom)
	if g.state.Winner != state.None {
		msg += fmt.Sprintf("\n%v has won", g.state.Winner)
	}
	if g.ctl.Message != "" {
		msg += "\n" + g.ctl.Message
	}
	if g.hoverOnMap {
		hx := g.lvl.Lines[g.hover.Line][g.hover.Col]
		msg += fmt.Sprintf("\nline %d col %d  %s", g.hover.Line, g.hover.Col, hx.Terrain)
	}
	ebitenutil.DebugPrint(screen, msg)
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
			ebitenutil.DebugPrintAt(dst, fmt.Sprintf("%d,%d", l, c), int(sx)-14, int(sy+22*g.cam.Zoom)-8)
		}
	}
}

// drawOverlays tints the hexes the selected unit may reach (blue) or attack
// (red) and outlines the selection, all taken from the rules engine's queries.
func (g *game) drawOverlays(dst *ebiten.Image) {
	c := g.ctl
	if !c.selected {
		return
	}
	tint := func(at hexgrid.Coord, col color.RGBA) {
		cx, cy := hexgrid.Center(at)
		sx, sy := g.cam.WorldToScreen(cx, cy)
		g.drawHex(dst, float32(sx), float32(sy), col, false)
	}
	for at := range c.moves {
		tint(at, color.RGBA{60, 120, 255, 90})
	}
	for at := range c.targets {
		tint(at, color.RGBA{255, 50, 40, 130})
	}
	cx, cy := hexgrid.Center(c.from)
	sx, sy := g.cam.WorldToScreen(cx, cy)
	g.drawHex(dst, float32(sx), float32(sy), color.RGBA{}, true)
}

// drawMinimap draws one small cell per hex with the same fog rules as the main
// view, so unseen terrain and hidden units never show, plus the camera rectangle.
func (g *game) drawMinimap(dst *ebiten.Image) {
	m := g.mini
	s := m.Size()
	vector.FillRect(dst, float32(m.Pos.X)-2, float32(m.Pos.Y)-2, float32(s.W)+4, float32(s.H)+4, color.RGBA{0, 0, 0, 220}, false)
	cw := float32(hexgrid.ColPitch * m.Scale)
	ch := float32(hexgrid.HexHeight * m.Scale)
	for l, line := range g.lvl.Lines {
		for c, hx := range line {
			coord := hexgrid.Coord{Line: l, Col: c}
			cx, cy := hexgrid.Center(coord)
			sx, sy := m.FromWorld(cx, cy)
			fog := g.fogOf(coord)
			col := fogShade(fog, g.fill(hx.Terrain))
			if g.state.At(coord).Unit != nil && showsUnit(fog) && !g.walking(coord) {
				col = color.RGBA{240, 240, 240, 255}
			}
			vector.FillRect(dst, float32(sx)-cw/2, float32(sy)-ch/2, cw, ch, col, false)
		}
	}
	x, y, w, h := m.ViewRect(g.cam)
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1.5, color.RGBA{255, 240, 80, 255}, false)
}

func (g *game) drawHex(dst *ebiten.Image, cx, cy float32, fill color.RGBA, outline bool) {
	z := float32(g.cam.Zoom)
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
		vector.StrokePath(dst, &p, &vector.StrokeOptions{Width: 3}, pathOptions(color.RGBA{255, 240, 80, 255}))
		return
	}
	vector.FillPath(dst, &p, &vector.FillOptions{}, pathOptions(fill))
	vector.StrokePath(dst, &p, &vector.StrokeOptions{Width: 1}, pathOptions(color.RGBA{0, 0, 0, 90}))
}

func pathOptions(c color.RGBA) *vector.DrawPathOptions {
	var o vector.DrawPathOptions
	o.AntiAlias = true
	o.ColorScale.ScaleWithColor(c)
	return &o
}

func (g *game) Layout(outsideW, outsideH int) (int, int) {
	if outsideW != g.w || outsideH != g.h {
		g.w, g.h = outsideW, outsideH
		g.cam.Resize(camera.Size{W: float64(outsideW), H: float64(outsideH)})
		g.layoutMinimap()
	}
	return outsideW, outsideH
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

func main() {
	name := flag.String("level", "level1_cp", "shipped level to load from levels/ (name without .json)")
	dir := flag.String("levels", "levels", "directory with the level files")
	side := flag.String("player", "", "side whose fog of war is shown: central, entente (default: the level's side)")
	reveal := flag.Bool("reveal", false, "dev: show the whole map, ignoring fog of war")
	hexNumbers := flag.Bool("hexnumbers", false, "dev: print line,col on every hex")
	flag.Parse()

	lvl, err := loadLevel(*dir, *name)
	if err != nil {
		log.Fatal(err)
	}
	set, err := data.Load()
	if err != nil {
		log.Fatal(err)
	}
	st, err := state.New(lvl, set)
	if err != nil {
		log.Fatal(err)
	}
	player := st.Human
	switch *side {
	case "":
	case "central":
		player = state.CentralPowers
	case "entente":
		player = state.Entente
	default:
		log.Fatalf("unknown -player %q (want central or entente)", *side)
	}
	ebiten.SetWindowSize(startWidth, startHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Historyline 1914-1918")
	gm := newGame(lvl, st, player, *reveal)
	gm.hexNumbers = *hexNumbers
	if err := ebiten.RunGame(gm); err != nil {
		log.Fatal(err)
	}
}

// drawReport shows the battle panel at the top centre while it is active.
func (g *game) drawReport(dst *ebiten.Image) {
	r := g.ctl.report
	if !r.Active() {
		return
	}
	const w, h = 360, 70
	x := (g.w - w) / 2
	vector.FillRect(dst, float32(x), 4, w, h, color.RGBA{20, 20, 40, 235}, false)
	for i, l := range r.Lines() {
		ebitenutil.DebugPrintAt(dst, l, x+6, 8+i*32)
	}
}
