package mapcompose

import (
	"io/fs"
	"math"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// shoreVertex is vertex v of the hex at c in world px, in the shoreline
// export's order (clockwise from the top left).
func shoreVertex(c hexgrid.Coord, v int) [2]float64 {
	x, y := hexgrid.Center(c)
	local := [6][2]float64{{25, 0}, {71, 0}, {96, 40}, {71, 80}, {25, 80}, {0, 40}}[v]
	return [2]float64{x - 48 + local[0], y - 40 + local[1]}
}

func TestVertexSharesMatchTheGeometry(t *testing.T) {
	for _, c := range []hexgrid.Coord{{Line: 4, Col: 3}, {Line: 5, Col: 3}} {
		around := hexgrid.Around(c)
		for v := range 6 {
			for _, n := range vertexShares[v] {
				a, b := shoreVertex(c, v), shoreVertex(around[n[0]], n[1])
				if math.Hypot(a[0]-b[0], a[1]-b[1]) > 1e-9 {
					t.Errorf("%+v vertex %d: neighbour %d vertex %d is at %v, not %v", c, v, n[0], n[1], b, a)
				}
			}
		}
		for i := range 6 {
			// The edge toward neighbour i has its midpoint halfway to its centre.
			a, b := edgeVertices(i)
			pa, pb := shoreVertex(c, a), shoreVertex(c, b)
			cx, cy := hexgrid.Center(c)
			x, y := hexgrid.Center(around[i])
			if math.Hypot((pa[0]+pb[0]-cx-x)/2, (pa[1]+pb[1]-cy-y)/2) > 1e-9 {
				t.Errorf("edge %d of %+v does not face its neighbour", i, c)
			}
		}
	}
}

func TestShoreShapes(t *testing.T) {
	for _, tt := range []struct {
		name string
		rows []string
		at   hexgrid.Coord
		want string
	}{
		// A pond: every vertex touches land.
		{"pond", []string{"grass grass grass", "grass grass", "grass water grass", "grass grass", "grass grass grass"}, hexgrid.Coord{Line: 2, Col: 1}, "shore-3f"},
		// Open water stays open.
		{"open", []string{"water water water", "water water", "water water water", "water water", "water water water"}, hexgrid.Coord{Line: 2, Col: 1}, "water"},
		// Land to the south: its two vertices are land.
		{"one land edge", []string{"water water water", "water water", "water water water", "water water", "water grass water"}, hexgrid.Coord{Line: 2, Col: 1}, "shore-18"},
		// A two-hex pond keeps the water between its halves.
		{"two-hex pond", []string{"grass grass grass", "grass grass", "grass water grass", "grass grass", "grass water grass", "grass grass", "grass grass grass"}, hexgrid.Coord{Line: 2, Col: 1}, "shore-3f-c08"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g, at := grid(tt.rows...)
			if got, gap := selectShore(g, at, tt.at, false, func(string) bool { return true }); got != tt.want || gap != nil {
				t.Errorf("shore %s (gap %v), want %s", got, gap, tt.want)
			}
			if s, ok := ParseShore(tt.want); tt.want != "water" && (!ok || s.ID() != tt.want) {
				t.Errorf("ParseShore(%s) = %+v, %v", tt.want, s, ok)
			}
		})
	}
}

// The spots of the original map 03 ("VENOM", campaign level 4) that met the
// grass along a straight hex edge.
func TestVenomCoastHasShores(t *testing.T) {
	b, err := fs.ReadFile(levels.Campaign, "campaign/level4_cp.json")
	if err != nil {
		t.Fatal(err)
	}
	l, err := level.Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	at := func(c hexgrid.Coord) Hex {
		h := l.Lines[c.Line][c.Col]
		return Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
	}
	for _, tt := range []struct {
		line, col int // 1-based, as the hex numbers show them
		want      string
	}{
		{8, 5, "shore-3f-c01"}, // the bay, open to the lake through its north edge
		{8, 7, "shore-3f"},     // the harbour pond between the houses
		{3, 8, "lakeside12"},   // the quays beside the bridge
		{5, 8, "lakeside123"},
	} {
		c := hexgrid.Coord{Line: tt.line - 1, Col: tt.col - 1}
		if got, _ := selectShore(l.Grid(), at, c, false, func(string) bool { return true }); got != tt.want {
			t.Errorf("%d,%d draws %s, want %s", tt.line, tt.col, got, tt.want)
		}
	}
}

