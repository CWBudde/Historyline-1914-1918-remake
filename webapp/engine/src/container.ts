// Container records (Move.Containers, 0x1c bytes each) hold the units a transporter carries and the units in a
// building. Transporters index them by live +0x0a; the building records follow the transporters' in three
// tables, which the end-cell check picks by terrain flag: 32e2:38b4 (two records, chosen by flag bit 1, for
// flag 0x40), 32e2:38ec (flag 0x100) and 32e2:3a04 (flag 0x400), ten records each, found by cell (0e40:1728).

import { GoError, at, setAt } from "./goruntime.ts";
import { div, int8, uint8 } from "./int.ts";
import {
  type MoveTables,
  type Movement,
  type Piece,
  owns,
  pieceAt,
  putWord,
  reach,
} from "./move.ts";
import { containerLen, word, type MoveOrder, type StepRun } from "./record.ts";
import {
  type MoveState,
  bitIndex,
  drive,
  flags,
  holdsAt,
  newMoveState,
  rec,
  setFlags,
  setPos,
} from "./route.ts";
import { ErrUndefined } from "./support.ts";

export const holdRecords = 92;
/** (0x38b4 - 0x310c) / 0x1c */
export const holdFlag40 = 70;
export const holdFlag100 = 72;
export const holdFlag400 = 82;
/** The records 0e40:1728 searches. */
const holdTable = 10;
/** A slot or map occupant above it holds no unit. */
export const slotFree = 0xf0;
/** The map occupant of a building cell with units in it. */
export const inBuilding = 0xf4;

/**
 * Hold is a container record: seven slot bytes per side (unit indices, above 0xf0 free) at +7*side, the cells
 * it covers as map offsets (four words at +0x0e; a transporter's first is its position 2*cell+1, marked by flag
 * 0x20), the flag word +0x19 (side bits 1 and 2 as in live +4, 2 also a neutral building; bits 0xfc0 the
 * classes a transporter takes, matched against live +6) and +0x1b, the occupant the map gets back where a unit
 * leaves it. Slots ([14]int) and cells ([4]int) are Go arrays, copied with the record: copy them before
 * changing a copy.
 */
export interface Hold {
  slots: number[];
  cells: number[];
  flags: number;
  occupant: number;
}

/** emptyHold is a record with free slots and zeros otherwise, as every record not given in Movement.holds is set up. */
export function emptyHold(): Hold {
  return {
    slots: Array.from({ length: 14 }, () => 0xff),
    cells: [0, 0, 0, 0],
    flags: 0,
    occupant: 0,
  };
}

/** holdRecord is Go's Hold.Bytes: the 0x1c-byte container record. */
export function holdRecord(h: Hold): Uint8Array {
  const b = new Uint8Array(containerLen);
  for (const [i, v] of h.slots.entries()) {
    setAt(b, i, uint8(v));
  }
  for (const [k, c] of h.cells.entries()) {
    putWord(b, 0x0e + 2 * k, c);
  }
  putWord(b, 0x19, h.flags);
  b[0x1b] = uint8(h.occupant);
  return b;
}

/** holdBytes are all container records of a movement. */
export function holdBytes(g: Movement): Map<number, Uint8Array> {
  return containerRecords(g.holds);
}

/** containerRecords are all container records, emptyHold where holds has none. */
export function containerRecords(
  holds: ReadonlyMap<number, Hold>,
): Map<number, Uint8Array> {
  const out = new Map<number, Uint8Array>();
  for (let i = 0; i < holdRecords; i++) {
    const h = holds.get(i);
    out.set(i, h === undefined ? emptyRecord.slice() : holdRecord(h));
  }
  return out;
}

// emptyRecord is holdRecord(emptyHold()) (14 free slots, zeros otherwise), copied for every record a movement
// does not give. It is spelled out, not computed, so that loading the module calls nothing across the
// move/route/container import cycle.
const emptyRecord = new Uint8Array(containerLen).fill(0xff, 0, 14);

