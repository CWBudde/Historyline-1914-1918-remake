/**
 * The steel and brass of internal/dospanel (paint.go, surface.go, internal/texture), painted pixel by pixel into
 * RGBA pixels: olive-grey steel modulated by the painted surface/steel tile, brown name plates with brass edges
 * and rivets from the painted surface/rivets sheet, recesses and hex plates. k is device pixels per reference
 * pixel.
 */
import type { Rect } from "../camera.ts";

export type RGB = [number, number, number];
const steelDark: RGB = [0x24, 0x24, 0x1c],
  steelLight: RGB = [0xb2, 0xb2, 0xa6],
  plateBrown: RGB = [0x42, 0x34, 0x26],
  plateEdge: RGB = [0x1c, 0x14, 0x0c],
  brassLight: RGB = [0xe0, 0xbc, 0x6c],
  brass: RGB = [0xa8, 0x80, 0x3c],
  brassDark: RGB = [0x58, 0x40, 0x1c];
/** Steel tones on the dark-to-light ramp. */
export const toneBack = 0.38,
  tonePlate = 0.46,
  toneWell = 0.04,
  toneBlank = 0.52,
  toneHex = 0.58,
  bevelTone = 0.3;
/** How many device pixels one steel tile spans per reference pixel, and how strongly it modulates. */
const steelSpan = 256,
  steelGain = 0.6;

export function lerp(a: RGB, b: RGB, t: number): RGB {
  t = Math.max(0, Math.min(1, t));
  return [
    Math.round(a[0] + (b[0] - a[0]) * t),
    Math.round(a[1] + (b[1] - a[1]) * t),
    Math.round(a[2] + (b[2] - a[2]) * t),
  ];
}
const ramp = (t: number) => lerp(steelDark, steelLight, t);
/** grain is a fixed noise in [-0.5, 0.5] per cell. */
function grain(x: number, y: number): number {
  let h = (Math.imul(x, 0x9e3779b1) ^ Math.imul(y, 0x85ebca77)) >>> 0;
  h ^= h >>> 15;
  h = Math.imul(h, 0x2c1b3c6d) >>> 0;
  h ^= h >>> 12;
  return (h & 0xffff) / 0xffff - 0.5;
}

/** Tile is a texture as its deviation from its mean per channel, with halved levels (internal/texture). */
export class Tile {
  private readonly mips: { w: number; h: number; d: Float32Array }[] = [];
  constructor(px: { width: number; height: number; data: ArrayLike<number> }) {
    const w = px.width,
      h = px.height,
      d = new Float32Array(3 * w * h),
      sum = [0, 0, 0];
    for (let i = 0; i < w * h; i++)
      for (let ch = 0; ch < 3; ch++) {
        const v = px.data[4 * i + ch] ?? 0;
        d[3 * i + ch] = v;
        sum[ch] = (sum[ch] ?? 0) + v;
      }
    for (let i = 0; i < d.length; i++) {
      const mean = (sum[i % 3] ?? 0) / (w * h);
      d[i] = mean > 0 ? (d[i] ?? 0) / mean - 1 : 0;
    }
    let m = { w, h, d };
    this.mips.push(m);
    while (m.w >= 8 && m.h >= 8) {
      const hw = m.w >> 1,
        hh = m.h >> 1,
        hd = new Float32Array(3 * hw * hh),
        src = m;
      for (let y = 0; y < hh; y++)
        for (let x = 0; x < hw; x++)
          for (let ch = 0; ch < 3; ch++) {
            const at = (dx: number, dy: number) =>
              src.d[3 * ((2 * y + dy) * src.w + 2 * x + dx) + ch] ?? 0;
            hd[3 * (y * hw + x) + ch] =
              (at(0, 0) + at(1, 0) + at(0, 1) + at(1, 1)) / 4;
          }
      m = { w: hw, h: hh, d: hd };
      this.mips.push(m);
    }
  }
  /** at is the deviation at device pixel (x, y) when one tile spans span device pixels, wrapping, bilinear. */
  at(x: number, y: number, span: number): RGB {
    let m = this.mips[0];
    if (m === undefined) return [0, 0, 0];
    for (const next of this.mips.slice(1)) {
      if (m.w / span <= 1.5) break;
      m = next;
    }
    const f = m.w / span,
      u = (x + 0.5) * f - 0.5,
      v = (y + 0.5) * f - 0.5,
      x0 = Math.floor(u),
      y0 = Math.floor(v),
      fx = u - x0,
      fy = v - y0,
      wrap = (i: number, n: number) => ((i % n) + n) % n,
      out: RGB = [0, 0, 0];
    for (const [dx, dy, w] of [
      [0, 0, (1 - fx) * (1 - fy)],
      [1, 0, fx * (1 - fy)],
      [0, 1, (1 - fx) * fy],
      [1, 1, fx * fy],
    ] as const) {
      const i = 3 * (wrap(y0 + dy, m.h) * m.w + wrap(x0 + dx, m.w));
      for (let ch = 0; ch < 3; ch++)
        out[ch] = (out[ch] ?? 0) + w * (m.d[i + ch] ?? 0);
    }
    return out;
  }
}
/** modulate scales each channel of c by one plus gain times its deviation. */
export function modulate(c: RGB, gain: number, d: RGB): RGB {
  const m = (v: number, dv: number) =>
    Math.max(0, Math.min(255, Math.round(v * (1 + gain * dv))));
  return [m(c[0], d[0]), m(c[1], d[1]), m(c[2], d[2])];
}

