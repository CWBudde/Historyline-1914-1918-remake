package main

import (
	"io/fs"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/campaign"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// Every campaign map has its month's front page, and every page fits its
// sheet: headline and deck across it, caption and orders above the button,
// the article in its two columns above "In brief".
func TestFrontPagesFit(t *testing.T) {
	issues := map[string]int{}
	for _, m := range testShell(t).maps {
		a, ok := m.news()
		if !ok {
			t.Errorf("%s has no front page", m.name)
			continue
		}
		month := m.month
		issues[month] = m.issue
		p := layoutNews(a, m.issue, m.briefing)
		if !p.fits {
			t.Errorf("%s does not fit its page", month)
		}
		for _, l := range p.lines {
			w := measure(l.k)
			left := l.x - l.al*w
			if left < newsLeft-0.5 || left+w > newsRight+0.5 {
				t.Errorf("%s: %q runs off the sheet", month, l.k.s)
			}
			if col := l.x >= newsColB && l.y >= newsTop; col && l.x < newsColC && left+w > newsColB+newsColW {
				t.Errorf("%s: %q is wider than its column", month, l.k.s)
			}
		}
	}
	pages, _ := fs.Glob(levels.Levels, "*/"+campaign.File)
	if issues["1914-08"] != 1 || issues["1918-07"] != len(pages) || len(pages) != 48 {
		t.Errorf("issues %d, %d of %d", issues["1914-08"], issues["1918-07"], len(pages))
	}
}

// Each side meets a unit type once in its campaign chain: the whole force in
// its first month, then only what is new.
func TestNewUnitsFollowTheChain(t *testing.T) {
	s := testShell(t)
	byName := map[string]mapLevel{}
	for _, m := range s.maps {
		byName[m.name] = m
	}
	ids := func(name string) []string {
		var out []string
		for _, c := range byName[name].newUnits {
			out = append(out, c.id)
		}
		slices.Sort(out)
		return out
	}
	if got := ids("1917-12:cp"); !slices.Equal(got, []string{"CP_a7v", "CP_fokkerDVII", "CP_markIV"}) {
		t.Errorf("1917-12 cp new units %v", got)
	}
	if got := ids("1917-10:ec"); !slices.Equal(got, []string{"EC_camel"}) {
		t.Errorf("1917-10 ec new units %v", got)
	}
	if got := ids("1914-09:cp"); len(got) != 0 {
		t.Errorf("1914-09 cp new units %v", got)
	}
	if n := len(ids("1914-08:cp")); n < 10 {
		t.Errorf("1914-08 cp introduces %d unit types, want the whole force", n)
	}
	for _, m := range s.maps {
		for _, c := range m.newUnits {
			if c.family == "" || c.note == "" || c.name == "" {
				t.Errorf("%s: card %+v", m.name, c)
			}
			if n := len(wrapText(c.note, textKey{size: 10}, unitNoteW)); n > 4 {
				t.Errorf("%s: note of %s takes %d lines", c.id, c.family, n)
			}
			for _, l := range []textKey{{s: c.name, bold: true, size: 11}, {s: c.attack, size: 9.5}} {
				if w := measure(l); w > unitNoteX-84-6 {
					t.Errorf("%s: %q is %.0f px wide", c.id, l.s, w)
				}
			}
		}
	}
}

// The page of new units turns with NEXT PAGE, back with the back key, and
// TO BATTLE loads from any page.
func TestUnitPages(t *testing.T) {
	s := testShell(t)
	s.code = 0
	m := s.maps[0]
	pages := unitPages(m.newUnits)
	if pages < 2 {
		t.Fatalf("map %s has %d pages of new units", m.name, pages)
	}
	s.afterNews()
	for i := 1; i < pages; i++ {
		if s.page != pageUnits || s.buttons()[s.focus].label != "NEXT PAGE" {
			t.Fatalf("page %d, focus %d", s.page, s.focus)
		}
		s.activate(s.focus)
	}
	if s.unitPage != pages-1 || len(s.buttons()) != 1 {
		t.Fatalf("last page %d, buttons %d", s.unitPage, len(s.buttons()))
	}
	s.key(keyBack)
	if s.unitPage != pages-2 {
		t.Fatalf("back to page %d", s.unitPage)
	}
	toBattle(t, s)
}
