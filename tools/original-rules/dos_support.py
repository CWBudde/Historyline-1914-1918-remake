"""Run the DOS pincer and blocking calculation 02dc:262a of HL.EXE on a synthetic map.

The function rates the neighbours of a battle: the six hexes around the defender add the
pincer percentage, the two hexes beside the defender's direction around the attacker add the
blocking percentage; results are the bytes at 32f7:1854 (pincer) and 32f7:1855 (blocking),
which the casualty routine reads as ctx +0x30/+0x31. The neighbour hexes come from the real
0b06:0157 and the side test from the real 0ca6:020a; only the stack check is hooked.

Memory it reads (segment 32f7): +4ca8/+4caa map width and height in cells; +18a4 live units of 26 bytes with
+4 flag word (side bits 1/2, target class bits 0x3c), +6 flag word (bit 2), +8 type index,
+0xb position (byte offset into the map plus 1); +09bc+69*type type records (flag word at +0xf
is read; 0x40 marks long-range units). The map is two bytes per cell; byte 1 is the unit index
(0xf1 and up means empty). Arguments: attacker, defender, map far pointer, selector (0 or 1).
"""
from dos_combat import BASE, SCRATCH, RET_SEG, STACK_LIMIT, DosCombat
from emu8086 import Machine

SEG = 0x02DC
ENTRY = 0x262A
DATA = 0x32F7


def run_support(dc, width, cells, units, types, attacker, defender, selector=0, terrain0=0,
                max_steps=200000):
    """cells: list of unit indices (0xff = empty) row-major; units: dict idx -> dict(flags4, flags6,
    type, cell). types: dict type idx -> flag word at +0xf. Returns (pincer, blocking)."""
    m = Machine()
    m.load(dc.image, dc.relocs, BASE)
    d = BASE + DATA
    m.hooks[Machine.lin(BASE, 0x1FC3)] = lambda mm: (_ for _ in ()).throw(RuntimeError('stack check'))
    m.ww(BASE + STACK_LIMIT[0], STACK_LIMIT[1], 0xFFFF)
    m.ww(d, 0x4CA8, width)
    m.ww(d, 0x4CAA, len(cells) // width)   # map height, read by the edge tests of 0b06:0254
    for idx, u in units.items():
        base = 0x18A4 + 26 * idx
        m.ww(d, base + 4, u['flags4'])
        m.ww(d, base + 6, u['flags6'])
        m.wb(d, base + 8, u['type'])
        for k in (0, 1):
            m.ww(d, base + 0x0B + 2 * k, 2 * u['cell'] + 1)
    for t, flags in types.items():
        m.ww(d, 0x09BC + 69 * t + 0x0F, flags)
    mapseg = SCRATCH + 0x200
    for i, c in enumerate(cells):
        m.wb(mapseg, 2 * i, terrain0 if i == 0 else 0)
        m.wb(mapseg, 2 * i + 1, c)

    ss = SCRATCH - 0x400
    m.s.update(cs=BASE + SEG, ds=BASE + 0x32E2, es=0, ss=ss)
    m.r['sp'] = 0x3F00
    for v in (selector, mapseg, 0, defender, attacker):
        m.push(v)
    m.push(RET_SEG)
    m.push(0)
    m.ip = ENTRY
    m.run(until=(RET_SEG, 0), max_steps=max_steps)
    return m.rb(d, 0x1854), m.rb(d, 0x1855)
