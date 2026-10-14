// Package hexgrid implements the map layout of the Python reference: hexes are
// stored in lines, where even lines (0, 2, …) hold Width hexes and odd lines hold
// Width-1 hexes shifted half a hex to the right. Line l and line l±2 are stacked
// directly above each other.
package hexgrid

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
	l, col := c.Line, c.Col
	var all [6]Coord
	if l%2 == 1 {
		all = [6]Coord{{l - 1, col + 1}, {l + 1, col + 1}, {l + 2, col}, {l + 1, col}, {l - 1, col}, {l - 2, col}}
	} else {
		all = [6]Coord{{l - 1, col}, {l + 1, col}, {l + 2, col}, {l + 1, col - 1}, {l - 1, col - 1}, {l - 2, col}}
	}
	out := make([]Coord, 0, 6)
	for _, n := range all {
		if g.Contains(n) {
			out = append(out, n)
		}
	}
	return out
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
