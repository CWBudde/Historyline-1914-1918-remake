package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/pixfont"
)

// Unit markers (-rules dos): every unit on the map carries its experience
// badge above the sprite and its group size at the lower right of its hex,
// for both sides. The original shows both only in the status line; the map
// markers make them readable at a glance.

const (
	// marksMinScale is the effective scale below which a hex is too small
	// for its markers.
	marksMinScale = 0.35
	// badgeSize is the badge's edge in world pixels; badgeY lifts its centre
	// above the hex centre.
	badgeSize, badgeY = 30.0, 28.0
	// sizeX, sizeY place the group size plate's centre from the hex centre,
	// and sizeH is the digits' height in world pixels.
	sizeX, sizeY, sizeH = 30.0, 24.0, 11.0
)

// badgeRanks is the rank art of each experience level 1–6 (the art has 11
// ranks for the remake rules: four stars, then the laurel rings).
var badgeRanks = [...]string{"", "rank/01", "rank/02", "rank/03", "rank/04", "rank/05", "rank/10"}

// badgeImage is the rank art of experience exp, nil at 0 or without the
// art.
func (g *game) badgeImage(exp int) *ebiten.Image {
	if exp <= 0 {
		return nil
	}
	return g.screens.image(badgeRanks[min(exp, len(badgeRanks)-1)], 2)
}

// marksShown reports whether the map at effective scale eff draws markers.
func marksShown(eff float64) bool { return eff >= marksMinScale }

// unitMark is the markers of one unit: its view and the screen points of the
// badge's and the size plate's centres.
type unitMark struct {
	at             hexgrid.Coord
	unit           dosplay.UnitView
	badgeX, badgeY float64
	sizeX, sizeY   float64
}

// unitMarks are the markers of every unit of the shown side's view on a hex
// in inView, in inView's order: none on a walking unit's hex or on a ship's
// rear (the ship is one unit, marked on its head), and none at all when the
// hexes are too small.
func (g *game) unitMarks(inView []hexgrid.Coord) []unitMark {
	eff := g.cam.Effective()
	if g.dos == nil || !marksShown(eff) {
		return nil
	}
	units := g.dos.unitsOf(g.shownSide(), g.reveal)
	var out []unitMark
	for _, c := range inView {
		u, ok := units[c]
		if !ok || g.walking(c) || u.Size <= 0 {
			continue
		}
		if u.Head != nil {
			if _, ok := units[*u.Head]; ok {
				continue // a ship is one unit: its head's hex carries the marks
			}
		}
		sx, sy := g.cam.WorldToScreen(hexgrid.Center(c))
		out = append(out, unitMark{at: c, unit: u, badgeX: sx, badgeY: sy - badgeY*eff, sizeX: sx + sizeX*eff, sizeY: sy + sizeY*eff})
	}
	return out
}

// drawUnitMarks draws the markers of the units on a hex in inView.
func (g *game) drawUnitMarks(dst *ebiten.Image, inView []hexgrid.Coord) {
	eff := g.cam.Effective()
	for _, m := range g.unitMarks(inView) {
		g.drawBadge(dst, m.unit.Exp, m.badgeX, m.badgeY, badgeSize*eff)
		drawSizePlate(dst, m.unit, m.sizeX, m.sizeY, eff)
	}
}

// drawBadge draws the badge of experience exp centred on (cx, cy), size
// pixels across: the rank art, or brass marks drawn in code without it.
func (g *game) drawBadge(dst *ebiten.Image, exp int, cx, cy, size float64) {
	if exp <= 0 {
		return
	}
	if img := g.badgeImage(exp); img != nil {
		b := img.Bounds()
		s := size / float64(max(b.Dx(), b.Dy()))
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(-float64(b.Dx())/2, -float64(b.Dy())/2)
		op.GeoM.Scale(s, s)
		op.GeoM.Translate(cx, cy)
		op.Filter = ebiten.FilterLinear
		dst.DrawImage(img, &op)
		return
	}
	drawBrass(dst, exp, cx, cy, size)
}

// The brass of the badges drawn in code.
var (
	brassLight = color.RGBA{0xFF, 0xF0, 0xC8, 0xFF}
	brassMid   = color.RGBA{0xD2, 0xB0, 0x6A, 0xFF}
	brassDark  = color.RGBA{0x5A, 0x40, 0x18, 0xFF}
)

// drawBrass draws experience exp as brass marks centred on (cx, cy), size
// pixels across: one to four beads, then a diamond, then a medal.
func drawBrass(dst *ebiten.Image, exp int, cx, cy, size float64) {
	r := float32(size / 7)
	bead := func(x, y float64) {
		vector.FillCircle(dst, float32(x), float32(y), r*1.35, brassDark, true)
		vector.FillCircle(dst, float32(x), float32(y), r, brassMid, true)
		vector.FillCircle(dst, float32(x)-r/3, float32(y)-r/3, r/2.5, brassLight, true)
	}
	q := size / 4
	switch exp {
	case 1:
		bead(cx, cy)
	case 2:
		bead(cx-q, cy)
		bead(cx+q, cy)
	case 3:
		bead(cx-q, cy+q/2)
		bead(cx+q, cy+q/2)
		bead(cx, cy-q/2)
	case 4:
		bead(cx-q, cy)
		bead(cx+q, cy)
		bead(cx, cy-q)
		bead(cx, cy+q)
	case 5:
		var p vector.Path
		h := float32(size * 0.45)
		x, y := float32(cx), float32(cy)
		p.MoveTo(x, y-h)
		p.LineTo(x+h, y)
		p.LineTo(x, y+h)
		p.LineTo(x-h, y)
		p.Close()
		var so vector.StrokeOptions
		so.Width = r / 1.5
		vector.FillPath(dst, &p, &vector.FillOptions{}, pathOptions(brassMid))
		vector.StrokePath(dst, &p, &so, pathOptions(brassDark))
	default:
		x, y, h := float32(cx), float32(cy), float32(size*0.45)
		vector.FillCircle(dst, x, y, h, brassDark, true)
		vector.FillCircle(dst, x, y, h*0.82, brassMid, true)
		vector.FillRect(dst, x-h*0.12, y-h*0.6, h*0.24, h*1.2, brassLight, true)
		vector.FillRect(dst, x-h*0.6, y-h*0.12, h*1.2, h*0.24, brassLight, true)
	}
}

// sizePlate is the dark plate behind a group size.
var sizePlate = color.RGBA{0x14, 0x16, 0x08, 0xD8}

// drawSizePlate draws u's group size as amber digits on a dark plate edged
// in its owner's colour, centred on (cx, cy) at effective scale eff.
func drawSizePlate(dst *ebiten.Image, u dosplay.UnitView, cx, cy, eff float64) {
	f := pixfont.Small
	_, h1 := f.Measure("0", 1)
	k := max(1, int(math.Round(sizeH*eff/float64(h1))))
	s := fmt.Sprint(u.Size)
	w, h := f.Measure(s, k)
	pad := k
	x0, y0 := int(math.Round(cx))-w/2-pad, int(math.Round(cy))-h/2-pad
	pw, ph := w+2*pad, h+2*pad
	edge := max(1, k/2)
	vector.FillRect(dst, float32(x0-edge), float32(y0-edge), float32(pw+2*edge), float32(ph+2*edge), ownerColour(u.Owner), false)
	vector.FillRect(dst, float32(x0), float32(y0), float32(pw), float32(ph), sizePlate, false)
	f.Draw(dst, s, x0+pad, y0+pad, k, dosscreen.NumberColour)
}
