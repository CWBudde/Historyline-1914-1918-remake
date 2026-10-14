package game

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// The commands port the rules of gui.py at Python parity, including the gaps
// that Phase 3 closes: movement counts steps, not terrain cost, ammo is never
// used and attacking does not end a unit's turn. Python plays every rule from
// the human's view (info.player, playerTurn); here that is the side to move.

const skillCapture = 1 // unit skill that may enter enemy buildings

var errGameOver = errors.New("the game is over")

// Opponent returns the other side; None for None.
func (p Player) Opponent() Player {
	switch p {
	case CentralPowers:
		return Entente
	case Entente:
		return CentralPowers
	}
	return None
}

// mover returns the unit whose stats decide where u may go from h, and checks
// that the side to move may pick u there (ActionMenu and ContentMenu). As in
// markMovableSquares, a unit leaving a carrier moves with the carrier's stats.
func (s *State) mover(h *Hex, u *Unit) (*Unit, error) {
	switch {
	case u == nil || h == nil:
		return nil, errors.New("no unit")
	case u.Owner != s.Turn:
		return nil, fmt.Errorf("%s belongs to %v", u.Type, u.Owner)
	case u.Moved:
		return nil, fmt.Errorf("%s has already moved", u.Type)
	case h.Unit == u:
		if u.Def.Speed == 0 {
			return nil, fmt.Errorf("%s cannot move", u.Type)
		}
		return u, nil
	case h.Unit != nil && h.Unit.Content != nil && slices.Contains(h.Unit.Content.Units(), u):
		if h.Unit.Owner != s.Turn {
			return nil, errors.New("the carrier belongs to the opponent")
		}
		if s.Rules.OwnStatsOnExit {
			return u, nil
		}
		return h.Unit, nil
	case h.Unit == nil && h.Container != nil && slices.Contains(h.Container.Units(), u):
		if h.Owner == s.Turn.Opponent() {
			return nil, errors.New("the building belongs to the opponent")
		}
		return u, nil
	}
	return nil, fmt.Errorf("%s is not on this hex", u.Type)
}

// Movable returns the hexes u may move to from `from`, each with its route
// (from first). u is the map unit of the hex or a unit inside its building or
// carrier. It is empty when u may not move.
func (s *State) Movable(from hexgrid.Coord, u *Unit) map[hexgrid.Coord]Route {
	m, err := s.mover(s.At(from), u)
	if err != nil || s.Winner != None {
		return nil
	}
	if s.Rules.MovementPoints {
		return s.movablePoints(from, m)
	}
	reach := s.movableSquares(from, m)
	out := map[hexgrid.Coord]Route{}
	for c := range reach {
		if s.Fog(s.Turn, c) == Unseen {
			continue
		}
		if path := s.findPath(from, c, reach); path != nil && len(path)-1 <= m.Def.Speed {
			steps := len(path) - 1
			out[c] = Route{Path: path, Cost: steps, Left: m.Def.Speed - steps}
		}
	}
	return out
}

// movableSquares ports getMovableSquares. Python keeps a list: a building in
// reach is appended once per neighbour in reach, and every obstruction removes
// one copy, so a building next to reachable hexes stays even when it is
// obstructed itself.
func (s *State) movableSquares(from hexgrid.Coord, m *Unit) map[hexgrid.Coord]bool {
	count := map[hexgrid.Coord]int{}
	s.within(from, m.Def.Speed, func(c hexgrid.Coord) { count[c] = 1 })
	delete(count, from)
	obstructed := map[hexgrid.Coord]int{}
	var buildings []hexgrid.Coord
	for c := range count {
		h := s.At(c)
		switch {
		case m.Def.MovementType != data.MoveNormal:
			if !slices.Contains(h.Mods.MovementAllowed, m.Def.MovementType) {
				obstructed[c]++
			}
		case s.Fog(s.Turn, c) == Unseen, h.Mods.Movement == nil:
			obstructed[c]++
		}
		if o := h.Unit; o != nil {
			switch {
			case m.Owner != s.Turn: // Python tests the mover here, not the occupant
				obstructed[c]++
			case o.Content == nil:
				obstructed[c]++
			case !o.Content.Fits(m):
				obstructed[c]++
			case o.Owner != s.Turn:
				obstructed[c]++
			}
		}
		if h.Container != nil && (h.Owner == s.Turn || slices.Contains(m.Def.Skills, skillCapture)) {
			buildings = append(buildings, c)
		}
	}
	for _, b := range buildings {
		for _, n := range s.Grid.Neighbors(b) {
			if count[n] > 0 {
				count[b]++
			}
		}
	}
	reach := map[hexgrid.Coord]bool{}
	for c, n := range count {
		if n > obstructed[c] {
			reach[c] = true
		}
	}
	return reach
}

