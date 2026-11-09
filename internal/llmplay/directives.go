package llmplay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Directives are a standing plan the Coach plays under: where named units
// go, what the factories build and when to hand back. Units no group names
// and buildings no rule covers play the coach's own doctrine. A player
// gives them once and lets many resolutions run (Game.Advance) or compares
// several (Try), instead of ordering every unit every phase.
type Directives struct {
	// Posture is one of the Lookahead's postures ("cautious", "bold",
	// "siege", "assault", "turtle", "grab", "grab-cautious"); "" or
	// "coach": the plain doctrine.
	Posture string      `json:"posture,omitempty"`
	Groups  []Group     `json:"groups,omitempty"`
	Build   []BuildRule `json:"build,omitempty"`
	Stop    Stop        `json:"stop,omitzero"`
}

// Group gives units one goal: "goto" (march to To, progress first, and
// hold there), "capture" (goto for a building entrance), "hold" (fight
// within Radius hexes of To), "escort" (stay next to unit Of) or "stay"
// (no move at all). Hexes are [line,col] as in the prompt.
type Group struct {
	Units  []int   `json:"units"`
	Goal   string  `json:"goal"`
	To     *[2]int `json:"to,omitempty"`
	Radius int     `json:"radius,omitempty"`
	Of     *int    `json:"of,omitempty"`
	// Quiet units give no attack orders (a raid that must not stop to
	// fight); Avoid are hexes they neither cross nor end on, each with
	// the hexes within AvoidRadius of it.
	Quiet       bool     `json:"quiet,omitempty"`
	Avoid       [][2]int `json:"avoid,omitempty"`
	AvoidRadius int      `json:"avoidRadius,omitempty"`
}

// BuildRule replaces the coach's production in building Hold (0: every
// own building): the first type of Types (type IDs or their endings,
// "eliteInf") the building can build and the energy allows while Reserve
// stays, Count per action phase (0: 1). No types: nothing is built there.
type BuildRule struct {
	Hold    int      `json:"hold,omitempty"`
	Types   []string `json:"types"`
	Reserve int      `json:"reserve,omitempty"`
	Count   int      `json:"count,omitempty"`
}

// Stop is when Advance hands back: after Res resolutions (0: 10), once
// Lost own units were destroyed, when an enemy comes within EnemyNear
// hexes of an own HQ, when a building changes hands (Captured: any; Taken:
// the side took one; LostBuilding: the enemy took one), when a goto or
// capture unit reached its hex or was lost (Arrived). The game's end
// always stops.
type Stop struct {
	Res          int  `json:"res,omitempty"`
	Lost         int  `json:"lost,omitempty"`
	EnemyNear    int  `json:"enemyNear,omitempty"`
	Captured     bool `json:"captured,omitempty"`
	Taken        bool `json:"taken,omitempty"`
	LostBuilding bool `json:"lostBuilding,omitempty"`
	Arrived      bool `json:"arrived,omitempty"`
}

const (
	dirGoto    = "goto"
	dirCapture = "capture"
	dirHold    = "hold"
	dirEscort  = "escort"
	dirStay    = "stay"
)

// ParseDirectives reads a plan; unknown fields are refused.
func ParseDirectives(text string) (*Directives, error) {
	dec := json.NewDecoder(bytes.NewReader([]byte(text)))
	dec.DisallowUnknownFields()
	var d Directives
	if err := dec.Decode(&d); err != nil {
		return nil, fmt.Errorf("plan: %w", err)
	}
	return &d, nil
}

// findPosture is the posture name of the Lookahead's lists.
func findPosture(name string) (posture, bool) {
	for _, p := range slices.Concat(postures, widePostures) {
		if p.name == name {
			return p, true
		}
	}
	return posture{}, false
}

