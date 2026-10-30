package llmplay

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// Blunder classes Review counts. Detection is by exact re-simulation:
// the transcript is replayed command by command and every model attack
// is re-scored with the same Engage+Gain oracle the prompt annotated —
// a "lesson" is only ever a counted pattern, never the model's claim.
const (
	BlunderLosingAttack = "losing-attack" // attack ordered with predicted gain <= 0
	BlunderUnitLost     = "unit-lost"     // a model unit was destroyed in a battle
	BlunderIdleEconomy  = "idle-economy"  // action phase ended with >= 35 energy and no build
)

// Finding is one counted mistake.
type Finding struct {
	Resolution int
	Class      string
	Detail     string
}

// Report is the outcome of Review.
type Report struct {
	Game        string
	End         *dosrules.End
	Resolutions int
	Findings    []Finding
	Counts      map[string]int
}

// Review replays a recorded game, verifying it hash-for-hash (via the
// same replay), and re-scores the model's commands. It writes
// review.md into the game dir and, when lessonsPath != "", appends
// lessons for classes seen at least twice.
func Review(dir, lessonsPath string) (*Report, error) {
	h, lines, err := ReadTranscript(dir)
	if err != nil {
		return nil, err
	}
	month, mode, err := ParseLevelRef(h.Level)
	if err != nil {
		return nil, err
	}
	lvl, err := ShipLevel(month, mode)
	if err != nil {
		return nil, err
	}
	s, err := NewSession(lvl, h.LLMSide)
	if err != nil {
		return nil, err
	}
	rep := &Report{Game: filepath.Base(dir), Counts: map[string]int{}}
	add := func(res int, class, detail string) {
		rep.Findings = append(rep.Findings, Finding{res, class, detail})
		rep.Counts[class]++
	}
	prof := s.Game.Profile
	for _, l := range lines {
		rep.Resolutions = l.Resolution
		for _, c := range l.Commands {
			// Re-score the model's attacks on the state before them.
			if c.Kind == dosrules.CmdAttack && c.Side == h.LLMSide {
				if bf, berr := s.Battlefield(h.LLMSide); berr == nil {
					if o, e := prof.Engage(bf, c.Unit, c.Cell); e == nil && o.Battle {
						if g := Gain(prof, bf, c.Unit, o); g <= 0 {
							add(l.Resolution, BlunderLosingAttack,
								fmt.Sprintf("u%d attacked cell %d with predicted gain %+d (def->%d, atk->%d)", c.Unit, c.Cell, g, o.DefenderSize, o.AttackerSize))
						}
					}
				}
			}
			evs, rerr := s.Replay([]dosrules.Command{c})
			if rerr != nil {
				return nil, fmt.Errorf("resolution %d: %w", l.Resolution, rerr)
			}
			for _, ev := range evs {
				if ev.Battle == nil {
					continue
				}
				for _, f := range []struct {
					side int
					u    int
					dead bool
				}{
					{ev.Battle.Attacker.Unit.Side, ev.Battle.Attacker.Unit.Index, ev.Battle.Attacker.Destroyed},
					{ev.Battle.Defender.Unit.Side, ev.Battle.Defender.Unit.Index, ev.Battle.Defender.Destroyed},
				} {
					if f.side == h.LLMSide && f.dead {
						add(l.Resolution, BlunderUnitLost, fmt.Sprintf("u%d destroyed: %s", f.u, battleStr(ev.Battle)))
					}
				}
			}
		}
		got, err := StateHash(s)
		if err != nil {
			return nil, err
		}
		if got != l.StateAfter {
			return nil, fmt.Errorf("resolution %d: replay diverged from transcript", l.Resolution)
		}
		// Energy here is post-income, so only a fat bank (>= 80: any
		// build was affordable) with neither build nor repair counts.
		if l.PhaseLLM == "action" && s.Energy(h.LLMSide) >= 80 && !l.Fallback {
			spent := false
			for _, c := range l.Commands {
				if (c.Kind == dosrules.CmdBuild || c.Kind == dosrules.CmdRepair) && c.Side == h.LLMSide {
					spent = true
				}
			}
			if !spent {
				add(l.Resolution, BlunderIdleEconomy,
					fmt.Sprintf("action phase ended with %d energy banked and no build/repair ordered", s.Energy(h.LLMSide)))
			}
		}
	}
	rep.End = s.Ended()
	if err := rep.Write(filepath.Join(dir, "review.md")); err != nil {
		return rep, err
	}
	if lessonsPath != "" {
		if err := rep.AppendLessons(lessonsPath); err != nil {
			return rep, err
		}
	}
	return rep, nil
}

// Write renders the per-resolution review table.
func (r *Report) Write(path string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# Review of %s\n\n", r.Game)
	if r.End != nil {
		fmt.Fprintf(&b, "Ended: winner side %d, won %v, score %d after %d resolutions.\n\n", r.End.Winner, r.End.Won, r.End.Score, r.Resolutions)
	} else {
		fmt.Fprintf(&b, "Game open after %d resolutions.\n\n", r.Resolutions)
	}
	b.WriteString("| Resolution | Class | Detail |\n|---|---|---|\n")
	for _, f := range r.Findings {
		fmt.Fprintf(&b, "| %d | %s | %s |\n", f.Resolution, f.Class, f.Detail)
	}
	b.WriteString("\nCounts:")
	classes := make([]string, 0, len(r.Counts))
	for c := range r.Counts {
		classes = append(classes, c)
	}
	sort.Strings(classes)
	for _, c := range classes {
		fmt.Fprintf(&b, " %s=%d", c, r.Counts[c])
	}
	b.WriteString("\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

var lessonText = map[string]string{
	BlunderLosingAttack: "Do not order attacks whose annotated Engage gain is <= 0 unless the target is a capturer or artillery and you state why.",
	BlunderUnitLost:     "A unit was destroyed in a battle the exact predictor could have foreseen: check every destination and attack against the gain annotations, and keep damaged units out of double-envelopes.",
	BlunderIdleEconomy:  "Banked energy wins nothing: repair (flat 15, full strength) or build a capturer instead of ending action phases with a fat bank.",
}

// AppendLessons appends one dated, evidence-linked bullet per class
// counted at least twice. The game preamble re-reads this file.
func (r *Report) AppendLessons(path string) error {
	var bullets []string
	classes := make([]string, 0, len(r.Counts))
	for c := range r.Counts {
		classes = append(classes, c)
	}
	sort.Strings(classes)
	for _, c := range classes {
		if r.Counts[c] < 2 {
			continue
		}
		text, ok := lessonText[c]
		if !ok {
			continue
		}
		var at []int
		for _, f := range r.Findings {
			if f.Class == c {
				at = append(at, f.Resolution)
			}
		}
		bullets = append(bullets, fmt.Sprintf("- %s — %s (evidence: %s resolutions %v, %s, counted %dx)\n",
			c, text, r.Game, at, time.Now().Format("2006-01-02"), r.Counts[c]))
	}
	if len(bullets) == 0 {
		return nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString(strings.Join(bullets, ""))
	return err
}

// Lessons reads the top lessons (at most n bullets) for a game preamble.
func Lessons(path string, n int) string {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var out []string
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "- ") {
			out = append(out, line)
		}
	}
	if len(out) > n {
		out = out[len(out)-n:]
	}
	if len(out) == 0 {
		return ""
	}
	return "LESSONS FROM EARLIER GAMES (counted patterns, obey them):\n" + strings.Join(out, "\n") + "\n"
}
