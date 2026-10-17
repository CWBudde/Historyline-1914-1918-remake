package dosoracle

import "github.com/CWBudde/Historyline-1914-1918-remake/internal/dos8086"

const supportMaxSteps = 200000

// RunExperienceWindow executes the experience window 02dc:0908-0abf with its
// near-call helpers hooked. alt says whether each side's type record uses the
// alternate strength byte. It returns the two experience values and the
// number of destroy calls.
func (e *Export) RunExperienceWindow(a, d ExpSide, alt [2]bool) (attacker, defender, destroyed int, err error) {
	m := e.machine()
	ar := newArena(m)
	side := func(s ExpSide, useAlt bool) (live, typ int) {
		l := make([]byte, 26)
		l[1] = byte(s.Exp)
		if useAlt {
			l[3] = byte(s.After)
		} else {
			l[2] = byte(s.After)
		}
		l[6] = byte(s.Flags6)
		rec := make([]byte, recordLen)
		if useAlt {
			rec[0x0F] = 4
		}
		return ar.alloc(l), ar.alloc(rec)
	}
	aLive, aType := side(a, alt[0])
	dLive, dType := side(d, alt[1])
	ctx := make([]byte, 0x20)
	farPtr(ctx, 0, aLive)
	farPtr(ctx, 4, dLive)
	farPtr(ctx, 0x0A, aType)
	farPtr(ctx, 0x0E, dType)
	c := ar.alloc(ctx)

	b := e.B
	code := baseSeg + b.Code
	m.Hooks[dos8086.Lin(code, b.ExpDestroy)] = func(mm *dos8086.Machine) error {
		destroyed++
		return nil
	}
	m.Hooks[dos8086.Lin(code, b.ExpRefresh)] = noHook

	ss := scratch - 0x400
	m.S = [4]uint16{dos8086.CS: uint16(code), dos8086.DS: uint16(baseSeg + b.DS), dos8086.SS: uint16(ss)}
	const bp = 0x3F00
	m.R[dos8086.BP], m.R[dos8086.SP], m.R[dos8086.DI] = bp, bp-0x60, 0
	m.WW(ss, bp-0x0C, 0)
	m.WW(ss, bp-0x0A, c)
	m.WB(ss, bp-0x1D, a.Before)
	m.WB(ss, bp-0x1E, d.Before)
	m.IP = uint16(b.ExpFirst)
	if err := m.Run(code, b.ExpLast, 0); err != nil {
		return 0, 0, 0, wrapRun("experience window", err)
	}
	return m.RB(aLive, 1), m.RB(dLive, 1), destroyed, nil
}
