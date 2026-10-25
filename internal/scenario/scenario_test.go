package scenario

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

func shipped(t *testing.T) []string {
	t.Helper()
	var paths []string
	for _, dir := range []string{"monthly", "skirmish"} {
		p, err := filepath.Glob(filepath.Join("..", "..", "levels", dir, "*.json"))
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, p...)
	}
	if len(paths) < 49 {
		t.Fatalf("%d scenario files, want at least 49", len(paths))
	}
	return paths
}

// Every shipped file parses, encodes to its own bytes, and every setup
// is a level the rules start.
func TestShippedRoundTrip(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range shipped(t) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			f, err := Parse(raw)
			if err != nil {
				t.Fatal(err)
			}
			out, err := f.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(out, raw) {
				t.Error("encoding differs from the file")
			}
			name := strings.TrimSuffix(filepath.Base(path), ".json")
			if f.Month != "" && !strings.HasPrefix(name, f.Month+"-") {
				t.Errorf("month %s, file %s", f.Month, name)
			}
			if f.Backdrop != "" {
				if _, err := os.Stat(filepath.Join(filepath.Dir(path), f.Backdrop)); err != nil {
					t.Errorf("backdrop: %v", err)
				}
			}
			for _, m := range Modes {
				if f.Setups[m] == nil {
					t.Errorf("no %s setup", m)
					continue
				}
				lv, err := f.Level(m)
				if err != nil {
					t.Error(err)
					continue
				}
				if _, err := game.New(lv, set); err != nil {
					t.Errorf("%s: %v", m, err)
				}
			}
		})
	}
}

const small = `{
	"format": "historyline-scenario",
	"version": 1,
	"month": "1914-08",
	"name": "Test",
	"backdrop": "test.jpg",
	"map": {
		"line1": [["grass", ""], ["hqN", ""], ["forest", ["stream25", "bridge36"]]],
		"line2": [["grass", "road25"], ["hqC", ""]],
		"line3": [["grass", ""], ["grass", ""], ["grass", ""]]
	},
	"setups": {
		"cp": {
			"code": "TESTA",
			"units": [
				[1, 1, "CP_infantry"],
				[3, 3, "EC_infantry"]
			],
			"buildings": [
				{"at": [1, 2], "owner": "cp", "credits": 5, "units": ["CP_cavalry"]}
			],
			"builds": ["CP_infantry"]
		},
		"ec": {"from": "cp", "code": "TESTB"}
	}
}
`

func TestLevel(t *testing.T) {
	f, err := Parse([]byte(small))
	if err != nil {
		t.Fatal(err)
	}
	if out, err := f.Marshal(); err != nil || string(out) != small {
		t.Fatalf("marshal: %v\n%s", err, out)
	}
	lv, err := f.Level(EC)
	if err != nil {
		t.Fatal(err)
	}
	if lv.Player != game.Entente.String() || lv.MapName != "TESTB" || lv.MapNo != 0 {
		t.Errorf("level %q %d %q", lv.MapName, lv.MapNo, lv.Player)
	}
	if lv.Lines[0][0].Unit != "CP_infantry" || lv.Lines[2][2].Unit != "EC_infantry" {
		t.Errorf("units %q %q", lv.Lines[0][0].Unit, lv.Lines[2][2].Unit)
	}
	c := lv.Lines[0][1].Content
	if c == nil || c.Owner != int(game.CentralPowers) || !reflect.DeepEqual(c.Units, []string{"CP_cavalry"}) {
		t.Errorf("content %+v", c)
	}
	if lv.Extra.Credits[hexgrid.Coord{Line: 0, Col: 1}] != 5 || !reflect.DeepEqual(lv.Extra.Builds, []string{"CP_infantry"}) || lv.Extra.Backdrop != "test.jpg" {
		t.Errorf("extra %+v", lv.Extra)
	}
	if got := lv.Lines[0][2].Infra; !got.List || len(got.IDs) != 2 {
		t.Errorf("infra %+v", got)
	}
	if s := f.Setup(EC); s.Code != "TESTB" || s.From != CP || len(s.Units) != 2 {
		t.Errorf("setup ec %+v", s)
	}
	if _, err := f.Level(TwoPlayer); err == nil {
		t.Error("a missing setup is a level")
	}

	lines, s := Split(lv)
	if !reflect.DeepEqual(lines, f.Map) {
		t.Error("Split map differs")
	}
	want := f.Setup(CP)
	if !reflect.DeepEqual(s.Units, want.Units) || !reflect.DeepEqual(s.Buildings, want.Buildings) || !reflect.DeepEqual(s.Builds, want.Builds) {
		t.Errorf("Split setup %+v, want %+v", s, want)
	}
}

func TestParseErrors(t *testing.T) {
	for name, edit := range map[string][2]string{
		"version":          {`"version": 1`, `"version": 2`},
		"backdrop path":    {`"test.jpg"`, `"../test.jpg"`},
		"backdrop type":    {`"test.jpg"`, `"test.gif"`},
		"month":            {`"1914-08"`, `"1914-8"`},
		"unknown key":      {`"name": "Test",`, `"name": "Test", "extra": 1,`},
		"line gap":         {`"line3"`, `"line4"`},
		"line width":       {`["hqC", ""]]`, `["hqC", ""], ["grass", ""]]`},
		"cell":             {`["grass", "road25"]`, `["grass", "road25", ""]`},
		"unit off the map": {`[3, 3, "EC_infantry"]`, `[2, 3, "EC_infantry"]`},
		"two units":        {`[3, 3, "EC_infantry"]`, `[1, 1, "EC_infantry"]`},
		"no building":      {`"at": [1, 2]`, `"at": [1, 1]`},
		"owner":            {`"owner": "cp"`, `"owner": "de"`},
		"mode":             {`"ec": {`, `"xx": {`},
		"from missing":     {`"from": "cp"`, `"from": "2p"`},
		"same code":        {`"TESTB"`, `"TESTA"`},
		"from with units":  {`"from": "cp",`, `"from": "cp", "builds": ["CP_infantry"],`},
	} {
		t.Run(name, func(t *testing.T) {
			if !strings.Contains(small, edit[0]) {
				t.Fatalf("%q is not in the file", edit[0])
			}
			if _, err := Parse([]byte(strings.Replace(small, edit[0], edit[1], 1))); err == nil {
				t.Error("parses")
			}
		})
	}
}

func TestLevelUnitOnBuilding(t *testing.T) {
	f, err := Parse([]byte(strings.Replace(small, `[1, 1, "CP_infantry"]`, `[1, 2, "CP_infantry"]`, 1)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Level(CP); err == nil {
		t.Error("a unit on a building's content is a level")
	}
}
