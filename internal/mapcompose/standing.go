package mapcompose

import "github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"

// The standing layer holds what rises above the ground of a hex: tree crowns,
// rocks, crests, buildings, bridges. Its art may overhang the hex by up to half a
// hex on every side, so it is drawn after the ground and infrastructure of all
// hexes, in line then column order, and before units.

// OverhangX and OverhangY are the most a standing item may reach beyond its hex,
// in logical pixels on each side.
const OverhangX, OverhangY = hexgrid.HexWidth / 2, hexgrid.HexHeight / 2

// StandingID is the runtime asset ID of the standing part of a terrain or
// infrastructure tile.
func StandingID(runtimeID string) string { return "standing/" + runtimeID }

// StandingFits reports whether a standing canvas of w×h logical pixels with its
// anchor on the hex centre stays within the allowed overhang.
func StandingFits(w, h float64, anchor [2]float64) bool {
	const hx, hy = hexgrid.HexWidth/2 + OverhangX, hexgrid.HexHeight/2 + OverhangY
	return anchor[0] >= 0 && anchor[1] >= 0 && anchor[0] <= hx && anchor[1] <= hy && w-anchor[0] <= hx && h-anchor[1] <= hy
}

// Standing is one standing item in drawing order.
type Standing struct {
	At hexgrid.Coord
	ID string
}

// Standing returns the standing items of the scene in drawing order: by line,
// then column, then the cell's own order.
func (s *Scene) Standing() []Standing {
	var out []Standing
	for l, line := range s.Cells {
		for col, c := range line {
			for _, id := range c.Standing {
				out = append(out, Standing{At: hexgrid.Coord{Line: l, Col: col}, ID: id})
			}
		}
	}
	return out
}

// standing returns the standing IDs of a hex showing terrain with infra, for those
// the catalog has.
func standing(terrain string, infra []string, has func(string) bool) []string {
	var out []string
	for _, id := range append([]string{terrain}, infra...) {
		if sid := StandingID(id); has(sid) {
			out = append(out, sid)
		}
	}
	return out
}
