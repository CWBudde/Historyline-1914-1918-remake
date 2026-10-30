package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
)

// dig has the first sapper of original map 03 dig next to it on the session
// of gm and returns the work's event.
func dig(t *testing.T, gm *game) dosplay.Event {
	t.Helper()
	s := gm.dos.s
	for side := range 2 {
		for _, u := range s.TrueView().Units {
			if u.Side != side || !s.Sapper(u.Index) {
				continue
			}
			for s.Phase(side) != dosrules.PhaseAction {
				s.Confirm(0)
				s.Confirm(1)
				if _, err := s.Resolve(); err != nil {
					t.Fatal(err)
				}
			}
			for _, to := range dosplay.Coords(s.Targets(side, u.Index)) {
				if _, taken := s.View(side).Units[to]; taken || s.Level.Lines[to.Line][to.Col].Terrain != "grass" {
					continue
				}
				evs, err := s.Attack(side, u.Index, to)
				if err != nil || len(evs) != 1 || evs[0].Kind != dosplay.Tasked {
					t.Fatalf("work order: %+v, %v", evs, err)
				}
				s.Confirm(0)
				s.Confirm(1)
				evs, err = s.Resolve()
				if err != nil {
					t.Fatal(err)
				}
				for _, ev := range evs {
					if ev.Kind == dosplay.Dug {
						return ev
					}
				}
				t.Fatalf("no work in %+v", evs)
			}
		}
	}
	t.Fatal("no sapper with free grass beside it")
	return dosplay.Event{}
}

// A sapper's work, once it plays back, changes the map drawn: the level copy
// and the composed terrain, not the session's level.
func TestSapperWorkRedrawsTerrain(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	lvl, b, err := loadMap(t.TempDir(), "original/03", set)
	skipMissingOriginal(t, err)
	if err != nil {
		t.Fatal(err)
	}
	player, err := state.PlayerByName(lvl.Player)
	if err != nil {
		t.Fatal(err)
	}
	gm, err := newMatch(matchConfig{lvl: lvl, board: b, set: set, player: player, rules: "dos", opponent: "pass"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	ev := dig(t, gm)
	// The catalog holds the tiles of the map as the work leaves it.
	after := reworked(lvl, gm.dos.s.Reworked())
	if gm.terrain, err = openTerrainArt(terrainCatalog(t, after, ""), assets.Options{AllowCandidates: true}, lvl); err != nil {
		t.Fatal(err)
	}
	trench := func(ids []string) bool {
		return slices.ContainsFunc(ids, func(id string) bool { return strings.Contains(id, "trench") })
	}
	if trench(gm.terrain.scene.Cells[ev.To.Line][ev.To.Col].Overlays) {
		t.Fatal("a trench before the work played")
	}

	gm.dos.queue = []dosplay.Event{ev}
	gm.dos.step()
	if !gm.dos.rework || gm.dos.Message == "" {
		t.Fatalf("the work played: rework %v, message %q", gm.dos.rework, gm.dos.Message)
	}
	gm.reworkTerrain()
	for _, c := range ev.Hexes {
		if !trench(gm.lvl.Lines[c.Line][c.Col].Infra.IDs) || !trench(gm.terrain.scene.Cells[c.Line][c.Col].Overlays) {
			t.Errorf("hex %v: level %+v, overlays %v", c, gm.lvl.Lines[c.Line][c.Col], gm.terrain.scene.Cells[c.Line][c.Col].Overlays)
		}
		if trench(lvl.Lines[c.Line][c.Col].Infra.IDs) {
			t.Errorf("hex %v: the session's level changed", c)
		}
	}
}
