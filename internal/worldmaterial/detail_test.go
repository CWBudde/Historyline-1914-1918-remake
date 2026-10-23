package worldmaterial

import (
	"image"
	"image/color"
	"math"
	"math/rand/v2"
	"testing"
)

// noise is an opaque image of uncorrelated grey values around v.
func noise(w, h int, v, spread float64, seed uint64) *image.NRGBA {
	r := rand.New(rand.NewPCG(seed, 1))
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			g := uint8(math.Max(0, math.Min(255, v+spread*(r.Float64()*2-1))))
			im.SetNRGBA(x, y, color.NRGBA{g, g, g, 255})
		}
	}
	return im
}

func TestLabRoundTrip(t *testing.T) {
	for _, c := range []color.NRGBA{{100, 107, 40, 255}, {0, 0, 0, 255}, {255, 255, 255, 255}, {18, 26, 16, 255}, {200, 30, 90, 255}} {
		l, a, b := toLab(c)
		if got := fromLab(l, a, b); got != c {
			t.Errorf("%v → %v", c, got)
		}
	}
	// The meadow mean is about L* 42.
	if l, _, _ := toLab(color.NRGBA{100, 107, 40, 255}); math.Abs(l-43.3) > 0.5 {
		t.Fatal(l)
	}
}

func TestBoxMeanWraps(t *testing.T) {
	v := []float64{0, 0, 0, 9, 0, 0}
	got := boxMean(v, 6, 1, 1)
	want := []float64{0, 0, 3, 3, 3, 0}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-12 {
			t.Fatal(got)
		}
	}
	if got := boxMean([]float64{9, 0, 0, 0}, 4, 1, 1); math.Abs(got[3]-3) > 1e-12 {
		t.Fatal("no wrap", got)
	}
}

func TestTiledRepeats(t *testing.T) {
	// A 6×4 period with three repeats: each copy is 2 world px wide, so at
	// density 1 the output repeats every 2 columns, and the crop keeps the
	// period aspect of the 15×8 source.
	src := noise(15, 8, 120, 60, 3)
	got, c, err := Tiled(src, image.Pt(6, 4), 1, 3, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	if got.Bounds() != image.Rect(0, 0, 6, 4) {
		t.Fatal(got.Bounds())
	}
	if c.Right-c.Left != 12 || c.Bottom-c.Top != 8 {
		t.Fatalf("crop %+v", c)
	}
	for y := range 4 {
		for x := range 4 {
			if got.NRGBAAt(x, y) != got.NRGBAAt(x+2, y) {
				t.Fatalf("no repeat at %d,%d", x, y)
			}
		}
	}
	if _, _, err := Tiled(src, image.Pt(6, 4), 1, 0, 0.5); err == nil {
		t.Fatal("zero repeats")
	}
}

func TestGrade(t *testing.T) {
	src := noise(40, 30, 110, 30, 9)
	got, err := Grade(src, 1.5, [3]float64{})
	if err != nil {
		t.Fatal(err)
	}
	m0, s0 := stats(newPlane(src).l)
	m1, s1 := stats(newPlane(got).l)
	if math.Abs(m1-m0) > 0.3 || math.Abs(s1/s0-1.5) > 0.02 {
		t.Fatalf("mean %v→%v, sigma %v→%v", m0, m1, s0, s1)
	}
	if same, _ := Grade(src, 1, [3]float64{}); same.NRGBAAt(3, 4) != src.NRGBAAt(3, 4) {
		t.Fatal("an empty grade changed the image")
	}
	// A b* offset moves the mean b* and leaves L*.
	shifted, err := Grade(solid(image.Rect(0, 0, 4, 4), color.NRGBA{100, 107, 40, 255}), 1, [3]float64{0, 0, -5})
	if err != nil {
		t.Fatal(err)
	}
	l0, _, b0 := toLab(color.NRGBA{100, 107, 40, 255})
	l1, _, b1 := toLab(shifted.NRGBAAt(1, 1))
	if math.Abs(l1-l0) > 0.5 || math.Abs(b1-b0+5) > 0.5 {
		t.Fatalf("L* %v→%v, b* %v→%v", l0, l1, b0, b1)
	}
	if _, err := Grade(src, 0, [3]float64{}); err == nil {
		t.Fatal("zero factor")
	}
	if _, err := Grade(src, 1, [3]float64{math.NaN(), 0, 0}); err == nil {
		t.Fatal("NaN offset")
	}
}

func TestCombine(t *testing.T) {
	tone := solid(image.Rect(0, 0, 40, 30), color.NRGBA{100, 107, 40, 255})
	detail := noise(40, 30, 110, 50, 7)
	same, err := Combine(tone, detail, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if same.NRGBAAt(5, 5) != tone.NRGBAAt(5, 5) {
		t.Fatal("zero gain changed the tone", same.NRGBAAt(5, 5))
	}
	got, err := Combine(tone, detail, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	// The high-pass has no mean, so the tone's mean L* stays; only L* varies.
	p, q := newPlane(tone), newPlane(got)
	mt, _ := stats(p.l)
	mg, sg := stats(q.l)
	if math.Abs(mg-mt) > 0.3 || sg < 2 {
		t.Fatalf("mean %v→%v, sigma %v", mt, mg, sg)
	}
	// Chroma stays the tone's on average; single pixels may clip at the sRGB gamut.
	ma, _ := stats(q.a)
	mb, _ := stats(q.b)
	if math.Abs(ma-p.a[0]) > 1 || math.Abs(mb-p.b[0]) > 1 {
		t.Fatal("chroma moved", p.a[0], ma, p.b[0], mb)
	}
	if _, err := Combine(tone, solid(image.Rect(0, 0, 4, 4), color.NRGBA{A: 255}), 1, 1); err == nil {
		t.Fatal("size")
	}
	if _, err := Combine(tone, detail, 1, math.NaN()); err == nil {
		t.Fatal("gain")
	}
}
