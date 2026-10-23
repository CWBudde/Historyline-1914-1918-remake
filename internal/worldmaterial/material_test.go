package worldmaterial

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func solid(r image.Rectangle, c color.NRGBA) *image.NRGBA {
	p := image.NewNRGBA(r)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			p.SetNRGBA(x, y, c)
		}
	}
	return p
}
func TestExportCropAndArea(t *testing.T) {
	src := solid(image.Rect(7, 9, 12, 13), color.NRGBA{A: 255})
	for y := 9; y < 13; y++ {
		for x := 7; x < 12; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: uint8((x - 7) * 50), A: 255})
		}
	}
	out, c, err := Export(src, image.Pt(1, 1), 1)
	if err != nil {
		t.Fatal(err)
	}
	if c.Left != .5 || c.Right != 4.5 || c.Top != 0 || c.Bottom != 4 {
		t.Fatalf("crop %+v", c)
	}
	if out.NRGBAAt(0, 0).R != 100 {
		t.Fatal(out.NRGBAAt(0, 0))
	}
	checker := solid(image.Rect(0, 0, 4, 4), color.NRGBA{A: 255})
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if (x+y)%2 == 0 {
				checker.SetNRGBA(x, y, color.NRGBA{255, 255, 255, 255})
			}
		}
	}
	p, _, err := Export(checker, image.Pt(1, 1), 1)
	if err != nil || p.NRGBAAt(0, 0).R != 128 {
		t.Fatalf("area %v %v", p, err)
	}
}
func TestSeamlessWrapsAndKeepsCentre(t *testing.T) {
	// A ramp jumps from its last column back to its first; the edges must take the
	// half-shifted neighbours instead, so the image continues across the wrap.
	src := image.NewNRGBA(image.Rect(3, 4, 15, 16))
	for y := 4; y < 16; y++ {
		for x := 3; x < 15; x++ {
			src.SetNRGBA(x, y, color.NRGBA{uint8(20 * (x - 3)), uint8(20 * (y - 4)), 40, 255})
		}
	}
	if h, v := continues(src, src.Bounds()); h || v {
		t.Fatal("ramp continues", h, v)
	}
	out, err := Seamless(src, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	if h, v := continues(out, out.Bounds()); !h || !v {
		t.Fatal("wrap", h, v)
	}
	if out.NRGBAAt(0, 6).R != src.NRGBAAt(3+6, 10).R || out.NRGBAAt(11, 6).R != src.NRGBAAt(3+5, 10).R {
		t.Fatal("edges are not the half-shifted columns", out.NRGBAAt(0, 6), out.NRGBAAt(11, 6))
	}
	if out.NRGBAAt(6, 6) != src.NRGBAAt(9, 10) || src.NRGBAAt(3, 4).R != 0 {
		t.Fatal("centre or source changed")
	}
	if _, err := New(out, out, image.Pt(12, 12), 1); err != nil {
		t.Fatal(err)
	}
}

func TestSpreadBlendKeepsContrast(t *testing.T) {
	mean := [3]float64{100, 100, 100}
	got := spreadBlend(color.NRGBA{110, 90, 100, 255}, color.NRGBA{110, 110, 100, 255}, .5, mean)
	// Equal deviations add to √2 times one deviation; opposite ones cancel.
	if got.R != 114 || got.G != 100 || got.B != 100 || got.A != 255 {
		t.Fatal(got)
	}
}
func TestSamplerPhaseMaskAndCoordinates(t *testing.T) {
	grass := solid(image.Rect(4, 5, 8, 9), color.NRGBA{20, 40, 60, 255})
	scrub := solid(image.Rect(-2, -3, 2, 1), color.NRGBA{100, 120, 140, 255})
	s, err := New(grass, scrub, image.Pt(2, 2), 2)
	if err != nil {
		t.Fatal(err)
	}
	if s.Phase(image.Pt(-5, 6)) != image.Pt(3, 2) {
		t.Fatal(s.Phase(image.Pt(-5, 6)))
	}
	mask := solid(image.Rect(8, 9, 10, 10), color.NRGBA{A: 128})
	mask.SetNRGBA(9, 9, color.NRGBA{})
	calls := 0
	out, err := s.Tile(mask, image.Pt(-1, 3), func(x, y float64) float64 {
		calls++
		if x != -.25 || y != 1.75 {
			t.Fatalf("coordinates %v %v", x, y)
		}
		return .25
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || out.Bounds() != image.Rect(0, 0, 2, 1) || out.NRGBAAt(0, 0) != (color.NRGBA{40, 60, 80, 128}) || out.NRGBAAt(1, 0) != (color.NRGBA{}) {
		t.Fatalf("tile %v", out.Pix)
	}
}
func TestInvalidInputs(t *testing.T) {
	opaque := solid(image.Rect(0, 0, 4, 4), color.NRGBA{A: 255})
	bad := solid(image.Rect(0, 0, 4, 4), color.NRGBA{})
	for _, p := range []image.Point{{0, 1}, {-1, 1}, {math.MaxInt, 2}, {6000, 6000}} {
		if _, _, err := Export(opaque, p, 1); err == nil {
			t.Fatal("invalid size", p)
		}
	}
	if _, _, err := Export(bad, image.Pt(1, 1), 1); err == nil {
		t.Fatal("alpha")
	}
	for _, b := range []float64{0, -1, 2, math.NaN(), math.Inf(1)} {
		if _, err := Seamless(opaque, b, 1); err == nil {
			t.Fatal("band", b)
		}
	}
	for y := range 4 {
		for x := range 4 {
			opaque.SetNRGBA(x, y, color.NRGBA{R: uint8(50 * x), A: 255})
		}
	}
	if _, err := New(opaque, opaque, image.Pt(4, 4), 1); err == nil {
		t.Fatal("nonperiodic")
	}
	s, err := New(bad, bad, image.Pt(4, 4), 1)
	if err == nil || s != nil {
		t.Fatal("transparent")
	}
	good := solid(image.Rect(0, 0, 4, 4), color.NRGBA{A: 255})
	s, err = New(good, good, image.Pt(4, 4), 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []float64{-1, 2, math.NaN(), math.Inf(1)} {
		if _, err := s.Tile(good, image.Point{}, func(float64, float64) float64 { return v }); err == nil {
			t.Fatal("weight", v)
		}
	}
}

func TestFractionalFootprints(t *testing.T) {
	src := solid(image.Rect(0, 0, 3, 2), color.NRGBA{A: 255})
	for y := 0; y < 2; y++ {
		src.SetNRGBA(0, y, color.NRGBA{R: 0, A: 255})
		src.SetNRGBA(1, y, color.NRGBA{R: 60, A: 255})
		src.SetNRGBA(2, y, color.NRGBA{R: 120, A: 255})
	}
	out, c, err := Export(src, image.Pt(1, 1), 2)
	if err != nil {
		t.Fatal(err)
	}
	if c != (Crop{.5, 0, 2.5, 2}) || out.NRGBAAt(0, 0).R != 30 || out.NRGBAAt(1, 0).R != 90 {
		t.Fatalf("fractional area %+v %v", c, out.Pix)
	}
}
func TestGlobalWrapAndNilWeight(t *testing.T) {
	src := solid(image.Rect(-4, 8, 1, 13), color.NRGBA{R: 10, A: 255})
	for y := 9; y < 12; y++ {
		for x := -3; x < 0; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: uint8((x + 4) * 30), A: 255})
		}
	}
	s, err := New(src, src, image.Pt(5, 5), 1)
	if err != nil {
		t.Fatal(err)
	}
	mask := solid(image.Rect(0, 0, 9, 8), color.NRGBA{A: 255})
	a, err := s.Tile(mask, image.Pt(-3, -2), nil)
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.Tile(mask, image.Pt(12, 13), nil)
	if err != nil {
		t.Fatal(err)
	}
	for y := 0; y < 8; y++ {
		for x := 0; x < 9; x++ {
			want := src.NRGBAAt(-4+mod(x-3, 5), 8+mod(y-2, 5))
			if a.NRGBAAt(x, y) != want || b.NRGBAAt(x, y) != want {
				t.Fatalf("global wrap %d,%d", x, y)
			}
		}
	}
}
