package hlwebapp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origmaps"
)

// TestOriginalExport writes the original maps' bindings and short traces from
// the local original files; it skips without them.
func TestOriginalExport(t *testing.T) {
	if _, err := origmaps.Full(); errors.Is(err, origmaps.ErrNotFound) {
		t.Skip(err)
	}
	out := t.TempDir()
	var stdout bytes.Buffer
	if err := Original(context.Background(), []string{"-out", out, "-resolutions", "1"}, &stdout, &stdout); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var index OriginalIndex
	if err := json.Unmarshal(raw, &index); err != nil {
		t.Fatal(err)
	}
	if len(index.Maps) != 73 || index.Maps[1] != "01" {
		t.Fatalf("maps %v", index.Maps)
	}
	raw, err = os.ReadFile(filepath.Join(out, "01.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m OriginalMap
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m.Board.Width != m.Level.Width || len(m.Board.Terrain) != m.Level.Width*m.Level.Height || len(m.Board.Units) == 0 || m.View == "" {
		t.Fatalf("map 01: board %dx%d with %d units", m.Board.Width, m.Board.Height, len(m.Board.Units))
	}
}

// TestOriginalExportWithoutFiles: without the original files the export
// writes nothing and succeeds.
func TestOriginalExportWithoutFiles(t *testing.T) {
	if _, err := origmaps.Full(); err == nil {
		t.Skip("the original files are present")
	}
	out := filepath.Join(t.TempDir(), "original")
	var stdout bytes.Buffer
	if err := Original(context.Background(), []string{"-out", out}, &stdout, &stdout); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(out); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("output written: %v", err)
	}
}
