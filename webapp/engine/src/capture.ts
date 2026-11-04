// Capture: the action pass of the order handler (02dc:0e91-19a4) after the action side's attacks and before the
// phase change. It walks the live records in order (work, which also runs the sappers' work in sapper.ts) and,
// for a unit whose arrival this cycle was onto another side's or a neutral building (action 4, 02dc:11a6), hands
// the building and every unit in it to the unit's side; the unit goes in. Only the movement side can have such an
// arrival, and only a type of flag +6 bit 1 (infantry and cavalry) reaches a foreign building's entrance. Taking
// an HQ wins the battle.
//
// Assumed, not verified: a captured unit's new serial number follows the side's highest one, as a built unit's
// does (the original counts per type and side).

import { at, setAt } from "./goruntime.ts";
import {
  holdFlag40,
  holdFlag100,
  holdFlag400,
  inBuilding,
  slotFree,
} from "./container.ts";
import {
  type Event,
  type Game,
  copyHold,
  evCaptured,
  holdOf,
  liveOf,
  moveTablesOrNull,
  newEvent,
  sortedKeys,
} from "./cycle.ts";
import { div } from "./int.ts";
import { owns } from "./move.ts";

/** The kind bits of a building record's flag word, as the record creators set them (159e:040e HQ, 159e:024b factory, 159e:032d depot). */
export const kindHQ = 0x04;
export const kindFactory = 0x08;
export const kindDepot = 0x10;
const kindBits = kindHQ | kindFactory | kindDepot;

/**
 * buildingKind is the kind bit of building record r: the HQ table (holdFlag40), the depot table (holdFlag100) or
 * the factory table (holdFlag400); 0 for a transporter's record.
 */
export function buildingKind(r: number): number {
  if (r >= holdFlag400) {
    return kindFactory;
  }
  if (r >= holdFlag100) {
    return kindDepot;
  }
  if (r >= holdFlag40) {
    return kindHQ;
  }
  return 0;
}

/** markKinds gives every building record its kind bit. */
export function markKinds(g: Game): void {
  for (const r of sortedKeys(g.holds)) {
    const h = holdOf(g, r);
    const k = buildingKind(r);
    if (k !== 0 && (h.flags & 0x8000) === 0) {
      h.flags = (h.flags & ~kindBits) | k;
      g.holds.set(r, h);
    }
  }
}

/**
 * capture is the capture of unit i in side a's action pass (work): its event, and whether it took an HQ; ok is
 * false where its building record is missing.
 */
export function capture(
  g: Game,
  a: number,
  i: number,
): [ev: Event, hq: boolean, ok: boolean] {
  const b = 1 - a;
  const u = liveOf(g, i);
  // The record by the flag word the arrival copied (+0x17) and its index (+0x15), 02dc:11a6-1208.
  let r = holdFlag100;
  const f = at(u.action, 4) | (at(u.action, 5) << 8);
  if ((f & kindHQ) !== 0) {
    r = holdFlag40;
  } else if ((f & kindFactory) !== 0) {
    r = holdFlag400;
  }
  r += at(u.action, 2);
  const found = g.holds.get(r);
  if (found === undefined) {
    return [newEvent({ kind: evCaptured }), false, false];
  }
  let h = copyHold(found);
  // B's slots take A's (137d), then the capturer's slot is B's first free one; in a full building the unit in
  // slot 0 is destroyed and the capturer takes its slot (13b4).
  for (let s = 0; s < 7; s++) {
    setAt(h.slots, 7 * b + s, at(h.slots, 7 * a + s));
  }
  let k = h.slots.slice(7 * b, 7 * b + 7).findIndex((s) => s > slotFree);
  if (k < 0) {
    k = 0;
    g.holds.set(r, copyHold(h));
    remove(g, at(h.slots, 7 * b));
    h = holdOf(g, r);
  }
  setAt(u.action, 6, k);
  // Every unit at the building's first cell as B sees it goes to B: the units held there and their cargo (1412).
  const first = at(h.cells, 0);
  for (const j of sortedKeys(g.units)) {
    const v = liveOf(g, j);
    if ((v.flags4 & 0x8000) !== 0 || at(v.pos, b) !== first) {
      continue;
    }
    v.flags4 = (v.flags4 & ~3) | b;
    v.serial = serial(g, b);
    g.units.set(j, v);
    if ((v.flags4 & 0x1000) !== 0) {
      const c = g.holds.get(v.hold);
      if (c !== undefined) {
        const x = copyHold(c);
        x.flags = (x.flags & ~3) | b;
        g.holds.set(v.hold, x);
      }
    }
  }
  // The entrance shows B's barrier on A's map (15d6); the phase change copies it to B's.
  const ma = at(g.maps, a);
  const [t, ok] = ownedTerrain(g, at(ma, first), b);
  if (ok) {
    setAt(ma, first, t);
  }
  const hq =
    (h.flags & (kindFactory | kindDepot)) === 0 && (h.flags & kindHQ) !== 0;
  h.flags = (h.flags & ~3) | b;
  setAt(h.slots, 7 * b + k, i);
  g.holds.set(r, h);
  // The capturer goes in (1749-1787): held, moved, at the building's first cell as B sees it; a transporter's
  // cargo follows.
  if (at(u.pos, b) < ma.length) {
    setAt(ma, at(u.pos, b), inBuilding);
  }
  setAt(u.pos, b, first);
  u.flags4 = (u.flags4 & ~0x100) | 0x4200;
  g.units.set(i, u);
  if ((u.flags4 & 0x1000) !== 0) {
    follow(g, i, b);
  }
  return [
    newEvent({
      kind: evCaptured,
      side: b,
      unit: i,
      cell: div(first, 2),
      hold: r,
    }),
    hq,
    true,
  ];
}

