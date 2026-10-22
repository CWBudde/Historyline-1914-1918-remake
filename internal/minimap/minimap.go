// Package minimap holds the minimap and the strategic map: the geometry of a scaled
// copy of the world at a fixed screen position, a cell-exact raster of the hex map
// (Raster), the press rule (Press) and the strategic-map panel (Strategic). It draws
// into plain images only; uploading them is up to the caller.
package minimap

import "github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"

// Minimap is the world drawn into the screen rectangle starting at Pos. Inset is
// the world margin left out on each side (the slanted hex tips at the left and
// right map edge, see Raster); ScaleY is the vertical scale, 0 meaning Scale.
type Minimap struct {
	World  camera.Size
	Pos    camera.Point
	Scale  float64 // minimap pixels per world pixel
	ScaleY float64
	Inset  camera.Point
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

func (m Minimap) scaleY() float64 {
	if m.ScaleY == 0 {
		return m.Scale
	}
	return m.ScaleY
}

// Size returns the on-screen size of the minimap.
func (m Minimap) Size() camera.Size {
	return camera.Size{W: (m.World.W - 2*m.Inset.X) * m.Scale, H: (m.World.H - 2*m.Inset.Y) * m.scaleY()}
}

// Contains reports whether screen position (sx, sy) is on the minimap.
func (m Minimap) Contains(sx, sy float64) bool {
	s := m.Size()
	return sx >= m.Pos.X && sx < m.Pos.X+s.W && sy >= m.Pos.Y && sy < m.Pos.Y+s.H
}

// Clamp moves screen position (sx, sy) onto the minimap, for a drag that leaves it.
func (m Minimap) Clamp(sx, sy float64) (float64, float64) {
	s := m.Size()
	return min(max(sx, m.Pos.X), m.Pos.X+s.W), min(max(sy, m.Pos.Y), m.Pos.Y+s.H)
}

// ToWorld converts a screen position on the minimap to a world position.
func (m Minimap) ToWorld(sx, sy float64) (wx, wy float64) {
	return (sx-m.Pos.X)/m.Scale + m.Inset.X, (sy-m.Pos.Y)/m.scaleY() + m.Inset.Y
}

// FromWorld converts a world position to a screen position on the minimap.
func (m Minimap) FromWorld(wx, wy float64) (sx, sy float64) {
	return m.Pos.X + (wx-m.Inset.X)*m.Scale, m.Pos.Y + (wy-m.Inset.Y)*m.scaleY()
}

// ViewRect returns the camera's visible area as a screen rectangle on the minimap,
// clipped to the minimap.
func (m Minimap) ViewRect(c *camera.Camera) (x, y, w, h float64) {
	vx, vy, vw, vh := c.Visible()
	x0, y0 := m.FromWorld(max(vx, m.Inset.X), max(vy, m.Inset.Y))
	x1, y1 := m.FromWorld(min(vx+vw, m.World.W-m.Inset.X), min(vy+vh, m.World.H-m.Inset.Y))
	return x0, y0, x1 - x0, y1 - y0
}
