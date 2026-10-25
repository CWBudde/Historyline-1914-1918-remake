package main

import (
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"unicode"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/campaign"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/saves"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scores"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// The menus around a battle, laid out like the original's: the main menu
// (start, map code, game values, control, extras, exit), the country selection
// before a battle, the outcome after it, and the question whether to leave a
// running battle. Every page is a frame over a painting on a 640×400 logical
// canvas; the shell holds what the pages show and does, the app draws it.

// page is one menu screen.
type page int

const (
	pageMain page = iota
	pageValues
	pageControl
	pageExtras
	pageCountry
	pageLoading // the battle is being prepared
	pageBattle  // the battle runs; the shell shows nothing
	pageOutcome
	pageLeave // the battle is paused: leave it?
	pageScores
	pageNews // the period's news, before a campaign battle
)

// mapLevel is a map of the menu: a setup of a monthly level
// (levels/monthly), opened by its code word.
type mapLevel struct {
	name       string // -level name: "monthly/1914-08:cp"
	file       string // the scenario file: "monthly/1914-08"
	mode       scenario.Mode
	lvl        *level.Level
	next       int    // the map a victory leads to, -1 for none
	password   string // the setup's code word
	players    int    // mission players (1 or 2); 0 means 1
	humanSides []int  // mission human sides
	winter     bool
	briefing   string // one line shown on the loading page
	month      string // "1914-08"
}

// news is the news page of the map's month.
func (m mapLevel) news() (campaign.Period, bool) {
	return campaign.News(m.month)
}

// briefingLine is the one-line briefing from the map's own facts; there is
// no briefing text in the map files (PLAN open question), so it is derived.
func briefingLine(code string, players int, humanSides []int, side state.Player, winter, knownSeason bool) string {
	season := ""
	if knownSeason {
		season = ", summer"
		if winter {
			season = ", winter"
		}
	}
	if players == 2 {
		return "Map " + code + season + " — two-player map; in v1 play one side vs computer or passive."
	}
	who := factionName(side)
	if len(humanSides) == 1 {
		who = factionName(dosbind.Factions[humanSides[0]])
	}
	return "Map " + code + season + " — command " + who + "; capture the enemy HQ."
}

func (m mapLevel) code() string {
	if m.password != "" {
		return m.password
	}
	return strings.ToUpper(m.lvl.MapName)
}

// campaignModes are the setups the menu campaign plays, in menu order: the
// Central Powers' chain, then the Entente's.
var campaignModes = []scenario.Mode{scenario.CP, scenario.EC}

// loadCampaign loads the campaign from the monthly levels (levels/monthly)
// in calendar order: one chain per side, every month's setup of that side
// with a code word, the Central Powers' first. A victory leads to the next
// month of the chain, none after its last.
func loadCampaign(dir string, set *data.Set) ([]mapLevel, error) {
	files, err := fs.Glob(levels.Monthly, "monthly/*.json")
	if err != nil {
		return nil, err
	}
	slices.Sort(files)
	scens := make([]*scenario.File, len(files))
	for i, f := range files {
		raw, err := readLevelFile(dir, levels.Monthly, f)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		if scens[i], err = scenario.Parse(raw); err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
	}
	var maps []mapLevel
	for _, mode := range campaignModes {
		first := len(maps)
		for i, f := range scens {
			s := f.Setup(mode)
			if s == nil || s.Code == "" {
				continue
			}
			file := strings.TrimSuffix(files[i], ".json")
			lvl, err := scenarioLevel(f, file, mode)
			if err != nil {
				return nil, err
			}
			m := mapLevel{name: file + ":" + string(mode), file: file, mode: mode, lvl: lvl, next: -1,
				password: s.Code, players: 1, humanSides: []int{dosbind.Side(mode.Human())}, winter: f.Winter(), month: f.Month}
			m.briefing = briefingLine(m.code(), m.players, m.humanSides, mode.Human(), m.winter, true)
			if len(maps) > first {
				maps[len(maps)-1].next = len(maps)
			}
			maps = append(maps, m)
		}
	}
	if len(maps) == 0 {
		return nil, fmt.Errorf("no campaign maps")
	}
	return maps, nil
}

// forSide is the menu map of the selected map's month for side: its setup
// for that side where the month has one, else the selected map.
func (s *shell) forSide(side state.Player) int {
	cur := s.maps[s.code]
	if cur.mode.Human() == side {
		return s.code
	}
	for i, m := range s.maps {
		if m.file == cur.file && m.mode.Human() == side {
			return i
		}
	}
	return s.code
}

// button is a control of a page, in logical canvas pixels.
type button struct {
	r     rect
	label string
	help  [2]string // the help box text while it has the focus
	off   bool      // disabled: shown, never activated
	left  bool      // label aligned left (menu rows), else centred
	do    func(s *shell)
}

// shellKey is a menu key, whatever the device.
type shellKey int

const (
	keyUp shellKey = iota
	keyDown
	keyLeft
	keyRight
	keyEnter
	keyBack
	keyErase
)

// scrollSpeeds are the map scroll speeds of the control menu.
var scrollSpeeds = []struct {
	name  string
	speed float64
}{{"NORMAL", scrollSpeed}, {"FAST", 2 * scrollSpeed}, {"SLOW", scrollSpeed / 2}}

// moveLimits are the move limits of the game values menu, 0 for none.
var moveLimits = []int{0, 5, 10, 15, 20}

// ruleModes are the rule sets of the game values menu, in cycle order:
// the original (DOS) rules are the default product mode; New (Standard)
// is the alternative remake engine.
var ruleModes = []struct {
	token string
	label string
}{
	{"dos", "RULES  ORIGINAL (DOS)"},
	{"standard", "RULES  NEW (STANDARD)"},
}

// outcome is how the last battle ended, from the human side.
type outcome struct {
	won, draw bool
	text      string // how it ended
	score     int
	next      string // the code of the next map after a victory, "" for none
}

type shell struct {
	page  page
	focus int // focused button of the page

	maps   []mapLevel
	code   int    // selected map
	typing bool   // the code row takes letters
	typed  string // the code typed so far
	note   [2]string
	side   state.Player // country chosen, or focused on the country page

	fights   bool // full walks and effects; off is reduced motion
	computer bool // player B is the computer; off passes every phase
	limit    int  // index into moveLimits
	scroll   int  // index into scrollSpeeds

	setFullscreen func(bool)  // nil in tests
	isFullscreen  func() bool // nil in tests

	result outcome
	quit   bool // leave the program

	rules       string
	scores      *scores.Store
	saves       *saves.Store
	scoreList   []scores.Entry
	pendingSave []byte // a loaded battle to resume, consumed by config
	pendingSide int
	saveBattle  func() error // set by the app; nil in tests
}

func newShell(maps []mapLevel, c matchConfig) *shell {
	rules := c.rules
	if rules == "" {
		rules = "dos"
	}
	return &shell{maps: maps, fights: !c.reduced, computer: c.opponent != "pass", rules: rules, scores: c.scores, saves: c.saves}
}

// config is the battle the menus describe: base with the selected map, side
// and game values.
func (s *shell) config(base matchConfig) matchConfig {
	c := base
	c.lvl, c.board = s.maps[s.code].lvl, nil
	c.player = s.side
	c.opponent = "ai"
	if !s.computer {
		c.opponent = "pass"
	}
	c.limit = moveLimits[s.limit]
	c.reduced = !s.fights
	c.scroll = scrollSpeeds[s.scroll].speed
	c.rules = s.rules
	if c.rules == "" {
		c.rules = "dos"
	}
	if s.pendingSave != nil {
		c.saved, c.player = s.pendingSave, dosbind.Factions[s.pendingSide]
		s.pendingSave = nil // consumed: a later START starts fresh
	}
	return c
}

// row is the n-th menu row of the left panel.
func row(n int) rect { return rect{40, 104 + n*44, 216, 30} }

// valueRow is the n-th row of the game values page, packed tighter so
// the rules row fits above the help box.
func valueRow(n int) rect { return rect{40, 96 + n*40, 216, 30} }

// ruleLabel is the game values label of the selected rule set.
func (s *shell) ruleLabel() string {
	for _, m := range ruleModes {
		if m.token == s.rules {
			return m.label
		}
	}
	return ruleModes[0].label
}

// cycleRules advances to the next rule set: dos -> standard -> dos.
func (s *shell) cycleRules() {
	for i, m := range ruleModes {
		if m.token == s.rules {
			s.rules = ruleModes[(i+1)%len(ruleModes)].token
			return
		}
	}
	s.rules = ruleModes[0].token
}

var (
	footerRect = rect{464, 346, 152, 30} // exit, back and continue
	cardRects  = [2]rect{{24, 88, 284, 238}, {332, 88, 284, 238}}
	stayRect   = rect{40, 266, 112, 30}
	saveRect   = rect{180, 266, 200, 30}
	leaveRect  = rect{408, 266, 160, 30}
)

// cardSides are the countries of the two cards: France on the left and
// Germany on the right, as in the original.
var cardSides = [2]state.Player{state.Entente, state.CentralPowers}

// buttons are the controls of the current page in focus order.
func (s *shell) buttons() []button {
	back := func(to page) button {
		return button{r: footerRect, label: "BACK", help: [2]string{"Back to the main menu."}, do: func(s *shell) { s.open(to) }}
	}
	switch s.page {
	case pageMain:
		code := "CODE  " + s.maps[s.code].code()
		if s.typing {
			code = "CODE  " + s.typed + "_"
		}
		return []button{
			{r: row(0), label: "START", left: true, help: [2]string{"Begin the battle for the map", "whose code is shown below."},
				do: func(s *shell) { s.side = s.levelSide(); s.open(pageCountry) }},
			{r: row(1), label: code, left: true, help: [2]string{"Type the code of a map you have", "reached, then press Enter."},
				do: func(s *shell) { s.typing, s.typed = true, "" }},
			{r: row(2), label: "GAME VALUES", left: true, help: [2]string{"Rules, battles, players and the", "move limit."},
				do: func(s *shell) { s.open(pageValues) }},
			{r: row(3), label: "CONTROL", left: true, help: [2]string{"Screen and scrolling."},
				do: func(s *shell) { s.open(pageControl) }},
			{r: row(4), label: "EXTRAS", left: true, help: [2]string{"Score list and saved games."},
				do: func(s *shell) { s.open(pageExtras) }},
			{r: footerRect, label: "EXIT", help: [2]string{"Leave Historyline 1914-1918."},
				do: func(s *shell) { s.quit = true }},
		}
	case pageValues:
		fights, b := "BATTLES  FULL", "PLAYER B  COMPUTER"
		if !s.fights {
			fights = "BATTLES  SHORT"
		}
		if !s.computer {
			b = "PLAYER B  PASSIVE"
		}
		limit := "MOVES  NO LIMIT"
		if n := moveLimits[s.limit]; n > 0 {
			limit = fmt.Sprintf("MOVES  %d A PHASE", n)
		}
		rulesHelp := [2]string{"Original (DOS) is the default mode.", "New: the alternative remake rules."}
		return []button{
			{r: valueRow(0), label: s.ruleLabel(), left: true, help: rulesHelp,
				do: func(s *shell) { s.cycleRules() }},
			{r: valueRow(1), label: fights, left: true, help: [2]string{"Full: units walk and fire in full.", "Short: moves and fire are instant."},
				do: func(s *shell) { s.fights = !s.fights }},
			{r: valueRow(2), label: "PLAYER A  HUMAN", left: true, off: true, help: [2]string{"You play the country you choose.", "Two human players are not in v1."}},
			{r: valueRow(3), label: b, left: true, help: [2]string{"Computer: the computer commands.", "Passive: it confirms without orders. (1-player v1)"},
				do: func(s *shell) { s.computer = !s.computer }},
			{r: valueRow(4), label: "DEPOTS  OWN ONLY", left: true, off: true, help: [2]string{"You see into your own buildings and", "transporters only."}},
			{r: valueRow(5), label: limit, left: true, help: [2]string{"Orders each side may give a phase.", "Fewer orders score higher."},
				do: func(s *shell) { s.limit = (s.limit + 1) % len(moveLimits) }},
			back(pageMain),
		}
	case pageControl:
		screen := "FULLSCREEN  OFF"
		if s.isFullscreen != nil && s.isFullscreen() {
			screen = "FULLSCREEN  ON"
		}
		return []button{
			{r: row(0), label: screen, left: true, help: [2]string{"Window or full screen (F11 in battle)."},
				do: func(s *shell) {
					if s.setFullscreen != nil && s.isFullscreen != nil {
						s.setFullscreen(!s.isFullscreen())
					}
				}},
			{r: row(1), label: "SCROLLING  " + scrollSpeeds[s.scroll].name, left: true, help: [2]string{"How fast the arrow keys move the map."},
				do: func(s *shell) { s.scroll = (s.scroll + 1) % len(scrollSpeeds) }},
			back(pageMain),
		}
	case pageExtras:
		loadOff, loadHelp := false, [2]string{"Continue a saved battle (one slot).", "Started from the leave menu: SAVE AND LEAVE."}
		switch {
		case s.rules != "dos":
			loadOff, loadHelp = true, [2]string{"Continue a saved battle.", "Saved battles need -rules dos."}
		case s.saves == nil || !s.saves.Has():
			loadOff, loadHelp = true, [2]string{"Continue a saved battle.", "No saved battle yet."}
		}
		return []button{
			{r: row(0), label: "SCORE LIST", left: true, help: [2]string{"The best scores of each map.", "Won battles are recorded automatically."},
				do: func(s *shell) {
					s.scoreList = nil
					if s.scores != nil {
						s.scoreList = s.scores.All()
					}
					s.open(pageScores)
				}},
			{r: row(1), label: "LOAD GAME", left: true, off: loadOff, help: loadHelp,
				do: func(s *shell) { s.loadSaved() }},
			back(pageMain),
		}
	case pageScores:
		return []button{{r: footerRect, label: "BACK", help: [2]string{"Back to extras."}, do: func(s *shell) { s.open(pageExtras) }}}
	case pageCountry:
		var bs []button
		for i, side := range cardSides {
			bs = append(bs, button{r: cardRects[i], label: countryName(side), help: [2]string{factionName(side)},
				do: func(s *shell) {
					s.side = side
					s.code = s.forSide(side)
					if _, ok := s.maps[s.code].news(); ok {
						s.open(pageNews)
					} else {
						s.open(pageLoading)
					}
				}})
		}
		return append(bs, back(pageMain))
	case pageNews:
		return []button{{r: rect{40, 344, 216, 30}, label: "TO BATTLE", left: true, help: [2]string{"On to the battle."}, do: func(s *shell) { s.open(pageLoading) }}}
	case pageOutcome:
		return []button{{r: footerRect, label: "CONTINUE", help: [2]string{"Back to the main menu."}, do: func(s *shell) { s.open(pageMain) }}}
	case pageLeave:
		saveOff := s.saveBattle == nil || s.rules != "dos" || s.saves == nil
		return []button{
			{r: stayRect, label: "STAY", help: [2]string{"Go on with the battle."}, do: func(s *shell) { s.open(pageBattle) }},
			{r: saveRect, label: "SAVE AND LEAVE", off: saveOff, help: [2]string{"Save this dos battle (one slot)", "and return to the main menu."}, do: func(s *shell) {
				if s.saveBattle != nil {
					if err := s.saveBattle(); err == nil {
						s.open(pageMain)
						s.note = [2]string{"Battle saved.", "LOAD GAME in EXTRAS continues it."}
					}
				}
			}},
			{r: leaveRect, label: "LEAVE", help: [2]string{"Give up the battle and return", "to the main menu."}, do: func(s *shell) { s.open(pageMain) }},
		}
	}
	return nil
}

// loadSaved stages the saved battle for the loading page.
func (s *shell) loadSaved() {
	if s.saves == nil {
		return
	}
	env, ok := s.saves.Get()
	if !ok {
		s.note = [2]string{"No saved battle yet."}
		return
	}
	for i, m := range s.maps {
		if m.name == env.Map {
			s.code, s.pendingSave, s.pendingSide = i, env.Data, env.Side
			s.side = dosbind.Factions[env.Side]
			s.open(pageLoading)
			return
		}
	}
	s.note = [2]string{"The saved map " + env.Map + " is not in the menu."}
}

// open shows page p with its first usable button focused (the card of the
// chosen side on the country page).
func (s *shell) open(p page) {
	s.page, s.typing, s.note = p, false, [2]string{}
	s.focus = s.step(-1, 1)
	if p == pageCountry && s.side == cardSides[1] {
		s.focus = 1
	}
}

// levelSide is the side the selected map is written for.
func (s *shell) levelSide() state.Player {
	p, err := state.PlayerByName(s.maps[s.code].lvl.Player)
	if err != nil {
		return state.CentralPowers
	}
	return p
}

// step returns the next usable button from i in direction dir, wrapping, or
// i when there is none.
func (s *shell) step(i, dir int) int {
	bs := s.buttons()
	if len(bs) == 0 {
		return -1
	}
	for range bs {
		i = ((i+dir)%len(bs) + len(bs)) % len(bs)
		if !bs[i].off {
			return i
		}
	}
	return max(i, 0)
}

// help is the help box text: a note, else the focused button's help.
func (s *shell) help() [2]string {
	if s.note != ([2]string{}) {
		return s.note
	}
	if bs := s.buttons(); s.focus >= 0 && s.focus < len(bs) {
		return bs[s.focus].help
	}
	return [2]string{}
}

// key handles a menu key.
func (s *shell) key(k shellKey) {
	if s.typing {
		switch k {
		case keyEnter:
			s.enterCode()
		case keyBack:
			s.typing = false
		case keyErase:
			if s.typed != "" {
				s.typed = s.typed[:len(s.typed)-1]
			}
		}
		return
	}
	switch k {
	case keyUp, keyLeft:
		s.focus, s.note = s.step(s.focus, -1), [2]string{}
	case keyDown, keyRight:
		s.focus, s.note = s.step(s.focus, 1), [2]string{}
	case keyEnter:
		s.activate(s.focus)
	case keyBack:
		switch s.page {
		case pageScores:
			s.open(pageExtras)
		case pageValues, pageControl, pageExtras, pageCountry, pageOutcome:
			s.open(pageMain)
		case pageLeave:
			s.open(pageBattle)
		}
	}
}

// typeRune adds a letter or digit to the code being typed.
func (s *shell) typeRune(r rune) {
	if s.typing && len(s.typed) < 12 && r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
		s.typed += string(unicode.ToUpper(r))
	}
}

