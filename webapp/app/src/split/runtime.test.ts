import { expect, it, vi } from "vitest";
import { ai, hexgrid, level, play } from "@historyline/engine";
import { Builds, Full } from "../../../engine/src/testing/tables.ts";
import { playData, shippedLevel } from "../../../engine/src/testing/play.ts";
import { Camera } from "../camera.ts";
import { MapBattle, computerStage } from "../mapBattle.ts";
import { BattleRuntime } from "../battle/runtime.ts";
import { SplitRuntime } from "./runtime.ts";
import type { Layout } from "../storage/store.ts";

function fixture(layout: Layout = "original") {
  const peer = { map: { ready: false } };
  const panels = [{ map: { ready: true } }, peer];
  const runtime = Object.create(SplitRuntime.prototype) as SplitRuntime;
  Object.assign(runtime, {
    settings: { layout },
    request: { side: 0 },
    panels,
    local: null,
    stopped: false,
  });
  return { runtime, peer };
}

it("holds optional upgrades until every visible split map is ready", () => {
  const { runtime, peer } = fixture();
  expect(runtime.ready).toBe(false);
  peer.map.ready = true;
  expect(runtime.ready).toBe(true);
});

it("does not wait for a hidden map or call an empty handover screen playable", () => {
  const { runtime } = fixture("single");
  expect(runtime.ready).toBe(true);
  Object.assign(runtime, {
    local: { mode: "hotseat", activeSide: 1, handover: 1 },
  });
  expect(runtime.ready).toBe(false);
  Object.assign(runtime, {
    local: { mode: "hotseat", activeSide: 1, handover: null },
  });
  expect(runtime.ready).toBe(false);
});

it("shows the split screen in the automatic layout on wide landscape windows only", () => {
  const { runtime } = fixture("auto");
  try {
    vi.stubGlobal("window", { innerWidth: 1920, innerHeight: 1080 });
    expect(runtime.original).toBe(true);
    vi.stubGlobal("window", { innerWidth: 412, innerHeight: 915 });
    expect(runtime.original).toBe(false);
    Object.assign(runtime, { settings: { layout: "original" } });
    expect(runtime.original).toBe(true);
    Object.assign(runtime, { settings: { layout: "single" } });
    vi.stubGlobal("window", { innerWidth: 1920, innerHeight: 1080 });
    expect(runtime.original).toBe(false);
  } finally {
    vi.unstubAllGlobals();
  }
});

/** panels are the two sides' single-map runtimes on one session, each with its own playback and camera, without Pixi. */
function panels() {
  const session = play.newSession(shippedLevel("1914-08"), playData, {
      profile: Full,
    }),
    grid = level.grid(session.level),
    [width, height] = hexgrid.bounds(grid);
  const panel = (side: 0 | 1) => {
    const battle = new MapBattle(session, side),
      camera = new Camera({ width, height }, { width: 400, height: 300 }),
      runtime = Object.create(BattleRuntime.prototype) as BattleRuntime;
    Object.assign(runtime, {
      battle,
      map: { camera, dirty: false, setCursor: () => {}, render: () => {} },
      input: null,
      paused: false,
      stopped: false,
      cursorDraw: () => null,
    });
    return runtime;
  };
  const both: [BattleRuntime, BattleRuntime] = [panel(0), panel(1)],
    runtime = Object.create(SplitRuntime.prototype) as SplitRuntime;
  Object.assign(runtime, {
    settings: { layout: "original" },
    request: { side: 0 },
    panels: both,
    local: null,
    paused: false,
    stopped: false,
    messages: [],
    wrappers: [{ visible: true }, { visible: true }],
    surfaces: [{ hidden: false }, { hidden: false }],
    // The whole-cell alignment of the split views takes a camera's move as its intent on the next frame.
    alignViews: () => {},
  });
  const middle = {
    line: Math.trunc(grid.lines / 2),
    col: Math.trunc(grid.width / 2),
  };
  /** resolve plays events on both sides' playbacks, as a resolution reaches both panels. */
  const resolve = (events: play.Event[]) => {
    for (const p of both) {
      const view = p.battle.view();
      p.battle.playback.resolution(view, view, events);
    }
  };
  return { runtime, both, middle, resolve };
}
function centre(camera: Camera): [number, number] {
  const c = camera.screenToWorld(camera.view.width / 2, camera.view.height / 2);
  return [c.x, c.y];
}

it("turns both panels to a blown mine or a gas release when its event plays", () => {
  const { runtime, both, middle, resolve } = panels();
  resolve([play.newEvent({ kind: play.Blown, to: middle })]);
  runtime.tick(16);
  for (const p of both) {
    expect(centre(p.map.camera)[0]).toBeCloseTo(hexgrid.center(middle)[0], 6);
    expect(centre(p.map.camera)[1]).toBeCloseTo(hexgrid.center(middle)[1], 6);
  }
  both[1].map.camera.scroll(90, 60);
  const panned = centre(both[1].map.camera);
  runtime.tick(16);
  expect(centre(both[1].map.camera)).toEqual(panned);
  const gas = { line: middle.line - 2, col: middle.col - 2 };
  resolve([play.newEvent({ kind: play.Released, side: 0, to: gas })]);
  runtime.tick(16);
  for (const p of both) {
    expect(centre(p.map.camera)[0]).toBeCloseTo(hexgrid.center(gas)[0], 6);
    expect(centre(p.map.camera)[1]).toBeCloseTo(hexgrid.center(gas)[1], 6);
  }
});

it("keeps both panels' views on other events", () => {
  const { runtime, both, middle, resolve } = panels(),
    still = both.map((p) => centre(p.map.camera));
  resolve([
    play.newEvent({ kind: play.Wired, side: 0, to: middle }),
    play.newEvent({ kind: play.Captured, side: 0, to: middle }),
  ]);
  runtime.tick(16);
  expect(both.map((p) => centre(p.map.camera))).toEqual(still);
});

it("words the computer's well: planning with its bar, orders under way while it acts, waiting once it confirmed", async () => {
  const { runtime, both } = panels();
  Object.assign(runtime, { request: { side: 1 } });
  const session = both[0].battle.session;
  let release = () => {};
  const gate = new Promise<void>((resolve) => {
    release = resolve;
  });
  const player = new MapBattle(session, 1, {
    commands: (s, side, progress) => {
      progress?.(1, 4);
      const snapshot = play.clone(s, Builds);
      return gate.then(() => ai.commands({}, snapshot, side));
    },
  });
  player.stage = computerStage(both[0].battle, () => true);
  Object.assign(both[1], { battle: player });
  expect(runtime.well(0)).toEqual({
    text: "MOVEMENT PHASE",
    unit: null,
    progress: -1,
  });
  runtime.tick(16);
  expect(player.computerState()).toBe("planning");
  expect(runtime.well(0)).toEqual({
    text: "PLANNING THE MOVES",
    unit: null,
    progress: 0.25,
  });
  // The player's own well is not the computer's.
  expect(runtime.well(1).progress).toBe(-1);
  release();
  await new Promise((resolve) => setTimeout(resolve, 0));
  while (player.computerState() === "planning") runtime.tick(16);
  expect(runtime.well(0).text).toBe("ORDERS UNDER WAY");
  while (player.computerState() !== null) runtime.tick(16);
  expect(play.ready(session, 0)).toBe(true);
  expect(runtime.well(0)).toEqual({
    text: "WAITING: ENEMY ATTACKS",
    unit: null,
    progress: -1,
  });
});
