package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origmaps"
)

func write(t *testing.T, path string, b []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

// A synthetic build of unpacked files: map 01 with a shop and a computer
// file, and a map without a number or shop.
func TestExtractSynthetic(t *testing.T) {
	dir, out := t.TempDir(), t.TempDir()
	write(t, filepath.Join(dir, "MAP", "01.FIN"), []byte{0, 2, 0, 1, 0x37, 0x24, 0x37, 0xff})
	shop := make([]byte, finmap.TypeCount+1)
	write(t, filepath.Join(dir, "MAP", "01.SHP"), shop)
	write(t, filepath.Join(dir, "MAP", "01.COM"), make([]byte, 6*finmap.TypeCount))
	write(t, filepath.Join(dir, "MAP", "STATMAP.FIN"), []byte{0, 1, 0, 1, 0x37, 0xff})
	codes := []byte{
		0x11, 0x11, 0x11, 0x11, 0x11, 2, 1, 1, 0, 0,
		0x12, 0x12, 0x12, 0x12, 0x12, 2, 1, 0, 2, 0,
	}
	write(t, filepath.Join(dir, "CODES.DAT"), codes)

	var log bytes.Buffer
	n, err := extract([]origmaps.Build{{Name: "test", Dir: dir, Tables: dosrules.FullTables}}, out, &log)
	if err != nil || n != 2 {
		t.Fatalf("extract: %d, %v\n%s", n, err, log.String())
	}
	raw, err := os.ReadFile(filepath.Join(out, "test", "01.json"))
	if err != nil {
		t.Fatal(err)
	}
	var l finmap.Level
	if err := json.Unmarshal(raw, &l); err != nil {
		t.Fatal(err)
	}
	if l.Version != finmap.Version || l.Source.Map != "01" || l.Mission == nil || l.Mission.Next != nil || !l.Mission.Winter || l.Mission.Code != "BBBBB" {
		t.Fatalf("level %s", raw)
	}
	if len(l.Units) != 1 || l.Units[0].Type != 18 || l.Units[0].Side != 0 || len(l.Producible) != finmap.TypeCount || !l.Computer {
		t.Fatalf("level %s", raw)
	}
	want := []string{"MAP/01.FIN", "MAP/01.SHP", "MAP/01.COM", "CODES.DAT"}
	if strings.Join(l.Source.Files, ",") != strings.Join(want, ",") {
		t.Fatalf("files %v", l.Source.Files)
	}
	raw, err = os.ReadFile(filepath.Join(out, "test", "STATMAP.json"))
	if err != nil || bytes.Contains(raw, []byte(`"mission"`)) {
		t.Fatalf("STATMAP: %v %s", err, raw)
	}
	var idx []indexEntry
	raw, err = os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil || json.Unmarshal(raw, &idx) != nil || len(idx) != 2 {
		t.Fatalf("index: %v %s", err, raw)
	}
}

func TestExtractRejects(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "MAP", "00.FIN"), []byte{0, 2, 0, 1, 0x37})
	write(t, filepath.Join(dir, "CODES.DAT"), []byte{0x11, 0x11, 0x11, 0x11, 0x11, 2, 1, 0, 0, 0})
	if _, err := extract([]origmaps.Build{{Name: "bad", Dir: dir, Tables: dosrules.FullTables}}, t.TempDir(), &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "00: finmap: FIN") {
		t.Fatalf("expected an error for the truncated map, got %v", err)
	}
	// A win on the last map would lead past the last record.
	write(t, filepath.Join(dir, "MAP", "00.FIN"), []byte{0, 1, 0, 1, 0x37, 0xff})
	write(t, filepath.Join(dir, "CODES.DAT"), []byte{0x11, 0x11, 0x11, 0x11, 0x11, 2, 1, 1, 0, 0})
	if _, err := extract([]origmaps.Build{{Name: "bad", Dir: dir, Tables: dosrules.FullTables}}, t.TempDir(), &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "next map") {
		t.Fatalf("expected an error for the next map, got %v", err)
	}
}

// Every map of the original full game and demo decodes; skipped without
// the original files.
func TestExtractOriginal(t *testing.T) {
	root, err := origmaps.Find(".")
	if err != nil {
		t.Skip("no original files")
	}
	var log bytes.Buffer
	n, err := extract(origmaps.Builds(root), t.TempDir(), &log)
	if err != nil {
		t.Fatalf("%v\n%s", err, log.String())
	}
	// 72 numbered maps and STATMAP in the full game, map 00 and STATMAP in
	// the demo.
	if n != 75 {
		t.Fatalf("%d maps", n)
	}
}

func TestSelectBuilds(t *testing.T) {
	all := origmaps.Builds("orig")
	got, err := selectBuilds(all, "full")
	if err != nil || len(got) != 1 || got[0].Name != "full" {
		t.Fatalf("full: %v %v", got, err)
	}
	if got, err := selectBuilds(all, "full,demo"); err != nil || len(got) != 2 {
		t.Fatalf("full,demo: %v %v", got, err)
	}
	for _, names := range []string{"full,amiga", "full,", "full,full"} {
		if _, err := selectBuilds(all, names); err == nil {
			t.Errorf("%q: expected an error", names)
		}
	}
}
