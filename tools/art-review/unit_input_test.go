package artreview

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func unitTestWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func unitTestPNG(t *testing.T, path string, im *image.NRGBA) connectionMaterial {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	unitTestWrite(t, path, b.Bytes())
	s := sha256.Sum256(b.Bytes())
	return connectionMaterial{Source: filepath.Base(path), SourceSHA256: hex.EncodeToString(s[:])}
}

func unitTestOwned(x, y, d int) bool {
	v := [][2]int{{25, 0}, {71, 0}, {96, 40}, {71, 80}, {25, 80}, {0, 40}}
	for i, a := range v {
		b := v[(i+1)%6]
		dx, dy := b[0]-a[0], b[1]-a[1]
		cross := dx*(2*y+1-2*d*a[1]) - dy*(2*x+1-2*d*a[0])
		tieIncluded := dy < 0 || dy == 0 && dx > 0
		if cross < 0 || cross == 0 && !tieIncluded {
			return false
		}
	}
	return true
}

func unitTestFixture(t *testing.T) (string, []byte) {
	t.Helper()
	return unitTestFamilyFixture(t, "infantry")
}

func TestUnitAtlasRejectsBackgroundMatte(t *testing.T) {
	const cell = 384
	atlas := image.NewNRGBA(image.Rect(0, 0, cell*2, cell*3))
	for face := range 6 {
		ox, oy := face%2*cell, face/2*cell
		for y := range cell {
			for x := range cell {
				p := color.NRGBA{180, 0, 180, 1} // faint whole-cell matte
				if x < 8 && y < 8 {
					p = color.NRGBA{} // the old 64-pixel minimum was insufficient
				} else if x >= 140 && x < 240 && y >= 140 && y < 240 {
					p = color.NRGBA{200, 0, 200, 255}
				}
				atlas.SetNRGBA(ox+x, oy+y, p)
			}
		}
	}
	if err := unitValidateAtlas(atlas); err == nil {
		t.Fatal("accepted a broad alpha-one background matte")
	}
	// Removing the matte, while retaining the authored object, is sufficient.
	for y := range atlas.Bounds().Dy() {
		for x := range atlas.Bounds().Dx() {
			if atlas.NRGBAAt(x, y).A == 1 {
				atlas.SetNRGBA(x, y, color.NRGBA{})
			}
		}
	}
	if err := unitValidateAtlas(atlas); err != nil {
		t.Fatal(err)
	}
}

func TestUnitAtlasRejectsArtworkCrossingCellEdge(t *testing.T) {
	path, _ := unitTestFixture(t)
	in, err := loadUnitInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	// A wing extending across an atlas seam is clipped in one view and becomes
	// a detached fragment of the next. Output padding cannot repair that source.
	cell := in.Atlas.Image.Bounds().Dx() / 2
	in.Atlas.Image.SetNRGBA(cell-1, cell/2, color.NRGBA{200, 0, 200, 255})
	in.Atlas.Image.SetNRGBA(cell, cell/2, color.NRGBA{200, 0, 200, 255})
	if err := unitValidateAtlas(in.Atlas.Image); err == nil {
		t.Fatal("accepted opaque artwork crossing an atlas cell edge")
	}
}

