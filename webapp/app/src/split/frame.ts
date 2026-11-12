/** Canvas overlay for the native internal/dosscreen frame. Map pixels remain
 * transparent; this owns only its canvas and decoded screen textures. */
import { hexgrid } from "@historyline/engine";
import { assetDensity, type AssetLoader, type LoadedImage } from "../assets.ts";
import { parseImageCatalog } from "../assetManifest.ts";
import type { Point, Rect } from "../camera.ts";
import { outline, rivets, type Screen, type View } from "./geometry.ts";

type RGB = [number, number, number];
interface Painting {
  bitmap: ImageBitmap;
  box: Rect;
  mean: RGB;
}
const steelDark: RGB = [36, 36, 28],
  steelLight: RGB = [178, 178, 166],
  gunmetalDark: RGB = [19, 20, 23],
  gunmetalLight: RGB = [46, 47, 51],
  brass: RGB = [140, 116, 72],
  brassLight: RGB = [194, 164, 102];
const css = (c: RGB) => `rgb(${c.map(Math.round).join(",")})`;
const mix = (a: RGB, b: RGB, t: number): RGB => {
  t = Math.max(0, Math.min(1, t));
  return [
    a[0] + (b[0] - a[0]) * t,
    a[1] + (b[1] - a[1]) * t,
    a[2] + (b[2] - a[2]) * t,
  ];
};
const scaled = (n: number, ui: number) => Math.max(Math.round(n * 2 * ui), 1);
function context(canvas: HTMLCanvasElement): CanvasRenderingContext2D {
  const ctx = canvas.getContext("2d", { willReadFrequently: true });
  if (!ctx) throw new Error("missing split frame canvas context");
  return ctx;
}

export class SplitFrame {
  readonly canvas = document.createElement("canvas");
  private readonly ctx = context(this.canvas);
  private readonly paintings = new Map<string, Painting>();
  private readonly noise = material();
  private readonly gunmetal = material(gunmetalDark, gunmetalLight);
  private steel: HTMLCanvasElement | null = null;
  private slimFrame: HTMLCanvasElement | null = null;
  private slimKey = "";
  private version = 0;
  private generation = 0;
  private density = 0;
  private pending: Promise<void> = Promise.resolve();
  private stopped = false;
  private readonly abort = new AbortController();
  private key = "";

  constructor(
    private readonly loader: AssetLoader,
    private readonly changed: () => void,
  ) {
    this.canvas.style.pointerEvents = "none";
    this.canvas.setAttribute("aria-hidden", "true");
  }

  initialize(uiScale: number, dpr: number): Promise<void> {
    if (this.stopped) return Promise.resolve();
    const desired = assetDensity(uiScale, dpr),
      packs = this.loader.index.packs.screens;
    const densities = Object.keys(packs ?? {})
        .map(Number)
        .sort((a, b) => a - b),
      density = densities.find((d) => d >= desired) ?? densities.at(-1);
    if (density === undefined) return Promise.resolve();
    if (this.density === density) return this.pending;
    this.density = density;
    const generation = ++this.generation;
    this.pending = this.load(density, generation);
    return this.pending;
  }

  private async load(density: number, generation: number): Promise<void> {
    try {
      const pair = this.loader.index.packs.screens?.[String(density)];
      if (!pair) return;
      const catalog = parseImageCatalog(
          await this.loader.store.json(pair.lite),
          density,
        ),
        ids = new Set(catalog.entries.map((entry) => entry.id)),
        selected: string[] = [];
      if (ids.has("surface/steel")) selected.push("surface/steel");
      for (let i = 1; ; i++) {
        const id = i === 1 ? "surface/plate" : `surface/plate-${i}`;
        if (!ids.has(id)) break;
        selected.push(id);
      }
      if (!selected.length || this.stopped || generation !== this.generation)
        return;
      await this.loader.progressive(
        "screens",
        selected,
        density,
        1,
        async (image) => {
          const painting = await decode(
            image,
            image.entry.id !== "surface/steel",
          );
          if (!painting) return;
          if (this.stopped || generation !== this.generation) {
            painting.bitmap.close();
            return;
          }
          this.paintings.get(image.entry.id)?.bitmap.close();
          this.paintings.set(image.entry.id, painting);
          this.steel = release(this.steel);
          this.version++;
          this.changed();
        },
        this.abort.signal,
      );
    } catch {
      // As in the native frame, absent art keeps the drawn material; a failed
      // full upgrade keeps the already decoded lite texture.
    }
  }

