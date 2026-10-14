"""Run the DOS casualty routine 2466:000e of HL.EXE in the 8086 interpreter.

Needs the local load image exported by disassemble.py (HL.image.bin, relocations.csv,
UNIT.DAT.bin); the original bytes stay under the ignored build/ directory. The routine
reads a context record through the far pointer at 3ac4:0010. Offsets used (all traced in
windows/combat.asm; names are descriptive, the record has other, unread fields):

    +0x00 far  attacker live unit        +0x04 far  defender live unit
    +0x0a far  attacker type record      +0x0e far  defender type record
    +0x12 far  attacker terrain record   +0x16 far  defender terrain record
    +0x1a/0x1c positions (to the distance helper, hooked here)
    +0x2e      return fire allowed (zero suppresses the attacker's losses)
    +0x30      pincer percentage         +0x31      blocking percentage
    +0x3a dword seed input               +0x6a far  RNG seed callback (hooked)

Live unit: +1 experience, +2 size, +3 alternate size, +4 word flags (0x10 air, 0x08 water
target; otherwise ground). The distance helper 0e47:2414 is replaced by a fixed value.
"""
import csv
import json
import os

from emu8086 import Machine

HL_SHA256 = 'fe4fa44699ef36df0f1152786557e790651f44b836f9f6bffd3ca5dcc38a9b92'
BASE = 0x1000                      # load segment of the image
ENTRY = (0x2466, 0x000E)
SCRATCH = 0x6000                   # segments from here up hold the synthetic structures
CTX_PTR = (0x3AC4, 0x0010)         # global far pointer to the context (unrelocated)
STACK_LIMIT = (0x3CD8, 0x11FA)
MAP_DIVISOR = (0x32F7, 0x4CA8)
RET_SEG = 0x5FF0                   # return address planted for the routine's retf


class DosCombat:
    def __init__(self, directory):
        with open(os.path.join(directory, 'manifest.json'), encoding='utf-8') as f:
            if json.load(f)['hl_sha256'] != HL_SHA256:
                raise ValueError('export is not from the identified HL.EXE build')
        with open(os.path.join(directory, 'HL.image.bin'), 'rb') as f:
            self.image = f.read()
        with open(os.path.join(directory, 'relocations.csv'), newline='') as f:
            self.relocs = [int(r['image_offset']) for r in csv.DictReader(f)]
        with open(os.path.join(directory, 'UNIT.DAT.bin'), 'rb') as f:
            self.units = f.read()

    def type_record(self, index):
        return bytearray(self.units[index * 69:(index + 1) * 69])

    def terrain_terms(self):
        """(attack, defence) signed terms of the nine terrain classes, from 3ac5:0008."""
        base = (0x3AC5 << 4) + 8
        sb = lambda v: v - 256 if v > 127 else v
        return [(sb(self.image[base + 2 * i]), sb(self.image[base + 2 * i + 1])) for i in range(9)]

    def battle(self, atk, dfn, distance=1, atk_terrain=7, dfn_terrain=7, pincer=0, blocking=0,
               return_fire=True):
        """atk/dfn: dict(type=index|record bytes, size, alt, exp, flags). Returns the result."""
        m = Machine()
        m.load(self.image, self.relocs, BASE)
        calls = []
        m.hooks[Machine.lin(BASE + 0x0E47, 0x2414)] = lambda mm: (
            calls.append(1), mm.r.__setitem__('ax', distance))
        m.hooks[Machine.lin(BASE + 0x0000, 0x1FC3)] = lambda mm: (_ for _ in ()).throw(
            RuntimeError('stack overflow check fired'))
        m.hooks[Machine.lin(SCRATCH + 0x100, 0)] = lambda mm: None   # seed callback

        seg = [SCRATCH + 0x200]

        def alloc(data):
            s = seg[0]
            m.mem[s * 16:s * 16 + len(data)] = bytes(data)
            seg[0] += (len(data) + 15) // 16 + 1
            return s

        def side(spec, terrain):
            rec = spec['type'] if not isinstance(spec['type'], int) else self.type_record(spec['type'])
            live = bytearray(26)
            live[1], live[2], live[3] = spec['exp'], spec['size'], spec.get('alt', 0)
            live[4] = spec.get('flags', 0) & 0xFF
            live[5] = spec.get('flags', 0) >> 8
            t = bytearray(6)
            t[5] = terrain
            return alloc(live), alloc(rec), alloc(t)

        a_live, a_type, a_terr = side(atk, atk_terrain)
        d_live, d_type, d_terr = side(dfn, dfn_terrain)
        ctx = bytearray(0x80)

        def far(off, s):
            ctx[off:off + 4] = bytes([0, 0, s & 0xFF, s >> 8])

        far(0x00, a_live)
        far(0x04, d_live)
        far(0x0A, a_type)
        far(0x0E, d_type)
        far(0x12, a_terr)
        far(0x16, d_terr)
        ctx[0x1A:0x1E] = bytes([0, 0, 0, 0])
        ctx[0x2E] = 1 if return_fire else 0
        ctx[0x30], ctx[0x31] = pincer, blocking
        ctx[0x6A:0x6E] = bytes([0, 0, (SCRATCH + 0x100) & 0xFF, (SCRATCH + 0x100) >> 8])
        c = alloc(ctx)
        m.ww(BASE + CTX_PTR[0], CTX_PTR[1], 0)
        m.ww(BASE + CTX_PTR[0], CTX_PTR[1] + 2, c)
        m.ww(BASE + STACK_LIMIT[0], STACK_LIMIT[1], 0xFFFF)
        m.ww(BASE + MAP_DIVISOR[0], MAP_DIVISOR[1], 1)

        m.s.update(cs=BASE + ENTRY[0], ds=BASE + 0x3AC5, es=0, ss=SCRATCH - 0x400)
        m.r['sp'] = 0x3F00
        m.push(RET_SEG)
        m.push(0)
        m.ip = ENTRY[1]
        m.run(until=(RET_SEG, 0))
        return {'attackerSize': m.rb(a_live, 2), 'attackerAlt': m.rb(a_live, 3),
                'defenderSize': m.rb(d_live, 2), 'defenderAlt': m.rb(d_live, 3),
                'distanceCalls': len(calls), 'steps': m.steps}
