//go:build campaigngfx

package artreview

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateWinterRoadBridgeTiles -count=1 -v
// Writes the winter road bridge tiles (tile/rBridge*_w) into
// assets/terrain/rBridge/ and lists them in assets/index.json; run
// `go run ./cmd/hl art index` afterwards. connected-mixed-v4 renders a road
// bridge's map tile empty under the standing deck in every season, so these
// are the summer tiles' exports under the winter IDs.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestGenerateWinterRoadBridgeTiles(t *testing.T) {
	root := filepath.Join("..", "..")
	mixed, err := loadMixedInputs(filepath.Join(root, "art/pipeline/connected-mixed-v4.json"))
	if err != nil {
		t.Fatal(err)
	}
	entries, _, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	entries = slices.DeleteFunc(entries, func(raw json.RawMessage) bool {
		var e campaignEntry
		return json.Unmarshal(raw, &e) == nil && strings.HasPrefix(e.ID, "tile/rBridge") && strings.HasSuffix(e.ID, "_w")
	})
	for _, v := range mixed.manifest.Variants {
		if v.Kind != "road-bridge" {
			continue
		}
		tiles, err := exportMixedTiles(mixed, v)
		if err != nil {
			t.Fatal(err)
		}
		id := v.RuntimeID + "_w"
		for _, d := range []int{2, 4} {
			path := fmt.Sprintf("terrain/rBridge/%s-%dx.png", id, d)
			var b bytes.Buffer
			if err := png.Encode(&b, tiles[d]); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "assets", path), b.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(b.Bytes())
			raw, err := json.Marshal(campaignEntry{ID: "tile/" + id, Pack: fmt.Sprintf("terrain-infrastructure-%dx", d), Density: d, Path: path, SHA256: hex.EncodeToString(sum[:]), Width: 96 * d, Height: 80 * d, Anchor: [2]float64{48, 40}, Art: "terrain/rBridge"})
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
	if err := os.WriteFile(filepath.Join(root, "assets/index.json"), out.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}
