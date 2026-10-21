package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"math"
	"slices"
	"testing"
)

func TestShoreMouthExportsEveryLandEdge(t *testing.T) {
	path, _ := shoreTestFixture(t)
	in, err := loadShoreInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	b, err := buildShoreBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Mouths []struct {
			ID      string       `json:"id"`
			Shore   string       `json:"shore"`
			Port    int          `json:"port"`
			Outputs []packOutput `json:"outputs"`
		} `json:"mouths"`
		Proofs []packOutput `json:"mouth_proofs"`
	}
	for _, a := range b.artifacts {
		if a.name == "report.json" {
			if err := json.Unmarshal(a.data, &report); err != nil {
				t.Fatal(err)
			}
		}
	}
	if len(report.Mouths) != 24 || len(report.Proofs) != 2 {
		t.Fatalf("mouth exports=%d proofs=%d; want24/2", len(report.Mouths), len(report.Proofs))
	}
	seen := map[string]bool{}
	for _, m := range report.Mouths {
		if m.ID != fmt.Sprintf("%s-outlet%d", m.Shore, m.Port) || seen[m.ID] || len(m.Outputs) != 2 {
			t.Fatalf("invalid/duplicate mouth %+v", m)
		}
		seen[m.ID] = true
		profile := shoreTestProfiles()[m.Shore]
		if !slices.Contains(profile, m.Port%6) || !slices.Contains(profile, (m.Port+1)%6) {
			t.Fatalf("mouth changes a mixed/water edge %+v", m)
		}
	}
	// Independently reassemble the three labelled proof panels from the
	// encoded/decoded output tiles. Presentation text stays outside the art.
	for _, scale := range []int{2, 4} {
		proof := b.pixels[fmt.Sprintf("shore-mouths-%dx.png", scale)]
		if proof == nil {
			t.Fatal("missing native mouth proof")
		}
		want := image.NewNRGBA(proof.Bounds())
		for i, tc := range []struct {
			id    string
			port  int
			delta image.Point
		}{{"lakeside12-outlet1", 1, image.Pt(71, -40)}, {"lakeside23-outlet4", 4, image.Pt(-71, 40)}, {"lakeside3456-outlet3", 3, image.Pt(0, 80)}} {
			at := image.Pt((12+254*i-min(0, tc.delta.X))*scale, (24-min(0, tc.delta.Y))*scale)
			shore := b.pixels[fmt.Sprintf("tiles/%s-%dx.png", tc.id, scale)]
			draw.Draw(want, shore.Bounds().Add(at), shore, image.Point{}, draw.Over)
			at = at.Add(tc.delta.Mul(scale))
			grass := b.pixels[fmt.Sprintf("tiles/grass-%dx.png", scale)]
			draw.Draw(want, grass.Bounds().Add(at), grass, image.Point{}, draw.Over)
			p := (tc.port-1)%3 + 1
			stream := b.pixels[fmt.Sprintf("mouth-proof/stream%d%d-%dx.png", p, p+3, scale)]
			draw.Draw(want, stream.Bounds().Add(at), stream, image.Point{}, draw.Over)
		}
		for y := range want.Bounds().Dy() {
			for x := range want.Bounds().Dx() {
				if c := want.NRGBAAt(x, y); c.A == 255 && c != proof.NRGBAAt(x, y) {
					t.Fatalf("%dx proof altered saved tile pixel%d,%d", scale, x, y)
				}
			}
		}
	}
}

