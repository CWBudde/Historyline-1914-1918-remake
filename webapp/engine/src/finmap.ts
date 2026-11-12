// internal/finmap: the original DOS level files, decoded and assembled into the versioned level JSON.
//
// A level is up to four files: MAP/NN.FIN (the map), MAP/NN.SHP (building and transporter inventories, energy and
// the types factories may produce), MAP/NN.COM (per-type parameters of the computer player, single-player maps
// only) and the map's 10-byte record in CODES.DAT (password, players, season, successor). All of them are
// TPWM-packed on disk; the decoders here take the unpacked bytes. Sides are 0 (Central Powers), 1 (Allies) and 2
// (neutral); cells are [row, col] in rows of width cells, odd columns half a cell lower.
import { GoError, at } from "./goruntime.ts";
import { type Coord } from "./hexgrid.ts";
import { type MoveTables } from "./move.ts";

/** typeCount is the number of unit types (UNIT.DAT records). */
export const typeCount = 51;
export const central = 0,
  allied = 1,
  neutral = 2;
/** emptyFrom is the lowest occupant value that holds no unit. */
export const emptyFrom = 0xf4;
/** slotEmpty is the lowest SHP slot value that creates no unit. */
export const slotEmpty = 0xf1;
/** maxUnits is the number of live unit records. */
export const maxUnits = 0xf1;

// Type class and terrain flag bits.
export const classTransporter = 0x1000,
  classTwoCell = 0x40,
  classHead = 0x80,
  classNeutral = 0x02,
  flagHQ = 0x40,
  flagDepot = 0x100,
  flagFactory = 0x400;

// Container table layout: 70 transporter holds, then the two HQs, ten depots and ten factories.
export const transporters = 70,
  hqBase = 70,
  depotBase = 72,
  factoryBase = 82,
  containers = 92,
  tableBuilding = 10;

/** The record kinds of an SHP file: 0 HQ, 1 factory, 2 depot, 3 transporter. */
export const kindNames = ["hq", "factory", "depot", "transporter"] as const;

export const format = "historyline-level",
  version = 1;

function fail(message: string): never {
  throw new GoError(`finmap: ${message}`);
}

/** FinMap is a decoded FIN file: the terrain record and occupant of every cell, row-major. */
export interface FinMap {
  width: number;
  height: number;
  terrain: Uint8Array;
  occupant: Uint8Array;
}

/** decodeMap decodes an unpacked FIN file: two big-endian size words, then two bytes per cell. */
export function decodeMap(b: Uint8Array): FinMap {
  if (b.length < 4) return fail("FIN header truncated");
  if (at(b, 0) !== 0 || at(b, 2) !== 0)
    return fail("FIN size words exceed a byte");
  const width = at(b, 1),
    height = at(b, 3);
  if (width === 0 || height === 0) return fail(`FIN size ${width}x${height}`);
  const n = width * height;
  if (b.length !== 4 + 2 * n)
    return fail(
      `FIN has ${b.length} bytes, ${width}x${height} needs ${4 + 2 * n}`,
    );
  const terrain = new Uint8Array(n),
    occupant = new Uint8Array(n);
  for (let i = 0; i < n; i++) {
    terrain[i] = at(b, 4 + 2 * i);
    occupant[i] = at(b, 5 + 2 * i);
  }
  return { width, height, terrain, occupant };
}

/** coord is the hexgrid coordinate of cell [row, col] (Go's finmap.Coord). */
export function coord(row: number, col: number): Coord {
  return { line: 2 * row + (col % 2), col: Math.trunc(col / 2) };
}

/** ShopRecord is one 12-byte SHP record. */
export interface ShopRecord {
  side: number;
  kind: number;
  index: number;
  energy: number;
  computerValue: number;
  slots: number[];
}
export interface Shop {
  locked: boolean[];
  records: ShopRecord[];
}

function recordUnits(r: ShopRecord): number[] {
  return r.slots.filter((s) => s < slotEmpty);
}
function recordContainer(r: ShopRecord): number {
  switch (r.kind) {
    case 0:
      return hqBase + r.index;
    case 1:
      return factoryBase + r.index;
    case 2:
      return depotBase + r.index;
  }
  return r.index;
}
function kindName(k: number): string {
  return kindNames[k] ?? `kind${k}`;
}

