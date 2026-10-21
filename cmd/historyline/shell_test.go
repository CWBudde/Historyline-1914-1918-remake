package main

import (
	"testing"

	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
)

func testShell(t *testing.T) *shell {
	t.Helper()
	maps, err := loadCampaign("../../levels")
	if err != nil {
		t.Fatal(err)
	}
	s := newShell(maps, matchConfig{opponent: "ai"})
	s.open(pageMain)
	return s
}

func TestCampaignOrderAndCodes(t *testing.T) {
	s := testShell(t)
	var codes []string
	for _, m := range s.maps {
		codes = append(codes, m.code())
	}
	if len(codes) != 3 || codes[0] != "PULSE" || codes[1] != "CIVIL" || codes[2] != "MOUSE" {
		t.Fatalf("codes %v", codes)
	}
}

func TestMainMenuNavigationWraps(t *testing.T) {
	s := testShell(t)
	if s.focus != 0 || s.buttons()[0].label != "START" {
		t.Fatalf("focus %d on %q", s.focus, s.buttons()[s.focus].label)
	}
	s.key(keyUp)
	if got := s.buttons()[s.focus].label; got != "EXIT" {
		t.Fatalf("up from START: %q", got)
	}
	s.key(keyDown)
	if s.focus != 0 {
		t.Fatalf("down from EXIT: %d", s.focus)
	}
}

func TestDisabledRowsAreSkippedAndInert(t *testing.T) {
	s := testShell(t)
	s.activate(2) // GAME VALUES
	if s.page != pageValues {
		t.Fatalf("page %d", s.page)
	}
	s.key(keyDown) // PLAYER A is disabled
	if got := s.buttons()[s.focus].label; got != "PLAYER B  COMPUTER" {
		t.Fatalf("down from BATTLES: %q", got)
	}
	s.activate(1)
	if s.page != pageValues || s.buttons()[1].label != "PLAYER A  HUMAN" {
		t.Fatal("disabled row did something")
	}
	s.key(keyBack)
	if s.page != pageMain {
		t.Fatalf("back: page %d", s.page)
	}
}

func TestCodeEntry(t *testing.T) {
	s := testShell(t)
	s.activate(1)
	for _, r := range "mo1use" {
		s.typeRune(r)
	}
	if s.typed != "MOUSE" {
		t.Fatalf("typed %q", s.typed)
	}
	s.key(keyEnter)
	if s.typing || s.code != 2 {
		t.Fatalf("typing %v, code %d", s.typing, s.code)
	}
	s.activate(1)
	for _, r := range "nowhere" {
		s.typeRune(r)
	}
	s.key(keyErase)
	s.key(keyEnter)
	if s.code != 2 || s.help()[0] != "Unknown code NOWHER." {
		t.Fatalf("code %d, help %q", s.code, s.help())
	}
	s.activate(1)
	s.typeRune('x')
	s.key(keyBack)
	if s.typing || s.buttons()[1].label != "CODE  MOUSE" {
		t.Fatalf("escape: typing %v, %q", s.typing, s.buttons()[1].label)
	}
}

func TestGameValuesMakeTheMatch(t *testing.T) {
	s := testShell(t)
	s.activate(2)
	s.activate(0) // battles short
	s.activate(2) // player B passive
	s.activate(4) // 5 moves a phase
	s.key(keyBack)
	s.activate(3) // control
	s.activate(1) // scrolling fast
	s.key(keyBack)
	s.activate(0) // start: the country page, Germany for a Central Powers map
	if s.page != pageCountry || s.focus != 1 {
		t.Fatalf("page %d, focus %d", s.page, s.focus)
	}
	s.key(keyLeft)
	s.key(keyEnter)
	if s.page != pageLoading {
		t.Fatalf("page %d", s.page)
	}
	c := s.config(matchConfig{rules: "dos"})
	if c.lvl != s.maps[0].lvl || c.player != state.Entente || c.opponent != "pass" || c.limit != 5 ||
		!c.reduced || c.scroll != 2*scrollSpeed || c.rules != "dos" {
		t.Fatalf("config %+v", c)
	}
}

func TestVictorySelectsTheNextMap(t *testing.T) {
	s := testShell(t)
	s.finish(outcome{won: true})
	if s.page != pageOutcome || s.code != 1 || s.result.next != "CIVIL" {
		t.Fatalf("page %d, code %d, next %q", s.page, s.code, s.result.next)
	}
	s.finish(outcome{})
	if s.code != 1 || s.result.next != "" {
		t.Fatal("a defeat changed the map")
	}
	s.code = len(s.maps) - 1
	s.finish(outcome{won: true})
	if s.code != len(s.maps)-1 || s.result.next != "" {
		t.Fatal("the last map has no next one")
	}
	s.key(keyEnter)
	if s.page != pageMain {
		t.Fatalf("continue: page %d", s.page)
	}
}

func TestLeaveBattle(t *testing.T) {
	s := testShell(t)
	s.open(pageLeave)
	if got := s.buttons()[s.focus].label; got != "STAY" {
		t.Fatalf("focus on %q", got)
	}
	s.key(keyBack)
	if s.page != pageBattle {
		t.Fatalf("escape: page %d", s.page)
	}
	s.open(pageLeave)
	s.activate(1)
	if s.page != pageMain {
		t.Fatalf("leave: page %d", s.page)
	}
}

func TestHitFindsButtons(t *testing.T) {
	s := testShell(t)
	r := row(2)
	if i := s.hit(r.X+1, r.Y+1); i != 2 {
		t.Fatalf("hit %d", i)
	}
	if i := s.hit(600, 10); i != -1 {
		t.Fatalf("hit %d outside", i)
	}
}
