// Sappers dig and fill trenches. A unit whose live +6 has bit 0x2000 (the sappers' type word +0x11) gets the
// attack order on a neighbouring cell whose terrain record has flag 0x8000 (0ca6:13c5, the target marking), with
// or without a unit on it. The order (0ad1:000c) sets the moved flag, +0x11 to the cell and +0x13 = 0x11 and
// shows message 0x6e; it does not enter the order list, so it does not count for the move limit and fights no
// battle.
//
// The work runs in the action pass after the battles, in unit order with the captures (02dc:197b, 0e47:1b4c), on
// the action side's map; the phase change copies it to the other side's. A target below the first trench record
// is dug, any other filled. The work lays three cells: the target and its two neighbours across the sapper's
// front, a straight line where the hexes allow one, a bend toward the sapper where the target lies straight
// above or below it. Each of them whose terrain has flag 0x8000 takes the trench record the table gives, or the
// filled record 0x51 (plain ground that plays as grass); the others keep theirs.

import type { Live } from "./board.ts";
import { destroyed, liveCell } from "./board.ts";
import {
  actionDepot,
  actionHalf,
  depotWork,
  halfDepotWork,
} from "./builder.ts";
import { capture } from "./capture.ts";
import { actionGas, releaseGas } from "./gas.ts";
import {
  type Event,
  type Game,
  evDug,
  evFilled,
  liveOf,
  moveTablesOrNull,
  newEvent,
  sortedKeys,
} from "./cycle.ts";
import { at, setAt } from "./goruntime.ts";
import { div, rem } from "./int.ts";
import { neighbours } from "./support.ts";
import { actionTunnel, startTunnel } from "./tunnel.ts";
import { actionWire, layWire } from "./wire.ts";

/** Live +6: a sapper. */
export const flagSapper = 0x2000;
/** The terrain flag word: a sapper digs and fills here. */
export const terrainWorked = 0x8000;
/** +0x13: the sapper's work order. */
export const actionTrench = 0x11;
/** The first trench record; a target from here on is filled. */
export const trenchFirst = 0xa5;
/** TerrainFilled is the record a filled cell takes. */
export const terrainFilled = 0x51;
/** MsgTrench is the message of a sapper's work order. */
export const msgTrench = 0x6e;

/** sapper reports whether the unit digs and fills trenches instead of attacking (Go's Live.Sapper). */
export function sapper(u: Readonly<Live>): boolean {
  return (u.flags6 & flagSapper) !== 0;
}

/** working reports whether the unit has a sapper's work order this phase (Go's Live.Working). */
export function working(u: Readonly<Live>): boolean {
  return at(u.action, 0) === actionTrench;
}

/** trenchCell is a cell of a sapper's work: its step from the sapper (column, row) and the trench record it takes. */
type TrenchCell = readonly [dx: number, dy: number, rec: number];

/**
 * trenchWork are the three cells of the work on each neighbour of the sapper, by the sapper's column parity and
 * the neighbour in the order of steps (32f7:0061 and 00cd, indexed there by the column and row differences); the
 * target is the middle one.
 */
export const trenchWork: readonly (readonly (readonly TrenchCell[])[])[] = [
  [
    [
      [-1, -1, 0xaa],
      [0, -1, 0xa7],
      [1, -1, 0xab],
    ],
    [
      [0, -1, 0xab],
      [1, -1, 0xa7],
      [2, 0, 0xab],
    ],
    [
      [2, 0, 0xaa],
      [1, 0, 0xa7],
      [0, 1, 0xaa],
    ],
    [
      [1, 0, 0xaa],
      [0, 1, 0xa7],
      [-1, 0, 0xab],
    ],
    [
      [0, 1, 0xab],
      [-1, 0, 0xa7],
      [-2, 0, 0xab],
    ],
    [
      [-2, 0, 0xaa],
      [-1, -1, 0xa7],
      [0, -1, 0xaa],
    ],
  ],
  [
    [
      [-1, 0, 0xaa],
      [0, -1, 0xa7],
      [1, 0, 0xab],
    ],
    [
      [0, -1, 0xab],
      [1, 0, 0xa7],
      [2, 0, 0xab],
    ],
    [
      [2, 0, 0xaa],
      [1, 1, 0xa7],
      [0, 1, 0xaa],
    ],
    [
      [1, 1, 0xaa],
      [0, 1, 0xa7],
      [-1, 1, 0xab],
    ],
    [
      [0, 1, 0xab],
      [-1, 1, 0xa7],
      [-2, 0, 0xab],
    ],
    [
      [-2, 0, 0xaa],
      [-1, 0, 0xa7],
      [0, -1, 0xaa],
    ],
  ],
];

/** trench reports whether a terrain record is a trench (Go's Trench). */
export function trench(rec: number): boolean {
  return rec >= trenchFirst && rec <= 0xad;
}

