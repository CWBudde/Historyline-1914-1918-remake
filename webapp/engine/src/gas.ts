// Gas pioneers release a lasting gas cloud. They are an extension type (extension.json, docs/new-units.md), not
// the original's: the original tables have no type with live +6 bit 0x8000.
//
// A gas pioneer gets the attack order on any neighbouring cell, with or without a unit on it, while it has a
// charge (live +0x0a, as a construction unit's; a repair refills them). The order sets the moved flag, +0x11 to
// the cell and +0x13 = actionGas and stays off the order list, like a sapper's. The release runs in the action
// pass after the battles (work): the cloud covers the three cells of a sapper's work on that neighbour
// (trenchWork) that lie on the map, whatever their terrain, each set to the game's gas duration, and a charge is
// spent.
//
// Once per game turn, in the resolution whose phase change counts the turn (side 0's action phase), every unit on
// the map standing on a gassed cell loses part of its full group, own units too: foot units (live +6 bits 0x01 or
// 0x40) the game's rate, other ground units and ships a third of it, air units nothing; at least 1. A unit held in
// a building or a transporter is safe. A unit left with no strength is destroyed by the phase change that
// follows. Then each gassed cell thins by one turn. (Go's gas.go.)

import { type Live, liveCell } from "./board.ts";
import { charged } from "./builder.ts";
import {
  type Event,
  type EventKind,
  type Game,
  liveOf,
  msgNoTarget,
  newEvent,
  phaseAction,
  refuse,
  sortedKeys,
  vacate,
} from "./cycle.ts";
import { type GoError, at } from "./goruntime.ts";
import { div, rem } from "./int.ts";
import { type Profile, tablesOf } from "./profile.ts";
import { trenchWork } from "./sapper.ts";
import { neighbours } from "./support.ts";

/** Live +6: a gas pioneer (extension types only). */
export const flagGas = 0x8000;
/** +0x13: the gas pioneers' release order. */
export const actionGas = 0x15;
/** GasCharges are a gas pioneer's charges when placed, built or repaired. */
export const gasCharges = 2;
/** GasTurns is the default duration of a cloud in game turns. */
export const gasTurnsDefault = 2;
/** GasLoss is the default loss of a foot unit per turn, in percent of its full group. */
export const gasLossDefault = 30;
/** MsgGas is the message of a gas pioneer's release order (an extension code above the original's message bytes). */
export const msgGas = 0x100;
/** MsgNoGas refuses a gas pioneer with no charge left. */
export const msgNoGas = 0x101;

/** gas pioneers released a cloud at cell over cells. */
export const evGasCloud = "gasCloud" satisfies EventKind;
/** Unit of side lost loss strength to the gas at cell. */
export const evGassed = "gassed" satisfies EventKind;
/** The cloud left cells. */
export const evGasCleared = "gasCleared" satisfies EventKind;

/** gasPioneer reports whether the unit releases gas instead of attacking (Go's Live.GasPioneer). */
export function gasPioneer(u: Readonly<Live>): boolean {
  return (u.flags6 & flagGas) !== 0;
}

/** releasing reports whether the unit has a gas pioneer's release order this phase (Go's Live.Releasing). */
export function releasing(u: Readonly<Live>): boolean {
  return at(u.action, 0) === actionGas;
}

/** gasType reports whether typ is a gas pioneer (Go's Profile.gasType). */
export function gasType(p: Profile, typ: number): boolean {
  const t = p.tables?.move ?? null;
  return (
    t !== null &&
    typ >= 0 &&
    typ < t.types.length &&
    (at(t.types, typ).flags & flagGas) !== 0
  );
}

/** gasTurnsOf is the game's cloud duration (Go's Game.gasTurns). */
export function gasTurnsOf(g: Game): number {
  const t = g.gasTurns ?? 0;
  return t > 0 ? t : gasTurnsDefault;
}

/** gasLossOfGame is the game's foot rate in percent (Go's Game.gasLoss). */
export function gasLossOfGame(g: Game): number {
  const l = g.gasLoss ?? 0;
  return l > 0 ? l : gasLossDefault;
}

/** checkGas is check for a gas pioneer's attack order. */
export function checkGas(
  g: Game,
  side: number,
  u: Readonly<Live>,
  cell: number,
): GoError | null {
  if (!charged(u)) {
    return refuse(msgNoGas, "no gas left");
  }
  if (!neighbours(liveCell(u, side), g.width, g.height).includes(cell)) {
    return refuse(msgNoTarget, "gas is released on a neighbouring cell");
  }
  return null;
}

/**
 * gasCells are the cells a cloud released by unit of side on cell would cover, in the order of the work; empty
 * where the unit may not release there (Go's Game.GasCells).
 */
export function gasCells(
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
    !gasPioneer(u)
  ) {
    return [];
  }
  const from = liveCell(u, side);
  if (!neighbours(from, g.width, g.height).includes(cell)) {
    return [];
  }
  return gasCellsFrom(g, from, cell);
}

