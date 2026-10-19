// Package dos8086 is a minimal 16-bit 8086 interpreter for running routines of
// the DOS HL.EXE. It implements the integer subset that Borland C++ 16-bit
// code uses (ALU, shifts, mul/div, jumps, calls, segment loads, stack), a flat
// 1 MiB memory and Go hooks for far or near addresses that are not worth
// interpreting (distance helper, stack check, RNG callback). There are no
// interrupts, string instructions, I/O or floating point: such an opcode
// returns an *UnsupportedError instead of guessing.
package dos8086

import (
	"errors"
	"fmt"
	"math/bits"
)

// 16-bit register indexes in encoding order.
const (
	AX = iota
	CX
	DX
	BX
	SP
	BP
	SI
	DI
)

// Segment register indexes in encoding order.
const (
	ES = iota
	CS
	SS
	DS
)

var (
	ErrStepLimit      = errors.New("step limit")
	ErrDivideError    = errors.New("divide error")
	ErrDivideOverflow = errors.New("divide overflow")
)

// UnsupportedError reports an opcode outside the implemented subset.
type UnsupportedError struct{ Msg string }

func (e *UnsupportedError) Error() string { return "unsupported: " + e.Msg }

// Hook replaces the routine at a linear address. A hooked near call returns
// at once; a hooked far call has the far return performed by the machine.
type Hook func(m *Machine) error

// Flags are the status flags.
type Flags struct{ C, Z, S, O, P, A bool }

// Machine is the CPU and its memory.
type Machine struct {
	Mem   []byte
	R     [8]uint16
	S     [4]uint16
	IP    uint16
	F     Flags
	Hooks map[int]Hook
	Steps int

	// At is the offset of the instruction being executed (in CS).
	At uint16
	// OnWrite, when set, is called before each memory write an instruction
	// makes, with the linear address, the width (0 byte, 1 word) and the
	// value. Stack pushes, hooks and setup writes are not reported.
	OnWrite func(m *Machine, lin, w, v int)
}

// New returns a machine with 1 MiB of zeroed memory.
func New() *Machine {
	return &Machine{Mem: make([]byte, 1<<20), Hooks: map[int]Hook{}}
}

// Lin is the linear address of seg:off.
func Lin(seg, off int) int { return (seg<<4 + off&0xFFFF) & 0xFFFFF }

func (m *Machine) RB(seg, off int) int { return int(m.Mem[Lin(seg, off)]) }
func (m *Machine) RW(seg, off int) int { return m.RB(seg, off) | m.RB(seg, off+1)<<8 }
func (m *Machine) WB(seg, off, v int)  { m.Mem[Lin(seg, off)] = byte(v) }
func (m *Machine) WW(seg, off, v int) {
	m.WB(seg, off, v)
	m.WB(seg, off+1, v>>8)
}

func (m *Machine) push(v int) {
	m.R[SP] -= 2
	m.WW(int(m.S[SS]), int(m.R[SP]), v)
}

func (m *Machine) pop() int {
	v := m.RW(int(m.S[SS]), int(m.R[SP]))
	m.R[SP] += 2
	return v
}

// Push and Pop expose the stack for hooks and callers.
func (m *Machine) Push(v int) { m.push(v) }
func (m *Machine) Pop() int   { return m.pop() }

// Load copies a header-free MZ load image to baseSeg:0 and adds baseSeg at
// each relocation offset.
func (m *Machine) Load(image []byte, relocations []int, baseSeg int) {
	start := baseSeg << 4
	copy(m.Mem[start:], image)
	for _, off := range relocations {
		a := start + off
		v := (int(m.Mem[a]) | int(m.Mem[a+1])<<8) + baseSeg
		m.Mem[a], m.Mem[a+1] = byte(v), byte(v>>8)
	}
}

func (m *Machine) fb() int {
	v := m.RB(int(m.S[CS]), int(m.IP))
	m.IP++
	return v
}

func (m *Machine) fw() int {
	lo := m.fb()
	return lo | m.fb()<<8
}

func (m *Machine) get(i, w int) int {
	if w != 0 {
		return int(m.R[i])
	}
	reg := int(m.R[i&3])
	if i&4 != 0 {
		return reg >> 8 & 0xFF
	}
	return reg & 0xFF
}

func (m *Machine) put(i, w, v int) {
	if w != 0 {
		m.R[i] = uint16(v)
		return
	}
	n := i & 3
	if i&4 != 0 {
		m.R[n] = m.R[n]&0xFF | uint16(v&0xFF)<<8
	} else {
		m.R[n] = m.R[n]&0xFF00 | uint16(v&0xFF)
	}
}

