/** Pure internal/dosscreen geometry in physical screen pixels. Views scroll in
 * whole column pairs and rows so the steel frame follows the visible hexes. */
import { hexgrid } from "@historyline/engine";
import type { Point, Rect, Size } from "../camera.ts";

export const minCols = 7;
export const targetRows = 8;
export const minRows = 6;
export const minScale = 0.5;
export const defaultZoom = 7;
export const zooms = [
  0.5, 0.55, 0.61, 0.67, 0.74, 0.82, 0.91, 1, 1.1, 1.22, 1.35, 1.49, 1.64, 1.81,
  2,
] as const;

const eps = 1e-6;
// Go's math.Round rounds ties away from zero, including negative offsets.
function round(x: number): number {
  return x < 0 ? -Math.round(-x) : Math.round(x);
}
function scaled(n: number, ui: number): number {
  return Math.max(round(n * 2 * ui), 1);
}

export function blockSize(cols: number, rows: number): Size {
  return {
    width: (cols - 1) * hexgrid.halfPitch + hexgrid.hexWidth,
    height: rows * hexgrid.hexHeight + (cols <= 1 ? 0 : hexgrid.hexHeight / 2),
  };
}

export function fit(
  panel: Rect,
  scale: number,
): { cols: number; rows: number } {
  const w = panel.width / scale,
    h = panel.height / scale;
  const cols = Math.floor((w - hexgrid.hexWidth) / hexgrid.halfPitch + eps) + 1;
  const rows = Math.floor(
    (h - (cols <= 1 ? 0 : hexgrid.hexHeight / 2)) / hexgrid.hexHeight + eps,
  );
  return { cols: Math.max(cols, 1), rows: Math.max(rows, 1) };
}

/**
 * layout is the split screen in physical pixels. The original frame (steel false) follows internal/dosscreen: the
 * panels fill the window, a thin strip under each. The steel housing of the painted and tile designs (ui/armour.css)
 * leaves a margin of plate round two rectangular openings, a riveted bar between them and a taller strip under each.
 */
export function layout(
  width: number,
  height: number,
  ui = 1,
  steel = false,
): Screen {
  return new Screen(width, height, ui, steel);
}

export class Screen {
  width: number;
  height: number;
  ui: number;
  panels: [Rect, Rect];
  divider: Rect;
  bolt: Rect;
  status: [Rect, Rect];
  cellScale: number;
  cols: number;
  rows: number;
  zoom = defaultZoom;
  fitScale: number;
  /** The steel housing: rectangular openings in a plate (layout). */
  steel: boolean;

  constructor(width: number, height: number, ui = 1, steel = false) {
    this.width = width = Math.max(Math.trunc(width), 1);
    this.height = height = Math.max(Math.trunc(height), 1);
    this.ui = ui = ui <= 0 ? 1 : ui;
    this.steel = steel;
    if (steel) {
      // A margin of plate, the riveted bar between the openings, the strip under each and its foot.
      const margin = scaled(7, ui),
        bar = scaled(16, ui),
        gap = scaled(5, ui),
        status = scaled(28, ui),
        foot = scaled(4, ui),
        pw = Math.max(Math.trunc((width - 2 * margin - bar) / 2), 1),
        ph = Math.max(height - margin - gap - status - foot, 1),
        right = width - margin - pw,
        sy = margin + ph + gap;
      this.panels = [
        { x: margin, y: margin, width: pw, height: ph },
        { x: right, y: margin, width: pw, height: ph },
      ];
      this.divider = {
        x: margin + pw,
        y: margin,
        width: Math.max(right - margin - pw, 0),
        height: ph,
      };
      this.status = [
        { x: margin, y: sy, width: pw, height: status },
        { x: right, y: sy, width: pw, height: status },
      ];
      this.bolt = { x: this.divider.x, y: sy, width: 0, height: 0 };
    } else {
      const gap = scaled(4, ui),
        status = scaled(14, ui),
        foot = scaled(2, ui);
      const pw = Math.max(Math.trunc((width - scaled(8, ui)) / 2), 1),
        ph = Math.max(height - gap - status - foot, 1);
      this.panels = [
        { x: 0, y: 0, width: pw, height: ph },
        { x: width - pw, y: 0, width: pw, height: ph },
      ];
      this.divider = {
        x: pw,
        y: 0,
        width: Math.max(width - 2 * pw, 0),
        height: ph + gap,
      };
      const sy = ph + gap;
      this.status = [
        { x: 0, y: sy, width: pw, height: status },
        { x: width - pw, y: sy, width: pw, height: status },
      ];
      const bolt = Math.min(scaled(8, ui), this.divider.width, status);
      this.bolt = {
        x: pw + Math.trunc((this.divider.width - bolt) / 2),
        y: sy + Math.trunc((status - bolt) / 2),
        width: bolt,
        height: bolt,
      };
    }
    const ph = this.panels[0].height,
      pw = this.panels[0].width;
    const block = blockSize(minCols, targetRows);
    this.fitScale = this.cellScale = Math.max(
      Math.min(ph / block.height, pw / block.width),
      minScale * ui,
    );
    const cells = fit(this.panels[0], this.cellScale);
    this.cols = cells.cols;
    this.rows = cells.rows;
  }

