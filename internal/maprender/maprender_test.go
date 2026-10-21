package maprender

import (
	"image"
	"image/color"
	"os"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// Catch odd-line shifts and anchors drifting between authored densities.
func TestPlace(t *testing.T) {
	for _, density := range []int{1, 2, 4} {
		src := image.NewNRGBA(image.Rect(0, 0, 96*density, 80*density))
		src.SetNRGBA(48*density, 40*density, color.NRGBA{R: 255, A: 255})
		dst := image.NewNRGBA(image.Rect(0, 0, 300*density, 200*density))
		Place(dst, src, hexgrid.Coord{Line: 1, Col: 0}, density, [2]float64{48, 40})
		if got := dst.NRGBAAt(119*density, 80*density); got.R != 255 {
			t.Errorf("%dx odd-line anchor: %+v", density, got)
		}
		clear(dst.Pix)
		Place(dst, src, hexgrid.Coord{Line: 1, Col: 0}, density, [2]float64{48, 49})
		if got := dst.NRGBAAt(119*density, 71*density); got.R != 255 {
			t.Errorf("%dx unit anchor: %+v", density, got)
		}
	}
}

func TestRenderPulse(t *testing.T) {
	r, err := New(os.DirFS("../../assets"))
	if err != nil {
		t.Fatal(err)
	}
	details := 0
	for id := range r.has {
		if strings.HasPrefix(id, "tile/grass-detail-") {
			details++
		}
	}
	if details != 16 {
		t.Errorf("%d grass details at both densities, want sixteen", details)
	}
	lvl, err := level.Load("../../levels/level1_cp.json")
	if err != nil {
		t.Fatal(err)
	}
	m, err := r.Render(lvl, 2, true)
	if err != nil {
		t.Fatal(err)
	}
	w, h := Size(lvl.Grid())
	if b := m.Image.Bounds(); b.Dx() != 2*w || b.Dy() != 2*h {
		t.Errorf("size %v, want %dx%d", b, 2*w, 2*h)
	}
	if m.Units == 0 {
		t.Error("no units drawn")
	}
}
