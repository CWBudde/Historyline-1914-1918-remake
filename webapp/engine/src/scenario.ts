// Pure scenario-folder decoding. The caller supplies parsed files and the terrain catalogue.
import {
  DecodeError,
  array,
  arrayField,
  int,
  intField,
  object,
  stringField,
} from "./decode.ts";
import { type Coord, contains } from "./hexgrid.ts";
import { type Hex, type Level, coordKey, grid, parseInfra } from "./level.ts";

export const modes = ["cp", "ec", "2p"] as const;
export type Mode = (typeof modes)[number];
export const seasons = ["summer", "autumn", "spring", "winter"] as const;
export type Season = (typeof seasons)[number];
export interface Unit {
  at: Coord;
  id: string;
}
export interface Building {
  at: Coord;
  owner: number;
  credits: number;
  units: string[] | null;
}
export interface Setup {
  from: Mode | "";
  code: string;
  units: Unit[];
  buildings: Building[];
  builds: string[];
  /** The unit types the computer may build on this map; empty: its side's default. */
  computerBuilds: number[];
}
export interface File {
  month: string;
  name: string;
  season: Season;
  seasons: Map<string, string>;
  backdrop: string;
  map: Hex[][];
  setups: Map<Mode, Setup>;
}
function fail(message: string): never {
  throw new DecodeError(`scenario: ${message}`);
}
function str(v: unknown): string {
  if (typeof v !== "string") return fail("not a string");
  return v;
}
function ss(v: readonly unknown[] | null): string[] {
  return (v ?? []).map(str);
}
function mode(v: string): Mode {
  if (v !== "cp" && v !== "ec" && v !== "2p") return fail(`no mode ${v}`);
  return v;
}
function season(v: string): Season {
  if (v !== "summer" && v !== "autumn" && v !== "spring" && v !== "winter")
    return fail(`no season ${v}`);
  return v;
}
const winterForms: Readonly<Record<string, string>> = {
  hills: "rough_w",
  mud: "shelled_w",
  stone: "stoney_w",
};
export function kind(t: string): string {
  for (const [k, w] of Object.entries(winterForms)) if (w === t) return k;
  return t.endsWith("_w") ? t.slice(0, -2) : t;
}
export function inSeason(
  k: string,
  s: string,
  terrains: ReadonlySet<string>,
): string {
  if (s !== "winter") return k;
  const w = winterForms[k] ?? `${k}_w`;
  return terrains.has(w) ? w : k;
}
function asLevel(f: File): Level {
  return { mapName: f.name, mapNo: 0, player: "", lines: f.map, extra: null };
}
function position(f: File, a: readonly unknown[]): Coord {
  if (a.length !== 2) return fail("want [line, column]");
  const c = { line: int(a[0], "line") - 1, col: int(a[1], "column") - 1 };
  if (!contains(grid(asLevel(f)), c)) return fail("position off the map");
  return c;
}
function building(t: string): boolean {
  return /^(hq|cmp|factory)/.test(t);
}

