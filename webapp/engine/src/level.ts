// The legacy level format (internal/level). File access belongs to the caller.
import { DecodeError, array, int, object } from "./decode.ts";
import { parse as parseJSON, goInt, type JSONNode } from "./gojson.ts";
import { type Coord, type Grid, contains } from "./hexgrid.ts";

export interface Infra {
  ids: string[];
  list: boolean;
}
export interface Content {
  owner: number;
  units: string[] | null;
}
export interface Hex {
  terrain: string;
  infra: Infra;
  unit: string;
  content: Content | null;
}
export interface Extra {
  credits: Map<string, number>;
  builds: string[];
  computerBuilds?: number[];
  backdrop: string;
  season: string;
  seasons: Map<string, string>;
  /** The scenario's campaign month ("1915-04"), which dates the extension units' rules; absent or "" for none. */
  month?: string;
}
export interface Level {
  mapName: string;
  mapNo: number;
  player: string;
  lines: Hex[][];
  extra: Extra | null;
}

/** Coordinates in level metadata use a value key rather than object identity. */
export function coordKey(c: Coord): string {
  return `${c.line},${c.col}`;
}
export function seasonAt(e: Extra, c: Coord): string {
  return e.seasons.get(coordKey(c)) ?? e.season;
}
export function grid(l: Level): Grid {
  return { width: l.lines[0]?.length ?? 0, lines: l.lines.length };
}

function str(v: unknown, what: string): string {
  if (typeof v !== "string") throw new DecodeError(`${what}: not a string`);
  return v;
}
function strings(v: unknown, what: string): string[] {
  return array(v, what).map((s) => str(s, what));
}
export function parseInfra(v: unknown): Infra {
  if (Array.isArray(v)) return { ids: strings(v, "infra"), list: true };
  const s = str(v, "infra");
  return { ids: s === "" ? [] : [s], list: false };
}
function parseContent(v: unknown): Content {
  const a = array(v, "content");
  if (a.length !== 1 && a.length !== 2)
    throw new DecodeError("content: want [owner] or [owner, [units]]");
  return {
    owner: int(a[0] ?? 0, "owner"),
    units: a.length === 2 ? strings(a[1], "units") : null,
  };
}

export function parse(raw: string | Uint8Array): Level {
  // Every number in this format is a Go int, including its JSON literal syntax.
  const value = (n: JSONNode): unknown => {
    switch (n.kind) {
      case "null":
        return null;
      case "bool":
      case "string":
        return n.value;
      case "number":
        return goInt.decode(n, 0, { struct: "", fields: [] });
      case "array":
        return n.items.map(value);
      case "object":
        return Object.fromEntries(n.entries.map(([k, v]) => [k, value(v)]));
    }
  };
  const input = value(
    parseJSON(typeof raw === "string" ? new TextEncoder().encode(raw) : raw),
  );
  const o = object(input, "level", ["mapName", "mapNo", "player", "tiles"]);
  let mapNo = o["mapNo"];
  if (typeof mapNo === "string") {
    if (!/^[+-]?\d+$/.test(mapNo))
      throw new DecodeError("mapNo is not a number");
    mapNo = Number(mapNo);
  }
  const tiles = object(o["tiles"], "tiles");
  if (Object.keys(tiles).length === 0) throw new DecodeError("no tiles");
  const l: Level = {
    mapName: str(o["mapName"], "mapName"),
    mapNo: int(mapNo, "mapNo"),
    player: str(o["player"], "player"),
    lines: [],
    extra: null,
  };
  for (let i = 0; i < Object.keys(tiles).length; i++) {
    const key = `line${i + 1}`;
    l.lines.push(
      array(tiles[key], key).map((v): Hex => {
        const h = array(v, key);
        if (h.length !== 3 && h.length !== 4)
          throw new DecodeError(`${key}: want 3 or 4 elements`);
        return {
          terrain: str(h[0], "terrain"),
          infra: parseInfra(h[1]),
          unit: str(h[2], "unit"),
          content: h.length === 4 ? parseContent(h[3]) : null,
        };
      }),
    );
  }
  const g = grid(l);
  for (const [line, hs] of l.lines.entries()) {
    if (
      hs.length === 0 ||
      !contains(g, { line, col: hs.length - 1 }) ||
      contains(g, { line, col: hs.length })
    )
      throw new DecodeError(
        `line${line + 1} has ${hs.length} hexes, which does not fit width ${g.width}`,
      );
  }
  return l;
}

/** marshal reproduces Go Level.Marshal's legacy whitespace and string restrictions. */
export function marshal(l: Level): string {
  const quote = (s: string): string => {
    for (const c of s) {
      if (
        c === '"' ||
        c === "'" ||
        c === "\\" ||
        c.charCodeAt(0) < 32 ||
        c.charCodeAt(0) === 127
      )
        throw new DecodeError("string cannot be written in the level format");
    }
    return `"${s}"`;
  };
  const ss = (a: string[]): string => `[${a.map(quote).join(", ")}]`;
  const hex = (h: Hex): string => {
    let infra: string;
    if (h.infra.list) infra = ss(h.infra.ids);
    else {
      if (h.infra.ids.length > 1)
        throw new DecodeError("infra needs the list form");
      infra = quote(h.infra.ids[0] ?? "");
    }
    let s = `[${quote(h.terrain)}, ${infra}, ${quote(h.unit)}`;
    if (h.content !== null)
      s += `, [${h.content.owner}${h.content.units === null ? "" : `, ${ss(h.content.units)}`}]`;
    return s + "]";
  };
  const lines = l.lines
    .map((line, i) => `\t\t\t\t"line${i + 1}" :\t[${line.map(hex).join(", ")}]`)
    .join(",\n");
  return `{\n\t"mapName" :\t${quote(l.mapName)},\n\t"mapNo" :\t${l.mapNo},\n\t"player" :\t${quote(l.player)},\n\t"tiles" :\t{\n${lines === "" ? "" : lines + "\n"}\t\t\t\t}\n}\n\n\n\n\n\n\n`;
}
