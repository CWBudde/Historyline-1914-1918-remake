package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	xdraw "golang.org/x/image/draw"
)

func loadPNG(file string) (image.Image, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return png.Decode(f)
}

// save writes img as JPEG (quality 90) or, for a .png name, lossless PNG.
func save(file string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	if filepath.Ext(file) == ".png" {
		err = png.Encode(f, img)
	} else {
		err = jpeg.Encode(f, flatten(img, color.NRGBA{40, 46, 50, 255}), &jpeg.Options{Quality: 90})
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}

// flatten composes img over an opaque background.
func flatten(img image.Image, bg color.NRGBA) *image.NRGBA {
	out := image.NewNRGBA(img.Bounds())
	draw.Draw(out, out.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)
	draw.Draw(out, out.Bounds(), img, img.Bounds().Min, draw.Over)
	return out
}

// halve reduces a 2x image to 1x by averaging 2×2 blocks, as the game's 1x view does.
func halve(src image.Image) *image.NRGBA {
	b := src.Bounds()
	out := image.NewNRGBA(image.Rect(0, 0, b.Dx()/2, b.Dy()/2))
	xdraw.ApproxBiLinear.Scale(out, out.Bounds(), src, b, draw.Src, nil)
	return out
}

// fit scales src into a w×h box, keeping its aspect ratio, centred on a transparent canvas.
func fit(src image.Image, w, h int) *image.NRGBA {
	b := src.Bounds()
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	sx, sy := float64(w)/float64(b.Dx()), float64(h)/float64(b.Dy())
	s := min(sx, sy)
	dw, dh := max(1, int(float64(b.Dx())*s+.5)), max(1, int(float64(b.Dy())*s+.5))
	r := image.Rect((w-dw)/2, (h-dh)/2, (w-dw)/2+dw, (h-dh)/2+dh)
	xdraw.CatmullRom.Scale(out, r, src, b, draw.Over, nil)
	return out
}

// cover scales and centre-crops src to exactly w×h.
func cover(src image.Image, w, h int) *image.NRGBA {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	if sw*h > sh*w {
		cw := sh * w / h
		b = image.Rect(b.Min.X+(sw-cw)/2, b.Min.Y, b.Min.X+(sw-cw)/2+cw, b.Max.Y)
	} else {
		ch := sw * h / w
		b = image.Rect(b.Min.X, b.Min.Y+(sh-ch)/2, b.Max.X, b.Min.Y+(sh-ch)/2+ch)
	}
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	xdraw.CatmullRom.Scale(out, out.Bounds(), src, b, draw.Src, nil)
	return out
}

// thumb scales src to at most n pixels on its long side.
func thumb(src image.Image, n int) *image.NRGBA {
	b := src.Bounds()
	if b.Dx() <= n && b.Dy() <= n {
		out := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(out, out.Bounds(), src, b.Min, draw.Src)
		return out
	}
	if b.Dx() >= b.Dy() {
		return fit(src, n, max(1, b.Dy()*n/b.Dx()))
	}
	return fit(src, max(1, b.Dx()*n/b.Dy()), n)
}

// crop copies r out of src into a new image with origin (0,0).
func crop(src image.Image, r image.Rectangle) *image.NRGBA {
	r = r.Intersect(src.Bounds())
	out := image.NewNRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	draw.Draw(out, out.Bounds(), src, r.Min, draw.Src)
	return out
}

// content returns the bounds of the pixels that differ from the background: alpha
// where the image has transparency, otherwise the colour of its four corners.
func content(src image.Image) image.Rectangle {
	b := src.Bounds()
	corner := func(x, y int) [4]uint32 {
		r, g, bb, a := src.At(x, y).RGBA()
		return [4]uint32{r >> 8, g >> 8, bb >> 8, a >> 8}
	}
	bg := corner(b.Min.X, b.Min.Y)
	transparent := bg[3] < 16
	differs := func(x, y int) bool {
		p := corner(x, y)
		if transparent {
			return p[3] > 24
		}
		d := 0
		for i := range 3 {
			v := int(p[i]) - int(bg[i])
			if v < 0 {
				v = -v
			}
			d += v
		}
		return d > 36
	}
	r := image.Rectangle{}
	first := true
	for y := b.Min.Y; y < b.Max.Y; y += 2 {
		for x := b.Min.X; x < b.Max.X; x += 2 {
			if !differs(x, y) {
				continue
			}
			if first {
				r, first = image.Rect(x, y, x+2, y+2), false
				continue
			}
			r = r.Union(image.Rect(x, y, x+2, y+2))
		}
	}
	if first {
		return b
	}
	return r.Intersect(b)
}