export function decode(
  files: ReadonlyMap<string, unknown>,
  terrains: ReadonlySet<string>,
): File {
  const mj = object(files.get("map.json"), "map.json", [
    "format",
    "version",
    "name",
    "season",
    "map",
    "seasons",
  ]);
  if (mj["format"] !== "historyline-scenario" || mj["version"] !== 3)
    return fail("want historyline-scenario version 3");
  const f: File = {
    month: "",
    name: stringField(mj, "name", "map"),
    season: season(stringField(mj, "season", "map")),
    seasons: new Map(),
    backdrop: "",
    map: [],
    setups: new Map(),
  };
  if (f.name === "") return fail("no name");
  const lines = object(mj["map"], "map");
  if (Object.keys(lines).length === 0) return fail("no map");
  for (let i = 0; i < Object.keys(lines).length; i++) {
    f.map.push(
      array(lines[`line${i + 1}`], "map line").map((v): Hex => {
        const a = array(v, "map hex");
        if (a.length !== 2) return fail("want [terrain, infrastructure]");
        const t = str(a[0]);
        if (t === "" || kind(t) !== t || !terrains.has(t))
          return fail("terrain must have no season suffix");
        const terrain = inSeason(t, f.season, terrains);
        if (!terrains.has(terrain)) return fail(`unknown terrain ${terrain}`);
        return { terrain, infra: parseInfra(a[1]), unit: "", content: null };
      }),
    );
  }
  const g = grid(asLevel(f));
  for (const [line, hs] of f.map.entries())
    if (
      hs.length === 0 ||
      !contains(g, { line, col: hs.length - 1 }) ||
      contains(g, { line, col: hs.length })
    )
      return fail("map line does not fit width");
  if (mj["seasons"] != null) {
    const overrides = object(mj["seasons"], "seasons");
    const codes: Readonly<Record<string, Season>> = {
      s: "summer",
      a: "autumn",
      p: "spring",
      w: "winter",
    };
    for (const [key, value] of Object.entries(overrides)) {
      if (!/^line[1-9]\d*$/.test(key)) return fail("bad seasons line");
      const line = Number(key.slice(4)) - 1;
      const hs = f.map[line];
      const text = str(value);
      if (hs === undefined || text.length !== hs.length)
        return fail("seasons line does not fit map");
      for (let col = 0; col < text.length; col++) {
        const ch = text.charAt(col);
        if (ch === ".") continue;
        const s = codes[ch];
        if (s === undefined) return fail("bad season character");
        const h = hs[col];
        if (
          h === undefined ||
          !terrains.has(inSeason(kind(h.terrain), s, terrains))
        )
          return fail("unknown seasonal terrain");
        if (s !== f.season) f.seasons.set(coordKey({ line, col }), s);
      }
    }
  }
  for (const [name, raw] of files) {
    if (name === "map.json") continue;
    if (!/^(cp|ec|2p)\.json$/.test(name)) return fail(`unknown file ${name}`);
    const m = mode(name.slice(0, -5));
    const o = object(raw, "setup", [
      "from",
      "code",
      "units",
      "buildings",
      "builds",
      "computerBuilds",
    ]);
    const from = stringField(o, "from", "setup");
    const s: Setup = {
      from: from === "" ? "" : mode(from),
      code: stringField(o, "code", "setup"),
      units: [],
      buildings: [],
      builds: ss(arrayField(o, "builds", "setup")),
      computerBuilds: (arrayField(o, "computerBuilds", "setup") ?? []).map(
        (v) => int(v, "computerBuilds"),
      ),
    };
    const units = arrayField(o, "units", "setup") ?? [];
    const buildings = arrayField(o, "buildings", "setup") ?? [];
    if (
      from !== "" &&
      (units.length !== 0 ||
        buildings.length !== 0 ||
        s.builds.length !== 0 ||
        s.computerBuilds.length !== 0)
    )
      return fail("alias has its own content");
    const taken = new Set<string>();
    for (const v of units) {
      const a = array(v, "unit");
      if (a.length !== 3) return fail("want [line, column, id]");
      const at = position(f, a.slice(0, 2));
      const id = str(a[2]);
      if (id === "" || taken.has(coordKey(at)))
        return fail("empty or duplicate unit");
      taken.add(coordKey(at));
      s.units.push({ at, id });
    }
    taken.clear();
    for (const v of buildings) {
      const b = object(v, "building", ["at", "owner", "credits", "units"]);
      const at = position(f, array(b["at"], "building.at"));
      if (taken.has(coordKey(at))) return fail("duplicate building");
      taken.add(coordKey(at));
      const h = f.map[at.line]?.[at.col];
      if (h === undefined || !building(h.terrain))
        return fail("content outside a building");
      const owner = stringField(b, "owner", "building");
      if (owner !== "" && owner !== "cp" && owner !== "ec")
        return fail("unknown owner");
      const held = arrayField(b, "units", "building");
      s.buildings.push({
        at,
        owner: owner === "cp" ? 1 : owner === "ec" ? 2 : 0,
        credits: intField(b, "credits", "building"),
        units: held === null ? null : ss(held),
      });
    }
    f.setups.set(m, s);
  }
  if (f.setups.size === 0) return fail("no setups");
  const codes = new Set<string>();
  for (const s of f.setups.values()) {
    if (s.from !== "") {
      const source = f.setups.get(s.from);
      if (source === undefined || source.from !== "")
        return fail("missing alias source or alias chain");
    }
    if (s.code !== "") {
      if (codes.has(s.code)) return fail("shared code");
      codes.add(s.code);
    }
  }
  return f;
}

export function setup(f: File, m: Mode): Setup | null {
  const s = f.setups.get(m);
  if (s === undefined) return null;
  if (s.from === "") return s;
  const source = f.setups.get(s.from);
  if (source === undefined) return fail("missing alias source");
  return { ...source, from: s.from, code: s.code };
}
export function level(f: File, m: Mode): Level {
  const s = setup(f, m);
  if (s === null) return fail(`no ${m} setup`);
  const l: Level = {
    mapName: s.code || f.name,
    mapNo: 0,
    player: m === "ec" ? "Entente" : "Central Powers",
    lines: f.map.map((line) =>
      line.map((h) => ({ ...h, infra: { ...h.infra, ids: [...h.infra.ids] } })),
    ),
    extra: {
      credits: new Map(),
      builds: [...s.builds],
      computerBuilds: [...s.computerBuilds],
      backdrop: f.backdrop,
      season: f.season,
      seasons: new Map(f.seasons),
    },
  };
  for (const b of s.buildings) {
    const h = l.lines[b.at.line]?.[b.at.col];
    if (h === undefined) return fail("building off the map");
    h.content = {
      owner: b.owner,
      units: b.units === null ? null : [...b.units],
    };
    if (b.credits !== 0) l.extra?.credits.set(coordKey(b.at), b.credits);
  }
  for (const u of s.units) {
    const h = l.lines[u.at.line]?.[u.at.col];
    if (h === undefined) return fail("unit off the map");
    if (h.content !== null) return fail("unit on building content");
    h.unit = u.id;
  }
  return l;
}
