package maprender

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
)

func lightSource(density, dx, dy int, scale uint8) *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, 16*density, 16*density))
	im.SetNRGBA(density+1+dx, 2*density+1+dy, color.NRGBA{R: 220, B: 70, A: 32 * scale})
	im.SetNRGBA(3*density+1+dx, 2*density+1+dy, color.NRGBA{G: 170, A: 96 * scale})
	return im
}

func lightSetSource(t *testing.T, r *Renderer, files fstest.MapFS, id string, density int, src image.Image) {
	t.Helper()
	e := r.files[fmt.Sprintf("%d/%s", density, id)]
	var raw bytes.Buffer
	if err := png.Encode(&raw, src); err != nil {
		t.Fatal(err)
	}
	files[e.Path].Data = raw.Bytes()
	e.SHA = fmt.Sprintf("%x", sha256.Sum256(raw.Bytes()))
	r.files[fmt.Sprintf("%d/%s", density, id)] = e
	delete(r.cache, e.Path)
}

func lightRenderer(t *testing.T, shadow bool, crownAnchor, shadowAnchor [2]float64) (*Renderer, fstest.MapFS) {
	t.Helper()
	assets := []canopyAsset{{"standing/forest", 16, 16, crownAnchor, 0}}
	if shadow {
		assets = append(assets, canopyAsset{"standing/shadow/forest", 16, 16, shadowAnchor, 0})
	}
	r, files := canopyRenderer(t, assets...)
	for _, density := range []int{2, 4} {
		lightSetSource(t, r, files, "standing/forest", density, lightSource(density, 0, 0, 2))
		if shadow {
			lightSetSource(t, r, files, "standing/shadow/forest", density, lightSource(density, 3*density, 4*density, 1))
		}
	}
	return r, files
}

func lightForest() *level.Level {
	return &level.Level{Lines: [][]level.Hex{{{Terrain: "forest"}}}}
}

func lightNear(got, want float64) bool {
	return math.Abs(got-want) <= 1e-10
}

func lightOnlyPair(t *testing.T, got ShadowPairs) ShadowPair {
	t.Helper()
	if len(got.Pairs) != 1 {
		t.Fatalf("pairs = %+v; want exactly one selected crown", got.Pairs)
	}
	return got.Pairs[0]
}

func TestCastShadowPairsAlphaCentroidTranslation(t *testing.T) {
	for _, density := range []int{2, 4} {
		r, _ := lightRenderer(t, true, [2]float64{}, [2]float64{})
		got, err := r.CastShadowPairs(lightForest(), density)
		if err != nil {
			t.Fatal(err)
		}
		if got.Density != density || len(got.Pairs) != 1 || len(got.Gaps) != 1 {
			t.Fatalf("result = %+v", got)
		}
		pair := lightOnlyPair(t, got)
		if pair.Reason != "" || pair.Standing.ID != "standing/forest" || pair.Shadow.ID != "standing/shadow/forest" || pair.Standing.SHA == "" || pair.Shadow.SHA == "" {
			t.Fatalf("pair = %+v", pair)
		}
		// The shadow has half the crown's opacity, with the same nonuniform
		// weights translated three logical pixels east and four south.
		if pair.Offset != [2]float64{3, 4} || !lightNear(pair.Distance, 5) || pair.Bearing == nil || !lightNear(*pair.Bearing, 143.13010235415598) {
			t.Errorf("translation = %+v", pair)
		}
		if density == 2 && (pair.StandingCentroid != [2]float64{51.25, 42.75} || pair.ShadowCentroid != [2]float64{54.25, 46.75}) {
			t.Errorf("weighted pixel-centre centroids = %+v", pair)
		}
	}
}

func TestCastShadowPairsIgnoresRGBAndMapClipping(t *testing.T) {
	r, files := lightRenderer(t, true, [2]float64{200, 200}, [2]float64{200, 200})
	src := lightSource(2, 6, 8, 1)
	for y := range src.Bounds().Dy() {
		for x := range src.Bounds().Dx() {
			c := src.NRGBAAt(x, y)
			c.R, c.G, c.B = uint8(x*7), uint8(y*5), 255
			src.SetNRGBA(x, y, c)
		}
	}
	lightSetSource(t, r, files, "standing/shadow/forest", 2, src)
	got, err := r.CastShadowPairs(lightForest(), 2)
	if err != nil {
		t.Fatal(err)
	}
	pair := lightOnlyPair(t, got)
	if pair.Reason != "" || pair.Offset != [2]float64{3, 4} || pair.StandingCentroid != [2]float64{-148.75, -157.25} {
		t.Fatalf("off-canvas source contribution = %+v", pair)
	}
}