// A mouth must continue the same exported stream on grass at its opening;
// the remaining shore edges must retain their original pixel guards.
func TestShoreMouthSharedStreamGuardsAndConnectedWater(t *testing.T) {
	path, m := shoreTestFixture(t)
	m["bank_profile"] = "authored-rocks-v1"
	shoreTestWrite(t, path, m)
	in, err := loadShoreInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	profiles := shoreTestProfiles()
	for _, tc := range []struct {
		id   string
		port int
	}{{"lakeside12", 1}, {"lakeside23", 2}, {"lakeside126", 6}, {"lakeside3456", 3}, {"lakeside23", 4}, {"lakeside56", 5}} {
		t.Run(fmt.Sprintf("%s-outlet%d", tc.id, tc.port), func(t *testing.T) {
			v := shoreVariant{tc.id, profiles[tc.id]}
			mouth, err := exportShoreMouth(in, v, tc.port)
			if err != nil {
				t.Fatal(err)
			}
			original, err := exportShoreTiles(in, v)
			if err != nil {
				t.Fatal(err)
			}
			c, f := shoreMouthStream(in)
			straight := connectionVariant{Ports: []int{tc.port, (tc.port+2)%6 + 1}}
			stream, err := exportConnectionTiles(c, &f, &straight)
			if err != nil {
				t.Fatal(err)
			}
			grass, err := exportGrassBaseTiles(in.Materials["grass"].Image)
			if err != nil {
				t.Fatal(err)
			}
			for _, scale := range []int{2, 4} {
				im := mouth[scale]
				streamGround := image.NewNRGBA(im.Bounds())
				draw.Draw(streamGround, streamGround.Bounds(), grass[scale], image.Point{}, draw.Src)
				draw.Draw(streamGround, streamGround.Bounds(), stream[scale], image.Point{}, draw.Over)
				changed, joined := 0, 0
				for y := range im.Bounds().Dy() {
					for x := range im.Bounds().Dx() {
						if im.NRGBAAt(x, y).A == 0 {
							continue
						}
						p := connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}
						other := math.Inf(1)
						for e := range 6 {
							if e != tc.port%6 {
								other = min(other, shoreMouthEdgeDistance(p, e))
							}
						}
						got := im.NRGBAAt(x, y)
						if got != original[scale].NRGBAAt(x, y) {
							changed++
						}
						if other <= 8 {
							if got != original[scale].NRGBAAt(x, y) {
								t.Fatalf("%dx other-edge guard changed at%d,%d", scale, x, y)
							}
						} else if shoreMouthEdgeDistance(p, tc.port%6) <= 8 {
							if got != streamGround.NRGBAAt(x, y) {
								t.Fatalf("%dx stream guard differs at%d,%d", scale, x, y)
							}
							joined++
						}
					}
				}
				if changed < 20*scale*scale || joined < 20*scale*scale {
					t.Fatalf("%dx missing channel/guard coverage: changed%d joined%d", scale, changed, joined)
				}
				// Flood from the mouth into an unchanged original-water pixel.
				port := sharedConnectionGeometry().Ports[tc.port-1]
				n := sharedConnectionGeometry().Neighbors[tc.port-1]
				n = n.mul(-2 / n.length())
				start := image.Pt(int((port[0]+n[0])*float64(scale)), int((port[1]+n[1])*float64(scale)))
				water := func(at image.Point) bool {
					col := im.NRGBAAt(at.X, at.Y)
					return col.A == 255 && int(col.B) > int(col.G)+25 && int(col.B) > int(col.R)+40
				}
				if !water(start) {
					t.Fatalf("%dx mouth is blocked at%v by%v", scale, start, im.NRGBAAt(start.X, start.Y))
				}
				queue := []image.Point{start}
				seen := map[image.Point]bool{start: true}
				connected := false
				for len(queue) > 0 {
					p := queue[0]
					queue = queue[1:]
					old := original[scale].NRGBAAt(p.X, p.Y)
					if int(old.B) > int(old.G)+25 && int(old.B) > int(old.R)+40 {
						connected = true
						break
					}
					for _, d := range []image.Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
						q := p.Add(d)
						if q.In(im.Bounds()) && !seen[q] && water(q) {
							seen[q] = true
							queue = append(queue, q)
						}
					}
				}
				if !connected {
					t.Fatalf("%dx stream does not reach original lake water", scale)
				}
			}
		})
	}
}
