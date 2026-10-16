package mapcompose

import (
	"strings"
	"testing"
)

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
		{"mountW_w", "winter", true}, {"house", "summer", true},
		{"house", "winter", false}, {"house_w", "summer", false},
		{"invented_w", "winter", false}, {"", "summer", false},
	} {
		if got := boundaryTerrain(tc.id, tc.season); got != tc.want {
			t.Errorf("%s/%s: got %t, want %t", tc.id, tc.season, got, tc.want)
		}
	}
}
