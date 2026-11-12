import { expect, it, vi } from "vitest";
import { SplitRuntime } from "./runtime.ts";
import type { Layout } from "../storage/store.ts";

function fixture(layout: Layout = "original") {
  const peer = { map: { ready: false } };
  const panels = [{ map: { ready: true } }, peer];
  const runtime = Object.create(SplitRuntime.prototype) as SplitRuntime;
  Object.assign(runtime, {
    settings: { layout },
    request: { side: 0 },
    panels,
    local: null,
    stopped: false,
  });
  return { runtime, peer };
}

it("holds optional upgrades until every visible split map is ready", () => {
  const { runtime, peer } = fixture();
  expect(runtime.ready).toBe(false);
  peer.map.ready = true;
  expect(runtime.ready).toBe(true);
});

it("does not wait for a hidden map or call an empty handover screen playable", () => {
  const { runtime } = fixture("single");
  expect(runtime.ready).toBe(true);
  Object.assign(runtime, {
    local: { mode: "hotseat", activeSide: 1, handover: 1 },
  });
  expect(runtime.ready).toBe(false);
  Object.assign(runtime, {
    local: { mode: "hotseat", activeSide: 1, handover: null },
  });
  expect(runtime.ready).toBe(false);
});

it("shows the split screen in the automatic layout on wide landscape windows only", () => {
  const { runtime } = fixture("auto");
  try {
    vi.stubGlobal("window", { innerWidth: 1920, innerHeight: 1080 });
    expect(runtime.original).toBe(true);
    vi.stubGlobal("window", { innerWidth: 412, innerHeight: 915 });
    expect(runtime.original).toBe(false);
    Object.assign(runtime, { settings: { layout: "original" } });
    expect(runtime.original).toBe(true);
    Object.assign(runtime, { settings: { layout: "single" } });
    vi.stubGlobal("window", { innerWidth: 1920, innerHeight: 1080 });
    expect(runtime.original).toBe(false);
  } finally {
    vi.unstubAllGlobals();
  }
});
