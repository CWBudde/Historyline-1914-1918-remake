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

// Cluster is a connected group of hexes of one kind.
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

// registerClusters finds the clusters of the scene and selects their stamps.
// The variant is chosen from the variants the catalog has, seeded by the
// cluster's first cell.
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
			slices.SortFunc(cl.Cells, func(a, b hexgrid.Coord) int {
				return cmp.Or(cmp.Compare(a.Line, b.Line), cmp.Compare(a.Col, b.Col))
			})
			cl.Shape = ShapeKey(cl.Cells)
			variants := 0
			for variants < MaxStampVariants && has(StampID(kind, cl.Shape, variants+1)) {
				variants++
			}
			if variants > 0 {
				first := cl.Cells[0]
				variant := int((coordinateHash(first.Line, first.Col)>>32)%uint64(variants)) + 1
				cl.Stamp = StampID(kind, cl.Shape, variant)
				if id := StandingID(cl.Stamp); has(id) {
					cl.Standing = id
				}
			}
			s.Clusters = append(s.Clusters, cl)
			for _, c := range cl.Cells {
				cell := &s.Cells[c.Line][c.Col]
				cell.Cluster = len(s.Clusters)
				if cl.Stamp != "" {
					cell.Under = baseTile(base, c.Line, c.Col, has)
				}
			}
		}
	}
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
