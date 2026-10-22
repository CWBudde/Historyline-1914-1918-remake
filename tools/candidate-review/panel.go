package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Unit panel geometry at 2x (tools/unit-panel compose: portrait at logical
// (16,64) 210×188, tactical unit at (266,94) 96×96 inside a box from (242,94)).
var (
	panelCrop    = image.Rect(16, 112, 784, 528)
	portraitRect = image.Rect(32, 128, 452, 504)
	unitBox      = image.Rect(486, 190, 766, 378)
	unitAt       = image.Pt(532, 188)
)

// panelBase renders the unit panel fixture once with tools/unit-panel.
func (r *renderer) panelBase() (*image.NRGBA, error) {
	if r.panel != nil {
		return r.panel, nil
	}
	out := filepath.Join(r.work, "panel")
	file := filepath.Join(out, "fixtures", "cp-ready-2x.png")
	if _, err := os.Stat(file); err != nil {
		_ = os.RemoveAll(out)
		cmd := exec.Command("go", "run", "./tools/unit-panel", "-manifest", "art/pipeline/unit-panel-v1.json", "-out", out)
		cmd.Dir = r.o.root
		if msg, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("unit-panel: %v: %s", err, msg)
		}
	}
	img, err := loadPNG(file)
	if err != nil {
		return nil, err
	}
	r.panel = image.NewNRGBA(img.Bounds())
	draw.Draw(r.panel, r.panel.Bounds(), img, img.Bounds().Min, draw.Src)
	return r.panel, nil
}

// inPanel puts a 420×376 portrait and the family's tactical sprite into the panel.
func (r *renderer) inPanel(portrait image.Image, unit image.Image) (*image.NRGBA, error) {
	base, err := r.panelBase()
	if err != nil {
		return nil, err
	}
	p := image.NewNRGBA(base.Bounds())
	draw.Draw(p, p.Bounds(), base, image.Point{}, draw.Src)
	draw.Draw(p, portraitRect, cover(portrait, portraitRect.Dx(), portraitRect.Dy()), image.Point{}, draw.Src)
	if unit != nil {
		fill := base.At(unitBox.Min.X+4, unitBox.Min.Y+4)
		draw.Draw(p, unitBox, &image.Uniform{fill}, image.Point{}, draw.Src)
		draw.Draw(p, unit.Bounds().Sub(unit.Bounds().Min).Add(unitAt), unit, unit.Bounds().Min, draw.Over)
	}
	return crop(p, panelCrop), nil
}

func (r *renderer) portraits(g *group) error {
	unit, _ := loadPNG(filepath.Join(r.o.root, "assets", "units", g.Family, "cp", "n-normal-2x.png"))
	if installed, err := loadPNG(filepath.Join(r.o.root, "assets", "portraits", g.Family, "portrait-2x.png")); err == nil {
		p, err := r.inPanel(installed, unit)
		if err != nil {
			return err
		}
		rel, err := r.media(g, "installed-panel.jpg", p)
		if err != nil {
			return err
		}
		g.Installed = &view{Label: "in the unit panel", Media: map[string]string{"img": rel}}
	}
	for _, c := range g.Candidates {
		src, err := r.rawThumb(g, c)
		if err != nil {
			return err
		}
		p, err := r.inPanel(src, unit)
		if err != nil {
			return err
		}
		rel, err := r.media(g, c.Exec+"-panel.jpg", p)
		if err != nil {
			return err
		}
		c.Views = append([]*view{{Label: "in the unit panel (centre crop 105:94)", Media: map[string]string{"img": rel}}}, c.Views...)
	}
	return nil
}

// ---- interface and screen sources ----

// counterpart finds the installed runtime file at 2x for a master such as
// art/interface/rank/07.png, falling back to the master itself.
func (r *renderer) counterpart(master string) (image.Image, string) {
	if master == "" {
		return nil, ""
	}
	key := path.Dir(strings.TrimPrefix(master, "art/"))
	name := strings.TrimSuffix(path.Base(master), ".png")
	for _, e := range r.assets {
		if e.Art == key && e.Density == 2 && path.Base(e.ID) == name {
			if img, err := loadPNG(filepath.Join(r.o.root, "assets", filepath.FromSlash(e.Path))); err == nil {
				return img, "assets/" + e.Path
			}
		}
	}
	img, err := loadPNG(filepath.Join(r.o.root, master))
	if err != nil {
		return nil, ""
	}
	return img, master
}

func (r *renderer) ui(g *group) error {
	for _, c := range g.Candidates {
		src, err := r.rawThumb(g, c)
		if err != nil {
			return err
		}
		if c.Group == "backdrop" {
			if err := r.backdrop(g, c); err != nil {
				return err
			}
			continue
		}
		inst, from := r.counterpart(c.Near)
		if inst == nil {
			continue
		}
		cand := crop(src, content(src))
		var pair *image.NRGBA
		if strings.HasPrefix(from, "assets/interface/map/") {
			pair, err = r.mapOverlayPair(inst, cand)
			if err != nil {
				return err
			}
		} else {
			pair = sideBySide(inst, cand)
		}
		rel, err := r.media(g, c.Exec+"-pair.jpg", pair)
		if err != nil {
			return err
		}
		c.Views = append([]*view{{Label: "left installed " + from + ", right candidate at the same size", Media: map[string]string{"img": rel}}}, c.Views...)
	}
	return nil
}

