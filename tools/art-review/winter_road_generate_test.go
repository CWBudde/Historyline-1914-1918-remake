//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateWinterRoads -count=1 -v
// Renders the winter roads (tile/road*_w, connected-road-winter-v1) into
// assets/terrain/road/, replacing earlier winter exports, and lists them in
// assets/index.json; run `go run ./cmd/hlart index` afterwards. The winter
// level crossings, cross2_w and cross3_w included, belong to
// TestGenerateWinterPathsAndRails (connected-mixed-winter-v1): rerun it after
// changing the winter road.
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

func TestGenerateWinterRoads(t *testing.T) {
	root := filepath.Join("..", "..")
	winter, err := loadConnectionInputs(filepath.Join(root, "art/pipeline/connected-road-winter-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	entries, _, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	winterTile := func(id string) bool {
		return strings.HasSuffix(id, "_w") && strings.HasPrefix(id, "tile/road")
	}
	entries = slices.DeleteFunc(entries, func(raw json.RawMessage) bool {
		var e campaignEntry
		return json.Unmarshal(raw, &e) == nil && winterTile(e.ID)
	})
	write := func(id, family string, tiles map[int]*image.NRGBA) {
		for _, d := range []int{2, 4} {
			path := fmt.Sprintf("terrain/%s/%s-%dx.png", family, id, d)
			var b bytes.Buffer
			if err := png.Encode(&b, tiles[d]); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "assets", path), b.Bytes(), 0644); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(b.Bytes())
			raw, err := json.Marshal(campaignEntry{ID: "tile/" + id, Pack: fmt.Sprintf("terrain-infrastructure-%dx", d), Density: d, Path: path, SHA256: hex.EncodeToString(sum[:]), Width: 96 * d, Height: 80 * d, Anchor: [2]float64{48, 40}, Art: "terrain/" + family})
			if err != nil {
				t.Fatal(err)
			}
			entries = append(entries, raw)
		}
	}
	for _, f := range winter.manifest.Families {
		for _, v := range f.Variants {
			tiles, err := exportConnectionTiles(winter, &f, &v)
			if err != nil {
				t.Fatal(err)
			}
			write(v.ID, "road", tiles)
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
