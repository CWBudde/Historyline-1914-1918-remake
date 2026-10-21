package main

import (
	"image"
	"image/color"
	"math"
)

const authoredBankProfile = "authored-rocks-v1"
const authoredBankDescription = " Authored-rocks-v1 override: single unmirrored authored sample, no coset average. Centre-crop to71:40, bilinear sample with8percent opposing-edge overlap only; bank periods35.5x20worldpx, water71x40. Each period divides every lattice translation exactly. Native1254px bank boulders remain about2..5logicalpx. Shore full rock band extends landward to5.2px then blends through8.4px; stream adds2px landward bank, modulated by authored rock brightness, through the existing8..16world-pixel interior envelope. Stream outer footprint and bank opacity retain the legacy profile throughout every8world-pixel port guard. Water core/ports/shore vertex topology unchanged; no synthesized stones or continuous foam line. Legacy profiles unchanged."

// The smaller rectangular period is already invariant under all six hex-lattice
// translations. Unlike coset averaging, it does not merge unrelated rock faces.
// A uniform centre crop preserves rounded authored shapes at the chosen scale.
func sampleAuthoredMaterial(src *image.NRGBA, p, period connectionPoint) color.NRGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	ratio := period[0] / period[1]
	if float64(w)/float64(h) > ratio {
		w = int(math.Floor(float64(h) * ratio))
	} else {
		h = int(math.Floor(float64(w) / ratio))
	}
	ox, oy := b.Min.X+(b.Dx()-w)/2, b.Min.Y+(b.Dy()-h)/2
	axis := func(v, period float64, size int) (float64, float64, float64) {
		v = math.Mod(v, period)
		if v < 0 {
			v += period
		}
		v = math.Round(v*1e9) / 1e9
		if v == period {
			v = 0
		}
		overlap := .08 * float64(size-1)
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
		out := [3]float64{}
		for dy := range 2 {
			for dx := range 2 {
				wx, wy := 1-fx, 1-fy
				if dx == 1 {
					wx = fx
				}
				if dy == 1 {
					wy = fy
				}
				c := src.NRGBAAt(ox+min(ix+dx, w-1), oy+min(iy+dy, h-1))
				out[0] += float64(c.R) * wx * wy
				out[1] += float64(c.G) * wx * wy
				out[2] += float64(c.B) * wx * wy
			}
		}
		return out
	}
	xa, xb, wx := axis(p[0], period[0], w)
	ya, yb, wy := axis(p[1], period[1], h)
	a, b0, c, d := bilinear(xa, ya), bilinear(xb, ya), bilinear(xa, yb), bilinear(xb, yb)
	out := [3]uint8{}
	for i := range 3 {
		top := b0[i] + (a[i]-b0[i])*wx
		bottom := d[i] + (c[i]-d[i])*wx
		out[i] = uint8(math.Floor(bottom + (top-bottom)*wy + .5))
	}
	return color.NRGBA{out[0], out[1], out[2], 255}
}
func sampleAuthoredBank(src *image.NRGBA, p connectionPoint) color.NRGBA {
	return sampleAuthoredMaterial(src, p, connectionPoint{35.5, 20})
}
func sampleAuthoredWater(src *image.NRGBA, p connectionPoint) color.NRGBA {
	return sampleAuthoredMaterial(src, p, connectionPoint{71, 40})
}
func authoredRockOffset(rock color.NRGBA) float64 {
	return (float64(rock.R)+float64(rock.G)+float64(rock.B))/765 - .45
}
func authoredShoreColor(in shoreInputs, p connectionPoint, d float64) color.NRGBA {
	water := sampleAuthoredWater(in.Materials["water"].Image, p)
	if d <= -7 {
		return water
	}
	grass := sampleGrassMaterial(in.Materials["grass"].Image, p)
	if d >= 10 {
		return grass
	}
	rock := sampleAuthoredBank(in.Materials["bank"].Image, p)
	d += 1.4 * authoredRockOffset(rock)
	wet := connectionBlend(rock, color.NRGBA{34, 52, 47, 255}, .56)
	c := connectionBlend(water, wet, connectionSmoothStep((d+1.1)/1.1))
	c = connectionBlend(c, rock, connectionSmoothStep((d-.1)/.8))
	return connectionBlend(c, grass, connectionSmoothStep((d-5.2)/3.2))
}
func authoredStreamColor(in connectionInputs, f connectionFamily, p connectionPoint, d float64) color.NRGBA {
	water := sampleAuthoredWater(in.materials[f.Material], p)
	rock := sampleAuthoredBank(in.materials[f.BorderMaterial], p)
	d = math.Round(d*1e9) / 1e9
	wet := connectionBlend(rock, color.NRGBA{34, 52, 47, 255}, .56)
	c := connectionBlend(water, wet, connectionSmoothStep((d+.45)/.7))
	return connectionBlend(c, rock, connectionSmoothStep((d-.15)/.8))
}
