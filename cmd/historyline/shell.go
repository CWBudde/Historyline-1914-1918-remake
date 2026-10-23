package main

import (
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
	"unicode"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
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
)

// mapLevel is a map of the menu: a shipped level, or an imported original map
// with its exact board. Its code is the map name in capitals, an original
// map's number ("05"), or its password when the local original files give it.
type mapLevel struct {
	name     string // -level name: file name without .json, or original/NN
	lvl      *level.Level
	board    *dosbind.Board // an original map's binding; nil for a shipped level
	next     int            // the map a victory leads to, -1 for none
	password string         // an original map's password, "" without the local original files
}

func (m mapLevel) code() string {
	if m.password != "" {
		return m.password
	}
	return strings.ToUpper(m.lvl.MapName)
}

// loadCampaign loads the shipped maps in their order (level1_cp, level2_cp, …),
// then the imported original maps 00-71, from dir where present, else the
// embedded copies. A shipped map leads to the next; the last one, an original
// map transcribed (MapNo n is map n-1), leads where that map does. An
// original map leads to its mission's next map: through each campaign, none
// after its last map.
func loadCampaign(dir string, set *data.Set) ([]mapLevel, error) {
	names, err := fs.Glob(levels.FS, "level?_cp.json")
	if err != nil {
		return nil, err
	}
	slices.Sort(names)
	var maps []mapLevel
	for i, n := range names {
		name := strings.TrimSuffix(n, ".json")
		lvl, err := loadLevel(dir, name)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		maps = append(maps, mapLevel{name: name, lvl: lvl, next: i + 1})
	}
	if len(maps) == 0 {
		return nil, fmt.Errorf("no shipped maps")
	}
	shipped := len(maps)
	files, err := fs.Glob(levels.Original, "original/full/[0-9][0-9].json")
	if err != nil {
		return nil, err
	}
	slices.Sort(files)
	for _, f := range files {
		stem := strings.TrimSuffix(path.Base(f), ".json")
		l, lvl, b, err := loadOriginal(dir, stem, set)
		if err != nil {
			return nil, err
		}
		m := mapLevel{name: originalPrefix + stem, lvl: lvl, board: b, next: -1}
		if l.Mission != nil && l.Mission.Next != nil {
			m.next = shipped + *l.Mission.Next
		}
		maps = append(maps, m)
	}
	last := &maps[shipped-1]
	last.next = -1
	if k := shipped + last.lvl.MapNo - 1; last.lvl.MapNo >= 1 && k < len(maps) {
		last.next = maps[k].next
	}
	for _, m := range maps {
		if m.next >= len(maps) {
			return nil, fmt.Errorf("%s leads to map %d of %d", m.name, m.next, len(maps))
		}
	}
	return maps, nil
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
}

func newShell(maps []mapLevel, c matchConfig) *shell {
	return &shell{maps: maps, fights: !c.reduced, computer: c.opponent != "pass"}
}

// config is the battle the menus describe: base with the selected map, side
// and game values.
func (s *shell) config(base matchConfig) matchConfig {
	c := base
	c.lvl, c.board = s.maps[s.code].lvl, s.maps[s.code].board
	c.player = s.side
	c.opponent = "ai"
	if !s.computer {
		c.opponent = "pass"
	}
	c.limit = moveLimits[s.limit]
	c.reduced = !s.fights
	c.scroll = scrollSpeeds[s.scroll].speed
	return c
}

// row is the n-th menu row of the left panel.
func row(n int) rect { return rect{40, 104 + n*44, 216, 30} }

var (
	footerRect = rect{464, 346, 152, 30} // exit, back and continue
	cardRects  = [2]rect{{24, 88, 284, 238}, {332, 88, 284, 238}}
	stayRect   = rect{132, 266, 112, 30}
	leaveRect  = rect{348, 266, 160, 30}
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
			{r: row(2), label: "GAME VALUES", left: true, help: [2]string{"Battles, players and the move", "limit."},
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
		return []button{
			{r: row(0), label: fights, left: true, help: [2]string{"Full: units walk and fire in full.", "Short: moves and fire are instant."},
				do: func(s *shell) { s.fights = !s.fights }},
			{r: row(1), label: "PLAYER A  HUMAN", left: true, off: true, help: [2]string{"You play the country you choose.", "Two human players are not possible yet."}},
			{r: row(2), label: b, left: true, help: [2]string{"Computer: the computer commands.", "Passive: it confirms without orders."},
				do: func(s *shell) { s.computer = !s.computer }},
			{r: row(3), label: "DEPOTS  OWN ONLY", left: true, off: true, help: [2]string{"You see into your own buildings and", "transporters only."}},
			{r: row(4), label: limit, left: true, help: [2]string{"Orders each side may give a phase.", "Fewer orders score higher."},
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
		return []button{
			{r: row(0), label: "SCORE LIST", left: true, off: true, help: [2]string{"The best scores of a map.", "No scores are kept yet."}},
			{r: row(1), label: "LOAD GAME", left: true, off: true, help: [2]string{"Continue a saved battle.", "Saving is not possible yet."}},
			back(pageMain),
		}
	case pageCountry:
		var bs []button
		for i, side := range cardSides {
			bs = append(bs, button{r: cardRects[i], label: countryName(side), help: [2]string{factionName(side)},
				do: func(s *shell) { s.side = side; s.open(pageLoading) }})
		}
		return append(bs, back(pageMain))
	case pageOutcome:
		return []button{{r: footerRect, label: "CONTINUE", help: [2]string{"Back to the main menu."}, do: func(s *shell) { s.open(pageMain) }}}
	case pageLeave:
		return []button{
			{r: stayRect, label: "STAY", help: [2]string{"Go on with the battle."}, do: func(s *shell) { s.open(pageBattle) }},
			{r: leaveRect, label: "LEAVE", help: [2]string{"Give up the battle and return", "to the main menu."}, do: func(s *shell) { s.open(pageMain) }},
		}
	}
	return nil
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
