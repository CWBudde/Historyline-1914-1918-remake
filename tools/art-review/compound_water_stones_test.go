package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func waterStonesScene() mountainSceneWant {
	return mountainSceneWant{"water-stones", "water-stones", "shared", 96, 80, []compoundTile{{Cell: "C", RuntimeID: "waterStones", Origin: connectionPoint{0, 0}}}}
}

func waterStonesFixture(t *testing.T) (string, compoundManifest) {
	t.Helper()
	path, m := mountainFixture(t)
	m.Grounds = map[string]connectionMaterial{"shared": m.Grounds["summer"]}
	m.Scenes = []compoundSceneInput{{ID: "water-stones", Artwork: m.Scenes[0].Artwork, Inset: .9}}
	compoundTestWriteManifest(t, path, m)
	return path, m
}

// Catch seasonal aliases, grass IDs, a quad-layout origin, or a fabricated tile ID.
func TestCompoundWaterStonesLiteralLayoutAndRuntimeOrigins(t *testing.T) {
	want := waterStonesScene()
	l, err := compoundLayoutFor(want.id)
	if err != nil {
		t.Fatal(err)
	}
	mountainCheckLayout(t, l, want)
	for parity, origin := range []connectionPoint{{0, 0}, {71, 40}} {
		p := compoundPatchFor(l, parity)
		mountainCheckRuntimePatch(t, p, want, parity)
		if p.SceneOrigin != origin || p.CanvasOrigin != origin.add(connectionPoint{-71, -80}) || len(p.Cells) != 7 {
			t.Fatalf("parity%d origins/count: %+v", parity, p)
		}
		neighbors := map[connectionPoint]bool{}
		for _, delta := range []connectionPoint{{71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}} {
			neighbors[origin.add(delta)] = true
		}
		for _, c := range p.Cells {
			id := "water"
			if c.Compound {
				id = "waterStones"
			} else if !neighbors[c.Origin] {
				t.Fatalf("unexpected water neighbor: %+v", c)
			}
			if c.RuntimeID != id || c.Season != "shared" {
				t.Fatalf("water runtime semantics: %+v", c)
			}
		}
	}
}

