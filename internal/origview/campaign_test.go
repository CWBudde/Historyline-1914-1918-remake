package origview

import (
	"reflect"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
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
	src := &finmap.Level{
		Format: finmap.Format, Version: finmap.Version,
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
		Mission: &finmap.Mission{Players: 1, HumanSides: []int{finmap.Allied}},
	}
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

// Every numbered original map makes a campaign level that the remake rules
// load under both rule sets with all of the map's units and building
// contents.
func TestOriginalCampaign(t *testing.T) {
	set := loadSet(t)
	for _, src := range originalMaps(t) {
		stem := src.Source.Map
		if stem == "STATMAP" {
			continue
		}
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
		if _, err := game.New(lv, set); err != nil {
			t.Errorf("%s: %v", stem, err)
		}
	}
}

// Check the shipped scenario maps (each once) as well as views
// reconstructed from original map records. New campaign transitions must
// ship with their art.
func TestCommittedCampaignArt(t *testing.T) {
	maps, err := scenario.Shipped(scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	if len(maps) < 49 {
		t.Fatalf("%d scenario maps, want the 52 months", len(maps))
	}
	has := terrainCatalog(t)
	for _, m := range maps {
		t.Run(m.Name, func(t *testing.T) {
			l := m.Level
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
