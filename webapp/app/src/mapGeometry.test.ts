import { expect, it } from "vitest";
import { hexgrid } from "@historyline/engine";
import {
  chooseMip,
  backdropDraw,
  visibleHexes,
  intersects,
} from "./mapGeometry.ts";
import type { Pyramid } from "./assetManifest.ts";

it("culls hexes conservatively in world coordinates, including odd lines", () => {
  const g = { width: 13, lines: 17 };
  for (const rect of [
    { x: 200, y: 140, width: 210, height: 150 },
    { x: -200, y: -180, width: 300, height: 280 },
    { x: 1400, y: 600, width: 400, height: 220 },
  ]) {
    const expected = [];
    const padded = {
      x: rect.x - 96,
      y: rect.y - 96,
      width: rect.width + 192,
      height: rect.height + 192,
    };
    for (let line = 0; line < g.lines; line++)
      for (let col = 0; col < g.width - (line % 2); col++) {
        const c = { line, col };
        const [x, y] = hexgrid.center(c);
        if (intersects({ x: x - 48, y: y - 40, width: 96, height: 80 }, padded))
          expected.push(c);
      }
    expect(visibleHexes(g, rect)).toEqual(expected);
  }
});

it("selects the coarsest non-magnified backdrop mip at effective DPR scale", () => {
  const p: Pyramid = {
    width: 4096,
    height: 2048,
    tileSize: 512,
    mips: [1, 2, 4, 8].map((scale) => ({
      scale,
      width: 4096 / scale,
      height: 2048 / scale,
      tiles: [],
    })),
  };
  for (const [eff, index] of [
    [0.25, 3],
    [0.5, 2],
    [1, 1],
    [1.001, 0],
    [4, 0],
  ])
    expect(chooseMip(p, { width: 2048, height: 1024 }, eff ?? 0)).toBe(index);
});

it("clips rounded mip edges while preserving gutter texture coordinates", () => {
  const file = { path: "", sha256: "", bytes: 0 };
  const tile = {
    x: 0,
    y: 0,
    width: 257,
    height: 129,
    sourceX: 0,
    sourceY: 0,
    sourceWidth: 257,
    sourceHeight: 129,
    file,
  };
  const mip = { scale: 4, width: 257, height: 129, tiles: [tile] };
  const p: Pyramid = { width: 1027, height: 515, tileSize: 512, mips: [mip] };
  const draw = backdropDraw(p, mip, tile, { width: 2054, height: 1030 });
  expect(draw.world).toEqual({ x: 0, y: 0, width: 2054, height: 1030 });
  expect(draw.frame).toEqual({ x: 0, y: 0, width: 256.75, height: 128.75 });
  const edge = {
    ...tile,
    x: 512,
    y: 512,
    width: 3,
    height: 3,
    sourceX: 510,
    sourceY: 510,
    sourceWidth: 5,
    sourceHeight: 5,
  };
  const other = backdropDraw(
    { ...p, width: 515, height: 515 },
    { ...mip, scale: 1 },
    edge,
    { width: 1030, height: 1030 },
  );
  expect(other.world).toEqual({ x: 1024, y: 1024, width: 6, height: 6 });
  expect(other.frame).toEqual({ x: 2, y: 2, width: 3, height: 3 });
});
