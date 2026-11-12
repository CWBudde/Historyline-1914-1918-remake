/**
 * The views that open inside a panel of the split screen, painted as internal/dospanel paints them: laid out in
 * the original's reference pixels (its 320×200 screen) at a scale in device pixels per reference pixel, centred in
 * the panel, the text in the pixel font. Drawing goes to a 2D context in the panel's device pixels.
 */
import { hexgrid } from "@historyline/engine";
import type { Rect } from "../camera.ts";
import type { PixelFace, PixelFonts } from "../pixfont.ts";
import type { Screen } from "./geometry.ts";
import {
  Painter,
  hexOf,
  toneBack,
  toneBlank,
  toneHex,
  tonePlate,
  type Surfaces,
} from "./painter.ts";

/** Text colours: cream on the plates, amber numbers, pale labels on the darkened map. */
export const cream = "#fbe3c7",
  amber = "#f0b858",
  pale = "#eeeee2";

/** refCell is a map cell's height in the reference's pixels. */
const refCell = 24;
/** panelScale is the plates' scale: the reference's proportion of a cell at the default zoom. */
export function panelScale(screen: Screen): number {
  return (screen.fitScale * hexgrid.hexHeight) / refCell;
}
/** plateRect is the part of side's panel a plate stays inside: the panel less half a cell all round. */
export function plateRect(screen: Screen, side: 0 | 1): Rect {
  const p = screen.panels[side],
    ix = Math.ceil((hexgrid.hexWidth / 2) * screen.cellScale),
    iy = Math.ceil((hexgrid.hexHeight / 2) * screen.cellScale);
  return {
    x: p.x + ix,
    y: p.y + iy,
    width: Math.max(p.width - 2 * ix, 0),
    height: Math.max(p.height - 2 * iy, 0),
  };
}
/** textScale is the pixel font's scale on a plate laid out at scale. */
export function textScale(scale: number): number {
  return Math.max(1, Math.floor(scale + 1e-9));
}

/** Grid places a block of reference pixels in a panel at scale k. */
export class Grid {
  constructor(
    readonly ox: number,
    readonly oy: number,
    readonly k: number,
  ) {}
  x(v: number): number {
    return Math.round(this.ox + v * this.k);
  }
  y(v: number): number {
    return Math.round(this.oy + v * this.k);
  }
  rect(x0: number, y0: number, x1: number, y1: number): Rect {
    const x = this.x(x0),
      y = this.y(y0);
    return { x, y, width: this.x(x1) - x, height: this.y(y1) - y };
  }
}
/** fit centres a w×h block in panel at scale, or at the largest scale that fits when the panel is too small. */
export function fit(panel: Rect, scale: number, w: number, h: number): Grid {
  const k = Math.max(Math.min(scale, panel.width / w, panel.height / h), 1e-3);
  return new Grid(
    panel.x + (panel.width - w * k) / 2,
    panel.y + (panel.height - h * k) / 2,
    k,
  );
}

/** Writer draws lines of the small face at scale k, each centred on its line at cy. */
export class Writer {
  constructor(
    readonly ctx: CanvasRenderingContext2D,
    readonly face: PixelFace,
    readonly k: number,
  ) {}
  left(s: string, x: number, cy: number, c: string): void {
    if (s === "") return;
    const h = this.face.measure(s, this.k).height;
    this.face.draw(this.ctx, s, x, cy - Math.trunc(h / 2), this.k, c);
  }
  right(s: string, x: number, cy: number, c: string): void {
    this.left(s, x - this.face.measure(s, this.k).width, cy, c);
  }
  centre(s: string, cx: number, cy: number, c: string): void {
    this.left(
      s,
      cx - Math.trunc(this.face.measure(s, this.k).width / 2),
      cy,
      c,
    );
  }
}

export interface Holdings {
  units: number;
  depots: number;
  factories: number;
}
/** GameInfo is what the game info shows (dospanel.GameInfo). */
export interface GameInfo {
  round: number;
  level: string;
  mode: string;
  best: number;
  score: number;
  sides: [Holdings, Holdings];
  map: Holdings;
  energy: [number, number];
  income: [number, number];
  moves: number;
  limit: number;
}
// The game info in reference pixels: 23 lines of 6 pixels, 130 wide.
const infoLines = 23,
  infoPitch = 6,
  infoW = 130,
  infoH = infoLines * infoPitch;
