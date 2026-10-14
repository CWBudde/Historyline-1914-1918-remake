package hexgrid

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// golden is testdata/adjacent.json, written by testdata/gen_adjacent.py from the
// Python adjacentHexes in hlrData.py.
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

	levels, err := filepath.Glob(filepath.Join("..", "..", "levels", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(levels) == 0 {
		t.Fatal("no shipped levels found")
	}
	for _, path := range levels {
		name := filepath.Base(path)
		lv, ok := g[name]
		if !ok {
			t.Fatalf("%s has no golden data; rerun testdata/gen_adjacent.py", name)
		}
		var file struct {
			Tiles map[string][]json.RawMessage `json:"tiles"`
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw, &file); err != nil {
			t.Fatal(err)
		}
		if lv.Width != len(file.Tiles["line1"]) || lv.Lines != len(file.Tiles) {
			t.Fatalf("%s: golden is %d×%d, level is %d×%d", name, lv.Width, lv.Lines,
				len(file.Tiles["line1"]), len(file.Tiles))
		}

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
	grid := Grid{Width: 8, Lines: 47} // level1_cp

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
			continue // same size as level1_cp
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
