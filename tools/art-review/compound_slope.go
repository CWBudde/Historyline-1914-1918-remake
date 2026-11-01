package artreview

import (
	"errors"
	"fmt"
	"image"
	"math"
)

// Slope fade bounds (world px) and its noise cells (world px).
const (
	compoundSlopeMin   = 8.0
	compoundSlopeMax   = 80.0
	compoundSlopeCell  = 11.0
	compoundSlopeCell2 = 3.6
)

// compoundValidateSlope allows a slope fade on scenes that keep the default
// guard: the fade is what lets a free-standing cut-out grow towards the outer
// boundary.
func compoundValidateSlope(framing string, width float64) error {
	if width == 0 {
		return nil
	}
	if framing == "reference-layout" {
		return errors.New("slope_fade excludes reference-layout framing")
	}
	if !(width >= compoundSlopeMin && width <= compoundSlopeMax) {
		return fmt.Errorf("slope_fade must lie in [%g,%g] world px", compoundSlopeMin, compoundSlopeMax)
	}
	return nil
}

// compoundSlopeFade fades a cut-out's alpha towards its own outline over about
// width world px (scale world px per source px), so its lower slopes break into
// patches over the ground instead of ending at a rim. b is the alpha blurred
// twice by a box of width/2: 1/2 on a straight outline, 1 width inside. Alpha is
// scaled by smoothstep of the smaller of (b-0.5)/0.3, which reaches zero on the
// outline, and (b-e+0.1)/0.2 with e between 0.45 and 0.9 by two octaves of value
// noise in world coordinates.
func compoundSlopeFade(src *image.NRGBA, scale, width float64) *image.NRGBA {
	w, h := src.Rect.Dx(), src.Rect.Dy()
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	a := make([]float64, w*h)
	for y := range h {
		for x := range w {
			a[y*w+x] = float64(src.Pix[y*src.Stride+x*4+3]) / 255
		}
	}
	r := max(1, int(math.Round(width/scale/2)))
	b := compoundBoxBlur(compoundBoxBlur(a, w, h, r), w, h, r)
	for y := range h {
		for x := range w {
			i := y*w + x
			p := connectionPoint{float64(x) * scale, float64(y) * scale}
			e := .45 + .45*(.65*compoundValueNoise(p, compoundSlopeCell, 41)+.35*compoundValueNoise(p, compoundSlopeCell2, 43))
			t := connectionSmoothStep(min((b[i]-e+.1)/.2, (b[i]-.5)/.3))
			s, d := src.Pix[y*src.Stride+x*4:], out.Pix[y*out.Stride+x*4:]
			copy(d[:3], s[:3])
			d[3] = uint8(math.Round(a[i] * t * 255))
		}
	}
	return out
}

// compoundBoxBlur averages v (w×h) over a (2r+1)² box, zero outside.
func compoundBoxBlur(v []float64, w, h, r int) []float64 {
	pass := func(in []float64, n, lines, stride, step int) []float64 {
		out := make([]float64, len(in))
		for j := range lines {
			at := func(i int) float64 {
				if i < 0 || i >= n {
					return 0
				}
				return in[j*stride+i*step]
			}
			sum := 0.0
			for i := -r; i <= r; i++ {
				sum += at(i)
			}
			for i := range n {
				out[j*stride+i*step] = sum / float64(2*r+1)
				sum += at(i+r+1) - at(i-r)
			}
		}
		return out
	}
	return pass(pass(v, w, h, w, 1), h, w, 1, w)
}
