// Package scenario reads and writes the level folders below levels/: one
// map shared by every mode and a unit setup per mode (the Central Powers
// against the computer, the Entente against the computer, two players),
// each in its own file.
//
//	levels/1914-08/map.json
//	{
//		"format": "historyline-scenario",
//		"version": 3,
//		"name": "Pulse",
//		"season": "summer",
//		"map": {
//			"line1": [["hills", ""], ["grass", "road236"], …],
//			…
//		},
//		"seasons": {"line7": "....aaa....", …}
//	}
//
//	levels/1914-08/cp.json
//	{
//		"code": "PULSE",
//		"units": [[4, 1, "CP_lArtillery"], …],
//		"buildings": [{"at": [11, 1], "owner": "cp", "credits": 5, "units": ["CP_infantry"]}, …],
//		"builds": ["CP_infantry", …]
//	}
//
//	levels/1914-08/2p.json
//	{"from": "cp", "code": "…"}
//
// The map holds the hexes of the legacy level format (internal/level)
// without units and content: [terrain, infrastructure], line1… as there,
// the terrain a kind without a season ("grass" and "hills", never
// "grass_w" or "rough_w"). The map's "season" ("summer", "autumn",
// "spring" or "winter") sets the rules season (winter plays by the
// original's winter tables) and the look of every hex; in winter the level
// holds each terrain's winter form (data.InSeason: "grass_w", hills
// "rough_w"). The optional "seasons" changes the look of single hexes only:
// a string per line that has any, a character per hex, "." for the map's
// season, "s" summer, "a" autumn, "p" spring, "w" winter
// (mapcompose.LevelHex draws them). Units and building entrances ("at") are [line,
// column], both counted from 1 as the map's line keys. A building's owner
// is "cp", "ec" or absent (none); its credits are per cycle. A setup with
// "from" reuses that setup's units, buildings and builds under its own
// code. A folder is named by its campaign month. Its backdrop.jpg is a
// painting of the whole map's ground, drawn in place of the terrain tiles;
// backdrop.png beside it is its lossless source, which the game does not
// ship.
package scenario

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// Format and Version name the file format.
const (
	Format  = "historyline-scenario"
	Version = 3
)

// The files of a level folder besides the setups (SetupFile).
const (
	MapFile        = "map.json"
	BackdropFile   = "backdrop.jpg"
	BackdropSource = "backdrop.png"
)

// Mode is a setup's key: who plays which side.
type Mode string

// The modes: a human Central Powers or Entente against the computer, or
// two human players.
const (
	CP        Mode = "cp"
	EC        Mode = "ec"
	TwoPlayer Mode = "2p"
)

// Modes are the modes in file order.
var Modes = []Mode{CP, EC, TwoPlayer}

// Human is the side a human plays first in the mode: the Entente in EC,
// the Central Powers otherwise (they move first with two players).
func (m Mode) Human() game.Player {
	if m == EC {
		return game.Entente
	}
	return game.CentralPowers
}

// SetupFile is the file of the mode's setup in a level folder: "cp.json".
func SetupFile(m Mode) string { return string(m) + ".json" }

// File is a level folder.
type File struct {
	Month string // "1914-08", the folder's name
	Name  string // "Pulse"
	// Season is the map's season, "summer", "autumn", "spring" or
	// "winter": the rules season and the look of every hex Seasons leaves.
	Season string
	// Seasons are the hexes drawn in another season than Season, by hex.
	Seasons map[hexgrid.Coord]string
	// Backdrop is the painting of the map's ground drawn in place of the
	// terrain tiles: BackdropFile when the folder has one, covering the
	// map's world rectangle as maprender draws it; "" for none.
	Backdrop string
	// Map[line][col] is terrain and infrastructure only, the terrain in
	// the map's season (data.InSeason): Seasons changes the look only.
	Map    [][]level.Hex
	Setups map[Mode]*Setup
}

// Setup is the units and buildings of one mode.
type Setup struct {
	From      Mode   // the setup whose units, buildings and builds this one reuses; "" for none
	Code      string // the code word that opens the setup; "" for none
	Units     []Unit
	Buildings []Building
	Builds    []string // the unit ids factories build
	// ComputerBuilds are the unit types (numbers) the computer player may build
	// on this map; nil for the default of its side (internal/dosai).
	ComputerBuilds []int
}

