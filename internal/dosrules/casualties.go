package dosrules

import "errors"

// ErrDivideError is the 8086 divide error the casualty routine raises on a
// zero divisor (a zero-armour unit); the game would stop with a runtime error.
var ErrDivideError = errors.New("divide error")

// Target-flag and live-flag bits the casualty routine reads.
const (
	classBits   = 0x3C // target classes in TargetFlags, AttackMask and the live word
	waterTarget = 0x08 // live word of the target: a boat
	airTarget   = 0x10 // live word of the target: an aircraft
	longRange   = 0x40 // TargetFlags: never supports, never returns fire
	noReturn    = 0x80 // TargetFlags: never returns fire
)

// Fighter is one side of a battle: its type and its live values.
type Fighter struct {
	Type      Type
	Size, Alt int // live group strength and alternate strength
	Exp       int // experience, 0-6 in the game (read as an unsigned byte)
	Flags     int // live flag word +4: side bits 1/2, target class bits
}

// Context are the battle inputs the handler supplies.
type Context struct {
	Distance           int
	ATerrain, DTerrain int // terrain classes 0-8
	Pincer, Blocking   int // support percentages (pincer a byte, blocking at most 150)
	ReturnFire         bool
}

// Result holds the strengths after the battle.
type Result struct{ AttackerSize, AttackerAlt, DefenderSize, DefenderAlt int }

// i32 truncates to a signed 32-bit value, as the 32-bit runtime helpers do.
func i32(v int64) int64 { return int64(int32(v)) }

// quo is the runtime's signed 32-bit division, truncating toward zero.
func quo(a, b int64) int64 {
	if b == 0 {
		panic(ErrDivideError)
	}
	return i32(a / b)
}

// percent raises v by pct percent in 16.16 fixed point and keeps the signed
// high word: 20 raised by 150 is 49, not 50.
func percent(v, pct int64) int64 {
	fx := i32((v & 0xFFFF) << 16)
	return int64(int16(i32(fx+i32(quo(fx, 100)*pct)) >> 16))
}

// strengthTerm is base·size + base·F·size/div + (base/2)·term, F = (exp<<8)/6 + 1.
func strengthTerm(base, size, exp, term, div int64) int64 {
	f := quo(i32(exp<<8), 6) + 1
	return i32(i32(base*size) + quo(i32(i32(base*f)*size), div) + i32(quo(base, 2)*term))
}

// weapon is the power and encoded range a type uses against a target with the
// live flag word flags.
func weapon(t Type, flags int) (power, rng int64) {
	switch {
	case flags&airTarget != 0:
		return int64(t.AirPower), int64(t.AirRange)
	case flags&waterTarget != 0:
		return int64(t.WaterPower), int64(t.SurfaceRange)
	}
	return int64(t.GroundPower), int64(t.SurfaceRange)
}

// groupSize is the strength that fights: the alternate strength for such
// types, clamped to 1..6.
func groupSize(f Fighter) int64 {
	n := f.Size
	if f.Type.alternate() {
		n = f.Alt
	}
	return int64(min(max(n, 1), 6))
}

// firesBack reports whether a defender of type d returns the fire of an
// attacker of type a: only when the context allows it (next to the attacker),
// its type neither never returns fire nor is long range, and the attacker's
// classes include it (2451:0037-009e in the demo).
func firesBack(a, d Type, returnFire bool) bool {
	return returnFire && d.TargetFlags&(noReturn|longRange) == 0 && a.AttackMask&d.TargetFlags&classBits != 0
}

// Casualties resolves one battle as the casualty routine (2451:0004 in the
// demo) does. The routine makes no random draw. The defender fires back
// unless the context forbids it, its type never returns fire or is long
// range, or the attacker's classes do not include it.
func (p Profile) Casualties(a, d Fighter, c Context) (res Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); !ok || !errors.Is(e, ErrDivideError) {
				panic(r)
			}
			err = ErrDivideError
		}
	}()
	terms := p.Tables.Terms
	suppressed := !firesBack(a.Type, d.Type, c.ReturnFire)

	sizeA, sizeD := groupSize(a), groupSize(d)
	power, rng := weapon(a.Type, d.Flags)
	if rng > 1 {
		power = i32(power - quo(i32(3*i32(power*int64(c.Distance-1))), 4*rng))
	}
	counter, _ := weapon(d.Type, a.Flags)
	expA, expD := int64(a.Exp+1), int64(d.Exp+1)
	armourA, armourD := int64(a.Type.Armour), int64(d.Type.Armour)

	power = percent(power, int64(c.Pincer))
	armourD = percent(armourD, int64(c.Blocking))

	at, dt := terms[c.ATerrain], terms[c.DTerrain]
	attackA := strengthTerm(power, sizeA, expA, int64(at.Attack), 256)
	defenceA := strengthTerm(armourA, sizeA, expA, int64(at.Defence), 256)
	attackD := strengthTerm(counter, sizeD, expD, int64(dt.Attack), 256)
	defenceD := strengthTerm(armourD, sizeD, expD, int64(dt.Defence), 512)

	// Experience parameters: the attacker's is E, 2 below 1 (which a byte
	// cannot reach) and 6 above 7; the defender's is 2 up to 1 and 6 from 7.
	pa, pd := expA, expD
	switch {
	case expA < 1:
		pa = 2
	case expA > 7:
		pa = 6
	}
	switch {
	case expD <= 1:
		pd = 2
	case expD >= 7:
		pd = 6
	}
	lossD := quo(i32(attackA-defenceD), quo(i32(armourD*(8-pa)), 2))
	lossA := quo(i32(attackD-defenceA), quo(i32(armourA*(8-pd)), 2))

	// At least one loss when the attack beats the per-unit defence and half
	// the armour; 7-11 losses become 6; never more than twice the opponent's
	// strength, never below zero or above the own strength.
	if lossD <= 0 && i32(quo(defenceD, sizeD)+quo(armourD, 2)) < attackA {
		lossD = 1
	}
	if lossA <= 0 && i32(quo(defenceA, sizeA)+quo(armourA, 2)) < attackD {
		lossA = 1
	}
	if lossD > 6 && lossD < 12 {
		lossD = 6
	}
	if lossA > 6 && lossA < 12 {
		lossA = 6
	}
	lossD = min(max(min(lossD, 2*sizeA), 0), sizeD)
	lossA = min(max(min(lossA, 2*sizeD), 0), sizeA)

	res = Result{a.Size, a.Alt, d.Size, d.Alt}
	if !suppressed {
		res.AttackerSize, res.AttackerAlt = writeBack(a, sizeA-lossA)
	}
	res.DefenderSize, res.DefenderAlt = writeBack(d, sizeD-lossD)
	return res, nil
}

// writeBack stores the remaining strength as a byte: into the alternate
// strength for such types, where zero also clears the group strength.
func writeBack(f Fighter, left int64) (size, alt int) {
	v := int(left) & 0xFF
	if !f.Type.alternate() {
		return v, f.Alt
	}
	if v == 0 {
		return 0, 0
	}
	return f.Size, v
}
