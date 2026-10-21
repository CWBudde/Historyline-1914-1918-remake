package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/image/font/opentype"
)

func screenTestInputs(t *testing.T) (string, []byte, []Loaded) {
	t.Helper()
	p := pipelineManifest("screens-v1")
	raw, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	_, loaded, e := screenInputs(p, raw)
	if e != nil {
		t.Fatal(e)
	}
	return p, raw, loaded
}

func TestHandoverNeverSamplesPreviousBackdrop(t *testing.T) {
	_, _, loaded := screenTestInputs(t)
	fonts := map[string]*opentype.Font{}
	for _, in := range loaded {
		if in.Spec.ID == "font/regular" || in.Spec.ID == "font/bold" {
			f, e := opentype.Parse(in.Bytes)
			if e != nil {
				t.Fatal(e)
			}
			fonts[in.Spec.ID] = f
		}
	}
	var handover Screen
	for _, s := range screenCatalog() {
		if s.ID == "handover" {
			handover = s
		}
	}
	if !handover.Privacy {
		t.Fatal("handover lost privacy flag")
	}
	red := image.NewNRGBA(image.Rect(0, 0, 1280, 800))
	blue := image.NewNRGBA(red.Bounds())
	for y := 0; y < 800; y++ {
		for x := 0; x < 1280; x++ {
			red.SetNRGBA(x, y, color.NRGBA{255, 0, 0, 255})
			blue.SetNRGBA(x, y, color.NRGBA{0, 0, 255, 255})
		}
	}
	a, e := screenPreview(handover, red, fonts)
	if e != nil {
		t.Fatal(e)
	}
	b, e := screenPreview(handover, blue, fonts)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(a.Pix, b.Pix) {
		t.Fatal("previous background leaked into handover")
	}
	for i := 3; i < len(a.Pix); i += 4 {
		if a.Pix[i] != 255 {
			t.Fatal("handover is not opaque")
		}
	}
	h := screenGallery([]Screen{handover})
	if strings.Contains(h, `src="backdrop/`) {
		t.Fatal("handover HTML loads backdrop")
	}
}

func TestScreenStrictInputsAndCrop(t *testing.T) {
	p, raw, loaded := screenTestInputs(t)
	var m ScreenManifest
	if e := json.Unmarshal(raw, &m); e != nil {
		t.Fatal(e)
	}
	bad := m
	bad.Fonts = append([]Input(nil), m.Fonts...)
	bad.Fonts[1] = bad.Fonts[0]
	if _, _, e := screenInputs(p, writeJSON(bad)); e == nil {
		t.Fatal("duplicate font accepted")
	}
	bad = m
	bad.Backdrop.SHA256 = strings.Repeat("0", 64)
	if _, _, e := screenInputs(p, writeJSON(bad)); e == nil {
		t.Fatal("tampered backdrop accepted")
	}
	bad = m
	bad.Backdrop.Source = "legacy/python/gfx/gameIcon.png"
	if _, _, e := screenInputs(p, writeJSON(bad)); e == nil {
		t.Fatal("reference source accepted")
	}
	if _, _, e := screenInputs(p, append(raw, []byte("{}")...)); e == nil {
		t.Fatal("trailing manifest accepted")
	}
	unknown := bytes.Replace(raw, []byte(`"profile"`), []byte(`"unknown":0,"profile"`), 1)
	if _, _, e := screenInputs(p, unknown); e == nil {
		t.Fatal("unknown field accepted")
	}
	for _, c := range [][4]int{{0, 0, 640, 400}, {0, 0, 1280, 801}, {-1, 0, 1504, 940}, {200, 0, 1504, 940}} {
		in := loaded[0]
		in.Spec.Crop = c
		if _, e := screenBackdrop(in); e == nil {
			t.Fatalf("invalid crop accepted: %v", c)
		}
	}
}

