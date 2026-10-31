package hlart

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite testdata/repo/art/index.json")

const fixture = "testdata/repo"

func TestFixtureIndexIsCurrent(t *testing.T) {
	if *update {
		var out bytes.Buffer
		if err := Run([]string{"index", "-root", fixture}, &out, &out); err != nil {
			t.Fatal(err, out.String())
		}
	}
	problems, err := check(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 0 {
		t.Fatalf("fixture problems:\n%s", strings.Join(problems, "\n"))
	}
}

func TestIndexContent(t *testing.T) {
	idx, problems, err := build(fixture)
	if err != nil || len(problems) != 0 {
		t.Fatal(err, problems)
	}
	ids := []string{}
	for _, a := range idx.Assets {
		ids = append(ids, a.ID)
	}
	if want := []string{"interface/action", "terrain/road", "units/infantry"}; !slices.Equal(ids, want) {
		t.Fatalf("ids %v, want %v", ids, want)
	}
	road := idx.Assets[1]
	if road.Masters[0].Path != "art/terrain/road/road-summer-v1.png" || road.Masters[0].Width != 16 || road.Masters[0].Height != 12 {
		t.Errorf("road master %+v", road.Masters[0])
	}
	if road.BasePrompt != "Top-down painted road on meadow, flat-topped hex." {
		t.Errorf("base prompt not unfenced: %q", road.BasePrompt)
	}
	if !slices.Equal(road.Avoid, []string{"Lane markings in yellow", "Kerbs"}) || len(road.Variants) != 2 || road.Variants[1].RuntimeIDs[0] != "road25" {
		t.Errorf("road avoid %v variants %+v", road.Avoid, road.Variants)
	}
	if road.Runtime.Count != 4 || !slices.Equal(road.Runtime.Densities, []int{2, 4}) ||
		!slices.Equal(road.Runtime.Samples, []string{"assets/terrain/road/road14-4x.png", "assets/terrain/road/road25-4x.png"}) {
		t.Errorf("road runtime %+v", road.Runtime)
	}
	units := idx.Assets[2].Runtime.Samples
	if len(units) != maxSamples || units[0] != "assets/units/infantry/cp/n-normal-2x.png" || units[5] != "assets/units/infantry/cp/ne-normal-2x.png" {
		t.Errorf("unit samples %v", units)
	}
	if s := idx.Assets[0].Runtime.Samples; s[0] != "assets/interface/action/attack-normal-2x.png" {
		t.Errorf("interface samples %v, want the normal state first", s)
	}
}

func TestRuntimeSamplesUseHighestDensityPerID(t *testing.T) {
	entries := []runtimeEntry{
		{ID: "tile/road14", Density: 4, Path: "terrain/road/road14-4x.png", Width: 16},
		{ID: "tile/road14", Density: 2, Path: "terrain/road/road14-2x.png", Width: 8},
		{ID: "tile/road25", Density: 2, Path: "terrain/road/road25-2x.png", Width: 8},
		{ID: "tile/road25", Density: 4, Path: "terrain/road/road25-4x.png", Width: 16},
		{ID: "tile/road14", Density: 8, Path: "terrain/road/palette.json"},
		{ID: "tile/road26", Density: 1, Path: "terrain/road/road26-1x.png", Width: 4},
	}
	for _, tc := range []struct {
		name    string
		entries []runtimeEntry
		want    []string
	}{
		{
			name:    "highest density regardless of entry order",
			entries: entries[:4],
			want:    []string{"assets/terrain/road/road25-4x.png", "assets/terrain/road/road14-4x.png"},
		},
		{
			name:    "mixed densities retain every representative and skip non-images",
			entries: append(slices.Clone(entries[:3]), entries[4:]...),
			want:    []string{"assets/terrain/road/road25-2x.png", "assets/terrain/road/road14-4x.png", "assets/terrain/road/road26-1x.png"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &prompt{RuntimeIDs: []string{"road25", "road14", "road26"}}
			got := summarise("terrain/road", p, tc.entries)
			if !slices.Equal(got.Samples, tc.want) {
				t.Errorf("samples %v, want %v", got.Samples, tc.want)
			}
		})
	}
}

func TestTerrainSamplesIncludeEveryDeclaredVariant(t *testing.T) {
	p := &prompt{RuntimeIDs: []string{"lakeside12", "lakeside123", "lakeside126", "lakeside156", "lakeside16", "lakeside23", "lakeside34", "lakeside3456", "lakeside56"}}
	var entries []runtimeEntry
	for _, id := range p.RuntimeIDs {
		for _, density := range []int{2, 4} {
			entries = append(entries, runtimeEntry{ID: "tile/" + id, Density: density, Path: fmt.Sprintf("terrain/lakeside/%s-%dx.png", id, density), Width: 96 * density})
		}
	}
	got := summarise("terrain/lakeside", p, entries)
	if len(got.Samples) != len(p.RuntimeIDs) {
		t.Fatalf("shore catalogue has %d samples, want all %d", len(got.Samples), len(p.RuntimeIDs))
	}
	for i, id := range p.RuntimeIDs {
		want := "assets/terrain/lakeside/" + id + "-4x.png"
		if got.Samples[i] != want {
			t.Errorf("variant %s: got %s, want %s", id, got.Samples[i], want)
		}
	}
}

func TestCheckFindsProblems(t *testing.T) {
	absent := map[string]string{"terrain type": "runtime ID grass"} // problems a case must not report
	for _, tc := range []struct {
		name, want string
		mutate     func(t *testing.T, root string)
	}{
		{"missing prompt", "art/units/infantry/: no prompt.md", func(t *testing.T, root string) {
			remove(t, root, "art/units/infantry/prompt.md")
		}},
		{"stray folder", "art/terrain/forest/: no prompt.md", func(t *testing.T, root string) {
			mkdir(t, root, "art/terrain/forest")
		}},
		{"art key without folder", "art/interface/action/prompt.md missing", func(t *testing.T, root string) {
			if err := os.RemoveAll(filepath.Join(root, "art/interface/action")); err != nil {
				t.Fatal(err)
			}
		}},
		{"changed master", "art/index.json is stale", func(t *testing.T, root string) {
			appendTo(t, root, "art/terrain/road/road-summer-v1.png", "x")
		}},
		{"missing master", "master gone.png", func(t *testing.T, root string) {
			replace(t, root, "art/terrain/road/prompt.md", "masters: road-summer-v1.png", "masters: road-summer-v1.png, gone.png")
		}},
		{"unlisted master", "art/terrain/road/extra.png: image not listed in masters", func(t *testing.T, root string) {
			appendTo(t, root, "art/terrain/road/extra.png", "")
		}},
		{"missing index", "art/index.json missing", func(t *testing.T, root string) {
			remove(t, root, "art/index.json")
		}},
		{"runtime sha", "assets/terrain/road/road14-2x.png: sha256", func(t *testing.T, root string) {
			appendTo(t, root, "assets/terrain/road/road14-2x.png", "x")
		}},
		{"runtime file gone", "assets/index.json tile/road25", func(t *testing.T, root string) {
			remove(t, root, "assets/terrain/road/road25-2x.png")
		}},
		{"unknown terrain id", "runtime ID road99: neither a terrain type", func(t *testing.T, root string) {
			replace(t, root, "art/terrain/road/prompt.md", "runtime_ids: road14, road25", "runtime_ids: road14, road25, road99")
		}},
		{"terrain type", "art/index.json is stale", func(t *testing.T, root string) {
			replace(t, root, "art/terrain/road/prompt.md", "runtime_ids: road14, road25", "runtime_ids: road14, road25, grass")
		}},
		{"unit of another family", "runtime ID CP_cavalry: drawn by family cavalry", func(t *testing.T, root string) {
			replace(t, root, "art/units/infantry/prompt.md", "CP_infantry, EC_infantry", "CP_infantry, EC_infantry, CP_cavalry")
		}},
		{"interface id", "runtime ID action/move: not in assets/index.json", func(t *testing.T, root string) {
			replace(t, root, "art/interface/action/prompt.md", "runtime_ids: action/attack", "runtime_ids: action/attack, action/move")
		}},
		{"variant id", `variant "Bend" runtime ID road26 is not in runtime_ids`, func(t *testing.T, root string) {
			replace(t, root, "art/terrain/road/prompt.md", "| Bend | road25 |", "| Bend | road26 |")
		}},
		{"wrong id", `id "terrain/street", want "terrain/road"`, func(t *testing.T, root string) {
			replace(t, root, "art/terrain/road/prompt.md", "id: terrain/road", "id: terrain/street")
		}},
		{"unknown layer", `layer "roof", want ground or standing`, func(t *testing.T, root string) {
			replace(t, root, "assets/index.json", `"art":"terrain/road"}`, `"layer":"roof","art":"terrain/road"}`)
		}},
		{"standing layer on a tile", "standing items have IDs standing/<id>", func(t *testing.T, root string) {
			replace(t, root, "assets/index.json", `"art":"terrain/road"}`, `"layer":"standing","art":"terrain/road"}`)
		}},
		{"ground tile off the hex", "want one hex on its centre", func(t *testing.T, root string) {
			replace(t, root, "assets/index.json", `"pack":"p-2x"`, `"pack":"terrain-backgrounds-2x"`)
		}},
		{"standing overhang", "overhangs by more than half a hex", func(t *testing.T, root string) {
			replace(t, root, "assets/index.json", `{"id":"tile/road14","pack":"p-2x","density":2,"path":"terrain/road/road14-2x.png","sha256":"96ed1291de2b253b33b93ba7b23820a6694f6230b401f178584c37a17d7b45c2","width":8,"height":6,"logical_anchor":[4,3]`,
				`{"id":"standing/road14","pack":"terrain-standing-2x","layer":"standing","density":2,"path":"terrain/road/road14-2x.png","sha256":"96ed1291de2b253b33b93ba7b23820a6694f6230b401f178584c37a17d7b45c2","width":8,"height":6,"logical_anchor":[100,3]`)
		}},
		{"stamp outside the stamp pack", "cluster stamps have IDs stamp/", func(t *testing.T, root string) {
			replace(t, root, "assets/index.json", `{"id":"tile/road14","pack":"p-2x"`, `{"id":"stamp/crater/1/1","pack":"p-2x"`)
		}},
		{"wrong kind", `kind "interface", want "unit"`, func(t *testing.T, root string) {
			replace(t, root, "art/units/infantry/prompt.md", "kind: unit", "kind: interface")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := copyFixture(t)
			tc.mutate(t, root)
			problems, err := check(root)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.ContainsFunc(problems, func(p string) bool { return strings.Contains(p, tc.want) }) {
				t.Fatalf("no problem %q in:\n%s", tc.want, strings.Join(problems, "\n"))
			}
			if a := absent[tc.name]; a != "" && slices.ContainsFunc(problems, func(p string) bool { return strings.Contains(p, a) }) {
				t.Fatalf("unexpected problem %q in:\n%s", a, strings.Join(problems, "\n"))
			}
		})
	}
}

func TestParsePromptProblems(t *testing.T) {
	good, err := os.ReadFile(filepath.Join(fixture, "art/terrain/road/prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	src := string(good)
	if _, errs := parsePrompt(src); len(errs) != 0 {
		t.Fatalf("good prompt: %v", errs)
	}
	for _, tc := range []struct{ name, old, new, want string }{
		{"no front matter", "---\nid:", "id:", "missing front matter"},
		{"missing key", "status: candidate\n", "", `front matter key "status" missing`},
		{"unknown key", "status: candidate\n", "status: candidate\nartist: me\n", `unknown front matter key "artist"`},
		{"bad status", "status: candidate", "status: draft", `status "draft"`},
		{"bad kind", "kind: terrain", "kind: tiles", `kind "tiles"`},
		{"empty runtime ids", "runtime_ids: road14, road25", "runtime_ids:", "runtime_ids is empty"},
		{"heading order", "## Post", "## Postprocessing", `want ["Look", "Base prompt", "Variants", "Edges & joins", "Post", "Avoid"]`},
		{"edges outside terrain", "kind: terrain", "kind: unit", `want ["Look", "Base prompt", "Variants", "Post", "Avoid"]`},
		{"empty look", "Grey asphalt with gravel shoulders.\n\nA dashed centre line.", "", "section Look is empty"},
		{"table header", "| Variant | Runtime IDs | Prompt delta |", "| Name | IDs | Delta |", "Variants table header"},
		{"short row", "| Bend | road25 | Turns through 120 degrees. |", "| Bend | road25 |", "has 2 cells"},
		{"preamble", "---\n\n## Look", "---\n\nIntro.\n\n## Look", "text before the first ## heading"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(src, tc.old) {
				t.Fatalf("fixture lacks %q", tc.old)
			}
			_, errs := parsePrompt(strings.Replace(src, tc.old, tc.new, 1))
			if !slices.ContainsFunc(errs, func(e string) bool { return strings.Contains(e, tc.want) }) {
				t.Fatalf("no error %q in %v", tc.want, errs)
			}
		})
	}
}

func TestFrontMatterTolerates(t *testing.T) {
	p, errs := parsePrompt("---\r\nid: terrain/road  # the road\r\ntitle: Road\r\nkind: unit\r\nstatus: approved\r\nruntime_ids: `a`,  b ,\r\nmasters:\r\n---\r\n# Road\r\n## Look\r\nx\r\n## Base prompt\r\ny\r\n## Variants\r\n## Post\r\n## Avoid\r\n")
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	if p.ID != "terrain/road" || !slices.Equal(p.RuntimeIDs, []string{"a", "b"}) || len(p.Masters) != 0 || len(p.Variants) != 0 || len(p.Avoid) != 0 {
		t.Fatalf("%+v", p)
	}
}

func TestSourceOnlyCandidateRequiresMaster(t *testing.T) {
	for _, tc := range []struct {
		name, kind, status, masters string
		wantValid                   bool
	}{
		{"painted screen candidate", "screen", "candidate", "illustration.png", true},
		{"masterless screen candidate", "screen", "candidate", "", false},
		{"placeholder master", "screen", "candidate", "—", false},
		{"revised screen", "screen", "revised", "illustration.png", false},
		{"approved screen", "screen", "approved", "illustration.png", false},
		{"terrain candidate", "terrain", "candidate", "illustration.png", false},
		{"unit candidate", "unit", "candidate", "illustration.png", false},
		{"painted portrait candidate", "portrait", "candidate", "illustration.png", true},
		{"masterless portrait candidate", "portrait", "candidate", "", false},
		{"portrait placeholder master", "portrait", "candidate", "—", false},
		{"revised portrait", "portrait", "revised", "illustration.png", false},
		{"approved portrait", "portrait", "approved", "illustration.png", false},
		{"interface candidate", "interface", "candidate", "illustration.png", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := "---\nid: screens/intermission\ntitle: Intermission\nkind: " + tc.kind +
				"\nstatus: " + tc.status + "\nruntime_ids:\nmasters: " + tc.masters +
				"\n---\n## Look\nQuiet rest stop.\n## Base prompt\nPaint an empty farmyard.\n## Variants\n"
			if tc.kind == "terrain" {
				src += "## Edges & joins\nContinuous meadow.\n"
			}
			src += "## Post\nCandidate master only.\n## Avoid\n- text\n"
			p, errs := parsePrompt(src)
			if tc.wantValid {
				if len(errs) != 0 {
					t.Fatalf("source-only candidate rejected: %v", errs)
				}
				if len(p.RuntimeIDs) != 0 || !slices.Equal(p.Masters, []string{"illustration.png"}) {
					t.Fatalf("candidate source metadata: %+v", p)
				}
			} else if !slices.Contains(errs, "runtime_ids is empty") {
				t.Fatalf("missing runtime IDs accepted: %v", errs)
			}
		})
	}
}

func TestShrink(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 40, 10))
	for x := range 20 {
		for y := range 10 {
			img.Set(x, y, color.White)
		}
	}
	got := shrink(img, 4)
	if b := got.Bounds(); b.Dx() != 4 || b.Dy() != 1 {
		t.Fatalf("size %v, want 4x1", b)
	}
	if r, _, _, a := got.At(0, 0).RGBA(); r != 0xffff || a != 0xffff {
		t.Errorf("left pixel %v", got.At(0, 0))
	}
	if _, _, _, a := got.At(3, 0).RGBA(); a != 0 {
		t.Errorf("right pixel %v", got.At(3, 0))
	}
	if small := image.NewRGBA(image.Rect(0, 0, 3, 3)); shrink(small, 4) != image.Image(small) {
		t.Error("an image within the limit is not kept")
	}
}

func copyFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	err := filepath.WalkDir(fixture, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(fixture, p)
		dst := filepath.Join(root, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, raw, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func remove(t *testing.T, root, rel string) {
	t.Helper()
	if err := os.Remove(filepath.Join(root, rel)); err != nil {
		t.Fatal(err)
	}
}

func mkdir(t *testing.T, root, rel string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, rel), 0o755); err != nil {
		t.Fatal(err)
	}
}

func appendTo(t *testing.T, root, rel, s string) {
	t.Helper()
	f, err := os.OpenFile(filepath.Join(root, rel), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(s); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func replace(t *testing.T, root, rel, old, new string) {
	t.Helper()
	p := filepath.Join(root, rel)
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte(old)) {
		t.Fatalf("%s lacks %q", rel, old)
	}
	if err := os.WriteFile(p, bytes.Replace(raw, []byte(old), []byte(new), 1), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestArtbookCompiles builds docs/artbook against the fixture when typst is installed.
func TestArtbookCompiles(t *testing.T) {
	typst, err := exec.LookPath("typst")
	if err != nil {
		t.Skip("typst not installed")
	}
	out := filepath.Join(t.TempDir(), "artbook.pdf")
	cmd := exec.Command(typst, "compile", "--root", ".", "--font-path", "docs/manual/fonts",
		"--input", "base=/internal/hlart/"+fixture, "--input", "style=false",
		"docs/artbook/artbook.typ", out)
	cmd.Dir = filepath.Join("..", "..")
	if msg, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("typst: %v\n%s", err, msg)
	}
	if info, err := os.Stat(out); err != nil || info.Size() == 0 {
		t.Fatalf("no PDF: %v", err)
	}
}
