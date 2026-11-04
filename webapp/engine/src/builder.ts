// Construction units build depots, half a depot per charge. Addresses in this file are the full build's (data
// segment 32f7). A unit whose live +6 has bit 0x4000 (the construction unit's type word +0x11; tested before the
// sappers' 0x2000) gets the attack order on a depot site (sites): with no charge left (+0x0a, a signed byte, not
// above 0) the selection shows message 0x4c, with no site message 0x4b (0ca6:1277-13bf). The order (0ad0:000d)
// sets the moved flag, +0x11 to the anchor and +0x13 = 0x0b, the first half, or 0x0d, the completion, where the
// anchor holds the half depot's N piece; a completion with ten depots on the map is refused (message 0x47).
// Either shows message 0x46 and stays off the order list.
//
// The work runs in the action pass with the captures and the sappers' (02dc:17b8, 02dc:1893), on the action
// side's map; the phase change copies it to the other side's.
//   - The first half lays the half depot's four pieces, provided all four cells, as the OTHER side's map shows
//     them, still have terrain sappers work (flag 0x8000) and no unit; otherwise nothing happens and the charge
//     stays. A charge is spent.
//   - The completion checks nothing: the next depot record (holdFlag100 plus the number of depots, 15a4:0330)
//     covers the four cells, owned by the completing side, the anchor takes that side's N piece and the others
//     the depot's E, S and W pieces. A charge is spent. The record credits no energy: the work clears the energy
//     bytes +0x16/+0x17 the creator set (the SHP record's energy, which a scenario depot gets at the level load),
//     and its +0x18 is 10.
//
// Either side may complete either side's half depot. Not modelled: the per-side depot count (32f7:4d3f, stride
// 0x17) the completion raises: only the statistics panel (130d:0eea) and the save header (1263) read it, and it
// equals the number of depot records the side owns.

import { type Live, liveCell } from "./board.ts";
import { kindDepot } from "./capture.ts";
import {
  emptyHold,
  holdFlag100,
  holdFlag400,
  holdTable,
  slotFree,
} from "./container.ts";
import {
  type Event,
  type Game,
  evDepot,
  evHalfDepot,
  evWork,
  liveOf,
  msgNoTarget,
  newEvent,
  occupant,
  refuse,
} from "./cycle.ts";
import { type GoError, at, setAt } from "./goruntime.ts";
import { div, int8, rem, uint8 } from "./int.ts";
import { worked } from "./sapper.ts";

/** Live +6: a construction unit. */
export const flagBuilder = 0x4000;
/** +0x13: the first half of a depot. */
export const actionHalf = 0x0b;
/** +0x13: the completion of a depot. */
export const actionDepot = 0x0d;
export const depotLimit = holdTable;

/** MsgDepot is the message of a construction unit's order. */
export const msgDepot = 0x46;
/** MsgDepotLimit refuses a completion with ten depots on the map. */
export const msgDepotLimit = 0x47;
/** MsgNoSite refuses a construction unit with no depot site. */
export const msgNoSite = 0x4b;
/** MsgNoCharge refuses a construction unit with no charge left. */
export const msgNoCharge = 0x4c;

/**
 * Depot terrain records (the constants file the game loads at 32f7:4c82), by the depot's pieces in the order N,
 * E, S, W: the half depot (4c9a), the complete depot's E, S and W pieces (4c9f), and its N piece by owner (4c85:
 * side 0, side 1).
 */
export const halfDepot: readonly number[] = [21, 24, 22, 23];
export const depotPiece: readonly number[] = [-1, 20, 18, 19];
export const depotN: readonly number[] = [16, 17];

/**
 * depotSites are the anchors a construction unit may build on, as (column, row) steps from its cell by its column
 * parity (32f7:0029/0033 even, 003d/0047 odd): the depots that touch the unit without covering it.
 */
const depotSites: readonly (readonly (readonly [number, number])[])[] = [
  [
    [0, -2],
    [1, -2],
    [2, -1],
    [2, 0],
    [1, 0],
    [0, 1],
    [-1, 0],
    [-2, 0],
    [-2, -1],
    [-1, -2],
  ],
  [
    [0, -2],
    [1, -1],
    [2, -1],
    [2, 0],
    [1, 1],
    [0, 1],
    [-1, 1],
    [-2, 0],
    [-2, -1],
    [-1, -1],
  ],
];

/**
 * depotSteps are a depot's cells N, E, S, W from its anchor, by the anchor's column parity (32f7:0051/0055 even,
 * 0059/005d odd; 0e47:1a8c).
 */
const depotSteps: readonly (readonly (readonly [number, number])[])[] = [
  [
    [0, 0],
    [1, 0],
    [0, 1],
    [-1, 0],
  ],
  [
    [0, 0],
    [1, 1],
    [0, 1],
    [-1, 1],
  ],
];

/** builder reports whether the unit builds depots instead of attacking (Go's Live.Builder). */
export function builder(u: Readonly<Live>): boolean {
  return (u.flags6 & flagBuilder) !== 0;
}

/** charged reports whether a construction unit has a charge left (+0x0a as a signed byte). */
export function charged(u: Readonly<Live>): boolean {
  return int8(uint8(u.hold)) > 0;
}

/** depotCells are the cells of the depot at anchor, N, E, S, W; a cell off the map is -1 (Go's Game.DepotCells). */
export function depotCells(g: Game, anchor: number): number[] {
  return depotCellsAt(g, rem(anchor, g.width), div(anchor, g.width));
}

