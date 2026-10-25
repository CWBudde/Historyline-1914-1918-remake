//go:build origtiles

package main

import (
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origview"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
	"github.com/hajimehoshi/ebiten/v2"
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

// TestOrigUnits decodes the local original unit sprites: six visible 24×24
// frames per unit type, for each side, moved or not. It writes a contact sheet
// (a row per type: side 0, side 1, side 0 moved) to build/review/orig-tiles.
func TestOrigUnits(t *testing.T) {
	dir, err := findDosFull(".")
	if err != nil {
		t.Skip(err)
	}
	units, err := readOrigUnits(dir)
	if err != nil {
		t.Fatal(err)
	}
	sets := [][]*image.NRGBA{units[0][0], units[1][0], units[0][1]}
	contact := image.NewNRGBA(image.Rect(0, 0, len(sets)*(6*26+8), origUnitTypes*26))
	for s, frames := range sets {
		if len(frames) != 6*origUnitTypes {
			t.Fatalf("set %d: %d frames, want %d", s, len(frames), 6*origUnitTypes)
		}
		for k, im := range frames {
			if im.Bounds() != image.Rect(0, 0, 24, 24) || visible(im).Empty() {
				t.Errorf("set %d frame %d: size %v, visible %v", s, k, im.Bounds(), visible(im))
			}
			p := image.Pt(s*(6*26+8)+k%6*26, k/6*26)
			draw.Draw(contact, im.Bounds().Add(p), im, image.Point{}, draw.Over)
		}
	}
	out := filepath.Join("..", "..", "build", "review", "orig-tiles")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(out, "units.png"))
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

// TestOrigPanelUnits: with the original art on, the contents' slots, the unit
// info and the battle show the original sprites, enlarged; off, the remake's.
func TestOrigPanelUnits(t *testing.T) {
	if _, err := findDosFull("."); err != nil {
		t.Skip(err)
	}
	g, d, f := newContentsTest(t)
	openFactory(t, g, d, f)
	if msg := g.toggleOrigArt(); !strings.HasSuffix(msg, "on") {
		t.Fatalf("F8 says %q", msg)
	}
	orig := map[int]*ebiten.Image{}
	for i, s := range g.contentsView().Slots {
		if !s.Occupied {
			continue
		}
		u := *d.boxSlots()[i].Unit
		want := g.origPanelUnit(u, unitsprite.N)
		if want == nil || want.Bounds().Dx() != origPanelScale*origHex || s.Sprite != want {
			t.Errorf("slot %d (%s): sprite %p, want the original %p", i, u.Name, s.Sprite, want)
		}
		if info := g.unitInfo(u); info.Sprite != want {
			t.Errorf("unit info of %s: sprite %p, want the original %p", u.Name, info.Sprite, want)
		}
		orig[i] = want
	}
	if len(orig) == 0 {
		t.Fatal("no unit in the factory")
	}
	g.drawContents(ebiten.NewImage(1920, 1080), image.Rect(0, 0, 640, 1080))
	g.toggleOrigArt()
	for i, s := range g.contentsView().Slots {
		if s.Occupied && s.Sprite == orig[i] {
			t.Errorf("slot %d keeps the original sprite with the art off", i)
		}
	}
}
