// The movement, route, container and visibility rules work on compact records: 26-byte live unit records,
// maps of two bytes per cell, the 0x1104-byte cost and reach-bit arrays and 0x1c-byte container records.
// board.ts reads them into typed values.
//
// Go's []byte is a Uint8Array here (shared like a Go slice, wrapping like a byte), []int8 an Int8Array, and
// map[int]X a Map<number, X>.

import { at } from "./goruntime.ts";

/** ContainerLen is the size of a container or building record. */
export const containerLen = 0x1c;

/** word is the little-endian word of rec at off. */
export function word(rec: Uint8Array, off: number): number {
  return at(rec, off) | (at(rec, off + 1) << 8);
}

/**
 * MoveOrder is a move order of the moving unit: the destination cell, the cell the order is changed to between
 * the order code and the start (-1: none), and cells whose reach bit for the side is left from an earlier
 * query (the reach search ORs its bits into the array; its callers clear it). Stale is absent where Go's is
 * nil (JSON omitempty).
 */
export interface MoveOrder {
  goal: number;
  retarget: number;
  stale?: number[];
}

/** StepRun is the executed outcome of one move order. Go's nil slices are empty arrays here. */
export interface StepRun {
  /** AX of the order code (1, 0x12 refused, |8 two cells). */
  order: number;
  /** AX of the move start. */
  start: number;
  /** Status messages shown: code, side. */
  messages: [number, number][];
  /** The mover's cell as its side sees it, after each step. */
  cells: number[];
  /** The live record of every placed unit afterwards. */
  live: Map<number, Uint8Array>;
  /** The side's and the other side's map afterwards. */
  maps: [Uint8Array, Uint8Array];
  /** The bit array afterwards. */
  bits: Uint8Array;
  /** The side's order list afterwards. */
  orders: number[];
  /** The route count afterwards. */
  count: number;
  /** The container records afterwards. */
  holds: Map<number, Uint8Array>;
  /** Container record 0 afterwards. */
  container: Uint8Array;
}

/** MaxSteps bounds the steps of one move (a route has at most 790 cells). */
export const maxSteps = 0x400;
