import { Container, Graphics, Sprite, type Application } from "pixi.js";
import { hexgrid, level, play } from "@historyline/engine";
import { Camera, type Point, type Rect } from "./camera.ts";
import { assetDensity, type AssetLoader } from "./assets.ts";
import { type Pyramid, record } from "./assetManifest.ts";
import {
  backdropDraw,
  chooseMip,
  intersects,
  visibleHexes,
} from "./mapGeometry.ts";
import { terrainDraws, type BakedLayers, type MapScene } from "./mapScene.ts";
import { reworkScene } from "./dynamicScene.ts";
import { MapImages, type MapImage } from "./mapImages.ts";
import {
  unitFrames,
  spriteID,
  squad,
  toward,
  type SpriteFamily,
  type UnitFrame,
} from "./unitFrame.ts";
import { rimFor } from "./unitDress.ts";
import type { MapBattle } from "./mapBattle.ts";

const ownerColor = (owner: number): number =>
  owner === 1 ? 0x55c3f3 : owner === 2 ? 0xe7d365 : 0xc4c4b8;
interface Catalog {
  density: number;
  images: ReadonlyMap<string, MapImage>;
}
export interface MapStats {
  hexes: number;
  terrain: number;
  units: number;
  backdropTiles: number;
  mip: number;
  density: number;
  decodedBytes: number;
}

/** Pixi objects express a culled frame; game commands and hidden enemy state
 * never enter this renderer. Every pass keeps the native row/column ordering. */
export class MapRenderer {
  readonly camera: Camera;
  readonly images: MapImages;
  readonly root = new Container();
  private readonly backdropLayer = new Container();
  private readonly terrainLayer = new Container();
  private readonly ownership = new Graphics();
  private readonly unitLayer = new Container();
  private readonly markers = new Container();
  private readonly plates = new Graphics();
  private readonly walkingLayer = new Container();
  private readonly overlays = new Graphics();
  private readonly effects = new Container();
  private readonly sprites = new Map<string, Sprite>();
  private readonly used = new Set<string>();
  private readonly counts = new Map<Container, number>();
  private readonly catalogs = new Map<string, Catalog>();
  private readonly requested = new Set<string>();
  private readonly desired = new Map<string, number>();
  private glyphs: ReadonlyMap<string, MapImage> = new Map();
  backdrop = true;
  dirty = true;
  stats: MapStats = {
    hexes: 0,
    terrain: 0,
    units: 0,
    backdropTiles: 0,
    mip: -1,
    density: 1,
    decodedBytes: 0,
  };
  private lastVersion = -1;
  private terrainVersion = -1;
  private workKey = "";
  private scene: MapScene;
  private drawnLevel: level.Level;
  private changed = new Set<string>();

