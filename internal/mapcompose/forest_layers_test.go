package mapcompose

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestForestContextStandingAndShadowSelection(t *testing.T) {
	g, at := grid("forest")
	d, _ := Descriptor(Forest, "forest", [6]string{"forest", "forest", "forest", "forest", "forest", "forest"}, [2]int{})
	ctx := ContextID(Forest, Key(d))
	crown, shadow := "standing/"+ctx, "standing/shadow/"+ctx
	for _, tc := range []struct {
		name   string
		ids    []string
		want   string
		shadow string
	}{
		{"split context", []string{ctx, crown, shadow, "standing/forest"}, crown, shadow},
		{"legacy standing", []string{ctx, "standing/forest"}, "standing/forest", ""},
		{"legacy shadow", []string{"standing/forest", "standing/shadow/forest"}, "standing/forest", "standing/shadow/forest"},
		{"unselected context", []string{crown, shadow, "standing/forest"}, "standing/forest", ""},
		{"orphan shadow", []string{ctx, shadow}, "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := Compose(g, at, func(id string) bool { return slices.Contains(tc.ids, id) })
			c := s.Cells[0][0]
			var want, shadows []string
			if tc.want != "" {
				want = []string{tc.want}
			}
			if tc.shadow != "" {
				shadows = []string{tc.shadow}
			}
			if !slices.Equal(c.Standing, want) || !slices.Equal(c.Shadows, shadows) {
				t.Fatalf("standing %v shadows %v; want %v %v", c.Standing, c.Shadows, want, shadows)
			}
		})
	}
}

func TestGrassTreePropsAvoidNeighbourBuildingsAndInfrastructure(t *testing.T) {
	g := hexgrid.Grid{Width: 10, Lines: 10}
	var target hexgrid.Coord
	found := false
	for line := 1; line < 9 && !found; line++ {
		for col := 1; col < 8; col++ {
			if TreePropID(line, col) != "" {
				target, found = hexgrid.Coord{Line: line, Col: col}, true
				break
			}
		}
	}
	if !found {
		t.Fatal("fixture has no prop")
	}
	neighbour := hexgrid.Around(target)[0]
	for _, tc := range []struct {
		name, terrain string
		infra         []string
		want          bool
	}{
		{"meadow", "grass", nil, true},
		{"forest edge", "forest", nil, true},
		{"HQ", "hqN", nil, false},
		{"house", "house", nil, false},
		{"water", "water", nil, false},
		{"road", "grass", []string{"road25"}, false},
		{"forest road", "forest", []string{"path25"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := Compose(g, func(c hexgrid.Coord) Hex {
				if c == neighbour {
					return Hex{Terrain: tc.terrain, Infra: tc.infra}
				}
				return Hex{Terrain: "grass"}
			}, func(id string) bool { return strings.HasPrefix(id, "standing/props/trees/") })
			if got := len(s.Cells[target.Line][target.Col].Standing) > 0; got != tc.want {
				t.Errorf("tree beside %s: %v want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestGrassTreePropsAreSparseStableAndAvoidInfrastructure(t *testing.T) {
	g := hexgrid.Grid{Width: 30, Lines: 30}
	have := func(id string) bool {
		for i := 1; i <= 4; i++ {
			if id == fmt.Sprintf("standing/props/trees/%d", i) || id == fmt.Sprintf("standing/shadow/props/trees/%d", i) {
				return true
			}
		}
		return false
	}
	at := func(hexgrid.Coord) Hex { return Hex{Terrain: "grass"} }
	a, b := Compose(g, at, have), Compose(g, at, have)
	count, variants := 0, map[string]bool{}
	var occupied hexgrid.Coord
	for line, cells := range a.Cells {
		for col, cell := range cells {
			if !slices.Equal(cell.Standing, b.Cells[line][col].Standing) {
				t.Fatal("props changed on recomposition")
			}
			if len(cell.Standing) == 0 {
				continue
			}
			count++
			occupied = hexgrid.Coord{Line: line, Col: col}
			variants[cell.Standing[0]] = true
			if !strings.HasPrefix(cell.Standing[0], "standing/props/trees/") || len(cell.Shadows) != 1 {
				t.Fatalf("unpaired prop %+v", cell)
			}
		}
	}
	if count < 50 || count > 200 || len(variants) != 4 {
		t.Fatalf("props %d variants %v", count, variants)
	}
	for _, terrain := range []string{"grass", "grass_w", "forest", "house"} {
		for _, infra := range [][]string{nil, {}, {"road25"}} {
			if terrain == "grass" && infra == nil {
				continue
			}
			s := Compose(g, func(c hexgrid.Coord) Hex {
				if c == occupied {
					return Hex{Terrain: terrain, Infra: infra}
				}
				return at(c)
			}, have)
			if got := s.Cells[occupied.Line][occupied.Col].Standing; len(got) != 0 {
				t.Errorf("%s infra %v gets prop %v", terrain, infra, got)
			}
		}
	}
	if s := Compose(g, at, none); len(s.Standing()) != 0 {
		t.Fatal("old catalog acquired props")
	}
}

func TestShadowsFollowSelectedStandingUnderCluster(t *testing.T) {
	s := Scene{Cells: [][]Cell{{{Standing: []string{"standing/mud", "standing/bridge25"}, Shadows: []string{"standing/shadow/mud", "standing/shadow/bridge25"}, Cluster: 1, Under: "tile/grass"}}}, Clusters: []Cluster{{Cells: []hexgrid.Coord{{}}, Stamp: "stamp/crater/test/1"}}}
	for _, known := range []bool{false, true} {
		got := s.ShadowsAt(hexgrid.Coord{}, func(*Cluster) bool { return known })
		want := []string{"standing/shadow/mud", "standing/shadow/bridge25"}
		if known {
			want = want[1:]
		}
		if !slices.Equal(got, want) {
			t.Errorf("known %v shadows %v want %v", known, got, want)
		}
	}
}