func TestRetainedUnitAtlasInputs(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "..", "art", "pipeline", "unit-*-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no retained unit manifests")
	}
	for _, path := range paths {
		if filepath.Base(path) == "unit-panel-v1.json" {
			continue
		}
		t.Run(filepath.Base(path), func(t *testing.T) {
			if strings.HasSuffix(path, "-long-v1.json") {
				if _, err := loadShipInputs(path); err != nil {
					t.Fatal(err)
				}
				return
			}
			if _, err := loadUnitInputs(path); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func unitTestFamilyFixture(t *testing.T, family string) (string, []byte) {
	t.Helper()
	dir := t.TempDir()
	atlas := image.NewNRGBA(image.Rect(0, 0, 768, 1152))
	for face := range 6 {
		ox, oy := face%2*384, face/2*384
		for y := 90; y < 300; y++ {
			for x := 150; x < 230; x++ {
				p := color.NRGBA{200, 0, 200, 255}
				if x < 170 {
					p = color.NRGBA{190, 140, uint8(70 + face*10), 255}
				}
				if y < 110 {
					p = color.NRGBA{uint8(30 + face*20), 60, 90, 255}
				}
				atlas.SetNRGBA(ox+x, oy+y, p)
			}
		}
		// Distinct small slot marker survives fixed framing but does not change body bounds.
		for y := 130; y < 145; y++ {
			for x := 90 + face*3; x < 105+face*3; x++ {
				atlas.SetNRGBA(ox+x, oy+y, color.NRGBA{80, 60, 40, 255})
			}
		}
	}
	atlasRef := unitTestPNG(t, filepath.Join(dir, "atlas.png"), atlas)
	grounds := map[string]map[string]connectionMaterial{}
	for idx, id := range []string{"grass", "grass_w", "mud"} {
		grounds[id] = map[string]connectionMaterial{}
		for _, d := range []int{2, 4} {
			im := image.NewNRGBA(image.Rect(0, 0, 96*d, 80*d))
			for y := range 80 * d {
				for x := range 96 * d {
					if unitTestOwned(x, y, d) {
						im.SetNRGBA(x, y, color.NRGBA{uint8(40 + idx*50), uint8(90 + x%17), uint8(30 + y%13), 255})
					}
				}
			}
			grounds[id][fmt.Sprintf("%dx", d)] = unitTestPNG(t, filepath.Join(dir, fmt.Sprintf("%s-%dx.png", id, d)), im)
		}
	}
	m := map[string]any{"schema_version": 1, "family": family, "atlas": atlasRef, "palettes": map[string]any{"cp": []int{104, 142, 83}, "ec": []int{190, 143, 92}, "neutral": []int{146, 146, 146}}, "grounds": grounds}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "manifest.json")
	unitTestWrite(t, path, b)
	return path, b
}

// Only the explicit bounded families may select the shared rendering pipeline.
func TestUnitFamilySelection(t *testing.T) {
	path, valid := unitTestFixture(t)
	for _, family := range []string{"infantry", "lightArtillery", "eliteInfantry", "lightartillery", "LightArtillery", "artillery", "lArtillery", "mediumArtillery", "heavyArtillery", "CP_lArtillery", "eliteinfantry", ""} {
		t.Run(family, func(t *testing.T) {
			b := bytes.Replace(valid, []byte(`"family":"infantry"`), []byte(fmt.Sprintf(`"family":%q`, family)), 1)
			unitTestWrite(t, path, b)
			in, err := loadUnitInputs(path)
			wantValid := family == "infantry" || family == "lightArtillery" || family == "eliteInfantry" || family == "mediumArtillery" || family == "heavyArtillery"
			if (err == nil) != wantValid {
				t.Fatalf("family %q: err=%v, want valid=%v", family, err, wantValid)
			}
			if wantValid && in.Manifest.Family != family {
				t.Fatal("family changed during loading")
			}
		})
	}
	unitTestWrite(t, path, bytes.Replace(valid, []byte(`"family":"infantry"`), []byte(`"family":null`), 1))
	if _, err := loadUnitInputs(path); err == nil {
		t.Fatal("accepted null family")
	}
}

// Case aliases, silent array truncation, numeric coercion, and duplicate keys must never alter the profile.
func TestUnitStrictManifest(t *testing.T) {
	path, valid := unitTestFixture(t)
	if _, err := loadUnitInputs(path); err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"null-root": "null", "array-root": "[]", "extra-document": string(valid) + "{}",
		"unknown":              strings.Replace(string(valid), `"family":"infantry"`, `"family":"infantry","facing":3`, 1),
		"case":                 strings.Replace(string(valid), `"family"`, `"Family"`, 1),
		"family":               strings.Replace(string(valid), `"infantry"`, `"eliteinfantry"`, 1),
		"version":              strings.Replace(string(valid), `"schema_version":1`, `"schema_version":1.0`, 1),
		"duplicate":            strings.Replace(string(valid), `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1),
		"version-null":         strings.Replace(string(valid), `"schema_version":1`, `"schema_version":null`, 1),
		"missing":              strings.Replace(string(valid), `"family":"infantry",`, "", 1),
		"palette-null":         strings.Replace(string(valid), `"cp":[104,142,83]`, `"cp":null`, 1),
		"palette-element-null": strings.Replace(string(valid), `[104,142,83]`, `[104,null,83]`, 1),
		"palette-short":        strings.Replace(string(valid), `[104,142,83]`, `[104,142]`, 1),
		"palette-long":         strings.Replace(string(valid), `[104,142,83]`, `[104,142,83,0]`, 1),
		"palette-float":        strings.Replace(string(valid), `[104,142,83]`, `[104,142,83.0]`, 1),
		"palette-negative":     strings.Replace(string(valid), `[104,142,83]`, `[-1,142,83]`, 1),
		"palette-range":        strings.Replace(string(valid), `[104,142,83]`, `[256,142,83]`, 1),
		"atlas-null":           strings.Replace(string(valid), `"atlas":{`, `"atlas":null,"unused":{`, 1),
		"atlas-case":           strings.Replace(string(valid), `"source":"atlas.png"`, `"Source":"atlas.png"`, 1),
		"atlas-path":           strings.Replace(string(valid), `"atlas.png"`, `"../atlas.png"`, 1),
		"atlas-original":       strings.Replace(string(valid), `"atlas.png"`, `"original/atlas.png"`, 1),
		"atlas-hash":           strings.Replace(string(valid), `"source":"atlas.png"`, `"source":"grass-2x.png"`, 1),
		"ground-null":          strings.Replace(string(valid), `"grass":{`, `"grass":null,"unexpected":{`, 1),
		"density-case":         strings.Replace(string(valid), `"2x":{`, `"2X":{`, 1),
	}
	for name, b := range cases {
		t.Run(name, func(t *testing.T) {
			unitTestWrite(t, path, []byte(b))
			if _, err := loadUnitInputs(path); err == nil {
				t.Fatal("accepted invalid manifest")
			}
		})
	}
}

// Independently pinned files still need semantic geometry/content checks.
func TestUnitRejectsInvalidInputs(t *testing.T) {
	for _, kind := range []string{"atlas-small", "atlas-aspect", "atlas-opaque", "atlas-empty-panel", "atlas-no-key", "ground-inside-hole", "ground-outside-alpha", "ground-outside-rgb", "ground-size", "symlink-escape"} {
		t.Run(kind, func(t *testing.T) {
			path, b := unitTestFixture(t)
			var m map[string]any
			if err := json.Unmarshal(b, &m); err != nil {
				t.Fatal(err)
			}
			dir := filepath.Dir(path)
			if strings.HasPrefix(kind, "atlas-") {
				im := unitTestLoad(t, filepath.Join(dir, "atlas.png"))
				switch kind {
				case "atlas-small":
					im = image.NewNRGBA(image.Rect(0, 0, 766, 1149))
				case "atlas-aspect":
					im = image.NewNRGBA(image.Rect(0, 0, 768, 1151))
				case "atlas-opaque":
					for y := range im.Bounds().Dy() {
						for x := range im.Bounds().Dx() {
							im.SetNRGBA(x, y, color.NRGBA{200, 0, 200, 255})
						}
					}
				case "atlas-empty-panel":
					for y := range 384 {
						for x := range 384 {
							im.SetNRGBA(x, y, color.NRGBA{})
						}
					}
				case "atlas-no-key":
					for y := range 384 {
						for x := range 384 {
							p := im.NRGBAAt(x, y)
							p.G = p.R
							im.SetNRGBA(x, y, p)
						}
					}
				}
				m["atlas"] = unitTestPNG(t, filepath.Join(dir, "atlas.png"), im)
			} else if kind == "symlink-escape" {
				outside := filepath.Join(t.TempDir(), "outside.png")
				unitTestWrite(t, outside, unitTestRead(t, filepath.Join(dir, "atlas.png")))
				if err := os.Symlink(outside, filepath.Join(dir, "link.png")); err != nil {
					t.Fatal(err)
				}
				m["atlas"].(map[string]any)["source"] = "link.png"
			} else {
				g := m["grounds"].(map[string]any)["grass"].(map[string]any)
				im := unitTestLoad(t, filepath.Join(dir, "grass-2x.png"))
				switch kind {
				case "ground-inside-hole":
					im.SetNRGBA(96, 80, color.NRGBA{})
				case "ground-outside-alpha":
					im.SetNRGBA(0, 0, color.NRGBA{1, 2, 3, 255})
				case "ground-outside-rgb":
					im.SetNRGBA(0, 0, color.NRGBA{1, 2, 3, 0})
				case "ground-size":
					im = image.NewNRGBA(image.Rect(0, 0, 193, 160))
				}
				g["2x"] = unitTestPNG(t, filepath.Join(dir, "grass-2x.png"), im)
			}
			altered, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			unitTestWrite(t, path, altered)
			if _, err := loadUnitInputs(path); err == nil {
				t.Fatal("accepted invalid input")
			}
		})
	}
}

func unitTestRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func unitTestLoad(t *testing.T, path string) *image.NRGBA {
	t.Helper()
	im, err := unitDecodePNG(unitTestRead(t, path))
	if err != nil {
		t.Fatal(err)
	}
	return im
}
