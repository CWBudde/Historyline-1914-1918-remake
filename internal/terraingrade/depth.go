package terraingrade

import "image"

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