/**
 * ownedTerrain is the terrain record of building terrain tr that shows side as its owner: the same flags but the
 * owner bits (4 side 0, 1 side 1, 2 neutral).
 */
export function ownedTerrain(
  g: Game,
  tr: number,
  side: number,
): [rec: number, ok: boolean] {
  const t = moveTablesOrNull(g.profile);
  if (t === null || tr < 0 || tr >= t.terrains.length) {
    return [0, false];
  }
  const f = at(t.terrains, tr).flags;
  if ((f & 0x540) === 0) {
    return [0, false];
  }
  const want = (f & ~7) | at([4, 1], side);
  const i = t.terrains.findIndex((x) => x.flags === want);
  return i < 0 ? [0, false] : [i, true];
}

/** serial is the serial number of a unit side gains: one above its highest. */
export function serial(g: Game, side: number): number {
  let s = -1;
  for (const u of g.units.values()) {
    if ((u.flags4 & 2) === 0 && owns(u.flags4, side)) {
      s = Math.max(s, u.serial);
    }
  }
  return s + 1;
}

/**
 * remove is the destroy helper 159e:0595 on unit i: its record holds no unit (+4 0x8000), nor do a ship's other
 * half and a transporter's cargo, whose own record is cleared (+0x19 0x8000).
 */
export function remove(g: Game, i: number): void {
  if (!g.units.has(i)) {
    return;
  }
  const u = liveOf(g, i);
  if ((u.flags4 & 0x40) !== 0) {
    const p = (u.flags4 & 0x80) !== 0 ? i + 1 : i - 1;
    if (g.units.has(p)) {
      const v = liveOf(g, p);
      v.flags4 = 0x8000;
      g.units.set(p, v);
    }
  }
  if ((u.flags4 & 0x1000) !== 0) {
    const found = g.holds.get(u.hold);
    if (found !== undefined) {
      const c = copyHold(found);
      for (const s of c.slots) {
        if (g.units.has(s) && s <= slotFree) {
          const v = liveOf(g, s);
          v.flags4 = 0x8000;
          g.units.set(s, v);
        }
      }
      c.flags = 0x8000;
      g.holds.set(u.hold, c);
    }
  }
  u.flags4 = 0x8000;
  g.units.set(i, u);
}

/**
 * follow is 159e:0805 on the game: the units in side's slots of transporter i's record and the record's first
 * cell take i's position as side sees it.
 */
export function follow(g: Game, i: number, side: number): void {
  const u = liveOf(g, i);
  const found = g.holds.get(u.hold);
  if (found === undefined) {
    return;
  }
  const c = copyHold(found);
  for (const s of c.slots.slice(7 * side, 7 * side + 7)) {
    if (g.units.has(s) && s <= slotFree) {
      const v = liveOf(g, s);
      setAt(v.pos, side, at(u.pos, side));
      g.units.set(s, v);
    }
  }
  setAt(c.cells, 0, at(u.pos, side));
  g.holds.set(u.hold, c);
}
