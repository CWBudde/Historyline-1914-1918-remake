package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
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

func TestCompoundFixedLayoutsAndOuterBoundary(t *testing.T) {
	for _, tc := range []struct {
		id                 string
		n, h               int
		first, last, entry string
	}{{"depot-summer", 4, 160, "cmpN", "cmpS", "N"}, {"depot-winter", 4, 160, "cmpN_w", "cmpS_w", "N"}, {"hq-summer", 7, 240, "hqN", "hqS", "N"}, {"hq-winter", 7, 240, "hqN_w", "hqS_w", "N"}, {"factory-winter", 4, 160, "factoryN_w", "factoryS_w", "E"}} {
		l, err := compoundLayoutFor(tc.id)
		if err != nil {
			t.Fatal(err)
		}
		if len(l.Tiles) != tc.n || l.Height != tc.h || l.Tiles[0].RuntimeID != tc.first || l.Tiles[tc.n-1].RuntimeID != tc.last || l.EntryCell != tc.entry {
			t.Fatalf("layout: %+v", l)
		}
		if l.Tiles[0].Origin != (connectionPoint{71, 0}) {
			t.Fatal("N origin")
		}
		expected := []connectionPoint{{71, 0}, {0, 40}, {142, 40}, {71, 80}}
		if tc.n == 7 {
			expected = []connectionPoint{{71, 0}, {0, 40}, {142, 40}, {71, 80}, {0, 120}, {142, 120}, {71, 160}}
		}
		names := []string{"N", "W", "E", "S"}
		prefix := "cmp"
		suffix := ""
		if tc.n == 7 {
			names = []string{"N", "NW", "NE", "C", "SW", "SE", "S"}
			prefix = "hq"
		}
		if strings.HasPrefix(tc.id, "factory-") {
			prefix = "factory"
		}
		if strings.HasSuffix(tc.id, "-winter") {
			suffix = "_w"
		}
		for i, p := range expected {
			tile := l.Tiles[i]
			runtimeID := prefix + names[i] + suffix
			if tile.Origin != p || tile.Cell != names[i] || tile.RuntimeID != runtimeID {
				t.Fatalf("%s tile%d got%+v want cell%s ID%s origin%v", tc.id, i, tile, names[i], runtimeID, p)
			}
		}
		// Shared N/W edge midpoint must be deep inside the union, never a guard edge.
		if d := compoundDistance(l, connectionPoint{83.5, 60}); d < 15 {
			t.Fatalf("internal join guarded: %g", d)
		}
		if d := compoundDistance(l, connectionPoint{119, 1}); d > 2 || d < 0 {
			t.Fatalf("outer edge: %g", d)
		}
	}
	if _, err := compoundLayoutFor("factory-summer"); err == nil {
		t.Fatal("unbounded layout")
	}
}

func TestCompoundAffineUniformCentersAndAlpha(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 1024, 1024))
	for y := 400; y < 624; y++ {
		for x := 200; x < 824; x++ {
			src.SetNRGBA(x, y, color.NRGBA{210, 70, 25, 255})
		}
	}
	l, _ := compoundLayoutFor("depot-summer")
	got, tr, err := compoundAffine(src, l, .9)
	if err != nil {
		t.Fatal(err)
	}
	want := .9 * 160 / 1024
	if math.Abs(tr.Scale-want) > 1e-12 || math.Abs(tr.Offset[0]-(238-1024*want)/2) > 1e-12 {
		t.Fatalf("transform %+v", tr)
	}
	if got.Bounds() != image.Rect(0, 0, 952, 640) {
		t.Fatal(got.Bounds())
	}
	visible := 0
	for y := 0; y < 640; y++ {
		for x := 0; x < 952; x++ {
			c := got.NRGBAAt(x, y)
			if c.A > 0 {
				visible++
				if c.R != 210 || c.G != 70 || c.B != 25 {
					t.Fatalf("alpha halo %v", c)
				}
			} else if c != (color.NRGBA{}) {
				t.Fatal("hidden RGB")
			}
		}
	}
	if visible < 10000 {
		t.Fatal("lost art")
	}
	if _, _, err := compoundAffine(src, l, math.NaN()); err == nil {
		t.Fatal("NaN inset")
	}
}

