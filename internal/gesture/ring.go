package gesture

import (
	"math"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// slant is the horizontal run of a plate's slanted edge in world px, as in
// hexgrid.Corners.
const slant = hexgrid.HexWidth - hexgrid.HalfPitch

// PlateCorners returns the corners of a plate centred on (cx, cy) at scale
// device px per world px: the hexgrid cell polygon, in the order of
// hexgrid.Corners.
func PlateCorners(cx, cy, scale float64) [6][2]float64 {
	c := hexgrid.Corners(hexgrid.Coord{})
	x0, y0 := hexgrid.Center(hexgrid.Coord{})
	for i := range c {
		c[i] = [2]float64{cx + (c[i][0]-x0)*scale, cy + (c[i][1]-y0)*scale}
	}
	return c
}

// InPlate reports whether (x, y) lies on the plate centred on (cx, cy).
func InPlate(x, y, cx, cy, scale float64) bool {
	dx, dy := math.Abs(x-cx)/scale, math.Abs(y-cy)/scale
	const hw, hh = hexgrid.HexWidth / 2, hexgrid.HexHeight / 2
	return dy <= hh && dx <= hw-slant*dy/hh
}

// RingPositions returns n points on a circle of radius around (cx, cy),
// evenly spaced clockwise from the top.
func RingPositions(n int, cx, cy, radius float64) [][2]float64 {
	if n <= 0 {
		return nil
	}
	out := make([][2]float64, n)
	for i := range out {
		x, y := ringPoint(i, n, cx, cy, radius)
		out[i] = [2]float64{x, y}
	}
	return out
}

func ringPoint(i, n int, cx, cy, radius float64) (x, y float64) {
	a := 2 * math.Pi * float64(i) / float64(n)
	return cx + radius*math.Sin(a), cy - radius*math.Cos(a)
}

// Ring is the right-click menu: a plate for each action around the cursor
// plate at (X, Y), clockwise from the top.
type Ring struct {
	X, Y    float64
	Radius  float64 // distance of the item centres in device px
	Scale   float64 // device px per world px of every plate
	Actions []Action
}

// ringGap is the clearance between plates in world px.
const ringGap = 8

// NewRing places acts around (x, y) as near as they fit: no plate meets the
// cursor plate or its neighbours (their circumscribed circles keep ringGap
// apart).
func NewRing(x, y, scale float64, acts []Action) Ring {
	d := float64(hexgrid.HexWidth + ringGap) // centre distance that keeps two plates apart
	r := d
	if n := len(acts); n > 2 {
		r = max(r, d/2/math.Sin(math.Pi/float64(n)))
	}
	return Ring{X: x, Y: y, Radius: r * scale, Scale: scale, Actions: acts}
}

// Pos returns the centre of item i.
func (r Ring) Pos(i int) (x, y float64) {
	return ringPoint(i, len(r.Actions), r.X, r.Y, r.Radius)
}

// Hit returns the item whose plate holds (x, y), or -1.
func (r Ring) Hit(x, y float64) int {
	for i := range r.Actions {
		if px, py := r.Pos(i); InPlate(x, y, px, py, r.Scale) {
			return i
		}
	}
	return -1
}
