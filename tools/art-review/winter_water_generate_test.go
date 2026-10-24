//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateWinterWater -count=1 -v
// Installs the winter open water (tile/water_w, tile/waterStones_w) as the
// installed summer pixels: the pack build grades a winter tile's water to the
// cold winter water (terraingrade), as it does the water of the winter shores,
// so open water and its coast stay one material. Writes them to
// assets/terrain/water/, replacing earlier winter exports, and lists them in
// assets/index.json; run `go run ./cmd/hlart index` afterwards.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestGenerateWinterWater(t *testing.T) {
	root := filepath.Join("..", "..")
	entries, exported, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{"tile/water", "tile/waterStones"}
	entries = slices.DeleteFunc(entries, func(raw json.RawMessage) bool {
		var e campaignEntry
		return json.Unmarshal(raw, &e) == nil && slices.Contains(ids, strings.TrimSuffix(e.ID, "_w")) && strings.HasSuffix(e.ID, "_w")
	})
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
			e := summer
			e.ID = id + "_w"
			e.Path = strings.Replace(summer.Path, "-", "_w-", 1)
			if err := os.WriteFile(filepath.Join(root, "assets", e.Path), b, 0644); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(b)
			e.SHA256 = hex.EncodeToString(sum[:])
			raw, err := json.Marshal(e)
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