  paint(
    screen: Screen,
    views: [View, View],
    grid: hexgrid.Grid,
    slim: boolean,
    dpr: number,
  ): void {
    if (this.stopped) return;
    const key = JSON.stringify([screen, views, grid, slim, dpr, this.version]);
    if (key === this.key) return;
    this.key = key;
    const c = this.canvas;
    c.width = screen.width;
    c.height = screen.height;
    c.style.width = `${screen.width / dpr}px`;
    c.style.height = `${screen.height / dpr}px`;
    const ctx = this.ctx;
    ctx.imageSmoothingEnabled = true;
    ctx.imageSmoothingQuality = "high";
    if (slim) {
      // Panning changes the world boundary, but not the frame material or its
      // openings. Retain those exact pixels instead of filling the pattern again.
      const slimKey = JSON.stringify([screen, dpr, this.version]);
      if (this.slimFrame === null || this.slimKey !== slimKey) {
        this.slimFrame = release(this.slimFrame);
        this.slim(screen);
        const cached = document.createElement("canvas");
        cached.width = screen.width;
        cached.height = screen.height;
        context(cached).drawImage(c, 0, 0);
        this.slimFrame = cached;
        this.slimKey = slimKey;
      } else ctx.drawImage(this.slimFrame, 0, 0);
      const [width, height] = hexgrid.bounds(grid);
      for (const side of [0, 1] as const)
        this.beyond(
          screen.slimOpening(side),
          views[side],
          width,
          height,
          screen.ui,
        );
    } else {
      this.fillSteel(screen.ui);
      for (const view of views) this.panel(view, grid);
      this.post(screen);
      this.rivet(
        {
          x: screen.bolt.x + screen.bolt.width / 2,
          y: screen.bolt.y + screen.bolt.height / 2,
        },
        screen.bolt.width / 2 - 0.5,
        false,
      );
      for (const well of screen.status) this.well(well, scaled(1, screen.ui));
    }
  }

