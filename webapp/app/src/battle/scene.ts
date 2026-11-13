import {
  Application,
  Container,
  Graphics,
  Sprite,
  Text,
  Texture,
  ColorMatrixFilter,
} from "pixi.js";
import { hexgrid, level, play } from "@historyline/engine";
import { MapImages, type MapImage } from "../mapImages.ts";
import { assetDensity, unitDensity, type AssetLoader } from "../assets.ts";
import { spriteID, type SpriteFamily } from "../unitFrame.ts";
import type { MapBattle } from "../mapBattle.ts";
import type { Point, Rect } from "../camera.ts";
import { loadPixelFonts, type PixelFonts } from "../pixfont.ts";
import { Painter, RivetSheet, Tile, type Surfaces } from "../split/painter.ts";
import type { BattleState, SideState } from "./timeline.ts";
import {
  battleLayout,
  figurePoint,
  hexCentre,
  hexScale,
  scaledLayout,
  type BattleLayout,
  type SideRects,
} from "./layout.ts";
import { paintBattle } from "./plate.ts";

const ownerColor = (owner: number) =>
  owner === 1 ? 0x55c3f3 : owner === 2 ? 0xe7d365 : 0xc4c4b8;
/** The rank art of experience 1–6 (unitmarks.go badgeRanks): four stars, then the laurel ring. */
const rank = ["", "01", "02", "03", "04", "05", "10"];
/** The explosion's frames (effect/explosion-00 to -08). */
const explosionFrames = 9,
  explosionID = (n: number) => `effect/explosion-${String(n).padStart(2, "0")}`;

// The scene's colours (battle_draw.go, paint.go, well.go).
const cream = 0xfbe3c7,
  amber = 0xf0b858,
  hitRed = 0xff5a38,
  flashCore = 0xfff6dc,
  fireAmber = 0xf4a840,
  smokeBrown = 0x3a2c1e,
  dustTan = 0xb89c74,
  countFloor = 0x1e160e,
  plateEdge = 0x1c140c,
  plateBrown = 0x423426,
  brass = 0xa8803c,
  brassDark = 0x58401c,
  flagBlack = 0x101010,
  flagWhite = 0xeeeee6,
  flagRed = 0xc82020,
  flagBlue = 0x2038a0;

const sepiaRows = [
  [0.393, 0.769, 0.189],
  [0.349, 0.686, 0.168],
  [0.272, 0.534, 0.131],
];
/** toneFilter is the photos' sepia, each output channel scaled by gain (fireTone: pushed towards amber). */
function toneFilter(
  gain: [number, number, number] = [1, 1, 1],
): ColorMatrixFilter {
  const f = new ColorMatrixFilter(),
    m: number[] = [];
  sepiaRows.forEach((row, i) => {
    const g = gain[i] ?? 1;
    m.push(...row.map((v) => v * g), 0, 0);
  });
  m.push(0, 0, 0, 1, 0);
  f.matrix = m as typeof f.matrix;
  // At the canvas's own resolution: the photos and figures stay sharp at any device pixel ratio.
  f.resolution = "inherit";
  return f;
}
/** mix is the colour a + (b − a)·t of two 0xRRGGBB colours. */
function mix(a: number, b: number, t: number): number {
  t = Math.max(0, Math.min(1, t));
  let out = 0;
  for (const shift of [16, 8, 0]) {
    const x = (a >> shift) & 255,
      y = (b >> shift) & 255;
    out |= Math.round(x + (y - x) * t) << shift;
  }
  return out;
}
const ease = (t: number) => 1 - (1 - t) * (1 - t);

/** glyphs are the original design's large figures on a 5×7 grid (battle_draw.go). */
const glyphs: Readonly<Record<string, readonly string[]>> = {
  "0": [" ### ", "#   #", "#  ##", "# # #", "##  #", "#   #", " ### "],
  "1": ["  #  ", " ##  ", "  #  ", "  #  ", "  #  ", "  #  ", " ### "],
  "2": [" ### ", "#   #", "    #", "   # ", "  #  ", " #   ", "#####"],
  "3": ["#####", "   # ", "  #  ", "   # ", "    #", "#   #", " ### "],
  "4": ["   # ", "  ## ", " # # ", "#  # ", "#####", "   # ", "   # "],
  "5": ["#####", "#    ", "#### ", "    #", "    #", "#   #", " ### "],
  "6": ["  ## ", " #   ", "#    ", "#### ", "#   #", "#   #", " ### "],
  "+": ["     ", "  #  ", "  #  ", "#####", "  #  ", "  #  ", "     "],
};
/** glyphsWidth is the width of s in glyphs of cell size q. */
export function glyphsWidth(s: string, q: number): number {
  return s.length === 0 ? 0 : q * (6 * s.length - 1);
}

