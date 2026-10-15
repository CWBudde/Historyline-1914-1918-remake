package main

import (
	"testing"

	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestEffectsNilSafe(t *testing.T) {
	var e *effects
	e.AddBattle(hexgrid.Coord{}, hexgrid.Coord{}, state.Battle{Damage: 3})
	e.Round(2, state.None)
	e.Victory(state.None)
	e.Tick()
	e.Dismiss()
	if e.Banner() != nil || e.Marks() != nil {
		t.Fatal("nil effects must be empty")
	}
}

func TestBattleMarks(t *testing.T) {
	var e effects
	from, to := hexgrid.Coord{Line: 1, Col: 1}, hexgrid.Coord{Line: 1, Col: 2}
	e.AddBattle(from, to, state.Battle{Damage: 4, DefenderDestroyed: true})
	m := e.Marks()
	if len(m) != 1 || m[0].At != to || m[0].Kind != markDeath {
		t.Fatalf("marks = %+v", m)
	}
	e.AddBattle(from, to, state.Battle{Damage: 2, CounterDamage: 1})
	if len(e.Marks()) != 3 {
		t.Fatalf("want defender+attacker hit added, got %+v", e.Marks())
	}
	e = effects{}
	e.AddBattle(from, to, state.Battle{})
	if len(e.Marks()) != 0 {
		t.Fatal("a miss with no damage must not mark")
	}
}

func TestEffectsBounded(t *testing.T) {
	var e effects
	e.AddBattle(hexgrid.Coord{}, hexgrid.Coord{Col: 1}, state.Battle{Damage: 1, AttackerDestroyed: true})
	e.Victory(state.Player(1))
	for range victoryFrames {
		e.Tick()
	}
	if len(e.Marks()) != 0 || e.Banner() != nil {
		t.Fatalf("effects outlived their duration: %+v %+v", e.Marks(), e.Banner())
	}
}

func TestBannerDismissAndVictoryReplacesRound(t *testing.T) {
	var e effects
	e.Round(3, state.Player(1))
	if b := e.Banner(); b == nil || b.Frames != roundFrames {
		t.Fatalf("round banner = %+v", b)
	}
	e.Victory(state.Player(1))
	if e.Banner().Frames != victoryFrames {
		t.Fatal("victory must replace the round banner")
	}
	e.Dismiss()
	if e.Banner() != nil {
		t.Fatal("click must dismiss the banner")
	}
}

func TestEndTurnShowsRoundBanner(t *testing.T) {
	c := newTestControl(t)
	c.EndTurn()
	if c.fx.Banner() == nil {
		t.Fatalf("no round banner after end turn: %q", c.Message)
	}
	c.Click(hexgrid.Coord{})
	if c.fx.Banner() != nil {
		t.Fatal("click must dismiss the round banner")
	}
}
