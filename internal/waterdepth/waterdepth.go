// Package waterdepth shades open water by the original's depths (water, the
// middle band waterDeep, the open sea waterSea) as one smooth field over the
// map, so the sea darkens without a step at any hex edge.
//
// A hex drawn as open water has its terrain's depth at its centre; a shore, a
// stone in water and land have none. Each hex vertex takes the mean depth of
// the hexes that meet there, and a point inside a hex is interpolated over the
// triangle of the centre and the two vertices around it, then smoothed. The
// same world point has the same depth through whichever tile draws it, and a
// coast is shallow, since every hex at a land vertex is land or a shore. A
// gentle world-periodic warp keeps the bands from tracing the hex outlines.
package waterdepth

import (
	"fmt"
	"image"
	"image/draw"
	"math"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/terraingrade"
)

// Of is the depth of terrain: 0.5 for the middle band waterDeep, 1 for the open
// sea waterSea, 0 for shallow water and anything else.
func Of(terrain string) float64 {
	switch terrain {
	case "waterDeep":
		return 0.5
	case "waterSea":
		return 1
	}
	return 0
}

// vertices are a hex's corners from its centre, in world pixels, clockwise
// from the top left (the 96 by 80 tile's hexagon, as mapmaterial's).
var vertices = [6][2]float64{{-23, -40}, {23, -40}, {48, 0}, {23, 40}, {-23, 40}, {-48, 0}}

// The warp: its amplitude in world pixels and its periods, the world
// material's (568 = 4 ColPitch, 400 = 10 LinePitch), so its phase at a hex
// centre takes few values.
const (
	warp                  = 8.0
	periodX, periodY      = 568.0, 400.0
	signatureRings        = 2
	digits                = "012"
	offMap           byte = 'x'
)

// Field is the depth of a level's water; it is immutable after New and safe
// for concurrent use.
type Field struct {
	grid   hexgrid.Grid
	wet    map[hexgrid.Coord]bool
	centre map[hexgrid.Coord]float64
	vertex map[hexgrid.Coord][6]float64
}

// New is the depth field of lvl as scene draws it: a hex is open water where
// its terrain is (mapcompose.OpenWater) and scene draws it with the water
// tile, not a shore.
func New(lvl *level.Level, scene *mapcompose.Scene) *Field {
	f := &Field{wet: map[hexgrid.Coord]bool{}, centre: map[hexgrid.Coord]float64{}, vertex: map[hexgrid.Coord][6]float64{}}
	if lvl == nil {
		return f
	}
	f.grid = lvl.Grid()
	type sum struct {
		depth float64
		n     int
	}
	at := map[image.Point]sum{}
	key := func(c hexgrid.Coord, k int) image.Point {
		x, y := hexgrid.Center(c)
		return image.Pt(int(math.Round(2*(x+vertices[k][0]))), int(math.Round(2*(y+vertices[k][1]))))
	}
	deep := false
	for line, row := range lvl.Lines {
		for col, h := range row {
			c := hexgrid.Coord{Line: line, Col: col}
			f.wet[c] = mapcompose.Wet(h.Terrain)
			d := Of(h.Terrain)
			if d > 0 && !openTile(scene, c) {
				d = 0
			}
			if d > 0 {
				f.centre[c] = d
				deep = true
			}
			for k := range 6 {
				s := at[key(c, k)]
				at[key(c, k)] = sum{s.depth + d, s.n + 1}
			}
		}
	}
	if !deep {
		return f
	}
	// The vertices of every hex on the map and of the ring beyond it, whose
	// corners the edge tiles reach.
	for line := -1; line <= f.grid.Lines; line++ {
		for col := -1; col <= f.grid.Width; col++ {
			c := hexgrid.Coord{Line: line, Col: col}
			var v [6]float64
			some := false
			for k := range 6 {
				if s := at[key(c, k)]; s.n > 0 {
					v[k] = s.depth / float64(s.n)
					some = some || v[k] > 0
				}
			}
			if !some {
				continue
			}
			f.vertex[c] = v
			if !f.grid.Contains(c) {
				// Beyond the map the sea goes on as its vertices have it.
				var m float64
				for _, x := range v {
					m = max(m, x)
				}
				f.centre[c] = m
			}
		}
	}
	return f
}

