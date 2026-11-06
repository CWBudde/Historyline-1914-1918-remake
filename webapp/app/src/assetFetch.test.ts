import { describe, expect, it } from "vitest";
import { AssetStore, type ContentCache } from "./assetFetch.ts";

async function reference(text: string) {
  const bytes = new TextEncoder().encode(text);
  const sum = new Uint8Array(await crypto.subtle.digest("SHA-256", bytes));
  const sha256 = Array.from(sum, (v) => v.toString(16).padStart(2, "0")).join(
    "",
  );
  return { path: `blobs/${sha256}.json`, sha256, bytes: bytes.length };
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
