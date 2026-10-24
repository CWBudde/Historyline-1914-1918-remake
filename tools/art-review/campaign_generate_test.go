//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateCampaignContexts -count=1 -v
// Normal tests never write artwork. Existing context exports are hash-checked
// and retained. Absent contexts are exported and added to their profiles and
// assets/index.json; run `go run ./cmd/hlart index` afterwards.
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

func TestGenerateCampaignContexts(t *testing.T) {
	root := filepath.Join("..", "..")
	read := func(p string) map[string]any {
		b, err := os.ReadFile(filepath.Join(root, p))
		if err != nil {
			t.Fatal(err)
		}
		var v map[string]any
		if err := json.Unmarshal(b, &v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	profiles := map[mapcompose.Family]map[string]any{}
	profileBefore := map[mapcompose.Family][]byte{}
	known := map[string]bool{}
	for _, f := range mapcompose.Families {
		b, err := os.ReadFile(filepath.Join(root, "art/pipeline/contexts/"+string(f)+".json"))
		if err != nil {
			t.Fatal(err)
		}
		profileBefore[f] = b
		r := read("art/pipeline/contexts/" + string(f) + ".json")
		profiles[f] = r
		for _, p := range r["profiles"].([]any) {
			known[p.(map[string]any)["context_key"].(string)] = true
		}
	}
	indexBefore, err := os.ReadFile(filepath.Join(root, "assets/index.json"))
	if err != nil {
		t.Fatal(err)
	}
	entries, exported, err := readCampaignIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range exported {
		if !strings.HasPrefix(e.ID, "context/") {
			continue
		}
		for _, d := range []int{2, 4} {
			if _, ok := exported[campaignExportKey(e.ID, d)]; !ok {
				t.Fatalf("partial retained context %s", e.ID)
			}
		}
		b, err := os.ReadFile(filepath.Join(root, "assets", e.Path))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(b)) != e.SHA256 {
			t.Fatalf("retained export hash differs: %s", e.Path)
		}
	}
	has := func(id string) bool {
		_, a := exported[campaignExportKey(id, 2)]
		_, b := exported[campaignExportKey(id, 4)]
		return a && b
	}
	contexts, err := campaignContexts(has)
	if err != nil {
		t.Fatal(err)
	}
	counts := map[mapcompose.Family]int{}
	for _, g := range contexts {
		counts[g.Family]++
	}
	t.Logf("absent exact-density contexts: %d, by family: %v", len(contexts), counts)
	ground, forest, hills, err := loadCampaignRenderers(root)
	if err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(contexts))
	for key := range contexts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	added := map[mapcompose.Family]int{}
	generated := 0
	newExports := 0
	for _, key := range keys {
		g := contexts[key]
		fields := map[string]string{}
		for _, line := range strings.Split(g.Descriptor, "\n") {
			k, v, ok := strings.Cut(line, "=")
			if ok {
				fields[k] = v
			}
		}
		var ns [6]string
		for i, k := range []string{"NE", "SE", "S", "SW", "NW", "N"} {
			ns[i] = fields[k]
		}
		p := map[string]any{"runtime_id": g.Terrain, "season": fields["season"], "context_key": key, "descriptor": g.Descriptor, "neighbor_ids": ns, "canonical": false}
		var tiles map[int]*image.NRGBA
		switch g.Family {
		case mapcompose.Ground:
			c, e := newGroundContext(g.Terrain, ns[:])
			if e != nil {
				t.Fatal(e)
			}
			if c.key() != key {
				t.Fatal("ground key mismatch")
			}
			tiles, err = ground.tiles(c)
		case mapcompose.Forest:
			// Summer's complete 64-mask catalogue uses layered floors/crowns/shadows.
			// The baked renderer below is used only for missing winter masks.
			if fields["season"] != "winter" {
				t.Fatalf("layered summer forest export missing: %s", key)
			}
			c, e := newForestContext(g.Terrain, ns[:])
			if e != nil {
				t.Fatal(e)
			}
			if c.key() != key {
				t.Fatal("forest key mismatch")
			}
			tiles, err = forest.tiles(c)
		case mapcompose.Hills:
			origin := connectionPoint{float64(142*g.At.Col + 71*(g.At.Line%2)), float64(40 * g.At.Line)}
			c, e := newHillsContext(g.Terrain, ns[:], origin)
			if e != nil {
				t.Fatal(e)
			}
			if c.key() != key {
				t.Fatal("hills key mismatch")
			}
			tiles, err = hills.tiles(c)
			p["phase"] = c.Phase
		}
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range []int{2, 4} {
			path := fmt.Sprintf("assets/terrain/%s/context/%s-%dx.png", g.Family, key, d)
			var b bytes.Buffer
			if err := png.Encode(&b, tiles[d]); err != nil {
				t.Fatal(err)
			}
			decoded, err := png.Decode(bytes.NewReader(b.Bytes()))
			if err != nil {
				t.Fatal(err)
			}
			if decoded.Bounds() != tiles[d].Bounds() {
				t.Fatal("PNG bounds changed")
			}
			full := filepath.Join(root, path)
			if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
				t.Fatal(err)
			}
			if old, e := os.ReadFile(full); e == nil {
				if !bytes.Equal(old, b.Bytes()) {
					t.Fatalf("existing bytes differ: %s", path)
				}
			} else if !os.IsNotExist(e) {
				t.Fatal(e)
			} else if e = os.WriteFile(full, b.Bytes(), 0644); e != nil {
				t.Fatal(e)
			}
			sum := sha256.Sum256(b.Bytes())
			hash := hex.EncodeToString(sum[:])
			e := campaignEntry{ID: mapcompose.ContextID(g.Family, key), Pack: fmt.Sprintf("terrain-context-%s-%dx", g.Family, d), Density: d, Path: strings.TrimPrefix(path, "assets/"), SHA256: hash, Width: tiles[d].Bounds().Dx(), Height: tiles[d].Bounds().Dy(), Anchor: [2]float64{48, 40}, Art: "terrain/" + string(g.Family)}
			raw, err := json.Marshal(e)
			if err != nil {
				t.Fatal(err)
			}
			entries = append(entries, raw)
			newExports++
		}
		generated++
		if generated%50 == 0 {
			t.Logf("exported %d/%d context pairs", generated, len(keys))
		}
		if known[key] {
			continue
		}
		r := profiles[g.Family]
		r["profiles"] = append(r["profiles"].([]any), p)
		added[g.Family]++
	}
	for _, f := range mapcompose.Families {
		if added[f] == 0 {
			continue
		}
		b, err := json.MarshalIndent(profiles[f], "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		name := filepath.Join(root, "art/pipeline/contexts/"+string(f)+".json")
		current, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(current, profileBefore[f]) {
			t.Fatalf("profile changed during export: %s", name)
		}
		if err := os.WriteFile(name, append(b, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if newExports > 0 {
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
		if err := os.WriteFile(name, out.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("newly exported context pairs: %d; new profiles by family: %v", generated, added)
}

func loadCampaignRenderers(root string) (*groundRenderer, *forestRenderer, *hillsRenderer, error) {
	gi, err := loadGroundInputs(filepath.Join(root, groundManifestPath))
	if err != nil {
		return nil, nil, nil, err
	}
	materials := map[string]*image.NRGBA{}
	for id, s := range gi.Sources {
		materials[id] = s.Image
	}
	ground := newLatticeGroundRenderer(materials, gi.Manifest.LatticeMaterials)
	ground.mudFeature = gi.Manifest.MudFeature
	ground.shelledFeature = gi.Manifest.ShelledFeature
	fi, err := loadForestInputs(filepath.Join(root, "art/pipeline/forest-terrain-v3.json"))
	if err != nil {
		return nil, nil, nil, err
	}
	grounds, atlases := map[string]*image.NRGBA{}, map[string]*image.NRGBA{}
	for id, s := range fi.Grounds {
		grounds[id] = s.Image
	}
	for id, s := range fi.Atlases {
		atlases[id] = s.Image
	}
	forest, err := newForestRenderer(grounds, atlases, fi.Manifest.SummerProfile)
	if err != nil {
		return nil, nil, nil, err
	}
	hi, err := loadHillsInputs(filepath.Join(root, "art/pipeline/hills-terrain-v2.json"))
	if err != nil {
		return nil, nil, nil, err
	}
	hills := newHillsRenderer(hi.Sources["grass"].Image, hi.Sources["height"].Image)
	if source, ok := hi.Sources["scrub"]; ok {
		hills.scrub = source.Image
	}
	return ground, forest, hills, nil
}
