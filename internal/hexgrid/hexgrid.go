// Package hexgrid implements the map layout of the Python reference: hexes are
// stored in lines, where even lines (0, 2, …) hold Width hexes and odd lines hold
// Width-1 hexes shifted half a hex to the right. Line l and line l±2 are stacked
// directly above each other.
package hexgrid

import "math"

// Coord addresses a hex by line and column, the (x, y) pair used by hlrData.py.
type Coord struct {
	Line, Col int
}

// Grid is a map of Lines lines, Width hexes wide on even lines.
type Grid struct {
	Width, Lines int
}

// Contains reports whether c lies on the map. Odd lines are one hex shorter.
func (g Grid) Contains(c Coord) bool {
	if c.Line < 0 || c.Line >= g.Lines || c.Col < 0 {
		return false
	}
	if c.Line%2 == 0 {
		return c.Col < g.Width
	}
	return c.Col < g.Width-1
}

// Size returns the number of hexes on the map.
func (g Grid) Size() int {
	even := (g.Lines + 1) / 2
	return even*g.Width + (g.Lines-even)*(g.Width-1)
}

// Neighbors returns the on-map neighbours of c in the same order as the Python
// adjacentHexes.
func (g Grid) Neighbors(c Coord) []Coord {
	out := make([]Coord, 0, 6)
	for _, n := range Around(c) {
		if g.Contains(n) {
			out = append(out, n)
		}
	}
	return out
}

// Around returns all six neighbours of c, on the map or not, in the order NE, SE,
// S, SW, NW, N (Python's adjacentHexes order).
func Around(c Coord) [6]Coord {
	l, col := c.Line, c.Col
	if l&1 == 1 {
		return [6]Coord{{l - 1, col + 1}, {l + 1, col + 1}, {l + 2, col}, {l + 1, col}, {l - 1, col}, {l - 2, col}}
	}
	return [6]Coord{{l - 1, col}, {l + 1, col}, {l + 2, col}, {l + 1, col - 1}, {l - 1, col - 1}, {l - 2, col}}
}

// Cube converts c to cube coordinates (q+r+s == 0). Doubling the column axis
// (2*Col, plus one on odd lines) gives doubled coordinates whose neighbours are
// (±1, ±1) and (0, ±2), which map directly onto cube axes.
func Cube(c Coord) (q, r, s int) {
	q = 2*c.Col + c.Line%2
	r = (c.Line - q) / 2
	return q, r, -q - r
}

// Distance returns the number of steps between a and b. It replaces the
// breadth-first search in gui.py calculateDistance.
func Distance(a, b Coord) int {
	aq, ar, as := Cube(a)
	bq, br, bs := Cube(b)
	return (abs(aq-bq) + abs(ar-br) + abs(as-bs)) / 2
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// Layout of the Python reference in map pixels, with (0, 0) at the top-left corner
// of the bounding box of hex (0, 0). Hexes are flat-topped, 96×80 pixels. A hex sits
// ColPitch pixels from its same-line neighbour and LinePitch pixels from the next
// line; odd lines are shifted by HalfPitch. This reproduces gui.py/main.py, where
// the map surface advances 142×40 pixels per hex column/line.
const (
	ColPitch  = 142
	HalfPitch = ColPitch / 2
	LinePitch = 40
	HexWidth  = 96
	HexHeight = 80
	centerX   = HexWidth / 2
	centerY   = HexHeight / 2
)

// Center returns the pixel centre of hex c. It is defined for any coordinate,
// on the map or not.
func Center(c Coord) (x, y float64) {
	return float64(c.Col*ColPitch + (c.Line&1)*HalfPitch + centerX), float64(c.Line*LinePitch + centerY)
}

// PixelToHex returns the hex under map pixel (x, y) and whether that hex is on the
// map. It replaces Python's hard-coded findHex: the cell of a hex is the set of
// points closer to its centre than to any other centre, which covers both line
// parities, the half-hex gaps at odd-line edges and points outside the map.
func (g Grid) PixelToHex(x, y float64) (Coord, bool) {
	line0 := int(math.Floor((y - centerY) / LinePitch))
	best, bestD := Coord{}, math.Inf(1)
	for line := line0 - 1; line <= line0+2; line++ {
		shift := float64((line & 1) * HalfPitch)
		col0 := int(math.Floor((x - centerX - shift) / ColPitch))
		for col := col0; col <= col0+1; col++ {
			c := Coord{line, col}
			cx, cy := Center(c)
			if d := (x-cx)*(x-cx) + (y-cy)*(y-cy); d < bestD {
				best, bestD = c, d
			}
		}
	}
	return best, g.Contains(best)
}

// Corners returns the six corners of hex c in map pixels, clockwise from the
// left vertex: left, top left, top right, right, bottom right, bottom left.
func Corners(c Coord) [6][2]float64 {
	x, y := Center(c)
	const q = HexWidth - HalfPitch // run of a slanted edge, so that neighbours share edges exactly
	return [6][2]float64{
		{x - centerX, y}, {x - centerX + q, y - centerY}, {x + centerX - q, y - centerY},
		{x + centerX, y}, {x + centerX - q, y + centerY}, {x - centerX + q, y + centerY},
	}
}

// Bounds returns the size of the map in pixels: the bounding box of all its hexes,
// with (0, 0) at the top-left corner of hex (0, 0). The widest line decides the
// width, so a one-line map is as wide as its Width hexes.
func (g Grid) Bounds() (w, h float64) {
	if g.Width <= 0 || g.Lines <= 0 {
		return 0, 0
	}
	w = float64((g.Width-1)*ColPitch + HexWidth)
	if g.Lines > 1 && g.Width > 1 {
		w = max(w, float64((g.Width-2)*ColPitch+HalfPitch+HexWidth))
	}
	return w, float64((g.Lines-1)*LinePitch + HexHeight)
}
