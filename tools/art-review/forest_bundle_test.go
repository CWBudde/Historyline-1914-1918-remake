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
	"image/png"
	"io/fs"
	"math"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func forestTestDescriptor(id string, neighbors [6]string) (string, string) {
	season, base := "summer", "grass"
	if strings.HasSuffix(id, "_w") {
		season, base = "winter", "grass_w"
	}
	if id == base {
		neighbors = [6]string{base, base, base, base, base, base}
	}
	descriptor := fmt.Sprintf("forest-v1\nseason=%s\ncenter=%s\n", season, id)
	for i, label := range []string{"NE", "SE", "S", "SW", "NW", "N"} {
		descriptor += label + "=" + neighbors[i] + "\n"
	}
	sum := sha256.Sum256([]byte(descriptor))
	return descriptor, hex.EncodeToString(sum[:])
}
func TestForestCLIActualPNGsAndFixtureTopology(t *testing.T) {
	path := forestTestFixture(t)
	out := filepath.Join(filepath.Dir(path), "bundle")
	cmd := exec.Command("go", "run", ".", "-forest", path, "-out", out)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("forest CLI %v: %s", err, b)
	}
	var r forestReport
	if err := json.Unmarshal(groundTestRead(t, filepath.Join(out, "report.json")), &r); err != nil {
		t.Fatal(err)
	}
	if r.SchemaVersion != 1 || len(r.Scenes) != 16 || len(r.Template) != 36 || len(r.Grounds) != 2 || len(r.Atlases) != 2 {
		t.Fatal("bounded report roster")
	}
	if len(r.Profiles) != 20 || len(r.Exports) != 72 {
		t.Fatal("missing profiles/scenes")
	}
	files := 0
	if err := filepath.WalkDir(out, func(_ string, entry fs.DirEntry, err error) error {
		if err == nil && !entry.IsDir() {
			files++
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if files != 79 {
		t.Fatalf("bundle files: %d; want 72 renders + 4 sources + manifest/report/gallery", files)
	}
	t.Logf("forest: %d profiles, %d scenes, %d PNGs", len(r.Profiles), len(r.Scenes), len(r.Exports))
	groundTestSnapshot(t, out, r.ManifestSnapshot, groundTestRead(t, path), r.ManifestSHA256)
	for _, group := range []map[string]connectionMaterialReport{r.Grounds, r.Atlases} {
		for _, source := range group {
			groundTestSnapshot(t, out, source.RetainedSource, groundTestRead(t, filepath.Join(filepath.Dir(path), source.Source)), source.SourceSHA256)
		}
	}
	actual := map[string]*image.NRGBA{}
	for _, output := range r.Exports {
		data := groundTestRead(t, filepath.Join(out, output.Path))
		sum := sha256.Sum256(data)
		if output.SHA256 != hex.EncodeToString(sum[:]) {
			t.Fatalf("wrong hash %s", output.Path)
		}
		im := groundTestLoad(t, filepath.Join(out, output.Path))
		actual[output.Path] = im
		if output.Dimensions != (dimensions{im.Bounds().Dx(), im.Bounds().Dy()}) {
			t.Fatal("wrong dimensions")
		}
	}
	profiles := map[string]forestProfileReport{}
	for _, profile := range r.Profiles {
		descriptor, key := forestTestDescriptor(profile.RuntimeID, profile.NeighborIDs)
		if profile.ContextKey != key || profile.Descriptor != descriptor {
			t.Fatal("wrong profile identity")
		}
		if _, duplicate := profiles[key]; duplicate {
			t.Fatal("duplicate profile")
		}
		profiles[key] = profile
		if len(profile.Outputs) != 2 {
			t.Fatal("profile density roster")
		}
		for _, output := range profile.Outputs {
			im := actual[output.Path]
			scale := im.Bounds().Dx() / 96
			if im.Bounds().Dy() != 80*scale || scale != 2 && scale != 4 {
				t.Fatal("tile size")
			}
			for y := range im.Bounds().Dy() {
				for x := range im.Bounds().Dx() {
					p := im.NRGBAAt(x, y)
					if groundTestOwned(x, y, scale) {
						if p.A != 255 {
							t.Fatal("owned hole")
						}
					} else if p != (color.NRGBA{}) {
						t.Fatal("outside ownership")
					}
				}
			}
		}
	}
	for _, id := range []string{"grass", "grass_w", "forest", "forest_w"} {
		for _, scale := range []int{2, 4} {
			if actual[fmt.Sprintf("tiles/%s-%dx.png", id, scale)] == nil {
				t.Fatal("missing canonical tile")
			}
		}
	}
	// Exact existing opaque base PNG bytes, independent of forest context.
	for _, season := range []string{"summer", "winter"} {
		id := "grass"
		if season == "winter" {
			id = "grass_w"
		}
		source := groundTestLoad(t, filepath.Join(filepath.Dir(path), r.Grounds[season].Source))
		existing, err := exportConnectionTiles(connectionInputs{manifest: connectionManifest{Geometry: sharedConnectionGeometry()}, materials: map[string]*image.NRGBA{"grass": source}}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, scale := range []int{2, 4} {
			var encoded bytes.Buffer
			if err := png.Encode(&encoded, existing[scale]); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(encoded.Bytes(), groundTestRead(t, filepath.Join(out, fmt.Sprintf("tiles/%s-%dx.png", id, scale)))) {
				t.Fatal("base PNG bytes changed")
			}
		}
	}
	seen := map[string]bool{}
	for _, scene := range r.Scenes {
		if seen[scene.ID] || len(scene.Cells) != 7 || len(scene.Sites) != 19 || len(scene.Outputs) != 2 {
			t.Fatal("scene roster")
		}
		seen[scene.ID] = true
		if scene.CenterOrigin != (connectionPoint{float64(71 * scene.CenterLineParity), float64(40 * scene.CenterLineParity)}) {
			t.Fatal("fake parity origin")
		}
		kind, base := "forest", "grass"
		if strings.HasPrefix(scene.ID, "winter-") {
			kind, base = "forest_w", "grass_w"
		}
		world := map[[2]int]string{}
		for _, site := range scene.Sites {
			expected := base
			relative := site.Origin.sub(scene.CenterOrigin)
			if strings.Contains(scene.ID, "-homogeneous-") || strings.Contains(scene.ID, "-isolated-") && relative == (connectionPoint{}) || strings.Contains(scene.ID, "-frontier-") && relative[0] >= 0 || strings.Contains(scene.ID, "-clearing-") && relative != (connectionPoint{}) {
				expected = kind
			}
			if site.RuntimeID != expected {
				t.Fatal("wrong scene case")
			}
			world[[2]int{int(site.Origin[0]), int(site.Origin[1])}] = site.RuntimeID
		}
		minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
		clearingOrientations := [6]bool{}
		for i, cell := range scene.Cells {
			d := [2]int{}
			if i > 0 {
				d = groundTestDeltas[i-1]
			}
			if cell.Origin != scene.CenterOrigin.add(connectionPoint{float64(d[0]), float64(d[1])}) {
				t.Fatal("wrong neighbor mapping")
			}
			if cell.RuntimeID != world[[2]int{int(cell.Origin[0]), int(cell.Origin[1])}] {
				t.Fatal("wrong cell ID")
			}
			for n, delta := range groundTestDeltas {
				if cell.NeighborIDs[n] != world[[2]int{int(cell.Origin[0]) + delta[0], int(cell.Origin[1]) + delta[1]}] {
					t.Fatal("context not derived from 19 real sites")
				}
			}
			_, key := forestTestDescriptor(cell.RuntimeID, cell.NeighborIDs)
			if cell.ContextKey != key {
				t.Fatal("cell context key")
			}
			if strings.Contains(scene.ID, "-clearing-") {
				if i == 0 {
					if cell.RuntimeID != base || cell.NeighborIDs != ([6]string{kind, kind, kind, kind, kind, kind}) {
						t.Fatal("clearing center must retain its actual six forest neighbors")
					}
				} else {
					basePort := [6]int{3, 4, 5, 0, 1, 2}[i-1]
					clearingOrientations[basePort] = true
					for port, neighbor := range cell.NeighborIDs {
						expected := kind
						if port == basePort {
							expected = base
						}
						if neighbor != expected {
							t.Fatal("clearing forest cell must have exactly the reciprocal oriented base neighbor")
						}
					}
					for _, output := range profiles[cell.ContextKey].Outputs {
						im := actual[output.Path]
						scale := im.Bounds().Dx() / 96
						baseTile := actual[fmt.Sprintf("tiles/%s-%dx.png", base, scale)]
						quiet := 0
						for y := range 80 * scale {
							for x := range 96 * scale {
								if !groundTestOwned(x, y, scale) {
									continue
								}
								px := (float64(x)+.5)/float64(scale) + cell.Origin[0] - scene.CenterOrigin[0]
								py := (float64(y)+.5)/float64(scale) + cell.Origin[1] - scene.CenterOrigin[1]
								if groundTestDistance(px, py) <= 4 {
									quiet++
									if im.NRGBAAt(x, y) != baseTile.NRGBAAt(x, y) {
										t.Fatalf("clearing %s port %d final %dx quiet band", scene.ID, basePort, scale)
									}
								}
							}
						}
						if quiet < 100 {
							t.Fatal("clearing oriented quiet band not exercised")
						}
					}
				}
			}
			minX = math.Min(minX, cell.Origin[0])
			minY = math.Min(minY, cell.Origin[1])
			maxX = math.Max(maxX, cell.Origin[0]+96)
			maxY = math.Max(maxY, cell.Origin[1]+80)
		}
		if strings.Contains(scene.ID, "-clearing-") && clearingOrientations != ([6]bool{true, true, true, true, true, true}) {
			t.Fatal("clearing does not cover all six oriented base neighbors")
		}
		if scene.CanvasOrigin != (connectionPoint{minX, minY}) {
			t.Fatal("wrong canvas origin")
		}
		for _, scale := range []int{2, 4} {
			reassembled := image.NewNRGBA(image.Rect(0, 0, int(maxX-minX)*scale, int(maxY-minY)*scale))
			for _, cell := range scene.Cells {
				profile := profiles[cell.ContextKey]
				var tile *image.NRGBA
				for _, output := range profile.Outputs {
					if actual[output.Path].Bounds().Dx() == 96*scale {
						tile = actual[output.Path]
					}
				}
				at := image.Pt(int(cell.Origin[0]-minX)*scale, int(cell.Origin[1]-minY)*scale)
				draw.Draw(reassembled, tile.Bounds().Add(at), tile, image.Point{}, draw.Over)
			}
			var saved *image.NRGBA
			for _, output := range scene.Outputs {
				if actual[output.Path].Bounds().Dx() == reassembled.Bounds().Dx() {
					saved = actual[output.Path]
				}
			}
			if saved == nil || !bytes.Equal(saved.Pix, reassembled.Pix) {
				t.Fatalf("actual PNG Over reassembly %s %dx", scene.ID, scale)
			}
			// No holes anywhere covered by any of the seven literal polygons.
			for y := range reassembled.Bounds().Dy() {
				for x := range reassembled.Bounds().Dx() {
					owned := false
					for _, cell := range scene.Cells {
						owned = owned || groundTestOwned(x-int(cell.Origin[0]-minX)*scale, y-int(cell.Origin[1]-minY)*scale, scale)
					}
					if owned && saved.NRGBAAt(x, y).A != 255 {
						t.Fatal("patch ownership gap")
					}
				}
			}
		}
	}
	for _, season := range []string{"summer", "winter"} {
		for _, scenario := range []string{"isolated", "homogeneous", "frontier", "clearing"} {
			for parity := range 2 {
				if !seen[fmt.Sprintf("%s-%s-parity%d", season, scenario, parity)] {
					t.Fatal("missing literal scene")
				}
			}
		}
	}
	// Actual report and gallery references must correspond to saved artifacts.
	gallery := string(groundTestRead(t, filepath.Join(out, "index.html")))
	for _, output := range r.Exports {
		if !strings.Contains(gallery, `src="`+output.Path+`"`) {
			t.Fatalf("missing gallery image %s", output.Path)
		}
	}
	if err := run([]string{"-forest", path, "-out", out}, &bytes.Buffer{}); err == nil {
		t.Fatal("overwrote existing output")
	}
}
