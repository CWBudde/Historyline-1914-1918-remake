// The typed view of the movement, container and visibility rules for the command cycle: the record-level
// functions run as executed, and their records are read into Live and Hold values. Go's Profile methods take the
// profile first (p.Move(g, o) is move(p, g, o)); they throw ErrNoMovement where the build has no movement tables.

import {
  type Action,
  type ExitOrder,
  type Hold,
  containerRecords,
  holdBytes,
  moveEndCheck,
  moveExit,
} from "./container.ts";
import { GoError, at, setAt } from "./goruntime.ts";
import { div, rem, uint8 } from "./int.ts";
import {
  type MoveTables,
  type Movement,
  liveLen,
  moveDestinations,
  moveSeason,
  path,
  putWord,
  reach,
} from "./move.ts";
import { type Profile, newProfile } from "./profile.ts";
import { type MoveOrder, type StepRun, word } from "./record.ts";
import { execute } from "./route.ts";
import { type PhaseChange, changePhase } from "./visibility.ts";

/** ErrNoMovement is thrown where the profile's build has no movement tables. */
export class ErrNoMovement extends GoError {
  constructor() {
    super("dosrules: the profile has no movement tables");
  }
}

/** Live is a 26-byte live unit record (Build.Units in the demo). Go copies it by value: copy before changing one. */
export interface Live {
  /** +0: the movement budget, from type byte 0. */
  budget: number;
  /** +1: experience, 0-6. */
  exp: number;
  /** +2: group size; 0 marks the unit destroyed. */
  size: number;
  /** +3: the alternate strength. */
  alt: number;
  /** +4 word: side bits 1/2 (owns), the type's class word +0x0d, 0x4000 in a container, 0x8000 no unit. */
  flags4: number;
  /** +6 word: the type's word +0x11. */
  flags6: number;
  /** +8: the type record. */
  type: number;
  /** +9: the per-side serial number. */
  serial: number;
  /** +0x0a: a transporter's container record, a construction unit's builder charges, 0xff otherwise. */
  hold: number;
  /** +0x0b + 2*side: the map offset 2*cell+1 of the occupant byte where each side sees the unit (liveCell). */
  pos: [number, number];
  /** +0x0f + side: the direction each side sees it face. */
  facing: [number, number];
  /** +0x11 word: the order's destination 2*cell, then the start's position. */
  target: number;
  /** +0x13 to +0x19: the action bytes (Go's [7]int). */
  action: number[];
}

/** zeroLive is Go's Live{}, what a map read of a missing index gives. */
export function zeroLive(): Live {
  return {
    budget: 0,
    exp: 0,
    size: 0,
    alt: 0,
    flags4: 0,
    flags6: 0,
    type: 0,
    serial: 0,
    hold: 0,
    pos: [0, 0],
    facing: [0, 0],
    target: 0,
    action: [0, 0, 0, 0, 0, 0, 0],
  };
}

/** copyLive is Go's assignment of a Live: a copy that shares nothing with u. */
export function copyLive(u: Readonly<Live>): Live {
  return {
    ...u,
    pos: [u.pos[0], u.pos[1]],
    facing: [u.facing[0], u.facing[1]],
    action: [...u.action],
  };
}

/** decodeLive reads a live record (it panics, throws a RangeError, on a record shorter than 26 bytes). */
export function decodeLive(b: Uint8Array): Live {
  const u: Live = {
    budget: at(b, 0),
    exp: at(b, 1),
    size: at(b, 2),
    alt: at(b, 3),
    flags4: word(b, 4),
    flags6: word(b, 6),
    type: at(b, 8),
    serial: at(b, 9),
    hold: at(b, 0x0a),
    pos: [word(b, 0x0b), word(b, 0x0d)],
    facing: [at(b, 0x0f), at(b, 0x10)],
    target: word(b, 0x11),
    action: [],
  };
  for (let i = 0; i < 7; i++) {
    u.action.push(at(b, 0x13 + i));
  }
  return u;
}

