package mapcompose

import (
	"cmp"
	"fmt"
	"math"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Meadow props (art/direction.md, G2) are small flat ground items scattered over
// open grass and hills: tufts, flowers, bare soil, straw and pebbles. Each class
// has its own Poisson-disc site set in map pixels, thinned by a density table per
// terrain family and by a slow drift noise, so items gather loosely without ever
// touching another of their class. Everything is seeded by map position. The
// items carry their own painted shadow and are drawn on the ground, under the
// infrastructure, cast shadows and standing items.

// propClass is one class of meadow prop.
type propClass struct {
	name   string
	sizes  []float64 // world px of each item, small to large
	radius float64   // least distance between two items of the class
	drift  float64   // how strongly the slow noise gathers the items, 0..1
	perHex map[string]float64
}

// propClasses is the density table: items per hex by terrain family. Families
// without an entry get none.
var propClasses = []propClass{
	{"tuft", []float64{3, 4.5, 6, 8}, 8, 0.5, map[string]float64{"grass": 8, "hills": 5}},
	{"flowers", []float64{6, 10, 14}, 14, 0.6, map[string]float64{"grass": 1.2}},
	{"soil", []float64{4, 7, 10}, 10, 0.6, map[string]float64{"grass": 0.8, "hills": 1.2}},
	{"straw", []float64{10, 16}, 16, 0.6, map[string]float64{"grass": 0.5, "hills": 0.8}},
	{"pebbles", []float64{4, 7, 10}, 10, 0.6, map[string]float64{"grass": 0.8, "hills": 1.5}},
}

const (
	// poissonYield is the share of grid candidates poissonSites keeps; the
	// process is scale-free, so the sites per area are poissonYield/r².
	poissonYield    = 0.3588
	hexArea         = hexgrid.ColPitch * hexgrid.LinePitch // map pixels per hex
	driftWavelength = 160.0
	propSalt        = 0x6d3ad0f1a5e9c247
)

// PropID is the standing asset ID of meadow prop n (1-based, small to large) of
// a class.
func PropID(class string, n int) string {
	return fmt.Sprintf("standing/scatter/meadow/%s-%d", class, n)
}

var propSizes = func() map[string]float64 {
	out := map[string]float64{}
	for _, c := range propClasses {
		for i, size := range c.sizes {
			out[PropID(c.name, i+1)] = size
		}
	}
	return out
}()

// propSize returns the world size of prop id, 0 when it is none.
func propSize(id string) float64 { return propSizes[id] }

func propClassOf(name string) propClass {
	i := slices.IndexFunc(propClasses, func(c propClass) bool { return c.name == name })
	return propClasses[i]
}

// propSet reports whether the catalog has every prop.
func propSet(has func(string) bool) bool {
	for id := range propSizes {
		if !has(id) {
			return false
		}
	}
	return true
}

// poissonSites returns points over [0, w)×[0, h) at least r apart: one jittered
// candidate per r×r grid cell, kept when no candidate within r has a higher
// priority. A candidate depends only on its grid cell, so the result is a pure
// function of map position.
func poissonSites(w, h, r float64, salt uint64) [][2]float64 {
	type candidate struct {
		x, y float64
		prio uint64
	}
	at := func(i, j int) candidate {
		x := mix(coordinateHash(j, i) ^ salt)
		px := (float64(i) + unit(x)) * r
		x = mix(x)
		py := (float64(j) + unit(x)) * r
		return candidate{px, py, mix(x)}
	}
	var out [][2]float64
	for j := 0; float64(j)*r < h; j++ {
		for i := 0; float64(i)*r < w; i++ {
			c, kept := at(i, j), true
			for dj := -1; dj <= 1 && kept; dj++ {
				for di := -1; di <= 1 && kept; di++ {
					if n := at(i+di, j+dj); (di != 0 || dj != 0) && n.prio > c.prio && math.Hypot(n.x-c.x, n.y-c.y) < r {
						kept = false
					}
				}
			}
			if kept {
				out = append(out, [2]float64{c.x, c.y})
			}
		}
	}
	return out
}

// scatterProps places the meadow props on open grass and hills.
func (s *Scene) scatterProps(at func(hexgrid.Coord) Hex, has func(string) bool) {
	if !propSet(has) {
		return
	}
	g := s.Grid
	open := func(x, y float64) (string, bool) {
		c, ok := g.PixelToHex(x, y)
		if !ok {
			return "", false
		}
		h := at(c)
		return h.Terrain, len(h.Infra) == 0 && (h.Terrain == "grass" || h.Terrain == "hills")
	}
	w, h := g.Bounds()
	for k, class := range propClasses {
		salt := propSalt ^ uint64(k+1)*0x9e3779b97f4a7c15
		base := class.radius * class.radius / (poissonYield * hexArea)
		for _, p := range poissonSites(w, h, class.radius, salt) {
			terrain, ok := open(p[0], p[1])
			if !ok || class.perHex[terrain] == 0 {
				continue
			}
			x := mix(coordinateHash(int(p[1]*16), int(p[0]*16)) ^ salt)
			drift := 1 + class.drift*(2*valueNoise(p[0]/driftWavelength, p[1]/driftWavelength, uint64(10+k))-1)
			if unit(x) >= class.perHex[terrain]*base*drift {
				continue
			}
			x = mix(x)
			n := min(len(class.sizes)-1, int(unit(x)*float64(len(class.sizes))))
			half := class.sizes[n] / 2
			if _, ok := open(p[0]-half, p[1]); !ok {
				continue
			}
			if _, ok := open(p[0]+half, p[1]); !ok {
				continue
			}
			if _, ok := open(p[0], p[1]-half); !ok {
				continue
			}
			if _, ok := open(p[0], p[1]+half); !ok {
				continue
			}
			owner, _ := g.PixelToHex(p[0], p[1])
			cx, cy := hexgrid.Center(owner)
			cell := &s.Cells[owner.Line][owner.Col]
			cell.Ground = append(cell.Ground, Placed{ID: PropID(class.name, n+1), X: p[0] - cx, Y: p[1] - cy})
		}
	}
	for l := range s.Cells {
		for col := range s.Cells[l] {
			slices.SortFunc(s.Cells[l][col].Ground, func(a, b Placed) int { return cmp.Or(cmp.Compare(a.Y, b.Y), cmp.Compare(a.X, b.X)) })
		}
	}
}

// GroundAt returns the ground props of hex c, drawn after the cluster stamps, in
// drawing order; none while c lies under a drawn cluster stamp.
func (s *Scene) GroundAt(c hexgrid.Coord, stamped func(*Cluster) bool) []Placed {
	cell := s.Cells[c.Line][c.Col]
	if cell.Cluster != 0 {
		if cl := &s.Clusters[cell.Cluster-1]; cl.Stamp != "" && stamped(cl) {
			return nil
		}
	}
	return cell.Ground
}
