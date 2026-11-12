// Twin of internal/dosai/works_test.go.

import { describe, expect, it } from "vitest";
import * as ai from "./ai.ts";
import * as play from "./play.ts";
import { type Coord, around, distance } from "./hexgrid.ts";
import { type Hex, type Level } from "./level.ts";
import { at } from "./goruntime.ts";
import { Extended, Full } from "./testing/tables.ts";
import { playData, shippedLevel } from "./testing/play.ts";

function session(lv: Level): play.Session {
  return play.newSession(lv, playData, {
    profile: Full,
    extended: Extended,
    computer: [true, true],
  });
}

function grassLevel(units: readonly [Coord, string][]): Level {
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
  return lv;
}

// digs plans side 1's action phase (side 1 starts in it) and returns its work orders.
function digs(s: play.Session): play.Event[] {
  expect(play.phase(s, 1)).toBe(2);
  return ai.plan({}, s, 1).filter((e) => e.kind === play.Tasked);
}

describe("TestTunnellersMineAnEnemy", () => {
  it("mines the enemy next to them, none without one", () => {
    const from = { line: 6, col: 3 },
      enemy = at(around(from), 3);
    let s = session(
      grassLevel([
        [from, "EC_tunnellers"],
        [enemy, "CP_infantry"],
      ]),
    );
    const tasked = digs(s);
    expect(tasked).toHaveLength(1);
    expect(at(tasked, 0).from).toStrictEqual(from);
    expect(at(tasked, 0).to).toStrictEqual(enemy);
    expect(play.tunneller(s, at(tasked, 0).unit)).toBe(true);

    // No enemy next to them: no tunnel.
    s = session(
      grassLevel([
        [from, "EC_tunnellers"],
        [{ line: 18, col: 3 }, "CP_infantry"],
      ]),
    );
    expect(digs(s)).toStrictEqual([]);
  });
});

describe("TestWiringPartyWiresTheFront", () => {
  it("wires toward a near enemy, not a far one", () => {
    const from = { line: 6, col: 3 },
      enemy = { line: 10, col: 3 };
    let s = session(
      grassLevel([
        [from, "EC_wiring"],
        [enemy, "CP_infantry"],
      ]),
    );
    const tasked = digs(s);
    expect(tasked).toHaveLength(1);
    expect(at(tasked, 0).from).toStrictEqual(from);
    expect(play.wiring(s, at(tasked, 0).unit)).toBe(true);
    expect(distance(at(tasked, 0).to, enemy)).toBeLessThan(
      distance(from, enemy),
    );

    // The enemy far away: no wire.
    s = session(
      grassLevel([
        [from, "EC_wiring"],
        [{ line: 18, col: 3 }, "CP_infantry"],
      ]),
    );
    expect(digs(s)).toStrictEqual([]);
  });
});

describe("TestWorksPlayOut", () => {
  it("wires, mines and heals with the new work units in place of some infantry", () => {
    const lv = shippedLevel("1915-05", "cp");
    const swap = ["_wiring", "_tunnellers", "_ambulance", "_flamethrower"];
    let n = 0;
    for (const line of lv.lines)
      for (const h of line)
        if (h.unit.endsWith("_infantry")) {
          if (n % 2 === 0)
            h.unit = h.unit.replace(
              /_infantry$/,
              at(swap, Math.trunc(n / 2) % swap.length),
            );
          n++;
        }
    const s = session(lv);
    const count = new Map<number, number>();
    const keep = (evs: readonly play.Event[]): void => {
      for (const ev of evs) count.set(ev.kind, (count.get(ev.kind) ?? 0) + 1);
    };
    for (let r = 0; r < 300 && play.ended(s) === null; r++) {
      for (const side of [0, 1]) keep(ai.plan({}, s, side));
      keep(play.resolve(s));
    }
    expect(count.get(play.Refused) ?? 0).toBe(0);
    const got = [
      play.Wired,
      play.Tunnelled,
      play.TunnelLost,
      play.Blown,
      play.Mined,
      play.Healed,
    ].map((k) => count.get(k) ?? 0);
    // The Go twin logs wired 4, tunnels 4 (lost 1), mines 3 hitting 1 units, healed 5.
    expect(got).toStrictEqual([4, 4, 1, 3, 1, 5]);
  }, 60_000);
});