  private plates(): Painting[] {
    return [...this.paintings.entries()]
      .filter(([id]) => id !== "surface/steel")
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([, painting]) => painting);
  }

  private face(): RGB {
    const plates = this.plates();
    if (!plates.length) return mix(steelDark, steelLight, 0.46);
    return plates.reduce<RGB>(
      (sum, p) => [
        sum[0] + p.mean[0] / plates.length,
        sum[1] + p.mean[1] / plates.length,
        sum[2] + p.mean[2] / plates.length,
      ],
      [0, 0, 0],
    );
  }

  private tone(t: number): RGB {
    const dt = t - 0.46,
      face = this.face();
    return dt >= 0
      ? mix(face, steelLight, dt / 0.54)
      : mix(face, steelDark, -dt / 0.46);
  }

  private fillSteel(ui: number): void {
    const ctx = this.ctx,
      painting = this.paintings.get("surface/steel");
    let source = this.noise;
    if (painting && this.plates().length) {
      this.steel ??= graded(painting, this.face());
      source = this.steel;
    }
    const pattern = ctx.createPattern(source, "repeat");
    if (pattern) {
      if (source !== this.noise)
        pattern.setTransform(
          new DOMMatrix().scale(
            (512 * ui) / source.width,
            (512 * ui) / source.height,
          ),
        );
      ctx.fillStyle = pattern;
    } else ctx.fillStyle = css(this.face());
    ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);
  }

  private panel(view: View, grid: hexgrid.Grid): void {
    const ctx = this.ctx,
      p = view.panel,
      plates = this.plates();
    ctx.save();
    ctx.beginPath();
    ctx.rect(p.x, p.y, p.width, p.height);
    ctx.clip();
    if (plates.length) {
      ctx.fillStyle = css(steelDark);
      ctx.fillRect(p.x, p.y, p.width, p.height);
      const a = view.screenToWorld(p.x, p.y),
        b = view.screenToWorld(p.x + p.width, p.y + p.height),
        origin = view.origin();
      for (
        let line = Math.floor(a.y / hexgrid.linePitch) - 2;
        line <= Math.ceil(b.y / hexgrid.linePitch) + 1;
        line++
      ) {
        for (
          let col = Math.floor(a.x / hexgrid.colPitch) - 2;
          col <= Math.ceil(b.x / hexgrid.colPitch) + 1;
          col++
        ) {
          const at = { line, col };
          if (view.shows(grid, at)) continue;
          const [cx, cy] = hexgrid.center(at),
            kx = Math.round((cx - origin.x) / hexgrid.halfPitch),
            ky = Math.round((cy - origin.y) / hexgrid.linePitch),
            plate =
              plates[
                Math.min(
                  plates.length - 1,
                  Math.floor(lattice(kx, ky, p.x) * plates.length),
                )
              ];
          if (!plate) continue;
          const top = view.worldToScreen(cx - 48, cy - 40),
            box = plate.box;
          ctx.save();
          ctx.clip(hexPath(view, at));
          ctx.drawImage(
            plate.bitmap,
            box.x,
            box.y,
            box.width,
            box.height,
            top.x,
            top.y,
            96 * view.scale,
            80 * view.scale,
          );
          ctx.restore();
        }
      }
    }
    const openings = new Path2D();
    for (const at of view.cells())
      if (hexgrid.contains(grid, at)) openings.addPath(hexPath(view, at));
    ctx.globalCompositeOperation = "destination-out";
    ctx.fill(openings);
    ctx.globalCompositeOperation = "source-over";
    const steel = new Path2D();
    steel.rect(p.x, p.y, p.width, p.height);
    steel.addPath(openings);
    ctx.clip(steel, "evenodd");
    const lip = Math.max(2 * view.scale, 1),
      bevel = plates.length ? 0 : Math.max(6 * view.scale, 2);
    ctx.lineJoin = "round";
    ctx.lineCap = "round";
    for (const { a, b, n } of outline(view, grid)) {
      ctx.beginPath();
      ctx.moveTo(a.x, a.y);
      ctx.lineTo(b.x, b.y);
      if (bevel) {
        ctx.lineWidth = 2 * (lip + bevel);
        ctx.strokeStyle = css(
          this.tone(0.46 + (0.42 * -(n.x + n.y)) / Math.SQRT2),
        );
        ctx.stroke();
      }
      ctx.lineWidth = 2 * lip;
      ctx.strokeStyle = css(mix(steelDark, steelLight, 0.06));
      ctx.stroke();
    }
    if (!plates.length)
      for (const point of rivets(view, grid))
        this.rivet(point, Math.max(4.5 * view.scale, 1.5), false);
    ctx.restore();
  }

  private post(screen: Screen): void {
    const ctx = this.ctx,
      p = screen.divider,
      line = Math.max(Math.round(screen.ui), 1),
      highlight = p.x + Math.trunc((p.width * 3) / 10);
    for (let x = p.x; x < p.x + p.width; x++) {
      let t = 0.3 + 0.25 * Math.sin((Math.PI * (x - p.x + 0.5)) / p.width);
      if (x === p.x || x === p.x + p.width - 1) t = 0.12;
      else if (x >= highlight && x < highlight + line) t = 0.92;
      ctx.fillStyle = css(this.tone(t));
      ctx.fillRect(x, p.y, 1, p.height);
    }
  }

  private well(r: Rect, width: number): void {
    const ctx = this.ctx;
    ctx.fillStyle = "#24180c";
    ctx.fillRect(r.x, r.y, r.width, r.height);
    ctx.fillStyle = "#453828";
    ctx.beginPath();
    ctx.moveTo(r.x, r.y + r.height);
    ctx.lineTo(r.x + r.width, r.y);
    ctx.lineTo(r.x + r.width, r.y + r.height);
    ctx.closePath();
    ctx.fill();
    ctx.fillStyle = "#181c00";
    ctx.fillRect(
      r.x + width,
      r.y + width,
      Math.max(r.width - width * 2, 0),
      Math.max(r.height - width * 2, 0),
    );
  }

  private rivet(point: Point, radius: number, gold: boolean): void {
    if (radius <= 0) return;
    const ctx = this.ctx,
      { x, y } = point;
    ctx.fillStyle = "rgba(5,5,6,.45)";
    ctx.beginPath();
    ctx.arc(x + radius * 0.35, y + radius * 0.35, radius, 0, Math.PI * 2);
    ctx.fill();
    const gradient = ctx.createRadialGradient(
      x - radius * 0.35,
      y - radius * 0.35,
      0,
      x,
      y,
      radius,
    );
    gradient.addColorStop(0, css(gold ? [244, 226, 180] : this.tone(0.92)));
    gradient.addColorStop(0.5, css(gold ? brass : this.tone(0.6)));
    gradient.addColorStop(1, css(gold ? [46, 34, 18] : this.tone(0.12)));
    ctx.fillStyle = gradient;
    ctx.beginPath();
    ctx.arc(x, y, radius, 0, Math.PI * 2);
    ctx.fill();
  }

  private slim(screen: Screen): void {
    const ctx = this.ctx,
      line = Math.max(Math.round(screen.ui), 1),
      rw = 4 * line,
      left = screen.slimOpening(0),
      right = screen.slimOpening(1);
    ctx.fillStyle =
      ctx.createPattern(this.gunmetal, "repeat") ??
      css(mix(gunmetalDark, gunmetalLight, 0.46));
    ctx.fillRect(0, 0, screen.width, screen.height);
    const postX = left.x + left.width + rw,
      postEnd = right.x - rw;
    for (let x = postX; x < postEnd; x++) {
      const u = (x - postX + 0.5) / (postEnd - postX),
        lift = 0.3 * Math.sin(Math.PI * Math.min(1, u * 1.25)) ** 2;
      ctx.globalCompositeOperation = "lighter";
      ctx.fillStyle = css([27 * lift, 27 * lift, 28 * lift]);
      ctx.fillRect(x, 0, 1, screen.height);
    }
    ctx.globalCompositeOperation = "source-over";
    for (const side of [0, 1] as const) {
      const o = screen.slimOpening(side),
        st = screen.status[side],
        y = o.y + o.height + rw,
        strip = {
          x: o.x,
          y,
          width: o.width,
          height: Math.max(st.y + st.height - y, 0),
        };
      ctx.globalCompositeOperation = "lighter";
      ctx.fillStyle = css([27 * 0.12, 27 * 0.12, 28 * 0.12]);
      ctx.fillRect(strip.x, strip.y, strip.width, strip.height);
      ctx.globalCompositeOperation = "source-over";
      this.ring(strip, line);
    }
    for (const side of [0, 1] as const) {
      const o = screen.slimOpening(side);
      this.ring(o, line);
      ctx.clearRect(o.x, o.y, o.width, o.height);
    }
    const b = screen.bolt;
    this.rivet(
      { x: b.x + b.width / 2, y: b.y + b.height / 2 },
      b.width / 2 - 0.5,
      true,
    );
  }

  private ring(r: Rect, line: number): void {
    const ctx = this.ctx;
    const band = (d0: number, d1: number, near: string, far: string) => {
      const w = d1 - d0;
      ctx.fillStyle = far;
      ctx.fillRect(r.x - d1, r.y - d1, r.width + 2 * d1, w);
      ctx.fillRect(r.x - d1, r.y - d0, w, r.height + 2 * d0);
      ctx.fillStyle = near;
      ctx.fillRect(r.x - d1, r.y + r.height + d0, r.width + 2 * d1, w);
      ctx.fillRect(r.x + r.width + d0, r.y - d0, w, r.height + 2 * d0);
    };
    band(0, line, "#0b0a08", "#0b0a08");
    band(line, 2 * line, css(brassLight), css(brass));
    band(2 * line, 3 * line, css(brass), css(brassLight));
    band(3 * line, 4 * line, "#050506", "#050506");
  }

  /** Native DrawBeyond: the playable map ends in gunmetal and a brass ring,
   * even when the underlying painted backdrop extends farther. */
  private beyond(
    open: Rect,
    view: View,
    width: number,
    height: number,
    ui: number,
  ): void {
    const a = view.worldToScreen(0, 0),
      b = view.worldToScreen(width, height),
      round = (n: number) => (n < 0 ? -Math.round(-n) : Math.round(n)),
      world = {
        x: round(a.x),
        y: round(a.y),
        width: round(b.x) - round(a.x),
        height: round(b.y) - round(a.y),
      },
      x0 = Math.max(open.x, world.x),
      y0 = Math.max(open.y, world.y),
      x1 = Math.min(open.x + open.width, world.x + world.width),
      y1 = Math.min(open.y + open.height, world.y + world.height);
    if (
      x0 === open.x &&
      y0 === open.y &&
      x1 === open.x + open.width &&
      y1 === open.y + open.height
    )
      return;
    const ctx = this.ctx;
    ctx.save();
    ctx.beginPath();
    ctx.rect(open.x, open.y, open.width, open.height);
    ctx.clip();
    ctx.fillStyle =
      ctx.createPattern(this.gunmetal, "repeat") ??
      css(mix(gunmetalDark, gunmetalLight, 0.46));
    if (x1 <= x0 || y1 <= y0)
      ctx.fillRect(open.x, open.y, open.width, open.height);
    else {
      ctx.fillRect(open.x, open.y, open.width, y0 - open.y);
      ctx.fillRect(open.x, y1, open.width, open.y + open.height - y1);
      ctx.fillRect(open.x, y0, x0 - open.x, y1 - y0);
      ctx.fillRect(x1, y0, open.x + open.width - x1, y1 - y0);
      this.ring(world, Math.max(Math.round(ui), 1));
    }
    ctx.restore();
  }

  dispose(): void {
    this.stopped = true;
    this.abort.abort();
    this.generation++;
    for (const painting of this.paintings.values()) painting.bitmap.close();
    this.paintings.clear();
    this.steel = release(this.steel);
    this.slimFrame = release(this.slimFrame);
    release(this.noise);
    release(this.gunmetal);
    release(this.canvas);
    this.canvas.remove();
  }
}

