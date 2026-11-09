// Twin of internal/dosplay/gas_test.go.

import { describe, expect, it } from "vitest";
import * as play from "./play.ts";
import { buildingKind, kindFactory } from "./capture.ts";
import { gasCharges, gasTurnsDefault } from "./gas.ts";
import { type Coord, around } from "./hexgrid.ts";
import { type Hex, type Level, coordKey } from "./level.ts";
import { at } from "./goruntime.ts";
import {
  Builds,
  Extended,
  ExtendedTables,
  Full,
  FullTables,
} from "./testing/tables.ts";
import { playData, shippedLevel } from "./testing/play.ts";

function grassLevel(units: readonly [Coord, string][]): Level {
  const lv: Level = {
    mapName: "grass",
    mapNo: 0,
    player: "Central Powers",
    extra: null,
    lines: Array.from({ length: 11 }, (_, l) =>
      Array.from({ length: 8 - (l % 2) }, (): Hex => ({
        terrain: "grass",
        infra: { ids: [], list: false },
        unit: "",
        content: null,
      })),
    ),
  };
  for (const [c, id] of units) at(at(lv.lines, c.line), c.col).unit = id;
  return lv;
}

function passUntilAction(s: play.Session, side: number): void {
  while (play.phase(s, side) !== 2) {
    play.confirm(s, 0);
    play.confirm(s, 1);
    play.resolve(s);
  }
}

const same = (a: Coord, b: Coord): boolean =>
  a.line === b.line && a.col === b.col;

describe("TestGasPioneers", () => {
  it("plays a level with gas pioneers on the extended rules", () => {
    const from = { line: 4, col: 3 },
      to = at(around(from), 3);
    const s = play.newSession(
      grassLevel([
        [from, "CP_gasPioneers"],
        [to, "EC_infantry"],
      ]),
      playData,
      { profile: Full, extended: Extended },
    );
    expect(s.game.profile.tables).toBe(ExtendedTables);
    const v = play.view(s, 0);
    const pioneer = v.units.get(coordKey(from))?.index ?? -1,
      enemy = v.units.get(coordKey(to))?.index ?? -1;
    expect(play.gasPioneer(s, pioneer)).toBe(true);
    expect(play.gasPioneer(s, enemy)).toBe(false);
    expect(play.charges(s, pioneer)).toBe(gasCharges);
    expect(play.charges(s, enemy)).toBe(-1);
    passUntilAction(s, 0);
    play.canRelease(s, 0, pioneer);
    const targets = play.targets(s, 0, pioneer);
    for (const c of around(from)) expect(targets.has(coordKey(c))).toBe(true);
    const cloud = play.cloud(s, 0, pioneer, to);
    expect(cloud).toHaveLength(3);
    expect(cloud.some((c) => same(c, to))).toBe(true);
    expect(play.cloud(s, 0, enemy, from)).toStrictEqual([]);
    let evs = play.attack(s, 0, pioneer, to);
    expect(evs.map((e) => [e.kind, e.text])).toStrictEqual([
      [play.Tasked, "Gas release ordered."],
    ]);
    const t = play.view(s, 0).units.get(coordKey(from))?.target;
    expect(t).toStrictEqual(to);
    play.confirm(s, 0);
    play.confirm(s, 1);
    evs = play.resolve(s);
    expect(evs.filter((e) => e.kind === play.Battle)).toHaveLength(0);
    const released = evs.filter((e) => e.kind === play.Released);
    expect(released).toHaveLength(1);
    expect(at(released, 0).from).toStrictEqual(from);
    expect(at(released, 0).to).toStrictEqual(to);
    expect(at(released, 0).hexes).toStrictEqual(cloud);
    const gassed = evs.filter((e) => e.kind === play.Gassed);
    expect(gassed).toHaveLength(1);
    expect(at(gassed, 0).unit).toBe(enemy);
    expect(at(gassed, 0).to).toStrictEqual(to);
    expect(at(gassed, 0).loss).toBe(2);
    expect(at(gassed, 0).text.endsWith("to gas.")).toBe(true);
    const gas = play.gas(s);
    expect(gas.size).toBe(3);
    expect(gas.get(coordKey(to))).toBe(gasTurnsDefault - 1);
    expect(play.gasAt(s, to)).toBe(gas.get(coordKey(to)));
    expect(play.charges(s, pioneer)).toBe(1);
    const r = play.loadBoard(s.board, s.level, playData, play.save(s), Builds);
    expect(play.gas(r)).toStrictEqual(gas);
  });
});

describe("TestExtendedUnitsOption", () => {
  it("brings the extension units into every factory", () => {
    const lv = shippedLevel("1918-02", "cp");
    lv.extra = lv.extra === null ? null : { ...lv.extra, month: "1918-02" };
    let s = play.newSession(lv, playData, {
      profile: Full,
      extended: Extended,
    });
    expect(s.game.profile.tables).toBe(FullTables);
    s = play.newSession(lv, playData, {
      profile: Full,
      extended: Extended,
      extendedUnits: true,
    });
    expect(s.game.profile.tables).toBe(ExtendedTables);
    expect([s.game.gasTurns, s.game.gasLoss]).toStrictEqual([4, 15]);
    let factories = 0;
    for (const [i, bd] of s.game.buildings)
      if (buildingKind(i) === kindFactory) {
        factories++;
        expect(bd.builds).toContain(51);
      }
    expect(factories).toBeGreaterThan(0);
    for (const [month, want] of [
      ["", [2, 30]],
      ["1915-04", [2, 30]],
      ["1916-01", [2, 15]],
      ["1917-07", [4, 15]],
    ] as const)
      expect(play.gasRates(month)).toStrictEqual(want);
  });
});
