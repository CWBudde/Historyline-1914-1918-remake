// Package dosai is the computer side on the original order cycle
// (internal/dosplay): a greedy planner that gives all orders of a side's
// phase through the session and confirms them. In the movement phase it
// moves each unit along the cheapest terrain route towards the enemy HQ
// and the enemy units it sees; in the
// action phase it queues the attacks whose exact outcome (dosrules.Engage on
// the side's map as the earlier orders leave it) gains most.
//
// It reads only what the side may know: its view (the other side's units
// where they stood at the last phase change), its battlefield, and the
// session's queries for its own units. It draws no random numbers and sorts
// every list, so equal sessions give equal plans.
package dosai

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Scores of the heuristic.
const (
	stepPenalty    = 10   // per hex
	costPenalty    = 2    // per movement point of the route to a goal (a hex costs 4 to 11)
	unreachable    = 1000 // a hex with no route to a goal
	immobileOffset = 40   // movement points a goal next to an immobile enemy unit counts extra
	adjacentCost   = 50   // a long-range unit next to an enemy unit
	killBonus      = 100
	hqRadius       = 3 // targets this close to an own HQ count hqFactor times
	hqFactor       = 3
	longRangeFlag  = 0x40 // Type.TargetFlags: long range
)

// Planner plans a side's orders. The zero Planner is ready to use.
//
// Progress, if set, is called during each Plan with the units planned so far
// and the units to plan in the phase: first with done 0, last with done
// equal to total. It does not change the plan.
type Planner struct {
	Progress func(done, total int)
}

// Plan gives all orders of side's current phase through the session, then
// confirms. It does nothing once the battle has ended or the side has
// confirmed. A command the session refuses is returned as an error.
func (p *Planner) Plan(s *dosplay.Session, side int) ([]dosplay.Event, error) {
	if side != 0 && side != 1 {
		return nil, fmt.Errorf("dosai: side %d", side)
	}
	if s.Ended() != nil || s.Ready(side) {
		return nil, nil
	}
	pl, err := newPlan(s, side)
	if err != nil {
		return nil, err
	}
	pl.progress = p.Progress
	var evs []dosplay.Event
	if s.Phase(side) == dosrules.PhaseMovement {
		evs, err = pl.move()
	} else {
		evs, err = pl.attack()
	}
	if err != nil {
		return evs, err
	}
	return append(evs, s.Confirm(side)...), nil
}

// PlayOut plays both sides until the battle ends or after maxTurns
// resolutions: each side plans (a nil planner only confirms), then the
// orders resolve. It returns the end, nil while the battle is open.
func PlayOut(s *dosplay.Session, p [2]*Planner, maxTurns int) (*dosrules.End, error) {
	return playOut(s, p, maxTurns, nil)
}

// playOut is PlayOut with every event passed to keep.
func playOut(s *dosplay.Session, p [2]*Planner, maxTurns int, keep func([]dosplay.Event)) (*dosrules.End, error) {
	for range maxTurns {
		if s.Ended() != nil {
			break
		}
		for k := range 2 {
			var evs []dosplay.Event
			if p[k] == nil {
				evs = s.Confirm(k)
			} else {
				var err error
				if evs, err = p[k].Plan(s, k); err != nil {
					return nil, err
				}
			}
			if keep != nil {
				keep(evs)
			}
		}
		evs, err := s.Resolve()
		if err != nil {
			return nil, err
		}
		if keep != nil {
			keep(evs)
		}
	}
	return s.Ended(), nil
}

// plan is the state of one Plan call.
type plan struct {
	s       *dosplay.Session
	side    int
	profile dosrules.Profile
	view    dosplay.View
	bf      *dosrules.Battlefield
	enemies []hexgrid.Coord               // the other side's units in the view, sorted
	hqs     []hexgrid.Coord               // the other side's HQ entrances, sorted
	ownHQs  []hexgrid.Coord               // the side's HQ entrances, sorted
	fields  map[int]map[hexgrid.Coord]int // by unit type

	progress func(done, total int) // Planner.Progress
}

// report passes the plan's progress to the callback, if any.
func (pl *plan) report(done, total int) {
	if pl.progress != nil {
		pl.progress(done, total)
	}
}

