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
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/gesture"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/pixfont"
)

// The phase change of -rules dos. F1 (or the mode-change gesture) opens a
// prompt over both panels in grey: F1 again confirms, fire or Esc goes back to
// the orders. Once both sides have confirmed the panels go black with only the
// cursor plates left; each fought attack order then shows the battle site in
// the human side's panel, grey and then in colour, while the other panel
// plays the sepia vignette of the attacker's domain and the site's well counts
// the losses. The new mode of each side follows on both maps in grey, then
// colour returns, the other side's moves walk and the computer plans. F9 turns
// the vignettes off; reduced motion skips them and shortens every stage.

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
	lookMap      panelLook = iota // the map (or strategic map) in colour
	lookGrey                      // the map in greyscale
	lookBlack                     // nothing but black
	lookVignette                  // the battle vignette
)

// Stage times in frames (tps per second), the second of each pair with
// reduced motion.
const (
	tps = 60

	blackFrames, blackFramesReduced     = 40, 10
	siteGreyFrames, siteGreyReduced     = 20, 6  // the battle site in grey before the fight
	stillFrames, stillFramesReduced     = 45, 20 // a battle without its vignette
	newModeFrames, newModeFramesReduced = 60, 15

	vignetteLoops = 3 // a vignette plays this often, about 2 s
)

// fight is the battle of stageBattle.
type fight struct {
	ev       dosplay.Event
	at       hexgrid.Coord // the defender's hex, where the site panel looks
	from     hexgrid.Coord // the attacker's hex
	domain   dospanel.Domain
	grey     int  // frames the site shows in grey
	vignette bool // the other panel plays the vignette
	struck   bool // the hit marks are up
}

// vignetteDomain is the vignette of a unit's battle domain (the two lists
// differ in order, so they meet by name).
func vignetteDomain(d dosplay.Domain) dospanel.Domain {
	for v := range dospanel.Domains {
		if v.String() == d.String() {
			return v
		}
	}
	return dospanel.Infantry
}

// animTicks is the length in frames of one loop of an animation whose frames
// show frameMS milliseconds each.
func animTicks(frameMS []int) int {
	ms := 0
	for _, f := range frameMS {
		ms += f
	}
	return (ms*tps + 999) / 1000
}

// animFrame is the frame of such an animation ticks frames after it started,
// looping.
func animFrame(frameMS []int, ticks int) int {
	ms := 0
	for _, f := range frameMS {
		ms += f
	}
	if ms <= 0 {
		return 0
	}
	t := ticks * 1000 / tps % ms
	for i, f := range frameMS {
		if t < f {
			return i
		}
		t -= f
	}
	return 0
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

// endStage ends the stage playing.
func (d *dosControl) endStage() { d.stage, d.stageAt, d.stageLen, d.fight = stageNone, 0, 0, nil }

// tickStage advances the stage by a frame: the battle's hits come up when
// its site turns to colour, and the stage ends when its time is up.
func (d *dosControl) tickStage() {
	d.stageAt++
	if f := d.fight; d.stage == stageBattle && !f.struck && d.stageAt >= f.grey {
		d.strike()
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
		domain: vignetteDomain(atk.Unit.Domain), vignette: !d.reduced && !d.noVignettes}
	f.grey = d.frames(siteGreyFrames, siteGreyReduced)
	n := d.frames(stillFrames, stillFramesReduced)
	if f.vignette {
		n = f.grey + vignetteLoops*animTicks(dospanel.Vignette(f.domain).FrameMS)
	}
	d.turn(atk.Unit.Index, f.from, f.at)
	d.turn(def.Unit.Index, f.at, f.from)
	d.lookAt(d.side, f.at)
	d.fight = f
	d.startStage(stageBattle, n)
}

// strike puts the hit marks up and takes the destroyed units off the frozen
// views.
func (d *dosControl) strike() {
	f := d.fight
	b := f.ev.Battle
	f.struck = true
	d.fx.AddHits(f.from, f.at, hit(b.Defender), b.Defender.Destroyed, hit(b.Attacker), b.Attacker.Destroyed)
	for _, x := range []dosplay.Fighter{b.Defender, b.Attacker} {
		if x.Destroyed {
			d.drop(x.Unit.Index)
		}
	}
}

// SkipBattle ends the battle playing at once (a click or a key).
func (d *dosControl) SkipBattle() {
	if d.stage != stageBattle {
		return
	}
	if !d.fight.struck {
		d.strike()
	}
	d.endStage()
}

// ToggleVignettes turns the battle vignettes off or on (F9).
func (d *dosControl) ToggleVignettes() {
	d.noVignettes = !d.noVignettes
	d.Message = "battle scenes " + map[bool]string{true: "off", false: "on"}[d.noVignettes]
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
		} else if f.vignette {
			out[1-d.side] = lookVignette
		}
		return out
	case stageNone:
	}
	if d.prompt {
		return [2]panelLook{lookGrey, lookGrey}
	}
	return [2]panelLook{}
}

// vignetteFrame is the vignette playing and its frame.
func (d *dosControl) vignetteFrame() (dospanel.Domain, int) {
	f := d.fight
	if f == nil {
		return dospanel.Infantry, 0
	}
	return f.domain, animFrame(dospanel.Vignette(f.domain).FrameMS, max(d.stageAt-f.grey, 0))
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
	case lookVignette:
		dom, frame := g.dos.vignetteFrame()
		dospanel.DrawVignette(screen, r, dom, frame)
	case lookGrey:
		img := g.panelImage(side, p)
		img.Clear()
		draw(img, float64(-p.X), float64(-p.Y))
		var op colorm.DrawImageOptions
		op.GeoM.Translate(float64(p.X), float64(p.Y))
		colorm.DrawImage(screen, img, greyscale, &op)
	}
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
