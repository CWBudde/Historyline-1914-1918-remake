"""Compare the interpreter running 2466:000e with the independent transcription.

    python3 -I tools/original-rules/compare_combat.py [--cases N] [--dir build/original-rules/dos-full]
"""
import argparse
import os
import random
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dos_combat import DosCombat          # noqa: E402
from dos_combat_model import resolve      # noqa: E402

TERRAIN_TABLE = (0x3AC5, 0x0008)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--cases', type=int, default=300)
    ap.add_argument('--dir', default='build/original-rules/dos-full')
    ap.add_argument('--seed', type=int, default=1914)
    args = ap.parse_args()
    dos = DosCombat(args.dir)
    terrain = dos.terrain_terms()
    armed = [i for i in range(51) if dos.type_record(i)[10:13] != bytes(3)]
    rng = random.Random(args.seed)
    bad = 0
    for n in range(args.cases):
        def side():
            t = rng.choice(armed)
            return {'type': t, 'rec': dos.type_record(t), 'size': rng.randint(1, 6), 'alt': rng.randint(0, 6),
                    'exp': rng.randint(0, 6), 'flags': rng.choice([0, 0, 0x08, 0x10])}
        a, d = side(), side()
        kw = dict(distance=rng.randint(1, 5), a_terrain=rng.randint(0, 8), d_terrain=rng.randint(0, 8),
                  pincer=rng.choice([0, 25, 50, 100, 175, 250]), blocking=rng.choice([0, 80, 150]),
                  return_fire=rng.random() < 0.8)
        try:
            want = resolve(a, d, terrain, **kw)
        except ZeroDivisionError:
            want = 'divide error'
        try:
            got = dos.battle(a, d, distance=kw['distance'], atk_terrain=kw['a_terrain'], dfn_terrain=kw['d_terrain'],
                             pincer=kw['pincer'], blocking=kw['blocking'], return_fire=kw['return_fire'])
            got = {k: got[k] for k in ('attackerSize', 'attackerAlt', 'defenderSize', 'defenderAlt')}
        except ZeroDivisionError:
            got = 'divide error'
        if got != want:
            bad += 1
            if bad <= 5:
                print('MISMATCH', n, a['type'], d['type'], kw, 'emu', got, 'model', want)
    print('%d cases, %d mismatches' % (args.cases, bad))
    return 1 if bad else 0


if __name__ == '__main__':
    sys.exit(main())
