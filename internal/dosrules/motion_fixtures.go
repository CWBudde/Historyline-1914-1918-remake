package dosrules

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
)

// The movement, step, container and visibility fixtures (testdata/, numbers
// only) use the movement tables of the build's profile (Tables.Move). The
// tests of this package replay them.

// FixturePiece is a live unit of a movement fixture.
type FixturePiece struct {
	Index int `json:"index"`
	Piece
}

// FixturePieces lists the units of a map in index order.
func FixturePieces(m map[int]Piece) []FixturePiece {
	out := []FixturePiece{}
	for _, i := range slices.Sorted(maps.Keys(m)) {
		out = append(out, FixturePiece{i, m[i]})
	}
	return out
}

// FixtureMap is the map and the moving unit of a movement or step fixture.
type FixtureMap struct {
	Width  int            `json:"width"`
	Height int            `json:"height"`
	Fill   int            `json:"fill"`  // terrain record of every cell not in Cells
	Cells  [][2]int       `json:"cells"` // [cell, terrain record]
	Units  []FixturePiece `json:"units"`
	Unseen []FixturePiece `json:"unseen"`
	Unit   int            `json:"unit"`
	Side   int            `json:"side"`
	Budget int            `json:"budget"`
	Block  int            `json:"block"`
}

// MapOf is the fixture form of a movement query; Fill is its most common
// terrain record.
func MapOf(g *Movement) FixtureMap {
	f := FixtureMap{Width: g.Width, Height: g.Height, Cells: [][2]int{}, Units: FixturePieces(g.Units),
		Unseen: FixturePieces(g.Unseen), Unit: g.Unit, Side: g.Side, Budget: g.Budget, Block: g.Block}
	counts := map[int]int{}
	for _, t := range g.Terrain {
		counts[t]++
	}
	for t, n := range counts {
		if n > counts[f.Fill] || n == counts[f.Fill] && t < f.Fill {
			f.Fill = t
		}
	}
	for cell, t := range g.Terrain {
		if t != f.Fill {
			f.Cells = append(f.Cells, [2]int{cell, t})
		}
	}
	return f
}

// Movement rebuilds the map of a fixture.
func (f FixtureMap) Movement() *Movement {
	g := &Movement{Width: f.Width, Height: f.Height, Terrain: make([]int, f.Width*f.Height),
		Units: map[int]Piece{}, Unseen: map[int]Piece{}, Unit: f.Unit, Side: f.Side,
		Budget: f.Budget, Block: f.Block}
	for i := range g.Terrain {
		g.Terrain[i] = f.Fill
	}
	for _, c := range f.Cells {
		g.Terrain[c[0]] = c[1]
	}
	for _, p := range f.Units {
		g.Units[p.Index] = p.Piece
	}
	for _, p := range f.Unseen {
		g.Unseen[p.Index] = p.Piece
	}
	return g
}

// MovementExpected is what the executed movement queries gave.
type MovementExpected struct {
	// Remaining is the cost buffer after the reach search, one signed byte
	// per cell in cell order: the budget left on reaching it, or below 0 the
	// negated entry cost, -80 closed, -102/-101 a zone of control of side
	// 0/1 that was not reached, -1 untouched.
	Remaining []int `json:"remaining"`
	Count     int   `json:"count"`
	// Destinations is null where the end-cell check reaches its container
	// branch (an own transporter or a building cell), which the container
	// fixtures record.
	Destinations []int `json:"destinations"`
	// Codes holds [cell, end code] for every reached cell outside the
	// container branch.
	Codes [][2]int `json:"codes"`
	// Paths holds, per goal, the cells from the goal back to the start, or
	// null where the search failed.
	Paths [][]int `json:"paths"`
}

// MovementFixture is a recorded set of movement queries on one map.
type MovementFixture struct {
	Schema  int    `json:"schema"`
	Profile string `json:"profile"`
	ID      string `json:"id"`
	Source  Source `json:"source"`
	FixtureMap
	// Winter: the query reads the winter terrain table (MoveTables.Season).
	Winter   bool             `json:"winter,omitempty"`
	Goals    []int            `json:"goals"`
	Expected MovementExpected `json:"expected"`
}

