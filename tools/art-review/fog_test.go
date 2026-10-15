package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// These report views are deliberately independent of the exporter's structs.
type fogTestCell struct {
	Label      string `json:"label"`
	Origin     [2]int `json:"origin"`
	RuntimeID  string `json:"runtime_id"`
	Visibility string `json:"visibility"`
}
type fogTestScene struct {
	ID      string        `json:"id"`
	Case    string        `json:"case"`
	Season  string        `json:"season"`
	Parity  int           `json:"center_line_parity"`
	Center  [2]int        `json:"center_origin"`
	Canvas  [2]int        `json:"canvas_origin"`
	Cells   []fogTestCell `json:"cells"`
	Outputs []packOutput  `json:"outputs"`
}
type fogTestReport struct {
	SchemaVersion     int                                 `json:"schema_version"`
	ManifestSHA256    string                              `json:"manifest_sha256"`
	ManifestSnapshot  string                              `json:"manifest_snapshot"`
	Geometry          connectionGeometry                  `json:"geometry"`
	Materials         map[string]connectionMaterialReport `json:"materials"`
	Scenes            []fogTestScene                      `json:"scenes"`
	Exports           []packOutput                        `json:"exports"`
	Profile           string                              `json:"profile"`
	Ownership         string                              `json:"ownership"`
	ResamplingChain   []string                            `json:"resampling_chain"`
	ConcealmentPolicy string                              `json:"concealment_policy"`
	Acceptance        string                              `json:"acceptance"`
}

func fogTestWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}
func fogTestJSON(t *testing.T, path string, v any) {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	fogTestWrite(t, path, b)
}
func fogTestPNG(t *testing.T, path string, im image.Image) connectionMaterial {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	fogTestWrite(t, path, b.Bytes())
	sum := sha256.Sum256(b.Bytes())
	return connectionMaterial{filepath.Base(path), hex.EncodeToString(sum[:])}
}
func fogTestFixture(t *testing.T) (string, map[string]any) {
	t.Helper()
	dir := t.TempDir()
	materials := map[string]connectionMaterial{}
	for i, name := range []string{"unseen", "summer", "winter"} {
		im := image.NewNRGBA(image.Rect(0, 0, 384, 320))
		for y := range 320 {
			for x := range 384 {
				im.SetNRGBA(x, y, color.NRGBA{uint8(x*3 + y + i*53), uint8(y*7 + x + i*19), uint8(x ^ y + i*41), 255})
			}
		}
		materials[name] = fogTestPNG(t, filepath.Join(dir, name+".png"), im)
	}
	m := map[string]any{"schema_version": 1, "materials": materials}
	path := filepath.Join(dir, "manifest.json")
	fogTestJSON(t, path, m)
	return path, m
}

