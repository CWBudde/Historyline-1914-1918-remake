//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateSeaTiles -count=1 -v
// Renders the open sea, art/terrain/water/water-sea-shared-v1.png, as one
// world-periodic material of 284×160 world px (worldmaterial.Export and
// Seamless) and cuts it into the eight phases a hex of the lattice can take,
// tile/sea-<kx><ky> with kx = (2·col + line&1) mod 4 and ky = line mod 4, in
// the shape of tile/water. The material is toned so that it grades to the mean
// colour of the graded tile/water, so the depth grade (terraingrade.Deepen) still sets
// how dark the sea is; internal/waterdepth blends it in by depth. The winter
// forms (suffix _w) are the same pixels under the winter water grade. Writes
// them to assets/terrain/water/, replacing earlier exports, and lists them in
// assets/index.json; run `go run ./cmd/hlart index` afterwards.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/terraingrade"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/worldmaterial"
)

const seaPeriodX, seaPeriodY = 284, 160

func TestGenerateSeaTiles(t *testing.T) {
	root := filepath.Join("..", "..")
	f, err := os.Open(filepath.Join(root, "art/terrain/water/water-sea-shared-v1.png"))
	if err != nil {
		t.Fatal(err)
	}
	src, err := png.Decode(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	entries, exported, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	entries = slices.DeleteFunc(entries, func(raw json.RawMessage) bool {
		var e campaignEntry
		return json.Unmarshal(raw, &e) == nil && strings.HasPrefix(e.ID, "tile/sea-")
	})
	for _, d := range []int{2, 4} {
		period, _, err := worldmaterial.Export(src, image.Pt(seaPeriodX, seaPeriodY), d)
		if err != nil {
			t.Fatal(err)
		}
		if period, err = worldmaterial.Seamless(period, 8, d); err != nil {
			t.Fatal(err)
		}
		water := installedTile(t, root, exported, "tile/water", d)
		// Toned so that it grades (terraingrade) to the graded water's mean.
		want := meanRGB(terraingrade.Grade(water, "tile/water", "terrain/water", nil))
		tone(period, meanRGB(water))
		for range 4 {
			have, graded := meanRGB(period), meanRGB(terraingrade.Grade(period, "tile/water", "terrain/water", nil))
			tone(period, [3]float64{have[0] + want[0] - graded[0], have[1] + want[1] - graded[1], have[2] + want[2] - graded[2]})
		}
		for ky := range 4 {
			for kx := ky % 2; kx < 4; kx += 2 {
				tile := image.NewNRGBA(water.Rect)
				for y := range water.Rect.Dy() {
					for x := range water.Rect.Dx() {
						a := water.NRGBAAt(x, y).A
						if a == 0 {
							continue
						}
						c := period.NRGBAAt((kx*71*d+x)%(seaPeriodX*d), (ky*40*d+y)%(seaPeriodY*d))
						c.A = a
						tile.SetNRGBA(x, y, c)
					}
				}
				var buf bytes.Buffer
				if err := png.Encode(&buf, tile); err != nil {
					t.Fatal(err)
				}
				sum := sha256.Sum256(buf.Bytes())
				for _, id := range []string{fmt.Sprintf("sea-%d%d", kx, ky), fmt.Sprintf("sea-%d%d_w", kx, ky)} {
					path := fmt.Sprintf("terrain/water/%s-%dx.png", id, d)
					if err := os.WriteFile(filepath.Join(root, "assets", path), buf.Bytes(), 0o644); err != nil {
						t.Fatal(err)
					}
					raw, err := json.Marshal(campaignEntry{ID: "tile/" + id, Pack: fmt.Sprintf("terrain-backgrounds-%dx", d), Density: d, Path: path, SHA256: hex.EncodeToString(sum[:]), Width: water.Rect.Dx(), Height: water.Rect.Dy(), Anchor: [2]float64{48, 40}, Art: "terrain/water"})
					if err != nil {
						t.Fatal(err)
					}
					entries = append(entries, raw)
				}
			}
		}
	}
	var out bytes.Buffer
	out.WriteString("{\n  \"schema_version\": 1,\n  \"assets\": [\n")
	for i, e := range entries {
		out.WriteString("    ")
		out.Write(e)
		if i+1 < len(entries) {
			out.WriteByte(',')
		}
		out.WriteByte('\n')
	}
	out.WriteString("  ]\n}\n")
	if err := os.WriteFile(filepath.Join(root, "assets/index.json"), out.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

// meanRGB is the mean colour of the opaque pixels of im.
func meanRGB(im *image.NRGBA) [3]float64 {
	var sum [3]float64
	n := 0.
	for y := range im.Rect.Dy() {
		for x := range im.Rect.Dx() {
			if c := im.NRGBAAt(x, y); c.A == 255 {
				sum[0] += float64(c.R)
				sum[1] += float64(c.G)
				sum[2] += float64(c.B)
				n++
			}
		}
	}
	return [3]float64{sum[0] / n, sum[1] / n, sum[2] / n}
}

// tone shifts every channel of im so that its mean is want.
func tone(im *image.NRGBA, want [3]float64) {
	have := meanRGB(im)
	shift := func(v uint8, i int) uint8 {
		return uint8(math.Round(math.Max(0, math.Min(255, float64(v)+want[i]-have[i]))))
	}
	for y := range im.Rect.Dy() {
		for x := range im.Rect.Dx() {
			c := im.NRGBAAt(x, y)
			im.SetNRGBA(x, y, color.NRGBA{shift(c.R, 0), shift(c.G, 1), shift(c.B, 2), c.A})
		}
	}
}
