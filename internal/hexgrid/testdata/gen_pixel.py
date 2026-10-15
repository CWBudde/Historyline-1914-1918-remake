"""Write pixel.json: the Python reference findHex result for a grid of map pixels.

findHex lives in main.py as a method that only reads its arguments, so the function
is taken from the source without importing pygame. The mouse position passed to it
is the map pixel plus the 21 pixel border. Run from the repo root:

    python3 -I internal/hexgrid/testdata/gen_pixel.py
"""
import ast
import json
import os

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..', '..'))
OUT = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'pixel.json')

with open(os.path.join(ROOT, 'legacy', 'python', 'main.py'), encoding='utf-8') as f:
    tree = ast.parse(f.read())
fn = next(m for c in tree.body if isinstance(c, ast.ClassDef) for m in c.body
          if isinstance(m, ast.FunctionDef) and m.name == 'findHex')
scope = {}
exec(compile(ast.Module(body=[fn], type_ignores=[]), 'main.py', 'exec'), scope)
findHex = scope['findHex']

STEP = 8
samples = []
for y in range(0, 23 * 40 + 41, STEP):
    for x in range(0, 8 * 142 + 1, STEP):
        r = findHex(None, x + 21, y + 21)
        # findHex returns [column, line] or False
        samples.append([x, y] + ([r[1], r[0]] if r else [-1, -1]))

with open(OUT, 'w', encoding='utf-8') as f:
    json.dump({'step': STEP, 'width': 8, 'lines': 23, 'samples': samples}, f, separators=(',', ':'))
    f.write('\n')
