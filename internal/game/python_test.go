package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// pythonRef is testdata/python.json, written by testdata/gen_python.py from the
// rule functions of gui.py.
type pythonRef struct {
	Levels []struct {
		Level  string  `json:"level"`
		Player Player  `json:"player"`
		Fog    [][]int `json:"fog"` // calculateFOW: 0 visible, 1 never seen, 2 seen
		Moves  []struct {
			From    [2]int              `json:"from"`
			Unit    string              `json:"unit"`
			In      string              `json:"in"` // map, carrier or building
			Row     int                 `json:"row"`
			Col     int                 `json:"col"`
			Targets map[string][][2]int `json:"targets"` // "line,col" → findPath result
		} `json:"moves"`
	} `json:"levels"`
	Battles []struct {
		Attacker       string `json:"attacker"`
		Defender       string `json:"defender"`
		Slot           int    `json:"slot"`
		Steps          int    `json:"steps"`
		AttackerExp    int    `json:"attackerExp"`
		DefenderExp    int    `json:"defenderExp"`
		AttackerSize   int    `json:"attackerSize"`
		DefenderSize   int    `json:"defenderSize"`
		DefenderEmpty  []int  `json:"defenderEmpty"` // defender slots set to ammo 0
		AttackerBattle int    `json:"attackerBattle"`
		DefenderBattle int    `json:"defenderBattle"`
		Rolls          [2]int `json:"rolls"`
		Counter        int    `json:"counter"` // defender slot that fires back, -1 for none
		AttackerKills  int    `json:"attackerKills"`
		DefenderKills  int    `json:"defenderKills"`
	} `json:"battles"`
}

func loadPython(t *testing.T) *pythonRef {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "python.json"))
	if err != nil {
		t.Fatal(err)
	}
	var ref pythonRef
	if err := json.Unmarshal(raw, &ref); err != nil {
		t.Fatal(err)
	}
	if len(ref.Levels) < 8 || len(ref.Battles) == 0 {
		t.Fatalf("python.json has %d level runs and %d battles", len(ref.Levels), len(ref.Battles))
	}
	return &ref
}
