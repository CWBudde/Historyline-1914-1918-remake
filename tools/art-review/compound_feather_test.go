package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// compoundFeatherCorridorTest is the entry corridor of the synthetic tests:
// the 22 px road's half width with its 2 px border, plus 2 world px.
const compoundFeatherCorridorTest = 22/2 + 2 + 2

// compoundEntryLateral returns how far p lies beside the entry port along the
// port's edge and how deep it lies inside the port, in world px.
func compoundEntryLateral(t *testing.T, l compoundLayout, p connectionPoint) (lateral, depth float64) {
	t.Helper()
	origin, port, err := compoundEntryPort(l)
	if err != nil {
		t.Fatal(err)
	}
	g := sharedConnectionGeometry()
	q := p.sub(origin).sub(g.Ports[port-1])
	in := g.Neighbors[port-1].mul(-1 / g.Neighbors[port-1].length())
	return math.Abs(q[0]*in[1] - q[1]*in[0]), q[0]*in[0] + q[1]*in[1]
}

// compoundFeatherContour marches inward from points along the outer boundary
// (outside the entry corridor, away from the corners) and returns each one's
// depth in world px where the 2x alpha first reaches 50 %.
func compoundFeatherContour(t *testing.T, l compoundLayout, art2 *image.NRGBA, corridor float64) []float64 {
	t.Helper()
	var depths []float64
	edges := compoundBoundary(l)
	for _, e := range edges {
		v := e.b.sub(e.a)
		n := v.length()
		u := v.mul(1 / n)
		in := connectionPoint{-u[1], u[0]}
		if mid := e.a.add(v.mul(.5)); compoundSignedDistance(l, edges, mid.add(in.mul(1))) < 0 {
			in = in.mul(-1)
		}
		for s := 3.; s <= n-3; s++ {
			p := e.a.add(u.mul(s))
			if lat, depth := compoundEntryLateral(t, l, p); lat <= corridor+1 && depth < 1 {
				continue
			}
			found := math.NaN()
			for d := 0.; d <= 16; d += .25 {
				q := p.add(in.mul(d))
				x, y := int(q[0]*2), int(q[1]*2)
				if image.Pt(x, y).In(art2.Bounds()) && art2.NRGBAAt(x, y).A >= 128 {
					// Its distance to the boundary, which near a corner
					// is less than the depth along this normal.
					found = compoundSignedDistance(l, edges, q)
					break
				}
			}
			depths = append(depths, found)
		}
	}
	return depths
}

