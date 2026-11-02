package dosbind

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// pythonPulse is the Python remake's PULSE (internal/level's fixture), the
// transcription of map 00 whose credits and builds records.json holds.
func pythonPulse(t *testing.T) testLevel {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "level", "testdata", "pulse.json"))
	if err != nil {
		t.Fatal(err)
	}
	pulse, err := level.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return testLevel{pulse, pulse.MapNo - 1}
}

// testLevel is a level and the original map it is a version of (00-71).
type testLevel struct {
	*level.Level
	Map int
}

// twoPlayerCodes are the code words of the original two-player maps 48–71;
// they fill campaign months between the original campaign's and after it
// (their cp setup).
var twoPlayerCodes = []string{
	"TRACK", "HUSAR", "BEAST", "PLATE", "LIGHT", "SCROL", "VIRUS", "BISON", "DRUCK", "TROLL", "UBOOT", "DROID",
	"GRAND", "ROYAL", "WATER", "SKILL", "SKULL", "AUDIO", "SPELL", "CAMEL", "FLAGS", "STORY", "SCOUT", "GREEN",
}

// shipped are the Python remake's PULSE (internal/level's fixture, map 00),
// then the setups that hold an original map's units: the cp setup of every
// campaign month made from one.
func shipped(t *testing.T) []testLevel {
	t.Helper()
	out := []testLevel{pythonPulse(t)}
	sh, err := scenario.Shipped(scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range sh {
		file, _, _ := strings.Cut(n.Name, ":")
		var y, m int
		if _, err := fmt.Sscanf(file, "%4d-%2d", &y, &m); err != nil {
			t.Fatalf("%s: %v", n.Name, err)
		}
		if k := slices.Index(twoPlayerCodes, n.Level.MapName); k >= 0 {
			out = append(out, testLevel{n.Level, 48 + k})
		} else if months := 12*(y-1914) + m - 8; months%2 == 0 && months < 48 {
			out = append(out, testLevel{n.Level, months / 2})
		}
	}
	if len(out) != 49 {
		t.Fatalf("%d levels of original maps, want 49", len(out))
	}
	return out
}

func bind(t *testing.T, lv *level.Level) *Board {
	t.Helper()
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	b, err := Bind(lv, set)
	if err != nil {
		t.Fatalf("%s: %v", lv.MapName, err)
	}
	return b
}

func TestBindShippedLevels(t *testing.T) {
	want := map[string]struct {
		w, h   int
		winter bool
	}{"Pulse": {16, 24, false}, "PULSE": {16, 24, false}, "CIVIL": {24, 24, true}, "MOUSE": {24, 24, true}, "WATER": {48, 64, true}}
	terrains := dosrules.Full.Tables.Move.Terrains
	for _, lv := range shipped(t) {
		b := bind(t, lv.Level)
		w, ok := want[lv.MapName]
		if ok && (b.Width != w.w || b.Height != w.h || b.Winter != w.winter) {
			t.Errorf("%s: %d×%d winter %v", lv.MapName, b.Width, b.Height, b.Winter)
		}
		g := lv.Grid()
		for l, line := range lv.Lines {
			for c := range line {
				p := hexgrid.Coord{Line: l, Col: c}
				cell, ok := b.Cell(p)
				if back, in := b.Coord(cell); !ok || !in || back != p {
					t.Fatalf("%s: %v -> cell %d -> %v", lv.MapName, p, cell, back)
				}
			}
		}
		out := 0
		for cell, rec := range b.Terrain {
			if rec < 0 || rec >= len(terrains) {
				t.Fatalf("%s: cell %d: record %d", lv.MapName, cell, rec)
			}
			if _, in := b.Coord(cell); !in {
				out++
				if !b.Edge(cell) || rec != Table.Edge {
					t.Errorf("%s: left-out cell %d is not edge (record %d)", lv.MapName, cell, rec)
				}
			}
			if b.Edge(cell) && rec != Table.Edge {
				t.Errorf("%s: edge cell %d has record %d", lv.MapName, cell, rec)
			}
		}
		if out != b.Width*b.Height-g.Size() {
			t.Errorf("%s: %d cells left out, want %d", lv.MapName, out, b.Width*b.Height-g.Size())
		}
		if _, ok := b.Cell(hexgrid.Coord{Line: g.Lines, Col: 0}); ok {
			t.Errorf("%s: a hex below the level has a cell", lv.MapName)
		}

		gm, err := b.Game(dosrules.Full)
		if err != nil {
			t.Fatalf("%s: NewGame: %v", lv.MapName, err)
		}
		if len(gm.Units) != len(b.Units) || len(gm.Buildings) != len(b.Buildings) {
			t.Errorf("%s: %d units, %d buildings in the game", lv.MapName, len(gm.Units), len(gm.Buildings))
		}
		// Every unit has its remake id but a two-cell ship's rear, the
		// record after its head.
		for i, u := range b.Units {
			if _, ok := b.IDs[i]; ok {
				continue
			}
			if head, ok := b.Units[i-1]; !ok || b.IDs[i-1] == "" || head.Type+1 != u.Type {
				t.Errorf("%s: unit %d (type %d) has no remake id and is no ship's rear", lv.MapName, i, u.Type)
			}
		}
		checkHolds(t, lv.MapName, b)
	}
}

// knownStacked are the cells where Bind places two units: addRear puts the
// rears of two destroyers (PARTS' ec and 2p setups, lines 39 and 42) on the
// one free water hex both touch, as it skips the level's units but not the
// rears it placed before.
var knownStacked = map[string][]int{"1916-02:ec": {670}, "1916-02:2p": {670}}

// checkHolds checks the container records: a building's first cell is its
// entrance, a transporter's its unit, every unit is in at most one record,
// and the buildings are numbered as the game numbers them.
func checkHolds(t *testing.T, name string, b *Board) {
	t.Helper()
	terrains := dosrules.Full.Tables.Move.Terrains
	in := map[int]int{}
	for i, h := range b.Holds {
		if i < dosrules.HoldFlag40 {
			u, ok := b.Units[h.Occupant]
			if h.Flags&carrier == 0 || !ok || h.Cells[0] != 2*u.Cell+1 {
				t.Errorf("%s: transporter record %d: %+v", name, i, h)
			}
			continue
		}
		f := terrains[b.Terrain[h.Cells[0]/2]].Flags
		var want int
		switch {
		case f&flagHQ != 0:
			want = dosrules.HoldFlag40 + f&1
		case f&flagDepot != 0 && i >= dosrules.HoldFlag100 && i < dosrules.HoldFlag400:
			want = i
		case f&flagFactory != 0 && i >= dosrules.HoldFlag400:
			want = i
		default:
			want = -1
		}
		if want != i || owner(h.Flags) != owner(f) || !dosrules.Owns(h.Flags, owner(f)) && owner(f) != 2 {
			t.Errorf("%s: building record %d on record %d (flags %#x): %+v", name, i, b.Terrain[h.Cells[0]/2], f, h)
		}
		if _, ok := b.Buildings[i]; !ok {
			t.Errorf("%s: building record %d without energy", name, i)
		}
		for k, u := range h.Slots[:7] {
			if u > dosrules.SlotFree {
				continue
			}
			if p := b.Units[u]; p.Cell != h.Cells[0]/2 || p.Owner != owner(h.Flags) {
				t.Errorf("%s: record %d holds unit %d at %+v", name, i, u, p)
			}
			if h.Slots[7+k] != u {
				t.Errorf("%s: record %d: the sides' slots differ", name, i)
			}
			in[u]++
		}
	}
	for u, n := range in {
		if n > 1 {
			t.Errorf("%s: unit %d in %d records", name, u, n)
		}
	}
	// The depot and factory tables run gap-free from their bases, in the
	// cell order of their entrances; every entrance has exactly one record.
	for _, base := range []int{dosrules.HoldFlag100, dosrules.HoldFlag400} {
		prev := -1
		for i := base; i < base+tableLen; i++ {
			h, ok := b.Holds[i]
			if !ok {
				for j := i + 1; j < base+tableLen; j++ {
					if _, ok := b.Holds[j]; ok {
						t.Errorf("%s: record %d without record %d", name, j, i)
					}
				}
				break
			}
			if h.Cells[0] <= prev {
				t.Errorf("%s: record %d at cell %d, the record before at %d", name, i, h.Cells[0]/2, prev/2)
			}
			prev = h.Cells[0]
		}
	}
	entered := map[int]int{}
	for i, h := range b.Holds {
		if i >= dosrules.HoldFlag40 {
			entered[h.Cells[0]/2]++
		}
	}
	for cell, rec := range b.Terrain {
		if terrains[rec].Flags&(flagHQ|flagDepot|flagFactory) != 0 && entered[cell] != 1 {
			t.Errorf("%s: entrance cell %d (record %d) entered by %d records", name, cell, rec, entered[cell])
		}
	}
	cells := map[int]int{}
	for i, u := range b.Units {
		if in[i] == 0 {
			if j, dup := cells[u.Cell]; dup && !slices.Contains(knownStacked[name], u.Cell) {
				t.Errorf("%s: units %d and %d on cell %d", name, i, j, u.Cell)
			}
			cells[u.Cell] = i
		}
	}
}

func TestKey(t *testing.T) {
	for _, c := range []struct {
		h    level.Hex
		want string
	}{
		{level.Hex{Terrain: "grass_w"}, "grass"},
		{level.Hex{Terrain: "grass", Infra: level.Infra{IDs: []string{"road25", "road25", "road25"}}}, "grass+road25"},
		{level.Hex{Terrain: "grass", Infra: level.Infra{IDs: []string{"stream25", "rBridge36", "road36"}}}, "grass+stream25+rBridge36+road36"},
	} {
		if got := Key(c.h); got != c.want {
			t.Errorf("Key(%+v) = %q, want %q", c.h, got, c.want)
		}
	}
}

func TestBindErrorsNameTheHex(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	lv := shipped(t)[0].Level
	for _, c := range []struct {
		edit func(h *level.Hex)
		want string
	}{
		{func(h *level.Hex) { h.Terrain = "lava" }, `line3[1]: no terrain record for "lava`},
		{func(h *level.Hex) { h.Unit = "CP_dragon" }, `line3[1]: unknown unit "CP_dragon"`},
	} {
		cp := *lv
		cp.Lines = make([][]level.Hex, len(lv.Lines))
		for i := range lv.Lines {
			cp.Lines[i] = slices.Clone(lv.Lines[i])
		}
		c.edit(&cp.Lines[2][1])
		if _, err := Bind(&cp, set); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("Bind: %v, want %q", err, c.want)
		}
	}
}

