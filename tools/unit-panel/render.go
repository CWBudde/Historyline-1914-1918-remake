package main

import (
	"fmt"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	"image"
	"image/draw"
	"math"
	"strings"
)

// All coordinates and bounds are logical pixels; Density scales both vectors and fonts.
type Draw struct {
	Kind       string     `json:"kind"`
	Path       string     `json:"path,omitempty"`
	Input      string     `json:"input,omitempty"`
	SHA256     string     `json:"sha256,omitempty"`
	X          float64    `json:"x"`
	Y          float64    `json:"y"`
	Width      float64    `json:"width,omitempty"`
	Height     float64    `json:"height,omitempty"`
	Density    int        `json:"density"`
	Text       string     `json:"text,omitempty"`
	Full       string     `json:"full_string,omitempty"`
	Font       string     `json:"font,omitempty"`
	FontPath   string     `json:"font_path,omitempty"`
	FontSHA256 string     `json:"font_sha256,omitempty"`
	Size       float64    `json:"size,omitempty"`
	Baseline   float64    `json:"baseline,omitempty"`
	HTMLTop    float64    `json:"html_top,omitempty"`
	Color      string     `json:"color,omitempty"`
	Bounds     [4]float64 `json:"bounds"`
	Clip       [4]float64 `json:"clip"`
}
type FixtureRender struct {
	ID      string `json:"id"`
	Density int    `json:"density"`
	Output  string `json:"output"`
	SHA256  string `json:"sha256"`
	Draws   []Draw `json:"draws"`
}

