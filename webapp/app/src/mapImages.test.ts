import { expect, it, vi } from "vitest";
import { Texture, TextureSource } from "pixi.js";
import { AssetLoader } from "./assets.ts";
import { MapImages, type MapImage, type ReadyImage } from "./mapImages.ts";
import type { FileRef } from "./assetManifest.ts";
import { AssetNetworkError } from "./assetFetch.ts";

async function catalogFixture(fullWidth = 8, fullFailure?: Error) {
  const files = new Map<string, string>();
  const ref = async (body: string): Promise<FileRef> => {
    const bytes = new TextEncoder().encode(body);
    const sha256 = Array.from(
      new Uint8Array(await crypto.subtle.digest("SHA-256", bytes)),
      (value) => value.toString(16).padStart(2, "0"),
    ).join("");
    const value = { path: `blobs/${sha256}.json`, sha256, bytes: bytes.length };
    files.set(value.path, body);
    return value;
  };
  const liteImage = await ref("lite pixels"),
    fullImage = await ref("full pixels");
  const catalog = (file: FileRef, width = 8) =>
    JSON.stringify({
      schema: 1,
      status: "approved",
      entries: [
        {
          id: "unit/test",
          width,
          height: 6,
          density: 1,
          logicalAnchor: [4, 3],
          file,
        },
      ],
    });
  const pair = {
    lite: await ref(catalog(liteImage)),
    full: await ref(catalog(fullImage, fullWidth)),
  };
  const all = { "1": pair, "2": pair, "4": pair };
  const index = {
    schema: 1,
    rules: {},
    levels: {},
    packs: {
      units: all,
      terrain: all,
      interface: all,
      screens: { "1": pair, "2": pair },
    },
    pixelFonts: {
      "5": { image: liteImage, metrics: liteImage },
      "7": { image: liteImage, metrics: liteImage },
    },
    fonts: { Noto: liteImage },
    fontLicense: liteImage,
  };
  const requests: string[] = [];
  let released = false;
  const loader = new AssetLoader(index, "https://game.test/", {
    cache: null,
    deferUpgrades: true,
    fetch: (input) => {
      const path = new URL(
        input instanceof Request ? input.url : String(input),
      ).pathname.slice(1);
      requests.push(path);
      if (path === pair.full.path && !released)
        return Promise.reject(
          new Error("full metadata requested before playable frame"),
        );
      if (path === pair.full.path && fullFailure !== undefined)
        return Promise.reject(fullFailure);
      const body = files.get(path);
      return Promise.resolve(
        new Response(body ?? "missing", {
          status: body === undefined ? 404 : 200,
        }),
      );
    },
  });
  let changed: () => void = () => {};
  const upgraded = new Promise<void>((resolve) => {
    changed = resolve;
  });
  let report: (error: Error) => void = () => {};
  const failure = new Promise<Error>((resolve) => {
    report = resolve;
  });
  const errors: Error[] = [];
  const images = new MapImages(loader, changed, (error) => {
    errors.push(error);
    report(error);
  });
  return {
    images,
    loader,
    requests,
    pair,
    liteImage,
    fullImage,
    files,
    upgraded,
    failure,
    errors,
    release: () => {
      released = true;
      loader.enableUpgrades();
    },
  };
}

it("keeps a usable lite catalogue when optional full metadata loses its connection after startup", async () => {
  const f = await catalogFixture(8, new TypeError("connection lost"));
  f.release();
  try {
    const catalog = await f.images.catalog("units", 1);
    expect(catalog.get("unit/test")?.lite).toEqual(f.liteImage);
    expect(catalog.get("unit/test")?.full).toEqual(f.liteImage);
    expect(await f.images.catalog("units", 1)).toBe(catalog);
    expect(f.errors).toEqual([]);
  } finally {
    f.images.dispose();
  }
});

