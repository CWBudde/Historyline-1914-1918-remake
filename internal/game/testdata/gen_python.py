"""Write python.json: fog of war, movement and battle results of the Python reference.

gui.py and hlrData.py start pygame on import, so the rule functions are taken from
their source (calculateFOW, getMovableSquares, markMovableSquares, findPath, the
calculation part of handleBattle, adjacentHexes, Content) and run against stub
hexes and units. Buildings follow internal/game: a factory piece other than
factoryE that has content in the level is a container too. Run from the repo root:

    python3 -I internal/game/testdata/gen_python.py
"""
import ast
import copy
import json
import os
import random

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..', '..'))
OUT = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'python.json')


def parse(name):
    with open(os.path.join(ROOT, 'legacy', 'python', name), encoding='utf-8') as f:
        return ast.parse(f.read())


hlr, gui = parse('hlrData.py'), parse('gui.py')


class _Stub:
    def __getattr__(self, _):
        return self

    def __call__(self, *a, **k):
        return self

    def __getitem__(self, _):
        return self


def literal(name):
    node = next(n for n in hlr.body if isinstance(n, ast.Assign) and n.targets[0].id == name)
    return eval(compile(ast.Expression(node.value), 'hlrData.py', 'eval'), {'pygame': _Stub()})


def define(tree, filename, names, scope):
    nodes = [n for n in tree.body if getattr(n, 'name', None) in names]
    exec(compile(ast.Module(body=nodes, type_ignores=[]), filename, 'exec'), scope)


terrain = literal('bgTilesModifiers')
unitsParameters = literal('unitsParameters')
weaponsParameters = literal('weaponsParameters')
scope = {'copy': copy}
define(hlr, 'hlrData.py', {'adjacentHexes', 'Content'}, scope)
adjacentHexes, Content = scope['adjacentHexes'], scope['Content']

# The infrastructure part of HexSquare.__init__, as in internal/data/testdata/gen_python.py.
hexsq = next(n for n in hlr.body if isinstance(n, ast.ClassDef) and n.name == 'HexSquare')
init = next(n for n in hexsq.body if isinstance(n, ast.FunctionDef) and n.name == '__init__')
start = next(i for i, s in enumerate(init.body)
             if isinstance(s, ast.Assign) and getattr(s.targets[0], 'id', '') == '_infrastructure')
infra_loop = compile(ast.Module(body=init.body[start:], type_ignores=[]), 'hlrData.py', 'exec')

# GUI methods, compiled into a class of our own.
gui_class = next(n for n in gui.body if isinstance(n, ast.ClassDef) and n.name == 'GUI')
methods = {n.name: n for n in gui_class.body if isinstance(n, ast.FunctionDef)}
cls = ast.ClassDef(name='Rules', bases=[], keywords=[], decorator_list=[], type_params=[],
                   body=[methods[m] for m in ('getMovableSquares', 'markMovableSquares', 'findPath',
                                              'calculateFOW', 'getSquare', 'calculateDistance')])
exec(compile(ast.fix_missing_locations(ast.Module(body=[cls], type_ignores=[])), 'gui.py', 'exec'), scope)
Rules = scope['Rules']

# handleBattle up to the first rendering statement.
battle = methods['handleBattle'].body
cut = next(i for i, s in enumerate(battle) if 'font40' in ast.unparse(s))
battle_code = compile(ast.Module(body=battle[:cut], type_ignores=[]), 'gui.py', 'exec')


class Obj:
    def __init__(self, **kw):
        self.__dict__.update(kw)


def weapon(key):
    if not key:
        return None
    w = weaponsParameters[key]
    return Obj(name=w['name'], rangeMin=w['rangeMin'], rangeMax=w['rangeMax'], power=w['power'], ammo=w['ammo'])


def unit(key):
    d = unitsParameters[key]
    return Obj(key=key, name=d['name'], armour=d['armour'], speed=d['speed'], weight=d['weight'], sight=d['sight'],
               experience=0, skills=d['skills'], movementType=d['movementType'], currentSize=10, moved=False,
               faction=1 if d['country'] in ['Germany', 'Austria', 'Bulgaria', 'Ottoman'] else 2,
               content=Content(d['storageMax']) if d['storageMax'] > 0 else False,
               weapons=[weapon(w) for w in d['weapons']])


def capacity(hex_type, has_content):
    base = hex_type[:-2] if hex_type.endswith('_w') else hex_type
    if base == 'hqN':
        return 50
    if base in ('cmpN', 'factoryE'):
        return 40
    if has_content and hex_type.startswith('factory'):
        return 40
    return 0


def hex_square(pos, hex_type, infrastructure, unit_key, content=None):
    h = Obj(position=pos, unit=unit(unit_key) if unit_key else None, unitToMove=None, fogofwar=None, seen=False,
            movementModifier=terrain[hex_type][0], battleModifier=terrain[hex_type][1],
            sightModifier=terrain[hex_type][2], movementAllowed=[0], infra=None, bgGrey=_Stub(), content=False)
    exec(infra_loop, {'infraIcons': _Stub(), 'greyscale': _Stub()}, {'self': h, 'infrastructure': infrastructure})
    cap = capacity(hex_type, content is not None)
    if cap:
        h.owner = content[0] if content else 0
        h.content = Content(cap)
        if content and len(content) == 2:
            for k in content[1]:
                h.content.addUnit(unit(k))
    return h


