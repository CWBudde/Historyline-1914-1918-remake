package main

import (
	"image"
	"image/color"

	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/minimap"
)

// The colours of the strategic map (internal/minimap): the side colours of
// units and buildings (greyish once acted) and the terrain colour of a hex.

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

// spentShade is the greyish look of a unit that has acted this turn, as its
// sprite turns grey: four fifths grey, one fifth the side colour.
func spentShade(s minimap.Shade) minimap.Shade {
	grey := func(c color.RGBA) color.RGBA {
		y := (299*int(c.R) + 587*int(c.G) + 114*int(c.B)) / 1000
		mix := func(v uint8) uint8 { return uint8((4*y + int(v)) / 5) }
		return color.RGBA{mix(c.R), mix(c.G), mix(c.B), c.A}
	}
	return minimap.Shade{Lit: grey(s.Lit), Mid: grey(s.Mid), Dark: grey(s.Dark)}
}

// unitShade is the minimap colour of a unit of owner, greyish once spent.
func unitShade(owner state.Player, spent bool) minimap.Shade {
	if spent {
		return spentShade(sideShade(owner))
	}
	return sideShade(owner)
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
