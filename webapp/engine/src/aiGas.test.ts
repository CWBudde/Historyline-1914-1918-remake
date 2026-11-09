// Twin of internal/dosai/gas_test.go.

import { describe, expect, it } from "vitest";
import * as ai from "./ai.ts";
import * as play from "./play.ts";
import { type Coord, around } from "./hexgrid.ts";
import { type Hex, type Level, coordKey } from "./level.ts";
import { at } from "./goruntime.ts";
import { Extended, Full } from "./testing/tables.ts";
import { playData, shippedLevel } from "./testing/play.ts";

function grass(units: readonly [Coord, string][]): play.Session {
  const lv: Level = {
    mapName: "grass",
    mapNo: 0,
    player: "Central Powers",
    extra: null,
    lines: Array.from({ length: 21 }, (_, l) =>
      Array.from({ length: 8 - (l % 2) }, (): Hex => ({
        terrain: "grass",
        infra: { ids: [], list: false },
        unit: "",
        content: null,
      })),
    ),
  };
  for (const [c, id] of units) at(at(lv.lines, c.line), c.col).unit = id;
  return play.newSession(lv, playData, {
    profile: Full,
    extended: Extended,
    computer: [true, true],
  });
}

// gasLevel is a session with side 1's gas pioneers at from and units on the hexes of the cloud on to.
function gasLevel(
  from: Coord,
  to: Coord,
  inCloud: Readonly<Record<number, string>>,
  more: readonly [Coord, string][] = [],
): [play.Session, Coord[]] {
  const s0 = grass([[from, "EC_gasPioneers"]]);
  const cloud = play.cloud(
    s0,
    1,
    play.view(s0, 1).units.get(coordKey(from))?.index ?? -1,
    to,
  );
  expect(cloud).toHaveLength(3);
  const units: [Coord, string][] = [[from, "EC_gasPioneers"], ...more];
  for (const [k, id] of Object.entries(inCloud))
    units.push([at(cloud, Number(k)), id]);
  return [grass(units), cloud];
}

function tasked(s: play.Session): play.Event[] {
  expect(play.phase(s, 1)).toBe(2);
  return ai.plan({}, s, 1).filter((e) => e.kind === play.Tasked);
}

const has = (cs: Coord[], c: Coord): boolean =>
  cs.some((p) => p.line === c.line && p.col === c.col);

describe("TestGasRelease", () => {
  it("releases over two enemies, never over an own unit", () => {
    const from = { line: 6, col: 3 },
      to = at(around(from), 3),
      far: [Coord, string][] = [[{ line: 18, col: 3 }, "CP_infantry"]];
    const [s, cloud] = gasLevel(from, to, {
      0: "CP_infantry",
      1: "CP_infantry",
    });
    const t = tasked(s);
    expect(t).toHaveLength(1);
    expect(at(t, 0).from).toStrictEqual(from);
    expect(has(cloud, at(t, 0).to)).toBe(true);
    const got = play.cloud(s, 1, at(t, 0).unit, at(t, 0).to);
    expect(has(got, at(cloud, 0)) && has(got, at(cloud, 1))).toBe(true);
    for (const units of [
      { 0: "CP_infantry" },
      { 0: "CP_infantry", 1: "EC_infantry", 2: "CP_infantry" },
    ]) {
      const [s2] = gasLevel(from, to, units, far);
      for (const ev of tasked(s2))
        expect(play.gasPioneer(s2, ev.unit)).toBe(false);
    }
  });
});

describe("TestMayBuildExtension", () => {
  it("lets both sides build the extension types", () => {
    for (const side of [0, 1]) {
      expect(ai.mayBuild([], side, 51)).toBe(true);
      expect(ai.mayBuild([18], side, 51)).toBe(true);
    }
    expect(ai.mayBuild([18], 0, 36)).toBe(false);
  });
});

describe("TestGasPlayOut", () => {
  it("gasses the enemy and keeps its own units out", () => {
    const lv = shippedLevel("1915-05", "cp");
    let n = 0;
    for (const line of lv.lines)
      for (const h of line)
        if (h.unit.endsWith("_infantry")) {
          if (n % 3 === 0)
            h.unit = h.unit.replace(/_infantry$/, "_gasPioneers");
          n++;
        }
    const s = play.newSession(lv, playData, {
      profile: Full,
      extended: Extended,
      computer: [true, true],
    });
    let released = 0,
      enemy = 0,
      own = 0;
    const sideOf = new Map<string, number>();
    for (let r = 0; r < 300 && play.ended(s) === null; r++) {
      for (const side of [0, 1]) ai.plan({}, s, side);
      for (const ev of play.resolve(s)) {
        if (ev.kind === play.Released) {
          released++;
          for (const c of ev.hexes) sideOf.set(coordKey(c), ev.side);
        } else if (ev.kind === play.Gassed) {
          if (sideOf.get(coordKey(ev.to)) === ev.side) own++;
          else enemy++;
        }
      }
    }
    expect(released).toBeGreaterThan(0);
    expect(enemy).toBeGreaterThan(0);
    expect(own).toBeLessThanOrEqual(enemy / 2);
    // The Go twin logs 4 releases, 7 enemy and 0 own losses.
    expect([released, enemy, own]).toStrictEqual([4, 7, 0]);
  });
});
