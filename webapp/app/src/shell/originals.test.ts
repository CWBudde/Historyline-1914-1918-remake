// The menu's original-game page: importing and removing the files through the host, and choosing an original map.
import { expect, it } from "vitest";
import { scenario } from "@historyline/engine";
import { defaultSettings } from "../storage/store.ts";
import { setOriginals, type Campaign, type Mission } from "./campaign.ts";
import { ShellController, type OriginalHooks } from "./controller.ts";

const file: scenario.File = {
  month: "",
  name: "Field",
  season: "summer",
  seasons: new Map(),
  backdrop: "",
  map: [],
  setups: new Map(),
};
function mission(key: string, side: 0 | 1, code: string): Mission {
  const [month = "", mode = "cp"] = key.split(":");
  return {
    key,
    month,
    mode: mode as scenario.Mode,
    side,
    code,
    name: "Field",
    season: "summer",
    briefing: "",
    next: null,
    news: null,
    issue: 0,
    newUnits: [],
    file,
  };
}
function campaign(): Campaign {
  const missions = [mission("1914-08:cp", 0, "PULSE")];
  const byKey = new Map(missions.map((m) => [m.key, m]));
  return {
    missions,
    byKey,
    forSide: (key) => byKey.get(key) ?? missions[0] ?? mission(key, 0, ""),
    findCode: () => null,
    first: () => missions[0] ?? mission("", 0, ""),
  };
}
const hooks = {
  start: () => Promise.resolve(),
  pause: () => undefined,
  discard: () => undefined,
  capture: () => ({ map: "", side: 0 as const, data: "" }),
  canSave: () => false,
  validate: () => mission("1914-08:cp", 0, "PULSE"),
  settings: () => undefined,
};

it("imports original maps, selects one and resets the map when they are removed", async () => {
  const c = campaign();
  const original = { ...mission("original/01:cp", 0, "CIVIL") };
  const host: OriginalHooks = {
    status: () => ({ maps: c.originals?.length ?? 0, art: true, skipped: [] }),
    import: (picked) => {
      expect(picked).toHaveLength(1);
      setOriginals(c, [
        {
          ...original,
          original: {
            stem: "01",
            map: {} as never,
            view: {} as never,
          },
        },
      ]);
      return Promise.resolve("1 original map imported.");
    },
    remove: () => {
      setOriginals(c, []);
      return Promise.resolve();
    },
  };
  const s = new ShellController(c, null, defaultSettings(), {
    ...hooks,
    originals: host,
  });
  expect(s.originals).toBe(host);
  s.open("original");
  expect(s.page).toBe("original");
  await s.importOriginals([{ path: "CODES.DAT", bytes: new Uint8Array() }]);
  expect(s.note).toBe("1 original map imported.");
  expect(c.byKey.has("original/01:cp")).toBe(true);
  s.selectOriginal("original/01:cp");
  expect([s.page, s.mission.key, s.note]).toEqual([
    "main",
    "original/01:cp",
    "Original map 01 (CIVIL) selected.",
  ]);
  s.open("original");
  await s.removeOriginals();
  expect(c.byKey.has("original/01:cp")).toBe(false);
  expect(s.mission.key).toBe("1914-08:cp");
  s.back();
  expect(s.page).toBe("extras");
});

it("reports a failed import and keeps the menu usable", async () => {
  const c = campaign();
  const s = new ShellController(c, null, defaultSettings(), {
    ...hooks,
    originals: {
      status: () => ({ maps: 0, art: false, skipped: [] }),
      import: () => Promise.reject(new Error("CODES.DAT is missing")),
      remove: () => Promise.resolve(),
    },
  });
  s.open("original");
  await s.importOriginals([]);
  expect([s.note, s.busy, s.page]).toEqual([
    "CODES.DAT is missing",
    false,
    "original",
  ]);
});

it("offers no original page without a host", () => {
  const s = new ShellController(campaign(), null, defaultSettings(), hooks);
  expect(s.originals).toBeNull();
});