/** The muzzle flash's outline in the shot's frame: angle from its direction and reach (drawMuzzle). */
const tongues = [
  [0, 1.7],
  [0.2, 0.45],
  [0.42, 1.0],
  [0.7, 0.4],
  [1.15, 0.62],
  [1.6, 0.3],
  [2.4, 0.34],
  [Math.PI, 0.26],
] as const;
/** The drawn fireball's puffs: offset and size as shares of its radius (drawBlast). */
const blobs = [
  [0, 0, 0.62],
  [-0.42, 0.12, 0.42],
  [0.4, 0.18, 0.44],
  [-0.16, -0.38, 0.46],
  [0.24, -0.34, 0.4],
  [0.02, 0.36, 0.36],
  [-0.5, -0.24, 0.3],
  [0.52, -0.1, 0.32],
] as const;

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
/** The face of the plates' lettering: stencilled on the painted designs (the menus' face). */
const stencil = '"Stardos Stencil", "HL Display", serif';

/** Persistent auxiliary Pixi canvas: ready before short reduced-motion scenes,
 * attached to either dock host without recreating its graphics context. It
 * draws the battle scene as the desktop game's dospanel.DrawBattle: a painted
 * steel plate (plate.ts), each side's sepia photo with flag, rank badge, count
 * box and name plate, the figures on their hex and the shots. */
