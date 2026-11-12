import { expect, it, vi } from "vitest";
import { AssetLoader, assetDensity, unitDensity } from "./assets.ts";
import {
  parseAssetIndex,
  parseImageCatalog,
  type AssetIndex,
  type FileRef,
} from "./assetManifest.ts";

async function fixture() {
  const files = new Map<string, string | Uint8Array<ArrayBuffer>>();
  const ref = async (
    body: string | Uint8Array<ArrayBuffer>,
    ext = "json",
  ): Promise<FileRef> => {
    const bytes =
      typeof body === "string" ? new TextEncoder().encode(body) : body;
    const hash = Array.from(
      new Uint8Array(await crypto.subtle.digest("SHA-256", bytes)),
      (v) => v.toString(16).padStart(2, "0"),
    ).join("");
    const path = `blobs/${hash}.${ext}`;
    files.set(path, body);
    return { path, sha256: hash, bytes: bytes.length };
  };
  const lite = await ref("lite image"),
    full = await ref("full image");
  const image = (d: number, file: FileRef) => ({
    id: "unit/example",
    width: 8 * d,
    height: 6 * d,
    density: d,
    logicalAnchor: [4, 3],
    file,
  });
  const pair = async (d: number) => ({
    lite: await ref(
      JSON.stringify({
        schema: 1,
        entries: [image(d, lite)],
      }),
    ),
    full: await ref(
      JSON.stringify({
        schema: 1,
        entries: [image(d, full)],
      }),
    ),
  });
  const all = { "1": await pair(1), "2": await pair(2), "4": await pair(4) };
  const data = await ref("{}");
  const index: AssetIndex = {
    schema: 1,
    rules: { "terrain.json": data },
    levels: {
      "1914-08": {
        files: {
          "map.json": data,
          "news.json": data,
          "cp.json": data,
          "ec.json": data,
          "2p.json": data,
        },
        layers: { cp: data, ec: data, "2p": data },
      },
    },
    packs: {
      units: all,
      terrain: all,
      interface: all,
      screens: { "1": all["1"], "2": all["2"] },
    },
    pixelFonts: {
      "5": { image: data, metrics: data },
      "7": { image: data, metrics: data },
    },
    fonts: { Noto: data },
    fontLicense: data,
  };
  const requests: string[] = [];
  const fetcher: typeof fetch = (url) => {
    const path = new URL(
      url instanceof Request ? url.url : url.toString(),
    ).pathname.replace("/next/data/", "");
    requests.push(path);
    const value =
      path === "index.json" ? JSON.stringify(index) : files.get(path);
    return Promise.resolve(
      new Response(value ?? "missing", {
        status: value === undefined ? 404 : 200,
      }),
    );
  };
  return { index, files, ref, lite, full, requests, fetcher };
}

function backdrop(file: FileRef) {
  return {
    width: 1,
    height: 1,
    tileSize: 1,
    mips: [
      {
        scale: 1,
        width: 1,
        height: 1,
        tiles: [
          {
            x: 0,
            y: 0,
            width: 1,
            height: 1,
            sourceX: 0,
            sourceY: 0,
            sourceWidth: 1,
            sourceHeight: 1,
            file,
          },
        ],
      },
    ],
  };
}

it("loads only the selected level's verified backdrop metadata", async () => {
  const f = await fixture();
  const entry = f.index.levels["1914-08"];
  if (entry === undefined) throw new Error("missing fixture");
  const pyramid = backdrop(f.lite);
  const selected = await f.ref(JSON.stringify(pyramid));
  const other = await f.ref("unselected metadata stays unloaded");
  entry.backdropFile = selected;
  f.index.levels["1914-09"] = { ...entry, backdropFile: other };
  const loader = await AssetLoader.open(
    "https://example.test/next/data/index.json",
    { fetch: f.fetcher, cache: null },
  );
  expect(f.requests).toEqual(["index.json"]);
  expect(await loader.backdrop("1914-08")).toEqual(pyramid);
  expect(f.requests).toContain(selected.path);
  expect(f.requests).not.toContain(other.path);
  const invalid = await f.ref(JSON.stringify({ ...pyramid, mips: [] }));
  loader.index.levels["1914-09"] = { ...entry, backdropFile: invalid };
  await expect(loader.backdrop("1914-09")).rejects.toThrow(/pyramid/);
});

