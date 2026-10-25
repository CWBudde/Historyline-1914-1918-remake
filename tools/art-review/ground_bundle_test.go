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
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func groundTestKey(id string, neighbors [6]string) (string, string) {
	season, base := "summer", "grass"
	if strings.HasSuffix(id, "_w") {
		season, base = "winter", "grass_w"
	}
	if id == base {
		for i := range neighbors {
			neighbors[i] = base
		}
	}
	descriptor := fmt.Sprintf("ground-v1\nseason=%s\ncenter=%s\n", season, id)
	for i, label := range []string{"NE", "SE", "S", "SW", "NW", "N"} {
		descriptor += label + "=" + neighbors[i] + "\n"
	}
	sum := sha256.Sum256([]byte(descriptor))
	return hex.EncodeToString(sum[:]), descriptor
}

func groundTestRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func groundTestLoad(t *testing.T, path string) *image.NRGBA {
	t.Helper()
	decoded, err := png.Decode(bytes.NewReader(groundTestRead(t, path)))
	if err != nil {
		t.Fatal(err)
	}
	im := image.NewNRGBA(decoded.Bounds())
	draw.Draw(im, im.Bounds(), decoded, decoded.Bounds().Min, draw.Src)
	return im
}

func groundTestSnapshot(t *testing.T, out, path string, want []byte, hash string) {
	t.Helper()
	got := groundTestRead(t, filepath.Join(out, path))
	sum := sha256.Sum256(got)
	if !bytes.Equal(got, want) || hash != hex.EncodeToString(sum[:]) {
		t.Fatalf("snapshot bytes/hash %s", path)
	}
}

func groundTestOwned(x, y, density int) bool {
	// Integer-scaled half-open polygon edge inequalities; no production geometry.
	px, py := 2*x+1, 2*y+1
	for i, a := range groundTestPolygon {
		b := groundTestPolygon[(i+1)%6]
		dx, dy := int(b[0]-a[0]), int(b[1]-a[1])
		qx, qy := px-2*density*int(a[0]), py-2*density*int(a[1])
		cross := dx*qy - dy*qx
		tieIncluded := dy < 0 || (dy == 0 && dx > 0)
		if cross < 0 || (cross == 0 && !tieIncluded) {
			return false
		}
	}
	return true
}

