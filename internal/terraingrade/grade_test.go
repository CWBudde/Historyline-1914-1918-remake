package terraingrade

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func flat(c color.NRGBA, w, h int) *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(im.Pix); i += 4 {
		im.Pix[i], im.Pix[i+1], im.Pix[i+2], im.Pix[i+3] = c.R, c.G, c.B, c.A
	}
	return im
}

func lab(c color.NRGBA) (l, a, b float64) { return toLab(c.R, c.G, c.B) }

// The land curve stretches the summer map's 5–95 % L* of 15–53 to about 8–64,
// lifts the meadow to 42–45 and leaves snow as painted.
func TestLandCurve(t *testing.T) {
	for _, tc := range []struct{ in, lo, hi float64 }{{15, 4, 8}, {41, 42, 45}, {53, 62, 66}, {88, 88, 88}, {0, 0, 0}, {100, 100, 100}} {
		if got := landL.at(tc.in); got < tc.lo-1e-9 || got > tc.hi+1e-9 {
			t.Errorf("land L* %v → %.2f, want %v–%v", tc.in, got, tc.lo, tc.hi)
		}
	}
	for l := 0.0; l < 100; l += 0.5 {
		if landL.at(l+0.5) < landL.at(l) || waterL.at(l+0.5) < waterL.at(l) {
			t.Fatalf("curves fall at L* %v", l)
		}
	}
}

// Land keeps its hue but is 4–5 b* less yellow; open water is darkened to the
// deep-water target and keeps its b*; snow and alpha stay as painted.
func TestPixelClasses(t *testing.T) {
	meadow := color.NRGBA{100, 107, 40, 255}
	water := color.NRGBA{60, 105, 125, 255}
	snow := color.NRGBA{226, 230, 236, 128}
	for _, tc := range []struct {
		name       string
		c          color.NRGBA
		dl, db     [2]float64 // allowed change of L* and b*
		wantAlpha  uint8
		identicalA bool
	}{
		{"meadow", meadow, [2]float64{1, 5}, [2]float64{-5, -4}, 255, true},
		{"water", water, [2]float64{-9, -5}, [2]float64{-0.6, 0.6}, 255, true},
		{"snow", snow, [2]float64{-0.6, 0.6}, [2]float64{-0.6, 0.6}, 128, true},
	} {
		got := Grade(flat(tc.c, 4, 4), "tile/x", "terrain/grass", nil).NRGBAAt(1, 1)
		l0, _, b0 := lab(tc.c)
		l1, _, b1 := lab(got)
		if dl, db := l1-l0, b1-b0; dl < tc.dl[0] || dl > tc.dl[1] || db < tc.db[0] || db > tc.db[1] || got.A != tc.wantAlpha {
			t.Errorf("%s %v → %v: ΔL* %.2f, Δb* %.2f, alpha %d", tc.name, tc.c, got, dl, db, got.A)
		}
	}
}

// Shadows, fog and other kinds are left as painted.
func TestUngraded(t *testing.T) {
	src := flat(color.NRGBA{100, 107, 40, 200}, 3, 3)
	for _, tc := range []struct{ id, art string }{
		{"standing/shadow/forest", "terrain/forest"},
		{"standing/shadow/scatter/crown/oak-16", "terrain/forest"},
		{"tile/unseen", "terrain/fog"},
		{"unit/a7v/blue/n-normal", "units/a7v"},
	} {
		if Applies(tc.id, tc.art) {
			t.Errorf("%s is graded", tc.id)
		}
		if got := Grade(src, tc.id, tc.art, nil); got.NRGBAAt(1, 1) != src.NRGBAAt(1, 1) {
			t.Errorf("%s changed: %v", tc.id, got.NRGBAAt(1, 1))
		}
	}
}

// A relief tile is darkened only where it differs from the meadow baked into it,
// so its meadow border meets a graded grass tile without a step.
func TestReliefSparesBakedMeadow(t *testing.T) {
	meadow := flat(color.NRGBA{100, 107, 40, 255}, 8, 2)
	hills := flat(color.NRGBA{100, 107, 40, 255}, 8, 2)
	for x := 4; x < 8; x++ {
		for y := range 2 {
			hills.SetNRGBA(x, y, color.NRGBA{112, 104, 60, 255})
		}
	}
	grass := Grade(meadow, "tile/grass", "terrain/grass", nil)
	got := Grade(hills, "context/hills/abc", "terrain/hills", meadow)
	if got.NRGBAAt(1, 0) != grass.NRGBAAt(1, 0) {
		t.Errorf("baked meadow %v, grass %v", got.NRGBAAt(1, 0), grass.NRGBAAt(1, 0))
	}
	plain := Grade(hills, "tile/x", "terrain/grass", nil)
	lr, _, _ := lab(got.NRGBAAt(5, 0))
	lp, _, _ := lab(plain.NRGBAAt(5, 0))
	if d := lr - lp; math.Abs(d-rows["terrain/hills"].l) > 0.6 {
		t.Errorf("relief pixel %.2f against land %.2f: offset %.2f, want %v", lr, lp, d, rows["terrain/hills"].l)
	}
}

