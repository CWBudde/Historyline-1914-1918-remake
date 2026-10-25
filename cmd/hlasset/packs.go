package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// catalog is one pack catalog of the packs command: inputs returns what it is
// built from, build writes it into a new directory.
type catalog struct {
	name   string
	inputs func() ([]byte, error)
	build  func(out string) error
}

// catalogs are the unit, terrain and screen catalogs built from assetsDir and
// artDir. Each one's inputs are the index entries it packs (their sha256 stands
// for the file); the screens add the art/screens paintings by content.
func catalogs(assetsDir, artDir string) []catalog {
	return []catalog{
		{"units",
			func() ([]byte, error) { return indexInputs(assetsDir, unitEntry) },
			func(out string) error { _, err := buildUnits(assetsDir, out); return err }},
		{"terrain",
			func() ([]byte, error) { return indexInputs(assetsDir, terrainEntry) },
			func(out string) error { _, err := buildTerrain(assetsDir, out); return err }},
		{"screens",
			func() ([]byte, error) { return screenInputs(assetsDir, artDir) },
			func(out string) error { _, err := buildScreens(assetsDir, artDir, out); return err }},
	}
}

// unitEntry and terrainEntry report whether e is packed into the unit or the
// terrain catalog.
func unitEntry(e indexEntry) bool    { return strings.HasPrefix(e.Pack, "unit-") }
func terrainEntry(e indexEntry) bool { return strings.HasPrefix(e.Pack, "terrain-") }

// indexInputs returns the index entries of assetsDir that keep selects, as JSON.
func indexInputs(assetsDir string, keep func(indexEntry) bool) ([]byte, error) {
	entries, err := readIndex(assetsDir)
	if err != nil {
		return nil, err
	}
	return json.Marshal(slices.DeleteFunc(entries, func(e indexEntry) bool { return !keep(e) }))
}

// screenInputs returns the screen entries of the index followed by the path and
// sha256 of every art/screens/<family>/*.png planScreens may read.
func screenInputs(assetsDir, artDir string) ([]byte, error) {
	b, err := indexInputs(assetsDir, screenEntry)
	if err != nil {
		return nil, err
	}
	srcs, err := filepath.Glob(filepath.Join(artDir, "screens", "*", "*.png"))
	if err != nil {
		return nil, err
	}
	slices.Sort(srcs)
	for _, src := range srcs {
		f, err := os.Open(src)
		if err != nil {
			return nil, err
		}
		h := sha256.New()
		_, err = io.Copy(h, f)
		_ = f.Close()
		if err != nil {
			return nil, err
		}
		rel, err := filepath.Rel(artDir, src)
		if err != nil {
			return nil, err
		}
		b = fmt.Appendf(b, "\n%s %x", filepath.ToSlash(rel), h.Sum(nil))
	}
	return b, nil
}

// executableSum is the sha256 of the running hlasset, so a change to any code
// it links rebuilds the catalogs.
func executableSum() ([]byte, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(exe)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// stampName is the file in each built catalog naming the key it was built for.
const stampName = "stamp"

// updatePacks brings each catalog under out up to date: a catalog whose stamp
// matches its key (builder and inputs) is kept, any other is built into
// out/<name>.new and then replaces out/<name>. A failed build keeps the old
// catalog. only, when not empty, limits the catalogs by name.
func updatePacks(out string, builder []byte, cs []catalog, only []string, w io.Writer) error {
	for _, name := range only {
		if !slices.ContainsFunc(cs, func(c catalog) bool { return c.name == name }) {
			return fmt.Errorf("unknown catalog %q", name)
		}
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	for _, c := range cs {
		if len(only) > 0 && !slices.Contains(only, c.name) {
			continue
		}
		inputs, err := c.inputs()
		if err != nil {
			return fmt.Errorf("%s: %w", c.name, err)
		}
		h := sha256.New()
		h.Write(builder)
		h.Write([]byte("\x00" + c.name + "\x00"))
		h.Write(inputs)
		key := hex.EncodeToString(h.Sum(nil)) + "\n"
		dir := filepath.Join(out, c.name)
		if b, err := os.ReadFile(filepath.Join(dir, stampName)); err == nil && bytes.Equal(b, []byte(key)) {
			if _, err := fmt.Fprintf(w, "%s: up to date\n", c.name); err != nil {
				return err
			}
			continue
		}
		start := time.Now()
		tmp := dir + ".new"
		if err := os.RemoveAll(tmp); err != nil {
			return err
		}
		if err := c.build(tmp); err != nil {
			_ = os.RemoveAll(tmp)
			return fmt.Errorf("%s: %w", c.name, err)
		}
		if err := os.WriteFile(filepath.Join(tmp, stampName), []byte(key), 0o644); err != nil {
			return err
		}
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
		if err := os.Rename(tmp, dir); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "%s: built in %s\n", c.name, time.Since(start).Round(time.Second)); err != nil {
			return err
		}
	}
	return nil
}
