package main

import (
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapmaterial"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/worldmaterial"
	"github.com/hajimehoshi/ebiten/v2"
)

type materialTileKey struct {
	phase           image.Point
	signature, mask string
	density         int
	bare            bool // with a transparent border, not a bleeding one
}

type terrainMaterials struct {
	samplers map[int]*worldmaterial.Sampler
	field    *mapmaterial.Field
	level    *level.Level
	tiles    map[materialTileKey]*ebiten.Image
	order    []materialTileKey
	bytes    int
}

// release also frees the derived GPU textures, which live outside artPacks.tex.
func (a *terrainArt) release() {
	if a == nil {
		return
	}
	if m := a.materials; m != nil {
		for _, img := range m.tiles {
			img.Deallocate()
		}
		m.tiles = make(map[materialTileKey]*ebiten.Image)
		m.order = nil
		m.bytes = 0
	}
	if a.depth != nil {
		a.depth.release(a)
	}
	for _, d := range a.dim {
		d.Deallocate()
	}
	a.dim = nil
	a.artPacks.release()
}

func validateMaterialManifest(m assets.Manifest, d int) error {
	if len(m.Entries) != 2 {
		return fmt.Errorf("materials must contain exactly meadow and scrub")
	}
	seen := map[string]bool{}
	for _, e := range m.Entries {
		if (e.ID != worldmaterial.MeadowID && e.ID != worldmaterial.ScrubID) || seen[e.ID] {
			return fmt.Errorf("unknown or duplicate material %q", e.ID)
		}
		seen[e.ID] = true
		if e.Density != d || e.Width != worldmaterial.PeriodWidth*d || e.Height != worldmaterial.PeriodHeight*d || e.Anchor != [2]float64{} {
			return fmt.Errorf("invalid material geometry %s at %dx", e.ID, d)
		}
	}
	return nil
}

func (a *terrainArt) openMaterials(lvl *level.Level) error {
	var found bool
	for k := range a.packs {
		if strings.HasPrefix(k.group, worldmaterial.Group+"-") {
			return fmt.Errorf("materials cannot be split into parts")
		}
		found = found || k.group == worldmaterial.Group
	}
	if !found {
		return nil
	}
	m := &terrainMaterials{samplers: map[int]*worldmaterial.Sampler{}, field: mapmaterial.New(lvl), level: lvl, tiles: map[materialTileKey]*ebiten.Image{}}
	for _, d := range terrainDensities {
		p := a.packs[packKey{worldmaterial.Group, d}]
		if p == nil {
			return fmt.Errorf("missing materials at %dx", d)
		}
		if err := validateMaterialManifest(p.Manifest(), d); err != nil {
			return err
		}
		meadow, err := p.Image(worldmaterial.MeadowID, d)
		if err != nil {
			return err
		}
		scrub, err := p.Image(worldmaterial.ScrubID, d)
		if err != nil {
			return err
		}
		s, err := worldmaterial.New(meadow, scrub, image.Pt(worldmaterial.PeriodWidth, worldmaterial.PeriodHeight), d)
		if err != nil {
			return fmt.Errorf("materials %dx: %w", d, err)
		}
		m.samplers[d] = s
	}
	a.materials = m
	return nil
}

func materialOrigin(c hexgrid.Coord, e assets.Entry, d int) image.Point {
	x, y := hexgrid.Center(c)
	return image.Pt(int(math.Round((x-e.Anchor[0])*float64(d))), int(math.Round((y-e.Anchor[1])*float64(d))))
}

// Only the actual summer ground background participates; overlays and unseen
// backgrounds retain their own art even when their owning cell is meadow.
func (m *terrainMaterials) eligible(c hexgrid.Coord, id string) bool {
	if m == nil || m.level == nil || c.Line < 0 || c.Line >= len(m.level.Lines) || c.Col < 0 || c.Col >= len(m.level.Lines[c.Line]) {
		return false
	}
	terrain := m.level.Lines[c.Line][c.Col].Terrain
	return id != "" && id != "tile/unseen" && (terrain == "grass" || terrain == "hills")
}

func (a *terrainArt) materialImage(c hexgrid.Coord, id string, d int, bare bool) (*ebiten.Image, assets.Entry, error) {
	e, err := a.entry(a.group[id], id, d)
	if err != nil {
		return nil, e, err
	}
	m := a.materials
	origin := materialOrigin(c, e, d)
	s := m.samplers[d]
	key := materialTileKey{s.Phase(origin), m.field.Signature(c), id, d, bare}
	if img := m.tiles[key]; img != nil {
		return img, e, nil
	}
	mask, err := a.packs[packKey{a.group[id], d}].Image(id, d)
	if err != nil {
		return nil, e, err
	}
	tile, err := sampleMaterial(s, m.field, mask, origin)
	if err != nil {
		return nil, e, err
	}
	img := ebiten.NewImage(tile.Bounds().Dx()+2*tilePad, tile.Bounds().Dy()+2*tilePad)
	img.WritePixels(padded(tile, tilePad, !bare))
	size := (tile.Bounds().Dx() + 2*tilePad) * (tile.Bounds().Dy() + 2*tilePad) * 4
	for m.bytes+size > 128*1024*1024 && len(m.order) > 0 {
		old := m.order[0]
		m.order = m.order[1:]
		previous := m.tiles[old]
		m.bytes -= previous.Bounds().Dx() * previous.Bounds().Dy() * 4
		a.forget(previous)
		previous.Deallocate()
		delete(m.tiles, old)
	}
	m.tiles[key] = img
	m.order = append(m.order, key)
	m.bytes += size
	return img, e, nil
}

func sampleMaterial(s *worldmaterial.Sampler, field *mapmaterial.Field, mask image.Image, origin image.Point) (*image.NRGBA, error) {
	return s.Tile(mask, origin, field.Weight)
}
