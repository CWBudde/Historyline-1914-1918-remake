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

// withCrowns reports the crown set, their shadows, every baked forest crown
// and the snow tile.
func withCrowns(id string) bool {
	return strings.HasPrefix(id, "standing/scatter/crown/") || strings.HasPrefix(id, "standing/shadow/scatter/crown/") ||
		strings.HasPrefix(id, "context/forest/") || strings.HasPrefix(id, "standing/context/forest/") ||
		strings.HasPrefix(id, "standing/shadow/context/forest/") || id == TileID("grass_w")
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
			if !strings.HasPrefix(cell.Background, "context/forest/") {
				t.Fatalf("%d,%d drops its forest floor for %s", l, col, cell.Background)
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

// winterMap is forestMap in winter.
func winterMap(w, l, from, to int) (hexgrid.Grid, func(hexgrid.Coord) Hex) {
	g, at := forestMap(w, l, from, to)
	return g, func(c hexgrid.Coord) Hex {
		h := at(c)
		h.Terrain += "_w"
		return h
	}
}

func TestCrownScatterFillsWinterForestWithWinterCrowns(t *testing.T) {
	g, at := winterMap(12, 24, 4, 20)
	s := Compose(g, at, withCrowns)
	sum, n := 0, 0
	species := map[string]bool{}
	for l := range s.Cells {
		for col := range s.Cells[l] {
			cell := s.Cells[l][col]
			if l >= 8 && l < 16 && col >= 4 && col < 8 {
				sum += len(cell.Scatter)
				n++
				if slices.ContainsFunc(cell.Standing, func(id string) bool { return strings.HasPrefix(id, "standing/context/forest/") }) {
					t.Fatalf("%d,%d keeps its baked crowns %v", l, col, cell.Standing)
				}
				// Bare winter woods stand on snow, not on the dark floor of the
				// baked contexts.
				if cell.Background != TileID("grass_w") || cell.Context != "" {
					t.Fatalf("%d,%d stands on %s (context %q), want the snow tile", l, col, cell.Background, cell.Context)
				}
			}
			for _, p := range cell.Scatter {
				species[strings.Split(strings.TrimPrefix(p.ID, "standing/scatter/crown/"), "-")[0]] = true
			}
		}
	}
	if mean := float64(sum) / float64(n); mean < 20 || mean > 30 {
		t.Fatalf("%.1f crowns per interior winter hex, want about 25", mean)
	}
	for sp := range species {
		if !slices.Contains(WinterCrownSpecies, sp) {
			t.Fatalf("winter forest scatters %q (species %v)", sp, species)
		}
	}
	if len(species) < 2 {
		t.Fatalf("one species only: %v", species)
	}
}

func TestCrownScatterNeedsTheSeasonsWholeSet(t *testing.T) {
	summerOnly := func(id string) bool {
		for _, sp := range WinterCrownSpecies {
			if strings.Contains(id, "/scatter/crown/"+sp+"-") {
				return false
			}
		}
		return withCrowns(id)
	}
	g, at := winterMap(12, 24, 4, 20)
	s := Compose(g, at, summerOnly)
	for l := range s.Cells {
		for col := range s.Cells[l] {
			if len(s.Cells[l][col].Scatter) != 0 {
				t.Fatalf("%d,%d scatters crowns in winter without the winter set", l, col)
			}
		}
	}
	if !slices.ContainsFunc(s.Cells[10][5].Standing, func(id string) bool { return strings.HasPrefix(id, "standing/context/forest/") }) {
		t.Fatalf("the baked winter crowns went away: %v", s.Cells[10][5].Standing)
	}
	g, at = forestMap(12, 24, 4, 20)
	if s := Compose(g, at, summerOnly); len(s.Cells[10][5].Scatter) == 0 {
		t.Fatal("summer forest needs the winter set")
	}
}

// A bare winter stand keeps about bareShare of its sites, so snow shows
// between the bare crowns; the conifers keep every site of the summer stand
// they replace (spruce for oak, fir for beech) at the same place and size.
func TestBareStandsThinOut(t *testing.T) {
	count := func(g hexgrid.Grid, at func(hexgrid.Coord) Hex) map[string][]Placed {
		s := Compose(g, at, withCrowns)
		out := map[string][]Placed{}
		for l := range s.Cells {
			for col := range s.Cells[l] {
				for _, p := range s.Cells[l][col].Scatter {
					sp := strings.Split(strings.TrimPrefix(p.ID, "standing/scatter/crown/"), "-")[0]
					out[sp] = append(out[sp], p)
				}
			}
		}
		return out
	}
	summer, winter := count(forestMap(40, 80, 2, 78)), count(winterMap(40, 80, 2, 78))
	for summerSp, winterSp := range map[string]string{"oak": "spruce", "beech": "fir"} {
		if len(summer[summerSp]) == 0 || len(winter[winterSp]) != len(summer[summerSp]) {
			t.Errorf("%s: %d crowns, %s: %d", summerSp, len(summer[summerSp]), winterSp, len(winter[winterSp]))
		}
	}
	birch, bare := float64(len(summer["birch"])), float64(len(winter["bare"]))
	if birch < 200 || math.Abs(bare/birch-bareShare) > 0.05 {
		t.Errorf("%v bare crowns for %v birches (%.2f), want about %.2f", bare, birch, bare/birch, bareShare)
	}
}