it.each(["geometry", "integrity"] as const)(
  "rejects an enabled full catalogue with invalid %s",
  async (fault) => {
    const f = await catalogFixture(fault === "geometry" ? 16 : 8);
    if (fault === "integrity") f.files.set(f.pair.full.path, "corrupt");
    f.release();
    try {
      await expect(f.images.catalog("units", 1)).rejects.toThrow(
        fault === "geometry" ? /geometry/ : /size/,
      );
    } finally {
      f.images.dispose();
    }
  },
);

it("cancels a disposed map's image transfer so the next required asset can use its network slot", async () => {
  const f = await catalogFixture();
  let arrived: () => void = () => {},
    releaseFetch: () => void = () => {},
    transportSignal: AbortSignal | null | undefined;
  const started = new Promise<void>((resolve) => {
    arrived = resolve;
  });
  const loader = new AssetLoader(f.loader.index, "https://game.test/", {
    cache: null,
    concurrency: 1,
    fetch: (input, options) => {
      const path = new URL(
        input instanceof Request ? input.url : String(input),
      ).pathname.slice(1);
      if (path !== f.liteImage.path)
        return Promise.resolve(new Response("full pixels"));
      transportSignal = options?.signal;
      return new Promise<Response>((resolve, reject) => {
        releaseFetch = () => {
          resolve(new Response("lite pixels"));
        };
        transportSignal?.addEventListener(
          "abort",
          () => {
            reject(new DOMException("Aborted image transfer", "AbortError"));
          },
          { once: true },
        );
        arrived();
      });
    },
  });
  const errors: Error[] = [];
  const images = new MapImages(
    loader,
    () => {},
    (error) => {
      errors.push(error);
    },
  );
  let next: Promise<ArrayBuffer> | undefined;
  try {
    images.get({
      id: "unit/blocked",
      width: 8,
      height: 6,
      density: 1,
      logicalAnchor: [4, 3],
      lite: f.liteImage,
      full: f.liteImage,
    });
    await started;
    next = loader.store.get(f.fullImage);
    images.dispose();
    await new Promise<void>((resolve) => {
      setImmediate(resolve);
    });
    expect(transportSignal?.aborted).toBe(true);
    expect(new TextDecoder().decode(await next)).toBe("full pixels");
    expect(errors).toEqual([]);
  } finally {
    images.dispose();
    releaseFetch();
    await Promise.allSettled(next === undefined ? [] : [next]);
    f.images.dispose();
  }
});

it("returns lite map geometry before loading full metadata and upgrades it after the playable frame", async () => {
  const f = await catalogFixture();
  try {
    const catalog = await f.images.catalog("units", 1);
    expect(catalog.get("unit/test")?.lite).toEqual(f.liteImage);
    expect(f.requests).not.toContain(f.pair.full.path);
    f.release();
    await f.upgraded;
    expect(catalog.get("unit/test")?.full).toEqual(f.fullImage);
    expect(f.errors).toEqual([]);
  } finally {
    f.images.dispose();
  }
});

it("refuses a deferred full catalogue with mismatched geometry and keeps lite usable", async () => {
  const f = await catalogFixture(16);
  try {
    const catalog = await f.images.catalog("units", 1);
    f.release();
    expect((await f.failure).message).toMatch(/geometry/);
    expect(catalog.get("unit/test")?.full).toEqual(f.liteImage);
  } finally {
    f.images.dispose();
  }
});

it("does not fetch deferred map metadata after the map has been disposed", async () => {
  const f = await catalogFixture();
  await f.images.catalog("units", 1);
  f.images.dispose();
  f.release();
  await new Promise<void>((resolve) => {
    setImmediate(resolve);
  });
  expect(f.requests).not.toContain(f.pair.full.path);
  expect(f.errors).toEqual([]);
});

it("cancels the disposed map's waiter while the shared startup gate remains held", async () => {
  const f = await catalogFixture();
  const waits = vi.spyOn(f.loader, "waitForUpgrades");
  await f.images.catalog("units", 1);
  const signal = waits.mock.calls[0]?.[0];
  const pending = waits.mock.results[0];
  try {
    expect(signal).toBeInstanceOf(AbortSignal);
    f.images.dispose();
    expect(signal?.aborted).toBe(true);
    if (pending?.type !== "return") throw new Error("missing gate waiter");
    await expect(pending.value).rejects.toThrow();
    expect(f.loader.upgradesDeferred).toBe(true);
    expect(f.requests).not.toContain(f.pair.full.path);
  } finally {
    f.images.dispose();
  }
});