// Unit is a unit on the map.
type Unit struct {
	At hexgrid.Coord
	ID string
}

// Building is the content of a building, at its entrance.
type Building struct {
	At      hexgrid.Coord
	Owner   game.Player
	Credits int
	Units   []string
}

// Read reads and checks the level folder dir of fsys (slash-separated): its
// map, the setups it has files for and whether it has a backdrop. The
// month is dir's name where that is one.
func Read(fsys fs.FS, dir string) (*File, error) {
	files := map[string][]byte{}
	for _, name := range append([]string{MapFile}, setupFiles()...) {
		b, err := fs.ReadFile(fsys, path.Join(dir, name))
		if name != MapFile && errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		files[name] = b
	}
	f, err := Decode(files)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", dir, err)
	}
	if base := path.Base(dir); monthRE.MatchString(base) {
		f.Month = base
	}
	if _, err := fs.Stat(fsys, path.Join(dir, BackdropFile)); err == nil {
		f.Backdrop = BackdropFile
	}
	return f, nil
}

// Load reads and checks the level folder dir from disk.
func Load(dir string) (*File, error) {
	dir = filepath.Clean(dir)
	return Read(os.DirFS(filepath.Dir(dir)), filepath.Base(dir))
}

// LoadLevel is the level of a setup of a level folder on disk.
func LoadLevel(dir string, m Mode) (*level.Level, error) {
	f, err := Load(dir)
	if err != nil {
		return nil, err
	}
	lv, err := f.Level(m)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", dir, err)
	}
	return lv, nil
}

func setupFiles() []string {
	out := make([]string, len(Modes))
	for i, m := range Modes {
		out[i] = SetupFile(m)
	}
	return out
}

type mapJSON struct {
	Format  string                         `json:"format"`
	Version int                            `json:"version"`
	Name    string                         `json:"name"`
	Season  string                         `json:"season"`
	Map     map[string][][]json.RawMessage `json:"map"`
	Seasons map[string]string              `json:"seasons"`
}

type setupJSON struct {
	From           Mode                `json:"from"`
	Code           string              `json:"code"`
	Units          [][]json.RawMessage `json:"units"`
	Buildings      []buildingJSON      `json:"buildings"`
	Builds         []string            `json:"builds"`
	ComputerBuilds []int               `json:"computerBuilds"`
}

type buildingJSON struct {
	At      [2]int   `json:"at"`
	Owner   string   `json:"owner"`
	Credits int      `json:"credits"`
	Units   []string `json:"units"`
}

var (
	monthRE = regexp.MustCompile(`^(19\d\d)-(0[1-9]|1[0-2])$`)
	owners  = map[string]game.Player{"": game.None, "cp": game.CentralPowers, "ec": game.Entente}
)

// strict decodes the single JSON object raw into v, unknown keys an error.
func strict(raw []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return errors.New("data after the object")
	}
	return nil
}

// Decode checks and decodes the files of a level folder by name (MapFile
// and a SetupFile per setup): unknown files or keys, gaps in the map lines,
// hexes off the map, two units on a hex, a building entrance that is no
// building, an unknown owner, a "from" chain, a code used twice, a missing or
// unknown season, a "seasons" line that does not fit its map line, a terrain
// with a season suffix and a terrain the terrain table lacks in its hex's
// season are errors. The month and the backdrop are the folder's, not its
// files'.
func Decode(files map[string][]byte) (*File, error) {
	var mj mapJSON
	raw, ok := files[MapFile]
	if !ok {
		return nil, errors.New("no " + MapFile)
	}
	if err := strict(raw, &mj); err != nil {
		return nil, fmt.Errorf("%s: %w", MapFile, err)
	}
	if mj.Format != Format || mj.Version != Version {
		return nil, fmt.Errorf("%s: format %q version %d, want %q version %d", MapFile, mj.Format, mj.Version, Format, Version)
	}
	setups := map[Mode]*setupJSON{}
	for name, raw := range files {
		if name == MapFile {
			continue
		}
		m := Mode(strings.TrimSuffix(name, ".json"))
		if !slices.Contains(Modes, m) || name != SetupFile(m) {
			return nil, fmt.Errorf("%s: no map or setup file", name)
		}
		var sj setupJSON
		if err := strict(raw, &sj); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		setups[m] = &sj
	}
	f, err := build(mj.Name, mj.Map, setups)
	if err != nil {
		return nil, err
	}
	if err := f.setSeasons(mj.Season, mj.Seasons); err != nil {
		return nil, fmt.Errorf("%s: %w", MapFile, err)
	}
	return f, nil
}

