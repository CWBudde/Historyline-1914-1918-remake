// Package campaign holds the campaign's front pages. Before each battle of a
// campaign month the player reads the month's front page of a neutral
// newspaper, newly written: the month's war in a few paragraphs, an
// illustration and a column of short items. A page is a file in the
// month's level folder, levels/1914-08/news.json.
package campaign

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path"
)

// File is the front page's file in a month's level folder (levels.Levels).
const File = "news.json"

// Article is the front page of a campaign month.
type Article struct {
	Label    string   `json:"label"`    // the issue's month: "August 1914"
	Headline string   `json:"headline"` // the page's headline
	Deck     string   `json:"deck"`     // the line under the headline
	Caption  string   `json:"caption"`  // under the illustration, art/screens/news/<month>.png
	Body     []string `json:"body"`     // the article's paragraphs
	Briefs   []string `json:"briefs"`   // the "In brief" column: news from elsewhere
}

// Month is the month of a front page file: "1914-08/news.json" is
// "1914-08".
func Month(file string) string {
	return path.Base(path.Dir(file))
}

// Parse reads a front page file; every part of the page must be there.
func Parse(raw []byte) (Article, error) {
	var a Article
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&a); err != nil {
		return a, err
	}
	switch {
	case a.Label == "", a.Headline == "", a.Deck == "", a.Caption == "":
		return a, errors.New("front page: label, headline, deck and caption are required")
	case len(a.Body) == 0, len(a.Briefs) == 0:
		return a, fmt.Errorf("front page %q: %d paragraphs, %d briefs", a.Label, len(a.Body), len(a.Briefs))
	}
	return a, nil
}
