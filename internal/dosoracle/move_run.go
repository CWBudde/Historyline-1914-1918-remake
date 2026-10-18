package dosoracle

import (
	"errors"
	"fmt"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dos8086"
)

const moveMaxSteps = 20_000_000

// guard fills the bytes the runner checks the routines leave alone.
const guard = 0xA5

// MoveRun is the executed outcome of one movement query.
type MoveRun struct {
	Buffer       []int8      // the cost buffer after the reach search, bufLen bytes
	Bits         []byte      // the reach-bit array at Move.Bits
	Count        int         // AX of the reach search
	Destinations []int       // cells, in the order Destinations lists them
	Codes        map[int]int // end-cell check of every reached cell
	Paths        [][]int     // per goal: cells from the goal back to the start, nil if not found
}

// RunMovement executes the movement queries on a synthetic map in the order
// the computer player calls them (1fe6, 23be): the reach search 0b06:0331
// (which calls the cost routine 0b06:09c1), the destination list 0b06:1383,
// which calls the end-cell check 10cc:0f74 for every reached cell, and, with
// each goal, the path search 0b06:0e12. Codes holds the end-cell check of every
// reached cell, called directly with the action record set as Destinations
// sets it. Only the stack check is hooked; the routines call nothing that
// draws. The side's map holds Units, the other side's map Units and Unseen.
func (e *Export) RunMovement(g *Movement) (MoveRun, error) {
	v, err := e.moveEnv(g)
	if err != nil {
		return MoveRun{}, err
	}
	m, data, mv := v.m, v.data, e.B.Move
	start := g.piece(g.Unit).Cell
	var r MoveRun
	if r.Buffer, r.Bits, r.Count, err = v.reach(g); err != nil {
		return r, err
	}
	sideMap, costSeg, guarded, call := v.sideMap, v.cost, v.guarded, v.call

	outSeg := guarded(2 * 0x316)
	if _, err := call("destinations", mv.Destinations, 0, outSeg, g.Side, g.Unit); err != nil {
		return r, err
	}
	n := m.RW(data, mv.Count)
	for i := range n {
		r.Destinations = append(r.Destinations, m.RW(outSeg, 2*i)/2)
	}

	bit := 1
	if g.Side != 0 {
		bit = 2
	}
	r.Codes = map[int]int{}
	for c := range g.Width * g.Height {
		if int(r.Bits[c/g.Width*64+c%g.Width])&bit == 0 {
			continue
		}
		m.WB(data, mv.Action, g.Unit)
		code, err := call("end cell", mv.EndCell, 2*c, g.Side, 0, sideMap)
		if err != nil {
			return r, err
		}
		r.Codes[c] = code
	}

	// The order code (10cc:02dd) passes the side's buffer that the reach
	// search filled as the node pool, and the route lands behind it at
	// +0x1f40; each search starts from the buffer as the reach left it.
	pool := costSeg
	for _, goal := range g.Goals {
		for i := range bufLen {
			m.WB(pool, i, int(byte(r.Buffer[i])))
		}
		ax, err := call("path", mv.Path, 0, pool, g.Unit, 2*start, 2*goal, g.Side, 0, sideMap)
		if err != nil {
			return r, err
		}
		var route []int
		if ax == 0 {
			route = []int{}
			for i := range m.RW(data, mv.Count) {
				route = append(route, m.RW(pool, 0x1F40+2*i)/2)
			}
		} else if ax != 0xFFFF || m.RW(data, mv.Count) != 0 {
			return r, fmt.Errorf("path: result %#x with count %d", ax, m.RW(data, mv.Count))
		}
		r.Paths = append(r.Paths, route)
		for i := range bufLen {
			if x := m.RB(data, mv.Bits+i); byte(x) != r.Bits[i] {
				return r, errors.New("path: reach bits changed")
			}
		}
	}
	return r, nil
}

// moveEnv is a machine set up for the movement routines on a synthetic map.
type moveEnv struct {
	m                 *dos8086.Machine
	mv                *Move
	ar                *arena
	data              int
	sideMap, otherMap int // segments of the side's and the other side's map
	cost              int // the guarded cost buffer the reach search fills
	orders            int // the guarded order list of the side (stepEnv)
	units             int // offset of the live unit table
	call              func(what string, at [2]int, args ...int) (int, error)
	guarded           func(n int) int
}

