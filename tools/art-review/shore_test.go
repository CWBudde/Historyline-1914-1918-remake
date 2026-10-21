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
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Independent numeric expectations, including lakeside23's effective 234 side mask.
func shoreTestProfiles() map[string][]int {
	return map[string][]int{
		"lakeside12": {1, 2, 3}, "lakeside16": {0, 1, 2}, "lakeside23": {2, 3, 4, 5},
		"lakeside34": {3, 4, 5}, "lakeside56": {0, 1, 5}, "lakeside123": {1, 2, 3, 4},
		"lakeside126": {0, 1, 2, 3}, "lakeside156": {0, 1, 2, 5}, "lakeside3456": {0, 1, 3, 4, 5},
	}
}
func shoreTestFixture(t *testing.T) (string, map[string]any) {
	t.Helper()
	dir := t.TempDir()
	materials := map[string]any{}
	for i, name := range []string{"grass", "water", "bank"} {
		im := image.NewNRGBA(image.Rect(0, 0, 384, 320))
		for y := range 320 {
			for x := range 384 {
				c := []color.NRGBA{{55, 115, 35, 255}, {20, 65, 125, 255}, {135, 100, 55, 255}}[i]
				c.R += uint8((x + 2*y) % 7)
				c.G += uint8((3*x + y) % 5)
				im.SetNRGBA(x, y, c)
			}
		}
		var b bytes.Buffer
		if err := png.Encode(&b, im); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name+".png"), b.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b.Bytes())
		materials[name] = map[string]any{"source": name + ".png", "source_sha256": hex.EncodeToString(sum[:])}
	}
	variants := []any{}
	for _, id := range []string{"lakeside12", "lakeside16", "lakeside23", "lakeside34", "lakeside56", "lakeside123", "lakeside126", "lakeside156", "lakeside3456"} {
		variants = append(variants, map[string]any{"id": id, "land_vertices": shoreTestProfiles()[id]})
	}
	m := map[string]any{"schema_version": 1, "materials": materials, "variants": variants}
	path := filepath.Join(dir, "manifest.json")
	shoreTestWrite(t, path, m)
	return path, m
}
func shoreTestWrite(t *testing.T, path string, m any) {
	t.Helper()
	b, e := json.Marshal(m)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(path, b, 0600); e != nil {
		t.Fatal(e)
	}
}
func TestShoreCLIRequiresExclusiveMode(t *testing.T) {
	for _, args := range [][]string{{"-shore", ""}, {"-shore", "x", "-out", "y", "-compound", ""}, {"-shore", "x", "-out", "y", "-resampling", "area"}, {"-shore", "x", "-out", "y", "pos"}} {
		err := run(args, io.Discard)
		if err == nil || !strings.Contains(err.Error(), "-shore requires") {
			t.Fatalf("%v: %v", args, err)
		}
	}
}
func TestShoreCLIActualPNGFamily(t *testing.T) {
	path, _ := shoreTestFixture(t)
	out := filepath.Join(filepath.Dir(path), "output")
	if err := run([]string{"-shore", path, "-out", out}, io.Discard); err != nil {
		t.Fatal(err)
	}
	var report struct {
		Variants []struct {
			ID   string `json:"id"`
			Land []int  `json:"land_vertices"`
		} `json:"variants"`
		Exports []packOutput `json:"exports"`
		Scenes  []shoreScene `json:"scenes"`
	}
	b, err := os.ReadFile(filepath.Join(out, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Variants) != 11 {
		t.Fatalf("variants %d", len(report.Variants))
	}
	for _, v := range report.Variants {
		if want, ok := shoreTestProfiles()[v.ID]; ok {
			if fmt.Sprint(v.Land) != fmt.Sprint(want) {
				t.Fatalf("profile %s: %v want %v", v.ID, v.Land, want)
			}
		} else if v.ID != "grass" && v.ID != "water" {
			t.Fatalf("unexpected %s", v.ID)
		}
		for _, density := range []int{2, 4} {
			im, err := loadPNG(filepath.Join(out, fmt.Sprintf("tiles/%s-%dx.png", v.ID, density)))
			if err != nil {
				t.Fatal(err)
			}
			if im.Bounds() != image.Rect(0, 0, 96*density, 80*density) {
				t.Fatal("dimensions")
			}
			for y := range im.Bounds().Dy() {
				for x := range im.Bounds().Dx() {
					inside := connectionInside(sharedConnectionGeometry(), connectionPoint{(float64(x) + .5) / float64(density), (float64(y) + .5) / float64(density)})
					c := im.NRGBAAt(x, y)
					if inside && c.A != 255 {
						t.Fatalf("%s owned alpha at %d,%d", v.ID, x, y)
					}
					if !inside && c != (color.NRGBA{}) {
						t.Fatal("outside RGBA")
					}
				}
			}
			if v.ID == "grass" || v.ID == "water" {
				material, err := loadPNG(filepath.Join(filepath.Dir(path), v.ID+".png"))
				if err != nil {
					t.Fatal(err)
				}
				adapter, err := exportConnectionTiles(connectionInputs{manifest: connectionManifest{Geometry: sharedConnectionGeometry()}, materials: map[string]*image.NRGBA{"grass": material}, grassSampling: v.ID == "grass"}, nil, nil)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(im.Pix, adapter[density].Pix) {
					t.Fatalf("%s %dx adapter differs", v.ID, density)
				}
			}
		}
	}
	if len(report.Exports) != 63 {
		t.Fatalf("expected 22 tiles +40 scenes +1 sheet, got %d", len(report.Exports))
	}
	// PNG hashes and dimensions are checked against bytes actually saved to disk.
	for _, e := range report.Exports {
		b, err := os.ReadFile(filepath.Join(out, e.Path))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != e.SHA256 {
			t.Fatal("export hash")
		}
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		if im.Bounds().Dx() != e.Dimensions.Width || im.Bounds().Dy() != e.Dimensions.Height {
			t.Fatal("export dimensions")
		}
	}
	// A bank and both base regions must survive the actual PNG chain.
	im, err := loadPNG(filepath.Join(out, "tiles/lakeside12-4x.png"))
	if err != nil {
		t.Fatal(err)
	}
	regions := [3]int{}
	for y := range im.Bounds().Dy() {
		for x := range im.Bounds().Dx() {
			c := im.NRGBAAt(x, y)
			if c.A == 0 {
				continue
			}
			if c.G > c.R+30 {
				regions[0]++
			}
			if c.B > c.G+30 {
				regions[1]++
			}
			if c.R > c.G+15 && c.G > c.B+15 {
				regions[2]++
			}
		}
	}
	for _, n := range regions {
		if n < 100 {
			t.Fatalf("missing material regions: %v", regions)
		}
	}
	shoreTestAssertScenes(t, out, report.Scenes)
	shoreTestAssertPNGGuards(t, path, out)
	second := filepath.Join(filepath.Dir(path), "second")
	if err = run([]string{"-shore", path, "-out", second}, io.Discard); err != nil {
		t.Fatal(err)
	}
	firstFiles := 0
	if err = filepath.WalkDir(out, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		firstFiles++
		rel, e := filepath.Rel(out, path)
		if e != nil {
			return e
		}
		left, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		right, e := os.ReadFile(filepath.Join(second, rel))
		if e != nil {
			return e
		}
		if !bytes.Equal(left, right) {
			return fmt.Errorf("nondeterministic file %s", rel)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	// Generated PNGs, three retained sources, manifest, report, gallery.
	if firstFiles != 69 {
		t.Fatalf("unexpected file list: %d", firstFiles)
	}
}

// Removing strict pin/schema/path checks must fail before a fresh output is created.
func TestShoreRejectsInvalidInputsBeforeOutput(t *testing.T) {
	path, base := shoreTestFixture(t)
	b, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	changes := []struct {
		name   string
		change func(map[string]any)
	}{
		{"schema", func(m map[string]any) { m["schema_version"] = 2 }},
		{"missing", func(m map[string]any) { delete(m, "materials") }},
		{"extra-material", func(m map[string]any) {
			m["materials"].(map[string]any)["snow"] = m["materials"].(map[string]any)["grass"]
		}},
		{"missing-material", func(m map[string]any) { delete(m["materials"].(map[string]any), "bank") }},
		{"short-roster", func(m map[string]any) { m["variants"] = m["variants"].([]any)[:8] }},
		{"duplicate-id", func(m map[string]any) { vs := m["variants"].([]any); vs[1] = vs[0] }},
		{"unknown-id", func(m map[string]any) { m["variants"].([]any)[0].(map[string]any)["id"] = "lakeside999" }},
		{"23-without-exception", func(m map[string]any) { m["variants"].([]any)[2].(map[string]any)["land_vertices"] = []int{2, 3, 4} }},
		{"unsorted", func(m map[string]any) { m["variants"].([]any)[0].(map[string]any)["land_vertices"] = []int{3, 2, 1} }},
		{"duplicate-vertex", func(m map[string]any) { m["variants"].([]any)[0].(map[string]any)["land_vertices"] = []int{1, 2, 2, 3} }},
		{"null-vertex", func(m map[string]any) {
			m["variants"].([]any)[0].(map[string]any)["land_vertices"] = []any{nil, 1, 2, 3}
		}},
	}
	for _, p := range []string{"../grass.png", "/tmp/grass.png", "source/../grass.png", "original/grass.png", "pilot/grass.png", "source\\grass.png"} {
		changes = append(changes, struct {
			name   string
			change func(map[string]any)
		}{"unsafe-" + p, func(m map[string]any) { m["materials"].(map[string]any)["grass"].(map[string]any)["source"] = p }})
	}
	changes = append(changes, struct {
		name   string
		change func(map[string]any)
	}{"changed-pin", func(m map[string]any) {
		m["materials"].(map[string]any)["grass"].(map[string]any)["source_sha256"] = strings.Repeat("0", 64)
	}})
	changes = append(changes, struct {
		name   string
		change func(map[string]any)
	}{"uppercase-pin", func(m map[string]any) {
		ref := m["materials"].(map[string]any)["grass"].(map[string]any)
		ref["source_sha256"] = strings.ToUpper(ref["source_sha256"].(string))
	}})
	for _, tc := range changes {
		t.Run(tc.name, func(t *testing.T) {
			var m map[string]any
			if e := json.Unmarshal(b, &m); e != nil {
				t.Fatal(e)
			}
			tc.change(m)
			shoreTestWrite(t, path, m)
			out := filepath.Join(filepath.Dir(path), "reject")
			if e := run([]string{"-shore", path, "-out", out}, io.Discard); e == nil {
				t.Fatal("invalid input accepted")
			}
			if _, e := os.Lstat(out); !os.IsNotExist(e) {
				t.Fatal("invalid input created output")
			}
		})
	}
	for name, data := range map[string]string{
		"null-root": "null", "array-root": "[]", "trailing": string(b) + " {}", "wrongtype": strings.Replace(string(b), `"schema_version":1`, `"schema_version":"1"`, 1),
		"root-alias":        strings.Replace(string(b), `"schema_version"`, `"Schema_version"`, 1),
		"ref-alias":         strings.Replace(string(b), `"source"`, `"Source"`, 1),
		"variant-alias":     strings.Replace(string(b), `"land_vertices"`, `"Land_vertices"`, 1),
		"duplicate-root":    strings.Replace(string(b), `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1),
		"duplicate-ref":     strings.Replace(string(b), `"source":"bank.png"`, `"source":"bank.png","source":"bank.png"`, 1),
		"duplicate-variant": strings.Replace(string(b), `"id":"lakeside12"`, `"id":"lakeside12","id":"lakeside12"`, 1),
		"null-materials":    strings.Replace(string(b), `"materials":{`, `"materials":null,"bad":{`, 1),
		"null-source":       strings.Replace(string(b), `"source":"bank.png"`, `"source":null`, 1),
		"null-vertices":     strings.Replace(string(b), `"land_vertices":[1,2,3]`, `"land_vertices":null`, 1),
		"vertex-wrongtype":  strings.Replace(string(b), `"land_vertices":[1,2,3]`, `"land_vertices":["1",2,3]`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if e := os.WriteFile(path, []byte(data), 0600); e != nil {
				t.Fatal(e)
			}
			if _, e := loadShoreInputs(path); e == nil {
				t.Fatalf("invalid JSON schema accepted: %s", data)
			}
		})
	}
	shoreTestWrite(t, path, base)
	// A lexical safe-looking symlink must not bypass the existing source exclusions.
	dir := filepath.Dir(path)
	if err = os.Mkdir(filepath.Join(dir, "original"), 0700); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(filepath.Join(dir, "grass.png"))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "original", "grass.png"), original, 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink("original/grass.png", filepath.Join(dir, "alias.png")); err != nil {
		t.Fatal(err)
	}
	base["materials"].(map[string]any)["grass"].(map[string]any)["source"] = "alias.png"
	shoreTestWrite(t, path, base)
	if _, err = loadShoreInputs(path); err == nil {
		t.Fatal("resolved original material accepted")
	}
}

func TestShoreRetainsLoadedSnapshotsAndRollsBack(t *testing.T) {
	path, _ := shoreTestFixture(t)
	in, err := loadShoreInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	manifest := append([]byte(nil), in.Bytes...)
	sources := map[string][]byte{}
	for name, src := range in.Materials {
		sources[name] = append([]byte(nil), src.Bytes...)
	}
	for _, name := range []string{"manifest.json", "grass.png", "water.png", "bank.png"} {
		if err = os.WriteFile(filepath.Join(filepath.Dir(path), name), []byte("changed after pin"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	bundle, err := buildShoreBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range bundle.artifacts {
		if a.name == "manifest.json" && !bytes.Equal(a.data, manifest) {
			t.Fatal("manifest reread")
		}
		for name, b := range sources {
			if a.name == "sources/"+name+".png" && !bytes.Equal(a.data, b) {
				t.Fatal("source reread")
			}
		}
	}
	sum := sha256.Sum256(manifest)
	if bundle.report.ManifestSHA256 != hex.EncodeToString(sum[:]) {
		t.Fatal("manifest snapshot hash")
	}
	path, _ = shoreTestFixture(t)
	out := filepath.Join(filepath.Dir(path), "rollback")
	if err = run([]string{"-shore", path, "-out", out}, connectedBrokenWriter{}); err == nil {
		t.Fatal("failed log writer ignored")
	}
	if _, err = os.Lstat(out); !os.IsNotExist(err) {
		t.Fatal("failed export retained output")
	}
	if err = os.Mkdir(out, 0700); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(out, "sentinel")
	if err = os.WriteFile(sentinel, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = run([]string{"-shore", path, "-out", out}, io.Discard); err == nil {
		t.Fatal("existing output accepted")
	}
	b, err := os.ReadFile(sentinel)
	if err != nil || string(b) != "keep" {
		t.Fatal("existing output changed")
	}
}

// Catches invented neighbor IDs, parity shifts, ring conflicts, and previews
// rerendered from geometry rather than ordinary Over of the saved tile PNGs.
func shoreTestAssertScenes(t *testing.T, out string, scenes []shoreScene) {
	t.Helper()
	if len(scenes) != 20 {
		t.Fatalf("scene count: %d", len(scenes))
	}
	profiles := shoreTestProfiles()
	profiles["grass"] = []int{0, 1, 2, 3, 4, 5}
	profiles["water"] = []int{}
	polygon := []image.Point{{25, 0}, {71, 0}, {96, 40}, {71, 80}, {25, 80}, {0, 40}}
	neighbors := []image.Point{{71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}}
	seen := map[string]bool{}
	for _, scene := range scenes {
		expectedID := fmt.Sprintf("%s-patch-parity%d", scene.CenterID, scene.CenterLineParity)
		if scene.ID != expectedID || seen[scene.ID] {
			t.Fatal("duplicate/incorrect scene ID")
		}
		seen[scene.ID] = true
		origin := image.Point{}
		if scene.CenterLineParity == 1 {
			origin = image.Pt(71, 40)
		} else if scene.CenterLineParity != 0 {
			t.Fatal("invalid parity")
		}
		positions := map[image.Point]bool{origin: true}
		for _, n := range neighbors {
			positions[origin.Add(n)] = true
		}
		if len(scene.Cells) != 7 {
			t.Fatal("seven cells required")
		}
		cells := map[image.Point]string{}
		worldValues := map[image.Point]bool{}
		worldCounts := map[image.Point]int{}
		minX, minY := 999, 999
		for _, c := range scene.Cells {
			p := image.Pt(int(c.Origin[0]), int(c.Origin[1]))
			if !positions[p] || cells[p] != "" {
				t.Fatal("invalid or duplicate runtime origin")
			}
			cells[p] = c.RuntimeID
			if p == origin && c.RuntimeID != scene.CenterID {
				t.Fatal("incorrect center")
			}
			profile, ok := profiles[c.RuntimeID]
			if !ok || fmt.Sprint(c.LandVertices) != fmt.Sprint(profile) {
				t.Fatal("invented ID/profile")
			}
			minX, minY = min(minX, p.X), min(minY, p.Y)
			for i, q := range polygon {
				at := p.Add(q)
				value := false
				for _, vertex := range profile {
					value = value || vertex == i
				}
				if old, ok := worldValues[at]; ok && old != value {
					t.Fatalf("three-cell vertex conflict %s at%v", scene.ID, at)
				}
				worldValues[at] = value
				worldCounts[at]++
			}
		}
		if scene.CanvasOrigin != (connectionPoint{float64(minX), float64(minY)}) {
			t.Fatal("canvas origin")
		}
		contacts, triples := 0, 0
		for p := range cells {
			for _, n := range neighbors {
				if cells[p.Add(n)] != "" {
					contacts++
				}
			}
		}
		for _, n := range worldCounts {
			if n == 3 {
				triples++
			}
		}
		if contacts != 24 || triples != 6 {
			t.Fatalf("expected12innercontacts/6threecellvertices: %d/%d", contacts, triples)
		}
		for _, density := range []int{2, 4} {
			name := scene.Output.Path
			if density == 4 {
				name = scene.Output4x.Path
			}
			if name != fmt.Sprintf("%s-%dx.png", scene.ID, density) {
				t.Fatal("preview path")
			}
			got, err := loadPNG(filepath.Join(out, name))
			if err != nil {
				t.Fatal(err)
			}
			want := image.NewNRGBA(image.Rect(0, 0, 238*density, 240*density))
			for _, c := range scene.Cells {
				tile, err := loadPNG(filepath.Join(out, fmt.Sprintf("tiles/%s-%dx.png", c.RuntimeID, density)))
				if err != nil {
					t.Fatal(err)
				}
				at := image.Pt((int(c.Origin[0])-minX)*density, (int(c.Origin[1])-minY)*density)
				draw.Draw(want, tile.Bounds().Add(at), tile, image.Point{}, draw.Over)
				// Every shared three-cell corner must be pure compatible base color.
				for i, q := range polygon {
					if worldCounts[image.Pt(int(c.Origin[0]), int(c.Origin[1])).Add(q)] != 3 {
						continue
					}
					baseID := "water"
					for _, vertex := range profiles[c.RuntimeID] {
						if vertex == i {
							baseID = "grass"
						}
					}
					base, err := loadPNG(filepath.Join(out, fmt.Sprintf("tiles/%s-%dx.png", baseID, density)))
					if err != nil {
						t.Fatal(err)
					}
					for y := max(0, (q.Y-2)*density); y < min(80*density, (q.Y+2)*density); y++ {
						for x := max(0, (q.X-2)*density); x < min(96*density, (q.X+2)*density); x++ {
							color := tile.NRGBAAt(x, y)
							if color.A > 0 && color != base.NRGBAAt(x, y) {
								t.Fatalf("corner base mismatch %s vertex%d %dx", c.RuntimeID, i, density)
							}
						}
					}
				}
			}
			if got.Bounds() != want.Bounds() || !bytes.Equal(got.Pix, want.Pix) {
				t.Fatalf("savedPNG Over assembly differs: %s %dx", scene.ID, density)
			}
		}
	}
}

// An independent projection oracle catches a field change in saved PNG guards.
// Material appearance is checked separately by the all-variant bank test.
// The same existing Lanczos kernel verifies the full unmasked
// filter support instead of comparing unrelated nearest boundary pixel centers.
func shoreTestProjectedColor(in shoreInputs, a, b connectionPoint, landA, landB bool, p connectionPoint) color.NRGBA {
	d := -32.
	if landA && landB {
		d = 32
	} else if landA != landB {
		e := b.sub(a)
		q := p.sub(a.add(b).mul(.5))
		d = (e[0]*q[0] + e[1]*q[1]) / math.Hypot(e[0], e[1])
		if landA {
			d = -d
		}
	}
	return shoreMaterialColor(in, p, d)
}
func shoreTestAssertPNGGuards(t *testing.T, path, out string) {
	t.Helper()
	in, err := loadShoreInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	profiles := shoreTestProfiles()
	profiles["grass"] = []int{0, 1, 2, 3, 4, 5}
	profiles["water"] = []int{}
	vertices := []connectionPoint{{25, 0}, {71, 0}, {96, 40}, {71, 80}, {25, 80}, {0, 40}}
	contains := func(v []int, i int) bool {
		for _, j := range v {
			if i == j {
				return true
			}
		}
		return false
	}
	for direction := 1; direction <= 6; direction++ {
		a, b := vertices[direction%6], vertices[(direction+1)%6]
		e := b.sub(a)
		length := e.length()
		for _, states := range [][2]bool{{false, false}, {false, true}, {true, false}, {true, true}} {
			raw := image.NewNRGBA(image.Rect(0, 0, 416, 352))
			for y := range 352 {
				for x := range 416 {
					p := connectionPoint{(float64(x)+.5)/4 - 4, (float64(y)+.5)/4 - 4}
					raw.SetNRGBA(x, y, shoreTestProjectedColor(in, a, b, states[0], states[1], p))
				}
			}
			raw2, err := resizeLanczos(raw, 208, 176)
			if err != nil {
				t.Fatal(err)
			}
			for id, profile := range profiles {
				if contains(profile, direction%6) != states[0] || contains(profile, (direction+1)%6) != states[1] {
					continue
				}
				for _, density := range []int{2, 4} {
					im, err := loadPNG(filepath.Join(out, fmt.Sprintf("tiles/%s-%dx.png", id, density)))
					if err != nil {
						t.Fatal(err)
					}
					oracle := raw
					if density == 2 {
						oracle = raw2
					}
					checked := 0
					for y := range 80 * density {
						for x := range 96 * density {
							p := connectionPoint{(float64(x) + .5) / float64(density), (float64(y) + .5) / float64(density)}
							q := p.sub(a)
							depth := (e[0]*q[1] - e[1]*q[0]) / length
							along := (e[0]*q[0] + e[1]*q[1]) / length
							if depth < 0 || depth > 6 || along < 10 || along > length-10 || im.NRGBAAt(x, y).A == 0 {
								continue
							}
							checked++
							want := oracle.NRGBAAt(x+4*density, y+4*density)
							if got := im.NRGBAAt(x, y); got != want {
								t.Fatalf("analytic PNG guard %s dir%d %dx at%d,%d got%v want%v", id, direction, density, x, y, got, want)
							}
						}
					}
					if checked < 100 {
						t.Fatal("guard test sampled too few owned pixels")
					}
				}
			}
		}
	}
}

func TestShoreEveryCompatiblePairReciprocalAndInteriorSmooth(t *testing.T) {
	for _, profile := range []string{"", authoredBankProfile} {
		t.Run(profile, func(t *testing.T) {
			path, m := shoreTestFixture(t)
			if profile != "" {
				m["bank_profile"] = profile
				shoreTestWrite(t, path, m)
			}
			in, err := loadShoreInputs(path)
			if err != nil {
				t.Fatal(err)
			}
			checkShoreCompatiblePairs(t, in)
		})
	}
}

func checkShoreCompatiblePairs(t *testing.T, in shoreInputs) {
	t.Helper()
	for _, v := range shoreVariants() {
		for i, p := range sharedConnectionGeometry().Polygon {
			land := false
			for _, vertex := range v.LandVertices {
				land = land || vertex == i
			}
			var want color.NRGBA
			if land {
				want = sampleGrassMaterial(in.Materials["grass"].Image, p)
			} else if in.Manifest.BankProfile == authoredBankProfile {
				want = sampleAuthoredWater(in.Materials["water"].Image, p)
			} else {
				want = sampleConnectionMaterial(in.Materials["water"].Image, p)
			}
			if got := shoreColor(in, v, p); got != want {
				t.Fatalf("three-tile corner material %s vertex%d: %v != %v", v.ID, i, got, want)
			}
		}
	}
	profiles := shoreTestProfiles()
	profiles["grass"] = []int{0, 1, 2, 3, 4, 5}
	profiles["water"] = []int{}
	vertices := []connectionPoint{{25, 0}, {71, 0}, {96, 40}, {71, 80}, {25, 80}, {0, 40}}
	neighbors := []connectionPoint{{71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}}
	bits := func(profile []int) [6]bool {
		var b [6]bool
		for _, i := range profile {
			b[i] = true
		}
		return b
	}
	pairs := 0
	for idA, profileA := range profiles {
		for idB, profileB := range profiles {
			aBits, bBits := bits(profileA), bits(profileB)
			for direction := 1; direction <= 6; direction++ {
				opposite := (direction+2)%6 + 1
				if aBits[direction%6] != bBits[(opposite+1)%6] || aBits[(direction+1)%6] != bBits[opposite%6] {
					continue
				}
				pairs++
				a, b := vertices[direction%6], vertices[(direction+1)%6]
				edge := b.sub(a)
				length := edge.length()
				tangent := edge.mul(1 / length)
				normal := connectionPoint{-tangent[1], tangent[0]}
				for _, depth := range []float64{-8, -4, 0, 4, 8} {
					for _, along := range []float64{.02, .25, .5, .75, .98} {
						p := a.add(edge.mul(along)).add(normal.mul(depth))
						q := p.sub(neighbors[direction-1])
						ca := shoreColor(in, shoreVariant{idA, profileA}, p)
						cb := shoreColor(in, shoreVariant{idB, profileB}, q)
						if ca != cb {
							t.Fatalf("reciprocal field color %s/%s dir%d depth%g t%g: %v/%v", idA, idB, direction, depth, along, ca, cb)
						}
					}
				}
			}
		}
	}
	if pairs != 234 {
		t.Fatalf("expected all 234 compatible ordered pairs: %d", pairs)
	}
	// A triangular fan introduces derivative jumps across center-to-vertex rays;
	// the smooth interior must retain a continuous transverse derivative there.
	for id, profile := range shoreTestProfiles() {
		for _, vertex := range vertices {
			ray := vertex.sub(connectionPoint{48, 40})
			p := connectionPoint{48, 40}.add(ray.mul(.3))
			n := connectionPoint{-ray[1], ray[0]}.mul(.01 / ray.length())
			v := shoreVariant{id, profile}
			left, center, right := shoreField(v, p.sub(n)), shoreField(v, p), shoreField(v, p.add(n))
			jump := math.Abs((right-center)/.01 - (center-left)/.01)
			if jump > .01 {
				t.Fatalf("interior derivative jump %s at%v: %g", id, p, jump)
			}
		}
	}
}

// Rejected first-pass art had a reverse bend at lakeside23's right shoulder
// and a flat cove wall on3456. Trace the actual scalar coast, independent of
// texture, so a smooth deep interior alone cannot hide the bad guard join.
func TestShoreCoastJoinsWithoutReverseBendsOrFlatCove(t *testing.T) {
	profile := shoreVariant{"lakeside23", []int{2, 3, 4, 5}}
	crossingY := func(x float64) float64 {
		lo, hi := 0., 80.
		for range 50 {
			mid := (lo + hi) / 2
			if shoreField(profile, connectionPoint{x, mid}) > 0 {
				hi = mid
			} else {
				lo = mid
			}
		}
		return (lo + hi) / 2
	}
	for x := 60.; x < 82; x += .25 {
		a, b := crossingY(x), crossingY(x+.25)
		if b-a > .02 {
			t.Fatalf("23 shoulder reverses coast at x%g: y%g ->%g", x, a, b)
		}
	}
	profile = shoreVariant{"lakeside3456", []int{0, 1, 3, 4, 5}}
	crossingX := func(y float64) float64 {
		lo, hi := 0., 96.
		for range 50 {
			mid := (lo + hi) / 2
			if shoreField(profile, connectionPoint{mid, y}) > 0 {
				lo = mid
			} else {
				hi = mid
			}
		}
		return (lo + hi) / 2
	}
	if bow := crossingX(32) + crossingX(48) - 2*crossingX(40); bow < 2 {
		t.Fatalf("3456 cove wall is flat: 16worldpx span has bow%g", bow)
	}
}

func TestShoreOpaqueMinimumMaterialDimensions(t *testing.T) {
	for _, tc := range []struct {
		name          string
		width, height int
		alpha         uint8
	}{{"short-width", 383, 320, 255}, {"short-height", 384, 319, 255}, {"translucent", 384, 320, 254}} {
		t.Run(tc.name, func(t *testing.T) {
			path, m := shoreTestFixture(t)
			im := image.NewNRGBA(image.Rect(0, 0, tc.width, tc.height))
			draw.Draw(im, im.Bounds(), image.NewUniform(color.NRGBA{60, 100, 40, 255}), image.Point{}, draw.Src)
			im.SetNRGBA(100, 100, color.NRGBA{60, 100, 40, tc.alpha})
			var b bytes.Buffer
			if err := png.Encode(&b, im); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(filepath.Dir(path), "grass.png"), b.Bytes(), 0600); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(b.Bytes())
			m["materials"].(map[string]any)["grass"].(map[string]any)["source_sha256"] = hex.EncodeToString(sum[:])
			shoreTestWrite(t, path, m)
			if _, err := loadShoreInputs(path); err == nil {
				t.Fatal("invalid pinned material accepted")
			}
		})
	}
}

// A stony bank must remain readable several world pixels inland on every
// crossing. Returning grass too early recreated the old hairline brown shore.
func TestShoreStonyBankReadableAcrossAllNineVariants(t *testing.T) {
	path, _ := shoreTestFixture(t)
	in, err := loadShoreInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	bank := in.Materials["bank"].Image
	draw.Draw(bank, bank.Bounds(), image.NewUniform(color.NRGBA{205, 195, 175, 255}), image.Point{}, draw.Src)
	g := sharedConnectionGeometry()
	for _, v := range shoreRoster() {
		for edge, a := range g.Polygon {
			b := g.Polygon[(edge+1)%6]
			landA, landB := false, false
			for _, vertex := range v.LandVertices {
				landA = landA || vertex == edge
				landB = landB || vertex == (edge+1)%6
			}
			if landA == landB {
				continue
			}
			e := b.sub(a)
			along := e.mul(1 / e.length())
			if landA {
				along = along.mul(-1)
			}
			p := a.add(b).mul(.5).add(along.mul(2.8))
			c := shoreColor(in, v, p)
			if c.R < 150 || c.G < 140 || c.B < 115 {
				t.Errorf("%s edge%d: bank disappears inland: %v", v.ID, edge, c)
			}
		}
	}
}
