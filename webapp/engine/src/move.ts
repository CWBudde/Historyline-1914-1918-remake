// The movement queries of segment 0b06 (demo addresses) work on a byte per cell in arrays of 64 cells per row
// (index row*64+column, 0x1104 bytes); the map itself is two bytes per cell in rows of Width cells (terrain
// record, unit index, 0xf1 and up empty). Column and row are cell % Width and cell / Width; odd columns sit
// half a cell lower.
//
// Go's []int8 cost buffer is an Int8Array and its []byte arrays Uint8Arrays, which wrap on a write as Go's
// conversions do; at and setAt keep Go's bounds checks.

import { type Hold, endCheckOn, moveEndCheck } from "./container.ts";
import { hexDistance } from "./distance.ts";
import { at, setAt } from "./goruntime.ts";
import { div, int8, int16, rem, uint8 } from "./int.ts";
import { newMoveState } from "./route.ts";
import { ErrUndefined } from "./support.ts";

/** BufLen is the size of the cost buffer and of the reach-bit array. */
export const bufLen = 0x1104;

// Cost buffer markers (signed bytes).
/** 0xb0: terrain, a neutral or enemy unit, or another unit for a ship. */
const costBlocked = -80;
/** 0x9a: next to an enemy of side 0. */
const zocSide0 = -102;
/** 0x9b: next to an enemy of side 1. */
const zocSide1 = -101;

/** MoveType holds the fields of a type record the movement queries read. */
export interface MoveType {
  /** Byte 0, copied into live byte 0 by the constructor. */
  budget: number;
  /** Byte 1, the displayed speed (not read here). */
  speed: number;
  /** Byte 5, ANDed with the terrain mask. */
  mask: number;
  /** Word +0x0d, copied into live +4 (0x40: a ship two cells long). */
  class: number;
  /** Word +0x11, copied into live +6 (0x10: a flat cost of 2). */
  flags: number;
  /** The path score weighs the distance to the goal by +0x15 (signed byte). */
  distanceWeight: number;
  /** The path score weighs the entry cost by +0x16 (signed byte). */
  costWeight: number;
  /** A transporter's capacity (+0x40), which the end-cell check compares (UNIT.DAT byte 64). */
  capacity: number;
  /** Every type's weight (+0x41, UNIT.DAT byte 65). */
  weight: number;
  /** The classes (container flags 0xfc0) a transporter takes, word +0x13. */
  carries: number;
}

/** MoveTerrain is a 6-byte terrain record as the movement queries read it. */
export interface MoveTerrain {
  /** Word 0: 0x4000 closed to air units; 0x400/0x100/0x40 buildings. */
  flags: number;
  /** Byte 2. */
  mask: number;
  /** Byte 3. */
  cost: number;
  /** Byte 4, the cost for units with live +4 bit 0x10. */
  airCost: number;
  /** Byte 5, the combat class the order handler reads (Battlefield.Terrain). */
  class: number;
}

/** MoveTables are a build's type and terrain records. */
export interface MoveTables {
  types: MoveType[];
  /** The summer table (WW1GNDS.DAT). */
  terrains: MoveTerrain[];
  /**
   * The winter table (WW1GNDW.DAT) where the build has one (Go's nil: null). In the full build it differs
   * from the summer table in the cost only.
   */
  winter: MoveTerrain[] | null;
}

/**
 * Season is the tables with the season's terrain records as terrains; null for winter where the build has no
 * winter table. Which season a map is played in is the scenario's. Winter is a new MoveTables sharing the
 * records (Go's &MoveTables{...}); summer is t itself.
 */
export function moveSeason(t: MoveTables, winter: boolean): MoveTables | null {
  if (!winter) {
    return t;
  }
  if (t.winter === null) {
    return null;
  }
  return { types: t.types, terrains: t.winter, winter: t.winter };
}

/**
 * Piece is a live unit on a movement map: flags at +4 (side bits 1/2, class bits from the type's +0x0d), flags
 * at +6, type and cell. Hold is live +0x0a, the record a transporter's cargo sits in (see Hold; JSON omitempty).
 * Go copies it by value: copy before changing one.
 */
export interface Piece {
  type: number;
  cell: number;
  flags4: number;
  flags6: number;
  hold: number;
}

