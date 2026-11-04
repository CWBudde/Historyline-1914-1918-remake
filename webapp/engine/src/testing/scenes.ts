// Test-only: the small scenes of the command cycle's Go tests (cycle_test.go, victory_test.go, energy_test.go,
// capture_test.go, sapper_test.go, builder_test.go, query_test.go and save_test.go's midRound), shared by their
// Vitest twins and the save tests.

import { expect } from "vitest";
import { destroyed } from "../board.ts";
import { type Hold, emptyHold, holdFlag40, holdFlag100 } from "../container.ts";
import {
  type Command,
  type Event,
  type Game,
  type Placement,
  Refusal,
  apply,
  cmdAttack,
  cmdMove,
  confirm,
  holdOf,
  liveOf,
  movement,
  newGame,
  resolve,
  zeroHold,
} from "../cycle.ts";
import { destinations } from "../board.ts";
import { at } from "../goruntime.ts";
import { type Profile, tablesOf } from "../profile.ts";
import { neighbours } from "../support.ts";
import { terrainWorked } from "../sapper.ts";
import {
  cBuilder,
  cInfantry,
  cPanzer,
  cSapper,
  cSupply,
  cTransportRec,
  moveTables,
} from "./tables.ts";

/** The scenes' map. */
export const cycleW = 9;
export const cycleH = 9;

/** The HQ entrance cells of side 0 and side 1 (victory_test.go). */
export const hq0 = 10;
export const hq1 = 70;

/** depotRec is a building record of terrain flag 0x100 that the scenes use, covering cell depot (energy_test.go). */
export const depotRec = holdFlag100;
export const depot = 76;

/** Units are a scene's placements by unit index (Go's map[int]Placement literal). */
export type Units = Readonly<Record<number, Placement>>;

/** Holds are a scene's container records by index (Go's map[int]Hold literal). */
export type Holds = Readonly<Record<number, Hold>>;

function unitMap(units: Units): Map<number, Placement> {
  return new Map(Object.entries(units).map(([k, v]) => [Number(k), v]));
}

function holdMap(holds: Holds | null): Map<number, Hold> {
  return new Map(Object.entries(holds ?? {}).map(([k, v]) => [Number(k), v]));
}

/** openTerrain is the cheapest terrain record without blocking or building flags that the scene's types may enter. */
export function openTerrain(p: Profile): number {
  const m = moveTables(p);
  let best = -1;
  for (const [i, tr] of m.terrains.entries()) {
    let ok = (tr.flags & 0x5547) === 0;
    for (const ty of [cInfantry, cPanzer, cSupply]) {
      ok = ok && (tr.mask & at(m.types, ty).mask) !== 0;
    }
    if (ok && (best < 0 || tr.cost < at(m.terrains, best).cost)) {
      best = i;
    }
  }
  if (best < 0) {
    throw new Error("no open terrain");
  }
  return best;
}

/** newCycle is a game on open terrain. */
export function newCycle(p: Profile, units: Units): Game {
  return newCycleHolds(p, units, null);
}

/** newCycleHolds is newCycle with container records. */
export function newCycleHolds(
  p: Profile,
  units: Units,
  holds: Holds | null,
): Game {
  const tr = openTerrain(p);
  const terrain = Array.from({ length: cycleW * cycleH }, () => tr);
  return newGame(p, cycleW, cycleH, terrain, unitMap(units), holdMap(holds));
}

/** applyOk is cycle_test.go apply: the command's events; it fails the test on an error. */
export function applyOk(g: Game, c: Command): Event[] {
  try {
    return apply(g, c);
  } catch (e) {
    throw new Error(`${JSON.stringify(c)}: ${String(e)}`, { cause: e });
  }
}

/** resolveOk is cycle_test.go resolve: both sides confirm and the orders resolve. */
export function resolveOk(g: Game): Event[] {
  confirm(g, 0);
  confirm(g, 1);
  return resolve(g);
}

/** refusal checks that err is the cycle's refusal with the message code. */
export function refusal(what: string, err: unknown, code: number): void {
  expect(err, `${what}: want a refusal`).toBeInstanceOf(Refusal);
  if (err instanceof Refusal) {
    expect(err.code, `${what}: ${err.message}`).toBe(code);
  }
}

/** transportHold is the container record of transporter unit at cell of side, taking every class, with the units of cargo in the side's slots. */
export function transportHold(
  cell: number,
  side: number,
  unit: number,
  ...cargo: number[]
): Hold {
  const h = emptyHold();
  h.cells[0] = 2 * cell + 1;
  h.flags = 0x20 | 0xfc0 | side;
  h.occupant = unit;
  for (const [k, u] of cargo.entries()) {
    h.slots[7 * side + k] = u;
  }
  return h;
}

/** hqCycle is newCycle with side 0's HQ entrance at hq0 and side 1's at hq1. */
export function hqCycle(p: Profile, units: Units): Game {
  return hqCycleHolds(p, units, null);
}

