package main

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"image"
	"image/draw"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type ScreenManifest struct {
	SchemaVersion int     `json:"schema_version"`
	Profile       string  `json:"profile"`
	Backdrop      Input   `json:"backdrop"`
	Fonts         []Input `json:"fonts"`
}

func screenInputs(manifestPath string, raw []byte) (ScreenManifest, []Loaded, error) {
	var m ScreenManifest
	if e := decodeStrict(raw, &m); e != nil {
		return m, nil, e
	}
	if m.SchemaVersion != 1 || (m.Profile != "screens-v1" && m.Profile != "screens-v2") || m.Backdrop.ID != "backdrop/western-front" {
		return m, nil, errors.New("requires schema 1 screens-v1/screens-v2 and western-front backdrop")
	}
	if len(m.Fonts) != 3 {
		return m, nil, errors.New("requires regular/bold/licence font inputs")
	}
	seen := map[string]bool{}
	for _, f := range m.Fonts {
		if (f.ID != "font/regular" && f.ID != "font/bold" && f.ID != "font/license") || seen[f.ID] {
			return m, nil, errors.New("unknown or duplicate screen font")
		}
		seen[f.ID] = true
	}
	root, e := os.OpenRoot(sourceBase(manifestPath))
	if e != nil {
		return m, nil, e
	}
	defer func() { _ = root.Close() }()
	var inputs []Loaded
	for _, s := range append([]Input{m.Backdrop}, m.Fonts...) {
		in, err := loadInput(root, s, s.ID == m.Backdrop.ID)
		if err != nil {
			return m, nil, err
		}
		inputs = append(inputs, in)
	}
	return m, inputs, nil
}

func screenBackdrop(in Loaded) (*image.NRGBA, error) {
	c := in.Spec.Crop
	r := image.Rect(c[0], c[1], c[0]+c[2], c[1]+c[3])
	if c[2] < 1280 || c[3] < 800 || c[2]*5 != c[3]*8 || !r.In(in.Image.Bounds()) {
		return nil, errors.New("backdrop requires a contained 8:5 crop at least 1280x800; no upscaling")
	}
	return resizeLanczos(in.Image.SubImage(r).(*image.NRGBA), 1280, 800)
}

