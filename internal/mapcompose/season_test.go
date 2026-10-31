package mapcompose

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// indexIDs are the IDs of assets/index.json at 4x.
func indexIDs(t *testing.T) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile("../../assets/index.json")
	if err != nil {
		t.Fatal(err)
	}
	var index struct {
		Assets []struct {
			ID      string `json:"id"`
			Density int    `json:"density"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(raw, &index); err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	for _, entry := range index.Assets {
		if entry.Density == 4 {
			have[entry.ID] = true
		}
	}
	return have
}

func TestSeasonalIDs(t *testing.T) {
	for _, c := range []struct {
		id     string
		season Season
		form   string
	}{
		{"tile/road36", Summer, "tile/road36"},
		{"tile/road36", Winter, "tile/road36_w"},
		{"tile/road36", Autumn, "tile/road36_au"},
		{"context/forest/abc", Spring, "context/forest/abc_sp"},
	} {
		form := SeasonalID(c.id, c.season)
		if form != c.form || IDSeason(form) != c.season {
			t.Errorf("%s in %s: %s (%s), want %s", c.id, c.season, form, IDSeason(form), c.form)
		}
		if c.season != Winter && SummerID(form) != c.id {
			t.Errorf("SummerID(%s) = %s", form, SummerID(form))
		}
	}
	if SummerID("tile/road36_w") != "tile/road36_w" {
		t.Error("SummerID strips the winter suffix")
	}
	if got := ShadowID("standing/context/forest/abc_au"); got != "standing/shadow/context/forest/abc" {
		t.Errorf("an autumn item's shadow is %s", got)
	}
	for id, want := range map[string]string{
		"tile/road36_au": "tile/road", "tile/hqNE": "tile/hq", "tile/rBridge25": "tile/rBridge",
		"tile/shore-12-c3_sp": "tile/shore", "standing/harbourW": "standing/harbour",
		"context/forest/abc_au": "context/forest", "stamp/knoll/2_0.2/1": "stamp/knoll",
		"standing/stamp/knoll/2_0.2/1": "standing/stamp/knoll", "standing/scatter/crown/oak-16_au": "standing/scatter/crown",
		"standing/props/trees/1": "standing/props/trees", "tile/barbedWire": "tile/barbedWire",
	} {
		if got := IDFamily(id); got != want {
			t.Errorf("IDFamily(%s) = %s, want %s", id, got, want)
		}
	}
}

func TestLevelSeason(t *testing.T) {
	l := campaignLevel(t, "1915-10:cp")
	if got := LevelSeason(l); got != Autumn {
		t.Errorf("1915-10: %s", got)
	}
	l.Extra = nil
	if got := LevelSeason(l); got != Summer {
		t.Errorf("1915-10 without its scenario: %s", got)
	}
	l = campaignLevel(t, "1916-12:cp")
	if got := LevelSeason(l); got != Winter {
		t.Errorf("1916-12: %s", got)
	}
	l.Extra = nil
	if got := LevelSeason(l); got != Winter {
		t.Errorf("1916-12 without its scenario: %s", got)
	}
	l.Extra = &level.Extra{Season: "monsoon"}
	if got := LevelSeason(l); got != Winter {
		t.Errorf("1916-12 with an unknown season: %s", got)
	}
}

// drawnIDs are the IDs of s an autumn or spring scene swaps, in a fixed order.
func drawnIDs(s Scene) []string {
	var out []string
	for _, line := range s.Cells {
		for _, c := range line {
			out = append(out, c.Background, c.Under)
			out = append(out, c.Overlays...)
			out = append(out, c.Standing...)
			for _, p := range slices.Concat(c.Scatter, c.Ground) {
				out = append(out, p.ID)
			}
		}
	}
	for _, cl := range s.Clusters {
		out = append(out, cl.Stamp, cl.Standing)
	}
	return out
}

// Summer and winter compose as Compose; autumn and spring without seasonal art
// draw the summer scene and count every drawn ID as a fallback; with every
// seasonal form they draw it in that form, the shadows and the layers
// unchanged.
func TestComposeSeason(t *testing.T) {
	have := indexIDs(t)
	summerHas := func(id string) bool { return have[id] }
	for _, name := range []string{"1915-08:cp", "1918-02:cp", "1916-12:cp"} {
		l := campaignLevel(t, name)
		g, at := l.Grid(), LevelHex(l)
		base := Compose(g, at, summerHas)
		for _, season := range []Season{Summer, Winter} {
			if s := ComposeSeason(g, at, summerHas, season); !reflect.DeepEqual(s, base) {
				t.Errorf("%s in %s differs from Compose", name, season)
			}
		}
		if mostlyWinter(g, at) {
			continue
		}
		for _, season := range []Season{Autumn, Spring} {
			s := ComposeSeason(g, at, summerHas, season)
			drawn := 0
			for _, id := range drawnIDs(base) {
				if id != "" {
					drawn++
				}
			}
			total := 0
			for _, n := range s.Fallbacks {
				total += n
			}
			if total != drawn {
				t.Errorf("%s in %s: %d fallbacks for %d drawn IDs", name, season, total, drawn)
			}
			s.Fallbacks = nil
			if !reflect.DeepEqual(s, base) {
				t.Errorf("%s in %s without seasonal art differs from summer", name, season)
			}

			sfx := season.suffix()
			all := func(id string) bool {
				return have[id] || have[strings.TrimSuffix(id, sfx)] && strings.HasSuffix(id, sfx)
			}
			s = ComposeSeason(g, at, all, season)
			if len(s.Fallbacks) != 0 {
				t.Errorf("%s in %s with every seasonal form: fallbacks %v", name, season, s.Fallbacks)
			}
			got, want := drawnIDs(s), drawnIDs(base)
			for i := range got {
				if want[i] == "" && got[i] != "" || want[i] != "" && got[i] != want[i]+sfx {
					t.Fatalf("%s in %s: %s drawn for %s", name, season, got[i], want[i])
				}
			}
			stamped := func(*Cluster) bool { return true }
			for line := range s.Cells {
				for col := range s.Cells[line] {
					c := hexgrid.Coord{Line: line, Col: col}
					if !slices.Equal(s.Cells[line][col].Shadows, base.Cells[line][col].Shadows) ||
						!slices.Equal(s.ShadowsAt(c, stamped), base.ShadowsAt(c, stamped)) {
						t.Fatalf("%s in %s: shadows at %v differ", name, season, c)
					}
					for i, p := range s.ScatterShadowsAt(c, stamped) {
						if p != base.ScatterShadowsAt(c, stamped)[i] {
							t.Fatalf("%s in %s: scattered shadow %v at %v", name, season, p, c)
						}
					}
					bg, st, ov, sd := s.Layers(c, stamped)
					wbg, wst, wov, wsd := base.Layers(c, stamped)
					if summerIDs(bg, st) != summerIDs(wbg, wst) || summerIDs(ov...) != summerIDs(wov...) || summerIDs(sd...) != summerIDs(wsd...) {
						t.Fatalf("%s in %s: layers at %v: %s %s %v %v, summer %s %s %v %v", name, season, c, bg, st, ov, sd, wbg, wst, wov, wsd)
					}
				}
			}
		}
	}
}

// summerIDs joins ids in their summer form.
func summerIDs(ids ...string) string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = SummerID(id)
	}
	return strings.Join(out, " ")
}