/** zeroPiece is Go's Piece{}, what a map read of a missing index gives. */
const zeroPiece: Readonly<Piece> = Object.freeze({
  type: 0,
  cell: 0,
  flags4: 0,
  flags6: 0,
  hold: 0,
});

/**
 * Movement is one movement query of the live unit Unit for player Side. Terrain holds a terrain record per
 * cell. Units are the units on the side's Owns map; Unseen ones are only on the other side's map, which none of
 * the queries read. Budget is the argument the callers take from live byte 0; Block is the terrain flag mask
 * that closes cells (the player's order code passes 0xffff for live +6 bit 1, else 6 for a side-1 unit and 3
 * for others). Goals are the path targets. Held units sit only in the slots of Holds, the container and
 * building records (index from Move.Containers; a record not given has empty slots and zeros otherwise).
 * Go's nil maps are empty Maps here. Go copies the struct but shares its maps and slices.
 */
export interface Movement {
  width: number;
  height: number;
  terrain: number[];
  units: Map<number, Piece>;
  unseen: Map<number, Piece>;
  held: Map<number, Piece>;
  holds: Map<number, Hold>;
  unit: number;
  side: number;
  budget: number;
  block: number;
  goals: number[];
}

/** newMovement is a Movement literal: the fields not given are Go's zero values (fresh empty maps). */
export function newMovement(fields: Partial<Movement>): Movement {
  return {
    width: 0,
    height: 0,
    terrain: [],
    units: new Map(),
    unseen: new Map(),
    held: new Map(),
    holds: new Map(),
    unit: 0,
    side: 0,
    budget: 0,
    block: 0,
    goals: [],
    ...fields,
  };
}

/**
 * pieceAt is Go's Movement.PieceAt: a placed unit's record as Movement gives it, zeros if none. The result is
 * shared with the map (or frozen): read it, never change it.
 */
export function pieceAt(g: Movement, idx: number): Readonly<Piece> {
  return g.units.get(idx) ?? g.unseen.get(idx) ?? g.held.get(idx) ?? zeroPiece;
}

/** occupants is Go's Movement.Occupants: the cells of the side's map mapped to unit indices. */
export function occupants(g: Movement): Map<number, number> {
  const o = new Map<number, number>();
  for (const [i, u] of g.units) {
    o.set(u.cell, i);
  }
  return o;
}

/** LiveLen is the size of a live unit record (Build.Units). */
export const liveLen = 26;

/**
 * LiveRecords are the live records of Units, Unseen and Held as the queries read them: flags +4 and +6, the
 * type at +8, the hold at +0x0a, the position 2*cell+1 at +0x0b and +0x0d (the cell as each side sees it) and,
 * for the moving unit, the budget in byte 0 (a moving unit not placed panics, as in Go).
 */
export function liveRecords(g: Movement): Map<number, Uint8Array> {
  const recs = new Map<number, Uint8Array>();
  const place = (idx: number, u: Readonly<Piece>): void => {
    const r = new Uint8Array(liveLen);
    putWord(r, 4, u.flags4);
    putWord(r, 6, u.flags6);
    r[8] = uint8(u.type);
    r[0x0a] = uint8(u.hold);
    for (let k = 0; k < 2; k++) {
      putWord(r, 0x0b + 2 * k, 2 * u.cell + 1);
    }
    recs.set(idx, r);
  };
  for (const [idx, u] of g.units) {
    place(idx, u);
  }
  for (const [idx, u] of g.unseen) {
    place(idx, u);
  }
  for (const [idx, u] of g.held) {
    place(idx, u);
  }
  setAt(recs.get(g.unit) ?? new Uint8Array(0), 0, uint8(g.budget));
  return recs;
}

/**
 * MovementMaps are the side's map, holding Units, and the other side's map, holding Units and Unseen: the
 * terrain record and the unit index per cell (0xff empty).
 */
export function movementMaps(
  g: Movement,
): [own: Uint8Array, other: Uint8Array] {
  const own = new Uint8Array(2 * g.width * g.height);
  const other = new Uint8Array(2 * g.width * g.height);
  for (const [c, t] of g.terrain.entries()) {
    setAt(own, 2 * c, uint8(t));
    setAt(own, 2 * c + 1, 0xff);
    setAt(other, 2 * c, uint8(t));
    setAt(other, 2 * c + 1, 0xff);
  }
  for (const [idx, u] of g.units) {
    setAt(own, 2 * u.cell + 1, uint8(idx));
    setAt(other, 2 * u.cell + 1, uint8(idx));
  }
  for (const [idx, u] of g.unseen) {
    setAt(other, 2 * u.cell + 1, uint8(idx));
  }
  return [own, other];
}