func TestFogCLIActualBundle(t *testing.T) {
	path, manifest := fogTestFixture(t)
	out := filepath.Join(filepath.Dir(path), "bundle")
	// Exercise the actual CLI, including main's exit behavior and flag routing.
	cmd := exec.Command("go", "run", ".", "-fog", path, "-out", out)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fog CLI: %v\n%s", err, b)
	}
	data, err := os.ReadFile(filepath.Join(out, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report fogTestReport
	if err = json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	wantGeometry := connectionGeometry{96, 80,
		[]connectionPoint{{25, 0}, {71, 0}, {96, 40}, {71, 80}, {25, 80}, {0, 40}},
		[]connectionPoint{{71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}},
		[]connectionPoint{{83.5, 20}, {83.5, 60}, {48, 80}, {12.5, 60}, {12.5, 20}, {48, 0}}}
	if report.SchemaVersion != 1 || !reflect.DeepEqual(report.Geometry, wantGeometry) {
		t.Fatalf("geometry/version: %+v", report)
	}
	if !strings.HasPrefix(report.Profile, "fog-v1:") || report.Ownership == "" || len(report.ResamplingChain) == 0 || !strings.Contains(report.ConcealmentPolicy, "opaque") || !strings.Contains(report.Acceptance, "Provisional") || !strings.Contains(report.Acceptance, "runtime") {
		t.Fatalf("missing bounded policy: %+v", report)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fogTestSnapshot(t, out, report.ManifestSnapshot, original, report.ManifestSHA256)
	materials := manifest["materials"].(map[string]connectionMaterial)
	if len(report.Materials) != 3 {
		t.Fatalf("materials: %v", report.Materials)
	}
	for name, ref := range materials {
		src := report.Materials[name]
		if src.connectionMaterial != ref || src.Dimensions != (dimensions{384, 320}) {
			t.Fatalf("source provenance: %+v", src)
		}
		original, err = os.ReadFile(filepath.Join(filepath.Dir(path), ref.Source))
		if err != nil {
			t.Fatal(err)
		}
		fogTestSnapshot(t, out, src.RetainedSource, original, src.SourceSHA256)
	}
	if len(report.Scenes) != 6 {
		t.Fatalf("scenes: %d", len(report.Scenes))
	}
	for i, scene := range report.Scenes {
		caseName := []string{"all-hidden", "summer-frontier", "winter-frontier"}[i/2]
		parity := i % 2
		fogTestSceneContract(t, scene, caseName, parity)
		for _, density := range []int{2, 4} {
			fogTestScenePixels(t, out, scene, density)
		}
	}
	for _, density := range []int{2, 4} {
		fog := fogTestLoad(t, out, fmt.Sprintf("tiles/unseen-%dx.png", density))
		fogTestOwnershipAndConcealment(t, fog, density)
		if density == 4 {
			source := fogTestLoad(t, filepath.Dir(path), materials["unseen"].Source)
			for parity := range 2 {
				// Actual runtime world origins, including negative lines, verify
				// source identity and shared phase without reusing the scene builder.
				for _, delta := range []image.Point{{0, 0}, {71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}} {
					origin := image.Pt(71*(parity%2), 40*parity).Add(delta)
					for _, point := range []image.Point{{48, 1}, {1, 40}, {94, 40}, {48, 78}, {48, 40}} {
						x, y := point.X*density, point.Y*density
						world := connectionPoint{float64(origin.X) + (float64(x)+.5)/4, float64(origin.Y) + (float64(y)+.5)/4}
						if got, want := fog.NRGBAAt(x, y), sampleConnectionMaterial(source, world); got != want {
							t.Fatalf("unseen source/phase at world %v: %v, want %v", world, got, want)
						}
					}
				}
			}
		}
		for _, pair := range [][2]string{{"summer", "grass"}, {"winter", "grass_w"}} {
			ground := fogTestLoad(t, out, fmt.Sprintf("ground/%s-%dx.png", pair[1], density))
			fogTestOwnershipAndConcealment(t, ground, density)
			// Compare encoded ground bytes against the established, unchanged renderer.
			source := fogTestLoad(t, filepath.Dir(path), materials[pair[0]].Source)
			exports, e := exportConnectionTiles(connectionInputs{manifest: connectionManifest{Geometry: wantGeometry}, materials: map[string]*image.NRGBA{"grass": source}}, nil, nil)
			if e != nil {
				t.Fatal(e)
			}
			var want bytes.Buffer
			if e = png.Encode(&want, exports[density]); e != nil {
				t.Fatal(e)
			}
			got, e := os.ReadFile(filepath.Join(out, fmt.Sprintf("ground/%s-%dx.png", pair[1], density)))
			if e != nil || !bytes.Equal(got, want.Bytes()) {
				t.Fatalf("changed ground export %s %dx: %v", pair[0], density, e)
			}
		}
	}
	// Rendered exports and separately pinned retained sources cover every PNG.
	if len(report.Exports) != 18 {
		t.Fatalf("rendered PNG report: %d, want 18", len(report.Exports))
	}
	listed := map[string]bool{}
	for _, source := range report.Materials {
		listed[source.RetainedSource] = true
	}
	for _, output := range report.Exports {
		if listed[output.Path] {
			t.Fatalf("duplicate output: %s", output.Path)
		}
		listed[output.Path] = true
		b, e := os.ReadFile(filepath.Join(out, output.Path))
		if e != nil {
			t.Fatal(e)
		}
		sum := sha256.Sum256(b)
		im := fogTestLoad(t, out, output.Path)
		if hex.EncodeToString(sum[:]) != output.SHA256 || output.Dimensions != (dimensions{im.Bounds().Dx(), im.Bounds().Dy()}) {
			t.Fatalf("PNG metadata mismatch: %+v", output)
		}
	}
	files := 0
	err = filepath.WalkDir(out, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			files++
			rel, e := filepath.Rel(out, path)
			if e != nil {
				return e
			}
			if strings.HasSuffix(rel, ".png") && !listed[filepath.ToSlash(rel)] {
				t.Errorf("unreported PNG %s", rel)
			}
		}
		return nil
	})
	if err != nil || files != 24 {
		t.Fatalf("bundle files: %d, err %v", files, err)
	}
	html, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, scene := range report.Scenes {
		for _, output := range scene.Outputs {
			if !bytes.Contains(html, []byte(output.Path)) {
				t.Fatalf("gallery missing %s", output.Path)
			}
		}
	}
	if !bytes.Contains(html, []byte("<img")) || bytes.Contains(html, []byte(filepath.Dir(path))) {
		t.Fatal("gallery is not self-contained relative auto-rendering HTML")
	}
}

