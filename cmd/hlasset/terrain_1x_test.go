package main

import (
	"image"
	"image/color"
	"testing"
)

func TestHalveAveragesAreasInLinearLight(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 4, 2))
	white, black := color.NRGBA{255, 255, 255, 255}, color.NRGBA{0, 0, 0, 255}
	// Left block: a black and white checker. Right block: two red texels
	// beside two transparent ones.
	src.Set(0, 0, white)
	src.Set(1, 1, white)
	src.Set(0, 1, black)
	src.Set(1, 0, black)
	src.Set(2, 0, color.NRGBA{200, 0, 0, 255})
	src.Set(3, 1, color.NRGBA{200, 0, 0, 255})
	got := halve(src)
	if got.Rect.Dx() != 2 || got.Rect.Dy() != 1 {
		t.Fatalf("halved to %v", got.Rect)
	}
	// Half the light of white, not half its sRGB value (128).
	if c := got.NRGBAAt(0, 0); c != (color.NRGBA{188, 188, 188, 255}) {
		t.Errorf("checker: %v", c)
	}
	// Transparent texels lend no colour, only coverage.
	if c := got.NRGBAAt(1, 0); c != (color.NRGBA{200, 0, 0, 128}) {
		t.Errorf("half covered: %v", c)
	}
}
