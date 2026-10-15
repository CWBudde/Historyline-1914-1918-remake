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
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"
)

var trenchTestColors = map[string]color.NRGBA{"trench": {65, 43, 27, 255}, "earth": {150, 110, 65, 255}, "timber": {185, 140, 75, 255}, "grass": {30, 80, 35, 255}}

func trenchTestFixture(t *testing.T) (string, map[string]any) {
	t.Helper()
	dir, m := connectedFixture(t)
	variants := []any{
		map[string]any{"id": "trench0", "ports": []int{}, "closed_ring": true},
		map[string]any{"id": "trench1", "ports": []int{1}},
		map[string]any{"id": "trench12", "ports": []int{1, 2}, "dugout_direction": 5},
		map[string]any{"id": "trench124", "ports": []int{1, 2, 4}, "dugout_direction": 5},
		map[string]any{"id": "trench125", "ports": []int{1, 2, 5}, "dugout_direction": 4},
		map[string]any{"id": "trench14", "ports": []int{1, 4}},
		map[string]any{"id": "trench15", "ports": []int{1, 5}, "blind_spurs": []int{3}, "dugout_direction": 4},
		map[string]any{"id": "trench24", "ports": []int{2, 4}, "dugout_direction": 5},
		map[string]any{"id": "trench25", "ports": []int{2, 5}},
		map[string]any{"id": "trench4", "ports": []int{4}},
		map[string]any{"id": "trench5", "ports": []int{5}},
		map[string]any{"id": "trench5a", "ports": []int{5}, "blind_spurs": []int{1, 3}, "dugout_direction": 4},
	}
	m["families"] = []any{map[string]any{"id": "trench", "width": 6, "border": 3, "material": "trench", "border_material": "earth", "variants": variants}}
	for id, c := range trenchTestColors {
		im := image.NewNRGBA(image.Rect(0, 0, 384, 320))
		draw.Draw(im, im.Bounds(), image.NewUniform(c), image.Point{}, draw.Src)
		var b bytes.Buffer
		if err := png.Encode(&b, im); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, id+".png"), b.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b.Bytes())
		m["materials"].(map[string]any)[id] = map[string]any{"source": id + ".png", "source_sha256": hex.EncodeToString(sum[:])}
	}
	return dir, m
}

// These cases catch permissive family/optional-material checks and leaking trench
// semantics into old families; they exercise the strict manifest reader.
func TestConnectedTrenchValidation(t *testing.T) {
	dir, m := trenchTestFixture(t)
	if _, err := loadConnectionInputs(writeConnectedManifest(t, dir, m)); err != nil {
		t.Fatalf("valid trench: %v", err)
	}
	for _, kind := range []string{"trench", "timber", "stray", "duplicate family", "width", "border", "empty", "ring ports", "ring spur", "ring dugout", "duplicate spur", "spur port", "bad spur", "dugout port", "dugout spur", "bad dugout", "other family"} {
		t.Run(kind, func(t *testing.T) {
			dir, m := trenchTestFixture(t)
			f := m["families"].([]any)[0].(map[string]any)
			vs := f["variants"].([]any)
			v := vs[1].(map[string]any)
			switch kind {
			case "trench", "timber":
				delete(m["materials"].(map[string]any), kind)
			case "stray":
				m["materials"].(map[string]any)["steel"] = m["materials"].(map[string]any)["earth"]
			case "duplicate family":
				m["families"] = []any{f, f}
			case "width":
				f["width"] = 3.99
			case "border":
				f["border"] = 1.49
			case "empty":
				v["ports"] = []int{}
			case "ring ports":
				vs[0].(map[string]any)["ports"] = []int{1}
			case "ring spur":
				vs[0].(map[string]any)["blind_spurs"] = []int{1}
			case "ring dugout":
				vs[0].(map[string]any)["dugout_direction"] = 1
			case "duplicate spur":
				v["blind_spurs"] = []int{2, 2}
			case "spur port":
				v["blind_spurs"] = []int{1}
			case "bad spur":
				v["blind_spurs"] = []int{7}
			case "dugout port":
				v["dugout_direction"] = 1
			case "dugout spur":
				v["blind_spurs"] = []int{2}
				v["dugout_direction"] = 2
			case "bad dugout":
				v["dugout_direction"] = -1
			case "other family":
				f["id"] = "road"
				f["material"] = "road"
				delete(m["materials"].(map[string]any), "trench")
				delete(m["materials"].(map[string]any), "timber")
			}
			if _, err := loadConnectionInputs(writeConnectedManifest(t, dir, m)); err == nil {
				t.Fatal("accepted invalid trench manifest")
			}
		})
	}
}

