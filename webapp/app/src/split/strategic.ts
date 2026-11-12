/**
 * The strategic map of internal/minimap (strategic.go, raster.go): the whole playable map on a dark steel plate in
 * a recessed frame, a block of the terrain's colour per cell, units as side-coloured dots (greyish once they have
 * their order), HQs as shields, depots as round D plates, factories as square F plates, and the panel's view as a
 * light rectangle; painted at whole pixels, magnified by the UI scale.
 */
import { hexgrid } from "@historyline/engine";
import type { Rect } from "../camera.ts";

type RGB = [number, number, number];
export interface Shade {
  lit: RGB;
  mid: RGB;
  dark: RGB;
}
export const blue: Shade = {
    lit: [85, 195, 243],
    mid: [56, 134, 174],
    dark: [32, 77, 109],
  },
  yellow: Shade = {
    lit: [231, 211, 101],
    mid: [166, 138, 52],
    dark: [101, 73, 16],
  },
  neutral: Shade = {
    lit: [196, 196, 184],
    mid: [140, 140, 128],
    dark: [72, 72, 62],
  };
/** sideShade is an owner's colours: the Central Powers blue, the Entente yellow. */
export function sideShade(owner: number): Shade {
  return owner === 1 ? blue : owner === 2 ? yellow : neutral;
}
/** spentShade is the greyish look of a unit with its order: four fifths grey, one fifth the side colour. */
export function spentShade(s: Shade): Shade {
  const grey = (c: RGB): RGB => {
    const y = Math.trunc((299 * c[0] + 587 * c[1] + 114 * c[2]) / 1000);
    return c.map((v) => Math.trunc((4 * y + v) / 5)) as RGB;
  };
  return { lit: grey(s.lit), mid: grey(s.mid), dark: grey(s.dark) };
}

export const plateColour: RGB = [48, 48, 36];
const frameOuter: RGB = [81, 81, 65],
  frameMiddle: RGB = [36, 36, 28],
  frameInner: RGB = [28, 28, 24],
  viewLine: RGB = [154, 154, 138],
  depotLit: RGB = [240, 236, 214],
  depotBody: RGB = [214, 210, 188],
  depotEdge: RGB = [92, 90, 76];
const cell = 4,
  frameWidth = 4,
  blockW = hexgrid.halfPitch,
  blockH = 2 * hexgrid.linePitch,
  inset = (hexgrid.hexWidth - hexgrid.halfPitch) / 2;

/** Pixels are RGBA pixels, as an ImageData holds them. */
export interface Pixels {
  width: number;
  height: number;
  data: Uint8ClampedArray<ArrayBuffer>;
}
export interface Building {
  at: hexgrid.Coord;
  kind: "hq" | "depot" | "factory";
  shade: Shade;
}
export interface Scene {
  terrain: (c: hexgrid.Coord) => RGB | null;
  units: { at: hexgrid.Coord; shade: Shade }[];
  buildings: Building[];
  /** The panel's view in world pixels. */
  view: Rect;
}

