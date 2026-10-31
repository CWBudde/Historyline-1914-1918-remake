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
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var railTestColors = map[string]color.NRGBA{"rail": {135, 125, 110, 255}, "timber": {105, 65, 30, 255}, "steel": {205, 215, 225, 255}, "earth": {90, 60, 35, 255}, "grass": {40, 80, 30, 255}}

func railTestFixture(t *testing.T) (string, map[string]any) {
	t.Helper()
	dir, m := connectedFixture(t)
	var variants []any
	for _, mask := range []string{"13", "14", "15", "236", "24", "25", "26", "346", "35", "356", "36", "46"} {
		var ports []int
		for _, p := range mask {
			ports = append(ports, int(p-'0'))
		}
		variants = append(variants, map[string]any{"id": "rail" + mask, "ports": ports})
	}
	m["families"] = []any{map[string]any{"id": "rail", "width": 11, "border": 1, "material": "rail", "border_material": "earth", "variants": variants}}
	for id, c := range railTestColors {
		im := image.NewNRGBA(image.Rect(0, 0, 384, 320))
		draw.Draw(im, im.Bounds(), image.NewUniform(c), image.Point{}, draw.Src)
		var b bytes.Buffer
		if err := png.Encode(&b, im); err != nil {
			t.Fatal(err)
		}
		name := id + ".png"
		if err := os.WriteFile(filepath.Join(dir, name), b.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b.Bytes())
		m["materials"].(map[string]any)[id] = map[string]any{"source": name, "source_sha256": hex.EncodeToString(sum[:])}
	}
	return dir, m
}

func TestConnectedRailRequiresExactlyThreeOptionalMaterials(t *testing.T) {
	dir, m := railTestFixture(t)
	if _, err := loadConnectionInputs(writeConnectedManifest(t, dir, m)); err != nil {
		t.Fatalf("valid rail manifest rejected: %v", err)
	}
	for _, kind := range []string{"rail", "timber", "steel", "stray", "unknown", "duplicate", "core", "border", "unsupported junction"} {
		t.Run(kind, func(t *testing.T) {
			dir, m := railTestFixture(t)
			materials := m["materials"].(map[string]any)
			family := m["families"].([]any)[0].(map[string]any)
			switch kind {
			case "rail", "timber", "steel":
				delete(materials, kind)
			case "stray":
				m["families"] = []any{connectedPathFamily()}
				materials["path"] = materials["road"]
			case "unknown":
				materials["iron"] = materials["steel"]
			case "duplicate":
				m["families"] = append(m["families"].([]any), family)
			case "core":
				family["material"] = "road"
			case "border":
				family["border_material"] = "grass"
			case "unsupported junction":
				family["variants"] = []any{map[string]any{"id": "rail123", "ports": []int{1, 2, 3}}}
			}
			if _, err := loadConnectionInputs(writeConnectedManifest(t, dir, m)); err == nil {
				t.Fatal("accepted invalid rail manifest")
			}
		})
	}
}

func TestConnectedRailCoreWidthProtectsFixedTrackLayers(t *testing.T) {
	for _, width := range []float64{2, 7.5, 8, 11} {
		t.Run(fmt.Sprintf("width_%g", width), func(t *testing.T) {
			dir, m := railTestFixture(t)
			family := m["families"].([]any)[0].(map[string]any)
			family["width"], family["border"] = width, 0
			_, err := loadConnectionInputs(writeConnectedManifest(t, dir, m))
			if width < 8 {
				if err == nil {
					t.Fatal("accepted a rail core too narrow for fixed track layers")
				}
				if !strings.Contains(err.Error(), "rail") || !strings.Contains(err.Error(), "width >=8") {
					t.Fatalf("missing rail width diagnostic: %v", err)
				}
			} else if err != nil {
				t.Fatalf("rejected supported rail core width: %v", err)
			}
		})
	}
}

// The 11-pixel track and the road-wide 22-pixel one (the same track scaled by
// 2, steel 1.4) both render their layers, ports and turnouts.
func TestConnectedRailPNGTrackLayersAndProvenance(t *testing.T) {
	for _, profile := range [][2]float64{{11, 1}, {22, 2}} {
		t.Run(fmt.Sprint(profile[0]), func(t *testing.T) { testConnectedRailPNG(t, profile[0], profile[1]) })
	}
}