func TestConnectedTrenchExportsAndProvenance(t *testing.T) {
	dir, m := trenchTestFixture(t)
	path := writeConnectedManifest(t, dir, m)
	out := filepath.Join(dir, "out")
	if err := runConnected(path, out, io.Discard); err != nil {
		t.Fatal(err)
	}
	var report map[string]json.RawMessage
	b, err := os.ReadFile(filepath.Join(out, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &report); err != nil {
		t.Fatal(err)
	}
	if len(report["trench_profile"]) == 0 {
		t.Fatal("missing structured trench profile")
	}
	var vr []map[string]any
	if err = json.Unmarshal(report["variants"], &vr); err != nil {
		t.Fatal(err)
	}
	if len(vr) != 12 || vr[0]["closed_ring"] != true || vr[11]["dugout_direction"] != float64(4) {
		t.Fatal("trench semantic metadata lost")
	}
	for _, id := range []string{"trench", "timber"} {
		got, err := os.ReadFile(filepath.Join(out, id+"-material-source.png"))
		if err != nil {
			t.Fatal(err)
		}
		want, err := os.ReadFile(filepath.Join(dir, id+".png"))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("%s provenance altered", id)
		}
	}
	assertConnectedPatchesUseExportedPNGs(t, out)
	in, err := loadConnectionInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	f := in.manifest.Families[0]
	g := in.manifest.Geometry
	for _, v := range f.Variants {
		for _, scale := range []int{2, 4} {
			tile, err := loadPNG(filepath.Join(out, fmt.Sprintf("%s-%dx.png", v.ID, scale)))
			if err != nil {
				t.Fatal(err)
			}
			// Sampling every edge near its middle catches unintended ring/spur/chamber openings.
			for port, p := range g.Ports {
				active := false
				for _, n := range v.Ports {
					active = active || n == port+1
				}
				if active {
					continue
				}
				n := connectionRailUnit(g.Neighbors[port])
				tangent := connectionRailNormal(n)
				for d := -8.; d <= 8; d += .5 {
					q := p.add(tangent.mul(d)).sub(n.mul(.5))
					if got := tile.NRGBAAt(int(q[0]*float64(scale)), int(q[1]*float64(scale))); got.A != 0 {
						t.Fatalf("%s %dx undeclared port%d at%v: %v", v.ID, scale, port+1, q, got)
					}
				}
			}
			if v.ID == "trench0" && tile.NRGBAAt(48*scale, 40*scale).A != 0 {
				t.Fatal("ring center must be transparent")
			}
			if len(v.Ports) == 1 {
				n := connectionRailUnit(g.Neighbors[v.Ports[0]-1])
				q := connectionPoint{48, 40}.sub(n.mul(23))
				if tile.NRGBAAt(int(q[0]*float64(scale)), int(q[1]*float64(scale))).A != 0 {
					t.Fatalf("%s cap continues beyond interior", v.ID)
				}
			}
		}
	}
	assertTrenchPNGPortProfiles(t, in, out)
	out2 := filepath.Join(dir, "again")
	if err := runConnected(path, out2, io.Discard); err != nil {
		t.Fatal(err)
	}
	b2, err := os.ReadFile(filepath.Join(out2, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, b2) {
		t.Fatal("report/output hashes not deterministic")
	}
}

// This compares ordinary decoded-PNG blits to an independent straight reference
// across every actual open port, with all paint (including timber) present.
func assertTrenchPNGPortProfiles(t *testing.T, in connectionInputs, out string) {
	t.Helper()
	f := in.manifest.Families[0]
	g := in.manifest.Geometry
	for _, v := range f.Variants {
		for _, port := range v.Ports {
			opposite := (port+2)%6 + 1
			straight := connectionVariant{ID: "synthetic", Ports: []int{port, opposite}}
			raw := renderConnectionRaw(in, &f, &straight)
			two, err := resizeLanczos(raw, 208, 176)
			if err != nil {
				t.Fatal(err)
			}
			suffix := "14"
			if port == 3 || port == 6 {
				suffix = "36"
			}
			if port == 2 || port == 5 {
				suffix = "25"
			}
			for _, scale := range []int{2, 4} {
				tile, err := loadPNG(filepath.Join(out, fmt.Sprintf("%s-%dx.png", v.ID, scale)))
				if err != nil {
					t.Fatal(err)
				}
				neighbor, err := loadPNG(filepath.Join(out, fmt.Sprintf("trench%s-%dx.png", suffix, scale)))
				if err != nil {
					t.Fatal(err)
				}
				ref := raw
				if scale == 2 {
					ref = two
				}
				expected := image.NewNRGBA(ref.Bounds())
				draw.Draw(expected, expected.Bounds(), ref, image.Point{}, draw.Over)
				for parity := 0; parity < 2; parity++ {
					origin := connectionPoint{}
					if parity == 1 {
						origin = connectionPoint{71, 40}
					}
					assembled := image.NewNRGBA(image.Rect(-800, -800, 1200, 1200))
					pos := image.Pt(int(origin[0])*scale, int(origin[1])*scale)
					draw.Draw(assembled, tile.Bounds().Add(pos), tile, image.Point{}, draw.Over)
					delta := g.Neighbors[port-1]
					np := image.Pt(int(origin[0]+delta[0])*scale, int(origin[1]+delta[1])*scale)
					draw.Draw(assembled, neighbor.Bounds().Add(np), neighbor, image.Point{}, draw.Over)
					axis := connectionRailUnit(delta)
					for y := 0; y < 80*scale; y++ {
						for x := 0; x < 96*scale; x++ {
							p := connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}
							q := p.sub(g.Ports[port-1])
							along := q[0]*axis[0] + q[1]*axis[1]
							across := math.Abs(-q[0]*axis[1] + q[1]*axis[0])
							if math.Abs(along) > 4 || across > 8 {
								continue
							}
							got := assembled.NRGBAAt(x+pos.X, y+pos.Y)
							want := expected.NRGBAAt(x+4*scale, y+4*scale)
							if !railTestSamePixel(got, want) {
								t.Fatalf("%s %dx parity%d port%d profile at%v: %v != %v", v.ID, scale, parity, port, p, got, want)
							}
						}
					}
				}
			}
		}
	}
}