// A native-resolution export must retain details that a 2x image cannot hold;
// a larger canvas alone must not permit upscaling a smaller selected crop.
func TestScreenBackdropNative4xRequiresSufficientCrop(t *testing.T) {
	_, raw, loaded := screenTestInputs(t)
	var manifest ScreenManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for i, in := range loaded[1:] {
		name := filepath.Base(in.Spec.Source)
		if err := os.WriteFile(filepath.Join(dir, name), in.Bytes, 0644); err != nil {
			t.Fatal(err)
		}
		manifest.Fonts[i].Source = name
	}
	source := image.NewNRGBA(image.Rect(0, 0, 2560, 1600))
	for y := 0; y < 1600; y++ {
		for x := 0; x < 2560; x++ {
			source.SetNRGBA(x, y, color.NRGBA{R: uint8((x % 2) * 255), G: uint8((y % 2) * 255), B: 75, A: 255})
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, source); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "backdrop.png"), encoded.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	manifest.Backdrop.Source = "backdrop.png"
	manifest.Backdrop.SHA256 = hash(encoded.Bytes())
	for _, tc := range []struct {
		name string
		crop [4]int
		four bool
	}{
		{"native", [4]int{0, 0, 2560, 1600}, true},
		{"small-crop-large-canvas", [4]int{0, 0, 1280, 800}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			manifest.Backdrop.Crop = tc.crop
			path := filepath.Join(dir, tc.name+".json")
			if err := os.WriteFile(path, writeJSON(manifest), 0644); err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(dir, tc.name)
			if err := export(path, out); err != nil {
				t.Fatal(err)
			}
			four, err := os.ReadFile(filepath.Join(out, "backdrop", "western-front-4x.png"))
			if !tc.four {
				if !os.IsNotExist(err) {
					t.Fatalf("small crop must not produce a 4x export: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("native 4x backdrop missing: %v", err)
			}
			im, err := png.Decode(bytes.NewReader(four))
			if err != nil {
				t.Fatal(err)
			}
			if im.Bounds() != image.Rect(0, 0, 2560, 1600) {
				t.Fatalf("native 4x dimensions: %v", im.Bounds())
			}
			for _, p := range []image.Point{{0, 0}, {1, 0}, {0, 1}, {123, 456}, {2559, 1599}} {
				if got := color.NRGBAModel.Convert(im.At(p.X, p.Y)); got != source.NRGBAAt(p.X, p.Y) {
					t.Fatalf("native detail changed at %v: %v", p, got)
				}
			}
		})
	}
}

func TestRetainedScreensRepeatAndGallery(t *testing.T) {
	root := freshExport(t, "screens-v1")
	raw, e := os.ReadFile(filepath.Join(root, "report.json"))
	if e != nil {
		t.Fatal(e)
	}
	var report Report
	if e = json.Unmarshal(raw, &report); e != nil {
		t.Fatal(e)
	}
	frames, controls, backdrops, previews, svgs, flags := 0, 0, 0, 0, 0, 0
	for _, f := range report.Files {
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(f.Path)))
		if err != nil {
			t.Fatal(err)
		}
		if hash(b) != f.SHA256 {
			t.Fatalf("saved pin mismatch: %s", f.Path)
		}
		if strings.HasSuffix(f.Path, ".svg") {
			svgs++
			if bytes.Contains(b, []byte("<text")) {
				t.Fatal("baked labels in runtime art")
			}
		}
		if f.Density == 0 {
			continue
		}
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		if im.Bounds().Dx() != f.Width || im.Bounds().Dy() != f.Height {
			t.Fatalf("saved dimensions: %s", f.Path)
		}
		switch f.Family {
		case "screen-frame":
			frames++
			if f.Width != 640*f.Density || f.Height != 400*f.Density {
				t.Fatal("frame density")
			}
		case "controls":
			controls++
		case "backdrop":
			backdrops++
			if f.Density > 2 {
				t.Fatal("upscaled backdrop")
			}
		case "review":
			previews++
			if f.Density != 2 {
				t.Fatal("preview density")
			}
		case "information":
			flags++
		}
	}
	if frames != 48 || controls != 15 || backdrops != 2 || previews != 16 || svgs != 23 || flags != 6 || len(report.Files) != 118 {
		t.Fatalf("coverage: %d frames %d controls %d backdrops %d previews %d SVG %d flags %d files", frames, controls, backdrops, previews, svgs, flags, len(report.Files))
	}
	// The sixteen screen frames in assets/ are this export, byte for byte.
	if n := checkRuntimePins(t, "screens", report.Files, func(f Output) bool {
		return f.Family == "screen-frame" && f.Density > 0 && strings.HasSuffix(f.Path, ".png")
	}); n != 48 {
		t.Fatalf("compared %d runtime frames, want 48", n)
	}
	h, e := os.ReadFile(filepath.Join(root, "index.html"))
	if e != nil {
		t.Fatal(e)
	}
	for _, m := range regexp.MustCompile(`(?:href|src)="([^"]+)"|url\('([^']+)'\)`).FindAllStringSubmatch(string(h), -1) {
		p := m[1]
		if p == "" {
			p = m[2]
		}
		if strings.HasPrefix(p, "#") {
			continue
		}
		if _, e = os.Stat(filepath.Join(root, filepath.FromSlash(p))); e != nil {
			t.Fatalf("broken gallery link %s", p)
		}
	}
	seen := map[string]bool{}
	for _, s := range screenCatalog() {
		if seen[s.ID] {
			t.Fatal("duplicate view")
		}
		seen[s.ID] = true
	}
	for _, id := range []string{"title", "side-selection", "mission-selection", "briefing", "campaign", "options", "pause", "save", "load", "handover", "victory", "defeat", "credits", "help", "editor", "new-map"} {
		if !seen[id] {
			t.Fatalf("missing separate view: %s", id)
		}
	}
	p, _, _ := screenTestInputs(t)
	out := filepath.Join(t.TempDir(), "repeat")
	if e = export(p, out); e != nil {
		t.Fatal(e)
	}
	for _, f := range append(report.Files, Output{Path: "report.json"}) {
		a, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(f.Path)))
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(f.Path)))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a, b) {
			t.Fatalf("non-reproducible %s", f.Path)
		}
	}
	if e = export(p, out); e == nil {
		t.Fatal("existing output overwritten")
	}
	b, e := os.ReadFile(filepath.Join(out, "report.json"))
	if e != nil || !bytes.Equal(b, raw) {
		t.Fatal("existing report changed")
	}
}