// ea is an effective address; reg is true for a register operand.
type ea struct {
	reg      bool
	seg, off int
}

func signed8(d int) int {
	if d > 127 {
		return d - 256
	}
	return d
}

// modrm decodes a mod/rm byte: it returns reg, rm and the operand location.
func (m *Machine) modrm(segOverride int) (reg, rm int, e ea) {
	b := m.fb()
	mod, reg, rm := b>>6, b>>3&7, b&7
	if mod == 3 {
		return reg, rm, ea{reg: true}
	}
	r := func(i int) int { return int(m.R[i]) }
	var off int
	seg := DS
	switch rm {
	case 0:
		off = r(BX) + r(SI)
	case 1:
		off = r(BX) + r(DI)
	case 2:
		off, seg = r(BP)+r(SI), SS
	case 3:
		off, seg = r(BP)+r(DI), SS
	case 4:
		off = r(SI)
	case 5:
		off = r(DI)
	case 6:
		off, seg = r(BP), SS
	case 7:
		off = r(BX)
	}
	switch {
	case mod == 0 && rm == 6:
		off, seg = m.fw(), DS
	case mod == 1:
		off += signed8(m.fb())
	case mod == 2:
		off += m.fw()
	}
	if segOverride >= 0 {
		seg = segOverride
	}
	return reg, rm, ea{seg: int(m.S[seg]), off: off & 0xFFFF}
}

func (m *Machine) rd(rm int, e ea, w int) int {
	if e.reg {
		return m.get(rm, w)
	}
	if w != 0 {
		return m.RW(e.seg, e.off)
	}
	return m.RB(e.seg, e.off)
}

func (m *Machine) wr(rm int, e ea, w, v int) {
	switch {
	case e.reg:
		m.put(rm, w, v)
	case w != 0:
		m.watch(e.seg, e.off, w, v)
		m.WW(e.seg, e.off, v)
	default:
		m.watch(e.seg, e.off, w, v)
		m.WB(e.seg, e.off, v)
	}
}

func (m *Machine) watch(seg, off, w, v int) {
	if m.OnWrite != nil {
		mask := 0xFF
		if w != 0 {
			mask = 0xFFFF
		}
		m.OnWrite(m, Lin(seg, off), w, v&mask)
	}
}

func (m *Machine) szp(v, w int) {
	mask, top := 0xFF, 0x80
	if w != 0 {
		mask, top = 0xFFFF, 0x8000
	}
	m.F.Z = v&mask == 0
	m.F.S = v&top != 0
	m.F.P = bits.OnesCount8(uint8(v))%2 == 0
}

// alu: op 0 add 1 or 2 adc 3 sbb 4 and 5 sub 6 xor 7 cmp. Returns the result
// (cmp: a).
func (m *Machine) alu(op, a, b, w int) int {
	mask, top := 0xFF, 0x80
	if w != 0 {
		mask, top = 0xFFFF, 0x8000
	}
	cin := 0
	var res int
	switch op {
	case 0, 2:
		if op == 2 && m.F.C {
			cin = 1
		}
		full := a + b + cin
		res = full & mask
		m.F.C = full > mask
		m.F.O = ^(a^b)&(a^res)&top != 0
		m.F.A = a&15+b&15+cin > 15
	case 3, 5, 7:
		if op == 3 && m.F.C {
			cin = 1
		}
		full := a - b - cin
		res = full & mask
		m.F.C = full < 0
		m.F.O = (a^b)&(a^res)&top != 0
		m.F.A = a&15-b&15-cin < 0
	default:
		switch op {
		case 1:
			res = a | b
		case 4:
			res = a & b
		case 6:
			res = a ^ b
		}
		res &= mask
		m.F.C, m.F.O, m.F.A = false, false, false
	}
	m.szp(res, w)
	if op == 7 {
		return a
	}
	return res
}

func (m *Machine) cond(n int) bool {
	f := m.F
	var t bool
	switch n >> 1 {
	case 0:
		t = f.O
	case 1:
		t = f.C
	case 2:
		t = f.Z
	case 3:
		t = f.C || f.Z
	case 4:
		t = f.S
	case 5:
		t = f.P
	case 6:
		t = f.S != f.O
	case 7:
		t = f.Z || f.S != f.O
	}
	return t != (n&1 != 0)
}

