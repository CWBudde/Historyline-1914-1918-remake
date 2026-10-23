package main

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dospanel"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/gesture"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/pixfont"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

// The phase change of -rules dos. F1 (or the mode-change gesture) opens a
// prompt over both panels in grey: F1 again confirms, fire or Esc goes back to
// the orders. Once both sides have confirmed the panels go black with only the
// cursor plates left; each fought attack order then shows the battle site in
// the human side's panel, grey and then in colour, while the other panel
// plays the battle scene (the two groups, their men falling one by one) and
// the site's well counts the losses; the hits land on the map with the scene's
// first loss. The new mode of each side follows on both maps in grey, then
// colour returns, the other side's moves walk and the computer plans. F9 turns
// the scenes off; reduced motion shows only their end and shortens every stage.

// stage is the step of the phase change playing.
type stage int

const (
	stageNone    stage = iota
	stageBlack         // both panels black, the cursor plates on them
	stageBattle        // a fought attack order
	stageNewMode       // the new mode of each side on the grey maps
)

// panelLook is how a panel is drawn.
type panelLook int

const (
	lookMap    panelLook = iota // the map (or strategic map) in colour
	lookGrey                    // the map in greyscale
	lookBlack                   // nothing but black
	lookBattle                  // the battle scene
)

// Stage times in frames (tps per second), the second of each pair with
// reduced motion.
const (
	tps = 60

	blackFrames, blackFramesReduced     = 40, 10
	siteGreyFrames, siteGreyReduced     = 20, 6  // the battle site in grey before the fight
	stillFrames, stillFramesReduced     = 45, 20 // a battle without its scene
	newModeFrames, newModeFramesReduced = 60, 15
)

// fight is the battle of stageBattle. Its fighters are in the scene's order:
// the attacker, then the defender.
type fight struct {
	ev       dosplay.Event
	at       hexgrid.Coord   // the defender's hex, where the site panel looks
	from     hexgrid.Coord   // the attacker's hex
	battle   dospanel.Battle // the scene's numbers; the art is added when it is drawn
	grey     int             // frames the site shows in grey
	scene    bool            // the other panel plays the battle scene
	strikeAt int             // the stage frame of the first loss: a hit without a man lost lands then
	struck   bool            // those hits are marked
	landed   [2]int          // the men each fighter has lost on the map so far
	gained   [2]bool         // each fighter's experience after the battle is on the map
	dying    [2]bool         // the fighter's explosion is on the map
	dropAt   [2]int          // the stage frame at which a dying fighter leaves the map
	dropped  [2]bool         // the fighter has left the map
}

// fighters are the fight's fighters and hexes, attacker first.
func (f *fight) fighters() ([2]dosplay.Fighter, [2]hexgrid.Coord) {
	return [2]dosplay.Fighter{f.ev.Battle.Attacker, f.ev.Battle.Defender}, [2]hexgrid.Coord{f.from, f.at}
}

// final is each fighter's losses and experience gain after the whole battle.
func (f *fight) final() (lost [2]int, gained [2]bool) {
	xs, _ := f.fighters()
	for i, x := range xs {
		lost[i], gained[i] = x.Before-x.After, true
		if x.Destroyed {
			lost[i] = x.Before
		}
	}
	return lost, gained
}

// battleScene is the scene of battle b without its art: the attacker on top,
// the defender below, a ship as one picture.
func battleScene(b *dosplay.BattleInfo) dospanel.Battle {
	fighter := func(x dosplay.Fighter) dospanel.Fighter {
		return dospanel.Fighter{Name: strings.ToUpper(x.Unit.Name), Before: x.Before, After: x.After, Full: x.Unit.Full,
			ExpBefore: x.ExpBefore, ExpAfter: x.ExpAfter, Destroyed: x.Destroyed, Ship: x.Unit.Domain == dosplay.Naval}
	}
	return dospanel.Battle{Attacker: fighter(b.Attacker), Defender: fighter(b.Defender), ReturnFire: b.ReturnFire}
}

// frames is n, or reduced with reduced motion.
func (d *dosControl) frames(n, reduced int) int {
	if d.reduced {
		return reduced
	}
	return n
}

// openPrompt asks the human side to confirm the phase end.
func (d *dosControl) openPrompt() {
	d.prompt = true
	d.Message = "F1 confirms the phase end, fire goes back"
}

// closePrompt goes back to the orders.
func (d *dosControl) closePrompt() {
	d.prompt = false
	d.Message = "back to the orders"
}

