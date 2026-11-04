// Test-only: strict readers of the movement, step, container and visibility fixtures, as Go's UnmarshalFixtures
// fills the types of motion_fixtures.go (unknown fields refused, absent fields zero).

import type { ExitOrder } from "../container.ts";
import {
  type JSONObject,
  array,
  boolField,
  intField,
  object,
  present,
} from "../decode.ts";
import type {
  ContainerExpected,
  ContainerFixture,
  FixtureHold,
  FixtureMap,
  FixturePiece,
  FixtureRecord,
  MovementExpected,
  MovementFixture,
  Pair,
  StepExpected,
  StepFixture,
  FixtureSighted,
  VisibilityExpected,
  VisibilityFixture,
} from "../motionFixtures.ts";
import type { MoveOrder } from "../record.ts";
import {
  decodeHeader,
  decodeInt,
  decodePair,
  headerKeys,
  list,
} from "./fixtures.ts";

const mapKeys = [
  "width",
  "height",
  "fill",
  "cells",
  "units",
  "unseen",
  "unit",
  "side",
  "budget",
  "block",
];

// inputs reads an input list: null and absent are empty, as Go ranges a nil slice.
function inputs<T>(
  o: JSONObject,
  key: string,
  what: string,
  elem: (v: unknown, what: string) => T,
): T[] {
  return list(o, key, what, elem) ?? [];
}

// decodeFixed reads a Go [n]int: null and missing elements are zero, elements past n an error (Go's decoder
// drops them; the fixtures never hold any).
function decodeFixed(v: unknown, n: number, what: string): number[] {
  const out = Array.from({ length: n }, () => 0);
  if (v === null || v === undefined) {
    return out;
  }
  const a = array(v, what);
  if (a.length > n) {
    throw new Error(`${what}: more than ${n} elements`);
  }
  for (const [i, x] of a.entries()) {
    out[i] = decodeInt(x, `${what}[${i}]`);
  }
  return out;
}

function decodePiece(v: unknown, what: string): FixturePiece {
  const o = object(v ?? {}, what, [
    "index",
    "type",
    "cell",
    "flags4",
    "flags6",
    "hold",
  ]);
  return {
    index: intField(o, "index", what),
    type: intField(o, "type", what),
    cell: intField(o, "cell", what),
    flags4: intField(o, "flags4", what),
    flags6: intField(o, "flags6", what),
    hold: intField(o, "hold", what),
  };
}

function decodeMap(o: JSONObject, what: string): FixtureMap {
  return {
    width: intField(o, "width", what),
    height: intField(o, "height", what),
    fill: intField(o, "fill", what),
    cells: inputs(o, "cells", what, decodePair),
    units: inputs(o, "units", what, decodePiece),
    unseen: inputs(o, "unseen", what, decodePiece),
    unit: intField(o, "unit", what),
    side: intField(o, "side", what),
    budget: intField(o, "budget", what),
    block: intField(o, "block", what),
  };
}

function decodeRoute(v: unknown, what: string): number[] | null {
  return v === null
    ? null
    : array(v, what).map((x, i) => decodeInt(x, `${what}[${i}]`));
}

function decodeMovementExpected(v: unknown, what: string): MovementExpected {
  const o = object(v ?? {}, what, [
    "remaining",
    "count",
    "destinations",
    "codes",
    "paths",
  ]);
  return {
    remaining: inputs(o, "remaining", what, decodeInt),
    count: intField(o, "count", what),
    destinations: list(o, "destinations", what, decodeInt),
    codes: inputs(o, "codes", what, decodePair),
    paths: inputs(o, "paths", what, decodeRoute),
  };
}

