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
// Keep the entire authored panel and one shared .8 framing; never crop to art.
func unitAffine(src *image.NRGBA) (*image.NRGBA, error) {
	cell := src.Bounds().Dx()
	if cell < 384 || cell > 2048 || src.Bounds().Dy() != cell {
		return nil, errors.New("unit affine requires a square source cell of size 384..2048")
	}
	scale := .8 * 384 / float64(cell)
	if err := compoundCheckTapBudget(384, 384, scale); err != nil {
		return nil, err
	}
	wx, wy := compoundAxis(384, cell, scale, 38.4), compoundAxis(384, cell, scale, 39.2)
	tmp := make([][4]float64, 384*cell)
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
			tmp[y*384+x] = q
		}
	}
	out := image.NewNRGBA(image.Rect(0, 0, 384, 384))
	for y, weights := range wy {
		for x := range 384 {
			var q [4]float64
			for i, w := range weights.weights {
				p := tmp[(weights.first+i)*384+x]
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
	Bounds     [4]int  `json:"alpha_bounds_pixels"`
	ClearWorld float64 `json:"minimum_clear_world_pixels"`
	GuardWorld int     `json:"required_zero_rgba_guard_world_pixels"`
}

func unitGuard(im *image.NRGBA, density int) (unitAlphaReport, error) {
	w, h := im.Bounds().Dx(), im.Bounds().Dy()
	r := unitAlphaReport{Bounds: [4]int{w, h, 0, 0}, GuardWorld: 6}
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
				return r, fmt.Errorf("unit art violates six-world-pixel guard at (%d,%d)", x, y)
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
