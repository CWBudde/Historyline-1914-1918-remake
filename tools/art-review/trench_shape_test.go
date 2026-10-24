package main

import (
	"io/fs"
	"math"
	"slices"
	"sort"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// campaignTrenches collects the derived trench tiles the shipped and campaign
// levels select that has lacks, by ID.
func campaignTrenches(has func(string) bool) (map[string]mapcompose.Gap, error) {
	out := map[string]mapcompose.Gap{}
	for _, source := range []struct {
		fsys    fs.FS
		pattern string
	}{{levels.FS, "*.json"}, {levels.Campaign, "campaign/*.json"}} {
		names, err := fs.Glob(source.fsys, source.pattern)
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			b, err := fs.ReadFile(source.fsys, name)
			if err != nil {
				return nil, err
			}
			l, err := level.Parse(b)
			if err != nil {
				return nil, err
			}
			s := mapcompose.Compose(l.Grid(), func(c hexgrid.Coord) mapcompose.Hex {
				h := l.Lines[c.Line][c.Col]
				return mapcompose.Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
			}, has)
			for _, g := range s.Gaps {
				if g.Family == mapcompose.Trench {
					out[g.Descriptor] = g
				}
			}
		}
	}
	return out, nil
}

// derivedTrenchShapes are the shapes the levels need beyond the roster.
func derivedTrenchShapes(t *testing.T) []mapcompose.TrenchShape {
	t.Helper()
	gaps, err := campaignTrenches(func(string) bool { return false })
	if err != nil {
		t.Fatal(err)
	}
	if len(gaps) < 30 {
		t.Fatalf("only %d derived trenches found", len(gaps))
	}
	ids := make([]string, 0, len(gaps))
	for id := range gaps {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var shapes []mapcompose.TrenchShape
	for _, id := range ids {
		s, ok := mapcompose.ParseTrench(id)
		if !ok {
			t.Fatalf("unparsable trench %s", id)
		}
		shapes = append(shapes, s)
	}
	return shapes
}

func trenchFixtureInputs(t *testing.T) (connectionInputs, connectionFamily) {
	t.Helper()
	dir, m := trenchTestFixture(t)
	in, err := loadConnectionInputs(writeConnectedManifest(t, dir, m))
	if err != nil {
		t.Fatal(err)
	}
	return in, in.manifest.Families[0]
}

func TestDerivedTrenchVariantsKeepRosterDecorations(t *testing.T) {
	_, f := trenchFixtureInputs(t)
	for _, tt := range []struct {
		id     string
		ports  []int
		spurs  []int
		dugout int
		ring   bool
	}{
		{id: "trench12-s4", ports: []int{1, 2}, spurs: []int{4}, dugout: 5}, // trench124's dugout
		{id: "trench1-s5", ports: []int{1}, spurs: []int{5, 3}, dugout: 4},  // trench15's spur and dugout
		{id: "trench1245", ports: []int{1, 2, 4, 5}, spurs: []int{}},        // no roster piece
		{id: "trench2-s45", ports: []int{2}, spurs: []int{4, 5}},            // trench245: none
		{id: "trench4-s2", ports: []int{4}, spurs: []int{2}, dugout: 5},     // trench24's dugout
		{id: "trench0-45", ports: []int{4, 5}, ring: true},                  // a position
		{id: "trench-s14", ports: []int{}, spurs: []int{1, 4}},              // a lone run
		{id: "trench24-s1", ports: []int{2, 4}, spurs: []int{1}, dugout: 5}, // trench124's dugout
	} {
		s, ok := mapcompose.ParseTrench(tt.id)
		if !ok {
			t.Fatalf("ParseTrench(%s)", tt.id)
		}
		v := derivedTrenchVariant(f.Variants, s)
		if v.ID != tt.id || v.ClosedRing != tt.ring || !slices.Equal(v.Ports, tt.ports) || !slices.Equal(v.BlindSpurs, tt.spurs) || v.DugoutDirection != tt.dugout {
			t.Errorf("%s: %+v", tt.id, v)
		}
		if !v.ClosedRing && len(v.Ports) > 0 {
			if err := validateConnectionTrenchVariant("trench", v); err != nil {
				t.Errorf("%s: %v", tt.id, err)
			}
		}
	}
}

// Two derived tiles that both link an edge paint the same colours through
// its band, so the derived trenches join each other and the roster.
func TestDerivedTrenchesJoinAcrossEdges(t *testing.T) {
	in, f := trenchFixtureInputs(t)
	checkDerivedTrenchesJoin(t, in, f)
}

func checkDerivedTrenchesJoin(t *testing.T, in connectionInputs, f connectionFamily) {
	t.Helper()
	g := in.manifest.Geometry
	var variants []connectionVariant
	for _, s := range derivedTrenchShapes(t) {
		variants = append(variants, derivedTrenchVariant(f.Variants, s))
	}
	variants = append(variants, f.Variants...)
	routes := make([]connectionTrenchGeometry, len(variants))
	for i, v := range variants {
		routes[i] = connectionTrenchRoutes(g, f, v)
	}
	pairs := 0
	for i, a := range variants {
		for j, b := range variants {
			for port := 1; port <= 6; port++ {
				opposite := (port+2)%6 + 1
				if !slices.Contains(a.Ports, port) || !slices.Contains(b.Ports, opposite) {
					continue
				}
				pairs++
				p0, p1 := g.Polygon[port%6], g.Polygon[(port+1)%6]
				edge := p1.sub(p0)
				tangent := edge.mul(1 / edge.length())
				normal := connectionPoint{-tangent[1], tangent[0]}
				for _, depth := range []float64{-6, -3, 0, 3, 6} {
					for _, along := range []float64{.3, .4, .45, .5, .55, .6, .7} {
						p := p0.add(edge.mul(along)).add(normal.mul(depth))
						q := p.sub(g.Neighbors[port-1])
						ca := connectionTrenchFeature(in, f, routes[i], p, .25)
						cb := connectionTrenchFeature(in, f, routes[j], q, .25)
						if ca != cb {
							t.Fatalf("%s/%s port%d depth%g t%g: %v/%v", a.ID, b.ID, port, depth, along, ca, cb)
						}
					}
				}
			}
		}
	}
	if pairs < 200 {
		t.Fatalf("only %d linked pairs", pairs)
	}
}

// A position's links run unbroken from their edge into the ring, and the
// ring keeps its open centre; blind spurs end inside the hex.
func TestDerivedTrenchTilesReachTheirEnds(t *testing.T) {
	in, f := trenchFixtureInputs(t)
	checkDerivedTrenchEnds(t, in, f)
}

func checkDerivedTrenchEnds(t *testing.T, in connectionInputs, f connectionFamily) {
	t.Helper()
	g := in.manifest.Geometry
	centre := connectionPoint{48, 40}
	for _, id := range []string{"trench0-45", "trench0-1236", "trench1-s4", "trench-s25"} {
		s, _ := mapcompose.ParseTrench(id)
		v := derivedTrenchVariant(f.Variants, s)
		tiles, err := exportConnectionTiles(in, &f, &v)
		if err != nil {
			t.Fatal(err)
		}
		tile := tiles[4]
		at := func(p connectionPoint) uint8 { return tile.NRGBAAt(int(p[0]*4), int(p[1]*4)).A }
		for _, port := range v.Ports {
			axis := g.Neighbors[port-1].mul(1 / g.Neighbors[port-1].length())
			end := centre
			if v.ClosedRing {
				end = trenchRingEntry(g.Ports[port-1].sub(axis.mul(16)))
			}
			for k := 0.; k <= 1; k += .02 {
				p := g.Ports[port-1].sub(axis.mul(1)).add(end.sub(g.Ports[port-1].sub(axis.mul(1))).mul(k))
				if a := at(p); a != 255 {
					t.Fatalf("%s: port %d course open at %v (alpha %d)", id, port, p, a)
				}
			}
		}
		for _, port := range v.BlindSpurs {
			axis := g.Neighbors[port-1].mul(1 / g.Neighbors[port-1].length())
			if at(centre.add(axis.mul(10))) != 255 {
				t.Errorf("%s: spur %d missing", id, port)
			}
			if a := at(g.Ports[port-1].sub(axis.mul(1))); a != 0 {
				t.Errorf("%s: spur %d reaches its edge (alpha %d)", id, port, a)
			}
		}
		if v.ClosedRing && at(centre) != 0 {
			t.Errorf("%s: ring centre is not open", id)
		}
	}
}

func TestTrenchRingEntryLiesOnTheRing(t *testing.T) {
	g := sharedConnectionGeometry()
	for port := 1; port <= 6; port++ {
		axis := g.Neighbors[port-1].mul(1 / g.Neighbors[port-1].length())
		guard := g.Ports[port-1].sub(axis.mul(16))
		if d := trenchRingDistance(guard); d <= 2 {
			t.Errorf("port %d: guard %v is %g from the ring", port, guard, d)
		}
		if e := trenchRingEntry(guard); math.Abs(trenchRingDistance(e)) > 1e-9 {
			t.Errorf("port %d: entry %v off the ring", port, e)
		}
	}
}
