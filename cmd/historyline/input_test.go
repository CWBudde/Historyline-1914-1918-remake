package main

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestKeyBindingsMatchPython(t *testing.T) {
	want := map[ebiten.Key]action{
		ebiten.KeyF1: actEndTurn, ebiten.KeyEscape: actCancel, ebiten.KeySpace: actSelectHex,
		ebiten.KeyPageUp: actViewTop, ebiten.KeyPageDown: actViewBottom, ebiten.KeyQ: actQuit,
		ebiten.KeyF11: actFullscreen, ebiten.KeyF9: actVignettes,
	}
	for k, a := range want {
		got := actionsFor(func(p ebiten.Key) bool { return p == k })
		if len(got) != 1 || got[0] != a {
			t.Errorf("key %v -> %v, want %v", k, got, a)
		}
	}
	if got := actionsFor(func(ebiten.Key) bool { return false }); len(got) != 0 {
		t.Errorf("no key pressed gave %v", got)
	}
	seen := map[ebiten.Key]bool{}
	for _, b := range keyBindings {
		if seen[b.Key] {
			t.Errorf("key %v bound twice", b.Key)
		}
		seen[b.Key] = true
	}
}

func TestEndTurnClearsSelectionAndReports(t *testing.T) {
	c := newTestControl(t)
	at := ownMovers(c)[0]
	c.Click(at)
	if !c.selected {
		t.Fatal("setup: nothing selected")
	}
	c.Message = ""
	c.EndTurn()
	if c.selected || c.unit != nil || c.moves != nil || c.targets != nil || c.MenuOpen() {
		t.Error("selection not cleared by end turn")
	}
	if c.Message == "" {
		t.Error("no message after end turn")
	}
}
