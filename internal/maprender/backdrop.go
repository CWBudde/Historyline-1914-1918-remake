package maprender

import (
	"fmt"
	"image"
	"math"

	xdraw "golang.org/x/image/draw"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapmaterial"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/waterdepth"
)

// RenderBackdrop draws lvl on a painted level backdrop instead of the
// terrain tiles, as the game does on a painted map: the backdrop covers
// the whole Size(grid) rectangle with its origin at the map rectangle's
// origin, any sapper-reworked hexes get their tile ground and overlays
// over the painting, and the units go on top. The backdrop may have any
// pixel size with the map rectangle's aspect (a level's backdrop.png
// master or its shipped backdrop.jpg); it is scaled to the density
// canvas with CatmullRom.
func (r *Renderer) RenderBackdrop(lvl *level.Level, density int, backdrop image.Image, reworked map[hexgrid.Coord]bool) (Map, error) {
	if density != 2 && density != 4 {
		return Map{}, fmt.Errorf("density %d: runtime art exists at 2x and 4x", density)
	}
	if backdrop == nil {
		return Map{}, fmt.Errorf("backdrop image is nil")
	}
	grid := lvl.Grid()
	w, h := Size(grid)
	bb := backdrop.Bounds()
	if bb.Dx() < 1 || bb.Dy() < 1 {
		return Map{}, fmt.Errorf("backdrop image is empty")
	}
	if got, want := float64(bb.Dx())/float64(bb.Dy()), float64(w)/float64(h); math.Abs(got-want)/want > 0.01 {
		return Map{}, fmt.Errorf("backdrop aspect %dx%d does not match the map's %dx%d", bb.Dx(), bb.Dy(), w, h)
	}
	field := mapmaterial.New(lvl)
	scene := mapcompose.Compose(grid, mapcompose.LevelHex(lvl), func(id string) bool { return r.has[id] })
	depth := waterdepth.New(lvl, &scene)
	m := Map{Image: image.NewNRGBA(image.Rect(0, 0, w*density, h*density)), Gaps: scene.Gaps, Issues: scene.Issues}
	xdraw.CatmullRom.Scale(m.Image, m.Image.Bounds(), backdrop, bb, xdraw.Src, nil)

	// Sapper-reworked hexes are the one piece of tile ground the game
	// still shows on a painted map.
	all := func(*mapcompose.Cluster) bool { return true }
	for line, cells := range scene.Cells {
		for col := range cells {
			at := hexgrid.Coord{Line: line, Col: col}
			if !reworked[at] {
				continue
			}
			bg, _, overlays, _ := scene.Layers(at, all)
			if bg != "" {
				terrain := lvl.Lines[line][col].Terrain
				switch {
				case depth.Signature(at) != "":
					if err := r.drawDeep(m.Image, bg, at, density, depth); err != nil {
						return Map{}, err
					}
				case r.materials[density] != nil && (terrain == "grass" || terrain == "hills"):
					if err := r.drawMaterial(m.Image, bg, at, density, field); err != nil {
						return Map{}, err
					}
				default:
					if err := r.draw(m.Image, bg, at, density); err != nil {
						return Map{}, err
					}
				}
			}
			for _, id := range overlays {
				if err := r.draw(m.Image, id, at, density); err != nil {
					return Map{}, err
				}
			}
		}
	}
	if err := r.drawUnits(&m, lvl, density); err != nil {
		return Map{}, err
	}
	return m, nil
}
