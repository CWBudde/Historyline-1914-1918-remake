// Twin of internal/dosrules/query_test.go: the container state newGame derives, the queries a session asks of a
// game and attacks of units off the map. TestSaveBuilding needs Save and LoadGame and is left to the save layer
// (save.test.ts).

import { describe, expect, it } from "vitest";
import { destinations } from "./board.ts";
import { airTarget, waterTarget } from "./casualties.ts";
import {
  type Hold,
  emptyHold,
  holdFlag40,
  holdFlag100,
  holdRecords,
  inBuilding,
  moveExit,
} from "./container.ts";
import {
  type Placement,
  check,
  cmdAttack,
  cmdMove,
  evMoved,
  exitMovement,
  holdAt,
  holdOf,
  liveOf,
  movement,
  newGame,
  occupant,
  weaponAgainst,
} from "./cycle.ts";
import { builderCharges } from "./energy.ts";
import { at, attempt } from "./goruntime.ts";
import { reach } from "./move.ts";
import { neighbours } from "./support.ts";
import {
  applyOk,
  containerScene,
  cycleH,
  cycleW,
  depot,
  depotRec,
  hq0,
  hq1,
  newCycleHolds,
  openTerrain,
  refusal,
  transportHold,
} from "./testing/scenes.ts";
import {
  Full,
  FullTables,
  Profiles,
  cInfantry,
  cSupply,
  cTransportRec,
  moveTables,
} from "./testing/tables.ts";

describe("TestNewGameContainers", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const g = containerScene(p);
    const tr = liveOf(g, 1);
    const cargo = liveOf(g, 10);
    const inside = liveOf(g, 5);
    expect(tr.hold).toBe(cTransportRec);
    expect(tr.flags4 & 0x4000).toBe(0);
    expect([occupant(g, 0, 40), occupant(g, 1, 40)]).toStrictEqual([1, 1]);
    expect(cargo.flags4 & 0x4000).not.toBe(0);
    expect(cargo.pos).toStrictEqual([2 * 40 + 1, 2 * 40 + 1]);
    expect(cargo.hold).toBe(0xff);
    expect(inside.flags4 & 0x4000).not.toBe(0);
    expect(inside.pos).toStrictEqual([2 * hq0, 2 * hq0]);
    expect([occupant(g, 0, hq0), occupant(g, 1, hq0)]).toStrictEqual([
      inBuilding,
      0xff,
    ]);
    for (let k = 0; k < 2; k++) {
      for (let c = 0; c < cycleW * cycleH; c++) {
        const o = occupant(g, k, c);
        expect(
          o === 5 || o === 10,
          `map ${k} shows held unit ${o} at ${c}`,
        ).toBe(false);
      }
    }
    expect([liveOf(g, 6).hold, liveOf(g, 0).hold]).toStrictEqual([
      builderCharges,
      0xff,
    ]);

    // A unit that reaches its own transporter: the destinations work and the move boards it.
    const dests = destinations(p, movement(g, 0, 0));
    expect(dests, "the transporter's cell is a destination").toContain(40);
    const evs = applyOk(g, { kind: cmdMove, side: 0, unit: 0, cell: 40 });
    expect(at(evs, 0).kind).toBe(evMoved);
    expect(holdOf(g, cTransportRec).slots[1]).toBe(0);
    expect(liveOf(g, 0).flags4 & 0x4000).not.toBe(0);
    expect(occupant(g, 0, 40)).toBe(1);
  });
});

describe("TestNewGameContainersRefused", () => {
  const terrain = Array.from({ length: cycleW * cycleH }, () =>
    openTerrain(Full),
  );
  const two = emptyHold();
  two.slots[0] = 3;
  it.each([
    [
      "an unplaced unit in a slot",
      [[1, { type: cInfantry, cell: 4 }]],
      [[holdFlag40, two]],
    ],
    [
      "a unit in two records",
      [[3, { type: cInfantry }]],
      [
        [holdFlag40, two],
        [holdFlag100, two],
      ],
    ],
    [
      "a transporter without a record",
      [[1, { type: cSupply, cell: 40 }]],
      [[0, transportHold(41, 0, 1)]],
    ],
    [
      "two transporters on one record",
      [
        [1, { type: cSupply, cell: 40 }],
        [2, { type: cSupply, cell: 40 }],
      ],
      [[0, transportHold(40, 0, 1)]],
    ],
    ["a record past the table", [], [[holdRecords, emptyHold()]]],
  ] as [string, [number, Placement][], [number, Hold][]][])(
    "%s",
    (_what, units, holds) => {
      const [, err] = attempt(() =>
        newGame(Full, cycleW, cycleH, terrain, new Map(units), new Map(holds)),
      );
      expect(err, "accepted").not.toBeNull();
    },
  );
});

