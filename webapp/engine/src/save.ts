// A saved game is the whole state of the command cycle as JSON: the profile by its tables' build and the season,
// both maps, the live and container records, each side's phase, confirmation and order list, the turn, the game
// options, the energy accounts and the buildings. Orders given but not confirmed and confirmed orders waiting
// for the other side are both part of that state (the order lists, the units' moved flags and targets), so a
// loaded game resolves exactly as the saved one would have.
//
// save writes the bytes Go's Game.Save writes (json.Marshal of savedGame, gojson.ts); loadGame reads them as
// Go's LoadGame does, with its checks in its order and its messages. Go's LoadGame knows the Full and Demo
// profiles; the port is given them (profilesOf). Go checks the units in its map's random order; the port in
// ascending order, which only decides which of several bad units a message names.

import { type Live, season } from "./board.ts";
import { markKinds } from "./capture.ts";
import { type Hold, slotFree } from "./container.ts";
import { type Game, type Player, action } from "./cycle.ts";
import type { Building } from "./energy.ts";
import { GoError } from "./goruntime.ts";
import {
  type Codec,
  encodeArray,
  encodeBool,
  encodeInt,
  encodeIntMap,
  encodeInts,
  encodeObject,
  encodeString,
  goArray,
  goBool,
  goInt,
  goIntMap,
  goPointer,
  goSlice,
  goString,
  goStruct,
  quote,
  unmarshal,
} from "./gojson.ts";
import type { Build, Profile } from "./profile.ts";
import type { End, EndReason } from "./victory.ts";

/** SaveSchema is the version of the saved command-cycle state. loadGame also reads schema 2, which had no season (summer). */
export const saveSchema = 3;

// savedPlayer is a Player as saved; orders null is Go's nil.
interface SavedPlayer {
  phase: number;
  ready: boolean;
  computer: boolean;
  orders: number[] | null;
}

// SavedBuilding is a Building as read; builds null is Go's nil.
interface SavedBuilding {
  credits: number;
  builds: number[] | null;
  computerBuilds: number[] | null;
}

// savedGame is the JSON of a save, Go's savedGame; null is Go's nil.
interface SavedGame {
  schema: number;
  build: string;
  winter: boolean;
  width: number;
  height: number;
  maps: (number[] | null)[];
  units: Map<number, Live> | null;
  holds: Map<number, Hold> | null;
  players: SavedPlayer[];
  turn: number;
  limit: number;
  concealed: boolean;
  ended: End | null;
  energy: number[];
  buildings: Map<number, SavedBuilding> | null;
}

// pair is a Go [2]T held as a TS pair.
function pair<T>(c: Codec<T[]>): Codec<[T, T]> {
  return c as unknown as Codec<[T, T]>;
}

const savedUnitCodec = goStruct<Live>("savedUnit", "dosrules.savedUnit", {
  budget: ["budget", goInt],
  exp: ["exp", goInt],
  size: ["size", goInt],
  alt: ["alt", goInt],
  flags4: ["flags4", goInt],
  flags6: ["flags6", goInt],
  type: ["type", goInt],
  serial: ["serial", goInt],
  hold: ["hold", goInt],
  pos: ["pos", pair(goArray(goInt, 2))],
  facing: ["facing", pair(goArray(goInt, 2))],
  target: ["target", goInt],
  action: ["action", goArray(goInt, 7)],
});

const holdCodec = goStruct<Hold>("Hold", "dosrules.Hold", {
  slots: ["slots", goArray(goInt, 14)],
  cells: ["cells", goArray(goInt, 4)],
  flags: ["flags", goInt],
  occupant: ["occupant", goInt],
});

const playerCodec = goStruct<SavedPlayer>(
  "savedPlayer",
  "dosrules.savedPlayer",
  {
    phase: ["phase", goInt],
    ready: ["ready", goBool],
    computer: ["computer", goBool],
    orders: ["orders", goSlice(goInt)],
  },
);

// The reasons are any string Go reads; the port's EndReason names the ones the rules give.
const endCodec = goStruct<End>("End", "dosrules.End", {
  over: ["over", goBool],
  winner: ["winner", goInt],
  won: [
    "won",
    pair(goArray(goString("dosrules.EndReason"), 2)) as Codec<
      [EndReason, EndReason]
    >,
  ],
  score: ["score", goInt],
});

const buildingCodec = goStruct<SavedBuilding>("Building", "dosrules.Building", {
  credits: ["credits", goInt],
  builds: ["builds", goSlice(goInt)],
  computerBuilds: ["computerBuilds", goSlice(goInt)],
});

