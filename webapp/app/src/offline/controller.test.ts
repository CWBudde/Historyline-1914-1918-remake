import { describe, expect, it } from "vitest";
import type { FileRef } from "../assetManifest.ts";
import type { OfflineMission } from "../offline.ts";
import {
  OfflineController,
  type OfflineServices,
  type StorageFigures,
} from "./controller.ts";
import { KeptRecords, type KeptRecord } from "./kept.ts";

const ref = (n: string, bytes: number): FileRef => ({
  path: `blobs/${n}`,
  sha256: n,
  bytes,
});
const mission = (month: string): OfflineMission => ({
  month,
  mode: "cp",
  next: null,
});
const shared = ref("shared", 100);

interface Fake extends OfflineServices {
  cache: Set<string>;
  persisted: number;
  gate: { release: () => void } | null;
  failOn: string | null;
  fullOn: string | null;
  kept: KeptRecords;
}
function fake(): Fake {
  const stored = new Map<string, string>();
  const records = new KeptRecords({
    keys: () => Promise.resolve([...stored.keys()]),
    get: (k) => Promise.resolve(stored.get(k)),
    put: (k, t) => {
      stored.set(k, t);
      return Promise.resolve();
    },
    delete: (k) => {
      stored.delete(k);
      return Promise.resolve();
    },
  });
  const f: Fake = {
    cache: new Set(),
    persisted: 0,
    gate: null,
    failOn: null,
    fullOn: null,
    kept: records,
    records: () => records.list(),
    plan(m) {
      const own = ref(`map-${m.month}`, 1000);
      const record: KeptRecord = {
        key: `${m.month}:${m.mode}`,
        month: m.month,
        mode: m.mode,
        label: m.month,
        keptAt: 1,
        files: { shared: 100, [own.sha256]: 1000 },
      };
      return Promise.resolve({ record, files: [shared, own] });
    },
    async fetch(files, signal, onFile) {
      for (const file of files) {
        signal.throwIfAborted();
        if (f.gate !== null) {
          await new Promise<void>((resolve) => {
            f.gate = { release: resolve };
          });
          signal.throwIfAborted();
        }
        if (file.sha256 === f.failOn) throw new Error("asset fetch 500");
        if (file.sha256 === f.fullOn)
          throw new DOMException(
            "The quota was exceeded",
            "QuotaExceededError",
          );
        f.cache.add(file.sha256);
        onFile(file);
      }
    },
    commit: (r) => records.put(r),
    remove: async (key) => {
      await records.remove(key, (hashes) => {
        for (const h of hashes) f.cache.delete(h);
        return Promise.resolve();
      });
    },
    discard: (hashes) => {
      for (const h of hashes) f.cache.delete(h);
      return Promise.resolve();
    },
    storage: () =>
      Promise.resolve<StorageFigures>({
        used: 5,
        quota: 50,
        persistent: false,
      }),
    persist: () => {
      f.persisted += 1;
      return Promise.resolve();
    },
  };
  return f;
}
const tick = () => new Promise((r) => setTimeout(r, 0));

