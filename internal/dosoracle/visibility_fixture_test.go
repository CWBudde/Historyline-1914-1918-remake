package dosoracle

import (
	"maps"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// Visibility fixtures (docs/research/visibility-fixture.schema.json) are
// recorded from the executed phase changes, numbers only (visibility.json).
// dosrules' TestVisibilityFixtures replays them without the original files.

// phaseFixtureStride records every k-th case of a visibility class.
var phaseFixtureStride = map[string]int{"random": 4, "holds": 2, "phase": 2}

// phaseFixtures records the visibility cases of an export.
func phaseFixtures(t *testing.T, e *Export) []byte {
	var fs []dosrules.VisibilityFixture
	perClass := map[string]int{}
	for _, c := range phaseCases() {
		perClass[c.class]++
		if k := phaseFixtureStride[c.class]; k > 1 && (perClass[c.class]-1)%k != 0 {
			continue
		}
		r, err := e.RunPhaseChange(c.g)
		if err != nil {
			t.Fatalf("%s/%s: %v", c.class, c.name, err)
		}
		g := c.g
		f := dosrules.VisibilityFixture{Schema: 1, Profile: "dos-0.1", ID: c.class + "/" + c.name, Source: source(e, "phase change"),
			Width: g.Width, Height: g.Height, Maps: g.Maps, Units: []dosrules.FixtureSighted{}, Records: dosrules.FixtureHolds(g.Holds),
			Attacker: g.Attacker, Phase: g.Phase, Count: g.Count, Expected: dosrules.VisibilityExpectedOf(g, r)}
		for _, i := range slices.Sorted(maps.Keys(g.Units)) {
			f.Units = append(f.Units, dosrules.FixtureSighted{Index: i, Sighted: g.Units[i]})
		}
		fs = append(fs, f)
	}
	out, err := dosrules.MarshalFixtures(fs)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d visibility cases", len(fs))
	return out
}
