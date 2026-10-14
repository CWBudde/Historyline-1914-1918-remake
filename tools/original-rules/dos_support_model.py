"""Independent transcription of the DOS pincer and blocking calculation (02dc:262a).

Written from windows/pincer_and_blocking.asm and the neighbour routines 0b06:0157/0254 and
0ca6:020a, not from the interpreter run in dos_support.py. Geometry: cells are numbered
row-major, odd columns sit half a cell lower; a neighbour is -1 when it falls off the map.

Quirks kept on purpose (all visible in the instructions):
  * the early exit tests the six neighbour bytes read at "offset + 1"; an off-map neighbour is
    read at map byte 0 (a terrain byte) instead, and still counts when it differs from the
    attacker;
  * when the attacker is not among the defender's neighbours and some neighbour is off the map,
    the direction comes from an uninitialised stack byte: `Undefined` is raised;
  * the pincer sum is a byte, so the 300 clamp is dead code;
  * one eligible neighbour of the defender next to the attacker adds 80 blocking, two give 150.
"""
WEIGHTS = (0, 25, 50, 100, 50, 25)          # indexed by the step from the attacker's direction
LONG_RANGE = 0x40


class Undefined(Exception):
    pass


def ring(cell, width, height):
    """The six neighbours of a cell in the DOS order (0 up, 1..2 right side, 3 down, 4..5 left)."""
    x, y = cell % width, cell // width
    n = [None] * 6
    if x & 1:
        n[0], n[3] = cell - width, cell + width
        n[1], n[5] = cell + 1, cell - 1
        n[4], n[2] = cell + width - 1, cell + width + 1
        if y <= 0:
            n[0] = None
        elif height - 1 <= y:
            n[2] = n[3] = n[4] = None
        if width - 1 <= x:
            n[1] = n[2] = None
    else:
        n[0], n[3] = cell - width, cell + width
        n[2], n[4] = cell + 1, cell - 1
        n[1], n[5] = cell - width + 1, cell - width - 1
        if y <= 0:
            n[0] = n[5] = n[1] = None
        elif height - 1 <= y:
            n[3] = None
        if x <= 0:
            n[4] = n[5] = None
        elif width - 1 <= x:
            n[1] = n[2] = None
    return n


def friendly(neighbour_flags, own_flags):
    """0ca6:020a: units of side bit 1 or 2 pair with the same bit; others pair with non-side-1."""
    if neighbour_flags & 1:
        return bool(own_flags & 1)
    if neighbour_flags & 2:
        return bool(own_flags & 2)
    return not own_flags & 1


def scan(width, height, cells, terrain0, units, types, anchor, reference):
    """02dc:2829. Eligible neighbours of `anchor`'s hex and the direction of `reference`."""
    b = units[anchor]
    classes, bit2 = b['flags4'] & 0x3C, b['flags6'] & 2
    slots = [0] * 6
    direction = None
    for k, c in enumerate(ring(b['cell'], width, height)):
        if c is None:
            continue
        u = cells[c]
        if u <= 0xF0:
            t = types[units[u]['type']]
            if (classes & t or bit2 & t) and not t & LONG_RANGE and \
                    friendly(units[u]['flags4'], units[reference]['flags4']):
                slots[k] = 1
        if u == reference:
            direction = k
    return slots, direction


def support(width, height, cells, terrain0, units, types, attacker, defender):
    """Returns (pincer, blocking) as the two bytes at 32f7:1854/1855."""
    near = ring(units[defender]['cell'], width, height)
    # off-map neighbours were read at map byte 0 (cell 0's terrain), valid ones at their unit byte
    differing = sum(1 for c in near if (cells[c] if c is not None else terrain0) != attacker)
    if differing == 6:
        return 0, 0
    slots, direction = scan(width, height, cells, terrain0, units, types, defender, attacker)
    if direction is None:
        raise Undefined('attacker is not next to the defender')
    pincer = 0
    for k in range(1, 6):
        if slots[(direction + k) % 6]:
            pincer = (pincer + WEIGHTS[k]) & 0xFF
    slots, direction = scan(width, height, cells, terrain0, units, types, attacker, defender)
    if direction is None:
        raise Undefined('defender is not next to the attacker')
    blocking = 0
    for k in (1, 5):
        if slots[(direction + k) % 6]:
            blocking += 80
    return pincer, min(blocking, 150)
