// internal/origview: an imported original map (finmap.ts) as a level the front end can draw. Every terrain record
// gets a look, the remake's terrain and infrastructure named in looks.json; the view decides only what is drawn,
// the rules bind the records themselves (original.ts). Lines come from cells through finmap.coord; the map edge
// cells it keeps take the look of a neighbour inside the map. Building entrances carry the owner as content.
import { type BindingData } from "./bind.ts";
import { DecodeError, object, stringField } from "./decode.ts";
import * as finmap from "./finmap.ts";
import { GoError, at } from "./goruntime.ts";
import { type Coord, type Grid, around, contains } from "./hexgrid.ts";
import { type Hex, type Level, grid as levelGrid } from "./level.ts";
import { remakeIDs } from "./original.ts";

/** records is the number of terrain records; edge is the map edge's. */
export const records = 175,
  edge = 174;

/** Look is how a terrain record is drawn: a terrain of terrain.json and the infrastructure overlays in order. */
export interface Look {
  terrain: string;
  winter: string;
  infra: string[];
}
export type Looks = ReadonlyMap<number, Look>;

/** parseLooks reads looks.json ({"looks": {"<record>": {terrain, winter?, infra?}}}). */
export function parseLooks(raw: unknown): Looks {
  const looks = object(object(raw, "looks.json")["looks"], "looks");
  const out = new Map<number, Look>();
  for (const [k, v] of Object.entries(looks)) {
    if (!/^\d+$/.test(k)) throw new DecodeError(`looks: record ${k}`);
    const o = object(v, k);
    const infra = o["infra"] ?? [];
    if (!Array.isArray(infra) || infra.some((s) => typeof s !== "string"))
      throw new DecodeError(`looks: ${k}: infra`);
    out.set(Number(k), {
      terrain: stringField(o, "terrain", k),
      winter: typeof o["winter"] === "string" ? o["winter"] : "",
      infra: [...(infra as string[])],
    });
  }
  return out;
}

/** ground is the terrain a look shows: in winter its winter form, else terrain+"_w" where terrains has it. */
export function ground(
  k: Look,
  winter: boolean,
  terrains: ReadonlySet<string>,
): string {
  if (!winter) return k.terrain;
  if (k.winter !== "") return k.winter;
  if (terrains.has(k.terrain + "_w")) return k.terrain + "_w";
  return k.terrain;
}

function hexOf(k: Look, winter: boolean, terrains: ReadonlySet<string>): Hex {
  return {
    terrain: ground(k, winter, terrains),
    infra: { ids: [...k.infra], list: k.infra.length > 1 },
    unit: "",
    content: null,
  };
}

/** The level owners of sides 0, 1 and 2 (game.Player: Central Powers 1, Entente 2, none 0). */
const owners = [1, 2, 0] as const;
const playerNames = ["Central Powers", "Entente Cordial"] as const;

function fail(message: string): never {
  throw new GoError(`origview: ${message}`);
}

/** edgeOrder is the order in which an edge hex looks for a neighbour inside the map: N, NW, NE, then the others. */
const edgeOrder = [5, 4, 0, 1, 3, 2] as const;

function building(terrain: string): boolean {
  return (
    terrain.startsWith("hq") ||
    terrain.startsWith("cmp") ||
    terrain.startsWith("factory")
  );
}

/**
 * view is the level of an imported map (Go's origview.Level): mapName the map's stem, mapNo its number plus one
 * (-1 when the stem is none), player the first human side.
 */
