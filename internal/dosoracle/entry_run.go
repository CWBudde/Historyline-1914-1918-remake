package dosoracle

import (
	"errors"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dos8086"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// Sizes of the player's order record and of its saved copy.
const (
	orderLen = 0x31 // the player's order record
	savedLen = 0x28 // the bytes 0ca0:044c copies
)

// Desk is the state the order-entry routines read and write.
type Desk struct {
	Live   map[int][]byte // the live record of every placed unit
	Maps   [2][]byte      // the side's map and the other side's
	Bits   []byte         // the bit array (dosrules.BufLen bytes)
	Holds  map[int][]byte // the container records
	Record []byte         // the player's order record
	Saved  []byte         // the saved order record (Cycle.Saved)
	// List is the side's order list (as many entries as its count byte).
	List                     []int
	Anim, Wait, Depots, Site int      // Move.Anim and Cycle's words
	Messages                 [][2]int // status messages: code, side
	Result                   int      // AX of the routine

	// The container records' offset and segment, which the order record
	// points into.
	HoldBase, HoldSeg int
}

// AttackOrder sets up an attack order: the cursor's cell, the bit-array
// bytes of the cells it gives (all others zero), the side's order count
// before, the depots under construction and the depot terrain index.
type AttackOrder struct {
	Cell         int
	Bits         map[int]int
	Count        int
	Depots, Site int
}

// CancelOrder sets up a cancel: the routine's mode argument (0x10 cancels),
// and either an exit (the exit branch runs first, then the cancel) or a
// planned move (the reach search runs, then the order code to Goal unless
// it is negative). The exit branch saves the order record (0ca0:044c) as
// the exit begins, so a cancel restores the record of that moment.
type CancelOrder struct {
	Mode int
	Exit *dosrules.ExitOrder
	Goal int
}

func (e *Export) entryEnv(g *Movement) (*moveEnv, *StepRun, *Cycle, error) {
	cy := e.B.Cycle
	if cy == nil {
		return nil, nil, nil, errors.New("the order-entry routines are not located in this build")
	}
	v, r, err := e.stepEnv(g)
	if err != nil {
		return nil, nil, nil, err
	}
	return v, r, cy, nil
}

// desk reads the state the entry routines share, with the order record at
// segment rec.
func (e *Export) desk(v *moveEnv, g *Movement, cy *Cycle, rec int, r *StepRun) Desk {
	m, data, mv := v.m, v.data, v.mv
	d := Desk{Live: map[int][]byte{}, Bits: v.bits(), Holds: map[int][]byte{},
		Record: make([]byte, orderLen), Saved: make([]byte, savedLen),
		Anim: m.RW(data, mv.Anim), Wait: m.RW(data, cy.Wait), Depots: m.RB(data, cy.Depots),
		Site: m.RB(data, cy.Site), Messages: append([][2]int{}, r.Messages...), Result: int(m.R[dos8086.AX]),
		HoldBase: mv.Containers, HoldSeg: data}
	for idx := range dosrules.LiveRecords(g) {
		d.Live[idx] = append([]byte{}, m.Mem[dos8086.Lin(data, v.units+dosrules.LiveLen*idx):][:dosrules.LiveLen]...)
	}
	for k, s := range [2]int{v.sideMap, v.otherMap} {
		d.Maps[k] = append([]byte{}, m.Mem[s*16:][:2*g.Width*g.Height]...)
	}
	for i := range dosrules.HoldRecords {
		d.Holds[i] = append([]byte{}, m.Mem[dos8086.Lin(data, mv.Containers+containerLen*i):][:containerLen]...)
	}
	copy(d.Record, m.Mem[rec*16:])
	copy(d.Saved, m.Mem[dos8086.Lin(data, cy.Saved):])
	for i := range m.RB(data, mv.Sides+sideLen*g.Side+sideOrderCount) {
		d.List = append(d.List, m.RB(v.orders, i))
	}
	return d
}

// RunAttackOrder executes the attack order (Cycle.Attack, 0ad1:000c in the
// demo) for Movement.Unit on the cursor's cell, as the menu calls it once the
// player picks a target (0ca0:0c38). Hooked: the stack check, the status
// message (recorded), the redraws and the timer. It returns the state before
// and after the call.
func (e *Export) RunAttackOrder(g *Movement, a AttackOrder) (Desk, Desk, error) {
	v, r, cy, err := e.entryEnv(g)
	if err != nil {
		return Desk{}, Desk{}, err
	}
	m, data, mv := v.m, v.data, v.mv
	for c, b := range a.Bits {
		m.WB(data, mv.Bits+c/g.Width*64+c%g.Width, b)
	}
	for i := range a.Count {
		m.WB(v.orders, i, 0xE0+i)
	}
	m.WB(data, mv.Sides+sideLen*g.Side+sideOrderCount, a.Count)
	m.WB(data, cy.Depots, a.Depots)
	m.WB(data, cy.Site, a.Site)
	m.WW(data, mv.Anim, 0)
	m.WW(data, cy.Wait, 0xAAAA)
	rec := make([]byte, orderLen)
	putWord(rec, 0, 2*a.Cell)
	putWord(rec, orderUnit, g.Unit)
	seg := v.ar.alloc(rec)

	before := e.desk(v, g, cy, seg, r)
	if _, err := v.call("attack order", cy.Attack, 0, seg, 0, v.sideMap, g.Side); err != nil {
		return before, Desk{}, err
	}
	return before, e.desk(v, g, cy, seg, r), e.listGuard(v)
}

// RunCancel executes the cancel (Cycle.Cancel, 10cc:0001 in the demo) after
// an exit or a planned move, as the menu calls it on the second press of the
// fire button (0ca0:0a0e). Hooked as RunAttackOrder, and the view redraws.
// It returns the state before and after the call.
func (e *Export) RunCancel(g *Movement, c CancelOrder) (Desk, Desk, error) {
	v, r, cy, err := e.entryEnv(g)
	if err != nil {
		return Desk{}, Desk{}, err
	}
	m, data, mv := v.m, v.data, v.mv
	m.WW(data, mv.Anim, 0)
	m.WW(data, cy.Wait, 0xAAAA)
	var seg int
	if c.Exit != nil {
		if seg, _, err = e.exitBranch(v, g, *c.Exit); err != nil {
			return Desk{}, Desk{}, err
		}
	} else {
		if _, _, _, err := v.reach(g); err != nil {
			return Desk{}, Desk{}, err
		}
		rec := make([]byte, orderLen)
		putWord(rec, orderUnit, g.Unit)
		seg = v.ar.alloc(rec)
		if c.Goal >= 0 {
			m.WW(seg, orderDest, 2*c.Goal)
			if _, err := v.call("order", mv.Order, 0, seg, 0, v.sideMap, g.Side); err != nil {
				return Desk{}, Desk{}, err
			}
		}
	}
	before := e.desk(v, g, cy, seg, r)
	if _, err := v.call("cancel", cy.Cancel, 0, seg, 0, v.sideMap, g.Side, c.Mode); err != nil {
		return before, Desk{}, err
	}
	return before, e.desk(v, g, cy, seg, r), e.listGuard(v)
}

func (e *Export) listGuard(v *moveEnv) error {
	for i := 16; i < 32; i++ {
		if v.m.RB(v.orders, i) != guard {
			return errors.New("wrote past the order list")
		}
	}
	return nil
}
