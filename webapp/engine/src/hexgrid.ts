// Package hexgrid (internal/hexgrid) implements the map layout of the Python reference: hexes are stored in
// lines, where even lines (0, 2, …) hold Width hexes and odd lines hold Width-1 hexes shifted half a hex to
// the right. Line l and line l±2 are stacked directly above each other. index.ts exports it as `hexgrid`.

import { div, rem } from "./int.ts";

/** Coord addresses a hex by line and column, the (x, y) pair used by hlrData.py. */
export interface Coord {
  readonly line: number;
  readonly col: number;
}

/** Grid is a map of Lines lines, Width hexes wide on even lines. */
export interface Grid {
  readonly width: number;
  readonly lines: number;
}

function coord(line: number, col: number): Coord {
  return { line, col };
}

/** contains reports whether c lies on the map. Odd lines are one hex shorter. */
export function contains(g: Grid, c: Coord): boolean {
  if (c.line < 0 || c.line >= g.lines || c.col < 0) {
    return false;
  }
  if (rem(c.line, 2) === 0) {
    return c.col < g.width;
  }
  return c.col < g.width - 1;
}

/** size returns the number of hexes on the map. */
export function size(g: Grid): number {
  const even = div(g.lines + 1, 2);
  return even * g.width + (g.lines - even) * (g.width - 1);
}

/** neighbors returns the on-map neighbours of c in the same order as the Python adjacentHexes. */
export function neighbors(g: Grid, c: Coord): Coord[] {
  return around(c).filter((n) => contains(g, n));
}

/**
 * around returns all six neighbours of c, on the map or not, in the order NE, SE, S, SW, NW, N (Python's
 * adjacentHexes order).
 */
export function around(c: Coord): Coord[] {
  const l = c.line;
  const col = c.col;
  if ((l & 1) === 1) {
    return [
      coord(l - 1, col + 1),
      coord(l + 1, col + 1),
      coord(l + 2, col),
      coord(l + 1, col),
      coord(l - 1, col),
      coord(l - 2, col),
    ];
  }
  return [
    coord(l - 1, col),
    coord(l + 1, col),
    coord(l + 2, col),
    coord(l + 1, col - 1),
    coord(l - 1, col - 1),
    coord(l - 2, col),
  ];
}

/**
 * cube converts c to cube coordinates [q, r, s] (q+r+s == 0). Doubling the column axis (2*Col, plus one on
 * odd lines) gives doubled coordinates whose neighbours are (±1, ±1) and (0, ±2), which map directly onto
 * cube axes.
 */
export function cube(c: Coord): [q: number, r: number, s: number] {
  const q = 2 * c.col + rem(c.line, 2);
  const r = div(c.line - q, 2);
  return [q, r, 0 - q - r];
}

/**
 * distance returns the number of steps between a and b. It replaces the breadth-first search in gui.py
 * calculateDistance.
 */
export function distance(a: Coord, b: Coord): number {
  const [aq, ar, as] = cube(a);
  const [bq, br, bs] = cube(b);
  return div(Math.abs(aq - bq) + Math.abs(ar - br) + Math.abs(as - bs), 2);
}

// Layout of the Python reference in map pixels, with (0, 0) at the top-left corner of the bounding box of hex
// (0, 0). Hexes are flat-topped, 96×80 pixels. A hex sits ColPitch pixels from its same-line neighbour and
// LinePitch pixels from the next line; odd lines are shifted by HalfPitch. This reproduces gui.py/main.py,
// where the map surface advances 142×40 pixels per hex column/line.
export const colPitch = 142;
export const halfPitch = colPitch / 2;
export const linePitch = 40;
export const hexWidth = 96;
export const hexHeight = 80;
export const centerX = hexWidth / 2;
export const centerY = hexHeight / 2;

/** center returns the pixel centre [x, y] of hex c. It is defined for any coordinate, on the map or not. */
export function center(c: Coord): [x: number, y: number] {
  return [
    c.col * colPitch + (c.line & 1) * halfPitch + centerX,
    c.line * linePitch + centerY,
  ];
}

/**
 * pixelToHex returns the hex under map pixel (x, y) and whether that hex is on the map. It replaces Python's
 * hard-coded findHex: the cell of a hex is the set of points closer to its centre than to any other centre,
 * which covers both line parities, the half-hex gaps at odd-line edges and points outside the map.
 */
export function pixelToHex(g: Grid, x: number, y: number): [Coord, boolean] {
  const line0 = Math.floor((y - centerY) / linePitch);
  let best = coord(0, 0);
  let bestD = Infinity;
  for (let line = line0 - 1; line <= line0 + 2; line++) {
    const shift = (line & 1) * halfPitch;
    const col0 = Math.floor((x - centerX - shift) / colPitch);
    for (let col = col0; col <= col0 + 1; col++) {
      const c = coord(line, col);
      const [cx, cy] = center(c);
      const d = (x - cx) * (x - cx) + (y - cy) * (y - cy);
      if (d < bestD) {
        best = c;
        bestD = d;
      }
    }
  }
  return [best, contains(g, best)];
}

/**
 * corners returns the six corners of hex c in map pixels, clockwise from the left vertex: left, top left, top
 * right, right, bottom right, bottom left.
 */
export function corners(c: Coord): [number, number][] {
  const [x, y] = center(c);
  const q = hexWidth - halfPitch; // run of a slanted edge, so that neighbours share edges exactly
  return [
    [x - centerX, y],
    [x - centerX + q, y - centerY],
    [x + centerX - q, y - centerY],
    [x + centerX, y],
    [x + centerX - q, y + centerY],
    [x - centerX + q, y + centerY],
  ];
}

/**
 * bounds returns the size [w, h] of the map in pixels: the bounding box of all its hexes, with (0, 0) at the
 * top-left corner of hex (0, 0). The widest line decides the width, so a one-line map is as wide as its Width
 * hexes.
 */
export function bounds(g: Grid): [w: number, h: number] {
  if (g.width <= 0 || g.lines <= 0) {
    return [0, 0];
  }
  let w = (g.width - 1) * colPitch + hexWidth;
  if (g.lines > 1 && g.width > 1) {
    w = Math.max(w, (g.width - 2) * colPitch + halfPitch + hexWidth);
  }
  return [w, (g.lines - 1) * linePitch + hexHeight];
}
