package hlwebapp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"math"
	"os"
	"path"
	"strconv"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/waterdepth"
)

// BakedBackground replaces only the matching Scene.Layers background. Geometry
// belongs to its original native pack entry, including the logical anchor.
type BakedBackground struct {
	ID      string     `json:"id"`
	Width   int        `json:"width"`
	Height  int        `json:"height"`
	Density int        `json:"density"`
	Anchor  [2]float64 `json:"logicalAnchor"`
	Lite    FileRef    `json:"lite"`
	Full    FileRef    `json:"full"`
}

type waterBakeKey struct {
	signature, id string
	density       int
	origin        image.Point
}
type waterAssetKey struct {
	id      string
	density int
}

// Only manifests stay resident across levels. Native Pack lazily decodes its
// verified snapshot, bounded at assets.MaxDecodedBytes; retain at most four
// packs for the background and optional sea/water/surf inputs rather than the
// full terrain catalog. Baked caches contain references, never decoded pixels.
type waterBaker struct {
	root    *os.Root
	blobs   blobWriter
	dirs    map[waterAssetKey]string
	anchors map[waterAssetKey][2]float64
	packs   map[string]*assets.Pack
	recent  []string
	cache   map[waterBakeKey]BakedBackground
}

func newWaterBaker(dir string, b blobWriter) (*waterBaker, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	w := &waterBaker{root: root, blobs: b, dirs: map[waterAssetKey]string{}, anchors: map[waterAssetKey][2]float64{}, packs: map[string]*assets.Pack{}, cache: map[waterBakeKey]BakedBackground{}}
	err = fs.WalkDir(root.FS(), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "manifest.json" {
			return nil
		}
		raw, err := fs.ReadFile(root.FS(), p)
		if err != nil {
			return err
		}
		m, err := assets.ParseManifest(raw)
		if err != nil {
			return err
		}
		for _, e := range m.Entries {
			key := waterAssetKey{e.ID, e.Density}
			if _, ok := w.dirs[key]; ok {
				return fmt.Errorf("duplicate water source %s at %dx", e.ID, e.Density)
			}
			w.dirs[key] = path.Dir(p)
			w.anchors[key] = e.Anchor
		}
		return nil
	})
	if err != nil {
		return nil, errors.Join(err, root.Close())
	}
	return w, nil
}

func (w *waterBaker) close() error { return w.root.Close() }

func (w *waterBaker) image(id string, d int) (image.Image, assets.Entry, error) {
	dir, ok := w.dirs[waterAssetKey{id, d}]
	if !ok {
		return nil, assets.Entry{}, fmt.Errorf("water bake: missing background %s at %dx", id, d)
	}
	pack := w.packs[dir]
	if pack == nil {
		// Drop the oldest pack before opening another bounded snapshot.
		if len(w.recent) == 4 {
			delete(w.packs, w.recent[0])
			w.recent = w.recent[1:]
		}
		sub, err := fs.Sub(w.root.FS(), dir)
		if err != nil {
			return nil, assets.Entry{}, err
		}
		pack, err = assets.Open(sub, "manifest.json")
		if err != nil {
			return nil, assets.Entry{}, err
		}
		w.packs[dir] = pack
	} else {
		for i, p := range w.recent {
			if p == dir {
				w.recent = append(w.recent[:i], w.recent[i+1:]...)
				break
			}
		}
	}
	w.recent = append(w.recent, dir)
	entry, err := pack.Entry(id, d)
	if err != nil {
		return nil, entry, err
	}
	im, err := pack.Image(id, d)
	return im, entry, err
}

func (w *waterBaker) optional(id string, d int) (image.Image, error) {
	if _, ok := w.dirs[waterAssetKey{id, d}]; !ok {
		id = mapcompose.SummerID(id)
	}
	if _, ok := w.dirs[waterAssetKey{id, d}]; !ok {
		return nil, nil
	}
	im, _, err := w.image(id, d)
	return im, err
}

