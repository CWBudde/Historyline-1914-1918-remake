package worldmaterial

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

// Opaque is the alpha from which a kit pixel counts as part of its item for
// the colour mean and the extent.
const Opaque = 128

// Haze is the fraction of pixels with an alpha in 1..cut-1: the faint matte
// the generator leaves between and around cut-out items.
func Haze(im *image.NRGBA, cut uint8) float64 {
	b := im.Bounds()
	n := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if a := im.NRGBAAt(x, y).A; a > 0 && a < cut {
				n++
			}
		}
	}
	return float64(n) / float64(b.Dx()*b.Dy())
}

// Cut returns a copy of im with every alpha below cut set to zero.
func Cut(im *image.NRGBA, cut uint8) *image.NRGBA {
	out := image.NewNRGBA(im.Bounds())
	copy(out.Pix, im.Pix)
	for i := 3; i < len(out.Pix); i += 4 {
		if out.Pix[i] < cut {
			out.Pix[i-3], out.Pix[i-2], out.Pix[i-1], out.Pix[i] = 0, 0, 0, 0
		}
	}
	return out
}

// Item is the object of one sheet cell: the 8-connected alpha components
// whose centroids lie in the cell. A component may reach into the next cell.
type Item struct {
	Cell       image.Point     // column, row from 0
	Bounds     image.Rectangle // of every pixel with alpha > 0
	Centroid   [2]float64      // alpha-weighted, in sheet pixels
	Area       int             // pixels with alpha > 0
	Components int
	pixels     []int   // indices into the sheet, row-major
	alpha      float64 // alpha sum, the centroid weight
}

// Image returns a sheet-sized copy of sheet holding only the item's pixels.
func (it Item) Image(sheet *image.NRGBA) *image.NRGBA {
	b := sheet.Bounds()
	out := image.NewNRGBA(b)
	for _, i := range it.pixels {
		x, y := b.Min.X+i%b.Dx(), b.Min.Y+i/b.Dx()
		out.SetNRGBA(x, y, sheet.NRGBAAt(x, y))
	}
	return out
}

// Items splits a cut sheet of cols×rows equal cells into one item per cell
// that holds any. Components smaller than minArea pixels are specks and are
// dropped. Items come in row-major cell order.
func Items(im *image.NRGBA, cols, rows, minArea int) ([]Item, error) {
	b := im.Bounds()
	w, h := b.Dx(), b.Dy()
	if cols <= 0 || rows <= 0 || w < cols || h < rows {
		return nil, fmt.Errorf("a %dx%d sheet has no %dx%d grid", w, h, cols, rows)
	}
	label := make([]int32, w*h)
	cells := map[image.Point]*Item{}
	var stack []int
	for start := range label {
		if label[start] != 0 || im.Pix[im.PixOffset(b.Min.X+start%w, b.Min.Y+start/w)+3] == 0 {
			continue
		}
		label[start] = 1
		stack = append(stack[:0], start)
		var pixels []int
		var sx, sy, sa float64
		for len(stack) > 0 {
			i := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			pixels = append(pixels, i)
			x, y := i%w, i/w
			a := float64(im.Pix[im.PixOffset(b.Min.X+x, b.Min.Y+y)+3])
			sx, sy, sa = sx+a*(float64(x)+.5), sy+a*(float64(y)+.5), sa+a
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					nx, ny := x+dx, y+dy
					if nx < 0 || ny < 0 || nx >= w || ny >= h {
						continue
					}
					if j := ny*w + nx; label[j] == 0 && im.Pix[im.PixOffset(b.Min.X+nx, b.Min.Y+ny)+3] != 0 {
						label[j] = 1
						stack = append(stack, j)
					}
				}
			}
		}
		if len(pixels) < minArea {
			continue
		}
		cx, cy := sx/sa, sy/sa
		cell := image.Pt(min(int(cx*float64(cols)/float64(w)), cols-1), min(int(cy*float64(rows)/float64(h)), rows-1))
		it := cells[cell]
		if it == nil {
			it = &Item{Cell: cell}
			cells[cell] = it
		}
		it.Centroid = [2]float64{(it.Centroid[0]*it.alpha + sx) / (it.alpha + sa), (it.Centroid[1]*it.alpha + sy) / (it.alpha + sa)}
		it.alpha += sa
		it.Area += len(pixels)
		it.Components++
		it.pixels = append(it.pixels, pixels...)
		for _, i := range pixels {
			it.Bounds = it.Bounds.Union(image.Rect(i%w, i/w, i%w+1, i/w+1))
		}
	}
	var out []Item
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if it := cells[image.Pt(c, r)]; it != nil {
				it.Bounds = it.Bounds.Add(b.Min)
				it.Centroid[0] += float64(b.Min.X)
				it.Centroid[1] += float64(b.Min.Y)
				out = append(out, *it)
			}
		}
	}
	return out, nil
}

