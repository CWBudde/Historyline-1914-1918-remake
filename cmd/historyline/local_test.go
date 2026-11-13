package main

import (
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/faction"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/gesture"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/hajimehoshi/ebiten/v2"
)

// The tests below play two humans on one screen as they do: each through its
// own seat, the left one on key set A, the right one on key set B, and the
// hot-seat through its curtain.

// newLocalTest is the split game of August 1914's two-player setup for two
// humans in mode m, the left panel's human served.
func newLocalTest(t *testing.T, m localMode) (*game, *dosControl) {
	t.Helper()
	lvl, err := loadLevel("", "1914-08:2p")
	if err != nil {
		t.Fatal(err)
	}
	s, err := dosplay.New(lvl, testSet(t), dosplay.Options{})
	if err != nil {
		t.Fatal(err)
	}
	g := newDosGame(lvl, s, 0, nil, false)
	g.setLocal(m)
	return g, g.dos
}

// keysOf is key set k with ks held.
func keysOf(k gesture.KeySet, ks ...ebiten.Key) gesture.Keys {
	return k.Read(func(key ebiten.Key) bool { return slices.Contains(ks, key) })
}

// moveOne moves the served human's first unit that can move towards the
// enemy and reports its index.
func moveOne(t *testing.T, d *dosControl) int {
	t.Helper()
	for _, at := range ownUnits(d) {
		d.Click(at)
		if !d.selected || len(d.moves) == 0 {
			d.Clear()
			continue
		}
		u := d.unit
		d.Click(nearestEnemy(d, dosplay.Coords(d.moves)))
		d.SkipWalk()
		return u
	}
	t.Fatal("no unit of the served side can move")
	return -1
}

// confirm ends the served human's phase: F1 and F1 in the prompt.
func confirm(d *dosControl) {
	d.EndTurn()
	d.EndTurn()
}

// playLocal plays the resolution back to its end.
func playLocal(t *testing.T, d *dosControl) {
	t.Helper()
	for range 5000 {
		if !d.busy() {
			return
		}
		d.SkipWalk()
		d.SkipScene()
		d.Tick()
	}
	t.Fatal("the resolution did not end")
}

// TestLocalBothHumansConfirm: each human keeps its own selection and gives
// its own orders; no planner confirms for the second human, one confirm never
// resolves, and the second one does.
func TestLocalBothHumansConfirm(t *testing.T) {
	g, d := newLocalTest(t, localSimultaneous)
	d.Tick()
	if d.s.Ready(0) || d.s.Ready(1) {
		t.Fatalf("after the first frame: ready %v %v; no planner may confirm", d.s.Ready(0), d.s.Ready(1))
	}
	at := ownUnits(d)[0]
	d.Click(at)
	if !d.selected {
		t.Fatalf("the left human could not select %v: %q", at, d.Message)
	}
	picked := d.unit
	g.useSeat(1)
	if d.side != 1 || d.selected {
		t.Fatalf("the right human's seat: side %d, selected %v", d.side, d.selected)
	}
	d.Click(ownUnits(d)[0])
	if !d.selected || d.s.Unit(d.unit).Side != 1 {
		t.Fatalf("the right human could not select its unit: %q", d.Message)
	}
	g.useSeat(0)
	if !d.selected || d.unit != picked {
		t.Fatalf("the left human's selection was lost: selected %v unit %d, want %d", d.selected, d.unit, picked)
	}
	d.Clear()
	moveOne(t, d)
	d.EndTurn()
	if looks := d.panelLooks(); looks != [2]panelLook{lookGrey, lookMap} {
		t.Errorf("the left human's prompt: looks %v, want only its panel grey", looks)
	}
	d.EndTurn()
	if !d.s.Ready(0) || d.s.Ready(1) {
		t.Fatalf("after the left human's confirm: ready %v %v", d.s.Ready(0), d.s.Ready(1))
	}
	for range 5 {
		d.Tick()
	}
	if d.busy() || d.s.Ready(1) {
		t.Fatalf("one confirm resolved (busy %v) or confirmed the other side (%v)", d.busy(), d.s.Ready(1))
	}
	phase := d.s.Phase(0)
	g.withSeat(1, func() {
		d.Clear()
		confirm(d)
	})
	if d.side != 0 || !d.s.Ready(1) {
		t.Fatalf("after the right human's confirm: served %d, ready %v", d.side, d.s.Ready(1))
	}
	d.Tick()
	if d.frozen == nil || d.otherFrozen == nil || d.s.Phase(0) == phase {
		t.Fatalf("the second confirm did not resolve: frozen %v %v, phase %d (was %d)", d.frozen != nil, d.otherFrozen != nil, d.s.Phase(0), phase)
	}
	if d.seatOf(1).selected || d.seatOf(1).prompt {
		t.Error("the right human's selection or prompt outlived the resolution")
	}
	playLocal(t, d)
	d.Tick()
	if d.s.Ready(0) || d.s.Ready(1) {
		t.Errorf("after the resolution: ready %v %v; both humans give orders again", d.s.Ready(0), d.s.Ready(1))
	}
}

