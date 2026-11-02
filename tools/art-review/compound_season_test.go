package artreview

import (
	"path/filepath"
	"testing"
)

// A seasonal compound export counts as edited only when a scene artwork,
// not merely the ground meadow, was replaced by an edit of that season.
func TestCompoundSeasonEdited(t *testing.T) {
	defer func() { exportSeason = "" }()
	in := func(sources ...string) compoundInputs {
		c := compoundInputs{Artwork: map[string]compoundSource{}}
		for i, s := range sources {
			c.Artwork[string(rune('a'+i))] = compoundSource{Ref: connectionMaterial{Source: s}}
		}
		return c
	}
	for _, tc := range []struct {
		season  string
		sources []string
		want    bool
	}{
		{"", []string{"art/terrain/hq/hq-autumn-v1.png"}, false},
		{"autumn", []string{"art/terrain/hq/hq-summer-v3.png", "art/terrain/hq/hq-winter-v4.png"}, false},
		{"autumn", []string{"art/terrain/hq/hq-autumn-v1.png", "art/terrain/hq/hq-winter-v4.png"}, true},
		{"spring", []string{"art/terrain/hq/hq-autumn-v1.png"}, false},
		{"spring", []string{"art/terrain/cmp/cmp-spring-v2.png"}, true},
	} {
		exportSeason = tc.season
		if got := compoundSeasonEdited(in(tc.sources...)); got != tc.want {
			t.Errorf("season %q, %v: got %v, want %v", tc.season, tc.sources, got, tc.want)
		}
	}
}

// The installed autumn compounds load their artwork edit; the winter-only
// factory manifest has none, so the generators skip it in autumn.
func TestCompoundAutumnManifests(t *testing.T) {
	defer func() { exportSeason = "" }()
	exportSeason = "autumn"
	for name, want := range map[string]bool{
		"compound-hq-v4.json":             true,
		"compound-depot-v6.json":          true,
		"compound-ruin-v2.json":           true,
		"compound-site-v1.json":           true,
		"compound-factory-summer-v4.json": true,
		"compound-factory-v6.json":        false,
	} {
		in, err := loadCompoundInputs(filepath.Join("..", "..", "art", "pipeline", name))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got := compoundSeasonEdited(in); got != want {
			t.Errorf("%s: autumn artwork edit %v, want %v", name, got, want)
		}
	}
}
