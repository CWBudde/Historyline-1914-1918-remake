package llmplay

import (
	"cmp"
	"maps"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

const (
	escortFrom = 120 // route cost to the enemy HQ at which the planes join the lead
	sealHexes  = 3   // most hexes a seal may take
	sealGround = 100 // route cost from the HQ of the ground behind the seal
	sealLimit  = 120 // route cost at which an enemy counts as shut out
	plugSize   = 3   // a plug below this size goes to repair if it can be relieved
	sealBroken = 3   // enemies behind the seal that end it
)

// findSeal is the small set of hexes that shuts the own HQ's ground off: with
// them closed no enemy building or unit is within sealLimit on foot, the
// widest movement type. dosai routes over the terrain alone, so its units
// jam behind a plug instead of going round. It takes, again and again, the
// hex of the route to the nearest enemy that lengthens that route most, on a
// tie the one with the fewest open neighbours, then the one nearest the HQ, and gives nil where sealHexes do not do.
func (p *coachPlan) findSeal() []hexgrid.Coord {
	if len(p.ownHQs) == 0 {
		return nil
	}
	hq := p.ownHQs[0]
	var foot *dosplay.UnitView
	for i, u := range p.own {
		if u.Domain == dosplay.Infantry && p.canCapture(u) {
			foot = &p.own[i]
			break
		}
	}
	if foot == nil {
		return nil
	}
	base, err := p.s.Costs(p.side, foot.Index)
	if err != nil {
		return nil
	}
	costs := maps.Clone(base)
	var foes []hexgrid.Coord
	for _, c := range p.foreign {
		if p.s.SideOf(p.view.Buildings[c].Owner) == p.opp {
			foes = append(foes, c)
		}
	}
	for _, e := range p.enemies {
		foes = append(foes, e.At)
	}
	nearestFoe := func(d map[hexgrid.Coord]int) (hexgrid.Coord, int) {
		best, bd := hexgrid.Coord{}, unreachable
		for _, f := range foes {
			if v, ok := d[f]; ok && v < bd {
				best, bd = f, v
			}
		}
		return best, bd
	}
	var seal []hexgrid.Coord
	for len(seal) <= sealHexes {
		d := dijkstra(costs, map[hexgrid.Coord]int{hq: 0})
		foe, bd := nearestFoe(d)
		if bd > sealLimit {
			return seal
		}
		if len(seal) == sealHexes {
			return nil
		}
		best, bestV, found := hexgrid.Coord{}, -1, false
		for at := foe; at != hq; {
			prev := at
			for _, n := range hexgrid.Around(at) {
				if v, ok := d[n]; ok && v+costs[n] == d[at] {
					prev = n
					break
				}
			}
			if prev == at {
				break
			}
			at = prev
			if at == hq || slices.Contains(foes, at) {
				continue
			}
			if _, isB := p.view.Buildings[at]; isB {
				continue
			}
			c := maps.Clone(costs)
			delete(c, at)
			_, v := nearestFoe(dijkstra(c, map[hexgrid.Coord]int{hq: 0}))
			open := func(h hexgrid.Coord) int {
				n := 0
				for _, a := range hexgrid.Around(h) {
					if _, ok := base[a]; ok {
						n++
					}
				}
				return n
			}
			tighter := found && v == bestV && cmp.Or(cmp.Compare(open(at), open(best)), cmp.Compare(hexgrid.Distance(at, hq), hexgrid.Distance(best, hq))) < 0
			if v > bestV || tighter {
				best, bestV, found = at, v, true
			}
		}
		if !found {
			return nil
		}
		delete(costs, best)
		seal = append(seal, best)
	}
	return nil
}

// planSeal assigns a plug to every seal hex: the unit on it while it is fit,
// else the unit that gets there soonest, the toughest on a tie. It marks the
// ground behind the seal, which the units on it do not leave.
func (p *coachPlan) planSeal(seal []hexgrid.Coord) {
	p.seal, p.plug, p.inside = seal, map[int]hexgrid.Coord{}, map[hexgrid.Coord]bool{}
	if len(seal) == 0 {
		return
	}
	var foot *dosplay.UnitView
	for i, u := range p.own {
		if u.Domain == dosplay.Infantry {
			foot = &p.own[i]
			break
		}
	}
	if foot != nil {
		if base, err := p.costsFor(*foot); err == nil {
			c := maps.Clone(base)
			for _, h := range seal {
				delete(c, h)
			}
			for at, v := range dijkstra(c, map[hexgrid.Coord]int{p.ownHQs[0]: 0}) {
				if v <= sealGround {
					p.inside[at] = true
				}
			}
		}
	}
	inNow := 0
	for _, e := range p.enemies {
		if p.inside[e.At] {
			inNow++
		}
	}
	if !holdOn {
		p.inside = map[hexgrid.Coord]bool{}
		if len(p.seal) > 0 {
			for _, h := range seal {
				if u, ok := p.view.Units[h]; ok && u.Side == p.side && u.Size >= plugSize {
					p.plug[u.Index] = h
				}
			}
		}
	}
	if holdOn && inNow >= sealBroken {
		p.seal, p.inside = nil, map[hexgrid.Coord]bool{}
		return
	}
	if holdOn {
		p.guardOf = map[int]hexgrid.Coord{} // with the ground shut, the doorstep needs no guard
	}
	for _, h := range seal {
		if u, ok := p.view.Units[h]; ok && u.Side == p.side && u.Size >= plugSize {
			p.plug[u.Index] = h
		}
	}
	taken := func(h hexgrid.Coord) bool {
		for _, g := range p.plug {
			if g == h {
				return true
			}
		}
		return false
	}
	// A bridge lies under all the guns of the far bank: planes first, which
	// only air power can hurt.
	byBridge := slices.Clone(seal)
	slices.SortStableFunc(byBridge, func(a, b hexgrid.Coord) int { return cmpBool(p.bridges[b], p.bridges[a]) })
	for _, h := range byBridge {
		if taken(h) || !p.bridges[h] {
			continue
		}
		type cand struct {
			idx, air, turns, tough int
		}
		var cands []cand
		for _, u := range p.own {
			if _, used := p.plug[u.Index]; used || u.Held || u.Carrier || u.Domain == dosplay.Naval || u.Budget == 0 || u.Size < plugSize {
				continue
			}
			if holdOn && !p.inside[u.At] && u.Domain != dosplay.Air && !slices.Contains(seal, u.At) {
				continue
			}
			costs, err := p.costsFor(u)
			if err != nil {
				continue
			}
			d := dijkstra(costs, map[hexgrid.Coord]int{h: 0})
			v, ok := d[u.At]
			if !ok {
				continue
			}
			air := 1
			if !holdOn && u.Domain != dosplay.Air {
				continue // the planes alone hold the bridges
			}
			if (u.Domain == dosplay.Air) == p.bridges[h] {
				air = 0
			}
			cands = append(cands, cand{u.Index, air, v / max(u.Budget, 1), u.Armour * u.Size})
		}
		slices.SortFunc(cands, func(a, b cand) int {
			return cmp.Or(cmp.Compare(a.air, b.air), cmp.Compare(a.turns, b.turns), cmp.Compare(b.tough, a.tough), cmp.Compare(a.idx, b.idx))
		})
		if len(cands) > 0 {
			p.plug[cands[0].idx] = h
		}
	}
}

// movePlug moves a plug onto or towards its hex.
func (p *coachPlan) movePlug(u dosplay.UnitView, h hexgrid.Coord) ([]dosplay.Event, error) {
	if u.At == h {
		return nil, nil
	}
	dests, err := p.s.Destinations(p.side, u.Index)
	if err != nil || len(dests) == 0 {
		return nil, nil //nolint:nilerr // A unit without destinations holds position.
	}
	if dests[h] {
		return p.s.Move(p.side, u.Index, h)
	}
	costs, err := p.costsFor(u)
	if err != nil {
		return nil, nil //nolint:nilerr // Unavailable movement costs mean the unit holds position.
	}
	d := dijkstra(costs, map[hexgrid.Coord]int{h: 0})
	best, bestV := u.At, unreachable
	if v, ok := d[u.At]; ok {
		bestV = v
	}
	for _, c := range dosplay.Coords(dests) {
		if v, ok := d[c]; ok && v < bestV {
			if _, isB := p.view.Buildings[c]; isB {
				continue
			}
			best, bestV = c, v
		}
	}
	if best == u.At {
		return nil, nil
	}
	return p.s.Move(p.side, u.Index, best)
}

func cmpBool(a, b bool) int {
	switch {
	case a == b:
		return 0
	case a:
		return 1
	}
	return -1
}

// holdField leads the units behind the seal to their weapon's reach from the
// plugs on the ground, where whatever comes up to the plug is under all guns.
func (p *coachPlan) holdField(u dosplay.UnitView, costs map[hexgrid.Coord]int, seed func(hexgrid.Coord, int)) {
	r := max(p.reach(u), 1)
	for _, h := range p.seal {
		if p.bridges[h] {
			continue
		}
		for c := range costs {
			if p.inside[c] && hexgrid.Distance(c, h) == r {
				seed(c, 0)
			}
		}
	}
}

// moveEscort flies a plane to the free hex ahead of the leading dasher on its
// route: only air power can shift it, and the enemy cannot stand where it
// stands, so the way stays open behind it.
func (p *coachPlan) moveEscort(u dosplay.UnitView, add func([]dosplay.Event, error) error) (bool, error) {
	if p.damaged(u) {
		return false, nil
	}
	var lead *dosplay.UnitView
	leadV := unreachable
	var field map[hexgrid.Coord]int
	for i, o := range p.own {
		if !p.dash[o.Index] {
			continue
		}
		f, err := p.field(o, goalDash)
		if err != nil {
			continue
		}
		if v, ok := f[o.At]; ok && v < leadV && v < escortFrom {
			lead, leadV, field = &p.own[i], v, f
		}
	}
	if lead == nil {
		return false, nil
	}
	dests, err := p.s.Destinations(p.side, u.Index)
	if err != nil {
		return false, nil //nolint:nilerr // A unit that cannot move flies no escort.
	}
	want := leadV - 12
	if u.Index%2 == 1 {
		want = leadV - 24
	}
	best, bestD := u.At, unreachable
	for _, c := range dosplay.Coords(dests) {
		v, ok := field[c]
		if !ok || v >= leadV {
			continue
		}
		if d := max(v-want, want-v); d < bestD {
			best, bestD = c, d
		}
	}
	if best != u.At {
		if err := add(p.s.Move(p.side, u.Index, best)); err != nil {
			return true, err
		}
	}
	return true, nil
}

var passHexes = []hexgrid.Coord{{Line: 27, Col: 18}, {Line: 28, Col: 18}, {Line: 27, Col: 19}, {Line: 28, Col: 19}, {Line: 27, Col: 20}, {Line: 28, Col: 20}, {Line: 29, Col: 17}, {Line: 30, Col: 17}}

// southDasher is every other dasher in the east: two ways in take twice the
// troops, and the pass lets one through per phase.
func (p *coachPlan) southDasher(u dosplay.UnitView) bool {
	return u.Index%2 == 0 && u.At.Col >= 20
}

const dashWaitF = 97

// DashGo is the turn the dash goes.
var DashGo = 16

func (p *coachPlan) turnNow() int { return p.s.Game.Turn }

var holdOn = false
