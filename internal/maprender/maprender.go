// Package maprender draws a whole level with the runtime art of assets/index.json,
// layered as in the game: backgrounds (context tiles chosen by mapcompose), cluster
// stamps, infrastructure, cast shadows, standing items, then units facing north in their
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
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapmaterial"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/terraingrade"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/terrainlight"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/waterdepth"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/worldmaterial"
)

// Entry is one runtime file of assets/index.json; Path is relative to assets/.
type Entry struct {
	ID      string     `json:"id"`
	Pack    string     `json:"pack"`
	Density int        `json:"density"`
	Path    string     `json:"path"`
	SHA     string     `json:"sha256"`
	Anchor  [2]float64 `json:"logical_anchor"`
	Width   int        `json:"width"`
	Height  int        `json:"height"`
	Layer   string     `json:"layer"`
	Art     string     `json:"art"`
}

// Map is a rendered level.
type Map struct {
	Image  *image.NRGBA
	Units  int
	Gaps   []mapcompose.Gap
	Issues []mapcompose.Issue
}

// Renderer draws levels from one assets tree. Ungraded draws the terrain as
// painted, without the pack build's grade, for comparisons; set it before the
// first Render.
type Renderer struct {
	Ungraded bool

	assets    fs.FS
	files     map[string]Entry       // terrain and unit IDs per density, keyed "density/id"
	has       map[string]bool        // tile, context, standing and stamp IDs present at every density
	cache     map[string]image.Image // decoded files as painted, by path
	graded    map[string]image.Image // images as drawn, keyed "density/id"
	set       *data.Set
	materials map[int]*worldmaterial.Sampler
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
	r := &Renderer{assets: assets, files: map[string]Entry{}, has: map[string]bool{}, cache: map[string]image.Image{}, graded: map[string]image.Image{}, set: set}
	densities := map[string]int{}
	pathHashes := map[string]string{}
	for _, e := range index.Assets {
		material := strings.HasPrefix(e.ID, "material/") || strings.HasPrefix(e.Pack, "terrain-materials")
		if !material && !strings.HasPrefix(e.Pack, "terrain-") && !strings.HasPrefix(e.Pack, "unit-") {
			continue
		}
		if previous, ok := pathHashes[e.Path]; ok && previous != e.SHA {
			return nil, fmt.Errorf("%s: conflicting sha256 pins", e.Path)
		}
		pathHashes[e.Path] = e.SHA
		if material {
			if err := r.addMaterial(e); err != nil {
				return nil, err
			}
			continue
		}
		r.files[fmt.Sprintf("%d/%s", e.Density, e.ID)] = e
		if strings.HasPrefix(e.ID, "context/") || strings.HasPrefix(e.ID, "tile/") || strings.HasPrefix(e.ID, "standing/") || strings.HasPrefix(e.ID, "stamp/") {
			densities[e.ID]++
		}
	}
	if err := r.loadMaterials(); err != nil {
		return nil, err
	}
	for id, n := range densities {
		if n != 2 {
			return nil, fmt.Errorf("%s lacks a 2x or 4x tile", id)
		}
		r.has[id] = true
	}
	return r, nil
}

// materialGround reports whether background id of a hex of terrain samples the
// world material: a summer meadow or hills background, not its autumn or
// spring form.
func materialGround(terrain, id string) bool {
	return mapcompose.IDSeason(id) == mapcompose.Summer && (terrain == "grass" || terrain == "hills")
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
	field := mapmaterial.New(lvl)
	scene := mapcompose.ComposeSeason(grid, mapcompose.LevelHex(lvl), func(id string) bool { return r.has[id] }, mapcompose.LevelSeason(lvl))
	depth := waterdepth.New(lvl, &scene)
	w, h := Size(grid)
	m := Map{Image: image.NewNRGBA(image.Rect(0, 0, w*density, h*density)), Gaps: scene.Gaps, Issues: scene.Issues}
	// Backgrounds, cluster stamps with the ground props, infrastructure, shadows and
	// standing are separate full-map passes, as in the game.
	all := func(*mapcompose.Cluster) bool { return true }
	for pass := range 5 {
		for line, cells := range scene.Cells {
			for col := range cells {
				at := hexgrid.Coord{Line: line, Col: col}
				bg, stamp, overlays, standing := scene.Layers(at, all)
				var shadows []string
				if pass == 3 {
					shadows = scene.ShadowsAt(at, all)
				}
				ids := [5][]string{{bg}, {stamp}, overlays, shadows, standing}[pass]
				var scattered []mapcompose.Placed
				switch pass {
				case 1:
					scattered = scene.GroundAt(at, all)
				case 3:
					scattered = scene.ScatterShadowsAt(at, all)
				case 4:
					scattered = scene.ScatterAt(at, all)
				}
				for _, id := range ids {
					if id == "" {
						continue
					}
					if pass == 0 && depth.Signature(at) != "" {
						if err := r.drawDeep(m.Image, id, at, density, depth); err != nil {
							return Map{}, err
						}
						continue
					}
					terrain := lvl.Lines[line][col].Terrain
					if pass == 0 && r.materials[density] != nil && materialGround(terrain, id) {
						if err := r.drawMaterial(m.Image, id, at, density, field); err != nil {
							return Map{}, err
						}
						continue
					}
					if err := r.draw(m.Image, id, at, density); err != nil {
						return Map{}, err
					}
				}
				for _, p := range scattered {
					if err := r.drawAt(m.Image, p, at, density); err != nil {
						return Map{}, err
					}
				}
			}
		}
	}
	if !units {
		return m, nil
	}
	if err := r.drawUnits(&m, lvl, density); err != nil {
		return Map{}, err
	}
	return m, nil
}

