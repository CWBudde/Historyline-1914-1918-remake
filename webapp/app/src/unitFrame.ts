import { hexgrid, play } from "@historyline/engine";
import type { Point } from "./camera.ts";

export type Facing = "n" | "ne" | "se" | "s" | "sw" | "nw";
export interface SpriteFamily {
  family: string;
  long?: string;
}
export interface UnitFrame {
  key: string;
  id: string;
  family: string;
  unit: play.UnitView;
  at: Point;
  drawAt: hexgrid.Coord;
  facing: Facing;
}
const facings: readonly Facing[] = ["n", "ne", "se", "s", "sw", "nw"];
export function toward(from: Point, to: Point): Facing {
  if (from.x === to.x && from.y === to.y) return "n";
  const index = Math.round(
    Math.atan2(to.x - from.x, from.y - to.y) / (Math.PI / 3),
  );
  return facings[(index + 6) % 6] ?? "n";
}
export function squad(family: string, size: number): string {
  if (family === "infantry" && size >= 1 && size < 6)
    return `${family}-${size}`;
  if (family === "elite-infantry" && size >= 1 && size <= 3)
    return `${family}-1`;
  return family;
}
export function spriteID(
  family: string,
  owner: number,
  facing: Facing,
  spent: boolean,
): string {
  return `unit/${family}/${owner === 1 ? "cp" : owner === 2 ? "ec" : "neutral"}/${facing}-${spent ? "spent" : "normal"}`;
}

/** A frame receives a side's map, never the live board's enemy coordinates. */
export function unitFrames(
  view: play.View,
  sprites: Readonly<Record<string, SpriteFamily>>,
  directions: ReadonlyMap<number, string>,
  hidden: ReadonlySet<number> = new Set(),
): UnitFrame[] {
  const result: UnitFrame[] = [];
  for (const u of view.units.values()) {
    if (u.head !== null || hidden.has(u.index)) continue;
    const meta = sprites[u.id];
    if (meta === undefined) throw new Error(`missing sprite family ${u.id}`);
    let family = squad(meta.family, u.size),
      drawAt = u.at;
    const [x, y] = hexgrid.center(u.at);
    let at = { x, y };
    const label = directions.get(u.index) ?? "n";
    if (!facings.includes(label as Facing))
      throw new Error(`invalid authored facing ${label}`);
    let facing = label as Facing;
    if (u.rear !== null) {
      if (meta.long === undefined)
        throw new Error(`missing long ship family ${u.id}`);
      family = meta.long;
      const [rx, ry] = hexgrid.center(u.rear);
      at = { x: (x + rx) / 2, y: (y + ry) / 2 };
      facing = toward({ x: rx, y: ry }, { x, y });
      if (u.rear.line > u.at.line) drawAt = u.rear;
    }
    result.push({
      key: `unit/${u.index}`,
      id: spriteID(family, u.owner, facing, u.moved && u.side === view.side),
      family,
      unit: u,
      at,
      drawAt,
      facing,
    });
  }
  result.sort(
    (a, b) => a.drawAt.line - b.drawAt.line || a.drawAt.col - b.drawAt.col,
  );
  return result;
}
