package llmplay

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// defaultStopRes is how many resolutions Advance plays when the plan
// names none; maxStopRes bounds any plan.
const (
	defaultStopRes = 10
	maxStopRes     = 60
)

// tally is what a resolution's events did to side.
type tally struct {
	lost, killed, captured []string
	built                  int
}

func (t *tally) add(evs []dosplay.Event, side int) {
	name := func(u dosplay.UnitView) string {
		p := "u"
		if u.Side != side {
			p = "e"
		}
		id := u.ID
		if _, rest, ok := strings.Cut(id, "_"); ok {
			id = rest
		}
		return fmt.Sprintf("%s%d %s", p, u.Index, id)
	}
	for _, ev := range evs {
		switch ev.Kind {
		case dosplay.Battle:
			if ev.Battle == nil {
				continue
			}
			for _, f := range []dosplay.Fighter{ev.Battle.Attacker, ev.Battle.Defender} {
				switch {
				case !f.Destroyed:
				case f.Unit.Side == side:
					t.lost = append(t.lost, name(f.Unit))
				default:
					t.killed = append(t.killed, name(f.Unit))
				}
			}
		case dosplay.Captured:
			t.captured = append(t.captured, fmt.Sprintf("%s took %s", Sides[ev.Side], show(ev.To)))
		case dosplay.Built:
			if ev.Side == side {
				t.built++
			}
		}
	}
}

// line is the tally in a few words; "" for a quiet resolution.
func (t tally) line() string {
	var parts []string
	if len(t.lost) > 0 {
		parts = append(parts, "lost "+strings.Join(t.lost, ", "))
	}
	if len(t.killed) > 0 {
		parts = append(parts, "killed "+strings.Join(t.killed, ", "))
	}
	parts = append(parts, t.captured...)
	if t.built > 0 {
		parts = append(parts, fmt.Sprintf("built %d", t.built))
	}
	return strings.Join(parts, " | ")
}

// planPlayer is the coach under a plan; the transcript's orders text
// names the plan.
type planPlayer struct {
	*Coach
	text string
}

func (p planPlayer) OrdersText() string { return p.text }

// AdvanceReport is what Advance played.
type AdvanceReport struct {
	Lines       []string // one per resolution
	Resolutions int
	Stop        string // why it handed back
}

func (r AdvanceReport) String() string {
	return strings.Join(r.Lines, "\n") + fmt.Sprintf("\nSTOP after %d resolutions: %s\n", r.Resolutions, r.Stop)
}

// enemyNearHQ reports whether side sees an enemy within n hexes of an own HQ.
func enemyNearHQ(s *dosplay.Session, side, n int) bool {
	v := s.View(side)
	for at, b := range v.Buildings {
		if !b.HQ || s.SideOf(b.Owner) != side {
			continue
		}
		for c, u := range v.Units {
			if u.Side == 1-side && hexgrid.Distance(c, at) <= n {
				return true
			}
		}
	}
	return false
}

// arrived names the first goto or capture unit of the plan that stands on
// its hex or is gone.
func (d *Directives) arrived(s *dosplay.Session, side int) string {
	own := ownUnits(s, side)
	for _, g := range d.Groups {
		if g.Goal != dirGoto && g.Goal != dirCapture || g.To == nil {
			continue
		}
		to := parseCoord(*g.To)
		for _, i := range g.Units {
			u, ok := own[i]
			switch {
			case !ok:
				return fmt.Sprintf("u%d is gone", i)
			case u.At == to:
				return fmt.Sprintf("u%d arrived at %s", i, show(to))
			}
		}
	}
	return ""
}

// stopReason is why the plan hands back after a resolution, "" to go on.
func (d *Directives) stopReason(s *dosplay.Session, side int, t tally, wasNear bool) string {
	st := d.Stop
	switch {
	case st.Lost > 0 && len(t.lost) >= st.Lost:
		return fmt.Sprintf("%d units lost", len(t.lost))
	case st.Captured && len(t.captured) > 0:
		return "a building changed hands"
	case st.EnemyNear > 0 && !wasNear && enemyNearHQ(s, side, st.EnemyNear):
		return fmt.Sprintf("enemy within %d hexes of your HQ", st.EnemyNear)
	}
	if st.Arrived {
		return d.arrived(s, side)
	}
	return ""
}

// Advance plays the game on with the coach under plan d (base: the
// level's coach configuration, nil for the default) until the plan's stop
// condition holds or the game ends. Every resolution is recorded like any
// other; the report has a line for each.
func (g *Game) Advance(d *Directives, base *CoachConfig) (AdvanceReport, error) {
	var rep AdvanceReport
	if errs := d.Check(g.S, g.Side); len(errs) > 0 {
		return rep, fmt.Errorf("plan refused: %s", strings.Join(errs, "; "))
	}
	coach, err := d.Coach(base)
	if err != nil {
		return rep, err
	}
	raw, err := json.Marshal(d)
	if err != nil {
		return rep, err
	}
	pl := planPlayer{coach, "advance " + string(raw)}
	limit := d.Stop.Res
	if limit <= 0 {
		limit = defaultStopRes
	}
	limit = min(limit, maxStopRes)
	var total tally
	var digest strings.Builder
	seen := knownEnemies(g.S, g.Side)
	for rep.Stop == "" {
		wasNear := d.Stop.EnemyNear > 0 && enemyNearHQ(g.S, g.Side, d.Stop.EnemyNear)
		phase := phaseName(g.S, g.Side)
		evs, err := g.StepPlayer(pl)
		if err != nil {
			return rep, err
		}
		rep.Resolutions++
		var t tally
		t.add(evs, g.Side)
		total.add(evs, g.Side)
		st := g.S.Stats()
		line := fmt.Sprintf("r%d T%d %s: %dv%d E%d", g.resolution, g.S.Turn(), phase[:3], st.Sides[g.Side].Units, st.Sides[1-g.Side].Units, g.S.Energy(g.Side))
		if s := t.line(); s != "" {
			line += " | " + s
		}
		rep.Lines = append(rep.Lines, line)
		digest.WriteString(BriefDigest(evs, g.Side))
		switch end := g.S.Ended(); {
		case end != nil:
			rep.Stop = fmt.Sprintf("GAME OVER winner=side%d won=%v score=%d", end.Winner, end.Won, end.Score)
		case rep.Resolutions >= limit:
			rep.Stop = fmt.Sprintf("%d resolutions played", limit)
		default:
			rep.Stop = d.stopReason(g.S, g.Side, total, wasNear)
		}
	}
	g.digest, g.seen = digest.String(), seen
	return rep, nil
}
