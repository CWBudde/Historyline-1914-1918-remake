import { Rectangle, Texture } from "pixi.js";
import {
  parseImageCatalog,
  type ImageEntry,
  type FileRef,
} from "./assetManifest.ts";
import { AssetLoader } from "./assets.ts";
import { AssetNetworkError } from "./assetFetch.ts";
import { RenderCache, type RenderResource } from "./renderCache.ts";
import { dress, dressMargin, padded, type Rim } from "./unitDress.ts";
import type { Rect } from "./camera.ts";

export interface MapImage {
  id: string;
  width: number;
  height: number;
  density: number;
  logicalAnchor: [number, number];
  lite: FileRef;
  full: FileRef;
  frame?: Rect;
  dressing?: { rim: Rim; flying: boolean };
  bleed?: boolean;
  nearest?: boolean;
  optional?: boolean;
}
export interface ReadyImage {
  texture: Texture;
  width: number;
  height: number;
  density: number;
  anchor: [number, number];
}

interface SharedBitmap {
  promise: Promise<ImageBitmap>;
  /** Draws that keep the bitmap as their texture source. */
  owners: number;
}
/** Decodes run off the main thread; the unit dress and bleed copies on it. */
const decodeConcurrency = Math.max(
  4,
  Math.min(
    8,
    (globalThis as { navigator?: { hardwareConcurrency?: number } }).navigator
      ?.hardwareConcurrency ?? 4,
  ),
);

/** Catalogs are parsed once per density, images only when a culled draw uses
 * them. Full images replace lite images without changing world geometry. */