/** gameInfoScale is the scale the game info takes in a panel at scale (GameInfoLayout). */
export function gameInfoScale(panel: Rect, scale: number): number {
  return fit(panel, scale, infoW, infoH).k;
}
/** drawGameInfo draws the game info over panel at scale: the map darkens and the text stands on it. */
export function drawGameInfo(
  ctx: CanvasRenderingContext2D,
  panel: Rect,
  scale: number,
  gi: GameInfo,
  fonts: PixelFonts,
): void {
  ctx.fillStyle = "rgba(0,0,0,0.6)";
  ctx.fillRect(panel.x, panel.y, panel.width, panel.height);
  const g = fit(panel, scale, infoW, infoH),
    w = new Writer(ctx, fonts.small, textScale(scale)),
    line = (n: number) => g.y(infoPitch / 2 + infoPitch * n),
    num = String;
  for (const [i, [label, value]] of [
    ["ROUND", num(gi.round)],
    ["LEVEL", gi.level],
    ["MODE", gi.mode],
    ["BEST", num(gi.best)],
    ["CURRENT", num(gi.score)],
  ].entries()) {
    w.left(label ?? "", g.x(19), line(i), pale);
    w.left(value ?? "", g.x(73), line(i), pale);
  }
  const cols = [91, 109, 127]; // right edges of the columns
  for (const [i, h] of ["ONE", "TWO", "MAP"].entries())
    w.centre(h, g.x((cols[i] ?? 0) - 6), line(9), pale);
  for (const r of [
    {
      n: 11,
      label: "UNITS",
      v: [gi.sides[0].units, gi.sides[1].units, gi.map.units],
    },
    {
      n: 13,
      label: "DEPOTS",
      v: [gi.sides[0].depots, gi.sides[1].depots, gi.map.depots],
    },
    {
      n: 14,
      label: "FACTORY UNITS",
      v: [gi.sides[0].factories, gi.sides[1].factories, gi.map.factories],
    },
    { n: 16, label: "RESOURCES TOTAL", v: gi.energy },
    { n: 17, label: "RESOURCES INCOME", v: gi.income },
  ]) {
    w.left(r.label, g.x(2), line(r.n), pale);
    for (const [i, v] of r.v.entries())
      w.right(num(v), g.x(cols[i] ?? 0), line(r.n), amber);
  }
  w.left("MOVES", g.x(2), line(20), pale);
  w.left(num(gi.moves), g.x(50), line(20), pale);
  w.left("LIMIT", g.x(2), line(22), pale);
  w.left(gi.limit > 0 ? num(gi.limit) : "NO LIMIT", g.x(50), line(22), pale);
}

export interface Weapon {
  power: number;
  range: number;
}
/** UnitInfo is what the unit info shows of a type (dospanel.UnitInfo). */
export interface UnitInfo {
  name: string;
  air: Weapon;
  ground: Weapon;
  water: Weapon;
  armour: number;
  speed: number;
  weight: number;
  group: number;
}
// The unit info in reference pixels: a 140×162 block, a 100×84 photo at the top left, the hex plate at its right.
const unitW = 140,
  unitH = 162,
  hexX = 124,
  hexY = 78,
  hexA = 13,
  hexB = 12;
/** unitInfoLayout lays out the unit info in panel at scale (UnitInfoLayout). */
export function unitInfoLayout(panel: Rect, scale: number) {
  const g = fit(panel, scale, unitW, unitH);
  return {
    scale: g.k,
    plate: g.rect(0, 0, unitW, unitH),
    photoWell: g.rect(4, 4, 110, 94),
    photo: g.rect(7, 7, 107, 91),
    hex: g.rect(hexX - hexA, hexY - hexB, hexX + hexA, hexY + hexB),
    name: g.rect(2, 97, 138, 108),
    stats: g.rect(2, 110, 138, 160),
  };
}

const plates = new Map<string, ImageData>();
/** cachedPlate is a plate painted once per kind, panel size, scale and surfaces. */
export function cachedPlate(
  key: string,
  width: number,
  height: number,
  paint: (p: Painter) => void,
  surfaces: Surfaces,
  scale: number,
): ImageData {
  const k = `${key}/${width}x${height}/${scale}/${surfaces.steel !== null}/${surfaces.rivets !== null}`;
  let img = plates.get(k);
  if (img === undefined) {
    if (plates.size >= 8) plates.clear();
    img = new ImageData(Math.max(width, 1), Math.max(height, 1));
    paint(new Painter(img.data, img.width, img.height, scale, surfaces));
    plates.set(k, img);
  }
  return img;
}

