package dospanel

import (
	"image"
	"image/color"
	"math"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/texture"
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
	steel, ground, rivets := texture.NewTile(s.Steel), texture.NewTile(s.Ground), newRivetSheet(s.Rivets)
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
	steel, ground *texture.Tile
	rivets        *rivetSheet
}

// steelColour is the steel of tone t at device pixel (x, y): the painted
// surface over the ramp, or the drawn grain without one.
func (p painter) steelColour(x, y int, t float64) color.RGBA {
	if s := surfaces.steel; s != nil {
		dr, dg, db := s.At(x, y, p.k*steelSpan)
		return texture.Modulate(ramp(t), steelGain, dr, dg, db)
	}
	q := max(1, int(math.Round(p.k)))
	return ramp(t + 0.06*grain(x/q, y/q))
}

// groundColour is the hex ground of tone t at device pixel (x, y).
func (p painter) groundColour(x, y int, t float64) color.RGBA {
	if s := surfaces.ground; s != nil {
		dr, dg, db := s.At(x, y, p.k*groundSpan)
		return texture.Modulate(lerp(fieldDark, fieldLight, t), groundGain, dr, dg, db)
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
