package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
)

func unitByte(v float64) uint8 { return uint8(math.Floor(max(0, min(255, v)) + .5)) }
func unitKey(p color.NRGBA) float64 {
	v := float64(max(p.R, p.G, p.B))
	q := max(0, (float64(min(p.R, p.B))-float64(p.G))/max(1, v))
	t := max(0, min(1, (q-.02)/.08))
	return t * t * t * (6*t*t - 15*t + 10)
}
func unitPlanes(src *image.NRGBA) (*image.NRGBA, *image.NRGBA) {
	n, k := image.NewNRGBA(image.Rect(0, 0, src.Bounds().Dx(), src.Bounds().Dy())), image.NewNRGBA(image.Rect(0, 0, src.Bounds().Dx(), src.Bounds().Dy()))
	for y := range n.Bounds().Dy() {
		for x := range n.Bounds().Dx() {
			p := src.NRGBAAt(src.Bounds().Min.X+x, src.Bounds().Min.Y+y)
			if p.A == 0 {
				continue
			}
			key := unitKey(p)
			energy := key * float64(max(p.R, p.G, p.B))
			gray := unitByte(energy)
			n.SetNRGBA(x, y, color.NRGBA{unitByte((1-key)*float64(p.R) + energy), unitByte((1-key)*float64(p.G) + energy), unitByte((1-key)*float64(p.B) + energy), p.A})
			k.SetNRGBA(x, y, color.NRGBA{gray, gray, gray, p.A})
		}
	}
	return n, k
}
func unitTint(base, mask *image.NRGBA, palette []int, spent bool) *image.NRGBA {
	out := image.NewNRGBA(base.Bounds())
	for y := range out.Bounds().Dy() {
		for x := range out.Bounds().Dx() {
			p := base.NRGBAAt(x, y)
			if p.A == 0 {
				continue
			}
			k := float64(mask.NRGBAAt(x, y).R)
			p.R = unitByte(float64(p.R) + k*(float64(palette[0])/255-1))
			p.G = unitByte(float64(p.G) + k*(float64(palette[1])/255-1))
			p.B = unitByte(float64(p.B) + k*(float64(palette[2])/255-1))
			if spent {
				p.G, p.B = p.R, p.R
			}
			out.SetNRGBA(x, y, p)
		}
	}
	return out
}

// Source and destination coordinates are pixel edges. The (48,49) logical
// anchor registers at terrain origin+(48,40), with top left at origin+(0,-9).
// Keep the entire authored panel and one shared framing value for all six views;
// the default is .8. Never crop to art or rescale individual facings.
func unitAffine(src *image.NRGBA) (*image.NRGBA, error) {
	return unitAffineInset(src, .8)
}

func unitAffineParameters(cell int, inset float64) (float64, connectionPoint) {
	if inset == .8 {
		// Preserve the original constant arithmetic and exact report offsets.
		return .8 * 384 / float64(cell), connectionPoint{38.4, 39.2}
	}
	return inset * 384 / float64(cell), connectionPoint{192 * (1 - inset), 196 * (1 - inset)}
}

func unitAffineInset(src *image.NRGBA, inset float64) (*image.NRGBA, error) {
	cell := src.Bounds().Dx()
	if cell < 384 || cell > 2048 || src.Bounds().Dy() != cell {
		return nil, errors.New("unit affine requires a square source cell of size 384..2048")
	}
	if inset < .8 || inset > 1 || math.IsNaN(inset) || math.IsInf(inset, 0) {
		return nil, errors.New("unit framing inset must be finite and between 0.8 and 1")
	}
	scale, offset := unitAffineParameters(cell, inset)
	return unitAffineTo(src, 384, scale, offset)
}