/** liveRecord is Go's Live.Bytes: the live record. */
export function liveRecord(u: Readonly<Live>): Uint8Array {
  const b = new Uint8Array(liveLen);
  b[0] = uint8(u.budget);
  b[1] = uint8(u.exp);
  b[2] = uint8(u.size);
  b[3] = uint8(u.alt);
  putWord(b, 4, u.flags4);
  putWord(b, 6, u.flags6);
  b[8] = uint8(u.type);
  b[9] = uint8(u.serial);
  b[0x0a] = uint8(u.hold);
  for (let k = 0; k < 2; k++) {
    putWord(b, 0x0b + 2 * k, at(u.pos, k));
    b[0x0f + k] = uint8(at(u.facing, k));
  }
  putWord(b, 0x11, u.target);
  for (const [i, v] of u.action.entries()) {
    setAt(b, 0x13 + i, uint8(v));
  }
  return b;
}

/** liveCell is Go's Live.Cell: the cell where side sees the unit. */
export function liveCell(u: Readonly<Live>, side: number): number {
  return div(at(u.pos, side) - 1, 2);
}

/** destroyed is Go's Live.Destroyed: the record holds no unit (+4 bit 0x8000) or a destroyed one (size 0). */
export function destroyed(u: Readonly<Live>): boolean {
  return (u.flags4 & 0x8000) !== 0 || u.size === 0;
}

/** decodeHold reads a container record (holdRecord). It drops the bytes Hold does not name, +0x16 to +0x18 and +0x1a. */
export function decodeHold(b: Uint8Array): Hold {
  const h: Hold = { slots: [], cells: [], flags: 0, occupant: 0 };
  for (let i = 0; i < 14; i++) {
    h.slots.push(at(b, i));
  }
  for (let k = 0; k < 4; k++) {
    h.cells.push(word(b, 0x0e + 2 * k));
  }
  h.flags = word(b, 0x19);
  h.occupant = at(b, 0x1b);
  return h;
}

/** Message is a status message the rules show to a side. */
export interface Message {
  code: number;
  side: number;
}

function messages(ms: readonly (readonly [number, number])[]): Message[] {
  return ms.map(([code, side]) => ({ code, side }));
}

function lives(recs: ReadonlyMap<number, Uint8Array>): Map<number, Live> {
  const out = new Map<number, Live>();
  for (const [i, b] of recs) {
    out.set(i, decodeLive(b));
  }
  return out;
}

function bytesEqual(a: Uint8Array | undefined, b: Uint8Array): boolean {
  const x = a ?? new Uint8Array(0);
  return x.length === b.length && x.every((v, i) => v === b[i]);
}

// changed are the container records that differ from before.
function changed(
  before: ReadonlyMap<number, Uint8Array>,
  after: ReadonlyMap<number, Uint8Array>,
): Map<number, Hold> {
  const out = new Map<number, Hold>();
  for (const [i, b] of after) {
    if (!bytesEqual(before.get(i), b)) {
      out.set(i, decodeHold(b));
    }
  }
  return out;
}

// mapOccupants are the occupied cells of a map of two bytes per cell (Go's occupants).
function mapOccupants(m: Uint8Array): Map<number, number> {
  const out = new Map<number, number>();
  for (let c = 0; c < div(m.length, 2); c++) {
    const v = at(m, 2 * c + 1);
    if (v !== 0xff) {
      out.set(c, v);
    }
  }
  return out;
}

/** Moved is the outcome of a move order or an exit. */
export interface Moved {
  /** The order code or the move start refused the order; the unit stays. */
  refused: boolean;
  /** The mover is a ship two cells long. */
  twoCell: boolean;
  /** The mover's cell, as its side sees it, after each step. */
  route: number[];
  messages: Message[];
  /** The record of every placed unit afterwards. */
  units: Map<number, Live>;
  /** The container records that changed. */
  holds: Map<number, Hold>;
  /** The occupied cells of the mover's side's map and of the other side's map afterwards. */
  occupants: [Map<number, number>, Map<number, number>];
  /** For an exit, the budget left on every cell its reach search reached; null (Go's nil) for a move. */
  remaining: Map<number, number> | null;
}

/**
 * season is the profile with the season's terrain records; in winter without a winter table its movement
 * functions throw ErrNoMovement. Summer, or a profile already in winter, is p itself; winter is a new profile
 * with new tables sharing the build's records.
 */
