package main

import (
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/settings"
)

// The game values of the menus (PLAY, BATTLES, PLAYER B, MOVES and
// SCROLLING) are kept between runs in settings.json (internal/settings).
// A value changed in the menus is kept at once; a flag's value holds for its
// run only.

// values are the menus' game values as settings keeps them.
func (s *shell) values() settings.Values {
	v := settings.Values{
		PlayMode: "single", Reduced: !s.fights, Opponent: "computer",
		Limit: moveLimits[s.limit], Scroll: strings.ToLower(scrollSpeeds[s.scroll].name),
	}
	if s.play != localNone {
		v.PlayMode = s.play.String()
	}
	if !s.computer {
		v.Opponent = "passive"
	}
	return v
}

// setValues makes v the menus' game values; a value the menus do not have
// leaves its default.
func (s *shell) setValues(v settings.Values) {
	s.fights, s.computer = !v.Reduced, v.Opponent != "passive"
	if m, err := parseLocalMode(v.PlayMode); err == nil {
		s.play = m
	}
	s.limit = max(slices.Index(moveLimits, v.Limit), 0)
	s.scroll = 0
	for i, sp := range scrollSpeeds {
		if strings.EqualFold(sp.name, v.Scroll) {
			s.scroll = i
		}
	}
}

// change runs fn, which changes game values in the menus, and keeps the
// values it changed; a value a flag set for the run stays unkept unless fn
// changed it. A store that cannot be written keeps the values for the run.
func (s *shell) change(fn func()) {
	before := s.values()
	fn()
	after := s.values()
	k := &s.kept
	if after.PlayMode != before.PlayMode {
		k.PlayMode = after.PlayMode
	}
	if after.Reduced != before.Reduced {
		k.Reduced = after.Reduced
	}
	if after.Opponent != before.Opponent {
		k.Opponent = after.Opponent
	}
	if after.Limit != before.Limit {
		k.Limit = after.Limit
	}
	if after.Scroll != before.Scroll {
		k.Scroll = after.Scroll
	}
	if s.settings != nil {
		_ = s.settings.Save(s.kept) //nolint:errcheck // An unwritable store keeps the values for this run.
	}
}

// setPlay sets PLAY to m and keeps it.
func (s *shell) setPlay(m localMode) { s.change(func() { s.play = m }) }