func fogTestSnapshot(t *testing.T, out, path string, want []byte, hash string) {
	t.Helper()
	got, err := os.ReadFile(filepath.Join(out, path))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(got)
	if !bytes.Equal(got, want) || hex.EncodeToString(sum[:]) != hash {
		t.Fatalf("snapshot mismatch: %s", path)
	}
}
func fogTestLoad(t *testing.T, out, path string) *image.NRGBA {
	t.Helper()
	im, err := loadPNG(filepath.Join(out, path))
	if err != nil {
		t.Fatal(err)
	}
	return im
}
func fogTestSceneContract(t *testing.T, scene fogTestScene, caseName string, parity int) {
	t.Helper()
	// Runtime formula x=142*col+71*(line mod 2), y=40*line, col=0.
	center := [2]int{142*0 + 71*(parity%2), 40 * parity}
	if scene.Case != caseName || scene.ID != fmt.Sprintf("%s-parity%d", caseName, parity) || scene.Parity != parity || scene.Center != center || scene.Canvas != ([2]int{center[0] - 71, center[1] - 80}) || len(scene.Cells) != 7 || len(scene.Outputs) != 2 {
		t.Fatalf("scene contract: %+v", scene)
	}
	season := "none"
	groundID := "unseen"
	if caseName == "summer-frontier" {
		season, groundID = "summer", "grass"
	}
	if caseName == "winter-frontier" {
		season, groundID = "winter", "grass_w"
	}
	if scene.Season != season {
		t.Fatalf("scene season: %s", scene.Season)
	}
	deltas := [][2]int{{0, 0}, {71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}}
	labels := []string{"C", "NE", "SE", "S", "SW", "NW", "N"}
	for i, cell := range scene.Cells {
		wantID, visibility := "unseen", "never-seen"
		if i >= 4 && caseName != "all-hidden" {
			wantID, visibility = groundID, "visible"
		}
		wantOrigin := [2]int{center[0] + deltas[i][0], center[1] + deltas[i][1]}
		if cell.Label != labels[i] || cell.Origin != wantOrigin || cell.RuntimeID != wantID || cell.Visibility != visibility {
			t.Fatalf("cell %d: %+v, want %s at %v", i, cell, wantID, wantOrigin)
		}
		line := cell.Origin[1] / 40
		positiveParity := (line%2 + 2) % 2
		column := float64(cell.Origin[0]-71*positiveParity) / 142
		if column != math.Trunc(column) {
			t.Fatalf("cell outside runtime lattice: %+v", cell)
		}
	}
}
func fogTestScenePixels(t *testing.T, out string, scene fogTestScene, density int) {
	t.Helper()
	name := fmt.Sprintf("scenes/%s-%dx.png", scene.ID, density)
	got := fogTestLoad(t, out, name)
	want := image.NewNRGBA(image.Rect(0, 0, 238*density, 240*density))
	if got.Bounds() != want.Bounds() {
		t.Fatalf("scene dimensions: %v", got.Bounds())
	}
	// Derive absolute positions from runtime line/column, independently of report.
	center := image.Pt(71*(scene.Parity%2), 40*scene.Parity)
	minimum := center.Sub(image.Pt(71, 80))
	for i, delta := range []image.Point{{0, 0}, {71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}} {
		path := fmt.Sprintf("tiles/unseen-%dx.png", density)
		if i >= 4 && scene.Case != "all-hidden" {
			id := "grass"
			if scene.Case == "winter-frontier" {
				id = "grass_w"
			}
			path = fmt.Sprintf("ground/%s-%dx.png", id, density)
		}
		tile := fogTestLoad(t, out, path)
		origin := center.Add(delta).Sub(minimum).Mul(density)
		draw.Draw(want, tile.Bounds().Add(origin), tile, image.Point{}, draw.Over)
	}
	if !bytes.Equal(got.Pix, want.Pix) {
		t.Fatalf("%s differs from ordinary decoded-PNG Over assembly", name)
	}
	// Exactly one opaque owner throughout the seven-cell union, no seams or holes.
	for y := range got.Bounds().Dy() {
		for x := range got.Bounds().Dx() {
			owners := 0
			for _, d := range []image.Point{{0, 0}, {71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}} {
				local := image.Pt(x, y).Sub(center.Add(d).Sub(minimum).Mul(density))
				if fogTestOwned(local.X, local.Y, density) {
					owners++
				}
			}
			c := got.NRGBAAt(x, y)
			if owners > 1 || (owners == 1 && c.A != 255) || (owners == 0 && c != (color.NRGBA{})) {
				t.Fatalf("%s union ownership at %d,%d: owners %d, %v", name, x, y, owners, c)
			}
		}
	}
}