// ownUnits are side's units by index: on the map and inside its buildings
// and carriers.
func ownUnits(s *dosplay.Session, side int) map[int]dosplay.UnitView {
	v := s.View(side)
	out := map[int]dosplay.UnitView{}
	for at, u := range v.Units {
		if u.Side != side || u.Head != nil {
			continue
		}
		out[u.Index] = u
		if u.Carrier {
			if _, slots, ok := s.Contents(side, at); ok {
				for _, c := range slots {
					out[c.Index] = c
				}
			}
		}
	}
	for at, b := range v.Buildings {
		if s.SideOf(b.Owner) != side {
			continue
		}
		if _, slots, ok := s.Contents(side, at); ok {
			for _, c := range slots {
				out[c.Index] = c
			}
		}
	}
	return out
}

func onBoard(s *dosplay.Session, c hexgrid.Coord) bool {
	_, ok := s.Board.Cell(c)
	return ok
}

// Check names what is wrong with the plan on the session: unknown
// postures, goals and hexes. A plan with errors is not played.
func (d *Directives) Check(s *dosplay.Session, side int) []string {
	errs, _ := d.check(s, side)
	return errs
}

// Warnings name what the plan asks that the position cannot give now: a
// unit that is gone or not there yet, a building not held. The plan is
// played without them, so it survives a loss and may name a building it
// is about to take.
func (d *Directives) Warnings(s *dosplay.Session, side int) []string {
	_, warns := d.check(s, side)
	return warns
}

func (d *Directives) check(s *dosplay.Session, side int) (errs, warns []string) {
	bad := func(format string, a ...any) { errs = append(errs, fmt.Sprintf(format, a...)) }
	warn := func(format string, a ...any) { warns = append(warns, fmt.Sprintf(format, a...)) }
	if d.Posture != "" {
		if _, ok := findPosture(d.Posture); !ok {
			bad("posture %q unknown", d.Posture)
		}
	}
	own := ownUnits(s, side)
	seen := map[int]bool{}
	for i, g := range d.Groups {
		switch g.Goal {
		case dirGoto, dirCapture, dirHold:
			if g.To == nil {
				bad("group %d (%s): no \"to\" hex", i+1, g.Goal)
			} else if c := parseCoord(*g.To); !onBoard(s, c) {
				bad("group %d: %s is off the map", i+1, show(c))
			} else if _, isB := s.View(side).Buildings[c]; g.Goal == dirCapture && !isB {
				bad("group %d: %s is no building entrance", i+1, show(c))
			}
		case dirEscort:
			if g.Of == nil {
				bad("group %d (escort): no \"of\" unit", i+1)
			} else if _, ok := own[*g.Of]; !ok {
				warn("group %d: escorted u%d is not yours (now)", i+1, *g.Of)
			}
		case dirStay:
		default:
			bad("group %d: goal %q unknown (goto, capture, hold, escort, stay)", i+1, g.Goal)
		}
		if len(g.Units) == 0 {
			bad("group %d: no units", i+1)
		}
		for _, a := range g.Avoid {
			if !onBoard(s, parseCoord(a)) {
				bad("group %d: avoid %s is off the map", i+1, show(parseCoord(a)))
			}
		}
		for _, u := range g.Units {
			if _, ok := own[u]; !ok {
				warn("group %d: u%d is not yours (now)", i+1, u)
			}
			if seen[u] {
				bad("group %d: u%d is in an earlier group", i+1, u)
			}
			seen[u] = true
		}
	}
	for i, r := range d.Build {
		if r.Hold == 0 {
			continue
		}
		found := false
		for at, b := range s.View(side).Buildings {
			if hold, _, ok := s.Contents(side, at); ok && hold == r.Hold && s.SideOf(b.Owner) == side {
				found = true
			}
		}
		if !found {
			warn("build rule %d: hold %d is not your building (now)", i+1, r.Hold)
		}
	}
	return errs, warns
}

// Coach is a coach that plays the plan on top of base (nil: the default
// configuration).
func (d *Directives) Coach(base *CoachConfig) (*Coach, error) {
	cfg := *DefaultCoachConfig()
	if base != nil {
		cfg = *base
	}
	if d.Posture != "" {
		p, ok := findPosture(d.Posture)
		if !ok {
			return nil, fmt.Errorf("plan: posture %q unknown", d.Posture)
		}
		p.apply(&cfg)
	}
	cfg.Directives = d
	return &Coach{Config: &cfg}, nil
}

