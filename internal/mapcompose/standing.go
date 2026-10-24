package mapcompose

import (
	"fmt"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

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

// ShadowID identifies a standing item's separately painted cast shadow. Both
// live in the standing pack and share its overhang limits.
func ShadowID(standingID string) string {
	return "standing/shadow/" + strings.TrimPrefix(standingID, "standing/")
}

// ShadowsAt returns the cast shadows paired with the selected standing items.
func (s *Scene) ShadowsAt(c hexgrid.Coord, stamped func(*Cluster) bool) []string {
	if len(s.Cells[c.Line][c.Col].Shadows) == 0 {
		return nil
	}
	_, _, _, standing := s.Layers(c, stamped)
	var out []string
	for _, id := range standing {
		if shadow := ShadowID(id); slices.Contains(s.Cells[c.Line][c.Col].Shadows, shadow) {
			out = append(out, shadow)
		}
	}
	return out
}

// TreePropID selects sparse meadow props without consuming random state. The
// coordinate hash selects about one in seven hexes and one of four authored
// layouts; the layouts keep the hex centre clear for units. An empty ID means
// no prop. Selection is independent of traversal order and visible bounds.
func TreePropID(line, col int) string {
	x := coordinateHash(line, col)
	if x%7 != 0 {
		return ""
	}
	return fmt.Sprintf("standing/props/trees/%d", (x/7)%4+1)
}

// coordinateHash mixes both coordinates with SplitMix64 without random state.
func coordinateHash(line, col int) uint64 {
	x := uint64(uint32(line))<<32 | uint64(uint32(col))
	x += 0x9e3779b97f4a7c15
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	x ^= x >> 31
	return x
}

// A prop can overhang its own meadow. Keep it away from buildings, water and
// infrastructure on neighbouring cells; forest edges may retain scattered trees.
func treePropNeighboursClear(g hexgrid.Grid, at func(hexgrid.Coord) Hex, c hexgrid.Coord) bool {
	for _, n := range hexgrid.Around(c) {
		if !g.Contains(n) {
			continue
		}
		h := at(n)
		if h.Infra != nil || (h.Terrain != "grass" && h.Terrain != "forest") {
			return false
		}
	}
	return true
}

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
// the catalog has. A road bridge draws the deck of what crosses it (bridgeDeck).
// The terrain carries its season; infrastructure does not, so on a winter map
// (mostlyWinter) its items take their winter form (suffix _w) where the catalog
// has it, as the overlays do (seasonalOverlay).
func standing(terrain string, infra []string, winter bool, has func(string) bool) []string {
	var out []string
	if sid := StandingID(terrain); has(sid) {
		out = append(out, sid)
	}
	for _, id := range infra {
		sid := StandingID(id)
		if deck := bridgeDeck(id, infra, has); deck != "" {
			sid = deck
		}
		if !has(sid) {
			continue
		}
		if w := sid + "_w"; winter && !strings.HasSuffix(sid, "_w") && has(w) {
			sid = w
		}
		out = append(out, sid)
	}
	return out
}

// bridgeDeck returns the standing deck of road bridge id chosen by the crossing
// family, or "" for the road deck. The levels write every bridge as rBridgeNN; a
// bridge without a road carries rail (the rail deck standing/rBridgeNN-rail) or
// a path (the footbridge standing/bridgeNN), where the catalog has them.
func bridgeDeck(id string, infra []string, has func(string) bool) string {
	ports, ok := strings.CutPrefix(id, "rBridge")
	if !ok || slices.ContainsFunc(infra, func(i string) bool { return strings.HasPrefix(i, "road") }) {
		return ""
	}
	switch {
	case slices.Contains(infra, "rail"+ports) && has(StandingID(id+"-rail")):
		return StandingID(id + "-rail")
	case slices.Contains(infra, "path"+ports) && has(StandingID("bridge"+ports)):
		return StandingID("bridge" + ports)
	}
	return ""
}
