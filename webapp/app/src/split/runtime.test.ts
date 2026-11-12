import { expect, it } from "vitest";
import { SplitRuntime } from "./runtime.ts";

function fixture(layout: "responsive" | "original" = "original") {
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
  const { runtime } = fixture("responsive");
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
