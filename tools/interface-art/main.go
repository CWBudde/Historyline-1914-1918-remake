// Command interface-art builds offline interface, weapon and effect candidates.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Input struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	SHA256 string `json:"sha256"`
	Crop   [4]int `json:"crop"`
}
type Manifest struct {
	SchemaVersion int     `json:"schema_version"`
	Profile       string  `json:"profile"`
	Weapons       []Input `json:"weapons"`
	Explosion     Input   `json:"explosion"`
	Fonts         []Input `json:"fonts"`
}
type Output struct {
	ID      string     `json:"id"`
	Family  string     `json:"family"`
	Path    string     `json:"path"`
	SHA256  string     `json:"sha256"`
	Width   int        `json:"width,omitempty"`
	Height  int        `json:"height,omitempty"`
	Density int        `json:"density,omitempty"`
	Anchor  [2]float64 `json:"logical_anchor"`
}
type Report struct {
	SchemaVersion  int      `json:"schema_version"`
	Profile        string   `json:"profile"`
	Acceptance     string   `json:"acceptance"`
	Filter         string   `json:"filter"`
	ManifestSHA256 string   `json:"manifest_sha256"`
	Files          []Output `json:"files"`
}
type Loaded struct {
	Spec  Input
	Bytes []byte
	Image *image.NRGBA
}

func hash(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func decodeStrict(b []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		return e
	}
	var extra any
	if e := d.Decode(&extra); e != io.EOF {
		return errors.New("manifest has trailing content")
	}
	return nil
}
func loadInput(root *os.Root, s Input, raster bool) (Loaded, error) {
	in := Loaded{Spec: s}
	p := filepath.ToSlash(s.Source)
	if p == "" || strings.Contains(p, ":") || strings.HasPrefix(p, "/") || strings.Contains(p, "\\") {
		return in, errors.New("source must be a contained slash-separated path")
	}
	for _, part := range strings.Split(p, "/") {
		if part == ".." || strings.EqualFold(part, "original") || strings.EqualFold(part, "legacy") {
			return in, errors.New("reference or escaping source rejected")
		}
	}
	b, e := root.ReadFile(filepath.FromSlash(p))
	if e != nil {
		return in, e
	}
	if len(s.SHA256) != 64 || hash(b) != s.SHA256 {
		return in, fmt.Errorf("%s: source pin mismatch", s.ID)
	}
	in.Bytes = b
	if !raster {
		return in, nil
	}
	im, e := png.Decode(bytes.NewReader(b))
	if e != nil {
		return in, e
	}
	n := image.NewNRGBA(im.Bounds())
	for y := n.Bounds().Min.Y; y < n.Bounds().Max.Y; y++ {
		for x := n.Bounds().Min.X; x < n.Bounds().Max.X; x++ {
			n.Set(x, y, im.At(x, y))
		}
	}
	in.Image = n
	return in, nil
}
func crop(im *image.NRGBA, c [4]int, w, h int) (*image.NRGBA, error) {
	r := image.Rect(c[0], c[1], c[0]+c[2], c[1]+c[3])
	if c[2] < w*4 || c[3] < h*4 || c[2]*h != c[3]*w || !r.In(im.Bounds()) {
		return nil, errors.New("crop must be contained, exact-aspect and at least 4x; no upscaling")
	}
	return im.SubImage(r).(*image.NRGBA), nil
}

func writeJSON(v any) []byte {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		panic(e)
	}
	return append(b, '\n')
}

type Builder struct {
	Dir    string
	Report Report
}