func groundTestBundleContract(t *testing.T, path, out string, r groundTestReport) {
	t.Helper()
	wantGeometry := connectionGeometry{96, 80, []connectionPoint{{25, 0}, {71, 0}, {96, 40}, {71, 80}, {25, 80}, {0, 40}}, []connectionPoint{{71, -40}, {71, 40}, {0, 80}, {-71, 40}, {-71, -40}, {0, -80}}, []connectionPoint{{83.5, 20}, {83.5, 60}, {48, 80}, {12.5, 60}, {12.5, 20}, {48, 0}}}
	if !reflect.DeepEqual(r.Geometry, wantGeometry) {
		t.Fatal("wrong fixed geometry")
	}
	if !strings.HasPrefix(r.Profile, "ground-v1:") || !strings.Contains(r.Profile, "PRODUCT") || !strings.Contains(r.Profile, "C1") || r.Ownership == "" || len(r.ResamplingChain) != 4 || !strings.Contains(r.Acceptance, "OPEN") {
		t.Fatal("missing declared fixed profile/limits")
	}
	groundTestSnapshot(t, out, r.ManifestSnapshot, groundTestRead(t, path), r.ManifestSHA256)
	var manifest struct {
		Materials map[string]connectionMaterial `json:"materials"`
	}
	if err := json.Unmarshal(groundTestRead(t, path), &manifest); err != nil {
		t.Fatal(err)
	}
	if len(r.Materials) != 7 {
		t.Fatal("source roster")
	}
	sources := map[string]*image.NRGBA{}
	for _, id := range []string{"grass", "grass_w", "mud", "stone", "rough_w", "stoney_w", "shelled_w"} {
		ref, ok := r.Materials[id]
		if !ok || ref.connectionMaterial != manifest.Materials[id] || ref.Dimensions != (dimensions{384, 320}) || ref.RetainedSource != "sources/"+id+".png" {
			t.Fatalf("source report %s: %+v", id, ref)
		}
		sourcePath := filepath.Join(filepath.Dir(path), manifest.Materials[id].Source)
		groundTestSnapshot(t, out, ref.RetainedSource, groundTestRead(t, sourcePath), ref.SourceSHA256)
		sources[id] = groundTestLoad(t, sourcePath)
	}
	expectedScenes := []struct {
		id, kind, scenario string
		parity             int
	}{}
	for _, scenario := range []string{"isolated", "homogeneous"} {
		for _, id := range []string{"mud", "stone", "rough_w", "stoney_w", "shelled_w"} {
			for parity := range 2 {
				expectedScenes = append(expectedScenes, struct {
					id, kind, scenario string
					parity             int
				}{fmt.Sprintf("%s-%s-parity%d", id, scenario, parity), id, scenario, parity})
			}
		}
	}
	for _, season := range []string{"summer", "winter"} {
		for parity := range 2 {
			expectedScenes = append(expectedScenes, struct {
				id, kind, scenario string
				parity             int
			}{fmt.Sprintf("%s-mixed-parity%d", season, parity), season, "mixed", parity})
		}
	}
	profileByKey := map[string]groundTestProfile{}
	expectedKeys := map[string]bool{}
	exports := map[string]packOutput{}
	pixels := map[string]*image.NRGBA{}
	requiredExports := map[string]bool{}
	for _, profile := range r.Profiles {
		key, descriptor := groundTestKey(profile.RuntimeID, profile.NeighborIDs)
		if profile.ContextKey != key || profile.Descriptor != descriptor || profileByKey[key].RuntimeID != "" {
			t.Fatalf("duplicate/incorrect profile descriptor %s", profile.RuntimeID)
		}
		season, base := "summer", "grass"
		if strings.HasSuffix(profile.RuntimeID, "_w") {
			season, base = "winter", "grass_w"
		}
		if _, ok := sources[profile.RuntimeID]; !ok {
			t.Fatal("invented runtime ID")
		}
		canonical := true
		for _, id := range profile.NeighborIDs {
			canonical = canonical && id == base
			if _, ok := sources[id]; !ok || strings.HasSuffix(id, "_w") != strings.HasSuffix(profile.RuntimeID, "_w") {
				t.Fatal("invalid profile neighbor roster")
			}
		}
		if profile.Season != season || profile.Canonical != canonical || len(profile.Outputs) != 2 {
			t.Fatal("incorrect profile declaration")
		}
		if profile.RuntimeID == base && !canonical {
			t.Fatal("base profile not normalized")
		}
		for i, density := range []int{2, 4} {
			dir, stem := "contexts/"+profile.RuntimeID, key
			if canonical {
				dir, stem = "tiles", profile.RuntimeID
				if profile.RuntimeID == base {
					dir = "ground"
				}
			}
			name := fmt.Sprintf("%s/%s-%dx.png", dir, stem, density)
			if profile.Outputs[i].Path != name || profile.Outputs[i].Dimensions != (dimensions{96 * density, 80 * density}) {
				t.Fatalf("profile output %s", name)
			}
			requiredExports[name] = true
		}
		profileByKey[key] = profile
	}
	for _, id := range []string{"grass", "grass_w", "mud", "stone", "rough_w", "stoney_w", "shelled_w"} {
		base := "grass"
		if strings.HasSuffix(id, "_w") {
			base = "grass_w"
		}
		ids := [6]string{}
		for i := range ids {
			ids[i] = base
		}
		key, _ := groundTestKey(id, ids)
		expectedKeys[key] = true
	}
	for _, output := range r.Exports {
		if _, ok := exports[output.Path]; ok {
			t.Fatalf("duplicate export %s", output.Path)
		}
		data := groundTestRead(t, filepath.Join(out, output.Path))
		hash := sha256.Sum256(data)
		if output.SHA256 != hex.EncodeToString(hash[:]) {
			t.Fatalf("export hash %s", output.Path)
		}
		im := groundTestLoad(t, filepath.Join(out, output.Path))
		if im.Bounds().Dx() != output.Dimensions.Width || im.Bounds().Dy() != output.Dimensions.Height {
			t.Fatal("export dimensions")
		}
		exports[output.Path] = output
		pixels[output.Path] = im
	}
	for i, expected := range expectedScenes {
		scene := r.Scenes[i]
		center := [2]int{71 * expected.parity, 40 * expected.parity}
		if scene.ID != expected.id || scene.Parity != expected.parity || scene.Center != center || scene.Canvas != ([2]int{center[0] - 71, center[1] - 80}) || len(scene.Cells) != 7 || len(scene.Sites) != 19 {
			t.Fatalf("scene layout %s: %+v", expected.id, scene)
		}
		world := groundTestWorld()
		base := "grass"
		if strings.HasSuffix(expected.kind, "_w") || expected.kind == "winter" {
			base = "grass_w"
		}
		for o := range world {
			world[o] = base
			if expected.scenario == "homogeneous" {
				world[o] = expected.kind
			}
		}
		visible := [][2]int{{0, 0}}
		visible = append(visible, groundTestDeltas[:]...)
		if expected.scenario == "isolated" {
			world[[2]int{}] = expected.kind
		}
		if expected.scenario == "mixed" {
			ids := []string{"mud", "stone", "grass", "mud", "stone", "grass", "mud"}
			if expected.kind == "winter" {
				ids = []string{"rough_w", "stoney_w", "shelled_w", "grass_w", "rough_w", "stoney_w", "shelled_w"}
			}
			for j, o := range visible {
				world[o] = ids[j]
			}
		}
		reportedWorld := map[[2]int]string{}
		for _, site := range scene.Sites {
			local := [2]int{site.Origin[0] - center[0], site.Origin[1] - center[1]}
			if world[local] != site.RuntimeID || reportedWorld[site.Origin] != "" {
				t.Fatalf("wrong/multiple scene site %s %+v", scene.ID, site)
			}
			reportedWorld[site.Origin] = site.RuntimeID
		}
		for j, o := range visible {
			cell := scene.Cells[j]
			wantOrigin := [2]int{center[0] + o[0], center[1] + o[1]}
			if cell.Label != []string{"C", "NE", "SE", "S", "SW", "NW", "N"}[j] || cell.Origin != wantOrigin || cell.RuntimeID != world[o] {
				t.Fatalf("cell layout %s %+v", scene.ID, cell)
			}
			var ids [6]string
			for n, d := range groundTestDeltas {
				ids[n] = world[[2]int{o[0] + d[0], o[1] + d[1]}]
				if ids[n] == "" || reportedWorld[[2]int{wantOrigin[0] + d[0], wantOrigin[1] + d[1]}] != ids[n] {
					t.Fatal("scene context is not in the 19 actual sites")
				}
			}
			key, _ := groundTestKey(cell.RuntimeID, ids)
			if cell.NeighborIDs != ids || cell.ContextKey != key || profileByKey[key].RuntimeID != cell.RuntimeID {
				t.Fatalf("wrong context %s %+v", scene.ID, cell)
			}
			expectedKeys[key] = true
		}
		if len(scene.Outputs) != 2 {
			t.Fatal("scene output count")
		}
		for j, density := range []int{2, 4} {
			name := fmt.Sprintf("scenes/%s-%dx.png", scene.ID, density)
			requiredExports[name] = true
			if scene.Outputs[j] != exports[name] || scene.Outputs[j].Dimensions != (dimensions{238 * density, 240 * density}) {
				t.Fatalf("scene output %s", name)
			}
			got := pixels[name]
			want := image.NewNRGBA(got.Bounds())
			for _, cell := range scene.Cells {
				tilePath := profileByKey[cell.ContextKey].Outputs[j].Path
				tile := pixels[tilePath]
				at := image.Pt((cell.Origin[0]-scene.Canvas[0])*density, (cell.Origin[1]-scene.Canvas[1])*density)
				draw.Draw(want, tile.Bounds().Add(at), tile, image.Point{}, draw.Over)
			}
			if !bytes.Equal(got.Pix, want.Pix) {
				t.Fatalf("scene is not ordinary actual-PNG Over assembly: %s", name)
			}
		}
	}
	if len(expectedKeys) != len(profileByKey) {
		t.Fatalf("profile roster missing or padded: expected %d got %d", len(expectedKeys), len(profileByKey))
	}
	for key := range expectedKeys {
		if profileByKey[key].ContextKey == "" {
			t.Fatalf("missing required profile %s", key)
		}
	}
	if len(requiredExports) != len(exports) {
		t.Fatal("rendered exports not exact")
	}
	for name := range exports {
		if !requiredExports[name] {
			t.Fatalf("extra export %s", name)
		}
	}
	fileCount := 0
	if err := filepath.WalkDir(out, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			fileCount++
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".png") {
			rel, e := filepath.Rel(out, path)
			if e != nil {
				return e
			}
			if !strings.HasPrefix(rel, "sources/") && !requiredExports[rel] {
				t.Errorf("unreported rendered PNG %s", rel)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if fileCount != len(exports)+7+3 {
		t.Fatalf("bundle has %d files; expected PNGs, seven snapshots, manifest/report/HTML", fileCount)
	}
	t.Logf("ground bundle has %d filesystem files", fileCount)
	for _, profile := range r.Profiles {
		for i, density := range []int{2, 4} {
			im := pixels[profile.Outputs[i].Path]
			for y := range 80 * density {
				for x := range 96 * density {
					c := im.NRGBAAt(x, y)
					if groundTestOwned(x, y, density) {
						if c.A != 255 {
							t.Fatalf("owned alpha %s %d,%d = %d", profile.RuntimeID, x, y, c.A)
						}
					} else if c != (color.NRGBA{}) {
						t.Fatal("outside ownership has nonzero RGBA")
					}
				}
			}
			homogeneous := true
			for _, id := range profile.NeighborIDs {
				homogeneous = homogeneous && id == profile.RuntimeID
			}
			if homogeneous {
				old, err := exportConnectionTiles(connectionInputs{manifest: connectionManifest{Geometry: wantGeometry}, materials: map[string]*image.NRGBA{"grass": sources[profile.RuntimeID]}, grassSampling: profile.RuntimeID == "grass"}, nil, nil)
				if err != nil {
					t.Fatal(err)
				}
				var pngBytes bytes.Buffer
				if err = png.Encode(&pngBytes, old[density]); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(groundTestRead(t, filepath.Join(out, profile.Outputs[i].Path)), pngBytes.Bytes()) {
					t.Fatalf("base/homogeneous PNG differs from unchanged material exporter: %s", profile.RuntimeID)
				}
			}
		}
	}
	groundTestActualPNGOracle(t, r, pixels, sources)
}

// Independent texture samplers and global-field transcription for actual
// PNG checks. It deliberately takes all 19 scene sites instead of local contexts.
func groundTestTexture(src *image.NRGBA, x, y float64) color.NRGBA {
	mirror := func(q, period float64) float64 {
		q = math.Mod(q, period)
		if q < 0 {
			q += period
		}
		if q > period/2 {
			q = period - q
		}
		return q / (period / 2)
	}
	bilinear := func(x, y float64) [3]float64 {
		x = mirror(x, 142) * float64(src.Bounds().Dx()-1)
		y = mirror(y, 80) * float64(src.Bounds().Dy()-1)
		ix, iy := int(x), int(y)
		fx, fy := x-float64(ix), y-float64(iy)
		var out [3]float64
		for dy := range 2 {
			for dx := range 2 {
				wx, wy := 1-fx, 1-fy
				if dx == 1 {
					wx = fx
				}
				if dy == 1 {
					wy = fy
				}
				c := src.NRGBAAt(min(ix+dx, src.Bounds().Dx()-1), min(iy+dy, src.Bounds().Dy()-1))
				out[0] += wx * wy * float64(c.R)
				out[1] += wx * wy * float64(c.G)
				out[2] += wx * wy * float64(c.B)
			}
		}
		return out
	}
	a, b := bilinear(x, y), bilinear(x+71, y+40)
	return color.NRGBA{uint8(math.Floor((a[0]+b[0])*.5 + .5000000001)), uint8(math.Floor((a[1]+b[1])*.5 + .5000000001)), uint8(math.Floor((a[2]+b[2])*.5 + .5000000001)), 255}
}

func groundTestWorldPixel(world []groundTestSite, p [2]float64, base string, sources map[string]*image.NRGBA) color.NRGBA {
	bc := groundTestTexture(sources[base], p[0], p[1])
	if base == "grass" {
		bc = groundTestGrassTexture(sources[base], p[0], p[1])
	}
	var sums [3]float64
	total, e := 0., 1.
	for _, site := range world {
		x, y := p[0]-float64(site.Origin[0]), p[1]-float64(site.Origin[1])
		if site.RuntimeID == base {
			e *= groundTestS(groundTestDistance(x, y))
			continue
		}
		t := math.Max(0, 1-((x-48)*(x-48)+(y-40)*(y-40))/6400)
		w := t * t * t
		c := groundTestTexture(sources[site.RuntimeID], p[0], p[1])
		total += w
		sums[0] += w * float64(c.R)
		sums[1] += w * float64(c.G)
		sums[2] += w * float64(c.B)
	}
	out := [3]float64{float64(bc.R), float64(bc.G), float64(bc.B)}
	if total > 0 {
		for i := range out {
			out[i] = out[i]*(1-e) + sums[i]/total*e
		}
	}
	return color.NRGBA{uint8(math.Floor(out[0] + .5)), uint8(math.Floor(out[1] + .5)), uint8(math.Floor(out[2] + .5)), 255}
}

func groundTestActualPNGOracle(t *testing.T, r groundTestReport, pixels map[string]*image.NRGBA, sources map[string]*image.NRGBA) {
	t.Helper()
	profiles := map[string]groundTestProfile{}
	for _, p := range r.Profiles {
		profiles[p.ContextKey] = p
	}
	compared := 0
	for _, scene := range r.Scenes {
		// Both mixed parities, plus isolated and homogeneous fixtures, test genuine
		// different terrain orders. All owned samples are inside the supported field.
		if scene.Parity == 1 && !strings.Contains(scene.ID, "mixed") {
			continue
		}
		base := "grass"
		if strings.Contains(scene.ID, "_w-") || strings.HasPrefix(scene.ID, "winter") {
			base = "grass_w"
		}
		for _, cell := range scene.Cells {
			if cell.RuntimeID == base {
				continue
			}
			profile := profiles[cell.ContextKey]
			for _, xy := range [][2]int{{48, 40}, {48, 1}, {48, 79}, {26, 1}, {70, 1}, {84, 20}, {84, 60}, {12, 20}, {12, 60}, {85, 40}, {90, 40}, {48, 6}, {48, 12}} {
				for i, density := range []int{2, 4} {
					x, y := xy[0]*density, xy[1]*density
					if !groundTestOwned(x, y, density) {
						continue
					}
					point := [2]float64{float64(cell.Origin[0]) + (float64(x)+.5)/float64(density), float64(cell.Origin[1]) + (float64(y)+.5)/float64(density)}
					want := groundTestWorldPixel(scene.Sites, point, base, sources)
					if density == 2 {
						// Raw 4x pixel centers around the 2x output center; independent Lanczos3
						// convolution, with the same byte quantization preceding the reduction.
						var sum [3]float64
						normalization := 0.
						kernel := func(d float64) float64 {
							u := math.Abs(d / 2)
							if u == 0 {
								return 1
							}
							if u >= 3 || u == 1 || u == 2 {
								return 0
							}
							v := math.Pi * u
							return math.Sin(v) / v * math.Sin(v/3) / (v / 3)
						}
						for sy := 2*y - 5; sy <= 2*y+6; sy++ {
							for sx := 2*x - 5; sx <= 2*x+6; sx++ {
								weight := kernel(float64(sx)-(2*float64(x)+.5)) * kernel(float64(sy)-(2*float64(y)+.5))
								p := [2]float64{float64(cell.Origin[0]) + (float64(sx)+.5)/4, float64(cell.Origin[1]) + (float64(sy)+.5)/4}
								c := groundTestWorldPixel(scene.Sites, p, base, sources)
								normalization += weight
								sum[0] += weight * float64(c.R)
								sum[1] += weight * float64(c.G)
								sum[2] += weight * float64(c.B)
							}
						}
						round := func(v float64) uint8 { return uint8(math.Floor(math.Max(0, math.Min(255, v/normalization)) + .5)) }
						want = color.NRGBA{round(sum[0]), round(sum[1]), round(sum[2]), 255}
					}
					got := pixels[profile.Outputs[i].Path].NRGBAAt(x, y)
					if got != want {
						t.Fatalf("independent global-field/filter oracle %s cell %s %dx (%d,%d): got %v want %v", scene.ID, cell.Label, density, x, y, got, want)
					}
					compared++
				}
			}
		}
	}
	if compared < 1000 {
		t.Fatalf("too few independent actual PNG oracle samples: %d", compared)
	}
}
