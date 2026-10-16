package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// Terrain tiles are 96×80 logical pixels registered on the hex centre.
const (
	tileWidth, tileHeight = 96, 80
)

var (
	terrainDensities = []int{2, 4}
	terrainAnchor    = [2]float64{48, 40}
)

// terrainPack is one built terrain group at one density.
type terrainPack struct {
	Group    string // backgrounds, infrastructure or context-<family>
	Density  int
	Manifest assets.Manifest
}

func terrainDir(d int, group string) string { return path.Join(fmt.Sprintf("%dx", d), group) }

// pin is a recipe's record of one exported PNG.
type pin struct {
	SHA256        string
	Width, Height int
}

// recipePins collects every {path, sha256, dimensions} record in a recipe,
// wherever it is nested (exports, profile outputs, context exports).
func recipePins(b []byte, pins map[string]pin) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var walk func(any) error
	walk = func(v any) error {
		switch x := v.(type) {
		case map[string]any:
			p, okP := x["path"].(string)
			sum, okS := x["sha256"].(string)
			dim, okD := x["dimensions"].(map[string]any)
			if okP && okS && okD && strings.HasSuffix(p, ".png") {
				w, _ := dim["width"].(float64)
				h, _ := dim["height"].(float64)
				n := pin{sum, int(w), int(h)}
				if old, ok := pins[p]; ok && old != n {
					return fmt.Errorf("%s pinned twice with different values", p)
				}
				pins[p] = n
			}
			for _, e := range x {
				if err := walk(e); err != nil {
					return err
				}
			}
		case []any:
			for _, e := range x {
				if err := walk(e); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return walk(v)
}

// terrainEntry turns a pinned export into a pack entry; its density follows from
// the pinned size, which the file name must agree with.
func terrainEntry(id, p string, pins map[string]pin) (assets.Entry, error) {
	pn, ok := pins[p]
	if !ok {
		return assets.Entry{}, fmt.Errorf("%s: %s has no pin in the recipes", id, p)
	}
	d := pn.Width / tileWidth
	if pn.Width != d*tileWidth || pn.Height != d*tileHeight || !slices.Contains(terrainDensities, d) {
		return assets.Entry{}, fmt.Errorf("%s: %s is %dx%d, not a 2x/4x tile", id, p, pn.Width, pn.Height)
	}
	if !strings.HasSuffix(p, fmt.Sprintf("-%dx.png", d)) {
		return assets.Entry{}, fmt.Errorf("%s: %s does not name density %dx", id, p, d)
	}
	return assets.Entry{ID: id, Path: p, SHA256: pn.SHA256, Width: pn.Width, Height: pn.Height, Density: d, Anchor: terrainAnchor}, nil
}

// planTerrain reads the coverage ledger and the recipes and returns the packs to
// build, before any image is read. Every tile needs a pinned export at every
// density; every context profile must reproduce its key from its neighbours.
func planTerrain(coverage, recipes string) ([]terrainPack, error) {
	b, err := readFileBounded(coverage, maxRecipe)
	if err != nil {
		return nil, err
	}
	rows, err := csv.NewReader(bytes.NewReader(b)).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", coverage, err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("%s: no rows", coverage)
	}
	col := map[string]int{}
	for i, h := range rows[0] {
		col[h] = i
	}
	for _, h := range []string{"kind", "runtime_id", "export_2x", "export_4x", "build_manifest"} {
		if _, ok := col[h]; !ok {
			return nil, fmt.Errorf("%s: no %s column", coverage, h)
		}
	}
	pins := map[string]pin{}
	read := map[string]bool{}
	readRecipe := func(name string) ([]byte, error) {
		b, err := readFileBounded(filepath.Join(recipes, name), maxRecipe)
		if err != nil {
			return nil, err
		}
		if !read[name] {
			read[name] = true
			if err := recipePins(b, pins); err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
		}
		return b, nil
	}
	entries := map[string][]assets.Entry{} // group/density → entries
	add := func(group string, e assets.Entry) {
		k := terrainDir(e.Density, group)
		entries[k] = append(entries[k], e)
	}
	seen := map[string]bool{}
	for _, r := range rows[1:] {
		group := map[string]string{"background": "backgrounds", "infrastructure": "infrastructure"}[r[col["kind"]]]
		id := mapcompose.TileID(r[col["runtime_id"]])
		if group == "" || seen[id] {
			return nil, fmt.Errorf("%s: row %s has kind %q or repeats", coverage, id, r[col["kind"]])
		}
		seen[id] = true
		if _, err := readRecipe(path.Base(r[col["build_manifest"]])); err != nil {
			return nil, err
		}
		for _, d := range terrainDensities {
			p := r[col[fmt.Sprintf("export_%dx", d)]]
			if p == "" {
				return nil, fmt.Errorf("%s: no %dx export", id, d)
			}
			e, err := terrainEntry(id, p, pins)
			if err != nil {
				return nil, err
			}
			if e.Density != d {
				return nil, fmt.Errorf("%s: export_%dx %s is %dx", id, d, p, e.Density)
			}
			add(group, e)
		}
	}
	for _, f := range mapcompose.Families {
		b, err := readRecipe(string(f) + "-terrain-v1.json")
		if err != nil {
			return nil, err
		}
		var recipe struct {
			Profiles []struct {
				RuntimeID  string    `json:"runtime_id"`
				Neighbors  [6]string `json:"neighbor_ids"`
				Phase      *[2]int   `json:"phase"`
				Descriptor string    `json:"descriptor"`
				Key        string    `json:"context_key"`
				Outputs    []struct {
					Path string `json:"path"`
				} `json:"outputs"`
			} `json:"profiles"`
		}
		if err := json.Unmarshal(b, &recipe); err != nil {
			return nil, err
		}
		if len(recipe.Profiles) == 0 {
			return nil, fmt.Errorf("%s recipe has no context profiles", f)
		}
		group := "context-" + string(f)
		for _, p := range recipe.Profiles {
			var phase [2]int
			if p.Phase != nil {
				phase = *p.Phase
			}
			d, ok := mapcompose.Descriptor(f, p.RuntimeID, p.Neighbors, phase)
			if !ok || d != p.Descriptor || mapcompose.Key(d) != p.Key {
				return nil, fmt.Errorf("%s profile %s: descriptor or context key %s does not follow the context contract", f, p.RuntimeID, p.Key)
			}
			id := mapcompose.ContextID(f, p.Key)
			if seen[id] {
				return nil, fmt.Errorf("%s: duplicate context profile", id)
			}
			seen[id] = true
			have := map[int]bool{}
			for _, o := range p.Outputs {
				e, err := terrainEntry(id, o.Path, pins)
				if err != nil {
					return nil, err
				}
				have[e.Density] = true
				add(group, e)
			}
			for _, d := range terrainDensities {
				if !have[d] {
					return nil, fmt.Errorf("%s: no %dx output", id, d)
				}
			}
		}
	}
	var packs []terrainPack
	for _, group := range []string{"backgrounds", "infrastructure", "context-ground", "context-forest", "context-hills"} {
		for _, d := range terrainDensities {
			packs = append(packs, terrainPack{Group: group, Density: d, Manifest: assets.Manifest{
				Schema: 1, ID: fmt.Sprintf("terrain-%s-%dx", group, d), Status: "candidate",
				Entries: entries[terrainDir(d, group)],
			}})
		}
	}
	return packs, nil
}

// buildTerrain builds the terrain catalog: out/<d>x/<group>/ holds one pack per
// group and density, each inside the loader's decoded-image budget (all 4x
// terrain in one pack would not be). Export paths are relative to root. Every
// pack is validated before out is created; out must be new.
func buildTerrain(coverage, recipes, root, out string) ([]terrainPack, error) {
	packs, err := planTerrain(coverage, recipes)
	if err != nil {
		return nil, err
	}
	r, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	with := func(p terrainPack, use func(*assets.Pack) error) error {
		pack, err := assets.Load(r.FS(), p.Manifest, assets.Options{AllowCandidates: true})
		if err != nil {
			return fmt.Errorf("%s: %w", p.Manifest.ID, err)
		}
		if err := use(pack); err != nil {
			return fmt.Errorf("%s: %w", p.Manifest.ID, err)
		}
		return nil
	}
	for _, p := range packs {
		if err := with(p, func(pack *assets.Pack) error { return pack.Validate() }); err != nil {
			return nil, err
		}
	}
	o, err := newOutput(out)
	if err != nil {
		return nil, err
	}
	defer o.close()
	for i := range packs {
		p := &packs[i]
		if err := with(*p, func(pack *assets.Pack) error {
			m, err := o.writePack(terrainDir(p.Density, p.Group), pack, p.Manifest)
			p.Manifest = m
			return err
		}); err != nil {
			return nil, err
		}
	}
	o.done()
	return packs, nil
}

// terrainGaps composes every shipped level against the context tiles of the
// catalog in dir and lists each hex drawn with a default tile instead of its
// context, plus stray compound pieces, for the art track.
func terrainGaps(dir string, w io.Writer) error {
	have := map[string]bool{}
	for _, f := range mapcompose.Families {
		r, err := os.OpenRoot(filepath.Join(dir, terrainDir(terrainDensities[0], "context-"+string(f))))
		if err != nil {
			return err
		}
		p, err := assets.Open(r.FS(), "manifest.json", assets.Options{AllowCandidates: true})
		_ = r.Close()
		if err != nil {
			return err
		}
		for _, e := range p.Manifest().Entries {
			have[e.ID] = true
		}
	}
	names, err := fs.Glob(levels.FS, "*.json")
	if err != nil {
		return err
	}
	for _, name := range names {
		b, err := fs.ReadFile(levels.FS, name)
		if err != nil {
			return err
		}
		l, err := level.Parse(b)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		s := mapcompose.Compose(l.Grid(), func(c hexgrid.Coord) mapcompose.Hex {
			h := l.Lines[c.Line][c.Col]
			return mapcompose.Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
		}, func(id string) bool { return have[id] })
		selected, count := 0, map[mapcompose.Reason]int{}
		for _, line := range s.Cells {
			for _, c := range line {
				if c.Context != "" {
					selected++
				}
			}
		}
		for _, g := range s.Gaps {
			count[g.Reason]++
		}
		if _, err := fmt.Fprintf(w, "%s: %d context tiles selected, %d missing, %d undefined, %d compound issues\n",
			strings.TrimSuffix(name, ".json"), selected, count[mapcompose.Missing], count[mapcompose.Undefined], len(s.Issues)); err != nil {
			return err
		}
		for _, g := range s.Gaps {
			why := g.Detail
			if g.Reason == mapcompose.Missing {
				why = strings.ReplaceAll(strings.TrimSuffix(g.Descriptor, "\n"), "\n", ";")
			}
			if _, err := fmt.Fprintf(w, "  %s %s %s (line %d, col %d): %s\n", g.Reason, g.Family, g.Terrain, g.At.Line, g.At.Col, why); err != nil {
				return err
			}
		}
		for _, is := range s.Issues {
			if _, err := fmt.Fprintf(w, "  compound (line %d, col %d): %s\n", is.At.Line, is.At.Col, is.Detail); err != nil {
				return err
			}
		}
	}
	return nil
}
