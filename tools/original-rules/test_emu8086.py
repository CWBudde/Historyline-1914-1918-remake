"""Tests for the 8086 interpreter with small hand-assembled programs (no original needed)."""
import unittest

from emu8086 import Machine, Unsupported, to_signed


def run(code, setup=None, hooks=None):
    m = Machine()
    m.s.update(cs=0x1000, ds=0x2000, ss=0x3000, es=0x2000)
    m.r['sp'] = 0x1000
    m.mem[0x10000:0x10000 + len(code)] = bytes(code)
    if setup:
        setup(m)
    m.hooks.update(hooks or {})
    m.s['cs'], m.ip = 0x1000, 0
    m.run(until=(0x1000, len(code)))
    return m


class EmuTest(unittest.TestCase):
    def test_mov_add_flags(self):
        # mov ax,0xFFFF; add ax,1
        m = run([0xB8, 0xFF, 0xFF, 0x05, 0x01, 0x00])
        self.assertEqual((m.r['ax'], m.f['c'], m.f['z']), (0, 1, 1))

    def test_signed_overflow(self):
        # mov ax,0x7FFF; add ax,1
        m = run([0xB8, 0xFF, 0x7F, 0x05, 0x01, 0x00])
        self.assertEqual((m.r['ax'], m.f['o'], m.f['s']), (0x8000, 1, 1))

    def test_sub_borrow_and_jcc(self):
        # mov ax,3; cmp ax,5; jl +3; mov bx,1; (skip) mov cx,7
        m = run([0xB8, 3, 0, 0x3D, 5, 0, 0x7C, 3, 0xBB, 1, 0, 0xB9, 7, 0])
        self.assertEqual((m.r['bx'], m.r['cx']), (0, 7))

    def test_adc_sbb_32bit(self):
        # dx:ax = 0x0001FFFF; add ax,1; adc dx,0 -> 0x00020000
        m = run([0xB8, 0xFF, 0xFF, 0xBA, 1, 0, 0x05, 1, 0, 0x83, 0xD2, 0])
        self.assertEqual((m.r['dx'], m.r['ax']), (2, 0))

    def test_cwd_and_idiv_truncates_toward_zero(self):
        # mov ax,-7; cwd; mov bx,2; idiv bx -> -3 rem -1
        m = run([0xB8, 0xF9, 0xFF, 0x99, 0xBB, 2, 0, 0xF7, 0xFB])
        self.assertEqual((to_signed(m.r['ax']), to_signed(m.r['dx'])), (-3, -1))

    def test_unsigned_div_and_mul(self):
        # mov ax,1000; mov bx,7; xor dx,dx; div bx; then mul bx
        m = run([0xB8, 0xE8, 3, 0xBB, 7, 0, 0x33, 0xD2, 0xF7, 0xF3])
        self.assertEqual((m.r['ax'], m.r['dx']), (142, 6))
        m = run([0xB8, 0xFF, 0xFF, 0xBB, 0xFF, 0xFF, 0xF7, 0xE3])
        self.assertEqual((m.r['dx'], m.r['ax']), (0xFFFE, 1))

    def test_shifts(self):
        # mov ax,0x8001; mov cl,2; shl ax,cl ; mov bx,0x8000; sar bx,1
        m = run([0xB8, 1, 0x80, 0xB1, 2, 0xD3, 0xE0, 0xBB, 0, 0x80, 0xD1, 0xFB])
        self.assertEqual((m.r['ax'], m.r['bx']), (4, 0xC000))

    def test_memory_modrm_and_segment_override(self):
        def setup(m):
            m.ww(0x2000, 0x0010, 0x1234)
            m.ww(0x4000, 0x0010, 0xBEEF)
            m.s['es'] = 0x4000
        # mov bx,0x10; mov ax,[bx]; mov dx,es:[bx]
        m = run([0xBB, 0x10, 0, 0x8B, 0x07, 0x26, 0x8B, 0x17], setup)
        self.assertEqual((m.r['ax'], m.r['dx']), (0x1234, 0xBEEF))

    def test_bp_is_stack_relative_and_les(self):
        def setup(m):
            m.ww(0x3000, 0x0F00, 0x0020)   # offset
            m.ww(0x3000, 0x0F02, 0x5000)   # segment
            m.r['bp'] = 0x0F00
        # les bx,[bp+0]  (c4 5e 00)
        m = run([0xC4, 0x5E, 0x00], setup)
        self.assertEqual((m.r['bx'], m.s['es']), (0x20, 0x5000))

    def test_byte_registers_and_xchg(self):
        # mov al,0x12; mov ah,0x34; xchg ax,cx
        m = run([0xB0, 0x12, 0xB4, 0x34, 0x91])
        self.assertEqual((m.r['cx'], m.r['ax']), (0x3412, 0))

    def test_far_call_hook_and_retf(self):
        seen = []

        def hook(m):
            seen.append(m.rw(m.s['ss'], m.r['sp'] + 4))   # first stacked argument
            m.r['ax'] = 99
        # push 5; call far 0x7000:0x0010 (hooked); add sp,2
        code = [0xB8, 5, 0, 0x50, 0x9A, 0x10, 0, 0, 0x70, 0x83, 0xC4, 2]
        m = run(code, hooks={Machine.lin(0x7000, 0x10): hook})
        self.assertEqual((seen, m.r['ax']), ([5], 99))

    def test_near_call_ret(self):
        # 0: call 5; 3: jmp 9 (past the routine); 5: mov ax,3; 8: ret; 9: nop
        code = [0xE8, 0x02, 0x00, 0xEB, 0x04, 0xB8, 3, 0, 0xC3, 0x90]
        m = run(code)
        self.assertEqual(m.r['ax'], 3)

    def test_unsupported_opcode_is_loud(self):
        with self.assertRaises(Unsupported):
            run([0xCD, 0x21])

    def test_relocation(self):
        m = Machine()
        m.load(bytes([0x34, 0x12, 0, 0]), [0], 0x100)
        self.assertEqual(m.rw(0x100, 0), 0x1234 + 0x100)


if __name__ == '__main__':
    unittest.main()
