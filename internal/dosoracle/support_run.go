package dosoracle

import "github.com/CWBudde/Historyline-1914-1918-remake/internal/dos8086"

// RunSupport executes the pincer and blocking caller (02dc:262a) on a
// synthetic map. It rates the neighbours of a battle: the six hexes around the defender
// add the pincer percentage, the two beside the defender's direction around
// the attacker add the blocking percentage; results are the bytes at
// 32f7:1854 and 32f7:1855, which the casualty routine reads as ctx
// +0x30/+0x31. The neighbour hexes come from the real 0b06:0157 and the side
// test from the real 0ca6:020a; only the stack check is hooked.
//
// Memory it reads (segment 32f7): +4ca8/+4caa map width and height in cells;
// +18a4 live units of 26 bytes with +4 flag word (side bits 1/2, target class
// bits 0x3c), +6 flag word (bit 2), +8 type index, +0xb position (byte offset
// into the map plus 1); Build.TypeFlags+69*type the type's target-flag word
// (+6 of the records at 32e2:09bb in the demo; 0x40 marks long-range units).
// The map is two bytes per cell; byte 1
// is the unit index (0xf1 and up means empty). Arguments: attacker, defender,
// map far pointer, selector (0 or 1).
func (e *Export) RunSupport(mp *Map, attacker, defender, selector int) (pincer, blocking int, err error) {
	m := e.machine()
	b := e.B
	dseg := baseSeg + b.Data
	m.WW(dseg, b.Width, mp.Width)
	m.WW(dseg, b.Width+2, len(mp.Cells)/mp.Width) // map height, read by the edge tests of 0b06:0254
	for idx, u := range mp.Units {
		base := b.Units + 26*idx
		m.WW(dseg, base+4, u.Flags4)
		m.WW(dseg, base+6, u.Flags6)
		m.WB(dseg, base+8, u.Type)
		for k := range 2 {
			m.WW(dseg, base+0x0B+2*k, 2*u.Cell+1)
		}
	}
	for t, flags := range mp.Types {
		m.WW(dseg, b.TypeFlags+recordLen*t, flags)
	}
	mapSeg := scratch + 0x200
	for i, c := range mp.Cells {
		terrain := 0
		if i == 0 {
			terrain = mp.Terrain0
		}
		m.WB(mapSeg, 2*i, terrain)
		m.WB(mapSeg, 2*i+1, c)
	}

	m.S = [4]uint16{dos8086.CS: uint16(baseSeg + b.Code), dos8086.DS: uint16(baseSeg + b.DS), dos8086.SS: scratch - 0x400}
	m.R[dos8086.SP] = 0x3F00
	for _, v := range [5]int{selector, mapSeg, 0, defender, attacker} {
		m.Push(v)
	}
	m.Push(retSeg)
	m.Push(0)
	m.IP = uint16(b.Support)
	if err := m.Run(retSeg, 0, supportMaxSteps); err != nil {
		return 0, 0, wrapRun("support", err)
	}
	return m.RB(dseg, b.Results), m.RB(dseg, b.Results+1), nil
}
