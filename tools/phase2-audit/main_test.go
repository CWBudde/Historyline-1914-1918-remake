package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLiteralLoaderContract(t *testing.T) {
	source := []byte("bgTiles = {'water': pygame.image.load('gfx/water.png'), 'none': None}\nother={'invented': pygame.image.load('gfx/no.png')}")
	ids, err := loaderIDs(source, "bgTiles")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids["water"] != "legacy/python/gfx/water.png" {
		t.Fatalf("wrong enabled IDs: %v", ids)
	}
	for _, source := range []string{"other = {}", "bgTiles = {'water': pygame.image.load('a'), 'water': pygame.image.load('b')}", "bgTiles = {'none': None}", "bgTiles = {"} {
		if _, err := loaderIDs([]byte(source), "bgTiles"); err == nil {
			t.Fatalf("accepted invalid loader %s", source)
		}
	}
}

func TestRejectEscapingPaths(t *testing.T) {
	for _, path := range []string{"../art/a.png", "art/../a.png", "C:/original/a.png", "/tmp/a.png", "art/../../original/a.png"} {
		if _, err := safePath(path); err == nil {
			t.Fatalf("accepted %q", path)
		}
	}
	if _, err := safePath("art/exports/a.png"); err != nil {
		t.Fatal(err)
	}
}

func newTestAuditor(t *testing.T) (*auditor, string) {
	t.Helper()
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	})
	return &auditor{root: root, pngs: map[string][2]int{}, pins: map[string]string{}, seen: map[string]bool{}}, dir
}

func TestTamperedPinAndConflictingDimensions(t *testing.T) {
	a, dir := newTestAuditor(t)
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewNRGBA(image.Rect(0, 0, 7, 9))); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "asset.png"), b.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	a.hash("asset.png", fmt.Sprintf("%x", sha256.Sum256([]byte("different bytes"))))
	a.png("asset.png", 7, 9)
	a.png("asset.png", 8, 9)
	if len(a.r.Problems) != 2 || !strings.Contains(a.r.Problems[0], "SHA256 mismatch") || !strings.Contains(a.r.Problems[1], "conflicting expected dimensions") {
		t.Fatalf("problems: %v", a.r.Problems)
	}
}

func TestPinAndDimensionDeduplication(t *testing.T) {
	a, dir := newTestAuditor(t)
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewNRGBA(image.Rect(0, 0, 7, 9))); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "asset.png"), b.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		a.hash("asset.png", fmt.Sprintf("%x", sha256.Sum256(b.Bytes())))
		a.png("asset.png", 0, 0)
	}
	a.png("asset.png", 7, 9)
	if len(a.r.Problems) != 0 || len(a.pins) != 1 || len(a.pngs) != 1 {
		t.Fatalf("unexpected result: %+v", a.r.Problems)
	}
}

func TestOptionalScreenStatesRejectCorruptReport(t *testing.T) {
	a, dir := newTestAuditor(t)
	a.r.Sections = map[string]section{}
	a.screenStates()
	if len(a.r.Problems) != 0 {
		t.Fatalf("absent optional bundle: %v", a.r.Problems)
	}
	base := filepath.Join(dir, "art", "exports", "screens-v2")
	if err := os.MkdirAll(base, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "report.json"), []byte("{broken json"), 0644); err != nil {
		t.Fatal(err)
	}
	a.screenStates()
	if len(a.r.Problems) == 0 || !strings.Contains(strings.Join(a.r.Problems, "\n"), "report.json") {
		t.Fatalf("corrupt optional report accepted: %v", a.r.Problems)
	}
}

func TestRepositoryCandidateCoverage(t *testing.T) {
	r, err := audit(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Problems) != 0 {
		t.Fatalf("coverage audit:\n%s", strings.Join(r.Problems, "\n"))
	}
	for name, want := range map[string]int{"tiles": 134, "units": 27, "interface": 117, "screens": 16} {
		s := r.Sections[name]
		if s.IDs != want || s.RuntimeAccepted != 0 {
			t.Errorf("%s: got %+v", name, s)
		}
	}
	if r.Sections["units"].Families != 15 || r.NativeTerrain.IDs != 350 || r.NativeTerrain.Breakdown["summer"] != 175 || r.NativeTerrain.Breakdown["winter"] != 175 {
		t.Fatalf("family/native ledger counts: %+v", r)
	}
	units := r.Sections["units"]
	if units.RecipePNGs != 1905 || units.CandidatePNGs != 1125 || units.PlanePNGs != 360 || units.ReviewPNGs != 420 || units.Breakdown["sprite_pngs"] != 1080 || units.Breakdown["portrait_pngs"] != 45 {
		t.Fatalf("candidate/review/plane counts mixed: %+v", units)
	}
	if len(r.FutureGaps) != 9 {
		t.Fatalf("future gaps: %+v", r.FutureGaps)
	}
	for _, gap := range r.FutureGaps {
		if gap.Background || gap.Infrastructure {
			t.Errorf("unconfirmed future background enabled: %+v", gap)
		}
	}
	if r.CheckedPNGs < 2500 || r.CheckedPins < 2500 {
		t.Fatalf("audit did not inspect complete retained candidates: %d PNGs, %d pins", r.CheckedPNGs, r.CheckedPins)
	}
}