/** Action is what the end-cell check leaves in the action record 32e2:1805 for the arrival 10cc:0ae2. */
export interface Action {
  /** +0x0c: 3 a plain move, 2 into a transporter, 4 onto a foreign building, 5 into an own building. */
  kind: number;
  /** +1: the unit on the cell, as the map holds it (a ship's rear). */
  occupant: number;
  /** +4: the free slot taken. */
  slot: number;
  /** The record +5 points at. */
  hold: number;
  /** +9: the transporter's live +0x0a, or the building's index in its table (-1: none). */
  index: number;
}

/** zeroAction is Go's Action{}. */
export function zeroAction(): Action {
  return { kind: 0, occupant: 0, slot: 0, hold: 0, index: 0 };
}

/**
 * moveEndCheck is 10cc:0f74 for the moving unit (Go's MoveTables.EndCheck): 0 if it may end on the cell, else
 * the code it shows, and the action it leaves. On a unit's cell (a two-cell ship's rear stands for its head) it
 * gives 0x37 for another side's unit and 0x39 for an own unit that carries nothing. An own transporter takes
 * the unit into a free slot of the moving side (action 2) unless the unit's live +6 shares none of the classes
 * in the record's +0x19 & 0xfc0 or the transporter has moved (+4 bit 0x200: 0x3a), or the side's seven slots
 * are taken or the weights exceed the transporter type's capacity (0x3b). The weights are the unit type's, the
 * cargo's of a transporter that moves (which also needs a free slot per cargo unit plus one) and the side's
 * cargo already in the record. A building cell of terrain flag 0x400, 0x100 or 0x40 is action 4 if the
 * building's +0x19 is another side's or neutral (2), else action 5 into a free slot (0x3b if none, or too few
 * for a moving transporter's cargo); the terrain is not checked then. Otherwise it gives 0x40 for terrain the
 * unit's mask does not enter or, for an air unit, terrain flag 0x4000. It throws ErrUndefined for a record
 * outside the ones the runner sets up.
 */
export function moveEndCheck(
  t: MoveTables,
  g: Movement,
  cell: number,
): [code: number, action: Action] {
  return endCheckOn(newMoveState(t, g), cell);
}

/** endCheckOn is moveEndCheck on the memory s (Go's moveState.endCheck). */
export function endCheckOn(
  s: MoveState,
  cell: number,
): [code: number, action: Action] {
  const t = s.t;
  const side = s.g.side;
  const mover = rec(s, s.g.unit);
  const a: Action = { ...zeroAction(), kind: 3 };
  const occ = at(s.maps[0], 2 * cell + 1);
  if (occ <= slotFree) {
    a.occupant = occ;
    let o = occ;
    const f = flags(s, o);
    if ((f & 0x40) !== 0 && (f & 0x80) === 0) {
      o--; // the rear of a two-cell ship
    }
    const of = flags(s, o);
    if (!owns(of, side)) {
      return [0x37, a];
    }
    if ((of & 0x1000) === 0) {
      return [0x39, a];
    }
    a.index = int8(at(rec(s, o), 0x0a));
    a.hold = a.index;
    const h = stateHold(s, a.hold);
    if ((word(mover, 6) & word(h, 0x19) & 0xfc0) === 0 || (of & 0x200) !== 0) {
      return [0x3a, a];
    }
    const slot = freeSlot(h, side);
    if (slot === 7) {
      return [0x3b, a];
    }
    let fits = true;
    let w = at(t.types, at(mover, 8)).weight;
    if ((word(mover, 4) & 0x1000) !== 0) {
      const cargo = stateHold(s, int8(at(mover, 0x0a)));
      w += cargoWeight(s, cargo);
      if (slotCount(h, side, false) < slotCount(cargo, side, true) + 1) {
        fits = false;
      }
    }
    w += cargoWeight(s, h);
    if (w > at(t.types, at(rec(s, o), 8)).capacity || !fits) {
      return [0x3b, a];
    }
    a.slot = slot;
    a.kind = 2;
    return [0, a];
  }

  const tr = at(t.terrains, at(s.maps[0], 2 * cell));
  let base = -1;
  let idx = 0;
  if ((tr.flags & 0x400) !== 0) {
    [base, idx] = [holdFlag400, findHold(s, holdFlag400, 2 * cell)];
  } else if ((tr.flags & 0x100) !== 0) {
    [base, idx] = [holdFlag100, findHold(s, holdFlag100, 2 * cell)];
  } else if ((tr.flags & 0x40) !== 0) {
    [base, idx] = [holdFlag40, tr.flags & 1];
  }
  if (base >= 0) {
    // A cell no record lists (-1) reads the record before the table.
    a.index = idx;
    a.hold = base + idx;
    const h = stateHold(s, a.hold);
    let slot = freeSlot(h, side);
    const hf = word(h, 0x19);
    if (!owns(hf, side) || (hf & 2) !== 0) {
      a.kind = 4;
      return [0, a];
    }
    if ((word(mover, 4) & 0x1000) !== 0) {
      const cargo = stateHold(s, int8(at(mover, 0x0a)));
      if (slotCount(h, side, false) < slotCount(cargo, side, true) + 1) {
        slot = 7;
      }
    }
    if (slot === 7) {
      return [0x3b, a];
    }
    a.slot = slot;
    a.kind = 5;
    return [0, a];
  }
  if (
    (at(t.types, at(mover, 8)).mask & tr.mask) === 0 ||
    ((word(mover, 4) & 0x10) !== 0 && (tr.flags & 0x4000) !== 0)
  ) {
    return [0x40, a];
  }
  return [0, a];
}

