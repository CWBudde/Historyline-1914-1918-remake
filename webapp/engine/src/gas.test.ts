// Twin of internal/dosrules/gas_test.go: gas pioneers release a lasting cloud.

import { describe, expect, it } from "vitest";
import {
  type Event,
  type EventKind,
  type Game,
  type Placement,
  check,
  cmdAttack,
  evBattle,
  evWork,
  liveOf,
  msgNoTarget,
  occupant,
} from "./cycle.ts";
import { charges } from "./energy.ts";
import {
  evGasCleared,
  evGasCloud,
  evGassed,
  flagGas,
  gasAt,
  gasCells,
  gasCharges,
  gasLossDefault,
  gasTurnsDefault,
  gasType,
  msgGas,
  msgNoGas,
  releasing,
} from "./gas.ts";
import { at } from "./goruntime.ts";
import { loadGame, save } from "./save.ts";
import { neighbours } from "./support.ts";
import {
  applyOk,
  cycleH,
  cycleW,
  newCycle,
  refusal,
  resolveOk,
} from "./testing/scenes.ts";
import {
  Builds,
  Extended,
  ExtendedTables,
  Full,
  FullTables,
  cInfantry,
  cSapper,
} from "./testing/tables.ts";

const cGas = 51;
const cArmCar = 0;
const cFighter = 5;

function gasScene(from: number, more: Record<number, Placement>): Game {
  return newCycle(Extended, {
    1: { type: cGas, owner: 1, cell: from },
    ...more,
  });
}

// release gives pioneer 1 its order on cell and resolves side 1's phase.
function release(g: Game, cell: number): Event[] {
  const evs = applyOk(g, { kind: cmdAttack, side: 1, unit: 1, cell });
  expect(evs).toHaveLength(1);
  expect(at(evs, 0).kind).toBe(evWork);
  expect(at(at(evs, 0).messages, 0).code).toBe(msgGas);
  const u = liveOf(g, 1);
  expect(u.flags4 & 0x200).not.toBe(0);
  expect(releasing(u)).toBe(true);
  expect(g.players[1].orders).toStrictEqual([]);
  return resolveOk(g);
}

function kinds(evs: Event[], k: EventKind): Event[] {
  return evs.filter((e) => e.kind === k);
}

describe("TestExtendedTables", () => {
  it("appends the extension to the full tables", () => {
    const x = ExtendedTables,
      f = FullTables;
    expect(f.types).toHaveLength(51);
    expect(f.move?.types).toHaveLength(51);
    expect(x.types.length).toBeGreaterThan(51);
    expect(x.move?.types.length).toBe(x.types.length);
    expect(x.types.slice(0, 51)).toStrictEqual(f.types);
    expect(x.move?.types.slice(0, 51)).toStrictEqual(f.move?.types);
    expect(x.build).toBe(f.build + "+units1");
    expect(x.terms).toStrictEqual(f.terms);
    expect(gasType(Extended, cGas)).toBe(true);
    expect(gasType(Full, cGas)).toBe(false);
    expect(gasType(Extended, cSapper)).toBe(false);
    for (const m of f.move?.types ?? []) expect(m.flags & flagGas).toBe(0);
  });
});

describe("TestGasTargets", () => {
  it("takes any neighbour while charged", () => {
    const n = neighbours(40, cycleW, cycleH);
    const g = gasScene(40, {
      0: { type: cInfantry, owner: 0, cell: at(n, 0) },
    });
    for (const c of n) {
      expect(
        check(g, { kind: cmdAttack, side: 1, unit: 1, cell: c }),
      ).toBeNull();
    }
    refusal(
      "a cell beyond the neighbours",
      check(g, { kind: cmdAttack, side: 1, unit: 1, cell: 42 }),
      msgNoTarget,
    );
    refusal(
      "its own cell",
      check(g, { kind: cmdAttack, side: 1, unit: 1, cell: 40 }),
      msgNoTarget,
    );
    expect(gasCells(g, 1, 1, 31)).toStrictEqual([30, 31, 32]);
    expect(gasCells(g, 1, 1, 42)).toStrictEqual([]);
    const u = liveOf(g, 1);
    u.hold = 0;
    g.units.set(1, u);
    refusal(
      "no charge",
      check(g, { kind: cmdAttack, side: 1, unit: 1, cell: 31 }),
      msgNoGas,
    );
  });
});

