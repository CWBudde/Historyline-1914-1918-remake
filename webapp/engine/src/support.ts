// The pincer and blocking support of the order handler (support.go).

import { classBits, longRange } from "./casualties.ts";
import { type Battlefield, unitAt, unitOf } from "./engage.ts";
import { GoError, at } from "./goruntime.ts";
import { div, rem } from "./int.ts";
import { type Profile, tablesOf } from "./profile.ts";

/**
 * ErrUndefined is the case where the original reads an uninitialised stack byte: a defender at the map edge,
 * not adjacent to its attacker, whose off-map neighbour reads map byte 0 as the attacker's index.
 */
export class ErrUndefined extends GoError {
  constructor() {
    super("undefined: direction from an uninitialised stack byte");
  }
}

/**
 * Neighbour steps (column, row) in the DOS order 0 up, 1-2 right, 3 down, 4-5 left, for even and odd columns;
 * odd columns sit half a cell lower.
 */
export const steps: readonly (readonly (readonly [number, number])[])[] = [
  [
    [0, -1],
    [1, -1],
    [1, 0],
    [0, 1],
    [-1, 0],
    [-1, -1],
  ],
  [
    [0, -1],
    [1, 0],
    [1, 1],
    [0, 1],
    [-1, 1],
    [-1, 0],
  ],
];

/** neighbours returns the six neighbours of a cell (0b06:0157), -1 off the map. Width and height are at least 2. */
export function neighbours(
  cell: number,
  width: number,
  height: number,
): number[] {
  const x = rem(cell, width);
  const y = div(cell, width);
  const n: number[] = [];
  for (const [sx, sy] of at(steps, x & 1)) {
    const nx = x + sx;
    const ny = y + sy;
    n.push(
      nx >= 0 && nx < width && ny >= 0 && ny < height ? ny * width + nx : -1,
    );
  }
  return n;
}

/**
 * friendly is the side test (0ca6:020a) on two live flag words: side bits 1 and 2 pair with themselves, a unit
 * without either pairs with non-side-1.
 */
export function friendly(unit: number, own: number): boolean {
  if ((unit & 1) !== 0) {
    return (own & 1) !== 0;
  }
  if ((unit & 2) !== 0) {
    return (own & 2) !== 0;
  }
  return (own & 1) === 0;
}

/** Pincer weights by the step from the attacker's direction around the defender. */
export const pincerWeights: readonly number[] = [0, 25, 50, 100, 50, 25];

/** The blocking percentage per eligible unit and its cap. */
export const blockStep = 80;
export const blockCap = 150;

/**
 * supporters marks the neighbours of anchor that would support against opponent and returns the direction in
 * which opponent stands (-1 if not). A supporter's type has one of anchor's live target classes (or anchor's
 * live +6 bit 2) in its target flags, is not long range, and stands on the opponent's side.
 */
function supporters(
  p: Profile,
  b: Battlefield,
  anchor: number,
  opponent: number,
): [slots: boolean[], dir: number] {
  const u = unitOf(b.units, anchor);
  const classes = u.flags4 & classBits;
  const bit2 = u.flags6 & 2;
  const slots = [false, false, false, false, false, false];
  let dir = -1;
  for (const [k, c] of neighbours(u.cell, b.width, b.height).entries()) {
    if (c < 0) {
      continue;
    }
    const idx = unitAt(b, c);
    if (idx !== undefined) {
      const n = unitOf(b.units, idx);
      const t = at(tablesOf(p).types, n.type).targetFlags;
      slots[k] =
        ((classes & t) !== 0 || (bit2 & t) !== 0) &&
        (t & longRange) === 0 &&
        friendly(n.flags4, unitOf(b.units, opponent).flags4);
      if (idx === opponent) {
        dir = k;
      }
    }
  }
  return [slots, dir];
}

/**
 * support rates the neighbours of a battle (02dc:262a; Go's Profile.Support): units on the attacker's side
 * around the defender add the pincer percentage, units on the defender's side beside the defender's direction
 * around the attacker add the blocking percentage. Only adjacent battles have support. The pincer is a byte
 * (five supporters give 250, the 300 clamp is dead code); blocking adds 80 per unit and stops at 150. It throws
 * ErrUndefined where the original reads an uninitialised stack byte.
 */
export function support(
  p: Profile,
  b: Battlefield,
  attacker: number,
  defender: number,
): [pincer: number, blocking: number] {
  let adjacent = false;
  for (const c of neighbours(
    unitOf(b.units, defender).cell,
    b.width,
    b.height,
  )) {
    // an off-map neighbour is read at map byte 0
    let v = b.mapByte0;
    if (c >= 0) {
      v = unitAt(b, c) ?? 0xff;
    }
    adjacent = adjacent || v === attacker;
  }
  if (!adjacent) {
    return [0, 0];
  }
  let [slots, dir] = supporters(p, b, defender, attacker);
  if (dir < 0) {
    throw new ErrUndefined();
  }
  let pincer = 0;
  for (let k = 1; k < 6; k++) {
    if (at(slots, (dir + k) % 6)) {
      pincer = (pincer + at(pincerWeights, k)) & 0xff;
    }
  }
  [slots, dir] = supporters(p, b, attacker, defender);
  if (dir < 0) {
    throw new ErrUndefined();
  }
  let blocking = 0;
  for (const k of [1, 5]) {
    if (at(slots, (dir + k) % 6)) {
      blocking += blockStep;
    }
  }
  return [pincer, Math.min(blocking, blockCap)];
}
