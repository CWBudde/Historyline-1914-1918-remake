package dospanel

import (
	"fmt"
	"image"
	"math"
	"testing"
)

func TestUnitInfoLayout(t *testing.T) {
	for _, c := range panelCases {
		t.Run(c.name, func(t *testing.T) {
			l := UnitInfoLayout(c.panel, c.scale)
			if l.Scale <= 0 || l.Scale > c.scale {
				t.Fatalf("scale %v for %v", l.Scale, c.scale)
			}
			checkInside(t, c.panel, map[string]image.Rectangle{"plate": l.Plate})
			checkInside(t, l.Plate, map[string]image.Rectangle{
				"photo well": l.PhotoWell, "hex": l.Hex, "name": l.Name, "stats": l.Stats,
			})
			checkInside(t, l.PhotoWell, map[string]image.Rectangle{"photo": l.Photo})
			checkDisjoint(t, map[string]image.Rectangle{
				"photo well": l.PhotoWell, "hex": l.Hex, "name": l.Name, "stats": l.Stats,
			})
			// The photo keeps the reference's 100×84.
			if a := float64(l.Photo.Dx()) / float64(l.Photo.Dy()); math.Abs(a-100.0/84) > 0.06 {
				t.Errorf("photo %v has aspect %.3f", l.Photo, a)
			}
			// Photo and hex side by side at the top, then the name, then the stats.
			if l.Hex.Min.X < l.PhotoWell.Max.X || l.Name.Min.Y < l.PhotoWell.Max.Y || l.Stats.Min.Y < l.Name.Max.Y {
				t.Errorf("order: photo %v hex %v name %v stats %v", l.PhotoWell, l.Hex, l.Name, l.Stats)
			}
		})
	}
}

// The layout uses the given scale while it fits and shrinks to fit the
// smallest panels.
func TestUnitInfoLayoutScale(t *testing.T) {
	if l := UnitInfoLayout(image.Rect(0, 0, 940, 990), 3); l.Scale != 3 {
		t.Errorf("1080p scale %v, want 3", l.Scale)
	}
	if l := UnitInfoLayout(image.Rect(0, 0, 300, 280), 2); l.Scale >= 2 {
		t.Errorf("smallest panel keeps scale %v", l.Scale)
	}
}

func TestContentsLayout(t *testing.T) {
	for _, c := range panelCases {
		t.Run(c.name, func(t *testing.T) {
			l := ContentsLayout(c.panel, c.scale)
			if l.Scale <= 0 || l.Scale > c.scale {
				t.Fatalf("scale %v for %v", l.Scale, c.scale)
			}
			checkInside(t, c.panel, map[string]image.Rectangle{"plate": l.Plate})
			plates := map[string]image.Rectangle{
				"photo well": l.PhotoWell, "info": l.Info, "name": l.Name, "details": l.Details,
			}
			checkInside(t, l.Plate, plates)
			checkInside(t, l.PhotoWell, map[string]image.Rectangle{"photo": l.Photo})
			checkDisjoint(t, plates)
			for i, s := range l.Slots {
				n := fmt.Sprintf("slot %d", i)
				checkInside(t, l.Plate, map[string]image.Rectangle{n: s})
				for pn, p := range plates {
					if s.Overlaps(p) {
						t.Errorf("%s %v overlaps %s %v", n, s, pn, p)
					}
				}
			}
			// Two staggered rows: even slots below, odd slots above, left to right.
			for i := 1; i < SlotCount; i++ {
				a, b := l.Slots[i-1], l.Slots[i]
				if b.Min.X <= a.Min.X {
					t.Errorf("slot %d %v not right of slot %d %v", i, b, i-1, a)
				}
				if i%2 == 1 && b.Min.Y >= a.Min.Y || i%2 == 0 && b.Min.Y <= a.Min.Y {
					t.Errorf("slot %d %v and slot %d %v not staggered", i, b, i-1, a)
				}
				if i >= 2 && l.Slots[i].Min.Y != l.Slots[i-2].Min.Y {
					t.Errorf("slots %d and %d not in one row", i-2, i)
				}
			}
		})
	}
}

func TestGameInfoLayout(t *testing.T) {
	for _, c := range panelCases {
		t.Run(c.name, func(t *testing.T) {
			l := GameInfoLayout(c.panel, c.scale)
			if l.Scale <= 0 || l.Scale > c.scale || l.Pitch <= 0 {
				t.Fatalf("scale %v pitch %v for %v", l.Scale, l.Pitch, c.scale)
			}
			checkInside(t, c.panel, map[string]image.Rectangle{"block": l.Block})
			if n := float64(l.Block.Dy()) / l.Pitch; n < gameInfoLines-0.5 {
				t.Errorf("block %v holds %.1f lines of %v", l.Block, n, l.Pitch)
			}
		})
	}
}