func (w *waterBaker) bake(ctx context.Context, lv *level.Level, layers *BakedLayers) error {
	field := waterdepth.New(lv, layers.Scene)
	layers.Backgrounds = map[string]map[string]BakedBackground{}
	type cell struct {
		at                 hexgrid.Coord
		key, signature, id string
	}
	var cells []cell
	for line, row := range lv.Lines {
		for col := range row {
			if err := ctx.Err(); err != nil {
				return err
			}
			c := hexgrid.Coord{Line: line, Col: col}
			signature := field.Signature(c)
			if signature == "" {
				continue
			}
			id, _, _, _ := layers.Scene.Layers(c, func(*mapcompose.Cluster) bool { return true })
			if id == "" || id == "tile/unseen" {
				continue
			}
			name := fmt.Sprintf("%d,%d", line, col)
			layers.Backgrounds[name] = map[string]BakedBackground{}
			cells = append(cells, cell{c, name, signature, id})
		}
	}
	// Finish a density before changing packs, so a tile never forces three
	// different resolutions of every background/material pack into the LRU.
	for _, d := range []int{1, 2, 4} {
		for _, c := range cells {
			if err := ctx.Err(); err != nil {
				return err
			}
			// Surf samples absolute positions beyond the periodic depth signature.
			// The exact WaterTile is also Shade's water reference, so landOf is
			// nil there and no surf can require the absolute origin.
			var origin image.Point
			if c.id != waterdepth.WaterTile(c.id) {
				origin = waterOrigin(c.at, w.anchors[waterAssetKey{c.id, d}], d)
			}
			key := waterBakeKey{c.signature, c.id, d, origin}
			result, ok := w.cache[key]
			if !ok {
				var err error
				result, err = w.shade(ctx, field, c.at, c.id, d)
				if err != nil {
					return fmt.Errorf("water bake %s %s at %dx: %w", c.key, c.id, d, err)
				}
				w.cache[key] = result
			}
			layers.Backgrounds[c.key][strconv.Itoa(d)] = result
		}
	}
	return nil
}

func (w *waterBaker) shade(ctx context.Context, field *waterdepth.Field, c hexgrid.Coord, id string, d int) (BakedBackground, error) {
	tile, e, err := w.image(id, d)
	if err != nil {
		return BakedBackground{}, err
	}
	var with waterdepth.Tiles
	for _, extra := range []struct {
		id     string
		target *image.Image
	}{{waterdepth.SeaTile(c, id), &with.Sea}, {waterdepth.WaterTile(id), &with.Water}, {waterdepth.SurfTile(id), &with.Surf}} {
		*extra.target, err = w.optional(extra.id, d)
		if err != nil {
			return BakedBackground{}, err
		}
	}
	origin := waterOrigin(c, e.Anchor, d)
	shaded := field.Shade(tile, with, origin, d)
	if err := ctx.Err(); err != nil {
		return BakedBackground{}, err
	}
	var raw bytes.Buffer
	if err := png.Encode(&raw, shaded); err != nil {
		return BakedBackground{}, err
	}
	result := BakedBackground{ID: id, Width: e.Width, Height: e.Height, Density: d, Anchor: e.Anchor}
	if result.Full, err = w.blobs.put("png", raw.Bytes()); err != nil {
		return result, err
	}
	result.Lite = result.Full
	compressed, err := cachedLite(shaded, raw.Bytes())
	if err != nil {
		return result, err
	}
	if len(compressed.Data) < raw.Len() {
		result.Lite, err = w.blobs.put(compressed.Ext, compressed.Data)
	}
	return result, err
}

func waterOrigin(c hexgrid.Coord, anchor [2]float64, density int) image.Point {
	x, y := hexgrid.Center(c)
	return image.Pt(int(math.Round((x-anchor[0])*float64(density))), int(math.Round((y-anchor[1])*float64(density))))
}
