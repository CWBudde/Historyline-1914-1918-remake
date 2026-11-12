import { describe, expect, it } from "vitest";
import {
  exclusiveFiles,
  keptBytes,
  KeptRecords,
  type KeptRecord,
  type RecordStorage,
} from "./kept.ts";

const record = (key: string, files: Record<string, number>): KeptRecord => ({
  key,
  month: key.split(":")[0] ?? "",
  mode: key.split(":")[1] ?? "",
  label: key,
  keptAt: 1,
  files,
});
function memory(): RecordStorage & { map: Map<string, string> } {
  const map = new Map<string, string>();
  return {
    map,
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
  };
}

describe("exclusiveFiles", () => {
  const a = record("a:cp", { shared: 5, backdropA: 10, both: 3 });
  const b = record("b:cp", { shared: 5, backdropB: 20, both: 3 });
  it("excludes what another kept month needs", () => {
    expect(exclusiveFiles([a, b], "a:cp")).toEqual({
      hashes: ["backdropA"],
      bytes: 10,
    });
  });
  it("frees the shared art with the last month", () => {
    const last = exclusiveFiles([a], "a:cp");
    expect(last.hashes.sort()).toEqual(["backdropA", "both", "shared"]);
    expect(last.bytes).toBe(18);
  });
  it("counts every shared file once in the total", () => {
    expect(keptBytes([a, b])).toBe(5 + 10 + 20 + 3);
  });
});

describe("KeptRecords", () => {
  it("lists records, skips damaged ones and sorts by key", async () => {
    const storage = memory();
    const kept = new KeptRecords(storage);
    await kept.put(record("b:cp", { x: 1 }));
    await kept.put(record("a:cp", { x: 1 }));
    storage.map.set("bad:cp", "{not json");
    expect((await kept.list()).map((r) => r.key)).toEqual(["a:cp", "b:cp"]);
  });
  it("removes the record and forgets only exclusive files", async () => {
    const kept = new KeptRecords(memory());
    await kept.put(record("a:cp", { shared: 5, onlyA: 10 }));
    await kept.put(record("b:cp", { shared: 5, onlyB: 20 }));
    const forgotten: string[] = [];
    const freed = await kept.remove("a:cp", (hashes) => {
      forgotten.push(...hashes);
      return Promise.resolve();
    });
    expect(forgotten).toEqual(["onlyA"]);
    expect(freed).toBe(10);
    expect((await kept.list()).map((r) => r.key)).toEqual(["b:cp"]);
    await kept.remove("b:cp", (hashes) => {
      forgotten.push(...hashes);
      return Promise.resolve();
    });
    expect(forgotten.sort()).toEqual(["onlyA", "onlyB", "shared"]);
  });
  it("is empty without storage", async () => {
    const kept = new KeptRecords(null);
    await kept.put(record("a:cp", {}));
    expect(await kept.list()).toEqual([]);
  });
});
