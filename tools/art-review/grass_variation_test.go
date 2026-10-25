package main

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"testing"
)

// A changed edge guard would show seams against any existing family. A lost
// source phase would leave repeated or tint-only interiors across the map.
func TestGrassVariationsKeepBoundariesAndDistinctInteriors(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 1536, 1024))
	for y := range 1024 {
		for x := range 1536 {
			src.SetNRGBA(x, y, color.NRGBA{uint8(x / 6), uint8(y / 4), uint8((x/11 + y/7) % 256), 255})
		}
	}
	base, err := exportGrassBaseTiles(src)
	if err != nil {
		t.Fatal(err)
	}
	signatures := map[color.NRGBA]bool{}
	for _, variant := range []int{0, 1, 5, 10, 15} {
		tiles, err := renderGrassVariation(src, variant)
		if err != nil {
			t.Fatal(err)
		}
		for _, scale := range []int{2, 4} {
			im := tiles[scale]
			for y := range 80 * scale {
				for x := range 96 * scale {
					p := connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}
					got, want := im.NRGBAAt(x, y), base[scale].NRGBAAt(x, y)
					if !connectionInside(sharedConnectionGeometry(), p) || mixedEdgeDistance(sharedConnectionGeometry(), p) < 6 {
						if got != want {
							t.Fatalf("variant %d/%dx changed canonical boundary at %v: %v != %v", variant, scale, p, got, want)
						}
					}
				}
			}
		}
		signatures[tiles[4].NRGBAAt(48*4, 40*4)] = true
	}
	if len(signatures) != 5 {
		t.Fatalf("distinct phase samples collapsed: %d", len(signatures))
	}
	for _, bad := range []int{-1, 16} {
		if _, err := renderGrassVariation(src, bad); err == nil {
			t.Errorf("accepted variant %d", bad)
		}
	}
}

func TestGrassSamplerIsPeriodicWithoutReflectionSymmetry(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 384, 320))
	for y := range 320 {
		for x := range 384 {
			src.SetNRGBA(x, y, color.NRGBA{uint8((x*x/37 + y*3) % 256), uint8((x + y*y/31) % 256), uint8((x/7 + y/3) % 256), 255})
		}
	}
	asymmetric := false
	for _, p := range []connectionPoint{{1.125, 5.375}, {34.25, 27.75}, {-25.125, -12.875}, {0, 0}} {
		want := sampleGrassMaterial(src, p)
		for _, shift := range sharedConnectionGeometry().Neighbors {
			if got := sampleGrassMaterial(src, p.add(shift)); got != want {
				t.Fatalf("lattice discontinuity at %v + %v: %v != %v", p, shift, got, want)
			}
		}
		reflected := sampleGrassMaterial(src, connectionPoint{-p[0], p[1]})
		asymmetric = asymmetric || reflected != want
	}
	if !asymmetric {
		t.Fatal("grass sampler still imposes ornamental reflection symmetry")
	}
	for _, p := range []connectionPoint{{0, 12}, {142, 31}, {18, 0}, {51, 80}} {
		a := sampleGrassMaterial(src, p.add(connectionPoint{-1e-7, -1e-7}))
		b := sampleGrassMaterial(src, p.add(connectionPoint{1e-7, 1e-7}))
		if math.Abs(float64(a.R)-float64(b.R)) > 1 || math.Abs(float64(a.G)-float64(b.G)) > 1 || math.Abs(float64(a.B)-float64(b.B)) > 1 {
			t.Fatalf("grass wrap discontinuity at %v: %v != %v", p, a, b)
		}
	}
}

func TestSummerGroundUsesGrassSamplerAndWinterKeepsClassicSampler(t *testing.T) {
	dir, manifest := connectedFixture(t)
	in, err := loadConnectionInputs(writeConnectedManifest(t, dir, manifest))
	if err != nil {
		t.Fatal(err)
	}
	src := in.materials["grass"]
	canonical, err := exportGrassBaseTiles(src)
	if err != nil {
		t.Fatal(err)
	}
	connected, err := exportConnectionTiles(in, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(canonical[4].Pix, connected[4].Pix) {
		t.Fatal("loaded connected export does not use canonical summer grass")
	}
	materials := map[string]*image.NRGBA{}
	for _, id := range groundMaterialIDs {
		materials[id] = src
	}
	p := connectionPoint{34.25, 27.75}
	for _, season := range []struct {
		id   string
		want color.NRGBA
	}{{"grass", sampleGrassMaterial(src, p)}, {"grass_w", sampleConnectionMaterial(src, p)}} {
		c, err := newGroundContext(season.id, []string{season.id, season.id, season.id, season.id, season.id, season.id})
		if err != nil {
			t.Fatal(err)
		}
		got := groundRounded(groundContextRGB(c, p, materials))
		if got != season.want {
			t.Errorf("%s ground sampler: %v != %v", season.id, got, season.want)
		}
	}
}

func TestGrassDetailMaterialChangesInteriorAndKeepsCanonicalBoundary(t *testing.T) {
	base, detail := image.NewNRGBA(image.Rect(0, 0, 768, 480)), image.NewNRGBA(image.Rect(0, 0, 768, 480))
	for y := range 480 {
		for x := range 768 {
			base.SetNRGBA(x, y, color.NRGBA{70, 90, 35, 255})
			detail.SetNRGBA(x, y, color.NRGBA{110, 120, 55, 255})
		}
	}
	canonical, err := exportGrassBaseTiles(base)
	if err != nil {
		t.Fatal(err)
	}
	tiles, err := renderGrassDetailVariation(base, detail, 5)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range []int{2, 4} {
		for y := range 80 * d {
			for x := range 96 * d {
				p := connectionPoint{(float64(x) + .5) / float64(d), (float64(y) + .5) / float64(d)}
				if mixedEdgeDistance(sharedConnectionGeometry(), p) < 6 && tiles[d].NRGBAAt(x, y) != canonical[d].NRGBAAt(x, y) {
					t.Fatalf("authored detail changed canonical guard at %v/%dx", p, d)
				}
			}
		}
		if got := tiles[d].NRGBAAt(48*d, 40*d); got != (color.NRGBA{96, 110, 48, 255}) {
			t.Fatalf("authored detail missing at center/%dx: %v", d, got)
		}
	}
}
