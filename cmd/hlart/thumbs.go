package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"slices"
)

// thumbs writes a PNG no larger than limit pixels on its long side for every
// master and runtime sample in idx to out/<path>. Up-to-date thumbnails are kept.
func thumbs(root, out string, idx *Index, limit int) (int, error) {
	n := 0
	for _, a := range idx.Assets {
		paths := slices.Clone(a.Runtime.Samples)
		for _, m := range a.Masters {
			paths = append(paths, m.Path)
		}
		for _, p := range paths {
			src := filepath.Join(root, filepath.FromSlash(p))
			dst := filepath.Join(out, filepath.FromSlash(p))
			if fresh(dst, src) {
				continue
			}
			if err := thumb(src, dst, limit); err != nil {
				return n, fmt.Errorf("%s: %w", p, err)
			}
			n++
		}
	}
	return n, nil
}

func fresh(dst, src string) bool {
	d, err := os.Stat(dst)
	if err != nil {
		return false
	}
	s, err := os.Stat(src)
	return err == nil && !d.ModTime().Before(s.ModTime())
}

func thumb(src, dst string, limit int) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	img, _, err := image.Decode(f)
	_ = f.Close()
	if err != nil {
		return err
	}
	out := shrink(img, limit)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	enc := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := enc.Encode(w, out); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

// shrink box-filters img down to at most limit pixels on its long side.
func shrink(img image.Image, limit int) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= limit && h <= limit {
		return img
	}
	tw, th := limit, h*limit/w
	if h > w {
		tw, th = w*limit/h, limit
	}
	tw, th = max(tw, 1), max(th, 1)
	src := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(src, src.Bounds(), img, b.Min, draw.Src)
	dst := image.NewRGBA(image.Rect(0, 0, tw, th))
	for y := range th {
		y0, y1 := y*h/th, max((y+1)*h/th, y*h/th+1)
		for x := range tw {
			x0, x1 := x*w/tw, max((x+1)*w/tw, x*w/tw+1)
			var sum [4]int
			for sy := y0; sy < y1; sy++ {
				row := src.Pix[sy*src.Stride:]
				for sx := x0; sx < x1; sx++ {
					p := row[sx*4 : sx*4+4]
					sum[0] += int(p[0])
					sum[1] += int(p[1])
					sum[2] += int(p[2])
					sum[3] += int(p[3])
				}
			}
			cnt := (y1 - y0) * (x1 - x0)
			o := dst.Pix[y*dst.Stride+x*4:]
			for c := range 4 {
				o[c] = uint8(sum[c] / cnt)
			}
		}
	}
	return dst
}