// Resample integrates the rectangle r of src onto a w×h image with an exact
// area box in premultiplied alpha, so transparent pixels add no colour to an
// item's edge.
func Resample(src *image.NRGBA, r image.Rectangle, w, h int) (*image.NRGBA, error) {
	r = r.Intersect(src.Bounds())
	if w <= 0 || h <= 0 || r.Empty() {
		return nil, fmt.Errorf("cannot resample %v to %dx%d", r, w, h)
	}
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	dx, dy := float64(r.Dx())/float64(w), float64(r.Dy())/float64(h)
	for y := 0; y < h; y++ {
		top, bottom := float64(y)*dy, float64(y+1)*dy
		for x := 0; x < w; x++ {
			left, right := float64(x)*dx, float64(x+1)*dx
			var sum [4]float64
			for sy := int(math.Floor(top)); sy < int(math.Ceil(bottom)) && sy < r.Dy(); sy++ {
				wy := math.Min(bottom, float64(sy+1)) - math.Max(top, float64(sy))
				for sx := int(math.Floor(left)); sx < int(math.Ceil(right)) && sx < r.Dx(); sx++ {
					weight := wy * (math.Min(right, float64(sx+1)) - math.Max(left, float64(sx)))
					c := src.NRGBAAt(r.Min.X+sx, r.Min.Y+sy)
					a := float64(c.A) * weight
					sum[0] += float64(c.R) * a
					sum[1] += float64(c.G) * a
					sum[2] += float64(c.B) * a
					sum[3] += a
				}
			}
			if sum[3] == 0 {
				continue
			}
			out.SetNRGBA(x, y, color.NRGBA{rounded(sum[0] / sum[3]), rounded(sum[1] / sum[3]), rounded(sum[2] / sum[3]), rounded(sum[3] / (dx * dy))})
		}
	}
	return out, nil
}

// OpaqueMean is the mean sRGB colour of the pixels with alpha >= Opaque.
func OpaqueMean(im *image.NRGBA) [3]float64 {
	var sum [3]float64
	n := 0
	for i := 0; i < len(im.Pix); i += 4 {
		if im.Pix[i+3] >= Opaque {
			sum[0] += float64(im.Pix[i])
			sum[1] += float64(im.Pix[i+1])
			sum[2] += float64(im.Pix[i+2])
			n++
		}
	}
	if n == 0 {
		return sum
	}
	return [3]float64{sum[0] / float64(n), sum[1] / float64(n), sum[2] / float64(n)}
}

// GradeTo shifts L*, a* and b* of the visible pixels by one offset so that the
// item's own colour lands on target; alpha stays. With reach > 0 the offset is
// weighted down with the CIELAB distance from the item's mean colour, falling
// to 1/e at reach, so that a stone's waterline, foam and translucent shadow on
// the water keep their colour; with reach 0 every pixel moves alike. The item's own colour is the mean
// of the opaque pixels under the same weights, which GradeTo also returns.
// The offset is refined a few times, because the sRGB mean is not linear in
// Lab.
func GradeTo(im *image.NRGBA, target [3]float64, reach float64) (*image.NRGBA, [3]float64, error) {
	m0 := OpaqueMean(im)
	if m0 == ([3]float64{}) {
		return nil, m0, fmt.Errorf("no opaque pixels to grade")
	}
	tl, ta, tb := toLab(color.NRGBA{rounded(target[0]), rounded(target[1]), rounded(target[2]), 255})
	ml0, ma0, mb0 := toLab(color.NRGBA{rounded(m0[0]), rounded(m0[1]), rounded(m0[2]), 255})
	lab := make([][3]float64, len(im.Pix)/4)
	weight := make([]float64, len(lab))
	for i := range lab {
		l, a, b := toLab(color.NRGBA{im.Pix[4*i], im.Pix[4*i+1], im.Pix[4*i+2], 255})
		lab[i] = [3]float64{l, a, b}
		d2 := (l-ml0)*(l-ml0) + (a-ma0)*(a-ma0) + (b-mb0)*(b-mb0)
		weight[i] = 1
		if reach > 0 {
			weight[i] = math.Exp(-d2 / (reach * reach))
		}
	}
	own := func(im *image.NRGBA) [3]float64 {
		var sum [3]float64
		total := 0.
		for i, w := range weight {
			if im.Pix[4*i+3] >= Opaque {
				sum[0], sum[1], sum[2] = sum[0]+w*float64(im.Pix[4*i]), sum[1]+w*float64(im.Pix[4*i+1]), sum[2]+w*float64(im.Pix[4*i+2])
				total += w
			}
		}
		return [3]float64{sum[0] / total, sum[1] / total, sum[2] / total}
	}
	var offset [3]float64
	out := im
	for range 6 {
		m := own(out)
		ml, ma, mb := toLab(color.NRGBA{rounded(m[0]), rounded(m[1]), rounded(m[2]), 255})
		offset = [3]float64{offset[0] + tl - ml, offset[1] + ta - ma, offset[2] + tb - mb}
		out = image.NewNRGBA(im.Bounds())
		for i, p := range lab {
			if im.Pix[4*i+3] == 0 {
				continue
			}
			w := weight[i]
			c := fromLab(p[0]+w*offset[0], p[1]+w*offset[1], p[2]+w*offset[2])
			out.Pix[4*i], out.Pix[4*i+1], out.Pix[4*i+2], out.Pix[4*i+3] = c.R, c.G, c.B, im.Pix[4*i+3]
		}
	}
	return out, own(out), nil
}
