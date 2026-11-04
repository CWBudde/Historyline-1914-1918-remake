package dosrules

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var updateSaves = flag.Bool("update", false, "rewrite testdata/saves from the current rules")

// TestSaveGolden pins saves written by Go in testdata/saves, which the web
// edition's port loads and must write back byte for byte: per build a round
// in progress (midRound) and the same game after the round resolved, and a
// winter game with containers and an HQ (containerScene). go test -update
// rewrites them.
func TestSaveGolden(t *testing.T) {
	cases := map[string]func(t *testing.T) *Game{
		"full-winter-containers": func(t *testing.T) *Game { return containerScene(t, Full.Season(true)) },
		"demo-ended": func(t *testing.T) *Game {
			g := containerScene(t, Demo)
			g.Buildings = map[int]Building{HoldFlag40: {Credits: 7, Builds: []int{cSupply, cInfantry}}}
			g.Energy = [2]int{120, 35}
			g.Players[1].Computer = true
			g.Ended = &End{Over: true, Winner: 0, Score: g.Score()}
			return g
		},
	}
	for build, p := range profiles {
		cases[build+"-midround"] = func(t *testing.T) *Game { return midRound(t, p) }
		cases[build+"-resolved"] = func(t *testing.T) *Game {
			g := midRound(t, p)
			g.Confirm(0)
			if _, err := g.Resolve(); err != nil {
				t.Fatal(err)
			}
			return g
		}
	}
	for name, scene := range cases {
		t.Run(name, func(t *testing.T) {
			b, err := scene(t).Save()
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join("testdata", "saves", name+".json")
			if *updateSaves {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, b, 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("%v (go test -run TestSaveGolden -update writes it)", err)
			}
			if !bytes.Equal(b, want) {
				t.Errorf("%s differs from the save the rules write now; rerun with -update if the change is intended", path)
			}
			if _, err := LoadGame(want); err != nil {
				t.Errorf("%s does not load: %v", path, err)
			}
		})
	}
}