it("supports inline backdrops and levels without a painting", async () => {
  const f = await fixture();
  const entry = f.index.levels["1914-08"];
  if (entry === undefined) throw new Error("missing fixture");
  const pyramid = backdrop(f.lite);
  f.index.levels["1914-09"] = { ...entry, backdrop: pyramid };
  const loader = await AssetLoader.open(
    "https://example.test/next/data/index.json",
    { fetch: f.fetcher, cache: null },
  );
  expect(await loader.backdrop("1914-09")).toEqual(pyramid);
  expect(await loader.backdrop("1914-08")).toBeUndefined();
  expect(f.requests).toEqual(["index.json"]);
  entry.backdrop = pyramid;
  entry.backdropFile = f.lite;
  expect(() => parseAssetIndex(f.index)).toThrow(/backdrop/);
  delete entry.backdrop;
  entry.backdropFile = { ...f.lite, path: "../outside" };
  expect(() => parseAssetIndex(f.index)).toThrow(/reference/);
});

it("selects the smallest density without magnification, capped at four", () => {
  for (const [zoom, dpr, want] of [
    [0.5, 2, 1],
    [1.001, 1, 2],
    [1, 2, 2],
    [1.01, 2, 4],
    [4, 3, 4],
  ])
    expect(assetDensity(zoom ?? 0, dpr ?? 0)).toBe(want);
  for (const bad of [0, -1, NaN, Infinity])
    expect(() => assetDensity(bad, 1)).toThrow();
});

it("loads manifest URLs, emits lite first, then full at zoom times DPR", async () => {
  const f = await fixture();
  const loader = await AssetLoader.open(
    "https://example.test/next/data/index.json",
    { fetch: f.fetcher, cache: null },
  );
  const stages: string[] = [];
  const result = await loader.progressive(
    "units",
    ["unit/example"],
    1.2,
    2,
    ({ quality, entry, bytes }) => {
      stages.push(quality);
      expect(entry.density).toBe(4);
      expect(new TextDecoder().decode(bytes)).toBe(
        quality === "lite" ? "lite image" : "full image",
      );
    },
  );
  expect(stages).toEqual(["lite", "full"]);
  expect(result.get("unit/example")?.entry.file).toEqual(f.full);
  expect(f.requests.indexOf(f.lite.path)).toBeLessThan(
    f.requests.indexOf(f.full.path),
  );
  const screen = await loader.progressive(
    "screens",
    ["unit/example"],
    4,
    3,
    () => {},
  );
  expect(screen.get("unit/example")?.entry.density).toBe(2);
  expect(await loader.rule("terrain.json")).toEqual({});
  expect((await loader.levelFiles("1914-08")).size).toBe(5);
  expect(await loader.baked("1914-08", "cp")).toEqual({});
});

it("keeps lite available on a failed full upgrade and rejects missing IDs", async () => {
  const f = await fixture();
  f.files.delete(f.full.path);
  const loader = new AssetLoader(f.index, "https://example.test/next/data/", {
    fetch: f.fetcher,
    cache: null,
  });
  const stages: string[] = [];
  await expect(
    loader.progressive("units", ["unit/example"], 1, 1, ({ quality }) => {
      stages.push(quality);
    }),
  ).rejects.toThrow(/404/);
  expect(stages).toEqual(["lite"]);
  await expect(
    loader.progressive("units", ["missing"], 1, 1, () => {}),
  ).rejects.toThrow(/missing/);
});

it("uses the complete lite art offline without requesting optional full upgrades", async () => {
  const f = await fixture();
  const loader = new AssetLoader(f.index, "https://example.test/next/data/", {
    fetch: f.fetcher,
    cache: null,
  });
  vi.stubGlobal("navigator", { onLine: false });
  try {
    const stages: string[] = [];
    await loader.progressive("units", ["unit/example"], 1, 1, (image) => {
      stages.push(image.quality);
    });
    expect(stages).toEqual(["lite"]);
    expect(f.requests).not.toContain(f.full.path);
    expect(loader.upgradesAllowed).toBe(false);
  } finally {
    vi.unstubAllGlobals();
  }
});

it("holds optional upgrades until the first playable frame while delivering lite art immediately", async () => {
  const f = await fixture();
  const loader = new AssetLoader(f.index, "https://example.test/next/data/", {
    fetch: f.fetcher,
    cache: null,
    deferUpgrades: true,
  });
  const stages: string[] = [];
  const work = loader.progressive("units", ["unit/example"], 1, 1, (image) => {
    stages.push(image.quality);
  });
  await vi.waitFor(() => {
    expect(stages).toEqual(["lite"]);
  });
  expect(f.requests).not.toContain(f.full.path);
  expect(loader.upgradesAllowed).toBe(false);
  expect(loader.enableUpgrades()).toBe(true);
  await work;
  expect(stages).toEqual(["lite", "full"]);
  expect(loader.enableUpgrades()).toBe(false);
});

