/**
 * The cursor of the desktop game (internal/gesture draw.go, layers.go, glyph.go, ring.go): a steel hex ring round
 * the cell, the solid action plate with an embossed glyph while fire is held or an order is aimed, and the ring
 * menu of plates. Pure geometry: each look is a list of layers, polygons filled in one colour in order, with holes
 * cut from the part that holds them.
 */
import { hexgrid } from "@historyline/engine";
import type { Action } from "./ui/ring.ts";

export type Pt = [number, number];
export interface Polygon {
  pts: Pt[];
  hole: boolean;
}
export interface Layer {
  colour: number;
  polys: Polygon[];
}
/** How a plate is drawn: raised, pushed in (dark top rim, glyph sunk), or not possible here (faded glyph). */
export type PlateState = "normal" | "pressed" | "disabled";

/** One pixel of the original's 24-px cell in world px: band, rims and glyph follow its proportions. */
const unit = hexgrid.hexHeight / 24;
// The placeholder steel, olive greys lit from the top left.
const outline = 0x181c00,
  steel = 0x515141,
  steelDark = 0x424234,
  speck = 0x9a9a8a,
  light = 0xc7c7be,
  face = 0x828271,
  shadow = 0x2c241c,
  shadowMid = 0x453828,
  faded = 0x646454;
// The ring: the original's steel band 3 wide between dark lines half a pixel wide, shaded in strips.
const ringLine = 0.5,
  ringBand = 3,
  ringStrips = 10;
const lightX = -Math.SQRT2 / 2,
  lightY = -Math.SQRT2 / 2;

/** plateCorners are a plate's corners centred on (cx, cy) at scale, in the order of hexgrid.corners. */
export function plateCorners(cx: number, cy: number, scale: number): Pt[] {
  const [x0, y0] = hexgrid.center({ line: 0, col: 0 });
  return hexgrid
    .corners({ line: 0, col: 0 })
    .map(([x, y]) => [cx + (x - x0) * scale, cy + (y - y0) * scale]);
}

/** inPlate reports whether (x, y) lies on the plate centred on (cx, cy). */
export function inPlate(
  x: number,
  y: number,
  cx: number,
  cy: number,
  scale: number,
): boolean {
  const dx = Math.abs(x - cx) / scale,
    dy = Math.abs(y - cy) / scale,
    hw = hexgrid.hexWidth / 2,
    hh = hexgrid.hexHeight / 2,
    slant = hexgrid.hexWidth - hexgrid.halfPitch;
  return dy <= hh && dx <= hw - (slant * dy) / hh;
}

/** inset moves every edge of the convex polygon p (positive shoelace area) inwards by d. */
export function inset(p: readonly Pt[], d: number): Pt[] {
  const n = p.length,
    lines = p.map((a, i) => {
      const b = p[(i + 1) % n] ?? a,
        ex = b[0] - a[0],
        ey = b[1] - a[1],
        l = Math.hypot(ex, ey);
      return { px: a[0] - (ey / l) * d, py: a[1] + (ex / l) * d, ex, ey };
    });
  return p.map((_, i) => {
    const l1 = lines[(i + n - 1) % n],
      l2 = lines[i];
    if (l1 === undefined || l2 === undefined) return [0, 0];
    const t =
      ((l2.px - l1.px) * l2.ey - (l2.py - l1.py) * l2.ex) /
      (l1.ex * l2.ey - l1.ey * l2.ex);
    return [l1.px + t * l1.ex, l1.py + t * l1.ey];
  });
}

/** edgeStrip is the band between insets d0 and d1 of hex along its corners from..to, wrapping past the last. */
function edgeStrip(
  hex: readonly Pt[],
  d0: number,
  d1: number,
  from: number,
  to: number,
): Pt[] {
  const o = inset(hex, d0),
    i = inset(hex, d1),
    out: Pt[] = [];
  for (let k = from; k <= to; k++) out.push(o[k % o.length] ?? [0, 0]);
  for (let k = to; k >= from; k--) out.push(i[k % i.length] ?? [0, 0]);
  return out;
}

