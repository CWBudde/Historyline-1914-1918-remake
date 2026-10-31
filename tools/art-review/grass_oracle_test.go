package artreview

import (
	"image"
	"image/color"
	"math"
)

// This independent oracle expands the separable overlap into weighted source
// taps, rather than calling the production sampler or its window helpers.
func groundTestGrassTexture(src *image.NRGBA, x, y float64) color.NRGBA {
	type tap struct{ at, weight float64 }
	axis := func(v, period float64, size int) []tap {
		phase := v - period*math.Floor(v/period)
		phase = math.Round(phase*1e9) / 1e9
		if phase == period {
			phase = 0
		}
		span := .75 * float64(size-1)
		at := phase / period * span
		overlap := float64(size-1) / 4
		if at >= overlap {
			return []tap{{at, 1}}
		}
		u := at / overlap
		w := u * u * (3 - 2*u)
		return []tap{{at, w}, {at + span, 1 - w}}
	}
	var rgb [3]float64
	for _, shift := range [][2]float64{{0, 0}, {71, 40}} {
		for _, xt := range axis(x+shift[0], 142, src.Bounds().Dx()) {
			for _, yt := range axis(y+shift[1], 80, src.Bounds().Dy()) {
				ix, iy := int(xt.at), int(yt.at)
				fx, fy := xt.at-float64(ix), yt.at-float64(iy)
				for dy := 0; dy < 2; dy++ {
					for dx := 0; dx < 2; dx++ {
						wx, wy := 1-fx, 1-fy
						if dx == 1 {
							wx = fx
						}
						if dy == 1 {
							wy = fy
						}
						c := src.NRGBAAt(src.Bounds().Min.X+min(ix+dx, src.Bounds().Dx()-1), src.Bounds().Min.Y+min(iy+dy, src.Bounds().Dy()-1))
						w := .5 * xt.weight * yt.weight * wx * wy
						rgb[0] += float64(c.R) * w
						rgb[1] += float64(c.G) * w
						rgb[2] += float64(c.B) * w
					}
				}
			}
		}
	}
	return color.NRGBA{uint8(math.Floor(rgb[0] + .5000000001)), uint8(math.Floor(rgb[1] + .5000000001)), uint8(math.Floor(rgb[2] + .5000000001)), 255}
}
