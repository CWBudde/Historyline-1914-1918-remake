import { describe, expect, it } from "vitest";
import { AssetStorageFullError } from "./assetFetch.ts";
import type { FileRef, Pyramid } from "./assetManifest.ts";
import {
  backdropRefs,
  bakedRefs,
  battleLoadError,
  keepPlanned,
  offlineMissions,
  saveDataRequested,
  type OfflinePlan,
} from "./offline.ts";
import { KeptRecords } from "./offline/kept.ts";

const ref = (n: string): FileRef => {
  const sha256 = n.repeat(64);
  return { path: `blobs/${sha256}.png`, sha256, bytes: 1 };
};

describe("offlineMissions", () => {
  const chain = (month: string, next: string | null) => ({
    month,
    mode: "cp",
    next,
  });
  const byKey = new Map([
    ["b:cp", chain("b", "c:cp")],
    ["c:cp", chain("c", "d:cp")],
    ["d:cp", chain("d", null)],
  ]);

  it("keeps the played month and the next two", () => {
    expect(
      offlineMissions(chain("a", "b:cp"), byKey).map((m) => m.month),
    ).toEqual(["a", "b", "c"]);
  });
  it("stops at the end of the chain and at unknown or repeated months", () => {
    expect(offlineMissions(chain("a", null), byKey)).toHaveLength(1);
    expect(offlineMissions(chain("a", "x:cp"), byKey)).toHaveLength(1);
    const loop = new Map([["a:cp", chain("a", "a:cp")]]);
    expect(offlineMissions(chain("a", "a:cp"), loop)).toHaveLength(1);
  });
});

describe("bakedRefs", () => {
  it("finds lite images and terrain IDs at any depth, without full images", () => {
    const lite = ref("a"),
      full = ref("b"),
      grass = ref("c");
    const baked = {
      scene: { Cells: [[{ Background: "grass", Overlays: ["unknown"] }]] },
      water: { w: { 1: { id: "w", lite, full } } },
    };
    const found = bakedRefs(baked, new Map([["grass", grass]]));
    expect(found.map((r) => r.sha256).sort()).toEqual(
      [lite, grass].map((r) => r.sha256).sort(),
    );
  });
});

describe("backdropRefs", () => {
  it("lists every tile of every mip, and nothing for a month without a backdrop", () => {
    const tile = (f: FileRef) => ({ file: f });
    const painted = {
      mips: [{ tiles: [tile(ref("a"))] }, { tiles: [tile(ref("b"))] }],
    } as unknown as Pyramid;
    expect(backdropRefs(painted)).toHaveLength(2);
    expect(backdropRefs(undefined)).toEqual([]);
  });
});

describe("battleLoadError", () => {
  it("names offline play only when the browser is offline", () => {
    const original = Error("asset fetch 500");
    expect(battleLoadError(original, true)).toBe(original);
    expect(battleLoadError(original, false).message).toContain("offline");
  });
});

describe("saveDataRequested", () => {
  it("is true only when the browser asks to save data", () => {
    expect(saveDataRequested({ connection: { saveData: true } })).toBe(true);
    expect(saveDataRequested({ connection: { saveData: false } })).toBe(false);
    expect(saveDataRequested({ connection: {} })).toBe(false);
    expect(saveDataRequested({})).toBe(false);
    expect(saveDataRequested(undefined)).toBe(false);
  });
});

