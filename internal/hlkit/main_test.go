package hlkit

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestPNG(t *testing.T, path string, im image.Image) string {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(b.Bytes()))
}

// testSheet is a 2×1 sheet: a 40 px square with a hazy rim in the left cell,
// a 20×10 px bar in the right cell, both on an alpha-20 matte.
func testSheet() *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, 200, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 200; x++ {
			im.SetNRGBA(x, y, color.NRGBA{120, 100, 60, 20})
		}
	}
	for y := 30; y < 70; y++ {
		for x := 30; x < 70; x++ {
			im.SetNRGBA(x, y, color.NRGBA{120, 100, 60, 255})
		}
	}
	for y := 45; y < 55; y++ {
		for x := 140; x < 160; x++ {
			im.SetNRGBA(x, y, color.NRGBA{60, 60, 60, 255})
		}
	}
	return im
}

func testProfile(t *testing.T, dir string, cells string) string {
	t.Helper()
	sheetSHA := writeTestPNG(t, filepath.Join(dir, "art/terrain/test/test-kit-shared-v1.png"), testSheet())
	ground := image.NewNRGBA(image.Rect(0, 0, 300, 200))
	for i := range ground.Pix {
		ground.Pix[i] = [4]uint8{100, 107, 40, 255}[i%4]
	}
	groundSHA := writeTestPNG(t, filepath.Join(dir, "ground.png"), ground)
	t.Chdir(dir)
	p := fmt.Sprintf(`{"schema_version":1,"densities":[2,4],"meadow":{"source":"ground.png","source_sha256":"%s"},
 "sheets":[{"id":"kit/test/kit","art":"terrain/test","source":"art/terrain/test/test-kit-shared-v1.png","source_sha256":"%s",
  "grid":[2,1],"cut":32,"min_area":4,"scale":10,"grade_rgb":[140,132,118],"cells":[%s]}]}`, groundSHA, sheetSHA, cells)
	if err := os.WriteFile("profile.json", []byte(p), 0o644); err != nil {
		t.Fatal(err)
	}
	return "profile.json"
}

type testReport struct {
	Sheets []sheetReport `json:"sheets"`
}

