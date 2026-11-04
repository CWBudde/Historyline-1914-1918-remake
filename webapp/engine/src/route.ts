// The move order of the player's order code: the order code, the move start and the steps until the unit
// arrives (route.go). The memory the routines read and change is a MoveState; Go's methods on it take it as
// their first parameter.

import {
  type Action,
  endCheckOn,
  inBuilding,
  holdBytes,
  slotFree,
  transfer,
  zeroAction,
} from "./container.ts";
import { GoError, at, setAt } from "./goruntime.ts";
import { div, int8, rem, uint8 } from "./int.ts";
import {
  type MoveTables,
  type Movement,
  liveLen,
  liveRecords,
  movementMaps,
  path,
  putWord,
  reach,
} from "./move.ts";
import { type MoveOrder, type StepRun, maxSteps, word } from "./record.ts";
import { ErrUndefined } from "./support.ts";

/** MoveState is the memory the move routines read and change (Go's moveState). */
export interface MoveState {
  t: MoveTables;
  g: Movement;
  live: Map<number, Uint8Array>;
  /** The side's map and the other side's map. */
  maps: [Uint8Array, Uint8Array];
  bits: Uint8Array;
  /** Cells from the goal back to the start, as the path search stores them. */
  route: number[];
  count: number;
  orders: number[];
  /** 32e2:1807 and 1808. */
  saved: [number, number];
  /** The container records. */
  holds: Map<number, Uint8Array>;
  /** The action record the order code's end-cell check left. */
  action: Action;
  messages: [number, number][];
}

/** newMoveState is the memory execute sets up for a movement. */
export function newMoveState(t: MoveTables, g: Movement): MoveState {
  return moveStateOn(t, g, liveRecords(g), movementMaps(g), holdBytes(g));
}

/**
 * moveStateOn is the memory of a whole game for a movement of g: the live records, both maps (the side's
 * first) and the container records as they are, in place of the ones Movement would set up. g must give the
 * side's map's units, as Game.movement projects them.
 */
export function moveStateOn(
  t: MoveTables,
  g: Movement,
  live: Map<number, Uint8Array>,
  maps: [Uint8Array, Uint8Array],
  holds: Map<number, Uint8Array>,
): MoveState {
  return {
    t,
    g,
    live,
    maps,
    bits: new Uint8Array(0),
    route: [],
    count: 0,
    orders: [],
    saved: [0, 0],
    holds,
    action: zeroAction(),
    messages: [],
  };
}

/** rec is a live record; one never placed reads as zeros, as in the game's array. */
export function rec(s: MoveState, idx: number): Uint8Array {
  let r = s.live.get(idx);
  if (r === undefined) {
    r = new Uint8Array(liveLen);
    s.live.set(idx, r);
  }
  return r;
}

export function pos(s: MoveState, idx: number): number {
  return word(rec(s, idx), 0x0b + 2 * s.g.side);
}

export function setPos(s: MoveState, idx: number, v: number): void {
  putWord(rec(s, idx), 0x0b + 2 * s.g.side, v);
}

export function flags(s: MoveState, idx: number): number {
  return word(rec(s, idx), 4);
}

export function setFlags(s: MoveState, idx: number, v: number): void {
  putWord(rec(s, idx), 4, v);
}

export function bitIndex(s: MoveState, cell: number): number {
  return div(cell, s.g.width) * 64 + rem(cell, s.g.width);
}

/** holdsAt is Go's s.holds[idx]: a record not set up is a nil slice, which panics on any index. */
export function holdsAt(s: MoveState, idx: number): Uint8Array {
  return s.holds.get(idx) ?? new Uint8Array(0);
}

function orBits(bits: Uint8Array, i: number, v: number): void {
  setAt(bits, i, at(bits, i) | uint8(v));
}

/**
 * execute is a move order as RunMove executes it (Go's MoveTables.Execute): reach, then the order code
 * 10cc:02dd, the move start 10cc:05c3 and the step 10cc:071e until it arrives (10cc:0ae2). It throws
 * ErrUndefined where path or the end-cell check do.
 */
export function execute(t: MoveTables, g: Movement, o: MoveOrder): StepRun {
  return executeOn(t, newMoveState(t, g), o);
}

