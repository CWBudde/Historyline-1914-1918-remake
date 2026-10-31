package artreview

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
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestMixedFlagExclusiveEvenEmpty(t *testing.T) {
	for _, args := range [][]string{{"-mixed", "", "-out", "x"}, {"-mixed", "missing", "-connected", "", "-out", "x"}, {"-mixed", "missing", "-candidate", "", "-out", "x"}} {
		err := run(args, io.Discard)
		if err == nil || !strings.Contains(err.Error(), "-mixed requires") {
			t.Fatalf("%v: %v", args, err)
		}
	}
}

func TestMixedSchemaRejectsAliasesAndInapplicableFields(t *testing.T) {
	for _, s := range []string{`{"schema_version":1,"manifests":[],"variants":[{"id":"bridge25","runtime_id":"bridge25","kind":"footbridge","ports":[2,5],"layers":[]}]}`, `{"schema_version":1,"manifests":[],"variants":[{"id":"bridge25","runtime_id":"bridge25","kind":"footbridge","Ports":[2,5]}]}`, `{"schema_version":1,"schema_version":1}`} {
		if _, err := decodeMixedManifest([]byte(s)); err == nil {
			t.Fatal("accepted malformed mixed manifest", s)
		}
	}
}

func TestMixedRawComposesBeforeFilteringAndFiniteBridge(t *testing.T) {
	in := mixedInputs{merged: connectedSolidInputs(t), components: map[string]mixedComponent{}}
	for _, f := range []connectionFamily{{ID: "path", Width: 7, Border: .75, Material: "path", BorderMaterial: "earth"}, {ID: "road", Width: 14, Border: 2, Material: "road", BorderMaterial: "earth"}, {ID: "rail", Width: 11, Border: 1, Material: "rail", BorderMaterial: "earth"}} {
		for _, axis := range []string{"14", "25", "36"} {
			v := connectionVariant{ID: f.ID + axis, Ports: []int{int(axis[0] - '0'), int(axis[1] - '0')}}
			in.components[v.ID] = mixedComponent{in.merged, f, v}
		}
	}
	// Distinct materials make sampling the wrong source detectable.
	for id, c := range map[string]color.NRGBA{"path": {200, 170, 70, 255}, "timber": {90, 40, 10, 255}, "rail": {80, 90, 100, 255}, "steel": {220, 230, 240, 255}} {
		im := image.NewNRGBA(image.Rect(0, 0, 384, 320))
		draw.Draw(im, im.Bounds(), image.NewUniform(c), image.Point{}, draw.Src)
		in.merged.materials[id] = im
	}
	for id, c := range in.components {
		c.input = in.merged
		in.components[id] = c
	}
	for _, v := range mixedRoster() {
		tiles, err := exportMixedTiles(in, v)
		if err != nil {
			t.Fatal(err)
		}
		for scale, im := range tiles {
			for y := 0; y < im.Bounds().Dy(); y++ {
				for x := 0; x < im.Bounds().Dx(); x++ {
					p := connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}
					if !connectionInside(sharedConnectionGeometry(), p) && im.NRGBAAt(x, y) != (color.NRGBA{}) {
						t.Fatal("outside polygon RGBA")
					}
				}
			}
			if v.Kind == "road-bridge" {
				if im.NRGBAAt(48*scale, 2*scale).A != 0 {
					t.Fatal("finite bridge touches boundary")
				}
				if im.NRGBAAt(48*scale, 40*scale).A == 0 {
					t.Fatal("missing bridge deck")
				}
			}
		}
	}
}

func mixedFixture(t *testing.T) (string, mixedManifest) {
	t.Helper()
	return mixedFixtureTracks(t, [2]float64{7, .75}, [2]float64{11, 1})
}