/** Strategic is the strategic map laid out in a picture of size w × h: the map's place and its cell size. */
export class Strategic {
  readonly cellW: number;
  readonly cellH: number;
  readonly map: Rect;
  readonly scale: number;
  constructor(
    readonly grid: hexgrid.Grid,
    readonly w: number,
    readonly h: number,
  ) {
    const base = this.size(cell, cell),
      k =
        base[0] > 0 && base[1] > 0
          ? Math.max(
              1,
              Math.min(
                Math.trunc((w - 2 * frameWidth) / base[0]),
                Math.trunc((h - 2 * frameWidth) / base[1]),
              ),
            )
          : 1;
    this.scale = k;
    this.cellW = k * cell;
    this.cellH = Math.max(k * cell + ((k * cell) & 1), 2);
    const [sw, sh] = this.size(this.cellW, this.cellH);
    this.map = {
      x: Math.trunc((w - sw) / 2),
      y: Math.trunc((h - sh) / 2),
      width: sw,
      height: sh,
    };
  }
  private size(cw: number, ch: number): [number, number] {
    const g = this.grid;
    if (g.width <= 0 || g.lines <= 0) return [0, 0];
    return [(2 * g.width - 1) * cw, Math.trunc(((g.lines + 1) * ch) / 2)];
  }
  /** block is cell c's block relative to the map. */
  block(c: hexgrid.Coord): Rect {
    return {
      x: (2 * c.col + (c.line & 1)) * this.cellW,
      y: Math.trunc((c.line * this.cellH) / 2),
      width: this.cellW,
      height: this.cellH,
    };
  }
  /** toWorld is the world point under picture pixel (px, py), null off the map. */
  toWorld(px: number, py: number): { x: number; y: number } | null {
    const m = this.map;
    if (px < m.x || py < m.y || px >= m.x + m.width || py >= m.y + m.height)
      return null;
    return {
      x: ((px - m.x) * blockW) / this.cellW + inset,
      y: ((py - m.y) * blockH) / this.cellH,
    };
  }
  /** render paints the scene into RGBA pixels of the picture's size. */
  render(sc: Scene): Pixels {
    const iw = Math.max(this.w, 1),
      ih = Math.max(this.h, 1),
      img: Pixels = {
        width: iw,
        height: ih,
        data: new Uint8ClampedArray(4 * iw * ih),
      },
      d = img.data,
      fill = (r: Rect, c: RGB) => {
        const x0 = Math.max(r.x, 0),
          y0 = Math.max(r.y, 0),
          x1 = Math.min(r.x + r.width, img.width),
          y1 = Math.min(r.y + r.height, img.height);
        for (let y = y0; y < y1; y++)
          for (let x = x0; x < x1; x++) {
            const i = 4 * (y * img.width + x);
            d[i] = c[0];
            d[i + 1] = c[1];
            d[i + 2] = c[2];
            d[i + 3] = 255;
          }
      },
      stroke = (r: Rect, c: RGB) => {
        fill({ x: r.x, y: r.y, width: r.width, height: 1 }, c);
        fill({ x: r.x, y: r.y + r.height - 1, width: r.width, height: 1 }, c);
        fill({ x: r.x, y: r.y, width: 1, height: r.height }, c);
        fill({ x: r.x + r.width - 1, y: r.y, width: 1, height: r.height }, c);
      },
      grow = (r: Rect, n: number): Rect => ({
        x: r.x - n,
        y: r.y - n,
        width: r.width + 2 * n,
        height: r.height + 2 * n,
      }),
      m = this.map;
    fill({ x: 0, y: 0, width: img.width, height: img.height }, plateColour);
    stroke(grow(m, 4), frameOuter);
    stroke(grow(m, 3), frameMiddle);
    stroke(grow(m, 2), frameMiddle);
    stroke(grow(m, 1), frameInner);
    fill(m, frameInner);
    const at = (r: Rect): Rect => ({ ...r, x: r.x + m.x, y: r.y + m.y });
    for (let line = 0; line < this.grid.lines; line++)
      for (let col = 0; col < this.grid.width; col++) {
        const c = { line, col };
        if (!hexgrid.contains(this.grid, c)) continue;
        const colour = sc.terrain(c);
        if (colour !== null) fill(at(this.block(c)), colour);
      }
    for (const u of sc.units) {
      const b = at(this.block(u.at)),
        n = Math.trunc(Math.min(this.cellW, this.cellH) / 4),
        r = {
          x: b.x + n,
          y: b.y + n,
          width: b.width - 2 * n,
          height: b.height - 2 * n,
        };
      for (let y = r.y; y < r.y + r.height; y++)
        for (let x = r.x; x < r.x + r.width; x++) {
          const lit = x === r.x || y === r.y,
            dark = x === r.x + r.width - 1 || y === r.y + r.height - 1;
          fill(
            { x, y, width: 1, height: 1 },
            lit && !dark
              ? u.shade.lit
              : dark && !lit
                ? u.shade.dark
                : u.shade.mid,
          );
        }
    }
    for (const b of sc.buildings) {
      const blk = at(this.block(b.at)),
        cx = blk.x + Math.trunc(blk.width / 2),
        cy = blk.y + Math.trunc(blk.height / 2),
        hq = b.kind === "hq",
        tones = hq ? shieldTones : b.kind === "depot" ? discTones : squareTones,
        pal: (RGB | null)[] = hq
          ? [null, b.shade.dark, b.shade.lit, b.shade.mid, b.shade.dark]
          : [
              null,
              b.shade === neutral ? depotEdge : b.shade.dark,
              depotLit,
              depotBody,
              frameMiddle,
            ],
        k = this.scale,
        h = tones.length,
        w = tones[0]?.length ?? 0,
        ox = cx - Math.trunc((w * k) / 2),
        oy = cy - Math.trunc((h * k) / 2);
      for (const [y, row] of tones.entries())
        for (const [x, tone] of row.entries()) {
          const c = pal[tone];
          if (c !== null && c !== undefined)
            fill({ x: ox + x * k, y: oy + y * k, width: k, height: k }, c);
        }
    }
    const v = this.viewRect(sc.view);
    if (v.width > 0 && v.height > 0) {
      fill(
        { x: v.x + v.width, y: v.y + 1, width: 1, height: v.height },
        frameOuter,
      );
      fill(
        { x: v.x + 1, y: v.y + v.height, width: v.width, height: 1 },
        frameOuter,
      );
      stroke(v, viewLine);
      fill(
        { x: v.x + 1, y: v.y + 1, width: 1, height: v.height - 2 },
        frameMiddle,
      );
      fill(
        { x: v.x + 1, y: v.y + 1, width: v.width - 2, height: 1 },
        frameMiddle,
      );
    }
    return img;
  }
  /** viewRect is the panel's view on the picture, cut to the map. */
  viewRect(view: Rect): Rect {
    const [w, h] = hexgrid.bounds(this.grid),
      from = (wx: number, wy: number) => [
        ((wx - inset) * this.cellW) / blockW,
        (wy * this.cellH) / blockH,
      ],
      [x0 = 0, y0 = 0] = from(Math.max(view.x, inset), Math.max(view.y, 0)),
      [x1 = 0, y1 = 0] = from(
        Math.min(view.x + view.width, w - inset),
        Math.min(view.y + view.height, h),
      ),
      m = this.map,
      ax = Math.max(Math.round(x0) + m.x, m.x),
      ay = Math.max(Math.round(y0) + m.y, m.y),
      bx = Math.min(Math.round(x1) + m.x, m.x + m.width),
      by = Math.min(Math.round(y1) + m.y, m.y + m.height);
    return {
      x: ax,
      y: ay,
      width: Math.max(bx - ax, 0),
      height: Math.max(by - ay, 0),
    };
  }
}

