"""Check the fixture schemas against hand-written cases and the recorded fixtures, using only the stdlib."""
import json
import os
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
SCHEMA = os.path.join(HERE, '..', '..', 'docs', 'research', 'battle-fixture.schema.json')
ENGAGEMENT = os.path.join(HERE, '..', '..', 'docs', 'research', 'engagement-fixture.schema.json')
FIXTURES = os.path.join(HERE, '..', '..', 'internal', 'dosrules', 'testdata')


def check(schema, value, root, path='$'):
    """Validate the JSON Schema subset the fixture schema uses; return error strings."""
    if '$ref' in schema:
        schema = root['$defs'][schema['$ref'].rsplit('/', 1)[1]]
    errs = []
    if 'const' in schema and value != schema['const']:
        errs.append('%s: not %r' % (path, schema['const']))
    if 'enum' in schema and value not in schema['enum']:
        errs.append('%s: not in enum' % path)
    t = schema.get('type')
    kinds = {'object': dict, 'array': list, 'string': str, 'integer': int, 'boolean': bool}
    if t and (not isinstance(value, kinds[t]) or (t == 'integer' and isinstance(value, bool))):
        return errs + ['%s: not %s' % (path, t)]
    if isinstance(value, int) and not isinstance(value, bool):
        if 'minimum' in schema and value < schema['minimum']:
            errs.append('%s: below minimum' % path)
        if 'maximum' in schema and value > schema['maximum']:
            errs.append('%s: above maximum' % path)
    if isinstance(value, str):
        import re
        if len(value) < schema.get('minLength', 0):
            errs.append('%s: too short' % path)
        if 'pattern' in schema and not re.search(schema['pattern'], value):
            errs.append('%s: pattern' % path)
    if isinstance(value, list) and 'items' in schema:
        for i, v in enumerate(value):
            errs += check(schema['items'], v, root, '%s[%d]' % (path, i))
    if isinstance(value, dict):
        props = schema.get('properties', {})
        for k in schema.get('required', []):
            if k not in value:
                errs.append('%s: missing %s' % (path, k))
        for k, v in value.items():
            if k in props:
                errs += check(props[k], v, root, path + '.' + k)
            elif 'additionalProperties' in schema:
                ap = schema['additionalProperties']
                if ap is False:
                    errs.append('%s: unexpected %s' % (path, k))
                elif isinstance(ap, dict):
                    errs += check(ap, v, root, path + '.' + k)
    return errs


def sample():
    return {
        'schema': 1, 'profile': 'dos-0.1', 'id': 'inf-vs-inf-adjacent',
        'source': {'kind': 'transcription', 'build': 'fe4fa446'},
        'attacker': {'type': 18, 'size': 6, 'experience': 0},
        'defender': {'type': 18, 'size': 6, 'experience': 3, 'raw': {'2': '14'}},
        'context': {'distance': 1, 'attackerTerrainClass': 7, 'defenderTerrainClass': 0},
        'expected': {'attackerSizeAfter': 5, 'defenderSizeAfter': 4},
    }


class FixtureSchemaTest(unittest.TestCase):
    def setUp(self):
        with open(SCHEMA, encoding='utf-8') as f:
            self.schema = json.load(f)

    def errors(self, v):
        return check(self.schema, v, self.schema)

    def test_sample_valid(self):
        self.assertEqual(self.errors(sample()), [])

    def test_rejects_bad_cases(self):
        for mutate in (
            lambda c: c['attacker'].update(experience=256),
            lambda c: c['context'].update(blocking=256),
            lambda c: c['attacker'].update(type=51),
            lambda c: c.pop('expected'),
            lambda c: c['defender']['raw'].update({'2': 'zz'}),
            lambda c: c.update(extra=1),
            lambda c: c['source'].update(kind='guess'),
        ):
            c = sample()
            mutate(c)
            self.assertNotEqual(self.errors(c), [], mutate)

    def test_record_side(self):
        c = sample()
        c['attacker'] = {'record': {k: 1 for k in self.schema['$defs']['record']['required']}, 'size': 6, 'experience': 0}
        self.assertEqual(self.errors(c), [])
        c['attacker']['record']['armour'] = 256
        self.assertNotEqual(self.errors(c), [])


def engagement():
    return {
        'schema': 1, 'profile': 'dos-0.1', 'id': 'duel', 'source': {'kind': 'emulated-routine', 'build': '83fc4e37'},
        'width': 9, 'height': 9, 'side': 0, 'attacker': 0, 'target': 31, 'mapByte0': 7,
        'units': [{'index': 0, 'type': 18, 'cell': 40, 'size': 6, 'alternateSize': 0, 'experience': 0, 'flags4': 5, 'flags6': 0}],
        'terrain': [{'cell': 40, 'class': 3}],
        'expected': {'battle': True, 'defender': 1, 'returnFire': True, 'pincer': 0, 'blocking': 0, 'attackerSize': 5,
                     'attackerAlternateSize': 0, 'defenderSize': 5, 'defenderAlternateSize': 0,
                     'attackerExperience': 1, 'defenderExperience': 1, 'destroyed': []},
    }


class EngagementSchemaTest(unittest.TestCase):
    def setUp(self):
        with open(ENGAGEMENT, encoding='utf-8') as f:
            self.schema = json.load(f)

    def errors(self, v):
        return check(self.schema, v, self.schema)

    def test_sample_valid(self):
        self.assertEqual(self.errors(engagement()), [])

    def test_rejects_bad_cases(self):
        for mutate in (
            lambda c: c['units'][0].update(type=51),
            lambda c: c['units'][0].pop('cell'),
            lambda c: c['terrain'][0].update(height=1),
            lambda c: c['expected'].update(blocking=160),
            lambda c: c['expected'].update(destroyed=[-1]),
            lambda c: c.update(units={}),
        ):
            c = engagement()
            mutate(c)
            self.assertNotEqual(self.errors(c), [], mutate)


class RecordedFixturesTest(unittest.TestCase):
    """Every committed fixture of internal/dosrules follows its schema."""

    def test_recorded(self):
        schemas = {}
        for name, path in (('casualties.json', SCHEMA), ('engagements.json', ENGAGEMENT)):
            with open(path, encoding='utf-8') as f:
                schemas[name] = json.load(f)
        seen = 0
        for build in sorted(os.listdir(FIXTURES)):
            for name, schema in schemas.items():
                path = os.path.join(FIXTURES, build, name)
                with open(path, encoding='utf-8') as f:
                    cases = json.load(f)
                for c in cases:
                    self.assertEqual(check(schema, c, schema), [], '%s %s' % (path, c['id']))
                seen += len(cases)
        self.assertGreater(seen, 0)


if __name__ == '__main__':
    unittest.main()
