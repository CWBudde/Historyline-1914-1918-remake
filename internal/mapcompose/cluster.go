package mapcompose

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// A cluster is a connected group of hexes that the art paints as one feature: a
// crater over several mud hexes, a lake over water and its shore. When the
// catalog has a stamp for the cluster's kind and shape, the stamp is drawn
// across the whole cluster on the base material instead of the per-hex art.
//
// A knoll or boulder group whose shape has no stamp is tiled with the smaller
// stamps the catalog has (Tile): each piece is a cluster of its own, and the
// hexes no stamp fits form unstamped clusters that keep their per-hex art.
// Crater fields are not tiled (a crater stamp is larger than a pit, so a field
// of them crowds), nor are lakes (their shore tiles join up across the water).

// Cluster is a connected group of hexes of one kind: a whole group of the map,
// or a piece of a group that the catalog has no stamp for.
type Cluster struct {
	Kind  string          // crater, boulders, knoll or lake; winter kinds end in _w
	Cells []hexgrid.Coord // in line, then column order; the stamp is anchored on Cells[0]
	Shape string          // ShapeKey(Cells)
	// Stamp is the stamp's asset ID, "" when the catalog has none for the shape;
	// Standing is the stamp's standing item, "" when it has none.
	Stamp, Standing string
}

// clusterKind returns the cluster kind and the seasonal base material of a
// terrain, or ok false when the terrain does not form clusters.
func clusterKind(terrain string) (kind, base string, ok bool) {
	switch {
	case terrain == "mud":
		return "crater", "grass", true
	case terrain == "shelled_w":
		return "crater_w", "grass_w", true
	case terrain == "stone":
		return "boulders", "grass", true
	case terrain == "stoney_w":
		return "boulders_w", "grass_w", true
	case terrain == "mountain":
		return "knoll", "grass", true
	case terrain == "mountain_w":
		return "knoll_w", "grass_w", true
	case terrain == "water", terrain == "waterStones", strings.HasPrefix(terrain, "lakeside"):
		return "lake", "grass", true
	}
	return "", "", false
}

// ShapeKey names the shape of cells (in line, then column order) independently of
// its position: the count, then each further cell's offset from the first in
// half-hex columns and lines. A single hex is "1", two hexes one above the other
// "2_0.2", a pair to the south-east "2_1.1".
func ShapeKey(cells []hexgrid.Coord) string {
	x0, y0 := doubled(cells[0])
	parts := []string{fmt.Sprint(len(cells))}
	for _, c := range cells[1:] {
		x, y := doubled(c)
		parts = append(parts, fmt.Sprintf("%d.%d", x-x0, y-y0))
	}
	return strings.Join(parts, "_")
}

func doubled(c hexgrid.Coord) (x, y int) { return 2*c.Col + c.Line&1, c.Line }

// MaxStampVariants bounds the variants of one stamp; variants are numbered
// from 1 without gaps.
const MaxStampVariants = 8

// StampID is the runtime asset ID of variant v (from 1) of a cluster stamp.
func StampID(kind, shape string, v int) string {
	return fmt.Sprintf("stamp/%s/%s/%d", kind, shape, v)
}

// registerClusters finds the clusters of the scene and selects their stamps,
// tiling a group without one (Tile). The variant is chosen from the variants
// the catalog has, seeded by the cluster's first cell.
func (s *Scene) registerClusters(at func(hexgrid.Coord) Hex, has func(string) bool) {
	g := s.Grid
	seen := map[hexgrid.Coord]bool{}
	for l := range s.Cells {
		for col := range s.Cells[l] {
			start := hexgrid.Coord{Line: l, Col: col}
			kind, base, ok := clusterKind(at(start).Terrain)
			if !ok || seen[start] {
				continue
			}
			cl := Cluster{Kind: kind}
			queue := []hexgrid.Coord{start}
			seen[start] = true
			for len(queue) > 0 {
				c := queue[0]
				queue = queue[1:]
				cl.Cells = append(cl.Cells, c)
				for _, n := range g.Neighbors(c) {
					if k, _, ok := clusterKind(at(n).Terrain); ok && k == kind && !seen[n] {
						seen[n] = true
						queue = append(queue, n)
					}
				}
			}
			slices.SortFunc(cl.Cells, byLineCol)
			parts := []Cluster{cl}
			if !stamp(&parts[0], has) && tiled[kind] {
				parts = Tile(g, kind, cl.Cells, has)
			}
			for _, p := range parts {
				s.Clusters = append(s.Clusters, p)
				for _, c := range p.Cells {
					cell := &s.Cells[c.Line][c.Col]
					cell.Cluster = len(s.Clusters)
					if p.Stamp != "" {
						cell.Under = baseTile(base, c.Line, c.Col, has)
					}
				}
			}
		}
	}
}

func byLineCol(a, b hexgrid.Coord) int {
	return cmp.Or(cmp.Compare(a.Line, b.Line), cmp.Compare(a.Col, b.Col))
}

// stamp sets the shape of cl (cells in line, then column order) and, when the
// catalog has a stamp for it, the stamp and its standing item, and reports
// whether it has. The variant is seeded by the first cell.
func stamp(cl *Cluster, has func(string) bool) bool {
	cl.Shape = ShapeKey(cl.Cells)
	variants := 0
	for variants < MaxStampVariants && has(StampID(cl.Kind, cl.Shape, variants+1)) {
		variants++
	}
	if variants == 0 {
		return false
	}
	first := cl.Cells[0]
	variant := int((coordinateHash(first.Line, first.Col)>>32)%uint64(variants)) + 1
	cl.Stamp = StampID(cl.Kind, cl.Shape, variant)
	if id := StandingID(cl.Stamp); has(id) {
		cl.Standing = id
	}
	return true
}

