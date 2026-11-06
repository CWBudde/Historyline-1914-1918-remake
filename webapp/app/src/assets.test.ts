import { expect, it } from "vitest";
import { AssetLoader, assetDensity } from "./assets.ts";
import {
  parseAssetIndex,
  parseImageCatalog,
  type AssetIndex,
  type FileRef,
} from "./assetManifest.ts";

async function fixture() {
  const files = new Map<string, string>();
  const ref = async (text: string): Promise<FileRef> => {
    const bytes = new TextEncoder().encode(text);
    const hash = Array.from(
      new Uint8Array(await crypto.subtle.digest("SHA-256", bytes)),
      (v) => v.toString(16).padStart(2, "0"),
    ).join("");
    const path = `blobs/${hash}.json`;
    files.set(path, text);
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
