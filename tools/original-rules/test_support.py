import os
import random
import unittest

from dos_support_model import Undefined, ring, support

DIR = os.path.join(os.path.dirname(__file__), '..', '..', 'build', 'original-rules', 'dos-full')


def scenario(rng):
    """A random map with an attacker, a defender and a few more units; adjacent most of the time."""
    width, height = rng.randint(3, 8), rng.randint(3, 8)
    cells = [0xFF] * (width * height)
    dcell = rng.randrange(width * height)
    near = [c for c in ring(dcell, width, height) if c is not None]
    acell = rng.choice(near) if near and rng.random() < .8 else rng.randrange(width * height)
    if acell == dcell:
        acell = near[0]
    spots = [c for c in range(width * height) if c not in (acell, dcell)]
    rng.shuffle(spots)
    placed = [(0, acell), (1, dcell)] + [(i + 2, c) for i, c in enumerate(spots[:rng.randint(6, 30)])]
    types = {0: rng.choice([0x04, 0x08, 0x10, 0x20, 0x0C, 0x44, 0x3C]),
             1: rng.choice([0x04, 0x08, 0x20, 0x14, 0x40]), 2: 0x3C}
    units = {}
    for idx, cell in placed:
        cells[cell] = idx
        units[idx] = dict(flags4=rng.choice([1, 2, 2, 1, 0]) | rng.choice([0x04, 0x08, 0x10, 0x20, 0x3C, 0x3C]),
                          flags6=rng.choice([0, 2]), type=rng.choice([0, 1, 2]), cell=cell)
    return width, height, cells, rng.choice([0, 1, 2, 7]), units, types


class ModelTest(unittest.TestCase):
    def test_ring_edges(self):
        self.assertEqual(ring(0, 4, 4), [None, None, 1, 4, None, None])
        self.assertEqual(sum(c is not None for c in ring(5, 4, 4)), 6)

    def test_single_supporter_and_blockers(self):
        w = h = 7
        c = 3 * w + 3
        cells = [0xFF] * (w * h)
        mk = lambda cell, side: dict(flags4=side | 4, flags6=0, type=0, cell=cell)
        # attacker left of the defender, a friend beside the defender, an enemy beside the attacker
        cells[c], cells[c + 1], cells[c + 2], cells[c - 1] = 0, 1, 2, 3
        units = {0: mk(c, 1), 1: mk(c + 1, 2), 2: mk(c + 2, 1), 3: mk(c - 1, 2)}
        self.assertEqual(support(w, h, cells, 0, units, {0: 4}, 0, 1), (50, 0))

    def test_far_apart_units_get_nothing(self):
        w = h = 7
        cells = [0xFF] * (w * h)
        cells[0], cells[48] = 0, 1
        units = {0: dict(flags4=5, flags6=0, type=0, cell=0), 1: dict(flags4=6, flags6=0, type=0, cell=48)}
        # cell 24's neighbours are all on the map; the attacker sits elsewhere
        units[1]['cell'] = 24
        cells[48], cells[24] = 0xFF, 1
        self.assertEqual(support(w, h, cells, 0, units, {0: 4}, 0, 1), (0, 0))


@unittest.skipUnless(os.path.exists(os.path.join(DIR, 'manifest.json')), 'local export absent')
class InterpreterAgreesTest(unittest.TestCase):
    def test_interpreter_matches_transcription(self):
        from dos_combat import DosCombat
        from dos_support import run_support
        dc = DosCombat(DIR)
        rng = random.Random(11)
        compared = nonzero = 0
        for _ in range(300):
            w, h, cells, t0, units, types = scenario(rng)
            try:
                want = support(w, h, cells, t0, units, types, 0, 1)
            except Undefined:
                continue
            got = run_support(dc, w, cells, units, types, 0, 1, terrain0=t0)
            self.assertEqual(got, want, (w, h, cells, t0, units, types))
            compared += 1
            nonzero += any(want)
        self.assertGreater(compared, 150)
        self.assertGreater(nonzero, 40)


if __name__ == '__main__':
    unittest.main()