const savedGameCodec = goStruct<SavedGame>("savedGame", "dosrules.savedGame", {
  schema: ["schema", goInt],
  build: ["build", goString("string")],
  winter: ["winter", goBool],
  width: ["width", goInt],
  height: ["height", goInt],
  maps: ["maps", goArray(goSlice(goInt), 2)],
  units: ["units", goIntMap(savedUnitCodec)],
  holds: ["holds", goIntMap(holdCodec)],
  players: ["players", goArray(playerCodec, 2)],
  turn: ["turn", goInt],
  limit: ["limit", goInt],
  concealed: ["concealed", goBool],
  ended: ["ended", goPointer(endCodec)],
  energy: ["energy", goArray(goInt, 2)],
  buildings: ["buildings", goIntMap(buildingCodec)],
});

const ints = (a: readonly number[]): string => encodeArray(a.map(encodeInt));

function encodeUnit(u: Live): string {
  return encodeObject([
    ["budget", encodeInt(u.budget)],
    ["exp", encodeInt(u.exp)],
    ["size", encodeInt(u.size)],
    ["alt", encodeInt(u.alt)],
    ["flags4", encodeInt(u.flags4)],
    ["flags6", encodeInt(u.flags6)],
    ["type", encodeInt(u.type)],
    ["serial", encodeInt(u.serial)],
    ["hold", encodeInt(u.hold)],
    ["pos", encodeInts(u.pos, 2)],
    ["facing", encodeInts(u.facing, 2)],
    ["target", encodeInt(u.target)],
    ["action", encodeInts(u.action, 7)],
  ]);
}

function encodeHold(h: Hold): string {
  return encodeObject([
    ["slots", encodeInts(h.slots, 14)],
    ["cells", encodeInts(h.cells, 4)],
    ["flags", encodeInt(h.flags)],
    ["occupant", encodeInt(h.occupant)],
  ]);
}

function encodePlayer(p: Player): string {
  const f: [string, string][] = [["phase", encodeInt(p.phase)]];
  if (p.ready) {
    f.push(["ready", encodeBool(true)]);
  }
  if (p.computer) {
    f.push(["computer", encodeBool(true)]);
  }
  if (p.orders.length > 0) {
    f.push(["orders", ints(p.orders)]);
  }
  return encodeObject(f);
}

function encodeEnd(e: End): string {
  return encodeObject([
    ["over", encodeBool(e.over)],
    ["winner", encodeInt(e.winner)],
    ["won", encodeArray([0, 1].map((k) => encodeString(e.won[k] ?? "")))],
    ["score", encodeInt(e.score)],
  ]);
}

function encodeBuilding(b: Building): string {
  const f: [string, string][] = [["credits", encodeInt(b.credits)]];
  if (b.builds.length > 0) {
    f.push(["builds", ints(b.builds)]);
  }
  if ((b.computerBuilds?.length ?? 0) > 0) {
    f.push(["computerBuilds", ints(b.computerBuilds ?? [])]);
  }
  return encodeObject(f);
}

/** save is Go's Game.Save: the game's state as JSON, the bytes Go writes. It throws a GoError without tables. */
export function save(g: Game): Uint8Array {
  const t = g.profile.tables;
  if (t === null) {
    throw new GoError("dosrules: save: no tables");
  }
  const f: [string, string][] = [
    ["schema", encodeInt(saveSchema)],
    ["build", encodeString(t.build)],
  ];
  if (g.profile.winter) {
    f.push(["winter", encodeBool(true)]);
  }
  f.push(
    ["width", encodeInt(g.width)],
    ["height", encodeInt(g.height)],
    ["maps", encodeArray(g.maps.map(ints))],
    ["units", encodeIntMap(g.units, encodeUnit)],
    ["holds", encodeIntMap(g.holds, encodeHold)],
    ["players", encodeArray(g.players.map(encodePlayer))],
    ["turn", encodeInt(g.turn)],
    ["limit", encodeInt(g.limit)],
  );
  if (g.concealed) {
    f.push(["concealed", encodeBool(true)]);
  }
  if (g.ended !== null) {
    f.push(["ended", encodeEnd(g.ended)]);
  }
  f.push(["energy", encodeInts(g.energy, 2)]);
  if (g.buildings.size > 0) {
    f.push(["buildings", encodeIntMap(g.buildings, encodeBuilding)]);
  }
  return new TextEncoder().encode(encodeObject(f));
}

// load wraps err as Go's fmt.Errorf("dosrules: load: %w", err).
function loadError(err: unknown): GoError {
  const msg = err instanceof Error ? err.message : String(err);
  return new GoError(`dosrules: load: ${msg}`, { cause: err });
}

