package dosrules

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// profiles maps each recorded build to its profile; testdata/<name> holds
// its fixtures. The full HL.EXE joins once its export is verified.
var profiles = map[string]Profile{"demo": Demo}

func load[T any](t *testing.T, build, name string) []T {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", build, name))
	if err != nil {
		t.Fatal(err)
	}
	fs, err := UnmarshalFixtures[T](b)
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) == 0 {
		t.Fatalf("%s/%s holds no fixtures", build, name)
	}
	return fs
}

func checkHeader(t *testing.T, p Profile, schema int, profile, id string, s Source, seen map[string]bool) {
	t.Helper()
	if schema != 1 || profile != p.Name || s.Kind != "emulated-routine" || s.Build != p.Tables.Build {
		t.Errorf("%s: header %d %s %+v", id, schema, profile, s)
	}
	if seen[id] {
		t.Errorf("duplicate fixture id %s", id)
	}
	seen[id] = true
}

func TestTables(t *testing.T) {
	if len(DemoTables.Types) != 51 || DemoTables.Build != "83fc4e37" {
		t.Fatalf("demo tables: %d types, build %q", len(DemoTables.Types), DemoTables.Build)
	}
}

// The recorded engagements are the R1 controlled battles: Engage must
// reproduce the handler's outcome field for field.
func TestEngagementFixtures(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			fs := load[EngagementFixture](t, build, "engagements.json")
			seen := map[string]bool{}
			var skipped, noReturn, pincer, blocked, destroyed, transposed int
			for i := range fs {
				f := &fs[i]
				checkHeader(t, p, f.Schema, f.Profile, f.ID, f.Source, seen)
				got := EngagementResult(p.Engage(f.Battlefield(), f.Attacker, f.Target))
				if !reflect.DeepEqual(got, f.Expected) {
					t.Errorf("%s:\nrecorded %+v\nresolved %+v", f.ID, f.Expected, got)
				}
				e := f.Expected
				switch {
				case !e.Battle:
					skipped++
				case !e.ReturnFire:
					noReturn++
				}
				if e.Pincer > 0 {
					pincer++
				}
				if e.Blocking == blockCap {
					blocked++
				}
				if len(e.Destroyed) > 0 {
					destroyed++
				}
				b := f.Battlefield()
				if a, ok := b.Units[f.Attacker]; ok && e.ReturnFire &&
					p.Distance(b.Width, a.Cell, f.Target) != 1 {
					transposed++
				}
			}
			// The corpus must reach the behaviours it stands for.
			for name, n := range map[string]int{"skipped order": skipped, "ranged": noReturn, "pincer": pincer,
				"blocking cap": blocked, "destruction": destroyed, "transposed neighbour": transposed} {
				if n == 0 {
					t.Errorf("no fixture with %s", name)
				}
			}
			t.Logf("%d engagements", len(fs))
		})
	}
}

// The recorded casualty-routine runs cover the R1 input classes.
func TestCasualtyFixtures(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			fs := load[CasualtyFixture](t, build, "casualties.json")
			seen := map[string]bool{}
			divide := 0
			for i := range fs {
				f := &fs[i]
				checkHeader(t, p, f.Schema, f.Profile, f.ID, f.Source, seen)
				a, d, c, err := f.Inputs(p.Tables)
				if err != nil {
					t.Fatalf("%s: %v", f.ID, err)
				}
				if got := CasualtyResult(p.Casualties(a, d, c)); got != f.Expected {
					t.Errorf("%s:\nrecorded %+v\nresolved %+v", f.ID, f.Expected, got)
				}
				if f.Expected.Error != "" {
					divide++
				}
			}
			if divide == 0 {
				t.Error("no fixture with a divide error")
			}
			t.Logf("%d casualty runs, %d divide errors", len(fs), divide)
		})
	}
}

// Identical battle states give identical results: there is no random draw,
// and resolving one battle leaves no state behind for the next.
func TestRepeatable(t *testing.T) {
	for build, p := range profiles {
		fs := load[EngagementFixture](t, build, "engagements.json")
		first := make([]EngagementExpected, len(fs))
		for i := range fs {
			first[i] = EngagementResult(p.Engage(fs[i].Battlefield(), fs[i].Attacker, fs[i].Target))
		}
		for i := len(fs) - 1; i >= 0; i-- {
			if got := EngagementResult(p.Engage(fs[i].Battlefield(), fs[i].Attacker, fs[i].Target)); !reflect.DeepEqual(got, first[i]) {
				t.Errorf("%s/%s: %+v, then %+v", build, fs[i].ID, first[i], got)
			}
		}
	}
}

// TransposedDistance puts one of the six neighbours at distance 2; without
// it every neighbour is at distance 1, and every cell's distance is the
// number of steps through neighbours.
func TestDistanceSwitch(t *testing.T) {
	const w, h = 9, 9
	layout := Profile{Tables: DemoTables}
	for cell := range w * h {
		far := 0
		for _, n := range neighbours(cell, w, h) {
			if n < 0 {
				continue
			}
			if d := layout.Distance(w, cell, n); d != 1 {
				t.Errorf("cell %d: neighbour %d at layout distance %d", cell, n, d)
			}
			if Demo.Distance(w, cell, n) != 1 {
				far++
			}
		}
		x, y := cell%w, cell/w
		if x > 0 && x < w-1 && y > 0 && y < h-1 && far != 1 {
			t.Errorf("cell %d: %d neighbours not at battle distance 1", cell, far)
		}
	}
	// Breadth-first steps from the centre match the layout distance.
	steps := map[int]int{40: 0}
	queue := []int{40}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		for _, n := range neighbours(c, w, h) {
			if _, ok := steps[n]; n >= 0 && !ok {
				steps[n] = steps[c] + 1
				queue = append(queue, n)
			}
		}
	}
	for c, s := range steps {
		if d := layout.Distance(w, 40, c); d != s {
			t.Errorf("cell %d: layout distance %d, %d steps", c, d, s)
		}
	}
}
