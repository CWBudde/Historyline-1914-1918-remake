package maprender

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
)

type canopyAsset struct {
	id     string
	w, h   int
	anchor [2]float64
	alpha  uint8
}

// Real PNG bytes and catalogue hashes exercise the same loading path as Render.
func canopyRenderer(t *testing.T, assets ...canopyAsset) (*Renderer, fstest.MapFS) {
	t.Helper()
	files := fstest.MapFS{}
	var entries []Entry
	for i, asset := range assets {
		for _, density := range []int{2, 4} {
			im := image.NewNRGBA(image.Rect(0, 0, asset.w*density, asset.h*density))
			draw.Draw(im, im.Bounds(), image.NewUniform(color.NRGBA{R: 210, G: 23, B: 171, A: asset.alpha}), image.Point{}, draw.Src)
			var raw bytes.Buffer
			if err := png.Encode(&raw, im); err != nil {
				t.Fatal(err)
			}
			path := fmt.Sprintf("test/%d-%d.png", i, density)
			files[path] = &fstest.MapFile{Data: raw.Bytes()}
			entries = append(entries, Entry{ID: asset.id, Pack: fmt.Sprintf("terrain-test-%dx", density), Density: density, Path: path, SHA: fmt.Sprintf("%x", sha256.Sum256(raw.Bytes())), Anchor: asset.anchor})
		}
	}
	raw, err := json.Marshal(struct {
		Schema int     `json:"schema_version"`
		Assets []Entry `json:"assets"`
	}{1, entries})
	if err != nil {
		t.Fatal(err)
	}
	files["index.json"] = &fstest.MapFile{Data: raw}
	r, err := New(files)
	if err != nil {
		t.Fatal(err)
	}
	return r, files
}

// A hex mask or default forest sprite must not replace selected context alpha.
func TestVisibleCanopyUsesSelectedContextAndOverhang(t *testing.T) {
	d, _ := mapcompose.Descriptor(mapcompose.Forest, "forest", [6]string{"forest", "forest", "forest", "forest", "forest", "forest"}, [2]int{})
	ctx := mapcompose.ContextID(mapcompose.Forest, mapcompose.Key(d))
	for _, density := range []int{2, 4} {
		t.Run(fmt.Sprint(density), func(t *testing.T) {
			r, _ := canopyRenderer(t,
				canopyAsset{ctx, 1, 1, [2]float64{}, 255},
				canopyAsset{"standing/" + ctx, 128, 112, [2]float64{64, 56}, 128},
				canopyAsset{"standing/forest", 1, 1, [2]float64{}, 255},
			)
			got, err := r.VisibleCanopy(&level.Level{Lines: [][]level.Hex{{{Terrain: "forest"}}}}, density)
			if err != nil {
				t.Fatal(err)
			}
			if got.Coverage == nil {
				t.Fatal("no canopy plane")
			}
			if got.Coverage.Bounds() != image.Rect(0, 0, 96*density, 80*density) {
				t.Fatalf("bounds %v", got.Coverage.Bounds())
			}
			// Whole selected artwork extends beyond the hex polygon at this map corner.
			for _, p := range []image.Point{{0, 0}, {48 * density, 40 * density}, {95 * density, 79 * density}} {
				if a := got.Coverage.AlphaAt(p.X, p.Y).A; a != 128 {
					t.Errorf("alpha at %v = %d; want128", p, a)
				}
			}
			if len(got.Unsupported) != 0 || len(got.Gaps) != 0 {
				t.Fatalf("unsupported %v gaps %v", got.Unsupported, got.Gaps)
			}
		})
	}
}

// Wrong per-cell ordering, counting shadows, or summing alpha breaks these values.
func TestVisibleCanopyUnionOcclusionAndShadows(t *testing.T) {
	for _, tc := range []struct {
		name, second string
		want         uint8
	}{
		{"crowns union", "forest", 192},
		{"later object occludes", "house", 64},
		{"shadow does not count", "grass", 128},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := canopyRenderer(t,
				canopyAsset{"standing/forest", 150, 4, [2]float64{}, 128},
				canopyAsset{"standing/house", 4, 4, [2]float64{}, 128},
				canopyAsset{"standing/shadow/forest", 240, 80, [2]float64{48, 40}, 255},
			)
			lvl := &level.Level{Lines: [][]level.Hex{{{Terrain: "forest"}, {Terrain: tc.second}}}}
			got, err := r.VisibleCanopy(lvl, 2)
			if err != nil {
				t.Fatal(err)
			}
			if got.Coverage == nil {
				t.Fatal("no canopy plane")
			}
			if a := got.Coverage.AlphaAt(190*2, 40*2).A; a != tc.want {
				t.Errorf("overlap alpha %d; want%d", a, tc.want)
			}
			if a := got.Coverage.AlphaAt(100*2, 60*2).A; a != 0 {
				t.Errorf("shadow became canopy: %d", a)
			}
		})
	}
}