func exportScreens(manifestPath, out string, raw []byte) error {
	m, inputs, e := screenInputs(manifestPath, raw)
	if e != nil {
		return e
	}
	backdrop, e := screenBackdrop(inputs[0])
	if e != nil {
		return e
	}
	fonts := map[string]*opentype.Font{}
	for _, in := range inputs {
		if in.Spec.ID == "font/regular" || in.Spec.ID == "font/bold" {
			f, err := opentype.Parse(in.Bytes)
			if err != nil {
				return err
			}
			fonts[in.Spec.ID] = f
		}
	}
	abs, e := filepath.Abs(out)
	if e != nil {
		return e
	}
	parent, e := filepath.EvalSymlinks(filepath.Dir(abs))
	if e != nil {
		return e
	}
	target := filepath.Join(parent, filepath.Base(abs))
	if _, e = os.Lstat(target); !os.IsNotExist(e) {
		return errors.New("output must be a fresh directory")
	}
	stage, e := os.MkdirTemp(parent, ".screen-art-")
	if e != nil {
		return e
	}
	defer func() {
		if filepath.Dir(stage) == parent && strings.HasPrefix(filepath.Base(stage), ".screen-art-") {
			_ = os.RemoveAll(stage)
		}
	}()
	b := Builder{Dir: stage, Report: Report{SchemaVersion: 1, Profile: m.Profile, ManifestSHA256: hash(raw), Acceptance: "offline artwork and layout fixtures; gameplay flows, original correspondence and runtime acceptance pending", Filter: "native 4x vectors -> 2x -> 1x signed premultiplied byte-RGB Lanczos3; painted backdrop 2x/1x only, no upscaling"}}
	if e = b.save("manifest.json", raw, Output{ID: "manifest", Family: "metadata"}); e != nil {
		return e
	}
	for _, in := range inputs {
		ext := filepath.Ext(in.Spec.Source)
		if e = b.save("inputs/"+in.Spec.ID+ext, in.Bytes, Output{ID: in.Spec.ID, Family: "source"}); e != nil {
			return e
		}
	}
	if e = b.png("backdrop/western-front-2x.png", backdrop, Output{ID: m.Backdrop.ID, Family: "backdrop", Density: 2}); e != nil {
		return e
	}
	one, e := resizeLanczos(backdrop, 640, 400)
	if e != nil {
		return e
	}
	if e = b.png("backdrop/western-front-1x.png", one, Output{ID: m.Backdrop.ID, Family: "backdrop", Density: 1}); e != nil {
		return e
	}
	screens := screenCatalog()
	if m.Profile == "screens-v2" {
		screens = screenStateCatalog()
	}
	if e = b.save("screens.json", writeJSON(screens), Output{ID: "screens", Family: "metadata"}); e != nil {
		return e
	}
	var arts []Art
	for _, s := range screens {
		arts = append(arts, screenFrame(s))
	}
	for _, state := range []string{"normal", "hover", "focus", "pressed", "disabled"} {
		arts = append(arts, Art{ID: "button/" + state, Family: "controls", Width: 160, Height: 40, Shapes: controlFrame(160, 40, state)})
	}
	for _, a := range catalog() {
		if strings.HasPrefix(a.ID, "country/") {
			arts = append(arts, a)
		}
	}
	if m.Profile == "screens-v2" {
		arts = append(arts, widgetCatalog()...)
		arts = append(arts, screenStateComponents()...)
		if e = b.save("widget-geometry.json", writeJSON(widgetGeometry()), Output{ID: "widget-geometry", Family: "metadata"}); e != nil {
			return e
		}
	}
	if e = b.save("vectors.json", writeJSON(arts), Output{ID: "vectors", Family: "source"}); e != nil {
		return e
	}
	for _, a := range arts {
		if e = b.save(a.ID+".svg", a.svg(), Output{ID: a.ID, Family: a.Family, Width: a.Width, Height: a.Height}); e != nil {
			return e
		}
		if e = b.densities(a.ID, a.Family, a.Anchor, a.raster(4)); e != nil {
			return e
		}
	}
	for _, s := range screens {
		im, err := screenPreview(s, backdrop, fonts)
		if err != nil {
			return fmt.Errorf("%s: %w", s.ID, err)
		}
		if e = b.png("fixtures/"+s.ID+"-2x.png", im, Output{ID: "fixture/" + s.ID, Family: "review", Density: 2}); e != nil {
			return e
		}
	}
	gallery := screenGallery(screens)
	if m.Profile == "screens-v2" {
		gallery = screenStateGallery(screens, arts)
	}
	if e = b.save("index.html", []byte(gallery), Output{ID: "gallery", Family: "review"}); e != nil {
		return e
	}
	sort.Slice(b.Report.Files, func(i, j int) bool { return b.Report.Files[i].Path < b.Report.Files[j].Path })
	if e = os.WriteFile(filepath.Join(stage, "report.json"), writeJSON(b.Report), 0644); e != nil {
		return e
	}
	return os.Rename(stage, target)
}

