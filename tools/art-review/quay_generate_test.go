//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateQuayTiles -count=1 -v
// Renders tile/quay14, quay25 and quay36 from the selected pier painting, and
// their winter forms (quay14_w, …) from the winter pier, into
// assets/terrain/quay/, replacing earlier exports, and lists them in
// assets/index.json; run `go run ./cmd/hlart index` afterwards.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// quaySeasons are the pier paintings, the suffix of the IDs exported from
// each and the finish the export gives the strip (quayWinterFinish: the
// winter deck's tone and the ice collar).
var quaySeasons = []struct {
	master, suffix string
	finish         func(quayStrip) quayStrip
}{
	{"quay-pier-shared-v1.png", "", nil},
	{"quay-pier-winter-v1.png", "_w", quayWinterFinish},
}

func TestGenerateQuayTiles(t *testing.T) {
	root := filepath.Join("..", "..")
	entries, _, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	isQuay := func(raw json.RawMessage) bool {
		var e campaignEntry
		return json.Unmarshal(raw, &e) == nil && e.Art == "terrain/quay"
	}
	// The new entries take the place of the earlier ones.
	at := slices.IndexFunc(entries, isQuay)
	entries = slices.DeleteFunc(entries, isQuay)
	if at < 0 {
		at = len(entries)
	}
	var quay []json.RawMessage
	if err := os.MkdirAll(filepath.Join(root, "assets/terrain/quay"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, season := range quaySeasons {
		f, err := os.Open(filepath.Join(root, "art/terrain/quay", season.master))
		if err != nil {
			t.Fatal(err)
		}
		src, err := png.Decode(f)
		_ = f.Close()
		if err != nil {
			t.Fatal(err)
		}
		im := image.NewNRGBA(src.Bounds())
		draw.Draw(im, im.Rect, src, src.Bounds().Min, draw.Src)
		strip, err := newQuayStrip(im)
		if err != nil {
			t.Fatal(err)
		}
		if season.finish != nil {
			strip = season.finish(strip)
		}
		for _, axis := range []string{"quay14", "quay25", "quay36"} {
			tiles, err := quayTiles(strip, axis)
			if err != nil {
				t.Fatal(err)
			}
			id := axis + season.suffix
			for _, d := range []int{2, 4} {
				path := fmt.Sprintf("terrain/quay/%s-%dx.png", id, d)
				var b bytes.Buffer
				if err := png.Encode(&b, tiles[d]); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, "assets", path), b.Bytes(), 0o644); err != nil {
					t.Fatal(err)
				}
				sum := sha256.Sum256(b.Bytes())
				raw, err := json.Marshal(campaignEntry{ID: "tile/" + id, Pack: fmt.Sprintf("terrain-infrastructure-%dx", d), Density: d, Path: path, SHA256: hex.EncodeToString(sum[:]), Width: 96 * d, Height: 80 * d, Anchor: [2]float64{48, 40}, Art: "terrain/quay"})
				if err != nil {
					t.Fatal(err)
				}
				quay = append(quay, raw)
			}
		}
	}
	entries = slices.Insert(entries, at, quay...)
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