// findPath ports gui.py findPath after markMovableSquares has marked every hex
// outside reach: a breadth-first search in Python's order that passes hexes
// with own units, and empty hexes in reach that are not unseen.
func (s *State) findPath(from, to hexgrid.Coord, reach map[hexgrid.Coord]bool) []hexgrid.Coord {
	passable := func(c hexgrid.Coord) bool {
		if u := s.At(c).Unit; u != nil {
			return u.Owner == s.Turn
		}
		return reach[c] && s.Fog(s.Turn, c) != Unseen
	}
	paths := [][]hexgrid.Coord{{from}}
	visited := map[hexgrid.Coord]bool{from: true}
	for len(paths) > 0 {
		var next [][]hexgrid.Coord
		for i := len(paths) - 1; i >= 0; i-- {
			p := paths[i]
			for _, n := range s.Grid.Neighbors(p[len(p)-1]) {
				if visited[n] || !passable(n) {
					continue
				}
				visited[n] = true
				np := append(slices.Clip(p), n)
				if n == to {
					return np
				}
				next = append(next, np)
			}
		}
		paths = next
	}
	return nil
}

// Move moves u from `from` to `to` along its path, as gui.py executeMove does:
// into the building at `to` when u fits (capturing it, and winning the game
// with an HQ), else into the carrier there when u fits, else onto the hex.
// It returns the path.
func (s *State) Move(from hexgrid.Coord, u *Unit, to hexgrid.Coord) ([]hexgrid.Coord, error) {
	if s.Winner != None {
		return nil, errGameOver
	}
	if _, err := s.mover(s.At(from), u); err != nil {
		return nil, err
	}
	route, ok := s.Movable(from, u)[to]
	if !ok {
		return nil, fmt.Errorf("%s cannot reach %v", u.Type, to)
	}
	src := s.At(from)
	switch {
	case src.Unit == u:
		src.Unit = nil
	case src.Unit != nil:
		src.Unit.Content.Remove(u)
	default:
		src.Container.Remove(u)
	}
	dst := s.At(to)
	switch {
	case dst.Container != nil && dst.Container.Fits(u):
		if dst.Owner != s.Turn {
			dst.Owner = s.Turn
			if strings.HasPrefix(dst.Terrain, "hqN") {
				s.Winner = s.Turn
			}
		}
		if err := dst.Container.Add(u); err != nil {
			return nil, err
		}
	case dst.Unit != nil && dst.Unit.Content != nil && dst.Unit.Content.Fits(u):
		if err := dst.Unit.Content.Add(u); err != nil {
			return nil, err
		}
	default:
		dst.Unit = u
	}
	u.Moved = true
	s.updateVision()
	return route.Path, nil
}

// Targets returns the hexes the map unit at `from` may attack, sorted, as
// markAttackableSquares finds them: visible enemy map units between the
// smallest minimum and the largest maximum range of all its weapons.
func (s *State) Targets(from hexgrid.Coord) []hexgrid.Coord {
	h := s.At(from)
	if h == nil || h.Unit == nil {
		return nil
	}
	lo, hi := 0, -1
	for _, w := range h.Unit.Weapons {
		if w == nil {
			continue
		}
		if hi < 0 || w.RangeMin < lo {
			lo = w.RangeMin
		}
		hi = max(hi, w.RangeMax)
	}
	var out []hexgrid.Coord
	s.within(from, hi, func(c hexgrid.Coord) {
		if hexgrid.Distance(from, c) < max(lo, 1) || s.Fog(s.Turn, c) != Visible {
			return
		}
		if u := s.At(c).Unit; u != nil && u.Owner != s.Turn {
			out = append(out, c)
		}
	})
	return out
}

// Battle is the outcome of an attack.
type Battle struct {
	Counter           int // defender weapon slot that fired back, -1 for none
	Damage            int // size the defender lost
	CounterDamage     int // size the attacker lost
	AttackerDestroyed bool
	DefenderDestroyed bool
}