// TestCycleAttackOffMap checks the assumed rule that an attacker must be on its side's map.
describe("TestCycleAttackOffMap", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const n = neighbours(40, cycleW, cycleH);
    const g = newCycleHolds(
      p,
      {
        0: { type: cInfantry, owner: 0, cell: 40 },
        1: { type: cSupply, owner: 1, cell: at(n, 0) },
        2: { type: cInfantry, owner: 1 },
        3: { type: cInfantry, owner: 1, cell: at(n, 3) },
        4: { type: cInfantry, owner: 1, cell: at(n, 1) },
      },
      { [cTransportRec]: transportHold(at(n, 0), 1, 1, 2) },
    );
    refusal(
      "a held unit's attack",
      check(g, { kind: cmdAttack, side: 1, unit: 2, cell: 40 }),
      0,
    );
    g.maps[1][2 * at(n, 3) + 1] = 4;
    refusal(
      "an attacker its map does not show",
      check(g, { kind: cmdAttack, side: 1, unit: 3, cell: 40 }),
      0,
    );
    expect(
      check(g, { kind: cmdAttack, side: 1, unit: 4, cell: 40 }),
      "an attacker on the map",
    ).toBeNull();
  });
});

describe("TestExitMovement", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const n = neighbours(40, cycleW, cycleH);
    const g = containerScene(p);
    const m = exitMovement(g, 0, cTransportRec, 0);
    expect([m.unit, m.budget, m.block]).toStrictEqual([10, 8, 3]);
    expect(m.held.get(10)?.cell).toBe(40);
    expect(m.units.has(10)).toBe(false);
    expect(destinations(p, m)).toContain(at(n, 2));

    // The exit's own reach search is the query's.
    const tab = moveTables(p);
    const r = moveExit(tab, movement(g, 0, 10), {
      hold: cTransportRec,
      slot: 0,
      goal: at(n, 2),
      retarget: -1,
    });
    const [buf] = reach(tab, m);
    expect(buf, "the exit's reach differs from exitMovement's").toStrictEqual(
      r.remaining,
    );

    // From the HQ, at the record's first cell.
    const h = exitMovement(g, 0, holdFlag40, 0);
    expect(h.unit).toBe(5);
    expect(h.held.get(5)?.cell).toBe(hq0);
    for (const [what, [side, hold, slot]] of [
      ["an empty slot", [0, cTransportRec, 1]],
      ["no record", [0, 9, 0]],
      ["a bad slot", [0, cTransportRec, 7]],
      ["a bad side", [2, cTransportRec, 0]],
    ] as const) {
      const [, err] = attempt(() => exitMovement(g, side, hold, slot));
      expect(err, `${what}: no error`).not.toBeNull();
    }
  });
});

describe("TestHoldAt", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const g = containerScene(p);
    const d = emptyHold();
    d.cells[0] = 2 * depot;
    g.holds.set(depotRec, d);
    const depotTerrain = moveTables(p).terrains.findIndex(
      (tr) => (tr.flags & 0x500) === 0x100,
    );
    for (const m of g.maps) {
      m[2 * depot] = depotTerrain;
      m[2 * (depot + 1)] = depotTerrain;
    }
    for (const [side, cell, rec, ok] of [
      [0, 40, cTransportRec, true],
      [1, 40, cTransportRec, true],
      [0, hq0, holdFlag40, true],
      [0, hq1, holdFlag40 + 1, true],
      [0, depot, depotRec, true],
      [0, depot + 1, 0, false],
      [0, 44, 0, false],
      [0, 41, 0, false],
      [0, -1, 0, false],
    ] as const) {
      expect(holdAt(g, side, cell), `side ${side} cell ${cell}`).toStrictEqual([
        rec,
        ok,
      ]);
    }
  });
});

describe("TestWeaponAgainst", () => {
  it("full", () => {
    for (const [i, ty] of FullTables.types.entries()) {
      for (const [flags, power, rng] of [
        [airTarget, ty.airPower, ty.airRange],
        [waterTarget, ty.waterPower, ty.surfaceRange],
        [4, ty.groundPower, ty.surfaceRange],
      ]) {
        expect(
          weaponAgainst(Full, i, flags ?? 0),
          `type ${i} against ${flags}`,
        ).toStrictEqual([power, rng]);
      }
    }
  });
});

it.todo("TestSaveBuilding: needs Save and LoadGame (the save layer)");
