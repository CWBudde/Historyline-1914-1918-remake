package main

import (
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/ai"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// playAI steps the computer's turn to its end, closing battle panels and
// finishing walks as a waiting player would; it fails if the turn never ends.
func playAI(t *testing.T, c *control) {
	t.Helper()
	for range 10000 {
		if !c.AITurn() {
			return
		}
		for !c.walk.Done() {
			c.walk.Tick()
		}
		c.DismissReport()
		c.StepAI()
	}
	t.Fatal("the computer's turn did not end")
}

func TestAITurnReturnsControl(t *testing.T) {
	c := newTestControl(t)
	c.ai = &ai.Planner{}
	round := c.st.Round
	mine := ownMovers(c)[0]
	c.EndTurn()
	if !c.AITurn() || c.Message != "enemy turn" {
		t.Fatalf("after end turn: AI turn %v, message %q", c.AITurn(), c.Message)
	}
	// Clicks wait while the computer plays.
	c.Click(mine)
	if c.selected {
		t.Error("selected a unit during the computer's turn")
	}
	playAI(t, c)
	if c.st.Turn != c.player || c.st.Round != round+1 {
		t.Errorf("turn %v round %d, want %v round %d", c.st.Turn, c.st.Round, c.player, round+1)
	}
	if c.aiMoves == 0 || !strings.Contains(c.Message, "the enemy made") {
		t.Errorf("computer did nothing: %d moves, %q", c.aiMoves, c.Message)
	}
	if c.fx.Banner() == nil {
		t.Error("no round banner when the turn came back")
	}
}

func TestOpponentOpensWhenPlayerMovesSecond(t *testing.T) {
	for _, planner := range []*ai.Planner{nil, {}} {
		c := newTestControl(t)
		c.player = c.st.Human.Opponent()
		c.ai = planner
		if !c.AITurn() {
			t.Fatal("the level's side should open")
		}
		playAI(t, c)
		if c.st.Turn != c.player {
			t.Errorf("ai %v: turn %v, want the player's", planner != nil, c.st.Turn)
		}
	}
}

func TestPassOpponentKeepsPythonFlow(t *testing.T) {
	c := newTestControl(t)
	round := c.st.Round
	c.EndTurn()
	if c.AITurn() || c.st.Round != round+1 || !strings.Contains(c.Message, "passed") {
		t.Errorf("pass: AI turn %v, round %d, %q", c.AITurn(), c.st.Round, c.Message)
	}
}

func TestWalkHexFollowsDrawnPosition(t *testing.T) {
	path := []hexgrid.Coord{{Line: 0, Col: 0}, {Line: 1, Col: 0}, {Line: 2, Col: 0}}
	w := newWalk(path)
	var got []hexgrid.Coord
	for !w.Done() {
		got = append(got, w.Hex())
		w.Tick()
	}
	if got[0] != path[0] || got[stepFrames/2-1] != path[0] || got[stepFrames/2] != path[1] || got[len(got)-1] != path[2] {
		t.Errorf("walk hexes %v", got)
	}
	if w.Hex() != path[2] {
		t.Errorf("done walk at %v", w.Hex())
	}
}

func TestCount(t *testing.T) {
	if got := count(1, "attack") + ", " + count(0, "move") + ", " + count(3, "move"); got != "1 attack, 0 moves, 3 moves" {
		t.Error(got)
	}
}

func TestHiddenAttackerReport(t *testing.T) {
	r := &report{Attacker: side{Name: "unknown", Hidden: true}, Defender: side{Name: "CP_infantry", Before: 10, After: 7, Max: 10}}
	l := r.Lines()
	if l[0] != "ATK unknown\n" || !strings.Contains(l[1], "10 -> 7") {
		t.Errorf("report %q", l)
	}
}
