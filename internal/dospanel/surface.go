package dospanel

import (
	"image"
	"image/color"
	"math"
)

// Surfaces are paintings of the plates' materials: tiles of the steel of
// the panels and the ground of the battle's hexes, and a sheet of rivet
// heads. Only their texture is used: each texel's colour relative to the
// tile's or the heads' mean modulates the palette's colour, so the plates
// keep its values. Nil keeps the drawn grain or rivet.
type Surfaces struct {
	Steel, Ground image.Image
	// Rivets holds rivetCols×rivetRows round heads lit from the top left,
	// one in each cell on a clear ground.
	Rivets image.Image
}

// The rivet sheet's grid.
const rivetCols, rivetRows = 4, 2

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
	steel, ground, rivets := newSurface(s.Steel), newSurface(s.Ground), newRivetSheet(s.Rivets)
	plates.Lock()
	defer plates.Unlock()
	surfaces.steel, surfaces.ground, surfaces.rivets = steel, ground, rivets
	for k, img := range plates.m {
		img.Deallocate()
		delete(plates.m, k)
	}
}

// surfaces are the graded paintings in use, guarded by plates.
var surfaces struct {
	steel, ground *surface
	rivets        *rivetSheet
}

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

// rivetSheet is a sheet of rivet heads: the colour and coverage of each
// texel, each head's place, and the heads' mean colour.
type rivetSheet struct {
	w, h  int
	d     []float32 // per texel r, g, b (not premultiplied) and the cover above half
	heads []rivetHead
	mean  [3]float64
}

// rivetHead is a head's centre and radius in texels.
type rivetHead struct{ cx, cy, r float64 }

// newRivetSheet measures the heads of img, nil for none: in each cell the
// centre and the radius of the texels at least half covered. The circle
// cuts the head out, so faint haze around it never shows.
func newRivetSheet(img image.Image) *rivetSheet {
	if img == nil || img.Bounds().Dx() < rivetCols || img.Bounds().Dy() < rivetRows {
		return nil
	}
	b := img.Bounds()
	s := &rivetSheet{w: b.Dx(), h: b.Dy(), d: make([]float32, 4*b.Dx()*b.Dy())}
	for y := range s.h {
		for x := range s.w {
			c := color.NRGBAModel.Convert(img.At(b.Min.X+x, b.Min.Y+y)).(color.NRGBA)
			copy(s.d[4*(y*s.w+x):], []float32{float32(c.R), float32(c.G), float32(c.B), float32(max(0, float64(c.A)/255-0.5) * 2)})
		}
	}
	var sum [3]float64
	var n float64
	for row := range rivetRows {
		for col := range rivetCols {
			x0, x1 := col*s.w/rivetCols, (col+1)*s.w/rivetCols
			y0, y1 := row*s.h/rivetRows, (row+1)*s.h/rivetRows
			var sx, sy, k float64
			for y := y0; y < y1; y++ {
				for x := x0; x < x1; x++ {
					if i := 4 * (y*s.w + x); s.d[i+3] > 0 {
						sx, sy, k = sx+float64(x)+0.5, sy+float64(y)+0.5, k+1
						for ch := range 3 {
							sum[ch] += float64(s.d[i+ch])
						}
					}
				}
			}
			if k >= 4 {
				s.heads = append(s.heads, rivetHead{sx / k, sy / k, math.Sqrt(k / math.Pi)})
				n += k
			}
		}
	}
	if len(s.heads) == 0 {
		return nil
	}
	for ch := range 3 {
		s.mean[ch] = math.Max(1, sum[ch]/n)
	}
	return s
}

// at is the colour of head h at (u, v), in head radii from its centre: the
// coverage-weighted mean of n×n bilinear samples over a footprint of f
// radii, evenly brightened so that the heads' mean is as light as base. The
// painting keeps its own hue.
func (s *rivetSheet) at(h rivetHead, u, v, f float64, base color.RGBA) color.RGBA {
	r := h.r * 0.96 // inside the head's antialiased edge
	px, py, foot := h.cx+u*r, h.cy+v*r, f*r
	n := int(math.Min(8, math.Max(1, math.Ceil(foot))))
	var c [3]float64
	var a float64
	for j := range n {
		for i := range n {
			x := px + foot*((float64(i)+0.5)/float64(n)-0.5) - 0.5
			y := py + foot*((float64(j)+0.5)/float64(n)-0.5) - 0.5
			x0, y0 := math.Floor(x), math.Floor(y)
			fx, fy := x-x0, y-y0
			for _, t := range [4]struct {
				dx, dy int
				w      float64
			}{{0, 0, (1 - fx) * (1 - fy)}, {1, 0, fx * (1 - fy)}, {0, 1, (1 - fx) * fy}, {1, 1, fx * fy}} {
				tx := min(s.w-1, max(0, int(x0)+t.dx))
				ty := min(s.h-1, max(0, int(y0)+t.dy))
				k := 4 * (ty*s.w + tx)
				w := t.w * float64(s.d[k+3])
				for ch := range 3 {
					c[ch] += w * float64(s.d[k+ch])
				}
				a += w
			}
		}
	}
	if a <= 0 {
		c, a = s.mean, 1
	}
	k := luma(float64(base.R), float64(base.G), float64(base.B)) / luma(s.mean[0], s.mean[1], s.mean[2]) / a
	m := func(ch int) uint8 { return uint8(math.Max(0, math.Min(255, math.Round(c[ch]*k)))) }
	return color.RGBA{m(0), m(1), m(2), 0xFF}
}

// luma is the lightness of a colour.
func luma(r, g, b float64) float64 { return 0.299*r + 0.587*g + 0.114*b }
