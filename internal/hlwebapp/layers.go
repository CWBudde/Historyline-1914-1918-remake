package hlwebapp

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/seasonblend"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/waterdepth"
)

// Raster is sampled at world pixel centres (x+.5,y+.5). Water uses 0..65535;
// season weights use 0..256. Both store the uint16 in PNG R (low) and G (high),
// with opaque alpha, so browser image decoding retains all bits. Divide by
// Divisor after reconstructing the integer. They need no JS noise port.
type Raster struct {
	X       int     `json:"x"`
	Y       int     `json:"y"`
	Width   int     `json:"width"`
	Height  int     `json:"height"`
	Divisor int     `json:"divisor"`
	File    FileRef `json:"file"`
}
type SeasonMask struct {
	At      hexgrid.Coord       `json:"at"`
	X       int                 `json:"x"`
	Y       int                 `json:"y"`
	Seasons []mapcompose.Season `json:"seasons"`
	Weights []Raster            `json:"weights"`
}
type BakedLayers struct {
	SeasonBakes *SeasonBakes                            `json:"seasonBakes,omitempty"`
	Backgrounds map[string]map[string]BakedBackground   `json:"backgrounds,omitempty"`
	Schema      int                                     `json:"schema"`
	Scene       *mapcompose.Scene                       `json:"scene"`
	Seasons     map[mapcompose.Season]*mapcompose.Scene `json:"seasons"`
	Water       map[string]Raster                       `json:"water"`
	Blend       map[string]SeasonMask                   `json:"blend"`
}

func (b blobWriter) layers(ctx context.Context, lv *level.Level, has func(string) bool) (BakedLayers, error) {
	scene := mapcompose.ComposeSeason(lv.Grid(), mapcompose.LevelHex(lv), has, mapcompose.SeasonAt(lv))
	depth := waterdepth.New(lv, &scene)
	blend := seasonblend.New(lv, has)
	out := BakedLayers{Schema: 1, Scene: &scene, Water: map[string]Raster{}, Blend: map[string]SeasonMask{}, Seasons: map[mapcompose.Season]*mapcompose.Scene{}}
	if blend != nil {
		out.Seasons = blend.Scenes
	}
	for line, row := range lv.Lines {
		for col, h := range row {
			if err := ctx.Err(); err != nil {
				return out, err
			}
			c := hexgrid.Coord{Line: line, Col: col}
			cx, cy := hexgrid.Center(c)
			key := fmt.Sprintf("%d,%d", line, col)
			if mapcompose.OpenWater(h.Terrain) {
				r := Raster{X: int(cx) - 48, Y: int(cy) - 40, Width: 96, Height: 80, Divisor: 65535}
				im := image.NewNRGBA(image.Rect(0, 0, r.Width, r.Height))
				some := false
				for y := range r.Height {
					for x := range r.Width {
						v := uint16(math.Round(depth.At(float64(r.X+x)+0.5, float64(r.Y+y)+0.5) * 65535))
						im.SetNRGBA(x, y, packed(v))
						some = some || v > 0
					}
				}
				if some {
					var err error
					r.File, err = b.png(im)
					if err != nil {
						return out, err
					}
					out.Water[key] = r
				}
			}
			if seasons := blend.Seasons(c); len(seasons) > 0 {
				m := SeasonMask{At: c, X: int(cx) - 64, Y: int(cy) - 56, Seasons: seasons}
				images := make([]*image.NRGBA, len(seasons))
				for i := range images {
					images[i] = image.NewNRGBA(image.Rect(0, 0, 128, 112))
				}
				for y := range 112 {
					for x := range 128 {
						q := blend.Weights(c, float64(m.X+x)+0.5, float64(m.Y+y)+0.5)
						for i, v := range q {
							images[i].SetNRGBA(x, y, packed(uint16(v)))
						}
					}
				}
				for _, im := range images {
					ref, err := b.png(im)
					if err != nil {
						return out, err
					}
					m.Weights = append(m.Weights, Raster{X: m.X, Y: m.Y, Width: 128, Height: 112, Divisor: 256, File: ref})
				}
				out.Blend[key] = m
			}
		}
	}
	return out, nil
}

func packed(v uint16) color.NRGBA { return color.NRGBA{R: uint8(v), G: uint8(v >> 8), A: 255} }
