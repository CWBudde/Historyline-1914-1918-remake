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
	b, mv := e.B, e.B.Move
	if mv == nil {
		return MoveRun{}, errors.New("the movement routines are not located in this build")
	}
	if g.Width > 64 || g.Height*64 > bufLen {
		return MoveRun{}, fmt.Errorf("map %dx%d exceeds the 64-column arrays", g.Width, g.Height)
	}
	m := e.machine()
	ar := newArena(m)
	data := baseSeg + b.Data

	m.WW(data, b.Width, g.Width)
	m.WW(data, b.Width+2, g.Height)
	for t := range len(e.Units) / recordLen {
		copy(m.Mem[dos8086.Lin(data, mv.Types+recordLen*t):], e.TypeRecord(t))
	}
	copy(m.Mem[dos8086.Lin(data, mv.Terrains):], e.Terrains)
	for i := range bufLen {
		m.WB(data, mv.Bits+i, 0)
	}
	place := func(idx int, u Piece) {
		base := b.Units + 26*idx
		m.WW(data, base+4, u.Flags4)
		m.WW(data, base+6, u.Flags6)
		m.WB(data, base+8, u.Type)
		for k := range 2 {
			m.WW(data, base+0x0B+2*k, 2*u.Cell+1)
		}
	}
	own, other := make([]byte, 2*g.Width*g.Height), make([]byte, 2*g.Width*g.Height)
	for c, t := range g.Terrain {
		own[2*c], own[2*c+1] = byte(t), 0xFF
		other[2*c], other[2*c+1] = byte(t), 0xFF
	}
	for idx, u := range g.Units {
		place(idx, u)
		own[2*u.Cell+1], other[2*u.Cell+1] = byte(idx), byte(idx)
	}
	for idx, u := range g.Unseen {
		place(idx, u)
		other[2*u.Cell+1] = byte(idx)
	}
	m.WB(data, b.Units+26*g.Unit, g.Budget)
	sideMap, otherMap := ar.alloc(own), ar.alloc(other)
	m.WW(data, mv.Maps[g.Side&1], 0)
	m.WW(data, mv.Maps[g.Side&1]+2, sideMap)
	m.WW(data, mv.Maps[1-g.Side&1], 0)
	m.WW(data, mv.Maps[1-g.Side&1]+2, otherMap)

	guarded := func(n int) int {
		buf := make([]byte, n+16)
		for i := range buf {
			buf[i] = guard
		}
		return ar.alloc(buf)
	}
	costSeg := guarded(0x1F40 + 2*0x317)
	call := func(what string, at [2]int, args ...int) (int, error) {
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
	start := g.Units[g.Unit].Cell
	var r MoveRun
	var err error
	if r.Count, err = call("reach", mv.Reach,
		2*start, 0, costSeg, g.Unit, g.Budget, g.Side, g.Block, 0, sideMap); err != nil {
		return r, err
	}
	r.Buffer = make([]int8, bufLen)
	for i := range bufLen {
		r.Buffer[i] = int8(m.RB(costSeg, i))
	}
	for i := bufLen; i < bufLen+16; i++ {
		if m.RB(costSeg, i) != guard {
			return r, errors.New("reach: wrote past the cost buffer")
		}
	}
	r.Bits = make([]byte, bufLen)
	for i := range bufLen {
		r.Bits[i] = byte(m.RB(data, mv.Bits+i))
	}

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
			if v := m.RB(data, mv.Bits+i); byte(v) != r.Bits[i] {
				return r, errors.New("path: reach bits changed")
			}
		}
	}
	return r, nil
}
