// Twin of internal/dosrules/builder_test.go: construction units build depots.

import { describe, expect, it } from "vitest";
import {
  depotCells,
  msgDepot,
  msgDepotLimit,
  msgNoCharge,
  msgNoSite,
  sites,
} from "./builder.ts";
import { kindDepot } from "./capture.ts";
import {
  type Hold,
  emptyHold,
  holdFlag100,
  holdTable,
  slotFree,
} from "./container.ts";
import {
  type Event,
  type Game,
  check,
  cmdAttack,
  cmdMove,
  evBattle,
  evDepot,
  evHalfDepot,
  evSkipped,
  evWork,
  holdAt,
  liveOf,
  msgNoTarget,
  newGame,
  occupant,
} from "./cycle.ts";
import { at } from "./goruntime.ts";
import {
  applyOk,
  builderScene,
  cycleH,
  cycleW,
  records,
  refusal,
  resolveOk,
  unworked,
} from "./testing/scenes.ts";
import { Builds, Full, Profiles, cBuilder } from "./testing/tables.ts";
import { loadGame, save } from "./save.ts";

// construct gives construction unit u of side its order on cell, resolves the phase and returns the work's
// events (evHalfDepot, evDepot).
function construct(g: Game, side: number, u: number, cell: number): Event[] {
  const evs = applyOk(g, { kind: cmdAttack, side, unit: u, cell });
  expect(evs).toHaveLength(1);
  expect(at(evs, 0).kind).toBe(evWork);
  expect(at(at(evs, 0).messages, 0).code).toBe(msgDepot);
  expect(liveOf(g, u).flags4 & 0x200).not.toBe(0);
  expect(at(g.players, side).orders).toStrictEqual([]);
  const done: Event[] = [];
  for (const ev of resolveOk(g)) {
    expect([evBattle, evSkipped], "the work fought").not.toContain(ev.kind);
    if (ev.kind === evHalfDepot || ev.kind === evDepot) {
      done.push(ev);
    }
  }
  return done;
}

const ascending = (v: readonly number[]): number[] =>
  [...v].sort((a, b) => a - b);

describe("TestBuilderSites", () => {
  it.each(Profiles)("%s", (_build, p) => {
    // Even column (4, 4): the ten anchors whose depot touches the builder without covering it.
    let g = builderScene(p, 40, 0, null, null);
    let want = [21, 22, 23, 29, 33, 38, 39, 41, 42, 49];
    expect(ascending(sites(g, 1, 1))).toStrictEqual(want);
    for (let c = 0; c < cycleW * cycleH; c++) {
      const err = check(g, { kind: cmdAttack, side: 1, unit: 1, cell: c });
      if (want.includes(c)) {
        expect(err, `site ${c}`).toBeNull();
      } else {
        refusal("no site", err, msgNoTarget);
      }
    }
    for (const a of want) {
      expect(
        depotCells(g, a).includes(40),
        `the depot at ${a} covers the builder`,
      ).toBe(false);
    }

    // Odd column (5, 4).
    g = builderScene(p, 41, 0, null, null);
    want = [23, 30, 31, 33, 34, 39, 43, 49, 50, 51];
    expect(ascending(sites(g, 1, 1)), "odd column").toStrictEqual(want);

    // A unit on a cell or terrain sappers leave alone rules out the depots over it; off the map's edge none is
    // offered.
    g = builderScene(p, 40, 50, { 22: unworked(p) }, null);
    for (const c of [41, 49, 22]) {
      refusal(
        "a site over a unit or rock",
        check(g, { kind: cmdAttack, side: 1, unit: 1, cell: c }),
        msgNoTarget,
      );
    }
    g = builderScene(p, 0, 80, null, null);
    for (const c of sites(g, 1, 1)) {
      for (const d of depotCells(g, c)) {
        expect(
          d >= 0 && d < cycleW * cycleH,
          `site ${c} reaches off the map`,
        ).toBe(true);
      }
    }
  });
});

describe("TestBuilderRefusals", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const rock = unworked(p);
    const walls: Record<number, number> = {};
    for (let c = 0; c < cycleW * cycleH; c++) {
      if (c !== 40) {
        walls[c] = rock;
      }
    }
    let g = builderScene(p, 40, 0, walls, null);
    refusal(
      "nowhere to build",
      check(g, { kind: cmdAttack, side: 1, unit: 1, cell: 41 }),
      msgNoSite,
    );

    g = builderScene(p, 40, 0, null, null);
    const u = liveOf(g, 1);
    u.hold = 0;
    g.units.set(1, u);
    refusal(
      "no charges",
      check(g, { kind: cmdAttack, side: 1, unit: 1, cell: 41 }),
      msgNoCharge,
    );
  });
});

