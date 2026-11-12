import { expect, it } from "vitest";
import { hexgrid } from "@historyline/engine";
import { Strategic, blue, sideShade, spentShade, yellow } from "./strategic.ts";

it("fits the strategic map into its picture at a whole scale, framed, and maps pixels back to the world", () => {
  const grid = { width: 13, lines: 34 },
    s = new Strategic(grid, 470, 500);
  // 25 blocks across and 17.5 rows of 4-px cells fit at scale 4 into 470 × 500 less the frame.
  expect(s.scale).toBe(Math.min(Math.trunc(462 / 100), Math.trunc(492 / 70)));
  expect(s.map.width).toBe(25 * 4 * s.scale);
  expect(s.map.x).toBe(Math.trunc((470 - s.map.width) / 2));
  expect(s.toWorld(s.map.x - 1, s.map.y)).toBeNull();
  const w = s.toWorld(s.map.x, s.map.y);
  expect(w).toEqual({ x: (hexgrid.hexWidth - hexgrid.halfPitch) / 2, y: 0 });
  const img = s.render({
    terrain: () => [10, 200, 10],
    units: [{ at: { line: 4, col: 2 }, shade: blue }],
    buildings: [{ at: { line: 10, col: 5 }, kind: "hq", shade: yellow }],
    view: { x: 0, y: 0, width: 400, height: 300 },
  });
  const px = (x: number, y: number) =>
    Array.from(
      img.data.slice(4 * (y * img.width + x), 4 * (y * img.width + x) + 3),
    );
  expect(px(0, 0)).toEqual([48, 48, 36]);
  expect(px(s.map.x - 4, s.map.y + 50)).toEqual([81, 81, 65]);
  const b = s.block({ line: 30, col: 1 });
  expect(px(s.map.x + b.x + 1, s.map.y + b.y + 1)).toEqual([10, 200, 10]);
  const v = s.viewRect({ x: 0, y: 0, width: 400, height: 300 });
  expect(px(v.x, v.y + 5)).toEqual([154, 154, 138]);
});

it("greys a unit with its order and colours the sides", () => {
  expect(sideShade(1)).toBe(blue);
  expect(sideShade(2)).toBe(yellow);
  const g = spentShade(blue);
  expect(Math.abs(g.lit[0] - g.lit[2])).toBeLessThan(
    Math.abs(blue.lit[0] - blue.lit[2]),
  );
});