const luma = (r: number, g: number, b: number) =>
  0.299 * r + 0.587 * g + 0.114 * b;
/** RivetSheet is the painted rivet heads (4×2 cells), sampled to paint a head of any size. */
export class RivetSheet {
  private readonly d: Float32Array;
  readonly heads: { cx: number; cy: number; r: number }[] = [];
  private readonly mean: RGB = [1, 1, 1];
  readonly w: number;
  readonly h: number;
  constructor(px: { width: number; height: number; data: ArrayLike<number> }) {
    const w = (this.w = px.width),
      h = (this.h = px.height);
    this.d = new Float32Array(4 * w * h);
    for (let i = 0; i < w * h; i++) {
      for (let ch = 0; ch < 3; ch++)
        this.d[4 * i + ch] = px.data[4 * i + ch] ?? 0;
      this.d[4 * i + 3] =
        Math.max(0, (px.data[4 * i + 3] ?? 0) / 255 - 0.5) * 2;
    }
    const sum = [0, 0, 0];
    let n = 0;
    for (let row = 0; row < 2; row++)
      for (let col = 0; col < 4; col++) {
        const x0 = Math.trunc((col * w) / 4),
          x1 = Math.trunc(((col + 1) * w) / 4),
          y0 = Math.trunc((row * h) / 2),
          y1 = Math.trunc(((row + 1) * h) / 2);
        let sx = 0,
          sy = 0,
          k = 0;
        for (let y = y0; y < y1; y++)
          for (let x = x0; x < x1; x++) {
            const i = 4 * (y * w + x);
            if ((this.d[i + 3] ?? 0) > 0) {
              sx += x + 0.5;
              sy += y + 0.5;
              k++;
              for (let ch = 0; ch < 3; ch++)
                sum[ch] = (sum[ch] ?? 0) + (this.d[i + ch] ?? 0);
            }
          }
        if (k >= 4) {
          this.heads.push({
            cx: sx / k,
            cy: sy / k,
            r: Math.sqrt(k / Math.PI),
          });
          n += k;
        }
      }
    for (let ch = 0; ch < 3; ch++)
      this.mean[ch] = Math.max(1, (sum[ch] ?? 0) / Math.max(n, 1));
  }
  at(
    head: { cx: number; cy: number; r: number },
    u: number,
    v: number,
    f: number,
    base: RGB,
  ): RGB {
    const r = head.r * 0.96,
      px = head.cx + u * r,
      py = head.cy + v * r,
      foot = f * r,
      n = Math.trunc(Math.min(8, Math.max(1, Math.ceil(foot)))),
      c: RGB = [0, 0, 0];
    let a = 0;
    for (let j = 0; j < n; j++)
      for (let i = 0; i < n; i++) {
        const x = px + foot * ((i + 0.5) / n - 0.5) - 0.5,
          y = py + foot * ((j + 0.5) / n - 0.5) - 0.5,
          x0 = Math.floor(x),
          y0 = Math.floor(y),
          fx = x - x0,
          fy = y - y0;
        for (const [dx, dy, wt] of [
          [0, 0, (1 - fx) * (1 - fy)],
          [1, 0, fx * (1 - fy)],
          [0, 1, (1 - fx) * fy],
          [1, 1, fx * fy],
        ] as const) {
          const tx = Math.min(this.w - 1, Math.max(0, x0 + dx)),
            ty = Math.min(this.h - 1, Math.max(0, y0 + dy)),
            k = 4 * (ty * this.w + tx),
            w = wt * (this.d[k + 3] ?? 0);
          for (let ch = 0; ch < 3; ch++)
            c[ch] = (c[ch] ?? 0) + w * (this.d[k + ch] ?? 0);
          a += w;
        }
      }
    let col = c;
    if (a <= 0) {
      col = [...this.mean];
      a = 1;
    }
    const k =
      luma(...base) / luma(this.mean[0], this.mean[1], this.mean[2]) / a;
    return col.map((v) => Math.max(0, Math.min(255, Math.round(v * k)))) as RGB;
  }
}

