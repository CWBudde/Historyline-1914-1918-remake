package dosoracle

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// Container fixtures (docs/research/container-fixture.schema.json) are
// recorded from the executed container cases, numbers only, beside the step
// fixtures, whose movement-tables.json they use. TestContainerFixtures
// replays them through the transcription without the original files.

// fixtureHold is a container record of a fixture.
type fixtureHold struct {
	Index int `json:"index"`
	Hold
}

// holdExpected is what the executed case gave: the step outcome, the
// container records that changed (all 0x1c bytes) and, for an exit, [cell,
// budget left] of every cell its reach search reached.
type holdExpected struct {
	stepExpected
	Holds     []fixtureRecord `json:"holds"`
	Remaining [][2]int        `json:"remaining"`
}

type holdFixture struct {
	Schema  int             `json:"schema"`
	Profile string          `json:"profile"`
	ID      string          `json:"id"`
	Source  dosrules.Source `json:"source"`
	fixtureMap
	Held     []fixturePiece `json:"held"`
	Records  []fixtureHold  `json:"records"`
	Order    ExitOrder      `json:"order"`
	Expected holdExpected   `json:"expected"`
}

// holdFixtureStride records every k-th case of a container class.
var holdFixtureStride = map[string]int{"load": 4, "exit": 2}

func holdFixtureOf(e *Export, c holdCase) holdFixture {
	f := holdFixture{Schema: 1, Profile: "dos-0.1", ID: fmt.Sprintf("%s/%s/to %d", c.class, c.name, c.o.Goal),
		Source: source(e, "container orders"), fixtureMap: fixtureMapOf(c.g), Held: pieces(c.g.Held),
		Records: []fixtureHold{}, Order: c.o}
	for _, i := range slices.Sorted(maps.Keys(c.g.Holds)) {
		f.Records = append(f.Records, fixtureHold{i, c.g.Holds[i]})
	}
	if f.Order.Stale == nil {
		f.Order.Stale = []int{}
	}
	return f
}

// holdExpectedOf is the recorded form of an outcome.
func holdExpectedOf(g *Movement, o ExitOrder, r ExitRun) (holdExpected, error) {
	x := holdExpected{Holds: []fixtureRecord{}, Remaining: [][2]int{}}
	var err error
	if x.stepExpected, err = stepExpectedOf(g, r.StepRun); err != nil {
		return x, err
	}
	before := holdBytes(g)
	for i := range holdRecords {
		if !bytes.Equal(before[i], r.Holds[i]) {
			x.Holds = append(x.Holds, fixtureRecord{i, ints(r.Holds[i])})
		}
	}
	if o.Hold >= 0 {
		for c := range g.Width * g.Height {
			if v := r.Remaining[c/g.Width*64+c%g.Width]; v >= 0 {
				x.Remaining = append(x.Remaining, [2]int{c, int(v)})
			}
		}
	}
	return x, nil
}

// holdFixtures records the container cases of an export.
func holdFixtures(t *testing.T, e *Export) []byte {
	tab := moveTablesOf(e)
	var fs []holdFixture
	perClass := map[string]int{}
	for _, c := range holdCases(tab) {
		perClass[c.class]++
		if k := holdFixtureStride[c.class]; k > 1 && (perClass[c.class]-1)%k != 0 {
			continue
		}
		var r ExitRun
		var err error
		if c.o.Hold < 0 {
			r.StepRun, err = e.RunMove(c.g, c.o.MoveOrder)
		} else {
			r, err = e.RunExit(c.g, c.o)
		}
		if err != nil {
			t.Fatalf("%s/%s to %d: %v", c.class, c.name, c.o.Goal, err)
		}
		f := holdFixtureOf(e, c)
		if f.Expected, err = holdExpectedOf(c.g, c.o, r); err != nil {
			t.Fatalf("%s/%s to %d: %v", c.class, c.name, c.o.Goal, err)
		}
		fs = append(fs, f)
	}
	out, err := dosrules.MarshalFixtures(fs)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d container cases", len(fs))
	return out
}

// movement rebuilds the map, the held units and the records of a fixture.
func (f holdFixture) movement() *Movement {
	g := f.fixtureMap.movement()
	g.Held, g.Holds = map[int]Piece{}, map[int]Hold{}
	for _, p := range f.Held {
		g.Held[p.Index] = p.Piece
	}
	for _, h := range f.Records {
		g.Holds[h.Index] = h.Hold
	}
	return g
}

// TestContainerFixtures replays the recorded container cases through the
// transcription; it needs no original files.
func TestContainerFixtures(t *testing.T) {
	tab, _ := loadMovementFixtures(t, "demo")
	b, err := os.ReadFile(filepath.Join("..", "dosrules", "testdata", "demo", "containers.json"))
	if err != nil {
		t.Fatal(err)
	}
	fs, err := dosrules.UnmarshalFixtures[holdFixture](b)
	if err != nil {
		t.Fatal(err)
	}
	classes := map[string]int{}
	for _, f := range fs {
		if err := replayHold(tab, f); err != nil {
			t.Errorf("%s: %v", f.ID, err)
		}
		class, _, _ := strings.Cut(f.ID, "/")
		classes[class]++
	}
	for _, class := range []string{"load", "slots", "weight", "moved", "nested", "building", "carry", "exit"} {
		if classes[class] == 0 {
			t.Errorf("no recorded %s case", class)
		}
	}
	t.Logf("%d recorded container cases: %v", len(fs), classes)
}

func replayHold(tab *MoveTables, f holdFixture) error {
	g := f.movement()
	var r ExitRun
	var err error
	if f.Order.Hold < 0 {
		r.StepRun, err = tab.Execute(g, f.Order.MoveOrder)
	} else {
		r, err = tab.Exit(g, f.Order)
	}
	if err != nil {
		return err
	}
	got, err := holdExpectedOf(g, f.Order, r)
	if err != nil {
		return err
	}
	for _, pair := range [][2]any{{got.stepExpected, f.Expected.stepExpected}, {got, f.Expected}} {
		gv, wv := reflect.ValueOf(pair[0]), reflect.ValueOf(pair[1])
		for i := range gv.NumField() {
			if gv.Type().Field(i).Anonymous {
				continue
			}
			if !reflect.DeepEqual(gv.Field(i).Interface(), wv.Field(i).Interface()) {
				return fmt.Errorf("%s is %v, recorded %v", gv.Type().Field(i).Name, gv.Field(i).Interface(), wv.Field(i).Interface())
			}
		}
	}
	return nil
}
