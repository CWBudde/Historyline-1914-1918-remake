package main

import (
	"cmp"
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
	// optionalGroups are built only when the index has standing items or cluster
	// stamps.
	optionalGroups = []string{standingGroup, stampGroup}
)

const standingGroup, stampGroup = "standing", "stamps"

// terrainPack is one built terrain group at one density.
type terrainPack struct {
	Group    string // backgrounds, infrastructure, context-<family>, standing or stamps
	Density  int
	Manifest assets.Manifest
}

func terrainDir(d int, group string) string { return path.Join(fmt.Sprintf("%dx", d), group) }

// planTerrain returns the terrain packs of the index, every group at every
// density, before any image is read. Tiles are registered on the hex centre;
// context tiles carry the family of their group. The standing and stamp groups
// are optional (optionalProblem).
func planTerrain(entries []indexEntry) ([]terrainPack, error) {
	for _, e := range entries {
		if !strings.HasPrefix(e.Pack, "terrain-") {
			continue
		}
		want := "ground"
		if strings.HasPrefix(e.Pack, "terrain-standing-") {
			want = "standing"
		}
		if layer := cmp.Or(e.Layer, "ground"); layer != want {
			return nil, fmt.Errorf("%s: %s has layer %s, want %s", e.Pack, e.ID, layer, want)
		}
	}
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
	for _, group := range optionalGroups {
		var built []terrainPack
		for _, d := range terrainDensities {
			id := fmt.Sprintf("terrain-%s-%dx", group, d)
			m, ok := byID[id]
			if !ok {
				continue
			}
			delete(byID, id)
			for _, e := range m.Entries {
				if msg := optionalProblem(group, e, d); msg != "" {
					return nil, fmt.Errorf("%s: %s %s (%dx%d, anchor %v)", id, e.ID, msg, e.Width, e.Height, e.Anchor)
				}
			}
			built = append(built, terrainPack{Group: group, Density: d, Manifest: m})
		}
		if len(built) != 0 && len(built) != len(terrainDensities) {
			return nil, fmt.Errorf("index has the %s pack at %d of %d densities", group, len(built), len(terrainDensities))
		}
		packs = append(packs, built...)
	}
	for id := range byID {
		return nil, fmt.Errorf("index has unknown terrain pack %s", id)
	}
	return packs, nil
}

// optionalProblem says why entry e cannot be in the optional group at density d,
// or "" when it can. Standing items stay within the overhang
// (mapcompose.StandingFits) unless they belong to a cluster stamp; stamps span
// their cluster from the anchor on its first hex.
func optionalProblem(group string, e assets.Entry, d int) string {
	if e.Density != d {
		return fmt.Sprintf("is not a %dx file", d)
	}
	if group == stampGroup {
		if !strings.HasPrefix(e.ID, "stamp/") {
			return "is not a stamp/<kind>/<shape>/<variant>"
		}
		return ""
	}
	if !strings.HasPrefix(e.ID, "standing/") {
		return "is not a standing/<id>"
	}
	w, h := float64(e.Width)/float64(d), float64(e.Height)/float64(d)
	if !strings.HasPrefix(e.ID, "standing/stamp/") && !mapcompose.StandingFits(w, h, e.Anchor) {
		return "overhangs its hex by more than half a hex"
	}
	return ""
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
	groups := []string{"backgrounds"}
	for _, f := range mapcompose.Families {
		groups = append(groups, "context-"+string(f))
	}
	if _, err := os.Stat(filepath.Join(dir, terrainDir(terrainDensities[0], stampGroup))); err == nil {
		groups = append(groups, stampGroup)
	}
	for _, g := range groups {
		r, err := os.OpenRoot(filepath.Join(dir, terrainDir(terrainDensities[0], g)))
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
		// Lakes use reusable shoreline tiles. Other feature clusters can still
		// request a stamp when their shape has no painting yet.
		for _, cl := range s.Clusters {
			if cl.Stamp != "" || cl.Kind == "lake" {
				continue
			}
			c := cl.Cells[0]
			if _, err := fmt.Fprintf(w, "  unstamped %s %s (line %d, col %d): %s\n", cl.Kind, cl.Shape, c.Line, c.Col, mapcompose.StampID(cl.Kind, cl.Shape, 1)); err != nil {
				return err
			}
		}
	}
	return nil
}
