package hlwebapp

import (
	"bytes"
	"context"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/seasonblend"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/waterdepth"
)

func TestBakedLayersUseNativeSceneAndFields(t *testing.T) {
	lv := &level.Level{Lines: [][]level.Hex{{{Terrain: "waterSea"}, {Terrain: "waterDeep"}, {Terrain: "grass"}}, {{Terrain: "water"}, {Terrain: "grass"}}, {{Terrain: "waterSea"}, {Terrain: "waterDeep"}, {Terrain: "grass"}}}, Extra: &level.Extra{Season: "summer", Seasons: map[hexgrid.Coord]string{{Line: 0, Col: 2}: "winter"}}}
	has := func(string) bool { return false }
	b := blobWriter{root: t.TempDir()}
	layers, err := b.layers(context.Background(), lv, has)
	if err != nil {
		t.Fatal(err)
	}
	scene := mapcompose.ComposeSeason(lv.Grid(), mapcompose.LevelHex(lv), has, mapcompose.SeasonAt(lv))
	if !reflect.DeepEqual(layers.Scene, &scene) {
		t.Fatal("draw lists differ")
	}
	field := waterdepth.New(lv, &scene)
	if len(layers.Water) == 0 || len(layers.Blend) == 0 || len(layers.Seasons) != 2 {
		t.Fatal("missing baked fields")
	}
	for _, r := range layers.Water {
		raw, err := os.ReadFile(filepath.Join(b.root, r.File.Path))
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		for y := 0; y < r.Height; y += 13 {
			for x := 0; x < r.Width; x += 13 {
				red, green, _, _ := im.At(x, y).RGBA()
				v := (red >> 8) | (green>>8)<<8
				want := uint32(math.Round(field.At(float64(r.X+x)+0.5, float64(r.Y+y)+0.5) * 65535))
				if v != want {
					t.Fatalf("depth %d != %d", v, want)
				}
			}
		}
	}
	blend := seasonblend.New(lv, has)
	for _, mask := range layers.Blend {
		q := blend.Weights(mask.At, float64(mask.X)+0.5, float64(mask.Y)+0.5)
		for i, r := range mask.Weights {
			raw, err := os.ReadFile(filepath.Join(b.root, r.File.Path))
			if err != nil {
				t.Fatal(err)
			}
			im, err := png.Decode(bytes.NewReader(raw))
			if err != nil {
				t.Fatal(err)
			}
			red, green, _, _ := im.At(0, 0).RGBA()
			v := (red >> 8) | (green>>8)<<8
			if int(v) != q[i] {
				t.Fatal("season weight differs")
			}
		}
	}
}
