package artreview

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"testing"
)

func gravelStreamInputs(t *testing.T) connectionInputs {
	t.Helper()
	in := connectedSolidInputs(t)
	bank := image.NewNRGBA(image.Rect(0, 0, 384, 320))
	draw.Draw(bank, bank.Bounds(), image.NewUniform(color.NRGBA{190, 180, 150, 255}), image.Point{}, draw.Src)
	in.materials["bank"] = bank
	return in
}

func TestConnectedGravelStreamManifestKeepsBankSeparateFromEarth(t *testing.T) {
	dir, m := connectedFixture(t)
	f := map[string]any{"id": "stream", "width": 18, "border": 2.5, "material": "water", "border_material": "bank", "variants": []any{map[string]any{"id": "stream36", "ports": []int{3, 6}}}}
	m["families"] = []any{f}
	m["materials"].(map[string]any)["bank"] = m["materials"].(map[string]any)["earth"]
	if _, err := loadConnectionInputs(writeConnectedManifest(t, dir, m)); err != nil {
		t.Fatal(err)
	}
	f["id"], f["material"] = "road", "road"
	if _, err := loadConnectionInputs(writeConnectedManifest(t, dir, m)); err == nil {
		t.Fatal("gravel stream bank accepted for another family")
	}
}

// Flat sampled water with an earth outline cannot produce the dark wet lip
// and intermittent pale foam that make this narrow river readable.
func TestConnectedGravelStreamHasWetLipFoamAndPaleBank(t *testing.T) {
	in := gravelStreamInputs(t)
	f := connectionFamily{ID: "stream", Width: 18, Border: 2.5, Material: "water", BorderMaterial: "bank"}
	r := connectionRoutes(in.manifest.Geometry, f, connectionVariant{Ports: []int{3, 6}})
	wet, foam, gravel := 0, 0, 0
	for y := 20.; y < 60; y += .25 {
		for x := 52.; x < 62; x += .25 {
			c := connectionFeature(in, f, r, connectionPoint{x, y}, connectionPoint{x, y}, .25)
			if c.A < 240 {
				continue
			}
			if c.R < 110 && c.G < 110 && c.B < 110 {
				wet++
			}
			if c.R > 100 && c.G > 140 && c.B > 160 {
				foam++
			}
			if c.R > 165 && c.G > 150 && c.B < 160 {
				gravel++
			}
		}
	}
	if wet < 20 || foam < 5 || gravel < 50 {
		t.Fatalf("river lacks material cues: wet%d foam%d gravel%d", wet, foam, gravel)
	}
}

func TestConnectedMarkedRoadVergeVariesInsideFixedGuard(t *testing.T) {
	in := gravelStreamInputs(t)
	f := connectionFamily{ID: "road", Width: 22, Border: 2, Material: "road", BorderMaterial: "earth", Markings: "dashed"}
	r := connectionRoutes(in.manifest.Geometry, f, connectionVariant{Ports: []int{3, 6}})
	visible, absent := 0, 0
	for y := 20.; y < 60; y += .25 {
		p := connectionPoint{61.1, y}
		c := connectionFeature(in, f, r, p, p, .25)
		if c.A > 20 {
			visible++
		}
		if c.A == 0 {
			absent++
		}
	}
	if visible < 5 || absent < 5 {
		t.Fatalf("road verge remains uniformly straight: visible%d absent%d", visible, absent)
	}
	for _, y := range []float64{0, 2, 4, 6, 8} {
		p := connectionPoint{61.5, y}
		if c := connectionFeature(in, f, r, p, p, .25); c.A != 0 {
			t.Fatalf("road verge changed fixed guard at%v: %v", p, c)
		}
	}
}

// Different bends on either side of a port must share the complete painted
// cross-section, including wet lips, foam, gravel and road shoulder alpha.
func TestConnectedRefdesignEdgesMatchEveryReciprocalPort(t *testing.T) {
	in := gravelStreamInputs(t)
	g := in.manifest.Geometry
	for _, f := range []connectionFamily{{ID: "stream", Width: 18, Border: 2.5, Material: "water", BorderMaterial: "bank"}, {ID: "stream", Width: 18, Border: 3.5, Material: "water", BorderMaterial: "bank"}, {ID: "stream", Width: 18, Border: 3.5, Material: "water", BorderMaterial: "bank", BankProfile: authoredBankProfile}, {ID: "road", Width: 22, Border: 2, Material: "road", BorderMaterial: "earth", Markings: "dashed"}} {
		for port := 1; port <= 6; port++ {
			opp := (port+2)%6 + 1
			a := connectionRoutes(g, f, connectionVariant{Ports: []int{port, opp}})
			b := connectionRoutes(g, f, connectionVariant{Ports: []int{opp, opp%6 + 1}})
			n := g.Neighbors[port-1].mul(1 / g.Neighbors[port-1].length())
			side := connectionPoint{-n[1], n[0]}
			for _, scale := range []int{2, 4} {
				for along := -3.; along <= 3; along += .5 {
					for across := -15.; across <= 15; across += .25 {
						p := g.Ports[port-1].add(n.mul(along)).add(side.mul(across))
						q := p.sub(g.Neighbors[port-1])
						x := connectionFeature(in, f, a, p, p, 1/float64(scale))
						y := connectionFeature(in, f, b, q, q, 1/float64(scale))
						if x.A != y.A || math.Abs(float64(x.R)-float64(y.R)) > 1 || math.Abs(float64(x.G)-float64(y.G)) > 1 || math.Abs(float64(x.B)-float64(y.B)) > 1 {
							t.Fatalf("%s port%d %dx mismatch at%v: %v != %v", f.ID, port, scale, p, x, y)
						}
					}
				}
			}
		}
	}
}

func TestWideGravelStreamBreaksInteriorBanksAndKeepsCanonicalGuard(t *testing.T) {
	g := sharedConnectionGeometry()
	f := connectionFamily{ID: "stream", Width: 18, Border: 3.5, Material: "water", BorderMaterial: "bank"}
	r := connectionRoutes(g, f, connectionVariant{Ports: []int{3, 6}})
	broken := 0
	for y := 24.; y <= 56; y++ {
		for x := 32.; x <= 64; x++ {
			p := connectionPoint{x, y}
			core, outer, _ := connectionGravelStreamProfile(g, f, r, p, p)
			border := core - outer
			if math.Abs(border-3.5) > .8 {
				broken++
			}
			if border < 0 || border > 4.6+1e-9 {
				t.Fatalf("wide bank out of bounds: %g", border)
			}
		}
	}
	if broken < 20 {
		t.Fatalf("wide bank remains a smooth outline: %d distinctly broken samples", broken)
	}
	for _, y := range []float64{0, 2, 4, 6, 8} {
		p := connectionPoint{48, y}
		core, outer, _ := connectionGravelStreamProfile(g, f, r, p, p)
		if math.Abs(core-outer-3.5) > 1e-9 {
			t.Fatalf("canonical port bank changed: %g", core-outer)
		}
	}
}