func (r *Renderer) drawUnits(m *Map, lvl *level.Level, density int) error {
	st, err := game.New(lvl, r.set)
	if err != nil {
		return err
	}
	for line, hs := range lvl.Lines {
		for col, hx := range hs {
			if hx.Unit == "" {
				continue
			}
			at := hexgrid.Coord{Line: line, Col: col}
			family, ok := unitsprite.Family(hx.Unit)
			if !ok {
				return fmt.Errorf("missing unit family %s", hx.Unit)
			}
			id := fmt.Sprintf("unit/%s/%s/n-normal", family, unitsprite.PaletteOf(st.At(at).Unit.Owner))
			if err := r.draw(m.Image, id, at, density); err != nil {
				return err
			}
			m.Units++
		}
	}
	return nil
}

// assetImage returns the entry of id at density and its image, graded as the
// pack build grades it (terraingrade), a painted cast shadow drawn from its
// standing item with the one light (terrainlight).
func (r *Renderer) assetImage(id string, density int) (Entry, image.Image, error) {
	key := fmt.Sprintf("%d/%s", density, id)
	e, ok := r.files[key]
	if !ok {
		return Entry{}, nil, fmt.Errorf("missing %dx asset %s", density, id)
	}
	if im := r.graded[key]; im != nil {
		return e, im, nil
	}
	im, err := r.source(e)
	if err != nil {
		return Entry{}, nil, err
	}
	if standing, opacity, ok := terrainlight.Recast(e.ID, e.Art); ok {
		s, ok := r.files[fmt.Sprintf("%d/%s", density, standing)]
		if !ok {
			return Entry{}, nil, fmt.Errorf("the shadow %s: missing %dx asset %s", id, density, standing)
		}
		item, err := r.source(s)
		if err != nil {
			return Entry{}, nil, err
		}
		im = terrainlight.Cast(item, density, terrainlight.Diameter(item, density), opacity)
	} else if !r.Ungraded && terraingrade.Applies(e.ID, e.Art) {
		var meadow image.Image
		g, ok := r.files[fmt.Sprintf("%d/%s", density, terraingrade.Meadow(e.ID))]
		if !ok { // an autumn or spring meadow the catalog lacks
			g, ok = r.files[fmt.Sprintf("%d/%s", density, mapcompose.SummerID(terraingrade.Meadow(e.ID)))]
		}
		if ok {
			if meadow, err = r.source(g); err != nil {
				return Entry{}, nil, err
			}
		}
		im = terraingrade.Grade(im, e.ID, e.Art, meadow)
	}
	if r.graded == nil {
		r.graded = map[string]image.Image{}
	}
	r.graded[key] = im
	return e, im, nil
}

// source decodes the file of e as painted, checked against its pin.
func (r *Renderer) source(e Entry) (image.Image, error) {
	if im := r.cache[e.Path]; im != nil {
		return im, nil
	}
	b, err := fs.ReadFile(r.assets, e.Path)
	if err != nil {
		return nil, err
	}
	if sum := fmt.Sprintf("%x", sha256.Sum256(b)); sum != e.SHA {
		return nil, fmt.Errorf("%s: sha256 %s, index pins %s", e.Path, sum, e.SHA)
	}
	im, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", e.Path, err)
	}
	r.cache[e.Path] = im
	return im, nil
}

func (r *Renderer) draw(dst draw.Image, id string, at hexgrid.Coord, density int) error {
	e, im, err := r.assetImage(id, density)
	if err != nil {
		return err
	}
	Place(dst, im, at, density, e.Anchor)
	return nil
}

// drawAt draws a scattered item with its anchor at its offset from hex at.
func (r *Renderer) drawAt(dst draw.Image, p mapcompose.Placed, at hexgrid.Coord, density int) error {
	e, im, err := r.assetImage(p.ID, density)
	if err != nil {
		return err
	}
	x, y := hexgrid.Center(at)
	PlaceAt(dst, im, x+p.X, y+p.Y, density, e.Anchor)
	return nil
}

// Place draws src with its logical anchor on the centre of hex at.
func Place(dst draw.Image, src image.Image, at hexgrid.Coord, density int, anchor [2]float64) {
	x, y := hexgrid.Center(at)
	PlaceAt(dst, src, x, y, density, anchor)
}

// PlaceAt draws src with its logical anchor on map point (x, y).
func PlaceAt(dst draw.Image, src image.Image, x, y float64, density int, anchor [2]float64) {
	p := image.Pt(int(math.Round((x-anchor[0])*float64(density))), int(math.Round((y-anchor[1])*float64(density))))
	draw.Draw(dst, src.Bounds().Sub(src.Bounds().Min).Add(p), src, src.Bounds().Min, draw.Over)
}