// Exercise the real CLI, retained pins, emitted PNG bytes, and ordinary Over
// assembly; no alternate renderer is substituted for the exported assets.
func TestCompoundWaterStonesCLIActualPNGs(t *testing.T) {
	path, manifest := waterStonesFixture(t)
	out := filepath.Join(filepath.Dir(path), "bundle")
	if err := run([]string{"-compound", path, "-out", out}, io.Discard); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(out, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report compoundReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Scenes) != 1 || len(report.Grounds) != 1 || len(report.Exports) != 14 {
		t.Fatalf("scene/ground/export count: %d/%d/%d", len(report.Scenes), len(report.Grounds), len(report.Exports))
	}
	want := waterStonesScene()
	scene := report.Scenes[0]
	mountainCheckLayout(t, scene.compoundLayout, want)
	if scene.ID != want.id || len(scene.Patches) != 2 || strings.Contains(scene.Profile, "grass only") || !strings.Contains(scene.Profile, "pinned shared water") || report.Profile != scene.Profile {
		t.Fatalf("water profile/scene contract: %+v", scene)
	}
	if scene.Transform.Source != (dimensions{1024, 640}) || math.Abs(scene.Transform.Scale-.084375) > 1e-12 || math.Abs(scene.Transform.Offset[0]-4.8) > 1e-12 || scene.Transform.Offset[1] != 13 || scene.Guard.GuardedOpaque != 0 {
		t.Fatalf("whole-source affine/guard: %+v %+v", scene.Transform, scene.Guard)
	}
	manifestBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mountainCheckSnapshot(t, out, report.ManifestSnapshot, manifestBytes, report.ManifestSHA256)
	for _, pair := range []struct {
		got connectionMaterialReport
		ref connectionMaterial
	}{{report.Grounds["shared"], manifest.Grounds["shared"]}, {scene.InputArtwork, manifest.Scenes[0].Artwork}} {
		if pair.got.connectionMaterial != pair.ref {
			t.Fatalf("source provenance changed: %+v", pair)
		}
		original, err := os.ReadFile(filepath.Join(filepath.Dir(path), pair.ref.Source))
		if err != nil {
			t.Fatal(err)
		}
		mountainCheckSnapshot(t, out, pair.got.RetainedSource, original, pair.got.SourceSHA256)
	}
	material := mountainLoadPNG(t, filepath.Dir(path), manifest.Grounds["shared"].Source)
	primitive, err := exportConnectionTiles(connectionInputs{manifest: connectionManifest{Geometry: sharedConnectionGeometry()}, materials: map[string]*image.NRGBA{"grass": material}}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, density := range []int{2, 4} {
		groundName := fmt.Sprintf("ground/water-%dx.png", density)
		ground := mountainLoadPNG(t, out, groundName)
		var encoded bytes.Buffer
		if err := png.Encode(&encoded, primitive[density]); err != nil {
			t.Fatal(err)
		}
		emitted, err := os.ReadFile(filepath.Join(out, groundName))
		if err != nil || !bytes.Equal(emitted, encoded.Bytes()) {
			t.Fatalf("ground PNG differs from existing opaque primitive at %dx: %v", density, err)
		}
		full := mountainLoadPNG(t, out, fmt.Sprintf("water-stones-scene-%dx.png", density))
		tile := mountainLoadPNG(t, out, fmt.Sprintf("tiles/waterStones-%dx.png", density))
		if full.Bounds() != image.Rect(0, 0, 96*density, 80*density) || tile.Bounds() != full.Bounds() || bytes.Equal(full.Pix, ground.Pix) {
			t.Fatalf("scene dimensions/artwork at %dx", density)
		}
		assembled := image.NewNRGBA(full.Bounds())
		draw.Draw(assembled, tile.Bounds(), tile, image.Point{}, draw.Over)
		mountainEqualPNG(t, assembled, full, "ordinary decoded tile Over assembly")
		mountainEqualPNG(t, mountainLoadPNG(t, out, fmt.Sprintf("water-stones-assembled-%dx.png", density)), full, "saved assembly")
		for y := range tile.Bounds().Dy() {
			for x := range tile.Bounds().Dx() {
				c := tile.NRGBAAt(x, y)
				if (mountainOwned(x, y, density) && c.A != 255) || (!mountainOwned(x, y, density) && c != (color.NRGBA{})) {
					t.Fatalf("owned opaque/exterior zero at %d,%d: %v", x, y, c)
				}
				// Independent distances to the six literal support lines. Every
				// owned pixel within three world pixels remains quiet water.
				px, py := (float64(x)+.5)/float64(density), (float64(y)+.5)/float64(density)
				slopeLength := math.Hypot(40, 25)
				d := min(py, 80-py, (40*px+25*py-1000)/slopeLength, (2840-40*px+25*py)/slopeLength, (4840-40*px-25*py)/slopeLength, (40*px-25*py+1000)/slopeLength)
				if mountainOwned(x, y, density) && d <= 3 && c != ground.NRGBAAt(x, y) {
					t.Fatalf("quiet water boundary band changed at %d,%d %dx", x, y, density)
				}
			}
		}
		for parity, origin := range []connectionPoint{{0, 0}, {71, 40}} {
			patch := scene.Patches[parity]
			mountainCheckRuntimePatch(t, patch, want, parity)
			actual := mountainLoadPNG(t, out, fmt.Sprintf("water-stones-patch-parity%d-%dx.png", parity, density))
			mountainEqualPNG(t, actual, mountainSurround(want, ground, map[string]*image.NRGBA{"waterStones": tile}, density, parity), "independent decoded PNG surround")
			phaseWitness := false
			for _, point := range []image.Point{{48, 1}, {1, 40}, {94, 40}, {48, 78}} {
				x, y := point.X*density, point.Y*density
				c := full.NRGBAAt(x, y)
				if c != ground.NRGBAAt(x, y) || actual.NRGBAAt(x+71*density, y+80*density) != c {
					t.Fatalf("quiet water changed at %v parity%d %dx", point, parity, density)
				}
				if density == 4 {
					local := connectionPoint{(float64(x) + .5) / 4, (float64(y) + .5) / 4}
					world := local.add(origin)
					if c != sampleConnectionMaterial(material, world) {
						t.Fatalf("absolute water phase at %v parity%d", point, parity)
					}
					phaseWitness = phaseWitness || c != sampleConnectionMaterial(material, world.add(connectionPoint{-71, 0}))
				}
			}
			if density == 4 && !phaseWitness {
				t.Fatal("fixture does not distinguish the incorrect quad-layout phase")
			}
			if density == 4 {
				for _, cell := range patch.Cells {
					if cell.Compound {
						continue
					}
					for _, point := range []image.Point{{48, 1}, {1, 40}, {94, 40}, {48, 78}} {
						local := connectionPoint{float64(point.X) + .125, float64(point.Y) + .125}
						x := (int(cell.Origin[0]-patch.CanvasOrigin[0]) + point.X) * density
						y := (int(cell.Origin[1]-patch.CanvasOrigin[1]) + point.Y) * density
						if actual.NRGBAAt(x, y) != sampleConnectionMaterial(material, local.add(cell.Origin)) {
							t.Fatalf("neighbor world water phase at %v + %v parity%d", cell.Origin, point, parity)
						}
					}
				}
			}
		}
	}
	files := 0
	if err := filepath.WalkDir(out, func(_ string, entry os.DirEntry, err error) error {
		if err == nil && !entry.IsDir() {
			files++
		}
		return err
	}); err != nil || files != 19 {
		t.Fatalf("bundle file count %d: %v", files, err)
	}
}

// Catch an ignored shared ground, incomplete registry, extra ground slots,
// aliases, and an off-by-one maximum before any bundle directory is created.
func TestCompoundWaterStonesTwelveSceneRegistryAndExactGrounds(t *testing.T) {
	path, m := hamletFixture(t)
	for _, id := range []string{"depot-summer", "depot-winter", "hq-summer", "hq-winter", "factory-winter", "mountain-summer", "mountain-winter", "mountain-single-summer", "mountain-single-winter", "water-stones"} {
		m.Scenes = append(m.Scenes, compoundSceneInput{ID: id, Artwork: m.Scenes[0].Artwork, Inset: .9})
	}
	m.Grounds["shared"] = m.Grounds["summer"]
	compoundTestWriteManifest(t, path, m)
	in, err := loadCompoundInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(in.Manifest.Scenes) != 12 || len(in.Artwork) != 12 || len(in.Grounds) != 3 || in.Grounds["shared"].Ref != m.Grounds["shared"] {
		t.Fatal("all twelve scenes and three exact grounds must load")
	}
	for _, tc := range []struct{ name, errorPart string }{
		{"missing-shared", "grounds must match exactly"},
		{"extra-shared", "grounds must match exactly"},
		{"unknown-ground", "unknown ground season"},
		{"summer-alias", "unsupported compound scene"},
		{"winter-alias", "unsupported compound scene"},
		{"runtime-alias", "unsupported compound scene"},
		{"duplicate", "duplicate compound scene"},
		{"thirteenth", "one to twelve"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bad := m
			bad.Grounds = map[string]connectionMaterial{}
			for k, v := range m.Grounds {
				bad.Grounds[k] = v
			}
			bad.Scenes = append([]compoundSceneInput(nil), m.Scenes...)
			switch tc.name {
			case "missing-shared":
				delete(bad.Grounds, "shared")
			case "extra-shared":
				bad.Scenes = bad.Scenes[:11]
			case "unknown-ground":
				bad.Grounds["ocean"] = bad.Grounds["shared"]
			case "summer-alias":
				bad.Scenes[11].ID = "water-stones-summer"
			case "winter-alias":
				bad.Scenes[11].ID = "water-stones-winter"
			case "runtime-alias":
				bad.Scenes[11].ID = "waterStones"
			case "duplicate":
				bad.Scenes[11].ID = "hamlet-summer"
			case "thirteenth":
				bad.Scenes = append(bad.Scenes, bad.Scenes[11])
			}
			compoundTestWriteManifest(t, path, bad)
			out := filepath.Join(filepath.Dir(path), tc.name)
			if err := run([]string{"-compound", path, "-out", out}, io.Discard); err == nil || !strings.Contains(err.Error(), tc.errorPart) {
				t.Fatalf("expected %q: %v", tc.errorPart, err)
			}
			if _, err := os.Stat(out); !os.IsNotExist(err) {
				t.Fatal("invalid manifest created output")
			}
		})
	}
}

func TestCompoundWaterStonesPinnedSourcesSurviveChanges(t *testing.T) {
	path, m := waterStonesFixture(t)
	in, err := loadCompoundInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{filepath.Base(path), m.Grounds["shared"].Source, m.Scenes[0].Artwork.Source} {
		if err := os.WriteFile(filepath.Join(filepath.Dir(path), name), []byte("changed after pinned read"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	b, err := buildCompoundBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]byte{"manifest.json": in.Bytes, "sources/ground-shared.png": in.Grounds["shared"].Bytes, "sources/water-stones.png": in.Artwork["water-stones"].Bytes}
	for _, a := range b.artifacts {
		if original, ok := want[a.name]; ok {
			if !bytes.Equal(a.data, original) {
				t.Fatalf("immutable snapshot changed: %s", a.name)
			}
			delete(want, a.name)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing snapshots: %v", want)
	}
}