func TestCastShadowPairsNonzeroBoundsOddRowAndUnequalAnchors(t *testing.T) {
	r, files := lightRenderer(t, true, [2]float64{1.25, 2.25}, [2]float64{0.75, 1.75})
	crown := image.NewNRGBA(image.Rect(10, 20, 18, 28))
	crown.SetNRGBA(12, 23, color.NRGBA{A: 128})
	shadow := image.NewNRGBA(image.Rect(-5, -8, 3, 0))
	shadow.SetNRGBA(-3, -5, color.NRGBA{A: 64})
	for id, src := range map[string]image.Image{"standing/forest": crown, "standing/shadow/forest": shadow} {
		lightSetSource(t, r, files, id, 2, src)
		// Decode and verify the PNG, then retain the equivalent source with its
		// nonzero image bounds to exercise the same cache contract as Place.
		e, _, err := r.assetImage(id, 2)
		if err != nil {
			t.Fatal(err)
		}
		r.cache[e.Path] = src
	}
	lvl := &level.Level{Lines: [][]level.Hex{{{Terrain: "grass"}, {Terrain: "grass"}}, {{Terrain: "forest"}}}}
	got, err := r.CastShadowPairs(lvl, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Pairs) != 1 {
		t.Fatalf("pairs = %+v", got.Pairs)
	}
	pair := lightOnlyPair(t, got)
	if pair.At != (hexgrid.Coord{Line: 1, Col: 0}) || pair.StandingCentroid != [2]float64{119.25, 79.75} || pair.ShadowCentroid != [2]float64{119.75, 80.25} || pair.Offset != [2]float64{0.5, 0.5} || pair.Bearing == nil || !lightNear(*pair.Bearing, 135) {
		t.Fatalf("odd-row placed centroids = %+v", pair)
	}
}

func TestCastShadowPairsCompassBearings(t *testing.T) {
	for _, tc := range []struct{ dx, dy, bearing float64 }{
		{0, -1, 0}, {1, -1, 45}, {1, 0, 90}, {1, 1, 135},
		{0, 1, 180}, {-1, 1, 225}, {-1, 0, 270}, {-1, -1, 315},
	} {
		r, files := lightRenderer(t, true, [2]float64{}, [2]float64{-tc.dx, -tc.dy})
		lightSetSource(t, r, files, "standing/shadow/forest", 2, lightSource(2, 0, 0, 1))
		got, err := r.CastShadowPairs(lightForest(), 2)
		if err != nil {
			t.Fatal(err)
		}
		pair := lightOnlyPair(t, got)
		if pair.Bearing == nil || !lightNear(*pair.Bearing, tc.bearing) {
			t.Errorf("offset (%v,%v) bearing = %v, want %v", tc.dx, tc.dy, pair.Bearing, tc.bearing)
		}
	}
}

func TestCastShadowPairsExclusions(t *testing.T) {
	for _, tc := range []struct{ name, reason string }{
		{"missing", "no_selected_cast_shadow"}, {"empty crown", "empty_standing_alpha"}, {"empty shadow", "empty_shadow_alpha"},
		{"clipped crown", "clipped_standing_alpha"}, {"clipped shadow", "clipped_shadow_alpha"}, {"zero", "zero_offset"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, files := lightRenderer(t, tc.name != "missing", [2]float64{}, [2]float64{})
			switch tc.name {
			case "empty crown", "empty shadow":
				id := "standing/forest"
				if tc.name == "empty shadow" {
					id = "standing/shadow/forest"
				}
				lightSetSource(t, r, files, id, 2, image.NewNRGBA(image.Rect(0, 0, 32, 32)))
			case "clipped crown", "clipped shadow":
				id := "standing/forest"
				if tc.name == "clipped shadow" {
					id = "standing/shadow/forest"
				}
				src := lightSource(2, 0, 0, 1)
				src.SetNRGBA(0, 12, color.NRGBA{A: 1})
				lightSetSource(t, r, files, id, 2, src)
			case "zero":
				lightSetSource(t, r, files, "standing/shadow/forest", 2, lightSource(2, 0, 0, 1))
			}
			got, err := r.CastShadowPairs(lightForest(), 2)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Pairs) != 1 || got.Pairs[0].Reason != tc.reason || got.Pairs[0].Bearing != nil {
				t.Fatalf("excluded pair = %+v; want reason %s and nil bearing", got.Pairs, tc.reason)
			}
			if got.Pairs[0].Standing.SHA == "" || got.Pairs[0].Shadow.ID != "standing/shadow/forest" {
				t.Fatalf("excluded source metadata = %+v", got.Pairs[0])
			}
		})
	}
}

