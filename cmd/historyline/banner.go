package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Display times in frames (60 TPS). Every effect is bounded and ends by itself.
const (
	hitFrames     = 24
	deathFrames   = explosionFrames * explosionMS * tps / 1000 // the explosion once
	roundFrames   = 120
	victoryFrames = 300
	captureFrames = 120

	// With reduced motion marks and banners are static and short.
	reducedFrames = 20
)

// The explosion of a destroyed unit (art effect/explosion-00..08, the death
// animation of assets/interface/animations.json): its frames, each shown
// explosionMS milliseconds, and the frame at which the fireball hides the
// unit.
const (
	explosionFrames = 9
	explosionMS     = 100
	explosionPeak   = 3

	// hitLast is the last frame of a hit's burst: the first frames, as
	// long as the hit's mark.
	hitLast = hitFrames*1000/tps/explosionMS - 1
)

type markKind int

const (
	markHit markKind = iota
	markDeath
)

// mark is a short flash on one hex: a hit marker, or a death flash when the
// unit there was destroyed.
type mark struct {
	At     hexgrid.Coord
	Kind   markKind
	Frames int
	Total  int
}

// Alpha is the fade of the mark, 1 at the start and 0 when it is gone.
func (m mark) Alpha() float64 {
	if m.Total <= 0 {
		return 0
	}
	return float64(m.Frames) / float64(m.Total)
}

// frame is the explosion frame of the mark now: a death plays them all, a
// hit the first hitLast+1 as a small burst. With reduced motion the mark
// shows its last frame briefly and still.
func (m mark) frame(reduced bool) int {
	last := explosionFrames - 1
	if m.Kind == markHit {
		last = hitLast
	}
	if reduced {
		return last
	}
	return min(max((m.Total-m.Frames)*1000/tps/explosionMS, 0), last)
}

// staticAlpha is the opacity of a mark with reduced motion: no fade.
const staticAlpha = 1.0

// banner is a line of text across the screen (round start, HQ capture).
type banner struct {
	Text   string
	Frames int
}

// effects holds the transient combat and turn presentation. It never reads or
// changes game state; control feeds it from engine results. The zero value is
// ready to use and every method accepts a nil receiver.
type effects struct {
	marks  []mark
	banner *banner

	Reduced bool // reduced motion: effects are short and do not fade or grow
}

// frames returns n, or the short static time with reduced motion.
func (e *effects) frames(n int) int {
	if e.Reduced {
		return min(n, reducedFrames)
	}
	return n
}

// Show puts text up as the banner for frames frames (fewer with reduced
// motion); it replaces the banner shown.
func (e *effects) Show(text string, frames int) {
	if e == nil {
		return
	}
	e.banner = &banner{Text: text, Frames: e.frames(frames)}
}

// Capture shows the capture-in-progress banner for a depot entered by a unit.
func (e *effects) Capture(unit string) { e.Show("CAPTURING DEPOT - "+unit, captureFrames) }

// AddBattle marks the defender and, when it fired back or died, the attacker,
// from the engine's Battle result.
func (e *effects) AddBattle(from, to hexgrid.Coord, b state.Battle) {
	e.AddHits(from, to, b.Damage > 0, b.DefenderDestroyed, b.CounterDamage > 0, b.AttackerDestroyed)
}

// Mark marks the hex at with a hit, or with an explosion when its unit was
// destroyed.
func (e *effects) Mark(at hexgrid.Coord, destroyed bool) {
	if e == nil {
		return
	}
	k, n := markHit, e.frames(hitFrames)
	if destroyed {
		k, n = markDeath, e.frames(deathFrames)
	}
	e.marks = append(e.marks, mark{At: at, Kind: k, Frames: n, Total: n})
}

// AddHits marks the defender at to when it was hit or destroyed, and the
// attacker at from when the reply hit or destroyed it.
func (e *effects) AddHits(from, to hexgrid.Coord, defHit, defDead, atkHit, atkDead bool) {
	if defHit || defDead {
		e.Mark(to, defDead)
	}
	if atkHit || atkDead {
		e.Mark(from, atkDead)
	}
}

// Round shows the round banner after a turn ended.
func (e *effects) Round(round int, turn state.Player) {
	e.Show(fmt.Sprintf("ROUND %d - %v to move", round, turn), roundFrames)
}

