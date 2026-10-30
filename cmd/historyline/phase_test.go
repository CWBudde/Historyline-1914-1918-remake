package main

import (
	"fmt"
	"image"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dospanel"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/gesture"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/hajimehoshi/ebiten/v2"
)

// The tests below drive the phase change of -rules dos: the prompt, then
// black, the battles, the new mode, the revealed moves and the computer's
// planning.

// confirmOrders asks for the phase end and confirms it (F1 twice).
func confirmOrders(t *testing.T, d *dosControl) {
	t.Helper()
	d.EndTurn()
	if !d.prompt {
		t.Fatalf("F1 opened no prompt: %q", d.Message)
	}
	d.EndTurn()
	if d.prompt || !d.s.Ready(d.side) {
		t.Fatalf("the second F1 did not confirm: prompt %v ready %v", d.prompt, d.s.Ready(d.side))
	}
}

// playOut ticks d until the resolution has played, skipping walks.
func playOut(t *testing.T, d *dosControl) {
	t.Helper()
	for range 100000 {
		if !d.busy() {
			return
		}
		d.SkipWalk()
		d.Tick()
	}
	t.Fatal("the resolution did not end")
}

// toBattles plays phases (advancing, then attacking with every own unit)
// until a resolution brings battles; it returns with that resolution just
// started.
func toBattles(t *testing.T, d *dosControl) {
	t.Helper()
	d.Tick()
	for range 60 {
		if d.s.Ended() != nil {
			break
		}
		if d.s.Phase(d.side) == dosrules.PhaseMovement {
			advance(t, d)
		} else {
			attackAll(t, d)
		}
		confirmOrders(t, d)
		d.Tick()
		for _, ev := range d.queue {
			if ev.Kind == dosplay.Battle {
				return
			}
		}
		reduced := d.reduced
		d.SetReducedMotion(true)
		playOut(t, d)
		d.SetReducedMotion(reduced)
		d.Tick() // the computer gives its orders
	}
	t.Fatal("no battle")
}

// stageLabel names what the panels show now, for the order of the sequence.
func stageLabel(d *dosControl) string {
	l := d.panelLooks()
	switch d.stage {
	case stageBlack:
		if l != [2]panelLook{lookBlack, lookBlack} {
			return "black?"
		}
		return "black"
	case stageBattle:
		switch l[d.side] {
		case lookGrey:
			return "site-grey"
		case lookMap:
			if l[1-d.side] == lookBattle {
				return "scene"
			}
			return "still"
		}
		return "battle?"
	case stageNewMode:
		if l != [2]panelLook{lookGrey, lookGrey} {
			return "newmode?"
		}
		return "newmode"
	case stageNone:
	}
	switch {
	case !d.walk.Done():
		return "reveal"
	case d.planning():
		return "planning"
	case d.busy():
		return "busy"
	}
	return "idle"
}

// sequence ticks d through the resolution just started and returns the
// labels in order (repeats folded) and how many frames each stage ran.
func sequence(t *testing.T, d *dosControl, g *game) (labels []string, frames map[stage][]int) {
	t.Helper()
	frames = map[stage][]int{}
	last, run := stageNone, 0
	for range 100000 {
		if d.stage != stageNone && !d.busy() {
			t.Fatal("a stage runs while the controller is not busy")
		}
		if g != nil {
			checkWells(t, g, d)
		}
		l := stageLabel(d)
		if n := len(labels); n == 0 || labels[n-1] != l {
			labels = append(labels, l)
		}
		if d.stage != last || d.stage == stageBattle && d.stageAt == 0 && run > 0 {
			if last != stageNone {
				frames[last] = append(frames[last], run)
			}
			last, run = d.stage, 0
		}
		run++
		if !d.busy() && (d.planning() || d.s.Ready(1-d.side) || d.over) {
			return labels, frames
		}
		if !d.walk.Done() {
			for range 10 {
				d.Tick()
			}
			d.SkipWalk()
		}
		d.Tick()
	}
	t.Fatal("the sequence did not end")
	return nil, nil
}