func TestCastShadowPairsUsesSelectedContextAndRejectsOrphans(t *testing.T) {
	d, _ := mapcompose.Descriptor(mapcompose.Forest, "forest", [6]string{"forest", "forest", "forest", "forest", "forest", "forest"}, [2]int{})
	ctx := mapcompose.ContextID(mapcompose.Forest, mapcompose.Key(d))
	for _, missing := range []bool{false, true} {
		assets := []canopyAsset{
			{ctx, 16, 16, [2]float64{}, 0}, {"standing/" + ctx, 16, 16, [2]float64{}, 0},
			{"standing/forest", 16, 16, [2]float64{}, 0}, {"standing/shadow/forest", 16, 16, [2]float64{}, 0},
			{"standing/shadow/orphan", 16, 16, [2]float64{}, 0},
		}
		if !missing {
			assets = append(assets, canopyAsset{mapcompose.ShadowID("standing/" + ctx), 16, 16, [2]float64{}, 0})
		}
		r, files := canopyRenderer(t, assets...)
		lightSetSource(t, r, files, "standing/"+ctx, 2, lightSource(2, 0, 0, 2))
		if !missing {
			lightSetSource(t, r, files, mapcompose.ShadowID("standing/"+ctx), 2, lightSource(2, 6, 8, 1))
		}
		got, err := r.CastShadowPairs(lightForest(), 2)
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Pairs) != 1 || got.Pairs[0].Standing.ID != "standing/"+ctx || got.Pairs[0].Shadow.ID != mapcompose.ShadowID("standing/"+ctx) || len(got.Gaps) != 0 {
			t.Fatalf("selected context = %+v", got)
		}
		if missing && got.Pairs[0].Reason != "no_selected_cast_shadow" || !missing && got.Pairs[0].Reason != "" {
			t.Errorf("missing=%v pair = %+v", missing, got.Pairs[0])
		}
	}
}

func TestCastShadowPairsReportsNoncanopyAndMissingCrown(t *testing.T) {
	r, _ := canopyRenderer(t, canopyAsset{"standing/forest_w", 16, 16, [2]float64{}, 0}, canopyAsset{"standing/house", 16, 16, [2]float64{}, 0})
	lvl := &level.Level{Lines: [][]level.Hex{{{Terrain: "forest"}, {Terrain: "forest_w"}, {Terrain: "house"}, {Terrain: "hqN"}}}}
	got, err := r.CastShadowPairs(lvl, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Pairs) != 0 || len(got.Unmeasured) != 4 || len(got.Gaps) != 2 || len(got.Issues) != 1 {
		t.Fatalf("omissions = %+v", got)
	}
	want := []ShadowOmission{
		{At: hexgrid.Coord{Line: 0, Col: 0}, ID: "tile/forest", Reason: "no_selected_crown"},
		{At: hexgrid.Coord{Line: 0, Col: 1}, ID: "standing/forest_w", Reason: "non_canopy_standing"},
		{At: hexgrid.Coord{Line: 0, Col: 1}, ID: "tile/forest_w", Reason: "no_selected_crown"},
		{At: hexgrid.Coord{Line: 0, Col: 2}, ID: "standing/house", Reason: "non_canopy_standing"},
	}
	for i := range want {
		if got.Unmeasured[i] != want[i] {
			t.Errorf("omission %d = %+v, want %+v", i, got.Unmeasured[i], want[i])
		}
	}
}

