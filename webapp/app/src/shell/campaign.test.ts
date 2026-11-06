import { readFile } from "node:fs/promises";
import { join } from "node:path";
import { expect, it } from "vitest";
import { scenario } from "@historyline/engine";
import { listDir, readJSON, repo } from "../../../engine/src/testing/files.ts";
import { AssetLoader } from "../assets.ts";
import type { AssetIndex } from "../assetManifest.ts";
import { loadCampaign } from "./campaign.ts";

const article = {
  label: "August 1914",
  headline: "War",
  deck: "Europe",
  caption: "Infantry",
  body: ["Report"],
  briefs: ["Bulletin"],
};
const unit = {
  name: "Infantry",
  country: "Germany",
  armour: 20,
  speed: 4,
  dos: 18,
  weapons: [
    null,
    {
      weapon: "rifle",
      attack: {
        power: 30,
        rangeMin: 1,
        rangeMax: 1,
        ground: true,
        water: true,
      },
    },
  ],
};
const units = {
  CP_infantry: unit,
  EC_infantry: { ...unit, country: "France" },
  CP_supplyCar: { ...unit, name: "Supply Car", dos: 22, weapons: [null] },
  CP_lArtillery: {
    ...unit,
    name: "Light Artillery",
    dos: 10,
    weapons: [
      {
        attack: {
          power: 45,
          rangeMin: 1,
          rangeMax: 3,
          ground: true,
          water: true,
        },
      },
      {
        attack: {
          power: 12,
          rangeMin: 2,
          rangeMax: 5,
          air: true,
          ground: true,
        },
      },
      { attack: { power: 0, rangeMin: 0, rangeMax: 99, air: true } },
      { weapon: "ornament" },
    ],
  },
};
function files(
  cp = "FIRST",
  ec = "OTHER",
  news: unknown = article,
): Map<string, unknown> {
  const result = new Map<string, unknown>([
    [
      "map.json",
      {
        format: "historyline-scenario",
        version: 3,
        name: "A field",
        season: "summer",
        map: {
          line1: [
            ["grass", ""],
            ["hq", ""],
            ["grass", ""],
          ],
        },
      },
    ],
    [
      "cp.json",
      {
        code: cp,
        units: [[1, 1, "CP_infantry"]],
        buildings: [{ at: [1, 2], units: ["CP_supplyCar"] }],
        builds: ["CP_lArtillery", "EC_infantry", "CP_unknown"],
      },
    ],
    ["ec.json", { from: "cp", code: ec }],
  ]);
  if (news !== undefined) result.set("news.json", news);
  return result;
}
function loader(months: Record<string, Map<string, unknown>>) {
  const index: AssetIndex = {
    schema: 1,
    rules: {},
    packs: {},
    pixelFonts: {},
    fonts: {},
    fontLicense: { path: "license.txt", sha256: "0".repeat(64), bytes: 0 },
    levels: Object.fromEntries(
      Object.keys(months).map((m) => [m, { files: {}, layers: {} }]),
    ),
  };
  const rules: Record<string, unknown> = {
    "terrain.json": { grass: {}, hq: {} },
    "units.json": units,
    "sprites.json": Object.fromEntries(
      Object.keys(units).map((id) => [id, { family: id.slice(3) }]),
    ),
    "notes.json": { infantry: "A foot soldier.", lArtillery: "A field gun." },
  };
  return {
    index,
    rule: (name: string): Promise<unknown> => Promise.resolve(rules[name]),
    levelFiles: (month: string): Promise<Map<string, unknown>> => {
      const f = months[month];
      if (f === undefined) throw new Error("missing fixture");
      return Promise.resolve(f);
    },
  };
}

it("links calendar-ordered side chains, skips empty codes and switches the selected month", async () => {
  const noNews = files("LAST", "");
  noNews.delete("news.json");
  const campaign = await loadCampaign(
    loader({
      "1914-10": noNews,
      "1914-08": files(),
      "1914-09": files("", "MIDDLE"),
    }),
  );
  expect(campaign.missions.map((m) => [m.key, m.next, m.issue])).toEqual([
    ["1914-08:cp", "1914-10:cp", 1],
    ["1914-10:cp", null, 0],
    ["1914-08:ec", "1914-09:ec", 1],
    ["1914-09:ec", null, 2],
  ]);
  expect(campaign.first(0).key).toBe("1914-08:cp");
  expect(campaign.first(1).key).toBe("1914-08:ec");
  expect(campaign.forSide("1914-08:cp", 1).key).toBe("1914-08:ec");
  expect(campaign.forSide("1914-10:cp", 1).key).toBe("1914-10:cp");
  expect(campaign.findCode("other")?.key).toBe("1914-08:ec");
  expect(campaign.findCode("missing")).toBeNull();
  expect(campaign.findCode("A FIELD")).toBeNull();
  expect(campaign.byKey.get("1914-08:cp")).toBe(campaign.first(0));
  expect(campaign.first(0)).toMatchObject({
    month: "1914-08",
    name: "A field",
    code: "FIRST",
    season: "summer",
    side: 0,
    mode: "cp",
    news: article,
    briefing:
      "Map FIRST, summer — command The Central Powers; capture the enemy HQ.",
  });
  expect(campaign.forSide("1914-10:cp", 0).news).toBeNull();
  expect(() => campaign.forSide("missing", 0)).toThrow();
});