// checkWells checks the wells against the stage.
func checkWells(t *testing.T, g *game, d *dosControl) {
	t.Helper()
	if d.over {
		return
	}
	own, other := g.wellOf(d.side), g.wellOf(1-d.side)
	switch d.stage {
	case stageBlack:
		if own.text != "" || other.text != "" || own.unit || other.progress >= 0 {
			t.Fatalf("wells in the black: %+v %+v", own, other)
		}
	case stageBattle:
		b := d.fight.ev.Battle
		if !strings.Contains(own.text, fmt.Sprintf("%d-%d", b.Defender.Before, b.Defender.After)) || other.text != "" {
			t.Fatalf("wells in a battle of %s on %s: %q %q", b.Attacker.Unit.Name, b.Defender.Unit.Name, own.text, other.text)
		}
	case stageNewMode:
		for side, w := range [2]well{g.wellOf(0), g.wellOf(1)} {
			if !strings.Contains(w.text, modeWord(d.s.Phase(side))) || !strings.Contains(w.text, "ROUND") {
				t.Fatalf("side %d's well at the new mode: %q", side, w.text)
			}
		}
	case stageNone:
	}
}

func TestPhaseSequencePrompt(t *testing.T) {
	g, d := newSplitTest(t, passPlanner{}, 1800, 1000)
	at, ok := g.orderableInView()
	if !ok {
		t.Fatal("no orderable unit")
	}
	g.openStrategic(d.side)

	// F1 opens the prompt: both panels grey, the strategic map closed, the
	// wells say how to go on.
	g.applyAction(actEndTurn)
	g.updateSplit()
	if !d.prompt || d.s.Ready(d.side) || d.busy() {
		t.Fatalf("after F1: prompt %v ready %v busy %v", d.prompt, d.s.Ready(d.side), d.busy())
	}
	if d.panelLooks() != [2]panelLook{lookGrey, lookGrey} || g.split.mode != [2]panelMode{} {
		t.Errorf("prompt looks %v, modes %v", d.panelLooks(), g.split.mode)
	}
	if own, other := g.wellOf(d.side), g.wellOf(1-d.side); !strings.Contains(own.text, "F1") || !strings.Contains(other.text, "FIRE") {
		t.Errorf("prompt wells %q / %q", own.text, other.text)
	}

	// No map input: the arrows leave the cursor, a drag gives no order.
	g.setCursor(at)
	feed(g, []gesture.Input{held(gesture.Keys{Right: true}), held(gesture.Keys{})})
	if g.hover != at || !d.prompt {
		t.Errorf("the arrows moved the cursor to %v in the prompt (prompt %v)", g.hover, d.prompt)
	}
	zoom := g.split.screen.Zoom
	p := g.split.screen.Panels[d.side]
	g.wheelSplit(1, p.X+p.W/2, p.Y+p.H/2)
	if g.split.screen.Zoom != zoom {
		t.Error("the wheel zoomed in the prompt")
	}

	// Fire goes back to the orders without selecting anything.
	feed(g, []gesture.Input{held(gesture.Keys{Fire: true}), held(gesture.Keys{})})
	if d.prompt || d.selected || d.s.Ready(d.side) {
		t.Fatalf("fire: prompt %v selected %v ready %v", d.prompt, d.selected, d.s.Ready(d.side))
	}
	if d.panelLooks() != [2]panelLook{} {
		t.Errorf("looks after the prompt %v", d.panelLooks())
	}

	// A left click goes back too.
	g.applyAction(actEndTurn)
	feed(g, clickAt(g, at))
	if d.prompt || d.selected {
		t.Errorf("click: prompt %v selected %v", d.prompt, d.selected)
	}

	// The mode-change gesture opens it, Esc leaves it.
	g.runGesture(gesture.ModeChange, at)
	if !d.prompt {
		t.Fatal("the mode-change gesture opened no prompt")
	}
	g.applyAction(actCancel)
	if d.prompt || d.s.Ready(d.side) {
		t.Errorf("Esc: prompt %v ready %v", d.prompt, d.s.Ready(d.side))
	}

	// F1 twice confirms; F1 again opens the orders while the other side
	// has not confirmed.
	g.applyAction(actEndTurn)
	g.applyAction(actEndTurn)
	if d.prompt || !d.s.Ready(d.side) {
		t.Fatalf("F1 twice: prompt %v ready %v", d.prompt, d.s.Ready(d.side))
	}
	g.applyAction(actEndTurn)
	if d.prompt || d.s.Ready(d.side) || !strings.Contains(d.Message, "open again") {
		t.Errorf("F1 once confirmed: prompt %v ready %v %q", d.prompt, d.s.Ready(d.side), d.Message)
	}
}

