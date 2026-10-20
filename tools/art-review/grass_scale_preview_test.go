//go:build grassgfx

package main

import (
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// Explicit visual proof at the1920x1080 detail capture's zoom1.5. This only
// writes ignored build artifacts and never updates a selected source or export.
func TestGenerateGrassScaleProof(t *testing.T) {
	root := filepath.Join("..", "..")
	master, err := loadPNG(filepath.Join(root, "art/source/terrain/refdesign-v1/grass-meadow-01.png"))
	if err != nil {
		t.Fatal(err)
	}
	for _, spec := range []struct {
		id string
		r  image.Rectangle
	}{{"full1536x1024", image.Rect(0, 0, 1536, 1024)}, {"medium1024x640", image.Rect(256, 192, 1280, 832)}, {"half768x480", image.Rect(384, 272, 1152, 752)}} {
		source := image.NewNRGBA(image.Rect(0, 0, spec.r.Dx(), spec.r.Dy()))
		draw.Draw(source, source.Bounds(), master, spec.r.Min, draw.Src)
		patch := image.NewNRGBA(image.Rect(0, 0, (4*142+96)*2, (8*40+80)*2))
		for variant := range 16 {
			tiles, err := renderGrassVariation(source, variant)
			if err != nil {
				t.Fatal(err)
			}
			for line := range 9 {
				for col := range 5 - line%2 {
					if 4*(line%4)+(col+3*line)%4 != variant {
						continue
					}
					x, y := (142*col+71*(line%2))*2, 40*line*2
					draw.Draw(patch, image.Rect(x, y, x+192, y+160), tiles[2], image.Point{}, draw.Over)
				}
			}
		}
		game, err := resizeLanczos(patch, patch.Bounds().Dx()*3/4, patch.Bounds().Dy()*3/4)
		if err != nil {
			t.Fatal(err)
		}
		out, err := os.Create(filepath.Join(root, "build/grass-refdesign-v1/scale-"+spec.id+"-game1.5.png"))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(out, game); err != nil {
			_ = out.Close()
			t.Fatal(err)
		}
		if err := out.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