/** worked reports whether a sapper digs and fills on a terrain record. */
export function worked(g: Game, rec: number): boolean {
  const t = moveTablesOrNull(g.profile);
  return (
    t !== null &&
    rec < t.terrains.length &&
    (at(t.terrains, rec).flags & terrainWorked) !== 0
  );
}

/** trenchTarget reports whether sapper u of side may work on cell: a neighbour of its cell on the side's map with worked terrain. */
export function trenchTarget(
  g: Game,
  side: number,
  u: Readonly<Live>,
  cell: number,
): boolean {
  const from = liveCell(u, side);
  const n = neighbours(from, g.width, g.height);
  return n.includes(cell) && worked(g, at(at(g.maps, side), 2 * cell));
}

/**
 * workCells is the work sapper unit of side would do on cell as the side's map shows it now, without doing it
 * (Go's Game.WorkCells): the cells it changes, the terrain record each takes and whether it fills (a trench on
 * cell) rather than digs. The cells are empty (Go's nil) for a cell the unit may not work on (no neighbour, or
 * terrain sappers leave alone).
 */
export function workCells(
  g: Game,
  side: number,
  unit: number,
  cell: number,
): [cells: number[], recs: number[], fill: boolean] {
  const u = g.units.get(unit);
  if (
    u === undefined ||
    (side !== 0 && side !== 1) ||
    cell < 0 ||
    cell >= g.width * g.height ||
    !trenchTarget(g, side, u, cell)
  ) {
    return [[], [], false];
  }
  return workCellsFrom(g, side, liveCell(u, side), cell);
}

/**
 * workCellsFrom is the work from cell from on cell to on side's map (Go's Game.workCells): the cells sappers work
 * among the three, the record each takes, and whether the work fills. A cell off the map is left alone (the
 * original's offset reads beyond the map's rows).
 */
export function workCellsFrom(
  g: Game,
  side: number,
  from: number,
  to: number,
): [cells: number[], recs: number[], fill: boolean] {
  const m = at(g.maps, side);
  const fill = at(m, 2 * to) >= trenchFirst;
  const cells: number[] = [];
  const recs: number[] = [];
  const k = neighbours(from, g.width, g.height).indexOf(to);
  if (k < 0) {
    return [cells, recs, fill];
  }
  const x = rem(from, g.width);
  const y = div(from, g.width);
  for (const [dx, dy, r] of at(at(trenchWork, x & 1), k)) {
    const cell = (y + dy) * g.width + x + dx;
    if (cell < 0 || cell >= g.width * g.height || !worked(g, at(m, 2 * cell))) {
      continue;
    }
    cells.push(cell);
    recs.push(fill ? terrainFilled : r);
  }
  return [cells, recs, fill];
}

/** gameTrench is the work of sapper i of side a on a's map, the terrain the cells take: dug below the first trench record, filled from there on (Go's Game.trench). */
export function gameTrench(g: Game, a: number, i: number): Event {
  const u = liveOf(g, i);
  const to = div(u.target, 2);
  const [cells, recs, fill] = workCellsFrom(g, a, liveCell(u, a), to);
  const ev = newEvent({ kind: evDug, side: a, unit: i, cell: to, cells });
  if (fill) {
    ev.kind = evFilled;
  }
  const m = at(g.maps, a);
  for (const [k, cell] of cells.entries()) {
    setAt(m, 2 * cell, at(recs, k));
  }
  return ev;
}

/**
 * work is the action pass for side a's resolution, in unit order: the captures (capture), the sappers' work
 * (gameTrench), the construction units' (builder.ts) and the extension work units' (gas.ts, wire.ts, tunnel.ts). It returns their events and the side that took an HQ
 * (-1 if none), the order handler's return value.
 */
export function work(g: Game, a: number): [evs: Event[], won: number] {
  let won = -1;
  const evs: Event[] = [];
  for (const i of sortedKeys(g.units)) {
    const u = liveOf(g, i);
    if (destroyed(u) || (at(u.action, 1) & 4) !== 0) {
      continue;
    }
    switch (at(u.action, 0)) {
      case 4: {
        const [ev, hq, ok] = capture(g, a, i);
        if (ok) {
          evs.push(ev);
          if (hq) {
            won = 1 - a;
          }
        }
        break;
      }
      case actionTrench:
        evs.push(gameTrench(g, a, i));
        break;
      case actionHalf: {
        const [ev, ok] = halfDepotWork(g, a, i);
        if (ok) {
          evs.push(ev);
        }
        break;
      }
      case actionDepot:
        evs.push(depotWork(g, a, i));
        break;
      case actionGas:
        evs.push(releaseGas(g, a, i));
        break;
      case actionWire:
        evs.push(layWire(g, a, i));
        break;
      case actionTunnel:
        evs.push(startTunnel(g, a, i));
        break;
    }
  }
  return [evs, won];
}
