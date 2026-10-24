//go:build campaigngfx

package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

type campaignEntry struct {
	ID      string     `json:"id"`
	Pack    string     `json:"pack"`
	Density int        `json:"density"`
	Path    string     `json:"path"`
	SHA256  string     `json:"sha256"`
	Width   int        `json:"width"`
	Height  int        `json:"height"`
	Anchor  [2]float64 `json:"logical_anchor"`
	Art     string     `json:"art"`
}

func campaignExportKey(id string, d int) string { return fmt.Sprintf("%s@%d", id, d) }

func readCampaignIndex(root string) ([]json.RawMessage, map[string]campaignEntry, error) {
	b, err := os.ReadFile(filepath.Join(root, "assets/index.json"))
	if err != nil {
		return nil, nil, err
	}
	var index struct {
		Schema int               `json:"schema_version"`
		Assets []json.RawMessage `json:"assets"`
	}
	if err := json.Unmarshal(b, &index); err != nil {
		return nil, nil, err
	}
	if index.Schema != 1 {
		return nil, nil, fmt.Errorf("unsupported runtime index schema")
	}
	exported := map[string]campaignEntry{}
	for _, raw := range index.Assets {
		var e campaignEntry
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, nil, err
		}
		if !strings.HasPrefix(e.Pack, "terrain-") {
			continue
		}
		k := campaignExportKey(e.ID, e.Density)
		if _, ok := exported[k]; ok {
			return nil, nil, fmt.Errorf("duplicate export %s", k)
		}
		exported[k] = e
	}
	return index.Assets, exported, nil
}

// Actual exact-density runtime IDs decide availability, including stamps that
// replace contexts. Profile membership alone does not mean an export exists.
func campaignContexts(has func(string) bool) (map[string]mapcompose.Gap, error) {
	contexts := map[string]mapcompose.Gap{}
	for _, source := range []struct {
		fsys    fs.FS
		pattern string
	}{{levels.FS, "*.json"}, {levels.Campaign, "campaign/*.json"}} {
		names, err := fs.Glob(source.fsys, source.pattern)
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			b, err := fs.ReadFile(source.fsys, name)
			if err != nil {
				return nil, err
			}
			l, err := level.Parse(b)
			if err != nil {
				return nil, err
			}
			s := mapcompose.Compose(l.Grid(), func(c hexgrid.Coord) mapcompose.Hex {
				h := l.Lines[c.Line][c.Col]
				return mapcompose.Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
			}, has)
			if len(s.Issues) != 0 {
				return nil, fmt.Errorf("%s: compound issues: %v", name, s.Issues)
			}
			for _, g := range s.Gaps {
				if g.Family == mapcompose.Shore {
					continue // TestGenerateCampaignShores
				}
				if g.Reason != mapcompose.Missing {
					return nil, fmt.Errorf("%s: undefined context: %+v", name, g)
				}
				contexts[mapcompose.Key(g.Descriptor)] = g
			}
		}
	}
	return contexts, nil
}

func TestCampaignContextDiscovery(t *testing.T) {
	all, err := campaignContexts(func(string) bool { return false })
	if err != nil {
		t.Fatal(err)
	}
	// The campaign contains far more transitions than the three shipped maps.
	if len(all) < 1500 {
		t.Fatalf("only %d contexts discovered; campaign levels missing", len(all))
	}
	for key, g := range all {
		if key != mapcompose.Key(g.Descriptor) || g.Reason != mapcompose.Missing {
			t.Fatalf("invalid context %+v", g)
		}
	}
	none, err := campaignContexts(func(string) bool { return true })
	if err != nil || len(none) != 0 {
		t.Fatalf("available contexts were selected for regeneration: %d, %v", len(none), err)
	}
}

func TestCampaignPreservesLayeredSummerForest(t *testing.T) {
	_, exported, err := readCampaignIndex(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	for mask := range 64 {
		neighbors := make([]string, 6)
		for i := range neighbors {
			neighbors[i] = "grass"
			if mask&(1<<i) != 0 {
				neighbors[i] = "forest"
			}
		}
		c, err := newForestContext("forest", neighbors)
		if err != nil {
			t.Fatal(err)
		}
		for _, id := range []string{"context/forest/" + c.key(), "standing/context/forest/" + c.key(), "standing/shadow/context/forest/" + c.key()} {
			for _, d := range []int{2, 4} {
				if _, ok := exported[campaignExportKey(id, d)]; !ok {
					t.Fatalf("selected layered forest export missing: %s at %dx", id, d)
				}
			}
		}
	}
}