/**
 * loadGame is Go's LoadGame: the game saved by save (data as bytes, or as a string, which is read as its UTF-8),
 * with the tables of the profile in profiles whose build the save names. It throws a GoError for a save of
 * another schema, of tables it does not know or a season they lack, or whose maps, units, order lists and phases
 * do not fit together. A map's occupant byte is a unit (up to slotFree) or a marker above it, such as inBuilding.
 */
export function loadGame(
  data: Uint8Array | string,
  profiles: Readonly<Record<Build, Profile>>,
): Game {
  const bytes =
    typeof data === "string" ? new TextEncoder().encode(data) : data;
  let s: SavedGame;
  try {
    s = unmarshal(bytes, savedGameCodec);
  } catch (e) {
    if (e instanceof GoError) {
      throw loadError(e);
    }
    throw e;
  }
  if (s.schema !== saveSchema && s.schema !== 2) {
    throw new GoError(
      `dosrules: load: schema ${String(s.schema)}, want ${String(saveSchema)}`,
    );
  }
  let p: Profile | null = null;
  for (const q of [profiles.full, profiles.demo]) {
    if (q.tables !== null && q.tables.build === s.build) {
      p = q;
    }
  }
  if (p === null) {
    throw new GoError(`dosrules: load: unknown tables ${quote(s.build)}`);
  }
  p = season(p, s.winter);
  const tables = p.tables;
  if (tables === null || tables.move === null) {
    throw new GoError(
      `dosrules: load: tables ${quote(s.build)} have no winter terrain`,
    );
  }
  if (s.width < 1 || s.height < 1) {
    throw new GoError(
      `dosrules: load: ${String(s.width)}\u00d7${String(s.height)} map`,
    );
  }
  const units = new Map<number, Live>();
  for (const i of [...(s.units?.keys() ?? [])].sort((a, b) => a - b)) {
    const u = s.units?.get(i);
    if (u === undefined) {
      continue;
    }
    if (i < 0 || i > slotFree || u.type < 0 || u.type >= tables.types.length) {
      throw new GoError(
        `dosrules: load: unit ${String(i)} of type ${String(u.type)}`,
      );
    }
    units.set(i, u);
  }
  const holds = new Map<number, Hold>();
  for (const r of [...(s.holds?.keys() ?? [])].sort((a, b) => a - b)) {
    const h = s.holds?.get(r);
    if (h !== undefined) {
      holds.set(r, h);
    }
  }
  const cells = s.width * s.height;
  const maps: [number[], number[]] = [s.maps[0] ?? [], s.maps[1] ?? []];
  for (const [k, m] of maps.entries()) {
    if (m.length !== 2 * cells) {
      throw new GoError(
        `dosrules: load: map ${String(k)} has ${String(m.length)} entries for ${String(cells)} cells`,
      );
    }
    for (let c = 0; c < cells; c++) {
      const occ = m[2 * c + 1] ?? 0;
      if (occ < 0 || occ > 0xff || (occ <= slotFree && !units.has(occ))) {
        throw new GoError(
          `dosrules: load: map ${String(k)} shows unit ${String(occ)} at ${String(c)}`,
        );
      }
    }
  }
  const players = s.players.map((sp, k): Player => {
    for (const u of sp.orders ?? []) {
      if (!units.has(u)) {
        throw new GoError(
          `dosrules: load: side ${String(k)} orders unit ${String(u)}`,
        );
      }
    }
    return {
      phase: sp.phase,
      ready: sp.ready,
      computer: sp.computer,
      orders: sp.orders ?? [],
    };
  });
  const buildings = new Map<number, Building>();
  for (const r of [...(s.buildings?.keys() ?? [])].sort((a, b) => a - b)) {
    const b = s.buildings?.get(r);
    if (b !== undefined) {
      buildings.set(r, {
        credits: b.credits,
        builds: b.builds ?? [],
        ...(b.computerBuilds === null
          ? {}
          : { computerBuilds: b.computerBuilds }),
      });
    }
  }
  const [p0, p1] = players;
  const [e0, e1] = s.energy;
  if (
    p0 === undefined ||
    p1 === undefined ||
    e0 === undefined ||
    e1 === undefined
  ) {
    throw new TypeError("dosrules: load: a Go array of two is not two"); // the codecs make them two
  }
  const g: Game = {
    profile: p,
    width: s.width,
    height: s.height,
    maps,
    units,
    holds,
    players: [p0, p1],
    turn: s.turn,
    limit: s.limit,
    concealed: s.concealed,
    ended: s.ended,
    energy: [e0, e1],
    buildings,
  };
  try {
    action(g);
  } catch (e) {
    if (e instanceof GoError) {
      throw loadError(e);
    }
    throw e;
  }
  markKinds(g);
  return g;
}
