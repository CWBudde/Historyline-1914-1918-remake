//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateWinterStreams -count=1 -v
// Renders the winter streams (tile/stream*_w, connected-stream-winter-v1) into
// assets/terrain/stream/ and re-renders the winter stream mouths
// (tile/<shore>-outlet<n>_w, shore-terrain-winter-v1), whose mouth now carries
// the winter stream, into assets/terrain/lakeside/, replacing earlier winter
// exports, and lists them in assets/index.json; run `go run ./cmd/hlart index`
// afterwards.
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

func TestGenerateWinterStreams(t *testing.T) {
	root := filepath.Join("..", "..")
	summer, err := loadConnectionInputs(filepath.Join(root, "art/pipeline/connected-stream-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	winter, err := loadConnectionInputs(filepath.Join(root, "art/pipeline/connected-stream-winter-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	shoreSummer, err := loadShoreInputs(filepath.Join(root, "art/pipeline/shore-terrain-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	shoreWinter, err := loadShoreInputs(filepath.Join(root, "art/pipeline/shore-terrain-winter-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	entries, exported, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	// The installed summer streams must come from connected-stream-v2, or the
	// winter forms would not keep their courses.
	sf := summer.manifest.Families[0]
	for _, v := range sf.Variants {
		tiles, err := exportConnectionTiles(summer, &sf, &v)
		if err != nil {
			t.Fatal(err)
		}
		if !winterInstalledEqual(t, root, exported, "tile/"+v.ID, tiles[4]) {
			t.Fatalf("installed %s differs from the connected-stream-v2 render", v.ID)
		}
	}
	var mouths []string
	for _, e := range exported {
		if e.Density == 4 && strings.HasSuffix(e.ID, "_w") && strings.Contains(e.ID, "-outlet") {
			mouths = append(mouths, strings.TrimPrefix(e.ID, "tile/"))
		}
	}
	slices.Sort(mouths)
	stale := func(id string) bool {
		return strings.HasSuffix(id, "_w") && (strings.HasPrefix(id, "tile/stream") || slices.Contains(mouths, strings.TrimPrefix(id, "tile/")))
	}
	entries = slices.DeleteFunc(entries, func(raw json.RawMessage) bool {
		var e campaignEntry
		return json.Unmarshal(raw, &e) == nil && stale(e.ID)
	})
	wf := winter.manifest.Families[0]
	for _, v := range wf.Variants {
		tiles, err := exportConnectionTiles(winter, &wf, &v)
		if err != nil {
			t.Fatal(err)
		}
		entries = append(entries, writeWinterTiles(t, root, v.ID, "stream", "infrastructure", tiles)...)
	}
	for _, id := range mouths {
		tiles, err := campaignShoreTiles(shoreSummer, shoreWinter, id)
		if err != nil {
			t.Fatal(err)
		}
		entries = append(entries, writeWinterTiles(t, root, id, "lakeside", "backgrounds", tiles)...)
	}
	t.Logf("wrote %d winter streams and %d winter mouths %v", len(wf.Variants), len(mouths), mouths)
	writeWinterIndex(t, root, entries)
}

// winterInstalledEqual reports whether the installed 4x export of id has the
// pixels of im.
func winterInstalledEqual(t *testing.T, root string, exported map[string]campaignEntry, id string, im *image.NRGBA) bool {
	t.Helper()
	e, ok := exported[campaignExportKey(id, 4)]
	if !ok {
		t.Fatalf("%s is not installed", id)
	}
	b, err := os.ReadFile(filepath.Join(root, "assets", e.Path))
	if err != nil {
		t.Fatal(err)
	}
	installed, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	n, ok := installed.(*image.NRGBA)
	return ok && bytes.Equal(n.Pix, im.Pix)
}

// writeWinterTiles writes the 2x and 4x exports of tile id of art family
// family into assets/terrain/<family>/ and returns their index entries in the
// terrain-<pack>-<d>x packs.
func writeWinterTiles(t *testing.T, root, id, family, pack string, tiles map[int]*image.NRGBA) []json.RawMessage {
	t.Helper()
	var out []json.RawMessage
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
		raw, err := json.Marshal(campaignEntry{ID: "tile/" + id, Pack: fmt.Sprintf("terrain-%s-%dx", pack, d), Density: d, Path: path, SHA256: hex.EncodeToString(sum[:]), Width: 96 * d, Height: 80 * d, Anchor: [2]float64{48, 40}, Art: "terrain/" + family})
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, raw)
	}
	return out
}

func writeWinterIndex(t *testing.T, root string, entries []json.RawMessage) {
	t.Helper()
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
