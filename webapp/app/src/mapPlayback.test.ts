import { expect, it } from "vitest";
import { play } from "@historyline/engine";
import { Full } from "../../engine/src/testing/tables.ts";
import { playData, shippedLevel } from "../../engine/src/testing/play.ts";
import { MapPlayback } from "./mapPlayback.ts";
import { battleTicks } from "./battle/timeline.ts";

function units() {
  const s = play.newSession(shippedLevel("1914-08"), playData, {
    profile: Full,
  });
  const u = [...play.view(s, 0).units.values()][0];
  if (u === undefined) throw new Error("fixture unit missing");
  const a = {
      ...u,
      index: 1,
      at: { line: 0, col: 0 },
      side: 0,
      owner: 1,
      size: 6,
      head: null,
      rear: null,
    },
    d = {
      ...u,
      index: 2,
      at: { line: 0, col: 1 },
      side: 1,
      owner: 2,
      size: 6,
      head: null,
      rear: null,
    };
  return { a, d };
}
function views() {
  const { a, d } = units();
  const before: play.View = {
    side: 0,
    buildings: new Map(),
    units: new Map([
      ["0,0", a],
      ["0,1", d],
    ]),
  };
  const after = structuredClone(before);
  after.units.delete("0,1");
  after.units.set("1,0", { ...d, at: { line: 1, col: 0 }, size: 4 });
  return { before, after, a, d };
}

it("interpolates only own movement for eight native ticks per edge", () => {
  const { a } = units();
  const p = new MapPlayback(0);
  const e = play.newEvent({
    kind: play.Moved,
    side: 0,
    unit: 1,
    path: [
      { line: 0, col: 0 },
      { line: 1, col: 0 },
      { line: 2, col: 0 },
    ],
  });
  p.move(e, a);
  p.advance((1000 * 4) / 60);
  expect(p.moving()?.at.x).toBeCloseTo(83.5, 9);
  expect(p.moving()?.at.y).toBeCloseTo(60, 9);
  expect(p.moving()?.facing).toBe("se");
  e.path[1] = { line: 99, col: 99 };
  p.advance((1000 * 12) / 60);
  expect(p.moving()).toBeNull();
  expect(p.directions.get(1)).toBe("sw");
  p.move(play.newEvent({ ...e, kind: play.Moved, side: 1 }), a);
  expect(p.moving()).toBeNull();
});

it("finishes a walk on the same boundary with individual frame intervals", () => {
  const { a } = units(),
    p = new MapPlayback(0);
  p.move(
    play.newEvent({
      kind: play.Moved,
      side: 0,
      unit: a.index,
      path: [
        { line: 0, col: 0 },
        { line: 1, col: 0 },
        { line: 2, col: 0 },
      ],
    }),
    a,
  );
  for (let i = 0; i < 16; i++) p.advance(1000 / 60);
  expect(p.busy()).toBe(false);
  expect(p.moving()).toBeNull();
});

it("freezes battle views and updates only copied fighters until phase change", () => {
  const { before, after, a, d } = views();
  const fighter = (u: play.UnitView, n: number) => ({
    unit: u,
    before: 6,
    after: n,
    expBefore: 0,
    expAfter: 1,
    altBefore: 0,
    altAfter: 0,
    destroyed: false,
  });
  const battle = play.newEvent({
    kind: play.Battle,
    from: a.at,
    to: d.at,
    battle: {
      attacker: fighter(a, 6),
      defender: fighter(d, 4),
      returnFire: false,
      pincer: 0,
      blocking: 0,
    },
  });
  const p = new MapPlayback(0);
  p.resolution(before, after, [
    battle,
    play.newEvent({ kind: play.PhaseChanged }),
  ]);
  expect(p.shown(after).units.get("0,1")?.size).toBe(4);
  expect(before.units.get("0,1")?.size).toBe(6);
  expect(p.shown(after).units.has("1,0")).toBe(false);
  expect(p.effects()).toHaveLength(1);
  if (battle.battle) battle.battle.defender.after = 0;
  p.advance(200);
  expect(p.effects()[0]?.frame).toBe("effect/explosion-02");
  expect(p.shown(after).units.get("0,1")?.size).toBe(4);
  p.advance(200);
  expect(p.busy()).toBe(false);
  expect(p.shown(after).units.has("1,0")).toBe(true);
});