// mixedFixtureTracks is mixedFixture with the given path and rail profiles.
func mixedFixtureTracks(t *testing.T, pathProfile, railProfile [2]float64) (string, mixedManifest) {
	t.Helper()
	dir := t.TempDir()
	mats := map[string]connectionMaterial{}
	colors := map[string]color.NRGBA{"grass": {30, 100, 40, 255}, "earth": {90, 55, 30, 255}, "water": {30, 80, 170, 255}, "road": {170, 160, 140, 255}, "path": {205, 175, 90, 255}, "rail": {100, 110, 120, 255}, "timber": {100, 45, 15, 255}, "steel": {220, 230, 240, 255}}
	for id, c := range colors {
		im := image.NewNRGBA(image.Rect(0, 0, 384, 320))
		draw.Draw(im, im.Bounds(), image.NewUniform(c), image.Point{}, draw.Src)
		var b bytes.Buffer
		if err := png.Encode(&b, im); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b.Bytes())
		name := id + ".png"
		if err := os.WriteFile(filepath.Join(dir, name), b.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
		mats[id] = connectionMaterial{name, hex.EncodeToString(sum[:])}
	}
	manifest := mixedManifest{SchemaVersion: 1, Variants: mixedRoster()}
	for i, families := range [][]string{{"road", "stream"}, {"path"}, {"rail"}} {
		m := connectionManifest{SchemaVersion: 1, Geometry: sharedConnectionGeometry(), Materials: map[string]connectionMaterial{}}
		needed := []string{"grass", "road", "water", "earth"}
		if i == 1 {
			needed = append(needed, "path")
		}
		if i == 2 {
			needed = append(needed, "rail", "timber", "steel")
		}
		for _, id := range needed {
			m.Materials[id] = mats[id]
		}
		for _, id := range families {
			width, border := 14., 2.
			material := id
			if id == "stream" {
				width, border, material = 18, 2.5, "water"
			}
			if id == "path" {
				width, border = pathProfile[0], pathProfile[1]
			}
			if id == "rail" {
				width, border = railProfile[0], railProfile[1]
			}
			f := connectionFamily{ID: id, Width: width, Border: border, Material: material, BorderMaterial: "earth"}
			for _, axis := range []string{"14", "25", "36"} {
				f.Variants = append(f.Variants, connectionVariant{ID: id + axis, Ports: []int{int(axis[0] - '0'), int(axis[1] - '0')}})
			}
			m.Families = append(m.Families, f)
		}
		b, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b)
		name := fmt.Sprintf("manifest-%d.json", i)
		if err = os.WriteFile(filepath.Join(dir, name), b, 0o644); err != nil {
			t.Fatal(err)
		}
		manifest.Manifests = append(manifest.Manifests, connectionMaterial{name, hex.EncodeToString(sum[:])})
	}
	b, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "mixed.json")
	if err = os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return path, manifest
}

func TestMixedLoadAndSchemaValidation(t *testing.T) {
	path, m := mixedFixture(t)
	if _, err := loadMixedInputs(path); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"duplicate-ref", "wrong-hash", "absolute", "original", "pilot", "duplicate-id", "runtime-case", "kind", "wrong-ports", "wrong-layer-order", "missing-component", "symlink-original", "missing-material"} {
		t.Run(name, func(t *testing.T) {
			path, m := mixedFixture(t)
			switch name {
			case "duplicate-ref":
				m.Manifests[1] = m.Manifests[0]
			case "wrong-hash":
				m.Manifests[0].SourceSHA256 = strings.Repeat("0", 64)
			case "absolute":
				m.Manifests[0].Source = filepath.Join(filepath.Dir(path), m.Manifests[0].Source)
			case "original", "pilot":
				dir := filepath.Join(filepath.Dir(path), name)
				if err := os.Mkdir(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(filepath.Join(filepath.Dir(path), m.Manifests[0].Source))
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(dir, "input.json"), data, 0o644); err != nil {
					t.Fatal(err)
				}
				m.Manifests[0].Source = name + "/input.json"
			case "symlink-original":
				dir := filepath.Join(filepath.Dir(path), "original")
				if err := os.Mkdir(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(filepath.Join(filepath.Dir(path), m.Manifests[0].Source))
				if err != nil {
					t.Fatal(err)
				}
				target := filepath.Join(dir, "input.json")
				if err = os.WriteFile(target, data, 0o644); err != nil {
					t.Fatal(err)
				}
				if err = os.Symlink(target, filepath.Join(filepath.Dir(path), "alias.json")); err != nil {
					t.Fatal(err)
				}
				m.Manifests[0].Source = "alias.json"
			case "duplicate-id":
				m.Variants[1] = m.Variants[0]
			case "runtime-case":
				m.Variants[2].RuntimeID = "rbridge14"
			case "kind":
				m.Variants[0].Kind = "road-bridge"
			case "wrong-ports":
				m.Variants[0].Ports = []int{1, 3}
			case "wrong-layer-order":
				m.Variants[5].Layers = []string{"rail36", "path14"}
			case "missing-component", "missing-material":
				p := filepath.Join(filepath.Dir(path), m.Manifests[1].Source)
				data, err := os.ReadFile(p)
				if err != nil {
					t.Fatal(err)
				}
				var c connectionManifest
				if err = json.Unmarshal(data, &c); err != nil {
					t.Fatal(err)
				}
				if name == "missing-component" {
					c.Families[0].Variants = c.Families[0].Variants[1:]
				} else {
					delete(c.Materials, "path")
				}
				data, err = json.Marshal(c)
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(p, data, 0o644); err != nil {
					t.Fatal(err)
				}
				sum := sha256.Sum256(data)
				m.Manifests[1].SourceSHA256 = hex.EncodeToString(sum[:])
			}
			data, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(path, data, 0o644); err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(filepath.Dir(path), "out")
			if err = run([]string{"-mixed", path, "-out", out}, io.Discard); err == nil {
				t.Fatal("accepted malformed input")
			}
			if _, err = os.Stat(out); !os.IsNotExist(err) {
				t.Fatal("invalid input wrote output")
			}
		})
	}
	// Canonical schema also accepts the complete roster in any declared order.
	m.Variants[0], m.Variants[1] = m.Variants[1], m.Variants[0]
	b, _ := json.Marshal(m)
	if _, err := decodeMixedManifest(b); err != nil {
		t.Fatal(err)
	}
}

