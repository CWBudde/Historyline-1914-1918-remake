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
	pages, err := fs.Glob(levels.Monthly, Dir+"/*.json")
	if err != nil || len(pages) != len(files) {
		t.Fatalf("%d front pages for %d monthly levels (%v)", len(pages), len(files), err)
	}
	for _, f := range files {
		month := strings.TrimSuffix(strings.TrimPrefix(f, "monthly/"), ".json")
		raw, err := fs.ReadFile(levels.Monthly, Dir+"/"+month+".json")
		if err != nil {
			t.Errorf("%s has no front page", month)
			continue
		}
		a, err := Parse(raw)
		if err != nil {
			t.Errorf("%s: %v", month, err)
			continue
		}
		if len(a.Body) < 2 || len(a.Briefs) != 3 {
			t.Errorf("%s: %d paragraphs, %d briefs", month, len(a.Body), len(a.Briefs))
		}
		if !strings.HasSuffix(a.Label, month[:4]) {
			t.Errorf("%s: label %q", month, a.Label)
		}
	}
}

func TestParseWantsAWholePage(t *testing.T) {
	for _, raw := range []string{
		`{"label": "May 1915", "headline": "H", "deck": "D", "caption": "C", "body": ["p"]}`,
		`{"label": "May 1915", "headline": "H", "deck": "D", "body": ["p"], "briefs": ["b"]}`,
		`{"label": "May 1915", "headline": "H", "deck": "D", "caption": "C", "body": ["p"], "briefs": ["b"], "author": "x"}`,
	} {
		if _, err := Parse([]byte(raw)); err == nil {
			t.Errorf("Parse(%s) took an incomplete page", raw)
		}
	}
	if Month("monthly/news/1916-02.json") != "1916-02" {
		t.Error(Month("monthly/news/1916-02.json"))
	}
}
