package dospanel

import (
	"image"
	"image/color"
	"math"
)

// Surfaces are tileable paintings of the plates' materials: the steel of
// the panels and the ground of the battle's hexes. Only their texture is
// used: each texel's colour relative to the tile's mean modulates the drawn
// colour, so the plates keep the palette's values. Nil keeps the drawn
// grain.
type Surfaces struct {
	Steel, Ground image.Image
}

// The reference pixels one tile of each surface spans, and how strongly
// its texture shows.
const (
	steelSpan  = 256
	groundSpan = 96
	steelGain  = 0.6
	groundGain = 0.8
)

// SetSurfaces makes s the plates' surfaces and repaints the plates.
func SetSurfaces(s Surfaces) {
	steel, ground := newSurface(s.Steel), newSurface(s.Ground)
	plates.Lock()
	defer plates.Unlock()
	surfaces.steel, surfaces.ground = steel, ground
	for k, img := range plates.m {
		img.Deallocate()
		delete(plates.m, k)
	}
}

// surfaces are the graded tiles in use, guarded by plates.
var surfaces struct{ steel, ground *surface }

// surface is a graded tile: per texel and channel the ratio to the tile's
// mean, minus one, at the full size and halved down to a few texels.
type surface struct {
	mips []mip
}

type mip struct {
	w, h int
	d    []float32 // 3 per texel
}

// newSurface grades img, nil for none or an empty image.
func newSurface(img image.Image) *surface {
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
	s := &surface{mips: []mip{m}}
	for m.w >= 8 && m.h >= 8 {
		m = m.half()
		s.mips = append(s.mips, m)
	}
	return s
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

// at is the surface's deviation at device pixel (x, y) of a plate at k
// device pixels per reference pixel when a tile spans span reference
// pixels: the tile wraps, sampled bilinearly from the level whose texels are
// no smaller than a device pixel.
func (s *surface) at(x, y int, k, span float64) (dr, dg, db float64) {
	m := s.mips[0]
	for _, next := range s.mips[1:] {
		if float64(m.w)/(span*k) <= 1.5 {
			break
		}
		m = next
	}
	f := float64(m.w) / (span * k)
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

// modulate is c with each channel scaled by one plus gain times its
// deviation.
func modulate(c color.RGBA, gain, dr, dg, db float64) color.RGBA {
	m := func(v uint8, d float64) uint8 {
		return uint8(math.Max(0, math.Min(255, math.Round(float64(v)*(1+gain*d)))))
	}
	return color.RGBA{m(c.R, dr), m(c.G, dg), m(c.B, db), c.A}
}

// steelColour is the steel of tone t at device pixel (x, y): the painted
// surface over the ramp, or the drawn grain without one.
func (p painter) steelColour(x, y int, t float64) color.RGBA {
	if s := surfaces.steel; s != nil {
		dr, dg, db := s.at(x, y, p.k, steelSpan)
		return modulate(ramp(t), steelGain, dr, dg, db)
	}
	q := max(1, int(math.Round(p.k)))
	return ramp(t + 0.06*grain(x/q, y/q))
}

// groundColour is the hex ground of tone t at device pixel (x, y).
func (p painter) groundColour(x, y int, t float64) color.RGBA {
	if s := surfaces.ground; s != nil {
		dr, dg, db := s.at(x, y, p.k, groundSpan)
		return modulate(lerp(fieldDark, fieldLight, t), groundGain, dr, dg, db)
	}
	q := max(1, int(math.Round(p.k)))
	return lerp(fieldDark, fieldLight, t+0.05*grain(x/q, y/q)+0.03*grain(x/(4*q), y/(3*q)))
}

// blend sets (x, y) to c over the pixel there at coverage a.
func (p painter) blend(x, y int, c color.RGBA, a float64) {
	if a >= 1 {
		p.img.SetRGBA(x, y, c)
		return
	}
	o := p.img.RGBAAt(x, y)
	p.img.SetRGBA(x, y, lerp(o, c, a))
}
