package dosrules

import (
	"maps"
	"slices"
)

// A field ambulance (live +6 bit 0x0080, an extension type) treats the
// wounded. Once per game turn, with the gas (turnEffects), every foot unit
// (live +6 bit 0x01) next to an ambulance of its side on the side's map
// regains AmbulanceHeal strength, up to its full group; a unit next to two
// ambulances regains it once. Neither the ambulance nor the unit may be held
// in a building or a transporter: a depot repairs there instead. The
// ambulance itself never fights (it has no weapon and never returns fire).

const (
	flagMedic = 0x0080 // live +6: a field ambulance (extension types only)

	// AmbulanceHeal is the strength a unit next to an ambulance regains per
	// turn.
	AmbulanceHeal = 1
)

// Ambulance reports whether the unit is a field ambulance.
func (u Live) Ambulance() bool { return u.Flags6&flagMedic != 0 }

// healTurn is the ambulances' turn, in unit order of the ambulances.
func (g *Game) healTurn() []Event {
	var evs []Event
	healed := map[int]bool{}
	for _, i := range slices.Sorted(maps.Keys(g.Units)) {
		a := g.Units[i]
		side := ownerOf(a)
		if !a.Ambulance() || a.Destroyed() || a.Action[1]&4 != 0 || side > 1 || a.Flags4&0x4000 != 0 {
			continue
		}
		for _, c := range neighbours(a.Cell(side), g.Width, g.Height) {
			if c < 0 {
				continue
			}
			j := g.Maps[side][2*c+1]
			u, ok := g.Units[j]
			if !ok || healed[j] || u.Destroyed() || u.Action[1]&4 != 0 || ownerOf(u) != side || u.Flags6&0x01 == 0 ||
				u.Flags4&0x4000 != 0 {
				continue
			}
			full := g.Profile.Tables.Types[u.Type].Group
			gain := min(AmbulanceHeal, full-u.Size)
			if gain <= 0 {
				continue
			}
			u.Size += gain
			g.Units[j] = u
			healed[j] = true
			evs = append(evs, Event{Kind: EvHealed, Side: side, Unit: j, Cell: c, Gain: gain})
		}
	}
	return evs
}

// turnEffects are the extension units' effects once per game turn, in the
// resolution whose phase change counts the turn (side 0's action phase): the
// tunnels (mineTurn), the gas (gasTurn), then the ambulances (healTurn). The
// units they destroy leave the map before the phase change marks them.
func (g *Game) turnEffects() []Event {
	if g.Players[0].Phase != PhaseAction {
		return nil
	}
	var gone []int
	evs := g.mineTurn(&gone)
	evs = append(evs, g.gasTurn(&gone)...)
	g.vacate(gone)
	return append(evs, g.healTurn()...)
}
