package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg" // the scenario files' backdrops
	_ "image/png"
	"io/fs"
	"math"
	"path/filepath"
	"strings"

	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// A backdrop (-backdrop) is one painted picture of a level's whole ground,
// drawn in place of the terrain tiles. It covers the level's world rectangle as
// maprender draws it (maprender.Size, origin at world (0, 0)) at any resolution
// of the same aspect. Fog, units and every overlay stay as they are; no grid
// outlines the hexes.

const (
	// backdropTileSize is the largest side of a backdrop texture, padding
	// included, well below every GPU's texture limit.
	backdropTileSize = 2048
	// backdropPad is the border each texture repeats of its neighbours, so
	// linear filtering and the smaller mipmap levels at a tile's edge sample the
	// true neighbouring pixels.
	backdropPad = 8
	// backdropAspect is how far the picture's aspect may stray from the world's.
	backdropAspect = 0.01
)

// backdropImage is a loaded backdrop picture, split into textures.
type backdropImage struct {
	w, h  int // the picture in pixels
	tiles []backdropTile
}

// backdropTile is one texture of a backdrop: the pixels it owns (core) and
// what it holds, the core with its padding (src), both in picture pixels.
type backdropTile struct {
	core, src image.Rectangle
	img       *ebiten.Image
}

// backdrop is a backdrop picture placed on a level's world.
type backdrop struct {
	*backdropImage
	scale  float64 // picture pixels per world pixel
	ww, wh float64 // the world it covers
}

// backdropScale is the picture pixels per world pixel of a w×h picture
// covering a world of ww×wh; the aspects must agree within backdropAspect.
func backdropScale(w, h int, ww, wh float64) (float64, error) {
	if w <= 0 || h <= 0 || ww <= 0 || wh <= 0 {
		return 0, fmt.Errorf("backdrop %dx%d on a %gx%g world: empty", w, h, ww, wh)
	}
	sx, sy := float64(w)/ww, float64(h)/wh
	if math.Abs(sx/sy-1) > backdropAspect {
		return 0, fmt.Errorf("backdrop %dx%d does not fit the level's %gx%g world: aspect %.4f, want %.4f within %g %%",
			w, h, ww, wh, float64(w)/float64(h), ww/wh, backdropAspect*100)
	}
	return sx, nil
}

// backdropTiles splits a w×h picture into cores of at most size-2*pad pixels a
// side, which partition it, each held with up to pad pixels of its neighbours.
func backdropTiles(w, h, size, pad int) []backdropTile {
	step := size - 2*pad
	all := image.Rect(0, 0, w, h)
	var out []backdropTile
	for y := 0; y < h; y += step {
		for x := 0; x < w; x += step {
			core := image.Rect(x, y, min(x+step, w), min(y+step, h))
			out = append(out, backdropTile{core: core, src: core.Inset(-pad).Intersect(all)})
		}
	}
	return out
}

// loadBackdrop reads the PNG at path (a file, or a file in an art directory
// URL with its files.txt) and makes its textures.
func loadBackdrop(path string) (*backdropImage, error) {
	dir, name := ".", path
	if i := strings.LastIndexAny(path, "/"+string(filepath.Separator)); i >= 0 {
		dir, name = path[:i+1], path[i+1:]
	}
	fsys, done, err := openArtDir(dir, nil)
	if err != nil {
		return nil, fmt.Errorf("backdrop: %w", err)
	}
	defer done()
	b, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("backdrop: %w", err)
	}
	return decodeBackdrop(path, b)
}

// levelBackdrop loads a scenario level's backdrop (level.Extra.Backdrop, a
// slash path below the levels directory as scenarioLevel sets it) from dir
// where present, else the embedded copy.
func levelBackdrop(dir, file string) (*backdropImage, error) {
	for prefix, emb := range scenarioDirs {
		if strings.HasPrefix(file, prefix) {
			b, err := readLevelFile(dir, emb, file)
			if err != nil {
				return nil, fmt.Errorf("backdrop: %w", err)
			}
			return decodeBackdrop(file, b)
		}
	}
	return nil, fmt.Errorf("backdrop %s: not in a scenario directory", file)
}

// decodeBackdrop decodes a PNG or JPEG picture and splits it into textures.
func decodeBackdrop(name string, b []byte) (*backdropImage, error) {
	pic, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("backdrop %s: %w", name, err)
	}
	return newBackdropImage(pic)
}

// newBackdropImage splits pic into textures.
func newBackdropImage(pic image.Image) (*backdropImage, error) {
	sub, ok := pic.(interface {
		SubImage(image.Rectangle) image.Image
	})
	if !ok {
		return nil, fmt.Errorf("backdrop: cannot split a %T", pic)
	}
	r := pic.Bounds()
	bi := &backdropImage{w: r.Dx(), h: r.Dy(), tiles: backdropTiles(r.Dx(), r.Dy(), backdropTileSize, backdropPad)}
	for i := range bi.tiles {
		t := &bi.tiles[i]
		t.img = ebiten.NewImageFromImage(sub.SubImage(t.src.Add(r.Min)))
	}
	return bi, nil
}

// on places the picture on the world of grid, refusing one whose aspect differs.
func (bi *backdropImage) on(grid hexgrid.Grid) (*backdrop, error) {
	ww, wh := maprender.Size(grid)
	s, err := backdropScale(bi.w, bi.h, float64(ww), float64(wh))
	if err != nil {
		return nil, err
	}
	return &backdrop{backdropImage: bi, scale: s, ww: float64(ww), wh: float64(wh)}, nil
}

// mapGround is the screen around the map.
var mapGround = color.RGBA{20, 20, 24, 255}

