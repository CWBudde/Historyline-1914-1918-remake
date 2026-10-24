package maprender

import (
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
)

// CanopyMap records the visible contribution of separately authored tree crowns.
type CanopyMap struct {
	Coverage    *image.Alpha
	Unsupported []hexgrid.Coord
	Gaps        []mapcompose.Gap
	Issues      []mapcompose.Issue
}

// VisibleCanopy measures selected crown artwork at runtime density 2 or 4.
// Coverage is the source-over alpha contribution of known crown-only standing
// art, attenuated by later noncanopy standing art. Separate cast shadows, ground,
// units and UI contribute nothing. Unsupported lists forest cells whose selected
// standing art has no known crown-only layer, including baked winter artwork.
func (r *Renderer) VisibleCanopy(lvl *level.Level, density int) (CanopyMap, error) {
	if density != 2 && density != 4 {
		return CanopyMap{}, fmt.Errorf("density %d: runtime art exists at 2x and 4x", density)
	}
	grid := lvl.Grid()
	scene := mapcompose.Compose(grid, mapcompose.LevelHex(lvl), func(id string) bool { return r.has[id] })
	w, h := Size(grid)
	m := CanopyMap{
		Coverage: image.NewAlpha(image.Rect(0, 0, w*density, h*density)),
		Gaps:     scene.Gaps, Issues: scene.Issues,
	}
	all := func(*mapcompose.Cluster) bool { return true }
	for line, cells := range scene.Cells {
		for col := range cells {
			at := hexgrid.Coord{Line: line, Col: col}
			_, _, _, ids := scene.Layers(at, all)
			cx, cy := hexgrid.Center(at)
			var items []mapcompose.Placed
			for _, id := range ids {
				items = append(items, mapcompose.Placed{ID: id})
			}
			items = append(items, scene.ScatterAt(at, all)...)
			supported := false
			for _, it := range items {
				id := it.ID
				canopy := isCanopy(id)
				supported = supported || canopy
				e, src, err := r.assetImage(id, density)
				if err != nil {
					return CanopyMap{}, err
				}
				// Allocate only the placed sprite's clipped rectangle. Place remains
				// responsible for the same anchor and rounding used by Render.
				x, y := cx+it.X, cy+it.Y
				p := image.Pt(int(math.Round((x-e.Anchor[0])*float64(density))), int(math.Round((y-e.Anchor[1])*float64(density))))
				bounds := src.Bounds().Sub(src.Bounds().Min).Add(p).Intersect(m.Coverage.Bounds())
				alpha := image.NewAlpha(bounds)
				PlaceAt(alpha, src, x, y, density, e.Anchor)
				for py := bounds.Min.Y; py < bounds.Max.Y; py++ {
					for px := bounds.Min.X; px < bounds.Max.X; px++ {
						a := int(alpha.AlphaAt(px, py).A)
						i := m.Coverage.PixOffset(px, py)
						v := (int(m.Coverage.Pix[i])*(255-a) + 127) / 255
						if canopy {
							v += a
						}
						m.Coverage.Pix[i] = uint8(v)
					}
				}
			}
			if strings.TrimSuffix(lvl.Lines[line][col].Terrain, "_w") == "forest" && !supported {
				m.Unsupported = append(m.Unsupported, at)
			}
		}
	}
	return m, nil
}

func isCanopy(id string) bool {
	return id == "standing/forest" || strings.HasPrefix(id, "standing/context/forest/") || strings.HasPrefix(id, "standing/props/trees/") ||
		strings.HasPrefix(id, "standing/scatter/crown/")
}
