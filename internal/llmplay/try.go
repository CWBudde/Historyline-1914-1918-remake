package llmplay

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosai"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// Candidate is one way to play on from a position: Orders (the JSONL of a
// step) for the current phase with the plain coach playing after it, or a
// Plan the coach plays under throughout. Neither: the plain coach.
type Candidate struct {
	Name   string
	Orders string
	Plan   *Directives
}

// TryResult is where a candidate leads.
type TryResult struct {
	Name        string
	Rejected    []string // the candidate's orders or plan were refused; nothing else is set
	Resolutions int
	Value       int // Evaluate at the end
	Units       [2]int
	Energy      int
	Income      [2]int
	Lost        []string
	Killed      []string
	Captured    []string
	End         *dosrules.End
	First       string // the first resolution's events (BriefDigest)
	Ends        string // where the side's units stand at the end
}

// Try plays every candidate res resolutions ahead from s on a clone of
// its own, against dosai's exact replies, and reports where each leads;
// s is left as it is. Like the Lookahead it is an oracle: the rollouts
// know the opponent's answers. base is the coach configuration (nil: the
// default).
func Try(s *dosplay.Session, side int, cands []Candidate, res int, base *CoachConfig) ([]TryResult, error) {
	out := make([]TryResult, len(cands))
	errs := make([]error, len(cands))
	var wg sync.WaitGroup
	for i, c := range cands {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out[i], errs[i] = tryOne(s, side, c, res, base)
		}()
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func tryOne(s *dosplay.Session, side int, c Candidate, res int, base *CoachConfig) (TryResult, error) {
	r := TryResult{Name: c.Name}
	clone, err := s.Clone()
	if err != nil {
		return r, err
	}
	coach := &Coach{Config: base}
	if c.Plan != nil {
		if errs := c.Plan.Check(clone, side); len(errs) > 0 {
			r.Rejected = errs
			return r, nil
		}
		if coach, err = c.Plan.Coach(base); err != nil {
			return r, err
		}
	}
	opp := &dosai.Planner{}
	var t tally
	for i := 0; i < res && clone.Ended() == nil; i++ {
		var evs []dosplay.Event
		if i == 0 && c.Orders != "" {
			orders, err := ParseOrders(c.Orders)
			if err != nil {
				r.Rejected = []string{err.Error()}
				return r, nil //nolint:nilerr // A parse error refuses the candidate, not the comparison.
			}
			_, rej, err := Apply(clone, side, orders)
			if err != nil {
				return r, err
			}
			if len(rej) > 0 {
				r.Rejected = rej
				return r, nil
			}
			evs = clone.Confirm(side)
		} else if evs, err = coach.PlayPhase(clone, side); err != nil {
			return r, err
		}
		oppEvs, err := opp.Plan(clone, 1-side)
		if err != nil {
			return r, err
		}
		resEvs, err := clone.Resolve()
		if err != nil {
			return r, err
		}
		evs = append(append(evs, oppEvs...), resEvs...)
		t.add(evs, side)
		if i == 0 {
			r.First = BriefDigest(evs, side)
		}
		r.Resolutions++
	}
	st := clone.Stats()
	r.Value = Evaluate(clone, side, r.Resolutions)
	r.Units = [2]int{st.Sides[side].Units, st.Sides[1-side].Units}
	r.Energy, r.Income = clone.Energy(side), [2]int{st.Income[side], st.Income[1-side]}
	r.Lost, r.Killed, r.Captured, r.End = t.lost, t.killed, t.captured, clone.Ended()
	own := ownUnits(clone, side)
	var ends []string
	for _, i := range slices.Sorted(maps.Keys(own)) {
		u := own[i]
		ends = append(ends, fmt.Sprintf("u%d %s %d/%d", i, show(u.At), u.Size, u.Full))
	}
	r.Ends = strings.Join(ends, ", ")
	return r, nil
}

// FormatTry is the comparison of the candidates from s: a row each, best
// value first marked with *, and with verbose the first resolution's events.
func FormatTry(s *dosplay.Session, side int, results []TryResult, verbose bool) string {
	st := s.Stats()
	var b strings.Builder
	fmt.Fprintf(&b, "NOW value %d units %dv%d energy %d income %d vs %d\n", Evaluate(s, side, 0),
		st.Sides[side].Units, st.Sides[1-side].Units, s.Energy(side), st.Income[side], st.Income[1-side])
	best := -1
	for i, r := range results {
		if r.Rejected == nil && (r.End == nil || r.End.Winner == side) && (best < 0 || r.Value > results[best].Value) {
			best = i
		}
	}
	for i, r := range results {
		mark := " "
		if i == best {
			mark = "*"
		}
		if r.Rejected != nil {
			fmt.Fprintf(&b, "%s %s: REJECTED\n   %s\n", mark, r.Name, strings.Join(r.Rejected, "\n   "))
			continue
		}
		fmt.Fprintf(&b, "%s %s: after %d res value %d units %dv%d energy %d income %d vs %d", mark, r.Name, r.Resolutions, r.Value,
			r.Units[0], r.Units[1], r.Energy, r.Income[0], r.Income[1])
		if r.End != nil {
			fmt.Fprintf(&b, " GAME OVER winner=side%d won=%v", r.End.Winner, r.End.Won)
		}
		b.WriteByte('\n')
		for _, l := range []struct {
			label string
			items []string
		}{{"lost", r.Lost}, {"killed", r.Killed}, {"buildings", r.Captured}} {
			if len(l.items) > 0 {
				fmt.Fprintf(&b, "   %s %d: %s\n", l.label, len(l.items), strings.Join(l.items, ", "))
			}
		}
		if verbose && r.Ends != "" {
			b.WriteString("   your units at the end: " + r.Ends + "\n")
		}
		if verbose && r.First != "" {
			b.WriteString("   first resolution:\n")
			for _, l := range strings.Split(strings.TrimRight(r.First, "\n"), "\n") {
				b.WriteString("   " + l + "\n")
			}
		}
	}
	return b.String()
}
