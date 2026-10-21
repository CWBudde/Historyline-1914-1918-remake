package dosbind

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/tpwm"
)

// The local original files; the test is skipped without them.
const originalDir = "../../original/extracted/dos-full"

// deviation is a cell where a shipped level shows other terrain than the
// original map: the cell and the original record.
type deviation struct{ Cell, Record int }

// knownTerrain are the cells whose bound record follows the level's terrain,
// which differs in its rules from the original map's record there.
var knownTerrain = map[string][]deviation{
	"Pulse": {{43, 55}, {45, 80}, {85, 167}, {99, 55}, {121, 64}, {259, 149}, {275, 146}, {291, 152},
		{328, 167}, {329, 168}, {330, 169}},
	"Mouse": {{253, 0}, {379, 54}, {381, 57}, {382, 49}, {404, 55}},
}

// unitDeviation is a cell where a shipped level places another unit than
// the original map: the original and the bound occupant (2*type + side,
// 0xff none).
type unitDeviation struct{ Cell, Original, Bound int }

// knownUnits are the cells where the level's units differ from the original
// map's.
var knownUnits = map[string][]unitDeviation{
	"Pulse": {{226, 27, 23}, {277, 0xFF, 36}},
	"Civil": {{132, 0xFF, 27}, {134, 0xFF, 45}, {226, 0xFF, 27}, {230, 27, 0xFF}},
	"Mouse": {{92, 52, 0xFF}},
}

// knownContent are the building records whose units in the level differ
// from the original scenario's: the original's types by record.
var knownContent = map[string]map[int][]int{
	"Civil": {82: {21, 21, 11, 18, 18, 17}, 85: {11, 18, 18, 17}},
	"Mouse": {83: {12, 10, 18, 18, 18, 18}},
}