  /** wellText is the status line's area in side's well: inside the bevel with a small inset, above the bar. */
  wellText(side: 0 | 1): Rect {
    const inner = this.wellInner(side),
      bar = this.wellBar(side),
      dx = scaled(2, this.ui),
      dy = scaled(1, this.ui);
    return {
      x: inner.x + dx,
      y: inner.y + dy,
      width: Math.max(inner.width - 2 * dx, 0),
      height: Math.max(bar.y - inner.y - dy, 0),
    };
  }

  /** wellBar is the track of the planning progress bar along the bottom of side's well. */
  wellBar(side: 0 | 1): Rect {
    const inner = this.wellInner(side),
      dx = scaled(2, this.ui),
      h = Math.min(scaled(2, this.ui), inner.height);
    return {
      x: inner.x + dx,
      y: inner.y + inner.height - h,
      width: Math.max(inner.width - 2 * dx, 0),
      height: h,
    };
  }

  private wellInner(side: 0 | 1): Rect {
    const st = this.status[side],
      bw = scaled(1, this.ui);
    return {
      x: st.x + bw,
      y: st.y + bw,
      width: Math.max(st.width - 2 * bw, 0),
      height: Math.max(st.height - 2 * bw, 0),
    };
  }

  view(side: 0 | 1, pair: number, row: number): View {
    return new View(this.panels[side], this.cellScale, pair, row);
  }

  zoomed(index: number): Screen {
    const s = new Screen(this.width, this.height, this.ui, this.steel);
    Object.assign(s, this);
    s.panels = [{ ...this.panels[0] }, { ...this.panels[1] }];
    s.status = [{ ...this.status[0] }, { ...this.status[1] }];
    s.divider = { ...this.divider };
    s.bolt = { ...this.bolt };
    s.zoom = this.clampZoom(index);
    s.cellScale = this.zoomScale(s.zoom);
    const cells = fit(s.panels[0], s.cellScale);
    s.cols = cells.cols;
    s.rows = cells.rows;
    return s;
  }

  zoomLevels(): number[] {
    const [lo, hi] = this.zoomRange();
    return Array.from({ length: hi - lo + 1 }, (_, i) => lo + i);
  }

  clampZoom(index: number): number {
    const [lo, hi] = this.zoomRange();
    return Math.min(Math.max(index, lo), hi);
  }

  slimOpening(side: number): Rect {
    const p = this.panels[(side & 1) as 0 | 1];
    // The steel housing's openings are the panels themselves.
    if (this.steel) return { ...p };
    const rim = scaled(3, this.ui),
      left = (side & 1) === 0 ? rim : 0;
    return {
      x: p.x + left,
      y: p.y + rim,
      width: Math.max(p.width - rim, 0),
      height: Math.max(p.height - rim, 0),
    };
  }

  slimFit(view: View, side: number, world: Rect): View {
    const v = new View(view.panel, view.scale, view.pair, view.row);
    v.cols = view.cols;
    v.rows = view.rows;
    const o = this.slimOpening(side),
      a = v.worldToScreen(world.x, world.y),
      b = v.worldToScreen(world.x + world.width, world.y + world.height);
    v.shiftX = -fitAxis(o.x, o.x + o.width, a.x, b.x) / v.scale;
    v.shiftY = -fitAxis(o.y, o.y + o.height, a.y, b.y) / v.scale;
    return v;
  }

  fillZoom(width: number, height: number): number {
    const [lo] = this.zoomRange(),
      o = this.slimOpening(0);
    for (let i = lo; i < defaultZoom; i++) {
      const scale = this.zoomScale(i);
      if (width * scale >= o.width - eps && height * scale >= o.height - eps)
        return i;
    }
    return defaultZoom;
  }

  private zoomRange(): [number, number] {
    let lo = defaultZoom,
      hi = defaultZoom;
    while (lo > 0 && this.zoomScale(lo - 1) >= minScale * this.ui - eps) lo--;
    while (hi < zooms.length - 1) {
      const scale = this.zoomScale(hi + 1),
        cells = fit(this.panels[0], scale);
      if (scale < minScale * this.ui - eps || cells.cols < 4 || cells.rows < 4)
        break;
      hi++;
    }
    return [lo, hi];
  }

  private zoomScale(index: number): number {
    return index === defaultZoom
      ? this.fitScale
      : this.fitScale * (zooms[index] ?? 1);
  }
}

export class View {
  cols: number;
  rows: number;
  shiftX = 0;
  shiftY = 0;