/** decodeShop decodes an unpacked SHP file: a lock byte per type, the record count, then the records. */
export function decodeShop(b: Uint8Array): Shop {
  if (b.length < typeCount + 1) return fail("SHP truncated");
  const n = at(b, typeCount);
  if (b.length !== typeCount + 1 + 12 * n)
    return fail(`SHP has ${b.length} bytes for ${n} records`);
  const s: Shop = { locked: [], records: [] };
  for (let t = 0; t < typeCount; t++) {
    if (at(b, t) > 1) return fail(`SHP type ${t} flag ${at(b, t)}`);
    s.locked.push(at(b, t) === 1);
  }
  const limit = [
    containers - hqBase,
    tableBuilding,
    tableBuilding,
    transporters,
  ];
  for (let i = 0; i < n; i++) {
    const p = typeCount + 1 + 12 * i;
    const r: ShopRecord = {
      side: at(b, p),
      kind: at(b, p + 1),
      index: at(b, p + 2),
      energy: at(b, p + 3),
      computerValue: at(b, p + 4),
      slots: Array.from(b.subarray(p + 5, p + 12)),
    };
    if (r.side > neutral) return fail(`SHP record ${i} side ${r.side}`);
    const l = limit[r.kind];
    if (l === undefined) return fail(`SHP record ${i}: kind ${r.kind}`);
    if (r.index >= l)
      return fail(`SHP record ${i}: ${kindName(r.kind)} index ${r.index}`);
    for (const t of recordUnits(r))
      if (t >= typeCount) return fail(`SHP record ${i} holds type ${t}`);
    s.records.push(r);
  }
  return s;
}

/** Computer is a decoded COM file: six bytes per unit type that only the computer player reads. */
export type Computer = Uint8Array[];

/** decodeComputer decodes an unpacked COM file. */
export function decodeComputer(b: Uint8Array): Computer {
  if (b.length !== 6 * typeCount)
    return fail(`COM has ${b.length} bytes, want ${6 * typeCount}`);
  return Array.from({ length: typeCount }, (_, t) => b.slice(6 * t, 6 * t + 6));
}

/** computerBuilds are the types the computer player may build: bit 0x400 of the word at +2. */
export function computerBuilds(c: Computer): number[] {
  const out: number[] = [];
  for (const [t, r] of c.entries()) if ((at(r, 3) & 0x04) !== 0) out.push(t);
  return out;
}

/** Code is a map's 10-byte record of CODES.DAT. */
export interface Code {
  password: string;
  players: number;
  step: number;
  winter: boolean;
  campaignEnd: boolean;
}

/** decodeCodes decodes an unpacked CODES.DAT. */
export function decodeCodes(b: Uint8Array): Code[] {
  if (b.length % 10 !== 0)
    return fail(`CODES.DAT has ${b.length} bytes, not whole records`);
  const out: Code[] = [];
  for (let i = 0; i < b.length / 10; i++) {
    const p = b.subarray(10 * i, 10 * i + 10);
    let password = "";
    for (const c of p.subarray(0, 5)) {
      if (c < 0x11 || c > 0x11 + 25)
        return fail(`CODES.DAT record ${i} letter ${c}`);
      password += String.fromCharCode(0x41 + c - 0x11);
    }
    if (at(p, 6) !== 1 && at(p, 6) !== 2)
      return fail(`CODES.DAT record ${i} players ${at(p, 6)}`);
    if ((at(p, 8) & ~3) !== 0)
      return fail(`CODES.DAT record ${i} flags ${at(p, 8)}`);
    out.push({
      password,
      players: at(p, 6),
      step: at(p, 7),
      winter: (at(p, 8) & 2) !== 0,
      campaignEnd: (at(p, 8) & 1) !== 0,
    });
  }
  return out;
}

/** Rules are the fields of a build's tables that loading a map reads. */
export interface Rules {
  terrainFlags: number[];
  typeClass: number[];
}
export function rulesFrom(t: MoveTables): Rules {
  return {
    terrainFlags: t.terrains.map((r) => r.flags),
    typeClass: t.types.map((r) => r.class),
  };
}

