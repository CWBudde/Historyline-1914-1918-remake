// The pincer and blocking support of the order handler. L1 holds the parts the fixtures' error names and the
// distance tests read (ErrUndefined, the neighbour steps); the support rules follow.

import { GoError, at } from "./goruntime.ts";
import { div, rem } from "./int.ts";

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
