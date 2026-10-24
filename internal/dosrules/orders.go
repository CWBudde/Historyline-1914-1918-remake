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
//
// A two-cell ship fights as one unit, unlike in the original, where each half
// keeps its own strength: a battle that hits either half leaves the other
// half the same strength and experience, and destroys both together
// (Destroyed lists the other half after the one fought).
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
		for _, k := range []int{o.Attacker, out.Defender} {
			if p, ok := shipPartner(r.Units, k); ok {
				r.Units[p] = whole(r.Units[p], r.Units[k])
				if slices.Contains(out.Destroyed, k) && !slices.Contains(out.Destroyed, p) {
					out.Destroyed = append(out.Destroyed, p)
				}
			}
		}
		r.Orders[len(r.Orders)-1] = out
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

// shipPartner is the other half of a two-cell ship among units: the next
// record from the head (+4 bit 0x80), the previous one from the rear. It
// reports false for any other unit and for a half whose partner is not there.
func shipPartner(units map[int]Unit, i int) (int, bool) {
	f := units[i].Flags4
	if f&0x40 == 0 {
		return 0, false
	}
	p := i - 1
	if f&0x80 != 0 {
		p = i + 1
	}
	q, ok := units[p]
	if !ok || q.Flags4&0x40 == 0 || q.Flags4&0x80 == f&0x80 {
		return 0, false
	}
	return p, true
}

// whole is half of a two-cell ship with the strength and experience of its
// other half.
func whole(half, other Unit) Unit {
	half.Size, half.Alt, half.Exp = other.Size, other.Alt, other.Exp
	return half
}