it("introduces deployed, held and buildable side types once, with native per-slot powers and ranges", async () => {
  const campaign = await loadCampaign(
    loader({ "1914-08": files(), "1914-09": files("NEXT", "ENTENTE") }),
  );
  expect(campaign.first(0).newUnits).toEqual([
    {
      id: "CP_lArtillery",
      family: "lArtillery",
      name: "Light Artillery",
      country: "Germany",
      armour: 20,
      speed: 4,
      dos: 10,
      attack: "Air 12 · Ground 45 · Water 45",
      reach: "Range 1–5",
      note: "A field gun.",
    },
    {
      id: "CP_infantry",
      family: "infantry",
      name: "Infantry",
      country: "Germany",
      armour: 20,
      speed: 4,
      dos: 18,
      attack: "Ground 30 · Water 30",
      reach: "Range 1",
      note: "A foot soldier.",
    },
    {
      id: "CP_supplyCar",
      family: "supplyCar",
      name: "Supply Car",
      country: "Germany",
      armour: 20,
      speed: 4,
      dos: 22,
      attack: "No weapons",
      reach: "",
      note: "",
    },
  ]);
  expect(campaign.first(1).newUnits.map((c) => c.id)).toEqual(["EC_infantry"]);
  expect(campaign.byKey.get("1914-09:cp")?.newUnits).toEqual([]);
  expect(campaign.byKey.get("1914-09:ec")?.newUnits).toEqual([]);
  expect(campaign.first(0).file.month).toBe("1914-08");
  expect(scenario.level(campaign.first(0).file, "cp").mapName).toBe("FIRST");
});

it("keeps local setups out of side chains, preserves them for either side and prefers campaign codes", async () => {
  const first = files();
  first.set("2p.json", { from: "cp" });
  const second = files("NEXT", "ENTENTE");
  second.set("2p.json", { from: "cp", code: "LOCAL" });
  const campaign = await loadCampaign(
    loader({ "1914-08": first, "1914-09": second }),
  );
  expect(campaign.missions).toHaveLength(4);
  expect(campaign.byKey.size).toBe(6);
  const local = campaign.byKey.get("1914-08:2p");
  expect(local).toMatchObject({
    key: "1914-08:2p",
    side: 0,
    mode: "2p",
    code: "FIRST",
    next: null,
    newUnits: [],
    news: article,
    issue: 1,
    season: "summer",
  });
  expect(campaign.forSide("1914-08:2p", 0)).toBe(local);
  expect(campaign.forSide("1914-08:2p", 1)).toBe(local);
  expect(campaign.findCode("first")?.key).toBe("1914-08:cp");
  expect(campaign.findCode("local")?.key).toBe("1914-09:2p");
});

it.each([
  { ...article, extra: true },
  { ...article, label: "" },
  { ...article, body: [] },
  { ...article, briefs: null },
  { ...article, caption: 3 },
  { ...article, body: [5] },
])("rejects malformed newspaper metadata: %j", async (news) => {
  await expect(
    loadCampaign(loader({ "1914-08": files("FIRST", "OTHER", news) })),
  ).rejects.toThrow();
});
it("rejects malformed scenario, unit and sprite metadata and an empty campaign", async () => {
  const bad = files();
  bad.set("map.json", {});
  await expect(loadCampaign(loader({ "1914-08": bad }))).rejects.toThrow();
  await expect(loadCampaign(loader({}))).rejects.toThrow(/no campaign/i);
  await expect(
    loadCampaign(loader({ "1914-08": files("", "") })),
  ).rejects.toThrow(/no campaign/i);
  for (const [name, value] of [
    ["units.json", { CP_infantry: { ...unit, speed: "fast" } }],
    ["sprites.json", { CP_infantry: { family: 3 } }],
    ["notes.json", { infantry: 4 }],
  ] as const) {
    const source = loader({ "1914-08": files() });
    const original = source.rule;
    source.rule = async (key) => (key === name ? value : original(key));
    await expect(loadCampaign(source), name).rejects.toThrow();
  }
});
it("bounds concurrent month loads and preserves calendar order despite response order", async () => {
  const source = loader(
    Object.fromEntries(
      Array.from({ length: 12 }, (_, i) => [
        `1914-${String(i + 1).padStart(2, "0")}`,
        files(`CP${i}`, `EC${i}`),
      ]),
    ),
  );
  const original = source.levelFiles;
  let active = 0,
    peak = 0;
  source.levelFiles = async (month) => {
    active++;
    peak = Math.max(peak, active);
    await new Promise((resolve) =>
      setTimeout(resolve, month.endsWith("01") ? 10 : 1),
    );
    try {
      return await original(month);
    } finally {
      active--;
    }
  };
  const campaign = await loadCampaign(source);
  expect(peak).toBeGreaterThan(1);
  expect(peak).toBeLessThanOrEqual(4);
  expect(campaign.missions.slice(0, 12).map((m) => m.code)).toEqual(
    Array.from({ length: 12 }, (_, i) => `CP${i}`),
  );
});

