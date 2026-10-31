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
	periodic := func(q connectionPoint) [3]float64 {
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
	a, b := periodic(p), periodic(p.add(connectionPoint{71, 40}))
	round := func(v float64) uint8 { return uint8(math.Floor(v + .5000000001)) }
	return color.NRGBA{round((a[0] + b[0]) / 2), round((a[1] + b[1]) / 2), round((a[2] + b[2]) / 2), 255}
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
