// Package level reads and writes the levels/*.json format of the Python
// reference. Marshal reproduces mapEditor.py saveData byte for byte, so a level
// that is loaded and saved again is unchanged.
package level

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Level is one map. Lines[l][c] is the hex at hexgrid.Coord{Line: l, Col: c}.
type Level struct {
	MapName string
	MapNo   int
	Player  string
	Lines   [][]Hex
}

// Hex is one [terrain, infra, unit] or [terrain, infra, unit, content] entry.
type Hex struct {
	Terrain string
	Infra   Infra
	Unit    string   // "" when the hex is empty
	Content *Content // nil for hexes without the fourth element
}

// Infra lists the infrastructure ids of a hex. The file stores them either as a
// string ("" or "road236") or as a list; List records which form was used.
type Infra struct {
	IDs  []string
	List bool
}

// Content is the [owner] or [owner, [units…]] element of an HQ, depot or
// factory. Units is nil for the [owner] form.
type Content struct {
	Owner int
	Units []string
}

// Grid returns the hex layout of the level: Width is the length of line1.
func (l *Level) Grid() hexgrid.Grid {
	if len(l.Lines) == 0 {
		return hexgrid.Grid{}
	}
	return hexgrid.Grid{Width: len(l.Lines[0]), Lines: len(l.Lines)}
}

type file struct {
	MapName *string                        `json:"mapName"`
	MapNo   *int                           `json:"mapNo"`
	Player  *string                        `json:"player"`
	Tiles   map[string][][]json.RawMessage `json:"tiles"`
}

// Load reads and parses a level file.
func Load(path string) (*Level, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	l, err := Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return l, nil
}

// Parse decodes a level. Unknown keys, missing lines and hexes that do not fit
// the offset layout are errors.
func Parse(raw []byte) (*Level, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var f file
	if err := dec.Decode(&f); err != nil {
		return nil, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("data after the level object")
	}
	if f.MapName == nil || f.MapNo == nil || f.Player == nil {
		return nil, errors.New("mapName, mapNo and player are required")
	}
	if len(f.Tiles) == 0 {
		return nil, errors.New("no tiles")
	}

	l := &Level{MapName: *f.MapName, MapNo: *f.MapNo, Player: *f.Player, Lines: make([][]Hex, len(f.Tiles))}
	for i := range l.Lines {
		key := "line" + strconv.Itoa(i+1)
		line, ok := f.Tiles[key]
		if !ok {
			return nil, fmt.Errorf("tiles: %s missing", key)
		}
		l.Lines[i] = make([]Hex, len(line))
		for col, elems := range line {
			h, err := parseHex(elems)
			if err != nil {
				return nil, fmt.Errorf("%s[%d]: %w", key, col, err)
			}
			l.Lines[i][col] = h
		}
	}

	g := l.Grid()
	for i, line := range l.Lines {
		if len(line) == 0 || !g.Contains(hexgrid.Coord{Line: i, Col: len(line) - 1}) ||
			g.Contains(hexgrid.Coord{Line: i, Col: len(line)}) {
			return nil, fmt.Errorf("line%d has %d hexes, which does not fit width %d", i+1, len(line), g.Width)
		}
	}
	return l, nil
}

func parseHex(elems []json.RawMessage) (Hex, error) {
	if len(elems) != 3 && len(elems) != 4 {
		return Hex{}, fmt.Errorf("%d elements, want 3 or 4", len(elems))
	}
	var h Hex
	var err error
	if h.Terrain, err = parseString(elems[0]); err != nil {
		return Hex{}, fmt.Errorf("terrain: %w", err)
	}
	if h.Infra, err = parseInfra(elems[1]); err != nil {
		return Hex{}, fmt.Errorf("infra: %w", err)
	}
	if h.Unit, err = parseString(elems[2]); err != nil {
		return Hex{}, fmt.Errorf("unit: %w", err)
	}
	if len(elems) == 4 {
		if h.Content, err = parseContent(elems[3]); err != nil {
			return Hex{}, fmt.Errorf("content: %w", err)
		}
	}
	return h, nil
}

func parseString(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || raw[0] != '"' {
		return "", fmt.Errorf("%s is not a string", raw)
	}
	var s string
	err := json.Unmarshal(raw, &s)
	return s, err
}