export function season(p: Profile, inWinter: boolean): Profile {
  if (!inWinter || p.winter || p.tables === null || p.tables.move === null) {
    return p;
  }
  return newProfile({
    ...p,
    tables: { ...p.tables, move: moveSeason(p.tables.move, true) },
    winter: true,
  });
}

/** winter reports whether season gave the profile the winter terrain records. */
export function winter(p: Profile): boolean {
  return p.winter;
}

function moveTablesOf(p: Profile): MoveTables {
  const t = p.tables?.move ?? null;
  if (t === null) {
    throw new ErrNoMovement();
  }
  return t;
}

function movedOf(g: Movement, r: StepRun): Moved {
  return {
    refused: (r.order & 1) === 0 || (r.start & 1) === 0,
    twoCell: (r.order & 8) !== 0,
    route: [...r.cells],
    messages: messages(r.messages),
    units: lives(r.live),
    holds: changed(holdBytes(g), r.holds),
    occupants: [mapOccupants(r.maps[0]), mapOccupants(r.maps[1])],
    remaining: null,
  };
}

/**
 * destinations are the cells the moving unit may be ordered to: the reach search, then the end-cell check on
 * every reached cell (0b06:1383). Go's nil (none) is an empty array.
 */
export function destinations(p: Profile, g: Movement): number[] {
  const t = moveTablesOf(p);
  const [, bits] = reach(t, g);
  return moveDestinations(t, g, bits);
}

/** route is the route the path search gives from the unit's cell to goal, start first; null where the search finds none. */
export function route(p: Profile, g: Movement, goal: number): number[] | null {
  const t = moveTablesOf(p);
  const [buf, bits] = reach(t, g);
  const cells = path(t, g, buf, bits, goal);
  return cells === null ? null : cells.reverse();
}

/**
 * endCheck is the end-cell check of the moving unit on a cell: 0 if it may end there, else the code it shows,
 * and the action an arrival there takes.
 */
export function endCheck(
  p: Profile,
  g: Movement,
  cell: number,
): [code: number, action: Action] {
  return moveEndCheck(moveTablesOf(p), g, cell);
}

/**
 * move executes a move order: the order code, the move start and the steps until the unit arrives, loading into
 * a transporter or a building where the destination is one.
 */
export function move(p: Profile, g: Movement, o: MoveOrder): Moved {
  return movedOf(g, execute(moveTablesOf(p), g, o));
}

/** exit takes a unit out of a container and executes its move order. */
export function exit(p: Profile, g: Movement, o: ExitOrder): Moved {
  const r = moveExit(moveTablesOf(p), g, o);
  const m = movedOf(g, r);
  const remaining = new Map<number, number>();
  for (let c = 0; c < g.width * g.height; c++) {
    const v = at(r.remaining, div(c, g.width) * 64 + rem(c, g.width));
    if (v >= 0) {
      remaining.set(c, v);
    }
  }
  m.remaining = remaining;
  return m;
}

/** Phased is the state the phase change leaves. */
export interface Phased {
  /** The records of PhaseChange.units. */
  units: Map<number, Live>;
  /** The container records that changed. */
  holds: Map<number, Hold>;
  /** The occupied cells of side 0's and side 1's map (identical). */
  occupants: [Map<number, number>, Map<number, number>];
  /** The sides' phase bytes. */
  phase: [number, number];
  /** The turn counter. */
  count: number;
  messages: Message[];
}

/**
 * profileChangePhase is Go's Profile.ChangePhase: the phase change at the end of a side's orders; both sides'
 * views are brought into line and the phases swap.
 */
export function profileChangePhase(_p: Profile, g: PhaseChange): Phased {
  const r = changePhase(g);
  const units = new Map<number, Live>();
  for (const i of g.units.keys()) {
    units.set(i, decodeLive(r.live.get(i) ?? new Uint8Array(0)));
  }
  return {
    units,
    holds: changed(containerRecords(g.holds), r.holds),
    occupants: [mapOccupants(r.maps[0]), mapOccupants(r.maps[1])],
    phase: [at(r.players[0], 0x16), at(r.players[1], 0x16)],
    count: r.count,
    messages: messages(r.messages),
  };
}
