package gesture

import (
	"math"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestRingPositions(t *testing.T) {
	tests := []struct {
		n    int
		want [][2]float64
	}{
		{1, [][2]float64{{10, -90}}},
		{2, [][2]float64{{10, -90}, {10, 110}}},
		{4, [][2]float64{{10, -90}, {110, 10}, {10, 110}, {-90, 10}}}, // up, right, down, left
	}
	for _, tt := range tests {
		got := RingPositions(tt.n, 10, 10, 100)
		if len(got) != len(tt.want) {
			t.Fatalf("n=%d: %d positions", tt.n, len(got))
		}
		for i := range got {
			if !near(got[i][0], tt.want[i][0]) || !near(got[i][1], tt.want[i][1]) {
				t.Errorf("n=%d: position %d = %v, want %v", tt.n, i, got[i], tt.want[i])
			}
		}
	}
	if RingPositions(0, 0, 0, 1) != nil {
		t.Error("n=0 gave positions")
	}
}

func TestPlateCornersMatchHexgrid(t *testing.T) {
	c := hexgrid.Coord{Line: 3, Col: 2}
	cx, cy := hexgrid.Center(c)
	want := hexgrid.Corners(c)
	got := PlateCorners(cx*0.5+7, cy*0.5-3, 0.5)
	for i := range want {
		if !near(got[i][0], want[i][0]*0.5+7) || !near(got[i][1], want[i][1]*0.5-3) {
			t.Errorf("corner %d = %v, want %v scaled", i, got[i], want[i])
		}
	}
}

func TestInPlate(t *testing.T) {
	const s = 0.5 // plate 48×40 device px
	tests := []struct {
		x, y float64
		want bool
	}{
		{0, 0, true},
		{23, 0, true},
		{-23, 0, true},
		{25, 0, false},
		{0, 19.9, true},
		{0, -20.1, false},
		{11, 19, true},    // under the top-right corner
		{20, 15, false},   // beyond the lower-right slant
		{-20, -15, false}, // beyond the upper-left slant
	}
	for _, tt := range tests {
		if got := InPlate(100+tt.x, 50+tt.y, 100, 50, s); got != tt.want {
			t.Errorf("InPlate(%v, %v) = %v, want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

func TestRingLayoutAndHit(t *testing.T) {
	for n := 1; n <= 9; n++ {
		acts := make([]Action, n)
		r := NewRing(300, 200, 0.75, acts)
		for i := range n {
			x, y := r.Pos(i)
			if r.Hit(x, y) != i {
				t.Errorf("n=%d: Hit(Pos(%d)) = %d", n, i, r.Hit(x, y))
			}
			// Plates keep clear of the cursor plate and of each other: their
			// circumscribed circles (radius half the plate width) do not meet.
			if d := math.Hypot(x-300, y-200); d < hexgrid.HexWidth*0.75 {
				t.Errorf("n=%d: item %d at distance %v overlaps the centre", n, i, d)
			}
			for j := range i {
				xj, yj := r.Pos(j)
				if d := math.Hypot(x-xj, y-yj); d < hexgrid.HexWidth*0.75 {
					t.Errorf("n=%d: items %d and %d are %v apart", n, i, j, d)
				}
			}
		}
		if r.Hit(300, 200) != -1 || r.Hit(-1000, -1000) != -1 {
			t.Errorf("n=%d: centre or far point hit an item", n)
		}
	}
	r := NewRing(0, 0, 1, []Action{Move, Map, Info, ModeChange})
	if x, y := r.Pos(0); !near(x, 0) || y >= 0 {
		t.Errorf("first item at (%v, %v), want straight above", x, y)
	}
}
