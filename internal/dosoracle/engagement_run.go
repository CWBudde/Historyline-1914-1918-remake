package dosoracle

import (
	"errors"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dos8086"
)

const engagementMaxSteps = 2_000_000

// RunDistance executes the hex distance helper (Build.Distance) on two
// (row, column) pairs.
func (e *Export) RunDistance(r1, c1, r2, c2 int) (int, error) {
	m := e.machine()
	b := e.B
	m.S = [4]uint16{dos8086.CS: uint16(baseSeg + b.Distance[0]), dos8086.SS: scratch - 0x400}
	m.R[dos8086.SP] = 0x3F00
	for _, v := range [4]int{c2, r2, c1, r1} {
		m.Push(v)
	}
	m.Push(retSeg)
	m.Push(0)
	m.IP = uint16(b.Distance[1])
	if err := m.Run(retSeg, 0, 0); err != nil {
		return 0, wrapRun("distance", err)
	}
	return dos8086.Signed16(int(m.R[dos8086.AX])), nil
}

// RunEngagement executes the order handler (Build.Handler, 02dc:0004 in the
// demo) for one attack order and stops at the end of its first battle.
//
// The handler reads the order from its side's list, takes the defender from
// the unit byte of the target cell, fills the battle context (live units,
// type and terrain records, positions, return fire from the attacker's six
// neighbours via 0b06:0122), calls the support routine and the battle-screen
// routine, which stores the context pointer and calls the casualty routine,
// and finishes with the experience window. Hooked: the stack check, the
// drawing, scrolling and sound routines of Handler.Present, the battle
// screen (only its context store and casualty call are performed, the rest
// draws into the rectangle at Handler.Screen), the experience window's
// destroy helper (recorded) and redraw helper. The context's seed callback at
// +0x6a is bound to the game's seed routine, as its initialisation does, and
// the RNG draw is armed to fail with ErrRandomDraw. The animation switch is
// set so that the handler takes the direct path; the animated path calls the
// same casualty routine once on the same context through 242a:000f.
func (e *Export) RunEngagement(g *Engagement) (Outcome, error) { return e.runEngagement(g, nil) }

// ErrRandomDraw reports that the battle path called the game's RNG.
var ErrRandomDraw = errors.New("random draw in the battle path")

// runEngagement is RunEngagement with the RNG state set first and read back
// at the end (nil keeps the image's initial state).
func (e *Export) runEngagement(g *Engagement, state *uint32) (Outcome, error) {
	m, err := e.handlerFrame(g, []Order{{g.Attacker, g.Target}}, nil, state)
	if err != nil {
		return Outcome{}, err
	}
	b, h := e.B, e.B.Handler
	code := baseSeg + b.Code
	var destroyed []int
	m.Hooks[dos8086.Lin(code, b.ExpDestroy)] = func(mm *dos8086.Machine) error {
		// "push cs; call near": the unit index lies above the pushed CS
		destroyed = append(destroyed, mm.RB(int(mm.S[dos8086.SS]), int(mm.R[dos8086.SP])+2))
		return nil
	}
	for int(m.S[dos8086.CS]) != code || int(m.IP) != h.Stop {
		if int(m.S[dos8086.CS]) == code && int(m.IP) == h.Done {
			return Outcome{}, nil
		}
		if m.Steps++; m.Steps > engagementMaxSteps {
			return Outcome{}, wrapRun("order handler", dos8086.ErrStepLimit)
		}
		if err := m.Step(); err != nil {
			return Outcome{}, wrapRun("order handler", err)
		}
	}
	if state != nil {
		rs := baseSeg + h.RandState[0]
		*state = uint32(m.RW(rs, h.RandState[1])) | uint32(m.RW(rs, h.RandState[1]+2))<<16
	}
	o := e.outcome(m)
	o.Destroyed = destroyed
	return o, nil
}

// outcome reads the battle the handler has just finished from its context
// and the live records.
func (e *Export) outcome(m *dos8086.Machine) Outcome {
	b, h := e.B, e.B.Handler
	data := baseSeg + b.Data
	att, def := m.RB(data, h.Ctx+8), m.RB(data, h.Ctx+9)
	live := func(idx, off int) int { return m.RB(data, b.Units+26*idx+off) }
	return Outcome{
		Battle: true, Defender: def,
		ReturnFire: m.RW(data, h.Ctx+0x2E) != 0,
		Pincer:     m.RB(data, h.Ctx+0x30), Blocking: m.RB(data, h.Ctx+0x31),
		Result:      Result{live(att, 2), live(att, 3), live(def, 2), live(def, 3)},
		AttackerExp: live(att, 1), DefenderExp: live(def, 1),
	}
}

