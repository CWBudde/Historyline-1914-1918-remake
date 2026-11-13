package saves

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPutGetRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "save.json")
	s := OpenFile(path)
	if s.Has() {
		t.Fatal("empty store has a save")
	}
	want := Envelope{Map: "original/05", Side: 1, Data: []byte{1, 2, 3}}
	if err := s.Put(want); err != nil {
		t.Fatal(err)
	}
	got, ok := OpenFile(path).Get()
	if !ok || got.Map != want.Map || got.Side != want.Side || string(got.Data) != string(want.Data) {
		t.Fatalf("got %+v ok %v", got, ok)
	}
	// Single slot: a second Put replaces the first.
	if err := s.Put(Envelope{Map: "1914-08:cp", Side: 0, Data: []byte{9}}); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Get(); got.Map != "1914-08:cp" {
		t.Fatalf("replaced with %q", got.Map)
	}
	var nilStore *Store
	if nilStore.Has() {
		t.Error("nil store has a save")
	}
	if err := nilStore.Put(want); err == nil {
		t.Error("nil store Put succeeded")
	}
}

// TestModeRoundTrip: a two-human save keeps its mode; a save without one
// (one human, or written before modes) reads as "".
func TestModeRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "save.json")
	s := OpenFile(path)
	if err := s.Put(Envelope{Map: "1914-08:2p", Side: 1, Data: []byte{7}, Mode: "hotseat"}); err != nil {
		t.Fatal(err)
	}
	if got, ok := s.Get(); !ok || got.Mode != "hotseat" || got.Side != 1 {
		t.Fatalf("got %+v ok %v", got, ok)
	}
	if err := os.WriteFile(path, []byte(`{"map":"1914-08:cp","side":0,"data":"AQ=="}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, ok := s.Get(); !ok || got.Mode != "" {
		t.Fatalf("an old save: %+v ok %v", got, ok)
	}
}
