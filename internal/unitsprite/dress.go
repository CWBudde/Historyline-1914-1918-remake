package unitsprite

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"sync"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
)

// A unit on the map is dressed for legibility (art/direction.md, Units): a rim
// 1.5 world px wide around its silhouette, and under it a soft contact shadow,
// the silhouette blurred and cast to the lower right, away from the upper-left
// light. Flying units cast it further off. The shadow stays below half opacity,
// so it never reads as part of the unit body. The rim is dark, except on the
// dark ground of forest and open water: a dark rim has no headroom there, so a
// muted pale one takes its place.

const (
	shadowOffset       = 2.0  // world px, toward the lower right
	flyingShadowOffset = 6.0  // world px, for aircraft
	shadowSigma        = 0.75 // world px of the Gaussian blur
	shadowOpacity      = 0.45
	rimWidth           = 1.5 // world px
)

var shadowColor = color.NRGBA{R: 18, G: 26, B: 16, A: 255}

// Rim is the colour of a dressed unit's rim.
type Rim int

const (
	Dark  Rim = iota // olive-black, on open ground and roads
	Light            // muted pale, in forest and on open water
)

// RimFor is the rim of a unit on a hex of terrain, with or without
// infrastructure: pale on forest (where the scattered crowns stand, as in
// mapcompose) and open water without infrastructure, dark elsewhere, also on
// bridges and forest roads.
func RimFor(terrain string, infra bool) Rim {
	switch terrain {
	case "forest", "forest_w", "water", "waterStones":
		if !infra {
			return Light
		}
	}
	return Dark
}

func (r Rim) colour() color.NRGBA {
	if r == Light {
		return color.NRGBA{R: 196, G: 190, B: 162, A: 255}
	}
	return color.NRGBA{R: 16, G: 18, B: 12, A: 255}
}

// DressMargin is the transparent border in texels that Dress adds on each side
// of a density-d sprite: whole logical pixels, so the logical anchor moves by
// DressMargin(d)/d.
func DressMargin(d int) int {
	return d * int(math.Ceil(flyingShadowOffset/math.Sqrt2+3*shadowSigma+1))
}

var flyingFamilies = sync.OnceValue(func() map[string]bool {
	out := map[string]bool{}
	set, err := data.Load()
	if err != nil {
		return out
	}
	for id, u := range set.Units {
		if f, ok := families[id]; ok && u.MovementType == data.MoveAir {
			out[f] = true
		}
	}
	return out
})

// Flies reports whether family draws aircraft.
func Flies(family string) bool { return flyingFamilies()[family] }