// seasonNames are the seasons of a map and its hexes.
var seasonNames = []string{"summer", "autumn", "spring", "winter"}

// seasonChars are the characters of a "seasons" line, by season.
var seasonChars = map[byte]string{'s': "summer", 'a': "autumn", 'p': "spring", 'w': "winter"}

// MonthSeason is the season of campaign month "1915-04", the default of a
// new map: spring from March to May, autumn from September to November,
// summer otherwise. Winter maps say so, since winter changes the rules.
func MonthSeason(month string) string {
	if m := monthRE.FindStringSubmatch(month); m != nil {
		switch m[2] {
		case "03", "04", "05":
			return "spring"
		case "09", "10", "11":
			return "autumn"
		}
	}
	return "summer"
}

// setSeasons sets the map's season and the hexes' seasons from map.json
// (SetSeasons). The terrain is a kind (data.Kind): "hills", never "hills_w"
// or "rough_w".
func (f *File) setSeasons(season string, lines map[string]string) error {
	for i, line := range f.Map {
		for col, h := range line {
			kind := data.Kind(h.Terrain)
			if !data.HasTerrain(kind) && data.HasTerrain(h.Terrain+"_w") {
				kind = data.Kind(h.Terrain + "_w") // "rough" names the winter form of "hills"
			}
			switch {
			case !data.HasTerrain(kind):
				return fmt.Errorf("map: line%d[%d]: unknown terrain %q", i+1, col, h.Terrain)
			case kind != h.Terrain:
				return fmt.Errorf("map: line%d[%d]: terrain %q: write %q, the season comes from season and seasons", i+1, col, h.Terrain, kind)
			}
		}
	}
	var seasons map[hexgrid.Coord]string
	for key, line := range lines {
		i, err := strconv.Atoi(strings.TrimPrefix(key, "line"))
		if err != nil || !strings.HasPrefix(key, "line") || i < 1 || i > len(f.Map) || key != "line"+strconv.Itoa(i) {
			return fmt.Errorf("seasons: %s is no map line", key)
		}
		if len(line) != len(f.Map[i-1]) {
			return fmt.Errorf("seasons: %s has %d hexes, the map line %d", key, len(line), len(f.Map[i-1]))
		}
		for col := range len(line) {
			s, ok := seasonChars[line[col]]
			switch {
			case line[col] == '.':
			case !ok:
				return fmt.Errorf("seasons: %s[%d]: %q, want ., s, a, p or w", key, col, line[col])
			default:
				if seasons == nil {
					seasons = map[hexgrid.Coord]string{}
				}
				seasons[hexgrid.Coord{Line: i - 1, Col: col}] = s
			}
		}
	}
	return f.SetSeasons(season, seasons)
}

// SetSeasons sets the map's season and the hexes drawn in another one
// (entries in the map's season are dropped) and puts every hex's terrain in
// the map's season (data.InSeason): the seasons of single hexes change their
// look only. An unknown season and a terrain the terrain table lacks in the
// map's season or its hex's are errors.
func (f *File) SetSeasons(season string, seasons map[hexgrid.Coord]string) error {
	if !slices.Contains(seasonNames, season) {
		return fmt.Errorf("season %q, want summer, autumn, spring or winter", season)
	}
	f.Season, f.Seasons = season, nil
	for c, s := range seasons {
		switch {
		case !slices.Contains(seasonNames, s):
			return fmt.Errorf("season %q at [%d, %d]", s, c.Line+1, c.Col+1)
		case !f.Grid().Contains(c):
			return fmt.Errorf("season %s at [%d, %d]: off the map", s, c.Line+1, c.Col+1)
		case s != season:
			if f.Seasons == nil {
				f.Seasons = map[hexgrid.Coord]string{}
			}
			f.Seasons[c] = s
		}
	}
	for i, line := range f.Map {
		for col := range line {
			h := &line[col]
			kind := data.Kind(h.Terrain)
			h.Terrain = data.InSeason(kind, season)
			for _, s := range []string{season, f.HexSeason(hexgrid.Coord{Line: i, Col: col})} {
				if !data.HasTerrain(data.InSeason(kind, s)) {
					return fmt.Errorf("map: line%d[%d]: no terrain %q in %s", i+1, col, kind, s)
				}
			}
		}
	}
	return nil
}