func TestPhaseSequenceOrder(t *testing.T) {
	g, d := newSplitTest(t, approachPlanner{}, 1800, 1000)
	toBattles(t, d)
	if d.stage != stageBlack || !d.busy() {
		t.Fatalf("the resolution starts with stage %v", d.stage)
	}
	labels, frames := sequence(t, d, g)
	got := strings.Join(slices.DeleteFunc(labels, func(l string) bool { return l == "busy" }), " ")
	t.Logf("sequence: %s; frames %v", got, frames)
	if !regexp.MustCompile(`^black (site-grey scene )+newmode( reveal)* idle$`).MatchString(got) {
		t.Errorf("sequence %q", got)
	}
	if fs := frames[stageBlack]; len(fs) != 1 || fs[0] != blackFrames {
		t.Errorf("black frames %v, want %d", fs, blackFrames)
	}
	if fs := frames[stageNewMode]; len(fs) != 1 || fs[0] != newModeFrames {
		t.Errorf("new mode frames %v, want %d", fs, newModeFrames)
	}
	for _, n := range frames[stageBattle] {
		if s := float64(n) / tps; s < 2 || s > 8 {
			t.Errorf("a battle ran %d frames", n)
		}
	}
	if d.fx.Banner() != nil {
		t.Errorf("a round banner under -rules dos: %+v", d.fx.Banner())
	}
}

func TestPhaseSequencePlanningAfter(t *testing.T) {
	d := newDosTest(t, "1914-08:cp", computerPlanner())
	opp := 1 - d.side
	tickUntil(t, d, func() bool { return d.s.Ready(opp) })
	confirmOrders(t, d)
	d.Tick()
	if d.stage != stageBlack {
		t.Fatalf("stage %v after both confirmed", d.stage)
	}
	sawNewMode := false
	tickUntil(t, d, func() bool {
		if d.stage == stageNewMode {
			sawNewMode = true
		}
		if d.planning() && d.busy() {
			t.Fatalf("the computer plans while the sequence plays (stage %v)", d.stage)
		}
		d.SkipWalk()
		return d.planning() || d.s.Ready(opp)
	})
	if !sawNewMode {
		t.Error("no new mode before the planning")
	}
}

func TestPhaseSequenceReduced(t *testing.T) {
	d := newDosTest(t, "1914-08:cp", approachPlanner{})
	d.SetReducedMotion(true)
	toBattles(t, d)
	labels, frames := sequence(t, d, nil)
	got := strings.Join(labels, " ")
	if !strings.Contains(got, "scene") || strings.Contains(got, "still") {
		t.Errorf("reduced motion sequence %q", got)
	}
	if fs := frames[stageBlack]; len(fs) != 1 || fs[0] != blackFramesReduced {
		t.Errorf("black frames %v, want %d", fs, blackFramesReduced)
	}
	if fs := frames[stageNewMode]; len(fs) != 1 || fs[0] != newModeFramesReduced {
		t.Errorf("new mode frames %v, want %d", fs, newModeFramesReduced)
	}
	for _, n := range frames[stageBattle] {
		if want := siteGreyReduced + dospanel.BattleTicks(dospanel.Battle{}, true); n != want {
			t.Errorf("a battle ran %d frames, want %d", n, want)
		}
	}
}

func TestBattlePanelSceneToggle(t *testing.T) {
	g, d := newSplitTest(t, approachPlanner{}, 1800, 1000)
	g.applyAction(actScenes)
	if !d.noScenes || !strings.Contains(d.Message, "off") {
		t.Fatalf("F9: off %v, %q", d.noScenes, d.Message)
	}
	toBattles(t, d)
	labels, frames := sequence(t, d, g)
	if got := strings.Join(labels, " "); strings.Contains(got, "scene") || !strings.Contains(got, "still") {
		t.Errorf("sequence without scenes %q", got)
	}
	for _, n := range frames[stageBattle] {
		if n < stillFrames || n > max(stillFrames, siteGreyFrames+deathFrames) {
			t.Errorf("a battle ran %d frames, want %d (longer only for an explosion)", n, stillFrames)
		}
	}
	g.applyAction(actScenes)
	if d.noScenes || !strings.Contains(d.Message, "on") {
		t.Errorf("F9 again: off %v, %q", d.noScenes, d.Message)
	}
}