const solid = (pts: Pt[]): Polygon => ({ pts, hole: false });
function mix(a: number, b: number, f: number): number {
  const ch = (c: number, s: number) => (c >> s) & 0xff,
    m = (s: number) => Math.round(ch(a, s) + f * (ch(b, s) - ch(a, s))) << s;
  return m(16) | m(8) | m(0);
}
/** rimShade is the steel of a slope facing the light by f (-1..1). */
function rimShade(f: number): number {
  const to = f < 0 ? shadow : light;
  return mix(steel, to, Math.min(Math.abs(f), 1) * 0.8);
}

/** ringLayers is the cursor: a rounded steel band between dark lines lit from the top left, open inside. */
export function ringLayers(cx: number, cy: number, scale: number): Layer[] {
  const hex = plateCorners(cx, cy, scale),
    u = unit * scale,
    layers: Layer[] = [
      {
        colour: outline,
        polys: [
          solid(hex),
          { pts: inset(hex, (2 * ringLine + ringBand) * u), hole: true },
        ],
      },
    ],
    w = ringBand / ringStrips;
  for (let i = 0; i < hex.length; i++) {
    const a = hex[i] ?? [0, 0],
      b = hex[(i + 1) % hex.length] ?? [0, 0],
      l = Math.hypot(b[0] - a[0], b[1] - a[1]),
      lit = ((b[1] - a[1]) * lightX - (b[0] - a[0]) * lightY) / l;
    for (let k = 0; k < ringStrips; k++) {
      const slope = 1 - (2 * k + 1) / ringStrips,
        d = ringLine + k * w;
      layers.push({
        colour: rimShade(lit * slope),
        polys: [solid(edgeStrip(hex, d * u, (d + w) * u, i, i + 1))],
      });
    }
  }
  return layers;
}

/** plateLayers is an action plate: the cell filled with raised steel and the glyph of a, 14 units across. */
export function plateLayers(
  cx: number,
  cy: number,
  scale: number,
  a: Action,
  s: PlateState,
): Layer[] {
  const hex = plateCorners(cx, cy, scale),
    u = unit * scale,
    layers: Layer[] = [
      { colour: outline, polys: [solid(hex)] },
      {
        colour: s === "pressed" ? steelDark : steel,
        polys: [solid(inset(hex, u))],
      },
    ];
  if (s === "normal")
    layers.push(
      { colour: speck, polys: [solid(edgeStrip(hex, u, 2 * u, 0, 2))] },
      { colour: light, polys: [solid(edgeStrip(hex, 2 * u, 3 * u, 4, 5))] },
    );
  else if (s === "pressed")
    layers.push({
      colour: shadow,
      polys: [solid(edgeStrip(hex, u, 2 * u, 0, 2))],
    });
  const g = (dx: number, dy: number) => glyphAt(a, cx + dx, cy + dy, 14 * u);
  if (s === "disabled")
    return [
      ...layers,
      { colour: shadowMid, polys: g(0.5 * u, 0.5 * u) },
      { colour: faded, polys: g(0, 0) },
    ];
  if (s === "pressed")
    return [
      ...layers,
      { colour: shadow, polys: g(1.5 * u, 1.5 * u) },
      { colour: face, polys: g(0.5 * u, 0.5 * u) },
    ];
  return [
    ...layers,
    { colour: light, polys: g(-u, -u) },
    { colour: shadow, polys: g(u, u) },
    { colour: face, polys: g(0, 0) },
  ];
}

