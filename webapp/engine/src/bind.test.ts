import { describe, expect, it } from "vitest";
import { bind, bindingData, cell, coord, game, key, playable } from "./bind.ts";
import { parse } from "./level.ts";
import { decode, level, modes } from "./scenario.ts";
import { object } from "./decode.ts";
import { usesExtension } from "./playGas.ts";
import { Extended, Full } from "./testing/tables.ts";
import { listDir, readBytes, readJSON } from "./testing/files.ts";

const data = bindingData(
  readJSON("internal/data/units.json"),
  readJSON("internal/dosbind/records.json"),
);
const terrains = new Set(
  Object.keys(object(readJSON("internal/data/terrain.json"), "terrains")),
);

function shipped(month: string, mode: "cp" | "ec" | "2p" = "cp") {
  const files = new Map<string, unknown>();
  for (const name of ["map.json", "cp.json", "ec.json", "2p.json"])
    files.set(name, readJSON(`levels/${month}/${name}`));
  return level(decode(files, terrains), mode);
}

describe("binding", () => {
  it("numbers MOUSE factories by entrances rather than first pieces", () => {
    const b = bind(shipped("1914-12"), data, Full);
    for (const [record, entrance, first] of [
      [82, { line: 15, col: 3 }, "CP_armCar"],
      [83, { line: 37, col: 9 }, "CP_hArtillery"],
      [84, { line: 38, col: 1 }, "CP_mArtillery"],
    ] as const) {
      const h = b.holds.get(record);
      expect(h?.cells[0]).toBe(2 * (cell(b, entrance) ?? -1));
      expect(b.ids.get(h?.slots[0] ?? -1)).toBe(first);
    }
    expect(b.holds.has(85)).toBe(false);
  });
  it("keeps touching PARTS depots and their content separate", () => {
    const l = shipped("1916-02");
    const b = bind(l, data, Full);
    expect(
      b.holds
        .get(72)
        ?.slots.slice(0, 6)
        .map((u) => b.ids.get(u)),
    ).toEqual([
      "CP_fokkerEI",
      "CP_lArtillery",
      "CP_eliteInf",
      "CP_infantry",
      "CP_infantry",
      "CP_sappers",
    ]);
    expect(
      b.holds
        .get(73)
        ?.slots.slice(0, 6)
        .map((u) => b.ids.get(u)),
    ).toEqual([
      "CP_lArtillery",
      "CP_lArtillery",
      "CP_eliteInf",
      "CP_eliteInf",
      "CP_infantry",
      "CP_infantry",
    ]);
    expect(b.holds.has(74)).toBe(false);
  });
  it("refuses an incomplete building, an orphan piece and duplicate content", () => {
    const l = shipped("1916-02");
    const missing = structuredClone(l);
    const south = missing.lines[61]?.[9];
    if (south === undefined) throw new Error("missing fixture hex");
    south.terrain = "grass";
    expect(() => bind(missing, data, Full)).toThrow();
    const orphan = structuredClone(l);
    const corner = orphan.lines[0]?.[0];
    if (corner === undefined) throw new Error("missing fixture hex");
    corner.terrain = "cmpW";
    expect(() => bind(orphan, data, Full)).toThrow();
    const duplicate = structuredClone(l);
    const west = duplicate.lines[60]?.[9];
    if (west === undefined) throw new Error("missing fixture hex");
    west.content = { owner: 0, units: ["CP_infantry"] };
    expect(() => bind(duplicate, data, Full)).toThrow();
  });
  it("assigns non-Central-Powers countries to the Entente like FactionOf", () => {
    const d = bindingData(
      { test: { dos: 18, country: "Portugal", name: "Infantry" } },
      readJSON("internal/dosbind/records.json"),
    );
    const l = parse(
      '{"mapName":"test","mapNo":0,"player":"p","tiles":{"line1":[["grass","","test"],["grass","",""]],"line2":[["grass","",""]],"line3":[["grass","",""],["grass","",""]]}}',
    );
    expect(bind(l, d, Full).units.get(0)?.owner).toBe(1);
  });
  it("puts a second ship's rear on the next free water neighbour", () => {
    // Both destroyers' first water neighbour is line2[0]: the first ship's
    // rear takes it, the second's the next free one, line4[0] (Go's
    // TestShipRearsTakeFreeWater).
    const w = '["water","",""]';
    const l = parse(
      `{"mapName":"rears","mapNo":0,"player":"p","tiles":{"line1":[["water","","CP_destroyer"],${w},${w}],"line2":[${w},${w}],"line3":[["water","","EC_destroyer"],${w},${w}],"line4":[${w},${w}],"line5":[${w},${w},${w}]}}`,
    );
    const b = bind(l, data, Full);
    expect([...b.units.values()].map((u) => coord(b, u.cell))).toEqual([
      { line: 0, col: 0 },
      { line: 1, col: 0 },
      { line: 2, col: 0 },
      { line: 3, col: 0 },
    ]);
  });
  it("binds the legacy PULSE fixture to original cells and containers", () => {
    const b = bind(
      parse(readBytes("internal/level/testdata/pulse.json")),
      data,
      Full,
    );
    expect([b.width, b.height, b.winter]).toEqual([16, 24, false]);
    expect(cell(b, { line: 5, col: 2 })).toBe(37);
    expect(coord(b, 37)).toEqual({ line: 5, col: 2 });
    expect(coord(b, -1)).toBeNull();
    expect(coord(b, 16 * 24)).toBeNull();
    expect(cell(b, { line: 47, col: 0 })).toBeNull();
    expect(playable(b)).toEqual({ width: 8, lines: 46 });
    const g = game(b, Full);
    expect(g.units.size).toBe(b.units.size);
    expect(g.buildings.size).toBe(b.buildings.size);
    expect(b.holds.has(70)).toBe(true);
    expect(b.holds.has(71)).toBe(true);
    expect([...b.units.values()].some((u) => u.type === 18)).toBe(true);
    const bd = b.buildings.get(70);
    const copied = g.buildings.get(70);
    expect(copied).toEqual(bd);
    expect(copied).not.toBe(bd);
  });
  it("normalizes a terrain key once, preserving infrastructure order", () => {
    expect(
      key({
        terrain: "grass_w",
        infra: { ids: ["road36", "road36", "path36"], list: true },
        unit: "",
        content: null,
      }),
    ).toBe("grass+road36+path36");
  });
  it("binds all 156 shipped setups, their building entrances, held units and ship rears", () => {
    let count = 0;
    for (const month of listDir("levels")
      .filter((s) => /^19\d\d-\d\d$/.test(s))
      .sort()) {
      const files = new Map<string, unknown>();
      for (const name of ["map.json", "cp.json", "ec.json", "2p.json"])
        files.set(name, readJSON(`levels/${month}/${name}`));
      const f = decode(files, terrains);
      for (const mode of modes) {
        const l = level(f, mode);
        // Types are looked up in the extended tables, a superset (Go's dosbind); a
        // setup that places an extension unit plays on them.
        const b = bind(l, data, Extended);
        expect(b.winter).toBe(f.season === "winter");
        expect(game(b, usesExtension(b) ? Extended : Full).units.size).toBe(
          b.units.size,
        );
        for (const [i, u] of b.units) {
          if (b.ids.has(i)) continue;
          expect(b.units.get(i - 1)?.type, `${month}:${mode} rear ${i}`).toBe(
            u.type - 1,
          );
        }
        for (const [i, h] of b.holds) {
          expect(h.slots.slice(0, 7)).toEqual(h.slots.slice(7));
          if (i >= 70) expect(b.buildings.has(i)).toBe(true);
          else
            expect(h.cells[0]).toBe(
              2 * (b.units.get(h.occupant)?.cell ?? -1) + 1,
            );
        }
        count++;
      }
    }
    expect(count).toBe(156);
  });
});
