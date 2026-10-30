package scenario

import (
	"bytes"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// levelsDir is the levels directory on disk.
var levelsDir = filepath.Join("..", "..", "levels")

// shipped are the shipped level folders' paths on disk.
func shipped(t *testing.T) []string {
	t.Helper()
	dirs, err := Folders(os.DirFS(levelsDir))
	if err != nil {
		t.Fatal(err)
	}
	if len(dirs) < 51 {
		t.Fatalf("%d level folders, want at least 51", len(dirs))
	}
	for i, d := range dirs {
		dirs[i] = filepath.Join(levelsDir, filepath.FromSlash(d))
	}
	return dirs
}

// Every shipped folder reads, encodes to its own files and holds no other
// JSON than them and a month's front page, and every setup is a level the
// rules start.
func TestShippedRoundTrip(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range shipped(t) {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			f, err := Load(dir)
			if err != nil {
				t.Fatal(err)
			}
			files, err := f.Files()
			if err != nil {
				t.Fatal(err)
			}
			for name, b := range files {
				raw, err := os.ReadFile(filepath.Join(dir, name))
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(b, raw) {
					t.Errorf("%s: encoding differs from the file", name)
				}
			}
			onDisk, err := filepath.Glob(filepath.Join(dir, "*.json"))
			if err != nil {
				t.Fatal(err)
			}
			for _, p := range onDisk {
				if name := filepath.Base(p); files[name] == nil && (name != "news.json" || f.Month == "") {
					t.Errorf("%s is no file of the level", name)
				}
			}
			if (f.Month == "") != (filepath.Base(filepath.Dir(dir)) == ExtraDir) {
				t.Errorf("month %q in %s", f.Month, dir)
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

// The embedded folders are the folders on disk, and the embedded files
// leave the backdrops' sources out.
func TestShippedFolders(t *testing.T) {
	disk, err := Folders(os.DirFS(levelsDir))
	if err != nil {
		t.Fatal(err)
	}
	emb, err := Folders(levels.Levels)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(disk, emb) {
		t.Errorf("embedded %v, on disk %v", emb, disk)
	}
	months, err := Months(levels.Levels)
	if err != nil || len(months) == 0 || months[0] != "1914-08" || slices.ContainsFunc(months, func(d string) bool { return strings.HasPrefix(d, ExtraDir) }) {
		t.Errorf("months %v, %v", months, err)
	}
	if src, _ := fs.Glob(levels.Levels, "*/"+BackdropSource); len(src) > 0 {
		t.Errorf("embedded backdrop sources %v", src)
	}
	all, err := Shipped(CP)
	if err != nil || len(all) != len(disk) || all[0].Name != "1914-08:cp" || all[0].Level.Extra.Backdrop != BackdropFile {
		t.Fatalf("shipped %d for %d folders, %v", len(all), len(disk), err)
	}
}

var small = map[string]string{
	MapFile: `{
	"format": "historyline-scenario",
	"version": 2,
	"name": "Test",
	"map": {
		"line1": [["grass", ""], ["hqN", ""], ["forest", ["stream25", "bridge36"]]],
		"line2": [["grass", "road25"], ["hqC", ""]],
		"line3": [["grass", ""], ["grass", ""], ["grass", ""]]
	}
}
`,
	"cp.json": `{
	"code": "TESTA",
	"units": [
		[1, 1, "CP_infantry"],
		[3, 3, "EC_infantry"]
	],
	"buildings": [
		{"at": [1, 2], "owner": "cp", "credits": 5, "units": ["CP_cavalry"]}
	],
	"builds": ["CP_infantry"]
}
`,
	"ec.json": `{"from": "cp", "code": "TESTB"}
`,
}

// smallFS is small as the level folder 1914-08, with a backdrop.
func smallFS(edit func(map[string]string)) fstest.MapFS {
	files := map[string]string{}
	for k, v := range small {
		files[k] = v
	}
	if edit != nil {
		edit(files)
	}
	fsys := fstest.MapFS{"1914-08/" + BackdropFile: {}}
	for name, s := range files {
		fsys[path.Join("1914-08", name)] = &fstest.MapFile{Data: []byte(s)}
	}
	return fsys
}

func TestLevel(t *testing.T) {
	f, err := Read(smallFS(nil), "1914-08")
	if err != nil {
		t.Fatal(err)
	}
	if f.Month != "1914-08" || f.Backdrop != BackdropFile {
		t.Errorf("month %q, backdrop %q", f.Month, f.Backdrop)
	}
	files, err := f.Files()
	if err != nil || len(files) != len(small) {
		t.Fatalf("files: %v, %d", err, len(files))
	}
	for name, s := range small {
		if string(files[name]) != s {
			t.Errorf("%s:\n%s", name, files[name])
		}
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
	if lv.Extra.Credits[hexgrid.Coord{Line: 0, Col: 1}] != 5 || !reflect.DeepEqual(lv.Extra.Builds, []string{"CP_infantry"}) || lv.Extra.Backdrop != BackdropFile {
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
	for name, edit := range map[string][3]string{
		"version":          {MapFile, `"version": 2`, `"version": 1`},
		"unknown key":      {MapFile, `"name": "Test",`, `"name": "Test", "extra": 1,`},
		"line gap":         {MapFile, `"line3"`, `"line4"`},
		"line width":       {MapFile, `["hqC", ""]]`, `["hqC", ""], ["grass", ""]]`},
		"cell":             {MapFile, `["grass", "road25"]`, `["grass", "road25", ""]`},
		"unit off the map": {"cp.json", `[3, 3, "EC_infantry"]`, `[2, 3, "EC_infantry"]`},
		"two units":        {"cp.json", `[3, 3, "EC_infantry"]`, `[1, 1, "EC_infantry"]`},
		"no building":      {"cp.json", `"at": [1, 2]`, `"at": [1, 1]`},
		"owner":            {"cp.json", `"owner": "cp"`, `"owner": "de"`},
		"setup key":        {"cp.json", `"code": "TESTA",`, `"code": "TESTA", "month": "1914-08",`},
		"from missing":     {"ec.json", `"from": "cp"`, `"from": "2p"`},
		"same code":        {"ec.json", `"TESTB"`, `"TESTA"`},
		"from with units":  {"ec.json", `"from": "cp",`, `"from": "cp", "builds": ["CP_infantry"],`},
	} {
		t.Run(name, func(t *testing.T) {
			if !strings.Contains(small[edit[0]], edit[1]) {
				t.Fatalf("%q is not in %s", edit[1], edit[0])
			}
			fsys := smallFS(func(files map[string]string) { files[edit[0]] = strings.Replace(files[edit[0]], edit[1], edit[2], 1) })
			if _, err := Read(fsys, "1914-08"); err == nil {
				t.Error("reads")
			}
		})
	}
	if _, err := Decode(map[string][]byte{MapFile: []byte(small[MapFile]), "xx.json": []byte(small["ec.json"])}); err == nil {
		t.Error("a file of no mode decodes")
	}
	if _, err := Decode(map[string][]byte{"cp.json": []byte(small["cp.json"])}); err == nil {
		t.Error("a folder without its map decodes")
	}
}

func TestLevelUnitOnBuilding(t *testing.T) {
	fsys := smallFS(func(files map[string]string) {
		files["cp.json"] = strings.Replace(files["cp.json"], `[1, 1, "CP_infantry"]`, `[1, 2, "CP_infantry"]`, 1)
	})
	f, err := Read(fsys, "1914-08")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Level(CP); err == nil {
		t.Error("a unit on a building's content is a level")
	}
}

// Write writes the files and removes those of the setups the file lacks.
func TestWrite(t *testing.T) {
	f, err := Read(smallFS(nil), "1914-08")
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "1914-08")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, SetupFile(TwoPlayer)), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := f.Write(dir); err != nil {
		t.Fatal(err)
	}
	g, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if g.Month != "1914-08" || g.Backdrop != "" || g.Setups[TwoPlayer] != nil || !reflect.DeepEqual(g.Setup(EC), f.Setup(EC)) {
		t.Errorf("read back %+v", g)
	}
}