// On every level no water hex meets land along a fully watery edge, and two
// water hexes agree on the vertices and channel of their shared edge.
func TestShoresMeetLandAndEachOther(t *testing.T) {
	for _, source := range []struct {
		fsys    fs.FS
		pattern string
	}{{levels.FS, "*.json"}, {levels.Campaign, "campaign/*.json"}} {
		names, err := fs.Glob(source.fsys, source.pattern)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range names {
			b, err := fs.ReadFile(source.fsys, name)
			if err != nil {
				t.Fatal(err)
			}
			l, err := level.Parse(b)
			if err != nil {
				t.Fatal(err)
			}
			g := l.Grid()
			at := func(c hexgrid.Coord) Hex {
				h := l.Lines[c.Line][c.Col]
				return Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
			}
			shape := func(c hexgrid.Coord) ShoreShape {
				id, _ := selectShore(g, at, c, false, func(string) bool { return true })
				if id == "water" {
					return ShoreShape{}
				}
				s, ok := ParseShore(id)
				if !ok {
					t.Fatalf("%s %+v: unknown shore %s", name, c, id)
				}
				return s
			}
			hard := 0
			for line, hexes := range l.Lines {
				for col, h := range hexes {
					c := hexgrid.Coord{Line: line, Col: col}
					if !wet(h.Terrain) || h.Terrain == "waterStones" {
						continue
					}
					s := shape(c)
					for i, n := range hexgrid.Around(c) {
						if !g.Contains(n) {
							continue
						}
						a, b := edgeVertices(i)
						t2 := at(n).Terrain
						if !wet(t2) {
							if s.Land&(1<<a) == 0 || s.Land&(1<<b) == 0 {
								hard++
							}
							continue
						}
						if t2 == "waterStones" {
							continue
						}
						o := shape(n)
						na, nb := edgeVertices((i + 3) % 6)
						bit := func(m uint8, v int) bool { return m&(1<<v) != 0 }
						if bit(s.Land, a) != bit(o.Land, nb) || bit(s.Land, b) != bit(o.Land, na) || bit(s.Channels, a) != bit(o.Channels, na) {
							t.Errorf("%s %+v and %+v disagree on their edge: %s / %s", name, c, n, s.ID(), o.ID())
						}
					}
				}
			}
			if hard != 0 {
				t.Errorf("%s: %d water edges meet land", strings.TrimSuffix(name, ".json"), hard)
			}
		}
	}
}

// A coast takes the season of most of its land: snow beside winter terrain,
// meadow beside summer terrain, and the map's season where only shores and
// water touch it.
func TestWinterCoastsTakeWinterShores(t *testing.T) {
	all := func(string) bool { return true }
	for _, tt := range []struct {
		name   string
		rows   []string
		winter bool
		want   string
	}{
		{"winter pond", []string{"grass_w grass_w grass_w", "grass_w grass_w", "grass_w water grass_w", "grass_w grass_w", "grass_w grass_w grass_w"}, false, "shore-3f_w"},
		{"summer pond on a winter map", []string{"grass grass grass", "grass grass", "grass water grass", "grass grass", "grass grass grass"}, true, "shore-3f"},
		{"mostly summer", []string{"water water water", "grass_w grass", "water water water", "water water", "water grass water"}, true, "shore-3f-c15"},
		{"tied, on a winter map", []string{"water water water", "water grass_w", "water water water", "water water", "water grass water"}, true, "shore-1e-c04_w"},
		{"roster shore", []string{"water water water", "water water", "water lakeside34 water", "water water", "water grass_w water"}, false, "lakeside34_w"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g, at := grid(tt.rows...)
			if got, gap := selectShore(g, at, hexgrid.Coord{Line: 2, Col: 1}, tt.winter, all); got != tt.want || gap != nil {
				t.Errorf("shore %s (gap %v), want %s", got, gap, tt.want)
			}
		})
	}
	// Without its winter form a winter coast keeps the summer shore and
	// reports the gap.
	g, at := grid("grass_w grass_w grass_w", "grass_w grass_w", "grass_w water grass_w", "grass_w grass_w", "grass_w grass_w grass_w")
	summer := func(id string) bool { return !strings.HasSuffix(id, "_w") }
	if got, gap := selectShore(g, at, hexgrid.Coord{Line: 2, Col: 1}, true, summer); got != "shore-3f" || gap == nil || gap.Descriptor != "shore-3f_w" {
		t.Errorf("shore %s, gap %+v", got, gap)
	}
	// Level 10 (map 09) lies in winter: its coast is snow.
	b, err := fs.ReadFile(levels.Campaign, "campaign/level10_cp.json")
	if err != nil {
		t.Fatal(err)
	}
	l, err := level.Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	s := Compose(l.Grid(), func(c hexgrid.Coord) Hex {
		h := l.Lines[c.Line][c.Col]
		return Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
	}, all)
	shores := 0
	for _, line := range s.Cells {
		for _, c := range line {
			if id := string(c.Background); strings.HasPrefix(id, "tile/lakeside") || strings.HasPrefix(id, "tile/shore-") {
				shores++
				if !strings.HasSuffix(id, "_w") {
					t.Errorf("summer shore %s on the winter coast", id)
				}
			}
		}
	}
	if shores < 40 {
		t.Errorf("only %d shores on level 10", shores)
	}
}

// Open water and stones in water take their winter form on a winter coast, by
// the vote the shores use, and stay summer without it or without the tile.
func TestWinterWaterTakesWinterTiles(t *testing.T) {
	all := func(string) bool { return true }
	summer := func(id string) bool { return !strings.HasSuffix(id, "_w") }
	open := []string{"water water water", "water water", "water water water", "water water", "water water water"}
	for _, tt := range []struct {
		name   string
		rows   []string
		winter bool
		has    func(string) bool
		want   string
	}{
		{"open water on a winter map", open, true, all, "water_w"},
		{"open water on a summer map", open, false, all, "water"},
		{"no winter tile", open, true, summer, "water"},
		{"stones on a winter coast", []string{"grass_w grass_w grass_w", "grass_w grass_w", "grass_w waterStones grass_w", "grass_w grass_w", "grass_w grass_w grass_w"}, false, all, "waterStones_w"},
		{"stones on a summer coast", []string{"grass grass grass", "grass grass", "grass waterStones grass", "grass grass", "grass grass grass"}, true, all, "waterStones"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g, at := grid(tt.rows...)
			if got, gap := selectShore(g, at, hexgrid.Coord{Line: 2, Col: 1}, tt.winter, tt.has); got != tt.want || gap != nil {
				t.Errorf("water %s (gap %v), want %s", got, gap, tt.want)
			}
		})
	}
}