// Integer half-planes for the literal six vertices, with complementary top/left ties.
func fogTestOwned(x, y, density int) bool {
	xx, yy, d := 2*x+1, 2*y+1, 2*density
	return yy >= 0 && yy < 80*d && 40*xx+25*yy >= 1000*d && 40*xx-25*yy < 2840*d && 40*xx+25*yy < 4840*d && 40*xx-25*yy >= -1000*d
}
func fogTestOwnershipAndConcealment(t *testing.T, fog *image.NRGBA, density int) {
	t.Helper()
	if fog.Bounds() != image.Rect(0, 0, 96*density, 80*density) {
		t.Fatalf("tile dimensions: %v", fog.Bounds())
	}
	underlays := []*image.NRGBA{image.NewNRGBA(fog.Bounds()), image.NewNRGBA(fog.Bounds())}
	for y := range fog.Bounds().Dy() {
		for x := range fog.Bounds().Dx() {
			underlays[0].SetNRGBA(x, y, color.NRGBA{uint8(255 * ((x + y) % 2)), 0, 255, 255})
			underlays[1].SetNRGBA(x, y, color.NRGBA{0, uint8(255 * ((x/3 + y/5) % 2)), 0, 255})
		}
	}
	composites := []*image.NRGBA{image.NewNRGBA(fog.Bounds()), image.NewNRGBA(fog.Bounds())}
	for i := range underlays {
		draw.Draw(composites[i], fog.Bounds(), underlays[i], image.Point{}, draw.Src)
		draw.Draw(composites[i], fog.Bounds(), fog, image.Point{}, draw.Over)
	}
	for y := range fog.Bounds().Dy() {
		for x := range fog.Bounds().Dx() {
			c := fog.NRGBAAt(x, y)
			if fogTestOwned(x, y, density) {
				if c.A != 255 || composites[0].NRGBAAt(x, y) != c || composites[1].NRGBAAt(x, y) != c {
					t.Fatalf("underlying information leaked at %d,%d: %v", x, y, c)
				}
			} else if c != (color.NRGBA{}) || composites[0].NRGBAAt(x, y) != underlays[0].NRGBAAt(x, y) || composites[1].NRGBAAt(x, y) != underlays[1].NRGBAAt(x, y) {
				t.Fatalf("outside ownership changed at %d,%d: %v", x, y, c)
			}
		}
	}
}

