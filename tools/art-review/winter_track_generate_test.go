//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateWinterPathsAndRails -count=1 -v
// Renders the winter paths (tile/path*_w, connected-paths-winter-v2) into
// assets/terrain/path/, the winter rails (tile/rail*_w,
// connected-rail-winter-v2) into assets/terrain/rail/, and the winter
// footbridge tiles and level crossings (tile/bridge25_w, bridge36_w,
// cross1_w..cross4_w, connected-mixed-winter-v2: the summer mixed tiles with
// winter path, road and rail components) into assets/terrain/bridge/ and
// assets/terrain/cross/, replacing earlier winter exports, and lists them in
// assets/index.json; run `go run ./cmd/hlart index` afterwards. This generator
// owns every winter crossing; TestGenerateWinterRoads owns the winter roads.
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

func TestGenerateWinterPathsAndRails(t *testing.T) {
	root := filepath.Join("..", "..")
	mixed, err := loadWinterMixedInputs(filepath.Join(root, "art/pipeline/connected-mixed-winter-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	entries, _, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	owned := func(id string) bool {
		if !strings.HasSuffix(id, "_w") {
			return false
		}
		for _, prefix := range []string{"tile/path", "tile/rail", "tile/cross"} {
			if strings.HasPrefix(id, prefix) {
				return true
			}
		}
		return id == "tile/bridge25_w" || id == "tile/bridge36_w"
	}
	entries = slices.DeleteFunc(entries, func(raw json.RawMessage) bool {
		var e campaignEntry
		return json.Unmarshal(raw, &e) == nil && owned(e.ID)
	})
	write := func(id, family string, tiles map[int]*image.NRGBA) {
		for _, d := range []int{2, 4} {
			path := fmt.Sprintf("terrain/%s/%s-%dx.png", family, id, d)
			var b bytes.Buffer
			if err := png.Encode(&b, tiles[d]); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "assets", path), b.Bytes(), 0o644); err != nil {
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
	for _, name := range []string{"connected-paths-winter-v2.json", "connected-rail-winter-v2.json"} {
		in, err := loadConnectionInputs(filepath.Join(root, "art/pipeline", name))
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range in.manifest.Families {
			for _, v := range f.Variants {
				tiles, err := exportConnectionTiles(in, &f, &v)
				if err != nil {
					t.Fatal(err)
				}
				write(v.ID, f.ID, tiles)
			}
		}
	}
	in := mixed.inputs()
	for _, v := range mixed.variants {
		tiles, err := exportMixedTiles(in, v)
		if err != nil {
			t.Fatal(err)
		}
		family := "cross"
		if v.Kind == "footbridge" {
			family = "bridge"
		}
		write(v.RuntimeID+"_w", family, tiles)
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
