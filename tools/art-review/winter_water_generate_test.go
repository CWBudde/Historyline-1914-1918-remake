//go:build campaigngfx

package artreview

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateWinterWater -count=1 -v
// Installs the winter open water tile/water_w as the installed summer pixels:
// the pack build grades a winter tile's water to the cold winter water
// (terraingrade), as it does the water of the winter shores, so open water and
// its coast stay one material. tile/waterStones_w is the winter water-stones
// scene (compound-water-stones-winter-v1.json, the snow-capped rocks under the
// summer rocks' alpha) over that same water. Writes them to
// assets/terrain/water/, replacing earlier winter exports, and lists them in
// assets/index.json; run `go run ./cmd/hl art index` afterwards.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateWinterWater(t *testing.T) {
	root := filepath.Join("..", "..")
	stones := waterStonesExports(t, root, waterStonesWinterManifest, waterStonesSummerManifest)
	entries, exported, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{"tile/water", "tile/waterStones"}
	for _, id := range ids {
		for _, d := range []int{2, 4} {
			summer, ok := exported[campaignExportKey(id, d)]
			if !ok {
				t.Fatalf("%s at %dx is not installed", id, d)
			}
			b, err := os.ReadFile(filepath.Join(root, "assets", summer.Path))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := png.Decode(bytes.NewReader(b)); err != nil {
				t.Fatalf("%s: %v", summer.Path, err)
			}
			if id == "tile/waterStones" {
				b = stones[d].stonesPNG
			}
			e := summer
			e.ID = id + "_w"
			e.Path = strings.Replace(summer.Path, "-", "_w-", 1)
			if err := os.WriteFile(filepath.Join(root, "assets", e.Path), b, 0o644); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(b)
			e.SHA256 = hex.EncodeToString(sum[:])
			raw, err := json.Marshal(e)
			if err != nil {
				t.Fatal(err)
			}
			entries = setCampaignEntry(entries, raw)
		}
	}
	writeCampaignIndex(t, root, entries)
}
