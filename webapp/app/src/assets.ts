import {
  AssetStore,
  AssetNetworkError,
  type StoreOptions,
} from "./assetFetch.ts";
import {
  parseAssetIndex,
  parseImageCatalog,
  parsePyramid,
  type AssetIndex,
  type ImageEntry,
  type Pyramid,
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
export interface LoaderOptions extends StoreOptions {
  deferUpgrades?: boolean;
}

export class AssetLoader {
  readonly index: AssetIndex;
  readonly store: AssetStore;
  private held: boolean;
  private readonly upgradeWaiters = new Set<() => void>();
  get upgradesAllowed(): boolean {
    const network: { navigator?: { onLine?: boolean } } = globalThis;
    return (
      !this.held &&
      this.store.networkAvailable &&
      (network.navigator?.onLine ?? true)
    );
  }
  get upgradesDeferred(): boolean {
    return this.held;
  }
  waitForUpgrades(signal?: AbortSignal): Promise<void> {
    signal?.throwIfAborted();
    if (!this.held) return Promise.resolve();
    return new Promise((resolve, reject) => {
      const done = () => {
        this.upgradeWaiters.delete(done);
        signal?.removeEventListener("abort", abort);
        resolve();
      };
      const abort = () => {
        this.upgradeWaiters.delete(done);
        signal?.removeEventListener("abort", abort);
        reject(
          signal?.reason instanceof Error
            ? signal.reason
            : new DOMException("Asset request cancelled", "AbortError"),
        );
      };
      this.upgradeWaiters.add(done);
      signal?.addEventListener("abort", abort, { once: true });
      if (signal?.aborted) abort();
    });
  }

  constructor(
    index: unknown,
    baseURL: string | URL,
    options: LoaderOptions = {},
  ) {
    this.index = parseAssetIndex(index);
    this.store = new AssetStore(baseURL, options);
    this.held = options.deferUpgrades ?? false;
  }
  enableUpgrades(): boolean {
    if (!this.held) return false;
    this.held = false;
    for (const done of this.upgradeWaiters) done();
    return true;
  }
  static async open(
    indexURL: string | URL,
    options: LoaderOptions = {},
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
  async backdrop(
    month: string,
    signal?: AbortSignal,
  ): Promise<Pyramid | undefined> {
    const entry = this.index.levels[month];
    if (entry === undefined) throw new Error(`missing level ${month}`);
    if (entry.backdropFile === undefined) return entry.backdrop;
    return parsePyramid(await this.store.json(entry.backdropFile, signal));
  }

  /** Only requested IDs are fetched. All lite images are delivered before any
   * full image; a failed upgrade leaves already-delivered lite images usable.
   * While full art is held, lite images come at the lowest density, unless
   * sharp asks for the drawn density (a painting that fills the window);
   * liteOnly stops after the lite image. */
  async progressive(
    catalog: string,
    ids: readonly string[],
    zoom: number,
    dpr: number,
    onImage: (image: LoadedImage) => void | Promise<void>,
    signal?: AbortSignal,
    sharp = false,
    liteOnly = false,
  ): Promise<Map<string, LoadedImage>> {
    signal?.throwIfAborted();
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
    const liteDensity =
      this.held && !sharp ? (densities[0] ?? density) : density;
    const litePair = packs[String(liteDensity)];
    if (litePair === undefined)
      throw new Error(`missing lite catalog density ${catalog}`);
    const requested = [...new Set(ids)];
    const result = new Map<string, LoadedImage>();
    let previous = new Map<string, ImageEntry>();
    for (const quality of ["lite", "full"] as const) {
      if (quality === "full") {
        if (liteOnly) break;
        await this.waitForUpgrades(signal);
        if (!this.upgradesAllowed) break;
      }
      try {
        const manifest = parseImageCatalog(
          await this.store.json(
            quality === "lite" ? litePair.lite : pair.full,
            signal,
          ),
          quality === "lite" ? liteDensity : density,
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
              old.width / old.density !== e.width / e.density ||
              old.height / old.density !== e.height / e.density ||
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
                : await this.store.get(entry.file, signal);
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
      } catch (error) {
        if (quality === "full" && error instanceof AssetNetworkError)
          return result;
        throw error;
      }
    }
    return result;
  }
}