func testConnectedRailPNG(t *testing.T, width, border float64) {
	dir, m := railTestFixture(t)
	family := m["families"].([]any)[0].(map[string]any)
	family["width"], family["border"] = width, border
	out := filepath.Join(dir, "out")
	if err := runConnected(writeConnectedManifest(t, dir, m), out, io.Discard); err != nil {
		t.Fatal(err)
	}
	d := connectionRailDimsFor(width)
	if width == 22 && (d.gauge != 9 || math.Abs(d.steel-1.4) > 1e-12 || d.sleeperLength != 15 || d.pitch != 8) {
		t.Fatalf("road-wide track %+v, want gauge 9, steel 1.4, sleepers 15 at pitch 8", d)
	}
	// Straight interior sleepers lie at 16 + n pitches; the gap between.
	gap, tie := 16+d.pitch/2, 16+d.pitch
	for _, scale := range []int{2, 4} {
		tile, err := loadPNG(filepath.Join(out, fmt.Sprintf("rail36-%dx.png", scale)))
		if err != nil {
			t.Fatal(err)
		}
		sample := func(x, y float64) color.NRGBA { return tile.NRGBAAt(int(x*float64(scale)), int(y*float64(scale))) }
		if scale == 2 {
			// Thin-band reduction can overshoot byte RGB in the existing
			// Lanczos filter. Check the visible material separation here;
			// 4x below pins the literal colours and gauge independently.
			if sample(48-d.gauge/2, gap).R < 195 || sample(48+d.gauge/2, gap).R < 195 || sample(48, gap).R > 145 || sample(48, tie).G > 80 {
				t.Fatal("2x reduction lost paired rails, open gauge or timber sleepers")
			}
			continue
		}
		for _, x := range []float64{48 - d.gauge/2, 48 + d.gauge/2} {
			if got := sample(x, gap); got != railTestColors["steel"] {
				t.Fatalf("%dx paired rail at x=%v: %v", scale, x, got)
			}
		}
		// The steel is d.steel wide: ballast just beyond it on either side.
		for _, x := range []float64{48 + d.gauge/2 - d.steel/2 - .3, 48 + d.gauge/2 + d.steel/2 + .3} {
			if got := sample(x, gap); got != railTestColors["rail"] {
				t.Fatalf("%dx steel wider than %v at x=%v: %v", scale, d.steel, x, got)
			}
		}
		if got := sample(48, gap); got != railTestColors["rail"] {
			t.Fatalf("%dx gauge interior: %v", scale, got)
		}
		if got := sample(48, tie); got != railTestColors["timber"] {
			t.Fatalf("%dx sleeper: %v", scale, got)
		}
		if got := sample(48+d.sleeperLength/2-.3, tie); got != railTestColors["timber"] {
			t.Fatalf("%dx sleeper end: %v", scale, got)
		}
		if got := sample(48+d.sleeperLength/2+.3, tie); got != railTestColors["rail"] {
			t.Fatalf("%dx sleeper longer than %v: %v", scale, d.sleeperLength, got)
		}
		if got := sample(48+4*d.scale, gap); got != railTestColors["rail"] {
			t.Fatalf("%dx ballast: %v", scale, got)
		}
		for y := 0; y < 80*scale; y++ {
			for x := 0; x < 96*scale; x++ {
				p := connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}
				if !connectionInside(sharedConnectionGeometry(), p) && tile.NRGBAAt(x, y) != (color.NRGBA{}) {
					t.Fatalf("outside polygon leaked at %v", p)
				}
			}
		}
	}
	b, err := os.ReadFile(filepath.Join(out, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report map[string]json.RawMessage
	if err = json.Unmarshal(b, &report); err != nil {
		t.Fatal(err)
	}
	var profile connectionRailProfile
	if err = json.Unmarshal(report["rail_profile"], &profile); err != nil || profile.Gauge != d.gauge || profile.RailWidth != d.steel || profile.SleeperPitch != d.pitch {
		t.Fatalf("rail profile %+v does not report the track %+v: %v", profile, d, err)
	}
	for _, id := range []string{"rail", "timber", "steel"} {
		got, err := os.ReadFile(filepath.Join(out, id+"-material-source.png"))
		if err != nil {
			t.Fatal(err)
		}
		want, err := os.ReadFile(filepath.Join(dir, id+".png"))
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("%s source changed", id)
		}
	}
	assertConnectedPatchesUseExportedPNGs(t, out)
	assertRailPNGPortProfiles(t, dir, m, out)
	assertRailPNGTurnouts(t, out, d)
}