describe("OfflineController", () => {
  it("offers the bytes a map needs, downloads it with progress and persists", async () => {
    const f = fake(),
      c = new OfflineController(f);
    await c.select(mission("1914-08"));
    expect(c.state.value).toMatchObject({ phase: "available", bytes: 1100 });
    const seen: number[] = [];
    c.state.subscribe((s) => seen.push(s.done));
    await c.download();
    expect(c.state.value.phase).toBe("ready");
    expect(seen).toContain(100);
    expect(seen).toContain(1100);
    expect(f.persisted).toBe(1);
    expect(c.kept.value.map((k) => k.key)).toEqual(["1914-08:cp"]);
    expect(c.storage.value.quota).toBe(50);
  });

  it("skips shared files another kept month holds", async () => {
    const f = fake(),
      c = new OfflineController(f);
    await c.select(mission("a"));
    await c.download();
    await c.select(mission("b"));
    expect(c.state.value).toMatchObject({ phase: "available", bytes: 1000 });
  });

  it("shows a kept map as ready without a plan", async () => {
    const f = fake(),
      c = new OfflineController(f);
    await c.select(mission("a"));
    await c.download();
    await c.select(mission("a"));
    expect(c.state.value.phase).toBe("ready");
  });

  it("cancels: no record, the arrived files are deleted", async () => {
    const f = fake(),
      c = new OfflineController(f);
    await c.select(mission("a"));
    f.gate = { release: () => undefined };
    const run = c.download();
    await tick();
    expect(c.state.value.phase).toBe("downloading");
    c.cancel();
    expect(c.state.value.phase).toBe("available");
    f.gate.release();
    await run;
    await tick();
    expect(c.state.value.phase).toBe("available");
    expect(c.kept.value).toEqual([]);
    expect(f.cache.size).toBe(0);
  });

  it("reports a failed file, records nothing and cleans up", async () => {
    const f = fake(),
      c = new OfflineController(f);
    await c.select(mission("a"));
    f.failOn = "map-a";
    await c.download();
    expect(c.state.value).toMatchObject({
      phase: "error",
      note: "asset fetch 500",
    });
    expect(f.cache.size).toBe(0);
    expect(await f.records()).toEqual([]);
    f.failOn = null;
    await c.select(mission("a"));
    await c.download();
    expect(c.state.value.phase).toBe("ready");
  });

  it("removes a month but keeps the shared art while another month is kept", async () => {
    const f = fake(),
      c = new OfflineController(f);
    await c.select(mission("a"));
    await c.download();
    await c.select(mission("b"));
    await c.download();
    expect(f.cache).toEqual(new Set(["shared", "map-a", "map-b"]));
    await c.remove("a:cp");
    expect(f.cache).toEqual(new Set(["shared", "map-b"]));
    await c.removeAll();
    expect(f.cache.size).toBe(0);
    expect(c.kept.value).toEqual([]);
  });

  it("offers the selected map again after removing it", async () => {
    const f = fake(),
      c = new OfflineController(f);
    await c.select(mission("a"));
    await c.download();
    await c.remove("a:cp");
    expect(c.state.value).toMatchObject({ phase: "available", bytes: 1100 });
  });

  it("ignores a stale plan after another selection", async () => {
    const f = fake(),
      c = new OfflineController(f);
    const first = c.select(mission("a"));
    const second = c.select(mission("b"));
    await Promise.all([first, second]);
    expect(c.state.value.key).toBe("b:cp");
  });
});

describe("OfflineController and the background keep", () => {
  const background = (month: string): KeptRecord => ({
    key: `${month}:cp`,
    month,
    mode: "cp",
    label: month,
    keptAt: 1,
    files: { shared: 100, [`map-${month}`]: 1000 },
  });

  it("pins what the player downloads", async () => {
    const f = fake(),
      c = new OfflineController(f);
    await c.select(mission("a"));
    await c.download();
    expect((await f.records())[0]?.pinned).toBe(true);
    expect(c.kept.value[0]?.pinned).toBe(true);
  });

  it("offers to keep a month the background keep stored, and pins it without a download", async () => {
    const f = fake(),
      c = new OfflineController(f);
    await f.kept.put(background("a"));
    await c.select(mission("a"));
    expect(c.state.value.phase).toBe("kept");
    expect(c.kept.value[0]?.pinned).toBe(false);
    await c.pin();
    expect(c.state.value.phase).toBe("ready");
    expect(f.persisted).toBe(1);
    expect(f.cache.size).toBe(0);
    expect((await f.records())[0]).toMatchObject({
      key: "a:cp",
      pinned: true,
    });
  });

  it("says the storage is full instead of failing, records nothing and cleans up", async () => {
    const f = fake(),
      c = new OfflineController(f);
    await c.select(mission("a"));
    f.fullOn = "map-a";
    await c.download();
    expect(c.state.value.phase).toBe("full");
    expect(c.state.value.note).toBe("");
    expect(f.cache.size).toBe(0);
    expect(await f.records()).toEqual([]);
    f.fullOn = null;
    await c.select(mission("a"));
    await c.download();
    expect(c.state.value.phase).toBe("ready");
  });
});
