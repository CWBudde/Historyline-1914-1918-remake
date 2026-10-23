package gesture

import (
	"fmt"
	"image/color"
	"math"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// State is how a plate is drawn.
type State int

// Plate states.
const (
	Normal   State = iota
	Pressed        // pushed in: dark top rim, glyph sunk
	Disabled       // not possible here: faded glyph without highlight
)

var stateNames = [...]string{"normal", "pressed", "disabled"}

func (s State) String() string {
	if s < 0 || int(s) >= len(stateNames) {
		return fmt.Sprintf("State(%d)", int(s))
	}
	return stateNames[s]
}

// unit is one pixel of the original's 24-px cell in world px. The ring band,
// rims and glyph follow the original's proportions in this unit.
const unit = hexgrid.HexHeight / 24.0

// The placeholder steel, olive greys lit from the top left.
var (
	outline   = color.RGBA{24, 28, 0, 255}
	steel     = color.RGBA{81, 81, 65, 255}
	steelDark = color.RGBA{66, 66, 52, 255}
	speck     = color.RGBA{154, 154, 138, 255}
	light     = color.RGBA{199, 199, 190, 255}
	face      = color.RGBA{130, 130, 113, 255}
	shadow    = color.RGBA{44, 36, 28, 255}
	shadowMid = color.RGBA{69, 56, 40, 255}
	faded     = color.RGBA{100, 100, 84, 255}
)

// layer is polygons in device px filled in one colour with the non-zero rule,
// so a reversed polygon cuts a hole. Drawing fills the layers in order.
type layer struct {
	c     color.RGBA
	polys [][][2]float64
}

// The cursor ring in units: the original's steel band 3 wide, between dark
// lines half its 1-pixel outlines wide, so the ring stays fine at high
// resolution; the band is shaded in ringStrips strips.
const (
	ringLine   = 0.5
	ringBand   = 3
	ringStrips = 10
)

// ringLayers is the cursor, as the original's: a steel band between dark
// lines, transparent inside. The band is a rounded rim lit from the top left:
// across it the slope turns from outwards to inwards, so each strip of each
// edge is shaded by how far its slope faces the light (the outer part of the
// top edge and the inner part of the bottom edge catch it).
func ringLayers(cx, cy, scale float64) []layer {
	hex := plate(cx, cy, scale)
	u := unit * scale
	ls := []layer{{outline, [][][2]float64{hex, reversed(inset(hex, (2*ringLine+ringBand)*u))}}}
	w := float64(ringBand) / ringStrips
	for i := range hex {
		a, b := hex[i], hex[(i+1)%len(hex)]
		l := math.Hypot(b[0]-a[0], b[1]-a[1])
		lit := ((b[1]-a[1])*lightDirX - (b[0]-a[0])*lightDirY) / l // outward normal · light
		for k := range ringStrips {
			slope := 1 - (2*float64(k)+1)/ringStrips // +1 at the outer edge, -1 at the inner
			d := ringLine + float64(k)*w
			ls = append(ls, layer{rimShade(lit * slope), [][][2]float64{edgeStrip(hex, d*u, (d+w)*u, i, i+1)}})
		}
	}
	return ls
}

// lightDirX, lightDirY is the direction towards the light: the top left.
var lightDirX, lightDirY = -math.Sqrt2 / 2, -math.Sqrt2 / 2

// rimShade is the steel of a slope facing the light by f (-1..1): towards
// light where it faces it, towards shadow where it faces away.
func rimShade(f float64) color.RGBA {
	to := light
	if f < 0 {
		to, f = shadow, -f
	}
	f = min(f, 1) * 0.8
	mix := func(a, b uint8) uint8 { return uint8(math.Round(float64(a) + f*(float64(b)-float64(a)))) }
	return color.RGBA{mix(steel.R, to.R), mix(steel.G, to.G), mix(steel.B, to.B), 255}
}

// plateLayers is an action plate: the cell outline filled with raised steel,
// a light rim along the top and a light strip near the bottom, and the glyph
// of a, 14 units across, embossed.
func plateLayers(cx, cy, scale float64, a Action, s State) []layer {
	hex := plate(cx, cy, scale)
	u := unit * scale
	body := steel
	if s == Pressed {
		body = steelDark
	}
	ls := []layer{{outline, [][][2]float64{hex}}, {body, [][][2]float64{inset(hex, u)}}}
	switch s {
	case Normal:
		ls = append(ls, layer{speck, [][][2]float64{edgeStrip(hex, u, 2*u, 0, 2)}},
			layer{light, [][][2]float64{edgeStrip(hex, 2*u, 3*u, 4, 5)}})
	case Pressed:
		ls = append(ls, layer{shadow, [][][2]float64{edgeStrip(hex, u, 2*u, 0, 2)}})
	}
	g := glyphAt(a, cx, cy, 14*u)
	switch s {
	case Disabled:
		return append(ls, layer{shadowMid, g(0.5*u, 0.5*u)}, layer{faded, g(0, 0)})
	case Pressed:
		return append(ls, layer{shadow, g(1.5*u, 1.5*u)}, layer{face, g(0.5*u, 0.5*u)})
	}
	return append(ls, layer{light, g(-u, -u)}, layer{shadow, g(u, u)}, layer{face, g(0, 0)})
}

// glyphAt returns the glyph of a in a box of side size centred on (cx, cy),
// moved by (dx, dy).
func glyphAt(a Action, cx, cy, size float64) func(dx, dy float64) [][][2]float64 {
	g := glyph(a)
	return func(dx, dy float64) [][][2]float64 {
		out := make([][][2]float64, len(g))
		for i, p := range g {
			q := make([][2]float64, len(p.pts))
			for j, pt := range p.pts {
				q[j] = [2]float64{cx + dx + pt[0]*size, cy + dy + pt[1]*size}
			}
			out[i] = q
		}
		return out
	}
}

func plate(cx, cy, scale float64) [][2]float64 {
	c := PlateCorners(cx, cy, scale)
	return c[:]
}

// edgeStrip is the band between insets d0 and d1 of hex along its corners
// from..to, wrapping past the last corner.
func edgeStrip(hex [][2]float64, d0, d1 float64, from, to int) [][2]float64 {
	o, i := inset(hex, d0), inset(hex, d1)
	var out [][2]float64
	for k := from; k <= to; k++ {
		out = append(out, o[k%len(o)])
	}
	for k := to; k >= from; k-- {
		out = append(out, i[k%len(i)])
	}
	return out
}

func reversed(p [][2]float64) [][2]float64 {
	out := make([][2]float64, len(p))
	for i, pt := range p {
		out[len(p)-1-i] = pt
	}
	return out
}
