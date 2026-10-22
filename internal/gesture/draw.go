package gesture

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DrawRing draws the cursor: a steel band around the cell centred on (cx, cy)
// at scale device px per world px, transparent inside.
func DrawRing(dst *ebiten.Image, cx, cy, scale float64) { draw(dst, ringLayers(cx, cy, scale)) }

// DrawPlate draws the action plate a on the cell centred on (cx, cy): the
// cell polygon filled with raised steel and an embossed glyph.
func DrawPlate(dst *ebiten.Image, cx, cy, scale float64, a Action, s State) {
	draw(dst, plateLayers(cx, cy, scale, a, s))
}

// DrawRingMenu draws the ring r: a cancel plate on the cursor and a plate for
// every action, the one under the pointer (hover, or -1) pressed.
func DrawRingMenu(dst *ebiten.Image, r Ring, hover int) {
	DrawPlate(dst, r.X, r.Y, r.Scale, Cancel, Normal)
	for i, a := range r.Actions {
		s := Normal
		if i == hover {
			s = Pressed
		}
		x, y := r.Pos(i)
		DrawPlate(dst, x, y, r.Scale, a, s)
	}
}

func draw(dst *ebiten.Image, ls []layer) {
	for _, l := range ls {
		var p vector.Path
		for _, poly := range l.polys {
			if len(poly) < 3 {
				continue
			}
			p.MoveTo(float32(poly[0][0]), float32(poly[0][1]))
			for _, pt := range poly[1:] {
				p.LineTo(float32(pt[0]), float32(pt[1]))
			}
			p.Close()
		}
		op := &vector.DrawPathOptions{AntiAlias: true}
		op.ColorScale.ScaleWithColor(l.c)
		vector.FillPath(dst, &p, nil, op)
	}
}
