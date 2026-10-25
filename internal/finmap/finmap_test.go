package finmap

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// fin builds an unpacked FIN file from terrain/occupant pairs in row-major
// order.
func fin(w, h int, cells ...[2]byte) []byte {
	b := []byte{0, byte(w), 0, byte(h)}
	for _, c := range cells {
		b = append(b, c[0], c[1])
	}
	return b
}

func TestDecodeMap(t *testing.T) {
	m, err := DecodeMap(fin(3, 2, [2]byte{1, 0xff}, [2]byte{2, 4}, [2]byte{3, 0xff}, [2]byte{4, 0xff}, [2]byte{5, 0xff}, [2]byte{6, 7}))
	if err != nil {
		t.Fatal(err)
	}
	if m.Width != 3 || m.Height != 2 {
		t.Fatalf("size %dx%d", m.Width, m.Height)
	}
	if m.TerrainAt(1, 2) != 6 || m.OccupantAt(1, 2) != 7 || m.TerrainAt(0, 1) != 2 || m.OccupantAt(0, 1) != 4 {
		t.Fatalf("cells %v %v", m.Terrain, m.Occupant)
	}
}

func TestDecodeMapRejects(t *testing.T) {
	cases := map[string][]byte{
		"short header":     {0, 1, 0},
		"zero width":       fin(0, 1),
		"cell count":       fin(2, 2, [2]byte{}, [2]byte{}, [2]byte{}),
		"trailing bytes":   append(fin(1, 1, [2]byte{}), 0),
		"high width byte":  {1, 1, 0, 1, 0, 0},
		"high height byte": {0, 1, 1, 1, 0, 0},
	}
	for name, b := range cases {
		if _, err := DecodeMap(b); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

// shp builds an SHP file: the locked types, then 12-byte records.
func shp(locked []int, recs ...[12]byte) []byte {
	b := make([]byte, TypeCount+1)
	for _, t := range locked {
		b[t] = 1
	}
	b[TypeCount] = byte(len(recs))
	for _, r := range recs {
		b = append(b, r[:]...)
	}
	return b
}

func rec(side, kind, index, energy, value byte, units ...byte) [12]byte {
	r := [12]byte{side, kind, index, energy, value, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	copy(r[5:], units)
	return r
}

func TestDecodeShop(t *testing.T) {
	s, err := DecodeShop(shp([]int{0, 50}, rec(2, 1, 3, 10, 40, 12, 0xf1, 18), rec(1, 3, 4, 0, 0)))
	if err != nil {
		t.Fatal(err)
	}
	if !s.Locked[0] || !s.Locked[50] || s.Locked[1] {
		t.Fatalf("locked %v", s.Locked)
	}
	want := Record{Side: 2, Kind: Factory, Index: 3, Energy: 10, ComputerValue: 40,
		Slots: [7]byte{12, 0xf1, 18, 0xff, 0xff, 0xff, 0xff}}
	if len(s.Records) != 2 || s.Records[0] != want {
		t.Fatalf("records %+v", s.Records)
	}
	if got := s.Records[0].Units(); !reflect.DeepEqual(got, []int{12, 18}) {
		t.Fatalf("units %v", got)
	}
	if s.Records[0].Container() != 85 || s.Records[1].Container() != 4 {
		t.Fatalf("containers %d %d", s.Records[0].Container(), s.Records[1].Container())
	}
}

func TestRecordContainer(t *testing.T) {
	for _, c := range []struct {
		kind  Kind
		index int
		want  int
	}{{HQ, 0, 70}, {HQ, 1, 71}, {HQ, 2, 72}, {Depot, 0, 72}, {Depot, 9, 81}, {Factory, 0, 82}, {Factory, 9, 91}, {Transporter, 69, 69}} {
		if got := (Record{Kind: c.kind, Index: c.index}).Container(); got != c.want {
			t.Errorf("%v %d: container %d, want %d", c.kind, c.index, got, c.want)
		}
	}
}

func TestDecodeShopRejects(t *testing.T) {
	flag2 := shp(nil)
	flag2[5] = 2
	cases := map[string][]byte{
		"short":         make([]byte, TypeCount),
		"length":        append(shp(nil, rec(0, 0, 0, 0, 0)), 0),
		"locked value":  flag2,
		"kind":          shp(nil, rec(0, 4, 0, 0, 0)),
		"side":          shp(nil, rec(3, 0, 0, 0, 0)),
		"slot type":     shp(nil, rec(0, 0, 0, 0, 0, TypeCount)),
		"factory index": shp(nil, rec(0, 1, 10, 0, 0)),
	}
	for name, b := range cases {
		if _, err := DecodeShop(b); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestDecodeComputer(t *testing.T) {
	b := make([]byte, TypeCount*6)
	b[6*50+5] = 9
	c, err := DecodeComputer(b)
	if err != nil {
		t.Fatal(err)
	}
	if c[50][5] != 9 {
		t.Fatalf("computer %v", c[50])
	}
	if _, err := DecodeComputer(b[1:]); err == nil {
		t.Fatal("expected error for a short file")
	}
}

func TestDecodeCodes(t *testing.T) {
	b := []byte{
		0x11, 0x12, 0x13, 0x2a, 0x11, 2, 1, 1, 2, 0,
		0x12, 0x13, 0x14, 0x15, 0x16, 2, 2, 0, 1, 0,
	}
	cs, err := DecodeCodes(b)
	if err != nil {
		t.Fatal(err)
	}
	want := []Code{
		{Password: "ABCZA", Raw5: 2, Players: 1, Step: 1, Winter: true},
		{Password: "BCDEF", Raw5: 2, Players: 2, Step: 0, CampaignEnd: true},
	}
	if !reflect.DeepEqual(cs, want) {
		t.Fatalf("codes %+v", cs)
	}
	for name, bad := range map[string][]byte{
		"length":  b[:9],
		"letter":  {0x10, 0x11, 0x11, 0x11, 0x11, 2, 1, 1, 0, 0},
		"players": {0x11, 0x11, 0x11, 0x11, 0x11, 2, 3, 1, 0, 0},
		"flags":   {0x11, 0x11, 0x11, 0x11, 0x11, 2, 1, 1, 4, 0},
	} {
		if _, err := DecodeCodes(bad); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

// rules for the synthetic maps: terrain 1/2 HQ of side 0/1, 3 a neutral
// factory, 4 a depot of side 0, 5 a neutral HQ; type 1 a transporter, 2/3 a ship's head and
// rear, 4 a neutral type.
var testRules = Rules{
	TerrainFlags: []int{0, 0x40 | 0x0c, 0x40 | 0x09, 0x400 | 0x0a, 0x100 | 0x0c, 0x40 | 0x0a},
	TypeClass:    append([]int{0, 0x1000, 0xc0, 0x40, 0x02}, make([]int, TypeCount-5)...),
}

func testMap() *Map {
	e := byte(0xff)
	m, err := DecodeMap(fin(4, 4,
		[2]byte{0, 1}, [2]byte{1, e}, [2]byte{0, e}, [2]byte{0, 5},
		[2]byte{0, e}, [2]byte{0, 2}, [2]byte{3, e}, [2]byte{0, 7},
		[2]byte{2, e}, [2]byte{0, e}, [2]byte{0, 6}, [2]byte{0, 0xf5},
		[2]byte{0, 8}, [2]byte{0, e}, [2]byte{0, 4}, [2]byte{4, e},
	))
	if err != nil {
		panic(err)
	}
	return m
}

func ptr(v int) *int { return &v }

func TestBuild(t *testing.T) {
	shop, err := DecodeShop(shp([]int{0, 2},
		rec(0, 0, 0, 5, 50, 0),
		rec(2, 1, 0, 10, 30, 1, 0),
		rec(0, 3, 0, 0, 0, 0, 0),
		rec(1, 3, 5, 0, 0),
		rec(0, 0, 2, 0, 0),
	))
	if err != nil {
		t.Fatal(err)
	}
	var com Computer
	com[3][1] = 7
	l, err := Build(Input{Index: 30, Map: testMap(), Shop: shop, Computer: &com,
		Code: &Code{Password: "ABCDE", Raw5: 2, Players: 1, Step: 1, Winter: true}}, testRules)
	if err != nil {
		t.Fatal(err)
	}
	if l.Format != Format || l.Version != Version || l.Width != 4 || l.Height != 4 {
		t.Fatalf("header %+v", l)
	}
	wantMission := &Mission{Code: "ABCDE", Players: 1, HumanSides: []int{1}, Winter: true, Next: ptr(31)}
	if !reflect.DeepEqual(l.Mission, wantMission) {
		t.Fatalf("mission %+v", l.Mission)
	}
	if !reflect.DeepEqual(l.Terrain[2], []int{2, 0, 0, 0}) || !reflect.DeepEqual(l.Terrain[3], []int{0, 0, 0, 4}) {
		t.Fatalf("terrain %v", l.Terrain)
	}
	wantUnits := []Unit{
		{Cell: [2]int{0, 0}, Type: 0, Side: 1},
		{Cell: [2]int{0, 3}, Type: 2, Side: 1, Rear: &[2]int{1, 3}},
		{Cell: [2]int{1, 1}, Type: 1, Side: 0, Container: ptr(0)},
		{Cell: [2]int{3, 0}, Type: 4, Side: Neutral},
		{Cell: [2]int{3, 2}, Type: 2, Side: 0, Rear: &[2]int{2, 2}},
	}
	if !reflect.DeepEqual(l.Units, wantUnits) {
		t.Fatalf("units %s", js(l.Units))
	}
	wantBuildings := []Building{
		{Kind: "hq", Cell: [2]int{0, 1}, Owner: 0, Container: 70},
		{Kind: "factory", Cell: [2]int{1, 2}, Owner: Neutral, Container: 82},
		{Kind: "hq", Cell: [2]int{2, 0}, Owner: 1, Container: 71},
		{Kind: "depot", Cell: [2]int{3, 3}, Owner: 0, Container: 72},
	}
	if !reflect.DeepEqual(l.Buildings, wantBuildings) {
		t.Fatalf("buildings %s", js(l.Buildings))
	}
	wantContainers := []Container{
		{Container: 70, Kind: "hq", Index: 0, Side: 0, Energy: 5, ComputerValue: 50, Units: []Cargo{{Type: 0}}, Holder: "building"},
		{Container: 82, Kind: "factory", Index: 0, Side: Neutral, Energy: 10, ComputerValue: 30,
			Units: []Cargo{{Type: 1, Container: ptr(1)}, {Type: 0}}, Holder: "building"},
		{Container: 0, Kind: "transporter", Index: 0, Side: 0, Units: []Cargo{{Type: 0}, {Type: 0}}, Holder: "unit"},
		{Container: 5, Kind: "transporter", Index: 5, Side: 1, Units: []Cargo{}, Holder: "none"},
		{Container: 72, Kind: "hq", Index: 2, Side: 0, Units: []Cargo{}, Holder: "building"},
	}
	if !reflect.DeepEqual(l.Containers, wantContainers) {
		t.Fatalf("containers %s", js(l.Containers))
	}
	wantProducible := []int{1}
	for ty := 3; ty < TypeCount; ty++ {
		wantProducible = append(wantProducible, ty)
	}
	if !reflect.DeepEqual(l.Producible, wantProducible) {
		t.Fatalf("producible %v", l.Producible)
	}
	if !l.Computer {
		t.Fatal("no computer file")
	}
}

func TestBuildMissionSides(t *testing.T) {
	for _, c := range []struct {
		index, players int
		want           []int
	}{{0, 1, []int{0}}, {23, 1, []int{0}}, {24, 1, []int{1}}, {48, 2, []int{0, 1}}} {
		l, err := Build(Input{Index: c.index, Map: testMap(), Code: &Code{Players: c.players}}, testRules)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(l.Mission.HumanSides, c.want) || l.Mission.Next != nil {
			t.Errorf("map %d: %+v", c.index, l.Mission)
		}
	}
}

func TestBuildWithoutShop(t *testing.T) {
	l, err := Build(Input{Index: -1, Map: testMap()}, testRules)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(l)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{`"containers":null`, `"producible":null`, `"computer":false`} {
		if !strings.Contains(string(raw), s) {
			t.Errorf("missing %s in %s", s, raw)
		}
	}
	if strings.Contains(string(raw), `"mission"`) {
		t.Errorf("unexpected mission in %s", raw)
	}
}

func TestBuildRejects(t *testing.T) {
	e := byte(0xff)
	cases := map[string]Input{
		// The rear of a side-1 ship must hold type+1 one row below.
		"missing rear":        {Map: mustMap(fin(1, 2, [2]byte{0, 5}, [2]byte{0, e}))},
		"rear side":           {Map: mustMap(fin(1, 2, [2]byte{0, 5}, [2]byte{0, 6}))},
		"stray rear":          {Map: mustMap(fin(1, 2, [2]byte{0, e}, [2]byte{0, 7}))},
		"rear outside":        {Map: mustMap(fin(1, 1, [2]byte{0, 4}))},
		"unit type":           {Map: mustMap(fin(1, 1, [2]byte{0, 2 * TypeCount}))},
		"terrain":             {Map: mustMap(fin(1, 1, [2]byte{9, e}))},
		"neutral HQ":          {Map: mustMap(fin(1, 1, [2]byte{5, e}))},
		"second HQ":           {Map: mustMap(fin(2, 1, [2]byte{1, e}, [2]byte{1, e}))},
		"code, no map number": {Index: -1, Map: mustMap(fin(1, 1, [2]byte{0, e})), Code: &Code{Players: 1}},
		"cells":               {Map: &Map{Width: 2, Height: 2}},
		"no cells":            {Map: &Map{}},
		"no map":              {},
		// 35 records of 7 units each exceed the unit records.
		"units": {Map: mustMap(fin(1, 1, [2]byte{0, e})), Shop: manyUnits(35)},
	}
	for name, in := range cases {
		if _, err := Build(in, testRules); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

// A later record for the same container replaces the earlier one's energy
// and computer value.
func TestBuildOverridden(t *testing.T) {
	shop, err := DecodeShop(shp(nil, rec(0, 0, 0, 10, 40), rec(0, 1, 0, 3, 3), rec(0, 0, 0, 0, 0)))
	if err != nil {
		t.Fatal(err)
	}
	l, err := Build(Input{Index: -1, Map: testMap(), Shop: shop}, testRules)
	if err != nil {
		t.Fatal(err)
	}
	var got []bool
	for _, c := range l.Containers {
		got = append(got, c.Overridden)
	}
	if !slices.Equal(got, []bool{true, false, false}) {
		t.Fatalf("overridden %v", got)
	}
}

// manyUnits is a shop of n depot records with seven infantry units each.
func manyUnits(n int) *Shop {
	s := &Shop{}
	for range n {
		s.Records = append(s.Records, Record{Kind: Depot, Slots: [7]byte{18, 18, 18, 18, 18, 18, 18}})
	}
	return s
}

func TestBuildUnitLimit(t *testing.T) {
	if _, err := Build(Input{Index: -1, Map: mustMap(fin(1, 1, [2]byte{0, 0xff})), Shop: manyUnits(34)}, testRules); err != nil {
		t.Fatalf("238 units: %v", err)
	}
}

// A head whose next type is not a rear has no rear to claim.
func TestBuildRearClass(t *testing.T) {
	r := Rules{TerrainFlags: testRules.TerrainFlags, TypeClass: slices.Clone(testRules.TypeClass)}
	r.TypeClass[3] = 0
	if _, err := Build(Input{Index: -1, Map: mustMap(fin(1, 2, [2]byte{0, 4}, [2]byte{0, 6}))}, r); err == nil {
		t.Fatal("expected an error for a rear type without the two-cell class")
	}
}

func TestCoord(t *testing.T) {
	for _, c := range []struct {
		row, col int
		want     hexgrid.Coord
	}{{0, 0, hexgrid.Coord{Line: 0, Col: 0}}, {0, 1, hexgrid.Coord{Line: 1, Col: 0}}, {2, 5, hexgrid.Coord{Line: 5, Col: 2}}, {3, 14, hexgrid.Coord{Line: 6, Col: 7}}} {
		if got := Coord(c.row, c.col); got != c.want {
			t.Errorf("Coord(%d, %d) = %v, want %v", c.row, c.col, got, c.want)
		}
	}
}

func mustMap(b []byte) *Map {
	m, err := DecodeMap(b)
	if err != nil {
		panic(err)
	}
	return m
}

func js(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func TestParseRejects(t *testing.T) {
	for name, raw := range map[string]string{
		"format":  `{"format":"x","version":1,"source":{},"width":1,"height":1,"terrain":[[0]]}`,
		"version": `{"format":"historyline-level","version":2,"source":{},"width":1,"height":1,"terrain":[[0]]}`,
		"rows":    `{"format":"historyline-level","version":1,"source":{},"width":1,"height":2,"terrain":[[0]]}`,
		"cells":   `{"format":"historyline-level","version":1,"source":{},"width":2,"height":1,"terrain":[[0]]}`,
		"field":   `{"format":"historyline-level","version":1,"source":{},"width":1,"height":1,"terrain":[[0]],"x":1}`,
		"trailer": `{"format":"historyline-level","version":1,"source":{},"width":1,"height":1,"terrain":[[0]]} {}`,
	} {
		if _, err := Parse([]byte(raw)); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	if _, err := Parse([]byte(`{"format":"historyline-level","version":1,"source":{},"width":1,"height":1,"terrain":[[0]]}`)); err != nil {
		t.Error(err)
	}
}
