// Twin of internal/dosrules/energy_test.go. Energy: the building credits, repair and production.

import { describe, expect, it } from "vitest";
import { type Live, destroyed, zeroLive } from "./board.ts";
import { emptyHold, holdFlag40, inBuilding, slotFree } from "./container.ts";
import {
  check,
  cmdBuild,
  cmdExit,
  cmdRepair,
  evBuilt,
  evCredit,
  evMoved,
  evPhase,
  evRepaired,
  holdOf,
  liveOf,
  occupant,
  phaseAction,
  phaseMovement,
} from "./cycle.ts";
import { array, intField, object } from "./decode.ts";
import { at } from "./goruntime.ts";
import { type Profile, tablesOf } from "./profile.ts";
import { neighbours } from "./support.ts";
import { readJSON } from "./testing/files.ts";
import {
  applyOk,
  cycleH,
  cycleW,
  depotRec,
  energyCycle,
  held,
  hq0,
  midRound,
  refusal,
  reload,
  resolveOk,
} from "./testing/scenes.ts";
import {
  Demo,
  Full,
  FullTables,
  Profiles,
  cBuilder,
  cInfantry,
  cPanzer,
  cSupply,
  cTransportRec,
  moveTables,
} from "./testing/tables.ts";

function typeOf(p: Profile, t: number): { group: number; cost: number } {
  return at(tablesOf(p).types, t);
}

describe("TestEnergyCredits", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const g = energyCycle(p, {
      0: { type: cInfantry, owner: 0, cell: 40 },
      1: { type: cInfantry, owner: 1, cell: 44 },
    });
    g.energy = [5, 1];
    const evs = resolveOk(g);
    expect(g.energy).toStrictEqual([25, 8]);
    const got = evs
      .filter((e) => e.kind === evCredit)
      .map((e) => [e.side, e.energy, e.turn]);
    expect(got).toStrictEqual([
      [0, 25, g.turn],
      [1, 8, g.turn],
    ]);
    const i = evs.findIndex((e) => e.kind === evCredit);
    expect(i, "credits do not follow the phase change").toBeGreaterThanOrEqual(
      1,
    );
    expect(at(evs, i - 1).kind).toBe(evPhase);
  });
});

describe("TestEnergyRepair", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const g = energyCycle(
      p,
      {
        0: { type: cInfantry, owner: 0, size: 2, exp: 4 },
        1: { type: cInfantry, owner: 1, cell: 44 },
        2: { type: cBuilder, owner: 0, size: 1 },
        3: { type: cInfantry, owner: 0, cell: 42, size: 1 },
      },
      { rec: holdFlag40, side: 0, slot: 0, unit: 0 },
      { rec: holdFlag40, side: 0, slot: 1, unit: 2 },
    );
    g.energy[0] = 65;
    const b = liveOf(g, 2);
    expect(b.hold, "a new construction unit's charges").toBe(2);
    b.hold = 0;
    g.units.set(2, b);

    refusal(
      "an empty slot",
      check(g, {
        kind: cmdRepair,
        side: 0,
        hold: holdFlag40,
        slot: 3,
        cell: 0,
      }),
      0,
    );
    refusal(
      "the other side's slot",
      check(g, {
        kind: cmdRepair,
        side: 1,
        hold: holdFlag40,
        slot: 0,
        cell: 0,
      }),
      0,
    );
    refusal(
      "no building",
      check(g, {
        kind: cmdRepair,
        side: 0,
        hold: cTransportRec,
        slot: 0,
        cell: 0,
      }),
      0,
    );

    const evs = applyOk(g, {
      kind: cmdRepair,
      side: 0,
      hold: holdFlag40,
      slot: 0,
      cell: 0,
    });
    const u = liveOf(g, 0);
    expect(evs).toHaveLength(1);
    expect([at(evs, 0).kind, at(evs, 0).unit, at(evs, 0).energy]).toStrictEqual(
      [evRepaired, 0, 50],
    );
    expect(u.size).toBe(typeOf(p, cInfantry).group);
    expect(u.exp).toBe(4);
    expect(u.flags4 & 0x200).not.toBe(0);
    expect(g.energy[0]).toBe(50);
    expect(
      g.players[0].orders,
      "the repair entered the order list",
    ).toStrictEqual([]);
    refusal(
      "a second repair",
      check(g, {
        kind: cmdRepair,
        side: 0,
        hold: holdFlag40,
        slot: 0,
        cell: 0,
      }),
      0,
    );

    applyOk(g, {
      kind: cmdRepair,
      side: 0,
      hold: holdFlag40,
      slot: 1,
      cell: 0,
    });
    const r = liveOf(g, 2);
    expect([r.size, r.hold, g.energy[0]]).toStrictEqual([
      typeOf(p, cBuilder).group,
      2,
      0,
    ]);
  });
});