func screenPreview(s Screen, backdrop *image.NRGBA, fonts map[string]*opentype.Font) (*image.NRGBA, error) {
	im := image.NewNRGBA(image.Rect(0, 0, 1280, 800))
	// Privacy screens never sample the backdrop (nor any previous gameplay image).
	if !s.Privacy {
		draw.Draw(im, im.Bounds(), backdrop, image.Point{}, draw.Src)
	}
	a := screenFrame(s)
	ordered := false
	for _, d := range s.Decorations {
		if d.Clip != nil {
			ordered = true
		}
	}
	for _, d := range s.Decorations {
		if ordered {
			continue
		}
		for _, source := range screenDecorationSources() {
			if source.ID == d.ArtID {
				r := d.Rect
				if d.Clip == nil {
					a.Shapes = append(a.Shapes, transform(source.Shapes, float64(r[2])/float64(source.Width), float64(r[3])/float64(source.Height), float64(r[0]), float64(r[1]), 0)...)
				}
			}
		}
	}
	for _, c := range s.Controls {
		a.Shapes = append(a.Shapes, screenControlShapes(c)...)
	}
	draw.Draw(im, im.Bounds(), a.raster(2), image.Point{}, draw.Over)
	for _, d := range s.Decorations {
		if ordered {
			if e := renderClippedDecoration(im, d); e != nil {
				return nil, e
			}
		}
	}
	text := append([]ScreenText(nil), s.Text...)
	for _, c := range s.Controls {
		col := cream
		if c.State == "disabled" {
			col = muted
		}
		text = append(text, ScreenText{c.Label, c.Rect[0] + 12, c.Rect[1] + 20, 13, false, col})
	}
	for _, t := range text {
		key := "font/regular"
		if t.Bold {
			key = "font/bold"
		}
		face, e := opentype.NewFace(fonts[key], &opentype.FaceOptions{Size: float64(t.Size * 2), DPI: 72, Hinting: font.HintingNone})
		if e != nil {
			return nil, e
		}
		bounds, _ := font.BoundString(face, t.Text)
		ink := bounds.Add(fixed.P(t.X*2, t.Y*2))
		if ink.Min.X < 0 || ink.Min.Y < 0 || ink.Max.X > fixed.I(1280) || ink.Max.Y > fixed.I(800) {
			_ = face.Close()
			return nil, fmt.Errorf("text outside screen: %q", t.Text)
		}
		d := font.Drawer{Dst: im, Src: image.NewUniform(colour(t.Color)), Face: face, Dot: fixed.P(t.X*2, t.Y*2)}
		d.DrawString(t.Text)
		_ = face.Close()
	}
	if s.Scroll != nil {
		if e := renderScreenScroll(im, *s.Scroll, fonts["font/regular"]); e != nil {
			return nil, e
		}
	}
	return im, nil
}