// HexSeason is the season hex c is drawn in: its entry in Seasons, else
// Season.
func (f *File) HexSeason(c hexgrid.Coord) string {
	if s, ok := f.Seasons[c]; ok {
		return s
	}
	return f.Season
}

// build checks and makes the file of a map and its setups.
func build(name string, lines map[string][][]json.RawMessage, setups map[Mode]*setupJSON) (*File, error) {
	if name == "" {
		return nil, errors.New("no name")
	}
	f := &File{Name: name, Setups: map[Mode]*Setup{}}
	if err := f.parseMap(lines); err != nil {
		return nil, err
	}
	if len(setups) == 0 {
		return nil, errors.New("no setups")
	}
	for m, sj := range setups {
		if !slices.Contains(Modes, m) {
			return nil, fmt.Errorf("setup %q: no such mode", m)
		}
		if sj == nil {
			return nil, fmt.Errorf("setup %s: null", m)
		}
		s, err := f.parseSetup(sj)
		if err != nil {
			return nil, fmt.Errorf("setup %s: %w", m, err)
		}
		f.Setups[m] = s
	}
	codes := map[string]Mode{}
	for _, m := range Modes {
		s := f.Setups[m]
		if s == nil {
			continue
		}
		if s.From != "" {
			src := f.Setups[s.From]
			switch {
			case src == nil:
				return nil, fmt.Errorf("setup %s: from %q, which the file lacks", m, s.From)
			case src.From != "":
				return nil, fmt.Errorf("setup %s: from %s, which is itself from %s", m, s.From, src.From)
			}
		}
		if s.Code == "" {
			continue
		}
		if o, dup := codes[s.Code]; dup {
			return nil, fmt.Errorf("setups %s and %s share the code %q", o, m, s.Code)
		}
		codes[s.Code] = m
	}
	return f, nil
}

func (f *File) parseMap(lines map[string][][]json.RawMessage) error {
	if len(lines) == 0 {
		return errors.New("no map")
	}
	f.Map = make([][]level.Hex, len(lines))
	for i := range f.Map {
		key := "line" + strconv.Itoa(i+1)
		line, ok := lines[key]
		if !ok {
			return fmt.Errorf("map: %s missing", key)
		}
		f.Map[i] = make([]level.Hex, len(line))
		for col, elems := range line {
			if len(elems) != 2 {
				return fmt.Errorf("map: %s[%d]: %d elements, want [terrain, infrastructure]", key, col, len(elems))
			}
			h := &f.Map[i][col]
			if err := json.Unmarshal(elems[0], &h.Terrain); err != nil || h.Terrain == "" {
				return fmt.Errorf("map: %s[%d]: terrain %s", key, col, elems[0])
			}
			infra, err := level.ParseInfra(elems[1])
			if err != nil {
				return fmt.Errorf("map: %s[%d]: infra: %w", key, col, err)
			}
			h.Infra = infra
		}
	}
	g := f.Grid()
	for i, line := range f.Map {
		if len(line) == 0 || !g.Contains(hexgrid.Coord{Line: i, Col: len(line) - 1}) ||
			g.Contains(hexgrid.Coord{Line: i, Col: len(line)}) {
			return fmt.Errorf("map: line%d has %d hexes, which does not fit width %d", i+1, len(line), g.Width)
		}
	}
	return nil
}

// at is the map hex of a file position [line, column], counted from 1.
func (f *File) at(line, col int) (hexgrid.Coord, error) {
	c := hexgrid.Coord{Line: line - 1, Col: col - 1}
	if !f.Grid().Contains(c) {
		return c, fmt.Errorf("[%d, %d] is off the map", line, col)
	}
	return c, nil
}

