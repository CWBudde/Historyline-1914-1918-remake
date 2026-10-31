package main

import (
	"cmp"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapmaterial"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/waterdepth"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/worldmaterial"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

// tilePad is the border around terrain textures, in texels. Background tiles
// own complementary pixels, but scaled, filtered quads do not meet exactly
// (Ebitengine snaps quad corners to a sub-pixel grid), which leaves seams.
// Each background's opaque edge grown into the border fills them from beneath
// (drawTerrain); overlays keep a transparent border.
const tilePad = 3

// terrainGroups are the pack groups hl assets build-terrain writes; the optional
// groups only when the catalog has standing items or cluster stamps.
var (
	terrainGroups  = []string{"backgrounds", "infrastructure", "context-ground", "context-forest", "context-hills"}
	optionalGroups = []string{standingGroup, standingStampGroup, stampGroup, worldmaterial.Group}
)

const standingGroup, standingStampGroup, stampGroup = "standing", "standing-stamps", "stamps"

// terrainArt is the runtime terrain catalog and the level composed from it.
type terrainArt struct {
	*artPacks
	materials *terrainMaterials
	depth     *depthTiles // the water backgrounds shaded by depth
	scene     mapcompose.Scene
	group     map[string]string // asset ID → pack group
	// dim holds the remembered look of each texture drawn dimmed so far.
	dim map[*ebiten.Image]*ebiten.Image
}

// openTerrainArt opens the terrain catalog and composes lvl with it. Context
// tiles are used where the catalog has them at every density; every other tile
// the level draws, and the unseen tile, must be present, so drawing never meets
// a missing tile.
func openTerrainArt(fsys fs.FS, opt assets.Options, lvl *level.Level) (*terrainArt, error) {
	groups, err := assets.PackGroups(fsys, terrainGroups, optionalGroups, terrainDensities)
	if err != nil {
		return nil, fmt.Errorf("terrain %w", err)
	}
	packs, err := openPacks(fsys, opt, groups, terrainDensities)
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
		for _, d := range terrainDensities {
			manifests[g] = append(manifests[g], packs.packs[packKey{g, d}].Manifest())
		}
		for _, e := range packs.packs[packKey{g, terrainDensities[0]}].Manifest().Entries {
			a.group[e.ID] = g
		}
	}
	if err := assets.CheckGroupDensities(manifests, terrainDensities); err != nil {
		return nil, err
	}
	if err := a.present(mapcompose.TileID("unseen")); err != nil {
		return nil, err
	}
	if err := a.compose(lvl); err != nil {
		return nil, err
	}
	if err := a.openMaterials(lvl); err != nil {
		return nil, err
	}
	return a, nil
}

// present checks that the catalog has tile id at every density.
func (a *terrainArt) present(id string) error {
	g, ok := a.group[id]
	if !ok {
		return fmt.Errorf("terrain catalog has no tile %s", id)
	}
	for _, d := range terrainDensities {
		if _, err := a.entry(g, id, d); err != nil {
			return fmt.Errorf("terrain pack %dx/%s: %w", d, g, err)
		}
	}
	return nil
}

// compose composes lvl in its season as the scene; on a missing tile the
// scene stays.
func (a *terrainArt) compose(lvl *level.Level) error {
	has := func(id string) bool { return a.present(id) == nil }
	s := mapcompose.ComposeSeason(lvl.Grid(), mapcompose.LevelHex(lvl), has, mapcompose.LevelSeason(lvl))
	var ids []string
	for _, line := range s.Cells {
		for _, c := range line {
			ids = slices.Concat(ids, []string{c.Background, c.Under}, c.Overlays, c.Shadows, c.Standing)
		}
	}
	for _, cl := range s.Clusters {
		ids = append(ids, cl.Stamp, cl.Standing)
	}
	for _, id := range ids {
		if id == "" {
			continue
		}
		if err := a.present(id); err != nil {
			return err
		}
	}
	a.scene = s
	if a.depth != nil {
		a.depth.release(a)
	}
	a.depth = &depthTiles{field: waterdepth.New(lvl, &a.scene), tiles: map[depthKey]*ebiten.Image{}}
	return nil
}

