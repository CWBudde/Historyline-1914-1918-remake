package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "settings.json")
	s := OpenFile(path)
	if got := s.Load(); got != Defaults() {
		t.Fatalf("a missing file: %+v, want the defaults", got)
	}
	want := Values{PlayMode: "hotseat", Reduced: true, Opponent: "passive", Limit: 15, Scroll: "fast"}
	if err := s.Save(want); err != nil {
		t.Fatal(err)
	}
	if got := OpenFile(path).Load(); got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	var none *Store
	if none.Load() != Defaults() {
		t.Error("a nil store has values")
	}
	if err := none.Save(want); err == nil {
		t.Error("a nil store saved")
	}
}

// TestUnreadable: a file that cannot be read or parsed counts as the
// defaults, and unknown values fall back to theirs one by one.
func TestUnreadable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	for _, body := range []string{"", "not json", "[1,2]", "null"} {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := OpenFile(path).Load(); got != Defaults() {
			t.Errorf("%q: %+v", body, got)
		}
	}
	if got := OpenFile(dir).Load(); got != Defaults() {
		t.Errorf("a directory: %+v", got)
	}
	body := `{"playMode":"triple","reduced":"yes","opponent":"passive","limit":7,"scroll":"slow","later":1}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	want := Defaults()
	want.Opponent, want.Scroll = "passive", "slow"
	if got := OpenFile(path).Load(); got != want {
		t.Errorf("unknown values: %+v, want %+v", got, want)
	}
}
