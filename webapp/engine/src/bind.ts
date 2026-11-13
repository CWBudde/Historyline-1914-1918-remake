// internal/dosbind: scenario hexes to the original cell, live and container records.
import { type Content, type Hex, type Level, coordKey, grid } from "./level.ts";
import { type Coord, type Grid, contains, neighbors } from "./hexgrid.ts";
import { compoundPiece, at as compoundAt } from "./compound.ts";
import { type Game, type Placement, newGame } from "./cycle.ts";
import {
  type Hold,
  emptyHold,
  holdFlag40,
  holdFlag100,
  holdFlag400,
  holdTable,
  inBuilding,
  slotFree,
} from "./container.ts";
import { type Building } from "./energy.ts";
import { type Profile } from "./profile.ts";
import { season, moveTablesOf } from "./board.ts";
import { GoError, at } from "./goruntime.ts";
import { arrayField, int, intField, object, stringField } from "./decode.ts";

interface MapRecords {
  terrain: ReadonlyMap<string, number>;
  credits: ReadonlyMap<string, number>;
  builds: number[];
}
export interface UnitDefinition {
  dos: number;
  country: string;
  name: string;
  weight: number;
  armour: number;
  /**
   * from and until are the first and the last campaign month ("1915-04") in which a factory builds an extension
   * type; "" for no bound (Go's data.Unit From, Until).
   */
  from: string;
  until: string;
}
export interface BindingData {
  units: ReadonlyMap<string, UnitDefinition>;
  edge: number;
  terrain: ReadonlyMap<string, number>;
  maps: ReadonlyMap<string, MapRecords>;
}
export interface Placed extends Placement {
  cell: number;
  owner: number;
}
export interface Board {
  width: number;
  height: number;
  winter: boolean;
  terrain: number[];
  units: Map<number, Placed>;
  ids: Map<number, string>;
  holds: Map<number, Hold>;
  buildings: Map<number, Building>;
  parts: Map<number, number>;
  grid: Grid;
}
const none: MapRecords = { terrain: new Map(), credits: new Map(), builds: [] };
function fail(message: string): never {
  throw new GoError(`dosbind: ${message}`);
}
function numbers(raw: unknown): ReadonlyMap<string, number> {
  return raw == null
    ? new Map()
    : new Map(
        Object.entries(object(raw, "records")).map(([k, v]) => [k, int(v, k)]),
      );
}
/** The browser's loader supplies the existing units.json and records.json data. */
export function bindingData(rawUnits: unknown, records: unknown): BindingData {
  const units = new Map<string, UnitDefinition>();
  for (const [id, v] of Object.entries(object(rawUnits, "units"))) {
    const o = object(v, id);
    units.set(id, {
      dos: intField(o, "dos", id),
      country: stringField(o, "country", id),
      name: stringField(o, "name", id),
      weight: intField(o, "weight", id),
      armour: intField(o, "armour", id),
      from: stringField(o, "from", id),
      until: stringField(o, "until", id),
    });
  }
  const r = object(records, "records");
  const maps = new Map<string, MapRecords>();
  for (const [id, v] of Object.entries(object(r["maps"], "maps"))) {
    const o = object(v, id);
    maps.set(id, {
      terrain: numbers(o["terrain"]),
      credits: numbers(o["credits"]),
      builds: (arrayField(o, "builds", id) ?? []).map((v) => int(v, "build")),
    });
  }
  return {
    units,
    edge: intField(r, "edge", "records"),
    terrain: numbers(r["terrain"]),
    maps,
  };
}
export function side(player: number): number {
  return player === 1 ? 0 : player === 2 ? 1 : 2;
}
export function cell(b: Board, c: Coord): number | null {
  return contains(b.grid, c)
    ? Math.trunc(c.line / 2) * b.width + 2 * c.col + (c.line % 2)
    : null;
}
export function coord(b: Board, c: number): Coord | null {
  if (c < 0 || c >= b.width * b.height) return null;
  const p = {
    line: 2 * Math.trunc(c / b.width) + ((c % b.width) % 2),
    col: Math.trunc((c % b.width) / 2),
  };
  return contains(b.grid, p) ? p : null;
}
export function edge(b: Board, c: number): boolean {
  return (
    Math.trunc(c / b.width) === b.height - 1 || c % b.width === b.width - 1
  );
}
export function playable(b: Board): Grid {
  return { width: b.width / 2, lines: 2 * (b.height - 1) };
}
export function game(b: Board, p: Profile): Game {
  const g = newGame(
    season(p, b.winter),
    b.width,
    b.height,
    b.terrain,
    b.units,
    b.holds,
  );
  g.buildings = new Map(
    [...b.buildings].map(([i, bd]) => [
      i,
      {
        credits: bd.credits,
        builds: [...bd.builds],
        ...(bd.computerBuilds === undefined
          ? {}
          : { computerBuilds: [...bd.computerBuilds] }),
      },
    ]),
  );
  return g;
}
export function key(h: Hex): string {
  return (
    h.terrain.replace(/_w$/, "") +
    [...new Set(h.infra.ids)].map((id) => "+" + id).join("")
  );
}
interface BuildingCells {
  n: Coord;
  cells: Coord[];
  entrance: Coord;
  content: Content | null;
}
function buildings(lv: Level): Map<string, BuildingCells> {
  const result = new Map<string, BuildingCells>();
  const g = grid(lv);
  const terrain = (p: Coord): string =>
    lv.lines[p.line]?.[p.col]?.terrain.replace(/_w$/, "") ?? "";
  for (const [line, hs] of lv.lines.entries())
    for (const [col, h] of hs.entries()) {
      const found = compoundPiece(h.terrain);
      if (
        found === null ||
        found.compound.entrance === "" ||
        found.piece !== "N"
      )
        continue;
      const cp = found.compound;
      const n = { line, col };
      const bd: BuildingCells = { n, cells: [], entrance: n, content: null };
      for (const piece of Object.keys(cp.pieces).sort()) {
        const p = compoundAt(cp, n, piece);
        if (!contains(g, p) || terrain(p) !== cp.prefix + piece)
          return fail(`incomplete ${cp.prefix} at ${coordKey(n)}`);
        result.set(coordKey(p), bd);
        bd.cells.push(p);
        if (piece === cp.entrance) bd.entrance = p;
      }
    }
  for (const [line, hs] of lv.lines.entries())
    for (const [col, h] of hs.entries()) {
      const found = compoundPiece(h.terrain);
      const bd = result.get(coordKey({ line, col }));
      if (found !== null && found.compound.entrance !== "" && bd === undefined)
        return fail("building piece outside a complete building");
      if (h.content === null) continue;
      if (bd === undefined) return fail("content outside a building");
      if (bd.content !== null) return fail("a second content of the building");
      bd.content = { ...h.content };
    }
  return result;
}
function owner(flags: number): number {
  return (flags & 1) !== 0 ? 1 : (flags & 2) !== 0 ? 2 : 0;
}
function countrySide(country: string): number {
  if (["Germany", "Austria", "Bulgaria", "Ottoman"].includes(country)) return 0;
  return 1;
}
export function bind(lv: Level, data: BindingData, p: Profile): Board {
  const g = grid(lv);
  if (g.width < 2 || g.lines < 2) return fail("level too small");
  const b: Board = {
    width: 2 * g.width,
    height: Math.trunc((g.lines + 1) / 2),
    winter: false,
    grid: g,
    terrain: [],
    units: new Map(),
    ids: new Map(),
    holds: new Map(),
    buildings: new Map(),
    parts: new Map(),
  };
  b.terrain = Array.from({ length: b.width * b.height }, () => data.edge);
  const mr = data.maps.get(String(lv.mapNo)) ?? none;
  const groups = buildings(lv);
  const lookup = (k: string): number | undefined =>
    mr.terrain.get(k) ?? data.terrain.get(k);
  const types = moveTablesOf(p).types;
  const terrains = moveTablesOf(p).terrains;
  const toCell = (c: Coord): number => cell(b, c) ?? 0;
  for (const [line, hs] of lv.lines.entries())
    for (const [col, h] of hs.entries()) {
      const pos = { line, col };
      if (h.terrain.endsWith("_w")) b.winter = true;
      let k = key(h);
      const bd = groups.get(coordKey(pos));
      if (bd !== undefined) {
        const state =
          bd.content === null ? "-" : String(side(bd.content.owner));
        if (lookup(k + "@" + state) !== undefined) k += "@" + state;
      }
      let rec = lookup(k);
      while (rec === undefined && k.includes("+")) {
        k = k.slice(0, k.lastIndexOf("+"));
        rec = lookup(k);
      }
      if (rec === undefined) return fail(`no terrain record for ${k}`);
      const c = toCell(pos);
      if (!edge(b, c)) b.terrain[c] = rec;
    }
  if (lv.extra !== null && lv.extra.season !== "")
    b.winter = lv.extra.season === "winter";
  const unitType = (
    id: string,
    allowShip = false,
  ): { type: number; owner: number; ship: boolean } => {
    const u = data.units.get(id);
    if (u === undefined) return fail(`unknown unit ${id}`);
    if (u.dos < 0 || u.dos >= types.length)
      return fail(`unit ${id}: no type ${u.dos}`);
    const ship = (at(types, u.dos).class & 0x40) !== 0;
    if (ship && !allowShip)
      return fail(`unit ${id}: two-cell ships are not bound here`);
    return { type: u.dos, owner: countrySide(u.country), ship };
  };
  const add = (id: string, u: Placed): number => {
    const i = b.units.size;
    if (i >= slotFree) return fail(`more than ${slotFree} units`);
    b.units.set(i, u);
    if (id !== "") b.ids.set(i, id);
    return i;
  };
  for (let c = 0; c < b.width * b.height; c++) {
    const pos = coord(b, c);
    if (pos === null) continue;
    const h = at(at(lv.lines, pos.line), pos.col);
    if (h.unit === "") continue;
    const u = unitType(h.unit, true);
    if (edge(b, c)) return fail("unit on the map edge");
    if ((at(terrains, at(b.terrain, c)).flags & (0x40 | 0x100 | 0x400)) !== 0)
      return fail("unit on a building entrance");
    add(h.unit, { type: u.type, owner: u.owner, cell: c });
    if (u.ship) {
      // The first water neighbour without a unit: neither one of the
      // level's nor one placed before, such as another ship's rear.
      const rear = neighbors(g, pos).find((n) => {
        const h = at(at(lv.lines, n.line), n.col);
        const nc = toCell(n);
        return (
          h.unit === "" &&
          h.terrain.replace(/_w$/, "").startsWith("water") &&
          !edge(b, nc) &&
          ![...b.units.values()].some((p) => p.cell === nc)
        );
      });
      if (rear === undefined)
        return fail("two-cell ship has no free water neighbour");
      add("", { type: u.type + 1, owner: u.owner, cell: toCell(rear) });
    }
  }
  const credits = new Map<BuildingCells, number>();
  const builds: number[] = [];
  if (lv.extra !== null) {
    for (const [key, n] of lv.extra.credits) {
      const bd = groups.get(key);
      if (bd === undefined) return fail("credits outside a building");
      credits.set(bd, (credits.get(bd) ?? 0) + n);
    }
    for (const id of lv.extra.builds) {
      const typ = unitType(id).type;
      if (!builds.includes(typ)) builds.push(typ);
    }
  }
  const order = [...new Set(groups.values())].sort(
    (a, z) => toCell(a.entrance) - toCell(z.entrance),
  );
  const next = new Map([
    [0x100, holdFlag100],
    [0x400, holdFlag400],
  ]);
  for (const bd of order) {
    const entrance = toCell(bd.entrance);
    const flags = at(terrains, at(b.terrain, entrance)).flags;
    let idx: number;
    if ((flags & 0x40) !== 0) idx = holdFlag40 + (flags & 1);
    else if ((flags & (0x100 | 0x400)) !== 0) {
      const kind = flags & (0x100 | 0x400);
      idx = next.get(kind) ?? 0;
      if (idx === (kind === 0x100 ? holdFlag100 : holdFlag400) + holdTable)
        return fail("too many buildings of the kind");
      next.set(kind, idx + 1);
    } else {
      if (bd.content !== null)
        return fail("content in terrain which is no container");
      continue;
    }
    if (b.holds.has(idx)) return fail("second building for record");
    const h = emptyHold();
    h.flags = owner(flags);
    h.occupant = inBuilding;
    h.cells[0] = 2 * entrance;
    const rest = bd.cells
      .map(toCell)
      .filter((c) => c !== entrance)
      .sort((a, z) => a - z);
    for (let i = 1; i < h.cells.length && i - 1 < rest.length; i++)
      h.cells[i] = 2 * at(rest, i - 1);
    for (const c of rest) b.parts.set(c, entrance);
    if (bd.content !== null) {
      const held = bd.content.units ?? [];
      if (held.length > 7) return fail("more than 7 units in a building");
      for (const [slot, id] of held.entries()) {
        const typ = unitType(id).type;
        const u = add(id, { type: typ, owner: owner(flags), cell: entrance });
        h.slots[slot] = u;
        h.slots[7 + slot] = u;
      }
    }
    b.holds.set(idx, h);
    b.buildings.set(idx, {
      credits:
        lv.extra === null
          ? (mr.credits.get(String(idx)) ?? 0)
          : (credits.get(bd) ?? 0),
      builds:
        (flags & 0x400) === 0
          ? []
          : [...(lv.extra === null ? mr.builds : builds)],
      ...((flags & 0x400) !== 0 && (lv.extra?.computerBuilds?.length ?? 0) > 0
        ? { computerBuilds: [...(lv.extra?.computerBuilds ?? [])] }
        : {}),
    });
  }
  if (lv.extra === null)
    for (const k of mr.credits.keys())
      if (!/^[+-]?\d+$/.test(k) || !b.holds.has(Number(k)))
        return fail(`credits for absent building record ${k}`);
  let nextCarrier = 0;
  for (const [i, u] of b.units) {
    const t = at(types, u.type);
    if ((t.class & 0x1000) === 0) continue;
    if (t.carries === 0) return fail("no classes for transporter");
    if (nextCarrier === holdFlag40) return fail("too many transporters");
    const h = emptyHold();
    h.cells[0] = 2 * u.cell + 1;
    h.flags = 0x20 | t.carries | (u.owner & 3);
    h.occupant = i;
    b.holds.set(nextCarrier++, h);
  }
  return b;
}
