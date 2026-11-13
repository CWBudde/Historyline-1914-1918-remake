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
  /** Pixels per reference pixel of the 152×180 block. */
  scale: number;
  sides: [SideRects, SideRects];
  /** The brass rivets along the seam between the halves. */
  rivets: Point[];
}
/** The scene's block in reference pixels (dospanel battleW, battleH). */
export const battleWidth = 152,
  battleHeight = 180;
/**
 * battleLayout is native dospanel.BattleLayout: the 152×180 block centred in view at the largest scale up to
 * maximum, its edges snapped to the device pixels of dpr. The attacker's half has the photo at the left and the
 * hex at the right; the defender's below it is its mirror image.
 */
export function battleLayout(
  view: Rect,
  maximum = Infinity,
  dpr = 1,
): BattleLayout {
  const k = Math.max(
      0.001,
      Math.min(maximum, view.width / battleWidth, view.height / battleHeight),
    ),
    ox = view.x + (view.width - battleWidth * k) / 2,
    oy = view.y + (view.height - battleHeight * k) / 2,
    pixel = (n: number) => Math.round(n * dpr) / dpr;
  const side = (s: number): SideRects => {
    const rect = (x0: number, y0: number, x1: number, y1: number): Rect => {
      if (s === 1) [x0, x1] = [battleWidth - x1, battleWidth - x0];
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
  const rivets: Point[] = [];
  for (let x = 10; x < battleWidth; x += 22)
    rivets.push({ x: pixel(ox + x * k), y: pixel(oy + 90 * k) });
  return { scale: k, sides: [side(0), side(1)], rivets };
}
/** scaledLayout is l with every length times f: a layout in device pixels as CSS pixels. */
export function scaledLayout(l: BattleLayout, f: number): BattleLayout {
  const r = (v: Rect): Rect => ({
      x: v.x * f,
      y: v.y * f,
      width: v.width * f,
      height: v.height * f,
    }),
    side = (s: SideRects): SideRects => ({
      half: r(s.half),
      photoWell: r(s.photoWell),
      photo: r(s.photo),
      name: r(s.name),
      hex: r(s.hex),
      flag: r(s.flag),
      badge: r(s.badge),
      count: r(s.count),
    });
  return {
    scale: l.scale * f,
    sides: [side(l.sides[0]), side(l.sides[1])],
    rivets: l.rivets.map((p) => ({ x: p.x * f, y: p.y * f })),
  };
}
const ring: readonly Point[] = [
  { x: 0, y: -15 },
  { x: 18, y: -7.5 },
  { x: 18, y: 7.5 },
  { x: 0, y: 15 },
  { x: -18, y: 7.5 },
  { x: -18, y: -7.5 },
];
/** hexScale is the pixels per reference pixel of a side's hex (62 reference pixels wide). */
export function hexScale(side: SideRects): number {
  return side.hex.width / 62;
}
/** hexCentre is the centre of a side's hex. */
export function hexCentre(side: SideRects): Point {
  return {
    x: side.hex.x + side.hex.width / 2,
    y: side.hex.y + side.hex.height / 2,
  };
}
/** figurePoint is the centre of figure index on the hex's ring; on a ship a place along its hull. */
export function figurePoint(
  side: SideRects,
  index: number,
  ship: boolean,
): Point {
  const i = Math.max(0, Math.min(5, index)),
    c = hexCentre(side),
    k = hexScale(side);
  const p = ship
    ? { x: ((i % 3) - 1) * 10, y: Math.floor(i / 3) * 6 - 4 }
    : (ring[i] ?? { x: 0, y: 0 });
  return { x: c.x + p.x * k, y: c.y + p.y * k };
}
