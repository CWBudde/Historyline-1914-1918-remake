package llmplay

// Coach is the strategy of docs/llmplay-strategy.md as executable Go: a
// PhasePlayer that plays a whole side through the session API (no text
// prompt, so self-play is fast). It differs from dosai.Planner where the
// losses said to:
//
//   - economy every action phase: repair first (flat cost, full
//     strength), then build by role (capturers, anti-tank, sappers,
//     armour), in every own factory;
//   - exact attacks like dosai, but valued by target cost/role, and
//     long-range units never trade lives: they only fire when the exact
//     outcome costs them nothing (duelling from range) or kills;
//   - movement by role over terrain route-cost fields: capturers make
//     deliberate detours to buildings (a capture dash is safe the
//     resolution it is ordered — battles fight on old positions),
//     damaged units retreat into buildings to repair, defenders keep
//     the HQ/factory screen instead of rushing, and every destination
//     is penalised by the enemy long-range envelopes around it;
//   - transports are used: building garrisons exit toward their goal,
//     isolated garrisons embark on ships, loaded ships sail to the
//     coast nearest the cargo's goal and unload there;
//   - aircraft camp on bridge hexes (a unit on a hex blocks it) when a
//     front bridge is out of AA reach, else hunt artillery/aircraft.
//
// Coach is deterministic: every list is sorted, no randomness.

