package dosrules

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// profiles maps each recorded build to its profile; testdata/<name> holds
// its fixtures.
var profiles = Profiles

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

// TestTables checks both builds' tables and that they differ only in the
// full build's winter table: any other difference between the builds would
// be a Profile field.
func TestTables(t *testing.T) {
	for _, c := range []struct {
		tab   *Tables
		build string
	}{{FullTables, "fe4fa446"}, {DemoTables, "83fc4e37"}} {
		if len(c.tab.Types) != 51 || c.tab.Build != c.build || c.tab.Move == nil || len(c.tab.Move.Terrains) != 175 {
			t.Fatalf("tables %q: %d types", c.tab.Build, len(c.tab.Types))
		}
	}
	if len(FullTables.Move.Winter) != 175 || DemoTables.Move.Winter != nil {
		t.Fatalf("winter tables: full %d records, demo %d", len(FullTables.Move.Winter), len(DemoTables.Move.Winter))
	}
	full, demo := *FullTables, *DemoTables
	move := *full.Move
	move.Winter = nil
	full.Move, full.Build, demo.Build = &move, "", ""
	if !reflect.DeepEqual(full, demo) {
		t.Error("the demo's tables differ from the full build's")
	}
	if Demo.TransposedDistance != Full.TransposedDistance {
		t.Error("the demo's switches differ from the full build's")
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
			for name, n := range map[string]int{
				"skipped order": skipped, "ranged": noReturn, "pincer": pincer,
				"blocking cap": blocked, "destruction": destroyed, "transposed neighbour": transposed,
			} {
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
	layout := Profile{Tables: FullTables}
	for cell := range w * h {
		far := 0
		for _, n := range neighbours(cell, w, h) {
			if n < 0 {
				continue
			}
			if d := layout.Distance(w, cell, n); d != 1 {
				t.Errorf("cell %d: neighbour %d at layout distance %d", cell, n, d)
			}
			if Full.Distance(w, cell, n) != 1 {
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
