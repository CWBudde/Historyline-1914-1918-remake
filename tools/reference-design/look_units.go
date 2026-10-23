package main

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
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/lookmeasure"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

type lookUnits struct {
	Available  bool              `json:"available"`
	Reason     string            `json:"reason,omitempty"`
	Scope      string            `json:"scope"`
	Definition string            `json:"definition"`
	Sprites    []lookUnitSprite  `json:"sprites"`
	Fixtures   []lookUnitFixture `json:"fixtures"`
	Cases      []lookUnitCase    `json:"cases"`
}

type lookUnitSprite struct {
	ID      string     `json:"id"`
	SHA256  string     `json:"source_sha256"`
	Density int        `json:"density"`
	Width   int        `json:"width"`
	Height  int        `json:"height"`
	Anchor  [2]float64 `json:"logical_anchor"`
}

type lookUnitFixture struct {
	ID           string `json:"id"`
	LevelSHA256  string `json:"canonical_level_sha256"`
	PixelsSHA256 string `json:"rgba_pixels_sha256"`
	Gaps         int    `json:"compositor_gaps"`
	Issues       int    `json:"compositor_issues"`
}

type lookUnitCase struct {
	Sprite        string            `json:"sprite"`
	Fixture       string            `json:"fixture"`
	Stats         unitContrastStats `json:"edge_contrast"`
	Preview       string            `json:"preview,omitempty"`
	PreviewSHA256 string            `json:"preview_sha256,omitempty"`
}

type unitContrastStats struct {
	Samples    int      `json:"samples"`
	Minimum    *float64 `json:"minimum_l_star"`
	P10        *float64 `json:"p10_l_star"`
	Median     *float64 `json:"median_l_star"`
	P90        *float64 `json:"p90_l_star"`
	Fraction30 *float64 `json:"fraction_delta_at_least_30"`
	Reason     string   `json:"reason,omitempty"`
}

type unitEntry struct {
	maprender.Entry
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Crop only after placing with the runtime anchor, and align the crop with the
// global export-density blocks. Unit silhouettes are never clipped to a hex.
func unitPatch(background *image.NRGBA, sprite image.Image, at hexgrid.Coord, density int, anchor [2]float64) (*image.NRGBA, *image.NRGBA, *image.Alpha, error) {
	if background == nil || sprite == nil || density != 2 {
		return nil, nil, nil, fmt.Errorf("unit patch requires an opaque 2x background and sprite")
	}
	for _, v := range anchor {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, nil, nil, fmt.Errorf("invalid unit anchor")
		}
	}
	b := sprite.Bounds()
	if b.Empty() {
		return nil, nil, nil, fmt.Errorf("empty unit sprite")
	}
	if err := unclipped(sprite); err != nil {
		return nil, nil, nil, err
	}
	x, y := hexgrid.Center(at)
	origin := image.Pt(int(math.Round((x-anchor[0])*float64(density))), int(math.Round((y-anchor[1])*float64(density))))
	placed := b.Sub(b.Min).Add(origin)
	down := func(v int) int { return int(math.Floor(float64(v)/float64(density)))*density - density }
	up := func(v int) int { return int(math.Ceil(float64(v)/float64(density)))*density + density }
	bounds := image.Rect(down(placed.Min.X), down(placed.Min.Y), up(placed.Max.X), up(placed.Max.Y))
	if bounds.Intersect(background.Bounds()) != bounds {
		return nil, nil, nil, fmt.Errorf("unit footprint outside background")
	}
	before, after := image.NewNRGBA(bounds), image.NewNRGBA(bounds)
	draw.Draw(before, bounds, background, bounds.Min, draw.Src)
	draw.Draw(after, bounds, background, bounds.Min, draw.Src)
	alpha := image.NewAlpha(bounds)
	maprender.Place(after, sprite, at, density, anchor)
	maprender.Place(alpha, sprite, at, density, anchor)
	return before, after, alpha, nil
}

// unclipped reports an error when the silhouette of sprite reaches its canvas edge.
func unclipped(sprite image.Image) error {
	b := sprite.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if x != b.Min.X && x != b.Max.X-1 && y != b.Min.Y && y != b.Max.Y-1 {
				continue
			}
			_, _, _, a := sprite.At(x, y).RGBA()
			if a >= 32768 {
				return fmt.Errorf("unit silhouette clipped at source canvas")
			}
		}
	}
	return nil
}

