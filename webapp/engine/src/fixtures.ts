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
import { type Battlefield, type Outcome, type Unit, unitOf } from "./engage.ts";
import { GoError, at, errorsIs } from "./goruntime.ts";
import { type Order, type Resolution, OrdersError } from "./orders.ts";
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

/** FixtureUnit is a live unit of an engagement fixture. */
export interface FixtureUnit {
  index: number;
  type: number;
  cell: number;
  size: number;
  alternateSize: number;
  experience: number;
  flags4: number;
  flags6: number;
}

/** FixtureTerrain is a cell's terrain class. */
export interface FixtureTerrain {
  cell: number;
  class: number;
}

/** EngagementExpected is the state the order handler left, or its error. destroyed null is Go's nil slice. */
export interface EngagementExpected {
  error?: string;
  battle: boolean;
  defender: number;
  returnFire: boolean;
  pincer: number;
  blocking: number;
  attackerSize: number;
  attackerAlternateSize: number;
  defenderSize: number;
  defenderAlternateSize: number;
  attackerExperience: number;
  defenderExperience: number;
  destroyed: number[] | null;
}

/** EngagementFixture is one recorded attack order through the order handler. Null slices are Go's nil. */
export interface EngagementFixture extends FixtureHeader {
  width: number;
  height: number;
  /** The player whose orders run (no effect on the outcome). */
  side: number;
  attacker: number;
  target: number;
  mapByte0: number;
  units: FixtureUnit[] | null;
  terrain: FixtureTerrain[] | null;
  expected: EngagementExpected;
}

/** fixtureBattlefield builds a Battlefield from fixture units and terrain (Go's EngagementFixture.Battlefield). */
function fixtureBattlefield(
  width: number,
  height: number,
  mapByte0: number,
  units: readonly FixtureUnit[] | null,
  terrain: readonly FixtureTerrain[] | null,
): Battlefield {
  const b: Battlefield = {
    width,
    height,
    units: new Map(),
    terrain: new Map(),
    mapByte0,
  };
  for (const u of units ?? []) {
    b.units.set(u.index, {
      type: u.type,
      cell: u.cell,
      size: u.size,
      alt: u.alternateSize,
      exp: u.experience,
      flags4: u.flags4,
      flags6: u.flags6,
    });
  }
  for (const c of terrain ?? []) {
    b.terrain.set(c.cell, c.class);
  }
  return b;
}

/** engagementBattlefield returns the fixture's map (Go's EngagementFixture.Battlefield). */
export function engagementBattlefield(f: EngagementFixture): Battlefield {
  return fixtureBattlefield(f.width, f.height, f.mapByte0, f.units, f.terrain);
}

function fixtureUnit(index: number, u: Unit): FixtureUnit {
  return {
    index,
    type: u.type,
    cell: u.cell,
    size: u.size,
    alternateSize: u.alt,
    experience: u.exp,
    flags4: u.flags4,
    flags6: u.flags6,
  };
}

/** engagement returns a fixture of a battlefield and an attack order, with nothing expected yet. */
export function engagement(
  b: Battlefield,
  attacker: number,
  target: number,
  side: number,
): EngagementFixture {
  const units = [...b.units]
    .map(([idx, u]) => fixtureUnit(idx, u))
    .sort((x, y) => x.index - y.index);
  const terrain = [...b.terrain]
    .map(([cell, class_]) => ({ cell, class: class_ }))
    .sort((x, y) => x.cell - y.cell);
  return {
    schema: 1,
    profile: "",
    id: "",
    source: { kind: "", build: "" },
    width: b.width,
    height: b.height,
    side,
    attacker,
    target,
    mapByte0: b.mapByte0,
    units,
    terrain,
    expected: zeroExpected(),
  };
}

// zeroExpected is the zero EngagementExpected (destroyed nil).
function zeroExpected(): EngagementExpected {
  return {
    battle: false,
    defender: 0,
    returnFire: false,
    pincer: 0,
    blocking: 0,
    attackerSize: 0,
    attackerAlternateSize: 0,
    defenderSize: 0,
    defenderAlternateSize: 0,
    attackerExperience: 0,
    defenderExperience: 0,
    destroyed: null,
  };
}

/**
 * engagementResult is the fixture form of an engage outcome: engagementResult(...attempt(() => engage(…))) as
 * Go's EngagementResult(p.Engage(…)).
 */
