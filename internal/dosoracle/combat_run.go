package dosoracle

import "github.com/CWBudde/Historyline-1914-1918-remake/internal/dos8086"

// The casualty routine (2466:000e in the full build) reads a context record
// through the far pointer Build.Ctx (3ac4:0010). Offsets used (all traced in windows/combat.asm; names
// are descriptive, the record has other, unread fields):
//
//	+0x00 far attacker live unit        +0x04 far defender live unit
//	+0x0a far attacker type record      +0x0e far defender type record
//	+0x12 far attacker terrain record   +0x16 far defender terrain record
//	+0x1a/0x1c positions (to the distance helper, hooked here)
//	+0x2e return fire allowed (zero suppresses the attacker's losses)
//	+0x30 pincer percentage             +0x31 blocking percentage
//	+0x6a far RNG seed callback (hooked)
//
// Live unit: +1 experience, +2 size, +3 alternate size, +4 word flags (0x10
// air, 0x08 water target; otherwise ground). The distance helper Build.Distance
// (0e47:2414) is replaced by a fixed value, and the map width it divides by
// is set to 1.

// CombatRun is the executed routine's outcome.
type CombatRun struct {
	Result
	DistanceCalls, Steps int
}

// Battle executes the casualty routine from the relocated image on a
// synthetic context.
// a.Rec and d.Rec must be 69-byte type records.
func (e *Export) Battle(a, d Side, p Params) (CombatRun, error) {
	m := e.machine()
	calls := 0
	m.Hooks[e.lin(e.B.Distance)] = func(mm *dos8086.Machine) error {
		calls++
		mm.R[dos8086.AX] = uint16(p.Distance)
		return nil
	}
	m.Hooks[dos8086.Lin(scratch+0x100, 0)] = noHook // seed callback

	ar := newArena(m)
	side := func(s Side, terrain int) (live, typ, terr int) {
		l := make([]byte, 26)
		l[1], l[2], l[3] = byte(s.Exp), byte(s.Size), byte(s.Alt)
		l[4], l[5] = byte(s.Flg), byte(s.Flg>>8)
		t := make([]byte, 6)
		t[5] = byte(terrain)
		return ar.alloc(l), ar.alloc(s.Rec), ar.alloc(t)
	}
	aLive, aType, aTerr := side(a, p.ATerrain)
	dLive, dType, dTerr := side(d, p.DTerrain)
	ctx := make([]byte, 0x80)
	farPtr(ctx, 0x00, aLive)
	farPtr(ctx, 0x04, dLive)
	farPtr(ctx, 0x0A, aType)
	farPtr(ctx, 0x0E, dType)
	farPtr(ctx, 0x12, aTerr)
	farPtr(ctx, 0x16, dTerr)
	if p.ReturnFire {
		ctx[0x2E] = 1
	}
	ctx[0x30], ctx[0x31] = byte(p.Pincer), byte(p.Blocking)
	farPtr(ctx, 0x6A, scratch+0x100)
	c := ar.alloc(ctx)
	b := e.B
	m.WW(baseSeg+b.Ctx[0], b.Ctx[1], 0)
	m.WW(baseSeg+b.Ctx[0], b.Ctx[1]+2, c)
	m.WW(baseSeg+b.Data, b.Width, 1)

	m.S = [4]uint16{dos8086.ES: 0, dos8086.CS: uint16(baseSeg + b.Combat[0]),
		dos8086.SS: scratch - 0x400, dos8086.DS: uint16(baseSeg + b.CombatDS)}
	m.R[dos8086.SP] = 0x3F00
	m.Push(retSeg)
	m.Push(0)
	m.IP = uint16(b.Combat[1])
	if err := m.Run(retSeg, 0, 0); err != nil {
		return CombatRun{}, wrapRun("battle", err)
	}
	return CombatRun{
		Result:        Result{m.RB(aLive, 2), m.RB(aLive, 3), m.RB(dLive, 2), m.RB(dLive, 3)},
		DistanceCalls: calls, Steps: m.Steps,
	}, nil
}
