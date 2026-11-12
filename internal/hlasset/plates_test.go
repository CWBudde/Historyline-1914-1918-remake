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

func writePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPlatesShareOneSizeAndSymbolsAreSquare(t *testing.T) {
	art := t.TempDir()
	dir := filepath.Join(art, "screens", "plates")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Two plate states of slightly different painted sizes in a hazy margin.
	for name, w := range map[string]int{"plate": 400, "plate-lit": 410} {
		img := image.NewNRGBA(image.Rect(0, 0, 500, 400))
		for y := range 400 {
			for x := range 500 {
				c := color.NRGBA{90, 90, 80, 10}
				if x >= 20 && x < 20+w && y >= 30 && y < 30+335 {
					c.A = 255
				}
				img.SetNRGBA(x, y, c)
			}
		}
		writePNG(t, filepath.Join(dir, name+".png"), img)
	}
	// A 4x2 symbol sheet of 300 px cells: a wide bar in every cell but the last.
	sheet := image.NewNRGBA(image.Rect(0, 0, 1200, 600))
	for i := range 7 {
		ox, oy := i%4*300, i/4*300
		for y := oy + 120; y < oy+180; y++ {
			for x := ox + 30; x < ox+270; x++ {
				sheet.SetNRGBA(x, y, color.NRGBA{214, 206, 186, 255})
			}
		}
	}
	writePNG(t, filepath.Join(dir, "icons.png"), sheet)
	for _, d := range screenDensities {
		files, err := planPlates(art, d)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 2+len(plateIcons) {
			t.Fatalf("%dx: %d files", d, len(files))
		}
		for _, f := range files {
			if f.Family != "plates" {
				t.Fatalf("%s: family %q", f.ID, f.Family)
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
				t.Fatalf("%s %dx: image %v, want %dx%d", f.ID, d, img.Bounds(), f.Width, f.Height)
			}
			switch f.ID {
			case "plates/plate", "plates/plate-lit":
				if f.Width != plateW*d || f.Height != plateH*d {
					t.Fatalf("%s %dx: %dx%d", f.ID, d, f.Width, f.Height)
				}
				// Trimmed: the plate fills its box, no haze left at the corners.
				if _, _, _, a := img.At(0, 0).RGBA(); a>>8 != 255 {
					t.Fatalf("%s: corner alpha %d", f.ID, a>>8)
				}
			case "plates/icon-move":
				if f.Width != plateIcon*d || f.Height != plateIcon*d {
					t.Fatalf("%s %dx: %dx%d", f.ID, d, f.Width, f.Height)
				}
				// The 240x60 bar fills the width, centred: clear above and below it.
				if _, _, _, a := img.At(f.Width/2, 2).RGBA(); a != 0 {
					t.Fatalf("%s: top alpha %d, want clear", f.ID, a>>8)
				}
				if _, _, _, a := img.At(f.Width/2, f.Height/2).RGBA(); a>>8 != 255 {
					t.Fatalf("%s: centre alpha %d", f.ID, a>>8)
				}
			}
		}
	}
}

func TestNoPlatesWithoutTheirArt(t *testing.T) {
	files, err := planPlates(t.TempDir(), 1)
	if err != nil || len(files) != 0 {
		t.Fatalf("%v %v", files, err)
	}
}
