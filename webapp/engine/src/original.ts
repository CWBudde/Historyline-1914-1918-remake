// internal/dosbind FromOriginal: an original map (finmap.ts) bound exactly as the map's files give it.
import { type Board, type BindingData, cell as cellOf, coord } from "./bind.ts";
import { moveTablesOf } from "./board.ts";
import {
  type Hold,
  emptyHold,
  holdFlag40,
  holdRecords,
  inBuilding,
  slotFree,
} from "./container.ts";
import { type Building } from "./energy.ts";
import type * as finmap from "./finmap.ts";
import { GoError, at } from "./goruntime.ts";
import { neighbors } from "./hexgrid.ts";
import { type Profile } from "./profile.ts";

const flagHQ = 0x40,
  flagDepot = 0x100,
  flagFactory = 0x400,
  transporter = 0x1000,
  carrier = 0x20,
  hqCredits = 50;
/** Terrain flags of a building's other cells, by the flag of its entrance. */
const partFlag = new Map([
  [flagHQ, 0x20],
  [flagDepot, 0x80],
  [flagFactory, 0x200],
]);

function fail(message: string): never {
  throw new GoError(`dosbind: ${message}`);
}
function owner(flags: number): number {
  return (flags & 1) !== 0 ? 1 : (flags & 2) !== 0 ? 2 : 0;
}
function countrySide(country: string): number {
  return ["Germany", "Austria", "Bulgaria", "Ottoman"].includes(country)
    ? 0
    : 1;
}

/** remakeIDs are the remake's unit ids by type and side ("type/side"), the first in id order where several share both. */
export function remakeIDs(data: BindingData): Map<string, string> {
  const out = new Map<string, string>();
  for (const id of [...data.units.keys()].sort()) {
    const u = data.units.get(id);
    if (u === undefined) continue;
    const k = `${u.dos}/${countrySide(u.country)}`;
    if (!out.has(k)) out.set(k, id);
  }
  return out;
}

function carrierHold(
  types: ReturnType<typeof moveTablesOf>["types"],
  u: { type: number; owner: number; cell: number },
  i: number,
): Hold | null {
  const classes = at(types, u.type).carries;
  if (classes === 0) return null;
  const h = emptyHold();
  h.cells[0] = 2 * u.cell + 1;
  h.flags = carrier | classes | (u.owner & 3);
  h.occupant = i;
  return h;
}

/** shape checks a unit of type class against the map: a ship's head has its rear, a transporter its record. */
function shape(cls: number, rear: boolean, container: boolean): void {
  if (((cls & 0xc0) === 0xc0) !== rear)
    fail(`class 0x${cls.toString(16)} with rear ${rear}`);
  if (((cls & transporter) !== 0) !== container)
    fail(`class 0x${cls.toString(16)} with container record ${container}`);
}

/**
 * fromOriginal binds an imported original map as the map's files give it (Go's dosbind.FromOriginal): the terrain
 * record of every cell, the units in the order the game creates them, the container records of the buildings and
 * transporters and the credits of the scenario's records. p is the full profile, whose tables the map's records
 * index.
 */