func TestBattlePanelStages(t *testing.T) {
	g, d := newSplitTest(t, approachPlanner{}, 1800, 1000)
	toBattles(t, d)
	for d.stage != stageBattle {
		d.Tick()
	}
	f := d.fight
	b := f.ev.Battle
	if a, x := f.battle.Attacker, f.battle.Defender; a.Before != b.Attacker.Before || a.After != b.Attacker.After ||
		x.Before != b.Defender.Before || x.After != b.Defender.After || x.Destroyed != b.Defender.Destroyed {
		t.Errorf("scene %+v / %+v for the battle %+v", a, x, b)
	}
	if d.report.Active() {
		t.Error("the HUD battle report shows under -rules dos")
	}
	// The site panel looks at the fight.
	if c, ok := d.takeFollow(d.side); !ok || c != f.at {
		t.Errorf("the site panel looks at %v %v, the fight is at %v", c, ok, f.at)
	}
	// Grey first, then colour with the scene from its start.
	if l := d.panelLooks(); l[d.side] != lookGrey || l[1-d.side] != lookBlack {
		t.Fatalf("battle opens with %v", l)
	}
	for d.panelLooks()[d.side] == lookGrey {
		d.Tick()
	}
	if d.panelLooks()[1-d.side] != lookBattle || d.sceneTick() != 0 {
		t.Fatalf("looks %v at scene tick %d", d.panelLooks(), d.sceneTick())
	}
	// The hits land with the scene's first loss.
	if want := f.grey + dospanel.FirstLoss(f.battle, false); f.strikeAt != want {
		t.Errorf("strike at %d, want %d", f.strikeAt, want)
	}
	for d.stageAt < f.strikeAt-1 {
		d.Tick()
	}
	if f.struck {
		t.Fatalf("struck at %d, before the first loss at %d", d.stageAt, f.strikeAt)
	}
	d.Tick()
	if !f.struck {
		t.Fatalf("not struck at the first loss, frame %d", d.stageAt)
	}
	// Drawing every look works.
	img := ebiten.NewImage(1800, 1000)
	g.drawSplit(img)
	g.drawWells(img)
	// A click skips the battle.
	d.Click(f.ev.To)
	if d.stage == stageBattle && d.fight == f {
		t.Error("a click did not skip the battle")
	}
	// Esc skips the next one.
	for d.busy() && d.stage != stageBattle {
		d.SkipWalk()
		d.Tick()
	}
	if d.stage == stageBattle {
		next := d.fight
		d.Cancel()
		if d.stage == stageBattle && d.fight == next {
			t.Error("Esc did not skip the battle")
		}
	}
}

func TestBattleScene(t *testing.T) {
	ship := dosplay.UnitView{Name: "Battleship", Domain: dosplay.Naval, Full: 1}
	inf := dosplay.UnitView{Name: "Infantry", Domain: dosplay.Infantry, Full: 6}
	s := battleScene(&dosplay.BattleInfo{
		Attacker:   dosplay.Fighter{Unit: inf, Before: 6, After: 5, ExpBefore: 1, ExpAfter: 2},
		Defender:   dosplay.Fighter{Unit: ship, Before: 1, After: 0, Destroyed: true},
		ReturnFire: true,
	})
	a, x := s.Attacker, s.Defender
	if a.Name != "INFANTRY" || a.Ship || a.Before != 6 || a.After != 5 || a.Full != 6 || a.ExpBefore != 1 || a.ExpAfter != 2 {
		t.Errorf("attacker %+v", a)
	}
	if !x.Ship || !x.Destroyed || x.After != 0 || !s.ReturnFire {
		t.Errorf("defender %+v, return fire %v", x, s.ReturnFire)
	}
	if flagImage(state.None) != nil || flagImage(state.CentralPowers) == nil || flagImage(state.Entente) == flagImage(state.CentralPowers) {
		t.Error("flags: none for the neutral side, one per power")
	}
}

// battleOn starts a battle of the own unit at against the nearest enemy of
// the frozen view, the enemy destroyed and the own unit losing a man.
func battleOn(t *testing.T, d *dosControl) (atk, def dosplay.UnitView) {
	t.Helper()
	v := d.s.View(d.side)
	d.frozen = &v
	at := ownUnits(d)[0]
	var to hexgrid.Coord
	best := -1
	for _, c := range dosplay.Coords(toSet(v.Units)) {
		if e := v.Units[c]; e.Side != d.side && (best < 0 || hexgrid.Distance(at, c) < best) {
			to, best = c, hexgrid.Distance(at, c)
		}
	}
	if best < 0 {
		t.Fatal("no enemy in view")
	}
	atk, def = v.Units[at], v.Units[to]
	b := &dosplay.BattleInfo{
		Attacker: dosplay.Fighter{Unit: atk, Before: atk.Size, After: atk.Size - 1, ExpBefore: atk.Exp, ExpAfter: atk.Exp + 1},
		Defender: dosplay.Fighter{Unit: def, Before: def.Size, After: 0, Destroyed: true},
	}
	d.startBattle(dosplay.Event{Kind: dosplay.Battle, From: at, To: to, Battle: b})
	return atk, def
}

