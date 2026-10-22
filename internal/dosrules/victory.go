package dosrules

// The end of a battle and its score. A side wins by capturing the other
// side's HQ (capture.go) or by destroying all of its units; the capture
// wins over an annihilation in the same resolution (0661:3d03); the score starts as
// the defence values of all units and falls as units disappear. The winner's
// final score adds 500 for annihilation and 100 for concealed inventories,
// multiplies by 4, 3 or 2 for a move limit of 4, 8 or 16 and is capped at
// 32,500.
//
// Assumed, not verified:
//   - A unit's defence value is its type's armour; the score counts every
//     unit not destroyed, of either side or neutral, in containers too.
//   - A side is annihilated when none of its units is left.
//   - Both sides annihilated at once is a draw without a score.
//   - The bonuses are added before the limit's factor.

// MaxScore caps the final score.
const MaxScore = 32500

// EndReason is how a side won.
type EndReason string

// The ways to win.
const (
	EndHQ           EndReason = "hq"
	EndAnnihilation EndReason = "annihilation"
)

// End is the state of a battle's end: whether it is over, the winner (-1
// while it is open or after a draw), how each side won, and the winner's
// final score.
type End struct {
	Over   bool         `json:"over"`
	Winner int          `json:"winner"`
	Won    [2]EndReason `json:"won"`
	Score  int          `json:"score"`
}

// Score is the current score: the armour of every unit not destroyed.
func (g *Game) Score() int {
	s := 0
	for _, u := range g.Units {
		if !u.Destroyed() {
			s += g.Profile.Tables.Types[u.Type].Armour
		}
	}
	return s
}

// end is the battle's end as the state stands; Resolve decides it, with hq
// the side that captured an HQ in the resolution (-1 none).
func (g *Game) end(hq int) End {
	r := End{Winner: -1}
	for k := range 2 {
		if g.annihilated(1 - k) {
			r.Won[k] = EndAnnihilation
		}
	}
	if hq >= 0 {
		r.Won = [2]EndReason{}
		r.Won[hq] = EndHQ
	}
	switch {
	case r.Won[0] != "" && r.Won[1] != "":
		r.Over = true
	case r.Won[0] != "" || r.Won[1] != "":
		r.Over, r.Winner = true, 0
		if r.Won[0] == "" {
			r.Winner = 1
		}
		r.Score = finalScore(g.Score(), r.Won[r.Winner] == EndAnnihilation, g.Concealed, g.Limit)
	}
	return r
}

// annihilated reports whether side has no unit left.
func (g *Game) annihilated(side int) bool {
	for _, u := range g.Units {
		if Owns(u.Flags4, side) && !u.Destroyed() {
			return false
		}
	}
	return true
}

// finalScore is the winner's score from the current one.
func finalScore(score int, annihilation, concealed bool, limit int) int {
	if annihilation {
		score += 500
	}
	if concealed {
		score += 100
	}
	switch limit {
	case 4:
		score *= 4
	case 8:
		score *= 3
	case 16:
		score *= 2
	}
	return min(score, MaxScore)
}