/** depotCellsAt are the cells of the depot at column x, row y (Go's Game.depotCells). */
export function depotCellsAt(g: Game, x: number, y: number): number[] {
  return at(depotSteps, x & 1).map(([sx, sy]) => {
    const cx = x + sx;
    const cy = y + sy;
    return cx >= 0 && cx < g.width && cy >= 0 && cy < g.height
      ? cy * g.width + cx
      : -1;
  });
}

/** buildable reports whether a cell of side's map takes a depot piece: terrain sappers work and no unit. */
export function buildable(g: Game, side: number, cell: number): boolean {
  return (
    cell >= 0 &&
    worked(g, at(at(g.maps, side), 2 * cell)) &&
    occupant(g, side, cell) > slotFree
  );
}

/**
 * sites are the anchors construction unit of side may build on as the side's map shows it (0e47:1898): of the ten
 * around it, those that hold a half depot's N piece, and those whose four cells are on the map with terrain
 * sappers work and no unit. Charges do not count here. A unit that builds no depots has none.
 */
export function sites(g: Game, side: number, unit: number): number[] {
  const u = g.units.get(unit);
  if (u === undefined || !builder(u) || (side !== 0 && side !== 1)) {
    return [];
  }
  const from = liveCell(u, side);
  if (from < 0 || from >= g.width * g.height) {
    return [];
  }
  const x = rem(from, g.width);
  const y = div(from, g.width);
  const m = at(g.maps, side);
  const out: number[] = [];
  for (const [sx, sy] of at(depotSites, x & 1)) {
    const ax = x + sx;
    const ay = y + sy;
    // The anchor's own record is read without a bounds check, so a column off either edge reads the
    // neighbouring row's cell.
    const a = ay * g.width + ax;
    if (a >= 0 && a < g.width * g.height && at(m, 2 * a) === at(halfDepot, 0)) {
      out.push(a);
      continue;
    }
    const cells = depotCellsAt(g, ax, ay);
    if (!cells.some((c) => !buildable(g, side, c))) {
      out.push(at(cells, 0));
    }
  }
  return out;
}

/** checkBuilder is check for a construction unit's order on cell. */
export function checkBuilder(
  g: Game,
  side: number,
  u: Readonly<Live>,
  unit: number,
  cell: number,
): GoError | null {
  if (!charged(u)) {
    return refuse(msgNoCharge, "no charge left");
  }
  const s = sites(g, side, unit);
  if (s.length === 0) {
    return refuse(msgNoSite, "nowhere to build a depot");
  }
  if (!s.includes(cell)) {
    return refuse(msgNoTarget, "no depot site there");
  }
  if (
    at(at(g.maps, side), 2 * cell) === at(halfDepot, 0) &&
    depots(g) >= depotLimit
  ) {
    return refuse(msgDepotLimit, "no more depots");
  }
  return null;
}

/** depots is the number of depots on the map (32f7:4cb6): the depot records. */
export function depots(g: Game): number {
  let n = 0;
  for (const r of g.holds.keys()) {
    if (r >= holdFlag100 && r < holdFlag400) {
      n++;
    }
  }
  return n;
}

/** orderDepot is the order of construction unit i of side on cell, its event. */
export function orderDepot(
  g: Game,
  side: number,
  i: number,
  cell: number,
): Event {
  const u = liveOf(g, i);
  u.flags4 |= 0x200;
  u.target = 2 * cell;
  setAt(u.action, 0, actionHalf);
  if (at(at(g.maps, side), 2 * cell) === at(halfDepot, 0)) {
    setAt(u.action, 0, actionDepot);
  }
  g.units.set(i, u);
  return newEvent({
    kind: evWork,
    side,
    unit: i,
    cell,
    messages: [{ code: msgDepot, side }],
  });
}

/**
 * halfDepotWork is the first half by construction unit i of side a on a's map; ok is false where the other
 * side's map leaves the depot no room.
 */
export function halfDepotWork(
  g: Game,
  a: number,
  i: number,
): [ev: Event, ok: boolean] {
  const u = liveOf(g, i);
  const anchor = div(u.target, 2);
  const cells = depotCells(g, anchor);
  for (const c of cells) {
    if (!buildable(g, 1 - a, c)) {
      return [newEvent({ kind: evHalfDepot }), false];
    }
  }
  const m = at(g.maps, a);
  for (const [k, c] of cells.entries()) {
    setAt(m, 2 * c, at(halfDepot, k));
  }
  u.hold--;
  g.units.set(i, u);
  return [
    newEvent({ kind: evHalfDepot, side: a, unit: i, cell: anchor, cells }),
    true,
  ];
}

/** depotWork is the completion by construction unit i of side a on a's map. */
export function depotWork(g: Game, a: number, i: number): Event {
  const u = liveOf(g, i);
  const p = u.target;
  const r = holdFlag100 + depots(g);
  // 15a4:0330: the cells N, W, S, E, the W and E steps by the parity of the cell number (the column's where the
  // map's width is even).
  const h = emptyHold();
  const row = 2 * g.width;
  h.cells = [p, p - 2, p + row, p + 2];
  if ((div(p, 2) & 1) !== 0) {
    h.cells[1] = p + row - 2;
    h.cells[3] = p + row + 2;
  }
  h.flags = a | kindDepot;
  h.occupant = 0xff;
  g.holds.set(r, h);
  g.buildings.set(r, { credits: 0, builds: [] });
  const cells = depotCells(g, div(p, 2));
  const m = at(g.maps, a);
  for (const [k, c] of cells.entries()) {
    if (c < 0) {
      continue;
    }
    setAt(m, 2 * c, k === 0 ? at(depotN, a) : at(depotPiece, k));
  }
  u.hold--;
  g.units.set(i, u);
  return newEvent({
    kind: evDepot,
    side: a,
    unit: i,
    cell: div(p, 2),
    cells,
    hold: r,
  });
}
