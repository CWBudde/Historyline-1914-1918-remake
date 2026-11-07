package dosscreen

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// The frame paints its chrome once per window size: zooming reuses it, a new
// window size paints it again.
func TestFrameCache(t *testing.T) {
	s := Layout(640, 400, 1)
	dst := ebiten.NewImage(640, 400)
	var f Frame
	defer f.Release()
	f.Draw(dst, s)
	if f.painted != 1 {
		t.Fatalf("first draw painted %d images, want the chrome", f.painted)
	}
	f.Draw(dst, s.Zoomed(s.ZoomLevels()[0]))
	if f.painted != 1 {
		t.Errorf("zooming repainted: %d", f.painted)
	}
	s2 := Layout(800, 500, 1)
	f.Draw(ebiten.NewImage(800, 500), s2)
	if f.painted != 2 {
		t.Errorf("a new window size painted %d, want 2", f.painted)
	}
}
