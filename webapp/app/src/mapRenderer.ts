import {
  Container,
  Graphics,
  Sprite,
  type Application,
  type Texture,
} from "pixi.js";
import { bind, hexgrid, level, play } from "@historyline/engine";
import { Camera, type Point, type Rect } from "./camera.ts";
import { assetDensity, type AssetLoader } from "./assets.ts";
import { type Pyramid, type Tile, record } from "./assetManifest.ts";
import {
  backdropDraw,
  backdropMips,
  chooseMip,
  intersects,
  visibleHexes,
} from "./mapGeometry.ts";
import { terrainDraws, type BakedLayers, type MapScene } from "./mapScene.ts";
import { composeScene, reworkScene } from "./dynamicScene.ts";
import type { OriginalTextures } from "./original/textures.ts";
import { MapImages, type MapImage, type ReadyImage } from "./mapImages.ts";
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

// The gas cloud (drawGas): puffs per hex, the turns left from which a cloud shows at full strength, a puff's peak
// opacity and its chlorine yellow-green, as the desktop draws it.
const gasPuffs = 7,
  gasFullTurn = 2,
  gasOpacity = 0.5,
  gasColour = 0xc4cc5c;
// The side's own mines (drawTunnels): dashes round the ring, its radius and line width, its earth brown and the
// opacity of the dashes lit and still dark, as the desktop draws it.
const tunnelDashes = 12,
  tunnelRadius = 30,
  tunnelWidth = 3,
  tunnelColour = 0x966432,
  tunnelLit = 230 / 255,
  tunnelDark = 90 / 255;
