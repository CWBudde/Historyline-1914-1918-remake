//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestRegenerateLatticeGroundContexts -count=1 -v
// Re-renders, from groundManifestPath, every exported ground context whose
// centre or neighbours use a material the manifest samples lattice-periodic
// (lattice_materials), canonical tiles included, overwriting their PNGs and
// updating their hashes in assets/index.json; run `go run ./cmd/hlart index`
// afterwards. TestGenerateCampaignContexts then finds them current.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image/png"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestRegenerateLatticeGroundContexts(t *testing.T) {
	root := filepath.Join("..", "..")
	gi, err := loadGroundInputs(filepath.Join(root, groundManifestPath))
	if err != nil {
		t.Fatal(err)
	}
	if len(gi.Manifest.LatticeMaterials) == 0 {
		t.Skip("no lattice materials")
	}
	ground, _, _, err := loadCampaignRenderers(root)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "art/pipeline/contexts/ground.json"))
	if err != nil {
		t.Fatal(err)
	}
	var profiles struct {
		Profiles []struct {
			RuntimeID   string    `json:"runtime_id"`
			ContextKey  string    `json:"context_key"`
			NeighborIDs [6]string `json:"neighbor_ids"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal(b, &profiles); err != nil {
		t.Fatal(err)
	}
	entries, exported, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	// New hashes by asset path: a canonical context shares its file with its tile.
	hashes := map[string][2]string{}
	for _, p := range profiles.Profiles {
		ids := append([]string{p.RuntimeID}, p.NeighborIDs[:]...)
		if !slices.ContainsFunc(ids, func(id string) bool { _, ok := gi.Manifest.LatticeMaterials[id]; return ok }) {
			continue
		}
		c, err := newGroundContext(p.RuntimeID, p.NeighborIDs[:])
		if err != nil {
			t.Fatal(err)
		}
		if c.key() != p.ContextKey {
			t.Fatalf("ground key mismatch: %s", p.ContextKey)
		}
		tiles, err := ground.tiles(c)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range []int{2, 4} {
			e, ok := exported[campaignExportKey("context/ground/"+p.ContextKey, d)]
			if !ok {
				continue
			}
			var buf bytes.Buffer
			if err := png.Encode(&buf, tiles[d]); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "assets", e.Path), buf.Bytes(), 0644); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(buf.Bytes())
			hashes[e.Path] = [2]string{e.SHA256, hex.EncodeToString(sum[:])}
		}
	}
	for i, raw := range entries {
		var e campaignEntry
		if json.Unmarshal(raw, &e) != nil {
			continue
		}
		if h, ok := hashes[e.Path]; ok && e.SHA256 == h[0] {
			entries[i] = bytes.Replace(raw, []byte(`"`+h[0]+`"`), []byte(`"`+h[1]+`"`), 1)
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
	t.Logf("re-rendered %d ground files for %s", len(hashes), strings.Join(slices.Sorted(maps.Keys(gi.Manifest.LatticeMaterials)), ", "))
}
