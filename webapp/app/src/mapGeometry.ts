import { hexgrid } from "@historyline/engine";
import type { Rect, Size } from "./camera.ts";
import type { Mip, Pyramid, Tile } from "./assetManifest.ts";

export function intersects(a: Rect, b: Rect): boolean {
  return (
    a.x <= b.x + b.width &&
    a.x + a.width >= b.x &&
    a.y <= b.y + b.height &&
    a.y + a.height >= b.y
  );
}

export function visibleHexes(
  grid: hexgrid.Grid,
  rect: Rect,
  margin = 96,
): hexgrid.Coord[] {
  const area = {
    x: rect.x - margin,
    y: rect.y - margin,
    width: rect.width + 2 * margin,
    height: rect.height + 2 * margin,
  };
  const result: hexgrid.Coord[] = [];
  const first = Math.max(0, Math.floor((area.y - 80) / 40));
  const last = Math.min(grid.lines - 1, Math.ceil((area.y + area.height) / 40));
  for (let line = first; line <= last; line++) {
    const shift = (line % 2) * 71;
    const left = Math.max(0, Math.floor((area.x - shift - 96) / 142));
    const right = Math.min(
      grid.width - 1 - (line % 2),
      Math.ceil((area.x + area.width - shift) / 142),
    );
    for (let col = left; col <= right; col++) {
      const coord = { line, col };
      const [x, y] = hexgrid.center(coord);
      if (intersects({ x: x - 48, y: y - 40, width: 96, height: 80 }, area))
        result.push(coord);
    }
  }
  return result;
}

/** Backdrops can have a different source density from the terrain packs. */
export function chooseMip(p: Pyramid, world: Size, effective: number): number {
  const density = p.width / world.width;
  let chosen = 0;
  for (let i = 0; i < p.mips.length; i++) {
    const mip = p.mips[i];
    if (mip !== undefined && density / mip.scale >= effective) chosen = i;
  }
  return chosen;
}

export function backdropDraw(
  p: Pyramid,
  mip: Mip,
  tile: Tile,
  world: Size,
): { world: Rect; frame: Rect } {
  const x = tile.x * mip.scale,
    y = tile.y * mip.scale;
  const width = Math.min(tile.width * mip.scale, p.width - x),
    height = Math.min(tile.height * mip.scale, p.height - y);
  return {
    world: {
      x: (x / p.width) * world.width,
      y: (y / p.height) * world.height,
      width: (width / p.width) * world.width,
      height: (height / p.height) * world.height,
    },
    frame: {
      x: tile.x - tile.sourceX,
      y: tile.y - tile.sourceY,
      width: width / mip.scale,
      height: height / mip.scale,
    },
  };
}