// FarCall pushes the return address and enters seg:off, running a hook if one
// is registered there.
func (m *Machine) FarCall(seg, off int) error {
	m.push(int(m.S[CS]))
	m.push(int(m.IP))
	m.S[CS], m.IP = uint16(seg), uint16(off)
	if h := m.Hooks[Lin(seg, off)]; h != nil {
		if err := h(m); err != nil {
			return err
		}
		m.IP = uint16(m.pop())
		m.S[CS] = uint16(m.pop())
	}
	return nil
}

// Run steps until CS:IP equals until, or maxSteps (0: two million) run out.
func (m *Machine) Run(untilSeg, untilOff, maxSteps int) error {
	if maxSteps == 0 {
		maxSteps = 2_000_000
	}
	for int(m.S[CS]) != untilSeg || int(m.IP) != untilOff {
		m.Steps++
		if m.Steps > maxSteps {
			return ErrStepLimit
		}
		if err := m.Step(); err != nil {
			return err
		}
	}
	return nil
}

func (m *Machine) jump(d int) { m.IP = uint16(int(m.IP) + d) }

var segPrefix = map[int]int{0x26: ES, 0x2E: CS, 0x36: SS, 0x3E: DS}

// Step executes one instruction.
func (m *Machine) Step() error {
	start := m.IP
	m.At = start
	segO := -1
	op := m.fb()
	for op == 0x26 || op == 0x2E || op == 0x36 || op == 0x3E {
		segO = segPrefix[op]
		op = m.fb()
	}
	r := &m.R

	switch {
	case op < 0x40 && op&7 < 6: // ALU
		kind, w := op>>3, op&1
		if op&7 < 4 {
			reg, rm, e := m.modrm(segO)
			var a, b int
			if op&2 == 0 {
				a, b = m.rd(rm, e, w), m.get(reg, w)
			} else {
				a, b = m.get(reg, w), m.rd(rm, e, w)
			}
			res := m.alu(kind, a, b, w)
			if kind != 7 {
				if op&2 != 0 {
					m.put(reg, w, res)
				} else {
					m.wr(rm, e, w, res)
				}
			}
		} else {
			var imm int
			if w != 0 {
				imm = m.fw()
			} else {
				imm = m.fb()
			}
			res := m.alu(kind, m.get(0, w), imm, w)
			if kind != 7 {
				m.put(0, w, res)
			}
		}
	case op == 0x06 || op == 0x0E || op == 0x16 || op == 0x1E:
		m.push(int(m.S[op>>3]))
	case op == 0x07 || op == 0x17 || op == 0x1F:
		m.S[op>>3] = uint16(m.pop())
	case op >= 0x40 && op < 0x50: // inc/dec r16
		c := m.F.C
		k := 0
		if op >= 0x48 {
			k = 5
		}
		r[op&7] = uint16(m.alu(k, int(r[op&7]), 1, 1))
		m.F.C = c
	case op >= 0x50 && op < 0x58:
		m.push(int(r[op&7]))
	case op >= 0x58 && op < 0x60:
		r[op&7] = uint16(m.pop())
	case op >= 0x70 && op < 0x80:
		d := m.fb()
		if m.cond(op & 15) {
			m.jump(signed8(d))
		}
	case op == 0x80 || op == 0x81 || op == 0x83:
		w := op & 1
		kind, rm, e := m.modrm(segO)
		var imm int
		if op == 0x81 {
			imm = m.fw()
		} else {
			imm = m.fb()
			if op == 0x83 && imm > 127 {
				imm |= 0xFF00
			}
		}
		if w == 0 {
			imm &= 0xFF
		} else {
			imm &= 0xFFFF
		}
		res := m.alu(kind, m.rd(rm, e, w), imm, w)
		if kind != 7 {
			m.wr(rm, e, w, res)
		}
	case op == 0x84 || op == 0x85:
		w := op & 1
		reg, rm, e := m.modrm(segO)
		m.alu(4, m.rd(rm, e, w), m.get(reg, w), w)
	case op == 0x86 || op == 0x87:
		w := op & 1
		reg, rm, e := m.modrm(segO)
		a, b := m.rd(rm, e, w), m.get(reg, w)
		m.wr(rm, e, w, b)
		m.put(reg, w, a)
	case op == 0x88 || op == 0x89:
		w := op & 1
		reg, rm, e := m.modrm(segO)
		m.wr(rm, e, w, m.get(reg, w))
	case op == 0x8A || op == 0x8B:
		w := op & 1
		reg, rm, e := m.modrm(segO)
		m.put(reg, w, m.rd(rm, e, w))
	case op == 0x8C:
		reg, rm, e := m.modrm(segO)
		m.wr(rm, e, 1, int(m.S[reg&3]))
	case op == 0x8E:
		reg, rm, e := m.modrm(segO)
		m.S[reg&3] = uint16(m.rd(rm, e, 1))
	case op == 0x8D:
		reg, _, e := m.modrm(segO)
		r[reg] = uint16(e.off)
	case op == 0x90:
	case op >= 0x91 && op < 0x98:
		i := op & 7
		r[AX], r[i] = r[i], r[AX]
	case op == 0x98:
		v := r[AX] & 0xFF
		if r[AX]&0x80 != 0 {
			v |= 0xFF00
		}
		r[AX] = v
	case op == 0x99:
		r[DX] = 0
		if r[AX]&0x8000 != 0 {
			r[DX] = 0xFFFF
		}
	case op == 0x9A:
		off := m.fw()
		seg := m.fw()
		return m.FarCall(seg, off)
	case op == 0xA0 || op == 0xA1:
		w := op & 1
		off := m.fw()
		m.put(0, w, m.memAt(segO, off, w))
	case op == 0xA2 || op == 0xA3:
		w := op & 1
		off := m.fw()
		seg := m.dsOr(segO)
		m.watch(seg, off, w, m.get(0, w))
		if w != 0 {
			m.WW(seg, off, m.get(0, w))
		} else {
			m.WB(seg, off, m.get(0, w))
		}
	case op == 0xA8 || op == 0xA9:
		w := op & 1
		var imm int
		if w != 0 {
			imm = m.fw()
		} else {
			imm = m.fb()
		}
		m.alu(4, m.get(0, w), imm, w)
	case op >= 0xB0 && op < 0xB8:
		m.put(op&7, 0, m.fb())
	case op >= 0xB8 && op < 0xC0:
		r[op&7] = uint16(m.fw())
	case op == 0xC3:
		m.IP = uint16(m.pop())
	case op == 0xC2:
		n := m.fw()
		m.IP = uint16(m.pop())
		r[SP] += uint16(n)
	case op == 0xCA || op == 0xCB:
		n := 0
		if op == 0xCA {
			n = m.fw()
		}
		m.IP = uint16(m.pop())
		m.S[CS] = uint16(m.pop())
		r[SP] += uint16(n)
	case op == 0xC4 || op == 0xC5:
		reg, _, e := m.modrm(segO)
		r[reg] = uint16(m.RW(e.seg, e.off))
		seg := ES
		if op == 0xC5 {
			seg = DS
		}
		m.S[seg] = uint16(m.RW(e.seg, e.off+2))
	case op == 0xC6 || op == 0xC7:
		w := op & 1
		_, rm, e := m.modrm(segO)
		var imm int
		if w != 0 {
			imm = m.fw()
		} else {
			imm = m.fb()
		}
		m.wr(rm, e, w, imm)
	case op >= 0xD0 && op <= 0xD3:
		w := op & 1
		kind, rm, e := m.modrm(segO)
		n := 1
		if op >= 0xD2 {
			n = int(r[CX]) & 0xFF
		}
		m.wr(rm, e, w, m.shift(kind, m.rd(rm, e, w), n, w))
	case op >= 0xE0 && op <= 0xE2: // loopnz, loopz, loop
		d := m.fb()
		r[CX]--
		if r[CX] != 0 && (op == 0xE2 || m.F.Z == (op == 0xE1)) {
			m.jump(signed8(d))
		}
	case op == 0xE3:
		d := m.fb()
		if r[CX] == 0 {
			m.jump(signed8(d))
		}
	case op == 0xE8:
		d := m.fw()
		target := (int(m.IP) + d) & 0xFFFF
		if h := m.Hooks[Lin(int(m.S[CS]), target)]; h != nil {
			return h(m) // a hooked near routine returns at once
		}
		m.push(int(m.IP))
		m.IP = uint16(target)
	case op == 0xE9:
		m.jump(m.fw())
	case op == 0xEB:
		m.jump(signed8(m.fb()))
	case op == 0xF6 || op == 0xF7:
		return m.group3(op&1, segO)
	case op == 0xFE || op == 0xFF:
		w := op & 1
		kind, rm, e := m.modrm(segO)
		switch kind {
		case 0, 1:
			c := m.F.C
			k := 0
			if kind == 1 {
				k = 5
			}
			m.wr(rm, e, w, m.alu(k, m.rd(rm, e, w), 1, w))
			m.F.C = c
		case 3: // call far [m]
			return m.FarCall(m.RW(e.seg, e.off+2), m.RW(e.seg, e.off))
		case 6:
			m.push(m.rd(rm, e, 1))
		case 2:
			m.push(int(m.IP))
			m.IP = uint16(m.rd(rm, e, 1))
		case 4:
			m.IP = uint16(m.rd(rm, e, 1))
		default:
			return &UnsupportedError{fmt.Sprintf("FF /%d at %04x", kind, start)}
		}
	default:
		return &UnsupportedError{fmt.Sprintf("opcode %02x at %04x:%04x", op, m.S[CS], start)}
	}
	return nil
}

