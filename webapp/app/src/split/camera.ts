import { hexgrid } from "@historyline/engine";
import type { Camera } from "../camera.ts";
import type { Screen, View } from "./geometry.ts";

/** Native whole-pair bounds; a smaller map has one centred negative origin. */
export function pairRange(grid: hexgrid.Grid, cols: number): [number, number] {
  const span = 2 * grid.width - 1 - cols;
  if (span >= 0) return [0, Math.trunc((span + 1) / 2)];
  const centre = Math.floor(span / 4);
  return [centre, centre];
}

export function rowRange(grid: hexgrid.Grid, rows: number): [number, number] {
  const span = Math.trunc((grid.lines + 1) / 2) - rows;
  if (span >= 0) return [0, span];
  const centre = Math.floor(span / 2);
  return [centre, centre];
}

export function clampView(
  grid: hexgrid.Grid,
  cols: number,
  rows: number,
  pair: number,
  row: number,
): { pair: number; row: number } {
  const [plo, phi] = pairRange(grid, cols),
    [rlo, rhi] = rowRange(grid, rows);
  return {
    pair: Math.min(Math.max(pair, plo), phi),
    row: Math.min(Math.max(row, rlo), rhi),
  };
}

export function centreView(
  grid: hexgrid.Grid,
  cols: number,
  rows: number,
  at: hexgrid.Coord,
): { pair: number; row: number } {
  const column = 2 * at.col + (at.line & 1);
  return clampView(
    grid,
    cols,
    rows,
    Math.floor((column - Math.trunc((cols - 1) / 2)) / 2),
    (at.line >> 1) - Math.trunc((rows - 1) / 2),
  );
}

/** Map rendering starts at the panel's own origin, in physical screen pixels. */
export function alignCamera(camera: Camera, view: View, dpr: number): void {
  camera.bounded = false;
  const origin = view.origin(),
    offset = view.offset();
  camera.scale = dpr;
  camera.zoom = view.scale / dpr;
  camera.view = { width: view.panel.width, height: view.panel.height };
  camera.origin = {
    x: origin.x - (offset.x - view.panel.x) / view.scale,
    y: origin.y - (offset.y - view.panel.y) / view.scale,
  };
}

// Native math.Round rounds ties away from zero.
function round(value: number): number {
  return value < 0 ? -Math.round(-value) : Math.round(value);
}

/** Snap free camera panning back to the current native screen's cell lattice. */
export function cameraRest(
  camera: Camera,
  screen: Screen,
  side: 0 | 1,
  grid: hexgrid.Grid,
): { pair: number; row: number } {
  const view = screen.view(side, 0, 0),
    offset = view.offset();
  return clampView(
    grid,
    screen.cols,
    screen.rows,
    round(
      (camera.origin.x + (offset.x - view.panel.x) / view.scale) /
        hexgrid.colPitch,
    ),
    round(
      (camera.origin.y + (offset.y - view.panel.y) / view.scale) /
        hexgrid.hexHeight,
    ),
  );
}
