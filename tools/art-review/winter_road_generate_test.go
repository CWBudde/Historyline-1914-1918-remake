//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateWinterRoads -count=1 -v
// Renders the winter roads (tile/road*_w, connected-road-winter-v1) into
// assets/terrain/road/ and the winter level crossings whose road carries the
// winter road (tile/cross2_w, cross3_w: the connected-mixed-v2 crossing with its
// road layer swapped for the winter road) into assets/terrain/cross/,
// replacing earlier winter exports, and lists them in assets/index.json; run
// `go run ./cmd/hlart index` afterwards.
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
	mixed, err := loadMixedInputs(filepath.Join(root, "art/pipeline/connected-mixed-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	entries, _, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	winterTile := func(id string) bool {
		return strings.HasSuffix(id, "_w") && (strings.HasPrefix(id, "tile/road") || strings.HasPrefix(id, "tile/cross"))
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
	roads := map[string]mixedComponent{}
	for _, f := range winter.manifest.Families {
		for _, v := range f.Variants {
			tiles, err := exportConnectionTiles(winter, &f, &v)
			if err != nil {
				t.Fatal(err)
			}
			write(v.ID, "road", tiles)
			roads[v.ID] = mixedComponent{winter, f, v}
		}
	}
	for _, v := range mixedRoster() {
		if v.Kind != "crossing" || !strings.HasPrefix(v.Layers[0], "road") {
			continue
		}
		road, ok := roads[v.Layers[0]+"_w"]
		if !ok {
			t.Fatalf("%s: no winter road %s_w", v.ID, v.Layers[0])
		}
		in := mixed
		in.components = map[string]mixedComponent{}
		for id, c := range mixed.components {
			in.components[id] = c
		}
		in.components[v.Layers[0]] = road
		tiles, err := exportMixedTiles(in, v)
		if err != nil {
			t.Fatal(err)
		}
		write(v.RuntimeID+"_w", "cross", tiles)
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