// Literal source colors separate the three materials. Removing wood, flattening
// the excavation's depth shade, or filling the ring hole breaks this test.
func TestConnectedTrenchReadableDepthAndTimber(t *testing.T) {
	dir, m := trenchTestFixture(t)
	in, err := loadConnectionInputs(writeConnectedManifest(t, dir, m))
	if err != nil {
		t.Fatal(err)
	}
	f := in.manifest.Families[0]
	v := f.Variants[0]
	tiles, err := exportConnectionTiles(in, &f, &v)
	if err != nil {
		t.Fatal(err)
	}
	for _, scale := range []int{2, 4} {
		floor, wood, bank := 0, 0, 0
		tile := tiles[scale]
		for y := 0; y < tile.Bounds().Dy(); y++ {
			for x := 0; x < tile.Bounds().Dx(); x++ {
				c := tile.NRGBAAt(x, y)
				if c.A != 255 {
					continue
				}
				if c == (color.NRGBA{47, 31, 19, 255}) {
					floor++
				}
				if c == (color.NRGBA{200, 151, 81, 255}) {
					wood++
				}
				if c.R >= 126 && c.R <= 156 && c.G >= 92 && c.G <= 114 && c.B >= 55 && c.B <= 68 {
					bank++
				}
			}
		}
		if floor < 100*scale || wood < 5*scale || bank < 100*scale {
			t.Fatalf("%dx excavation unreadable: floor%d timber%d bank%d", scale, floor, wood, bank)
		}
		if tile.NRGBAAt(48*scale, 40*scale).A != 0 {
			t.Fatal("ring filled")
		}
	}
}

