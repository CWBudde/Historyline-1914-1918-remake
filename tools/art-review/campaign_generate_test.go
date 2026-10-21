//go:build campaigngfx

package main

// Explicit offline generator: go test -tags campaigngfx ./tools/art-review -run TestGenerateCampaignContexts -count=1 -v
// Normal tests never write artwork. Existing images are checked byte for byte
// against assets/index.json. New contexts are written to assets/terrain/<family>/context/
// and appended to art/pipeline/contexts/<family>.json; run `go run ./cmd/hlart index` afterwards.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
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
	known := map[string]bool{}
	for _, f := range mapcompose.Families {
		r := read("art/pipeline/contexts/" + string(f) + ".json")
		profiles[f] = r
		for _, p := range r["profiles"].([]any) {
			known[p.(map[string]any)["context_key"].(string)] = true
		}
	}
	// Exported context tiles by ID and density: path below the repository root and hash.
	type pin struct{ path, sha256 string }
	exported := map[string]pin{}
	for _, e := range read("assets/index.json")["assets"].([]any) {
		entry := e.(map[string]any)
		if id := entry["id"].(string); strings.HasPrefix(id, "context/") {
			exported[fmt.Sprintf("%s@%v", id, entry["density"])] = pin{"assets/" + entry["path"].(string), entry["sha256"].(string)}
		}
	}
	contexts := map[string]mapcompose.Gap{}
	names, err := fs.Glob(levels.FS, "*.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		b, err := fs.ReadFile(levels.FS, name)
		if err != nil {
			t.Fatal(err)
		}
		l, err := level.Parse(b)
		if err != nil {
			t.Fatal(err)
		}
		s := mapcompose.Compose(l.Grid(), func(c hexgrid.Coord) mapcompose.Hex {
			h := l.Lines[c.Line][c.Col]
			return mapcompose.Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
		}, func(string) bool { return false })
		for _, g := range s.Gaps {
			if g.Reason == mapcompose.Missing {
				contexts[mapcompose.Key(g.Descriptor)] = g
			}
		}
	}
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
			var expected string
			if known[key] {
				old, ok := exported[fmt.Sprintf("context/%s/%s@%d", g.Family, key, d)]
				if !ok {
					t.Fatalf("existing profile %s lacks a %dx export", key, d)
				}
				path, expected = old.path, old.sha256
			}
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
			if expected != "" && hash != expected {
				t.Fatalf("regenerated hash differs: %s", path)
			}
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
		if err := os.WriteFile(filepath.Join(root, "art/pipeline/contexts/"+string(f)+".json"), append(b, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("defined campaign profiles: %d; newly exported by family: %v", len(keys), added)
}

func loadCampaignRenderers(root string) (*groundRenderer, *forestRenderer, *hillsRenderer, error) {
	gi, err := loadGroundInputs(filepath.Join(root, "art/pipeline/ground-terrain-v2.json"))
	if err != nil {
		return nil, nil, nil, err
	}
	materials := map[string]*image.NRGBA{}
	for id, s := range gi.Sources {
		materials[id] = s.Image
	}
	ground := newGroundRenderer(materials)
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
