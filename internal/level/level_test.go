package level

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func shippedLevels(t *testing.T) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join("..", "..", "levels", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) < 4 {
		t.Fatalf("found %d level files, want at least 4", len(paths))
	}
	return paths
}

func TestRoundTrip(t *testing.T) {
	for _, path := range shippedLevels(t) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			l, err := Parse(raw)
			if err != nil {
				t.Fatal(err)
			}
			out, err := l.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(out, raw) {
				a, b := strings.Split(string(raw), "\n"), strings.Split(string(out), "\n")
				for i := 0; i < len(a) && i < len(b); i++ {
					if a[i] != b[i] {
						t.Fatalf("line %d differs:\nfile:  %q\nsaved: %q", i+1, a[i], b[i])
					}
				}
				t.Fatalf("saved %d lines, file has %d", len(b), len(a))
			}
		})
	}
}

func TestSaveLoad(t *testing.T) {
	l, err := Load(filepath.Join("..", "..", "levels", "level3_cp.json"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "out.json")
	if err := l.Save(path); err != nil {
		t.Fatal(err)
	}
	back, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(l, back) {
		t.Fatal("level changed after Save and Load")
	}
}

func TestParseForms(t *testing.T) {
	l, err := Load(filepath.Join("..", "..", "levels", "level1_cp.json"))
	if err != nil {
		t.Fatal(err)
	}
	if l.MapName != "Pulse" || l.MapNo != 1 || l.Player != "Central Powers" {
		t.Errorf("header = %q %d %q", l.MapName, l.MapNo, l.Player)
	}
	if g := l.Grid(); g != (hexgrid.Grid{Width: 8, Lines: 47}) {
		t.Errorf("Grid() = %+v", g)
	}
	cases := []struct {
		line, col int
		want      Hex
	}{
		{0, 0, Hex{Terrain: "hills"}},
		{1, 0, Hex{Terrain: "grass", Infra: Infra{IDs: []string{"road236"}}}},
		{3, -1, Hex{Terrain: "grass", Infra: Infra{IDs: []string{"stream25", "rBridge36", "road36"}, List: true}, Unit: "CP_lArtillery"}},
		{5, -1, Hex{Terrain: "hqN", Content: &Content{Owner: 1}}},
		{18, -1, Hex{Terrain: "cmpN", Content: &Content{Owner: 0, Units: []string{"CP_hArtillery"}}}},
	}
	for _, c := range cases {
		line := l.Lines[c.line]
		found := false
		for col, h := range line {
			if (c.col < 0 || col == c.col) && reflect.DeepEqual(h, c.want) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("line%d has no hex %+v", c.line+1, c.want)
		}
	}
}

const header = "{\"mapName\": \"m\", \"mapNo\": 1, \"player\": \"p\", \"tiles\": {"

func TestParseStringMapNo(t *testing.T) {
	// generateMap in mapEditor.py writes the map number as a string.
	l, err := Parse([]byte(`{"mapName": "m", "mapNo": "7", "player": "p", "tiles": {"line1": [["grass","",""]]}}`))
	if err != nil {
		t.Fatal(err)
	}
	if l.MapNo != 7 {
		t.Fatalf("MapNo = %d, want 7", l.MapNo)
	}
	out, err := l.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "\t\"mapNo\" :\t7,\n") {
		t.Fatalf("saved mapNo not written as a number like saveData:\n%s", out)
	}
}

func TestParseRejects(t *testing.T) {
	cases := map[string]string{
		"missing line":       header + `"line1": [["grass","",""],["grass","",""]], "line3": [["grass","",""],["grass","",""]]}}`,
		"odd line too long":  header + `"line1": [["grass","",""],["grass","",""]], "line2": [["grass","",""],["grass","",""]]}}`,
		"even line short":    header + `"line1": [["grass","",""],["grass","",""]], "line2": [["grass","",""]], "line3": [["grass","",""]]}}`,
		"two elements":       header + `"line1": [["grass",""]]}}`,
		"five elements":      header + `"line1": [["grass","","",[0],"x"]]}}`,
		"terrain not string": header + `"line1": [[1,"",""]]}}`,
		"infra number":       header + `"line1": [["grass",5,""]]}}`,
		"content no owner":   header + `"line1": [["hqN","","",[]]]}}`,
		"content extra":      header + `"line1": [["hqN","","",[0,[],1]]]}}`,
		"bad line key":       header + `"row1": [["grass","",""]]}}`,
		"no tiles":           header + `}}`,
		"mapNo not a number": `{"mapName": "m", "mapNo": "x", "player": "p", "tiles": {"line1": [["grass","",""]]}}`,
		"unknown key":        `{"mapName": "m", "mapNo": 1, "player": "p", "extra": 1, "tiles": {"line1": [["grass","",""]]}}`,
		"trailing data":      header + `"line1": [["grass","",""]]}} {}`,
	}
	for name, in := range cases {
		if _, err := Parse([]byte(in)); err == nil {
			t.Errorf("%s: Parse accepted %s", name, in)
		}
	}
}

func TestMarshalRejectsUnrepresentable(t *testing.T) {
	for _, s := range []string{`a"b`, "a'b", `a\b`, "a\nb"} {
		l := &Level{MapName: "m", Player: "p", Lines: [][]Hex{{{Terrain: "grass", Unit: s}}}}
		if _, err := l.Marshal(); err == nil {
			t.Errorf("Marshal accepted unit %q", s)
		}
	}
}

func TestIDsKnown(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	unit := func(path string, id string) {
		if _, ok := set.Units[id]; !ok {
			t.Errorf("%s: unknown unit %q", path, id)
		}
	}
	for _, path := range shippedLevels(t) {
		l, err := Load(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range l.Lines {
			for _, h := range line {
				if _, ok := set.Terrain[h.Terrain]; !ok {
					t.Errorf("%s: unknown terrain %q", path, h.Terrain)
				}
				if h.Unit != "" {
					unit(path, h.Unit)
				}
				if h.Content != nil {
					for _, u := range h.Content.Units {
						unit(path, u)
					}
				}
			}
		}
	}
}
