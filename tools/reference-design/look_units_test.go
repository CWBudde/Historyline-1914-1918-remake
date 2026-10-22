package main

import (
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	"io/fs"
	"math"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func TestUnitContrastComposesBeforeReduction(t *testing.T) {
	background := image.NewNRGBA(image.Rect(0, 0, 200, 200))
	draw.Draw(background, background.Bounds(), image.NewUniform(color.NRGBA{A: 255}), image.Point{}, draw.Src)
	sprite := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	// Three white 2x subpixels give 75% white after composition on black.
	sprite.SetNRGBA(2, 2, color.NRGBA{255, 255, 255, 255})
	sprite.SetNRGBA(3, 2, color.NRGBA{255, 255, 255, 255})
	sprite.SetNRGBA(2, 3, color.NRGBA{255, 255, 255, 255})
	sprite.SetNRGBA(3, 3, color.NRGBA{R: 255, A: 0}) // hiddenRGB mustnot contribute
	before, after, alpha, err := unitPatch(background, sprite, hexgrid.Coord{}, 2, [2]float64{2, 2})
	if err != nil {
		t.Fatal(err)
	}
	b, c := halve(before), halve(after)
	mask, err := averageCanopy(alpha, 2)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for i, a := range mask.alpha {
		if a > .5 {
			found = true
			want := uint8(191)
			if got := c.NRGBAAt(i%c.Bounds().Dx(), i/c.Bounds().Dx()); got != (color.NRGBA{want, want, want, 255}) {
				t.Fatalf("composite-before-reduce: %v", got)
			}
			if got := b.NRGBAAt(i%b.Bounds().Dx(), i/b.Bounds().Dx()); got != (color.NRGBA{A: 255}) {
				t.Fatal(got)
			}
		}
	}
	if !found {
		t.Fatal("missing75%alpha pixel")
	}
}

func TestUnitPatchRespectsOddRowAndFractionalAnchor(t *testing.T) {
	background := image.NewNRGBA(image.Rect(0, 0, 500, 500))
	draw.Draw(background, background.Bounds(), image.White, image.Point{}, draw.Src)
	sprite := image.NewNRGBA(image.Rect(9, 11, 17, 19))
	sprite.SetNRGBA(12, 14, color.NRGBA{A: 255})
	at := hexgrid.Coord{Line: 1, Col: 0}
	anchor := [2]float64{2.25, 3.75}
	before, after, alpha, err := unitPatch(background, sprite, at, 2, anchor)
	if err != nil {
		t.Fatal(err)
	}
	if before.Bounds().Min.X%2 != 0 || before.Bounds().Min.Y%2 != 0 {
		t.Fatal("notworld-aligned")
	}
	x, y := hexgrid.Center(at)
	want := image.Pt(int(math.Round((x-anchor[0])*2))+3, int(math.Round((y-anchor[1])*2))+3)
	if alpha.AlphaAt(want.X, want.Y).A != 255 || after.NRGBAAt(want.X, want.Y) != (color.NRGBA{A: 255}) {
		t.Fatalf("misplaced at%v", want)
	}
	if _, _, _, err := unitPatch(background, sprite, at, 2, [2]float64{1000, 1000}); err == nil {
		t.Fatal("clipped unit accepted")
	}
}

func TestUnitReportCoversCatalogAndBackgrounds(t *testing.T) {
	p := loadFixture(t)
	got, err := p.measureUnits("")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Available || len(got.Sprites) != 144 || len(got.Fixtures) != 4 || len(got.Cases) != 576 {
		t.Fatalf("scope available%v sprites%dfixtures%dcases%d", got.Available, len(got.Sprites), len(got.Fixtures), len(got.Cases))
	}
	seen := map[string]bool{}
	for _, c := range got.Cases {
		key := c.Sprite + "/" + c.Fixture
		if seen[key] {
			t.Fatal("duplicate", key)
		}
		seen[key] = true
		if c.Stats.Samples == 0 || c.Stats.Median == nil || c.Stats.Fraction30 == nil {
			t.Fatalf("unmeasuredcase %+v", c)
		}
	}
	for _, f := range got.Fixtures {
		if len(f.PixelsSHA256) != 64 || len(f.LevelSHA256) != 64 || f.Gaps != 0 || f.Issues != 0 {
			t.Fatalf("fixture provenance %+v", f)
		}
	}
}

func TestUnitPatchRejectsClippedSpriteAndTranslucentGround(t *testing.T) {
	ground := image.NewNRGBA(image.Rect(0, 0, 200, 200))
	draw.Draw(ground, ground.Bounds(), image.White, image.Point{}, draw.Src)
	sprite := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	sprite.SetNRGBA(0, 3, color.NRGBA{A: 255})
	if _, _, _, err := unitPatch(ground, sprite, hexgrid.Coord{}, 2, [2]float64{2, 2}); err == nil {
		t.Fatal("clipped source accepted")
	}
	sprite.SetNRGBA(0, 3, color.NRGBA{})
	sprite.SetNRGBA(3, 3, color.NRGBA{A: 255})
	sprite.SetNRGBA(2, 3, color.NRGBA{A: 255})
	sprite.SetNRGBA(2, 2, color.NRGBA{A: 255})
	sprite.SetNRGBA(3, 2, color.NRGBA{A: 255})
	draw.Draw(ground, ground.Bounds(), image.NewUniform(color.NRGBA{255, 255, 255, 254}), image.Point{}, draw.Src)
	if _, _, err := unitMeasurement(ground, sprite, hexgrid.Coord{}, [2]float64{2, 2}); err == nil {
		t.Fatal("translucent background silently omitted")
	}
}

type unitAssetOverlay struct {
	fs.FS
	files map[string][]byte
}

func (o unitAssetOverlay) Open(name string) (fs.File, error) {
	if b, ok := o.files[name]; ok {
		return fstest.MapFS{name: &fstest.MapFile{Data: b}}.Open(name)
	}
	return o.FS.Open(name)
}

func TestUnitCatalogueRejectsMissingAndChangedSources(t *testing.T) {
	for _, kind := range []string{"missing", "changed"} {
		t.Run(kind, func(t *testing.T) {
			p := loadFixture(t)
			raw, err := fs.ReadFile(p.assets, "index.json")
			if err != nil {
				t.Fatal(err)
			}
			var index struct {
				Schema int         `json:"schema_version"`
				Assets []unitEntry `json:"assets"`
			}
			if err := json.Unmarshal(raw, &index); err != nil {
				t.Fatal(err)
			}
			overlay := unitAssetOverlay{FS: p.assets, files: map[string][]byte{}}
			found := false
			for i, e := range index.Assets {
				if e.Density == 2 && e.ID == "unit/a7v/cp/n-normal" {
					found = true
					if kind == "missing" {
						index.Assets = append(index.Assets[:i], index.Assets[i+1:]...)
						raw, err = json.Marshal(index)
						if err != nil {
							t.Fatal(err)
						}
						overlay.files["index.json"] = raw
					} else {
						overlay.files[e.Path] = []byte("changed")
					}
					break
				}
			}
			if !found {
				t.Fatal("missing test sprite")
			}
			p.assets = overlay
			_, err = p.measureUnits("")
			if err == nil || !strings.Contains(err.Error(), map[string]string{"missing": "missing 2x unit", "changed": "source hash mismatch"}[kind]) {
				t.Fatalf("error%v", err)
			}
		})
	}
}

func TestUnitMeasurementHandlesOpaqueWhite(t *testing.T) {
	for _, whiteGround := range []bool{false, true} {
		ground := image.NewNRGBA(image.Rect(0, 0, 200, 200))
		background, foreground := color.NRGBA{A: 255}, color.NRGBA{255, 255, 255, 255}
		if whiteGround {
			background, foreground = foreground, background
		}
		draw.Draw(ground, ground.Bounds(), image.NewUniform(background), image.Point{}, draw.Src)
		sprite := image.NewNRGBA(image.Rect(0, 0, 8, 8))
		draw.Draw(sprite, image.Rect(2, 2, 6, 6), image.NewUniform(foreground), image.Point{}, draw.Src)
		stats, _, err := unitMeasurement(ground, sprite, hexgrid.Coord{}, [2]float64{2, 2})
		if err != nil {
			t.Fatalf("whiteGround%v: %v", whiteGround, err)
		}
		if stats.Median == nil || math.Abs(*stats.Median-100) > 1e-9 {
			t.Fatalf("white/black contrast %+v", stats)
		}
	}
}
