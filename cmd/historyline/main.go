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
)

type game struct {
	lvl         *level.Level
	state       *state.State
	player      state.Player // whose fog is shown
	reveal      bool         // dev: ignore fog
	grid        hexgrid.Grid
	cam         *camera.Camera
	mini        minimap.Minimap
	hover       hexgrid.Coord
	hoverOnMap  bool
	w, h        int
	terrainFill map[string]color.RGBA
	ctl         *control
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
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.ctl.EndTurn()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.ctl.clear()
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
			if hx.Unit != "" && showsUnit(fog) {
				vector.FillCircle(screen, float32(sx), float32(sy), float32(14*g.cam.Zoom), color.RGBA{230, 230, 230, 255}, true)
			}
		}
	}
	g.drawOverlays(screen)
	if g.hoverOnMap {
		cx, cy := hexgrid.Center(g.hover)
		sx, sy := g.cam.WorldToScreen(cx, cy)
		g.drawHex(screen, float32(sx), float32(sy), color.RGBA{}, true)
	}
	g.drawMinimap(screen)
	msg := fmt.Sprintf("%s  round %d, %v to move  zoom %.2f\narrows/WASD scroll, wheel zoom, click unit then hex, F1 end turn, Esc/right-click cancel",
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
			if hx.Unit != "" && showsUnit(fog) {
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
	if err := ebiten.RunGame(newGame(lvl, st, player, *reveal)); err != nil {
		log.Fatal(err)
	}
}
