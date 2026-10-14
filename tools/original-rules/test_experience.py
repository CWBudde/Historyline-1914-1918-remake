import os
import random
import unittest

from dos_combat import DosCombat
from dos_experience import award

DIR = os.path.join(os.path.dirname(__file__), '..', '..', 'build', 'original-rules', 'dos-full')


def side(exp=0, before=10, after=10, flags6=0):
    return dict(exp=exp, before=before, after=after, flags6=flags6)


class AwardTest(unittest.TestCase):
    def test_no_losses_no_experience(self):
        self.assertEqual(award(side(), side()), (0, 0, []))

    def test_damage_awards_one_to_the_other_side(self):
        self.assertEqual(award(side(), side(after=7)), (1, 0, []))
        self.assertEqual(award(side(after=7), side()), (0, 1, []))

    def test_kill_awards_two(self):
        self.assertEqual(award(side(), side(after=0)), (2, 0, [1]))

    def test_victim_flag_withholds_the_damage_award_but_not_the_kill_award(self):
        self.assertEqual(award(side(), side(after=7, flags6=0x20)), (0, 0, []))
        self.assertEqual(award(side(), side(after=0, flags6=0x20)), (1, 0, [1]))

    def test_cap_is_six(self):
        self.assertEqual(award(side(exp=5), side(after=0)), (6, 0, [1]))


@unittest.skipUnless(os.path.exists(os.path.join(DIR, 'manifest.json')), 'local export absent')
class WindowAgreesTest(unittest.TestCase):
    def test_interpreter_matches_transcription(self):
        from dos_experience import run_window
        dc = DosCombat(DIR)
        rng = random.Random(7)
        for _ in range(150):
            def pick():
                before = rng.randint(1, 15)
                return side(rng.randint(0, 6), before, rng.choice([0, before, rng.randint(0, before)]),
                            rng.choice([0, 0x20, 0x01, 0x21]))
            a, d = pick(), pick()
            alt = (rng.random() < .3, rng.random() < .3)
            got = run_window(dc, a, d, alt)
            want = award(a, d)
            self.assertEqual(got, (want[0], want[1], len(want[2])), (a, d, alt))


if __name__ == '__main__':
    unittest.main()
