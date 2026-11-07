package dosscreen

import (
	"math"
	"testing"
)

// The steel texture tiles seamlessly: across the wrap it changes no more than
// between any two neighbouring pixels inside the tile.
func TestSteelTiles(t *testing.T) {
	inside := 0.0
	for y := range steelTile - 1 {
		for x := range steelTile - 1 {
			inside = math.Max(inside, math.Max(math.Abs(steelAt(x+1, y)-steelAt(x, y)), math.Abs(steelAt(x, y+1)-steelAt(x, y))))
		}
	}
	for k := range steelTile {
		if d := math.Abs(steelAt(steelTile-1, k) - steelAt(steelTile, k)); d > inside+1e-12 {
			t.Errorf("row %d jumps by %.4f at the wrap (inside at most %.4f)", k, d, inside)
		}
		if d := math.Abs(steelAt(k, -1) - steelAt(k, 0)); d > inside+1e-12 {
			t.Errorf("column %d jumps by %.4f at the wrap (inside at most %.4f)", k, d, inside)
		}
	}
}
