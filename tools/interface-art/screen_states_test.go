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
	"unicode"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func screenStateFonts(t *testing.T) map[string]*opentype.Font {
	t.Helper()
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
	return fonts
}

func TestScreenStateWrappingAndClipContainment(t *testing.T) {
	fonts := screenStateFonts(t)
	face, e := opentype.NewFace(fonts["font/regular"], &opentype.FaceOptions{Size: 26, DPI: 72, Hinting: font.HintingNone})
	if e != nil {
		t.Fatal(e)
	}
	defer func() { _ = face.Close() }()
	strip := func(s string) string {
		return strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, s)
	}
	for _, text := range []string{"i " + strings.Repeat("W", 100), "München — Épernay " + strings.Repeat("Üé", 70), "Line one\n\nLine two"} {
		for _, width := range []int{40, 70, 150, 536} {
			lines := wrapScreenText(face, text, width)
			for _, line := range lines {
				if font.MeasureString(face, line) > fixed.I(width) {
					t.Fatalf("wrap overflow %d: %q", width, line)
				}
			}
			if strip(strings.Join(lines, "")) != strip(text) {
				t.Fatal("wrapping lost source characters")
			}
		}
	}
	for _, single := range []bool{false, true} {
		im := image.NewNRGBA(image.Rect(0, 0, 1280, 800))
		for i := range im.Pix {
			im.Pix[i] = 27
		}
		before := bytes.Clone(im.Pix)
		s := ScreenScroll{Rect: [4]int{44, 104, 216, 24}, Text: strings.Repeat("Long name — München / Épernay ", 30), LineHeight: 24, SingleLine: single}
		if e = renderScreenScroll(im, s, fonts["font/regular"]); e != nil {
			t.Fatal(e)
		}
		r := image.Rect(88, 208, 520, 256)
		changed := false
		for y := 0; y < 800; y++ {
			for x := 0; x < 1280; x++ {
				i := im.PixOffset(x, y)
				same := bytes.Equal(im.Pix[i:i+4], before[i:i+4])
				if !image.Pt(x, y).In(r) && !same {
					t.Fatal("text escaped clip")
				}
				if !same {
					changed = true
				}
			}
		}
		if !changed {
			t.Fatal("empty scroll composition")
		}
	}
}

func TestScreenSliderFillClipsWithoutScalingOrCoveringThumb(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 1280, 800))
	d := ScreenDecoration{ArtID: "widget/slider/fill/normal", Rect: [4]int{324, 200, 220, 32}, Clip: &[4]int{0, 0, 110, 32}}
	if e := renderClippedDecoration(im, d); e != nil {
		t.Fatal(e)
	}
	if im.NRGBAAt(330*2, 216*2).A != 0 || im.NRGBAAt(337*2, 216*2).A == 0 || im.NRGBAAt(435*2, 216*2).A != 0 {
		t.Fatal("fill scaled instead of clipped")
	}
	var s Screen
	for _, state := range screenStateCatalog() {
		if state.ID == "settings-controls" {
			s = state
		}
	}
	preview, e := screenPreview(s, image.NewNRGBA(im.Bounds()), screenStateFonts(t))
	if e != nil {
		t.Fatal(e)
	}
	var thumb Art
	for _, a := range widgetCatalog() {
		if a.ID == "widget/slider/thumb/normal" {
			thumb = a
		}
	}
	// An opaque point in the left side of the thumb must remain above the fill.
	want := thumb.raster(2).NRGBAAt(8*2, 16*2)
	if got := preview.NRGBAAt(430*2, 216*2); got != want {
		t.Fatalf("fill covered thumb: %v vs %v", got, want)
	}
}

func TestRetainedScreenStatesRepeatPinsAndLinks(t *testing.T) {
	root := freshExport(t, "screens-v2")
	raw, e := os.ReadFile(filepath.Join(root, "report.json"))
	if e != nil {
		t.Fatal(e)
	}
	var r Report
	if e = json.Unmarshal(raw, &r); e != nil {
		t.Fatal(e)
	}
	counts := map[string]int{}
	ids := map[string]map[int]bool{}
	for _, f := range r.Files {
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(f.Path)))
		if err != nil {
			t.Fatal(err)
		}
		if hash(b) != f.SHA256 {
			t.Fatalf("saved pin: %s", f.Path)
		}
		if f.Density == 0 {
			continue
		}
		counts[f.Family]++
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		if im.Bounds().Dx() != f.Width || im.Bounds().Dy() != f.Height {
			t.Fatalf("saved dimensions: %s", f.Path)
		}
		if f.Family == "widgets" || f.Family == "screen-component" {
			for y := 0; y < f.Height; y++ {
				for x := 0; x < f.Width; x++ {
					p := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
					if p.A == 0 && (p.R != 0 || p.G != 0 || p.B != 0) {
						t.Fatal("hidden RGB in transparent widget")
					}
				}
			}
		}
		if f.Family != "review" && f.Family != "backdrop" {
			if ids[f.ID] == nil {
				ids[f.ID] = map[int]bool{}
			}
			ids[f.ID][f.Density] = true
		}
	}
	if counts["widgets"] != 183 || counts["screen-frame"] != 60 || counts["screen-component"] != 12 || counts["review"] != 20 {
		t.Fatalf("incomplete state/widget coverage: %v", counts)
	}
	for id, densities := range ids {
		if !densities[1] || !densities[2] || !densities[4] {
			t.Fatalf("missing density: %s", id)
		}
	}
	// The runtime copies in assets/ are this export, byte for byte.
	if n := checkRuntimePins(t, "screens", r.Files, func(f Output) bool {
		return f.Family != "review" && f.Density > 0 && strings.HasSuffix(f.Path, ".png")
	}); n != 278 {
		t.Fatalf("compared %d runtime files, want 278", n)
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
			t.Fatalf("gallery link: %s", p)
		}
	}
	if !bytes.Contains(h, []byte(`role="region" tabindex="0"`)) || !bytes.Contains(h, []byte("appearance:none;padding-right:44px")) {
		t.Fatal("review keyboard/chevron contract missing")
	}
	out := filepath.Join(t.TempDir(), "repeat")
	if e = export(pipelineManifest("screens-v2"), out); e != nil {
		t.Fatal(e)
	}
	n := 0
	if e = filepath.WalkDir(root, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		rel, e := filepath.Rel(root, p)
		if e != nil {
			return e
		}
		want, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		got, e := os.ReadFile(filepath.Join(out, rel))
		if e != nil {
			return e
		}
		if !bytes.Equal(want, got) {
			t.Errorf("repeat differs: %s", rel)
		}
		n++
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	if n != len(r.Files)+1 || n != 400 {
		t.Fatalf("unreported/unexpected files: %d", n)
	}
}
