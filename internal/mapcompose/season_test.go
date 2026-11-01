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
	for id, want := range map[string]bool{
		"tile/trench-s12": true, "standing/rBridge25": true, "tile/barbedWire": true,
		"tile/rail14": true, "standing/harbourW": true, "standing/hangar": true, "tile/hangar": false, "tile/road36": false, "tile/cross14": false, "tile/water": false, "tile/hqNE": false,
	} {
		if got := SeasonNeutral(id); got != want {
			t.Errorf("SeasonNeutral(%s) = %v", id, got)
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
// draw the summer scene and count every drawn ID but the season-neutral ones
// as a fallback; with every
// seasonal form they draw it in that form, the shadows and the layers
// unchanged.
func TestComposeSeason(t *testing.T) {
	have := indexIDs(t)
	summerHas := func(id string) bool { return have[id] && SummerID(id) == id } // the catalog without seasonal art
	for _, name := range []string{"1915-08:cp", "1918-02:cp", "1916-12:cp"} {
		l := campaignLevel(t, name)
		g, at := l.Grid(), LevelHex(l)
		base := Compose(g, at, summerHas)
		for _, season := range []Season{Summer, Winter} {
			if s := ComposeSeason(g, at, summerHas, Everywhere(season)); !reflect.DeepEqual(s, base) {
				t.Errorf("%s in %s differs from Compose", name, season)
			}
		}
		if mostlyWinter(g, at) {
			continue
		}
		for _, season := range []Season{Autumn, Spring} {
			s := ComposeSeason(g, at, summerHas, Everywhere(season))
			drawn := 0
			for _, id := range drawnIDs(base) {
				if id != "" && !SeasonNeutral(id) {
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
			s = ComposeSeason(g, at, all, Everywhere(season))
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

// A level's Seasons draw single hexes in their season (SeasonAt): only those
// hexes take their autumn forms.
func TestComposeHexSeasons(t *testing.T) {
	have := indexIDs(t)
	all := func(id string) bool {
		return have[id] || have[strings.TrimSuffix(id, "_au")] && strings.HasSuffix(id, "_au")
	}
	l := campaignLevel(t, "1915-08:cp")
	autumn := map[hexgrid.Coord]bool{{Line: 4, Col: 3}: true, {Line: 4, Col: 4}: true, {Line: 5, Col: 3}: true}
	l.Extra.Seasons = map[hexgrid.Coord]string{}
	for c := range autumn {
		l.Extra.Seasons[c] = string(Autumn)
	}
	g, at := l.Grid(), LevelHex(l)
	base := Compose(g, at, all)
	s := ComposeSeason(g, at, all, SeasonAt(l))
	if len(s.Fallbacks) != 0 {
		t.Errorf("fallbacks %v", s.Fallbacks)
	}
	for line := range s.Cells {
		for col, c := range s.Cells[line] {
			want := base.Cells[line][col].Background
			if autumn[hexgrid.Coord{Line: line, Col: col}] {
				want = SeasonalID(want, Autumn)
			}
			if c.Background != want {
				t.Errorf("[%d, %d]: %s, want %s", line, col, c.Background, want)
			}
		}
	}
	if got := SeasonAt(l)(hexgrid.Coord{}); got != Summer {
		t.Errorf("an unlisted hex is in %s", got)
	}
}

// A hex drawn in winter on a summer level takes the winter look of its
// terrain, hills, mud and stone the original's rough, shelled and stony
// ground, and its shores, roads and standing items follow it.
func TestLevelHexWinter(t *testing.T) {
	l := campaignLevel(t, "1915-08:cp")
	l.Extra.Seasons = map[hexgrid.Coord]string{}
	want := map[string]string{"hills": "rough_w", "mud": "shelled_w", "stone": "stoney_w", "grass": "grass_w", "water": "water"}
	at := LevelHex(l)
	seen := map[string]bool{}
	for li, line := range l.Lines {
		for col, h := range line {
			c := hexgrid.Coord{Line: li, Col: col}
			w, ok := want[h.Terrain]
			if !ok || seen[h.Terrain] {
				continue
			}
			seen[h.Terrain] = true
			l.Extra.Seasons[c] = string(Winter)
			if got := at(c); got.Terrain != w || got.Season != Winter {
				t.Errorf("%s at %v in winter: %s (%s)", h.Terrain, c, got.Terrain, got.Season)
			}
			if l.Lines[li][col].Terrain != h.Terrain {
				t.Errorf("the level's terrain at %v changed", c)
			}
		}
	}
	if len(seen) < 4 {
		t.Fatalf("1915-08 has only %v", seen)
	}
	if got := at(hexgrid.Coord{}); got.Season != "" {
		t.Errorf("a hex in the map's season has Season %q", got.Season)
	}
}

// A context next to a hex drawn in another season continues as if that hex
// were in its own season: a winter east adds no undefined context, and
// LevelHexIn draws every hex in its season.
func TestContextAcrossSeasons(t *testing.T) {
	have := indexIDs(t)
	has := func(id string) bool { return have[id] }
	l := campaignLevel(t, "1915-08:cp")
	undefined := func() int {
		n := 0
		for _, gap := range Compose(l.Grid(), LevelHex(l), has).Gaps {
			if gap.Reason == Undefined {
				n++
			}
		}
		return n
	}
	before := undefined()
	l.Extra.Seasons = map[hexgrid.Coord]string{}
	for li, line := range l.Lines {
		for col := range line {
			if col >= 8 {
				l.Extra.Seasons[hexgrid.Coord{Line: li, Col: col}] = string(Winter)
			}
		}
	}
	if got := undefined(); got != before {
		t.Errorf("%d undefined contexts with a winter east, %d without", got, before)
	}
	at := LevelHexIn(l, Winter)
	if h := at(hexgrid.Coord{Line: 4, Col: 2}); h.Season != Winter || (h.Terrain != "water" && !strings.HasSuffix(h.Terrain, "_w")) {
		t.Errorf("a summer hex in winter: %+v", h)
	}
	if h := LevelHexIn(l, Summer)(hexgrid.Coord{Line: 4, Col: 9}); h.Season != Summer || strings.HasSuffix(h.Terrain, "_w") {
		t.Errorf("a winter hex in summer: %+v", h)
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
