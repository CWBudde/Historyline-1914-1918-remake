import { fileRef, type FileRef } from "./assetManifest.ts";

export interface ContentCache {
  readonly persistent?: boolean;
  get(hash: string): Promise<ArrayBuffer | undefined>;
  put(hash: string, bytes: ArrayBuffer): Promise<void>;
  delete(hash: string): Promise<void>;
}
export interface StoreOptions {
  fetch?: typeof fetch;
  cache?: ContentCache | null;
  concurrency?: number;
}

export class AssetNetworkError extends Error {
  constructor(cause: TypeError) {
    super(cause.message, { cause });
    this.name = "AssetNetworkError";
  }
}

interface PendingAsset {
  promise: Promise<ArrayBuffer>;
  controller: AbortController;
  consumers: number;
  settled: boolean;
}

function abortError(signal: AbortSignal): Error {
  return signal.reason instanceof Error
    ? signal.reason
    : new DOMException("Asset request cancelled", "AbortError");
}

function cancellable<T>(promise: Promise<T>, signal?: AbortSignal): Promise<T> {
  if (signal === undefined) return promise;
  return new Promise((resolve, reject) => {
    const abort = () => {
      signal.removeEventListener("abort", abort);
      reject(abortError(signal));
    };
    signal.addEventListener("abort", abort, { once: true });
    if (signal.aborted) abort();
    void promise.then(
      (value) => {
        signal.removeEventListener("abort", abort);
        resolve(value);
      },
      (error: unknown) => {
        signal.removeEventListener("abort", abort);
        reject(error instanceof Error ? error : new Error(String(error)));
      },
    );
  });
}

class MemoryCache implements ContentCache {
  readonly persistent = false;
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
        persistent: true,
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
  private online = true;
  private waiters: (() => void)[] = [];
  private inflight = new Map<string, PendingAsset>();

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
  get networkAvailable(): boolean {
    return this.online;
  }
  async get(ref: FileRef, signal?: AbortSignal): Promise<ArrayBuffer> {
    fileRef(ref);
    signal?.throwIfAborted();
    const key = `${ref.sha256}:${ref.bytes}`;
    let pending = this.inflight.get(key);
    if (pending === undefined) {
      const controller = new AbortController();
      pending = {
        controller,
        consumers: 0,
        settled: false,
        promise: this.read(ref, controller.signal),
      };
      this.inflight.set(key, pending);
      const entry = pending;
      const settled = () => {
        entry.settled = true;
        if (this.inflight.get(key) === entry) this.inflight.delete(key);
      };
      void pending.promise.then(settled, settled);
    }
    pending.consumers++;
    try {
      return (await cancellable(pending.promise, signal)).slice(0);
    } finally {
      pending.consumers--;
      if (pending.consumers === 0 && !pending.settled) {
        if (this.inflight.get(key) === pending) this.inflight.delete(key);
        pending.controller.abort();
      }
    }
  }
  async json(ref: FileRef, signal?: AbortSignal): Promise<unknown> {
    const bytes = await this.get(ref, signal);
    return JSON.parse(new TextDecoder().decode(bytes)) as unknown;
  }
  async isPersistent(): Promise<boolean> {
    return (await this.cache)?.persistent === true;
  }
  /** Checks verified local bytes without falling back to the network. */
  async has(ref: FileRef): Promise<boolean> {
    fileRef(ref);
    const cache = await this.cache;
    try {
      const bytes = await cache?.get(ref.sha256);
      if (bytes === undefined) return false;
      await this.verified(ref, bytes);
      return true;
    } catch {
      try {
        await cache?.delete(ref.sha256);
      } catch {
        /* An unavailable cache is not an offline copy. */
      }
      return false;
    }
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
  private acquire(signal: AbortSignal): Promise<void> {
    signal.throwIfAborted();
    if (this.active < this.limit) {
      this.active++;
      return Promise.resolve();
    }
    return new Promise((resolve, reject) => {
      const next = () => {
        signal.removeEventListener("abort", abort);
        resolve();
      };
      const abort = () => {
        const index = this.waiters.indexOf(next);
        if (index >= 0) this.waiters.splice(index, 1);
        reject(abortError(signal));
      };
      this.waiters.push(next);
      signal.addEventListener("abort", abort, { once: true });
      if (signal.aborted) abort();
    });
  }
  private async transport<T>(
    operation: () => Promise<T>,
    signal: AbortSignal,
  ): Promise<T> {
    try {
      return await cancellable(operation(), signal);
    } catch (error) {
      if (!signal.aborted && error instanceof TypeError) {
        this.online = false;
        throw new AssetNetworkError(error);
      }
      throw error;
    }
  }
  private async read(ref: FileRef, signal: AbortSignal): Promise<ArrayBuffer> {
    const cache = await cancellable(this.cache, signal);
    try {
      const bytes =
        cache === null
          ? undefined
          : await cancellable(cache.get(ref.sha256), signal);
      if (bytes !== undefined) {
        await this.verified(ref, bytes);
        return bytes;
      }
    } catch {
      signal.throwIfAborted();
      try {
        await cache?.delete(ref.sha256);
      } catch {
        /* Cache failures cannot prevent a network load. */
      }
    }
    await this.acquire(signal);
    let bytes: ArrayBuffer;
    try {
      signal.throwIfAborted();
      const response = await this.transport(
        () => this.fetcher(new URL(ref.path, this.baseURL), { signal }),
        signal,
      );
      if (!response.ok)
        throw new Error(`asset fetch ${response.status}: ${ref.path}`);
      this.online = true;
      bytes = await this.transport(() => response.arrayBuffer(), signal);
      await this.verified(ref, bytes);
    } finally {
      const next = this.waiters.shift();
      if (next === undefined) this.active--;
      else next();
    }
    signal.throwIfAborted();
    try {
      await cache?.put(ref.sha256, bytes);
    } catch {
      /* Quota failures leave the verified response usable. */
    }
    return bytes;
  }
}
