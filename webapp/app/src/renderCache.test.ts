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
