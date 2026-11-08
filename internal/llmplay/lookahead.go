package llmplay

import (
	"maps"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosai"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Lookahead is the Coach with a plan search, a chess engine's search and
// evaluation over postures instead of single moves (a phase moves every
// unit, far too many moves to search one by one). Before each movement
// phase it plays every posture Horizon resolutions ahead on a clone,
// against dosai.Planner (the real opponent: dosai is deterministic, so the
// rollout is exactly what would happen while the posture is kept), values
// the end position and plays the best posture's phase. It is an oracle
// player: it knows the opponent's replies.
type Lookahead struct {
	Config  *CoachConfig // the base doctrine; nil: DefaultCoachConfig
	Horizon int          // resolutions played ahead; 0: 8
	Exact   bool         // rollouts replan with the coach's own oracle (3x slower)
	Wide    bool         // also the widePostures
	// Trace, if set, gets each decision: the posture values and the pick.
	Trace func(turn int, values []int, pick int)

	c       Coach
	list    []posture
	cfgs    []*CoachConfig
	posture int
}

// posture is a change to the base doctrine the search may pick.
type posture struct {
	name  string
	apply func(*CoachConfig)
}

// postures are the plans Lookahead chooses between; the first is the
// plain coach.
var postures = []posture{
	{"coach", func(*CoachConfig) {}},
	{"cautious", func(c *CoachConfig) { c.Caution = 300 }},
	{"bold", func(c *CoachConfig) { c.Caution = 30 }},
	{"siege", func(c *CoachConfig) { c.Siege, c.Caution = 1, 50 }},
	{"assault", func(c *CoachConfig) { c.Siege = 2 }},
	{"turtle", func(c *CoachConfig) { c.Turtle, c.Dash = true, false }},
}

// widePostures are further plans a Wide Lookahead considers: the foot
// capturers take buildings instead of racing for the enemy HQ.
var widePostures = []posture{
	{"grab", func(c *CoachConfig) { c.Dash = false }},
	{"grab-cautious", func(c *CoachConfig) { c.Dash, c.Caution = false, 300 }},
}

// keepBonus favours the posture in play: switching on a hair's difference
// makes the army march back and forth.
const keepBonus = 40

// PlayPhase implements PhasePlayer.
func (l *Lookahead) PlayPhase(s *dosplay.Session, side int) ([]dosplay.Event, error) {
	if s.Ended() != nil || s.Ready(side) {
		return nil, nil
	}
	if l.cfgs == nil {
		base := l.Config
		if base == nil {
			base = DefaultCoachConfig()
		}
		l.list = postures
		if l.Wide {
			l.list = append(slices.Clone(postures), widePostures...)
		}
		for _, p := range l.list {
			c := *base
			p.apply(&c)
			l.cfgs = append(l.cfgs, &c)
		}
	}
	if s.Phase(side) == dosrules.PhaseMovement {
		l.choose(s, side)
	}
	l.c.Config = l.cfgs[l.posture]
	return l.c.PlayPhase(s, side)
}

// Posture is the name of the posture in play.
func (l *Lookahead) Posture() string { return l.list[l.posture].name }

func (l *Lookahead) choose(s *dosplay.Session, side int) {
	values := make([]int, len(l.list))
	best, bestV := l.posture, 0
	found := false
	for i := range l.list {
		v, ok := l.rollout(s, side, i)
		if !ok {
			values[i] = -1 << 40
			continue
		}
		if i == l.posture {
			v += keepBonus
		}
		values[i] = v
		if !found || v > bestV {
			best, bestV, found = i, v, true
		}
	}
	l.posture = best
	if l.Trace != nil {
		l.Trace(s.Turn(), values, best)
	}
}

// rollout plays posture i Horizon resolutions ahead on a clone and values
// the end; ok=false when the rollout could not be played.
func (l *Lookahead) rollout(s *dosplay.Session, side, i int) (int, bool) {
	clone, err := s.Clone()
	if err != nil {
		return 0, false
	}
	cc := Coach{Config: l.cfgs[i], stickyReinforce: maps.Clone(l.c.stickyReinforce), seal: l.c.seal, sealDone: l.c.sealDone, quick: !l.Exact}
	opp := &dosai.Planner{}
	h := l.Horizon
	if h <= 0 {
		h = 8
	}
	r := 0
	for ; r < h && clone.Ended() == nil; r++ {
		if _, err := cc.PlayPhase(clone, side); err != nil {
			return 0, false
		}
		if _, err := opp.Plan(clone, 1-side); err != nil {
			return 0, false
		}
		if _, err := clone.Resolve(); err != nil {
			return 0, false
		}
	}
	return Evaluate(clone, side, r), true
}

// Evaluate values the position for side after r resolutions of a rollout:
// a win or loss outweighs everything (sooner is better), else material
// (each unit's cost at its strength), buildings, and capturers at an HQ —
// the enemy's at ours count against, ours at the enemy's for.
func Evaluate(s *dosplay.Session, side, r int) int {
	if end := s.Ended(); end != nil {
		switch end.Winner {
		case side:
			return 1<<30 - r*1000
		case 1 - side:
			return -1<<30 + r*1000
		}
	}
	prof := s.Game.Profile
	cost := func(typ int) int {
		if c := prof.Tables.Types[typ].Cost; c < 255 {
			return c
		}
		return 70
	}
	mv := prof.Tables.Move
	captures := func(typ int) bool {
		return mv != nil && typ < len(mv.Types) && mv.Types[typ].Flags&captureFlag != 0
	}
	v := s.View(side)
	var hqs [2][]hexgrid.Coord
	for at, b := range v.Buildings {
		if o := s.SideOf(b.Owner); b.HQ && (o == 0 || o == 1) {
			hqs[o] = append(hqs[o], at)
		}
	}
	val := 0
	var onMap [2]int
	for _, u := range v.Units {
		if u.Head != nil || u.Side != 0 && u.Side != 1 {
			continue
		}
		onMap[u.Side]++
		w := cost(u.Type) * u.Size / max(u.Full, 1)
		if captures(u.Type) {
			// A capturer next to the other side's HQ is a threat of the
			// game's end: worth more the closer it stands.
			for _, hq := range hqs[1-u.Side] {
				if d := hexgrid.Distance(u.At, hq); d <= 4 {
					w += 60 * (5 - d)
				}
			}
		}
		if u.Side == side {
			val += w
		} else {
			val -= w
		}
	}
	st := s.Stats()
	for sd := range 2 {
		h := st.Sides[sd]
		w := 30*(h.Units-onMap[sd]) + 40*h.Factories + 15*h.Depots // held units, buildings
		if sd == side {
			val += w
		} else {
			val -= w
		}
	}
	return val
}
