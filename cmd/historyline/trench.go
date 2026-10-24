package main

import (
	"log"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// Sappers dig and fill trenches during a battle (-rules dos). The session's
// level stays as loaded; the map draws a copy with the hexes the work
// changed (dosplay.Session.Reworked), composed again when a work plays back.

// reworkTerrain draws the hexes the sappers changed, and their terrain
// records in the original tiles (F8).
func (g *game) reworkTerrain() {
	s := g.dos.s
	g.lvl = reworked(s.Level, s.Reworked())
	g.records = make([]int, len(s.Board.Terrain))
	for c := range g.records {
		g.records[c] = s.Game.Maps[0][2*c]
	}
	if g.terrain == nil {
		return
	}
	if err := g.terrain.recompose(g.lvl); err != nil {
		log.Printf("terrain after a sapper's work: %v", err)
	}
}

// reworked is a copy of base with hexes in place of its own.
func reworked(base *level.Level, hexes map[hexgrid.Coord]level.Hex) *level.Level {
	l := *base
	l.Lines = make([][]level.Hex, len(base.Lines))
	for i, line := range base.Lines {
		l.Lines[i] = slices.Clone(line)
	}
	for c, h := range hexes {
		l.Lines[c.Line][c.Col] = h
	}
	return &l
}