// handlerFrame loads a machine for a handler run on g's battlefield and
// leaves it at the handler's entry: the maps, the live units, the type and
// terrain records, the attacking side's order list (orders, in list order:
// each attacker's live +0x11 holds its target), the other side's list (idle,
// nil leaves it empty), the side selector, the RNG (state nil keeps the
// image's) and the hooks. The destroy helper is left to the caller.
func (e *Export) handlerFrame(g *Engagement, orders []Order, idle []int, state *uint32) (*dos8086.Machine, error) {
	b, h := e.B, e.B.Handler
	if h == nil {
		return nil, errors.New("the order handler is not located in this build")
	}
	m := e.machine()
	ar := newArena(m)
	data, code := baseSeg+b.Data, baseSeg+b.Code

	cells := g.cells()
	mp := make([]byte, 2*len(cells))
	for i, u := range cells {
		mp[2*i], mp[2*i+1] = byte(g.class(i)), byte(u) // terrain index = combat class
	}
	own := ar.alloc(mp)
	other := own // the other side's map
	if len(g.Moved) > 0 {
		mo := make([]byte, len(mp))
		for i := range cells {
			mo[2*i], mo[2*i+1] = byte(g.class(i)), 0xFF
		}
		for idx := range g.Units {
			mo[2*g.seen(idx)+1] = byte(idx)
		}
		other = ar.alloc(mo)
	}
	for k := range 2 { // Maps holds side 1's map, then side 0's
		view := other
		if 1-k == g.Side&1 {
			view = own
		}
		m.WW(data, h.Maps+4*k, 0)
		m.WW(data, h.Maps+4*k+2, view)
	}
	m.WW(data, b.Width, g.Width)
	m.WW(data, b.Width+2, g.Height)
	for t := range len(e.Units) / recordLen {
		copy(m.Mem[dos8086.Lin(data, h.Types+recordLen*t):], e.TypeRecord(t))
	}
	for c := range 9 {
		m.WB(data, h.Terrains+6*c+5, c)
	}
	for idx, u := range g.Units {
		base := b.Units + 26*idx
		m.WB(data, base+1, u.Exp)
		m.WB(data, base+2, u.Size)
		m.WB(data, base+3, u.Alt)
		m.WW(data, base+4, u.Flags4)
		m.WW(data, base+6, u.Flags6)
		m.WB(data, base+8, u.Type)
		for k := range 2 {
			cell := u.Cell
			if k != g.Side&1 {
				cell = g.seen(idx)
			}
			m.WW(data, base+0x0B+2*k, 2*cell+1)
		}
	}
	attackers := make([]byte, len(orders))
	for i, o := range orders {
		m.WW(data, b.Units+26*o.Attacker+0x11, 2*o.Target)
		attackers[i] = byte(o.Attacker)
	}
	list := func(side int, units []byte) {
		order := h.Orders + 0x17*side
		m.WW(data, order, 0)
		m.WW(data, order+2, ar.alloc(append(units, 0)))
		m.WB(data, order+4, len(units))
	}
	list(g.Side, attackers)
	if idle != nil {
		idx := make([]byte, len(idle))
		for i, u := range idle {
			idx[i] = byte(u)
		}
		list(1-g.Side, idx)
	}
	turn := 1
	if g.Side == 0 {
		turn = 2
	}
	m.WB(data, h.Turn, turn)
	m.WW(data, h.Flags, 0)
	m.WW(data, h.Tiles, 0)
	m.WW(data, h.Tiles+2, ar.alloc(make([]byte, 0x400)))
	m.WW(baseSeg+h.Screen[0], h.Screen[1], 0)
	m.WW(baseSeg+h.Screen[0], h.Screen[1]+2, ar.alloc(make([]byte, 0x40)))
	m.WB(baseSeg+h.Option[0], h.Option[1], 1)
	// The game's own seed routine; a draw anywhere in the battle path fails.
	m.WW(data, h.Ctx+0x6A, h.Srand[1])
	m.WW(data, h.Ctx+0x6C, baseSeg+h.Srand[0])
	m.Hooks[e.lin(h.Rand)] = errHook(ErrRandomDraw)
	if state != nil {
		rs := baseSeg + h.RandState[0]
		m.WW(rs, h.RandState[1], int(*state&0xFFFF))
		m.WW(rs, h.RandState[1]+2, int(*state>>16))
	}

	for _, p := range h.Present {
		m.Hooks[e.lin(p)] = noHook
	}
	m.Hooks[e.lin(h.Caller)] = func(mm *dos8086.Machine) error {
		ss, sp := int(mm.S[dos8086.SS]), int(mm.R[dos8086.SP])
		mm.WW(baseSeg+b.Ctx[0], b.Ctx[1], mm.RW(ss, sp+4))
		mm.WW(baseSeg+b.Ctx[0], b.Ctx[1]+2, mm.RW(ss, sp+6))
		if mm.RW(ss, sp+8) != 0 {
			return errors.New("battle screen called without the casualty routine")
		}
		cs, ip := mm.S[dos8086.CS], mm.IP
		mm.Push(retSeg)
		mm.Push(0)
		mm.S[dos8086.CS], mm.IP = uint16(baseSeg+b.Combat[0]), uint16(b.Combat[1])
		if err := mm.Run(retSeg, 0, engagementMaxSteps); err != nil {
			return err
		}
		mm.S[dos8086.CS], mm.IP = cs, ip
		return nil
	}
	m.Hooks[dos8086.Lin(code, b.ExpRefresh)] = noHook

	m.S = [4]uint16{dos8086.CS: uint16(code), dos8086.DS: uint16(baseSeg + b.DS), dos8086.SS: scratch - 0x400}
	m.R[dos8086.SP] = 0x3F00
	for range 10 {
		m.Push(0)
	}
	m.Push(retSeg)
	m.Push(0)
	m.IP = uint16(h.Entry)
	return m, nil
}