func newFace(f *opentype.Font, size float64, d int) (font.Face, error) {
	return opentype.NewFace(f, &opentype.FaceOptions{Size: size * float64(d), DPI: 72, Hinting: font.HintingNone})
}
func textWidth(f font.Face, s string) float64 { return float64(font.MeasureString(f, s)) / 64 }
func fitText(f font.Face, s string, max float64) string {
	if textWidth(f, s) <= max {
		return s
	}
	r := []rune(s)
	for len(r) > 0 {
		r = r[:len(r)-1]
		v := strings.TrimSpace(string(r)) + "…"
		if textWidth(f, v) <= max {
			return v
		}
	}
	if textWidth(f, "…") <= max {
		return "…"
	}
	return ""
}
func titleLines(f font.Face, s string) (string, string) {
	if textWidth(f, s) <= 368 {
		return s, ""
	}
	r := []rune(s)
	n := 0
	for n < len(r) && textWidth(f, string(r[:n+1])) <= 368 {
		n++
	}
	if n == 0 {
		return "", fitText(f, s, 368)
	}
	split := n
	for i := n - 1; i > n/2; i-- {
		if r[i] == ' ' {
			split = i
			break
		}
	}
	return strings.TrimSpace(string(r[:split])), fitText(f, strings.TrimSpace(string(r[split:])), 368)
}
func compose(in *Inputs, p Panel, d int, assets map[string]Asset) ([]Draw, error) {
	var records []Draw
	img := func(path, input string, x, y, w, h float64) {
		a := assets[path]
		records = append(records, Draw{Kind: "image", Path: path, Input: input, SHA256: a.SHA256, X: x, Y: y, Width: w, Height: h, Density: d, Bounds: [4]float64{x, y, x + w, y + h}, Clip: [4]float64{x, y, x + w, y + h}})
	}
	text := func(s, full, fontName string, size, x, baseline, max float64, c string) error {
		f, e := newFace(in.Fonts[fontName], size, 1)
		if e != nil {
			return e
		}
		defer func() { _ = f.Close() }()
		s = fitText(f, s, max)
		bounds, _ := font.BoundString(f, s)
		left := float64(bounds.Min.X) / 64
		right := float64(bounds.Max.X) / 64
		if left < 0 {
			x -= left
		}
		if right-left > max {
			return fmt.Errorf("glyph ink exceeds text region: %q", s)
		}
		// Ink bounds are measured again at native density for independent pixel reassembly.
		native, e := newFace(in.Fonts[fontName], size, d)
		if e != nil {
			return e
		}
		defer func() { _ = native.Close() }()
		nb, _ := font.BoundString(native, s)
		den := 64 * float64(d)
		b := [4]float64{x + float64(nb.Min.X)/den, baseline + float64(nb.Min.Y)/den, x + float64(nb.Max.X)/den, baseline + float64(nb.Max.Y)/den}
		item := in.item(fontName)
		metrics := f.Metrics()
		top := baseline - (size+float64(metrics.Ascent-metrics.Descent)/64)/2
		records = append(records, Draw{Kind: "text", Text: s, Full: full, Font: fontName, FontPath: item.Output, FontSHA256: item.SHA256, Size: size, X: x, Y: baseline - size, Baseline: baseline, HTMLTop: top, Color: c, Density: d, Bounds: b, Clip: [4]float64{x, baseline - size - 2, x + max, baseline + size*.3}})
		return nil
	}
	var first error
	tx := func(s, full, f string, size, x, y, w float64, c string) {
		if e := text(s, full, f, size, x, y, w, c); e != nil && first == nil {
			first = e
		}
	}
	imageAsset := func(name string, x, y, w, h float64) {
		path := fmt.Sprintf("assets/%s.%dx.png", name, d)
		img(path, "vector-source.json#"+strings.TrimPrefix(name, "icons/"), x, y, w, h)
	}
	inputAsset := func(key string, x, y, w, h float64) {
		item := in.item(fmt.Sprintf("%s-%dx", key, d))
		img(item.Output, item.Source, x, y, w, h)
	}
	icon := func(name string, x, y float64) { imageAsset("icons/"+name, x, y, 24, 24) }
	imageAsset("frame", 0, 0, 400, 640)
	if !p.Available {
		tx("No unit selected", "No unit selected", "regular", 18, 16, 46, 368, cream)
		return records, first
	}
	f, e := newFace(in.Fonts["bold"], 18, 1)
	if e != nil {
		return nil, e
	}
	line1, line2 := titleLines(f, p.Title)
	_ = f.Close()
	tx(line1, p.Title, "bold", 18, 16, 26, 368, cream)
	if line2 != "" {
		tx(line2, p.Title, "bold", 18, 16, 46, 368, cream)
	}
	inputAsset("portrait", 16, 64, 210, 188)
	inputAsset("unit-"+p.Tactical, 266, 94, 96, 96)
	ownerColor := brass
	if p.Tactical == "ec" {
		ownerColor = "#9fc2d7"
	}
	if p.Tactical == "neutral" {
		ownerColor = "#b4bebd"
	}
	tx(p.Owner, p.Owner, "bold", 12, 250, 82, 126, ownerColor)
	icon("rank", 242, 190)
	tx(fmt.Sprintf("Rank %d", p.Rank), "", "regular", 16, 272, 208, 112, cream)
	tx(p.Country, p.Country, "regular", 14, 242, 229, 142, cream)
	if p.Capture {
		icon("capture", 242, 234)
		tx("Capture", "Capture", "regular", 14, 272, 251, 112, cream)
	}
	for i, s := range p.Stats {
		x := float64(16 + (i%3)*126)
		y := float64(268 + (i/3)*64)
		icon(s.Icon, x+8, y+13)
		tx(s.Label, s.Label, "regular", 11, x+40, y+19, 68, "#bcc9cc")
		tx(s.Value, s.Value, "bold", 20, x+40, y+43, 68, cream)
	}
	tx("Weapons", "Weapons", "bold", 16, 16, 414, 368, cream)
	counts := [3]int{}
	for i, w := range p.Weapons {
		y := float64(426 + i*46)
		icon(w.Icon, 24, y+10)
		if w.Empty {
			tx("Empty", "Empty", "regular", 14, 60, y+20, 316, "#a7b5bc")
			continue
		}
		tx(w.Name, w.Name, "bold", 14, 60, y+20, 316, cream)
		classes := []string{}
		if w.Air {
			classes = append(classes, "Air")
			counts[0]++
		}
		if w.Ground {
			classes = append(classes, "Ground")
			counts[1]++
		}
		if w.Water {
			classes = append(classes, "Water")
			counts[2]++
		}
		metric := fmt.Sprintf("Power %d · Range %d–%d · Ammo %s · %s", w.Power, w.Min, w.Max, w.Ammo, strings.Join(classes, "/"))
		tx(metric, metric, "regular", 11, 60, y+37, 316, "#c2ced0")
	}
	tx(p.Action, p.Action, "regular", 12, 16, 634, 204, cream)
	summary := fmt.Sprintf("Air %d · Ground %d · Water %d", counts[0], counts[1], counts[2])
	tx(summary, summary, "regular", 10, 224, 634, 160, brass)
	return records, first
}
func renderPanel(in *Inputs, records []Draw, images map[string]image.Image, d int) (*image.RGBA, error) {
	out := image.NewRGBA(image.Rect(0, 0, 400*d, 640*d))
	for _, r := range records {
		switch r.Kind {
		case "image":
			im := images[r.Path]
			if im == nil {
				return nil, fmt.Errorf("missing image %s", r.Path)
			}
			if im.Bounds().Dx() != int(r.Width*float64(d)) || im.Bounds().Dy() != int(r.Height*float64(d)) {
				return nil, fmt.Errorf("same-density size mismatch %s", r.Path)
			}
			pt := image.Pt(int(r.X*float64(d)), int(r.Y*float64(d)))
			draw.Draw(out, image.Rectangle{Min: pt, Max: pt.Add(im.Bounds().Size())}, im, im.Bounds().Min, draw.Over)
		case "text":
			f, e := newFace(in.Fonts[r.Font], r.Size, d)
			if e != nil {
				return nil, e
			}
			dr := font.Drawer{Dst: out, Src: image.NewUniform(rgba(r.Color)), Face: f, Dot: fixed.Point26_6{X: fixed.Int26_6(math.Round(r.X * float64(d) * 64)), Y: fixed.Int26_6(math.Round(r.Baseline * float64(d) * 64))}}
			dr.DrawString(r.Text)
			_ = f.Close()
		}
	}
	return out, nil
}