/** decodeMovementFixture reads one record of movements.json. */
export function decodeMovementFixture(
  v: unknown,
  what: string,
): MovementFixture {
  const o = object(v, what, [
    ...headerKeys,
    ...mapKeys,
    "winter",
    "goals",
    "expected",
  ]);
  const f: MovementFixture = {
    ...decodeHeader(o, what),
    ...decodeMap(o, what),
    goals: inputs(o, "goals", what, decodeInt),
    expected: decodeMovementExpected(o["expected"], `${what}.expected`),
  };
  if (boolField(o, "winter", what)) {
    f.winter = true;
  }
  return f;
}

function decodeMoveOrder(o: JSONObject, what: string): MoveOrder {
  const m: MoveOrder = {
    goal: intField(o, "goal", what),
    retarget: intField(o, "retarget", what),
  };
  if (present(o, "stale")) {
    m.stale = inputs(o, "stale", what, decodeInt);
  }
  return m;
}

function decodeRecord(v: unknown, what: string): FixtureRecord {
  const o = object(v ?? {}, what, ["index", "bytes"]);
  return {
    index: intField(o, "index", what),
    bytes: list(o, "bytes", what, decodeInt),
  };
}

// decodeOccupants reads a Go [2][][2]int: null and missing elements are nil.
function decodeOccupants(
  v: unknown,
  what: string,
): [Pair[] | null, Pair[] | null] {
  if (v === null || v === undefined) {
    return [null, null];
  }
  const a = array(v, what);
  if (a.length > 2) {
    throw new Error(`${what}: more than two elements`);
  }
  const side = (i: number): Pair[] | null => {
    const s = a[i];
    return s === null || s === undefined
      ? null
      : array(s, `${what}[${i}]`).map((p, k) =>
          decodePair(p, `${what}[${i}][${k}]`),
        );
  };
  return [side(0), side(1)];
}

const stepExpectedKeys = [
  "order",
  "start",
  "messages",
  "cells",
  "live",
  "occupants",
  "bits",
  "orders",
  "count",
  "container",
];

function decodeStepFields(o: JSONObject, what: string): StepExpected {
  return {
    order: intField(o, "order", what),
    start: intField(o, "start", what),
    messages: list(o, "messages", what, decodePair),
    cells: list(o, "cells", what, decodeInt),
    live: list(o, "live", what, decodeRecord),
    occupants: decodeOccupants(o["occupants"], `${what}.occupants`),
    bits: list(o, "bits", what, decodePair),
    orders: list(o, "orders", what, decodeInt),
    count: intField(o, "count", what),
    container: list(o, "container", what, decodeInt),
  };
}

/** decodeStepFixture reads one record of steps.json. */
export function decodeStepFixture(v: unknown, what: string): StepFixture {
  const o = object(v, what, [...headerKeys, ...mapKeys, "order", "expected"]);
  const x = `${what}.expected`;
  return {
    ...decodeHeader(o, what),
    ...decodeMap(o, what),
    order: decodeMoveOrder(
      object(o["order"] ?? {}, `${what}.order`, ["goal", "retarget", "stale"]),
      `${what}.order`,
    ),
    expected: decodeStepFields(
      object(o["expected"] ?? {}, x, stepExpectedKeys),
      x,
    ),
  };
}

function decodeHold(v: unknown, what: string): FixtureHold {
  const o = object(v ?? {}, what, [
    "index",
    "slots",
    "cells",
    "flags",
    "occupant",
  ]);
  return {
    index: intField(o, "index", what),
    slots: decodeFixed(o["slots"], 14, `${what}.slots`),
    cells: decodeFixed(o["cells"], 4, `${what}.cells`),
    flags: intField(o, "flags", what),
    occupant: intField(o, "occupant", what),
  };
}

function decodeExitOrder(v: unknown, what: string): ExitOrder {
  const o = object(v ?? {}, what, [
    "hold",
    "slot",
    "goal",
    "retarget",
    "stale",
  ]);
  return {
    hold: intField(o, "hold", what),
    slot: intField(o, "slot", what),
    ...decodeMoveOrder(o, what),
  };
}