/** putWord writes v as the little-endian word of b at off. */
export function putWord(b: Uint8Array, off: number, v: number): void {
  setAt(b, off, uint8(v));
  setAt(b, off + 1, uint8(v >> 8));
}

/**
 * neighbours64 is 0b06:005f with its edge clip 0b06:021f: the six neighbours of (col, row) as indices of 64 per
 * row, -1 off the map.
 */
export function neighbours64(
  col: number,
  row: number,
  width: number,
  height: number,
): number[] {
  const i = row * 64 + col;
  const n = [i - 64, 0, 0, i + 64, 0, 0];
  if ((col & 1) !== 0) {
    [n[1], n[5], n[4], n[2]] = [i + 1, i - 1, i + 63, i + 65];
    if (row <= 0) {
      n[0] = -1;
    } else if (height - 1 <= row) {
      [n[3], n[4], n[2]] = [-1, -1, -1];
    }
  } else {
    [n[2], n[4], n[5], n[1]] = [i + 1, i - 1, i - 65, i - 63];
    if (row <= 0) {
      [n[0], n[5], n[1]] = [-1, -1, -1];
    } else if (height - 1 <= row) {
      n[3] = -1;
    }
  }
  if (col <= 0) {
    [n[4], n[5]] = [-1, -1];
  } else if (width - 1 <= col) {
    [n[1], n[2]] = [-1, -1];
  }
  return n;
}

/** costs is 0b06:09c1 (Go's MoveTables.Costs): the negated entry cost of every cell, or a marker. */
export function costs(t: MoveTables, g: Movement): Int8Array {
  const buf = new Int8Array(bufLen).fill(-1);
  const u = pieceAt(g, g.unit);
  const typ = at(t.types, u.type);
  let mode = 2; // ground: byte 3
  if ((u.flags4 & 0x10) !== 0) {
    mode = 0; // air: byte 4
  }
  if ((u.flags6 & 0x10) !== 0) {
    mode = 1; // a flat cost of 2
  }
  for (let row = 0; row < g.height; row++) {
    for (let col = 0; col < g.width; col++) {
      const tr = at(t.terrains, at(g.terrain, row * g.width + col));
      let v = costBlocked;
      if ((tr.mask & typ.mask) !== 0) {
        switch (mode) {
          case 0:
            v = int8(0 - tr.airCost);
            break;
          case 1:
            v = -2;
            break;
          default:
            v = int8(0 - tr.cost);
        }
      }
      if (g.block !== 0xffff && (tr.flags & g.block) !== 0) {
        v = costBlocked;
      }
      setAt(buf, row * 64 + col, v);
    }
  }
  const occ = occupants(g);
  for (let row = 0; row < g.height; row++) {
    for (let col = 0; col < g.width; col++) {
      const idx = occ.get(row * g.width + col);
      if (idx === undefined) {
        continue;
      }
      const i = row * 64 + col;
      const f = pieceAt(g, idx).flags4;
      let zoc = 0;
      let own = false;
      if ((f & 2) !== 0) {
        setAt(buf, i, costBlocked);
      } else if ((f & 1) !== 0) {
        [own, zoc] = [g.side === 1, zocSide0];
      } else {
        [own, zoc] = [g.side === 0, zocSide1];
      }
      if (zoc !== 0) {
        if (own) {
          if ((f & 0x3000) !== 0) {
            setAt(buf, i, -2);
          }
        } else {
          setAt(buf, i, costBlocked);
          for (const n of neighbours64(col, row, g.width, g.height)) {
            if (n >= 0 && at(buf, n) !== costBlocked) {
              setAt(buf, n, zoc);
            }
          }
        }
      }
      if ((typ.class & 0x40) !== 0 && idx !== g.unit) {
        setAt(buf, i, costBlocked);
      }
    }
  }
  return buf;
}

