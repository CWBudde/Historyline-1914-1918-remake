package main

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

// The read side of the original order cycle: what the map, the status line
// and the info box show of a session. There is no fog: every hex is visible,
// and the units are those the side's own map shows.

// newDosGame is the game on a session, played by side.
func newDosGame(lvl *level.Level, s *dosplay.Session, side int, planner dosPlanner, reveal bool) *game {
	g := newMapGame(lvl, reveal)
	g.dos = newDosControl(s, side, planner, reveal)
	g.layoutSplit()
	return g
}

// dosWalking reports whether the unit standing on at is still walking there, on
// the human side's map (the walker is drawn there only).
func (g *game) dosWalking(at hexgrid.Coord) bool {
	w := g.dos.walk
	if g.shownSide() != g.dos.side || w.Done() || w.path[len(w.path)-1] != at {
		return false
	}
	u, ok := g.dos.units(g.reveal)[at]
	return ok && w.unit == u.Index
}

// dosSpent reports whether u is drawn spent: a unit of the map's side with its
// order. The other side's orders stay hidden unless the true view is shown.
func (g *game) dosSpent(u dosplay.UnitView) bool {
	return u.Moved && (u.Side == g.shownSide() || g.reveal)
}

// drawDosUnit draws the unit on c at screen point (sx, sy).
func (g *game) drawDosUnit(dst *ebiten.Image, c hexgrid.Coord, sx, sy float64) {
	u := g.dos.unitsOf(g.shownSide(), g.reveal)[c]
	if g.drawUnitSprite(dst, sx, sy, u.ID, u.Owner, g.dosSpent(u), false, g.dos.Facing(u.Index)) {
		return
	}
	vector.FillCircle(dst, float32(sx), float32(sy), float32(14*g.cam.Effective()), color.RGBA{230, 230, 230, 255}, true)
}

// drawWalker draws the walking unit u (the engine's key) and reports whether
// a sprite was drawn.
func (g *game) drawWalker(dst *ebiten.Image, unit any, sx, sy float64, face unitsprite.Facing) bool {
	switch u := unit.(type) {
	case *state.Unit:
		return u != nil && g.drawUnitSprite(dst, sx, sy, u.Type, u.Owner, unitSpent(u), false, face)
	case int:
		if g.dos == nil {
			return false
		}
		v := g.dos.s.Unit(u)
		return g.drawUnitSprite(dst, sx, sy, v.ID, v.Owner, false, false, face)
	}
	return false
}

// orderColour marks queued attack orders.
var orderColour = color.RGBA{255, 90, 40, 230}

// drawOrders draws the side's queued attack orders: a line from the attacker
// to its target and a ring on the target.
func (g *game) drawOrders(dst *ebiten.Image) {
	if g.dos == nil {
		return
	}
	e, w := float32(g.cam.Effective()), 3*float32(g.view.UI)
	for _, o := range g.dos.s.Orders(g.dos.side) {
		x0, y0 := g.cam.WorldToScreen(hexgrid.Center(o.From))
		x1, y1 := g.cam.WorldToScreen(hexgrid.Center(o.To))
		vector.StrokeLine(dst, float32(x0), float32(y0), float32(x1), float32(y1), w, orderColour, true)
		vector.StrokeCircle(dst, float32(x1), float32(y1), 28*e, w, orderColour, true)
	}
}

// drawOrderNumbers prints each queued attack order's number beside its
// target (HUD layer); orders on the same target stand side by side.
func (g *game) drawOrderNumbers(dst *ebiten.Image) {
	if g.dos == nil {
		return
	}
	per := map[hexgrid.Coord]int{}
	for i, o := range g.dos.s.Orders(g.dos.side) {
		sx, sy := g.cam.WorldToScreen(hexgrid.Center(o.To))
		hx, hy := g.view.toHUD(int(sx), int(sy))
		label := fmt.Sprint(i + 1)
		x, y := hx+12+per[o.To]*(textWidth(label)+6), hy-24
		per[o.To]++
		vector.FillRect(dst, float32(x-2), float32(y), float32(textWidth(label)+4), 16, color.RGBA{60, 20, 10, 230}, false)
		ebitenutil.DebugPrintAt(dst, label, x, y)
	}
}

// dosUnitInfo describes a unit of the session.
func dosUnitInfo(u dosplay.UnitView, own bool) []string {
	out := []string{
		fmt.Sprintf("%s (%v)", u.Name, u.Owner),
		fmt.Sprintf("size %d/%d  exp %d  budget %d", u.Size, u.Full, u.Exp, u.Budget),
	}
	switch {
	case !own:
	case u.Target != nil:
		out = append(out, fmt.Sprintf("attacks hex %d,%d", u.Target.Line, u.Target.Col))
	case u.Moved:
		out = append(out, "has its order")
	}
	return out
}

// dosInfoLines is the dev info box of the order cycle (with the hex numbers):
// the selected unit, the hovered hex with its building and the unit on it.
func (g *game) dosInfoLines() []string {
	d := g.dos
	var out []string
	if d.selected && d.unit >= 0 {
		out = append(out, "selected:")
		out = append(out, dosUnitInfo(d.s.Unit(d.unit), true)...)
	}
	if !g.hoverOnMap {
		return out
	}
	c := g.hover
	if len(out) > 0 {
		out = append(out, "")
	}
	out = append(out, fmt.Sprintf("hex %d,%d:", c.Line, c.Col))
	if c.Line >= 0 && c.Line < len(g.lvl.Lines) && c.Col >= 0 && c.Col < len(g.lvl.Lines[c.Line]) {
		h := g.lvl.Lines[c.Line][c.Col]
		out = append(out, h.Terrain)
		if len(h.Infra.IDs) > 0 {
			out = append(out, strings.Join(h.Infra.IDs, " "))
		}
	}
	if b, ok := d.buildings()[c]; ok {
		kind := "building"
		if b.HQ {
			kind = "HQ"
		}
		out = append(out, fmt.Sprintf("%s, owner %v", kind, b.Owner))
		for _, u := range b.Units {
			out = append(out, fmt.Sprintf("  %s (size %d/%d)", u.Name, u.Size, u.Full))
		}
	}
	if u, ok := d.units(g.reveal)[c]; ok && !g.dosWalking(c) && (!d.selected || u.Index != d.unit) {
		out = append(out, dosUnitInfo(u, u.Side == d.side)...)
	}
	return out
}