  constructor(
    public panel: Rect,
    public scale: number,
    public pair: number,
    public row: number,
  ) {
    this.panel = { ...panel };
    const cells = fit(panel, scale);
    this.cols = cells.cols;
    this.rows = cells.rows;
  }

  origin(): Point {
    return {
      x: this.pair * hexgrid.colPitch + this.shiftX,
      y: this.row * hexgrid.hexHeight + this.shiftY,
    };
  }

  offset(): Point {
    const b = blockSize(this.cols, this.rows);
    return {
      x: this.panel.x + round((this.panel.width - b.width * this.scale) / 2),
      y: this.panel.y + round((this.panel.height - b.height * this.scale) / 2),
    };
  }

  worldToScreen(x: number, y: number): Point {
    const o = this.origin(),
      b = this.offset();
    return { x: b.x + (x - o.x) * this.scale, y: b.y + (y - o.y) * this.scale };
  }

  screenToWorld(x: number, y: number): Point {
    const o = this.origin(),
      b = this.offset();
    return { x: (x - b.x) / this.scale + o.x, y: (y - b.y) / this.scale + o.y };
  }

  cell(j: number, i: number): hexgrid.Coord {
    return { line: 2 * (this.row + i) + (j & 1), col: this.pair + (j >> 1) };
  }

  cells(): hexgrid.Coord[] {
    const cells: hexgrid.Coord[] = [];
    for (let j = 0; j < this.cols; j++) {
      for (let i = 0; i < this.rows; i++) cells.push(this.cell(j, i));
    }
    return cells;
  }

  shows(grid: hexgrid.Grid, coord: hexgrid.Coord): boolean {
    const j = 2 * (coord.col - this.pair) + (coord.line & 1),
      i = (coord.line >> 1) - this.row;
    return (
      j >= 0 &&
      j < this.cols &&
      i >= 0 &&
      i < this.rows &&
      hexgrid.contains(grid, coord)
    );
  }
}

export interface Segment {
  a: Point;
  b: Point;
  /** Unit normal into the opening. */
  n: Point;
}

// Across successive hexgrid.corners edges, in hexgrid.around order.
const edgeNeighbour = [4, 5, 0, 1, 2, 3];

export function outline(view: View, grid: hexgrid.Grid): Segment[] {
  const out: Segment[] = [];
  for (const c of view.cells()) {
    if (!hexgrid.contains(grid, c)) continue;
    const [x, y] = hexgrid.center(c),
      centre = view.worldToScreen(x, y),
      corners = hexgrid.corners(c),
      around = hexgrid.around(c);
    for (const [k, neighbour] of edgeNeighbour.entries()) {
      const adjacent = around[neighbour],
        pa = corners[k],
        pb = corners[(k + 1) % 6];
      if (!adjacent || !pa || !pb || view.shows(grid, adjacent)) continue;
      const a = view.worldToScreen(pa[0], pa[1]),
        b = view.worldToScreen(pb[0], pb[1]);
      let nx = -(b.y - a.y),
        ny = b.x - a.x;
      const length = Math.hypot(nx, ny);
      nx /= length;
      ny /= length;
      if ((centre.x - a.x) * nx + (centre.y - a.y) * ny < 0) {
        nx = -nx;
        ny = -ny;
      }
      out.push({ a, b, n: { x: nx, y: ny } });
    }
  }
  return out;
}

export function rivets(view: View, grid: hexgrid.Grid): Point[] {
  const vertices = new Map<string, { p: Point; n: Point }>();
  const add = (p: Point, n: Point) => {
    const key = `${round(p.x * 16)},${round(p.y * 16)}`;
    let v = vertices.get(key);
    if (!v) {
      v = { p, n: { x: 0, y: 0 } };
      vertices.set(key, v);
    }
    v.n.x += n.x;
    v.n.y += n.y;
  };
  for (const { a, b, n } of outline(view, grid)) {
    add(a, n);
    add(b, n);
  }
  const scale = view.scale,
    off =
      (Math.max(4.5 * scale, 1.5) + Math.max(2 * scale, 1) + 1) /
      Math.sin((56 * Math.PI) / 180),
    out: Point[] = [];
  for (const { p, n } of vertices.values()) {
    const length = Math.hypot(n.x, n.y);
    if (length < 1e-9) continue;
    out.push({ x: p.x - (n.x / length) * off, y: p.y - (n.y / length) * off });
  }
  return out;
}

/** Whole physical pixels to translate a world span onto an opening. */
function fitAxis(o0: number, o1: number, w0: number, w1: number): number {
  if (w1 - w0 < o1 - o0) return round((o0 + o1 - w0 - w1) / 2);
  if (w0 > o0) return -Math.ceil(w0 - o0 - eps);
  if (w1 < o1) return Math.ceil(o1 - w1 - eps);
  return 0;
}