func readReport(t *testing.T) testReport {
	t.Helper()
	b, err := os.ReadFile("out/kits.json")
	if err != nil {
		t.Fatal(err)
	}
	var r testReport
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestRunPreparesSizedGradedSprites(t *testing.T) {
	p := testProfile(t, t.TempDir(), `{"cell":"r1c1","size":5},{"cell":"r1c2"}`)
	if err := run(p, "out", true); err != nil {
		t.Fatal(err)
	}
	r := readReport(t)
	s := r.Sheets[0]
	if !s.Pass || s.HazeFrom == 0 || s.Haze != 0 || len(s.Items) != 2 || len(s.Contact) != 2 {
		t.Fatalf("sheet report %+v", s)
	}
	square, bar := s.Items[0], s.Items[1]
	if square.ExtentFrom != 4 || square.Extent != 5 {
		t.Fatalf("square extent %v -> %v, want 4 -> 5", square.ExtentFrom, square.Extent)
	}
	// Without a size the bar keeps the painting scale: 20 px at 10 px per world px.
	if bar.Extent != 2 || bar.Outputs[1].Width != 8 || bar.Outputs[1].Height != 4 {
		t.Fatalf("bar %+v", bar)
	}
	if square.Outputs[0].Width != 10 || square.Outputs[1].Width != 20 || square.Outputs[0].Anchor != [2]float64{2.5, 2.5} {
		t.Fatalf("square outputs %+v", square.Outputs)
	}
	f, err := os.Open(filepath.Join("out", filepath.FromSlash(square.Outputs[1].Path)))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	im, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, a := im.At(0, 0).RGBA(); a != 0xffff {
		t.Fatal("the sprite lost its opaque corner")
	}
}

func TestRunFailsAnEmptyOrUnlistedCell(t *testing.T) {
	p := testProfile(t, t.TempDir(), `{"cell":"r1c1"}`)
	err := run(p, "out", true)
	if err == nil || !strings.Contains(err.Error(), "kit/test/kit") {
		t.Fatalf("an unlisted item passed: %v", err)
	}
	if err := run(p, "out", false); err != nil {
		t.Fatalf("report mode failed: %v", err)
	}
	s := readReport(t).Sheets[0]
	if s.Pass || !strings.Contains(s.Checks[len(s.Checks)-1].Name, "r1c2") {
		t.Fatalf("checks %+v", s.Checks)
	}
}

// testIndex is an assets/index.json with one standing entry, an earlier
// install to be replaced and a closing ground entry.
const testIndex = `{
  "schema_version": 1,
  "assets": [
    {"id":"standing/props/trees/1","pack":"terrain-standing-2x","density":2,"path":"terrain/forest/props-1-0-2x.png","sha256":"aa","width":256,"height":224,"logical_anchor":[64,56],"layer":"standing","art":"terrain/forest"},
    {"id":"standing/scatter/old","pack":"terrain-standing-2x","density":2,"path":"terrain/test/scatter-old-2x.png","sha256":"bb","width":2,"height":2,"logical_anchor":[1,1],"layer":"standing","art":"terrain/test"},
    {"id":"tile/grass","pack":"terrain-backgrounds-2x","density":2,"path":"terrain/grass/grass-2x.png","sha256":"cc","width":192,"height":160,"logical_anchor":[48,40],"art":"terrain/grass"}
  ]
}
`

func TestInstallWritesSpritesShadowsAndIndex(t *testing.T) {
	p := testProfile(t, t.TempDir(), `{"cell":"r1c1","size":5,"runtime":"standing/scatter/test/square"},{"cell":"r1c2"}`)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	b = bytes.Replace(b, []byte(`"cells":`), []byte(`"shadow":{"rgb":[18,26,16],"opacity":0.6,"length":0.25},"cells":`), 1)
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("assets", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("assets/index.json", []byte(testIndex), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := writePNG("assets/terrain/test/scatter-old-2x.png", image.NewNRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	if err := install(p, "out", "assets"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("assets/index.json")
	if err != nil {
		t.Fatal(err)
	}
	var index struct {
		Assets []struct {
			ID      string     `json:"id"`
			Pack    string     `json:"pack"`
			Density int        `json:"density"`
			Path    string     `json:"path"`
			SHA     string     `json:"sha256"`
			Width   int        `json:"width"`
			Anchor  [2]float64 `json:"logical_anchor"`
			Layer   string     `json:"layer"`
			Art     string     `json:"art"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(raw, &index); err != nil {
		t.Fatalf("index no longer parses: %v\n%s", err, raw)
	}
	var ids []string
	byKey := map[string]int{}
	for i, e := range index.Assets {
		ids = append(ids, fmt.Sprintf("%s@%d", e.ID, e.Density))
		byKey[fmt.Sprintf("%s@%d", e.ID, e.Density)] = i
	}
	want := []string{
		"standing/props/trees/1@2", "standing/scatter/test/square@2", "standing/scatter/test/square@4",
		"standing/shadow/scatter/test/square@2", "standing/shadow/scatter/test/square@4", "tile/grass@2",
	}
	if strings.Join(ids, " ") != strings.Join(want, " ") {
		t.Fatalf("index ids %v, want %v", ids, want)
	}
	crown, shadow := index.Assets[byKey["standing/scatter/test/square@4"]], index.Assets[byKey["standing/shadow/scatter/test/square@4"]]
	if crown.Pack != "terrain-standing-4x" || crown.Layer != "standing" || crown.Art != "terrain/test" || crown.Width != 20 || crown.Anchor != [2]float64{2.5, 2.5} {
		t.Fatalf("crown entry %+v", crown)
	}
	// Light from the upper left: the shadow falls 0.25 × 5 world px to the lower right.
	off := 0.25 * 5 / math.Sqrt2
	if shadow.Width != 20 || math.Abs(shadow.Anchor[0]-(2.5-off)) > 0.01 || math.Abs(shadow.Anchor[1]-(2.5-off)) > 0.01 {
		t.Fatalf("shadow entry %+v", shadow)
	}
	for _, e := range []int{byKey["standing/scatter/test/square@4"], byKey["standing/shadow/scatter/test/square@4"]} {
		f, err := os.ReadFile(filepath.Join("assets", index.Assets[e].Path))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(f)) != index.Assets[e].SHA {
			t.Fatalf("%s: hash differs from the index", index.Assets[e].Path)
		}
	}
	f, err := os.Open(filepath.Join("assets", shadow.Path))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	im, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if c := color.NRGBAModel.Convert(im.At(10, 10)).(color.NRGBA); c != (color.NRGBA{18, 26, 16, 153}) {
		t.Fatalf("shadow pixel %v", c)
	}
	if _, err := os.Stat("assets/terrain/test/scatter-old-2x.png"); err == nil {
		t.Fatal("the earlier install's file is left behind")
	}
}

func TestInstallRefusesAFailingSheet(t *testing.T) {
	p := testProfile(t, t.TempDir(), `{"cell":"r1c1","size":5,"runtime":"standing/scatter/test/square"}`)
	if err := os.MkdirAll("assets", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("assets/index.json", []byte(testIndex), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := install(p, "out", "assets"); err == nil {
		t.Fatal("installed a sheet whose checks fail")
	}
	if b, _ := os.ReadFile("assets/index.json"); string(b) != testIndex {
		t.Fatal("a failed install changed the index")
	}
}

func TestValidateRejectsBadCells(t *testing.T) {
	for _, cells := range []string{`{"cell":"r2c1"}`, `{"cell":"r1c1"},{"cell":"r1c1"}`, `{"cell":"x1"}`, `{"cell":"r01c1"}`} {
		p := testProfile(t, t.TempDir(), cells)
		if err := run(p, "out", false); err == nil || !strings.Contains(err.Error(), "cell") {
			t.Errorf("cells %s accepted: %v", cells, err)
		}
	}
}

// The shipped profile's cast shadows are the one light (terrainlight); a
// shadow that leaves it is refused.
func TestProfileShadowsAreTheOneLight(t *testing.T) {
	raw, err := os.ReadFile("../../art/pipeline/world-kits-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var p profile
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatal(err)
	}
	if err := validate(p); err != nil {
		t.Fatal(err)
	}
	for i, s := range p.Sheets {
		if s.Shadow == nil {
			continue
		}
		old := *s.Shadow
		for _, sh := range []shadow{{RGB: [3]float64{24, 29, 22}, Opacity: old.Opacity, Length: old.Length}, {RGB: old.RGB, Opacity: 0.32, Length: old.Length}, {RGB: old.RGB, Opacity: old.Opacity, Length: 0.3}} {
			*p.Sheets[i].Shadow = sh
			if validate(p) == nil {
				t.Errorf("%s: shadow %+v passes", s.ID, sh)
			}
		}
		*p.Sheets[i].Shadow = old
	}
}
