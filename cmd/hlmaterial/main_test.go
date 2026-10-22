package main

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func TestExportMaterial(t *testing.T) {
	src := image.NewNRGBA(image.Rect(3, 7, 13, 15))
	for y := 7; y < 15; y++ {
		for x := 3; x < 13; x++ {
			src.SetNRGBA(x, y, color.NRGBA{uint8(x * 10), uint8(y * 5), 80, 255})
		}
	}
	got, c, err := exportMaterial(src, image.Pt(6, 4), 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.Bounds() != image.Rect(0, 0, 12, 8) {
		t.Fatal(got.Bounds())
	}
	if math.Abs(c.Top-2.0/3) > 1e-12 || c.Left != 0 {
		t.Fatalf("crop %+v", c)
	}
	for y := 0; y < 8; y++ {
		if got.NRGBAAt(0, y) != got.NRGBAAt(11, y) {
			t.Fatal("nonperiodic x")
		}
	}
	for x := 0; x < 12; x++ {
		if got.NRGBAAt(x, 0) != got.NRGBAAt(x, 7) {
			t.Fatal("nonperiodic y")
		}
	}
}
