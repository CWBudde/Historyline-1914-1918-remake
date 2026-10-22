// Command finextract imports every map of the original DOS full game and
// demo (MAP/*.FIN with their .SHP and .COM files and CODES.DAT) into the
// versioned level JSON of internal/finmap, one file per map plus index.json.
//
//	go run ./cmd/finextract [-original dir] [-out build/scratch/levels] [-build full,demo]
//
// just levels-import writes the full game's maps to levels/original, the
// committed copy.
package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/tpwm"
)

// build is one installation of the game and the tables of its executable.
type build struct {
	Name   string
	Dir    string
	Tables *dosrules.Tables
}

func defaultBuilds(original string) []build {
	return []build{
		{Name: "full", Dir: filepath.Join(original, "extracted", "dos-full"), Tables: dosrules.FullTables},
		{Name: "demo", Dir: filepath.Join(original, "demo", "HL"), Tables: dosrules.DemoTables},
	}
}

// findOriginal looks for the original/ directory from dir upwards.
func findOriginal(dir string) (string, error) {
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
			return "", errors.New("no original/extracted/dos-full/MAP above the working directory")
		}
		dir = parent
	}
}

// repoRoot is the directory of go.mod from dir upwards.
func repoRoot(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no go.mod above the working directory")
		}
		dir = parent
	}
}

func main() {
	original := flag.String("original", "", "original/ directory (default: found above the working directory)")
	out := flag.String("out", "", "output directory (default: build/scratch/levels in the repository)")
	names := flag.String("build", "full,demo", "comma-separated builds to import")
	flag.Parse()
	if *original == "" {
		o, err := findOriginal(".")
		if err != nil {
			log.Fatal(err)
		}
		*original = o
	}
	if *out == "" {
		root, err := repoRoot(".")
		if err != nil {
			log.Fatal(err)
		}
		*out = filepath.Join(root, "build", "scratch", "levels")
	}
	builds, err := selectBuilds(defaultBuilds(*original), *names)
	if err != nil {
		log.Fatal(err)
	}
	n, err := extract(builds, *out, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %d levels to %s\n", n, *out)
}

// selectBuilds returns the builds named in the comma-separated list.
func selectBuilds(all []build, names string) ([]build, error) {
	var out []build
	for _, name := range strings.Split(names, ",") {
		i := slices.IndexFunc(all, func(b build) bool { return b.Name == name })
		switch {
		case i < 0:
			return nil, fmt.Errorf("unknown build %q", name)
		case slices.ContainsFunc(out, func(b build) bool { return b.Name == name }):
			return nil, fmt.Errorf("build %q named twice", name)
		}
		out = append(out, all[i])
	}
	return out, nil
}

func rules(t *dosrules.Tables) (finmap.Rules, error) {
	if t == nil || t.Move == nil {
		return finmap.Rules{}, errors.New("tables without terrain and type records")
	}
	return finmap.RulesFrom(t), nil
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

// code is the level's password for the log, "" without a mission.
func code(l *finmap.Level) string {
	if l.Mission == nil {
		return ""
	}
	return l.Mission.Code
}

// indexEntry summarises one written level.
type indexEntry struct {
	Build     string `json:"build"`
	Map       string `json:"map"`
	File      string `json:"file"`
	Players   int    `json:"players,omitempty"`
	Winter    bool   `json:"winter,omitempty"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Units     int    `json:"units"`
	Buildings int    `json:"buildings"`
}

// extract writes the levels of every build to out/<build>/<map>.json and
// out/index.json, and returns how many it wrote.
func extract(builds []build, out string, logw io.Writer) (int, error) {
	var index []indexEntry
	for _, b := range builds {
		r, err := rules(b.Tables)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", b.Name, err)
		}
		codesRaw, err := unpack(filepath.Join(b.Dir, "CODES.DAT"))
		if err != nil {
			return 0, err
		}
		codes, err := finmap.DecodeCodes(codesRaw)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", b.Name, err)
		}
		fins, err := filepath.Glob(filepath.Join(b.Dir, "MAP", "*.FIN"))
		if err != nil {
			return 0, err
		}
		sort.Strings(fins)
		if len(fins) == 0 {
			return 0, fmt.Errorf("%s: no maps in %s", b.Name, filepath.Join(b.Dir, "MAP"))
		}
		if err := os.MkdirAll(filepath.Join(out, b.Name), 0o755); err != nil {
			return 0, err
		}
		for _, fin := range fins {
			l, err := level(b, fin, codes, r)
			if err != nil {
				return 0, fmt.Errorf("%s: %w", b.Name, err)
			}
			raw, err := json.Marshal(l)
			if err != nil {
				return 0, err
			}
			file := filepath.Join(b.Name, l.Source.Map+".json")
			if err := os.WriteFile(filepath.Join(out, file), append(raw, '\n'), 0o644); err != nil {
				return 0, err
			}
			e := indexEntry{Build: b.Name, Map: l.Source.Map, File: filepath.ToSlash(file), Width: l.Width, Height: l.Height,
				Units: len(l.Units), Buildings: len(l.Buildings)}
			if m := l.Mission; m != nil {
				e.Players, e.Winter = m.Players, m.Winter
			}
			index = append(index, e)
			if _, err := fmt.Fprintf(logw, "%s %-7s %2dx%-2d %5s units=%d buildings=%d\n", b.Name, l.Source.Map, l.Width, l.Height, code(l), e.Units, e.Buildings); err != nil {
				return 0, err
			}
		}
	}
	raw, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return 0, err
	}
	if err := os.WriteFile(filepath.Join(out, "index.json"), append(raw, '\n'), 0o644); err != nil {
		return 0, err
	}
	return len(index), nil
}

// level decodes one map with its SHP, COM and CODES.DAT record.
func level(b build, fin string, codes []finmap.Code, r finmap.Rules) (*finmap.Level, error) {
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
