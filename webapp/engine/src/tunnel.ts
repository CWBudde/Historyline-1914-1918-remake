// Tunnellers (live +6 bit 0x0400, an extension type) dig a mine under a neighbouring cell and blow it. The attack
// order on any neighbouring cell on the map that is no water starts the tunnel, with or without a unit on it,
// unless the unit digs one already. The order sets the moved flag, +0x11 to the cell and +0x13 = actionTunnel;
// the work (work) after the battles opens the tunnel: Game.tunnels holds it until it blows.
//
// Once per game turn, with the gas (turnEffects), every tunnel whose tunnellers still stand on the cell it started
// from gains a turn; the others are abandoned. A tunnel with tunnelTurns blows: every unit on the mined cell that
// is no aircraft and not held in a building or transporter loses tunnelLoss percent of its full group, at least
// 1, own units too; a unit left with no strength is destroyed. Where the cell's terrain is worked ground, both
// sides' maps take terrainCrater, which costs more to cross and shelters like a trench. (Go's tunnel.go.)

import { type Live, destroyed, liveCell } from "./board.ts";
import {
  type Event,
  type EventKind,
  type Game,
  liveOf,
  moveTablesOrNull,
  msgNoTarget,
  newEvent,
  refuse,
  sortedKeys,
} from "./cycle.ts";
import { ownerOf, unitGasCell } from "./gas.ts";
import { type GoError, at, setAt } from "./goruntime.ts";
import { div } from "./int.ts";
import { tablesOf } from "./profile.ts";
import { worked } from "./sapper.ts";
import { neighbours } from "./support.ts";
import { terrainCrater } from "./wire.ts";

/** Live +6: tunnellers (extension types only). */
export const flagTunnel = 0x0400;
/** +0x13: the tunnellers' order. */
export const actionTunnel = 0x17;
/** TunnelTurns are the game turns a tunnel takes to blow. */
export const tunnelTurns = 2;
/** TunnelLoss is the loss of a unit on a blown mine, in percent of its full group. */
export const tunnelLoss = 70;
/** MsgTunnel is the message of the tunnellers' order. */
export const msgTunnel = 0x104;
/** MsgTunnelBusy refuses tunnellers that dig a tunnel already. */
export const msgTunnelBusy = 0x105;

/** Tunnellers began a mine under cell. */
export const evTunnel = "tunnel" satisfies EventKind;
/** The tunnellers left or lost their mine under cell. */
export const evTunnelLost = "tunnelLost" satisfies EventKind;
/** The mine under cell blew; cells holds cell where it left a crater. */
export const evMine = "mine" satisfies EventKind;
/** Unit of side lost loss strength to the mine at cell. */
export const evMined = "mined" satisfies EventKind;

/** Tunnel is a mine being dug: the tunnellers (unit) of side dig from cell from under the neighbouring cell and have dug dug turns. */
export interface Tunnel {
  unit: number;
  side: number;
  from: number;
  cell: number;
  dug: number;
}

/** tunneller reports whether the unit digs mines instead of attacking (Go's Live.Tunneller). */
export function tunneller(u: Readonly<Live>): boolean {
  return (u.flags6 & flagTunnel) !== 0;
}

/** tunnelOf is the tunnel unit digs, null where it digs none (Go's Game.TunnelOf). */
export function tunnelOf(g: Game, unit: number): Tunnel | null {
  for (const t of g.tunnels ?? []) {
    if (t.unit === unit) {
      return { ...t };
    }
  }
  return null;
}

/** mineable reports whether tunnellers may dig under cell on side's map: on the map and no water. */
export function mineable(g: Game, side: number, cell: number): boolean {
  const t = moveTablesOrNull(g.profile);
  if (t === null || cell < 0 || cell >= g.width * g.height) {
    return false;
  }
  const rec = at(at(g.maps, side), 2 * cell);
  return rec < t.terrains.length && (at(t.terrains, rec).mask & 0x80) === 0;
}

/** checkTunnel is check for the tunnellers' attack order. */
export function checkTunnel(
  g: Game,
  side: number,
  u: Readonly<Live>,
  idx: number,
  cell: number,
): GoError | null {
  if (tunnelOf(g, idx) !== null) {
    return refuse(msgTunnelBusy, "the tunnellers dig a tunnel already");
  }
  const n = neighbours(liveCell(u, side), g.width, g.height);
  if (!n.includes(cell) || !mineable(g, side, cell)) {
    return refuse(msgNoTarget, "no tunnel there");
  }
  return null;
}

/** startTunnel is the work of tunnellers i of side a: the tunnel opens. */
export function startTunnel(g: Game, a: number, i: number): Event {
  const u = liveOf(g, i);
  const t: Tunnel = {
    unit: i,
    side: a,
    from: liveCell(u, a),
    cell: div(u.target, 2),
    dug: 0,
  };
  g.tunnels = [...(g.tunnels ?? []).filter((x) => x.unit !== i), t];
  return newEvent({ kind: evTunnel, side: a, unit: i, cell: t.cell });
}

/**
 * mineTurn is the tunnels' turn: abandoned tunnels close, the others gain a turn, and those with tunnelTurns blow.
 * Units the mines destroy join gone.
 */
export function mineTurn(g: Game, gone: number[]): Event[] {
  const evs: Event[] = [];
  const left: Tunnel[] = [];
  for (const x of g.tunnels ?? []) {
    const t = { ...x };
    const u = g.units.get(t.unit);
    if (
      u === undefined ||
      destroyed(u) ||
      (at(u.action, 1) & 4) !== 0 ||
      unitGasCell(u) !== t.from ||
      ownerOf(u) !== t.side
    ) {
      evs.push(
        newEvent({
          kind: evTunnelLost,
          side: t.side,
          unit: t.unit,
          cell: t.cell,
        }),
      );
      continue;
    }
    t.dug++;
    if (t.dug < tunnelTurns) {
      left.push(t);
      continue;
    }
    evs.push(...blow(g, t, gone));
  }
  // Go's nil slice: the field goes once no tunnel is left.
  if (left.length > 0) {
    g.tunnels = left;
  } else {
    delete g.tunnels;
  }
  return evs;
}

/** blow is the mine of tunnel t going up: the crater, then the losses. */
function blow(g: Game, t: Readonly<Tunnel>, gone: number[]): Event[] {
  const ev = newEvent({
    kind: evMine,
    side: t.side,
    unit: t.unit,
    cell: t.cell,
  });
  for (const m of g.maps) {
    if (worked(g, at(m, 2 * t.cell))) {
      setAt(m, 2 * t.cell, terrainCrater);
      ev.cells = [t.cell];
    }
  }
  const evs = [ev];
  for (const i of sortedKeys(g.units)) {
    const u = liveOf(g, i);
    if (
      destroyed(u) ||
      (at(u.action, 1) & 4) !== 0 ||
      (u.flags4 & 0x10) !== 0 ||
      unitGasCell(u) !== t.cell
    ) {
      continue;
    }
    const group = at(tablesOf(g.profile).types, u.type).group;
    const loss = Math.min(
      Math.max(1, div(group * tunnelLoss + 99, 100)),
      u.size,
    );
    u.size -= loss;
    if (u.size === 0) {
      u.action[1] = at(u.action, 1) | 4;
      gone.push(i);
    }
    g.units.set(i, u);
    evs.push(
      newEvent({
        kind: evMined,
        side: ownerOf(u),
        unit: i,
        cell: t.cell,
        loss,
      }),
    );
  }
  return evs;
}
