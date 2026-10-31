package hlasset

import (
	"bytes"
	"cmp"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/terraingrade"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/terrainlight"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/worldmaterial"
)

// Terrain tiles are 96×80 logical pixels registered on the hex centre.
const (
	tileWidth, tileHeight = 96, 80
)

var (
	terrainDensities = []int{2, 4}
	terrainAnchor    = [2]float64{48, 40}
	terrainGroups    = []string{"backgrounds", "infrastructure", "context-ground", "context-forest", "context-hills"}
	// optionalGroups are built only when the index has their entries.
	optionalGroups = []string{standingGroup, standingStampGroup, stampGroup, worldmaterial.Group}
)

const standingGroup, standingStampGroup, stampGroup = "standing", "standing-stamps", "stamps"

// terrainPack is one built terrain group at one density.
type terrainPack struct {
	Group    string // backgrounds, infrastructure, context-<family>, standing, standing-stamps, stamps or materials
	Density  int
	Manifest assets.Manifest
}

func terrainDir(d int, group string) string { return path.Join(fmt.Sprintf("%dx", d), group) }

// planTerrain returns the terrain packs of the index, every group at every
// density, before any image is read. Tiles are registered on the hex centre;
// context tiles carry the family of their group. Standing, stamp and material
// groups are optional (optionalProblem).
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
			if group == worldmaterial.Group {
				ids := map[string]bool{}
				for _, e := range m.Entries {
					ids[e.ID] = true
				}
				if len(m.Entries) != 2 || !ids[worldmaterial.MeadowID] || !ids[worldmaterial.ScrubID] {
					return nil, fmt.Errorf("%s: requires both summer grass and hills materials exactly once", id)
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
	return splitTerrain(packs)
}

// optionalProblem says why entry e cannot be in the optional group at density d,
// or "" when it can. Materials use the fixed world period and zero anchor.
// Standing items stay within the overhang
// (mapcompose.StandingFits) unless they belong to a cluster stamp; stamps span
// their cluster from the anchor on its first hex.
func optionalProblem(group string, e assets.Entry, d int) string {
	if e.Density != d {
		return fmt.Sprintf("is not a %dx file", d)
	}
	if group == worldmaterial.Group {
		if e.ID != worldmaterial.MeadowID && e.ID != worldmaterial.ScrubID {
			return "is not a summer grass or hills material"
		}
		if e.Width != worldmaterial.PeriodWidth*d || e.Height != worldmaterial.PeriodHeight*d || e.Anchor != [2]float64{} {
			return "is not a 568x400 world material with zero anchor"
		}
		return ""
	}
	if group == stampGroup {
		if !strings.HasPrefix(e.ID, "stamp/") {
			return "is not a stamp/<kind>/<shape>/<variant>"
		}
		return ""
	}
	if group == standingStampGroup && !strings.HasPrefix(e.ID, "standing/stamp/") {
		return "is not a standing/stamp/<kind>/<shape>/<variant>"
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
// created; out must be new. The images are graded (terraingrade), the painted
// cast shadows drawn with the one light (terrainlight), and out/grade names
// both. The 1x packs are derived from the graded 2x ones (deriveTerrain).
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
	grade, err := terrainRecoder(assetsDir, entries)
	if err != nil {
		return nil, err
	}
	written, err := writePacks(assetsDir, out, manifests, func(i int) string { return terrainDir(packs[i].Density, packs[i].Group) }, grade)
	if err != nil {
		return nil, err
	}
	for i := range packs {
		packs[i].Manifest = written[i]
	}
	derived, err := deriveTerrain(out, packs)
	if err != nil {
		_ = os.RemoveAll(out)
		return nil, err
	}
	packs = append(derived, packs...)
	if err := os.WriteFile(filepath.Join(out, "grade"), []byte(terraingrade.Version+"\n"+terrainlight.Version+"\n"), 0o644); err != nil {
		return nil, err
	}
	return packs, nil
}

// terrainRecoder returns the recoder that grades each terrain image by its
// art family, against the meadow tile of its density and season, and draws each
// painted cast shadow from its standing item with the one light (terrainlight).
func terrainRecoder(assetsDir string, entries []indexEntry) (recoder, error) {
	art := map[string]string{}
	meadow := map[string]image.Image{}
	for _, e := range entries {
		art[e.ID] = e.Art
		if e.ID == "tile/grass" || e.ID == "tile/grass_w" {
			b, err := os.ReadFile(filepath.Join(assetsDir, filepath.FromSlash(e.Path)))
			if err != nil {
				return nil, err
			}
			if meadow[fmt.Sprintf("%d/%s", e.Density, e.ID)], err = png.Decode(bytes.NewReader(b)); err != nil {
				return nil, fmt.Errorf("%s: %w", e.Path, err)
			}
		}
	}
	return func(e assets.Entry, p *assets.Pack) ([]byte, error) {
		var buf bytes.Buffer
		if standing, opacity, ok := terrainlight.Recast(e.ID, art[e.ID]); ok {
			im, err := p.Image(standing, e.Density)
			if err != nil {
				return nil, fmt.Errorf("the shadow %s: %w", e.ID, err)
			}
			err = png.Encode(&buf, terrainlight.Cast(im, e.Density, terrainlight.Diameter(im, e.Density), opacity))
			return buf.Bytes(), err
		}
		if !terraingrade.Applies(e.ID, art[e.ID]) {
			err := p.WritePNG(&buf, e.ID, e.Density)
			return buf.Bytes(), err
		}
		im, err := p.Image(e.ID, e.Density)
		if err != nil {
			return nil, err
		}
		err = png.Encode(&buf, terraingrade.Grade(im, e.ID, art[e.ID], meadow[fmt.Sprintf("%d/%s", e.Density, terraingrade.Meadow(e.ID))]))
		return buf.Bytes(), err
	}, nil
}

// terrainGaps composes every shipped map (a scenario's cp setup) against the context tiles of the
// catalog in dir and lists each hex drawn with a default tile instead of its
// context, plus stray compound pieces, for the art track.
func terrainGaps(dir string, w io.Writer) error {
	have := map[string]bool{}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	groups, err := assets.PackGroups(root.FS(), terrainGroups, optionalGroups, terrainDensities)
	if err != nil {
		return err
	}
	manifests := map[string][]assets.Manifest{}
	for _, g := range groups {
		for _, d := range terrainDensities {
			fsys, err := fs.Sub(root.FS(), terrainDir(d, g))
			if err != nil {
				return err
			}
			p, err := assets.Open(fsys, "manifest.json", assets.Options{AllowCandidates: true})
			if err != nil {
				return err
			}
			manifests[g] = append(manifests[g], p.Manifest())
		}
	}
	if err := assets.CheckGroupDensities(manifests, terrainDensities); err != nil {
		return err
	}
	for _, ms := range manifests {
		for _, e := range ms[0].Entries {
			have[e.ID] = true
		}
	}
	shipped, err := scenario.Shipped(scenario.CP)
	if err != nil {
		return err
	}
	for _, sh := range shipped {
		name, l := sh.Name, sh.Level
		s := mapcompose.Compose(l.Grid(), mapcompose.LevelHex(l), func(id string) bool { return have[id] })
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
			name, selected, count[mapcompose.Missing], count[mapcompose.Undefined], len(s.Issues)); err != nil {
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
