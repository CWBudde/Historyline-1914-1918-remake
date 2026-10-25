package main

import (
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
	"github.com/hajimehoshi/ebiten/v2"
)

// A two-cell ship (the original's battleship, destroyer and submarine) is two
// unit records on neighbouring hexes, its head and its rear, which dosplay
// links. The map draws it once, with its long family's sprite across both
// hexes and the bow at the head; each half keeps its own markers.

// shipHalves returns the hexes of the head and the rear of the two-cell ship
// whose half u stands on c.
func shipHalves(u dosplay.UnitView, c hexgrid.Coord) (head, rear hexgrid.Coord, ok bool) {
	switch {
	case u.Rear != nil:
		return c, *u.Rear, true
	case u.Head != nil:
		return *u.Head, c, true
	}
	return c, c, false
}

// shipHead resolves the rear of a two-cell ship on c to its head: the unit
// and the hex to select or give an order.
func shipHead(units map[hexgrid.Coord]dosplay.UnitView, c hexgrid.Coord) (dosplay.UnitView, hexgrid.Coord, bool) {
	u, ok := units[c]
	if ok && u.Head != nil {
		if h, ok := units[*u.Head]; ok {
			return h, *u.Head, true
		}
	}
	return u, c, ok
}

// shipFacing is the facing of a ship with its head at world point (hx, hy)
// and its rear at (rx, ry): from the rear toward the head.
func shipFacing(hx, hy, rx, ry float64) unitsprite.Facing {
	f, _ := unitsprite.Toward(hx-rx, hy-ry)
	return f
}

// shipDrawnFrom reports whether the ship with halves head and rear is drawn
// from hex c of dst: from the half lower on the map, so the hexes drawn after
// it lie below, unless that half is out of dst's sight.
func (g *game) shipDrawnFrom(dst *ebiten.Image, c, head, rear hexgrid.Coord) bool {
	later, earlier := head, rear
	if rear.Line > head.Line {
		later, earlier = rear, head
	}
	if c == later {
		return g.shipHalfShown(dst, later)
	}
	return c == earlier && !g.shipHalfShown(dst, later)
}

// shipHalfShown reports whether hex c lies within half a hex of dst's bounds,
// well inside the hexes the map layer draws.
func (g *game) shipHalfShown(dst *ebiten.Image, c hexgrid.Coord) bool {
	if !g.grid.Contains(c) {
		return false
	}
	sx, sy := g.cam.WorldToScreen(hexgrid.Center(c))
	e := g.cam.Effective()
	b := dst.Bounds()
	mx, my := hexgrid.HexWidth/2*e, hexgrid.HexHeight/2*e
	return sx >= float64(b.Min.X)-mx && sx < float64(b.Max.X)+mx && sy >= float64(b.Min.Y)-my && sy < float64(b.Max.Y)+my
}

// drawShip draws the long sprite of a two-cell ship of unit type typ with its
// head at world point (hx, hy) and its rear at (rx, ry), dressed with the rim
// of hex c, and reports whether it could.
func (g *game) drawShip(dst *ebiten.Image, c hexgrid.Coord, typ string, owner state.Player, spent bool, hx, hy, rx, ry float64) bool {
	family, ok := unitsprite.Family(typ)
	if !ok {
		return false
	}
	long, ok := unitsprite.Long(family)
	if !ok {
		return false
	}
	st := unitsprite.Normal
	if spent {
		st = unitsprite.Spent
	}
	sx, sy := g.cam.WorldToScreen((hx+rx)/2, (hy+ry)/2)
	return g.drawSprite(dst, c, long, unitsprite.ID(long, unitsprite.PaletteOf(owner), shipFacing(hx, hy, rx, ry), st), sx, sy, false)
}
