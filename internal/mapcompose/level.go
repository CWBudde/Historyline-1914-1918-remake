package mapcompose

import (
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// LevelHex reads the hexes of l for Compose: terrain, infrastructure, and
// Vacant where the level gives a hex no building content. A hex the
// scenario draws in another season (level.Extra.Seasons) takes its terrain
// in that season (data.InSeason: "grass_w", "rough_w" for hills) and that
// Season.
func LevelHex(l *level.Level) func(hexgrid.Coord) Hex {
	var seasons map[hexgrid.Coord]string
	if l.Extra != nil {
		seasons = l.Extra.Seasons
	}
	return func(c hexgrid.Coord) Hex {
		h := l.Lines[c.Line][c.Col]
		out := Hex{Terrain: h.Terrain, Infra: h.Infra.IDs, Vacant: h.Content == nil}
		if s, ok := seasons[c]; ok {
			out.Terrain, out.Season = data.InSeason(data.Kind(h.Terrain), s), Season(s)
		}
		return out
	}
}
