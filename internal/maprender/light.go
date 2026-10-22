package maprender

import (
	"fmt"
	"image"
	"math"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
)

// ShadowPair measures the complete source-alpha centroids of one selected crown
// and its selected cast shadow. Bearing is compass degrees clockwise from north.
// Reason excludes a pair from direction measurement; its Bearing is then nil.
type ShadowPair struct {
	At                               hexgrid.Coord
	Standing, Shadow                 Entry
	StandingCentroid, ShadowCentroid [2]float64
	Offset                           [2]float64
	Distance                         float64
	Bearing                          *float64
	Reason                           string
}

// ShadowOmission describes selected artwork or a forest cell without a
// separately measurable crown. Multiple reasons may describe the same cell.
type ShadowOmission struct {
	At         hexgrid.Coord
	ID, Reason string
}

// ShadowPairs records selected crown/shadow pairs and composition omissions.
type ShadowPairs struct {
	Density    int
	Pairs      []ShadowPair
	Unmeasured []ShadowOmission
	Gaps       []mapcompose.Gap
	Issues     []mapcompose.Issue
}

// CastShadowPairs measures source asset displacement at runtime density 2 or 4.
// It follows the actual scene selection, without map clipping or attenuation by
// other objects. The result describes asset direction, not global illumination.
func (r *Renderer) CastShadowPairs(lvl *level.Level, density int) (ShadowPairs, error) {
	if density != 2 && density != 4 {
		return ShadowPairs{}, fmt.Errorf("density %d: runtime art exists at 2x and 4x", density)
	}
	if lvl == nil {
		return ShadowPairs{}, fmt.Errorf("missing level")
	}
	scene := mapcompose.Compose(lvl.Grid(), func(at hexgrid.Coord) mapcompose.Hex {
		h := lvl.Lines[at.Line][at.Col]
		return mapcompose.Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
	}, func(id string) bool { return r.has[id] })
	result := ShadowPairs{Density: density, Gaps: scene.Gaps, Issues: scene.Issues}
	all := func(*mapcompose.Cluster) bool { return true }
	for line, cells := range scene.Cells {
		for col := range cells {
			at := hexgrid.Coord{Line: line, Col: col}
			background, _, _, standing := scene.Layers(at, all)
			shadows := scene.ShadowsAt(at, all)
			crownSelected := false
			for _, id := range standing {
				if !isCanopy(id) {
					result.Unmeasured = append(result.Unmeasured, ShadowOmission{At: at, ID: id, Reason: "non_canopy_standing"})
					continue
				}
				crownSelected = true
				pair, err := r.castShadowPair(at, id, slices.Contains(shadows, mapcompose.ShadowID(id)), density)
				if err != nil {
					return ShadowPairs{}, err
				}
				result.Pairs = append(result.Pairs, pair)
			}
			if strings.TrimSuffix(lvl.Lines[line][col].Terrain, "_w") == "forest" && !crownSelected {
				result.Unmeasured = append(result.Unmeasured, ShadowOmission{At: at, ID: background, Reason: "no_selected_crown"})
			}
		}
	}
	return result, nil
}

func (r *Renderer) castShadowPair(at hexgrid.Coord, id string, selectedShadow bool, density int) (ShadowPair, error) {
	pair := ShadowPair{At: at, Shadow: Entry{ID: mapcompose.ShadowID(id)}}
	standing, src, err := r.assetImage(id, density)
	if err != nil {
		return ShadowPair{}, err
	}
	pair.Standing = standing
	centroid, standingReason, err := shadowAlphaCentroid(src, standing, at, density, "standing")
	if err != nil {
		return ShadowPair{}, err
	}
	pair.StandingCentroid = centroid
	if !selectedShadow {
		pair.Reason = "no_selected_cast_shadow"
		return pair, nil
	}
	shadow, src, err := r.assetImage(pair.Shadow.ID, density)
	if err != nil {
		return ShadowPair{}, err
	}
	pair.Shadow = shadow
	centroid, shadowReason, err := shadowAlphaCentroid(src, shadow, at, density, "shadow")
	if err != nil {
		return ShadowPair{}, err
	}
	pair.ShadowCentroid = centroid
	pair.Reason = standingReason
	if pair.Reason == "" {
		pair.Reason = shadowReason
	}
	if pair.Reason != "" {
		return pair, nil
	}
	pair.Offset = [2]float64{pair.ShadowCentroid[0] - pair.StandingCentroid[0], pair.ShadowCentroid[1] - pair.StandingCentroid[1]}
	pair.Distance = math.Hypot(pair.Offset[0], pair.Offset[1])
	if math.IsNaN(pair.Distance) || math.IsInf(pair.Distance, 0) {
		return ShadowPair{}, fmt.Errorf("%s: nonfinite centroid displacement", id)
	}
	if pair.Distance == 0 {
		pair.Reason = "zero_offset"
		return pair, nil
	}
	bearing := math.Mod(math.Atan2(pair.Offset[0], -pair.Offset[1])*180/math.Pi+360, 360)
	pair.Bearing = &bearing
	return pair, nil
}

func shadowAlphaCentroid(src image.Image, entry Entry, at hexgrid.Coord, density int, role string) ([2]float64, string, error) {
	x, y := hexgrid.Center(at)
	origin := [2]float64{}
	for i, center := range [2]float64{x, y} {
		if math.IsNaN(entry.Anchor[i]) || math.IsInf(entry.Anchor[i], 0) {
			return [2]float64{}, "", fmt.Errorf("%s: nonfinite anchor", entry.ID)
		}
		origin[i] = math.Round((center-entry.Anchor[i])*float64(density)) / float64(density)
		if math.IsNaN(origin[i]) || math.IsInf(origin[i], 0) {
			return [2]float64{}, "", fmt.Errorf("%s: nonfinite anchor placement", entry.ID)
		}
	}
	bounds := src.Bounds()
	var weight, weightedX, weightedY float64
	clipped := false
	for py := bounds.Min.Y; py < bounds.Max.Y; py++ {
		for px := bounds.Min.X; px < bounds.Max.X; px++ {
			_, _, _, alpha := src.At(px, py).RGBA()
			if alpha == 0 {
				continue
			}
			a := float64(alpha)
			weight += a
			weightedX += (float64(px-bounds.Min.X) + 0.5) * a
			weightedY += (float64(py-bounds.Min.Y) + 0.5) * a
			clipped = clipped || px == bounds.Min.X || py == bounds.Min.Y || px == bounds.Max.X-1 || py == bounds.Max.Y-1
		}
	}
	if weight == 0 {
		return [2]float64{}, "empty_" + role + "_alpha", nil
	}
	centroid := [2]float64{origin[0] + weightedX/weight/float64(density), origin[1] + weightedY/weight/float64(density)}
	if clipped {
		return centroid, "clipped_" + role + "_alpha", nil
	}
	return centroid, "", nil
}
