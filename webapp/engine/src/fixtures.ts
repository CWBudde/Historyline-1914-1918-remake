// Fixtures record expected results as numbers only (internal/dosrules/testdata/). The types mirror the JSON
// exactly: Go's omitempty fields are optional here and absent when zero, so a decoded record and an adapter's
// result compare with toStrictEqual as Go's == and reflect.DeepEqual do. Reading the files is test-only
// (src/testing/); the types and adapters are engine code, as in Go.

import {
  type Context,
  type Fighter,
  type Result,
  ErrDivideError,
} from "./casualties.ts";
import { GoError, at, errorsIs } from "./goruntime.ts";
import type { Tables, Type } from "./profile.ts";
import { ErrUndefined } from "./support.ts";

/** Source says how a fixture's expectation was obtained. */
export interface Source {
  /** emulated-routine, controlled-game, transcription, manual. */
  kind: string;
  /** SHA-256 prefix of the executable. */
  build: string;
  note?: string;
}

/** FixtureHeader is the envelope every fixture record shares (no Go type: each fixture repeats its fields). */
export interface FixtureHeader {
  schema: number;
  profile: string;
  id: string;
  source: Source;
}

function isObject(v: unknown): v is Record<string, unknown> {
  return typeof v === "object" && v !== null && !Array.isArray(v);
}

/** isFixtureHeader reports whether the parsed JSON v carries the fixture envelope. */
export function isFixtureHeader(v: unknown): v is FixtureHeader {
  if (!isObject(v) || !isObject(v["source"])) {
    return false;
  }
  const src = v["source"];
  return (
    typeof v["schema"] === "number" &&
    typeof v["profile"] === "string" &&
    typeof v["id"] === "string" &&
    typeof src["kind"] === "string" &&
    typeof src["build"] === "string" &&
    (src["note"] === undefined || typeof src["note"] === "string")
  );
}

/** FixtureSide is one side of a casualty fixture: a type of the build's tables or a synthetic record, and the live values. */
export interface FixtureSide {
  type?: number;
  record?: Type;
  size: number;
  alternateSize: number;
  experience: number;
  flags: number;
}

/** FixtureContext are a casualty fixture's battle inputs. */
export interface FixtureContext {
  distance: number;
  attackerTerrainClass: number;
  defenderTerrainClass: number;
  pincer: number;
  blocking: number;
  returnFire: boolean;
}

/** CasualtyExpected is the routine's result, or the error it raised. */
export interface CasualtyExpected {
  error?: string;
  attackerSizeAfter: number;
  attackerAlternateAfter: number;
  defenderSizeAfter: number;
  defenderAlternateAfter: number;
}

/** CasualtyFixture is one recorded run of the casualty routine. */
export interface CasualtyFixture extends FixtureHeader {
  attacker: FixtureSide;
  defender: FixtureSide;
  context: FixtureContext;
  expected: CasualtyExpected;
}

function fighter(s: FixtureSide, t: Tables): Fighter {
  let typ: Type;
  if (s.record !== undefined) {
    typ = { ...s.record };
  } else if (s.type !== undefined && s.type >= 0 && s.type < t.types.length) {
    typ = { ...at(t.types, s.type) };
  } else {
    throw new GoError("fixture side has neither a known type nor a record");
  }
  return {
    type: typ,
    size: s.size,
    alt: s.alternateSize,
    exp: s.experience,
    flags: s.flags,
  };
}

/** casualtyInputs returns the fixture's fighters and context (Go's CasualtyFixture.Inputs). */
export function casualtyInputs(
  f: CasualtyFixture,
  t: Tables,
): [a: Fighter, d: Fighter, c: Context] {
  const a = fighter(f.attacker, t);
  const d = fighter(f.defender, t);
  const x = f.context;
  return [
    a,
    d,
    {
      distance: x.distance,
      aTerrain: x.attackerTerrainClass,
      dTerrain: x.defenderTerrainClass,
      pincer: x.pincer,
      blocking: x.blocking,
      returnFire: x.returnFire,
    },
  ];
}

/** errorName is the fixture spelling of an error ("" for none). */
export function errorName(err: Error | null): string {
  if (err === null) {
    return "";
  }
  if (errorsIs(err, ErrDivideError)) {
    return "divide-error";
  }
  if (errorsIs(err, ErrUndefined)) {
    return "undefined";
  }
  return err.message;
}

/**
 * casualtyResult is the fixture form of a casualties result: CasualtyResult(...attempt(() => casualties(…)))
 * as Go's CasualtyResult(p.Casualties(…)).
 */
export function casualtyResult(
  ...[r, err]: readonly [Result, null] | readonly [null, Error]
): CasualtyExpected {
  if (err !== null) {
    return {
      error: errorName(err),
      attackerSizeAfter: 0,
      attackerAlternateAfter: 0,
      defenderSizeAfter: 0,
      defenderAlternateAfter: 0,
    };
  }
  return {
    attackerSizeAfter: r.attackerSize,
    attackerAlternateAfter: r.attackerAlt,
    defenderSizeAfter: r.defenderSize,
    defenderAlternateAfter: r.defenderAlt,
  };
}
