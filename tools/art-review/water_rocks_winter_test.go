package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/terraingrade"
)

// The summer lake rocks stay as they are.
var summerRocksSHA = map[string]string{
	"terrain/water/waterStones-2x.png": "9e032238a22e7d5409614bdc1130def60631f704082ecc97ccdff8d798e97c2e",
	"terrain/water/waterStones-4x.png": "6382c845eeea54f5df7896855eb874eff2e7304df7bbf62696a7499e1661dc87",
	"terrain/water/waterRocks-2x.png":  "01eadc582ceae7374d94de74e2d179804f7c914901cfd6cd5502285a1dcba864",
	"terrain/water/waterRocks-4x.png":  "a82af40d9201ea56c1af155a24321857bc55d893b1113704b0cca2e4a8cd051d",
}

// installedRuntimeTile loads the installed runtime file of asset id at density d.
func installedRuntimeTile(t *testing.T, root, id string, d int) *image.NRGBA {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "assets/index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var index struct {
		Assets []struct {
			ID      string `json:"id"`
			Density int    `json:"density"`
			Path    string `json:"path"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(b, &index); err != nil {
		t.Fatal(err)
	}
	for _, e := range index.Assets {
		if e.ID == id && e.Density == d {
			im, err := loadPNG(filepath.Join(root, "assets", e.Path))
			if err != nil {
				t.Fatal(err)
			}
			return im
		}
	}
	t.Fatalf("%s at %dx is not installed", id, d)
	return nil
}

// TestWaterRocksWinterSnowCaps checks the winter lake rocks at 4x, graded as the
// pack build grades them (terraingrade, winter water): the summer rocks' outline
// with settled snow on their tops, no summer green, and the water around them
// as the winter water tile has it.
func TestWaterRocksWinterSnowCaps(t *testing.T) {
	root := filepath.Join("..", "..")
	for path, want := range summerRocksSHA {
		b, err := os.ReadFile(filepath.Join(root, "assets", path))
		if err != nil {
			t.Fatal(err)
		}
		if sum := sha256.Sum256(b); hex.EncodeToString(sum[:]) != want {
			t.Errorf("summer %s changed", path)
		}
	}
	const d = 4
	summerRocks := installedRuntimeTile(t, root, "tile/waterRocks", d)
	summerStones := installedRuntimeTile(t, root, "tile/waterStones", d)
	rocks := installedRuntimeTile(t, root, "tile/waterRocks_w", d)
	stones := installedRuntimeTile(t, root, "tile/waterStones_w", d)
	water := installedRuntimeTile(t, root, "tile/water_w", d)
	if got := compoundCompose(water, rocks); !bytes.Equal(got.Pix, stones.Pix) {
		t.Error("the winter rocks over tile/water_w differ from tile/waterStones_w")
	}
	r := rocks.Rect
	solid := func(im *image.NRGBA, x, y int) bool { return im.NRGBAAt(x, y).A >= 128 }

	// The outline: alpha >= 128 against the summer rocks.
	inter, union := 0, 0
	for y := range r.Dy() {
		for x := range r.Dx() {
			a, b := solid(summerRocks, x, y), solid(rocks, x, y)
			if a && b {
				inter++
			}
			if a || b {
				union++
			}
		}
	}
	iou := float64(inter) / float64(union)
	t.Logf("rock mask IoU %.4f (%d/%d)", iou, inter, union)
	if iou < .97 {
		t.Errorf("rock mask IoU %.4f, want >= 0.97", iou)
	}

	// Snow and green on the rock area (the summer rocks at alpha >= 200).
	type gradedTile struct {
		id string
		im *image.NRGBA
	}
	for _, g := range []gradedTile{
		{"tile/waterRocks_w", terraingrade.Grade(rocks, "tile/waterRocks_w", "terrain/water", nil)},
		{"tile/waterStones_w", terraingrade.Grade(stones, "tile/waterStones_w", "terrain/water", nil)},
	} {
		area, snow, green := 0, 0, 0
		var ax, ay, sx, sy float64
		for y := range r.Dy() {
			for x := range r.Dx() {
				if summerRocks.NRGBAAt(x, y).A < 200 {
					continue
				}
				area++
				ax, ay = ax+float64(x), ay+float64(y)
				c := g.im.NRGBAAt(x, y)
				l, a, b := winterLab(c)
				if l >= 75 && math.Hypot(a, b) <= 10 {
					snow++
					sx, sy = sx+float64(x), sy+float64(y)
				}
				if summerGreen(c) {
					green++
				}
			}
		}
		snowShare, greenShare := float64(snow)/float64(area), float64(green)/float64(area)
		t.Logf("%s: rock area %d px, snow %.1f %%, summer green %.2f %%", g.id, area, 100*snowShare, 100*greenShare)
		if snowShare < .25 {
			t.Errorf("%s: snow (L* >= 75, C* <= 10) on %.1f %% of the rock area, want >= 25 %%", g.id, 100*snowShare)
		}
		if greenShare >= .01 {
			t.Errorf("%s: summer green on %.2f %% of the rock area, want < 1 %%", g.id, 100*greenShare)
		}
		if snow > 0 {
			// The snow lies on the lit side: its centre is up-left of the rocks'.
			dx, dy := sx/float64(snow)-ax/float64(area), sy/float64(snow)-ay/float64(area)
			t.Logf("%s: snow centre offset %.1f,%.1f px", g.id, dx, dy)
			if dx+dy > 0 {
				t.Errorf("%s: snow centre lies %.1f,%.1f px toward the lower right of the rocks", g.id, dx, dy)
			}
		}
	}

	// The water: outside both rock masks dilated by 2 world px, the winter
	// stones tile is the winter water tile's (the summer stones' pixels).
	reach := 2 * d
	near := func(x, y int) bool {
		for v := max(0, y-reach); v <= min(r.Dy()-1, y+reach); v++ {
			for u := max(0, x-reach); u <= min(r.Dx()-1, x+reach); u++ {
				if (u-x)*(u-x)+(v-y)*(v-y) <= reach*reach && (solid(summerRocks, u, v) || solid(rocks, u, v)) {
					return true
				}
			}
		}
		return false
	}
	gs := terraingrade.Grade(stones, "tile/waterStones_w", "terrain/water", nil)
	gw := terraingrade.Grade(summerStones, "tile/waterStones_w", "terrain/water", nil)
	var sum float64
	n := 0
	for y := range r.Dy() {
		for x := range r.Dx() {
			if stones.NRGBAAt(x, y).A != 255 || near(x, y) {
				continue
			}
			p, q := gs.NRGBAAt(x, y), gw.NRGBAAt(x, y)
			sum += math.Abs(float64(p.R)-float64(q.R)) + math.Abs(float64(p.G)-float64(q.G)) + math.Abs(float64(p.B)-float64(q.B))
			n += 3
		}
	}
	mad := sum / float64(n)
	t.Logf("water outside the rocks: mean |diff| %.3f/255 over %d px", mad, n/3)
	if mad > 1 {
		t.Errorf("water outside the rocks differs by %.3f/255 on average, want <= 1/255", mad)
	}
}
