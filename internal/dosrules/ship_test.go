package dosrules

import (
	"slices"
	"testing"
)

// A two-cell ship is one unit: battles against its rear leave the head the
// same strength, and the ship goes down whole.
func TestShipFightsWhole(t *testing.T) {
	for build, p := range profiles {
		b := &Battlefield{Width: 10, Height: 10, Units: map[int]Unit{
			0: {Type: 34, Cell: 22, Size: 6, Flags4: 0xC9},
			1: {Type: 35, Cell: 12, Size: 6, Flags4: 0x49},
			2: {Type: 32, Cell: 24, Size: 6, Flags4: 0xCA},
			3: {Type: 33, Cell: 23, Size: 6, Flags4: 0x4A},
		}}
		var orders []Order
		for range 12 {
			orders = append(orders, Order{0, 23})
		}
		r, err := p.ResolveOrders(b, orders)
		if err != nil {
			t.Fatal(err)
		}
		fought := 0
		for _, o := range r.Orders {
			if o.Battle {
				fought++
			}
		}
		head, rear := r.Units[2], r.Units[3]
		if fought == 0 || head.Size != rear.Size || head.Exp != rear.Exp || head.Size == 6 {
			t.Errorf("%s: %d battles leave head %+v, rear %+v", build, fought, head, rear)
		}
		if gone := slices.Contains(r.Gone, 3); gone != slices.Contains(r.Gone, 2) || gone != (rear.Size == 0) {
			t.Errorf("%s: gone %v with head %+v, rear %+v", build, r.Gone, head, rear)
		}
		if !slices.Contains(r.Gone, 3) {
			t.Errorf("%s: twelve attacks leave the destroyer at %d", build, rear.Size)
		}
	}
}
