package dosplay

import (
	"maps"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Fog of war (Options.Fog, the modern rules) adds a line of sight to the
// original's maps. A side sees a hex within the sight (data.Unit.Sight, in
// hexes) of one of its units on its own map, or within BuildingSight of the
// entrance of one of its buildings; a held unit does not look out. Fog
// hides the other side's and neutral units out of sight:
//
//   - View leaves them out; a two-cell ship shows while either half is in
//     sight;
//   - Targets skips their hexes and Attack refuses them (no target), except
//     for a sapper's work, which goes to a hex whether a unit is on it or not;
//   - Resolve drops the Revealed moves whose goal the action side does not
//     see after the phase change, and EventsFor filters any event list for
//     a side.
//
// Fog does not hide the terrain, the buildings or their owners (a building
// shows its true owner), the battles a side's own units fight, or what the
// rules find on the side's map: Destinations, Route and Move still avoid a
// hex a hidden unit holds, a move refused there says so, and Battlefield
// and the outcome of an attack count the hidden units' support terms.
// TrueView, Unit and Orders stay unfiltered. Fog keeps no memory: a unit
// that leaves sight leaves the view.

// BuildingSight is how far a side's building sees around its entrance under
// fog.
const BuildingSight = 2

// DefaultSight is the sight of a unit the remake has no unit id for.
const DefaultSight = 2

// Fog reports whether the session plays with fog of war.
func (s *Session) Fog() bool { return s.Game.Fog }

// eye is a hex a side looks from and how far.
type eye struct {
	at    hexgrid.Coord
	sight int
}

// eyes are the hexes side looks from: its units on its map and the
// entrances of its buildings, in cell and record order.
func (s *Session) eyes(side int) []eye {
	var out []eye
	m := s.Game.Maps[side]
	for c := range s.Game.Width * s.Game.Height {
		occ := m[2*c+1]
		u, live := s.Game.Units[occ]
		p, ok := s.Board.Coord(c)
		if occ > dosrules.SlotFree || !live || !ok || u.Destroyed() || sideOf(u.Flags4) != side {
			continue
		}
		sight := DefaultSight
		if d, ok := s.set.Units[s.unitID(occ)]; ok {
			sight = d.Sight
		}
		out = append(out, eye{at: p, sight: sight})
	}
	for _, i := range slices.Sorted(maps.Keys(s.Game.Holds)) {
		h := s.Game.Holds[i]
		if i < dosrules.HoldFlag40 || sideOf(h.Flags) != side {
			continue
		}
		if p, ok := s.Board.Coord(h.Cells[0] / 2); ok {
			out = append(out, eye{at: p, sight: BuildingSight})
		}
	}
	return out
}

// sees reports whether one of eyes sees c.
func sees(eyes []eye, c hexgrid.Coord) bool {
	for _, e := range eyes {
		if hexgrid.Distance(e.at, c) <= e.sight {
			return true
		}
	}
	return false
}

// fogged reports whether fog hides anything from side: fog is on and side is
// a player side.
func (s *Session) fogged(side int) bool { return s.Game.Fog && (side == 0 || side == 1) }

// InSight reports whether side sees hex c now; without fog every hex of the
// level.
func (s *Session) InSight(side int, c hexgrid.Coord) bool {
	if _, ok := s.Board.Cell(c); !ok {
		return false
	}
	return !s.fogged(side) || sees(s.eyes(side), c)
}

// Sight is the set of hexes side sees now under fog, for a front end to
// shade the rest; nil without fog, when it sees every hex.
func (s *Session) Sight(side int) map[hexgrid.Coord]bool {
	if !s.fogged(side) {
		return nil
	}
	eyes := s.eyes(side)
	out := map[hexgrid.Coord]bool{}
	for c := range s.Game.Width * s.Game.Height {
		if p, ok := s.Board.Coord(c); ok && sees(eyes, p) {
			out[p] = true
		}
	}
	return out
}

// hidden reports whether fog hides unit idx, on cell c of side's map, from
// side; a two-cell ship half shows while either half is in sight. eyes are
// side's (eyes).
func (s *Session) hidden(side, idx, c int, eyes []eye) bool {
	u, ok := s.Game.Units[idx]
	if !s.fogged(side) || !ok || sideOf(u.Flags4) == side {
		return false
	}
	if p, ok := s.Board.Coord(c); ok && sees(eyes, p) {
		return false
	}
	other := -1
	switch u.Flags4 & 0xC0 {
	case 0xC0:
		other = idx + 1
	case 0x40:
		other = idx - 1
	}
	if o, ok := s.Game.Units[other]; ok && !o.Destroyed() {
		if p, ok := s.Board.Coord(o.Cell(side)); ok && sees(eyes, p) {
			return false
		}
	}
	return true
}

// hiddenAt reports whether fog hides the unit side's map shows on cell c.
func (s *Session) hiddenAt(side, c int, eyes []eye) bool {
	occ := s.Game.Maps[side][2*c+1]
	return occ <= dosrules.SlotFree && s.hidden(side, occ, c, eyes)
}

// EventsFor are the events of evs side may see under fog, in order; without
// fog all of them. It keeps the side's own events and the ones without a
// place (confirmations, phase changes, energy, the end, captures and trench
// work, which change only buildings and terrain); of the others it keeps a
// battle with one of side's units or on a hex in sight, and a move, a
// reveal, an order, a skipped order, a repair or a build whose hex (a
// move's goal) is in sight now.
func (s *Session) EventsFor(side int, evs []Event) []Event {
	if !s.fogged(side) {
		return evs
	}
	eyes := s.eyes(side)
	out := make([]Event, 0, len(evs))
	for _, ev := range evs {
		if s.shown(side, ev, eyes) {
			out = append(out, ev)
		}
	}
	return out
}

// shown reports whether side sees ev under fog.
func (s *Session) shown(side int, ev Event, eyes []eye) bool {
	switch ev.Kind {
	case Battle:
		if ev.Battle == nil {
			return ev.Side == side
		}
		return ev.Battle.Attacker.Unit.Side == side || ev.Battle.Defender.Unit.Side == side ||
			sees(eyes, ev.From) || sees(eyes, ev.To)
	case Moved, Refused, Revealed:
		if ev.Side == side {
			return true
		}
		goal := ev.To
		if len(ev.Path) > 0 {
			goal = ev.Path[len(ev.Path)-1]
		}
		return sees(eyes, goal)
	case Ordered, Tasked:
		return ev.Side == side || sees(eyes, ev.From)
	case Skipped, Repaired, Built:
		return ev.Side == side || sees(eyes, ev.To)
	}
	return true
}
