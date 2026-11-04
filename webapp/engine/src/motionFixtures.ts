// The movement, step and container fixtures (internal/dosrules/testdata/, numbers only) use the movement tables
// of the build's profile (Tables.move); the tests replay them (motion_fixtures.go). The visibility fixtures
// follow with visibility.ts.
//
// The types mirror the JSON: omitempty fields are optional and absent when zero, and an expectation's null is
// Go's nil slice, so a decoded record and an adapter's result compare with toStrictEqual as Go's
// reflect.DeepEqual does. Input lists (a fixture's map, units and records) read null as empty, as Go ranges a
// nil slice.

import {
  type ExitOrder,
  type ExitRun,
  type Hold,
  holdBytes,
  holdRecords,
} from "./container.ts";
import { GoError, at } from "./goruntime.ts";
import { div, rem } from "./int.ts";
import { type Movement, type Piece, newMovement } from "./move.ts";
import type { FixtureHeader } from "./fixtures.ts";
import type { MoveOrder, StepRun } from "./record.ts";

/** Pair is a Go [2]int. */
export type Pair = [number, number];

/** FixturePiece is a live unit of a movement fixture. */
export interface FixturePiece extends Piece {
  index: number;
}

function pieceOf(p: Readonly<Piece>): Piece {
  return {
    type: p.type,
    cell: p.cell,
    flags4: p.flags4,
    flags6: p.flags6,
    hold: p.hold,
  };
}

/** fixturePieces lists the units of a map in index order. */
export function fixturePieces(m: ReadonlyMap<number, Piece>): FixturePiece[] {
  return [...m]
    .sort(([a], [b]) => a - b)
    .map(([i, p]) => ({ index: i, ...pieceOf(p) }));
}

/** FixtureMap is the map and the moving unit of a movement or step fixture. */
export interface FixtureMap {
  width: number;
  height: number;
  /** Terrain record of every cell not in cells. */
  fill: number;
  /** [cell, terrain record]. */
  cells: Pair[];
  units: FixturePiece[];
  unseen: FixturePiece[];
  unit: number;
  side: number;
  budget: number;
  block: number;
}

/** mapOf is the fixture form of a movement query; fill is its most common terrain record. */
export function mapOf(g: Movement): FixtureMap {
  const f: FixtureMap = {
    width: g.width,
    height: g.height,
    fill: 0,
    cells: [],
    units: fixturePieces(g.units),
    unseen: fixturePieces(g.unseen),
    unit: g.unit,
    side: g.side,
    budget: g.budget,
    block: g.block,
  };
  const counts = new Map<number, number>();
  for (const t of g.terrain) {
    counts.set(t, (counts.get(t) ?? 0) + 1);
  }
  for (const [t, n] of counts) {
    const best = counts.get(f.fill) ?? 0;
    if (n > best || (n === best && t < f.fill)) {
      f.fill = t;
    }
  }
  for (const [cell, t] of g.terrain.entries()) {
    if (t !== f.fill) {
      f.cells.push([cell, t]);
    }
  }
  return f;
}

/** fixtureMovement rebuilds the map of a fixture (Go's FixtureMap.Movement). */
export function fixtureMovement(f: FixtureMap): Movement {
  const g = newMovement({
    width: f.width,
    height: f.height,
    terrain: Array.from({ length: f.width * f.height }, () => f.fill),
    unit: f.unit,
    side: f.side,
    budget: f.budget,
    block: f.block,
  });
  for (const [cell, t] of f.cells) {
    g.terrain[cell] = t;
  }
  for (const p of f.units) {
    g.units.set(p.index, pieceOf(p));
  }
  for (const p of f.unseen) {
    g.unseen.set(p.index, pieceOf(p));
  }
  return g;
}

/** MovementExpected is what the executed movement queries gave. */
export interface MovementExpected {
  /**
   * The cost buffer after the reach search, one signed byte per cell in cell order: the budget left on reaching
   * it, or below 0 the negated entry cost, -80 closed, -102/-101 a zone of control of side 0/1 that was not
   * reached, -1 untouched.
   */
  remaining: number[];
  count: number;
  /**
   * Null where the end-cell check reaches its container branch (an own transporter or a building cell), which
   * the container fixtures record.
   */
  destinations: number[] | null;
  /** [cell, end code] for every reached cell outside the container branch. */
  codes: Pair[];
  /** Per goal, the cells from the goal back to the start, or null where the search failed. */
  paths: (number[] | null)[];
}

/** MovementFixture is a recorded set of movement queries on one map. */
export interface MovementFixture extends FixtureHeader, FixtureMap {
  /** The query reads the winter terrain table (moveSeason); absent when false. */
  winter?: boolean;
  goals: number[];
  expected: MovementExpected;
}

/** movementQuery rebuilds the query of a fixture (Go's MovementFixture.Query). */
export function movementQuery(f: MovementFixture): Movement {
  const g = fixtureMovement(f);
  g.goals = f.goals;
  return g;
}

/** FixtureRecord is a record after the run, its bytes in order. */
export interface FixtureRecord {
  index: number;
  bytes: number[] | null;
}

/** ints is a byte array as numbers (Go's ints, never nil). */
export function ints(b: Uint8Array): number[] {
  return Array.from(b);
}