/** Input are the decoded files of one map. */
export interface Input {
  index: number;
  map: FinMap;
  shop?: Shop;
  computer?: Computer;
  code?: Code;
}

// The level JSON of Go's finmap.Level, with its JSON field names; optional fields are absent where Go omits them.
export interface Source {
  build: string;
  map: string;
  files: string[];
}
export interface Mission {
  code: string;
  players: number;
  humanSides: number[];
  winter: boolean;
  next: number | null;
  campaignEnd: boolean;
}
export interface Unit {
  cell: [number, number];
  type: number;
  side: number;
  rear?: [number, number];
  container?: number;
}
export interface Building {
  kind: string;
  cell: [number, number];
  owner: number;
  container: number;
}
export interface Cargo {
  type: number;
  container?: number;
}
export interface Container {
  container: number;
  kind: string;
  index: number;
  side: number;
  energy: number;
  computerValue: number;
  units: Cargo[];
  overridden?: boolean;
  holder: string;
}
export interface Level {
  format: string;
  version: number;
  source: Source;
  mission?: Mission;
  width: number;
  height: number;
  terrain: number[][];
  units: Unit[];
  buildings: Building[];
  containers: Container[] | null;
  producible: number[] | null;
  computer: boolean;
  computerBuilds?: number[];
}

function ownerOf(flags: number): number {
  if ((flags & 2) !== 0) return neutral;
  if ((flags & 1) !== 0) return allied;
  return central;
}

/** build assembles the level JSON from a map's decoded files (Go's finmap.Build). */
export function build(input: Input, r: Rules): Level {
  const m = input.map;
  if (
    m.width < 1 ||
    m.height < 1 ||
    m.terrain.length !== m.width * m.height ||
    m.occupant.length !== m.width * m.height
  )
    return fail(`${m.width}×${m.height} map with mismatched cells`);
  if (input.code !== undefined && input.index < 0)
    return fail("a CODES.DAT record without a map number");
  const l: Level = {
    format,
    version,
    source: { build: "", map: "", files: [] },
    width: m.width,
    height: m.height,
    terrain: [],
    units: [],
    buildings: [],
    containers: null,
    producible: null,
    computer: false,
  };
  for (let row = 0; row < m.height; row++) {
    const line: number[] = [];
    for (let col = 0; col < m.width; col++) {
      const t = at(m.terrain, row * m.width + col);
      if (t >= r.terrainFlags.length)
        return fail(`cell ${row},${col} terrain ${t}`);
      line.push(t);
    }
    l.terrain.push(line);
  }
  addBuildings(l, m, r);
  const units = addUnits(l, m, r);
  let created = units.created;
  if (input.shop !== undefined) addShop(l, input.shop, r, units.holds);
  for (const c of l.containers ?? []) created += c.units.length;
  if (created > maxUnits) return fail(`${created} units, at most ${maxUnits}`);
  l.computer = input.computer !== undefined;
  if (input.computer !== undefined) {
    const builds = computerBuilds(input.computer);
    if (builds.length > 0) l.computerBuilds = builds;
  }
  const c = input.code;
  if (c !== undefined) {
    l.mission = {
      code: c.password,
      players: c.players,
      humanSides:
        c.players === 2
          ? [central, allied]
          : input.index < 24
            ? [central]
            : [allied],
      winter: c.winter,
      next: c.step === 0 ? null : input.index + c.step,
      campaignEnd: c.campaignEnd,
    };
  }
  return l;
}

function addBuildings(l: Level, m: FinMap, r: Rules): void {
  let depots = 0,
    factories = 0;
  const hqs = new Set<number>();
  for (let row = 0; row < m.height; row++)
    for (let col = 0; col < m.width; col++) {
      const f = at(r.terrainFlags, at(m.terrain, row * m.width + col));
      const owner = ownerOf(f);
      let kind: string, container: number;
      if ((f & flagFactory) !== 0) {
        kind = "factory";
        container = factoryBase + factories++;
      } else if ((f & flagDepot) !== 0) {
        kind = "depot";
        container = depotBase + depots++;
      } else if ((f & flagHQ) !== 0) {
        if (owner === neutral || hqs.has(owner))
          return fail(
            `cell ${row},${col}: a second or neutral HQ (owner ${owner})`,
          );
        hqs.add(owner);
        kind = "hq";
        container = hqBase + owner;
      } else continue;
      l.buildings.push({ kind, cell: [row, col], owner, container });
    }
  if (depots > tableBuilding || factories > tableBuilding)
    fail(
      `${depots} depots and ${factories} factories, the tables hold ${tableBuilding} each`,
    );
}