it("reveals enemy moves immediately and uses static reduced battle effects", () => {
  const { before, after, a, d } = views();
  const p = new MapPlayback(0, true);
  p.move(
    play.newEvent({
      kind: play.Moved,
      side: 0,
      unit: 1,
      path: [a.at, { line: 1, col: 0 }],
    }),
    a,
  );
  expect(p.moving()).toBeNull();
  p.resolution(before, after, [
    play.newEvent({ kind: play.PhaseChanged }),
    play.newEvent({
      kind: play.Revealed,
      side: 1,
      unit: 2,
      path: [d.at, { line: 1, col: 0 }],
    }),
  ]);
  expect(p.moving()).toBeNull();
  expect(p.shown(after).units.has("1,0")).toBe(true);
  const fighter = (u: play.UnitView, n: number) => ({
    unit: u,
    before: 6,
    after: n,
    expBefore: 0,
    expAfter: 1,
    altBefore: 0,
    altAfter: 0,
    destroyed: n === 0,
  });
  p.resolution(before, after, [
    play.newEvent({
      kind: play.Battle,
      from: a.at,
      to: d.at,
      battle: {
        attacker: fighter(a, 6),
        defender: fighter(d, 0),
        returnFire: false,
        pincer: 0,
        blocking: 0,
      },
    }),
    play.newEvent({ kind: play.PhaseChanged }),
  ]);
  p.advance(16);
  expect(p.effects()[0]?.frame).toBe("effect/explosion-08");
  expect(p.effects()[0]?.alpha).toBe(1);
  p.advance(400);
  expect(p.effects()).toHaveLength(0);
  expect(p.busy()).toBe(false);
});

it("follows the native scene's individual losses and experience without changing rules views", () => {
  const { before, after, a, d } = views(),
    p = new MapPlayback(0);
  const fighter = (u: play.UnitView, n: number): play.Fighter => ({
    unit: u,
    before: 6,
    after: n,
    expBefore: 0,
    expAfter: 1,
    altBefore: 0,
    altAfter: 0,
    destroyed: false,
  });
  const b: play.BattleInfo = {
    attacker: fighter(a, 6),
    defender: fighter(d, 4),
    returnFire: false,
    pincer: 0,
    blocking: 0,
  };
  p.setScenes(true);
  p.resolution(before, after, [
    play.newEvent({ kind: play.Battle, from: a.at, to: d.at, battle: b }),
    play.newEvent({ kind: play.PhaseChanged }),
  ]);
  expect(p.scene()?.state.sides[0].shown).toBe(1);
  expect(p.shown(after).units.get("0,1")?.size).toBe(6);
  p.advance((97 * 1000) / 60);
  expect(p.shown(after).units.get("0,1")?.size).toBe(6);
  p.advance(1000 / 60);
  expect(p.shown(after).units.get("0,1")?.size).toBe(5);
  p.advance((28 * 1000) / 60);
  expect(p.shown(after).units.get("0,1")?.size).toBe(4);
  expect(p.shown(after).units.get("0,1")?.exp).toBe(0);
  p.advance((30 * 1000) / 60);
  expect(p.shown(after).units.get("0,1")?.exp).toBe(1);
  p.advance(((battleTicks(b, false) - 156) * 1000) / 60);
  expect(p.scene()).toBeNull();
  expect(p.busy()).toBe(false);
  expect(before.units.get("0,1")?.size).toBe(6);
});

