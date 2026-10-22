package origview

import (
	"io/fs"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

func loadSet(t *testing.T) *data.Set {
	t.Helper()
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	return set
}

func TestCampaign(t *testing.T) {
	hold := 0
	src := &finmap.Level{Format: finmap.Format, Version: finmap.Version,
		Source: finmap.Source{Build: "full", Map: "30"}, Width: 4, Height: 3,
		Terrain: [][]int{{1, 0, 0, Edge}, {16, 0, 0, Edge}, {Edge, Edge, Edge, Edge}},
		Units: []finmap.Unit{
			{Cell: [2]int{0, 1}, Type: 18, Side: finmap.Central},
			{Cell: [2]int{1, 2}, Type: 22, Side: finmap.Allied, Container: &hold},
		},
		Buildings: []finmap.Building{
			{Kind: "hq", Cell: [2]int{0, 0}, Owner: finmap.Allied, Container: 71},
			{Kind: "depot", Cell: [2]int{1, 0}, Owner: finmap.Neutral, Container: 72},
		},
		Containers: []finmap.Container{
			{Container: 71, Kind: "building", Side: finmap.Allied, Units: []finmap.Cargo{{Type: 17}}, Holder: "building"},
			{Container: 0, Kind: "unit", Side: finmap.Allied, Units: []finmap.Cargo{{Type: 18}}, Holder: "unit"},
			// Neutral: the human side's unit, the Charron the Entente's anyway.
			{Container: 72, Kind: "building", Side: finmap.Neutral, Units: []finmap.Cargo{{Type: 18}, {Type: 46}}, Holder: "building"},
			{Container: 71, Kind: "building", Side: finmap.Allied, Units: []finmap.Cargo{{Type: 10}}, Holder: "building", Overridden: true},
		},
		Mission: &finmap.Mission{Players: 1, HumanSides: []int{finmap.Allied}}}
	got, err := Campaign(src, loadSet(t))
	if err != nil {
		t.Fatal(err)
	}
	unit := func(c [2]int) level.Hex {
		p := finmap.Coord(c[0], c[1])
		return got.Lines[p.Line][p.Col]
	}
	if h := unit([2]int{0, 1}); h.Unit != "CP_infantry" || h.Content != nil {
		t.Errorf("map unit: %+v", h)
	}
	if h := unit([2]int{1, 2}); h.Unit != "EC_supplyCar" || h.Content != nil {
		t.Errorf("transporter (its cargo left out): %+v", h)
	}
	if h := unit([2]int{0, 0}); !reflect.DeepEqual(h.Content, &level.Content{Owner: int(game.Entente), Units: []string{"EC_eliteInf", "EC_lArtillery"}}) {
		t.Errorf("HQ content: %+v", h.Content)
	}
	if h := unit([2]int{1, 0}); !reflect.DeepEqual(h.Content, &level.Content{Owner: int(game.None), Units: []string{"EC_infantry", "EC_charron"}}) {
		t.Errorf("neutral depot content: %+v", h.Content)
	}
	if got.MapName != "30" || got.MapNo != 31 || got.Player != game.Entente.String() {
		t.Errorf("header %q %d %q", got.MapName, got.MapNo, got.Player)
	}

	for name, mutate := range map[string]func(l *finmap.Level){
		"unit off the level": func(l *finmap.Level) { l.Units[0].Cell = [2]int{2, 1} },
		"unit on a building": func(l *finmap.Level) { l.Units[0].Cell = [2]int{0, 0} },
		"two units on a hex": func(l *finmap.Level) { l.Units[1].Cell = [2]int{0, 1} },
		"ship rear":          func(l *finmap.Level) { l.Units[0].Type = 30 },
		"no such building":   func(l *finmap.Level) { l.Containers[0].Container = 73 },
	} {
		l := *src
		l.Units = append([]finmap.Unit(nil), src.Units...)
		l.Containers = append([]finmap.Container(nil), src.Containers...)
		mutate(&l)
		if _, err := Campaign(&l, loadSet(t)); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

func TestFileName(t *testing.T) {
	cp, ec, two := []int{finmap.Central}, []int{finmap.Allied}, []int{finmap.Central, finmap.Allied}
	for _, c := range []struct {
		m     string
		sides []int
		want  string
	}{
		{"00", cp, "level1_cp"}, {"03", cp, "level4_cp"}, {"23", cp, "level24_cp"},
		{"24", ec, "level1_ec"}, {"47", ec, "level24_ec"},
		{"48", two, "level1_2p"}, {"71", two, "level24_2p"},
		{"05", ec, ""}, {"30", cp, ""}, {"50", cp, ""}, {"72", cp, ""}, {"STATMAP", cp, ""},
	} {
		got, err := FileName(&finmap.Level{Source: finmap.Source{Map: c.m}, Mission: &finmap.Mission{HumanSides: c.sides}})
		if got != c.want || (err == nil) != (c.want != "") {
			t.Errorf("FileName(%s, %v) = %q, %v; want %q", c.m, c.sides, got, err, c.want)
		}
	}
	if _, err := FileName(&finmap.Level{Source: finmap.Source{Map: "03"}}); err == nil {
		t.Error("a map without a mission has a campaign name")
	}
}

// Every committed map makes a campaign level that the remake rules load under
// both rule sets with all of the map's units and building contents.
func TestCommittedCampaign(t *testing.T) {
	names, err := fs.Glob(levels.Original, "original/full/[0-9][0-9].json")
	if err != nil {
		t.Fatal(err)
	}
	set := loadSet(t)
	for _, name := range names {
		raw, err := fs.ReadFile(levels.Original, name)
		if err != nil {
			t.Fatal(err)
		}
		src, err := finmap.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		stem := strings.TrimSuffix(path.Base(name), ".json")
		lv, err := Campaign(src, set)
		if err != nil {
			t.Errorf("%s: %v", stem, err)
			continue
		}
		var units, inside int
		for _, line := range lv.Lines {
			for _, h := range line {
				if h.Unit != "" {
					units++
				}
				if h.Content != nil {
					inside += len(h.Content.Units)
				}
			}
		}
		want := 0
		for _, r := range src.Containers {
			if r.Holder == "building" {
				want += len(r.Units)
			}
		}
		if units != len(src.Units) || inside != want {
			t.Errorf("%s: %d units and %d in buildings, the map has %d and %d", stem, units, inside, len(src.Units), want)
		}
		for _, rules := range []game.Rules{game.Classic, game.Standard} {
			if _, err := game.NewWithRules(lv, set, rules); err != nil {
				t.Errorf("%s: %v", stem, err)
			}
		}
	}
}

// Check the actual committed gameplay levels as well as views reconstructed
// from original map records. New campaign transitions must ship with their art.
func TestCommittedCampaignArt(t *testing.T) {
	names, err := fs.Glob(levels.Campaign, "campaign/*.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 72 {
		t.Fatalf("%d campaign levels, want 72", len(names))
	}
	has := terrainCatalog(t)
	for _, name := range names {
		t.Run(strings.TrimSuffix(path.Base(name), ".json"), func(t *testing.T) {
			b, err := fs.ReadFile(levels.Campaign, name)
			if err != nil {
				t.Fatal(err)
			}
			l, err := level.Parse(b)
			if err != nil {
				t.Fatal(err)
			}
			s := mapcompose.Compose(l.Grid(), func(c hexgrid.Coord) mapcompose.Hex {
				h := l.Lines[c.Line][c.Col]
				return mapcompose.Hex{Terrain: h.Terrain, Infra: h.Infra.IDs}
			}, has)
			if len(s.Gaps) != 0 || len(s.Issues) != 0 {
				t.Errorf("%d context gaps and %d compound issues", len(s.Gaps), len(s.Issues))
			}
		})
	}
}
