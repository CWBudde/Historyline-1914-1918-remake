import {
  Application,
  Container,
  Graphics,
  Sprite,
  Text,
  ColorMatrixFilter,
} from "pixi.js";
import { hexgrid, level, play } from "@historyline/engine";
import { MapImages, type MapImage } from "../mapImages.ts";
import { assetDensity, type AssetLoader } from "../assets.ts";
import { spriteID, type SpriteFamily } from "../unitFrame.ts";
import type { MapBattle } from "../mapBattle.ts";
import type { Point, Rect } from "../camera.ts";
import { battleLayout, figurePoint } from "./layout.ts";

const ownerColor = (owner: number) =>
  owner === 1 ? 0x55c3f3 : owner === 2 ? 0xe7d365 : 0xc4c4b8;
const rank = ["", "01", "02", "03", "04", "05", "10"];
function sepia(): ColorMatrixFilter {
  const f = new ColorMatrixFilter();
  f.matrix = [
    0.393, 0.769, 0.189, 0, 0, 0.349, 0.686, 0.168, 0, 0, 0.272, 0.534, 0.131,
    0, 0, 0, 0, 0, 1, 0,
  ];
  return f;
}
function cropped(image: MapImage, ratio: number): MapImage {
  let width = image.width,
    height = image.height;
  if (width / height > ratio) width = height * ratio;
  else height = width / ratio;
  return {
    ...image,
    frame: {
      x: (image.width - width) / 2,
      y: (image.height - height) / 2,
      width,
      height,
    },
  };
}
/** Persistent auxiliary Pixi canvas: ready before short reduced-motion scenes,
 * attached to either dock host without recreating its graphics context. */
