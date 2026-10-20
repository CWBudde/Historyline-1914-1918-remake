package main

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

// A uniformly smooth bank silhouette ignores clods in its source material.
// Texture may roughen the interior edge, but must not move reciprocal ports.
func TestConnectedTrenchBankTextureRoughensInteriorAndPreservesPortGuard(t *testing.T) {
	makeInputs := func(bank color.NRGBA) connectionInputs {
		materials := map[string]*image.NRGBA{}
		for id, c := range map[string]color.NRGBA{"earth": bank, "trench": {50, 40, 30, 255}, "timber": {100, 80, 60, 255}} {
			im := image.NewNRGBA(image.Rect(0, 0, 2, 2))
			draw.Draw(im, im.Bounds(), image.NewUniform(c), image.Point{}, draw.Src)
			materials[id] = im
		}
		return connectionInputs{manifest: connectionManifest{Geometry: sharedConnectionGeometry()}, materials: materials}
	}
	f := connectionFamily{ID: "trench", Width: 6, Border: 3}
	r := connectionTrenchGeometry{course: connectionRoute{segments: []connectionSegment{{connectionPoint{48, -10}, connectionPoint{48, 90}}}}}
	dark, light := makeInputs(color.NRGBA{0, 0, 0, 255}), makeInputs(color.NRGBA{255, 255, 255, 255})
	a := connectionTrenchFeature(dark, f, r, connectionPoint{53.9, 40}, .125)
	b := connectionTrenchFeature(light, f, r, connectionPoint{53.9, 40}, .125)
	if a.A >= b.A {
		t.Fatalf("bank clods did not change silhouette: dark=%v light=%v", a, b)
	}
	a = connectionTrenchFeature(dark, f, r, connectionPoint{53.9, 4}, .125)
	b = connectionTrenchFeature(light, f, r, connectionPoint{53.9, 4}, .125)
	if a.A != b.A {
		t.Fatalf("bank texture changed fixed port width: %v %v", a, b)
	}
}

func TestConnectedTrenchTimberHasBrokenInteriorPlanksAndFixedGuard(t *testing.T) {
	in := connectedSolidInputs(t)
	for id, c := range trenchTestColors {
		im := image.NewNRGBA(image.Rect(0, 0, 2, 2))
		draw.Draw(im, im.Bounds(), image.NewUniform(c), image.Point{}, draw.Src)
		in.materials[id] = im
	}
	f := connectionFamily{ID: "trench", Width: 6, Border: 3}
	r := connectionTrenchGeometry{course: connectionRoute{segments: []connectionSegment{{connectionPoint{48, -10}, connectionPoint{48, 90}}}}}
	planks, gaps := 0, 0
	for y := 20.; y < 60; y += .125 {
		c := connectionTrenchFeature(in, f, r, connectionPoint{50.9, y}, .25)
		if c.R > 160 {
			planks++
		}
		if c.R < 100 {
			gaps++
		}
	}
	if planks < 20 || gaps < 20 {
		t.Fatalf("timber reads as continuous rails: planks%d gaps%d", planks, gaps)
	}
	for _, y := range []float64{0, 2, 4, 6} {
		if c := connectionTrenchFeature(in, f, r, connectionPoint{50.9, y}, .25); c.R < 160 {
			t.Fatalf("timber guard changed at%g: %v", y, c)
		}
	}
}
