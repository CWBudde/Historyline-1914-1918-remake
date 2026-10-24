package mapcompose

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Forest crowns are scattered over the world instead of baked into one
// template per hex: sites on a jittered grid in map pixels, each one a crown when
// the hex under the site, moved by a low-frequency wander, is forest without
// infrastructure in a season whose crown set the catalog has. The wander lets the
// woodland edge leave the hex edge by up to edgeWander pixels either way; species
// follow a slower noise, so they grow in stands. A crown belongs to the hex under its own site and is drawn with
// that hex's standing items. Everything is seeded by map position, so the result
// does not depend on traversal order or the visible part of the map.

// Placed is a standing item drawn with its anchor at (X, Y) logical pixels from
// its hex centre.
type Placed struct {
	ID   string
	X, Y float64
}

// CrownSpecies and CrownSizes name the scattered crown set (art/direction.md, F2):
// three species, each small, medium and large, in world pixels. Winter forest
// takes WinterCrownSpecies (F3) in their place: two conifers and a bare
// broadleaf, so a stand keeps its place and size across the seasons.
var (
	CrownSpecies       = []string{"oak", "beech", "birch"}
	WinterCrownSpecies = []string{"spruce", "fir", "bare"}
	CrownSizes         = []int{16, 22, 28}
	// crownWeights pick the sizes in that order.
	crownWeights = []float64{0.25, 0.4, 0.35}
)

// CrownID is the standing asset ID of one scattered crown.
func CrownID(species string, size int) string {
	return fmt.Sprintf("standing/scatter/crown/%s-%d", species, size)
}

const (
	crownPitch      = 15.0 // map pixels between grid sites: about 25 crowns and 90 % cover per interior hex
	crownJitter     = 0.42 // site jitter as a fraction of the pitch, each way
	edgeWander      = 12.0 // most the woodland edge moves from the hex edge
	edgeWavelength  = 64.0
	standWavelength = 190.0
	scatterSalt     = 0x5ca77e7c0f0e57a1
)

// bareShare of a bare winter stand's sites carry a crown, so snow shows between
// the leafless branches instead of a wash of overlapping twigs.
const bareShare = 0.45

// crownSeason is the forest and open terrain of one season and its species.
// With snowFloor its scattered forest stands on the open ground's tile: bare
// winter woods show snow between their branches, not the dark floor of the
// baked contexts.
type crownSeason struct {
	forest, grass string
	species       []string
	snowFloor     bool
}

var crownSeasons = []crownSeason{{"forest", "grass", CrownSpecies, false}, {"forest_w", "grass_w", WinterCrownSpecies, true}}

// crownSet reports whether the catalog has every crown of species and its shadow.
func crownSet(species []string, has func(string) bool) bool {
	for _, sp := range species {
		for _, size := range CrownSizes {
			if id := CrownID(sp, size); !has(id) || !has(ShadowID(id)) {
				return false
			}
		}
	}
	return true
}

