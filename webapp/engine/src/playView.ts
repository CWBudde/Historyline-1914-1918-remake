import * as rules from "./cycle.ts";
import { destroyed, liveCell } from "./board.ts";
import { at } from "./goruntime.ts";
import { coord as boundCoord, cell } from "./bind.ts";
import { holdFlag40, holdFlag100, holdFlag400, slotFree } from "./container.ts";
import { coordKey } from "./level.ts";
import type { Coord } from "./hexgrid.ts";
import { tablesOf, type Tables } from "./profile.ts";
import { working } from "./sapper.ts";
import {
  type Session,
  type UnitView,
  type View,
  type BuildingView,
  type SlotView,
  Infantry,
  Armour,
  Artillery,
  Naval,
  Air,
  HQ,
  Depot,
  Factory,
  faction,
  flagSide,
  hex,
} from "./playModel.ts";

export function typeDomain(t: Tables, typ: number): number {
  const rt = t.types[typ];
  if (rt === undefined) return Infantry;
  const mt = t.move?.types[typ],
    cls = mt?.class ?? rt.attackMask,
    flags = mt?.flags ?? 0;
  return cls & 0x10
    ? Air
    : cls & 8
      ? Naval
      : flags & 1
        ? Infantry
        : rt.surfaceRange - 1 >= 2 || rt.airRange - 1 >= 2
          ? Artillery
          : Armour;
}
export function domainOf(s: Session, idx: number): number {
  const u = s.game.units.get(idx);
  return u === undefined
    ? Infantry
    : typeDomain(tablesOf(s.game.profile), u.type);
}
export function kindOf(hold: number): number {
  return hold < holdFlag100 ? HQ : hold < holdFlag400 ? Depot : Factory;
}
export function typeID(s: Session, typ: number, side: number): string {
  for (const id of [...s.data.units.keys()].sort()) {
    const d = s.data.units.get(id);
    if (d === undefined) continue;
    const faction = ["Germany", "Austria", "Bulgaria", "Ottoman"].includes(
      d.country,
    )
      ? 0
      : 1;
    if (d.dos === typ && (side === 2 || side === faction)) return id;
  }
  return "";
}
export function typeName(s: Session, id: string, typ: number): string {
  return id === "" ? `Type ${typ}` : (s.data.units.get(id)?.name ?? "");
}
export function unitID(s: Session, idx: number): string {
  const u = rules.liveOf(s.game, idx),
    original = s.board.ids.get(idx),
    cached = s.ids.get(idx);
  if (original !== undefined && s.data.units.get(original)?.dos === u.type)
    return original;
  if (cached !== undefined && s.data.units.get(cached)?.dos === u.type)
    return cached;
  if ((u.flags4 & 0xc0) === 0x40) return unitID(s, idx - 1);
  const id = typeID(s, u.type, flagSide(u.flags4));
  s.ids.set(idx, id);
  return id;
}
export function container(
  s: Session,
  side: number,
  unit: number,
): { hold: number; slot: number } | null {
  for (const i of rules.sortedKeys(s.game.holds)) {
    const h = rules.holdOf(s.game, i);
    for (let k = 0; k < 7; k++)
      if (h.slots[7 * side + k] === unit) return { hold: i, slot: k };
  }
  return null;
}
export function startCell(h: { cells: number[]; flags: number }): number {
  return Math.trunc(((h.cells[0] ?? 0) - ((h.flags & 0x20) !== 0 ? 1 : 0)) / 2);
}
export function unitView(s: Session, idx: number, side: number): UnitView {
  const u = rules.liveOf(s.game, idx),
    id = unitID(s, idx),
    t = at(tablesOf(s.game.profile).types, u.type),
    c = container(s, side, idx);
  const held = (u.flags4 & 0x4000) !== 0,
    owner = flagSide(u.flags4);
  return {
    index: idx,
    type: u.type,
    id,
    name: typeName(s, id, u.type),
    domain: typeDomain(tablesOf(s.game.profile), u.type),
    side: owner,
    owner: faction(owner),
    at: hex(s, Math.trunc(at(u.pos, side) / 2)),
    size: u.size,
    full: t.group,
    alt: u.alt,
    exp: u.exp,
    budget: u.budget,
    armour: t.armour,
    moved: (u.flags4 & 0x200) !== 0,
    held,
    carrier: (u.flags4 & 0x1000) !== 0,
    destroyed: destroyed(u),
    hold: (u.flags4 & 0x1000) !== 0 ? u.hold : -1,
    in: held ? (c?.hold ?? -1) : -1,
    slot: held ? (c?.slot ?? -1) : -1,
    target: null,
    rear: null,
    head: null,
  };
}
export function unit(s: Session, idx: number): UnitView {
  const u = s.game.units.get(idx);
  if (u !== undefined) return unitView(s, idx, flagSide(u.flags4) & 1);
  return {
    index: idx,
    type: 0,
    id: "",
    name: "",
    domain: 0,
    side: 0,
    owner: 0,
    at: { line: 0, col: 0 },
    size: 0,
    full: 0,
    alt: 0,
    exp: 0,
    budget: 0,
    armour: 0,
    moved: false,
    held: false,
    carrier: false,
    destroyed: true,
    hold: -1,
    in: -1,
    slot: -1,
    target: null,
    rear: null,
    head: null,
  };
}
export function target(s: Session, side: number, idx: number): Coord | null {
  if (
    (side !== 0 && side !== 1) ||
    at(s.game.players, side).phase !== rules.phaseAction ||
    (!at(s.game.players, side).orders.includes(idx) &&
      !working(rules.liveOf(s.game, idx)))
  )
    return null;
  return hex(s, Math.trunc(rules.liveOf(s.game, idx).target / 2));
}
export const Slots = 7;
export function slotViews(s: Session, hold: number, side: number): SlotView[] {
  const h = rules.holdOf(s.game, hold);
  return Array.from({ length: Slots }, (_, slot) => {
    const u = h.slots[Slots * side + slot] ?? 0;
    if (u > slotFree || !s.game.units.has(u)) return { slot, unit: null };
    return {
      slot,
      unit: {
        ...unitView(s, u, side),
        at: hex(s, startCell(h)),
        held: true,
        in: hold,
        slot,
      },
    };
  });
}
export function slots(s: Session, hold: number, side: number): UnitView[] {
  return slotViews(s, hold, side).flatMap((v) =>
    v.unit === null ? [] : [v.unit],
  );
}
export function buildings(s: Session, side: number): Map<string, BuildingView> {
  const out = new Map<string, BuildingView>();
  for (const i of rules.sortedKeys(s.game.holds)) {
    if (i < holdFlag40) continue;
    const h = rules.holdOf(s.game, i),
      p = boundCoord(s.board, Math.trunc((h.cells[0] ?? 0) / 2));
    if (p === null) continue;
    const owner = flagSide(h.flags),
      k = side < 0 ? owner & 1 : side;
    out.set(coordKey(p), {
      hold: i,
      owner: faction(owner),
      kind: kindOf(i),
      hq: kindOf(i) === HQ,
      units: slots(s, i, k),
    });
  }
  return out;
}
function linkShips(s: Session, units: Map<string, UnitView>): void {
  const positions = new Map([...units].map(([c, u]) => [u.index, c]));
  for (const [key, u] of units) {
    const f = rules.liveOf(s.game, u.index).flags4;
    if ((f & 0x40) === 0 || (f & 0x80) !== 0) continue;
    const at = positions.get(u.index - 1),
      head = at === undefined ? undefined : units.get(at);
    if (head === undefined) continue;
    head.rear = u.at;
    u.head = head.at;
    units.set(key, u);
  }
}
export function view(s: Session, side: number): View {
  const v: View = { side, units: new Map(), buildings: buildings(s, side) },
    m = at(s.game.maps, side);
  for (let c = 0; c < s.game.width * s.game.height; c++) {
    const occ = m[2 * c + 1] ?? 0,
      p = boundCoord(s.board, c);
    if (occ > slotFree || !s.game.units.has(occ) || p === null) continue;
    const u = unitView(s, occ, side);
    u.at = p;
    if (u.side === side) u.target = target(s, side, occ);
    v.units.set(coordKey(p), u);
  }
  linkShips(s, v.units);
  return v;
}
export function trueView(s: Session): View {
  const v: View = { side: -1, units: new Map(), buildings: buildings(s, -1) };
  for (const i of rules.sortedKeys(s.game.units)) {
    const u = rules.liveOf(s.game, i);
    if (destroyed(u) || (u.flags4 & 0x4000) !== 0) continue;
    const side = flagSide(u.flags4) & 1;
    if (boundCoord(s.board, liveCell(u, side)) === null) continue;
    const uv = unitView(s, i, side);
    uv.target = target(s, side, i);
    v.units.set(coordKey(uv.at), uv);
  }
  linkShips(s, v.units);
  return v;
}
export function entrance(s: Session, c: Coord): Coord | null {
  const pos = cell(s.board, c);
  if (pos === null) return null;
  const part = s.board.parts.get(pos);
  if (part !== undefined) return hex(s, part);
  for (const [i, h] of s.game.holds)
    if (i >= holdFlag40 && h.cells[0] === 2 * pos) return c;
  for (const [i, h] of s.game.holds)
    if (i >= holdFlag40 && h.cells.slice(1).includes(2 * pos))
      return boundCoord(s.board, Math.trunc((h.cells[0] ?? 0) / 2));
  return null;
}
export function footprint(s: Session, e: Coord): Coord[] {
  const p = entrance(s, e);
  if (p === null || coordKey(p) !== coordKey(e)) return [];
  const pos = cell(s.board, e) ?? 0,
    rest: number[] = [];
  for (const [p, en] of s.board.parts) if (en === pos) rest.push(p);
  if (rest.length === 0)
    for (const [i, h] of s.game.holds)
      if (i >= holdFlag40 && h.cells[0] === 2 * pos)
        for (const c of h.cells.slice(1))
          if (c !== h.cells[0]) rest.push(Math.trunc(c / 2));
  return [e, ...rest.sort((a, b) => a - b).map((c) => hex(s, c))];
}
export function holdOf(s: Session, side: number, c: Coord): number | null {
  const pos = cell(s.board, c);
  if (pos === null || (side !== 0 && side !== 1)) return null;
  const [hold, found] = rules.holdAt(s.game, side, pos);
  if (!found) return null;
  const h = s.game.holds.get(hold);
  if (h === undefined) return null;
  if (hold < holdFlag40) {
    const u = rules.liveOf(s.game, at(s.game.maps, side)[2 * pos + 1] ?? 0);
    if (flagSide(u.flags4) !== side) return null;
  } else if (flagSide(h.flags) !== side) return null;
  return hold;
}
export function contents(
  s: Session,
  side: number,
  c: Coord,
): { hold: number; slots: UnitView[] } | null {
  const hold = holdOf(s, side, c);
  return hold === null ? null : { hold, slots: slots(s, hold, side) };
}
export function contentSlots(
  s: Session,
  side: number,
  c: Coord,
): { hold: number; slots: SlotView[] } | null {
  const hold = holdOf(s, side, c);
  return hold === null ? null : { hold, slots: slotViews(s, hold, side) };
}