// openTile reports whether scene draws hex c with the open water tile.
func openTile(scene *mapcompose.Scene, c hexgrid.Coord) bool {
	if scene == nil {
		return true
	}
	if c.Line < 0 || c.Line >= len(scene.Cells) || c.Col < 0 || c.Col >= len(scene.Cells[c.Line]) {
		return false
	}
	return strings.TrimSuffix(scene.Cells[c.Line][c.Col].Background, "_w") == mapcompose.TileID("water")
}

// At is the depth at world point (x, y), 0 to 1.
func (f *Field) At(x, y float64) float64 {
	if f == nil || len(f.vertex) == 0 {
		return 0
	}
	return f.depthAt(x+warp*math.Sin(2*math.Pi*(x/periodX+2*y/periodY)), y+warp*math.Sin(2*math.Pi*(2*x/periodX-y/periodY)))
}

// depthAt is the depth at world point (x, y) without the warp.
func (f *Field) depthAt(x, y float64) float64 {
	c, ok := hexAt(x, y)
	if !ok {
		return 0
	}
	v, ok := f.vertex[c]
	if !ok {
		return f.centre[c]
	}
	cx, cy := hexgrid.Center(c)
	px, py := x-cx, y-cy
	for k := range 6 {
		a, b := vertices[k], vertices[(k+1)%6]
		// Barycentric coordinates of p in the triangle centre, a, b.
		det := a[0]*b[1] - b[0]*a[1]
		wa := (px*b[1] - b[0]*py) / det
		wb := (a[0]*py - px*a[1]) / det
		if wa < -1e-9 || wb < -1e-9 || wa+wb > 1+1e-9 {
			continue
		}
		t := (1-wa-wb)*f.centre[c] + wa*v[k] + wb*v[(k+1)%6]
		t = math.Max(0, math.Min(1, t))
		return t * t * (3 - 2*t)
	}
	return f.centre[c]
}

// hexAt is the hex whose hexagon holds world point (x, y), on the map or not.
func hexAt(x, y float64) (hexgrid.Coord, bool) {
	near, _ := hexgrid.Grid{}.PixelToHex(x, y)
	around := hexgrid.Around(near)
	for _, c := range append([]hexgrid.Coord{near}, around[:]...) {
		cx, cy := hexgrid.Center(c)
		dx, dy := math.Abs(x-cx), math.Abs(y-cy)
		if dy <= 40+1e-9 && dx <= 48-25*dy/40+1e-9 {
			return c, true
		}
	}
	return hexgrid.Coord{}, false
}

// The open sea's material repeats every 284 by 160 world pixels (2 ColPitch,
// 4 LinePitch), so a hex of the lattice meets it in one of eight phases, each
// a tile of its own: tile/sea-<kx><ky>, kx and ky the hex's top-left corner
// in steps of 71 and 40 world pixels.
func seaPhase(c hexgrid.Coord) string {
	kx := ((2*c.Col+(c.Line&1))%4 + 4) % 4
	ky := (c.Line%4 + 4) % 4
	return fmt.Sprintf("%d%d", kx, ky)
}

// SeaTile is the tile of the open sea's material for hex c whose background
// is background: its winter form (suffix _w) under a winter background.
func SeaTile(c hexgrid.Coord, background string) string {
	id := mapcompose.TileID("sea-" + seaPhase(c))
	if strings.HasSuffix(background, "_w") {
		id += "_w"
	}
	return id
}

// WaterTile is the open water tile of background's season, whose pixels the
// shores share where they are open water.
func WaterTile(background string) string {
	id := mapcompose.TileID("water")
	if strings.HasSuffix(background, "_w") {
		id += "_w"
	}
	return id
}

// seaFrom and seaFull are the depths over which the open sea's material
// replaces the shallows' water: the coast and the shallows keep the water of
// the shores, the middle band is mostly sea.
const seaFrom, seaFull = 0.15, 0.6

