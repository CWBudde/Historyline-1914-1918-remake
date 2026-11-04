// Test-only: reading the fixture files of internal/dosrules/testdata/<build>/ as Go's UnmarshalFixtures does,
// refusing unknown fields, and the header check of dosrules_test.go.

import { describe, expect, it } from "vitest";
import {
  type JSONObject,
  array,
  arrayField,
  boolField,
  int,
  intField,
  object,
  present,
  stringField,
} from "../decode.ts";
import type {
  CasualtyExpected,
  CasualtyFixture,
  EngagementExpected,
  EngagementFixture,
  FixtureContext,
  FixtureHeader,
  FixtureSide,
  FixtureTerrain,
  FixtureUnit,
  OrdersExpected,
  OrdersFixture,
  Source,
} from "../fixtures.ts";
import { type Profile, decodeType, tablesOf } from "../profile.ts";
import { readJSON } from "./files.ts";
import { Profiles, moveTables } from "./tables.ts";

/** load reads testdata/<build>/<name> and decodes every record; the file must hold at least one. */
export function load<T>(
  build: string,
  name: string,
  decode: (v: unknown, what: string) => T,
): T[] {
  const what = `${build}/${name}`;
  const records = array(
    readJSON(`internal/dosrules/testdata/${build}/${name}`),
    what,
  );
  if (records.length === 0) {
    throw new Error(`${what} holds no fixtures`);
  }
  return records.map((r, i) => decode(r, `${what}[${i}]`));
}

/** checkHeader is dosrules_test.go checkHeader: schema 1, the profile's name and build, ids unique. */
export function checkHeader(
  p: Profile,
  h: FixtureHeader,
  seen: Set<string>,
): void {
  expect(
    h.schema === 1 &&
      h.profile === p.name &&
      h.source.kind === "emulated-routine" &&
      h.source.build === tablesOf(p).build,
    `${h.id}: header ${h.schema} ${h.profile} ${JSON.stringify(h.source)}`,
  ).toBe(true);
  expect(seen.has(h.id), `duplicate fixture id ${h.id}`).toBe(false);
  seen.add(h.id);
}

const headerKeys = ["schema", "profile", "id", "source"];

/** decodeHeader reads the envelope fields of a record already read as o. */
export function decodeHeader(o: JSONObject, what: string): FixtureHeader {
  return {
    schema: intField(o, "schema", what),
    profile: stringField(o, "profile", what),
    id: stringField(o, "id", what),
    source: decodeSource(o["source"], `${what}.source`),
  };
}

/** decodeSource reads a Source; note is absent when empty (omitempty). */
export function decodeSource(v: unknown, what: string): Source {
  const o = object(v ?? {}, what, ["kind", "build", "note"]);
  const s: Source = {
    kind: stringField(o, "kind", what),
    build: stringField(o, "build", what),
  };
  const note = stringField(o, "note", what);
  if (note !== "") {
    s.note = note;
  }
  return s;
}

function decodeSide(v: unknown, what: string): FixtureSide {
  const o = object(v ?? {}, what, [
    "type",
    "record",
    "size",
    "alternateSize",
    "experience",
    "flags",
  ]);
  const s: FixtureSide = {
    size: intField(o, "size", what),
    alternateSize: intField(o, "alternateSize", what),
    experience: intField(o, "experience", what),
    flags: intField(o, "flags", what),
  };
  if (present(o, "type")) {
    s.type = intField(o, "type", what);
  }
  if (present(o, "record")) {
    s.record = decodeType(o["record"], `${what}.record`, true);
  }
  return s;
}

function decodeContext(v: unknown, what: string): FixtureContext {
  const o = object(v ?? {}, what, [
    "distance",
    "attackerTerrainClass",
    "defenderTerrainClass",
    "pincer",
    "blocking",
    "returnFire",
  ]);
  return {
    distance: intField(o, "distance", what),
    attackerTerrainClass: intField(o, "attackerTerrainClass", what),
    defenderTerrainClass: intField(o, "defenderTerrainClass", what),
    pincer: intField(o, "pincer", what),
    blocking: intField(o, "blocking", what),
    returnFire: boolField(o, "returnFire", what),
  };
}