describe("TestGasCloud", () => {
  it("gasses friend and foe once a turn, then thins", () => {
    const g = gasScene(40, {
      0: { type: cInfantry, owner: 0, cell: 31 },
      2: { type: cInfantry, owner: 1, cell: 30 },
      3: { type: cInfantry, owner: 0, cell: 22 },
    });
    const full = at(ExtendedTables.types, cInfantry).group;
    expect(liveOf(g, 1).hold).toBe(gasCharges);
    let evs = release(g, 31);
    const cloud = kinds(evs, evGasCloud);
    expect(cloud).toHaveLength(1);
    expect(at(cloud, 0).cell).toBe(31);
    expect(at(cloud, 0).cells).toStrictEqual([30, 31, 32]);
    expect(kinds(evs, evBattle)).toHaveLength(0);
    expect(kinds(evs, evGassed)).toHaveLength(0);
    expect(liveOf(g, 1).hold).toBe(gasCharges - 1);
    expect(gasAt(g, 31)).toBe(gasTurnsDefault);
    expect(gasAt(g, 22)).toBe(0);

    const loss = Math.trunc((full * gasLossDefault + 99) / 100);
    evs = resolveOk(g);
    const gassed = kinds(evs, evGassed);
    expect(gassed.map((e) => [e.unit, e.side, e.loss])).toStrictEqual([
      [0, 0, loss],
      [2, 1, loss],
    ]);
    expect(liveOf(g, 0).size).toBe(full - loss);
    expect(liveOf(g, 2).size).toBe(full - loss);
    expect(liveOf(g, 3).size).toBe(full);
    expect(liveOf(g, 1).size).toBe(full);
    expect(gasAt(g, 31)).toBe(gasTurnsDefault - 1);

    evs = resolveOk(g);
    expect(kinds(evs, evGassed)).toHaveLength(0);
    expect(gasAt(g, 31)).toBe(gasTurnsDefault - 1);
    evs = resolveOk(g);
    expect(kinds(evs, evGassed)).toHaveLength(2);
    expect(liveOf(g, 0).size).toBe(full - 2 * loss);
    const cl = kinds(evs, evGasCleared);
    expect(cl).toHaveLength(1);
    expect(at(cl, 0).cells).toStrictEqual([30, 31, 32]);
    expect(g.gas).toBeUndefined();
  });
});

describe("TestGasLossByKind", () => {
  it("weighs foot, vehicles and aircraft", () => {
    const g = gasScene(40, {
      0: { type: cInfantry, owner: 0, cell: 31 },
      2: { type: cArmCar, owner: 0, cell: 30 },
      3: { type: cFighter, owner: 0, cell: 32 },
    });
    g.gasLoss = 15;
    release(g, 31);
    const got = new Map(
      kinds(resolveOk(g), evGassed).map((e) => [e.unit, e.loss]),
    );
    expect([...got]).toStrictEqual([
      [0, 1],
      [2, 1],
    ]);
    expect(liveOf(g, 3).size).toBe(at(ExtendedTables.types, cFighter).group);
  });
});

describe("TestGasDestroys", () => {
  it("destroys a unit left with no strength", () => {
    const g = gasScene(40, {
      0: { type: cInfantry, owner: 0, cell: 31, size: 1 },
      2: { type: cInfantry, owner: 0, cell: 13 },
    });
    g.gasTurns = 4;
    release(g, 31);
    expect(gasAt(g, 31)).toBe(4);
    const ev = kinds(resolveOk(g), evGassed);
    expect(ev).toHaveLength(1);
    expect(at(ev, 0).loss).toBe(1);
    const u = liveOf(g, 0);
    expect(u.size).toBe(0);
    expect(u.flags4 & 0x8000).not.toBe(0);
    for (const k of [0, 1]) expect(occupant(g, k, 31)).not.toBe(0);
  });
});

describe("TestGasCharges", () => {
  it("spends the charges", () => {
    const g = gasScene(40, { 0: { type: cInfantry, owner: 0, cell: 4 } });
    release(g, 31);
    resolveOk(g);
    release(g, 39);
    expect(liveOf(g, 1).hold).toBe(0);
    resolveOk(g);
    refusal(
      "spent",
      check(g, { kind: cmdAttack, side: 1, unit: 1, cell: 31 }),
      msgNoGas,
    );
    expect(charges(Extended, cGas)).toStrictEqual([gasCharges, true]);
    expect(charges(Extended, cInfantry)[1]).toBe(false);
  });
});

describe("TestGasSave", () => {
  it("saves the cloud only when there is one", () => {
    const g = gasScene(40, { 0: { type: cInfantry, owner: 0, cell: 31 } });
    g.gasLoss = 15;
    let b = new TextDecoder().decode(save(g));
    expect(b).not.toContain('"gas"');
    release(g, 31);
    const raw = save(g);
    b = new TextDecoder().decode(raw);
    const h = loadGame(raw, Builds);
    expect(h.profile.tables).toBe(ExtendedTables);
    expect(h.gas).toStrictEqual(g.gas);
    expect(h.gasLoss).toBe(15);
    expect(new TextDecoder().decode(save(h))).toBe(b);
    const bad = JSON.parse(b) as { gas: number[] };
    bad.gas = bad.gas.slice(0, 3);
    expect(() => loadGame(JSON.stringify(bad), Builds)).toThrow(/gas on 3/);
  });
});
