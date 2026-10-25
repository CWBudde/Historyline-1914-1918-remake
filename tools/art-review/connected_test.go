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
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func connectedFixture(t *testing.T) (string, map[string]any) {
	t.Helper()
	dir, m, _ := packFixture(t, 384, 320, false)
	a := m["assets"].([]any)[0].(map[string]any)
	materials := map[string]any{}
	for _, id := range []string{"grass", "road", "water", "earth"} {
		materials[id] = map[string]any{"source": a["source"], "source_sha256": a["source_sha256"]}
	}
	return dir, map[string]any{"schema_version": 1, "geometry": map[string]any{"width": 96, "height": 80, "polygon": [][2]float64{{25, 0}, {71, 0}, {96, 40}, {71, 80}, {25, 80}, {0, 40}}, "neighbors": [][2]float64{{71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}}, "ports": [][2]float64{{83.5, 20}, {83.5, 60}, {48, 80}, {12.5, 60}, {12.5, 20}, {48, 0}}}, "materials": materials, "families": []any{map[string]any{"id": "road", "width": 14, "border": 2, "material": "road", "border_material": "earth", "variants": []any{map[string]any{"id": "road36", "ports": []int{3, 6}}, map[string]any{"id": "road13", "ports": []int{1, 3}}, map[string]any{"id": "road136", "ports": []int{1, 3, 6}}, map[string]any{"id": "road14", "ports": []int{1, 4}}, map[string]any{"id": "road25", "ports": []int{2, 5}}}}}}
}

func writeConnectedManifest(t *testing.T, dir string, m map[string]any) string {
	t.Helper()
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "connected.json")
	if err = os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func connectedPathFamily() map[string]any {
	return map[string]any{"id": "path", "width": 7, "border": .75, "material": "path", "border_material": "earth", "variants": []any{map[string]any{"id": "path14", "ports": []int{1, 4}}, map[string]any{"id": "path25", "ports": []int{2, 5}}, map[string]any{"id": "path36", "ports": []int{3, 6}}}}
}

