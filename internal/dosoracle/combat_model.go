// Package dosoracle holds the research oracle for the DOS Historyline rules:
// independent transcriptions of routines of HL.EXE (this file and its
// siblings *_model.go, pure arithmetic) and runners that execute the real
// routines from a locally exported load image in the dos8086 interpreter.
// Agreement of the two shows the interpreter runs the routine as written; it
// does not show that the routine matches the running game.
package dosoracle

import "errors"

// ErrDivideError is a zero divisor in the transcribed 32-bit division.
var ErrDivideError = errors.New("divide error")

// wrap truncates to a signed 32-bit value.
func wrap(v int64) int64 { return int64(int32(v)) }

func mul(a, b int64) int64 { return wrap(a * b) }

// div is signed division truncating toward zero (0000:03dc). A zero divisor
// faults; the panic is recovered by Resolve.
func div(a, b int64) int64 {
	if b == 0 {
		panic(ErrDivideError)
	}
	return wrap(a / b)
}

func s16(v int64) int64 { return int64(int16(v)) }

func word(rec []byte, off int) int { return int(rec[off]) | int(rec[off+1])<<8 }

// adjust adds pct percent in 16.16 fixed point: (value<<16) plus
// trunc((value<<16)/100)*pct, then the high word, signed.
func adjust(value, pct int64) int64 {
	shifted := wrap((value & 0xFFFF) << 16)
	return s16(wrap(shifted+mul(div(shifted, 100), pct)) >> 16)
}

// terms is base*size + trunc(base*F*size/divisor) + trunc(base/2)*term with
// F = trunc((exp1<<8)/6)+1.
func terms(base, size, exp1, term, divisor int64) int64 {
	f := div(wrap(exp1<<8), 6) + 1
	return wrap(mul(base, size) + div(mul(mul(base, f), size), divisor) + mul(div(base, 2), term))
}

// powerFor returns power and encoded range of a record against a target with
// the given live flag word.
func powerFor(rec []byte, targetFlags int) (power, rng int64) {
	switch {
	case targetFlags&0x10 != 0:
		return int64(rec[10]), int64(rec[8])
	case targetFlags&0x08 != 0:
		return int64(rec[12]), int64(rec[9])
	}
	return int64(rec[11]), int64(rec[9])
}

// Side is one combatant: a 69-byte UNIT.DAT type record plus live values.
type Side struct {
	Rec                 []byte
	Size, Alt, Exp, Flg int
}

// Result holds the four strength bytes written back by the routine.
type Result struct{ AttackerSize, AttackerAlt, DefenderSize, DefenderAlt int }

// Params are the battle inputs other than the two sides.
type Params struct {
	Distance           int
	ATerrain, DTerrain int
	Pincer, Blocking   int
	ReturnFire         bool
	TerrainTerms       [9][2]int // (attack, defence) signed bytes of 3ac5:0008
}

func clampSize(rec []byte, s Side) int64 {
	n := s.Size
	if word(rec, 0x0F)&4 != 0 {
		n = s.Alt
	}
	switch {
	case n == 0:
		return 1
	case n > 6:
		return 6
	}
	return int64(n)
}

// Resolve transcribes casualty routine 2466:000e. Operation order, signed
// truncation and fixed-point adjustments follow the disassembly.
func Resolve(a, d Side, p Params) (res Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok && errors.Is(e, ErrDivideError) {
				err = e
				return
			}
			panic(r)
		}
	}()
	ra, rd := a.Rec, d.Rec
	suppressed := word(rd, 6)&0x80 != 0 || !p.ReturnFire
	if word(ra, 0x0D)&word(rd, 6)&0x3C == 0 {
		suppressed = true
	}
	if p.ReturnFire && word(rd, 6)&0x40 != 0 {
		suppressed = true
	}

	szA, szD := clampSize(ra, a), clampSize(rd, d)
	pw, rng := powerFor(ra, d.Flg)
	eD, armD := int64(d.Exp+1), int64(rd[2])
	if rng > 1 {
		pw = wrap(pw - div(mul(3, mul(pw, int64(p.Distance-1))), 4*rng))
	}
	sw, _ := powerFor(rd, a.Flg)
	eA, armA := int64(a.Exp+1), int64(ra[2])

	pw = adjust(pw, int64(p.Pincer))
	armD = adjust(armD, int64(p.Blocking))

	tt := p.TerrainTerms
	atkA := terms(pw, szA, eA, int64(tt[p.ATerrain][0]), 256)
	defA := terms(armA, szA, eA, int64(tt[p.ATerrain][1]), 256)
	atkD := terms(sw, szD, eD, int64(tt[p.DTerrain][0]), 256)
	defD := terms(armD, szD, eD, int64(tt[p.DTerrain][1]), 512)

	m1, m2 := wrap(atkA-defD), wrap(atkD-defA)
	pa, pd := eA, eD
	switch {
	case eA < 1:
		pa = 2
	case eA > 7:
		pa = 6
	}
	switch {
	case eD <= 1:
		pd = 2
	case eD >= 7:
		pd = 6
	}
	lossD := div(m1, div(mul(armD, 8-pa), 2))
	lossA := div(m2, div(mul(armA, 8-pd), 2))

	if lossD <= 0 && wrap(div(defD, szD)+div(armD, 2)) < atkA {
		lossD = 1
	}
	if lossA <= 0 && wrap(div(defA, szA)+div(armA, 2)) < atkD {
		lossA = 1
	}
	if lossD > 6 && lossD < 12 {
		lossD = 6
	}
	if lossA > 6 && lossA < 12 {
		lossA = 6
	}
	lossD = min(lossD, 2*szA)
	lossA = min(lossA, 2*szD)
	lossD = min(max(lossD, 0), szD)
	lossA = min(max(lossA, 0), szA)

	res = Result{a.Size, a.Alt, d.Size, d.Alt}
	if !suppressed {
		if word(ra, 0x0F)&4 != 0 {
			res.AttackerAlt = int(szA-lossA) & 0xFF
			if res.AttackerAlt == 0 {
				res.AttackerSize = 0
			}
		} else {
			res.AttackerSize = int(szA-lossA) & 0xFF
		}
	}
	if word(rd, 0x0F)&4 != 0 {
		res.DefenderAlt = int(szD-lossD) & 0xFF
		if res.DefenderAlt == 0 {
			res.DefenderSize = 0
		}
	} else {
		res.DefenderSize = int(szD-lossD) & 0xFF
	}
	return res, nil
}
