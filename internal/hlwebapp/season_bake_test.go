package hlwebapp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/seasonblend"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/waterdepth"
)

func TestSeasonBakeMatchesNativeMix(t *testing.T) {
	dir, lv, layers := seasonBakeFixture(t)
	b := blobWriter{root: t.TempDir()}
	w, err := newWaterBaker(dir, b)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := w.close(); err != nil {
			t.Error(err)
		}
	})
	before, err := json.Marshal(layers)
	if err != nil {
		t.Fatal(err)
	}
	got, err := w.bakeSeasons(t.Context(), lv, layers)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Layers) == 0 {
		t.Fatal("missing mixed border layers")
	}
	blend := seasonblend.New(lv, seasonHas(w))
	blend.Scenes = layers.Seasons
	field := waterdepth.New(lv, layers.Scene)
	overhang := false
	for line, row := range lv.Lines {
		for col := range row {
			c := hexgrid.Coord{Line: line, Col: col}
			if len(blend.Seasons(c)) == 0 {
				continue
			}
			key := fmt.Sprintf("%d,%d", line, col)
			for _, pass := range []string{"background", "infrastructure", "standing"} {
				for _, d := range []int{1, 2, 4} {
					parts := nativeSeasonParts(t, w, blend, field, c, pass, d)
					want := blend.Mix(c, parts, d)
					entry, ok := got.Layers[key][pass][strconv.Itoa(d)]
					if want == nil {
						if ok {
							t.Fatalf("empty %s produced a texture", pass)
						}
						continue
					}
					if !ok {
						t.Fatalf("missing %s %s %dx", key, pass, d)
					}
					x, y := hexgrid.Center(c)
					anchor := [2]float64{x - float64(want.Bounds().Min.X)/float64(d), y - float64(want.Bounds().Min.Y)/float64(d)}
					if entry.Width != want.Bounds().Dx() || entry.Height != want.Bounds().Dy() || entry.Anchor != anchor || entry.Density != d {
						t.Fatalf("mixed geometry %+v, bounds %v", entry, want.Bounds())
					}
					if pass == "standing" && entry.Width > 128*d && entry.Height > 112*d {
						overhang = true
					}
					raw, err := os.ReadFile(filepath.Join(b.root, entry.Full.Path))
					if err != nil {
						t.Fatal(err)
					}
					pic, err := png.Decode(bytes.NewReader(raw))
					if err != nil {
						t.Fatal(err)
					}
					actual := image.NewRGBA(pic.Bounds())
					draw.Draw(actual, actual.Rect, pic, pic.Bounds().Min, draw.Src)
					for py := range entry.Height {
						for px := range entry.Width {
							if actual.RGBAAt(px, py) != want.RGBAAt(want.Rect.Min.X+px, want.Rect.Min.Y+py) {
								t.Fatalf("%s %s %dx pixel %d,%d differs: %v != %v", key, pass, d, px, py, actual.RGBAAt(px, py), want.RGBAAt(want.Rect.Min.X+px, want.Rect.Min.Y+py))
							}
						}
					}
					if entry.Lite.Bytes > entry.Full.Bytes {
						t.Fatal("lite larger than full")
					}
				}
			}
			ground, _ := blend.Scattered(c, func(s *mapcompose.Scene) []mapcompose.Placed { return s.GroundAt(c, allSeasonsKnown) })
			standing, _ := blend.Scattered(c, func(s *mapcompose.Scene) []mapcompose.Placed { return s.ScatterAt(c, allSeasonsKnown) })
			if !reflect.DeepEqual(got.Ground[key], ground) || !reflect.DeepEqual(got.Standing[key], standing) {
				t.Fatalf("%s scattered placements differ", key)
			}
		}
	}
	if !overhang {
		t.Fatal("fixture did not exercise overhang beyond baked weight raster")
	}
	after, err := json.Marshal(layers)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("baking mutated the scene or its shadows")
	}
}