  constructor(
    readonly app: Application,
    readonly loader: AssetLoader,
    readonly baked: BakedLayers,
    readonly battle: MapBattle,
    readonly families: Readonly<Record<string, SpriteFamily>>,
    private readonly pyramid: Pyramid | undefined,
    private readonly report: (message: string) => void,
  ) {
    this.scene = baked.scene;
    this.drawnLevel = battle.session.level;
    const [width, height] = hexgrid.bounds(baked.scene.grid);
    this.camera = new Camera(
      { width, height },
      { width: app.screen.width, height: app.screen.height },
    );
    this.camera.viewport(
      app.screen.width,
      app.screen.height,
      app.renderer.resolution,
    );
    this.images = new MapImages(
      loader,
      () => {
        this.dirty = true;
      },
      (e) => {
        report(e.message);
      },
    );
    this.root.addChild(
      this.backdropLayer,
      this.terrainLayer,
      this.ownership,
      this.unitLayer,
      this.markers,
      this.walkingLayer,
      this.overlays,
      this.effects,
    );
    this.markers.addChild(this.plates);
    app.stage.addChild(this.root);
    const hq = [...battle.view().buildings].find(
      ([, b]) => b.hq && b.owner === battle.side + 1,
    );
    const first = [...battle.view().units.values()].find(
      (u) => u.side === battle.side,
    );
    const c = hq === undefined ? first?.at : play.coordFromKey(hq[0]);
    if (c !== undefined) this.camera.centerOn(...hexgrid.center(c));
  }
  async initialize(): Promise<void> {
    const d = assetDensity(this.camera.zoom, this.camera.scale);
    await Promise.all([
      this.load("terrain", d),
      this.load("units", d),
      this.load("screens", Math.min(2, d)),
      this.loadGlyphs(),
    ]);
  }
  resize(width: number, height: number, dpr: number): void {
    this.camera.viewport(width, height, dpr);
    this.dirty = true;
  }
  design(): void {
    this.backdrop = !this.backdrop;
    this.dirty = true;
  }
  render(): void {
    if (!this.dirty && this.lastVersion === this.battle.version) return;
    this.dirty = false;
    this.lastVersion = this.battle.version;
    const camera = this.camera,
      rect = camera.visible(),
      d = assetDensity(camera.zoom, camera.scale);
    for (const name of ["terrain", "units", "screens"])
      this.request(name, name === "screens" ? Math.min(2, d) : d);
    this.root.scale.set(camera.zoom);
    this.root.position.set(
      -camera.origin.x * camera.zoom,
      -camera.origin.y * camera.zoom,
    );
    this.images.cache.begin();
    // A worker may still be planning while the rules-side resolution is
    // pending. Fetch all explosion frames before their short on-map playback.
    if (this.battle.busy() && this.battle.playback.moving() === null) {
      const screens = this.catalogs.get("screens");
      for (let i = 0; i < 9; i++) {
        const image = screens?.images.get(
          `effect/explosion-${String(i).padStart(2, "0")}`,
        );
        if (image !== undefined) this.images.get(image);
      }
    }
    this.used.clear();
    this.counts.clear();
    this.counts.set(this.markers, 1);
    this.ownership.clear();
    this.plates.clear();
    this.overlays.clear();
    this.refreshTerrain();
    const cells = visibleHexes(this.scene.grid, rect, 180),
      view = this.battle.view();
    this.stats = {
      hexes: cells.length,
      terrain: 0,
      units: 0,
      backdropTiles: 0,
      mip: -1,
      density: d,
      decodedBytes: this.images.cache.bytes,
    };
    const painted = this.backdrop && this.pyramid !== undefined;
    if (painted) this.paintBackdrop(this.pyramid, rect);
    if (!painted || this.changed.size > 0) {
      const catalog = this.catalogs.get("terrain");
      if (catalog !== undefined)
        for (const draw of terrainDraws(
          this.scene,
          cells,
          this.baked.seasonBakes,
          this.changed,
        )) {
          if (
            painted &&
            (!this.changed.has(level.coordKey(draw.coord)) ||
              !["bleed", "background", "infrastructure"].includes(draw.pass))
          )
            continue;
          let image = catalog.images.get(draw.id);
          const mixed = this.baked.seasonBakes.layers
            .get(level.coordKey(draw.coord))
            ?.get(draw.pass === "bleed" ? "background" : draw.pass)
            ?.get(catalog.density);
          if (mixed?.id === draw.id) image = mixed;
          if (draw.pass === "background" || draw.pass === "bleed") {
            const baked = this.baked.backgrounds
              .get(level.coordKey(draw.coord))
              ?.get(catalog.density);
            if (baked?.id === draw.id) image = baked;
            if (image !== undefined && draw.pass === "bleed")
              image = { ...image, bleed: true };
          }
          if (image !== undefined) {
            this.sprite(draw.key, this.terrainLayer, image, draw.at);
            this.stats.terrain++;
          }
        }
    }
    const visible = new Set(cells.map(level.coordKey));
    for (const [key, b] of view.buildings)
      if (b.owner !== 0 && visible.has(key)) {
        const [x, y] = hexgrid.center(play.coordFromKey(key));
        this.ownership.rect(x - 20, y - 34, 40, 6).fill(ownerColor(b.owner));
      }
    const moving = this.battle.playback.moving();
    const frames = unitFrames(
      view,
      this.families,
      this.battle.playback.directions,
      moving === null ? new Set() : new Set([moving.unit.index]),
    );
    for (const frame of frames) {
      if (
        !visible.has(level.coordKey(frame.drawAt)) &&
        !visible.has(level.coordKey(frame.unit.at))
      )
        continue;
      this.unit(frame, this.unitLayer);
      this.stats.units++;
      if (camera.effective() >= 0.35 && frame.unit.size > 0)
        this.mark(frame.unit);
    }
    if (moving !== null) {
      const meta = this.families[moving.unit.id];
      if (meta !== undefined) {
        const family =
          moving.rear === null
            ? squad(meta.family, moving.unit.size)
            : (meta.long ?? meta.family);
        const at =
          moving.rear === null
            ? moving.at
            : {
                x: (moving.at.x + moving.rear.x) / 2,
                y: (moving.at.y + moving.rear.y) / 2,
              };
        const facing =
          moving.rear === null ? moving.facing : toward(moving.rear, moving.at);
        const [coord] = hexgrid.pixelToHex(
          this.baked.scene.grid,
          moving.at.x,
          moving.at.y,
        );
        this.unit(
          {
            key: "walking",
            id: spriteID(family, moving.unit.owner, facing, false),
            family,
            unit: moving.unit,
            at,
            drawAt: coord,
            facing,
          },
          this.walkingLayer,
        );
      }
    }
    this.drawOverlays(view);
    const screens = this.catalogs.get("screens");
    for (const effect of this.battle.playback.effects()) {
      const image = screens?.images.get(effect.frame);
      if (image !== undefined)
        this.sprite(effect.key, this.effects, image, effect.at, effect.alpha);
    }
    for (const [key, sprite] of this.sprites)
      if (!this.used.has(key)) {
        sprite.destroy();
        this.sprites.delete(key);
      }
    this.images.cache.end();
    this.stats.decodedBytes = this.images.cache.bytes;
  }
  private sprite(
    key: string,
    parent: Container,
    image: MapImage,
    at: Point,
    alpha = 1,
    tint = 0xffffff,
    size?: Rect,
  ): Sprite | null {
    const ready = this.images.get(image);
    if (ready === null) return null;
    this.used.add(key);
    let sprite = this.sprites.get(key);
    if (sprite === undefined) {
      sprite = new Sprite(ready.texture);
      this.sprites.set(key, sprite);
      parent.addChild(sprite);
    } else sprite.texture = ready.texture;
    const index = this.counts.get(parent) ?? 0;
    if (sprite.parent !== parent) parent.addChild(sprite);
    if (parent.getChildIndex(sprite) !== index)
      parent.setChildIndex(sprite, index);
    this.counts.set(parent, index + 1);
    sprite.position.set(
      size?.x ?? at.x - ready.anchor[0],
      size?.y ?? at.y - ready.anchor[1],
    );
    sprite.width = size?.width ?? ready.width / ready.density;
    sprite.height = size?.height ?? ready.height / ready.density;
    sprite.alpha = alpha;
    sprite.tint = tint;
    return sprite;
  }
  private unit(frame: UnitFrame, parent: Container): void {
    const image = this.catalogs.get("units")?.images.get(frame.id);
    if (image === undefined) return;
    const h = this.drawnLevel.lines[frame.drawAt.line]?.[frame.drawAt.col];
    this.sprite(
      frame.key,
      parent,
      {
        ...image,
        dressing: {
          rim: rimFor(h?.terrain ?? "", (h?.infra.ids.length ?? 0) > 0),
          flying: frame.unit.domain === play.Air,
        },
      },
      frame.at,
    );
  }
  private paintBackdrop(p: Pyramid, rect: Rect): void {
    const index = chooseMip(p, this.camera.world, this.camera.effective()),
      mip = p.mips[index];
    if (mip === undefined) return;
    this.stats.mip = index;
    for (const tile of mip.tiles) {
      const draw = backdropDraw(p, mip, tile, this.camera.world);
      if (!intersects(rect, draw.world)) continue;
      this.sprite(
        `backdrop/${index}/${tile.x}/${tile.y}`,
        this.backdropLayer,
        {
          id: "backdrop",
          width: tile.sourceWidth,
          height: tile.sourceHeight,
          density: 1,
          logicalAnchor: [0, 0],
          lite: tile.file,
          full: tile.file,
          frame: draw.frame,
        },
        { x: 0, y: 0 },
        1,
        0xffffff,
        draw.world,
      );
      this.stats.backdropTiles++;
    }
  }
  private mark(u: play.UnitView): void {
    const [x, y] = hexgrid.center(u.at),
      label = String(u.size),
      e = this.camera.effective(),
      k = Math.max(1, Math.round((11 * e) / 5)),
      pixel = k / e;
    const width = (label.length * 4 - 1) * pixel,
      height = 5 * pixel,
      pad = pixel,
      edge = Math.max(1, Math.floor(k / 2)) / e;
    const centre = this.camera.worldToScreen(x + 30, y + 24),
      origin = this.camera.screenToWorld(
        Math.round(centre.x) - Math.floor((width * e) / 2),
        Math.round(centre.y) - Math.floor((height * e) / 2),
      ),
      left = origin.x,
      top = origin.y;
    this.plates
      .rect(
        left - pad - edge,
        top - pad - edge,
        width + 2 * (pad + edge),
        height + 2 * (pad + edge),
      )
      .fill(ownerColor(u.owner));
    this.plates
      .rect(left - pad, top - pad, width + 2 * pad, height + 2 * pad)
      .fill({ color: 0x141608, alpha: 216 / 255 });
    for (let i = 0; i < label.length; i++) {
      const glyph = this.glyphs.get(label[i] ?? "");
      if (glyph !== undefined)
        this.sprite(
          `size/${u.index}/${i}`,
          this.markers,
          glyph,
          { x: 0, y: 0 },
          1,
          0xe7d365,
          { x: left + i * 4 * pixel, y: top, width: 3 * pixel, height },
        );
    }
    if (u.exp > 0) {
      const rank = ["", "01", "02", "03", "04", "05", "10"][Math.min(u.exp, 6)];
      const image = this.catalogs
        .get("screens")
        ?.images.get(`rank/${rank ?? "10"}`);
      if (image !== undefined) {
        const scale = 30 / Math.max(image.width, image.height);
        this.sprite(
          `badge/${u.index}`,
          this.markers,
          image,
          { x: 0, y: 0 },
          1,
          0xffffff,
          {
            x: x - (image.width * scale) / 2,
            y: y - 28 - (image.height * scale) / 2,
            width: image.width * scale,
            height: image.height * scale,
          },
        );
      }
    }
  }
  private drawOverlays(view: play.View): void {
    const g = this.overlays,
      width = 2 / this.camera.zoom;
    const hex = (at: hexgrid.Coord, color: number, fill: boolean) => {
      g.poly(hexgrid.corners(at).flat());
      if (fill) g.fill({ color, alpha: 0.17 });
      g.stroke({ color, width, alpha: 0.85 });
    };
    for (const key of this.battle.destinations())
      hex(play.coordFromKey(key), 0x55c3f3, true);
    for (const key of this.battle.targets())
      hex(play.coordFromKey(key), 0xff5a28, true);
    const selected = [...view.units.values()].find(
      (u) => u.index === this.battle.selected,
    );
    if (selected !== undefined) hex(selected.at, 0xfff0c8, false);
    if (this.battle.hover !== null) {
      const path = this.battle.route(this.battle.hover);
      for (let i = 0; i < path.length; i++) {
        const at = path[i];
        if (at === undefined) continue;
        const [x, y] = hexgrid.center(at);
        if (i === 0) g.moveTo(x, y);
        else g.lineTo(x, y);
      }
      g.stroke({ color: 0xfff0c8, width: 3 / this.camera.zoom, alpha: 0.9 });
      hex(this.battle.hover, 0xc4a773, false);
    }
    for (const order of this.battle.orders()) {
      const [x, y] = hexgrid.center(order.from),
        [tx, ty] = hexgrid.center(order.to);
      g.moveTo(x, y)
        .lineTo(tx, ty)
        .stroke({ color: 0xff5a28, width: 3 / this.camera.zoom, alpha: 0.9 });
      g.circle(tx, ty, 28).stroke({ color: 0xff5a28, width, alpha: 0.9 });
    }
  }
  private refreshTerrain(): void {
    if (
      this.battle.playback.busy() ||
      this.terrainVersion === play.version(this.battle.session)
    )
      return;
    const catalog = this.catalogs.get("terrain");
    if (catalog === undefined) return;
    this.terrainVersion = play.version(this.battle.session);
    const reworked = play.reworked(this.battle.session),
      key = JSON.stringify(
        [...reworked].sort((a, b) => a[0].localeCompare(b[0])),
      );
    if (key === this.workKey) return;
    this.workKey = key;
    if (reworked.size === 0) {
      this.scene = this.baked.scene;
      this.drawnLevel = this.battle.session.level;
      this.changed.clear();
      return;
    }
    const result = reworkScene(
      this.baked.scene,
      this.battle.session.level,
      reworked,
      (id) => catalog.images.has(id),
    );
    this.scene = result.scene;
    this.drawnLevel = result.level;
    this.changed = result.changed;
  }
  private request(name: string, d: number): void {
    const key = `${name}/${d}`;
    this.desired.set(name, d);
    if (this.catalogs.get(name)?.density === d || this.requested.has(key))
      return;
    this.requested.add(key);
    void this.load(name, d)
      .catch((e: unknown) => {
        this.report(e instanceof Error ? e.message : String(e));
      })
      .finally(() => {
        this.requested.delete(key);
      });
  }
  private async load(name: string, d: number): Promise<void> {
    const images = await this.images.catalog(name, d);
    if (this.desired.get(name) === undefined || this.desired.get(name) === d)
      this.catalogs.set(name, { density: d, images });
    this.dirty = true;
  }
  private async loadGlyphs(): Promise<void> {
    const font = this.loader.index.pixelFonts["5"];
    if (font === undefined) throw new Error("missing small pixel font");
    const raw = record(await this.loader.store.json(font.metrics)),
      glyphs = record(raw.glyphs);
    if (typeof raw.width !== "number" || typeof raw.height !== "number")
      throw new Error("invalid pixel font geometry");
    const images = new Map<string, MapImage>();
    for (const c of "0123456789") {
      const glyph = record(glyphs[c]);
      if (
        typeof glyph.x !== "number" ||
        typeof glyph.width !== "number" ||
        typeof glyph.height !== "number"
      )
        throw new Error("invalid pixel glyph geometry");
      images.set(c, {
        id: `glyph/${c}`,
        nearest: true,
        width: raw.width,
        height: raw.height,
        density: 1,
        logicalAnchor: [0, 0],
        lite: font.image,
        full: font.image,
        frame: { x: glyph.x, y: 0, width: glyph.width, height: glyph.height },
      });
    }
    this.glyphs = images;
  }
  dispose(): void {
    this.root.destroy({ children: true });
    this.sprites.clear();
    this.images.dispose();
  }
}
