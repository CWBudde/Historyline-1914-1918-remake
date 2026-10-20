package dosoracle

import (
	"fmt"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// Step fixtures (docs/research/step-fixture.schema.json) are recorded from
// the executed move orders, numbers only, beside the movement fixtures (steps.json).
// dosrules' TestStepFixtures replays them without the original files.

// stepFixtureStride records every k-th compared case of a class.
var stepFixtureStride = map[string]int{"terrain": 4, "transition": 8, "units": 3, "random": 6, "edge": 2}

// stepFixtures records the compared step cases of an export.
func stepFixtures(t *testing.T, e *Export) []byte {
	tab := moveTablesOf(e)
	var fs []dosrules.StepFixture
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
		x, err := dosrules.StepExpectedOf(c.g, r)
		if err != nil {
			t.Fatalf("%s/%s to %d: %v", c.class, c.name, c.o.Goal, err)
		}
		o := c.o
		if o.Stale == nil {
			o.Stale = []int{}
		}
		fs = append(fs, dosrules.StepFixture{Schema: 1, Profile: "dos-0.1", ID: fmt.Sprintf("%s/%s/to %d", c.class, c.name, c.o.Goal),
			Source: source(e, "move orders"), FixtureMap: dosrules.MapOf(c.g), Order: o, Expected: x})
	}
	out, err := dosrules.MarshalFixtures(fs)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d step cases", len(fs))
	return out
}
