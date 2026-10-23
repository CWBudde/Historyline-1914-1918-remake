package main

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
)

type lookLight struct {
	Available          bool                `json:"available"`
	Reason             string              `json:"reason,omitempty"`
	GlobalAvailable    bool                `json:"global_illumination_available"`
	GlobalReason       string              `json:"global_illumination_unavailable_reason"`
	Definition         string              `json:"definition"`
	Density            int                 `json:"density"`
	AllSelected        int                 `json:"all_rows_selected_crown_instances"`
	Excluded           int                 `json:"excluded_row_instances"`
	Instances          int                 `json:"included_crown_instances"`
	Clean              int                 `json:"clean_crown_instances"`
	UniquePairs        int                 `json:"unique_source_pairs"`
	UniqueShadowHashes int                 `json:"unique_shadow_hashes"`
	Summary            lookLightSummary    `json:"instance_summary"`
	UniqueSummary      lookLightSummary    `json:"unique_source_pair_summary"`
	Pairs              []lookLightPair     `json:"pairs"`
	Unmeasured         []lookLightOmission `json:"unmeasured_standing_and_forest_cells"`
	Gaps               int                 `json:"compositor_gaps"`
	Issues             int                 `json:"compositor_issues"`
}

type lookLightPair struct {
	Cell             string          `json:"cell"`
	Clean            bool            `json:"clean_cell"`
	Site             [2]float64      `json:"site_offset_world"`
	Standing         maprender.Entry `json:"standing"`
	Shadow           maprender.Entry `json:"shadow"`
	StandingCentroid *[2]float64     `json:"standing_centroid_world"`
	ShadowCentroid   *[2]float64     `json:"shadow_centroid_world"`
	Offset           *[2]float64     `json:"centroid_offset_world"`
	Distance         *float64        `json:"centroid_displacement_world_px"`
	Bearing          *float64        `json:"shadow_compass_bearing_degrees"`
	Reason           string          `json:"reason,omitempty"`
}

type lookLightOmission struct {
	Cell   string `json:"cell"`
	ID     string `json:"standing_or_background_id"`
	Reason string `json:"reason"`
}

type lookLightSummary struct {
	Measured       int      `json:"measured_pairs"`
	Unavailable    int      `json:"unavailable_pairs"`
	MeanBearing    *float64 `json:"circular_mean_bearing_degrees"`
	Concentration  *float64 `json:"circular_resultant_length"`
	MeanDistance   *float64 `json:"mean_centroid_displacement_world_px"`
	MaxTargetError *float64 `json:"maximum_error_from_135_degrees"`
	WithinTen      int      `json:"pairs_within_135_plus_minus_10_degrees"`
	Reason         string   `json:"reason,omitempty"`
}

func lightAngleDifference(a, b float64) float64 { return math.Abs(math.Mod(a-b+540, 360) - 180) }

func summarizeLight(pairs []lookLightPair) lookLightSummary {
	result := lookLightSummary{}
	x, y, length, maxError := 0., 0., 0., 0.
	for _, pair := range pairs {
		if pair.Reason != "" || pair.Bearing == nil {
			result.Unavailable++
			continue
		}
		result.Measured++
		radians := *pair.Bearing * math.Pi / 180
		x += math.Sin(radians)
		y += math.Cos(radians)
		if pair.Distance != nil {
			length += *pair.Distance
		}
		e := lightAngleDifference(*pair.Bearing, 135)
		maxError = max(maxError, e)
		if e <= 10 {
			result.WithinTen++
		}
	}
	if result.Measured == 0 {
		result.Reason = "no_measured_pairs"
		return result
	}
	n := float64(result.Measured)
	concentration := math.Hypot(x, y) / n
	result.Concentration = &concentration
	meanLength := length / n
	result.MeanDistance = &meanLength
	result.MaxTargetError = &maxError
	if concentration < 1e-12 {
		result.Reason = "direction_vectors_cancel"
		return result
	}
	bearing := math.Mod(math.Atan2(x, y)*180/math.Pi+360, 360)
	result.MeanBearing = &bearing
	return result
}

func (p *project) lightFromPairs(raw maprender.ShadowPairs) lookLight {
	result := lookLight{Density: raw.Density, AllSelected: len(raw.Pairs), Gaps: len(raw.Gaps), Issues: len(raw.Issues), Pairs: []lookLightPair{}, Unmeasured: []lookLightOmission{}, GlobalReason: "separate crown/cast-shadow alpha displacement only; embedded shadows, painted form shading, other families and global illumination remain unmeasured", Definition: "every selected baked or scattered crown with its own cast shadow (a scattered crown at its site offset from the hex centre); complete unoccluded hash-verified paired source alpha, weighted pixel-centre centroids placed with runtime density/anchors/rounding; vector is shadow minus crown; magnitude is centroid displacement, not physical shadow length; compass north=0 east=90 clockwise, bearing atan2(dx,-dy); unavailable pairs have null values; both instance and unique SHA+anchor pair summaries use unweighted circular unit-vector means; no independent evidence is inferred from repeated contexts; global light acceptance is not asserted"}
	seen, shadowHashes := map[string]bool{}, map[string]bool{}
	var unique []lookLightPair
	for _, pair := range raw.Pairs {
		if p.excluded[pair.At.Line+1] {
			result.Excluded++
			continue
		}
		c := cell{Row: pair.At.Line + 1, Col: pair.At.Col + 1}
		h, ok := p.hex(c)
		clean := ok && len(h.Infra.IDs) == 0 && h.Content == nil
		v := lookLightPair{Cell: c.String(), Clean: clean, Site: pair.Site, Standing: pair.Standing, Shadow: pair.Shadow, Reason: pair.Reason}
		if pair.Reason == "" && pair.Bearing != nil {
			v.StandingCentroid = &pair.StandingCentroid
			v.ShadowCentroid = &pair.ShadowCentroid
			v.Offset = &pair.Offset
			v.Distance = &pair.Distance
			v.Bearing = pair.Bearing
		}
		result.Pairs = append(result.Pairs, v)
		if clean {
			result.Clean++
		}
		key, _ := json.Marshal([]any{pair.Standing.SHA, pair.Standing.Anchor, pair.Shadow.SHA, pair.Shadow.Anchor})
		if !seen[string(key)] {
			seen[string(key)] = true
			unique = append(unique, v)
		}
		if pair.Shadow.SHA != "" {
			shadowHashes[pair.Shadow.SHA] = true
		}
	}
	for _, omission := range raw.Unmeasured {
		if p.excluded[omission.At.Line+1] {
			continue
		}
		c := cell{Row: omission.At.Line + 1, Col: omission.At.Col + 1}
		result.Unmeasured = append(result.Unmeasured, lookLightOmission{c.String(), omission.ID, omission.Reason})
	}
	result.Instances = len(result.Pairs)
	result.UniquePairs = len(unique)
	result.UniqueShadowHashes = len(shadowHashes)
	result.Summary = summarizeLight(result.Pairs)
	result.UniqueSummary = summarizeLight(unique)
	result.Available = result.Summary.Measured > 0
	if !result.Available {
		result.Reason = "no_measured_selected_pairs"
	}
	return result
}

func (p *project) measureLight(r *maprender.Renderer, density int) (lookLight, error) {
	raw, err := r.CastShadowPairs(p.level, density)
	if err != nil {
		return lookLight{}, fmt.Errorf("paired cast shadows: %w", err)
	}
	return p.lightFromPairs(raw), nil
}
