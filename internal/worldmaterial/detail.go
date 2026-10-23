package worldmaterial

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

// HighPassRadius is the box radius, in world pixels, of the three blur passes
// that Combine subtracts from the detail's L*: about a 12-world-pixel kernel.
const HighPassRadius = 4

// ToneBand is the edge cross-fade, in world pixels, of a tone layer before
// the detail is added: about its largest patches, so a patch fades into the
// half-period shift instead of changing within the 4-pixel edge band.
const ToneBand = 40

// Tiled centre-crops src to the period aspect, makes it seamless with a
// worldBand feather and wraps it repeats×repeats times across one period at
// density, integrating each output pixel's source footprint with an area box.
// The tile is period/repeats world pixels, which need not be whole.
func Tiled(src image.Image, period image.Point, density, repeats int, worldBand float64) (*image.NRGBA, Crop, error) {
	w, h, err := dimensions(period, density)
	if err != nil {
		return nil, Crop{}, err
	}
	if repeats < 1 {
		return nil, Crop{}, fmt.Errorf("invalid detail repeats %d", repeats)
	}
	r, err := bounds(src)
	if err != nil {
		return nil, Crop{}, err
	}
	if err = opaque(src, r); err != nil {
		return nil, Crop{}, err
	}
	aspect := float64(period.X) / float64(period.Y)
	cw, ch := r.Dx(), r.Dy()
	if float64(cw)/float64(ch) > aspect {
		cw = int(math.Round(float64(ch) * aspect))
	} else {
		ch = int(math.Round(float64(cw) / aspect))
	}
	x0, y0 := (r.Dx()-cw)/2, (r.Dy()-ch)/2
	c := Crop{float64(x0), float64(y0), float64(x0 + cw), float64(y0 + ch)}
	tile := image.NewNRGBA(image.Rect(0, 0, cw, ch))
	for y := range ch {
		for x := range cw {
			tile.SetNRGBA(x, y, rgb(src, r.Min.X+x0+x, r.Min.Y+y0+y))
		}
	}
	// The feather is worldBand world pixels of the tile, in tile pixels.
	if tile, err = Seamless(tile, worldBand*float64(cw*repeats)/float64(period.X), 1); err != nil {
		return nil, Crop{}, err
	}
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	dx, dy := float64(cw*repeats)/float64(w), float64(ch*repeats)/float64(h)
	for y := range h {
		top, bottom := float64(y)*dy, float64(y+1)*dy
		for x := range w {
			left, right := float64(x)*dx, float64(x+1)*dx
			var sum [3]float64
			for sy := int(math.Floor(top)); sy < int(math.Ceil(bottom)); sy++ {
				wy := math.Min(bottom, float64(sy+1)) - math.Max(top, float64(sy))
				for sx := int(math.Floor(left)); sx < int(math.Ceil(right)); sx++ {
					weight := wy * (math.Min(right, float64(sx+1)) - math.Max(left, float64(sx)))
					v := tile.NRGBAAt(mod(sx, cw), mod(sy, ch))
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

// highPass returns l minus three periodic box blurs of l at radius.
func highPass(l []float64, w, h, radius int) []float64 {
	blur := l
	for range 3 {
		blur = boxMean(blur, w, h, radius)
	}
	out := make([]float64, len(l))
	for i := range l {
		out[i] = l[i] - blur[i]
	}
	return out
}

// Grade scales the L* of im about its mean by k and then adds offset to L*, a*
// and b*. It grades a tone layer whose patches came out flatter, or whose
// colour came out further from the target, than the checks allow.
func Grade(im *image.NRGBA, k float64, offset [3]float64) (*image.NRGBA, error) {
	if im == nil || !finite(k) || k <= 0 || !finite(offset[0]) || !finite(offset[1]) || !finite(offset[2]) {
		return nil, fmt.Errorf("invalid tone grade %v %v", k, offset)
	}
	p := newPlane(im)
	mean, _ := stats(p.l)
	out := image.NewNRGBA(image.Rect(0, 0, p.w, p.h))
	for y := range p.h {
		for x := range p.w {
			i := y*p.w + x
			out.SetNRGBA(x, y, fromLab(mean+k*(p.l[i]-mean)+offset[0], p.a[i]+offset[1], p.b[i]+offset[2]))
		}
	}
	return out, nil
}

// Combine adds the high-passed L* of detail, times gain, to the L* of tone;
// a* and b* stay tone's. Both are periodic images of one size at density.
func Combine(tone, detail *image.NRGBA, density int, gain float64) (*image.NRGBA, error) {
	if tone == nil || detail == nil || tone.Bounds().Size() != detail.Bounds().Size() {
		return nil, fmt.Errorf("tone and detail differ in size")
	}
	if density <= 0 || !finite(gain) || gain < 0 {
		return nil, fmt.Errorf("invalid density or gain")
	}
	t, d := newPlane(tone), newPlane(detail)
	hp := highPass(d.l, d.w, d.h, HighPassRadius*density)
	out := image.NewNRGBA(image.Rect(0, 0, t.w, t.h))
	for y := range t.h {
		for x := range t.w {
			i := y*t.w + x
			out.SetNRGBA(x, y, fromLab(t.l[i]+gain*hp[i], t.a[i], t.b[i]))
		}
	}
	return out, nil
}