export function fromOriginal(
  l: finmap.Level,
  data: BindingData,
  p: Profile,
): Board {
  if (
    l.width < 2 ||
    l.width % 2 !== 0 ||
    l.height < 2 ||
    l.terrain.length !== l.height
  )
    return fail(`an original map of ${l.width}×${l.height} cells`);
  const b: Board = {
    width: l.width,
    height: l.height,
    winter: l.mission?.winter ?? false,
    terrain: [],
    grid: { width: l.width / 2, lines: 2 * l.height - 1 },
    units: new Map(),
    ids: new Map(),
    holds: new Map(),
    parts: new Map(),
    buildings: new Map(),
  };
  const { terrains, types } = moveTablesOf(p);
  for (const [row, line] of l.terrain.entries()) {
    if (line.length !== l.width)
      return fail(`terrain row ${row} of ${line.length} cells`);
    for (const [col, t] of line.entries())
      if (t < 0 || t >= terrains.length)
        return fail(`cell ${row},${col}: terrain record ${t}`);
    b.terrain.push(...line);
  }
  const cellAt = (c: readonly [number, number]): number =>
    c[0] < 0 || c[0] >= b.height || c[1] < 0 || c[1] >= b.width
      ? -1
      : c[0] * b.width + c[1];
  const ids = remakeIDs(data);
  const place = (typ: number, side: number, cell: number): number => {
    if (
      typ < 0 ||
      typ >= types.length ||
      side < 0 ||
      side > 2 ||
      cell < 0 ||
      cell >= b.terrain.length
    )
      return fail(`type ${typ} of side ${side} at cell ${cell}`);
    const i = b.units.size;
    if (i >= slotFree) return fail(`more than ${slotFree} units`);
    b.units.set(i, { type: typ, owner: side, cell });
    const id = ids.get(`${typ}/${side}`) ?? "";
    if (id !== "") b.ids.set(i, id);
    return i;
  };
  const addTransporter = (i: number, idx: number): void => {
    const u = b.units.get(i);
    if (u === undefined) return fail(`no unit ${i}`);
    const h = carrierHold(types, u, i);
    if (h === null) return fail(`no classes for transporter type ${u.type}`);
    if (b.holds.has(idx) || idx < 0 || idx >= holdFlag40)
      return fail(`container record ${idx}`);
    b.holds.set(idx, h);
  };
  for (const u of l.units) {
    try {
      const i = place(u.type, u.side, cellAt(u.cell));
      shape(
        at(types, u.type).class,
        u.rear !== undefined,
        u.container !== undefined,
      );
      if (u.rear !== undefined) place(u.type + 1, u.side, cellAt(u.rear));
      if (u.container !== undefined) addTransporter(i, u.container);
    } catch (e) {
      return fail(
        `unit at ${u.cell[0]},${u.cell[1]}: ${e instanceof Error ? e.message : String(e)}`,
      );
    }
  }
  originalBuildings(b, l, p, cellAt);

  const slots = new Map<number, number>();
  for (const c of l.containers ?? []) {
    if (c.holder === "none") {
      if (c.units.length > 0)
        return fail(
          `record for container ${c.container}, which the map lacks, creates units`,
        );
      continue;
    }
    const h = b.holds.get(c.container);
    if (h === undefined)
      return fail(`record for container ${c.container}, which the map lacks`);
    const bd = b.buildings.get(c.container);
    if (bd !== undefined && c.overridden !== true) bd.credits = c.energy;
    for (const u of c.units) {
      const k = slots.get(c.container) ?? 0;
      if (k === 7) return fail(`more than 7 units in container ${c.container}`);
      try {
        const i = place(u.type, c.side, Math.trunc(at(h.cells, 0) / 2));
        shape(at(types, u.type).class, false, u.container !== undefined);
        if (u.container !== undefined) addTransporter(i, u.container);
        h.slots[k] = i;
        h.slots[7 + k] = i;
      } catch (e) {
        return fail(
          `unit in container ${c.container}: ${e instanceof Error ? e.message : String(e)}`,
        );
      }
      slots.set(c.container, k + 1);
    }
  }
  return b;
}

/** originalBuildings gives every building its container record and credits. */
function originalBuildings(
  b: Board,
  l: finmap.Level,
  p: Profile,
  cellAt: (c: readonly [number, number]) => number,
): void {
  const terrains = moveTablesOf(p).terrains;
  const flags = (cell: number): number =>
    at(terrains, at(b.terrain, cell)).flags;
  const claimed = new Map<number, number>();
  for (const bd of l.buildings) {
    const entrance = cellAt(bd.cell);
    if (entrance < 0)
      return fail(`building at ${bd.cell[0]},${bd.cell[1]} off the map`);
    const f = flags(entrance);
    const kind = f & (flagHQ | flagDepot | flagFactory);
    const part = partFlag.get(kind);
    if (part === undefined)
      return fail(
        `building at ${bd.cell[0]},${bd.cell[1]} on terrain flags 0x${f.toString(16)}`,
      );
    if (
      b.holds.has(bd.container) ||
      bd.container < holdFlag40 ||
      bd.container >= holdRecords
    )
      return fail(
        `building at ${bd.cell[0]},${bd.cell[1]}: container record ${bd.container}`,
      );
    const rest: number[] = [];
    const stack = [entrance];
    while (stack.length > 0) {
      const c = stack.pop() ?? 0;
      const pos = coord(b, c);
      if (pos === null) continue;
      for (const n of neighbors(b.grid, pos)) {
        const nc = cellOf(b, n);
        if (nc === null) continue;
        const nf = flags(nc);
        if (
          (nf & part) === 0 ||
          (nf & (flagHQ | flagDepot | flagFactory)) !== 0 ||
          nc === entrance
        )
          continue;
        const e = claimed.get(nc);
        if (e !== undefined) {
          if (e !== entrance)
            return fail(
              `cell ${nc} belongs to the buildings at cells ${e} and ${entrance}`,
            );
          continue;
        }
        claimed.set(nc, entrance);
        rest.push(nc);
        b.parts.set(nc, entrance);
        stack.push(nc);
      }
    }
    const h = emptyHold();
    h.flags = owner(f);
    h.occupant = inBuilding;
    h.cells[0] = 2 * entrance;
    const sorted = [...rest].sort((x, y) => x - y);
    for (let k = 1; k < h.cells.length && k - 1 < sorted.length; k++)
      h.cells[k] = 2 * at(sorted, k - 1);
    b.holds.set(bd.container, h);
    const credit: Building = { credits: 0, builds: [] };
    if (kind === flagHQ) credit.credits = hqCredits;
    else if (kind === flagFactory) {
      credit.builds = [...(l.producible ?? [])];
      // Go's nil without a COM file's builds (the level JSON omits an empty list).
      if ((l.computerBuilds?.length ?? 0) > 0)
        credit.computerBuilds = [...(l.computerBuilds ?? [])];
    }
    b.buildings.set(bd.container, credit);
  }
}
