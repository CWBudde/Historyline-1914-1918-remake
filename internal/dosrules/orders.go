package dosrules

import "slices"

// Order is one attack order: the unit Attacker against the cell Target.
type Order struct{ Attacker, Target int }

// Resolution is what the order handler leaves after a side's attack list.
type Resolution struct {
	Orders []Outcome    // per order in list order; Battle false: skipped
	Units  map[int]Unit // every unit afterwards, destroyed ones included
	Gone   []int        // the units the destroy helper took off the map, ascending
	Cells  map[int]int  // the occupied cells of the attacking side's map afterwards
}

// ResolveOrders resolves a side's attack list on the battlefield as that side
// sees it, as the handler's loop (02dc:0117–0d55) does: the orders run in
// list order, each one as Engage on the state the earlier ones left. There
// is no alive check: an order whose target cell holds no unit is skipped,
// any other is fought, so a target that survives is fought again. The
// destroy helper (02dc:2083), called for every unit a battle destroys, takes
// it off the attacking side's map at once, so a destroyed defender is no
// longer a target and a destroyed attacker no longer supports. The unit
// keeps its record (its cell and strength), and Engage still runs for an
// order of a destroyed attacker; the attack order never lets a side target
// its own units, so that order cannot arise in the game.
func (p Profile) ResolveOrders(b *Battlefield, orders []Order) (Resolution, error) {
	r := Resolution{Units: map[int]Unit{}, Cells: map[int]int{}}
	for i, u := range b.Units {
		r.Units[i] = u
	}
	gone := map[int]bool{}
	for _, o := range orders {
		view := *b
		view.Units = map[int]Unit{}
		for i, u := range r.Units {
			if !gone[i] {
				view.Units[i] = u
			}
		}
		out, err := p.Engage(&view, o.Attacker, o.Target)
		if err != nil {
			return r, err
		}
		r.Orders = append(r.Orders, out)
		if !out.Battle {
			continue
		}
		a, d := r.Units[o.Attacker], r.Units[out.Defender]
		a.Size, a.Alt, a.Exp = out.AttackerSize, out.AttackerAlt, out.AttackerExp
		d.Size, d.Alt, d.Exp = out.DefenderSize, out.DefenderAlt, out.DefenderExp
		r.Units[o.Attacker], r.Units[out.Defender] = a, d
		for _, k := range out.Destroyed {
			gone[k] = true
		}
	}
	for i, u := range r.Units {
		if gone[i] {
			r.Gone = append(r.Gone, i)
		} else {
			r.Cells[u.Cell] = i
		}
	}
	slices.Sort(r.Gone)
	return r, nil
}
