package saves

import (
	"bytes"
	"encoding/json"
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

// TestPendingRoundTrip: a hot-seat save keeps the playback one human has not
// seen; a save without one writes no field, an old save reads without one,
// and a malformed one is dropped while the battle still loads.
func TestPendingRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "save.json")
	s := OpenFile(path)
	want := &Pending{Side: 1, Before: []byte{4, 5, 6}, Events: json.RawMessage(`[{"Kind":5,"Side":0}]`)}
	if err := s.Put(Envelope{Map: "1914-08:2p", Side: 0, Data: []byte{7}, Mode: "hotseat", Pending: want}); err != nil {
		t.Fatal(err)
	}
	got, ok := s.Get()
	if !ok || got.Pending == nil || got.Pending.Side != 1 || !bytes.Equal(got.Pending.Before, want.Before) {
		t.Fatalf("got %+v ok %v", got, ok)
	}
	var evs []map[string]int
	if err := json.Unmarshal(got.Pending.Events, &evs); err != nil || len(evs) != 1 || evs[0]["Kind"] != 5 {
		t.Fatalf("events %s: %v", got.Pending.Events, err)
	}
	if err := s.Put(Envelope{Map: "1914-08:2p", Side: 0, Data: []byte{7}, Mode: "hotseat"}); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(path); err != nil || bytes.Contains(b, []byte("pending")) {
		t.Errorf("a save without a pending playback writes the field:\n%s", b)
	}
	for _, body := range []string{
		`{"map":"1914-08:2p","side":0,"data":"AQ==","mode":"hotseat"}`,
		`{"map":"1914-08:2p","side":0,"data":"AQ==","mode":"hotseat","pending":null}`,
		`{"map":"1914-08:2p","side":0,"data":"AQ==","mode":"hotseat","pending":5}`,
		`{"map":"1914-08:2p","side":0,"data":"AQ==","mode":"hotseat","pending":{"side":2,"before":"AQ==","events":[]}}`,
		`{"map":"1914-08:2p","side":0,"data":"AQ==","mode":"hotseat","pending":{"side":1,"before":"","events":[]}}`,
		`{"map":"1914-08:2p","side":0,"data":"AQ==","mode":"hotseat","pending":{"side":1,"before":"AQ==","events":{}}}`,
		`{"map":"1914-08:2p","side":0,"data":"AQ==","mode":"hotseat","pending":{"side":1,"before":"!!","events":[]}}`,
	} {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if got, ok := s.Get(); !ok || got.Pending != nil || got.Map != "1914-08:2p" || string(got.Data) != "\x01" {
			t.Errorf("%s: got %+v ok %v", body, got, ok)
		}
	}
}