// dressed is the 2x unit sprite of family as the game's map draws it, with its
// rim and contact shadow (unitsprite.Dress), and its logical anchor.
func dressed(sprite image.Image, family string, anchor [2]float64) (image.Image, [2]float64, error) {
	if err := unclipped(sprite); err != nil {
		return nil, anchor, err
	}
	m := float64(unitsprite.DressMargin(2)) / 2
	return unitsprite.Dress(sprite, 2, unitsprite.Flies(family)), [2]float64{anchor[0] + m, anchor[1] + m}, nil
}

func summarizeUnit(samples []lookmeasure.EdgeSample) unitContrastStats {
	s := unitContrastStats{Samples: len(samples)}
	if len(samples) == 0 {
		s.Reason = "insufficient_samples"
		return s
	}
	values := make([]float64, len(samples))
	passing := 0
	for i, p := range samples {
		values[i] = p.Delta
		if p.Delta >= 30 {
			passing++
		}
	}
	slices.Sort(values)
	quantile := func(p float64) *float64 {
		at := p * float64(len(values)-1)
		lo := int(at)
		hi := min(lo+1, len(values)-1)
		v := values[lo] + (at-float64(lo))*(values[hi]-values[lo])
		return &v
	}
	s.Minimum, s.P10, s.Median, s.P90 = quantile(0), quantile(.1), quantile(.5), quantile(.9)
	fraction := float64(passing) / float64(len(samples))
	s.Fraction30 = &fraction
	return s
}

func unitMeasurement(background *image.NRGBA, sprite image.Image, at hexgrid.Coord, anchor [2]float64) (unitContrastStats, *image.NRGBA, error) {
	before, after, alpha, err := unitPatch(background, sprite, at, 2, anchor)
	if err != nil {
		return unitContrastStats{}, nil, err
	}
	ground, visible := newLabPlane(halve(before)), newLabPlane(halve(after))
	opacity, err := averageCanopy(alpha, 2)
	if err != nil {
		return unitContrastStats{}, nil, err
	}
	body, valid := make([]bool, len(opacity.alpha)), make([]bool, len(opacity.alpha))
	for i, a := range opacity.alpha {
		body[i] = a >= .5
		valid[i] = ground.ok[i] && visible.ok[i]
		if valid[i] {
			// The D65 conversion coefficients produce L*=100.0000039 for
			// sRGB white. Normalize that roundoff at the producer boundary;
			// the measurement backend retains its strict [0,100] contract.
			ground.l[i] = max(0, min(100, ground.l[i]))
			visible.l[i] = max(0, min(100, visible.l[i]))
		}
	}
	edges, err := lookmeasure.EdgeContrast(ground.l, visible.l, body, valid, ground.w, ground.h)
	if err != nil {
		return unitContrastStats{}, nil, err
	}
	return summarizeUnit(edges), after, nil
}

func unitFixture(id string) *level.Level {
	lv := &level.Level{MapName: "unit-contrast-" + id, Lines: make([][]level.Hex, 7)}
	for l := range lv.Lines {
		lv.Lines[l] = make([]level.Hex, 5-l%2)
		for c := range lv.Lines[l] {
			terrain := id
			if id == "road" {
				terrain = "grass"
			}
			lv.Lines[l][c].Terrain = terrain
			if id == "road" && l%2 == 1 && c == 2 {
				lv.Lines[l][c].Infra.IDs = []string{"road36"}
			}
		}
	}
	return lv
}