func newPlan(s *dosplay.Session, side int) (*plan, error) {
	bf, err := s.Battlefield(side)
	if err != nil {
		return nil, err
	}
	pl := &plan{s: s, side: side, profile: s.Game.Profile, view: s.View(side), bf: bf,
		fields: map[int]map[hexgrid.Coord]int{}}
	for _, c := range sorted(pl.view.Units) {
		if pl.view.Units[c].Side == 1-side {
			pl.enemies = append(pl.enemies, c)
		}
	}
	for _, c := range sorted(pl.view.Buildings) {
		switch b := pl.view.Buildings[c]; {
		case !b.HQ:
		case b.Owner == s.Faction(1-side):
			pl.hqs = append(pl.hqs, c)
		case b.Owner == s.Faction(side):
			pl.ownHQs = append(pl.ownHQs, c)
		}
	}
	return pl, nil
}

// sorted are the keys of a map by hex in line, then column order.
func sorted[V any](m map[hexgrid.Coord]V) []hexgrid.Coord {
	set := make(map[hexgrid.Coord]bool, len(m))
	for c := range m {
		set[c] = true
	}
	return dosplay.Coords(set)
}

// nearest is the distance from c to the nearest hex of cs, -1 for none.
func nearest(c hexgrid.Coord, cs []hexgrid.Coord) int {
	d := -1
	for _, g := range cs {
		if n := hexgrid.Distance(c, g); d < 0 || n < d {
			d = n
		}
	}
	return d
}

// mover is a unit of the side that may get a move or an exit order.
type mover struct {
	unit int
	at   hexgrid.Coord
	held bool
	dist int // to the nearest goal
}

// goals are the hexes the movement phase heads for: the enemy HQ entrances
// and the enemy units.
func (pl *plan) goals() []hexgrid.Coord { return append(slices.Clone(pl.hqs), pl.enemies...) }

// movers are the side's units that may move or exit now, nearest to a goal
// first (ties by line, column and unit).
func (pl *plan) movers() []mover {
	goals := pl.goals()
	var out []mover
	add := func(u dosplay.UnitView, at hexgrid.Coord, held bool) {
		if u.Side != pl.side || pl.s.CanOrder(pl.side, u.Index) != nil {
			return
		}
		out = append(out, mover{unit: u.Index, at: at, held: held, dist: nearest(at, goals)})
	}
	for _, c := range sorted(pl.view.Units) {
		u := pl.view.Units[c]
		add(u, c, false)
		if u.Side == pl.side && u.Carrier {
			if _, slots, ok := pl.s.Contents(pl.side, c); ok {
				for _, v := range slots {
					add(v, c, true)
				}
			}
		}
	}
	for _, c := range sorted(pl.view.Buildings) {
		if b := pl.view.Buildings[c]; b.Owner == pl.s.Faction(pl.side) {
			for _, v := range b.Units {
				add(v, c, true)
			}
		}
	}
	slices.SortStableFunc(out, func(a, b mover) int {
		return cmp.Or(cmp.Compare(a.dist, b.dist), cmp.Compare(a.at.Line, b.at.Line),
			cmp.Compare(a.at.Col, b.at.Col), cmp.Compare(a.unit, b.unit))
	})
	return out
}

// move gives the movement phase's orders.
func (pl *plan) move() ([]dosplay.Event, error) {
	if len(pl.enemies) == 0 && len(pl.hqs) == 0 {
		pl.report(0, 0)
		return nil, nil
	}
	var evs []dosplay.Event
	movers := pl.movers()
	pl.report(0, len(movers))
	for i, m := range movers {
		if pl.s.CanOrder(pl.side, m.unit) == nil {
			ev, err := pl.moveUnit(m)
			if err != nil {
				return evs, err
			}
			evs = append(evs, ev...)
			// The side's own map changed; its view follows.
			pl.view = pl.s.View(pl.side)
		}
		pl.report(i+1, len(movers))
	}
	return evs, nil
}

// moveUnit moves or exits one unit where it scores best, if that beats
// staying.
func (pl *plan) moveUnit(m mover) ([]dosplay.Event, error) {
	u := pl.s.Unit(m.unit)
	var dests map[hexgrid.Coord]bool
	var err error
	if m.held {
		if u.In < 0 {
			return nil, nil
		}
		dests, err = pl.s.ExitDestinations(pl.side, u.In, u.Slot)
	} else {
		dests, err = pl.s.Destinations(pl.side, m.unit)
	}
	if err != nil {
		return nil, err
	}
	to, ok := pl.choose(u, m.at, dosplay.Coords(dests))
	if !ok {
		return nil, nil
	}
	if m.held {
		return pl.s.Exit(pl.side, u.In, u.Slot, to)
	}
	return pl.s.Move(pl.side, m.unit, to)
}