function decodeContainerExpected(v: unknown, what: string): ContainerExpected {
  const o = object(v ?? {}, what, [...stepExpectedKeys, "holds", "remaining"]);
  return {
    ...decodeStepFields(o, what),
    holds: list(o, "holds", what, decodeRecord),
    remaining: list(o, "remaining", what, decodePair),
  };
}

/** decodeContainerFixture reads one record of containers.json. */
export function decodeContainerFixture(
  v: unknown,
  what: string,
): ContainerFixture {
  const o = object(v, what, [
    ...headerKeys,
    ...mapKeys,
    "held",
    "records",
    "order",
    "expected",
  ]);
  return {
    ...decodeHeader(o, what),
    ...decodeMap(o, what),
    held: inputs(o, "held", what, decodePiece),
    records: inputs(o, "records", what, decodeHold),
    order: decodeExitOrder(o["order"], `${what}.order`),
    expected: decodeContainerExpected(o["expected"], `${what}.expected`),
  };
}

function decodeSighted(v: unknown, what: string): FixtureSighted {
  const o = object(v ?? {}, what, [
    "index",
    "flags4",
    "size",
    "pos",
    "facing",
    "hold",
  ]);
  return {
    index: intField(o, "index", what),
    flags4: intField(o, "flags4", what),
    size: intField(o, "size", what),
    pos: decodePair(o["pos"] ?? null, `${what}.pos`),
    facing: decodePair(o["facing"] ?? null, `${what}.facing`),
    hold: intField(o, "hold", what),
  };
}

// decodeSides reads a Go [2][]int: null and missing elements are nil.
function decodeSides(
  v: unknown,
  what: string,
): [number[] | null, number[] | null] {
  if (v === null || v === undefined) {
    return [null, null];
  }
  const a = array(v, what);
  if (a.length > 2) {
    throw new Error(`${what}: more than two elements`);
  }
  const side = (i: number): number[] | null => {
    const s = a[i];
    return s === null || s === undefined
      ? null
      : array(s, `${what}[${i}]`).map((x, k) =>
          decodeInt(x, `${what}[${i}][${k}]`),
        );
  };
  return [side(0), side(1)];
}

function decodeVisibilityExpected(
  v: unknown,
  what: string,
): VisibilityExpected {
  const o = object(v ?? {}, what, [
    "maps",
    "live",
    "holds",
    "players",
    "orders",
    "count",
    "anim",
    "messages",
  ]);
  return {
    maps: decodeSides(o["maps"], `${what}.maps`),
    live: list(o, "live", what, decodeRecord),
    holds: list(o, "holds", what, decodeRecord),
    players: decodeSides(o["players"], `${what}.players`),
    orders: decodePair(o["orders"] ?? null, `${what}.orders`),
    count: intField(o, "count", what),
    anim: intField(o, "anim", what),
    messages: list(o, "messages", what, decodePair),
  };
}

/** decodeVisibilityFixture reads one record of visibility.json. */
export function decodeVisibilityFixture(
  v: unknown,
  what: string,
): VisibilityFixture {
  const o = object(v, what, [
    ...headerKeys,
    "width",
    "height",
    "maps",
    "units",
    "records",
    "attacker",
    "phase",
    "count",
    "expected",
  ]);
  // The input maps: a nil map is an empty one, as Go ranges it.
  const [m0, m1] = decodeSides(o["maps"], `${what}.maps`);
  return {
    ...decodeHeader(o, what),
    width: intField(o, "width", what),
    height: intField(o, "height", what),
    maps: [m0 ?? [], m1 ?? []],
    units: inputs(o, "units", what, decodeSighted),
    records: inputs(o, "records", what, decodeHold),
    attacker: intField(o, "attacker", what),
    phase: decodePair(o["phase"] ?? null, `${what}.phase`),
    count: intField(o, "count", what),
    expected: decodeVisibilityExpected(o["expected"], `${what}.expected`),
  };
}
