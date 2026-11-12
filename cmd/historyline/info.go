package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// infoLines is the info box text, a dev aid shown with the hex numbers only:
// the battle screen has its plates and wells instead.
func (g *game) infoLines() []string {
	if g.dos == nil || !g.hexNumbers {
		return nil
	}
	return g.dosInfoLines()
}

// drawInfo draws the info box at the top right, below the dev indicators.
func (g *game) drawInfo(dst *ebiten.Image) {
	lines := g.infoLines()
	if len(lines) == 0 {
		return
	}
	w := 0
	for _, l := range lines {
		w = max(w, textWidth(l))
	}
	r := g.hud.infoRect(w, len(lines))
	vector.FillRect(dst, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), color.RGBA{20, 20, 40, 220}, false)
	for i, l := range lines {
		ebitenutil.DebugPrintAt(dst, l, r.X+infoPad, r.Y+infoPad/2+i*infoLineHeight)
	}
}
