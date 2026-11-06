import { expect, it } from "vitest";
import { Camera } from "./camera.ts";

it("matches the native camera bounds and round trips", () => {
  const c = new Camera(
    { width: 1000, height: 800 },
    { width: 400, height: 300 },
  );
  c.scroll(-50, -50);
  expect(c.origin).toEqual({ x: 0, y: 0 });
  c.scroll(5000, 5000);
  expect(c.origin).toEqual({ x: 600, y: 500 });
  const p = c.screenToWorld(123, 45);
  expect(c.worldToScreen(p.x, p.y)).toEqual({ x: 123, y: 45 });
  const small = new Camera(
    { width: 100, height: 100 },
    { width: 400, height: 300 },
  );
  expect(small.worldToScreen(0, 0)).toEqual({ x: 150, y: 100 });
});

it("keeps the cursor's world point through zoom and clamps native limits", () => {
  const c = new Camera(
    { width: 4000, height: 3000 },
    { width: 800, height: 600 },
  );
  c.centerOn(2000, 1500);
  const before = c.screenToWorld(300, 200);
  for (const factor of [1.5, 0.8, 1.2, 0.9]) {
    c.zoomAt(factor, 300, 200);
    const p = c.screenToWorld(300, 200);
    expect(p.x).toBeCloseTo(before.x, 9);
    expect(p.y).toBeCloseTo(before.y, 9);
  }
  c.zoomAt(100, 400, 300);
  expect(c.zoom).toBe(4);
  c.zoomAt(0.000001, 400, 300);
  expect(c.zoom).toBe(0.25);
});

it("preserves the centre and user zoom across DPR and viewport changes", () => {
  const c = new Camera(
    { width: 4000, height: 3000 },
    { width: 800, height: 600 },
  );
  c.centerOn(2000, 1500);
  c.zoomAt(1.5, 400, 300);
  c.setScale(2);
  expect(c.zoom).toBe(1.5);
  expect(c.effective()).toBe(3);
  expect(c.screenToWorld(400, 300)).toEqual({ x: 2000, y: 1500 });
  expect(c.worldToScreen(2010, 1500).x).toBeCloseTo(430, 9);
  expect(c.worldToScreen(2010, 1500).y).toBeCloseTo(300, 9);
  expect(c.visible().width).toBeCloseTo(800 / 3);
  expect(c.visible().height).toBe(200);
  c.resize({ width: 1200, height: 900 });
  expect(c.screenToWorld(600, 450)).toEqual({ x: 2000, y: 1500 });
  c.viewport(600, 450, 3);
  expect(c.view).toEqual({ width: 1800, height: 1350 });
  expect(c.scale).toBe(3);
  expect(c.zoom).toBe(1.5);
  expect(c.screenToWorld(900, 675)).toEqual({ x: 2000, y: 1500 });
});