// Victory shows the HQ-capture banner; it replaces a round banner.
func (e *effects) Victory(winner state.Player) {
	if winner == state.None {
		return
	}
	e.Show(fmt.Sprintf("%v captured the HQ and wins", winner), victoryFrames)
}

// Tick counts every effect down and drops finished ones.
func (e *effects) Tick() {
	if e == nil {
		return
	}
	kept := e.marks[:0]
	for _, m := range e.marks {
		if m.Frames--; m.Frames > 0 {
			kept = append(kept, m)
		}
	}
	e.marks = kept
	if e.banner != nil {
		if e.banner.Frames--; e.banner.Frames <= 0 {
			e.banner = nil
		}
	}
}

// Dismiss ends the banner early (a click); marks are too short to need it.
func (e *effects) Dismiss() {
	if e != nil {
		e.banner = nil
	}
}

// Banner returns the active banner, or nil.
func (e *effects) Banner() *banner {
	if e == nil {
		return nil
	}
	return e.banner
}

// Marks returns the active hex marks.
func (e *effects) Marks() []mark {
	if e == nil {
		return nil
	}
	return e.marks
}

// drawMarks paints hit and death marks over their hexes (map layer, device pixels).
func (g *game) drawMarks(dst *ebiten.Image) {
	fx := &g.drv().ui().fx
	for _, m := range fx.Marks() {
		if g.fogOf(m.At) != state.Visible {
			continue // no hint of what the player cannot see
		}
		cx, cy := hexgrid.Center(m.At)
		sx, sy := g.cam.WorldToScreen(cx, cy)
		alpha := m.Alpha()
		if fx.Reduced {
			alpha = staticAlpha
		}
		a := uint8(255 * alpha)
		r := float32(g.cam.Effective()) * 34
		if g.drawEffect(dst, m, fx.Reduced, sx, sy, alpha) {
			continue
		}
		if m.Kind == markDeath {
			if !fx.Reduced {
				r *= 2 - float32(alpha)
			}
			vector.FillCircle(dst, float32(sx), float32(sy), r, color.RGBA{255, 140, 0, a / 2}, true)
			continue
		}
		red, w := color.RGBA{255, 40, 40, a}, 4*float32(g.view.UI)
		vector.StrokeLine(dst, float32(sx)-r/2, float32(sy)-r/2, float32(sx)+r/2, float32(sy)+r/2, w, red, true)
		vector.StrokeLine(dst, float32(sx)-r/2, float32(sy)+r/2, float32(sx)+r/2, float32(sy)-r/2, w, red, true)
	}
}

// effectScale is the size of a death's explosion on the map against its
// logical size (96 wide, it covers its hex and a little beyond); hitScale
// is a hit's burst against that.
const effectScale, hitScale = 1.3, 0.55

// drawEffect draws mark m with the explosion art centred on (sx, sy): a
// death's frame whole, a hit's small and fading with alpha. It reports false
// without the art, and the caller draws the plain mark.
func (g *game) drawEffect(dst *ebiten.Image, m mark, reduced bool, sx, sy, alpha float64) bool {
	eff := g.cam.Effective()
	d := 1
	if eff > 1 {
		d = 2
	}
	img := g.screens.image(fmt.Sprintf("effect/explosion-%02d", m.frame(reduced)), d)
	if img == nil {
		return false
	}
	s := effectScale * eff / float64(d)
	if m.Kind == markHit {
		s *= hitScale
	} else {
		alpha = 1
	}
	b := img.Bounds()
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(-float64(b.Dx())/2, -float64(b.Dy())/2)
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(sx, sy)
	op.Filter = ebiten.FilterLinear
	op.ColorScale.ScaleAlpha(float32(alpha))
	dst.DrawImage(img, &op)
	return true
}

// drawBanner paints the active banner across the HUD.
func (g *game) drawBanner(dst *ebiten.Image) {
	if b := g.drv().ui().fx.Banner(); b != nil {
		h := g.hud.Banner.H
		y := float32(g.hud.Banner.Y)
		vector.FillRect(dst, 0, y, float32(g.hud.Banner.W), float32(h), color.RGBA{20, 20, 40, 220}, false)
		ebitenutil.DebugPrintAt(dst, b.Text, (g.w-textWidth(b.Text))/2, int(y)+h/2-8)
	}
}