export class MapImages {
  private readonly catalogs = new Map<
    string,
    Promise<ReadonlyMap<string, MapImage>>
  >();
  private stopped = false;
  private readonly abort = new AbortController();
  private readonly decoding = new Map<string, SharedBitmap>();
  private readonly keys = new WeakMap<ReadyImage, string>();
  readonly cache: RenderCache<ReadyImage>;
  constructor(
    readonly loader: AssetLoader,
    private readonly changed: () => void,
    private readonly error: (e: Error) => void,
  ) {
    this.cache = new RenderCache(
      192 * 1024 * 1024,
      decodeConcurrency,
      changed,
      error,
    );
  }
  catalog(
    name: string,
    density: number,
  ): Promise<ReadonlyMap<string, MapImage>> {
    const key = `${name}/${density}`;
    let pending = this.catalogs.get(key);
    if (pending === undefined) {
      pending = this.loadCatalog(name, density);
      this.catalogs.set(key, pending);
    }
    return pending;
  }
  private async loadCatalog(
    name: string,
    density: number,
  ): Promise<ReadonlyMap<string, MapImage>> {
    const pair = this.loader.index.packs[name]?.[String(density)];
    if (pair === undefined)
      throw new Error(`missing ${name} catalog at ${density}x`);
    const lite = parseImageCatalog(
      await this.loader.store.json(pair.lite, this.abort.signal),
      density,
    );
    const images = new Map(
      lite.entries.map((e) => [e.id, { ...e, lite: e.file, full: e.file }]),
    );
    const upgrade = async () => {
      await this.loader.waitForUpgrades(this.abort.signal);
      if (this.isStopped()) return;
      const full = parseImageCatalog(
        await this.loader.store.json(pair.full, this.abort.signal),
        density,
      );
      if (this.isStopped()) return;
      const complete = new Map(full.entries.map((e) => [e.id, e]));
      for (const e of lite.entries) {
        const next = complete.get(e.id);
        if (next === undefined || !sameGeometry(e, next))
          throw new Error(`asset upgrade geometry differs: ${e.id}`);
      }
      for (const [id, image] of images) {
        const next = complete.get(id);
        if (next !== undefined) image.full = next.file;
      }
      this.changed();
    };
    if (this.loader.upgradesDeferred) {
      void upgrade().catch((error: unknown) => {
        if (!this.stopped && !(error instanceof AssetNetworkError))
          this.error(error instanceof Error ? error : new Error(String(error)));
      });
    } else {
      try {
        await upgrade();
      } catch (error) {
        if (!(error instanceof AssetNetworkError)) throw error;
      }
    }
    return images;
  }
  private key(image: MapImage, ref: FileRef): string {
    return `${ref.sha256}/${JSON.stringify(image.frame ?? null)}/${JSON.stringify(image.dressing ?? null)}/${image.density}/${image.bleed === true}/${image.nearest === true}/${image.logicalAnchor.join(",")}`;
  }
  peek(image: MapImage): ReadyImage | null {
    return (
      this.cache.peek(this.key(image, image.full)) ??
      this.cache.peek(this.key(image, image.lite))
    );
  }
  /**
   * hold keeps an image decoded earlier for the current frame, where it stands in for its successor (another
   * density, lite before full) until that has decoded; false once it has been released.
   */
  hold(ready: ReadyImage): boolean {
    const key = this.keys.get(ready);
    return key !== undefined && this.cache.hold(key);
  }
  get(image: MapImage): ReadyImage | null {
    const full = this.key(image, image.full),
      lite = this.key(image, image.lite);
    if (this.cache.peek(full) !== null)
      return this.cache.get(full, () => this.decode(image, image.full, full));
    const report = (error: Error) => {
      if (!(error instanceof AssetNetworkError)) this.error(error);
    };
    const current = this.cache.get(
      lite,
      () => this.decode(image, image.lite, lite),
      image.optional === true ? report : undefined,
    );
    if (current !== null && lite !== full && this.loader.upgradesAllowed)
      this.cache.get(full, () => this.decode(image, image.full, full), report);
    return current;
  }
  /**
   * bitmap fetches and decodes a file once for every draw that starts on it while it decodes: a background tile
   * and its bleed copy are one decode. A draw that keeps the bitmap as its texture owns it; the others copy it.
   */
  private bitmap(
    ref: FileRef,
    owner: boolean,
  ): Promise<{ bitmap: ImageBitmap; shared: SharedBitmap }> {
    let shared = this.decoding.get(ref.sha256);
    // Two textures never share one bitmap: each closes its own.
    if (shared === undefined || (owner && shared.owners > 0)) {
      const fresh: SharedBitmap = {
        owners: 0,
        promise: (async () => {
          const bytes = await this.loader.store.get(ref, this.abort.signal);
          return createImageBitmap(new Blob([bytes]));
        })(),
      };
      if (shared === undefined) this.decoding.set(ref.sha256, fresh);
      shared = fresh;
      const settled = () => {
        if (this.decoding.get(ref.sha256) === fresh)
          this.decoding.delete(ref.sha256);
      };
      fresh.promise.then(settled, settled);
    }
    if (owner) shared.owners++;
    const entry = shared;
    return shared.promise.then((bitmap) => ({ bitmap, shared: entry }));
  }
  private async decode(
    image: MapImage,
    ref: FileRef,
    key: string,
  ): Promise<RenderResource<ReadyImage>> {
    const copied = image.dressing !== undefined || image.bleed === true;
    const { bitmap, shared } = await this.bitmap(ref, !copied);
    // A copy closes the bitmap only when no texture keeps it.
    const release = () => {
      if (!copied || shared.owners === 0) bitmap.close();
    };
    if (bitmap.width !== image.width || bitmap.height !== image.height) {
      release();
      throw new Error(`image geometry differs: ${image.id}`);
    }
    let source: ImageBitmap | HTMLCanvasElement = bitmap;
    const anchor: [number, number] = [...image.logicalAnchor];
    try {
      if (copied) {
        const canvas = document.createElement("canvas");
        canvas.width = bitmap.width;
        canvas.height = bitmap.height;
        const context = canvas.getContext("2d", { willReadFrequently: true });
        if (context === null) throw new Error("missing sprite canvas context");
        context.drawImage(bitmap, 0, 0);
        release();
        const raw = context.getImageData(0, 0, canvas.width, canvas.height);
        const dressed =
          image.dressing === undefined
            ? padded(raw, 3, true)
            : dress(
                raw,
                image.density,
                image.dressing.flying,
                image.dressing.rim,
              );
        canvas.width = dressed.width;
        canvas.height = dressed.height;
        const pixels = context.createImageData(dressed.width, dressed.height);
        pixels.data.set(dressed.data);
        context.putImageData(pixels, 0, 0);
        source = canvas;
        const margin =
          (image.dressing === undefined ? 3 : dressMargin(image.density)) /
          image.density;
        anchor[0] += margin;
        anchor[1] += margin;
      }
      const base = Texture.from(source, true);
      base.source.style.scaleMode =
        image.nearest === true ? "nearest" : "linear";
      const texture =
        image.frame === undefined
          ? base
          : new Texture({
              source: base.source,
              frame: new Rectangle(
                image.frame.x,
                image.frame.y,
                image.frame.width,
                image.frame.height,
              ),
            });
      const value: ReadyImage = {
        texture,
        width: texture.width,
        height: texture.height,
        density: image.density,
        anchor,
      };
      this.keys.set(value, key);
      return {
        value,
        bytes: source.width * source.height * 4,
        dispose: () => {
          if (texture !== base) texture.destroy();
          base.destroy(true);
          if (!copied) bitmap.close();
        },
      };
    } catch (error) {
      if (!copied) bitmap.close();
      throw error;
    }
  }
  dispose(): void {
    this.stopped = true;
    this.abort.abort();
    this.cache.dispose();
  }
  private isStopped(): boolean {
    return this.stopped;
  }
}
function sameGeometry(a: ImageEntry, b: ImageEntry): boolean {
  return (
    a.width === b.width &&
    a.height === b.height &&
    a.density === b.density &&
    a.logicalAnchor[0] === b.logicalAnchor[0] &&
    a.logicalAnchor[1] === b.logicalAnchor[1]
  );
}
