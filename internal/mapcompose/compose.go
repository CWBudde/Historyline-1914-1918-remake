package mapcompose

import (
	"fmt"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Hex is the terrain of one map hex: its background ID and its infrastructure
// IDs in stored (drawing) order.
type Hex struct {
	Terrain string
	Infra   []string
}

// Cell is what the map view draws on one hex: a background tile, then the
// infrastructure overlays in order.
type Cell struct {
	Background string   // asset ID
	Overlays   []string // asset IDs
	Context    Family   // set when Background is a context tile
}

// Reason says why a context tile could not be selected.
type Reason string

const (
	// Missing: the contract defines the context, but it has not been exported.
	Missing Reason = "missing"
	// Undefined: an unknown terrain or a neighbour from another season.
	// The contract defines no tile for it.
	Undefined Reason = "undefined"
)

// Gap is a hex drawn with its default tile because its context tile is not
// available. The default tile is not a valid replacement; gaps are listed so the
// art track can export or define them.
type Gap struct {
	At         hexgrid.Coord
	Terrain    string
	Family     Family
	Reason     Reason
	Descriptor string // the missing context's descriptor (Missing only)
	Detail     string // the offending neighbours (Undefined only)
}

// Issue is a compound-building piece outside a complete group. It is drawn as
// is; the compositor never invents the missing pieces.
type Issue struct {
	At     hexgrid.Coord
	Detail string
}

// Scene is a composed map: Cells[line][col].
type Scene struct {
	Grid   hexgrid.Grid
	Cells  [][]Cell
	Gaps   []Gap
	Issues []Issue
}

// Compose selects the tiles of every hex of g. at returns a hex's terrain; has
// reports whether a context asset ID is available.
func Compose(g hexgrid.Grid, at func(hexgrid.Coord) Hex, has func(id string) bool) Scene {
	s := Scene{Grid: g, Cells: make([][]Cell, g.Lines)}
	for l := range g.Lines {
		width := g.Width
		if l&1 == 1 {
			width--
		}
		s.Cells[l] = make([]Cell, width)
		for col := range width {
			c := hexgrid.Coord{Line: l, Col: col}
			h := at(c)
			cell := &s.Cells[l][col]
			cell.Background = TileID(h.Terrain)
			for _, id := range h.Infra {
				cell.Overlays = append(cell.Overlays, TileID(id))
			}
			if f, ok := FamilyOf(h.Terrain); ok {
				if id, gap := selectContext(g, at, c, f, h.Terrain, has); gap != nil {
					s.Gaps = append(s.Gaps, *gap)
				} else {
					cell.Background, cell.Context = id, f
				}
			}
		}
	}
	s.registerCompounds(at)
	return s
}

func selectContext(g hexgrid.Grid, at func(hexgrid.Coord) Hex, c hexgrid.Coord, f Family, centre string, has func(string) bool) (string, *Gap) {
	var neighbors [6]string
	var bad []string
	season, base, _ := member(f, centre)
	for i, n := range hexgrid.Around(c) {
		if !g.Contains(n) {
			// Continue the material beyond the visible map boundary.
			neighbors[i] = centre
			continue
		}
		neighbors[i] = at(n).Terrain
		if s, _, ok := member(f, neighbors[i]); !ok || s != season {
			if boundaryTerrain(neighbors[i], season) {
				neighbors[i] = base
			} else {
				bad = append(bad, directions[i]+"="+neighbors[i])
			}
		}
	}
	gap := &Gap{At: c, Terrain: centre, Family: f}
	if len(bad) > 0 {
		gap.Reason, gap.Detail = Undefined, strings.Join(bad, ", ")
		return "", gap
	}
	x, y := Phase(c.Line, c.Col)
	d, _ := Descriptor(f, centre, neighbors, [2]int{x, y})
	if id := ContextID(f, Key(d)); has(id) {
		return id, nil
	}
	gap.Reason, gap.Descriptor = Missing, d
	return "", gap
}

// A compound building covers several hexes, each with its own piece tile. Steps
// walk from the group's N piece to each piece through hexgrid.Around directions.
type compound struct {
	prefix string
	pieces map[string][]int // piece suffix → Around directions from N
}

const (
	dirSE, dirS, dirSW = 1, 2, 3
)

var diamond = map[string][]int{"N": nil, "W": {dirSW}, "E": {dirSE}, "S": {dirS}}

var compounds = []compound{
	{"hq", map[string][]int{
		"N": nil, "NW": {dirSW}, "NE": {dirSE}, "C": {dirS},
		"SW": {dirS, dirSW}, "SE": {dirS, dirSE}, "S": {dirS, dirS},
	}},
	{"cmp", diamond},
	{"factory", diamond},
	{"mount", diamond},
}

// piece splits a compound terrain ID into its compound, piece and season suffix.
func piece(id string) (cp *compound, name, suffix string) {
	base, winter := strings.CutSuffix(id, "_w")
	if winter {
		suffix = "_w"
	}
	for i := range compounds {
		p, ok := strings.CutPrefix(base, compounds[i].prefix)
		if _, known := compounds[i].pieces[p]; ok && known {
			return &compounds[i], p, suffix
		}
	}
	return nil, "", ""
}

// registerCompounds checks that every compound piece belongs to a complete group
// of the same season. A mountain quarter outside a group is shown as a whole
// mountain (display only); any other stray piece is an Issue.
func (s *Scene) registerCompounds(at func(hexgrid.Coord) Hex) {
	g := s.Grid
	claimed := map[hexgrid.Coord]bool{}
	for l := range s.Cells {
		for col := range s.Cells[l] {
			c := hexgrid.Coord{Line: l, Col: col}
			cp, name, sfx := piece(at(c).Terrain)
			if cp == nil || name != "N" {
				continue
			}
			group := map[hexgrid.Coord]bool{c: true}
			for p, steps := range cp.pieces {
				x := c
				for _, d := range steps {
					x = hexgrid.Around(x)[d]
				}
				if p == "N" {
					continue
				}
				if !g.Contains(x) || at(x).Terrain != cp.prefix+p+sfx {
					group = nil
					break
				}
				group[x] = true
			}
			for x := range group {
				claimed[x] = true
			}
		}
	}
	for l := range s.Cells {
		for col := range s.Cells[l] {
			c := hexgrid.Coord{Line: l, Col: col}
			t := at(c).Terrain
			cp, _, sfx := piece(t)
			if cp == nil || claimed[c] {
				continue
			}
			if cp.prefix == "mount" {
				s.Cells[l][col].Background = TileID("mountain" + sfx)
				continue
			}
			s.Issues = append(s.Issues, Issue{At: c, Detail: fmt.Sprintf("%s outside a complete %s group", t, cp.prefix)})
		}
	}
}