func TestSeasonBakeEmptyLayersAndCancellation(t *testing.T) {
	dir, lv, layers := seasonBakeFixture(t)
	w, err := newWaterBaker(dir, blobWriter{root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := w.close(); err != nil {
			t.Error(err)
		}
	})
	for _, s := range layers.Seasons {
		for _, row := range s.Cells {
			for col := range row {
				row[col].Overlays = nil
				row[col].Standing = nil
			}
		}
	}
	got, err := w.bakeSeasons(t.Context(), lv, layers)
	if err != nil {
		t.Fatal(err)
	}
	for key, passes := range got.Layers {
		for _, pass := range []string{"infrastructure", "standing"} {
			entries, ok := passes[pass]
			if !ok || len(entries) != 0 {
				t.Fatalf("%s must explicitly suppress empty %s", key, pass)
			}
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := w.bakeSeasons(ctx, nil, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
	lv.Extra.Seasons = nil
	got, err = w.bakeSeasons(t.Context(), lv, layers)
	if err != nil || len(got.Layers) != 0 {
		t.Fatalf("one-season map: %+v %v", got, err)
	}
}

func TestSeasonBakeRejectsMissingImage(t *testing.T) {
	dir, lv, layers := seasonBakeFixture(t)
	w, err := newWaterBaker(dir, blobWriter{root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := w.close(); err != nil {
			t.Error(err)
		}
	})
	for _, s := range layers.Seasons {
		s.Cells[0][0].Background = "tile/missing"
	}
	if _, err := w.bakeSeasons(t.Context(), lv, layers); err == nil {
		t.Fatal("missing required image accepted")
	}
}

func TestSeasonBakeShadesWaterBeforeMix(t *testing.T) {
	dir, _, _, _ := waterBakeFixture(t)
	w, err := newWaterBaker(dir, blobWriter{root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := w.close(); err != nil {
			t.Error(err)
		}
	})
	lv := &level.Level{Extra: &level.Extra{Season: "summer", Seasons: map[hexgrid.Coord]string{{Line: 1, Col: 0}: "autumn"}}}
	for line := range 3 {
		lv.Lines = append(lv.Lines, make([]level.Hex, 2-line%2))
		for col := range lv.Lines[line] {
			lv.Lines[line][col].Terrain = "waterSea"
		}
	}
	has := seasonHas(w)
	scene := mapcompose.ComposeSeason(lv.Grid(), mapcompose.LevelHex(lv), has, mapcompose.SeasonAt(lv))
	blend := seasonblend.New(lv, has)
	field := waterdepth.New(lv, &scene)
	got, err := w.bakeSeasons(t.Context(), lv, &BakedLayers{Scene: &scene, Seasons: blend.Scenes})
	if err != nil {
		t.Fatal(err)
	}
	c := hexgrid.Coord{Line: 1, Col: 0}
	if field.Signature(c) == "" {
		t.Fatal("fixture has no depth shading")
	}
	for _, d := range []int{1, 2, 4} {
		want := blend.Mix(c, nativeSeasonParts(t, w, blend, field, c, "background", d), d)
		if want == nil {
			t.Fatal("missing native water mix")
		}
		entry := got.Layers["1,0"]["background"][strconv.Itoa(d)]
		raw, err := os.ReadFile(filepath.Join(w.blobs.root, entry.Full.Path))
		if err != nil {
			t.Fatal(err)
		}
		pic, err := png.Decode(bytes.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		actual := image.NewRGBA(pic.Bounds())
		draw.Draw(actual, actual.Rect, pic, pic.Bounds().Min, draw.Src)
		for y := range entry.Height {
			for x := range entry.Width {
				if actual.RGBAAt(x, y) != want.RGBAAt(want.Rect.Min.X+x, want.Rect.Min.Y+y) {
					t.Fatalf("%dx water pixel %d,%d differs", d, x, y)
				}
			}
		}
	}
}

func TestStraightSeasonPreservesEveryPremultipliedChannel(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 256, 256))
	for a := range 256 {
		for v := 0; v <= a; v++ {
			src.SetRGBA(v, a, color.RGBA{uint8(v), uint8(v), uint8(v), uint8(a)})
		}
	}
	var raw bytes.Buffer
	if err := png.Encode(&raw, straightSeason(src)); err != nil {
		t.Fatal(err)
	}
	decoded, err := png.Decode(bytes.NewReader(raw.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	got := image.NewRGBA(src.Bounds())
	draw.Draw(got, got.Rect, decoded, decoded.Bounds().Min, draw.Src)
	if !bytes.Equal(got.Pix, src.Pix) {
		t.Fatal("straight PNG changed native premultiplied values")
	}
}

func seasonHas(w *waterBaker) func(string) bool {
	return func(id string) bool {
		for _, d := range []int{1, 2, 4} {
			if _, ok := w.dirs[waterAssetKey{id, d}]; !ok {
				return false
			}
		}
		return true
	}
}

func allSeasonsKnown(*mapcompose.Cluster) bool { return true }

func nativeSeasonParts(t *testing.T, w *waterBaker, b *seasonblend.Blend, field *waterdepth.Field, c hexgrid.Coord, pass string, d int) [][]seasonblend.Part {
	t.Helper()
	parts := make([][]seasonblend.Part, len(b.Seasons(c)))
	for i, season := range b.Seasons(c) {
		bg, _, overlays, standing := b.Scenes[season].Layers(c, allSeasonsKnown)
		ids := map[string][]string{"background": {bg}, "infrastructure": overlays, "standing": standing}[pass]
		for _, id := range ids {
			if id == "" {
				continue
			}
			pic, entry, err := w.image(id, d)
			if err != nil {
				t.Fatal(err)
			}
			origin := waterOrigin(c, entry.Anchor, d)
			if pass == "background" && field.Signature(c) != "" {
				var with waterdepth.Tiles
				with.Sea, err = w.optional(waterdepth.SeaTile(c, id), d)
				if err != nil {
					t.Fatal(err)
				}
				with.Water, err = w.optional(waterdepth.WaterTile(id), d)
				if err != nil {
					t.Fatal(err)
				}
				with.Surf, err = w.optional(waterdepth.SurfTile(id), d)
				if err != nil {
					t.Fatal(err)
				}
				pic = field.Shade(pic, with, origin, d)
			}
			parts[i] = append(parts[i], seasonblend.Part{Image: pic, At: origin})
		}
	}
	return parts
}

func seasonBakeFixture(t *testing.T) (string, *level.Level, *BakedLayers) {
	t.Helper()
	dir := t.TempDir()
	for _, d := range []int{1, 2, 4} {
		folder := filepath.Join(dir, strconv.Itoa(d)+"x", "fixture")
		if err := os.MkdirAll(folder, 0o755); err != nil {
			t.Fatal(err)
		}
		m := assets.Manifest{Schema: 1, ID: fmt.Sprintf("season-test/%d", d), Status: "approved"}
		for i, id := range []string{"tile/grass", "tile/grass_w", "tile/road14", "tile/road14_w", "standing/grass", "standing/grass_w"} {
			width, height, anchor := 96, 80, [2]float64{47.25, 39.75}
			if i >= 4 {
				width, height, anchor = 180, 140, [2]float64{90.25, 70.75}
			}
			pic := image.NewNRGBA(image.Rect(0, 0, width*d, height*d))
			for y := range pic.Rect.Dy() {
				for x := range pic.Rect.Dx() {
					pic.SetNRGBA(x, y, color.NRGBA{uint8((x+17*i)%180 + 20), uint8((y+11*i)%180 + 20), 70 + uint8(i*20), 60 + uint8((x+y+i)%190)})
				}
			}
			var raw bytes.Buffer
			if err := png.Encode(&raw, pic); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(raw.Bytes())
			name := fmt.Sprintf("%d.png", i)
			m.Entries = append(m.Entries, assets.Entry{ID: id, Path: name, SHA256: hex.EncodeToString(sum[:]), Width: width * d, Height: height * d, Density: d, Anchor: anchor})
			if err := os.WriteFile(filepath.Join(folder, name), raw.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		raw, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(folder, "manifest.json"), raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	lv := &level.Level{Extra: &level.Extra{Season: "summer", Seasons: map[hexgrid.Coord]string{{Line: 1, Col: 0}: "winter"}}}
	for line := range 3 {
		lv.Lines = append(lv.Lines, make([]level.Hex, 2-line%2))
		for col := range lv.Lines[line] {
			lv.Lines[line][col] = level.Hex{Terrain: "grass", Infra: level.Infra{IDs: []string{"road14"}}}
		}
	}
	has := func(id string) bool {
		for _, known := range []string{"tile/grass", "tile/grass_w", "tile/road14", "tile/road14_w", "standing/grass", "standing/grass_w"} {
			if id == known {
				return true
			}
		}
		return false
	}
	scene := mapcompose.ComposeSeason(lv.Grid(), mapcompose.LevelHex(lv), has, mapcompose.SeasonAt(lv))
	blend := seasonblend.New(lv, has)
	for season, s := range blend.Scenes {
		for line, row := range s.Cells {
			for col := range row {
				row[col].Ground = []mapcompose.Placed{{ID: "ground/" + string(season), X: 1.25, Y: -2}}
				row[col].Scatter = []mapcompose.Placed{{ID: "crown/" + string(season), X: -3.5, Y: 4}}
				if line == 0 && col == 0 {
					row[col].Shadows = []string{"unchanged-shadow"}
				}
			}
		}
	}
	return dir, lv, &BakedLayers{Scene: &scene, Seasons: blend.Scenes}
}