// tiled are the cluster kinds a group without a stamp is tiled for.
var tiled = map[string]bool{"knoll": true, "knoll_w": true, "boulders": true, "boulders_w": true}

// MaxPieceCells bounds the pieces Tile tries.
const MaxPieceCells = 6

// Tile covers a group of kind (cells in line, then column order) with the
// catalog's stamps for smaller shapes. Going through the cells in order, the
// first cell no piece covers yet anchors the largest stamped shape of at most
// MaxPieceCells uncovered cells (the smaller shape key among equals); a cell no
// stamp fits stays uncovered. The pieces come first, in anchor order, then the
// uncovered cells as connected unstamped clusters.
func Tile(g hexgrid.Grid, kind string, cells []hexgrid.Coord, has func(string) bool) []Cluster {
	free := map[hexgrid.Coord]bool{}
	for _, c := range cells {
		free[c] = true
	}
	stamped := map[string]bool{} // shape -> the catalog has a stamp for it
	var pieces []Cluster
	for _, c := range cells {
		if !free[c] {
			continue
		}
		var best []hexgrid.Coord
		bestShape := ""
		for _, sub := range connectedSets(g, c, free, MaxPieceCells) {
			if len(sub) < len(best) {
				continue
			}
			shape := ShapeKey(sub)
			ok, known := stamped[shape]
			if !known {
				ok = has(StampID(kind, shape, 1))
				stamped[shape] = ok
			}
			if ok && (len(sub) > len(best) || shape < bestShape) {
				best, bestShape = sub, shape
			}
		}
		if best == nil {
			continue
		}
		for _, x := range best {
			free[x] = false
		}
		piece := Cluster{Kind: kind, Cells: best}
		stamp(&piece, has)
		pieces = append(pieces, piece)
	}
	seen := map[hexgrid.Coord]bool{}
	for _, c := range cells {
		if !free[c] || seen[c] {
			continue
		}
		rest := Cluster{Kind: kind}
		queue := []hexgrid.Coord{c}
		seen[c] = true
		for len(queue) > 0 {
			x := queue[0]
			queue = queue[1:]
			rest.Cells = append(rest.Cells, x)
			for _, n := range g.Neighbors(x) {
				if free[n] && !seen[n] {
					seen[n] = true
					queue = append(queue, n)
				}
			}
		}
		slices.SortFunc(rest.Cells, byLineCol)
		rest.Shape = ShapeKey(rest.Cells)
		pieces = append(pieces, rest)
	}
	return pieces
}

// connectedSets returns every connected set of at most limit cells of free that
// contains anchor and no cell before it in line, then column order, each sorted
// and each once (Redelmeier's enumeration).
func connectedSets(g hexgrid.Grid, anchor hexgrid.Coord, free map[hexgrid.Coord]bool, limit int) [][]hexgrid.Coord {
	marked := map[hexgrid.Coord]bool{anchor: true}
	var out [][]hexgrid.Coord
	var grow func(set, untried []hexgrid.Coord)
	grow = func(set, untried []hexgrid.Coord) {
		for len(untried) > 0 {
			c := untried[len(untried)-1]
			untried = untried[:len(untried)-1]
			next := append(slices.Clip(set), c)
			out = append(out, slices.SortedFunc(slices.Values(next), byLineCol))
			if len(next) == limit {
				continue
			}
			var added []hexgrid.Coord
			for _, n := range g.Neighbors(c) {
				if free[n] && byLineCol(n, anchor) > 0 && !marked[n] {
					marked[n] = true
					added = append(added, n)
				}
			}
			grow(next, append(slices.Clip(untried), added...))
			for _, n := range added {
				delete(marked, n)
			}
		}
	}
	grow(nil, []hexgrid.Coord{anchor})
	return out
}

// Layers returns what to draw on hex c in each pass: the background, a stamp
// anchored on c, the overlays and the standing items. A stamped cluster is drawn
// as its stamp on the base tiles when stamped(cluster) is true (the game asks
// that every cell is known); otherwise its cells keep their per-hex art. Under a
// stamp the terrain's own standing items give way to the stamp's; those of the
// infrastructure stay.
func (s *Scene) Layers(c hexgrid.Coord, stamped func(*Cluster) bool) (bg, stamp string, overlays, standing []string) {
	cell := s.Cells[c.Line][c.Col]
	bg, overlays, standing = cell.Background, cell.Overlays, cell.Standing
	if cell.Cluster == 0 {
		return bg, "", overlays, standing
	}
	cl := &s.Clusters[cell.Cluster-1]
	if cl.Stamp == "" || !stamped(cl) {
		return bg, "", overlays, standing
	}
	var kept []string
	if cl.Cells[0] == c {
		stamp = cl.Stamp
		if cl.Standing != "" {
			kept = append(kept, cl.Standing)
		}
	}
	for _, id := range standing {
		rest, _ := strings.CutPrefix(id, "standing/")
		if _, _, terrain := clusterKind(rest); !terrain {
			kept = append(kept, id)
		}
	}
	return cell.Under, stamp, overlays, kept
}
