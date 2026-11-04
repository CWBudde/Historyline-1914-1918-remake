package llmplay

// Temporary CANDL probes (gate: CANDL_PROBE=1). Not part of the suite.

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosai"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

func candlSession(t *testing.T) *dosplay.Session {
	t.Helper()
	lvl, err := ShipLevel("1918-04", scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewSession(lvl, 0)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func pc(c hexgrid.Coord) string { return fmt.Sprintf("[%d,%d]", c.Line+1, c.Col+1) }

func oppPlanner() *dosai.Planner { return &dosai.Planner{} }

func TestHexDist(t *testing.T) {
	if os.Getenv("CANDL_PROBE") == "" {
		t.Skip("probe")
	}
	hart := hexgrid.Coord{Line: 85, Col: 7} // EC_hArtillery [86,8] 1-based
	ecHQ := hexgrid.Coord{Line: 86, Col: 12}
	fmt.Printf("Around(hArt): %s\n", ringStr(hart))
	fmt.Printf("Around(ecHQ): %s\n", ringStr(ecHQ))
	for _, at := range []hexgrid.Coord{{Line: 88, Col: 2}, {Line: 86, Col: 3}, {Line: 87, Col: 3}, {Line: 85, Col: 3}, {Line: 88, Col: 3}, {Line: 89, Col: 2}, {Line: 87, Col: 2}, {Line: 86, Col: 2}} {
		fmt.Printf(" %s dHArt=%d\n", pc(at), hexgrid.Distance(at, hart))
	}
}

func TestCandlProbe(t *testing.T) {
	if os.Getenv("CANDL_PROBE") == "" {
		t.Skip("probe")
	}
	s := candlSession(t)
	v := s.TrueView()
	// Buildings
	for _, at := range sortedCoords(v.Buildings) {
		b := v.Buildings[at]
		hold, slots, ok := s.Contents(0, at)
		fmt.Printf("B %s owner=%v HQ=%v hold=%d ok=%v slots=%d ring=%v builds0=%v\n", pc(at), b.Owner, b.HQ, hold, ok, len(slots), ringStr(at), buildsStr(s, 0, hold))
		hold1, slots1, ok1 := s.Contents(1, at)
		_ = hold1
		if ok1 && len(slots1) > 0 {
			fmt.Printf("  EC contents visible? %d\n", len(slots1))
		}
	}
	// Type table for present types
	seen := map[int]bool{}
	var types []int
	for _, u := range v.Units {
		if !seen[u.Type] {
			seen[u.Type] = true
			types = append(types, u.Type)
		}
	}
	slices.Sort(types)
	for _, typ := range types {
		tv, ok := s.TypeInfo(0, typ)
		if !ok {
			continue
		}
		tv1, _ := s.TypeInfo(1, typ)
		fmt.Printf("T %-22s / %-22s dom=%v bud? spd=%d wt=%d grp=%d arm=%d cost=%d air=%d/r%d gnd=%d/r%d wat=%d\n",
			tv.ID, tv1.ID, tv.Domain, tv.Speed, tv.Weight, tv.Group, tv.Armour, tv.Cost,
			tv.Air.Power, tv.Air.Range, tv.Ground.Power, tv.Ground.Range, tv.Water.Power)
	}
	// Per-unit opening facts
	for _, at := range sortedCoords(v.Units) {
		u := v.Units[at]
		if u.Side != 0 {
			continue
		}
		costs, err := s.Costs(0, u.Index)
		nc := 0
		if err == nil {
			nc = len(costs)
		}
		dests, derr := s.Destinations(0, u.Index)
		nd := 0
		if derr == nil {
			nd = len(dests)
		}
		fmt.Printf("U u%-3d %-18s %s size%d bud=%d arm=%d carrier=%v held=%v costs=%d dests=%d\n",
			u.Index, u.ID, pc(u.At), u.Size, u.Budget, u.Armour, u.Carrier, u.Held, nc, nd)
	}
	// Route distance (Dijkstra field cost) from each CP unit to HQ entrance and doorsteps
	hq := hexgrid.Coord{Line: 7, Col: 10}
	fmt.Printf("HQ ring: %v\n", ringStr(hq))
	for _, at := range sortedCoords(v.Units) {
		u := v.Units[at]
		if u.Side != 0 {
			continue
		}
		costs, err := s.Costs(0, u.Index)
		if err != nil {
			fmt.Printf("R u%-3d %-18s %s costs err %v\n", u.Index, u.ID, pc(u.At), err)
			continue
		}
		f := dijkstra(costs, map[hexgrid.Coord]int{hq: 0})
		// phases estimate: best progress per move = follow destinations greedily? just field value / budget
		fv, ok := f[u.At]
		ph := -1
		if ok && u.Budget > 0 {
			ph = (fv + u.Budget - 1) / u.Budget
		}
		fmt.Printf("R u%-3d %-18s %s fieldHQ=%d ok=%v estMovePhases>=%d hexdist=%d\n", u.Index, u.ID, pc(u.At), fv, ok, ph, hexgrid.Distance(u.At, hq))
	}
}

func ringStr(c hexgrid.Coord) string {
	var out []string
	ring := hexgrid.Around(c)
	for _, n := range ring {
		out = append(out, pc(n))
	}
	return strings.Join(out, " ")
}

func buildsStr(s *dosplay.Session, side, hold int) string {
	opts := s.Buildable(side, hold)
	var out []string
	for _, o := range opts {
		out = append(out, o.ID)
	}
	return strings.Join(out, ",")
}

func zoneCounts(s *dosplay.Session) string {
	v := s.TrueView()
	hq := hexgrid.Coord{Line: 7, Col: 10}
	cp, ec := 0, 0
	cpHQ, ecHQ, cpWest, cpMid, cpSouth, cpEast := 0, 0, 0, 0, 0, 0
	for _, u := range v.Units {
		if u.Head != nil || u.Held {
			continue
		}
		switch u.Side {
		case 0:
			cp++
			if hexgrid.Distance(u.At, hq) <= 5 {
				cpHQ++
			}
			switch {
			case u.At.Line >= 60:
				cpSouth++
			case u.At.Col <= 4:
				cpWest++
			case u.At.Col >= 12:
				cpEast++
			default:
				cpMid++
			}
		case 1:
			ec++
			if hexgrid.Distance(u.At, hq) <= 6 {
				ecHQ++
			}
		}
	}
	own := 0
	for _, at := range sortedCoords(v.Buildings) {
		if s.SideOf(v.Buildings[at].Owner) == 0 {
			own++
		}
	}
	return fmt.Sprintf("CP %d (HQzone %d, west %d, mid %d, east %d, south %d) EC %d (HQzone %d) energy %d buildings %d",
		cp, cpHQ, cpWest, cpMid, cpEast, cpSouth, ec, ecHQ, s.Energy(0), own)
}

// Probe2: HQ entry costs, idle-CP capture timing, carrier capacity,
// HQ exits, builder targets, balloon engage maths.
func TestCandlProbe2(t *testing.T) {
	if os.Getenv("CANDL_PROBE") == "" {
		t.Skip("probe")
	}
	s := candlSession(t)
	hq := hexgrid.Coord{Line: 7, Col: 10}
	// entry costs per hex for infantry / cavalry types
	for _, name := range []string{"u27inf", "u22cav", "u25a7v"} {
		_ = name
	}
	inf := s.Unit(27)
	costs, err := s.Costs(0, inf.Index)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("infantry costs: HQ entrance=%d(ok=%v) ring:", costs[hq], func() bool { _, ok := costs[hq]; return ok }())
	ring := hexgrid.Around(hq)
	for _, n := range ring {
		fmt.Printf(" %s=%d", pc(n), costs[n])
	}
	fmt.Printf("\nouter ring terrain/costs:\n")
	seen := map[hexgrid.Coord]bool{}
	for _, n := range ring {
		r2 := hexgrid.Around(n)
		for _, m := range r2 {
			if seen[m] || m == hq {
				continue
			}
			seen[m] = true
			terr := ""
			if m.Line >= 0 && m.Line < len(s.Level.Lines) && m.Col >= 0 && m.Col < len(s.Level.Lines[m.Line]) {
				terr = s.Level.Lines[m.Line][m.Col].Terrain
			}
			fmt.Printf(" %s cost=%d terr=%s distHQ=%d\n", pc(m), costs[m], terr, hexgrid.Distance(m, hq))
		}
	}
	// terrain of ring itself
	for _, n := range ring {
		fmt.Printf("ring %s terr=%s\n", pc(n), s.Level.Lines[n.Line][n.Col].Terrain)
	}
	// Builder targets at start
	for _, idx := range []int{80, 87} {
		u := s.Unit(idx)
		fmt.Printf("builder u%d %s at %s targets=%v\n", idx, u.ID, pc(u.At), coordsStr(s.Targets(0, idx)))
	}
	// Carrier loading experiments on a clone (movement phase 1)
	cl, err := s.Clone()
	if err != nil {
		t.Fatal(err)
	}
	try := func(unit int, to hexgrid.Coord, what string) {
		_, err := cl.Move(0, unit, to)
		fmt.Printf("load %s u%d -> %s err=%v\n", what, unit, pc(to), err)
	}
	// train u13 at [8,2] 1-based = (7,1); infantry u18 at (8,1)
	train := cl.Unit(13)
	try(18, train.At, "infantry>train")
	try(11, train.At, "infantry2>train")
	try(19, train.At, "lArt>train?")
	holdT, slotsT, okT := cl.Contents(0, train.At)
	fmt.Printf("train hold=%d ok=%v slots=%d\n", holdT, okT, len(slotsT))
	for _, sl := range slotsT {
		fmt.Printf("  slot %d u%d %s\n", sl.Slot, sl.Index, sl.ID)
	}
	// a7v onto train u15 (7,2)
	train2 := cl.Unit(15)
	try(23, train2.At, "a7v>train?")
	if _, sl, ok := cl.Contents(0, train2.At); ok {
		fmt.Printf("train2 slots=%d\n", len(sl))
	}
	// balloon onto supplyCar u9 (5,10)? balloon u3 too far; skip dest check: print dests
	// supplyCar capacity: infantry u27/u28 next to u9? u27 (11,8), not in range of (5,10). print u9 contents after staticAA boards?
	sc := cl.Unit(9)
	try(10, sc.At, "staticAA>supplyCar")
	if _, sl, ok := cl.Contents(0, sc.At); ok {
		fmt.Printf("supplyCar slots=%d\n", len(sl))
	}
	// Move loaded train, then exits next phase happens later; print train destinations now (after loads it has not moved)
	dests, err := cl.Destinations(0, 13)
	fmt.Printf("loaded train dests err=%v: %v\n", err, coordsStr(dests))
	// HQ entry + exit destinations: fresh clone
	cl2, err := s.Clone()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cl2.Move(0, 9, hq); err != nil {
		fmt.Printf("supplyCar->HQ err=%v\n", err)
	}
	cl2.Confirm(0)
	cl2.Confirm(1)
	if _, err := cl2.Resolve(); err != nil {
		t.Fatal(err)
	}
	// now action phase; confirm through to next movement
	cl2.Confirm(0)
	cl2.Confirm(1)
	if _, err := cl2.Resolve(); err != nil {
		t.Fatal(err)
	}
	holdH, slotsH, okH := cl2.Contents(0, hq)
	fmt.Printf("HQ hold=%d ok=%v slots=%d phase0=%d\n", holdH, okH, len(slotsH), cl2.Phase(0))
	for _, sl := range slotsH {
		ed, err := cl2.ExitDestinations(0, holdH, sl.Slot)
		fmt.Printf(" HQ exit u%d %s slot%d err=%v dests=%v\n", sl.Index, sl.ID, sl.Slot, err, coordsStr(ed))
	}
	// Idle CP: how/when does EC capture, what are EC dests near HQ?
	s2 := candlSession(t)
	pl := &struct{}{}
	_ = pl
	for res := 1; res <= 36; res++ {
		if s2.Phase(1) == 0 { // movement? PhaseMovement value unknown; just query every other res on clone
		}
		cl3, err := s2.Clone()
		if err == nil {
			v := cl3.TrueView()
			for _, at := range sortedCoords(v.Units) {
				u := v.Units[at]
				if u.Side != 1 || hexgrid.Distance(u.At, hq) > 2 {
					continue
				}
				d, err := cl3.Destinations(1, u.Index)
				_, canHQ := d[hq]
				fmt.Printf("idle res%d EC u%d %s at %s distHQ=%d dests=%d canEnterHQ=%v err=%v phaseEC=%d\n",
					res, u.Index, u.ID, pc(u.At), hexgrid.Distance(u.At, hq), len(d), canHQ, err, cl3.Phase(1))
			}
		}
		s2.Confirm(0)
		opp := oppPlanner()
		if _, err := opp.Plan(s2, 1); err != nil {
			t.Fatal(err)
		}
		if _, err := s2.Resolve(); err != nil {
			t.Fatal(err)
		}
		if s2.Ended() != nil {
			fmt.Printf("idle CP: ended res%d %v\n", res, s2.Ended())
			break
		}
		if res == 36 {
			fmt.Printf("idle CP: not ended after 36: %s\n", zoneCounts(s2))
		}
	}
}

func coordsStr(m map[hexgrid.Coord]bool) string {
	var out []string
	for _, c := range dosplay.Coords(m) {
		out = append(out, pc(c))
	}
	return strings.Join(out, " ")
}

func TestCandlCoachDiag(t *testing.T) {
	if os.Getenv("CANDL_PROBE") == "" {
		t.Skip("probe")
	}
	s := candlSession(t)
	g, err := NewGame(s, 0, t.TempDir(), "1918-04:cp", "CANDL", false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = g.Close() }()
	coach := &Coach{}
	fmt.Printf("res 0: %s\n", zoneCounts(s))
	for r := 1; r <= 40; r++ {
		evs, err := g.StepPlayer(coach)
		if err != nil {
			t.Fatalf("res %d: %v", r, err)
		}
		killsCP, killsEC := 0, 0
		for _, ev := range evs {
			if ev.Battle != nil {
				if ev.Battle.Defender.Destroyed {
					if ev.Battle.Defender.Unit.Side == 0 {
						killsCP++
					} else {
						killsEC++
					}
				}
				if ev.Battle.Attacker.Destroyed {
					if ev.Battle.Attacker.Unit.Side == 0 {
						killsCP++
					} else {
						killsEC++
					}
				}
			}
		}
		fmt.Printf("res %d: %s lostCP %d lostEC %d end=%v\n", r, zoneCounts(s), killsCP, killsEC, s.Ended())
		if s.Ended() != nil {
			break
		}
	}
}

// ---------- CANDL general v1 ----------

func refusedEvs(evs []dosplay.Event) bool {
	for _, ev := range evs {
		if ev.Kind == dosplay.Refused {
			return true
		}
	}
	return false
}

// candlGeneral: Coach action phase + explicit CANDL movement doctrine.
// Hold the HQ flower with plugged doorsteps, shuttle the west/south
// armies by rail, retreat the east cut-off group to the HQ, and let the
// coach's movement (with overridden assignments) fight the rest.
type candlGeneral struct {
	coach   *Coach
	sticky  map[int]bool
	freeze  map[int]bool // units the coach must not move this phase
	diag    bool
	balloon map[int]hexgrid.Coord // sticky balloon -> doorstep assignment
	dashers map[int]bool          // units committed to the EC-HQ dash
	goDash  bool                  // latched: the muster has departed
}

func (g *candlGeneral) logf(format string, args ...any) {
	if g.diag {
		fmt.Printf(format+"\n", args...)
	}
}

func doorstepsOf(s *dosplay.Session, hq hexgrid.Coord) []hexgrid.Coord {
	var out []hexgrid.Coord
	ring := hexgrid.Around(hq)
	for _, n := range ring {
		if n.Line < 0 || n.Line >= len(s.Level.Lines) || n.Col < 0 || n.Col >= len(s.Level.Lines[n.Line]) {
			continue
		}
		t := s.Level.Lines[n.Line][n.Col].Terrain
		if strings.HasPrefix(t, "hq") || strings.HasPrefix(t, "water") || strings.HasPrefix(t, "lakeside") ||
			strings.HasPrefix(t, "factory") || strings.HasPrefix(t, "cmp") {
			continue
		}
		out = append(out, n)
	}
	return out
}

func (g *candlGeneral) PlayPhase(s *dosplay.Session, side int) ([]dosplay.Event, error) {
	if s.Ended() != nil || s.Ready(side) {
		return nil, nil
	}
	if s.Phase(side) != dosrules.PhaseMovement {
		if os.Getenv("CANDL_MYACT") != "" {
			return g.act(s, side)
		}
		return g.coach.PlayPhase(s, side)
	}
	// Oracle wrapper, as Coach.moveOracle: plan on a clone, replan
	// doomed units at survival threat, replay the commands.
	clone, err := s.Clone()
	if err != nil {
		return nil, err
	}
	mark := clone.Mark()
	if _, err := g.planMovement(clone, side, nil); err != nil {
		return nil, err
	}
	cmds := clone.CommandsSince(mark)
	if doomed := g.coach.doomedInTwo(clone, side); len(doomed) > 0 {
		clone2, err := s.Clone()
		if err != nil {
			return nil, err
		}
		mark2 := clone2.Mark()
		if _, err := g.planMovement(clone2, side, doomed); err == nil {
			cmds = clone2.CommandsSince(mark2)
		}
	}
	evs, err := s.Replay(cmds)
	if err != nil {
		return nil, err
	}
	return append(evs, s.Confirm(side)...), nil
}

// act is the action phase: repairs, CANDL build doctrine, depot
// building, sapper work, then the coach's exact attacks.
func (g *candlGeneral) act(s *dosplay.Session, side int) ([]dosplay.Event, error) {
	p := newCoachPlan(s, side, g.sticky)
	g.sticky = p.reinforce
	var evs []dosplay.Event
	// Repairs first (flat 15, as the coach).
	type job struct{ hold, slot, missing int }
	var jobs []job
	for _, at := range p.ownB {
		hold, slots, ok := s.Contents(side, at)
		if !ok {
			continue
		}
		for _, u := range slots {
			if u.Size < u.Full {
				jobs = append(jobs, job{hold, u.Slot, u.Full - u.Size})
			}
		}
	}
	slices.SortFunc(jobs, func(a, b job) int { return b.missing - a.missing })
	for _, j := range jobs {
		if s.CanRepair(side, j.hold, j.slot) != nil {
			continue
		}
		ev, err := s.Repair(side, j.hold, j.slot)
		if err != nil {
			break
		}
		evs = append(evs, ev...)
	}
	// Builds: a far factory builds what arrives in time (fast),
	// a near one builds the doorstep war (guns, armour, pool).
	for _, at := range p.ownB {
		hold, slots, ok := s.Contents(side, at)
		if !ok {
			continue
		}
		opts := s.Buildable(side, hold)
		if len(opts) == 0 {
			continue
		}
		var pref []string
		if at.Line >= 40 {
			// The south builds the army that later takes the EC
			// HQ: ranged guns first (free damage at range), then
			// armour, then capturers.
			nAT, nLA, nArm, nFast := 0, 0, 0, 0
			count := func(u dosplay.UnitView) {
				switch {
				case strings.HasSuffix(u.ID, "_antiTank"):
					nAT++
				case strings.HasSuffix(u.ID, "_lArtillery") || strings.HasSuffix(u.ID, "_hArtillery") || strings.HasSuffix(u.ID, "_mArtillery"):
					nLA++
				case strings.HasSuffix(u.ID, "_a7v") || strings.HasSuffix(u.ID, "_markIV") || strings.HasSuffix(u.ID, "_markI"):
					nArm++
				case strings.HasSuffix(u.ID, "_cavalry") || strings.HasSuffix(u.ID, "_eliteInf") || strings.HasSuffix(u.ID, "_armCar"):
					nFast++
				}
			}
			for _, u := range p.own {
				if u.At.Line >= 40 {
					count(u)
				}
			}
			for _, u := range slots {
				count(u)
			}
			nAA := 0
			for _, u := range p.own {
				if u.At.Line >= 40 && (strings.HasSuffix(u.ID, "_mobileAA") || strings.HasSuffix(u.ID, "_staticAA")) {
					nAA++
				}
			}
			// Dash composition: AA (mobileAA is not buildable in
			// this scenario, staticAA is), tanks (save up for
			// them), fast capturers, then guns.
			switch {
			case nAA < 1:
				pref = []string{"_staticAA"}
			case nArm < 4:
				pref = []string{"_a7v", "_markIV"}
			case nFast < 8:
				pref = []string{"_cavalry", "_eliteInf", "_armCar"}
			case nAT < 4:
				pref = []string{"_antiTank"}
			case nLA < 2:
				pref = []string{"_lArtillery"}
			default:
				pref = []string{"_eliteInf", "_cavalry", "_antiTank", "_infantry"}
			}
		} else {
			pref = []string{"_antiTank", "_eliteInf", "_a7v", "_markIV", "_cavalry", "_infantry", "_lArtillery"}
		}
		for len(slots) < 5 {
			built := false
			for _, suf := range pref {
				for _, o := range opts {
					if strings.HasSuffix(o.ID, suf) && s.CanBuild(side, hold, o.Type) == nil {
						ev, err := s.Build(side, hold, o.Type)
						if err != nil {
							break
						}
						evs = append(evs, ev...)
						_, slots, _ = s.Contents(side, at)
						built = true
						g.logf(" build %s at %s", o.ID, pc(at))
						break
					}
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
	// Construction units build a depot at their first site.
	for _, u := range p.own {
		if u.Moved || u.Held || !s.Builder(u.Index) {
			continue
		}
		targets := dosplay.Coords(s.Targets(side, u.Index))
		if len(targets) == 0 {
			continue
		}
		ev, err := s.Attack(side, u.Index, targets[0])
		if err != nil {
			return evs, err
		}
		evs = append(evs, ev...)
		g.logf(" depot order u%d -> %s", u.Index, pc(targets[0]))
		p.markMoved(u.Index)
	}
	// Sappers dig near own buildings or under a near enemy.
	for _, u := range p.own {
		if u.Moved || u.Held || !s.Sapper(u.Index) {
			continue
		}
		nearB, nearE := 999, 999
		for _, b := range p.ownB {
			if d := hexgrid.Distance(u.At, b); d < nearB {
				nearB = d
			}
		}
		for _, e := range p.enemies {
			if d := hexgrid.Distance(u.At, e.At); d < nearE {
				nearE = d
			}
		}
		if nearB > 4 && nearE > 5 {
			continue
		}
		var best hexgrid.Coord
		bestD := 999
		found := false
		for _, t := range dosplay.Coords(s.Targets(side, u.Index)) {
			if _, occ := p.view.Units[t]; occ {
				continue
			}
			for _, e := range p.enemies {
				if d := hexgrid.Distance(t, e.At); d < bestD {
					best, bestD, found = t, d, true
				}
			}
		}
		if found {
			ev, err := s.Attack(side, u.Index, best)
			if err != nil {
				return evs, err
			}
			evs = append(evs, ev...)
			p.markMoved(u.Index)
		}
	}
	ev, err := p.attacks()
	evs = append(evs, ev...)
	if err != nil {
		return evs, err
	}
	return append(evs, s.Confirm(side)...), nil
}

func (g *candlGeneral) planMovement(s *dosplay.Session, side int, safe map[int]bool) ([]dosplay.Event, error) {
	g.freeze = map[int]bool{}
	var evs []dosplay.Event
	add := func(ev []dosplay.Event, err error) error {
		evs = append(evs, ev...)
		return err
	}
	p0 := newCoachPlan(s, side, g.sticky)
	if len(p0.ownHQs) == 0 {
		ev, err := p0.move()
		if err != nil {
			return evs, err
		}
		return append(evs, ev...), nil
	}
	hq := p0.ownHQs[0]
	steps := doorstepsOf(s, hq)

	// 1. Unload carriers that reached the north (cargo walks to the HQ).
	{
		v := s.TrueView()
		for _, at := range sortedCoords(v.Units) {
			u := v.Units[at]
			if u.Side != side || !u.Carrier || u.Moved {
				continue
			}
			if u.At.Line > 15 {
				continue
			}
			hold, slots, ok := s.Contents(side, u.At)
			if !ok || len(slots) == 0 {
				continue
			}
			// Only at a real railhead: the cargo's walk to the HQ
			// from here must be short (bridge area), else ride on.
			if f, ferr := p0.field(slots[0], goalHQ); ferr == nil {
				if fv, ok := f[u.At]; ok && fv > 70 {
					continue
				}
			}
			for _, c := range slots {
				dests, err := s.ExitDestinations(side, hold, c.Slot)
				if err != nil {
					continue
				}
				f, ferr := p0.field(c, goalHQ)
				best, bestV := hexgrid.Coord{}, 1<<30
				for _, d := range dosplay.Coords(dests) {
					if ou, occ := v.Units[d]; occ && ou.Carrier {
						continue
					}
					val := 1 << 29
					if ferr == nil {
						if fv, ok := f[d]; ok {
							val = fv
						}
					}
					if val < bestV {
						best, bestV = d, val
					}
				}
				if bestV < 1<<30 {
					if err := add(s.Exit(side, hold, c.Slot, best)); err != nil {
						return evs, err
					}
					g.logf(" unload u%d %s from u%d -> %s", c.Index, c.ID, u.Index, pc(best))
				}
			}
		}
	}

	// 2. Balloons plug doorsteps (tanks/infantry have no air power).
	{
		v := s.TrueView()
		var balloons []dosplay.UnitView
		for _, at := range sortedCoords(v.Units) {
			u := v.Units[at]
			if u.Side == side && strings.HasSuffix(u.ID, "_balloon") && !u.Moved {
				balloons = append(balloons, u)
			}
		}
		slices.SortFunc(balloons, func(a, b dosplay.UnitView) int {
			return hexgrid.Distance(a.At, hq) - hexgrid.Distance(b.At, hq)
		})
		// Sticky assignment: computed once from the opening so
		// targets never flip (v10 balloons oscillated at the bridge).
		if g.balloon == nil {
			g.balloon = map[int]hexgrid.Coord{}
			free := slices.Clone(steps)
			for _, b := range balloons {
				bestJ, bestD := -1, 999
				for j, d := range free {
					if dd := hexgrid.Distance(b.At, d); dd < bestD {
						bestJ, bestD = j, dd
					}
				}
				if bestJ >= 0 {
					g.balloon[b.Index] = free[bestJ]
					free = slices.Delete(free, bestJ, bestJ+1)
				}
			}
		}
		for _, b := range balloons {
			target, found := g.balloon[b.Index]
			if !found {
				// A later balloon (or freed step): take any free step.
				usedSteps := map[hexgrid.Coord]bool{}
				for _, d := range g.balloon {
					usedSteps[d] = true
				}
				for _, d := range steps {
					if !usedSteps[d] {
						target, found = d, true
						g.balloon[b.Index] = d
						break
					}
				}
			}
			route := target
			if !found || b.At == target {
				g.freeze[b.Index] = true
				continue
			}
			dests, err := s.Destinations(side, b.Index)
			if err != nil || len(dests) == 0 {
				g.freeze[b.Index] = true
				continue
			}
			// The ground holder of the target step vacates into
			// the HQ when the balloon can take its place.
			if hexgrid.Distance(b.At, target) <= 1 {
				if occ, ok := s.TrueView().Units[target]; ok && occ.Side == side && occ.Domain != dosplay.Air && !occ.Moved {
					if od, err := s.Destinations(side, occ.Index); err == nil && od[hq] {
						if err := add(s.Move(side, occ.Index, hq)); err != nil {
							return evs, err
						}
						g.logf(" balloon vacate u%d %s %s -> HQ", occ.Index, occ.ID, pc(target))
						dests, err = s.Destinations(side, b.Index)
						if err != nil {
							g.freeze[b.Index] = true
							continue
						}
					}
				}
			}
			// Route over the air-cost field, threat-aware: the
			// straight line crosses the EC west wing (v4 balloons
			// from the west both died on their first steps).
			best := b.At
			bestScore := 1 << 30
			if costs, cerr := s.Costs(side, b.Index); cerr == nil {
				f := dijkstra(costs, map[hexgrid.Coord]int{route: 0})
				score := func(c hexgrid.Coord) int {
					fv, ok := f[c]
					if !ok {
						fv = 1 << 29
					}
					return fv*4 + p0.threat(b, c)
				}
				bestScore = score(b.At)
				for _, d := range dosplay.Coords(dests) {
					if sc := score(d); sc < bestScore {
						best, bestScore = d, sc
					}
				}
			}
			if best != b.At {
				ev, err := s.Move(side, b.Index, best)
				if err := add(ev, err); err != nil {
					return evs, err
				}
				if !refusedEvs(ev) {
					g.logf(" balloon u%d %s -> %s (target %s)", b.Index, pc(b.At), pc(best), pc(target))
				}
			}
			g.freeze[b.Index] = true
		}
	}

	// 3. Doorstep management, fully explicit: swap damaged/soft
	// holders into the HQ and refill from the reserve pool in the
	// same phase. Every HQ exit hex stays plugged, so the coach's
	// exits cannot deploy the pool away (no destinations).
	stock := 0
	{
		holdHQ, hqSlots, _ := s.Contents(side, hq)
		stock = len(hqSlots)
		plugs := slices.Clone(steps)
		isStep := map[hexgrid.Coord]bool{}
		for _, d := range steps {
			isStep[d] = true
		}
		if len(hqSlots) > 0 {
			if dests, err := s.ExitDestinations(side, holdHQ, hqSlots[0].Slot); err == nil {
				for _, d := range dosplay.Coords(dests) {
					if !isStep[d] {
						plugs = append(plugs, d)
					}
				}
			}
		}
		tough := func(u dosplay.UnitView) int {
			key := u.Armour*10 + u.Size*5
			if tv, ok := s.TypeInfo(side, u.Type); !ok || tv.Ground.Power == 0 || tv.Ground.Range < 1 {
				key -= 400
			}
			return key
		}
		bestHeld := func() (dosplay.UnitView, bool) {
			_, slots, ok := s.Contents(side, hq)
			if !ok {
				return dosplay.UnitView{}, false
			}
			best, bestKey := -1, -1<<30
			for _, c := range slots {
				if c.Size != c.Full {
					continue
				}
				if key := tough(c); key > bestKey {
					best, bestKey = c.Index, key
					_ = best
				}
			}
			for _, c := range slots {
				if c.Size == c.Full && tough(c) == bestKey {
					return c, true
				}
			}
			return dosplay.UnitView{}, false
		}
		for _, d := range plugs {
			v := s.TrueView()
			u, occ := v.Units[d]
			if occ && u.Side == side && u.Domain == dosplay.Air {
				continue // balloon plug
			}
			needFill := !occ || u.Side != side
			if occ && u.Side == side && !u.Moved {
				swap := u.Size <= 4 && u.Size < u.Full
				if !swap && isStep[d] {
					if r, ok := bestHeld(); ok && tough(r) > tough(u)+100 {
						swap = true // a much tougher reserve waits inside
					}
				}
				if swap {
					// Only swap into a refill, never into a hole.
					_, hasHeld := bestHeld()
					if dests, err := s.Destinations(side, u.Index); err == nil && dests[hq] && hasHeld {
						ev, err := s.Move(side, u.Index, hq)
						if err := add(ev, err); err != nil {
							return evs, err
						}
						if !refusedEvs(ev) {
							needFill = true
							g.logf(" swap-in u%d %s %d/%d from %s", u.Index, u.ID, u.Size, u.Full, pc(d))
						}
					}
				}
				if !needFill {
					g.freeze[u.Index] = true // holder stays
					continue
				}
			}
			if !needFill {
				continue
			}
			// Refill: toughest healthy reserve exits onto the plug,
			// else the toughest ground unit that can reach it.
			filled := false
			if r, ok := bestHeld(); ok {
				if dests, err := s.ExitDestinations(side, holdHQ, r.Slot); err == nil && dests[d] {
					ev, err := s.Exit(side, holdHQ, r.Slot, d)
					if err := add(ev, err); err != nil {
						return evs, err
					}
					if !refusedEvs(ev) {
						filled = true
						g.logf(" refill u%d %s HQ -> %s", r.Index, r.ID, pc(d))
					}
				}
			}
			if !filled {
				v2 := s.TrueView()
				best, bestKey := -1, -1<<30
				var bestU dosplay.UnitView
				for _, at := range sortedCoords(v2.Units) {
					o := v2.Units[at]
					if o.Side != side || o.Moved || o.Held || o.Carrier || o.Domain == dosplay.Air || o.Budget == 0 {
						continue
					}
					if dests, err := s.Destinations(side, o.Index); err == nil && dests[d] {
						if key := tough(o); key > bestKey {
							best, bestKey, bestU = o.Index, key, o
						}
					}
				}
				if best >= 0 {
					ev, err := s.Move(side, bestU.Index, d)
					if err := add(ev, err); err != nil {
						return evs, err
					}
					if !refusedEvs(ev) {
						filled = true
						g.logf(" fill u%d %s %s -> %s", bestU.Index, bestU.ID, pc(bestU.At), pc(d))
					}
				}
			}
			_ = filled
		}
		_, hqSlots, _ = s.Contents(side, hq)
		stock = len(hqSlots)
	}
	// Stock march: foot units in the HQ zone walk into the HQ
	// (the coach would never end a move on its own entrance).
	{
		v := s.TrueView()
		holderAt := map[hexgrid.Coord]bool{}
		for _, d := range steps {
			holderAt[d] = true
		}
		type candS struct {
			u dosplay.UnitView
			d int
		}
		var cands []candS
		for _, at := range sortedCoords(v.Units) {
			u := v.Units[at]
			if u.Side != side || u.Moved || u.Held || holderAt[u.At] {
				continue
			}
			if u.Domain != dosplay.Infantry && !u.Carrier {
				continue
			}
			if strings.HasSuffix(u.ID, "_antiTank") || strings.HasSuffix(u.ID, "_sappers") {
				continue // the gun shoots, the sappers dig: both stay out
			}
			if d := hexgrid.Distance(u.At, hq); d <= 8 {
				cands = append(cands, candS{u, d})
			}
		}
		slices.SortFunc(cands, func(a, b candS) int { return a.d - b.d })
		want := 5 - stock
		for _, c := range cands {
			if want <= 0 {
				break
			}
			u := c.u
			want--
			dests, err := s.Destinations(side, u.Index)
			if err != nil {
				continue
			}
			if dests[hq] {
				ev, err := s.Move(side, u.Index, hq)
				if err := add(ev, err); err != nil {
					return evs, err
				}
				if !refusedEvs(ev) {
					stock++
					want = 5 - stock
					g.logf(" stock u%d %s from %s", u.Index, u.ID, pc(u.At))
				}
				continue
			}
			f, ferr := p0.field(u, goalHQ)
			if ferr != nil {
				continue
			}
			best, bestV := u.At, 1<<30
			if fv, ok := f[u.At]; ok {
				bestV = fv
			}
			for _, d := range dosplay.Coords(dests) {
				if ou, occ := v.Units[d]; occ && ou.Carrier {
					continue // do not board a carrier by accident
				}
				if fv, ok := f[d]; ok && fv < bestV {
					best, bestV = d, fv
				}
			}
			if best != u.At {
				ev, err := s.Move(side, u.Index, best)
				if err := add(ev, err); err != nil {
					return evs, err
				}
				if !refusedEvs(ev) {
					g.logf(" stock-march u%d %s %s -> %s", u.Index, u.ID, pc(u.At), pc(best))
				}
			}
			g.freeze[u.Index] = true
		}
	}

	// 4. Boarding: shuttle units board a reachable carrier.
	// North-west: foot + towed guns in the west strip (line<20).
	// South: anyone near the factory (line>=75) or at a pickup.
	full := map[int]bool{}
	board := func(u dosplay.UnitView) bool {
		dests, err := s.Destinations(side, u.Index)
		if err != nil {
			return false
		}
		v := s.TrueView()
		var cands []dosplay.UnitView
		for _, d := range dosplay.Coords(dests) {
			if ou, occ := v.Units[d]; occ && ou.Side == side && ou.Carrier && !ou.Moved && !full[ou.Index] {
				if u.At.Line >= 40 && strings.HasSuffix(ou.ID, "_transTrain") {
					continue // the south marches: the middle rail is no faster than walking
				}
				cands = append(cands, ou)
			}
		}
		slices.SortFunc(cands, func(a, b dosplay.UnitView) int { return a.Index - b.Index })
		for _, c := range cands {
			ev, err := s.Move(side, u.Index, c.At)
			if err != nil {
				continue
			}
			evs = append(evs, ev...)
			if refusedEvs(ev) {
				full[c.Index] = true
				continue
			}
			g.logf(" board u%d %s -> carrier u%d %s", u.Index, u.ID, c.Index, pc(c.At))
			return true
		}
		return false
	}
	{
		v := s.TrueView()
		// Two west foot stay behind as decoys/screen: without them
		// EC's west eliteInf hunts the balloons crossing east.
		var nwFoot []int
		for _, at := range sortedCoords(v.Units) {
			u := v.Units[at]
			if u.Side == side && !u.Moved && !u.Held && u.At.Col <= 4 && u.At.Line < 10 && u.Domain == dosplay.Infantry {
				nwFoot = append(nwFoot, u.Index)
			}
		}
		keepWest := map[int]bool{}
		for i, idx := range nwFoot {
			if i >= len(nwFoot)-2 {
				keepWest[idx] = true
			}
		}
		for _, at := range sortedCoords(v.Units) {
			u := v.Units[at]
			if u.Side != side || u.Moved || u.Held || u.Carrier || u.Domain == dosplay.Air {
				continue
			}
			northWest := u.At.Col <= 4 && u.At.Line < 10 && (u.Domain == dosplay.Infantry || strings.HasSuffix(u.ID, "_lArtillery")) && !keepWest[u.Index]
			south := u.At.Line >= 60 && u.Budget <= 12 // only the slow ride the supply car
			if northWest || south {
				board(u)
			}
		}
	}

	// 5. Move carriers: loaded ones run north, the empty south train
	// runs to the factory pickup, the south supply car follows suit.
	{
		v := s.TrueView()
		var factory hexgrid.Coord
		haveFactory := false
		for _, at := range p0.ownB {
			if s.SideOf(v.Buildings[at].Owner) == side && !v.Buildings[at].HQ && hexgrid.Distance(at, hq) > 20 {
				// The south railhead is the depot (no build list):
				// the rail crosses east and never reaches the factory.
				if len(s.Buildable(side, mustHold(s, side, at))) == 0 {
					factory, haveFactory = at, true
				}
			}
		}
		for _, at := range sortedCoords(v.Units) {
			u := v.Units[at]
			if u.Side != side || !u.Carrier || u.Moved {
				continue
			}
			_, slots, _ := s.Contents(side, u.At)
			dests, err := s.Destinations(side, u.Index)
			if err != nil || len(dests) == 0 {
				continue
			}
			if len(slots) > 0 {
				// The factory supply car waits for a built gun.
				if strings.HasSuffix(u.ID, "_supplyCar") && u.At.Line >= 75 {
					waitGun := false
					for _, bat := range p0.ownB {
						if bat.Line < 40 {
							continue
						}
						if _, bslots, ok := s.Contents(side, bat); ok {
							for _, c := range bslots {
								if strings.HasSuffix(c.ID, "_hArtillery") || strings.HasSuffix(c.ID, "_mArtillery") || strings.HasSuffix(c.ID, "_lArtillery") {
									waitGun = true
								}
							}
						}
					}
					for _, at2 := range sortedCoords(v.Units) {
						o := v.Units[at2]
						if o.Side == side && o.At.Line >= 75 && (strings.HasSuffix(o.ID, "_hArtillery") || strings.HasSuffix(o.ID, "_mArtillery")) {
							waitGun = true
						}
					}
					if waitGun {
						g.freeze[u.Index] = true
						continue
					}
				}
				// The south train waits at the depot until it is
				// worth the run: more units are still walking to it.
				if u.At.Line >= 50 && len(slots) < 4 {
					waiting := false
					for _, at2 := range sortedCoords(v.Units) {
						o := v.Units[at2]
						if o.Side == side && !o.Carrier && !o.Held && o.At.Line > u.At.Line && hexgrid.Distance(o.At, u.At) <= 18 {
							waiting = true
							break
						}
					}
					if waiting {
						g.freeze[u.Index] = true
						continue
					}
				}
				f, ferr := p0.field(slots[0], goalHQ)
				if ferr != nil {
					continue
				}
				score := func(c hexgrid.Coord) int {
					fv, ok := f[c]
					if !ok {
						fv = 1 << 29
					}
					// A loaded carrier runs its route north, but not
					// into a kill zone: threat counts at par with a
					// few hexes of route progress.
					th := p0.threat(u, c)
					if safe[u.Index] {
						th = p0.threat(u, c) * 3
					}
					return fv*4 + th + c.Line
				}
				best, bestV := u.At, score(u.At)
				for _, d := range dosplay.Coords(dests) {
					if sc := score(d); sc < bestV {
						best, bestV = d, sc
					}
				}
				if best != u.At {
					ev, err := s.Move(side, u.Index, best)
					if err := add(ev, err); err != nil {
						return evs, err
					}
					if !refusedEvs(ev) {
						g.logf(" carrier u%d %s loaded %s -> %s", u.Index, u.ID, pc(u.At), pc(best))
					}
					continue
				}
				g.freeze[u.Index] = true // at railhead; unload next phase
				continue
			}
			// Empty: south carriers head for the pickup, others park.
			if strings.HasSuffix(u.ID, "_supplyCar") && u.At.Line >= 75 {
				g.freeze[u.Index] = true // load at the factory first
				continue
			}
			if haveFactory && u.At.Line >= 40 && !full[u.Index] {
				costs, err := s.Costs(side, u.Index)
				if err != nil {
					continue
				}
				target, bestD := u.At, 999
				for c := range costs {
					if dd := hexgrid.Distance(c, factory); dd < bestD {
						target, bestD = c, dd
					}
				}
				f := dijkstra(costs, map[hexgrid.Coord]int{target: 0})
				best, bestV := u.At, 1<<30
				if fv, ok := f[u.At]; ok {
					bestV = fv
				}
				for _, d := range dosplay.Coords(dests) {
					if fv, ok := f[d]; ok && fv < bestV {
						best, bestV = d, fv
					}
				}
				if best != u.At {
					ev, err := s.Move(side, u.Index, best)
					if err := add(ev, err); err != nil {
						return evs, err
					}
					if !refusedEvs(ev) {
						g.logf(" carrier u%d empty -> pickup %s -> %s", u.Index, pc(u.At), pc(best))
					}
					continue
				}
			}
			g.freeze[u.Index] = true
		}
	}

	// 5b. Muster: the south army masses in the band just north of the
	// factory (lines 86-89), clear of its exit ring, until the
	// dash. (Depot mustering marched it through the trainArt
	// gauntlet twice.)
	{
		var factory hexgrid.Coord
		haveFactory := false
		for _, at := range p0.ownB {
			if at.Line >= 40 && len(s.Buildable(side, mustHold(s, side, at))) > 0 {
				factory, haveFactory = at, true
			}
		}
		southForce := 0
		{
			vv := s.TrueView()
			for _, at := range sortedCoords(vv.Units) {
				o := vv.Units[at]
				if o.Side == side && !o.Carrier && !o.Held && o.At.Line >= 78 && o.Domain != dosplay.Air {
					southForce++
				}
			}
		}
		dashReady := g.goDash || (len(p0.enemies) <= 45 && southForce >= 24)
		if haveFactory && !dashReady {
			fields := map[int]map[hexgrid.Coord]int{}
			v := s.TrueView()
			for _, at := range sortedCoords(v.Units) {
				u := v.Units[at]
				if u.Side != side || u.Moved || u.Held || u.Carrier || u.Domain == dosplay.Air || u.At.Line < 75 {
					continue
				}
				if u.At.Line >= 85 && u.At.Line <= 89 && hexgrid.Distance(u.At, factory) > 2 {
					g.freeze[u.Index] = true // in the muster band: hold
					continue
				}
				f, ok := fields[u.Type]
				if !ok {
					costs, err := s.Costs(side, u.Index)
					if err != nil {
						continue
					}
					start := map[hexgrid.Coord]int{}
					for c := range costs {
						if c.Line >= 82 && c.Line <= 89 && c.Col <= 6 && hexgrid.Distance(c, factory) > 2 {
							start[c] = 0
						}
					}
					f = dijkstra(costs, start)
					fields[u.Type] = f
				}
				dests, err := s.Destinations(side, u.Index)
				if err != nil {
					continue
				}
				best, bestV := u.At, 1<<30
				if fv, ok := f[u.At]; ok {
					bestV = fv
				}
				for _, d := range dosplay.Coords(dests) {
					if fv, ok := f[d]; ok && fv < bestV {
						best, bestV = d, fv
					}
				}
				if best != u.At {
					ev, err := s.Move(side, u.Index, best)
					if err := add(ev, err); err != nil {
						return evs, err
					}
				}
				g.freeze[u.Index] = true
			}
		}
	}

	// 5c. Assault: with EC broken, offence units close on the
	// enemy under their own field, ignoring threat (the coach's
	// scoring deadlocks against artillery envelopes). The south
	// force masses for a dash on the EC HQ (15 hexes from the
	// factory, defended by one heavy gun): bypassing the bunker
	// line beats trading with it.
	if len(p0.enemies) <= 45 {
		fields0 := map[int]map[hexgrid.Coord]int{}
		dashFields := map[int]map[hexgrid.Coord]int{}
		var ecHQ hexgrid.Coord
		haveEcHQ := false
		for _, b := range p0.foreign {
			if bv := p0.view.Buildings[b]; bv.HQ && s.SideOf(bv.Owner) == 1-side {
				ecHQ, haveEcHQ = b, true
			}
		}
		// The whole muster departs in one phase when it is big
		// enough (latched): a wide simultaneous close saturates
		// the one or two guns covering the EC HQ.
		if g.dashers == nil {
			g.dashers = map[int]bool{}
		}
		southForce := 0
		{
			vv := s.TrueView()
			for _, at := range sortedCoords(vv.Units) {
				o := vv.Units[at]
				if o.Side == side && !o.Carrier && !o.Held && o.Domain != dosplay.Air && (o.At.Line >= 78 || g.dashers[o.Index]) {
					southForce++
				}
			}
		}
		if haveEcHQ && southForce >= 24 {
			g.goDash = true
		}
		dash := haveEcHQ && g.goDash
		g.logf(" assault-check ecHQ=%v at %s southForce=%d dash=%v enemies=%d", haveEcHQ, pc(ecHQ), southForce, dash, len(p0.enemies))
		v := s.TrueView()
		for _, at := range sortedCoords(v.Units) {
			u := v.Units[at]
			if u.Side != side || u.Moved || u.Held || u.Carrier || u.Budget == 0 || u.Domain == dosplay.Air {
				continue
			}
			if hexgrid.Distance(u.At, hq) <= 6 {
				continue // HQ defence stays
			}
			if !dash || (u.At.Line < 78 && !g.dashers[u.Index]) {
				continue // only the mustered south force dashes; the rest is preserved
			}
			g.dashers[u.Index] = true
			isDash := true
			stage := 0
			if u.At.Col >= 5 && u.At.Line >= 70 {
				stage = 1 // across the southern crossing: at the HQ shell
			}
			_ = isDash
			fmap := dashFields
			if stage == 0 {
				fmap = fields0
			}
			_ = isDash
			f, ok := fmap[u.Type]
			if !ok {
				costs, err := s.Costs(side, u.Index)
				if err != nil {
					continue
				}
				start := map[hexgrid.Coord]int{}
				seed := func(c hexgrid.Coord) {
					if _, ok := costs[c]; ok {
						start[c] = 0
					}
				}
				if stage == 1 {
					// End within own reach of the HQ: a wide shell,
					// not six contested ring hexes.
					r := max(p0.reach(u), 1)
					for c := range costs {
						if d := hexgrid.Distance(c, ecHQ); d >= 1 && d <= r {
							seed(c)
						}
					}
					seed(ecHQ)
				} else {
					// East-bank hexes of the southern crossing.
					for c := range costs {
						if c.Line >= 80 && c.Line <= 86 && c.Col >= 5 && c.Col <= 7 {
							seed(c)
						}
					}
				}
				f = dijkstra(costs, start)
				fmap[u.Type] = f
				g.logf(" assault-field u%d %s stage=%d seeds=%d field=%d at=%s fAt=%d", u.Index, u.ID, stage, len(start), len(f), pc(u.At), f[u.At])
			}
			dests, err := s.Destinations(side, u.Index)
			if err != nil || len(dests) == 0 {
				continue
			}
			// Capture dash if legal this move.
			captureTo := hexgrid.Coord{}
			canCap := false
			if p0.canCapture(u) {
				for _, b := range p0.foreign {
					if dests[b] {
						captureTo, canCap = b, true
						break
					}
				}
			}
			if canCap {
				ev, err := s.Move(side, u.Index, captureTo)
				if err := add(ev, err); err != nil {
					return evs, err
				}
				g.logf(" assault-capture u%d %s -> %s", u.Index, u.ID, pc(captureTo))
				continue
			}
			// No standoff: the whole muster departs in one phase
			// and speed is the defence (EC's south guns farm
			// experience on trickles until they one-shot tanks).
			best, bestV := u.At, 1<<30
			if fv, ok := f[u.At]; ok {
				bestV = fv
			}
			for _, d := range dosplay.Coords(dests) {
				if ou, occ := v.Units[d]; occ && ou.Carrier {
					continue
				}
				fv, ok := f[d]
				if !ok {
					continue
				}
				if fv < bestV || (fv == bestV && p0.threat(u, d) < p0.threat(u, best)) {
					best, bestV = d, fv
				}
			}
			if best != u.At {
				ev, err := s.Move(side, u.Index, best)
				if err := add(ev, err); err != nil {
					return evs, err
				}
				if !refusedEvs(ev) {
					g.logf(" assault u%d %s %s -> %s", u.Index, u.ID, pc(u.At), pc(best))
				}
			}
			g.freeze[u.Index] = true
		}
	}

	// 6. Coach movement for everyone else, assignments overridden:
	// nobody races foreign buildings; south/east march to the HQ,
	// the rest screens and fights locally.
	p := newCoachPlan(s, side, g.sticky)
	g.sticky = p.reinforce
	if !p.siege {
		p.homeGuard = map[int]bool{}
		p.reinforce = map[int]bool{}
		// One west heavy tank is the HQ reserve from the first
		// phase: soft holders die in one focused EC action phase.
		westReserve := -1
		{
			bestD := 999
			for _, u := range p.own {
				if u.Held || u.Domain != dosplay.Armour || u.Armour < 50 || u.At.Col > 4 {
					continue
				}
				if d := hexgrid.Distance(u.At, hq); d < bestD {
					bestD, westReserve = d, u.Index
				}
			}
		}
		for _, u := range p.own {
			if u.Held {
				continue
			}
			switch {
			case u.Domain == dosplay.Air:
				p.markMoved(u.Index)
			case strings.HasSuffix(u.ID, "_transTrain"):
				p.markMoved(u.Index)
			case g.freeze[u.Index]:
				p.markMoved(u.Index)
			case u.Index == westReserve && hexgrid.Distance(u.At, hq) > 3:
				p.reinforce[u.Index] = true
			case len(p.enemies) <= 35 && hexgrid.Distance(u.At, hq) > 6 && u.At.Line < 55 && nearestEnemyDist(p, u) > 6 && !p.damaged(u):
				p.markMoved(u.Index) // preserve the west/north army: the balloon lock needs it alive
			case len(p.enemies) <= 35 && hexgrid.Distance(u.At, hq) > 6 && u.At.Line < 55:
				p.homeGuard[u.Index] = true // fight only what comes close
			case u.At.Line >= 55 && len(p.enemies) <= 22 && p.canCapture(u):
				// EC broken: capturers race its buildings/HQ
				// (goalFor falls through to goalCapture).
			case u.At.Line >= 34 || (u.At.Col >= 11 && u.At.Line >= 15 && hexgrid.Distance(u.At, hq) > 8):
				p.reinforce[u.Index] = true
			case u.Domain == dosplay.Infantry && hexgrid.Distance(u.At, hq) > 3:
				p.reinforce[u.Index] = true // foot marches to the HQ pool, armour/guns fight
			case u.At.Col <= 3 && nearestEnemyDist(p, u) > 7:
				p.reinforce[u.Index] = true // west battle won: armour/guns cross to the HQ
			default:
				p.homeGuard[u.Index] = true
			}
		}
		// Doorstep guards: current occupants + toughest reachable +
		// first healthy stocked unit per empty step (for exits).
		p.guardOf = map[int]hexgrid.Coord{}
		v := s.TrueView()
		used := map[int]bool{}
		// AA held in the HQ exits to a back hex, never a doorstep.
		if _, slots, ok := s.Contents(side, hq); ok {
			for _, c := range slots {
				if strings.HasSuffix(c.ID, "_staticAA") || strings.HasSuffix(c.ID, "_mobileAA") {
					back := hexgrid.Coord{}
					bestLine := 999
					ring := hexgrid.Around(hq)
					for _, n := range ring {
						r2 := hexgrid.Around(n)
						for _, m := range r2 {
							if hexgrid.Distance(m, hq) == 2 && m.Line < bestLine {
								back, bestLine = m, m.Line
							}
						}
					}
					p.guardOf[c.Index] = back
					used[c.Index] = true
				}
			}
		}
		for _, d := range steps {
			if u, occ := v.Units[d]; occ && u.Side == side {
				if u.Domain != dosplay.Air {
					p.guardOf[u.Index] = d
					used[u.Index] = true
				}
				continue
			}
			best, bestKey := -1, -1
			for _, u := range p.own {
				if u.Held || u.Carrier || u.Domain == dosplay.Air || u.Domain == dosplay.Naval || used[u.Index] || g.freeze[u.Index] {
					continue
				}
				if p.s.Sapper(u.Index) {
					continue // sappers dig, they do not plug
				}
				if hexgrid.Distance(u.At, d) > 7 {
					continue
				}
				costs, err := p.costsFor(u)
				if err != nil {
					continue
				}
				if _, ok := costs[d]; !ok {
					continue
				}
				key := u.Armour*10 + u.Size*5 - hexgrid.Distance(u.At, d)*20
				if tv, ok := s.TypeInfo(side, u.Type); !ok || tv.Ground.Power == 0 || tv.Ground.Range < 1 {
					key -= 400 // a holder should shoot back, but reachability wins
				}
				if dests, err := s.Destinations(side, u.Index); err == nil && dests[d] {
					key += 10000 // can plug the step this very phase
				}
				if key > bestKey {
					best, bestKey = u.Index, key
				}
			}
			if best >= 0 {
				p.guardOf[best] = d
				used[best] = true
				continue
			}
			if _, slots, ok := s.Contents(side, hq); ok {
				bestC, bestKeyC := -1, -1
				for _, c := range slots {
					if c.Size != c.Full || used[c.Index] || strings.HasSuffix(c.ID, "_staticAA") || strings.HasSuffix(c.ID, "_mobileAA") || strings.HasSuffix(c.ID, "_sappers") {
						continue
					}
					if tv, ok := s.TypeInfo(side, c.Type); !ok || tv.Ground.Power == 0 || tv.Ground.Range < 1 {
						continue
					}
					if key := c.Armour*10 + c.Size*5; key > bestKeyC {
						bestC, bestKeyC = c.Index, key
					}
				}
				if bestC >= 0 {
					p.guardOf[bestC] = d
					used[bestC] = true
				}
			}
		}
		// Outer ring: pack the hexes at distance 2 from the HQ so
		// holders fight with full support and EC gets few pincer
		// slots; outer deaths open no capture (candl4 lesson).
		{
			outerSeen := map[hexgrid.Coord]bool{}
			var outer []hexgrid.Coord
			ringAll := append(slices.Clone(steps), hq)
			for _, c := range ringAll {
				ring := hexgrid.Around(c)
				for _, n := range ring {
					if outerSeen[n] || hexgrid.Distance(n, hq) != 2 {
						continue
					}
					outerSeen[n] = true
					if n.Line < 0 || n.Line >= len(s.Level.Lines) || n.Col < 0 || n.Col >= len(s.Level.Lines[n.Line]) {
						continue
					}
					t := s.Level.Lines[n.Line][n.Col].Terrain
					if strings.HasPrefix(t, "hq") || strings.HasPrefix(t, "water") || strings.HasPrefix(t, "lakeside") {
						continue
					}
					outer = append(outer, n)
				}
			}
			slices.SortFunc(outer, func(a, b hexgrid.Coord) int {
				if a.Line != b.Line {
					return a.Line - b.Line
				}
				return a.Col - b.Col
			})
			for _, d := range outer {
				if u, occ := v.Units[d]; occ && u.Side == side {
					if u.Domain != dosplay.Air && !used[u.Index] {
						p.guardOf[u.Index] = d
						used[u.Index] = true
					}
					continue
				}
				best, bestKey := -1, -1
				for _, u := range p.own {
					if u.Held || u.Carrier || u.Domain == dosplay.Air || u.Domain == dosplay.Naval || used[u.Index] || g.freeze[u.Index] {
						continue
					}
					if p.s.Sapper(u.Index) || u.Budget == 0 {
						continue
					}
					if hexgrid.Distance(u.At, d) > 6 {
						continue
					}
					costs, err := p.costsFor(u)
					if err != nil {
						continue
					}
					if _, ok := costs[d]; !ok {
						continue
					}
					key := u.Armour*10 + u.Size*5 - hexgrid.Distance(u.At, d)*20
					if tv, ok := s.TypeInfo(side, u.Type); ok && tv.Ground.Power > 0 {
						key += tv.Ground.Range * 80 // guns on the ring shoot the steps clear
					} else {
						key -= 400
					}
					if dests, err := s.Destinations(side, u.Index); err == nil && dests[d] {
						key += 10000
					}
					if key > bestKey {
						best, bestKey = u.Index, key
					}
				}
				if best >= 0 {
					p.guardOf[best] = d
					used[best] = true
				}
			}
		}
	}
	for idx, d := range p.guardOf {
		g.logf(" guard u%d -> %s", idx, pc(d))
	}
	p.safe = safe
	ev, err := p.move()
	evs = append(evs, ev...)
	if err != nil {
		return evs, err
	}
	return evs, nil
}

func nearestEnemyDist(p *coachPlan, u dosplay.UnitView) int {
	best := 999
	for _, e := range p.enemies {
		if d := hexgrid.Distance(u.At, e.At); d < best {
			best = d
		}
	}
	return best
}

func mustHold(s *dosplay.Session, side int, at hexgrid.Coord) int {
	hold, _, _ := s.Contents(side, at)
	return hold
}

func TestCandlGeneralDiag(t *testing.T) {
	if os.Getenv("CANDL_PROBE") == "" {
		t.Skip("probe")
	}
	s := candlSession(t)
	outDir := os.Getenv("CANDL_OUT")
	if outDir == "" {
		outDir = t.TempDir()
	}
	g, err := NewGame(s, 0, outDir, "1918-04:cp", "CANDL", false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = g.Close() }()
	gen := &candlGeneral{coach: &Coach{}, diag: os.Getenv("CANDL_DIAG") != ""}
	hq := hexgrid.Coord{Line: 7, Col: 10}
	steps := doorstepsOf(s, hq)
	fmt.Printf("res 0: %s steps=%v\n", zoneCounts(s), steps)
	maxRes := 120
	if v := os.Getenv("CANDL_MAXRES"); v != "" {
		fmt.Sscanf(v, "%d", &maxRes)
	}
	for r := 1; r <= maxRes; r++ {
		evs, err := g.StepPlayer(gen)
		if err != nil {
			t.Fatalf("res %d: %v", r, err)
		}
		lostCP, lostEC := 0, 0
		for _, ev := range evs {
			if ev.Battle != nil && os.Getenv("CANDL_BATTLES") != "" && r >= 30 {
				b := ev.Battle
				fmt.Printf("  B res%d u%d %s@%s -> u%d %s@%s def %d->%d%s att %d->%d%s\n", r,
					b.Attacker.Unit.Index, b.Attacker.Unit.ID, pc(b.Attacker.Unit.At),
					b.Defender.Unit.Index, b.Defender.Unit.ID, pc(b.Defender.Unit.At),
					b.Defender.Before, b.Defender.After, map[bool]string{true: " X"}[b.Defender.Destroyed],
					b.Attacker.Before, b.Attacker.After, map[bool]string{true: " X"}[b.Attacker.Destroyed])
			}
			if ev.Battle != nil {
				for _, f := range []dosplay.Fighter{ev.Battle.Attacker, ev.Battle.Defender} {
					if f.Destroyed {
						if f.Unit.Side == 0 {
							lostCP++
							if gen.diag {
								fmt.Printf("  DIED u%d %s at %s\n", f.Unit.Index, f.Unit.ID, pc(f.Unit.At))
							}
						} else {
							lostEC++
						}
					}
				}
			}
		}
		v := s.TrueView()
		var occ []string
		for _, d := range steps {
			if u, ok := v.Units[d]; ok {
				occ = append(occ, fmt.Sprintf("%s:u%d %s %d/%d", pc(d), u.Index, strings.TrimPrefix(strings.TrimPrefix(u.ID, "CP_"), "EC_"), u.Size, u.Full))
			} else {
				occ = append(occ, pc(d)+":--")
			}
		}
		_, hqSlots, _ := s.Contents(0, hq)
		var stockStr []string
		for _, sl := range hqSlots {
			stockStr = append(stockStr, fmt.Sprintf("u%d:%s%d/%d", sl.Index, strings.TrimPrefix(sl.ID, "CP_"), sl.Size, sl.Full))
		}
		fmt.Printf("res %d: %s lost %d/%d hqStock %d[%s] steps [%s] end=%v\n", r, zoneCounts(s), lostCP, lostEC, len(hqSlots), strings.Join(stockStr, " "), strings.Join(occ, " "), s.Ended())
		if os.Getenv("CANDL_ROSTER") != "" && r%50 == 0 {
			for _, at := range sortedCoords(v.Units) {
				u := v.Units[at]
				fmt.Printf("  R%d u%d %s %s %d/%d side%d\n", r, u.Index, u.ID, pc(u.At), u.Size, u.Full, u.Side)
			}
			// Why does no EC capturer enter the HQ?
			if cl, err := s.Clone(); err == nil {
				// Advance the clone to EC's next movement phase.
				for i := 0; i < 3 && cl.Phase(1) != dosrules.PhaseMovement; i++ {
					cl.Confirm(0)
					cl.Confirm(1)
					_, _ = cl.Resolve()
				}
				cv := cl.TrueView()
				for _, at := range sortedCoords(cv.Units) {
					u := cv.Units[at]
					if u.Side != 1 || hexgrid.Distance(u.At, hq) > 4 {
						continue
					}
					d, err := cl.Destinations(1, u.Index)
					_, canHQ := d[hq]
					fmt.Printf("  CAPCHK u%d %s at %s distHQ=%d dests=%d canEnterHQ=%v err=%v phaseEC=%d destList=%s\n",
						u.Index, u.ID, pc(u.At), hexgrid.Distance(u.At, hq), len(d), canHQ, err, cl.Phase(1), coordsStr(d))
				}
				cmds, err := oppPlanner().Commands(cl, 1)
				fmt.Printf("  CAPCHK dosai cmds err=%v n=%d\n", err, len(cmds))
				for _, c := range cmds {
					fmt.Printf("    cmd %s unit=%d cell=%d\n", c.Kind, c.Unit, c.Cell)
				}
			}
		}
		if os.Getenv("CANDL_SOUTH") != "" && r%4 == 1 {
			for _, at := range sortedCoords(v.Units) {
				u := v.Units[at]
				if u.Side == 0 && u.At.Line >= 40 {
					cargo := ""
					if u.Carrier {
						_, sl, _ := s.Contents(0, u.At)
						cargo = fmt.Sprintf(" cargo=%d", len(sl))
					}
					fmt.Printf("  S u%d %s %s%s\n", u.Index, u.ID, pc(u.At), cargo)
				}
			}
			for _, at := range sortedCoords(v.Buildings) {
				b := v.Buildings[at]
				fmt.Printf("  SB %s owner=%v HQ=%v\n", pc(at), b.Owner, b.HQ)
			}
		}
		if s.Ended() != nil {
			break
		}
	}
}