// choose is the destination of a unit standing at from: an enemy HQ
// entrance if it reaches one, else the best scoring hex if it beats from.
func (pl *plan) choose(u dosplay.UnitView, from hexgrid.Coord, dests []hexgrid.Coord) (hexgrid.Coord, bool) {
	for _, c := range dests {
		if slices.Contains(pl.hqs, c) {
			return c, true
		}
	}
	f, err := pl.field(u)
	if err != nil {
		return from, false
	}
	best, bestScore := from, pl.score(u, f, from)
	for _, c := range dests {
		if _, ok := pl.view.Buildings[c]; ok {
			continue // loading, or a building the rules do not capture
		}
		if _, ok := pl.view.Units[c]; ok {
			continue
		}
		if sc := pl.score(u, f, c); sc > bestScore {
			best, bestScore = c, sc
		}
	}
	return best, best != from
}

// score rates hex c for unit u: a lower route cost to the nearest place to
// fight from (field) is better, a long-range unit wants its weapon's full
// reach to the nearest enemy unit and no enemy next to it, and the
// terrain's defence term counts.
func (pl *plan) score(u dosplay.UnitView, f map[hexgrid.Coord]int, c hexgrid.Coord) int {
	sc := pl.defence(c)
	if cost, ok := f[c]; ok {
		sc -= costPenalty * cost
	} else {
		sc -= unreachable + stepPenalty*nearest(c, pl.goals())
	}
	if pl.profile.Tables.Types[u.Type].TargetFlags&longRangeFlag != 0 && len(pl.enemies) > 0 {
		e, d := pl.nearestEnemy(c)
		if reach := pl.reach(u, e); d <= reach {
			sc -= stepPenalty * (reach - d)
		}
		if d <= 1 {
			sc -= adjacentCost
		}
	}
	return sc
}

// reach is the distance unit u's weapon reaches against the enemy unit on
// e, at least 1.
func (pl *plan) reach(u dosplay.UnitView, e hexgrid.Coord) int {
	_, rng := pl.profile.WeaponAgainst(u.Type, pl.flags4(e))
	return max(1, rng-1)
}

// field is the route cost from every hex unit u may enter to the nearest
// place to fight from, on the terrain alone: an enemy HQ entrance, a hex
// next to an enemy unit, for a long-range unit a hex within its reach but
// not next to it. Immobile enemy units count immobileOffset more. Fields
// are kept per type for the Plan call; the goals do not change during it.
func (pl *plan) field(u dosplay.UnitView) (map[hexgrid.Coord]int, error) {
	if f, ok := pl.fields[u.Type]; ok {
		return f, nil
	}
	costs, err := pl.s.Costs(pl.side, u.Index)
	if err != nil {
		return nil, err
	}
	long := pl.profile.Tables.Types[u.Type].TargetFlags&longRangeFlag != 0
	start := map[hexgrid.Coord]int{}
	seed := func(c hexgrid.Coord, v int) {
		if _, ok := costs[c]; !ok {
			return
		}
		if old, ok := start[c]; !ok || v < old {
			start[c] = v
		}
	}
	for _, c := range pl.hqs {
		seed(c, 0)
	}
	for _, e := range pl.enemies {
		off := 0
		if pl.view.Units[e].Budget == 0 {
			off = immobileOffset
		}
		lo, hi := 1, 1
		if long {
			hi = pl.reach(u, e)
			lo = min(2, hi)
		}
		for c := range costs {
			if d := hexgrid.Distance(c, e); d >= lo && d <= hi {
				seed(c, off)
			}
		}
	}
	f := dijkstra(costs, start)
	pl.fields[u.Type] = f
	return f, nil
}

// nearestEnemy is the nearest enemy unit to c (the first in hex order on a
// tie) and its distance.
func (pl *plan) nearestEnemy(c hexgrid.Coord) (hexgrid.Coord, int) {
	best, bd := pl.enemies[0], -1
	for _, e := range pl.enemies {
		if d := hexgrid.Distance(c, e); bd < 0 || d < bd {
			best, bd = e, d
		}
	}
	return best, bd
}

// flags4 is the live +4 word of the unit the side's map shows on c.
func (pl *plan) flags4(c hexgrid.Coord) int {
	if u, ok := pl.bf.Units[pl.view.Units[c].Index]; ok {
		return u.Flags4
	}
	return 0
}