function addUnits(
  l: Level,
  m: FinMap,
  r: Rules,
): { created: number; holds: number } {
  const occupantAt = (row: number, col: number): number =>
    at(m.occupant, row * m.width + col);
  const rears = new Set<number>();
  let holds = 0;
  for (let row = 0; row < m.height; row++)
    for (let col = 0; col < m.width; col++) {
      const occ = occupantAt(row, col);
      if (occ >= emptyFrom) continue;
      const t = occ >> 1,
        side = occ & 1;
      if (t >= typeCount || t >= r.typeClass.length)
        return fail(`cell ${row},${col} occupant ${occ}`);
      const cls = at(r.typeClass, t);
      if ((cls & classTwoCell) !== 0 && (cls & classHead) === 0) continue;
      const u: Unit = { cell: [row, col], type: t, side };
      if ((cls & classTwoCell) !== 0) {
        if (
          t + 1 >= r.typeClass.length ||
          (at(r.typeClass, t + 1) & (classTwoCell | classHead)) !== classTwoCell
        )
          return fail(`ship at ${row},${col}: type ${t + 1} is no rear`);
        const rr = side === allied ? row + 1 : row - 1;
        if (
          rr < 0 ||
          rr >= m.height ||
          occupantAt(rr, col) !== 2 * (t + 1) + side
        )
          return fail(`ship at ${row},${col} has no rear at ${rr},${col}`);
        u.rear = [rr, col];
        rears.add(rr * m.width + col);
      } else if ((cls & classNeutral) !== 0) u.side = neutral;
      if ((cls & classTransporter) !== 0) {
        if (holds === transporters)
          return fail(`more than ${transporters} transporters`);
        u.container = holds++;
      }
      l.units.push(u);
    }
  for (let row = 0; row < m.height; row++)
    for (let col = 0; col < m.width; col++) {
      const occ = occupantAt(row, col);
      if (occ >= emptyFrom) continue;
      const c = at(r.typeClass, occ >> 1);
      if (
        (c & classTwoCell) !== 0 &&
        (c & classHead) === 0 &&
        !rears.has(row * m.width + col)
      )
        return fail(`ship rear at ${row},${col} without a head`);
    }
  return { created: l.units.length + rears.size, holds };
}

function addShop(l: Level, s: Shop, r: Rules, holds: number): number {
  const out: Container[] = [];
  l.producible = [];
  for (const [t, locked] of s.locked.entries())
    if (!locked) l.producible.push(t);
  const holder = new Map<number, string>();
  for (const b of l.buildings) holder.set(b.container, "building");
  for (const u of l.units)
    if (u.container !== undefined) holder.set(u.container, "unit");
  for (const rec of s.records) {
    const c: Container = {
      container: recordContainer(rec),
      kind: kindName(rec.kind),
      index: rec.index,
      side: rec.side,
      energy: rec.energy,
      computerValue: rec.computerValue,
      units: [],
      holder: "",
    };
    for (const t of recordUnits(rec)) {
      const cargo: Cargo = { type: t };
      if (
        t < r.typeClass.length &&
        (at(r.typeClass, t) & classTransporter) !== 0
      ) {
        if (holds === transporters)
          return fail(`more than ${transporters} transporters`);
        cargo.container = holds;
        holder.set(holds, "cargo");
        holds++;
      }
      c.units.push(cargo);
    }
    out.push(c);
  }
  const last = new Map<number, number>();
  for (const [i, c] of out.entries()) last.set(c.container, i);
  for (const [i, c] of out.entries()) {
    if (last.get(c.container) !== i) c.overridden = true;
    c.holder = holder.get(c.container) ?? "none";
  }
  l.containers = out;
  return holds;
}