// inFrozen is the unit of record index on the frozen view.
func inFrozen(d *dosControl, index int) (dosplay.UnitView, bool) {
	for _, u := range d.frozen.Units {
		if u.Index == index {
			return u, true
		}
	}
	return dosplay.UnitView{}, false
}

func TestBattleDrop(t *testing.T) {
	d := newDosTest(t, "1914-08:cp", passPlanner{})
	atk, def := battleOn(t, d)
	f := d.fight
	if def.Size < 2 {
		t.Fatalf("defender of size %d", def.Size)
	}
	// The map follows the panel man by man: the defender's size counts down
	// one at a time with a hit each, its last man goes up in an explosion and
	// it stays on the map until the explosion's peak.
	sizes, hits, deaths, lastLoss := []int{def.Size}, 0, 0, -1
	for d.stage == stageBattle && !f.dropped[1] {
		before := len(d.fx.Marks())
		d.tickStage()
		for _, m := range d.fx.Marks()[min(before, len(d.fx.Marks())):] {
			if m.At == f.at && m.Kind == markHit {
				hits++
			}
			if m.At == f.at && m.Kind == markDeath {
				deaths++
				lastLoss = d.stageAt
			}
		}
		if u, ok := inFrozen(d, def.Index); ok && u.Size != sizes[len(sizes)-1] {
			sizes = append(sizes, u.Size)
		}
	}
	for i := 1; i < len(sizes); i++ {
		if sizes[i] != sizes[i-1]-1 {
			t.Fatalf("defender sizes on the map %v, want one man at a time", sizes)
		}
	}
	if sizes[len(sizes)-1] != 0 || hits != def.Size-1 || deaths != 1 {
		t.Errorf("sizes %v, %d hits and %d explosions for %d men", sizes, hits, deaths, def.Size)
	}
	peak := explosionPeak * explosionMS * tps / 1000
	if _, ok := inFrozen(d, def.Index); ok || d.stageAt-lastLoss != peak {
		t.Errorf("the destroyed unit left %d frames after its last man (on the map %v), want %d", d.stageAt-lastLoss, ok, peak)
	}
	if d.stageLen < lastLoss+deathFrames {
		t.Errorf("the stage ends at %d, before the explosion at %d is over", d.stageLen, lastLoss)
	}
	// The attacker's loss and gain land as the scene shows them.
	for d.stage == stageBattle {
		d.tickStage()
	}
	if f.landed != [2]int{1, def.Size} || f.gained != [2]bool{true, true} {
		t.Errorf("landed %v, gained %v", f.landed, f.gained)
	}

	// Skipping the battle, or its end, lands it all and drops the destroyed
	// unit at once.
	d = newDosTest(t, "1914-08:cp", passPlanner{})
	atk, def = battleOn(t, d)
	d.SkipScene()
	if _, ok := inFrozen(d, def.Index); ok || d.stage != stageNone {
		t.Errorf("skipped: the destroyed unit stays (%v), stage %v", ok, d.stage)
	}
	if u, _ := inFrozen(d, atk.Index); u.Size != atk.Size-1 || u.Exp != atk.Exp+1 {
		t.Errorf("skipped: attacker %+v, want size %d exp %d", u, atk.Size-1, atk.Exp+1)
	}
	d = newDosTest(t, "1914-08:cp", passPlanner{})
	_, def = battleOn(t, d)
	for !d.fight.struck {
		d.tickStage()
	}
	d.endStage()
	if _, ok := inFrozen(d, def.Index); ok {
		t.Error("the stage ended with the destroyed unit on the map")
	}
}

