import { expect, it, vi } from "vitest";
import { paintingScale, UiAssets } from "./assets.ts";
import { assetDensity, type AssetLoader, type LoadedImage } from "../assets.ts";

function source(fail = false) {
  let release: (() => void) | undefined;
  const gate = new Promise<void>((resolve) => {
    release = resolve;
  });
  const calls: unknown[][] = [];
  const loader: Pick<AssetLoader, "progressive" | "upgradesAllowed"> = {
    upgradesAllowed: true,
    progressive: async (
      catalog,
      ids,
      zoom,
      dpr,
      onImage,
      _signal,
      sharp,
      liteOnly,
    ) => {
      calls.push([
        catalog,
        ids,
        zoom,
        dpr,
        ...(sharp === true ? [true] : []),
        ...(liteOnly === true ? ["lite only"] : []),
      ]);
      const image = (quality: "lite" | "full"): LoadedImage => ({
        quality,
        bytes: new TextEncoder().encode(quality).buffer,
        entry: {
          id: ids[0] ?? "",
          width: 8,
          height: 6,
          density: assetDensity(zoom, dpr),
          logicalAnchor: [4, 3],
          file: { path: "unused", sha256: quality, bytes: quality.length },
        },
      });
      await onImage(image("lite"));
      if (liteOnly === true) return new Map();
      await gate;
      if (fail) throw new Error("upgrade failed");
      const full = image("full");
      await onImage(full);
      return new Map([[full.entry.id, full]]);
    },
  };
  return {
    loader,
    calls,
    release: () => {
      release?.();
    },
  };
}

it("shares progressive DOM art, choosing density from UI scale and DPR", async () => {
  const f = source(),
    assets = new UiAssets(f.loader, () => 2);
  const image = assets.image("screens", "portrait/infantry", 1.25);
  expect(assets.image("screens", "portrait/infantry", 1.25)).toBe(image);
  await vi.waitFor(() => {
    expect(image.value.quality).toBe("lite");
  });
  const lite = image.value.url;
  if (lite === null) throw new Error("missing lite URL");
  expect(await (await fetch(lite)).text()).toBe("lite");
  f.release();
  await vi.waitFor(() => {
    expect(image.value.quality).toBe("full");
  });
  const full = image.value.url;
  if (full === null) throw new Error("missing full URL");
  expect(await (await fetch(full)).text()).toBe("full");
  expect(f.calls).toEqual([["screens", ["portrait/infantry"], 4, 1]]);
  assets.dispose();
  await expect(fetch(lite)).rejects.toThrow();
  await expect(fetch(full)).rejects.toThrow();
});

it("retains lite art after failed upgrades and ignores late disposed results", async () => {
  const failed = source(true),
    assets = new UiAssets(failed.loader, () => 1);
  const image = assets.image("screens", "portrait/infantry");
  await vi.waitFor(() => {
    expect(image.value.quality).toBe("lite");
  });
  failed.release();
  await vi.waitFor(() => {
    expect(image.value.error).toBe("upgrade failed");
  });
  expect(image.value.quality).toBe("lite");
  assets.dispose();

  const delayed = source(),
    disposed = new UiAssets(delayed.loader, () => 1);
  const late = disposed.image("screens", "portrait/infantry");
  await vi.waitFor(() => {
    expect(late.value.quality).toBe("lite");
  });
  disposed.dispose();
  delayed.release();
  await Promise.resolve();
  await Promise.resolve();
  expect(late.value.url).toBeNull();
  expect(disposed.image("screens", "portrait/infantry").value.url).toBeNull();
});

it("chooses a window-covering painting's art for its drawn size", () => {
  // 640×400 logical pixels: cover-scaled 3× at 1080p, 2.29× on a portrait phone, never below the UI scale.
  expect(paintingScale(1920, 1080, 1)).toBe(3);
  expect(paintingScale(412, 915, 1)).toBeCloseTo(2.2875);
  expect(paintingScale(320, 200, 1.5)).toBe(1.5);
  const f = source(),
    assets = new UiAssets(f.loader, () => 1);
  const sharp = assets.image("screens", "scene/options", 3, true);
  expect(assets.image("screens", "scene/options", 3)).not.toBe(sharp);
  expect(f.calls).toEqual([
    ["screens", ["scene/options"], 4, 1, true],
    ["screens", ["scene/options"], 4, 1],
  ]);
  assets.dispose();
});

it("brings a sharp painting's full art only when a page asks for it after full art is allowed", async () => {
  const f = source(),
    loader = f.loader as { upgradesAllowed: boolean },
    assets = new UiAssets(f.loader, () => 1);
  loader.upgradesAllowed = false;
  const image = assets.image("screens", "scene/options", 3, true);
  await vi.waitFor(() => {
    expect(image.value.quality).toBe("lite");
  });
  expect(assets.image("screens", "scene/options", 3, true)).toBe(image);
  expect(f.calls).toEqual([
    ["screens", ["scene/options"], 4, 1, true, "lite only"],
  ]);
  loader.upgradesAllowed = true;
  assets.image("screens", "scene/options", 3, true);
  assets.image("screens", "scene/options", 3, true);
  f.release();
  await vi.waitFor(() => {
    expect(image.value.quality).toBe("full");
  });
  expect(f.calls).toEqual([
    ["screens", ["scene/options"], 4, 1, true, "lite only"],
    ["screens", ["scene/options"], 4, 1, true],
  ]);
  assets.dispose();
});
