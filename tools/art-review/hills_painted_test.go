package artreview

import (
	"encoding/json"
	"image"
	"image/color"
	"path/filepath"
	"testing"
)

// Rejecting an optional painted source or overlooking its alpha would break authored hills inputs.
func TestHillsPaintedInput(t *testing.T) {
	path := hillsTestFixture(t)
	var m hillsManifest
	if err := json.Unmarshal(groundTestRead(t, path), &m); err != nil {
		t.Fatal(err)
	}
	paint := image.NewNRGBA(image.Rect(0, 0, 384, 320))
	for y := range 320 {
		for x := range 384 {
			paint.SetNRGBA(x, y, color.NRGBA{180, 160, 80, 255})
		}
	}
	for _, alpha := range []uint8{255, 254} {
		paint.SetNRGBA(0, 0, color.NRGBA{180, 160, 80, alpha})
		m.Materials["scrub"] = forestTestWritePNG(t, filepath.Join(filepath.Dir(path), "scrub.png"), paint)
		b, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		groundTestWrite(t, path, b)
		in, err := loadHillsInputs(path)
		if alpha != 255 {
			if err == nil {
				t.Fatal("accepted transparent painted ground")
			}
			continue
		}
		if err != nil {
			t.Fatalf("opaque painted meadow rejected: %v", err)
		}
		if len(in.Sources) != 3 {
			t.Fatal("painted source not retained")
		}
	}
}

// Reapplying height lighting, an isolated bump, or painting through a grass-facing edge breaks this.
func TestHillsPaintedBundlePreservesMaterialAndGrassBorders(t *testing.T) {
	in, err := loadHillsInputs(hillsTestFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	paint := image.NewNRGBA(image.Rect(0, 0, 384, 320))
	for y := range 320 {
		for x := range 384 {
			paint.SetNRGBA(x, y, color.NRGBA{180, 160, 80, 255})
		}
	}
	in.Sources["scrub"] = compoundSource{Image: paint}
	b, err := buildHillsBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	base, err := exportGrassBaseTiles(in.Sources["grass"].Image)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range b.report.Profiles {
		if p.RuntimeID != "hills" {
			continue
		}
		if p.IsolatedCue.Enabled {
			t.Fatal("painted material received isolated height cue")
		}
		for _, out := range p.Outputs {
			im := b.pixels[out.Path]
			scale := im.Bounds().Dx() / 96
			if im.NRGBAAt(48*scale, 40*scale) != (color.NRGBA{180, 160, 80, 255}) {
				t.Fatal("painted centre shaded or omitted")
			}
			for y := range im.Bounds().Dy() {
				for x := range im.Bounds().Dx() {
					if !groundTestOwned(x, y, scale) {
						continue
					}
					for n, d := range groundTestDeltas {
						if p.NeighborIDs[n] == "grass" && groundTestDistance((float64(x)+.5)/float64(scale)-float64(d[0]), (float64(y)+.5)/float64(scale)-float64(d[1])) <= 4 && im.NRGBAAt(x, y) != base[scale].NRGBAAt(x, y) {
							t.Fatal("painted material leaked into grass border")
						}
					}
				}
			}
		}
	}
}
