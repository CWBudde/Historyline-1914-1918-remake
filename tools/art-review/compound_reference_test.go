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
	raw := compoundReferenceArtwork(l)
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

// A road on the north/east edge, a fence one world pixel from the indentation,
// and an outside pixel distinguish full reference framing from a relaxed fade.
func compoundReferenceArtwork(l compoundLayout) *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, l.Width*4, l.Height*4))
	road := image.Rect(111*4, 0, 127*4, 100*4)
	if l.Layout == "factory" {
		road = image.Rect(100*4, 74*4, l.Width*4, 86*4)
	}
	draw.Draw(im, road, image.NewUniform(color.NRGBA{80, 82, 83, 255}), image.Point{}, draw.Src)
	draw.Draw(im, image.Rect(27*4, 41*4, 211*4, 43*4), image.NewUniform(color.NRGBA{190, 40, 30, 255}), image.Point{}, draw.Src)
	im.SetNRGBA(0, 0, color.NRGBA{240, 10, 240, 255})
	return im
}

func TestCompoundReferenceFramingActualPNGAssembly(t *testing.T) {
	path, m := compoundTestFixture(t)
	m.Grounds["winter"] = m.Grounds["summer"]
	m.Scenes = nil
	for _, id := range []string{"hq-summer", "hq-winter", "depot-summer", "depot-winter", "factory-winter"} {
		l, err := compoundLayoutFor(id)
		if err != nil {
			t.Fatal(err)
		}
		var encoded bytes.Buffer
		if err = png.Encode(&encoded, compoundReferenceArtwork(l)); err != nil {
			t.Fatal(err)
		}
		name := id + ".png"
		if err = os.WriteFile(filepath.Join(filepath.Dir(path), name), encoded.Bytes(), 0600); err != nil {
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
			if !bytes.Equal(art4.Pix, in.Artwork[scene.ID].Image.Pix) {
				t.Fatal("1:1 affine reference artwork was changed")
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
				fence := whole.NRGBAAt(40*density, 41*density)
				if fence.R < 150 || fence.A != 255 {
					t.Fatalf("%dx original boundary fence faded: %v", density, fence)
				}
				if scene.Layout != "factory" {
					road := whole.NRGBAAt(119*density, 0)
					if road.R < 70 || road.B < 60 || road.A != 255 {
						t.Fatalf("%dx north road does not reach union edge: %v", density, road)
					}
				} else {
					road := whole.NRGBAAt(scene.Width*density-1, 80*density)
					if road.R < 70 || road.B < 60 || road.A != 255 {
						t.Fatalf("%dx east road does not reach union edge: %v", density, road)
					}
				}
			}
		})
	}
}
