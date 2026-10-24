package main

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

// origArt is the original game's map art: the terrain tiles, summer and
// winter, by terrain record, and the unit sprites. F8 draws them in place of
// the remake's art under -rules dos, for comparison on a local machine: only
// builds with -tags origtiles (just play-original) can load them, from the
// git-ignored original/ folder; other builds answer F8 with a message.
type origArt struct {
	tiles [2][]*ebiten.Image // summer, winter
	// units are the unit frames by side (0, 1, 2 neutral) and moved (0, 1):
	// six facings per unit type, clockwise from north.
	units [3][2][]*ebiten.Image
	on    bool
	err   string // why they could not be loaded
}

// The original tile is a 24×24 box filled by a hex with an 8-pixel top edge,
// so its hexes repeat every 32 pixels in a line and every 12 pixels down. It
// is scaled from these pitches onto the remake's (hexgrid.ColPitch,
// hexgrid.LinePitch), which lets the hexes interlock as in the original. A
// unit sprite fills the same box.
const (
	origHex                     = 24
	origColPitch, origLinePitch = 32, 12
)

// toggleOrigArt turns the original art on or off, loading it on first use,
// and tells the player.
func (g *game) toggleOrigArt() string {
	if g.dos == nil {
		return "original art: only under -rules dos"
	}
	if g.orig == nil {
		g.orig = &origArt{}
		if a, err := loadOrigArt(); err != nil {
			g.orig.err = err.Error()
		} else {
			g.orig.tiles, g.orig.units = a.tiles, a.units
		}
	}
	if g.orig.err != "" {
		return "original art: " + g.orig.err
	}
	g.orig.on = !g.orig.on
	return "original tiles and units " + map[bool]string{true: "on", false: "off"}[g.orig.on]
}

// origOn reports whether the original art is drawn.
func (g *game) origOn() bool { return g.orig != nil && g.orig.on && g.dos != nil }

// origGeoM places the original 24×24 box on the hex centred on screen point
// (sx, sy).
func (g *game) origGeoM(sx, sy float64) ebiten.GeoM {
	e := g.cam.Effective()
	var m ebiten.GeoM
	m.Translate(-origHex/2, -origHex/2)
	// One screen pixel more each way, so that rounding leaves no seam.
	m.Scale((e*hexgrid.ColPitch+1)/origColPitch, (e*hexgrid.LinePitch+1)/origLinePitch)
	m.Translate(sx, sy)
	return m
}

// drawOrigTiles draws the original tile of every hex in view, scaled onto the
// hex box. It reports false when the original art is off.
func (g *game) drawOrigTiles(dst *ebiten.Image, inView []hexgrid.Coord) bool {
	if !g.origOn() {
		return false
	}
	b := g.dos.s.Board
	season := 0
	if b.Winter {
		season = 1
	}
	tiles := g.orig.tiles[season]
	for _, c := range inView {
		cell, ok := b.Cell(c)
		if !ok || cell >= len(b.Terrain) || b.Terrain[cell] >= len(tiles) {
			continue
		}
		sx, sy := g.cam.WorldToScreen(hexgrid.Center(c))
		op := ebiten.DrawImageOptions{GeoM: g.origGeoM(sx, sy)}
		dst.DrawImage(tiles[b.Terrain[cell]], &op)
	}
	return true
}

// drawOrigUnit draws the original sprite of unit u facing face on the hex
// centred on (sx, sy), grey when moved. It reports false when the original
// art is off or has no sprite for u.
func (g *game) drawOrigUnit(dst *ebiten.Image, u dosplay.UnitView, sx, sy float64, face unitsprite.Facing, moved bool) bool {
	if !g.origOn() || u.Side < 0 || u.Side > 2 {
		return false
	}
	frames := g.orig.units[u.Side][map[bool]int{false: 0, true: 1}[moved]]
	// The remake's facings run anticlockwise from north, the frames clockwise.
	k := 6*u.Type + (int(unitsprite.Facings)-int(face))%int(unitsprite.Facings)
	if u.Type < 0 || face < 0 || k >= len(frames) {
		return false
	}
	op := ebiten.DrawImageOptions{GeoM: g.origGeoM(sx, sy)}
	dst.DrawImage(frames[k], &op)
	return true
}