func TestCompoundGuardRejectsArchitectureAndPreservesInternalRoof(t *testing.T) {
	l, _ := compoundLayoutFor("depot-summer")
	raw := image.NewNRGBA(image.Rect(0, 0, 952, 640))
	raw.SetNRGBA(119*4, 1*4, color.NRGBA{100, 80, 50, 255})
	if _, _, err := compoundGuard(raw, l); err == nil {
		t.Fatal("opaque boundary accepted")
	}
	raw.SetNRGBA(119*4, 1*4, color.NRGBA{})
	raw.SetNRGBA(83*4, 60*4, color.NRGBA{100, 80, 50, 255})
	guarded, stats, err := compoundGuard(raw, l)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw.Pix, guarded.Pix) || stats.GuardedOpaque != 0 {
		t.Fatal("internal architecture altered")
	}
}

// Mutating case spelling, source traversal, or pinned bytes must reject before output exists.
func TestCompoundStrictManifest(t *testing.T) {
	for _, data := range []string{
		`{"schema_version":1,"Schema_Version":1,"grounds":{},"scenes":[]}`,
		`{"schema_version":1,"grounds":{},"scenes":[],"scenes":[]}`,
		`{"schema_version":1,"grounds":{},"scenes":[],"unknown":0}`,
		`{"schema_version":1,"grounds":{},"scenes":[{"id":"depot-summer","artwork":{"Source":"x","source_sha256":"x"},"inset":0.9}]}`,
	} {
		var m compoundManifest
		if err := decodeCompoundManifest([]byte(data), &m); err == nil {
			t.Fatalf("accepted %s", data)
		}
	}
	for _, p := range []string{"", "/tmp/x", "../x", "source/../x", "source/original/x", "pilot/x", "source/PILOT/x", "source\\x"} {
		if err := compoundSafeRelative(p); err == nil {
			t.Fatalf("unsafe %q", p)
		}
	}
}

func TestCompoundCLIExclusivity(t *testing.T) {
	for _, args := range [][]string{{"-compound", ""}, {"-compound", "x", "-out", "y", "-candidate", ""}, {"-compound", "x", "-out", "y", "-connected", ""}, {"-compound", "x", "-out", "y", "pos"}} {
		var b bytes.Buffer
		if err := run(args, &b); err == nil || !bytes.Contains([]byte(err.Error()), []byte("-compound")) {
			t.Fatalf("%v: %v", args, err)
		}
	}
}

