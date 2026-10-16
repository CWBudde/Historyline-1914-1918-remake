package main

import (
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

func at(l, c int) hexgrid.Coord { return hexgrid.Coord{Line: l, Col: c} }

func onGrass(id string) level.Hex { return level.Hex{Terrain: "grass", Unit: id} }

func building(terrain string, owner state.Player, units ...string) level.Hex {
	return level.Hex{Terrain: terrain, Content: &level.Content{Owner: int(owner), Units: units}}
}

// boardLevel is a 6-wide, 12-line grass map for the Central Powers with the
// given hexes placed on it.
func boardLevel(place map[hexgrid.Coord]level.Hex) *level.Level {
	lv := &level.Level{MapName: "test", MapNo: 1, Player: "Central Powers"}
	for l := range 12 {
		line := make([]level.Hex, 6-l%2)
		for c := range line {
			line[c] = level.Hex{Terrain: "grass"}
			if h, ok := place[at(l, c)]; ok {
				line[c] = h
			}
		}
		lv.Lines = append(lv.Lines, line)
	}
	return lv
}

// boardControl is a control for the Central Powers on boardLevel(place).
func boardControl(t *testing.T, r state.Rules, place map[hexgrid.Coord]level.Hex) *control {
	t.Helper()
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	st, err := state.NewWithRules(boardLevel(place), set, r)
	if err != nil {
		t.Fatal(err)
	}
	return &control{st: st, player: state.CentralPowers}
}

// rows renders menu rows as "label" or "label!" for disabled rows.
func rows(m []WeaponOption) string {
	var out []string
	for _, o := range m {
		s := o.Label
		if !o.Usable {
			s += "!"
		}
		out = append(out, s)
	}
	return strings.Join(out, " | ")
}

func TestActionMenuRows(t *testing.T) {
	cases := []struct {
		name  string
		rules state.Rules
		place map[hexgrid.Coord]level.Hex
		sel   hexgrid.Coord
		setup func(u *state.Unit)
		want  string
	}{
		{"fresh unit with a target", state.Standard,
			map[hexgrid.Coord]level.Hex{at(4, 2): onGrass("CP_infantry"), at(6, 2): onGrass("EC_infantry")},
			at(4, 2), nil, "Move | Attack | Exit"},
		{"no target in range", state.Standard,
			map[hexgrid.Coord]level.Hex{at(4, 2): onGrass("CP_infantry")},
			at(4, 2), nil, "Move | Attack: no target in range! | Exit"},
		{"moved under Standard", state.Standard,
			map[hexgrid.Coord]level.Hex{at(4, 2): onGrass("CP_infantry"), at(6, 2): onGrass("EC_infantry")},
			at(4, 2), func(u *state.Unit) { u.Moved = true }, "Move: already moved! | Attack: already moved! | Exit"},
		{"moved under Classic may still attack", state.Classic,
			map[hexgrid.Coord]level.Hex{at(4, 2): onGrass("CP_infantry"), at(6, 2): onGrass("EC_infantry")},
			at(4, 2), func(u *state.Unit) { u.Moved = true }, "Move: already moved! | Attack | Exit"},
		{"attacked under Standard", state.Standard,
			map[hexgrid.Coord]level.Hex{at(4, 2): onGrass("CP_infantry"), at(6, 2): onGrass("EC_infantry")},
			at(4, 2), func(u *state.Unit) { u.Attacked = true }, "Move: already attacked! | Attack: already attacked! | Exit"},
		{"static unit", state.Standard,
			map[hexgrid.Coord]level.Hex{at(4, 2): onGrass("CP_bunker")},
			at(4, 2), nil, "Move: cannot move! | Attack: no target in range! | Contents: empty! | Exit"},
		{"empty carrier", state.Standard,
			map[hexgrid.Coord]level.Hex{at(4, 2): onGrass("CP_supplyCar")},
			at(4, 2), nil, "Move | Contents: empty! | Exit"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := boardControl(t, tc.rules, tc.place)
			if tc.setup != nil {
				tc.setup(c.st.At(tc.sel).Unit)
			}
			c.Click(tc.sel)
			if !c.selected || c.menuKind != menuActions {
				t.Fatalf("click did not open the action menu: selected %v, kind %v", c.selected, c.menuKind)
			}
			if got := rows(c.menu); got != tc.want {
				t.Errorf("rows %q, want %q", got, tc.want)
			}
		})
	}
}

func TestActionMenuNarrowsOverlays(t *testing.T) {
	place := map[hexgrid.Coord]level.Hex{at(4, 2): onGrass("CP_infantry"), at(6, 2): onGrass("EC_infantry")}
	c := boardControl(t, state.Standard, place)
	c.Click(at(4, 2))
	if len(c.moves) == 0 || len(c.targets) == 0 {
		t.Fatalf("selection lacks overlays: %d moves, %d targets", len(c.moves), len(c.targets))
	}
	c.ChooseMenu(0) // Move
	if c.MenuOpen() || len(c.moves) == 0 || len(c.targets) != 0 || !c.selected {
		t.Errorf("Move: menu %v, %d moves, %d targets", c.MenuOpen(), len(c.moves), len(c.targets))
	}

	c.Click(at(6, 2)) // not a move target any more: an enemy, so the selection goes
	c.Click(at(4, 2))
	c.ChooseMenu(1) // Attack
	if c.MenuOpen() || len(c.moves) != 0 || !c.targets[at(6, 2)] {
		t.Errorf("Attack: menu %v, %d moves, targets %v", c.MenuOpen(), len(c.moves), c.targets)
	}
	c.Click(at(6, 2))
	if c.menuKind != menuWeapons || !c.MenuOpen() {
		t.Errorf("a red hex should open the weapon menu, kind %v", c.menuKind)
	}
}