function decodeCasualtyExpected(v: unknown, what: string): CasualtyExpected {
  const o = object(v ?? {}, what, [
    "error",
    "attackerSizeAfter",
    "attackerAlternateAfter",
    "defenderSizeAfter",
    "defenderAlternateAfter",
  ]);
  const e: CasualtyExpected = {
    attackerSizeAfter: intField(o, "attackerSizeAfter", what),
    attackerAlternateAfter: intField(o, "attackerAlternateAfter", what),
    defenderSizeAfter: intField(o, "defenderSizeAfter", what),
    defenderAlternateAfter: intField(o, "defenderAlternateAfter", what),
  };
  const err = stringField(o, "error", what);
  if (err !== "") {
    e.error = err;
  }
  return e;
}

/** decodeCasualtyFixture reads one record of casualties.json. */
export function decodeCasualtyFixture(
  v: unknown,
  what: string,
): CasualtyFixture {
  const o = object(v, what, [
    ...headerKeys,
    "attacker",
    "defender",
    "context",
    "expected",
  ]);
  return {
    ...decodeHeader(o, what),
    attacker: decodeSide(o["attacker"], `${what}.attacker`),
    defender: decodeSide(o["defender"], `${what}.defender`),
    context: decodeContext(o["context"], `${what}.context`),
    expected: decodeCasualtyExpected(o["expected"], `${what}.expected`),
  };
}

function decodeUnit(v: unknown, what: string): FixtureUnit {
  const o = object(v ?? {}, what, [
    "index",
    "type",
    "cell",
    "size",
    "alternateSize",
    "experience",
    "flags4",
    "flags6",
  ]);
  return {
    index: intField(o, "index", what),
    type: intField(o, "type", what),
    cell: intField(o, "cell", what),
    size: intField(o, "size", what),
    alternateSize: intField(o, "alternateSize", what),
    experience: intField(o, "experience", what),
    flags4: intField(o, "flags4", what),
    flags6: intField(o, "flags6", what),
  };
}

function decodeTerrain(v: unknown, what: string): FixtureTerrain {
  const o = object(v ?? {}, what, ["cell", "class"]);
  return { cell: intField(o, "cell", what), class: intField(o, "class", what) };
}

// list reads o[key] as a slice of T with each element decoded by elem; null and absent are Go's nil slice.
function list<T>(
  o: JSONObject,
  key: string,
  what: string,
  elem: (v: unknown, what: string) => T,
): T[] | null {
  const a = arrayField(o, key, what);
  return a === null ? null : a.map((v, i) => elem(v, `${what}.${key}[${i}]`));
}

function decodeInt(v: unknown, what: string): number {
  return v === null ? 0 : int(v, what);
}

// decodePair reads a Go [2]int: null and missing elements are zero, elements past two an error (Go's decoder
// drops them; the fixtures never hold any, so the reader refuses them to catch a schema change).
function decodePair(v: unknown, what: string): [number, number] {
  if (v === null) {
    return [0, 0];
  }
  const p = array(v, what);
  if (p.length > 2) {
    throw new Error(`${what}: more than two elements`);
  }
  return [
    decodeInt(p[0] ?? 0, `${what}[0]`),
    decodeInt(p[1] ?? 0, `${what}[1]`),
  ];
}

function decodeEngagementExpected(
  v: unknown,
  what: string,
): EngagementExpected {
  const o = object(v ?? {}, what, [
    "error",
    "battle",
    "defender",
    "returnFire",
    "pincer",
    "blocking",
    "attackerSize",
    "attackerAlternateSize",
    "defenderSize",
    "defenderAlternateSize",
    "attackerExperience",
    "defenderExperience",
    "destroyed",
  ]);
  const e: EngagementExpected = {
    battle: boolField(o, "battle", what),
    defender: intField(o, "defender", what),
    returnFire: boolField(o, "returnFire", what),
    pincer: intField(o, "pincer", what),
    blocking: intField(o, "blocking", what),
    attackerSize: intField(o, "attackerSize", what),
    attackerAlternateSize: intField(o, "attackerAlternateSize", what),
    defenderSize: intField(o, "defenderSize", what),
    defenderAlternateSize: intField(o, "defenderAlternateSize", what),
    attackerExperience: intField(o, "attackerExperience", what),
    defenderExperience: intField(o, "defenderExperience", what),
    destroyed: list(o, "destroyed", what, decodeInt),
  };
  const err = stringField(o, "error", what);
  if (err !== "") {
    e.error = err;
  }
  return e;
}