// sequence reports whether a stage of the phase change plays.
func (d *dosControl) sequence() bool { return d.stage != stageNone }

// locked reports whether the human side's map takes no input: in the prompt
// and while a stage plays.
func (d *dosControl) locked() bool { return d.prompt || d.sequence() }

// startStage starts st for n frames.
func (d *dosControl) startStage(st stage, n int) { d.stage, d.stageAt, d.stageLen = st, 0, n }

// endStage ends the stage playing; a battle's losses all land and its
// destroyed fighters leave the map now.
func (d *dosControl) endStage() {
	if f := d.fight; f != nil {
		d.land(f.final())
		xs, _ := f.fighters()
		for i, x := range xs {
			if x.Destroyed && !f.dropped[i] {
				f.dropped[i] = true
				d.drop(x.Unit.Index)
			}
		}
	}
	d.stage, d.stageAt, d.stageLen, d.fight = stageNone, 0, 0, nil
}

// tickStage advances the stage by a frame, the map following the battle
// (playFight), and ends the stage when its time is up.
func (d *dosControl) tickStage() {
	d.stageAt++
	if d.stage == stageBattle {
		d.playFight()
	}
	if d.stageAt >= d.stageLen {
		d.endStage()
	}
}

// startBattle shows a fought attack order: the site panel looks at it, the
// fighters turn to each other and the stage runs grey, then the fight.
func (d *dosControl) startBattle(ev dosplay.Event) {
	b := ev.Battle
	atk, def := b.Attacker, b.Defender
	f := &fight{ev: ev, from: d.frozenAt(atk.Unit.Index, ev.From), at: d.frozenAt(def.Unit.Index, ev.To),
		battle: battleScene(b), scene: !d.noScenes}
	f.grey = d.frames(siteGreyFrames, siteGreyReduced)
	f.strikeAt = f.grey
	n := d.frames(stillFrames, stillFramesReduced)
	if f.scene {
		f.strikeAt += dospanel.FirstLoss(f.battle, d.reduced)
		n = f.grey + dospanel.BattleTicks(f.battle, d.reduced)
	}
	d.turn(atk.Unit.Index, f.from, f.at)
	d.turn(def.Unit.Index, f.at, f.from)
	d.lookAt(d.side, f.at)
	d.fight = f
	d.startStage(stageBattle, n)
}

// losses is each fighter's losses and experience gain the battle shows now:
// the scene's count as its men fall, without a scene all at once when the
// site turns to colour.
func (d *dosControl) losses() (lost [2]int, gained [2]bool) {
	f := d.fight
	switch {
	case d.stageAt < f.grey:
		return lost, gained
	case !f.scene:
		return f.final()
	}
	st := dospanel.BattleAt(f.battle, d.sceneTick(), d.reduced)
	for i := range 2 {
		lost[i], gained[i] = st.Sides[i].Lost, st.Gained[i]
	}
	return lost, gained
}

// playFight brings the map up to the battle as it plays: the losses as the
// panel shows them (land), a hit that cost no man with the first loss, and a
// destroyed fighter off the map at its explosion's peak.
func (d *dosControl) playFight() {
	f := d.fight
	d.land(d.losses())
	xs, hexes := f.fighters()
	if !f.struck && d.stageAt >= f.strikeAt {
		f.struck = true
		lost, _ := f.final()
		for i, x := range xs {
			if hit(x) && lost[i] == 0 {
				d.fx.Mark(hexes[i], false)
			}
		}
	}
	for i, x := range xs {
		if f.dying[i] && !f.dropped[i] && d.stageAt >= f.dropAt[i] {
			f.dropped[i] = true
			d.drop(x.Unit.Index)
		}
	}
}

// land puts each fighter's losses and gain on the map: the group size counts
// down with a hit on its hex per step, the last man of a destroyed fighter
// goes up in an explosion that hides the unit at its peak (the stage lasts
// until the explosion is over), and the experience changes with the gain.
func (d *dosControl) land(lost [2]int, gained [2]bool) {
	f := d.fight
	xs, hexes := f.fighters()
	for i, x := range xs {
		if lost[i] <= f.landed[i] && gained[i] == f.gained[i] {
			continue
		}
		hits := lost[i] > f.landed[i]
		f.landed[i], f.gained[i] = max(lost[i], f.landed[i]), gained[i]
		exp := x.ExpBefore
		if gained[i] {
			exp = x.ExpAfter
		}
		d.strength(x.Unit.Index, x.Before-f.landed[i], exp)
		if !hits {
			continue
		}
		dead := x.Destroyed && f.landed[i] >= x.Before
		d.fx.Mark(hexes[i], dead)
		if dead && !f.dying[i] {
			f.dying[i] = true
			f.dropAt[i] = d.stageAt + d.frames(explosionPeak*explosionMS*tps/1000, 0)
			d.stageLen = max(d.stageLen, d.stageAt+d.fx.frames(deathFrames))
		}
	}
}