// defence is the defence term of a hex's combat class.
func (pl *plan) defence(c hexgrid.Coord) int {
	cell, ok := pl.s.Board.Cell(c)
	if !ok {
		return 0
	}
	class, ok := pl.bf.Terrain[cell]
	if !ok {
		class = dosrules.DefaultClass
	}
	return pl.profile.Tables.Terms[class].Defence
}

// candidate is a possible attack order.
type candidate struct {
	unit   int
	target hexgrid.Coord
	cell   int
	near   bool // within hqRadius of an own HQ
}

// attack gives the action phase's orders: repeatedly the candidate whose
// battle, fought on the state the orders so far leave, gains most, while one
// gains anything.
func (pl *plan) attack() ([]dosplay.Event, error) {
	var cands []candidate
	for _, c := range sorted(pl.view.Units) {
		u := pl.view.Units[c]
		if u.Side != pl.side || u.Moved {
			continue
		}
		for _, t := range dosplay.Coords(pl.s.Targets(pl.side, u.Index)) {
			if pl.view.Units[t].Side != 1-pl.side {
				continue // never a neutral unit
			}
			cell, _ := pl.s.Board.Cell(t)
			d := nearest(t, pl.ownHQs)
			cands = append(cands, candidate{unit: u.Index, target: t, cell: cell, near: d >= 0 && d <= hqRadius})
		}
	}
	var units []int
	for _, c := range cands {
		if !slices.Contains(units, c.unit) {
			units = append(units, c.unit)
		}
	}
	pl.report(0, len(units))
	var orders []dosrules.Order
	var evs []dosplay.Event
	for len(cands) > 0 {
		state, err := pl.after(orders)
		if err != nil {
			return evs, err
		}
		best, bestScore := -1, 0
		for i, c := range cands {
			if sc, ok := pl.gain(state, c); ok && sc > bestScore {
				best, bestScore = i, sc
			}
		}
		if best < 0 {
			break
		}
		c := cands[best]
		ev, err := pl.s.Attack(pl.side, c.unit, c.target)
		if err != nil {
			return evs, err
		}
		evs = append(evs, ev...)
		orders = append(orders, dosrules.Order{Attacker: c.unit, Target: c.cell})
		cands = slices.DeleteFunc(cands, func(o candidate) bool { return o.unit == c.unit })
		if len(orders) < len(units) {
			pl.report(len(orders), len(units))
		}
	}
	if len(units) > 0 {
		// The units left gain nothing from an attack.
		pl.report(len(units), len(units))
	}
	return evs, nil
}

// after is the side's battlefield as the orders leave it: the new strengths,
// the destroyed units taken off.
func (pl *plan) after(orders []dosrules.Order) (*dosrules.Battlefield, error) {
	if len(orders) == 0 {
		return pl.bf, nil
	}
	r, err := pl.profile.ResolveOrders(pl.bf, orders)
	if err != nil {
		return nil, err
	}
	b := *pl.bf
	b.Units = make(map[int]dosrules.Unit, len(r.Units))
	for i, u := range r.Units {
		if !slices.Contains(r.Gone, i) {
			b.Units[i] = u
		}
	}
	return &b, nil
}

// gain scores an attack on the state: the armour-weighted strength the enemy
// loses less the attacker's, a bonus for each unit destroyed (a cost for the
// attacker), times hqFactor near an own HQ. It is false for an order that
// would not be fought or that the rules cannot resolve.
func (pl *plan) gain(b *dosrules.Battlefield, c candidate) (int, bool) {
	a, ok := b.Units[c.unit]
	if !ok {
		return 0, false
	}
	o, err := pl.profile.Engage(b, c.unit, c.cell)
	if err != nil || !o.Battle {
		return 0, false
	}
	d := b.Units[o.Defender]
	ta, td := pl.profile.Tables.Types[a.Type], pl.profile.Tables.Types[d.Type]
	sc := td.Armour*(strength(td, d.Size, d.Alt)-strength(td, o.DefenderSize, o.DefenderAlt)) -
		ta.Armour*(strength(ta, a.Size, a.Alt)-strength(ta, o.AttackerSize, o.AttackerAlt))
	for _, k := range o.Destroyed {
		if k == c.unit {
			sc -= killBonus
		} else {
			sc += killBonus
		}
	}
	if c.near {
		sc *= hqFactor
	}
	return sc, true
}

// strength is the strength a type counts: its alternate strength for the
// types that count it, else the group size.
func strength(t dosrules.Type, size, alt int) int {
	if t.StrengthFlags&4 != 0 {
		return alt
	}
	return size
}
