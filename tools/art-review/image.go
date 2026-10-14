package main

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"math"
)

type coverage struct {
	ReferencePixels         int      `json:"reference_visible_pixels"`
	CandidatePixels         int      `json:"candidate_visible_pixels"`
	MissingPixels           int      `json:"missing_visible_pixels"`
	ExtraPixels             int      `json:"extra_visible_pixels"`
	SharedPixels            int      `json:"shared_visible_pixels"`
	AlphaMismatchPixels     int      `json:"alpha_mismatch_pixels"`
	NormalizedAlphaDiff     float64  `json:"normalized_alpha_difference"`
	NormalizedRGBDifference *float64 `json:"normalized_rgb_difference"`
}

func boxDownsample(src *image.NRGBA, factor int) (*image.NRGBA, error) {
	if factor <= 0 || src.Bounds().Empty() || src.Bounds().Dx()%factor != 0 || src.Bounds().Dy()%factor != 0 {
		return nil, errors.New("box factor must be positive and divide both image dimensions")
	}
	return resizeArea(src, src.Bounds().Dx()/factor, src.Bounds().Dy()/factor)
}

// resizeArea uses integer rational-coordinate overlaps. RGB is weighted by
// alpha before averaging, preventing hidden transparent colors from bleeding
// into visible edges. Returned RGB is unpremultiplied.
func resizeArea(src *image.NRGBA, width, height int) (*image.NRGBA, error) {
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	if width <= 0 || height <= 0 || width > sw || height > sh {
		return nil, errors.New("area resize requires positive dimensions without upscaling")
	}
	dst := image.NewNRGBA(image.Rect(0, 0, width, height))
	area := int64(sw) * int64(sh)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var sumA, sumR, sumG, sumB int64
			for sy := y * sh / height; sy < ((y+1)*sh+height-1)/height; sy++ {
				wy := min((sy+1)*height, (y+1)*sh) - max(sy*height, y*sh)
				for sx := x * sw / width; sx < ((x+1)*sw+width-1)/width; sx++ {
					wx := min((sx+1)*width, (x+1)*sw) - max(sx*width, x*sw)
					p := src.NRGBAAt(sx+src.Bounds().Min.X, sy+src.Bounds().Min.Y)
					weightA := int64(wx) * int64(wy) * int64(p.A)
					sumA += weightA
					sumR += weightA * int64(p.R)
					sumG += weightA * int64(p.G)
					sumB += weightA * int64(p.B)
				}
			}
			if sumA > 0 {
				p := color.NRGBA{R: uint8((sumR + sumA/2) / sumA), G: uint8((sumG + sumA/2) / sumA), B: uint8((sumB + sumA/2) / sumA), A: uint8((sumA + area/2) / area)}
				if p.A > 0 {
					dst.SetNRGBA(x, y, p)
				}
			}
		}
	}
	return dst, nil
}

func applyFootprint(src, ref *image.NRGBA) *image.NRGBA {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := src.NRGBAAt(x+src.Bounds().Min.X, y+src.Bounds().Min.Y)
			a := ref.NRGBAAt(x*ref.Bounds().Dx()/w+ref.Bounds().Min.X, y*ref.Bounds().Dy()/h+ref.Bounds().Min.Y).A
			p.A = min(p.A, a)
			if p.A > 0 {
				dst.SetNRGBA(x, y, p)
			}
		}
	}
	return dst
}

func sevenHex(tile *image.NRGBA) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, 238, 240))
	// Center (2,1); even-line neighbors sorted for original raster blits.
	for _, cell := range [][2]int{{0, 1}, {1, 0}, {1, 1}, {2, 1}, {3, 0}, {3, 1}, {4, 1}} {
		line, col := cell[0], cell[1]
		p := image.Pt(col*142+(line%2)*71-71, line*40)
		draw.Draw(dst, tile.Bounds().Sub(tile.Bounds().Min).Add(p), tile, tile.Bounds().Min, draw.Over)
	}
	return dst
}

func measure(ref, candidate *image.NRGBA) coverage {
	var result coverage
	var rgbSum, sharedWeight, alphaSum float64
	for y := 0; y < ref.Bounds().Dy(); y++ {
		for x := 0; x < ref.Bounds().Dx(); x++ {
			r, c := ref.NRGBAAt(x+ref.Bounds().Min.X, y+ref.Bounds().Min.Y), candidate.NRGBAAt(x+candidate.Bounds().Min.X, y+candidate.Bounds().Min.Y)
			if r.A > 0 {
				result.ReferencePixels++
			}
			if c.A > 0 {
				result.CandidatePixels++
			}
			if r.A > 0 && c.A == 0 {
				result.MissingPixels++
			}
			if r.A == 0 && c.A > 0 {
				result.ExtraPixels++
			}
			if r.A != c.A {
				result.AlphaMismatchPixels++
			}
			alphaSum += float64(abs(int(r.A) - int(c.A)))
			if r.A > 0 && c.A > 0 {
				result.SharedPixels++
				weight := float64(min(r.A, c.A))
				sharedWeight += weight
				rgbSum += weight * float64(abs(int(r.R)-int(c.R))+abs(int(r.G)-int(c.G))+abs(int(r.B)-int(c.B)))
			}
		}
	}
	if sharedWeight > 0 {
		v := rgbSum / (sharedWeight * 3 * 255)
		result.NormalizedRGBDifference = &v
	}
	if !ref.Bounds().Empty() {
		result.NormalizedAlphaDiff = alphaSum / (float64(ref.Bounds().Dx()*ref.Bounds().Dy()) * 255)
	}
	return result
}

