import { describe, expect, it } from "vitest";
import type { FileRef, Pyramid } from "./assetManifest.ts";
import {
  backdropRefs,
  bakedRefs,
  battleLoadError,
  offlineMissions,
  paintingRefs,
} from "./offline.ts";

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

describe("paintingRefs", () => {
  it("picks the menu paintings among the screens", () => {
    const screens = new Map([
      ["backdrop/western-front", ref("w")],
      ["scene/options", ref("o")],
      ["portrait/infantry", ref("p")],
      ["news/1914-08", ref("n")],
    ]);
    expect(paintingRefs(screens)).toEqual([ref("w"), ref("o")]);
  });
});

describe("battleLoadError", () => {
  it("names offline play only when the browser is offline", () => {
    const original = Error("asset fetch 500");
    expect(battleLoadError(original, true)).toBe(original);
    expect(battleLoadError(original, false).message).toContain("offline");
  });
});
