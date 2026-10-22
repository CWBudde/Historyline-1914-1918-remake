package scores

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecordKeepsTheBest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "scores.json")
	s := OpenFile(path)
	if got := s.Best("PULSE"); got != 0 {
		t.Fatalf("empty store: best %d", got)
	}
	steps := []struct {
		score    int
		beat     bool
		wantBest int
	}{
		{861, true, 861},
		{500, false, 861},
		{861, false, 861},
		{900, true, 900},
	}
	for _, st := range steps {
		beat, err := s.Record("PULSE", st.score)
		if err != nil {
			t.Fatal(err)
		}
		if beat != st.beat || s.Best("PULSE") != st.wantBest {
			t.Errorf("record %d: beat %v best %d, want %v %d", st.score, beat, s.Best("PULSE"), st.beat, st.wantBest)
		}
	}
	// A new store reads what the first one kept; other levels stay apart.
	again := OpenFile(path)
	if got := again.Best("PULSE"); got != 900 {
		t.Errorf("reopened best %d, want 900", got)
	}
	if got := again.Best("CIVIL"); got != 0 {
		t.Errorf("CIVIL best %d, want 0", got)
	}
	// A first record of 0 still counts as recorded.
	if beat, _ := again.Record("MOUSE", 0); !beat {
		t.Error("a first score of 0 is not recorded")
	}
}

func TestUnreadableStoreIsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scores.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := OpenFile(path)
	if got := s.Best("PULSE"); got != 0 {
		t.Errorf("best %d from a broken file", got)
	}
	if _, err := s.Record("PULSE", 10); err != nil {
		t.Fatal(err)
	}
	if got := OpenFile(path).Best("PULSE"); got != 10 {
		t.Errorf("rewritten best %d, want 10", got)
	}
}

func TestNoStoreKeepsTheSession(t *testing.T) {
	s := none()
	beat, err := s.Record("PULSE", 5)
	if !beat || err == nil {
		t.Errorf("record without a store: beat %v err %v", beat, err)
	}
	if got := s.Best("PULSE"); got != 5 {
		t.Errorf("session best %d, want 5", got)
	}
	var nilStore *Store
	if nilStore.Best("PULSE") != 0 {
		t.Error("a nil store has a best")
	}
}