func screenGallery(screens []Screen) string {
	var b bytes.Buffer
	b.WriteString(`<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Full-screen artwork candidates</title>
<style>@font-face{font-family:Go;src:url('inputs/font/regular.ttf')}@font-face{font-family:Go;src:url('inputs/font/bold.ttf');font-weight:700}
*{box-sizing:border-box}body{margin:24px;background:#101b22;color:#eee5d2;font:16px Go,sans-serif}a{color:#c4a773}p{max-width:960px;line-height:1.5}nav{display:flex;gap:16px;flex-wrap:wrap}.scroll{overflow:auto;padding:4px}.canvas{position:relative;width:640px;height:400px;isolation:isolate}.layer{position:absolute;inset:0;width:640px;height:400px;pointer-events:none}.live{position:absolute;white-space:pre;margin:0;line-height:1}.canvas button{position:absolute;border:0;padding:0 12px;text-align:left;font:13px Go;color:#eee5d2;background:transparent url('button/normal-2x.png') center/100% 100% no-repeat;cursor:pointer}.canvas button:hover{background-image:url('button/hover-2x.png')}.canvas button:focus-visible{background-image:url('button/focus-2x.png');outline:2px solid #eee5d2;outline-offset:2px}.canvas button:active,.canvas button[aria-pressed=true]{background-image:url('button/pressed-2x.png')}.canvas button:disabled{background-image:url('button/disabled-2x.png');color:#6e8088;cursor:default}.sample-focus{box-shadow:0 0 0 1px #c4a773}article{margin:32px 0}summary{cursor:pointer}img.preview{display:block;width:640px;height:400px}code{color:#c4a773}</style>
<h1>Full-screen artwork candidates</h1><p>Sixteen separate views in the existing slate/brass vocabulary. All labels and example values below are live HTML. Controls demonstrate artwork states; they perform no gameplay, saving, campaign unlocking or editor operations. These are fixed 640×400 layout studies. Native font composition, responsive layouts, keyboard flows and runtime acceptance require review.</p><p>The painted backdrop is retained at 1×/2× only: its source cannot support a true 4× export. Text-free vector frames and five shared button states have 1×/2×/4× PNGs and editable SVGs. Saved 2× compositions are review fixtures only, never runtime artwork. Handover uses an opaque frame and loads no backdrop.</p><p><a href="manifest.json">Pinned manifest</a> · <a href="screens.json">Layouts and live fixtures</a> · <a href="vectors.json">Vector source</a> · <a href="report.json">Output pins</a> · <a href="inputs/font/license.txt">Font licence</a></p><nav>`)
	for _, s := range screens {
		fmt.Fprintf(&b, `<a href="#%s">%s</a>`, s.ID, html.EscapeString(s.Title))
	}
	b.WriteString("</nav>")
	for _, s := range screens {
		fmt.Fprintf(&b, `<article id="%s"><h2>%s</h2><p><code>%s</code> · <a href="screen/%s.svg">Editable frame</a> · PNG <a href="screen/%s-1x.png">1×</a> / <a href="screen/%s-2x.png">2×</a> / <a href="screen/%s-4x.png">4×</a></p><div class="scroll"><div class="canvas" aria-label="%s artwork fixture">`, s.ID, html.EscapeString(s.Title), s.InventoryID, s.ID, s.ID, s.ID, s.ID, html.EscapeString(s.Title))
		if !s.Privacy {
			b.WriteString(`<img class="layer" src="backdrop/western-front-2x.png" alt="">`)
		}
		fmt.Fprintf(&b, `<img class="layer" src="screen/%s-2x.png" alt="">`, s.ID)
		for _, d := range s.Decorations {
			r := d.Rect
			clip := ""
			if d.Clip != nil {
				c := d.Clip
				clip = fmt.Sprintf(";clip-path:inset(%dpx %dpx %dpx %dpx)", c[1], r[2]-c[0]-c[2], r[3]-c[1]-c[3], c[0])
			}
			fmt.Fprintf(&b, `<img class="live" src="%s-2x.png" alt="%s" style="left:%dpx;top:%dpx;width:%dpx;height:%dpx%s">`, d.ArtID, html.EscapeString(d.ArtID), r[0], r[1], r[2], r[3], clip)
		}
		for _, t := range s.Text {
			weight := 400
			if t.Bold {
				weight = 700
			}
			fmt.Fprintf(&b, `<span class="live" style="left:%dpx;top:%.2fpx;font-size:%dpx;font-weight:%d;color:%s">%s</span>`, t.X, float64(t.Y)-float64(t.Size)*.79, t.Size, weight, t.Color, html.EscapeString(t.Text))
		}
		for _, c := range s.Controls {
			attrs := ""
			if c.State == "disabled" {
				attrs = " disabled"
			}
			if c.State == "pressed" {
				attrs = ` aria-pressed="true"`
			}
			if c.State == "focus" {
				attrs += ` class="sample-focus"`
			}
			fmt.Fprintf(&b, `<button type="button"%s style="left:%dpx;top:%dpx;width:%dpx;height:%dpx" title="Artwork fixture; no game operation">%s</button>`, attrs, c.Rect[0], c.Rect[1], c.Rect[2], c.Rect[3], html.EscapeString(c.Label))
		}
		fmt.Fprintf(&b, `</div></div><details><summary>Saved native 2× font composition (review only)</summary><div class="scroll"><img class="preview" src="fixtures/%s-2x.png" alt="Saved %s composition"></div></details></article>`, s.ID, html.EscapeString(s.Title))
	}
	b.WriteString("</html>\n")
	return b.String()
}
