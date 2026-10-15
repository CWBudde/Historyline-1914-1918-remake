package dos8086

import (
	"errors"
	"slices"
	"testing"
)

// run executes hand-assembled code at 1000:0000 until it falls off the end.
func run(t *testing.T, code []byte, setup func(*Machine), hooks map[int]Hook) *Machine {
	t.Helper()
	m := New()
	m.S = [4]uint16{ES: 0x2000, CS: 0x1000, SS: 0x3000, DS: 0x2000}
	m.R[SP] = 0x1000
	copy(m.Mem[0x10000:], code)
	if setup != nil {
		setup(m)
	}
	for k, h := range hooks {
		m.Hooks[k] = h
	}
	m.S[CS], m.IP = 0x1000, 0
	if err := m.Run(0x1000, len(code), 0); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestMovAddFlags(t *testing.T) {
	m := run(t, []byte{0xB8, 0xFF, 0xFF, 0x05, 0x01, 0x00}, nil, nil)
	if m.R[AX] != 0 || !m.F.C || !m.F.Z {
		t.Fatalf("ax=%x c=%v z=%v", m.R[AX], m.F.C, m.F.Z)
	}
}

func TestSignedOverflow(t *testing.T) {
	m := run(t, []byte{0xB8, 0xFF, 0x7F, 0x05, 0x01, 0x00}, nil, nil)
	if m.R[AX] != 0x8000 || !m.F.O || !m.F.S {
		t.Fatalf("ax=%x o=%v s=%v", m.R[AX], m.F.O, m.F.S)
	}
}

func TestSubBorrowAndJcc(t *testing.T) {
	m := run(t, []byte{0xB8, 3, 0, 0x3D, 5, 0, 0x7C, 3, 0xBB, 1, 0, 0xB9, 7, 0}, nil, nil)
	if m.R[BX] != 0 || m.R[CX] != 7 {
		t.Fatalf("bx=%d cx=%d", m.R[BX], m.R[CX])
	}
}

func TestAdcCarriesIntoHighWord(t *testing.T) {
	m := run(t, []byte{0xB8, 0xFF, 0xFF, 0xBA, 1, 0, 0x05, 1, 0, 0x83, 0xD2, 0}, nil, nil)
	if m.R[DX] != 2 || m.R[AX] != 0 {
		t.Fatalf("dx=%d ax=%d", m.R[DX], m.R[AX])
	}
}

func TestIdivTruncatesTowardZero(t *testing.T) {
	m := run(t, []byte{0xB8, 0xF9, 0xFF, 0x99, 0xBB, 2, 0, 0xF7, 0xFB}, nil, nil)
	if Signed16(int(m.R[AX])) != -3 || Signed16(int(m.R[DX])) != -1 {
		t.Fatalf("ax=%d dx=%d", Signed16(int(m.R[AX])), Signed16(int(m.R[DX])))
	}
}

func TestUnsignedDivAndMul(t *testing.T) {
	m := run(t, []byte{0xB8, 0xE8, 3, 0xBB, 7, 0, 0x33, 0xD2, 0xF7, 0xF3}, nil, nil)
	if m.R[AX] != 142 || m.R[DX] != 6 {
		t.Fatalf("ax=%d dx=%d", m.R[AX], m.R[DX])
	}
	m = run(t, []byte{0xB8, 0xFF, 0xFF, 0xBB, 0xFF, 0xFF, 0xF7, 0xE3}, nil, nil)
	if m.R[DX] != 0xFFFE || m.R[AX] != 1 {
		t.Fatalf("dx=%x ax=%x", m.R[DX], m.R[AX])
	}
}

func TestDivideErrors(t *testing.T) {
	// mov ax,1; mov bx,0; div bx
	m := New()
	m.S = [4]uint16{CS: 0x1000, SS: 0x3000, DS: 0x2000}
	m.R[SP] = 0x1000
	copy(m.Mem[0x10000:], []byte{0xB8, 1, 0, 0xBB, 0, 0, 0xF7, 0xF3})
	if err := m.Run(0x1000, 8, 0); !errors.Is(err, ErrDivideError) {
		t.Fatalf("got %v", err)
	}
}

func TestShifts(t *testing.T) {
	m := run(t, []byte{0xB8, 1, 0x80, 0xB1, 2, 0xD3, 0xE0, 0xBB, 0, 0x80, 0xD1, 0xFB}, nil, nil)
	if m.R[AX] != 4 || m.R[BX] != 0xC000 {
		t.Fatalf("ax=%x bx=%x", m.R[AX], m.R[BX])
	}
}

func TestMemoryModrmAndSegmentOverride(t *testing.T) {
	setup := func(m *Machine) {
		m.WW(0x2000, 0x10, 0x1234)
		m.WW(0x4000, 0x10, 0xBEEF)
		m.S[ES] = 0x4000
	}
	m := run(t, []byte{0xBB, 0x10, 0, 0x8B, 0x07, 0x26, 0x8B, 0x17}, setup, nil)
	if m.R[AX] != 0x1234 || m.R[DX] != 0xBEEF {
		t.Fatalf("ax=%x dx=%x", m.R[AX], m.R[DX])
	}
}

func TestBPIsStackRelativeAndLes(t *testing.T) {
	setup := func(m *Machine) {
		m.WW(0x3000, 0x0F00, 0x0020)
		m.WW(0x3000, 0x0F02, 0x5000)
		m.R[BP] = 0x0F00
	}
	m := run(t, []byte{0xC4, 0x5E, 0x00}, setup, nil)
	if m.R[BX] != 0x20 || m.S[ES] != 0x5000 {
		t.Fatalf("bx=%x es=%x", m.R[BX], m.S[ES])
	}
}

func TestByteRegistersAndXchg(t *testing.T) {
	m := run(t, []byte{0xB0, 0x12, 0xB4, 0x34, 0x91}, nil, nil)
	if m.R[CX] != 0x3412 || m.R[AX] != 0 {
		t.Fatalf("cx=%x ax=%x", m.R[CX], m.R[AX])
	}
}

func TestFarCallHookAndReturn(t *testing.T) {
	var seen []int
	hook := func(m *Machine) error {
		seen = append(seen, m.RW(int(m.S[SS]), int(m.R[SP])+4)) // first stacked argument
		m.R[AX] = 99
		return nil
	}
	code := []byte{0xB8, 5, 0, 0x50, 0x9A, 0x10, 0, 0, 0x70, 0x83, 0xC4, 2}
	m := run(t, code, nil, map[int]Hook{Lin(0x7000, 0x10): hook})
	if !slices.Equal(seen, []int{5}) || m.R[AX] != 99 {
		t.Fatalf("seen=%v ax=%d", seen, m.R[AX])
	}
}

func TestNearCallRet(t *testing.T) {
	m := run(t, []byte{0xE8, 0x02, 0x00, 0xEB, 0x04, 0xB8, 3, 0, 0xC3, 0x90}, nil, nil)
	if m.R[AX] != 3 {
		t.Fatalf("ax=%d", m.R[AX])
	}
}

func TestUnsupportedOpcodeIsLoud(t *testing.T) {
	m := New()
	m.S[CS] = 0x1000
	copy(m.Mem[0x10000:], []byte{0xCD, 0x21})
	var u *UnsupportedError
	if err := m.Run(0x1000, 2, 0); !errors.As(err, &u) {
		t.Fatalf("got %v", err)
	}
}

func TestRelocation(t *testing.T) {
	m := New()
	m.Load([]byte{0x34, 0x12, 0, 0}, []int{0}, 0x100)
	if got := m.RW(0x100, 0); got != 0x1234+0x100 {
		t.Fatalf("got %x", got)
	}
}
