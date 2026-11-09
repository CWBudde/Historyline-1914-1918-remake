package llmplay

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// QueryForms lists the queries Query answers.
const QueryForms = `dist A B | near A R | eta uN [l,c] | eta uN buildings | reach eN | threat [l,c] | threat uN | type ID|N|uN|eN | income   (A, B: [l,c], uN or eN)`

// setsUpFlag is Type.StrengthFlags bit 0: the type holds fire in the action
// phase after a move.
const setsUpFlag = 0x01

// Limits of the answers, which a player pays for in tokens.
const (
	queryHexes     = 15 // destinations listed in full
	queryNearest   = 8  // destinations listed of a longer set
	queryBuildings = 12 // buildings of eta uN buildings
	queryThreats   = 12 // attackers of threat
)

// Query answers one planning query of side as a few lines of text in the
// prompt's notation (uN own units, eN enemies where side's map shows them,
// [l,c] 1-based hexes). Like Observe it reads side's View and map only,
// never the other side's.
func Query(s *dosplay.Session, side int, q string) (string, error) {
	if side != 0 && side != 1 {
		return "", fmt.Errorf("query: side %d", side)
	}
	f := strings.Fields(q)
	x := &querier{s: s, side: side, v: s.View(side)}
	bad := func() (string, error) {
		return "", fmt.Errorf("unknown query %q; forms: %s", q, QueryForms)
	}
	if len(f) == 0 {
		return bad()
	}
	switch {
	case f[0] == "dist" && len(f) == 3:
		return x.dist(f[1], f[2])
	case f[0] == "near" && len(f) == 3:
		return x.near(f[1], f[2])
	case f[0] == "eta" && len(f) == 3:
		return x.eta(f[1], f[2])
	case f[0] == "reach" && len(f) == 2:
		return x.reach(f[1])
	case f[0] == "threat" && len(f) == 2:
		return x.threat(f[1])
	case f[0] == "type" && len(f) == 2:
		return x.typ(f[1])
	case f[0] == "income" && len(f) == 1:
		return x.income(), nil
	}
	return bad()
}

type querier struct {
	s    *dosplay.Session
	side int
	v    dosplay.View
}

// queryRef is a hex a query names, with the unit when it named one.
type queryRef struct {
	at   hexgrid.Coord
	unit *dosplay.UnitView
}

var (
	queryHex  = regexp.MustCompile(`^\[?(\d+),(\d+)\]?$`)
	queryUnit = regexp.MustCompile(`^([ue])(\d+)$`)
)

// label is the prompt's name of a unit: uN for side's own, eN otherwise.
func (x *querier) label(u dosplay.UnitView) string {
	if u.Side == x.side {
		return fmt.Sprintf("u%d", u.Index)
	}
	return fmt.Sprintf("e%d", u.Index)
}

// queryInt is the number of a run of digits, -1 for one too long.
func queryInt(digits string) int {
	n, err := strconv.Atoi(digits)
	if err != nil {
		return -1
	}
	return n
}

func (x *querier) ref(tok string) (queryRef, error) {
	if m := queryHex.FindStringSubmatch(tok); m != nil {
		at := parseCoord([2]int{queryInt(m[1]), queryInt(m[2])})
		if _, ok := x.s.Board.Cell(at); !ok {
			return queryRef{}, fmt.Errorf("hex %s is off the map (%d lines)", show(at), len(x.s.Level.Lines))
		}
		return queryRef{at: at}, nil
	}
	m := queryUnit.FindStringSubmatch(tok)
	if m == nil {
		return queryRef{}, fmt.Errorf("%q is no hex [l,c] and no unit uN or eN", tok)
	}
	n := queryInt(m[2])
	if m[1] == "u" {
		u := x.s.Unit(n)
		if u.Destroyed || u.Side != x.side {
			return queryRef{}, fmt.Errorf("no own unit u%d", n)
		}
		return queryRef{at: u.At, unit: &u}, nil
	}
	for _, at := range sortedCoords(x.v.Units) {
		if u := x.v.Units[at]; u.Index == n && u.Side != x.side && u.Head == nil {
			return queryRef{at: at, unit: &u}, nil
		}
	}
	return queryRef{}, fmt.Errorf("no enemy unit e%d on your map", n)
}

