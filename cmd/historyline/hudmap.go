package main

import (
	"image"
	"image/color"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/minimap"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// The minimap of the single-map HUD (the remake rules): the cell-exact raster
// of internal/minimap in the bottom-right corner, terrain under the player's
// fog, unit dots in the side colours and the camera rectangle. Only a press
// that starts on it moves the camera.

// miniMap is the minimap's raster, its picture and the press on it.
type miniMap struct {
	raster minimap.Raster
	img    *ebiten.Image
	fills  []color.RGBA   // the hex colours the picture shows
	units  []minimap.Unit // and its dots
	press  minimap.Press
}

// layoutMinimap fits the raster into the HUD's minimap corner.
func (g *game) layoutMinimap() {
	g.mini, g.miniMap.raster = minimap.FitGrid(g.grid, camera.Size{W: float64(g.hud.MinimapMax), H: float64(g.hud.MinimapMax)},
		camera.Size{W: float64(g.w), H: float64(g.h)}, minimapMargin)
}

// sideShade is the minimap colour of a player: the Central Powers (the
// German side) blue, the Entente yellow.
func sideShade(p state.Player) minimap.Shade {
	switch p {
	case state.CentralPowers:
		return minimap.Blue
	case state.Entente:
		return minimap.Yellow
	}
	return minimap.Neutral
}

// dimShade is the dimmed look of a remembered unit.
func dimShade(s minimap.Shade) minimap.Shade {
	dim := func(c color.RGBA) color.RGBA {
		return color.RGBA{c.R / 2, c.G / 2, c.B / 2, c.A}
	}
	return minimap.Shade{Lit: dim(s.Lit), Mid: dim(s.Mid), Dark: dim(s.Dark)}
}

// hexColour is the minimap colour of c's terrain: its mean colour in the
// backdrop unless the sappers changed it, else the average colour of its
// terrain tile when terrain art is loaded, else the placeholder fill.
func (g *game) hexColour(c hexgrid.Coord) color.RGBA {
	if b := g.backdrop; b != nil && !g.reworkedHex[c] {
		if col, ok := b.tint[c]; ok {
			return col
		}
	}
	if g.terrain != nil {
		if bg := g.terrain.scene.Cells[c.Line][c.Col].Background; bg != "" {
			if col, ok := g.tileColour(bg); ok {
				return col
			}
		}
	}
	return g.fill(g.lvl.Lines[c.Line][c.Col].Terrain)
}

// tileColour is the average colour of terrain tile id's opaque texels at the
// smallest loaded density, computed once.
func (g *game) tileColour(id string) (color.RGBA, bool) {
	if c, ok := g.tileTint[id]; ok {
		return c, c.A != 0
	}
	if g.tileTint == nil {
		g.tileTint = map[string]color.RGBA{}
	}
	d := terrainDensities[0]
	var c color.RGBA
	if p, ok := g.terrain.packs[packKey{g.terrain.group[id], d}]; ok {
		if img, err := p.Image(id, d); err == nil {
			c = averageColour(img)
		}
	}
	g.tileTint[id] = c
	return c, c.A != 0
}

// averageColour is the alpha-weighted mean colour of img, transparent when img
// has no opaque texel.
func averageColour(img image.Image) color.RGBA {
	var r, gr, b, a uint64
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pr, pg, pb, pa := img.At(x, y).RGBA() // premultiplied
			r, gr, b, a = r+uint64(pr), gr+uint64(pg), b+uint64(pb), a+uint64(pa)
		}
	}
	if a == 0 {
		return color.RGBA{}
	}
	return color.RGBA{uint8(r * 255 / a), uint8(gr * 255 / a), uint8(b * 255 / a), 255}
}

// minimapScene is what the minimap shows: a colour per hex in line order and
// the unit dots, with the walking unit on the hex it walks through.
func (g *game) minimapScene() (fills []color.RGBA, units []minimap.Unit) {
	for l, line := range g.lvl.Lines {
		for c := range line {
			coord := hexgrid.Coord{Line: l, Col: c}
			fills = append(fills, fogShade(g.fogOf(coord), g.hexColour(coord)))
			shown, remembered := g.unitShown(coord)
			switch {
			case !shown:
			case remembered:
				if m, _ := g.state.Remembered(g.player, coord); m.Unit != nil {
					units = append(units, minimap.Unit{At: coord, Shade: dimShade(sideShade(m.Unit.Owner))})
				}
			default:
				units = append(units, minimap.Unit{At: coord, Shade: sideShade(g.state.At(coord).Unit.Owner)})
			}
		}
	}
	if w := g.drv().ui().walk; !w.Done() && g.fogOf(w.Hex()) == state.Visible {
		if u, ok := w.unit.(*state.Unit); ok && u != nil {
			units = append(units, minimap.Unit{At: w.Hex(), Shade: sideShade(u.Owner)})
		}
	}
	return fills, units
}

// drawMinimap draws the minimap with the same fog rules as the main view, so
// unseen terrain and hidden units never show, plus the camera rectangle. The
// picture is redrawn only when what it shows changed.
func (g *game) drawMinimap(dst *ebiten.Image) {
	m, mm := g.mini, &g.miniMap
	size := mm.raster.Size()
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	fills, units := g.minimapScene()
	if mm.img == nil || mm.img.Bounds().Size() != size || !slices.Equal(fills, mm.fills) || !slices.Equal(units, mm.units) {
		if mm.img == nil || mm.img.Bounds().Size() != size {
			if mm.img != nil {
				mm.img.Deallocate()
			}
			mm.img = ebiten.NewImage(size.X, size.Y)
		}
		widths := make([]int, len(g.lvl.Lines)+1)
		for l, line := range g.lvl.Lines {
			widths[l+1] = widths[l] + len(line)
		}
		pix := mm.raster.Render(func(c hexgrid.Coord) color.RGBA {
			if c.Line < 0 || c.Line >= len(g.lvl.Lines) || c.Col < 0 || c.Col >= len(g.lvl.Lines[c.Line]) {
				return color.RGBA{}
			}
			return fills[widths[c.Line]+c.Col]
		})
		for _, u := range units {
			mm.raster.DrawUnit(pix, image.Point{}, u)
		}
		mm.img.WritePixels(pix.Pix)
		mm.fills, mm.units = fills, units
	}
	vector.FillRect(dst, float32(m.Pos.X)-2, float32(m.Pos.Y)-2, float32(size.X)+4, float32(size.Y)+4, color.RGBA{0, 0, 0, 220}, false)
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(m.Pos.X, m.Pos.Y)
	dst.DrawImage(mm.img, &op)
	x, y, w, h := m.ViewRect(g.cam)
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1.5, color.RGBA{255, 240, 80, 255}, false)
}
