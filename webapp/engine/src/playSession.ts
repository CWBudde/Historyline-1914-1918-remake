import * as rules from "./cycle.ts";
import * as board from "./board.ts";
import * as binding from "./bind.ts";
import { type Coord } from "./hexgrid.ts";
import { type Level, grid, coordKey } from "./level.ts";
import { type Profiles } from "./profile.ts";
import { extend, usesExtension } from "./playGas.ts";
import { gasPioneer as ruleGasPioneer } from "./gas.ts";
import { save as saveGame, loadGame } from "./save.ts";
import { surrender as surrenderGame } from "./victory.ts";
import { slotFree } from "./container.ts";
import { costs as moveCosts, pieceAt, type Movement } from "./move.ts";
import { sapper as ruleSapper } from "./sapper.ts";
import { builder as ruleBuilder } from "./builder.ts";
import { GoError, at } from "./goruntime.ts";
import {
  type Session,
  type Options,
  type Command,
  type Event,
  Moved,
  Revealed,
  copyEvent,
  zeroRawEvent,
  hex,
} from "./playModel.ts";
import { unitID, unitView, container, startCell } from "./playView.ts";
import {
  event,
  moved,
  wrap,
  refusal,
  ReplayError,
  type Battles,
} from "./playEvent.ts";

function fits(b: binding.Board, l: Level): void {
  const g = grid(l);
  if (2 * g.width !== b.width || g.lines !== 2 * b.height - 1)
    throw new GoError(
      `dosplay: a level of ${g.lines} lines of ${g.width} hexes for a ${b.width}×${b.height} board`,
    );
}
function session(
  g: rules.Game,
  b: binding.Board,
  l: Level,
  d: binding.BindingData,
): Session {
  return {
    game: g,
    board: b,
    level: l,
    data: d,
    version: 0,
    routes: [[], []],
    ids: new Map(),
    log: [],
    light: [],
    lightWater: -1,
    lightDone: false,
  };
}
export function newSession(
  l: Level,
  d: binding.BindingData,
  o: Options,
): Session {
  return newBoard(binding.bind(l, d, o.extended ?? o.profile), l, d, o);
}
export function newBoard(
  b: binding.Board,
  l: Level,
  d: binding.BindingData,
  o: Options,
): Session {
  fits(b, l);
  let p = o.profile;
  if (o.extendedUnits === true || usesExtension(b)) {
    if (o.extended === undefined)
      throw new GoError(
        "dosplay: the extension units need the extended profile",
      );
    p = o.extended;
  }
  const g = binding.game(b, p);
  if (o.extended !== undefined && p === o.extended)
    extend(g, d, l.extra?.month ?? "", o.extendedUnits === true);
  g.limit = o.limit ?? 0;
  for (const k of [0, 1]) at(g.players, k).computer = o.computer?.[k] ?? false;
  return session(g, b, l, d);
}
export function load(
  l: Level,
  d: binding.BindingData,
  raw: Uint8Array | string,
  profiles: Profiles,
): Session {
  return loadBoard(
    binding.bind(l, d, profiles.extended ?? profiles.full),
    l,
    d,
    raw,
    profiles,
  );
}
export function loadBoard(
  b: binding.Board,
  l: Level,
  d: binding.BindingData,
  raw: Uint8Array | string,
  profiles: Profiles,
): Session {
  fits(b, l);
  const g = loadGame(raw, profiles);
  if (
    g.width !== b.width ||
    g.height !== b.height ||
    board.winter(g.profile) !== b.winter
  )
    throw new GoError(
      `dosplay: a saved ${g.width}×${g.height} battle (winter ${board.winter(g.profile)}) on a ${b.width}×${b.height} level (winter ${b.winter})`,
    );
  return session(g, b, l, d);
}
export function save(s: Session): Uint8Array {
  return saveGame(s.game);
}
export function phase(s: Session, side: number): number {
  return at(s.game.players, side).phase;
}
export function ready(s: Session, side: number): boolean {
  return at(s.game.players, side).ready;
}
export function turn(s: Session): number {
  return s.game.turn;
}
export function ended(s: Session) {
  return s.game.ended;
}
export function energy(s: Session, side: number): number {
  return at(s.game.energy, side);
}
export function version(s: Session): number {
  return s.version;
}
export function ordersLeft(s: Session, side: number): number {
  const p = at(s.game.players, side);
  return s.game.limit <= 0 || p.computer
    ? -1
    : Math.max(0, s.game.limit - p.orders.length);
}
function cell(s: Session, c: Coord): number {
  const p = binding.cell(s.board, c);
  if (p === null)
    throw new GoError(`dosplay: hex ${coordKey(c)} is not on the level`);
  return p;
}
function command(
  s: Session,
  side: number,
  unit: number,
  cell: number,
): rules.Command {
  if (side !== 0 && side !== 1) throw new GoError(`dosplay: side ${side}`);
  const u = s.game.units.get(unit);
  if (u === undefined) throw refusal("no such unit");
  const c: rules.Command = { kind: "move", side, unit, cell };
  if (phase(s, side) === rules.phaseAction) c.kind = "attack";
  else if ((u.flags4 & 0x4000) !== 0) {
    const h = container(s, side, unit);
    if (h === null) throw refusal("the unit is not the side's to order");
    c.kind = "exit";
    c.hold = h.hold;
    c.slot = h.slot;
  }
  return c;
}
function checked(s: Session, c: rules.Command): void {
  const err = rules.check(s.game, c);
  if (err !== null) throw wrap(err);
}
export function canOrder(s: Session, side: number, unit: number): void {
  const u = s.game.units.get(unit);
  if (u === undefined) throw refusal("no such unit");
  const c = command(s, side, unit, Math.trunc(at(u.pos, side & 1) / 2)),
    err = rules.check(s.game, c);
  if (
    c.kind === "attack" &&
    err instanceof rules.Refusal &&
    err.code === rules.msgNoTarget
  )
    return;
  if (err !== null) throw wrap(err);
}
function canMove(s: Session, side: number, unit: number): void {
  canOrder(s, side, unit);
  if (
    phase(s, side) !== rules.phaseMovement ||
    (rules.liveOf(s.game, unit).flags4 & 0x4000) !== 0
  )
    throw refusal("the unit does not move from the map now");
}
function hexes(s: Session, cells: number[], skip: number): Set<string> {
  return new Set(
    cells.flatMap((c) => {
      const p = binding.coord(s.board, c);
      return p !== null && c !== skip ? [coordKey(p)] : [];
    }),
  );
}
function lightCells(s: Session): void {
  if (s.lightDone) return;
  s.lightDone = true;
  s.lightWater = -1;
  for (const [line, hs] of s.level.lines.entries())
    for (const [col, h] of hs.entries()) {
      const c = binding.cell(s.board, { line, col });
      if (c === null || binding.edge(s.board, c)) continue;
      const ids = h.infra.ids,
        water = ids.some((id) => /^(stream|lake)/.test(id)),
        bridge = ids.some((id) => /^(bridge|rBridge)/.test(id)),
        deck = ids.some((id) => /^(path|bridge)/.test(id)),
        full = ids.some((id) => /^(road|rail)/.test(id));
      if (water && bridge && deck && !full) s.light.push(c);
      else if (s.lightWater < 0 && water && !bridge)
        s.lightWater = at(s.board.terrain, c);
    }
}
function crossesLight(s: Session, unit: number): boolean {
  const u = s.game.units.get(unit);
  if (u === undefined) return false;
  if ((u.flags4 & 0x10) !== 0) return true;
  const d = s.data.units.get(unitID(s, unit));
  return d !== undefined && d.weight <= 3 && d.armour <= 25;
}
export function patchLight(s: Session, m: Movement, unit: number): void {
  lightCells(s);
  if (s.light.length === 0 || s.lightWater < 0 || crossesLight(s, unit)) return;
  for (const c of s.light) m.terrain[c] = s.lightWater;
}
export function movement(s: Session, side: number, unit: number): Movement {
  const m = rules.movement(s.game, side, unit);
  patchLight(s, m, unit);
  return m;
}
function blockLight(s: Session, unit: number): () => void {
  lightCells(s);
  if (s.light.length === 0 || s.lightWater < 0 || crossesLight(s, unit))
    return () => {};
  const old: { side: number; cell: number; old: number }[] = [];
  for (let k = 0; k < 2; k++)
    for (const c of s.light) {
      const m = at(s.game.maps, k);
      old.push({ side: k, cell: c, old: at(m, 2 * c) });
      m[2 * c] = s.lightWater;
    }
  return () => {
    for (const x of old) at(s.game.maps, x.side)[2 * x.cell] = x.old;
  };
}
export function destinations(
  s: Session,
  side: number,
  unit: number,
): Set<string> {
  canMove(s, side, unit);
  return hexes(
    s,
    board.destinations(s.game.profile, movement(s, side, unit)),
    board.liveCell(rules.liveOf(s.game, unit), side),
  );
}
export function route(
  s: Session,
  side: number,
  unit: number,
  to: Coord,
): Coord[] {
  canMove(s, side, unit);
  return (
    board.route(s.game.profile, movement(s, side, unit), cell(s, to)) ?? []
  ).map((c) => hex(s, c));
}
export function costs(
  s: Session,
  side: number,
  unit: number,
): Map<string, number> {
  if (side !== 0 && side !== 1) throw new GoError(`dosplay: side ${side}`);
  const u = s.game.units.get(unit);
  if (u === undefined || board.destroyed(u)) throw refusal("no such unit");
  const t = board.moveTablesOf(s.game.profile),
    m = movement(s, side, unit),
    p = pieceAt(m, unit);
  m.units = new Map();
  if ((u.flags4 & 0x4000) === 0) m.units.set(unit, { ...p });
  const buf = moveCosts(t, m),
    out = new Map<string, number>();
  for (let c = 0; c < s.game.width * s.game.height; c++) {
    const v = buf[Math.trunc(c / s.game.width) * 64 + (c % s.game.width)] ?? 0,
      pos = binding.coord(s.board, c);
    if (v < 0 && v > -80 && pos !== null) out.set(coordKey(pos), -v);
  }
  return out;
}
export function canExit(
  s: Session,
  side: number,
  hold: number,
  slot: number,
): void {
  const h = s.game.holds.get(hold);
  if (h === undefined) throw refusal("no such container");
  checked(s, { kind: "exit", side, hold, slot, cell: startCell(h) });
}
export function exitDestinations(
  s: Session,
  side: number,
  hold: number,
  slot: number,
): Set<string> {
  canExit(s, side, hold, slot);
  let m: Movement;
  try {
    m = rules.exitMovement(s.game, side, hold, slot);
  } catch (e) {
    throw wrap(e);
  }
  const h = rules.holdOf(s.game, hold),
    u = h.slots[7 * side + slot] ?? 0;
  if (u <= slotFree) patchLight(s, m, u);
  return hexes(s, board.destinations(s.game.profile, m), startCell(h));
}
export function targets(s: Session, side: number, unit: number): Set<string> {
  const out = new Set<string>();
  if ((side !== 0 && side !== 1) || phase(s, side) !== rules.phaseAction)
    return out;
  try {
    canOrder(s, side, unit);
  } catch {
    return out;
  }
  const work =
      ruleSapper(rules.liveOf(s.game, unit)) ||
      ruleBuilder(rules.liveOf(s.game, unit)) ||
      ruleGasPioneer(rules.liveOf(s.game, unit)),
    m = at(s.game.maps, side);
  for (let c = 0; c < s.game.width * s.game.height; c++)
    if (
      (work || at(m, 2 * c + 1) <= slotFree) &&
      binding.coord(s.board, c) !== null &&
      rules.check(s.game, { kind: "attack", side, unit, cell: c }) === null
    )
      out.add(coordKey(hex(s, c)));
  return out;
}
export function battlefield(s: Session, side: number) {
  return rules.battlefield(s.game, side);
}
function apply(s: Session, c: rules.Command, start: number): Event[] {
  let raw: rules.Event[];
  try {
    raw = rules.apply(s.game, c);
  } catch (e) {
    throw wrap(e);
  }
  s.version++;
  s.log.push({ ...c });
  return raw.map((r) => {
    const ev =
      r.kind === "moved" || r.kind === "refused"
        ? moved(s, r, start)
        : event(s, r);
    if (ev.kind === Moved) at(s.routes, c.side).push(copyEvent(ev));
    return ev;
  });
}
export function move(
  s: Session,
  side: number,
  unit: number,
  to: Coord,
): Event[] {
  const pos = cell(s, to);
  canMove(s, side, unit);
  const u = rules.liveOf(s.game, unit),
    rear =
      (u.flags4 & 0xc0) === 0xc0
        ? board.liveCell(rules.liveOf(s.game, unit + 1), side)
        : -1;
  const restore = blockLight(s, unit);
  let evs: Event[];
  try {
    evs = apply(
      s,
      { kind: "move", side, unit, cell: pos },
      board.liveCell(u, side),
    );
  } finally {
    restore();
  }
  for (const ev of evs)
    if (rear >= 0 && ev.kind === Moved && ev.unit === unit)
      ev.rear = [hex(s, rear), ...ev.path.slice(0, -1)];
  return evs;
}
export function exit(
  s: Session,
  side: number,
  hold: number,
  slot: number,
  to: Coord,
): Event[] {
  const pos = cell(s, to),
    h = s.game.holds.get(hold);
  if (h === undefined) throw refusal("no such container");
  if (pos === startCell(h)) throw refusal("the unit must leave the container");
  const u = at(h.slots, 7 * side + slot),
    restore = u <= slotFree ? blockLight(s, u) : () => {};
  try {
    return apply(
      s,
      { kind: "exit", side, hold, slot, cell: pos },
      startCell(h),
    );
  } finally {
    restore();
  }
}
export function attack(
  s: Session,
  side: number,
  unit: number,
  to: Coord,
): Event[] {
  return apply(s, { kind: "attack", side, unit, cell: cell(s, to) }, -1);
}
export function repair(
  s: Session,
  side: number,
  hold: number,
  slot: number,
): Event[] {
  return apply(s, { kind: "repair", side, hold, slot, cell: 0 }, -1);
}
export function build(
  s: Session,
  side: number,
  hold: number,
  type: number,
): Event[] {
  return apply(s, { kind: "build", side, hold, type, cell: 0 }, -1);
}
export function confirm(s: Session, side: number): Event[] {
  s.version++;
  s.log.push({ kind: "confirm", side, cell: 0 });
  return rules.confirm(s.game, side).map((r) => event(s, r));
}
export function unconfirm(s: Session, side: number): Event[] {
  const raw = rules.unconfirm(s.game, side);
  if (raw.length > 0) {
    s.version++;
    s.log.push({ kind: "unconfirm", side, cell: 0 });
  }
  return raw.map((r) => event(s, r));
}
export function surrender(s: Session, side: number): Event[] {
  const raw = surrenderGame(s.game, side);
  if (raw.length === 0) return [];
  s.version++;
  s.log.push({ kind: "surrender", side, cell: 0 });
  return raw.map((r) => event(s, r));
}
export function resolve(s: Session): Event[] {
  const a = rules.action(s.game),
    b: Battles = { views: new Map(), state: new Map(), half: new Map() };
  for (const [i, u] of s.game.units) {
    b.views.set(i, unitView(s, i, a));
    b.state.set(i, [u.size, u.alt, u.exp]);
    if ((u.flags4 & 0xc0) === 0xc0) {
      b.half.set(i, i + 1);
      b.half.set(i + 1, i);
    }
  }
  const raw = rules.resolve(s.game);
  if (raw.length > 0) {
    s.version++;
    s.log.push({ kind: "resolve", side: 0, cell: 0 });
  }
  const out: Event[] = [];
  for (const r of raw) {
    out.push(event(s, r, b));
    if (r.kind !== "phase") continue;
    for (const mv of at(s.routes, 1 - a))
      if (!board.destroyed(rules.liveOf(s.game, mv.unit)))
        out.push({ ...mv, kind: Revealed, text: "" });
    s.routes = [[], []];
  }
  return out;
}
export function orders(s: Session, side: number): Event[] {
  if ((side !== 0 && side !== 1) || phase(s, side) !== rules.phaseAction)
    return [];
  return at(s.game.players, side).orders.map((u) => {
    const l = rules.liveOf(s.game, u);
    return {
      ...event(
        s,
        rules.newEvent({
          kind: "attack",
          side,
          unit: u,
          cell: Math.trunc(l.target / 2),
        }),
      ),
      text: "",
      raw: zeroRawEvent(),
      from: hex(s, board.liveCell(l, side)),
    };
  });
}
export function clone(s: Session, profiles: Profiles): Session {
  // The game's own profile reads its save back, an extended one too.
  const g = loadGame(save(s), {
    ...profiles,
    extended: profiles.extended ?? s.game.profile,
  });
  g.profile = s.game.profile;
  return {
    ...session(g, s.board, s.level, s.data),
    version: s.version,
    ids: new Map(s.ids),
    log: s.log.map((c) => ({ ...c })),
    routes: [s.routes[0].map(copyEvent), s.routes[1].map(copyEvent)],
  };
}
export function mark(s: Session): number {
  return s.log.length;
}
export function commandsSince(s: Session, mark: number): Command[] {
  return mark < 0 || mark >= s.log.length
    ? []
    : s.log.slice(mark).map((c) => ({ ...c }));
}
export function replay(s: Session, cmds: readonly Command[]): Event[] {
  const out: Event[] = [];
  for (const [i, c] of cmds.entries()) {
    try {
      switch (c.kind) {
        case "repair":
          out.push(...repair(s, c.side, c.hold ?? 0, c.slot ?? 0));
          continue;
        case "build":
          out.push(...build(s, c.side, c.hold ?? 0, c.type ?? 0));
          continue;
        case "resolve":
          out.push(...resolve(s));
          continue;
      }
      if (c.side !== 0 && c.side !== 1)
        throw new GoError(`dosplay: side ${c.side}`);
      switch (c.kind) {
        case "confirm":
          out.push(...confirm(s, c.side));
          continue;
        case "unconfirm":
          out.push(...unconfirm(s, c.side));
          continue;
        case "surrender":
          out.push(...surrender(s, c.side));
          continue;
      }
      const to = binding.coord(s.board, c.cell);
      if (to === null)
        throw new GoError(`dosplay: cell ${c.cell} is not on the level`);
      switch (c.kind) {
        case "move":
          out.push(...move(s, c.side, c.unit ?? 0, to));
          break;
        case "exit":
          out.push(...exit(s, c.side, c.hold ?? 0, c.slot ?? 0, to));
          break;
        case "attack":
          out.push(...attack(s, c.side, c.unit ?? 0, to));
          break;
        default:
          throw new GoError(`dosplay: unknown command ${String(c.kind)}`);
      }
    } catch (e) {
      if (!(e instanceof GoError)) throw e;
      throw new ReplayError(i, c.kind, e, out);
    }
  }
  return out;
}
