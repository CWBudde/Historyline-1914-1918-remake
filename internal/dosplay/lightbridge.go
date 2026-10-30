package dosplay

import (
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Light bridges (data.LightBridge: a path or plain bridge over water, as
// NONNE's southern crossing) are a remake rule the original records do
// not carry: their terrain record admits every ground unit. The session
// therefore shows a unit too heavy for such a bridge the water under it
// instead: the movement queries of a heavy unit see the bare stream
// record on the bridge cells, and a move or exit order runs with the
// maps swapped the same way, so the order code refuses a crossing it
// would otherwise allow. dosrules itself stays the original's.

// lightCells are the board cells of the level's light bridges, and
// lightWater is the terrain record of the water they cross (a stream or
// lake hex without a bridge), -1 where the level has none.
func (s *Session) lightCells() ([]int, int) {
	if s.lightDone {
		return s.light, s.lightWater
	}
	s.lightDone, s.lightWater = true, -1
	for l, line := range s.Level.Lines {
		for c, h := range line {
			cell, ok := s.Board.Cell(hexgrid.Coord{Line: l, Col: c})
			if !ok || s.Board.Edge(cell) {
				continue
			}
			switch {
			case data.LightBridge(h.Infra.IDs):
				s.light = append(s.light, cell)
			case s.lightWater < 0 && openWater(h.Infra.IDs):
				s.lightWater = s.Board.Terrain[cell]
			}
		}
	}
	return s.light, s.lightWater
}

// openWater reports whether infra is a stream or lake without a bridge.
func openWater(infra []string) bool {
	water, bridge := false, false
	for _, id := range infra {
		switch {
		case strings.HasPrefix(id, "stream"), strings.HasPrefix(id, "lake"):
			water = true
		case strings.HasPrefix(id, "bridge"), strings.HasPrefix(id, "rBridge"):
			bridge = true
		}
	}
	return water && !bridge
}

// crossesLight reports whether unit may cross a light bridge: aircraft
// fly over it, ships sail under it on the water record, and ground units
// follow data.Unit.CrossesLightBridge. A unit the remake has no data for
// stays off.
func (s *Session) crossesLight(unit int) bool {
	u, ok := s.Game.Units[unit]
	if !ok {
		return false
	}
	if u.Flags4&0x10 != 0 {
		return true // air: the query reads the air cost
	}
	if d, ok := s.set.Units[s.unitID(unit)]; ok {
		return d.CrossesLightBridge()
	}
	return false
}

// patchLight turns the light bridge cells of the movement query m into
// the water under them for a unit too heavy to cross.
func (s *Session) patchLight(m *dosrules.Movement, unit int) {
	cells, water := s.lightCells()
	if len(cells) == 0 || water < 0 || s.crossesLight(unit) {
		return
	}
	for _, c := range cells {
		m.Terrain[c] = water
	}
}

// movement is the side's movement query of unit with patchLight applied.
func (s *Session) movement(side, unit int) *dosrules.Movement {
	m := s.Game.Movement(side, unit)
	s.patchLight(m, unit)
	return m
}

// blockLight swaps the light bridge cells of both maps to the water
// record while an order of unit runs, so the order code routes and
// refuses like the queries do; the returned function restores the maps.
func (s *Session) blockLight(unit int) func() {
	cells, water := s.lightCells()
	if len(cells) == 0 || water < 0 || s.crossesLight(unit) {
		return func() {}
	}
	type swap struct {
		side, cell, old int
	}
	var old []swap
	for k := range s.Game.Maps {
		for _, c := range cells {
			old = append(old, swap{k, c, s.Game.Maps[k][2*c]})
			s.Game.Maps[k][2*c] = water
		}
	}
	return func() {
		for _, sw := range old {
			s.Game.Maps[sw.side][2*sw.cell] = sw.old
		}
	}
}