/** gasCellsFrom are the cells of a sapper's work from cell from on neighbour to that lie on the map. */
export function gasCellsFrom(g: Game, from: number, to: number): number[] {
  const k = neighbours(from, g.width, g.height).indexOf(to);
  if (k < 0) {
    return [];
  }
  const x = rem(from, g.width);
  const y = div(from, g.width);
  const cells: number[] = [];
  for (const [dx, dy] of at(at(trenchWork, x & 1), k)) {
    const cx = x + dx;
    const cy = y + dy;
    if (cx >= 0 && cx < g.width && cy >= 0 && cy < g.height) {
      cells.push(cy * g.width + cx);
    }
  }
  return cells;
}

/** orderGas carries out a checked release order. */
export function orderGas(
  g: Game,
  side: number,
  unit: number,
  cell: number,
): Event {
  const u = liveOf(g, unit);
  u.flags4 |= 0x200;
  u.target = 2 * cell;
  u.action[0] = actionGas;
  g.units.set(unit, u);
  return newEvent({
    kind: "work",
    side,
    unit,
    cell,
    messages: [{ code: msgGas, side }],
  });
}

/** releaseGas is the release of gas pioneer i of side a: the cloud's cells take the full duration and a charge is spent. */
export function releaseGas(g: Game, a: number, i: number): Event {
  const u = liveOf(g, i);
  const to = div(u.target, 2);
  const cells = gasCellsFrom(g, liveCell(u, a), to);
  if (g.gas === undefined || g.gas === null) {
    g.gas = Array.from({ length: g.width * g.height }, () => 0);
  }
  for (const c of cells) {
    g.gas[c] = gasTurnsOf(g);
  }
  if (charged(u)) {
    u.hold--;
  }
  g.units.set(i, u);
  return newEvent({ kind: evGasCloud, side: a, unit: i, cell: to, cells });
}

/** gasAt is the number of turns the cloud on cell has left, 0 for none (Go's Game.GasAt). */
export function gasAt(g: Game, cell: number): number {
  const gas = g.gas ?? null;
  if (gas === null || cell < 0 || cell >= gas.length) {
    return 0;
  }
  return at(gas, cell);
}

/** gasLossOf is the loss of unit u per turn on a gassed cell. */
export function gasLossOf(g: Game, u: Readonly<Live>): number {
  if ((u.flags4 & 0x10) !== 0) {
    return 0;
  }
  let pct = gasLossOfGame(g);
  if ((u.flags6 & 0x41) === 0) {
    pct = div(pct, 3);
  }
  const group = at(tablesOf(g.profile).types, u.type).group;
  return Math.max(1, div(group * pct + 99, 100));
}

/** unitGasCell is the cell a unit stands on as its own side sees it (side 0's view for a neutral one), -1 for a held unit. */
export function unitGasCell(u: Readonly<Live>): number {
  if ((u.flags4 & 0x4000) !== 0) {
    return -1;
  }
  return liveCell(u, (u.flags4 & 1) !== 0 ? 1 : 0);
}

/** ownerOf is the side a unit belongs to: 0, 1 or 2 for neutral. */
function ownerOf(u: Readonly<Live>): number {
  if ((u.flags4 & 1) !== 0) {
    return 1;
  }
  return (u.flags4 & 2) !== 0 ? 2 : 0;
}

/**
 * gasTurn is the turn's gas in a resolution: the losses of the units on gassed cells, then the thinning. It runs
 * only when the coming phase change counts the turn (Go's Game.gasTurn).
 */
export function gasTurn(g: Game): Event[] {
  if (
    g.gas === undefined ||
    g.gas === null ||
    g.players[0].phase !== phaseAction
  ) {
    return [];
  }
  const evs: Event[] = [];
  const gone: number[] = [];
  const cleared: number[] = [];
  for (const i of sortedKeys(g.units)) {
    const u = liveOf(g, i);
    if (
      (u.flags4 & 0x8000) !== 0 ||
      u.size === 0 ||
      (u.flags4 & 0xc0) === 0x40
    ) {
      continue; // a ship's rear goes with its head
    }
    let on = gasAt(g, unitGasCell(u)) > 0;
    const rear = g.units.get(i + 1);
    const ship =
      rear !== undefined &&
      (u.flags4 & 0xc0) === 0xc0 &&
      (rear.flags4 & 0xc0) === 0x40;
    if (ship && !on) {
      on = gasAt(g, unitGasCell(rear)) > 0;
    }
    let loss = gasLossOf(g, u);
    if (!on || loss === 0) {
      continue;
    }
    loss = Math.min(loss, u.size);
    u.size -= loss;
    if (u.size === 0) {
      u.action[1] = at(u.action, 1) | 4;
      gone.push(i);
    }
    g.units.set(i, u);
    if (ship) {
      const r = liveOf(g, i + 1);
      r.size = u.size;
      r.action[1] = at(r.action, 1) | (at(u.action, 1) & 4);
      g.units.set(i + 1, r);
    }
    evs.push(
      newEvent({
        kind: evGassed,
        side: ownerOf(u),
        unit: i,
        cell: unitGasCell(u),
        loss,
      }),
    );
  }
  let left = false;
  const gas = g.gas;
  for (const [c, t] of gas.entries()) {
    if (t <= 0) {
      continue;
    }
    gas[c] = t - 1;
    if (t === 1) {
      cleared.push(c);
    } else {
      left = true;
    }
  }
  if (!left) {
    delete g.gas;
  }
  vacate(g, gone);
  if (cleared.length > 0) {
    evs.push(newEvent({ kind: evGasCleared, side: -1, cells: cleared }));
  }
  return evs;
}