// everySetup are the Python remake's PULSE and every setup of the shipped
// level folders, of every mode.
func everySetup(t *testing.T) []scenario.Named {
	t.Helper()
	sh, err := scenario.Shipped()
	if err != nil {
		t.Fatal(err)
	}
	return append([]scenario.Named{{Name: "Pulse", Level: pythonPulse(t).Level}}, sh...)
}

// shippedLevel is setup m of the shipped level folder month.
func shippedLevel(t *testing.T, month string, m scenario.Mode) *level.Level {
	t.Helper()
	lv, err := scenario.ShippedLevel(month, m)
	if err != nil {
		t.Fatal(err)
	}
	return lv
}

// TestDroppedContent binds every setup: each unit of a building's content
// is held in the building's container record, none is dropped.
func TestDroppedContent(t *testing.T) {
	for _, n := range everySetup(t) {
		b := bind(t, n.Level)
		content := 0
		for _, line := range n.Level.Lines {
			for _, h := range line {
				if h.Content != nil {
					content += len(h.Content.Units)
				}
			}
		}
		held := 0
		for i, h := range b.Holds {
			if i < dosrules.HoldFlag40 {
				continue
			}
			for _, u := range h.Slots[:7] {
				if u <= dosrules.SlotFree {
					held++
				}
			}
		}
		if held != content {
			t.Errorf("%s: %d of %d content units held", n.Name, held, content)
		}
	}
}