export function engagementResult(
  ...[o, err]: readonly [Outcome, null] | readonly [null, Error]
): EngagementExpected {
  if (err !== null) {
    return { ...zeroExpected(), error: errorName(err), destroyed: [] };
  }
  return {
    battle: o.battle,
    defender: o.defender,
    returnFire: o.returnFire,
    pincer: o.pincer,
    blocking: o.blocking,
    attackerSize: o.attackerSize,
    attackerAlternateSize: o.attackerAlt,
    defenderSize: o.defenderSize,
    defenderAlternateSize: o.defenderAlt,
    attackerExperience: o.attackerExp,
    defenderExperience: o.defenderExp,
    destroyed: [...o.destroyed],
  };
}

/**
 * OrdersExpected is the state the order handler left after a whole attack list, or the error that stopped it:
 * each order's outcome in list order, every unit afterwards, the units the destroy helper took off the map and
 * [cell, unit] for every occupied cell of the attacking side's map. Null slices are Go's nil.
 */
export interface OrdersExpected {
  error?: string;
  orders: EngagementExpected[] | null;
  units: FixtureUnit[] | null;
  gone: number[] | null;
  occupants: [number, number][] | null;
}

/**
 * OrdersFixture is one recorded run of the order handler over an attack list. The battlefield is the attacking
 * side's (side) view; seen gives [unit, cell] for the units the other side sees elsewhere, and idle is the
 * other side's list, which the handler leaves alone.
 */
export interface OrdersFixture extends FixtureHeader {
  width: number;
  height: number;
  side: number;
  mapByte0: number;
  units: FixtureUnit[] | null;
  terrain: FixtureTerrain[] | null;
  /** [attacker, target cell] in list order. */
  orders: [number, number][] | null;
  seen: [number, number][] | null;
  idle: number[] | null;
  expected: OrdersExpected;
}

/** ordersInput returns a fixture of a battlefield and an attack list, with nothing expected yet. */
export function ordersInput(
  b: Battlefield,
  orders: readonly Order[],
  side: number,
): OrdersFixture {
  const e = engagement(b, 0, 0, side);
  return {
    schema: 1,
    profile: "",
    id: "",
    source: { kind: "", build: "" },
    width: b.width,
    height: b.height,
    side,
    mapByte0: b.mapByte0,
    units: e.units,
    terrain: e.terrain,
    orders: orders.map((o) => [o.attacker, o.target]),
    seen: [],
    idle: [],
    expected: { orders: null, units: null, gone: null, occupants: null },
  };
}

/** ordersBattlefield returns the fixture's map as the attacking side sees it (Go's OrdersFixture.Battlefield). */
export function ordersBattlefield(f: OrdersFixture): Battlefield {
  return fixtureBattlefield(f.width, f.height, f.mapByte0, f.units, f.terrain);
}

/** ordersList returns the fixture's attack list (Go's OrdersFixture.List). */
export function ordersList(f: OrdersFixture): Order[] {
  return (f.orders ?? []).map(([attacker, target]) => ({ attacker, target }));
}

/**
 * ordersResult is the fixture form of a resolveOrders result: ordersResult(...attempt(() => resolveOrders(…)))
 * as Go's OrdersResult(p.ResolveOrders(…)). On an error it reads the partial resolution an OrdersError
 * carries, which Go returns beside the error.
 */
export function ordersResult(
  ...[res, err]: readonly [Resolution, null] | readonly [null, Error]
): OrdersExpected {
  let r: Resolution;
  if (err === null) {
    r = res;
  } else if (err instanceof OrdersError) {
    r = err.resolution;
  } else {
    r = { orders: [], units: new Map(), gone: [], cells: new Map() };
  }
  const x: OrdersExpected = {
    orders: r.orders.map((o) => engagementResult(o, null)),
    units: [...r.units.keys()]
      .sort((a, b) => a - b)
      .map((i) => fixtureUnit(i, unitOf(r.units, i))),
    gone: [...r.gone],
    occupants: [...r.cells.keys()]
      .sort((a, b) => a - b)
      .map((c): [number, number] => [c, r.cells.get(c) ?? 0]),
  };
  if (err !== null) {
    x.error = errorName(err);
  }
  return x;
}
