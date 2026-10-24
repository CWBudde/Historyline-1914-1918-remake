package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const repoRoot = "../.."

// The committed bunker sprites are exactly what the manifest builds, every
// owner shares the neutral outline, and spent is the grey of normal.
func TestCommittedBunkerReproduces(t *testing.T) {
	m, base, err := readManifest(repoRoot, filepath.Join(repoRoot, "art/pipeline/unit-bunker-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	files, err := build(repoRoot, m, base, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3*6*2*len(m.Densities) {
		t.Fatalf("%d files, want every palette, facing, state and density", len(files))
	}
	for path, want := range files {
		got, err := readPNG(filepath.Join(repoRoot, "assets", path))
		if err != nil {
			t.Fatal(err)
		}
		g, ok := got.(*image.NRGBA)
		if !ok || g.Bounds() != want.Bounds() || !bytes.Equal(g.Pix, want.Pix) {
			t.Errorf("%s differs from the manifest's build; run go run ./tools/unit-topview -manifest art/pipeline/unit-bunker-v2.json", path)
		}
	}
	for _, d := range []string{"2x", "4x"} {
		neutral := files["units/bunker/neutral/n-normal-"+d+".png"]
		for _, p := range []string{"cp", "ec"} {
			im := files["units/bunker/"+p+"/sw-normal-"+d+".png"]
			for i := 3; i < len(im.Pix); i += 4 {
				if im.Pix[i] != neutral.Pix[i] {
					t.Fatalf("%s %s: outline differs from neutral", p, d)
				}
			}
			if bytes.Equal(im.Pix, neutral.Pix) {
				t.Errorf("%s %s: no insignia", p, d)
			}
			sp := files["units/bunker/"+p+"/sw-spent-"+d+".png"]
			for i := 0; i < len(sp.Pix); i += 4 {
				if sp.Pix[i] != im.Pix[i] || sp.Pix[i+1] != im.Pix[i] || sp.Pix[i+2] != im.Pix[i] || sp.Pix[i+3] != im.Pix[i+3] {
					t.Fatalf("%s %s: spent is not the grey of normal", p, d)
				}
			}
		}
	}
	// A bunker fills its hex: the solid box spans the diameter around the anchor.
	r, err := solidBounds(files["units/bunker/neutral/n-normal-4x.png"])
	if err != nil {
		t.Fatal(err)
	}
	if w := float64(r.Dx()) / 4; math.Abs(w-m.Diameter) > 1.5 {
		t.Errorf("solid width %.1f logical, want %.0f", w, m.Diameter)
	}
	if cx, cy := float64(r.Min.X+r.Max.X)/8, float64(r.Min.Y+r.Max.Y)/8; math.Abs(cx-anchorX) > 1 || math.Abs(cy-anchorY) > 1 {
		t.Errorf("solid centre (%.1f,%.1f), want the anchor (%v,%v)", cx, cy, anchorX, anchorY)
	}
}

// ringImage is a grey square with a dark ring of radius 40 and spokes around (cx, cy).
func ringImage(cx, cy float64) *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, 200, 200))
	for y := range 200 {
		for x := range 200 {
			dx, dy := float64(x)+.5-cx, float64(y)+.5-cy
			v := uint8(150)
			if d := math.Hypot(dx, dy); math.Abs(d-40) < 2 || (d > 40 && math.Abs(dx) < 1.5) || (d > 30 && math.Abs(dy-dx*.3) < 1.5) {
				v = 60
			}
			im.SetNRGBA(x, y, color.NRGBA{v, v, v, 255})
		}
	}
	return im
}

func TestRegisterFindsTheShift(t *testing.T) {
	base, edit := ringImage(100, 100), ringImage(103, 98)
	if dx, dy := register(base, edit, 100, 100, 30, 60, 6); dx != 3 || dy != -2 {
		t.Errorf("shift (%d,%d), want (3,-2)", dx, dy)
	}
	if dx, dy := register(base, base, 100, 100, 30, 60, 6); dx != 0 || dy != 0 {
		t.Errorf("identical images shift (%d,%d)", dx, dy)
	}
}

// The inlay replaces only the disc; outside it every pixel stays the base's.
func TestInlayKeepsTheOutside(t *testing.T) {
	base := ringImage(100, 100)
	edit := ringImage(100, 100)
	for y := 85; y < 115; y++ {
		for x := 85; x < 115; x++ {
			edit.SetNRGBA(x, y, color.NRGBA{200, 30, 30, 255})
		}
	}
	out, _ := inlay(base, edit, 100, 100, 30, 4)
	for y := range 200 {
		for x := range 200 {
			d := math.Hypot(float64(x)+.5-100, float64(y)+.5-100)
			if d > 31 && out.NRGBAAt(x, y) != base.NRGBAAt(x, y) {
				t.Fatalf("(%d,%d) outside the disc changed", x, y)
			}
		}
	}
	if c := out.NRGBAAt(100, 100); c.R < 150 || c.G > 60 {
		t.Errorf("centre %v, want the edit's red", c)
	}
}

func TestInstallRepinsOnlyItsEntries(t *testing.T) {
	dir := t.TempDir()
	index := `{"schema_version":1,"assets":[
    {"id":"unit/x/cp/n-normal","path":"units/x/cp/n-normal-2x.png","sha256":"` + strings.Repeat("0", 64) + `","width":2,"height":2},
    {"id":"unit/y/cp/n-normal","path":"units/y/cp/n-normal-2x.png","sha256":"` + strings.Repeat("1", 64) + `","width":2,"height":2}
]}
`
	if err := os.WriteFile(filepath.Join(dir, "index.json"), []byte(index), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "units/x/cp"), 0o755); err != nil {
		t.Fatal(err)
	}
	im := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	if err := install(dir, map[string]*image.NRGBA{"units/z/cp/n-normal-2x.png": im}); err == nil {
		t.Error("a file without an index entry was installed")
	}
	if err := install(dir, map[string]*image.NRGBA{"units/x/cp/n-normal-2x.png": im}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(got), "\n")
	if strings.Contains(lines[1], strings.Repeat("0", 64)) || !strings.Contains(lines[2], strings.Repeat("1", 64)) {
		t.Errorf("index after install:\n%s", got)
	}
	f, err := os.Open(filepath.Join(dir, "units/x/cp/n-normal-2x.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if _, err := png.Decode(f); err != nil {
		t.Fatal(err)
	}
}
