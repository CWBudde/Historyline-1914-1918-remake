package campaign

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

func TestEveryMonthHasAFrontPage(t *testing.T) {
	files, err := fs.Glob(levels.Monthly, "monthly/*.json")
	if err != nil || len(files) == 0 {
		t.Fatal(files, err)
	}
	for _, f := range files {
		month := strings.TrimSuffix(strings.TrimPrefix(f, "monthly/"), ".json")
		if _, ok := News(month); !ok {
			t.Errorf("%s has no front page", month)
		}
	}
	if len(Articles) != len(files) {
		t.Errorf("%d front pages for %d monthly levels", len(Articles), len(files))
	}
	for month, a := range Articles {
		if a.Label == "" || a.Headline == "" || a.Deck == "" || a.Caption == "" || len(a.Body) < 2 || len(a.Briefs) != 3 {
			t.Errorf("%s: incomplete page %+v", month, a)
		}
		if !strings.HasSuffix(a.Label, month[:4]) {
			t.Errorf("%s: label %q", month, a.Label)
		}
	}
	if _, ok := News("1918-08"); ok {
		t.Error("1918-08 has a front page")
	}
}
