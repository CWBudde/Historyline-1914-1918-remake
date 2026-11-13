import { expect, it } from "vitest";
import {
  battleLayout,
  figurePoint,
  hexCentre,
  hexScale,
  scaledLayout,
  type SideRects,
} from "./layout.ts";

it("preserves native 152x180 proportions and mirrored fighter plates", () => {
  const l = battleLayout({ x: 0, y: 0, width: 320, height: 400 }, 2);
  expect(l.scale).toBe(2);
  expect(l.sides[0].photo).toEqual({ x: 22, y: 34, width: 152, height: 128 });
  expect(l.sides[1].photo).toEqual({ x: 146, y: 214, width: 152, height: 128 });
  expect(l.sides[0].hex).toEqual({ x: 182, y: 52, width: 124, height: 116 });
  expect(l.sides[1].hex).toEqual({ x: 14, y: 232, width: 124, height: 116 });
  expect(figurePoint(l.sides[0], 0, false)).toEqual({ x: 244, y: 80 });
  expect(figurePoint(l.sides[0], 5, true)).toEqual({ x: 264, y: 114 });
});
it("fits every plate within narrow or offset viewports and rounds to device pixels", () => {
  for (const rect of [
    { x: 4, y: 8, width: 286, height: 260 },
    { x: 0, y: 0, width: 240, height: 190 },
  ]) {
    const l = battleLayout(rect, Infinity, 2);
    for (const side of l.sides)
      for (const r of Object.values({ ...side })) {
        expect(r.x).toBeGreaterThanOrEqual(rect.x - 0.5);
        expect(r.y).toBeGreaterThanOrEqual(rect.y - 0.5);
        expect(r.x + r.width).toBeLessThanOrEqual(rect.x + rect.width + 0.5);
        expect(r.y + r.height).toBeLessThanOrEqual(rect.y + rect.height + 0.5);
        expect(r.x * 2).toBe(Math.round(r.x * 2));
      }
  }
});
it("puts the seam's rivets between the halves, every 22 reference pixels", () => {
  const l = battleLayout({ x: 0, y: 0, width: 304, height: 360 });
  expect(l.scale).toBe(2);
  expect(l.rivets.map((p) => p.x)).toEqual([20, 64, 108, 152, 196, 240, 284]);
  for (const p of l.rivets) {
    expect(p.y).toBe(180);
    expect(p.y).toBeGreaterThan(
      l.sides[0].half.y + l.sides[0].half.height - 0.5,
    );
    expect(p.y).toBeLessThan(l.sides[1].half.y + 0.5);
  }
});
it("fills a host of the scene's proportions edge to edge at any size", () => {
  for (const width of [306, 366, 675, 1000]) {
    const height = Math.round((width * 180) / 152),
      l = battleLayout({ x: 0, y: 0, width, height });
    // The halves reach within two reference pixels of every edge: no band is left over.
    const top = l.sides[0].half,
      bottom = l.sides[1].half;
    expect(top.x).toBeLessThanOrEqual(2 * l.scale + 1);
    expect(top.y).toBeLessThanOrEqual(2 * l.scale + 1);
    expect(width - (top.x + top.width)).toBeLessThanOrEqual(2 * l.scale + 1);
    expect(height - (bottom.y + bottom.height)).toBeLessThanOrEqual(
      2 * l.scale + 1,
    );
  }
});
it("lays out in device pixels and scales back to CSS pixels on the same pixels", () => {
  const dpr = 2.625,
    device = battleLayout({ x: 0, y: 0, width: 1024, height: 1213 }),
    css = scaledLayout(device, 1 / dpr);
  expect(css.scale).toBeCloseTo(device.scale / dpr);
  for (const s of [0, 1] as const)
    for (const name of Object.keys(css.sides[s]) as (keyof SideRects)[]) {
      const r = css.sides[s][name],
        d = device.sides[s][name];
      expect(r.x * dpr).toBeCloseTo(d.x);
      expect(r.width * dpr).toBeCloseTo(d.width);
      expect(Number.isInteger(d.x) && Number.isInteger(d.width)).toBe(true);
    }
  expect(hexScale(css.sides[0])).toBeCloseTo(hexScale(device.sides[0]) / dpr);
  expect(hexCentre(css.sides[1]).x * dpr).toBeCloseTo(
    hexCentre(device.sides[1]).x,
  );
});
