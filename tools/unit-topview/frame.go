package unittopview

import (
	"fmt"
	"image"
	"image/color"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

// Sprite geometry shared with every unit family: a 96×96 logical sprite whose
// logical anchor (48,49) sits on the hex centre.
const (
	logicalSize = 96
	anchorX     = 48.0
	anchorY     = 49.0
)

// framing maps master pixels to logical sprite pixels: the master point centre
// lands on the anchor, and scale master pixels make one logical pixel.
type framing struct {
	centreX, centreY float64
	scale            float64
}

// cutHaze clears the alpha haze a generated cut-out carries outside its object:
// every pixel below cut becomes fully transparent.
func cutHaze(src image.Image, cut uint8) *image.NRGBA {
	b := src.Bounds()
	out := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := range b.Dy() {
		for x := range b.Dx() {
			c := color.NRGBAModel.Convert(src.At(b.Min.X+x, b.Min.Y+y)).(color.NRGBA)
			if c.A < cut {
				continue
			}
			out.SetNRGBA(x, y, c)
		}
	}
	return out
}

// solidBounds is the box of the pixels with alpha above 128.
func solidBounds(im *image.NRGBA) (image.Rectangle, error) {
	r := image.Rectangle{}
	b := im.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if im.NRGBAAt(x, y).A > 128 {
				r = r.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	if r.Empty() {
		return r, fmt.Errorf("master has no solid pixels")
	}
	return r, nil
}

// frameOf centres the solid box of base on the anchor and scales its larger side
// to diameter logical pixels.
func frameOf(base *image.NRGBA, diameter float64) (framing, error) {
	r, err := solidBounds(base)
	if err != nil {
		return framing{}, err
	}
	side := max(r.Dx(), r.Dy())
	return framing{
		centreX: float64(r.Min.X+r.Max.X) / 2,
		centreY: float64(r.Min.Y+r.Max.Y) / 2,
		scale:   float64(side) / diameter,
	}, nil
}

// render resamples the master into a 96×96 logical sprite at density d with an
// antialiased Catmull-Rom kernel on premultiplied colour.
func render(master *image.NRGBA, f framing, d int) *image.NRGBA {
	size := logicalSize * d
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	// Destination pixel p maps to logical p/d, then to the master point
	// centre + (logical - anchor) * scale.
	s := f.scale / float64(d)
	m := f64.Aff3{
		1 / s, 0, anchorX*float64(d) - f.centreX/s,
		0, 1 / s, anchorY*float64(d) - f.centreY/s,
	}
	xdraw.CatmullRom.Transform(dst, m, master, master.Bounds(), xdraw.Src, nil)
	out := image.NewNRGBA(dst.Bounds())
	for y := range size {
		for x := range size {
			c := color.NRGBAModel.Convert(dst.RGBAAt(x, y)).(color.NRGBA)
			if c.A == 0 {
				c = color.NRGBA{}
			}
			out.SetNRGBA(x, y, c)
		}
	}
	return out
}

// spent is the acted state every unit family uses: the red channel as grey,
// alpha unchanged.
func spent(im *image.NRGBA) *image.NRGBA {
	out := image.NewNRGBA(im.Bounds())
	for i := 0; i < len(im.Pix); i += 4 {
		r, a := im.Pix[i], im.Pix[i+3]
		out.Pix[i], out.Pix[i+1], out.Pix[i+2], out.Pix[i+3] = r, r, r, a
	}
	return out
}