export interface Surfaces {
  steel: Tile | null;
  rivets: RivetSheet | null;
}

/** Painter paints placeholder steel and brass into an RGBA buffer of w × h device pixels at scale k. */
export class Painter {
  constructor(
    readonly px: Uint8ClampedArray,
    readonly w: number,
    readonly h: number,
    readonly k: number,
    readonly surfaces: Surfaces,
  ) {}
  /** bw is the width of a bevel or edge line. */
  bw(): number {
    return Math.max(1, Math.round(this.k));
  }
  private set(x: number, y: number, c: RGB): void {
    const i = 4 * (y * this.w + x);
    this.px[i] = c[0];
    this.px[i + 1] = c[1];
    this.px[i + 2] = c[2];
    this.px[i + 3] = 255;
  }
  private get(x: number, y: number): RGB {
    const i = 4 * (y * this.w + x);
    return [this.px[i] ?? 0, this.px[i + 1] ?? 0, this.px[i + 2] ?? 0];
  }
  blend(x: number, y: number, c: RGB, a: number): void {
    if (a >= 1) this.set(x, y, c);
    else this.set(x, y, lerp(this.get(x, y), c, a));
  }
  private clip(r: Rect): [number, number, number, number] {
    return [
      Math.max(r.x, 0),
      Math.max(r.y, 0),
      Math.min(r.x + r.width, this.w),
      Math.min(r.y + r.height, this.h),
    ];
  }
  steelColour(x: number, y: number, t: number): RGB {
    const s = this.surfaces.steel;
    if (s !== null)
      return modulate(ramp(t), steelGain, s.at(x, y, this.k * steelSpan));
    const q = Math.max(1, Math.round(this.k));
    return ramp(t + 0.06 * grain(Math.trunc(x / q), Math.trunc(y / q)));
  }
  /** steel fills r with steel of tone t. */
  steel(r: Rect, t: number): void {
    const [x0, y0, x1, y1] = this.clip(r);
    for (let y = y0; y < y1; y++)
      for (let x = x0; x < x1; x++) this.set(x, y, this.steelColour(x, y, t));
  }
  /** edges calls fn for the pixels within w of r's border: lit on the top and left sides. */
  private edges(
    r: Rect,
    w: number,
    fn: (x: number, y: number, lit: boolean) => void,
  ): void {
    const [x0, y0, x1, y1] = this.clip(r);
    for (let y = y0; y < y1; y++)
      for (let x = x0; x < x1; x++) {
        const dl = Math.min(y - r.y, x - r.x),
          dr = Math.min(r.y + r.height - 1 - y, r.x + r.width - 1 - x);
        if (dl < w && dl <= dr) fn(x, y, true);
        else if (dr < w) fn(x, y, false);
      }
  }
  bevel(r: Rect, t: number, raised: boolean): void {
    this.edges(r, this.bw(), (x, y, lit) => {
      this.set(
        x,
        y,
        this.steelColour(x, y, t + (lit !== raised ? -bevelTone : bevelTone)),
      );
    });
  }
  plate(r: Rect, t: number): void {
    this.steel(r, t);
    this.bevel(r, t, true);
  }
  well(r: Rect): void {
    this.steel(r, toneWell);
    this.bevel(r, toneBack, false);
  }
  /** namePlate is a brown plate with a dark edge, a brass line and a brass rivet in each corner. */
  namePlate(r: Rect): void {
    const [x0, y0, x1, y1] = this.clip(r),
      w = this.bw(),
      s = this.surfaces.steel;
    for (let y = y0; y < y1; y++)
      for (let x = x0; x < x1; x++)
        this.set(
          x,
          y,
          s !== null
            ? modulate(
                lerp(plateBrown, plateEdge, 0.12),
                1.5 * steelGain,
                s.at(x, y, this.k * steelSpan),
              )
            : lerp(
                plateBrown,
                plateEdge,
                0.12 + 0.24 * grain(Math.trunc(x / w), Math.trunc(y / w)),
              ),
        );
    this.edges(r, w, (x, y) => {
      this.set(x, y, plateEdge);
    });
    this.edges(
      {
        x: r.x + w,
        y: r.y + w,
        width: r.width - 2 * w,
        height: r.height - 2 * w,
      },
      w,
      (x, y, lit) => {
        this.set(x, y, lit ? brass : brassDark);
      },
    );
    const inset = 3.5 * this.k,
      rad = Math.max(1.2, 1.3 * this.k),
      xa = r.x + inset,
      xb = r.x + r.width - inset;
    let ys = [r.y + inset, r.y + r.height - inset];
    if ((ys[1] ?? 0) - (ys[0] ?? 0) < 4 * rad) ys = [r.y + r.height / 2];
    for (const y of ys) {
      this.rivet(xa, y, rad);
      this.rivet(xb, y, rad);
    }
  }
  /** rivet is a brass head of radius rad at (cx, cy) with a soft shadow to the bottom right. */
  rivet(cx: number, cy: number, rad: number): void {
    const sheet = this.surfaces.rivets,
      head =
        sheet === null || sheet.heads.length === 0
          ? null
          : (sheet.heads[
              Math.min(
                sheet.heads.length - 1,
                Math.trunc(
                  (grain(Math.trunc(cx), Math.trunc(cy)) + 0.5) *
                    sheet.heads.length,
                ),
              )
            ] ?? null);
    for (
      let y = Math.trunc(cy - rad - 2);
      y <= Math.trunc(cy + rad * 1.5 + 2);
      y++
    )
      for (
        let x = Math.trunc(cx - rad - 2);
        x <= Math.trunc(cx + rad * 1.5 + 2);
        x++
      ) {
        if (x < 0 || y < 0 || x >= this.w || y >= this.h) continue;
        const dx = x + 0.5 - cx,
          dy = y + 0.5 - cy,
          d = Math.hypot(dx, dy),
          sd = Math.hypot(dx - rad * 0.4, dy - rad * 0.4);
        if (sd < rad + 0.5 && d >= rad)
          this.blend(x, y, ramp(0.1), 0.5 * Math.min(1, rad + 0.5 - sd));
        const cov = Math.min(1, rad + 0.5 - d);
        if (cov <= 0) continue;
        if (sheet !== null && head !== null) {
          this.blend(
            x,
            y,
            sheet.at(head, dx / rad, dy / rad, 1 / rad, brass),
            cov,
          );
          continue;
        }
        const nx = dx / rad,
          ny = dy / rad,
          l =
            -0.5 * nx -
            0.5 * ny +
            0.707 * Math.sqrt(Math.max(0, 1 - nx * nx - ny * ny));
        this.blend(
          x,
          y,
          l > 0.5
            ? lerp(brass, brassLight, (l - 0.5) * 2)
            : lerp(brassDark, brass, l * 2),
          cov,
        );
      }
  }
  /** hex fills the hex of half size a × b round (cx, cy) with steel of tone t, its edge lit when raised. */
  hex(
    cx: number,
    cy: number,
    a: number,
    b: number,
    t: number,
    raised: boolean,
  ): void {
    const bw = this.bw(),
      sign = raised ? 1 : -1;
    for (let y = Math.trunc(cy - b - 1); y <= Math.trunc(cy + b + 1); y++)
      for (let x = Math.trunc(cx - a - 1); x <= Math.trunc(cx + a + 1); x++) {
        const [d, l] = hexEdge(x + 0.5 - cx, y + 0.5 - cy, a, b),
          cov = Math.min(1, d + 0.5);
        if (cov <= 0 || x < 0 || y < 0 || x >= this.w || y >= this.h) continue;
        const bevel = Math.max(0, Math.min(1, bw - d + 0.5));
        this.blend(
          x,
          y,
          this.steelColour(x, y, t + bevel * sign * bevelTone * l),
          cov,
        );
      }
  }
}

/** hexEdge is the distance of (dx, dy) from the edge of a flat-topped hex (negative outside) and its light. */
export function hexEdge(
  dx: number,
  dy: number,
  a: number,
  b: number,
): [number, number] {
  const ax = Math.abs(dx),
    ay = Math.abs(dy),
    sx = Math.sign(dx) || 1,
    sy = Math.sign(dy) || 1,
    top = b - ay,
    s = Math.hypot(b, (2 * a) / 3),
    slant = ((a - ax - ((2 * a) / 3) * (ay / b)) * b) / s;
  if (top <= slant) return [top, (-sy * Math.SQRT2) / 2];
  const nx = (sx * b) / s,
    ny = (sy * ((2 * a) / 3)) / s;
  return [slant, (-(nx + ny) * Math.SQRT2) / 2];
}
/** hexOf is the centre and half size of the hex inscribed in r. */
export function hexOf(r: Rect): [number, number, number, number] {
  return [r.x + r.width / 2, r.y + r.height / 2, r.width / 2, r.height / 2];
}
