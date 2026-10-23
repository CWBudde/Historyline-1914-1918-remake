package dospanel

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func TestSurfacesKeepThePalette(t *testing.T) {
	t.Cleanup(func() { SetSurfaces(Surfaces{}) })
	p := painter{image.NewRGBA(image.Rect(0, 0, 64, 64)), 2}

	// A flat surface has no texture: the steel is the ramp itself.
	flat := image.NewUniform(color.RGBA{0x60, 0x50, 0x40, 0xFF})
	SetSurfaces(Surfaces{Steel: subImage(flat, 16)})
	if got, want := p.steelColour(5, 7, toneBack), ramp(toneBack); got != want {
		t.Errorf("flat steel %v, want the ramp's %v", got, want)
	}

	// A checker averages out to the ramp over a tile.
	checker := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := range 64 {
		for x := range 64 {
			v := uint8(80)
			if (x/4+y/4)%2 == 0 {
				v = 120
			}
			checker.SetRGBA(x, y, color.RGBA{v, v, v, 0xFF})
		}
	}
	SetSurfaces(Surfaces{Steel: checker, Ground: checker})
	n := int(steelSpan * p.k)
	var sum float64
	for y := range n {
		for x := range n {
			sum += float64(p.steelColour(x, y, toneBack).G)
		}
	}
	if got, want := sum/float64(n*n), float64(ramp(toneBack).G); math.Abs(got-want) > 1 {
		t.Errorf("textured steel averages %.1f, want the ramp's %.0f", got, want)
	}
	if a, b := p.groundColour(0, 0, 0.3), p.groundColour(int(2*p.k), 0, 0.3); a == b {
		t.Errorf("ground shows no texture: %v at both", a)
	}

	// Without surfaces the drawn grain is back.
	SetSurfaces(Surfaces{})
	if surfaces.steel != nil || surfaces.ground != nil {
		t.Error("surfaces kept")
	}
}

// subImage is an n×n image of src.
func subImage(src image.Image, n int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, n, n))
	for y := range n {
		for x := range n {
			img.Set(x, y, src.At(x, y))
		}
	}
	return img
}

func TestRivetSheetHeads(t *testing.T) {
	t.Cleanup(func() { SetSurfaces(Surfaces{}) })
	// A sheet of flat discs of radius 10, off centre in their 40×40 cells,
	// over a faint red haze.
	sheet := image.NewNRGBA(image.Rect(0, 0, 40*rivetCols, 40*rivetRows))
	for y := range sheet.Rect.Dy() {
		for x := range sheet.Rect.Dx() {
			c := color.NRGBA{0xFF, 0, 0, 20}
			if math.Hypot(float64(x%40)+0.5-22, float64(y%40)+0.5-18) < 10 {
				c = color.NRGBA{0x50, 0x40, 0x20, 0xFF}
			}
			sheet.SetNRGBA(x, y, c)
		}
	}
	SetSurfaces(Surfaces{Rivets: sheet})
	s := surfaces.rivets
	if s == nil || len(s.heads) != rivetCols*rivetRows {
		t.Fatalf("sheet %v, want %d heads", s, rivetCols*rivetRows)
	}
	for i, h := range s.heads {
		cx, cy := float64(i%rivetCols*40+22), float64(i/rivetCols*40+18)
		if math.Abs(h.cx-cx) > 0.1 || math.Abs(h.cy-cy) > 0.1 || math.Abs(h.r-10) > 0.3 {
			t.Errorf("head %d at (%.1f, %.1f) r %.1f, want (%.0f, %.0f) r 10", i, h.cx, h.cy, h.r, cx, cy)
		}
	}
	// A flat head keeps its hue at the brass's lightness, to its rim, and no
	// haze shows.
	flat := s.at(s.heads[3], 0, 0, 0.3, brass)
	if got, want := luma(float64(flat.R), float64(flat.G), float64(flat.B)), luma(float64(brass.R), float64(brass.G), float64(brass.B)); math.Abs(got-want) > 1 {
		t.Errorf("flat head %v has lightness %.1f, want %.1f", flat, got, want)
	}
	if math.Abs(float64(flat.R)/float64(flat.B)-2.5) > 0.05 {
		t.Errorf("flat head %v lost its hue", flat)
	}
	for _, uv := range [][2]float64{{-0.9, 0}, {0.6, 0.6}} {
		if got := s.at(s.heads[3], uv[0], uv[1], 0.3, brass); got != flat {
			t.Errorf("flat head at %v is %v, want %v", uv, got, flat)
		}
	}
	// The painted head is drawn inside the rivet's circle.
	p := painter{image.NewRGBA(image.Rect(0, 0, 32, 32)), 4}
	p.rivet(16, 16, 6)
	if got := p.img.RGBAAt(16, 16); got != flat {
		t.Errorf("rivet centre %v, want %v", got, flat)
	}
	if got := p.img.RGBAAt(2, 2); got != (color.RGBA{}) {
		t.Errorf("rivet reached (2, 2): %v", got)
	}

	SetSurfaces(Surfaces{})
	if surfaces.rivets != nil {
		t.Error("rivets kept")
	}
}
