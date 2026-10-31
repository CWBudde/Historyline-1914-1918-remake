package unittopview

import (
	"fmt"
	"image"
	"math"
)

// Owner masters are edits of the base that add the insignia, but an edit
// redraws the whole picture: its outline and roof tone drift by a few pixels and
// levels. Only the centre disc of an edit is used, registered on the base and
// matched to its tone, so every owner keeps the base's outline.

// lumAt is the luma of an NRGBA pixel.
func lumAt(im *image.NRGBA, x, y int) float64 {
	p := im.Pix[im.PixOffset(x, y):]
	return 0.299*float64(p[0]) + 0.587*float64(p[1]) + 0.114*float64(p[2])
}

// register finds the shift (dx, dy) within ±reach that best lays edit onto base:
// the smallest mean absolute luma gradient difference over the annulus
// [r0, r1) around (cx, cy), which holds the disc's edge and the roof's seams.
func register(base, edit *image.NRGBA, cx, cy, r0, r1, sq float64, reach int) (int, int) {
	grad := func(im *image.NRGBA, x, y int) float64 {
		return math.Abs(lumAt(im, x+1, y)-lumAt(im, x-1, y)) + math.Abs(lumAt(im, x, y+1)-lumAt(im, x, y-1))
	}
	b := base.Bounds().Intersect(edit.Bounds()).Inset(reach + 1)
	type pt struct{ x, y int }
	var ring []pt
	for y := int(cy - r1*sq); y <= int(cy+r1*sq); y += 2 {
		for x := int(cx - r1); x <= int(cx+r1); x += 2 {
			d := math.Hypot(float64(x)+.5-cx, (float64(y)+.5-cy)/sq)
			if d >= r0 && d < r1 && image.Pt(x, y).In(b) {
				ring = append(ring, pt{x, y})
			}
		}
	}
	want := make([]float64, len(ring))
	for i, p := range ring {
		want[i] = grad(base, p.x, p.y)
	}
	bestX, bestY, best := 0, 0, math.Inf(1)
	for dy := -reach; dy <= reach; dy++ {
		for dx := -reach; dx <= reach; dx++ {
			var sum float64
			for i, p := range ring {
				sum += math.Abs(grad(edit, p.x+dx, p.y+dy) - want[i])
			}
			// Ties keep the smaller shift, so an exact copy registers at (0, 0).
			if sum < best-1e-9 {
				bestX, bestY, best = dx, dy, sum
			}
		}
	}
	return bestX, bestY
}

// inlay replaces the disc of radius r around (cx, cy) of base with the
// registered disc of edit, scaled to the base's mean concrete luma just inside the disc
// edge, blended over the last feather pixels. The note reports shift and gain.
func inlay(base, edit *image.NRGBA, cx, cy, r, sq, feather float64) (*image.NRGBA, string) {
	dx, dy := register(base, edit, cx, cy, r-20, r+50, sq, 10)
	out := image.NewNRGBA(base.Bounds())
	copy(out.Pix, base.Pix)
	// The tone is matched on bare concrete only: grey in the edit and near the
	// base's luma, so paint that reaches the edge does not count.
	var lb, le, n float64
	for y := int(cy - r*sq); y <= int(cy+r*sq); y++ {
		for x := int(cx - r); x <= int(cx+r); x++ {
			if d := math.Hypot(float64(x)+.5-cx, (float64(y)+.5-cy)/sq); d < r-2*feather || d >= r {
				continue
			}
			p := edit.Pix[edit.PixOffset(x+dx, y+dy):]
			chroma := int(max(p[0], p[1], p[2])) - int(min(p[0], p[1], p[2]))
			b, e := lumAt(base, x, y), lumAt(edit, x+dx, y+dy)
			if chroma < 30 && math.Abs(e-b) < 35 {
				lb, le, n = lb+b, le+e, n+1
			}
		}
	}
	gain := 1.0
	if n > 0 && le > 0 {
		gain = lb / le
	}
	for y := int(cy - r*sq - 1); y <= int(cy+r*sq+1); y++ {
		for x := int(cx - r - 1); x <= int(cx+r+1); x++ {
			d := math.Hypot(float64(x)+.5-cx, (float64(y)+.5-cy)/sq)
			w := math.Max(0, math.Min(1, (r-d)/feather))
			if w == 0 {
				continue
			}
			w = w * w * (3 - 2*w)
			i, j := out.PixOffset(x, y), edit.PixOffset(x+dx, y+dy)
			for ch := range 3 {
				v := float64(base.Pix[i+ch])*(1-w) + math.Min(255, float64(edit.Pix[j+ch])*gain)*w
				out.Pix[i+ch] = uint8(math.Round(v))
			}
		}
	}
	return out, fmt.Sprintf("shift (%d,%d), luma gain %.3f", dx, dy, gain)
}
