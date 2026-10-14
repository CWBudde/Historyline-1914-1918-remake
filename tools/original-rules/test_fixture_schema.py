"""Check the battle fixture schema against hand-written cases, using only the stdlib."""
import json
import os
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
SCHEMA = os.path.join(HERE, '..', '..', 'docs', 'research', 'battle-fixture.schema.json')


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
    kinds = {'object': dict, 'string': str, 'integer': int, 'boolean': bool}
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
            lambda c: c['attacker'].update(experience=7),
            lambda c: c['context'].update(blocking=151),
            lambda c: c['attacker'].update(type=51),
            lambda c: c.pop('expected'),
            lambda c: c['defender']['raw'].update({'2': 'zz'}),
            lambda c: c.update(extra=1),
            lambda c: c['source'].update(kind='guess'),
        ):
            c = sample()
            mutate(c)
            self.assertNotEqual(self.errors(c), [], mutate)


if __name__ == '__main__':
    unittest.main()
