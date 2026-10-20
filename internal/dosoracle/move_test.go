package dosoracle

import (
	"errors"
	"fmt"
	"maps"
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// moveTablesOf reads the movement fields of the export's type and terrain
// records.
func moveTablesOf(e *Export) *MoveTables {
	t := &MoveTables{}
	for i := range len(e.Units) / recordLen {
		r := e.TypeRecord(i)
		t.Types = append(t.Types, MoveType{Budget: int(r[0]), Speed: int(r[1]), Mask: int(r[5]),
			Class: word(r, 0x0D), Flags: word(r, 0x11), DistanceWeight: int(int8(r[0x15])), CostWeight: int(int8(r[0x16])),
			Capacity: int(r[0x40]), Weight: int(r[0x41])})
	}
	t.Terrains, t.Winter = moveTerrains(e.Terrains), moveTerrains(e.Winter)
	return t
}

func moveTerrains(b []byte) []MoveTerrain {
	var out []MoveTerrain
	for i := 0; i+6 <= len(b); i += 6 {
		r := b[i : i+6]
		out = append(out, MoveTerrain{Flags: word(r, 0), Mask: int(r[2]), Cost: int(r[3]), AirCost: int(r[4])})
	}
	return out
}

func needMove(t *testing.T, e *Export) {
	t.Helper()
	if e.B.Move == nil {
		t.Skip("the movement routines are not located in this build")
	}
}

// piece is a unit of type typ for a side (0, 1, or 2 for neutral) with its
// live flags as the constructor 159e:0006 sets them: +4 the type's +0x0d word
// ORed with the side bit, +6 the type's +0x11 word.
func piece(tab *MoveTables, typ, cell, side int) Piece {
	bit := [3]int{0, 1, 2}[side]
	return Piece{Type: typ, Cell: cell, Flags4: tab.Types[typ].Class | bit, Flags6: tab.Types[typ].Flags}
}

// playerBlock is the terrain flag mask the player's order code passes.
func playerBlock(p Piece) int {
	switch {
	case p.Flags6&1 != 0:
		return 0xFFFF
	case p.Flags4&1 != 0:
		return 6
	}
	return 3
}

// moveCase is one movement query of a coverage class.
type moveCase struct {
	class, name string
	g           *Movement
}

// uniform is a width x height map of one terrain with the unit of type typ
// for side 0 in the middle, at the type's own budget.
func uniform(tab *MoveTables, typ, terrain, width, height int) *Movement {
	g := &Movement{Width: width, Height: height, Terrain: make([]int, width*height), Units: map[int]Piece{}}
	for i := range g.Terrain {
		g.Terrain[i] = terrain
	}
	p := piece(tab, typ, height/2*width+width/2, 0)
	g.Units[0] = p
	g.Budget, g.Block = tab.Types[typ].Budget, playerBlock(p)
	return g
}

// movesAgree runs a query in the emulator and in the transcription and
// compares every output; it returns the emulator's outcome.
func movesAgree(t *testing.T, e *Export, tab *MoveTables, c moveCase) MoveRun {
	t.Helper()
	got, err := e.RunMovement(c.g)
	if err != nil {
		t.Fatalf("%s/%s: %v", c.class, c.name, err)
	}
	buf, bits, count := tab.Reach(c.g)
	if !slices.Equal(got.Buffer, buf) {
		for i := range buf {
			if got.Buffer[i] != buf[i] {
				t.Errorf("%s/%s: cost buffer at row %d column %d: executed %d, transcribed %d",
					c.class, c.name, i/64, i%64, got.Buffer[i], buf[i])
				break
			}
		}
	}
	if !slices.Equal(got.Bits, bits) || got.Count != count {
		t.Errorf("%s/%s: reach bits or count differ (count executed %d, transcribed %d)", c.class, c.name, got.Count, count)
	}
	for cell, code := range got.Codes {
		want, err := tab.EndCode(c.g, cell)
		if err != nil || want != code {
			t.Errorf("%s/%s: end code of cell %d: executed %#x, transcribed %#x (%v)", c.class, c.name, cell, code, want, err)
		}
	}
	if dest, err := tab.Destinations(c.g, bits); err == nil && !slices.Equal(dest, got.Destinations) {
		t.Errorf("%s/%s: destinations executed %v, transcribed %v", c.class, c.name, got.Destinations, dest)
	}
	for k, goal := range c.g.Goals {
		want, err := tab.Path(c.g, buf, bits, goal)
		if errors.Is(err, ErrUndefined) {
			continue
		}
		if !slices.Equal(want, got.Paths[k]) || (want == nil) != (got.Paths[k] == nil) {
			t.Errorf("%s/%s: path to %d executed %v, transcribed %v", c.class, c.name, goal, got.Paths[k], want)
		}
	}
	return got
}

// Terrain records of the demo's summer table used by name below; the
// semantic names are not established (R2 item 1), so they are described by
// their numbers.
const (
	tOpen     = 121 // flags 0x0008, mask 0x10, cost 5
	tCheap    = 105 // flags 0x2008, mask 0x10, cost 4
	tRail     = 85  // mask 0x02, cost 5
	tShallow  = 48  // mask 0x01, cost 6
	tDeep     = 47  // mask 0x04, cost 6
	tClosed   = 3   // flags 0x4028, mask 0x80, cost 255 (closed to air units too)
	tBuilding = 14  // flags 0x0e09 (0x400 building, flag bit 1), mask 0x10, cost 4
)

// Type records used by name: 0 infantry-like (mask 0x32), 22 a ground
// transporter (+0x0d 0x1004), 24 a train (mask 0x02), 27 a ship (mask 0x05),
// 29 a two-cell ship head (+0x0d 0xc8), 30 its rear (0x48), 2 an air unit
// (+0x0d 0x10).
const (
	uInfantry, uTransporter, uTrain, uShip, uShipHead, uShipRear, uAir = 0, 22, 24, 27, 29, 30, 2
)

// distinctTypes returns the first type record of every distinct combination
// of the fields the movement queries read; the capacity and weight only
// matter on a transporter's cell (container_test.go).
func distinctTypes(tab *MoveTables) []int {
	var out []int
	seen := map[MoveType]bool{}
	for i, ty := range tab.Types {
		k := ty
		k.Speed, k.Capacity, k.Weight = 0, 0, 0
		if !seen[k] {
			seen[k] = true
			out = append(out, i)
		}
	}
	return out
}

// distinctTerrains returns the first terrain record of every distinct record.
func distinctTerrains(tab *MoveTables) []int {
	var out []int
	seen := map[MoveTerrain]bool{}
	for i, tr := range tab.Terrains {
		if !seen[tr] {
			seen[tr] = true
			out = append(out, i)
		}
	}
	return out
}

// winterCases are the terrain cases on the winter table: every distinct type
// on every distinct terrain record whose winter cost differs from its summer
// one; nil without a winter table.
func winterCases(tab *MoveTables) []moveCase {
	wt := tab.Season(true)
	if wt == nil {
		return nil
	}
	var cs []moveCase
	seen := map[[2]MoveTerrain]bool{}
	for tr, w := range wt.Terrains {
		k := [2]MoveTerrain{tab.Terrains[tr], w}
		if w == k[0] || seen[k] {
			continue
		}
		seen[k] = true
		for _, ty := range distinctTypes(tab) {
			g := uniform(wt, ty, tr, 9, 9)
			cs = append(cs, moveCase{"winter", fmt.Sprintf("type %d terrain %d", ty, tr), withPaths(wt, g, pathLimit["terrain"])})
		}
	}
	return cs
}

func building(tab *MoveTables, terrain int) bool {
	return tab.Terrains[terrain].Flags&(0x400|0x100|0x40) != 0
}

// withPaths sets the goals to the transcribed destinations, as the order code
// only asks for a route to a legal destination; limit > 0 keeps that many,
// evenly spaced.
func withPaths(tab *MoveTables, g *Movement, limit int) *Movement {
	_, bits, _ := tab.Reach(g)
	dest, _ := tab.Destinations(g, bits)
	if limit > 0 && len(dest) > limit {
		kept := make([]int, limit)
		for i := range kept {
			kept[i] = dest[i*len(dest)/limit]
		}
		dest = kept
	}
	g.Goals = dest
	return g
}

// pathLimit caps the routes of the two large classes.
var pathLimit = map[string]int{"terrain": 6, "transition": 6}

// toward is the cell one step from cell in direction k (Ring order), or -1.
func toward(cell, k, width, height int) int { return Ring(cell, width, height)[k] }

// moveCases enumerates the movement input classes.
func moveCases(tab *MoveTables) []moveCase {
	var cs []moveCase
	add := func(class, name string, g *Movement) {
		cs = append(cs, moveCase{class, name, withPaths(tab, g, pathLimit[class])})
	}

	// Every distinct type on every distinct terrain: costs, masks, air costs,
	// budgets below a single step and first steps.
	for _, ty := range distinctTypes(tab) {
		for _, tr := range distinctTerrains(tab) {
			add("terrain", fmt.Sprintf("type %d terrain %d", ty, tr), uniform(tab, ty, tr, 9, 9))
		}
	}

	// Transitions: the left half of one terrain, the right half of another,
	// for one type of every distinct mask, between terrains of distinct mask
	// and cost.
	var reps []int
	seenMC := map[[2]int]bool{}
	for _, tr := range distinctTerrains(tab) {
		k := [2]int{tab.Terrains[tr].Mask, tab.Terrains[tr].Cost}
		if !building(tab, tr) && !seenMC[k] {
			seenMC[k] = true
			reps = append(reps, tr)
		}
	}
	var movers []int
	seenMask := map[[2]int]bool{}
	for _, ty := range distinctTypes(tab) {
		k := [2]int{tab.Types[ty].Mask, tab.Types[ty].Class & 0x10}
		if tab.Types[ty].Budget > 0 && !seenMask[k] {
			seenMask[k] = true
			movers = append(movers, ty)
		}
	}
	for _, ty := range movers {
		for _, a := range reps {
			if tab.Types[ty].Mask&tab.Terrains[a].Mask == 0 {
				continue
			}
			for _, b := range reps {
				if a == b {
					continue
				}
				g := uniform(tab, ty, a, 9, 9)
				for c := range g.Terrain {
					if c%9 >= 5 {
						g.Terrain[c] = b
					}
				}
				g.Units[0] = piece(tab, ty, 4*9+3, 0)
				add("transition", fmt.Sprintf("type %d terrain %d to %d", ty, a, b), g)
			}
		}
	}

	// Units: own, enemy and neutral units next to the mover and two cells
	// away in every direction, from an even and an odd column, for both sides.
	for _, side := range []int{0, 1} {
		for _, start := range []int{4*9 + 4, 4*9 + 3} {
			for k := range 6 {
				for _, rel := range []struct {
					name  string
					side  int
					steps int
				}{{"own", side, 1}, {"enemy", 1 - side, 1}, {"enemy at 2", 1 - side, 2}, {"neutral", 2, 1}} {
					g := uniform(tab, uInfantry, tOpen, 9, 9)
					g.Side = side
					g.Units[0] = piece(tab, uInfantry, start, side)
					g.Block = playerBlock(g.Units[0])
					at := start
					for range rel.steps {
						at = toward(at, k, 9, 9)
					}
					g.Units[1] = piece(tab, uInfantry, at, rel.side)
					add("units", fmt.Sprintf("side %d column %d %s direction %d", side, start%9, rel.name, k), g)
				}
			}
		}
	}
	{
		g := uniform(tab, uInfantry, tOpen, 9, 9)
		g.Units[1] = piece(tab, uInfantry, 2*9+6, 1)
		g.Units[2] = piece(tab, uInfantry, 6*9+6, 1)
		add("units", "two enemies with overlapping zones", g)

		g = uniform(tab, uInfantry, tCheap, 9, 9)
		for i, c := range []int{3*9 + 4, 3*9 + 5, 4*9 + 5} {
			g.Units[i+1] = piece(tab, uInfantry, c, 0)
		}
		g.Units[4] = piece(tab, uInfantry, 1*9+5, 1)
		add("units", "own units in front of an enemy", g)

		g = uniform(tab, uInfantry, tOpen, 9, 9)
		g.Units[1] = piece(tab, uTransporter, 4*9+5, 0)
		add("units", "own transporter", g)

		g = uniform(tab, uInfantry, tOpen, 9, 9)
		g.Units[0] = Piece{Type: uInfantry, Cell: 40, Flags4: tab.Types[uInfantry].Class, Flags6: tab.Types[uInfantry].Flags | 0x10}
		add("units", "flat cost (+6 bit 0x10)", g)

		g = uniform(tab, uShip, tShallow, 9, 9)
		g.Units[5] = piece(tab, uShipHead, 4*9+6, 1)
		g.Units[6] = piece(tab, uShipRear, 3*9+6, 1)
		add("units", "enemy two-cell ship", g)

		g = uniform(tab, uShipHead, tShallow, 9, 9)
		g.Units[1] = piece(tab, uShipRear, 3*9+4, 0)
		g.Units[2] = piece(tab, uShip, 4*9+6, 0)
		add("units", "two-cell ship among own ships", g)

		g = uniform(tab, uAir, tOpen, 9, 9)
		for c := range g.Terrain {
			if c%9 >= 5 {
				g.Terrain[c] = tClosed
			}
		}
		g.Units[1] = piece(tab, uInfantry, 2*9+3, 1)
		add("units", "air unit, enemy and closed terrain", g)
	}

	// Hidden units: an enemy on the other side's map only.
	for k := range 6 {
		g := uniform(tab, uInfantry, tOpen, 9, 9)
		g.Unseen = map[int]Piece{1: piece(tab, uInfantry, toward(40, k, 9, 9), 1)}
		add("unseen", fmt.Sprintf("enemy only on the other map, direction %d", k), g)
	}

	// The terrain flag mask: building terrain with flag bits 1, 2 and 4
	// around the mover, for each mask the callers pass.
	for _, block := range []int{3, 6, 0x1B, 0xFFFF} {
		g := uniform(tab, uInfantry, tOpen, 9, 9)
		for k, tr := range []int{12, 13, 14, 15, 16, 17} {
			g.Terrain[toward(40, k, 9, 9)] = tr
		}
		g.Block = block
		add("block", fmt.Sprintf("mask %#x", block), g)
	}

	// Budgets: from none to past the 0xb0 marker (a cost of 80) and the sign
	// of the byte, next to closed terrain.
	for _, budget := range []int{0, 1, 4, 5, 6, 10, 15, 16, 17, 30, 42, 79, 80, 81, 100, 127, 128, 200, 255} {
		g := uniform(tab, uInfantry, tOpen, 9, 9)
		g.Terrain[toward(40, 0, 9, 9)] = tDeep
		g.Budget = budget
		add("budget", fmt.Sprintf("budget %d", budget), g)
	}

	// Edges: the mover on every border cell of a 9 x 7 map, and maps one
	// cell wide or high.
	for c := range 9 * 7 {
		if c%9 != 0 && c%9 != 8 && c/9 != 0 && c/9 != 6 {
			continue
		}
		g := uniform(tab, uInfantry, tOpen, 9, 7)
		g.Units[0] = piece(tab, uInfantry, c, 0)
		add("edge", fmt.Sprintf("cell %d", c), g)
	}
	for _, wh := range [][2]int{{1, 9}, {9, 1}, {2, 2}, {64, 3}, {3, 68}} {
		g := uniform(tab, uTrain, tRail, wh[0], wh[1])
		add("edge", fmt.Sprintf("%dx%d", wh[0], wh[1]), g)
	}

	// Random maps of all non-building terrains with own, enemy and neutral
	// units.
	rng := rand.New(rand.NewSource(18))
	var open []int
	for i := range tab.Terrains {
		if !building(tab, i) {
			open = append(open, i)
		}
	}
	for n := range 200 {
		ty := movers[rng.Intn(len(movers))]
		g := uniform(tab, ty, open[0], 12, 12)
		for c := range g.Terrain {
			g.Terrain[c] = open[rng.Intn(len(open))]
		}
		g.Side = rng.Intn(2)
		g.Units[0] = piece(tab, ty, rng.Intn(144), g.Side)
		g.Block = playerBlock(g.Units[0])
		used := map[int]bool{g.Units[0].Cell: true}
		for i := 1; i <= rng.Intn(5); i++ {
			c := rng.Intn(144)
			if used[c] {
				continue
			}
			used[c] = true
			g.Units[i] = piece(tab, []int{uInfantry, uTrain, uShip, 10}[rng.Intn(4)], c, rng.Intn(3))
		}
		add("random", fmt.Sprint(n), g)
	}
	return cs
}

// TestMovementCoverage runs every movement case in the emulator and the
// transcription and checks the rules the cases are there to show.
func TestMovementCoverage(t *testing.T) { eachExport(t, movementCoverage) }

func movementCoverage(t *testing.T, e *Export) {
	needMove(t, e)
	tab := moveTablesOf(e)
	perClass := map[string]int{}
	took := map[string]time.Duration{}
	seen := map[string]int{}
	paths := 0
	for _, c := range moveCases(tab) {
		t0 := time.Now()
		r := movesAgree(t, e, tab, c)
		took[c.class] += time.Since(t0)
		perClass[c.class]++
		paths += len(r.Paths)
		moveRules(t, tab, c, r, seen)
		for i, v := range r.Buffer {
			if (i%64 >= c.g.Width || i/64 >= c.g.Height) && v != -1 {
				// The edge clips are else-ifs: a map one cell wide or high leaks
				// into column or row 1.
				if c.g.Width > 1 && c.g.Height > 1 {
					t.Fatalf("%s/%s: buffer written outside the map at %d", c.class, c.name, i)
				}
				t.Logf("%s/%s: buffer written outside the map at row %d column %d", c.class, c.name, i/64, i%64)
			}
		}
	}
	// The winter table: the reach must differ from the summer one somewhere,
	// or the cases would not show that the runner reads it.
	if w := e.winter(); w != nil {
		wt, differs := tab.Season(true), 0
		for _, c := range winterCases(tab) {
			r := movesAgree(t, w, wt, c)
			perClass[c.class]++
			moveRules(t, wt, c, r, seen)
			if buf, _, _ := tab.Reach(c.g); !slices.Equal(buf, r.Buffer) {
				differs++
			}
		}
		if differs == 0 {
			t.Error("no winter case reaches differently from summer")
		}
		t.Logf("winter: %d cases, %d reach differently from summer", perClass["winter"], differs)
	}
	for _, class := range []string{"terrain", "transition", "units", "unseen", "block", "budget", "edge", "random"} {
		if perClass[class] == 0 {
			t.Errorf("no %s cases", class)
		}
		t.Logf("%s: %d cases, %v", class, perClass[class], took[class])
	}
	t.Logf("%d paths", paths)
	for _, k := range []string{"spent against the terrain cost", "no first step", "own cell not a destination",
		"own unit passed", "zone of control: entered with 0 left", "unseen enemy cell is a destination",
		"blocked cell entered at budget >= 80", "blocked cell closed below 80", "budget >= 128 reaches nothing"} {
		if seen[k] == 0 {
			t.Errorf("no case shows: %s", k)
		}
	}
	for _, k := range slices.Sorted(maps.Keys(seen)) {
		t.Logf("%s: %d", k, seen[k])
	}
}

// moveRules checks, on the executed outcome, the rules the cases show and
// counts each one.
func moveRules(t *testing.T, tab *MoveTables, c moveCase, r MoveRun, seen map[string]int) {
	t.Helper()
	g := c.g
	at := func(cell int) int { return int(r.Buffer[cell/g.Width*64+cell%g.Width]) }
	start := g.Units[g.Unit].Cell
	costs := tab.Costs(g)
	cost := func(cell int) int { return -int(costs[cell/g.Width*64+cell%g.Width]) }
	fail := func(format string, args ...any) {
		t.Errorf("%s/%s: "+format, append([]any{c.class, c.name}, args...)...)
	}
	occ := g.Occupants()

	if code, ok := r.Codes[start]; ok && g.Units[g.Unit].Flags4&0x1000 == 0 {
		if code != 0x39 {
			fail("own cell has end code %#x", code)
		}
		seen["own cell not a destination"]++
	}
	// No guaranteed first step: a neighbour the unit may enter, but at a cost
	// above its whole budget, is not reached.
	if g.Budget > 0 && g.Budget < 128 {
		for _, n := range Ring(start, g.Width, g.Height) {
			if _, held := occ[n]; n >= 0 && !held && cost(n) > g.Budget && cost(n) < 80 {
				if at(n) >= 0 {
					fail("cell %d at cost %d reached with budget %d", n, cost(n), g.Budget)
				}
				seen["no first step"]++
			}
		}
	}
	if g.Budget >= 128 {
		if r.Count != 0 {
			fail("budget %d reaches %d cells", g.Budget, r.Count)
		}
		seen["budget >= 128 reaches nothing"]++
	}
	if c.class == "terrain" {
		for cell := range g.Width * g.Height {
			if cell == start || at(cell) < 0 {
				continue
			}
			if k := cost(cell); k <= 0 || (g.Budget-at(cell))%k != 0 {
				fail("cell %d keeps %d of %d at cost %d", cell, at(cell), g.Budget, k)
			}
			seen["spent against the terrain cost"]++
		}
	}
	if c.class == "budget" && g.Budget < 128 {
		deep := toward(start, 0, g.Width, g.Height)
		switch {
		case g.Budget >= 80 && at(deep) != g.Budget-80:
			fail("blocked cell keeps %d", at(deep))
		case g.Budget >= 80:
			seen["blocked cell entered at budget >= 80"]++
		case at(deep) >= 0:
			fail("blocked cell reached at budget %d", g.Budget)
		default:
			seen["blocked cell closed below 80"]++
		}
	}
	for idx, u := range g.Units {
		if idx == g.Unit {
			continue
		}
		switch {
		case u.Flags4&2 == 0 && dosrules.Owns(u.Flags4, g.Side):
			if at(u.Cell) >= 0 && r.Codes[u.Cell] == 0x39 && tab.Types[g.Units[g.Unit].Type].Class&0x40 == 0 {
				seen["own unit passed"]++
			}
		case u.Flags4&2 == 0:
			for _, n := range Ring(u.Cell, g.Width, g.Height) {
				if _, held := occ[n]; n < 0 || n == start || held || at(n) < 0 {
					continue
				}
				if at(n) != 0 {
					fail("cell %d next to enemy %d keeps %d", n, idx, at(n))
				}
				seen["zone of control: entered with 0 left"]++
			}
		}
	}
	for _, u := range g.Unseen {
		if at(u.Cell) >= 0 && r.Codes[u.Cell] == 0 && slices.Contains(r.Destinations, u.Cell) {
			seen["unseen enemy cell is a destination"]++
		}
	}
	for k, route := range r.Paths {
		if route == nil {
			seen["path not found"]++
			continue
		}
		// Price the route by the cost buffer before the search; a cell
		// holding a marker (a zone of control) is priced by its terrain.
		spent, zoc := 0, false
		for i, cell := range route[:len(route)-1] {
			k := cost(cell)
			if k > 80 {
				if i > 0 {
					zoc = true
				}
				k = stepCost(tab, g, cell)
			}
			spent += k
		}
		goal := g.Goals[k]
		switch {
		case zoc:
			seen["route through a zone of control"]++
		case at(goal) > 0 && spent > g.Budget:
			seen["route costs more than the budget"]++
		case at(goal) > 0 && spent > g.Budget-at(goal):
			seen["route dearer than the cheapest"]++
		case at(goal) > 0:
			seen["route as cheap as the reach"]++
		}
	}
}

// stepCost is the terrain cost of entering a cell in the mover's mode.
func stepCost(tab *MoveTables, g *Movement, cell int) int {
	u := g.Units[g.Unit]
	tr := tab.Terrains[g.Terrain[cell]]
	switch {
	case u.Flags6&0x10 != 0:
		return 2
	case u.Flags4&0x10 != 0:
		return tr.AirCost
	}
	return tr.Cost
}