/** Ring is the ring menu in world px: a plate per action round the cancel plate at (x, y), clockwise from the top. */
export interface Ring {
  x: number;
  y: number;
  radius: number;
  actions: Action[];
}
/** The clearance between ring plates in world px. */
const ringGap = 8;
/** newRing places actions round (x, y) as near as they fit without plates meeting. */
export function newRing(x: number, y: number, actions: Action[]): Ring {
  const d = hexgrid.hexWidth + ringGap,
    n = actions.length;
  return {
    x,
    y,
    radius: n > 2 ? Math.max(d, d / 2 / Math.sin(Math.PI / n)) : d,
    actions,
  };
}
/** ringPos is the centre of the ring's item i. */
export function ringPos(r: Ring, i: number): Pt {
  const a = (2 * Math.PI * i) / r.actions.length;
  return [r.x + r.radius * Math.sin(a), r.y - r.radius * Math.cos(a)];
}
/** hitRing is the item whose plate holds (x, y), "cancel" on the centre plate, or null. */
export function hitRing(
  r: Ring,
  x: number,
  y: number,
): number | "cancel" | null {
  for (let i = 0; i < r.actions.length; i++) {
    const [px, py] = ringPos(r, i);
    if (inPlate(x, y, px, py, 1)) return i;
  }
  return inPlate(x, y, r.x, r.y, 1) ? "cancel" : null;
}
/** fitRing moves r as little as it takes to keep all its plates inside the box, where they fit. */
export function fitRing(
  r: Ring,
  box: { x: number; y: number; width: number; height: number },
): Ring {
  const hw = hexgrid.hexWidth / 2,
    hh = hexgrid.hexHeight / 2;
  let x0 = r.x - hw,
    y0 = r.y - hh,
    x1 = r.x + hw,
    y1 = r.y + hh;
  for (let i = 0; i < r.actions.length; i++) {
    const [x, y] = ringPos(r, i);
    x0 = Math.min(x0, x - hw);
    y0 = Math.min(y0, y - hh);
    x1 = Math.max(x1, x + hw);
    y1 = Math.max(y1, y + hh);
  }
  const into = (lo: number, hi: number, a: number, b: number) =>
    lo < a ? a - lo : hi > b ? Math.max(b - hi, a - lo) : 0;
  return {
    ...r,
    x: r.x + into(x0, x1, box.x, box.x + box.width),
    y: r.y + into(y0, y1, box.y, box.y + box.height),
  };
}
/** ringMenuLayers are the ring's plates: cancel on the cursor, every action, the hovered one pressed. */
export function ringMenuLayers(r: Ring, hover: number | null): Layer[] {
  const out = plateLayers(r.x, r.y, 1, "cancel", "normal");
  for (const [i, a] of r.actions.entries()) {
    const [x, y] = ringPos(r, i);
    out.push(...plateLayers(x, y, 1, a, i === hover ? "pressed" : "normal"));
  }
  return out;
}

/** cursorLook is what the cursor shows (cursorPlate): a plate of an action in a state, or the plain ring (null). */
export function cursorLook(
  pick: Action,
  onTarget: boolean,
  held: boolean,
  centre: boolean,
  picked: Action,
  allowed: boolean,
): { action: Action; state: PlateState } | null {
  if (pick !== "cancel" && held && centre)
    return { action: pick, state: onTarget ? "pressed" : "disabled" };
  if (pick !== "cancel" && !held)
    return onTarget ? { action: pick, state: "normal" } : null;
  if (!held) return null;
  return { action: picked, state: allowed ? "normal" : "disabled" };
}