// directed is a unit's goal under the plan.
type directed struct {
	goal   string
	to     hexgrid.Coord
	radius int
	group  int // index into Directives.Groups: its Quiet and Avoid
}

// march reports whether the unit values progress to its hex above all.
func (d directed) march() bool { return d.goal != dirHold && d.goal != dirStay }

// applyDirectives resolves the plan's groups on the position and takes
// their units out of the coach's own roles.
func (p *coachPlan) applyDirectives() {
	d := p.cfg.Directives
	if d == nil {
		return
	}
	p.directed = map[int]directed{}
	var at map[int]dosplay.UnitView
	for gi, g := range d.Groups {
		r := directed{goal: g.Goal, radius: g.Radius, group: gi}
		switch g.Goal {
		case dirGoto, dirCapture, dirHold:
			if g.To == nil {
				continue
			}
			r.to = parseCoord(*g.To)
		case dirEscort:
			if at == nil {
				at = ownUnits(p.s, p.side)
			}
			if g.Of == nil {
				continue
			}
			lead, ok := at[*g.Of]
			if !ok {
				continue // the escorted unit is gone: back to the doctrine
			}
			r.to, r.radius = lead.At, max(g.Radius, 1)
		case dirStay:
		default:
			continue
		}
		for _, u := range g.Units {
			p.directed[u] = r
			delete(p.homeGuard, u)
			delete(p.reinforce, u)
			delete(p.dash, u)
			delete(p.guardOf, u)
			delete(p.plug, u)
		}
	}
}

// quiet reports whether the plan keeps unit from attacking.
func (p *coachPlan) quiet(unit int) bool {
	d, ok := p.directed[unit]
	return ok && p.cfg.Directives.Groups[d.group].Quiet
}

// avoided reports whether the plan bars c for a unit under d.
func (p *coachPlan) avoided(d directed, c hexgrid.Coord) bool {
	g := p.cfg.Directives.Groups[d.group]
	for _, a := range g.Avoid {
		if hexgrid.Distance(parseCoord(a), c) <= g.AvoidRadius {
			return true
		}
	}
	return false
}

// directedSeeds seeds the field of a directed unit: the hexes within its
// radius of the goal hex, the goal's neighbours where the unit cannot
// enter the hex itself (a foreign entrance for a unit that cannot capture).
func directedSeeds(r directed, costs map[hexgrid.Coord]int, seed func(hexgrid.Coord, int)) {
	n := 0
	for c := range costs {
		if hexgrid.Distance(c, r.to) <= r.radius {
			seed(c, 0)
			n++
		}
	}
	if n == 0 {
		for _, c := range hexgrid.Around(r.to) {
			seed(c, 1)
		}
	}
}

// buildByRules is the production the plan's build rules order; done
// reports that rules exist, so the coach's own production is off.
func (p *coachPlan) buildByRules() (evs []dosplay.Event, done bool) {
	d := p.cfg.Directives
	if d == nil || len(d.Build) == 0 {
		return nil, false
	}
	for _, at := range p.ownB {
		hold, _, ok := p.s.Contents(p.side, at)
		if !ok {
			continue
		}
		var rule *BuildRule
		for i := range d.Build {
			if r := &d.Build[i]; r.Hold == hold || r.Hold == 0 && rule == nil {
				rule = r
				if r.Hold == hold {
					break
				}
			}
		}
		if rule == nil {
			continue
		}
		opts := p.s.Buildable(p.side, hold)
		for range max(rule.Count, 1) {
			built := false
			for _, t := range rule.Types {
				for _, o := range opts {
					if o.ID != t && !strings.HasSuffix(o.ID, "_"+t) {
						continue
					}
					if p.s.Energy(p.side)-o.Cost < rule.Reserve || p.s.CanBuild(p.side, hold, o.Type) != nil {
						continue
					}
					ev, err := p.s.Build(p.side, hold, o.Type)
					if err != nil {
						return evs, true
					}
					evs, built = append(evs, ev...), true
					break
				}
				if built {
					break
				}
			}
			if !built {
				break
			}
		}
	}
	return evs, true
}
