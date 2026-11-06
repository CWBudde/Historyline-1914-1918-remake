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
	"image/png"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/waterdepth"
)

// Omitting the native sea/surf inputs, selecting Cell.Background instead of
// Layers, or rounding the density origin incorrectly must change these pixels.
func TestWaterBakeMatchesNativeShade(t *testing.T) {
	dir, lv, scene, pictures := waterBakeFixture(t)
	b := blobWriter{root: t.TempDir()}
	baker, err := newWaterBaker(dir, b)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := baker.close(); err != nil {
			t.Error(err)
		}
	}()
	layers := BakedLayers{Scene: scene}
	if err := baker.bake(context.Background(), lv, &layers); err != nil {
		t.Fatal(err)
	}
	field := waterdepth.New(lv, scene)
	count := 0
	for line, row := range lv.Lines {
		for col := range row {
			c := hexgrid.Coord{Line: line, Col: col}
			key := fmt.Sprintf("%d,%d", line, col)
			baked := layers.Backgrounds[key]
			if field.Signature(c) == "" {
				if len(baked) != 0 {
					t.Fatalf("unshaded cell %s baked", key)
				}
				continue
			}
			count++
			id, _, _, _ := scene.Layers(c, func(*mapcompose.Cluster) bool { return true })
			for _, d := range []int{1, 2, 4} {
				e := baked[strconv.Itoa(d)]
				if e.ID != id || e.Density != d || e.Width != 96*d || e.Height != 80*d || e.Anchor != [2]float64{47.25, 39.75} {
					t.Fatalf("geometry %s: %+v", key, e)
				}
				raw, err := os.ReadFile(filepath.Join(b.root, e.Full.Path))
				if err != nil {
					t.Fatal(err)
				}
				got, err := png.Decode(bytes.NewReader(raw))
				if err != nil {
					t.Fatal(err)
				}
				x, y := hexgrid.Center(c)
				with := waterdepth.Tiles{Sea: pictures[d][mapcompose.SummerID(waterdepth.SeaTile(c, id))], Water: pictures[d]["tile/water"], Surf: pictures[d]["tile/surf"]}
				want := field.Shade(pictures[d][id], with, image.Pt(int(math.Round((x-47.25)*float64(d))), int(math.Round((y-39.75)*float64(d)))), d)
				for y := range e.Height {
					for x := range e.Width {
						if color.NRGBAModel.Convert(got.At(x, y)) != want.NRGBAAt(x, y) {
							t.Fatalf("%s %dx pixel %d,%d differs", key, d, x, y)
						}
					}
				}
				if e.Lite.Bytes > e.Full.Bytes {
					t.Fatal("lite expanded image")
				}
				if _, err := os.Stat(filepath.Join(b.root, e.Lite.Path)); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	if count < 3 {
		t.Fatal("fixture lacks shaded shore and water")
	}
	// Cached references remain usable without reading the source again, including
	// across different setup objects with identical terrain and signatures.
	first := layers.Backgrounds
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	layers.Backgrounds = nil
	otherSetup := *lv
	if err := baker.bake(context.Background(), &otherSetup, &layers); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, layers.Backgrounds) {
		t.Fatal("setup cache changed references")
	}
}

func TestWaterBakeRejectsChangedSource(t *testing.T) {
	dir, lv, scene, _ := waterBakeFixture(t)
	baker, err := newWaterBaker(dir, blobWriter{root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := baker.close(); err != nil {
			t.Error(err)
		}
	}()
	if err := os.WriteFile(filepath.Join(dir, "1x", "backgrounds", "0.png"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := baker.bake(context.Background(), lv, &BakedLayers{Scene: scene}); err == nil {
		t.Fatal("accepted a source whose pixels no longer match its manifest hash")
	}
}

// Translating four columns preserves the depth/warp/sea signature, while surf
// samples different absolute world positions. These tiles must not share the
// first level's baked pixels even though their depth signatures match.
func TestWaterBakeKeepsTranslatedSurfOriginsSeparate(t *testing.T) {
	dir, _, _, pictures := waterBakeFixture(t)
	b := blobWriter{root: t.TempDir()}
	baker, err := newWaterBaker(dir, b)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := baker.close(); err != nil {
			t.Error(err)
		}
	}()
	var signature string
	var previous map[int]*image.NRGBA
	for _, col := range []int{4, 8} {
		lv, scene, c := waterBakeIsland(col, "tile/shore-test_au")
		field := waterdepth.New(lv, scene)
		gotSignature := field.Signature(c)
		if gotSignature == "" || signature != "" && signature != gotSignature {
			t.Fatal("fixture must collide in the native depth signature")
		}
		signature = gotSignature
		layers := BakedLayers{Scene: scene}
		if err := baker.bake(context.Background(), lv, &layers); err != nil {
			t.Fatal(err)
		}
		native := map[int]*image.NRGBA{}
		for _, d := range []int{1, 2, 4} {
			id := "tile/shore-test_au"
			x, y := hexgrid.Center(c)
			origin := image.Pt(int(math.Round((x-47.25)*float64(d))), int(math.Round((y-39.75)*float64(d))))
			with := waterdepth.Tiles{Water: pictures[d]["tile/water"], Sea: pictures[d][mapcompose.SummerID(waterdepth.SeaTile(c, id))], Surf: pictures[d]["tile/surf"]}
			want := field.Shade(pictures[d][id], with, origin, d)
			if previous != nil && bytes.Equal(previous[d].Pix, want.Pix) {
				t.Fatalf("%dx fixture must change native surf pixels with origin", d)
			}
			native[d] = want
			e := layers.Backgrounds[fmt.Sprintf("%d,%d", c.Line, c.Col)][strconv.Itoa(d)]
			raw, err := os.ReadFile(filepath.Join(b.root, e.Full.Path))
			if err != nil {
				t.Fatal(err)
			}
			got, err := png.Decode(bytes.NewReader(raw))
			if err != nil {
				t.Fatal(err)
			}
			for py := range e.Height {
				for px := range e.Width {
					if color.NRGBAModel.Convert(got.At(px, py)) != want.NRGBAAt(px, py) {
						t.Fatalf("translated coast %d,%d %dx reused the wrong surf origin at pixel %d,%d", c.Line, c.Col, d, px, py)
					}
				}
			}
		}
		previous = native
	}
}

func TestWaterBakeReusesTranslatedOpenWater(t *testing.T) {
	dir, _, _, pictures := waterBakeFixture(t)
	b := blobWriter{root: t.TempDir()}
	baker, err := newWaterBaker(dir, b)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := baker.close(); err != nil {
			t.Error(err)
		}
	}()
	var signature string
	var previous map[int]*image.NRGBA
	var firstCacheEntries int
	for _, col := range []int{4, 8} {
		id := "tile/water_au"
		lv, scene, c := waterBakeIsland(col, id)
		field := waterdepth.New(lv, scene)
		if s := field.Signature(c); s == "" || signature != "" && signature != s {
			t.Fatal("fixture needs matching nonempty signatures")
		} else {
			signature = s
		}
		layers := BakedLayers{Scene: scene}
		if err := baker.bake(context.Background(), lv, &layers); err != nil {
			t.Fatal(err)
		}
		native := map[int]*image.NRGBA{}
		for _, d := range []int{1, 2, 4} {
			x, y := hexgrid.Center(c)
			origin := image.Pt(int(math.Round((x-47.25)*float64(d))), int(math.Round((y-39.75)*float64(d))))
			with := waterdepth.Tiles{Water: pictures[d][id], Sea: pictures[d][mapcompose.SummerID(waterdepth.SeaTile(c, id))], Surf: pictures[d]["tile/surf"]}
			want := field.Shade(pictures[d][id], with, origin, d)
			if previous != nil && !bytes.Equal(previous[d].Pix, want.Pix) {
				t.Fatalf("%dx translated pure water must have identical native pixels", d)
			}
			native[d] = want
			e := layers.Backgrounds[fmt.Sprintf("%d,%d", c.Line, c.Col)][strconv.Itoa(d)]
			raw, err := os.ReadFile(filepath.Join(b.root, e.Full.Path))
			if err != nil {
				t.Fatal(err)
			}
			got, err := png.Decode(bytes.NewReader(raw))
			if err != nil {
				t.Fatal(err)
			}
			for py := range e.Height {
				for px := range e.Width {
					if color.NRGBAModel.Convert(got.At(px, py)) != want.NRGBAAt(px, py) {
						t.Fatalf("%dx translated open water differs at %d,%d", d, px, py)
					}
				}
			}
		}
		if previous != nil && len(baker.cache) != firstCacheEntries {
			t.Fatal("translated pure water performed extra bakes despite matching native signatures and pixels")
		}
		previous = native
		firstCacheEntries = len(baker.cache)
	}
}

func waterBakeIsland(col int, id string) (*level.Level, *mapcompose.Scene, hexgrid.Coord) {
	c := hexgrid.Coord{Line: 4, Col: col}
	lv := &level.Level{Lines: make([][]level.Hex, 9)}
	scene := &mapcompose.Scene{Cells: make([][]mapcompose.Cell, 9)}
	for line := range lv.Lines {
		lv.Lines[line] = make([]level.Hex, 14-line%2)
		scene.Cells[line] = make([]mapcompose.Cell, len(lv.Lines[line]))
		for x := range lv.Lines[line] {
			lv.Lines[line][x].Terrain = "grass"
			scene.Cells[line][x].Background = "tile/grass"
		}
	}
	for _, n := range hexgrid.Around(c) {
		lv.Lines[n.Line][n.Col].Terrain = "waterSea"
		scene.Cells[n.Line][n.Col].Background = "tile/water_au"
	}
	lv.Lines[c.Line][c.Col].Terrain = "waterStones"
	if id == "tile/water_au" {
		lv.Lines[c.Line][c.Col].Terrain = "waterSea"
	}
	scene.Cells[c.Line][c.Col].Background = id
	scene.Grid = lv.Grid()
	return lv, scene, c
}

func TestWaterBakeRetainsCompleteFourPackWorkingSet(t *testing.T) {
	dir, _, _, _ := waterBakeFixture(t)
	// Native shores, sea, water and surf live in four distinct pack groups.
	// Split the synthetic pack the same way, with its existing verified bytes.
	for _, d := range []int{1, 2, 4} {
		base := filepath.Join(dir, strconv.Itoa(d)+"x")
		original := filepath.Join(base, "backgrounds")
		raw, err := os.ReadFile(filepath.Join(original, "manifest.json"))
		if err != nil {
			t.Fatal(err)
		}
		m, err := assets.ParseManifest(raw)
		if err != nil {
			t.Fatal(err)
		}
		groups := map[string]assets.Manifest{}
		for _, e := range m.Entries {
			group := "water"
			switch {
			case strings.HasPrefix(e.ID, "tile/shore-"):
				group = "backgrounds"
			case strings.HasPrefix(e.ID, "tile/sea-"):
				group = "sea"
			case e.ID == "tile/surf":
				group = "surf"
			}
			next := groups[group]
			next.Schema, next.Status, next.ID = m.Schema, m.Status, m.ID+"/"+group
			next.Entries = append(next.Entries, e)
			groups[group] = next
			if group != "backgrounds" {
				if err := os.MkdirAll(filepath.Join(base, group), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(filepath.Join(original, e.Path), filepath.Join(base, group, e.Path)); err != nil {
					t.Fatal(err)
				}
			}
		}
		for group, manifest := range groups {
			raw, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(base, group, "manifest.json"), raw, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	baker, err := newWaterBaker(dir, blobWriter{root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := baker.close(); err != nil {
			t.Error(err)
		}
	}()
	lv, scene, c := waterBakeIsland(4, "tile/shore-test_au")
	first, err := baker.shade(context.Background(), waterdepth.New(lv, scene), c, "tile/shore-test_au", 4)
	if err != nil {
		t.Fatal(err)
	}
	// The complete immutable working set must still serve a different shore
	// after source files are unavailable, without reopening any of its packs.
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	lv, scene, c = waterBakeIsland(8, "tile/shore-test_au")
	second, err := baker.shade(context.Background(), waterdepth.New(lv, scene), c, "tile/shore-test_au", 4)
	if err != nil {
		t.Fatalf("four-pack working set was unnecessarily reopened: %v", err)
	}
	if first.Full == second.Full {
		t.Fatal("translated shore must retain its distinct native surf pixels")
	}
}

func TestWaterBakeCancellationAndErrors(t *testing.T) {
	dir, lv, scene, _ := waterBakeFixture(t)
	b := blobWriter{root: t.TempDir()}
	baker, err := newWaterBaker(dir, b)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := baker.close(); err != nil {
			t.Error(err)
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := baker.bake(ctx, lv, &BakedLayers{Scene: scene}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
	scene.Cells[0][1].Background = "tile/missing"
	if err := baker.bake(context.Background(), lv, &BakedLayers{Scene: scene}); err == nil {
		t.Fatal("missing background accepted")
	}
}

func waterBakeFixture(t *testing.T) (string, *level.Level, *mapcompose.Scene, map[int]map[string]image.Image) {
	t.Helper()
	dir := t.TempDir()
	lv := &level.Level{Lines: [][]level.Hex{{{Terrain: "waterStones"}, {Terrain: "waterSea"}, {Terrain: "grass"}}, {{Terrain: "waterDeep"}, {Terrain: "waterSea"}}, {{Terrain: "waterSea"}, {Terrain: "waterSea"}, {Terrain: "grass"}}}}
	scene := &mapcompose.Scene{Grid: lv.Grid(), Cells: [][]mapcompose.Cell{{{Background: "tile/unused", Under: "tile/shore-test_au", Cluster: 1}, {Background: "tile/water_au"}, {Background: "tile/grass"}}, {{Background: "tile/water_au"}, {Background: "tile/water_au"}}, {{Background: "tile/water_au"}, {Background: "tile/water_au"}, {Background: "tile/grass"}}}, Clusters: []mapcompose.Cluster{{Cells: []hexgrid.Coord{{}}, Stamp: "stamp/test"}}}
	pictures := map[int]map[string]image.Image{}
	for _, d := range []int{1, 2, 4} {
		packDir := filepath.Join(dir, strconv.Itoa(d)+"x", "backgrounds")
		if err := os.MkdirAll(packDir, 0o755); err != nil {
			t.Fatal(err)
		}
		m := assets.Manifest{Schema: 1, ID: fmt.Sprintf("test/%d", d), Status: "approved"}
		pictures[d] = map[string]image.Image{}
		ids := []string{"tile/water_au", "tile/shore-test_au", "tile/water", "tile/surf", "tile/sea-00", "tile/sea-02", "tile/sea-11", "tile/sea-13", "tile/sea-20", "tile/sea-22", "tile/sea-31", "tile/sea-33"}
		for i, id := range ids {
			im := image.NewNRGBA(image.Rect(0, 0, 96*d, 80*d))
			for y := range im.Rect.Dy() {
				for x := range im.Rect.Dx() {
					p := color.NRGBA{40 + uint8((x+y)%13), 90, 110, 255}
					if id == "tile/shore-test_au" && x < 20*d {
						p = color.NRGBA{140, 130, 70, 255}
					}
					if id == "tile/surf" {
						p = color.NRGBA{180 + uint8(x/d%70), 220 + uint8(y/d%30), 240, 80 + uint8((x/d+y/d)%150)}
					}
					if i >= 4 {
						p = color.NRGBA{25, 70 + uint8(i), 100, 255}
					}
					im.SetNRGBA(x, y, p)
				}
			}
			pictures[d][id] = im
			var raw bytes.Buffer
			if err := png.Encode(&raw, im); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(raw.Bytes())
			name := fmt.Sprintf("%d.png", i)
			m.Entries = append(m.Entries, assets.Entry{ID: id, Path: name, SHA256: hex.EncodeToString(sum[:]), Width: 96 * d, Height: 80 * d, Density: d, Anchor: [2]float64{47.25, 39.75}})
			if err := os.WriteFile(filepath.Join(packDir, name), raw.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		raw, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(packDir, "manifest.json"), raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir, lv, scene, pictures
}