// TestBindHoldsEverySetup checks the container records of every setup.
func TestBindHoldsEverySetup(t *testing.T) {
	for _, n := range everySetup(t) {
		checkHolds(t, n.Name, bind(t, n.Level))
	}
}

// TestBindNumbersByEntrance binds MOUSE, whose two lower factories come in
// the other order by their first cells: the game numbers depots and
// factories by their entrances in cell order.
func TestBindNumbersByEntrance(t *testing.T) {
	b := bind(t, shippedLevel(t, "1914-12", scenario.CP))
	for _, c := range []struct {
		record   int
		entrance hexgrid.Coord
		first    string
	}{
		{dosrules.HoldFlag400, hexgrid.Coord{Line: 15, Col: 3}, "CP_armCar"},
		{dosrules.HoldFlag400 + 1, hexgrid.Coord{Line: 37, Col: 9}, "CP_hArtillery"},
		{dosrules.HoldFlag400 + 2, hexgrid.Coord{Line: 38, Col: 1}, "CP_mArtillery"},
	} {
		cell, _ := b.Cell(c.entrance)
		h, ok := b.Holds[c.record]
		if !ok || h.Cells[0] != 2*cell || b.IDs[h.Slots[0]] != c.first {
			t.Errorf("record %d: %+v (%v), want entrance cell %d holding %s first", c.record, h, ok, cell, c.first)
		}
	}
	if _, ok := b.Holds[dosrules.HoldFlag400+3]; ok {
		t.Error("a fourth factory record")
	}
}

