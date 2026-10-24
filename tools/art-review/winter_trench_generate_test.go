//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateWinterTrenches -count=1 -v
// Renders the winter form (suffix _w, connected-trench-winter-v1) of every
// trench piece and of every derived trench the levels select
// (mapcompose.TrenchShape, as TestGenerateCampaignTrenches lists them) into
// assets/terrain/trench/, replacing earlier winter exports, and lists them in
// assets/index.json; run `go run ./cmd/hlart index` afterwards.
import (
	"encoding/json"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
)

func TestGenerateWinterTrenches(t *testing.T) {
	root := filepath.Join("..", "..")
	summer, err := loadConnectionInputs(filepath.Join(root, "art/pipeline/connected-trench-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	winter, err := loadConnectionInputs(filepath.Join(root, "art/pipeline/connected-trench-winter-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	entries, exported, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	sf, wf := summer.manifest.Families[0], winter.manifest.Families[0]
	roster := map[string]bool{}
	for _, v := range sf.Variants {
		roster[v.ID] = true
		tiles, err := exportConnectionTiles(summer, &sf, &v)
		if err != nil {
			t.Fatal(err)
		}
		if !winterInstalledEqual(t, root, exported, "tile/"+v.ID, tiles[4]) {
			t.Fatalf("installed %s differs from the connected-trench-v2 render", v.ID)
		}
	}
	gaps, err := campaignTrenches(func(string) bool { return false })
	if err != nil {
		t.Fatal(err)
	}
	var derived []string
	for id := range gaps {
		if !roster[id] {
			derived = append(derived, id)
		}
	}
	sort.Strings(derived)
	entries = slices.DeleteFunc(entries, func(raw json.RawMessage) bool {
		var e campaignEntry
		return json.Unmarshal(raw, &e) == nil && strings.HasPrefix(e.ID, "tile/trench") && strings.HasSuffix(e.ID, "_w")
	})
	variants := slices.Clone(wf.Variants)
	for _, id := range derived {
		shape, ok := mapcompose.ParseTrench(id)
		if !ok {
			t.Fatalf("unknown trench %s", id)
		}
		if _, ok := exported[campaignExportKey(mapcompose.TileID(id), 4)]; !ok {
			t.Fatalf("summer %s is not installed; run TestGenerateCampaignTrenches first", id)
		}
		v := derivedTrenchVariant(wf.Variants, shape)
		v.ID += "_w"
		variants = append(variants, v)
	}
	for _, v := range variants {
		tiles, err := exportConnectionTiles(winter, &wf, &v)
		if err != nil {
			t.Fatal(err)
		}
		entries = append(entries, writeWinterTiles(t, root, v.ID, "trench", "infrastructure", tiles)...)
	}
	t.Logf("wrote %d winter trench pieces and %d derived winter trenches", len(wf.Variants), len(derived))
	writeWinterIndex(t, root, entries)
}