// Catches paths sampling the road material or using the wider road profile.
func TestConnectedPathBuildUsesItsMaterialAndNarrowProfile(t *testing.T) {
	dir, m := connectedFixture(t)
	m["families"] = []any{connectedPathFamily()}
	core := color.NRGBA{210, 175, 115, 255}
	im := image.NewNRGBA(image.Rect(0, 0, 384, 320))
	draw.Draw(im, im.Bounds(), image.NewUniform(core), image.Point{}, draw.Src)
	var source bytes.Buffer
	if err := png.Encode(&source, im); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "path.png"), source.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(source.Bytes())
	m["materials"].(map[string]any)["path"] = map[string]any{"source": "path.png", "source_sha256": hex.EncodeToString(sum[:])}
	out := filepath.Join(dir, "out")
	if err := run([]string{"-connected", writeConnectedManifest(t, dir, m), "-out", out}, io.Discard); err != nil {
		t.Fatal(err)
	}
	for _, scale := range []int{2, 4} {
		tile, err := loadPNG(filepath.Join(out, fmt.Sprintf("path36-%dx.png", scale)))
		if err != nil {
			t.Fatal(err)
		}
		for _, x := range []int{47, 48, 49} {
			if got := tile.NRGBAAt(x*scale, 40*scale); got != core {
				t.Fatalf("%dx path core at x=%d: %v, want %v", scale, x, got, core)
			}
		}
		for _, p := range []image.Point{{0, 0}, {42, 40}, {54, 40}, {95, 79}} {
			if got := tile.NRGBAAt(p.X*scale, p.Y*scale); got != (color.NRGBA{}) {
				t.Fatalf("%dx outside narrow path at %v: %v", scale, p, got)
			}
		}
		if scale == 4 {
			// The protected port guard keeps seven world pixels (28 samples),
			// from x=44.625 through x=51.375 on this vertical route.
			for x := 0; x < 384; x++ {
				if got := tile.NRGBAAt(x, 16) == core; got != (x >= 178 && x <= 205) {
					t.Fatalf("4x path guard core width incorrect at x=%d: %v", x, tile.NRGBAAt(x, 16))
				}
			}
		}
	}
	retained, err := os.ReadFile(filepath.Join(out, "path-material-source.png"))
	if err != nil || !bytes.Equal(retained, source.Bytes()) {
		t.Fatalf("path source not retained exactly: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(out, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report connectionReport
	if err = json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if got := report.Materials["path"]; got.SourceSHA256 != hex.EncodeToString(sum[:]) || got.RetainedSource != "path-material-source.png" {
		t.Fatalf("missing path provenance: %+v", got)
	}
	assertConnectedPatchesUseExportedPNGs(t, out)
}

func TestConnectedPathMaterialRequiredExactlyForPathFamily(t *testing.T) {
	for _, kind := range []string{"missing path", "stray path", "unknown fifth", "missing fixed", "wrong core", "wrong border", "duplicate path"} {
		t.Run(kind, func(t *testing.T) {
			dir, m := connectedFixture(t)
			materials := m["materials"].(map[string]any)
			materials["path"] = materials["road"]
			path := connectedPathFamily()
			m["families"] = append(m["families"].([]any), path)
			switch kind {
			case "missing path":
				delete(materials, "path")
			case "stray path":
				m["families"] = m["families"].([]any)[:1]
			case "unknown fifth":
				materials["sand"] = materials["path"]
				delete(materials, "path")
			case "missing fixed":
				delete(materials, "road")
			case "wrong core":
				path["material"] = "road"
			case "wrong border":
				path["border_material"] = "grass"
			case "duplicate path":
				m["families"] = append(m["families"].([]any), connectedPathFamily())
			}
			out := filepath.Join(dir, "out")
			if err := run([]string{"-connected", writeConnectedManifest(t, dir, m), "-out", out}, io.Discard); err == nil {
				t.Fatal("accepted invalid path manifest")
			}
			if _, err := os.Lstat(out); !os.IsNotExist(err) {
				t.Fatal("created output before validation")
			}
		})
	}
}

// Catches a missing command dispatch and clipped/opaque overlay exports.
func TestConnectedBuildFreshDeterministicAndTransparent(t *testing.T) {
	dir, m := connectedFixture(t)
	path := writeConnectedManifest(t, dir, m)
	a, b := filepath.Join(dir, "a"), filepath.Join(dir, "b")
	for _, out := range []string{a, b} {
		if err := run([]string{"-connected", path, "-out", out}, io.Discard); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(a)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		x, err := os.ReadFile(filepath.Join(a, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		y, err := os.ReadFile(filepath.Join(b, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(x, y) {
			t.Errorf("nondeterministic %s", entry.Name())
		}
	}
	assertConnectedPatchesUseExportedPNGs(t, a)
	for _, scale := range []int{2, 4} {
		suffix := "2x"
		if scale == 4 {
			suffix = "4x"
		}
		im, err := loadPNG(filepath.Join(a, "road36-"+suffix+".png"))
		if err != nil {
			t.Fatal(err)
		}
		if im.Bounds() != image.Rect(0, 0, 96*scale, 80*scale) {
			t.Fatal(im.Bounds())
		}
		for _, p := range []image.Point{{0, 0}, {20 * scale, 40 * scale}, {95 * scale, 79 * scale}} {
			if im.NRGBAAt(p.X, p.Y).A != 0 {
				t.Errorf("outside feature opaque at %v", p)
			}
		}
		if im.NRGBAAt(48*scale, 40*scale).A != 255 {
			t.Fatal("missing road core")
		}
	}
	if err := run([]string{"-connected", path, "-out", a}, io.Discard); err == nil {
		t.Fatal("overwrote existing output")
	}
}

func TestConnectedRejectsInvalidManifestBeforeWriting(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"geometry", func(m map[string]any) { m["geometry"].(map[string]any)["width"] = 95 }},
		{"missing ports", func(m map[string]any) { connectedVariant(m)["ports"] = []int{} }},
		{"duplicate ports", func(m map[string]any) { connectedVariant(m)["ports"] = []int{1, 1} }},
		{"bad port", func(m map[string]any) { connectedVariant(m)["ports"] = []int{0, 7} }},
		{"duplicate id", func(m map[string]any) { connectedVariant(m)["id"] = "road13" }},
		{"unsafe id", func(m map[string]any) { connectedVariant(m)["id"] = "../bad" }},
		{"bad hash", func(m map[string]any) {
			m["materials"].(map[string]any)["grass"].(map[string]any)["source_sha256"] = strings.Repeat("0", 64)
		}},
		{"unknown family", func(m map[string]any) { m["families"].([]any)[0].(map[string]any)["id"] = "rail" }},
		{"missing material", func(m map[string]any) { delete(m["materials"].(map[string]any), "earth") }},
		{"wide profile", func(m map[string]any) { m["families"].([]any)[0].(map[string]any)["width"] = 80 }},
		{"negative bank", func(m map[string]any) { m["families"].([]any)[0].(map[string]any)["border"] = -1 }},
		{"unknown field", func(m map[string]any) { m["surprise"] = true }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, m := connectedFixture(t)
			tt.mutate(m)
			out := filepath.Join(dir, "out")
			if err := run([]string{"-connected", writeConnectedManifest(t, dir, m), "-out", out}, io.Discard); err == nil {
				t.Fatal("accepted invalid manifest")
			}
			if _, err := os.Lstat(out); !os.IsNotExist(err) {
				t.Fatal("created output before validation")
			}
		})
	}
}

func connectedVariant(m map[string]any) map[string]any {
	return m["families"].([]any)[0].(map[string]any)["variants"].([]any)[0].(map[string]any)
}

func TestConnectedRejectsConflictingFlags(t *testing.T) {
	for _, flags := range [][]string{{"-pack", "x"}, {"-reference", "x"}, {"-candidate", "x"}, {"-neighborhood", "x"}, {"-resampling", "area"}, {"-reference-platform", "new"}, {"positional"}} {
		args := append([]string{"-connected", "missing", "-out", filepath.Join(t.TempDir(), "out")}, flags...)
		err := run(args, io.Discard)
		if err == nil || !strings.Contains(err.Error(), "excludes") {
			t.Fatalf("flags %v: %v", flags, err)
		}
	}
}

func TestConnectedRejectsDerivedOutputNameCollision(t *testing.T) {
	dir, m := connectedFixture(t)
	f := m["families"].([]any)[0].(map[string]any)
	f["variants"] = append(f["variants"].([]any), map[string]any{"id": "road13-grass", "ports": []int{1, 4}})
	if _, err := loadConnectionInputs(writeConnectedManifest(t, dir, m)); err == nil {
		t.Fatal("accepted overlay name colliding with another variant's grass composite")
	}
}

func TestConnectedRejectsSourceDimensionsAlphaAndLegacyPaths(t *testing.T) {
	for _, kind := range []string{"size", "alpha", "original", "pilot"} {
		t.Run(kind, func(t *testing.T) {
			dir, m := connectedFixture(t)
			w := 384
			if kind == "size" {
				w = 383
			}
			im := image.NewNRGBA(image.Rect(0, 0, w, 320))
			for y := 0; y < 320; y++ {
				for x := 0; x < w; x++ {
					im.SetNRGBA(x, y, color.NRGBA{80, 100, 40, 255})
				}
			}
			if kind == "alpha" {
				im.SetNRGBA(0, 0, color.NRGBA{})
			}
			var b bytes.Buffer
			if err := png.Encode(&b, im); err != nil {
				t.Fatal(err)
			}
			name := "input.png"
			if kind == "original" || kind == "pilot" {
				if err := os.Mkdir(filepath.Join(dir, kind), 0o755); err != nil {
					t.Fatal(err)
				}
				name = kind + "/input.png"
			}
			if err := os.WriteFile(filepath.Join(dir, name), b.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(b.Bytes())
			m["materials"].(map[string]any)["grass"] = map[string]any{"source": name, "source_sha256": hex.EncodeToString(sum[:])}
			out := filepath.Join(dir, "out")
			if err := run([]string{"-connected", writeConnectedManifest(t, dir, m), "-out", out}, io.Discard); err == nil {
				t.Fatal("accepted forbidden source")
			}
			if _, err := os.Lstat(out); !os.IsNotExist(err) {
				t.Fatal("output created")
			}
		})
	}
}

// Catches a phase drift between neighboring independent tile exports.
func TestConnectedMaterialLatticeInvariance(t *testing.T) {
	_, _, src := packFixture(t, 401, 337, false)
	for _, p := range []connectionPoint{{0, 0}, {.125, .875}, {-22.5, 81.5}, {141.99, 79.99}, {71, 40}} {
		want := sampleConnectionMaterial(src, p)
		for _, delta := range sharedConnectionGeometry().Neighbors {
			got := sampleConnectionMaterial(src, p.add(delta))
			if got != want {
				t.Errorf("phase %v + %v: %v != %v", p, delta, got, want)
			}
		}
	}
}

// Catches nonreciprocal routes, missing straight guards, width drift and bank gaps.
func TestConnectedReciprocalProfilesAndGuards(t *testing.T) {
	g := sharedConnectionGeometry()
	f := connectionFamily{Width: 14, Border: 2}
	for port := 1; port <= 6; port++ {
		opposite := (port+2)%6 + 1
		p := g.Ports[port-1]
		q := g.Neighbors[port-1].add(g.Ports[opposite-1])
		if p != q {
			t.Fatalf("port %d not reciprocal", port)
		}
		v := connectionVariant{Ports: []int{port, opposite}}
		route := connectionRoutes(g, f, v)
		n := g.Neighbors[port-1].mul(1 / g.Neighbors[port-1].length())
		normal := connectionPoint{-n[1], n[0]}
		for _, depth := range []float64{-3, 0, 3, 8} {
			for _, offset := range []float64{0, 5, 7, 8.5, 10} {
				point := p.add(n.mul(-depth)).add(normal.mul(offset))
				dist := connectionRouteDistance(route, point)
				if math.Abs(dist-offset) > .001 {
					t.Fatalf("port %d depth %g offset %g distance %g", port, depth, offset, dist)
				}
			}
		}
	}
	bend := connectionVariant{Ports: []int{1, 3}}
	route := connectionRoutes(g, f, bend)
	n := g.Neighbors[0].mul(1 / g.Neighbors[0].length())
	for _, d := range []float64{0, 4, 8} {
		if got := connectionRouteDistance(route, g.Ports[0].sub(n.mul(d))); got > 1e-6 {
			t.Errorf("bend guard distance %g", got)
		}
	}
}

// Catches polygon holes at either line parity and double ownership at shared vertices.
func TestConnectedPolygonTessellationHasSingleOwner(t *testing.T) {
	g := sharedConnectionGeometry()
	var cells []connectionCell
	for line := -6; line <= 6; line++ {
		for col := -3; col <= 3; col++ {
			parity := ((line % 2) + 2) % 2
			cells = append(cells, connectionCell{Origin: connectionPoint{float64(col*142 + parity*71), float64(line * 40)}})
		}
	}
	for y := -80; y < 160; y++ {
		for x := -142; x < 284; x++ {
			p := connectionPoint{float64(x) + .125, float64(y) + .375}
			owners := 0
			for _, c := range cells {
				if connectionInside(g, p.sub(c.Origin)) {
					owners++
				}
			}
			if owners != 1 {
				t.Fatalf("%d polygons at non-boundary %v", owners, p)
			}
		}
	}
	// Include exact shared vertices and edge midpoints, where closed polygon
	// tests would select two or three owners instead of one.
	for _, p := range []connectionPoint{{25, 0}, {71, 0}, {96, 40}, {71, 80}, {25, 80}, {0, 40}, {83.5, 20}, {48, 80}} {
		owners := 0
		for _, c := range cells {
			if connectionInside(g, p.sub(c.Origin)) {
				owners++
			}
		}
		if owners != 1 {
			t.Fatalf("shared vertex/edge %v has %d owners", p, owners)
		}
	}
}

// Catches mismatched bank/core materials across reciprocal edges, and a hole in
// the merged junction center. Literal colors distinguish bank from core.
func TestConnectedBoundaryPixelsAndJunctionCenter(t *testing.T) {
	dir, m := connectedFixture(t)
	in, err := loadConnectionInputs(writeConnectedManifest(t, dir, m))
	if err != nil {
		t.Fatal(err)
	}
	for id, c := range map[string]color.NRGBA{"road": {160, 130, 90, 255}, "earth": {90, 60, 35, 255}, "grass": {40, 80, 30, 255}} {
		im := image.NewNRGBA(image.Rect(0, 0, 384, 320))
		for y := 0; y < 320; y++ {
			for x := 0; x < 384; x++ {
				im.SetNRGBA(x, y, c)
			}
		}
		in.materials[id] = im
	}
	f := in.manifest.Families[0]
	g := in.manifest.Geometry
	for port := 1; port <= 6; port++ {
		opposite := (port+2)%6 + 1
		v := connectionVariant{Ports: []int{port, opposite}}
		route := connectionRoutes(g, f, v)
		neighborRoute := connectionRoutes(g, f, connectionVariant{Ports: []int{opposite, port}})
		n := g.Neighbors[port-1].mul(1 / g.Neighbors[port-1].length())
		normal := connectionPoint{-n[1], n[0]}
		for _, sample := range []struct {
			offset float64
			want   color.NRGBA
		}{{0, color.NRGBA{160, 130, 90, 255}}, {6.5, color.NRGBA{160, 130, 90, 255}}, {8, color.NRGBA{90, 60, 35, 255}}, {9, color.NRGBA{90, 60, 35, 128}}, {10, color.NRGBA{}}} {
			p := g.Ports[port-1].add(normal.mul(sample.offset))
			a := connectionFeature(in, f, route, p, p, .25)
			b := connectionFeature(in, f, neighborRoute, p.sub(g.Neighbors[port-1]), p, .25)
			if a != sample.want || a != b {
				t.Fatalf("port %d offset %g: %v / %v want %v", port, sample.offset, a, b, sample.want)
			}
		}
	}
	junction := connectionRoutes(g, f, connectionVariant{Ports: []int{1, 3, 6}})
	center := connectionPoint{48, 40}
	if c := connectionFeature(in, f, junction, center, center, .25); c != (color.NRGBA{160, 130, 90, 255}) {
		t.Fatal("junction center", c)
	}
}

// Catches invented or unmatched internal patch ports, including ring neighbors.
func TestConnectedSceneTopologyIsReciprocal(t *testing.T) {
	dir, m := connectedFixture(t)
	in, err := loadConnectionInputs(writeConnectedManifest(t, dir, m))
	if err != nil {
		t.Fatal(err)
	}
	f := in.manifest.Families[0]
	g := in.manifest.Geometry
	for _, v := range f.Variants {
		for parity := 0; parity < 2; parity++ {
			scene, err := buildConnectionScene(g, f, v, parity)
			if err != nil {
				t.Fatal(err)
			}
			for _, cell := range scene.Cells {
				line := int(cell.Origin[1]) / 40
				lineParity := (line%2 + 2) % 2
				if (int(cell.Origin[0])-lineParity*71)%142 != 0 {
					t.Fatalf("%s cell %v is outside runtime lattice", scene.ID, cell.Origin)
				}
				for _, port := range cell.Ports {
					target := cell.Origin.add(g.Neighbors[port-1])
					for _, neighbor := range scene.Cells {
						if neighbor.Origin != target {
							continue
						}
						reciprocal := false
						for _, p := range neighbor.Ports {
							if p == (port+2)%6+1 {
								reciprocal = true
							}
						}
						if !reciprocal {
							t.Fatalf("%s internal port %d from %v ends in %v", scene.ID, port, cell.Origin, neighbor.Origin)
						}
					}
				}
			}
		}
	}
}

type connectedBrokenWriter struct{}

func (connectedBrokenWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestConnectedRollsBackFailedOutputAndRejectsDuplicateJSON(t *testing.T) {
	dir, m := connectedFixture(t)
	f := m["families"].([]any)[0].(map[string]any)
	f["variants"] = []any{connectedVariant(m)}
	path := writeConnectedManifest(t, dir, m)
	// A vertical straight only needs itself for both continuations.
	out := filepath.Join(dir, "out")
	if err := run([]string{"-connected", path, "-out", out}, connectedBrokenWriter{}); err == nil {
		t.Fatal("lost writer error")
	}
	if _, err := os.Lstat(out); !os.IsNotExist(err) {
		t.Fatal("failed build left output")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	b = bytes.Replace(b, []byte(`"schema_version":1`), []byte(`"schema_version":1,"schema_version":1`), 1)
	if err = os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err = loadConnectionInputs(path); err == nil {
		t.Fatal("accepted duplicate JSON key")
	}
}

// Exercise decoded PNGs with the runtime's ordinary source-over composition.
// Polygon-edge AA in individual exports used to leave alpha 191/255 here.
// The narrow path and the road-wide one both assemble without gaps.
func TestConnectedPNGExportsAssembleWithoutSharedEdgeGaps(t *testing.T) {
	for _, profile := range [][2]float64{{7, .75}, {22, 2}} {
		t.Run(fmt.Sprint(profile[0]), func(t *testing.T) { testConnectedPNGAssembly(t, profile) })
	}
}

func testConnectedPNGAssembly(t *testing.T, pathProfile [2]float64) {
	dir, m := connectedFixture(t)
	colors := map[string]color.NRGBA{"grass": {40, 80, 30, 255}, "road": {160, 130, 90, 255}, "earth": {90, 60, 35, 255}, "water": {40, 90, 130, 255}, "path": {210, 175, 115, 255}}
	for id, c := range colors {
		im := image.NewNRGBA(image.Rect(0, 0, 384, 320))
		for y := 0; y < 320; y++ {
			for x := 0; x < 384; x++ {
				im.SetNRGBA(x, y, c)
			}
		}
		var b bytes.Buffer
		if err := png.Encode(&b, im); err != nil {
			t.Fatal(err)
		}
		name := id + ".png"
		if err := os.WriteFile(filepath.Join(dir, name), b.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b.Bytes())
		m["materials"].(map[string]any)[id] = map[string]any{"source": name, "source_sha256": hex.EncodeToString(sum[:])}
	}
	families := m["families"].([]any)
	families = append(families, map[string]any{"id": "stream", "width": 18, "border": 2.5, "material": "water", "border_material": "earth", "variants": []any{map[string]any{"id": "stream14", "ports": []int{1, 4}}, map[string]any{"id": "stream25", "ports": []int{2, 5}}, map[string]any{"id": "stream36", "ports": []int{3, 6}}}})
	pathFamily := connectedPathFamily()
	pathFamily["width"], pathFamily["border"] = pathProfile[0], pathProfile[1]
	families = append(families, pathFamily)
	m["families"] = families
	out := filepath.Join(dir, "out")
	path := writeConnectedManifest(t, dir, m)
	if err := run([]string{"-connected", path, "-out", out}, io.Discard); err != nil {
		t.Fatal(err)
	}
	in, err := loadConnectionInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	// An unmasked feature reference isolates final polygon ownership from
	// feature-edge AA/filtering. It cannot hide a shared-edge alpha gap.
	unclipped := map[string]map[int]*image.NRGBA{}
	for _, f := range in.manifest.Families {
		for _, v := range f.Variants {
			if len(v.Ports) != 2 || (v.Ports[0]+2)%6+1 != v.Ports[1] {
				continue
			}
			raw := renderConnectionRaw(in, &f, &v)
			two, err := resizeLanczos(raw, 208, 176)
			if err != nil {
				t.Fatal(err)
			}
			unclipped[v.ID] = map[int]*image.NRGBA{}
			for scale, pixels := range map[int]*image.NRGBA{2: two, 4: raw} {
				// Apply the same one source-over conversion to the unmasked
				// reference; NRGBA conversion can round faint alpha-1 RGB by 1.
				composited := image.NewNRGBA(pixels.Bounds())
				draw.Draw(composited, composited.Bounds(), pixels, image.Point{}, draw.Over)
				unclipped[v.ID][scale] = composited
			}
		}
	}
	g := sharedConnectionGeometry()
	for _, family := range []struct {
		id            string
		width, border float64
		core          color.NRGBA
	}{{"road", 14, 2, colors["road"]}, {"stream", 18, 2.5, colors["water"]}, {"path", pathProfile[0], pathProfile[1], colors["path"]}} {
		for _, scale := range []int{2, 4} {
			for parity := 0; parity < 2; parity++ {
				for port := 1; port <= 6; port++ {
					suffix := "36"
					if port == 1 || port == 4 {
						suffix = "14"
					}
					if port == 2 || port == 5 {
						suffix = "25"
					}
					im, err := loadPNG(filepath.Join(out, fmt.Sprintf("%s%s-%dx.png", family.id, suffix, scale)))
					if err != nil {
						t.Fatal(err)
					}
					delta := g.Neighbors[port-1]
					n := delta.mul(1 / delta.length())
					shared := g.Ports[port-1]
					center := connectionPoint{0, 0}
					if parity == 1 {
						center = connectionPoint{71, 40}
					}
					stitched := image.NewNRGBA(image.Rect(-400, -400, 800, 800))
					for _, origin := range []connectionPoint{center, center.add(delta)} {
						pos := image.Pt(int(origin[0])*scale, int(origin[1])*scale)
						draw.Draw(stitched, im.Bounds().Add(pos), im, image.Point{}, draw.Over)
					}
					for y := 0; y < 80*scale; y++ {
						for x := 0; x < 96*scale; x++ {
							p := connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}
							q := p.sub(shared)
							along := q[0]*n[0] + q[1]*n[1]
							across := math.Abs(-q[0]*n[1] + q[1]*n[0])
							if math.Abs(along) > 1/float64(scale) {
								continue
							}
							c := stitched.NRGBAAt(x+int(center[0])*scale, y+int(center[1])*scale)
							want := unclipped[family.id+suffix][scale].NRGBAAt(x+4*scale, y+4*scale)
							// A transparent neighboring draw can round NRGBA RGB one
							// further level while preserving alpha. Keep alpha and
							// opaque core/bank checks exact at both densities.
							faintRGBRoundoff := c.A == want.A && c.A > 0 && c.A < 255 && math.Abs(float64(c.R)-float64(want.R)) <= 1 && math.Abs(float64(c.G)-float64(want.G)) <= 1 && math.Abs(float64(c.B)-float64(want.B)) <= 1
							if across < family.width/2+family.border+2 && c != want && !faintRGBRoundoff {
								t.Fatalf("%s %dx parity%d port%d assembled %v != unclipped %v at %v", family.id, scale, parity, port, c, want, p)
							}
							if across < family.width/2-2 && c != family.core {
								t.Fatalf("%s %dx parity%d port%d core %v at %v", family.id, scale, parity, port, c, p)
							}
							if scale == 4 && across > family.width/2+.25 && across < family.width/2+family.border-.25 && c != colors["earth"] {
								t.Fatalf("%s %dx parity%d port%d bank %v at %v", family.id, scale, parity, port, c, p)
							}
							if across > family.width/2+family.border+2 && across < family.width/2+family.border+3 && c.A != 0 {
								t.Fatalf("outside profile leaked %v", c)
							}
						}
					}
				}
			}
		}
	}
}

func TestConnectedGrassExportsHaveExactlyOneOwnerAtEveryPixelCenter(t *testing.T) {
	dir, m := connectedFixture(t)
	out := filepath.Join(dir, "out")
	if err := run([]string{"-connected", writeConnectedManifest(t, dir, m), "-out", out}, io.Discard); err != nil {
		t.Fatal(err)
	}
	for _, scale := range []int{2, 4} {
		im, err := loadPNG(filepath.Join(out, fmt.Sprintf("grass-%dx.png", scale)))
		if err != nil {
			t.Fatal(err)
		}
		for y := 0; y < 80*scale; y++ {
			for x := 0; x < 142*scale; x++ {
				owners := 0
				for line := -3; line <= 3; line++ {
					for col := -1; col <= 1; col++ {
						px := x - (col*142+((line%2+2)%2)*71)*scale
						py := y - line*40*scale
						a := im.NRGBAAt(px, py).A
						if a != 0 && a != 255 {
							t.Fatalf("fractional polygon ownership alpha %d", a)
						}
						if a == 255 {
							owners++
						}
					}
				}
				if owners != 1 {
					t.Fatalf("%dx pixel %d,%d has %d owners", scale, x, y, owners)
				}
			}
		}
	}
}

func assertConnectedPatchesUseExportedPNGs(t *testing.T, out string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(out, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report connectionReport
	if err = json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	cache := map[string]*image.NRGBA{}
	load := func(name string) *image.NRGBA {
		if im := cache[name]; im != nil {
			return im
		}
		im, err := loadPNG(filepath.Join(out, name))
		if err != nil {
			t.Fatal(err)
		}
		cache[name] = im
		return im
	}
	for _, scene := range report.Scenes {
		for _, scale := range []int{2, 4} {
			minX, minY, maxX, maxY := 100000, 100000, -100000, -100000
			for _, cell := range scene.Cells {
				x, y := int(cell.Origin[0]), int(cell.Origin[1])
				minX = min(minX, x)
				minY = min(minY, y)
				maxX = max(maxX, x+96)
				maxY = max(maxY, y+80)
			}
			assembled := image.NewNRGBA(image.Rect(0, 0, (maxX-minX)*scale, (maxY-minY)*scale))
			for _, layer := range []string{"ground", "overlay"} {
				for _, cell := range scene.Cells {
					id := "grass"
					if layer == "overlay" {
						id = cell.Variant
						if id == "" {
							continue
						}
					}
					tile := load(fmt.Sprintf("%s-%dx.png", id, scale))
					pos := image.Pt((int(cell.Origin[0])-minX)*scale, (int(cell.Origin[1])-minY)*scale)
					draw.Draw(assembled, tile.Bounds().Add(pos), tile, image.Point{}, draw.Over)
				}
			}
			name := scene.Output.Path
			if scale == 4 {
				name = scene.Output4x.Path
			}
			actual := load(name)
			if !bytes.Equal(assembled.Pix, actual.Pix) || assembled.Bounds() != actual.Bounds() {
				t.Fatalf("%s is not assembled from its actual exported PNGs", name)
			}
			var encoded bytes.Buffer
			if err := png.Encode(&encoded, assembled); err != nil {
				t.Fatal(err)
			}
			stored, err := os.ReadFile(filepath.Join(out, name))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(encoded.Bytes(), stored) {
				t.Fatalf("%s reassembled PNG bytes differ", name)
			}
		}
	}
}

func connectedSolidInputs(t *testing.T) connectionInputs {
	t.Helper()
	dir, m := connectedFixture(t)
	in, err := loadConnectionInputs(writeConnectedManifest(t, dir, m))
	if err != nil {
		t.Fatal(err)
	}
	for id, c := range map[string]color.NRGBA{"road": {160, 130, 90, 255}, "water": {40, 90, 130, 255}, "earth": {90, 60, 35, 255}, "grass": {40, 80, 30, 255}, "path": {210, 175, 115, 255}} {
		im := image.NewNRGBA(image.Rect(0, 0, 384, 320))
		for y := 0; y < 320; y++ {
			for x := 0; x < 384; x++ {
				im.SetNRGBA(x, y, c)
			}
		}
		in.materials[id] = im
	}
	return in
}

// Catches an organic path effect leaking into reciprocal guards, leaving
// mathematically uniform interiors, or punching holes through the route core.
func TestConnectedOrganicPathWearsInteriorAndProtectsPorts(t *testing.T) {
	for _, profile := range [][2]float64{{7, .75}, {22, 2}} {
		t.Run(fmt.Sprint(profile[0]), func(t *testing.T) { testConnectedOrganicPath(t, profile[0], profile[1]) })
	}
}

// testConnectedOrganicPath checks the organic path wear on a path of the
// given profile; a wider path wears in proportion.
func testConnectedOrganicPath(t *testing.T, width, border float64) {
	in := connectedSolidInputs(t)
	g := in.manifest.Geometry
	f := connectionFamily{ID: "path", Width: width, Border: border, Material: "path", BorderMaterial: "earth"}
	baseline := f
	baseline.ID = "road"
	v := connectionVariant{Ports: []int{1, 3, 5}}
	route := connectionRoutes(g, f, v)
	for i, a := range g.Polygon {
		edge := g.Polygon[(i+1)%6].sub(a)
		inward := connectionPoint{-edge[1], edge[0]}.mul(1 / edge.length())
		mid := a.add(g.Polygon[(i+1)%6]).mul(.5)
		tangent := edge.mul(1 / edge.length())
		for _, depth := range []float64{-4, 0, 4, 8} {
			for offset := -7; offset <= 7; offset++ {
				p := mid.add(inward.mul(depth)).add(tangent.mul(float64(offset)))
				if got, want := connectionFeature(in, f, route, p, p, .25), connectionFeature(in, baseline, route, p, p, .25); got != want {
					t.Fatalf("path changed protected edge %d depth %g at %v: %v != %v", i, depth, p, got, want)
				}
			}
		}
	}
	rounded, worn, pockets := 0, 0, 0
	for y := 16; y < 64; y++ {
		for x := 20; x < 76; x++ {
			p := connectionPoint{float64(x) + .125, float64(y) + .375}
			got := connectionFeature(in, f, route, p, p, .25)
			want := connectionFeature(in, baseline, route, p, p, .25)
			if p.sub(connectionPoint{48, 40}).length() < .85*f.Width && got == (color.NRGBA{210, 175, 115, 255}) && want != got {
				rounded++
			}
			if p.sub(connectionPoint{48, 40}).length() > 8 && want.A == 255 && got.A < 200 {
				worn++
			}
			if got == (color.NRGBA{90, 60, 35, got.A}) && got.A > 0 && got.A < 240 && want.A == 255 {
				pockets++
			}
			if connectionRouteDistance(route, p) < 2 && got.A != 255 {
				t.Fatalf("path effect made a core hole at %v: %v", p, got)
			}
			for _, delta := range g.Neighbors {
				if shifted := connectionFeature(in, f, route, p, p.add(delta), .25); got != shifted {
					t.Fatalf("path phase changes under lattice shift %v at %v: %v != %v", delta, p, got, shifted)
				}
			}
		}
	}
	if rounded < 3 || worn < 5 || pockets < 5 {
		t.Fatalf("path remained too mechanical: rounded=%d worn=%d translucent earth=%d", rounded, worn, pockets)
	}
	organic, err := exportConnectionTiles(in, &f, &v)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := exportConnectionTiles(in, &baseline, &v)
	if err != nil {
		t.Fatal(err)
	}
	for _, scale := range []int{2, 4} {
		for y := 0; y < 80*scale; y++ {
			for x := 0; x < 96*scale; x++ {
				p := connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}
				distance := math.Inf(1)
				for i, a := range g.Polygon {
					edge := g.Polygon[(i+1)%6].sub(a)
					q := p.sub(a)
					distance = math.Min(distance, (edge[0]*q[1]-edge[1]*q[0])/edge.length())
				}
				if distance <= 6 && organic[scale].NRGBAAt(x, y) != plain[scale].NRGBAAt(x, y) {
					t.Fatalf("path effect escaped filtered port guard at %dx %v", scale, p)
				}
				if distance > 6 && connectionRouteDistance(route, p) < 1 && organic[scale].NRGBAAt(x, y).A != 255 {
					t.Fatalf("exported path core not opaque at %dx %v", scale, p)
				}
			}
		}
	}
}

func TestConnectedOrganicStreamProtectsEdgesAndRoads(t *testing.T) {
	in := connectedSolidInputs(t)
	g := in.manifest.Geometry
	f := connectionFamily{ID: "stream", Width: 18, Border: 2.5, Material: "water", BorderMaterial: "earth"}
	baseline := f
	baseline.ID = "road"
	v := connectionVariant{Ports: []int{1, 3, 5}}
	route := connectionRoutes(g, f, v)
	for i, a := range g.Polygon {
		b := g.Polygon[(i+1)%6]
		edge := b.sub(a)
		inward := connectionPoint{-edge[1], edge[0]}.mul(1 / edge.length())
		mid := a.add(b).mul(.5)
		tangent := edge.mul(1 / edge.length())
		for _, depth := range []float64{-4, 0, 4, 8} {
			for offset := -14; offset <= 14; offset++ {
				p := mid.add(inward.mul(depth)).add(tangent.mul(float64(offset)))
				got := connectionFeature(in, f, route, p, p, .25)
				want := connectionFeature(in, baseline, route, p, p, .25)
				if got != want {
					t.Fatalf("stream changed protected edge %d depth %g at %v: %v != %v", i, depth, p, got, want)
				}
			}
		}
	}
	// Road behavior remains the existing constant-width analytic profile.
	road := connectionFamily{ID: "road", Width: 14, Border: 2, Material: "road", BorderMaterial: "earth"}
	other := road
	other.ID = "unchanged"
	for y := 0; y < 80; y++ {
		for x := 0; x < 96; x++ {
			p := connectionPoint{float64(x) + .125, float64(y) + .375}
			if connectionFeature(in, road, route, p, p, .25) != connectionFeature(in, other, route, p, p, .25) {
				t.Fatal("organic field changed a road")
			}
		}
	}
}

func TestConnectedOrganicStreamRoundsConfluenceAndBreaksInteriorBanks(t *testing.T) {
	in := connectedSolidInputs(t)
	f := connectionFamily{ID: "stream", Width: 18, Border: 2.5, Material: "water", BorderMaterial: "earth"}
	baseline := f
	baseline.ID = "road"
	v := connectionVariant{Ports: []int{1, 3, 5}}
	route := connectionRoutes(in.manifest.Geometry, f, v)
	rounded, bankPockets := 0, 0
	for y := 16; y < 64; y++ {
		for x := 20; x < 76; x++ {
			p := connectionPoint{float64(x) + .125, float64(y) + .375}
			got := connectionFeature(in, f, route, p, p, .25)
			want := connectionFeature(in, baseline, route, p, p, .25)
			if got == (color.NRGBA{40, 90, 130, 255}) && want == (color.NRGBA{90, 60, 35, 255}) && p.sub(connectionPoint{48, 40}).length() < 12.25 {
				rounded++
			}
			if got.R == 90 && got.G == 60 && got.B == 35 && got.A > 0 && got.A < 240 && want.A == 255 {
				bankPockets++
			}
			if connectionRouteDistance(route, p) < 5 && got.A != 255 {
				t.Fatalf("organic field made a core hole at %v: %v", p, got)
			}
		}
	}
	if rounded < 5 {
		t.Fatalf("confluence gained only %d opaque water pixels inside former pointed banks", rounded)
	}
	if bankPockets < 5 {
		t.Fatalf("only %d interior bank pixels reveal underlying terrain", bankPockets)
	}
	first, err := exportConnectionTiles(in, &f, &v)
	if err != nil {
		t.Fatal(err)
	}
	second, err := exportConnectionTiles(in, &f, &v)
	if err != nil {
		t.Fatal(err)
	}
	baselineTiles, err := exportConnectionTiles(in, &baseline, &v)
	if err != nil {
		t.Fatal(err)
	}
	for _, scale := range []int{2, 4} {
		if !bytes.Equal(first[scale].Pix, second[scale].Pix) {
			t.Fatal("organic stream is nondeterministic")
		}
		im := first[scale]
		for y := 0; y < 80*scale; y++ {
			for x := 0; x < 96*scale; x++ {
				p := connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}
				distance := math.Inf(1)
				for i, a := range in.manifest.Geometry.Polygon {
					edge := in.manifest.Geometry.Polygon[(i+1)%6].sub(a)
					q := p.sub(a)
					distance = math.Min(distance, (edge[0]*q[1]-edge[1]*q[0])/edge.length())
				}
				if distance <= 6 && im.NRGBAAt(x, y) != baselineTiles[scale].NRGBAAt(x, y) {
					t.Fatalf("organic effect escaped filtered guard at %dx %v", scale, p)
				}
			}
		}
		// Every active port remains connected through opaque water to the
		// rounded confluence, even after final filtering and ownership.
		seen := make([]bool, 96*80*scale*scale)
		queue := []image.Point{{48 * scale, 40 * scale}}
		seen[queue[0].Y*96*scale+queue[0].X] = true
		for head := 0; head < len(queue); head++ {
			p := queue[head]
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					q := p.Add(image.Pt(dx, dy))
					if !q.In(im.Bounds()) {
						continue
					}
					index := q.Y*96*scale + q.X
					if seen[index] || im.NRGBAAt(q.X, q.Y) != (color.NRGBA{40, 90, 130, 255}) {
						continue
					}
					seen[index] = true
					queue = append(queue, q)
				}
			}
		}
		for _, port := range v.Ports {
			g := in.manifest.Geometry
			n := g.Neighbors[port-1].mul(1 / g.Neighbors[port-1].length())
			p := g.Ports[port-1].sub(n.mul(3))
			if !seen[int(p[1]*float64(scale))*96*scale+int(p[0]*float64(scale))] {
				t.Fatalf("%dx organic stream core disconnected port %d", scale, port)
			}
		}
	}
}

func TestConnectedOrganicStreamProfileBoundsAndLatticePhase(t *testing.T) {
	g := sharedConnectionGeometry()
	f := connectionFamily{ID: "stream", Width: 18, Border: 2.5}
	route := connectionRoutes(g, f, connectionVariant{Ports: []int{3, 6}})
	variation := 0.0
	for y := 8; y < 72; y += 2 {
		for x := 16; x < 80; x += 2 {
			p := connectionPoint{float64(x) + .125, float64(y) + .375}
			core, outer, opacity := connectionStreamProfile(g, f, route, p, p)
			baseline := connectionRouteDistance(route, p) - 9
			change := math.Abs(core - baseline)
			variation = math.Max(variation, change)
			if change > 2.5 {
				t.Fatalf("unbounded stream warp/radius %g", change)
			}
			bank := core - outer
			if bank < 1.4-1e-9 || bank > 2.5+1e-9 || opacity < .35 || opacity > 1 {
				t.Fatalf("unbounded bank width/opacity %g/%g", bank, opacity)
			}
			for _, delta := range g.Neighbors {
				a, b, c := connectionStreamProfile(g, f, route, p, p.add(delta))
				if math.Abs(core-a) > 1e-9 || math.Abs(outer-b) > 1e-9 || math.Abs(opacity-c) > 1e-9 {
					t.Fatal("organic stream field changes under a lattice shift")
				}
			}
		}
	}
	if variation < .5 {
		t.Fatal("stream remained mechanically uniform")
	}
}