describe("keepPlanned", () => {
  const mission = (month: string) => ({ month, mode: "cp", next: null });
  const shared = ref("5");
  const plan = (months: Record<string, FileRef[]>): OfflinePlan => ({
    shared: [shared],
    months: Object.entries(months).map(([month, refs]) => ({
      mission: mission(month),
      refs,
    })),
  });
  /** A content cache that holds room files and then refuses as a full quota does. */
  function store(room = Infinity) {
    const cache = new Set<string>();
    return {
      cache,
      keep(file: FileRef): Promise<void> {
        if (!cache.has(file.sha256) && cache.size >= room)
          return Promise.reject(new AssetStorageFullError(file.path));
        cache.add(file.sha256);
        return Promise.resolve();
      },
      forget(hashes: readonly string[]): Promise<void> {
        for (const h of hashes) cache.delete(h);
        return Promise.resolve();
      },
    };
  }
  function records() {
    const map = new Map<string, string>();
    return new KeptRecords({
      keys: () => Promise.resolve([...map.keys()]),
      get: (k) => Promise.resolve(map.get(k)),
      put: (k, t) => {
        map.set(k, t);
        return Promise.resolve();
      },
      delete: (k) => {
        map.delete(k);
        return Promise.resolve();
      },
    });
  }
  const files = (...refs: FileRef[]) =>
    Object.fromEntries(refs.map((r) => [r.sha256, r.bytes]));
  const record = (month: string, pinned?: boolean, ...refs: FileRef[]) => ({
    key: `${month}:cp`,
    month,
    mode: "cp",
    label: month,
    keptAt: 1,
    files: files(shared, ...refs),
    ...(pinned === undefined ? {} : { pinned }),
  });

  it("evicts the unpinned months outside the window once the keep completes", async () => {
    const s = store(),
      kept = records();
    const a = ref("a"),
      b = ref("b"),
      c = ref("c"),
      old = ref("d"),
      legacy = ref("e"),
      mine = ref("f"),
      sharedOld = ref("0");
    for (const r of [shared, old, legacy, mine, sharedOld])
      s.cache.add(r.sha256);
    await kept.put(record("1914-08", false, old, sharedOld));
    await kept.put(record("1914-09", undefined, legacy));
    await kept.put(record("1914-10", true, mine));
    const result = await keepPlanned(
      s,
      plan({ "1915-01": [a, sharedOld], "1915-02": [b], "1915-03": [c] }),
      new AbortController().signal,
      kept,
    );
    expect(result).toEqual({
      kept: ["1915-01:cp", "1915-02:cp", "1915-03:cp"],
      evicted: ["1914-08:cp", "1914-09:cp"],
      full: false,
    });
    expect((await kept.list()).map((r) => [r.key, r.pinned])).toEqual([
      ["1914-10:cp", true],
      ["1915-01:cp", false],
      ["1915-02:cp", false],
      ["1915-03:cp", false],
    ]);
    // The evicted months' own files go; a file the new month shares with one stays.
    expect(s.cache.has(old.sha256)).toBe(false);
    expect(s.cache.has(legacy.sha256)).toBe(false);
    for (const r of [shared, sharedOld, mine, a, b, c])
      expect(s.cache.has(r.sha256)).toBe(true);
  });

  it("keeps a month the player pinned pinned when it is also in the window", async () => {
    const s = store(),
      kept = records();
    const a = ref("a");
    await kept.put(record("1915-01", true, a));
    await keepPlanned(
      s,
      plan({ "1915-01": [a] }),
      new AbortController().signal,
      kept,
    );
    expect((await kept.list()).map((r) => [r.key, r.pinned])).toEqual([
      ["1915-01:cp", true],
    ]);
  });

  it("stops quietly on a full quota and records only the months that arrived complete", async () => {
    // Room for the shared art and two months' files.
    const s = store(3),
      kept = records();
    const a = ref("a"),
      b = ref("b"),
      c = ref("c"),
      old = ref("d");
    await kept.put(record("1914-08", false, old));
    s.cache.add(shared.sha256);
    s.cache.add(old.sha256);
    const result = await keepPlanned(
      s,
      plan({ "1915-01": [a], "1915-02": [b], "1915-03": [c] }),
      new AbortController().signal,
      kept,
    );
    expect(result.full).toBe(true);
    expect(result.kept).toEqual(["1915-01:cp"]);
    // The month that left the window still goes, which makes room for the next keep.
    expect(result.evicted).toEqual(["1914-08:cp"]);
    expect((await kept.list()).map((r) => r.key)).toEqual(["1915-01:cp"]);
    expect(s.cache.has(c.sha256)).toBe(false);
  });

  it("records nothing and evicts nothing when the battle is left", async () => {
    const s = store(),
      kept = records();
    await kept.put(record("1914-08", false, ref("d")));
    const abort = new AbortController();
    abort.abort();
    await expect(
      keepPlanned(s, plan({ "1915-01": [ref("a")] }), abort.signal, kept),
    ).rejects.toThrow();
    expect((await kept.list()).map((r) => r.key)).toEqual(["1914-08:cp"]);
  });
});
