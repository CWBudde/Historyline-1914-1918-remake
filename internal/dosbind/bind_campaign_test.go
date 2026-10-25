package dosbind_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// The shipped campaign levels (levels/campaign, the editable remake
// transcriptions of the original maps) must all bind for -rules dos.
func TestBindCampaignLevels(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob("../../levels/campaign/*.json")
	if err != nil || len(files) != 72 {
		t.Fatalf("campaign files %d, err %v", len(files), err)
	}
	sort.Strings(files)
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		lv, err := level.Parse(raw)
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		b, err := dosbind.Bind(lv, set)
		if err != nil {
			t.Errorf("%s: %v", filepath.Base(f), err)
			continue
		}
		if _, err := b.Game(dosrules.Full); err != nil {
			t.Errorf("%s: game: %v", filepath.Base(f), err)
		}
	}
}
