package hlart

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTerrainMasterNamesRequireFamilySeasonAndVersion(t *testing.T) {
	for _, tc := range []struct {
		key, name string
		wantError bool
	}{
		{"terrain/grass", "grass-summer-v1.png", false},
		{"terrain/grass", "grass-meadow-winter-v2.png", false},
		{"terrain/grass", "grass-flowers-material-summer-v12.png", false},
		{"terrain/water", "water-rocks-shared-v2.png", false},
		{"terrain/rBridge", "rbridge-timber-shared-v1.png", false},
		{"terrain/grass", "grass.png", true},
		{"terrain/grass", "grass_w.png", true},
		{"terrain/grass", "meadow-winter-master-v2.png", true},
		{"terrain/grass", "grass-meadow-master-summer-v2.png", true},
		{"terrain/grass", "forest-crowns-summer-v2.png", true},
		{"terrain/grass", "grassland-summer-v1.png", true},
		{"terrain/grass", "Grass-summer-v1.png", true},
		{"terrain/grass", "grass-summer-v0.png", true},
		{"terrain/grass", "grass-summer-v01.png", true},
		{"terrain/grass", "grass-autumn-v1.png", false},
		{"terrain/road", "road-spring-v3.png", false},
		{"terrain/grass", "grass-fall-v1.png", true},
		{"terrain/grass", "grass-summer-v1.jpg", true},
		{"terrain/grass", "grass-summer-v1.png.bak", true},
		{"units/infantry", "atlas.png", false},
		{"portraits/infantry", "infantry-master.jpg", false},
		{"screens/menu", "menu_w.png", false},
	} {
		t.Run(tc.key+"/"+tc.name, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, "art", filepath.FromSlash(tc.key))
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			// Reuse the fixture's PNG bytes; naming validation must inspect the
			// declared source, without changing the image or runtime IDs.
			raw, err := os.ReadFile(filepath.Join(fixture, "art/units/infantry/atlas.png"))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, tc.name), raw, 0o644); err != nil {
				t.Fatal(err)
			}
			kind, _, _ := strings.Cut(tc.key, "/")
			p := &prompt{ID: tc.key, Kind: kindDirs[kind], Masters: []string{tc.name}}
			_, problems := asset(root, tc.key, p, nil)
			if tc.wantError {
				if len(problems) != 1 || !strings.Contains(problems[0], "terrain master") || !strings.Contains(problems[0], tc.name) {
					t.Fatalf("master %q: want terrain naming problem, got %v", tc.name, problems)
				}
			} else if len(problems) != 0 {
				t.Fatalf("valid master %q rejected: %v", tc.name, problems)
			}
		})
	}
}