func (p *project) measureUnits(out string) (lookUnits, error) {
	result := lookUnits{Scope: "all authored families and owner palettes; north-facing normal state only; native 2x runtime art dressed as the map draws it (dark rim and contact shadow, unitsprite.Dress) and composited before 1x area reduction; four deterministic 5x7 scenes, with a continuous road through grass and homogeneous forest/water/grass elsewhere", Definition: "silhouette is mean unit alpha >= 0.5; edge pixels have a four-connected nonbody neighbour; delta = abs(L*(composite) - L*(bare background)) at the same world pixel; clipped or invalid edges are errors; no passing aggregation is asserted", Sprites: []lookUnitSprite{}, Fixtures: []lookUnitFixture{}, Cases: []lookUnitCase{}}
	raw, err := fs.ReadFile(p.assets, "index.json")
	if err != nil {
		return result, err
	}
	var index struct {
		Schema int         `json:"schema_version"`
		Assets []unitEntry `json:"assets"`
	}
	if err := json.Unmarshal(raw, &index); err != nil {
		return result, err
	}
	if index.Schema != 1 {
		return result, fmt.Errorf("unsupported unit catalogue schema")
	}
	entries := map[string]unitEntry{}
	for _, e := range index.Assets {
		if !strings.HasPrefix(e.Pack, "unit-") || e.Density != 2 {
			continue
		}
		if _, exists := entries[e.ID]; exists {
			return result, fmt.Errorf("duplicate unit entry %s", e.ID)
		}
		entries[e.ID] = e
	}
	renderer, err := maprender.New(p.assets)
	if err != nil {
		return result, err
	}
	backgrounds := map[string]*image.NRGBA{}
	for _, id := range []string{"grass", "forest", "road", "water"} {
		lv := unitFixture(id)
		m, err := renderer.Render(lv, 2, false)
		if err != nil {
			return result, err
		}
		data, err := lv.Marshal()
		if err != nil {
			return result, err
		}
		fixture := lookUnitFixture{id, fmt.Sprintf("%x", sha256.Sum256(data)), fmt.Sprintf("%x", sha256.Sum256(m.Image.Pix)), len(m.Gaps), len(m.Issues)}
		if fixture.Gaps != 0 || fixture.Issues != 0 {
			return result, fmt.Errorf("unit fixture %s has %d gaps/%d issues", id, fixture.Gaps, fixture.Issues)
		}
		backgrounds[id] = m.Image
		result.Fixtures = append(result.Fixtures, fixture)
	}
	examples := map[string]bool{"infantry": true, "armored-car": true, "a7v": true, "battleship": true, "spad-xiii": true}
	for _, family := range unitsprite.Families() {
		for _, palette := range unitsprite.Palettes {
			id := unitsprite.ID(family, palette, unitsprite.N, unitsprite.Normal)
			e, ok := entries[id]
			if !ok {
				return result, fmt.Errorf("missing 2x unit %s", id)
			}
			data, err := fs.ReadFile(p.assets, e.Path)
			if err != nil {
				return result, err
			}
			if fmt.Sprintf("%x", sha256.Sum256(data)) != e.SHA {
				return result, fmt.Errorf("unit %s source hash mismatch", id)
			}
			src, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				return result, err
			}
			if src.Bounds().Dx() != e.Width || src.Bounds().Dy() != e.Height {
				return result, fmt.Errorf("unit %s dimensions differ from catalogue", id)
			}
			result.Sprites = append(result.Sprites, lookUnitSprite{id, e.SHA, 2, e.Width, e.Height, e.Anchor})
			sprite, anchor, err := dressed(src, family, e.Anchor)
			if err != nil {
				return result, fmt.Errorf("%s: %w", id, err)
			}
			for _, fixture := range result.Fixtures {
				stats, composed, err := unitMeasurement(backgrounds[fixture.ID], sprite, hexgrid.Coord{Line: 3, Col: 2}, anchor)
				if err != nil {
					return result, fmt.Errorf("%s on %s: %w", id, fixture.ID, err)
				}
				c := lookUnitCase{Sprite: id, Fixture: fixture.ID, Stats: stats}
				if out != "" && examples[family] && palette == unitsprite.CP {
					c.Preview = "units/" + strings.ReplaceAll(id, "/", "-") + "-" + fixture.ID + "-2x.png"
					var b bytes.Buffer
					if err := png.Encode(&b, composed); err != nil {
						return result, err
					}
					dest := filepath.Join(out, c.Preview)
					if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
						return result, err
					}
					if err := os.WriteFile(dest, b.Bytes(), 0644); err != nil {
						return result, err
					}
					c.PreviewSHA256 = fmt.Sprintf("%x", sha256.Sum256(b.Bytes()))
				}
				result.Cases = append(result.Cases, c)
			}
		}
	}
	result.Available = len(result.Cases) > 0
	if !result.Available {
		result.Reason = "unit_catalogue_unavailable"
	}
	return result, nil
}