// An opaque ruin scene fades out towards its outer boundary along an
// irregular, deterministic line; the entry corridor keeps the road open to the
// port, and the interior stays untouched.
func TestCompoundFeatherFadesOuterBoundary(t *testing.T) {
	l, err := compoundLayoutFor("ruin-winter")
	if err != nil {
		t.Fatal(err)
	}
	raw := image.NewNRGBA(image.Rect(0, 0, l.Width*4, l.Height*4))
	draw.Draw(raw, raw.Bounds(), image.NewUniform(color.NRGBA{220, 224, 228, 255}), image.Point{}, draw.Src)
	before := append([]byte(nil), raw.Pix...)
	const width = 9.
	art4, stats, err := compoundFeatherEdge(raw, l, compoundGuardReport{}, width, compoundFeatherCorridorTest, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw.Pix, before) {
		t.Fatal("feather changed its input")
	}
	again, _, err := compoundFeatherEdge(raw, l, compoundGuardReport{}, width, compoundFeatherCorridorTest, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(art4.Pix, again.Pix) {
		t.Fatal("feather is not deterministic")
	}
	if stats.ClippedPixels == 0 || stats.ClippedAlphaMass == 0 || stats.FeatherWidth != width || stats.FeatherCorridor != compoundFeatherCorridorTest {
		t.Fatalf("feather report: %+v", stats)
	}
	art2, err := resizeLanczos(art4, l.Width*2, l.Height*2)
	if err != nil {
		t.Fatal(err)
	}
	edges := compoundBoundary(l)
	boundary, corridor, interior := 0, 0, 0
	for y := range l.Height * 2 {
		for x := range l.Width * 2 {
			p := connectionPoint{(float64(x) + .5) / 2, (float64(y) + .5) / 2}
			d := compoundSignedDistance(l, edges, p)
			if d < 0 {
				continue
			}
			a := art2.NRGBAAt(x, y).A
			lat, depth := compoundEntryLateral(t, l, p)
			switch {
			case lat <= compoundFeatherCorridorTest-4 && depth < width+2:
				corridor++
				if a < 250 {
					t.Fatalf("entry corridor faded at (%.1f,%.1f): alpha %d", p[0], p[1], a)
				}
			case d <= 2 && !(lat <= compoundFeatherCorridorTest && depth < width+2):
				boundary++
				if a > 64 {
					t.Fatalf("hard edge at (%.1f,%.1f), %.2f px inside: alpha %d", p[0], p[1], d, a)
				}
			case d >= width+2:
				interior++
				if a != 255 {
					t.Fatalf("interior faded at (%.1f,%.1f): alpha %d", p[0], p[1], a)
				}
			}
		}
	}
	if boundary < 500 || corridor < 50 || interior < 10000 {
		t.Fatalf("too few samples: boundary %d, corridor %d, interior %d", boundary, corridor, interior)
	}
	// The 50 % contour wanders: it is no soft copy of the hex outline.
	depths := compoundFeatherContour(t, l, art2, compoundFeatherCorridorTest)
	mean, lo, hi := 0., math.Inf(1), math.Inf(-1)
	for _, d := range depths {
		if math.IsNaN(d) {
			t.Fatal("50 % contour missing within 16 world px")
		}
		mean += d
		lo, hi = min(lo, d), max(hi, d)
	}
	mean /= float64(len(depths))
	variance := 0.
	for _, d := range depths {
		variance += (d - mean) * (d - mean)
	}
	sd := math.Sqrt(variance / float64(len(depths)))
	t.Logf("50 %% contour over %d boundary samples: mean %.2f, sd %.2f, range %.2f..%.2f world px", len(depths), mean, sd, lo, hi)
	if sd < 1 || hi-lo < 3.5 || lo < 2 || hi > width+1 {
		t.Fatalf("50 %% contour too regular or outside the band: sd %.2f, range %.2f..%.2f", sd, lo, hi)
	}
}

func TestCompoundFeatherValidation(t *testing.T) {
	square := []connectionPoint{{10, 10}, {20, 10}, {20, 20}, {10, 20}}
	for _, tc := range []struct {
		id      string
		framing string
		width   float64
		entry   bool
		keep    [][]connectionPoint
		wantErr bool
	}{
		{"ruin-summer", "reference-layout", 9, true, nil, false},
		{"ruin-winter", "reference-layout", 9, true, nil, false},
		{"ruin-summer", "reference-layout", 0, false, nil, false},
		{"ruin-summer", "reference-layout", 9, false, nil, true},
		{"ruin-summer", "", 9, false, nil, true},
		{"ruin-summer", "reference-layout", 3, true, nil, true},
		{"ruin-summer", "reference-layout", 13, true, nil, true},
		{"ruin-summer", "reference-layout", math.NaN(), true, nil, true},
		{"depot-summer", "reference-layout", 9, true, nil, false},
		{"hq-winter", "reference-layout", 9, true, [][]connectionPoint{square}, false},
		{"factory-summer", "reference-layout", 9, true, [][]connectionPoint{square, square}, false},
		{"depot-winter", "reference-layout", 9, false, nil, true},
		{"mountain-summer", "", 9, false, nil, true},
		// Keep polygons need a feather, three to 32 points and points near the box.
		{"depot-summer", "reference-layout", 0, true, [][]connectionPoint{square}, true},
		{"depot-summer", "reference-layout", 9, true, [][]connectionPoint{square[:2]}, true},
		{"depot-summer", "reference-layout", 9, true, [][]connectionPoint{{{10, 10}, {20, 10}, {20, 169}}}, true},
		{"depot-summer", "reference-layout", 9, true, [][]connectionPoint{{{10, 10}, {math.NaN(), 10}, {20, 20}}}, true},
		{"depot-summer", "reference-layout", 9, true, [][]connectionPoint{{{-8, -8}, {246, -8}, {246, 168}}}, false},
	} {
		l, err := compoundLayoutFor(tc.id)
		if err != nil {
			t.Fatal(err)
		}
		if err := compoundValidateFeather(l, tc.framing, tc.width, tc.entry, tc.keep); (err != nil) != tc.wantErr {
			t.Errorf("%s framing %q width %v entry %v keep %v: error %v, want error %v", tc.id, tc.framing, tc.width, tc.entry, tc.keep, err, tc.wantErr)
		}
	}
}

// Keep polygons protect what reaches the boundary: inside them the artwork
// keeps its alpha exactly, beyond the keep ramp the boundary fades as before.
func TestCompoundFeatherKeepPolygons(t *testing.T) {
	l, err := compoundLayoutFor("depot-winter")
	if err != nil {
		t.Fatal(err)
	}
	raw := image.NewNRGBA(image.Rect(0, 0, l.Width*4, l.Height*4))
	draw.Draw(raw, raw.Bounds(), image.NewUniform(color.NRGBA{120, 110, 100, 255}), image.Point{}, draw.Src)
	// A shed across the W cell's bottom edge (y 120) and a tower on the E cell's top edge (y 40).
	keep := [][]connectionPoint{{{50, 106}, {70, 106}, {70, 124}, {50, 124}}, {{180, 36}, {196, 36}, {200, 50}, {176, 50}}}
	const width = 9.
	plain, _, err := compoundFeatherEdge(raw, l, compoundGuardReport{}, width, compoundFeatherCorridorTest, nil)
	if err != nil {
		t.Fatal(err)
	}
	art4, stats, err := compoundFeatherEdge(raw, l, compoundGuardReport{}, width, compoundFeatherCorridorTest, keep)
	if err != nil {
		t.Fatal(err)
	}
	if stats.ClippedPixels == 0 {
		t.Fatal("nothing feathered")
	}
	edges := compoundBoundary(l)
	kept, faded, eased := 0, 0, 0
	for y := range l.Height * 4 {
		for x := range l.Width * 4 {
			p := connectionPoint{(float64(x) + .5) / 4, (float64(y) + .5) / 4}
			d := compoundSignedDistance(l, edges, p)
			if d < 0 || d >= width {
				continue
			}
			a, b := art4.NRGBAAt(x, y), plain.NRGBAAt(x, y)
			switch k := compoundKeepDistance(keep, p); {
			case k == 0:
				kept++
				if a != raw.NRGBAAt(x, y) {
					t.Fatalf("keep polygon changed at (%.2f,%.2f): %v", p[0], p[1], a)
				}
			case k >= compoundKeepRamp:
				if a != b {
					t.Fatalf("beyond the keep ramp at (%.2f,%.2f): %v, without keep %v", p[0], p[1], a, b)
				}
				if a.A < 255 {
					faded++
				}
			default:
				if a.A < b.A {
					t.Fatalf("keep ramp fades more at (%.2f,%.2f): %d < %d", p[0], p[1], a.A, b.A)
				}
				if a.A > b.A {
					eased++
				}
			}
		}
	}
	if kept < 1000 || faded < 10000 || eased < 100 {
		t.Fatalf("too few samples: kept %d, faded %d, eased %d", kept, faded, eased)
	}
}

// Through the manifest and the whole bundle: the feathered ruin keeps the
// connected road at its north port and fades its other edges into the ground.
func TestCompoundFeatherBundle(t *testing.T) {
	roadDir, cm := connectedFixture(t)
	f := cm["families"].([]any)[0].(map[string]any)
	f["width"], f["border"] = 22, 2
	green := image.NewNRGBA(image.Rect(0, 0, 384, 320))
	draw.Draw(green, green.Bounds(), image.NewUniform(color.NRGBA{40, 200, 40, 255}), image.Point{}, draw.Src)
	cm["materials"].(map[string]any)["road"] = unitTestPNG(t, filepath.Join(roadDir, "green.png"), green)
	roadPath := writeConnectedManifest(t, roadDir, cm)
	roadBytes, err := os.ReadFile(roadPath)
	if err != nil {
		t.Fatal(err)
	}
	roadSum := sha256.Sum256(roadBytes)

	path, m := compoundTestFixture(t)
	l, err := compoundLayoutFor("ruin-summer")
	if err != nil {
		t.Fatal(err)
	}
	// An opaque scene with the reference road to the north edge.
	art := compoundReferenceArtwork(l)
	rest := image.NewNRGBA(art.Bounds())
	draw.Draw(rest, rest.Bounds(), image.NewUniform(color.NRGBA{150, 120, 90, 255}), image.Point{}, draw.Src)
	draw.Draw(rest, rest.Bounds(), art, image.Point{}, draw.Over)
	rest.SetNRGBA(0, 0, color.NRGBA{}) // the artwork must have a transparent pixel
	var encoded bytes.Buffer
	if err = png.Encode(&encoded, rest); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(filepath.Dir(path), "ruin.png"), encoded.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	artSum := sha256.Sum256(encoded.Bytes())
	rel, err := filepath.Rel(filepath.Dir(path), roadPath)
	if err != nil {
		t.Fatal(err)
	}
	entry := &connectionMaterial{Source: rel, SourceSHA256: hex.EncodeToString(roadSum[:])}
	scene := compoundSceneInput{ID: "ruin-summer", Artwork: connectionMaterial{Source: "ruin.png", SourceSHA256: hex.EncodeToString(artSum[:])}, Inset: 1, Framing: "reference-layout", EntryRoad: entry, EdgeFeather: 9}
	m.Scenes = []compoundSceneInput{scene}
	compoundTestWriteManifest(t, path, m)
	in, err := loadCompoundInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	b, err := buildCompoundBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	report := b.report.Scenes[0]
	if report.EdgeFeather != 9 || report.Guard.FeatherWidth != 9 || report.Guard.FeatherCorridor != compoundFeatherCorridorTest || report.Guard.ClippedPixels == 0 {
		t.Fatalf("feather missing from the report: %+v", report)
	}
	for _, density := range []int{2, 4} {
		whole := b.pixels[fmt.Sprintf("ruin-summer-scene-%dx.png", density)]
		if c := whole.NRGBAAt(119*density, 0); c.G < 170 || c.R > 70 || c.A != 255 {
			t.Fatalf("%dx gate does not start as the road tile: %v", density, c)
		}
		// Beside the corridor on the north edge and on the west tip only the
		// ground (61,91,43) shows.
		for _, p := range []connectionPoint{{100, 0.5}, {138, 0.5}, {1.5, 80}, {236.5, 80}, {119, 159.5}} {
			c := whole.NRGBAAt(int(p[0]*float64(density)), int(p[1]*float64(density)))
			if c.R > 75 || c.G < 85 {
				t.Errorf("%dx artwork at the boundary (%.0f,%.1f): %v", density, p[0], p[1], c)
			}
		}
	}
	// Keep polygons need the feather.
	scene.EdgeFeather, scene.EdgeKeep = 0, [][]connectionPoint{{{10, 10}, {20, 10}, {20, 20}}}
	m.Scenes = []compoundSceneInput{scene}
	compoundTestWriteManifest(t, path, m)
	if _, err := loadCompoundInputs(path); err == nil {
		t.Error("edge_keep accepted without edge_feather")
	}
}
