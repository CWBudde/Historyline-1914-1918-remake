package dosai

import (
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// The extension's work units beyond the gas pioneers (docs/new-units.md):
// wiring parties wire the ground in front of the enemy, tunnellers mine an
// enemy unit next to them and stay until it blows, field ambulances keep
// next to the side's wounded foot units.
const (
	wireRadius     = 4 // a wiring party wires within this distance of an enemy unit
	wireFront      = 2 // a wired hex nearer the nearest enemy unit than the party counts this much
	ambulanceBonus = 6 // per wounded own foot unit next to an ambulance's destination
)

// wirers are the side's wiring parties on the map that may lay wire now and
// have an enemy unit within wireRadius, in hex order.
func (pl *plan) wirers() []int {
	var out []int
	for _, c := range sorted(pl.view.Units) {
		u := pl.view.Units[c]
		if u.Side == pl.side && !u.Moved && pl.s.Wiring(u.Index) && pl.s.CanOrder(pl.side, u.Index) == nil &&
			pl.s.CanWire(pl.side, u.Index) == nil {
			if e := nearest(c, pl.enemies); e >= 0 && e <= wireRadius {
				out = append(out, u.Index)
			}
		}
	}
	return out
}

// wire gives each wiring party the wire order that scores best (wireScore),
// if one scores above 0. A hex an order of this plan wires already counts
// no more. The progress runs from done to done+len(ws) of total.
func (pl *plan) wire(ws []int, done, total int) ([]dosplay.Event, error) {
	var evs []dosplay.Event
	wired := map[hexgrid.Coord]bool{}
	for i, unit := range ws {
		at := pl.s.Game.Units[unit].Cell(pl.side)
		from, _ := pl.s.Board.Coord(at)
		var best hexgrid.Coord
		var bestHexes []hexgrid.Coord
		bestScore := 0
		for _, t := range dosplay.Coords(pl.s.Targets(pl.side, unit)) {
			hexes := pl.s.WireHexes(pl.side, unit, t)
			// Between equal scores the target nearer the enemy.
			sc := pl.wireScore(from, hexes, wired)
			if sc > bestScore || sc == bestScore && sc > 0 && nearest(t, pl.enemies) < nearest(best, pl.enemies) {
				best, bestHexes, bestScore = t, hexes, sc
			}
		}
		if bestScore > 0 {
			ev, err := pl.s.Attack(pl.side, unit, best)
			if err != nil {
				return evs, err
			}
			evs = append(evs, ev...)
			for _, c := range bestHexes {
				wired[c] = true
			}
		}
		pl.report(done+i+1, total)
	}
	return evs, nil
}

// wireScore rates wire from a party on from over hexes: per hex no order of
// this plan wires, 1, or wireFront where it lies nearer the nearest enemy
// unit than the party; 0 when a hex holds a unit or lies next to an own HQ
// entrance, which the wire would close to vehicles.
func (pl *plan) wireScore(from hexgrid.Coord, hexes []hexgrid.Coord, wired map[hexgrid.Coord]bool) int {
	sc := 0
	own := nearest(from, pl.enemies)
	for _, c := range hexes {
		if _, ok := pl.view.Units[c]; ok {
			return 0
		}
		if h := nearest(c, pl.ownHQs); h >= 0 && h <= 1 {
			return 0
		}
		if wired[c] {
			continue
		}
		if nearest(c, pl.enemies) < own {
			sc += wireFront
		} else {
			sc++
		}
	}
	return sc
}

// tunnellers are the side's tunnellers on the map that may start a tunnel
// now, in hex order.
func (pl *plan) tunnellers() []int {
	var out []int
	for _, c := range sorted(pl.view.Units) {
		u := pl.view.Units[c]
		if u.Side == pl.side && !u.Moved && pl.s.Tunneller(u.Index) && pl.s.CanOrder(pl.side, u.Index) == nil &&
			pl.s.CanTunnel(pl.side, u.Index) == nil {
			out = append(out, u.Index)
		}
	}
	return out
}

// mine starts each tunneller's tunnel under the neighbouring enemy unit on
// the map that is worth most: an aircraft or a unit held in a building is
// safe from it. The progress runs from done to done+len(ts) of total.
func (pl *plan) mine(ts []int, done, total int) ([]dosplay.Event, error) {
	var evs []dosplay.Event
	for i, unit := range ts {
		var best hexgrid.Coord
		bestScore := 0
		for _, t := range dosplay.Coords(pl.s.Targets(pl.side, unit)) {
			v, ok := pl.view.Units[t]
			if !ok || v.Side != 1-pl.side || v.Held || v.Domain == dosplay.Air {
				continue
			}
			if sc := v.Size * pl.profile.Tables.Types[v.Type].Cost; sc > bestScore {
				best, bestScore = t, sc
			}
		}
		if bestScore > 0 {
			ev, err := pl.s.Attack(pl.side, unit, best)
			if err != nil {
				return evs, err
			}
			evs = append(evs, ev...)
		}
		pl.report(done+i+1, total)
	}
	return evs, nil
}

// digging reports whether unit digs a tunnel, which it loses if it moves.
func (pl *plan) digging(unit int) bool {
	_, ok := pl.s.TunnelOf(unit)
	return ok
}

// ambulanceScore is what hex c is worth to field ambulance u: ambulanceBonus
// per wounded own foot unit next to it; 0 for any other unit.
func (pl *plan) ambulanceScore(u dosplay.UnitView, c hexgrid.Coord) int {
	if !pl.s.Ambulance(u.Index) {
		return 0
	}
	sc := 0
	for _, n := range hexgrid.Around(c) {
		if v, ok := pl.view.Units[n]; ok && v.Side == pl.side && v.Index != u.Index && v.Domain == dosplay.Infantry && v.Size < v.Full {
			sc += ambulanceBonus
		}
	}
	return sc
}
