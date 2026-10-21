package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	xdraw "golang.org/x/image/draw"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

//go:embed page.html
var pageSource string

var page = template.Must(template.New("page").Parse(pageSource))

const (
	tileDensity = 2
	thumbName   = "reference.png"
	thumbScale  = 2 // the thumbnail is the reference at half size
	compareDir  = "compare"
)

type tileView struct {
	ID, File, Info string
	Left, Top      int // assembled position in display pixels (features only)
}

type familyView struct {
	ID, Kind, Keys, ArtLink, Art, Notes string
	Exemplars, All                      []tileView
	Map                                 template.HTML
	Compare                             familyCompare
}

type featureView struct {
	ID, Title, Notes, Crop string
	Tiles                  []tileView
	Width, Height          int
	Map                    template.HTML
}

type pageView struct {
	Image, Level    string
	Width, Height   int
	ThumbW, ThumbH  int
	Thumb           string
	Cells, Excluded int
	Families        []familyView
	Features        []featureView
	Compare         *comparison
}

// render writes the tiles, a reference thumbnail and index.html into out.
func (p *project) render(out string) error {
	src, present, err := p.source()
	if err != nil {
		return err
	}
	if !present {
		return fmt.Errorf("reference image %s is missing: place the local copy at the repository root (sha256 %s)", p.ref.Image.Path, p.ref.Image.SHA256)
	}

	if err := os.RemoveAll(filepath.Join(out, "tiles")); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(out, "tiles"), 0o755); err != nil {
		return err
	}
	view := pageView{Image: p.ref.Image.Path, Level: p.ref.Level, Width: p.ref.Image.Width, Height: p.ref.Image.Height, Thumb: thumbName}
	for _, c := range p.cells() {
		if p.excluded[c.Row] {
			view.Excluded++
			continue
		}
		view.Cells++
		ox, oy := tileOrigin(c.Row-1, c.Col-1)
		im, _ := renderTile(src, p.ref.Registration, ox, oy, tileDensity)
		if err := writePNG(filepath.Join(out, "tiles", c.fileName()), im); err != nil {
			return err
		}
	}

	b := src.Bounds()
	thumb := image.NewNRGBA(image.Rect(0, 0, (b.Dx()+thumbScale-1)/thumbScale, (b.Dy()+thumbScale-1)/thumbScale))
	xdraw.CatmullRom.Scale(thumb, thumb.Bounds(), src, b, xdraw.Src, nil)
	if err := writePNG(filepath.Join(out, thumbName), thumb); err != nil {
		return err
	}
	view.ThumbW, view.ThumbH = thumb.Bounds().Dx(), thumb.Bounds().Dy()

	rootRel, err := relPath(out, p.root)
	if err != nil {
		return err
	}
	cmp, err := p.compare(src, filepath.Join(out, compareDir), formatPNG)
	if err != nil {
		return err
	}
	if err := p.writeTargets(cmp); err != nil {
		return err
	}
	view.Compare = cmp
	for i, f := range p.ref.Families {
		fv := familyView{ID: f.ID, Kind: f.Kind, Art: f.Art, Notes: f.Notes, ArtLink: rootRel + "/" + f.Art + "/prompt.md", Compare: cmp.Families[i]}
		if f.Kind == "terrain" {
			fv.Keys = strings.Join(f.RuntimeIDs, ", ")
		} else {
			fv.Keys = strings.Join(f.InfraPrefixes, "*, ") + "*"
		}
		var exemplars []cell
		for _, s := range f.Exemplars {
			c, _ := parseCell(s)
			exemplars = append(exemplars, c)
			fv.Exemplars = append(fv.Exemplars, p.tileView(c))
		}
		occ := p.occurrences(f)
		for _, c := range occ {
			fv.All = append(fv.All, p.tileView(c))
		}
		fv.Map = p.overlay(occ, exemplars)
		view.Families = append(view.Families, fv)
	}
	for i, ft := range p.ref.Features {
		fv := p.featureView(ft)
		fv.Crop = cmp.Features[i].Crop
		view.Features = append(view.Features, fv)
	}

	var html bytes.Buffer
	if err := page.Execute(&html, view); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, "index.html"), html.Bytes(), 0o644)
}

func (p *project) tileView(c cell) tileView {
	h, _ := p.hex(c)
	return tileView{ID: c.String(), File: "tiles/" + c.fileName(), Info: describe(h)}
}

// featureView places the feature's tiles at their true hex positions, in display
// pixels at 1.5× the game scale.
func (p *project) featureView(ft feature) featureView {
	const num, den = 3, 2
	fv := featureView{ID: ft.ID, Title: ft.Title, Notes: ft.Notes}
	var cells []cell
	minX, minY, maxX, maxY := 1<<30, 1<<30, -1<<30, -1<<30
	for _, s := range ft.Cells {
		c, _ := parseCell(s)
		cells = append(cells, c)
		x, y := tileOrigin(c.Row-1, c.Col-1)
		minX, minY, maxX, maxY = min(minX, x), min(minY, y), max(maxX, x), max(maxY, y)
	}
	for _, c := range cells {
		x, y := tileOrigin(c.Row-1, c.Col-1)
		t := p.tileView(c)
		t.Left, t.Top = (x-minX)*num/den, (y-minY)*num/den
		fv.Tiles = append(fv.Tiles, t)
	}
	fv.Width = (maxX - minX + hexgrid.HexWidth) * num / den
	fv.Height = (maxY - minY + hexgrid.HexHeight) * num / den
	fv.Map = p.overlay(cells, nil)
	return fv
}

// overlay draws the reference thumbnail with the given cells outlined in source
// coordinates; marked cells are filled more strongly.
func (p *project) overlay(cells, marked []cell) template.HTML {
	strong := map[cell]bool{}
	for _, c := range marked {
		strong[c] = true
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<svg class="map" viewBox="0 0 %d %d" role="img" aria-label="Locations on the reference"><image href="%s" width="%d" height="%d"/>`,
		p.ref.Image.Width, p.ref.Image.Height, thumbName, p.ref.Image.Width, p.ref.Image.Height)
	for _, c := range cells {
		ox, oy := tileOrigin(c.Row-1, c.Col-1)
		class := "hit"
		if strong[c] {
			class = "hit strong"
		}
		fmt.Fprintf(&b, `<polygon class="%s" points="`, class)
		for i, v := range hexPolygon {
			sx, sy := p.ref.Registration.sourcePoint(float64(ox)+v[0], float64(oy)+v[1])
			if i > 0 {
				b.WriteByte(' ')
			}
			fmt.Fprintf(&b, "%.1f,%.1f", sx, sy)
		}
		fmt.Fprintf(&b, `"><title>%s</title></polygon>`, c)
	}
	b.WriteString(`</svg>`)
	return template.HTML(b.String()) // numbers and cell ids only
}

func writePNG(path string, im image.Image) error {
	var b bytes.Buffer
	if err := (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(&b, im); err != nil {
		return err
	}
	return os.WriteFile(path, b.Bytes(), 0o644)
}

// relPath returns the slash-separated path from directory from to target.
func relPath(from, target string) (string, error) {
	a, err := filepath.Abs(from)
	if err != nil {
		return "", err
	}
	t, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	r, err := filepath.Rel(a, t)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(r), nil
}

// source decodes the verified reference image without its exterior black canvas;
// present is false when the local image does not exist.
func (p *project) source() (*image.NRGBA, bool, error) {
	raw, present, err := p.checkImage()
	if err != nil || !present {
		return nil, present, err
	}
	decoded, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, true, err
	}
	return prepareSource(decoded), true, nil
}
