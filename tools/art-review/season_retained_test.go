//go:build retainedgfx

package artreview

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// A seasonal bundle is supplemented with the summer profiles alone, rendered
// from the seasonal sources into suffixed files.
func TestRetainedContextSupplementSeason(t *testing.T) {
	root := t.TempDir()
	old := sourceRoot
	sourceRoot = root
	defer func() { sourceRoot = old }()
	manifest := seasonTestGround(t, root)
	var profiles []groundProfileReport
	for _, ids := range [][7]string{{"mud", "grass", "mud", "grass", "stone", "mud", "grass"}, {"shelled_w", "grass_w", "shelled_w", "grass_w", "grass_w", "shelled_w", "grass_w"}} {
		c, err := newGroundContext(ids[0], ids[1:])
		if err != nil {
			t.Fatal(err)
		}
		season, _ := groundSeason(ids[0])
		profiles = append(profiles, groundProfileReport{RuntimeID: c.RuntimeID, Season: season, ContextKey: c.key(), Descriptor: c.descriptor(), NeighborIDs: c.NeighborIDs})
	}
	recipe := seasonTestJSON(t, root, "recipe.json", map[string]any{"profiles": profiles})
	input, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(input)
	bundles := map[string]string{}
	for _, season := range []string{"", "autumn"} {
		bundle := filepath.Join(t.TempDir(), "bundle")
		if err = os.Mkdir(bundle, 0o755); err != nil {
			t.Fatal(err)
		}
		groundTestWrite(t, filepath.Join(bundle, "manifest.json"), input)
		report, _ := json.Marshal(map[string]any{"schema_version": 1, "manifest_sha256": hex.EncodeToString(sum[:]), "profiles": []any{}, "exports": []any{}, "season": season})
		groundTestWrite(t, filepath.Join(bundle, "report.json"), report)
		if err = augmentRetainedContextBundle("ground", manifest, recipe, bundle); err != nil {
			t.Fatal(err)
		}
		bundles[season] = bundle
	}
	if exportSeason != "" {
		t.Fatal("supplement leaves the export season set")
	}
	var summer, autumn groundReport
	for bundle, r := range map[string]*groundReport{bundles[""]: &summer, bundles["autumn"]: &autumn} {
		if err = json.Unmarshal(groundTestRead(t, filepath.Join(bundle, "report.json")), r); err != nil {
			t.Fatal(err)
		}
	}
	if len(summer.Profiles) != 2 || len(autumn.Profiles) != 1 || autumn.Profiles[0].ContextKey != profiles[0].ContextKey {
		t.Fatalf("summer %d profiles, autumn %+v", len(summer.Profiles), autumn.Profiles)
	}
	pairs := map[string]string{}
	for i, o := range summer.Profiles[0].Outputs {
		a := autumn.Profiles[0].Outputs[i].Path
		if a != seasonTestAutumnName(o.Path) {
			t.Fatalf("autumn output %s for %s", a, o.Path)
		}
		pairs[o.Path] = a
	}
	seasonTestCompare(t, bundles[""], bundles["autumn"], pairs, func(string) bool { return true })
	if err = augmentRetainedContextBundle("forest", manifest, recipe, bundles["autumn"]); err == nil {
		t.Fatal("seasonal forest supplement accepted")
	}
}
