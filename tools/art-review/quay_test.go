package main

import (
	"image"
	"image/color"
	"math"
	"testing"
)

// quayTestStrip is a 1536x1024 strip with an opaque band of 260 rows whose
// colour varies along it, like the painting.
func quayTestStrip(t *testing.T) quayStrip {
	t.Helper()
	im := image.NewNRGBA(image.Rect(0, 0, 1536, 1024))
	for y := 382; y < 642; y++ {
		for x := range 1536 {
			im.SetNRGBA(x, y, color.NRGBA{uint8(100 + x%37), uint8(90 + (x/7)%41), 80, 255})
		}
	}
	im.SetNRGBA(10, 100, color.NRGBA{255, 0, 0, 20}) // haze
	s, err := newQuayStrip(im)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// One period on from any phase the strip repeats exactly, so each hex of a
// chain, sampled from phase zero at its centre, meets the next one.
func TestQuayStripIsPeriodic(t *testing.T) {
	s := quayTestStrip(t)
	if math.Abs(s.k*quayWidth-260) > 1e-9 || s.centre != 512 {
		t.Fatalf("strip centre %g, %g px per world px", s.centre, s.k)
	}
	for id, step := range quayAxes {
		p := step.length() * s.k
		for _, phase := range []float64{0, 1.5, p / 8, p / 4, p / 2, p - .5} {
			a, b := s.periodic(phase, p, 500), s.periodic(phase+p, p, 500)
			for i := range a {
				if math.Abs(a[i]-b[i]) > 1e-9 {
					t.Fatalf("%s phase %g: %v vs %v", id, phase, a, b)
				}
			}
		}
		// Approaching the period from below meets phase zero.
		a, b := s.periodic(p-1e-6, p, 500), s.periodic(0, p, 500)
		for i := range a {
			if math.Abs(a[i]-b[i]) > 1e-3 {
				t.Fatalf("%s seam: %v vs %v", id, a, b)
			}
		}
	}
	if c := s.sample(10.5, 100.5); c[3] != 0 {
		t.Errorf("haze kept: %v", c)
	}
}

// The vertical quay crosses its hex 26 world px wide on the centre line, from
// the top edge to the bottom edge, and leaves the corners clear.
func TestQuayTilesCrossTheHex(t *testing.T) {
	s := quayTestStrip(t)
	for id := range quayAxes {
		tiles, err := quayTiles(s, id)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range []int{2, 4} {
			im := tiles[d]
			if im.Bounds() != image.Rect(0, 0, 96*d, 80*d) {
				t.Fatalf("%s %dx bounds %v", id, d, im.Bounds())
			}
			if a := im.NRGBAAt(48*d, 40*d).A; a != 255 {
				t.Errorf("%s %dx centre alpha %d", id, d, a)
			}
			if a := im.NRGBAAt(0, 0).A; a != 0 {
				t.Errorf("%s %dx outside the hex alpha %d", id, d, a)
			}
		}
		if id != "quay36" {
			continue
		}
		im := tiles[4]
		opaque := 0
		for x := range 96 * 4 {
			if im.NRGBAAt(x, 40*4).A == 255 {
				opaque++
			}
		}
		if w := float64(opaque) / 4; math.Abs(w-quayWidth) > 1 {
			t.Errorf("quay36 is %g world px wide, want %g", w, quayWidth)
		}
		for _, y := range []int{1, 79} {
			if a := im.NRGBAAt(48*4, y*4).A; a != 255 {
				t.Errorf("quay36 does not reach the port at y %d: alpha %d", y, a)
			}
		}
		// The shadow falls on the lower right only.
		beside := int(quayWidth/2) + 1
		left, right := im.NRGBAAt((48-beside)*4, 40*4).A, im.NRGBAAt((48+beside)*4, 40*4).A
		if left != 0 || right == 0 || right == 255 {
			t.Errorf("shadow beside the pier: left %d, right %d", left, right)
		}
	}
}