// recompose draws lvl, a level whose hexes changed during the battle: the
// scene and the ground materials follow it.
func (a *terrainArt) recompose(lvl *level.Level) error {
	if err := a.compose(lvl); err != nil {
		return err
	}
	if m := a.materials; m != nil {
		m.field, m.level = mapmaterial.New(lvl), lvl
	}
	return nil
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

// dimmed is texture img in the remembered look, made on first use. Terrain is
// drawn with DrawImage, not colorm, since only DrawImage shrinks with mipmaps.
func (a *terrainArt) dimmed(img *ebiten.Image) *ebiten.Image {
	if d, ok := a.dim[img]; ok {
		return d
	}
	if a.dim == nil {
		a.dim = map[*ebiten.Image]*ebiten.Image{}
	}
	d := ebiten.NewImage(img.Bounds().Dx(), img.Bounds().Dy())
	colorm.DrawImage(d, img, remembered, nil)
	a.dim[img] = d
	return d
}

// forget releases the remembered look of texture img.
func (a *terrainArt) forget(img *ebiten.Image) {
	if d, ok := a.dim[img]; ok {
		d.Deallocate()
		delete(a.dim, img)
	}
}

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

// scatterLayers are the scattered items drawn after the tiles of a pass: ground
// props after the cluster stamps, then the crowns' cast shadows, then the crowns.
var scatterLayers = [5]func(*mapcompose.Scene, hexgrid.Coord, func(*mapcompose.Cluster) bool) []mapcompose.Placed{
	1: (*mapcompose.Scene).GroundAt,
	3: (*mapcompose.Scene).ScatterShadowsAt,
	4: (*mapcompose.Scene).ScatterAt,
}

// terrainScatter returns the items of one scatter layer of hex c under the fog
// rules of standing items.
func terrainScatter(s *mapcompose.Scene, c hexgrid.Coord, fog func(hexgrid.Coord) state.Fog, layer func(*mapcompose.Scene, hexgrid.Coord, func(*mapcompose.Cluster) bool) []mapcompose.Placed) ([]mapcompose.Placed, bool) {
	f := fog(c)
	if f == state.Unseen {
		return nil, false
	}
	known := func(cl *mapcompose.Cluster) bool {
		return !slices.ContainsFunc(cl.Cells, func(x hexgrid.Coord) bool { return fog(x) == state.Unseen })
	}
	return layer(s, c, known), f == state.Seen
}

// drawTerrainTile draws id with its anchor on screen point (sx, sy); a bare
// background without its bleeding border.
func (g *game) drawTerrainTile(dst *ebiten.Image, id string, sx, sy float64, dim bool, at *hexgrid.Coord, bare bool) {
	d := terrainDensity(g.cam.Effective())
	var img *ebiten.Image
	var e assets.Entry
	var err error
	signature := ""
	if at != nil {
		signature = g.terrain.depthSignature(*at, id)
	}
	if signature != "" {
		img, e, err = g.terrain.depthImage(*at, signature, id, d, bare)
	} else if at != nil && g.terrain.materials.eligible(*at, id) {
		img, e, err = g.terrain.materialImage(*at, id, d, bare)
	} else if bare {
		img, e, err = g.terrain.bareImage(g.terrain.group[id], id, d)
	} else {
		img, e, err = g.terrain.image(g.terrain.group[id], id, d)
	}
	if err != nil {
		g.terrainErr = err.Error()
		return
	}
	if dim {
		img = g.terrain.dimmed(img)
	}
	var op ebiten.DrawImageOptions
	op.GeoM = anchorGeoM(e, d, g.terrain.pad, g.cam.Effective(), sx, sy)
	op.Filter = ebiten.FilterLinear
	dst.DrawImage(img, &op)
}

// drawTerrain draws the composed terrain of the hexes in view (in line, then
// column order): every background first, then the cluster stamps, the
// infrastructure, cast shadows and standing items, each pass with its scattered
// items (ground props after the stamps). Every shadow is drawn before every crown,
// so a later hex's shadow cannot cover an earlier crown. The backgrounds' bleeding
// borders go beneath all backgrounds, which are drawn bare over them: the borders
// fill the gaps between scaled tiles, but none covers a neighbour's own texels.
func (g *game) drawTerrain(dst *ebiten.Image, inView []hexgrid.Coord) {
	s := &g.terrain.scene
	cells := withClusterAnchors(s, inView)
	for _, c := range cells {
		bg, _, _, _, dim := terrainLayers(s, c, g.fogOf)
		sx, sy := g.cam.WorldToScreen(hexgrid.Center(c))
		g.drawTerrainTile(dst, bg, sx, sy, dim, g.materialHex(c), false)
	}
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
					var at *hexgrid.Coord
					if pass == 0 {
						at = g.materialHex(c)
					}
					g.drawTerrainTile(dst, id, sx, sy, dim, at, true)
				}
			}
			if layer := scatterLayers[pass]; layer != nil {
				scattered, dim := terrainScatter(s, c, g.fogOf, layer)
				cx, cy := hexgrid.Center(c)
				for _, p := range scattered {
					px, py := g.cam.WorldToScreen(cx+p.X, cy+p.Y)
					g.drawTerrainTile(dst, p.ID, px, py, dim, nil, true)
				}
			}
		}
	}
}

// materialHex is c for a background that may sample the world material: any
// but an unseen hex's.
func (g *game) materialHex(c hexgrid.Coord) *hexgrid.Coord {
	if g.fogOf(c) == state.Unseen {
		return nil
	}
	return &c
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