describe("TestEnergyRepairRefused", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const g = energyCycle(
      p,
      {
        0: { type: cInfantry, owner: 0, size: 2 },
        1: { type: cInfantry, owner: 1, size: 2 },
        2: { type: cBuilder, owner: 0, size: 1 },
      },
      { rec: holdFlag40, side: 0, slot: 0, unit: 0 },
      { rec: depotRec, side: 1, slot: 0, unit: 1 },
      { rec: holdFlag40, side: 0, slot: 1, unit: 2 },
    );
    g.energy = [49, 100];
    refusal(
      "a builder without 50",
      check(g, {
        kind: cmdRepair,
        side: 0,
        hold: holdFlag40,
        slot: 1,
        cell: 0,
      }),
      0,
    );
    refusal(
      "the movement phase",
      check(g, { kind: cmdRepair, side: 1, hold: depotRec, slot: 0, cell: 0 }),
      0,
    );
    g.energy[0] = 14;
    refusal(
      "less than 15",
      check(g, {
        kind: cmdRepair,
        side: 0,
        hold: holdFlag40,
        slot: 0,
        cell: 0,
      }),
      0,
    );

    // A unit of side 0 left in a record side 1 holds is not repaired there.
    held(g, depotRec, 0, 1, 2);
    g.energy[0] = 100;
    refusal(
      "a foreign building",
      check(g, { kind: cmdRepair, side: 0, hold: depotRec, slot: 1, cell: 0 }),
      0,
    );
  });
});

describe("TestEnergyBuild", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const g = energyCycle(p, {
      0: { type: cInfantry, owner: 0, cell: 40 },
      1: { type: cInfantry, owner: 1, cell: 44 },
      3: { type: cInfantry, owner: 0, cell: 42 },
    });
    const cost = typeOf(p, cPanzer).cost;
    g.energy[0] = cost + typeOf(p, cBuilder).cost;
    const s = liveOf(g, 3);
    s.serial = 5;
    g.units.set(3, s);

    const build = (type: number, side = 0, hold = holdFlag40) =>
      ({ kind: cmdBuild, side, hold, type, cell: 0 }) as const;
    const evs = applyOk(g, build(cPanzer));
    expect(evs).toHaveLength(1);
    expect([at(evs, 0).kind, at(evs, 0).unit, at(evs, 0).cell]).toStrictEqual([
      evBuilt,
      2,
      hq0,
    ]);
    const u = liveOf(g, 2);
    const h = holdOf(g, holdFlag40);
    const m = at(moveTables(p).types, cPanzer);
    const want: Live = {
      ...zeroLive(),
      budget: m.budget,
      size: typeOf(p, cPanzer).group,
      flags4: m.class | 0x4200,
      flags6: m.flags,
      type: cPanzer,
      serial: 6,
      hold: 0xff,
      pos: [2 * hq0, 2 * hq0],
    };
    expect(u, "built unit").toStrictEqual(want);
    expect(h.slots[0]).toBe(2);
    expect(g.energy[0]).toBe(typeOf(p, cBuilder).cost);
    expect(occupant(g, 0, hq0)).toBe(inBuilding);
    expect(
      g.players[0].orders,
      "the build entered the order list",
    ).toStrictEqual([]);

    applyOk(g, build(cBuilder));
    const b = liveOf(g, 4);
    expect([
      b.type,
      b.hold,
      g.energy[0],
      holdOf(g, holdFlag40).slots[1],
    ]).toStrictEqual([cBuilder, 2, 0, 4]);

    g.energy[0] = 1000;
    g.players[0].phase = phaseMovement;
    refusal("the movement phase", check(g, build(cInfantry)), 0);
    g.players[0].phase = phaseAction;
    refusal("a type the building does not build", check(g, build(cSupply)), 0);
    refusal("a type never sold", check(g, build(13)), 0);
    refusal("no build list", check(g, build(cInfantry, 1, depotRec)), 0);
    refusal("another side's building", check(g, build(cInfantry, 1)), 0);
    g.energy[0] = typeOf(p, cInfantry).cost - 1;
    refusal("too little energy", check(g, build(cInfantry)), 0);
    g.energy[0] = 1000;
    for (let k = 0; k < 5; k++) {
      applyOk(g, build(cInfantry));
    }
    refusal("no free slot", check(g, build(cInfantry)), 0);
  });
});

// TestEnergyBuildReusesDestroyedRecord checks that a build can take the record of a destroyed unit once every
// live record has been used.
describe("TestEnergyBuildReusesDestroyedRecord", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const g = energyCycle(p, {
      0: { type: cInfantry, owner: 0, cell: 40 },
      1: { type: cInfantry, owner: 1, cell: 44 },
    });
    for (let i = 2; i <= slotFree; i++) {
      g.units.set(i, { ...zeroLive(), type: cInfantry, size: 1, flags4: 2 });
    }
    g.units.set(2, { ...zeroLive(), type: cInfantry, flags4: 0x8000 });
    g.energy[0] = typeOf(p, cInfantry).cost;

    const evs = applyOk(g, {
      kind: cmdBuild,
      side: 0,
      hold: holdFlag40,
      type: cInfantry,
      cell: 0,
    });
    expect(evs.map((e) => [e.kind, e.unit])).toStrictEqual([[evBuilt, 2]]);
    const u = liveOf(g, 2);
    expect(destroyed(u)).toBe(false);
    expect(u.type).toBe(cInfantry);
    expect(holdOf(g, holdFlag40).slots[0]).toBe(2);
  });
});

