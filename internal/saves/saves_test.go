package saves

import (
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
	if err := s.Put(Envelope{Map: "monthly/1914-08-pulse:cp", Side: 0, Data: []byte{9}}); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Get(); got.Map != "monthly/1914-08-pulse:cp" {
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