// enterCode selects the map whose code was typed.
func (s *shell) enterCode() {
	s.typing = false
	for i, m := range s.maps {
		if m.code() == s.typed || strings.ToUpper(m.lvl.MapName) == s.typed {
			s.code = i
			s.note = [2]string{"Map " + m.lvl.MapName + " selected."}
			return
		}
	}
	s.note = [2]string{"Unknown code " + s.typed + ".", "The map stays " + s.maps[s.code].lvl.MapName + "."}
}

// activate does what button i does.
func (s *shell) activate(i int) {
	bs := s.buttons()
	if i < 0 || i >= len(bs) || bs[i].off || bs[i].do == nil {
		return
	}
	s.focus = i
	s.note = [2]string{}
	bs[i].do(s)
}

// hit is the button under logical point (x, y), -1 for none.
func (s *shell) hit(x, y int) int {
	for i, b := range s.buttons() {
		if x >= b.r.X && x < b.r.X+b.r.W && y >= b.r.Y && y < b.r.Y+b.r.H {
			return i
		}
	}
	return -1
}

// finish shows the outcome of a battle; a victory selects the next map.
func (s *shell) finish(o outcome) {
	if next := s.maps[s.code].next; o.won && next >= 0 {
		s.code = next
		o.next = s.maps[s.code].code()
	}
	s.result = o
	s.open(pageOutcome)
}

// countryName is the country the original lets a side choose.
func countryName(p state.Player) string {
	if p == state.Entente {
		return "FRANCE"
	}
	return "GERMANY"
}

// factionName names the alliance of a side.
func factionName(p state.Player) string {
	if p == state.Entente {
		return "The Entente"
	}
	return "The Central Powers"
}
