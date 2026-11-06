import { expect, it } from "vitest";
import { battleLayout, figurePoint } from "./layout.ts";

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
