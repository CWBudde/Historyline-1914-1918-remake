package dosoracle

import (
	"fmt"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// Container fixtures (docs/research/container-fixture.schema.json) are
// recorded from the executed container cases, numbers only, beside the step
// fixtures (containers.json). dosrules' TestContainerFixtures replays them
// without the original files.

// holdFixtureStride records every k-th case of a container class.
var holdFixtureStride = map[string]int{"load": 4, "exit": 2}

func holdFixtureOf(e *Export, c holdCase) dosrules.ContainerFixture {
	f := dosrules.ContainerFixture{Schema: 1, Profile: "dos-0.1", ID: fmt.Sprintf("%s/%s/to %d", c.class, c.name, c.o.Goal),
		Source: source(e, "container orders"), FixtureMap: dosrules.MapOf(c.g), Held: dosrules.FixturePieces(c.g.Held),
		Records: dosrules.FixtureHolds(c.g.Holds), Order: c.o}
	if f.Order.Stale == nil {
		f.Order.Stale = []int{}
	}
	return f
}

// holdFixtures records the container cases of an export.
func holdFixtures(t *testing.T, e *Export) []byte {
	tab := moveTablesOf(e)
	var fs []dosrules.ContainerFixture
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
		if f.Expected, err = dosrules.ContainerExpectedOf(c.g, c.o, r); err != nil {
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