// TestLocalKeySets: key set B drives the right panel's cursor and its mode
// change while key set A drives the left one's, each leaving the other alone.
func TestLocalKeySets(t *testing.T) {
	g, d := newLocalTest(t, localSimultaneous)
	step := func(c hexgrid.Coord) (gesture.Dir, ebiten.Key, ebiten.Key) {
		if g.playGrid().Contains(gesture.Neighbor(c, gesture.Down)) {
			return gesture.Down, ebiten.KeyArrowDown, ebiten.KeyC
		}
		return gesture.Up, ebiten.KeyArrowUp, ebiten.KeyD
	}
	h0, h1 := g.hover, g.split.seats[1].hover
	dir, _, keyB := step(h1)
	g.feedSecond(keysOf(gesture.KeySetB, keyB))
	g.feedSecond(keysOf(gesture.KeySetB))
	want1 := gesture.Neighbor(h1, dir)
	if d.side != 0 || g.hover != h0 || g.split.seats[1].hover != want1 {
		t.Fatalf("key set B: served %d, left cursor %v (was %v), right cursor %v, want %v", d.side, g.hover, h0, g.split.seats[1].hover, want1)
	}
	dir, keyA, _ := step(h0)
	g.feedGesture(gesture.Input{Keys: keysOf(gesture.KeySetA, keyA)})
	g.feedGesture(gesture.Input{Keys: keysOf(gesture.KeySetA)})
	if g.hover != gesture.Neighbor(h0, dir) || g.split.seats[1].hover != want1 {
		t.Fatalf("key set A: left cursor %v, want %v; right cursor %v, want %v", g.hover, gesture.Neighbor(h0, dir), g.split.seats[1].hover, want1)
	}
	// Ctrl with X runs the mode change on a cell that is no own container:
	// the prompt, then the confirm.
	g.withSeat(1, func() {
		for _, c := range g.splitView(1).Cells() {
			if g.playGrid().Contains(c) && g.panelShows(1, c) && !d.ownContainer(c) {
				g.setCursor(c)
				return
			}
		}
		t.Fatal("no free cell in the right panel")
	})
	modeChange := func() {
		g.feedSecond(keysOf(gesture.KeySetB, ebiten.KeyControlLeft))
		g.feedSecond(keysOf(gesture.KeySetB, ebiten.KeyControlLeft, ebiten.KeyX))
		g.feedSecond(keysOf(gesture.KeySetB))
	}
	modeChange()
	if !d.seatOf(1).prompt || d.prompt || d.s.Ready(1) {
		t.Fatalf("the first mode change: right prompt %v, left prompt %v, ready %v", d.seatOf(1).prompt, d.prompt, d.s.Ready(1))
	}
	if looks := d.panelLooks(); looks != [2]panelLook{lookMap, lookGrey} {
		t.Errorf("the right human's prompt: looks %v", looks)
	}
	var text string
	g.withSeat(1, func() { text = g.wellOf(1).text })
	if text != "CTRL+X: END THE PHASE" {
		t.Errorf("the right well in its prompt: %q", text)
	}
	modeChange()
	if !d.s.Ready(1) || d.s.Ready(0) || d.seatOf(1).prompt {
		t.Fatalf("the second mode change: ready %v %v, prompt %v", d.s.Ready(0), d.s.Ready(1), d.seatOf(1).prompt)
	}
	// Drawing both humans' panels, cursors and wells works.
	img := ebiten.NewImage(1800, 1000)
	g.Draw(img)
}

