package artreview

import (
	"image"
	"image/color"
	"testing"
)

func TestCompoundValidateSlope(t *testing.T) {
	for _, tc := range []struct {
		framing string
		width   float64
		ok      bool
	}{{"", 0, true}, {"", 56, true}, {"", 7, false}, {"", 81, false}, {"reference-layout", 56, false}, {"reference-layout", 0, true}} {
		if err := compoundValidateSlope(tc.framing, tc.width); (err == nil) != tc.ok {
			t.Errorf("%q %g: %v", tc.framing, tc.width, err)
		}
	}
}

// The fade clears the outline, keeps the middle of a large cut-out and its
// colours, and never raises alpha.
func TestCompoundSlopeFade(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 400, 300))
	for y := 50; y < 250; y++ {
		for x := 50; x < 350; x++ {
			src.SetNRGBA(x, y, color.NRGBA{uint8(x), uint8(y), 7, 255})
		}
	}
	got := compoundSlopeFade(src, .5, 20)
	for y := range 300 {
		for x := range 400 {
			s, g := src.NRGBAAt(x, y), got.NRGBAAt(x, y)
			if g.A > s.A || (s.A > 0 && (g.R != s.R || g.G != s.G || g.B != s.B)) {
				t.Fatalf("%d,%d: %v from %v", x, y, g, s)
			}
		}
	}
	for _, p := range []image.Point{{50, 150}, {349, 150}, {200, 50}, {200, 249}} {
		if a := got.NRGBAAt(p.X, p.Y).A; a != 0 {
			t.Errorf("outline %v keeps alpha %d", p, a)
		}
	}
	if a := got.NRGBAAt(200, 150).A; a != 255 {
		t.Errorf("middle alpha %d", a)
	}
}