/** drawPhoto draws img over r, scaled to cover it, in sepia. */
export function drawPhoto(
  ctx: CanvasRenderingContext2D,
  r: Rect,
  img: CanvasImageSource & { width: number; height: number },
): void {
  if (img.width === 0 || r.width <= 0) return;
  const s = Math.max(r.width / img.width, r.height / img.height);
  ctx.save();
  ctx.beginPath();
  ctx.rect(r.x, r.y, r.width, r.height);
  ctx.clip();
  ctx.filter = "sepia(1)";
  ctx.imageSmoothingEnabled = true;
  ctx.drawImage(
    img,
    r.x + (r.width - img.width * s) / 2,
    r.y + (r.height - img.height * s) / 2,
    img.width * s,
    img.height * s,
  );
  ctx.restore();
}
/** drawSprite draws img centred in r, scaled to fit share of it. */
export function drawSprite(
  ctx: CanvasRenderingContext2D,
  r: Rect,
  img: CanvasImageSource & { width: number; height: number },
  share: number,
): void {
  if (img.width === 0 || r.width <= 0) return;
  const s = share * Math.min(r.width / img.width, r.height / img.height);
  ctx.imageSmoothingEnabled = true;
  ctx.drawImage(
    img,
    r.x + (r.width - img.width * s) / 2,
    r.y + (r.height - img.height * s) / 2,
    img.width * s,
    img.height * s,
  );
}

type Picture = CanvasImageSource & { width: number; height: number };
/** drawUnitInfo draws the unit info of u over panel: a steel plate with the sepia photo, the sprite on its hex plate, the name plate and the stats. */
export function drawUnitInfo(
  ctx: CanvasRenderingContext2D,
  panel: Rect,
  scale: number,
  u: UnitInfo,
  fonts: PixelFonts,
  surfaces: Surfaces,
  portrait: Picture | null,
  sprite: Picture | null,
): void {
  const l = unitInfoLayout(panel, scale);
  ctx.putImageData(
    cachedPlate(
      "unit",
      panel.width,
      panel.height,
      (p) => {
        const local = unitInfoLayout(
          { x: 0, y: 0, width: panel.width, height: panel.height },
          scale,
        );
        p.steel({ x: 0, y: 0, width: p.w, height: p.h }, toneBack);
        p.plate(local.plate, tonePlate);
        p.well(local.photoWell);
        p.plate(local.photo, toneBlank);
        const [cx, cy, a, b] = hexOf(local.hex);
        p.hex(cx, cy, a, b, toneHex, true);
        p.namePlate(local.name);
        p.namePlate(local.stats);
      },
      surfaces,
      scale,
    ),
    panel.x,
    panel.y,
  );
  if (portrait !== null) drawPhoto(ctx, l.photo, portrait);
  if (sprite !== null) drawSprite(ctx, l.hex, sprite, 0.8);
  const g = fit(panel, scale, unitW, unitH),
    w = new Writer(ctx, fonts.small, textScale(scale));
  w.centre(u.name.toUpperCase(), g.x(unitW / 2), g.y(102.5), cream);
  w.left("WEAPONS", g.x(8), g.y(117), cream);
  w.left("(POWER/RANGE)", g.x(8), g.y(123), cream);
  for (const [i, [label, wp]] of (
    [
      ["AIR", u.air],
      ["GROUND", u.ground],
      ["WATER", u.water],
    ] as const
  ).entries()) {
    const y = g.y(116 + 6 * i);
    w.left(label, g.x(70), y, cream);
    w.right(String(wp.power), g.x(113), y, amber);
    w.centre("/", g.x(117), y, cream);
    w.left(String(wp.range), g.x(121), y, amber);
  }
  for (const [i, [label, v]] of (
    [
      ["ARMOUR", u.armour],
      ["SPEED", u.speed],
      ["WEIGHT", u.weight],
      ["MAX GROUP SIZE", u.group],
    ] as const
  ).entries()) {
    const y = g.y(137 + 6 * i);
    w.left(label, g.x(8), y, cream);
    w.right(String(v), g.x(104), y, amber);
  }
}
