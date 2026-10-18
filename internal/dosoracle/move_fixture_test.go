package dosoracle

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// Movement fixtures (docs/research/movement-fixture.schema.json) are recorded
// from the executed movement queries, numbers only, next to the dosrules
// fixtures: movement-tables.json holds the fields of the build's type and
// terrain records they read, movements.json the cases. TestMovementFixtures
// replays them through the transcription without the original files.

// fixturePiece is a live unit of a movement fixture.
type fixturePiece struct {
	Index int `json:"index"`
	Piece
}

// moveExpected is what the executed queries gave.
type moveExpected struct {
	// Remaining is the cost buffer after the reach search, one signed byte
	// per cell in cell order: the budget left on reaching it, or below 0 the
	// negated entry cost, -80 closed, -102/-101 a zone of control of side
	// 0/1 that was not reached, -1 untouched.
	Remaining []int `json:"remaining"`
	Count     int   `json:"count"`
	// Destinations is null where the end-cell check reaches its container
	// branch (an own transporter or a building cell), which is not recorded.
	Destinations []int `json:"destinations"`
	// Codes holds [cell, end code] for every reached cell outside the
	// container branch.
	Codes [][2]int `json:"codes"`
	// Paths holds, per goal, the cells from the goal back to the start, or
	// null where the search failed.
	Paths [][]int `json:"paths"`
}

type moveFixture struct {
	Schema   int             `json:"schema"`
	Profile  string          `json:"profile"`
	ID       string          `json:"id"`
	Source   dosrules.Source `json:"source"`
	Width    int             `json:"width"`
	Height   int             `json:"height"`
	Fill     int             `json:"fill"`  // terrain record of every cell not in Cells
	Cells    [][2]int        `json:"cells"` // [cell, terrain record]
	Units    []fixturePiece  `json:"units"`
	Unseen   []fixturePiece  `json:"unseen"`
	Unit     int             `json:"unit"`
	Side     int             `json:"side"`
	Budget   int             `json:"budget"`
	Block    int             `json:"block"`
	Goals    []int           `json:"goals"`
	Expected moveExpected    `json:"expected"`
}

// fixtureMoveStride records every k-th case of a movement class.
var fixtureMoveStride = map[string]int{"terrain": 3, "transition": 5, "random": 4}

func pieces(m map[int]Piece) []fixturePiece {
	out := []fixturePiece{}
	for _, i := range slices.Sorted(maps.Keys(m)) {
		out = append(out, fixturePiece{i, m[i]})
	}
	return out
}

// moveFixtureOf records a case; the end codes, destinations and routes the
// transcription leaves undefined by its inputs are left out.
func moveFixtureOf(e *Export, tab *MoveTables, c moveCase, r MoveRun) moveFixture {
	g := c.g
	f := moveFixture{Schema: 1, Profile: "dos-0.1", ID: c.class + "/" + c.name,
		Source: source(e, "movement queries"), Width: g.Width, Height: g.Height,
		Cells: [][2]int{}, Units: pieces(g.Units), Unseen: pieces(g.Unseen),
		Unit: g.Unit, Side: g.Side, Budget: g.Budget, Block: g.Block, Goals: []int{}}
	counts := map[int]int{}
	for _, t := range g.Terrain {
		counts[t]++
	}
	for t, n := range counts {
		if n > counts[f.Fill] || n == counts[f.Fill] && t < f.Fill {
			f.Fill = t
		}
	}
	for cell, t := range g.Terrain {
		if t != f.Fill {
			f.Cells = append(f.Cells, [2]int{cell, t})
		}
	}
	x := &f.Expected
	for cell := range g.Width * g.Height {
		x.Remaining = append(x.Remaining, int(r.Buffer[cell/g.Width*64+cell%g.Width]))
	}
	x.Count = r.Count
	x.Codes = [][2]int{}
	container := false
	for _, cell := range slices.Sorted(maps.Keys(r.Codes)) {
		if _, err := tab.EndCode(g, cell); errors.Is(err, ErrContainer) {
			container = true
			continue
		}
		x.Codes = append(x.Codes, [2]int{cell, r.Codes[cell]})
	}
	if !container {
		x.Destinations = append([]int{}, r.Destinations...)
	}
	buf, bits, _ := tab.Reach(g)
	x.Paths = [][]int{}
	for k, goal := range g.Goals {
		if _, err := tab.Path(g, buf, bits, goal); errors.Is(err, ErrUndefined) {
			continue
		}
		f.Goals = append(f.Goals, goal)
		x.Paths = append(x.Paths, r.Paths[k])
	}
	return f
}

