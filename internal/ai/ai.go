// Package ai is a greedy compatibility opponent: each call picks the best
// immediate attack, else the best move, for the side to move. It plays only
// through game commands and that side's vision. It is not the original
// game's AI on the R3 order scheduler; its move+attack planning is no
// scheduler contract.
package ai

import (
	"cmp"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Scores of the heuristic.
const (
	hqBonus      = 100000 // capturing the enemy HQ wins the game
	captureBonus = 1000   // capturing a depot or factory
	stepPenalty  = 10     // per step to the nearest goal
	killBonus    = 10     // an attack that destroys the defender
	lossPenalty  = 10     // an attack that loses the attacker
	battleScale  = 10     // a hex's battle modifier is divided by this
)

// Order is one command for the side to move: an attack with weapon Slot of
// the map unit at From on To, or a move of Unit from From to To.
type Order struct {
	Attack   bool
	From, To hexgrid.Coord
	Unit     *game.Unit
	Slot     int
}

// Planner hands out one order at a time. It allows each unit one move and one
// attack per turn, under Rules.OneActionPerTurn one of the two. The zero
// Planner is ready to use.
type Planner struct {
	state    *game.State
	round    int
	turn     game.Player
	attacked map[*game.Unit]bool // given an attack this turn
	moved    map[*game.Unit]bool // given a move this turn
	settled  map[*game.Unit]bool // considered for a move this turn
}

// Next returns the next order for s.Turn, or false when that side has nothing
// left to do. The caller executes it with s.Attack or s.Move. The unit is
// marked when the order is handed out, so a failing order is not repeated.
func (p *Planner) Next(s *game.State) (Order, bool) {
	if s.Winner != game.None {
		return Order{}, false
	}
	p.reset(s)
	if o, ok := p.attack(s); ok {
		p.attacked[o.Unit] = true
		return o, true
	}
	return p.move(s)
}

// reset clears the bookkeeping when a new turn has begun.
func (p *Planner) reset(s *game.State) {
	if p.state == s && p.round == s.Round && p.turn == s.Turn {
		return
	}
	p.state, p.round, p.turn = s, s.Round, s.Turn
	p.attacked = map[*game.Unit]bool{}
	p.moved = map[*game.Unit]bool{}
	p.settled = map[*game.Unit]bool{}
}

// mayAttack reports whether u may still be given an attack.
func (p *Planner) mayAttack(s *game.State, u *game.Unit) bool {
	if u.Attacked || p.attacked[u] {
		return false
	}
	return !s.Rules.OneActionPerTurn || !u.Moved && !p.moved[u]
}

// mayMove reports whether u may still be considered for a move.
func (p *Planner) mayMove(s *game.State, u *game.Unit) bool {
	if u.Moved || p.settled[u] || u.Def.Speed == 0 {
		return false
	}
	return !s.Rules.OneActionPerTurn || !u.Attacked && !p.attacked[u]
}

// attack returns the best worthwhile attack of the side to move, scored by
// Predict. Ties go to the first in line, column and slot order.
func (p *Planner) attack(s *game.State) (Order, bool) {
	var best Order
	bestScore, found := 0, false
	eachHex(s, func(c hexgrid.Coord, h *game.Hex) {
		a := h.Unit
		if a == nil || a.Owner != s.Turn || !p.mayAttack(s, a) {
			return
		}
		for _, t := range s.Targets(c) {
			d := s.At(t).Unit // Targets lists visible hexes only
			steps := hexgrid.Distance(c, t)
			for slot, w := range a.Weapons {
				if w == nil || !s.CanFire(a, slot, d) || steps < max(w.RangeMin, 1) || steps > w.RangeMax {
					continue
				}
				b, err := s.Predict(c, t, slot)
				if err != nil || !worth(b) {
					continue
				}
				if v := score(b); !found || v > bestScore {
					best, bestScore, found = Order{Attack: true, From: c, To: t, Unit: a, Slot: slot}, v, true
				}
			}
		}
	})
	return best, found
}

// worth reports whether an attack is worth making: it hurts the defender, and
// no more than it hurts the attacker unless it destroys the defender.
func worth(b game.Battle) bool {
	return b.Damage > 0 && (b.Damage >= b.CounterDamage || b.DefenderDestroyed)
}

func score(b game.Battle) int {
	v := b.Damage - b.CounterDamage
	if b.DefenderDestroyed {
		v += killBonus
	}
	if b.AttackerDestroyed {
		v -= lossPenalty
	}
	return v
}

// mover is a unit that may move and the hex it starts from.
type mover struct {
	from hexgrid.Coord
	unit *game.Unit
}

// move returns a move for the first unit, in line and column order, whose
// best destination beats staying. Units that stay are not considered again
// this turn.
func (p *Planner) move(s *game.State) (Order, bool) {
	dist := distances(s, goals(s, s.Turn))
	for _, m := range movers(s, s.Turn) {
		if !p.mayMove(s, m.unit) {
			continue
		}
		p.settled[m.unit] = true
		if to, ok := bestMove(s, dist, m); ok {
			p.moved[m.unit] = true
			return Order{From: m.from, To: to, Unit: m.unit}, true
		}
	}
	return Order{}, false
}

// movers lists the units of side that may be picked: map units, units in
// their carriers and units in side's buildings, in line and column order.
func movers(s *game.State, side game.Player) []mover {
	var out []mover
	eachHex(s, func(c hexgrid.Coord, h *game.Hex) {
		switch u := h.Unit; {
		case u != nil && u.Owner == side:
			out = append(out, mover{c, u})
			if u.Content != nil {
				for _, v := range u.Content.Units() {
					out = append(out, mover{c, v})
				}
			}
		case u == nil && h.Container != nil && h.Owner == side:
			for _, v := range h.Container.Units() {
				out = append(out, mover{c, v})
			}
		}
	})
	return out
}

// bestMove returns the destination of m that scores best, if it beats
// staying. Ties go to the first in line and column order.
func bestMove(s *game.State, dist [][]int, m mover) (hexgrid.Coord, bool) {
	armed := slices.ContainsFunc(m.unit.Weapons[:], func(w *data.Weapon) bool { return w != nil })
	best, found := hexgrid.Coord{}, false
	bestScore := 0
	if armed {
		bestScore = position(s, dist, m.from)
	}
	routes := s.Movable(m.from, m.unit)
	dests := make([]hexgrid.Coord, 0, len(routes))
	for c := range routes {
		dests = append(dests, c)
	}
	slices.SortFunc(dests, compare)
	for _, c := range dests {
		v, ok := value(s, dist, c, armed)
		if ok && v > bestScore {
			best, bestScore, found = c, v, true
		}
	}
	return best, found
}

// value scores moving to c; ok is false for a destination not worth taking:
// one that only stores the unit in an own building or carrier, or, for an
// unarmed unit, one that captures nothing.
func value(s *game.State, dist [][]int, c hexgrid.Coord, armed bool) (int, bool) {
	m, _ := s.Remembered(s.Turn, c) // destinations are never Unseen
	if m.Unit != nil && m.Unit.Owner == s.Turn {
		return 0, false
	}
	if h := s.At(c); h.Container != nil {
		switch {
		case m.Owner == s.Turn:
			return 0, false
		case strings.HasPrefix(h.Terrain, "hqN"):
			return hqBonus, true
		}
		return captureBonus, true
	}
	if !armed {
		return 0, false
	}
	return position(s, dist, c), true
}

// position scores standing at c: closer to a goal is better, and cover helps.
func position(s *game.State, dist [][]int, c hexgrid.Coord) int {
	v := 0
	if d := dist[c.Line][c.Col]; d >= 0 {
		v -= stepPenalty * d
	}
	if b := s.At(c).Mods.Battle; b != nil {
		v += *b / battleScale
	}
	return v
}

// goals lists what side heads for: hexes where it remembers an enemy unit,
// and seen buildings it does not own. With none known, it explores: every
// Unseen hex is a goal.
func goals(s *game.State, side game.Player) []hexgrid.Coord {
	var known, unseen []hexgrid.Coord
	eachHex(s, func(c hexgrid.Coord, h *game.Hex) {
		m, ok := s.Remembered(side, c)
		switch {
		case !ok:
			unseen = append(unseen, c)
		case m.Unit != nil && m.Unit.Owner != side,
			h.Container != nil && m.Owner != side:
			known = append(known, c)
		}
	})
	if len(known) > 0 {
		return known
	}
	return unseen
}

// distances returns the steps from every hex to the nearest goal, -1 where
// there is none, by a breadth-first search over the grid (steps equal
// hexgrid.Distance on it).
func distances(s *game.State, goals []hexgrid.Coord) [][]int {
	dist := make([][]int, len(s.Hexes))
	for i := range dist {
		dist[i] = make([]int, len(s.Hexes[i]))
		for j := range dist[i] {
			dist[i][j] = -1
		}
	}
	queue := slices.Clone(goals)
	for _, g := range goals {
		dist[g.Line][g.Col] = 0
	}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		for _, n := range s.Grid.Neighbors(c) {
			if dist[n.Line][n.Col] < 0 {
				dist[n.Line][n.Col] = dist[c.Line][c.Col] + 1
				queue = append(queue, n)
			}
		}
	}
	return dist
}

// eachHex calls f for every hex in line and column order.
func eachHex(s *game.State, f func(hexgrid.Coord, *game.Hex)) {
	for l := range s.Hexes {
		for c := range s.Hexes[l] {
			f(hexgrid.Coord{Line: l, Col: c}, &s.Hexes[l][c])
		}
	}
}

func compare(a, b hexgrid.Coord) int {
	return cmp.Or(cmp.Compare(a.Line, b.Line), cmp.Compare(a.Col, b.Col))
}
