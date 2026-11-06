import { play, hexgrid, level } from "@historyline/engine";
import { toward, type Facing } from "./unitFrame.ts";
import type { Point } from "./camera.ts";

const stepMS = (8 * 1000) / 60,
  reducedMS = (20 * 1000) / 60,
  epsilonMS = 1e-7;
interface Walk {
  unit: play.UnitView;
  path: hexgrid.Coord[];
  rear: hexgrid.Coord[];
  elapsed: number;
}
interface Mark {
  at: hexgrid.Coord;
  death: boolean;
  start: number;
  duration: number;
  reduced: boolean;
}
export interface WalkingUnit {
  unit: play.UnitView;
  at: Point;
  rear: Point | null;
  facing: Facing;
}
export interface MapEffect {
  key: string;
  at: Point;
  frame: string;
  alpha: number;
}

/** Immutable visual snapshots follow events after the rules have already run. */
export class MapPlayback {
  readonly directions = new Map<number, Facing>();
  private walk: Walk | null = null;
  private frozen: play.View | null = null;
  private result: play.View | null = null;
  private queue: play.Event[] = [];
  private wait = 0;
  private time = 0;
  private marks: Mark[] = [];
  constructor(
    readonly side: number,
    private reduced = false,
  ) {}

  move(event: play.Event, unit: play.UnitView): void {
    if (event.kind === play.Revealed) {
      this.face(event);
      return;
    }
    if (event.kind !== play.Moved || event.side !== this.side) return;
    this.face(event);
    if (this.reduced || event.path.length < 2) return;
    this.walk = {
      unit: structuredClone(unit),
      path: structuredClone(event.path),
      rear: structuredClone(event.rear),
      elapsed: 0,
    };
  }
  skipWalk(): boolean {
    if (this.walk === null) return false;
    this.walk = null;
    return true;
  }
  moving(): WalkingUnit | null {
    const w = this.walk;
    if (w === null) return null;
    const last = w.path.length - 1;
    if (w.elapsed >= last * stepMS) return null;
    const i = Math.min(Math.floor(w.elapsed / stepMS), last - 1),
      f = (w.elapsed % stepMS) / stepMS;
    const along = (path: hexgrid.Coord[]): Point => {
      const a = path[i],
        b = path[i + 1];
      if (a === undefined || b === undefined)
        throw new Error("invalid visual route");
      const [x, y] = hexgrid.center(a),
        [xx, yy] = hexgrid.center(b);
      return { x: x + (xx - x) * f, y: y + (yy - y) * f };
    };
    const a = w.path[i],
      b = w.path[i + 1];
    if (a === undefined || b === undefined) return null;
    const [x, y] = hexgrid.center(a),
      [xx, yy] = hexgrid.center(b);
    return {
      unit: w.unit,
      at: along(w.path),
      rear: w.rear.length === w.path.length ? along(w.rear) : null,
      facing: toward({ x, y }, { x: xx, y: yy }),
    };
  }
  shown(live: play.View): play.View {
    return this.frozen ?? live;
  }
  busy(): boolean {
    return (
      this.walk !== null ||
      this.frozen !== null ||
      this.queue.length > 0 ||
      this.wait > 0
    );
  }
  resolution(
    before: play.View,
    after: play.View,
    events: readonly play.Event[],
  ): void {
    this.walk = null;
    this.frozen = structuredClone(before);
    this.result = structuredClone(after);
    this.queue = structuredClone([...events]);
    this.wait = 0;
    this.marks = [];
    this.normalize();
  }
  advance(ms: number): void {
    if (!Number.isFinite(ms) || ms < 0)
      throw new Error("invalid presentation interval");
    if (this.walk !== null) {
      this.walk.elapsed += ms;
      if (this.walk.elapsed + epsilonMS >= (this.walk.path.length - 1) * stepMS)
        this.walk = null;
    }
    let remaining = ms;
    while (this.frozen !== null || this.queue.length > 0 || this.wait > 0) {
      this.normalize();
      if (this.wait === 0) break;
      if (remaining === 0) break;
      const used = Math.min(remaining, this.wait);
      this.time += used;
      remaining -= used;
      this.wait -= used;
      if (this.wait < epsilonMS) this.wait = 0;
    }
    this.normalize();
    this.time += remaining;
    this.marks = this.marks.filter(
      (m) => this.time - m.start + epsilonMS < m.duration,
    );
  }
  effects(): MapEffect[] {
    return this.marks
      .filter((m) => this.time - m.start < m.duration)
      .map((m, i) => {
        const elapsed = this.time - m.start,
          last = m.death ? 8 : 3,
          frame = m.reduced ? last : Math.min(last, Math.floor(elapsed / 100));
        const [x, y] = hexgrid.center(m.at);
        return {
          key: `effect/${m.start}/${i}`,
          at: { x, y },
          frame: `effect/explosion-${String(frame).padStart(2, "0")}`,
          alpha: m.reduced ? 1 : Math.max(0, 1 - elapsed / m.duration),
        };
      });
  }
  setReducedMotion(reduced: boolean): void {
    if (reduced === this.reduced) return;
    this.reduced = reduced;
    if (reduced) {
      this.walk = null;
      for (const m of this.marks) {
        m.duration = Math.min(reducedMS, m.duration - (this.time - m.start));
        m.start = this.time;
        m.reduced = true;
      }
      this.wait = Math.min(this.wait, reducedMS);
    }
  }
  private face(event: play.Event): void {
    const n = event.path.length,
      a = n > 1 ? event.path[n - 2] : event.from,
      b = n > 1 ? event.path[n - 1] : event.to;
    if (a === undefined || b === undefined) return;
    const [x, y] = hexgrid.center(a),
      [xx, yy] = hexgrid.center(b);
    this.directions.set(event.unit, toward({ x, y }, { x: xx, y: yy }));
  }
  private normalize(): void {
    while (this.wait === 0) {
      const e = this.queue.shift();
      if (e === undefined) {
        this.frozen = null;
        this.result = null;
        return;
      }
      if (e.kind === play.Battle && e.battle !== null) {
        this.battle(e);
        return;
      }
      if (e.kind === play.PhaseChanged) {
        this.frozen = null;
        continue;
      }
      if (e.kind === play.Revealed) {
        this.face(e);
        continue;
      }
      if (
        e.kind === play.Captured &&
        this.frozen !== null &&
        this.result !== null
      ) {
        for (const [key, b] of this.frozen.buildings) {
          if (b.hold === e.hold) {
            const current = this.result.buildings.get(key);
            if (current !== undefined)
              this.frozen.buildings.set(key, structuredClone(current));
          }
        }
      }
    }
  }
  private battle(e: play.Event): void {
    const b = e.battle;
    if (b === null) return;
    const point = (c: hexgrid.Coord): Point => {
      const [x, y] = hexgrid.center(c);
      return { x, y };
    };
    this.directions.set(
      b.attacker.unit.index,
      toward(point(e.from), point(e.to)),
    );
    if (b.returnFire)
      this.directions.set(
        b.defender.unit.index,
        toward(point(e.to), point(e.from)),
      );
    let duration = 400;
    for (const f of [b.defender, b.attacker]) {
      const old =
          this.frozen === null
            ? undefined
            : [...this.frozen.units.values()].find(
                (u) => u.index === f.unit.index,
              ),
        at = old?.at ?? f.unit.at;
      const hit = f.after < f.before || f.altAfter < f.altBefore;
      if (hit || f.destroyed) {
        const ms = this.reduced ? reducedMS : f.destroyed ? 900 : 400;
        duration = Math.max(duration, ms);
        this.marks.push({
          at: { ...at },
          death: f.destroyed,
          start: this.time,
          duration: ms,
          reduced: this.reduced,
        });
      }
      if (this.frozen !== null) {
        const indices = new Set([f.unit.index]);
        for (const partner of [old?.head, old?.rear]) {
          if (partner !== undefined && partner !== null) {
            const u = this.frozen.units.get(level.coordKey(partner));
            if (u !== undefined) indices.add(u.index);
          }
        }
        for (const [key, u] of this.frozen.units) {
          if (!indices.has(u.index)) continue;
          if (f.destroyed) this.frozen.units.delete(key);
          else
            this.frozen.units.set(key, {
              ...u,
              size: f.after,
              alt: f.altAfter,
              exp: f.expAfter,
            });
        }
      }
    }
    this.wait = this.reduced ? Math.min(duration, reducedMS) : duration;
  }
}