/** executeOn is execute on the memory s (Go's MoveTables.execute). */
export function executeOn(t: MoveTables, s: MoveState, o: MoveOrder): StepRun {
  const g = s.g;
  const [buf, bits] = reach(t, g);
  for (const c of o.stale ?? []) {
    orBits(bits, bitIndex(s, c), 1 + g.side);
  }
  s.bits = bits;
  return drive(s, buf, o, -1);
}

/**
 * drive runs the order code, the start and the steps on the reach search's buffer and bits; from is the
 * container record an exiting unit leaves (-1: none).
 */
export function drive(
  s: MoveState,
  buf: Int8Array,
  o: MoveOrder,
  from: number,
): StepRun {
  const g = s.g;
  const ord = order(s, buf, o.goal);
  let started = 0;
  const cells: number[] = [];
  let dest = 2 * o.goal;
  if (o.retarget >= 0) {
    dest = 2 * o.retarget;
  }
  if ((ord & 1) !== 0) {
    started = start(s, dest, from);
  }
  while ((started & 1) !== 0) {
    const entry = s.count;
    step(s);
    cells.push(div(pos(s, g.unit) - 1, 2));
    if (entry === 0) {
      break;
    }
    if (cells.length === maxSteps) {
      throw new GoError("step: no arrival");
    }
  }
  const live = new Map<number, Uint8Array>();
  for (const idx of liveRecords(g).keys()) {
    live.set(idx, s.live.get(idx) ?? new Uint8Array(0));
  }
  return {
    order: ord,
    start: started,
    messages: s.messages,
    cells,
    live,
    maps: [s.maps[0], s.maps[1]],
    bits: s.bits,
    orders: s.orders,
    count: s.count,
    holds: s.holds,
    container: holdsAt(s, 0),
  };
}

/**
 * order is 10cc:02dd: it refuses (0x12) a destination without the side's reach bit, the unit's own cell,
 * terrain flag 0x4000 for an air unit and a non-zero end code, which it shows. Otherwise it searches the path
 * (which leaves the bits as it found them); without one it shows 0x48 and refuses. With one it clears the
 * side's reach bits everywhere, sets them on the route's cells and stores the destination in live +0x11. A
 * two-cell ship adds 8.
 */
export function order(s: MoveState, buf: Int8Array, goal: number): number {
  const { g, t } = s;
  const u = g.unit;
  const reachBit = g.side !== 0 ? 2 : 1;
  let res = 1;
  if ((at(s.bits, bitIndex(s, goal)) & reachBit) === 0) {
    res = 0x12;
  }
  if (pos(s, u) === 2 * goal + 1) {
    res = 0x12;
  }
  if (
    (flags(s, u) & 0x10) !== 0 &&
    (at(t.terrains, at(g.terrain, goal)).flags & 0x4000) !== 0
  ) {
    res = 0x12;
  }
  if ((res & 1) !== 0) {
    const [code, a] = endCheckOn(s, goal);
    s.action = a;
    if (code !== 0) {
      s.messages.push([code, g.side]);
      res = 0x12;
    }
  }
  if ((flags(s, u) & 0x40) !== 0) {
    res |= 8;
  }
  if ((res & 1) === 0) {
    return res;
  }
  const route = path(t, g, buf, s.bits, goal);
  // The path search clears the side's visit bit everywhere when it returns (0b06:000f), an exit's selection
  // bit included.
  const visit = uint8(0x40 << g.side);
  for (const [i, b] of s.bits.entries()) {
    s.bits[i] = b & ~visit;
  }
  if (route === null) {
    s.messages.push([0x48, g.side]);
    return 0x12;
  }
  for (const [i, b] of s.bits.entries()) {
    s.bits[i] = b & ~reachBit;
  }
  for (const c of route) {
    orBits(s.bits, bitIndex(s, c), reachBit);
  }
  s.route = route;
  s.count = route.length;
  putWord(rec(s, u), 0x11, 2 * goal);
  return res;
}

/**
 * start is 10cc:05c3: it refuses (0x12) an order whose destination is no longer the one in live +0x11.
 * Otherwise it appends the unit to the side's order list, stores the unit's cell in live +0x11, skips the
 * route's start and the first count step (count - 2) and clears both saved occupants; an exit order (from >=
 * 0) takes the first from the record's +0x1b.
 */
