package maprender

import (
	"fmt"
	"image"
	"math"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapmaterial"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/seasonblend"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/waterdepth"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/worldmaterial"
)

func (r *Renderer) addMaterial(e Entry) error {
	if (e.ID != worldmaterial.MeadowID && e.ID != worldmaterial.ScrubID) || (e.Density != 2 && e.Density != 4) || e.Pack != fmt.Sprintf("terrain-materials-%dx", e.Density) || e.Width != worldmaterial.PeriodWidth*e.Density || e.Height != worldmaterial.PeriodHeight*e.Density || e.Anchor != [2]float64{} || (e.Layer != "" && e.Layer != "ground") {
		return fmt.Errorf("invalid material entry %s in %s", e.ID, e.Pack)
	}
	key := fmt.Sprintf("%d/%s", e.Density, e.ID)
	if _, exists := r.files[key]; exists {
		return fmt.Errorf("duplicate material %s", key)
	}
	r.files[key] = e
	return nil
}

func (r *Renderer) loadMaterials() error {
	count := 0
	for _, d := range []int{2, 4} {
		for _, id := range []string{worldmaterial.MeadowID, worldmaterial.ScrubID} {
			if _, ok := r.files[fmt.Sprintf("%d/%s", d, id)]; ok {
				count++
			}
		}
	}
	if count == 0 {
		return nil
	}
	if count != 4 {
		return fmt.Errorf("materials require both IDs at 2x and 4x")
	}
	r.materials = make(map[int]*worldmaterial.Sampler)
	for _, d := range []int{2, 4} {
		_, grass, err := r.assetImage(worldmaterial.MeadowID, d)
		if err != nil {
			return err
		}
		_, scrub, err := r.assetImage(worldmaterial.ScrubID, d)
		if err != nil {
			return err
		}
		sampler, err := worldmaterial.New(grass, scrub, image.Pt(worldmaterial.PeriodWidth, worldmaterial.PeriodHeight), d)
		if err != nil {
			return fmt.Errorf("%dx materials: %w", d, err)
		}
		r.materials[d] = sampler
	}
	return nil
}

// origin is the top-left corner, in map pixels at density, of an image of
// entry e placed on hex at.
func origin(e Entry, at hexgrid.Coord, density int) image.Point {
	x, y := hexgrid.Center(at)
	return image.Pt(int(math.Round((x-e.Anchor[0])*float64(density))), int(math.Round((y-e.Anchor[1])*float64(density))))
}

// ground is background id of hex at as drawn, placed at its top-left corner:
// water shaded by depth, a summer meadow or hills background sampled from the
// world material, any other as painted.
func (r *Renderer) ground(id string, at hexgrid.Coord, density int, depth *waterdepth.Field, field *mapmaterial.Field, terrain string) (seasonblend.Part, error) {
	if depth != nil && depth.Signature(at) != "" {
		return r.deep(id, at, density, depth)
	}
	e, im, err := r.assetImage(id, density)
	if err != nil {
		return seasonblend.Part{}, err
	}
	o := origin(e, at, density)
	if r.materials[density] != nil && materialGround(terrain, id) {
		tile, err := r.materials[density].Tile(im, o, field.Weight)
		if err != nil {
			return seasonblend.Part{}, err
		}
		return seasonblend.Part{Image: tile, At: o}, nil
	}
	return seasonblend.Part{Image: im, At: o}, nil
}

// deep is background id on hex at with its water shaded by the sea's depth
// (waterdepth), the open sea's material blended in and surf along a coast by
// the sea where the catalog has them.
func (r *Renderer) deep(id string, at hexgrid.Coord, density int, field *waterdepth.Field) (seasonblend.Part, error) {
	e, im, err := r.assetImage(id, density)
	if err != nil {
		return seasonblend.Part{}, err
	}
	o := origin(e, at, density)
	var with waterdepth.Tiles
	for _, t := range []struct {
		id  string
		set *image.Image
	}{{waterdepth.SeaTile(at, id), &with.Sea}, {waterdepth.WaterTile(id), &with.Water}, {waterdepth.SurfTile(id), &with.Surf}} {
		if !r.has[t.id] {
			t.id = mapcompose.SummerID(t.id) // an autumn or spring form the catalog lacks
		}
		if _, im, err := r.assetImage(t.id, density); err == nil {
			*t.set = im
		}
	}
	return seasonblend.Part{Image: field.Shade(im, with, o, density), At: o}, nil
}
