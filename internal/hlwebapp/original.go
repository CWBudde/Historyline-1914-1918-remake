package hlwebapp

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origmaps"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origview"
)

// OriginalIndex lists the maps of an original export, in map order.
type OriginalIndex struct {
	Schema      int      `json:"schema"`
	Resolutions int      `json:"resolutions"`
	Maps        []string `json:"maps"`
}

// OriginalMap is a map's binding as Go makes it, numbers and remake ids only:
// the finmap level, the exact board (dosbind.FromOriginal) and the view the
// front end draws (origview.Campaign, in the level format).
type OriginalMap struct {
	Schema int           `json:"schema"`
	Level  *finmap.Level `json:"level"`
	Board  OriginalBoard `json:"board"`
	View   string        `json:"view"`
}

// OriginalBoard is dosbind.Board with its maps as sorted lists.
type OriginalBoard struct {
	Width     int                `json:"width"`
	Height    int                `json:"height"`
	Winter    bool               `json:"winter"`
	Terrain   []int              `json:"terrain"`
	Units     [][4]int           `json:"units"` // index, type, owner, cell
	IDs       map[string]string  `json:"ids"`
	Holds     []OriginalHold     `json:"holds"`
	Buildings []OriginalBuilding `json:"buildings"`
	Parts     map[string]int     `json:"parts"`
}

type OriginalHold struct {
	Index int `json:"index"`
	dosrules.Hold
}

type OriginalBuilding struct {
	Index          int   `json:"index"`
	Credits        int   `json:"credits"`
	Builds         []int `json:"builds"`
	ComputerBuilds []int `json:"computerBuilds"`
}

// Original writes the original maps' Go bindings and AI-vs-AI traces from the
// local original files (git-ignored original/) to build/webapp/original, which
// the web edition's tests check field for field. Without the original files it
// writes nothing and says so.
func Original(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	f := flag.NewFlagSet("hl webapp original", flag.ContinueOnError)
	f.SetOutput(stderr)
	out := f.String("out", "build/webapp/original", "output directory")
	resolutions := f.Int("resolutions", 300, "maximum resolutions per map trace (0 exports initial states only)")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 || *resolutions < 0 {
		return errors.New("hl webapp original: expected flags and a nonnegative resolution cap")
	}
	levels, err := origmaps.Full()
	if errors.Is(err, origmaps.ErrNotFound) {
		_, err = fmt.Fprintln(stdout, "no original files: nothing to export")
		return err
	}
	if err != nil {
		return err
	}
	set, err := data.Load()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	index := OriginalIndex{Schema: 1, Resolutions: *resolutions}
	for _, l := range levels {
		if err := ctx.Err(); err != nil {
			return err
		}
		stem := l.Source.Map
		if err := writeOriginal(ctx, *out, l, set, *resolutions); err != nil {
			return fmt.Errorf("original/%s: %w", stem, err)
		}
		index.Maps = append(index.Maps, stem)
	}
	raw, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(*out, "index.json"), append(raw, '\n'), 0o644); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "wrote %d original maps to %s\n", len(index.Maps), *out)
	return err
}

func writeOriginal(ctx context.Context, out string, l *finmap.Level, set *data.Set, resolutions int) (err error) {
	stem := l.Source.Map
	b, err := dosbind.FromOriginal(l, set)
	if err != nil {
		return err
	}
	lvl, err := origview.Campaign(l, set)
	if err != nil {
		return err
	}
	view, err := lvl.Marshal()
	if err != nil {
		return err
	}
	raw, err := json.Marshal(OriginalMap{Schema: 1, Level: l, Board: originalBoard(b), View: string(view)})
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, stem+".json"), append(raw, '\n'), 0o644); err != nil {
		return err
	}
	s, err := dosplay.NewBoard(b, lvl, set, dosplay.Options{Computer: [2]bool{true, true}})
	if err != nil {
		return err
	}
	file, err := os.Create(filepath.Join(out, stem+".jsonl.gz"))
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}()
	z := gzip.NewWriter(file)
	err = writeSessionTrace(ctx, z, "original/"+stem, s, resolutions)
	if closeErr := z.Close(); err == nil {
		err = closeErr
	}
	return err
}

func originalBoard(b *dosbind.Board) OriginalBoard {
	o := OriginalBoard{
		Width: b.Width, Height: b.Height, Winter: b.Winter, Terrain: b.Terrain,
		Units: [][4]int{}, IDs: map[string]string{}, Holds: []OriginalHold{}, Buildings: []OriginalBuilding{}, Parts: map[string]int{},
	}
	for _, i := range slices.Sorted(maps.Keys(b.Units)) {
		u := b.Units[i]
		o.Units = append(o.Units, [4]int{i, u.Type, u.Owner, u.Cell})
	}
	for i, id := range b.IDs {
		o.IDs[strconv.Itoa(i)] = id
	}
	for _, i := range slices.Sorted(maps.Keys(b.Holds)) {
		o.Holds = append(o.Holds, OriginalHold{Index: i, Hold: b.Holds[i]})
	}
	for _, i := range slices.Sorted(maps.Keys(b.Buildings)) {
		bd := b.Buildings[i]
		o.Buildings = append(o.Buildings, OriginalBuilding{Index: i, Credits: bd.Credits, Builds: orEmpty(bd.Builds), ComputerBuilds: orEmpty(bd.ComputerBuilds)})
	}
	for c, e := range b.Parts {
		o.Parts[strconv.Itoa(c)] = e
	}
	return o
}

func orEmpty(s []int) []int {
	if s == nil {
		return []int{}
	}
	return s
}
