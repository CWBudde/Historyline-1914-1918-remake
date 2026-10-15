package camera

import (
	"math"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestRoundTripAndBounds(t *testing.T) {
	c := New(Size{1000, 800}, Size{400, 300})
	c.Scroll(-50, -50)
	if c.Origin != (Point{}) {
		t.Fatalf("scrolled above the map: %v", c.Origin)
	}
	c.Scroll(5000, 5000)
	if c.Origin != (Point{600, 500}) {
		t.Fatalf("scrolled past the map: %v", c.Origin)
	}
	wx, wy := c.ScreenToWorld(123, 45)
	if sx, sy := c.WorldToScreen(wx, wy); !near(sx, 123) || !near(sy, 45) {
		t.Fatalf("round trip = %v,%v", sx, sy)
	}
}

func TestSmallMapIsCentred(t *testing.T) {
	c := New(Size{100, 100}, Size{400, 300})
	if sx, sy := c.WorldToScreen(0, 0); !near(sx, 150) || !near(sy, 100) {
		t.Fatalf("map corner at %v,%v", sx, sy)
	}
}

// TestPickedHexSurvivesViewChanges is the G2 acceptance: the hex under a fixed
// screen point stays the same through zoom and resize, and picking through the
// camera agrees with picking the world point directly.
func TestPickedHexSurvivesViewChanges(t *testing.T) {
	g := hexgrid.Grid{Width: 8, Lines: 23}
	world := Size{8*hexgrid.ColPitch + hexgrid.HalfPitch, 23*hexgrid.LinePitch + 40}
	c := New(world, Size{600, 400})
	c.Scroll(120, 90)
	pick := func(sx, sy float64) (hexgrid.Coord, bool) {
		wx, wy := c.ScreenToWorld(sx, sy)
		return g.PixelToHex(wx, wy)
	}
	sx, sy := 300.0, 200.0
	want, ok := pick(sx, sy)
	if !ok {
		t.Fatal("test point is off the map")
	}
	for _, f := range []float64{1.5, 0.8, 1.2, 0.9} {
		c.ZoomAt(f, sx, sy)
		if got, _ := pick(sx, sy); got != want {
			t.Fatalf("after zoom %.2f: %v, want %v", c.Zoom, got, want)
		}
	}
	// The hex centre must map to the screen and back to the same hex.
	cx, cy := hexgrid.Center(want)
	scx, scy := c.WorldToScreen(cx, cy)
	if got, _ := pick(scx, scy); got != want {
		t.Fatalf("centre of %v picks %v", want, got)
	}
	c.Resize(Size{900, 700})
	cwx, cwy := hexgrid.Center(want)
	scx, scy = c.WorldToScreen(cwx, cwy)
	if got, _ := pick(scx, scy); got != want {
		t.Fatalf("after resize centre of %v picks %v", want, got)
	}
}

func TestZoomLimits(t *testing.T) {
	c := New(Size{1000, 1000}, Size{200, 200})
	c.ZoomAt(100, 0, 0)
	if c.Zoom != MaxZoom {
		t.Fatalf("zoom = %v", c.Zoom)
	}
	c.ZoomAt(1e-6, 0, 0)
	if c.Zoom != MinZoom {
		t.Fatalf("zoom = %v", c.Zoom)
	}
}
