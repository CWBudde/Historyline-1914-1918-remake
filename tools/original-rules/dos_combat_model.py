"""Independent transcription of the DOS casualty routine 2466:000e (windows/combat.asm).

Pure integer arithmetic with 32-bit wrap-around and truncating signed division, written from
the disassembly and the findings in docs/research/original-dos-rules.md, not from the
interpreter. Comparing it with dos_combat.DosCombat checks that the interpreter runs the
routine as written; it does not show that the routine matches the running game.

Inputs are plain numbers: type records are 69-byte sequences (UNIT.DAT layout), `terrain`
maps a terrain class to its (attack term, defence term) signed bytes from 3ac5:0008.
"""


def wrap(v):
    v &= 0xFFFFFFFF
    return v - (1 << 32) if v >> 31 else v


def mul(a, b):
    return wrap(a * b)


def div(a, b):
    """Signed division truncating toward zero (0000:03dc). A zero divisor faults."""
    if b == 0:
        raise ZeroDivisionError('divide error')
    q = abs(a) // abs(b)
    return wrap(q if (a < 0) == (b < 0) else -q)


def s16(v):
    v &= 0xFFFF
    return v - 0x10000 if v & 0x8000 else v


def word(rec, off):
    return rec[off] | rec[off + 1] << 8


def adjust(value, pct):
    """(value << 16) plus trunc((value << 16) / 100) * pct, then the high word, signed."""
    shifted = wrap((value & 0xFFFF) << 16)
    return s16(wrap(shifted + mul(div(shifted, 100), pct)) >> 16)


def terms(base, size, exp1, term, divisor):
    """base*size + trunc(base*F*size/divisor) + trunc(base/2)*term, F = trunc((exp1<<8)/6)+1."""
    f = div(wrap(exp1 << 8), 6) + 1
    return wrap(mul(base, size) + div(mul(mul(base, f), size), divisor) + mul(div(base, 2), term))


def power_for(rec, target_flags):
    """(power, encoded range) for a target with the given live flag word."""
    if target_flags & 0x10:
        return rec[10], rec[8]
    if target_flags & 0x08:
        return rec[12], rec[9]
    return rec[11], rec[9]


def clamp_size(rec, unit):
    n = unit['alt'] if word(rec, 0x0F) & 4 else unit['size']
    return 1 if n == 0 else 6 if n > 6 else n


def resolve(a, d, terrain, distance, a_terrain, d_terrain, pincer, blocking, return_fire):
    """a, d: dict(rec=69 bytes, size, alt, exp, flags). Returns the live bytes written back."""
    ra, rd = a['rec'], d['rec']
    suppressed = bool(word(rd, 6) & 0x80) or not return_fire
    if not word(ra, 0x0D) & word(rd, 6) & 0x3C:
        suppressed = True
    if return_fire and word(rd, 6) & 0x40:
        suppressed = True

    sz_a, sz_d = clamp_size(ra, a), clamp_size(rd, d)
    p, rng = power_for(ra, d['flags'])
    e_d, arm_d = d['exp'] + 1, rd[2]
    if rng > 1:
        p = wrap(p - div(mul(3, mul(p, distance - 1)), 4 * rng))
    s, _ = power_for(rd, a['flags'])
    e_a, arm_a = a['exp'] + 1, ra[2]

    p = adjust(p, pincer)
    arm_d = adjust(arm_d, blocking)

    atk_a = terms(p, sz_a, e_a, terrain[a_terrain][0], 256)
    def_a = terms(arm_a, sz_a, e_a, terrain[a_terrain][1], 256)
    atk_d = terms(s, sz_d, e_d, terrain[d_terrain][0], 256)
    def_d = terms(arm_d, sz_d, e_d, terrain[d_terrain][1], 512)

    m1, m2 = wrap(atk_a - def_d), wrap(atk_d - def_a)
    pa = 2 if e_a < 1 else e_a if e_a <= 7 else 6
    pd = 2 if e_d <= 1 else e_d if e_d < 7 else 6
    loss_d = div(m1, div(mul(arm_d, 8 - pa), 2))
    loss_a = div(m2, div(mul(arm_a, 8 - pd), 2))

    if loss_d <= 0 and wrap(div(def_d, sz_d) + div(arm_d, 2)) < atk_a:
        loss_d = 1
    if loss_a <= 0 and wrap(div(def_a, sz_a) + div(arm_a, 2)) < atk_d:
        loss_a = 1
    if 6 < loss_d < 12:
        loss_d = 6
    if 6 < loss_a < 12:
        loss_a = 6
    loss_d = min(loss_d, 2 * sz_a)
    loss_a = min(loss_a, 2 * sz_d)
    loss_d = min(max(loss_d, 0), sz_d)
    loss_a = min(max(loss_a, 0), sz_a)

    out = {'attackerSize': a['size'], 'attackerAlt': a['alt'],
           'defenderSize': d['size'], 'defenderAlt': d['alt']}
    if not suppressed:
        if word(ra, 0x0F) & 4:
            out['attackerAlt'] = (sz_a - loss_a) & 0xFF
            if out['attackerAlt'] == 0:
                out['attackerSize'] = 0
        else:
            out['attackerSize'] = (sz_a - loss_a) & 0xFF
    if word(rd, 0x0F) & 4:
        out['defenderAlt'] = (sz_d - loss_d) & 0xFF
        if out['defenderAlt'] == 0:
            out['defenderSize'] = 0
    else:
        out['defenderSize'] = (sz_d - loss_d) & 0xFF
    return out