func comparison(ref, candidate *image.NRGBA) *image.NRGBA {
	w, h := ref.Bounds().Dx(), ref.Bounds().Dy()
	diff := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, c := ref.NRGBAAt(x, y), candidate.NRGBAAt(x, y)
			if r.A == 0 && c.A == 0 {
				continue
			}
			ad := abs(int(r.A) - int(c.A))
			channel := func(a, b uint8) uint8 {
				return uint8(max(ad, abs((int(a)*int(r.A)+127)/255-(int(b)*int(c.A)+127)/255)))
			}
			diff.SetNRGBA(x, y, color.NRGBA{R: channel(r.R, c.R), G: channel(r.G, c.G), B: channel(r.B, c.B), A: 255})
		}
	}
	dst := image.NewNRGBA(image.Rect(0, 0, 3*w*4, h*5))
	for i, panel := range []*image.NRGBA{ref, candidate, diff} {
		draw.Draw(dst, image.Rect(i*w, 0, (i+1)*w, h), panel, image.Point{}, draw.Src)
		for y := 0; y < h*4; y++ {
			for x := 0; x < w*4; x++ {
				dst.SetNRGBA(i*w*4+x, h+y, panel.NRGBAAt(x/4, y/4))
			}
		}
	}
	return dst
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// resizeBilinear reduces with a separable tent widened by the source/destination
// ratio for antialiasing. Pixel centers map as (d+0.5)*source/destination-0.5.
// Support outside the image is discarded and the remaining weights normalized.
// Integer-coordinate weights avoid coordinate drift; float64 sums are evaluated
// in row-major order. RGB bytes (no linear-light conversion) are weighted by
// alpha, then unpremultiplied; channels round half up once, after the full 2D
// filter. Fully transparent output is zero RGB. There is no intermediate
// quantization or dependence on a platform graphics API.
func resizeBilinear(src *image.NRGBA, width, height int) (*image.NRGBA, error) {
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	if width <= 0 || height <= 0 || width > sw || height > sh {
		return nil, errors.New("bilinear resize requires positive dimensions without upscaling")
	}
	dst := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		y0, y1, cy := bilinearSpan(y, sh, height)
		for x := 0; x < width; x++ {
			x0, x1, cx := bilinearSpan(x, sw, width)
			var sumW, sumA, sumR, sumG, sumB float64
			for sy := y0; sy < y1; sy++ {
				dy := (2*int64(sy)+1)*int64(height) - cy
				if dy < 0 {
					dy = -dy
				}
				wy := 2*int64(sh) - dy
				if wy <= 0 {
					continue
				}
				for sx := x0; sx < x1; sx++ {
					dx := (2*int64(sx)+1)*int64(width) - cx
					if dx < 0 {
						dx = -dx
					}
					wx := 2*int64(sw) - dx
					if wx <= 0 {
						continue
					}
					weight := float64(wx) * float64(wy)
					p := src.NRGBAAt(sx+src.Bounds().Min.X, sy+src.Bounds().Min.Y)
					weightA := weight * float64(p.A)
					sumW += weight
					sumA += weightA
					sumR += weightA * float64(p.R)
					sumG += weightA * float64(p.G)
					sumB += weightA * float64(p.B)
				}
			}
			if sumA > 0 {
				a := uint8(math.Floor(sumA/sumW + 0.5))
				if a > 0 {
					dst.SetNRGBA(x, y, color.NRGBA{R: uint8(math.Floor(sumR/sumA + 0.5)), G: uint8(math.Floor(sumG/sumA + 0.5)), B: uint8(math.Floor(sumB/sumA + 0.5)), A: a})
				}
			}
		}
	}
	return dst, nil
}

// Centers and tent radii use units of 1/(2*destination). The initial index may
// include a zero-weight sample; resizeBilinear skips it. Clip, do not wrap edges.
func bilinearSpan(d, source, destination int) (first, end int, center int64) {
	s, t := int64(source), int64(destination)
	center = (2*int64(d) + 1) * s
	first = max(0, int((center-2*s-t)/(2*t)))
	end = min(source, int((center+2*s+t-1)/(2*t)))
	return
}
