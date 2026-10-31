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
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompoundReferenceFramingPreservesBoundaryArtwork(t *testing.T) {
	l, err := compoundLayoutFor("depot-summer")
	if err != nil {
		t.Fatal(err)
	}
	raw, _, err := compoundAffine(compoundReferenceArtwork(l), l, 1, compoundFit{Fill: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = compoundFrame(raw, l, ""); err == nil {
		t.Fatal("default framing accepted opaque boundary road and fence")
	}
	before := append([]byte(nil), raw.Pix...)
	framed, stats, err := compoundFrame(raw, l, "reference-layout")
	if err != nil {
		t.Fatal(err)
	}
	if framed != raw || !bytes.Equal(framed.Pix, before) {
		t.Fatal("reference framing masked or faded artwork before resampling")
	}
	if stats.GuardedOpaque == 0 || stats.BoundaryPixels == 0 || stats.MaxBoundaryAlpha != 255 || stats.MinimumOpaqueDistance == nil || *stats.MinimumOpaqueDistance >= 0 {
		t.Fatalf("missing honest boundary diagnostics: %+v", stats)
	}
	if stats.ClippedPixels != 0 || stats.ClippedAlphaMass != 0 || !strings.Contains(stats.Policy, "diagnostic only") {
		t.Fatalf("reference framing claims clipping or a guard pass: %+v", stats)
	}
	if _, _, err = compoundFrame(raw, l, "reference-layout-typo"); err == nil {
		t.Fatal("unknown framing accepted by renderer")
	}
}

func TestCompoundReferenceFramingManifestValidation(t *testing.T) {
	path, base := compoundTestFixture(t)
	for _, tc := range []struct {
		id      string
		framing string
		wantErr bool
	}{
		{"depot-summer", "", false},
		{"hq-summer", "reference-layout", false},
		{"hq-winter", "reference-layout", false},
		{"depot-summer", "reference-layout", false},
		{"depot-winter", "reference-layout", false},
		{"ruin-summer", "reference-layout", false},
		{"ruin-winter", "reference-layout", false},
		{"factory-summer", "reference-layout", false},
		{"factory-winter", "reference-layout", false},
		{"depot-summer", "unknown", true},
		{"mountain-summer", "reference-layout", true},
		{"mountain-single-summer", "reference-layout", true},
		{"hamlet-summer", "reference-layout", true},
		{"water-stones", "reference-layout", true},
	} {
		t.Run(tc.id+"/"+tc.framing, func(t *testing.T) {
			m := base
			m.Scenes = []compoundSceneInput{{ID: tc.id, Artwork: base.Scenes[0].Artwork, Inset: 1, Framing: tc.framing}}
			l, err := compoundLayoutFor(tc.id)
			if err != nil {
				t.Fatal(err)
			}
			m.Grounds = map[string]connectionMaterial{l.Season: base.Grounds["summer"]}
			compoundTestWriteManifest(t, path, m)
			_, err = loadCompoundInputs(path)
			if (err != nil) != tc.wantErr {
				t.Fatalf("load error = %v, want error %v", err, tc.wantErr)
			}
		})
	}
	data, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"Framing":"reference-layout",`, `"framing":17,`, `"framing":null,`, `"framing":"","framing":"reference-layout",`} {
		bad := bytes.Replace(data, []byte(`"inset":`), []byte(field+`"inset":`), 1)
		var m compoundManifest
		if err = decodeCompoundManifest(bad, &m); err == nil {
			t.Fatalf("accepted noncanonical framing: %s", field)
		}
	}
}

func TestCompoundShiftManifestValidation(t *testing.T) {
	path, base := compoundTestFixture(t)
	for _, tc := range []struct {
		framing string
		shift   []float64
		wantErr bool
	}{
		{"reference-layout", []float64{-0.77, 0}, false},
		{"reference-layout", []float64{2, -2}, false},
		{"reference-layout", []float64{8.5, 0}, true},
		{"reference-layout", []float64{1}, true},
		{"", []float64{0.5, 0}, true},
	} {
		m := base
		m.Scenes = []compoundSceneInput{{ID: "depot-summer", Artwork: base.Scenes[0].Artwork, Inset: 1, Framing: tc.framing, Shift: tc.shift}}
		m.Grounds = map[string]connectionMaterial{"summer": base.Grounds["summer"]}
		compoundTestWriteManifest(t, path, m)
		if _, err := loadCompoundInputs(path); (err != nil) != tc.wantErr {
			t.Errorf("framing %q shift %v: error %v, want error %v", tc.framing, tc.shift, err, tc.wantErr)
		}
	}
	// A shift must keep the filled artwork covering the box.
	l, err := compoundLayoutFor("depot-summer")
	if err != nil {
		t.Fatal(err)
	}
	src := compoundReferenceArtwork(l)
	if _, _, err := compoundAffine(src, l, 1, compoundFit{Fill: true, Shift: connectionPoint{0, 0.5}}); err == nil {
		t.Error("a shift that uncovers the box edge was accepted")
	}
	if _, _, err := compoundAffine(src, l, 1, compoundFit{Shift: connectionPoint{0.5, 0}}); err == nil {
		t.Error("a shift without fill was accepted")
	}
	// Zoom enlarges the filled artwork, so a shift that would uncover the box
	// edge fits once there is slack.
	if _, _, err := compoundAffine(src, l, 1, compoundFit{Fill: true, Zoom: 1.035, Shift: connectionPoint{3.5, 2.4}}); err != nil {
		t.Errorf("zoomed shift within the slack: %v", err)
	}
	if _, _, err := compoundAffine(src, l, 1, compoundFit{Zoom: 1.035}); err == nil {
		t.Error("a zoom without fill was accepted")
	}
	for _, tc := range []struct {
		framing string
		zoom    float64
		wantErr bool
	}{{"reference-layout", 1.035, false}, {"reference-layout", 1.2, true}, {"reference-layout", 0.9, true}, {"", 1.035, true}} {
		m := base
		m.Scenes = []compoundSceneInput{{ID: "depot-summer", Artwork: base.Scenes[0].Artwork, Inset: 1, Framing: tc.framing, Zoom: tc.zoom}}
		m.Grounds = map[string]connectionMaterial{"summer": base.Grounds["summer"]}
		compoundTestWriteManifest(t, path, m)
		if _, err := loadCompoundInputs(path); (err != nil) != tc.wantErr {
			t.Errorf("framing %q zoom %v: error %v, want error %v", tc.framing, tc.zoom, err, tc.wantErr)
		}
	}
}

// The entry street turns into the connected road at the gate: the road tile's
// own pixels at the port, the painted street beyond the blend.
func TestCompoundEntryRoadBlend(t *testing.T) {
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
	l, err := compoundLayoutFor("depot-summer")
	if err != nil {
		t.Fatal(err)
	}
	var encoded bytes.Buffer
	if err = png.Encode(&encoded, compoundReferenceArtwork(l)); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(filepath.Dir(path), "depot.png"), encoded.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	artSum := sha256.Sum256(encoded.Bytes())
	rel, err := filepath.Rel(filepath.Dir(path), roadPath)
	if err != nil {
		t.Fatal(err)
	}
	entry := &connectionMaterial{Source: rel, SourceSHA256: hex.EncodeToString(roadSum[:])}
	m.Scenes = []compoundSceneInput{{ID: "depot-summer", Artwork: connectionMaterial{Source: "depot.png", SourceSHA256: hex.EncodeToString(artSum[:])}, Inset: 1, Framing: "reference-layout", EntryRoad: entry}}
	compoundTestWriteManifest(t, path, m)
	in, err := loadCompoundInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	b, err := buildCompoundBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	for _, density := range []int{2, 4} {
		whole := b.pixels[fmt.Sprintf("depot-summer-scene-%dx.png", density)]
		if c := whole.NRGBAAt(119*density, 0); c.G < 170 || c.R > 70 || c.A != 255 {
			t.Fatalf("%dx gate does not start as the road tile: %v", density, c)
		}
		if c := whole.NRGBAAt(119*density, (compoundEntryBlend+2)*density); c.R < 70 || c.R > 90 || c.G > 92 {
			t.Fatalf("%dx painted street replaced beyond the blend: %v", density, c)
		}
	}

	// A scene without an entry cell, or a framing that is not the reference
	// layout, cannot take an entry road.
	for _, s := range []compoundSceneInput{
		{ID: "mountain-summer", Artwork: m.Scenes[0].Artwork, Inset: 1, EntryRoad: entry},
		{ID: "depot-summer", Artwork: m.Scenes[0].Artwork, Inset: 1, EntryRoad: entry},
	} {
		m.Scenes = []compoundSceneInput{s}
		compoundTestWriteManifest(t, path, m)
		if _, err := loadCompoundInputs(path); err == nil {
			t.Errorf("%s with framing %q accepted entry_road", s.ID, s.Framing)
		}
	}
}

// A road to the north/east canvas edge, a fence one world pixel from the
// indentation, and an outside patch distinguish full reference framing from a
// relaxed fade. Like the painted masters, the canvas has another aspect than
// the scene box (wider for the depot, square for the HQ, narrower for the
// factory), so letterboxing would leave a band between the road and the edge.
func compoundReferenceArtwork(l compoundLayout) *image.NRGBA {
	w, h := 1536, 1024
	switch l.Layout {
	case "hq":
		w, h = 1254, 1254
	case "factory":
		w, h = 1400, 1000
	}
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	// World rectangles through the fill transform the renderer uses.
	k := max(float64(l.Width)/float64(w), float64(l.Height)/float64(h))
	ox, oy := (float64(l.Width)-float64(w)*k)/2, (float64(l.Height)-float64(h)*k)/2
	rect := func(x0, y0, x1, y1 float64) image.Rectangle {
		return image.Rect(int(math.Floor((x0-ox)/k)), int(math.Floor((y0-oy)/k)), int(math.Ceil((x1-ox)/k)), int(math.Ceil((y1-oy)/k)))
	}
	road := rect(111, -20, 127, 100)
	if l.Layout == "factory" {
		road = rect(100, 74, float64(l.Width)+20, 86)
	}
	draw.Draw(im, road, image.NewUniform(color.NRGBA{80, 82, 83, 255}), image.Point{}, draw.Src)
	draw.Draw(im, rect(27, 40, 211, 44), image.NewUniform(color.NRGBA{190, 40, 30, 255}), image.Point{}, draw.Src)
	draw.Draw(im, rect(0, 0, 6, 6), image.NewUniform(color.NRGBA{240, 10, 240, 255}), image.Point{}, draw.Src)
	return im
}

func TestCompoundReferenceFramingActualPNGAssembly(t *testing.T) {
	path, m := compoundTestFixture(t)
	m.Grounds["winter"] = m.Grounds["summer"]
	m.Scenes = nil
	for _, id := range []string{"hq-summer", "hq-winter", "depot-summer", "depot-winter", "factory-summer", "factory-winter"} {
		l, err := compoundLayoutFor(id)
		if err != nil {
			t.Fatal(err)
		}
		var encoded bytes.Buffer
		if err = png.Encode(&encoded, compoundReferenceArtwork(l)); err != nil {
			t.Fatal(err)
		}
		name := id + ".png"
		if err = os.WriteFile(filepath.Join(filepath.Dir(path), name), encoded.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(encoded.Bytes())
		m.Scenes = append(m.Scenes, compoundSceneInput{ID: id, Artwork: connectionMaterial{Source: name, SourceSHA256: hex.EncodeToString(hash[:])}, Inset: 1, Framing: "reference-layout"})
	}
	compoundTestWriteManifest(t, path, m)
	in, err := loadCompoundInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	b, err := buildCompoundBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.report.Profile, "Per-scene profiles") {
		t.Fatal("top-level profile incorrectly claims universal strict guard")
	}
	for _, scene := range b.report.Scenes {
		t.Run(scene.ID, func(t *testing.T) {
			if scene.Inset != 1 || scene.Framing != "reference-layout" || scene.Guard.GuardedOpaque == 0 || scene.Guard.ClippedPixels != 0 {
				t.Fatalf("wrong reference framing report: %+v", scene)
			}
			if strings.Contains(scene.Profile, "reject any sampled alpha >=250") || !strings.Contains(scene.Profile, "before final union clipping") {
				t.Fatal("scene profile describes the strict guard")
			}
			art4 := b.pixels[scene.ID+"-artwork-4x.png"]
			tr := scene.Transform
			if tr.Offset[0] > 0 || tr.Offset[1] > 0 || tr.Offset[0]+float64(tr.Source.Width)*tr.Scale < float64(scene.Width)-1e-9 || tr.Offset[1]+float64(tr.Source.Height)*tr.Scale < float64(scene.Height)-1e-9 {
				t.Fatalf("artwork does not fill the scene box: %+v", tr)
			}
			want2, err := resizeLanczos(art4, scene.Width*2, scene.Height*2)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(want2.Pix, b.pixels[scene.ID+"-artwork-2x.png"].Pix) {
				t.Fatal("2x source was clipped before Lanczos reduction")
			}
			for _, density := range []int{2, 4} {
				assembled := image.NewNRGBA(image.Rect(0, 0, scene.Width*density, scene.Height*density))
				for _, tile := range scene.Tiles {
					name := fmt.Sprintf("tiles/%s-%dx.png", tile.RuntimeID, density)
					var im image.Image
					for _, artifact := range b.artifacts {
						if artifact.name == name {
							im, err = png.Decode(bytes.NewReader(artifact.data))
							if err != nil {
								t.Fatal(err)
							}
							break
						}
					}
					if im == nil {
						t.Fatalf("actual tile PNG missing: %s", name)
					}
					at := image.Pt(int(tile.Origin[0])*density, int(tile.Origin[1])*density)
					draw.Draw(assembled, im.Bounds().Add(at), im, im.Bounds().Min, draw.Over)
				}
				whole := b.pixels[fmt.Sprintf("%s-scene-%dx.png", scene.ID, density)]
				if !bytes.Equal(assembled.Pix, whole.Pix) {
					t.Fatalf("%dx actual tile PNG assembly differs", density)
				}
				if whole.NRGBAAt(0, 0).A != 0 {
					t.Fatalf("%dx final union did not clip outside artwork", density)
				}
				fence := whole.NRGBAAt(40*density, 42*density)
				if fence.R < 150 || fence.A != 255 {
					t.Fatalf("%dx original boundary fence faded: %v", density, fence)
				}
				// The road covers the outermost pixel row (column) at the union
				// edge across its whole width: no ground band at the gate.
				road := func(x, y int) {
					c := whole.NRGBAAt(x, y)
					if c.R < 70 || c.R > 90 || c.G < 72 || c.G > 92 || c.B < 73 || c.B > 93 || c.A != 255 {
						t.Fatalf("%dx road does not reach the union edge at (%d,%d): %v", density, x, y, c)
					}
				}
				if scene.Layout != "factory" {
					for x := 114 * density; x < 124*density; x++ {
						road(x, 0)
					}
				} else {
					// Only the vertex row reaches the last column.
					road(scene.Width*density-1, 80*density)
				}
			}
		})
	}
}
