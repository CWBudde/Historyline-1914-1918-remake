import { expect, it } from "vitest";
import { hexgrid } from "@historyline/engine";
import { Camera } from "../camera.ts";
import { layout } from "./geometry.ts";
import {
  alignCamera,
  cameraRest,
  centreView,
  clampView,
  pairRange,
  rowRange,
} from "./camera.ts";

it("pans away from a native plate edge without the generic world clamp jumping the camera", () => {
  const grid = { width: 8, lines: 44 },
    screen = layout(1280, 720),
    [pair] = pairRange(grid, screen.cols).slice(-1),
    [width, height] = hexgrid.bounds(grid),
    camera = new Camera({ width, height }, { width: 632, height: 680 });
  if (pair === undefined) throw Error("Missing native edge");
  alignCamera(camera, screen.view(0, pair, 0), 1);
  const before = { ...camera.origin };
  camera.scroll(-10, 0);
  expect(camera.origin.x).toBeCloseTo(before.x - 10 / camera.effective(), 9);
  expect(camera.origin.y).toBe(before.y);
});

it.each([
  [8, 24, 7, 7, [0, 4], [0, 5]],
  [8, 24, 8, 7, [0, 4], [0, 5]],
  [4, 14, 7, 7, [0, 0], [0, 0]],
  [8, 23, 7, 7, [0, 4], [0, 5]],
  [3, 6, 13, 7, [-2, -2], [-2, -2]],
  [3, 6, 6, 4, [-1, -1], [-1, -1]],
  [40, 12, 11, 8, [0, 34], [-1, -1]],
  [4, 80, 9, 8, [-1, -1], [0, 32]],
])(
  "keeps native bounds for grid %dx%d and block %dx%d",
  (width, lines, cols, rows, pairs, rowBounds) => {
    const grid = { width, lines };
    expect(pairRange(grid, cols)).toEqual(pairs);
    expect(rowRange(grid, rows)).toEqual(rowBounds);
  },
);

it("clamps both axes while retaining the centred negative origin of small maps", () => {
  const grid = { width: 8, lines: 24 };
  expect(clampView(grid, 7, 7, -3, -9)).toEqual({ pair: 0, row: 0 });
  expect(clampView(grid, 7, 7, 2, 3)).toEqual({ pair: 2, row: 3 });
  expect(clampView(grid, 7, 7, 9, 99)).toEqual({ pair: 4, row: 5 });
  expect(clampView({ width: 3, lines: 6 }, 13, 7, 99, 99)).toEqual({
    pair: -2,
    row: -2,
  });
});

it("centres odd and even hex lines using native integer middle columns", () => {
  const grid = { width: 8, lines: 24 };
  expect(centreView(grid, 8, 6, { line: 10, col: 4 })).toEqual({
    pair: 2,
    row: 3,
  });
  expect(centreView(grid, 8, 6, { line: 11, col: 4 })).toEqual({
    pair: 3,
    row: 3,
  });
  expect(centreView(grid, 7, 7, { line: 0, col: 0 })).toEqual({
    pair: 0,
    row: 0,
  });
  expect(centreView(grid, 7, 7, { line: 23, col: 6 })).toEqual({
    pair: 4,
    row: 5,
  });
});

it("keeps every centred map hex visible for native block sizes", () => {
  const grid = { width: 8, lines: 24 };
  for (const [cols, rows] of [
    [7, 7],
    [8, 6],
    [9, 7],
    [12, 9],
  ]) {
    if (cols === undefined || rows === undefined) throw Error("Missing size");
    for (let line = 0; line < grid.lines; line++) {
      for (let col = 0; col < grid.width - (line & 1); col++) {
        const at = { line, col },
          rest = centreView(grid, cols, rows, at),
          j = 2 * (col - rest.pair) + (line & 1),
          i = (line >> 1) - rest.row;
        expect(j).toBeGreaterThanOrEqual(0);
        expect(j).toBeLessThan(cols);
        expect(i).toBeGreaterThanOrEqual(0);
        expect(i).toBeLessThan(rows);
      }
    }
  }
});

