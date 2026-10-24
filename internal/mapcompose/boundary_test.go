package mapcompose

import (
	"slices"
	"strings"
	"testing"
)

var sharedShoreTestIDs = []string{
	"lakeside12", "lakeside123", "lakeside126", "lakeside156", "lakeside16",
	"lakeside23", "lakeside34", "lakeside3456", "lakeside56",
}

func TestKnownShoreBoundariesValidInBothSeasons(t *testing.T) {
	for _, shore := range sharedShoreTestIDs {
		for _, season := range []string{"summer", "winter"} {
			t.Run(shore+"/"+season, func(t *testing.T) {
				if !boundaryTerrain(shore, season) {
					t.Errorf("known shared shore %s rejected in %s", shore, season)
				}
			})
		}
	}
}

func TestWinterMaterialContextsBesideSharedShores(t *testing.T) {
	for _, terrain := range []string{"rough_w", "stoney_w", "shelled_w", "forest_w"} {
		for _, shore := range sharedShoreTestIDs {
			t.Run(terrain+"/"+shore, func(t *testing.T) {
				g, at := grid(terrain+" "+shore, terrain, terrain+" "+terrain)
				s := Compose(g, at, none)
				s.Gaps = slices.DeleteFunc(s.Gaps, func(g Gap) bool { return g.Family == Shore })
				if len(s.Gaps) != 4 {
					t.Fatalf("got %d gaps, want four material contexts", len(s.Gaps))
				}
				for _, gap := range s.Gaps {
					if gap.Reason != Missing {
						t.Fatalf("shared shore prevented a descriptor: %+v", gap)
					}
					if gap.At.Line == 1 && (!strings.Contains(gap.Descriptor, "NE=grass_w\n") || !strings.Contains(gap.Descriptor, "SE="+terrain+"\n") || !strings.Contains(gap.Descriptor, "NW="+terrain+"\n")) {
						t.Errorf("winter boundary lost its seasonal base or material connections: %s", gap.Descriptor)
					}
				}
			})
		}
	}
}

func TestWinterBoundaryRejectsUnknownAndMixedSeasonNeighbors(t *testing.T) {
	for _, terrain := range []string{"rough_w", "forest_w"} {
		for _, neighbor := range []string{"grass", "house", "forest", "lakeside999", "lakeside12_w"} {
			t.Run(terrain+"/"+neighbor, func(t *testing.T) {
				g, at := grid(terrain+" "+neighbor, terrain, terrain+" "+terrain)
				s := Compose(g, at, none)
				for _, gap := range s.Gaps {
					if gap.At.Line == 1 {
						if gap.Reason != Undefined || !strings.Contains(gap.Detail, "NE="+neighbor) {
							t.Fatalf("invalid winter neighbor acquired a context: %+v", gap)
						}
						return
					}
				}
				t.Fatal("invalid winter neighbor produced no gap")
			})
		}
	}
}

func TestCampaignBoundariesHaveContexts(t *testing.T) {
	for _, name := range []string{"level1_cp", "level2_cp", "level3_cp"} {
		g, at := loadLevel(t, name)
		s := Compose(g, at, func(string) bool { return true })
		if len(s.Gaps) != 0 {
			t.Errorf("%s: %d undefined boundaries, first %+v", name, len(s.Gaps), s.Gaps[0])
		}
	}
}

func TestBoundaryContextsPreserveConnections(t *testing.T) {
	g, at := grid("forest grass", "forest", "hills grass")
	s := Compose(g, at, none)
	for _, gap := range s.Gaps {
		if gap.Reason != Missing {
			t.Fatalf("boundary has no descriptor: %+v", gap)
		}
		if gap.At.Line == 0 && (!strings.Contains(gap.Descriptor, "SE=forest\n") || !strings.Contains(gap.Descriptor, "S=grass\n") || !strings.Contains(gap.Descriptor, "N=forest\n")) {
			t.Errorf("forest boundary: %s", gap.Descriptor)
		}
	}
}

func TestBoundarySeasonValidation(t *testing.T) {
	for _, tc := range []struct {
		id, season string
		want       bool
	}{
		{"water", "winter", true}, {"water", "summer", true},
		{"waterStones", "winter", true}, {"waterStones", "summer", true},
		{"mountW_w", "winter", true}, {"house", "summer", true},
		{"house", "winter", false}, {"house_w", "summer", false},
		{"invented_w", "winter", false}, {"", "summer", false},
		{"lakeside999", "winter", false}, {"lakeside12_w", "winter", false},
	} {
		if got := boundaryTerrain(tc.id, tc.season); got != tc.want {
			t.Errorf("%s/%s: got %t, want %t", tc.id, tc.season, got, tc.want)
		}
	}
}
