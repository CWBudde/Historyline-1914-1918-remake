package mapcompose

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// pythonPulse is the Python remake's PULSE (internal/level's fixture), the
// map of the PULSE reference design (art/reference/pulse).
const pythonPulse = "../level/testdata/pulse"

// loadLevel is a setup of a shipped scenario ("monthly/1914-08-pulse:cp"), or
// a level file of the legacy format (pythonPulse), as a grid and its hexes.
func loadLevel(t *testing.T, name string) (hexgrid.Grid, func(hexgrid.Coord) Hex) {
	t.Helper()
	var l *level.Level
	if strings.Contains(name, ":") {
		l = campaignLevel(t, name)
	} else {
		var err error
		if l, err = level.Load(name + ".json"); err != nil {
			t.Fatal(err)
		}
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
	l, err := campaignLevelErr("monthly/1916-02-parts:cp")
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
	g, at := grid("grass_w grass_w grass_w", "grass_w mountW_w", "grass_w grass_w grass_w")
	s := Compose(g, at, none)
	if got := s.Cells[1][1].Background; got != "tile/mountain_w" {
		t.Errorf("orphan quarter drawn as %s", got)
	}
	if len(s.Issues) != 0 {
		t.Errorf("compound issues: %v", s.Issues)
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
		pythonPulse:                "forest/selected:81 ground/selected:24 hills/selected:51",
		"monthly/1914-08-pulse:cp": "forest/selected:76 ground/selected:29 hills/selected:53 shore/missing:3 trench/missing:3",
		"monthly/1914-10-civil:cp": "forest/selected:168 ground/selected:64 trench/missing:9",
		"monthly/1914-12-mouse:cp": "forest/selected:97 ground/selected:86 trench/missing:3",
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

// Roads and level crossings carry no season in the levels: on a winter map they
// draw their winter form where the catalog has it, on a summer map never.
func TestComposeWinterRoads(t *testing.T) {
	winterTiles := func(id string) bool {
		return strings.HasPrefix(id, "tile/road") || strings.HasPrefix(id, "tile/cross")
	}
	roads := func(s Scene) (summer, winter int) {
		for _, line := range s.Cells {
			for _, c := range line {
				for _, id := range c.Overlays {
					if !strings.HasPrefix(id, "tile/road") && !strings.HasPrefix(id, "tile/cross") {
						continue
					}
					if strings.HasSuffix(id, "_w") {
						winter++
					} else {
						summer++
					}
				}
			}
		}
		return summer, winter
	}
	load := func(name string) (hexgrid.Grid, func(hexgrid.Coord) Hex) {
		l := campaignLevel(t, name)
		return l.Grid(), func(c hexgrid.Coord) Hex {
			h := l.Lines[c.Line][c.Col]
			return Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
		}
	}
	g, at := load("monthly/1916-02-parts:cp")
	if summer, winter := roads(Compose(g, at, winterTiles)); summer != 0 || winter == 0 {
		t.Errorf("winter map with winter roads: %d summer, %d winter road overlays", summer, winter)
	}
	if summer, winter := roads(Compose(g, at, func(string) bool { return false })); summer == 0 || winter != 0 {
		t.Errorf("winter map without winter roads: %d summer, %d winter road overlays", summer, winter)
	}
	g, at = load("monthly/1914-08-pulse:cp")
	if summer, winter := roads(Compose(g, at, winterTiles)); summer == 0 || winter != 0 {
		t.Errorf("summer map: %d summer, %d winter road overlays", summer, winter)
	}
	for _, tc := range []struct {
		id     string
		winter bool
		want   string
	}{
		{"road36", true, "road36_w"},
		{"cross2", true, "cross2_w"},
		{"road36", false, "road36"},
		{"rail36", true, "rail36_w"},
		{"path14", true, "path14_w"},
		{"trench14-s2", true, "trench14-s2_w"},
		{"harbourN_w", true, "harbourN_w"},
		{"stream25", true, "stream25"}, // no winter form in the catalog
	} {
		has := func(id string) bool { return id != TileID("stream25_w") }
		if got := seasonalOverlay(tc.id, tc.winter, has); got != tc.want {
			t.Errorf("seasonalOverlay(%s, %v) = %s, want %s", tc.id, tc.winter, got, tc.want)
		}
	}
}
