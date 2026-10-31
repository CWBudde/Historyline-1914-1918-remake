package artreview

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/draw"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/terraingrade"
)

// The summer skerries and the shoals of both seasons stay as they are.
var summerSeaKitSHA = map[string]string{
	"terrain/water/skerry1-2x.png":  "12b3cb3087f63f509bc16a674d1a1d379327679195ea0250c6d0f08240c08be9",
	"terrain/water/skerry1-4x.png":  "67d0eb3235d75ca5165a7599c1c4bd09190f9c972bbfdba045d045a239558360",
	"terrain/water/skerry2-2x.png":  "45bcbcae6211122a0e0a597ae4e3baad83d9cd64578f6690f697f3e39f6a0a51",
	"terrain/water/skerry2-4x.png":  "ca5fd5187fe850e02787322f6fd7fcd4296fab8502317f2589d78fac1709d1a8",
	"terrain/water/skerry3-2x.png":  "351d7719169d6725f9583f9ee0bb988aef1c588ecebe9c11acee65bf7c3958a0",
	"terrain/water/skerry3-4x.png":  "36c792a007ebb0c1551f8eb18210c41f781d44ca4be885b68e5606fb06a16682",
	"terrain/water/shoal1-2x.png":   "188b110415e568cce5d2a1b5b9208e8cbd77bf057fe72a5714c7975acc688a3b",
	"terrain/water/shoal1-4x.png":   "588cc38b6d51d61f2afea6363f7642f5ebb1cd2c6721c5872dc2693809ae2586",
	"terrain/water/shoal2-2x.png":   "79624bcf446a756665afad8cca9dc8f01e8a3a34ceea1d93974a7beb40e3dc4e",
	"terrain/water/shoal2-4x.png":   "bd11ad313a493800835ca4d618fdc9d0f2df1088164d715e6e2d1b026f6751ff",
	"terrain/water/shoal3-2x.png":   "256e24ab613dac731f8c13f04cce13b25b6a2a0cca603614eede0578b7ac28a8",
	"terrain/water/shoal3-4x.png":   "36e2d687753e5bfb6fb4574590a037630f7290933429ab68139de8aebb472711",
	"terrain/water/shoal4-2x.png":   "e7ae0c6fb209f3eb5284272098958153bc790c0df54ba9e72fc1003261eb52b0",
	"terrain/water/shoal4-4x.png":   "ad4c14bc3aa1b0da4435f268984dc300e1a960153c0a0f68b78cc80493fa4112",
	"terrain/water/shoal1_w-2x.png": "188b110415e568cce5d2a1b5b9208e8cbd77bf057fe72a5714c7975acc688a3b",
	"terrain/water/shoal1_w-4x.png": "588cc38b6d51d61f2afea6363f7642f5ebb1cd2c6721c5872dc2693809ae2586",
	"terrain/water/shoal2_w-2x.png": "79624bcf446a756665afad8cca9dc8f01e8a3a34ceea1d93974a7beb40e3dc4e",
	"terrain/water/shoal2_w-4x.png": "bd11ad313a493800835ca4d618fdc9d0f2df1088164d715e6e2d1b026f6751ff",
	"terrain/water/shoal3_w-2x.png": "256e24ab613dac731f8c13f04cce13b25b6a2a0cca603614eede0578b7ac28a8",
	"terrain/water/shoal3_w-4x.png": "36e2d687753e5bfb6fb4574590a037630f7290933429ab68139de8aebb472711",
	"terrain/water/shoal4_w-2x.png": "e7ae0c6fb209f3eb5284272098958153bc790c0df54ba9e72fc1003261eb52b0",
	"terrain/water/shoal4_w-4x.png": "ad4c14bc3aa1b0da4435f268984dc300e1a960153c0a0f68b78cc80493fa4112",
}

// TestSeaKitWinterSnowCaps checks the winter skerries at 4x, each graded as the
// pack build grades it (terraingrade) and drawn over the graded winter water:
// the summer skerries' outline with settled snow on the rocks, no summer
// green, and the summer skerries and the shoals unchanged.
func TestSeaKitWinterSnowCaps(t *testing.T) {
	root := filepath.Join("..", "..")
	for path, want := range summerSeaKitSHA {
		b, err := os.ReadFile(filepath.Join(root, "assets", path))
		if err != nil {
			t.Fatal(err)
		}
		if sum := sha256.Sum256(b); hex.EncodeToString(sum[:]) != want {
			t.Errorf("%s changed", path)
		}
	}
	const d = 4
	water := terraingrade.Grade(installedRuntimeTile(t, root, "tile/water_w", d), "tile/water_w", "terrain/water", nil)
	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("tile/skerry%d", i)
		summer := installedRuntimeTile(t, root, id, d)
		winter := installedRuntimeTile(t, root, id+"_w", d)
		r := summer.Rect

		// The outline: alpha >= 128 against the summer skerry.
		inter, union := 0, 0
		for y := range r.Dy() {
			for x := range r.Dx() {
				a, b := summer.NRGBAAt(x, y).A >= 128, winter.NRGBAAt(x, y).A >= 128
				if a && b {
					inter++
				}
				if a || b {
					union++
				}
			}
		}
		iou := float64(inter) / float64(union)
		t.Logf("%s_w: alpha IoU %.4f (%d/%d)", id, iou, inter, union)
		if iou < .97 {
			t.Errorf("%s_w: alpha IoU %.4f, want >= 0.97", id, iou)
		}

		// Snow and green on the rock area: the summer skerry's opaque
		// (alpha >= 200) pixels of rock and weed hue (b* >= 0, hue under 150°),
		// so the blue and teal water and the foam around the rocks do not count.
		over := image.NewNRGBA(r)
		draw.Draw(over, r, water, image.Point{}, draw.Src)
		draw.Draw(over, r, terraingrade.Grade(winter, id+"_w", "terrain/water", nil), image.Point{}, draw.Over)
		area, snow, green := 0, 0, 0
		for y := range r.Dy() {
			for x := range r.Dx() {
				s := summer.NRGBAAt(x, y)
				if _, a, b := winterLab(s); s.A < 200 || b < 0 || math.Atan2(b, a) >= 150*math.Pi/180 {
					continue
				}
				area++
				c := over.NRGBAAt(x, y)
				l, a, b := winterLab(c)
				if l >= 75 && math.Hypot(a, b) <= 10 {
					snow++
				}
				if summerGreen(c) {
					green++
				}
			}
		}
		if area == 0 {
			t.Fatalf("%s: no rock area", id)
		}
		snowShare, greenShare := float64(snow)/float64(area), float64(green)/float64(area)
		t.Logf("%s_w: rock area %d px, snow %.1f %%, summer green %.2f %%", id, area, 100*snowShare, 100*greenShare)
		if snowShare < .25 {
			t.Errorf("%s_w: snow (L* >= 75, C* <= 10) on %.1f %% of the rock area, want >= 25 %%", id, 100*snowShare)
		}
		if greenShare >= .01 {
			t.Errorf("%s_w: summer green on %.2f %% of the rock area, want < 1 %%", id, 100*greenShare)
		}
	}
}
