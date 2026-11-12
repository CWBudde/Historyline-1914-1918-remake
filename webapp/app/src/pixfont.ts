/**
 * The capitals pixel font of the desktop game (internal/pixfont) from the shipped atlases: Small (5 font pixels
 * high) for the status wells, Large (7) for the overlays. Text is drawn at an integer scale, device pixels per font
 * pixel, without smoothing; digits may take a colour of their own. Layout is pure, drawing goes to a 2D context.
 */
import { record } from "./assetManifest.ts";
import type { AssetLoader } from "./assets.ts";

export interface Glyph {
  x: number;
  width: number;
  advance: number;
}
export interface FaceMetrics {
  height: number;
  glyphs: ReadonlyMap<string, Glyph>;
  fallback: Glyph;
  plain: ReadonlyMap<string, string>;
  expand: ReadonlyMap<string, string>;
}
/** A colour as a CSS string, such as "#bed392". */
export type Colour = string;

function glyphOf(v: unknown): Glyph {
  const g = record(v);
  if (
    typeof g.x !== "number" ||
    typeof g.width !== "number" ||
    typeof g.advance !== "number"
  )
    throw new Error("invalid pixel glyph geometry");
  return { x: g.x, width: g.width, advance: g.advance };
}
function strings(v: unknown): Map<string, string> {
  const out = new Map<string, string>();
  for (const [k, s] of Object.entries(record(v ?? {})))
    if (typeof s === "string") out.set(k, s);
  return out;
}
/** parseMetrics reads an atlas's metrics (internal/pixfont/bitmap.Metadata). */
export function parseMetrics(v: unknown): FaceMetrics {
  const m = record(v);
  if (typeof m.height !== "number")
    throw new Error("invalid pixel font geometry");
  const glyphs = new Map<string, Glyph>();
  for (const [k, g] of Object.entries(record(m.glyphs)))
    glyphs.set(k, glyphOf(g));
  return {
    height: m.height,
    glyphs,
    fallback: glyphOf(m.fallback),
    plain: strings(m.plain),
    expand: strings(m.expand),
  };
}

/** A placed glyph: its atlas column, width and left edge in font pixels, and whether it is a digit. */
export interface Placed {
  glyph: Glyph;
  x: number;
  digit: boolean;
}

/** PixelFace measures and lays out one face; drawing needs its atlas (attach). */
export class PixelFace {
  private atlas: CanvasImageSource | null = null;
  private readonly tinted = new Map<Colour, HTMLCanvasElement>();
  constructor(readonly metrics: FaceMetrics) {}
  get height(): number {
    return this.metrics.height;
  }
  attach(atlas: CanvasImageSource): void {
    this.atlas = atlas;
    this.tinted.clear();
  }
  private lookup(r: string): Glyph {
    const m = this.metrics,
      up = r.toUpperCase();
    return (
      m.glyphs.get(up) ??
      m.glyphs.get(m.plain.get(up) ?? "") ??
      // Upper-casing may spell out a letter (ß → SS): the plain glyph of its first.
      m.glyphs.get(r) ??
      m.fallback
    );
  }
  private expanded(s: string): string {
    let out = "";
    for (const r of s) out += this.metrics.expand.get(r) ?? r;
    return out;
  }
  /** layout places every glyph of s; width is in font pixels, up to the last glyph's last column. */
  layout(s: string): { glyphs: Placed[]; width: number } {
    const glyphs: Placed[] = [];
    let x = 0;
    for (const r of this.expanded(s)) {
      const glyph = this.lookup(r);
      glyphs.push({ glyph, x, digit: r >= "0" && r <= "9" });
      x += glyph.width + 1;
    }
    return { glyphs, width: Math.max(x - 1, 0) };
  }
  /** measure is s's width and cap height in device pixels at scale (below 1 counts as 1). */
  measure(s: string, scale: number): { width: number; height: number } {
    const k = Math.max(Math.trunc(scale), 1);
    return { width: this.layout(s).width * k, height: this.height * k };
  }
  /**
   * fit is s if it fits into maxWidth device pixels at scale, otherwise its longest start that fits followed by
   * "...", trailing spaces dropped before the dots; "" when not even the dots fit.
   */
  fit(s: string, scale: number, maxWidth: number): string {
    if (this.measure(s, scale).width <= maxWidth) return s;
    const runes = Array.from(s); // runes, as Go's []rune(s)
    for (let n = runes.length - 1; n >= 0; n--) {
      const t = runes.slice(0, n).join("").trimEnd() + "...";
      if (this.measure(t, scale).width <= maxWidth) return t;
    }
    return "";
  }
  /** draw draws s with its top-left corner at (x, y) in device pixels, digits in numbers when given. */
  draw(
    ctx: CanvasRenderingContext2D,
    s: string,
    x: number,
    y: number,
    scale: number,
    text: Colour,
    numbers: Colour = text,
  ): void {
    const k = Math.max(Math.trunc(scale), 1),
      h = this.height;
    const smoothing = ctx.imageSmoothingEnabled;
    ctx.imageSmoothingEnabled = false;
    for (const p of this.layout(s).glyphs) {
      const ink = this.ink(p.digit ? numbers : text);
      if (ink === null) break;
      ctx.drawImage(
        ink,
        p.glyph.x,
        0,
        p.glyph.width,
        h,
        Math.round(x) + p.x * k,
        Math.round(y),
        p.glyph.width * k,
        h * k,
      );
    }
    ctx.imageSmoothingEnabled = smoothing;
  }
  /** drawShadow draws s over a copy in shadow one font pixel to the right and down. */
  drawShadow(
    ctx: CanvasRenderingContext2D,
    s: string,
    x: number,
    y: number,
    scale: number,
    colour: Colour,
    shadow: Colour,
  ): void {
    const k = Math.max(Math.trunc(scale), 1);
    this.draw(ctx, s, x + k, y + k, k, shadow);
    this.draw(ctx, s, x, y, k, colour);
  }
  /** ink is the atlas in one colour: white ink on transparent, tinted once per colour. */
  private ink(colour: Colour): HTMLCanvasElement | null {
    const atlas = this.atlas;
    if (atlas === null) return null;
    let canvas = this.tinted.get(colour);
    if (canvas === undefined) {
      canvas = document.createElement("canvas");
      const w = (atlas as { width: number }).width,
        h = (atlas as { height: number }).height;
      canvas.width = w;
      canvas.height = h;
      const ctx = canvas.getContext("2d");
      if (ctx === null) return null;
      ctx.drawImage(atlas, 0, 0);
      ctx.globalCompositeOperation = "source-in";
      ctx.fillStyle = colour;
      ctx.fillRect(0, 0, w, h);
      this.tinted.set(colour, canvas);
    }
    return canvas;
  }
}

export interface PixelFonts {
  small: PixelFace;
  large: PixelFace;
}

/** loadPixelFonts fetches and decodes both shipped atlases. */
export async function loadPixelFonts(
  loader: Pick<AssetLoader, "index" | "store">,
): Promise<PixelFonts> {
  const face = async (height: "5" | "7") => {
    const ref = loader.index.pixelFonts[height];
    if (ref === undefined) throw new Error("missing pixel font");
    const [metrics, bytes] = await Promise.all([
      loader.store.json(ref.metrics),
      loader.store.get(ref.image),
    ]);
    const f = new PixelFace(parseMetrics(metrics));
    f.attach(await createImageBitmap(new Blob([bytes], { type: "image/png" })));
    return f;
  };
  const [small, large] = await Promise.all([face("5"), face("7")]);
  return { small, large };
}
