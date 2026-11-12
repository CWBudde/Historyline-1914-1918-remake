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
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/seasonblend"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/waterdepth"
)

type SeasonBakes struct {
	Layers   map[string]map[string]map[string]BakedBackground `json:"layers"`
	Ground   map[string][]mapcompose.Placed                   `json:"ground"`
	Standing map[string][]mapcompose.Placed                   `json:"standing"`
}

// bakeSeasons preserves the native border-layer mix, including its complete
// image union outside the scalar-weight raster. Empty pass maps suppress the
// main scene's corresponding layer; empty placement entries do the same for
// its ground props and scattered crowns. Stamps and shadows remain unchanged.
func (w *waterBaker) bakeSeasons(ctx context.Context, lv *level.Level, layers *BakedLayers) (SeasonBakes, error) {
	out := SeasonBakes{Layers: map[string]map[string]map[string]BakedBackground{}, Ground: map[string][]mapcompose.Placed{}, Standing: map[string][]mapcompose.Placed{}}
	if err := ctx.Err(); err != nil {
		return out, err
	}
	if lv == nil || layers == nil || layers.Scene == nil {
		return out, fmt.Errorf("season bake: missing level or scene")
	}
	has := func(id string) bool {
		for _, d := range []int{1, 2, 4} {
			if _, ok := w.dirs[waterAssetKey{id, d}]; !ok {
				return false
			}
		}
		return true
	}
	blend := seasonblend.New(lv, has)
	if blend == nil {
		return out, nil
	}
	if len(layers.Seasons) > 0 {
		blend.Scenes = layers.Seasons
	}
	field := waterdepth.New(lv, layers.Scene)
	known := func(*mapcompose.Cluster) bool { return true }
	var border []hexgrid.Coord
	for line, row := range lv.Lines {
		for col := range row {
			if err := ctx.Err(); err != nil {
				return out, err
			}
			c := hexgrid.Coord{Line: line, Col: col}
			if len(blend.Seasons(c)) == 0 {
				continue
			}
			for _, season := range blend.Seasons(c) {
				if blend.Scenes[season] == nil {
					return out, fmt.Errorf("season bake: missing %s scene", season)
				}
			}
			key := fmt.Sprintf("%d,%d", line, col)
			out.Layers[key] = map[string]map[string]BakedBackground{}
			for _, pass := range []string{"background", "infrastructure", "standing"} {
				out.Layers[key][pass] = map[string]BakedBackground{}
			}
			out.Ground[key], _ = blend.Scattered(c, func(s *mapcompose.Scene) []mapcompose.Placed { return s.GroundAt(c, known) })
			out.Standing[key], _ = blend.Scattered(c, func(s *mapcompose.Scene) []mapcompose.Placed { return s.ScatterAt(c, known) })
			border = append(border, c)
		}
	}
	// Keep native image snapshots at one density at a time, as the water bake.
	for _, d := range []int{1, 2, 4} {
		for _, c := range border {
			for _, pass := range []string{"background", "infrastructure", "standing"} {
				if err := ctx.Err(); err != nil {
					return out, err
				}
				entry, present, err := w.seasonLayer(ctx, blend, field, c, pass, d)
				if err != nil {
					return out, fmt.Errorf("season bake %d,%d %s %dx: %w", c.Line, c.Col, pass, d, err)
				}
				if present {
					out.Layers[fmt.Sprintf("%d,%d", c.Line, c.Col)][pass][strconv.Itoa(d)] = entry
				}
			}
		}
	}
	return out, nil
}

func (w *waterBaker) seasonLayer(ctx context.Context, blend *seasonblend.Blend, field *waterdepth.Field, c hexgrid.Coord, pass string, d int) (BakedBackground, bool, error) {
	seasons := blend.Seasons(c)
	parts := make([][]seasonblend.Part, len(seasons))
	for i, season := range seasons {
		bg, _, overlays, standing := blend.Scenes[season].Layers(c, func(*mapcompose.Cluster) bool { return true })
		var ids []string
		switch pass {
		case "background":
			ids = []string{bg}
		case "infrastructure":
			ids = overlays
		case "standing":
			ids = standing
		}
		for _, id := range ids {
			if err := ctx.Err(); err != nil {
				return BakedBackground{}, false, err
			}
			if id == "" {
				continue
			}
			im, entry, err := w.image(id, d)
			if err != nil {
				return BakedBackground{}, false, err
			}
			origin := waterOrigin(c, entry.Anchor, d)
			if pass == "background" && field.Signature(c) != "" {
				var with waterdepth.Tiles
				with.Sea, err = w.optional(waterdepth.SeaTile(c, id), d)
				if err != nil {
					return BakedBackground{}, false, err
				}
				with.Water, err = w.optional(waterdepth.WaterTile(id), d)
				if err != nil {
					return BakedBackground{}, false, err
				}
				with.Surf, err = w.optional(waterdepth.SurfTile(id), d)
				if err != nil {
					return BakedBackground{}, false, err
				}
				im = field.Shade(im, with, origin, d)
			}
			if pass == "background" {
				if im, err = w.reworked(blend.Scenes[season], c, id, im, origin, d); err != nil {
					return BakedBackground{}, false, err
				}
			}
			parts[i] = append(parts[i], seasonblend.Part{Image: im, At: origin})
		}
	}
	mixed := blend.Mix(c, parts, d)
	if mixed == nil {
		return BakedBackground{}, false, nil
	}
	if err := ctx.Err(); err != nil {
		return BakedBackground{}, false, err
	}
	x, y := hexgrid.Center(c)
	result := BakedBackground{ID: fmt.Sprintf("mixed/%s/%d,%d", pass, c.Line, c.Col), Width: mixed.Bounds().Dx(), Height: mixed.Bounds().Dy(), Density: d, Anchor: [2]float64{x - float64(mixed.Bounds().Min.X)/float64(d), y - float64(mixed.Bounds().Min.Y)/float64(d)}}
	var raw bytes.Buffer
	straight := straightSeason(mixed)
	if err := png.Encode(&raw, straight); err != nil {
		return result, false, err
	}
	var err error
	if result.Full, err = w.blobs.put("png", raw.Bytes()); err != nil {
		return result, false, err
	}
	result.Lite = result.Full
	compressed, err := cachedLite(straight, raw.Bytes(), hlasset.LiteWebP)
	if err != nil {
		return result, false, err
	}
	if len(compressed.Data) < raw.Len() {
		result.Lite, err = w.blobs.put(compressed.Ext, compressed.Data)
	}
	return result, true, err
}

// PNG stores straight alpha. Rounding unpremultiplication down can lose a full
// native RGB value on decoding; the ceiling preserves every premultiplied
// 8-bit channel when the straight pixel is drawn back into an RGBA surface.
func straightSeason(src *image.RGBA) *image.NRGBA {
	out := image.NewNRGBA(src.Bounds())
	for y := src.Rect.Min.Y; y < src.Rect.Max.Y; y++ {
		for x := src.Rect.Min.X; x < src.Rect.Max.X; x++ {
			i, j := src.PixOffset(x, y), out.PixOffset(x, y)
			a := int(src.Pix[i+3])
			out.Pix[j+3] = uint8(a)
			if a == 0 {
				continue
			}
			for ch := range 3 {
				denominator := a * 0x101
				out.Pix[j+ch] = uint8((int(src.Pix[i+ch])*0xff00 + denominator - 1) / denominator)
			}
		}
	}
	return out
}