func (m *Machine) dsOr(segO int) int {
	if segO >= 0 {
		return int(m.S[segO])
	}
	return int(m.S[DS])
}

func (m *Machine) memAt(segO, off, w int) int {
	if w != 0 {
		return m.RW(m.dsOr(segO), off)
	}
	return m.RB(m.dsOr(segO), off)
}

func (m *Machine) shift(kind, v, n, w int) int {
	mask, top := 0xFF, 0x80
	if w != 0 {
		mask, top = 0xFFFF, 0x8000
	}
	n &= 31
	for range n {
		switch kind {
		case 0: // rol
			m.F.C = v&top != 0
			v = (v<<1 | b2i(m.F.C)) & mask
		case 1: // ror
			m.F.C = v&1 != 0
			v >>= 1
			if m.F.C {
				v |= top
			}
		case 2: // rcl
			c := b2i(m.F.C)
			m.F.C = v&top != 0
			v = (v<<1 | c) & mask
		case 3: // rcr
			c := m.F.C
			m.F.C = v&1 != 0
			v >>= 1
			if c {
				v |= top
			}
		case 4, 6: // shl
			m.F.C = v&top != 0
			v = v << 1 & mask
		case 5: // shr
			m.F.C = v&1 != 0
			v >>= 1
		default: // sar
			m.F.C = v&1 != 0
			v = v>>1 | v&top
		}
	}
	if n != 0 && kind >= 4 {
		m.szp(v, w)
	}
	return v
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (m *Machine) group3(w, segO int) error {
	kind, rm, e := m.modrm(segO)
	v := m.rd(rm, e, w)
	mask, top := 0xFF, 0x80
	if w != 0 {
		mask, top = 0xFFFF, 0x8000
	}
	r := &m.R
	switch kind {
	case 0, 1:
		var imm int
		if w != 0 {
			imm = m.fw()
		} else {
			imm = m.fb()
		}
		m.alu(4, v, imm, w)
	case 2:
		m.wr(rm, e, w, ^v&mask)
	case 3:
		m.wr(rm, e, w, m.alu(5, 0, v, w))
		m.F.C = v != 0
	case 4: // mul
		if w != 0 {
			p := int(r[AX]) * v
			r[AX], r[DX] = uint16(p), uint16(p>>16)
			m.F.C = r[DX] != 0
		} else {
			p := int(r[AX]&0xFF) * v
			r[AX] = uint16(p)
			m.F.C = p > 0xFF
		}
		m.F.O = m.F.C
	case 5: // imul
		sv := v
		if v&top != 0 {
			sv = v - (mask + 1)
		}
		if w != 0 {
			a := int(int16(r[AX]))
			p := a * sv
			r[AX], r[DX] = uint16(p), uint16(p>>16)
			m.F.C = p < -0x8000 || p >= 0x8000
		} else {
			a := int(int8(r[AX]))
			p := a * sv
			r[AX] = uint16(p)
			m.F.C = p < -128 || p >= 128
		}
		m.F.O = m.F.C
	default: // div / idiv
		if v == 0 {
			return ErrDivideError
		}
		var n int
		if w != 0 {
			n = int(r[DX])<<16 | int(r[AX])
		} else {
			n = int(r[AX])
		}
		var q, rem int
		if kind == 6 {
			q, rem = n/v, n%v
			if q > mask {
				return ErrDivideOverflow
			}
		} else {
			if w != 0 {
				n = int(int32(n))
			} else {
				n = int(int16(n))
			}
			d := v
			if v&top != 0 {
				d = v - (mask + 1)
			}
			q = n / d // Go truncates toward zero like idiv
			rem = n - q*d
			if q < -top || q >= top {
				return ErrDivideOverflow
			}
			q, rem = q&mask, rem&mask
		}
		if w != 0 {
			r[AX], r[DX] = uint16(q), uint16(rem)
		} else {
			r[AX] = uint16(q | rem<<8)
		}
	}
	return nil
}

// Signed16 interprets v as a signed 16-bit value.
func Signed16(v int) int { return int(int16(v)) }