/** decodeEngagementFixture reads one record of engagements.json. */
export function decodeEngagementFixture(
  v: unknown,
  what: string,
): EngagementFixture {
  const o = object(v, what, [
    ...headerKeys,
    "width",
    "height",
    "side",
    "attacker",
    "target",
    "mapByte0",
    "units",
    "terrain",
    "expected",
  ]);
  return {
    ...decodeHeader(o, what),
    width: intField(o, "width", what),
    height: intField(o, "height", what),
    side: intField(o, "side", what),
    attacker: intField(o, "attacker", what),
    target: intField(o, "target", what),
    mapByte0: intField(o, "mapByte0", what),
    units: list(o, "units", what, decodeUnit),
    terrain: list(o, "terrain", what, decodeTerrain),
    expected: decodeEngagementExpected(o["expected"], `${what}.expected`),
  };
}

function decodeOrdersExpected(v: unknown, what: string): OrdersExpected {
  const o = object(v ?? {}, what, [
    "error",
    "orders",
    "units",
    "gone",
    "occupants",
  ]);
  const e: OrdersExpected = {
    orders: list(o, "orders", what, decodeEngagementExpected),
    units: list(o, "units", what, decodeUnit),
    gone: list(o, "gone", what, decodeInt),
    occupants: list(o, "occupants", what, decodePair),
  };
  const err = stringField(o, "error", what);
  if (err !== "") {
    e.error = err;
  }
  return e;
}

/** decodeOrdersFixture reads one record of orders.json. */
export function decodeOrdersFixture(v: unknown, what: string): OrdersFixture {
  const o = object(v, what, [
    ...headerKeys,
    "width",
    "height",
    "side",
    "mapByte0",
    "units",
    "terrain",
    "orders",
    "seen",
    "idle",
    "expected",
  ]);
  return {
    ...decodeHeader(o, what),
    width: intField(o, "width", what),
    height: intField(o, "height", what),
    side: intField(o, "side", what),
    mapByte0: intField(o, "mapByte0", what),
    units: list(o, "units", what, decodeUnit),
    terrain: list(o, "terrain", what, decodeTerrain),
    orders: list(o, "orders", what, decodePair),
    seen: list(o, "seen", what, decodePair),
    idle: list(o, "idle", what, decodeInt),
    expected: decodeOrdersExpected(o["expected"], `${what}.expected`),
  };
}

/**
 * replay is motion_test.go replay: per build, every record of testdata/<build>/<name> (exactly count of
 * them) passes the header check and run (which compares with expect), and every class (the id up to its
 * first "/") in classes is present.
 */
export function replay<T extends FixtureHeader>(
  name: string,
  decode: (v: unknown, what: string) => T,
  count: number,
  run: (p: Profile, f: T) => void,
  ...classes: string[]
): void {
  describe.each(Profiles)(`${name}, %s build`, (build, p) => {
    it(`every record of ${name}`, () => {
      moveTables(p);
      const fs = load(build, name, decode);
      expect(fs).toHaveLength(count);
      const seen = new Map<string, number>();
      const ids = new Set<string>();
      let passed = 0;
      for (const f of fs) {
        checkHeader(p, f, ids);
        run(p, f);
        passed++;
        const cut = f.id.indexOf("/");
        const class_ = cut < 0 ? f.id : f.id.slice(0, cut);
        seen.set(class_, (seen.get(class_) ?? 0) + 1);
      }
      expect(passed).toBe(count);
      for (const c of classes) {
        expect(
          seen.get(c) ?? 0,
          `${build}: no recorded ${c} case`,
        ).toBeGreaterThan(0);
      }
    });
  });
}