it("keeps optional backdrop detail network failures quiet but reports integrity errors", async () => {
  const f = await catalogFixture();
  const errors: Error[] = [];
  const images = new MapImages(
    f.loader,
    () => {},
    (error) => {
      errors.push(error);
    },
  );
  const get = vi.spyOn(f.loader.store, "get");
  const image: MapImage = {
    id: "backdrop",
    width: 8,
    height: 6,
    density: 1,
    logicalAnchor: [0, 0],
    lite: f.liteImage,
    full: f.liteImage,
    optional: true,
  };
  try {
    get.mockRejectedValue(new AssetNetworkError(new TypeError("offline")));
    images.get(image);
    await new Promise<void>((resolve) => {
      setImmediate(resolve);
    });
    expect(errors).toEqual([]);
    get.mockRejectedValue(new Error("asset hash mismatch"));
    images.get({ ...image, lite: f.fullImage, full: f.fullImage });
    await new Promise<void>((resolve) => {
      setImmediate(resolve);
    });
    expect(errors.map((error) => error.message)).toEqual([
      "asset hash mismatch",
    ]);
  } finally {
    images.dispose();
    f.images.dispose();
  }
});

it("does not schedule full image decodes offline and retains an already decoded full image", () => {
  // This guard only consumes the loader's upgrade policy, not its catalogue.
  const loader = Object.create(AssetLoader.prototype) as AssetLoader;
  Object.defineProperty(loader, "store", { value: { networkAvailable: true } });
  const images = new MapImages(
    loader,
    () => {},
    () => {},
  );
  const image: MapImage = {
    id: "unit/test",
    width: 8,
    height: 6,
    density: 1,
    logicalAnchor: [4, 3],
    lite: { path: "lite", sha256: "lite", bytes: 4 },
    full: { path: "full", sha256: "full", bytes: 4 },
  };
  const ready: ReadyImage = {
    texture: Texture.EMPTY,
    width: 8,
    height: 6,
    density: 1,
    anchor: [4, 3],
  };
  const get = vi.spyOn(images.cache, "get").mockReturnValue(ready);
  const peek = vi.spyOn(images.cache, "peek").mockReturnValue(null);
  vi.stubGlobal("navigator", { onLine: false });
  try {
    expect(images.get(image)).toBe(ready);
    expect(get).toHaveBeenCalledTimes(1);
    expect(get.mock.calls[0]?.[0]).toMatch(/^lite\//);
    peek.mockReturnValue(ready);
    expect(images.get(image)).toBe(ready);
    expect(get.mock.calls[1]?.[0]).toMatch(/^full\//);
  } finally {
    vi.unstubAllGlobals();
    images.cache.dispose();
  }
});

it("decodes a background tile and its bleed copy from one fetch and one decode", async () => {
  const f = await catalogFixture();
  const get = vi.spyOn(f.loader.store, "get");
  let closed = 0;
  const decode = vi.fn(() =>
    Promise.resolve({
      width: 8,
      height: 6,
      close: () => {
        closed++;
      },
    }),
  );
  const context = {
    drawImage: () => {},
    getImageData: (_x: number, _y: number, width: number, height: number) => ({
      width,
      height,
      data: new Uint8ClampedArray(width * height * 4),
    }),
    createImageData: (width: number, height: number) => ({
      width,
      height,
      data: new Uint8ClampedArray(width * height * 4),
    }),
    putImageData: () => {},
  };
  vi.stubGlobal("createImageBitmap", decode);
  vi.stubGlobal("document", {
    createElement: () => ({ width: 0, height: 0, getContext: () => context }),
  });
  const from = vi
    .spyOn(Texture, "from")
    .mockImplementation(() => new Texture());
  const tile: MapImage = {
    id: "tile/test",
    width: 8,
    height: 6,
    density: 1,
    logicalAnchor: [4, 3],
    lite: f.liteImage,
    full: f.liteImage,
  };
  try {
    f.images.cache.begin();
    f.images.get({ ...tile, bleed: true });
    f.images.get(tile);
    f.images.cache.end();
    await vi.waitFor(() => {
      expect(f.images.peek(tile)).not.toBeNull();
      expect(f.images.peek({ ...tile, bleed: true })).not.toBeNull();
    });
    expect(decode).toHaveBeenCalledTimes(1);
    expect(get.mock.calls.filter(([ref]) => ref === f.liteImage)).toHaveLength(
      1,
    );
    // The bare tile's texture keeps the bitmap; the bleed copy leaves it open.
    expect(closed).toBe(0);
    expect(f.images.peek({ ...tile, bleed: true })?.anchor).toEqual([7, 6]);
  } finally {
    f.images.dispose();
    from.mockRestore();
    vi.unstubAllGlobals();
  }
  expect(closed).toBe(1);
});

it("holds an image decoded earlier as the stand-in of its successor", async () => {
  const f = await catalogFixture();
  vi.stubGlobal(
    "createImageBitmap",
    vi.fn(() => Promise.resolve({ width: 8, height: 6, close: () => {} })),
  );
  const from = vi
    .spyOn(Texture, "from")
    .mockImplementation(() => new Texture());
  const image: MapImage = {
    id: "tile/test",
    width: 8,
    height: 6,
    density: 1,
    logicalAnchor: [4, 3],
    lite: f.liteImage,
    full: f.liteImage,
  };
  try {
    f.images.cache.begin();
    f.images.get(image);
    f.images.cache.end();
    await vi.waitFor(() => {
      expect(f.images.peek(image)).not.toBeNull();
    });
    const ready = f.images.peek(image);
    if (ready === null) throw new Error("not decoded");
    f.images.cache.begin();
    expect(f.images.hold(ready)).toBe(true);
    expect(f.images.hold({ ...ready })).toBe(false);
    f.images.cache.end();
  } finally {
    f.images.dispose();
    from.mockRestore();
    vi.unstubAllGlobals();
  }
});

it("mipmaps map art so that a zoomed-out map shrinks it smoothly, but not pixel glyphs", async () => {
  const f = await catalogFixture();
  vi.stubGlobal(
    "createImageBitmap",
    vi.fn(() => Promise.resolve({ width: 8, height: 6, close: () => {} })),
  );
  const from = vi
    .spyOn(Texture, "from")
    .mockImplementation(() => new Texture({ source: new TextureSource() }));
  const tile: MapImage = {
    id: "tile/test",
    width: 8,
    height: 6,
    density: 1,
    logicalAnchor: [4, 3],
    lite: f.liteImage,
    full: f.liteImage,
  };
  const glyph: MapImage = { ...tile, id: "glyph/1", nearest: true };
  try {
    f.images.cache.begin();
    f.images.get(tile);
    f.images.get(glyph);
    f.images.cache.end();
    await vi.waitFor(() => {
      expect(f.images.peek(tile)).not.toBeNull();
      expect(f.images.peek(glyph)).not.toBeNull();
    });
    const art = f.images.peek(tile)?.texture.source,
      pixel = f.images.peek(glyph)?.texture.source;
    expect(art?.autoGenerateMipmaps).toBe(true);
    expect(art?.style.minFilter).toBe("linear");
    expect(art?.style.mipmapFilter).toBe("nearest");
    expect(pixel?.autoGenerateMipmaps).toBe(false);
    expect(pixel?.style.minFilter).toBe("nearest");
    // The budget counts the mipmaps' third.
    expect(f.images.cache.bytes).toBe(Math.ceil((8 * 6 * 16) / 3) + 8 * 6 * 4);
  } finally {
    f.images.dispose();
    from.mockRestore();
    vi.unstubAllGlobals();
  }
});
