package minimap

import (
	"math"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
)

func TestFitAnchorsBottomRight(t *testing.T) {
	world := camera.Size{W: 2000, H: 1000}
	view := camera.Size{W: 800, H: 600}
	m := Fit(world, camera.Size{W: 200, H: 200}, view, 10)
	if m.Scale != 0.1 {
		t.Fatalf("scale = %v, want 0.1", m.Scale)
	}
	s := m.Size()
	if m.Pos.X+s.W != 790 || m.Pos.Y+s.H != 590 {
		t.Errorf("bottom-right = (%v, %v), want (790, 590)", m.Pos.X+s.W, m.Pos.Y+s.H)
	}
	if !m.Contains(m.Pos.X, m.Pos.Y) || m.Contains(m.Pos.X-1, m.Pos.Y) || m.Contains(790, 590) {
		t.Error("Contains boundaries wrong")
	}
}

func TestWorldRoundTrip(t *testing.T) {
	m := Fit(camera.Size{W: 3000, H: 1500}, camera.Size{W: 240, H: 240}, camera.Size{W: 1000, H: 700}, 8)
	for _, p := range [][2]float64{{0, 0}, {1500, 700}, {2999, 1499}} {
		sx, sy := m.FromWorld(p[0], p[1])
		wx, wy := m.ToWorld(sx, sy)
		if math.Abs(wx-p[0]) > 1e-9 || math.Abs(wy-p[1]) > 1e-9 {
			t.Errorf("round trip of %v = (%v, %v)", p, wx, wy)
		}
	}
}

// A click on the minimap centres the camera on that world point.
func TestClickCentresCamera(t *testing.T) {
	world := camera.Size{W: 4000, H: 2000}
	view := camera.Size{W: 800, H: 600}
	cam := camera.New(world, view)
	m := Fit(world, camera.Size{W: 200, H: 100}, view, 10)
	sx, sy := m.FromWorld(2500, 1200)
	wx, wy := m.ToWorld(sx, sy)
	cam.CenterOn(wx, wy)
	vx, vy, vw, vh := cam.Visible()
	if math.Abs(vx+vw/2-2500) > 1e-6 || math.Abs(vy+vh/2-1200) > 1e-6 {
		t.Errorf("view centre = (%v, %v), want (2500, 1200)", vx+vw/2, vy+vh/2)
	}
}

func TestViewRectClipped(t *testing.T) {
	world := camera.Size{W: 4000, H: 2000}
	view := camera.Size{W: 800, H: 600}
	cam := camera.New(world, view)
	m := Fit(world, camera.Size{W: 400, H: 200}, view, 0)
	x, y, w, h := m.ViewRect(cam)
	if x != m.Pos.X || y != m.Pos.Y || w != 80 || h != 60 {
		t.Errorf("rect = %v %v %v %v", x, y, w, h)
	}
	cam.Zoom = camera.MinZoom // sees more than the whole map
	cam.Resize(camera.Size{W: 800, H: 600})
	x, y, w, h = m.ViewRect(cam)
	s := m.Size()
	if x < m.Pos.X || y < m.Pos.Y || x+w > m.Pos.X+s.W+1e-9 || y+h > m.Pos.Y+s.H+1e-9 {
		t.Errorf("rect %v %v %v %v leaves the minimap", x, y, w, h)
	}
}
