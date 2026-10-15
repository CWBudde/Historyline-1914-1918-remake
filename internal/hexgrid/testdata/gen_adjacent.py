"""Write adjacent.json: the Python reference neighbours of every hex of every shipped level.

hlrData.py starts pygame on import, so only the pure adjacentHexes function is taken
from its source. Run from the repo root:

    python3 -I internal/hexgrid/testdata/gen_adjacent.py
"""
import ast
import json
import os

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..', '..'))
OUT = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'adjacent.json')

with open(os.path.join(ROOT, 'legacy', 'python', 'hlrData.py'), encoding='utf-8') as f:
    tree = ast.parse(f.read())
fn = next(n for n in tree.body if isinstance(n, ast.FunctionDef) and n.name == 'adjacentHexes')
scope = {}
exec(compile(ast.Module(body=[fn], type_ignores=[]), 'hlrData.py', 'exec'), scope)
adjacentHexes = scope['adjacentHexes']

levels = {}
for name in sorted(os.listdir(os.path.join(ROOT, 'levels'))):
    if not name.endswith('.json'):
        continue
    with open(os.path.join(ROOT, 'levels', name), encoding='utf-8') as f:
        tiles = json.load(f)['tiles']
    width, lines = len(tiles['line1']), len(tiles)  # Info.mapWidth, Info.mapHeight
    hexes = []
    for x in range(lines):
        for y in range(len(tiles['line%d' % (x + 1)])):
            hexes.append({'hex': [x, y], 'neighbors': [list(c) for c in adjacentHexes(x, y, width, lines)]})
    levels[name] = {'width': width, 'lines': lines, 'hexes': hexes}

with open(OUT, 'w', encoding='utf-8') as f:
    json.dump(levels, f, separators=(',', ':'))
    f.write('\n')
