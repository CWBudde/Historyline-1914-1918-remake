"""A minimal 16-bit 8086 interpreter for running routines of the DOS HL.EXE.

It implements the integer subset that Borland C++ 16-bit code uses (ALU, shifts, mul/div,
jumps, calls, segment loads, stack), a flat 1 MiB memory and Python hooks for far
addresses that are not worth interpreting (distance helper, stack check, RNG callback).
There are no interrupts, string instructions, I/O or floating point: such an opcode
raises ``Unsupported`` instead of guessing.
"""


class Unsupported(Exception):
    pass


MASK = (0xFFFF, 0xFF)
REG16 = ('ax', 'cx', 'dx', 'bx', 'sp', 'bp', 'si', 'di')
SEG = ('es', 'cs', 'ss', 'ds')


def parity(v):
    return bin(v & 0xFF).count('1') % 2 == 0


class Machine:
    def __init__(self):
        self.mem = bytearray(1 << 20)
        self.r = dict.fromkeys(REG16, 0)
        self.s = dict.fromkeys(SEG, 0)
        self.ip = 0
        self.f = dict(c=0, z=0, s=0, o=0, p=0, a=0)
        self.hooks = {}   # linear address of a far entry -> fn(machine)
        self.steps = 0

    # -- memory -------------------------------------------------------------
    @staticmethod
    def lin(seg, off):
        return ((seg << 4) + (off & 0xFFFF)) & 0xFFFFF

    def rb(self, seg, off):
        return self.mem[self.lin(seg, off)]

    def rw(self, seg, off):
        return self.rb(seg, off) | self.rb(seg, off + 1) << 8

    def wb(self, seg, off, v):
        self.mem[self.lin(seg, off)] = v & 0xFF

    def ww(self, seg, off, v):
        self.wb(seg, off, v)
        self.wb(seg, off + 1, v >> 8)

    def push(self, v):
        self.r['sp'] = (self.r['sp'] - 2) & 0xFFFF
        self.ww(self.s['ss'], self.r['sp'], v)

    def pop(self):
        v = self.rw(self.s['ss'], self.r['sp'])
        self.r['sp'] = (self.r['sp'] + 2) & 0xFFFF
        return v

    # -- loading ------------------------------------------------------------
    def load(self, image, relocations, base_seg):
        """Copy a header-free MZ load image to base_seg:0; add base_seg at each relocation."""
        start = base_seg << 4
        self.mem[start:start + len(image)] = image
        for off in relocations:
            a = start + off
            v = (self.mem[a] | self.mem[a + 1] << 8) + base_seg
            self.mem[a], self.mem[a + 1] = v & 0xFF, v >> 8 & 0xFF

    # -- fetch --------------------------------------------------------------
    def fb(self):
        v = self.rb(self.s['cs'], self.ip)
        self.ip = (self.ip + 1) & 0xFFFF
        return v

    def fw(self):
        return self.fb() | self.fb() << 8

    # -- registers ----------------------------------------------------------
    def get(self, i, w):
        if w:
            return self.r[REG16[i]]
        reg = self.r[REG16[i & 3]]
        return reg >> 8 & 0xFF if i & 4 else reg & 0xFF

    def put(self, i, w, v):
        if w:
            self.r[REG16[i]] = v & 0xFFFF
            return
        n = REG16[i & 3]
        if i & 4:
            self.r[n] = self.r[n] & 0xFF | (v & 0xFF) << 8
        else:
            self.r[n] = self.r[n] & 0xFF00 | v & 0xFF

    # -- modrm --------------------------------------------------------------
    def modrm(self, seg_override):
        """Decode modrm: returns (mod, reg, rm, ea) where ea=(seg, off) for memory operands."""
        b = self.fb()
        mod, reg, rm = b >> 6, b >> 3 & 7, b & 7
        if mod == 3:
            return mod, reg, rm, None
        r = self.r
        base = (
            (r['bx'] + r['si'], 'ds'), (r['bx'] + r['di'], 'ds'),
            (r['bp'] + r['si'], 'ss'), (r['bp'] + r['di'], 'ss'),
            (r['si'], 'ds'), (r['di'], 'ds'), (r['bp'], 'ss'), (r['bx'], 'ds'))[rm]
        off, seg = base
        if mod == 0 and rm == 6:
            off, seg = 0, 'ds'
            off = self.fw()
        elif mod == 1:
            d = self.fb()
            off += d - 256 if d > 127 else d
        elif mod == 2:
            off += self.fw()
        return mod, reg, rm, (self.s[seg_override or seg], off & 0xFFFF)

    def rd(self, rm, ea, w):
        if ea is None:
            return self.get(rm, w)
        return self.rw(*ea) if w else self.rb(*ea)

    def wr(self, rm, ea, w, v):
        if ea is None:
            self.put(rm, w, v)
        elif w:
            self.ww(*ea, v)
        else:
            self.wb(*ea, v)

    # -- flags --------------------------------------------------------------
    def szp(self, v, w):
        m = MASK[not w]
        top = 0x8000 if w else 0x80
        self.f['z'] = int(v & m == 0)
        self.f['s'] = int(bool(v & top))
        self.f['p'] = int(parity(v))

    def alu(self, op, a, b, w):
        """op: 0 add 1 or 2 adc 3 sbb 4 and 5 sub 6 xor 7 cmp. Returns result (cmp: a)."""
        m, top = (0xFFFF, 0x8000) if w else (0xFF, 0x80)
        c = self.f['c']
        if op in (0, 2):
            cin = c if op == 2 else 0
            full = a + b + cin
            res = full & m
            self.f['c'] = int(full > m)
            self.f['o'] = int(bool(~(a ^ b) & (a ^ res) & top))
            self.f['a'] = int(((a & 15) + (b & 15) + cin) > 15)
        elif op in (3, 5, 7):
            cin = c if op == 3 else 0
            full = a - b - cin
            res = full & m
            self.f['c'] = int(full < 0)
            self.f['o'] = int(bool((a ^ b) & (a ^ res) & top))
            self.f['a'] = int(((a & 15) - (b & 15) - cin) < 0)
        else:
            res = {1: a | b, 4: a & b, 6: a ^ b}[op] & m
            self.f['c'] = self.f['o'] = self.f['a'] = 0
        self.szp(res, w)
        return a if op == 7 else res

    def cond(self, n):
        f = self.f
        t = (f['o'], f['c'], f['z'], f['c'] or f['z'], f['s'], f['p'],
             f['s'] != f['o'], f['z'] or f['s'] != f['o'])[n >> 1]
        return bool(t) != bool(n & 1)

    # -- calls and returns --------------------------------------------------
    def far_call(self, seg, off):
        self.push(self.s['cs'])
        self.push(self.ip)
        self.s['cs'], self.ip = seg, off & 0xFFFF
        h = self.hooks.get(self.lin(seg, off))
        if h:
            h(self)
            self.ip = self.pop()
            self.s['cs'] = self.pop()

    def run(self, until, max_steps=2_000_000):
        """Run until CS:IP equals the (seg, off) pair `until` or steps run out."""
        while (self.s['cs'], self.ip) != until:
            self.steps += 1
            if self.steps > max_steps:
                raise RuntimeError('step limit')
            self.step()

    # -- instruction --------------------------------------------------------
    def step(self):
        start = self.ip
        seg_o = None
        op = self.fb()
        while op in (0x26, 0x2E, 0x36, 0x3E):
            seg_o = {0x26: 'es', 0x2E: 'cs', 0x36: 'ss', 0x3E: 'ds'}[op]
            op = self.fb()
        r, s = self.r, self.s

        if op < 0x40 and op & 7 < 6:                      # ALU
            kind, w = op >> 3, op & 1
            if op & 7 < 4:
                _, reg, rm, ea = self.modrm(seg_o)
                a, b = (self.rd(rm, ea, w), self.get(reg, w)) if not op & 2 else \
                    (self.get(reg, w), self.rd(rm, ea, w))
                res = self.alu(kind, a, b, w)
                if kind != 7:
                    if op & 2:
                        self.put(reg, w, res)
                    else:
                        self.wr(rm, ea, w, res)
            else:
                imm = self.fw() if w else self.fb()
                res = self.alu(kind, self.get(0, w), imm, w)
                if kind != 7:
                    self.put(0, w, res)
        elif op in (0x06, 0x0E, 0x16, 0x1E):
            self.push(s[SEG[op >> 3]])
        elif op in (0x07, 0x17, 0x1F):
            s[SEG[op >> 3]] = self.pop()
        elif 0x40 <= op < 0x50:                          # inc/dec r16
            i, c = op & 7, self.f['c']
            res = self.alu(0 if op < 0x48 else 5, r[REG16[i]], 1, 1)
            r[REG16[i]] = res
            self.f['c'] = c
        elif 0x50 <= op < 0x58:
            self.push(r[REG16[op & 7]])
        elif 0x58 <= op < 0x60:
            r[REG16[op & 7]] = self.pop()
        elif 0x70 <= op < 0x80:
            d = self.fb()
            if self.cond(op & 15):
                self.ip = (self.ip + (d - 256 if d > 127 else d)) & 0xFFFF
        elif op in (0x80, 0x81, 0x83):
            w = op & 1
            _, kind, rm, ea = self.modrm(seg_o)
            if op == 0x81:
                imm = self.fw()
            else:
                imm = self.fb()
                if op == 0x83 and imm > 127:
                    imm |= 0xFF00
            res = self.alu(kind, self.rd(rm, ea, w), imm & MASK[not w], w)
            if kind != 7:
                self.wr(rm, ea, w, res)
        elif op in (0x84, 0x85):
            w = op & 1
            _, reg, rm, ea = self.modrm(seg_o)
            self.alu(4, self.rd(rm, ea, w), self.get(reg, w), w)
        elif op in (0x86, 0x87):
            w = op & 1
            _, reg, rm, ea = self.modrm(seg_o)
            a, b = self.rd(rm, ea, w), self.get(reg, w)
            self.wr(rm, ea, w, b)
            self.put(reg, w, a)
        elif op in (0x88, 0x89):
            w = op & 1
            _, reg, rm, ea = self.modrm(seg_o)
            self.wr(rm, ea, w, self.get(reg, w))
        elif op in (0x8A, 0x8B):
            w = op & 1
            _, reg, rm, ea = self.modrm(seg_o)
            self.put(reg, w, self.rd(rm, ea, w))
        elif op == 0x8C:
            _, reg, rm, ea = self.modrm(seg_o)
            self.wr(rm, ea, 1, s[SEG[reg & 3]])
        elif op == 0x8E:
            _, reg, rm, ea = self.modrm(seg_o)
            s[SEG[reg & 3]] = self.rd(rm, ea, 1)
        elif op == 0x8D:
            _, reg, rm, ea = self.modrm(seg_o)
            r[REG16[reg]] = ea[1]
        elif op == 0x90:
            pass
        elif 0x91 <= op < 0x98:
            i = op & 7
            r['ax'], r[REG16[i]] = r[REG16[i]], r['ax']
        elif op == 0x98:
            r['ax'] = r['ax'] & 0xFF | (0xFF00 if r['ax'] & 0x80 else 0)
        elif op == 0x99:
            r['dx'] = 0xFFFF if r['ax'] & 0x8000 else 0
        elif op == 0x9A:
            off, seg = self.fw(), self.fw()
            self.far_call(seg, off)
        elif op in (0xA0, 0xA1):
            w = op & 1
            off = self.fw()
            v = self.rw(s[seg_o or 'ds'], off) if w else self.rb(s[seg_o or 'ds'], off)
            self.put(0, w, v)
        elif op in (0xA2, 0xA3):
            w = op & 1
            off = self.fw()
            (self.ww if w else self.wb)(s[seg_o or 'ds'], off, self.get(0, w))
        elif op in (0xA8, 0xA9):
            w = op & 1
            self.alu(4, self.get(0, w), self.fw() if w else self.fb(), w)
        elif 0xB0 <= op < 0xB8:
            self.put(op & 7, 0, self.fb())
        elif 0xB8 <= op < 0xC0:
            r[REG16[op & 7]] = self.fw()
        elif op == 0xC3:
            self.ip = self.pop()
        elif op == 0xC2:
            n = self.fw()
            self.ip = self.pop()
            r['sp'] = (r['sp'] + n) & 0xFFFF
        elif op in (0xCA, 0xCB):
            n = self.fw() if op == 0xCA else 0
            self.ip = self.pop()
            s['cs'] = self.pop()
            r['sp'] = (r['sp'] + n) & 0xFFFF
        elif op in (0xC4, 0xC5):
            _, reg, rm, ea = self.modrm(seg_o)
            r[REG16[reg]] = self.rw(*ea)
            s['es' if op == 0xC4 else 'ds'] = self.rw(ea[0], ea[1] + 2)
        elif op in (0xC6, 0xC7):
            w = op & 1
            _, _, rm, ea = self.modrm(seg_o)
            self.wr(rm, ea, w, self.fw() if w else self.fb())
        elif op in (0xD0, 0xD1, 0xD2, 0xD3):
            w = op & 1
            _, kind, rm, ea = self.modrm(seg_o)
            n = 1 if op < 0xD2 else r['cx'] & 0xFF
            self.wr(rm, ea, w, self.shift(kind, self.rd(rm, ea, w), n, w))
        elif op in (0xE0, 0xE1, 0xE2):                   # loopnz, loopz, loop
            d = self.fb()
            r['cx'] = (r['cx'] - 1) & 0xFFFF
            take = r['cx'] != 0 and (op == 0xE2 or bool(self.f['z']) == (op == 0xE1))
            if take:
                self.ip = (self.ip + (d - 256 if d > 127 else d)) & 0xFFFF
        elif op == 0xE3:
            d = self.fb()
            if r['cx'] == 0:
                self.ip = (self.ip + (d - 256 if d > 127 else d)) & 0xFFFF
        elif op == 0xE8:
            d = self.fw()
            self.push(self.ip)
            self.ip = (self.ip + d) & 0xFFFF
        elif op == 0xE9:
            d = self.fw()
            self.ip = (self.ip + d) & 0xFFFF
        elif op == 0xEB:
            d = self.fb()
            self.ip = (self.ip + (d - 256 if d > 127 else d)) & 0xFFFF
        elif op in (0xF6, 0xF7):
            self.group3(op & 1, seg_o)
        elif op in (0xFE, 0xFF):
            w = op & 1
            _, kind, rm, ea = self.modrm(seg_o)
            if kind < 2:
                c = self.f['c']
                self.wr(rm, ea, w, self.alu(0 if kind == 0 else 5, self.rd(rm, ea, w), 1, w))
                self.f['c'] = c
            elif kind == 3:                              # call far [m]
                self.far_call(self.rw(ea[0], ea[1] + 2), self.rw(*ea))
            elif kind == 6:
                self.push(self.rd(rm, ea, 1))
            elif kind == 2:
                self.push(self.ip)
                self.ip = self.rd(rm, ea, 1)
            elif kind == 4:
                self.ip = self.rd(rm, ea, 1)
            else:
                raise Unsupported('FF /%d at %04x' % (kind, start))
        else:
            raise Unsupported('opcode %02x at %04x:%04x' % (op, s['cs'], start))

    def shift(self, kind, v, n, w):
        bits, m, top = (16, 0xFFFF, 0x8000) if w else (8, 0xFF, 0x80)
        n &= 31
        for _ in range(n):
            if kind == 0:    # rol
                self.f['c'] = int(bool(v & top))
                v = (v << 1 | self.f['c']) & m
            elif kind == 1:  # ror
                self.f['c'] = v & 1
                v = v >> 1 | (top if self.f['c'] else 0)
            elif kind == 2:  # rcl
                c = self.f['c']
                self.f['c'] = int(bool(v & top))
                v = (v << 1 | c) & m
            elif kind == 3:  # rcr
                c = self.f['c']
                self.f['c'] = v & 1
                v = v >> 1 | (top if c else 0)
            elif kind in (4, 6):  # shl
                self.f['c'] = int(bool(v & top))
                v = v << 1 & m
            elif kind == 5:  # shr
                self.f['c'] = v & 1
                v >>= 1
            else:            # sar
                self.f['c'] = v & 1
                v = v >> 1 | v & top
        if n and kind >= 4:
            self.szp(v, w)
        return v

    def group3(self, w, seg_o):
        _, kind, rm, ea = self.modrm(seg_o)
        v = self.rd(rm, ea, w)
        m, top = (0xFFFF, 0x8000) if w else (0xFF, 0x80)
        r = self.r
        if kind in (0, 1):
            self.alu(4, v, self.fw() if w else self.fb(), w)
        elif kind == 2:
            self.wr(rm, ea, w, ~v & m)
        elif kind == 3:
            res = self.alu(5, 0, v, w)
            self.wr(rm, ea, w, res)
            self.f['c'] = int(v != 0)
        elif kind == 4:   # mul
            if w:
                p = r['ax'] * v
                r['ax'], r['dx'] = p & 0xFFFF, p >> 16
                self.f['c'] = self.f['o'] = int(r['dx'] != 0)
            else:
                p = r['ax'] & 0xFF
                p *= v
                r['ax'] = p
                self.f['c'] = self.f['o'] = int(p > 0xFF)
        elif kind == 5:   # imul
            sv = v - (m + 1) if v & top else v
            if w:
                a = r['ax'] - 0x10000 if r['ax'] & 0x8000 else r['ax']
                p = a * sv
                r['ax'], r['dx'] = p & 0xFFFF, p >> 16 & 0xFFFF
                self.f['c'] = self.f['o'] = int(not -0x8000 <= p < 0x8000)
            else:
                a = (r['ax'] & 0xFF)
                a = a - 256 if a & 0x80 else a
                p = a * sv
                r['ax'] = p & 0xFFFF
                self.f['c'] = self.f['o'] = int(not -128 <= p < 128)
        else:             # div / idiv
            if v == 0:
                raise ZeroDivisionError('divide error')
            if kind == 6:
                n = (r['dx'] << 16 | r['ax']) if w else r['ax']
                q, rem = divmod(n, v)
                lim = m
            else:
                n = (r['dx'] << 16 | r['ax']) if w else r['ax']
                n = n - (1 << (32 if w else 16)) if n >> (31 if w else 15) else n
                d = v - (m + 1) if v & top else v
                q = abs(n) // abs(d) * (1 if (n < 0) == (d < 0) else -1)
                rem = n - q * d
                if not -(top) <= q < top:
                    raise OverflowError('divide overflow')
                q, rem, lim = q & m, rem & m, m
            if q > lim:
                raise OverflowError('divide overflow')
            if w:
                r['ax'], r['dx'] = q, rem
            else:
                r['ax'] = q | rem << 8


def to_signed(v, bits=16):
    return v - (1 << bits) if v >> (bits - 1) else v