function release(canvas: HTMLCanvasElement | null): null {
  if (canvas) canvas.width = canvas.height = 0;
  return null;
}

function hexPath(view: View, at: hexgrid.Coord): Path2D {
  const path = new Path2D();
  for (const [i, [x = 0, y = 0]] of hexgrid.corners(at).entries()) {
    const p = view.worldToScreen(x, y);
    if (i === 0) path.moveTo(p.x, p.y);
    else path.lineTo(p.x, p.y);
  }
  path.closePath();
  return path;
}

async function decode(
  image: LoadedImage,
  crop: boolean,
): Promise<Painting | null> {
  const bitmap = await createImageBitmap(new Blob([image.bytes]));
  try {
    if (
      bitmap.width !== image.entry.width ||
      bitmap.height !== image.entry.height
    )
      throw new Error("split frame image geometry differs");
    const scratch = document.createElement("canvas");
    scratch.width = bitmap.width;
    scratch.height = bitmap.height;
    const ctx = context(scratch);
    ctx.drawImage(bitmap, 0, 0);
    const data = ctx.getImageData(0, 0, scratch.width, scratch.height).data;
    let x0 = bitmap.width,
      y0 = bitmap.height,
      x1 = 0,
      y1 = 0;
    const sum: [number, number, number, number] = [0, 0, 0, 0];
    for (let y = 0; y < bitmap.height; y++)
      for (let x = 0; x < bitmap.width; x++) {
        const i = (y * bitmap.width + x) * 4,
          alpha = (data[i + 3] ?? 0) / 255;
        if (alpha > 0.5) {
          x0 = Math.min(x0, x);
          y0 = Math.min(y0, y);
          x1 = Math.max(x1, x + 1);
          y1 = Math.max(y1, y + 1);
        }
        sum[0] += (data[i] ?? 0) * alpha;
        sum[1] += (data[i + 1] ?? 0) * alpha;
        sum[2] += (data[i + 2] ?? 0) * alpha;
        sum[3] += alpha;
      }
    release(scratch);
    if (!sum[3] || (crop && (x0 >= x1 || y0 >= y1))) {
      bitmap.close();
      return null;
    }
    return {
      bitmap,
      box: crop
        ? { x: x0, y: y0, width: x1 - x0, height: y1 - y0 }
        : { x: 0, y: 0, width: bitmap.width, height: bitmap.height },
      mean: [sum[0] / sum[3], sum[1] / sum[3], sum[2] / sum[3]],
    };
  } catch (error) {
    bitmap.close();
    throw error;
  }
}