it.each([1, 2])("uses panel-local physical pixels at DPR %s", (dpr) => {
  const screen = layout(1280, 720);
  for (const side of [0, 1] as const) {
    const view = screen.view(side, 2, 3),
      camera = new Camera(
        { width: 5000, height: 5000 },
        { width: 1, height: 1 },
      );
    alignCamera(camera, view, dpr);
    expect(camera.scale).toBe(dpr);
    expect(camera.zoom).toBe(1 / dpr);
    expect(camera.view).toEqual({ width: 632, height: 680 });
    expect(camera.origin).toEqual({ x: 264, y: 240 });
    expect(camera.worldToScreen(426, 400)).toEqual({ x: 162, y: 160 });
    for (const [x, y] of [
      [0, 0],
      [426, 400],
      [1234.5, 987.25],
    ] as const) {
      const native = view.worldToScreen(x, y),
        local = camera.worldToScreen(x, y);
      expect(local.x).toBeCloseTo(native.x - view.panel.x, 10);
      expect(local.y).toBeCloseTo(native.y - view.panel.y, 10);
      const world = camera.screenToWorld(local.x, local.y);
      expect(world.x).toBeCloseTo(x, 10);
      expect(world.y).toBeCloseTo(y, 10);
    }
  }
});

it("preserves fractional zoom and slim shifts without generic camera clamping", () => {
  const screen = layout(1920, 1080, 2).zoomed(9),
    view = screen.view(1, -2, -3),
    camera = new Camera({ width: 10, height: 10 }, { width: 1, height: 1 });
  view.shiftX = 0.25;
  view.shiftY = -1.5;
  alignCamera(camera, view, 2);
  expect(camera.effective()).toBeCloseTo(view.scale, 12);
  expect(camera.world).toEqual({ width: 10, height: 10 });
  for (const [x, y] of [
    [0, 0],
    [431.5, 217.25],
  ] as const) {
    const native = view.worldToScreen(x, y),
      local = camera.worldToScreen(x, y);
    expect(local.x).toBeCloseTo(native.x - view.panel.x, 10);
    expect(local.y).toBeCloseTo(native.y - view.panel.y, 10);
  }
});

it("snaps a panned camera to whole pairs and rows, then clamps to native bounds", () => {
  const screen = layout(1280, 720),
    grid = { width: 20, lines: 40 },
    camera = new Camera({ width: 5000, height: 5000 }, { width: 1, height: 1 });
  alignCamera(camera, screen.view(1, 2, 3), 2);
  camera.origin.x += hexgrid.colPitch / 2;
  camera.origin.y += hexgrid.hexHeight / 2;
  expect(cameraRest(camera, screen, 1, grid)).toEqual({ pair: 3, row: 4 });
  camera.origin = { x: -10000, y: -10000 };
  expect(cameraRest(camera, screen, 1, grid)).toEqual({ pair: 0, row: 0 });
  camera.origin = { x: 10000, y: 10000 };
  expect(cameraRest(camera, screen, 1, grid)).toEqual({ pair: 16, row: 12 });
});

it.each([1, 2])(
  "restores native camera rests after alignment at DPR %s",
  (dpr) => {
    for (const screen of [layout(1280, 720), layout(1920, 1080, 2).zoomed(9)]) {
      for (const side of [0, 1] as const) {
        for (const [grid, pair, row] of [
          [{ width: 40, lines: 80 }, 3, 4],
          [{ width: 1, lines: 1 }, -2, -4],
        ] as const) {
          const rest = clampView(grid, screen.cols, screen.rows, pair, row),
            view = screen.view(side, rest.pair, rest.row),
            camera = new Camera(
              { width: 10, height: 10 },
              { width: 1, height: 1 },
            );
          alignCamera(camera, view, dpr);
          expect(cameraRest(camera, screen, side, grid)).toEqual(rest);
          if (grid.width === 1) {
            expect(rest.pair).toBeLessThan(0);
            expect(rest.row).toBeLessThan(0);
          }
        }
      }
    }
  },
);
