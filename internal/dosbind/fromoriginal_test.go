package dosbind

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// TestPlayableCutsTheEdge checks Board.Playable on imported maps and the
// shipped levels: it holds a hex exactly when the hex is a cell off the edge,
// and the last row off the edge keeps both its lines.
func TestPlayableCutsTheEdge(t *testing.T) {
	set := loadSet(t)
	var boards []*Board
	var names []string
	for _, n := range []int{0, 9, 23, 47, 71} {
		_, b := fromOriginal(t, n, set)
		boards, names = append(boards, b), append(names, fmt.Sprintf("map %02d", n))
	}
	for _, lv := range shipped(t) {
		boards, names = append(boards, bind(t, lv)), append(names, lv.MapName)
	}
	for k, b := range boards {
		p := b.Playable()
		for cell := range b.Width * b.Height {
			c, ok := b.Coord(cell)
			if want := ok && !b.Edge(cell); p.Contains(c) != want {
				t.Errorf("%s: cell %d (%v): playable %v, want %v", names[k], cell, c, p.Contains(c), want)
			}
		}
		for l := range p.Lines {
			for col := range p.Width {
				c := hexgrid.Coord{Line: l, Col: col}
				if _, ok := b.Cell(c); p.Contains(c) && !ok {
					t.Errorf("%s: playable %v is off the level", names[k], c)
				}
			}
		}
		last := b.Height - 2
		for _, c := range []hexgrid.Coord{{Line: 2 * last, Col: 0}, {Line: 2*last + 1, Col: 0}} {
			if !p.Contains(c) {
				t.Errorf("%s: line %d of the last row is not playable", names[k], c.Line)
			}
		}
	}
}

// importedMap reads the imported original map n (levels/original/full).
func importedMap(t *testing.T, n int) *finmap.Level {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("../../levels/original/full", fmt.Sprintf("%02d.json", n)))
	if err != nil {
		t.Fatal(err)
	}
	l, err := finmap.Parse(raw)
	if err != nil {
		t.Fatalf("map %02d: %v", n, err)
	}
	return l
}

func loadSet(t *testing.T) *data.Set {
	t.Helper()
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	return set
}

func fromOriginal(t *testing.T, n int, set *data.Set) (*finmap.Level, *Board) {
	t.Helper()
	l := importedMap(t, n)
	b, err := FromOriginal(l, set)
	if err != nil {
		t.Fatalf("map %02d: %v", n, err)
	}
	return l, b
}

// TestFromOriginalImportedMaps binds the 72 imported maps and starts a game
// on each: the units, rears, cargo, container records and building credits
// are what the map implies.
func TestFromOriginalImportedMaps(t *testing.T) {
	set := loadSet(t)
	types := dosrules.Full.Tables.Move.Types
	for n := range 72 {
		l, b := fromOriginal(t, n, set)
		name := fmt.Sprintf("map %02d", n)
		g, err := b.Game(dosrules.Full)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		winter := l.Mission != nil && l.Mission.Winter
		if b.Winter != winter || g.Profile.Winter() != winter || b.Width != l.Width || b.Height != l.Height {
			t.Errorf("%s: %d×%d winter %v, game winter %v", name, b.Width, b.Height, b.Winter, g.Profile.Winter())
		}
		for row, line := range l.Terrain {
			if !slices.Equal(b.Terrain[row*l.Width:(row+1)*l.Width], line) {
				t.Fatalf("%s: terrain row %d differs", name, row)
			}
		}

		units, holds := len(l.Units), len(l.Buildings)
		for _, u := range l.Units {
			if u.Rear != nil {
				units++
			}
			if u.Container != nil {
				holds++
			}
		}
		cargo := 0
		effective := map[int]int{}
		for _, c := range l.Containers {
			if c.Holder == "none" {
				continue
			}
			cargo += len(c.Units)
			for _, u := range c.Units {
				if u.Container != nil {
					holds++
				}
			}
			if !c.Overridden {
				effective[c.Container] = c.Energy
			}
		}
		if len(b.Units) != units+cargo || len(g.Units) != units+cargo || len(b.Holds) != holds || len(g.Holds) != holds ||
			len(b.Buildings) != len(l.Buildings) || len(g.Buildings) != len(l.Buildings) {
			t.Errorf("%s: %d units (game %d), %d holds (game %d), %d buildings; want %d units, %d holds, %d buildings",
				name, len(b.Units), len(g.Units), len(b.Holds), len(g.Holds), len(b.Buildings), units+cargo, holds, len(l.Buildings))
		}

		held := 0
		for _, u := range g.Units {
			if u.Flags4&0x4000 != 0 {
				held++
			}
		}
		if held != cargo {
			t.Errorf("%s: %d held units, want %d", name, held, cargo)
		}

		i := 0
		for _, u := range l.Units {
			p := b.Units[i]
			if p.Type != u.Type || p.Owner != u.Side || p.Cell != u.Cell[0]*l.Width+u.Cell[1] {
				t.Errorf("%s: unit %d %+v, map %+v", name, i, p, u)
			}
			i++
			if u.Rear != nil {
				r := b.Units[i]
				if r.Type != u.Type+1 || r.Owner != u.Side || r.Cell != u.Rear[0]*l.Width+u.Rear[1] ||
					g.Maps[0][2*r.Cell+1] != i || g.Maps[1][2*r.Cell+1] != i {
					t.Errorf("%s: rear %d %+v of %+v", name, i, r, u)
				}
				i++
			}
		}

		for _, bd := range l.Buildings {
			want := effective[bd.Container]
			if _, ok := effective[bd.Container]; !ok && bd.Kind == "hq" {
				want = 50
			}
			got, ok := b.Buildings[bd.Container]
			if !ok || got.Credits != want {
				t.Errorf("%s: building %+v: credits %d (%v), want %d", name, bd, got.Credits, ok, want)
			}
			wantBuilds := []int(nil)
			if bd.Kind == "factory" {
				wantBuilds = l.Producible
			}
			if !slices.Equal(got.Builds, wantBuilds) {
				t.Errorf("%s: building %+v builds %v, want %v", name, bd, got.Builds, wantBuilds)
			}
			h := b.Holds[bd.Container]
			f := dosrules.Full.Tables.Move.Terrains[l.Terrain[bd.Cell[0]][bd.Cell[1]]].Flags
			if h.Cells[0] != 2*(bd.Cell[0]*l.Width+bd.Cell[1]) || h.Occupant != dosrules.InBuilding || h.Flags != owner(f) {
				t.Errorf("%s: building %+v: record %+v", name, bd, h)
			}
		}

		ids := RemakeIDs(set)
		for i, p := range b.Units {
			want := ids[[2]int{p.Type, p.Owner}]
			if id := b.IDs[i]; id != want {
				t.Errorf("%s: unit %d %+v named %q, want %q", name, i, p, id, want)
			}
			if u := set.Units[want]; want != "" && (u.Dos != p.Type || Side(game.FactionOf(u.Country)) != p.Owner) {
				t.Errorf("%s: unit %d %+v: remake unit %q", name, i, p, want)
			}
		}
		// Every transporter takes the record the map gives it.
		for r, h := range b.Holds {
			if r < dosrules.HoldFlag40 && g.Units[h.Occupant].Hold != r {
				t.Errorf("%s: record %d of transporter %d, which took record %d", name, r, h.Occupant, g.Units[h.Occupant].Hold)
			}
		}
		for i, p := range b.Units {
			if types[p.Type].Class&transporter == 0 {
				continue
			}
			if h, ok := b.Holds[g.Units[i].Hold]; !ok || h.Cells[0] != 2*p.Cell+1 || h.Flags&carrier == 0 {
				t.Errorf("%s: transporter %d: record %d %+v", name, i, g.Units[i].Hold, h)
			}
		}
	}
}

