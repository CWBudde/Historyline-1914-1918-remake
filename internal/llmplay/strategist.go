package llmplay

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosai"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Strategist is the Lookahead searching standing plans as well as
// postures. Before each movement phase it lists the plans the position
// offers (every posture, and for each foreign building in reach a squad
// of the nearest capturers sent to take it, alone or with fighters),
// plays each Horizon resolutions ahead on a clone against dosai.Planner,
// values the end with EvaluatePlan and plays the best plan's phase. A
// posture moves the whole army one way; a plan can send a few units for a
// far building while the rest holds, which no posture does. Like the
// Lookahead it is an oracle player. The rollouts run side by side.
type Strategist struct {
	Config  *CoachConfig // the base doctrine; nil: DefaultCoachConfig
	Horizon int          // resolutions played ahead; 0: 16
	// Trace, if set, gets each decision: the plans' names and values and the pick.
	Trace func(turn int, names []string, values []int, pick int)

	c       Coach
	current string // the plan in play, by name
}

const (
	strategistHorizon = 16
	objectiveReach    = 8 // movement phases a squad may need for its building
	objectiveMax      = 6 // buildings weighed per decision
	squadSize         = 3
	escortSize        = 3
)

// plan is a candidate of the search.
type plan struct {
	name string
	d    *Directives
}

// EvaluatePlan is Evaluate with what decides a long game beside the
// material: the income each side has from here on and the energy in hand.
// dosai turns income into units every turn, so a position that is even in
// units but behind in income is lost in the long run.
func EvaluatePlan(s *dosplay.Session, side, r int) int {
	v := Evaluate(s, side, r)
	if s.Ended() != nil {
		return v
	}
	st := s.Stats()
	return v + 10*(st.Income[side]-st.Income[1-side]) + (st.Energy[side]-st.Energy[1-side])/2
}

// PlayPhase implements PhasePlayer.
func (l *Strategist) PlayPhase(s *dosplay.Session, side int) ([]dosplay.Event, error) {
	if s.Ended() != nil || s.Ready(side) {
		return nil, nil
	}
	base := l.Config
	if base == nil {
		base = DefaultCoachConfig()
	}
	if s.Phase(side) == dosrules.PhaseMovement || l.c.Config == nil {
		cfg, err := l.choose(s, side, base)
		if err != nil {
			return nil, err
		}
		l.c.Config = cfg
	}
	return l.c.PlayPhase(s, side)
}

// Plan is the name of the plan in play.
func (l *Strategist) Plan() string { return l.current }

// OrdersText names the plan in the transcript.
func (l *Strategist) OrdersText() string { return "plan " + l.current }

