import { expect, it, vi } from "vitest";
import { ai, hexgrid, level, play } from "@historyline/engine";
import { Builds, Full } from "../../../engine/src/testing/tables.ts";
import { playData, shippedLevel } from "../../../engine/src/testing/play.ts";
import { Camera } from "../camera.ts";
import { MapBattle, actRestMs, computerStage } from "../mapBattle.ts";
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
    slides: [null, null],
    // A panel shows the hexes in its camera's rectangle.
    panelShows: (side: 0 | 1, at: hexgrid.Coord) => {
      const r = both[side].map.camera.visible(),
        [x, y] = hexgrid.center(at);
      return x >= r.x && y >= r.y && x < r.x + r.width && y < r.y + r.height;
    },
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

/** longest is side's movable unit and its farthest destination by route. */
function longest(session: play.Session, side: number) {
  let best: { u: play.UnitView; to: hexgrid.Coord; n: number } | null = null;
  for (const u of play.view(session, side).units.values()) {
    if (u.side !== side || u.head !== null) continue;
    try {
      for (const to of play.coords(play.destinations(session, side, u.index))) {
        const n = play.route(session, side, u.index, to).length;
        if (best === null || n > best.n) best = { u, to, n };
      }
    } catch {
      /* Units without a legal movement order are skipped. */
    }
  }
  if (best === null) throw Error("no movable unit");
  return best;
}
function shows(camera: Camera, at: hexgrid.Coord): boolean {
  const r = camera.visible(),
    [x, y] = hexgrid.center(at);
  return x >= r.x && y >= r.y && x < r.x + r.width && y < r.y + r.height;
}

it("slides the computer's panel to its walker and follows the walk, the player's panel stays", () => {
  const { runtime, both } = panels();
  Object.assign(runtime, { request: { side: 1 } });
  const session = both[0].battle.session,
    { u, to } = longest(session, 0),
    camera = both[0].map.camera;
  // The computer's panel looks at the far corner of the map from the unit.
  const [ux, uy] = hexgrid.center(u.at);
  camera.centerOn(
    ux < camera.world.width / 2 ? camera.world.width : 0,
    uy < camera.world.height / 2 ? camera.world.height : 0,
  );
  expect(shows(camera, u.at)).toBe(false);
  const player = centre(both[1].map.camera);
  expect(
    both[0].battle.act(play.move(session, 0, u.index, to), actRestMs),
  ).toBe(true);
  // It slides there while the move rests, not in one jump.
  runtime.tick(16);
  expect(shows(camera, u.at)).toBe(false);
  for (let ms = 16; ms < actRestMs; ms += 16) runtime.tick(16);
  expect(shows(camera, u.at)).toBe(true);
  // The walk stays in sight to its end.
  while (both[0].battle.playback.walking()) {
    runtime.tick(16);
    const w = both[0].battle.playback.moving();
    if (w !== null) {
      const [at] = hexgrid.pixelToHex(
        level.grid(session.level),
        w.at.x,
        w.at.y,
      );
      // A slide may take its few frames to catch up with the walker.
      if (!shows(camera, at)) for (let i = 0; i < 9; i++) runtime.tick(16);
    }
  }
  for (let i = 0; i < 9; i++) runtime.tick(16);
  expect(shows(camera, to)).toBe(true);
  expect(centre(both[1].map.camera)).toEqual(player);
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