import (
	"container/heap"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosai"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// PhasePlayer plays one side's whole phase including Confirm.
type PhasePlayer interface {
	PlayPhase(s *dosplay.Session, side int) ([]dosplay.Event, error)
}

const (
	longRangeFlag = 0x40 // Type.TargetFlags
	captureFlag   = 0x01 // MoveType.Flags: enters foreign buildings
	unreachable   = 1 << 29
)

// Coach plays the strategy. The zero Coach is ready to use; use a
// pointer so reinforcement assignments stay sticky across phases
// (recomputed-from-scratch assignments oscillated: units marched
// west one phase and back east the next, v21).
type Coach struct {
	stickyReinforce map[int]bool
}

type coachPlan struct {
	s         *dosplay.Session
	side      int
	opp       int
	view      dosplay.View
	prof      dosrules.Profile
	own       []dosplay.UnitView
	enemies   []dosplay.UnitView
	ownB      []hexgrid.Coord // own building entrances
	foreign   []hexgrid.Coord // enemy+neutral entrances
	ownHQs    []hexgrid.Coord
	bridges   map[hexgrid.Coord]bool
	fields    map[fieldKey]map[hexgrid.Coord]int
	costs     map[int]map[hexgrid.Coord]int // by unit Type
	loaded    map[int]bool                  // carriers loaded this phase
	guardOf   map[int]hexgrid.Coord         // unit index -> HQ entrance it must hold
	congested map[hexgrid.Coord]bool        // hexes around full buildings: vacate!
	homeGuard map[int]bool                  // nearest ground units to an own HQ: they screen, never rush
	reinforce map[int]bool                  // eastern units marching to the HQ screen
	siege     bool                          // enemy broken (<=12 left): all-out on its HQ
	siegeHard bool                          // only immobile/artillery enemies left: tanks assault too
	exitHexes map[hexgrid.Coord]bool        // building exit hexes: do not park here
	safe      map[int]bool                  // oracle-doomed units: weight threat heavily
}

type fieldKey struct {
	typ  int
	goal goalKind
}

type goalKind int

const (
	goalFight goalKind = iota
	goalCapture
	goalHome
	goalHQ   // own HQs only (sappers: trench the HQ doorstep)
	goalRace // east/NE fighters: escort the dash at the enemy HQ
)

func newCoachPlan(s *dosplay.Session, side int, sticky map[int]bool) *coachPlan {
	p := &coachPlan{
		s: s, side: side, opp: 1 - side, view: s.View(side), prof: s.Game.Profile,
		fields: map[fieldKey]map[hexgrid.Coord]int{}, costs: map[int]map[hexgrid.Coord]int{}, loaded: map[int]bool{},
	}
	for _, at := range sortedCoords(p.view.Units) {
		u := p.view.Units[at]
		if u.Head != nil {
			continue // rear half of a two-cell ship: not an orderable unit
		}
		switch u.Side {
		case side:
			p.own = append(p.own, u)
		case p.opp:
			p.enemies = append(p.enemies, u)
		}
	}
	for _, at := range sortedCoords(p.view.Buildings) {
		bv := p.view.Buildings[at]
		if s.SideOf(bv.Owner) == side {
			p.ownB = append(p.ownB, at)
			if bv.HQ {
				p.ownHQs = append(p.ownHQs, at)
			}
		} else {
			p.foreign = append(p.foreign, at)
		}
	}
	// Congested doorsteps: a building holding 2+ units cannot unload
	// while its exit hexes are parked on (coach v4/v5 banked 400
	// energy behind a full factory; exit reach is 2 hexes, so the
	// vacate radius is 2, applied to where own units actually stand).
	p.congested = map[hexgrid.Coord]bool{}
	for _, at := range p.ownB {
		if _, slots, ok := s.Contents(side, at); ok && len(slots) >= 2 {
			for _, u := range p.own {
				if hexgrid.Distance(u.At, at) <= 2 {
					p.congested[u.At] = true
				}
			}
		}
	}
	// Home guard: the 8 ground units nearest an own HQ form its
	// screen. In v1–v6 the west infantry rushed foreign buildings and
	// the HQ front thinned to two units by the time EC arrived.
	p.homeGuard = map[int]bool{}
	if len(p.ownHQs) > 0 {
		type du struct {
			idx, d int
		}
		var near []du
		for _, u := range p.own {
			if u.Held || u.Carrier || u.Domain == dosplay.Air || u.Domain == dosplay.Naval {
				continue
			}
			d := 999
			for _, hq := range p.ownHQs {
				if dd := hexgrid.Distance(u.At, hq); dd < d {
					d = dd
				}
			}
			near = append(near, du{u.Index, d})
		}
		slices.SortFunc(near, func(a, b du) int { return a.d - b.d })
		for i, n := range near {
			if i >= 8 || n.d > 10 {
				break // eastern units are not the HQ screen (v20 drafted them and clamped the offence)
			}
			p.homeGuard[n.idx] = true
		}
	}
	// Reinforcements: when the HQ screen thins below 7 ground units,
	// the nearest eastern fighters march west to rebuild it (v19 held
	// 46 resolutions and led 55:27 in units, but the screen died of
	// old age with the whole army in the east).
	p.reinforce = map[int]bool{}
	for idx := range sticky {
		p.reinforce[idx] = true
	}
	// Arrival ends the assignment: a reinforcement inside the screen
	// radius becomes a normal screen unit.
	for _, u := range p.own {
		if !p.reinforce[u.Index] {
			continue
		}
		for _, hq := range p.ownHQs {
			if hexgrid.Distance(u.At, hq) <= 8 {
				delete(p.reinforce, u.Index)
			}
		}
	}
	if len(p.ownHQs) > 0 {
		west := 0
		for _, u := range p.own {
			if u.Held || u.Carrier || u.Domain == dosplay.Air || u.Domain == dosplay.Naval {
				continue
			}
			for _, hq := range p.ownHQs {
				if hexgrid.Distance(u.At, hq) <= 8 {
					west++
					break
				}
			}
		}
		// Threshold 12 (starting west strength is 9): a westward
		// column marches from the first phase. Assignments made when
		// the screen was already thin (v20–v23) arrived 20
		// resolutions after the HQ fell — the march itself takes
		// ~25 resolutions across the map.
		if need := 12 - west; need > 0 {
			type du struct{ idx, d int }
			var east []du
			for _, u := range p.own {
				if u.Held || u.Carrier || u.Domain == dosplay.Air || u.Domain == dosplay.Naval || p.homeGuard[u.Index] {
					continue
				}
				if p.s.Sapper(u.Index) {
					continue
				}
				d := 999
				for _, hq := range p.ownHQs {
					if dd := hexgrid.Distance(u.At, hq); dd < d {
						d = dd
					}
				}
				if d > 8 {
					east = append(east, du{u.Index, d})
				}
			}
			slices.SortFunc(east, func(a, b du) int { return a.d - b.d })
			for i, e := range east {
				if i >= need {
					break
				}
				p.reinforce[e.idx] = true
			}
		}
	}
	// Siege: with the enemy broken, the whole army converges on its
	// HQ and its bunker ring. Without it v24 stalled 550 resolutions
	// at 46:7 — the armour mass sat in a support cluster in the east
	// while 6 bunkers guarded the EC HQ untouched.
	p.siege = len(p.enemies) <= 12
	p.siegeHard = p.siege
	for _, e := range p.enemies {
		if e.Budget > 0 && !p.longRange(e) && e.Domain != dosplay.Artillery {
			p.siegeHard = false
		}
	}
	p.bridges = bridgeHexes(s)
	// HQ doorstep guards: a unit cannot park ON its own entrance
	// (moving there enters the building), so up to three tough ground
	// units hold the entrance's enterable neighbour hexes — every
	// capturer must end a move on one of them first, inside the
	// screen's exact fire (v2–v5 kept losing the HQ this way; the
	// scout's doorstep analysis names these hexes).
	p.guardOf = map[int]hexgrid.Coord{}
	for _, hq := range p.ownHQs {
		var steps []hexgrid.Coord
		ring := hexgrid.Around(hq)
		for _, n := range ring {
			if n.Line < 0 || n.Line >= len(s.Level.Lines) || n.Col < 0 || n.Col >= len(s.Level.Lines[n.Line]) {
				continue
			}
			t := s.Level.Lines[n.Line][n.Col].Terrain
			if strings.HasPrefix(t, "hq") || strings.HasPrefix(t, "water") || strings.HasPrefix(t, "lakeside") || strings.HasPrefix(t, "factory") || strings.HasPrefix(t, "cmp") {
				continue
			}
			steps = append(steps, n)
		}
		type candG struct {
			u     dosplay.UnitView
			tough int
		}
		var cands []candG
		for _, u := range p.own {
			if u.Held || u.Carrier || u.Domain == dosplay.Air || u.Domain == dosplay.Naval || p.canCapture(u) {
				continue
			}
			if _, taken := p.guardOf[u.Index]; taken {
				continue
			}
			if d := hexgrid.Distance(u.At, hq); d > 9 {
				continue
			}
			cands = append(cands, candG{u, u.Armour * u.Size})
		}
		slices.SortFunc(cands, func(a, b candG) int { return b.tough - a.tough })
		for i, step := range steps {
			if i >= len(cands) {
				break
			}
			// Nearest tough unit to this doorstep hex.
			bestJ, bestD := -1, 999
			for j, c := range cands {
				if _, taken := p.guardOf[c.u.Index]; taken {
					continue
				}
				if d := hexgrid.Distance(c.u.At, step); d < bestD {
					bestJ, bestD = j, d
				}
			}
			if bestJ >= 0 {
				p.guardOf[cands[bestJ].u.Index] = step
			}
		}
		// Bridge guards: tough ground units sit ON the bridge hexes
		// near the HQ that aircraft are not plugging (a unit on the
		// hex blocks the crossing completely).
		var bridgesNear []hexgrid.Coord
		for b := range p.bridges {
			if hexgrid.Distance(b, hq) <= 9 {
				if u, occ := p.view.Units[b]; occ && u.Side == side {
					continue // already plugged (plane or guard)
				}
				bridgesNear = append(bridgesNear, b)
			}
		}
		slices.SortFunc(bridgesNear, func(a, b hexgrid.Coord) int {
			if d := hexgrid.Distance(a, hq) - hexgrid.Distance(b, hq); d != 0 {
				return d
			}
			if a.Line != b.Line {
				return a.Line - b.Line
			}
			return a.Col - b.Col
		})
		for _, b := range bridgesNear {
			best, bestD := -1, 999
			for _, u := range p.own {
				if u.Held || u.Carrier || u.Domain == dosplay.Air || u.Domain == dosplay.Naval {
					continue
				}
				if _, taken := p.guardOf[u.Index]; taken {
					continue
				}
				if u.Armour*u.Size < 120 { // tough enough to sit under fire
					continue
				}
				if d := hexgrid.Distance(u.At, b); d < bestD {
					best, bestD = u.Index, d
				}
			}
			if best >= 0 && bestD <= 8 {
				p.guardOf[best] = b
			}
		}
	}
	return p
}

// bridgeHexes are the passage hexes an aircraft can block: road/rail
// infra over water terrain, or an infra named bridge.
func bridgeHexes(s *dosplay.Session) map[hexgrid.Coord]bool {
	out := map[hexgrid.Coord]bool{}
	for l, line := range s.Level.Lines {
		for c, h := range line {
			water := strings.HasPrefix(h.Terrain, "water") || strings.HasPrefix(h.Terrain, "lakeside")
			for _, id := range h.Infra.IDs {
				if strings.Contains(strings.ToLower(id), "bridge") || water && (strings.HasPrefix(id, "road") || strings.HasPrefix(id, "rail")) {
					out[hexgrid.Coord{Line: l, Col: c}] = true
				}
			}
		}
	}
	return out
}

func (p *coachPlan) canCapture(u dosplay.UnitView) bool {
	m := p.prof.Tables.Move
	return m != nil && u.Type < len(m.Types) && m.Types[u.Type].Flags&captureFlag != 0
}

func (p *coachPlan) longRange(u dosplay.UnitView) bool {
	return p.prof.Tables.Types[u.Type].TargetFlags&longRangeFlag != 0
}

func (p *coachPlan) costsFor(u dosplay.UnitView) (map[hexgrid.Coord]int, error) {
	if c, ok := p.costs[u.Type]; ok {
		return c, nil
	}
	c, err := p.s.Costs(p.side, u.Index)
	if err != nil {
		return nil, err
	}
	p.costs[u.Type] = c
	return c, nil
}

// field is the least terrain route cost from every hex to the goal set
// of kind g, for units of u's movement type (Dijkstra over Costs, as in
// dosai). Fight goals ring the enemy units at the unit's own reach.
func (p *coachPlan) field(u dosplay.UnitView, g goalKind) (map[hexgrid.Coord]int, error) {
	k := fieldKey{u.Type, g}
	if f, ok := p.fields[k]; ok {
		return f, nil
	}
	costs, err := p.costsFor(u)
	if err != nil {
		return nil, err
	}
	start := map[hexgrid.Coord]int{}
	seed := func(c hexgrid.Coord, v int) {
		if _, ok := costs[c]; !ok {
			return
		}
		if old, ok := start[c]; !ok || v < old {
			start[c] = v
		}
	}
	// Buildings are seeded by their ring too: a non-capturer's Costs
	// omit foreign entrances entirely (Block mask), so seeding only
	// the entrance left tanks with an empty field and they milled in
	// place 550 resolutions from the EC HQ (v24/v25 stall).
	seedBuilding := func(c hexgrid.Coord, v int) {
		seed(c, v)
		ring := hexgrid.Around(c)
		for _, n := range ring {
			seed(n, v)
		}
	}
	switch g {
	case goalCapture:
		// The race: neutral prize first (income), enemy HQ preferred,
		// other enemy buildings only when they are on the way (+80).
		for _, c := range p.foreign {
			bv := p.view.Buildings[c]
			v := 80
			switch {
			case p.s.SideOf(bv.Owner) == 2:
				v = 0
			case bv.HQ:
				v = -40
			}
			seed(c, v)
		}
	case goalRace:
		for _, c := range p.foreign {
			bv := p.view.Buildings[c]
			if bv.HQ && p.s.SideOf(bv.Owner) == p.opp {
				seedBuilding(c, 0)
			}
			if p.s.SideOf(bv.Owner) == 2 {
				seedBuilding(c, 0)
			}
		}
		ringDist := max(p.reach(u), 1)
		if strings.HasSuffix(u.ID, "_antiTank") {
			ringDist = 1 // anti-tank guns one-shot bunkers adjacent (probe: 3->0 for 1 own loss)
		}
		for _, e := range p.enemies {
			if hexgrid.Distance(u.At, e.At) <= 8 {
				for c := range costs {
					if hexgrid.Distance(c, e.At) == ringDist {
						seed(c, 0)
					}
				}
			}
		}
	case goalHome:
		for _, c := range p.ownB {
			seed(c, 0)
		}
	case goalHQ:
		for _, c := range p.ownHQs {
			seed(c, 0)
		}
	case goalFight:
		for _, c := range p.foreign {
			seed(c, 0)
		}
		lo, hi := 1, 1
		if p.longRange(u) {
			hi = max(p.reach(u), 1)
			lo = min(2, hi)
		}
		foes := p.enemies
		if u.Domain == dosplay.Air {
			// Aircraft pick their prey: ground-attack planes hunt the
			// artillery that shells our screen (it cannot answer air),
			// fighters hunt aircraft. Everything else is a fallback.
			tv, _ := p.s.TypeInfo(p.side, u.Type)
			var pri []dosplay.UnitView
			for _, e := range foes {
				if tv.Ground.Power >= 50 && (p.longRange(e) || e.Domain == dosplay.Artillery) {
					pri = append(pri, e)
				}
				if tv.Ground.Power < 50 && e.Domain == dosplay.Air {
					pri = append(pri, e)
				}
			}
			if len(pri) > 0 {
				foes = pri
			}
		}
		for _, e := range foes {
			for c := range costs {
				if d := hexgrid.Distance(c, e.At); d >= lo && d <= hi {
					seed(c, 0)
				}
			}
		}
	}
	f := dijkstra(costs, start)
	p.fields[k] = f
	return f, nil
}

// reach is the unit's longest weapon range as displayed (hexes).
func (p *coachPlan) reach(u dosplay.UnitView) int {
	tv, ok := p.s.TypeInfo(p.side, u.Type)
	if !ok {
		return 1
	}
	return max(tv.Air.Range, tv.Ground.Range, tv.Water.Range, 1)
}

type cItem struct {
	c hexgrid.Coord
	d int
}
type cQueue []cItem

func (q cQueue) Len() int           { return len(q) }
func (q cQueue) Less(i, j int) bool { return q[i].d < q[j].d }
func (q cQueue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q *cQueue) Push(x any)        { *q = append(*q, x.(cItem)) }
func (q *cQueue) Pop() any {
	old := *q
	it := old[len(old)-1]
	*q = old[:len(old)-1]
	return it
}

func dijkstra(costs map[hexgrid.Coord]int, start map[hexgrid.Coord]int) map[hexgrid.Coord]int {
	dist := make(map[hexgrid.Coord]int, len(costs))
	q := &cQueue{}
	for c, v := range start {
		dist[c] = v
		heap.Push(q, cItem{c, v})
	}
	for q.Len() > 0 {
		it := heap.Pop(q).(cItem)
		if it.d > dist[it.c] {
			continue
		}
		d := it.d + costs[it.c]
		for _, n := range hexgrid.Around(it.c) {
			if _, ok := costs[n]; !ok {
				continue
			}
			if old, ok := dist[n]; !ok || d < old {
				dist[n] = d
				heap.Push(q, cItem{n, d})
			}
		}
	}
	return dist
}

// PlayPhase implements PhasePlayer.
func (c *Coach) PlayPhase(s *dosplay.Session, side int) ([]dosplay.Event, error) {
	if s.Ended() != nil || s.Ready(side) {
		return nil, nil
	}
	if s.Phase(side) == dosrules.PhaseMovement {
		return c.moveOracle(s, side)
	}
	p := newCoachPlan(s, side, c.stickyReinforce)
	c.stickyReinforce = p.reinforce
	evs, err := p.act()
	if err != nil {
		return evs, err
	}
	return append(evs, s.Confirm(side)...), nil
}

// moveOracle plans the movement phase on a Clone, then simulates the
// deterministic opponent two resolutions ahead: its shots in the
// resolution after next use the positions our moves choose now (its
// own repositioning in between included, exactly, via
// dosai.Planner on the clone). Units predicted destroyed there get
// one replan with threat weighted at survival rate. The chosen plan's
// commands then Replay onto the real session. This is the asymmetry
// the heuristic versions (v1–v12, all lost the HQ by ~24) lacked:
// dosai's exact focus fire can simply be read in advance.
func (c *Coach) moveOracle(s *dosplay.Session, side int) ([]dosplay.Event, error) {
	hardSiege := false
	planOn := func(clone *dosplay.Session, safe map[int]bool) ([]dosrules.Command, error) {
		mark := clone.Mark()
		p := newCoachPlan(clone, side, c.stickyReinforce)
		c.stickyReinforce = p.reinforce
		hardSiege = p.siegeHard
		p.safe = safe
		if _, err := p.move(); err != nil {
			return nil, err
		}
		return clone.CommandsSince(mark), nil
	}
	clone, err := s.Clone()
	if err != nil {
		return nil, err
	}
	cmds, err := planOn(clone, nil)
	if err != nil {
		return nil, err
	}
	// In a hard siege the oracle must not veto the assault: bunkers
	// are killed by units that close under fire and trade (v30 froze
	// 30:6 for 400 resolutions with every approach predicted costly).
	if !hardSiege {
		if doomed := c.doomedInTwo(clone, side); len(doomed) > 0 {
			clone2, err := s.Clone()
			if err != nil {
				return nil, err
			}
			if cmds2, err := planOn(clone2, doomed); err == nil {
				cmds = cmds2
			}
		}
	}
	evs, err := s.Replay(cmds)
	if err != nil {
		return nil, err
	}
	return append(evs, s.Confirm(side)...), nil
}

// doomedInTwo advances the clone (already holding our planned moves)
// through: opponent's current shots, our action phase, the opponent's
// move, our shots, and the opponent's next attack plan — and reports
// our units destroyed by that final attack. Any failure means no
// oracle (empty set), never an aborted real phase.
func (c *Coach) doomedInTwo(clone *dosplay.Session, side int) map[int]bool {
	saved := c.stickyReinforce
	defer func() { c.stickyReinforce = saved }()
	opp := 1 - side
	pl := &dosai.Planner{}
	clone.Confirm(side)
	if _, err := pl.Plan(clone, opp); err != nil {
		return nil
	}
	if _, err := clone.Resolve(); err != nil {
		return nil
	}
	if clone.Ended() != nil {
		return nil
	}
	if _, err := c.PlayPhase(clone, side); err != nil { // our action phase
		return nil
	}
	if _, err := pl.Plan(clone, opp); err != nil { // opponent moves
		return nil
	}
	if _, err := clone.Resolve(); err != nil { // our shots
		return nil
	}
	if clone.Ended() != nil {
		return nil
	}
	if _, err := pl.Plan(clone, opp); err != nil { // opponent's next attacks
		return nil
	}
	clone.Confirm(side)
	evs, err := clone.Resolve()
	if err != nil {
		return nil
	}
	doomed := map[int]bool{}
	for _, ev := range evs {
		if ev.Battle == nil {
			continue
		}
		for _, f := range []dosplay.Fighter{ev.Battle.Attacker, ev.Battle.Defender} {
			if f.Unit.Side == side && f.Destroyed {
				doomed[f.Unit.Index] = true
			}
		}
	}
	return doomed
}

// threat is the danger of ending a move at c for unit u: every enemy
// whose (next-phase) envelope — its move reach plus weapon range — can
// touch c adds its power against u's domain, long-range and many-hex
// envelopes weighted most (those are what killed careless movers).
func (p *coachPlan) threat(u dosplay.UnitView, c hexgrid.Coord) int {
	danger, worst := 0, 0
	for _, e := range p.enemies {
		tv, ok := p.s.TypeInfo(p.opp, e.Type)
		if !ok {
			continue
		}
		w := tv.Ground
		switch u.Domain {
		case dosplay.Air:
			w = tv.Air
		case dosplay.Naval:
			w = tv.Water
		}
		if w.Power == 0 {
			continue
		}
		// Cap the speculative envelope at range+2: the full
		// move+range radius (up to 9 for heavy artillery) froze
		// strategic movement (sappers never left the east, v14).
		// Real deaths further out are the oracle's job.
		moveHexes := min(e.Budget/6, 2)
		radius := moveHexes + max(w.Range, 1)
		if d := hexgrid.Distance(c, e.At); d <= radius {
			pen := w.Power * (radius - d + 1)
			if p.longRange(e) {
				pen *= 2
			}
			danger += pen
			if pen > worst {
				worst = pen
			}
		}
	}
	// The worst envelope dominates; the sum only shades it. (v1 summed
	// everything and the whole army froze out of range.)
	return worst + danger/8
}

func terrainBonus(s *dosplay.Session, rew map[hexgrid.Coord]struct{}, c hexgrid.Coord) int {
	_ = rew
	if c.Line < 0 || c.Line >= len(s.Level.Lines) || c.Col < 0 || c.Col >= len(s.Level.Lines[c.Line]) {
		return 0
	}
	h := s.Level.Lines[c.Line][c.Col]
	for _, id := range h.Infra.IDs {
		if strings.HasPrefix(id, "trench") {
			return 40
		}
	}
	t := strings.TrimSuffix(h.Terrain, "_w")
	switch {
	case t == "forest":
		return 20
	case strings.HasPrefix(t, "mount"):
		return 25
	case t == "house":
		return 20
	case t == "stone":
		return 15
	case t == "hills":
		return 10
	case t == "mud":
		return -10
	}
	return 0
}

func (p *coachPlan) damaged(u dosplay.UnitView) bool {
	// In a hard siege the assault units press: retreating at 3 to
	// repair and returning looped forever 3 hexes from the bunkers
	// (v31 equilibrium). They trade at adjacency instead.
	if p.siegeHard && (strings.HasSuffix(u.ID, "_antiTank") || p.canCapture(u)) {
		return u.Size <= 1
	}
	// Retreat at 3, not 2: the enemy action phase still shoots the
	// old position, so a 2-strength unit ordered home usually dies
	// on the way (the u101 lesson).
	return u.Size <= 3 && u.Size < u.Full || u.Size*2 <= u.Full && p.cost(u.Type) >= 50
}

func (p *coachPlan) cost(typ int) int {
	c := p.prof.Tables.Types[typ].Cost
	if c >= 255 {
		return 70
	}
	return c
}

func (p *coachPlan) goalFor(u dosplay.UnitView) goalKind {
	if _, isGuard := p.guardOf[u.Index]; isGuard {
		return goalHome
	}
	switch {
	case p.damaged(u) && len(p.ownB) > 0:
		return goalHome
	case p.s.Sapper(u.Index):
		return goalHQ // trench the HQ doorstep, wherever it was built
	case p.reinforce[u.Index]:
		return goalHQ // march west and rebuild the HQ screen
	case p.homeGuard[u.Index]:
		return goalFight // screen the HQ; foreign buildings are others' job
	case p.canCapture(u):
		return goalCapture
	case u.At.Col >= 14 && !u.Carrier && u.Domain != dosplay.Air:
		return goalRace // east/NE fighters escort the HQ dash
	default:
		return goalFight
	}
}

// isDefender: a gun/bunker starting near an own building keeps the
// screen (dosai's rush left the CP HQ open; the baseline lost by HQ).
func (p *coachPlan) isDefender(u dosplay.UnitView) bool {
	if p.siege {
		_, isGuard := p.guardOf[u.Index]
		return isGuard // enemy broken: only the doorstep guards stay home
	}
	if p.homeGuard[u.Index] {
		return true
	}
	id := u.ID
	// AA guns cannot shoot ground targets at all: never screen with them.
	gun := strings.Contains(id, "antiTank") || strings.Contains(id, "Artillery") ||
		strings.Contains(id, "bunker")
	if !gun {
		return false
	}
	// Only the HQ screen is clamped. Guns at a factory must advance:
	// clamping them there packed the east ring so solid that the
	// factory had no exit hexes at all (v8).
	for _, b := range p.ownHQs {
		if hexgrid.Distance(u.At, b) <= 6 {
			return true
		}
	}
	return false
}

// scoreDest values ending unit u's move at c (higher is better).
func (p *coachPlan) scoreDest(u dosplay.UnitView, from, c hexgrid.Coord, f map[hexgrid.Coord]int) int {
	fv, ok := f[c]
	if !ok {
		fv = unreachable
	}
	fromV, ok := f[from]
	if !ok {
		fromV = unreachable
	}
	score := (fromV - fv) * 3
	if g, isGuard := p.guardOf[u.Index]; isGuard && c == g {
		score += 20000 // hold the HQ entrance
	}
	if bv, isB := p.view.Buildings[c]; isB {
		owner := p.s.SideOf(bv.Owner)
		if owner != p.side && p.canCapture(u) {
			score += 10000 // capture resolves this resolution, before enemy fire at the new hex
		}
		if owner == p.side {
			if p.damaged(u) {
				score += 5000 // enter to repair
			} else {
				// Ending on our own entrance loads the unit into the
				// building: healthy units churned in/out of the east
				// factory for whole phases this way (probe, v11).
				score -= 5000
			}
		}
	}
	// Strategic marchers (reinforcements, sappers heading for the HQ)
	// value progress above all: v22's reinforcements sat in a forest
	// cluster at col 21 for 25 resolutions because support+terrain
	// bonuses outweighed the route gradient.
	siegeAssault := p.siege && !p.homeGuard[u.Index] &&
		(strings.HasSuffix(u.ID, "_antiTank") || p.canCapture(u) || u.Domain == dosplay.Air || p.siegeHard)
	if p.reinforce[u.Index] || p.s.Sapper(u.Index) && p.goalFor(u) == goalHQ || siegeAssault {
		return score*4 - p.threat(u, c)/20
	}
	score += terrainBonus(p.s, nil, c) * 2
	// Support: friendly neighbours (support terms and pincer defence).
	for _, n := range hexgrid.Around(c) {
		if ou, ok := p.view.Units[n]; ok && ou.Side == p.side {
			score += 12
		}
	}
	// Long-range units belong at their reach from the enemy, not adjacent.
	nd := 999
	for _, e := range p.enemies {
		if d := hexgrid.Distance(c, e.At); d < nd {
			nd = d
		}
	}
	// Closing into firing range during the movement phase shoots
	// first: battles next resolution fight on these positions, the
	// enemy's own move does not count yet. Ending exactly at weapon
	// reach of the nearest enemy is the attacking position (v4 stayed
	// out of range and was shot without reply).
	if nd == p.reach(u) && !u.Carrier {
		score += 70
	}
	if p.longRange(u) && nd <= 1 {
		score -= 120
	}
	if p.isDefender(u) {
		near := 999
		for _, b := range p.ownHQs {
			if d := hexgrid.Distance(c, b); d < near {
				near = d
			}
		}
		if near > 6 {
			score -= 400
		}
	}
	if p.exitHexes[c] && c != from {
		score -= 800 // keep the factory doors free (v7: 14 built units trapped inside)
	}
	// Foot units are the first victims of exact focus fire (v9):
	// they pay threat at a higher rate than armour. Oracle-doomed
	// units pay it at the maximum rate: survival first.
	div := 10
	if u.Domain == dosplay.Infantry {
		div = 4
	}
	if p.safe[u.Index] {
		div = 1
	}
	if p.reinforce[u.Index] {
		div = 20 // marching to the sound of the guns: progress first
	}
	score -= p.threat(u, c) / div
	if c == from {
		score += 5
		if p.congested[from] {
			if _, isGuard := p.guardOf[u.Index]; !isGuard {
				// Must beat the exit-hex penalty (-800) on the other
				// ring hexes, or ring units shuffle and stay (v10).
				score -= 2000
			}
		}
	}
	return score
}

// bestDest picks the highest-scored destination; ok=false means staying
// is at least as good.
func (p *coachPlan) bestDest(u dosplay.UnitView, from hexgrid.Coord, dests map[hexgrid.Coord]bool) (hexgrid.Coord, bool, error) {
	// Static screen: a home-guard unit already in defensible terrain
	// (or on its guard hex) holds position. Shuffling west units lost
	// terrain/support every phase and dosai's exact fire punished
	// the old hex anyway (v17/v18 west fell at 20).
	if p.homeGuard[u.Index] && !p.damaged(u) && !p.siege {
		if g, isGuard := p.guardOf[u.Index]; isGuard && from == g {
			return from, false, nil
		}
		if _, isGuard := p.guardOf[u.Index]; !isGuard && terrainBonus(p.s, nil, from) >= 15 {
			return from, false, nil
		}
	}
	f, err := p.field(u, p.goalFor(u))
	if err != nil {
		// A destroyed cargo record can linger in a hold; its Costs
		// refuse with "no such unit". Staying put must not abort the
		// whole phase (coach v2/v3 died at resolution 17 on this).
		return from, false, nil
	}
	best, bestScore := from, p.scoreDest(u, from, from, f)
	for _, c := range dosplay.Coords(dests) {
		if sc := p.scoreDest(u, from, c, f); sc > bestScore {
			best, bestScore = c, sc
		}
	}
	return best, best != from, nil
}

// move runs the movement phase: unload, exit garrisons, then move in
// role order (capturers, damaged, fighters, aircraft, carriers).
func (p *coachPlan) move() ([]dosplay.Event, error) {
	var evs []dosplay.Event
	add := func(ev []dosplay.Event, err error) error {
		evs = append(evs, ev...)
		return err
	}
	// 1. Unload carriers whose cargo can exit toward its goal on land.
	for _, u := range p.own {
		if !u.Carrier || u.Moved {
			continue
		}
		hold, slots, ok := p.s.Contents(p.side, u.At)
		if !ok {
			continue
		}
		for _, c := range slots {
			dests, err := p.s.ExitDestinations(p.side, hold, c.Slot)
			if err != nil {
				if err := add(nil, err); err != nil {
					return evs, err
				}
			}
			land := map[hexgrid.Coord]bool{}
			for d := range dests {
				if ou, occ := p.view.Units[d]; occ && ou.Carrier {
					continue // another carrier: not landfall
				}
				land[d] = true
			}
			if len(land) == 0 {
				continue
			}
			best, goNow, err := p.bestDest(c, u.At, land)
			if err != nil {
				return evs, err
			}
			if goNow {
				if err := add(p.s.Exit(p.side, hold, c.Slot, best)); err != nil {
					return evs, err
				}
				// The exited unit has its order; mark it in the snapshot.
				p.markMoved(c.Index)
			}
		}
	}
	// 2. Building exits run AFTER the moves below (doorstep units
	// vacate the exit hexes first; exitHexes keeps movers off them).
	// The closure is defined here so the exit hexes are known before
	// any unit moves.
	p.exitHexes = map[hexgrid.Coord]bool{}
	for _, at := range p.ownB {
		hold, slots, ok := p.s.Contents(p.side, at)
		if !ok {
			continue
		}
		for _, c := range slots {
			if dests, err := p.s.ExitDestinations(p.side, hold, c.Slot); err == nil {
				for d := range dests {
					p.exitHexes[d] = true
				}
			}
		}
		// Geometric fallback: when the ring is packed, the engine
		// reports no exit destinations at all — reserve the whole
		// 2-hex doorstep anyway so it can clear.
		if len(slots) >= 2 {
			for _, u := range p.own {
				if hexgrid.Distance(u.At, at) <= 2 {
					p.exitHexes[u.At] = true
				}
			}
			ring := hexgrid.Around(at)
			for _, n := range ring {
				p.exitHexes[n] = true
			}
		}
	}
	exitBuildings := func() error {
		for _, at := range p.ownB {
			hold, slots, ok := p.s.Contents(p.side, at)
			if !ok {
				continue
			}
			for _, c := range slots {
				dests, err := p.s.ExitDestinations(p.side, hold, c.Slot)
				if err != nil {
					continue
				}
				f, err := p.field(c, p.goalFor(c))
				if err != nil {
					continue // destroyed record in a slot; skip it
				}
				_, landReachable := f[at]
				var landBest hexgrid.Coord
				landScore, embarkScore := -1<<30, -1<<30
				var embarkBest hexgrid.Coord
				for _, d := range dosplay.Coords(dests) {
					sc := p.scoreDest(c, at, d, f)
					if ou, occ := p.view.Units[d]; occ && ou.Carrier {
						if sc > embarkScore {
							embarkBest, embarkScore = d, sc
						}
						continue
					}
					if sc > landScore {
						landBest, landScore = d, sc
					}
				}
				switch {
				case landReachable && landScore > -1<<29 && !p.damaged(c):
					// Garrisons always deploy: a unit inside a building
					// shoots nothing and is lost with it on capture (v1
					// let built units sit inside until the HQ fell).
					if err := add(p.s.Exit(p.side, hold, c.Slot, landBest)); err != nil {
						return err
					}
					p.markMoved(c.Index)
				case !landReachable && embarkScore > -1<<29 && !p.carrierMoved(embarkBest):
					if err := add(p.s.Exit(p.side, hold, c.Slot, embarkBest)); err != nil {
						return err
					}
					if ou, occ := p.view.Units[embarkBest]; occ {
						p.loaded[ou.Index] = true
					}
					p.markMoved(c.Index)
				}
			}
		}
		return nil
	}
	// 3. Moves in role order.
	ordered := slices.Clone(p.own)
	rank := func(u dosplay.UnitView) int {
		switch {
		case p.canCapture(u):
			return 0
		case p.damaged(u):
			return 1
		case u.Domain == dosplay.Air:
			return 3
		case u.Carrier:
			return 4
		default:
			return 2
		}
	}
	slices.SortFunc(ordered, func(a, b dosplay.UnitView) int {
		if rank(a) != rank(b) {
			return rank(a) - rank(b)
		}
		return a.Index - b.Index
	})
	for _, u := range ordered {
		if u.Moved || u.Held || u.Budget == 0 {
			continue
		}
		if u.Carrier {
			if err := add(p.moveCarrier(u)); err != nil {
				return evs, err
			}
			continue
		}
		if u.Domain == dosplay.Air && p.isPlugger(u) {
			if done, err := p.movePlane(u, add); err != nil {
				return evs, err
			} else if done {
				continue
			}
		}
		dests, err := p.s.Destinations(p.side, u.Index)
		if err != nil || len(dests) == 0 {
			continue
		}
		best, goNow, err := p.bestDest(u, u.At, dests)
		if err != nil {
			return evs, err
		}
		if goNow {
			if err := add(p.s.Move(p.side, u.Index, best)); err != nil {
				return evs, err
			}
		}
	}
	// 4. Now the doors are free: exit the garrisons.
	if err := exitBuildings(); err != nil {
		return evs, err
	}
	return evs, nil
}

// carrierMoved reports whether the carrier on hex c already has its
// order this phase (loading onto it would be refused).
func (p *coachPlan) carrierMoved(c hexgrid.Coord) bool {
	ou, occ := p.view.Units[c]
	return occ && p.s.Unit(ou.Index).Moved
}

func (p *coachPlan) markMoved(idx int) {
	for i, u := range p.own {
		if u.Index == idx {
			p.own[i].Moved = true
		}
	}
}

// moveCarrier: loaded ships sail to the destination whose neighbouring
// land is closest (route cost) to the cargo's goal; empty carriers go
// back toward the nearest own building still holding units, staying
// out of threat where possible. Combat carriers are rare; they simply
// use bestDest on the fight field.
func (p *coachPlan) moveCarrier(u dosplay.UnitView) ([]dosplay.Event, error) {
	dests, err := p.s.Destinations(p.side, u.Index)
	if err != nil || len(dests) == 0 {
		return nil, err
	}
	_, slots, _ := p.s.Contents(p.side, u.At)
	hasCargo := len(slots) > 0 || p.loaded[u.Index]
	if !hasCargo {
		// Wait by a loaded building: exits (step 4) embark onto us,
		// and loading onto an already-moved carrier is refused.
		for _, at := range p.ownB {
			if hexgrid.Distance(u.At, at) <= 3 {
				if _, waiting, ok := p.s.Contents(p.side, at); ok && len(waiting) > 0 {
					return nil, nil
				}
			}
		}
		best, goNow, err := p.bestDest(u, u.At, dests)
		if err != nil || !goNow {
			return nil, err
		}
		return p.s.Move(p.side, u.Index, best)
	}
	// Field of the first cargo unit (or of the ship itself as proxy
	// when the cargo was loaded this phase and Contents lags).
	cargo := u
	if len(slots) > 0 {
		cargo = slots[0]
	}
	f, err := p.field(cargo, p.goalFor(cargo))
	if err != nil {
		return nil, nil // destroyed cargo record: hold position
	}
	best, bestScore := u.At, -1<<30
	for _, c := range dosplay.Coords(dests) {
		near := unreachable
		ring := hexgrid.Around(c)
		for _, n := range append(ring[:], c) {
			if v, ok := f[n]; ok && v < near {
				near = v
			}
		}
		sc := -near*3 - p.threat(u, c)/4
		if sc > bestScore {
			best, bestScore = c, sc
		}
	}
	if best == u.At {
		return nil, nil
	}
	return p.s.Move(p.side, u.Index, best)
}

// isPlugger: the ground-attack aircraft plugs bridges (it also
// shoots tanks from there for free); the fighter never plugs — its
// job is killing the Camels that farm our soft units, and the EC
// mobileAA that shot both pluggers down in v16. If only one aircraft
// is left, it plugs.
func (p *coachPlan) isPlugger(u dosplay.UnitView) bool {
	tv, ok := p.s.TypeInfo(p.side, u.Type)
	if !ok {
		return true
	}
	if tv.Ground.Power >= 50 {
		return true
	}
	for _, o := range p.own {
		if o.Index != u.Index && o.Domain == dosplay.Air {
			if otv, ok := p.s.TypeInfo(p.side, o.Type); ok && otv.Ground.Power >= 50 {
				return false // the Junkers plugs; this one hunts
			}
		}
	}
	return true
}

// movePlane parks the aircraft on a front bridge hex when that blocks
// an enemy approach without sitting in AA reach; reports whether it
// handled the unit.
func (p *coachPlan) movePlane(u dosplay.UnitView, add func([]dosplay.Event, error) error) (bool, error) {
	if len(p.bridges) == 0 {
		return false, nil
	}
	// Already plugging a bridge: stay plugged. Ground units without
	// air power cannot hurt an aircraft; only AA/camels force it off.
	if p.bridges[u.At] && p.threat(u, u.At) <= 200 {
		return true, nil
	}
	dests, err := p.s.Destinations(p.side, u.Index)
	if err != nil || len(dests) == 0 {
		return false, err
	}
	best, bestScore := u.At, -1<<30
	for _, c := range dosplay.Coords(dests) {
		if !p.bridges[c] {
			continue
		}
		// Front bridge: an enemy is within 7 hexes of it.
		nd := 999
		for _, e := range p.enemies {
			if d := hexgrid.Distance(c, e.At); d < nd {
				nd = d
			}
		}
		if nd > 7 {
			continue
		}
		// Not under AA/artillery air fire we cannot survive: use the
		// threat sum as the guard (air power only, via threat()).
		if p.threat(u, c) > 200 {
			continue
		}
		sc := 1000 - nd*10 - p.threat(u, c)/8
		if sc > bestScore {
			best, bestScore = c, sc
		}
	}
	if bestScore == -1<<30 || best == u.At {
		return false, nil
	}
	return true, add(p.s.Move(p.side, u.Index, best))
}

// act runs the action phase: economy, sapper work, then exact attacks.
func (p *coachPlan) act() ([]dosplay.Event, error) {
	var evs []dosplay.Event
	evs = append(evs, p.service()...)
	// Sappers dig toward the enemy near own buildings (their Targets
	// include unit-free work hexes; an attack order tasks them).
	for _, u := range p.own {
		if u.Moved || u.Held || !p.s.Sapper(u.Index) {
			continue
		}
		nearB := 999
		for _, b := range p.ownB {
			if d := hexgrid.Distance(u.At, b); d < nearB {
				nearB = d
			}
		}
		if nearB > 4 {
			continue
		}
		var best hexgrid.Coord
		bestD := 999
		found := false
		for _, t := range dosplay.Coords(p.s.Targets(p.side, u.Index)) {
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
			ev, err := p.s.Attack(p.side, u.Index, best)
			if err != nil {
				return evs, err
			}
			evs = append(evs, ev...)
			p.markMoved(u.Index)
		}
	}
	ev, err := p.attacks()
	evs = append(evs, ev...)
	return evs, err
}

// service is the action-phase economy: repair every damaged held unit
// (cheapest full-strength in the game), then build by role in every
// own building with a free slot.
func (p *coachPlan) service() []dosplay.Event {
	var evs []dosplay.Event
	type job struct {
		hold, slot, missing int
	}
	var jobs []job
	for _, at := range p.ownB {
		hold, slots, ok := p.s.Contents(p.side, at)
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
		if p.s.CanRepair(p.side, j.hold, j.slot) != nil {
			continue
		}
		ev, err := p.s.Repair(p.side, j.hold, j.slot)
		if err != nil {
			return evs
		}
		evs = append(evs, ev...)
	}
	// Role counts drive production. ONE build per building per action
	// phase: exits drain about one unit per movement phase through a
	// single free hex, and building faster only jams the slots (probe
	// v11: 7-deep backlog while energy banked to 900).
	capturers, sappers, antiTanks, aaGuns, enemyAir, footCapturers := 0, 0, 0, 0, 0, 0
	for _, u := range p.own {
		if p.canCapture(u) {
			capturers++
			if u.Domain == dosplay.Infantry {
				footCapturers++
			}
		}
		if p.s.Sapper(u.Index) {
			sappers++
		}
		if strings.HasSuffix(u.ID, "_antiTank") {
			antiTanks++
		}
		if strings.HasSuffix(u.ID, "_mobileAA") || strings.HasSuffix(u.ID, "_staticAA") {
			aaGuns++
		}
	}
	for _, e := range p.enemies {
		if e.Domain == dosplay.Air {
			enemyAir++
		}
	}
	want := []struct {
		match func(id string) bool
		when  bool
	}{
		{func(id string) bool { return strings.HasSuffix(id, "_mobileAA") }, enemyAir > 0 && aaGuns < 2},
		{func(id string) bool { return strings.HasSuffix(id, "_sappers") }, sappers < 2},
		{func(id string) bool { return strings.HasSuffix(id, "_infantry") || strings.HasSuffix(id, "_eliteInf") }, capturers < 4 || footCapturers < 3 || p.siege && footCapturers < 6},
		{func(id string) bool { return strings.HasSuffix(id, "_antiTank") }, antiTanks < 8},
		{func(id string) bool { return strings.HasSuffix(id, "_a7v") || strings.HasSuffix(id, "_markIV") }, p.s.Energy(p.side) >= 150},
		{func(id string) bool { return strings.HasSuffix(id, "_sappers") }, sappers < 2},
		{func(id string) bool { return strings.HasSuffix(id, "_eliteInf") }, true},
		{func(id string) bool { return strings.HasSuffix(id, "_antiTank") }, true},
	}
	for _, at := range p.ownB {
		hold, slots, ok := p.s.Contents(p.side, at)
		if !ok {
			continue
		}
		opts := p.s.Buildable(p.side, hold)
		if len(opts) == 0 {
			continue
		}
		for len(slots) <= 4 { // short exit pipeline, but spend the bank
			built := false
			for _, w := range want {
				if !w.when {
					continue
				}
				for _, o := range opts {
					if w.match(o.ID) && p.s.CanBuild(p.side, hold, o.Type) == nil {
						ev, err := p.s.Build(p.side, hold, o.Type)
						if err != nil {
							return evs
						}
						evs = append(evs, ev...)
						_, slots, _ = p.s.Contents(p.side, at)
						built = true
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
	return evs
}

type cand struct {
	unit   int
	target hexgrid.Coord
	cell   int
}

// attacks queues exact attacks greedily on the state the earlier ones
// leave (ResolveOrders lookahead, as dosai), valued by gain2 and
// filtered by the long-range/no-suicide rules.
func (p *coachPlan) attacks() ([]dosplay.Event, error) {
	bf, err := p.s.Battlefield(p.side)
	if err != nil {
		return nil, err
	}
	var cands []cand
	for _, u := range p.own {
		if u.Moved || u.Held || p.s.Sapper(u.Index) {
			continue
		}
		for _, t := range dosplay.Coords(p.s.Targets(p.side, u.Index)) {
			tv, ok := p.view.Units[t]
			if !ok || tv.Side != p.opp {
				continue
			}
			cell, ok := p.s.Board.Cell(t)
			if !ok {
				continue
			}
			cands = append(cands, cand{u.Index, t, cell})
		}
	}
	var orders []dosrules.Order
	var evs []dosplay.Event
	after := func() (*dosrules.Battlefield, error) {
		if len(orders) == 0 {
			return bf, nil
		}
		r, err := p.prof.ResolveOrders(bf, orders)
		if err != nil {
			return nil, err
		}
		b := *bf
		b.Units = make(map[int]dosrules.Unit, len(r.Units))
		for i, u := range r.Units {
			if !slices.Contains(r.Gone, i) {
				b.Units[i] = u
			}
		}
		return &b, nil
	}
	for len(cands) > 0 {
		state, err := after()
		if err != nil {
			return evs, err
		}
		best, bestScore := -1, 0
		for i, c := range cands {
			if sc, ok := p.gain2(state, c); ok && sc > bestScore {
				best, bestScore = i, sc
			}
		}
		if best < 0 {
			break
		}
		c := cands[best]
		ev, err := p.s.Attack(p.side, c.unit, c.target)
		if err != nil {
			return evs, err
		}
		evs = append(evs, ev...)
		orders = append(orders, dosrules.Order{Attacker: c.unit, Target: c.cell})
		cands = slices.DeleteFunc(cands, func(o cand) bool { return o.unit == c.unit })
	}
	return evs, nil
}

// rangedAttacker reports whether attacker shoots defender from beyond
// the next hex (displayed range >= 2 against the defender's domain):
// anti-tank guns, bunkers, light artillery and trains count, not only
// the long-range-flagged medium/heavy artillery.
func (p *coachPlan) rangedAttacker(attacker, defender int) bool {
	av, dv := p.s.Unit(attacker), p.s.Unit(defender)
	tv, ok := p.s.TypeInfo(p.side, av.Type)
	if !ok {
		return p.longRange(av)
	}
	r := tv.Ground.Range
	switch dv.Domain {
	case dosplay.Air:
		r = tv.Air.Range
	case dosplay.Naval:
		r = tv.Water.Range
	}
	return r >= 2
}

// gain2 values one attack on the (lookahead) battlefield: the exact
// armour-weighted exchange (Gain), plus the worth of what dies, plus
// role bonuses for the targets that matter (capturers, guns,
// aircraft-killers). Long-range attackers may not lose lives for a
// non-kill, and nobody trades its own destruction for less.
func (p *coachPlan) gain2(b *dosrules.Battlefield, c cand) (int, bool) {
	a, ok := b.Units[c.unit]
	if !ok {
		return 0, false
	}
	o, err := p.prof.Engage(b, c.unit, c.cell)
	if err != nil || !o.Battle {
		return 0, false
	}
	d := b.Units[o.Defender]
	attDead := slices.Contains(o.Destroyed, c.unit)
	defDead := slices.Contains(o.Destroyed, o.Defender)
	if p.rangedAttacker(c.unit, o.Defender) && o.AttackerSize < a.Size && !defDead {
		return 0, false // ranged units duel from range or not at all
	}
	if attDead && !defDead {
		return 0, false
	}
	sc := Gain(p.prof, b, c.unit, o)
	if defDead {
		sc += p.cost(d.Type)
	}
	if attDead {
		sc -= p.cost(a.Type)
	}
	dv := p.s.Unit(o.Defender)
	nearOwn := 999
	for _, b := range p.ownB {
		if d := hexgrid.Distance(c.target, b); d < nearOwn {
			nearOwn = d
		}
	}
	isCap := p.canCapture(dv)
	if isCap {
		sc += 60
		if nearOwn <= 4 {
			sc += 200 // a capturer at our door is the loss condition
		}
	}
	if p.longRange(dv) {
		sc += 40
	}
	for _, hq := range p.ownHQs {
		if hexgrid.Distance(c.target, hq) <= 3 {
			sc = sc * 3 / 2
			break
		}
	}
	// Defending the doorstep may trade at a small exact loss.
	desperate := isCap && nearOwn <= 3 && !attDead && sc > -80
	return sc, sc > 0 || desperate
}
