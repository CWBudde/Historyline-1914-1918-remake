package dosoracle

import (
	"errors"
	"maps"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// Movement fixtures (docs/research/movement-fixture.schema.json) are recorded
// from the executed movement queries, numbers only, into the dosrules
// testdata (movements.json); the profile's tables (Tables.Move) hold the
// fields of the build's type and terrain records they read. dosrules'
// TestMovementFixtures replays them without the original files.

// fixtureMoveStride records every k-th case of a movement class.
var fixtureMoveStride = map[string]int{"terrain": 3, "transition": 5, "random": 4}

// holdCell reports whether the end-cell check takes its container branch on
// the cell (an own transporter, or building terrain without a unit). The
// movement and step fixtures leave those cells to the container fixtures.
func holdCell(tab *MoveTables, g *Movement, cell int) bool {
	if idx, ok := g.Occupants()[cell]; ok {
		o := g.PieceAt(idx)
		if o.Flags4&0x40 != 0 && o.Flags4&0x80 == 0 {
			o = g.PieceAt(idx - 1)
		}
		return dosrules.Owns(o.Flags4, g.Side) && o.Flags4&0x1000 != 0
	}
	return tab.Terrains[g.Terrain[cell]].Flags&(0x400|0x100|0x40) != 0
}

// moveFixtureOf records a case; the end codes, destinations and routes the
// transcription leaves undefined by its inputs are left out.
func moveFixtureOf(e *Export, tab *MoveTables, c moveCase, r MoveRun) dosrules.MovementFixture {
	g := c.g
	f := dosrules.MovementFixture{Schema: 1, Profile: "dos-0.1", ID: c.class + "/" + c.name,
		Source: source(e, "movement queries"), FixtureMap: dosrules.MapOf(g), Goals: []int{}}
	x := &f.Expected
	for cell := range g.Width * g.Height {
		x.Remaining = append(x.Remaining, int(r.Buffer[cell/g.Width*64+cell%g.Width]))
	}
	x.Count = r.Count
	x.Codes = [][2]int{}
	container := false
	for _, cell := range slices.Sorted(maps.Keys(r.Codes)) {
		if holdCell(tab, g, cell) {
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
func movementFixtures(t *testing.T, e *Export) []byte {
	tab := moveTablesOf(e)
	var fs []dosrules.MovementFixture
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
	cases, err := dosrules.MarshalFixtures(fs)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d movement cases", len(fs))
	return cases
}
