import { AssetStore, type StoreOptions } from "./assetFetch.ts";
import {
  parseAssetIndex,
  parseImageCatalog,
  type AssetIndex,
  type ImageEntry,
} from "./assetManifest.ts";

/** The native assetDensity rule: smallest non-magnified pack, capped at 4x. */
export function assetDensity(
  zoom: number,
  devicePixelRatio: number,
): 1 | 2 | 4 {
  const effective = zoom * devicePixelRatio;
  if (!Number.isFinite(effective) || zoom <= 0 || devicePixelRatio <= 0)
    throw new Error("invalid asset scale");
  return effective <= 1 ? 1 : effective <= 2 ? 2 : 4;
}
export interface LoadedImage {
  entry: ImageEntry;
  bytes: ArrayBuffer;
  quality: "lite" | "full";
}

export class AssetLoader {
  readonly index: AssetIndex;
  readonly store: AssetStore;

  constructor(
    index: unknown,
    baseURL: string | URL,
    options: StoreOptions = {},
  ) {
    this.index = parseAssetIndex(index);
    this.store = new AssetStore(baseURL, options);
  }
  static async open(
    indexURL: string | URL,
    options: StoreOptions = {},
  ): Promise<AssetLoader> {
    const url = new URL(indexURL);
    const response = await (options.fetch ?? fetch)(url, { cache: "no-cache" });
    if (!response.ok) throw new Error(`asset index fetch ${response.status}`);
    const index: unknown = await response.json();
    return new AssetLoader(index, new URL("./", url), options);
  }
  async rule(name: string): Promise<unknown> {
    const ref = this.index.rules[name];
    if (ref === undefined) throw new Error(`missing rule ${name}`);
    return this.store.json(ref);
  }
  async levelFiles(month: string): Promise<Map<string, unknown>> {
    const level = this.index.levels[month];
    if (level === undefined) throw new Error(`missing level ${month}`);
    return new Map(
      await Promise.all(
        Object.entries(level.files).map(
          async ([name, ref]) => [name, await this.store.json(ref)] as const,
        ),
      ),
    );
  }
  async baked(month: string, mode: string): Promise<unknown> {
    const ref = this.index.levels[month]?.layers[mode];
    if (ref === undefined)
      throw new Error(`missing baked layers ${month}:${mode}`);
    return this.store.json(ref);
  }

  /** Only requested IDs are fetched. All lite images are delivered before any
   * full image; a failed upgrade leaves already-delivered lite images usable. */
  async progressive(
    catalog: string,
    ids: readonly string[],
    zoom: number,
    dpr: number,
    onImage: (image: LoadedImage) => void | Promise<void>,
  ): Promise<Map<string, LoadedImage>> {
    const packs = this.index.packs[catalog];
    if (packs === undefined) throw new Error(`missing catalog ${catalog}`);
    const desired = assetDensity(zoom, dpr);
    const densities = Object.keys(packs)
      .map(Number)
      .sort((a, b) => a - b);
    const density = densities.find((d) => d >= desired) ?? densities.at(-1);
    const pair = density === undefined ? undefined : packs[String(density)];
    if (pair === undefined || density === undefined)
      throw new Error(`missing catalog density ${catalog}`);
    const requested = [...new Set(ids)];
    const result = new Map<string, LoadedImage>();
    let previous = new Map<string, ImageEntry>();
    for (const quality of ["lite", "full"] as const) {
      const manifest = parseImageCatalog(
        await this.store.json(pair[quality]),
        density,
      );
      const byID = new Map(manifest.entries.map((e) => [e.id, e]));
      const selected = requested.map((id) => {
        const entry = byID.get(id);
        if (entry === undefined)
          throw new Error(`missing asset ${id} at ${density}x`);
        return entry;
      });
      if (quality === "full") {
        for (const e of selected) {
          const old = previous.get(e.id);
          if (
            old === undefined ||
            old.width !== e.width ||
            old.height !== e.height ||
            old.density !== e.density ||
            old.logicalAnchor[0] !== e.logicalAnchor[0] ||
            old.logicalAnchor[1] !== e.logicalAnchor[1]
          )
            throw new Error(`asset upgrade geometry differs: ${e.id}`);
        }
      }
      await Promise.all(
        selected.map(async (entry) => {
          const old = result.get(entry.id);
          const bytes =
            old?.entry.file.sha256 === entry.file.sha256
              ? old.bytes
              : await this.store.get(entry.file);
          const image: LoadedImage = {
            entry,
            bytes,
            quality,
          };
          result.set(entry.id, image);
          if (old?.entry.file.sha256 !== entry.file.sha256)
            await onImage(image);
        }),
      );
      previous = byID;
    }
    return result;
  }
}
