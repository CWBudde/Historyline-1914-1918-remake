package dosoracle

import (
	"errors"
	"fmt"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dos8086"
)

// RunPhaseChange executes the phase change in the tail of the order handler
// (Handler.Phase, 02dc:1a09-1cc6 in the demo) on a frame as the handler
// leaves it after the attacking side's orders: DI the attacking side,
// [bp-2] the other side, [bp-6] the far pointer to the attacking side's map.
// Both maps, the live records 0-0xf0, the container records, the per-side
// records (playerBytes), the order counts (0x55), the animation flags
// (animStart), the map size and the turn counter are set up from g. The
// container sync, the map copy and the side test run; the status message is
// hooked and recorded, and code in any other segment fails the run.
func (e *Export) RunPhaseChange(g *PhaseChange) (PhaseRun, error) {
	b, h, mv := e.B, e.B.Handler, e.B.Move
	if h == nil || mv == nil || h.Phase[1] == 0 {
		return PhaseRun{}, errors.New("the phase change is not located in this build")
	}
	size := 2 * g.Width * g.Height
	for k := range 2 {
		if len(g.Maps[k]) != size {
			return PhaseRun{}, fmt.Errorf("map %d holds %d bytes, not %d", k, len(g.Maps[k]), size)
		}
	}
	m := e.machine()
	ar := newArena(m)
	data, code := baseSeg+b.Data, baseSeg+b.Code
	att := g.Attacker & 1

	var maps [2]int
	for k := range 2 {
		mp := make([]byte, size)
		for i, v := range g.Maps[k] {
			mp[i] = byte(v)
		}
		maps[k] = ar.alloc(mp)
		m.WW(data, mv.Maps[k], 0)
		m.WW(data, mv.Maps[k]+2, maps[k])
	}
	m.WW(data, b.Width, g.Width)
	m.WW(data, b.Width+2, g.Height)
	m.WW(data, h.Size, size)
	for idx, rec := range g.live() {
		copy(m.Mem[dos8086.Lin(data, b.Units+liveLen*idx):], rec)
	}
	for idx, rec := range holdsOf(g.Holds) {
		copy(m.Mem[dos8086.Lin(data, mv.Containers+containerLen*idx):], rec)
	}
	for k := range 2 {
		copy(m.Mem[dos8086.Lin(data, h.Players+playerLen*k):], playerBytes(k, g.Phase[k]))
		m.WB(data, mv.Sides+sideLen*k+sideOrderCount, 0x55)
	}
	m.WW(data, mv.Anim, animStart)
	m.WW(data, h.Count, g.Count)

	r := PhaseRun{Live: map[int][]byte{}, Holds: map[int][]byte{}}
	m.Hooks[e.lin(mv.Message)] = func(mm *dos8086.Machine) error {
		ss, sp := int(mm.S[dos8086.SS]), int(mm.R[dos8086.SP])
		r.Messages = append(r.Messages, [2]int{mm.RW(ss, sp+4), mm.RW(ss, sp+6)})
		return nil
	}
	ss := scratch - 0x400
	m.S = [4]uint16{dos8086.CS: uint16(code), dos8086.DS: uint16(baseSeg + b.DS), dos8086.SS: uint16(ss)}
	const bp = 0x3F00
	m.R[dos8086.BP], m.R[dos8086.SP], m.R[dos8086.DI] = bp, bp-0x30, uint16(att)
	m.WW(ss, bp-2, 1-att)
	m.WW(ss, bp-6, 0)
	m.WW(ss, bp-4, maps[att])
	m.IP = uint16(h.Phase[0])
	// The window calls only the side test and the map copy (and the hooked
	// message); code in any other segment fails the run.
	allowed := map[int]bool{code: true, baseSeg + h.Copy[0]: true, baseSeg + h.Owner[0]: true}
	for m.Steps = 0; int(m.S[dos8086.CS]) != code || int(m.IP) != h.Phase[1]; m.Steps++ {
		if m.Steps > moveMaxSteps {
			return r, wrapRun("phase change", dos8086.ErrStepLimit)
		}
		if err := m.Step(); err != nil {
			return r, wrapRun("phase change", err)
		}
		if cs := int(m.S[dos8086.CS]); !allowed[cs] {
			return r, fmt.Errorf("phase change: entered segment %04x", cs-baseSeg)
		}
	}

	for k := range 2 {
		r.Maps[k] = make([]byte, size)
		copy(r.Maps[k], m.Mem[maps[k]*16:])
		r.Players[k] = make([]byte, playerLen)
		copy(r.Players[k], m.Mem[dos8086.Lin(data, h.Players+playerLen*k):])
		r.Orders[k] = m.RB(data, mv.Sides+sideLen*k+sideOrderCount)
	}
	for idx := range g.Units {
		r.Live[idx] = make([]byte, liveLen)
		copy(r.Live[idx], m.Mem[dos8086.Lin(data, b.Units+liveLen*idx):])
	}
	for i := range holdRecords {
		r.Holds[i] = make([]byte, containerLen)
		copy(r.Holds[i], m.Mem[dos8086.Lin(data, mv.Containers+containerLen*i):])
	}
	r.Count = m.RW(data, h.Count)
	r.Anim = m.RW(data, mv.Anim)
	return r, nil
}