func (x *querier) dist(a, b string) (string, error) {
	ra, err := x.ref(a)
	if err != nil {
		return "", err
	}
	rb, err := x.ref(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d hexes %s -> %s\n", hexgrid.Distance(ra.at, rb.at), show(ra.at), show(rb.at)), nil
}

func (x *querier) near(a, radius string) (string, error) {
	ra, err := x.ref(a)
	if err != nil {
		return "", err
	}
	r, err := strconv.Atoi(radius)
	if err != nil || r < 0 {
		return "", fmt.Errorf("near: radius %q is no number of hexes", radius)
	}
	type row struct {
		d    int
		text string
	}
	var rows []row
	for _, at := range sortedCoords(x.v.Units) {
		u := x.v.Units[at]
		if d := hexgrid.Distance(ra.at, at); d <= r {
			rows = append(rows, row{d, fmt.Sprintf("%s %s %d/%d %s", x.label(u), u.ID, u.Size, u.Full, show(at))})
		}
	}
	for _, at := range sortedCoords(x.v.Buildings) {
		bv := x.v.Buildings[at]
		if d := hexgrid.Distance(ra.at, at); d <= r {
			rows = append(rows, row{d, fmt.Sprintf("%s %s %s", bv.Kind, Sides[x.s.SideOf(bv.Owner)], show(at))})
		}
	}
	if len(rows) == 0 {
		return fmt.Sprintf("nothing within %d of %s\n", r, show(ra.at)), nil
	}
	slices.SortStableFunc(rows, func(a, b row) int { return a.d - b.d })
	var b strings.Builder
	fmt.Fprintf(&b, "within %d of %s (distance first):\n", r, show(ra.at))
	for _, w := range rows {
		fmt.Fprintf(&b, " %d %s\n", w.d, w.text)
	}
	return b.String(), nil
}

// terrainReach is the reach search of the original (dosrules Reach) on entry
// costs alone: from start with budget, the hexes entered while the budget
// lasts, inside the search's window of budget cells around the start. A
// blocked hex is closed, a hex in zoc (next to an opposing unit) is entered
// whatever it costs and ends the move there. The result maps each reached
// hex, start included, to the budget left on it.
func (x *querier) terrainReach(costs map[hexgrid.Coord]int, start hexgrid.Coord, budget int, blocked, zoc map[hexgrid.Coord]bool) map[hexgrid.Coord]int {
	left := map[hexgrid.Coord]int{start: budget}
	if budget <= 0 {
		return left
	}
	w, h := x.s.Game.Width, x.s.Game.Height
	cell, _ := x.s.Board.Cell(start)
	colLo, rowLo := max(cell%w-budget/2, 0), max(cell/w-budget/2, 0)
	colHi, rowHi := min(colLo+budget, w), min(rowLo+budget, h)
	inWindow := func(c hexgrid.Coord) bool {
		k, ok := x.s.Board.Cell(c)
		return ok && k%w >= colLo && k%w < colHi && k/w >= rowLo && k/w < rowHi
	}
	levels := make([][]hexgrid.Coord, budget+1)
	levels[budget] = []hexgrid.Coord{start}
	for level := budget; level > 0; level-- {
		for _, c := range levels[level] {
			if !inWindow(c) {
				continue
			}
			for _, n := range hexgrid.Around(c) {
				cost, ok := costs[n]
				if _, seen := left[n]; !ok || seen || blocked[n] {
					continue
				}
				switch {
				case zoc[n]:
					left[n] = 0
				case cost <= level:
					left[n] = level - cost
					levels[level-cost] = append(levels[level-cost], n)
				}
			}
		}
	}
	return left
}

// typeBudget is the movement budget of a type, the one every unit of it
// starts a movement phase with.
func (x *querier) typeBudget(typ int) int {
	if m := x.s.Game.Profile.Tables.Move; m != nil && typ >= 0 && typ < len(m.Types) {
		return m.Types[typ].Budget
	}
	return 0
}

// etaField is a unit's way over the map in movement phases.
type etaField struct {
	phase map[hexgrid.Coord]int           // movement phases to end a move on the hex
	from  map[hexgrid.Coord]hexgrid.Coord // where the phase that reaches it starts
	exact bool                            // the first phase is the engine's destination set
}

// etaField counts the movement phases unit u needs to every hex, up to goal
// when given. The first phase is the unit's legal destinations now (its exits
// when it is held) where the session gives them. Every later phase, and the
// first one of a unit that cannot move now, is terrainReach with the type's
// budget on the terrain costs alone (Session.Costs): no unit blocks or stops
// the way, and a hex another unit holds now counts as a stop.
func (x *querier) etaField(u dosplay.UnitView, goal *hexgrid.Coord) (*etaField, error) {
	costs, err := x.s.Costs(x.side, u.Index)
	if err != nil {
		return nil, err
	}
	budget := x.typeBudget(u.Type)
	f := &etaField{phase: map[hexgrid.Coord]int{u.At: 0}, from: map[hexgrid.Coord]hexgrid.Coord{}}
	frontier := []hexgrid.Coord{u.At}
	var first map[hexgrid.Coord]bool
	if x.s.Phase(x.side) == dosrules.PhaseMovement && !u.Moved {
		if u.Held {
			first, err = x.s.ExitDestinations(x.side, u.In, u.Slot)
		} else {
			first, err = x.s.Destinations(x.side, u.Index)
		}
		f.exact = err == nil
	}
	for k := 1; len(frontier) > 0; k++ {
		if goal != nil {
			if _, ok := f.phase[*goal]; ok {
				break
			}
		}
		var next []hexgrid.Coord
		add := func(c, from hexgrid.Coord) {
			if _, ok := f.phase[c]; !ok {
				f.phase[c], f.from[c] = k, from
				next = append(next, c)
			}
		}
		if k == 1 && f.exact {
			for _, c := range dosplay.Coords(first) {
				add(c, u.At)
			}
		} else {
			if goal != nil {
				// Stops nearer the goal first, so a way advances early.
				slices.SortStableFunc(frontier, func(a, b hexgrid.Coord) int {
					return hexgrid.Distance(a, *goal) - hexgrid.Distance(b, *goal)
				})
			}
			for _, c := range frontier {
				for _, n := range sortedCoords(x.terrainReach(costs, c, budget, nil, nil)) {
					add(n, c)
				}
			}
		}
		frontier = next
	}
	return f, nil
}

// stops are the hexes the unit ends its movement phases on up to c.
func (f *etaField) stops(c hexgrid.Coord) []hexgrid.Coord {
	var out []hexgrid.Coord
	for f.phase[c] > 0 {
		out = append(out, c)
		c = f.from[c]
	}
	slices.Reverse(out)
	return out
}

// resolution is the resolution, counted from the current one as 1, in which
// the unit's k-th movement phase from now is played: a side moves every
// second resolution.
func (x *querier) resolution(u dosplay.UnitView, k int) int {
	switch {
	case x.s.Phase(x.side) != dosrules.PhaseMovement:
		return 2 * k
	case u.Moved:
		return 2*k + 1
	}
	return 2*k - 1
}

func (x *querier) etaText(u dosplay.UnitView, k int) string {
	if k == 0 {
		return "0 moves (there now)"
	}
	return fmt.Sprintf("%d moves (there after %d resolutions)", k, x.resolution(u, k))
}

// approach is the hex next to c the unit reaches first, for a hex it cannot
// enter.
func (f *etaField) approach(c hexgrid.Coord) (hexgrid.Coord, bool) {
	best, found := hexgrid.Coord{}, false
	for _, n := range hexgrid.Around(c) {
		if k, ok := f.phase[n]; ok && (!found || k < f.phase[best]) {
			best, found = n, true
		}
	}
	return best, found
}

func (x *querier) eta(unit, target string) (string, error) {
	ru, err := x.ref(unit)
	if err != nil {
		return "", err
	}
	if ru.unit == nil || ru.unit.Side != x.side {
		return "", fmt.Errorf("eta: %q is no own unit uN", unit)
	}
	u := *ru.unit
	note := ""
	if target == "buildings" {
		f, err := x.etaField(u, nil)
		if err != nil {
			return "", err
		}
		if !f.exact {
			note = ", terrain only"
		}
		return x.etaBuildings(u, f, note), nil
	}
	rt, err := x.ref(target)
	if err != nil {
		return "", err
	}
	f, err := x.etaField(u, &rt.at)
	if err != nil {
		return "", err
	}
	if !f.exact {
		note = " (terrain only: it cannot move now)"
	}
	head := fmt.Sprintf("u%d %s %s -> %s: ", u.Index, u.ID, show(u.At), show(rt.at))
	k, ok := f.phase[rt.at]
	if !ok {
		// The search stopped at no goal: the field is whole.
		if n, ok := f.approach(rt.at); ok {
			if f.phase[n] == 0 {
				return head + "cannot enter; it stands next to it\n", nil
			}
			return fmt.Sprintf("%scannot enter; next to it at %s in %s, stops %s%s\n", head, show(n), x.etaText(u, f.phase[n]), showAll(f.stops(n)), note), nil
		}
		return head + "unreachable\n", nil
	}
	if k == 0 {
		return head + x.etaText(u, 0) + "\n", nil
	}
	return fmt.Sprintf("%s%s, stops %s%s\n", head, x.etaText(u, k), showAll(f.stops(rt.at)), note), nil
}

func showAll(cs []hexgrid.Coord) string {
	parts := make([]string, len(cs))
	for i, c := range cs {
		parts[i] = show(c)
	}
	return strings.Join(parts, " ")
}

func (x *querier) etaBuildings(u dosplay.UnitView, f *etaField, note string) string {
	type row struct {
		k, d int
		text string
	}
	var rows []row
	for _, at := range sortedCoords(x.v.Buildings) {
		bv := x.v.Buildings[at]
		if x.s.SideOf(bv.Owner) == x.side {
			continue
		}
		name := fmt.Sprintf("%s %s %s", show(at), bv.Kind, Sides[x.s.SideOf(bv.Owner)])
		d := hexgrid.Distance(u.At, at)
		if k, ok := f.phase[at]; ok {
			text := name + " " + x.etaText(u, k)
			if st := f.stops(at); len(st) > 1 {
				text += " first " + show(st[0])
			}
			rows = append(rows, row{k, d, text})
		} else if n, ok := f.approach(at); ok {
			rows = append(rows, row{f.phase[n] + 1000, d, fmt.Sprintf("%s cannot enter; next to it at %s in %s", name, show(n), x.etaText(u, f.phase[n]))})
		} else {
			rows = append(rows, row{1 << 20, d, name + " unreachable"})
		}
	}
	slices.SortStableFunc(rows, func(a, b row) int { return cmp.Or(a.k-b.k, a.d-b.d) })
	var b strings.Builder
	fmt.Fprintf(&b, "u%d %s %s to foreign buildings (nearest first%s):\n", u.Index, u.ID, show(u.At), note)
	for i, w := range rows {
		if i == queryBuildings {
			fmt.Fprintf(&b, " (+%d more)\n", len(rows)-i)
			break
		}
		b.WriteString(" " + w.text + "\n")
	}
	return b.String()
}

// weaponOn is the weapon of an enemy type against a unit of domain d, as the
// prompt's "!" mark and the coach's threat envelope pick it.
func weaponOn(tv dosplay.TypeView, d dosplay.Domain) dosplay.Weapon {
	switch d {
	case dosplay.Air:
		return tv.Air
	case dosplay.Naval:
		return tv.Water
	}
	return tv.Ground
}

// enemyReach are the hexes enemy e may end one movement phase on, as far as
// side can tell. The session answers destinations only for a side's own
// units, so this is an approximation on side's map: terrainReach from e's
// last-known hex with its type's budget over the terrain costs for its type,
// side's and neutral units closing their hexes and side's units stopping it
// next to them, and no hex a unit of the map stands on as an end. It leaves
// out what side cannot know: where e and its fellows really stand, and
// loading into carriers and own buildings.
//
// The costs are Session.Costs on side's map. That query closes the hex of a
// unit of the other side and the hexes around it, so those seven are taken
// from the same query on e's side, which is terrain alone as well.
func (x *querier) enemyReach(e dosplay.UnitView) ([]hexgrid.Coord, error) {
	costs, err := x.s.Costs(x.side, e.Index)
	if err != nil {
		return nil, err
	}
	if theirs, err := x.s.Costs(e.Side&1, e.Index); err == nil {
		ring := hexgrid.Around(e.At)
		for _, c := range append(ring[:], e.At) {
			if v, ok := theirs[c]; ok {
				costs[c] = v
			}
		}
	}
	return x.viewReach(e, costs), nil
}

// viewReach is terrainReach for unit u among the units of side's view, as
// the original's cost query sets them: a unit of another side closes its
// hex, one of the opposing side stops u next to it, a carrier of u's side is
// passed at a cost of 2, and no hex a unit stands on is an end.
func (x *querier) viewReach(u dosplay.UnitView, costs map[hexgrid.Coord]int) []hexgrid.Coord {
	blocked, zoc := map[hexgrid.Coord]bool{}, map[hexgrid.Coord]bool{}
	for at, o := range x.v.Units {
		switch {
		case o.Side != u.Side:
			blocked[at] = true
			if o.Side+u.Side == 1 {
				for _, n := range hexgrid.Around(at) {
					zoc[n] = true
				}
			}
		case o.Carrier && o.Index != u.Index:
			if _, ok := costs[at]; ok {
				costs[at] = 2
			}
		}
	}
	var out []hexgrid.Coord
	for _, c := range sortedCoords(x.terrainReach(costs, u.At, x.typeBudget(u.Type), blocked, zoc)) {
		if _, taken := x.v.Units[c]; !taken && c != u.At {
			out = append(out, c)
		}
	}
	return out
}

// ownOnMap are side's units on the map, by record.
func (x *querier) ownOnMap() []dosplay.UnitView {
	var out []dosplay.UnitView
	for _, at := range sortedCoords(x.v.Units) {
		if u := x.v.Units[at]; u.Side == x.side {
			out = append(out, u)
		}
	}
	slices.SortStableFunc(out, func(a, b dosplay.UnitView) int { return a.Index - b.Index })
	return out
}

// firesAfterMove reports whether a type may attack in the action phase
// after it moved.
func (x *querier) firesAfterMove(typ int) bool {
	return x.s.Game.Profile.Tables.Types[typ].StrengthFlags&setsUpFlag == 0
}

func (x *querier) reach(enemy string) (string, error) {
	re, err := x.ref(enemy)
	if err != nil {
		return "", err
	}
	if re.unit == nil || re.unit.Side == x.side {
		return "", fmt.Errorf("reach: %q is no enemy unit eN", enemy)
	}
	e := *re.unit
	dests, err := x.enemyReach(e)
	if err != nil {
		return "", err
	}
	own := x.ownOnMap()
	var b strings.Builder
	fmt.Fprintf(&b, "e%d %s %s budget %d: %d hexes in one move (terrain estimate on your map)", e.Index, e.ID, show(e.At), x.typeBudget(e.Type), len(dests))
	list := dests
	if len(dests) > queryHexes {
		// The ones that matter: nearest to side's units and buildings.
		var marks []hexgrid.Coord
		for _, u := range own {
			marks = append(marks, u.At)
		}
		for at, bv := range x.v.Buildings {
			if x.s.SideOf(bv.Owner) == x.side {
				marks = append(marks, at)
			}
		}
		near := func(c hexgrid.Coord) int {
			best := 1 << 20
			for _, m := range marks {
				best = min(best, hexgrid.Distance(c, m))
			}
			return best
		}
		list = slices.Clone(dests)
		slices.SortStableFunc(list, func(a, b hexgrid.Coord) int { return near(a) - near(b) })
		list = list[:queryNearest]
		fmt.Fprintf(&b, ", the %d nearest to you", queryNearest)
	}
	if len(list) > 0 {
		b.WriteString(": " + showAll(list))
	}
	b.WriteByte('\n')
	for _, c := range dests {
		if bv, ok := x.v.Buildings[c]; ok && x.s.SideOf(bv.Owner) != e.Side {
			fmt.Fprintf(&b, " can enter %s %s %s\n", show(c), bv.Kind, Sides[x.s.SideOf(bv.Owner)])
		}
	}
	tv, _ := x.s.TypeInfo(e.Side&1, e.Type)
	var now, later []string
	for _, u := range own {
		w := weaponOn(tv, u.Domain)
		if w.Power == 0 || x.s.GasPioneer(e.Index) {
			continue
		}
		if d := hexgrid.Distance(e.At, u.At); d >= 1 && d <= w.Range {
			now = append(now, fmt.Sprintf("u%d %s", u.Index, show(u.At)))
			continue
		}
		if !x.firesAfterMove(e.Type) {
			continue
		}
		for _, c := range dests {
			if d := hexgrid.Distance(c, u.At); d >= 1 && d <= w.Range {
				later = append(later, fmt.Sprintf("u%d %s from %s", u.Index, show(u.At), show(c)))
				break
			}
		}
	}
	if len(now) > 0 {
		fmt.Fprintf(&b, " in range now: %s\n", strings.Join(now, ", "))
	}
	if len(later) > 0 {
		fmt.Fprintf(&b, " in range after its move: %s\n", strings.Join(later, ", "))
	}
	if len(now)+len(later) == 0 {
		b.WriteString(" none of your units in its range, now or after its move\n")
	}
	return b.String(), nil
}

// attacker is an enemy that may attack a hex: from where it stands (now) or
// from the hex it moves to first.
type attacker struct {
	e      dosplay.UnitView
	from   hexgrid.Coord
	now    bool
	o      dosrules.Outcome // against the own unit on the hex, if any
	fought bool
	alts   []hexgrid.Coord // every hex it may attack from after a move, the hardest hit first
}

func (x *querier) threat(target string) (string, error) {
	rt, err := x.ref(target)
	if err != nil {
		return "", err
	}
	var mine *dosplay.UnitView
	if u, ok := x.v.Units[rt.at]; ok && u.Side == x.side {
		mine = &u
	} else if rt.unit != nil {
		return "", fmt.Errorf("threat: %q is no own unit on the map", target)
	}
	domain := dosplay.Infantry
	if mine != nil {
		domain = mine.Domain
	}
	prof := x.s.Game.Profile
	bf, err := x.s.Battlefield(x.side) // side's own map: enemies where it last saw them
	if err != nil {
		return "", err
	}
	cell, _ := x.s.Board.Cell(rt.at)
	// engage is the exact outcome of e attacking the hex from a hex.
	engage := func(e dosplay.UnitView, from hexgrid.Coord) (dosrules.Outcome, bool) {
		if mine == nil {
			return dosrules.Outcome{}, false
		}
		was, ok := bf.Units[e.Index]
		fromCell, in := x.s.Board.Cell(from)
		if !ok || !in {
			return dosrules.Outcome{}, false
		}
		moved := was
		moved.Cell = fromCell
		bf.Units[e.Index] = moved
		o, err := prof.Engage(bf, e.Index, cell)
		bf.Units[e.Index] = was
		return o, err == nil && o.Battle
	}
	size := func(a attacker) int {
		if a.fought {
			return a.o.DefenderSize
		}
		return 1 << 20
	}
	var list []attacker
	for _, at := range sortedCoords(x.v.Units) {
		e := x.v.Units[at]
		if e.Side != 1-x.side || e.Head != nil || x.s.GasPioneer(e.Index) {
			continue
		}
		tv, ok := x.s.TypeInfo(e.Side, e.Type)
		w := weaponOn(tv, domain)
		if !ok || w.Power == 0 {
			continue
		}
		d := hexgrid.Distance(at, rt.at)
		if d >= 1 && d <= w.Range {
			a := attacker{e: e, from: at, now: true}
			a.o, a.fought = engage(e, at)
			list = append(list, a)
			continue
		}
		budget := x.typeBudget(e.Type)
		if !x.firesAfterMove(e.Type) || budget == 0 || d > budget+w.Range {
			continue
		}
		dests, err := x.enemyReach(e)
		if err != nil {
			continue
		}
		var opts []attacker
		for _, c := range dests {
			if dc := hexgrid.Distance(c, rt.at); dc < 1 || dc > w.Range {
				continue
			}
			a := attacker{e: e, from: c}
			a.o, a.fought = engage(e, c)
			opts = append(opts, a)
		}
		if len(opts) == 0 {
			continue
		}
		slices.SortStableFunc(opts, func(p, q attacker) int { return size(p) - size(q) })
		best := opts[0]
		for _, a := range opts {
			best.alts = append(best.alts, a.from)
		}
		list = append(list, best)
	}
	var b strings.Builder
	phase := "action: it fires this resolution from where it stands"
	if x.s.Phase(1-x.side) == dosrules.PhaseMovement {
		phase = "movement: it moves this resolution and fires in the next"
	}
	if mine != nil {
		fmt.Fprintf(&b, "u%d %s %d/%d %s", mine.Index, mine.ID, mine.Size, mine.Full, show(rt.at))
	} else {
		fmt.Fprintf(&b, "%s (as a ground unit)", show(rt.at))
	}
	fmt.Fprintf(&b, ": %d enemies in reach; enemy phase %s\n", len(list), phase)
	if len(list) == 0 {
		return b.String(), nil
	}
	slices.SortStableFunc(list, func(p, q attacker) int {
		if p.now != q.now {
			if p.now {
				return -1
			}
			return 1
		}
		return size(p) - size(q)
	})
	for i, a := range list {
		if i == queryThreats {
			fmt.Fprintf(&b, " (+%d more)\n", len(list)-i)
			break
		}
		fmt.Fprintf(&b, " e%d %s %d/%d %s", a.e.Index, a.e.ID, a.e.Size, a.e.Full, show(a.e.At))
		if a.now {
			b.WriteString(" now")
		} else {
			fmt.Fprintf(&b, " after move to %s", show(a.from))
		}
		if a.fought {
			fmt.Fprintf(&b, ": you %d->%d it %d->%d", mine.Size, a.o.DefenderSize, a.e.Size, a.o.AttackerSize)
			if slices.Contains(a.o.Destroyed, mine.Index) {
				b.WriteString(" DESTROYED")
			}
		}
		b.WriteByte('\n')
	}
	if mine == nil {
		return b.String(), nil
	}
	// Focus fire: the attacks one after another on the battlefield, as the
	// order handler fights a side's list (ResolveOrders).
	focus := func(title string, pick func(attacker) bool) {
		field := *bf
		field.Units = make(map[int]dosrules.Unit, len(bf.Units))
		for i, u := range bf.Units {
			field.Units[i] = u
		}
		taken := map[int]bool{}
		for _, u := range field.Units {
			taken[u.Cell] = true
		}
		var orders []dosrules.Order
		for _, a := range list {
			if !pick(a) || !a.fought {
				continue
			}
			if !a.now {
				to := -1
				for _, c := range a.alts {
					if k, ok := x.s.Board.Cell(c); ok && !taken[k] {
						to = k
						break
					}
				}
				if to < 0 {
					continue // the other attackers took its hexes
				}
				u := field.Units[a.e.Index]
				taken[to] = true
				u.Cell = to
				field.Units[a.e.Index] = u
			}
			orders = append(orders, dosrules.Order{Attacker: a.e.Index, Target: cell})
		}
		if len(orders) == 0 {
			return
		}
		r, err := prof.ResolveOrders(&field, orders)
		if err != nil {
			return
		}
		fmt.Fprintf(&b, " focus fire %s (%d attackers): you %d->%d", title, len(orders), mine.Size, r.Units[mine.Index].Size)
		if slices.Contains(r.Gone, mine.Index) {
			b.WriteString(" DESTROYED")
		}
		b.WriteByte('\n')
	}
	focus("now", func(a attacker) bool { return a.now })
	if slices.ContainsFunc(list, func(a attacker) bool { return !a.now }) {
		focus("after their move", func(attacker) bool { return true })
	}
	return b.String(), nil
}

func (x *querier) typ(id string) (string, error) {
	typ, tside := -1, x.side
	types := x.s.Game.Profile.Tables.Types
	if queryUnit.MatchString(id) {
		r, err := x.ref(id)
		if err != nil {
			return "", err
		}
		typ, tside = r.unit.Type, r.unit.Side&1
	} else if n, err := strconv.Atoi(strings.TrimPrefix(id, "type")); err == nil {
		typ = n
		if tv, _ := x.s.TypeInfo(x.side, n); tv.ID == "" {
			tside = 1 - x.side
		}
	} else {
	find:
		for _, k := range []int{x.side, 1 - x.side} {
			for t := range types {
				if tv, _ := x.s.TypeInfo(k, t); tv.ID == id {
					typ, tside = t, k
					break find
				}
			}
		}
	}
	tv, ok := x.s.TypeInfo(tside, typ)
	if !ok {
		return "", fmt.Errorf("no unit type %q (a type number, a unit id as the prompt prints it, uN or eN)", id)
	}
	t := types[typ]
	cost := "not built"
	if tv.Cost >= 0 {
		cost = strconv.Itoa(tv.Cost)
	}
	var flags []string
	if m := x.s.Game.Profile.Tables.Move; m != nil && typ < len(m.Types) {
		if m.Types[typ].Flags&captureFlag != 0 {
			flags = append(flags, "capturer")
		}
		if c := m.Types[typ].Capacity; c > 0 {
			flags = append(flags, fmt.Sprintf("carrier (capacity %d)", c))
		}
	}
	if t.TargetFlags&longRangeFlag != 0 {
		flags = append(flags, "long-range")
	}
	if t.StrengthFlags&setsUpFlag != 0 {
		flags = append(flags, "no attack after a move")
	}
	if t.TargetFlags&0x80 != 0 {
		flags = append(flags, "never returns fire")
	}
	if slices.Contains(x.s.Game.Profile.Builders, typ) {
		flags = append(flags, "builds depots")
	}
	if len(flags) == 0 {
		flags = []string{"none"}
	}
	weapon := func(name string, w dosplay.Weapon) string {
		if w.Power == 0 {
			return name + " -"
		}
		return fmt.Sprintf("%s power %d range %d", name, w.Power, w.Range)
	}
	return fmt.Sprintf("type%d %s %q %s: budget %d speed %d armour %d group %d weight %d cost %s\n weapons: %s, %s, %s\n flags: %s\n",
		typ, tv.ID, tv.Name, tv.Domain, x.typeBudget(typ), tv.Speed, tv.Armour, tv.Group, tv.Weight, cost,
		weapon("ground", tv.Ground), weapon("air", tv.Air), weapon("water", tv.Water), strings.Join(flags, ", ")), nil
}

// income lists what every building pays its owner at a phase change. Like
// the original's game info (Session.Stats) it shows both sides' accounts.
func (x *querier) income() string {
	st := x.s.Stats()
	var b strings.Builder
	rows := [3][]string{}
	for _, at := range sortedCoords(x.v.Buildings) {
		bv := x.v.Buildings[at]
		if c := x.s.Game.Buildings[bv.Hold].Credits; c > 0 {
			o := x.s.SideOf(bv.Owner)
			rows[o] = append(rows[o], fmt.Sprintf("%s %s +%d", show(at), bv.Kind, c))
		}
	}
	for _, k := range []int{x.side, 1 - x.side} {
		who := "you"
		if k != x.side {
			who = "enemy"
		}
		fmt.Fprintf(&b, "%s %s: energy %d, income %d per phase change: %s\n", who, Sides[k], st.Energy[k], st.Income[k], strings.Join(rows[k], ", "))
	}
	if len(rows[2]) > 0 {
		fmt.Fprintf(&b, "neutral (pays whoever takes it): %s\n", strings.Join(rows[2], ", "))
	}
	return b.String()
}
