// Extension work units (extension.json, docs/new-units.md) beyond the gas pioneers (gas.ts): the flamethrower,
// the wiring party, the tunnellers (tunnel.ts) and the field ambulance (ambulance.ts). Their abilities ride on
// live +6 bits the original's types leave unused.
//
// A flamethrower (bit 0x0800) attacks like any unit, but the defender's terrain lends it no defence: a positive
// defence term counts as 0 (Context.flame).
//
// A wiring party (bit 0x1000) gets the attack order on a neighbouring cell whose terrain a sapper works (flag
// 0x8000) and that is neither a trench nor wire yet, with or without a unit on it, while it has a charge (live
// +0x0a, as a construction unit's; a repair refills them). The order sets the moved flag, +0x11 to the cell and
// +0x13 = actionWire, like a sapper's. The work runs in the action pass after the battles (work) on the action
// side's map: of the three cells of a sapper's work on that neighbour (trenchWork), each such cell takes
// terrainWire, and a charge is spent. Wire plays like a trench for who may enter (terrain mask 0x08: foot units,
// heavy tanks and light guns, not cavalry, cars or light tanks) but costs a whole foot move and shelters nobody;
// sappers clear it as they fill a trench. (Go's wire.go.)

import { type Live, liveCell } from "./board.ts";
import { charged } from "./builder.ts";
import { flagFlame } from "./casualties.ts";
import {
  type Event,
  type EventKind,
  type Game,
  evWork,
  liveOf,
  msgNoTarget,
  newEvent,
  refuse,
} from "./cycle.ts";
import { gasCellsFrom } from "./gas.ts";
import { type GoError, at, setAt } from "./goruntime.ts";
import { div } from "./int.ts";
import type { Profile } from "./profile.ts";
import { trench, worked } from "./sapper.ts";
import { neighbours } from "./support.ts";

/** Live +6: a flamethrower (extension types only; casualties.ts holds it, where Context.flame reads it). */
export { flagFlame };
/** Live +6: a wiring party. */
export const flagWire = 0x1000;
/** +0x13: the wiring party's work order. */
export const actionWire = 0x16;
/** TerrainWire is the extension terrain record of barbed wire on open ground. */
export const terrainWire = 0xaf;
/** TerrainCrater is the extension terrain record of a mine's crater (tunnel.ts). */
export const terrainCrater = 0xb0;
/** WireCharges are a wiring party's charges when placed, built or repaired. */
export const wireCharges = 2;
/** MsgWire is the message of a wiring party's order (an extension code, like msgGas). */
export const msgWire = 0x102;
/** MsgNoWire refuses a wiring party with no charge left. */
export const msgNoWire = 0x103;

/** A wiring party laid wire at cell over cells. */
export const evWired = "wired" satisfies EventKind;

/** flamethrower reports whether the unit's attacks ignore the defender's terrain defence (Go's Live.Flamethrower). */
export function flamethrower(u: Readonly<Live>): boolean {
  return (u.flags6 & flagFlame) !== 0;
}

/** wiring reports whether the unit lays barbed wire instead of attacking (Go's Live.Wiring). */
export function wiring(u: Readonly<Live>): boolean {
  return (u.flags6 & flagWire) !== 0;
}

/** wire reports whether a terrain record is barbed wire (Go's Wire). */
export function wire(rec: number): boolean {
  return rec === terrainWire;
}

/** typeFlag reports whether type typ has live +6 bit flag (Go's Profile.typeFlag). */
export function typeFlag(p: Profile, typ: number, flag: number): boolean {
  const t = p.tables?.move ?? null;
  return (
    t !== null &&
    typ >= 0 &&
    typ < t.types.length &&
    (at(t.types, typ).flags & flag) !== 0
  );
}

/** wireable reports whether a wiring party lays wire on a terrain record: worked ground that is no trench and no wire yet. */
export function wireable(g: Game, rec: number): boolean {
  return worked(g, rec) && !trench(rec) && !wire(rec);
}

/** checkWire is check for a wiring party's attack order. */
export function checkWire(
  g: Game,
  side: number,
  u: Readonly<Live>,
  cell: number,
): GoError | null {
  if (!charged(u)) {
    return refuse(msgNoWire, "no wire left");
  }
  const n = neighbours(liveCell(u, side), g.width, g.height);
  if (!n.includes(cell) || !wireable(g, at(at(g.maps, side), 2 * cell))) {
    return refuse(msgNoTarget, "no wire work there");
  }
  return null;
}

/**
 * wireCells are the cells a wiring party unit of side would wire from its cell on cell as the side's map shows it
 * now, ascending in the order of the work; empty (Go's nil) where it may not (Go's Game.WireCells).
 */
export function wireCells(
  g: Game,
  side: number,
  unit: number,
  cell: number,
): number[] {
  const u = g.units.get(unit);
  if (
    u === undefined ||
    (side !== 0 && side !== 1) ||
    cell < 0 ||
    cell >= g.width * g.height ||
    !wiring(u)
  ) {
    return [];
  }
  const n = neighbours(liveCell(u, side), g.width, g.height);
  if (!n.includes(cell) || !wireable(g, at(at(g.maps, side), 2 * cell))) {
    return [];
  }
  return wireCellsFrom(g, side, liveCell(u, side), cell);
}

/** wireCellsFrom are the cells of a sapper's work from from on to that take wire on side's map (Go's Game.wireCells). */
export function wireCellsFrom(
  g: Game,
  side: number,
  from: number,
  to: number,
): number[] {
  const m = at(g.maps, side);
  return gasCellsFrom(g, from, to).filter((c) => wireable(g, at(m, 2 * c)));
}

/** orderWork carries out a checked work order with action byte action and message msg (the wiring party's, the tunnellers'). */
export function orderWork(
  g: Game,
  side: number,
  unit: number,
  cell: number,
  action: number,
  msg: number,
): Event {
  const u = liveOf(g, unit);
  u.flags4 |= 0x200;
  u.target = 2 * cell;
  setAt(u.action, 0, action);
  g.units.set(unit, u);
  return newEvent({
    kind: evWork,
    side,
    unit,
    cell,
    messages: [{ code: msg, side }],
  });
}

/** layWire is the work of wiring party i of side a on a's map. */
export function layWire(g: Game, a: number, i: number): Event {
  const u = liveOf(g, i);
  const to = div(u.target, 2);
  const cells = wireCellsFrom(g, a, liveCell(u, a), to);
  const m = at(g.maps, a);
  for (const c of cells) {
    setAt(m, 2 * c, terrainWire);
  }
  if (charged(u)) {
    u.hold--;
  }
  g.units.set(i, u);
  return newEvent({ kind: evWired, side: a, unit: i, cell: to, cells });
}