export class BattleScenes {
  private readonly app = new Application();
  /** The painted plate of the battle scene, one texel per device pixel. */
  private readonly plate = new Sprite();
  private readonly plates = new Graphics();
  private readonly ground = new Graphics();
  private readonly pictures = new Container();
  private readonly emblems = new Container();
  private readonly flag = new Graphics();
  private readonly medallions = new Graphics();
  private readonly counts = new Graphics();
  private readonly labels = new Container();
  private readonly glows = new Container();
  private readonly fire = new Container();
  private readonly flashes = new Graphics();
  private readonly tone = toneFilter();
  private readonly fireTone = toneFilter([1.08, 0.9, 0.66]);
  private readonly images: MapImages;
  private readonly sprites = new Map<string, Sprite>();
  private readonly texts = new Map<string, Text>();
  private readonly glowSprites = new Map<string, Sprite>();
  private readonly pixelNames = new Map<string, Texture>();
  private readonly used = new Set<string>();
  private readonly order = new Map<Container, number>();
  private screen: ReadonlyMap<string, MapImage> = new Map();
  private units: ReadonlyMap<string, MapImage> = new Map();
  private surfaces: Surfaces = { steel: null, rivets: null, ground: null };
  private surfacesFor: ReadonlyMap<string, MapImage> | null = null;
  private plateTexture: Texture | null = null;
  private plateKey = "";
  private glowTexture: Texture | null = null;
  private pixelFonts: PixelFonts | null = null;
  private pixelFontsRequested = false;
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
    /** The map design drawn now: the original's draws the plates' lettering in its pixel font. */
    private readonly design: () => string = () => "painted",
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
      this.app.destroy(
        { removeView: true, releaseGlobalResources: false },
        { children: true },
      );
      return;
    }
    this.initialized = true;
    this.pictures.filters = [this.tone];
    this.fire.filters = [this.fireTone];
    this.emblems.addChild(this.flag, this.medallions);
    this.app.stage.addChild(
      this.plate,
      this.plates,
      this.ground,
      this.pictures,
      this.emblems,
      this.counts,
      this.labels,
      this.glows,
      this.fire,
      this.flashes,
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
    // The lettering's face may still be loading; its texts are made again once it is there.
    void document.fonts
      .load(`700 32px ${stencil}`)
      .then(() => {
        if (this.isStopped()) return;
        for (const text of this.texts.values()) text.destroy();
        this.texts.clear();
        this.dirty = true;
      })
      .catch(() => undefined);
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
        this.images.catalog("units", unitDensity(d)),
      ]);
      if (!this.isStopped() && this.density === d) {
        this.screen = screen;
        this.units = units;
        this.dirty = true;
        if (this.surfacesFor !== screen) {
          this.surfacesFor = screen;
          void this.loadSurfaces(screen).catch((e: unknown) => {
            this.report(e);
          });
        }
      }
    } catch (error) {
      if (!this.isStopped() && this.density === d) throw error;
    }
  }
  /** loadSurfaces decodes the plate's painted steel, ground and rivets (dospanel.SetSurfaces). */
  private async loadSurfaces(
    screen: ReadonlyMap<string, MapImage>,
  ): Promise<void> {
    const decode = async (id: string): Promise<ImageData | null> => {
      const image = screen.get(id);
      if (image === undefined) return null;
      const bytes = await this.loader.store.get(image.full);
      if (this.isStopped()) return null;
      const bitmap = await createImageBitmap(new Blob([bytes]));
      try {
        const canvas = document.createElement("canvas");
        canvas.width = bitmap.width;
        canvas.height = bitmap.height;
        const c = canvas.getContext("2d", { willReadFrequently: true });
        if (c === null) return null;
        c.drawImage(bitmap, 0, 0);
        return c.getImageData(0, 0, canvas.width, canvas.height);
      } finally {
        bitmap.close();
      }
    };
    const [steel, ground, rivets] = await Promise.all([
      decode("surface/steel"),
      decode("surface/ground"),
      decode("surface/rivets"),
    ]);
    if (this.isStopped() || this.surfacesFor !== screen) return;
    this.surfaces = {
      steel: steel === null ? null : new Tile(steel),
      ground: ground === null ? null : new Tile(ground),
      rivets: rivets === null ? null : new RivetSheet(rivets),
    };
    this.plateKey = "";
    this.dirty = true;
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
      key = `${width}/${height}/${dpr}/${scene?.tick ?? this.battle.version}/${scene?.reduced ?? false}/${this.design()}/${JSON.stringify(view)}`;
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
    this.order.set(this.emblems, 2);
    this.images.cache.begin();
    for (const g of [
      this.plates,
      this.ground,
      this.flag,
      this.medallions,
      this.counts,
      this.flashes,
    ])
      g.clear();
    this.plate.visible = scene !== null;
    if (scene !== null)
      this.drawBattle(scene.battle, scene.state, width, height, dpr);
    else if (this.battle.panel === "map") this.drawStrategy(width, height);
    for (const [key, sprite] of this.sprites)
      if (!this.used.has(key)) {
        sprite.destroy();
        this.sprites.delete(key);
      }
    for (const [key, sprite] of this.glowSprites)
      if (!this.used.has(key)) {
        sprite.destroy();
        this.glowSprites.delete(key);
      }
    for (const [key, text] of this.texts)
      if (!this.used.has(key)) {
        text.destroy();
        this.texts.delete(key);
      }
    this.images.cache.end();
  }
  /**
   * drawBattle draws the scene: the plate painted in device pixels and the rest laid out on it in CSS pixels, so
   * every part keeps to the plate's pixels at any device pixel ratio.
   */
  private drawBattle(
    b: play.BattleInfo,
    st: BattleState,
    width: number,
    height: number,
    dpr: number,
  ): void {
    const dw = Math.max(1, Math.round(width * dpr)),
      dh = Math.max(1, Math.round(height * dpr)),
      device = battleLayout({ x: 0, y: 0, width: dw, height: dh }),
      l = scaledLayout(device, 1 / dpr),
      fighters = [b.attacker, b.defender] as const,
      pixel = this.design() === "original";
    this.paintPlate(device, dw, dh, dpr);
    if (pixel && !this.pixelFontsRequested) {
      this.pixelFontsRequested = true;
      void loadPixelFonts(this.loader)
        .then((fonts) => {
          if (this.isStopped()) return;
          this.pixelFonts = fonts;
          this.dirty = true;
        })
        .catch((e: unknown) => {
          this.report(e);
        });
    }
    for (const side of [0, 1] as const) {
      const f = fighters[side],
        r = l.sides[side],
        state = st.sides[side],
        meta = this.families[f.unit.id];
      if (meta !== undefined) {
        const photo = this.screen.get(`portrait/${meta.family}`);
        if (photo)
          this.picture(
            `photo/${side}`,
            cropped(photo, r.photo.width / r.photo.height),
            r.photo,
            this.pictures,
          );
      }
      this.drawFlag(r.flag, f.unit.owner);
      this.drawBadge(side, r, f, st, l.scale, dpr);
      this.drawCount(side, r, state, l.scale, dpr, pixel);
      this.drawName(side, r, f.unit.name.toUpperCase(), l.scale, dpr, pixel);
      if (meta !== undefined)
        this.drawFigures(
          side,
          r,
          state,
          f,
          this.units.get(
            spriteID(meta.family, f.unit.owner, side === 0 ? "s" : "n", false),
          ),
        );
    }
    // Every frame of the explosion decodes as the scene opens, so none is missing when its shot lands.
    for (let n = 0; n < explosionFrames; n++) {
      const frame = this.screen.get(explosionID(n));
      if (frame !== undefined) this.images.get(frame);
    }
    for (const [i, flash] of st.flash.entries()) this.drawFlash(i, flash, l, b);
  }
  /** paintPlate paints the plate for device layout l in a canvas of w × h device pixels, once per size. */
  private paintPlate(l: BattleLayout, w: number, h: number, dpr: number): void {
    const s = this.surfaces,
      key = `${w}x${h}/${s.steel !== null}/${(s.ground ?? null) !== null}/${s.rivets !== null}`;
    if (key !== this.plateKey) {
      const canvas = document.createElement("canvas");
      canvas.width = w;
      canvas.height = h;
      const c = canvas.getContext("2d");
      if (c === null) return;
      const img = c.createImageData(w, h);
      paintBattle(new Painter(img.data, w, h, l.scale, s), l);
      c.putImageData(img, 0, 0);
      this.plateTexture?.destroy(true);
      this.plateTexture = Texture.from(canvas);
      this.plate.texture = this.plateTexture;
      this.plateKey = key;
    }
    this.plate.position.set(0, 0);
    this.plate.scale.set(1 / dpr);
  }
  /** drawFlag draws the nation's flag in the proportions of the well's, the Central Powers' bands across. */
  private drawFlag(r: Rect, owner: number): void {
    if (owner !== 1 && owner !== 2) return;
    const w = r.width,
      h = (r.width * 2) / 3,
      y = r.y + (r.height - h) / 2,
      bands =
        owner === 1
          ? [flagBlack, flagWhite, flagRed]
          : [flagBlue, flagWhite, flagRed];
    bands.forEach((color, i) => {
      if (owner === 1)
        this.flag.rect(r.x, y + (i * h) / 3, w, h / 3).fill(color);
      else this.flag.rect(r.x + (i * w) / 3, y, w / 3, h).fill(color);
    });
  }
  /** drawBadge draws the rank badge on a brass-rimmed medallion; a gain switches to the new one with a glow and a "+1". */
  private drawBadge(
    side: 0 | 1,
    r: SideRects,
    f: play.Fighter,
    st: BattleState,
    k: number,
    dpr: number,
  ): void {
    const exp = st.gained[side] ? f.expAfter : f.expBefore,
      gl = st.glow[side],
      art =
        exp > 0
          ? (this.screen.get(`rank/${rank[Math.min(6, exp)] ?? "10"}`) ?? null)
          : null;
    if (art === null && gl <= 0) return;
    const cx = r.badge.x + r.badge.width / 2,
      cy = r.badge.y + r.badge.height / 2,
      rad = r.badge.width / 2,
      px = 1 / dpr;
    this.medallions
      .circle(cx, cy, rad)
      .fill(plateEdge)
      .circle(cx, cy, rad - Math.max(px, k * 0.8))
      .fill(brass)
      .circle(cx, cy, rad - Math.max(2 * px, k * 1.8))
      .fill(countFloor);
    if (gl > 0)
      this.glow(
        `badge/${side}`,
        cx,
        cy,
        rad * (1.6 + 0.8 * gl),
        fireAmber,
        0.9 * gl,
      );
    if (art !== null) {
      // The badges are drawn with a wide margin: twice the medallion fits the mark inside it.
      const sprite = this.sprite(`badge/${side}`, art, this.emblems);
      if (sprite) {
        const size = 2 * rad * 1.9 * (1 + 0.25 * gl);
        sprite.position.set(cx, cy);
        sprite.scale.set(
          size / Math.max(sprite.texture.width, sprite.texture.height),
        );
        sprite.rotation = 0;
        sprite.alpha = 1;
        sprite.tint = 0xffffff;
      }
    }
    if (gl > 0) {
      const size = 9 * k,
        x =
          side === 0
            ? r.badge.x + r.badge.width + k + size * 0.6
            : r.badge.x - k - size * 0.6;
      this.label(
        `gain/${side}`,
        "+1",
        x,
        cy - (1 - gl) * 4 * k,
        size,
        mix(amber, flashCore, gl),
      );
    }
  }
  /**
   * drawCount draws the count box on the photo: the group's size large and, once men fell, the size before the
   * battle small and struck out; the figures flash red as a man falls.
   */
  private drawCount(
    side: 0 | 1,
    r: SideRects,
    ss: SideState,
    k: number,
    dpr: number,
    pixel: boolean,
  ): void {
    const b = r.count,
      bw = Math.max(1, Math.round(k * dpr)) / dpr;
    this.counts
      .rect(b.x, b.y, b.width, b.height)
      .fill(plateEdge)
      .rect(b.x + bw, b.y + bw, b.width - 2 * bw, b.height - 2 * bw)
      .fill(brassDark)
      .rect(b.x + 2 * bw, b.y + 2 * bw, b.width - 4 * bw, b.height - 4 * bw)
      .fill(countFloor);
    const big = String(ss.count),
      col = mix(amber, hitRed, ss.hit),
      old = String(ss.count + ss.lost),
      oldCol = mix(cream, plateBrown, 0.25);
    if (pixel) {
      const q = Math.max(1, Math.floor((b.height * dpr * 0.62) / 7)) / dpr,
        dy = b.y + (b.height - 7 * q) / 2;
      if (ss.lost === 0) {
        this.glyphs(
          big,
          b.x + (b.width - glyphsWidth(big, q)) / 2,
          dy,
          q,
          col,
          dpr,
        );
        return;
      }
      this.glyphs(big, b.x + 3.5 * k, dy, q, col, dpr);
      const sq = Math.max(1, Math.round(q * dpr * 0.55)) / dpr,
        ox = b.x + b.width - 3.5 * k - glyphsWidth(old, sq),
        oy = b.y + b.height - 3.5 * k - 7 * sq;
      this.glyphs(old, ox, oy, sq, oldCol, dpr);
      this.strike(
        ox - sq,
        oy + 8 * sq,
        ox + glyphsWidth(old, sq) + sq,
        oy - sq,
        Math.max(1 / dpr, sq * 0.9),
      );
      return;
    }
    // Painted designs: stencilled figures, the cap height about three fifths of the box.
    const size = b.height * 0.9;
    this.label(
      `count/${side}`,
      big,
      ss.lost === 0 ? b.x + b.width / 2 : b.x + 3 * k + size * 0.28,
      b.y + b.height / 2 + size * 0.04,
      size,
      col,
      Infinity,
      ss.lost === 0 ? 0.5 : 0.5,
    );
    if (ss.lost === 0) return;
    const small = size * 0.48,
      ox = b.x + b.width - 3 * k - small * 0.3,
      oy = b.y + b.height - 3 * k - small * 0.32;
    this.label(`count-old/${side}`, old, ox, oy, small, oldCol);
    this.strike(
      ox - small * 0.42,
      oy + small * 0.36,
      ox + small * 0.42,
      oy - small * 0.36,
      Math.max(1 / dpr, small * 0.11),
    );
  }
  private strike(
    x0: number,
    y0: number,
    x1: number,
    y1: number,
    w: number,
  ): void {
    this.counts
      .moveTo(x0, y0)
      .lineTo(x1, y1)
      .stroke({ color: hitRed, width: w, cap: "round" });
  }
  /** glyphs draws s in the original design's 5×7 figures of cell q from its top left, with a shadow below right. */
  private glyphs(
    s: string,
    x: number,
    y: number,
    q: number,
    color: number,
    dpr: number,
  ): void {
    const sh = Math.max(1, Math.round((q * dpr) / 2)) / dpr;
    for (const [pass, c] of [
      [0, 0x000000],
      [1, color],
    ] as const) {
      const off = pass === 0 ? sh : 0;
      Array.from(s, (ch, n) => {
        const g = glyphs[ch];
        if (g === undefined) return;
        g.forEach((line, row) => {
          for (let i = 0; i < line.length; i++)
            if (line[i] === "#")
              this.counts
                .rect(x + off + (6 * n + i) * q, y + off + row * q, q, q)
                .fill(c);
        });
      });
    }
  }
  /** drawName letters the name plate: stencilled on the painted designs, the original's pixel font on its own. */
  private drawName(
    side: 0 | 1,
    r: SideRects,
    name: string,
    k: number,
    dpr: number,
    pixel: boolean,
  ): void {
    const n = r.name,
      room = n.width - 2 * (3.5 * k + 2.6 * k),
      cx = n.x + n.width / 2,
      cy = n.y + n.height / 2;
    if (!pixel) {
      this.label(
        `name/${side}`,
        name,
        cx,
        cy + n.height * 0.04,
        n.height * 0.72,
        cream,
        room,
        0.5,
        true,
      );
      return;
    }
    const face = this.pixelFonts?.small;
    if (face === undefined) return;
    // The largest whole scale of the font whose name still fits.
    let q = Math.max(1, Math.floor((n.height * dpr * 0.6) / face.height));
    while (q > 1 && face.measure(name, q).width > room * dpr) q--;
    const key = `${name}/${q}`;
    let texture = this.pixelNames.get(key);
    if (texture === undefined) {
      const text = face.fit(name, q, Math.round(room * dpr)),
        m = face.measure(text, q),
        canvas = document.createElement("canvas");
      canvas.width = Math.max(1, m.width + q);
      canvas.height = Math.max(1, m.height + q);
      const c = canvas.getContext("2d");
      if (c === null) return;
      face.drawShadow(c, text, 0, 0, q, "#fbe3c7", "#000000");
      texture = Texture.from(canvas);
      texture.source.scaleMode = "nearest";
      if (this.pixelNames.size > 16) {
        for (const t of this.pixelNames.values()) t.destroy(true);
        this.pixelNames.clear();
      }
      this.pixelNames.set(key, texture);
    }
    const sprite = this.plainSprite(`pixel-name/${side}`, texture, this.labels);
    sprite.anchor.set(0, 0);
    sprite.scale.set(1 / dpr);
    sprite.position.set(
      Math.round((cx - texture.width / dpr / 2) * dpr) / dpr,
      Math.round((cy - (face.height * q) / dpr / 2) * dpr) / dpr,
    );
  }
  /** drawFigures draws the ring's places as ground shadows and the men on them; a ship as one picture. */
  private drawFigures(
    side: 0 | 1,
    r: SideRects,
    state: SideState,
    f: play.Fighter,
    image: MapImage | undefined,
  ): void {
    const hk = hexScale(r),
      ship = f.unit.domain === play.Naval;
    if (ship) {
      if (state.shown === 0 || image === undefined) return;
      const c = hexCentre(r),
        lost = state.lost >= state.shown,
        e = ease(state.rise);
      if (lost)
        this.ground
          .ellipse(c.x, c.y + 8 * hk, 26 * hk, 7 * hk)
          .fill({ color: 0x080502, alpha: 0.7 });
      this.figure(
        `ship/${side}`,
        image,
        { x: c.x, y: c.y + (lost ? 5 * hk : 0) },
        64 * hk * (lost ? 1 : 1.1 - 0.1 * e),
        lost ? 0.16 : 0,
        lost ? 0.32 : 0.95,
        lost ? 1 : e,
      );
      return;
    }
    // The places a full group would fill: the men missing from it leave empty foxholes.
    for (
      let i = 0;
      i < Math.min(6, Math.max(f.unit.full, state.shown, 0));
      i++
    ) {
      const p = figurePoint(r, i, false);
      this.ground
        .ellipse(p.x, p.y + 7 * hk, 7 * hk, 2.6 * hk)
        .fill({ color: 0, alpha: 0x50 / 255 });
    }
    if (image === undefined) return;
    const size = 2 * 9.5 * hk * 1.25;
    // From the top of the ring down, so the lower ones stand in front.
    for (const i of [0, 5, 1, 4, 2, 3]) {
      if (i >= state.shown) continue;
      const p = figurePoint(r, i, false),
        lost = i >= state.shown - state.lost,
        e = i === state.shown - 1 ? ease(state.rise) : 1;
      if (lost) {
        // A dark wreck tilted on a scorch mark.
        this.ground
          .ellipse(p.x, p.y + 4 * hk, 10 * hk, 4.2 * hk)
          .fill({ color: 0x080502, alpha: 0xb0 / 255 })
          .ellipse(p.x - 2 * hk, p.y + 3 * hk, 5 * hk, 2 * hk)
          .fill({ color: 0, alpha: 0xc0 / 255 });
        this.figure(
          `figure/${side}/${i}`,
          image,
          { x: p.x, y: p.y + 3 * hk },
          size * 0.92,
          i % 2 ? -0.45 : 0.45,
          0.32,
          1,
        );
      } else
        this.figure(
          `figure/${side}/${i}`,
          image,
          { x: p.x, y: p.y - (1 - e) * 6 * hk },
          size * (1.2 - 0.2 * e),
          0,
          0.95,
          e,
        );
    }
  }
  /** drawFlash draws one effect: a muzzle flash towards the other hex, a blast, or a puff of dust. */
  private drawFlash(
    i: number,
    flash: BattleState["flash"][number],
    l: BattleLayout,
    b: play.BattleInfo,
  ): void {
    const side = flash.side === 0 ? 0 : 1,
      r = l.sides[side],
      fighters = [b.attacker, b.defender] as const,
      ship = fighters[side].unit.domain === play.Naval,
      k = hexScale(r),
      p = flash.progress,
      at = flash.figure < 0 ? hexCentre(r) : figurePoint(r, flash.figure, ship);
    if (flash.kind === 0) {
      const t = hexCentre(l.sides[side === 0 ? 1 : 0]),
        d = Math.max(1, Math.hypot(t.x - at.x, t.y - at.y)),
        dx = (t.x - at.x) / d,
        dy = (t.y - at.y) / d,
        env = p <= 0.2 ? p / 0.2 : (1 - p) / 0.8;
      this.muzzle(
        `muzzle/${i}`,
        at.x + dx * 7 * k,
        at.y + dy * 7 * k,
        dx,
        dy,
        9 * k * env,
        env,
      );
    } else if (flash.kind === 1)
      this.blast(`blast/${i}`, at.x, at.y - 2 * k, (ship ? 50 : 40) * k, p);
    else {
      const s = flash.figure % 2 === 0 ? -1 : 1,
        x = at.x + s * 11 * k,
        y = at.y + 5 * k,
        size = 9 * k,
        a = (200 / 255) * (1 - p),
        g = ease(p);
      for (const [n, [ox, oy]] of (
        [
          [0, 0],
          [-0.45, -0.25],
          [0.4, -0.35],
        ] as const
      ).entries())
        this.flashes
          .circle(
            x + ox * size * g,
            y + oy * size * g - size * 0.4 * g,
            size * (0.3 + 0.35 * g) * (1 - 0.15 * n),
          )
          .fill({ color: dustTan, alpha: a });
    }
  }
  /** muzzle draws a star of fire at (x, y) of radius r, long towards (dx, dy), at strength a. */
  private muzzle(
    key: string,
    x: number,
    y: number,
    dx: number,
    dy: number,
    r: number,
    a: number,
  ): void {
    if (r <= 0 || a <= 0) return;
    a = Math.min(1, a);
    this.glow(key, x, y, r * 2.6, fireAmber, 0.9 * a);
    const pts: number[] = [],
      pt = (ang: number, reach: number) => {
        const u = Math.cos(ang) * reach * r,
          v = Math.sin(ang) * reach * r;
        pts.push(x + u * dx - v * dy, y + u * dy + v * dx);
      };
    for (const [ang, reach] of tongues) pt(ang, reach);
    for (let n = tongues.length - 2; n > 0; n--) {
      const t = tongues[n];
      if (t !== undefined) pt(-t[0], t[1]);
    }
    this.flashes
      .poly(pts)
      .fill({ color: fireAmber, alpha: (240 / 255) * a })
      .circle(x + dx * r * 0.15, y + dy * r * 0.15, r * 0.36)
      .fill({ color: flashCore, alpha: a });
  }
  /** blast draws an explosion of size at (x, y) at progress p: the art's frames in an amber sepia, or a fireball. */
  private blast(
    key: string,
    x: number,
    y: number,
    size: number,
    p: number,
  ): void {
    const fade = 1 - p;
    this.glow(
      key,
      x,
      y,
      size * 0.9 * (0.6 + 0.4 * fade),
      fireAmber,
      0.5 * fade * fade,
    );
    const frame = this.screen.get(
      explosionID(
        Math.min(explosionFrames - 1, Math.floor(p * explosionFrames)),
      ),
    );
    if (frame !== undefined) {
      // The art's frame, or only the glow while it is still decoding.
      const sprite = this.sprite(key, frame, this.fire);
      if (sprite === null) return;
      sprite.position.set(x, y);
      sprite.scale.set(
        (size * 1.25 * (0.8 + 0.2 * ease(p))) /
          Math.max(sprite.texture.width, sprite.texture.height),
      );
      sprite.rotation = 0;
      sprite.alpha = Math.min(1, fade / 0.2);
      sprite.tint = 0xffffff;
      return;
    }
    const grow = ease(Math.min(1, p / 0.55)),
      rad = size * 0.5 * (0.3 + 0.7 * grow),
      rise = rad * 0.35 * p,
      smoke = (230 / 255) * Math.min(1, fade * 1.5),
      heat = Math.max(0, 1 - p / 0.75);
    for (const [bx, by, bs] of blobs)
      this.flashes
        .circle(x + bx * rad, y + by * rad - rise, bs * rad * 1.15)
        .fill({ color: smokeBrown, alpha: smoke });
    if (heat > 0)
      for (const [bx, by, bs] of blobs)
        this.flashes
          .circle(
            x + bx * rad * 0.9,
            y + by * rad * 0.9 - rise,
            bs * rad * heat,
          )
          .fill({ color: fireAmber, alpha: heat });
    if (heat > 0)
      this.flashes
        .circle(x, y - rise, rad * 0.42 * heat * heat)
        .fill(flashCore);
    for (let n = 0; n < 7; n++) {
      const a = (n * 2 * Math.PI) / 7 + 0.3,
        d = rad * (0.7 + 0.9 * grow);
      this.flashes
        .circle(
          x + d * Math.cos(a),
          y + d * Math.sin(a) - rise * 2,
          size * 0.025 * fade + 0.5,
        )
        .fill({ color: smokeBrown, alpha: fade });
    }
  }
  /** glow adds a soft round light of colour and radius r at (x, y), at strength a. */
  private glow(
    key: string,
    x: number,
    y: number,
    r: number,
    color: number,
    a: number,
  ): void {
    if (r <= 0 || a <= 0) return;
    if (this.glowTexture === null) {
      const n = 64,
        canvas = document.createElement("canvas");
      canvas.width = n;
      canvas.height = n;
      const c = canvas.getContext("2d");
      if (c === null) return;
      const img = c.createImageData(n, n);
      for (let j = 0; j < n; j++)
        for (let i = 0; i < n; i++) {
          const d = Math.hypot(i + 0.5 - n / 2, j + 0.5 - n / 2) / (n / 2),
            v = Math.round(255 * Math.max(0, 1 - d) ** 2),
            o = 4 * (j * n + i);
          img.data[o] = img.data[o + 1] = img.data[o + 2] = 255;
          img.data[o + 3] = v;
        }
      c.putImageData(img, 0, 0);
      this.glowTexture = Texture.from(canvas);
    }
    key = `glow/${key}`;
    this.used.add(key);
    let sprite = this.glowSprites.get(key);
    if (sprite === undefined) {
      sprite = new Sprite(this.glowTexture);
      sprite.anchor.set(0.5);
      sprite.blendMode = "add";
      this.glows.addChild(sprite);
      this.glowSprites.set(key, sprite);
    }
    sprite.position.set(x, y);
    sprite.width = sprite.height = 2 * r;
    sprite.tint = color;
    sprite.alpha = Math.min(1, a);
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
    return this.plainSprite(key, ready.texture, parent);
  }
  /** plainSprite is the sprite under key showing texture, made once and kept in its parent's draw order. */
  private plainSprite(
    key: string,
    texture: Texture,
    parent: Container,
  ): Sprite {
    this.used.add(key);
    let sprite = this.sprites.get(key);
    if (sprite === undefined) {
      sprite = new Sprite(texture);
      sprite.anchor.set(0.5);
      this.sprites.set(key, sprite);
      parent.addChild(sprite);
    } else sprite.texture = texture;
    if (sprite.parent !== parent) parent.addChild(sprite);
    const index = Math.min(
      this.order.get(parent) ?? 0,
      parent.children.length - 1,
    );
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
    anchor = 0.5,
    shadow = false,
  ): void {
    this.used.add(key);
    const face = key.startsWith("hq/") ? "HL Sans" : stencil;
    let text = this.texts.get(key);
    if (text === undefined) {
      text = new Text({
        text: value,
        style: {
          fontFamily: face,
          fontWeight: face === stencil ? "700" : "600",
          fontSize: size,
          fill: color,
        },
      });
      this.labels.addChild(text);
      this.texts.set(key, text);
    }
    text.anchor.set(anchor, 0.5);
    text.text = value;
    text.style.fontSize = size;
    text.style.fill = color;
    text.style.dropShadow = shadow
      ? {
          color: 0x000000,
          alpha: 0.85,
          blur: 0,
          angle: Math.PI / 4,
          distance: Math.max(1, size * 0.08),
        }
      : false;
    text.position.set(x, y);
    text.scale.set(1);
    if (text.width > maxWidth) text.scale.x = maxWidth / text.width;
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
    this.fireTone.destroy();
    if (this.initialized) {
      // The map renderer survives this canvas; its active batches share Pixi's
      // global pools, so only document teardown may release those resources.
      this.app.destroy(
        { removeView: true, releaseGlobalResources: false },
        { children: true },
      );
      this.initialized = false;
    }
    this.plateTexture?.destroy(true);
    this.plateTexture = null;
    this.glowTexture?.destroy(true);
    this.glowTexture = null;
    for (const t of this.pixelNames.values()) t.destroy(true);
    this.pixelNames.clear();
    this.hosts.battle = null;
    this.hosts.strategy = null;
    this.host = null;
    this.sprites.clear();
    this.glowSprites.clear();
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
