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
