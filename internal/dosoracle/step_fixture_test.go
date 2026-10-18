package dosoracle

import (
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

// Step fixtures (docs/research/step-fixture.schema.json) are recorded from
// the executed move orders, numbers only, beside the movement fixtures, whose
// movement-tables.json they use. TestStepFixtures replays them through the
// transcription without the original files.

// fixtureRecord is a live record after the move, its 26 bytes in order.
type fixtureRecord struct {
	Index int   `json:"index"`
	Bytes []int `json:"bytes"`
}

// stepExpected is what the executed order gave.
type stepExpected struct {
	Order    int             `json:"order"` // AX of the order code: 1, 9 for a two-cell ship, 0x12 refused
	Start    int             `json:"start"` // AX of the move start, 0 where the order was refused
	Messages [][2]int        `json:"messages"`
	Cells    []int           `json:"cells"` // the mover's cell after each step
	Live     []fixtureRecord `json:"live"`
	// Occupants holds [cell, unit] for every occupied cell of the side's map
	// and of the other side's map afterwards; the terrain bytes stay.
	Occupants [2][][2]int `json:"occupants"`
	Bits      [][2]int    `json:"bits"` // [row*64+column, value] of every non-zero byte
	Orders    []int       `json:"orders"`
	Count     int         `json:"count"`
	Container []int       `json:"container"` // container record 0
}

type stepFixture struct {
	Schema  int             `json:"schema"`
	Profile string          `json:"profile"`
	ID      string          `json:"id"`
	Source  dosrules.Source `json:"source"`
	fixtureMap
	Order    MoveOrder    `json:"order"`
	Expected stepExpected `json:"expected"`
}

// stepFixtureStride records every k-th compared case of a class.
var stepFixtureStride = map[string]int{"terrain": 4, "transition": 8, "units": 3, "random": 6, "edge": 2}

func ints(b []byte) []int {
	out := make([]int, len(b))
	for i, v := range b {
		out[i] = int(v)
	}
	return out
}

// stepExpectedOf is the recorded form of an outcome; it fails where a map's
// terrain bytes changed, which the fixture does not record.
func stepExpectedOf(g *Movement, r StepRun) (stepExpected, error) {
	x := stepExpected{Order: r.Order, Start: r.Start, Messages: r.Messages, Cells: r.Cells,
		Orders: r.Orders, Count: r.Count, Container: ints(r.Container)}
	if x.Messages == nil {
		x.Messages = [][2]int{}
	}
	if x.Cells == nil {
		x.Cells = []int{}
	}
	if x.Orders == nil {
		x.Orders = []int{}
	}
	for _, idx := range slices.Sorted(maps.Keys(r.Live)) {
		x.Live = append(x.Live, fixtureRecord{idx, ints(r.Live[idx])})
	}
	for k, mp := range r.Maps {
		x.Occupants[k] = [][2]int{}
		for c := range g.Width * g.Height {
			if int(mp[2*c]) != g.Terrain[c] {
				return x, fmt.Errorf("map %d: terrain of cell %d changed", k, c)
			}
			if mp[2*c+1] != 0xFF {
				x.Occupants[k] = append(x.Occupants[k], [2]int{c, int(mp[2*c+1])})
			}
		}
	}
	x.Bits = [][2]int{}
	for i, v := range r.Bits {
		if v != 0 {
			x.Bits = append(x.Bits, [2]int{i, int(v)})
		}
	}
	return x, nil
}

// stepFixtures records the compared step cases of an export.
func stepFixtures(t *testing.T, e *Export) []byte {
	tab := moveTablesOf(e)
	var fs []stepFixture
	perClass := map[string]int{}
	for _, c := range stepCases(tab) {
		if _, err := tab.Execute(c.g, c.o); err != nil {
			continue
		}
		perClass[c.class]++
		if k := stepFixtureStride[c.class]; k > 1 && (perClass[c.class]-1)%k != 0 {
			continue
		}
		r, err := e.RunMove(c.g, c.o)
		if err != nil {
			t.Fatalf("%s/%s to %d: %v", c.class, c.name, c.o.Goal, err)
		}
		x, err := stepExpectedOf(c.g, r)
		if err != nil {
			t.Fatalf("%s/%s to %d: %v", c.class, c.name, c.o.Goal, err)
		}
		o := c.o
		if o.Stale == nil {
			o.Stale = []int{}
		}
		fs = append(fs, stepFixture{Schema: 1, Profile: "dos-0.1", ID: fmt.Sprintf("%s/%s/to %d", c.class, c.name, c.o.Goal),
			Source: source(e, "move orders"), fixtureMap: fixtureMapOf(c.g), Order: o, Expected: x})
	}
	out, err := dosrules.MarshalFixtures(fs)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d step cases", len(fs))
	return out
}

// TestStepFixtures replays the recorded move orders through the
// transcription; it needs no original files.
func TestStepFixtures(t *testing.T) {
	tab, _ := loadMovementFixtures(t, "demo")
	b, err := os.ReadFile(filepath.Join("..", "dosrules", "testdata", "demo", "steps.json"))
	if err != nil {
		t.Fatal(err)
	}
	fs, err := dosrules.UnmarshalFixtures[stepFixture](b)
	if err != nil {
		t.Fatal(err)
	}
	classes := map[string]int{}
	for _, f := range fs {
		if err := replayStep(tab, f); err != nil {
			t.Errorf("%s: %v", f.ID, err)
		}
		class, _, _ := strings.Cut(f.ID, "/")
		classes[class]++
	}
	for _, class := range []string{"terrain", "transition", "units", "unseen", "block", "budget", "edge", "random", "order", "retarget", "pass"} {
		if classes[class] == 0 {
			t.Errorf("no recorded %s case", class)
		}
	}
	t.Logf("%d recorded step cases: %v", len(fs), classes)
}

func replayStep(tab *MoveTables, f stepFixture) error {
	g := f.movement()
	r, err := tab.Execute(g, f.Order)
	if err != nil {
		return err
	}
	got, err := stepExpectedOf(g, r)
	if err != nil {
		return err
	}
	want := f.Expected
	gv, wv := reflect.ValueOf(got), reflect.ValueOf(want)
	for i := range gv.NumField() {
		if !reflect.DeepEqual(gv.Field(i).Interface(), wv.Field(i).Interface()) {
			return fmt.Errorf("%s is %v, recorded %v", gv.Type().Field(i).Name, gv.Field(i).Interface(), wv.Field(i).Interface())
		}
	}
	return nil
}