// movementFixtures records the movement cases of an export.
func movementFixtures(t *testing.T, e *Export) (tables []byte, cases []byte) {
	tab := moveTablesOf(e)
	var fs []moveFixture
	perClass := map[string]int{}
	for _, c := range moveCases(tab) {
		perClass[c.class]++
		if k := fixtureMoveStride[c.class]; k > 1 && (perClass[c.class]-1)%k != 0 {
			continue
		}
		r, err := e.RunMovement(c.g)
		if err != nil {
			t.Fatalf("%s/%s: %v", c.class, c.name, err)
		}
		fs = append(fs, moveFixtureOf(e, tab, c, r))
	}
	tables, err := json.MarshalIndent(tab, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if cases, err = dosrules.MarshalFixtures(fs); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d movement cases", len(fs))
	return append(tables, '\n'), cases
}

// movementOf rebuilds the query of a fixture.
func movementOf(f moveFixture) *Movement {
	g := &Movement{Width: f.Width, Height: f.Height, Terrain: make([]int, f.Width*f.Height),
		Units: map[int]Piece{}, Unseen: map[int]Piece{}, Unit: f.Unit, Side: f.Side,
		Budget: f.Budget, Block: f.Block, Goals: f.Goals}
	for i := range g.Terrain {
		g.Terrain[i] = f.Fill
	}
	for _, c := range f.Cells {
		g.Terrain[c[0]] = c[1]
	}
	for _, p := range f.Units {
		g.Units[p.Index] = p.Piece
	}
	for _, p := range f.Unseen {
		g.Unseen[p.Index] = p.Piece
	}
	return g
}

func loadMovementFixtures(t *testing.T, build string) (*MoveTables, []moveFixture) {
	t.Helper()
	dir := filepath.Join("..", "dosrules", "testdata", build)
	b, err := os.ReadFile(filepath.Join(dir, "movement-tables.json"))
	if err != nil {
		t.Fatal(err)
	}
	tab := &MoveTables{}
	if err := json.Unmarshal(b, tab); err != nil {
		t.Fatal(err)
	}
	if b, err = os.ReadFile(filepath.Join(dir, "movements.json")); err != nil {
		t.Fatal(err)
	}
	fs, err := dosrules.UnmarshalFixtures[moveFixture](b)
	if err != nil {
		t.Fatal(err)
	}
	return tab, fs
}

// TestMovementFixtures replays the recorded movement queries through the
// transcription; it needs no original files.
func TestMovementFixtures(t *testing.T) {
	tab, fs := loadMovementFixtures(t, "demo")
	classes := map[string]int{}
	for _, f := range fs {
		if err := replayMovement(tab, f); err != nil {
			t.Errorf("%s: %v", f.ID, err)
		}
		class, _, _ := strings.Cut(f.ID, "/")
		classes[class]++
	}
	for _, class := range []string{"terrain", "transition", "units", "unseen", "block", "budget", "edge", "random"} {
		if classes[class] == 0 {
			t.Errorf("no recorded %s case", class)
		}
	}
	t.Logf("%d recorded movement cases: %v", len(fs), classes)
}

func replayMovement(tab *MoveTables, f moveFixture) error {
	g := movementOf(f)
	x := f.Expected
	buf, bits, count := tab.Reach(g)
	for cell, want := range x.Remaining {
		if got := int(buf[cell/g.Width*64+cell%g.Width]); got != want {
			return fmt.Errorf("cell %d keeps %d, recorded %d", cell, got, want)
		}
	}
	if count != x.Count {
		return fmt.Errorf("%d cells reached, recorded %d", count, x.Count)
	}
	for _, c := range x.Codes {
		if got, err := tab.EndCode(g, c[0]); err != nil || got != c[1] {
			return fmt.Errorf("end code of cell %d is %#x (%v), recorded %#x", c[0], got, err, c[1])
		}
	}
	if x.Destinations != nil {
		got, err := tab.Destinations(g, bits)
		if err != nil || !slices.Equal(got, x.Destinations) {
			return fmt.Errorf("destinations %v (%v), recorded %v", got, err, x.Destinations)
		}
	}
	if len(x.Paths) != len(f.Goals) {
		return fmt.Errorf("%d routes for %d goals", len(x.Paths), len(f.Goals))
	}
	for k, goal := range f.Goals {
		got, err := tab.Path(g, buf, bits, goal)
		if err != nil || !slices.Equal(got, x.Paths[k]) || (got == nil) != (x.Paths[k] == nil) {
			return fmt.Errorf("route to %d is %v (%v), recorded %v", goal, got, err, x.Paths[k])
		}
	}
	return nil
}