func TestPhaseSequenceDraw(t *testing.T) {
	g, d := newSplitTest(t, passPlanner{}, 1280, 720)
	img := ebiten.NewImage(1280, 720)
	d.EndTurn()
	g.drawSplit(img)
	d.EndTurn()
	d.Tick()
	if d.stage != stageBlack {
		t.Fatalf("stage %v", d.stage)
	}
	g.drawSplit(img)
	for d.stage != stageNewMode && d.busy() {
		d.Tick()
	}
	g.drawSplit(img)
	g.drawWells(img)
	p := g.split.screen.Panels[d.side]
	if g.split.grey[d.side] == nil || g.split.grey[d.side].Bounds().Size() != image.Pt(p.W, p.H) {
		t.Error("no offscreen panel of the panel's size")
	}
}

func TestBattlePanelRemakeUnaffected(t *testing.T) {
	c := newTestControl(t)
	g := newGame(testLevel(t), c.st, c.player, true)
	g.applyAction(actScenes)
	if c.Message != "" || g.dos != nil {
		t.Errorf("F9 under the remake rules: %q", c.Message)
	}
}

func TestPhaseSequenceCaptureViews(t *testing.T) {
	g, d := newSplitTest(t, passPlanner{}, 1920, 1080)
	d.sceneArt = func() sceneArt { return sceneArt{capture: true, finale: true, surrender: true} }
	want := map[string][2]panelLook{
		"split-prompt":           {lookGrey, lookGrey},
		"split-black":            {lookBlack, lookBlack},
		"split-battle-grey":      {lookGrey, lookBlack},
		"split-battle":           {lookMap, lookBattle},
		"split-battle-loss":      {lookMap, lookBattle},
		"split-battle-destroyed": {lookMap, lookBattle},
		"split-battle-hold":      {lookMap, lookBattle},
		"split-taken-grey":       {lookGrey, lookBlack},
		"split-taken":            {lookMap, lookCaptured},
		"split-taken-plant":      {lookMap, lookCaptured},
		"split-taken-hold":       {lookMap, lookCaptured},
		"split-finale-doors":     {lookMap, lookMap},
		"split-finale":           {lookMap, lookMap},
		"split-surrender-offer":  {lookMap, lookMap},
		"split-surrender":        {lookMap, lookMap},
		"split-newmode":          {lookGrey, lookGrey},
		"split":                  {lookMap, lookMap},
	}
	planned := map[string]bool{}
	for _, f := range splitFramePlan() {
		planned[f.view] = true
	}
	for view, looks := range want {
		if !planned[view] {
			t.Errorf("%s is not captured", view)
		}
		g.stageCapture(view)
		if got := d.panelLooks(); got != looks {
			t.Errorf("%s: looks %v, want %v", view, got, looks)
		}
	}
	g.stageCapture("split-battle")
	if w := g.wellOf(d.side); !strings.ContainsAny(w.text, "0123456789") {
		t.Errorf("battle well %q", w.text)
	}
	// The destroyed defender stands under its explosion.
	g.stageCapture("split-battle-destroyed")
	def := d.fight.ev.Battle.Defender.Unit
	if _, ok := inFrozen(d, def.Index); !ok || !slices.ContainsFunc(d.fx.Marks(), func(m mark) bool { return m.Kind == markDeath }) {
		t.Errorf("destroyed view: defender on the map %v, marks %+v", ok, d.fx.Marks())
	}
	// The building changes hands on the map as the pin goes in; the finale
	// holds after it has played.
	for view, mine := range map[string]bool{"split-taken": false, "split-taken-plant": true} {
		g.stageCapture(view)
		to := d.taking.ev.To
		if got := d.frozen.Buildings[to].Owner == d.s.Faction(d.side); got != mine {
			t.Errorf("%s: the building is ours %v, want %v", view, got, mine)
		}
		if w := g.wellOf(d.side); !strings.Contains(w.text, "CAPTURED") {
			t.Errorf("%s: well %q", view, w.text)
		}
	}
	g.stageCapture("split-finale")
	if d.ending == nil || d.sequence() {
		t.Errorf("finale: ending %v, stage %v", d.ending, d.stage)
	}
	g.stageCapture("split-surrender-offer")
	if d.ending == nil || !d.sequence() || d.surrenderTick() <= 0 {
		t.Errorf("surrender offer: ending %v, stage %v at %d", d.ending, d.stage, d.stageAt)
	}
	g.stageCapture("split-surrender")
	if d.ending == nil || d.sequence() {
		t.Errorf("surrender: ending %v, stage %v", d.ending, d.stage)
	}
	g.stageCapture("split")
	if d.frozen != nil || d.otherFrozen != nil {
		t.Error("a battle view leaves its frozen views behind")
	}
	if d.ending != nil {
		t.Error("the finale outlasts its view")
	}
}
