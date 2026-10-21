package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadFixture(t *testing.T) *project {
	t.Helper()
	p, err := load("testdata", "testdata/reference.json")
	if err != nil {
		t.Fatal(err)
	}
	p.assets = os.DirFS("../../assets")
	p.targetsPath = filepath.Join(t.TempDir(), "targets.json")
	return p
}

func TestFixtureIsValid(t *testing.T) {
	p := loadFixture(t)
	if errs := p.validate(); len(errs) > 0 {
		t.Fatal(errors.Join(errs...))
	}
	if _, present, err := p.checkImage(); err != nil || present {
		t.Fatalf("missing image: present=%t err=%v, want absent without error", present, err)
	}
}

func TestValidateRejects(t *testing.T) {
	fam := func(p *project, id string) *family {
		for i := range p.ref.Families {
			if p.ref.Families[i].ID == id {
				return &p.ref.Families[i]
			}
		}
		t.Fatalf("no family %s", id)
		return nil
	}
	for _, tc := range []struct {
		name, want string
		edit       func(p *project)
	}{
		{"duplicate family", "duplicate id", func(p *project) { p.ref.Families = append(p.ref.Families, p.ref.Families[1]) }},
		{"art path", `want "art/terrain/forest"`, func(p *project) { fam(p, "forest").Art = "art/forest" }},
		{"uncovered terrain", "terrain hills is covered by 0", func(p *project) {
			p.ref.Families = p.ref.Families[:2] // drops hills and the infra families
		}},
		{"doubly covered terrain", "terrain grass is covered by 2", func(p *project) {
			f := fam(p, "forest")
			f.RuntimeIDs = append(f.RuntimeIDs, "grass")
		}},
		{"uncovered infra", "infra path14 is covered by 0", func(p *project) { fam(p, "path").InfraPrefixes = []string{"trail"} }},
		{"doubly covered infra", "infra road25 is covered by 2", func(p *project) {
			f := fam(p, "path")
			f.InfraPrefixes = append(f.InfraPrefixes, "ro")
		}},
		{"exemplar terrain", "exemplar r1c1 is terrain=grass infra=road25", func(p *project) { fam(p, "forest").Exemplars[0] = "r1c1" }},
		{"exemplar infra", "exemplar r1c2 is terrain=forest", func(p *project) { fam(p, "road").Exemplars[0] = "r1c2" }},
		{"exemplar excluded", "lies in excluded row 4", func(p *project) { fam(p, "forest").Exemplars[0] = "r4c2" }},
		{"exemplar off the level", "r2c3 is not on the level", func(p *project) { fam(p, "hills").Exemplars[0] = "r2c3" }},
		{"exemplar malformed", "not of the form", func(p *project) { fam(p, "hills").Exemplars[0] = "r02c" }},
		{"exemplar twice", "listed twice", func(p *project) { fam(p, "forest").Exemplars[1] = "r1c2" }},
		{"too few exemplars", "2 exemplars, want 3 to 6", func(p *project) { fam(p, "meadow").Exemplars = []string{"r1c3", "r3c1"} }},
		{"too many exemplars", "7 exemplars", func(p *project) {
			fam(p, "meadow").Exemplars = []string{"r1c3", "r3c1", "r2c1", "r3c2", "r1c1", "r1c3", "r3c1"}
		}},
		{"kind", `kind "rail"`, func(p *project) { fam(p, "hills").Kind = "rail" }},
		{"terrain family with prefixes", "lists runtime_ids only", func(p *project) { fam(p, "hills").InfraPrefixes = []string{"x"} }},
		{"absent family", "does not occur", func(p *project) {
			p.ref.Families = append(p.ref.Families, family{ID: "rail", Kind: "infra", InfraPrefixes: []string{"rail"}, Art: "art/terrain/rail"})
		}},
		{"feature excluded cell", "feature crossing: cell r4c1 lies in excluded row 4", func(p *project) { p.ref.Features[0].Cells[0] = "r4c1" }},
		{"feature notes", "title and notes are required", func(p *project) { p.ref.Features[0].Notes = " " }},
		{"duplicate feature", "duplicate", func(p *project) { p.ref.Features = append(p.ref.Features, p.ref.Features[0]) }},
		{"sha", "sha256 must be 64 hex digits", func(p *project) { p.ref.Image.SHA256 = "abc" }},
		{"registration", "registration", func(p *project) { p.ref.Registration.ScaleY[1] = 0 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := loadFixture(t)
			tc.edit(p)
			err := errors.Join(p.validate()...)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("errors %v, want one containing %q", err, tc.want)
			}
		})
	}
}

func TestFeatureCellsNeedNoFamilyMatch(t *testing.T) {
	p := loadFixture(t)
	p.ref.Features[0].Cells = append(p.ref.Features[0].Cells, "r1c2", "r2c2")
	if errs := p.validate(); len(errs) > 0 {
		t.Fatal(errors.Join(errs...))
	}
}

func TestExcludedRowsNeedNoFamily(t *testing.T) {
	p := loadFixture(t)
	p.excluded = map[int]bool{}
	err := errors.Join(p.validate()...)
	if err == nil || !strings.Contains(err.Error(), "terrain water is covered by 0") {
		t.Fatalf("errors %v, want water uncovered once row 4 is included", err)
	}
}