// moveEnv loads the width and height, the type and terrain records, a zeroed
// bit array, the live units (liveRecords), the container records
// (holdBytes) and both maps (movementMaps), and points the map far pointers
// at them, the current map pointer at the side's.
func (e *Export) moveEnv(g *Movement) (*moveEnv, error) {
	b, mv := e.B, e.B.Move
	if mv == nil {
		return nil, errors.New("the movement routines are not located in this build")
	}
	if g.Width > 64 || g.Height*64 > bufLen {
		return nil, fmt.Errorf("map %dx%d exceeds the 64-column arrays", g.Width, g.Height)
	}
	m := e.machine()
	v := &moveEnv{m: m, mv: mv, ar: newArena(m), data: baseSeg + b.Data, units: b.Units}
	data := v.data

	m.WW(data, b.Width, g.Width)
	m.WW(data, b.Width+2, g.Height)
	for t := range len(e.Units) / recordLen {
		copy(m.Mem[dos8086.Lin(data, mv.Types+recordLen*t):], e.TypeRecord(t))
	}
	copy(m.Mem[dos8086.Lin(data, mv.Terrains):], e.Terrains)
	for i := range bufLen {
		m.WB(data, mv.Bits+i, 0)
	}
	for idx, rec := range liveRecords(g) {
		copy(m.Mem[dos8086.Lin(data, b.Units+liveLen*idx):], rec)
	}
	for idx, rec := range holdBytes(g) {
		copy(m.Mem[dos8086.Lin(data, mv.Containers+containerLen*idx):], rec)
	}
	own, other := movementMaps(g)
	v.sideMap, v.otherMap = v.ar.alloc(own), v.ar.alloc(other)
	m.WW(data, mv.Maps[g.Side&1], 0)
	m.WW(data, mv.Maps[g.Side&1]+2, v.sideMap)
	m.WW(data, mv.Maps[1-g.Side&1], 0)
	m.WW(data, mv.Maps[1-g.Side&1]+2, v.otherMap)
	m.WW(data, mv.Current, 0)
	m.WW(data, mv.Current+2, v.sideMap)

	v.guarded = func(n int) int {
		buf := make([]byte, n+16)
		for i := range buf {
			buf[i] = guard
		}
		return v.ar.alloc(buf)
	}
	v.cost = v.guarded(0x1F40 + 2*0x317)
	v.call = func(what string, at [2]int, args ...int) (int, error) {
		m.S = [4]uint16{dos8086.CS: uint16(baseSeg + at[0]), dos8086.DS: uint16(baseSeg + b.DS), dos8086.SS: scratch - 0x400}
		m.R[dos8086.SP] = 0x3F00
		for i := len(args) - 1; i >= 0; i-- {
			m.Push(args[i])
		}
		m.Push(retSeg)
		m.Push(0)
		m.IP = uint16(at[1])
		m.Steps = 0
		if err := m.Run(retSeg, 0, moveMaxSteps); err != nil {
			return 0, wrapRun(what, err)
		}
		return int(m.R[dos8086.AX]), nil
	}
	return v, nil
}

// reach runs the reach search 0b06:0331 as the player's order code calls it
// and reads the cost buffer, the bit array and the count.
func (v *moveEnv) reach(g *Movement) (buffer []int8, bits []byte, count int, err error) {
	m := v.m
	start := g.piece(g.Unit).Cell
	if count, err = v.call("reach", v.mv.Reach,
		2*start, 0, v.cost, g.Unit, g.Budget, g.Side, g.Block, 0, v.sideMap); err != nil {
		return nil, nil, 0, err
	}
	buffer = make([]int8, bufLen)
	for i := range bufLen {
		buffer[i] = int8(m.RB(v.cost, i))
	}
	for i := bufLen; i < bufLen+16; i++ {
		if m.RB(v.cost, i) != guard {
			return nil, nil, 0, errors.New("reach: wrote past the cost buffer")
		}
	}
	return buffer, v.bits(), count, nil
}

// bits reads the reach-bit array.
func (v *moveEnv) bits() []byte {
	bits := make([]byte, bufLen)
	for i := range bufLen {
		bits[i] = byte(v.m.RB(v.data, v.mv.Bits+i))
	}
	return bits
}
