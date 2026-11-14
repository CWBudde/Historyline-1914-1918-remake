package main

import (
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosai"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/saves"
)

// The tests below save a hot-seat battle while one human has a resolution
// still to see, resume it, and check the playback against the battle that
// went on without the save.

// battleEve is August 1914's two-player setup just before a resolution that
// fights a battle: the computer gave both sides' orders up to there and both
// have confirmed. The session is loaded from the save before that
// resolution, as the web edition loads it.
func battleEve(t *testing.T) *dosplay.Session {
	t.Helper()
	lvl, err := loadLevel("", "1914-08:2p")
	if err != nil {
		t.Fatal(err)
	}
	s, err := dosplay.New(lvl, testSet(t), dosplay.Options{})
	if err != nil {
		t.Fatal(err)
	}
	for range 60 {
		for side := range 2 {
			if _, err := (&dosai.Planner{}).Plan(s, side); err != nil {
				t.Fatal(err)
			}
		}
		before, err := s.Save()
		if err != nil {
			t.Fatal(err)
		}
		evs, err := s.Resolve()
		if err != nil {
			t.Fatal(err)
		}
		if slices.ContainsFunc(evs, func(ev dosplay.Event) bool { return ev.Kind == dosplay.Battle }) {
			r, err := dosplay.Load(lvl, testSet(t), before)
			if err != nil {
				t.Fatal(err)
			}
			return r
		}
	}
	t.Fatal("no battle in sixty resolutions")
	return nil
}

// pendingGame is the hot-seat battle at battleEve after the left human took
// over and saw the resolution: the right human has it still to see, and the
// left one gives its orders.
func pendingGame(t *testing.T) (*game, *dosControl) {
	t.Helper()
	s := battleEve(t)
	g := newDosGame(s.Level, s, 0, nil, false)
	g.setLocal(localHotseat)
	d := g.dos
	g.curtainInput(true, false)
	d.Tick()
	if d.pending[1] == nil || !slices.ContainsFunc(d.pending[1].events, func(ev dosplay.Event) bool { return ev.Kind == dosplay.Battle }) {
		t.Fatal("the right human keeps no resolution with a battle")
	}
	playLocal(t, d)
	return g, d
}

// saveGame saves g's battle into the save file at path.
func saveGame(t *testing.T, g *game, path string) saves.Envelope {
	t.Helper()
	s := localShell(t)
	s.localOn, s.play = true, g.dos.local
	store := saves.OpenFile(path)
	a := &app{shell: s, base: matchConfig{saves: store}, battle: g, press: -1}
	if err := a.saveCurrentBattle(); err != nil {
		t.Fatal(err)
	}
	env, ok := store.Get()
	if !ok {
		t.Fatal("the save does not read")
	}
	return env
}

// resumeGame resumes the battle of the save file at path through LOAD GAME.
func resumeGame(t *testing.T, path string) (*game, *dosControl) {
	t.Helper()
	s := localShell(t)
	s.saves = saves.OpenFile(path)
	s.open(pageExtras)
	s.activate(1) // LOAD GAME
	if s.page != pageLoading || s.play != localHotseat {
		t.Fatalf("LOAD GAME: page %d, mode %v", s.page, s.play)
	}
	gm, err := newDosMatch(s.config(matchConfig{set: testSet(t), opponent: "ai"}))
	if err != nil {
		t.Fatal(err)
	}
	return gm, gm.dos
}

// playback is what side's takeover plays: the view shown until the battles
// have played and the events.
type playback struct {
	before dosplay.View
	events []dosplay.Event
}

// takeOverAs lets side's human take over behind its curtain and returns the
// playback that starts.
func takeOverAs(t *testing.T, g *game, side int) playback {
	t.Helper()
	d := g.dos
	if c, ok := d.curtain(); !ok || c != side {
		t.Fatalf("the curtain waits for %d (%v), want %d", c, ok, side)
	}
	g.curtainInput(true, false)
	if d.side != side || !d.busy() || d.frozen == nil || d.stage != stageBlack || d.pending[side] != nil {
		t.Fatalf("side %d's takeover: served %d, busy %v, stage %v, still kept %v", side, d.side, d.busy(), d.stage, d.pending[side] != nil)
	}
	return playback{before: *d.frozen, events: slices.Clone(d.queue)}
}

// TestHotseatSavePending: a save made while the right human has a resolution
// still to see, during the left human's turn or behind the right human's
// curtain, keeps it; the resumed battle plays it at the right human's
// takeover from the view before, exactly as the battle without the save.
func TestHotseatSavePending(t *testing.T) {
	g, d := pendingGame(t)
	dir := t.TempDir()
	turn := filepath.Join(dir, "turn.json")
	if env := saveGame(t, g, turn); env.Side != 0 || env.Mode != "hotseat" || env.Pending == nil || env.Pending.Side != 1 {
		t.Fatalf("the save in the left human's turn: %+v", env)
	}
	confirm(d)
	curtain := filepath.Join(dir, "curtain.json")
	if env := saveGame(t, g, curtain); env.Side != 1 || env.Pending == nil || env.Pending.Side != 1 {
		t.Fatalf("the save behind the right human's curtain: %+v", env)
	}
	want := takeOverAs(t, g, 1)

	gm, r := resumeGame(t, curtain)
	if r.pending[1] == nil || r.pending[0] != nil {
		t.Fatalf("resumed behind the curtain: kept %v %v", r.pending[0] != nil, r.pending[1] != nil)
	}
	if got := takeOverAs(t, gm, 1); !reflect.DeepEqual(got, want) {
		t.Error("the resumed playback differs from the one without the save")
	}
	playLocal(t, r)

	gm, r = resumeGame(t, turn)
	g.curtainInput(true, false)
	gm.curtainInput(true, false)
	if r.side != 0 || r.busy() || r.pending[1] == nil {
		t.Fatalf("resumed in the left human's turn: served %d, busy %v, kept %v", r.side, r.busy(), r.pending[1] != nil)
	}
	confirm(r)
	if got := takeOverAs(t, gm, 1); !reflect.DeepEqual(got, want) {
		t.Error("the playback resumed from the left human's turn differs from the one without the save")
	}
}