func assertRailPNGPortProfiles(t *testing.T, dir string, m map[string]any, out string) {
	t.Helper()
	in, err := loadConnectionInputs(writeConnectedManifest(t, dir, m))
	if err != nil {
		t.Fatal(err)
	}
	f := in.manifest.Families[0]
	g := in.manifest.Geometry
	for _, v := range f.Variants {
		for _, port := range v.Ports {
			opposite := (port+2)%6 + 1
			straight := connectionVariant{ID: "reference", Ports: []int{port, opposite}}
			raw := renderConnectionRaw(in, &f, &straight)
			two, err := resizeLanczos(raw, 208, 176)
			if err != nil {
				t.Fatal(err)
			}
			suffix := "36"
			if port == 1 || port == 4 {
				suffix = "14"
			}
			if port == 2 || port == 5 {
				suffix = "25"
			}
			for _, scale := range []int{2, 4} {
				tile, err := loadPNG(filepath.Join(out, fmt.Sprintf("%s-%dx.png", v.ID, scale)))
				if err != nil {
					t.Fatal(err)
				}
				neighbor, err := loadPNG(filepath.Join(out, fmt.Sprintf("rail%s-%dx.png", suffix, scale)))
				if err != nil {
					t.Fatal(err)
				}
				reference := raw
				if scale == 2 {
					reference = two
				}
				// Match the one NRGBA source-over conversion of the real PNG assembly.
				expected := image.NewNRGBA(reference.Bounds())
				draw.Draw(expected, expected.Bounds(), reference, image.Point{}, draw.Over)
				for parity := 0; parity < 2; parity++ {
					origin := connectionPoint{}
					if parity == 1 {
						origin = connectionPoint{71, 40}
					}
					assembled := image.NewNRGBA(image.Rect(-800, -800, 1200, 1200))
					pos := image.Pt(int(origin[0])*scale, int(origin[1])*scale)
					delta := g.Neighbors[port-1]
					draw.Draw(assembled, tile.Bounds().Add(pos), tile, image.Point{}, draw.Over)
					neighborPos := image.Pt(int(origin[0]+delta[0])*scale, int(origin[1]+delta[1])*scale)
					draw.Draw(assembled, neighbor.Bounds().Add(neighborPos), neighbor, image.Point{}, draw.Over)
					axis := connectionRailUnit(delta)
					for y := 0; y < 80*scale; y++ {
						for x := 0; x < 96*scale; x++ {
							p := connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}
							q := p.sub(g.Ports[port-1])
							along := q[0]*axis[0] + q[1]*axis[1]
							across := math.Abs(-q[0]*axis[1] + q[1]*axis[0])
							if math.Abs(along) > 5 || across > f.Width/2+f.Border+2.5 {
								continue
							}
							got := assembled.NRGBAAt(x+pos.X, y+pos.Y)
							want := expected.NRGBAAt(x+4*scale, y+4*scale)
							if !railTestSamePixel(got, want) {
								t.Fatalf("%s %dx parity%d port%d PNG profile at %v: %v != %v", v.ID, scale, parity, port, p, got, want)
							}
						}
					}
				}
			}
		}
	}
}

// Existing output operations may round only partial-alpha RGB by one byte.
func railTestSamePixel(a, b color.NRGBA) bool {
	if a == b {
		return true
	}
	return a.A == b.A && a.A > 0 && a.A < 255 && math.Abs(float64(a.R)-float64(b.R)) <= 1 && math.Abs(float64(a.G)-float64(b.G)) <= 1 && math.Abs(float64(a.B)-float64(b.B)) <= 1
}

// railTestFamilies are the narrow rail profile and the road-wide one.
var railTestFamilies = []connectionFamily{{ID: "rail", Width: 11, Border: 1}, {ID: "rail", Width: 22, Border: 2}}

func TestConnectedRailPortGaugeAndTurnoutTangents(t *testing.T) {
	g := sharedConnectionGeometry()
	for _, f := range railTestFamilies {
		dims := connectionRailDimsFor(f.Width)
		for _, ports := range [][]int{{1, 3}, {1, 5}, {2, 4}, {2, 6}, {3, 5}, {4, 6}, {2, 3, 6}, {3, 4, 6}, {3, 5, 6}} {
			r := connectionRailRoutes(g, f, connectionVariant{Ports: ports})
			for _, port := range ports {
				inward := connectionRailUnit(g.Neighbors[port-1]).mul(-1)
				normal := connectionRailNormal(inward)
				for _, d := range []float64{0, 4, 8} {
					for _, side := range []float64{-1, 1} {
						p := g.Ports[port-1].add(inward.mul(d)).add(normal.mul(side * dims.gauge / 2))
						if got := connectionRouteDistance(r.rails, p); got > 1e-8 {
							t.Fatalf("width %v ports %v port %d d=%v side=%v: protected rail centre drift %v", f.Width, ports, port, d, side, got)
						}
					}
				}
			}
			if len(ports) == 3 {
				// The turnout must preserve both straight spine rails through the centre.
				for y := 0.0; y <= 80; y += .5 {
					for _, x := range []float64{48 - dims.gauge/2, 48 + dims.gauge/2} {
						if connectionRouteDistance(r.rails, connectionPoint{x, y}) > 1e-8 {
							t.Fatalf("width %v turnout %v cuts through spine", f.Width, ports)
						}
					}
				}
			}
		}
	}
}