// The public manifest only needs through 14/25. Explicit synthetic variants
// exercise the north/south contract and six-way topology without an ID fallback.
func TestConnectedTrenchSyntheticSixDirections(t *testing.T) {
	dir, m := trenchTestFixture(t)
	f := m["families"].([]any)[0].(map[string]any)
	f["variants"] = []any{
		map[string]any{"id": "trench14", "ports": []int{1, 4}},
		map[string]any{"id": "trench25", "ports": []int{2, 5}},
		map[string]any{"id": "trench36", "ports": []int{3, 6}},
		map[string]any{"id": "synthetic6", "ports": []int{1, 2, 3, 4, 5, 6}},
	}
	path := writeConnectedManifest(t, dir, m)
	out := filepath.Join(dir, "out")
	if err := runConnected(path, out, io.Discard); err != nil {
		t.Fatal(err)
	}
	in, err := loadConnectionInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	assertTrenchPNGPortProfiles(t, in, out)
}

func TestConnectedTrenchSharedTimberAndFiveFamilies(t *testing.T) {
	dir, m := trenchTestFixture(t)
	materials := m["materials"].(map[string]any)
	families := m["families"].([]any)
	for _, id := range []string{"road", "stream", "path", "rail"} {
		material := id
		if id == "stream" {
			material = "water"
		}
		width := 8
		families = append(families, map[string]any{"id": id, "width": width, "border": 2, "material": material, "border_material": "earth", "variants": []any{map[string]any{"id": id + "14", "ports": []int{1, 4}}}})
	}
	for _, id := range []string{"path", "rail", "steel"} {
		materials[id] = materials["earth"]
	}
	m["families"] = families
	if _, err := loadConnectionInputs(writeConnectedManifest(t, dir, m)); err != nil {
		t.Fatalf("five families with shared timber rejected: %v", err)
	}
}

func TestConnectedTrenchOptionalFieldsRejectedOutsideFamily(t *testing.T) {
	for key, value := range map[string]any{
		"closed_ring": false, "blind_spurs": []int{}, "dugout_direction": 0,
		"CLOSED_RING": false, "BLIND_SPURS": []int{}, "DUGOUT_DIRECTION": 0,
		"Closed_Ring": false, "Blind_Spurs": []int{}, "Dugout_Direction": 0,
	} {
		t.Run(key, func(t *testing.T) {
			dir, m := connectedFixture(t)
			connectedVariant(m)[key] = value
			if _, err := loadConnectionInputs(writeConnectedManifest(t, dir, m)); err == nil {
				t.Fatal("accepted trench-only field outside trench")
			}
		})
	}
}

// encoding/json accepts case aliases for struct tags. This fresh schema
// requires canonical keys so aliases cannot silently override declared values.
func TestConnectedTrenchRejectsNoncanonicalFieldAliases(t *testing.T) {
	for key, value := range map[string]any{
		"CLOSED_RING": false, "BLIND_SPURS": []int{}, "DUGOUT_DIRECTION": 0,
		"Closed_Ring": false, "Blind_Spurs": []int{}, "Dugout_Direction": 0,
	} {
		t.Run(key, func(t *testing.T) {
			dir, m := trenchTestFixture(t)
			family := m["families"].([]any)[0].(map[string]any)
			family["variants"].([]any)[1].(map[string]any)[key] = value
			if _, err := loadConnectionInputs(writeConnectedManifest(t, dir, m)); err == nil {
				t.Fatal("accepted noncanonical trench field alias")
			}
		})
	}
}

// These literal distal points lie beyond the ordinary centerline floor. They
// catch silently dropping either a capped blind spur or the separate chamber.
func TestConnectedTrenchBlindSpurAndSeparateChamber(t *testing.T) {
	dir, m := trenchTestFixture(t)
	in, err := loadConnectionInputs(writeConnectedManifest(t, dir, m))
	if err != nil {
		t.Fatal(err)
	}
	f := in.manifest.Families[0]
	v := f.Variants[11]
	tiles, err := exportConnectionTiles(in, &f, &v)
	if err != nil {
		t.Fatal(err)
	}
	// unit(71,-40)*13 from center; dugout4 chamber + 3 along + 1.5 across.
	for _, p := range []connectionPoint{{59.33, 33.62}, {31.58, 47.53}} {
		for _, scale := range []int{2, 4} {
			c := tiles[scale].NRGBAAt(int(p[0]*float64(scale)), int(p[1]*float64(scale)))
			if c.A != 255 || c.R > 90 {
				t.Fatalf("%dx missing dark excavated blind spur/chamber at%v: %v", scale, p, c)
			}
		}
	}
}

