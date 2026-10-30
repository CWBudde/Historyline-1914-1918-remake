package campaign

import (
	"io/fs"
	"path"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

func TestEveryMonthHasAFrontPage(t *testing.T) {
	maps, err := fs.Glob(levels.Levels, "19*/map.json")
	if err != nil || len(maps) == 0 {
		t.Fatal(maps, err)
	}
	pages, err := fs.Glob(levels.Levels, "*/"+File)
	if err != nil || len(pages) != len(maps) {
		t.Fatalf("%d front pages for %d campaign months (%v)", len(pages), len(maps), err)
	}
	for _, m := range maps {
		month := path.Dir(m)
		raw, err := fs.ReadFile(levels.Levels, month+"/"+File)
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
	if Month("1916-02/news.json") != "1916-02" {
		t.Error(Month("1916-02/news.json"))
	}
}