describe("TestBuilderHalfDepot", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const g = builderScene(p, 40, 0, null, null);
    const evs = construct(g, 1, 1, 41);
    expect(evs).toHaveLength(1);
    expect([at(evs, 0).kind, at(evs, 0).cell, at(evs, 0).cells]).toStrictEqual([
      evHalfDepot,
      41,
      [41, 51, 50, 49],
    ]);
    expect(records(g, 41, 51, 50, 49), "the half depot").toStrictEqual([
      21, 24, 22, 23,
    ]);
    expect(liveOf(g, 1).hold, "charges").toBe(1);
    expect(
      g.holds.has(holdFlag100),
      "a half depot has a container record",
    ).toBe(false);
    // The half depot is the anchor's site, whatever stands on it.
    expect(sites(g, 1, 1)).toContain(41);
  });
});

describe("TestBuilderCompletesDepot", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const g = builderScene(p, 40, 0, null, null);
    construct(g, 1, 1, 41);
    resolveOk(g); // side 0's action phase
    const evs = construct(g, 1, 1, 41);
    expect(evs).toHaveLength(1);
    expect([at(evs, 0).kind, at(evs, 0).hold, at(evs, 0).cells]).toStrictEqual([
      evDepot,
      holdFlag100,
      [41, 51, 50, 49],
    ]);
    expect(records(g, 41, 51, 50, 49), "side 1's depot").toStrictEqual([
      17, 20, 18, 19,
    ]);
    const h = g.holds.get(holdFlag100);
    expect(h, "no container record").toBeDefined();
    if (h === undefined) {
      return;
    }
    expect(h.cells).toStrictEqual([82, 98, 100, 102]);
    expect(h.flags).toBe(1 | kindDepot);
    expect(h.occupant).toBe(0xff);
    for (const s of h.slots) {
      expect(s).toBeGreaterThan(slotFree);
    }
    expect(
      g.buildings.get(holdFlag100),
      "a building that credits nothing",
    ).toStrictEqual({ credits: 0, builds: [] });
    expect(holdAt(g, 1, 41)).toStrictEqual([holdFlag100, true]);
    expect(liveOf(g, 1).hold, "charges").toBe(0);
    resolveOk(g);
    refusal(
      "no charges left",
      check(g, { kind: cmdAttack, side: 1, unit: 1, cell: 33 }),
      msgNoCharge,
    );
  });
});

describe("TestBuilderCompletesOthersHalf", () => {
  it.each(Profiles)("%s", (_build, p) => {
    // Side 1 lays half a depot; side 0's builder completes it and owns it.
    const tr = Array.from({ length: cycleW * cycleH }, () => 0);
    const g = newGame(
      p,
      cycleW,
      cycleH,
      tr,
      new Map([
        [0, { type: cBuilder, owner: 0, cell: 58 }],
        [1, { type: cBuilder, owner: 1, cell: 40 }],
      ]),
    );
    construct(g, 1, 1, 41);
    const evs = construct(g, 0, 0, 41);
    expect(evs.map((e) => e.kind)).toStrictEqual([evDepot]);
    expect(records(g, 41), "side 0's 16").toStrictEqual([16]);
    expect(g.holds.get(holdFlag100)?.flags).toBe(0 | kindDepot);
  });
});

describe("TestBuilderChecksOtherMap", () => {
  it.each(Profiles)("%s", (_build, p) => {
    // The first half counts the depot's cells on the other side's map: side 0 moves onto one in the same cycle
    // and nothing is built, nor a charge spent.
    const g = builderScene(p, 40, 59, null, null);
    applyOk(g, { kind: cmdAttack, side: 1, unit: 1, cell: 41 });
    applyOk(g, { kind: cmdMove, side: 0, unit: 0, cell: 50 });
    expect(occupant(g, 0, 50), "the move did not arrive").toBe(0);
    for (const ev of resolveOk(g)) {
      expect([evHalfDepot, evDepot], "built").not.toContain(ev.kind);
    }
    expect(records(g, 41, 51, 50, 49)).toStrictEqual([0, 0, 0, 0]);
    expect(liveOf(g, 1).hold, "charges").toBe(2);
  });
});

describe("TestBuilderDepotLimit", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const holds: Record<number, Hold> = {};
    for (let k = 0; k < holdTable; k++) {
      const h = emptyHold();
      h.cells[0] = 2 * (k + 1);
      holds[holdFlag100 + k] = h;
    }
    const g = builderScene(p, 40, 0, { 41: 21, 51: 24, 50: 22, 49: 23 }, holds);
    refusal(
      "the eleventh depot",
      check(g, { kind: cmdAttack, side: 1, unit: 1, cell: 41 }),
      msgDepotLimit,
    );
    expect(
      check(g, { kind: cmdAttack, side: 1, unit: 1, cell: 33 }),
      "a first half beyond the limit",
    ).toBeNull();
  });
});

describe("TestBuilderSaveResume", () => {
  it("full", () => {
    const g = builderScene(Full, 40, 0, null, null);
    construct(g, 1, 1, 41);
    resolveOk(g);
    construct(g, 1, 1, 41);
    const r = loadGame(save(g), Builds);
    expect(r.holds.get(holdFlag100), "resumed depot").toStrictEqual(
      g.holds.get(holdFlag100),
    );
    expect(liveOf(r, 1).hold, "resumed charges").toBe(0);
    expect(
      r.buildings.has(holdFlag100),
      "the resumed depot is a building",
    ).toBe(true);
  });
});
