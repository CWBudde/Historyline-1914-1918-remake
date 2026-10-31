package referencedesign

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math"
	"os"
	"slices"
	"strings"
)

// tolerance is how far a rendered family may drift from its reference targets:
// delta_e is the CIE76 distance of the mean colours, the others are relative.
type tolerance struct {
	DeltaE   float64 `json:"delta_e"`
	Contrast float64 `json:"contrast"`
	Chroma   float64 `json:"chroma"`
	Edges    float64 `json:"edges"`
}

func (t tolerance) valid() bool {
	return t.DeltaE > 0 && t.Contrast > 0 && t.Chroma > 0 && t.Edges > 0
}

// outside names the metrics of got that lie outside the tolerance around want.
func (t tolerance) outside(want, got metrics) []string {
	out := []string{}
	if got.deltaE(want) > t.DeltaE {
		out = append(out, "colour")
	}
	rel := func(name string, w, g, tol float64) {
		if math.Abs(g-w) > tol*math.Max(w, 0.01) {
			out = append(out, name)
		}
	}
	rel("contrast", want.Contrast, got.Contrast, t.Contrast)
	rel("chroma", want.Chroma, got.Chroma, t.Chroma)
	rel("edges", want.Edges, got.Edges, t.Edges)
	return out
}

// targets is targets.json next to reference.json: the reference metrics of each
// family's exemplars, written by -out so -check works without the image.
type targets struct {
	Image    string   `json:"image_sha256"`
	Families []target `json:"families"`
}

type target struct {
	ID        string   `json:"id"`
	Exemplars []string `json:"exemplars"`
	metrics
}

func (p *project) writeTargets(cmp *comparison) error {
	t := targets{Image: p.ref.Image.SHA256}
	for _, f := range cmp.Families {
		t.Families = append(t.Families, target{ID: f.ID, Exemplars: f.Exemplars, metrics: f.Reference})
	}
	var b bytes.Buffer
	b.WriteString("{\n  \"image_sha256\": ")
	sha, _ := json.Marshal(t.Image)
	b.Write(sha)
	b.WriteString(",\n  \"families\": [\n")
	for i, f := range t.Families {
		line, err := json.Marshal(f)
		if err != nil {
			return err
		}
		b.WriteString("    ")
		b.Write(line)
		if i < len(t.Families)-1 {
			b.WriteByte(',')
		}
		b.WriteByte('\n')
	}
	b.WriteString("  ]\n}\n")
	return os.WriteFile(p.targetsPath, b.Bytes(), 0o644)
}

// loadTargets returns nil without error when targets.json does not exist yet.
func (p *project) loadTargets() (*targets, error) {
	raw, err := os.ReadFile(p.targetsPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var t targets
	if err := dec.Decode(&t); err != nil {
		return nil, fmt.Errorf("%s: %w", p.targetsPath, err)
	}
	return &t, nil
}

// checkTargets reports targets that no longer match reference.json; they need a
// new measurement with the image (just reference).
func (p *project) checkTargets(t *targets) []error {
	var errs []error
	fail := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf("targets: "+format+" (measure again with just reference)", args...))
	}
	if t.Image != p.ref.Image.SHA256 {
		fail("measured on image %.12s…, reference.json names %.12s…", t.Image, p.ref.Image.SHA256)
	}
	byID := map[string]target{}
	for _, f := range t.Families {
		byID[f.ID] = f
	}
	for _, f := range p.ref.Families {
		got, ok := byID[f.ID]
		switch {
		case !ok:
			fail("family %s has no target", f.ID)
		case !slices.Equal(got.Exemplars, f.Exemplars):
			fail("family %s was measured on exemplars %s, reference.json lists %s", f.ID, strings.Join(got.Exemplars, ","), strings.Join(f.Exemplars, ","))
		}
		delete(byID, f.ID)
	}
	for _, id := range slices.Sorted(maps.Keys(byID)) {
		fail("family %s is not in reference.json", id)
	}
	return errs
}

type drift struct {
	ID      string
	Want    metrics
	Got     metrics
	Outside []string
	DeltaE  float64
}

// drift measures the current render against the targets.
func (p *project) drift(t *targets) ([]drift, error) {
	render, err := p.renderMap()
	if err != nil {
		return nil, err
	}
	got := p.renderMetrics(render)
	var out []drift
	for _, f := range t.Families {
		g := got[f.ID]
		out = append(out, drift{ID: f.ID, Want: f.metrics, Got: g, DeltaE: round2(g.deltaE(f.metrics)), Outside: p.ref.Tolerance.outside(f.metrics, g)})
	}
	return out, nil
}

func driftTable(rows []drift) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-10s %6s  %-17s %-17s %-13s %-13s %-11s %s\n", "family", "ΔE", "L a b target", "L a b render", "contrast", "chroma", "edges", "outside")
	for _, d := range rows {
		fmt.Fprintf(&b, "%-10s %6.1f  %5.1f %5.1f %5.1f %5.1f %5.1f %5.1f  %5.1f→%-5.1f  %5.1f→%-5.1f  %4.2f→%-4.2f  %s\n",
			d.ID, d.DeltaE, d.Want.L, d.Want.A, d.Want.B, d.Got.L, d.Got.A, d.Got.B,
			d.Want.Contrast, d.Got.Contrast, d.Want.Chroma, d.Got.Chroma, d.Want.Edges, d.Got.Edges, strings.Join(d.Outside, ","))
	}
	return b.String()
}
