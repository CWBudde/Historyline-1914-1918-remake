package gesture

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	xvector "golang.org/x/image/vector"
)

func area(p [][2]float64) float64 {
	var a float64
	for i := range p {
		j := (i + 1) % len(p)
		a += p[i][0]*p[j][1] - p[j][0]*p[i][1]
	}
	return a / 2
}

// Every glyph is non-empty, fits its unit box and has its solid parts wound
// one way and its holes the other, so a non-zero fill cuts the holes.
func TestGlyphs(t *testing.T) {
	for a := Cancel; a <= Repair; a++ {
		polys := glyph(a)
		if len(polys) == 0 {
			t.Errorf("%v: no glyph", a)
		}
		var solid float64
		for _, p := range polys {
			for _, pt := range p.pts {
				if math.Abs(pt[0]) > 0.5+1e-9 || math.Abs(pt[1]) > 0.5+1e-9 {
					t.Errorf("%v: point %v outside the unit box", a, pt)
				}
			}
			ar := area(p.pts)
			if p.hole != (ar < 0) || ar == 0 {
				t.Errorf("%v: polygon area %v, hole %v", a, ar, p.hole)
			}
			solid += ar
		}
		if solid < 0.08 {
			t.Errorf("%v: glyph covers only %v of its box", a, solid)
		}
	}
	if len(glyph(Action(99))) != 0 {
		t.Error("glyph for an unknown action")
	}
}

func TestInset(t *testing.T) {
	c := PlateCorners(0, 0, 1)
	in := inset(c[:], 10)
	// The top edge moves down by 10, the left vertex inwards along the axis by
	// 10 / sin of the slant's angle.
	if !near(in[1][1], -30) || !near(in[2][1], -30) || !near(in[4][1], 30) {
		t.Errorf("inset top/bottom = %v %v %v", in[1], in[2], in[4])
	}
	slant := math.Atan2(40, 25) // angle of the slanted edge against the horizontal
	if want := -48 + 10/math.Sin(slant); !near(in[0][0], want) || !near(in[0][1], 0) {
		t.Errorf("inset left vertex = %v, want (%v, 0)", in[0], want)
	}
	if area(in) <= 0 || area(in) >= area(c[:]) {
		t.Errorf("inset area %v of %v", area(in), area(c[:]))
	}
}

// Drawing needs no running game: the commands queue on offscreen images. The
// pixels cannot be read back in a plain test, so this only checks that every
// piece draws at small and large scales without panicking.
func TestDrawSmoke(t *testing.T) {
	dst := ebiten.NewImage(256, 256)
	for _, s := range []float64{0.25, 0.5, 1, 2} {
		DrawRing(dst, 128, 128, s)
		for a := Cancel; a <= Repair; a++ {
			for st := Normal; st <= Disabled; st++ {
				DrawPlate(dst, 128, 128, s, a, st)
			}
		}
		r := NewRing(128, 128, s, ForMap(false, true).RingActions())
		for hover := -1; hover < len(r.Actions); hover++ {
			DrawRingMenu(dst, r, hover)
		}
	}
	if Normal.String() != "normal" || State(7).String() != "State(7)" {
		t.Error("state names")
	}
}

// raster fills the layers in software, as draw does on the GPU, so their
// pixels can be checked in a plain test.
func raster(ls []layer, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for _, l := range ls {
		z := xvector.NewRasterizer(w, h)
		for _, p := range l.polys {
			z.MoveTo(float32(p[0][0]), float32(p[0][1]))
			for _, pt := range p[1:] {
				z.LineTo(float32(pt[0]), float32(pt[1]))
			}
			z.ClosePath()
		}
		z.Draw(dst, dst.Bounds(), image.NewUniform(l.c), image.Point{})
	}
	return dst
}

func TestRingPixels(t *testing.T) {
	const s = 1 // a 96×80 cell
	img := raster(ringLayers(60, 50, s), 120, 100)
	tests := []struct {
		name string
		x, y int
		want color.RGBA
	}{
		{"inside is transparent", 60, 50, color.RGBA{}},
		{"inside near the band is transparent", 60, 50 - 40 + 15, color.RGBA{}},
		{"outside is transparent", 2, 2, color.RGBA{}},
		{"outline at the bottom edge", 60, 50 + 40 - 1, outline},
		{"inner line", 60, 50 + 40 - 13, outline},
	}
	for _, tt := range tests {
		if got := img.RGBAAt(tt.x, tt.y); got != tt.want {
			t.Errorf("%s: pixel (%d, %d) = %v, want %v", tt.name, tt.x, tt.y, got, tt.want)
		}
	}
	// The rim is lit from the top left: the outer part of the top edge and the
	// inner part of the bottom edge are lighter than plain steel, the outer
	// part of the bottom edge and the inner part of the top edge darker.
	luma := func(x, y int) int { c := img.RGBAAt(x, y); return int(c.R) + int(c.G) + int(c.B) }
	base := int(steel.R) + int(steel.G) + int(steel.B)
	for _, tt := range []struct {
		name    string
		x, y    int
		lighter bool
	}{
		{"outer top rim", 60, 50 - 40 + 2, true},
		{"inner top rim", 60, 50 - 40 + 10, false},
		{"outer bottom rim", 60, 50 + 40 - 3, false},
		{"inner bottom rim", 60, 50 + 40 - 11, true},
	} {
		if l := luma(tt.x, tt.y); (l > base) != tt.lighter || l == base {
			t.Errorf("%s: pixel (%d, %d) = %v, lighter than steel %v", tt.name, tt.x, tt.y, img.RGBAAt(tt.x, tt.y), tt.lighter)
		}
	}
}

// A plate covers its cell polygon and nothing outside it, at every scale and
// state; its centre is never transparent.
func TestPlatePixels(t *testing.T) {
	for _, s := range []float64{0.25, 0.5, 1} {
		w, h := int(120*s)+4, int(100*s)+4
		cx, cy := float64(w)/2, float64(h)/2
		for a := Cancel; a <= Repair; a++ {
			for st := Normal; st <= Disabled; st++ {
				img := raster(plateLayers(cx, cy, s, a, st), w, h)
				for y := range h {
					for x := range w {
						px, py := float64(x)+0.5, float64(y)+0.5
						in := InPlate(px, py, cx, cy, s)
						edge := InPlate(px, py, cx, cy, s+1.5/40) && !InPlate(px, py, cx, cy, s-1.5/40)
						if got := img.RGBAAt(x, y); !edge && (got.A != 0) != in {
							t.Fatalf("scale %v %v %v: pixel (%d, %d) = %v, in plate %v", s, a, st, x, y, got, in)
						}
					}
				}
				if img.RGBAAt(int(cx), int(cy)).A != 255 {
					t.Errorf("scale %v %v %v: centre not opaque", s, a, st)
				}
			}
		}
	}
	// The glyph shows: the normal plate has face, highlight and shadow pixels.
	img := raster(plateLayers(60, 50, 1, Move, Normal), 120, 100)
	count := map[color.RGBA]int{}
	for y := range 100 {
		for x := range 120 {
			count[img.RGBAAt(x, y)]++
		}
	}
	for _, c := range []color.RGBA{face, light, shadow, steel, speck, outline} {
		if count[c] < 20 {
			t.Errorf("colour %v on %d pixels", c, count[c])
		}
	}
}