// drawBackdrop draws the backdrop in the world rectangle (vx, vy, vw, vh), the
// hexes the sappers changed over it from the terrain catalog, the fog of the
// hexes in view: an unseen hex black, a seen one in the remembered grey. No
// hex grid is drawn: the hexes show only through the game's own overlays.
func (g *game) drawBackdrop(dst *ebiten.Image, vx, vy, vw, vh float64, inView []hexgrid.Coord) {
	var seen, unseen []hexgrid.Coord
	for _, c := range inView {
		switch g.fogOf(c) {
		case state.Seen:
			seen = append(seen, c)
		case state.Unseen:
			unseen = append(unseen, c)
		}
	}
	// A seen hex shows the picture greyed, taken from a copy of the view.
	target, at := dst, dst.Bounds().Min
	if len(seen) > 0 {
		size := dst.Bounds().Size()
		if s := g.backdropView; s == nil || s.Bounds().Size() != size {
			if s != nil {
				s.Deallocate()
			}
			g.backdropView = ebiten.NewImage(size.X, size.Y)
		}
		target = g.backdropView
		target.Clear()
	} else {
		at = image.Point{}
	}
	b, e := g.backdrop, g.cam.Effective()
	ox, oy := g.cam.WorldToScreen(0, 0)
	for _, t := range b.tiles {
		x0, y0 := float64(t.core.Min.X)/b.scale, float64(t.core.Min.Y)/b.scale
		x1, y1 := float64(t.core.Max.X)/b.scale, float64(t.core.Max.Y)/b.scale
		if x1 < vx || x0 > vx+vw || y1 < vy || y0 > vy+vh {
			continue
		}
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(float64(t.src.Min.X), float64(t.src.Min.Y))
		op.GeoM.Scale(e/b.scale, e/b.scale)
		op.GeoM.Translate(ox-float64(at.X), oy-float64(at.Y))
		op.Filter = ebiten.FilterLinear
		target.DrawImage(t.img, &op)
	}
	if target != dst {
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(float64(at.X), float64(at.Y))
		dst.DrawImage(target, &op)
		var vs []ebiten.Vertex
		var is []uint16
		for _, c := range seen {
			n := uint16(len(vs))
			for _, v := range hexgrid.Corners(c) {
				x, y := g.cam.WorldToScreen(v[0], v[1])
				vs = append(vs, ebiten.Vertex{
					DstX: float32(x), DstY: float32(y), SrcX: float32(x) - float32(at.X), SrcY: float32(y) - float32(at.Y),
					ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1,
				})
			}
			for i := uint16(1); i < 5; i++ {
				is = append(is, n, n+i, n+i+1)
			}
		}
		colorm.DrawTriangles(dst, vs, is, target, remembered, &colorm.DrawTrianglesOptions{})
	}
	g.drawReworked(dst, inView)
	if len(unseen) > 0 {
		var p vector.Path
		g.hexPath(&p, unseen)
		vector.FillPath(dst, &p, nil, pathOptions(fogShade(state.Unseen, color.RGBA{})))
	}
	if g.split != nil && g.split.frame.Slim {
		return // the slim frame shows the whole picture, edge to edge
	}
	// Elsewhere the picture beyond the outer hexes stays empty, as the terrain
	// tiles leave it.
	var edge vector.Path
	x0, y0 := g.cam.WorldToScreen(0, 0)
	x1, y1 := g.cam.WorldToScreen(b.ww, b.wh)
	edge.MoveTo(float32(x0), float32(y0))
	edge.LineTo(float32(x1), float32(y0))
	edge.LineTo(float32(x1), float32(y1))
	edge.LineTo(float32(x0), float32(y1))
	edge.Close()
	play := g.playGrid()
	var all []hexgrid.Coord
	for l, line := range g.lvl.Lines {
		for c := range line {
			if at := (hexgrid.Coord{Line: l, Col: c}); play.Contains(at) {
				all = append(all, at)
			}
		}
	}
	g.hexPath(&edge, all)
	vector.FillPath(dst, &edge, &vector.FillOptions{FillRule: vector.FillRuleEvenOdd}, pathOptions(mapGround))
}

// hexPath adds the outline of each of cells on screen to p.
func (g *game) hexPath(p *vector.Path, cells []hexgrid.Coord) {
	for _, c := range cells {
		for i, v := range hexgrid.Corners(c) {
			x, y := g.cam.WorldToScreen(v[0], v[1])
			if i == 0 {
				p.MoveTo(float32(x), float32(y))
			} else {
				p.LineTo(float32(x), float32(y))
			}
		}
		p.Close()
	}
}

// drawReworked draws the hexes in view the sappers changed (-rules dos), which
// the picture cannot show: their ground and overlays from the terrain catalog,
// or a coloured hex without one.
func (g *game) drawReworked(dst *ebiten.Image, inView []hexgrid.Coord) {
	for _, c := range inView {
		if !g.reworkedHex[c] {
			continue
		}
		sx, sy := g.cam.WorldToScreen(hexgrid.Center(c))
		if g.terrain == nil {
			hx := g.lvl.Lines[c.Line][c.Col]
			g.drawHex(dst, float32(sx), float32(sy), fogShade(g.fogOf(c), g.fill(hx.Terrain)), false)
			continue
		}
		bg, _, overlays, _, dim := terrainLayers(&g.terrain.scene, c, g.fogOf)
		g.drawTerrainTile(dst, bg, sx, sy, dim, g.materialHex(c), false)
		for _, id := range overlays {
			g.drawTerrainTile(dst, id, sx, sy, dim, nil, true)
		}
	}
}
