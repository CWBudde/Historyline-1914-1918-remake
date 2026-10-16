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

// Scale is the device factor under the user zoom: it multiplies the transform,
// keeps the view centre and leaves Zoom and its limits alone.
func TestScaleKeepsCentreAndZoom(t *testing.T) {
	c := New(Size{4000, 3000}, Size{800, 600})
	c.CenterOn(2000, 1500)
	c.ZoomAt(1.5, 400, 300)
	c.SetScale(2)
	if c.Zoom != 1.5 || !near(c.Effective(), 3) {
		t.Fatalf("zoom %v effective %v", c.Zoom, c.Effective())
	}
	if wx, wy := c.ScreenToWorld(400, 300); !near(wx, 2000) || !near(wy, 1500) {
		t.Fatalf("centre moved to %v,%v", wx, wy)
	}
	if sx, sy := c.WorldToScreen(2010, 1500); !near(sx, 430) || !near(sy, 300) {
		t.Fatalf("10 world px at effective 3 = %v,%v", sx, sy)
	}
	if _, _, w, h := c.Visible(); !near(w, 800/3.0) || !near(h, 200) {
		t.Fatalf("visible %vx%v", w, h)
	}
	c.ZoomAt(100, 400, 300)
	if c.Zoom != MaxZoom || !near(c.Effective(), 2*MaxZoom) {
		t.Fatalf("zoom limit applies to the user zoom: %v / %v", c.Zoom, c.Effective())
	}
}
