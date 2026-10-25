// Command hlmaterial exports the selected meadow and scrub as reusable periodic
// world materials. It writes PNGs and a provenance report with the meadow
// checks; it does not select art.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/worldmaterial"
)

// detail is a close-up layer whose high-passed L* is wrapped repeats×repeats
// over the material and added with gain (profile schema 2, meadow only). The
// tone is graded first: tone_contrast scales its L* about its mean (omitted:
// 1), then tone_offset adds to its L*, a* and b*.
type detail struct {
	Source       string     `json:"source"`
	SHA          string     `json:"source_sha256"`
	Repeats      int        `json:"repeats"`
	Gain         float64    `json:"gain"`
	ToneContrast float64    `json:"tone_contrast,omitempty"`
	ToneOffset   [3]float64 `json:"tone_offset,omitzero"`
}

// toneContrast is the tone's L* factor, 1 when the profile omits it.
func (d *detail) toneContrast() float64 {
	if d.ToneContrast == 0 {
		return 1
	}
	return d.ToneContrast
}

type material struct {
	ID     string  `json:"id"`
	Art    string  `json:"art"`
	Source string  `json:"source"`
	SHA    string  `json:"source_sha256"`
	Detail *detail `json:"detail,omitempty"`
}
type profile struct {
	Schema    int        `json:"schema_version"`
	Period    [2]int     `json:"period"`
	Densities []int      `json:"densities"`
	Band      float64    `json:"edge_band_world_pixels"`
	Materials []material `json:"materials"`
}
type detailOutput struct {
	Source       string             `json:"source"`
	SourceSHA    string             `json:"source_sha256"`
	Crop         worldmaterial.Crop `json:"source_crop"`
	Repeats      int                `json:"repeats"`
	Gain         float64            `json:"gain"`
	ToneContrast float64            `json:"tone_contrast"`
	ToneOffset   [3]float64         `json:"tone_offset"`
}
type output struct {
	ID        string             `json:"id"`
	Pack      string             `json:"pack"`
	Density   int                `json:"density"`
	Path      string             `json:"path"`
	SHA       string             `json:"sha256"`
	Width     int                `json:"width"`
	Height    int                `json:"height"`
	Anchor    [2]float64         `json:"logical_anchor"`
	Art       string             `json:"art"`
	Source    string             `json:"source"`
	SourceSHA string             `json:"source_sha256"`
	Crop      worldmaterial.Crop `json:"source_crop"`
	Detail    *detailOutput      `json:"detail,omitempty"`
}

// checks are the meadow's measurements at 1×: tone and detail only with a
// detail layer.
type checks struct {
	ID            string                `json:"id"`
	SuggestedGain *float64              `json:"suggested_gain,omitempty"`
	Tone          []worldmaterial.Check `json:"tone,omitempty"`
	Detail        []worldmaterial.Check `json:"detail,omitempty"`
	Result        []worldmaterial.Check `json:"result"`
	Pass          bool                  `json:"pass"`
}

