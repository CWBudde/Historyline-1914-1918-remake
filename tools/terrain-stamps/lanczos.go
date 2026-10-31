package terrainstamps

import (
	"errors"
	"image"
	"image/color"
	"math"
)

// Lanczos3 implementation retained from tools/art-review/image.go.
func resizeLanczos(src *image.NRGBA, width, height int) (*image.NRGBA, error) {
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	if width <= 0 || height <= 0 || width > sw || height > sh {
		return nil, errors.New("lanczos resize requires positive dimensions without upscaling")
	}
	dst := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		y0, wy := lanczosWeights(y, sh, height)
		for x := 0; x < width; x++ {
			x0, wx := lanczosWeights(x, sw, width)
			var sumW, sumA, sumR, sumG, sumB float64
			for j, vy := range wy {
				if vy == 0 {
					continue
				}
				for i, vx := range wx {
					if vx == 0 {
						continue
					}
					weight := vx * vy
					p := src.NRGBAAt(x0+i+src.Bounds().Min.X, y0+j+src.Bounds().Min.Y)
					weightA := weight * float64(p.A)
					sumW += weight
					sumA += weightA
					sumR += weightA * float64(p.R)
					sumG += weightA * float64(p.G)
					sumB += weightA * float64(p.B)
				}
			}
			rawAlpha := sumA / sumW
			if rawAlpha <= 0 {
				continue
			}
			a := uint8(math.Floor(min(rawAlpha, 255) + 0.5))
			if a == 0 {
				continue
			}
			channel := func(sum float64) uint8 {
				// Cancel the common normalization factor before division;
				// normalizing both operands first can perturb exact half ties.
				premultiplied := min(max(sum, 0), 255*sumA)
				return uint8(math.Floor(premultiplied/sumA + 0.5))
			}
			dst.SetNRGBA(x, y, color.NRGBA{R: channel(sumR), G: channel(sumG), B: channel(sumB), A: a})
		}
	}
	return dst, nil
}

// Coordinates use units of 1/(2*destination), as in bilinearSpan; the support
// radius is three times the source/destination ratio. Retain signed lobes.
func lanczosWeights(d, source, destination int) (first int, weights []float64) {
	s, t := int64(source), int64(destination)
	center := (2*int64(d) + 1) * s
	first = max(0, int((center-6*s-t)/(2*t)))
	end := min(source, int((center+6*s+t-1)/(2*t)))
	weights = make([]float64, end-first)
	for i := range weights {
		u := math.Abs(float64((2*int64(first+i)+1)*t-center) / float64(2*s))
		switch {
		case u == 0:
			weights[i] = 1
		case u >= 3 || u == 1 || u == 2:
			// Exact sinc zeros also preserve identity reductions.
		default:
			v := math.Pi * u
			weights[i] = math.Sin(v) / v * (math.Sin(v/3) / (v / 3))
		}
	}
	return
}