// The original tile is a 24×24 box filled by a hex whose hexes repeat every 32 pixels along a line and every 12
// pixels down; frames run clockwise from north.
const originalHex = 24,
  originalColPitch = 32,
  originalLinePitch = 12,
  originalFacings = ["n", "ne", "se", "s", "sw", "nw"];
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
  /** The extension's gas cloud over the units (drawGas). */
  private readonly gasLayer = new Graphics();
  private gasCells: hexgrid.Coord[] = [];
  /** The side's own mines being dug (drawTunnels). */
  private readonly tunnelLayer = new Graphics();
  private readonly overlays = new Graphics();
  private readonly effects = new Container();
  private readonly sprites = new Map<string, Sprite>();
  /** The decoded image each sprite shows, kept while its successor decodes. */
  private readonly shown = new Map<string, ReadyImage>();
  /** The backdrop mip levels the last frame drew. */
  private backdropShown: number[] = [];
  private readonly used = new Set<string>();
  private readonly counts = new Map<Container, number>();
  private readonly catalogs = new Map<string, Catalog>();
  private readonly requested = new Set<string>();
  private readonly desired = new Map<string, number>();
  private glyphs: ReadonlyMap<string, MapImage> = new Map();
  backdrop = true;
  dirty = true;
  get ready(): boolean {
    return this.stats.hexes > 0 && this.images.cache.ready;
  }
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
  private baked: BakedLayers;
  /** A level without a baked scene (an original map) composes it once the terrain catalog is known. */
  private composing: boolean;
  /** The original design's art (original/state.ts); null draws the remake's art. */
  original: OriginalTextures | null = null;
  private readonly originalLayer = new Container();

  constructor(
    readonly app: Application,
    readonly loader: AssetLoader,
    baked: BakedLayers | null,
    readonly battle: MapBattle,
    readonly families: Readonly<Record<string, SpriteFamily>>,
    private readonly pyramid: Pyramid | undefined,
    private readonly report: (message: string) => void,
  ) {
    this.composing = baked === null;
    this.baked = baked ?? {
      scene: composeScene(battle.session.level, () => false),
      seasons: new Map(),
      water: new Map(),
      blend: new Map(),
      backgrounds: new Map(),
      seasonBakes: {
        layers: new Map(),
        ground: new Map(),
        standing: new Map(),
      },
    };
    this.scene = this.baked.scene;
    this.drawnLevel = battle.session.level;
    const [width, height] = hexgrid.bounds(this.scene.grid);
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
        // Offline, full-quality upgrades and unvisited art are expected to be missing.
        if (navigator.onLine) report(e.message);
      },
    );
    this.root.addChild(
      this.backdropLayer,
      this.terrainLayer,
      this.originalLayer,
      this.ownership,
      this.unitLayer,
      this.markers,
      this.walkingLayer,
      this.gasLayer,
      this.tunnelLayer,
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
    if (!this.dirty && this.lastVersion === this.battle.version) {
      if (!this.battle.playback.reducedMotion) this.drawGas();
      return;
    }
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
    const original = this.original;
    const painted =
      original === null && this.backdrop && this.pyramid !== undefined;
    if (original !== null) this.paintOriginal(original, cells);
    else if (painted) this.paintBackdrop(this.pyramid, rect);
    if (original === null && (!painted || this.changed.size > 0))
      for (const { draw, image } of this.terrainImages(cells, painted)) {
        this.sprite(draw.key, this.terrainLayer, image, draw.at);
        this.stats.terrain++;
      }
    const visible = new Set(cells.map(level.coordKey));
    for (const [key, b] of view.buildings)
      if (b.owner !== 0 && visible.has(key)) {
        const [x, y] = hexgrid.center(play.coordFromKey(key));
        this.ownership.rect(x - 20, y - 34, 40, 6).fill(ownerColor(b.owner));
      }
    const moving = this.battle.playback.moving();
    const frames =
      original !== null
        ? []
        : unitFrames(
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
    if (original !== null) this.originalUnits(original, view, visible, moving);
    else if (moving !== null) {
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
    if (original === null) this.prefetch(rect, visible, painted, frames);
    this.gasCells = cells;
    this.drawGas();
    this.drawTunnels();
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
        this.shown.delete(key);
      }
    this.images.cache.end();
    this.stats.decodedBytes = this.images.cache.bytes;
  }
  /** terrainImages are the terrain images of cells with their draws, as the frame shows them. */
  private *terrainImages(
    cells: readonly hexgrid.Coord[],
    painted: boolean,
  ): Generator<{
    draw: ReturnType<typeof terrainDraws>[number];
    image: MapImage;
  }> {
    const catalog = this.catalogs.get("terrain");
    if (catalog === undefined) return;
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
        if (baked?.id === draw.id && !(painted && baked.tileOnly))
          image = baked;
        if (image !== undefined && draw.pass === "bleed")
          image = { ...image, bleed: true };
      }
      if (image !== undefined) yield { draw, image };
    }
  }
  /**
   * prefetch decodes, after what the frame shows, the art of a ring around the view a third of its size wide, so
   * that panning reveals drawn hexes, units and backdrop tiles instead of slate. It waits for full art to be
   * allowed (the first playable frame stays small) and skips far zoom, where the view holds most of a map.
   */
  private prefetch(
    rect: Rect,
    visible: ReadonlySet<string>,
    painted: boolean,
    frames: readonly UnitFrame[],
  ): void {
    if (this.loader.upgradesDeferred || this.camera.effective() < 0.75) return;
    const mx = rect.width / 3,
      my = rect.height / 3,
      wide = {
        x: rect.x - mx,
        y: rect.y - my,
        width: rect.width + 2 * mx,
        height: rect.height + 2 * my,
      };
    const ring = visibleHexes(this.scene.grid, wide, 0).filter(
      (c) => !visible.has(level.coordKey(c)),
    );
    if (!painted || this.changed.size > 0)
      for (const { image } of this.terrainImages(ring, painted))
        this.images.prefetch(image);
    if (painted && this.pyramid !== undefined) {
      const p = this.pyramid,
        index = chooseMip(
          p,
          this.camera.world,
          Math.min(1, this.camera.effective()),
        ),
        mip = p.mips[index];
      for (const tile of mip?.tiles ?? []) {
        if (mip === undefined) break;
        const draw = backdropDraw(p, mip, tile, this.camera.world);
        if (intersects(wide, draw.world) && !intersects(rect, draw.world))
          this.images.prefetch(backdropImage(tile, draw.frame));
      }
    }
    const near = new Set(ring.map(level.coordKey));
    for (const frame of frames)
      if (near.has(level.coordKey(frame.unit.at))) {
        const image = this.unitImage(frame);
        if (image !== undefined) this.images.prefetch(image);
      }
  }
  /**
   * drawTunnels marks the hexes the shown side's tunnellers mine with a dashed ring that closes as the tunnel nears
   * its blast; the other side's stay hidden.
   */
  private drawTunnels(): void {
    const g = this.tunnelLayer;
    g.clear();
    for (const t of play.tunnels(this.battle.session, this.battle.side)) {
      const [cx, cy] = hexgrid.center(t.mine),
        lit = Math.trunc((tunnelDashes * (t.dug + 1)) / (t.turns + 1));
      for (let k = 0; k < tunnelDashes; k++) {
        const a0 = (2 * Math.PI * k) / tunnelDashes,
          a1 = a0 + Math.PI / tunnelDashes;
        g.moveTo(
          cx + tunnelRadius * Math.cos(a0),
          cy + tunnelRadius * Math.sin(a0),
        )
          .lineTo(
            cx + tunnelRadius * Math.cos(a1),
            cy + tunnelRadius * Math.sin(a1),
          )
          .stroke({
            color: tunnelColour,
            width: tunnelWidth,
            alpha: k < lit ? tunnelLit : tunnelDark,
            cap: "round",
          });
      }
    }
  }
  /**
   * drawGas draws the gas pioneers' cloud (docs/new-units.md) on the visible hexes: soft yellow-green puffs that
   * drift slowly and fade as the cloud thins; still under reduced motion. Each puff is four concentric discs, a
   * cheap falloff to its rim.
   */
  private drawGas(): void {
    const g = this.gasLayer;
    g.clear();
    const gas = play.gas(this.battle.session);
    if (gas.size === 0) return;
    const t = this.battle.playback.reducedMotion
      ? 0
      : performance.now() / (1000 / 60);
    for (const c of this.gasCells) {
      const turns = gas.get(level.coordKey(c)) ?? 0;
      if (turns === 0) continue;
      const strength = Math.min(1, 0.55 + (0.45 * turns) / gasFullTurn),
        [cx, cy] = hexgrid.center(c),
        seed = c.line * 31 + c.col * 17;
      for (let k = 0; k < gasPuffs; k++) {
        const a = seed + k * 2.39996, // the golden angle spreads the puffs round the centre
          r = 8 + 22 * Math.abs(Math.sin(a * 1.7)),
          dx = r * Math.cos(a) + 6 * Math.sin(t / 95 + a),
          dy = 0.7 * r * Math.sin(a) + 4 * Math.cos(t / 120 + a * 1.3),
          rad = 40 + 10 * Math.sin(t / 75 + a * 0.8),
          alpha = gasOpacity * strength * (0.7 + 0.3 * Math.sin(t / 60 + a));
        for (const f of [1, 0.75, 0.5, 0.25])
          g.circle(cx + dx, cy + dy, rad * f).fill({
            color: gasColour,
            alpha: alpha * 0.3,
          });
      }
    }
  }
  private sprite(
    key: string,
    parent: Container,
    image: MapImage,
    at: Point,
    alpha = 1,
    tint = 0xffffff,
    size?: Rect,
    decodedOnly = false,
  ): Sprite | null {
    let ready = decodedOnly ? this.images.peek(image) : this.images.get(image);
    if (decodedOnly && ready !== null && !this.images.hold(ready)) ready = null;
    if (ready === null) {
      // Until the new image has decoded, the sprite keeps the one it shows (another density, lite art): a frame
      // never loses art it had.
      const last = this.shown.get(key);
      if (last === undefined || !this.images.hold(last)) return null;
      ready = last;
    }
    this.shown.set(key, ready);
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
  /** unitImage is a unit frame's image, dressed for the ground it stands on. */
  private unitImage(frame: UnitFrame): MapImage | undefined {
    const image = this.catalogs.get("units")?.images.get(frame.id);
    if (image === undefined) return undefined;
    const h = this.drawnLevel.lines[frame.drawAt.line]?.[frame.drawAt.col];
    return {
      ...image,
      dressing: {
        rim: rimFor(h?.terrain ?? "", (h?.infra.ids.length ?? 0) > 0),
        flying: frame.unit.domain === play.Air,
      },
    };
  }
  private unit(frame: UnitFrame, parent: Container): void {
    const image = this.unitImage(frame);
    if (image !== undefined) this.sprite(frame.key, parent, image, frame.at);
  }
  private paintBackdrop(p: Pyramid, rect: Rect): void {
    const effective = this.camera.effective();
    const lite = chooseMip(p, this.camera.world, Math.min(1, effective));
    const decoded = (index: number, tile: Tile) => {
      const mip = p.mips[index];
      if (mip === undefined) return false;
      const draw = backdropDraw(p, mip, tile, this.camera.world);
      return (
        this.images.peek(backdropImage(tile, draw.frame, index !== lite)) !==
        null
      );
    };
    const tiers = backdropMips(
      p,
      this.camera.world,
      effective,
      rect,
      !this.loader.upgradesDeferred,
      (mip, tile) => decoded(p.mips.indexOf(mip), tile),
    );
    // A new base level (zoomed past a mip) decodes under the levels the last frame drew, so that no tile goes blank.
    const base = tiers[0] ?? lite,
      baseMip = p.mips[base],
      complete =
        baseMip === undefined ||
        baseMip.tiles.every(
          (tile) =>
            !intersects(
              rect,
              backdropDraw(p, baseMip, tile, this.camera.world).world,
            ) || decoded(base, tile),
        );
    const under = complete
      ? []
      : this.backdropShown.filter((index) => !tiers.includes(index));
    this.backdropShown = [];
    for (const index of [...under, ...tiers]) {
      const mip = p.mips[index];
      if (mip === undefined) continue;
      const decodedOnly = under.includes(index);
      let drawn = false;
      for (const tile of mip.tiles) {
        const draw = backdropDraw(p, mip, tile, this.camera.world);
        if (!intersects(rect, draw.world)) continue;
        const sprite = this.sprite(
          `backdrop/${index}/${tile.x}/${tile.y}`,
          this.backdropLayer,
          backdropImage(tile, draw.frame, index !== lite),
          { x: 0, y: 0 },
          1,
          0xffffff,
          draw.world,
          decodedOnly,
        );
        if (!decodedOnly) this.stats.backdropTiles++;
        if (sprite !== null) drawn = true;
      }
      if (drawn) this.backdropShown.push(index);
      if (!decodedOnly) this.stats.mip = index;
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
    if (name === "terrain" && this.composing) {
      this.composing = false;
      this.baked = {
        ...this.baked,
        scene: composeScene(this.battle.session.level, (id) => images.has(id)),
      };
      this.scene = this.baked.scene;
      this.workKey = "";
      this.terrainVersion = -1;
    }
    this.dirty = true;
  }
  /**
   * paintOriginal draws the original tile of every hex in view, scaled from the original's hex pitches (32 pixels
   * along a line, 12 down) onto the remake's, so that the hexes interlock as in the original: the terrain record
   * the board holds now, sappers' work included (side 0's map, as the desktop drew it).
   */
  private paintOriginal(
    art: OriginalTextures,
    cells: readonly hexgrid.Coord[],
  ): void {
    const s = this.battle.session,
      tiles = art.tiles[s.board.winter ? 1 : 0],
      map = s.game.maps[0];
    for (const c of cells) {
      const cell = bind.cell(s.board, c);
      if (cell === null) continue;
      const tile = tiles[map[2 * cell] ?? -1];
      if (tile !== undefined)
        this.originalSprite(`original/tile/${c.line},${c.col}`, tile, c);
    }
  }
  private originalSprite(
    key: string,
    texture: Texture,
    c: hexgrid.Coord | Point,
    alpha = 1,
  ): void {
    const [x, y] = "line" in c ? hexgrid.center(c) : [c.x, c.y];
    this.used.add(key);
    let sprite = this.sprites.get(key);
    if (sprite === undefined) {
      sprite = new Sprite(texture);
      this.sprites.set(key, sprite);
    } else sprite.texture = texture;
    const parent = this.originalLayer,
      index = this.counts.get(parent) ?? 0;
    if (sprite.parent !== parent) parent.addChild(sprite);
    if (parent.getChildIndex(sprite) !== index)
      parent.setChildIndex(sprite, index);
    this.counts.set(parent, index + 1);
    // One world pixel more each way, so that rounding leaves no seam.
    const sx = (hexgrid.colPitch + 1) / originalColPitch,
      sy = (hexgrid.linePitch + 1) / originalLinePitch;
    sprite.width = texture.width * sx;
    sprite.height = texture.height * sy;
    sprite.position.set(x - (originalHex / 2) * sx, y - (originalHex / 2) * sy);
    sprite.alpha = alpha;
    sprite.tint = 0xffffff;
  }
  /** originalUnits draws the shown side's units with the original sprites, grey when moved. */
  private originalUnits(
    art: OriginalTextures,
    view: play.View,
    visible: ReadonlySet<string>,
    moving: ReturnType<MapBattle["playback"]["moving"]>,
  ): void {
    const frame = (u: play.UnitView, facing: string, moved: boolean) => {
      const k = 6 * u.type + Math.max(0, originalFacings.indexOf(facing));
      return art.units[u.side]?.[moved ? 1 : 0]?.[k];
    };
    const units = [...view.units.values()].sort(
      (a, b) => a.at.line - b.at.line || a.at.col - b.at.col,
    );
    for (const u of units) {
      if (moving?.unit.index === u.index || !visible.has(level.coordKey(u.at)))
        continue;
      const facing = this.battle.playback.directions.get(u.index) ?? "n";
      const t = frame(u, facing, u.moved && u.side === view.side);
      if (t === undefined) continue;
      this.originalSprite(`original/unit/${u.index}`, t, u.at);
      this.stats.units++;
      if (this.camera.effective() >= 0.35 && u.size > 0) this.mark(u);
    }
    if (moving !== null) {
      const t = frame(moving.unit, moving.facing, false);
      if (t !== undefined)
        this.originalSprite("original/walking", t, moving.at);
    }
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

function backdropImage(tile: Tile, frame: Rect, optional = false): MapImage {
  return {
    id: "backdrop",
    width: tile.sourceWidth,
    height: tile.sourceHeight,
    density: 1,
    logicalAnchor: [0, 0],
    lite: tile.file,
    full: tile.file,
    frame,
    optional,
  };
}