/** The native steel tile's channel deviation modulates the plates' mean. */
function graded(painting: Painting, face: RGB): HTMLCanvasElement {
  const canvas = document.createElement("canvas");
  canvas.width = painting.bitmap.width;
  canvas.height = painting.bitmap.height;
  const ctx = context(canvas);
  ctx.drawImage(painting.bitmap, 0, 0);
  const pixels = ctx.getImageData(0, 0, canvas.width, canvas.height);
  for (let i = 0; i < pixels.data.length; i += 4) {
    for (const ch of [0, 1, 2] as const) {
      const v = pixels.data[i + ch] ?? 0,
        mean = painting.mean[ch];
      pixels.data[i + ch] = Math.round(
        face[ch] * (1 + 0.6 * (mean > 0 ? v / mean - 1 : 0)),
      );
    }
    pixels.data[i + 3] = 255;
  }
  ctx.putImageData(pixels, 0, 0);
  return canvas;
}

function lattice(i: number, j: number, seed: number): number {
  let h =
    Math.imul(i, 0x8da6b343) ^
    Math.imul(j, 0xd8163841) ^
    Math.imul(seed, 0xcb1ab31f);
  h ^= h >>> 13;
  h = Math.imul(h, 0x5bd1e995);
  h ^= h >>> 15;
  return (h & 0xffff) / 0xffff;
}

