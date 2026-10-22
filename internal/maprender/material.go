package maprender

import (
	"fmt"
	"image"
	"image/draw"
	"math"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapmaterial"
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

func (r *Renderer) drawMaterial(dst draw.Image, id string, at hexgrid.Coord, density int, field *mapmaterial.Field) error {
	e, mask, err := r.assetImage(id, density)
	if err != nil {
		return err
	}
	x, y := hexgrid.Center(at)
	origin := image.Pt(int(math.Round((x-e.Anchor[0])*float64(density))), int(math.Round((y-e.Anchor[1])*float64(density))))
	tile, err := r.materials[density].Tile(mask, origin, field.Weight)
	if err != nil {
		return err
	}
	draw.Draw(dst, tile.Bounds().Add(origin), tile, image.Point{}, draw.Over)
	return nil
}