export function start(s: MoveState, dest: number, from: number): number {
  const u = s.g.unit;
  let res = 1;
  if (word(rec(s, u), 0x11) !== dest) {
    res = 0x12;
  }
  if ((flags(s, u) & 0x40) !== 0) {
    res |= 8;
  }
  if ((res & 1) === 0) {
    return res;
  }
  s.orders.push(u);
  putWord(rec(s, u), 0x11, pos(s, u));
  s.count -= 2;
  s.saved = [0xff, 0xff];
  if (from >= 0) {
    s.saved[0] = at(holdsAt(s, from), 0x1b);
  }
  return res;
}

/**
 * step is 10cc:071e: the unit enters route[count] on its side's map only. The cell it leaves gets back the
 * occupant saved when it entered (0xff at the start) and the occupant of the cell it enters is saved. A
 * two-cell ship's rear (the next record) takes the head's old cell and facing, its own old cell getting back
 * the second saved occupant. The facing +0x0f+side is the direction of the step (facing). A transporter moves
 * its cargo (cargo). With count 0 the unit has arrived (arrive); otherwise count drops by one.
 */
export function step(s: MoveState): void {
  const g = s.g;
  const u = g.unit;
  const side = g.side;
  const mp = s.maps[0];
  const next = 2 * at(s.route, s.count);
  if ((flags(s, u) & 0x40) !== 0) {
    const rear = u + 1;
    if (pos(s, rear) === 0) {
      // A head without its rear record (the game places both): the arrival marks the byte before the bit
      // array.
      throw new ErrUndefined();
    }
    const old = pos(s, u);
    setAt(mp, pos(s, rear), uint8(s.saved[1]));
    s.saved[1] = s.saved[0];
    s.saved[0] = at(mp, next + 1);
    setAt(mp, next + 1, uint8(u));
    setAt(rec(s, u), 0x0f + side, uint8(facing(old - 1, next, g.width)));
    setPos(s, u, next + 1);
    setAt(mp, old, uint8(u + 1));
    setAt(rec(s, rear), 0x0f + side, at(rec(s, u), 0x0f + side));
    setPos(s, rear, old);
  } else {
    const old = pos(s, u);
    setAt(mp, old, uint8(s.saved[0]));
    s.saved[0] = at(mp, next + 1);
    setAt(mp, next + 1, uint8(u));
    setAt(rec(s, u), 0x0f + side, uint8(facing(old - 1, next, g.width)));
    setPos(s, u, next + 1);
  }
  if ((flags(s, u) & 0x1000) !== 0) {
    cargo(s, u);
  }
  if (s.count > 0) {
    s.count--;
    return;
  }
  arrive(s);
}

/**
 * arrive is 10cc:0ae2 for the action the end-cell check left, then the end of 071e: it shows the arrival's
 * code (0x38) and clears live +4 bit 0x100, for the ship partner too.
 */
