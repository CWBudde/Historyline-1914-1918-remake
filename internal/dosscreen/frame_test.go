package dosscreen

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// The frame paints a panel once per window size, zoom and pattern of on-map cells:
// scrolling inside the map reuses it, reaching the map edge paints a new one.
func TestFrameCache(t *testing.T) {
	g := hexgrid.Grid{Width: 40, Lines: 40}
	s := Layout(640, 400, 1)
	dst := ebiten.NewImage(640, 400)
	var f Frame
	defer f.Release()
	f.Draw(dst, s, [2]View{s.View(0, 5, 5), s.View(1, 9, 9)}, g)
	if f.painted != 3 {
		t.Fatalf("first draw painted %d images, want chrome and one per panel pattern", f.painted)
	}
	f.Draw(dst, s, [2]View{s.View(0, 6, 7), s.View(1, 8, 10)}, g)
	if f.painted != 3 {
		t.Errorf("scrolling inside the map repainted: %d", f.painted)
	}
	f.Draw(dst, s, [2]View{s.View(0, -1, 0), s.View(1, 8, 10)}, g)
	if f.painted != 4 {
		t.Errorf("the map edge did not paint a new panel: %d", f.painted)
	}
	zoomed := NewView(s.Panels[0], s.CellScale*1.5, 5, 5)
	f.Draw(dst, s, [2]View{zoomed, s.View(1, 8, 10)}, g)
	if f.painted != 5 {
		t.Errorf("a new zoom did not paint a new panel: %d", f.painted)
	}
	s2 := Layout(800, 500, 1)
	f.Draw(ebiten.NewImage(800, 500), s2, [2]View{s2.View(0, 5, 5), s2.View(1, 5, 5)}, g)
	if f.painted != 8 {
		t.Errorf("a new window size painted %d, want chrome and both panels", f.painted)
	}
	if len(f.panels) > maxCached {
		t.Errorf("%d panels cached, limit %d", len(f.panels), maxCached)
	}
}
