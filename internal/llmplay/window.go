package llmplay

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Window narrows a prompt to the fronts the player is deciding on. Lines
// keeps only the units, buildings, threats, orders and events on the
// 1-based map lines of its spans (header and holdings stay whole); Full
// lists every legal destination instead of a sample. The zero Window is
// the whole prompt.
type Window struct {
	Lines [][2]int
	Full  bool
	// Brief drops what a player has seen once: the order format and the
	// destination legend, and keeps events short (BriefDigest).
	Brief bool
}

// ParseLines reads a line window such as "1-20,75-95" or "8".
func ParseLines(spec string) ([][2]int, error) {
	var out [][2]int
	for part := range strings.SplitSeq(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		lo, hi, isSpan := strings.Cut(part, "-")
		a, err := strconv.Atoi(lo)
		if err != nil {
			return nil, fmt.Errorf("line window %q: %w", part, err)
		}
		b := a
		if isSpan {
			if b, err = strconv.Atoi(hi); err != nil {
				return nil, fmt.Errorf("line window %q: %w", part, err)
			}
		}
		if a < 1 || b < a {
			return nil, fmt.Errorf("line window %q: want 1 <= from <= to", part)
		}
		out = append(out, [2]int{a, b})
	}
	return out, nil
}

// Keeps reports whether hex c lies in the window.
func (w Window) Keeps(c hexgrid.Coord) bool {
	if len(w.Lines) == 0 {
		return true
	}
	for _, s := range w.Lines {
		if line := c.Line + 1; line >= s[0] && line <= s[1] {
			return true
		}
	}
	return false
}

func (w Window) header() string {
	if len(w.Lines) == 0 {
		return ""
	}
	parts := make([]string, len(w.Lines))
	for i, s := range w.Lines {
		parts[i] = fmt.Sprintf("%d-%d", s[0], s[1])
	}
	return "WINDOW lines " + strings.Join(parts, ",") + ": units, buildings, threats, orders and events elsewhere are omitted (omitted units hold position)\n"
}

var hexRef = regexp.MustCompile(`\[(\d+),(\d+)\]`)

// digest keeps the event lines that name no hex or a hex in the window.
func (w Window) digest(d string) string {
	if len(w.Lines) == 0 {
		return d
	}
	var b strings.Builder
	for line := range strings.SplitAfterSeq(d, "\n") {
		refs := hexRef.FindAllStringSubmatch(line, -1)
		keep := len(refs) == 0
		for _, r := range refs {
			l, errL := strconv.Atoi(r[1])
			c, errC := strconv.Atoi(r[2])
			if errL == nil && errC == nil && w.Keeps(hexgrid.Coord{Line: l - 1, Col: c - 1}) {
				keep = true
				break
			}
		}
		if keep {
			b.WriteString(line)
		}
	}
	return b.String()
}

// gun is an enemy long-range piece: it fires without moving, so its reach
// now is where a mover ends up under fire next enemy action phase.
type gun struct {
	at                 hexgrid.Coord
	air, ground, water dosplay.Weapon
}

// destNotes annotates a destination for unit u: "!" inside the current
// reach of an enemy long-range gun against u's domain, ">uN" boarding
// the own carrier uN standing there.
type destNotes func(u dosplay.UnitView, c hexgrid.Coord) string

func newDestNotes(s *dosplay.Session, side int, v dosplay.View) destNotes {
	var guns []gun
	types := s.Game.Profile.Tables.Types
	for _, e := range v.Units {
		if e.Side != 1-side || e.Type < 0 || e.Type >= len(types) || types[e.Type].TargetFlags&longRangeFlag == 0 {
			continue
		}
		if tv, ok := s.TypeInfo(e.Side, e.Type); ok {
			guns = append(guns, gun{at: e.At, air: tv.Air, ground: tv.Ground, water: tv.Water})
		}
	}
	return func(u dosplay.UnitView, c hexgrid.Coord) string {
		note := ""
		for _, g := range guns {
			w := g.ground
			switch u.Domain {
			case dosplay.Air:
				w = g.air
			case dosplay.Naval:
				w = g.water
			}
			if w.Power > 0 && hexgrid.Distance(c, g.at) <= w.Range {
				note = "!"
				break
			}
		}
		if o, ok := v.Units[c]; ok && o.Side == side && o.Carrier && o.Index != u.Index {
			note += fmt.Sprintf(">u%d", o.Index)
		}
		return note
	}
}