func TestMixedExportProtectedPortsAndVisibleStructure(t *testing.T) {
	path, _ := mixedFixture(t)
	in, err := loadMixedInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range mixedRoster() {
		tiles, err := exportMixedTiles(in, v)
		if err != nil {
			t.Fatal(err)
		}
		if v.Kind == "road-bridge" {
			axis := mixedAxis(v.Ports)
			for _, family := range []string{"road", "path", "rail"} {
				c := in.components[family+axis]
				upper, err := exportConnectionTiles(c.input, &c.family, &c.variant)
				if err != nil {
					t.Fatal(err)
				}
				for _, scale := range []int{2, 4} {
					surviving := 0
					for y := 0; y < 80*scale; y++ {
						for x := 0; x < 96*scale; x++ {
							a, b := tiles[scale].NRGBAAt(x, y), upper[scale].NRGBAAt(x, y)
							p := connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}
							if mixedEdgeDistance(sharedConnectionGeometry(), p) < 8 && a.A != 0 {
								t.Fatalf("%s finite structure touched protected boundary", v.ID)
							}
							if a.A == 255 && b.A == 0 {
								surviving++
							}
						}
					}
					if surviving < 50*scale*scale {
						t.Fatalf("%s under %s %dx: only %d visible structure pixels", v.ID, family, scale, surviving)
					}
				}
			}
			continue
		}
		layers := v.Layers
		if v.Kind == "footbridge" {
			layers = []string{"path" + mixedAxis(v.Ports)}
		}
		// Expected contacts come from actual independently filtered connected PNGs.
		imported := map[string]map[int]*image.NRGBA{}
		for _, id := range layers {
			c := in.components[id]
			im, err := exportConnectionTiles(c.input, &c.family, &c.variant)
			if err != nil {
				t.Fatal(err)
			}
			imported[id] = im
		}
		for _, scale := range []int{2, 4} {
			contacts := 0
			for y := 0; y < 80*scale; y++ {
				for x := 0; x < 96*scale; x++ {
					p := connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}
					d := mixedEdgeDistance(sharedConnectionGeometry(), p)
					if d < 0 || d > 8 {
						continue
					}
					want := color.NRGBA{}
					for _, id := range layers {
						want = mixedOver(want, imported[id][scale].NRGBAAt(x, y))
					}
					got := tiles[scale].NRGBAAt(x, y)
					if want.A > 0 {
						contacts++
					}
					if got != want {
						t.Fatalf("%s %dx protected contact (%d,%d): got %v want %v", v.ID, scale, x, y, got, want)
					}
				}
			}
			if contacts == 0 {
				t.Fatal("no protected port contacts tested")
			}
		}
	}
}