// SkipBattle ends the battle playing at once (a click or a key).
func (d *dosControl) SkipBattle() {
	if d.stage == stageBattle {
		d.endStage()
	}
}

// ToggleScenes turns the battle scenes off or on (F9).
func (d *dosControl) ToggleScenes() {
	d.noScenes = !d.noScenes
	d.Message = "battle scenes " + map[bool]string{true: "off", false: "on"}[d.noScenes]
}

// startNewMode shows the new mode of each side after the phase change ev.
func (d *dosControl) startNewMode(ev dosplay.Event) {
	d.newMode = ev
	d.startStage(stageNewMode, d.frames(newModeFrames, newModeFramesReduced))
}

// panelLooks is how each panel is drawn now.
func (d *dosControl) panelLooks() [2]panelLook {
	switch d.stage {
	case stageBlack:
		return [2]panelLook{lookBlack, lookBlack}
	case stageNewMode:
		return [2]panelLook{lookGrey, lookGrey}
	case stageBattle:
		var out [2]panelLook
		f := d.fight
		out[d.side], out[1-d.side] = lookMap, lookBlack
		if d.stageAt < f.grey {
			out[d.side] = lookGrey
		} else if f.scene {
			out[1-d.side] = lookBattle
		}
		return out
	case stageNone:
	}
	if d.prompt {
		return [2]panelLook{lookGrey, lookGrey}
	}
	return [2]panelLook{}
}

// sceneTick is the frame of the battle scene playing.
func (d *dosControl) sceneTick() int {
	if d.fight == nil {
		return 0
	}
	return max(d.stageAt-d.fight.grey, 0)
}

// stageWell is side's well in the prompt or a stage, ok false otherwise.
func (g *game) stageWell(side int) (w well, ok bool) {
	d := g.dos
	w.progress = -1
	switch {
	case d.prompt:
		w.text = "FIRE: BACK TO THE ORDERS"
		if side == d.side {
			w.text = "F1: END THE PHASE"
		}
	case d.stage == stageBattle && side == d.side:
		w.text = g.casualties(side, d.fight.ev.Battle)
	case d.stage == stageNewMode:
		w.text = fmt.Sprintf("ROUND %d: %s PHASE", d.newMode.Turn+1, modeWord(d.newMode.Phase[side]))
	case d.sequence():
	default:
		return w, false
	}
	return w, true
}

// casualties is the loss line of a battle, the names shortened (the longer
// first) until it fits side's well: "ATTACKER 6-4 / DEFENDER 5-2".
func (g *game) casualties(side int, b *dosplay.BattleInfo) string {
	a, d := []rune(strings.ToUpper(b.Attacker.Unit.Name)), []rune(strings.ToUpper(b.Defender.Unit.Name))
	line := func() string {
		return fmt.Sprintf("%s %d-%d / %s %d-%d", strings.TrimSpace(string(a)), b.Attacker.Before, b.Attacker.After,
			strings.TrimSpace(string(d)), b.Defender.Before, b.Defender.After)
	}
	if g.split == nil {
		return line()
	}
	sc := g.split.screen
	k := wellScale(sc, side)
	room := sc.WellText(side).W - 2*k
	for {
		if w, _ := pixfont.Small.Measure(line(), k); w <= room || len(a)+len(d) == 0 {
			return line()
		}
		if len(a) >= len(d) {
			a = a[:len(a)-1]
		} else {
			d = d[:len(d)-1]
		}
	}
}

// greyscale draws a panel without colour.
var greyscale = func() colorm.ColorM {
	var m colorm.ColorM
	m.ChangeHSV(0, 0, 1)
	return m
}()

// panelImage is side's offscreen panel, made again when the panel's size
// changed.
func (g *game) panelImage(side int, p dosscreen.Rect) *ebiten.Image {
	img := &g.split.grey[side]
	if *img != nil && (*img).Bounds().Size() != image.Pt(p.W, p.H) {
		(*img).Deallocate()
		*img = nil
	}
	if *img == nil {
		*img = ebiten.NewImage(p.W, p.H)
	}
	return *img
}