it("starts held high-DPR UI images with 1x lite pixels and preserves their logical geometry at 4x", async () => {
  const f = await fixture();
  const loader = new AssetLoader(f.index, "https://example.test/next/data/", {
    fetch: f.fetcher,
    cache: null,
    deferUpgrades: true,
  });
  const stages: { quality: string; density: number; width: number }[] = [];
  const work = loader.progressive(
    "units",
    ["unit/example"],
    1.2,
    2,
    ({ quality, entry }) => {
      stages.push({ quality, density: entry.density, width: entry.width });
    },
  );
  try {
    await vi.waitFor(() => {
      expect(stages).toHaveLength(1);
    });
    expect(stages).toEqual([{ quality: "lite", density: 1, width: 8 }]);
    expect(f.requests).not.toContain(f.index.packs.units?.["4"]?.lite.path);
    loader.enableUpgrades();
    await work;
    expect(stages).toEqual([
      { quality: "lite", density: 1, width: 8 },
      { quality: "full", density: 4, width: 32 },
    ]);
  } finally {
    loader.enableUpgrades();
    await work;
  }
});

it("brings a held sharp painting's lite pixels at the drawn density", async () => {
  const f = await fixture();
  const loader = new AssetLoader(f.index, "https://example.test/next/data/", {
    fetch: f.fetcher,
    cache: null,
    deferUpgrades: true,
  });
  const stages: { quality: string; density: number }[] = [];
  const work = loader.progressive(
    "units",
    ["unit/example"],
    1.2,
    2,
    ({ quality, entry }) => {
      stages.push({ quality, density: entry.density });
    },
    undefined,
    true,
  );
  try {
    await vi.waitFor(() => {
      expect(stages).toHaveLength(1);
    });
    expect(stages).toEqual([{ quality: "lite", density: 4 }]);
  } finally {
    loader.enableUpgrades();
    await work;
  }
});

it("rejects a cross-density upgrade that changes logical size before fetching full pixels", async () => {
  const f = await fixture();
  const pair = f.index.packs.units?.["4"];
  if (pair === undefined) throw new Error("missing fixture");
  const body = f.files.get(pair.full.path);
  if (typeof body !== "string") throw new Error("missing catalogue");
  const catalog = JSON.parse(body) as { entries: { width: number }[] };
  const entry = catalog.entries[0];
  if (entry === undefined) throw new Error("missing entry");
  entry.width = 40;
  pair.full = await f.ref(JSON.stringify(catalog));
  const loader = new AssetLoader(f.index, "https://example.test/next/data/", {
    fetch: f.fetcher,
    cache: null,
    deferUpgrades: true,
  });
  const stages: number[] = [];
  const work = loader.progressive(
    "units",
    ["unit/example"],
    1,
    4,
    ({ entry }) => {
      stages.push(entry.density);
    },
  );
  await vi.waitFor(() => {
    expect(stages).toHaveLength(1);
  });
  loader.enableUpgrades();
  await expect(work).rejects.toThrow(/geometry/);
  expect(stages).toEqual([1]);
  expect(f.requests).not.toContain(f.full.path);
});

it("cancels a progressive image's held upgrade without opening the shared startup gate", async () => {
  const f = await fixture();
  const loader = new AssetLoader(f.index, "https://example.test/next/data/", {
    fetch: f.fetcher,
    cache: null,
    deferUpgrades: true,
  });
  const waits = vi.spyOn(loader, "waitForUpgrades");
  const abort = new AbortController();
  const stages: string[] = [];
  const work = loader.progressive(
    "units",
    ["unit/example"],
    1,
    1,
    ({ quality }) => {
      stages.push(quality);
    },
    abort.signal,
  );
  try {
    await vi.waitFor(() => {
      expect(stages).toEqual(["lite"]);
    });
    expect(waits).toHaveBeenCalledWith(abort.signal);
    abort.abort();
    await expect(work).rejects.toThrow();
    expect(loader.upgradesDeferred).toBe(true);
    expect(f.requests).not.toContain(f.full.path);
  } finally {
    loader.enableUpgrades();
    await work.catch(() => {});
  }
});

