"""Tests for the combat transcription (synthetic) and, when the local export exists, for the
interpreter against it. Without build/original-rules/dos-full the second part is skipped."""
import os
import unittest

import dos_combat_model as M

EXPORT = os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', '..', 'build', 'original-rules', 'dos-full')
FLAT = [(0, 0)] * 9


def record(armour=20, power=30, air_range=1, ground_range=1, mask=0x04, flags=0):
    r = bytearray(69)
    r[2], r[8], r[9], r[10], r[11], r[12] = armour, air_range, ground_range, 0, power, 0
    r[6] = mask
    r[0x0D] = 0x04
    r[0x0F] = flags
    return bytes(r)


def side(rec, size=6, exp=0):
    return {'rec': rec, 'size': size, 'alt': 0, 'exp': exp, 'flags': 0}


class ModelTest(unittest.TestCase):
    def test_division_truncates_toward_zero(self):
        self.assertEqual((M.div(-7, 2), M.div(7, -2), M.div(7, 2)), (-3, -3, 3))
        with self.assertRaises(ZeroDivisionError):
            M.div(1, 0)

    def test_adjust_is_identity_without_bonus_and_scales_with_it(self):
        self.assertEqual(M.adjust(20, 0), 20)
        self.assertEqual(M.adjust(100, 50), 150)
        self.assertEqual(M.adjust(20, 150), 49)   # fixed-point truncation: 13107 * 150 falls just short of 30.0

    def test_equal_infantry(self):
        # power 30 vs armour 20, six men, experience 0: 80 margin / 70 denominator = 1 loss.
        out = M.resolve(side(record()), side(record()), FLAT, 1, 0, 0, 0, 0, True)
        self.assertEqual((out['attackerSize'], out['defenderSize']), (5, 5))

    def test_no_return_fire_leaves_attacker_untouched(self):
        out = M.resolve(side(record()), side(record()), FLAT, 1, 0, 0, 0, 0, False)
        self.assertEqual((out['attackerSize'], out['defenderSize']), (6, 5))

    def test_losses_never_exceed_group(self):
        out = M.resolve(side(record(power=250)), side(record(armour=1), size=3), FLAT, 1, 0, 0, 0, 0, False)
        self.assertEqual(out['defenderSize'], 0)

    def test_blocking_and_pincer_change_the_result(self):
        base = M.resolve(side(record(power=45)), side(record()), FLAT, 1, 0, 0, 0, 0, False)['defenderSize']
        pincer = M.resolve(side(record(power=45)), side(record()), FLAT, 1, 0, 0, 100, 0, False)['defenderSize']
        block = M.resolve(side(record(power=45)), side(record()), FLAT, 1, 0, 0, 0, 150, False)['defenderSize']
        self.assertLess(pincer, base)
        self.assertGreater(block, base)


@unittest.skipUnless(os.path.exists(os.path.join(EXPORT, 'HL.image.bin')), 'local DOS export not present')
class InterpreterAgreesTest(unittest.TestCase):
    def test_interpreter_matches_transcription(self):
        import compare_combat
        import sys
        old = sys.argv
        sys.argv = ['compare_combat', '--cases', '120', '--dir', EXPORT]
        try:
            self.assertEqual(compare_combat.main(), 0)
        finally:
            sys.argv = old


if __name__ == '__main__':
    unittest.main()