func TestVisibleCanopyOddLineAnchorAndLaterCrown(t *testing.T) {
	r, _ := canopyRenderer(t,
		canopyAsset{"standing/house", 200, 120, [2]float64{48, 40}, 255},
		canopyAsset{"standing/forest", 2, 2, [2]float64{1.25, 1.25}, 128},
	)
	lvl := &level.Level{Lines: [][]level.Hex{{{Terrain: "house"}, {Terrain: "house"}}, {{Terrain: "forest"}}}}
	got, err := r.VisibleCanopy(lvl, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got.Coverage == nil {
		t.Fatal("no canopy plane")
	}
	// Odd-line centre(119,80), anchor(1.25,1.25), rounded placement(236,158).
	if a := got.Coverage.AlphaAt(236, 158).A; a != 128 {
		t.Errorf("odd-line crown alpha%d; want128", a)
	}
	if a := got.Coverage.AlphaAt(235, 158).A; a != 0 {
		t.Errorf("anchor shifted left: %d", a)
	}
}

func TestVisibleCanopyMeadowProps(t *testing.T) {
	var assets []canopyAsset
	for i := 1; i <= 4; i++ {
		assets = append(assets, canopyAsset{fmt.Sprintf("standing/props/trees/%d", i), 2, 2, [2]float64{}, 255})
	}
	r, _ := canopyRenderer(t, assets...)
	lvl := &level.Level{Lines: [][]level.Hex{make([]level.Hex, 30)}}
	for col := range lvl.Lines[0] {
		lvl.Lines[0][col].Terrain = "grass"
	}
	got, err := r.VisibleCanopy(lvl, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got.Coverage == nil {
		t.Fatal("no canopy plane")
	}
	count := 0
	for _, a := range got.Coverage.Pix {
		if a != 0 {
			count++
		}
	}
	if count == 0 {
		t.Fatal("selected meadow props omitted")
	}
}

// Transparent gaps and faint leaves must survive; neither RGB nor sprite bounds
// establishes visible crown coverage.
func TestVisibleCanopyRetainsSparseAuthoredAlpha(t *testing.T) {
	r, files := canopyRenderer(t, canopyAsset{"standing/forest", 4, 4, [2]float64{}, 255})
	src := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	src.SetNRGBA(1, 1, color.NRGBA{R: 255, B: 255, A: 37})
	src.SetNRGBA(6, 6, color.NRGBA{A: 255})
	var raw bytes.Buffer
	if err := png.Encode(&raw, src); err != nil {
		t.Fatal(err)
	}
	files["test/0-2.png"].Data = raw.Bytes()
	e := r.files["2/standing/forest"]
	e.SHA = fmt.Sprintf("%x", sha256.Sum256(raw.Bytes()))
	r.files["2/standing/forest"] = e
	got, err := r.VisibleCanopy(&level.Level{Lines: [][]level.Hex{{{Terrain: "forest"}}}}, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		x, y int
		want uint8
	}{{97, 81, 37}, {102, 86, 255}, {100, 84, 0}} {
		if a := got.Coverage.AlphaAt(tc.x, tc.y).A; a != tc.want {
			t.Errorf("alpha at (%d,%d)=%d; want%d", tc.x, tc.y, a, tc.want)
		}
	}
}

func TestVisibleCanopyReportsUnsupportedForest(t *testing.T) {
	r, _ := canopyRenderer(t, canopyAsset{"standing/forest_w", 96, 80, [2]float64{48, 40}, 255})
	lvl := &level.Level{Lines: [][]level.Hex{{{Terrain: "forest"}, {Terrain: "forest_w"}, {Terrain: "grass"}, {Terrain: "hqN"}}}}
	got, err := r.VisibleCanopy(lvl, 2)
	if err != nil {
		t.Fatal(err)
	}
	want := []hexgrid.Coord{{Line: 0, Col: 0}, {Line: 0, Col: 1}}
	if !slices.Equal(got.Unsupported, want) {
		t.Errorf("unsupported %v; want%v", got.Unsupported, want)
	}
	if len(got.Gaps) != 2 {
		t.Errorf("composition gaps%d; want2", len(got.Gaps))
	}
	if len(got.Issues) != 1 || got.Issues[0].At != (hexgrid.Coord{Line: 0, Col: 3}) {
		t.Errorf("incomplete compound issues %v", got.Issues)
	}
}

func TestVisibleCanopyRejectsInvalidDensityAndAssets(t *testing.T) {
	lvl := &level.Level{Lines: [][]level.Hex{{{Terrain: "forest"}}}}
	for _, tc := range []struct {
		name    string
		density int
		failure string
	}{
		{"density", 1, "density"}, {"missing", 2, "file does not exist"}, {"hash", 2, "sha256"}, {"decode", 2, "PNG"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, files := canopyRenderer(t, canopyAsset{"standing/forest", 2, 2, [2]float64{}, 255})
			switch tc.name {
			case "missing":
				delete(files, "test/0-2.png")
			case "hash":
				files["test/0-2.png"].Data = []byte("broken")
			case "decode":
				files["test/0-2.png"].Data = []byte("not a PNG image")
				e := r.files["2/standing/forest"]
				e.SHA = fmt.Sprintf("%x", sha256.Sum256([]byte("not a PNG image")))
				r.files["2/standing/forest"] = e
			}
			_, err := r.VisibleCanopy(lvl, tc.density)
			if err == nil || !strings.Contains(err.Error(), tc.failure) {
				t.Fatalf("error %v; want%q", err, tc.failure)
			}
		})
	}
}