// candidates are the plans the position offers, the postures first.
func (l *Strategist) candidates(s *dosplay.Session, side int, base *CoachConfig) []plan {
	var out []plan
	for _, p := range slices.Concat(postures, widePostures) {
		out = append(out, plan{p.name, &Directives{Posture: p.name}})
	}
	p := newCoachPlan(s, side, nil, base)
	var squadPool, escortPool []dosplay.UnitView
	for _, u := range p.own {
		switch {
		case u.Held || u.Budget == 0 || u.Carrier || u.Domain == dosplay.Air || u.Domain == dosplay.Naval:
		case p.canCapture(u):
			squadPool = append(squadPool, u)
		case !p.longRange(u) && !p.s.Sapper(u.Index) && !p.s.Builder(u.Index):
			escortPool = append(escortPool, u)
		}
	}
	type objective struct {
		at      hexgrid.Coord
		eta     int
		squad   []int
		escorts []int
	}
	var objs []objective
	for _, at := range p.foreign {
		type cand struct {
			idx, eta int
		}
		var near []cand
		fields := map[int]map[hexgrid.Coord]int{}
		for _, u := range squadPool {
			f, ok := fields[u.Type]
			if !ok {
				costs, err := p.costsFor(u)
				if err != nil {
					continue
				}
				f = dijkstra(costs, map[hexgrid.Coord]int{at: 0})
				fields[u.Type] = f
			}
			if d, ok := f[u.At]; ok {
				near = append(near, cand{u.Index, (d + u.Budget - 1) / u.Budget})
			}
		}
		slices.SortFunc(near, func(a, b cand) int { return cmp.Or(a.eta-b.eta, a.idx-b.idx) })
		if len(near) == 0 || near[0].eta > objectiveReach {
			continue
		}
		o := objective{at: at, eta: near[0].eta}
		for _, n := range near[:min(len(near), squadSize)] {
			o.squad = append(o.squad, n.idx)
		}
		esc := slices.Clone(escortPool)
		slices.SortFunc(esc, func(a, b dosplay.UnitView) int {
			return cmp.Or(hexgrid.Distance(a.At, at)-hexgrid.Distance(b.At, at), a.Index-b.Index)
		})
		for _, u := range esc[:min(len(esc), escortSize)] {
			o.escorts = append(o.escorts, u.Index)
		}
		objs = append(objs, o)
	}
	slices.SortFunc(objs, func(a, b objective) int {
		return cmp.Or(a.eta-b.eta, a.at.Line-b.at.Line, a.at.Col-b.at.Col)
	})
	for _, o := range objs[:min(len(objs), objectiveMax)] {
		to := &[2]int{o.at.Line + 1, o.at.Col + 1}
		squad := Group{Units: o.squad, Goal: dirCapture, To: to}
		out = append(out, plan{"take " + show(o.at), &Directives{Groups: []Group{squad}}})
		if len(o.escorts) > 0 {
			out = append(out, plan{"storm " + show(o.at), &Directives{Groups: []Group{
				squad, {Units: o.escorts, Goal: dirGoto, To: to, Radius: 2},
			}}})
		}
	}
	return out
}

// choose plays every candidate ahead and returns the best one's
// configuration.
func (l *Strategist) choose(s *dosplay.Session, side int, base *CoachConfig) (*CoachConfig, error) {
	cands := l.candidates(s, side, base)
	cfgs := make([]*CoachConfig, len(cands))
	values := make([]int, len(cands))
	oks := make([]bool, len(cands))
	var wg sync.WaitGroup
	for i, c := range cands {
		coach, err := c.d.Coach(base)
		if err != nil {
			return nil, err
		}
		cfgs[i] = coach.Config
		wg.Add(1)
		go func() {
			defer wg.Done()
			values[i], oks[i] = l.rollout(s, side, cfgs[i])
		}()
	}
	wg.Wait()
	best := -1
	names := make([]string, len(cands))
	for i, c := range cands {
		names[i] = c.name
		if !oks[i] {
			values[i] = -1 << 40
			continue
		}
		if c.name == l.current {
			values[i] += keepBonus
		}
		if best < 0 || values[i] > values[best] {
			best = i
		}
	}
	if best < 0 {
		return nil, fmt.Errorf("llmplay: no plan could be played ahead")
	}
	l.current = cands[best].name
	if l.Trace != nil {
		l.Trace(s.Turn(), names, values, best)
	}
	return cfgs[best], nil
}

// rollout plays the configuration Horizon resolutions ahead on a clone
// and values the end; ok=false when the rollout could not be played.
func (l *Strategist) rollout(s *dosplay.Session, side int, cfg *CoachConfig) (int, bool) {
	clone, err := s.Clone()
	if err != nil {
		return 0, false
	}
	cc := Coach{Config: cfg, stickyReinforce: maps.Clone(l.c.stickyReinforce), seal: l.c.seal, sealDone: l.c.sealDone, quick: true}
	opp := &dosai.Planner{}
	h := l.Horizon
	if h <= 0 {
		h = strategistHorizon
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
	return EvaluatePlan(clone, side, r), true
}
