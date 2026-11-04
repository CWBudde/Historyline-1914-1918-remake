// The demo has no sight: no routine reads a sight range from the type or terrain records, and nothing hides a
// unit because it is out of view. Each side has its own map and its own position word in every live record
// (+0x0b + 2*side). During a phase a move changes only the moving side's (10cc:071e), and the order handler
// 02dc:0004 resolves the attacking side's orders on the attacking side's map. At its end the phase change brings
// both views into line; changePhase transcribes it.

import {
  type Hold,
  containerRecords,
  holdFlag40,
  holdFlag100,
  holdFlag400,
  holdTable,
  slotFree,
} from "./container.ts";
import { GoError, at, setAt } from "./goruntime.ts";
import { uint8 } from "./int.ts";
import { liveLen, owns, putWord } from "./move.ts";
import { word } from "./record.ts";
import { ErrUndefined } from "./support.ts";

/** LiveUnits: the phase change walks live records 0-0xf0. */
export const liveUnits = 0xf1;
/** PlayerLen is the size of the per-side records (Handler.Players). */
export const playerLen = 0x31;
/** AnimStart is the animation flags before the phase change. */
export const animStart = 0x0101;

/** Sighted is a live unit as the phase change reads it. Go copies it by value. */
export interface Sighted {
  flags4: number;
  /** +2: 0 marks the unit destroyed. */
  size: number;
  /** +0x0b + 2*side: 2*cell+1 as side 0 and side 1 see it. */
  pos: [number, number];
  /** +0x0f + side: the direction each side sees it face. */
  facing: [number, number];
  /** +0x0a (JSON omitempty). */
  hold: number;
}

/** sightedRecord is Go's Sighted.Bytes: the live record the phase change reads. */
export function sightedRecord(u: Readonly<Sighted>): Uint8Array {
  const r = new Uint8Array(liveLen);
  r[2] = uint8(u.size);
  putWord(r, 4, u.flags4);
  r[0x0a] = uint8(u.hold);
  for (let k = 0; k < 2; k++) {
    putWord(r, 0x0b + 2 * k, at(u.pos, k));
    r[0x0f + k] = uint8(at(u.facing, k));
  }
  return r;
}

/**
 * PhaseChange is what the phase change reads: both sides' maps (two bytes per cell, the terrain record and the
 * occupant, rows of width cells), the live units (every other record of 0-0xf0 carries 0x8000, no unit), the
 * container records (emptyHold where none is given), the side whose attack orders ran, the phase bytes (+0x16
 * of the per-side records; side 0's is Handler.Turn) and the turn counter. Go copies the struct but shares its
 * maps and slices.
 */
export interface PhaseChange {
  width: number;
  height: number;
  maps: [number[], number[]];
  units: Map<number, Sighted>;
  holds: Map<number, Hold>;
  attacker: number;
  phase: [number, number];
  count: number;
}

/** PhaseRun is the state the phase change leaves. */
export interface PhaseRun {
  /** Side 0's and side 1's map. */
  maps: [Uint8Array, Uint8Array];
  /** The records of PhaseChange.units. */
  live: Map<number, Uint8Array>;
  /** All container records. */
  holds: Map<number, Uint8Array>;
  /** The per-side records. */
  players: [Uint8Array, Uint8Array];
  /** The count bytes of the two attack order lists. */
  orders: [number, number];
  /** The turn counter. */
  count: number;
  /** The animation flags (animStart before). */
  anim: number;
  /** Status messages shown: code, side. */
  messages: [number, number][];
}

// noUnit is Sighted{Flags4: 0x8000}, the record of every index not in PhaseChange.units.
const noUnit: Readonly<Sighted> = {
  flags4: 0x8000,
  size: 0,
  pos: [0, 0],
  facing: [0, 0],
  hold: 0,
};

/** phaseRecords are the live records 0-0xf0 the phase change walks (Go's PhaseChange.Records). */
export function phaseRecords(g: PhaseChange): Map<number, Uint8Array> {
  const out = new Map<number, Uint8Array>();
  for (let idx = 0; idx < liveUnits; idx++) {
    out.set(idx, sightedRecord(g.units.get(idx) ?? noUnit));
  }
  return out;
}

/** playerRecord is a side's per-side record before the phase change: filler that tells the bytes apart, and the phase byte. */
export function playerRecord(side: number, phase: number): Uint8Array {
  const b = new Uint8Array(playerLen);
  for (let i = 0; i < b.length; i++) {
    b[i] = uint8(0x40 + playerLen * side + i);
  }
  b[0x16] = uint8(phase);
  return b;
}

