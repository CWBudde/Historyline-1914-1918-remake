import { expect, it } from "vitest";
import {
  cursorLook,
  glyph,
  inPlate,
  inset,
  plateCorners,
  plateLayers,
  ringLayers,
  type Pt,
} from "./gesture.ts";
import type { Action } from "./ui/ring.ts";

const area = (p: readonly Pt[]) =>
  p.reduce((a, [x, y], i) => {
    const [nx, ny] = p[(i + 1) % p.length] ?? [0, 0];
    return a + x * ny - nx * y;
  }, 0) / 2;
const inside = (p: readonly Pt[], [x, y]: Pt) => {
  let c = false;
  for (let i = 0, j = p.length - 1; i < p.length; j = i++) {
    const [xi, yi] = p[i] ?? [0, 0],
      [xj, yj] = p[j] ?? [0, 0];
    if (yi > y !== yj > y && x < ((xj - xi) * (y - yi)) / (yj - yi) + xi)
      c = !c;
  }
  return c;
};

it("lays plates on the cell polygon and insets every edge by the same distance", () => {
  const hex = plateCorners(100, 50, 2);
  expect(hex[0]).toEqual([4, 50]);
  expect(hex[3]).toEqual([196, 50]);
  expect(area(hex)).toBeGreaterThan(0);
  const small = inset(hex, 6);
  // The left slant (corners 0 to 1) moved 6 px inwards along its normal.
  const [ax, ay] = small[0] ?? [0, 0],
    [hx, hy] = hex[0] ?? [0, 0],
    [bx, by] = hex[1] ?? [0, 0];
  expect(
    Math.abs((by - hy) * (ax - hx) - (bx - hx) * (ay - hy)) /
      Math.hypot(bx - hx, by - hy),
  ).toBeCloseTo(6, 9);
  expect(small[1]?.[1]).toBeCloseTo(50 - 80 + 6, 6);
  expect(inPlate(100, 50, 100, 50, 2)).toBe(true);
  expect(inPlate(195, 50, 100, 50, 2)).toBe(true);
  expect(inPlate(190, 15, 100, 50, 2)).toBe(false);
});

it("draws the ring open inside and the plate solid, every glyph hole inside the part before it", () => {
  const ring = ringLayers(0, 0, 1);
  expect(ring[0]?.polys.map((p) => p.hole)).toEqual([false, true]);
  expect(ring).toHaveLength(1 + 6 * 10);
  for (const a of [
    "cancel",
    "move",
    "attack",
    "mode",
    "open",
    "map",
    "info",
    "exit",
    "produce",
    "repair",
  ] as Action[]) {
    const g = glyph(a);
    expect(g.length).toBeGreaterThan(0);
    for (const [i, p] of g.entries()) {
      for (const [x, y] of p.pts) {
        expect(Math.abs(x)).toBeLessThanOrEqual(0.5 + 1e-9);
        expect(Math.abs(y)).toBeLessThanOrEqual(0.5 + 1e-9);
      }
      if (!p.hole) continue;
      const owner = [...g.slice(0, i)].reverse().find((q) => !q.hole);
      expect(owner).toBeDefined();
      const cx = p.pts.reduce((s, q) => s + q[0], 0) / p.pts.length,
        cy = p.pts.reduce((s, q) => s + q[1], 0) / p.pts.length;
      expect(inside(owner?.pts ?? [], [cx, cy])).toBe(true);
    }
    for (const s of ["normal", "pressed", "disabled"] as const)
      expect(plateLayers(0, 0, 1, a, s).length).toBeGreaterThan(3);
  }
});

it("turns the ring into the plate of the held or aimed action as the desktop does", () => {
  // Nothing aimed, fire up: the plain ring; fire held: the picked plate, faded where not possible.
  expect(cursorLook("cancel", false, false, true, "cancel", true)).toBeNull();
  expect(cursorLook("cancel", false, true, true, "cancel", true)).toEqual({
    action: "cancel",
    state: "normal",
  });
  expect(cursorLook("cancel", false, true, false, "info", true)).toEqual({
    action: "info",
    state: "normal",
  });
  expect(cursorLook("cancel", false, true, false, "attack", false)).toEqual({
    action: "attack",
    state: "disabled",
  });
  // An aimed order: its plate on cells in reach, the ring elsewhere; pressed while fire is held there.
  expect(cursorLook("move", true, false, true, "cancel", true)).toEqual({
    action: "move",
    state: "normal",
  });
  expect(cursorLook("move", false, false, true, "cancel", true)).toBeNull();
  expect(cursorLook("move", true, true, true, "cancel", true)).toEqual({
    action: "move",
    state: "pressed",
  });
  expect(cursorLook("attack", false, true, true, "cancel", true)).toEqual({
    action: "attack",
    state: "disabled",
  });
});
