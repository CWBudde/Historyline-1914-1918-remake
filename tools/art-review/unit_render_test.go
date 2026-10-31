package artreview

import (
	"image"
	"image/color"
	"math"
	"testing"
)

// Replacing tint energy with coverage, keying skin, or premultiplying twice breaks these witnesses.
func TestUnitPlanesWitnessesAndTint(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 5, 1))
	samples := []color.NRGBA{{200, 0, 200, 173}, {190, 140, 90, 255}, {80, 60, 40, 200}, {200, 192, 200, 255}, {200, 0, 200, 0}}
	for i, p := range samples {
		src.SetNRGBA(i, 0, p)
	}
	base, mask := unitPlanes(src)
	wantN := []color.NRGBA{{200, 200, 200, 173}, {190, 140, 90, 255}, {80, 60, 40, 200}, {200, 192, 200, 255}, {}}
	wantK := []color.NRGBA{{200, 200, 200, 173}, {0, 0, 0, 255}, {0, 0, 0, 200}, {41, 41, 41, 255}, {}}
	// q=.04 gives smoothstep(.25)=.103515625, K=20.703125 ->21; N=(200,192.828125,200).
	wantN[3] = color.NRGBA{200, 193, 200, 255}
	wantK[3] = color.NRGBA{21, 21, 21, 255}
	for i := range samples {
		if got := base.NRGBAAt(i, 0); got != wantN[i] {
			t.Errorf("base %d got %v want %v", i, got, wantN[i])
		}
		if got := mask.NRGBAAt(i, 0); got != wantK[i] {
			t.Errorf("mask %d got %v want %v", i, got, wantK[i])
		}
	}
	normal := unitTint(base, mask, []int{104, 142, 83}, false)
	spent := unitTint(base, mask, []int{104, 142, 83}, true)
	if got := normal.NRGBAAt(0, 0); got != (color.NRGBA{82, 111, 65, 173}) {
		t.Fatalf("tint %v", got)
	}
	for x := range 5 {
		n, s := normal.NRGBAAt(x, 0), spent.NRGBAAt(x, 0)
		if s != (color.NRGBA{n.R, n.R, n.R, n.A}) {
			t.Fatal("spent is not post-tint red-channel gray")
		}
	}
	// Negative lobes can produce N<K: the final colour must saturate, not wrap.
	base.SetNRGBA(0, 0, color.NRGBA{10, 20, 30, 255})
	mask.SetNRGBA(0, 0, color.NRGBA{200, 200, 200, 255})
	if p := unitTint(base, mask, []int{0, 0, 0}, false).NRGBAAt(0, 0); p != (color.NRGBA{0, 0, 0, 255}) {
		t.Fatal(p)
	}
}

// The direct 2D oracle includes transparent full support, signed lobes and the (48,49) registration.
func TestUnitAffineIndependentOracle(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 384, 384))
	for y := 75; y < 310; y++ {
		for x := 80; x < 290; x++ {
			src.SetNRGBA(x, y, color.NRGBA{uint8(x % 201), uint8(y % 219), 90, uint8(70 + (x+y)%186)})
		}
	}
	got, err := unitAffine(src)
	if err != nil {
		t.Fatal(err)
	}
	weight := func(u float64) float64 {
		u = math.Abs(u)
		if u == 0 {
			return 1
		}
		if u >= 3 {
			return 0
		}
		return math.Sin(math.Pi*u) / (math.Pi * u) * math.Sin(math.Pi*u/3) / (math.Pi * u / 3)
	}
	// cell384 -> scale .8, offset(38.4,39.2), source anchor(192,196).
	for _, p := range []image.Point{{0, 0}, {50, 101}, {101, 101}, {103, 104}, {192, 196}, {269, 269}, {345, 300}} {
		cx, cy := (float64(p.X)+.5-38.4)/.8-.5, (float64(p.Y)+.5-39.2)/.8-.5
		var sum [4]float64
		total := 0.
		for y := int(math.Ceil(cy - 3/.8)); y <= int(math.Floor(cy+3/.8)); y++ {
			for x := int(math.Ceil(cx - 3/.8)); x <= int(math.Floor(cx+3/.8)); x++ {
				w := weight((float64(x)-cx)*.8) * weight((float64(y)-cy)*.8)
				total += w
				q := src.NRGBAAt(x, y)
				sum[3] += w * float64(q.A)
				sum[0] += w * float64(q.A) * float64(q.R)
				sum[1] += w * float64(q.A) * float64(q.G)
				sum[2] += w * float64(q.A) * float64(q.B)
			}
		}
		want := color.NRGBA{}
		a := math.Floor(math.Max(0, math.Min(255, sum[3]/total)) + .5)
		if a > 0 {
			c := func(v float64) uint8 { return uint8(math.Floor(math.Max(0, math.Min(255, v/sum[3])) + .5)) }
			want = color.NRGBA{c(sum[0]), c(sum[1]), c(sum[2]), uint8(a)}
		}
		actual := got.NRGBAAt(p.X, p.Y)
		for i, v := range []uint8{actual.R, actual.G, actual.B, actual.A} {
			wv := []uint8{want.R, want.G, want.B, want.A}[i]
			if math.Abs(float64(v)-float64(wv)) > 1 {
				t.Fatalf("%v got %v want %v", p, actual, want)
			}
		}
	}
	if _, err = unitAffine(image.NewNRGBA(image.Rect(0, 0, 1, 1))); err == nil {
		t.Fatal("upscale accepted")
	}
}

