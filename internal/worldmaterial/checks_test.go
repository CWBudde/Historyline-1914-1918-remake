package worldmaterial

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func checkNamed(t *testing.T, cs []Check, name string) Check {
	t.Helper()
	for _, c := range cs {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no check %q in %v", name, cs)
	return Check{}
}

func TestChecksOnFlatMeadow(t *testing.T) {
	flat := solid(image.Rect(0, 0, 568, 400), color.NRGBA{100, 107, 40, 255})
	tone := ToneChecks(flat)
	if !checkNamed(t, tone, "mean G").Pass || checkNamed(t, tone, "sigma L* of the radius-10 box mean").Pass {
		t.Fatal(tone)
	}
	result := ResultChecks(flat, image.Pt(568, 400), 3)
	if c := checkNamed(t, result, "sigma L*"); c.Pass || c.Value != 0 {
		t.Fatal(c)
	}
	if c := checkNamed(t, result, "Sobel edge density"); c.Value != 0 {
		t.Fatal(c)
	}
	// A flat field has no correlation to report: it counts as zero, not as repetition.
	if c := checkNamed(t, result, "autocorrelation at the detail repeat (189,0)"); !c.Pass {
		t.Fatal(c)
	}
	if Passed(result) {
		t.Fatal("flat meadow passed")
	}
}

func TestPearsonFindsTheRepeat(t *testing.T) {
	// Noise repeating every 10 columns correlates fully at (10,0) and hardly at (3,0).
	base := noise(10, 40, 120, 60, 11)
	im := image.NewNRGBA(image.Rect(0, 0, 40, 40))
	for y := range 40 {
		for x := range 40 {
			im.SetNRGBA(x, y, base.NRGBAAt(x%10, y))
		}
	}
	p := newPlane(im)
	if r := pearson(p, image.Pt(10, 0)); math.Abs(r-1) > 1e-9 {
		t.Fatal(r)
	}
	if r := pearson(p, image.Pt(3, 0)); math.Abs(r) > 0.2 {
		t.Fatal(r)
	}
}

func TestDarkWidth(t *testing.T) {
	im := solid(image.Rect(0, 0, 60, 60), color.NRGBA{120, 120, 120, 255})
	// A 15×3 dark bar is wider than 12 world px; a 3×3 speck is not.
	for y := 20; y < 23; y++ {
		for x := 10; x < 25; x++ {
			im.SetNRGBA(x, y, color.NRGBA{40, 40, 40, 255})
		}
	}
	for y := 40; y < 43; y++ {
		for x := 40; x < 43; x++ {
			im.SetNRGBA(x, y, color.NRGBA{40, 40, 40, 255})
		}
	}
	if w := darkWidth(newPlane(im)); w != 15 {
		t.Fatal(w)
	}
}

func TestDetailChecksOnNoise(t *testing.T) {
	// Pixel noise has far more energy at 1/3 than at 3/10 and is even.
	cs := DetailChecks(noise(200, 150, 110, 40, 5))
	if c := checkNamed(t, cs, "band ratio RMS(1/3) / RMS(3/10)"); c.Pass || c.Value < 2 {
		t.Fatal(c)
	}
	if c := checkNamed(t, cs, "CV of the 3/10 band RMS over 50 px windows"); !c.Pass {
		t.Fatal(c)
	}
}

func TestSuggestGain(t *testing.T) {
	tone := solid(image.Rect(0, 0, 120, 80), color.NRGBA{100, 107, 40, 255})
	detail := noise(120, 80, 110, 40, 9)
	g := SuggestGain(tone, detail)
	if g <= 0 {
		t.Fatal(g)
	}
	// Applying it keeps the 3/10 band at or under the cap (within rounding).
	got, err := Combine(tone, detail, 1, g)
	if err != nil {
		t.Fatal(err)
	}
	if b := rms(band(newPlane(got), 3, 10)); b > meadowBandCap+0.2 {
		t.Fatal(b)
	}
	if SuggestGain(tone, tone) != 0 {
		t.Fatal("flat detail")
	}
}