// Attack fires weapon slot of the map unit at `from` at the unit at `to`, as
// gui.py handleBattle does. roll returns a number in -5..5 (random.randint); it
// is called for the attacker and then for the defender.
func (s *State) Attack(from, to hexgrid.Coord, slot int, roll func() int) (Battle, error) {
	if s.Winner != None {
		return Battle{}, errGameOver
	}
	fh, th := s.At(from), s.At(to)
	if fh == nil || fh.Unit == nil || fh.Unit.Owner != s.Turn {
		return Battle{}, errors.New("no own unit to attack with")
	}
	a := fh.Unit
	if slot < 0 || slot >= len(a.Weapons) || a.Weapons[slot] == nil {
		return Battle{}, fmt.Errorf("%s has no weapon in slot %d", a.Type, slot)
	}
	if !slices.Contains(s.Targets(from), to) {
		return Battle{}, fmt.Errorf("%v is not a target of %s", to, a.Type)
	}
	steps := hexgrid.Distance(from, to)
	if steps > a.Weapons[slot].RangeMax {
		return Battle{}, fmt.Errorf("%v is out of range of %s", to, a.Weapons[slot].Name)
	}
	if fh.Mods.Battle == nil || th.Mods.Battle == nil {
		return Battle{}, errors.New("a battle hex has no battle modifier")
	}
	d := th.Unit
	b := Battle{}
	b.Counter, b.Damage, b.CounterDamage = resolve(a, d, slot, steps, *fh.Mods.Battle, *th.Mods.Battle, roll)

	d.Size -= b.Damage
	if b.Damage > 0 {
		promote(a)
	}
	if d.Size < 1 {
		d.Size = 0
		promote(a)
		b.DefenderDestroyed = true
	}
	a.Size -= b.CounterDamage
	if b.CounterDamage > 0 {
		promote(d)
	}
	if a.Size < 1 {
		a.Size = 0
		promote(d)
		b.AttackerDestroyed = true
	}
	a.Attacked = true
	if b.DefenderDestroyed {
		th.Unit = nil
	}
	if b.AttackerDestroyed {
		fh.Unit = nil
	}
	s.updateVision()
	return b, nil
}

func promote(u *Unit) {
	if u.Experience < 10 {
		u.Experience++
	}
}

// resolve computes a battle as handleBattle does, from the step distance and
// the battle modifiers of both hexes. It returns the defender's counter slot
// (-1 for none) and the size each side loses.
func resolve(a, d *Unit, slot, steps, aBattle, dBattle int, roll func() int) (counter, damage, counterDamage int) {
	distance := steps - 1 // calculateDistance counts the start hex
	counter = -1
	for i, w := range d.Weapons {
		if w == nil || w.RangeMax < distance || d.Ammo[i] != nil && *d.Ammo[i] == 0 {
			continue
		}
		if counter < 0 || w.Power > d.Weapons[counter].Power {
			counter = i
		}
	}
	strike := func(power int, from, to *Unit, fromBattle, toBattle int) int {
		v := float64(power-to.Def.Armour-distance*3) + float64(fromBattle-toBattle)/3
		v += float64((from.Experience - to.Experience) * 3)
		v += float64((from.Size - to.Size) * 3)
		v += float64(roll())
		if v > 0 {
			return int(v / 10)
		}
		return 0
	}
	damage = strike(a.Weapons[slot].Power, a, d, aBattle, dBattle)
	// Without a counter weapon Python still runs the formula with power 0
	// against nothing: the base attack is 0, not "power - armour".
	if counter >= 0 {
		counterDamage = strike(d.Weapons[counter].Power, d, a, dBattle, aBattle)
	} else {
		counterDamage = strike(a.Def.Armour, d, a, dBattle, aBattle)
	}
	return counter, damage, counterDamage
}

// EndTurn passes the turn to the other side. Python starts a new round when the
// human ends a turn and resets the moved flags before the human moves again;
// here the side that starts its turn gets its flags reset.
func (s *State) EndTurn() error {
	if s.Winner != None {
		return errGameOver
	}
	if s.Turn == s.Human {
		s.Round++
	}
	s.Turn = s.Turn.Opponent()
	for _, u := range s.Units(s.Turn) {
		u.Moved, u.Attacked = false, false
	}
	s.updateVision()
	return nil
}