// Signature names everything a tile drawn at hex c needs to be shaded: the
// depths of the hexes within two rings, the sea's phase, and, where the depths
// differ, the warp's phase at c. It is "" where c is not water or no depth
// reaches its tile.
func (f *Field) Signature(c hexgrid.Coord) string {
	if f == nil || len(f.vertex) == 0 || !f.wet[c] {
		return ""
	}
	var b strings.Builder
	b.WriteByte('0' + byte(c.Line&1))
	seen := map[hexgrid.Coord]bool{}
	ring := []hexgrid.Coord{c}
	for range signatureRings + 1 {
		var next []hexgrid.Coord
		for _, h := range ring {
			if seen[h] {
				continue
			}
			seen[h] = true
			switch d, ok := f.centre[h]; {
			case !f.grid.Contains(h):
				if ok {
					b.WriteByte(digits[int(math.Round(2*d))])
				} else {
					b.WriteByte(offMap)
				}
			default:
				b.WriteByte(digits[int(math.Round(2*d))])
			}
			around := hexgrid.Around(h)
			next = append(next, around[:]...)
		}
		ring = next
	}
	s := b.String()
	depths := s[1:]
	if strings.Trim(depths, "0"+string(offMap)) == "" {
		return ""
	}
	s += "~" + seaPhase(c)
	if strings.Trim(depths, depths[:1]) == "" {
		return s
	}
	x, y := hexgrid.Center(c)
	return fmt.Sprintf("%s@%d,%d", s, int(math.Mod(x, periodX)), int(math.Mod(y, periodY)))
}

// Tiles are the graded tiles of the same hex and size that Shade draws with
// besides the background; any may be nil.
type Tiles struct {
	// Sea is the hex's SeaTile: the open sea's material, which replaces the
	// water from depth seaFrom to seaFull.
	Sea image.Image
	// Water is the season's WaterTile, whose pixels a shore shares where it
	// is open water.
	Water image.Image
	// Surf is the season's SurfTile, the foam laid along coasts by the sea.
	Surf image.Image
}

// Shade is tile, a graded background whose top left lies at origin in pixels
// of density (world pixels times density), with its water darkened by depth,
// the open sea's material blended in and surf along a coast by the sea, as
// with has them.
func (f *Field) Shade(tile image.Image, with Tiles, origin image.Point, density int) *image.NRGBA {
	b := tile.Bounds()
	out := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(out, out.Rect, tile, b.Min, draw.Src)
	d := float64(density)
	depth := make([]float64, b.Dx()*b.Dy())
	for y := range b.Dy() {
		for x := range b.Dx() {
			depth[y*b.Dx()+x] = f.At((float64(origin.X+x)+0.5)/d, (float64(origin.Y+y)+0.5)/d)
		}
	}
	water := sized(with.Water, b.Size())
	land := landOf(out, water)
	if sea := sized(with.Sea, b.Size()); sea != nil {
		terraingrade.Blend(out, sea, water, func(x, y int) float64 {
			t := math.Max(0, math.Min(1, (depth[y*b.Dx()+x]-seaFrom)/(seaFull-seaFrom)))
			return t * t * (3 - 2*t)
		})
	}
	terraingrade.Deepen(out, func(x, y int) float64 { return depth[y*b.Dx()+x] })
	if surf := sized(with.Surf, b.Size()); surf != nil && land != nil {
		f.surf(out, land, surf, origin, density)
	}
	return out
}

// sized is im as an NRGBA at the origin when it has size, else nil.
func sized(im image.Image, size image.Point) *image.NRGBA {
	if im == nil || im.Bounds().Size() != size {
		return nil
	}
	out := image.NewNRGBA(image.Rect(0, 0, size.X, size.Y))
	draw.Draw(out, out.Rect, im, im.Bounds().Min, draw.Src)
	return out
}

// SurfTile is the surf of background's season (Tiles.Surf).
func SurfTile(background string) string {
	id := mapcompose.TileID("surf")
	if strings.HasSuffix(background, "_w") {
		id += "_w"
	}
	return id
}
