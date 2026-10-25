package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/saves"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scores"
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
	// The shipped levels, then the original maps 00-71 by their passwords.
	if len(codes) != 75 || codes[0] != "PULSE" || codes[1] != "CIVIL" || codes[2] != "MOUSE" ||
		s.maps[3].name != "original/00" || codes[3] != "PULSE" || s.maps[74].name != "original/71" {
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

func TestCampaignBriefingsAndPlayers(t *testing.T) {
	s := testShell(t)
	if len(s.maps) != 75 {
		t.Fatalf("maps %d", len(s.maps))
	}
	for i, m := range s.maps {
		if m.briefing == "" {
			t.Errorf("map %d %s: no briefing", i, m.name)
		}
	}
	if s.maps[3+23].next != -1 || s.maps[3+47].next != -1 || s.maps[3+71].next != -1 {
		t.Error("campaign ends should have no next")
	}
	if s.maps[3+48].players != 2 || len(s.maps[3+48].humanSides) != 2 {
		t.Errorf("map 48 players %d sides %v", s.maps[3+48].players, s.maps[3+48].humanSides)
	}
	if !strings.Contains(s.maps[3+48].briefing, "two-player map") || !strings.Contains(s.maps[3+48].briefing, "not in v1") && !strings.Contains(s.maps[3+48].briefing, "in v1") {
		t.Errorf("2P briefing %q", s.maps[3+48].briefing)
	}
}

func TestSaveLoadWiring(t *testing.T) {
	s := testShell(t)
	set := testSet(t)
	// A real dos battle saved through the app callback loads via LOAD GAME.
	sess, err := dosplay.New(s.maps[0].lvl, set, dosplay.Options{})
	if err != nil {
		t.Fatal(err)
	}
	store := saves.OpenFile(filepath.Join(t.TempDir(), "save.json"))
	scoreStore := scores.OpenFile(filepath.Join(t.TempDir(), "scores.json"))
	if _, err := scoreStore.Record("PULSE", 42); err != nil {
		t.Fatal(err)
	}
	s.saves, s.scores, s.saveBattle = store, scoreStore, func() error {
		data, err := sess.Save()
		if err != nil {
			return err
		}
		return store.Put(saves.Envelope{Map: s.maps[s.code].name, Side: 0, Data: data})
	}
	// EXTRAS: LOAD starts disabled (no save), SCORE LIST opens.
	s.open(pageExtras)
	if !s.buttons()[1].off {
		t.Fatal("LOAD GAME enabled without a save")
	}
	s.activate(0)
	if s.page != pageScores || len(s.scoreList) != 1 || s.scoreList[0].Score != 42 {
		t.Fatalf("scores page %d list %v", s.page, s.scoreList)
	}
	s.key(keyBack)
	// SAVE AND LEAVE on the leave page writes the slot and returns to main.
	s.open(pageLeave)
	s.activate(1)
	if s.page != pageMain || !store.Has() {
		t.Fatalf("after save: page %d has %v", s.page, store.Has())
	}
	// LOAD GAME now resumes: config carries the saved bytes once.
	s.open(pageExtras)
	if s.buttons()[1].off {
		t.Fatal("LOAD GAME disabled with a save")
	}
	s.activate(1)
	if s.page != pageLoading {
		t.Fatalf("load page %d", s.page)
	}
	c := s.config(matchConfig{rules: "dos", set: set})
	if c.saved == nil || c.lvl != s.maps[0].lvl {
		t.Fatalf("loaded config saved %v lvl %v", c.saved != nil, c.lvl)
	}
	if _, err := dosplay.Load(c.lvl, set, c.saved); err != nil {
		t.Fatal(err)
	}
	if s.config(matchConfig{rules: "dos"}).saved != nil {
		t.Fatal("pending save was not consumed")
	}
	// An app with a battle saves through saveCurrentBattle.
	gm := newDosGame(s.maps[0].lvl, sess, 0, passPlanner{}, false)
	a := &app{shell: s, base: matchConfig{saves: store, rules: "dos"}, battle: gm, press: -1}
	if err := a.saveCurrentBattle(); err != nil {
		t.Fatal(err)
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
	s.key(keyDown) // RULES -> BATTLES
	if got := s.buttons()[s.focus].label; got != "BATTLES  FULL" {
		t.Fatalf("down from RULES: %q", got)
	}
	s.key(keyDown) // PLAYER A is disabled
	if got := s.buttons()[s.focus].label; got != "PLAYER B  COMPUTER" {
		t.Fatalf("down from BATTLES: %q", got)
	}
	s.activate(2)
	if s.page != pageValues || s.buttons()[2].label != "PLAYER A  HUMAN" {
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
	// The rules row cycles dos (default) -> standard -> dos,
	// and config carries the choice while the base says otherwise.
	if got := s.buttons()[0].label; got != "RULES  ORIGINAL (DOS)" {
		t.Fatalf("rules row %q", got)
	}
	for _, want := range []string{"standard", "dos"} {
		s.activate(0)
		if c := s.config(matchConfig{rules: "standard"}); c.rules != want {
			t.Fatalf("cycled rules %q, want %q", c.rules, want)
		}
	}
	if got := s.buttons()[0].label; got != "RULES  ORIGINAL (DOS)" {
		t.Fatalf("rules row after cycle %q", got)
	}
	s.activate(1) // battles short
	s.activate(3) // player B passive
	s.activate(5) // 5 moves a phase
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

func TestRulesChoiceStartsBothEngines(t *testing.T) {
	s := testShell(t)
	set := testSet(t)
	s.open(pageValues)
	for _, want := range []string{"dos", "standard"} {
		for s.rules != want {
			s.cycleRules()
		}
		c := s.config(matchConfig{set: set, opponent: "pass"})
		c.player = s.levelSide()
		gm, err := newMatch(c, nil)
		if err != nil {
			t.Fatalf("%s: %v", want, err)
		}
		if want == "dos" && gm.dos == nil {
			t.Fatalf("%s: no dos session", want)
		}
		if want != "dos" && gm.state == nil {
			t.Fatalf("%s: no remake state", want)
		}
	}
	// The removed Classic (Python) mode stays rejected.
	if _, err := newMatch(matchConfig{set: set, rules: "classic"}, nil); err == nil {
		t.Error("classic rules accepted")
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
		want, shown := s.maps[c.from].name, ""
		if c.next != "" {
			want, shown = originalPrefix+c.next, s.maps[s.code].password
		}
		if s.maps[s.code].name != want || s.result.next != shown {
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
	s.activate(2) // LEAVE (SAVE AND LEAVE is disabled without a battle)
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
	// Every original map has its own five-letter password; maps 00-02 share
	// theirs with the shipped levels transcribed from them.
	seen := map[string]bool{}
	for i, m := range s.maps[3:] {
		if len(m.password) != 5 || strings.ToUpper(m.password) != m.password || seen[m.password] {
			t.Errorf("map %02d: password %q", i, m.password)
		}
		seen[m.password] = true
		if i < 3 && m.password != s.maps[i].code() {
			t.Errorf("map %02d: password %q, shipped %q", i, m.password, s.maps[i].code())
		}
	}
	enter := func(code string) {
		s.activate(1)
		for _, r := range code {
			s.typeRune(r)
		}
		s.key(keyEnter)
	}
	// The password selects its map; the number still does.
	enter(strings.ToLower(s.maps[3+30].password))
	if s.maps[s.code].name != "original/30" || s.buttons()[1].label != "CODE  "+s.maps[3+30].password {
		t.Fatalf("map 30's password selected %s, %q", s.maps[s.code].name, s.buttons()[1].label)
	}
	enter("05")
	if s.maps[s.code].name != "original/05" || s.maps[s.code].code() != "RIGHT" {
		t.Fatalf("05 selected %s, code %q", s.maps[s.code].name, s.maps[s.code].code())
	}
	// The shipped level keeps its code; its original map is reached by number.
	enter("MOUSE")
	if s.maps[s.code].name != "level3_cp" {
		t.Fatalf("MOUSE selected %s", s.maps[s.code].name)
	}
	enter("02")
	if s.maps[s.code].name != "original/02" {
		t.Fatalf("02 selected %s", s.maps[s.code].name)
	}
	// A victory shows the next map's password.
	enter("05")
	s.finish(outcome{won: true})
	if s.maps[s.code].name != "original/06" || s.result.next != s.maps[3+6].password {
		t.Fatalf("after 05: %s, shown %q", s.maps[s.code].name, s.result.next)
	}
}
