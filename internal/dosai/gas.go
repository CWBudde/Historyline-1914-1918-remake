package dosai

import (
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Scores of the gas pioneers' release (docs/new-units.md): a cloud is worth
// releasing over at least gasEnemies enemy units and never over an own one.
const (
	gasEnemies = 2  // enemy units a cloud must cover
	gasFoot    = 3  // the worth of a foot unit in a cloud (it loses the full rate)
	gasOther   = 1  // the worth of another ground unit or a ship (a third of it)
	gasCost    = 60 // a hex in a cloud, for a unit that would end its move there
)

// gasWorth is what a unit in a cloud is worth: nothing for an aircraft.
func gasWorth(u dosplay.UnitView) int {
	switch u.Domain {
	case dosplay.Air:
		return 0
	case dosplay.Infantry:
		return gasFoot
	}
	return gasOther
}

// gassers are the side's gas pioneers on the map that may release gas now,
// in hex order.
func (pl *plan) gassers() []int {
	var out []int
	for _, c := range sorted(pl.view.Units) {
		u := pl.view.Units[c]
		if u.Side == pl.side && !u.Moved && pl.s.GasPioneer(u.Index) && pl.s.CanOrder(pl.side, u.Index) == nil &&
			pl.s.CanRelease(pl.side, u.Index) == nil {
			out = append(out, u.Index)
		}
	}
	return out
}

// release gives each gas pioneer the release that scores best (gasScore),
// if one covers enough enemy units and none of the side's. A hex a cloud
// of this plan already covers counts no more. The progress runs from done
// to done+len(gs) of total.
func (pl *plan) release(gs []int, done, total int) ([]dosplay.Event, error) {
	var evs []dosplay.Event
	covered := map[hexgrid.Coord]bool{}
	for i, unit := range gs {
		var best hexgrid.Coord
		var bestHexes []hexgrid.Coord
		bestScore := 0
		for _, t := range dosplay.Coords(pl.s.Targets(pl.side, unit)) {
			hexes := pl.s.Cloud(pl.side, unit, t)
			if sc := pl.gasScore(hexes, covered); sc > bestScore {
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
				covered[c] = true
			}
		}
		pl.report(done+i+1, total)
	}
	return evs, nil
}

// gasScore rates a cloud over hexes: the worth of the enemy units on hexes
// no cloud covers yet; 0 with an own unit in it or fewer than gasEnemies
// enemy units.
func (pl *plan) gasScore(hexes []hexgrid.Coord, covered map[hexgrid.Coord]bool) int {
	sc, n := 0, 0
	for _, c := range hexes {
		v, ok := pl.view.Units[c]
		if !ok {
			continue
		}
		if v.Side == pl.side {
			return 0
		}
		if v.Side != 1-pl.side || covered[c] || pl.s.GasAt(c) > 1 {
			continue
		}
		if w := gasWorth(v); w > 0 {
			sc += w
			n++
		}
	}
	if n < gasEnemies {
		return 0
	}
	return sc
}

// gasPenalty is what ending a move on hex c costs unit u: gasCost on a
// gassed hex for any unit but an aircraft.
func (pl *plan) gasPenalty(u dosplay.UnitView, c hexgrid.Coord) int {
	if u.Domain == dosplay.Air || pl.s.GasAt(c) == 0 {
		return 0
	}
	return gasCost
}
