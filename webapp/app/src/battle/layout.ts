import type { Point, Rect } from "../camera.ts";

export interface SideRects {
  half: Rect;
  photoWell: Rect;
  photo: Rect;
  name: Rect;
  hex: Rect;
  flag: Rect;
  badge: Rect;
  count: Rect;
}
export interface BattleLayout {
  scale: number;
  sides: [SideRects, SideRects];
}
/** Native dospanel.BattleLayout in CSS pixels, snapped to the current DPR. */
export function battleLayout(
  view: Rect,
  maximum = Infinity,
  dpr = 1,
): BattleLayout {
  const k = Math.max(
      0.001,
      Math.min(maximum, view.width / 152, view.height / 180),
    ),
    ox = view.x + (view.width - 152 * k) / 2,
    oy = view.y + (view.height - 180 * k) / 2,
    pixel = (n: number) => Math.round(n * dpr) / dpr;
  const side = (s: number): SideRects => {
    const rect = (x0: number, y0: number, x1: number, y1: number): Rect => {
      if (s === 1) [x0, x1] = [152 - x1, 152 - x0];
      const x = pixel(ox + x0 * k),
        y = pixel(oy + (y0 + s * 90) * k);
      return {
        x,
        y,
        width: pixel(ox + x1 * k) - x,
        height: pixel(oy + (y1 + s * 90) * k) - y,
      };
    };
    return {
      half: rect(2, 2, 150, 88),
      photoWell: rect(5, 5, 85, 73),
      photo: rect(7, 7, 83, 71),
      name: rect(5, 75, 85, 86),
      hex: rect(87, 16, 149, 74),
      flag: rect(10, 10, 25, 25),
      badge: rect(27, 9, 44, 26),
      count: rect(54, 50, 80, 68),
    };
  };
  return { scale: k, sides: [side(0), side(1)] };
}
const ring: readonly Point[] = [
  { x: 0, y: -15 },
  { x: 18, y: -7.5 },
  { x: 18, y: 7.5 },
  { x: 0, y: 15 },
  { x: -18, y: 7.5 },
  { x: -18, y: -7.5 },
];
export function figurePoint(
  side: SideRects,
  index: number,
  ship: boolean,
): Point {
  const i = Math.max(0, Math.min(5, index)),
    cx = side.hex.x + side.hex.width / 2,
    cy = side.hex.y + side.hex.height / 2,
    k = side.hex.width / 62;
  const p = ship
    ? { x: ((i % 3) - 1) * 10, y: Math.floor(i / 3) * 6 - 4 }
    : (ring[i] ?? { x: 0, y: 0 });
  return { x: cx + p.x * k, y: cy + p.y * k };
}
