package main

import (
	"image"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/waterdepth"
	"github.com/hajimehoshi/ebiten/v2"
)

// The water backgrounds are shaded by the sea's depth (waterdepth): a tile is
// shaded once per depth signature and kept while it fits the budget, oldest
// out first.

const depthBudget = 96 * 1024 * 1024

type depthKey struct {
	signature, id string
	density       int
	bare          bool
}

type depthTiles struct {
	field *waterdepth.Field
	tiles map[depthKey]*ebiten.Image
	order []depthKey
	bytes int
}

// release frees the shaded tiles.
func (t *depthTiles) release(a *terrainArt) {
	for _, img := range t.tiles {
		a.forget(img)
		img.Deallocate()
	}
	t.tiles, t.order, t.bytes = map[depthKey]*ebiten.Image{}, nil, 0
}

// depthSignature is the depth signature of background id on hex c; "" where it
// is drawn unshaded.
func (a *terrainArt) depthSignature(c hexgrid.Coord, id string) string {
	if a.depth == nil || id == "" || id == "tile/unseen" {
		return ""
	}
	return a.depth.field.Signature(c)
}

// depthImage is background id of hex c at density d shaded by depth, with the
// open sea's material and surf where the catalog has them, as image or
// bareImage gives it unshaded.
func (a *terrainArt) depthImage(c hexgrid.Coord, signature, id string, d int, bare bool) (*ebiten.Image, assets.Entry, error) {
	e, err := a.entry(a.group[id], id, d)
	if err != nil {
		return nil, e, err
	}
	t := a.depth
	key := depthKey{signature, id, d, bare}
	if img := t.tiles[key]; img != nil {
		return img, e, nil
	}
	tile, err := a.packs[packKey{a.group[id], d}].Image(id, d)
	if err != nil {
		return nil, e, err
	}
	var with waterdepth.Tiles
	for _, w := range []struct {
		id  string
		set *image.Image
	}{{waterdepth.SeaTile(c, id), &with.Sea}, {waterdepth.WaterTile(id), &with.Water}, {waterdepth.SurfTile(id), &with.Surf}} {
		if a.group[w.id] == "" {
			continue
		}
		if *w.set, err = a.packs[packKey{a.group[w.id], d}].Image(w.id, d); err != nil {
			return nil, e, err
		}
	}
	shaded := t.field.Shade(tile, with, materialOrigin(c, e, d), d)
	w, h := shaded.Bounds().Dx()+2*tilePad, shaded.Bounds().Dy()+2*tilePad
	img := ebiten.NewImage(w, h)
	img.WritePixels(padded(shaded, tilePad, !bare))
	for t.bytes+w*h*4 > depthBudget && len(t.order) > 0 {
		old := t.order[0]
		t.order = t.order[1:]
		previous := t.tiles[old]
		t.bytes -= previous.Bounds().Dx() * previous.Bounds().Dy() * 4
		a.forget(previous)
		previous.Deallocate()
		delete(t.tiles, old)
	}
	t.tiles[key] = img
	t.order = append(t.order, key)
	t.bytes += w * h * 4
	return img, e, nil
}
