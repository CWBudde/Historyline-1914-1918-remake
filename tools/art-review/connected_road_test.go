package artreview

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"path/filepath"
	"testing"
)

func TestConnectedRoadMarkingValidation(t *testing.T) {
	for _, tc := range []struct {
		json string
		bad  bool
	}{
		{`{"families":[{"id":"road","markings":"dashed"}]}`, false},
		{`{"families":[{"id":"road"}]}`, false},
		{`{"families":[{"id":"path","markings":"dashed"}]}`, true},
		{`{"families":[{"id":"road","Markings":"dashed"}]}`, true},
		{`{"families":[{"id":"road","markings":null}]}`, true},
		{`{"families":[{"id":"road","markings":""}]}`, true},
		{`{"families":[{"id":"road","markings":"solid"}]}`, true},
		{`{"families":[{"id":"road","markings":"ruts"}]}`, false},
		{`{"families":[{"id":"road","markings":"ruts","verge":"snow"}]}`, false},
		{`{"families":[{"id":"road","verge":"snow"}]}`, true},
		{`{"families":[{"id":"road","markings":"ruts","verge":"mud"}]}`, true},
		{`{"families":[{"id":"road","markings":"ruts","Verge":"snow"}]}`, true},
		// Winter paths carry ruts and snow verges; winter rails snow verges
		// without markings.
		{`{"families":[{"id":"path","markings":"ruts","verge":"snow"}]}`, false},
		{`{"families":[{"id":"path","verge":"snow"}]}`, true},
		{`{"families":[{"id":"rail","verge":"snow"}]}`, false},
		{`{"families":[{"id":"rail","markings":"ruts","verge":"snow"}]}`, true},
		{`{"families":[{"id":"rail","verge":"frost"}]}`, true},
		{`{"families":[{"id":"stream","verge":"snow"}]}`, true},
		{`{"families":[{"id":"trench","verge":"snow"}]}`, true},
	} {
		if err := validateConnectionRoadFields([]byte(tc.json)); (err != nil) != tc.bad {
			t.Fatalf("%s: %v", tc.json, err)
		}
	}
}

func TestConnectedRoadDashesJoinAcrossEveryPort(t *testing.T) {
	g := sharedConnectionGeometry()
	for port := 1; port <= 6; port++ {
		opposite := (port+2)%6 + 1
		n := g.Neighbors[port-1].mul(1 / g.Neighbors[port-1].length())
		side := connectionPoint{-n[1], n[0]}
		a := connectionRoadPaintFor(g, connectionVariant{Ports: []int{port, port%6 + 1}})
		for _, ports := range [][]int{{opposite, (opposite+2)%6 + 1}, {opposite, opposite%6 + 1}, {opposite, opposite%6 + 1, (opposite+1)%6 + 1}} {
			b := connectionRoadPaintFor(g, connectionVariant{Ports: ports})
			for _, density := range []int{2, 4} {
				for along := -7.5; along <= 7.5; along += .5 {
					for cross := -4.; cross <= 4; cross += .25 {
						p := g.Ports[port-1].add(n.mul(along)).add(side.mul(cross))
						x := connectionRoadPaintAt(a, p, 1/float64(density))
						y := connectionRoadPaintAt(b, p.sub(g.Neighbors[port-1]), 1/float64(density))
						if math.Abs(x-y) > 1e-6 {
							t.Fatalf("port%d %dx along%v cross%v: %v != %v", port, density, along, cross, x, y)
						}
					}
				}
			}
		}
	}
}

func TestConnectedMarkedRoadActualPNG(t *testing.T) {
	dir, m := connectedFixture(t)
	f := m["families"].([]any)[0].(map[string]any)
	f["markings"] = "dashed"
	f["width"] = 22
	f["border"] = 1
	f["variants"] = []any{map[string]any{"id": "road36", "ports": []int{3, 6}}}
	surface := image.NewNRGBA(image.Rect(0, 0, 384, 320))
	draw.Draw(surface, surface.Bounds(), image.NewUniform(color.NRGBA{119, 111, 120, 255}), image.Point{}, draw.Src)
	ref := unitTestPNG(t, filepath.Join(dir, "paved.png"), surface)
	m["materials"].(map[string]any)["road"] = ref
	in, err := loadConnectionInputs(writeConnectedManifest(t, dir, m))
	if err != nil {
		t.Fatal(err)
	}
	tiles, err := exportConnectionTiles(in, &in.manifest.Families[0], &in.manifest.Families[0].Variants[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, density := range []int{2, 4} {
		im := tiles[density]
		light, dark := 0, 0
		for y := 0; y < 80*density; y++ {
			p := im.NRGBAAt(48*density, y)
			if p.R > 220 {
				light++
			}
			if p.R < 150 {
				dark++
			}
		}
		if light < 20*density || dark < 10*density {
			t.Fatalf("%dx missing dash/gap pattern: %d/%d", density, light, dark)
		}
		if im.NRGBAAt(48*density, 0).R < 220 || im.NRGBAAt(48*density, 80*density-1).R < 220 {
			t.Fatal("edge dash does not continue across tile join")
		}
		if im.NRGBAAt(35*density, 40*density).A >= 128 || im.NRGBAAt(60*density, 40*density).A >= 128 {
			t.Fatal("road exceeds reference 24px footprint")
		}
		if im.NRGBAAt(37*density, 40*density).A < 250 {
			t.Fatal("reference-width road is too narrow")
		}
	}
}

// Winter roads: no painted lines, two darker slush ruts either side of the
// centre line, and light snow shoulders.
func TestConnectedWinterRoadActualPNG(t *testing.T) {
	dir, m := connectedFixture(t)
	f := m["families"].([]any)[0].(map[string]any)
	f["markings"], f["verge"] = "ruts", "snow"
	f["width"], f["border"] = 22, 2
	f["variants"] = []any{map[string]any{"id": "road36_w", "ports": []int{3, 6}}}
	surface := image.NewNRGBA(image.Rect(0, 0, 384, 320))
	draw.Draw(surface, surface.Bounds(), image.NewUniform(color.NRGBA{140, 138, 140, 255}), image.Point{}, draw.Src)
	m["materials"].(map[string]any)["road"] = unitTestPNG(t, filepath.Join(dir, "snow-road.png"), surface)
	in, err := loadConnectionInputs(writeConnectedManifest(t, dir, m))
	if err != nil {
		t.Fatal(err)
	}
	tiles, err := exportConnectionTiles(in, &in.manifest.Families[0], &in.manifest.Families[0].Variants[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, density := range []int{2, 4} {
		im := tiles[density]
		y := 40 * density
		centre := im.NRGBAAt(48*density, y)
		for x := 0; x < 96*density; x++ {
			if p := im.NRGBAAt(x, y); p.A > 0 && p.R > 230 {
				t.Fatalf("%dx painted line at x=%d: %v", density, x, p)
			}
		}
		for _, x := range []float64{48 - 4.5, 48 + 4.5} {
			rut := im.NRGBAAt(int(x*float64(density)), y)
			if int(rut.R)+20 > int(centre.R) {
				t.Fatalf("%dx rut at %v not darker than the centre: %v vs %v", density, x, rut, centre)
			}
		}
		for _, x := range []float64{48 - 12, 48 + 12} {
			shoulder := im.NRGBAAt(int(x*float64(density)), y)
			if shoulder.A < 128 || shoulder.R < 175 {
				t.Fatalf("%dx shoulder at %v is not snow: %v", density, x, shoulder)
			}
		}
	}
}
