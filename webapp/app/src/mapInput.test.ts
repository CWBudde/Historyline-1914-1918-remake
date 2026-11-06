import { expect, it } from "vitest";
import { Camera } from "./camera.ts";
import { MapGesture } from "./mapInput.ts";

function setup() {
  const camera = new Camera(
    { width: 4000, height: 3000 },
    { width: 800, height: 600 },
  );
  camera.setScale(2);
  camera.centerOn(2000, 1500);
  return { camera, input: new MapGesture(camera) };
}

it("separates taps from physical-pixel dragging at high DPR", () => {
  const { camera, input } = setup();
  const before = { ...camera.origin };
  input.down(1, 100, 100);
  expect(input.up(1, 102, 100)).toEqual({ x: 204, y: 200 });
  expect(camera.origin).toEqual(before);
  const anchor = camera.screenToWorld(200, 200);
  input.down(1, 100, 100);
  input.move(1, 130, 100);
  expect(input.up(1, 130, 100)).toBeNull();
  const after = camera.screenToWorld(260, 200);
  expect(after.x).toBeCloseTo(anchor.x, 9);
  expect(after.y).toBeCloseTo(anchor.y, 9);
});

it("pinches around the moving midpoint without later generating taps", () => {
  const { camera, input } = setup();
  const anchor = camera.screenToWorld(300, 200);
  input.down(1, 100, 100);
  input.down(2, 200, 100);
  input.move(1, 50, 100);
  input.move(2, 250, 100);
  expect(camera.zoom).toBeCloseTo(2, 9);
  const after = camera.screenToWorld(300, 200);
  expect(after.x).toBeCloseTo(anchor.x, 9);
  expect(after.y).toBeCloseTo(anchor.y, 9);
  expect(input.up(1, 50, 100)).toBeNull();
  expect(input.up(2, 250, 100)).toBeNull();
});

it("anchors wheel zoom and clears cancelled pointers", () => {
  const { camera, input } = setup();
  const anchor = camera.screenToWorld(300, 200);
  input.wheel(-100, 0, 150, 100);
  expect(camera.zoom).toBeCloseTo(1.1, 9);
  expect(camera.screenToWorld(300, 200).x).toBeCloseTo(anchor.x, 9);
  input.down(3, 100, 100);
  input.cancel();
  const before = { ...camera.origin };
  expect(input.move(3, 300, 100)).toBe(false);
  expect(input.up(3, 300, 100)).toBeNull();
  expect(camera.origin).toEqual(before);
});

it.each([0.75, 1, 2, 3])(
  "scrolls with the native 12 CSS pixels per tick at DPR %s",
  (dpr) => {
    const { camera, input } = setup();
    camera.setScale(dpr);
    const before = { ...camera.origin };
    input.keyboard(new Set(["d", "s"]), 1000 / 60);
    expect(camera.origin.x).toBeCloseTo(before.x + 12, 9);
    expect(camera.origin.y).toBeCloseTo(before.y + 12, 9);
  },
);

it("applies the native control-menu scroll speed to keyboard camera motion", () => {
  const c = new Camera(
      { width: 10000, height: 10000 },
      { width: 400, height: 300 },
    ),
    g = new MapGesture(c);
  g.keyboard(new Set(["d"]), 1000 / 60, 2);
  expect(c.origin.x).toBeCloseTo(24);
  g.keyboard(new Set(["d"]), 1000 / 60, 0.5);
  expect(c.origin.x).toBeCloseTo(30);
});