// Query rebuilds the query of a fixture.
func (f MovementFixture) Query() *Movement {
	g := f.Movement()
	g.Goals = f.Goals
	return g
}

// FixtureRecord is a record after the run, its bytes in order.
type FixtureRecord struct {
	Index int   `json:"index"`
	Bytes []int `json:"bytes"`
}

func ints(b []byte) []int {
	out := make([]int, len(b))
	for i, v := range b {
		out[i] = int(v)
	}
	return out
}

// StepExpected is what an executed move order gave.
type StepExpected struct {
	Order    int             `json:"order"` // AX of the order code: 1, 9 for a two-cell ship, 0x12 refused
	Start    int             `json:"start"` // AX of the move start, 0 where the order was refused
	Messages [][2]int        `json:"messages"`
	Cells    []int           `json:"cells"` // the mover's cell after each step
	Live     []FixtureRecord `json:"live"`
	// Occupants holds [cell, unit] for every occupied cell of the side's map
	// and of the other side's map afterwards; the terrain bytes stay.
	Occupants [2][][2]int `json:"occupants"`
	Bits      [][2]int    `json:"bits"` // [row*64+column, value] of every non-zero byte
	Orders    []int       `json:"orders"`
	Count     int         `json:"count"`
	Container []int       `json:"container"` // container record 0
}

// StepExpectedOf is the recorded form of an outcome; it fails where a map's
// terrain bytes changed, which the fixture does not record.
func StepExpectedOf(g *Movement, r StepRun) (StepExpected, error) {
	x := StepExpected{Order: r.Order, Start: r.Start, Messages: r.Messages, Cells: r.Cells,
		Orders: r.Orders, Count: r.Count, Container: ints(r.Container)}
	if x.Messages == nil {
		x.Messages = [][2]int{}
	}
	if x.Cells == nil {
		x.Cells = []int{}
	}
	if x.Orders == nil {
		x.Orders = []int{}
	}
	for _, idx := range slices.Sorted(maps.Keys(r.Live)) {
		x.Live = append(x.Live, FixtureRecord{idx, ints(r.Live[idx])})
	}
	for k, mp := range r.Maps {
		x.Occupants[k] = [][2]int{}
		for c := range g.Width * g.Height {
			if int(mp[2*c]) != g.Terrain[c] {
				return x, fmt.Errorf("map %d: terrain of cell %d changed", k, c)
			}
			if mp[2*c+1] != 0xFF {
				x.Occupants[k] = append(x.Occupants[k], [2]int{c, int(mp[2*c+1])})
			}
		}
	}
	x.Bits = [][2]int{}
	for i, v := range r.Bits {
		if v != 0 {
			x.Bits = append(x.Bits, [2]int{i, int(v)})
		}
	}
	return x, nil
}

// StepFixture is a recorded move order.
type StepFixture struct {
	Schema  int    `json:"schema"`
	Profile string `json:"profile"`
	ID      string `json:"id"`
	Source  Source `json:"source"`
	FixtureMap
	Order    MoveOrder    `json:"order"`
	Expected StepExpected `json:"expected"`
}

// FixtureHold is a container record of a fixture.
type FixtureHold struct {
	Index int `json:"index"`
	Hold
}

// FixtureHolds lists container records in index order.
func FixtureHolds(m map[int]Hold) []FixtureHold {
	out := []FixtureHold{}
	for _, i := range slices.Sorted(maps.Keys(m)) {
		out = append(out, FixtureHold{i, m[i]})
	}
	return out
}

// ContainerExpected is what an executed container case gave: the step
// outcome, the container records that changed (all 0x1c bytes) and, for an
// exit, [cell, budget left] of every cell its reach search reached.
type ContainerExpected struct {
	StepExpected
	Holds     []FixtureRecord `json:"holds"`
	Remaining [][2]int        `json:"remaining"`
}

// ContainerExpectedOf is the recorded form of an outcome.
func ContainerExpectedOf(g *Movement, o ExitOrder, r ExitRun) (ContainerExpected, error) {
	x := ContainerExpected{Holds: []FixtureRecord{}, Remaining: [][2]int{}}
	var err error
	if x.StepExpected, err = StepExpectedOf(g, r.StepRun); err != nil {
		return x, err
	}
	x.Holds = changedHolds(HoldBytes(g), r.Holds)
	if o.Hold >= 0 {
		for c := range g.Width * g.Height {
			if v := r.Remaining[c/g.Width*64+c%g.Width]; v >= 0 {
				x.Remaining = append(x.Remaining, [2]int{c, int(v)})
			}
		}
	}
	return x, nil
}

