// A direct port of internal/dosai. Only the side's view and public session queries inform its orders.
import * as play from "./play.ts";
import { type Coord, distance, around } from "./hexgrid.ts";
import { coordKey } from "./level.ts";
import { cell } from "./bind.ts";
import { type Profile, type Type, tablesOf } from "./profile.ts";
import { moveTablesOf } from "./board.ts";
import { type Battlefield, defaultClass, engage } from "./engage.ts";
import { type Order, resolveOrders } from "./orders.ts";
import { weaponAgainst } from "./cycle.ts";
import { GoError, at } from "./goruntime.ts";
import { dijkstra } from "./aiField.ts";
export { dijkstra } from "./aiField.ts";

export interface Planner {
  progress?: (done: number, total: number) => void;
}
export class PlanError extends GoError {
  readonly events: play.Event[];
  constructor(cause: GoError, events: play.Event[]) {
    super(cause.message, { cause });
    this.events = events;
  }
}
interface Plan {
  s: play.Session;
  side: number;
  profile: Profile;
  view: play.View;
  bf: Battlefield;
  enemies: Coord[];
  hqs: Coord[];
  ownHQs: Coord[];
  prizes: Coord[];
  fields: Map<number, Map<string, number>>;
  missions: Map<number, Mission>;
  hopFields: Map<string, Map<string, number>>;
  routes: Map<string, Map<string, number>>;
  progress: ((done: number, total: number) => void) | undefined;
}
function sorted<V>(m: ReadonlyMap<string, V>): Coord[] {
  return play.coords(new Set(m.keys()));
}
function nearest(c: Coord, cs: readonly Coord[]): number {
  let d = -1;
  for (const g of cs) {
    const n = distance(c, g);
    if (d < 0 || n < d) d = n;
  }
  return d;
}
function includes(cs: readonly Coord[], c: Coord): boolean {
  return cs.some((p) => p.line === c.line && p.col === c.col);
}
function report(pl: Plan, done: number, total: number): void {
  pl.progress?.(done, total);
}
function newPlan(s: play.Session, side: number, p: Planner): Plan {
  const pl: Plan = {
    s,
    side,
    profile: s.game.profile,
    view: play.view(s, side),
    bf: play.battlefield(s, side),
    enemies: [],
    hqs: [],
    ownHQs: [],
    prizes: [],
    fields: new Map(),
    missions: new Map(),
    hopFields: new Map(),
    routes: new Map(),
    progress: p.progress,
  };
  for (const c of sorted(pl.view.units))
    if (pl.view.units.get(coordKey(c))?.side === 1 - side) pl.enemies.push(c);
  for (const c of sorted(pl.view.buildings)) {
    const b = pl.view.buildings.get(coordKey(c));
    if (b === undefined) continue;
    if (!b.hq) {
      if (b.owner !== play.faction(side)) pl.prizes.push(c);
    } else if (b.owner === play.faction(1 - side)) pl.hqs.push(c);
    else if (b.owner === play.faction(side)) pl.ownHQs.push(c);
  }
  return pl;
}
function canOrder(pl: Plan, unit: number): boolean {
  try {
    play.canOrder(pl.s, pl.side, unit);
    return true;
  } catch (e) {
    if (e instanceof GoError) return false;
    throw e;
  }
}
interface Mover {
  unit: number;
  at: Coord;
  held: boolean;
  leave: boolean;
  dist: number;
  value: number;
}
function goals(pl: Plan): Coord[] {
  return [...pl.hqs, ...pl.enemies];
}
function movers(pl: Plan): Mover[] {
  const out: Mover[] = [],
    gs = goals(pl);
  const add = (
    u: play.UnitView,
    c: Coord,
    held: boolean,
    leave: boolean,
  ): void => {
    if (u.side === pl.side && canOrder(pl, u.index))
      out.push({
        unit: u.index,
        at: c,
        held,
        leave,
        dist: nearest(c, gs),
        value: pl.missions.get(u.index)?.value ?? 0,
      });
  };
  for (const c of sorted(pl.view.units)) {
    const u = pl.view.units.get(coordKey(c));
    if (u === undefined) continue;
    add(u, c, false, false);
    if (u.side === pl.side && u.carrier)
      for (const v of play.contents(pl.s, pl.side, c)?.slots ?? [])
        add(v, c, true, false);
  }
  for (const c of sorted(pl.view.buildings)) {
    const b = pl.view.buildings.get(coordKey(c));
    if (b?.owner === play.faction(pl.side))
      for (const v of b.units) add(v, c, true, true);
  }
  return out.sort(
    (a, b) =>
      Number(a.leave) - Number(b.leave) ||
      b.value - a.value ||
      a.dist - b.dist ||
      a.at.line - b.at.line ||
      a.at.col - b.at.col ||
      a.unit - b.unit,
  );
}
function flags4(pl: Plan, c: Coord): number {
  const idx = pl.view.units.get(coordKey(c))?.index ?? 0;
  return pl.bf.units.get(idx)?.flags4 ?? 0;
}
function reach(pl: Plan, u: play.UnitView, e: Coord): number {
  return Math.max(1, weaponAgainst(pl.profile, u.type, flags4(pl, e))[1] - 1);
}
function defence(pl: Plan, c: Coord): number {
  const n = cell(pl.s.board, c);
  return n === null
    ? 0
    : at(tablesOf(pl.profile).terms, pl.bf.terrain.get(n) ?? defaultClass)
        .defence;
}
function field(pl: Plan, u: play.UnitView): Map<string, number> {
  const cached = pl.fields.get(u.type);
  if (cached !== undefined) return cached;
  const costs = play.costs(pl.s, pl.side, u.index),
    long = (at(tablesOf(pl.profile).types, u.type).targetFlags & 0x40) !== 0,
    start = new Map<string, number>();
  // Parse value keys once per field, rather than once for every enemy and cell.
  const cells = [...costs.keys()].map((k) => play.coordFromKey(k));
  const seed = (c: Coord, v: number): void => {
    const k = coordKey(c);
    if (!costs.has(k)) return;
    const old = start.get(k);
    if (old === undefined || v < old) start.set(k, v);
  };
  for (const c of pl.hqs) seed(c, 0);
  for (const e of pl.enemies) {
    const off = pl.view.units.get(coordKey(e))?.budget === 0 ? 40 : 0;
    const hi = long ? reach(pl, u, e) : 1,
      lo = long ? Math.min(2, hi) : 1;
    for (const c of cells) {
      const d = distance(c, e);
      if (d >= lo && d <= hi) seed(c, off);
    }
  }
  const f = dijkstra(costs, start);
  pl.fields.set(u.type, f);
  return f;
}
function nearestEnemy(pl: Plan, c: Coord): [Coord, number] {
  let best = at(pl.enemies, 0),
    bd = -1;
  for (const e of pl.enemies) {
    const d = distance(c, e);
    if (bd < 0 || d < bd) {
      best = e;
      bd = d;
    }
  }
  return [best, bd];
}
function score(
  pl: Plan,
  u: play.UnitView,
  f: ReadonlyMap<string, number>,
  c: Coord,
): number {
  let sc = defence(pl, c);
  if (
    (at(moveTablesOf(pl.profile).types, u.type).flags & 1) === 0 &&
    nearest(c, pl.hqs) === 1
  )
    sc -= 50;
  const cost = f.get(coordKey(c));
  sc -= cost === undefined ? 1000 + 10 * nearest(c, goals(pl)) : 2 * cost;
  if (
    (at(tablesOf(pl.profile).types, u.type).targetFlags & 0x40) !== 0 &&
    pl.enemies.length > 0
  ) {
    const [e, d] = nearestEnemy(pl, c),
      r = reach(pl, u, e);
    if (d <= r) sc -= 10 * (r - d);
    if (d <= 1) sc -= 50;
  }
  return sc;
}
function hasLower(f: ReadonlyMap<string, number>, c: Coord, b: Coord): boolean {
  const fc = f.get(coordKey(c)),
    fb = f.get(coordKey(b));
  return fc !== undefined && (fb === undefined || fc < fb);
}
function choose(
  pl: Plan,
  u: play.UnitView,
  from: Coord,
  dests: Coord[],
  leave: boolean,
): Coord | null {
  for (const c of dests) if (includes(pl.hqs, c)) return c;
  let f: Map<string, number>;
  try {
    f = missionField(pl, u) ?? field(pl, u);
  } catch (e) {
    if (e instanceof GoError) return null;
    throw e;
  }
  for (const c of dests) {
    const cost = f.get(coordKey(c));
    if (
      cost !== undefined &&
      includes(pl.prizes, c) &&
      cost <= (f.get(coordKey(from)) ?? 0)
    )
      return c;
  }
  let best = from,
    bestScore = score(pl, u, f, from);
  if (u.held) bestScore -= defence(pl, from);
  if (leave) bestScore = -Infinity;
  for (const c of dests) {
    const k = coordKey(c);
    if (pl.view.units.has(k) || k === coordKey(from)) continue;
    let sc = score(pl, u, f, c);
    if (pl.view.buildings.has(k)) sc -= defence(pl, c);
    if (sc > bestScore || (sc === bestScore && hasLower(f, c, best))) {
      best = c;
      bestScore = sc;
    }
  }
  return coordKey(best) === coordKey(from) ? null : best;
}
function move(pl: Plan, out: play.Event[]): void {
  if (
    pl.enemies.length === 0 &&
    pl.hqs.length === 0 &&
    pl.prizes.length === 0
  ) {
    report(pl, 0, 0);
    return;
  }
  assignMissions(pl);
  const ms = movers(pl);
  report(pl, 0, ms.length);
  for (const [i, m] of ms.entries()) {
    if (canOrder(pl, m.unit)) {
      const u = play.unit(pl.s, m.unit);
      if (!m.held || u.in >= 0) {
        const dst = m.held
            ? play.exitDestinations(pl.s, pl.side, u.in, u.slot)
            : play.destinations(pl.s, pl.side, m.unit),
          to = choose(pl, u, m.at, play.coords(dst), m.leave);
        if (to !== null)
          out.push(
            ...(m.held
              ? play.exit(pl.s, pl.side, u.in, u.slot, to)
              : play.move(pl.s, pl.side, m.unit, to)),
          );
        pl.view = play.view(pl.s, pl.side);
      }
    }
    report(pl, i + 1, ms.length);
  }
}
interface Candidate {
  unit: number;
  target: Coord;
  cell: number;
  near: boolean;
}
function candidates(pl: Plan): [Candidate[], number[]] {
  const cands: Candidate[] = [];
  for (const c of sorted(pl.view.units)) {
    const u = pl.view.units.get(coordKey(c));
    if (
      u === undefined ||
      u.side !== pl.side ||
      u.moved ||
      play.sapper(pl.s, u.index) ||
      play.builder(pl.s, u.index)
    )
      continue;
    for (const t of play.coords(play.targets(pl.s, pl.side, u.index))) {
      if ((pl.view.units.get(coordKey(t))?.side ?? 0) !== 1 - pl.side) continue;
      const d = nearest(t, pl.ownHQs);
      cands.push({
        unit: u.index,
        target: t,
        cell: cell(pl.s.board, t) ?? 0,
        near: d >= 0 && d <= 3,
      });
    }
  }
  return [cands, [...new Set(cands.map((c) => c.unit))]];
}
function after(pl: Plan, orders: Order[]): Battlefield {
  if (orders.length === 0) return pl.bf;
  const r = resolveOrders(pl.profile, pl.bf, orders);
  return {
    ...pl.bf,
    units: new Map([...r.units].filter(([i]) => !r.gone.includes(i))),
  };
}
function strength(t: Type, size: number, alt: number): number {
  return (t.strengthFlags & 4) !== 0 ? alt : size;
}
function gain(pl: Plan, b: Battlefield, c: Candidate): number | null {
  const a = b.units.get(c.unit);
  if (a === undefined) return null;
  let o;
  try {
    o = engage(pl.profile, b, c.unit, c.cell);
  } catch (e) {
    if (e instanceof GoError) return null;
    throw e;
  }
  if (!o.battle) return null;
  const d = b.units.get(o.defender);
  if (d === undefined) return null;
  const ta = at(tablesOf(pl.profile).types, a.type),
    td = at(tablesOf(pl.profile).types, d.type);
  let sc =
    td.armour *
      (strength(td, d.size, d.alt) -
        strength(td, o.defenderSize, o.defenderAlt)) -
    ta.armour *
      (strength(ta, a.size, a.alt) -
        strength(ta, o.attackerSize, o.attackerAlt));
  for (const k of o.destroyed) sc += k === c.unit ? -100 : 100;
  if (c.near) sc *= 3;
  return sc;
}
function attack(
  pl: Plan,
  cands: Candidate[],
  n: number,
  total: number,
  out: play.Event[],
): void {
  const orders: Order[] = [];
  while (cands.length > 0) {
    const state = after(pl, orders);
    let best = -1,
      bestScore = 0;
    for (const [i, c] of cands.entries()) {
      const sc = gain(pl, state, c);
      if (sc !== null && sc > bestScore) {
        best = i;
        bestScore = sc;
      }
    }
    if (best < 0) break;
    const c = at(cands, best);
    out.push(...play.attack(pl.s, pl.side, c.unit, c.target));
    orders.push({ attacker: c.unit, target: c.cell });
    cands = cands.filter((o) => o.unit !== c.unit);
    if (orders.length < n) report(pl, orders.length, total);
  }
  if (n > 0) report(pl, n, total);
}
interface Digger {
  unit: number;
  at: Coord;
  near: boolean;
}
interface Work {
  target: Coord;
  hexes: Coord[];
}
function diggers(pl: Plan): Digger[] {
  const out: Digger[] = [];
  for (const c of sorted(pl.view.units)) {
    const u = pl.view.units.get(coordKey(c));
    if (
      u === undefined ||
      u.side !== pl.side ||
      u.moved ||
      !play.sapper(pl.s, u.index) ||
      !canOrder(pl, u.index)
    )
      continue;
    const e = nearest(c, pl.enemies),
      h = nearest(c, pl.ownHQs),
      near = h >= 0 && h <= 3;
    if (near || (e >= 0 && e <= 4)) out.push({ unit: u.index, at: c, near });
  }
  return out;
}
function clashes(given: Work[], target: Coord, hexes: Coord[]): boolean {
  return given.some(
    (w) => includes(w.hexes, target) || includes(hexes, w.target),
  );
}
function digScore(pl: Plan, d: Digger, hexes: Coord[], recs: number[]): number {
  const t = moveTablesOf(pl.profile),
    front = nearest(d.at, pl.enemies),
    blocked = new Set<number>();
  let sc = 0;
  for (const [k, c] of hexes.entries()) {
    const tr = at(t.terrains, at(recs, k)),
      gain = at(tablesOf(pl.profile).terms, tr.class).defence - defence(pl, c);
    sc += gain;
    const v = pl.view.units.get(coordKey(c));
    if (v?.side === pl.side) sc += 3 * gain;
    else if (v?.side === 1 - pl.side) sc -= 4 * gain;
    if (front >= 0 && nearest(c, pl.enemies) < front) sc += 2;
    for (const n of [...around(c), c]) {
      const v = pl.view.units.get(coordKey(n));
      if (
        v !== undefined &&
        v.side === pl.side &&
        !blocked.has(v.index) &&
        (at(t.types, v.type).mask & tr.mask) === 0
      ) {
        blocked.add(v.index);
        sc -= 8;
      }
    }
  }
  return sc;
}
function dig(
  pl: Plan,
  ds: Digger[],
  done: number,
  total: number,
  out: play.Event[],
): void {
  const given: Work[] = [];
  for (const [i, d] of ds.entries()) {
    let best: Work = { target: { line: 0, col: 0 }, hexes: [] },
      bestScore = 0;
    for (const t of play.coords(play.targets(pl.s, pl.side, d.unit))) {
      if (pl.view.units.get(coordKey(t))?.side === 1 - pl.side) continue;
      const w = play.work(pl.s, pl.side, d.unit, t);
      if (w.fill || w.hexes.length === 0 || clashes(given, t, w.hexes))
        continue;
      let sc = digScore(pl, d, w.hexes, w.recs);
      if (d.near) sc *= 3;
      if (
        sc > bestScore ||
        (sc === bestScore &&
          sc > 0 &&
          nearest(t, pl.enemies) < nearest(best.target, pl.enemies))
      ) {
        best = { target: t, hexes: w.hexes };
        bestScore = sc;
      }
    }
    if (bestScore > 0) {
      out.push(...play.attack(pl.s, pl.side, d.unit, best.target));
      given.push(best);
    }
    report(pl, done + i + 1, total);
  }
}
// The computer's economy, as the original's computer side plans it (service.go).
export const repairFloor = 14,
  reserve = 61,
  maxBuilds = 10;