func TestConnectedRailSleeperArcLengthPhase(t *testing.T) {
	g := sharedConnectionGeometry()
	for _, f := range railTestFamilies {
		pitch := connectionRailDimsFor(f.Width).pitch
		for _, ports := range [][]int{{1, 3}, {1, 4}, {1, 5}, {2, 4}, {2, 5}, {2, 6}, {3, 5}, {3, 6}, {4, 6}, {2, 3, 6}} {
			r := connectionRailRoutes(g, f, connectionVariant{Ports: ports})
			for _, port := range ports {
				tangent := connectionRailUnit(g.Neighbors[port-1]).mul(-1)
				// The port and every pitch inward up to the 8-pixel guard.
				for d := 0.; d <= 8; d += pitch {
					center := g.Ports[port-1].add(tangent.mul(d))
					found := false
					for _, s := range r.sleepers {
						if s.center.sub(center).length() < 1e-8 {
							found = true
							if math.Abs(s.tangent[0]*tangent[1]-s.tangent[1]*tangent[0]) > 1e-8 {
								t.Fatalf("width %v: sleeper rotated in protected guard: ports %v port %d d=%v", f.Width, ports, port, d)
							}
						}
					}
					if !found {
						t.Fatalf("width %v: missing fixed sleeper: ports %v port %d d=%v", f.Width, ports, port, d)
					}
				}
			}
		}
		// A length that is not a multiple of the pitch needs adjusted interior
		// spacing, while reciprocal port guards remain independent of it.
		r := connectionRailGeometry{dims: connectionRailDimsFor(f.Width)}
		points := connectionRailCubic(connectionPoint{0, 0}, connectionPoint{20, 0}, connectionPoint{20, 30}, connectionPoint{40, 30})
		r.interiorTies(points, false)
		length := connectionRailLength(points)
		intervals := int(math.Round(length / pitch))
		if len(r.sleepers) != intervals-1 {
			t.Fatalf("width %v: interior tie count does not fit an integer number of intervals", f.Width)
		}
		for i, s := range r.sleepers {
			p, n := connectionRailAt(points, length*float64(i+1)/float64(intervals))
			if s.center.sub(p).length() > 1e-8 || s.tangent.sub(n).length() > 1e-8 {
				t.Fatal("curved sleeper is not placed by arc length and route normal")
			}
		}
	}
}

// Read the exported PNGs to catch painting rails around the merged ballast,
// clipping a through spine, or leaving branch pairs short of the merge.
func assertRailPNGTurnouts(t *testing.T, out string, dims connectionRailDims) {
	t.Helper()
	for _, c := range []struct {
		mask string
		port int
	}{{"236", 2}, {"346", 4}, {"356", 5}} {
		tile, err := loadPNG(filepath.Join(out, "rail"+c.mask+"-4x.png"))
		if err != nil {
			t.Fatal(err)
		}
		sample := func(p connectionPoint) color.NRGBA { return tile.NRGBAAt(int(p[0]*4), int(p[1]*4)) }
		for y := 0; y < 320; y++ {
			for _, x := range []int{int((48 - dims.gauge/2) * 4), int((48 + dims.gauge/2) * 4)} {
				if got := tile.NRGBAAt(x, y); got != railTestColors["steel"] {
					t.Fatalf("turnout %s through spine at %d,%d: %v", c.mask, x, y, got)
				}
			}
		}
		g := sharedConnectionGeometry()
		inward := g.Neighbors[c.port-1].mul(-1 / g.Neighbors[c.port-1].length())
		a := g.Ports[c.port-1].add(inward.mul(8))
		b := a.add(inward.mul(16))
		d := connectionPoint{48, 40}
		merge := connectionPoint{0, -1}
		if c.port == 5 {
			merge = connectionPoint{0, 1}
		}
		control := d.sub(merge.mul(12))
		for _, u := range []float64{.1, .25, .5, .75, .9, 1} {
			s := 1 - u
			p := a.mul(s * s * s).add(b.mul(3 * s * s * u)).add(control.mul(3 * s * u * u)).add(d.mul(u * u * u))
			tangent := b.sub(a).mul(3 * s * s).add(control.sub(b).mul(6 * s * u)).add(d.sub(control).mul(3 * u * u))
			normal := connectionPoint{-tangent[1], tangent[0]}.mul(1 / tangent.length())
			for _, side := range []float64{-1, 1} {
				point := p.add(normal.mul(side * dims.gauge / 2))
				if got := sample(point); got != railTestColors["steel"] {
					t.Fatalf("turnout %s branch t=%v side=%v at %v: %v", c.mask, u, side, point, got)
				}
			}
		}
	}
}
