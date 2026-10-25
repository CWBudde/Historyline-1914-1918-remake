// Package origview turns an imported original map (finmap.Level) into a
// level.Level the front end can draw: every terrain record gets a look, the
// remake's terrain and infrastructure named in looks.json. The view decides
// only what is drawn; the rules bind the records themselves.
//
// Lines come from cells through finmap.Coord, so the level has 2*Height-1
// lines of Width/2 hexes (odd lines one fewer) and leaves out the last column
// and the odd cells of the last row. The map edge cells it keeps (record
// Edge, the even cells of the last row) take the look of a neighbour inside
// the map. A winter map shows each look's winter terrain (Look.Ground).
// Building entrances carry the owner as content; Level leaves units out (no
// Unit, no content units), the session draws them, and Campaign adds them.
//
// Open water keeps the original's depths: shallow water and ponds are water,
// the middle band waterDeep and the open sea waterSea.
//
// Several looks stand in for original terrain the remake has no art for:
// ponds show open water, town streets a road, quays and piers a road bridge on
// water, forest and mountain variants plain forest and
// mountain, junctions and level crossings without a tile of their own two overlays.
// Airstrips are a road carrying runway36 and their hangars a house carrying
// hangar, which the map draws as an airfield. The harbour buildings are houses
// carrying harbourN (the warehouse) and harbourW (the customs house), their yard
// a road carrying harbourS, which the map draws as a harbour.
package origview

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// Records is the number of terrain records; Edge is the map edge's.
const (
	Records = 175
	Edge    = 174
)

// Look is how a terrain record is drawn: a terrain of terrain.json and the
// infrastructure overlays in drawing order.
type Look struct {
	Terrain string `json:"terrain"`
	// Winter is the winter terrain where it is not Terrain+"_w" (or Terrain
	// when terrain.json has no winter form).
	Winter string   `json:"winter,omitempty"`
	Infra  []string `json:"infra,omitempty"`
}

//go:embed looks.json
var looksJSON []byte

// Looks are the looks of the terrain records (looks.json).
var Looks = mustLooks(looksJSON)

// terrains are the terrain names of terrain.json.
var terrains = mustTerrains()

func mustLooks(b []byte) map[int]Look {
	var f struct {
		Looks map[int]Look `json:"looks"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		panic("origview: " + err.Error())
	}
	return f.Looks
}

func mustTerrains() map[string]bool {
	set, err := data.Load()
	if err != nil {
		panic("origview: " + err.Error())
	}
	w := map[string]bool{}
	for name := range set.Terrain {
		w[name] = true
	}
	return w
}

// Ground is the terrain the look shows in summer or winter: in winter Winter,
// else Terrain+"_w" where terrain.json has it, else Terrain.
func (k Look) Ground(winter bool) string {
	switch {
	case !winter:
		return k.Terrain
	case k.Winter != "":
		return k.Winter
	case terrains[k.Terrain+"_w"]:
		return k.Terrain + "_w"
	}
	return k.Terrain
}

// Hex is the level hex of the look, without unit and content.
func (k Look) Hex(winter bool) level.Hex {
	h := level.Hex{Terrain: k.Ground(winter)}
	if len(k.Infra) > 0 {
		h.Infra = level.Infra{IDs: append([]string(nil), k.Infra...), List: len(k.Infra) > 1}
	}
	return h
}

// owners are the level owners of sides 0, 1 and 2 (neutral).
var owners = [3]game.Player{game.CentralPowers, game.Entente, game.None}

// Level is the view of an imported map. MapName is the map's file stem
// ("05", "STATMAP"), MapNo its number plus one (-1 when the stem is none),
// as the legacy levels count (the Python PULSE is map 00), Player the first human
// side (Central Powers without a mission). An unknown terrain record,
// building owner or human side is an error.
func Level(l *finmap.Level) (*level.Level, error) {
	g := hexgrid.Grid{Width: l.Width / 2, Lines: 2*l.Height - 1}
	if l.Width < 4 || l.Width%2 != 0 || l.Height < 2 || len(l.Terrain) != l.Height {
		return nil, fmt.Errorf("origview: a %d×%d map", l.Width, l.Height)
	}
	record := func(c hexgrid.Coord) int { return l.Terrain[c.Line/2][2*c.Col+c.Line%2] }
	for r, row := range l.Terrain {
		if len(row) != l.Width {
			return nil, fmt.Errorf("origview: row %d has %d cells", r, len(row))
		}
		for c, t := range row {
			if _, ok := Looks[t]; !ok {
				return nil, fmt.Errorf("origview: cell %d,%d: no look for terrain record %d", r, c, t)
			}
		}
	}
	winter := l.Mission != nil && l.Mission.Winter
	lv := &level.Level{MapName: l.Source.Map, MapNo: -1, Player: game.CentralPowers.String(), Lines: make([][]level.Hex, g.Lines)}
	if n, err := strconv.Atoi(l.Source.Map); err == nil {
		lv.MapNo = n + 1
	}
	if m := l.Mission; m != nil && len(m.HumanSides) > 0 {
		s := m.HumanSides[0]
		if s != finmap.Central && s != finmap.Allied {
			return nil, fmt.Errorf("origview: human side %d", s)
		}
		lv.Player = owners[s].String()
	}
	for line := range lv.Lines {
		lv.Lines[line] = make([]level.Hex, g.Width-line%2)
		for col := range lv.Lines[line] {
			c := hexgrid.Coord{Line: line, Col: col}
			t := record(c)
			if t == Edge {
				t = inside(g, c, record)
			}
			lv.Lines[line][col] = Looks[t].Hex(winter)
		}
	}
	for _, b := range l.Buildings {
		c := finmap.Coord(b.Cell[0], b.Cell[1])
		if !g.Contains(c) || b.Owner < 0 || b.Owner >= len(owners) {
			return nil, fmt.Errorf("origview: %s at %v owned by %d", b.Kind, b.Cell, b.Owner)
		}
		lv.Lines[c.Line][c.Col].Content = &level.Content{Owner: int(owners[b.Owner])}
	}
	return lv, nil
}

// edgeOrder is the order in which an edge hex looks for a neighbour inside
// the map: N, NW, NE, then the others (hexgrid.Around indices).
var edgeOrder = [6]int{5, 4, 0, 1, 3, 2}

// inside is the record of the first neighbour of edge hex c that is neither
// the edge nor a building piece (a copy would be a stray piece); Edge, which
// looks like grass, when there is none.
func inside(g hexgrid.Grid, c hexgrid.Coord, record func(hexgrid.Coord) int) int {
	around := hexgrid.Around(c)
	for _, i := range edgeOrder {
		if n := around[i]; g.Contains(n) && record(n) != Edge && !building(Looks[record(n)].Terrain) {
			return record(n)
		}
	}
	return Edge
}

// building reports whether terrain is a piece of an HQ, depot or factory.
func building(terrain string) bool {
	return strings.HasPrefix(terrain, "hq") || strings.HasPrefix(terrain, "cmp") || strings.HasPrefix(terrain, "factory")
}
