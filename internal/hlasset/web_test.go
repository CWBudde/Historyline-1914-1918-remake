package hlasset

import (
	"image"
	"image/color"
	"testing"
)

func TestBleedSpreadsColourIntoTheMaskOnly(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 12, 4))
	for y := range 4 {
		for x := 6; x < 12; x++ {
			im.SetNRGBA(x, y, color.NRGBA{200, 100, 50, 255})
		}
	}
	im.SetNRGBA(5, 1, color.NRGBA{10, 20, 30, 128}) // a soft edge keeps its own colour
	out := Bleed(im, 3)
	for y := range 4 {
		for x := range 12 {
			got, was := out.NRGBAAt(x, y), im.NRGBAAt(x, y)
			if got.A != was.A || was.A != 0 && got != was {
				t.Fatalf("(%d,%d): %v became %v", x, y, was, got)
			}
		}
	}
	if got := out.NRGBAAt(4, 0); got.A != 0 || got.R == 0 {
		t.Fatalf("mask next to the picture is %v, want its colour spread", got)
	}
	if got := out.NRGBAAt(0, 0); got != (color.NRGBA{}) {
		t.Fatalf("mask beyond the passes changed: %v", got)
	}
}

func TestEncodeAtKeepsTheMaskSharp(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	for y := range 64 {
		for x := 32; x < 64; x++ {
			im.SetNRGBA(x, y, color.NRGBA{uint8(x * 4), uint8(y * 4), 90, 255})
		}
	}
	ext, data, err := EncodeAt(im, FieldWebP)
	if err != nil || ext != ".webp" || len(data) == 0 {
		t.Fatalf("%q %d bytes: %v", ext, len(data), err)
	}
}
