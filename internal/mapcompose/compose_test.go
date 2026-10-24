package mapcompose

import (
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

func loadLevel(t *testing.T, name string) (hexgrid.Grid, func(hexgrid.Coord) Hex) {
	t.Helper()
	b, err := fs.ReadFile(levels.FS, name+".json")
	if err != nil {
		t.Fatal(err)
	}
	l, err := level.Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	return l.Grid(), func(c hexgrid.Coord) Hex {
		h := l.Lines[c.Line][c.Col]
		return Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
	}
}

// grid builds a small map from rows of terrain IDs (odd rows one shorter).
func grid(rows ...string) (hexgrid.Grid, func(hexgrid.Coord) Hex) {
	var lines [][]string
	for _, r := range rows {
		lines = append(lines, strings.Fields(r))
	}
	g := hexgrid.Grid{Width: len(lines[0]), Lines: len(lines)}
	return g, func(c hexgrid.Coord) Hex { return Hex{Terrain: lines[c.Line][c.Col]} }
}

func none(string) bool { return false }

func TestOverlaysKeepStoredOrder(t *testing.T) {
	g := hexgrid.Grid{Width: 1, Lines: 1}
	s := Compose(g, func(hexgrid.Coord) Hex {
		return Hex{Terrain: "water", Infra: []string{"stream25", "bridge25", "road25"}}
	}, none)
	c := s.Cells[0][0]
	if c.Background != "tile/water" || !slices.Equal(c.Overlays, []string{"tile/stream25", "tile/bridge25", "tile/road25"}) {
		t.Errorf("cell = %+v", c)
	}
}

func TestContextSelectedWhenExported(t *testing.T) {
	g, at := grid(
		"grass grass grass",
		"grass grass",
		"grass mud grass",
		"grass grass",
		"grass grass grass",
	)
	mud := hexgrid.Coord{Line: 2, Col: 1}
	d, _ := Descriptor(Ground, "mud", [6]string{"grass", "grass", "grass", "grass", "grass", "grass"}, [2]int{})
	id := ContextID(Ground, Key(d))
	s := Compose(g, at, func(x string) bool { return x == id })
	if got := s.Cells[mud.Line][mud.Col]; got.Background != id || got.Context != Ground {
		t.Errorf("mud cell = %+v, want %s", got, id)
	}
	if len(s.Gaps) != 0 {
		t.Errorf("gaps %v", s.Gaps)
	}
	// Without the export the default tile is drawn and the gap is reported.
	s = Compose(g, at, none)
	if got := s.Cells[mud.Line][mud.Col].Background; got != "tile/mud" {
		t.Errorf("fallback = %s", got)
	}
	if len(s.Gaps) != 1 || s.Gaps[0].Reason != Missing || s.Gaps[0].Descriptor != d {
		t.Errorf("gaps %+v", s.Gaps)
	}
}

func TestUndefinedContexts(t *testing.T) {
	for _, neighbor := range []string{"grass_w", "unknown"} {
		g, at := grid("forest grass", neighbor, "grass grass")
		s := Compose(g, at, none)
		if len(s.Gaps) != 1 || s.Gaps[0].Reason != Undefined || !strings.Contains(s.Gaps[0].Detail, neighbor) {
			t.Errorf("unsupported boundary %s: %+v", neighbor, s.Gaps)
		}
	}
}

func TestCompoundGroups(t *testing.T) {
	// A complete depot diamond: N, then W/E on the next line, S two lines down.
	g, at := grid(
		"grass cmpN grass",
		"cmpW cmpE",
		"grass cmpS grass",
	)
	if s := Compose(g, at, none); len(s.Issues) != 0 {
		t.Errorf("complete depot: %v", s.Issues)
	}
	g, at = grid(
		"grass cmpN grass",
		"cmpW cmpE",
		"grass grass grass",
	)
	s := Compose(g, at, none)
	if len(s.Issues) != 3 {
		t.Fatalf("broken depot: %v", s.Issues)
	}
	// The pieces still draw their own tiles; nothing is invented.
	if s.Cells[0][1].Background != "tile/cmpN" {
		t.Errorf("cmpN drawn as %s", s.Cells[0][1].Background)
	}
}

func TestDestroyedDepotDrawsRuin(t *testing.T) {
	ruins := func(id string) bool { return strings.HasPrefix(id, "tile/ruin") }
	cells := map[string]hexgrid.Coord{"N": {Line: 0, Col: 1}, "W": {Line: 1, Col: 0}, "E": {Line: 1, Col: 1}, "S": {Line: 2, Col: 1}}
	for _, sfx := range []string{"", "_w"} {
		g, plain := grid(
			"grass cmpN"+sfx+" grass",
			"cmpW"+sfx+" cmpE"+sfx,
			"grass cmpS"+sfx+" grass",
		)
		vacant := func(c hexgrid.Coord) Hex {
			h := plain(c)
			h.Vacant = c == cells["N"]
			return h
		}
		for _, tc := range []struct {
			name   string
			at     func(hexgrid.Coord) Hex
			has    func(string) bool
			prefix string
		}{
			{"destroyed", vacant, ruins, "ruin"},
			{"held", plain, ruins, "cmp"},
			{"no ruin art", vacant, none, "cmp"},
		} {
			s := Compose(g, tc.at, tc.has)
			if len(s.Issues) != 0 {
				t.Errorf("%s%s: issues %v", tc.name, sfx, s.Issues)
			}
			for p, c := range cells {
				if got, want := s.Cells[c.Line][c.Col].Background, TileID(tc.prefix+p+sfx); got != want {
					t.Errorf("%s%s: %s piece drawn as %s, want %s", tc.name, sfx, p, got, want)
				}
			}
		}
	}
}

// The destroyed depot of level 10 (PARTS) at line59[11]: its entrance has no
// content, the held depot beside it has.
func TestLevelDestroyedDepot(t *testing.T) {
	b, err := fs.ReadFile(levels.Campaign, "campaign/level10_cp.json")
	if err != nil {
		t.Fatal(err)
	}
	l, err := level.Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	s := Compose(l.Grid(), LevelHex(l), func(id string) bool { return strings.HasPrefix(id, "tile/ruin") })
	n := 0
	for line, cells := range s.Cells {
		for col, c := range cells {
			if strings.HasPrefix(c.Background, "tile/ruin") {
				n++
				if !strings.HasSuffix(c.Background, "_w") {
					t.Errorf("line%d[%d]: summer ruin %s on a winter map", line+1, col, c.Background)
				}
			}
		}
	}
	if got := s.Cells[58][11].Background; got != "tile/ruinN_w" {
		t.Errorf("line59[11] drawn as %s, want tile/ruinN_w", got)
	}
	if n != 4 {
		t.Errorf("%d ruin pieces, want one depot of 4", n)
	}
}

func TestIsolatedMountainQuarterShowsMountain(t *testing.T) {
	g, at := loadLevel(t, "level3_cp")
	s := Compose(g, at, none)
	at21 := hexgrid.Coord{Line: 21, Col: 6}
	if at(at21).Terrain != "mountW_w" {
		t.Fatalf("level 3 (21,6) is %s", at(at21).Terrain)
	}
	if got := s.Cells[21][6].Background; got != "tile/mountain_w" {
		t.Errorf("orphan quarter drawn as %s", got)
	}
	if len(s.Issues) != 0 {
		t.Errorf("level 3 compound issues: %v", s.Issues)
	}
}

// Golden gap counts on the shipped levels with no context exported, and with
// the exported profiles: they change only when the levels or the art change.
func TestShippedLevelGaps(t *testing.T) {
	exported := map[string]bool{}
	for _, f := range Families {
		for _, p := range readProfiles(t, f) {
			exported[ContextID(f, p.Key)] = true
		}
	}
	want := map[string]string{
		"level1_cp": "forest/selected:81 ground/selected:24 hills/selected:51",
		"level2_cp": "forest/selected:174 ground/selected:59",
		"level3_cp": "forest/selected:97 ground/selected:78 trench/missing:1",
	}
	for name := range want {
		g, at := loadLevel(t, name)
		s := Compose(g, at, func(id string) bool { return exported[id] })
		if len(s.Issues) != 0 {
			t.Errorf("%s compound issues: %v", name, s.Issues)
		}
		got := map[string]int{}
		for _, x := range s.Gaps {
			got[fmt.Sprintf("%s/%s", x.Family, x.Reason)]++
		}
		for _, line := range s.Cells {
			for _, c := range line {
				if c.Context != "" {
					got[fmt.Sprintf("%s/selected", c.Context)]++
				}
			}
		}
		var keys []string
		for k, n := range got {
			keys = append(keys, fmt.Sprintf("%s:%d", k, n))
		}
		slices.Sort(keys)
		if g := strings.Join(keys, " "); g != want[name] {
			t.Errorf("%s:\n got %s\nwant %s", name, g, want[name])
		}
	}
}
