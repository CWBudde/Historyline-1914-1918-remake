"""Experience awards after a DOS battle: the code at 02dc:0908-0abf of HL.EXE.

The attack-resolution function at 02dc:0004 runs the casualty routine and then, in this
window, turns the strength changes of both units into experience. `award` is a plain
transcription; `run_window` executes the window itself in the interpreter (near calls to
the destroy and refresh helpers are hooked) so the two can be compared.

Per side, with `before` the strength saved ahead of the battle (frame byte bp-0x1d / -0x1e)
and `after` the live strength now (byte 3 when the type record's flag word +0xf has 0x04,
otherwise byte 2):
  * a side whose `after` is zero is destroyed and the opponent gains 1 experience;
  * a side that lost strength (`before - after` as a byte, non-zero) gives the opponent
    1 experience, unless that side's live flag word at +6 has bit 0x20;
  * experience is clamped to 6.
So a kill awards 2 and plain damage 1; the receiving side's own losses never cost experience.
"""
from dos_combat import BASE, SCRATCH, DosCombat
from emu8086 import Machine

CAP = 6
SEG = 0x02DC
FIRST, LAST = 0x0908, 0x0ABF
NO_AWARD_FLAG = 0x20     # bit of the victim's live word at +6


def award(a, d):
    """a, d: dict(exp, before, after, flags6). Returns (attacker exp, defender exp, kills)."""
    gain = [0, 0]
    kills = []
    for me, other, i in ((a, d, 0), (d, a, 1)):
        if me['after'] == 0:
            gain[1 - i] += 1
            kills.append(i)
        if (me['before'] - me['after']) & 0xFF and not me['flags6'] & NO_AWARD_FLAG:
            gain[1 - i] += 1
    return (min(CAP, a['exp'] + gain[0]), min(CAP, d['exp'] + gain[1]), kills)


def run_window(dc, a, d, alt=(False, False)):
    """Execute the window. `alt` says whether each side's type record uses the alternate byte."""
    m = Machine()
    m.load(dc.image, dc.relocs, BASE)
    seg = [SCRATCH + 0x200]

    def alloc(data):
        s = seg[0]
        m.mem[s * 16:s * 16 + len(data)] = bytes(data)
        seg[0] += (len(data) + 15) // 16 + 1
        return s

    def side(spec, use_alt):
        live = bytearray(26)
        live[1] = spec['exp']
        live[3 if use_alt else 2] = spec['after']
        live[6] = spec['flags6']
        rec = bytearray(69)
        rec[0x0F] = 4 if use_alt else 0
        return alloc(live), alloc(rec)

    a_live, a_type = side(a, alt[0])
    d_live, d_type = side(d, alt[1])
    ctx = bytearray(0x20)
    for off, s in ((0, a_live), (4, d_live), (0x0A, a_type), (0x0E, d_type)):
        ctx[off:off + 4] = bytes([0, 0, s & 0xFF, s >> 8])
    c = alloc(ctx)

    destroyed = []
    code = BASE + SEG
    m.hooks[Machine.lin(code, 0x2083)] = lambda mm: destroyed.append(mm.rw(mm.s['ss'], mm.r['sp'] + 2))
    m.hooks[Machine.lin(code, 0x254B)] = lambda mm: None

    ss = SCRATCH - 0x400
    m.s.update(cs=code, ds=BASE + 0x32E2, es=0, ss=ss)
    bp = 0x3F00
    m.r['bp'], m.r['sp'], m.r['di'] = bp, bp - 0x60, 0
    m.ww(ss, bp - 0x0C, 0)
    m.ww(ss, bp - 0x0A, c)
    m.mem[Machine.lin(ss, bp - 0x1D)] = a['before']
    m.mem[Machine.lin(ss, bp - 0x1E)] = d['before']
    m.ip = FIRST
    m.run(until=(code, LAST))
    return (m.rb(a_live, 1), m.rb(d_live, 1), len(destroyed))
