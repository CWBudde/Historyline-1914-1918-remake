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