/**
 * stateHold is a container record by index (Go's moveState.hold); one outside the records the runner sets up
 * is undefined.
 */
export function stateHold(s: MoveState, idx: number): Uint8Array {
  const h = s.holds.get(idx);
  if (h === undefined) {
    throw new ErrUndefined();
  }
  return h;
}

/** findHold is 0e40:1728: the record of the table at base listing the map offset among its four cells, or -1. */
export function findHold(s: MoveState, base: number, offset: number): number {
  for (let i = 0; i < holdTable; i++) {
    for (let k = 0; k < 4; k++) {
      if (word(holdsAt(s, base + i), 0x0e + 2 * k) === offset) {
        return i;
      }
    }
  }
  return -1;
}

/** freeSlot is the side's first free slot, 7 if none. */
export function freeSlot(h: Uint8Array, side: number): number {
  for (let i = 0; i < 7; i++) {
    if (at(h, 7 * side + i) > slotFree) {
      return i;
    }
  }
  return 7;
}

/** slotCount is 10cc:17cb: the side's taken slots, or its free ones. */
export function slotCount(h: Uint8Array, side: number, taken: boolean): number {
  let n = 0;
  for (let i = 0; i < 7; i++) {
    if (at(h, 7 * side + i) <= slotFree === taken) {
      n++;
    }
  }
  return n;
}

/** cargoWeight is 1307:15fd: the weights of the moving side's units in the record. */
export function cargoWeight(s: MoveState, h: Uint8Array): number {
  let w = 0;
  for (let i = 0; i < 7; i++) {
    const u = at(h, 7 * s.g.side + i);
    if (u <= slotFree) {
      w += at(s.t.types, at(rec(s, u), 8)).weight;
    }
  }
  return w;
}

/**
 * transfer is 10cc:1840: each unit in the moving side's slots of src moves to the first free slot of dst, if
 * there is one, and takes dst's first cell as its position.
 */
export function transfer(s: MoveState, src: Uint8Array, dst: Uint8Array): void {
  const side = s.g.side;
  for (let i = 0; i < 7; i++) {
    const u = at(src, 7 * side + i);
    if (u > slotFree) {
      continue;
    }
    for (let k = 0; k < 7; k++) {
      if (at(dst, 7 * side + k) > slotFree) {
        setAt(dst, 7 * side + k, u);
        setPos(s, u, word(dst, 0x0e));
        setAt(src, 7 * side + i, 0xff);
        break;
      }
    }
  }
}

/** ExitOrder is a unit leaving a container: the moving side's slot of record Hold, then a move order to Goal. */
export interface ExitOrder extends MoveOrder {
  hold: number;
  slot: number;
}