func TestCompoundActualPNGBundle(t *testing.T) {
	dir := t.TempDir()
	ground := image.NewNRGBA(image.Rect(0, 0, 384, 320))
	for y := range 320 {
		for x := range 384 {
			ground.SetNRGBA(x, y, color.NRGBA{uint8(x*7 + y*3), uint8(x*2 + y*11), uint8(x ^ y), 255})
		}
	}
	art := image.NewNRGBA(image.Rect(0, 0, 1024, 1024))
	for y := 450; y < 574; y++ {
		for x := 300; x < 724; x++ {
			art.SetNRGBA(x, y, color.NRGBA{uint8(x), uint8(y), uint8(x + y), 255})
		}
	}
	source := func(name string, im *image.NRGBA) connectionMaterial {
		var b bytes.Buffer
		if err := png.Encode(&b, im); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), b.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b.Bytes())
		return connectionMaterial{name, hex.EncodeToString(sum[:])}
	}
	g, a := source("ground.png", ground), source("art.png", art)
	m := compoundManifest{SchemaVersion: 1, Grounds: map[string]connectionMaterial{"summer": g, "winter": g}}
	for _, id := range []string{"depot-summer", "depot-winter", "hq-summer", "hq-winter", "factory-winter"} {
		m.Scenes = append(m.Scenes, compoundSceneInput{ID: id, Artwork: a, Inset: .9})
	}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "manifest.json")
	if err = os.WriteFile(manifest, b, 0600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "bundle")
	if err = run([]string{"-compound", manifest, "-out", out}, io.Discard); err != nil {
		t.Fatal(err)
	}
	// All comparisons use decoded exported bytes, independently blitted with Over.
	for _, id := range []string{"depot-summer", "depot-winter", "hq-summer", "hq-winter", "factory-winter"} {
		l, _ := compoundLayoutFor(id)
		edges := compoundBoundary(l)
		for _, density := range []int{2, 4} {
			full, err := loadPNG(filepath.Join(out, fmt.Sprintf("%s-scene-%dx.png", id, density)))
			if err != nil {
				t.Fatal(err)
			}
			assembled := image.NewNRGBA(full.Bounds())
			groundName := "grass"
			if l.Season == "winter" {
				groundName = "grass_w"
			}
			tileGround, err := loadPNG(filepath.Join(out, fmt.Sprintf("ground/%s-%dx.png", groundName, density)))
			if err != nil {
				t.Fatal(err)
			}
			expectedGround, err := exportConnectionTiles(connectionInputs{manifest: connectionManifest{Geometry: sharedConnectionGeometry()}, materials: map[string]*image.NRGBA{"grass": ground}}, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(tileGround.Pix, expectedGround[density].Pix) {
				t.Fatal("ground phase changed")
			}
			base := image.NewNRGBA(full.Bounds())
			for _, tile := range l.Tiles {
				at := image.Pt(int(tile.Origin[0])*density, int(tile.Origin[1])*density)
				draw.Draw(base, tileGround.Bounds().Add(at), tileGround, image.Point{}, draw.Over)
				got, err := loadPNG(filepath.Join(out, fmt.Sprintf("tiles/%s-%dx.png", tile.RuntimeID, density)))
				if err != nil {
					t.Fatal(err)
				}
				for y := range got.Bounds().Dy() {
					for x := range got.Bounds().Dx() {
						c := got.NRGBAAt(x, y)
						inside := connectionInside(sharedConnectionGeometry(), connectionPoint{(float64(x) + .5) / float64(density), (float64(y) + .5) / float64(density)})
						if inside && c.A != 255 {
							t.Fatal("owned pixel not opaque")
						}
						if !inside && c != (color.NRGBA{}) {
							t.Fatal("outside RGBA")
						}
					}
				}
				draw.Draw(assembled, got.Bounds().Add(at), got, image.Point{}, draw.Over)
			}
			if !bytes.Equal(full.Pix, assembled.Pix) {
				t.Fatalf("actual PNG assembly differs %s %dx", id, density)
			}
			for y := range full.Bounds().Dy() {
				for x := range full.Bounds().Dx() {
					d := compoundSignedDistance(l, edges, connectionPoint{(float64(x) + .5) / float64(density), (float64(y) + .5) / float64(density)})
					if d >= 0 && d < 2 && full.NRGBAAt(x, y) != base.NRGBAAt(x, y) {
						t.Fatal("quiet band ground phase mismatch")
					}
				}
			}
			for parity := range 2 {
				got, err := loadPNG(filepath.Join(out, fmt.Sprintf("%s-patch-parity%d-%dx.png", id, parity, density)))
				if err != nil {
					t.Fatal(err)
				}
				// Independent runtime-lattice surround construction, without production patch helpers.
				offset := image.Pt(-71, 0)
				if parity == 1 {
					offset = image.Pt(0, 40)
				}
				cells := map[image.Point]string{}
				for _, tile := range l.Tiles {
					p := image.Pt(int(tile.Origin[0]), int(tile.Origin[1])).Add(offset)
					cells[p] = tile.RuntimeID
				}
				for _, tile := range l.Tiles {
					p := image.Pt(int(tile.Origin[0]), int(tile.Origin[1])).Add(offset)
					for _, n := range []image.Point{{71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}} {
						q := p.Add(n)
						if _, ok := cells[q]; !ok {
							cells[q] = ""
						}
					}
				}
				minX, minY, maxX, maxY := 999, 999, -999, -999
				for p := range cells {
					minX = min(minX, p.X)
					minY = min(minY, p.Y)
					maxX = max(maxX, p.X+96)
					maxY = max(maxY, p.Y+80)
				}
				want := image.NewNRGBA(image.Rect(0, 0, (maxX-minX)*density, (maxY-minY)*density))
				for p := range cells {
					at := image.Pt((p.X-minX)*density, (p.Y-minY)*density)
					draw.Draw(want, tileGround.Bounds().Add(at), tileGround, image.Point{}, draw.Over)
				}
				for p, runtimeID := range cells {
					if runtimeID == "" {
						continue
					}
					tile, err := loadPNG(filepath.Join(out, fmt.Sprintf("tiles/%s-%dx.png", runtimeID, density)))
					if err != nil {
						t.Fatal(err)
					}
					at := image.Pt((p.X-minX)*density, (p.Y-minY)*density)
					draw.Draw(want, tile.Bounds().Add(at), tile, image.Point{}, draw.Over)
				}
				if got.Bounds() != want.Bounds() || !bytes.Equal(got.Pix, want.Pix) {
					t.Fatalf("actual parity patch mismatch %s parity%d %dx", id, parity, density)
				}
			}
		}
	}
	// Byte determinism across a fresh output directory includes source/manifest snapshots.
	second := filepath.Join(dir, "second")
	if err = runCompound(manifest, second, io.Discard); err != nil {
		t.Fatal(err)
	}
	if err = filepath.WalkDir(out, func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(out, path)
		if err != nil {
			return err
		}
		left, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		right, err := os.ReadFile(filepath.Join(second, rel))
		if err != nil {
			return err
		}
		if !bytes.Equal(left, right) {
			return fmt.Errorf("nondeterministic %s", rel)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err = runCompound(manifest, out, io.Discard); err == nil {
		t.Fatal("existing output accepted")
	}
}

func TestCompoundGalleryLinksTwoToFour(t *testing.T) {
	r := compoundReport{Scenes: []compoundSceneReport{{ID: "depot-summer", compoundLayout: compoundLayout{Tiles: []compoundTile{{RuntimeID: "cmpN", Outputs: []packOutput{{Path: "tiles/cmpN-2x.png"}, {Path: "tiles/cmpN-4x.png"}}}}}, Reference: []packOutput{{Path: "scene-2x.png"}, {Path: "scene-4x.png"}}, Patches: []compoundPatch{{Outputs: []packOutput{{Path: "patch-2x.png"}, {Path: "patch-4x.png"}}}}}}}
	var b bytes.Buffer
	if err := template.Must(template.New("gallery").Parse(compoundGallery)).Execute(&b, r); err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]string{{"tiles/cmpN-4x.png", "tiles/cmpN-2x.png"}, {"scene-4x.png", "scene-2x.png"}, {"patch-4x.png", "patch-2x.png"}} {
		want := fmt.Sprintf(`href="%s"><img`, pair[0])
		start := bytes.Index(b.Bytes(), []byte(want))
		if start < 0 {
			t.Fatal("missing 4x link")
		}
		tail := b.String()[start:]
		end := strings.Index(tail, "</a>")
		if end < 0 || !strings.Contains(tail[:end], `src="`+pair[1]+`"`) {
			t.Fatalf("2x preview must link to 4x: %v", pair)
		}
	}
}

func TestCompoundAxisSupportSignedLobesAndFractionalCenter(t *testing.T) {
	w := compoundAxis(30, 200, .25, 1.25)[10]
	if w.first != 25 || len(w.weights) != 24 {
		t.Fatalf("source-space radius must widen to 12: %+v", w)
	}
	sum, moment := 0., 0.
	negative := false
	for i, v := range w.weights {
		sum += v
		moment += float64(w.first+i) * v
		negative = negative || v < 0
	}
	if !negative || math.Abs(sum-1) > 1e-12 || math.Abs(moment-36.5) > 1e-12 {
		t.Fatalf("signed centered kernel sum=%g center=%g negative=%v", sum, moment, negative)
	}
	// Transparent extension keeps missing taps in normalization, unlike edge renormalization.
	edge := compoundAxis(1, 8, .5, .5)[0]
	sum = 0
	for _, v := range edge.weights {
		sum += v
	}
	if sum >= .9 {
		t.Fatalf("source edge renormalized: %g", sum)
	}
}

func compoundTestFixture(t *testing.T) (string, compoundManifest) {
	t.Helper()
	dir := t.TempDir()
	ground := image.NewNRGBA(image.Rect(0, 0, 384, 320))
	draw.Draw(ground, ground.Bounds(), image.NewUniform(color.NRGBA{61, 91, 43, 255}), image.Point{}, draw.Src)
	art := image.NewNRGBA(image.Rect(0, 0, 1024, 1024))
	for y := 480; y < 544; y++ {
		for x := 480; x < 544; x++ {
			art.SetNRGBA(x, y, color.NRGBA{200, 30, 10, 255})
		}
	}
	ref := func(name string, im *image.NRGBA) connectionMaterial {
		var b bytes.Buffer
		if err := png.Encode(&b, im); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), b.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b.Bytes())
		return connectionMaterial{name, hex.EncodeToString(sum[:])}
	}
	m := compoundManifest{SchemaVersion: 1, Grounds: map[string]connectionMaterial{"summer": ref("ground.png", ground)}, Scenes: []compoundSceneInput{{ID: "depot-summer", Artwork: ref("art.png", art), Inset: .9}}}
	path := filepath.Join(dir, "manifest.json")
	compoundTestWriteManifest(t, path, m)
	return path, m
}
func compoundTestWriteManifest(t *testing.T, path string, m compoundManifest) {
	t.Helper()
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestCompoundPinnedSnapshots(t *testing.T) {
	path, _ := compoundTestFixture(t)
	in, err := loadCompoundInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	original := append([]byte(nil), in.Bytes...)
	art := append([]byte(nil), in.Artwork["depot-summer"].Bytes...)
	for _, name := range []string{"manifest.json", "art.png", "ground.png"} {
		if err = os.WriteFile(filepath.Join(filepath.Dir(path), name), []byte("changed after pin"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	bundle, err := buildCompoundBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, a := range bundle.artifacts {
		if a.name == "manifest.json" {
			found++
			if !bytes.Equal(a.data, original) {
				t.Fatal("manifest reread")
			}
		}
		if a.name == "sources/depot-summer.png" {
			found++
			if !bytes.Equal(a.data, art) {
				t.Fatal("source reread")
			}
		}
	}
	if found != 2 {
		t.Fatal("snapshots missing")
	}
}

func TestCompoundInputRejections(t *testing.T) {
	path, base := compoundTestFixture(t)
	for _, tc := range []struct {
		name   string
		change func(*compoundManifest)
	}{
		{"schema", func(m *compoundManifest) { m.SchemaVersion = 2 }},
		{"missing-ground", func(m *compoundManifest) { m.Grounds = map[string]connectionMaterial{} }},
		{"unused-ground", func(m *compoundManifest) {
			m.Grounds = map[string]connectionMaterial{"summer": base.Grounds["summer"], "winter": base.Grounds["summer"]}
		}},
		{"duplicate", func(m *compoundManifest) { m.Scenes = append(m.Scenes, m.Scenes[0]) }},
		{"unsupported", func(m *compoundManifest) { m.Scenes[0].ID = "factory-summer" }},
		{"inset", func(m *compoundManifest) { m.Scenes[0].Inset = .49 }},
		{"hash", func(m *compoundManifest) { m.Scenes[0].Artwork.SourceSHA256 = strings.Repeat("0", 64) }},
		{"traversal", func(m *compoundManifest) { m.Scenes[0].Artwork.Source = "../art.png" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := base
			m.Scenes = append([]compoundSceneInput(nil), base.Scenes...)
			tc.change(&m)
			compoundTestWriteManifest(t, path, m)
			out := filepath.Join(filepath.Dir(path), "reject-"+tc.name)
			if err := runCompound(path, out, io.Discard); err == nil {
				t.Fatal("accepted invalid input")
			}
			if _, err := os.Lstat(out); !os.IsNotExist(err) {
				t.Fatal("invalid input created output")
			}
		})
	}
	compoundTestWriteManifest(t, path, base)
	// The lexical path is innocuous; only resolution reveals the forbidden directory.
	forbidden := filepath.Join(filepath.Dir(path), "original")
	if err := os.Mkdir(forbidden, 0700); err != nil {
		t.Fatal(err)
	}
	art, err := os.ReadFile(filepath.Join(filepath.Dir(path), "art.png"))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(forbidden, "art.png"), art, 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink("original/art.png", filepath.Join(filepath.Dir(path), "alias.png")); err != nil {
		t.Fatal(err)
	}
	m := base
	m.Scenes = append([]compoundSceneInput(nil), base.Scenes...)
	m.Scenes[0].Artwork.Source = "alias.png"
	compoundTestWriteManifest(t, path, m)
	if _, err = loadCompoundInputs(path); err == nil {
		t.Fatal("resolved original source accepted")
	}
}

type compoundFailWriter struct{}

func (compoundFailWriter) Write([]byte) (int, error) {
	return 0, errors.New("intentional writer failure")
}
func TestCompoundRollbackAndOpaquePreflight(t *testing.T) {
	path, m := compoundTestFixture(t)
	out := filepath.Join(filepath.Dir(path), "output")
	if err := runCompound(path, out, compoundFailWriter{}); err == nil {
		t.Fatal("writer failure lost")
	}
	if _, err := os.Lstat(out); !os.IsNotExist(err) {
		t.Fatal("failed export directory retained")
	}
	// Opaque rectangle covers outer union boundary after fit: must reject before any output.
	art := image.NewNRGBA(image.Rect(0, 0, 1024, 1024))
	for y := 10; y < 1014; y++ {
		for x := 10; x < 1014; x++ {
			art.SetNRGBA(x, y, color.NRGBA{200, 30, 10, 255})
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, art); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), "art.png"), b.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b.Bytes())
	m.Scenes[0].Artwork.SourceSHA256 = hex.EncodeToString(sum[:])
	compoundTestWriteManifest(t, path, m)
	err := runCompound(path, out, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "unsafe compound framing") {
		t.Fatalf("opaque preflight: %v", err)
	}
	if _, err = os.Lstat(out); !os.IsNotExist(err) {
		t.Fatal("preflight created output")
	}
}

func TestCompoundAffineTallSourceScalarReference(t *testing.T) {
	// A tall legal source exercises the cheaper vertical-first intermediate. A
	// scalar 2D convolution checks premultiplication/clamping independently of
	// the separable pass order, including fractional source center positions.
	src := image.NewNRGBA(image.Rect(0, 0, 64, 2048))
	for y := range 2048 {
		for x := range 64 {
			src.SetNRGBA(x, y, color.NRGBA{uint8(x*17 + y*3), uint8(y * 7), uint8(x + y), uint8((x*9 + y*5) % 256)})
		}
	}
	l, _ := compoundLayoutFor("depot-summer")
	got, tr, err := compoundAffine(src, l, .9)
	if err != nil {
		t.Fatal(err)
	}
	wx := compoundAxis(952, 64, tr.Scale*4, tr.Offset[0]*4)
	wy := compoundAxis(640, 2048, tr.Scale*4, tr.Offset[1]*4)
	for _, p := range []image.Point{{473, 40}, {477, 159}, {471, 320}, {480, 550}} {
		var q [4]float64
		for j, v := range wy[p.Y].weights {
			for i, u := range wx[p.X].weights {
				c := src.NRGBAAt(wx[p.X].first+i, wy[p.Y].first+j)
				a := u * v * float64(c.A)
				q[0] += a * float64(c.R)
				q[1] += a * float64(c.G)
				q[2] += a * float64(c.B)
				q[3] += a
			}
		}
		want := color.NRGBA{}
		if q[3] > 0 {
			want.A = uint8(math.Floor(min(q[3], 255) + .5))
			if want.A > 0 {
				ch := func(v float64) uint8 { return uint8(math.Floor(min(max(v, 0), 255*q[3])/q[3] + .5)) }
				want.R = ch(q[0])
				want.G = ch(q[1])
				want.B = ch(q[2])
			}
		}
		if c := got.NRGBAAt(p.X, p.Y); c != want {
			t.Fatalf("scalar source-alpha reference at%v: got%v want%v", p, c, want)
		}
	}
}

func TestCompoundTapBudgetRejectsPathologicalDimensionsWithoutAllocation(t *testing.T) {
	for _, source := range []dimensions{{1, 67108864}, {67108864, 1}} {
		scale := 4 * .9 * min(238/float64(source.Width), 160/float64(source.Height))
		if err := compoundCheckTapBudget(952, 640, scale); err == nil || !strings.Contains(err.Error(), "tap budget") {
			t.Fatalf("pathological source %+v accepted: %v", source, err)
		}
	}
	for _, scale := range []float64{0, -1, math.NaN(), math.Inf(1), math.SmallestNonzeroFloat64} {
		if err := compoundCheckTapBudget(952, 640, scale); err == nil {
			t.Fatalf("unsafe scale %g accepted", scale)
		}
	}
	for _, scene := range []dimensions{{238, 160}, {238, 240}} {
		scale := 4 * .9 * min(float64(scene.Width)/1536, float64(scene.Height)/1024)
		if err := compoundCheckTapBudget(scene.Width*4, scene.Height*4, scale); err != nil {
			t.Fatal(err)
		}
	}
	// Include transparent-extension work: a one-pixel source still incurs the
	// complete widened kernel, not merely its single retained source tap.
	if err := compoundCheckTapBudget(952, 640, .001); err == nil {
		t.Fatal("full-support work omitted")
	}
	if err := compoundCheckTapBudget(0, 640, 1); err == nil {
		t.Fatal("zero dimension")
	}
}