func TestLoadRejectsBadExcludedRow(t *testing.T) {
	dir := t.TempDir()
	raw, err := os.ReadFile("testdata/reference.json")
	if err != nil {
		t.Fatal(err)
	}
	raw = bytes.Replace(raw, []byte(`"excluded_rows": [4]`), []byte(`"excluded_rows": [5]`), 1)
	level, err := os.ReadFile("testdata/level.json")
	if err != nil {
		t.Fatal(err)
	}
	for name, b := range map[string][]byte{"reference.json": raw, "level.json": level} {
		if err := os.WriteFile(filepath.Join(dir, name), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := load(dir, filepath.Join(dir, "reference.json")); err == nil || !strings.Contains(err.Error(), "excluded row 5") {
		t.Fatalf("load error %v, want excluded row 5 rejected", err)
	}
}

func TestDumpCells(t *testing.T) {
	got := loadFixture(t).dump()
	for _, line := range []string{"r1c1 terrain=grass infra=road25\n", "r2c1 terrain=grass infra=stream14,bridge25\n", "r1c2 terrain=forest\n", "r4c1 terrain=water excluded\n"} {
		if !strings.Contains(got, line) {
			t.Errorf("dump lacks %q:\n%s", line, got)
		}
	}
	if n := strings.Count(got, "\n"); n != 10 {
		t.Errorf("dump has %d lines, want 10", n)
	}
}

func TestParseCell(t *testing.T) {
	for _, s := range []string{"r12c3", "r1c1"} {
		c, err := parseCell(s)
		if err != nil || c.String() != s {
			t.Errorf("parseCell(%q) = %v, %v", s, c, err)
		}
	}
	for _, s := range []string{"r0c1", "r01c1", "c1r1", "r1c", "r1c1 "} {
		if _, err := parseCell(s); err == nil {
			t.Errorf("parseCell(%q) accepted", s)
		}
	}
	if got := (cell{3, 7}).fileName(); got != "r03c07.png" {
		t.Errorf("fileName = %s", got)
	}
}

func TestRegistrationMatchesPulseLattice(t *testing.T) {
	for _, w := range [][2]float64{{0, 0}, {48, 40}, {1090, 1880}, {355.5, 77.25}} {
		x, y := pulseRegistration.sourcePoint(w[0], w[1])
		wantX, wantY := w[0]*123/142+42-48.0*123/142, w[1]*35/40
		if math.Abs(x-wantX) > 1e-9 || math.Abs(y-wantY) > 1e-9 {
			t.Errorf("sourcePoint(%v) = (%v,%v), want (%v,%v)", w, x, y, wantX, wantY)
		}
	}
}

// writeImage stores a plain painted PNG as the fixture image and records its hash.
func writeImage(t *testing.T, p *project) {
	t.Helper()
	im := image.NewNRGBA(image.Rect(0, 0, p.ref.Image.Width, p.ref.Image.Height))
	for i := range im.Pix {
		im.Pix[i] = 200
	}
	im.SetNRGBA(0, 0, color.NRGBA{A: 255})
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	p.root = t.TempDir()
	if err := os.WriteFile(filepath.Join(p.root, p.ref.Image.Path), b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b.Bytes())
	p.ref.Image.SHA256 = hex.EncodeToString(sum[:])
}

func TestCheckImage(t *testing.T) {
	p := loadFixture(t)
	writeImage(t, p)
	if _, present, err := p.checkImage(); err != nil || !present {
		t.Fatalf("present=%t err=%v, want a verified image", present, err)
	}
	good := p.ref.Image.SHA256
	p.ref.Image.SHA256 = strings.Repeat("0", 64)
	if _, _, err := p.checkImage(); err == nil || !strings.Contains(err.Error(), "sha256") {
		t.Fatalf("err %v, want sha mismatch", err)
	}
	p.ref.Image.SHA256, p.ref.Image.Width = good, 99
	if _, _, err := p.checkImage(); err == nil || !strings.Contains(err.Error(), "want 99×80") {
		t.Fatalf("err %v, want size mismatch", err)
	}
}

func TestRenderNeedsImage(t *testing.T) {
	err := loadFixture(t).render(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "missing.png is missing") {
		t.Fatalf("err %v, want a missing-image message", err)
	}
}

func TestRender(t *testing.T) {
	p := loadFixture(t)
	p.ref.Image.Width, p.ref.Image.Height = 300, 160
	writeImage(t, p)
	out := t.TempDir()
	if err := os.MkdirAll(filepath.Join(out, "tiles"), 0o755); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(out, "tiles", "r09c09.png")
	if err := os.WriteFile(stale, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := p.render(out); err != nil {
		t.Fatal(err)
	}
	tiles, err := filepath.Glob(filepath.Join(out, "tiles", "*.png"))
	if err != nil {
		t.Fatal(err)
	}
	if len(tiles) != 8 { // rows 1-3; row 4 is excluded
		t.Fatalf("%d tiles, want 8: %v", len(tiles), tiles)
	}
	raw, err := os.ReadFile(filepath.Join(out, "tiles", "r02c01.png"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(raw))
	if err != nil || cfg.Width != 192 || cfg.Height != 160 {
		t.Fatalf("tile %dx%d (%v), want 192x160", cfg.Width, cfg.Height, err)
	}
	html, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`id="meadow"`, `id="feature-crossing"`, "A road meets a stream.", `src="tiles/r02c01.png"`, "All occurrences (5)", `id="comparison"`, `src="compare/band-1.png"`, `src="compare/feature-crossing.png"`, "art/terrain/bridge/prompt.md", `src="reference.png"`} {
		if !bytes.Contains(html, []byte(want)) {
			t.Errorf("index.html lacks %q", want)
		}
	}
	first := append([]byte(nil), html...)
	if err := p.render(out); err != nil {
		t.Fatal(err)
	}
	if again, _ := os.ReadFile(filepath.Join(out, "index.html")); !bytes.Equal(first, again) {
		t.Error("index.html differs between two renders")
	}
}
