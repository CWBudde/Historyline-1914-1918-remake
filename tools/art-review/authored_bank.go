package main

import (
	"image"
	"image/color"
	"math"
)

const authoredBankProfile = "authored-rocks-v1"
const authoredBankDescription = " Authored-rocks-v1 materials: single unmirrored authored sample, no coset average. Centre-crop to71:40, bilinear sample with8percent opposing-edge overlap only; bank periods35.5x20worldpx, water71x40. Each period divides every lattice translation exactly. " +
	"Shore interiors: E=smoothstep((nearestEdge-10)/10), P=.7*sin(.065*x+.105*y+.8)+.3*sin(.14*x-.075*y+1.3), R=meanRockRGB/255-.45. Coastoffset(1.4+1.8*E)*R, rockreach=clamp(5.2+E*(.6+5.2*P+1.2*R),2.2,11.5), basefeather=3.2+E*(.3-.7*P). Width fields retain5.2px rockreach and3.2px basefeather through10worldpx. " +
	"Soft shore materials: water shows through submerged rocks fromd=-4.8; wetbluegrey faces blend to dryrock over2.5px without a continuous darkoutline. Meadowspan=basefeather+3.6 centred on the accepted outerboundary, with paintedrockcrevices and grassmarks controlling localmosspenetration. Sharedcoastprofiles soften consistently at every reciprocal edge. No blur; sourcepixels, coasts and widthfields unchanged. " +
	"Streams retain their existing bank: add2px landward, modulated by authored rock brightness, through the8..16worldpx interior envelope. Stream outer footprint and bank opacity retain the legacy profile throughout every8worldpx port guard. Water core/ports/shore vertex topology unchanged; no synthesized stones or continuous foam line."

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

// Leave the eight-pixel shared edge profile and its Lanczos filter support
// untouched. Only the tile interior grows broader rock shelves and narrows
// back to small stony lips; the material and world scale remain shared.
func authoredShoreVariation(p connectionPoint) (envelope, patch float64) {
	edge := math.Inf(1)
	for i := range 6 {
		edge = min(edge, shoreMouthEdgeDistance(p, i))
	}
	envelope = connectionSmoothStep((edge - 10) / 10)
	patch = .7*math.Sin(.065*p[0]+.105*p[1]+.8) + .3*math.Sin(.14*p[0]-.075*p[1]+1.3)
	return envelope, patch
}

func authoredShoreColor(in shoreInputs, p connectionPoint, d float64) color.NRGBA {
	water := sampleAuthoredWater(in.Materials["water"].Image, p)
	if d <= -7 {
		return water
	}
	grass := sampleGrassMaterial(in.Materials["grass"].Image, p)
	if d >= 18 {
		return grass
	}
	rock := sampleAuthoredBank(in.Materials["bank"].Image, p)
	envelope, patch := authoredShoreVariation(p)
	stone := authoredRockOffset(rock)
	d += (1.4 + 1.8*envelope) * stone
	width := 5.2 + envelope*(.6+5.2*patch+1.2*stone)
	width = max(2.2, min(11.5, width))
	feather := 3.2 + envelope*(.3-.7*patch)
	// Water still shows through the submerged stones before the wet faces
	// become dry rock. Broader colour ramps avoid a dark cut-out outline;
	// the authored surfaces themselves remain sharp and unfiltered.
	submerged := connectionBlend(water, rock, .24)
	wet := connectionBlend(connectionBlend(rock, water, .52), color.NRGBA{34, 52, 47, 255}, .16)
	c := connectionBlend(water, submerged, connectionSmoothStep((d+4.8)/3.6))
	c = connectionBlend(c, wet, connectionSmoothStep((d+2.2)/2.6))
	c = connectionBlend(c, rock, connectionSmoothStep((d+.1)/2.5))
	if in.Manifest.Season == "winter" {
		// Dry stones turn cold grey and carry snow on their light tops; the
		// wet faces at the waterline stay dark.
		lum := float64(rock.R) + float64(rock.G) + float64(rock.B)
		grey := uint8(lum / 3)
		dry := connectionSmoothStep((d - .5) / 2.5)
		c = connectionBlend(c, color.NRGBA{grey, grey, grey + 4, 255}, .45*dry)
		c = connectionBlend(c, grass, .85*dry*connectionSmoothStep((lum/765-.38)/.22))
	}
	// Centre the wider fringe on the accepted rock/grass boundary. Existing
	// painted crevices and meadow marks let vegetation enter between stones,
	// instead of fading every rock through an identical parallel band.
	crevice := 1 - connectionSmoothStep((stone+.22)/.42)
	grassMark := max(0, min(1, (float64(grass.G)-float64(grass.R))/96))
	fringe := connectionSmoothStep((d - width + 1.8 + .6*(crevice-.5) + .25*(grassMark-.5)) / (feather + 3.6))
	rooted := .8*crevice + .2*grassMark
	c = connectionBlend(c, grass, .3*rooted*4*fringe*(1-fringe))
	return connectionBlend(c, grass, fringe)
}
func authoredStreamColor(in connectionInputs, f connectionFamily, p connectionPoint, d float64) color.NRGBA {
	water := sampleAuthoredWater(in.materials[f.Material], p)
	rock := sampleAuthoredBank(in.materials[f.BorderMaterial], p)
	d = math.Round(d*1e9) / 1e9
	wet := connectionBlend(rock, color.NRGBA{34, 52, 47, 255}, .56)
	c := connectionBlend(water, wet, connectionSmoothStep((d+.45)/.7))
	return connectionBlend(c, rock, connectionSmoothStep((d-.15)/.8))
}
