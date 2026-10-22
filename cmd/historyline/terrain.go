package main

import (
	"cmp"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
)

// tilePad is the border around terrain textures, in texels. Background tiles
// own complementary pixels, but scaled, filtered quads do not meet exactly
// (Ebitengine snaps quad corners to a sub-pixel grid), which leaves seams.
// Growing each background's opaque edge into the border makes neighbours
// overlap instead; overlays keep a transparent border.
const tilePad = 3

// terrainGroups are the pack groups hlasset build-terrain writes; the optional
// groups only when the catalog has standing items or cluster stamps.
var (
	terrainGroups  = []string{"backgrounds", "infrastructure", "context-ground", "context-forest", "context-hills"}
	optionalGroups = []string{standingGroup, standingStampGroup, stampGroup}
)

const standingGroup, standingStampGroup, stampGroup = "standing", "standing-stamps", "stamps"

// terrainArt is the runtime terrain catalog and the level composed from it.
type terrainArt struct {
	*artPacks
	scene mapcompose.Scene
	group map[string]string // asset ID → pack group
}

// levelHex reads a hex's terrain from the level for the compositor.
func levelHex(l *level.Level) func(hexgrid.Coord) mapcompose.Hex {
	return func(c hexgrid.Coord) mapcompose.Hex {
		h := l.Lines[c.Line][c.Col]
		return mapcompose.Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
	}
}

// openTerrainArt opens the terrain catalog and composes lvl with it. Context
// tiles are used where the catalog has them at every density; every other tile
// the level draws, and the unseen tile, must be present, so drawing never meets
// a missing tile.
func openTerrainArt(fsys fs.FS, opt assets.Options, lvl *level.Level) (*terrainArt, error) {
	groups, err := assets.PackGroups(fsys, terrainGroups, optionalGroups, spriteDensities)
	if err != nil {
		return nil, fmt.Errorf("terrain %w", err)
	}
	packs, err := openPacks(fsys, opt, groups, spriteDensities)
	if err != nil {
		return nil, fmt.Errorf("terrain %w", err)
	}
	packs.pad = tilePad
	packs.bleed = map[string]bool{}
	for _, g := range groups {
		packs.bleed[g] = g == "backgrounds" || strings.HasPrefix(g, "backgrounds-") || strings.HasPrefix(g, "context-")
	}
	a := &terrainArt{artPacks: packs, group: map[string]string{}}
	manifests := map[string][]assets.Manifest{}
	for _, g := range groups {
		for _, d := range spriteDensities {
			manifests[g] = append(manifests[g], packs.packs[packKey{g, d}].Manifest())
		}
		for _, e := range packs.packs[packKey{g, spriteDensities[0]}].Manifest().Entries {
			a.group[e.ID] = g
		}
	}
	if err := assets.CheckGroupDensities(manifests, spriteDensities); err != nil {
		return nil, err
	}
	present := func(id string) error {
		g, ok := a.group[id]
		if !ok {
			return fmt.Errorf("terrain catalog has no tile %s", id)
		}
		for _, d := range spriteDensities {
			if _, err := a.entry(g, id, d); err != nil {
				return fmt.Errorf("terrain pack %dx/%s: %w", d, g, err)
			}
		}
		return nil
	}
	a.scene = mapcompose.Compose(lvl.Grid(), levelHex(lvl), func(id string) bool { return present(id) == nil })
	if err := present(mapcompose.TileID("unseen")); err != nil {
		return nil, err
	}
	var ids []string
	for _, line := range a.scene.Cells {
		for _, c := range line {
			ids = slices.Concat(ids, []string{c.Background, c.Under}, c.Overlays, c.Shadows, c.Standing)
		}
	}
	for _, cl := range a.scene.Clusters {
		ids = append(ids, cl.Stamp, cl.Standing)
	}
	for _, id := range ids {
		if id == "" {
			continue
		}
		if err := present(id); err != nil {
			return nil, err
		}
	}
	return a, nil
}

// loadTerrainArt opens the terrain catalog in dir for lvl, counting its fetches in m.
func loadTerrainArt(dir string, allowCandidates bool, lvl *level.Level, m *meter) (*terrainArt, error) {
	fsys, done, err := openArtDir(dir, m)
	if err != nil {
		return nil, err
	}
	defer done()
	return openTerrainArt(fsys, assets.Options{AllowCandidates: allowCandidates}, lvl)
}