// drawPanelLook draws side's panel as look asks; draw draws its map (or
// strategic map) in colour onto an image at screen coordinates shifted by
// (dx, dy).
func (g *game) drawPanelLook(screen *ebiten.Image, side int, look panelLook, draw func(dst *ebiten.Image, dx, dy float64)) {
	p := g.split.screen.Panels[side]
	r := image.Rect(p.X, p.Y, p.Right(), p.Bottom())
	switch look {
	case lookMap:
		draw(screen, 0, 0)
	case lookBlack:
		vector.FillRect(screen, float32(p.X), float32(p.Y), float32(p.W), float32(p.H), color.Black, false)
	case lookBattle:
		g.drawBattle(screen, side, r)
	case lookGrey:
		img := g.panelImage(side, p)
		img.Clear()
		draw(img, float64(-p.X), float64(-p.Y))
		var op colorm.DrawImageOptions
		op.GeoM.Translate(float64(p.X), float64(p.Y))
		colorm.DrawImage(screen, img, greyscale, &op)
	}
}

// drawBattle draws the scene of the battle playing into side's panel r,
// inside the frame's teeth like the plates, with the fighters' photos, their
// sprites facing each other, flags and badges.
func (g *game) drawBattle(dst *ebiten.Image, side int, r image.Rectangle) {
	f := g.dos.fight
	if f == nil {
		return
	}
	b := f.battle
	units := [2]dosplay.UnitView{f.ev.Battle.Attacker.Unit, f.ev.Battle.Defender.Unit}
	for i, x := range [2]*dospanel.Fighter{&b.Attacker, &b.Defender} {
		u := units[i]
		x.Photo = g.portrait(u.ID)
		x.Sprite = g.unitFacing(u.ID, u.Owner, [2]unitsprite.Facing{unitsprite.S, unitsprite.N}[i])
		x.Flag = flagImage(u.Owner)
		x.Badge = g.badgeImage
	}
	k := g.plateScale(side, battleFit)
	// The text scale of the plates, smaller until both names fit theirs.
	t := panelText{textScale(k)}
	room := dospanel.BattleLayout(r, k).Sides[0].Name.Dx() * 9 / 10
	for t.k > 1 {
		wa, _ := t.Measure(b.Attacker.Name, dospanel.Small)
		wd, _ := t.Measure(b.Defender.Name, dospanel.Small)
		if max(wa, wd) <= room {
			break
		}
		t.k--
	}
	dospanel.DrawBattle(dst, r, k, b, g.dos.sceneTick(), g.dos.reduced, g.battleArt(), t)
}

// battleArt is the scene's explosion, empty without all of its frames.
func (g *game) battleArt() dospanel.BattleArt {
	var a dospanel.BattleArt
	for i := range explosionFrames {
		img := g.screens.image(fmt.Sprintf("effect/explosion-%02d", i), 2)
		if img == nil {
			return dospanel.BattleArt{}
		}
		a.Explosion = append(a.Explosion, img)
	}
	return a
}

// flagImages holds each owner's flag as an image, painted once.
var flagImages = map[state.Player]*ebiten.Image{}

// flagImage is owner's flag in the proportions of the well's, nil for the
// neutral side.
func flagImage(owner state.Player) *ebiten.Image {
	if owner == state.None {
		return nil
	}
	img := flagImages[owner]
	if img == nil {
		img = ebiten.NewImage(48, 32)
		drawFlag(img, owner, 0, 0, 48, 32)
		flagImages[owner] = img
	}
	return img
}

// cursorPlates draws the plates of the phase change, one per panel: the human
// side's at its cursor cell, the other side's on its HQ (else the middle of
// its view), where the view shows them.
func (g *game) cursorPlates(dst *ebiten.Image) {
	for side := range 2 {
		c, ok := g.plateCell(side)
		if !ok {
			continue
		}
		cam := g.splitCamera(side)
		x, y := cam.WorldToScreen(hexgrid.Center(c))
		gesture.DrawPlate(dst, x, y, cam.Effective(), gesture.ModeChange, gesture.Normal)
	}
}

// plateCell is where side's plate of the phase change sits.
func (g *game) plateCell(side int) (hexgrid.Coord, bool) {
	v := g.splitView(side)
	if side == g.dos.side {
		return g.hover, g.hoverOnMap && viewShows(v, g.grid, g.hover)
	}
	if hq := g.startHex(side); viewShows(v, g.grid, hq) {
		return hq, true
	}
	c := v.Cell(v.Cols/2, v.Rows/2)
	return c, viewShows(v, g.grid, c)
}
