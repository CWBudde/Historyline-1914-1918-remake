package llmplay

// Strike wins RHEIN (1918-02:cp) by capturing the EC HQ at resolution 23,
// three resolutions before the CP HQ would fall. It rests on how dosai walks:
// every unit heads for the nearest seen enemy, planes included, so
//
//   - two planes parked at the south edge pull the EC mouth force (lines 30-40,
//     cols 12-16) away from the neck and hold it there; most of it cannot hurt
//     them, and the same force no longer marches on the CP HQ (falls r26, not r18);
//   - the eastern ship sails at r7 with two built cavalry and the four old
//     capturers and lands at {31 17} at r13. Cavalry is the fast capturer;
//     the slower old units stay in the neck, where they absorb what comes back;
//   - the Junkers leaves the bait at r11 and bombs the heavy gun at {19 14}, the
//     one mobile-range threat in the pocket, then the pocket's last mobile units;
//   - landed units stop where the bunkers and long guns reach least, and hold
//     their fire: return fire costs strength the run needs.
//
// A hex next to an enemy unit ends a move, so all routes price it as a lost move.

import (
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Strike is a PhasePlayer for side 0 of RHEIN. The zero value is ready to use;
// use a pointer. It sees only the side's view and is deterministic.
type Strike struct {
	r       int // resolutions played
	landed  map[int]bool
	sailing bool
	view    dosplay.View
}

const (
	strikeSail      = 7  // the resolution the lift ship loads and sails
	strikeJunkers   = 11 // the resolution the Junkers leaves the bait
	strikeShip      = 65 // the eastern transport ship's record
	strikeDanger    = 3  // route cost per unit of fixed fire on a stopping hex
	strikeShipZOC   = 30 // a hex next to an enemy costs a ship its move
	typeCavalry     = 21
	typeAntiTank    = 19
	typeBunker      = 13
	typeHeavyGun    = 12
	typeMediumGun   = 11
	bunkerReach     = 3
	heavyGunReach   = 6
	mediumGunReach  = 5
	heavyGunFire    = 3
	mediumGunFire   = 2
	strikePocketHex = 10 // a heavy gun this near the EC HQ is the Junkers' target
)

var (
	ecHQ       = hexgrid.Coord{Line: 14, Col: 14}
	liftShip   = hexgrid.Coord{Line: 30, Col: 22}
	cpFactory  = hexgrid.Coord{Line: 28, Col: 22}
	landingSea = hexgrid.Coord{Line: 31, Col: 17}
	fokkerBait = hexgrid.Coord{Line: 44, Col: 14}
	junkerBait = hexgrid.Coord{Line: 44, Col: 9}
	liftOld    = []hexgrid.Coord{{Line: 31, Col: 22}, {Line: 30, Col: 23}, {Line: 28, Col: 23}, {Line: 27, Col: 22}}
)

// PlayPhase implements PhasePlayer.
func (st *Strike) PlayPhase(s *dosplay.Session, side int) ([]dosplay.Event, error) {
	if s.Ended() != nil || s.Ready(side) {
		return nil, nil
	}
	if st.landed == nil {
		st.landed = map[int]bool{}
	}
	st.r++
	var evs []dosplay.Event
	var err error
	if s.Phase(side) == dosrules.PhaseMovement {
		evs, err = st.move(s, side)
	} else {
		evs, err = st.act(s, side)
	}
	if err != nil {
		return evs, err
	}
	return append(evs, s.Confirm(side)...), nil
}

func coordsOf[V any](m map[hexgrid.Coord]V) []hexgrid.Coord {
	set := map[hexgrid.Coord]bool{}
	for c := range m {
		set[c] = true
	}
	return dosplay.Coords(set)
}

// ownUnits are the side's units on the map and those held in its carriers and buildings.
func (st *Strike) ownUnits(s *dosplay.Session, side int) (mapped, held []dosplay.UnitView) {
	v := s.View(side)
	for _, c := range coordsOf(v.Units) {
		u := v.Units[c]
		if u.Side != side || u.Head != nil {
			continue
		}
		mapped = append(mapped, u)
		if u.Carrier {
			if _, sl, ok := s.Contents(side, c); ok {
				held = append(held, sl...)
			}
		}
	}
	for _, c := range coordsOf(v.Buildings) {
		if b := v.Buildings[c]; b.Owner == s.Faction(side) {
			if _, sl, ok := s.Contents(side, c); ok {
				held = append(held, sl...)
			}
		}
	}
	return mapped, held
}

// find is a unit of the side on the map by record.
func (st *Strike) find(s *dosplay.Session, side, idx int) (dosplay.UnitView, bool) {
	v := s.View(side)
	for _, c := range coordsOf(v.Units) {
		if u := v.Units[c]; u.Index == idx && u.Side == side {
			return u, true
		}
	}
	return dosplay.UnitView{}, false
}

// addZOC closes enemy hexes and adds pen to the entry cost of the hexes next to one.
func (st *Strike) addZOC(s *dosplay.Session, side int, co map[hexgrid.Coord]int, pen int) {
	v := s.View(side)
	seen := map[hexgrid.Coord]bool{}
	for c, u := range v.Units {
		if u.Side == 1-side {
			delete(co, c)
			for _, n := range hexgrid.Around(c) {
				if _, ok := co[n]; ok && !seen[n] {
					co[n] += pen
					seen[n] = true
				}
			}
		}
	}
	for c, u := range v.Units {
		if u.Side == 1-side {
			delete(co, c)
		}
	}
}

// hqField is the route cost to the EC HQ for unit's type.
func (st *Strike) hqField(s *dosplay.Session, side, unit int) map[hexgrid.Coord]int {
	co, err := s.Costs(side, unit)
	if err != nil {
		return nil
	}
	st.addZOC(s, side, co, s.Unit(unit).Budget)
	return dijkstra(co, map[hexgrid.Coord]int{ecHQ: 0})
}

// danger is the fixed fire on a hex: bunkers and long guns in reach.
func (st *Strike) danger(v dosplay.View, side int, c hexgrid.Coord) int {
	n := 0
	for at, u := range v.Units {
		if u.Side != 1-side {
			continue
		}
		d := hexgrid.Distance(at, c)
		switch {
		case u.Type == typeBunker && d <= bunkerReach:
			n++
		case u.Type == typeHeavyGun && d <= heavyGunReach:
			n += heavyGunFire
		case u.Type == typeMediumGun && d <= mediumGunReach:
			n += mediumGunFire
		}
	}
	return n
}

// sailTo moves a ship along its water towards dest.
func (st *Strike) sailTo(s *dosplay.Session, side int, u dosplay.UnitView, dest hexgrid.Coord) ([]dosplay.Event, error) {
	if u.At == dest || u.Moved {
		return nil, nil
	}
	co, err := s.Costs(side, u.Index)
	if err != nil {
		return nil, err
	}
	st.addZOC(s, side, co, strikeShipZOC)
	f := dijkstra(co, map[hexgrid.Coord]int{dest: 0})
	d, err := s.Destinations(side, u.Index)
	if err != nil {
		return nil, err
	}
	best, bv := u.At, f[u.At]
	if _, ok := f[u.At]; !ok {
		bv = 1 << 28
	}
	for _, c := range dosplay.Coords(d) {
		if v, ok := f[c]; ok && v < bv {
			best, bv = c, v
		}
	}
	if best == u.At {
		return nil, nil
	}
	return s.Move(side, u.Index, best)
}

// lift loads what waits in the factory and the old capturers at strikeSail.
func (st *Strike) lift(s *dosplay.Session, side int, ship dosplay.UnitView) ([]dosplay.Event, error) {
	var evs []dosplay.Event
	_, inF, _ := s.Contents(side, cpFactory)
	for _, u := range inF {
		if u.Side != side || s.CanOrder(side, u.Index) != nil {
			continue
		}
		if d, err := s.ExitDestinations(side, u.In, u.Slot); err == nil && d[ship.At] {
			e, err := s.Exit(side, u.In, u.Slot, ship.At)
			evs = append(evs, e...)
			if err != nil {
				return evs, err
			}
		}
	}
	for _, at := range liftOld {
		if u, ok := s.View(side).Units[at]; ok && u.Side == side {
			if d, err := s.Destinations(side, u.Index); err == nil && d[ship.At] {
				e, err := s.Move(side, u.Index, ship.At)
				evs = append(evs, e...)
				if err != nil {
					return evs, err
				}
			}
		}
	}
	return evs, nil
}

func (st *Strike) move(s *dosplay.Session, side int) ([]dosplay.Event, error) {
	var evs []dosplay.Event
	if ship, ok := st.find(s, side, strikeShip); ok && st.r >= strikeSail {
		if !st.sailing && ship.At == liftShip {
			e, err := st.lift(s, side, ship)
			evs = append(evs, e...)
			if err != nil {
				return evs, err
			}
			st.sailing = true
			ship, _ = st.find(s, side, strikeShip)
		}
		if st.sailing {
			e, err := st.sailTo(s, side, ship, landingSea)
			evs = append(evs, e...)
			if err != nil {
				return evs, err
			}
		}
	}
	mapped, _ := st.ownUnits(s, side)
	st.view = s.View(side)
	for _, u := range mapped {
		if u.Domain != dosplay.Air || u.Moved {
			continue
		}
		e, err := st.goTo(s, side, u, st.waypoint(u))
		evs = append(evs, e...)
		if err != nil {
			return evs, err
		}
	}
	e, err := st.strikers(s, side)
	return append(evs, e...), err
}

// strikers lands the cargo of the arrived ship and marches every landed unit,
// the nearest to the EC HQ first.
func (st *Strike) strikers(s *dosplay.Session, side int) ([]dosplay.Event, error) {
	var evs []dosplay.Event
	mapped, held := st.ownUnits(s, side)
	type mv struct {
		u dosplay.UnitView
		f int
	}
	var list []mv
	for _, u := range held {
		if s.CanOrder(side, u.Index) != nil || u.In < 0 {
			continue
		}
		if v, ok := st.find(s, side, strikeShip); ok && v.At == u.At && u.At == landingSea {
			list = append(list, mv{u, 0})
			st.landed[u.Index] = true
		}
	}
	for _, u := range mapped {
		if st.landed[u.Index] && s.CanOrder(side, u.Index) == nil {
			list = append(list, mv{u, 0})
		}
	}
	for i := range list {
		list[i].f = st.hqField(s, side, list[i].u.Index)[list[i].u.At]
	}
	slices.SortStableFunc(list, func(a, b mv) int { return a.f - b.f })
	for _, m := range list {
		u := m.u
		field := st.hqField(s, side, u.Index)
		var d map[hexgrid.Coord]bool
		var err error
		if u.Held {
			d, err = s.ExitDestinations(side, u.In, u.Slot)
		} else {
			d, err = s.Destinations(side, u.Index)
		}
		if err != nil {
			continue
		}
		v := s.View(side)
		score := func(c hexgrid.Coord, x int) int { return x + strikeDanger*st.danger(v, side, c) }
		best, bv := u.At, 1<<28
		if !u.Held {
			if x, ok := field[u.At]; ok {
				bv = score(u.At, x)
			}
		}
		for _, c := range dosplay.Coords(d) {
			if o, occ := v.Units[c]; occ && o.Side == side {
				continue
			}
			if c == ecHQ {
				best, bv = c, -1<<28
				continue
			}
			if x, ok := field[c]; ok && score(c, x) < bv {
				best, bv = c, score(c, x)
			}
		}
		if best == u.At {
			continue
		}
		var e []dosplay.Event
		if u.Held {
			e, err = s.Exit(side, u.In, u.Slot, best)
		} else {
			e, err = s.Move(side, u.Index, best)
		}
		evs = append(evs, e...)
		if err != nil {
			return evs, err
		}
	}
	return evs, nil
}

// waypoint is a plane's goal: the south edge, for the Junkers from strikeJunkers
// the heavy gun by the EC HQ, then the mobile unit nearest the EC HQ in the pocket.
func (st *Strike) waypoint(u dosplay.UnitView) hexgrid.Coord {
	if !strings.Contains(u.ID, "junkers") {
		return fokkerBait
	}
	if st.r < strikeJunkers {
		return junkerBait
	}
	for _, c := range coordsOf(st.view.Units) {
		if x := st.view.Units[c]; x.Side == 1 && x.Type == typeHeavyGun && hexgrid.Distance(c, ecHQ) < strikePocketHex {
			return c
		}
	}
	best, bd := junkerBait, 1<<20
	for _, c := range coordsOf(st.view.Units) {
		x := st.view.Units[c]
		if x.Side != 1 || x.Budget == 0 || x.Domain == dosplay.Air || c.Col < 12 || c.Col > 17 || c.Line > 28 {
			continue
		}
		if d := hexgrid.Distance(c, ecHQ); d < bd {
			best, bd = c, d
		}
	}
	return best
}

// goTo moves u to the free destination nearest wp.
func (st *Strike) goTo(s *dosplay.Session, side int, u dosplay.UnitView, wp hexgrid.Coord) ([]dosplay.Event, error) {
	d, err := s.Destinations(side, u.Index)
	if err != nil {
		return nil, nil //nolint:nilerr // A plane that may not move stays.
	}
	v := s.View(side)
	best, bv := u.At, hexgrid.Distance(u.At, wp)
	for _, c := range dosplay.Coords(d) {
		if _, occ := v.Units[c]; occ {
			continue
		}
		if x := hexgrid.Distance(c, wp); x < bv {
			best, bv = c, x
		}
	}
	if best == u.At {
		return nil, nil
	}
	return s.Move(side, u.Index, best)
}

// act fires with everything but the landed runners (exact attacks, as Coach
// values them), then builds cavalry with all the energy there is.
func (st *Strike) act(s *dosplay.Session, side int) ([]dosplay.Event, error) {
	p := newCoachPlan(s, side, nil)
	p.own = slices.DeleteFunc(p.own, func(u dosplay.UnitView) bool { return st.landed[u.Index] && u.Type != typeAntiTank })
	evs, err := p.attacks()
	if err != nil {
		return evs, err
	}
	v := s.View(side)
	if b, ok := v.Buildings[cpFactory]; ok && b.Owner == s.Faction(side) {
		for s.CanBuild(side, b.Hold, typeCavalry) == nil {
			e, err := s.Build(side, b.Hold, typeCavalry)
			evs = append(evs, e...)
			if err != nil {
				return evs, err
			}
		}
	}
	return evs, nil
}