it("keeps a dying unit until the map explosion peak and skips or reduces scenes safely", () => {
  const { before, a, d } = views(),
    after = structuredClone(before),
    p = new MapPlayback(0);
  after.units.delete("0,1");
  const fighter = (u: play.UnitView, n: number): play.Fighter => ({
    unit: u,
    before: 6,
    after: n,
    expBefore: 0,
    expAfter: 0,
    altBefore: 0,
    altAfter: 0,
    destroyed: n === 0,
  });
  const b: play.BattleInfo = {
    attacker: fighter(a, 6),
    defender: fighter(d, 0),
    returnFire: false,
    pincer: 0,
    blocking: 0,
  };
  const events = [
    play.newEvent({ kind: play.Battle, from: a.at, to: d.at, battle: b }),
    play.newEvent({ kind: play.PhaseChanged }),
  ];
  p.setScenes(true);
  p.resolution(before, after, events);
  p.advance((238 * 1000) / 60);
  expect(p.shown(after).units.get("0,1")?.size).toBe(0);
  p.advance((17 * 1000) / 60);
  expect(p.shown(after).units.has("0,1")).toBe(true);
  p.advance(1000 / 60);
  expect(p.shown(after).units.has("0,1")).toBe(false);
  expect(p.skipScene()).toBe(true);
  expect(p.busy()).toBe(false);
  p.resolution(before, after, events);
  p.setReducedMotion(true);
  expect(p.scene()?.state.sides[1].count).toBe(0);
  expect(p.shown(after).units.has("0,1")).toBe(false);
  p.advance((20 * 1000) / 60);
  expect(p.busy()).toBe(false);
  expect(p.skipScene()).toBe(false);
});

it("allows native cancel to skip the map battle wait when scene panels are off", () => {
  const { before, after, a, d } = views(),
    p = new MapPlayback(0);
  const fighter = (u: play.UnitView): play.Fighter => ({
    unit: u,
    before: 6,
    after: 6,
    expBefore: 0,
    expAfter: 0,
    altBefore: 0,
    altAfter: 0,
    destroyed: false,
  });
  p.resolution(before, after, [
    play.newEvent({
      kind: play.Battle,
      from: a.at,
      to: d.at,
      battle: {
        attacker: fighter(a),
        defender: fighter(d),
        returnFire: false,
        pincer: 0,
        blocking: 0,
      },
    }),
    play.newEvent({ kind: play.PhaseChanged }),
  ]);
  expect(p.scene()).toBeNull();
  expect(p.busy()).toBe(true);
  p.setScenes(false);
  expect(p.busy()).toBe(true);
  expect(p.skipScene()).toBe(true);
  expect(p.busy()).toBe(false);
  expect(p.shown(after)).toBe(after);
});

it("asks the views to look at a blown mine or a gas release once, when playback reaches it", () => {
  const { before, after, a, d } = views(),
    p = new MapPlayback(0),
    mine = { line: 5, col: 7 },
    gas = { line: 9, col: 3 };
  const fighter = (u: play.UnitView, n: number): play.Fighter => ({
    unit: u,
    before: 6,
    after: n,
    expBefore: 0,
    expAfter: 0,
    altBefore: 0,
    altAfter: 0,
    destroyed: false,
  });
  p.resolution(before, after, [
    play.newEvent({
      kind: play.Battle,
      from: a.at,
      to: d.at,
      battle: {
        attacker: fighter(a, 6),
        defender: fighter(d, 4),
        returnFire: false,
        pincer: 0,
        blocking: 0,
      },
    }),
    play.newEvent({ kind: play.Blown, to: mine }),
    play.newEvent({
      kind: play.Battle,
      from: a.at,
      to: d.at,
      battle: {
        attacker: fighter(a, 6),
        defender: fighter(d, 4),
        returnFire: false,
        pincer: 0,
        blocking: 0,
      },
    }),
    play.newEvent({ kind: play.Released, side: 1, to: gas }),
    play.newEvent({ kind: play.Wired, side: 1, to: { line: 1, col: 1 } }),
  ]);
  // The rules resolved the mine already; the views look only once the battle before it has played.
  expect(p.takeLook()).toBeNull();
  p.advance(399);
  expect(p.takeLook()).toBeNull();
  p.advance(1);
  expect(p.takeLook()).toEqual(mine);
  expect(p.takeLook()).toBeNull();
  p.advance(400);
  expect(p.takeLook()).toEqual(gas);
  expect(p.busy()).toBe(false);
  expect(p.takeLook()).toBeNull();
  p.resolution(before, after, [
    play.newEvent({ kind: play.Wired, side: 0, to: { line: 2, col: 2 } }),
    play.newEvent({ kind: play.Captured, to: { line: 3, col: 3 } }),
  ]);
  p.advance(1000);
  expect(p.takeLook()).toBeNull();
});
