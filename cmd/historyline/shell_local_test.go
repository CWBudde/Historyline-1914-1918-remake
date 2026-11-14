package main

import (
	"path/filepath"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/faction"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/saves"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// The tests below start two humans' battles from the menu and resume them.

// localShell is the menu with every month's two-player setup.
func localShell(t *testing.T) *shell {
	t.Helper()
	s := testShell(t)
	locals, err := loadLocals("../../levels", testSet(t))
	if err != nil {
		t.Fatal(err)
	}
	s.locals = locals
	return s
}

// TestLocalsEveryMonth: every campaign month has a two-player setup.
func TestLocalsEveryMonth(t *testing.T) {
	s := localShell(t)
	months, err := scenario.Months(levelsFS("../../levels"))
	if err != nil {
		t.Fatal(err)
	}
	for _, month := range months {
		m, ok := s.locals[month]
		if !ok || m.lvl == nil || m.name != month+":2p" || m.players != 2 {
			t.Errorf("%s: two-player setup %+v", month, m.name)
		}
	}
}

// TestMenuStartsTwoHumans: the play mode of the game values starts the
// month's two-player setup for two humans, through its front page and
// without country or unit cards, in either mode, and no campaign map follows.
func TestMenuStartsTwoHumans(t *testing.T) {
	s := localShell(t)
	for _, want := range []localMode{localSimultaneous, localHotseat} {
		s.open(pageValues)
		s.activate(5)
		if bs := s.buttons(); s.play != want || !bs[2].off || bs[2].label != "PLAYER B  HUMAN" {
			t.Fatalf("play mode %v, player B %q off %v; want %v", s.play, bs[2].label, bs[2].off, want)
		}
		s.key(keyBack)
		s.activate(0) // START
		if s.page != pageNews || !s.localOn {
			t.Fatalf("START with two humans: page %d, local %v", s.page, s.localOn)
		}
		s.activate(0) // CONTINUE
		if s.page != pageLoading {
			t.Fatalf("after the front page: page %d, want loading", s.page)
		}
		c := s.config(matchConfig{set: testSet(t), opponent: "ai"})
		if c.lvl != s.locals["1914-08"].lvl || c.local != want || c.player != faction.CentralPowers {
			t.Fatalf("config: %s, mode %v, player %v", c.lvl.MapName, c.local, c.player)
		}
		gm, err := newDosMatch(c)
		if err != nil {
			t.Fatal(err)
		}
		d := gm.dos
		d.Tick()
		if d.local != want || d.s.Game.Players[1].Computer || d.s.Ready(1) {
			t.Fatalf("the battle: mode %v, computer %v, ready %v", d.local, d.s.Game.Players[1].Computer, d.s.Ready(1))
		}
		if _, ok := d.curtain(); ok != (want == localHotseat) {
			t.Errorf("%v starts behind a curtain: %v", want, ok)
		}
		s.finish(outcome{won: true})
		if s.result.next != "" {
			t.Errorf("two humans' battle leads on to %q", s.result.next)
		}
		s.open(pageMain)
	}
	// Back to one human: START picks a country again.
	s.open(pageValues)
	s.activate(5)
	s.key(keyBack)
	s.activate(0)
	if s.play != localNone || s.page != pageCountry || s.localOn {
		t.Fatalf("one human: mode %v, page %d, local %v", s.play, s.page, s.localOn)
	}
	if c := s.config(matchConfig{}); c.local != localNone || c.lvl != s.maps[s.code].lvl {
		t.Errorf("one human's config: mode %v", c.local)
	}
}

// TestHotseatSaveResume: a save made at the handover names the next human
// and the mode, and LOAD GAME resumes behind that human's curtain.
func TestHotseatSaveResume(t *testing.T) {
	g, d := newLocalTest(t, localHotseat)
	g.curtainInput(true, false)
	moveOne(t, d)
	confirm(d)
	s := localShell(t)
	s.localOn, s.play = true, localHotseat
	store := saves.OpenFile(filepath.Join(t.TempDir(), "save.json"))
	s.saves = store
	a := &app{shell: s, base: matchConfig{saves: store}, battle: g, press: -1}
	if err := a.saveCurrentBattle(); err != nil {
		t.Fatal(err)
	}
	env, ok := store.Get()
	if !ok || env.Map != "1914-08:2p" || env.Side != 1 || env.Mode != "hotseat" {
		t.Fatalf("saved %+v", env)
	}
	s.localOn, s.play = false, localNone
	s.open(pageExtras)
	s.activate(1) // LOAD GAME
	if s.page != pageLoading || !s.localOn || s.play != localHotseat {
		t.Fatalf("LOAD GAME: page %d, local %v, mode %v", s.page, s.localOn, s.play)
	}
	c := s.config(matchConfig{set: testSet(t), opponent: "ai"})
	if c.saved == nil || c.local != localHotseat || c.player != faction.Entente || c.lvl != s.locals["1914-08"].lvl {
		t.Fatalf("the resumed config: saved %v, mode %v, player %v", c.saved != nil, c.local, c.player)
	}
	gm, err := newDosMatch(c)
	if err != nil {
		t.Fatal(err)
	}
	r := gm.dos
	if side, ok := r.curtain(); !ok || side != 1 || r.side != 1 || !r.s.Ready(0) || r.s.Ready(1) {
		t.Fatalf("resumed: curtain %d %v, served %d, ready %v %v", side, ok, r.side, r.s.Ready(0), r.s.Ready(1))
	}
	gm.curtainInput(true, false)
	if r.covered(1) || !r.covered(0) {
		t.Errorf("after the takeover: covered %v %v", r.covered(0), r.covered(1))
	}
}

// TestCodeWordPlay: a code word selects its month and PLAY picks the mode:
// one human starts the month's map of the code's side, two humans its
// two-player setup in their mode. A code only a two-player setup has
// switches one human to TWO HUMANS and starts that setup.
func TestCodeWordPlay(t *testing.T) {
	enter := func(s *shell, code string) {
		s.open(pageMain)
		s.typing, s.typed = true, code
		s.enterCode()
	}
	for _, tc := range []struct {
		code       string
		play, want localMode
		month      string
	}{
		{"TRACK", localNone, localNone, "1915-05"},
		{"TRACK", localSimultaneous, localSimultaneous, "1915-05"},
		{"TRACK", localHotseat, localHotseat, "1915-05"},
		{"WYTSC", localNone, localSimultaneous, "1914-11"},
		{"WYTSC", localSimultaneous, localSimultaneous, "1914-11"},
		{"WYTSC", localHotseat, localHotseat, "1914-11"},
	} {
		s := localShell(t)
		s.play = tc.play
		enter(s, tc.code)
		if m := s.maps[s.code]; m.file != tc.month || s.play != tc.want {
			t.Errorf("%s with %v: month %s, play %v; want %s, %v (%q)", tc.code, tc.play, m.file, s.play, tc.month, tc.want, s.note)
			continue
		}
		s.activate(0) // START
		c := s.config(matchConfig{set: testSet(t), opponent: "ai"})
		switch tc.want {
		case localNone:
			if s.page != pageCountry || s.localOn || c.local != localNone || c.lvl != s.maps[s.code].lvl || s.maps[s.code].code() != tc.code {
				t.Errorf("%s with one human: page %d, local %v, map %s", tc.code, s.page, s.localOn, s.maps[s.code].name)
			}
		default:
			if !s.localOn || c.local != tc.want || c.lvl != s.locals[tc.month].lvl {
				t.Errorf("%s with %v: local %v, mode %v, level %s", tc.code, tc.want, s.localOn, c.local, c.lvl.MapName)
			}
		}
	}
}