// knownNumbering are the building records Bind numbers otherwise than the
// game (and the imported map): the game numbers depots and factories by
// their entrances in cell order, Bind by the buildings' first cells. The
// bound record by the imported one.
var knownNumbering = map[string]map[int]int{"Mouse": {83: 84, 84: 83}}

// TestFromOriginalMatchesBind compares the building records of the shipped
// levels with those of their original maps.
func TestFromOriginalMatchesBind(t *testing.T) {
	set := loadSet(t)
	for _, lv := range shipped(t) {
		bound := bind(t, lv)
		_, b := fromOriginal(t, lv.MapNo-1, set)
		n := 0
		for i, o := range b.Holds {
			if i < dosrules.HoldFlag40 {
				continue
			}
			n++
			j, ok := knownNumbering[lv.MapName][i]
			if !ok {
				j = i
			}
			h, ok := bound.Holds[j]
			if !ok || o.Cells != h.Cells || o.Flags != h.Flags || o.Occupant != h.Occupant {
				t.Errorf("%s: record %d: original %+v, bound %d %+v", lv.MapName, i, o, j, h)
			}
		}
		if n != buildingHolds(bound) {
			t.Errorf("%s: %d building records, bound %d", lv.MapName, n, buildingHolds(bound))
		}
	}
}

// TestFromOriginalShipMoves moves a two-cell ship on an imported map: the
// rear, the record after the head, takes the head's old cell.
func TestFromOriginalShipMoves(t *testing.T) {
	l, b := fromOriginal(t, 9, loadSet(t))
	g, err := b.Game(dosrules.Full)
	if err != nil {
		t.Fatal(err)
	}
	head := -1
	i := 0
	for _, u := range l.Units {
		if u.Rear != nil && u.Side == 0 && head < 0 {
			head = i
		}
		i++
		if u.Rear != nil {
			i++
		}
	}
	if head < 0 {
		t.Fatal("no ship of side 0")
	}
	old, oldRear := g.Units[head].Cell(0), g.Units[head+1].Cell(0)
	dests, err := g.Profile.Destinations(g.Movement(0, head))
	if err != nil || len(dests) == 0 {
		t.Fatalf("ship %d at %d: destinations %v, %v", head, old, dests, err)
	}
	dest := dests[0]
	// The rear moves with its head and takes no order of its own.
	if err := g.Check(dosrules.Command{Kind: dosrules.CmdMove, Side: 0, Unit: head + 1, Cell: dest}); err == nil {
		t.Error("a move order for the rear was accepted")
	}
	ev, err := g.Apply(dosrules.Command{Kind: dosrules.CmdMove, Side: 0, Unit: head, Cell: dest})
	if err != nil || len(ev) != 1 || ev[0].Kind != dosrules.EvMoved {
		t.Fatalf("move to %d: %+v, %v", dest, ev, err)
	}
	route := ev[0].Route
	behind := old
	if len(route) > 1 {
		behind = route[len(route)-2]
	}
	if g.Units[head].Cell(0) != dest || g.Units[head+1].Cell(0) != behind {
		t.Errorf("head at %d (want %d), rear at %d (want %d); route %v", g.Units[head].Cell(0), dest, g.Units[head+1].Cell(0), behind, route)
	}
	m := g.Maps[0]
	if m[2*dest+1] != head || m[2*behind+1] != head+1 {
		t.Errorf("map: %#x at %d, %#x at %d", m[2*dest+1], dest, m[2*behind+1], behind)
	}
	if oldRear != behind && oldRear != dest && m[2*oldRear+1] != 0xFF {
		t.Errorf("the rear's old cell %d holds %#x", oldRear, m[2*oldRear+1])
	}
}
