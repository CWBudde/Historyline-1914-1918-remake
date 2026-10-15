// Package camera maps between world pixels (the hexgrid layout) and screen pixels.
// It has no rendering dependency, so the same transforms serve drawing, picking and
// the minimap, and are testable without a window.
package camera

import "math"

// Zoom limits keep tiles readable and the transform well conditioned.
const (
	MinZoom = 0.25
	MaxZoom = 4.0
)

// Camera shows part of a World.W × World.H pixel map in a View.W × View.H screen
// area. Origin is the world point at the top-left of the view; a world point (wx, wy)
// appears at screen ((wx-Origin.X)*Zoom, (wy-Origin.Y)*Zoom).
type Camera struct {
	World  Size
	View   Size
	Zoom   float64
	Origin Point
}

// Point and Size are in pixels.
type (
	Point struct{ X, Y float64 }
	Size  struct{ W, H float64 }
)

// New returns a camera at zoom 1 looking at the top-left of the world.
func New(world, view Size) *Camera {
	c := &Camera{World: world, View: view, Zoom: 1}
	c.clamp()
	return c
}

// ScreenToWorld converts a screen position to a world position.
func (c *Camera) ScreenToWorld(sx, sy float64) (wx, wy float64) {
	return sx/c.Zoom + c.Origin.X, sy/c.Zoom + c.Origin.Y
}

// WorldToScreen converts a world position to a screen position.
func (c *Camera) WorldToScreen(wx, wy float64) (sx, sy float64) {
	return (wx - c.Origin.X) * c.Zoom, (wy - c.Origin.Y) * c.Zoom
}

// Scroll moves the view by a screen-space delta.
func (c *Camera) Scroll(dsx, dsy float64) {
	c.Origin.X += dsx / c.Zoom
	c.Origin.Y += dsy / c.Zoom
	c.clamp()
}

// ZoomAt multiplies the zoom by factor, keeping the world point under screen
// position (sx, sy) fixed so the cursor stays on the same hex.
func (c *Camera) ZoomAt(factor, sx, sy float64) {
	wx, wy := c.ScreenToWorld(sx, sy)
	c.Zoom = math.Min(MaxZoom, math.Max(MinZoom, c.Zoom*factor))
	c.Origin = Point{wx - sx/c.Zoom, wy - sy/c.Zoom}
	c.clamp()
}

// Resize changes the view size, keeping the world point at the view centre.
func (c *Camera) Resize(view Size) {
	cx, cy := c.ScreenToWorld(c.View.W/2, c.View.H/2)
	c.View = view
	c.Origin = Point{cx - view.W/2/c.Zoom, cy - view.H/2/c.Zoom}
	c.clamp()
}

// CenterOn puts world point (wx, wy) in the middle of the view (minimap jumps).
func (c *Camera) CenterOn(wx, wy float64) {
	c.Origin = Point{wx - c.View.W/2/c.Zoom, wy - c.View.H/2/c.Zoom}
	c.clamp()
}

// Visible returns the world rectangle shown (for the minimap rectangle and culling).
func (c *Camera) Visible() (x, y, w, h float64) {
	return c.Origin.X, c.Origin.Y, c.View.W / c.Zoom, c.View.H / c.Zoom
}

// clamp keeps the view inside the map; a map smaller than the view is centred.
func (c *Camera) clamp() {
	c.Origin.X = clampAxis(c.Origin.X, c.World.W, c.View.W/c.Zoom)
	c.Origin.Y = clampAxis(c.Origin.Y, c.World.H, c.View.H/c.Zoom)
}

func clampAxis(origin, world, view float64) float64 {
	if world <= view {
		return (world - view) / 2
	}
	return math.Min(math.Max(origin, 0), world-view)
}
