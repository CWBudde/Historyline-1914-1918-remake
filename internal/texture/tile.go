// Package texture grades seamless paintings of materials so drawn surfaces
// keep their own colours and take only the painting's texture.
package texture

import (
	"image"
	"image/color"
	"math"
)

// Tile is a graded tile: per texel and channel the ratio to the tile's mean,
// minus one, at the full size and halved down to a few texels.
type Tile struct {
	mips []mip
}

type mip struct {
	w, h int
	d    []float32 // 3 per texel
}

// NewTile grades img, nil for none or an empty image.
func NewTile(img image.Image) *Tile {
	if img == nil || img.Bounds().Empty() {
		return nil
	}
	b := img.Bounds()
	m := mip{b.Dx(), b.Dy(), make([]float32, 3*b.Dx()*b.Dy())}
	var sum [3]float64
	for y := range m.h {
		for x := range m.w {
			c := color.NRGBAModel.Convert(img.At(b.Min.X+x, b.Min.Y+y)).(color.NRGBA)
			i := 3 * (y*m.w + x)
			for ch, v := range [3]uint8{c.R, c.G, c.B} {
				m.d[i+ch] = float32(v)
				sum[ch] += float64(v)
			}
		}
	}
	n := float64(m.w * m.h)
	for i := range m.d {
		if mean := sum[i%3] / n; mean > 0 {
			m.d[i] = float32(float64(m.d[i])/mean - 1)
		} else {
			m.d[i] = 0
		}
	}
	t := &Tile{mips: []mip{m}}
	for m.w >= 8 && m.h >= 8 {
		m = m.half()
		t.mips = append(t.mips, m)
	}
	return t
}

// half is m at half its size, each texel the mean of four.
func (m mip) half() mip {
	h := mip{m.w / 2, m.h / 2, make([]float32, 3*(m.w/2)*(m.h/2))}
	for y := range h.h {
		for x := range h.w {
			for ch := range 3 {
				at := func(dx, dy int) float32 { return m.d[3*((2*y+dy)*m.w+2*x+dx)+ch] }
				h.d[3*(y*h.w+x)+ch] = (at(0, 0) + at(1, 0) + at(0, 1) + at(1, 1)) / 4
			}
		}
	}
	return h
}

// At is the tile's deviation at device pixel (x, y) when one tile spans span
// device pixels: the tile wraps, sampled bilinearly from the level whose
// texels are no smaller than a device pixel.
func (t *Tile) At(x, y int, span float64) (dr, dg, db float64) {
	m := t.mips[0]
	for _, next := range t.mips[1:] {
		if float64(m.w)/span <= 1.5 {
			break
		}
		m = next
	}
	f := float64(m.w) / span
	u, v := (float64(x)+0.5)*f-0.5, (float64(y)+0.5)*f-0.5
	x0, y0 := math.Floor(u), math.Floor(v)
	fx, fy := u-x0, v-y0
	wrap := func(i, n int) int { return ((i % n) + n) % n }
	var d [3]float64
	for _, c := range [4]struct {
		dx, dy int
		w      float64
	}{{0, 0, (1 - fx) * (1 - fy)}, {1, 0, fx * (1 - fy)}, {0, 1, (1 - fx) * fy}, {1, 1, fx * fy}} {
		i := 3 * (wrap(int(y0)+c.dy, m.h)*m.w + wrap(int(x0)+c.dx, m.w))
		for ch := range 3 {
			d[ch] += c.w * float64(m.d[i+ch])
		}
	}
	return d[0], d[1], d[2]
}

// Modulate is c with each channel scaled by one plus gain times its
// deviation.
func Modulate(c color.RGBA, gain, dr, dg, db float64) color.RGBA {
	m := func(v uint8, d float64) uint8 {
		return uint8(math.Max(0, math.Min(255, math.Round(float64(v)*(1+gain*d)))))
	}
	return color.RGBA{m(c.R, dr), m(c.G, dg), m(c.B, db), c.A}
}
