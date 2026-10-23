package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
)

var shellSet *data.Set

func testSet(t *testing.T) *data.Set {
	t.Helper()
	if shellSet == nil {
		set, err := data.Load()
		if err != nil {
			t.Fatal(err)
		}
		shellSet = set
	}
	return shellSet
}

func testShell(t *testing.T) *shell {
	t.Helper()
	maps, err := loadCampaign("../../levels", testSet(t))
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
	// The shipped levels, then the original maps 00-71.
	if len(codes) != 75 || codes[0] != "PULSE" || codes[1] != "CIVIL" || codes[2] != "MOUSE" ||
		codes[3] != "00" || codes[74] != "71" {
		t.Fatalf("codes %v", codes)
	}
	for i, m := range s.maps {
		if (m.board != nil) != (i >= 3) {
			t.Errorf("%s: board %v", m.name, m.board != nil)
		}
	}
	// An Allies campaign map is played as the Entente.
	s.code = 3 + 30
	if s.levelSide() != state.Entente {
		t.Errorf("map 30 side %v", s.levelSide())
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
	for _, r := range "mo-use" {
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
	// An original map's code is its number.
	s.activate(1)
	for _, r := range "05" {
		s.typeRune(r)
	}
	s.key(keyEnter)
	if s.maps[s.code].name != "original/05" {
		t.Fatalf("05 selected %s", s.maps[s.code].name)
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
	if c.lvl != s.maps[0].lvl || c.board != nil || c.player != state.Entente || c.opponent != "pass" || c.limit != 5 ||
		!c.reduced || c.scroll != 2*scrollSpeed || c.rules != "dos" {
		t.Fatalf("config %+v", c)
	}
	// An original map plays its exact board.
	s.code = 3 + 6
	if c := s.config(matchConfig{rules: "dos"}); c.lvl != s.maps[9].lvl || c.board == nil || c.board != s.maps[9].board {
		t.Fatalf("original map config %+v", c)
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
	// Mouse is original map 02: a victory leads on to map 03, and the
	// campaigns run to their last map.
	for _, c := range []struct {
		from int
		next string
	}{{2, "03"}, {3 + 3, "04"}, {3 + 22, "23"}, {3 + 23, ""}, {3 + 24, "25"}, {3 + 47, ""}, {3 + 48, "49"}, {3 + 71, ""}} {
		s.code = c.from
		s.finish(outcome{won: true})
		if s.result.next != c.next || c.next == "" && s.code != c.from {
			t.Errorf("victory on %s: next %q, map %s", s.maps[c.from].name, s.result.next, s.maps[s.code].name)
		}
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

func TestCodeEntryPasswords(t *testing.T) {
	s := testShell(t)
	pw := make([]string, 72)
	for i := range pw {
		pw[i] = fmt.Sprintf("PW%03d", i)
	}
	withPasswords(s.maps, pw)
	if s.maps[0].code() != "PULSE" || s.maps[3+5].code() != "PW005" {
		t.Fatalf("codes %q %q", s.maps[0].code(), s.maps[3+5].code())
	}
	enter := func(code string) {
		s.activate(1)
		for _, r := range code {
			s.typeRune(r)
		}
		s.key(keyEnter)
	}
	// The password selects its map; the number still does.
	enter("pw030")
	if s.maps[s.code].name != "original/30" || s.buttons()[1].label != "CODE  PW030" {
		t.Fatalf("pw030 selected %s, %q", s.maps[s.code].name, s.buttons()[1].label)
	}
	enter("05")
	if s.maps[s.code].name != "original/05" {
		t.Fatalf("05 selected %s", s.maps[s.code].name)
	}
	// A victory shows the next map's password.
	next := s.maps[s.code].next
	s.finish(outcome{won: true})
	if next < 0 || s.result.next != s.maps[next].password {
		t.Fatalf("next %d, shown %q", next, s.result.next)
	}
}

func TestLocalPasswords(t *testing.T) {
	pw := localPasswords()
	if pw == nil {
		t.Skip("no original files")
	}
	if len(pw) != 72 {
		t.Fatalf("%d passwords", len(pw))
	}
	for i, p := range pw {
		if len(p) != 5 || strings.ToUpper(p) != p {
			t.Errorf("map %d: password of %d letters", i, len(p))
		}
	}
}