// sideBySide shows the installed item at its runtime size and the candidate's content
// scaled into the same box, both on the interface slate.
func sideBySide(inst, cand image.Image) *image.NRGBA {
	b := inst.Bounds()
	w, h := b.Dx(), b.Dy()
	if w > 900 || h > 700 { // large panels: halve both for the page
		w, h = w/2, h/2
	}
	pad := 24
	out := image.NewNRGBA(image.Rect(0, 0, 2*w+3*pad, h+2*pad))
	draw.Draw(out, out.Bounds(), &image.Uniform{slate}, image.Point{}, draw.Src)
	draw.Draw(out, image.Rect(pad, pad, pad+w, pad+h), fit(inst, w, h), image.Point{}, draw.Over)
	draw.Draw(out, image.Rect(2*pad+w, pad, 2*pad+2*w, pad+h), fit(cand, w, h), image.Point{}, draw.Over)
	return out
}

// mapOverlayPair places a hex overlay (240×200 at 2x) on the map target hex, installed left.
func (r *renderer) mapOverlayPair(inst, cand image.Image) (*image.NRGBA, error) {
	if err := r.prepare(); err != nil {
		return nil, err
	}
	st := r.land
	anchor := [2]float64{60, 50}
	views := make([]*image.NRGBA, 2)
	for i, src := range []image.Image{inst, fit(cand, inst.Bounds().Dx(), inst.Bounds().Dy())} {
		m := image.NewNRGBA(st.img.Bounds())
		draw.Draw(m, m.Bounds(), st.img, image.Point{}, draw.Src)
		placeOverlay(m, src, st.target, anchor)
		views[i] = crop(m, st.window)
	}
	w, h := views[0].Bounds().Dx(), views[0].Bounds().Dy()
	out := image.NewNRGBA(image.Rect(0, 0, 2*w+16, h))
	draw.Draw(out, out.Bounds(), &image.Uniform{slate}, image.Point{}, draw.Src)
	draw.Draw(out, image.Rect(0, 0, w, h), views[0], image.Point{}, draw.Src)
	draw.Draw(out, image.Rect(w+16, 0, 2*w+16, h), views[1], image.Point{}, draw.Src)
	return out, nil
}

func placeOverlay(dst draw.Image, src image.Image, at hexgrid.Coord, anchor [2]float64) {
	x, y := hexgrid.Center(at)
	p := image.Pt(int((x-anchor[0])*2), int((y-anchor[1])*2))
	draw.Draw(dst, src.Bounds().Sub(src.Bounds().Min).Add(p), src, src.Bounds().Min, draw.Over)
}

// backdrop renders the screen fixtures with the candidate as the menu backdrop
// through tools/interface-art, next to the installed backdrop.
func (r *renderer) backdrop(g *group, c *candidate) error {
	render := func(name, source, sum string, cropRect []int) (string, error) {
		out := filepath.Join(r.work, "screens-"+name)
		if _, err := os.Stat(filepath.Join(out, "fixtures", "loading-050-2x.png")); err == nil {
			return out, nil
		}
		raw, err := os.ReadFile(filepath.Join(r.o.root, "art", "pipeline", "screens-v2.json"))
		if err != nil {
			return "", err
		}
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			return "", err
		}
		if source != "" {
			b := m["backdrop"].(map[string]any)
			b["source"], b["sha256"], b["crop"] = source, sum, cropRect
		}
		_ = os.RemoveAll(out)
		if err := os.MkdirAll(r.work, 0o755); err != nil {
			return "", err
		}
		file := filepath.Join(r.work, "screens-"+name+".json")
		data, _ := json.MarshalIndent(m, "", "  ")
		if err := os.WriteFile(file, data, 0o644); err != nil {
			return "", err
		}
		cmd := exec.Command("go", "run", "./tools/interface-art", "-manifest", file, "-out", out)
		cmd.Dir = r.o.root
		if msg, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("interface-art: %v: %s", err, msg)
		}
		return out, nil
	}
	src, err := loadPNG(c.raw)
	if err != nil {
		return err
	}
	b := src.Bounds()
	n := min(b.Dx()/8, b.Dy()/5) // exact 8:5
	w, h := 8*n, 5*n
	x0, y0 := (b.Dx()-w)/2, (b.Dy()-h)/2
	sum, err := fileSHA(c.raw)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(r.o.root, c.raw)
	if err != nil {
		return err
	}
	inst, err := render("installed", "", "", nil)
	if err != nil {
		return err
	}
	cand, err := render(c.Exec, filepath.ToSlash(rel), sum, []int{x0, y0, w, h})
	if err != nil {
		c.Verdict = err.Error()
		return nil
	}
	v := &view{Label: "as the menu backdrop", Media: map[string]string{}}
	iv := &view{Label: "as the menu backdrop", Media: map[string]string{}}
	for _, f := range []string{"loading-050", "confirm-exit"} {
		for _, side := range []struct {
			dir string
			v   *view
			pre string
		}{{inst, iv, "installed"}, {cand, v, c.Exec}} {
			img, err := loadPNG(filepath.Join(side.dir, "fixtures", f+"-2x.png"))
			if err != nil {
				return err
			}
			rel, err := r.media(g, side.pre+"-"+f+".jpg", halve(img))
			if err != nil {
				return err
			}
			side.v.Media[f] = rel
		}
	}
	g.Installed = iv
	c.Views = append([]*view{v}, c.Views...)
	return nil
}
