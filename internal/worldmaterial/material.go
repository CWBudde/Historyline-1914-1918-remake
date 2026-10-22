// Package worldmaterial exports and samples opaque periodic world materials.
package worldmaterial

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

const (
	MeadowID     = "material/grass/summer"
	ScrubID      = "material/hills/summer"
	Group        = "materials"
	PeriodWidth  = 568
	PeriodHeight = 400
	maxPixels    = 32 * 1024 * 1024
)

// Crop describes continuous source pixel boundaries relative to Bounds().Min.
// A source pixel's centre is (x+.5,y+.5) in this coordinate system.
type Crop struct{ Left, Top, Right, Bottom float64 }

func dimensions(p image.Point, d int) (int, int, error) {
	if p.X <= 0 || p.Y <= 0 || d <= 0 || p.X > maxPixels/d || p.Y > maxPixels/d {
		return 0, 0, fmt.Errorf("invalid material dimensions or density")
	}
	w, h := p.X*d, p.Y*d
	if w > maxPixels/h {
		return 0, 0, fmt.Errorf("material exceeds %d pixels", maxPixels)
	}
	return w, h, nil
}
func bounds(src image.Image) (image.Rectangle, error) {
	if src == nil {
		return image.Rectangle{}, fmt.Errorf("nil image")
	}
	r := src.Bounds()
	if r.Max.X <= r.Min.X || r.Max.Y <= r.Min.Y || uint(r.Max.X)-uint(r.Min.X) > maxPixels || uint(r.Max.Y)-uint(r.Min.Y) > maxPixels {
		return r, fmt.Errorf("invalid image bounds")
	}
	return r, nil
}
func opaque(src image.Image, r image.Rectangle) error {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			_, _, _, a := src.At(x, y).RGBA()
			if a != 65535 {
				return fmt.Errorf("nonopaque material at %d,%d", x, y)
			}
		}
	}
	return nil
}
func rgb(src image.Image, x, y int) color.NRGBA {
	if p, ok := src.(*image.NRGBA); ok {
		return p.NRGBAAt(x, y)
	}
	return color.NRGBAModel.Convert(src.At(x, y)).(color.NRGBA)
}
func rounded(v float64) uint8 { return uint8(math.Floor(v + .5)) }

// Export centre-crops to the exact period aspect and integrates each output
// pixel's fractional source footprint with an area-box filter.
func Export(src image.Image, period image.Point, density int) (*image.NRGBA, Crop, error) {
	w, h, err := dimensions(period, density)
	if err != nil {
		return nil, Crop{}, err
	}
	r, err := bounds(src)
	if err != nil {
		return nil, Crop{}, err
	}
	if err = opaque(src, r); err != nil {
		return nil, Crop{}, err
	}
	sw, sh := float64(r.Dx()), float64(r.Dy())
	cw, ch := sw, sh
	aspect := float64(period.X) / float64(period.Y)
	if sw/sh > aspect {
		cw = sh * aspect
	} else {
		ch = sw / aspect
	}
	c := Crop{(sw - cw) / 2, (sh - ch) / 2, (sw + cw) / 2, (sh + ch) / 2}
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	dx, dy := cw/float64(w), ch/float64(h)
	for y := 0; y < h; y++ {
		top, bottom := c.Top+float64(y)*dy, c.Top+float64(y+1)*dy
		for x := 0; x < w; x++ {
			left, right := c.Left+float64(x)*dx, c.Left+float64(x+1)*dx
			var sum [3]float64
			for sy := int(math.Floor(top)); sy < int(math.Ceil(bottom)) && sy < r.Dy(); sy++ {
				wy := math.Min(bottom, float64(sy+1)) - math.Max(top, float64(sy))
				for sx := int(math.Floor(left)); sx < int(math.Ceil(right)) && sx < r.Dx(); sx++ {
					weight := wy * (math.Min(right, float64(sx+1)) - math.Max(left, float64(sx)))
					v := rgb(src, r.Min.X+sx, r.Min.Y+sy)
					sum[0] += float64(v.R) * weight
					sum[1] += float64(v.G) * weight
					sum[2] += float64(v.B) * weight
				}
			}
			area := (right - left) * (bottom - top)
			out.SetNRGBA(x, y, color.NRGBA{rounded(sum[0] / area), rounded(sum[1] / area), rounded(sum[2] / area), 255})
		}
	}
	return out, c, nil
}
func smooth(v float64) float64 { return v * v * (3 - 2*v) }
func blend(a, b color.NRGBA, t float64) color.NRGBA {
	return color.NRGBA{rounded(float64(a.R)*(1-t) + float64(b.R)*t), rounded(float64(a.G)*(1-t) + float64(b.G)*t), rounded(float64(a.B)*(1-t) + float64(b.B)*t), a.A}
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// Seamless returns a normalized copy with smooth endpoint feathering, first
// horizontally then vertically. The source remains unchanged.
func Seamless(src *image.NRGBA, worldBand float64, density int) (*image.NRGBA, error) {
	if src == nil {
		return nil, fmt.Errorf("nil image")
	}
	r, err := bounds(src)
	if err != nil {
		return nil, err
	}
	if _, _, err = dimensions(r.Size(), 1); err != nil {
		return nil, err
	}
	band := worldBand * float64(density)
	if density <= 0 || !finite(worldBand) || !finite(band) || band <= 0 || band >= float64(r.Dx())/2 || band >= float64(r.Dy())/2 {
		return nil, fmt.Errorf("invalid feather band")
	}
	if err = opaque(src, r); err != nil {
		return nil, err
	}
	out := image.NewNRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			out.SetNRGBA(x, y, src.NRGBAAt(r.Min.X+x, r.Min.Y+y))
		}
	}
	for y := 0; y < r.Dy(); y++ {
		mean := blend(out.NRGBAAt(0, y), out.NRGBAAt(r.Dx()-1, y), .5)
		for x := 0; x < r.Dx(); x++ {
			distance := float64(min(x, r.Dx()-1-x))
			if distance < band {
				out.SetNRGBA(x, y, blend(mean, out.NRGBAAt(x, y), smooth(distance/band)))
			}
		}
	}
	for x := 0; x < r.Dx(); x++ {
		mean := blend(out.NRGBAAt(x, 0), out.NRGBAAt(x, r.Dy()-1), .5)
		for y := 0; y < r.Dy(); y++ {
			distance := float64(min(y, r.Dy()-1-y))
			if distance < band {
				out.SetNRGBA(x, y, blend(mean, out.NRGBAAt(x, y), smooth(distance/band)))
			}
		}
	}
	return out, nil
}