func changedHolds(before, after map[int][]byte) []FixtureRecord {
	out := []FixtureRecord{}
	for i := range HoldRecords {
		if !bytes.Equal(before[i], after[i]) {
			out = append(out, FixtureRecord{i, ints(after[i])})
		}
	}
	return out
}

// ContainerFixture is a recorded load, unload or exit.
type ContainerFixture struct {
	Schema  int    `json:"schema"`
	Profile string `json:"profile"`
	ID      string `json:"id"`
	Source  Source `json:"source"`
	FixtureMap
	Held     []FixturePiece    `json:"held"`
	Records  []FixtureHold     `json:"records"`
	Order    ExitOrder         `json:"order"`
	Expected ContainerExpected `json:"expected"`
}

// Movement rebuilds the map, the held units and the records of a fixture.
func (f ContainerFixture) Movement() *Movement {
	g := f.FixtureMap.Movement()
	g.Held, g.Holds = map[int]Piece{}, map[int]Hold{}
	for _, p := range f.Held {
		g.Held[p.Index] = p.Piece
	}
	for _, h := range f.Records {
		g.Holds[h.Index] = h.Hold
	}
	return g
}

// FixtureSighted is a live unit of a visibility fixture.
type FixtureSighted struct {
	Index int `json:"index"`
	Sighted
}

// VisibilityExpected is what an executed phase change gave: both maps, the
// records of the listed units, the container records that changed, the
// per-side records, the order counts, the turn counter, the animation flags
// and the messages.
type VisibilityExpected struct {
	Maps     [2][]int        `json:"maps"`
	Live     []FixtureRecord `json:"live"`
	Holds    []FixtureRecord `json:"holds"`
	Players  [2][]int        `json:"players"`
	Orders   [2]int          `json:"orders"`
	Count    int             `json:"count"`
	Anim     int             `json:"anim"`
	Messages [][2]int        `json:"messages"`
}

// VisibilityExpectedOf is the recorded form of a phase change.
func VisibilityExpectedOf(g *PhaseChange, r PhaseRun) VisibilityExpected {
	x := VisibilityExpected{Live: []FixtureRecord{}, Orders: r.Orders, Count: r.Count,
		Anim: r.Anim, Messages: r.Messages}
	for k := range 2 {
		x.Maps[k], x.Players[k] = ints(r.Maps[k]), ints(r.Players[k])
	}
	for _, i := range slices.Sorted(maps.Keys(r.Live)) {
		x.Live = append(x.Live, FixtureRecord{i, ints(r.Live[i])})
	}
	x.Holds = changedHolds(ContainerRecords(g.Holds), r.Holds)
	return x
}

// VisibilityFixture is a recorded phase change.
type VisibilityFixture struct {
	Schema   int                `json:"schema"`
	Profile  string             `json:"profile"`
	ID       string             `json:"id"`
	Source   Source             `json:"source"`
	Width    int                `json:"width"`
	Height   int                `json:"height"`
	Maps     [2][]int           `json:"maps"`
	Units    []FixtureSighted   `json:"units"`
	Records  []FixtureHold      `json:"records"`
	Attacker int                `json:"attacker"`
	Phase    [2]int             `json:"phase"`
	Count    int                `json:"count"`
	Expected VisibilityExpected `json:"expected"`
}

// PhaseChange rebuilds the input of a fixture.
func (f VisibilityFixture) PhaseChange() *PhaseChange {
	g := &PhaseChange{Width: f.Width, Height: f.Height, Maps: f.Maps, Units: map[int]Sighted{},
		Holds: map[int]Hold{}, Attacker: f.Attacker, Phase: f.Phase, Count: f.Count}
	for _, u := range f.Units {
		g.Units[u.Index] = u.Sighted
	}
	for _, h := range f.Records {
		g.Holds[h.Index] = h.Hold
	}
	return g
}
