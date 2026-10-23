package mapcompose

import (
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// forestMap is a w×l map of grass with a forest block in lines and columns
// [from, to).
func forestMap(w, l, from, to int) (hexgrid.Grid, func(hexgrid.Coord) Hex) {
	g := hexgrid.Grid{Width: w, Lines: l}
	return g, func(c hexgrid.Coord) Hex {
		if c.Line >= from && c.Line < to && c.Col >= from/2 && c.Col < to/2 {
			return Hex{Terrain: "forest"}
		}
		return Hex{Terrain: "grass"}
	}
}

// withCrowns reports the crown set, their shadows and every baked forest crown.
func withCrowns(id string) bool {
	return strings.HasPrefix(id, "standing/scatter/crown/") || strings.HasPrefix(id, "standing/shadow/scatter/crown/") ||
		strings.HasPrefix(id, "context/forest/") || strings.HasPrefix(id, "standing/context/forest/") ||
		strings.HasPrefix(id, "standing/shadow/context/forest/")
}

func TestCrownScatterFillsInteriorForest(t *testing.T) {
	g, at := forestMap(12, 24, 4, 20)
	s := Compose(g, at, withCrowns)
	var counts []int
	for l := 8; l < 16; l++ {
		for col := 4; col < 8; col++ {
			cell := s.Cells[l][col]
			counts = append(counts, len(cell.Scatter))
			if slices.ContainsFunc(cell.Standing, func(id string) bool { return strings.HasPrefix(id, "standing/context/forest/") }) {
				t.Fatalf("%d,%d keeps its baked crowns %v", l, col, cell.Standing)
			}
			if len(cell.Shadows) != 0 {
				t.Fatalf("%d,%d keeps baked shadows %v", l, col, cell.Shadows)
			}
		}
	}
	sum := 0
	for _, n := range counts {
		sum += n
	}
	if mean := float64(sum) / float64(len(counts)); mean < 20 || mean > 30 {
		t.Fatalf("%.1f crowns per interior hex (%v), want about 25", mean, counts)
	}
}

func TestCrownScatterIsDeterministicAndOrdered(t *testing.T) {
	g, at := forestMap(12, 24, 4, 20)
	a, b := Compose(g, at, withCrowns), Compose(g, at, withCrowns)
	species := map[string]bool{}
	for l := range a.Cells {
		for col := range a.Cells[l] {
			pa, pb := a.Cells[l][col].Scatter, b.Cells[l][col].Scatter
			if !slices.Equal(pa, pb) {
				t.Fatalf("%d,%d differs between runs", l, col)
			}
			for i, p := range pa {
				if i > 0 && (pa[i-1].Y > p.Y || pa[i-1].Y == p.Y && pa[i-1].X > p.X) {
					t.Fatalf("%d,%d not in drawing order: %v", l, col, pa)
				}
				species[strings.Split(strings.TrimPrefix(p.ID, "standing/scatter/crown/"), "-")[0]] = true
			}
		}
	}
	if len(species) < 2 {
		t.Fatalf("one species only: %v", species)
	}
}

func TestCrownScatterEdgeWandersAroundTheHexEdge(t *testing.T) {
	g, at := forestMap(12, 24, 4, 20)
	s := Compose(g, at, withCrowns)
	spilled, far := 0, 0
	for l := range s.Cells {
		for col := range s.Cells[l] {
			c := hexgrid.Coord{Line: l, Col: col}
			if at(c).Terrain != "grass" {
				continue
			}
			cx, cy := hexgrid.Center(c)
			for _, p := range s.Cells[l][col].Scatter {
				spilled++
				// A spilled crown sits within the wander of a forest hex.
				x, y := cx+p.X, cy+p.Y
				near := false
				for dy := -edgeWander; dy <= edgeWander; dy += 2 {
					for dx := -edgeWander; dx <= edgeWander; dx += 2 {
						if h, ok := g.PixelToHex(x+dx, y+dy); ok && at(h).Terrain == "forest" {
							near = true
						}
					}
				}
				if !near {
					far++
				}
			}
		}
	}
	if spilled == 0 {
		t.Fatal("the woodland edge follows the hex edge exactly")
	}
	if far != 0 {
		t.Fatalf("%d of %d spilled crowns lie beyond the edge wander", far, spilled)
	}
}

func TestCrownScatterNeedsTheWholeSet(t *testing.T) {
	g, at := forestMap(12, 24, 4, 20)
	s := Compose(g, at, func(id string) bool {
		return withCrowns(id) && id != CrownID("birch", 28)
	})
	for l := range s.Cells {
		for col := range s.Cells[l] {
			if len(s.Cells[l][col].Scatter) != 0 {
				t.Fatalf("%d,%d scatters crowns from an incomplete set", l, col)
			}
		}
	}
	if !slices.ContainsFunc(s.Cells[10][5].Standing, func(id string) bool { return strings.HasPrefix(id, "standing/context/forest/") }) {
		t.Fatalf("the baked crowns went away: %v", s.Cells[10][5].Standing)
	}
}

func TestCrownScatterSkipsInfrastructure(t *testing.T) {
	g, base := forestMap(12, 24, 4, 20)
	road := hexgrid.Coord{Line: 10, Col: 5}
	at := func(c hexgrid.Coord) Hex {
		h := base(c)
		if c == road {
			h.Infra = []string{"road-1-4"}
		}
		return h
	}
	s := Compose(g, at, withCrowns)
	if n := len(s.Cells[road.Line][road.Col].Scatter); n != 0 {
		t.Fatalf("%d crowns on the road hex", n)
	}
	cx, cy := hexgrid.Center(road)
	for l := range s.Cells {
		for col := range s.Cells[l] {
			x0, y0 := hexgrid.Center(hexgrid.Coord{Line: l, Col: col})
			for _, p := range s.Cells[l][col].Scatter {
				if math.Hypot(x0+p.X-cx, y0+p.Y-cy) < 30 {
					t.Fatalf("crown %v of %d,%d stands on the road hex", p, l, col)
				}
			}
		}
	}
}

func TestScatterShadowsPairTheirCrowns(t *testing.T) {
	g, at := forestMap(12, 24, 4, 20)
	s := Compose(g, at, withCrowns)
	all := func(*Cluster) bool { return true }
	c := hexgrid.Coord{Line: 10, Col: 5}
	crowns, shadows := s.ScatterAt(c, all), s.ScatterShadowsAt(c, all)
	if len(crowns) == 0 || len(shadows) != len(crowns) {
		t.Fatalf("%d crowns, %d shadows", len(crowns), len(shadows))
	}
	for i := range crowns {
		if shadows[i].ID != ShadowID(crowns[i].ID) || shadows[i].X != crowns[i].X || shadows[i].Y != crowns[i].Y {
			t.Fatalf("shadow %v for crown %v", shadows[i], crowns[i])
		}
	}
}
