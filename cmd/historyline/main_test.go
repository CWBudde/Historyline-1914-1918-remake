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
	for _, name := range []string{"level1_cp", "level2_cp", "level3_cp"} {
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

func newTestControl(t *testing.T) *control {
	t.Helper()
	lvl, err := loadLevel("", "level1_cp")
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
		t.Fatal("no movable own unit in level1_cp")
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
}
