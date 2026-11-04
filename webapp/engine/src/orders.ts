// A side's attack list through the order handler's loop (orders.go).

import {
  type Battlefield,
  type Outcome,
  type Unit,
  engage,
  unitOf,
} from "./engage.ts";
import { GoError } from "./goruntime.ts";
import type { Profile } from "./profile.ts";

/** Order is one attack order: the unit attacker against the cell target. */
export interface Order {
  attacker: number;
  target: number;
}

/** Resolution is what the order handler leaves after a side's attack list. */
export interface Resolution {
  /** Per order in list order; battle false: skipped. */
  orders: Outcome[];
  /** Every unit afterwards, destroyed ones included. */
  units: Map<number, Unit>;
  /** The units the destroy helper took off the map, ascending (Go's nil is empty). */
  gone: number[];
  /** The occupied cells of the attacking side's map afterwards: cell to unit. */
  cells: Map<number, number>;
}

/**
 * OrdersError is the error that stopped resolveOrders (its cause, whose message it keeps) with the resolution
 * as far as it got, which Go returns beside the error: the orders before the failing one and the units they
 * left, gone and cells empty.
 */
export class OrdersError extends GoError {
  readonly resolution: Resolution;

  constructor(err: GoError, resolution: Resolution) {
    super(err.message, { cause: err });
    this.resolution = resolution;
  }
}

/**
 * resolveOrders resolves a side's attack list on the battlefield as that side sees it, as the handler's loop
 * (02dc:0117–0d55) does (Go's Profile.ResolveOrders): the orders run in list order, each one as engage on the
 * state the earlier ones left. There is no alive check: an order whose target cell holds no unit is skipped,
 * any other is fought, so a target that survives is fought again. The destroy helper (02dc:2083), called for
 * every unit a battle destroys, takes it off the attacking side's map at once, so a destroyed defender is no
 * longer a target and a destroyed attacker no longer supports. The unit keeps its record (its cell and
 * strength), and engage still runs for an order of a destroyed attacker (as Go's zero unit, since it is off
 * the map); the attack order never lets a side target its own units, so that order cannot arise in the game.
 *
 * A two-cell ship fights as one unit, unlike in the original, where each half keeps its own strength: a battle
 * that hits either half leaves the other half the same strength and experience, and destroys both together
 * (destroyed lists the other half after the one fought).
 *
 * b is left as it is: each order fights on a copy of it with its own units map (its terrain map shared, as Go
 * copies the struct), and the resolution holds copies of the units. An error engage throws comes back as an
 * OrdersError carrying the partial resolution.
 */
export function resolveOrders(
  p: Profile,
  b: Battlefield,
  orders: readonly Order[],
): Resolution {
  const r: Resolution = {
    orders: [],
    units: new Map(),
    gone: [],
    cells: new Map(),
  };
  for (const [i, u] of b.units) {
    r.units.set(i, { ...u });
  }
  const gone = new Set<number>();
  for (const o of orders) {
    const view: Battlefield = { ...b, units: new Map() };
    for (const [i, u] of r.units) {
      if (!gone.has(i)) {
        view.units.set(i, u);
      }
    }
    let out: Outcome;
    try {
      out = engage(p, view, o.attacker, o.target);
    } catch (e) {
      if (e instanceof GoError) {
        throw new OrdersError(e, r);
      }
      throw e;
    }
    r.orders.push(out);
    if (!out.battle) {
      continue;
    }
    // Go reads both records before writing either (they are value copies).
    const a: Unit = {
      ...unitOf(r.units, o.attacker),
      size: out.attackerSize,
      alt: out.attackerAlt,
      exp: out.attackerExp,
    };
    const d: Unit = {
      ...unitOf(r.units, out.defender),
      size: out.defenderSize,
      alt: out.defenderAlt,
      exp: out.defenderExp,
    };
    r.units.set(o.attacker, a);
    r.units.set(out.defender, d);
    for (const k of [o.attacker, out.defender]) {
      const q = shipPartner(r.units, k);
      if (q !== undefined) {
        r.units.set(q, whole(unitOf(r.units, q), unitOf(r.units, k)));
        if (out.destroyed.includes(k) && !out.destroyed.includes(q)) {
          out.destroyed.push(q);
        }
      }
    }
    for (const k of out.destroyed) {
      gone.add(k);
    }
  }
  for (const [i, u] of r.units) {
    if (gone.has(i)) {
      r.gone.push(i);
    } else {
      r.cells.set(u.cell, i);
    }
  }
  r.gone.sort((x, y) => x - y);
  return r;
}

/**
 * shipPartner is the other half of a two-cell ship among units: the next record from the head (+4 bit 0x80),
 * the previous one from the rear. It is undefined for any other unit and for a half whose partner is not there.
 */
export function shipPartner(
  units: ReadonlyMap<number, Unit>,
  i: number,
): number | undefined {
  const f = unitOf(units, i).flags4;
  if ((f & 0x40) === 0) {
    return undefined;
  }
  const p = (f & 0x80) !== 0 ? i + 1 : i - 1;
  const q = units.get(p);
  if (
    q === undefined ||
    (q.flags4 & 0x40) === 0 ||
    (q.flags4 & 0x80) === (f & 0x80)
  ) {
    return undefined;
  }
  return p;
}

/** whole is half of a two-cell ship with the strength and experience of its other half. */
export function whole(half: Readonly<Unit>, other: Readonly<Unit>): Unit {
  return { ...half, size: other.size, alt: other.alt, exp: other.exp };
}