func TestActionMenuDirectClickStillMoves(t *testing.T) {
	c := boardControl(t, state.Standard, map[hexgrid.Coord]level.Hex{at(4, 2): onGrass("CP_infantry")})
	c.Click(at(4, 2))
	c.Click(at(6, 2))
	if c.st.At(at(6, 2)).Unit == nil {
		t.Errorf("a blue hex click with the action menu open did not move: %q", c.Message)
	}
}

func TestActionMenuExitDeselects(t *testing.T) {
	c := boardControl(t, state.Standard, map[hexgrid.Coord]level.Hex{at(4, 2): onGrass("CP_infantry")})
	c.Click(at(4, 2))
	c.ChooseMenu(len(c.menu) - 1)
	if c.selected || c.MenuOpen() || c.moves != nil {
		t.Errorf("Exit left selected %v, menu %v", c.selected, c.MenuOpen())
	}
}

func TestActionMenuContentsAndLeave(t *testing.T) {
	c := boardControl(t, state.Standard, map[hexgrid.Coord]level.Hex{
		at(4, 2): onGrass("CP_infantry"), at(6, 2): onGrass("CP_supplyCar"),
	})
	c.Click(at(4, 2))
	c.Click(at(6, 2)) // enter the car
	if got := c.st.At(at(6, 2)).Unit.Content.Units(); len(got) != 1 {
		t.Fatalf("car holds %d units: %q", len(got), c.Message)
	}
	c.EndTurn() // the opponent passes; the infantry may move again
	c.Click(at(6, 2))
	if got := rows(c.menu); got != "Move | Contents | Exit" {
		t.Fatalf("carrier rows %q", got)
	}
	c.ChooseMenu(1) // Contents
	if c.menuKind != menuContents || len(c.menu) != 1 {
		t.Fatalf("contents menu: kind %v, %q", c.menuKind, rows(c.menu))
	}
	c.ChooseMenu(0) // the infantry
	if got := rows(c.menu); c.menuKind != menuActions || got != "Move | Exit" {
		t.Fatalf("a stored unit gets Move/Exit, got kind %v %q", c.menuKind, got)
	}
	c.ChooseMenu(0) // Move
	var to hexgrid.Coord
	for to = range c.moves {
		break
	}
	c.Click(to)
	if c.st.At(to).Unit == nil || len(c.st.At(at(6, 2)).Unit.Content.Units()) != 0 {
		t.Errorf("the infantry did not leave: %q", c.Message)
	}
}

func TestActionMenuRect(t *testing.T) {
	const w, h, mw = 400, 300, 100
	n := 4
	mh := n * menuRowHeight
	cases := []struct {
		name         string
		cx, cy, half int
		wantX, wantY int
	}{
		{"right of the unit", 100, 150, 20, 100 + 20 + actionMenuGap, 150 - mh/2},
		{"flips left at the right edge", 350, 150, 20, 350 - 20 - actionMenuGap - mw, 150 - mh/2},
		{"clamped at the top", 100, 5, 20, 100 + 20 + actionMenuGap, hudMargin},
		{"clamped at the bottom", 100, 299, 20, 100 + 20 + actionMenuGap, h - hudMargin - mh},
		{"top right corner", 390, 0, 20, 390 - 20 - actionMenuGap - mw, hudMargin},
		{"bottom left corner", 0, 300, 20, 20 + actionMenuGap, h - hudMargin - mh},
		{"wide hex at both edges stays on screen", 200, 150, 190, w - hudMargin - mw, 150 - mh/2},
	}
	for _, tc := range cases {
		r := actionMenuRect(tc.cx, tc.cy, tc.half, mw, n, w, h)
		if r.X != tc.wantX || r.Y != tc.wantY || r.W != mw || r.H != mh {
			t.Errorf("%s: %+v, want x %d y %d", tc.name, r, tc.wantX, tc.wantY)
		}
		if r.X < 0 || r.Y < 0 || r.X+r.W > w || r.Y+r.H > h {
			t.Errorf("%s: %+v leaves the %dx%d screen", tc.name, r, w, h)
		}
	}
	// A menu wider or taller than the screen starts at the margin.
	if r := actionMenuRect(10, 10, 5, 1000, 50, w, h); r.X != hudMargin || r.Y != hudMargin {
		t.Errorf("oversized menu at %+v", r)
	}
}

func TestMenuPressFiresOnReleaseOverSameRow(t *testing.T) {
	p := newMenuPress()
	if row, used := p.update(true, false, true, 2); row != menuNone || !used || p.Row() != 2 {
		t.Fatalf("press: row %d used %v pressed %d", row, used, p.Row())
	}
	if row, used := p.update(false, false, true, 2); row != menuNone || !used {
		t.Errorf("hold: row %d used %v", row, used)
	}
	if row, used := p.update(false, true, false, 2); row != 2 || !used || p.Row() != menuNone {
		t.Errorf("release: row %d used %v pressed %d", row, used, p.Row())
	}
}

func TestMenuPressCancelledByDraggingOff(t *testing.T) {
	p := newMenuPress()
	p.update(true, false, true, 1)
	if row, used := p.update(false, true, false, menuNone); row != menuNone || !used || p.Row() != menuNone {
		t.Errorf("release off the row: row %d used %v", row, used)
	}
	if row, used := p.update(true, false, true, menuNone); row != menuNone || used {
		t.Errorf("a press outside the menu must fall through: row %d used %v", row, used)
	}
	p.update(true, false, true, 0)
	if _, used := p.update(false, false, false, 0); used || p.Row() != menuNone {
		t.Error("a lost release must clear the press")
	}
}