// Crowns follow their own curve: the mean crown lands in the forest's 26–31,
// gaps stay at most 18 and lit tops at most 58.
func TestCrownCurve(t *testing.T) {
	for _, tc := range []struct{ in, lo, hi float64 }{{8, 0, 18}, {25.7, 26, 31}, {46, 50, 58}} {
		if got := crownL.at(tc.in); got < tc.lo || got > tc.hi {
			t.Errorf("crown L* %v → %.2f, want %v–%v", tc.in, got, tc.lo, tc.hi)
		}
	}
	c := color.NRGBA{50, 62, 26, 255}
	l0, _, _ := lab(c)
	crown := Grade(flat(c, 2, 2), "standing/scatter/crown/oak-22", "terrain/forest", nil).NRGBAAt(0, 0)
	if lc, _, _ := lab(crown); math.Abs(lc-crownL.at(l0)) > 0.6 {
		t.Errorf("crown %.2f → %.2f, want %.2f", l0, lc, crownL.at(l0))
	}
}

// Buildings take their own b* offset instead of the land's where they differ
// from the meadow, and the meadow baked around them is graded as land.
func TestBuildingsOwnB(t *testing.T) {
	meadow := flat(color.NRGBA{100, 107, 40, 255}, 4, 1)
	house := flat(color.NRGBA{100, 107, 40, 255}, 4, 1)
	roof := color.NRGBA{150, 120, 70, 255}
	house.SetNRGBA(3, 0, roof)
	got := Grade(house, "tile/house", "terrain/house", meadow)
	_, _, b0 := lab(roof)
	if _, _, b1 := lab(got.NRGBAAt(3, 0)); math.Abs(b1-b0-rows["terrain/house"].b) > 0.8 {
		t.Errorf("roof b* %.2f → %.2f, want %.2f", b0, b1, b0+rows["terrain/house"].b)
	}
	if grass := Grade(meadow, "tile/grass", "terrain/grass", nil); got.NRGBAAt(0, 0) != grass.NRGBAAt(0, 0) {
		t.Errorf("baked meadow %v, grass %v", got.NRGBAAt(0, 0), grass.NRGBAAt(0, 0))
	}
}

// Grading is a function of the pixel alone: the same colour gives the same
// result in every tile and at every density.
func TestDeterministic(t *testing.T) {
	a := Grade(flat(color.NRGBA{90, 99, 52, 255}, 2, 2), "tile/grass", "terrain/grass", nil)
	b := Grade(flat(color.NRGBA{90, 99, 52, 255}, 9, 5), "context/ground/x", "terrain/ground", nil)
	if a.NRGBAAt(0, 0) != b.NRGBAAt(4, 3) {
		t.Errorf("%v against %v", a.NRGBAAt(0, 0), b.NRGBAAt(4, 3))
	}
}

func TestLabRoundTrip(t *testing.T) {
	for _, c := range []color.NRGBA{{0, 0, 0, 255}, {255, 255, 255, 255}, {100, 107, 40, 255}, {22, 82, 100, 255}, {200, 30, 90, 255}} {
		l, a, b := toLab(c.R, c.G, c.B)
		if got := fromLab(l, a, b, c.A); got != c {
			t.Errorf("%v → %v", c, got)
		}
	}
}

// The shipped summer building tiles come to the direction's b* 16 where they
// differ from the meadow.
func TestShippedBuildings(t *testing.T) {
	load := func(name string) image.Image {
		f, err := os.Open(filepath.Join("../../assets/terrain", name))
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = f.Close() }()
		im, err := png.Decode(f)
		if err != nil {
			t.Fatal(err)
		}
		return im
	}
	meadow := load("grass/grass-2x.png")
	m := Grade(meadow, "x", "none", nil)
	for _, family := range []string{"house", "hq", "cmp", "factory"} {
		files, err := filepath.Glob(filepath.Join("../../assets/terrain", family, "*-2x.png"))
		if err != nil || len(files) == 0 {
			t.Fatalf("%s: no tiles (%v)", family, err)
		}
		var n, painted, graded float64
		for _, f := range files {
			if strings.HasSuffix(f, "_w-2x.png") {
				continue
			}
			src := Grade(load(filepath.Join(family, filepath.Base(f))), "x", "none", nil)
			got := Grade(src, "tile/"+family, "terrain/"+family, meadow)
			for i := 0; i < len(src.Pix); i += 4 {
				if src.Pix[i+3] != 255 || diff(src, i, m.Pix[i:i+4]) < 8 {
					continue
				}
				_, _, b0 := toLab(src.Pix[i], src.Pix[i+1], src.Pix[i+2])
				_, _, b1 := toLab(got.Pix[i], got.Pix[i+1], got.Pix[i+2])
				n, painted, graded = n+1, painted+b0, graded+b1
			}
		}
		if painted, graded = painted/n, graded/n; math.Abs(graded-16) > 0.5 {
			t.Errorf("%s: b* %.2f painted, %.2f graded (%d pixels), want about 16", family, painted, graded, int(n))
		} else {
			t.Logf("%s: b* %.2f painted, %.2f graded", family, painted, graded)
		}
	}
}
