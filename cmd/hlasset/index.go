package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
)

// maxIndex bounds assets/index.json.
const maxIndex = 16 << 20

// indexEntry is one runtime file of assets/index.json. Packed PNGs name their
// pack; all paths are relative to the assets directory.
type indexEntry struct {
	ID      string      `json:"id"`
	Pack    string      `json:"pack,omitempty"`
	Density int         `json:"density,omitempty"`
	Path    string      `json:"path"`
	SHA256  string      `json:"sha256"`
	Width   int         `json:"width,omitempty"`
	Height  int         `json:"height,omitempty"`
	Anchor  *[2]float64 `json:"logical_anchor,omitempty"`
	Art     string      `json:"art"`
}

// readIndex reads dir/index.json.
func readIndex(dir string) ([]indexEntry, error) {
	f, err := os.Open(filepath.Join(dir, "index.json"))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	b, err := io.ReadAll(io.LimitReader(f, maxIndex+1))
	if err != nil {
		return nil, err
	}
	if len(b) > maxIndex {
		return nil, fmt.Errorf("%s/index.json: larger than %d bytes", dir, maxIndex)
	}
	var index struct {
		Schema int          `json:"schema_version"`
		Assets []indexEntry `json:"assets"`
	}
	if err := json.Unmarshal(b, &index); err != nil {
		return nil, fmt.Errorf("%s/index.json: %w", dir, err)
	}
	if index.Schema != 1 {
		return nil, fmt.Errorf("%s/index.json: schema_version %d, want 1", dir, index.Schema)
	}
	for _, e := range index.Assets {
		if e.ID == "" || !filepath.IsLocal(e.Path) || strings.Contains(e.Path, `\`) {
			return nil, fmt.Errorf("index entry %q: invalid id or path %q", e.ID, e.Path)
		}
		if e.Pack != "" && (e.Anchor == nil || e.Density == 0) {
			return nil, fmt.Errorf("index entry %s: packed without density or anchor", e.ID)
		}
	}
	return index.Assets, nil
}

// indexPacks returns the candidate manifest of every pack whose ID starts with
// prefix, in index order; each pack's entries keep their index order.
func indexPacks(entries []indexEntry, prefix string) []assets.Manifest {
	var packs []assets.Manifest
	at := map[string]int{}
	for _, e := range entries {
		if e.Pack == "" || !strings.HasPrefix(e.Pack, prefix) {
			continue
		}
		i, ok := at[e.Pack]
		if !ok {
			i = len(packs)
			at[e.Pack] = i
			packs = append(packs, assets.Manifest{Schema: 1, ID: e.Pack, Status: "candidate"})
		}
		packs[i].Entries = append(packs[i].Entries, assets.Entry{
			ID: e.ID, Path: e.Path, SHA256: e.SHA256, Width: e.Width, Height: e.Height, Density: e.Density, Anchor: *e.Anchor,
		})
	}
	return packs
}

// loadPack loads m from the assets directory root and hands it to use; the
// decoded pack is dropped afterwards so only one is held at a time.
func loadPack(root *os.Root, m assets.Manifest, use func(*assets.Pack) error) error {
	p, err := assets.Load(root.FS(), m, assets.Options{AllowCandidates: true})
	if err != nil {
		return fmt.Errorf("%s: %w", m.ID, err)
	}
	if err := use(p); err != nil {
		return fmt.Errorf("%s: %w", m.ID, err)
	}
	return nil
}

// writePacks validates every pack, then writes each to dirOf(pack) under the new
// directory out. Nothing is created when a pack is invalid.
func writePacks(assetsDir, out string, packs []assets.Manifest, dirOf func(int) string) ([]assets.Manifest, error) {
	r, err := os.OpenRoot(assetsDir)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	for _, m := range packs {
		if err := loadPack(r, m, func(p *assets.Pack) error { return p.Validate() }); err != nil {
			return nil, err
		}
	}
	o, err := newOutput(out)
	if err != nil {
		return nil, err
	}
	defer o.close()
	written := make([]assets.Manifest, len(packs))
	for i, m := range packs {
		if err := loadPack(r, m, func(p *assets.Pack) error {
			w, err := o.writePack(dirOf(i), p, m)
			written[i] = w
			return err
		}); err != nil {
			return nil, err
		}
	}
	o.done()
	return written, nil
}
