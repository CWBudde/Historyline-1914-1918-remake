package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Reuse the independent PNG/lattice checks, with literal house runtime IDs.
func hamletScenes() []mountainSceneWant {
	return []mountainSceneWant{
		{"hamlet-summer", "hamlet", "summer", 96, 80, []compoundTile{{Cell: "C", RuntimeID: "house", Origin: connectionPoint{0, 0}}}},
		{"hamlet-winter", "hamlet", "winter", 96, 80, []compoundTile{{Cell: "C", RuntimeID: "house_w", Origin: connectionPoint{0, 0}}}},
	}
}

func hamletFixture(t *testing.T) (string, compoundManifest) {
	t.Helper()
	path, m := mountainFixture(t)
	m.Scenes = []compoundSceneInput{
		{ID: "hamlet-summer", Artwork: m.Scenes[0].Artwork, Inset: .9},
		{ID: "hamlet-winter", Artwork: m.Scenes[1].Artwork, Inset: .9},
	}
	compoundTestWriteManifest(t, path, m)
	return path, m
}

func TestCompoundHamletLiteralLayoutAndRuntimeOrigins(t *testing.T) {
	for _, want := range hamletScenes() {
		t.Run(want.id, func(t *testing.T) {
			layout, err := compoundLayoutFor(want.id)
			if err != nil {
				t.Fatal(err)
			}
			mountainCheckLayout(t, layout, want)
			for parity, origin := range []connectionPoint{{0, 0}, {71, 40}} {
				patch := compoundPatchFor(layout, parity)
				mountainCheckRuntimePatch(t, patch, want, parity)
				if patch.SceneOrigin != origin || patch.CanvasOrigin != origin.add(connectionPoint{-71, -80}) || len(patch.Cells) != 7 {
					t.Fatalf("parity%d origins/cell count: %+v", parity, patch)
				}
			}
		})
	}
}

func TestCompoundHamletCLIActualPNGs(t *testing.T) {
	path, manifest := hamletFixture(t)
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
	if len(report.Scenes) != 2 || len(report.Grounds) != 2 {
		t.Fatalf("scene/ground count: %d/%d", len(report.Scenes), len(report.Grounds))
	}
	for i, want := range hamletScenes() {
		t.Run(want.id, func(t *testing.T) {
			scene := report.Scenes[i]
			if scene.ID != want.id || scene.Profile != compoundProfile || len(scene.Patches) != 2 {
				t.Fatalf("scene contract: %+v", scene)
			}
			mountainCheckLayout(t, scene.compoundLayout, want)
			if scene.Transform.Source != (dimensions{1024, 640}) || math.Abs(scene.Transform.Scale-.084375) > 1e-12 || math.Abs(scene.Transform.Offset[0]-4.8) > 1e-12 || scene.Transform.Offset[1] != 13 || scene.Guard.GuardedOpaque != 0 {
				t.Fatalf("whole-master transform/guard: %+v %+v", scene.Transform, scene.Guard)
			}
			material := mountainLoadPNG(t, filepath.Dir(path), manifest.Grounds[want.season].Source)
			groundID := "grass"
			if want.season == "winter" {
				groundID = "grass_w"
			}
			for _, density := range []int{2, 4} {
				full := mountainLoadPNG(t, out, fmt.Sprintf("%s-scene-%dx.png", want.id, density))
				tile := mountainLoadPNG(t, out, fmt.Sprintf("tiles/%s-%dx.png", want.tiles[0].RuntimeID, density))
				ground := mountainLoadPNG(t, out, fmt.Sprintf("ground/%s-%dx.png", groundID, density))
				if full.Bounds() != image.Rect(0, 0, 96*density, 80*density) || tile.Bounds() != full.Bounds() {
					t.Fatalf("scene/tile dimensions: %v %v", full.Bounds(), tile.Bounds())
				}
				assembled := image.NewNRGBA(full.Bounds())
				draw.Draw(assembled, tile.Bounds(), tile, image.Point{}, draw.Over)
				mountainEqualPNG(t, assembled, full, "ordinary decoded PNG assembly")
				mountainEqualPNG(t, mountainLoadPNG(t, out, fmt.Sprintf("%s-assembled-%dx.png", want.id, density)), full, "saved assembly")
				if bytes.Equal(full.Pix, ground.Pix) {
					t.Fatal("house artwork missing")
				}
				for y := range tile.Bounds().Dy() {
					for x := range tile.Bounds().Dx() {
						c := tile.NRGBAAt(x, y)
						if (mountainOwned(x, y, density) && c.A != 255) || (!mountainOwned(x, y, density) && c != (color.NRGBA{})) {
							t.Fatalf("tile ownership at %d,%d: %v", x, y, c)
						}
					}
				}
				for parity, origin := range []connectionPoint{{0, 0}, {71, 40}} {
					patch := scene.Patches[parity]
					mountainCheckRuntimePatch(t, patch, want, parity)
					if patch.SceneOrigin != origin || patch.CanvasOrigin != origin.add(connectionPoint{-71, -80}) {
						t.Fatalf("absolute world origins: %+v", patch)
					}
					actual := mountainLoadPNG(t, out, fmt.Sprintf("%s-patch-parity%d-%dx.png", want.id, parity, density))
					mountainEqualPNG(t, actual, mountainSurround(want, ground, map[string]*image.NRGBA{want.tiles[0].RuntimeID: tile}, density, parity), "independent decoded PNG surround")
					phaseWitness := false
					for _, point := range []image.Point{{48, 1}, {1, 40}, {94, 40}, {48, 78}} {
						x, y := point.X*density, point.Y*density
						c := full.NRGBAAt(x, y)
						if c != ground.NRGBAAt(x, y) || actual.NRGBAAt(x+71*density, y+80*density) != c {
							t.Fatalf("quiet ground changed at %v parity%d %dx", point, parity, density)
						}
						if density == 4 {
							// Normalizing a patch hides a wrong global translation. Check
							// the reported absolute world phase against runtime placement.
							local := connectionPoint{(float64(x) + .5) / 4, (float64(y) + .5) / 4}
							world := local.add(origin)
							if c != sampleConnectionMaterial(material, world) || c != sampleConnectionMaterial(material, local.add(patch.SceneOrigin)) {
								t.Fatalf("absolute world material phase at %v parity%d", point, parity)
							}
							phaseWitness = phaseWitness || c != sampleConnectionMaterial(material, world.add(connectionPoint{-71, 0}))
						}
					}
					if density == 4 && !phaseWitness {
						t.Fatal("fixture does not distinguish the incorrect quad-layout phase")
					}
				}
			}
		})
	}
}