// A source-border fleck is legitimate; fixed .8 framing supplies the runtime guard.
func TestUnitFixedFrameFullPanelNoAdjacentBleed(t *testing.T) {
	atlas := image.NewNRGBA(image.Rect(0, 0, 768, 1152))
	for face := range 6 {
		ox, oy := face%2*384, face/2*384
		for y := range 384 {
			for x := range 384 {
				atlas.SetNRGBA(ox+x, oy+y, color.NRGBA{uint8(25 + face*20), 180, 50, 255})
			}
		}
	}
	for face := range 6 {
		rect := image.Rect(face%2*384, face/2*384, (face%2+1)*384, (face/2+1)*384)
		src := atlas.SubImage(rect).(*image.NRGBA)
		n, _ := unitPlanes(src)
		out, err := unitAffine(n)
		if err != nil {
			t.Fatal(err)
		}
		want := color.NRGBA{uint8(25 + face*20), 180, 50, 255}
		if p := out.NRGBAAt(192, 196); p != want {
			t.Fatalf("face %d got %v want %v", face, p, want)
		}
		for _, at := range []image.Point{{37, 196}, {38, 196}, {39, 196}, {345, 196}, {346, 196}} {
			p := out.NRGBAAt(at.X, at.Y)
			if p.A > 0 && (p.R != want.R || p.G != want.G || p.B != want.B) {
				t.Fatal("adjacent panel bleed", face, at, p)
			}
		}
		if _, err = unitGuard(out, 4); err != nil {
			t.Fatal(err)
		}
	}
	src := image.NewNRGBA(image.Rect(0, 0, 384, 384))
	for y := 90; y < 310; y++ {
		for x := 140; x < 240; x++ {
			src.SetNRGBA(x, y, color.NRGBA{180, 140, 90, 255})
		}
	}
	src.SetNRGBA(383, 383, color.NRGBA{120, 90, 50, 255})
	got, err := unitAffine(src)
	if err != nil {
		t.Fatal(err)
	}
	if got.NRGBAAt(192, 196) != (color.NRGBA{180, 140, 90, 255}) {
		t.Fatal("registration drift")
	}
	if _, err = unitGuard(got, 4); err != nil {
		t.Fatal("source edge fleck must fit fixed runtime guard", err)
	}
}

func TestUnitGuardRejectsVisibleBorderAndHiddenRGB(t *testing.T) {
	for _, p := range []color.NRGBA{{1, 2, 3, 255}, {1, 2, 3, 0}} {
		im := image.NewNRGBA(image.Rect(0, 0, 192, 192))
		im.SetNRGBA(0, 0, p)
		im.SetNRGBA(96, 96, color.NRGBA{90, 80, 70, 255})
		if _, err := unitGuard(im, 2); err == nil {
			t.Fatal("invalid guard accepted", p)
		}
	}
	if _, err := unitGuard(image.NewNRGBA(image.Rect(0, 0, 192, 192)), 2); err == nil {
		t.Fatal("empty output accepted")
	}
}