const computerBuilds: readonly (readonly number[])[] = [
  [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 14, 15, 17, 18, 19, 21],
  [
    10, 11, 12, 14, 15, 17, 18, 21, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46,
    47, 48, 49, 50,
  ],
];
/** mayBuild: the building's map list holds both sides' types, side 0 builds none above 35, side 1 none below 10; an empty list is the side's default. */
export function mayBuild(
  list: readonly number[],
  side: number,
  type: number,
): boolean {
  if (list.length === 0) return at(computerBuilds, side).includes(type);
  if ((side === 0 && type > 35) || (side === 1 && type < 10)) return false;
  return list.includes(type);
}
export function pick(
  turn: number,
  side: number,
  hold: number,
  slot: number,
  n: number,
): number {
  let x =
    Math.imul(turn, 73856093) ^
    Math.imul(side + 1, 19349663) ^
    Math.imul(hold, 83492791) ^
    Math.imul(slot + 1, 2654435761);
  x ^= x >>> 15;
  x = Math.imul(x, 2654435761);
  x ^= x >>> 13;
  return (x >>> 0) % n;
}
function ownBuildings(pl: Plan): play.BuildingView[] {
  const out: play.BuildingView[] = [];
  for (const c of sorted(pl.view.buildings)) {
    const b = pl.view.buildings.get(coordKey(c));
    if (b?.owner === play.faction(pl.side)) out.push(b);
  }
  return out.sort((a, b) => a.hold - b.hold);
}
function refused(check: () => void): boolean {
  try {
    check();
    return false;
  } catch (e) {
    if (e instanceof GoError) return true;
    throw e;
  }
}
function repairs(pl: Plan, energy: number, out: play.Event[]): void {
  if (energy <= repairFloor) return;
  const units = ownBuildings(pl)
    .flatMap((b) => b.units)
    .filter((u) => u.side === pl.side && u.size < u.full)
    .sort((a, b) => a.index - b.index);
  for (const v of units) {
    const u = play.unit(pl.s, v.index);
    if (
      refused(() => {
        play.canRepair(pl.s, pl.side, u.in, u.slot);
      })
    )
      continue;
    out.push(...play.repair(pl.s, pl.side, u.in, u.slot));
  }
}
function produce(pl: Plan, energy: number, out: play.Event[]): void {
  let built = 0;
  for (const b of ownBuildings(pl)) {
    if (b.kind !== play.Factory) continue;
    let e = energy;
    const list = pl.s.game.buildings.get(b.hold)?.computerBuilds ?? [];
    for (
      let slot = b.units.length;
      slot < play.Slots && built < maxBuilds;
      slot++
    ) {
      const budget = e - reserve;
      if (budget <= 0) break;
      const opts = play
        .buildable(pl.s, pl.side, b.hold)
        .filter((o) => mayBuild(list, pl.side, o.type) && o.cost <= budget);
      if (opts.length === 0) break;
      const o = at(
        opts,
        pick(play.turn(pl.s), pl.side, b.hold, slot, opts.length),
      );
      e -= o.cost;
      built++;
      if (
        refused(() => {
          play.canBuild(pl.s, pl.side, b.hold, o.type);
        })
      )
        continue;
      out.push(...play.build(pl.s, pl.side, b.hold, o.type));
    }
  }
}
export function plan(p: Planner, s: play.Session, side: number): play.Event[] {
  if (side !== 0 && side !== 1) throw new GoError(`dosai: side ${side}`);
  if (play.ended(s) !== null || play.ready(s, side)) return [];
  const pl = newPlan(s, side, p),
    out: play.Event[] = [];
  try {
    if (play.phase(s, side) === 1) move(pl, out);
    else {
      const energy = play.energy(s, side),
        [cands, units] = candidates(pl),
        ds = diggers(pl),
        total = units.length + ds.length;
      report(pl, 0, total);
      attack(pl, cands, units.length, total, out);
      dig(pl, ds, units.length, total, out);
      repairs(pl, energy, out);
      produce(pl, energy, out);
    }
    out.push(...play.confirm(s, side));
  } catch (e) {
    if (e instanceof GoError) throw new PlanError(e, out);
    throw e;
  }
  return out;
}
export function commands(
  p: Planner,
  s: play.Session,
  side: number,
): play.Command[] {
  const mark = play.mark(s);
  plan(p, s, side);
  return play.commandsSince(s, mark);
}
export function playOut(
  s: play.Session,
  planners: readonly [Planner | null, Planner | null],
  maxTurns: number,
) {
  for (let n = 0; n < maxTurns; n++) {
    if (play.ended(s) !== null) break;
    for (const side of [0, 1]) {
      const p = at(planners, side);
      if (p === null) play.confirm(s, side);
      else plan(p, s, side);
    }
    play.resolve(s);
  }
  return play.ended(s);
}

