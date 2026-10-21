package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"slices"

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

// terrainGroups are the pack groups hlasset build-terrain writes;
// standingGroup is written only when the catalog has standing items.
var terrainGroups = []string{"backgrounds", "infrastructure", "context-ground", "context-forest", "context-hills"}

const standingGroup = "standing"

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
	groups := terrainGroups
	if _, err := fs.Stat(fsys, path.Join(fmt.Sprintf("%dx", spriteDensities[0]), standingGroup, "manifest.json")); err == nil {
		groups = append(slices.Clip(groups), standingGroup)
	}
	packs, err := openPacks(fsys, opt, groups, spriteDensities)
	if err != nil {
		return nil, fmt.Errorf("terrain %w", err)
	}
	packs.pad = tilePad
	packs.bleed = map[string]bool{}
	for _, g := range groups {
		packs.bleed[g] = g != "infrastructure" && g != standingGroup
	}
	a := &terrainArt{artPacks: packs, group: map[string]string{}}
	for _, g := range groups {
		for _, e := range packs.packs[packKey{g, spriteDensities[0]}].Manifest().Entries {
			a.group[e.ID] = g
		}
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
	for _, line := range a.scene.Cells {
		for _, c := range line {
			for _, id := range slices.Concat([]string{c.Background}, c.Overlays, c.Standing) {
				if err := present(id); err != nil {
					return nil, err
				}
			}
		}
	}
	return a, nil
}

// loadTerrainArt opens the terrain catalog in dir for lvl.
func loadTerrainArt(dir string, allowCandidates bool, lvl *level.Level) (*terrainArt, error) {
	r, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	return openTerrainArt(r.FS(), assets.Options{AllowCandidates: allowCandidates}, lvl)
}

// terrainLayers returns the tiles drawn on a hex the player knows as f: nothing
// of an unseen hex but the fog tile (its standing items would give its terrain
// away), and the remembered terrain of a seen hex dimmed to grey. Standing items
// of a known hex may overhang an unseen neighbour.
func terrainLayers(c mapcompose.Cell, f state.Fog) (bg string, overlays, standing []string, dim bool) {
	if f == state.Unseen {
		return mapcompose.TileID("unseen"), nil, nil, false
	}
	return c.Background, c.Overlays, c.Standing, f == state.Seen
}

// remembered is the grey, darkened look of a seen but not visible hex.
var remembered = func() colorm.ColorM {
	var m colorm.ColorM
	m.ChangeHSV(0, 0, 0.5)
	return m
}()

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
// column order): every background first, then the infrastructure, then the
// standing items, so neither is covered by a later hex's ground.
func (g *game) drawTerrain(dst *ebiten.Image, inView []hexgrid.Coord) {
	for pass := range 3 {
		for _, c := range inView {
			bg, overlays, standing, dim := terrainLayers(g.terrain.scene.Cells[c.Line][c.Col], g.fogOf(c))
			sx, sy := g.cam.WorldToScreen(hexgrid.Center(c))
			ids := [3][]string{{bg}, overlays, standing}[pass]
			for _, id := range ids {
				g.drawTile(dst, id, sx, sy, dim)
			}
		}
	}
}
