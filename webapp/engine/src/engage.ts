// One attack order through the order handler (engage.go): the battlefield, the defender, support, casualties
// and the experience window.

import { type Result, casualties, firesBack } from "./casualties.ts";
import { distance } from "./distance.ts";
import { at } from "./goruntime.ts";
import { type Profile, type Type, alternate, tablesOf } from "./profile.ts";
import { neighbours, support } from "./support.ts";

/** Unit is a live unit on the battlefield. Go copies it by value: copy before changing one. */
export interface Unit {
  /** Index into the profile's types. */
  type: number;
  cell: number;
  size: number;
  alt: number;
  exp: number;
  /** Live flag word +4: side bits 1/2 and target classes. */
  flags4: number;
  /** Live flag word +6: bit 0x20 no award. */
  flags6: number;
}

/** zeroUnit is Go's Unit{}, what a map read of a missing index gives. */
const zeroUnit: Readonly<Unit> = Object.freeze({
  type: 0,
  cell: 0,
  size: 0,
  alt: 0,
  exp: 0,
  flags4: 0,
  flags6: 0,
});

/**
 * unitOf is Go's units[i]: the unit at index i, or the zero unit when there is none. The result is shared with
 * the map (or frozen): read it, never change it.
 */
export function unitOf(
  units: ReadonlyMap<number, Unit>,
  i: number,
): Readonly<Unit> {
  return units.get(i) ?? zeroUnit;
}

/**
 * Battlefield is a DOS map: width × height cells numbered row by row, odd columns half a cell lower. terrain
 * gives a cell's combat class (0-8; missing cells have defaultClass). mapByte0 is the low byte of cell 0 (its
 * terrain index), which the support routine reads for off-map neighbours. Go copies the struct but shares its
 * maps.
 */
export interface Battlefield {
  width: number;
  height: number;
  units: Map<number, Unit>;
  terrain: Map<number, number>;
  mapByte0: number;
}

/** defaultClass is the terrain class of cells without an entry; its terms are 0/0. */
export const defaultClass = 7;

/** terrainClass is a cell's combat class (Go's Battlefield.class). */
export function terrainClass(b: Battlefield, cell: number): number {
  return b.terrain.get(cell) ?? defaultClass;
}

/**
 * unitAt returns the index of the unit in a cell, undefined if there is none (Go's Battlefield.at). Go ranges
 * over a map; two units in one cell never arise, so the order does not matter.
 */
export function unitAt(b: Battlefield, cell: number): number | undefined {
  for (const [idx, u] of b.units) {
    if (u.cell === cell) {
      return idx;
    }
  }
  return undefined;
}

/** Outcome is the state after one attack order. */
export interface Outcome extends Result {
  /** False: no unit in the target cell, the order is skipped. */
  battle: boolean;
  defender: number;
  /** The defender stood next to the attacker. */
  returnFire: boolean;
  /** And fired back: its type can (firesBack). */
  returned: boolean;
  pincer: number;
  blocking: number;
  attackerExp: number;
  defenderExp: number;
  /** Units destroyed, attacker first (Go's nil is empty). */
  destroyed: number[];
}

/** zeroOutcome is Go's Outcome{}: a skipped order. */
export function zeroOutcome(): Outcome {
  return {
    battle: false,
    defender: 0,
    returnFire: false,
    returned: false,
    pincer: 0,
    blocking: 0,
    attackerSize: 0,
    attackerAlt: 0,
    defenderSize: 0,
    defenderAlt: 0,
    attackerExp: 0,
    defenderExp: 0,
    destroyed: [],
  };
}

export const expCap = 6;
export const noAwardFlag = 0x20;

/**
 * award is the experience window after a battle (02dc:0908): a side that lost strength gives its opponent 1
 * unless its live +6 word has bit 0x20, a destroyed side gives 1 more; experience stops at 6.
 */
export function award(
  expA: number,
  expD: number,
  lostA: boolean,
  lostD: boolean,
  deadA: boolean,
  deadD: boolean,
  flagsA: number,
  flagsD: number,
): [number, number] {
  let gainA = 0;
  let gainD = 0;
  if (deadA) {
    gainD++;
  }
  if (lostA && (flagsA & noAwardFlag) === 0) {
    gainD++;
  }
  if (deadD) {
    gainA++;
  }
  if (lostD && (flagsD & noAwardFlag) === 0) {
    gainA++;
  }
  return [Math.min(expCap, expA + gainA), Math.min(expCap, expD + gainD)];
}

/** strength is the strength the experience window compares. */
export function strength(t: Type, size: number, alt: number): number {
  return alternate(t) ? alt : size;
}

/**
 * engage resolves the attack order of unit attacker on the target cell as the order handler (02dc:0004 in the
 * demo) does (Go's Profile.Engage): the defender is the unit in the cell, it may fire back when it stands next
 * to the attacker, support rates the neighbours, distance and the two cells' terrain classes feed casualties,
 * and the experience window compares the strengths before and after. An attacker that is not on the
 * battlefield fights as Go's zero unit. It throws ErrUndefined or ErrDivideError where Go returns them (Go's
 * partial outcome beside the error is dropped; no caller reads it).
 */
export function engage(
  p: Profile,
  b: Battlefield,
  attacker: number,
  target: number,
): Outcome {
  const def = unitAt(b, target);
  if (def === undefined) {
    return zeroOutcome();
  }
  const a = unitOf(b.units, attacker);
  const d = unitOf(b.units, def);
  const types = tablesOf(p).types;
  const ta = at(types, a.type);
  const td = at(types, d.type);
  const returnFire = neighbours(a.cell, b.width, b.height).includes(d.cell);
  const [pincer, blocking] = support(p, b, attacker, def);
  const r = casualties(
    p,
    { type: ta, size: a.size, alt: a.alt, exp: a.exp, flags: a.flags4 },
    { type: td, size: d.size, alt: d.alt, exp: d.exp, flags: d.flags4 },
    {
      distance: distance(p, b.width, a.cell, d.cell),
      aTerrain: terrainClass(b, a.cell),
      dTerrain: terrainClass(b, d.cell),
      pincer,
      blocking,
      returnFire,
    },
  );

  const beforeA = strength(ta, a.size, a.alt);
  const afterA = strength(ta, r.attackerSize, r.attackerAlt);
  const beforeD = strength(td, d.size, d.alt);
  const afterD = strength(td, r.defenderSize, r.defenderAlt);
  const [attackerExp, defenderExp] = award(
    a.exp,
    d.exp,
    ((beforeA - afterA) & 0xff) !== 0,
    ((beforeD - afterD) & 0xff) !== 0,
    afterA === 0,
    afterD === 0,
    a.flags6,
    d.flags6,
  );
  const destroyed: number[] = [];
  if (afterA === 0) {
    destroyed.push(attacker);
  }
  if (afterD === 0) {
    destroyed.push(def);
  }
  return {
    battle: true,
    defender: def,
    returnFire,
    returned: firesBack(ta, td, returnFire),
    pincer,
    blocking,
    attackerSize: r.attackerSize,
    attackerAlt: r.attackerAlt,
    defenderSize: r.defenderSize,
    defenderAlt: r.defenderAlt,
    attackerExp,
    defenderExp,
    destroyed,
  };
}
