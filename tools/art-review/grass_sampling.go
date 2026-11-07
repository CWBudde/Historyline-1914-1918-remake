package artreview

import (
	"image"
	"image/color"
	"math"
)

const grassSamplingProfile = "Summer grass: unmirrored periodic bilinear source sampling. Per source axis, overlap=(size-1)/4 and span=3*(size-1)/4. At world phase t in [0,period), x=t/period*span. For x<overlap blend source(x+span) toward source(x) with cubic smoothstep(x/overlap); otherwise source(x). Apply separable windows on both axes, periods142,80. Round wrapped world phases to1e-9 before source lookup. Average RGB of f(x,y),f(x+71,y+40), then round half-up once. This is invariant under all six neighbor translations without reflected source marks. Non-grass materials, unseen and winter retain the existing sampler."

// sampleGrassMaterial wraps the authored texture without reflecting its marks.
// A quarter-source overlap crossfades each opposing edge with a cubic window.
// Averaging the two (71,40) cosets makes the result exactly lattice-periodic,
// with rectangular periods (142,80), without introducing mirror ornaments.
func sampleGrassMaterial(src *image.NRGBA, p connectionPoint) color.NRGBA {
	a, b := grassPeriodicRGB(src, p), grassPeriodicRGB(src, p.add(connectionPoint{71, 40}))
	return color.NRGBA{grassRound((a[0] + b[0]) / 2), grassRound((a[1] + b[1]) / 2), grassRound((a[2] + b[2]) / 2), 255}
}

func grassRound(v float64) uint8 { return uint8(math.Floor(v + .5000000001)) }

// grassPeriodicRGB is one coset's unmirrored sample, with rectangular periods
// (142,80) only.
func grassPeriodicRGB(src *image.NRGBA, q connectionPoint) [3]float64 {
	axis := func(v, period float64, size int) (float64, float64, float64) {
		v = math.Mod(v, period)
		if v < 0 {
			v += period
		}
		v = math.Round(v*1e9) / 1e9
		if v == period {
			v = 0
		}
		overlap := float64(size-1) / 4
		span := float64(size-1) - overlap
		x := v / period * span
		if x < overlap {
			return x, x + span, connectionSmoothStep(x / overlap)
		}
		return x, x, 1
	}
	bilinear := func(x, y float64) [3]float64 {
		ix, iy := int(x), int(y)
		fx, fy := x-float64(ix), y-float64(iy)
		var rgb [3]float64
		for dy := range 2 {
			for dx := range 2 {
				wx, wy := 1-fx, 1-fy
				if dx == 1 {
					wx = fx
				}
				if dy == 1 {
					wy = fy
				}
				c := src.NRGBAAt(src.Bounds().Min.X+min(ix+dx, src.Bounds().Dx()-1), src.Bounds().Min.Y+min(iy+dy, src.Bounds().Dy()-1))
				rgb[0] += float64(c.R) * wx * wy
				rgb[1] += float64(c.G) * wx * wy
				rgb[2] += float64(c.B) * wx * wy
			}
		}
		return rgb
	}
	xa, xb, wx := axis(q[0], 142, src.Bounds().Dx())
	ya, yb, wy := axis(q[1], 80, src.Bounds().Dy())
	aa := bilinear(xa, ya)
	// Outside the wrap windows the coordinates coincide. Avoid fetching
	// those identical samples for every grass pixel in every family.
	if wx == 1 && wy == 1 {
		return aa
	}
	var out [3]float64
	if wx == 1 {
		ab := bilinear(xa, yb)
		for i := range out {
			out[i] = wy*aa[i] + (1-wy)*ab[i]
		}
		return out
	}
	if wy == 1 {
		ba := bilinear(xb, ya)
		for i := range out {
			out[i] = wx*aa[i] + (1-wx)*ba[i]
		}
		return out
	}
	ab, ba, bb := bilinear(xa, yb), bilinear(xb, ya), bilinear(xb, yb)
	for i := range out {
		out[i] = wy*(wx*aa[i]+(1-wx)*ba[i]) + (1-wy)*(wx*ab[i]+(1-wx)*bb[i])
	}
	return out
}

// Selecting between the two cosets instead of averaging them keeps the
// painted contrast. The weight field flips under (71,40) while the cosets swap,
// so the result stays exactly lattice-periodic. The noise holds only harmonics
// with k+m odd, which change sign under that shift; the stone key, on luminance
// blurred over a few world px so that whole patches switch and not single grains,
// lets the darker coset win near a seam: seams run around stones, not through them.
const (
	latticeSelectStoneKey = .9  // weight of the luminance difference against the noise
	latticeSelectSoftness = .10 // half-width of the weight's ramp
	latticeSelectBlur     = 2.5 // world px radius of the stone key's luminance
)

var latticeSelectHarmonics = [...]struct{ k, m, amp, phase float64 }{
	{1, 0, .30, .4}, {0, 1, .30, 2.1}, {2, 1, .28, 4.0}, {1, 2, .28, 5.3}, {3, 2, .18, 1.2}, {2, 3, .18, 3.1},
}

func latticeSelectNoise(p connectionPoint) float64 {
	n := 0.
	for _, h := range latticeSelectHarmonics {
		n += h.amp * math.Sin(2*math.Pi*(h.k*p[0]/142+h.m*p[1]/80)+h.phase)
	}
	return n
}

// latticeSelectLuma is one coset's luminance in 0..1, blurred over a ring of
// latticeSelectBlur around p.
func latticeSelectLuma(src *image.NRGBA, p connectionPoint) float64 {
	sum := 0.
	for i := range 8 {
		s, c := math.Sincos(float64(i) * math.Pi / 4)
		rgb := grassPeriodicRGB(src, p.add(connectionPoint{latticeSelectBlur * c, latticeSelectBlur * s}))
		sum += .2126*rgb[0] + .7152*rgb[1] + .0722*rgb[2]
	}
	return sum / (8 * 255)
}

// sampleSelectedMaterial is sampleGrassMaterial with the cosets selected by the
// antisymmetric weight field instead of averaged.
func sampleSelectedMaterial(src *image.NRGBA, p connectionPoint) color.NRGBA {
	a, b := grassPeriodicRGB(src, p), grassPeriodicRGB(src, p.add(connectionPoint{71, 40}))
	n := latticeSelectNoise(p) + latticeSelectStoneKey*(latticeSelectLuma(src, p.add(connectionPoint{71, 40}))-latticeSelectLuma(src, p))
	w := connectionSmoothStep(.5 + n/(2*latticeSelectSoftness))
	mix := func(i int) uint8 { return grassRound(w*a[i] + (1-w)*b[i]) }
	return color.NRGBA{mix(0), mix(1), mix(2), 255}
}

func renderGrassBaseRaw(src *image.NRGBA) *image.NRGBA {
	raw := image.NewNRGBA(image.Rect(0, 0, 416, 352))
	for y := range 352 {
		for x := range 416 {
			p := connectionPoint{(float64(x)+.5)/4 - 4, (float64(y)+.5)/4 - 4}
			raw.SetNRGBA(x, y, sampleGrassMaterial(src, p))
		}
	}
	return raw
}

func exportGrassBaseTiles(src *image.NRGBA) (map[int]*image.NRGBA, error) {
	return grassTilesFromRaw(renderGrassBaseRaw(src))
}