// Dress returns the density-d sprite src with rim r and its contact shadow, on a
// canvas grown by DressMargin(d) on each side.
func Dress(src image.Image, d int, flying bool, r Rim) *image.NRGBA {
	b := src.Bounds()
	m := DressMargin(d)
	sprite := image.NewNRGBA(image.Rect(0, 0, b.Dx()+2*m, b.Dy()+2*m))
	draw.Draw(sprite, b.Sub(b.Min).Add(image.Pt(m, m)), src, b.Min, draw.Src)
	// Work only around the opaque texels: the rim reaches width texels out, the
	// shadow its blur radius plus its shift.
	used := image.Rectangle{}
	for y := range sprite.Rect.Dy() {
		for x := range sprite.Rect.Dx() {
			if sprite.Pix[sprite.PixOffset(x, y)+3] != 0 {
				used = used.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	if used.Empty() {
		return sprite
	}
	offset := shadowOffset
	if flying {
		offset = flyingShadowOffset
	}
	shift := int(math.Round(offset / math.Sqrt2 * float64(d)))
	sigma := shadowSigma * float64(d)
	width := int(math.Round(rimWidth * float64(d)))
	reach := width + int(math.Ceil(3*sigma))
	area := image.Rect(used.Min.X-reach, used.Min.Y-reach, used.Max.X+reach+shift, used.Max.Y+reach+shift).Intersect(sprite.Rect)
	w, h := area.Dx(), area.Dy()
	alpha := make([]float64, w*h)
	for y := range h {
		for x := range w {
			alpha[y*w+x] = float64(sprite.Pix[sprite.PixOffset(area.Min.X+x, area.Min.Y+y)+3]) / 255
		}
	}
	rim := dilate(alpha, w, h, width)
	shadow := blur(alpha, w, h, sigma)

	out := image.NewNRGBA(sprite.Rect)
	for y := range h {
		for x := range w {
			i := y*w + x
			var sa float64
			if sx, sy := x-shift, y-shift; sx >= 0 && sy >= 0 {
				sa = shadowOpacity * shadow[sy*w+sx]
			}
			// Shadow, rim and sprite, bottom to top.
			p := sprite.PixOffset(area.Min.X+x, area.Min.Y+y)
			c := layer{}.over(shadowColor, sa).over(r.colour(), rim[i]).over(color.NRGBA{sprite.Pix[p], sprite.Pix[p+1], sprite.Pix[p+2], sprite.Pix[p+3]}, alpha[i])
			out.Pix[p] = uint8(math.Round(c.r))
			out.Pix[p+1] = uint8(math.Round(c.g))
			out.Pix[p+2] = uint8(math.Round(c.b))
			out.Pix[p+3] = uint8(math.Round(c.a * 255))
		}
	}
	return out
}

// layer is a straight-alpha colour being composited.
type layer struct{ r, g, b, a float64 }

// over returns c at alpha ca composited over l.
func (l layer) over(c color.NRGBA, ca float64) layer {
	na := ca + l.a*(1-ca)
	if na == 0 {
		return l
	}
	return layer{
		(float64(c.R)*ca + l.r*l.a*(1-ca)) / na,
		(float64(c.G)*ca + l.g*l.a*(1-ca)) / na,
		(float64(c.B)*ca + l.b*l.a*(1-ca)) / na,
		na,
	}
}

// dilate returns the largest alpha within a disc of radius r texels: per row of
// the disc, the row maximum over its half width.
func dilate(alpha []float64, w, h, r int) []float64 {
	// rows[k] is the largest alpha within k texels left or right.
	rows := [][]float64{alpha}
	for k := 1; k <= r; k++ {
		prev, next := rows[k-1], make([]float64, len(alpha))
		for y := range h {
			for x := range w {
				v := prev[y*w+x]
				if x >= k {
					v = max(v, alpha[y*w+x-k])
				}
				if x+k < w {
					v = max(v, alpha[y*w+x+k])
				}
				next[y*w+x] = v
			}
		}
		rows = append(rows, next)
	}
	out := make([]float64, len(alpha))
	for dy := -r; dy <= r; dy++ {
		row := rows[int(math.Sqrt(float64(r*r-dy*dy)))]
		for y := max(0, -dy); y < min(h, h-dy); y++ {
			for x := range w {
				out[y*w+x] = max(out[y*w+x], row[(y+dy)*w+x])
			}
		}
	}
	return out
}

// blur returns alpha convolved with a Gaussian of sigma texels.
func blur(alpha []float64, w, h int, sigma float64) []float64 {
	rad := int(math.Ceil(3 * sigma))
	kernel := make([]float64, 2*rad+1)
	var sum float64
	for i := range kernel {
		v := math.Exp(-float64((i-rad)*(i-rad)) / (2 * sigma * sigma))
		kernel[i], sum = v, sum+v
	}
	for i := range kernel {
		kernel[i] /= sum
	}
	pass := func(in []float64, dx, dy int) []float64 {
		out := make([]float64, len(in))
		for y := range h {
			for x := range w {
				var v float64
				for k, kv := range kernel {
					nx, ny := x+(k-rad)*dx, y+(k-rad)*dy
					if nx >= 0 && ny >= 0 && nx < w && ny < h {
						v += kv * in[ny*w+nx]
					}
				}
				out[y*w+x] = v
			}
		}
		return out
	}
	return pass(pass(alpha, 1, 0), 0, 1)
}
