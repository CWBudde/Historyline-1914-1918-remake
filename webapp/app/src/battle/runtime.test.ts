import { expect, it } from "vitest";
import { hexgrid, level, play } from "@historyline/engine";
import { Full } from "../../../engine/src/testing/tables.ts";
import { playData, shippedLevel } from "../../../engine/src/testing/play.ts";
import { Camera } from "../camera.ts";
import { MapBattle } from "../mapBattle.ts";
import { BattleRuntime } from "./runtime.ts";

/** panel is a single-map runtime around a real battle and camera, without Pixi. */
function panel() {
  const session = play.newSession(shippedLevel("1914-08"), playData, {
      profile: Full,
    }),
    grid = level.grid(session.level),
    [width, height] = hexgrid.bounds(grid),
    battle = new MapBattle(session, 0),
    camera = new Camera({ width, height }, { width: 400, height: 300 }),
    map = { camera, dirty: false },
    runtime = Object.create(BattleRuntime.prototype) as BattleRuntime;
  Object.assign(runtime, {
    battle,
    map,
    input: null,
    paused: false,
    stopped: false,
  });
  const middle = {
    line: Math.trunc(grid.lines / 2),
    col: Math.trunc(grid.width / 2),
  };
  return { runtime, battle, camera, map, middle };
}
function centre(camera: Camera): [number, number] {
  const c = camera.screenToWorld(camera.view.width / 2, camera.view.height / 2);
  return [c.x, c.y];
}

it("looks at a blown mine or a gas release when playback shows it, and leaves the player's panning alone", () => {
  const { runtime, battle, camera, map, middle } = panel(),
    view = battle.view();
  battle.playback.resolution(view, view, [
    play.newEvent({ kind: play.Blown, to: middle }),
  ]);
  runtime.tick(16, false);
  expect(centre(camera)[0]).toBeCloseTo(hexgrid.center(middle)[0], 6);
  expect(centre(camera)[1]).toBeCloseTo(hexgrid.center(middle)[1], 6);
  expect(map.dirty).toBe(true);
  camera.scroll(-120, -80);
  const panned = centre(camera);
  runtime.tick(16, false);
  expect(centre(camera)).toEqual(panned);
  const gas = { line: middle.line + 2, col: middle.col - 3 };
  battle.playback.resolution(view, view, [
    play.newEvent({ kind: play.Released, side: 1, to: gas }),
  ]);
  runtime.tick(16, false);
  expect(centre(camera)[0]).toBeCloseTo(hexgrid.center(gas)[0], 6);
  expect(centre(camera)[1]).toBeCloseTo(hexgrid.center(gas)[1], 6);
});

it("keeps the camera where it is on other events", () => {
  const { runtime, battle, camera, middle } = panel(),
    view = battle.view(),
    still = centre(camera);
  battle.playback.resolution(view, view, [
    play.newEvent({ kind: play.Wired, side: 1, to: middle }),
    play.newEvent({ kind: play.Captured, side: 1, to: middle }),
    play.newEvent({ kind: play.PhaseChanged }),
  ]);
  runtime.tick(16, false);
  expect(centre(camera)).toEqual(still);
});
