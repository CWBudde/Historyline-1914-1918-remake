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

/** Catalogs are parsed once per density, images only when a culled draw uses
 * them. Full images replace lite images without changing world geometry. */
export class MapImages {
  private readonly catalogs = new Map<
    string,
    Promise<ReadonlyMap<string, MapImage>>
  >();
  private stopped = false;
  private readonly abort = new AbortController();
  readonly cache: RenderCache<ReadyImage>;
  constructor(
    readonly loader: AssetLoader,
    private readonly changed: () => void,
    private readonly error: (e: Error) => void,
  ) {
    this.cache = new RenderCache(192 * 1024 * 1024, 3, changed, error);
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
  get(image: MapImage): ReadyImage | null {
    const full = this.key(image, image.full),
      lite = this.key(image, image.lite);
    if (this.cache.peek(full) !== null)
      return this.cache.get(full, () => this.decode(image, image.full));
    const report = (error: Error) => {
      if (!(error instanceof AssetNetworkError)) this.error(error);
    };
    const current = this.cache.get(
      lite,
      () => this.decode(image, image.lite),
      image.optional === true ? report : undefined,
    );
    if (current !== null && lite !== full && this.loader.upgradesAllowed)
      this.cache.get(full, () => this.decode(image, image.full), report);
    return current;
  }
  private async decode(
    image: MapImage,
    ref: FileRef,
  ): Promise<RenderResource<ReadyImage>> {
    const bytes = await this.loader.store.get(ref, this.abort.signal);
    const bitmap = await createImageBitmap(new Blob([bytes]));
    if (bitmap.width !== image.width || bitmap.height !== image.height) {
      bitmap.close();
      throw new Error(`image geometry differs: ${image.id}`);
    }
    let source: ImageBitmap | HTMLCanvasElement = bitmap;
    const anchor: [number, number] = [...image.logicalAnchor];
    try {
      if (image.dressing !== undefined || image.bleed === true) {
        const canvas = document.createElement("canvas");
        canvas.width = bitmap.width;
        canvas.height = bitmap.height;
        const context = canvas.getContext("2d", { willReadFrequently: true });
        if (context === null) throw new Error("missing sprite canvas context");
        context.drawImage(bitmap, 0, 0);
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
        bitmap.close();
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
      return {
        value: {
          texture,
          width: texture.width,
          height: texture.height,
          density: image.density,
          anchor,
        },
        bytes: source.width * source.height * 4,
        dispose: () => {
          if (texture !== base) texture.destroy();
          base.destroy(true);
          bitmap.close();
        },
      };
    } catch (error) {
      bitmap.close();
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