export class BattleScenes {
  private readonly app = new Application();
  private readonly plates = new Graphics();
  private readonly ground = new Graphics();
  private readonly pictures = new Container();
  private readonly emblems = new Container();
  private readonly flag = new Graphics();
  private readonly flashes = new Graphics();
  private readonly labels = new Container();
  private readonly tone = sepia();
  private readonly images: MapImages;
  private readonly sprites = new Map<string, Sprite>();
  private readonly texts = new Map<string, Text>();
  private readonly used = new Set<string>();
  private readonly order = new Map<Container, number>();
  private screen: ReadonlyMap<string, MapImage> = new Map();
  private units: ReadonlyMap<string, MapImage> = new Map();
  private readonly hosts: Record<"battle" | "strategy", HTMLDivElement | null> =
    { battle: null, strategy: null };
  private host: HTMLDivElement | null = null;
  private initialized = false;
  private stopped = false;
  private dirty = true;
  private density = 1;
  private key = "";
  private lastBattle: play.BattleInfo | null = null;
  private mini: { scale: number; offset: Point } | null = null;
  private palette = new Map<string, number>();
  private paletteRequested = false;
  constructor(
    private readonly battle: MapBattle,
    private readonly loader: AssetLoader,
    private readonly families: Readonly<Record<string, SpriteFamily>>,
    private readonly scale: () => number,
    private readonly viewport: () => Rect,
    private readonly choose: (at: hexgrid.Coord) => void,
    private readonly error: (e: Error) => void,
  ) {
    this.images = new MapImages(
      loader,
      () => {
        this.dirty = true;
      },
      error,
    );
  }
  async initialize(): Promise<void> {
    if (this.isStopped()) return;
    await this.app.init({
      width: 304,
      height: 360,
      resolution: window.devicePixelRatio,
      autoDensity: true,
      background: 0x24241c,
      antialias: true,
      autoStart: false,
    });
    if (this.isStopped()) {
      this.app.destroy(true, { children: true });
      return;
    }
    this.initialized = true;
    this.pictures.filters = [this.tone];
    this.emblems.addChild(this.flag);
    this.app.stage.addChild(
      this.plates,
      this.ground,
      this.pictures,
      this.emblems,
      this.flashes,
      this.labels,
    );
    this.app.canvas.setAttribute("aria-label", "Battle scene or strategic map");
    this.app.canvas.addEventListener("pointerup", (e) => {
      if (this.battle.playback.scene() !== null) {
        this.battle.playback.skipScene();
        this.battle.version++;
        this.battle.onChange?.();
        return;
      }
      if (this.battle.panel !== "map" || this.mini === null) return;
      const r = this.app.canvas.getBoundingClientRect(),
        x = (e.clientX - r.left - this.mini.offset.x) / this.mini.scale,
        y = (e.clientY - r.top - this.mini.offset.y) / this.mini.scale;
      const [at, onMap] = hexgrid.pixelToHex(
        level.grid(this.battle.session.level),
        x,
        y,
      );
      if (onMap) this.choose(at);
    });
    this.app.ticker.add(() => {
      this.draw();
    });
    this.app.stop();
    await this.loadCatalogs();
    this.updateHost();
  }
  attach(mode: "battle" | "strategy", host: HTMLDivElement | null): void {
    if (this.stopped) return;
    this.hosts[mode] = host;
    this.updateHost();
  }
  /** Re-evaluate conditional hosts after controller/settings changes, including
   * when the ticker is stopped because neither host was active. */
  refresh(): void {
    if (this.stopped) return;
    this.dirty = true;
    this.updateHost();
  }
  private updateHost(): void {
    if (!this.initialized || this.stopped) return;
    const host =
      this.battle.playback.scene() !== null
        ? this.hosts.battle
        : this.battle.panel === "map"
          ? this.hosts.strategy
          : null;
    if (host === this.host) {
      if (host !== null) this.app.start();
      return;
    }
    this.host = host;
    this.dirty = true;
    if (host === null) {
      this.app.stop();
      this.app.canvas.remove();
    } else {
      host.appendChild(this.app.canvas);
      this.app.start();
    }
  }
  private async loadCatalogs(): Promise<void> {
    if (this.isStopped()) return;
    const d = assetDensity(this.scale(), window.devicePixelRatio);
    this.density = d;
    try {
      const [screen, units] = await Promise.all([
        this.images.catalog("screens", Math.min(2, d)),
        this.images.catalog("units", d),
      ]);
      if (!this.isStopped() && this.density === d) {
        this.screen = screen;
        this.units = units;
        this.dirty = true;
      }
    } catch (error) {
      if (!this.isStopped() && this.density === d) throw error;
    }
  }
  // Calls re-read lifecycle state after awaits; TypeScript otherwise keeps
  // the pre-await property narrowing even though dispose can run meanwhile.
  private isStopped(): boolean {
    return this.stopped;
  }
  private report(error: unknown): void {
    if (!this.stopped)
      this.error(error instanceof Error ? error : new Error(String(error)));
  }
  private draw(): void {
    if (this.stopped || this.host === null) return;
    const width = this.host.clientWidth,
      height = this.host.clientHeight,
      dpr = window.devicePixelRatio;
    if (width <= 0 || height <= 0) return;
    if (this.app.renderer.resolution !== dpr)
      this.app.renderer.resolution = dpr;
    if (this.app.screen.width !== width || this.app.screen.height !== height) {
      this.app.renderer.resize(width, height);
      this.dirty = true;
    }
    if (this.density !== assetDensity(this.scale(), dpr))
      void this.loadCatalogs().catch((e: unknown) => {
        this.report(e);
      });
    const scene = this.battle.playback.scene(),
      view = this.viewport(),
      key = `${width}/${height}/${dpr}/${scene?.tick ?? this.battle.version}/${scene?.reduced ?? false}/${JSON.stringify(view)}`;
    if (
      !this.dirty &&
      this.key === key &&
      this.lastBattle === (scene?.battle ?? null)
    )
      return;
    this.key = key;
    this.lastBattle = scene?.battle ?? null;
    this.dirty = false;
    this.used.clear();
    this.order.clear();
    this.order.set(this.emblems, 1);
    this.images.cache.begin();
    this.plates.clear();
    this.ground.clear();
    this.flag.clear();
    this.flashes.clear();
    if (scene !== null) {
      const l = battleLayout({ x: 0, y: 0, width, height }, Infinity, dpr),
        fighters = [scene.battle.attacker, scene.battle.defender] as const;
      for (const side of [0, 1] as const) {
        const f = fighters[side],
          r = l.sides[side],
          state = scene.state.sides[side],
          k = l.scale,
          meta = this.families[f.unit.id];
        this.plate(r.half, 0x414136);
        this.plate(r.photoWell, 0x181c12);
        this.plate(r.name, 0x453828);
        const hx = r.hex.x + r.hex.width / 2,
          hy = r.hex.y + r.hex.height / 2;
        this.plates
          .poly(
            [-31, -0, -15.5, -29, 15.5, -29, 31, 0, 15.5, 29, -15.5, 29].map(
              (v, i) => (i % 2 === 0 ? hx : hy) + v * k,
            ),
          )
          .fill(0x3d3021)
          .stroke({ color: 0x92927e, width: 1.5 * k });
        if (meta !== undefined) {
          const photo = this.screen.get(`portrait/${meta.family}`);
          if (photo)
            this.picture(
              `photo/${side}`,
              cropped(photo, r.photo.width / r.photo.height),
              r.photo,
              this.pictures,
            );
          const image = this.units.get(
            spriteID(meta.family, f.unit.owner, side === 0 ? "s" : "n", false),
          );
          if (image !== undefined) {
            const ship = f.unit.domain === play.Naval,
              ease = (v: number) => 1 - (1 - v) * (1 - v);
            if (ship && state.shown > 0) {
              const lost = state.lost >= state.shown,
                size = 64 * k * (lost ? 1 : 1.1 - 0.1 * ease(state.rise));
              this.figure(
                `ship/${side}`,
                image,
                { x: hx, y: hy + (lost ? 5 * k : 0) },
                size,
                lost ? 0.16 : 0,
                lost ? 0.32 : 0.95,
                lost ? 1 : ease(state.rise),
              );
            } else if (!ship) {
              for (
                let i = 0;
                i < Math.min(6, Math.max(f.unit.full, state.shown));
                i++
              ) {
                const p = figurePoint(r, i, false);
                this.ground
                  .ellipse(p.x, p.y + 7 * k, 7 * k, 2.6 * k)
                  .fill({ color: 0, alpha: 0.31 });
              }
              for (const i of [0, 5, 1, 4, 2, 3])
                if (i < state.shown) {
                  const p = figurePoint(r, i, false),
                    lost = i >= state.shown - state.lost,
                    e = i === state.shown - 1 ? ease(state.rise) : 1;
                  if (lost)
                    this.ground
                      .ellipse(p.x, p.y + 4 * k, 10 * k, 4.2 * k)
                      .fill({ color: 0x080502, alpha: 0.7 });
                  this.figure(
                    `figure/${side}/${i}`,
                    image,
                    { x: p.x, y: p.y + (lost ? 3 * k : -(1 - e) * 6 * k) },
                    23.75 * k * (lost ? 0.92 : 1.2 - 0.2 * e),
                    lost ? (i % 2 ? -0.45 : 0.45) : 0,
                    lost ? 0.32 : 0.95,
                    lost ? 1 : e,
                  );
                }
            }
          }
        }
        this.drawFlag(r.flag, f.unit.owner);
        const exp = scene.state.gained[side] ? f.expAfter : f.expBefore,
          badge = this.screen.get(`rank/${rank[Math.min(6, exp)] ?? "10"}`);
        if (exp > 0 && badge)
          this.picture(`badge/${side}`, badge, r.badge, this.emblems);
        this.plate(r.count, 0x18120c);
        this.label(
          `count/${side}`,
          String(state.count),
          r.count.x + r.count.width / 2,
          r.count.y + r.count.height / 2,
          17 * k,
          state.hit > 0 ? 0xf07040 : 0xf0b858,
        );
        this.label(
          `name/${side}`,
          f.unit.name,
          r.name.x + r.name.width / 2,
          r.name.y + r.name.height / 2,
          Math.max(8, 6 * k),
          0xfbe3c7,
          r.name.width - 4 * k,
        );
        if (scene.state.glow[side] > 0)
          this.label(
            `gain/${side}`,
            "+1",
            r.badge.x + r.badge.width + 6 * k,
            r.badge.y + 5 * k,
            8 * k,
            0xfff0c8,
          );
      }
      for (const [i, flash] of scene.state.flash.entries()) {
        const side = flash.side === 0 ? 0 : 1,
          r = l.sides[side],
          f = fighters[side],
          p = figurePoint(r, flash.figure, f.unit.domain === play.Naval),
          k = l.scale;
        if (flash.kind === 0) {
          this.flashes
            .circle(p.x, p.y, 4 * k * (1 - flash.progress))
            .fill({ color: 0xffe8ac, alpha: 0.9 });
        } else if (flash.kind === 1) {
          const image = this.screen.get(
            `effect/explosion-${String(Math.min(8, Math.floor(flash.progress * 9))).padStart(2, "0")}`,
          );
          if (image) this.figure(`blast/${i}`, image, p, 30 * k, 0, 1, 1);
        } else
          this.flashes
            .ellipse(p.x, p.y, 8 * k, 4 * k)
            .fill({ color: 0xb39a76, alpha: 0.5 * (1 - flash.progress) });
      }
    } else if (this.battle.panel === "map") this.drawStrategy(width, height);
    for (const [key, sprite] of this.sprites)
      if (!this.used.has(key)) {
        sprite.destroy();
        this.sprites.delete(key);
      }
    for (const [key, text] of this.texts)
      if (!this.used.has(key)) {
        text.destroy();
        this.texts.delete(key);
      }
    this.images.cache.end();
  }
  private plate(r: Rect, color: number): void {
    this.plates
      .rect(r.x, r.y, r.width, r.height)
      .fill(color)
      .stroke({ color: 0x777765, width: 1 });
  }
  private picture(
    key: string,
    image: MapImage,
    r: Rect,
    parent: Container,
  ): void {
    const sprite = this.sprite(key, image, parent);
    if (!sprite) return;
    sprite.position.set(r.x + r.width / 2, r.y + r.height / 2);
    sprite.width = r.width;
    sprite.height = r.height;
    sprite.alpha = 1;
    sprite.rotation = 0;
    sprite.tint = 0xffffff;
  }
  private figure(
    key: string,
    image: MapImage,
    p: Point,
    size: number,
    rotation: number,
    tone: number,
    alpha: number,
  ): void {
    const sprite = this.sprite(key, image, this.pictures);
    if (!sprite) return;
    const s = size / Math.max(sprite.texture.width, sprite.texture.height);
    sprite.position.set(p.x, p.y);
    sprite.scale.set(s);
    sprite.rotation = rotation;
    sprite.alpha = alpha;
    const c = Math.round(255 * tone);
    sprite.tint = (c << 16) | (c << 8) | c;
  }
  private sprite(
    key: string,
    image: MapImage,
    parent: Container,
  ): Sprite | null {
    const ready = this.images.get(image);
    if (ready === null) return null;
    this.used.add(key);
    let sprite = this.sprites.get(key);
    if (sprite === undefined) {
      sprite = new Sprite(ready.texture);
      sprite.anchor.set(0.5);
      this.sprites.set(key, sprite);
      parent.addChild(sprite);
    } else sprite.texture = ready.texture;
    if (sprite.parent !== parent) parent.addChild(sprite);
    const index = this.order.get(parent) ?? 0;
    if (parent.getChildIndex(sprite) !== index)
      parent.setChildIndex(sprite, index);
    this.order.set(parent, index + 1);
    return sprite;
  }
  private label(
    key: string,
    value: string,
    x: number,
    y: number,
    size: number,
    color: number,
    maxWidth = Infinity,
  ): void {
    this.used.add(key);
    let text = this.texts.get(key);
    if (text === undefined) {
      text = new Text({
        text: value,
        style: {
          fontFamily: "HL Sans",
          fontWeight: "600",
          fontSize: size,
          fill: color,
        },
      });
      text.anchor.set(0.5);
      this.labels.addChild(text);
      this.texts.set(key, text);
    }
    text.text = value;
    text.style.fontSize = size;
    text.style.fill = color;
    text.position.set(x, y);
    text.scale.set(1);
    if (text.width > maxWidth) text.scale.x = maxWidth / text.width;
  }
  private drawFlag(r: Rect, owner: number): void {
    this.flag.rect(r.x, r.y, r.width, r.height).fill(0x15140f);
    const colors =
      owner === 1
        ? [0x171714, 0xf6f2e4, 0xa92320]
        : owner === 2
          ? [0x173c72, 0xf6f2e4, 0xb42c28]
          : [0x747469];
    colors.forEach((color, i) => {
      const vertical = owner === 2,
        w = r.width / colors.length,
        h = r.height / colors.length;
      this.flag
        .rect(
          r.x + (vertical ? i * w : 0),
          r.y + (vertical ? 0 : i * h),
          vertical ? w : r.width,
          vertical ? r.height : h,
        )
        .fill(color);
    });
  }
  private drawStrategy(width: number, height: number): void {
    const lv = this.battle.session.level,
      grid = level.grid(lv),
      [ww, wh] = hexgrid.bounds(grid),
      scale = Math.min((width - 16) / ww, (height - 16) / wh),
      offset = { x: (width - ww * scale) / 2, y: (height - wh * scale) / 2 };
    this.mini = { scale, offset };
    const point = (x: number, y: number) => ({
        x: offset.x + x * scale,
        y: offset.y + y * scale,
      }),
      reworked = play.reworked(this.battle.session);
    if (!this.paletteRequested) {
      this.paletteRequested = true;
      void this.loadPalette().catch((error: unknown) => {
        this.report(error);
      });
    }
    for (const [line, row] of lv.lines.entries())
      for (const [col, original] of row.entries()) {
        const c = { line, col },
          h = reworked.get(level.coordKey(c)) ?? original,
          color = this.palette.get(h.terrain) ?? fallbackColor(h.terrain);
        this.plates
          .poly(
            hexgrid.corners(c).flatMap(([x, y]) => {
              const p = point(x, y);
              return [p.x, p.y];
            }),
          )
          .fill(color);
      }
    const view = this.battle.view();
    for (const [key, b] of view.buildings) {
      const [x, y] = hexgrid.center(play.coordFromKey(key)),
        p = point(x, y);
      this.ground.rect(p.x - 3, p.y - 3, 6, 6).fill(ownerColor(b.owner));
      if (b.hq) this.label(`hq/${key}`, "H", p.x, p.y, 9, 0x141608);
    }
    for (const u of view.units.values()) {
      const [x, y] = hexgrid.center(u.at),
        p = point(x, y);
      this.ground
        .circle(p.x, p.y, 1.5)
        .fill(
          u.side === this.battle.side && u.moved
            ? 0x787880
            : ownerColor(u.owner),
        );
    }
    const v = this.viewport(),
      p = point(v.x, v.y);
    this.ground
      .rect(p.x, p.y, v.width * scale, v.height * scale)
      .stroke({ color: 0xfff0c8, width: 1.5 });
  }
  private async loadPalette(): Promise<void> {
    if (this.isStopped()) return;
    const catalog = await this.images.catalog("terrain", 1);
    if (this.isStopped()) return;
    const terrains = new Set(
      this.battle.session.level.lines.flatMap((row) =>
        row.map((h) => h.terrain),
      ),
    );
    await Promise.all(
      [...terrains].map(async (terrain) => {
        const image =
          catalog.get(`tile/${terrain}`) ??
          catalog.get(`tile/${terrain.replace(/_w$/u, "")}`);
        if (!image) return;
        const bytes = await this.loader.store.get(image.full);
        if (this.isStopped()) return;
        const bitmap = await createImageBitmap(new Blob([bytes]));
        try {
          if (this.isStopped()) return;
          const canvas = document.createElement("canvas");
          canvas.width = 16;
          canvas.height = 16;
          const c = canvas.getContext("2d");
          if (c === null) return;
          c.drawImage(bitmap, 0, 0, 16, 16);
          const data = c.getImageData(0, 0, 16, 16).data;
          let a = 0,
            r = 0,
            g = 0,
            b = 0;
          for (let i = 0; i < data.length; i += 4) {
            const alpha = data[i + 3] ?? 0;
            a += alpha;
            r += (data[i] ?? 0) * alpha;
            g += (data[i + 1] ?? 0) * alpha;
            b += (data[i + 2] ?? 0) * alpha;
          }
          if (a > 0)
            this.palette.set(
              terrain,
              (Math.round(r / a) << 16) |
                (Math.round(g / a) << 8) |
                Math.round(b / a),
            );
        } finally {
          bitmap.close();
        }
      }),
    );
    if (!this.isStopped()) this.dirty = true;
  }
  dispose(): void {
    if (this.stopped) return;
    this.stopped = true;
    this.images.dispose();
    this.tone.destroy();
    if (this.initialized) {
      this.app.destroy(true, { children: true });
      this.initialized = false;
    }
    this.hosts.battle = null;
    this.hosts.strategy = null;
    this.host = null;
    this.sprites.clear();
    this.texts.clear();
    this.palette.clear();
  }
}
/** Native game's fallback for terrain absent from the selected art palette. */
function fallbackColor(id: string): number {
  let hash = 2166136261;
  for (const b of new TextEncoder().encode(id))
    hash = Math.imul(hash ^ b, 16777619) >>> 0;
  return (
    ((70 + (((hash >>> 16) & 255) % 120)) << 16) |
    ((90 + (((hash >>> 8) & 255) % 120)) << 8) |
    (60 + ((hash & 255) % 100))
  );
}