func (f *File) parseSetup(sj *setupJSON) (*Setup, error) {
	s := &Setup{From: sj.From, Code: sj.Code, Builds: sj.Builds, ComputerBuilds: sj.ComputerBuilds}
	if s.From != "" {
		if !slices.Contains(Modes, s.From) {
			return nil, fmt.Errorf("from %q: no such mode", s.From)
		}
		if len(sj.Units) > 0 || len(sj.Buildings) > 0 || len(sj.Builds) > 0 || len(sj.ComputerBuilds) > 0 {
			return nil, errors.New("a setup from another has no units, buildings or builds of its own")
		}
		return s, nil
	}
	taken := map[hexgrid.Coord]bool{}
	for _, raw := range sj.Units {
		var line, col int
		var id string
		if len(raw) != 3 || json.Unmarshal(raw[0], &line) != nil || json.Unmarshal(raw[1], &col) != nil ||
			json.Unmarshal(raw[2], &id) != nil || id == "" {
			b, err := json.Marshal(raw)
			if err != nil {
				return nil, fmt.Errorf("invalid unit: %w", err)
			}
			return nil, fmt.Errorf("unit %s: want [line, column, id]", b)
		}
		c, err := f.at(line, col)
		if err != nil {
			return nil, fmt.Errorf("unit %s: %w", id, err)
		}
		if taken[c] {
			return nil, fmt.Errorf("unit %s: a second unit at [%d, %d]", id, line, col)
		}
		taken[c] = true
		s.Units = append(s.Units, Unit{At: c, ID: id})
	}
	seen := map[hexgrid.Coord]bool{}
	for _, bj := range sj.Buildings {
		c, err := f.at(bj.At[0], bj.At[1])
		if err != nil {
			return nil, fmt.Errorf("building: %w", err)
		}
		if seen[c] {
			return nil, fmt.Errorf("building: a second entry at %v", bj.At)
		}
		seen[c] = true
		if !building(f.Map[c.Line][c.Col].Terrain) {
			return nil, fmt.Errorf("building at %v: %q is no building", bj.At, f.Map[c.Line][c.Col].Terrain)
		}
		o, ok := owners[bj.Owner]
		if !ok {
			return nil, fmt.Errorf("building at %v: owner %q, want cp, ec or none", bj.At, bj.Owner)
		}
		s.Buildings = append(s.Buildings, Building{At: c, Owner: o, Credits: bj.Credits, Units: bj.Units})
	}
	return s, nil
}

// building reports whether terrain is a piece of an HQ, depot or factory.
func building(terrain string) bool {
	t := strings.TrimSuffix(terrain, "_w")
	return strings.HasPrefix(t, "hq") || strings.HasPrefix(t, "cmp") || strings.HasPrefix(t, "factory")
}

// Grid is the hex layout of the map: Width is the length of line1.
func (f *File) Grid() hexgrid.Grid {
	if len(f.Map) == 0 {
		return hexgrid.Grid{}
	}
	return hexgrid.Grid{Width: len(f.Map[0]), Lines: len(f.Map)}
}

// Setup is the setup of a mode with "from" resolved; nil when the file has
// none.
func (f *File) Setup(m Mode) *Setup {
	s := f.Setups[m]
	if s == nil || s.From == "" {
		return s
	}
	src := *f.Setups[s.From]
	src.From, src.Code = s.From, s.Code
	return &src
}

// Winter reports whether the map plays by the winter rules.
func (f *File) Winter() bool { return f.Season == "winter" }

// Level is a setup as a level of the legacy format: the map with the
// setup's units and building content, Player the mode's first human side,
// MapName the setup's code word (the file's name without one: the game info
// and the best scores go by it), MapNo 0, and Extra the setup's credits and
// builds and the file's backdrop, Season and Seasons. A unit on a
// building's content hex is an error.
func (f *File) Level(m Mode) (*level.Level, error) {
	s := f.Setup(m)
	if s == nil {
		return nil, fmt.Errorf("no %s setup", m)
	}
	name := f.Name
	if s.Code != "" {
		name = s.Code
	}
	lv := &level.Level{
		MapName: name, Player: m.Human().String(), Lines: make([][]level.Hex, len(f.Map)),
		Extra: &level.Extra{
			Credits: map[hexgrid.Coord]int{}, Builds: slices.Clone(s.Builds), ComputerBuilds: slices.Clone(s.ComputerBuilds), Backdrop: f.Backdrop,
			Season: f.Season, Seasons: maps.Clone(f.Seasons), Month: f.Month,
		},
	}
	for i, line := range f.Map {
		lv.Lines[i] = make([]level.Hex, len(line))
		for col, h := range line {
			h.Infra.IDs = slices.Clone(h.Infra.IDs)
			lv.Lines[i][col] = h
		}
	}
	for _, b := range s.Buildings {
		lv.Lines[b.At.Line][b.At.Col].Content = &level.Content{Owner: int(b.Owner), Units: slices.Clone(b.Units)}
		if b.Credits != 0 {
			lv.Extra.Credits[b.At] = b.Credits
		}
	}
	for _, u := range s.Units {
		h := &lv.Lines[u.At.Line][u.At.Col]
		if h.Content != nil {
			return nil, fmt.Errorf("setup %s: unit %s on the building at [%d, %d]", m, u.ID, u.At.Line+1, u.At.Col+1)
		}
		h.Unit = u.ID
	}
	return lv, nil
}

