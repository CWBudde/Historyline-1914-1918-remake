package origview

import (
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

func TestEveryRecordHasALook(t *testing.T) {
	if len(Looks) != Records {
		t.Errorf("%d looks, want %d", len(Looks), Records)
	}
	for r := range Records {
		if k, ok := Looks[r]; !ok || k.Terrain == "" {
			t.Errorf("record %d has no look", r)
		}
	}
}

func TestLooksAreTerrains(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	for r, k := range Looks {
		for _, winter := range []bool{false, true} {
			if g := k.Ground(winter); !hasTerrain(set, g) {
				t.Errorf("record %d: terrain %q (winter %v) is not in terrain.json", r, g, winter)
			}
		}
		if _, ok := set.Terrain[k.Terrain+"_w"]; ok && k.Winter == "" && k.Ground(true) != k.Terrain+"_w" {
			t.Errorf("record %d: winter %q, terrain.json has %s_w", r, k.Ground(true), k.Terrain)
		}
		for _, id := range k.Infra {
			if id == "" || strings.ContainsAny(id, `"'\, `) {
				t.Errorf("record %d: infra %q", r, id)
			}
		}
	}
}

func TestFactoryLooksUseSeasonalTiles(t *testing.T) {
	for _, tc := range []struct {
		record  int
		terrain string
	}{
		{9, "factoryN"},
		{10, "factoryS"},
		{11, "factoryW"},
		{12, "factoryE"},
		{13, "factoryE"},
		{14, "factoryE"},
	} {
		for _, winter := range []bool{false, true} {
			want := tc.terrain
			if winter {
				want += "_w"
			}
			if got := Looks[tc.record].Hex(winter); got.Terrain != want || len(got.Infra.IDs) != 0 {
				t.Errorf("record %d (winter %v): %+v, want terrain %s without overlays", tc.record, winter, got, want)
			}
		}
	}
}

func hasTerrain(set *data.Set, name string) bool {
	_, ok := set.Terrain[name]
	return ok
}

// terrainCatalog reports the IDs the runtime terrain catalog has at every
// density the game draws (the terrain-* packs of assets/index.json).
func terrainCatalog(t *testing.T) func(string) bool {
	t.Helper()
	raw, err := os.ReadFile("../../assets/index.json")
	if err != nil {
		t.Fatal(err)
	}
	var index struct {
		Assets []struct {
			ID      string `json:"id"`
			Pack    string `json:"pack"`
			Density int    `json:"density"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(raw, &index); err != nil {
		t.Fatal(err)
	}
	at := map[string]map[int]bool{}
	for _, e := range index.Assets {
		if !strings.HasPrefix(e.Pack, "terrain-") {
			continue
		}
		if at[e.ID] == nil {
			at[e.ID] = map[int]bool{}
		}
		at[e.ID][e.Density] = true
	}
	return func(id string) bool { return at[id][2] && at[id][4] }
}

func TestLooksHaveTiles(t *testing.T) {
	has := terrainCatalog(t)
	for r, k := range Looks {
		for _, winter := range []bool{false, true} {
			h := k.Hex(winter)
			ground := h.Terrain
			if mapcompose.OpenWater(ground) {
				ground = "water"
			}
			for _, id := range append([]string{ground}, h.Infra.IDs...) {
				if !has(mapcompose.TileID(id)) {
					t.Errorf("record %d (winter %v): the terrain catalog has no %s", r, winter, mapcompose.TileID(id))
				}
			}
		}
	}
}

// key is dosbind.Key: the terrain without the winter suffix and the infra ids.
func key(h level.Hex) string {
	return strings.Join(append([]string{strings.TrimSuffix(h.Terrain, "_w")}, h.Infra.IDs...), "+")
}

// Wherever the binding table names a record, the look shows one of its keys.
func TestLooksAgreeWithRecords(t *testing.T) {
	raw, err := os.ReadFile("../dosbind/records.json")
	if err != nil {
		t.Fatal(err)
	}
	var table struct {
		Terrain map[string]int `json:"terrain"`
		Maps    map[string]struct {
			Terrain map[string]int `json:"terrain"`
		} `json:"maps"`
	}
	if err := json.Unmarshal(raw, &table); err != nil {
		t.Fatal(err)
	}
	keys := map[int][]string{}
	add := func(m map[string]int) {
		for k, r := range m {
			k, _, _ = strings.Cut(k, "@")
			keys[r] = append(keys[r], k)
		}
	}
	add(table.Terrain)
	for _, m := range table.Maps {
		add(m.Terrain)
	}
	if len(keys) < 100 {
		t.Fatalf("records.json names %d records", len(keys))
	}
	for r, ks := range keys {
		k := Looks[r]
		if s, w := key(k.Hex(false)), key(k.Hex(true)); !slices.Contains(ks, s) && !slices.Contains(ks, w) {
			t.Errorf("record %d looks like %s / %s, records.json names it %v", r, s, w, ks)
		}
	}
}

func TestLevel(t *testing.T) {
	src := &finmap.Level{Format: finmap.Format, Version: finmap.Version,
		Source: finmap.Source{Build: "full", Map: "07"}, Width: 4, Height: 3,
		Terrain:   [][]int{{1, 57, 36, Edge}, {25, 81, 0, Edge}, {Edge, Edge, Edge, Edge}},
		Buildings: []finmap.Building{{Kind: "hq", Cell: [2]int{0, 0}, Owner: finmap.Allied}},
		Mission:   &finmap.Mission{Players: 1, HumanSides: []int{finmap.Allied}, Winter: true}}
	got, err := Level(src)
	if err != nil {
		t.Fatal(err)
	}
	hex := func(terrain string, infra ...string) level.Hex {
		return level.Hex{Terrain: terrain, Infra: level.Infra{IDs: infra}}
	}
	hq := hex("hqN_w")
	hq.Content = &level.Content{Owner: int(game.Entente)}
	want := &level.Level{MapName: "07", MapNo: 8, Player: game.Entente.String(), Lines: [][]level.Hex{
		{hq, hex("grass_w", "stream25")},
		{hex("forest_w")},
		{hex("water"), hex("grass_w")},
		{hex("shelled_w")},
		{hex("water"), hex("grass_w")},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Level =\n%+v\nwant\n%+v", got, want)
	}

	src.Mission, src.Source.Map = nil, "STATMAP"
	got, err = Level(src)
	if err != nil {
		t.Fatal(err)
	}
	if got.MapName != "STATMAP" || got.MapNo != -1 || got.Player != game.CentralPowers.String() ||
		got.Lines[0][0].Terrain != "hqN" || got.Lines[3][0].Terrain != "mud" {
		t.Errorf("a map without a mission: %s %d %q %+v", got.MapName, got.MapNo, got.Player, got.Lines)
	}

	src.Terrain[1][1] = Records
	if _, err := Level(src); err == nil {
		t.Error("an unknown record gave no error")
	}
}

func TestLevelInfraForms(t *testing.T) {
	src := &finmap.Level{Source: finmap.Source{Map: "00"}, Width: 4, Height: 2,
		Terrain: [][]int{{139, 0, 0, Edge}, {Edge, Edge, Edge, Edge}}}
	got, err := Level(src)
	if err != nil {
		t.Fatal(err)
	}
	if h := got.Lines[0][0]; !h.Infra.List || !reflect.DeepEqual(h.Infra.IDs, []string{"road25", "road24"}) {
		t.Errorf("two infra ids: %+v", h.Infra)
	}
	src.Terrain[0][0] = 127
	if got, err = Level(src); err != nil {
		t.Fatal(err)
	}
	if h := got.Lines[0][0]; h.Infra.List || !reflect.DeepEqual(h.Infra.IDs, []string{"road36"}) {
		t.Errorf("one infra id: %+v", h.Infra)
	}
}

// The committed original maps: the view has the shipped level shape, survives
// a save, shows each cell's look, puts the owners on the building entrances,
// and composes with the runtime terrain catalog.
func TestCommittedMaps(t *testing.T) {
	names, err := fs.Glob(levels.Original, "original/full/*.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 73 {
		t.Fatalf("%d committed maps", len(names))
	}
	has := terrainCatalog(t)
	var missing, undefined int
	for _, name := range names {
		raw, err := fs.ReadFile(levels.Original, name)
		if err != nil {
			t.Fatal(err)
		}
		src, err := finmap.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		lv, err := Level(src)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		stem := strings.TrimSuffix(path.Base(name), ".json")
		checkShape(t, stem, src, lv)
		checkCells(t, stem, src, lv)

		out, err := lv.Marshal()
		if err != nil {
			t.Fatalf("%s: %v", stem, err)
		}
		back, err := level.Parse(out)
		if err != nil {
			t.Fatalf("%s: %v", stem, err)
		}
		if !reflect.DeepEqual(back, lv) {
			t.Errorf("%s: the view changes when saved and read again", stem)
		}

		s := mapcompose.Compose(lv.Grid(), func(c hexgrid.Coord) mapcompose.Hex {
			h := lv.Lines[c.Line][c.Col]
			return mapcompose.Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
		}, has)
		ids := []string{}
		for _, line := range s.Cells {
			for _, c := range line {
				ids = slices.Concat(ids, []string{c.Background, c.Under}, c.Overlays, c.Standing)
			}
		}
		for _, cl := range s.Clusters {
			ids = append(ids, cl.Stamp, cl.Standing)
		}
		for _, id := range ids {
			if id != "" && !has(id) {
				t.Errorf("%s: the terrain catalog has no %s", stem, id)
			}
		}
		for _, is := range s.Issues {
			// STATMAP shows the building entrances alone.
			if stem == "STATMAP" {
				t.Logf("%s: %v: %s", stem, is.At, is.Detail)
			} else {
				t.Errorf("%s: %v: %s", stem, is.At, is.Detail)
			}
		}
		var m, u int
		for _, g := range s.Gaps {
			switch g.Reason {
			case mapcompose.Missing:
				m++
			case mapcompose.Undefined:
				u++
			}
		}
		if m+u > 0 {
			if stem == "STATMAP" {
				t.Logf("%s: %d missing and %d undefined context tiles", stem, m, u)
			} else {
				t.Errorf("%s: %d missing and %d undefined context tiles", stem, m, u)
			}
		}
		missing += m
		undefined += u
	}
	t.Logf("all maps: %d missing and %d undefined context tiles", missing, undefined)
}

func checkShape(t *testing.T, stem string, src *finmap.Level, lv *level.Level) {
	t.Helper()
	if lv.MapName != stem {
		t.Errorf("%s: MapName %q", stem, lv.MapName)
	}
	if no, err := strconv.Atoi(stem); err == nil && lv.MapNo != no+1 || err != nil && lv.MapNo != -1 {
		t.Errorf("%s: MapNo %d", stem, lv.MapNo)
	}
	player := game.CentralPowers
	if src.Mission != nil && src.Mission.HumanSides[0] == finmap.Allied {
		player = game.Entente
	}
	if p, err := game.PlayerByName(lv.Player); err != nil || p != player {
		t.Errorf("%s: Player %q, want %v", stem, lv.Player, player)
	}
	if len(lv.Lines) != 2*src.Height-1 {
		t.Fatalf("%s: %d lines for %d rows", stem, len(lv.Lines), src.Height)
	}
	for l, line := range lv.Lines {
		if want := src.Width/2 - l%2; len(line) != want {
			t.Errorf("%s: line %d has %d hexes, want %d", stem, l, len(line), want)
		}
	}
}

func checkCells(t *testing.T, stem string, src *finmap.Level, lv *level.Level) {
	t.Helper()
	winter := src.Mission != nil && src.Mission.Winter
	g := lv.Grid()
	owners := map[hexgrid.Coord]int{}
	for _, b := range src.Buildings {
		owners[finmap.Coord(b.Cell[0], b.Cell[1])] = int([]game.Player{game.CentralPowers, game.Entente, game.None}[b.Owner])
	}
	for row := range src.Height {
		for col := range src.Width {
			c := finmap.Coord(row, col)
			if !g.Contains(c) {
				continue
			}
			h := lv.Lines[c.Line][c.Col]
			if h.Unit != "" {
				t.Errorf("%s: %v holds unit %q", stem, c, h.Unit)
			}
			want := Looks[src.Terrain[row][col]].Hex(winter)
			if src.Terrain[row][col] == Edge {
				// A neighbour's look, or the edge's own (grass) among buildings.
				want = Looks[Edge].Hex(winter)
				for _, n := range g.Neighbors(c) {
					nh := lv.Lines[n.Line][n.Col]
					if nh.Terrain == h.Terrain && slices.Equal(nh.Infra.IDs, h.Infra.IDs) && !building(nh.Terrain) {
						want = h
					}
				}
			}
			if o, ok := owners[c]; ok {
				want.Content = &level.Content{Owner: o}
				if e := strings.TrimSuffix(h.Terrain, "_w"); e != "hqN" && e != "cmpN" && e != "factoryE" {
					t.Errorf("%s: building entrance %v shows %s", stem, c, h.Terrain)
				}
			}
			if !reflect.DeepEqual(h, want) {
				t.Errorf("%s: %v (record %d) is %+v, want %+v", stem, c, src.Terrain[row][col], h, want)
			}
		}
	}
}