// TestHotseatHandover: the curtain hides both maps from the start and after
// each confirm until the next human takes over with a key or click; the
// resolution waits for that human, and the other one sees it behind its own
// curtain on its next turn.
func TestHotseatHandover(t *testing.T) {
	g, d := newLocalTest(t, localHotseat)
	img := ebiten.NewImage(1280, 720)
	if side, ok := d.curtain(); !ok || side != 0 || !g.behindCurtain() || !d.covered(0) || !d.covered(1) {
		t.Fatalf("the start: curtain %d %v, covered %v %v", side, ok, d.covered(0), d.covered(1))
	}
	if d.shown() != nil {
		t.Errorf("behind the curtain the panels %v show", d.shown())
	}
	g.Draw(img)
	if g.curtainInput(true, true) {
		t.Fatal("a key still held took over")
	}
	if !g.curtainInput(false, false) || g.behindCurtain() {
		t.Fatal("releasing the key did not take over")
	}
	if d.side != 0 || d.covered(0) || !d.covered(1) {
		t.Fatalf("after the takeover: served %d, covered %v %v", d.side, d.covered(0), d.covered(1))
	}
	if looks := g.splitLooks(); looks[0] != lookMap || looks[1] != lookBlack {
		t.Errorf("the left human's turn: looks %v, want its map and the other panel black", looks)
	}
	g.Draw(img)
	moveOne(t, d)
	d.EndTurn()
	if w := g.wellOf(1); w.text == "FIRE: BACK TO THE ORDERS" {
		t.Errorf("the covered well tells the prompt: %q", w.text)
	}
	d.EndTurn()
	if side, ok := d.curtain(); !ok || side != 1 || d.saveSide() != 1 {
		t.Fatalf("after the left confirm: curtain %d %v, save side %d", side, ok, d.saveSide())
	}
	g.curtainInput(true, false)
	if d.side != 1 || d.covered(1) || !d.covered(0) || g.behindCurtain() {
		t.Fatalf("the right human's turn: served %d, covered %v %v", d.side, d.covered(0), d.covered(1))
	}
	if g.hover != g.startHex(1) {
		t.Errorf("the right human's cursor %v, want its HQ %v", g.hover, g.startHex(1))
	}
	confirm(d)
	if side, ok := d.curtain(); !ok || side != 0 || !d.s.Ready(0) || !d.s.Ready(1) {
		t.Fatalf("after the right confirm: curtain %d %v", side, ok)
	}
	d.Tick()
	if d.busy() {
		t.Fatal("the resolution played behind the curtain")
	}
	g.curtainInput(true, false)
	d.Tick()
	if d.side != 0 || d.frozen == nil || d.pending[1] == nil {
		t.Fatalf("after the takeover: served %d, frozen %v, kept for the right human %v", d.side, d.frozen != nil, d.pending[1] != nil)
	}
	playLocal(t, d)
	confirm(d)
	g.curtainInput(true, false)
	if d.side != 1 || !d.busy() || d.frozen == nil || d.stage != stageBlack || d.pending[1] != nil {
		t.Fatalf("the right human's turn: served %d, busy %v, stage %v, still kept %v", d.side, d.busy(), d.stage, d.pending[1] != nil)
	}
	if looks := g.splitLooks(); looks[0] != lookBlack {
		t.Errorf("the left panel during the right human's playback: %v", looks[0])
	}
	playLocal(t, d)
	if d.s.Ready(1) || !d.s.Ready(0) || g.behindCurtain() {
		t.Errorf("after the playback: ready %v %v, curtain %v", d.s.Ready(0), d.s.Ready(1), g.behindCurtain())
	}
}

// TestLocalFlagMatch: -play makes the battle of -level two humans', whatever
// -opponent says; without it the opponent plays as before.
func TestLocalFlagMatch(t *testing.T) {
	lvl, err := loadLevel("", "1914-08:2p")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range []string{"simultaneous", "hotseat", ""} {
		local, err := parseLocalMode(m)
		if err != nil {
			t.Fatal(err)
		}
		gm, err := newDosMatch(matchConfig{lvl: lvl, set: testSet(t), player: faction.CentralPowers, opponent: "ai", local: local})
		if err != nil {
			t.Fatal(err)
		}
		if gm.dos.local != local || gm.dos.s.Game.Players[1].Computer != (local == localNone) || local.String() != m {
			t.Errorf("-play %q: mode %v, computer %v", m, gm.dos.local, gm.dos.s.Game.Players[1].Computer)
		}
	}
	if _, err := parseLocalMode("both"); err == nil {
		t.Error("an unknown mode parsed")
	}
}

// TestSingleHumanUnchanged: one human has one seat, sees both panels and
// never meets a curtain; its prompt greys both panels as before.
func TestSingleHumanUnchanged(t *testing.T) {
	g, d := newSplitTest(t, passPlanner{}, 1280, 720)
	if d.local != localNone || g.behindCurtain() || d.covered(0) || d.covered(1) || d.saveSide() != d.side {
		t.Fatalf("one human: mode %v, curtain %v, covered %v %v", d.local, g.behindCurtain(), d.covered(0), d.covered(1))
	}
	g.useSeat(1)
	if d.side != 0 {
		t.Fatal("one human's seat was swapped")
	}
	d.EndTurn()
	if looks := d.panelLooks(); looks != [2]panelLook{lookGrey, lookGrey} {
		t.Errorf("one human's prompt: looks %v", looks)
	}
	d.EndTurn()
	d.Tick()
	if d.frozen == nil {
		t.Error("the passing side did not confirm")
	}
}