function noise(
  x: number,
  y: number,
  px: number,
  py: number,
  seed: number,
): number {
  const ix = Math.floor(x),
    iy = Math.floor(y),
    fx = (x - ix) ** 2 * (3 - 2 * (x - ix)),
    fy = (y - iy) ** 2 * (3 - 2 * (y - iy)),
    at = (i: number, j: number) =>
      lattice(((i % px) + px) % px, ((j % py) + py) % py, seed),
    top = at(ix, iy) + fx * (at(ix + 1, iy) - at(ix, iy)),
    bottom = at(ix, iy + 1) + fx * (at(ix + 1, iy + 1) - at(ix, iy + 1));
  return top + fy * (bottom - top);
}

function material(dark = steelDark, light = steelLight): HTMLCanvasElement {
  const canvas = document.createElement("canvas");
  canvas.width = canvas.height = 128;
  const ctx = context(canvas),
    pixels = ctx.createImageData(128, 128);
  for (let y = 0; y < 128; y++)
    for (let x = 0; x < 128; x++) {
      let t = 0.46;
      for (const [k, amp] of [0.1, 0.06, 0.04, 0.025].entries()) {
        const p = 4 << k;
        t += amp * (noise((x / 128) * p, (y / 128) * p, p, p, k) - 0.5);
      }
      t += 0.05 * (noise(x / 64, y / 2, 2, 64, 7) - 0.5);
      const color = mix(dark, light, t),
        i = (y * 128 + x) * 4;
      pixels.data[i] = Math.round(color[0]);
      pixels.data[i + 1] = Math.round(color[1]);
      pixels.data[i + 2] = Math.round(color[2]);
      pixels.data[i + 3] = 255;
    }
  ctx.putImageData(pixels, 0, 0);
  return canvas;
}
