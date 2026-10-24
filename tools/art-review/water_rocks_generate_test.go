//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateWaterRocks -count=1 -v
// Renders tile/waterRocks (and tile/waterRocks_w, the same pixels under the
// winter water grade) from the water-stones scene of
// art/pipeline/compound-water-stones-v2.json: the rocks alone, without their
// water, so a stones hex draws them over the shore its neighbours give it.
// Over the open water tile they are tile/waterStones. Writes them to
// assets/terrain/water/, replacing earlier exports, and lists them in
// assets/index.json; run `go run ./cmd/hlart index` afterwards.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestGenerateWaterRocks(t *testing.T) {
	root := filepath.Join("..", "..")
	in, err := loadCompoundInputs(filepath.Join(root, "art/pipeline/compound-water-stones-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := buildCompoundBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	entries, exported, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	entries = slices.DeleteFunc(entries, func(raw json.RawMessage) bool {
		var e campaignEntry
		return json.Unmarshal(raw, &e) == nil && strings.TrimSuffix(e.ID, "_w") == "tile/waterRocks"
	})
	for _, d := range []int{2, 4} {
		art := b.pixels[fmt.Sprintf("water-stones-artwork-%dx.png", d)]
		if art == nil {
			t.Fatalf("no water-stones artwork at %dx", d)
		}
		rocks := image.NewNRGBA(image.Rect(0, 0, 96*d, 80*d))
		for y := range 80 * d {
			for x := range 96 * d {
				if connectionInside(sharedConnectionGeometry(), connectionPoint{(float64(x) + .5) / float64(d), (float64(y) + .5) / float64(d)}) {
					rocks.SetNRGBA(x, y, art.NRGBAAt(x, y))
				}
			}
		}
		// The rocks over the installed water are the installed stones tile.
		water, stones := installedTile(t, root, exported, "tile/water", d), installedTile(t, root, exported, "tile/waterStones", d)
		if got := compoundCompose(water, rocks); !bytes.Equal(got.Pix, stones.Pix) {
			t.Fatalf("%dx: the rocks over tile/water differ from tile/waterStones", d)
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, rocks); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(buf.Bytes())
		for _, id := range []string{"waterRocks", "waterRocks_w"} {
			path := fmt.Sprintf("terrain/water/%s-%dx.png", id, d)
			if err := os.WriteFile(filepath.Join(root, "assets", path), buf.Bytes(), 0644); err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(campaignEntry{ID: "tile/" + id, Pack: fmt.Sprintf("terrain-infrastructure-%dx", d), Density: d, Path: path, SHA256: hex.EncodeToString(sum[:]), Width: 96 * d, Height: 80 * d, Anchor: [2]float64{48, 40}, Art: "terrain/water"})
			if err != nil {
				t.Fatal(err)
			}
			entries = append(entries, raw)
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
	if err := os.WriteFile(filepath.Join(root, "assets/index.json"), out.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
}

func installedTile(t *testing.T, root string, exported map[string]campaignEntry, id string, d int) *image.NRGBA {
	t.Helper()
	e, ok := exported[campaignExportKey(id, d)]
	if !ok {
		t.Fatalf("%s at %dx is not installed", id, d)
	}
	f, err := os.Open(filepath.Join(root, "assets", e.Path))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	im, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	out := image.NewNRGBA(im.Bounds())
	for y := range im.Bounds().Dy() {
		for x := range im.Bounds().Dx() {
			out.Set(x, y, im.At(x, y))
		}
	}
	return out
}
