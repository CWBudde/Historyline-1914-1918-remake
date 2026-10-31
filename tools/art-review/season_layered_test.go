//go:build campaigngfx

package artreview

// The layered forest exports all 64 summer masks twice (about 90 s):
// go test -tags campaigngfx ./tools/art-review -run '^TestSeasonLayeredForest$' -count=1
import (
	"encoding/json"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func TestSeasonLayeredForest(t *testing.T) {
	root := t.TempDir()
	m := layeredForestManifest{SchemaVersion: 1}
	for i, rel := range []string{"art/terrain/grass/grass-summer-v1.png", "art/terrain/forest/forest-floor-summer-v4.png"} {
		im := seasonTestMaterial(i)
		ref := seasonTestPNG(t, root, rel, im)
		seasonTestPNG(t, root, strings.Replace(rel, "-summer-", "-autumn-", 1), seasonTestEdit(im, 0x55))
		if i == 0 {
			m.Ground = ref
		} else {
			m.Floor = ref
		}
	}
	for _, species := range []string{"oak", "beech", "birch"} {
		m.Atlases = append(m.Atlases, seasonTestPNG(t, root, "art/terrain/forest/forest-"+species+"-summer-v4.png", forestTestAtlas()))
		seasonTestPNG(t, root, "art/terrain/forest/forest-"+species+"-autumn-v1.png", seasonTestEdit(forestTestAtlas(), 0x33))
	}
	manifest := seasonTestJSON(t, root, "layered.json", m)
	summerDir, autumnDir := filepath.Join(root, "summer"), filepath.Join(root, "autumn")
	for _, args := range [][]string{{"-out", summerDir}, {"-season", "autumn", "-out", autumnDir}} {
		if err := run(append([]string{"-root", root, "-forest-layered", manifest}, args...), io.Discard); err != nil {
			t.Fatal(err)
		}
	}
	var summer, autumn []layeredForestEntry
	for dir, entries := range map[string]*[]layeredForestEntry{summerDir: &summer, autumnDir: &autumn} {
		b, err := os.ReadFile(filepath.Join(dir, "index-entries.json"))
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(b, entries); err != nil {
			t.Fatal(err)
		}
	}
	var season struct {
		seasonReport
		Sources []connectionMaterial `json:"sources"`
	}
	b, err := os.ReadFile(filepath.Join(autumnDir, "season.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &season); err != nil {
		t.Fatal(err)
	}
	if season.IDSuffix != "_au" || len(season.Sources) != 5 {
		t.Fatalf("season record %+v", season)
	}
	for _, s := range season.Sources {
		if !strings.Contains(s.Source, "-autumn-") {
			t.Fatalf("autumn export read %s", s.Source)
		}
	}
	if len(autumn) != len(summer) {
		t.Fatalf("autumn entries %d, summer %d", len(autumn), len(summer))
	}
	pairs := map[string]string{}
	for i, s := range summer {
		a := autumn[i]
		if a.ID != s.ID+"_au" || a.Path != seasonTestAutumnName(s.Path) || a.Width != s.Width || a.Height != s.Height || a.Anchor != s.Anchor || a.Pack != s.Pack {
			t.Fatalf("autumn entry %+v, summer %+v", a, s)
		}
		pairs[s.Path] = a.Path
	}
	// Shadows (shadow-*, props-<v>-1-*) are cast from the crowns' alpha alone
	// and keep their colour.
	seasonTestCompare(t, summerDir, autumnDir, pairs, func(p string) bool {
		name := strings.Split(path.Base(p), "-")
		return name[0] != "shadow" && (name[0] != "props" || name[2] != "1")
	})
}
