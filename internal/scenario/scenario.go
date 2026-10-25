// Package scenario reads and writes the level files of levels/monthly and
// levels/skirmish: one map shared by every mode and a unit setup per mode
// (the Central Powers against the computer, the Entente against the
// computer, two players).
//
//	{
//		"format": "historyline-scenario",
//		"version": 1,
//		"month": "1914-08",
//		"name": "Pulse",
//		"backdrop": "1914-08.jpg",
//		"map": {
//			"line1": [["hills", ""], ["grass", "road236"], …],
//			…
//		},
//		"setups": {
//			"cp": {
//				"code": "PULSE",
//				"units": [[4, 1, "CP_lArtillery"], …],
//				"buildings": [{"at": [11, 1], "owner": "cp", "credits": 5, "units": ["CP_infantry"]}, …],
//				"builds": ["CP_infantry", …]
//			},
//			"ec": {…},
//			"2p": {"from": "cp", "code": "…"}
//		}
//	}
//
// The map holds the hexes of the legacy level format (internal/level)
// without units and content: [terrain, infrastructure], line1… as there.
// Units and building entrances ("at") are [line, column], both counted from
// 1 as the map's line keys. A building's owner is "cp", "ec" or absent
// (none); its credits are per cycle. A setup with "from" reuses that
// setup's units, buildings and builds under its own code. Skirmish files
// have no month. The optional backdrop names a painting of the whole map's
// ground next to the file, drawn in place of the terrain tiles.
package scenario

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// Format and Version name the file format.
const (
	Format  = "historyline-scenario"
	Version = 1
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

// File is a scenario file.
type File struct {
	Month string // "1914-08"; "" for a skirmish map
	Name  string // "Pulse"
	// Backdrop is the painting of the map's ground drawn in place of the
	// terrain tiles: a PNG or JPEG file next to the scenario file
	// ("1914-08.jpg") covering the map's world rectangle as
	// maprender draws it; "" for none.
	Backdrop string
	Map      [][]level.Hex // Map[line][col]: terrain and infrastructure only
	Setups   map[Mode]*Setup
}

// Setup is the units and buildings of one mode.
type Setup struct {
	From      Mode   // the setup whose units, buildings and builds this one reuses; "" for none
	Code      string // the code word that opens the setup; "" for none
	Units     []Unit
	Buildings []Building
	Builds    []string // the unit ids factories build
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

// Load reads and parses a scenario file.
func Load(path string) (*File, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f, err := Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return f, nil
}

// LoadLevel is the level of a setup of a scenario file.
func LoadLevel(path string, m Mode) (*level.Level, error) {
	f, err := Load(path)
	if err != nil {
		return nil, err
	}
	lv, err := f.Level(m)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return lv, nil
}

type fileJSON struct {
	Format   string                         `json:"format"`
	Version  int                            `json:"version"`
	Month    string                         `json:"month"`
	Name     string                         `json:"name"`
	Backdrop string                         `json:"backdrop"`
	Map      map[string][][]json.RawMessage `json:"map"`
	Setups   map[Mode]*setupJSON            `json:"setups"`
}

type setupJSON struct {
	From      Mode                `json:"from"`
	Code      string              `json:"code"`
	Units     [][]json.RawMessage `json:"units"`
	Buildings []buildingJSON      `json:"buildings"`
	Builds    []string            `json:"builds"`
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

// Parse decodes and checks a scenario file: unknown keys, gaps in the map
// lines, hexes off the map, two units on a hex, a building entrance that is
// no building, an unknown owner, a "from" chain and a code used twice are
// errors.
func Parse(raw []byte) (*File, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var fj fileJSON
	if err := dec.Decode(&fj); err != nil {
		return nil, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("data after the scenario object")
	}
	if fj.Format != Format || fj.Version != Version {
		return nil, fmt.Errorf("format %q version %d, want %q version %d", fj.Format, fj.Version, Format, Version)
	}
	if fj.Month != "" && !monthRE.MatchString(fj.Month) {
		return nil, fmt.Errorf("month %q is not YYYY-MM", fj.Month)
	}
	if fj.Name == "" {
		return nil, errors.New("no name")
	}
	if b := fj.Backdrop; b != "" && (path.Base(b) != b || strings.HasPrefix(b, ".") || !slices.Contains([]string{".png", ".jpg", ".jpeg"}, strings.ToLower(path.Ext(b)))) {
		return nil, fmt.Errorf("backdrop %q is not a PNG or JPEG file name", b)
	}
	f := &File{Month: fj.Month, Name: fj.Name, Backdrop: fj.Backdrop, Setups: map[Mode]*Setup{}}
	if err := f.parseMap(fj.Map); err != nil {
		return nil, err
	}
	if len(fj.Setups) == 0 {
		return nil, errors.New("no setups")
	}
	for m, sj := range fj.Setups {
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
	s := &Setup{From: sj.From, Code: sj.Code, Builds: sj.Builds}
	if s.From != "" {
		if !slices.Contains(Modes, s.From) {
			return nil, fmt.Errorf("from %q: no such mode", s.From)
		}
		if len(sj.Units) > 0 || len(sj.Buildings) > 0 || len(sj.Builds) > 0 {
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
			b, _ := json.Marshal(raw)
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

// Winter reports whether the map carries winter terrain.
func (f *File) Winter() bool {
	for _, line := range f.Map {
		for _, h := range line {
			if strings.HasSuffix(h.Terrain, "_w") {
				return true
			}
		}
	}
	return false
}

// Level is a setup as a level of the legacy format: the map with the
// setup's units and building content, Player the mode's first human side,
// MapName the setup's code word (the file's name without one: the game info
// and the best scores go by it), MapNo 0, and Extra the setup's credits and
// builds and the file's backdrop. A unit on a building's content hex is an
// error.
func (f *File) Level(m Mode) (*level.Level, error) {
	s := f.Setup(m)
	if s == nil {
		return nil, fmt.Errorf("no %s setup", m)
	}
	name := f.Name
	if s.Code != "" {
		name = s.Code
	}
	lv := &level.Level{MapName: name, Player: m.Human().String(), Lines: make([][]level.Hex, len(f.Map)),
		Extra: &level.Extra{Credits: map[hexgrid.Coord]int{}, Builds: slices.Clone(s.Builds), Backdrop: f.Backdrop}}
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

// Marshal encodes the file: a map line per row, a unit and a building per
// row, setups in Modes order, units and buildings in hex order.
func (f *File) Marshal() ([]byte, error) {
	var b bytes.Buffer
	str := func(s string) {
		q, _ := json.Marshal(s)
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
	fmt.Fprintf(&b, "{\n\t\"format\": %q,\n\t\"version\": %d,\n", Format, Version)
	if f.Month != "" {
		b.WriteString("\t\"month\": ")
		str(f.Month)
		b.WriteString(",\n")
	}
	b.WriteString("\t\"name\": ")
	str(f.Name)
	if f.Backdrop != "" {
		b.WriteString(",\n\t\"backdrop\": ")
		str(f.Backdrop)
	}
	b.WriteString(",\n\t\"map\": {\n")
	for i, line := range f.Map {
		fmt.Fprintf(&b, "\t\t\"line%d\": [", i+1)
		for col, h := range line {
			if col > 0 {
				b.WriteString(", ")
			}
			b.WriteString("[")
			str(h.Terrain)
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
	}
	b.WriteString("\t},\n\t\"setups\": {\n")
	var modes []Mode
	for _, m := range Modes {
		if f.Setups[m] != nil {
			modes = append(modes, m)
		}
	}
	for i, m := range modes {
		s := f.Setups[m]
		fmt.Fprintf(&b, "\t\t%q: {", m)
		var fields []func()
		if s.From != "" {
			fields = append(fields, func() { fmt.Fprintf(&b, "\"from\": %q", s.From) })
		}
		if s.Code != "" {
			fields = append(fields, func() { b.WriteString("\"code\": "); str(s.Code) })
		}
		if s.From == "" {
			units := slices.Clone(s.Units)
			slices.SortStableFunc(units, func(x, y Unit) int { return order(x.At, y.At) })
			fields = append(fields, func() {
				b.WriteString("\"units\": [")
				for k, u := range units {
					if k > 0 {
						b.WriteString(",")
					}
					fmt.Fprintf(&b, "\n\t\t\t\t[%d, %d, ", u.At.Line+1, u.At.Col+1)
					str(u.ID)
					b.WriteString("]")
				}
				if len(units) > 0 {
					b.WriteString("\n\t\t\t")
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
					fmt.Fprintf(&b, "\n\t\t\t\t{\"at\": [%d, %d]", bd.At.Line+1, bd.At.Col+1)
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
					b.WriteString("\n\t\t\t")
				}
				b.WriteString("]")
			})
			if len(s.Builds) > 0 {
				fields = append(fields, func() { b.WriteString("\"builds\": "); strs(s.Builds) })
			}
		}
		for k, w := range fields {
			if k > 0 {
				b.WriteString(",")
			}
			if s.From == "" {
				b.WriteString("\n\t\t\t")
			} else if k > 0 {
				b.WriteString(" ")
			}
			w()
		}
		if s.From == "" {
			b.WriteString("\n\t\t")
		}
		b.WriteString("}")
		if i+1 < len(modes) {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString("\t}\n}\n")
	if _, err := Parse(b.Bytes()); err != nil {
		return nil, fmt.Errorf("scenario: the encoded file does not parse: %w", err)
	}
	return b.Bytes(), nil
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
	}
	return lines, s
}
