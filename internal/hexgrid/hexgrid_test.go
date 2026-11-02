package hexgrid

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// golden is testdata/adjacent.json, a frozen golden of the Python remake's
// adjacentHexes in hlrData.py.
type golden map[string]struct {
	Width int `json:"width"`
	Lines int `json:"lines"`
	Hexes []struct {
		Hex       [2]int   `json:"hex"`
		Neighbors [][2]int `json:"neighbors"`
	} `json:"hexes"`
}

func loadGolden(t *testing.T) golden {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "adjacent.json"))
	if err != nil {
		t.Fatal(err)
	}
	var g golden
	if err := json.Unmarshal(raw, &g); err != nil {
		t.Fatal(err)
	}
	return g
}

func coord(p [2]int) Coord { return Coord{Line: p[0], Col: p[1]} }

func TestNeighborsMatchPython(t *testing.T) {
	g := loadGolden(t)

	if len(g) == 0 {
		t.Fatal("no golden grids")
	}
	// The grids are the sizes of the Python remake's levels.
	for name, lv := range g {
		grid := Grid{Width: lv.Width, Lines: lv.Lines}
		count := 0
		for _, h := range lv.Hexes {
			c := coord(h.Hex)
			if !grid.Contains(c) {
				t.Errorf("%s: %v should be on the map", name, c)
			}
			want := make([]Coord, 0, len(h.Neighbors))
			for _, n := range h.Neighbors {
				want = append(want, coord(n))
			}
			if got := grid.Neighbors(c); !reflect.DeepEqual(got, want) {
				t.Errorf("%s: Neighbors(%v) = %v, want %v", name, c, got, want)
			}
			count++
		}
		if count != grid.Size() {
			t.Errorf("%s: golden has %d hexes, grid has %d", name, count, grid.Size())
		}
	}
}

func TestEdgeCases(t *testing.T) {
	grid := Grid{Width: 8, Lines: 47} // PULSE

	contains := []struct {
		c    Coord
		want bool
	}{
		{Coord{0, 0}, true},
		{Coord{0, 7}, true},   // last hex of an even line
		{Coord{1, 6}, true},   // last hex of an odd line
		{Coord{1, 7}, false},  // odd lines are one hex shorter
		{Coord{46, 7}, true},  // last line is even
		{Coord{47, 0}, false}, // past the last line
		{Coord{-1, 0}, false},
		{Coord{0, -1}, false},
		{Coord{0, 8}, false},
	}
	for _, tc := range contains {
		if got := grid.Contains(tc.c); got != tc.want {
			t.Errorf("Contains(%v) = %v, want %v", tc.c, got, tc.want)
		}
	}

	neighbors := []struct {
		c    Coord
		want []Coord
	}{
		{Coord{0, 0}, []Coord{{1, 0}, {2, 0}}},
		{Coord{0, 7}, []Coord{{2, 7}, {1, 6}}},
		{Coord{1, 0}, []Coord{{0, 1}, {2, 1}, {3, 0}, {2, 0}, {0, 0}}},
		{Coord{1, 6}, []Coord{{0, 7}, {2, 7}, {3, 6}, {2, 6}, {0, 6}}},
		{Coord{2, 0}, []Coord{{1, 0}, {3, 0}, {4, 0}, {0, 0}}},
		{Coord{2, 7}, []Coord{{4, 7}, {3, 6}, {1, 6}, {0, 7}}},
		{Coord{20, 3}, []Coord{{19, 3}, {21, 3}, {22, 3}, {21, 2}, {19, 2}, {18, 3}}},
		{Coord{21, 3}, []Coord{{20, 4}, {22, 4}, {23, 3}, {22, 3}, {20, 3}, {19, 3}}},
		{Coord{46, 0}, []Coord{{45, 0}, {44, 0}}},
		{Coord{46, 7}, []Coord{{45, 6}, {44, 7}}},
	}
	for _, tc := range neighbors {
		if got := grid.Neighbors(tc.c); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Neighbors(%v) = %v, want %v", tc.c, got, tc.want)
		}
	}

	if got := grid.Size(); got != 24*8+23*7 {
		t.Errorf("Size() = %d, want %d", got, 24*8+23*7)
	}
}

// TestDistanceMatchesBFS checks that the cube distance equals the breadth-first
// step count over the Python adjacency, which is what gui.py calculateDistance
// computes.
func TestDistanceMatchesBFS(t *testing.T) {
	for name, lv := range loadGolden(t) {
		if strings.HasPrefix(name, "BACK-") {
			continue // same size as PULSE
		}
		adj := make(map[Coord][]Coord, len(lv.Hexes))
		for _, h := range lv.Hexes {
			for _, n := range h.Neighbors {
				adj[coord(h.Hex)] = append(adj[coord(h.Hex)], coord(n))
			}
		}
		for from := range adj {
			dist := map[Coord]int{from: 0}
			queue := []Coord{from}
			for len(queue) > 0 {
				c := queue[0]
				queue = queue[1:]
				for _, n := range adj[c] {
					if _, seen := dist[n]; !seen {
						dist[n] = dist[c] + 1
						queue = append(queue, n)
					}
				}
			}
			if len(dist) != len(adj) {
				t.Fatalf("%s: only %d of %d hexes reachable from %v", name, len(dist), len(adj), from)
			}
			for to, want := range dist {
				if got := Distance(from, to); got != want {
					t.Errorf("%s: Distance(%v, %v) = %d, BFS %d", name, from, to, got, want)
				}
			}
		}
	}
}

func TestCubeInvariant(t *testing.T) {
	grid := Grid{Width: 12, Lines: 47}
	for line := range grid.Lines {
		for col := range grid.Width {
			c := Coord{line, col}
			if !grid.Contains(c) {
				continue
			}
			q, r, s := Cube(c)
			if q+r+s != 0 {
				t.Errorf("Cube(%v) = (%d, %d, %d), sum not 0", c, q, r, s)
			}
			for _, n := range grid.Neighbors(c) {
				if d := Distance(c, n); d != 1 {
					t.Errorf("Distance(%v, %v) = %d, want 1", c, n, d)
				}
			}
		}
	}
}

func TestAroundIsGeometric(t *testing.T) {
	// Around lists all six neighbours, on the map or not, in the order NE, SE, S,
	// SW, NW, N; their centres lie in exactly those directions.
	want := [6][2]float64{{71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}}
	for _, c := range []Coord{{0, 0}, {1, 0}, {4, 3}, {7, 5}} {
		cx, cy := Center(c)
		for i, n := range Around(c) {
			nx, ny := Center(n)
			if nx-cx != want[i][0] || ny-cy != want[i][1] {
				t.Errorf("Around(%v)[%d] = %v, offset (%v,%v), want %v", c, i, n, nx-cx, ny-cy, want[i])
			}
		}
	}
	g := Grid{Width: 3, Lines: 4}
	var on []Coord
	for _, n := range Around(Coord{0, 0}) {
		if g.Contains(n) {
			on = append(on, n)
		}
	}
	if got := g.Neighbors(Coord{0, 0}); !slices.Equal(got, on) {
		t.Errorf("Neighbors = %v, want the on-map part of Around %v", got, on)
	}
}
