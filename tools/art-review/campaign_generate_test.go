//go:build campaigngfx

package main

// Explicit offline recipe: go test -tags campaigngfx ./tools/art-review -run TestGenerateCampaignContexts -count=1 -v
// Normal tests never write artwork. Existing images are checked byte for byte.
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
	recipes := map[mapcompose.Family]map[string]any{}
	known := map[string]map[string]any{}
	for _, f := range mapcompose.Families {
		r := read("art/recipes/" + string(f) + "-terrain-v1.json")
		recipes[f] = r
		for _, p := range r["profiles"].([]any) {
			profile := p.(map[string]any)
			known[profile["context_key"].(string)] = profile
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
	gi, err := loadGroundInputs(filepath.Join(root, "art/pipeline/ground-terrain-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	materials := map[string]*image.NRGBA{}
	for id, s := range gi.Sources {
		materials[id] = s.Image
	}
	ground := newGroundRenderer(materials)
	ground.mudFeature = gi.Manifest.MudFeature
	fi, err := loadForestInputs(filepath.Join(root, "art/pipeline/forest-terrain-v1.json"))
	if err != nil {
		t.Fatal(err)
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
		t.Fatal(err)
	}
	hi, err := loadHillsInputs(filepath.Join(root, "art/pipeline/hills-terrain-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	hills := newHillsRenderer(hi.Sources["grass"].Image, hi.Sources["height"].Image)
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
			p["instances"] = forest.instances(c)
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
			p["isolated_cue"] = hills.cueFor(c)
		}
		if err != nil {
			t.Fatal(err)
		}
		var outputs []any
		for _, d := range []int{2, 4} {
			path := fmt.Sprintf("assets/terrain/%s/context/%s-%dx.png", g.Family, key, d)
			var expected string
			if old := known[key]; old != nil {
				for _, o := range old["outputs"].([]any) {
					out := o.(map[string]any)
					if int(out["dimensions"].(map[string]any)["width"].(float64)) == 96*d {
						path = out["path"].(string)
						expected = out["sha256"].(string)
					}
				}
				if expected == "" {
					t.Fatal("existing profile lacks density")
				}
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
			outputs = append(outputs, map[string]any{"path": path, "sha256": hash, "dimensions": dimensions{96 * d, 80 * d}})
		}
		if known[key] != nil {
			continue
		}
		p["outputs"] = outputs
		r := recipes[g.Family]
		r["profiles"] = append(r["profiles"].([]any), p)
		field := "context_exports"
		if g.Family == mapcompose.Ground {
			field = "exports"
		}
		for _, o := range outputs {
			r[field] = append(r[field].([]any), o)
		}
		added[g.Family]++
	}
	for _, f := range mapcompose.Families {
		if added[f] == 0 {
			continue
		}
		r := recipes[f]
		r["campaign_context_generation"] = "go test -tags campaigngfx ./tools/art-review -run TestGenerateCampaignContexts -count=1 -v"
		b, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "art/recipes/"+string(f)+"-terrain-v1.json"), append(b, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("defined campaign profiles: %d; newly exported by family: %v", len(keys), added)
}
