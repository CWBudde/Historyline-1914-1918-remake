package llmplay

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// TerrainCode is the 2-letter code of a terrain name in the static grid.
func TerrainCode(terrain string, infra level.Infra) string {
	for _, id := range infra.IDs {
		switch {
		case strings.HasPrefix(id, "road"):
			return "=="
		case strings.HasPrefix(id, "rail"):
			return "++"
		case strings.HasPrefix(id, "trench"):
			return "tt"
		}
	}
	t := strings.TrimSuffix(terrain, "_w")
	switch {
	case t == "grass" || t == "":
		return ".."
	case t == "forest":
		return "ff"
	case t == "water":
		return "~~"
	case t == "waterDeep":
		return "~D"
	case t == "waterSea":
		return "~S"
	case strings.HasPrefix(t, "lakeside"):
		return "ls"
	case t == "stone":
		return "st"
	case t == "mud":
		return "mu"
	case strings.HasPrefix(t, "mount"):
		return "^^"
	case t == "hills":
		return "hi"
	case strings.Contains(t, "hq"):
		return "HQ"
	case strings.Contains(t, "factory"):
		return "fy"
	case strings.Contains(t, "depot"):
		return "dp"
	case strings.Contains(t, "house") || strings.Contains(t, "building"):
		return "ho"
	default:
		if len(t) >= 2 {
			return t[:2]
		}
		return "??"
	}
}

func terrainAt(s *dosplay.Session, reworked map[hexgrid.Coord]level.Hex, c hexgrid.Coord) string {
	if h, ok := reworked[c]; ok {
		return h.Terrain
	}
	if c.Line < 0 || c.Line >= len(s.Level.Lines) || c.Col < 0 || c.Col >= len(s.Level.Lines[c.Line]) {
		return "offmap"
	}
	return s.Level.Lines[c.Line][c.Col].Terrain
}

