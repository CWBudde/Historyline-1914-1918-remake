import { expect, it } from "vitest";
import { layout } from "./geometry.ts";
import {
  fit,
  gameInfoScale,
  panelScale,
  plateRect,
  textScale,
  unitInfoLayout,
} from "./overlay.ts";
import { Painter, Tile, hexEdge } from "./painter.ts";
import type { Rect } from "../camera.ts";

const inside = (outer: Rect, inner: Rect) =>
  inner.x >= outer.x &&
  inner.y >= outer.y &&
  inner.x + inner.width <= outer.x + outer.width &&
  inner.y + inner.height <= outer.y + outer.height;

it("lays out the unit info as the desktop does, at its scale while it fits", () => {
  expect(unitInfoLayout({ x: 0, y: 0, width: 940, height: 990 }, 3).scale).toBe(
    3,
  );
  expect(
    unitInfoLayout({ x: 0, y: 0, width: 300, height: 280 }, 2).scale,
  ).toBeLessThan(2);
  for (const [w, h, s] of [
    [952, 1040, 3],
    [1270, 1380, 4],
    [300, 280, 2],
  ] as const) {
    const panel = { x: 10, y: 20, width: w, height: h },
      l = unitInfoLayout(panel, s);
    expect(inside(panel, l.plate)).toBe(true);
    for (const r of [l.photoWell, l.hex, l.name, l.stats])
      expect(inside(l.plate, r)).toBe(true);
    expect(inside(l.photoWell, l.photo)).toBe(true);
    expect(Math.abs(l.photo.width / l.photo.height - 100 / 84)).toBeLessThan(
      0.06,
    );
    expect(l.hex.x).toBeGreaterThanOrEqual(l.photoWell.x + l.photoWell.width);
    expect(l.name.y).toBeGreaterThanOrEqual(l.photoWell.y + l.photoWell.height);
    expect(l.stats.y).toBeGreaterThanOrEqual(l.name.y + l.name.height);
  }
});

it("scales the plates from the cell scale, inside the frame's teeth", () => {
  const s = layout(1920, 1080, 1),
    base = panelScale(s);
  expect(base).toBeCloseTo((s.fitScale * 80) / 24, 9);
  const room = plateRect(s, 0);
  expect(room.width).toBeLessThan(s.panels[0].width);
  // At 1080p the game info's 138 reference pixels of height fit the plate room at a smaller scale.
  const k = gameInfoScale(room, base);
  expect(k).toBeLessThanOrEqual(base);
  expect(138 * k).toBeLessThanOrEqual(room.height + 1e-9);
  expect(textScale(k)).toBe(Math.floor(k));
  const g = fit({ x: 0, y: 0, width: 100, height: 50 }, 2, 40, 10);
  expect([g.x(0), g.y(0), g.k]).toEqual([10, 15, 2]);
});

it("models a texture as its deviation and paints steel, bevels and hexes inside their rectangles", () => {
  const px = new Uint8ClampedArray(4 * 16 * 16);
  for (let i = 0; i < 256; i++) {
    px[4 * i] = 100 + (i % 2) * 50;
    px[4 * i + 1] = 120;
    px[4 * i + 2] = 80;
    px[4 * i + 3] = 255;
  }
  const t = new Tile({ width: 16, height: 16, data: px }),
    [dr, dg] = t.at(0, 0, 16);
  expect(dg).toBeCloseTo(0, 9);
  expect(Math.abs(dr)).toBeGreaterThan(0.1);
  const out = new Uint8ClampedArray(4 * 40 * 30),
    p = new Painter(out, 40, 30, 2, { steel: null, rivets: null });
  p.plate({ x: 5, y: 5, width: 20, height: 10 }, 0.46);
  expect(out[4 * (4 * 40 + 4) + 3]).toBe(0);
  expect(out[4 * (10 * 40 + 10) + 3]).toBe(255);
  // The lit top bevel is lighter than the shaded bottom one.
  expect(out[4 * (5 * 40 + 15)] ?? 0).toBeGreaterThan(
    out[4 * (14 * 40 + 15)] ?? 0,
  );
  p.namePlate({ x: 0, y: 16, width: 40, height: 14 });
  expect(out[4 * (23 * 40 + 20) + 3]).toBe(255);
  // The centre lies nearer the slanted edges than the top: 13·12 / hypot(12, 26/3).
  expect(hexEdge(0, 0, 13, 12)[0]).toBeCloseTo(156 / Math.hypot(12, 26 / 3), 9);
  expect(hexEdge(20, 0, 13, 12)[0]).toBeLessThan(0);
});
