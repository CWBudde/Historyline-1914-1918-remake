//go:build origtiles

package main

import (
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origview"
)

// TestOrigTiles decodes the local original tiles: every terrain record has a
// visible 24×24 tile in both seasons. It writes contact sheets to
// build/review/orig-tiles.
func TestOrigTiles(t *testing.T) {
	dir, err := findDosFull(".")
	if err != nil {
		t.Skip(err)
	}
	sheets, err := readOrigTiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join("..", "..", "build", "review", "orig-tiles")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	for season, sheet := range sheets {
		if len(sheet) != origview.Records {
			t.Fatalf("season %d: %d tiles, want %d", season, len(sheet), origview.Records)
		}
		const cols = 16
		contact := image.NewNRGBA(image.Rect(0, 0, cols*26, (len(sheet)+cols-1)/cols*26))
		for r, im := range sheet {
			if im.Bounds() != image.Rect(0, 0, 24, 24) {
				t.Errorf("season %d record %d: size %v", season, r, im.Bounds())
			}
			if visible(im).Empty() {
				t.Errorf("season %d record %d: no visible pixel", season, r)
			}
			p := image.Pt(r%cols*26, r/cols*26)
			draw.Draw(contact, im.Bounds().Add(p), im, image.Point{}, draw.Over)
		}
		t.Logf("season %d: record 0 visible %v", season, visible(sheet[0]))
		f, err := os.Create(filepath.Join(out, [2]string{"summer.png", "winter.png"}[season]))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, contact); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

// visible is the bounding box of im's opaque pixels.
func visible(im *image.NRGBA) image.Rectangle {
	var r image.Rectangle
	for y := im.Rect.Min.Y; y < im.Rect.Max.Y; y++ {
		for x := im.Rect.Min.X; x < im.Rect.Max.X; x++ {
			if im.NRGBAAt(x, y).A != 0 {
				r = r.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	return r
}