// Explicit blind spurs supersede the ordinary one-port tail: trench5a's SE
// quadrant must stay open instead of inventing a fourth fighting-trench arm.
func TestConnectedTrenchExplicitSpursSuppressImplicitTail(t *testing.T) {
	dir, m := trenchTestFixture(t)
	in, err := loadConnectionInputs(writeConnectedManifest(t, dir, m))
	if err != nil {
		t.Fatal(err)
	}
	f := in.manifest.Families[0]
	v := f.Variants[11]
	tiles, err := exportConnectionTiles(in, &f, &v)
	if err != nil {
		t.Fatal(err)
	}
	for _, scale := range []int{2, 4} {
		if c := tiles[scale].NRGBAAt(int(58.46*float64(scale)), int(45.89*float64(scale))); c.A != 0 {
			t.Fatalf("%dx invented SE arm for explicit-spur variant: %v", scale, c)
		}
	}
}

// A dugout must have an actual room wider than its access neck. Literal
// cross-course samples catch reusing an ordinary full-width capped arm.
func TestConnectedTrenchDugoutRoomAndNarrowNeck(t *testing.T) {
	dir, m := trenchTestFixture(t)
	in, err := loadConnectionInputs(writeConnectedManifest(t, dir, m))
	if err != nil {
		t.Fatal(err)
	}
	f := in.manifest.Families[0]
	v := f.Variants[11]
	tiles, err := exportConnectionTiles(in, &f, &v)
	if err != nil {
		t.Fatal(err)
	}
	for _, scale := range []int{2, 4} {
		tile := tiles[scale]
		// Dugout4 unit axis=(-.8712,.4908), normal=(-.4908,-.8712).
		// Lateral floor point at chamber center + 3.2 normal.
		room := tile.NRGBAAt(int(33.36*float64(scale)), int(44.58*float64(scale)))
		// Neck at 7 along + 2.3 normal lies beyond its 1.75px floor radius.
		neckSide := tile.NRGBAAt(int(40.77*float64(scale)), int(41.43*float64(scale)))
		if room.A != 255 || room.R >= 100 {
			t.Fatalf("%dx missing broad dark dugout room: %v", scale, room)
		}
		if neckSide.A != 255 || neckSide.R < 100 {
			t.Fatalf("%dx dugout neck still full trench width: %v", scale, neckSide)
		}
		// Along the complete access path, the floor remains opaque and dark.
		for d := 0.; d <= 15; d += .5 {
			p := connectionPoint{48 - .8712*d, 40 + .4908*d}
			c := tile.NRGBAAt(int(p[0]*float64(scale)), int(p[1]*float64(scale)))
			if c.A != 255 || c.R >= 100 {
				t.Fatalf("%dx disconnected room entrance at%v: %v", scale, p, c)
			}
		}
	}
}

func TestConnectedTrenchDirectionalInnerWallShadow(t *testing.T) {
	dir, m := trenchTestFixture(t)
	in, err := loadConnectionInputs(writeConnectedManifest(t, dir, m))
	if err != nil {
		t.Fatal(err)
	}
	f := in.manifest.Families[0]
	v := connectionVariant{ID: "synthetic36", Ports: []int{3, 6}}
	r := connectionTrenchRoutes(in.manifest.Geometry, f, v)
	lit := connectionTrenchFeature(in, f, r, connectionPoint{46, 10}, .25)
	shaded := connectionTrenchFeature(in, f, r, connectionPoint{50, 10}, .25)
	if lit.A != 255 || shaded.A != 255 || int(lit.R)-int(shaded.R) < 6 {
		t.Fatalf("inner wall lacks consistent directional depth: lit%v shadow%v", lit, shaded)
	}
}
