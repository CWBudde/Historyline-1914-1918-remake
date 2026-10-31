package maprender

import (
	"image"
	"image/color"
	"math"
	"os"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// RenderBackdrop paints the level on the given backdrop: an empty hex
// keeps the backdrop's colour, and the units are still drawn.
func TestRenderBackdrop(t *testing.T) {
	ok := true
	r, err := New(os.DirFS("../../assets"))
	if err != nil {
		t.Fatal(err)
	}
	lvl, err := level.Load("../../art/reference/pulse/level.json")
	if err != nil {
		t.Fatal(err)
	}
	w, h := Size(lvl.Grid())
	magenta := color.NRGBA{255, 0, 255, 255}
	bd := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			bd.SetNRGBA(x, y, magenta)
		}
	}
	m, err := r.RenderBackdrop(lvl, 2, bd, nil)
	if err != nil {
		t.Fatal(err)
	}
	if m.Units == 0 {
		t.Error("no units drawn on the backdrop")
	}
	// An empty grass hex shows the backdrop at its centre.
	found := false
outer:
	for line, hs := range lvl.Lines {
		for col, hx := range hs {
			if hx.Terrain != "grass" || hx.Unit != "" || len(hx.Infra.IDs) != 0 {
				continue
			}
			cx, cy := hexgrid.Center(hexgrid.Coord{Line: line, Col: col})
			px := m.Image.NRGBAAt(int(math.Round(cx*2)), int(math.Round(cy*2)))
			if px != magenta {
				t.Errorf("empty grass hex [%d,%d] centre = %v, want the backdrop colour", line+1, col+1, px)
				ok = false
			}
			found = true
			break outer
		}
	}
	if !found {
		t.Fatal("no empty grass hex in the PULSE fixture")
	}
	if _, err := r.RenderBackdrop(lvl, 2, image.NewNRGBA(image.Rect(0, 0, 100, 100)), nil); err == nil {
		t.Error("a backdrop with the wrong aspect was accepted")
	}
	_ = ok
}
