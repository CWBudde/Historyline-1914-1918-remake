// Command hlkit prepares the selected cut-out sheets (crowns, props, stones,
// wire) as world-scale sprites: it clears the alpha haze, splits the sheet
// into its items, scales each to its world size and grades its colour. It
// writes review PNGs and a report with the checks; it does not select art.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/worldmaterial"
)

// cell is one sheet cell: the long side of its visible item (alpha at or
// above the cut) in world px (0: the sheet's scale) and an optional colour
// target of its opaque mean.
type cell struct {
	Cell  string      `json:"cell"`
	Size  float64     `json:"size,omitempty"`
	Grade *[3]float64 `json:"grade_rgb,omitempty"`
}

// sheet is one selected cut-out master. scale is its painting scale in
// source px per world px; grade_rgb is the target of the whole sheet's
// opaque mean, applied before a cell's own target; grade_reach (CIELAB) spares
// the colours far from the item's, such as a stone's waterline (0: none).
type sheet struct {
	ID      string      `json:"id"`
	Art     string      `json:"art"`
	Source  string      `json:"source"`
	SHA     string      `json:"source_sha256"`
	Grid    [2]int      `json:"grid"`
	Cut     uint8       `json:"cut"`
	MinArea int         `json:"min_area"`
	Scale   float64     `json:"scale"`
	Grade   *[3]float64 `json:"grade_rgb,omitempty"`
	Reach   float64     `json:"grade_reach,omitempty"`
	Cells   []cell      `json:"cells"`
}

type source struct {
	Source string `json:"source"`
	SHA    string `json:"source_sha256"`
}

type profile struct {
	Schema    int     `json:"schema_version"`
	Densities []int   `json:"densities"`
	Meadow    source  `json:"meadow"`
	Sheets    []sheet `json:"sheets"`
}

type output struct {
	Density int        `json:"density"`
	Path    string     `json:"path"`
	SHA     string     `json:"sha256"`
	Width   int        `json:"width"`
	Height  int        `json:"height"`
	Anchor  [2]float64 `json:"logical_anchor"`
}

type itemReport struct {
	Cell        string                `json:"cell"`
	Bounds      [4]int                `json:"source_bounds"`
	Components  int                   `json:"components"`
	ExtentFrom  float64               `json:"extent_world_px_at_painting_scale"`
	Extent      float64               `json:"extent_world_px"`
	WorldPerSrc float64               `json:"world_px_per_source_px"`
	MeanFrom    [3]float64            `json:"mean_rgb_before"`
	Mean        [3]float64            `json:"mean_rgb"`
	Own         *[3]float64           `json:"graded_colour,omitempty"`
	Outputs     []output              `json:"outputs"`
	Checks      []worldmaterial.Check `json:"checks"`
}

type sheetReport struct {
	ID        string                `json:"id"`
	Art       string                `json:"art"`
	Source    string                `json:"source"`
	SourceSHA string                `json:"source_sha256"`
	Canvas    [2]int                `json:"canvas"`
	HazeFrom  float64               `json:"haze_before"`
	Haze      float64               `json:"haze"`
	MeanFrom  [3]float64            `json:"mean_rgb_before"`
	Mean      [3]float64            `json:"mean_rgb"`
	Own       *[3]float64           `json:"graded_colour,omitempty"`
	Contact   []string              `json:"contact_sheets"`
	Items     []itemReport          `json:"items"`
	Checks    []worldmaterial.Check `json:"checks"`
	Pass      bool                  `json:"pass"`
}

