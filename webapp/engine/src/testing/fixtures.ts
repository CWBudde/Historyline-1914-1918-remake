// Test-only: reading the fixture files of internal/dosrules/testdata/<build>/ as Go's UnmarshalFixtures does,
// refusing unknown fields, and the header check of dosrules_test.go.

import { expect } from "vitest";
import {
  type JSONObject,
  array,
  boolField,
  intField,
  object,
  present,
  stringField,
} from "../decode.ts";
import type {
  CasualtyExpected,
  CasualtyFixture,
  FixtureContext,
  FixtureHeader,
  FixtureSide,
  Source,
} from "../fixtures.ts";
import { type Profile, decodeType, tablesOf } from "../profile.ts";
import { readJSON } from "./files.ts";

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
