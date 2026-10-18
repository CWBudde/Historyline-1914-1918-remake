package dosoracle

import (
	"errors"
	"fmt"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dos8086"
)

// Order record fields the exit branch reads.
const (
	orderSlot = 0x0E // word: the slot the unit leaves
	orderHold = 0x24 // far pointer to the container record
)

// RunExit executes a unit leaving a container the way the main loop does it:
// the exit branch 0661:2f0a-30fe on a frame holding the order record (the
// slot and the record), then, with the destination set, the order code, the
// start and the steps as RunMove runs them. Hooked as in RunMove, and the
// view centring and view redraw; the order-record copy 0ca0:044c runs.
// Movement.Unit must be the unit in the slot.
func (e *Export) RunExit(g *Movement, o ExitOrder) (ExitRun, error) {
	v, r, err := e.stepEnv(g)
	if err != nil {
		return ExitRun{}, err
	}
	m, data, mv := v.m, v.data, v.mv
	rec := make([]byte, 0x30)
	putWord(rec, orderSlot, o.Slot)
	putWord(rec, orderHold, mv.Containers+containerLen*o.Hold)
	putWord(rec, orderHold+2, data)
	order := v.ar.alloc(rec)

	code := baseSeg + mv.Exit[0]
	ss := scratch - 0x400
	m.S = [4]uint16{dos8086.CS: uint16(code), dos8086.DS: uint16(baseSeg + e.B.DS), dos8086.SS: uint16(ss)}
	const bp = 0x3F00
	m.R[dos8086.BP], m.R[dos8086.SP], m.R[dos8086.SI] = bp, bp-0x60, uint16(g.Side)
	m.WW(ss, bp-0x38, 0)
	m.WW(ss, bp-0x36, order)
	m.IP = uint16(mv.Exit[1])
	m.Steps = 0
	if err := m.Run(code, mv.ExitEnd, moveMaxSteps); err != nil {
		return ExitRun{}, wrapRun("exit", err)
	}
	x := ExitRun{Unit: m.RW(order, orderUnit), Remaining: make([]int8, bufLen)}
	if x.Unit != g.Unit {
		return x, fmt.Errorf("exit: the slot holds unit %d, not Movement.Unit %d", x.Unit, g.Unit)
	}
	for i := range bufLen {
		x.Remaining[i] = int8(m.RB(v.cost, i))
	}
	for i := bufLen; i < bufLen+16; i++ {
		if m.RB(v.cost, i) != guard {
			return x, errors.New("exit: the reach search wrote past the cost buffer")
		}
	}
	x.StepRun, err = v.drive(g, o.MoveOrder, r, order)
	return x, err
}
