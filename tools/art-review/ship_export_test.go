package main

import (
	"fmt"
	"math"
	"path/filepath"
	"testing"
)

// Each long ship covers the centres of both its hexes, the canvas centre ±
// half the step toward its facing, and ends about hull/2 from the centre.
func TestShipSpritesSpanTwoHexes(t *testing.T) {
	steps := map[string][2]float64{"n": {0, -80}, "nw": {-71, -40}, "sw": {-71, 40}, "s": {0, 80}, "se": {71, 40}, "ne": {71, -40}}
	for _, family := range []string{"battleship", "destroyer", "submarine"} {
		in, err := loadShipInputs(filepath.Join("..", "..", "art", "pipeline", "unit-"+family+"-long-v1.json"))
		if err != nil {
			t.Fatal(err)
		}
		sprites, frames, err := shipSprites(in)
		if err != nil {
			t.Fatal(err)
		}
		if len(sprites) != 6*3*2*2 {
			t.Fatalf("%s: %d sprites", family, len(sprites))
		}
		hull := in.Manifest.Hull
		for _, f := range frames {
			t.Logf("%s %s: hull %.0f source px at (%.0f,%.0f), scale %.3f", family, f.Facing, f.Length, f.Centre[0], f.Centre[1], f.Scale4x)
			im := sprites[fmt.Sprintf("%s-long/cp/%s-normal-4x.png", family, f.Facing)]
			if im.Bounds().Dx() != 4*shipCanvas || im.Bounds().Dy() != 4*shipCanvas {
				t.Fatalf("%s %s: %v", family, f.Facing, im.Bounds())
			}
			// The master must resolve the 2x runtime hull without upscaling;
			// the 4x export may enlarge it by up to two times.
			if f.Scale4x/2 > 1 {
				t.Errorf("%s %s: 2x hull upscaled %.2f from its master", family, f.Facing, f.Scale4x/2)
			}
			step := steps[f.Facing]
			l := math.Hypot(step[0], step[1])
			ux, uy := step[0]/l, step[1]/l
			c := float64(shipCanvas) / 2
			// The hull passes within 6 logical px of both hex centres (the
			// painted diagonals run a few degrees steeper than the hex axis).
			for _, s := range []float64{-.5, .5} {
				x, y := 4*(c+s*step[0]), 4*(c+s*step[1])
				near := math.Inf(1)
				for py := range im.Bounds().Dy() {
					for px := range im.Bounds().Dx() {
						if im.NRGBAAt(px, py).A >= 128 {
							near = min(near, math.Hypot(float64(px)-x, float64(py)-y)/4)
						}
					}
				}
				if near > 6 {
					t.Errorf("%s %s: hull %.1f logical px from the hex centre %+.1f step", family, f.Facing, near, s)
				}
			}
			// Nothing opaque more than hull/2 + 8 from the centre along the axis.
			for y := range im.Bounds().Dy() {
				for x := range im.Bounds().Dx() {
					if im.NRGBAAt(x, y).A < 128 {
						continue
					}
					tx := (float64(x)/4-c)*ux + (float64(y)/4-c)*uy
					if math.Abs(tx) > hull/2+8 {
						t.Fatalf("%s %s: hull at %.1f logical px from the centre", family, f.Facing, tx)
					}
				}
			}
		}
	}
}