/** ExitRun is the executed outcome of an exit: the unit, the cost buffer the exit's reach search left, and the move order. */
export interface ExitRun extends StepRun {
  unit: number;
  remaining: Int8Array;
}

/**
 * moveExit is the exit branch of the main loop (0661:2f0a-30fe), then the move order as execute runs it (Go's
 * MoveTables.Exit). The unit is taken from the slot (which becomes free) and placed at the record's first
 * cell, one less for a transporter (flag 0x20); its live +4 gains 0x100, the side's map shows it there and the
 * cell gets the side's selection bit (0x40, 0x80 for side 1). The reach search runs from there with budget 4
 * for live +6 bit 0x10 and 8 otherwise, whatever live byte 0 holds, and the flag mask 3 (6 for live +4 bit 1),
 * with 0x18 for live +6 bit 4. The move start then takes the record's +0x1b as the occupant the cell gets
 * back. Undefined: an empty slot and an odd start offset (the map write would hit a terrain byte).
 */
export function moveExit(t: MoveTables, g: Movement, o: ExitOrder): ExitRun {
  return exitOn(t, newMoveState(t, g), o);
}

/**
 * exitQuery is the movement query of the exit from slot of the container record h on the query g of the side:
 * the unit in the slot, held at the record's first cell (one less for a transporter, flag 0x20), budget 4 for
 * live +6 bit 0x10 and 8 otherwise, the flag mask 3 (6 for live +4 bit 1), with 0x18 for live +6 bit 4. live
 * gives the unit's live record. It also returns the unit and the start offset; an empty slot and an odd start
 * offset or one past the map are undefined (ErrUndefined).
 */
export function exitQuery(
  g: Movement,
  h: Uint8Array,
  slot: number,
  live: (i: number) => Uint8Array,
): [moved: Movement, u: number, start: number] {
  const u = at(h, 7 * g.side + slot);
  let start = word(h, 0x0e);
  if ((word(h, 0x19) & 0x20) !== 0) {
    start--;
  }
  if (
    u > slotFree ||
    (start & 1) !== 0 ||
    div(start, 2) >= g.width * g.height
  ) {
    throw new ErrUndefined();
  }
  const m: Movement = { ...g, unit: u, budget: 8, block: 3 };
  const l = live(u);
  if ((word(l, 6) & 0x10) !== 0) {
    m.budget = 4;
  }
  if ((word(l, 4) & 1) !== 0) {
    m.block = 6;
  }
  if ((word(l, 6) & 4) !== 0) {
    m.block |= 0x18;
  }
  const p: Piece = { ...pieceAt(g, u), cell: div(start, 2) };
  m.held = new Map([[u, p]]);
  for (const [idx, q] of g.held) {
    if (idx !== u) {
      m.held.set(idx, q);
    }
  }
  m.units = new Map();
  for (const [idx, q] of g.units) {
    if (idx !== u) {
      m.units.set(idx, q);
    }
  }
  return [m, u, start];
}

/** exitOn is moveExit on the memory s (Go's MoveTables.exit). */
export function exitOn(t: MoveTables, s: MoveState, o: ExitOrder): ExitRun {
  const g = s.g;
  const side = g.side;
  const h = stateHold(s, o.hold);
  const [moved, u, start] = exitQuery(g, h, o.slot, (i) => rec(s, i));
  if (u !== g.unit) {
    throw new GoError(
      `exit: slot ${o.slot} holds unit ${u}, not Movement.Unit ${g.unit}`,
    );
  }
  setAt(h, 7 * side + o.slot, 0xff);
  setPos(s, u, start + 1);
  s.g = moved;

  const [buf, bits] = reach(t, moved);
  setAt(s.maps[0], start + 1, uint8(u));
  setFlags(s, u, flags(s, u) | 0x100);
  const i = bitIndex(s, div(start, 2));
  setAt(bits, i, at(bits, i) | uint8(0x40 << side));
  s.bits = bits;
  const move: MoveOrder = { goal: o.goal, retarget: o.retarget };
  if (o.stale !== undefined) {
    move.stale = o.stale;
  }
  return { unit: u, remaining: buf, ...drive(s, buf, move, o.hold) };
}
