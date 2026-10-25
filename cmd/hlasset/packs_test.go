package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdatePacksBuildsOnlyChangedCatalogs(t *testing.T) {
	out := t.TempDir()
	inputs := map[string]string{"a": "one", "b": "two"}
	builds := map[string]int{}
	fail := ""
	var cs []catalog
	for _, name := range []string{"a", "b"} {
		cs = append(cs, catalog{name,
			func() ([]byte, error) { return []byte(inputs[name]), nil },
			func(dir string) error {
				builds[name]++
				if fail == name {
					return errors.New("broken")
				}
				if err := os.Mkdir(dir, 0o755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(dir, "content"), []byte(inputs[name]), 0o644)
			}})
	}
	update := func(builder string, only ...string) string {
		t.Helper()
		var log bytes.Buffer
		if err := updatePacks(out, []byte(builder), cs, only, &log); err != nil {
			t.Fatal(err)
		}
		return log.String()
	}
	content := func(name string) string {
		b, err := os.ReadFile(filepath.Join(out, name, "content"))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	update("v1")
	if log := update("v1"); builds["a"] != 1 || builds["b"] != 1 || strings.Count(log, "up to date") != 2 {
		t.Fatalf("second run rebuilt: %v\n%s", builds, log)
	}
	inputs["a"] = "three"
	update("v1")
	if builds["a"] != 2 || builds["b"] != 1 || content("a") != "three" {
		t.Fatalf("changed input: builds %v, a holds %q", builds, content("a"))
	}
	update("v2", "b")
	if builds["a"] != 2 || builds["b"] != 2 {
		t.Fatalf("new builder with -only b: builds %v", builds)
	}

	inputs["a"], fail = "four", "a"
	if err := updatePacks(out, []byte("v2"), cs, nil, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "broken") {
		t.Fatal("failed build not reported", err)
	}
	if content("a") != "three" {
		t.Fatal("failed build replaced the old catalog")
	}
	if _, err := os.Stat(filepath.Join(out, "a.new")); !os.IsNotExist(err) {
		t.Fatal("failed build left its directory", err)
	}
	if err := updatePacks(out, []byte("v2"), cs, []string{"c"}, &bytes.Buffer{}); err == nil {
		t.Fatal("unknown catalog accepted")
	}
}

// A unit's new art changes the units' inputs and neither the terrain's nor the
// screens'; the screens' inputs include the art/screens paintings.
func TestCatalogInputsFollowTheirEntries(t *testing.T) {
	index, err := os.ReadFile(filepath.Join(assetsRoot, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var unit indexEntry
	for _, e := range retainedIndex(t) {
		if unitEntry(e) {
			unit = e
			break
		}
	}
	changed := t.TempDir()
	newSum := strings.Repeat("0", len(unit.SHA256))
	if err := os.WriteFile(filepath.Join(changed, "index.json"), bytes.Replace(index, []byte(unit.SHA256), []byte(newSum), 1), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name    string
		keep    func(indexEntry) bool
		changes bool
	}{
		{"units", unitEntry, true},
		{"terrain", terrainEntry, false},
		{"screens", screenEntry, false},
	} {
		was, err := indexInputs(assetsRoot, c.keep)
		if err != nil {
			t.Fatal(err)
		}
		is, err := indexInputs(changed, c.keep)
		if err != nil {
			t.Fatal(err)
		}
		if len(was) < 1000 || bytes.Equal(was, is) == c.changes {
			t.Fatalf("%s: %d bytes of inputs, changed %v, want %v", c.name, len(was), !bytes.Equal(was, is), c.changes)
		}
	}
	screens, err := screenInputs(assetsRoot, filepath.Join("..", "..", "art"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(screens, []byte("\nscreens/surface/")) {
		t.Fatal("screen inputs miss the art/screens paintings")
	}
}