func parseStrings(raw json.RawMessage) ([]string, error) {
	var elems []json.RawMessage
	if len(raw) == 0 || raw[0] != '[' {
		return nil, fmt.Errorf("%s is not a list", raw)
	}
	if err := json.Unmarshal(raw, &elems); err != nil {
		return nil, err
	}
	out := make([]string, len(elems))
	for i, e := range elems {
		s, err := parseString(e)
		if err != nil {
			return nil, err
		}
		out[i] = s
	}
	return out, nil
}

func parseInfra(raw json.RawMessage) (Infra, error) {
	if len(raw) > 0 && raw[0] == '[' {
		ids, err := parseStrings(raw)
		return Infra{IDs: ids, List: true}, err
	}
	s, err := parseString(raw)
	if err != nil || s == "" {
		return Infra{}, err
	}
	return Infra{IDs: []string{s}}, nil
}

func parseContent(raw json.RawMessage) (*Content, error) {
	var elems []json.RawMessage
	if len(raw) == 0 || raw[0] != '[' {
		return nil, fmt.Errorf("%s is not a list", raw)
	}
	if err := json.Unmarshal(raw, &elems); err != nil {
		return nil, err
	}
	if len(elems) != 1 && len(elems) != 2 {
		return nil, fmt.Errorf("%d elements, want [owner] or [owner, [units]]", len(elems))
	}
	c := &Content{}
	if err := json.Unmarshal(elems[0], &c.Owner); err != nil {
		return nil, fmt.Errorf("owner: %w", err)
	}
	if len(elems) == 2 {
		units, err := parseStrings(elems[1])
		if err != nil {
			return nil, fmt.Errorf("units: %w", err)
		}
		c.Units = units
	}
	return c, nil
}

// Save writes the level to path in the format of Marshal.
func (l *Level) Save(path string) error {
	raw, err := l.Marshal()
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

// Marshal encodes the level exactly as mapEditor.py saveData does: Python str()
// of each line with ' replaced by ". Strings that this format cannot carry are
// rejected.
func (l *Level) Marshal() ([]byte, error) {
	var b strings.Builder
	w := &writer{b: &b}
	b.WriteString("{\n")
	b.WriteString("\t\"mapName\" :\t")
	w.str(l.MapName)
	b.WriteString(",\n\t\"mapNo\" :\t" + strconv.Itoa(l.MapNo) + ",\n")
	b.WriteString("\t\"player\" :\t")
	w.str(l.Player)
	b.WriteString(",\n\t\"tiles\" :\t{\n")
	for i, line := range l.Lines {
		fmt.Fprintf(&b, "\t\t\t\t\"line%d\" :\t[", i+1)
		for col, h := range line {
			if col > 0 {
				b.WriteString(", ")
			}
			w.hex(h)
		}
		b.WriteString("]")
		if i+1 != len(l.Lines) {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString("\t\t\t\t}\n}\n\n\n\n\n\n\n")
	if w.err != nil {
		return nil, w.err
	}
	return []byte(b.String()), nil
}

type writer struct {
	b   *strings.Builder
	err error
}

func (w *writer) str(s string) {
	bad := strings.ContainsFunc(s, func(r rune) bool { return r == '"' || r == '\'' || r == '\\' || r < 0x20 || r == 0x7f })
	if bad && w.err == nil {
		w.err = fmt.Errorf("%q cannot be written in the level format", s)
	}
	w.b.WriteString(`"` + s + `"`)
}

func (w *writer) strs(ss []string) {
	w.b.WriteString("[")
	for i, s := range ss {
		if i > 0 {
			w.b.WriteString(", ")
		}
		w.str(s)
	}
	w.b.WriteString("]")
}

func (w *writer) hex(h Hex) {
	w.b.WriteString("[")
	w.str(h.Terrain)
	w.b.WriteString(", ")
	switch {
	case h.Infra.List:
		w.strs(h.Infra.IDs)
	case len(h.Infra.IDs) == 0:
		w.str("")
	case len(h.Infra.IDs) == 1:
		w.str(h.Infra.IDs[0])
	default:
		w.str(strings.Join(h.Infra.IDs, ","))
		if w.err == nil {
			w.err = fmt.Errorf("infra %v needs the list form", h.Infra.IDs)
		}
	}
	w.b.WriteString(", ")
	w.str(h.Unit)
	if c := h.Content; c != nil {
		w.b.WriteString(", [" + strconv.Itoa(c.Owner))
		if c.Units != nil {
			w.b.WriteString(", ")
			w.strs(c.Units)
		}
		w.b.WriteString("]")
	}
	w.b.WriteString("]")
}
