package mapcompose

import (
	"slices"
	"strings"
)

// A harbour is three hexes, named like compound pieces: the quayside
// warehouse (infrastructure harbourN), the customs house at its SW (harbourW)
// and the quay yard at its S (harbourS), which the levels keep as two houses
// and a grass hex with a road. A harbour hex draws the open ground of its
// season, the granite paving of its piece (tile/harbour*) and its building or
// cargo (standing/harbour*) with its cast shadow; the yard keeps its road
// beneath the paving, so the road runs into the yard. On a winter hex every
// piece is drawn in its _w form where the catalog has one. The rules still see
// the houses and the road.

// harbourPiece returns the harbour piece h holds and the catalog can draw.
func harbourPiece(h Hex, has func(string) bool) (string, bool) {
	for _, id := range h.Infra {
		if strings.HasPrefix(id, "harbour") && has(TileID(id)) && has(StandingID(id)) {
			return id, true
		}
	}
	return "", false
}

// harbour returns infra with the harbour piece of h in the season of h.
func harbour(infra []string, h Hex, has func(string) bool) []string {
	id, ok := harbourPiece(h, has)
	if !ok {
		return infra
	}
	i := slices.Index(infra, id)
	return slices.Replace(slices.Clone(infra), i, i+1, seasonal(id, h.Terrain, has))
}

// harbourGround is the background of hex (line, col) holding h: the open
// ground of its season on a harbour hex, otherwise bg.
func harbourGround(h Hex, bg string, line, col int, has func(string) bool) string {
	if _, ok := harbourPiece(h, has); !ok {
		return bg
	}
	if strings.HasSuffix(h.Terrain, "_w") {
		return baseTile("grass_w", line, col, has)
	}
	return baseTile("grass", line, col, has)
}
