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
// destroy helper (recorded) and redraw helper, and the RNG seed callback at
// context +0x6a, which the game's initialisation binds to a routine that only
// stores the seed. The animation switch is set so that the handler takes the
// direct path; the animated path calls the same casualty routine once on
// the same context through 242a:000f.
func (e *Export) RunEngagement(g *Engagement) (Outcome, error) {
	b, h := e.B, e.B.Handler
	if h == nil {
		return Outcome{}, errors.New("the order handler is not located in this build")
	}
	m := e.machine()
	ar := newArena(m)
	data, code := baseSeg+b.Data, baseSeg+b.Code

	cells := g.cells()
	mp := make([]byte, 2*len(cells))
	for i, u := range cells {
		mp[2*i], mp[2*i+1] = byte(g.class(i)), byte(u) // terrain index = combat class
	}
	mapSeg := ar.alloc(mp)
	for k := range 2 {
		m.WW(data, h.Maps+4*k, 0)
		m.WW(data, h.Maps+4*k+2, mapSeg)
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
			m.WW(data, base+0x0B+2*k, 2*u.Cell+1)
		}
	}
	m.WW(data, b.Units+26*g.Attacker+0x11, 2*g.Target)

	order := h.Orders + 0x17*g.Side
	m.WW(data, order, 0)
	m.WW(data, order+2, ar.alloc([]byte{byte(g.Attacker)}))
	m.WB(data, order+4, 1)
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
	m.WW(data, h.Ctx+0x6A, 0)
	m.WW(data, h.Ctx+0x6C, scratch+0x100)
	m.Hooks[dos8086.Lin(scratch+0x100, 0)] = noHook

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
	var destroyed []int
	m.Hooks[dos8086.Lin(code, b.ExpDestroy)] = func(mm *dos8086.Machine) error {
		// "push cs; call near": the unit index lies above the pushed CS
		destroyed = append(destroyed, mm.RB(int(mm.S[dos8086.SS]), int(mm.R[dos8086.SP])+2))
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

	ctx := h.Ctx
	att, def := m.RB(data, ctx+8), m.RB(data, ctx+9)
	live := func(idx, off int) int { return m.RB(data, b.Units+26*idx+off) }
	return Outcome{
		Battle: true, Defender: def,
		ReturnFire: m.RW(data, ctx+0x2E) != 0,
		Pincer:     m.RB(data, ctx+0x30), Blocking: m.RB(data, ctx+0x31),
		Result:      Result{live(att, 2), live(att, 3), live(def, 2), live(def, 3)},
		AttackerExp: live(att, 1), DefenderExp: live(def, 1),
		Destroyed: destroyed,
	}, nil
}
