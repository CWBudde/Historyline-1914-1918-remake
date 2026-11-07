package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/saves"
)

// -snapshot reads both the save slot file and a bare dos save, and either
// resumes the battle where it was saved.
func TestReadSnapshot(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	lvl, _, err := loadMap("../../levels", "1918-04:cp", set)
	if err != nil {
		t.Fatal(err)
	}
	s, err := dosplay.New(lvl, set, dosplay.Options{Computer: [2]bool{false, true}})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := s.Save()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	bare := filepath.Join(dir, "bare.json")
	if err := os.WriteFile(bare, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	slot := filepath.Join(dir, "slot.json")
	if err := saves.OpenFile(slot).Put(saves.Envelope{Map: "1918-04:cp", Side: 1, Data: raw}); err != nil {
		t.Fatal(err)
	}

	got, err := readSnapshot(bare)
	if err != nil || got.level != "" || got.side != -1 || string(got.data) != string(raw) {
		t.Fatalf("bare save: %+v %v", got.level, err)
	}
	got, err = readSnapshot(slot)
	if err != nil || got.level != "1918-04:cp" || got.side != 1 || string(got.data) != string(raw) {
		t.Fatalf("slot file: level %q side %d err %v", got.level, got.side, err)
	}
	if _, err := dosplay.Load(lvl, set, got.data); err != nil {
		t.Fatal(err)
	}

	if _, err := startSnapshot(slot, "dos", true, "1918-02:cp"); err == nil {
		t.Error("a -level other than the save's was accepted")
	}
	if _, err := startSnapshot(slot, "standard", false, ""); err == nil {
		t.Error("-snapshot accepted the standard rules")
	}
	junk := filepath.Join(dir, "junk")
	if err := os.WriteFile(junk, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readSnapshot(junk); err == nil {
		t.Error("a non-JSON file was accepted")
	}
}
