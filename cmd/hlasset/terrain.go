package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
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
	terrainGroups    = []string{"backgrounds", "infrastructure", "context-ground", "context-forest", "context-hills"}
)

// terrainPack is one built terrain group at one density.
type terrainPack struct {
	Group    string // backgrounds, infrastructure or context-<family>
	Density  int
	Manifest assets.Manifest
}

func terrainDir(d int, group string) string { return path.Join(fmt.Sprintf("%dx", d), group) }

// planTerrain returns the terrain packs of the index, every group at every
// density, before any image is read. Tiles are registered on the hex centre;
// context tiles carry the family of their group.
func planTerrain(entries []indexEntry) ([]terrainPack, error) {
	byID := map[string]assets.Manifest{}
	for _, m := range indexPacks(entries, "terrain-") {
		byID[m.ID] = m
	}
	var packs []terrainPack
	for _, group := range terrainGroups {
		prefix := "tile/"
		if f, ok := strings.CutPrefix(group, "context-"); ok {
			prefix = "context/" + f + "/"
		}
		for _, d := range terrainDensities {
			id := fmt.Sprintf("terrain-%s-%dx", group, d)
			m, ok := byID[id]
			if !ok {
				return nil, fmt.Errorf("index has no %s pack", id)
			}
			delete(byID, id)
			for _, e := range m.Entries {
				if !strings.HasPrefix(e.ID, prefix) || e.Density != d || e.Width != d*tileWidth || e.Height != d*tileHeight || e.Anchor != terrainAnchor {
					return nil, fmt.Errorf("%s: %s is not a %s %dx tile (%dx%d, anchor %v)", id, e.ID, prefix, d, e.Width, e.Height, e.Anchor)
				}
			}
			packs = append(packs, terrainPack{Group: group, Density: d, Manifest: m})
		}
	}
	for id := range byID {
		return nil, fmt.Errorf("index has unknown terrain pack %s", id)
	}
	return packs, nil
}

// buildTerrain builds the terrain catalog: out/<d>x/<group>/ holds one pack per
// group and density, each inside the loader's decoded-image budget (all 4x
// terrain in one pack would not be). Every pack is validated before out is
// created; out must be new.
func buildTerrain(assetsDir, out string) ([]terrainPack, error) {
	entries, err := readIndex(assetsDir)
	if err != nil {
		return nil, err
	}
	packs, err := planTerrain(entries)
	if err != nil {
		return nil, err
	}
	manifests := make([]assets.Manifest, len(packs))
	for i, p := range packs {
		manifests[i] = p.Manifest
	}
	written, err := writePacks(assetsDir, out, manifests, func(i int) string { return terrainDir(packs[i].Density, packs[i].Group) })
	if err != nil {
		return nil, err
	}
	for i := range packs {
		packs[i].Manifest = written[i]
	}
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