// TestEnergyBuildExit builds a unit in side 0's action phase and takes it out of the HQ in the next movement
// phase.
describe("TestEnergyBuildExit", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const g = energyCycle(p, {
      0: { type: cInfantry, owner: 0, cell: 40 },
      1: { type: cInfantry, owner: 1, cell: 44 },
    });
    g.energy[0] = 100;
    applyOk(g, {
      kind: cmdBuild,
      side: 0,
      hold: holdFlag40,
      type: cInfantry,
      cell: 0,
    });
    refusal(
      "an exit in the phase it was built",
      check(g, {
        kind: cmdExit,
        side: 0,
        hold: holdFlag40,
        slot: 0,
        cell: hq0 + 1,
      }),
      0,
    );
    resolveOk(g);
    expect(g.players[0].phase).toBe(phaseMovement);
    const n = neighbours(hq0, cycleW, cycleH);
    const evs = applyOk(g, {
      kind: cmdExit,
      side: 0,
      hold: holdFlag40,
      slot: 0,
      cell: at(n, 0),
    });
    expect(at(evs, 0).kind).toBe(evMoved);
    expect(occupant(g, 0, at(n, 0))).toBe(2);
  });
});

// TestEnergyBuildTransporter builds a supply car: it gets its own container record, as a scenario's transporter
// does, and leaves the HQ in the next movement phase with the record following it.
describe("TestEnergyBuildTransporter", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const g = energyCycle(p, {
      0: { type: cInfantry, owner: 0, cell: 40 },
      1: { type: cInfantry, owner: 1, cell: 44 },
    });
    g.buildings.set(holdFlag40, { credits: 0, builds: [cSupply] });
    g.energy[0] = 100;
    const taken = emptyHold();
    taken.flags = 0x20;
    g.holds.set(0, taken);
    applyOk(g, {
      kind: cmdBuild,
      side: 0,
      hold: holdFlag40,
      type: cSupply,
      cell: 0,
    });
    const u = liveOf(g, 2);
    const want = emptyHold();
    want.cells[0] = 2 * hq0 + 1;
    want.flags = 0x20 | at(moveTables(p).types, cSupply).carries;
    want.occupant = 2;
    expect(u.hold).toBe(1);
    expect(g.holds.get(1)).toStrictEqual(want);
    resolveOk(g);
    const n = neighbours(hq0, cycleW, cycleH);
    const evs = applyOk(g, {
      kind: cmdExit,
      side: 0,
      hold: holdFlag40,
      slot: 0,
      cell: at(n, 0),
    });
    expect(at(evs, 0).kind).toBe(evMoved);
    expect(occupant(g, 0, at(n, 0))).toBe(2);
    expect(holdOf(g, u.hold).cells[0]).toBe(2 * at(n, 0) + 1);
  });
});

// TestEnergyBuildReusesClearedCarrierRecord checks that building a transporter can take a cleared transporter
// record.
describe("TestEnergyBuildReusesClearedCarrierRecord", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const g = energyCycle(p, {
      0: { type: cInfantry, owner: 0, cell: 40 },
      1: { type: cInfantry, owner: 1, cell: 44 },
    });
    g.buildings.set(holdFlag40, { credits: 0, builds: [cSupply] });
    g.energy[0] = 100;
    for (let i = 0; i < holdFlag40; i++) {
      const h = emptyHold();
      h.flags = 0x20;
      g.holds.set(i, h);
    }
    const cleared = holdOf(g, 0);
    cleared.flags = 0x8000;
    g.holds.set(0, cleared);

    applyOk(g, {
      kind: cmdBuild,
      side: 0,
      hold: holdFlag40,
      type: cSupply,
      cell: 0,
    });
    const u = liveOf(g, 2);
    expect(u.hold).toBe(0);
    expect(holdOf(g, 0).flags & 0x8000).toBe(0);
    expect(holdOf(g, 0).occupant).toBe(2);
  });
});

describe("TestEnergySave", () => {
  it("full", () => {
    const g = midRound(Full);
    g.energy = [12, 34];
    g.buildings = new Map([
      [holdFlag40, { credits: 20, builds: [cInfantry] }],
      [depotRec, { credits: 7, builds: [] }],
    ]);
    expect(reload(g), "loaded game differs").toStrictEqual(g);
  });
});

// TestCosts checks the prices against the full build's unit table.
describe("TestCosts", () => {
  it("full", () => {
    const recs = array(
      readJSON("internal/data/testdata/unitdat.json"),
      "unitdat.json",
    ).map((r, i) => intField(object(r, `unitdat[${i}]`), "cost", "unitdat"));
    expect(recs).toHaveLength(FullTables.types.length);
    for (const [i, cost] of recs.entries()) {
      expect(at(FullTables.types, i).cost, `type ${i}`).toBe(cost);
    }
    expect(Full.builders).toStrictEqual([cBuilder]);
    expect(Demo.builders).toStrictEqual(Full.builders);
  });
});
