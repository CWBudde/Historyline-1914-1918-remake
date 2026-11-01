package dosai

import (
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Scores of the sappers' work.
const (
	digRadius  = 4 // a sapper digs with an enemy unit this close, or within hqRadius of an own HQ
	holdWeight = 3 // the defence a dug hex gains counts this many times more with a unit on it
	frontBonus = 2 // per dug hex closer to the nearest enemy unit than the sapper
	blockCost  = 8 // per own unit on or next to the dug hexes that cannot cross a trench
)

// digger is a sapper of the side that may get a work order now.
type digger struct {
	unit int
	at   hexgrid.Coord
	near bool // within hqRadius of an own HQ
}

// diggers are the side's sappers on the map that may get their work order
// and have a reason to dig: an enemy unit within digRadius or an own HQ
// within hqRadius. They are in hex order.
func (pl *plan) diggers() []digger {
	var out []digger
	for _, c := range sorted(pl.view.Units) {
		u := pl.view.Units[c]
		if u.Side != pl.side || u.Moved || !pl.s.Sapper(u.Index) || pl.s.CanOrder(pl.side, u.Index) != nil {
			continue
		}
		e, h := nearest(c, pl.enemies), nearest(c, pl.ownHQs)
		near := h >= 0 && h <= hqRadius
		if near || e >= 0 && e <= digRadius {
			out = append(out, digger{unit: u.Index, at: c, near: near})
		}
	}
	return out
}

// work is a work order given in this plan: its target and the hexes it
// changes.
type work struct {
	target hexgrid.Coord
	hexes  []hexgrid.Coord
}

// dig gives each digger the work order that scores best (digScore), if one
// scores above 0. It never orders a fill: a target with a trench, or the
// target or the hexes of an order given before, which the work, in unit
// order after the battles, could find dug and so fill. The progress runs
// from done to done+len(ds) of total.
func (pl *plan) dig(ds []digger, done, total int) ([]dosplay.Event, error) {
	var evs []dosplay.Event
	var given []work
	for i, d := range ds {
		var best work
		bestScore := 0
		for _, t := range dosplay.Coords(pl.s.Targets(pl.side, d.unit)) {
			if v, ok := pl.view.Units[t]; ok && v.Side == 1-pl.side {
				continue
			}
			hexes, recs, fill := pl.s.Work(pl.side, d.unit, t)
			if fill || len(hexes) == 0 || clashes(given, t, hexes) {
				continue
			}
			sc := pl.digScore(d, hexes, recs)
			if d.near {
				sc *= hqFactor
			}
			// Between equal scores the target nearer the enemy.
			if sc > bestScore || sc == bestScore && sc > 0 && nearest(t, pl.enemies) < nearest(best.target, pl.enemies) {
				best, bestScore = work{t, hexes}, sc
			}
		}
		if bestScore > 0 {
			ev, err := pl.s.Attack(pl.side, d.unit, best.target)
			if err != nil {
				return evs, err
			}
			evs = append(evs, ev...)
			given = append(given, best)
		}
		pl.report(done+i+1, total)
	}
	return evs, nil
}

// clashes reports whether a work on target changing hexes meets a work
// given before: either's target among the other's hexes.
func clashes(given []work, target hexgrid.Coord, hexes []hexgrid.Coord) bool {
	for _, w := range given {
		if slices.Contains(w.hexes, target) || slices.Contains(hexes, w.target) {
			return true
		}
	}
	return false
}

// digScore rates a dig on hexes taking the terrain records recs: the
// defence each hex gains, holdWeight times more for a unit of the side on
// it and holdWeight+1 times less for an enemy unit on it, frontBonus for each hex
// closer to the nearest enemy unit than the sapper, and blockCost for each
// own unit on or next to the hexes whose type cannot cross the trench.
func (pl *plan) digScore(d digger, hexes []hexgrid.Coord, recs []int) int {
	t := pl.profile.Tables.Move
	front := nearest(d.at, pl.enemies)
	blocked := map[int]bool{}
	sc := 0
	for k, c := range hexes {
		tr := t.Terrains[recs[k]]
		gain := pl.profile.Tables.Terms[tr.Class].Defence - pl.defence(c)
		sc += gain
		if v, ok := pl.view.Units[c]; ok {
			switch v.Side {
			case pl.side:
				sc += holdWeight * gain
			case 1 - pl.side:
				sc -= (holdWeight + 1) * gain
			}
		}
		if front >= 0 && nearest(c, pl.enemies) < front {
			sc += frontBonus
		}
		around := hexgrid.Around(c)
		for _, n := range append(around[:], c) {
			v, ok := pl.view.Units[n]
			if ok && v.Side == pl.side && !blocked[v.Index] && t.Types[v.Type].Mask&tr.Mask == 0 {
				blocked[v.Index] = true
				sc -= blockCost
			}
		}
	}
	return sc
}
