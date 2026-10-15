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
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

const (
	startWidth  = 1800
	startHeight = 1000
	scrollSpeed = 12 // screen pixels per frame
)

// Pixel outline of a hex around its centre (flat top, 96×80).
var hexOutline = [6][2]float32{{-48, 0}, {-25, -40}, {25, -40}, {48, 0}, {25, 40}, {-25, 40}}

type game struct {
	lvl         *level.Level
	grid        hexgrid.Grid
	cam         *camera.Camera
	hover       hexgrid.Coord
	hoverOnMap  bool
	w, h        int
	terrainFill map[string]color.RGBA
}

func newGame(lvl *level.Level) *game {
	g := lvl.Grid()
	world := camera.Size{
		W: float64(g.Width*hexgrid.ColPitch + hexgrid.HalfPitch),
		H: float64((g.Lines-1)*hexgrid.LinePitch + hexgrid.HexHeight),
	}
	return &game{
		lvl: lvl, grid: g,
		cam: camera.New(world, camera.Size{W: startWidth, H: startHeight}),
		w:   startWidth, h: startHeight,
		terrainFill: map[string]color.RGBA{},
	}
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

	mx, my := ebiten.CursorPosition()
	if _, wheel := ebiten.Wheel(); wheel != 0 {
		factor := 1.1
		if wheel < 0 {
			factor = 1 / 1.1
		}
		g.cam.ZoomAt(factor, float64(mx), float64(my))
	}
	wx, wy := g.cam.ScreenToWorld(float64(mx), float64(my))
	g.hover, g.hoverOnMap = g.grid.PixelToHex(wx, wy)
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
			g.drawHex(screen, float32(sx), float32(sy), g.fill(hx.Terrain), false)
			if hx.Unit != "" {
				vector.FillCircle(screen, float32(sx), float32(sy), float32(14*g.cam.Zoom), color.RGBA{230, 230, 230, 255}, true)
			}
		}
	}
	if g.hoverOnMap {
		cx, cy := hexgrid.Center(g.hover)
		sx, sy := g.cam.WorldToScreen(cx, cy)
		g.drawHex(screen, float32(sx), float32(sy), color.RGBA{}, true)
	}
	msg := fmt.Sprintf("%s  zoom %.2f  arrows/WASD scroll, wheel zoom", g.lvl.MapName, g.cam.Zoom)
	if g.hoverOnMap {
		hx := g.lvl.Lines[g.hover.Line][g.hover.Col]
		msg += fmt.Sprintf("\nline %d col %d  %s", g.hover.Line, g.hover.Col, hx.Terrain)
	}
	ebitenutil.DebugPrint(screen, msg)
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
	}
	return outsideW, outsideH
}

func main() {
	name := flag.String("level", "level1_cp", "shipped level to load from levels/ (name without .json)")
	dir := flag.String("levels", "levels", "directory with the level files")
	flag.Parse()

	lvl, err := level.Load(filepath.Join(*dir, *name+".json"))
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(startWidth, startHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Historyline 1914-1918")
	if err := ebiten.RunGame(newGame(lvl)); err != nil {
		log.Fatal(err)
	}
}
