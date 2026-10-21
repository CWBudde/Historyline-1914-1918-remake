package dosbind

import (
	"io/fs"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// shipped are the shipped levels in file order.
func shipped(t *testing.T) []*level.Level {
	t.Helper()
	names, err := fs.Glob(levels.FS, "level?_cp.json")
	if err != nil || len(names) == 0 {
		t.Fatalf("no shipped levels: %v", err)
	}
	var out []*level.Level
	for _, n := range names {
		raw, err := levels.FS.ReadFile(n)
		if err != nil {
			t.Fatal(err)
		}
		lv, err := level.Parse(raw)
		if err != nil {
			t.Fatalf("%s: %v", n, err)
		}
		out = append(out, lv)
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
	}{"Pulse": {16, 24, false}, "Civil": {24, 24, true}, "Mouse": {24, 24, true}}
	terrains := dosrules.Full.Tables.Move.Terrains
	for _, lv := range shipped(t) {
		b := bind(t, lv)
		w, ok := want[lv.MapName]
		if !ok || b.Width != w.w || b.Height != w.h || b.Winter != w.winter {
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
		if len(gm.Units) != len(b.Units) || len(gm.Buildings) != len(b.Buildings) || len(b.IDs) != len(b.Units) {
			t.Errorf("%s: %d units, %d buildings in the game", lv.MapName, len(gm.Units), len(gm.Buildings))
		}
		checkHolds(t, lv.MapName, b)
	}
}

// checkHolds checks the container records: a building's first cell is its
// entrance, a transporter's its unit, and every unit is in at most one
// record.
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
	cells := map[int]int{}
	for i, u := range b.Units {
		if in[i] == 0 {
			if j, dup := cells[u.Cell]; dup {
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
	lv := shipped(t)[0]
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

func TestDroppedContent(t *testing.T) {
	for _, lv := range shipped(t) {
		b := bind(t, lv)
		n := 0
		for _, ids := range b.Dropped {
			n += len(ids)
		}
		want := map[string]int{"Civil": 2}[lv.MapName]
		if n != want {
			t.Errorf("%s: %d dropped units %v, want %d", lv.MapName, n, b.Dropped, want)
		}
	}
}