/** hqCycleHolds is hqCycle with further container records; the HQ records are given where holds has none. */
export function hqCycleHolds(
  p: Profile,
  units: Units,
  holds: Holds | null,
): Game {
  const m = moveTables(p);
  const open = openTerrain(p);
  const terrain = Array.from({ length: cycleW * cycleH }, () => open);
  for (const [i, tr] of m.terrains.entries()) {
    if ((tr.flags & 0x40) !== 0) {
      terrain[at([hq0, hq1], tr.flags & 1)] = i;
    }
  }
  if (
    (at(m.terrains, at(terrain, hq0)).flags & 0x40) === 0 ||
    (at(m.terrains, at(terrain, hq1)).flags & 0x40) === 0
  ) {
    throw new Error("no HQ terrain");
  }
  const h0 = emptyHold();
  const h1 = emptyHold();
  h0.cells[0] = 2 * hq0;
  h1.flags = 1;
  h1.cells[0] = 2 * hq1;
  const all = new Map<number, Hold>([
    [holdFlag40, h0],
    [holdFlag40 + 1, h1],
  ]);
  for (const [k, h] of holdMap(holds)) {
    all.set(k, h);
  }
  return newGame(p, cycleW, cycleH, terrain, unitMap(units), all);
}

/** armour is the sum of the armour of the units not destroyed. */
export function armour(g: Game): number {
  let s = 0;
  for (const u of g.units.values()) {
    if (!destroyed(u)) {
      s += at(tablesOf(g.profile).types, u.type).armour;
    }
  }
  return s;
}

/** InSlot is a unit placed in a slot of side's part of a record. */
export interface InSlot {
  rec: number;
  side: number;
  slot: number;
  unit: number;
}

/**
 * energyCycle is hqCycle in side 0's action phase (repairs and builds are given there) with side 0's HQ record
 * covering hq0, a depot record of side 1 covering depot, the HQ crediting 20 and building infantry and panzers,
 * the depot crediting 7, a neutral record crediting 99, the units of units and the slots of slots.
 */
export function energyCycle(
  p: Profile,
  units: Units,
  ...slots: InSlot[]
): Game {
  const h = emptyHold();
  h.cells[0] = 2 * hq0;
  const d = emptyHold();
  d.flags = 1;
  d.cells[0] = 2 * depot;
  const n = emptyHold();
  n.flags = 2;
  const holds: Record<number, Hold> = {
    [holdFlag40]: h,
    [depotRec]: d,
    [depotRec + 1]: n,
  };
  for (const s of slots) {
    const r = holds[s.rec] ?? zeroHold();
    r.slots[7 * s.side + s.slot] = s.unit;
    holds[s.rec] = r;
  }
  const g = hqCycleHolds(p, units, holds);
  g.buildings = new Map([
    [holdFlag40, { credits: 20, builds: [cInfantry, cPanzer, cBuilder, 13] }],
    [depotRec, { credits: 7, builds: [] }],
    [depotRec + 1, { credits: 99, builds: [] }],
  ]);
  g.players[0].phase = 2;
  g.players[1].phase = 1;
  return g;
}

/** held puts unit i into slot of side's part of record r. */
export function held(
  g: Game,
  r: number,
  side: number,
  slot: number,
  i: number,
): void {
  const h = holdOf(g, r);
  h.slots[7 * side + slot] = i;
  g.holds.set(r, h);
  const u = liveOf(g, i);
  u.flags4 |= 0x4000;
  u.pos = [at(h.cells, 0), at(h.cells, 0)];
  g.units.set(i, u);
  for (const m of g.maps) {
    for (let c = 0; c < g.width * g.height; c++) {
      if (m[2 * c + 1] === i) {
        m[2 * c + 1] = 0xff;
      }
    }
  }
}

/** inSlots reports whether any slot of any container record holds unit i. */
export function inSlots(g: Game, i: number): boolean {
  return [...g.holds.values()].some((h) => h.slots.includes(i));
}

/**
 * containerScene is a game on hqCycle's map: side 0's transporter at 40 carrying unit 10, infantry 0 next to it,
 * infantry 5 in side 0's HQ, a construction unit 6 and side 1's infantry 7 (query_test.go).
 */
export function containerScene(p: Profile): Game {
  const n = neighbours(40, cycleW, cycleH);
  const hq = emptyHold();
  hq.cells[0] = 2 * hq0;
  hq.slots[0] = 5;
  return hqCycleHolds(
    p,
    {
      0: { type: cInfantry, owner: 0, cell: at(n, 0) },
      1: { type: cSupply, owner: 0, cell: 40 },
      5: { type: cInfantry, owner: 0 },
      6: { type: cBuilder, owner: 0, cell: 60 },
      7: { type: cInfantry, owner: 1, cell: 44 },
      10: { type: cInfantry, owner: 0 },
    },
    { [cTransportRec]: transportHold(40, 0, 1, 10), [holdFlag40]: hq },
  );
}

