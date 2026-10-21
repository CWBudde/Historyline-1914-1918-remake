// Package maprender draws a whole level with the runtime art of assets/index.json,
// layered as in the game: backgrounds (context tiles chosen by mapcompose), cluster
// stamps, infrastructure, standing items, then units facing north in their
// owner's colours. Fog, owner bars and UI are left out.
package maprender

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/fs"
	"math"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

// Entry is one runtime file of assets/index.json; Path is relative to assets/.
type Entry struct {
	ID      string     `json:"id"`
	Pack    string     `json:"pack"`
	Density int        `json:"density"`
	Path    string     `json:"path"`
	SHA     string     `json:"sha256"`
	Anchor  [2]float64 `json:"logical_anchor"`
}

// Map is a rendered level.
type Map struct {
	Image  *image.NRGBA
	Units  int
	Gaps   []mapcompose.Gap
	Issues []mapcompose.Issue
}

// Renderer draws levels from one assets tree.
type Renderer struct {
	assets fs.FS
	files  map[string]Entry // terrain and unit IDs per density, keyed "density/id"
	has    map[string]bool  // tile, context, standing and stamp IDs present at every density
	cache  map[string]image.Image
	set    *data.Set
}

// New reads index.json of the assets tree.
func New(assets fs.FS) (*Renderer, error) {
	raw, err := fs.ReadFile(assets, "index.json")
	if err != nil {
		return nil, err
	}
	var index struct {
		Schema int     `json:"schema_version"`
		Assets []Entry `json:"assets"`
	}
	if err := json.Unmarshal(raw, &index); err != nil {
		return nil, fmt.Errorf("assets/index.json: %w", err)
	}
	if index.Schema != 1 || len(index.Assets) == 0 {
		return nil, fmt.Errorf("assets/index.json: unsupported or empty index")
	}
	set, err := data.Load()
	if err != nil {
		return nil, err
	}
	r := &Renderer{assets: assets, files: map[string]Entry{}, has: map[string]bool{}, cache: map[string]image.Image{}, set: set}
	densities := map[string]int{}
	for _, e := range index.Assets {
		if !strings.HasPrefix(e.Pack, "terrain-") && !strings.HasPrefix(e.Pack, "unit-") {
			continue
		}
		r.files[fmt.Sprintf("%d/%s", e.Density, e.ID)] = e
		if strings.HasPrefix(e.ID, "context/") || strings.HasPrefix(e.ID, "tile/") || strings.HasPrefix(e.ID, "standing/") || strings.HasPrefix(e.ID, "stamp/") {
			densities[e.ID]++
		}
	}
	for id, n := range densities {
		if n != 2 {
			return nil, fmt.Errorf("%s lacks a 2x or 4x tile", id)
		}
		r.has[id] = true
	}
	return r, nil
}

// Size returns the logical (1x) pixel size of a grid's map.
func Size(g hexgrid.Grid) (int, int) {
	return (g.Width-1)*142 + 96, (g.Lines-1)*40 + 80
}

// Render draws lvl at density 2 or 4, with or without its units.
func (r *Renderer) Render(lvl *level.Level, density int, units bool) (Map, error) {
	if density != 2 && density != 4 {
		return Map{}, fmt.Errorf("density %d: runtime art exists at 2x and 4x", density)
	}
	grid := lvl.Grid()
	scene := mapcompose.Compose(grid, func(at hexgrid.Coord) mapcompose.Hex {
		h := lvl.Lines[at.Line][at.Col]
		return mapcompose.Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
	}, func(id string) bool { return r.has[id] })
	w, h := Size(grid)
	m := Map{Image: image.NewNRGBA(image.Rect(0, 0, w*density, h*density)), Gaps: scene.Gaps, Issues: scene.Issues}
	// Backgrounds, cluster stamps, infrastructure and standing items are separate
	// full-map passes, as in the game.
	all := func(*mapcompose.Cluster) bool { return true }
	for pass := range 4 {
		for line, cells := range scene.Cells {
			for col := range cells {
				at := hexgrid.Coord{Line: line, Col: col}
				bg, stamp, overlays, standing := scene.Layers(at, all)
				ids := [4][]string{{bg}, {stamp}, overlays, standing}[pass]
				for _, id := range ids {
					if id == "" {
						continue
					}
					if err := r.draw(m.Image, id, at, density); err != nil {
						return Map{}, err
					}
				}
			}
		}
	}
	if !units {
		return m, nil
	}
	st, err := game.New(lvl, r.set)
	if err != nil {
		return Map{}, err
	}
	for line, hs := range lvl.Lines {
		for col, hx := range hs {
			if hx.Unit == "" {
				continue
			}
			at := hexgrid.Coord{Line: line, Col: col}
			family, ok := unitsprite.Family(hx.Unit)
			if !ok {
				return Map{}, fmt.Errorf("missing unit family %s", hx.Unit)
			}
			id := fmt.Sprintf("unit/%s/%s/n-normal", family, unitsprite.PaletteOf(st.At(at).Unit.Owner))
			if err := r.draw(m.Image, id, at, density); err != nil {
				return Map{}, err
			}
			m.Units++
		}
	}
	return m, nil
}

func (r *Renderer) draw(dst draw.Image, id string, at hexgrid.Coord, density int) error {
	e, ok := r.files[fmt.Sprintf("%d/%s", density, id)]
	if !ok {
		return fmt.Errorf("missing %dx asset %s", density, id)
	}
	im := r.cache[e.Path]
	if im == nil {
		b, err := fs.ReadFile(r.assets, e.Path)
		if err != nil {
			return err
		}
		if sum := fmt.Sprintf("%x", sha256.Sum256(b)); sum != e.SHA {
			return fmt.Errorf("%s: sha256 %s, index pins %s", e.Path, sum, e.SHA)
		}
		if im, err = png.Decode(bytes.NewReader(b)); err != nil {
			return fmt.Errorf("%s: %w", e.Path, err)
		}
		r.cache[e.Path] = im
	}
	Place(dst, im, at, density, e.Anchor)
	return nil
}

// Place draws src with its logical anchor on the centre of hex at.
func Place(dst draw.Image, src image.Image, at hexgrid.Coord, density int, anchor [2]float64) {
	x, y := hexgrid.Center(at)
	p := image.Pt(int(math.Round((x-anchor[0])*float64(density))), int(math.Round((y-anchor[1])*float64(density))))
	draw.Draw(dst, src.Bounds().Sub(src.Bounds().Min).Add(p), src, src.Bounds().Min, draw.Over)
}
