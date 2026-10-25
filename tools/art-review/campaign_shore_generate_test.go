//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateCampaignShores -count=1 -v
// Every derived shore tile the levels select (mapcompose.ShoreShape) that the
// runtime index lacks is rendered with the shore-terrain-v2 materials (winter
// forms, suffix _w, with shore-terrain-winter-v1), written to
// assets/terrain/lakeside/ and added to assets/index.json; run
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
	"sort"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
)

func TestGenerateCampaignShores(t *testing.T) {
	root := filepath.Join("..", "..")
	in, err := loadShoreInputs(filepath.Join(root, "art/pipeline/shore-terrain-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	winter, err := loadShoreInputs(filepath.Join(root, "art/pipeline/shore-terrain-winter-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	indexBefore, err := os.ReadFile(filepath.Join(root, "assets/index.json"))
	if err != nil {
		t.Fatal(err)
	}
	entries, exported, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	// The installed roster tiles must come from these inputs, or a derived tile
	// would not join its lakeside neighbours.
	for _, v := range shoreRoster() {
		tiles, err := exportShoreTiles(in, v)
		if err != nil {
			t.Fatal(err)
		}
		e := exported[campaignExportKey("tile/"+v.ID, 4)]
		b, err := os.ReadFile(filepath.Join(root, "assets", e.Path))
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(im.(*image.NRGBA).Pix, tiles[4].Pix) {
			t.Fatalf("installed %s differs from the shore-terrain-v2 render", v.ID)
		}
	}
	has := func(id string) bool {
		_, a := exported[campaignExportKey(id, 2)]
		_, b := exported[campaignExportKey(id, 4)]
		return a && b
	}
	// A winter mouth shows up as a gap only once its winter shore exists, so
	// export until the levels select no absent shore.
	written := 0
	for {
		gaps, err := campaignShores(has)
		if err != nil {
			t.Fatal(err)
		}
		if len(gaps) == 0 {
			break
		}
		ids := make([]string, 0, len(gaps))
		for id := range gaps {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		t.Logf("absent shore tiles: %d", len(ids))
		for _, id := range ids {
			tiles, err := campaignShoreTiles(in, winter, id)
			if err != nil {
				t.Fatal(err)
			}
			for _, d := range []int{2, 4} {
				path := fmt.Sprintf("assets/terrain/lakeside/%s-%dx.png", id, d)
				var b bytes.Buffer
				if err := png.Encode(&b, tiles[d]); err != nil {
					t.Fatal(err)
				}
				full := filepath.Join(root, path)
				if _, err := os.Stat(full); err == nil {
					t.Fatalf("unindexed export exists: %s", path)
				}
				if err := os.WriteFile(full, b.Bytes(), 0o644); err != nil {
					t.Fatal(err)
				}
				sum := sha256.Sum256(b.Bytes())
				e := campaignEntry{ID: mapcompose.TileID(id), Pack: fmt.Sprintf("terrain-backgrounds-%dx", d), Density: d, Path: strings.TrimPrefix(path, "assets/"), SHA256: hex.EncodeToString(sum[:]), Width: tiles[d].Bounds().Dx(), Height: tiles[d].Bounds().Dy(), Anchor: [2]float64{48, 40}, Art: "terrain/lakeside"}
				raw, err := json.Marshal(e)
				if err != nil {
					t.Fatal(err)
				}
				entries = append(entries, raw)
				exported[campaignExportKey(e.ID, d)] = e
				written++
			}
		}
	}
	if written == 0 {
		return
	}
	name := filepath.Join(root, "assets/index.json")
	current, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(current, indexBefore) {
		t.Fatal("runtime index changed during export")
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
	if err := os.WriteFile(name, out.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

// campaignShoreTiles renders the shore tile id: a derived shape, or the winter
// form (suffix _w) of a roster shore, a stream mouth or a derived shape.
func campaignShoreTiles(summer, winter shoreInputs, id string) (map[int]*image.NRGBA, error) {
	in := summer
	base, cold := strings.CutSuffix(id, "_w")
	if cold {
		in = winter
	}
	shore, port := base, 0
	if i := strings.Index(base, "-outlet"); i >= 0 {
		shore = base[:i]
		if _, err := fmt.Sscanf(base[i:], "-outlet%d", &port); err != nil {
			return nil, fmt.Errorf("shore %s: %w", id, err)
		}
	}
	for _, v := range shoreRoster() {
		if v.ID != shore {
			continue
		}
		canonical, err := exportShoreTiles(in, v)
		if err != nil || port == 0 {
			return canonical, err
		}
		return exportShoreMouthOnBase(in, v, port, canonical)
	}
	shape, ok := mapcompose.ParseShore(base)
	if !ok || port != 0 {
		return nil, fmt.Errorf("unknown shore %s", id)
	}
	return exportShoreShapeTiles(in, shape)
}
