package mapcompose

import (
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// LevelHex reads the hexes of l for Compose: terrain, infrastructure, and
// Vacant where the level gives a hex no building content.
func LevelHex(l *level.Level) func(hexgrid.Coord) Hex {
	return func(c hexgrid.Coord) Hex {
		h := l.Lines[c.Line][c.Col]
		return Hex{Terrain: h.Terrain, Infra: h.Infra.IDs, Vacant: h.Content == nil}
	}
}