// The plates' glyphs and shapes: a shield with "HQ", a disc with "D", a square with "F".
const glyphH = ["#.#", "#.#", "###", "#.#", "#.#"],
  glyphQ = [".#.", "#.#", "#.#", "##.", ".##"],
  glyphD = ["##.", "#.#", "#.#", "#.#", "##."],
  glyphF = ["###", "#..", "##.", "#..", "#.."];
function plateTones(
  w: number,
  h: number,
  inside: (x: number, y: number) => boolean,
  marks: [number, number, string[]][],
): number[][] {
  const interior = (x: number, y: number) =>
    inside(x, y) &&
    inside(x - 1, y) &&
    inside(x + 1, y) &&
    inside(x, y - 1) &&
    inside(x, y + 1);
  const t: number[][] = Array.from({ length: h }, (_, y) =>
    Array.from({ length: w }, (_, x): number =>
      !inside(x, y)
        ? 0
        : !interior(x, y)
          ? 1
          : !interior(x - 1, y) || !interior(x, y - 1)
            ? 2
            : 3,
    ),
  );
  for (const [mx, my, g] of marks)
    for (const [gy, row] of g.entries())
      for (const [gx, ch] of Array.from(row).entries()) {
        const r = t[my + gy];
        if (ch === "#" && r !== undefined) r[mx + gx] = 4;
      }
  return t;
}
const shieldW = 11,
  shieldH = 12,
  disc = 9;
const shieldTones = plateTones(
    shieldW,
    shieldH,
    (x, y) => {
      if (x < 0 || y < 0 || y >= shieldH) return false;
      const d = Math.max(0, y - 6);
      return x >= d && x <= shieldW - 1 - d;
    },
    [
      [2, 2, glyphH],
      [6, 2, glyphQ],
    ],
  ),
  discTones = plateTones(
    disc,
    disc,
    (x, y) => {
      const dx = x - Math.trunc(disc / 2),
        dy = y - Math.trunc(disc / 2);
      return dx * dx + dy * dy <= 20;
    },
    [[3, 2, glyphD]],
  ),
  squareTones = plateTones(
    disc,
    disc,
    (x, y) => x >= 0 && y >= 0 && x < disc && y < disc,
    [[3, 2, glyphF]],
  );