// Capture missions (internal/dosai/capture.go): every depot and factory counts
// the same and above the enemy HQ, so that the route decides.
const buildingValue = 3850,
  hqValue = 450;
interface Mission {
  goal: Coord;
  value: number;
}
interface CaptureGoal {
  at: Coord;
  value: number;
}
function goalValue(b: play.BuildingView): number {
  return b.hq ? hqValue : buildingValue;
}
function assignMissions(pl: Plan): void {
  pl.missions = new Map();
  const goals = captureGoals(pl),
    cs = capturers(pl);
  for (const g of goals) assign(pl, g, cs, false);
  for (const g of goals) {
    let served = false;
    for (const m of pl.missions.values())
      served ||= coordKey(m.goal) === coordKey(g.at);
    if (!served) assign(pl, g, cs, true);
  }
}
function captureGoals(pl: Plan): CaptureGoal[] {
  const goals: CaptureGoal[] = [];
  for (const c of sorted(pl.view.buildings)) {
    const b = pl.view.buildings.get(coordKey(c));
    if (
      b !== undefined &&
      b.owner !== play.faction(pl.side) &&
      (!b.hq || b.owner === play.faction(1 - pl.side))
    )
      goals.push({ at: c, value: goalValue(b) });
  }
  // Array.prototype.sort is stable, as Go's SortStableFunc.
  return goals.sort((a, b) => b.value - a.value);
}
function capturers(pl: Plan): Coord[] {
  const out: Coord[] = [];
  for (const c of sorted(pl.view.units)) {
    const u = pl.view.units.get(coordKey(c));
    if (
      u !== undefined &&
      u.side === pl.side &&
      (at(moveTablesOf(pl.profile).types, u.type).flags & 1) !== 0 &&
      canOrder(pl, u.index)
    )
      out.push(c);
  }
  const index = (c: Coord): number =>
    pl.view.units.get(coordKey(c))?.index ?? 0;
  return out.sort((a, b) => index(a) - index(b));
}
function assign(
  pl: Plan,
  g: CaptureGoal,
  cs: readonly Coord[],
  free: boolean,
): void {
  let best = -1,
    bestHops = 0;
  for (const c of cs) {
    const u = pl.view.units.get(coordKey(c));
    if (u === undefined) continue;
    const m = pl.missions.get(u.index);
    if (free && m !== undefined) continue;
    const h = hops(pl, u, g.at, c);
    if (
      h !== null &&
      g.value - h > (m?.value ?? 0) &&
      (best < 0 || h < bestHops)
    ) {
      best = u.index;
      bestHops = h;
    }
  }
  if (best >= 0)
    pl.missions.set(best, { goal: g.at, value: g.value - bestHops });
}
function hops(
  pl: Plan,
  u: play.UnitView,
  goal: Coord,
  from: Coord,
): number | null {
  const k = `${u.type}:${coordKey(goal)}`;
  let dist = pl.hopFields.get(k);
  if (dist === undefined) {
    const costs = play.costs(pl.s, pl.side, u.index);
    dist = new Map();
    if (costs.has(coordKey(goal))) {
      dist.set(coordKey(goal), 0);
      const q: Coord[] = [goal];
      for (let i = 0; i < q.length; i++) {
        const p = at(q, i),
          d = dist.get(coordKey(p)) ?? 0;
        for (const n of around(p)) {
          const nk = coordKey(n);
          if (costs.has(nk) && !dist.has(nk)) {
            dist.set(nk, d + 1);
            q.push(n);
          }
        }
      }
    }
    pl.hopFields.set(k, dist);
  }
  return dist.get(coordKey(from)) ?? null;
}
function missionField(
  pl: Plan,
  u: play.UnitView,
): Map<string, number> | undefined {
  const m = pl.missions.get(u.index);
  if (
    m === undefined ||
    pl.view.buildings.get(coordKey(m.goal))?.owner === play.faction(pl.side)
  )
    return undefined;
  const k = `${u.type}:${coordKey(m.goal)}`;
  let f = pl.routes.get(k);
  if (f === undefined) {
    f = dijkstra(
      play.costs(pl.s, pl.side, u.index),
      new Map([[coordKey(m.goal), 0]]),
    );
    pl.routes.set(k, f);
  }
  return f;
}