// scatterCrowns places the crowns and drops the baked crowns and shadows of the
// forest hexes the scatter covers. Forest with infrastructure keeps its art, and
// so does the forest of a season whose crown set the catalog lacks.
func (s *Scene) scatterCrowns(at func(hexgrid.Coord) Hex, has func(string) bool) {
	var seasons []crownSeason
	for _, cs := range crownSeasons {
		if crownSet(cs.species, has) {
			seasons = append(seasons, cs)
		}
	}
	if len(seasons) == 0 {
		return
	}
	g := s.Grid
	// forestOf returns the season whose scattered forest c is.
	forestOf := func(c hexgrid.Coord) (crownSeason, bool) {
		h := at(c)
		for _, cs := range seasons {
			if h.Terrain == cs.forest && len(h.Infra) == 0 {
				return cs, true
			}
		}
		return crownSeason{}, false
	}
	open := func(c hexgrid.Coord, cs crownSeason) bool {
		h := at(c)
		return len(h.Infra) == 0 && (h.Terrain == cs.forest || h.Terrain == cs.grass)
	}
	for l := range s.Cells {
		for col := range s.Cells[l] {
			if cs, ok := forestOf(hexgrid.Coord{Line: l, Col: col}); ok {
				cell := &s.Cells[l][col]
				cell.Standing = slices.DeleteFunc(cell.Standing, func(id string) bool { return strings.HasPrefix(id, "standing/context/forest/") })
				if floor := baseTile(cs.grass, l, col, has); cs.snowFloor && has(floor) {
					cell.Background, cell.Context = floor, ""
				}
			}
		}
	}
	w, h := g.Bounds()
	for j := -1; float64(j)*crownPitch < h+crownPitch; j++ {
		for i := -1; float64(i)*crownPitch < w+crownPitch; i++ {
			x := coordinateHash(j, i) ^ scatterSalt
			x = mix(x)
			px := (float64(i) + crownJitter*signed(x)) * crownPitch
			x = mix(x)
			py := (float64(j) + crownJitter*signed(x)) * crownPitch
			wx := edgeWander * (2*valueNoise(px/edgeWavelength, py/edgeWavelength, 1) - 1)
			wy := edgeWander * (2*valueNoise(px/edgeWavelength, py/edgeWavelength, 2) - 1)
			src, ok := g.PixelToHex(px+wx, py+wy)
			if !ok {
				continue
			}
			cs, ok := forestOf(src)
			if !ok {
				continue
			}
			owner, ok := g.PixelToHex(px, py)
			if !ok || !open(owner, cs) {
				continue
			}
			sp := cs.species[min(len(cs.species)-1, int(valueNoise(px/standWavelength, py/standWavelength, 3)*float64(len(cs.species))))]
			x = mix(x)
			size := CrownSizes[pick(unit(x), crownWeights)]
			if sp == "bare" {
				if x = mix(x); unit(x) >= bareShare {
					continue
				}
			}
			cx, cy := hexgrid.Center(owner)
			cell := &s.Cells[owner.Line][owner.Col]
			cell.Scatter = append(cell.Scatter, Placed{ID: CrownID(sp, size), X: px - cx, Y: py - cy})
		}
	}
	for l := range s.Cells {
		for col := range s.Cells[l] {
			slices.SortFunc(s.Cells[l][col].Scatter, func(a, b Placed) int { return cmp.Or(cmp.Compare(a.Y, b.Y), cmp.Compare(a.X, b.X)) })
		}
	}
}

// ScatterAt returns the scattered items drawn with hex c's standing items, in
// drawing order; none while c lies under a drawn cluster stamp.
func (s *Scene) ScatterAt(c hexgrid.Coord, stamped func(*Cluster) bool) []Placed {
	cell := s.Cells[c.Line][c.Col]
	if cell.Cluster != 0 {
		if cl := &s.Clusters[cell.Cluster-1]; cl.Stamp != "" && stamped(cl) {
			return nil
		}
	}
	return cell.Scatter
}

// ScatterShadowsAt returns the cast shadows of ScatterAt, drawn before all
// standing items. A shadow shares its crown's position; its canvas carries the
// offset.
func (s *Scene) ScatterShadowsAt(c hexgrid.Coord, stamped func(*Cluster) bool) []Placed {
	items := s.ScatterAt(c, stamped)
	if len(items) == 0 {
		return nil
	}
	out := make([]Placed, len(items))
	for i, p := range items {
		out[i] = Placed{ID: ShadowID(p.ID), X: p.X, Y: p.Y}
	}
	return out
}

func mix(x uint64) uint64 {
	x += 0x9e3779b97f4a7c15
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	return x ^ (x >> 31)
}

// unit maps x to [0, 1); signed to [-1, 1).
func unit(x uint64) float64   { return float64(x>>11) / (1 << 53) }
func signed(x uint64) float64 { return 2*unit(x) - 1 }

// pick returns the index of the weight band that u in [0, 1) falls in.
func pick(u float64, weights []float64) int {
	for i, w := range weights {
		if u < w {
			return i
		}
		u -= w
	}
	return len(weights) - 1
}

// valueNoise is smooth lattice noise in [0, 1) at (x, y) lattice units; channel
// selects an independent field.
func valueNoise(x, y float64, channel uint64) float64 {
	x0, y0 := math.Floor(x), math.Floor(y)
	fx, fy := smooth(x-x0), smooth(y-y0)
	at := func(i, j float64) float64 {
		return unit(mix(coordinateHash(int(j), int(i)) ^ channel*0x632be59bd9b4e019))
	}
	top := at(x0, y0) + fx*(at(x0+1, y0)-at(x0, y0))
	bottom := at(x0, y0+1) + fx*(at(x0+1, y0+1)-at(x0, y0+1))
	return top + fy*(bottom-top)
}

func smooth(t float64) float64 { return t * t * (3 - 2*t) }