func (b *Builder) save(path string, data []byte, o Output) error {
	full := filepath.Join(b.Dir, filepath.FromSlash(path))
	if e := os.MkdirAll(filepath.Dir(full), 0755); e != nil {
		return e
	}
	if e := os.WriteFile(full, data, 0644); e != nil {
		return e
	}
	o.Path = path
	o.SHA256 = hash(data)
	b.Report.Files = append(b.Report.Files, o)
	return nil
}
func (b *Builder) png(path string, im image.Image, o Output) error {
	var buf bytes.Buffer
	if e := png.Encode(&buf, im); e != nil {
		return e
	}
	o.Width, o.Height = im.Bounds().Dx(), im.Bounds().Dy()
	return b.save(path, buf.Bytes(), o)
}
func (b *Builder) densities(id, family string, anchor [2]float64, master *image.NRGBA) error {
	im := master
	for _, d := range []int{4, 2, 1} {
		if d != 4 {
			var e error
			im, e = resizeLanczos(im, master.Bounds().Dx()*d/4, master.Bounds().Dy()*d/4)
			if e != nil {
				return e
			}
		}
		o := Output{ID: id, Family: family, Density: d, Anchor: anchor}
		if e := b.png(fmt.Sprintf("%s-%dx.png", id, d), im, o); e != nil {
			return e
		}
	}
	return nil
}
func export(manifestPath, out string) error {
	raw, e := os.ReadFile(manifestPath)
	if e != nil {
		return e
	}
	var m Manifest
	if e = decodeStrict(raw, &m); e != nil {
		return e
	}
	if m.SchemaVersion != 1 || m.Profile != "interface-v1" {
		return errors.New("requires schema 1 interface-v1")
	}
	root, e := os.OpenRoot(filepath.Dir(manifestPath))
	if e != nil {
		return e
	}
	defer func() { _ = root.Close() }()
	if e = validateBindings(m); e != nil {
		return e
	}
	var loaded []Loaded
	for _, s := range append(append(append([]Input{}, m.Weapons...), m.Explosion), m.Fonts...) {
		in, err := loadInput(root, s, !strings.HasPrefix(s.ID, "font/"))
		if err != nil {
			return err
		}
		loaded = append(loaded, in)
	}
	// Resolve parent once. Only remove the exclusively-created staging directory under it.
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
	stage, e := os.MkdirTemp(parent, ".interface-art-")
	if e != nil {
		return e
	}
	defer func() {
		if filepath.Dir(stage) == parent && strings.HasPrefix(filepath.Base(stage), ".interface-art-") {
			_ = os.RemoveAll(stage)
		}
	}()
	b := Builder{Dir: stage, Report: Report{SchemaVersion: 1, Profile: m.Profile, ManifestSHA256: hash(raw), Acceptance: "offline candidates; reference likeness, native mapping, responsive flows and runtime acceptance pending", Filter: "native 4x vector raster or source crop -> 4x -> 2x -> 1x; signed premultiplied byte-RGB Lanczos3, no upscaling"}}
	if e = b.save("manifest.json", raw, Output{ID: "manifest", Family: "metadata"}); e != nil {
		return e
	}
	arts := catalog()
	if e = b.save("vectors.json", writeJSON(arts), Output{ID: "vectors", Family: "source"}); e != nil {
		return e
	}
	for _, a := range arts {
		if e = b.save(a.ID+".svg", a.svg(), Output{ID: a.ID, Family: a.Family, Width: a.Width, Height: a.Height, Anchor: a.Anchor}); e != nil {
			return e
		}
		if e = b.densities(a.ID, a.Family, a.Anchor, a.raster(4)); e != nil {
			return e
		}
	}
	for _, in := range loaded {
		ext := ".png"
		if strings.HasPrefix(in.Spec.ID, "font/") {
			ext = filepath.Ext(in.Spec.Source)
		}
		if e = b.save("inputs/"+in.Spec.ID+ext, in.Bytes, Output{ID: in.Spec.ID, Family: "source"}); e != nil {
			return e
		}
		if in.Spec.ID == "effect/explosion-atlas" {
			// Explicit equal cells and optional gutters avoid fractional grid crops.
			cell := in.Spec.Crop[2]
			gap := in.Spec.Crop[3]
			if cell < 384 || gap < 0 || in.Image.Bounds().Dx() != 3*cell+2*gap || in.Image.Bounds().Dy() != 3*cell+2*gap {
				return errors.New("explosion grid dimensions do not match cell/gap")
			}
			for i := 0; i < 9; i++ {
				x, y := (i%3)*(cell+gap), (i/3)*(cell+gap)
				im, err := crop(in.Image, [4]int{x, y, cell, cell}, 96, 96)
				if err != nil {
					return err
				}
				im, err = resizeLanczos(im, 384, 384)
				if err != nil {
					return err
				}
				if e = b.densities(fmt.Sprintf("effect/explosion-%02d", i), "effects", [2]float64{48, 48}, im); e != nil {
					return e
				}
			}
			continue
		}
		if strings.HasPrefix(in.Spec.ID, "weapon/") {
			im, err := fitWeapon(in.Image, in.Spec.Crop)
			if err != nil {
				return err
			}
			if e = b.densities(in.Spec.ID, "weapons", [2]float64{126, 25}, im); e != nil {
				return e
			}
		}
	}
	if e = b.save("animations.json", writeJSON(animations()), Output{ID: "animations", Family: "metadata"}); e != nil {
		return e
	}
	if e = b.save("bindings.json", writeJSON(bindings()), Output{ID: "bindings", Family: "metadata"}); e != nil {
		return e
	}
	if e = b.save("index.html", []byte(gallery(b.Report.Files)), Output{ID: "gallery", Family: "review"}); e != nil {
		return e
	}
	sort.Slice(b.Report.Files, func(i, j int) bool { return b.Report.Files[i].Path < b.Report.Files[j].Path })
	if e = os.WriteFile(filepath.Join(stage, "report.json"), writeJSON(b.Report), 0644); e != nil {
		return e
	}
	return os.Rename(stage, target)
}

