package dosrules

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"
)

// midRound is a round in progress: side 0 has moved a unit but not
// confirmed, side 1 has queued an attack on the cell where its map still
// shows that unit and has confirmed.
func midRound(t *testing.T, p Profile) *Game {
	t.Helper()
	n := neighbours(40, cycleW, cycleH)
	g := newCycle(t, p, map[int]Placement{
		0: {Type: cInfantry, Owner: 0, Cell: 40}, 1: {Type: cPanzer, Owner: 1, Cell: n[0]},
		2: {Type: cInfantry, Owner: 0, Cell: 70}, 3: {Type: cInfantry, Owner: 1, Cell: 10, Exp: 3},
	})
	g.Limit, g.Concealed = 8, true
	dests, err := p.Destinations(g.Movement(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	dest := -1
	for _, c := range dests {
		if !slices.Contains(n[:], c) && c != 40 {
			dest = c
			break
		}
	}
	if dest < 0 {
		t.Fatalf("no destination away from cell 40 in %v", dests)
	}
	apply(t, g, Command{Kind: CmdMove, Side: 0, Unit: 0, Cell: dest})
	apply(t, g, Command{Kind: CmdAttack, Side: 1, Unit: 1, Cell: 40})
	g.Confirm(1)
	return g
}

func reload(t *testing.T, g *Game) *Game {
	t.Helper()
	b, err := g.Save()
	if err != nil {
		t.Fatal(err)
	}
	h, err := LoadGame(b)
	if err != nil {
		t.Fatalf("load: %v\n%s", err, b)
	}
	return h
}

// TestSaveResume saves a round with an unconfirmed side and a confirmed side
// with a queued attack: the loaded game equals the saved one, and both run
// the rest of the round to the same events and state.
func TestSaveResume(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			g := midRound(t, p)
			if g.Players[0].Ready || !g.Players[1].Ready || len(g.Players[0].Orders) != 1 || len(g.Players[1].Orders) != 1 {
				t.Fatalf("set-up: %+v", g.Players)
			}
			h := reload(t, g)
			if !reflect.DeepEqual(g, h) {
				t.Fatalf("loaded game differs:\n got %+v\nwant %+v", h, g)
			}

			var evs [2][]Event
			for k, x := range []*Game{g, h} {
				x.Confirm(0)
				e, err := x.Resolve()
				if err != nil {
					t.Fatal(err)
				}
				evs[k] = e
			}
			if len(evs[0]) == 0 || evs[0][0].Kind != EvBattle || !reflect.DeepEqual(evs[0], evs[1]) {
				t.Errorf("events differ:\n got %+v\nwant %+v", evs[1], evs[0])
			}
			if !reflect.DeepEqual(g, h) {
				t.Errorf("state after the round differs")
			}

			// A save after the phase change reloads equal too.
			if r := reload(t, g); !reflect.DeepEqual(g, r) {
				t.Errorf("reloaded after the round differs:\n got %+v\nwant %+v", r, g)
			}
		})
	}
}

// TestLoadGameRefuses checks that a save that does not fit together is
// refused rather than loaded.
func TestLoadGameRefuses(t *testing.T) {
	g := midRound(t, Full)
	b, err := g.Save()
	if err != nil {
		t.Fatal(err)
	}
	for what, edit := range map[string]func(m map[string]any){
		"another schema":     func(m map[string]any) { m["schema"] = 1 },
		"an unknown build":   func(m map[string]any) { m["build"] = "00000000" },
		"a truncated map":    func(m map[string]any) { mp := m["maps"].([]any); mp[1] = mp[1].([]any)[:10] },
		"an unknown type":    func(m map[string]any) { m["units"].(map[string]any)["0"].(map[string]any)["type"] = 99 },
		"a missing occupant": func(m map[string]any) { delete(m["units"].(map[string]any), "1") },
		"an unknown order":   func(m map[string]any) { m["players"].([]any)[1].(map[string]any)["orders"] = []int{77} },
		"a phase pair":       func(m map[string]any) { m["players"].([]any)[0].(map[string]any)["phase"] = 2 },
	} {
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatal(err)
		}
		edit(m)
		bad, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := LoadGame(bad); err == nil {
			t.Errorf("%s: loaded", what)
		}
	}
}
