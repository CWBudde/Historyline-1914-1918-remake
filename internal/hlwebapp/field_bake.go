package hlwebapp

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"strconv"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hlasset"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
)

type fieldBakeKey struct {
	id      string
	phase   image.Point
	edges   [6]bool
	at      hexgrid.Coord // of a fringe, which is not periodic; zero in a field
	density int
}

type fieldMaterialKey struct {
	exemplar string
	density  int
}

// reworked is background im, id with its entry's origin on hex c of scene s at
// density d, as the tile map draws it; im itself where it is not reworked.
func (w *waterBaker) reworked(s *mapcompose.Scene, c hexgrid.Coord, id string, im image.Image, origin image.Point, d int) (image.Image, error) {
	if f := s.FringeAt(c, id); f != nil {
		exemplar, _, err := w.image(f.Exemplar, d)
		if err != nil {
			return nil, err
		}
		return f.Rework(im, exemplar, origin, c, d), nil
	}
	f := s.FieldAt(c, id)
	if f == nil {
		return im, nil
	}
	k := fieldMaterialKey{f.Exemplar, d}
	m := w.fields[k]
	if m == nil {
		exemplar, e, err := w.image(f.Exemplar, d)
		if err != nil {
			return nil, err
		}
		m = mapcompose.NewFieldMaterial(exemplar, e.Anchor, d)
		if w.fields == nil {
			w.fields = map[fieldMaterialKey]*mapcompose.FieldMaterial{}
		}
		w.fields[k] = m
	}
	return m.Rework(im, origin, c, f.Edges), nil
}

// bakeFields adds to the water's baked backgrounds the crater fields and stamp
// borders, marked as the tile map's alone. Nearly every hex is its own tile, so
// they hold the lite encoding only, about 100 MB over the 52 months.
func (w *waterBaker) bakeFields(ctx context.Context, lv *level.Level, layers *BakedLayers) error {
	scene := layers.Scene
	all := func(*mapcompose.Cluster) bool { return true }
	type cell struct {
		at  hexgrid.Coord
		key string
		id  string
	}
	var cells []cell
	for line, row := range lv.Lines {
		for col := range row {
			if err := ctx.Err(); err != nil {
				return err
			}
			c := hexgrid.Coord{Line: line, Col: col}
			id, _, _, _ := scene.Layers(c, all)
			if id == "" || (scene.FieldAt(c, id) == nil && scene.FringeAt(c, id) == nil) {
				continue
			}
			name := fmt.Sprintf("%d,%d", line, col)
			layers.Backgrounds[name] = map[string]BakedBackground{}
			cells = append(cells, cell{c, name, id})
		}
	}
	for _, d := range []int{1, 2, 4} {
		for _, c := range cells {
			if err := ctx.Err(); err != nil {
				return err
			}
			origin := waterOrigin(c.at, w.anchors[waterAssetKey{c.id, d}], d)
			key := fieldBakeKey{id: c.id, density: d}
			if f := scene.FieldAt(c.at, c.id); f != nil {
				key.phase, key.edges = mapcompose.FieldPhase(origin, d), f.Edges
			} else {
				key.at = c.at
			}
			result, ok := w.fieldCache[key]
			if !ok {
				var err error
				result, err = w.bakeField(scene, c.at, c.id, d)
				if err != nil {
					return fmt.Errorf("field bake %s %s at %dx: %w", c.key, c.id, d, err)
				}
				if w.fieldCache == nil {
					w.fieldCache = map[fieldBakeKey]BakedBackground{}
				}
				w.fieldCache[key] = result
			}
			layers.Backgrounds[c.key][strconv.Itoa(d)] = result
		}
	}
	return nil
}

func (w *waterBaker) bakeField(s *mapcompose.Scene, c hexgrid.Coord, id string, d int) (BakedBackground, error) {
	tile, e, err := w.image(id, d)
	if err != nil {
		return BakedBackground{}, err
	}
	im, err := w.reworked(s, c, id, tile, waterOrigin(c, e.Anchor, d), d)
	if err != nil {
		return BakedBackground{}, err
	}
	var raw bytes.Buffer
	if err := png.Encode(&raw, im); err != nil {
		return BakedBackground{}, err
	}
	result := BakedBackground{ID: id, Width: e.Width, Height: e.Height, Density: d, Anchor: e.Anchor, TileOnly: true}
	compressed, err := cachedLite(im, raw.Bytes(), hlasset.FieldWebP)
	if err != nil {
		return result, err
	}
	if len(compressed.Data) < raw.Len() {
		result.Lite, err = w.blobs.put(compressed.Ext, compressed.Data)
	} else {
		result.Lite, err = w.blobs.fullArt("png", im, raw.Bytes())
	}
	result.Full = result.Lite
	return result, err
}
