package main

import (
	"image"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/seasonblend"
	"github.com/hajimehoshi/ebiten/v2"
)

// Near a season border (seasonblend) a hex's background, overlays and
// standing items are each drawn as one texture mixed from the hex's layer in
// every season there, made once per layer and kept while it fits the budget,
// oldest out first.

const blendBudget = 96 * 1024 * 1024

type blendKey struct {
	at      hexgrid.Coord
	ids     string // the layer's IDs in each season
	density int
	bare    bool
}

type blendTile struct {
	img *ebiten.Image
	e   assets.Entry
}

type blendTiles struct {
	blend *seasonblend.Blend
	tiles map[blendKey]blendTile
	order []blendKey
	bytes int
}

// release frees the mixed textures.
func (t *blendTiles) release(a *terrainArt) {
	for _, tile := range t.tiles {
		a.forget(tile.img)
		tile.img.Deallocate()
	}
	t.tiles, t.order, t.bytes = map[blendKey]blendTile{}, nil, 0
}

// knownClusters reports whether a cluster is known to the player: none of
// its hexes is unseen.
func knownClusters(fog func(hexgrid.Coord) state.Fog) func(*mapcompose.Cluster) bool {
	return func(cl *mapcompose.Cluster) bool {
		return !slices.ContainsFunc(cl.Cells, func(x hexgrid.Coord) bool { return fog(x) == state.Unseen })
	}
}

// blendSeasons are the seasons mixed on hex c under the player's fog; nil
// where c is drawn as composed.
func (g *game) blendSeasons(c hexgrid.Coord) []mapcompose.Season {
	if g.terrain.blend == nil || g.fogOf(c) == state.Unseen {
		return nil
	}
	return g.terrain.blend.blend.Seasons(c)
}

// drawBlended draws layer pass (0 the background, 2 the overlays, 4 the
// standing items) of hex c mixed across its seasons, with its anchor on
// screen point (sx, sy), and reports whether c is mixed.
func (g *game) drawBlended(dst *ebiten.Image, c hexgrid.Coord, pass int, sx, sy float64, bare bool) bool {
	if g.blendSeasons(c) == nil {
		return false
	}
	d := terrainDensity(g.cam.Effective())
	img, e, err := g.terrain.blendImage(c, pass, d, bare, knownClusters(g.fogOf))
	if err != nil {
		g.terrainErr = err.Error()
		return true
	}
	if img == nil {
		return true
	}
	if g.fogOf(c) == state.Seen {
		img = g.terrain.dimmed(img)
	}
	var op ebiten.DrawImageOptions
	op.GeoM = anchorGeoM(e, d, g.terrain.pad, g.cam.Effective(), sx, sy)
	op.Filter = ebiten.FilterLinear
	dst.DrawImage(img, &op)
	return true
}

// blendImage is layer pass of hex c at density d mixed across its seasons,
// with an entry whose anchor is the hex centre; nil where the layer draws
// nothing.
func (a *terrainArt) blendImage(c hexgrid.Coord, pass, d int, bare bool, known func(*mapcompose.Cluster) bool) (*ebiten.Image, assets.Entry, error) {
	t := a.blend
	seasons := t.blend.Seasons(c)
	ids := make([][]string, len(seasons))
	var key strings.Builder
	for i, s := range seasons {
		bg, _, overlays, standing := t.blend.Scenes[s].Layers(c, known)
		ids[i] = [5][]string{{bg}, nil, overlays, nil, standing}[pass]
		key.WriteString(strings.Join(ids[i], ","))
		key.WriteByte(';')
	}
	k := blendKey{c, key.String(), d, bare}
	if tile, ok := t.tiles[k]; ok {
		return tile.img, tile.e, nil
	}
	layers := make([][]seasonblend.Part, len(seasons))
	for i := range seasons {
		for _, id := range ids[i] {
			if id == "" {
				continue
			}
			im, e, err := a.pixels(c, id, d, pass == 0)
			if err != nil {
				return nil, assets.Entry{}, err
			}
			layers[i] = append(layers[i], seasonblend.Part{Image: im, At: materialOrigin(c, e, d)})
		}
	}
	mixed := t.blend.Mix(c, layers, d)
	if mixed == nil {
		t.tiles[k] = blendTile{}
		return nil, assets.Entry{}, nil
	}
	x, y := hexgrid.Center(c)
	o := mixed.Bounds().Min
	e := assets.Entry{Anchor: [2]float64{x - float64(o.X)/float64(d), y - float64(o.Y)/float64(d)}}
	w, h := mixed.Bounds().Dx()+2*tilePad, mixed.Bounds().Dy()+2*tilePad
	img := ebiten.NewImage(w, h)
	img.WritePixels(padded(mixed, tilePad, pass == 0 && !bare))
	for t.bytes+w*h*4 > blendBudget && len(t.order) > 0 {
		old := t.order[0]
		t.order = t.order[1:]
		if previous := t.tiles[old].img; previous != nil {
			t.bytes -= previous.Bounds().Dx() * previous.Bounds().Dy() * 4
			a.forget(previous)
			previous.Deallocate()
		}
		delete(t.tiles, old)
	}
	t.tiles[k] = blendTile{img, e}
	t.order = append(t.order, k)
	t.bytes += w * h * 4
	return img, e, nil
}

// pixels is drawn ID id of hex c at density d as painted, a background as
// drawn: water shaded by depth, a summer meadow sampled from the world
// material.
func (a *terrainArt) pixels(c hexgrid.Coord, id string, d int, background bool) (image.Image, assets.Entry, error) {
	e, err := a.entry(a.group[id], id, d)
	if err != nil {
		return nil, e, err
	}
	tile, err := a.packs[packKey{a.group[id], d}].Image(id, d)
	if err != nil || !background {
		return tile, e, err
	}
	if a.depthSignature(c, id) != "" {
		shaded, err := a.shadeDepth(c, id, tile, e, d)
		return shaded, e, err
	}
	if m := a.materials; m.eligible(c, id) {
		sampled, err := sampleMaterial(m.samplers[d], m.field, tile, materialOrigin(c, e, d))
		return sampled, e, err
	}
	return tile, e, nil
}
