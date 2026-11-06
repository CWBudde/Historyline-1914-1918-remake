import { Rectangle, Texture } from "pixi.js";
import {
  parseImageCatalog,
  type ImageEntry,
  type FileRef,
} from "./assetManifest.ts";
import { AssetLoader } from "./assets.ts";
import { RenderCache, type RenderResource } from "./renderCache.ts";
import { dress, dressMargin, padded, type Rim } from "./unitDress.ts";
import type { Rect } from "./camera.ts";

export class CandidateArtError extends Error {
  constructor() {
    super("The current art pack is marked candidate.");
  }
}

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
  readonly cache: RenderCache<ReadyImage>;
  constructor(
    readonly loader: AssetLoader,
    changed: () => void,
    error: (e: Error) => void,
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
    const [lite, full] = await Promise.all(
      [pair.lite, pair.full].map(async (ref) =>
        parseImageCatalog(await this.loader.store.json(ref), density),
      ),
    );
    if (lite === undefined || full === undefined)
      throw new Error("missing image catalog");
    if (
      !this.loader.candidatesAllowed &&
      (lite.status === "candidate" || full.status === "candidate")
    )
      throw new CandidateArtError();
    const complete = new Map(full.entries.map((e) => [e.id, e]));
    return new Map(
      lite.entries.map((e) => {
        const next = complete.get(e.id);
        if (next === undefined || !sameGeometry(e, next))
          throw new Error(`asset upgrade geometry differs: ${e.id}`);
        return [e.id, { ...e, lite: e.file, full: next.file }];
      }),
    );
  }
  get(image: MapImage): ReadyImage | null {
    const key = (ref: FileRef) =>
      `${ref.sha256}/${JSON.stringify(image.frame ?? null)}/${JSON.stringify(image.dressing ?? null)}/${image.density}/${image.bleed === true}/${image.nearest === true}/${image.logicalAnchor.join(",")}`;
    const full = key(image.full),
      lite = key(image.lite);
    if (this.cache.peek(full) !== null)
      return this.cache.get(full, () => this.decode(image, image.full));
    const current = this.cache.get(lite, () => this.decode(image, image.lite));
    if (current !== null && lite !== full)
      this.cache.get(full, () => this.decode(image, image.full));
    return current;
  }
  private async decode(
    image: MapImage,
    ref: FileRef,
  ): Promise<RenderResource<ReadyImage>> {
    const bytes = await this.loader.store.get(ref);
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
    this.cache.dispose();
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