// TestTouchingCampsBindApart binds PARTS, whose depot at line 60 touches
// the camp at line 59: each takes its own four hexes. The camp without
// content is a half depot (no container record), the depot holds its six units.
func TestTouchingCampsBindApart(t *testing.T) {
	b := bind(t, shippedLevel(t, "1916-02", scenario.CP))
	cells := func(cs ...hexgrid.Coord) []int {
		var out []int
		for _, c := range cs {
			cell, _ := b.Cell(c)
			out = append(out, cell)
		}
		return out
	}
	for _, c := range []struct {
		record int
		camp   []int // entrance, then its other cells in cell order
		units  []string
	}{
		{
			dosrules.HoldFlag100,
			cells(hexgrid.Coord{Line: 57, Col: 5}, hexgrid.Coord{Line: 58, Col: 5}, hexgrid.Coord{Line: 59, Col: 5}, hexgrid.Coord{Line: 58, Col: 6}),
			[]string{"CP_fokkerEI", "CP_lArtillery", "CP_eliteInf", "CP_infantry", "CP_infantry", "CP_sappers"},
		},
		{
			dosrules.HoldFlag100 + 1,
			cells(hexgrid.Coord{Line: 59, Col: 9}, hexgrid.Coord{Line: 60, Col: 9}, hexgrid.Coord{Line: 61, Col: 9}, hexgrid.Coord{Line: 60, Col: 10}),
			[]string{"CP_lArtillery", "CP_lArtillery", "CP_eliteInf", "CP_eliteInf", "CP_infantry", "CP_infantry"},
		},
	} {
		h, ok := b.Holds[c.record]
		var ids []string
		for _, u := range h.Slots[:7] {
			if u <= dosrules.SlotFree {
				ids = append(ids, b.IDs[u])
			}
		}
		if !ok || !slices.Equal(ids, c.units) {
			t.Errorf("record %d holds %v (%v), want %v", c.record, ids, ok, c.units)
		}
		for k, cell := range c.camp {
			if h.Cells[k] != 2*cell {
				t.Errorf("record %d: cells %v, want %v (doubled)", c.record, h.Cells, c.camp)
				break
			}
			if e, ok := b.Parts[cell]; k > 0 && (!ok || e != c.camp[0]) {
				t.Errorf("record %d: cell %d is part of %d (%v)", c.record, cell, e, ok)
			}
		}
	}
	if _, ok := b.Holds[dosrules.HoldFlag100+2]; ok {
		t.Error("a third depot record")
	}
	half := cells(hexgrid.Coord{Line: 58, Col: 11}, hexgrid.Coord{Line: 59, Col: 10}, hexgrid.Coord{Line: 59, Col: 11}, hexgrid.Coord{Line: 60, Col: 11})
	for k, rec := range []int{Table.Terrain["cmpN@-"], Table.Terrain["cmpW@-"], Table.Terrain["cmpE@-"], Table.Terrain["cmpS@-"]} {
		cell := half[k]
		if b.Terrain[cell] != rec {
			t.Errorf("half depot cell %d: record %d, want %d", cell, b.Terrain[cell], rec)
		}
		if e, ok := b.Parts[cell]; ok {
			t.Errorf("half depot cell %d is part of the building at %d", cell, e)
		}
		for i, h := range b.Holds {
			if i >= dosrules.HoldFlag40 && slices.Contains(h.Cells[:], 2*cell) {
				t.Errorf("half depot cell %d in record %d", cell, i)
			}
		}
	}
}

// TestBindBuildingErrors binds PARTS with a building broken: an error names
// the hex.
func TestBindBuildingErrors(t *testing.T) {
	set := loadSet(t)
	lv := shippedLevel(t, "1916-02", scenario.CP)
	for _, c := range []struct {
		edit func(lines [][]level.Hex)
		want string
	}{
		// The depot entered at line60[9] without its S piece.
		{func(lines [][]level.Hex) { lines[61][9].Terrain = "grass" }, "line62[9]: no cmpS of the building at line60[9]"},
		// A piece that no building's N piece stamps.
		{func(lines [][]level.Hex) { lines[0][0].Terrain = "cmpW" }, "line1[0]: cmpW outside a complete building"},
		// Content on the W piece of a depot that has its own.
		{
			func(lines [][]level.Hex) { lines[60][9].Content = &level.Content{Units: []string{"CP_infantry"}} },
			"line61[9]: a second content of the building at line60[9]",
		},
	} {
		cp := *lv
		cp.Lines = make([][]level.Hex, len(lv.Lines))
		for i := range lv.Lines {
			cp.Lines[i] = slices.Clone(lv.Lines[i])
		}
		c.edit(cp.Lines)
		if _, err := Bind(&cp, set); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("Bind: %v, want %q", err, c.want)
		}
	}
	// Content on any piece of a building is the building's.
	cp := *lv
	cp.Lines = make([][]level.Hex, len(lv.Lines))
	for i := range lv.Lines {
		cp.Lines[i] = slices.Clone(lv.Lines[i])
	}
	cp.Lines[60][9].Content, cp.Lines[59][9].Content = cp.Lines[59][9].Content, nil
	b, err := Bind(&cp, set)
	if err != nil {
		t.Fatal(err)
	}
	if h := b.Holds[dosrules.HoldFlag100+1]; b.IDs[h.Slots[0]] != "CP_lArtillery" {
		t.Errorf("content on the depot's W piece: record %+v", h)
	}
}