// unitAffineTo resamples the square panel src onto a size×size canvas:
// destination = offset + scale*source, in pixel edges.
func unitAffineTo(src *image.NRGBA, size int, scale float64, offset connectionPoint) (*image.NRGBA, error) {
	cell := src.Bounds().Dx()
	if err := compoundCheckTapBudget(size, size, scale); err != nil {
		return nil, err
	}
	wx, wy := compoundAxis(size, cell, scale, offset[0], false), compoundAxis(size, cell, scale, offset[1], false)
	tmp := make([][4]float64, size*cell)
	for y := range cell {
		for x, weights := range wx {
			var q [4]float64
			for i, w := range weights.weights {
				p := src.NRGBAAt(src.Bounds().Min.X+weights.first+i, src.Bounds().Min.Y+y)
				a := float64(p.A) * w
				q[0] += a * float64(p.R)
				q[1] += a * float64(p.G)
				q[2] += a * float64(p.B)
				q[3] += a
			}
			tmp[y*size+x] = q
		}
	}
	out := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y, weights := range wy {
		for x := range size {
			var q [4]float64
			for i, w := range weights.weights {
				p := tmp[(weights.first+i)*size+x]
				for c := range q {
					q[c] += p[c] * w
				}
			}
			if q[3] <= 0 {
				continue
			}
			a := unitByte(q[3])
			if a == 0 {
				continue
			}
			out.SetNRGBA(x, y, color.NRGBA{unitByte(q[0] / q[3]), unitByte(q[1] / q[3]), unitByte(q[2] / q[3]), a})
		}
	}
	return out, nil
}

type unitAlphaReport struct {
	Bounds             [4]int  `json:"alpha_bounds_pixels"`
	ClearWorld         float64 `json:"minimum_clear_world_pixels"`
	GuardWorld         int     `json:"required_zero_rgba_guard_world_pixels"`
	ContrastGuardWorld int     `json:"required_alpha_guard_world_pixels,omitempty"`
	GuardAlphaMin      uint8   `json:"guard_minimum_alpha,omitempty"`
	FringeSamples      int     `json:"low_alpha_guard_samples,omitempty"`
	FringeMaximum      uint8   `json:"low_alpha_guard_maximum,omitempty"`
	GuardPolicy        string  `json:"guard_policy,omitempty"`
}

func unitGuard(im *image.NRGBA, density int) (unitAlphaReport, error) {
	return unitGuardAlpha(im, density, 1)
}

// Explicit reference framing retains faint generated alpha fringes unchanged.
// The original strict guard remains the default; reports never call the wider
// frame a zero-RGBA pass when low-alpha samples occupy the perimeter.
func unitGuardAlpha(im *image.NRGBA, density int, minimumAlpha uint8) (unitAlphaReport, error) {
	w, h := im.Bounds().Dx(), im.Bounds().Dy()
	r := unitAlphaReport{Bounds: [4]int{w, h, 0, 0}, GuardWorld: 6}
	if minimumAlpha != 1 {
		r.GuardWorld = 0
		r.ContrastGuardWorld, r.GuardAlphaMin = 6, minimumAlpha
		r.GuardPolicy = "Explicit reference framing: six-world-pixel guard for alpha >=8; lower-alpha source fringes are retained unchanged and counted. Not a zero-RGBA guard pass."
	}
	found := false
	for y := range h {
		for x := range w {
			p := im.NRGBAAt(x, y)
			if p.A == 0 {
				if p != (color.NRGBA{}) {
					return r, errors.New("unit alpha-zero pixel contains hidden RGB")
				}
				continue
			}
			if x < 6*density || y < 6*density || x >= w-6*density || y >= h-6*density {
				if p.A >= minimumAlpha {
					return r, fmt.Errorf("unit art violates six-world-pixel guard at (%d,%d), alpha=%d threshold=%d", x, y, p.A, minimumAlpha)
				}
				r.FringeSamples++
				r.FringeMaximum = max(r.FringeMaximum, p.A)
			}
			found = true
			r.Bounds[0] = min(r.Bounds[0], x)
			r.Bounds[1] = min(r.Bounds[1], y)
			r.Bounds[2] = max(r.Bounds[2], x+1)
			r.Bounds[3] = max(r.Bounds[3], y+1)
		}
	}
	if !found {
		return r, errors.New("unit output is empty")
	}
	r.ClearWorld = float64(min(r.Bounds[0], r.Bounds[1], w-r.Bounds[2], h-r.Bounds[3])) / float64(density)
	return r, nil
}
