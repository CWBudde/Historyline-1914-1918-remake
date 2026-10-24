package mapcompose

import (
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// An airfield is a runway of north-south hexes, each a road carrying the
// infrastructure runway36, beside hangars: hexes whose infrastructure is
// hangar. The runway is drawn instead of its road (the rules still see the
// road). A hangar hex draws the open ground of its season (grass or grass_w),
// the apron in front of the hangar's doors (tile/hangar) and the hangar
// itself (standing/hangar); the apron runs on over the runway hexes at its NE
// and SE (tile/apron-sw, apron-nw: the hangar lies SW or NW of them). On a
// winter hex every piece is drawn in its _w form where the catalog has one.

const (
	runway = "runway36"
	hangar = "hangar"
)

// seasonal is id in the season of terrain: id+"_w" on a winter hex where the
// catalog has that tile.
func seasonal(id, terrain string, has func(string) bool) string {
	if strings.HasSuffix(terrain, "_w") && has(TileID(id+"_w")) {
		return id + "_w"
	}
	return id
}

// hasHangar reports whether h holds a hangar the catalog can draw.
func hasHangar(h Hex, has func(string) bool) bool {
	return slices.Contains(h.Infra, hangar) && has(TileID(hangar)) && has(StandingID(hangar))
}

// airfield returns the infrastructure drawn on hex c: on a runway the runway
// in place of its road, then the aprons of the hangars at its SW and NW; on a
// hangar hex the hangar in its season.
func airfield(g hexgrid.Grid, at func(hexgrid.Coord) Hex, c hexgrid.Coord, h Hex, has func(string) bool) []string {
	infra := h.Infra
	if hasHangar(h, has) {
		i := slices.Index(infra, hangar)
		return slices.Replace(slices.Clone(infra), i, i+1, seasonal(hangar, h.Terrain, has))
	}
	if !slices.Contains(infra, runway) || !has(TileID(runway)) {
		return infra
	}
	out := slices.DeleteFunc(slices.Clone(infra), func(id string) bool { return id == "road36" || id == runway })
	out = append(out, seasonal(runway, h.Terrain, has))
	around := hexgrid.Around(c)
	for _, n := range []struct {
		at    hexgrid.Coord
		apron string
	}{{around[3], "apron-sw"}, {around[4], "apron-nw"}} {
		if g.Contains(n.at) && hasHangar(at(n.at), has) && has(TileID(n.apron)) {
			out = append(out, seasonal(n.apron, h.Terrain, has))
		}
	}
	return out
}

// hangarGround is the background of a hangar hex: the open ground of its
// season, or ok false when h holds no hangar the catalog can draw.
func hangarGround(h Hex, line, col int, has func(string) bool) (string, bool) {
	if !hasHangar(h, has) {
		return "", false
	}
	if strings.HasSuffix(h.Terrain, "_w") {
		return baseTile("grass_w", line, col, has), true
	}
	return baseTile("grass", line, col, has), true
}
