package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestUnitOptionalFramingValidation(t *testing.T) {
	_, raw := unitTestFixture(t)
	for _, tc := range []struct {
		field string
		bad   bool
	}{
		{`"framing_inset":0.95,`, false},
		{`"framing_inset":1,`, false},
		{`"framing_inset":null,`, true},
		{`"framing_inset":0.79,`, true},
		{`"framing_inset":1.01,`, true},
		{`"framing_inset":"0.95",`, true},
		{`"Framing_Inset":0.95,`, true},
	} {
		var m unitManifest
		data := bytes.Replace(raw, []byte("{"), []byte("{"+tc.field), 1)
		if err := decodeUnitManifest(data, &m); (err != nil) != tc.bad {
			t.Fatalf("%s: %v", tc.field, err)
		}
	}
}

func TestUnitReferenceFramingKeepsAnchorAndGuard(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 384, 384))
	draw.Draw(src, image.Rect(60, 60, 324, 324), image.NewUniform(color.NRGBA{180, 140, 110, 255}), image.Point{}, draw.Src)
	// A small contrasting mark at the logical pivot checks registration while
	// the body grows. All six atlas panels use this same whole-panel transform.
	draw.Draw(src, image.Rect(188, 192, 196, 200), image.NewUniform(color.NRGBA{240, 220, 170, 255}), image.Point{}, draw.Src)
	for _, inset := range []float64{.8, .92, .95, 1} {
		im, err := unitAffineInset(src, inset)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = unitGuard(im, 4); err != nil {
			t.Fatal(err)
		}
		if p := im.NRGBAAt(192, 196); p.R < 230 || p.A != 255 {
			t.Fatalf("inset%v moved logical pivot: %v", inset, p)
		}
	}
	// Authoring outside the fixed safe margin is still rejected, never cropped
	// or repaired to make a larger unit appear to pass the existing guard.
	draw.Draw(src, src.Bounds(), image.NewUniform(color.NRGBA{160, 120, 80, 255}), image.Point{}, draw.Src)
	im, err := unitAffineInset(src, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = unitGuard(im, 4); err == nil {
		t.Fatal("full-canvas artwork bypassed six-world-pixel guard")
	}
}

func TestUnitReferenceGuardRetainsAndReportsFaintFringe(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 384, 384))
	im.SetNRGBA(192, 196, color.NRGBA{180, 150, 120, 255})
	im.SetNRGBA(8, 8, color.NRGBA{80, 40, 60, 7})
	before := append([]byte(nil), im.Pix...)
	if _, err := unitGuard(im, 4); err == nil {
		t.Fatal("default strict guard accepted faint border")
	}
	r, err := unitGuardAlpha(im, 4, 8)
	if err != nil {
		t.Fatal(err)
	}
	if r.GuardWorld != 0 || r.ContrastGuardWorld != 6 || r.GuardAlphaMin != 8 || r.FringeSamples != 1 || r.FringeMaximum != 7 || r.Bounds[0] != 8 {
		t.Fatalf("dishonest fringe report: %+v", r)
	}
	if !bytes.Equal(before, im.Pix) {
		t.Fatal("guard altered source pixels")
	}
	im.SetNRGBA(8, 8, color.NRGBA{80, 40, 60, 8})
	if _, err = unitGuardAlpha(im, 4, 8); err == nil {
		t.Fatal("reference framing accepted visible border")
	}
	im.SetNRGBA(8, 8, color.NRGBA{80, 40, 60, 0})
	if _, err = unitGuardAlpha(im, 4, 8); err == nil {
		t.Fatal("reference framing accepted hidden zero-alpha RGB")
	}
}
