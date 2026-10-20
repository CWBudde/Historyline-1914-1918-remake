package dosrules

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// The recorded movement queries, move orders, container cases and phase
// changes replay through the record-level functions on the profile's
// movement tables; each class of case must be present.

// fieldDiff names the first field of two structs of one type that differs;
// embedded structs are compared field by field too.
func fieldDiff(got, want any) error {
	gv, wv := reflect.ValueOf(got), reflect.ValueOf(want)
	for i := range gv.NumField() {
		g, w := gv.Field(i).Interface(), wv.Field(i).Interface()
		if gv.Type().Field(i).Anonymous {
			if err := fieldDiff(g, w); err != nil {
				return err
			}
			continue
		}
		if !reflect.DeepEqual(g, w) {
			return fmt.Errorf("%s is %v, recorded %v", gv.Type().Field(i).Name, g, w)
		}
	}
	return nil
}

func moveTables(t *testing.T, p Profile) *MoveTables {
	t.Helper()
	if p.Tables.Move == nil || len(p.Tables.Move.Types) != 51 || len(p.Tables.Move.Terrains) == 0 {
		t.Fatalf("%s: no movement tables", p.Tables.Build)
	}
	return p.Tables.Move
}

// header is the common head of a fixture.
type header struct {
	schema      int
	profile, id string
	source      Source
}

// replay runs every recorded fixture of a file and counts the classes of
// their ids.
func replay[T any](t *testing.T, name string, head func(T) header, run func(Profile, T) error, classes ...string) {
	for build, p := range profiles {
		moveTables(t, p)
		seen, ids := map[string]int{}, map[string]bool{}
		fs := load[T](t, build, name)
		for _, f := range fs {
			h := head(f)
			checkHeader(t, p, h.schema, h.profile, h.id, h.source, ids)
			if err := run(p, f); err != nil {
				t.Errorf("%s: %v", h.id, err)
			}
			class, _, _ := strings.Cut(h.id, "/")
			seen[class]++
		}
		for _, c := range classes {
			if seen[c] == 0 {
				t.Errorf("%s: no recorded %s case", build, c)
			}
		}
		t.Logf("%s: %d recorded cases: %v", build, len(fs), seen)
	}
}

func TestMovementFixtures(t *testing.T) {
	replay(t, "movements.json", func(f MovementFixture) header { return header{f.Schema, f.Profile, f.ID, f.Source} }, replayMovement,
		"terrain", "transition", "units", "unseen", "block", "budget", "edge", "random")
}

func replayMovement(p Profile, f MovementFixture) error {
	tab, g, x := p.Tables.Move, f.Query(), f.Expected
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

func TestStepFixtures(t *testing.T) {
	replay(t, "steps.json", func(f StepFixture) header { return header{f.Schema, f.Profile, f.ID, f.Source} }, replayStep,
		"terrain", "transition", "units", "unseen", "block", "budget", "edge", "random", "order", "retarget", "pass")
}

func replayStep(p Profile, f StepFixture) error {
	g := f.Movement()
	r, err := p.Tables.Move.Execute(g, f.Order)
	if err != nil {
		return err
	}
	got, err := StepExpectedOf(g, r)
	if err != nil {
		return err
	}
	return fieldDiff(got, f.Expected)
}

func TestContainerFixtures(t *testing.T) {
	replay(t, "containers.json", func(f ContainerFixture) header { return header{f.Schema, f.Profile, f.ID, f.Source} }, replayContainer,
		"load", "slots", "weight", "moved", "nested", "building", "carry", "exit")
}

func replayContainer(p Profile, f ContainerFixture) error {
	g := f.Movement()
	var r ExitRun
	var err error
	if f.Order.Hold < 0 {
		r.StepRun, err = p.Tables.Move.Execute(g, f.Order.MoveOrder)
	} else {
		r, err = p.Tables.Move.Exit(g, f.Order)
	}
	if err != nil {
		return err
	}
	got, err := ContainerExpectedOf(g, f.Order, r)
	if err != nil {
		return err
	}
	return fieldDiff(got, f.Expected)
}

func TestVisibilityFixtures(t *testing.T) {
	replay(t, "visibility.json", func(f VisibilityFixture) header { return header{f.Schema, f.Profile, f.ID, f.Source} }, replayVisibility,
		"phase", "sides", "life", "stale", "ships", "terrain", "holds", "random")
}

func replayVisibility(_ Profile, f VisibilityFixture) error {
	g := f.PhaseChange()
	r, err := ChangePhase(g)
	if err != nil {
		return err
	}
	return fieldDiff(VisibilityExpectedOf(g, r), f.Expected)
}