// Glyphs: each symbol in its unit box, x and y in [-0.5, 0.5], y down, newly drawn from bars, triangles and discs.
function rect(x0: number, y0: number, x1: number, y1: number): Pt[] {
  return [
    [x0, y0],
    [x1, y0],
    [x1, y1],
    [x0, y1],
  ];
}
function tri(...p: [number, number, number, number, number, number]): Pt[] {
  return [
    [p[0], p[1]],
    [p[2], p[3]],
    [p[4], p[5]],
  ];
}
function bar(x0: number, y0: number, x1: number, y1: number, w: number): Pt[] {
  const dx = x1 - x0,
    dy = y1 - y0,
    l = Math.hypot(dx, dy),
    nx = ((-dy / l) * w) / 2,
    ny = ((dx / l) * w) / 2;
  return [
    [x0 + nx, y0 + ny],
    [x1 + nx, y1 + ny],
    [x1 - nx, y1 - ny],
    [x0 - nx, y0 - ny],
  ];
}
function disc(cx: number, cy: number, r: number, n: number): Pt[] {
  return Array.from({ length: n }, (_, i): Pt => {
    const a = (2 * Math.PI * i) / n;
    return [cx + r * Math.cos(a), cy + r * Math.sin(a)];
  });
}
const hole = (pts: Pt[]): Polygon => ({ pts, hole: true });
function question(): Polygon[] {
  const w = 0.13,
    r = 0.2,
    cy = -0.16,
    pts: Pt[] = [];
  for (let deg = 190; deg <= 420; deg += 23) {
    const a = (deg * Math.PI) / 180;
    pts.push([r * Math.cos(a), cy + r * Math.sin(a)]);
  }
  pts.push([0, 0.08], [0, 0.17]);
  const parts: Pt[][] = [];
  for (let i = 1; i < pts.length; i++) {
    const a = pts[i - 1] ?? [0, 0],
      b = pts[i] ?? [0, 0];
    parts.push(bar(a[0], a[1], b[0], b[1], w));
    if (i < pts.length - 1) parts.push(disc(b[0], b[1], w / 2, 8));
  }
  parts.push(disc(0, 0.36, 0.08, 12));
  return parts.map(solid);
}
/** glyph is the embossed symbol of a. */
export function glyph(a: Action): Polygon[] {
  switch (a) {
    case "cancel":
      return [
        bar(-0.36, -0.36, 0.36, 0.36, 0.16),
        bar(-0.36, 0.36, 0.36, -0.36, 0.16),
      ].map(solid);
    case "move":
      return [
        bar(-0.3, 0, 0.3, 0, 0.12),
        bar(0, -0.3, 0, 0.3, 0.12),
        tri(0, -0.5, 0.16, -0.28, -0.16, -0.28),
        tri(0, 0.5, -0.16, 0.28, 0.16, 0.28),
        tri(-0.5, 0, -0.28, 0.16, -0.28, -0.16),
        tri(0.5, 0, 0.28, -0.16, 0.28, 0.16),
      ].map(solid);
    case "attack":
      return [
        rect(-0.28, -0.12, 0.26, 0.3),
        rect(-0.28, -0.34, -0.16, -0.08),
        rect(-0.14, -0.36, -0.02, -0.08),
        rect(0, -0.36, 0.12, -0.08),
        rect(0.14, -0.32, 0.26, -0.08),
        rect(-0.4, -0.06, -0.24, 0.18),
        rect(-0.18, 0.3, 0.18, 0.46),
      ].map(solid);
    case "mode":
      return [
        bar(-0.38, -0.16, 0.18, -0.16, 0.1),
        tri(0.44, -0.16, 0.16, 0.02, 0.16, -0.34),
        bar(-0.18, 0.16, 0.38, 0.16, 0.1),
        tri(-0.44, 0.16, -0.16, -0.02, -0.16, 0.34),
      ].map(solid);
    case "open":
      return [
        solid(rect(-0.36, -0.06, 0.36, 0.38)),
        hole(rect(-0.12, 0.06, 0.12, 0.14)),
        solid(bar(-0.38, -0.16, 0.34, -0.38, 0.12)),
      ];
    case "map":
      return [
        solid(rect(-0.3, -0.3, 0.3, 0.3)),
        hole(rect(-0.18, -0.14, 0.18, -0.08)),
        hole(rect(-0.18, 0.04, 0.1, 0.1)),
        solid(rect(-0.4, -0.42, 0.4, -0.28)),
        solid(rect(-0.4, 0.28, 0.4, 0.42)),
      ];
    case "info":
      return question();
    case "exit":
      return [
        bar(0, 0.14, 0, -0.2, 0.14),
        tri(0, -0.48, 0.24, -0.2, -0.24, -0.2),
        rect(-0.4, 0, -0.3, 0.42),
        rect(-0.4, 0.32, 0.4, 0.42),
        rect(0.3, 0, 0.4, 0.42),
      ].map(solid);
    case "produce":
      return [
        bar(-0.3, 0.4, 0.13, -0.13, 0.12),
        bar(-0.05, -0.35, 0.35, 0.05, 0.2),
      ].map(solid);
    case "repair":
      return [
        solid(bar(-0.34, 0.34, 0.1, -0.1, 0.14)),
        solid(disc(0.18, -0.18, 0.2, 16)),
        hole(disc(0.18, -0.18, 0.09, 6)),
      ];
  }
}
/** glyphAt is the glyph of a in a box of side size centred on (cx, cy). */
function glyphAt(a: Action, cx: number, cy: number, size: number): Polygon[] {
  return glyph(a).map((p) => ({
    hole: p.hole,
    pts: p.pts.map(([x, y]): Pt => [cx + x * size, cy + y * size]),
  }));
}
