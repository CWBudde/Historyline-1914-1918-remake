package mapmaterial

import (
	"math"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

func fixture(cells ...hexgrid.Coord) *level.Level {
	l := &level.Level{Lines: make([][]level.Hex, 15)}
	for y := range l.Lines {
		n := 9
		if y&1 == 1 {
			n--
		}
		l.Lines[y] = make([]level.Hex, n)
		for x := range l.Lines[y] {
			l.Lines[y][x].Terrain = "grass"
		}
	}
	for _, c := range cells {
		l.Lines[c.Line][c.Col].Terrain = "hills"
	}
	return l
}

func TestEmptyAndOutside(t *testing.T) {
	for _, l := range []*level.Level{nil, {}, fixture()} {
		f := New(l)
		for _, p := range [][2]float64{{0, 0}, {100, 100}, {-100, -100}, {math.NaN(), 0}, {0, math.Inf(1)}} {
			if got := f.Weight(p[0], p[1]); got != 0 {
				t.Fatalf("empty weight %v", got)
			}
		}
	}
}

func TestIsolatedActualPolygonAndFiniteEdges(t *testing.T) {
	c := hexgrid.Coord{Line: 6, Col: 3}
	f := New(fixture(c))
	x, y := hexgrid.Center(c)
	for _, p := range [][2]float64{{0, 0}, {0, 20}, {20, 0}} {
		if got := f.Weight(x+p[0], y+p[1]); got != 1 {
			t.Fatalf("interior %v: %v", p, got)
		}
	}
	for _, p := range [][2]float64{{0, -40}, {48, 0}, {23, -40}, {24, -40}, {48, 1}, {-49, 0}, {0, -41}} {
		if got := f.Weight(x+p[0], y+p[1]); got != 0 {
			t.Fatalf("boundary/exterior %v: %v", p, got)
		}
	}
	a, b := f.Weight(x, y-39.9), f.Weight(x, y-39.8)
	if a <= 0 || b <= a {
		t.Fatalf("positive continuous falloff: %v %v", a, b)
	}
}

func TestEverySharedPortAndThreeWayVertex(t *testing.T) {
	// Hex vertices, clockwise; Around ports NE SE S SW NW N correspond to edges 1..5,0.
	v := [6][2]float64{{-23, -40}, {23, -40}, {48, 0}, {23, 40}, {-23, 40}, {-48, 0}}
	for _, line := range []int{6, 7} {
		c := hexgrid.Coord{Line: line, Col: 3}
		x, y := hexgrid.Center(c)
		ns := hexgrid.Around(c)
		for port, n := range ns {
			e := (port + 1) % 6
			a, b := v[e], v[(e+1)%6]
			px, py := x+(a[0]+b[0])/2, y+(a[1]+b[1])/2
			f := New(fixture(c, n))
			if got := f.Weight(px, py); got != 1 {
				t.Fatalf("parity %d port %d shared midpoint=%v", line&1, port, got)
			}
			for _, delta := range []float64{-1e-5, 1e-5} {
				if got := f.Weight(px+delta, py+delta); math.Abs(got-1) > 1e-8 {
					t.Fatalf("shared edge %d: %v", port, got)
				}
			}
			// Add the neighbor on the other edge meeting vertex a: all three cells own that vertex.
			third := ns[(port+5)%6]
			f = New(fixture(c, n, third))
			vx, vy := x+a[0], y+a[1]
			if got := f.Weight(vx, vy); got != 1 {
				t.Fatalf("parity %d port %d three-way vertex=%v", line&1, port, got)
			}
			for _, delta := range []float64{-1e-5, 1e-5} {
				if got := f.Weight(vx+delta, vy+delta); math.Abs(got-1) > 1e-8 {
					t.Fatalf("vertex neighborhood %d: %v", port, got)
				}
			}
		}
	}
}

func TestUnionAndSignature(t *testing.T) {
	c := hexgrid.Coord{Line: 6, Col: 3}
	cells := []hexgrid.Coord{c}
	for _, n := range hexgrid.Around(c) {
		cells = append(cells, n)
	}
	f := New(fixture(cells...))
	x, y := hexgrid.Center(c)
	for _, p := range [][2]float64{{0, 0}, {0, -39.9}, {47.9, 0}, {23, -40}} {
		if got := f.Weight(x+p[0], y+p[1]); got != 1 {
			t.Fatalf("full union interior %v: %v", p, got)
		}
	}
	a := New(fixture(c))
	near := New(fixture(c, hexgrid.Around(c)[0]))
	far := New(fixture(c, hexgrid.Coord{Line: 14, Col: 8}))
	if a.Signature(c) == near.Signature(c) {
		t.Fatal("near geometry omitted")
	}
	if a.Signature(c) != far.Signature(c) {
		t.Fatal("irrelevant distant geometry included")
	}
	for yy := -50.; yy <= 50; yy += .71 {
		for xx := -55.; xx <= 55; xx += .83 {
			w := f.Weight(x+xx, y+yy)
			if math.IsNaN(w) || w < 0 || w > 1 {
				t.Fatalf("invalid weight %v", w)
			}
		}
	}
}
