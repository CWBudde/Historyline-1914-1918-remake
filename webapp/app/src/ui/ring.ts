import { hexgrid } from "@historyline/engine";
import type { Point, Rect } from "../camera.ts";

export type Action =
  | "cancel"
  | "move"
  | "attack"
  | "mode"
  | "open"
  | "map"
  | "info"
  | "exit"
  | "produce"
  | "repair";
export type Direction = "centre" | "up" | "left" | "right" | "down";
export type Placement = Record<Direction, Action>;
export function placement(
  attack: boolean,
  ownContainer: boolean,
  contents = false,
): Placement {
  return contents
    ? {
        centre: "cancel",
        up: "exit",
        left: "produce",
        right: "cancel",
        down: "repair",
      }
    : {
        centre: "cancel",
        up: attack ? "attack" : "move",
        left: ownContainer ? "open" : "mode",
        right: "map",
        down: "info",
      };
}
export function ringActions(
  p: Placement,
  allowed: (a: Action) => boolean,
): Action[] {
  return [p.up, p.right, p.down, p.left].filter(
    (a) => a !== "cancel" && allowed(a),
  );
}
export interface RingItem extends Point {
  action: Action;
  points: Point[];
}
export interface RingLayout extends Point {
  scale: number;
  radius: number;
  items: RingItem[];
}
export function layoutRing(
  actions: readonly Action[],
  anchor: Point,
  viewport: Rect,
  requestedScale = 1,
): RingLayout {
  if (
    ![viewport.width, viewport.height, requestedScale].every(
      (n) => Number.isFinite(n) && n > 0,
    )
  )
    throw new Error("invalid ring viewport or scale");
  const distance = 104,
    radius =
      actions.length > 2
        ? Math.max(distance, distance / 2 / Math.sin(Math.PI / actions.length))
        : distance;
  const positions = actions.map((action, i) => {
    const angle = (2 * Math.PI * i) / actions.length;
    // Cardinal positions stay exact, like the native reference's intended geometry.
    const clean = (n: number) => (Math.abs(n) < 1e-10 ? 0 : n);
    return {
      action,
      x: clean(radius * Math.sin(angle)),
      y: clean(-radius * Math.cos(angle)),
    };
  });
  const minX = Math.min(0, ...positions.map((p) => p.x)) - 48,
    maxX = Math.max(0, ...positions.map((p) => p.x)) + 48,
    minY = Math.min(0, ...positions.map((p) => p.y)) - 40,
    maxY = Math.max(0, ...positions.map((p) => p.y)) + 40,
    pad = Math.min(8, viewport.width / 4, viewport.height / 4),
    scale = Math.min(
      requestedScale,
      (viewport.width - 2 * pad) / (maxX - minX),
      (viewport.height - 2 * pad) / (maxY - minY),
    ),
    x = Math.max(
      viewport.x + pad - minX * scale,
      Math.min(anchor.x, viewport.x + viewport.width - pad - maxX * scale),
    ),
    y = Math.max(
      viewport.y + pad - minY * scale,
      Math.min(anchor.y, viewport.y + viewport.height - pad - maxY * scale),
    );
  return {
    x,
    y,
    scale,
    radius: radius * scale,
    items: positions.map((p) => {
      const px = x + p.x * scale,
        py = y + p.y * scale;
      return {
        action: p.action,
        x: px,
        y: py,
        points: hexgrid.corners({ line: 0, col: 0 }).map(([cx, cy]) => ({
          x: px + (cx - 48) * scale,
          y: py + (cy - 40) * scale,
        })),
      };
    }),
  };
}
function inPlate(p: Point, centre: Point, scale: number): boolean {
  const dx = Math.abs(p.x - centre.x) / scale,
    dy = Math.abs(p.y - centre.y) / scale;
  return dy <= 40 && dx <= 48 - (24 * dy) / 40;
}
export function hitRing(ring: RingLayout, p: Point): Action | null {
  for (const item of ring.items)
    if (inPlate(p, item, ring.scale)) return item.action;
  return inPlate(p, ring, ring.scale) ? "cancel" : null;
}
