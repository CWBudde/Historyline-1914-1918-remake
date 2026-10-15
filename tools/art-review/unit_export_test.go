package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// This covers every real exported image, not a surrogate sprite or preview-only render.
func TestUnitActualPNGBundle(t *testing.T) {
	for _, tc := range []struct {
		family  string
		classes []string
		title   string
	}{
		{"infantry", []string{"CP_infantry", "EC_infantry"}, "Infantry"},
		{"lightArtillery", []string{"CP_lArtillery", "EC_lArtillery"}, "Light artillery"},
		{"eliteInfantry", []string{"CP_eliteInf", "EC_eliteInf"}, "Elite infantry"},
	} {
		t.Run(tc.family, func(t *testing.T) {
			unitTestActualFamilyPNGBundle(t, tc.family, tc.classes, tc.title)
		})
	}
}

func unitTestActualFamilyPNGBundle(t *testing.T, family string, classes []string, title string) {
	t.Helper()
	path, manifest := unitTestFamilyFixture(t, family)
	out := filepath.Join(t.TempDir(), "new")
	if err := run([]string{"-unit", path, "-out", out}, io.Discard); err != nil {
		t.Fatal(err)
	}
	var r unitReport
	if err := json.Unmarshal(unitTestRead(t, filepath.Join(out, "report.json")), &r); err != nil {
		t.Fatal(err)
	}
	if len(r.Exports) != 124 || len(r.Frames) != 6 || len(r.Variants) != 36 || len(r.Sheets) != 4 || len(r.Scenes) != 24 {
		t.Fatalf("counts %d/%d/%d/%d/%d", len(r.Exports), len(r.Frames), len(r.Variants), len(r.Sheets), len(r.Scenes))
	}
	profilePrefix := "unit-infantry-v1:"
	if family == "lightArtillery" {
		profilePrefix = "unit-light-artillery-v1:"
	}
	if family == "eliteInfantry" {
		profilePrefix = "unit-elite-infantry-v1:"
	}
	if r.Family != family || !reflect.DeepEqual(r.ClassIDs, classes) || !strings.HasPrefix(r.Profile, profilePrefix) || !strings.Contains(r.Profile, classes[0]+" and "+classes[1]) || !strings.Contains(r.PaletteSelector, "independent of class ID/country") {
		t.Fatal("family report metadata", r.Family, r.ClassIDs, r.Profile, r.PaletteSelector)
	}
	ownerPalettes := map[string][]int{"cp": {104, 142, 83}, "ec": {190, 143, 92}, "neutral": {146, 146, 146}}
	if !reflect.DeepEqual(r.Palettes, ownerPalettes) {
		t.Fatal("owner palettes changed")
	}
	gallery := string(unitTestRead(t, filepath.Join(out, "index.html")))
	if !strings.Contains(gallery, "<title>"+title+" family candidate</title>") || !strings.Contains(gallery, "<h1>"+title+" family candidate</h1>") {
		t.Fatal("gallery family title")
	}
	if family == "lightArtillery" {
		if strings.Contains(strings.ToLower(r.Acceptance+r.Profile+r.PaletteSelector+gallery), "infantry") || !strings.Contains(r.Acceptance, "light artillery") || !strings.Contains(gallery, `alt="Seven light artillery over pinned terrain PNGs"`) {
			t.Fatal("artillery metadata or gallery describes infantry")
		}
	}
	if !bytes.Equal(manifest, unitTestRead(t, filepath.Join(out, r.ManifestSnapshot))) {
		t.Fatal("manifest bytes changed")
	}
	files, pngs := 0, 0
	if err := filepath.WalkDir(out, func(p string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() {
			files++
			if filepath.Ext(p) == ".png" {
				pngs++
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if files != 134 || pngs != 131 {
		t.Fatalf("physical files %d PNG %d", files, pngs)
	}
	all := map[string]*image.NRGBA{}
	canonicalOutputs := map[string]packOutput{}
	for _, p := range r.Exports {
		b := unitTestRead(t, filepath.Join(out, p.Path))
		sum := sha256.Sum256(b)
		if p.SHA256 != hex.EncodeToString(sum[:]) {
			t.Fatal("output hash", p.Path)
		}
		im := unitTestLoad(t, filepath.Join(out, p.Path))
		if p.Dimensions != (dimensions{im.Bounds().Dx(), im.Bounds().Dy()}) {
			t.Fatal("dimensions", p.Path)
		}
		if all[p.Path] != nil {
			t.Fatal("duplicate output")
		}
		all[p.Path] = im
		canonicalOutputs[p.Path] = p
	}
	seen := map[string]bool{}
	for face, f := range r.Frames {
		if f.FacingIndex != face || f.FacingLabel != []string{"n", "nw", "sw", "s", "se", "ne"}[face] || f.HeadingDegrees != []int{0, 300, 240, 180, 120, 60}[face] || f.SourceCell != [4]int{face % 2 * 384, face / 2 * 384, (face%2 + 1) * 384, (face/2 + 1) * 384} || f.SourceAnchor != (connectionPoint{192, 196}) || f.LogicalAnchor != (connectionPoint{48, 49}) || f.TopLeftOffset != (connectionPoint{0, -9}) || f.Inset != .8 || math.Abs(f.Scale4x-.8) > 1e-14 || f.Offset4x != (connectionPoint{38.4, 39.2}) {
			t.Fatal("frame registration/order", f)
		}
		for _, p := range f.Planes {
			if p.Base.Path != fmt.Sprintf("planes/%s/%s-base-%dx.png", family, f.FacingLabel, p.Density) || p.Mask.Path != fmt.Sprintf("planes/%s/%s-mask-%dx.png", family, f.FacingLabel, p.Density) {
				t.Fatal("family plane path")
			}
			n, k := all[p.Base.Path], all[p.Mask.Path]
			if p.Density == 4 {
				want := color.NRGBA{190, 140, uint8(70 + face*10), 255}
				if got := n.NRGBAAt(166, 199); got != want {
					t.Fatalf("authored panel order %d: got %v want %v", face, got, want)
				}
			}
			if n == nil || k == nil || n.Bounds() != image.Rect(0, 0, 96*p.Density, 96*p.Density) {
				t.Fatal("planes")
			}
			unitTestPlaneGuard(t, n, k, p.Density)
			seen[p.Base.Path] = true
			seen[p.Mask.Path] = true
			if p.Density == 2 {
				var n4, k4 *image.NRGBA
				for _, p4 := range f.Planes {
					if p4.Density == 4 {
						n4, k4 = all[p4.Base.Path], all[p4.Mask.Path]
					}
				}
				nr, e := resizeLanczos(n4, 192, 192)
				if e != nil {
					t.Fatal(e)
				}
				kr, e := resizeLanczos(k4, 192, 192)
				if e != nil {
					t.Fatal(e)
				}
				if !bytes.Equal(n.Pix, nr.Pix) || !bytes.Equal(k.Pix, kr.Pix) {
					t.Fatal("2x planes are not whole-4x Lanczos")
				}
			}
		}
	}
	variantKeys := map[string]bool{}
	for _, v := range r.Variants {
		if v.Family != family || !reflect.DeepEqual(v.ClassIDs, classes) || v.FacingLabel != []string{"n", "nw", "sw", "s", "se", "ne"}[v.FacingIndex] {
			t.Fatal("family bindings")
		}
		variantKey := v.FacingLabel + "/" + v.Palette + "/" + v.State
		if variantKeys[variantKey] || len(v.Outputs) != 2 {
			t.Fatal("duplicate variant or missing density", variantKey)
		}
		variantKeys[variantKey] = true
		for _, o := range v.Outputs {
			im := all[o.Path]
			d := o.Dimensions.Width / 96
			if o.Path != fmt.Sprintf("sprites/%s/%s/%s-%s-%dx.png", family, v.Palette, v.FacingLabel, v.State, d) || o != canonicalOutputs[o.Path] {
				t.Fatal("family variant mapping")
			}
			var n, k *image.NRGBA
			for _, p := range r.Frames[v.FacingIndex].Planes {
				if p.Density == d {
					n, k = all[p.Base.Path], all[p.Mask.Path]
				}
			}
			palette := ownerPalettes[v.Palette]
			for y := range im.Bounds().Dy() {
				for x := range im.Bounds().Dx() {
					b, m := n.NRGBAAt(x, y), k.NRGBAAt(x, y)
					want := color.NRGBA{}
					if b.A > 0 {
						round := func(v float64) uint8 {
							if v < 0 {
								v = 0
							}
							if v > 255 {
								v = 255
							}
							return uint8(v + .5)
						}
						want = color.NRGBA{round(float64(b.R) + float64(m.R)*(float64(palette[0])/255-1)), round(float64(b.G) + float64(m.R)*(float64(palette[1])/255-1)), round(float64(b.B) + float64(m.R)*(float64(palette[2])/255-1)), b.A}
						if v.State == "spent" {
							want.G, want.B = want.R, want.R
						}
					}
					if im.NRGBAAt(x, y) != want {
						t.Fatalf("composition %s (%d,%d)", o.Path, x, y)
					}
				}
			}
			unitTestPlaneGuard(t, im, k, d)
			seen[o.Path] = true
		}
	}
	if len(seen) != 96 {
		t.Fatalf("profile images %d", len(seen))
	}
	for _, face := range []string{"n", "nw", "sw", "s", "se", "ne"} {
		for _, palette := range []string{"cp", "ec", "neutral"} {
			for _, state := range []string{"normal", "spent"} {
				if !variantKeys[face+"/"+palette+"/"+state] {
					t.Fatal("missing facing/palette/state variant")
				}
			}
		}
	}
	for _, s := range r.Sheets {
		if s.Output.Path != fmt.Sprintf("sheets/%s-%s-%dx.png", family, s.Background, s.Density) {
			t.Fatal("family sheet path")
		}
		im := all[s.Output.Path]
		d := s.Density
		if im.Bounds() != image.Rect(0, 0, 576*d, 576*d) {
			t.Fatal("sheet dimensions")
		}
		for face := range 6 {
			for col := range 6 {
				palette := []string{"cp", "ec", "neutral"}[col%3]
				state := "normal"
				if col >= 3 {
					state = "spent"
				}
				sprite := all[fmt.Sprintf("sprites/%s/%s/%s-%s-%dx.png", family, palette, r.Frames[face].FacingLabel, state, d)]
				for _, p := range []image.Point{{0, 0}, {96 * d / 2, 96 * d / 2}, {96*d - 1, 96*d - 1}} {
					want := image.NewNRGBA(image.Rect(0, 0, 1, 1))
					wx, wy := col*96*d+p.X, face*96*d+p.Y
					c := color.NRGBA{255, 255, 255, 255}
					if s.Background == "checker" && ((wx/(8*d)+wy/(8*d))%2 == 1) {
						c = color.NRGBA{205, 205, 205, 255}
					}
					want.SetNRGBA(0, 0, c)
					draw.Draw(want, want.Bounds(), sprite, p, draw.Over)
					if im.NRGBAAt(wx, wy) != want.NRGBAAt(0, 0) {
						t.Fatal("sheet order/background")
					}
				}
			}
		}
	}
	offsets := [][2]int{{0, 0}, {71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}}
	labels := []string{"C", "NE", "SE", "S", "SW", "NW", "N"}
	facings := []int{0, 5, 4, 3, 2, 1, 0}
	faceLabels := []string{"n", "ne", "se", "s", "sw", "nw", "n"}
	palettes := []string{"neutral", "ec", "ec", "cp", "ec", "cp", "ec"}
	classIndices := []int{0, 0, 1, 0, 1, 1, 1}
	canonicalGrounds := map[string]connectionMaterialReport{}
	var inputManifest unitManifest
	if err := json.Unmarshal(manifest, &inputManifest); err != nil {
		t.Fatal(err)
	}
	for _, ground := range []string{"grass", "grass_w", "mud"} {
		for _, d := range []int{2, 4} {
			key := fmt.Sprintf("%dx", d)
			want := connectionMaterialReport{inputManifest.Grounds[ground][key], dimensions{96 * d, 80 * d}, fmt.Sprintf("sources/%s-%s.png", ground, key)}
			if r.Grounds[ground][key] != want {
				t.Fatal("canonical ground metadata", ground, key)
			}
			canonicalGrounds[ground+key] = want
		}
	}
	sceneKeys := map[string]bool{}
	for _, s := range r.Scenes {
		key := fmt.Sprintf("%s-%s-p%d-%dx", s.Ground, s.State, s.Parity, s.Density)
		if sceneKeys[key] || s.ID != key || s.Output.Path != "scenes/"+key+".png" || s.CanvasSize != (dimensions{238, 256}) {
			t.Fatal("scene Cartesian key/size", key)
		}
		sceneKeys[key] = true
		d := s.Density
		cx, cy := 0, 0
		if s.Parity == 1 {
			cx, cy = 71, 40
		}
		if s.CenterOrigin != (connectionPoint{float64(cx), float64(cy)}) || s.CanvasOrigin != (connectionPoint{float64(cx - 71), float64(cy - 89)}) || len(s.Cells) != 7 {
			t.Fatal("scene geometry")
		}
		want := image.NewNRGBA(image.Rect(0, 0, 238*d, 256*d))
		for i, c := range s.Cells {
			ox, oy := cx+offsets[i][0], cy+offsets[i][1]
			if c.Label != labels[i] || c.TerrainOrigin != (connectionPoint{float64(ox), float64(oy)}) || c.UnitTopLeft != (connectionPoint{float64(ox), float64(oy - 9)}) || c.RegistrationWorld != (connectionPoint{float64(ox + 48), float64(oy + 40)}) || c.FacingIndex != facings[i] || c.State != s.State {
				t.Fatal("cell anchor/facing", c)
			}
			ground := s.Ground
			if ground == "mud" && i > 0 {
				ground = "grass"
			}
			if c.Terrain != ground {
				t.Fatal("ground context")
			}
			// Derive the expected selection independently before trusting any nested pin
			// in the ordinary-Over reconstruction below.
			spriteName := fmt.Sprintf("sprites/%s/%s/%s-%s-%dx.png", family, palettes[i], faceLabels[i], s.State, d)
			canonicalSprite, exists := canonicalOutputs[spriteName]
			if !exists || c.SpriteInput != canonicalSprite || c.SpriteInput.Path != spriteName || c.GroundInput != canonicalGrounds[ground+fmt.Sprintf("%dx", d)] || c.Family != family || c.RuntimeID != classes[classIndices[i]] || c.Palette != palettes[i] || c.FacingLabel != faceLabels[i] {
				t.Fatalf("scene family selection or nested pins %s %s: %+v", s.ID, c.Label, c)
			}
			im := unitTestLoad(t, filepath.Join(out, c.GroundInput.RetainedSource))
			at := image.Pt((ox-cx+71)*d, (oy-cy+89)*d)
			draw.Draw(want, im.Bounds().Add(at), im, image.Point{}, draw.Over)
		}
		if s.Cells[1].RuntimeID != classes[0] || s.Cells[1].Palette != "ec" || s.Cells[5].RuntimeID != classes[1] || s.Cells[5].Palette != "cp" || s.Cells[0].Palette != "neutral" {
			t.Fatal("captured owner independent of class")
		}
		cells := append([]unitSceneCell(nil), s.Cells...)
		sort.SliceStable(cells, func(i, j int) bool {
			if cells[i].UnitTopLeft[1] != cells[j].UnitTopLeft[1] {
				return cells[i].UnitTopLeft[1] < cells[j].UnitTopLeft[1]
			}
			return cells[i].UnitTopLeft[0] < cells[j].UnitTopLeft[0]
		})
		for _, c := range cells {
			im := unitTestLoad(t, filepath.Join(out, c.SpriteInput.Path))
			at := image.Pt(int(c.UnitTopLeft[0]-s.CanvasOrigin[0])*d, int(c.UnitTopLeft[1]-s.CanvasOrigin[1])*d)
			draw.Draw(want, im.Bounds().Add(at), im, image.Point{}, draw.Over)
		}
		if !bytes.Equal(want.Pix, all[s.Output.Path].Pix) {
			t.Fatal("scene is not ordinary Over of pinned actual PNGs", s.ID)
		}
	}
	for _, ground := range []string{"grass", "grass_w", "mud"} {
		for _, state := range []string{"normal", "spent"} {
			for parity := range 2 {
				for _, d := range []int{2, 4} {
					if !sceneKeys[fmt.Sprintf("%s-%s-p%d-%dx", ground, state, parity, d)] {
						t.Fatal("missing scene Cartesian combination")
					}
				}
			}
		}
	}
}
func unitTestPlaneGuard(t *testing.T, n, k *image.NRGBA, d int) {
	t.Helper()
	for y := range n.Bounds().Dy() {
		for x := range n.Bounds().Dx() {
			p, m := n.NRGBAAt(x, y), k.NRGBAAt(x, y)
			if p.A != m.A {
				t.Fatal("plane alpha mismatch")
			}
			if p.A == 0 && p != (color.NRGBA{}) {
				t.Fatal("hidden RGB")
			}
			if (x < 6*d || y < 6*d || x >= 90*d || y >= 90*d) && p != (color.NRGBA{}) {
				t.Fatal("world guard")
			}
		}
	}
}

// Snapshot and render must consume the bytes loaded before any source mutation.
func TestUnitLoadedBytesSnapshotsAndExclusiveRollback(t *testing.T) {
	path, manifest := unitTestFixture(t)
	in, err := loadUnitInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	atlas := append([]byte(nil), in.Atlas.Bytes...)
	ground := append([]byte(nil), in.Grounds["grass"][2].Bytes...)
	unitTestWrite(t, path, []byte("changed"))
	unitTestWrite(t, filepath.Join(filepath.Dir(path), "atlas.png"), []byte("changed"))
	unitTestWrite(t, filepath.Join(filepath.Dir(path), "grass-2x.png"), []byte("changed"))
	b, err := buildUnitBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	snap := map[string][]byte{}
	for _, a := range b.artifacts {
		snap[a.name] = a.data
	}
	if !bytes.Equal(snap["manifest.json"], manifest) || !bytes.Equal(snap["sources/atlas.png"], atlas) || !bytes.Equal(snap["sources/grass-2x.png"], ground) {
		t.Fatal("snapshot reread")
	}
	valid, _ := unitTestFixture(t)
	dir := t.TempDir()
	out := filepath.Join(dir, "existing")
	if err = os.Mkdir(out, 0700); err != nil {
		t.Fatal(err)
	}
	unitTestWrite(t, filepath.Join(out, "marker"), []byte("keep"))
	before, err := os.Stat(out)
	if err != nil {
		t.Fatal(err)
	}
	if runUnit(valid, out, io.Discard) == nil {
		t.Fatal("existing directory accepted")
	}
	after, err := os.Stat(out)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(before, after) || string(unitTestRead(t, filepath.Join(out, "marker"))) != "keep" {
		t.Fatal("existing output changed")
	}
	failed := filepath.Join(dir, "failed")
	if runUnit(valid, failed, unitTestFailWriter{}) == nil {
		t.Fatal("log failure ignored")
	}
	if _, err = os.Lstat(failed); !os.IsNotExist(err) {
		t.Fatal("failed output not rolled back")
	}
}

type unitTestFailWriter struct{}

func (unitTestFailWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("intentional final writer failure")
}

// A rollback must not delete a replacement directory that the exporter did not create.
func TestUnitRollbackPreservesReplacementIdentity(t *testing.T) {
	path, _ := unitTestFixture(t)
	out := filepath.Join(t.TempDir(), "owned")
	writer := unitTestReplacingWriter{t: t, out: out}
	if runUnit(path, out, writer) == nil {
		t.Fatal("final writer error ignored")
	}
	if got := string(unitTestRead(t, filepath.Join(out, "foreign-marker"))); got != "preserve" {
		t.Fatal("replacement output was removed")
	}
}

type unitTestReplacingWriter struct {
	t   *testing.T
	out string
}

func (w unitTestReplacingWriter) Write([]byte) (int, error) {
	w.t.Helper()
	if err := os.Rename(w.out, w.out+"-renamed"); err != nil {
		w.t.Fatal(err)
	}
	if err := os.Mkdir(w.out, 0700); err != nil {
		w.t.Fatal(err)
	}
	unitTestWrite(w.t, filepath.Join(w.out, "foreign-marker"), []byte("preserve"))
	return 0, fmt.Errorf("intentional replaced-directory writer failure")
}
