package main

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"testing"
)

// Use flat, distinct materials so texture contrast cannot hide a constant-width
// ribbon. At the same distance inland, different interior stretches must have
// visibly different amounts of bank before returning to meadow.
func TestShoreInteriorBankHasBroadAndNarrowStretches(t *testing.T) {
	path, manifest := shoreTestFixture(t)
	manifest["bank_profile"] = authoredBankProfile
	shoreTestWrite(t, path, manifest)
	in, err := loadShoreInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	for name, c := range map[string]color.NRGBA{
		"grass": {35, 110, 35, 255},
		"water": {20, 65, 125, 255},
		"bank":  {205, 195, 175, 255},
	} {
		im := in.Materials[name].Image
		draw.Draw(im, im.Bounds(), image.NewUniform(c), image.Point{}, draw.Src)
	}
	minimum, maximum := math.Inf(1), 0.
	for _, p := range []connectionPoint{{48, 20}, {48, 28}, {48, 36}, {48, 44}, {48, 52}, {48, 60}} {
		reach := 0.
		for d := 2.; d <= 18; d += .1 {
			if shoreMaterialColor(in, p, d).R >= 130 {
				reach = d
			}
		}
		minimum, maximum = min(minimum, reach), max(maximum, reach)
	}
	if maximum-minimum < 4 {
		t.Fatalf("bank is still an even ribbon: narrowest %.1f, widest %.1f logical pixels", minimum, maximum)
	}
	if minimum < 2 || maximum > 15 {
		t.Fatalf("bank loses its readable scale: %.1f..%.1f", minimum, maximum)
	}
}
