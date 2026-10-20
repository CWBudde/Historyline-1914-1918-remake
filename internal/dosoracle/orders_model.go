package dosoracle

import "slices"

// Order is one attack order: the live unit Attacker against the cell Target.
type Order struct{ Attacker, Target int }

// OrderList is one run of the order handler over a list of attack orders.
// The battlefield is the Engagement's (its Attacker and Target are unused):
// Units stand where the attacking side Side sees them. Orders is that side's
// list in list order; Idle is the other side's list, which the handler only
// sees the count of.
type OrderList struct {
	Engagement
	Orders []Order
	Idle   []int
}

// Resolution is what a handler run over an order list leaves.
type Resolution struct {
	Orders []Outcome       // per order in list order; Battle false: skipped
	Units  map[int]Fighter // every placed unit afterwards
	// Cells is the unit byte of every cell of the attacking side's map
	// afterwards (0xff empty).
	Cells []int
	// Marked are the live units whose +0x14 gained bit 4 (the destroy
	// helper's mark), ascending.
	Marked []int
}

// ResolveOrders transcribes the handler's loop (02dc:0117–0d55): it takes
// the orders in list order and resolves each as Engage on the state the
// earlier ones left. There is no alive check: an order whose target cell
// holds no unit byte is skipped, any other is fought. The destroy helper
// (02dc:2083), called for each unit the experience window destroys, writes
// 0xff over that unit's cell of the attacking side's map and sets bit 4 of
// its live +0x14; the unit's record keeps its cell and its strength. So a
// destroyed defender is no longer a target, a destroyed attacker no longer
// supports, and both stay off the attacking side's map. Idle is not read.
func ResolveOrders(g *OrderList, rec func(int) []byte, terms [9][2]int) (Resolution, error) {
	units := map[int]Fighter{}
	for i, u := range g.Units {
		units[i] = u
	}
	gone := map[int]bool{}
	var r Resolution
	for _, o := range g.Orders {
		eg := g.Engagement
		eg.Units = map[int]Fighter{}
		for i, u := range units {
			if !gone[i] {
				eg.Units[i] = u
			}
		}
		eg.Attacker, eg.Target = o.Attacker, o.Target
		out, err := Engage(&eg, rec, terms)
		if err != nil {
			return r, err
		}
		r.Orders = append(r.Orders, out)
		if !out.Battle {
			continue
		}
		a, d := units[o.Attacker], units[out.Defender]
		a.Size, a.Alt, a.Exp = out.AttackerSize, out.AttackerAlt, out.AttackerExp
		units[o.Attacker] = a
		d.Size, d.Alt, d.Exp = out.DefenderSize, out.DefenderAlt, out.DefenderExp
		units[out.Defender] = d
		for _, k := range out.Destroyed {
			gone[k] = true
		}
	}
	r.Units = units
	r.Cells = make([]int, g.Width*g.Height)
	for i := range r.Cells {
		r.Cells[i] = 0xFF
	}
	for i, u := range g.Units {
		if !gone[i] {
			r.Cells[u.Cell] = i
		}
	}
	for i := range gone {
		r.Marked = append(r.Marked, i)
	}
	slices.Sort(r.Marked)
	return r, nil
}
