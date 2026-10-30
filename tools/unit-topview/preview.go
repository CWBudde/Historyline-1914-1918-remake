package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
	xdraw "golang.org/x/image/draw"
)

// previewLevel is the campaign map whose bunkers the preview crops (its
// cp setup).
const previewLevel = "levels/1916-02"

// renderPreview renders the installed family with the runtime art: every unit of
// the family on previewLevel, and a summer and a winter field with each owner's
// sprite, normal and spent, next to infantry for scale. Each image is written at
// 2x and reduced to 1x and 0.5x.
func renderPreview(root, out, family string) error {
	if clean := filepath.ToSlash(filepath.Clean(out)); !strings.HasPrefix(clean, "build/") && !strings.Contains(clean, "/build/") {
		return fmt.Errorf("-preview must be under build/")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	assetsDir := filepath.Join(root, "assets")
	r, err := maprender.New(os.DirFS(assetsDir))
	if err != nil {
		return err
	}
	lv, err := scenario.LoadLevel(filepath.Join(root, previewLevel), scenario.CP)
	if err != nil {
		return err
	}
	m, err := r.Render(lv, 2, true)
	if err != nil {
		return err
	}
	n := 0
	for line, hs := range lv.Lines {
		for col, hx := range hs {
			if !strings.HasSuffix(hx.Unit, "_"+family) {
				continue
			}
			n++
			if err := writeScaled(out, fmt.Sprintf("map-%02d", n), crop(m.Image, hexgrid.Coord{Line: line, Col: col}, 3, 2)); err != nil {
				return err
			}
		}
	}
	for _, season := range []string{"grass", "grass_w"} {
		field := &level.Level{MapName: "preview", Lines: make([][]level.Hex, 5)}
		for l := range field.Lines {
			field.Lines[l] = make([]level.Hex, 6-l%2)
			for c := range field.Lines[l] {
				field.Lines[l][c].Terrain = season
			}
		}
		fm, err := r.Render(field, 2, false)
		if err != nil {
			return err
		}
		place := func(at hexgrid.Coord, path string) error {
			im, err := readPNG(filepath.Join(assetsDir, path))
			if err != nil {
				return err
			}
			maprender.Place(fm.Image, im, at, 2, [2]float64{anchorX, anchorY})
			return nil
		}
		// Line 1: cp, ec and neutral; line 3: cp and ec spent, infantry of both sides.
		for i, p := range []string{
			"units/%s/cp/n-normal-2x.png", "units/%s/ec/n-normal-2x.png", "units/%s/neutral/n-normal-2x.png",
			"units/%s/cp/n-spent-2x.png", "units/%s/ec/n-spent-2x.png", "units/infantry/cp/s-normal-2x.png",
		} {
			at := hexgrid.Coord{Line: 1 + 2*(i/3), Col: 1 + i%3}
			if strings.Contains(p, "%s") {
				p = fmt.Sprintf(p, family)
			}
			if err := place(at, p); err != nil {
				return err
			}
		}
		if err := place(hexgrid.Coord{Line: 2, Col: 4}, "units/infantry/ec/s-normal-2x.png"); err != nil {
			return err
		}
		if err := writeScaled(out, "field-"+season, fm.Image); err != nil {
			return err
		}
	}
	fmt.Printf("preview: %d %s units on %s, two fields, in %s\n", n, family, previewLevel, out)
	return nil
}

// crop is the 2x window of w hexes either side and h lines above and below at.
func crop(im *image.NRGBA, at hexgrid.Coord, w, h float64) image.Image {
	x, y := hexgrid.Center(at)
	r := image.Rect(int(x-w*71)*2, int(y-h*40-40)*2, int(x+w*71)*2, int(y+h*40+40)*2).Intersect(im.Bounds())
	return im.SubImage(r)
}

// writeScaled writes name-2x.png and its reductions to 1x and 0.5x.
func writeScaled(dir, name string, im image.Image) error {
	b := im.Bounds()
	for _, s := range []struct {
		label string
		div   int
	}{{"2x", 1}, {"1x", 2}, {"0.5x", 4}} {
		dst := image.NewNRGBA(image.Rect(0, 0, b.Dx()/s.div, b.Dy()/s.div))
		if s.div == 1 {
			draw.Draw(dst, dst.Bounds(), im, b.Min, draw.Src)
		} else {
			xdraw.CatmullRom.Scale(dst, dst.Bounds(), im, b, xdraw.Src, nil)
		}
		f, err := os.Create(filepath.Join(dir, name+"-"+s.label+".png"))
		if err != nil {
			return err
		}
		if err := png.Encode(f, dst); err != nil {
			_ = f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}

func readPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return png.Decode(f)
}
