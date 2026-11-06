import { fileRef, type FileRef } from "./assetManifest.ts";

export interface ContentCache {
  get(hash: string): Promise<ArrayBuffer | undefined>;
  put(hash: string, bytes: ArrayBuffer): Promise<void>;
  delete(hash: string): Promise<void>;
}
export interface StoreOptions {
  fetch?: typeof fetch;
  cache?: ContentCache | null;
  concurrency?: number;
}

class MemoryCache implements ContentCache {
  private values = new Map<string, ArrayBuffer>();
  private size = 0;
  get(hash: string): Promise<ArrayBuffer | undefined> {
    const v = this.values.get(hash);
    return Promise.resolve(v?.slice(0));
  }
  put(hash: string, bytes: ArrayBuffer): Promise<void> {
    this.size -= this.values.get(hash)?.byteLength ?? 0;
    this.values.delete(hash);
    this.values.set(hash, bytes.slice(0));
    this.size += bytes.byteLength;
    while (this.size > 64 * 1024 * 1024) {
      const key = this.values.keys().next().value;
      if (key === undefined) break;
      this.size -= this.values.get(key)?.byteLength ?? 0;
      this.values.delete(key);
    }
    return Promise.resolve();
  }
  delete(hash: string): Promise<void> {
    this.size -= this.values.get(hash)?.byteLength ?? 0;
    this.values.delete(hash);
    return Promise.resolve();
  }
}

async function browserCache(): Promise<ContentCache> {
  if (typeof caches !== "undefined") {
    try {
      const cache = await caches.open("historyline-content-v1");
      const key = (hash: string) =>
        new URL(`/__historyline_content__/${hash}`, location.origin).href;
      return {
        get: async (hash) => {
          const r = await cache.match(key(hash));
          return r?.arrayBuffer();
        },
        put: async (hash, bytes) => {
          await cache.put(key(hash), new Response(bytes));
        },
        delete: async (hash) => {
          await cache.delete(key(hash));
        },
      };
    } catch {
      /* Private browsing and quota restrictions retain a bounded memory cache. */
    }
  }
  return new MemoryCache();
}

/** SHA-pinned fetches shared across catalogues, with a bounded network queue. */
export class AssetStore {
  readonly baseURL: URL;
  private readonly fetcher: typeof fetch;
  private readonly cache: Promise<ContentCache | null>;
  private readonly limit: number;
  private active = 0;
  private waiters: (() => void)[] = [];
  private inflight = new Map<string, Promise<ArrayBuffer>>();

  constructor(baseURL: string | URL, options: StoreOptions = {}) {
    this.baseURL = new URL(baseURL);
    this.fetcher = options.fetch ?? fetch.bind(globalThis);
    this.cache =
      options.cache === undefined
        ? browserCache()
        : Promise.resolve(options.cache);
    this.limit = options.concurrency ?? 6;
    if (!Number.isSafeInteger(this.limit) || this.limit < 1)
      throw new Error("invalid asset fetch concurrency");
  }
  async get(ref: FileRef): Promise<ArrayBuffer> {
    fileRef(ref);
    const key = `${ref.sha256}:${ref.bytes}`;
    let promise = this.inflight.get(key);
    if (promise === undefined) {
      promise = this.read(ref);
      this.inflight.set(key, promise);
      void promise
        .finally(() => {
          this.inflight.delete(key);
        })
        .catch(() => {
          /* The caller receives the original error. */
        });
    }
    return (await promise).slice(0);
  }
  async json(ref: FileRef): Promise<unknown> {
    return JSON.parse(new TextDecoder().decode(await this.get(ref))) as unknown;
  }
  private async verified(ref: FileRef, bytes: ArrayBuffer): Promise<void> {
    if (bytes.byteLength !== ref.bytes)
      throw new Error(`asset size mismatch: ${ref.path}`);
    const hash = Array.from(
      new Uint8Array(await crypto.subtle.digest("SHA-256", bytes)),
      (v) => v.toString(16).padStart(2, "0"),
    ).join("");
    if (hash !== ref.sha256)
      throw new Error(`asset hash mismatch: ${ref.path}`);
  }
  private async read(ref: FileRef): Promise<ArrayBuffer> {
    const cache = await this.cache;
    try {
      const bytes = await cache?.get(ref.sha256);
      if (bytes !== undefined) {
        await this.verified(ref, bytes);
        return bytes;
      }
    } catch {
      try {
        await cache?.delete(ref.sha256);
      } catch {
        /* Cache failures cannot prevent a network load. */
      }
    }
    if (this.active < this.limit) this.active++;
    else await new Promise<void>((resolve) => this.waiters.push(resolve));
    let bytes: ArrayBuffer;
    try {
      const response = await this.fetcher(new URL(ref.path, this.baseURL));
      if (!response.ok)
        throw new Error(`asset fetch ${response.status}: ${ref.path}`);
      bytes = await response.arrayBuffer();
      await this.verified(ref, bytes);
    } finally {
      const next = this.waiters.shift();
      if (next === undefined) this.active--;
      else next();
    }
    try {
      await cache?.put(ref.sha256, bytes);
    } catch {
      /* Quota failures leave the verified response usable. */
    }
    return bytes;
  }
}