export function arrive(s: MoveState): void {
  const u = s.g.unit;
  const side = s.g.side;
  const a = s.action;
  const l = rec(s, u);
  switch (a.kind) {
    case 3:
      // The plain move: live +4 loses 0x4100 and gains the moved bit 0x200, which 0e40:1ddd also sets in the
      // bit array (0x10, 0x20 for side 1) at the unit's cell and at its ship partner's. +0x13 takes the
      // action, +0x14 gains 1 and +0x15 to +0x19 are cleared.
      setFlags(s, u, (flags(s, u) & 0xbeff) | 0x200);
      setMoved(s, u, 0x200);
      l[0x13] = 3;
      if ((flags(s, u) & 0x1000) !== 0) {
        cargo(s, u);
      }
      l[0x14] = at(l, 0x14) | 1;
      l.fill(0, 0x15, 0x1a);
      break;
    case 2: {
      // Into a transporter: the unit takes the free slot, the map shows the transporter again, live +4 loses
      // 0x100 and gains 0x4200 (the bit array gets no moved bit) and a transporter hands its cargo over.
      // +0x19 holds the slot, +0x15 the record, +0x16 the transporter.
      const h = holdsAt(s, a.hold);
      setAt(h, 7 * side + a.slot, uint8(u));
      setAt(s.maps[0], pos(s, u), uint8(a.occupant));
      setFlags(s, u, (flags(s, u) & 0xfeff) | 0x4200);
      if ((flags(s, u) & 0x1000) !== 0) {
        transfer(s, holdsAt(s, int8(at(l, 0x0a))), h);
      }
      l[0x13] = 2;
      l[0x19] = uint8(a.slot);
      l[0x15] = uint8(a.index);
      l[0x16] = uint8(a.occupant);
      l[0x14] = at(l, 0x14) | 1;
      break;
    }
    case 4:
      // Onto a foreign or neutral building: a plain arrival on the map; +0x17 takes the building's flag word,
      // +0x15 its index.
      setFlags(s, u, (flags(s, u) & 0xbeff) | 0x200);
      setMoved(s, u, 0x200);
      l[0x13] = 4;
      l[0x14] = at(l, 0x14) | 1;
      if ((flags(s, u) & 0x1000) !== 0) {
        cargo(s, u);
      }
      putWord(l, 0x17, word(holdsAt(s, a.hold), 0x19));
      l[0x15] = uint8(a.index);
      l[0x16] = 0;
      break;
    case 5: {
      // Into an own building: the unit takes the free slot, a transporter hands its cargo over, the map shows
      // 0xf4 and the unit's position becomes the building's first cell (an even map offset); live +4 loses
      // 0x100 and gains 0x4200.
      const h = holdsAt(s, a.hold);
      l[0x13] = 5;
      l[0x14] = at(l, 0x14) | 1;
      l[0x19] = uint8(a.slot);
      l[0x15] = uint8(a.index);
      putWord(l, 0x17, word(h, 0x19));
      setAt(h, 7 * side + a.slot, uint8(u));
      if ((flags(s, u) & 0x1000) !== 0) {
        transfer(s, holdsAt(s, int8(at(l, 0x0a))), h);
      }
      setAt(s.maps[0], pos(s, u), inBuilding);
      setPos(s, u, word(h, 0x0e));
      setFlags(s, u, (flags(s, u) & 0xfeff) | 0x4200);
      break;
    }
    default:
      throw new GoError(
        `arrival: action ${a.kind} is not set by the end-cell check`,
      );
  }
  s.messages.push([0x38, side]);
  setFlags(s, u, flags(s, u) & ~0x100);
  const p = partner(s, u);
  if (p !== null) {
    setFlags(s, p, flags(s, p) & ~0x100);
  }
}

/**
 * partner is the other record of a two-cell ship: the next one from the head (+4 bit 0x80), the previous one
 * from the rear; null for any other unit (Go's false).
 */
export function partner(s: MoveState, u: number): number | null {
  const f = flags(s, u);
  if ((f & 0x40) === 0) {
    return null;
  }
  if ((f & 0x80) !== 0) {
    return u + 1;
  }
  return u - 1;
}

/**
 * setMoved is 0e40:1ddd setting flag 0x200: live +4 and the side's moved bit at the unit's cell, for both
 * records of a ship.
 */
export function setMoved(s: MoveState, u: number, flag: number): void {
  const bit = (s.g.side & 1) !== 0 ? 0x20 : 0x10;
  const mark = (idx: number): void => {
    setFlags(s, idx, flags(s, idx) | flag);
    orBits(s.bits, bitIndex(s, div(pos(s, idx) - 1, 2)), bit);
  };
  mark(u);
  const p = partner(s, u);
  if (p !== null) {
    mark(p);
  }
}

/**
 * cargo is 159e:0805: the units in the side's seven slots of the unit's container record (live +0x0a) and the
 * record's first cell take the unit's position.
 */
export function cargo(s: MoveState, u: number): void {
  const c = holdsAt(s, int8(at(rec(s, u), 0x0a)));
  const p = pos(s, u);
  for (let k = 0; k < 7; k++) {
    const occ = at(c, 7 * s.g.side + k);
    if (occ <= slotFree) {
      setPos(s, occ, p);
    }
  }
  putWord(c, 0x0e, p);
}

/** facing is 10cc:143e: the direction (neighbours64 order) from map offset from to map offset to (2*cell each). */
export function facing(from: number, to: number, width: number): number {
  const fc = rem(div(from, 2), width);
  const fr = div(div(from, 2), width);
  const tc = rem(div(to, 2), width);
  const tr = div(div(to, 2), width);
  if (tc === fc) {
    return tr < fr ? 0 : 3;
  }
  if ((fc & 1) !== 0) {
    if (tc < fc && tr > fr) {
      return 4;
    }
    if (tc < fc) {
      return 5;
    }
    if (tr > fr) {
      return 2;
    }
    return 1;
  }
  if (tc < fc && tr < fr) {
    return 5;
  }
  if (tc < fc) {
    return 4;
  }
  if (tr < fr) {
    return 1;
  }
  return 2;
}