func TestCastShadowPairsFailsClosedOnDensityAnchorsAndAssets(t *testing.T) {
	for _, which := range []string{"standing/forest", "standing/shadow/forest"} {
		for _, failure := range []string{"density", "missing", "hash", "decode", "anchor NaN", "anchor Inf"} {
			t.Run(which+"/"+failure, func(t *testing.T) {
				r, files := lightRenderer(t, true, [2]float64{}, [2]float64{})
				density := 2
				e := r.files["2/"+which]
				want := failure
				switch failure {
				case "density":
					density = 1
				case "missing":
					delete(files, e.Path)
					want = "file does not exist"
				case "hash":
					files[e.Path].Data = []byte("broken")
					want = "sha256"
				case "decode":
					files[e.Path].Data = []byte("invalid PNG")
					e.SHA = fmt.Sprintf("%x", sha256.Sum256(files[e.Path].Data))
					r.files["2/"+which] = e
					want = "PNG"
				case "anchor NaN", "anchor Inf":
					if failure == "anchor Inf" {
						e.Anchor[1] = math.Inf(1)
					} else {
						e.Anchor[0] = math.NaN()
					}
					r.files["2/"+which] = e
					want = "anchor"
				}
				got, err := r.CastShadowPairs(lightForest(), density)
				if err == nil || !strings.Contains(err.Error(), want) || len(got.Pairs) != 0 {
					t.Fatalf("result %+v, error %v; want %s error without partial pairs", got, err, want)
				}
			})
		}
	}
}

// Scattered crowns are measured one by one with their own shadows, each placed
// at its site's offset from the hex centre.
func TestCastShadowPairsMeasuresScatteredCrowns(t *testing.T) {
	var assets []canopyAsset
	for _, sp := range mapcompose.CrownSpecies {
		for _, size := range mapcompose.CrownSizes {
			id := mapcompose.CrownID(sp, size)
			assets = append(assets, canopyAsset{id, 16, 16, [2]float64{8, 8}, 0}, canopyAsset{mapcompose.ShadowID(id), 16, 16, [2]float64{8, 8}, 0})
		}
	}
	r, files := canopyRenderer(t, assets...)
	for _, a := range assets {
		for _, density := range []int{2, 4} {
			if strings.HasPrefix(a.id, "standing/shadow/") {
				lightSetSource(t, r, files, a.id, density, lightSource(density, 3*density, 4*density, 1))
			} else {
				lightSetSource(t, r, files, a.id, density, lightSource(density, 0, 0, 2))
			}
		}
	}
	lvl := lightForest()
	scene := mapcompose.Compose(lvl.Grid(), func(hexgrid.Coord) mapcompose.Hex { return mapcompose.Hex{Terrain: "forest"} }, func(id string) bool { return r.has[id] })
	sites := scene.ScatterAt(hexgrid.Coord{}, func(*mapcompose.Cluster) bool { return true })
	if len(sites) == 0 {
		t.Fatal("no scattered crowns in the test forest")
	}
	// A cut-out's alpha may reach its canvas edge.
	edge := lightSource(2, 0, 0, 2)
	edge.SetNRGBA(0, 12, color.NRGBA{A: 1})
	lightSetSource(t, r, files, sites[0].ID, 2, edge)
	for _, density := range []int{2, 4} {
		got, err := r.CastShadowPairs(lvl, density)
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Pairs) != len(sites) || len(got.Unmeasured) != 0 {
			t.Fatalf("%dx: %d pairs, %d unmeasured; want %d scattered crowns", density, len(got.Pairs), len(got.Unmeasured), len(sites))
		}
		for i, pair := range got.Pairs {
			if pair.Reason != "" || pair.Standing.ID != sites[i].ID || pair.Shadow.ID != mapcompose.ShadowID(sites[i].ID) || pair.Site != [2]float64{sites[i].X, sites[i].Y} {
				t.Fatalf("%dx pair %d = %+v; want site %+v", density, i, pair, sites[i])
			}
			if pair.Standing.ID == sites[0].ID && density == 2 {
				continue // the edge texel moves its centroid
			}
			if pair.Offset != [2]float64{3, 4} || pair.Bearing == nil || !lightNear(*pair.Bearing, 143.13010235415598) {
				t.Errorf("%dx pair %d offset %v", density, i, pair.Offset)
			}
			// The crown sits at its site with anchor (8,8), rounded to the density
			// grid like PlaceAt; lightSource weighs its pixel centres to
			// (2.5, 2) + 1.5/density logical px from the canvas origin.
			x, y := hexgrid.Center(hexgrid.Coord{})
			d := float64(density)
			want := [2]float64{math.Round((x+sites[i].X-8)*d)/d + 2.5 + 1.5/d, math.Round((y+sites[i].Y-8)*d)/d + 2 + 1.5/d}
			if !lightNear(pair.StandingCentroid[0], want[0]) || !lightNear(pair.StandingCentroid[1], want[1]) {
				t.Errorf("%dx pair %d crown centroid %v, want %v", density, i, pair.StandingCentroid, want)
			}
		}
	}
}
