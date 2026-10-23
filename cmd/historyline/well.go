package main

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/pixfont"
)

// The status wells under the panels (-rules dos): one line of capitals per
// side in the panel's colour, numbers in amber. With the cursor on an own unit
// the human side's well shows the unit line (group size, experience badge,
// name, the owner's flag); otherwise each well says what its side is doing,
// and while the computer plans a thin bar along the bottom shows how far it
// got.

// msgFrames is how long a fresh message beats the unit line.
const msgFrames = 180

// well is what one status well shows: a centred line or the unit line, and
// the planning progress from 0 to 1 (negative: no bar).
type well struct {
	text string

	unit  bool // the unit line instead of text
	size  int
	exp   int
	name  string
	owner state.Player

	progress float64
}

// modeWord names a phase as the wells do.
func modeWord(p int) string {
	if p == dosrules.PhaseAction {
		return "ATTACK"
	}
	return "MOVEMENT"
}

// wellOf is side's status well now.
func (g *game) wellOf(side int) well {
	d, s := g.dos, g.dos.s
	w := well{progress: -1}
	if pw, ok := g.stageWell(side); ok && !d.over {
		return pw // the prompt and the phase change (phase.go), already in capitals
	}
	switch {
	case d.over || s.Ended() != nil:
		w.text = d.endText()
	case side != d.side && d.planning():
		w.progress = d.planProgress()
		if s.Phase(side) == dosrules.PhaseAction {
			w.text = "WEIGHING THE ATTACKS"
		} else {
			w.text = "PLANNING THE MOVES"
		}
	case d.busy():
		w.text = "ORDERS UNDER WAY"
	case s.Ready(side):
		w.text = "WAITING: ENEMY MOVES"
		if s.Phase(1-side) == dosrules.PhaseAction {
			w.text = "WAITING: ENEMY ATTACKS"
		}
	case side != d.side:
		w.text = modeWord(s.Phase(side)) + " PHASE"
	case d.Message != "" && g.split.msgAge < msgFrames:
		w.text = d.Message
	default:
		if u, ok := g.cursorUnit(); ok {
			w.unit, w.size, w.exp, w.name, w.owner = true, u.Size, u.Exp, u.Name, u.Owner
			break
		}
		w.text = modeWord(s.Phase(side)) + " PHASE  " + g.orderCount(side)
	}
	w.text = strings.ToUpper(w.text)
	w.name = strings.ToUpper(w.name)
	return w
}

// cursorUnit is the human side's own unit under the cursor, if any.
func (g *game) cursorUnit() (u unitLine, ok bool) {
	if !g.hoverOnMap || g.dosWalking(g.hover) {
		return u, false
	}
	v, ok := g.dos.unitsOf(g.dos.side, false)[g.hover]
	if !ok || v.Side != g.dos.side {
		return u, false
	}
	return unitLine{Size: v.Size, Exp: v.Exp, Name: v.Name, Owner: v.Owner}, true
}

// unitLine is the unit line's data.
type unitLine struct {
	Size  int
	Exp   int
	Name  string
	Owner state.Player
}

// orderCount is side's orders this phase, against the limit if there is one.
func (g *game) orderCount(side int) string {
	s := g.dos.s
	n := len(s.Game.Players[side].Orders)
	if s.Game.Limit > 0 {
		return fmt.Sprintf("%d/%d ORDERS", n, s.Game.Limit)
	}
	return count(n, "ORDER")
}

// tickWells ages the human side's message (updateSplit).
func (g *game) tickWells() {
	s := g.split
	if m := g.dos.Message; m != s.lastMsg {
		s.lastMsg, s.msgAge = m, 0
		return
	}
	s.msgAge++
}

// wellScale is the font scale (device pixels per font pixel) of the wells:
// the art scale times the UI scale, as large as the text area holds.
func wellScale(sc dosscreen.Screen, side int) int {
	r := sc.WellText(side)
	_, h := pixfont.Small.Measure("A", 1)
	return max(1, min(int(math.Floor(2*sc.UI+1e-9)), r.H/h))
}

// drawWells draws both status wells on the screen, in device pixels.
func (g *game) drawWells(dst *ebiten.Image) {
	sc := g.split.screen
	for side := range 2 {
		drawWell(dst, sc, side, g.wellOf(side), g.drawBadge)
	}
}

// drawWell draws w into side's well of sc, the unit line's experience
// through badge.
func drawWell(dst *ebiten.Image, sc dosscreen.Screen, side int, w well, badge func(dst *ebiten.Image, exp int, cx, cy, size float64)) {
	r := sc.WellText(side)
	k := wellScale(sc, side)
	f := pixfont.Small
	text, num := dosscreen.TextColour(side), dosscreen.NumberColour
	_, th := f.Measure("A", k)
	y := r.Y + (r.H-th)/2
	if w.unit {
		size := fmt.Sprint(w.size)
		fh := th
		fw := fh * 8 / 5
		f.Draw(dst, size, r.X+k, y, k, num)
		sw, _ := f.Measure(size, k)
		if badge != nil && w.exp > 0 {
			bs := math.Min(float64(th)*1.8, float64(r.H))
			badge(dst, w.exp, float64(r.X+sw+3*k)+bs/2, float64(y)+float64(th)/2, bs)
			sw += 3*k + int(bs)
		}
		room := r.W - 2*max(sw+2*k, fw+2*k)
		name := f.Fit(w.name, k, room)
		nw, _ := f.Measure(name, k)
		f.Draw(dst, name, r.X+(r.W-nw)/2, y, k, text)
		drawFlag(dst, w.owner, r.Right()-k-fw, y, fw, fh)
	} else {
		line := f.Fit(w.text, k, r.W-2*k)
		lw, _ := f.Measure(line, k)
		f.DrawNumbers(dst, line, r.X+(r.W-lw)/2, y, k, text, num)
	}
	if w.progress >= 0 {
		b := sc.WellBar(side)
		fill := int(math.Round(float64(b.W) * min(w.progress, 1)))
		vector.FillRect(dst, float32(b.X), float32(b.Y), float32(fill), float32(b.H), dosscreen.BarColour, false)
	}
}

// Flag colours: the German Empire's black, white and red bands, the French
// blue, white and red.
var (
	flagBlack = color.RGBA{0x10, 0x10, 0x10, 0xFF}
	flagWhite = color.RGBA{0xEE, 0xEE, 0xE6, 0xFF}
	flagRed   = color.RGBA{0xC8, 0x20, 0x20, 0xFF}
	flagBlue  = color.RGBA{0x20, 0x38, 0xA0, 0xFF}
)

// drawFlag draws owner's flag in the w × h rectangle at (x, y): the Central
// Powers in horizontal bands, the Entente in vertical ones.
func drawFlag(dst *ebiten.Image, owner state.Player, x, y, w, h int) {
	bands := [3]color.RGBA{flagBlack, flagWhite, flagRed}
	across := owner != state.Entente
	if !across {
		bands = [3]color.RGBA{flagBlue, flagWhite, flagRed}
	}
	for i, c := range bands {
		if across {
			y0, y1 := y+h*i/3, y+h*(i+1)/3
			vector.FillRect(dst, float32(x), float32(y0), float32(w), float32(y1-y0), c, false)
			continue
		}
		x0, x1 := x+w*i/3, x+w*(i+1)/3
		vector.FillRect(dst, float32(x0), float32(y), float32(x1-x0), float32(h), c, false)
	}
}