it("keeps verified lite art when a stale online hint permits an upgrade during disconnection", async () => {
  const f = await fixture();
  const loader = new AssetLoader(f.index, "https://example.test/next/data/", {
    cache: null,
    fetch: (input, options) =>
      (input instanceof Request ? input.url : input.toString()).endsWith(
        f.full.path,
      )
        ? Promise.reject(new TypeError("offline"))
        : f.fetcher(input, options),
  });
  const stages: string[] = [];
  const result = await loader.progressive(
    "units",
    ["unit/example"],
    1,
    1,
    (image) => {
      stages.push(image.quality);
    },
  );
  expect(stages).toEqual(["lite"]);
  expect(result.get("unit/example")?.quality).toBe("lite");
  expect(loader.upgradesAllowed).toBe(false);
});

it("keeps lite art if the deferred full catalogue cannot be fetched after disconnection", async () => {
  const f = await fixture();
  const full = f.index.packs.units?.["1"]?.full;
  if (full === undefined) throw new Error("missing fixture");
  const loader = new AssetLoader(f.index, "https://example.test/next/data/", {
    cache: null,
    fetch: (input, options) =>
      (input instanceof Request ? input.url : String(input)).endsWith(full.path)
        ? Promise.reject(new TypeError("offline"))
        : f.fetcher(input, options),
  });
  const stages: string[] = [];
  const result = await loader.progressive(
    "units",
    ["unit/example"],
    1,
    1,
    ({ quality }) => {
      stages.push(quality);
    },
  );
  expect(stages).toEqual(["lite"]);
  expect(result.get("unit/example")?.quality).toBe("lite");
  expect(loader.upgradesAllowed).toBe(false);
});

it("does not hide an upgrade integrity failure after an unrelated network failure", async () => {
  const f = await fixture();
  const unrelated = await f.ref("unrelated");
  let start: () => void = () => {},
    finish: (bytes: ArrayBuffer) => void = () => {};
  const started = new Promise<void>((resolve) => {
    start = resolve;
  });
  const body = new Promise<ArrayBuffer>((resolve) => {
    finish = resolve;
  });
  const loader = new AssetLoader(f.index, "https://example.test/next/data/", {
    cache: null,
    fetch: (input, options) => {
      const url = input instanceof Request ? input.url : input.toString();
      if (url.endsWith(unrelated.path))
        return Promise.reject(new TypeError("offline"));
      if (url.endsWith(f.full.path)) {
        const response = new Response("bad image");
        Object.defineProperty(response, "arrayBuffer", {
          value: () => {
            start();
            return body;
          },
        });
        return Promise.resolve(response);
      }
      return f.fetcher(input, options);
    },
  });
  const upgrading = loader.progressive(
    "units",
    ["unit/example"],
    1,
    1,
    () => {},
  );
  const failed = expect(upgrading).rejects.toThrow(/hash/);
  await started;
  await expect(loader.store.get(unrelated)).rejects.toThrow("offline");
  expect(loader.upgradesAllowed).toBe(false);
  finish(new TextEncoder().encode("bad image!").buffer);
  await failed;
});

it("rejects schema gaps, hash paths, missing mip tiles and catalog geometry", async () => {
  const f = await fixture();
  expect(parseAssetIndex(f.index)).toEqual(f.index);
  for (const mutate of [
    (i: AssetIndex) => {
      i.schema = 2 as 1;
    },
    (i: AssetIndex) => {
      delete i.packs.units;
    },
    (i: AssetIndex) => {
      i.fontLicense.path = "../escape";
    },
    (i: AssetIndex) => {
      const l = i.levels["1914-08"];
      if (l)
        l.backdrop = {
          width: 513,
          height: 12,
          tileSize: 512,
          mips: [{ scale: 1, width: 513, height: 12, tiles: [] }],
        };
    },
  ]) {
    const copy = structuredClone(f.index);
    mutate(copy);
    expect(() => parseAssetIndex(copy)).toThrow();
  }
  expect(() =>
    parseImageCatalog(
      {
        schema: 1,
        entries: [
          {
            id: "a",
            density: 2,
            width: 3,
            height: 4,
            logicalAnchor: [0, 0],
            file: f.lite,
          },
        ],
      },
      2,
    ),
  ).toThrow(/geometry/);
});

it("draws units from 2x art at least, as the desktop game does", () => {
  expect([1, 2, 4].map(unitDensity)).toEqual([2, 2, 4]);
});
