package main

import (
	"strings"
	"testing"

	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// The tests below drive whole commands through control clicks, as the player
// does, and check the rules engine's state afterwards.

func TestEnterOwnDepotAndLeave(t *testing.T) {
	c := boardControl(t, state.Standard, map[hexgrid.Coord]level.Hex{
		at(4, 2): onGrass("CP_infantry"), at(6, 2): building("cmpN", state.CentralPowers),
	})
	inf := c.st.At(at(4, 2)).Unit
	c.Click(at(4, 2))
	if c.moves[at(6, 2)].Path == nil {
		t.Fatal("own depot not in the move overlay")
	}
	c.Click(at(6, 2))
	depot := c.st.At(at(6, 2))
	if depot.Unit != nil || len(depot.Container.Units()) != 1 || depot.Container.Units()[0] != inf {
		t.Fatalf("infantry not stored: %q", c.Message)
	}
	c.EndTurn()
	c.Click(at(6, 2)) // no map unit: the depot's contents open
	if c.menuKind != menuContents || !c.menu[0].Usable {
		t.Fatalf("depot contents: kind %v %q", c.menuKind, rows(c.menu))
	}
	c.ChooseMenu(0)
	c.Click(at(8, 2))
	if c.st.At(at(8, 2)).Unit != inf || len(depot.Container.Units()) != 0 {
		t.Errorf("infantry did not leave the depot: %q", c.Message)
	}
}

func TestCaptureDepotThroughControl(t *testing.T) {
	for _, owner := range []state.Player{state.Entente, state.None} {
		c := boardControl(t, state.Standard, map[hexgrid.Coord]level.Hex{
			at(4, 2): onGrass("CP_cavalry"), at(8, 2): building("factoryE_w", owner),
		})
		c.Click(at(4, 2))
		c.Click(at(8, 2))
		if got := c.st.At(at(8, 2)).Owner; got != state.CentralPowers {
			t.Errorf("owner %v: depot owner %v after entering (%q)", owner, got, c.Message)
		}
		if !strings.Contains(c.Message, "capturing") || c.fx.Banner() == nil {
			t.Errorf("owner %v: no capture message or banner: %q", owner, c.Message)
		}
		if c.st.Winner != state.None {
			t.Errorf("owner %v: a depot won the game", owner)
		}
	}
}

func TestCaptureHQWinsThroughControl(t *testing.T) {
	c := boardControl(t, state.Standard, map[hexgrid.Coord]level.Hex{
		at(4, 2): onGrass("CP_infantry"), at(6, 2): building("hqN", state.Entente),
	})
	c.Click(at(4, 2))
	c.Click(at(6, 2))
	if c.st.Winner != state.CentralPowers || !strings.Contains(c.Message, "wins") {
		t.Errorf("winner %v, %q", c.st.Winner, c.Message)
	}
	c.SkipWalk()
	c.Click(at(4, 2))
	if !strings.Contains(c.Message, "has won") {
		t.Errorf("a click after the victory: %q", c.Message)
	}
}

func TestRejectedMoveShowsEngineError(t *testing.T) {
	c := boardControl(t, state.Standard, map[hexgrid.Coord]level.Hex{at(4, 2): onGrass("CP_infantry")})
	inf := c.st.At(at(4, 2)).Unit
	c.Click(at(4, 2))
	if c.moves[at(6, 2)].Path == nil {
		t.Fatal("no move overlay")
	}
	// The overlay is stale: the engine no longer allows the move.
	inf.Moved = true
	c.Click(at(6, 2))
	if c.st.At(at(4, 2)).Unit != inf || c.st.At(at(6, 2)).Unit != nil {
		t.Fatal("a refused move changed the map")
	}
	if !strings.Contains(c.Message, "already moved") {
		t.Errorf("the engine's refusal is not shown: %q", c.Message)
	}
}
