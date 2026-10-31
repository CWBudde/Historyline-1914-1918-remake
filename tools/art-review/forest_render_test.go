package artreview

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"reflect"
	"testing"
)

func TestForestPinnedSeedReproducesSpeciesCenters(t *testing.T) {
	// Independent bounded replay pins every center, not merely point count.
	var points []connectionPoint
	state := uint64(0x19141918)
	attempts := 0
	deltas := [][2]int{{0, 0}, {71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}}
	for ; attempts < 200000 && len(points) < 36; attempts++ {
		state = state*6364136223846793005 + 1442695040888963407
		x := int((state >> 32) % 384)
		state = state*6364136223846793005 + 1442695040888963407
		y := int((state >> 32) % 320)
		owned := true
		for i, a := range groundTestPolygon {
			b := groundTestPolygon[(i+1)%6]
			dx, dy := int(b[0]-a[0]), int(b[1]-a[1])
			cross := dx*(y-4*int(a[1])) - dy*(x-4*int(a[0]))
			tieOwned := dy < 0 || dy == 0 && dx > 0
			if cross < 0 || cross == 0 && !tieOwned {
				owned = false
			}
		}
		if !owned {
			continue
		}
		for _, point := range points {
			for _, d := range deltas {
				if math.Hypot(float64(x)/4-point[0]-float64(d[0]), float64(y)/4-point[1]-float64(d[1])) < 10 {
					owned = false
				}
			}
		}
		if owned {
			points = append(points, connectionPoint{float64(x) / 4, float64(y) / 4})
		}
	}
	if attempts != 1934 || !reflect.DeepEqual(points, forestTemplate()) {
		t.Fatalf("seed no longer reproduces template: attempts %d", attempts)
	}
}

func TestForestSharedLocalFieldsAndCrossingCrowns(t *testing.T) {
	for _, profile := range []string{"", testForestCanopyProfile} {
		t.Run(profile, func(t *testing.T) { forestTestSharedLocalFieldsAndCrossingCrowns(t, profile) })
	}
}

func forestTestSharedLocalFieldsAndCrossingCrowns(t *testing.T, profile string) {
	t.Helper()
	path := forestTestFixture(t)
	in, err := loadForestInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	grounds, atlases := map[string]*image.NRGBA{}, map[string]*image.NRGBA{}
	for _, season := range []string{"summer", "winter"} {
		grounds[season] = in.Grounds[season].Image
		atlases[season] = in.Atlases[season].Image
	}
	renderer, err := newForestRenderer(grounds, atlases, profile)
	if err != nil {
		t.Fatal(err)
	}
	for _, season := range []string{"summer", "winter"} {
		for _, scenario := range []string{"homogeneous", "frontier", "clearing"} {
			scene, err := forestSceneFor(season, scenario, 0)
			if err != nil {
				t.Fatal(err)
			}
			raw := map[connectionPoint]*image.NRGBA{}
			for _, cell := range scene.Cells {
				context, err := newForestContext(cell.RuntimeID, cell.NeighborIDs[:])
				if err != nil {
					t.Fatal(err)
				}
				raw[cell.Origin] = renderer.raw(context)
			}
			center := raw[connectionPoint{}]
			crossings := 0
			for _, cell := range scene.Cells[1:] {
				delta := cell.Origin
				other := raw[delta]
				matched := 0
				for y := range 352 {
					for x := range 416 {
						p := connectionPoint{(float64(x)+.5)/4 - 4, (float64(y)+.5)/4 - 4}
						if groundTestDistance(p[0], p[1]) > 4 || groundTestDistance(p[0]-delta[0], p[1]-delta[1]) > 4 {
							continue
						}
						qx, qy := x-int(delta[0]*4), y-int(delta[1]*4)
						if !image.Pt(qx, qy).In(other.Bounds()) {
							continue
						}
						matched++
						if center.NRGBAAt(x, y) != other.NRGBAAt(qx, qy) {
							t.Fatalf("local context field disagreement %s %s delta %v p %v", season, scenario, delta, p)
						}
						if scenario == "homogeneous" && center.NRGBAAt(x, y) != renderer.baseRaw[season].NRGBAAt(x, y) {
							crossings++
						}
					}
				}
				if matched < 100 {
					t.Fatalf("shared field not exercised %v", delta)
				}
			}
			if scenario == "homogeneous" && crossings < 100 {
				t.Fatalf("homogeneous forest lost crowns along shared cuts: %d", crossings)
			}
			// Every final quiet-band pixel equals the unchanged textured base, including
			// the 2x Lanczos result rather than merely raw 4x samples.
			if scenario == "frontier" {
				centerContext, _ := newForestContext(scene.Cells[0].RuntimeID, scene.Cells[0].NeighborIDs[:])
				tiles, err := renderer.tiles(centerContext)
				if err != nil {
					t.Fatal(err)
				}
				base := "grass"
				if season == "winter" {
					base = "grass_w"
				}
				baseContext, _ := newForestContext(base, []string{base, base, base, base, base, base})
				bases, err := renderer.tiles(baseContext)
				if err != nil {
					t.Fatal(err)
				}
				quiet := 0
				for scale, im := range tiles {
					for y := range 80 * scale {
						for x := range 96 * scale {
							if !groundTestOwned(x, y, scale) {
								continue
							}
							px, py := (float64(x)+.5)/float64(scale), (float64(y)+.5)/float64(scale)
							isQuiet := false
							for _, site := range scene.Sites {
								if site.RuntimeID == base && groundTestDistance(px-site.Origin[0], py-site.Origin[1]) <= 4 {
									isQuiet = true
								}
							}
							if isQuiet {
								quiet++
								if im.NRGBAAt(x, y) != bases[scale].NRGBAAt(x, y) {
									t.Fatalf("textured final base band leaked %s %dx %d,%d", season, scale, x, y)
								}
							}
						}
					}
				}
				if quiet < 100 {
					t.Fatal("quiet band not exercised")
				}
			}
		}
	}
	// Whole transparent panels remain distinct in row-major species order.
	for i, crown := range renderer.crowns["summer"] {
		pixels := 96
		if profile != "" {
			pixels = 106
		}
		if crown.Bounds() != image.Rect(0, 0, pixels, pixels) || crown.RGBAAt(0, 0) != (color.RGBA{}) {
			t.Fatal("whole panel transparent margin lost")
		}
		for j, other := range renderer.crowns["summer"] {
			if i != j && bytes.Equal(crown.Pix, other.Pix) {
				t.Fatal("species mixed")
			}
		}
	}
}