/**
 * reach is 0b06:0331 (Go's MoveTables.Reach): the cost buffer after the search (a value >= 0 is the budget left
 * on reaching that cell), the side's reach bits it ORs into 0x3b50 and the number of reached cells. The search
 * expands, from the budget down to 1, every cell holding that value inside a window of Budget cells around the
 * start; a neighbour still negative takes the value minus its cost if that is not negative, and a
 * zone-of-control cell takes 0.
 */
export function reach(
  t: MoveTables,
  g: Movement,
): [buf: Int8Array, bits: Uint8Array, count: number] {
  const buf = costs(t, g);
  const start = pieceAt(g, g.unit).cell;
  const col = rem(start, g.width);
  const row = div(start, g.width);
  const half = g.budget >> 1;
  const colLo = Math.max(col - half, 0);
  const rowLo = Math.max(row - half, 0);
  const colHi = Math.min(colLo + g.budget, g.width);
  const rowHi = Math.min(rowLo + g.budget, g.height);
  const zoc = g.side !== 0 ? zocSide1 : zocSide0;
  setAt(buf, row * 64 + col, int8(g.budget));
  for (let level = at(buf, row * 64 + col); level > 0; level--) {
    for (let r = rowLo; r < rowHi; r++) {
      for (let c = colLo; c < colHi; c++) {
        if (at(buf, r * 64 + c) !== level) {
          continue;
        }
        for (const n of neighbours64(c, r, g.width, g.height)) {
          if (n < 0 || at(buf, n) >= 0) {
            continue;
          }
          const v = at(buf, n);
          if (v === zoc) {
            setAt(buf, n, 0);
          } else if (v + level >= 0) {
            setAt(buf, n, v + int8(level));
          }
        }
      }
    }
  }
  const bit = g.side === 1 ? 2 : 1;
  const bits = new Uint8Array(bufLen);
  let count = 0;
  for (const [i, v] of buf.entries()) {
    if (v >= 0) {
      count++;
      bits[i] = at(bits, i) | bit;
    }
  }
  return [buf, bits, count];
}

/** owns is 0ca0:020c: whether live flags +4 belong to player side. */
export function owns(flags: number, side: number): boolean {
  if ((flags & 1) !== 0) {
    return (side & 1) !== 0;
  }
  if ((flags & 2) !== 0) {
    return (side & 2) !== 0;
  }
  return (side & 1) === 0;
}

/**
 * endCode is 10cc:0f74 for the moving unit (moveEndCheck): 0 if it may end on the cell, else the code it
 * shows. It throws ErrUndefined where moveEndCheck does.
 */
export function endCode(t: MoveTables, g: Movement, cell: number): number {
  return moveEndCheck(t, g, cell)[0];
}

/**
 * moveDestinations is 0b06:1383 after reach (Go's MoveTables.Destinations): the cells with the side's reach
 * bit, column by column, whose end code is 0 (at most 790). The end-cell check only reads the memory, so all
 * cells are checked on one. Go's nil (none) is an empty array.
 */
export function moveDestinations(
  t: MoveTables,
  g: Movement,
  bits: Uint8Array,
): number[] {
  const bit = g.side !== 0 ? 2 : 1;
  const s = newMoveState(t, g);
  const out: number[] = [];
  for (let col = 0; col < g.width; col++) {
    for (let row = 0; row < g.height; row++) {
      if ((at(bits, row * 64 + col) & bit) === 0) {
        continue;
      }
      const [code] = endCheckOn(s, row * g.width + col);
      if (code === 0) {
        out.push(row * g.width + col);
        if (out.length >= 0x316) {
          return out;
        }
      }
    }
  }
  return out;
}

/**
 * PathNodes are the 10-byte nodes of the path search (column, row, parent, score and the sorted list links),
 * a field per array, zeros as Go's make([]pathNode, n) leaves them.
 */
class PathNodes {
  readonly col: Int32Array;
  readonly row: Int32Array;
  readonly parent: Int32Array;
  readonly score: Int32Array;
  readonly prev: Int32Array;
  readonly next: Int32Array;

  constructor(n: number) {
    this.col = new Int32Array(n);
    this.row = new Int32Array(n);
    this.parent = new Int32Array(n);
    this.score = new Int32Array(n);
    this.prev = new Int32Array(n);
    this.next = new Int32Array(n);
  }