/**
 * midRound is a round in progress: side 0 has moved a unit but not confirmed, side 1 has queued an attack on the
 * cell where its map still shows that unit and has confirmed (save_test.go).
 */
export function midRound(p: Profile): Game {
  const n = neighbours(40, cycleW, cycleH);
  const g = newCycle(p, {
    0: { type: cInfantry, owner: 0, cell: 40 },
    1: { type: cPanzer, owner: 1, cell: at(n, 0) },
    2: { type: cInfantry, owner: 0, cell: 70 },
    3: { type: cInfantry, owner: 1, cell: 10, exp: 3 },
  });
  g.limit = 8;
  g.concealed = true;
  const dest = destinations(p, movement(g, 0, 0)).find(
    (c) => !n.includes(c) && c !== 40,
  );
  if (dest === undefined) {
    throw new Error("no destination away from cell 40");
  }
  applyOk(g, { kind: cmdMove, side: 0, unit: 0, cell: dest });
  applyOk(g, { kind: cmdAttack, side: 1, unit: 1, cell: 40 });
  confirm(g, 1);
  return g;
}

/** terrainOf is the terrain record of cell on both maps, -1 where they differ (sapper_test.go). */
export function terrainOf(g: Game, cell: number): number {
  const a = at(g.maps[0], 2 * cell);
  return a === at(g.maps[1], 2 * cell) ? a : -1;
}

/** records are the terrain records of cells, -1 where the two maps differ (builder_test.go). */
export function records(g: Game, ...cells: number[]): number[] {
  return cells.map((c) => terrainOf(g, c));
}

/** unworked is a terrain record sappers leave alone. */
export function unworked(p: Profile): number {
  const i = moveTables(p).terrains.findIndex(
    (tr) => (tr.flags & (terrainWorked | 0x5547)) === 0,
  );
  if (i < 0) {
    throw new Error("no terrain without the sapper flag");
  }
  return i;
}

function scene(
  p: Profile,
  unit: number,
  from: number,
  enemy: number,
  terrain: Readonly<Record<number, number>> | null,
  holds: Holds | null,
): Game {
  const tr = Array.from({ length: cycleW * cycleH }, () => 0);
  for (const [c, r] of Object.entries(terrain ?? {})) {
    tr[Number(c)] = r;
  }
  return newGame(
    p,
    cycleW,
    cycleH,
    tr,
    unitMap({
      0: { type: cInfantry, owner: 0, cell: enemy },
      1: { type: unit, owner: 1, cell: from },
    }),
    holdMap(holds),
  );
}

/**
 * sapperScene is a game on grass (record 0, which sappers work) with side 1's sapper 1 at from and side 0's
 * infantry 0 at enemy; the records in terrain override the grass. Side 1 gives its orders first.
 */
export function sapperScene(
  p: Profile,
  from: number,
  enemy: number,
  terrain: Readonly<Record<number, number>> | null,
): Game {
  return scene(p, cSapper, from, enemy, terrain, null);
}

/**
 * builderScene is a game on grass with side 1's construction unit 1 at from and side 0's infantry 0 at enemy;
 * the records in terrain override the grass and holds are the container records. Side 1 is in the action phase.
 */
export function builderScene(
  p: Profile,
  from: number,
  enemy: number,
  terrain: Readonly<Record<number, number>> | null,
  holds: Holds | null,
): Game {
  return scene(p, cBuilder, from, enemy, terrain, holds);
}

/** buildingTerrain is the terrain record of kind flag (0x40, 0x100, 0x400) with owner bits owner (4 side 0, 1 side 1, 2 neutral). */
export function buildingTerrain(
  p: Profile,
  flag: number,
  owner: number,
): number {
  const i = moveTables(p).terrains.findIndex(
    (tr) => (tr.flags & flag) !== 0 && (tr.flags & 7) === owner,
  );
  if (i < 0) {
    throw new Error(`no terrain of flag ${flag} and owner ${owner}`);
  }
  return i;
}

/**
 * depotCycle is newCycle with a depot record of owner (0, 1, 2 neutral) at cell depot, its entrance terrain
 * showing the owner, the units of inside in both sides' slots (capture_test.go).
 */
export function depotCycle(
  p: Profile,
  owner: number,
  units: Units,
  ...inside: number[]
): Game {
  const tr = openTerrain(p);
  const terrain = Array.from({ length: cycleW * cycleH }, () => tr);
  terrain[depot] = buildingTerrain(p, 0x100, at([4, 1, 2], owner));
  const h = emptyHold();
  h.flags = owner;
  h.cells[0] = 2 * depot;
  for (const [k, u] of inside.entries()) {
    h.slots[k] = u;
    h.slots[7 + k] = u;
  }
  return newGame(
    p,
    cycleW,
    cycleH,
    terrain,
    unitMap(units),
    new Map([[depotRec, h]]),
  );
}
