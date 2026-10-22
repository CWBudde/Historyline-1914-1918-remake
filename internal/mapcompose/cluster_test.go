package mapcompose

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// A painting fitted to PULSE's whole lake must not replace the reusable water
// and shoreline tiles in the candidate catalog used by the game.
func TestPULSELakeUsesReusableTerrainTiles(t *testing.T) {
	raw, err := os.ReadFile("../../assets/index.json")
	if err != nil {
		t.Fatal(err)
	}
	var index struct {
		Assets []struct {
			ID      string `json:"id"`
			Density int    `json:"density"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(raw, &index); err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	for _, entry := range index.Assets {
		if entry.Density == 4 {
			have[entry.ID] = true
		}
	}
	g, at := loadLevel(t, "level1_cp")
	s := Compose(g, at, func(id string) bool { return have[id] })
	for _, cell := range []struct {
		line, col  int
		background string
	}{
		{26, 6, "tile/lakeside156"},
		{27, 5, "tile/lakeside56"},
		{27, 6, "tile/lakeside16"},
		{28, 5, "tile/lakeside3456"},
		{28, 6, "tile/waterStones"},
		{28, 7, "tile/lakeside126"},
		{29, 5, "tile/lakeside34"},
		{29, 6, "tile/water"},
		{30, 6, "tile/lakeside34"},
		{30, 7, "tile/lakeside123"},
		{31, 6, "tile/lakeside23-outlet4"},
	} {
		c := hexgrid.Coord{Line: cell.line, Col: cell.col}
		bg, stamp, _, _ := s.Layers(c, func(*Cluster) bool { return true })
		if bg != cell.background || stamp != "" {
			t.Errorf("lake at %+v draws %s and stamp %q, want %s without a whole-lake painting", c, bg, stamp, cell.background)
		}
		if !have[cell.background] {
			t.Errorf("candidate catalog lacks %s", cell.background)
		}
	}
	outlet := hexgrid.Coord{Line: 32, Col: 6}
	_, _, overlays, standing := s.Layers(outlet, func(*Cluster) bool { return true })
	if !slices.Equal(overlays, []string{"tile/stream14", "tile/bridge25"}) || !slices.Contains(standing, "standing/bridge25") {
		t.Errorf("lake outlet lost its stream or bridge: overlays %v, standing %v", overlays, standing)
	}
}

func TestClusters(t *testing.T) {
	g, at := grid(
		"mud grass stone water",
		"mud grass lakeside12",
		"grass grass grass water",
		"grass shelled_w grass",
	)
	s := Compose(g, at, none)
	type want struct {
		kind, shape string
		cells       []hexgrid.Coord
	}
	got := []want{}
	for _, cl := range s.Clusters {
		got = append(got, want{cl.Kind, cl.Shape, cl.Cells})
		if cl.Stamp != "" {
			t.Errorf("%s stamped without art: %s", cl.Kind, cl.Stamp)
		}
	}
	exp := []want{
		// (0,0) and its SE neighbour (1,0).
		{"crater", "2_1.1", []hexgrid.Coord{{Line: 0, Col: 0}, {Line: 1, Col: 0}}},
		{"boulders", "1", []hexgrid.Coord{{Line: 0, Col: 2}}},
		// water (0,3), lakeside SW of it (1,2), water S of it (2,3).
		{"lake", "3_-1.1_0.2", []hexgrid.Coord{{Line: 0, Col: 3}, {Line: 1, Col: 2}, {Line: 2, Col: 3}}},
		// Seasons never share a cluster.
		{"crater_w", "1", []hexgrid.Coord{{Line: 3, Col: 1}}},
	}
	if !slices.EqualFunc(got, exp, func(a, b want) bool {
		return a.kind == b.kind && a.shape == b.shape && slices.Equal(a.cells, b.cells)
	}) {
		t.Errorf("clusters %+v\nwant %+v", got, exp)
	}
	if s.Cells[1][0].Cluster != 1 || s.Cells[0][1].Cluster != 0 {
		t.Errorf("cluster index %d, %d", s.Cells[1][0].Cluster, s.Cells[0][1].Cluster)
	}
}

func TestStampVariantsMixEvenCoordinatesAndRemainStable(t *testing.T) {
	g := hexgrid.Grid{Width: 14, Lines: 16}
	mud := map[hexgrid.Coord]bool{}
	for _, line := range []int{0, 6, 12} {
		for _, col := range []int{2, 6, 10} {
			for _, c := range []hexgrid.Coord{{Line: line, Col: col}, {Line: line + 1, Col: col - 1}, {Line: line + 1, Col: col}, {Line: line + 2, Col: col}} {
				mud[c] = true
			}
		}
	}
	at := func(c hexgrid.Coord) Hex {
		if mud[c] {
			return Hex{Terrain: "mud"}
		}
		return Hex{Terrain: "grass"}
	}
	const shape = "4_-1.1_1.1_0.2"
	one, two := StampID("crater", shape, 1), StampID("crater", shape, 2)
	has := func(id string) bool { return id == one || id == two }
	a, b := Compose(g, at, has), Compose(g, at, has)
	if !reflect.DeepEqual(a.Clusters, b.Clusters) {
		t.Fatal("stamp selection changed on recomposition")
	}
	if len(a.Clusters) != 9 {
		t.Fatalf("got %d clusters, want nine isolated diamonds", len(a.Clusters))
	}
	selected := map[string]bool{}
	for _, cl := range a.Clusters {
		if cl.Shape != shape || cl.Cells[0].Line%2 != 0 || cl.Cells[0].Col%2 != 0 || !has(cl.Stamp) {
			t.Fatalf("unexpected cluster: %+v", cl)
		}
		selected[cl.Stamp] = true
	}
	if !selected[one] || !selected[two] {
		t.Fatalf("even-coordinate clusters selected %v, want both available variants", selected)
	}
}

func TestPULSENorthAndSouthCraterVariants(t *testing.T) {
	g, at := loadLevel(t, "level1_cp")
	const shape = "4_-1.1_1.1_0.2"
	one, two := StampID("crater", shape, 1), StampID("crater", shape, 2)
	s := Compose(g, at, func(id string) bool { return id == one || id == two })
	for _, tc := range []struct {
		at    hexgrid.Coord
		stamp string
	}{
		{hexgrid.Coord{Line: 11, Col: 2}, one},
		{hexgrid.Coord{Line: 39, Col: 4}, two},
	} {
		cell := s.Cells[tc.at.Line][tc.at.Col]
		if cell.Cluster == 0 {
			t.Fatalf("crater at %+v has no cluster", tc.at)
		}
		cl := s.Clusters[cell.Cluster-1]
		if cl.Cells[0] != tc.at || cl.Shape != shape || cl.Stamp != tc.stamp {
			t.Errorf("crater at %+v: %+v, want %s", tc.at, cl, tc.stamp)
		}
	}
}

func TestStampedClusterLayers(t *testing.T) {
	g, at := grid(
		"grass mud grass",
		"mud grass",
		"grass grass grass",
	)
	atInfra := func(c hexgrid.Coord) Hex {
		h := at(c)
		if c == (hexgrid.Coord{Line: 1, Col: 0}) {
			h.Infra = []string{"bridge25"}
		}
		return h
	}
	shape := "2_-1.1"
	have := map[string]bool{
		StampID("crater", shape, 1): true, StampID("crater", shape, 2): true,
		StandingID(StampID("crater", shape, 2)): true, StandingID(StampID("crater", shape, 1)): true,
		"standing/mud": true, "standing/bridge25": true,
	}
	s := Compose(g, atInfra, func(id string) bool { return have[id] })
	if len(s.Clusters) != 1 || s.Clusters[0].Shape != shape {
		t.Fatalf("clusters %+v", s.Clusters)
	}
	// The first cell's mixed coordinate hash selects variant 1 at (0,1).
	cl := s.Clusters[0]
	if cl.Stamp != StampID("crater", shape, 1) || cl.Standing != StandingID(cl.Stamp) {
		t.Errorf("stamp %s, standing %s", cl.Stamp, cl.Standing)
	}
	first, second := hexgrid.Coord{Line: 0, Col: 1}, hexgrid.Coord{Line: 1, Col: 0}
	yes := func(*Cluster) bool { return true }
	if bg, stamp, _, st := s.Layers(first, yes); bg != "tile/grass" || stamp != cl.Stamp || !slices.Equal(st, []string{cl.Standing}) {
		t.Errorf("first cell: %s %s %v", bg, stamp, st)
	}
	// The mud's own standing item gives way; the bridge's stays.
	if bg, stamp, ov, st := s.Layers(second, yes); bg != "tile/grass" || stamp != "" || len(ov) != 1 || !slices.Equal(st, []string{"standing/bridge25"}) {
		t.Errorf("second cell: %s %s %v %v", bg, stamp, ov, st)
	}
	// Without the stamp the cells keep their per-hex art.
	no := func(*Cluster) bool { return false }
	if bg, stamp, _, st := s.Layers(second, no); bg != "tile/mud" || stamp != "" || !slices.Equal(st, []string{"standing/mud", "standing/bridge25"}) {
		t.Errorf("unstamped: %s %s %v", bg, stamp, st)
	}
}

func TestShapeKey(t *testing.T) {
	for _, c := range []struct {
		cells []hexgrid.Coord
		want  string
	}{
		{[]hexgrid.Coord{{Line: 3, Col: 2}}, "1"},
		{[]hexgrid.Coord{{Line: 3, Col: 2}, {Line: 5, Col: 2}}, "2_0.2"},
		{[]hexgrid.Coord{{Line: 2, Col: 2}, {Line: 3, Col: 2}}, "2_1.1"},
		{[]hexgrid.Coord{{Line: 3, Col: 2}, {Line: 4, Col: 3}}, "2_1.1"}, // same shape from an odd line
	} {
		if got := ShapeKey(c.cells); got != c.want {
			t.Errorf("%v: %s, want %s", c.cells, got, c.want)
		}
	}
}

// A knoll or boulder group whose shape has no stamp is tiled with the stamps the
// catalog has: the largest shape first, from the first uncovered cell; cells no
// stamp fits keep their per-hex art as unstamped clusters. Lakes and crater
// fields stay whole.
func TestTile(t *testing.T) {
	g, at := grid(
		"stone grass grass water",
		"stone grass water",
		"stone grass grass water",
		"stone grass mud",
		"stone grass mud grass",
	)
	have := map[string]bool{
		StampID("boulders", "2_0.2", 1): true, // two hexes one above the other
		StampID("boulders", "1", 1):     true,
		StampID("lake", "1", 1):         true,
		StampID("crater", "1", 1):       true,
	}
	s := Compose(g, at, func(id string) bool { return have[id] })
	type want struct {
		kind, shape, stamp string
		cells              []hexgrid.Coord
	}
	var got []want
	for _, cl := range s.Clusters {
		got = append(got, want{cl.Kind, cl.Shape, cl.Stamp, cl.Cells})
	}
	c := func(line, col int) hexgrid.Coord { return hexgrid.Coord{Line: line, Col: col} }
	exp := []want{
		// The five stones zig-zag down lines 0 to 4: two pairs one above the
		// other, then a single.
		{"boulders", "2_0.2", StampID("boulders", "2_0.2", 1), []hexgrid.Coord{c(0, 0), c(2, 0)}},
		{"boulders", "2_0.2", StampID("boulders", "2_0.2", 1), []hexgrid.Coord{c(1, 0), c(3, 0)}},
		{"boulders", "1", StampID("boulders", "1", 1), []hexgrid.Coord{c(4, 0)}},
		// The lake has no stamp for its shape and is not tiled.
		{"lake", "3_-1.1_0.2", "", []hexgrid.Coord{c(0, 3), c(1, 2), c(2, 3)}},
		// Crater fields are not tiled, though a single crater has a stamp.
		{"crater", "2_-1.1", "", []hexgrid.Coord{c(3, 2), c(4, 2)}},
	}
	if !slices.EqualFunc(got, exp, func(a, b want) bool {
		return a.kind == b.kind && a.shape == b.shape && a.stamp == b.stamp && slices.Equal(a.cells, b.cells)
	}) {
		t.Fatalf("clusters\n%+v\nwant\n%+v", got, exp)
	}
	yes := func(*Cluster) bool { return true }
	for _, x := range []struct {
		at    hexgrid.Coord
		stamp string
	}{{c(0, 0), exp[0].stamp}, {c(2, 0), ""}, {c(1, 0), exp[1].stamp}, {c(4, 0), exp[2].stamp}} {
		if bg, stamp, _, _ := s.Layers(x.at, yes); bg != "tile/grass" || stamp != x.stamp {
			t.Errorf("%v draws %s and stamp %q, want the grass base and %q", x.at, bg, stamp, x.stamp)
		}
	}
	if bg, stamp, _, _ := s.Layers(c(3, 2), yes); bg != "tile/mud" || stamp != "" {
		t.Errorf("an untiled crater draws %s and stamp %q", bg, stamp)
	}

	// A remainder no stamp fits stays unstamped and connected.
	delete(have, StampID("boulders", "1", 1))
	s = Compose(g, at, func(id string) bool { return have[id] })
	if cl := s.Clusters[2]; cl.Stamp != "" || !slices.Equal(cl.Cells, []hexgrid.Coord{c(4, 0)}) {
		t.Errorf("remainder %+v", cl)
	}
}

// connectedSets lists each connected set through the anchor once, and no set
// with a cell before the anchor: the same sets a brute force over all subsets
// finds.
func TestConnectedSets(t *testing.T) {
	g := hexgrid.Grid{Width: 4, Lines: 5}
	free := map[hexgrid.Coord]bool{}
	var all []hexgrid.Coord
	for line := range g.Lines {
		for col := range g.Width - line%2 {
			c := hexgrid.Coord{Line: line, Col: col}
			free[c], all = true, append(all, c)
		}
	}
	free[hexgrid.Coord{Line: 2, Col: 2}] = false
	anchor := hexgrid.Coord{Line: 1, Col: 1}
	const limit = 4
	got := map[string]bool{}
	for _, set := range connectedSets(g, anchor, free, limit) {
		k := fmt.Sprint(set)
		if got[k] {
			t.Errorf("%s listed twice", k)
		}
		got[k] = true
	}
	want := map[string]bool{}
	for mask := 1; mask < 1<<len(all); mask++ {
		var set []hexgrid.Coord
		for i, c := range all {
			if mask&(1<<i) != 0 {
				set = append(set, c)
			}
		}
		if len(set) > limit || !slices.Contains(set, anchor) || slices.ContainsFunc(set, func(c hexgrid.Coord) bool {
			return !free[c] || byLineCol(c, anchor) < 0
		}) || !connected(g, set) {
			continue
		}
		want[fmt.Sprint(set)] = true
	}
	if !maps.Equal(got, want) {
		t.Errorf("%d sets, want %d", len(got), len(want))
	}
}

func connected(g hexgrid.Grid, set []hexgrid.Coord) bool {
	seen := map[hexgrid.Coord]bool{set[0]: true}
	queue := []hexgrid.Coord{set[0]}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		for _, n := range g.Neighbors(c) {
			if slices.Contains(set, n) && !seen[n] {
				seen[n] = true
				queue = append(queue, n)
			}
		}
	}
	return len(seen) == len(set)
}