  /** set is nodes[i] = pathNode{...}. */
  set(
    i: number,
    col: number,
    row: number,
    parent: number,
    score: number,
    prev: number,
    next: number,
  ): void {
    setAt(this.col, i, col);
    setAt(this.row, i, row);
    setAt(this.parent, i, parent);
    setAt(this.score, i, score);
    setAt(this.prev, i, prev);
    setAt(this.next, i, next);
  }
}

/**
 * path is 0b06:0e12 after reach (Go's MoveTables.Path): a best-first search over the cells with the side's
 * reach bit, in a list sorted by score = DistanceWeight x distance to the goal + CostWeight x entry cost (0 at
 * the goal), each cell entered once. It returns the cells from the goal back to the start, or null when the
 * search runs dry or out of its 790 nodes. The score is not summed along the route, so the route need not be
 * the cheapest. The node pool is the buffer the reach search left (buf); the sentinel's column and row stay
 * buffer bytes 20 and 21, so a goal they match, a score past the sentinel's 30000, or expanding the sentinel
 * past the bit array (a search that runs dry) throws ErrUndefined. The order code only asks for destinations.
 */
export function path(
  t: MoveTables,
  g: Movement,
  buf: Int8Array,
  bits: Uint8Array,
  goal: number,
): number[] | null {
  const w = g.width;
  const u = pieceAt(g, g.unit);
  const typ = at(t.types, u.type);
  let reachBit = 1;
  let visit = 0x40;
  if (g.side !== 0) {
    [reachBit, visit] = [2, 0x80];
  }
  bits = bits.slice();
  const gcol = rem(goal, w);
  const grow = div(goal, w);
  const start = u.cell;
  const nodes = new PathNodes(0x317 + 1);
  setAt(nodes.next, 0, 1);
  nodes.set(1, rem(start, w), div(start, w), 1, 0, 0, 2);
  nodes.set(2, uint8(at(buf, 20)), uint8(at(buf, 21)), 0, 30000, 1, 0);
  let free = 3;
  for (let cur = at(nodes.next, 0); cur !== 0; cur = at(nodes.next, cur)) {
    const ncur = at(nodes.col, cur);
    const rcur = at(nodes.row, cur);
    if (ncur === gcol && rcur === grow) {
      if (cur === 2) {
        throw new ErrUndefined(); // the sentinel's column and row are buffer bytes
      }
      const route = [grow * w + gcol];
      while (cur > 1) {
        cur = at(nodes.parent, cur);
        route.push(at(nodes.row, cur) * w + at(nodes.col, cur));
        if (route.length > 0x316) {
          return null;
        }
      }
      return route;
    }
    for (const k of neighbours64(ncur, rcur, g.width, g.height)) {
      if (k >= bits.length) {
        // The sentinel's column and row (buffer bytes 20 and 21) lie past the array: the search reads the
        // data behind it.
        throw new ErrUndefined();
      }
      if (
        k < 0 ||
        (at(bits, k) & reachBit) === 0 ||
        (at(bits, k) & visit) !== 0
      ) {
        continue;
      }
      const ncol = rem(k, 64);
      const nrow = div(k, 64);
      const d = int8(hexDistance(ncol, nrow, gcol, grow));
      // A cell past the right edge (see neighbours64) reads the map bytes of the next row's first cell.
      const cell = nrow * w + ncol;
      if (cell >= g.terrain.length) {
        throw new ErrUndefined();
      }
      const tr = at(t.terrains, at(g.terrain, cell));
      let cost = tr.cost;
      if ((u.flags4 & 0x10) !== 0) {
        cost = tr.airCost;
      }
      let score = 0;
      if (d !== 0) {
        score = int16(d * typ.distanceWeight + int8(cost) * typ.costWeight);
      }
      let pos = at(nodes.next, cur);
      while (at(nodes.score, pos) <= score) {
        pos = at(nodes.next, pos);
        if (pos === 0) {
          throw new ErrUndefined(); // past the sentinel into the list head
        }
      }
      const prev = at(nodes.prev, pos);
      nodes.set(free, ncol, nrow, cur, score, prev, pos);
      setAt(nodes.next, prev, free);
      setAt(nodes.prev, pos, free);
      bits[k] = at(bits, k) | visit;
      free++;
      if (free > 0x316) {
        return null;
      }
    }
  }
  return null;
}
