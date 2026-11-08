package llmplay

import (
	"cmp"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// ForkPoint is a recorded game on a target and the resolution where the
// recorded side stood best (most units over the enemy's, the latest such
// resolution): a start for a played game (Fork) that skips the opening.
type ForkPoint struct {
	Dir         string
	Driver      string
	Resolutions int    // recorded
	Result      string // "won", "lost" or "open" (no end recorded)
	Units       [2]int // the recorded side's and the enemy's at the end
	At          int    // the best resolution
	AtUnits     [2]int // and the units there
}

// ForkPoints lists the games below root/*/ (the sweep folders) whose header
// names target ref, best margin first, each game once. They are not
// replayed: a game made before a dosai change no longer forks.
func ForkPoints(root, ref string) ([]ForkPoint, error) {
	name := strings.NewReplacer(":", "-").Replace(ref)
	dirs, err := filepath.Glob(filepath.Join(root, "*", name+"-*"))
	if err != nil {
		return nil, err
	}
	var out []ForkPoint
	seen := map[string]bool{} // a game swept twice is listed once
	for _, dir := range dirs {
		h, lines, err := ReadTranscript(dir)
		if err != nil || h.Level != ref || len(lines) == 0 || seen[lines[len(lines)-1].StateAfter] {
			continue
		}
		last := lines[len(lines)-1]
		seen[last.StateAfter] = true
		p := ForkPoint{Dir: dir, Driver: h.PlayedBy(), Resolutions: len(lines), Result: "open", Units: [2]int{last.UnitsLLM, last.UnitsOpponent}}
		for _, ev := range last.Events {
			if ev.Kind != "ended" {
				continue
			}
			switch {
			case strings.Contains(ev.End, "winner="+strconv.Itoa(h.LLMSide)):
				p.Result = "won"
			case strings.Contains(ev.End, "winner="):
				p.Result = "lost"
			}
		}
		best := 0
		for i, l := range lines {
			if i == 0 || l.UnitsLLM-l.UnitsOpponent >= best {
				best, p.At, p.AtUnits = l.UnitsLLM-l.UnitsOpponent, l.Resolution, [2]int{l.UnitsLLM, l.UnitsOpponent}
			}
		}
		out = append(out, p)
	}
	slices.SortStableFunc(out, func(a, b ForkPoint) int {
		return cmp.Or(cmp.Compare(b.AtUnits[0]-b.AtUnits[1], a.AtUnits[0]-a.AtUnits[1]), cmp.Compare(b.At, a.At))
	})
	return out, nil
}
