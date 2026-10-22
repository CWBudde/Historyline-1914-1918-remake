package hexgrid

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestCenterRoundTrip(t *testing.T) {
	g := Grid{Width: 8, Lines: 23}
	for line := 0; line < g.Lines; line++ {
		for col := 0; col < g.Width; col++ {
			c := Coord{line, col}
			if !g.Contains(c) {
				continue
			}
			x, y := Center(c)
			for _, d := range [][2]float64{{0, 0}, {-20, 0}, {20, 0}, {0, -30}, {0, 30}, {10, 20}} {
				got, ok := g.PixelToHex(x+d[0], y+d[1])
				if !ok || got != c {
					t.Fatalf("hex %v: point offset %v resolved to %v (on map %v)", c, d, got, ok)
				}
			}
		}
	}
}

func TestPixelToHexOutside(t *testing.T) {
	g := Grid{Width: 8, Lines: 23}
	for _, p := range [][2]float64{{-5, 40}, {-100, -100}, {5000, 40}, {48, 23*40 + 200}, {48, -60}} {
		if c, ok := g.PixelToHex(p[0], p[1]); ok {
			t.Errorf("point %v is outside the map but resolved to %v", p, c)
		}
	}
	// Odd lines are one hex shorter: the slot right of the last odd hex is empty.
	x, y := Center(Coord{1, 8})
	if c, ok := g.PixelToHex(x, y); ok {
		t.Errorf("missing last odd-line hex resolved to %v", c)
	}
	// A hex of an odd line reaches half a hex left of column 0 of the even lines.
	x, y = Center(Coord{1, 0})
	if c, ok := g.PixelToHex(x, y); !ok || c != (Coord{1, 0}) {
		t.Errorf("first odd-line hex = %v, %v", c, ok)
	}
}

// TestPixelToHexMatchesPython compares with findHex of main.py. Python's diagonal
// edges are only approximate (its intercept is 23, not the 40 of the hex outline),
// so disagreement is tolerated only where the point is within 12 pixels of a cell edge.
func TestPixelToHexMatchesPython(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "pixel.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fx struct {
		Width, Lines int
		Samples      [][4]int
	}
	if err := json.Unmarshal(raw, &fx); err != nil {
		t.Fatal(err)
	}
	g := Grid{Width: fx.Width, Lines: fx.Lines}
	same, edge := 0, 0
	for _, s := range fx.Samples {
		// Right of the last column's flat top Python rejects even-line hexes and accepts
		// the missing last hex of odd lines; its edge handling there is simply wrong.
		if s[0] >= (g.Width-1)*ColPitch+74 {
			continue
		}
		x, y := float64(s[0]), float64(s[1])
		got, ok := g.PixelToHex(x, y)
		want := Coord{s[2], s[3]}
		pyOK := s[2] >= 0
		if ok == pyOK && (!ok || got == want) {
			same++
			continue
		}
		if m := edgeMargin(g, x, y); m > 12 {
			t.Errorf("(%v,%v): got %v/%v, Python %v/%v, %.1f px from an edge", x, y, got, ok, want, pyOK, m)
			continue
		}
		edge++
	}
	t.Logf("%d samples agree, %d differ near an edge", same, edge)
}

// edgeMargin is half the difference between the distances to the nearest and the
// second nearest hex centre, which is the distance to the cell edge.
func edgeMargin(g Grid, x, y float64) float64 {
	best, second := math.Inf(1), math.Inf(1)
	line0 := int(math.Floor((y - centerY) / LinePitch))
	for line := line0 - 2; line <= line0+3; line++ {
		for col := -2; col < g.Width+2; col++ {
			cx, cy := Center(Coord{line, col})
			d := math.Hypot(x-cx, y-cy)
			if d < best {
				best, second = d, best
			} else if d < second {
				second = d
			}
		}
	}
	return (second - best) / 2
}

func TestCornersAndBounds(t *testing.T) {
	g := Grid{Width: 8, Lines: 23}
	var maxX, maxY float64
	for line := 0; line < g.Lines; line++ {
		for col := 0; col < g.Width; col++ {
			c := Coord{line, col}
			if !g.Contains(c) {
				continue
			}
			for _, p := range Corners(c) {
				maxX, maxY = math.Max(maxX, p[0]), math.Max(maxY, p[1])
			}
		}
	}
	if w, h := g.Bounds(); w != maxX || h != maxY {
		t.Fatalf("Bounds() = %v×%v, corners reach %v×%v", w, h, maxX, maxY)
	}
	// Hex (0, 0) touches the origin: left vertex at x 0, top edge at y 0.
	cs := Corners(Coord{})
	if cs[0] != [2]float64{0, 40} || cs[1] != [2]float64{25, 0} || cs[3] != [2]float64{96, 40} {
		t.Fatalf("corners of hex 0,0: %v", cs)
	}
	// Neighbours tile without gaps: the top-right corner of a hex is the left
	// vertex of its NE neighbour, its right vertex the neighbour's bottom-left corner.
	c := Coord{4, 3}
	ne := Corners(Around(c)[0])
	own := Corners(c)
	if own[2] != ne[0] || own[3] != ne[5] {
		t.Fatalf("NE edge %v-%v, neighbour %v-%v", own[2], own[3], ne[0], ne[5])
	}
	if w, h := (Grid{}).Bounds(); w != 0 || h != 0 {
		t.Fatalf("empty grid bounds %v×%v", w, h)
	}
}
