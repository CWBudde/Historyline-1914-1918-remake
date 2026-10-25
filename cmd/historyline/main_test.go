package main

import (
	"image/color"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// The browser build has no levels directory; the embedded copies must load.
func TestLoadLevelEmbeddedFallback(t *testing.T) {
	for _, name := range []string{"monthly/1914-08:cp", "monthly/1914-10:cp", "monthly/1914-12:cp"} {
		if _, err := loadLevel(t.TempDir(), name); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := loadLevel(t.TempDir(), "nope"); err == nil {
		t.Error("unknown level loaded")
	}
}

// Fog must not leak terrain colour or units of hexes the player cannot see.
func TestFogPolicy(t *testing.T) {
	c := color.RGBA{R: 200, G: 100, B: 50, A: 255}
	if got := fogShade(state.Visible, c); got != c {
		t.Errorf("visible shade = %v, want unchanged", got)
	}
	u1, u2 := fogShade(state.Unseen, c), fogShade(state.Unseen, color.RGBA{R: 10, G: 240, B: 90, A: 255})
	if u1 != u2 {
		t.Errorf("unseen shade depends on terrain: %v vs %v", u1, u2)
	}
	if s := fogShade(state.Seen, c); s.R >= c.R || s.R > s.B {
		t.Errorf("seen shade %v should be dimmed and grey", s)
	}
	for f, want := range map[state.Fog]bool{state.Unseen: false, state.Seen: false, state.Visible: true} {
		if showsUnit(f) != want {
			t.Errorf("showsUnit(%v) = %v, want %v", f, !want, want)
		}
	}
}

// pythonPulse is the Python remake's PULSE (internal/level's fixture), a
// German unit within reach of the enemy at the start, which advancing on
// the enemy wins.
const pythonPulse = "../../internal/level/testdata/pulse.json"

func newTestControl(t *testing.T) *control {
	t.Helper()
	lvl, err := loadLevel("", pythonPulse)
	if err != nil {
		t.Fatal(err)
	}
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	st, err := state.New(lvl, set)
	if err != nil {
		t.Fatal(err)
	}
	return &control{st: st, player: st.Human}
}

// ownMovers lists the player's visible map units that have somewhere to go.
func ownMovers(c *control) (out []hexgrid.Coord) {
	for l, line := range c.st.Hexes {
		for col := range line {
			at := hexgrid.Coord{Line: l, Col: col}
			if u := c.st.At(at).Unit; u != nil && u.Owner == c.player && len(c.st.Movable(at, u)) > 0 {
				out = append(out, at)
			}
		}
	}
	return out
}

func TestControlSelectMoveAndEndTurn(t *testing.T) {
	c := newTestControl(t)
	movers := ownMovers(c)
	if len(movers) == 0 {
		t.Fatal("no movable own unit in PULSE")
	}
	from := movers[0]
	u := c.st.At(from).Unit

	c.Click(from)
	if !c.selected || len(c.moves) == 0 {
		t.Fatalf("selecting %v gave selected=%v moves=%d", from, c.selected, len(c.moves))
	}
	var to hexgrid.Coord
	for at := range c.moves {
		to = at
		break
	}
	c.Click(to)
	if !u.Moved {
		t.Fatalf("click on %v did not move the unit: %q", to, c.Message)
	}

	for !c.walk.Done() {
		c.walk.Tick()
	}
	// An enemy unit is never selectable.
	for l, line := range c.st.Hexes {
		for col := range line {
			at := hexgrid.Coord{Line: l, Col: col}
			if e := c.st.At(at).Unit; e != nil && e.Owner != c.player {
				c.Click(at)
				if c.selected {
					t.Fatalf("selected enemy unit at %v", at)
				}
			}
		}
	}

	round := c.st.Round
	c.EndTurn()
	if c.st.Turn != c.player || c.st.Round != round+1 {
		t.Errorf("after end turn: turn %v round %d, want %v round %d", c.st.Turn, c.st.Round, c.player, round+1)
	}
	if u.Moved {
		t.Error("moved flag not reset for the new turn")
	}
}

func TestControlIgnoresClicksOutsideOwnTurn(t *testing.T) {
	c := newTestControl(t)
	c.player = c.st.Human.Opponent()
	for _, at := range ownMovers(&control{st: c.st, player: c.st.Human}) {
		c.Click(at)
		if c.selected {
			t.Fatal("selected a unit while it is not the player's turn")
		}
	}
}

func TestDevIndicators(t *testing.T) {
	g := &game{}
	if got := g.devIndicators(); got != "" {
		t.Errorf("no dev flags: got %q", got)
	}
	g.reveal, g.hexNumbers = true, true
	if got := g.devIndicators(); got != "DEV: REVEAL HEXNUMBERS" {
		t.Errorf("got %q", got)
	}
	// Unapproved unit art is labelled too, so screenshots are never mistaken for it.
	g.reveal, g.hexNumbers, g.art = false, false, &artPacks{candidate: true}
	if got := g.devIndicators(); got != "DEV: CANDIDATE-ART" {
		t.Errorf("got %q", got)
	}
	g.art.candidate = false
	if got := g.devIndicators(); got != "" {
		t.Errorf("approved art: got %q", got)
	}
}

func TestWeaponMenu(t *testing.T) {
	c := newTestControl(t)
	for l, line := range c.st.Hexes {
		for col := range line {
			at := hexgrid.Coord{Line: l, Col: col}
			u := c.st.At(at).Unit
			if u == nil || u.Owner != c.player {
				continue
			}
			c.Select(at)
			for tgt := range c.targets {
				c.OpenWeaponMenu(tgt)
				if !c.MenuOpen() {
					t.Fatal("menu empty for a legal target")
				}
				usable := -1
				for i, o := range c.menu {
					if o.Usable {
						usable = i
					}
				}
				if usable < 0 {
					t.Fatalf("legal target %v but no usable weapon: %+v", tgt, c.menu)
				}
				c.ChooseWeapon(usable)
				if c.MenuOpen() || !u.Attacked {
					t.Fatalf("weapon did not fire: %q", c.Message)
				}
				return
			}
		}
	}
	t.Skip("PULSE has no unit with a target at the start")
}

func TestWeaponMenuRejectsUnusable(t *testing.T) {
	c := &control{presentation: presentation{menu: []WeaponOption{{Slot: 0, Label: "gun: no ammo"}}}}
	c.ChooseWeapon(0)
	if !c.MenuOpen() || c.Message != "gun: no ammo" {
		t.Errorf("menu %v message %q", c.menu, c.Message)
	}
}

// ownBuilding finds a visible own building and stores an infantry unit in it.
func ownBuilding(t *testing.T, c *control) (hexgrid.Coord, *state.Unit) {
	t.Helper()
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	for l, line := range c.st.Hexes {
		for col := range line {
			at := hexgrid.Coord{Line: l, Col: col}
			h := c.st.At(at)
			if h.Container == nil || h.Owner != c.player || h.Unit != nil || c.st.Fog(c.player, at) != state.Visible {
				continue
			}
			for _, u := range h.Container.Units() {
				h.Container.Remove(u)
			}
			u, err := state.NewUnit(set, "CP_infantry")
			if err != nil {
				t.Skip("no infantryCP unit type:", err)
			}
			if err := h.Container.Add(u); err != nil {
				t.Fatal(err)
			}
			return at, u
		}
	}
	t.Skip("PULSE has no visible own building without a unit")
	return hexgrid.Coord{}, nil
}

func TestContentMenuLeave(t *testing.T) {
	c := newTestControl(t)
	at, u := ownBuilding(t, c)

	c.Click(at)
	if !c.MenuOpen() || len(c.menu) != 1 || !c.menu[0].Usable {
		t.Fatalf("content menu %+v, message %q", c.menu, c.Message)
	}
	c.ChooseMenu(0)
	if c.menuKind != menuActions || !c.selected || c.unit != u || len(c.moves) == 0 {
		t.Fatalf("stored unit not selected: selected=%v moves=%d %q", c.selected, len(c.moves), c.Message)
	}
	var to hexgrid.Coord
	for h := range c.moves {
		if c.st.At(h).Unit == nil && c.st.At(h).Container == nil {
			to = h
			break
		}
	}
	c.Click(to)
	if !u.Moved || c.st.At(to).Unit != u {
		t.Fatalf("unit did not leave to %v: %q", to, c.Message)
	}
	if slicesContains(c.st.At(at).Container.Units(), u) {
		t.Error("unit still stored after leaving")
	}
}

func TestContentMenuRejectsMovedUnit(t *testing.T) {
	c := newTestControl(t)
	at, u := ownBuilding(t, c)
	u.Moved = true
	c.OpenContents(at)
	if len(c.menu) != 1 || c.menu[0].Usable {
		t.Fatalf("moved unit offered as usable: %+v", c.menu)
	}
	c.ChooseMenu(0)
	if c.selected || !c.MenuOpen() {
		t.Error("unusable row changed state")
	}
}

func slicesContains(us []*state.Unit, u *state.Unit) bool {
	for _, x := range us {
		if x == u {
			return true
		}
	}
	return false
}

// A unit the player saw stays drawn (dimmed) where it was last seen, and a unit
// on a hex the player cannot see now is never drawn from live state.
func TestRememberedUnitShown(t *testing.T) {
	c := newTestControl(t)
	lvl, err := loadLevel("", "monthly/1914-08:cp")
	if err != nil {
		t.Fatal(err)
	}
	st := c.st
	g := newGame(lvl, st, c.player, false)
	var enemy, spot hexgrid.Coord
	found := false
	for l, line := range st.Hexes {
		for col := range line {
			at := hexgrid.Coord{Line: l, Col: col}
			u := st.At(at).Unit
			if u == nil || u.Owner == c.player || st.Fog(c.player, at) != state.Unseen {
				continue
			}
			for _, n := range st.Grid.Neighbors(at) {
				if h := st.At(n); h.Unit == nil && h.Container == nil && st.Fog(c.player, n) == state.Unseen {
					enemy, spot, found = at, n, true
				}
			}
			if found {
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Fatal("no hidden enemy with a free hidden neighbour in PULSE")
	}
	if shown, _ := g.unitShown(enemy); shown {
		t.Fatal("hidden enemy drawn")
	}
	scout := st.At(ownMovers(c)[0]).Unit
	st.At(spot).Unit = scout
	st.UpdateVision(c.player)
	if shown, rem := g.unitShown(enemy); !shown || rem {
		t.Fatalf("visible enemy: shown %v remembered %v", shown, rem)
	}
	st.At(spot).Unit = nil
	st.UpdateVision(c.player)
	if st.Fog(c.player, enemy) != state.Seen {
		t.Skipf("enemy hex is %v after the scout left; another own unit sees it", st.Fog(c.player, enemy))
	}
	gone := st.At(enemy).Unit
	st.At(enemy).Unit = nil
	if shown, rem := g.unitShown(enemy); !shown || !rem {
		t.Errorf("left enemy: shown %v remembered %v, want last-seen snapshot", shown, rem)
	}
	st.At(enemy).Unit = gone
	st.At(spot).Unit = gone // an enemy walks onto a hex that is only remembered
	if shown, _ := g.unitShown(spot); shown {
		t.Error("enemy on a remembered-empty hex drawn from live state")
	}
	st.At(spot).Unit = nil
	g.reveal = true
	if shown, rem := g.unitShown(enemy); !shown || rem {
		t.Errorf("reveal: shown %v remembered %v", shown, rem)
	}
}

func TestOwnerShown(t *testing.T) {
	c := newTestControl(t)
	lvl, err := loadLevel("", "monthly/1914-08:cp")
	if err != nil {
		t.Fatal(err)
	}
	g := newGame(lvl, c.st, c.player, false)
	at, _ := ownBuilding(t, c)
	if got := g.ownerShown(at); got != c.player {
		t.Errorf("own visible building shows owner %v", got)
	}
	for l, line := range c.st.Hexes {
		for col := range line {
			p := hexgrid.Coord{Line: l, Col: col}
			if h := c.st.At(p); h.Container != nil && c.st.Fog(c.player, p) == state.Unseen {
				if got := g.ownerShown(p); got != state.None {
					t.Errorf("unseen building %v shows owner %v", p, got)
				}
			}
		}
	}
}