func main() {
	p := flag.String("profile", "art/pipeline/world-materials-v1.json", "pinned source profile")
	out := flag.String("out", "build/review/world-materials", "output directory (paths match assets/)")
	strict := flag.Bool("check", false, "fail unless every meadow check passes (the report is written anyway)")
	flag.Parse()
	if err := run(*p, *out, *strict); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func exportMaterial(src image.Image, period image.Point, density int, band float64) (*image.NRGBA, worldmaterial.Crop, error) {
	im, c, err := worldmaterial.Export(src, period, density)
	if err != nil {
		return nil, c, err
	}
	im, err = worldmaterial.Seamless(im, band, density)
	return im, c, err
}

// layers is one material's sources: tone, and with a detail block the detail.
type layers struct {
	tone, detail image.Image
	d            *detail
}

// build returns the material at density with its tone and detail layers
// before the edge feather (detail is nil without a detail block).
func (l layers) build(period image.Point, density int, band float64) (result, tone, det *image.NRGBA, c, dc worldmaterial.Crop, err error) {
	if l.d == nil {
		result, c, err = exportMaterial(l.tone, period, density, band)
		return result, nil, nil, c, dc, err
	}
	if tone, c, err = worldmaterial.Export(l.tone, period, density); err != nil {
		return
	}
	if k := l.d.toneContrast(); k != 1 || l.d.ToneOffset != [3]float64{} {
		if tone, err = worldmaterial.Grade(tone, k, l.d.ToneOffset); err != nil {
			return
		}
	}
	if tone, err = worldmaterial.Seamless(tone, worldmaterial.ToneBand, density); err != nil {
		return
	}
	if det, dc, err = worldmaterial.Tiled(l.detail, period, density, l.d.Repeats, band); err != nil {
		return
	}
	// The tone is seamless and the detail wraps a whole number of times, so the
	// sum continues across the period; a second cross-fade would make the
	// sparse detail denser along the edges.
	result, err = worldmaterial.Combine(tone, det, density, l.d.Gain)
	return
}

func readSource(path, sha string) (image.Image, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if sum := fmt.Sprintf("%x", sha256.Sum256(b)); sum != sha {
		return nil, fmt.Errorf("%s: sha256 %s, profile pins %s", path, sum, sha)
	}
	return png.Decode(bytes.NewReader(b))
}

func validate(p profile) error {
	if (p.Schema != 1 && p.Schema != 2) || p.Period != [2]int{worldmaterial.PeriodWidth, worldmaterial.PeriodHeight} || len(p.Densities) != 2 || p.Densities[0] != 2 || p.Densities[1] != 4 || p.Band != 4 || len(p.Materials) != 2 {
		return fmt.Errorf("unsupported world material profile")
	}
	seen := map[string]bool{}
	for _, m := range p.Materials {
		if (m.ID != worldmaterial.MeadowID && m.ID != worldmaterial.ScrubID) || seen[m.ID] {
			return fmt.Errorf("unknown or duplicate material %q", m.ID)
		}
		seen[m.ID] = true
		if m.Art != "terrain/"+strings.Split(m.ID, "/")[1] {
			return fmt.Errorf("%s: mismatched art family", m.ID)
		}
		if d := m.Detail; d != nil {
			if p.Schema != 2 || m.ID != worldmaterial.MeadowID {
				return fmt.Errorf("%s: a detail layer needs schema 2 and the meadow", m.ID)
			}
			if d.Repeats < 2 || d.Repeats > 8 || math.IsNaN(d.Gain) || math.IsInf(d.Gain, 0) || d.Gain < 0 {
				return fmt.Errorf("%s: detail repeats 2..8 and a finite gain >= 0", m.ID)
			}
			if k := d.ToneContrast; math.IsNaN(k) || k < 0 || k > 3 || (k > 0 && k < 0.25) {
				return fmt.Errorf("%s: tone contrast 0.25..3", m.ID)
			}
			for _, v := range d.ToneOffset {
				if math.IsNaN(v) || math.Abs(v) > 20 {
					return fmt.Errorf("%s: tone offset within ±20 per Lab axis", m.ID)
				}
			}
		}
	}
	return nil
}

func run(profilePath, outDir string, strict bool) error {
	raw, err := os.ReadFile(profilePath)
	if err != nil {
		return err
	}
	var p profile
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err = dec.Decode(&p); err != nil {
		return err
	}
	if err = validate(p); err != nil {
		return err
	}
	period := image.Pt(p.Period[0], p.Period[1])
	var outputs []output
	var measured []checks
	for _, m := range p.Materials {
		l := layers{d: m.Detail}
		if l.tone, err = readSource(m.Source, m.SHA); err != nil {
			return err
		}
		if m.Detail != nil {
			if l.detail, err = readSource(m.Detail.Source, m.Detail.SHA); err != nil {
				return err
			}
		}
		family := strings.Split(m.ID, "/")[1]
		for _, d := range p.Densities {
			im, _, _, c, dc, e := l.build(period, d, p.Band)
			if e != nil {
				return fmt.Errorf("%s: %w", m.ID, e)
			}
			name := fmt.Sprintf("terrain/%s/%s-summer-material-%dx.png", family, family, d)
			dst := filepath.Join(outDir, filepath.FromSlash(name))
			if e = os.MkdirAll(filepath.Dir(dst), 0o755); e != nil {
				return e
			}
			var encoded bytes.Buffer
			if e = png.Encode(&encoded, im); e != nil {
				return e
			}
			if e = os.WriteFile(dst, encoded.Bytes(), 0o644); e != nil {
				return e
			}
			o := output{ID: m.ID, Pack: fmt.Sprintf("terrain-materials-%dx", d), Density: d, Path: name, SHA: fmt.Sprintf("%x", sha256.Sum256(encoded.Bytes())), Width: im.Bounds().Dx(), Height: im.Bounds().Dy(), Art: m.Art, Source: m.Source, SourceSHA: m.SHA, Crop: c}
			if m.Detail != nil {
				o.Detail = &detailOutput{m.Detail.Source, m.Detail.SHA, dc, m.Detail.Repeats, m.Detail.Gain, m.Detail.toneContrast(), m.Detail.ToneOffset}
			}
			outputs = append(outputs, o)
		}
		if m.ID == worldmaterial.MeadowID {
			ch, e := measure(l, period, p.Band)
			if e != nil {
				return fmt.Errorf("%s: %w", m.ID, e)
			}
			measured = append(measured, ch)
		}
	}
	algorithm := "native aspect centre crop; exact area box; contrast-keeping cross-fade into the half-period shift at the edges"
	if p.Schema == 2 {
		algorithm += fmt.Sprintf("; detail: tone L* scaled about its mean by tone_contrast and tone_offset added in Lab, then cross-faded over %d world px at the edges; seamless centre crop wrapped repeats x repeats, L* high-pass by three periodic box blurs of radius %d world px, times gain onto the tone's L*", worldmaterial.ToneBand, worldmaterial.HighPassRadius)
	}
	report := struct {
		Schema    int      `json:"schema_version"`
		Algorithm string   `json:"algorithm"`
		Period    [2]int   `json:"period"`
		Band      float64  `json:"edge_band_world_pixels"`
		Outputs   []output `json:"outputs"`
		Checks    []checks `json:"checks"`
	}{p.Schema, algorithm, p.Period, p.Band, outputs, measured}
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(outDir, "materials.json"), append(b, '\n'), 0o644); err != nil {
		return err
	}
	if strict {
		var failed []string
		for _, ch := range measured {
			if !ch.Pass {
				failed = append(failed, ch.ID)
			}
		}
		if len(failed) > 0 {
			return errors.New("meadow checks fail: " + strings.Join(failed, ", ") + " (see materials.json)")
		}
	}
	return nil
}

// measure builds the material at 1× and runs the meadow checks.
func measure(l layers, period image.Point, band float64) (checks, error) {
	result, tone, det, _, _, err := l.build(period, 1, band)
	if err != nil {
		return checks{}, err
	}
	ch := checks{ID: worldmaterial.MeadowID}
	repeats := 1
	if l.d != nil {
		repeats = l.d.Repeats
		ch.Tone, ch.Detail = worldmaterial.ToneChecks(tone), worldmaterial.DetailChecks(det)
		g := math.Round(worldmaterial.SuggestGain(tone, det)*1000) / 1000
		ch.SuggestedGain = &g
	}
	ch.Result = worldmaterial.ResultChecks(result, period, repeats)
	ch.Pass = worldmaterial.Passed(ch.Tone) && worldmaterial.Passed(ch.Detail) && worldmaterial.Passed(ch.Result)
	return ch, nil
}
