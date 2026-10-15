// Package minimap holds the geometry of the minimap: a scaled copy of the world
// shown at a fixed screen position. It uses the same world size as package camera,
// so clicks and the camera rectangle agree with the main view. It has no rendering
// dependency.
package minimap

import "github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"

// Minimap is the world drawn at Scale into the screen rectangle starting at Pos.
type Minimap struct {
	World camera.Size
	Pos   camera.Point
	Scale float64 // minimap pixels per world pixel
}

// Fit returns a minimap of world that fits in max (keeping the aspect ratio),
// anchored to the bottom-right corner of a view of the given size with margin.
func Fit(world, max, view camera.Size, margin float64) Minimap {
	scale := max.W / world.W
	if s := max.H / world.H; s < scale {
		scale = s
	}
	m := Minimap{World: world, Scale: scale}
	m.Pos = camera.Point{X: view.W - margin - world.W*scale, Y: view.H - margin - world.H*scale}
	return m
}

// Size returns the on-screen size of the minimap.
func (m Minimap) Size() camera.Size {
	return camera.Size{W: m.World.W * m.Scale, H: m.World.H * m.Scale}
}

// Contains reports whether screen position (sx, sy) is on the minimap.
func (m Minimap) Contains(sx, sy float64) bool {
	s := m.Size()
	return sx >= m.Pos.X && sx < m.Pos.X+s.W && sy >= m.Pos.Y && sy < m.Pos.Y+s.H
}

// ToWorld converts a screen position on the minimap to a world position.
func (m Minimap) ToWorld(sx, sy float64) (wx, wy float64) {
	return (sx - m.Pos.X) / m.Scale, (sy - m.Pos.Y) / m.Scale
}

// FromWorld converts a world position to a screen position on the minimap.
func (m Minimap) FromWorld(wx, wy float64) (sx, sy float64) {
	return m.Pos.X + wx*m.Scale, m.Pos.Y + wy*m.Scale
}

// ViewRect returns the camera's visible area as a screen rectangle on the minimap,
// clipped to the minimap.
func (m Minimap) ViewRect(c *camera.Camera) (x, y, w, h float64) {
	vx, vy, vw, vh := c.Visible()
	x0, y0 := m.FromWorld(max(vx, 0), max(vy, 0))
	x1, y1 := m.FromWorld(min(vx+vw, m.World.W), min(vy+vh, m.World.H))
	return x0, y0, x1 - x0, y1 - y0
}
