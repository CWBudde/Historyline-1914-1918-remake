package mapcompose

import (
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// withProps reports the meadow prop set.
func withProps(id string) bool { return strings.HasPrefix(id, "standing/scatter/meadow/") }

// plainMap is a w×l map of one terrain.
func plainMap(w, l int, terrain string) (hexgrid.Grid, func(hexgrid.Coord) Hex) {
	return hexgrid.Grid{Width: w, Lines: l}, func(hexgrid.Coord) Hex { return Hex{Terrain: terrain} }
}

// classOf is the class of prop id: "tuft" for standing/scatter/meadow/tuft-2.
func classOf(id string) string {
	name, _, _ := strings.Cut(strings.TrimPrefix(id, "standing/scatter/meadow/"), "-")
	return name
}

// absolute returns every prop of s in map pixels, by class.
func absolute(s Scene) map[string][][2]float64 {
	out := map[string][][2]float64{}
	for l := range s.Cells {
		for col := range s.Cells[l] {
			cx, cy := hexgrid.Center(hexgrid.Coord{Line: l, Col: col})
			for _, p := range s.Cells[l][col].Ground {
				out[classOf(p.ID)] = append(out[classOf(p.ID)], [2]float64{cx + p.X, cy + p.Y})
			}
		}
	}
	return out
}

func TestPropsKeepTheirClassDistance(t *testing.T) {
	g, at := plainMap(12, 24, "grass")
	for class, ps := range absolute(Compose(g, at, withProps)) {
		r := propClassOf(class).radius
		for i := range ps {
			for j := i + 1; j < len(ps); j++ {
				if d := math.Hypot(ps[i][0]-ps[j][0], ps[i][1]-ps[j][1]); d < r {
					t.Fatalf("%s at %v and %v: %.1f apart, want at least %.0f", class, ps[i], ps[j], d, r)
				}
			}
		}
	}
}

func TestPropDensityFollowsTheTable(t *testing.T) {
	for _, terrain := range []string{"grass", "hills"} {
		g, at := plainMap(20, 60, terrain)
		s := Compose(g, at, withProps)
		counts, hexes := map[string]int{}, 0
		for l := 4; l < 56; l++ {
			for col := 2; col < 17; col++ {
				hexes++
				for _, p := range s.Cells[l][col].Ground {
					counts[classOf(p.ID)]++
				}
			}
		}
		for _, c := range propClasses {
			want := c.perHex[terrain]
			got := float64(counts[c.name]) / float64(hexes)
			if want == 0 && got != 0 || math.Abs(got-want) > 0.2*want {
				t.Errorf("%s %s: %.2f per hex, want %.2f", terrain, c.name, got, want)
			}
		}
	}
}

func TestPropsAreSeededByWorldPosition(t *testing.T) {
	g, at := plainMap(12, 30, "grass")
	far := hexgrid.Coord{Line: 24, Col: 6}
	changed := func(c hexgrid.Coord) Hex {
		if c == far {
			return Hex{Terrain: "water"}
		}
		return at(c)
	}
	a, b, again := Compose(g, at, withProps), Compose(g, changed, withProps), Compose(g, at, withProps)
	here := hexgrid.Coord{Line: 6, Col: 6}
	if len(a.Cells[here.Line][here.Col].Ground) == 0 {
		t.Fatal("no props")
	}
	if !slices.Equal(a.Cells[here.Line][here.Col].Ground, b.Cells[here.Line][here.Col].Ground) {
		t.Fatal("a hex 9 hexes away changed this hex's props")
	}
	for l := range a.Cells {
		for col := range a.Cells[l] {
			pa := a.Cells[l][col].Ground
			if !slices.Equal(pa, again.Cells[l][col].Ground) {
				t.Fatalf("%d,%d differs between runs", l, col)
			}
			for i := 1; i < len(pa); i++ {
				if pa[i-1].Y > pa[i].Y || pa[i-1].Y == pa[i].Y && pa[i-1].X > pa[i].X {
					t.Fatalf("%d,%d not in drawing order: %v", l, col, pa)
				}
			}
		}
	}
}

func TestPropsStayOnOpenMeadow(t *testing.T) {
	g, base := forestMap(12, 24, 4, 20)
	road := hexgrid.Coord{Line: 2, Col: 5}
	at := func(c hexgrid.Coord) Hex {
		h := base(c)
		if c == road {
			h.Infra = []string{"road-1-4"}
		}
		return h
	}
	s := Compose(g, at, withProps)
	n := 0
	for l := range s.Cells {
		for col := range s.Cells[l] {
			cx, cy := hexgrid.Center(hexgrid.Coord{Line: l, Col: col})
			for _, p := range s.Cells[l][col].Ground {
				n++
				half := propSize(p.ID) / 2
				for _, d := range [][2]float64{{0, 0}, {-half, 0}, {half, 0}, {0, -half}, {0, half}} {
					h, ok := g.PixelToHex(cx+p.X+d[0], cy+p.Y+d[1])
					if !ok || at(h).Terrain != "grass" || len(at(h).Infra) != 0 {
						t.Fatalf("%s of %d,%d reaches %v (%+v)", p.ID, l, col, h, at(h))
					}
				}
			}
		}
	}
	if n == 0 {
		t.Fatal("no props")
	}
}

func TestPropsNeedTheWholeSet(t *testing.T) {
	g, at := plainMap(12, 24, "grass")
	s := Compose(g, at, func(id string) bool { return withProps(id) && id != PropID("straw", 2) })
	for l := range s.Cells {
		for col := range s.Cells[l] {
			if len(s.Cells[l][col].Ground) != 0 {
				t.Fatalf("%d,%d scatters props from an incomplete set", l, col)
			}
		}
	}
}

func TestGroundAtHidesUnderADrawnStamp(t *testing.T) {
	ground := []Placed{{ID: PropID("tuft", 1)}}
	s := Scene{Cells: [][]Cell{{{Ground: ground, Cluster: 1}}}, Clusters: []Cluster{{Stamp: "stamp/x"}}}
	c := hexgrid.Coord{}
	if got := s.GroundAt(c, func(*Cluster) bool { return true }); got != nil {
		t.Fatalf("props under a drawn stamp: %v", got)
	}
	if got := s.GroundAt(c, func(*Cluster) bool { return false }); !slices.Equal(got, ground) {
		t.Fatalf("props without the stamp: %v", got)
	}
}
