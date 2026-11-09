package dosplay

import (
	"errors"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Gas pioneers release a lasting gas cloud (dosrules gas.go), an extension
// type that only a game on the extended rules fields (docs/new-units.md).

// GasPioneer reports whether unit releases gas instead of attacking: its
// action order goes to a neighbouring hex, where it releases a cloud over
// three hexes after the battles, and fights no battle.
func (s *Session) GasPioneer(unit int) bool { return s.Game.Units[unit].GasPioneer() }

// Charges is what a construction unit or a gas pioneer has left to work
// with, -1 for a unit without charges.
func (s *Session) Charges(unit int) int {
	u, ok := s.Game.Units[unit]
	if !ok || !u.Builder() && !u.GasPioneer() {
		return -1
	}
	return int(int8(byte(u.Hold)))
}

// CanRelease is why gas pioneer unit of side can release no gas now (none
// left), nil when it can.
func (s *Session) CanRelease(side, unit int) error {
	u, ok := s.Game.Units[unit]
	if !ok || !u.GasPioneer() {
		return refusal("not a gas pioneer")
	}
	err := s.Game.Check(dosrules.Command{Kind: dosrules.CmdAttack, Side: side, Unit: unit, Cell: u.Cell(side)})
	var r *dosrules.Refusal
	if errors.As(err, &r) && r.Code == dosrules.MsgNoTarget {
		return nil // a neighbour: its own cell is none
	}
	return wrap(err)
}

// Cloud is the hexes a cloud released by gas pioneer unit of side on hex to
// would cover, without ordering it; nil for a unit that is no gas pioneer
// or a hex that is no neighbour of it.
func (s *Session) Cloud(side, unit int, to hexgrid.Coord) []hexgrid.Coord {
	cell, ok := s.Board.Cell(to)
	if !ok {
		return nil
	}
	return s.levelHexes(s.Game.GasCells(side, unit, cell))
}

// Gas are the gassed hexes and the turns each has left. Both sides see the
// cloud.
func (s *Session) Gas() map[hexgrid.Coord]int {
	out := map[hexgrid.Coord]int{}
	for c, t := range s.Game.Gas {
		if p, ok := s.Board.Coord(c); ok && t > 0 {
			out[p] = t
		}
	}
	return out
}

// GasAt is the turns the cloud on hex has left, 0 for none.
func (s *Session) GasAt(at hexgrid.Coord) int {
	cell, ok := s.Board.Cell(at)
	if !ok {
		return 0
	}
	return s.Game.GasAt(cell)
}

// levelHexes are the level hexes of cells, leaving out cells off the level.
func (s *Session) levelHexes(cells []int) []hexgrid.Coord {
	var out []hexgrid.Coord
	for _, c := range cells {
		if p, ok := s.Board.Coord(c); ok {
			out = append(out, p)
		}
	}
	return out
}

// GasRates are the cloud's duration in turns and a foot unit's loss per
// turn in percent by the level's month ("1915-04"; "" for the defaults):
// gas masks halve the loss from 1916, mustard gas doubles the duration
// from July 1917.
func GasRates(month string) (turns, loss int) {
	turns, loss = dosrules.GasTurns, dosrules.GasLoss
	if month >= "1916-01" {
		loss = 15
	}
	if month >= "1917-07" {
		turns = 4
	}
	return turns, loss
}

// extensionTypes are the extension types of the remake's units, ascending.
func extensionTypes(set *data.Set) []int {
	var out []int
	for _, u := range set.Units {
		if u.Extension() && !slices.Contains(out, u.Dos) {
			out = append(out, u.Dos)
		}
	}
	slices.Sort(out)
	return out
}

// usesExtension reports whether a board places or builds an extension type.
func usesExtension(b *dosbind.Board) bool {
	for _, u := range b.Units {
		if u.Type >= data.OriginalTypes {
			return true
		}
	}
	for _, bd := range b.Buildings {
		for _, t := range slices.Concat(bd.Builds, bd.ComputerBuilds) {
			if t >= data.OriginalTypes {
				return true
			}
		}
	}
	return false
}

// extend sets a new game on the extended rules up: every factory also
// builds the extension types, and the cloud follows the month.
func extend(g *dosrules.Game, set *data.Set, month string, builds bool) {
	if builds {
		ext := extensionTypes(set)
		for i, bd := range g.Buildings {
			if dosrules.BuildingKind(i) != dosrules.KindFactory {
				continue
			}
			for _, t := range ext {
				if !slices.Contains(bd.Builds, t) {
					bd.Builds = append(bd.Builds, t)
				}
			}
			g.Buildings[i] = bd
		}
	}
	turns, loss := GasRates(month)
	if turns != dosrules.GasTurns {
		g.GasTurns = turns
	}
	if loss != dosrules.GasLoss {
		g.GasLoss = loss
	}
}
