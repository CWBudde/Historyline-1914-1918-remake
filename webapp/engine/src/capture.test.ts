// Twin of internal/dosrules/capture_test.go. Capture: a unit of the movement side that ends on another side's or
// a neutral building takes it and the units in it at the next resolution.

import { describe, expect, it } from "vitest";
import { destinations, destroyed } from "./board.ts";
import { kindDepot } from "./capture.ts";
import {
  cmdMove,
  evCaptured,
  evPhase,
  holdOf,
  liveOf,
  movement,
  newEvent,
  occupant,
} from "./cycle.ts";
import { at } from "./goruntime.ts";
import { owns } from "./move.ts";
import { neighbours } from "./support.ts";
import {
  applyOk,
  buildingTerrain,
  cycleH,
  cycleW,
  depot,
  depotCycle,
  depotRec,
  resolveOk,
} from "./testing/scenes.ts";
import { Profiles, cInfantry, cPanzer } from "./testing/tables.ts";

describe("TestCaptureNeutralDepot", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const n = neighbours(depot, cycleW, cycleH);
    const g = depotCycle(
      p,
      2,
      {
        0: { type: cInfantry, owner: 0, cell: at(n, 0) },
        1: { type: cPanzer, owner: 1, cell: 0 },
        2: { type: cPanzer, owner: 2, cell: depot },
      },
      2,
    );
    g.buildings = new Map([[depotRec, { credits: 7, builds: [] }]]);
    expect(
      holdOf(g, depotRec).flags,
      "depot flags: the neutral bit and the depot kind",
    ).toBe(2 | kindDepot);
    applyOk(g, { kind: cmdMove, side: 0, unit: 0, cell: depot });
    const evs = resolveOk(g);
    const i = evs.findIndex((e) => e.kind === evCaptured);
    expect(i, "no capture event").toBeGreaterThanOrEqual(0);
    expect(at(evs, i)).toStrictEqual(
      newEvent({
        kind: evCaptured,
        side: 0,
        unit: 0,
        cell: depot,
        hold: depotRec,
      }),
    );
    const j = evs.findIndex((e) => e.kind === evPhase);
    expect(j, "the capture follows the phase change").toBeGreaterThanOrEqual(i);
    const h = holdOf(g, depotRec);
    expect(h.flags, "side 0's depot").toBe(kindDepot);
    for (let k = 0; k < 2; k++) {
      expect(
        h.slots.slice(7 * k, 7 * k + 2),
        `side ${k}'s slots: the unit found and the capturer`,
      ).toStrictEqual([2, 0]);
      expect(at(at(g.maps, k), 2 * depot)).toBe(buildingTerrain(p, 0x100, 4));
      const occ = occupant(g, k, depot);
      expect(occ !== 0 && occ !== 2, `side ${k} shows unit ${occ}`).toBe(true);
    }
    const found = liveOf(g, 2);
    const capturer = liveOf(g, 0);
    expect(owns(found.flags4, 0)).toBe(true);
    expect(found.flags4 & 3).toBe(0);
    expect(capturer.flags4 & 0x4000).not.toBe(0);
    expect(capturer.pos).toStrictEqual([2 * depot, 2 * depot]);
    expect(g.energy, "the depot's 7 for side 0").toStrictEqual([7, 0]);
  });
});

// TestCaptureEnemyDepotFull takes a depot whose seven slots are full: the unit in the first slot is destroyed and
// the capturer takes its place.
describe("TestCaptureEnemyDepotFull", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const n = neighbours(depot, cycleW, cycleH);
    const units: Record<number, { type: number; owner: number; cell: number }> =
      { 0: { type: cInfantry, owner: 0, cell: at(n, 0) } };
    const inside: number[] = [];
    for (let i = 1; i <= 7; i++) {
      units[i] = { type: cInfantry, owner: 1, cell: depot };
      inside.push(i);
    }
    const g = depotCycle(p, 1, units, ...inside);
    applyOk(g, { kind: cmdMove, side: 0, unit: 0, cell: depot });
    resolveOk(g);
    const h = holdOf(g, depotRec);
    expect(h.flags).toBe(kindDepot);
    expect(h.slots.slice(0, 7)).toStrictEqual([0, 2, 3, 4, 5, 6, 7]);
    expect(destroyed(liveOf(g, 1)), "the unit in the first slot survives").toBe(
      true,
    );
    for (let i = 2; i <= 7; i++) {
      const u = liveOf(g, i);
      expect(u.flags4 & 3, `unit ${i}`).toBe(0);
      expect(destroyed(u), `unit ${i}`).toBe(false);
    }
  });
});

// TestCaptureOnlyFoot keeps a unit without type flag bit 1 off a foreign building's entrance.
describe("TestCaptureOnlyFoot", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const n = neighbours(depot, cycleW, cycleH).filter((c) => c >= 0);
    const g = depotCycle(p, 2, {
      0: { type: cPanzer, owner: 0, cell: at(n, 0) },
      1: { type: cInfantry, owner: 0, cell: at(n, 1) },
    });
    for (const [u, want] of [false, true].entries()) {
      const cells = destinations(p, movement(g, 0, u));
      expect(cells.includes(depot), `unit ${u} reaches the neutral depot`).toBe(
        want,
      );
    }
  });
});

// TestCaptureOnce: the action bytes are cleared at the phase change, so a later resolution does not capture
// again.
describe("TestCaptureOnce", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const n = neighbours(depot, cycleW, cycleH);
    const g = depotCycle(p, 2, {
      0: { type: cInfantry, owner: 0, cell: at(n, 0) },
      1: { type: cPanzer, owner: 1, cell: 0 },
    });
    applyOk(g, { kind: cmdMove, side: 0, unit: 0, cell: depot });
    resolveOk(g);
    for (let k = 0; k < 2; k++) {
      for (const e of resolveOk(g)) {
        expect(e.kind, "captured again").not.toBe(evCaptured);
      }
    }
  });
});