// Files encodes the folder's files by name: MapFile, its season, a map line
// per row with the terrain without its season and a "seasons" line per row
// with hexes in another season, and a SetupFile per setup in Modes order, a
// unit and a building per row in hex order. A terrain that is not in the
// map's season (data.InSeason) is an error.
func (f *File) Files() (map[string][]byte, error) {
	var b bytes.Buffer
	str := func(s string) {
		q, _ := json.Marshal(s) //nolint:errcheck // A string always has a JSON representation.
		b.Write(q)
	}
	strs := func(ss []string) {
		b.WriteString("[")
		for i, s := range ss {
			if i > 0 {
				b.WriteString(", ")
			}
			str(s)
		}
		b.WriteString("]")
	}
	out := map[string][]byte{}
	fmt.Fprintf(&b, "{\n\t\"format\": %q,\n\t\"version\": %d,\n\t\"name\": ", Format, Version)
	str(f.Name)
	b.WriteString(",\n\t\"season\": ")
	str(f.Season)
	b.WriteString(",\n\t\"map\": {\n")
	var seasons []string
	for i, line := range f.Map {
		fmt.Fprintf(&b, "\t\t\"line%d\": [", i+1)
		chars, mixed := make([]byte, len(line)), false
		for col, h := range line {
			c := hexgrid.Coord{Line: i, Col: col}
			chars[col] = '.'
			if s, ok := f.Seasons[c]; ok && s != f.Season {
				chars[col], mixed = s[0], true
				if s == "spring" {
					chars[col] = 'p'
				}
			}
			kind := data.Kind(h.Terrain)
			if data.InSeason(kind, f.Season) != h.Terrain {
				return nil, fmt.Errorf("scenario: line%d[%d]: terrain %q on a %s map", i+1, col, h.Terrain, f.Season)
			}
			if col > 0 {
				b.WriteString(", ")
			}
			b.WriteString("[")
			str(kind)
			b.WriteString(", ")
			switch {
			case h.Infra.List || len(h.Infra.IDs) > 1:
				strs(h.Infra.IDs)
			case len(h.Infra.IDs) == 1:
				str(h.Infra.IDs[0])
			default:
				str("")
			}
			b.WriteString("]")
		}
		b.WriteString("]")
		if i+1 < len(f.Map) {
			b.WriteString(",")
		}
		b.WriteString("\n")
		if mixed {
			seasons = append(seasons, fmt.Sprintf("\t\t\"line%d\": %q", i+1, chars))
		}
	}
	b.WriteString("\t}")
	if len(seasons) > 0 {
		fmt.Fprintf(&b, ",\n\t\"seasons\": {\n%s\n\t}", strings.Join(seasons, ",\n"))
	}
	b.WriteString("\n}\n")
	out[MapFile] = slices.Clone(b.Bytes())
	for _, m := range Modes {
		s := f.Setups[m]
		if s == nil {
			continue
		}
		b.Reset()
		if s.From != "" {
			fmt.Fprintf(&b, "{\"from\": %q", s.From)
			if s.Code != "" {
				b.WriteString(", \"code\": ")
				str(s.Code)
			}
			b.WriteString("}\n")
			out[SetupFile(m)] = slices.Clone(b.Bytes())
			continue
		}
		var fields []func()
		if s.Code != "" {
			fields = append(fields, func() { b.WriteString("\"code\": "); str(s.Code) })
		}
		units := slices.Clone(s.Units)
		slices.SortStableFunc(units, func(x, y Unit) int { return order(x.At, y.At) })
		fields = append(fields, func() {
			b.WriteString("\"units\": [")
			for k, u := range units {
				if k > 0 {
					b.WriteString(",")
				}
				fmt.Fprintf(&b, "\n\t\t[%d, %d, ", u.At.Line+1, u.At.Col+1)
				str(u.ID)
				b.WriteString("]")
			}
			if len(units) > 0 {
				b.WriteString("\n\t")
			}
			b.WriteString("]")
		})
		bs := slices.Clone(s.Buildings)
		slices.SortStableFunc(bs, func(x, y Building) int { return order(x.At, y.At) })
		fields = append(fields, func() {
			b.WriteString("\"buildings\": [")
			for k, bd := range bs {
				if k > 0 {
					b.WriteString(",")
				}
				fmt.Fprintf(&b, "\n\t\t{\"at\": [%d, %d]", bd.At.Line+1, bd.At.Col+1)
				switch bd.Owner {
				case game.CentralPowers:
					b.WriteString(", \"owner\": \"cp\"")
				case game.Entente:
					b.WriteString(", \"owner\": \"ec\"")
				}
				if bd.Credits != 0 {
					fmt.Fprintf(&b, ", \"credits\": %d", bd.Credits)
				}
				if len(bd.Units) > 0 {
					b.WriteString(", \"units\": ")
					strs(bd.Units)
				}
				b.WriteString("}")
			}
			if len(bs) > 0 {
				b.WriteString("\n\t")
			}
			b.WriteString("]")
		})
		if len(s.Builds) > 0 {
			fields = append(fields, func() { b.WriteString("\"builds\": "); strs(s.Builds) })
		}
		if len(s.ComputerBuilds) > 0 {
			fields = append(fields, func() {
				b.WriteString("\"computerBuilds\": ")
				b.WriteString("[")
				for i, t := range s.ComputerBuilds {
					if i > 0 {
						b.WriteString(",")
					}
					b.WriteString(strconv.Itoa(t))
				}
				b.WriteString("]")
			})
		}
		b.WriteString("{")
		for k, w := range fields {
			if k > 0 {
				b.WriteString(",")
			}
			b.WriteString("\n\t")
			w()
		}
		b.WriteString("\n}\n")
		out[SetupFile(m)] = slices.Clone(b.Bytes())
	}
	if _, err := Decode(out); err != nil {
		return nil, fmt.Errorf("scenario: the encoded files do not decode: %w", err)
	}
	return out, nil
}