/** StepExpected is what an executed move order gave. Null is Go's nil slice. */
export interface StepExpected {
  /** AX of the order code: 1, 9 for a two-cell ship, 0x12 refused. */
  order: number;
  /** AX of the move start, 0 where the order was refused. */
  start: number;
  messages: Pair[] | null;
  /** The mover's cell after each step. */
  cells: number[] | null;
  live: FixtureRecord[] | null;
  /**
   * [cell, unit] for every occupied cell of the side's map and of the other side's map afterwards; the terrain
   * bytes stay.
   */
  occupants: [Pair[] | null, Pair[] | null];
  /** [row*64+column, value] of every non-zero byte. */
  bits: Pair[] | null;
  orders: number[] | null;
  count: number;
  /** Container record 0. */
  container: number[] | null;
}

/**
 * stepExpectedOf is the recorded form of an outcome; it throws where a map's terrain bytes changed, which the
 * fixture does not record.
 */
export function stepExpectedOf(g: Movement, r: StepRun): StepExpected {
  const live = [...r.live.keys()]
    .sort((a, b) => a - b)
    .map((idx) => ({
      index: idx,
      bytes: ints(r.live.get(idx) ?? new Uint8Array(0)),
    }));
  const x: StepExpected = {
    order: r.order,
    start: r.start,
    messages: r.messages,
    cells: r.cells,
    // Go appends to a nil slice: no record leaves it nil.
    live: live.length === 0 ? null : live,
    occupants: [[], []],
    bits: [],
    orders: r.orders,
    count: r.count,
    container: ints(r.container),
  };
  for (const [k, mp] of r.maps.entries()) {
    const occ: Pair[] = [];
    for (let c = 0; c < g.width * g.height; c++) {
      if (at(mp, 2 * c) !== at(g.terrain, c)) {
        throw new GoError(`map ${k}: terrain of cell ${c} changed`);
      }
      if (at(mp, 2 * c + 1) !== 0xff) {
        occ.push([c, at(mp, 2 * c + 1)]);
      }
    }
    x.occupants[k] = occ;
  }
  const bits: Pair[] = [];
  for (const [i, v] of r.bits.entries()) {
    if (v !== 0) {
      bits.push([i, v]);
    }
  }
  x.bits = bits;
  return x;
}

/** StepFixture is a recorded move order. */
export interface StepFixture extends FixtureHeader, FixtureMap {
  order: MoveOrder;
  expected: StepExpected;
}

/** FixtureHold is a container record of a fixture. */
export interface FixtureHold extends Hold {
  index: number;
}

/** fixtureHolds lists container records in index order. */
export function fixtureHolds(m: ReadonlyMap<number, Hold>): FixtureHold[] {
  return [...m]
    .sort(([a], [b]) => a - b)
    .map(([i, h]) => ({ index: i, ...copyHold(h) }));
}

function copyHold(h: Hold): Hold {
  return {
    slots: [...h.slots],
    cells: [...h.cells],
    flags: h.flags,
    occupant: h.occupant,
  };
}

/**
 * ContainerExpected is what an executed container case gave: the step outcome, the container records that
 * changed (all 0x1c bytes) and, for an exit, [cell, budget left] of every cell its reach search reached.
 */
export interface ContainerExpected extends StepExpected {
  holds: FixtureRecord[] | null;
  remaining: Pair[] | null;
}

/** containerExpectedOf is the recorded form of an outcome. */
export function containerExpectedOf(
  g: Movement,
  o: ExitOrder,
  r: ExitRun,
): ContainerExpected {
  const remaining: Pair[] = [];
  if (o.hold >= 0) {
    for (let c = 0; c < g.width * g.height; c++) {
      const v = at(r.remaining, div(c, g.width) * 64 + rem(c, g.width));
      if (v >= 0) {
        remaining.push([c, v]);
      }
    }
  }
  return {
    ...stepExpectedOf(g, r),
    holds: changedHolds(holdBytes(g), r.holds),
    remaining,
  };
}

function bytesEqual(a: Uint8Array, b: Uint8Array): boolean {
  return a.length === b.length && a.every((v, i) => v === b[i]);
}

/** changedHolds are the records of after that differ from before (Go's bytes.Equal; a missing one is nil). */
export function changedHolds(
  before: ReadonlyMap<number, Uint8Array>,
  after: ReadonlyMap<number, Uint8Array>,
): FixtureRecord[] {
  const out: FixtureRecord[] = [];
  for (let i = 0; i < holdRecords; i++) {
    const b = before.get(i) ?? new Uint8Array(0);
    const a = after.get(i) ?? new Uint8Array(0);
    if (!bytesEqual(b, a)) {
      out.push({ index: i, bytes: ints(a) });
    }
  }
  return out;
}

/** ContainerFixture is a recorded load, unload or exit. */
export interface ContainerFixture extends FixtureHeader, FixtureMap {
  held: FixturePiece[];
  records: FixtureHold[];
  order: ExitOrder;
  expected: ContainerExpected;
}

/** containerMovement rebuilds the map, the held units and the records of a fixture (Go's ContainerFixture.Movement). */
export function containerMovement(f: ContainerFixture): Movement {
  const g = fixtureMovement(f);
  for (const p of f.held) {
    g.held.set(p.index, pieceOf(p));
  }
  for (const h of f.records) {
    g.holds.set(h.index, copyHold(h));
  }
  return g;
}
