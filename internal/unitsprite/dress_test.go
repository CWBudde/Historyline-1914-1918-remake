package unitsprite

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

// square is a density-d sprite of a 20×20 logical canvas with an opaque red
// square from logical 5 to 15.
func square(d int) *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, 20*d, 20*d))
	draw.Draw(im, image.Rect(5*d, 5*d, 15*d, 15*d), image.NewUniform(color.NRGBA{R: 255, A: 255}), image.Point{}, draw.Src)
	return im
}

func TestDressKeepsTheSpriteAndGrowsTheCanvas(t *testing.T) {
	for _, d := range []int{2, 4} {
		src := square(d)
		out := Dress(src, d, false, Dark)
		m := DressMargin(d)
		if m%d != 0 {
			t.Fatalf("%dx margin %d is not whole logical pixels", d, m)
		}
		if got, want := out.Bounds(), image.Rect(0, 0, 20*d+2*m, 20*d+2*m); got != want {
			t.Fatalf("%dx bounds %v want %v", d, got, want)
		}
		for y := 5 * d; y < 15*d; y++ {
			for x := 5 * d; x < 15*d; x++ {
				if got := out.NRGBAAt(x+m, y+m); got != (color.NRGBA{R: 255, A: 255}) {
					t.Fatalf("%dx sprite texel %d,%d became %v", d, x, y, got)
				}
			}
		}
	}
}

func TestDressRimIsOneAndAHalfWorldPixels(t *testing.T) {
	for _, rim := range []Rim{Dark, Light} {
		for _, d := range []int{2, 4} {
			out, m := Dress(square(d), d, false, rim), DressMargin(d)
			y, width := 10*d+m, 3*d/2
			for i := 1; i <= width; i++ {
				if got := out.NRGBAAt(5*d+m-i, y); got != rim.colour() {
					t.Errorf("rim %d %dx: texel %d left of the sprite is %v, want %v", rim, d, i, got, rim.colour())
				}
			}
			if got := out.NRGBAAt(5*d+m-width-1, y); got.A >= 128 {
				t.Errorf("rim %d %dx is wider than 1.5 world px: %v", rim, d, got)
			}
		}
	}
}

// The light rim, for units in forest and on water, stands out against the dark
// ground without reading as a white sticker.
func TestLightRimIsPale(t *testing.T) {
	l := func(c color.NRGBA) float64 { return 0.2126*float64(c.R) + 0.7152*float64(c.G) + 0.0722*float64(c.B) }
	if dark, light := l(Dark.colour()), l(Light.colour()); dark > 30 || light < 170 || light > 235 {
		t.Errorf("rim luma dark %.0f, light %.0f: want dark ≤ 30 and a muted pale light, 170–235", dark, light)
	}
}

func TestDressShadowFallsToTheLowerRight(t *testing.T) {
	for _, flying := range []bool{false, true} {
		d := 2
		out, m := Dress(square(d), d, flying, Dark), DressMargin(d)
		var upperLeft, lowerRight int
		for y := range out.Bounds().Dy() {
			for x := range out.Bounds().Dx() {
				lx, ly := float64(x-m)/float64(d), float64(y-m)/float64(d)
				if lx >= 3 && lx < 17 && ly >= 3 && ly < 17 {
					continue // sprite and rim
				}
				a := int(out.NRGBAAt(x, y).A)
				if a >= 128 {
					t.Fatalf("flying %v: shadow texel %d,%d has alpha %d, would count as the unit", flying, x, y, a)
				}
				if lx+ly < 20 {
					upperLeft += a
				} else {
					lowerRight += a
				}
			}
		}
		if lowerRight == 0 || upperLeft*4 > lowerRight {
			t.Errorf("flying %v: shadow alpha upper left %d, lower right %d", flying, upperLeft, lowerRight)
		}
	}
}

func TestFlyingShadowSitsApart(t *testing.T) {
	// The alpha-weighted centre of the bare shadow texels moves further to the
	// lower right for a flying unit.
	centre := func(flying bool) float64 {
		out := Dress(square(2), 2, flying, Dark)
		var sum, weight float64
		for y := range out.Bounds().Dy() {
			for x := range out.Bounds().Dx() {
				if a := float64(out.NRGBAAt(x, y).A); a < 128 {
					sum += a * float64(x+y)
					weight += a
				}
			}
		}
		return sum / weight
	}
	if ground, air := centre(false), centre(true); air <= ground+4 {
		t.Errorf("flying shadow centre %.1f, grounded %.1f", air, ground)
	}
	if !Flies("spad-xiii") || Flies("infantry") || Flies("static-balloon") {
		t.Error("Flies follows the aircraft movement type")
	}
}

func TestDressEmptySpriteStaysEmpty(t *testing.T) {
	out := Dress(image.NewNRGBA(image.Rect(0, 0, 8, 8)), 2, false, Light)
	for i := 3; i < len(out.Pix); i += 4 {
		if out.Pix[i] != 0 {
			t.Fatal("an empty sprite gained a rim or shadow")
		}
	}
}

// A 4x unit sprite is dressed once, at its first draw.
func BenchmarkDress4x(b *testing.B) {
	src := image.NewNRGBA(image.Rect(0, 0, 384, 384))
	draw.Draw(src, image.Rect(96, 96, 288, 288), image.NewUniform(color.NRGBA{R: 255, A: 255}), image.Point{}, draw.Src)
	for b.Loop() {
		Dress(src, 4, false, Dark)
	}
}

func TestRimForDarkGround(t *testing.T) {
	for _, c := range []struct {
		terrain string
		infra   bool
		want    Rim
	}{{"forest", false, Light}, {"forest_w", false, Light}, {"water", false, Light}, {"waterStones", false, Light}, {"forest", true, Dark}, {"water", true, Dark}, {"grass", false, Dark}, {"grass_w", false, Dark}, {"lakeside12", false, Dark}} {
		if got := RimFor(c.terrain, c.infra); got != c.want {
			t.Errorf("RimFor(%q, %v) = %d, want %d", c.terrain, c.infra, got, c.want)
		}
	}
}