func fitWeapon(im *image.NRGBA, c [4]int) (*image.NRGBA, error) {
	r := image.Rect(c[0], c[1], c[0]+c[2], c[1]+c[3])
	if c[2] <= 0 || c[3] <= 0 || !r.In(im.Bounds()) {
		return nil, errors.New("weapon crop outside source")
	}
	if e := guardCrop(im, c); e != nil {
		return nil, e
	}
	// Fit without stretching into a 252x50 logical row. Small pistols retain their
	// naturally smaller width; all artwork is retained, with transparent letterboxing.
	scale := min(1000.0/float64(c[2]), 192.0/float64(c[3]))
	if scale > 1 {
		return nil, errors.New("weapon source too small; no upscaling")
	}
	w, h := int(float64(c[2])*scale), int(float64(c[3])*scale)
	fitted, e := resizeLanczos(im.SubImage(r).(*image.NRGBA), w, h)
	if e != nil {
		return nil, e
	}
	out := image.NewNRGBA(image.Rect(0, 0, 1008, 200))
	pt := image.Pt((1008-w)/2, (200-h)/2)
	draw.Draw(out, fitted.Bounds().Add(pt), fitted, image.Point{}, draw.Src)
	return out, nil
}

func guardCrop(im *image.NRGBA, c [4]int) error {
	// No silently amputated painted parts: all nontrivial alpha must be inside the crop.
	r := image.Rect(c[0]+2, c[1]+2, c[0]+c[2]-2, c[1]+c[3]-2)
	for y := im.Bounds().Min.Y; y < im.Bounds().Max.Y; y++ {
		for x := im.Bounds().Min.X; x < im.Bounds().Max.X; x++ {
			if im.NRGBAAt(x, y).A > 8 && !image.Pt(x, y).In(r) {
				return errors.New("visible artwork outside safe crop")
			}
		}
	}
	return nil
}
func run(args []string) error {
	fs := flag.NewFlagSet("interface-art", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	manifest := fs.String("manifest", "", "pinned authored inputs")
	out := fs.String("out", "", "fresh output directory")
	seen := map[string]bool{}
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			key := strings.SplitN(strings.TrimLeft(a, "-"), "=", 2)[0]
			if seen[key] {
				return fmt.Errorf("repeated flag %s", key)
			}
			seen[key] = true
		}
	}
	if e := fs.Parse(args); e != nil {
		return e
	}
	if fs.NArg() != 0 || *manifest == "" || *out == "" {
		return errors.New("usage: interface-art -manifest FILE -out FRESH_DIR")
	}
	return export(*manifest, *out)
}
func main() {
	if e := run(os.Args[1:]); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