func TestFogStrictManifest(t *testing.T) {
	path, manifest := fogTestFixture(t)
	valid, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"null": "null", "array": "[]", "empty": "{}", "schema-null": strings.Replace(string(valid), `"schema_version":1`, `"schema_version":null`, 1),
		"schema-string":      strings.Replace(string(valid), `"schema_version":1`, `"schema_version":"1"`, 1),
		"schema-version":     strings.Replace(string(valid), `"schema_version":1`, `"schema_version":2`, 1),
		"schema-fraction":    strings.Replace(string(valid), `"schema_version":1`, `"schema_version":1.5`, 1),
		"root-alias":         strings.Replace(string(valid), `"schema_version"`, `"Schema_Version"`, 1),
		"materials-alias":    strings.Replace(string(valid), `"materials"`, `"Materials"`, 1),
		"unknown":            strings.Replace(string(valid), `"schema_version":1`, `"schema_version":1,"unknown":0`, 1),
		"duplicate":          strings.Replace(string(valid), `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1),
		"source-duplicate":   strings.Replace(string(valid), `"source":"summer.png"`, `"source":"summer.png","source":"summer.png"`, 1),
		"source-alias":       strings.Replace(string(valid), `"source":"summer.png"`, `"Source":"summer.png"`, 1),
		"hash-alias":         strings.Replace(string(valid), `"source_sha256"`, `"Source_SHA256"`, 1),
		"source-null":        strings.Replace(string(valid), `"source":"summer.png"`, `"source":null`, 1),
		"source-wrongtype":   strings.Replace(string(valid), `"source":"summer.png"`, `"source":[]`, 1),
		"source-empty":       strings.Replace(string(valid), `"source":"summer.png"`, `"source":""`, 1),
		"source-missing":     strings.Replace(string(valid), `"source":"summer.png",`, ``, 1),
		"ref-extra":          strings.Replace(string(valid), `"source":"summer.png"`, `"source":"summer.png","note":"x"`, 1),
		"ref-null":           strings.Replace(string(valid), `"summer":{`, `"summer":null,"unused":{`, 1),
		"missing-material":   strings.Replace(string(valid), `"summer":`, `"autumn":`, 1),
		"material-alias":     strings.Replace(string(valid), `"unseen":`, `"Unseen":`, 1),
		"materials-null":     `{"schema_version":1,"materials":null}`,
		"materials-array":    `{"schema_version":1,"materials":[]}`,
		"extra-material":     strings.Replace(string(valid), `"materials":{`, `"materials":{"extra":null,`, 1),
		"material-duplicate": strings.Replace(string(valid), `"unseen":{`, `"unseen":null,"unseen":{`, 1),
		"material-null":      `{"schema_version":1,"materials":{"unseen":null,"summer":{},"winter":{}}}`,
		"extra-json":         string(valid) + ` {}`,
	}
	hash := manifest["materials"].(map[string]connectionMaterial)["summer"].SourceSHA256
	for name, replacement := range map[string]string{"hash-null": "null", "hash-wrongtype": "42", "hash-empty": `""`} {
		cases[name] = strings.Replace(string(valid), `"source_sha256":"`+hash+`"`, `"source_sha256":`+replacement, 1)
	}
	cases["hash-missing"] = strings.Replace(string(valid), `,"source_sha256":"`+hash+`"`, "", 1)
	for name, bad := range cases {
		t.Run(name, func(t *testing.T) {
			fogTestWrite(t, path, []byte(bad))
			out := filepath.Join(filepath.Dir(path), "out-"+name)
			err := run([]string{"-fog", path, "-out", out}, io.Discard)
			if err == nil {
				t.Fatalf("accepted invalid manifest: %s", bad)
			}
			if _, e := os.Lstat(out); !os.IsNotExist(e) {
				t.Fatalf("invalid input created output: %v", e)
			}
		})
	}
}

func TestFogLoadedSnapshotsAndSourceIndependence(t *testing.T) {
	path, _ := fogTestFixture(t)
	in, err := loadFogInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	// Once loaded, both provenance and rendering use these exact source bytes.
	fogTestWrite(t, path, []byte("changed after loading"))
	for _, source := range in.Sources {
		fogTestWrite(t, filepath.Join(filepath.Dir(path), source.Ref.Source), []byte("changed after loading"))
	}
	first, err := buildFogBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	retained := map[string][]byte{}
	for _, artifact := range first.artifacts {
		retained[artifact.name] = artifact.data
	}
	if !bytes.Equal(retained["manifest.json"], in.Bytes) {
		t.Fatal("manifest was reread")
	}
	for id, source := range in.Sources {
		if !bytes.Equal(retained[first.report.Materials[id].RetainedSource], source.Bytes) {
			t.Fatalf("source %s was reread", id)
		}
	}
	// Replace both terrain sources; unseen assets and all-hidden scenes must
	// remain byte-for-byte identical despite high-contrast underlying patterns.
	other := image.NewNRGBA(image.Rect(0, 0, 384, 320))
	for y := range 320 {
		for x := range 384 {
			other.SetNRGBA(x, y, color.NRGBA{uint8(255 * ((x + y) % 2)), 255, uint8(255 * ((x/3 + y/5) % 2)), 255})
		}
	}
	ref := fogTestPNG(t, filepath.Join(filepath.Dir(path), "other.png"), other)
	source, err := compoundReadSource(filepath.Dir(path), ref, true)
	if err != nil {
		t.Fatal(err)
	}
	in.Sources["summer"], in.Sources["winter"] = source, source
	second, err := buildFogBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	for _, artifact := range second.artifacts {
		if strings.HasPrefix(artifact.name, "tiles/unseen-") || strings.HasPrefix(artifact.name, "scenes/all-hidden-") {
			if !bytes.Equal(artifact.data, retained[artifact.name]) {
				t.Fatalf("underlying terrain changed %s", artifact.name)
			}
		}
	}
}

func TestFogOversizedSourceRejectsBeforeDecode(t *testing.T) {
	path, m := fogTestFixture(t)
	materials := m["materials"].(map[string]connectionMaterial)
	data, err := os.ReadFile(filepath.Join(filepath.Dir(path), materials["unseen"].Source))
	if err != nil {
		t.Fatal(err)
	}
	// Valid IHDR with an over-budget size; reject before decoding its image body.
	binary.BigEndian.PutUint32(data[16:20], 8193)
	binary.BigEndian.PutUint32(data[20:24], 8192)
	binary.BigEndian.PutUint32(data[29:33], crc32.ChecksumIEEE(data[12:29]))
	fogTestWrite(t, filepath.Join(filepath.Dir(path), "huge.png"), data)
	sum := sha256.Sum256(data)
	materials["unseen"] = connectionMaterial{"huge.png", hex.EncodeToString(sum[:])}
	fogTestJSON(t, path, m)
	out := filepath.Join(filepath.Dir(path), "huge-out")
	err = run([]string{"-fog", path, "-out", out}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "64 megapixels") {
		t.Fatalf("oversized input: %v", err)
	}
	if _, err = os.Lstat(out); !os.IsNotExist(err) {
		t.Fatal("oversized input created output")
	}
}

func TestFogModeConflicts(t *testing.T) {
	for _, other := range []string{"reference", "candidate", "reference-platform", "neighborhood", "connected", "compound", "shore", "mixed", "pack", "resampling"} {
		for _, value := range []string{"", "x"} {
			t.Run(other+"/"+value, func(t *testing.T) {
				var log bytes.Buffer
				err := run([]string{"-fog", "manifest.json", "-out", "out", "-" + other + "=" + value}, &log)
				if err == nil || !strings.Contains(err.Error(), "excludes all other flags") {
					t.Fatalf("mode conflict %s=%q: %v", other, value, err)
				}
			})
		}
	}
	for _, args := range [][]string{{"-fog=", "-out", "out"}, {"-fog", "x"}, {"-fog", "x", "-out", "out", "extra"}, {"-fog", "x", "-out="}, {"-fog=", "-reference", "x", "-candidate", "y", "-out", "z"}} {
		if err := run(args, io.Discard); err == nil || !strings.Contains(err.Error(), "-fog requires") {
			t.Fatalf("accepted invalid CLI: %v: %v", args, err)
		}
	}
}

func TestFogSourcesAndFreshOutput(t *testing.T) {
	path, m := fogTestFixture(t)
	materials := m["materials"].(map[string]connectionMaterial)
	originalRef := materials["unseen"]
	for _, tc := range []struct{ name, path, hash string }{
		{"stale-pin", originalRef.Source, strings.Repeat("0", 64)},
		{"uppercase-pin", originalRef.Source, strings.ToUpper(originalRef.SourceSHA256)},
		{"short-pin", originalRef.Source, "aa"},
		{"empty-pin", originalRef.Source, ""},
		{"absolute", filepath.Join(filepath.Dir(path), originalRef.Source), originalRef.SourceSHA256},
		{"traversal", "../unseen.png", originalRef.SourceSHA256},
		{"dot", "./unseen.png", originalRef.SourceSHA256},
		{"backslash", `dir\unseen.png`, originalRef.SourceSHA256},
		{"original", "original/unseen.png", originalRef.SourceSHA256},
		{"pilot", "PiLoT/unseen.png", originalRef.SourceSHA256},
	} {
		t.Run(tc.name, func(t *testing.T) {
			materials["unseen"] = connectionMaterial{tc.path, tc.hash}
			fogTestJSON(t, path, m)
			out := filepath.Join(filepath.Dir(path), "out-"+tc.name)
			if err := run([]string{"-fog", path, "-out", out}, io.Discard); err == nil {
				t.Fatal("accepted unsafe source")
			}
			if _, err := os.Lstat(out); !os.IsNotExist(err) {
				t.Fatal("unsafe source created output")
			}
		})
	}
	materials["unseen"] = originalRef
	for _, name := range []string{"unseen", "summer", "winter"} {
		t.Run("nonopaque-"+name, func(t *testing.T) {
			im := image.NewNRGBA(image.Rect(0, 0, 384, 320))
			draw.Draw(im, im.Bounds(), image.NewUniform(color.NRGBA{100, 120, 140, 255}), image.Point{}, draw.Src)
			im.SetNRGBA(383, 319, color.NRGBA{100, 120, 140, 254})
			old := materials[name]
			materials[name] = fogTestPNG(t, filepath.Join(filepath.Dir(path), "alpha.png"), im)
			fogTestJSON(t, path, m)
			out := filepath.Join(filepath.Dir(path), "out-alpha-"+name)
			if err := run([]string{"-fog", path, "-out", out}, io.Discard); err == nil {
				t.Fatal("accepted nonopaque material")
			}
			if _, err := os.Lstat(out); !os.IsNotExist(err) {
				t.Fatal("nonopaque input created output")
			}
			materials[name] = old
		})
	}
	for _, size := range []image.Point{{383, 320}, {384, 319}, {1, 1}} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			im := image.NewNRGBA(image.Rectangle{Max: size})
			draw.Draw(im, im.Bounds(), image.NewUniform(color.NRGBA{20, 30, 40, 255}), image.Point{}, draw.Src)
			materials["unseen"] = fogTestPNG(t, filepath.Join(filepath.Dir(path), "small.png"), im)
			fogTestJSON(t, path, m)
			if err := run([]string{"-fog", path, "-out", filepath.Join(filepath.Dir(path), "small-out")}, io.Discard); err == nil {
				t.Fatal("accepted undersized source")
			}
		})
	}
	materials["unseen"] = originalRef
	outside := t.TempDir()
	contents, err := os.ReadFile(filepath.Join(filepath.Dir(path), originalRef.Source))
	if err != nil {
		t.Fatal(err)
	}
	fogTestWrite(t, filepath.Join(outside, "escape.png"), contents)
	for _, tc := range []struct{ name, target string }{
		{"escape", filepath.Join(outside, "escape.png")},
		{"original-link", filepath.Join(filepath.Dir(path), "original", "hidden.png")},
	} {
		if tc.name == "original-link" {
			if err = os.Mkdir(filepath.Dir(tc.target), 0700); err != nil {
				t.Fatal(err)
			}
			fogTestWrite(t, tc.target, contents)
		}
		link := tc.name + ".png"
		if err = os.Symlink(tc.target, filepath.Join(filepath.Dir(path), link)); err != nil {
			t.Fatal(err)
		}
		materials["unseen"] = connectionMaterial{link, originalRef.SourceSHA256}
		fogTestJSON(t, path, m)
		if err = run([]string{"-fog", path, "-out", filepath.Join(filepath.Dir(path), tc.name+"-out")}, io.Discard); err == nil {
			t.Fatal("accepted unsafe symlink")
		}
	}
	materials["unseen"] = originalRef
	fogTestJSON(t, path, m)
	existing := filepath.Join(filepath.Dir(path), "existing")
	if err = os.Mkdir(existing, 0700); err != nil {
		t.Fatal(err)
	}
	fogTestWrite(t, filepath.Join(existing, "sentinel"), []byte("preserve"))
	link := filepath.Join(filepath.Dir(path), "output-link")
	if err = os.Symlink(existing, link); err != nil {
		t.Fatal(err)
	}
	dangling := filepath.Join(filepath.Dir(path), "dangling-link")
	if err = os.Symlink(filepath.Join(filepath.Dir(path), "missing"), dangling); err != nil {
		t.Fatal(err)
	}
	for _, out := range []string{existing, link, dangling} {
		if err = run([]string{"-fog", path, "-out", out}, io.Discard); err == nil {
			t.Fatalf("accepted existing output %s", out)
		}
	}
	if b, e := os.ReadFile(filepath.Join(existing, "sentinel")); e != nil || string(b) != "preserve" {
		t.Fatal("existing output modified")
	}
	out := filepath.Join(filepath.Dir(path), "rollback")
	if err = run([]string{"-fog", path, "-out", out}, connectedBrokenWriter{}); err == nil {
		t.Fatal("writer error lost")
	}
	if _, err = os.Lstat(out); !os.IsNotExist(err) {
		t.Fatal("failed output left artifacts")
	}
}
