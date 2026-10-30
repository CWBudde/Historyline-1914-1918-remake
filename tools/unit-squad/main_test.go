package main

import (
	"image"
	"image/color"
	"testing"
)

// fixture is an atlas with three solid 80x120 men in every cell.
func fixture() *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, 1024, 1536))
	for f := range cells {
		ox, oy := f%2*512, f/2*512
		for m := range men {
			for y := 0; y < 120; y++ {
				for x := 0; x < 80; x++ {
					im.SetNRGBA(ox+20+m*150+x, oy+200+y, color.NRGBA{200, 40, 120, 255})
				}
			}
		}
	}
	return im
}

func TestAtlasShowsNMen(t *testing.T) {
	for n := 1; n <= maxCount; n++ {
		out, err := atlas(fixture(), n, .5)
		if err != nil {
			t.Fatal(err)
		}
		for f := range cells {
			if f != 0 && f != 3 {
				continue // the boxes of the fixture touch once the formation turns
			}
			cell := image.NewNRGBA(image.Rect(0, 0, 512, 512))
			for y := range 512 {
				for x := range 512 {
					cell.SetNRGBA(x, y, out.NRGBAAt(f%2*512+x, f/2*512+y))
				}
			}
			if _, err := cutMen(cell, n); err != nil {
				t.Errorf("n=%d cell %d: %v", n, f, err)
			}
		}
	}
}
