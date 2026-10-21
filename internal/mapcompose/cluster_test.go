package mapcompose

import (
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

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
	// The variant is seeded by the first cell (0,1): 31*0+1 = 1 mod 2 → variant 2.
	cl := s.Clusters[0]
	if cl.Stamp != StampID("crater", shape, 2) || cl.Standing != StandingID(cl.Stamp) {
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
