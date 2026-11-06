import { play } from "@historyline/engine";

export interface SideState {
  shown: number;
  lost: number;
  count: number;
  rise: number;
  hit: number;
}
export interface Flash {
  kind: 0 | 1 | 2;
  side: number;
  figure: number;
  progress: number;
}
export interface BattleState {
  sides: [SideState, SideState];
  flash: Flash[];
  gained: [boolean, boolean];
  glow: [number, number];
}

// Native internal/dospanel/battle.go timings, at 60 ticks per second.
const maxFigures = 6,
  buildStep = 6,
  buildPause = 12,
  shotStep = 28,
  muzzleTicks = 10,
  blastFrom = 8,
  blastTicks = 20,
  landAt = 14,
  volleyGap = 16,
  holdTicks = 70,
  glowTicks = 40,
  hitTicks = 16,
  reducedHold = 20;
type Side = 0 | 1;
const sides = [0, 1] as const;
const shooterOrder = [
  [3, 2, 4, 1, 5, 0],
  [0, 1, 5, 2, 4, 3],
] as const;
interface Schedule {
  before: [number, number];
  after: [number, number];
  exp: [[number, number], [number, number]];
  builds: { at: number; side: Side; shown: number }[];
  shots: {
    at: number;
    side: Side;
    figure: number;
    shooter: number;
    miss: boolean;
  }[];
  holdAt: number;
  total: number;
}

function plan(b: play.BattleInfo): Schedule {
  const fighters = [b.attacker, b.defender] as const,
    s: Schedule = {
      before: [0, 0],
      after: [0, 0],
      exp: [
        [0, 0],
        [0, 0],
      ],
      builds: [],
      shots: [],
      holdAt: 0,
      total: 0,
    };
  for (const side of sides) {
    const f = fighters[side];
    s.before[side] = Math.min(Math.max(f.before, 0), maxFigures);
    s.after[side] = f.destroyed
      ? 0
      : Math.min(Math.max(f.after, 0), s.before[side]);
    s.exp[side] = [f.expBefore, f.expAfter];
  }
  let t = 0;
  for (let i = 0; i < maxFigures; i++)
    for (const side of sides) {
      const ship = fighters[side].unit.domain === play.Naval;
      if (i >= s.before[side] || (ship && i > 0)) continue;
      s.builds.push({ at: t, side, shown: ship ? s.before[side] : i + 1 });
      t += buildStep;
    }
  t += buildPause;

  const shooter = (side: Side, standing: number, k: number): number => {
    const living = shooterOrder[side].filter((f) => f < standing);
    return living[k % living.length] ?? -1;
  };
  const fire = (target: Side, standing: number) => {
    for (let k = 0; k < s.before[target] - s.after[target]; k++) {
      s.shots.push({
        at: t,
        side: target,
        figure: s.before[target] - 1 - k,
        shooter: shooter(target === 0 ? 1 : 0, standing, k),
        miss: false,
      });
      t += shotStep;
    }
  };
  fire(1, s.before[0]);
  if (s.shots.length > 0) t += volleyGap;
  if (s.after[0] < s.before[0]) fire(0, s.after[1]);
  else if (b.returnFire && s.after[1] > 0 && s.before[0] > 0) {
    s.shots.push({
      at: t,
      side: 0,
      figure: s.before[0] - 1,
      shooter: shooter(1, s.after[1], 0),
      miss: true,
    });
    t += shotStep;
  }
  s.holdAt = t;
  s.total = t + holdTicks;
  return s;
}

/** Length of the build-up, volleys and final hold; reduced motion holds only the end. */
export function battleTicks(b: play.BattleInfo, reduced: boolean): number {
  return reduced ? reducedHold : plan(b).total;
}

/** First falling figure, or the start of the final hold when no figure falls. */
export function firstLoss(b: play.BattleInfo, reduced: boolean): number {
  const s = plan(b);
  for (const shot of s.shots)
    if (!shot.miss) return reduced ? 0 : shot.at + landAt;
  return reduced ? reducedHold : s.holdAt;
}

const ramp01 = (v: number, at: number, n: number): number =>
  Math.max(0, Math.min(1, (v - at) / n));

/** Immutable event numbers drive the native scene; no combat is resolved here. */
export function battleAt(
  b: play.BattleInfo,
  tick: number,
  reduced: boolean,
): BattleState {
  const s = plan(b),
    empty = (): SideState => ({ shown: 0, lost: 0, count: 0, rise: 0, hit: 0 }),
    st: BattleState = {
      sides: [empty(), empty()],
      flash: [],
      gained: [false, false],
      glow: [0, 0],
    };
  if (reduced) {
    for (const side of sides) {
      st.sides[side] = {
        shown: s.before[side],
        lost: s.before[side] - s.after[side],
        count: s.after[side],
        rise: 1,
        hit: 0,
      };
      st.gained[side] = s.exp[side][1] > s.exp[side][0];
    }
    return st;
  }
  tick = Math.min(Math.max(tick, 0), s.total);
  for (const side of sides) st.sides[side].rise = 1;
  for (const build of s.builds)
    if (build.at <= tick) {
      st.sides[build.side].shown = build.shown;
      st.sides[build.side].rise = ramp01(tick, build.at, buildStep);
    }
  for (const shot of s.shots) {
    if (!shot.miss && shot.at + landAt <= tick) {
      const side = st.sides[shot.side];
      side.lost++;
      side.hit = 1 - ramp01(tick, shot.at + landAt, hitTicks);
    }
    if (shot.shooter >= 0 && tick >= shot.at && tick < shot.at + muzzleTicks)
      st.flash.push({
        kind: 0,
        side: 1 - shot.side,
        figure: shot.shooter,
        progress: ramp01(tick, shot.at, muzzleTicks),
      });
    if (tick >= shot.at + blastFrom && tick < shot.at + blastFrom + blastTicks)
      st.flash.push({
        kind: shot.miss ? 2 : 1,
        side: shot.side,
        figure: shot.figure,
        progress: ramp01(tick, shot.at + blastFrom, blastTicks),
      });
  }
  for (const side of sides) {
    st.sides[side].count = s.before[side] - st.sides[side].lost;
    if (s.exp[side][1] > s.exp[side][0] && tick >= s.holdAt) {
      st.gained[side] = true;
      st.glow[side] = 1 - ramp01(tick, s.holdAt, glowTicks);
    }
  }
  return st;
}
