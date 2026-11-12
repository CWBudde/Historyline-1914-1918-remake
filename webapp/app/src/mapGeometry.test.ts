import { expect, it } from "vitest";
import { hexgrid } from "@historyline/engine";
import {
  chooseMip,
  backdropMips,
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

it("keeps the lite backdrop until every visible detail tile is decoded, then uses full DPR", () => {
  const p: Pyramid = {
    width: 512,
    height: 256,
    tileSize: 128,
    mips: [1, 2, 4].map((scale) => {
      const width = 512 / scale,
        height = 256 / scale;
      const tiles = [];
      for (let y = 0; y < height; y += 128)
        for (let x = 0; x < width; x += 128) {
          const tw = Math.min(128, width - x),
            th = Math.min(128, height - y);
          const sourceX = Math.max(0, x - 2),
            sourceY = Math.max(0, y - 2);
          tiles.push({
            x,
            y,
            width: tw,
            height: th,
            sourceX,
            sourceY,
            sourceWidth: Math.min(width, x + tw + 2) - sourceX,
            sourceHeight: Math.min(height, y + th + 2) - sourceY,
            file: { path: `${scale}/${x}/${y}`, sha256: "", bytes: 1 },
          });
        }
      return { scale, width, height, tiles };
    }),
  };
  const world = { width: 256, height: 128 },
    rect = { x: 0, y: 0, width: 120, height: 120 };
  expect(
    backdropMips(p, world, 2, rect, false, () => {
      throw new Error("detail queried before playable frame");
    }),
  ).toEqual([1]);
  expect(backdropMips(p, world, 2, rect, true, () => false)).toEqual([1, 0]);
  expect(
    backdropMips(p, world, 2, rect, true, (_mip, tile) => tile.x < 256),
  ).toEqual([0]);
  expect(
    backdropMips(
      p,
      world,
      2,
      { ...rect, x: 180, width: 60 },
      true,
      (_mip, tile) => tile.x < 256,
    ),
  ).toEqual([1, 0]);
  expect(backdropMips(p, world, 1, rect, true, () => false)).toEqual([1]);
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