it("loads all 104 shipped missions with both 52-month chains, codes, issues and unit introductions", async () => {
  const root = join(repo, "build/webapp/data");
  const fetcher: typeof fetch = async (url) => {
    const pathname = new URL(
      url instanceof Request ? url.url : url.toString(),
    ).pathname.replace("/data/", "");
    return new Response(await readFile(join(root, pathname)));
  };
  const source = await AssetLoader.open(
    "https://example.test/data/index.json",
    { fetch: fetcher, cache: null },
  );
  const campaign = await loadCampaign(source);
  const months = listDir("levels")
    .filter((m) => /^19\d\d-\d\d$/.test(m))
    .sort();
  expect(months).toHaveLength(52);
  expect(campaign.missions).toHaveLength(104);
  expect(campaign.byKey.size).toBe(156);
  for (const [i, month] of months.entries()) {
    const local = campaign.byKey.get(`${month}:2p`);
    if (local === undefined) throw Error(`missing local mission ${month}`);
    const raw = readJSON(`levels/${month}/2p.json`) as { code?: string };
    const cp = readJSON(`levels/${month}/cp.json`) as { code: string };
    expect(local).toMatchObject({
      month,
      mode: "2p",
      side: 0,
      code: raw.code || cp.code,
      next: null,
      newUnits: [],
      issue: i + 1,
    });
    expect(local.file).toBe(campaign.byKey.get(`${month}:cp`)?.file);
    expect(local.news).toBe(campaign.byKey.get(`${month}:cp`)?.news);
    expect(campaign.forSide(local.key, 1)).toBe(local);
    expect(() => scenario.level(local.file, "2p")).not.toThrow();
  }
  for (const side of [0, 1] as const) {
    const mode = side === 0 ? "cp" : "ec";
    const chain = campaign.missions.slice(side * 52, (side + 1) * 52);
    expect(chain.map((m) => m.month)).toEqual(months);
    const seen = new Set<string>();
    for (const [i, mission] of chain.entries()) {
      interface RawSetup {
        code: string;
        from?: string;
        units?: [number, number, string][];
        buildings?: { units?: string[] }[];
        builds?: string[];
      }
      const raw = readJSON(`levels/${mission.month}/${mode}.json`) as RawSetup;
      const content = raw.from
        ? (readJSON(`levels/${mission.month}/${raw.from}.json`) as RawSetup)
        : raw;
      const available = new Set([
        ...(content.units ?? []).map((u) => u[2]),
        ...(content.buildings ?? []).flatMap((b) => b.units ?? []),
        ...(content.builds ?? []),
      ]);
      const introduced = [...available].filter(
        (id) => id.startsWith(side === 0 ? "CP_" : "EC_") && !seen.has(id),
      );
      expect(mission.newUnits.map((c) => c.id).sort(), mission.key).toEqual(
        introduced.sort(),
      );
      expect(mission.code).toBe(raw.code);
      expect(campaign.findCode(raw.code.toLowerCase())).toBe(mission);
      expect(mission.next).toBe(
        i === 51 ? null : `${months[i + 1] ?? ""}:${mode}`,
      );
      expect(mission.issue).toBe(i + 1);
      expect(mission.news?.label).toContain(mission.month.slice(0, 4));
      expect(campaign.forSide(mission.key, side === 0 ? 1 : 0).key).toBe(
        `${mission.month}:${side === 0 ? "ec" : "cp"}`,
      );
      for (const c of mission.newUnits) {
        expect(c.id.startsWith(side === 0 ? "CP_" : "EC_")).toBe(true);
        expect(seen.has(c.id), c.id).toBe(false);
        seen.add(c.id);
        expect(c.family).not.toBe("");
        expect(c.note).not.toBe("");
      }
      expect(
        mission.newUnits.map(
          (c) => `${String(c.dos).padStart(2, "0")}:${c.id}`,
        ),
      ).toEqual(
        mission.newUnits
          .map((c) => `${String(c.dos).padStart(2, "0")}:${c.id}`)
          .sort(),
      );
    }
    expect(seen.size).toBe(side === 0 ? 36 : 39);
  }
  expect(campaign.first(0).newUnits.map((c) => c.id)).toEqual([
    "CP_armCar",
    "CP_fokkerEI",
    "CP_lArtillery",
    "CP_mArtillery",
    "CP_hArtillery",
    "CP_bunker",
    "CP_construction",
    "CP_eliteInf",
    "CP_infantry",
    "CP_cavalry",
    "CP_supplyCar",
  ]);
  expect(
    campaign.first(0).newUnits.find((c) => c.id === "CP_lArtillery"),
  ).toMatchObject({
    family: "light-artillery",
    attack: "Ground 45 · Water 45",
    reach: "Range 1–3",
    armour: 25,
    speed: 2,
  });
});