// StaticMap is the one-time terrain preamble: a 2-char-per-hex grid with
// 1-based line numbers; odd file lines are indented one space (they sit
// half a hex right and hold Width-1 hexes).
func StaticMap(s *dosplay.Session) string {
	var b strings.Builder
	fmt.Fprintf(&b, "STATIC MAP %dx%d legend: .. grass ff forest ~~ water ~D deep ~S sea ls lakeside st stone mu mud ^^ mountain hi hills ho house == road ++ rail tt trench HQ/fy/dp building\n", s.Level.Grid().Width, s.Level.Grid().Lines)
	for l, line := range s.Level.Lines {
		fmt.Fprintf(&b, "%3d ", l+1)
		if l%2 == 1 {
			b.WriteString(" ")
		}
		for _, h := range line {
			b.WriteString(TerrainCode(h.Terrain, h.Infra))
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func unitLine(s *dosplay.Session, rew map[hexgrid.Coord]level.Hex, u dosplay.UnitView) string {
	state := "ready"
	if u.Moved {
		state = "ordered"
	}
	if u.Held {
		state += " held"
	}
	id := u.ID
	if id == "" {
		id = fmt.Sprintf("type%d", u.Type)
	}
	return fmt.Sprintf("u%d %s %d/%d e%d %s %s %s", u.Index, id, u.Size, u.Full, u.Exp, show(u.At), terrainAt(s, rew, u.At), state)
}

func sortedCoords[V any](m map[hexgrid.Coord]V) []hexgrid.Coord {
	out := slices.SortedFunc(maps.Keys(m), func(a, b hexgrid.Coord) int {
		if a.Line != b.Line {
			return a.Line - b.Line
		}
		return a.Col - b.Col
	})
	return out
}

// Observe renders the state side may fairly know (its View, never
// TrueView) plus this phase's legal orders. Attack targets carry exact
// Engage predictions and the dosai gain score, so the model decides on
// numbers, not on reading sprites. lastDigest is Digest of the events
// since the model's previous orders ("" at game start).
func Observe(s *dosplay.Session, side int, lastDigest string) string {
	rew := s.Reworked()
	v := s.View(side)
	st := s.Stats()
	var b strings.Builder
	fmt.Fprintf(&b, "GAME %s YOU=%s side%d TURN %d PHASE %s ENERGY %d (income %d/phase-change) ENEMY income %d SCORE %d\n",
		s.Level.MapName, Sides[side], side, s.Turn(), phaseName(s, side), s.Energy(side), st.Income[side], st.Income[1-side], st.Score)
	fmt.Fprintf(&b, "HOLDINGS you %d units/%d depots/%d factories, enemy %d/%d/%d, neutral %d/%d/%d\n",
		st.Sides[side].Units, st.Sides[side].Depots, st.Sides[side].Factories,
		st.Sides[1-side].Units, st.Sides[1-side].Depots, st.Sides[1-side].Factories,
		st.Sides[2].Units, st.Sides[2].Depots, st.Sides[2].Factories)
	if len(rew) > 0 {
		b.WriteString("REWORKED (overrides static map):")
		for _, c := range sortedCoords(rew) {
			fmt.Fprintf(&b, " %s=%s", show(c), TerrainCode(rew[c].Terrain, rew[c].Infra))
		}
		b.WriteByte('\n')
	}
	b.WriteString("BUILDINGS (entrance kind owner hold contents-if-yours):\n")
	for _, at := range sortedCoords(v.Buildings) {
		bv := v.Buildings[at]
		owner := Sides[s.SideOf(bv.Owner)]
		line := fmt.Sprintf(" %s %s %s hold%d", show(at), bv.Kind, owner, bv.Hold)
		if s.SideOf(bv.Owner) == side {
			_, slots, _ := s.Contents(side, at)
			var parts []string
			for _, u := range slots {
				parts = append(parts, fmt.Sprintf("slot%d u%d %s %d/%d", u.Slot, u.Index, u.ID, u.Size, u.Full))
			}
			line += " holds: " + strings.Join(parts, ", ")
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("YOUR UNITS:\n")
	var mine, enemy []dosplay.UnitView
	for _, at := range sortedCoords(v.Units) {
		u := v.Units[at]
		switch u.Side {
		case side:
			mine = append(mine, u)
		case 1 - side:
			enemy = append(enemy, u)
		}
	}
	slices.SortFunc(mine, func(a, b dosplay.UnitView) int { return a.Index - b.Index })
	for _, u := range mine {
		b.WriteString(" " + unitLine(s, rew, u) + "\n")
		if u.Carrier {
			if hold, slots, ok := s.Contents(side, u.At); ok {
				for _, c := range slots {
					fmt.Fprintf(&b, "  cargo of u%d: hold%d slot%d u%d %s %d/%d\n", u.Index, hold, c.Slot, c.Index, c.ID, c.Size, c.Full)
				}
			}
		}
	}
	b.WriteString("ENEMY LAST-KNOWN (positions stale since last phase change):\n")
	slices.SortFunc(enemy, func(a, b dosplay.UnitView) int { return a.Index - b.Index })
	for _, u := range enemy {
		fmt.Fprintf(&b, " e%d %s %d/%d %s\n", u.Index, u.ID, u.Size, u.Full, show(u.At))
	}
	if lastDigest != "" {
		b.WriteString("EVENTS SINCE YOUR LAST ORDERS:\n" + lastDigest)
	}
	b.WriteString(threats(s, side, v))
	b.WriteString(legal(s, side, v, rew))
	b.WriteString("Reply with JSONL orders only, one per line: {\"op\":\"move\",\"unit\":N,\"to\":[l,c]} | {\"op\":\"exit\",\"hold\":H,\"slot\":S,\"to\":[l,c]} | {\"op\":\"attack\",\"unit\":N,\"target\":[l,c]} | {\"op\":\"repair\",\"hold\":H,\"slot\":S} | {\"op\":\"build\",\"hold\":H,\"type\":T} | {\"op\":\"surrender\"}. Coordinates 1-based. Omitted units hold position. Whole batch is validated atomically.\n")
	return b.String()
}

// threats warns about the resolution semantics that kill careless
// movers: when the opponent is the action side, its battles are fought
// on the positions as they stand NOW — a unit ordered to move away this
// phase is still shot where it stands. It lists every enemy attack the
// exact Engage oracle predicts to destroy or halve one of side's units.
func threats(s *dosplay.Session, side int, v dosplay.View) string {
	opp := 1 - side
	if s.Phase(opp) != dosrules.PhaseAction {
		return ""
	}
	bf, err := s.Battlefield(opp)
	if err != nil {
		return ""
	}
	prof := s.Game.Profile
	var b strings.Builder
	for _, at := range sortedCoords(v.Units) {
		e := v.Units[at]
		if e.Side != opp {
			continue
		}
		for _, t := range dosplay.Coords(s.Targets(opp, e.Index)) {
			mine, ok := v.Units[t]
			if !ok || mine.Side != side {
				continue
			}
			cell, ok := s.Board.Cell(t)
			if !ok {
				continue
			}
			o, err := prof.Engage(bf, e.Index, cell)
			if err != nil || !o.Battle {
				continue
			}
			if slices.Contains(o.Destroyed, mine.Index) {
				fmt.Fprintf(&b, " THREAT e%d %s can DESTROY u%d %s at %s this resolution (moving away does not save it)\n", e.Index, e.ID, mine.Index, mine.ID, show(t))
			} else if o.DefenderSize*2 <= mine.Size && o.DefenderSize < mine.Size {
				fmt.Fprintf(&b, " THREAT e%d %s can halve u%d %s (%d->%d) at %s this resolution\n", e.Index, e.ID, mine.Index, mine.ID, mine.Size, o.DefenderSize, show(t))
			}
		}
	}
	if b.Len() == 0 {
		return ""
	}
	return "THREATS (opponent is action side; battles resolve on CURRENT positions):\n" + b.String()
}

// legal annotates this phase's legal orders, capped per unit.
func legal(s *dosplay.Session, side int, v dosplay.View, rew map[hexgrid.Coord]level.Hex) string {
	var b strings.Builder
	b.WriteString("LEGAL THIS PHASE:\n")
	if s.Phase(side) == dosrules.PhaseMovement {
		for _, at := range sortedCoords(v.Units) {
			u := v.Units[at]
			if u.Side != side || u.Moved || u.Held {
				continue
			}
			dests, err := s.Destinations(side, u.Index)
			if err != nil || len(dests) == 0 {
				continue
			}
			cs := dosplay.Coords(dests)
			var caps []string
			for _, c := range cs {
				if bv, ok := v.Buildings[c]; ok && s.SideOf(bv.Owner) != side {
					caps = append(caps, show(c)+" "+bv.Kind.String())
				}
			}
			fmt.Fprintf(&b, " u%d move-> %d hexes: %s", u.Index, len(cs), destSample(s, side, v, cs))
			if len(caps) > 0 {
				fmt.Fprintf(&b, " CAPTURE: %s", strings.Join(caps, " "))
			}
			b.WriteByte('\n')
		}
		for _, at := range sortedCoords(v.Buildings) {
			bv := v.Buildings[at]
			if s.SideOf(bv.Owner) != side {
				continue
			}
			hold, slots, ok := s.Contents(side, at)
			if !ok {
				continue
			}
			for _, u := range slots {
				dests, err := s.ExitDestinations(side, hold, u.Slot)
				if err != nil || len(dests) == 0 {
					continue
				}
				cs := dosplay.Coords(dests)
				var sample []string
				for _, c := range cs[:min(6, len(cs))] {
					sample = append(sample, show(c))
				}
				fmt.Fprintf(&b, " hold%d slot%d (%s) exit-> %d hexes: %s%s\n", hold, u.Slot, u.ID, len(cs), strings.Join(sample, " "), more(len(cs), 6))
			}
		}
		// Cargo exits from own transporters (ships, trains, trucks).
		for _, at := range sortedCoords(v.Units) {
			u := v.Units[at]
			if u.Side != side || !u.Carrier {
				continue
			}
			hold, slots, ok := s.Contents(side, at)
			if !ok {
				continue
			}
			for _, c := range slots {
				dests, err := s.ExitDestinations(side, hold, c.Slot)
				if err != nil || len(dests) == 0 {
					continue
				}
				cs := dosplay.Coords(dests)
				fmt.Fprintf(&b, " hold%d slot%d (%s, cargo of u%d) exit-> %d hexes: %s\n", hold, c.Slot, c.ID, u.Index, len(cs), destSample(s, side, v, cs))
			}
		}
		return b.String()
	}
	// Action phase: exact Engage per target + repair/build options.
	bf, err := s.Battlefield(side)
	prof := s.Game.Profile
	for _, at := range sortedCoords(v.Units) {
		u := v.Units[at]
		if u.Side != side || u.Moved || u.Held {
			continue
		}
		targets := dosplay.Coords(s.Targets(side, u.Index))
		if len(targets) == 0 {
			continue
		}
		var parts []string
		for _, t := range targets {
			tv, occupied := v.Units[t]
			if s.Sapper(u.Index) && !occupied {
				parts = append(parts, fmt.Sprintf("%s dig/work", show(t)))
				continue
			}
			if err == nil && occupied {
				if cell, ok := s.Board.Cell(t); ok {
					if o, e := prof.Engage(bf, u.Index, cell); e == nil && o.Battle {
						g := Gain(prof, bf, u.Index, o)
						killed := ""
						if slices.Contains(o.Destroyed, o.Defender) {
							killed = " KILL"
						}
						if slices.Contains(o.Destroyed, u.Index) {
							killed += " YOU-DIE"
						}
						parts = append(parts, fmt.Sprintf("%s %s def %d->%d you %d->%d gain %+d%s", show(t), tv.ID, tv.Size, o.DefenderSize, u.Size, o.AttackerSize, g, killed))
						continue
					}
				}
			}
			parts = append(parts, fmt.Sprintf("%s %s (unpredicted)", show(t), tv.ID))
		}
		fmt.Fprintf(&b, " u%d attack-> %s\n", u.Index, strings.Join(parts, "; "))
	}
	for _, at := range sortedCoords(v.Buildings) {
		bv := v.Buildings[at]
		if s.SideOf(bv.Owner) != side {
			continue
		}
		hold, slots, _ := s.Contents(side, at)
		for _, u := range slots {
			if u.Size < u.Full && s.CanRepair(side, hold, u.Slot) == nil {
				fmt.Fprintf(&b, " hold%d slot%d repair %s %d/%d -> full (cost 15)\n", hold, u.Slot, u.ID, u.Size, u.Full)
			}
		}
		for _, o := range s.Buildable(side, hold) {
			if s.CanBuild(side, hold, o.Type) == nil {
				fmt.Fprintf(&b, " hold%d build type%d %s %q cost %d\n", hold, o.Type, o.ID, o.Name, o.Cost)
			}
		}
	}
	return b.String()
}

// destSample lists destinations usefully: all when few, else the six
// closing most on the nearest enemy building (advance) plus the two
// opening most distance (retreat). Any other legal hex may still be
// ordered — Apply validates against the full set.
func destSample(s *dosplay.Session, side int, v dosplay.View, cs []hexgrid.Coord) string {
	showAll := func(list []hexgrid.Coord) string {
		parts := make([]string, len(list))
		for i, c := range list {
			parts[i] = show(c)
		}
		return strings.Join(parts, " ")
	}
	if len(cs) <= 12 {
		return showAll(cs)
	}
	var goals []int
	for at, bv := range v.Buildings {
		if s.SideOf(bv.Owner) != side {
			if cell, ok := s.Board.Cell(at); ok {
				goals = append(goals, cell)
			}
		}
	}
	prof, w := s.Game.Profile, s.Game.Width
	dist := func(c hexgrid.Coord) int {
		cell, ok := s.Board.Cell(c)
		if !ok || len(goals) == 0 {
			return 1 << 30
		}
		best := 1 << 30
		for _, g := range goals {
			if d := prof.Distance(w, cell, g); d < best {
				best = d
			}
		}
		return best
	}
	sorted := slices.Clone(cs)
	slices.SortFunc(sorted, func(a, b hexgrid.Coord) int { return dist(a) - dist(b) })
	advance := sorted[:6]
	retreat := sorted[len(sorted)-2:]
	return fmt.Sprintf("advance: %s retreat: %s (+%d more legal, any may be ordered)", showAll(advance), showAll(retreat), len(cs)-8)
}

func more(total, shown int) string {
	if total > shown {
		return fmt.Sprintf(" …+%d", total-shown)
	}
	return ""
}

// Gain is dosai's attack score (dosai.go gain, private there): armour-
// weighted strength the defender loses minus the attacker's, ±100 per
// unit destroyed. Positive gain favours the attacker.
func Gain(p dosrules.Profile, b *dosrules.Battlefield, attacker int, o dosrules.Outcome) int {
	a, ok := b.Units[attacker]
	if !ok {
		return 0
	}
	d := b.Units[o.Defender]
	ta, td := p.Tables.Types[a.Type], p.Tables.Types[d.Type]
	str := func(t dosrules.Type, size, alt int) int {
		if t.StrengthFlags&4 != 0 {
			return alt
		}
		return size
	}
	sc := td.Armour*(str(td, d.Size, d.Alt)-str(td, o.DefenderSize, o.DefenderAlt)) -
		ta.Armour*(str(ta, a.Size, a.Alt)-str(ta, o.AttackerSize, o.AttackerAlt))
	for _, k := range o.Destroyed {
		if k == attacker {
			sc -= 100
		} else {
			sc += 100
		}
	}
	return sc
}