func TestCompoundHamletElevenSceneBound(t *testing.T) {
	path, m := hamletFixture(t)
	for _, id := range []string{"depot-summer", "depot-winter", "hq-summer", "hq-winter", "factory-winter", "mountain-summer", "mountain-winter", "mountain-single-summer", "mountain-single-winter"} {
		m.Scenes = append(m.Scenes, compoundSceneInput{ID: id, Artwork: m.Scenes[0].Artwork, Inset: .9})
	}
	compoundTestWriteManifest(t, path, m)
	in, err := loadCompoundInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(in.Manifest.Scenes) != 11 || len(in.Artwork) != 11 {
		t.Fatal("all eleven fixed scenes were not loaded")
	}
	for _, tc := range []struct{ name, id, errorPart string }{
		{"thirteenth", "hamlet-summer", "one to twelve"},
		{"duplicate", "hamlet-summer", "duplicate compound scene"},
		{"unknown-season", "hamlet-autumn", "unsupported compound scene"},
		{"unknown-alias", "house-summer", "unsupported compound scene"},
		{"fabricated-cell", "houseC-summer", "unsupported compound scene"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bad := m
			bad.Scenes = append([]compoundSceneInput(nil), m.Scenes...)
			if tc.name == "thirteenth" {
				for range 2 {
					bad.Scenes = append(bad.Scenes, compoundSceneInput{ID: tc.id, Artwork: m.Scenes[0].Artwork, Inset: .9})
				}
			} else {
				bad.Scenes[1].ID = tc.id
			}
			compoundTestWriteManifest(t, path, bad)
			out := filepath.Join(filepath.Dir(path), tc.name)
			if err := run([]string{"-compound", path, "-out", out}, io.Discard); err == nil || !strings.Contains(err.Error(), tc.errorPart) {
				t.Fatalf("expected %q: %v", tc.errorPart, err)
			}
			if _, err := os.Stat(out); !os.IsNotExist(err) {
				t.Fatal("invalid input created output")
			}
		})
	}
}
