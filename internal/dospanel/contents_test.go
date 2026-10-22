package dospanel

import (
	"image"
	"testing"
)

func TestSlotAt(t *testing.T) {
	for _, c := range panelCases {
		t.Run(c.name, func(t *testing.T) {
			l := ContentsLayout(c.panel, c.scale)
			for i, s := range l.Slots {
				x, y := centre(s)
				if got, ok := SlotAt(l, x, y); !ok || got != i {
					t.Errorf("centre of slot %d hits %d, %v", i, got, ok)
				}
			}
			for _, p := range [][2]int{{c.panel.Min.X, c.panel.Min.Y}, centreOf(l.Details), centreOf(l.PhotoWell)} {
				if i, ok := SlotAt(l, p[0], p[1]); ok {
					t.Errorf("%v hits slot %d", p, i)
				}
			}
			// Between two slots of one row lies the slot of the other row.
			a, b := l.Slots[0], l.Slots[2]
			x, y := (a.Max.X+b.Min.X)/2, a.Min.Y+1
			if i, ok := SlotAt(l, x, y); !ok || i != 1 {
				t.Errorf("(%d, %d) between slots 0 and 2 hits %d, %v; want 1", x, y, i, ok)
			}
		})
	}
}

func centreOf(r image.Rectangle) [2]int {
	x, y := centre(r)
	return [2]int{x, y}
}

var opposite = map[Dir]Dir{Up: Down, Down: Up, Left: Right, Right: Left}

func TestSlotNeighbor(t *testing.T) {
	l := ContentsLayout(image.Rect(0, 0, 304, 360), 2)
	for i := range SlotCount {
		for d, back := range opposite {
			j, ok := SlotNeighbor(i, d)
			if !ok {
				continue
			}
			if j < 0 || j >= SlotCount || j == i {
				t.Fatalf("slot %d %v gives %d", i, d, j)
			}
			if k, ok := SlotNeighbor(j, back); !ok || k != i {
				t.Errorf("slot %d %v to %d, back %v gives %d, %v", i, d, j, back, k, ok)
			}
			// The neighbour lies that way on the screen.
			ax, ay := centre(l.Slots[i])
			bx, by := centre(l.Slots[j])
			if d == Up && by >= ay || d == Down && by <= ay || d == Left && bx >= ax || d == Right && bx <= ax {
				t.Errorf("slot %d %v gives %d at (%d, %d) from (%d, %d)", i, d, j, bx, by, ax, ay)
			}
		}
	}
	tests := []struct {
		i    int
		d    Dir
		want int
		ok   bool
	}{
		{0, Left, 0, false},
		{0, Right, 1, true},
		{SlotCount - 1, Right, 0, false},
		{3, Left, 2, true},
		{0, Up, 1, true},
		{1, Down, 0, true},
		{1, Up, 0, false},
		{0, Down, 0, false},
		{2, Centre, 0, false},
		{-1, Right, 0, false},
		{SlotCount, Left, 0, false},
	}
	for _, tt := range tests {
		got, ok := SlotNeighbor(tt.i, tt.d)
		if ok != tt.ok || ok && got != tt.want {
			t.Errorf("SlotNeighbor(%d, %v) = %d, %v; want %d, %v", tt.i, tt.d, got, ok, tt.want, tt.ok)
		}
	}
}