/**
 * changePhase is the phase change in the tail of the order handler, 02dc:1a09-1cc6. A is the side whose orders
 * ran, B the other.
 *   - Container sync (02dc:23e5 on the transporter records and the three building tables): records with +0x19 &
 *     0x8002 are skipped. Of a record B owns (0ca0:020c with the side passed as the flags: B's flags & 1, or for
 *     side 0 none), B's slots are copied to A's, a slot holding a unit already marked 0x8000 cleared first; of
 *     any other, A's slots to B's.
 *   - Every occupant byte of A's map becomes 0xff.
 *   - Each unit of B (flags & 1, or for side 0 flags & 3 == 0) without 0x8000 gets B's position word and facing
 *     in A's.
 *   - Units with +2 == 0 are marked 0x8000 (02dc:2b1d).
 *   - Every unit without 0x8000 or 0x4000 (in a container) is written into A's map at its +0x0b, side 0's
 *     position word, whichever side A is.
 *   - A's whole map, terrain too, is copied over B's (0e40:1d9d).
 *   - The per-side records get +0x17..+0x19 = 0 and +0x1b = 5, their phase bytes swap between 1 and 2 (side 0
 *     becomes 2 if it was 1, else 1), both attack order counts become 0, the animation flags gain 0x18, messages
 *     0x35/0x36 go to side 0 and side 1 (0x35 to the side whose phase becomes 1), and the turn counter counts up
 *     when side 0's becomes 1.
 *
 * Undefined (ErrUndefined): a slot of B's other than a unit or 0xff in a record B owns, and a position word
 * +0x0b beyond the map. A map of the wrong size throws a GoError.
 */
export function changePhase(g: PhaseChange): PhaseRun {
  const a = g.attacker & 1;
  const b = 1 - (g.attacker & 1);
  const size = 2 * g.width * g.height;
  const maps: Uint8Array[] = [];
  for (let k = 0; k < 2; k++) {
    const m = at(g.maps, k);
    if (m.length !== size) {
      throw new GoError(`map ${k} holds ${m.length} bytes, not ${size}`);
    }
    maps.push(Uint8Array.from(m, uint8));
  }
  const r: PhaseRun = {
    maps: [at(maps, 0), at(maps, 1)],
    live: new Map(),
    holds: containerRecords(g.holds),
    players: [new Uint8Array(0), new Uint8Array(0)],
    orders: [0, 0],
    count: 0,
    anim: animStart | 0x18,
    messages: [],
  };
  const live = phaseRecords(g);
  const rec = (idx: number): Uint8Array => live.get(idx) ?? new Uint8Array(0);

  for (const [first, n] of [
    [0, 0x46],
    [holdFlag100, holdTable],
    [holdFlag400, holdTable],
    [holdFlag40, 2],
  ] as const) {
    for (let i = first; i < first + n; i++) {
      const h = r.holds.get(i) ?? new Uint8Array(0);
      const f = word(h, 0x19);
      if ((f & 0x8002) !== 0) {
        continue;
      }
      for (let k = 0; k < 7; k++) {
        if (!owns(b, f)) {
          setAt(h, 7 * b + k, at(h, 7 * a + k));
          continue;
        }
        let u = at(h, 7 * b + k);
        if (u <= slotFree && (word(rec(u), 4) & 0x8000) !== 0) {
          setAt(h, 7 * b + k, 0xff);
          u = 0xff;
        } else if (u > slotFree && u !== 0xff) {
          throw new ErrUndefined();
        }
        setAt(h, 7 * a + k, u);
      }
    }
  }

  const am = at(r.maps, a);
  for (let i = 1; i < size; i += 2) {
    am[i] = 0xff;
  }
  for (let idx = 0; idx < liveUnits; idx++) {
    const l = rec(idx);
    const f = word(l, 4);
    if (
      (f & 0x8000) !== 0 ||
      (b === 1 && (f & 1) === 0) ||
      (b === 0 && (f & 3) !== 0)
    ) {
      continue;
    }
    putWord(l, 0x0b + 2 * a, word(l, 0x0b + 2 * b));
    setAt(l, 0x0f + a, at(l, 0x0f + b));
  }
  for (let idx = 0; idx < liveUnits; idx++) {
    const l = rec(idx);
    if ((word(l, 4) & 0x8000) === 0 && at(l, 2) === 0) {
      putWord(l, 4, word(l, 4) | 0x8000);
    }
  }
  for (let idx = 0; idx < liveUnits; idx++) {
    const l = rec(idx);
    if ((word(l, 4) & 0xc000) !== 0) {
      continue;
    }
    const p = word(l, 0x0b);
    if (p >= size) {
      throw new ErrUndefined();
    }
    am[p] = idx;
  }
  at(r.maps, b).set(am);

  for (let k = 0; k < 2; k++) {
    const p = playerRecord(k, at(g.phase, k));
    p[0x17] = 0;
    p[0x18] = 0;
    p[0x19] = 0;
    p[0x1b] = 5;
    setAt(r.players, k, p);
  }
  let codes: [number, number] = [0x36, 0x35];
  r.count = g.count;
  const [p0, p1] = r.players;
  if (p0[0x16] === 1) {
    p0[0x16] = 2;
    p1[0x16] = 1;
  } else {
    p0[0x16] = 1;
    p1[0x16] = 2;
    codes = [0x35, 0x36];
    r.count = (g.count + 1) & 0xffff;
  }
  r.messages = [
    [codes[0], 0],
    [codes[1], 1],
  ];
  for (const idx of g.units.keys()) {
    // A unit past record 0xf0 has none (Go's nil record).
    r.live.set(idx, rec(idx));
  }
  return r;
}