func main() {
	p := flag.String("profile", "art/pipeline/world-kits-v1.json", "pinned source profile")
	out := flag.String("out", "build/review/world-kits", "output directory")
	strict := flag.Bool("check", false, "fail unless every kit check passes (the report is written anyway)")
	flag.Parse()
	if err := run(*p, *out, *strict); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func readSource(path, sha string) (*image.NRGBA, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if sum := fmt.Sprintf("%x", sha256.Sum256(b)); sum != sha {
		return nil, fmt.Errorf("%s: sha256 %s, profile pins %s", path, sum, sha)
	}
	im, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	out := image.NewNRGBA(image.Rect(0, 0, im.Bounds().Dx(), im.Bounds().Dy()))
	draw.Draw(out, out.Bounds(), im, im.Bounds().Min, draw.Src)
	return out, nil
}

// parseCell reads "r<row>c<col>", 1-based, into a 0-based column and row.
func parseCell(s string, grid [2]int) (image.Point, error) {
	var r, c int
	if n, err := fmt.Sscanf(s, "r%dc%d", &r, &c); n != 2 || err != nil || fmt.Sprintf("r%dc%d", r, c) != s || c < 1 || r < 1 || c > grid[0] || r > grid[1] {
		return image.Point{}, fmt.Errorf("cell %q outside the %dx%d grid", s, grid[0], grid[1])
	}
	return image.Pt(c-1, r-1), nil
}

func cellName(p image.Point) string { return fmt.Sprintf("r%dc%d", p.Y+1, p.X+1) }

func validRGB(c *[3]float64) bool {
	if c == nil {
		return true
	}
	for _, v := range c {
		if math.IsNaN(v) || v < 0 || v > 255 {
			return false
		}
	}
	return true
}

func validate(p profile) error {
	if p.Schema != 1 || len(p.Densities) != 2 || p.Densities[0] != 2 || p.Densities[1] != 4 || p.Meadow.Source == "" {
		return fmt.Errorf("unsupported world kit profile")
	}
	ids := map[string]bool{}
	for _, s := range p.Sheets {
		if s.ID == "" || ids[s.ID] || strings.Contains(s.ID, "..") {
			return fmt.Errorf("missing, duplicate or unsafe kit id %q", s.ID)
		}
		ids[s.ID] = true
		if !strings.HasPrefix(s.Source, "art/"+s.Art+"/") {
			return fmt.Errorf("%s: source outside art/%s", s.ID, s.Art)
		}
		if s.Grid[0] < 1 || s.Grid[1] < 1 || s.Grid[0] > 8 || s.Grid[1] > 8 || s.Cut == 0 || s.MinArea < 1 || !(s.Scale > 0) || !(s.Reach >= 0) || len(s.Cells) == 0 || !validRGB(s.Grade) {
			return fmt.Errorf("%s: grid 1..8, cut > 0, min_area >= 1, scale > 0, rgb 0..255 and cells", s.ID)
		}
		seen := map[string]bool{}
		for _, c := range s.Cells {
			if _, err := parseCell(c.Cell, s.Grid); err != nil || seen[c.Cell] {
				return fmt.Errorf("%s: invalid or duplicate cell %q", s.ID, c.Cell)
			}
			seen[c.Cell] = true
			if math.IsNaN(c.Size) || c.Size < 0 || !validRGB(c.Grade) {
				return fmt.Errorf("%s %s: size >= 0 and rgb 0..255", s.ID, c.Cell)
			}
		}
	}
	return nil
}

// alphaBounds is the box of the pixels with alpha >= from.
func alphaBounds(im *image.NRGBA, from uint8) image.Rectangle {
	var r image.Rectangle
	b := im.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if im.NRGBAAt(x, y).A >= from {
				r = r.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	return r
}

func longSide(r image.Rectangle) float64 { return float64(max(r.Dx(), r.Dy())) }

func round2(v float64) float64 { return math.Round(v*100) / 100 }
func round4(v float64) float64 { return math.Round(v*1e4) / 1e4 }

func roundRGB(c [3]float64) [3]float64 {
	return [3]float64{round2(c[0]), round2(c[1]), round2(c[2])}
}

func check(name string, v float64, want string, pass bool) worldmaterial.Check {
	return worldmaterial.Check{Name: name, Value: round2(v), Want: want, Pass: pass}
}

// rgbCheck is the largest per-channel distance of an item's own graded
// colour from its target.
func rgbCheck(mean, target [3]float64) worldmaterial.Check {
	d := 0.
	for i := range mean {
		d = math.Max(d, math.Abs(mean[i]-target[i]))
	}
	return check(fmt.Sprintf("graded colour within 8 of RGB %g,%g,%g", target[0], target[1], target[2]), d, "<= 8", d <= 8)
}

func writePNG(path string, im image.Image) (string, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(b.Bytes())), os.WriteFile(path, b.Bytes(), 0644)
}

// prepared is one item's sprites by density.
type prepared map[int]*image.NRGBA

func prepare(s sheet, src *image.NRGBA, densities []int, outDir string) (sheetReport, []prepared, error) {
	rep := sheetReport{ID: s.ID, Art: s.Art, Source: s.Source, SourceSHA: s.SHA, Canvas: [2]int{src.Bounds().Dx(), src.Bounds().Dy()}}
	rep.HazeFrom, rep.MeanFrom = round4(worldmaterial.Haze(src, s.Cut)), roundRGB(worldmaterial.OpaqueMean(src))
	cut := worldmaterial.Cut(src, s.Cut)
	if s.Grade != nil {
		g, own, err := worldmaterial.GradeTo(cut, *s.Grade, s.Reach)
		if err != nil {
			return rep, nil, fmt.Errorf("%s: %w", s.ID, err)
		}
		cut, own = g, roundRGB(own)
		rep.Own = &own
	}
	rep.Haze, rep.Mean = round4(worldmaterial.Haze(cut, s.Cut)), roundRGB(worldmaterial.OpaqueMean(cut))
	rep.Checks = append(rep.Checks, check("haze after the cut", rep.Haze, "0", rep.Haze == 0))
	if s.Grade != nil {
		rep.Checks = append(rep.Checks, rgbCheck(*rep.Own, *s.Grade))
	}
	items, err := worldmaterial.Items(cut, s.Grid[0], s.Grid[1], s.MinArea)
	if err != nil {
		return rep, nil, fmt.Errorf("%s: %w", s.ID, err)
	}
	byCell := map[image.Point]worldmaterial.Item{}
	listed := map[image.Point]bool{}
	for _, it := range items {
		byCell[it.Cell] = it
	}
	var sprites []prepared
	for _, c := range s.Cells {
		at, _ := parseCell(c.Cell, s.Grid)
		listed[at] = true
		ir := itemReport{Cell: c.Cell}
		it, ok := byCell[at]
		ir.Checks = append(ir.Checks, check("cell holds an item", float64(it.Area), "> 0", ok))
		if !ok {
			rep.Items = append(rep.Items, ir)
			continue
		}
		im := it.Image(cut)
		ir.MeanFrom = roundRGB(worldmaterial.OpaqueMean(im))
		if c.Grade != nil {
			var own [3]float64
			if im, own, err = worldmaterial.GradeTo(im, *c.Grade, s.Reach); err != nil {
				return rep, nil, fmt.Errorf("%s %s: %w", s.ID, c.Cell, err)
			}
			own = roundRGB(own)
			ir.Own = &own
		}
		ir.Mean = roundRGB(worldmaterial.OpaqueMean(im))
		// The visible item: the cut has cleared everything fainter.
		ob := it.Bounds
		ir.Bounds = [4]int{it.Bounds.Min.X, it.Bounds.Min.Y, it.Bounds.Max.X, it.Bounds.Max.Y}
		ir.Components = it.Components
		ir.ExtentFrom = round2(longSide(ob) / s.Scale)
		ir.WorldPerSrc = 1 / s.Scale
		if c.Size > 0 {
			ir.WorldPerSrc = c.Size / longSide(ob)
		}
		edge := it.Bounds.Min.X == 0 || it.Bounds.Min.Y == 0 || it.Bounds.Max.X == src.Bounds().Dx() || it.Bounds.Max.Y == src.Bounds().Dy()
		ir.Checks = append(ir.Checks, check("item clear of the canvas edge", 0, "true", !edge))
		if c.Grade != nil {
			ir.Checks = append(ir.Checks, rgbCheck(*ir.Own, *c.Grade))
		}
		sp := prepared{}
		for _, d := range densities {
			k := ir.WorldPerSrc * float64(d)
			w, h := max(1, int(math.Round(float64(it.Bounds.Dx())*k))), max(1, int(math.Round(float64(it.Bounds.Dy())*k)))
			out, err := worldmaterial.Resample(im, it.Bounds, w, h)
			if err != nil {
				return rep, nil, fmt.Errorf("%s %s: %w", s.ID, c.Cell, err)
			}
			sp[d] = out
			name := filepath.ToSlash(filepath.Join(s.ID, fmt.Sprintf("%s@%dx.png", c.Cell, d)))
			sum, err := writePNG(filepath.Join(outDir, filepath.FromSlash(name)), out)
			if err != nil {
				return rep, nil, err
			}
			anchor := [2]float64{round2((it.Centroid[0] - float64(it.Bounds.Min.X)) * ir.WorldPerSrc), round2((it.Centroid[1] - float64(it.Bounds.Min.Y)) * ir.WorldPerSrc)}
			ir.Outputs = append(ir.Outputs, output{d, name, sum, w, h, anchor})
		}
		top := densities[len(densities)-1]
		// Downscaled thin blades and wires fall below the opaque alpha, so the
		// result is measured at the cut: what still shows.
		ir.Extent = round2(longSide(alphaBounds(sp[top], s.Cut)) / float64(top))
		if c.Size > 0 {
			tol := math.Max(0.5, 0.05*c.Size)
			ir.Checks = append(ir.Checks, check(fmt.Sprintf("extent within %g of %g world px", round2(tol), c.Size), math.Abs(ir.Extent-c.Size), fmt.Sprintf("<= %g", round2(tol)), math.Abs(ir.Extent-c.Size) <= tol))
		}
		sprites = append(sprites, sp)
		rep.Items = append(rep.Items, ir)
	}
	for _, it := range items {
		if !listed[it.Cell] {
			rep.Checks = append(rep.Checks, check("item in an unlisted cell "+cellName(it.Cell), float64(it.Area), "0", false))
		}
	}
	rep.Pass = worldmaterial.Passed(rep.Checks)
	for _, ir := range rep.Items {
		rep.Pass = rep.Pass && worldmaterial.Passed(ir.Checks)
	}
	return rep, sprites, nil
}

// contact lays the sheet's sprites at density d out in its grid shape over
// the meadow, so sizes read against the ground they will stand on.
func contact(meadow *image.NRGBA, sprites []prepared, cols, d int) *image.NRGBA {
	cw, ch := 8*d, 8*d
	for _, sp := range sprites {
		cw, ch = max(cw, sp[d].Bounds().Dx()+8*d), max(ch, sp[d].Bounds().Dy()+8*d)
	}
	rows := (len(sprites) + cols - 1) / cols
	out := image.NewNRGBA(image.Rect(0, 0, cols*cw, max(1, rows)*ch))
	mb := meadow.Bounds()
	for y := 0; y < out.Bounds().Dy(); y++ {
		for x := 0; x < out.Bounds().Dx(); x++ {
			out.SetNRGBA(x, y, meadow.NRGBAAt(x%mb.Dx(), y%mb.Dy()))
		}
	}
	for i, sp := range sprites {
		im := sp[d]
		at := image.Pt((i%cols)*cw+(cw-im.Bounds().Dx())/2, (i/cols)*ch+(ch-im.Bounds().Dy())/2)
		draw.Draw(out, im.Bounds().Add(at), im, image.Point{}, draw.Over)
	}
	return out
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
	ground, err := readSource(p.Meadow.Source, p.Meadow.SHA)
	if err != nil {
		return err
	}
	period := image.Pt(worldmaterial.PeriodWidth, worldmaterial.PeriodHeight)
	meadows := map[int]*image.NRGBA{}
	for _, d := range p.Densities {
		if meadows[d], _, err = worldmaterial.Export(ground, period, d); err != nil {
			return fmt.Errorf("meadow: %w", err)
		}
	}
	var reports []sheetReport
	var failed []string
	for _, s := range p.Sheets {
		src, err := readSource(s.Source, s.SHA)
		if err != nil {
			return err
		}
		rep, sprites, err := prepare(s, src, p.Densities, outDir)
		if err != nil {
			return err
		}
		for _, d := range p.Densities {
			name := filepath.ToSlash(filepath.Join(s.ID, fmt.Sprintf("contact@%dx.png", d)))
			if _, err = writePNG(filepath.Join(outDir, filepath.FromSlash(name)), contact(meadows[d], sprites, s.Grid[0], d)); err != nil {
				return err
			}
			rep.Contact = append(rep.Contact, name)
		}
		if !rep.Pass {
			failed = append(failed, s.ID)
		}
		reports = append(reports, rep)
	}
	report := struct {
		Schema    int           `json:"schema_version"`
		Algorithm string        `json:"algorithm"`
		Sheets    []sheetReport `json:"sheets"`
	}{p.Schema, "alpha below cut cleared; sheet graded to grade_rgb by one Lab offset; 8-connected alpha components grouped by the cell of their centroid, specks under min_area dropped; cell graded to its grade_rgb; opaque long side scaled to size (or 1/scale); exact area box in premultiplied alpha per density", reports}
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(outDir, "kits.json"), append(b, '\n'), 0644); err != nil {
		return err
	}
	if strict && len(failed) > 0 {
		return errors.New("kit checks fail: " + strings.Join(failed, ", ") + " (see kits.json)")
	}
	return nil
}
