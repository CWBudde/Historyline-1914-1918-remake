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
	// The Central Powers' chain through the months, then the Entente's:
	// the original maps on every second month from August 1914, Marne in
	// September 1914 on both.
	if len(codes) != 50 || codes[0] != "PULSE" || codes[1] != "KLUCK" || codes[2] != "CIVIL" || codes[24] != "STERN" ||
		codes[25] != "BATLE" || codes[26] != "MARNE" || codes[27] != "GOOSE" || codes[49] != "SODOM" {
		t.Fatalf("codes %v", codes)
	}
	if s.maps[0].name != "monthly/1914-08:cp" || s.maps[25].name != "monthly/1914-08:ec" {
		t.Fatalf("names %s, %s", s.maps[0].name, s.maps[25].name)
	}
	for i, m := range s.maps {
		want := state.CentralPowers
		if i >= 25 {
			want = state.Entente
		}
		if s.code = i; s.levelSide() != want || m.mode.Human() != want {
			t.Errorf("%s: side %v", m.name, s.levelSide())
		}
	}
}

func TestCampaignBriefingsAndPlayers(t *testing.T) {
	s := testShell(t)
	for i, m := range s.maps {
		if m.briefing == "" || m.players != 1 || len(m.humanSides) != 1 {
			t.Errorf("map %d %s: briefing %q, players %d %v", i, m.name, m.briefing, m.players, m.humanSides)
		}
	}
	if s.maps[24].next != -1 || s.maps[49].next != -1 {
		t.Error("campaign ends should have no next")
	}
	if !s.maps[2].winter || s.maps[0].winter || !strings.Contains(s.maps[2].briefing, "winter") {
		t.Errorf("CIVIL winter %v, PULSE winter %v, %q", s.maps[2].winter, s.maps[0].winter, s.maps[2].briefing)
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
	if s.typing || s.code != 3 {
		t.Fatalf("typing %v, code %d", s.typing, s.code)
	}
	s.activate(1)
	for _, r := range "nowhere" {
		s.typeRune(r)
	}
	s.key(keyErase)
	s.key(keyEnter)
	if s.code != 3 || s.help()[0] != "Unknown code NOWHER." {
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
	if s.page != pageNews {
		t.Fatalf("page %d, want the period's news", s.page)
	}
	s.activate(0) // TO BATTLE
	if s.page != pageLoading {
		t.Fatalf("page %d", s.page)
	}
	// France takes the month's Entente setup.
	c := s.config(matchConfig{rules: "dos"})
	if c.lvl != s.maps[25].lvl || c.board != nil || c.player != state.Entente || c.opponent != "pass" || c.limit != 5 ||
		!c.reduced || c.scroll != 2*scrollSpeed || c.rules != "dos" {
		t.Fatalf("config %+v", c)
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
	if s.page != pageOutcome || s.code != 1 || s.result.next != "KLUCK" {
		t.Fatalf("page %d, code %d, next %q", s.page, s.code, s.result.next)
	}
	s.finish(outcome{})
	if s.code != 1 || s.result.next != "" {
		t.Fatal("a defeat changed the map")
	}
	// Each side's chain runs through the months to its last.
	for _, c := range []struct {
		from int
		next string
	}{{1, "monthly/1914-10:cp"}, {23, "monthly/1918-06:cp"}, {24, ""},
		{25, "monthly/1914-09:ec"}, {26, "monthly/1914-10:ec"}, {49, ""}} {
		s.code = c.from
		s.finish(outcome{won: true})
		want, shown := s.maps[c.from].name, ""
		if c.next != "" {
			want, shown = c.next, s.maps[s.code].password
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
	// Every setup has its own five-letter code word.
	seen := map[string]bool{}
	for i, m := range s.maps {
		if len(m.password) != 5 || strings.ToUpper(m.password) != m.password || seen[m.password] {
			t.Errorf("map %02d: password %q", i, m.password)
		}
		seen[m.password] = true
	}
	enter := func(code string) {
		s.activate(1)
		for _, r := range code {
			s.typeRune(r)
		}
		s.key(keyEnter)
	}
	enter(strings.ToLower(s.maps[30].password))
	if s.code != 30 || s.buttons()[1].label != "CODE  "+s.maps[30].password {
		t.Fatalf("map 30's password selected %s, %q", s.maps[s.code].name, s.buttons()[1].label)
	}
	enter("GOOSE")
	if s.maps[s.code].name != "monthly/1914-10:ec" {
		t.Fatalf("GOOSE selected %s", s.maps[s.code].name)
	}
	enter("MOUSE")
	if s.maps[s.code].name != "monthly/1914-12:cp" {
		t.Fatalf("MOUSE selected %s", s.maps[s.code].name)
	}
	// A victory shows the next map's password.
	s.finish(outcome{won: true})
	if s.maps[s.code].name != "monthly/1915-02:cp" || s.result.next != "VENOM" {
		t.Fatalf("after MOUSE: %s, shown %q", s.maps[s.code].name, s.result.next)
	}
}

func TestCampaignPeriodsAndNews(t *testing.T) {
	s := testShell(t)
	// The original maps show their period's news, Marne its own month's.
	for idx, want := range map[int]string{0: "August – September 1914", 1: "September 1914", 2: "October – November 1914",
		24: "June – July 1918", 25: "August – September 1914", 26: "September 1914"} {
		if p, ok := s.maps[idx].news(); !ok || p.Label != want {
			t.Errorf("maps[%d] (%s) news %q, want %q", idx, s.maps[idx].month, p.Label, want)
		}
	}
	// Starting a campaign map shows its news first; TO BATTLE then loads.
	s.code = 0
	s.open(pageCountry)
	s.activate(1) // Germany
	if s.page != pageNews || s.code != 0 {
		t.Fatalf("campaign map start page %d code %d, want news of map 0", s.page, s.code)
	}
	s.activate(0)
	if s.page != pageLoading {
		t.Fatalf("news continue page %d, want loading", s.page)
	}
}

// The country card picks the month's setup for its side.
func TestCountryPicksTheSidesSetup(t *testing.T) {
	s := testShell(t)
	s.code = 2 // CIVIL, the Central Powers' October 1914
	s.open(pageCountry)
	s.activate(0) // France
	if s.maps[s.code].name != "monthly/1914-10:ec" || s.side != state.Entente {
		t.Fatalf("France on CIVIL: %s, side %v", s.maps[s.code].name, s.side)
	}
	c := s.config(matchConfig{})
	if c.lvl != s.maps[s.code].lvl || c.player != state.Entente {
		t.Fatalf("config plays %s as %v", c.lvl.MapName, c.player)
	}
	// Under the original rules both sides earn the setup's credits.
	sess, err := dosplay.New(c.lvl, testSet(t), dosplay.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if in := sess.Stats().Income; in[0] <= 0 || in[1] <= 0 {
		t.Errorf("GOOSE income %v", in)
	}
	s.open(pageCountry)
	s.activate(1) // Germany
	if s.maps[s.code].name != "monthly/1914-10:cp" {
		t.Fatalf("Germany on GOOSE: %s, want CIVIL", s.maps[s.code].name)
	}
}

func TestMonthlyMarne(t *testing.T) {
	s := testShell(t)
	m := s.maps[26]
	if m.code() != "MARNE" || m.name != "monthly/1914-09:ec" || m.month != "1914-09" || m.next != 27 {
		t.Fatalf("monthly map %+v", m)
	}
	if k := s.maps[1]; k.code() != "KLUCK" || k.file != m.file || k.mode.Human() != state.CentralPowers {
		t.Fatalf("Marne's German setup %+v", k)
	}
	// The code word selects it.
	s.typing, s.typed = true, "MARNE"
	s.key(keyEnter)
	if s.code != 26 {
		t.Fatalf("code MARNE selected map %d", s.code)
	}
	if s.levelSide() != state.Entente || !strings.Contains(m.briefing, "The Entente") {
		t.Fatalf("side %v, briefing %q", s.levelSide(), m.briefing)
	}
	// Its own September 1914 news comes before the battle.
	if p, ok := m.news(); !ok || p.Label != "September 1914" {
		t.Fatalf("news %+v ok %v", p, ok)
	}
	s.open(pageCountry)
	s.activate(0)
	if s.page != pageNews {
		t.Fatalf("monthly start page %d, want news", s.page)
	}
	s.activate(0)
	if s.page != pageLoading {
		t.Fatalf("news continue page %d, want loading", s.page)
	}
	// Easy by design: the Entente field force outnumbers the German one.
	var ec, cp int
	for _, line := range m.lvl.Lines {
		for _, h := range line {
			switch {
			case strings.HasPrefix(h.Unit, "EC_"):
				ec++
			case strings.HasPrefix(h.Unit, "CP_"):
				cp++
			}
		}
	}
	if ec <= cp {
		t.Fatalf("field units: Entente %d, Central Powers %d", ec, cp)
	}
	// Both engines start a battle on it.
	set := testSet(t)
	if _, err := dosplay.New(m.lvl, set, dosplay.Options{}); err != nil {
		t.Fatalf("dos: %v", err)
	}
	if _, err := state.New(m.lvl, set); err != nil {
		t.Fatalf("standard: %v", err)
	}
}
