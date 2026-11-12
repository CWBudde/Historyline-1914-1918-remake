package hlasset

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestArmourPartsAreTrimmedAndHalved(t *testing.T) {
	art := t.TempDir()
	dir := filepath.Join(art, "screens", "armour")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// A 9x7 opaque block at (10,4) in haze of alpha 12 on a 40x20 canvas.
	src := image.NewNRGBA(image.Rect(0, 0, 40, 20))
	for y := range 20 {
		for x := range 40 {
			c := color.NRGBA{200, 100, 50, 12}
			if x >= 10 && x < 19 && y >= 4 && y < 11 {
				c.A = 255
			}
			src.SetNRGBA(x, y, c)
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, src); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "part.png"), b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, d := range screenDensities {
		files, err := planArmour(art, d)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 1 || files[0].ID != "armour/part" || files[0].Family != "armour" {
			t.Fatalf("%dx: %+v", d, files)
		}
		// Trimmed to 9x7, then to the even 8x6: 4x3 logical.
		f := files[0]
		if f.Width != 4*d || f.Height != 3*d {
			t.Fatalf("%dx: %dx%d", d, f.Width, f.Height)
		}
		data, err := f.read()
		if err != nil {
			t.Fatal(err)
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		if img.Bounds().Dx() != f.Width || img.Bounds().Dy() != f.Height {
			t.Fatalf("%dx: image %v", d, img.Bounds())
		}
		for y := range f.Height {
			for x := range f.Width {
				if _, _, _, a := img.At(x, y).RGBA(); a>>8 != 255 {
					t.Fatalf("%dx: (%d,%d) alpha %d, want the haze gone", d, x, y, a>>8)
				}
			}
		}
	}
}