// Sampler retains immutable, opaque periodic material images. Callers must not
// mutate either source after New; validation is performed only at construction.
type Sampler struct {
	grass, scrub image.Image
	gb, sb       image.Rectangle
	size         image.Point
	density      int
}

// New validates dimensions, opacity and matching opposite RGB endpoints once.
func New(grass, scrub image.Image, period image.Point, density int) (*Sampler, error) {
	w, h, err := dimensions(period, density)
	if err != nil {
		return nil, err
	}
	var rs [2]image.Rectangle
	for i, src := range []image.Image{grass, scrub} {
		r, e := bounds(src)
		if e != nil {
			return nil, e
		}
		if r.Size() != image.Pt(w, h) {
			return nil, fmt.Errorf("material shape %v, expected %dx%d", r.Size(), w, h)
		}
		if e = opaque(src, r); e != nil {
			return nil, e
		}
		for y := r.Min.Y; y < r.Max.Y; y++ {
			if rgb(src, r.Min.X, y) != rgb(src, r.Max.X-1, y) {
				return nil, fmt.Errorf("horizontal endpoints differ")
			}
		}
		for x := r.Min.X; x < r.Max.X; x++ {
			if rgb(src, x, r.Min.Y) != rgb(src, x, r.Max.Y-1) {
				return nil, fmt.Errorf("vertical endpoints differ")
			}
		}
		rs[i] = r
	}
	return &Sampler{grass, scrub, rs[0], rs[1], image.Pt(w, h), density}, nil
}
func mod(v, n int) int {
	v %= n
	if v < 0 {
		v += n
	}
	return v
}

// Phase returns the positive-modulo global texel origin.
func (s *Sampler) Phase(origin image.Point) image.Point {
	return image.Pt(mod(origin.X, s.size.X), mod(origin.Y, s.size.Y))
}

// Tile samples global material texels under a mask, preserving mask alpha and
// zeroing hidden RGB. Weight receives world pixel centres and must return 0..1.
func (s *Sampler) Tile(mask image.Image, origin image.Point, weight func(float64, float64) float64) (*image.NRGBA, error) {
	r, err := bounds(mask)
	if err != nil {
		return nil, err
	}
	if _, _, err = dimensions(r.Size(), 1); err != nil {
		return nil, err
	}
	out := image.NewNRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	phase := s.Phase(origin)
	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			_, _, _, alpha := mask.At(r.Min.X+x, r.Min.Y+y).RGBA()
			if alpha == 0 {
				continue
			}
			t := 0.
			if weight != nil {
				t = weight((float64(origin.X)+float64(x)+.5)/float64(s.density), (float64(origin.Y)+float64(y)+.5)/float64(s.density))
				if !finite(t) || t < 0 || t > 1 {
					return nil, fmt.Errorf("invalid material weight at %d,%d", x, y)
				}
			}
			tx, ty := (phase.X+x)%s.size.X, (phase.Y+y)%s.size.Y
			v := blend(rgb(s.grass, s.gb.Min.X+tx, s.gb.Min.Y+ty), rgb(s.scrub, s.sb.Min.X+tx, s.sb.Min.Y+ty), t)
			v.A = uint8(alpha >> 8)
			out.SetNRGBA(x, y, v)
		}
	}
	return out, nil
}
