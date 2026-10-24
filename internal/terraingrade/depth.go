package terraingrade

import (
	"image"
	"math"
)

// The original maps tell three depths of open water apart, shallow water, a
// middle band and the open sea, and draw each a step darker: L* falls by 7 %
// and then by 18 %, which calms the ripples by as much, and the teal turns a
// little less blue (a* to 0.93, b* to 0.66 in the open sea). Deepen gives a
// graded tile the same steps by depth, 0 (shallow) to 1 (open sea).
var (
	deepL = curve{{0, 1}, {0.5, 0.93}, {1, 0.823}}
	deepA = curve{{0, 1}, {0.5, 1}, {1, 0.925}}
	deepB = curve{{0, 1}, {0.5, 0.86}, {1, 0.663}}
)

// Blend mixes graded tile sea, the open sea's material at the same place and
// size, into the water of graded tile im in place by weight(x, y), 0 to 1.
// Only open water takes it: a pixel the graded open water tile water (same
// place and size, nil for none) has too, ripple crests included, and elsewhere
// by how blue and dark a pixel is against the sea's own water, so banks,
// stones and snow keep their pixels. The mix keeps each channel's spread about
// the sea's mean, so the two ripple patterns do not wash each other out where
// they meet.
func Blend(im, sea, water *image.NRGBA, weight func(x, y int) float64) {
	b := im.Rect
	var mean [3]float64
	var wetness, n float64
	for y := range b.Dy() {
		for x := range b.Dx() {
			i := sea.PixOffset(sea.Rect.Min.X+x, sea.Rect.Min.Y+y)
			s := sea.Pix[i : i+4 : i+4]
			if s[3] == 0 {
				continue
			}
			l, _, bb := toLab(s[0], s[1], s[2])
			wetness += wet(l, bb)
			for k := range 3 {
				mean[k] += float64(s[k])
			}
			n++
		}
	}
	if n == 0 || wetness == 0 {
		return
	}
	for k := range mean {
		mean[k] /= n
	}
	wetness /= n
	rowsOf(b.Dy(), func(y int) {
		for x := range b.Dx() {
			i := im.PixOffset(b.Min.X+x, b.Min.Y+y)
			p := im.Pix[i : i+4 : i+4]
			j := sea.PixOffset(sea.Rect.Min.X+x, sea.Rect.Min.Y+y)
			s := sea.Pix[j : j+4 : j+4]
			if p[3] == 0 || s[3] == 0 {
				continue
			}
			t := weight(b.Min.X+x, b.Min.Y+y)
			if t <= 0 {
				continue
			}
			open := false
			if water != nil {
				w := water.PixOffset(water.Rect.Min.X+x, water.Rect.Min.Y+y)
				open = [4]uint8(p) == [4]uint8(water.Pix[w:w+4])
			}
			if !open {
				l, _, bb := toLab(p[0], p[1], p[2])
				if t *= clamp(wet(l, bb) / wetness); t <= 0 {
					continue
				}
			}
			k := 1 / math.Hypot(1-t, t)
			for c := range 3 {
				v := mean[c] + k*((1-t)*(float64(p[c])-mean[c])+t*(float64(s[c])-mean[c]))
				p[c] = uint8(math.Round(math.Max(0, math.Min(255, v))))
			}
		}
	})
}

// Deepen shades the water of graded tile im in place by depth(x, y) at each
// pixel, 0 (shallow) to 1 (open sea), ripple crests with it. Snow and foam
// (L* 80 and above) fade out of it by L* 90 and keep their colour.
func Deepen(im *image.NRGBA, depth func(x, y int) float64) {
	b := im.Rect
	rowsOf(b.Dy(), func(y int) {
		for x := range b.Dx() {
			i := im.PixOffset(b.Min.X+x, b.Min.Y+y)
			p := im.Pix[i : i+4 : i+4]
			if p[3] == 0 {
				continue
			}
			d := depth(b.Min.X+x, b.Min.Y+y)
			if d <= 0 {
				continue
			}
			l, a, bb := toLab(p[0], p[1], p[2])
			w := clamp((90 - l) / 10)
			if w == 0 {
				continue
			}
			k := func(c curve) float64 { return 1 + w*(c.at(d)-1) }
			c := fromLab(l*k(deepL), a*k(deepA), bb*k(deepB), p[3])
			p[0], p[1], p[2] = c.R, c.G, c.B
		}
	})
}

// Wet is how much a pixel of colour r, g, b is open water, 0 to 1: blue and
// not light, the test the grade gives water its own curve by.
func Wet(r, g, b uint8) float64 {
	l, _, bb := toLab(r, g, b)
	return wet(l, bb)
}