export function view(
  l: finmap.Level,
  looks: Looks,
  terrains: ReadonlySet<string>,
): Level {
  const g: Grid = { width: l.width / 2, lines: 2 * l.height - 1 };
  if (
    l.width < 4 ||
    l.width % 2 !== 0 ||
    l.height < 2 ||
    l.terrain.length !== l.height
  )
    return fail(`a ${l.width}×${l.height} map`);
  const record = (c: Coord): number =>
    at(at(l.terrain, c.line >> 1), 2 * c.col + (c.line % 2));
  const look = (t: number): Look => {
    const k = looks.get(t);
    if (k === undefined) return fail(`no look for terrain record ${t}`);
    return k;
  };
  for (const [r, row] of l.terrain.entries()) {
    if (row.length !== l.width) return fail(`row ${r} has ${row.length} cells`);
    for (const [c, t] of row.entries())
      if (!looks.has(t))
        return fail(`cell ${r},${c}: no look for terrain record ${t}`);
  }
  const winter = l.mission?.winter ?? false;
  const stem = l.source.map;
  const lv: Level = {
    mapName: stem,
    mapNo: /^[+-]?\d+$/.test(stem) ? Number(stem) + 1 : -1,
    player: playerNames[0],
    lines: [],
    extra: null,
  };
  const human = l.mission?.humanSides[0];
  if (human !== undefined) {
    if (human !== finmap.central && human !== finmap.allied)
      return fail(`human side ${human}`);
    lv.player = playerNames[human];
  }
  const inside = (c: Coord): number => {
    const a = around(c);
    for (const i of edgeOrder) {
      const n = at(a, i);
      if (
        contains(g, n) &&
        record(n) !== edge &&
        !building(look(record(n)).terrain)
      )
        return record(n);
    }
    return edge;
  };
  for (let line = 0; line < g.lines; line++) {
    const hs: Hex[] = [];
    for (let col = 0; col < g.width - (line % 2); col++) {
      const c = { line, col };
      let t = record(c);
      if (t === edge) t = inside(c);
      hs.push(hexOf(look(t), winter, terrains));
    }
    lv.lines.push(hs);
  }
  for (const b of l.buildings) {
    const c = finmap.coord(b.cell[0], b.cell[1]);
    const o = owners[b.owner];
    if (!contains(g, c) || o === undefined)
      return fail(`${b.kind} at ${b.cell.join(",")} owned by ${b.owner}`);
    at(lv.lines, c.line)[c.col] = {
      ...at(at(lv.lines, c.line), c.col),
      content: { owner: o, units: null },
    };
  }
  return lv;
}

/**
 * campaign is the remake level of an imported map (Go's origview.Campaign): the view with the map's units by remake
 * id and the units the container records put into the buildings as the entrances' content, in record order. A unit
 * of a neutral record takes the first human side's remake unit, or the other side's.
 */
export function campaign(
  l: finmap.Level,
  data: BindingData,
  looks: Looks,
  terrains: ReadonlySet<string>,
): Level {
  const lv = view(l, looks, terrains);
  const ids = remakeIDs(data);
  const human = l.mission?.humanSides[0] ?? finmap.central;
  const id = (typ: number, side: number): string => {
    const sides = side === finmap.neutral ? [human, 1 - human] : [side];
    for (const s of sides) {
      const found = ids.get(`${typ}/${s}`);
      if (found !== undefined) return found;
    }
    return fail(
      `map ${l.source.map}: no remake unit for type ${typ} of side ${side}`,
    );
  };
  const g = levelGrid(lv);
  for (const u of l.units) {
    const c = finmap.coord(u.cell[0], u.cell[1]);
    if (!contains(g, c))
      return fail(
        `map ${l.source.map}: unit at ${u.cell.join(",")} off the level`,
      );
    const h = at(at(lv.lines, c.line), c.col);
    if (h.unit !== "" || h.content !== null)
      return fail(
        `map ${l.source.map}: unit at ${u.cell.join(",")} on an occupied hex or a building entrance`,
      );
    h.unit = id(u.type, u.side);
  }
  const entrances = new Map<number, Coord>();
  for (const b of l.buildings)
    entrances.set(b.container, finmap.coord(b.cell[0], b.cell[1]));
  for (const r of l.containers ?? []) {
    if (r.holder !== "building") continue;
    const c = entrances.get(r.container);
    if (c === undefined)
      return fail(
        `map ${l.source.map}: record for building ${r.container}, which the map lacks`,
      );
    const content = at(at(lv.lines, c.line), c.col).content;
    if (content === null) return fail(`no content at building ${r.container}`);
    for (const u of r.units) (content.units ??= []).push(id(u.type, r.side));
  }
  return lv;
}