def game(path, player):
    with open(path, encoding='utf-8') as f:
        tiles = json.load(f)['tiles']
    g = Rules()
    g.parent = Obj(info=Obj(player=player, opponent=1 if player == 2 else 2, mapWidth=len(tiles['line1']),
                            mapHeight=len(tiles)), playerTurn=player, cmdArgs=Obj(reveal=False))
    g.mainMap = [[hex_square((x, y), *sq) for y, sq in enumerate(line)] for x, line in enumerate(tiles.values())]
    return g


def contained(content):
    return [u for row in content.units for u in row if u]


def movers(g, player):
    """Units the human could pick: own map units, and own unmoved units in own carriers or
    in buildings not owned by the opponent (ContentMenu.endMenu)."""
    out = []
    for line in g.mainMap:
        for h in line:
            if h.unit and h.unit.faction == player:
                if h.unit.speed:
                    out.append((h, h.unit, 'map', -1, -1))
                if h.unit.content:
                    for r, row in enumerate(h.unit.content.units):
                        for c, u in enumerate(row):
                            if u and u.faction == player:
                                out.append((h, u, 'carrier', r, c))
            elif h.content and h.owner != g.parent.info.opponent:
                for r, row in enumerate(h.content.units):
                    for c, u in enumerate(row):
                        if u and u.faction == player:
                            out.append((h, u, 'building', r, c))
    return out


levels = []
for name in sorted(os.listdir(os.path.join(ROOT, 'levels'))):
    if not name.endswith('.json'):
        continue
    for player in (1, 2):
        g = game(os.path.join(ROOT, 'levels', name), player)
        g.calculateFOW()
        fog = [[h.fogofwar for h in line] for line in g.mainMap]
        moves = []
        for h, u, where, r, c in movers(g, player):
            g.calculateFOW()
            if where != 'map':
                h.unitToMove = u
            movable = g.markMovableSquares(h)
            targets = {}
            for x, y in sorted(movable):
                if g.mainMap[x][y].fogofwar in (0, 2):
                    targets['%d,%d' % (x, y)] = [list(p) for p in g.findPath(h.position, (x, y))]
            h.unitToMove = None
            moves.append({'from': list(h.position), 'unit': u.key, 'in': where, 'row': r, 'col': c,
                          'targets': targets})
        levels.append({'level': name, 'player': player, 'fog': fog, 'moves': moves})


# Battle cases: random but fixed combinations of the formula's inputs.
rng = random.Random(1914)
battle_values = sorted({v[1] for v in terrain.values() if v[1] is not None} | {0, 10})
armed = sorted(k for k, d in unitsParameters.items() if d and any(d['weapons']))
cases = []
for _ in range(600):
    att, dfn = unit(rng.choice(armed)), unit(rng.choice(armed))
    slot = rng.choice([i for i, w in enumerate(att.weapons) if w])
    w = att.weapons[slot]
    steps = rng.randint(max(1, w.rangeMin), w.rangeMax)
    att.experience, dfn.experience = rng.randint(0, 10), rng.randint(0, 10)
    att.currentSize, dfn.currentSize = rng.randint(1, 10), rng.randint(1, 10)
    empty = [i for i, x in enumerate(dfn.weapons) if x and rng.random() < 0.2]
    for i in empty:
        dfn.weapons[i].ammo = 0
    rolls = [rng.randint(-5, 5), rng.randint(-5, 5)]
    fb, tb = rng.choice(battle_values), rng.choice(battle_values)
    queue = list(rolls)
    ns = {'self': Obj(calculateDistance=lambda a, b, r, s=steps: s),
          'attackFromSquare': Obj(position=None, battleModifier=fb, unit=att),
          'attackToSquare': Obj(position=None, battleModifier=tb, unit=dfn),
          'weapon': w, 'random': Obj(randint=lambda lo, hi: queue.pop(0)), 'sorted': sorted}
    exec(battle_code, ns)
    counter = -1 if ns['enemyWeapon'] is None else next(i for i, x in enumerate(dfn.weapons) if x is ns['enemyWeapon'])
    cases.append({'attacker': att.key, 'defender': dfn.key, 'slot': slot, 'steps': steps,
                  'attackerExp': att.experience, 'defenderExp': dfn.experience,
                  'attackerSize': att.currentSize, 'defenderSize': dfn.currentSize,
                  'defenderEmpty': empty, 'attackerBattle': fb, 'defenderBattle': tb, 'rolls': rolls,
                  'counter': counter, 'attackerKills': ns['fFinal'], 'defenderKills': ns['eFinal']})

with open(OUT, 'w', encoding='utf-8') as f:
    json.dump({'levels': levels, 'battles': cases}, f, separators=(',', ':'))
    f.write('\n')
