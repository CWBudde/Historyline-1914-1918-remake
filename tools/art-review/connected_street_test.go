package main

import (
	"image"
	"image/color"
	"image/draw"
	"path/filepath"
	"testing"
)

func TestStreetCentreMarkingLeavesPavementVisible(t *testing.T) {
	paint := connectionRoadPaintFor(sharedConnectionGeometry(), connectionVariant{Ports: []int{3, 6}})
	checked := 0
	for y := 20.; y < 60; y += .5 {
		if connectionRoadPaintAt(paint, connectionPoint{48, y}, .25) < .95 {
			continue
		}
		if got := connectionRoadPaintAt(paint, connectionPoint{49.3, y}, .25); got > .01 {
			t.Fatalf("centre marking obscures adjacent pavement at y%g: %g", y, got)
		}
		checked++
	}
	if checked < 10 {
		t.Fatal("insufficient dash samples")
	}
}

func TestStreetEdgeLinesSurviveBothExportDensities(t *testing.T) {
	dir, m := connectedFixture(t)
	f := m["families"].([]any)[0].(map[string]any)
	f["markings"], f["width"], f["border"] = "dashed", 22, 1
	f["variants"] = []any{map[string]any{"id": "road36", "ports": []int{3, 6}}}
	surface := image.NewNRGBA(image.Rect(0, 0, 384, 320))
	draw.Draw(surface, surface.Bounds(), image.NewUniform(color.NRGBA{55, 58, 57, 255}), image.Point{}, draw.Src)
	m["materials"].(map[string]any)["road"] = unitTestPNG(t, filepath.Join(dir, "paved.png"), surface)
	in, err := loadConnectionInputs(writeConnectedManifest(t, dir, m))
	if err != nil {
		t.Fatal(err)
	}
	tiles, err := exportConnectionTiles(in, &in.manifest.Families[0], &in.manifest.Families[0].Variants[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, density := range []int{2, 4} {
		for _, x := range []int{38, 58} {
			c := tiles[density].NRGBAAt(x*density, 35*density)
			if c.R < 120 || c.G < 120 || c.B < 120 {
				t.Errorf("%dx street edge line missing at x%d: %v", density, x, c)
			}
		}
	}
}
