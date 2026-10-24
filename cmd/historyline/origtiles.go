package main

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// origTiles are the original game's terrain tiles, summer and winter, by
// terrain record. F8 draws them in place of the remake's terrain under
// -rules dos, for comparison on a local machine: only builds with
// -tags origtiles (just play-original) can load them, from the git-ignored
// original/ folder; other builds answer F8 with a message.
type origTiles struct {
	tiles [2][]*ebiten.Image // summer, winter
	on    bool
	err   string // why they could not be loaded
}

// The original tile is a 24×24 box filled by a hex with an 8-pixel top edge,
// so its hexes repeat every 32 pixels in a line and every 12 pixels down. It
// is scaled from these pitches onto the remake's (hexgrid.ColPitch,
// hexgrid.LinePitch), which lets the hexes interlock as in the original.
const (
	origHex                     = 24
	origColPitch, origLinePitch = 32, 12
)

// toggleOrigTiles turns the original tiles on or off, loading them on first
// use, and tells the player.
func (g *game) toggleOrigTiles() string {
	if g.dos == nil {
		return "original tiles: only under -rules dos"
	}
	if g.orig == nil {
		g.orig = &origTiles{}
		if t, err := loadOrigTiles(); err != nil {
			g.orig.err = err.Error()
		} else {
			g.orig.tiles = t.tiles
		}
	}
	if g.orig.err != "" {
		return "original tiles: " + g.orig.err
	}
	g.orig.on = !g.orig.on
	return "original tiles " + map[bool]string{true: "on", false: "off"}[g.orig.on]
}

// drawOrigTiles draws the original tile of every hex in view, scaled onto the
// hex box. It reports false when the original tiles are off.
func (g *game) drawOrigTiles(dst *ebiten.Image, inView []hexgrid.Coord) bool {
	if g.orig == nil || !g.orig.on || g.dos == nil {
		return false
	}
	b := g.dos.s.Board
	season := 0
	if b.Winter {
		season = 1
	}
	tiles := g.orig.tiles[season]
	e := g.cam.Effective()
	for _, c := range inView {
		cell, ok := b.Cell(c)
		if !ok || cell >= len(b.Terrain) || b.Terrain[cell] >= len(tiles) {
			continue
		}
		sx, sy := g.cam.WorldToScreen(hexgrid.Center(c))
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(-origHex/2, -origHex/2)
		// One screen pixel more each way, so that rounding leaves no seam.
		op.GeoM.Scale((e*hexgrid.ColPitch+1)/origColPitch, (e*hexgrid.LinePitch+1)/origLinePitch)
		op.GeoM.Translate(sx, sy)
		dst.DrawImage(tiles[b.Terrain[cell]], &op)
	}
	return true
}
