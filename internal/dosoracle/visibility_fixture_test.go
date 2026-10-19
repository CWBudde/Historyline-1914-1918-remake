package dosoracle

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// Visibility fixtures (docs/research/visibility-fixture.schema.json) are
// recorded from the executed phase changes, numbers only.
// TestVisibilityFixtures replays them through ChangePhase without the
// original files.

// fixtureSighted is a live unit of a visibility fixture.
type fixtureSighted struct {
	Index int `json:"index"`
	Sighted
}

// phaseExpected is what the executed phase change gave: both maps, the
// records of the listed units, the container records that changed, the
// per-side records, the order counts, the turn counter, the animation flags
// and the messages.
type phaseExpected struct {
	Maps     [2][]int        `json:"maps"`
	Live     []fixtureRecord `json:"live"`
	Holds    []fixtureRecord `json:"holds"`
	Players  [2][]int        `json:"players"`
	Orders   [2]int          `json:"orders"`
	Count    int             `json:"count"`
	Anim     int             `json:"anim"`
	Messages [][2]int        `json:"messages"`
}

type phaseFixture struct {
	Schema   int              `json:"schema"`
	Profile  string           `json:"profile"`
	ID       string           `json:"id"`
	Source   dosrules.Source  `json:"source"`
	Width    int              `json:"width"`
	Height   int              `json:"height"`
	Maps     [2][]int         `json:"maps"`
	Units    []fixtureSighted `json:"units"`
	Records  []fixtureHold    `json:"records"`
	Attacker int              `json:"attacker"`
	Phase    [2]int           `json:"phase"`
	Count    int              `json:"count"`
	Expected phaseExpected    `json:"expected"`
}

// phaseFixtureStride records every k-th case of a visibility class.
var phaseFixtureStride = map[string]int{"random": 4, "holds": 2, "phase": 2}

func phaseExpectedOf(g *PhaseChange, r PhaseRun) phaseExpected {
	x := phaseExpected{Live: []fixtureRecord{}, Holds: []fixtureRecord{}, Orders: r.Orders, Count: r.Count,
		Anim: r.Anim, Messages: r.Messages}
	for k := range 2 {
		x.Maps[k], x.Players[k] = ints(r.Maps[k]), ints(r.Players[k])
	}
	for _, i := range slices.Sorted(maps.Keys(r.Live)) {
		x.Live = append(x.Live, fixtureRecord{i, ints(r.Live[i])})
	}
	before := holdsOf(g.Holds)
	for i := range holdRecords {
		if !bytes.Equal(before[i], r.Holds[i]) {
			x.Holds = append(x.Holds, fixtureRecord{i, ints(r.Holds[i])})
		}
	}
	return x
}

// phaseFixtures records the visibility cases of an export.
func phaseFixtures(t *testing.T, e *Export) []byte {
	var fs []phaseFixture
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
		f := phaseFixture{Schema: 1, Profile: "dos-0.1", ID: c.class + "/" + c.name, Source: source(e, "phase change"),
			Width: g.Width, Height: g.Height, Maps: g.Maps, Units: []fixtureSighted{}, Records: []fixtureHold{},
			Attacker: g.Attacker, Phase: g.Phase, Count: g.Count, Expected: phaseExpectedOf(g, r)}
		for _, i := range slices.Sorted(maps.Keys(g.Units)) {
			f.Units = append(f.Units, fixtureSighted{i, g.Units[i]})
		}
		for _, i := range slices.Sorted(maps.Keys(g.Holds)) {
			f.Records = append(f.Records, fixtureHold{i, g.Holds[i]})
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

func (f phaseFixture) phaseChange() *PhaseChange {
	g := &PhaseChange{Width: f.Width, Height: f.Height, Maps: f.Maps, Units: map[int]Sighted{},
		Holds: map[int]Hold{}, Attacker: f.Attacker, Phase: f.Phase, Count: f.Count}
	for _, u := range f.Units {
		g.Units[u.Index] = u.Sighted
	}
	for _, h := range f.Records {
		g.Holds[h.Index] = h.Hold
	}
	return g
}

// TestVisibilityFixtures replays the recorded phase changes through
// ChangePhase; it needs no original files.
func TestVisibilityFixtures(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "dosrules", "testdata", "demo", "visibility.json"))
	if err != nil {
		t.Fatal(err)
	}
	fs, err := dosrules.UnmarshalFixtures[phaseFixture](b)
	if err != nil {
		t.Fatal(err)
	}
	classes := map[string]int{}
	for _, f := range fs {
		if err := replayPhase(f); err != nil {
			t.Errorf("%s: %v", f.ID, err)
		}
		class, _, _ := strings.Cut(f.ID, "/")
		classes[class]++
	}
	for _, class := range []string{"phase", "sides", "life", "stale", "ships", "terrain", "holds", "random"} {
		if classes[class] == 0 {
			t.Errorf("no recorded %s case", class)
		}
	}
	t.Logf("%d recorded visibility cases: %v", len(fs), classes)
}

func replayPhase(f phaseFixture) error {
	g := f.phaseChange()
	r, err := ChangePhase(g)
	if err != nil {
		return err
	}
	got, want := phaseExpectedOf(g, r), f.Expected
	if d := fieldDiff(got, want); d != "" {
		return fmt.Errorf("%s", d)
	}
	return nil
}
