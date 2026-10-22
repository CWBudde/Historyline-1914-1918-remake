// Command hlmaterial exports the selected meadow and scrub as reusable periodic
// world materials. It writes PNGs and a provenance report; it does not select art.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/worldmaterial"
)

type material struct {
	ID     string `json:"id"`
	Art    string `json:"art"`
	Source string `json:"source"`
	SHA    string `json:"source_sha256"`
}
type profile struct {
	Schema    int        `json:"schema_version"`
	Period    [2]int     `json:"period"`
	Densities []int      `json:"densities"`
	Band      float64    `json:"edge_band_world_pixels"`
	Materials []material `json:"materials"`
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
}

func main() {
	p := flag.String("profile", "art/pipeline/world-materials-v1.json", "pinned source profile")
	out := flag.String("out", "build/review/world-materials", "output directory (paths match assets/)")
	flag.Parse()
	if err := run(*p, *out); err != nil {
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
func run(profilePath, outDir string) error {
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
	if p.Schema != 1 || p.Period != [2]int{worldmaterial.PeriodWidth, worldmaterial.PeriodHeight} || len(p.Densities) != 2 || p.Densities[0] != 2 || p.Densities[1] != 4 || p.Band != 4 || len(p.Materials) != 2 {
		return fmt.Errorf("unsupported world material profile")
	}
	seen := map[string]bool{}
	var outputs []output
	for _, m := range p.Materials {
		if (m.ID != worldmaterial.MeadowID && m.ID != worldmaterial.ScrubID) || seen[m.ID] {
			return fmt.Errorf("unknown or duplicate material %q", m.ID)
		}
		seen[m.ID] = true
		family := strings.Split(m.ID, "/")[1]
		if m.Art != "terrain/"+family {
			return fmt.Errorf("%s: mismatched art family", m.ID)
		}
		b, e := os.ReadFile(m.Source)
		if e != nil {
			return e
		}
		if sum := fmt.Sprintf("%x", sha256.Sum256(b)); sum != m.SHA {
			return fmt.Errorf("%s: sha256 %s, profile pins %s", m.Source, sum, m.SHA)
		}
		src, e := png.Decode(bytes.NewReader(b))
		if e != nil {
			return e
		}
		for _, d := range p.Densities {
			im, c, e := exportMaterial(src, image.Pt(p.Period[0], p.Period[1]), d, p.Band)
			if e != nil {
				return fmt.Errorf("%s: %w", m.ID, e)
			}
			name := fmt.Sprintf("terrain/%s/%s-summer-material-%dx.png", family, family, d)
			dst := filepath.Join(outDir, filepath.FromSlash(name))
			if e = os.MkdirAll(filepath.Dir(dst), 0755); e != nil {
				return e
			}
			var encoded bytes.Buffer
			if e = png.Encode(&encoded, im); e != nil {
				return e
			}
			if e = os.WriteFile(dst, encoded.Bytes(), 0644); e != nil {
				return e
			}
			outputs = append(outputs, output{ID: m.ID, Pack: fmt.Sprintf("terrain-materials-%dx", d), Density: d, Path: name, SHA: fmt.Sprintf("%x", sha256.Sum256(encoded.Bytes())), Width: im.Bounds().Dx(), Height: im.Bounds().Dy(), Art: m.Art, Source: m.Source, SourceSHA: m.SHA, Crop: c})
		}
	}
	report := struct {
		Schema    int      `json:"schema_version"`
		Algorithm string   `json:"algorithm"`
		Period    [2]int   `json:"period"`
		Band      float64  `json:"edge_band_world_pixels"`
		Outputs   []output `json:"outputs"`
	}{1, "native aspect centre crop; exact area box; smoothstep opposite endpoint feather", p.Period, p.Band, outputs}
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "materials.json"), append(b, '\n'), 0644)
}
