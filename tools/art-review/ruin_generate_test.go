//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateRuinTiles -count=1 -v
// Cuts the destroyed depot (art/pipeline/compound-ruin-v2.json) into tile/ruinN,
// ruinW, ruinE and ruinS in both seasons through the compound pipeline, writes
// them to assets/terrain/ruin/, replacing earlier exports, and lists them in
// assets/index.json; run `go run ./cmd/hlart index` afterwards.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestGenerateRuinTiles(t *testing.T) {
	root := filepath.Join("..", "..")
	in, err := loadCompoundInputs(filepath.Join(root, "art/pipeline/compound-ruin-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := buildCompoundBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{}
	for _, a := range bundle.artifacts {
		files[a.name] = a.data
	}
	entries, _, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	entries = slices.DeleteFunc(entries, func(raw json.RawMessage) bool {
		var e campaignEntry
		return json.Unmarshal(raw, &e) == nil && e.Art == "terrain/ruin"
	})
	if err := os.MkdirAll(filepath.Join(root, "assets/terrain/ruin"), 0755); err != nil {
		t.Fatal(err)
	}
	for _, d := range []int{2, 4} {
		for _, sfx := range []string{"", "_w"} {
			for _, p := range []string{"E", "N", "S", "W"} {
				id := "ruin" + p + sfx
				data, ok := files[fmt.Sprintf("tiles/%s-%dx.png", id, d)]
				if !ok {
					t.Fatalf("compound bundle has no %s at %dx", id, d)
				}
				path := fmt.Sprintf("terrain/ruin/%s-%dx.png", id, d)
				if err := os.WriteFile(filepath.Join(root, "assets", path), data, 0644); err != nil {
					t.Fatal(err)
				}
				sum := sha256.Sum256(data)
				raw, err := json.Marshal(campaignEntry{ID: "tile/" + id, Pack: fmt.Sprintf("terrain-backgrounds-%dx", d), Density: d, Path: path, SHA256: hex.EncodeToString(sum[:]), Width: 96 * d, Height: 80 * d, Anchor: [2]float64{48, 40}, Art: "terrain/ruin"})
				if err != nil {
					t.Fatal(err)
				}
				entries = append(entries, raw)
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
	if err := os.WriteFile(filepath.Join(root, "assets/index.json"), out.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
}