func unpack(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(originalDir, name))
	if err != nil {
		t.Skipf("no local original files: %v", err)
	}
	out, err := tpwm.Unpack(b, binary.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// equivalent reports whether two terrain records have the same rules in
// both seasons.
func equivalent(a, b int) bool {
	m := dosrules.Full.Tables.Move
	return m.Terrains[a] == m.Terrains[b] && m.Winter[a] == m.Winter[b]
}

// TestBindMatchesOriginalMaps checks every bound level against its original
// map (mapNo-1): the terrain record of every cell, the unit on every cell
// (the map's occupant byte is 2*type + side), and the building records of
// the scenario file (per record: side, table, index, credits, a number not
// bound, seven unit types; before them a byte per type, 0 where the
// factories build it).
func TestBindMatchesOriginalMaps(t *testing.T) {
	for _, lv := range shipped(t) {
		b := bind(t, lv)
		name := fmt.Sprintf("MAP/%02d", lv.MapNo-1)
		fin := unpack(t, name+".FIN")
		w, h := int(fin[1]), int(fin[3])
		cells := fin[4:]
		if w != b.Width || h != b.Height || len(cells) != 2*w*h {
			t.Fatalf("%s: original %d×%d, bound %d×%d", lv.MapName, w, h, b.Width, b.Height)
		}

		var dev []deviation
		var udev []unitDeviation
		occupant := map[int]int{}
		for i, u := range b.Units {
			if !held(b, i) {
				occupant[u.Cell] = 2*u.Type + u.Owner
			}
		}
		for c := range w * h {
			orig := int(cells[2*c])
			if b.Edge(c) && orig != Table.Edge {
				t.Errorf("%s: edge cell %d holds record %d", lv.MapName, c, orig)
			}
			if !equivalent(b.Terrain[c], orig) {
				dev = append(dev, deviation{c, orig})
			}
			want, ok := occupant[c]
			if !ok {
				want = 0xFF
			}
			if o := int(cells[2*c+1]); o != want {
				udev = append(udev, unitDeviation{c, o, want})
			}
		}
		if !slices.Equal(udev, knownUnits[lv.MapName]) {
			t.Errorf("%s: unit deviations %v, known %v", lv.MapName, udev, knownUnits[lv.MapName])
		}
		if !slices.Equal(dev, knownTerrain[lv.MapName]) {
			t.Errorf("%s: terrain deviations %v, known %v", lv.MapName, dev, knownTerrain[lv.MapName])
		}

		shp := unpack(t, name+".SHP")
		const types = 51
		var builds []int
		for k := range types {
			if shp[k] == 0 {
				builds = append(builds, k)
			}
		}
		credits := map[string]int{}
		bound := 0
		for k := range int(shp[types]) {
			r := shp[types+1+12*k : types+13+12*k]
			side, table, idx := int(r[0]), int(r[1]), int(r[2])
			var hold int
			switch table {
			case 0:
				hold = dosrules.HoldFlag40 + idx
			case 1:
				hold = dosrules.HoldFlag400 + idx
			case 2:
				hold = dosrules.HoldFlag100 + idx
			default:
				continue // a depot yet to be built
			}
			bound++
			credits[strconv.Itoa(hold)] = int(r[3])
			hd, ok := b.Holds[hold]
			if !ok {
				t.Errorf("%s: no building record %d", lv.MapName, hold)
				continue
			}
			if owner(hd.Flags) != side {
				t.Errorf("%s: record %d of side %d, bound %d", lv.MapName, hold, side, owner(hd.Flags))
			}
			var orig, got []int
			for _, x := range r[5:] {
				if x != 0xFF {
					orig = append(orig, int(x))
				}
			}
			for k, u := range hd.Slots[:7] {
				if u <= dosrules.SlotFree {
					got = append(got, b.Units[u].Type)
					if hd.Slots[7+k] != u {
						t.Errorf("%s: record %d: the sides' slots differ", lv.MapName, hold)
					}
				}
			}
			want := knownContent[lv.MapName][hold]
			if want == nil {
				want = orig
			} else {
				got = orig
			}
			if !sameTypes(got, want) {
				t.Errorf("%s: record %d holds types %v, original %v", lv.MapName, hold, got, orig)
			}
			bd := b.Buildings[hold]
			wantBuilds := []int(nil)
			if table == 1 {
				wantBuilds = builds
			}
			if bd.Credits != int(r[3]) || !slices.Equal(bd.Builds, wantBuilds) {
				t.Errorf("%s: record %d: credits %d builds %v, original %d %v", lv.MapName, hold, bd.Credits, bd.Builds, r[3], wantBuilds)
			}
		}
		if n := buildingHolds(b); n != bound {
			t.Errorf("%s: %d building records bound, original %d", lv.MapName, n, bound)
		}
		mr := Table.Maps[strconv.Itoa(lv.MapNo)]
		if !maps.Equal(mr.Credits, credits) || !slices.Equal(mr.Builds, builds) {
			j, _ := json.Marshal(MapRecords{Credits: credits, Builds: builds})
			t.Errorf("%s: records.json maps %d differs from the original: %s", lv.MapName, lv.MapNo, j)
		}
	}
}

// TestCarriesMatchUnitTable checks the classes a transporter takes against
// UNIT.DAT word +0x13 (records of 69 bytes).
func TestCarriesMatchUnitTable(t *testing.T) {
	u := unpack(t, "UNIT.DAT")
	for i, typ := range dosrules.Full.Tables.Move.Types {
		w := int(binary.LittleEndian.Uint16(u[69*i+0x13:]))
		got, ok := Table.Carries[strconv.Itoa(i)]
		if typ.Class&transporter != 0 && (!ok || got != w) {
			t.Errorf("type %d: carries %d (%v), unit table %d", i, got, ok, w)
		}
		if typ.Class&transporter == 0 && ok {
			t.Errorf("type %d carries but is no transporter", i)
		}
	}
}

func held(b *Board, unit int) bool {
	for _, h := range b.Holds {
		if h.Flags&carrier == 0 && slices.Contains(h.Slots[:], unit) {
			return true
		}
	}
	return false
}

func buildingHolds(b *Board) int {
	n := 0
	for i := range b.Holds {
		if i >= dosrules.HoldFlag40 {
			n++
		}
	}
	return n
}

func sameTypes(a, b []int) bool {
	a, b = slices.Clone(a), slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}