// terrainLayers returns the tiles drawn on hex c of scene s given the player's
// fog: nothing of an unseen hex but the fog tile (its standing items would give
// its terrain away), and the remembered terrain of a seen hex dimmed to grey.
// Standing items of a known hex may overhang an unseen neighbour. A cluster is
// drawn as its stamp only while all its hexes are known; otherwise its known
// hexes keep their per-hex art.
func terrainLayers(s *mapcompose.Scene, c hexgrid.Coord, fog func(hexgrid.Coord) state.Fog) (bg, stamp string, overlays, standing []string, dim bool) {
	f := fog(c)
	if f == state.Unseen {
		return mapcompose.TileID("unseen"), "", nil, nil, false
	}
	known := func(cl *mapcompose.Cluster) bool {
		return !slices.ContainsFunc(cl.Cells, func(x hexgrid.Coord) bool { return fog(x) == state.Unseen })
	}
	bg, stamp, overlays, standing = s.Layers(c, known)
	return bg, stamp, overlays, standing, f == state.Seen
}

// remembered is the grey, darkened look of a seen but not visible hex.
var remembered = func() colorm.ColorM {
	var m colorm.ColorM
	m.ChangeHSV(0, 0, 0.5)
	return m
}()

// terrainShadows shares standing items' fog rules, including suppression under
// a known cluster stamp and dimming of remembered terrain.
func terrainShadows(s *mapcompose.Scene, c hexgrid.Coord, fog func(hexgrid.Coord) state.Fog) ([]string, bool) {
	f := fog(c)
	if f == state.Unseen {
		return nil, false
	}
	known := func(cl *mapcompose.Cluster) bool {
		return !slices.ContainsFunc(cl.Cells, func(x hexgrid.Coord) bool { return fog(x) == state.Unseen })
	}
	return s.ShadowsAt(c, known), f == state.Seen
}

// drawTile draws terrain tile id with its anchor on screen point (sx, sy).
func (g *game) drawTile(dst *ebiten.Image, id string, sx, sy float64, dim bool) {
	d := spriteDensity(g.cam.Effective())
	img, e, err := g.terrain.image(g.terrain.group[id], id, d)
	if err != nil {
		g.terrainErr = err.Error()
		return
	}
	var op colorm.DrawImageOptions
	op.GeoM = anchorGeoM(e, d, g.terrain.pad, g.cam.Effective(), sx, sy)
	op.Filter = ebiten.FilterLinear
	var cm colorm.ColorM
	if dim {
		cm = remembered
	}
	colorm.DrawImage(dst, img, cm, &op)
}

// drawTerrain draws the composed terrain of the hexes in view (in line, then
// column order): every background first, then the cluster stamps, the
// infrastructure, cast shadows and standing items. Every shadow is drawn before
// every crown, so a later hex's shadow cannot cover an earlier crown.
func (g *game) drawTerrain(dst *ebiten.Image, inView []hexgrid.Coord) {
	s := &g.terrain.scene
	cells := withClusterAnchors(s, inView)
	for pass := range 5 {
		for _, c := range cells {
			bg, stamp, overlays, standing, dim := terrainLayers(s, c, g.fogOf)
			var shadows []string
			if pass == 3 {
				shadows, _ = terrainShadows(s, c, g.fogOf)
			}
			sx, sy := g.cam.WorldToScreen(hexgrid.Center(c))
			for _, id := range [5][]string{{bg}, {stamp}, overlays, shadows, standing}[pass] {
				if id != "" {
					g.drawTile(dst, id, sx, sy, dim)
				}
			}
		}
	}
}

// withClusterAnchors adds to the hexes in view the anchor hex of every stamped
// cluster reaching into the view, since a stamp may span more than the view's
// margin, and keeps line, then column order.
func withClusterAnchors(s *mapcompose.Scene, inView []hexgrid.Coord) []hexgrid.Coord {
	in := map[hexgrid.Coord]bool{}
	for _, c := range inView {
		in[c] = true
	}
	out := inView
	for _, cl := range s.Clusters {
		if cl.Stamp != "" && !in[cl.Cells[0]] && slices.ContainsFunc(cl.Cells, func(c hexgrid.Coord) bool { return in[c] }) {
			out = append(slices.Clip(out), cl.Cells[0])
			in[cl.Cells[0]] = true
		}
	}
	if len(out) > len(inView) {
		slices.SortFunc(out, func(a, b hexgrid.Coord) int { return cmp.Or(cmp.Compare(a.Line, b.Line), cmp.Compare(a.Col, b.Col)) })
	}
	return out
}
