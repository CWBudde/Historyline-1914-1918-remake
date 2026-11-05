import { describe, expect, it } from "vitest";
import { decode, level, setup } from "./scenario.ts";
import { readJSON, listDir } from "./testing/files.ts";
import { object } from "./decode.ts";

const terrains = new Set(
  Object.keys(object(readJSON("internal/data/terrain.json"), "terrains")),
);
const map = {
  format: "historyline-scenario",
  version: 3,
  name: "Fixture",
  season: "winter",
  map: {
    line1: [
      ["hills", ""],
      ["grass", ""],
    ],
    line2: [["grass", ["road236"]]],
    line3: [
      ["grass", ""],
      ["grass", ""],
    ],
  },
  seasons: { line1: "s." },
};
const decodeFiles = (
  m: unknown = map,
  cp: unknown = { code: "ONE", units: [[2, 1, "CP_infantry"]] },
  others: Record<string, unknown> = {},
) =>
  decode(
    new Map(Object.entries({ "map.json": m, "cp.json": cp, ...others })),
    terrains,
  );

describe("scenario folders", () => {
  it("resolves aliases with their own codes, seasonal ground and independent level data", () => {
    const f = decodeFiles(map, undefined, {
      "2p.json": { from: "cp", code: "TWO" },
    });
    f.month = "1914-08";
    f.backdrop = "backdrop.jpg";
    expect(setup(f, "2p")?.code).toBe("TWO");
    const l = level(f, "2p");
    expect(l.mapName).toBe("TWO");
    expect(l.player).toBe("Central Powers");
    expect(l.lines[0]?.[0]?.terrain).toBe("rough_w");
    expect(l.lines[1]?.[0]?.unit).toBe("CP_infantry");
    expect(l.extra?.seasons.get("0,0")).toBe("summer");
    expect(l.extra?.backdrop).toBe("backdrop.jpg");
    const first = l.lines[1]?.[0];
    if (first !== undefined) first.infra.ids.push("road36");
    expect(level(f, "cp").lines[1]?.[0]?.infra.ids).toEqual(["road236"]);
  });
  it("places building content, credits and factory builds", () => {
    const f = decodeFiles(
      {
        ...map,
        season: "summer",
        map: {
          ...map.map,
          line1: [
            ["hqN", ""],
            ["grass", ""],
          ],
        },
      },
      {
        buildings: [
          { at: [1, 1], owner: "ec", credits: 5, units: ["EC_infantry"] },
        ],
        builds: ["CP_infantry"],
      },
    );
    const l = level(f, "cp");
    expect(l.mapName).toBe("Fixture");
    expect(l.lines[0]?.[0]?.content).toEqual({
      owner: 2,
      units: ["EC_infantry"],
    });
    expect(l.extra?.credits.get("0,0")).toBe(5);
    expect(l.extra?.builds).toEqual(["CP_infantry"]);
  });
  it.each([
    { ...map, version: 2 },
    { ...map, extra: 1 },
    { ...map, name: "" },
    { ...map, season: "" },
    { ...map, season: "rain" },
    { ...map, map: { line1: [["grass_w", ""]] }, seasons: {} },
    { ...map, map: { line1: [["unknown", ""]] }, seasons: {} },
    { ...map, seasons: { line1: "w" } },
    { ...map, seasons: { line1: "x." } },
    { ...map, seasons: { line8: ".." } },
    { ...map, seasons: { line01: ".." } },
    { ...map, seasons: { "line+1": ".." } },
    { ...map, map: { line1: [["rough", ""]] }, seasons: {} },
    { ...map, map: { line1: [["shelled", ""]] }, seasons: {} },
  ])("refuses invalid map %#", (m) => {
    expect(() => decodeFiles(m, {})).toThrow();
  });
  it.each([
    { units: [[0, 1, "CP_infantry"]] },
    { units: [[1, 3, "CP_infantry"]] },
    {
      units: [
        [1, 1, "CP_infantry"],
        [1, 1, "CP_sapper"],
      ],
    },
    { units: [[1, 1, ""]] },
    { buildings: [{ at: [1, 1], owner: "cp" }] },
    { from: "ec" },
    { from: "cp" },
    { from: "2p", units: [[1, 1, "CP_infantry"]] },
    { unknown: true },
  ])("refuses invalid setup %#", (cp) => {
    expect(() => decodeFiles(map, cp)).toThrow();
  });
  it("refuses shared codes, alias chains, unknown files and no setups", () => {
    expect(() =>
      decodeFiles(map, { code: "X" }, { "ec.json": { code: "X" } }),
    ).toThrow();
    expect(() =>
      decodeFiles(
        map,
        {},
        { "ec.json": { from: "cp" }, "2p.json": { from: "ec" } },
      ),
    ).toThrow();
    expect(() => decodeFiles(map, {}, { "other.json": {} })).toThrow();
    expect(() => decode(new Map([["map.json", map]]), terrains)).toThrow();
  });
  it("refuses a map unit on building content when making a level", () => {
    const f = decodeFiles(
      {
        ...map,
        map: {
          line1: [
            ["hqN", ""],
            ["grass", ""],
          ],
        },
        seasons: {},
      },
      { units: [[1, 1, "CP_infantry"]], buildings: [{ at: [1, 1] }] },
    );
    expect(() => level(f, "cp")).toThrow();
    expect(() => level(f, "ec")).toThrow();
  });
  it("decodes every shipped month and mode", () => {
    const months = listDir("levels")
      .filter((name) => /^19\d\d-\d\d$/.test(name))
      .sort();
    expect(months).toHaveLength(52);
    let count = 0;
    for (const month of months) {
      const files = new Map<string, unknown>();
      for (const name of listDir(`levels/${month}`).filter((name) =>
        /^(map|cp|ec|2p)\.json$/.test(name),
      ))
        files.set(name, readJSON(`levels/${month}/${name}`));
      const f = decode(files, terrains);
      for (const mode of ["cp", "ec", "2p"] as const) {
        const l = level(f, mode);
        expect(l.lines.length, `${month}:${mode}`).toBeGreaterThan(0);
        expect(l.player).toBe(mode === "ec" ? "Entente" : "Central Powers");
        count++;
      }
    }
    expect(count).toBe(156);
  });
});
