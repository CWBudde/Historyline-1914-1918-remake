package main

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// The gas cloud of the extension's gas pioneers (docs/new-units.md) is drawn
// over the units of both panels' maps: soft yellow-green puffs on each
// gassed hex that drift slowly and fade as the cloud thins. Reduced motion
// holds them still.
const (
	gasPuffs    = 7   // puffs per hex
	gasFullTurn = 2   // turns left from which a cloud shows at full strength
	gasOpacity  = 0.5 // a puff's peak opacity at full strength
	gasPuffSize = 64  // the puff texture's size in pixels
)

// gasColour is a puff's colour, chlorine yellow-green.
var gasColour = color.RGBA{196, 204, 92, 255}

// gasPuff is a white disc whose opacity falls off smoothly to its rim.
var gasPuff = func() *ebiten.Image {
	const n = gasPuffSize
	im := image.NewRGBA(image.Rect(0, 0, n, n))
	for y := range n {
		for x := range n {
			dx, dy := (float64(x)+.5)/n*2-1, (float64(y)+.5)/n*2-1
			r2 := dx*dx + dy*dy
			if r2 >= 1 {
				continue
			}
			a := uint8(255 * (1 - r2) * (1 - r2))
			im.SetRGBA(x, y, color.RGBA{a, a, a, a})
		}
	}
	return ebiten.NewImageFromImage(im)
}()

// drawGas draws the cloud on the cells in view.
func (g *game) drawGas(dst *ebiten.Image, cells []hexgrid.Coord) {
	if g.dos == nil {
		return
	}
	gas := g.dos.s.Gas()
	if len(gas) == 0 {
		return
	}
	t := float64(g.gasTick)
	if g.dos.reduced {
		t = 0
	}
	e := g.cam.Effective()
	for _, c := range cells {
		turns := gas[c]
		if turns == 0 {
			continue
		}
		strength := math.Min(1, 0.55+0.45*float64(turns)/gasFullTurn)
		cx, cy := hexgrid.Center(c)
		sx, sy := g.cam.WorldToScreen(cx, cy)
		seed := float64(c.Line*31 + c.Col*17)
		for k := range gasPuffs {
			a := seed + float64(k)*2.39996 // the golden angle spreads the puffs round the centre
			r := 8 + 22*math.Abs(math.Sin(a*1.7))
			dx := r*math.Cos(a) + 6*math.Sin(t/95+a)
			dy := 0.7*r*math.Sin(a) + 4*math.Cos(t/120+a*1.3)
			rad := 40 + 10*math.Sin(t/75+a*0.8) // hexes are 96 wide at zoom 1: puffs reach into their neighbours
			op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
			s := 2 * rad * e / gasPuffSize
			op.GeoM.Translate(-gasPuffSize/2, -gasPuffSize/2)
			op.GeoM.Scale(s, s)
			op.GeoM.Translate(sx+dx*e, sy+dy*e)
			alpha := float32(gasOpacity * strength * (0.7 + 0.3*math.Sin(t/60+a)))
			op.ColorScale.Scale(float32(gasColour.R)/255*alpha, float32(gasColour.G)/255*alpha, float32(gasColour.B)/255*alpha, alpha)
			dst.DrawImage(gasPuff, op)
		}
	}
}
