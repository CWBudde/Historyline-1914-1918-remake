package main

import (
	"image"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/hajimehoshi/ebiten/v2"
)

// The crater fields' backgrounds and the border of a winter crater stamp with
// rough ground are reworked at draw time (mapcompose.FieldMaterial,
// mapcompose.StampFringe): a tile is made once per world phase and kept while it
// fits the budget, oldest out first. The tile map only: a painted map's sapper
// hexes keep the tile as it is.

const fieldBudget = 96 * 1024 * 1024

type fieldKey struct {
	id      string
	phase   image.Point
	edges   [6]bool
	at      hexgrid.Coord // of a fringe, which is not periodic; zero in a field
	density int
	bare    bool
}

type fieldTiles struct {
	materials map[fieldMaterialKey]*mapcompose.FieldMaterial
	tiles     map[fieldKey]*ebiten.Image
	order     []fieldKey
	bytes     int
}

type fieldMaterialKey struct {
	exemplar string
	density  int
}

// release frees the reworked tiles.
func (t *fieldTiles) release(a *terrainArt) {
	for _, img := range t.tiles {
		a.forget(img)
		img.Deallocate()
	}
	t.tiles, t.order, t.bytes = map[fieldKey]*ebiten.Image{}, nil, 0
}

// reworksGround reports whether background id of hex c in scene s is reworked.
func reworksGround(s *mapcompose.Scene, c hexgrid.Coord, id string) bool {
	return s.FieldAt(c, id) != nil || s.FringeAt(c, id) != nil
}

// reworkedPixels is background tile, id of entry e on hex c of scene s at density
// d, reworked if it is; tile itself otherwise.
func (a *terrainArt) reworkedPixels(s *mapcompose.Scene, c hexgrid.Coord, id string, tile image.Image, e assets.Entry, d int) (image.Image, error) {
	origin := materialOrigin(c, e, d)
	if f := s.FringeAt(c, id); f != nil {
		exemplar, err := a.packs[packKey{a.group[f.Exemplar], d}].Image(f.Exemplar, d)
		if err != nil {
			return nil, err
		}
		return f.Rework(tile, exemplar, origin, c, d), nil
	}
	f := s.FieldAt(c, id)
	if f == nil {
		return tile, nil
	}
	if a.fields == nil {
		a.fields = &fieldTiles{tiles: map[fieldKey]*ebiten.Image{}}
	}
	if a.fields.materials == nil {
		a.fields.materials = map[fieldMaterialKey]*mapcompose.FieldMaterial{}
	}
	k := fieldMaterialKey{f.Exemplar, d}
	m := a.fields.materials[k]
	if m == nil {
		pack := a.packs[packKey{a.group[f.Exemplar], d}]
		exemplar, err := pack.Image(f.Exemplar, d)
		if err != nil {
			return nil, err
		}
		entry, err := pack.Entry(f.Exemplar, d)
		if err != nil {
			return nil, err
		}
		m = mapcompose.NewFieldMaterial(exemplar, entry.Anchor, d)
		a.fields.materials[k] = m
	}
	return m.Rework(tile, origin, c, f.Edges), nil
}

// fieldImage is background id of hex c at density d reworked, as image or
// bareImage gives it as it is.
func (a *terrainArt) fieldImage(c hexgrid.Coord, id string, d int, bare bool) (*ebiten.Image, assets.Entry, error) {
	e, err := a.entry(a.group[id], id, d)
	if err != nil {
		return nil, e, err
	}
	if a.fields == nil {
		a.fields = &fieldTiles{tiles: map[fieldKey]*ebiten.Image{}}
	}
	t := a.fields
	key := fieldKey{id: id, density: d, bare: bare}
	if f := a.scene.FringeAt(c, id); f != nil {
		key.at = c
	} else {
		f := a.scene.FieldAt(c, id)
		o := materialOrigin(c, e, d)
		key.phase = mapcompose.FieldPhase(o, d)
		key.edges = f.Edges
	}
	if img := t.tiles[key]; img != nil {
		return img, e, nil
	}
	tile, err := a.packs[packKey{a.group[id], d}].Image(id, d)
	if err != nil {
		return nil, e, err
	}
	reworked, err := a.reworkedPixels(&a.scene, c, id, tile, e, d)
	if err != nil {
		return nil, e, err
	}
	w, h := reworked.Bounds().Dx()+2*tilePad, reworked.Bounds().Dy()+2*tilePad
	img := ebiten.NewImage(w, h)
	img.WritePixels(padded(reworked, tilePad, !bare))
	for t.bytes+w*h*4 > fieldBudget && len(t.order) > 0 {
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
