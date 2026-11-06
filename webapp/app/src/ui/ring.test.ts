import { expect, it } from "vitest";
import {
  placement,
  ringActions,
  layoutRing,
  hitRing,
  type Action,
} from "./ring.ts";

it("keeps native clockwise map and contents actions with disabled actions omitted", () => {
  expect(ringActions(placement(false, false), () => true)).toEqual([
    "move",
    "map",
    "info",
    "mode",
  ]);
  expect(ringActions(placement(true, true), () => true)).toEqual([
    "attack",
    "map",
    "info",
    "open",
  ]);
  expect(
    ringActions(placement(false, true, true), (a) => a !== "repair"),
  ).toEqual(["exit", "produce"]);
});

it("uses native plate spacing and selects hex plates rather than bounding boxes", () => {
  const ring = layoutRing(
    ["move", "map", "info", "mode"],
    { x: 600, y: 450 },
    { x: 0, y: 0, width: 1200, height: 900 },
  );
  expect(ring.scale).toBe(1);
  expect(ring.items.map((i) => [i.x, i.y])).toEqual([
    [600, 346],
    [704, 450],
    [600, 554],
    [496, 450],
  ]);
  expect(hitRing(ring, { x: 600, y: 346 })).toBe("move");
  expect(hitRing(ring, { x: 647, y: 307 })).toBeNull();
  expect(hitRing(ring, { x: 600, y: 450 })).toBe("cancel");
});

it("fits all actions within phone and tiny viewports, including near corners", () => {
  const actions: Action[] = ["move", "map", "info", "open"];
  for (const [width, height] of [
    [390, 420],
    [240, 160],
    [50, 40],
  ]) {
    const viewport = { x: 12, y: 64, width: width ?? 1, height: height ?? 1 };
    for (const point of [
      { x: 12, y: 64 },
      { x: 402, y: 484 },
    ]) {
      const ring = layoutRing(actions, point, viewport, 2);
      for (const item of ring.items) {
        for (const p of item.points) {
          expect(p.x).toBeGreaterThanOrEqual(viewport.x - 1e-9);
          expect(p.x).toBeLessThanOrEqual(viewport.x + viewport.width + 1e-9);
          expect(p.y).toBeGreaterThanOrEqual(viewport.y - 1e-9);
          expect(p.y).toBeLessThanOrEqual(viewport.y + viewport.height + 1e-9);
        }
        expect(hitRing(ring, item)).toBe(item.action);
      }
    }
  }
});
