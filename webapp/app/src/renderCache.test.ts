import { expect, it, vi } from "vitest";
import { RenderCache } from "./renderCache.ts";

it("bounds concurrent decodes, drops stale queued work and retains visible images", async () => {
  let finish: (() => void) | undefined;
  const disposed: string[] = [],
    started: string[] = [];
  const cache = new RenderCache<string>(4, 1, () => {});
  const load =
    (key: string, blocked = false) =>
    async () => {
      started.push(key);
      if (blocked)
        await new Promise<void>((resolve) => {
          finish = resolve;
        });
      return {
        value: key,
        bytes: 4,
        dispose: () => {
          disposed.push(key);
        },
      };
    };
  cache.begin();
  expect(cache.get("a", load("a", true))).toBeNull();
  cache.get("stale", load("stale"));
  cache.end();
  cache.begin();
  cache.get("b", load("b"));
  cache.end();
  finish?.();
  await vi.waitFor(() => {
    expect(cache.peek("b")).toBe("b");
  });
  expect(started).toEqual(["a", "b"]);
  expect(disposed).toEqual(["a"]);
  expect(cache.bytes).toBe(4);
  cache.dispose();
  expect(disposed).toEqual(["a", "b"]);
});

it("reports a frame ready only after every requested image has decoded", async () => {
  const cache = new RenderCache<string>(4, 1, () => {});
  let finish: () => void = () => {};
  cache.begin();
  cache.get("visible", async () => {
    await new Promise<void>((resolve) => {
      finish = resolve;
    });
    return { value: "visible", bytes: 4, dispose: () => {} };
  });
  cache.end();
  expect(cache.ready).toBe(false);
  finish();
  await vi.waitFor(() => {
    expect(cache.ready).toBe(true);
  });
  cache.dispose();
});

it("disposes late results and reports failed loads once", async () => {
  const errors: string[] = [],
    disposed: string[] = [];
  let finish: (() => void) | undefined;
  const cache = new RenderCache<string>(
    4,
    1,
    () => {},
    (e) => {
      errors.push(e.message);
    },
  );
  cache.begin();
  const fail = () => Promise.reject(new Error("decode failed"));
  cache.get("bad", fail);
  await vi.waitFor(() => {
    expect(errors).toEqual(["decode failed"]);
  });
  cache.get("bad", fail);
  cache.get("slow", async () => {
    await new Promise<void>((resolve) => {
      finish = resolve;
    });
    return {
      value: "slow",
      bytes: 4,
      dispose: () => {
        disposed.push("slow");
      },
    };
  });
  cache.dispose();
  finish?.();
  await vi.waitFor(() => {
    expect(disposed).toEqual(["slow"]);
  });
  expect(errors).toHaveLength(1);
  expect(cache.bytes).toBe(0);
});

it("handles optional upgrade failures separately from required image failures", async () => {
  const errors: string[] = [],
    optional: string[] = [];
  const cache = new RenderCache<string>(
    4,
    1,
    () => {},
    (error) => {
      errors.push(error.message);
    },
  );
  cache.begin();
  cache.get(
    "full",
    () => Promise.reject(new TypeError("offline")),
    (error) => {
      optional.push(error.message);
    },
  );
  await vi.waitFor(() => {
    expect(optional).toEqual(["offline"]);
  });
  expect(errors).toEqual([]);
  cache.get("lite", () => Promise.reject(new Error("required image missing")));
  await vi.waitFor(() => {
    expect(errors).toEqual(["required image missing"]);
  });
  cache.dispose();
});

it("holds a decoded stand-in through eviction and counts failed loads as settled", async () => {
  const disposed: string[] = [];
  const cache = new RenderCache<string>(4, 2, () => {});
  const load = (key: string) => () =>
    Promise.resolve({
      value: key,
      bytes: 4,
      dispose: () => {
        disposed.push(key);
      },
    });
  cache.begin();
  cache.get("old", load("old"));
  cache.end();
  await vi.waitFor(() => {
    expect(cache.peek("old")).toBe("old");
  });
  // The next frame wants a successor that has not decoded yet and keeps the old image meanwhile.
  cache.begin();
  cache.get("new", load("new"));
  expect(cache.hold("old")).toBe(true);
  expect(cache.hold("missing")).toBe(false);
  cache.end();
  await vi.waitFor(() => {
    expect(cache.peek("new")).toBe("new");
  });
  expect(disposed).toEqual([]);
  expect(cache.progress()).toEqual({ settled: 2, total: 2 });
  // Once no frame holds it, the budget evicts it.
  cache.begin();
  cache.get("new", load("new"));
  cache.end();
  expect(disposed).toEqual(["old"]);
  expect(cache.hold("old")).toBe(false);
  cache.begin();
  cache.get("bad", () => Promise.reject(new Error("decode failed")));
  cache.end();
  await vi.waitFor(() => {
    expect(cache.ready).toBe(true);
  });
  expect(cache.progress()).toEqual({ settled: 1, total: 1 });
  cache.dispose();
});

it("calls a frame complete while only optional upgrades are still decoding", async () => {
  const cache = new RenderCache<string>(64, 2, () => {});
  let finish: () => void = () => {};
  const slow = () =>
    new Promise<{ value: string; bytes: number; dispose: () => void }>(
      (resolve) => {
        finish = () => {
          resolve({ value: "full", bytes: 4, dispose: () => {} });
        };
      },
    );
  cache.begin();
  cache.get("lite", () =>
    Promise.resolve({ value: "lite", bytes: 4, dispose: () => {} }),
  );
  cache.get("full", slow, undefined, true);
  cache.end();
  await vi.waitFor(() => {
    expect(cache.ready).toBe(true);
  });
  expect(cache.progress()).toEqual({ settled: 1, total: 1 });
  // Asked for as required in the same frame, it counts.
  cache.begin();
  cache.get("full", slow, undefined, true);
  cache.get("full", slow);
  cache.end();
  expect(cache.ready).toBe(false);
  finish();
  await vi.waitFor(() => {
    expect(cache.ready).toBe(true);
  });
  cache.dispose();
});

it("starts what a frame shows before prefetches queued earlier", async () => {
  const started: string[] = [];
  const cache = new RenderCache<string>(64, 1, () => {});
  let finish: () => void = () => {};
  const load =
    (key: string, blocked = false) =>
    async () => {
      started.push(key);
      if (blocked)
        await new Promise<void>((resolve) => {
          finish = resolve;
        });
      return { value: key, bytes: 4, dispose: () => {} };
    };
  cache.begin();
  cache.get("busy", load("busy", true));
  cache.get("near", load("near"), undefined, true);
  cache.end();
  cache.begin();
  cache.get("busy", load("busy", true));
  cache.get("near", load("near"), undefined, true);
  cache.get("shown", load("shown"));
  cache.end();
  finish();
  await vi.waitFor(() => {
    expect(started).toEqual(["busy", "shown", "near"]);
  });
  cache.dispose();
});
