package finmap

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// committedDir holds the imported maps of the full game (just levels-import).
const committedDir = "../../levels/original"

// committedLevels reads the committed maps 00-71 in map order; they need no
// original files.
func committedLevels(t *testing.T) []*Level {
	t.Helper()
	var ls []*Level
	for i := range 72 {
		raw, err := os.ReadFile(filepath.Join(committedDir, "full", fmt.Sprintf("%02d.json", i)))
		if err != nil {
			t.Fatal(err)
		}
		l, err := Parse(raw)
		if err != nil {
			t.Fatalf("map %02d: %v", i, err)
		}
		ls = append(ls, l)
	}
	return ls
}

func TestCommittedIndex(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(committedDir, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var index []struct{ Build, Map, File string }
	if err := json.Unmarshal(raw, &index); err != nil {
		t.Fatal(err)
	}
	// 72 numbered maps and STATMAP.
	if len(index) != 73 || index[0].Map != "00" || index[72].Map != "STATMAP" {
		t.Fatalf("index has %d entries", len(index))
	}
	for _, e := range index {
		if e.Build != "full" || e.File != "full/"+e.Map+".json" {
			t.Errorf("index entry %+v", e)
		}
		if _, err := os.Stat(filepath.Join(committedDir, e.File)); err != nil {
			t.Error(err)
		}
	}
}

func TestCommittedMaps(t *testing.T) {
	terrains, types := len(dosrules.FullTables.Move.Terrains), len(dosrules.FullTables.Move.Types)
	// Winter maps by position in their series of 24: both campaigns, then
	// the two-player series.
	winter := [][]int{{1, 2, 8, 9, 13, 14, 19, 20}, {1, 2, 8, 9, 13, 14, 19, 20}, {1, 2, 7, 8, 13, 14, 19, 20}}
	for i, l := range committedLevels(t) {
		if l.Format != Format || l.Version != Version || l.Source.Build != "full" || l.Source.Map != fmt.Sprintf("%02d", i) {
			t.Fatalf("map %02d: format %q %d source %+v", i, l.Format, l.Version, l.Source)
		}
		in := func(c [2]int) bool { return c[0] >= 0 && c[0] < l.Height && c[1] >= 0 && c[1] < l.Width }
		if len(l.Terrain) != l.Height {
			t.Fatalf("map %02d: %d rows of %d", i, len(l.Terrain), l.Height)
		}
		for _, row := range l.Terrain {
			if len(row) != l.Width || slices.Max(row) >= terrains || slices.Min(row) < 0 {
				t.Fatalf("map %02d: row %v", i, row)
			}
		}
		for _, u := range l.Units {
			if !in(u.Cell) || u.Type < 0 || u.Type >= types || u.Side < Central || u.Side > Neutral {
				t.Errorf("map %02d: unit %+v", i, u)
			}
		}
		hq := map[int]int{}
		for _, b := range l.Buildings {
			if !in(b.Cell) || b.Container < HQBase || b.Container >= Containers {
				t.Errorf("map %02d: building %+v", i, b)
			}
			if b.Kind == HQ.String() {
				hq[b.Owner]++
			}
		}
		if len(hq) != 2 || hq[Central] != 1 || hq[Allied] != 1 {
			t.Errorf("map %02d: HQs %v", i, hq)
		}
		for _, c := range l.Containers {
			for _, u := range c.Units {
				if u.Type < 0 || u.Type >= types {
					t.Errorf("map %02d: container %d holds type %d", i, c.Container, u.Type)
				}
			}
		}

		m := l.Mission
		if m == nil {
			t.Fatalf("map %02d: no mission", i)
		}
		last := i%24 == 23
		if (m.Next == nil) != last || (m.Next != nil && *m.Next != i+1) || m.CampaignEnd != (last && i < 48) {
			t.Errorf("map %02d: next %v campaign end %v", i, m.Next, m.CampaignEnd)
		}
		if m.Winter != slices.Contains(winter[i/24], i%24) {
			t.Errorf("map %02d: winter %v", i, m.Winter)
		}
		if (m.Players == 1) != (i < 48) || l.Computer != (i < 48) {
			t.Errorf("map %02d: players %d computer %v", i, m.Players, l.Computer)
		}
	}
}

func TestParseRejects(t *testing.T) {
	for name, raw := range map[string]string{
		"format":  `{"format":"x","version":1,"source":{},"width":1,"height":1,"terrain":[[0]]}`,
		"version": `{"format":"historyline-level","version":2,"source":{},"width":1,"height":1,"terrain":[[0]]}`,
		"rows":    `{"format":"historyline-level","version":1,"source":{},"width":1,"height":2,"terrain":[[0]]}`,
		"cells":   `{"format":"historyline-level","version":1,"source":{},"width":2,"height":1,"terrain":[[0]]}`,
		"field":   `{"format":"historyline-level","version":1,"source":{},"width":1,"height":1,"terrain":[[0]],"x":1}`,
		"trailer": `{"format":"historyline-level","version":1,"source":{},"width":1,"height":1,"terrain":[[0]]} {}`,
	} {
		if _, err := Parse([]byte(raw)); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	if _, err := Parse([]byte(`{"format":"historyline-level","version":1,"source":{},"width":1,"height":1,"terrain":[[0]]}`)); err != nil {
		t.Error(err)
	}
}