func TestMixedRejectsIncompatibleRehashedProfile(t *testing.T) {
	for _, family := range []string{"road", "path", "rail", "stream"} {
		t.Run(family, func(t *testing.T) {
			path, m := mixedFixture(t)
			for i, ref := range m.Manifests {
				p := filepath.Join(filepath.Dir(path), ref.Source)
				b, err := os.ReadFile(p)
				if err != nil {
					t.Fatal(err)
				}
				var manifest connectionManifest
				if err = json.Unmarshal(b, &manifest); err != nil {
					t.Fatal(err)
				}
				changed := false
				for j := range manifest.Families {
					if manifest.Families[j].ID == family {
						manifest.Families[j].Width++
						changed = true
					}
				}
				if !changed {
					continue
				}
				b, err = json.Marshal(manifest)
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(p, b, 0o644); err != nil {
					t.Fatal(err)
				}
				sum := sha256.Sum256(b)
				m.Manifests[i].SourceSHA256 = hex.EncodeToString(sum[:])
			}
			b, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(path, b, 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err = loadMixedInputs(path); err == nil || !strings.Contains(err.Error(), "profile") {
				t.Fatalf("incompatible rehashed %s profile accepted: %v", family, err)
			}
		})
	}
}

func TestMixedCrossingMaterialOrderAndFootApproach(t *testing.T) {
	path, _ := mixedFixture(t)
	in, err := loadMixedInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	v := mixedRoster()[6] // road14 crossed by vertical rail36
	raw := renderMixedRaw(in, v)
	steel := in.merged.materials["steel"].NRGBAAt(0, 0)
	center := raw.NRGBAAt((48+4)*4, (41+4)*4)
	if center.R < 125 || center.G > 130 || center.B > 90 || center.A != 255 {
		t.Fatalf("intersection must have lightened warm timber over road and ballast: %v", center)
	}
	// Fixed-gauge rail centerline is x=48 +/- gauge/2, and must cross timber.
	sx := int((48 + connectionRailGauge/2 + 4) * 4)
	if got := raw.NRGBAAt(sx, (40+4)*4); got != steel {
		t.Fatalf("continuous steel must remain above infill: %v want %v", got, steel)
	}
	foot := renderMixedRaw(in, mixedRoster()[1])
	c := in.components["path36"]
	approach := renderConnectionRaw(c.input, &c.family, &c.variant)
	if got, want := foot.NRGBAAt((48+4)*4, (4+4)*4), approach.NRGBAAt((48+4)*4, (4+4)*4); got != want {
		t.Fatalf("foot approach material: %v want %v", got, want)
	}
}

func TestMixedBundleDeterministicAndActualPNGAssembly(t *testing.T) {
	path, _ := mixedFixture(t)
	base := filepath.Dir(path)
	before := map[string][]byte{}
	entries, err := os.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		b, err := os.ReadFile(filepath.Join(base, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		before[e.Name()] = b
	}
	var first mixedReport
	for pass := 0; pass < 2; pass++ {
		out := filepath.Join(base, fmt.Sprintf("out%d", pass))
		if err = run([]string{"-mixed", path, "-out", out}, io.Discard); err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(filepath.Join(out, "report.json"))
		if err != nil {
			t.Fatal(err)
		}
		var report mixedReport
		if err = json.Unmarshal(b, &report); err != nil {
			t.Fatal(err)
		}
		if len(report.Variants) != 9 || len(report.Scenes) != 32 || len(report.Outputs) != 127 {
			t.Fatalf("bundle counts: %d variants/%d scenes/%d PNGs", len(report.Variants), len(report.Scenes), len(report.Outputs))
		}
		if pass == 0 {
			first = report
		} else {
			a, _ := json.Marshal(first)
			if !bytes.Equal(a, bytes.TrimSpace(b)) { // report is indented; compare semantic canonical JSON.
				canonical, _ := json.Marshal(report)
				if !bytes.Equal(a, canonical) {
					t.Fatal("nondeterministic report")
				}
			}
		}
		tiles := map[string]map[int]*image.NRGBA{}
		for _, o := range report.Outputs {
			data, err := os.ReadFile(filepath.Join(out, o.Path))
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(data)
			if hex.EncodeToString(sum[:]) != o.SHA256 {
				t.Fatal("wrong PNG hash", o.Path)
			}
			im, err := loadPNG(filepath.Join(out, o.Path))
			if err != nil {
				t.Fatal(err)
			}
			if im.Bounds().Dx() != o.Dimensions.Width || im.Bounds().Dy() != o.Dimensions.Height {
				t.Fatal("wrong PNG dimensions", o.Path)
			}
			name := strings.TrimPrefix(o.Path, "dependencies/")
			for _, scale := range []int{2, 4} {
				suffix := fmt.Sprintf("-%dx.png", scale)
				if strings.HasSuffix(name, suffix) && !strings.Contains(name, "-grass-") && !strings.Contains(name, "patch-") {
					id := strings.TrimSuffix(name, suffix)
					if tiles[id] == nil {
						tiles[id] = map[int]*image.NRGBA{}
					}
					tiles[id][scale] = im
				}
			}
		}
		for _, s := range report.Scenes {
			assertMixedSceneTypedTopology(t, s)
			if len(s.Cells) != 7 {
				t.Fatal("not seven cells")
			}
			for _, scale := range []int{2, 4} {
				// Assemble independently using decoded exports and ordinary integer Over.
				minX, minY := math.Inf(1), math.Inf(1)
				for _, c := range s.Cells {
					minX = math.Min(minX, c.Origin[0])
					minY = math.Min(minY, c.Origin[1])
				}
				o := s.Output
				if scale == 4 {
					o = s.Output4x
				}
				expected := image.NewNRGBA(image.Rect(0, 0, o.Dimensions.Width, o.Dimensions.Height))
				for _, c := range s.Cells {
					p := image.Pt(int(c.Origin[0]-minX)*scale, int(c.Origin[1]-minY)*scale)
					tile := tiles["grass"][scale]
					draw.Draw(expected, tile.Bounds().Add(p), tile, image.Point{}, draw.Over)
				}
				for _, c := range s.Cells {
					p := image.Pt(int(c.Origin[0]-minX)*scale, int(c.Origin[1]-minY)*scale)
					for _, l := range c.Layers {
						tile := tiles[l.ID][scale]
						if tile == nil {
							t.Fatal("missing exported layer", l.ID)
						}
						draw.Draw(expected, tile.Bounds().Add(p), tile, image.Point{}, draw.Over)
					}
				}
				actual, err := loadPNG(filepath.Join(out, o.Path))
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(actual.Pix, expected.Pix) {
					t.Fatal("scene does not use actual PNG assembly", s.ID, scale)
				}

			}
		}
		if err = run([]string{"-mixed", path, "-out", out}, io.Discard); err == nil {
			t.Fatal("reused output accepted")
		}
	}
	for name, want := range before {
		got, err := os.ReadFile(filepath.Join(base, name))
		if err != nil || !bytes.Equal(got, want) {
			t.Fatal("source mutated", name, err)
		}
	}
}

func TestMixedCanonicalAndPresenceFieldsOnOtherwiseValidRoster(t *testing.T) {
	_, m := mixedFixture(t)
	valid, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = decodeMixedManifest(valid); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"top-alias", "kind-alias", "ports-alias", "bridge-empty-layers", "bridge-null-layers", "crossing-empty-ports", "crossing-null-ports", "ref-alias", "unknown", "duplicate-key"} {
		t.Run(name, func(t *testing.T) {
			var raw map[string]any
			if err := json.Unmarshal(valid, &raw); err != nil {
				t.Fatal(err)
			}
			variants := raw["variants"].([]any)
			bridge := variants[0].(map[string]any)
			cross := variants[5].(map[string]any)
			switch name {
			case "top-alias":
				raw["Schema_Version"] = raw["schema_version"]
				delete(raw, "schema_version")
			case "kind-alias":
				bridge["Kind"] = bridge["kind"]
				delete(bridge, "kind")
			case "ports-alias":
				bridge["Ports"] = bridge["ports"]
				delete(bridge, "ports")
			case "bridge-empty-layers":
				bridge["layers"] = []any{}
			case "bridge-null-layers":
				bridge["layers"] = nil
			case "crossing-empty-ports":
				cross["ports"] = []any{}
			case "crossing-null-ports":
				cross["ports"] = nil
			case "ref-alias":
				ref := raw["manifests"].([]any)[0].(map[string]any)
				ref["Source"] = ref["source"]
				delete(ref, "source")
			case "unknown":
				bridge["extra"] = 1
			}
			b, err := json.Marshal(raw)
			if err != nil {
				t.Fatal(err)
			}
			if name == "duplicate-key" {
				b = bytes.Replace(b, []byte(`"kind":"footbridge"`), []byte(`"kind":"footbridge","kind":"footbridge"`), 1)
			}
			if _, err = decodeMixedManifest(b); err == nil {
				t.Fatal("accepted malformed otherwise valid roster")
			}
		})
	}
}

func TestMixedRollsBackFailedOutput(t *testing.T) {
	path, _ := mixedFixture(t)
	out := filepath.Join(filepath.Dir(path), "failed")
	if err := run([]string{"-mixed", path, "-out", out}, connectedBrokenWriter{}); err == nil {
		t.Fatal("logging failure ignored")
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("failed bundle was not removed: %v", err)
	}
}

func TestMixedBridgeReadableBoardsAndPosts(t *testing.T) {
	path, _ := mixedFixture(t)
	in, err := loadMixedInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []mixedVariant{mixedRoster()[1], mixedRoster()[4]} {
		tiles, err := exportMixedTiles(in, v)
		if err != nil {
			t.Fatal(err)
		}
		for _, scale := range []int{2, 4} {
			im := tiles[scale]
			sum, count := 0, 0
			lo, hi := 255, 0
			for y := 24 * scale; y < 56*scale; y++ {
				c := im.NRGBAAt(48*scale, y)
				sum += int(c.R)
				count++
				lo = min(lo, int(c.R))
				hi = max(hi, int(c.R))
			}
			if sum/count < 125 || hi-lo < 35 {
				t.Fatalf("%s %dx wood lacks readable light boards/seams: mean=%d range=%d", v.ID, scale, sum/count, hi-lo)
			}
			across := 6.75
			if v.Kind == "road-bridge" {
				across = 11.25
			}
			post := im.NRGBAAt(int((48+across)*float64(scale)), 40*scale)
			beam := im.NRGBAAt(int((48+across)*float64(scale)), 44*scale)
			if int(post.R)-int(beam.R) < 15 {
				t.Fatalf("%s %dx no distinct guard post/cap: post=%v beam=%v", v.ID, scale, post, beam)
			}
		}
	}
}

func TestMixedLoadedImportMustMatchVerifiedSnapshot(t *testing.T) {
	path, m := mixedFixture(t)
	ref := m.Manifests[0]
	importPath := filepath.Join(filepath.Dir(path), ref.Source)
	verified, err := os.ReadFile(importPath)
	if err != nil {
		t.Fatal(err)
	}
	// A real, valid connected definition changes after the initial verification.
	// Whitespace leaves behavior valid while changing exactly the retained hash.
	changed := append(append([]byte{}, verified...), '\n')
	if err = os.WriteFile(importPath, changed, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err = loadPinnedMixedConnection(importPath, ref.SourceSHA256); err == nil || !strings.Contains(err.Error(), "loaded source manifest SHA256") {
		t.Fatalf("changed loaded definition accepted under old pin: %v", err)
	}
	sum := sha256.Sum256(changed)
	if _, err = loadPinnedMixedConnection(importPath, hex.EncodeToString(sum[:])); err != nil {
		t.Fatalf("matching loaded definition rejected: %v", err)
	}
}

func TestMixedRetainedTopManifestUsesInitialSnapshot(t *testing.T) {
	path, _ := mixedFixture(t)
	initial, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	in, err := loadMixedInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, []byte("changed after load"), 0o644); err != nil {
		t.Fatal(err)
	}
	artifact := in.retainedManifest()
	sum := sha256.Sum256(artifact.data)
	if artifact.name != "mixed-manifest.json" || !bytes.Equal(artifact.data, initial) || hex.EncodeToString(sum[:]) != in.hash {
		t.Fatalf("retained top manifest does not match loaded snapshot/hash: %s", artifact.name)
	}
}

func assertMixedSceneTypedTopology(t *testing.T, scene mixedScene) {
	t.Helper()
	type fixture struct{ center, continuations []string }
	fixtures := map[string]fixture{
		"bridge25":                           {[]string{"stream36", "bridge25"}, []string{"stream36", "path25"}},
		"bridge36":                           {[]string{"stream14", "bridge36"}, []string{"stream14", "path36"}},
		"rbridge14":                          {[]string{"stream36", "rbridge14", "road14"}, []string{"stream36", "road14"}},
		"rbridge25":                          {[]string{"stream36", "rbridge25", "road25"}, []string{"stream36", "road25"}},
		"rbridge36":                          {[]string{"stream14", "rbridge36", "road36"}, []string{"stream14", "road36"}},
		"cross1":                             {[]string{"cross1"}, []string{"path14", "rail36"}},
		"cross2":                             {[]string{"cross2"}, []string{"road14", "rail36"}},
		"cross3":                             {[]string{"cross3"}, []string{"road25", "rail36"}},
		"cross4":                             {[]string{"cross4"}, []string{"path36", "rail14"}},
		"campaign-rbridge25-stream36-road25": {[]string{"stream36", "rbridge25", "road25"}, []string{"stream36", "road25"}},
	}
	for _, stream := range []string{"stream14", "stream25"} {
		for _, route := range []string{"road36", "path36", "rail36"} {
			fixtures["campaign-rbridge36-"+stream+"-"+route] = fixture{[]string{stream, "rbridge36", route}, []string{stream, route}}
		}
	}
	key := strings.TrimSuffix(scene.ID, fmt.Sprintf("-patch-parity%d", scene.Parity))
	spec, ok := fixtures[key]
	if !ok {
		t.Fatal("unexpected scene fixture", scene.ID)
	}
	mixedPorts := map[string][]int{"bridge25": {2, 5}, "bridge36": {3, 6}, "rbridge14": {1, 4}, "rbridge25": {2, 5}, "rbridge36": {3, 6}, "cross1": {1, 3, 4, 6}, "cross2": {1, 3, 4, 6}, "cross3": {2, 3, 5, 6}, "cross4": {1, 3, 4, 6}}
	ports := func(id string) []int {
		if p, ok := mixedPorts[id]; ok {
			return p
		}
		return []int{int(id[len(id)-2] - '0'), int(id[len(id)-1] - '0')}
	}
	center := connectionPoint{}
	if scene.Parity == 1 {
		center = connectionPoint{71, 40}
	} else if scene.Parity != 0 {
		t.Fatal("invalid parity")
	}
	neighbors := []connectionPoint{{71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}}
	expected := map[connectionPoint][]string{center: spec.center}
	for i, delta := range neighbors {
		origin := connectionPoint{center[0] + delta[0], center[1] + delta[1]}
		expected[origin] = []string{}
		for _, id := range spec.continuations {
			for _, p := range ports(id) {
				if p == i+1 {
					expected[origin] = append(expected[origin], id)
					break
				}
			}
		}
	}
	seen := map[connectionPoint]bool{}
	for _, cell := range scene.Cells {
		ids, ok := expected[cell.Origin]
		if !ok || seen[cell.Origin] {
			t.Fatal("unexpected/duplicate cell origin", scene.ID, cell.Origin)
		}
		seen[cell.Origin] = true
		if len(cell.Layers) != len(ids) {
			t.Fatalf("%s cell%v layers=%v want%v", scene.ID, cell.Origin, cell.Layers, ids)
		}
		for i, id := range ids {
			layer := cell.Layers[i]
			want := ports(id)
			alpha := want
			if strings.HasPrefix(id, "rbridge") {
				alpha = []int{}
			}
			if layer.ID != id || !slices.Equal(layer.Ports, want) || !slices.Equal(layer.AlphaPorts, alpha) {
				t.Fatalf("%s cell%v layer%d=%+v want id=%s semantic=%v alpha=%v", scene.ID, cell.Origin, i, layer, id, want, alpha)
			}
			if cell.Origin != center {
				// Exact component ID includes family, and every neighbor must contain the
				// reciprocal port back to the center on the same straight component.
				delta := cell.Origin.sub(center)
				port := 0
				for j, d := range neighbors {
					if d == delta {
						port = j + 1
					}
				}
				reciprocal := (port+2)%6 + 1
				if !slices.Contains(layer.Ports, reciprocal) {
					t.Fatal("missing typed reciprocal continuation", scene.ID, layer.ID, port)
				}
			}
		}
	}
	if len(seen) != 7 {
		t.Fatal("missing seven-cell topology", scene.ID)
	}
}