// Write writes the folder's files into dir, which it makes where missing,
// and removes the files of the setups f lacks.
func (f *File) Write(dir string) error {
	files, err := f.Files()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for name, b := range files {
		if err := os.WriteFile(filepath.Join(dir, name), b, 0o644); err != nil {
			return err
		}
	}
	for _, m := range Modes {
		if f.Setups[m] != nil {
			continue
		}
		if err := os.Remove(filepath.Join(dir, SetupFile(m))); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	return nil
}

// order sorts hexes by line, then column.
func order(a, b hexgrid.Coord) int {
	if a.Line != b.Line {
		return a.Line - b.Line
	}
	return a.Col - b.Col
}

// Split is the reverse of Level: the map of a legacy level and its units
// and building content as a setup (credits and builds from lv.Extra, when
// it has them).
func Split(lv *level.Level) (lines [][]level.Hex, s *Setup) {
	s = &Setup{}
	lines = make([][]level.Hex, len(lv.Lines))
	for i, line := range lv.Lines {
		lines[i] = make([]level.Hex, len(line))
		for col, h := range line {
			c := hexgrid.Coord{Line: i, Col: col}
			if h.Unit != "" {
				s.Units = append(s.Units, Unit{At: c, ID: h.Unit})
			}
			if h.Content != nil {
				bd := Building{At: c, Owner: game.Player(h.Content.Owner), Units: slices.Clone(h.Content.Units)}
				if lv.Extra != nil {
					bd.Credits = lv.Extra.Credits[c]
				}
				s.Buildings = append(s.Buildings, bd)
			}
			lines[i][col] = level.Hex{Terrain: h.Terrain, Infra: level.Infra{IDs: slices.Clone(h.Infra.IDs), List: h.Infra.List}}
		}
	}
	if lv.Extra != nil {
		s.Builds = slices.Clone(lv.Extra.Builds)
		s.ComputerBuilds = slices.Clone(lv.Extra.ComputerBuilds)
	}
	return lines, s
}
