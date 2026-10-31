// Package origmaps reads the maps of the local original game files
// (original/ above the working directory, git-ignored) into finmap levels:
// each map's MAP/*.FIN with its .SHP and .COM files and its CODES.DAT
// record. The repository ships none of them; internal/hlcampaign writes the
// scenarios from the full game's maps, -level original/NN plays one exactly,
// and the tests that check the original rules on the real maps skip without
// them (ErrNotFound).
package origmaps

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/tpwm"
)

// ErrNotFound is the error of Find and Full without the original files.
var ErrNotFound = errors.New("origmaps: no original/extracted/dos-full/MAP above the working directory")

// Build is one installation of the game and the tables of its executable.
type Build struct {
	Name   string
	Dir    string
	Tables *dosrules.Tables
}

// Builds are the full game and the demo below the original/ directory.
func Builds(original string) []Build {
	return []Build{
		{Name: "full", Dir: filepath.Join(original, "extracted", "dos-full"), Tables: dosrules.FullTables},
		{Name: "demo", Dir: filepath.Join(original, "demo", "HL"), Tables: dosrules.DemoTables},
	}
}

// Find looks for the original/ directory from dir upwards.
func Find(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		o := filepath.Join(dir, "original")
		if st, err := os.Stat(filepath.Join(o, "extracted", "dos-full", "MAP")); err == nil && st.IsDir() {
			return o, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNotFound
		}
		dir = parent
	}
}

// encoded is the full game's maps in finmap's JSON, in map order.
type encoded struct {
	names []string
	raw   map[string][]byte
}

// full is the full game's maps above the working directory, read once.
var full = sync.OnceValues(func() (*encoded, error) {
	o, err := Find(".")
	if err != nil {
		return nil, err
	}
	ls, err := Read(Builds(o)[0])
	if err != nil {
		return nil, err
	}
	e := &encoded{raw: map[string][]byte{}}
	for _, l := range ls {
		if e.raw[l.Source.Map], err = json.Marshal(l); err != nil {
			return nil, err
		}
		e.names = append(e.names, l.Source.Map)
	}
	return e, nil
})

// Full are the full game's maps in map order, 00 to 71 and STATMAP, from
// original/ above the working directory; ErrNotFound without the files.
// The files are read once, each call returns new levels.
func Full() ([]*finmap.Level, error) {
	e, err := full()
	if err != nil {
		return nil, err
	}
	out := make([]*finmap.Level, len(e.names))
	for i, n := range e.names {
		if out[i], err = Map(n); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// Map is a new level of the full game's map stem ("05"), as Full.
func Map(stem string) (*finmap.Level, error) {
	e, err := full()
	if err != nil {
		return nil, err
	}
	b, ok := e.raw[stem]
	if !ok {
		return nil, fmt.Errorf("origmaps: no map %q", stem)
	}
	return finmap.Parse(b)
}

// Read decodes every map of the build in file name order.
func Read(b Build) ([]*finmap.Level, error) {
	if b.Tables == nil || b.Tables.Move == nil {
		return nil, fmt.Errorf("%s: tables without terrain and type records", b.Name)
	}
	r := finmap.RulesFrom(b.Tables)
	codesRaw, err := unpack(filepath.Join(b.Dir, "CODES.DAT"))
	if err != nil {
		return nil, err
	}
	codes, err := finmap.DecodeCodes(codesRaw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", b.Name, err)
	}
	fins, err := filepath.Glob(filepath.Join(b.Dir, "MAP", "*.FIN"))
	if err != nil {
		return nil, err
	}
	sort.Strings(fins)
	if len(fins) == 0 {
		return nil, fmt.Errorf("%s: no maps in %s", b.Name, filepath.Join(b.Dir, "MAP"))
	}
	var out []*finmap.Level
	for _, fin := range fins {
		l, err := level(b, fin, codes, r)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", b.Name, err)
		}
		out = append(out, l)
	}
	return out, nil
}

// unpack reads a file and undoes its TPWM packing; plain files pass through.
func unpack(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	u, err := tpwm.Unpack(b, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return u, nil
}

// level decodes one map with its SHP, COM and CODES.DAT record.
func level(b Build, fin string, codes []finmap.Code, r finmap.Rules) (*finmap.Level, error) {
	stem := strings.TrimSuffix(filepath.Base(fin), filepath.Ext(fin))
	raw, err := unpack(fin)
	if err != nil {
		return nil, err
	}
	m, err := finmap.DecodeMap(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", stem, err)
	}
	in := finmap.Input{Index: -1, Map: m}
	files := []string{"MAP/" + stem + ".FIN"}
	if raw, err := unpack(filepath.Join(b.Dir, "MAP", stem+".SHP")); err == nil {
		if in.Shop, err = finmap.DecodeShop(raw); err != nil {
			return nil, fmt.Errorf("%s: %w", stem, err)
		}
		files = append(files, "MAP/"+stem+".SHP")
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if raw, err := unpack(filepath.Join(b.Dir, "MAP", stem+".COM")); err == nil {
		if in.Computer, err = finmap.DecodeComputer(raw); err != nil {
			return nil, fmt.Errorf("%s: %w", stem, err)
		}
		files = append(files, "MAP/"+stem+".COM")
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if i, err := strconv.Atoi(stem); err == nil && i >= 0 {
		in.Index = i
		if i < len(codes) {
			in.Code = &codes[i]
			files = append(files, "CODES.DAT")
		}
	}
	l, err := finmap.Build(in, r)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", stem, err)
	}
	if m := l.Mission; m != nil && m.Next != nil && *m.Next >= len(codes) {
		return nil, fmt.Errorf("%s: next map %d, CODES.DAT has %d", stem, *m.Next, len(codes))
	}
	l.Source = finmap.Source{Build: b.Name, Map: stem, Files: files}
	return l, nil
}
