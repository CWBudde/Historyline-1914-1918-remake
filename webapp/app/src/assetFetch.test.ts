import { describe, expect, it, vi } from "vitest";
import {
  AssetStorageFullError,
  AssetStore,
  isStorageFull,
  type ContentCache,
} from "./assetFetch.ts";

async function reference(body: string | Uint8Array<ArrayBuffer>, ext = "json") {
  const bytes =
    typeof body === "string" ? new TextEncoder().encode(body) : body;
  const sum = new Uint8Array(await crypto.subtle.digest("SHA-256", bytes));
  const sha256 = Array.from(sum, (v) => v.toString(16).padStart(2, "0")).join(
    "",
  );
  return { path: `blobs/${sha256}.${ext}`, sha256, bytes: bytes.length };
}

class MemoryCache implements ContentCache {
  values = new Map<string, ArrayBuffer>();
  get(hash: string) {
    return Promise.resolve(this.values.get(hash));
  }
  put(hash: string, data: ArrayBuffer) {
    this.values.set(hash, data);
    return Promise.resolve();
  }
  delete(hash: string) {
    this.values.delete(hash);
    return Promise.resolve();
  }
}

describe("verified content storage", () => {
  it("verifies JSON before decoding it and reuses it from cache", async () => {
    const value = { name: "1914 — Österreich", values: [0, -12, 0.125] };
    const bytes = new TextEncoder().encode(JSON.stringify(value));
    const ref = await reference(bytes);
    const cache = new MemoryCache();
    const fetcher = vi.fn(() => Promise.resolve(new Response(bytes)));
    const store = new AssetStore("https://example.test/", {
      cache,
      fetch: fetcher,
    });
    expect(await store.json(ref)).toEqual(value);
    expect(await store.json(ref)).toEqual(value);
    expect(fetcher).toHaveBeenCalledTimes(1);
    const wrongHash = {
      ...ref,
      sha256: "0".repeat(64),
      path: `blobs/${"0".repeat(64)}.json`,
    };
    await expect(store.json(wrongHash)).rejects.toThrow(/hash/);
    // An integrity error is not a lost connection.
    expect(store.networkAvailable).toBe(true);
  });

  it("learns network failures separately from integrity errors and resets on a successful fetch", async () => {
    const a = await reference("a"),
      b = await reference("b");
    let online = false;
    const store = new AssetStore("https://example.test/", {
      cache: null,
      fetch: () =>
        online
          ? Promise.resolve(new Response("b"))
          : Promise.reject(new TypeError("offline")),
    });
    await expect(store.get(a)).rejects.toThrow("offline");
    expect(store.networkAvailable).toBe(false);
    online = true;
    await store.get(b);
    expect(store.networkAvailable).toBe(true);
    await expect(store.get(a)).rejects.toThrow(/hash/);
    expect(store.networkAvailable).toBe(true);
  });
  it("cancels one consumer without aborting another consumer's shared request", async () => {
    const ref = await reference("shared");
    let complete: (value: Response) => void = () => {};
    let started: () => void = () => {};
    const ready = new Promise<void>((resolve) => {
      started = resolve;
    });
    const response = new Promise<Response>((resolve) => {
      complete = resolve;
    });
    let networkSignal: AbortSignal | undefined;
    const store = new AssetStore("https://example.test/", {
      cache: null,
      fetch: (_input, options) => {
        networkSignal = options?.signal ?? undefined;
        started();
        return response;
      },
    });
    const a = new AbortController();
    const cancelled = store.get(ref, a.signal);
    const shared = store.get(ref);
    await ready;
    a.abort();
    await expect(cancelled).rejects.toThrow();
    expect(networkSignal?.aborted).toBe(false);
    complete(new Response("shared"));
    expect(new TextDecoder().decode(await shared)).toBe("shared");
  });

  it("releases cancelled active and queued requests so a retry can fetch", async () => {
    const a = await reference("a"),
      b = await reference("b"),
      c = await reference("c");
    const bodies = new Map([
      [a.path, "a"],
      [b.path, "b"],
      [c.path, "c"],
    ]);
    const calls: string[] = [];
    let started: () => void = () => {};
    const ready = new Promise<void>((resolve) => {
      started = resolve;
    });
    let firstSignal: AbortSignal | undefined;
    const store = new AssetStore("https://example.test/", {
      cache: null,
      concurrency: 1,
      fetch: (input, options) => {
        const path = new URL(
          input instanceof Request ? input.url : String(input),
        ).pathname.slice(1);
        calls.push(path);
        if (calls.length === 1) {
          firstSignal = options?.signal ?? undefined;
          started();
          return new Promise<Response>(() => {});
        }
        return Promise.resolve(new Response(bodies.get(path) ?? "missing"));
      },
    });
    const active = new AbortController(),
      queued = new AbortController();
    const first = store.get(a, active.signal);
    await ready;
    const second = store.get(b, queued.signal);
    queued.abort();
    await expect(second).rejects.toThrow();
    active.abort();
    await expect(first).rejects.toThrow();
    expect(firstSignal?.aborted).toBe(true);
    expect(new TextDecoder().decode(await store.get(c))).toBe("c");
    expect(new TextDecoder().decode(await store.get(a))).toBe("a");
    expect(calls).toEqual([a.path, c.path, a.path]);
  });
  it("reports durable cached content without fetching and rejects quota or corrupt copies", async () => {
    const cache = Object.assign(new MemoryCache(), { persistent: true });
    const ref = await reference("saved content");
    let fetched = 0;
    const store = new AssetStore("https://example.test/data/", {
      cache,
      fetch: () => {
        fetched++;
        return Promise.resolve(new Response("saved content"));
      },
    });
    expect(await store.isPersistent()).toBe(true);
    expect(await store.has(ref)).toBe(false);
    expect(fetched).toBe(0);
    await store.get(ref);
    expect(await store.has(ref)).toBe(true);
    cache.values.set(ref.sha256, new TextEncoder().encode("broken").buffer);
    expect(await store.has(ref)).toBe(false);
    expect(fetched).toBe(1);
    expect(cache.values.has(ref.sha256)).toBe(false);
    const quota = new AssetStore("https://example.test/data/", {
      cache: {
        persistent: true,
        get: () => Promise.resolve(undefined),
        put: () => Promise.reject(new Error("quota")),
        delete: () => Promise.resolve(),
      },
      fetch: () => Promise.resolve(new Response("saved content")),
    });
    await quota.get(ref);
    expect(await quota.has(ref)).toBe(false);
    expect(
      await new AssetStore("https://example.test/", {
        cache: null,
      }).isPersistent(),
    ).toBe(false);
    expect(
      await new AssetStore("https://example.test/", {
        cache: new MemoryCache(),
      }).isPersistent(),
    ).toBe(false);
  });
  it("keeps a file for offline play, and names a cache that refuses it as full", async () => {
    const ref = await reference("kept content");
    const cache = new MemoryCache();
    let refuse = true,
      fetched = 0;
    const store = new AssetStore("https://example.test/data/", {
      cache: {
        get: (hash) => cache.get(hash),
        put: (hash, bytes) =>
          refuse
            ? Promise.reject(
                new DOMException(
                  "The quota was exceeded",
                  "QuotaExceededError",
                ),
              )
            : cache.put(hash, bytes),
        delete: (hash) => cache.delete(hash),
      },
      fetch: () => {
        fetched++;
        return Promise.resolve(new Response("kept content"));
      },
    });
    // The game itself goes on with bytes the cache refused.
    expect(new TextDecoder().decode(await store.get(ref))).toBe("kept content");
    const full = await store.keep(ref).catch((error: unknown) => error);
    expect(full).toBeInstanceOf(AssetStorageFullError);
    expect(isStorageFull(full)).toBe(true);
    refuse = false;
    await store.keep(ref);
    expect(cache.values.has(ref.sha256)).toBe(true);
    await store.keep(ref);
    expect(fetched).toBe(3);
  });
  it("tells a full storage from other failures", () => {
    expect(isStorageFull(new DOMException("full", "QuotaExceededError"))).toBe(
      true,
    );
    expect(isStorageFull(new Error("asset fetch 500"))).toBe(false);
    expect(isStorageFull(undefined)).toBe(false);
  });
  it("keeps the default browser fetch bound to its global receiver", async () => {
    const ref = await reference("browser");
    vi.stubGlobal("fetch", function (this: unknown) {
      if (this !== globalThis) throw new TypeError("Illegal invocation");
      return Promise.resolve(new Response("browser"));
    });
    try {
      const store = new AssetStore("https://example.test/data/", {
        cache: null,
      });
      expect(new TextDecoder().decode(await store.get(ref))).toBe("browser");
    } finally {
      vi.unstubAllGlobals();
    }
  });
  it("deduplicates concurrent requests and reuses content across releases", async () => {
    const ref = await reference('{"test":1}');
    const cache = new MemoryCache();
    const urls: string[] = [];
    const fetcher: typeof fetch = (url) => {
      urls.push(url instanceof Request ? url.url : url.toString());
      return Promise.resolve(new Response('{"test":1}'));
    };
    const a = new AssetStore("https://example.test/next/data/", {
      cache,
      fetch: fetcher,
    });
    const values = await Promise.all([a.json(ref), a.json(ref)]);
    expect(values).toEqual([{ test: 1 }, { test: 1 }]);
    const b = new AssetStore("https://example.test/release/data/", {
      cache,
      fetch: fetcher,
    });
    expect(await b.json(ref)).toEqual({ test: 1 });
    expect(urls).toEqual([`https://example.test/next/data/${ref.path}`]);
  });

  it("repairs corrupt cached bytes and never caches a bad network response", async () => {
    const ref = await reference("good");
    const cache = new MemoryCache();
    cache.values.set(ref.sha256, new TextEncoder().encode("evil").buffer);
    const store = new AssetStore("https://example.test/data/", {
      cache,
      fetch: () => Promise.resolve(new Response("good")),
    });
    expect(new TextDecoder().decode(await store.get(ref))).toBe("good");
    const wrong = new AssetStore("https://example.test/data/", {
      cache: new MemoryCache(),
      fetch: () => Promise.resolve(new Response("bad!")),
    });
    await expect(wrong.get(ref)).rejects.toThrow(/hash/);
    await expect(wrong.get({ ...ref, bytes: 99 })).rejects.toThrow(/size/);
  });

  it("caps network concurrency and lets a failed request retry", async () => {
    let active = 0;
    let peak = 0;
    let fail = true;
    const data = new Map<string, string>();
    const fetcher: typeof fetch = async (url) => {
      active++;
      peak = Math.max(peak, active);
      await new Promise((r) => setTimeout(r, 2));
      active--;
      if (fail) {
        fail = false;
        return new Response("missing", { status: 404 });
      }
      const key = new URL(
        url instanceof Request ? url.url : url.toString(),
      ).pathname.replace("/data/", "");
      return new Response(data.get(key) ?? "data");
    };
    const store = new AssetStore("https://example.test/data/", {
      fetch: fetcher,
      concurrency: 2,
      cache: null,
    });
    const ref = await reference("data");
    await expect(store.get(ref)).rejects.toThrow(/404/);
    expect(new TextDecoder().decode(await store.get(ref))).toBe("data");
    const refs = await Promise.all(
      Array.from({ length: 8 }, async (_, i) => {
        const text = `data ${i}`;
        const r = await reference(text);
        data.set(r.path, text);
        return r;
      }),
    );
    await Promise.all(refs.map((r) => store.get(r)));
    expect(peak).toBe(2);
  });

  it("rejects escaping or mismatched paths before fetching", async () => {
    let requests = 0;
    const ref = await reference("data");
    const store = new AssetStore("https://example.test/data/", {
      fetch: () => {
        requests++;
        return Promise.resolve(new Response("data"));
      },
      cache: null,
    });
    for (const path of [
      "../secret",
      "https://other.test/a",
      "blobs/not-the-hash.png",
    ]) {
      await expect(store.get({ ...ref, path })).rejects.toThrow(/reference/);
    }
    expect(requests).toBe(0);
  });
});
