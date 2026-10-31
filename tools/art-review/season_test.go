package artreview

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// seasonTestPNG writes im to root/rel and returns its pin.
func seasonTestPNG(t *testing.T, root, rel string, im *image.NRGBA) connectionMaterial {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, b.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b.Bytes())
	return connectionMaterial{rel, hex.EncodeToString(sum[:])}
}

func seasonTestMaterial(seed int) *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, 384, 320))
	for y := range 320 {
		for x := range 384 {
			im.SetNRGBA(x, y, color.NRGBA{uint8(x*3 + y + seed*41), uint8(x + y*7 + seed*31), uint8(x ^ y + seed*61), 255})
		}
	}
	return im
}

// seasonTestEdit repaints every pixel and keeps alpha, as a seasonal edit keeps
// every outline.
func seasonTestEdit(im *image.NRGBA, k uint8) *image.NRGBA {
	out := image.NewNRGBA(im.Bounds())
	for i := 0; i < len(im.Pix); i += 4 {
		r, g, b := im.Pix[i], im.Pix[i+1], im.Pix[i+2]
		out.Pix[i], out.Pix[i+1], out.Pix[i+2], out.Pix[i+3] = g^k, b, r, im.Pix[i+3]
	}
	return out
}

func seasonTestJSON(t *testing.T, root, rel string, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(root, rel)
	if err = os.WriteFile(p, b, 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

// seasonTestBundles exports manifest in summer and in autumn and returns both
// reports, decoded into report.
func seasonTestBundles[R any](t *testing.T, root, mode, manifest string) (summerDir, autumnDir string, summer, autumn R) {
	t.Helper()
	summerDir, autumnDir = filepath.Join(root, "build", "summer"), filepath.Join(root, "build", "autumn")
	if err := run([]string{"-root", root, mode, manifest, "-out", summerDir}, io.Discard); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-root", root, "-season", "autumn", mode, manifest, "-out", autumnDir}, io.Discard); err != nil {
		t.Fatal(err)
	}
	if exportSeason != "" {
		t.Fatal("run leaves the export season set")
	}
	for dir, r := range map[string]*R{summerDir: &summer, autumnDir: &autumn} {
		b, err := os.ReadFile(filepath.Join(dir, "report.json"))
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(b, r); err != nil {
			t.Fatal(err)
		}
	}
	return summerDir, autumnDir, summer, autumn
}

// seasonTestAutumnName is a summer runtime file's autumn name: the stem before
// the density suffix gains _au.
func seasonTestAutumnName(p string) string {
	dir, name := path.Split(p)
	i := strings.LastIndex(name, "-")
	return dir + name[:i] + "_au" + name[i:]
}

// seasonTestCompare checks that each autumn file has its summer file's size
// and alpha at every pixel, and that its colours differ exactly where differ
// says the edited sources reach.
func seasonTestCompare(t *testing.T, summerDir, autumnDir string, pairs map[string]string, differ func(summer string) bool) {
	t.Helper()
	if len(pairs) == 0 {
		t.Fatal("no runtime files compared")
	}
	for s, a := range pairs {
		x, err := loadPNG(filepath.Join(summerDir, s))
		if err != nil {
			t.Fatal(err)
		}
		y, err := loadPNG(filepath.Join(autumnDir, a))
		if err != nil {
			t.Fatal(err)
		}
		if x.Bounds() != y.Bounds() {
			t.Fatalf("%s: %v, autumn %s: %v", s, x.Bounds(), a, y.Bounds())
		}
		changed := 0
		for i := 0; i < len(x.Pix); i += 4 {
			if x.Pix[i+3] != y.Pix[i+3] {
				t.Fatalf("%s: alpha differs in %s", s, a)
			}
			if x.Pix[i+3] != 0 && !bytes.Equal(x.Pix[i:i+3], y.Pix[i:i+3]) {
				changed++
			}
		}
		if want := differ(s); want != (changed > 0) {
			t.Errorf("%s: %d autumn pixels changed, want changes %v", s, changed, want)
		}
	}
}

func seasonTestOutputs(outputs []packOutput) []string {
	var paths []string
	for _, o := range outputs {
		paths = append(paths, o.Path)
	}
	return paths
}

func TestSeasonConnectedRoad(t *testing.T) {
	root := t.TempDir()
	materials := map[string]connectionMaterial{}
	for i, m := range []struct{ id, rel, edit string }{
		{"grass", "art/terrain/grass/grass-summer-v1.png", "art/terrain/grass/grass-autumn-v2.png"},
		{"road", "art/terrain/road/road-asphalt-shared-v2.png", "art/terrain/road/road-asphalt-autumn-v1.png"},
		{"water", "art/terrain/water/water-shared-v3.png", ""},
		{"earth", "art/terrain/ground/ground-earth-bank-shared-v2.png", ""},
	} {
		im := seasonTestMaterial(i)
		materials[m.id] = seasonTestPNG(t, root, m.rel, im)
		if m.edit != "" {
			seasonTestPNG(t, root, m.edit, seasonTestEdit(im, 0x55))
		}
	}
	// An older edit loses to the newest one.
	seasonTestPNG(t, root, "art/terrain/grass/grass-autumn-v1.png", seasonTestEdit(seasonTestMaterial(0), 0x11))
	_, m := connectedFixture(t)
	m["materials"] = materials
	// Every port of the junction has its straight continuation for the patches.
	m["families"] = []any{map[string]any{"id": "road", "width": 14, "border": 2, "material": "road", "border_material": "earth", "variants": []any{map[string]any{"id": "road136", "ports": []int{1, 3, 6}}, map[string]any{"id": "road14", "ports": []int{1, 4}}, map[string]any{"id": "road36", "ports": []int{3, 6}}}}}
	manifest := seasonTestJSON(t, root, "road.json", m)
	summerDir, autumnDir, summer, autumn := seasonTestBundles[connectionReport](t, root, "-connected", manifest)

	if autumn.Season != "autumn" || autumn.IDSuffix != "_au" || summer.Season != "" {
		t.Fatalf("seasons %+v, %+v", summer.seasonReport, autumn.seasonReport)
	}
	if got := autumn.Materials["grass"].Source; got != "art/terrain/grass/grass-autumn-v2.png" {
		t.Fatalf("grass source %s", got)
	}
	if autumn.Materials["road"].Source != "art/terrain/road/road-asphalt-autumn-v1.png" || autumn.Materials["water"] != summer.Materials["water"] {
		t.Fatalf("materials %+v", autumn.Materials)
	}
	if len(autumn.Variants) != len(summer.Variants) {
		t.Fatalf("variants %d, summer %d", len(autumn.Variants), len(summer.Variants))
	}
	pairs := map[string]string{}
	for i, v := range summer.Variants {
		a := autumn.Variants[i]
		if a.ID != v.ID+"_au" || !slices.Equal(a.Ports, v.Ports) {
			t.Fatalf("variant %s ports %v, summer %s %v", a.ID, a.Ports, v.ID, v.Ports)
		}
		got, want := seasonTestOutputs(a.Outputs), seasonTestOutputs(v.Outputs)
		for j := range want {
			want[j] = seasonTestAutumnName(want[j])
		}
		if !slices.Equal(got, want) {
			t.Fatalf("outputs %v, want %v", got, want)
		}
		for j, o := range v.Outputs {
			pairs[o.Path] = a.Outputs[j].Path
		}
	}
	pairs["grass-2x.png"] = "grass-2x.png"
	seasonTestCompare(t, summerDir, autumnDir, pairs, func(string) bool { return true })
}

// seasonTestGround writes a ground manifest whose meadow alone has an autumn edit.
func seasonTestGround(t *testing.T, root string) string {
	t.Helper()
	materials := map[string]connectionMaterial{}
	for i, m := range []struct{ id, rel string }{
		{"grass", "art/terrain/grass/grass-summer-v1.png"},
		{"grass_w", "art/terrain/grass/grass-winter-v1.png"},
		{"mud", "art/terrain/ground/ground-shelled-summer-v1.png"},
		{"stone", "art/terrain/ground/ground-stone-summer-v2.png"},
		{"rough_w", "art/terrain/ground/ground-rough-winter-v3.png"},
		{"stoney_w", "art/terrain/ground/ground-stony-winter-v2.png"},
		{"shelled_w", "art/terrain/ground/ground-shelled-winter-v3.png"},
	} {
		im := seasonTestMaterial(i)
		materials[m.id] = seasonTestPNG(t, root, m.rel, im)
		if m.id == "grass" {
			// Only the meadow is repainted; mud and stone keep their summer source.
			seasonTestPNG(t, root, "art/terrain/grass/grass-autumn-v1.png", seasonTestEdit(im, 0x55))
		}
	}
	return seasonTestJSON(t, root, "ground.json", map[string]any{"schema_version": 1, "materials": materials})
}

func TestSeasonGroundContexts(t *testing.T) {
	root := t.TempDir()
	manifest := seasonTestGround(t, root)
	summerDir, autumnDir, summer, autumn := seasonTestBundles[groundReport](t, root, "-ground", manifest)

	if autumn.IDSuffix != "_au" || autumn.Materials["mud"] != summer.Materials["mud"] || autumn.Materials["grass"].Source != "art/terrain/grass/grass-autumn-v1.png" {
		t.Fatalf("autumn sources %+v", autumn.Materials)
	}
	want := map[string]groundProfileReport{}
	for _, p := range summer.Profiles {
		if p.Season == "summer" {
			want[p.ContextKey] = p
		}
	}
	if len(autumn.Profiles) != len(want) || len(want) < 3 {
		t.Fatalf("autumn profiles %d, summer %d", len(autumn.Profiles), len(want))
	}
	pairs := map[string]string{}
	grass := map[string]bool{}
	for _, a := range autumn.Profiles {
		s, ok := want[a.ContextKey]
		if !ok || a.Descriptor != s.Descriptor || a.RuntimeID != s.RuntimeID || a.Canonical != s.Canonical {
			t.Fatalf("autumn profile %s %s has no equal summer profile", a.RuntimeID, a.ContextKey)
		}
		for i, o := range s.Outputs {
			if a.Outputs[i].Path != seasonTestAutumnName(o.Path) {
				t.Fatalf("autumn output %s for %s", a.Outputs[i].Path, o.Path)
			}
			pairs[o.Path] = a.Outputs[i].Path
			grass[o.Path] = s.RuntimeID == "grass" || slices.Contains(s.NeighborIDs[:], "grass")
		}
	}
	if !slices.Contains(seasonTestOutputs(autumn.Exports), "ground/grass_au-2x.png") {
		t.Fatal("no autumn base grass")
	}
	// Contexts without meadow use only the unedited mud and stone.
	seasonTestCompare(t, summerDir, autumnDir, pairs, func(p string) bool { return grass[p] })
}

func TestSeasonForestContexts(t *testing.T) {
	root := t.TempDir()
	grounds, atlases := map[string]connectionMaterial{}, map[string]connectionMaterial{}
	for i, season := range []string{"summer", "winter"} {
		ground := seasonTestMaterial(i)
		grounds[season] = seasonTestPNG(t, root, "art/terrain/grass/grass-"+season+"-v1.png", ground)
		atlases[season] = seasonTestPNG(t, root, "art/terrain/forest/forest-crowns-inset-"+season+"-v3.png", forestTestAtlas())
		if season == "summer" {
			seasonTestPNG(t, root, "art/terrain/grass/grass-autumn-v1.png", seasonTestEdit(ground, 0x55))
			seasonTestPNG(t, root, "art/terrain/forest/forest-crowns-inset-autumn-v1.png", seasonTestEdit(forestTestAtlas(), 0x33))
		}
	}
	manifest := seasonTestJSON(t, root, "forest.json", forestManifest{SchemaVersion: 1, Grounds: grounds, Atlases: atlases})
	summerDir, autumnDir, summer, autumn := seasonTestBundles[forestReport](t, root, "-forest", manifest)

	if autumn.Atlases["summer"].Source != "art/terrain/forest/forest-crowns-inset-autumn-v1.png" || autumn.Atlases["winter"] != summer.Atlases["winter"] {
		t.Fatalf("autumn atlases %+v", autumn.Atlases)
	}
	want := map[string]forestProfileReport{}
	for _, p := range summer.Profiles {
		if p.Season == "summer" {
			want[p.ContextKey] = p
		}
	}
	if len(autumn.Profiles) != len(want) || len(want) < 3 {
		t.Fatalf("autumn profiles %d, summer %d", len(autumn.Profiles), len(want))
	}
	pairs := map[string]string{}
	for _, a := range autumn.Profiles {
		s, ok := want[a.ContextKey]
		if !ok || a.Descriptor != s.Descriptor || len(a.Instances) != len(s.Instances) {
			t.Fatalf("autumn profile %s %s has no equal summer profile", a.RuntimeID, a.ContextKey)
		}
		for i, o := range s.Outputs {
			if a.Outputs[i].Path != seasonTestAutumnName(o.Path) {
				t.Fatalf("autumn output %s for %s", a.Outputs[i].Path, o.Path)
			}
			pairs[o.Path] = a.Outputs[i].Path
		}
	}
	seasonTestCompare(t, summerDir, autumnDir, pairs, func(string) bool { return true })
}

func TestSeasonCompoundTiles(t *testing.T) {
	root := t.TempDir()
	grounds := map[string]connectionMaterial{}
	var scenes []compoundSceneInput
	for i, season := range []string{"summer", "winter"} {
		ground := seasonTestMaterial(i)
		grounds[season] = seasonTestPNG(t, root, "art/terrain/grass/grass-"+season+"-v1.png", ground)
		art := image.NewNRGBA(image.Rect(0, 0, 1024, 1024))
		for y := 440; y < 600; y++ {
			for x := 400; x < 620; x++ {
				art.SetNRGBA(x, y, color.NRGBA{uint8(150 + x%50), uint8(30 + y%40), 10, uint8(255 - (x+y)%3*40)})
			}
		}
		rel := "art/terrain/cmp/cmp-" + season + "-v4.png"
		scenes = append(scenes, compoundSceneInput{ID: "depot-" + season, Artwork: seasonTestPNG(t, root, rel, art), Inset: .9})
		if season == "summer" {
			seasonTestPNG(t, root, "art/terrain/grass/grass-autumn-v1.png", seasonTestEdit(ground, 0x55))
			seasonTestPNG(t, root, "art/terrain/cmp/cmp-autumn-v1.png", seasonTestEdit(art, 0x33))
		}
	}
	manifest := seasonTestJSON(t, root, "depot.json", compoundManifest{SchemaVersion: 1, Grounds: grounds, Scenes: scenes})
	summerDir, autumnDir, summer, autumn := seasonTestBundles[compoundReport](t, root, "-compound", manifest)

	if len(autumn.Scenes) != 1 || autumn.Scenes[0].ID != "depot-summer" || autumn.Scenes[0].Season != "summer" {
		t.Fatalf("autumn scenes %+v", autumn.Scenes)
	}
	if _, ok := autumn.Grounds["winter"]; ok {
		t.Fatal("autumn export kept the winter ground")
	}
	s, a := summer.Scenes[0], autumn.Scenes[0]
	if s.ID != "depot-summer" || a.Transform != s.Transform || len(a.Tiles) != len(s.Tiles) {
		t.Fatalf("summer %s, autumn transform %+v", s.ID, a.Transform)
	}
	pairs := map[string]string{"ground/grass-2x.png": "ground/grass_au-2x.png", "ground/grass-4x.png": "ground/grass_au-4x.png"}
	for i, tile := range s.Tiles {
		if a.Tiles[i].RuntimeID != tile.RuntimeID+"_au" || a.Tiles[i].Origin != tile.Origin {
			t.Fatalf("tile %s at %v, summer %s at %v", a.Tiles[i].RuntimeID, a.Tiles[i].Origin, tile.RuntimeID, tile.Origin)
		}
		for j, o := range tile.Outputs {
			if a.Tiles[i].Outputs[j].Path != seasonTestAutumnName(o.Path) {
				t.Fatalf("autumn tile %s for %s", a.Tiles[i].Outputs[j].Path, o.Path)
			}
			pairs[o.Path] = a.Tiles[i].Outputs[j].Path
		}
	}
	seasonTestCompare(t, summerDir, autumnDir, pairs, func(string) bool { return true })
}

func TestSeasonRejections(t *testing.T) {
	root := t.TempDir()
	grass := seasonTestMaterial(0)
	materials := map[string]connectionMaterial{}
	for i, id := range []string{"grass", "road", "water", "earth"} {
		materials[id] = seasonTestPNG(t, root, "art/terrain/"+id+"/"+id+"-shared-v1.png", seasonTestMaterial(i))
	}
	_, m := connectedFixture(t)
	m["materials"] = materials
	manifest := seasonTestJSON(t, root, "road.json", m)
	for _, tt := range []struct {
		args []string
		want string
	}{
		{[]string{"-season", "winter", "-connected", manifest}, "autumn or spring"},
		{[]string{"-season", "autumn", "-fog", manifest}, "no season"},
		{[]string{"-season", "autumn", "-unit", manifest}, "no season"},
		{[]string{"-season", "autumn", "-connected", manifest}, "no autumn edit"},
	} {
		err := run(append([]string{"-root", root, "-out", filepath.Join(root, "out")}, tt.args...), io.Discard)
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%v: %v, want %q", tt.args, err, tt.want)
		}
	}
	// A winter family has no autumn form, and an edit keeps the source's size.
	seasonTestPNG(t, root, "art/terrain/grass/grass-autumn-v1.png", seasonTestEdit(grass, 0x55))
	for _, v := range m["families"].([]any)[0].(map[string]any)["variants"].([]any) {
		v.(map[string]any)["id"] = v.(map[string]any)["id"].(string) + "_w"
	}
	winter := seasonTestJSON(t, root, "winter.json", m)
	if err := run([]string{"-root", root, "-season", "autumn", "-connected", winter, "-out", filepath.Join(root, "out")}, io.Discard); err == nil || !strings.Contains(err.Error(), "only winter") {
		t.Errorf("winter family: %v", err)
	}
	seasonTestPNG(t, root, "art/terrain/grass/grass-autumn-v2.png", image.NewNRGBA(image.Rect(0, 0, 400, 320)))
	if err := run([]string{"-root", root, "-season", "autumn", "-connected", manifest, "-out", filepath.Join(root, "out")}, io.Discard); err == nil || !strings.Contains(err.Error(), "geometry") {
		t.Errorf("resized edit: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "out")); !os.IsNotExist(err) {
		t.Fatalf("rejected export wrote output: %v", err)
	}
}
