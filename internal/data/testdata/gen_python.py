"""Write python.json: the Python reference data and infrastructure-rule results.

hlrData.py starts pygame on import, so the dict literals and the infrastructure loop
of HexSquare.__init__ are taken from its source and run against stubs. Run from the
repo root:

    python3 -I internal/data/testdata/gen_python.py
"""
import ast
import json
import os

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..', '..'))
OUT = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'python.json')

with open(os.path.join(ROOT, 'legacy', 'python', 'hlrData.py'), encoding='utf-8') as f:
    tree = ast.parse(f.read())


class _Stub:
    """Stands in for pygame and the surfaces the infrastructure loop touches."""

    def __getattr__(self, _):
        return self

    def __call__(self, *a, **k):
        return self

    def __getitem__(self, _):
        return self


def literal(name):
    node = next(n for n in tree.body if isinstance(n, ast.Assign) and n.targets[0].id == name)
    return eval(compile(ast.Expression(node.value), 'hlrData.py', 'eval'), {'pygame': _Stub()})


terrain = literal('bgTilesModifiers')
infra_ids = [k for k in literal('infraIcons') if k != '<none>']
units = {k: {f: v for f, v in u.items() if f not in ('icon', 'picture')}
         for k, u in literal('unitsParameters').items() if u}
weapons = {k: {f: v for f, v in w.items() if f != 'picture'}
           for k, w in literal('weaponsParameters').items()}

# The infrastructure part of HexSquare.__init__: from `_infrastructure = []` to the end.
hexsq = next(n for n in tree.body if isinstance(n, ast.ClassDef) and n.name == 'HexSquare')
init = next(n for n in hexsq.body if isinstance(n, ast.FunctionDef) and n.name == '__init__')
start = next(i for i, s in enumerate(init.body)
             if isinstance(s, ast.Assign) and getattr(s.targets[0], 'id', '') == '_infrastructure')
loop = compile(ast.Module(body=init.body[start:], type_ignores=[]), 'hlrData.py', 'exec')


def apply(hex_type, infrastructure):
    self = _Stub()
    self.__dict__.update(movementModifier=terrain[hex_type][0], battleModifier=terrain[hex_type][1],
                         sightModifier=terrain[hex_type][2], movementAllowed=[0], infra=None, bgGrey=_Stub())
    exec(loop, {'infraIcons': _Stub(), 'greyscale': _Stub()},
         {'self': self, 'infrastructure': infrastructure})
    return {'terrain': hex_type, 'infra': infrastructure if isinstance(infrastructure, list) else [infrastructure],
            'movement': self.movementModifier, 'battle': self.battleModifier, 'sight': self.sightModifier,
            'movementAllowed': self.movementAllowed}


cases = [('grass', i) for i in infra_ids]
cases += [('grass', 'barbedWire'), ('grass', 'lake12'), ('forest', 'barbedWire'),
          ('grass', ['road13', 'rail24']), ('grass', ['rail24', 'road13']), ('grass', ['stream13', 'rail24']),
          ('grass', ['road25', 'stream13']), ('forest', ['trench12', 'path13']), ('house', ['bridge25', 'trench5'])]
for name in sorted(os.listdir(os.path.join(ROOT, 'levels'))):
    if not name.endswith('.json'):
        continue
    with open(os.path.join(ROOT, 'levels', name), encoding='utf-8') as f:
        for line in json.load(f)['tiles'].values():
            for hex_type, infrastructure, *_ in line:
                if infrastructure:
                    cases.append((hex_type, infrastructure))
seen, results = set(), []
for hex_type, infrastructure in cases:
    key = json.dumps([hex_type, infrastructure])
    if key not in seen:
        seen.add(key)
        results.append(apply(hex_type, infrastructure))

with open(OUT, 'w', encoding='utf-8') as f:
    json.dump({'terrain': terrain, 'units': units, 'weapons': weapons, 'infra': results}, f, indent=1)
    f.write('\n')
