package main

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"testing"
)

func shoreFlatBlendFixture(t *testing.T) shoreInputs {
	t.Helper()
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
	return in
}

// Flat materials isolate the transition from the texture. A subpixel dry-rock
// ramp creates a strong artificial outline even when the coast itself curves.
func TestShoreWaterRockBlendHasNoSubpixelColourCliff(t *testing.T) {
	in := shoreFlatBlendFixture(t)
	p := connectionPoint{48, 40}
	previous := shoreMaterialColor(in, p, -6)
	largest := 0.
	for d := -5.75; d <= 4; d += .25 {
		c := shoreMaterialColor(in, p, d)
		for _, pair := range [][2]uint8{{c.R, previous.R}, {c.G, previous.G}, {c.B, previous.B}} {
			largest = max(largest, math.Abs(float64(pair[0])-float64(pair[1])))
		}
		previous = c
	}
	if largest > 24 {
		t.Fatalf("water/rock edge still has a colour cliff: %.0f levels per quarter logical pixel", largest)
	}
	if shallow := shoreMaterialColor(in, p, -3); shallow == (color.NRGBA{20, 65, 125, 255}) {
		t.Fatal("missing submerged transition before the wet stones")
	}
	if core := shoreMaterialColor(in, p, 3); core.R < 195 {
		t.Fatalf("softening washed out the dry stone core: %v", core)
	}
}

func TestShoreMeadowBlendHasRoomForMossAndGrass(t *testing.T) {
	in := shoreFlatBlendFixture(t)
	p := connectionPoint{48, 40}
	start, finish := -1., -1.
	for d := 3.; d <= 18; d += .05 {
		c := shoreMaterialColor(in, p, d)
		coverage := (205 - float64(c.R)) / 170
		if start < 0 && coverage >= .1 {
			start = d
		}
		if finish < 0 && coverage >= .9 {
			finish = d
		}
	}
	if start < 0 || finish < 0 || finish-start < 3.4 {
		t.Fatalf("rock/meadow fringe too abrupt: 10..90percent span %.2f logical pixels", finish-start)
	}
}
