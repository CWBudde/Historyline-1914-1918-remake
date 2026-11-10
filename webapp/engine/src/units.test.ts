// Twin of internal/dosrules/units_test.go: the extension types beyond the gas pioneers, the flamethrower, the
// wiring party, the tunnellers and the field ambulance.

import { describe, expect, it } from "vitest";
import { ambulanceHeal, evHealed, flagMedic } from "./ambulance.ts";
import { destroyed } from "./board.ts";
import { casualties } from "./casualties.ts";
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
import { flagGas } from "./gas.ts";
import { at } from "./goruntime.ts";
import { setsUp } from "./profile.ts";
import { loadGame, save } from "./save.ts";
import { trench, trenchFirst, worked } from "./sapper.ts";
import { applyOk, newCycle, refusal, resolveOk } from "./testing/scenes.ts";
import {
  Builds,
  Extended,
  ExtendedTables,
  FullTables,
  cInfantry,
} from "./testing/tables.ts";
import {
  evMine,
  evMined,
  evTunnel,
  evTunnelLost,
  flagTunnel,
  msgTunnel,
  msgTunnelBusy,
  tunnelLoss,
} from "./tunnel.ts";
import {
  flagFlame,
  flagWire,
  msgNoWire,
  msgWire,
  terrainCrater,
  terrainWire,
  typeFlag,
  wire,
  wireCells,
  wireCharges,
  evWired,
} from "./wire.ts";

// The extension types beyond the gas pioneers (extension.json).
const cMarkV = 56;
const cWhippet = 55;
const cZeppelin = 54;
const cParisGun = 58;
const cFlame = 59;
const cWiring = 60;
const cTunnel = 61;
const cAmbulance = 62;
const cCavalry = 21;
const cArmCar = 0;

function kinds(evs: Event[], k: EventKind): Event[] {
  return evs.filter((e) => e.kind === k);
}

describe("TestExtensionRecords", () => {
  it("appends the type and terrain records", () => {
    const x = ExtendedTables,
      f = FullTables;
    const xm = x.move,
      fm = f.move;
    if (xm === null || fm === null || xm.winter === null || fm.winter === null)
      throw new Error("no movement tables");
    expect(x.types).toHaveLength(63);
    expect(xm.types).toHaveLength(63);
    expect(xm.terrains).toHaveLength(fm.terrains.length + 2);
    expect(xm.winter).toHaveLength(fm.winter.length + 2);
    expect(xm.terrains.slice(0, fm.terrains.length)).toStrictEqual(fm.terrains);
    expect(xm.winter.slice(0, fm.winter.length)).toStrictEqual(fm.winter);
    expect(fm.terrains).toHaveLength(terrainWire);
    for (const m of fm.types) {
      expect(
        m.flags & (flagFlame | flagWire | flagTunnel | flagMedic | flagGas),
      ).toBe(0);
    }
    for (const [typ, flag] of [
      [cFlame, flagFlame],
      [cWiring, flagWire],
      [cTunnel, flagTunnel],
      [cAmbulance, flagMedic],
    ] as const) {
      expect(typeFlag(Extended, typ, flag), `type ${String(typ)}`).toBe(true);
    }
    // Wire lets foot units and heavy tanks in, not cavalry or light tanks.
    const w = at(xm.terrains, terrainWire);
    for (const [typ, inside] of [
      [cInfantry, true],
      [cMarkV, true],
      [cCavalry, false],
      [cWhippet, false],
      [cWiring, true],
    ] as const) {
      expect(
        (w.mask & at(xm.types, typ).mask) !== 0,
        `type ${String(typ)} on wire`,
      ).toBe(inside);
    }
    expect(w.cost).toBeGreaterThanOrEqual(at(xm.types, cInfantry).budget);
    expect(at(x.types, cParisGun).surfaceRange).toBeGreaterThan(
      at(x.types, 24).surfaceRange,
    );
    expect(setsUp(at(x.types, cParisGun))).toBe(true);
    expect(at(xm.types, cZeppelin).class & 0x10).not.toBe(0);
  });
});

describe("TestFlameIgnoresTerrain", () => {
  it("strips the defender's cover, not the return fire", () => {
    const terms = ExtendedTables.terms;
    let cls = -1;
    for (const [c, tm] of terms.entries()) {
      if (tm.defence > 0 && (cls < 0 || tm.defence > at(terms, cls).defence)) {
        cls = c;
      }
    }
    expect(cls).toBeGreaterThanOrEqual(0);
    const ty = ExtendedTables.types;
    const a = { type: at(ty, cFlame), size: 6, alt: 0, exp: 0, flags: 1 | 4 };
    const d = { type: at(ty, cInfantry), size: 6, alt: 0, exp: 0, flags: 4 };
    const ctx = {
      distance: 1,
      aTerrain: 8,
      dTerrain: cls,
      pincer: 0,
      blocking: 0,
      returnFire: true,
      flame: false,
    };
    const plain = casualties(Extended, a, d, ctx);
    const flame = casualties(Extended, a, d, { ...ctx, flame: true });
    expect(flame.defenderSize).toBeLessThan(plain.defenderSize);
    expect(flame.attackerSize).toBe(plain.attackerSize);
  });
});

// workScene is an Extended game on worked open ground (the grass record 0) with side 1's unit 1 of typ at from
// and the units of more.
function workScene(
  typ: number,
  from: number,
  more: Record<number, Placement>,
): Game {
  const g = newCycle(Extended, {
    1: { type: typ, owner: 1, cell: from },
    ...more,
  });
  for (const m of g.maps) {
    for (let c = 0; c < g.width * g.height; c++) {
      m[2 * c] = 0;
    }
  }
  return g;
}

// order gives unit 1 of side 1 its work order on cell with message msg.
function order(g: Game, cell: number, msg: number): void {
  const evs = applyOk(g, { kind: cmdAttack, side: 1, unit: 1, cell });
  expect(evs).toHaveLength(1);
  expect(at(evs, 0).kind).toBe(evWork);
  expect(at(at(evs, 0).messages, 0).code).toBe(msg);
  expect(liveOf(g, 1).flags4 & 0x200).not.toBe(0);
  expect(g.players[1].orders).toStrictEqual([]);
}

describe("TestWire", () => {
  it("lays wire on worked ground and spends a charge", () => {
    const g = workScene(cWiring, 40, {
      0: { type: cInfantry, owner: 0, cell: 31 },
    });
    expect(liveOf(g, 1).hold).toBe(wireCharges);
    expect(wireCells(g, 1, 1, 31)).toStrictEqual([30, 31, 32]);
    g.maps[1][2 * 32] = trenchFirst;
    expect(wireCells(g, 1, 1, 31)).toStrictEqual([30, 31]);
    refusal(
      "a trench",
      check(g, { kind: cmdAttack, side: 1, unit: 1, cell: 32 }),
      msgNoTarget,
    );
    refusal(
      "beyond the neighbours",
      check(g, { kind: cmdAttack, side: 1, unit: 1, cell: 42 }),
      msgNoTarget,
    );
    order(g, 31, msgWire);
    const evs = resolveOk(g);
    const w = kinds(evs, evWired);
    expect(w).toHaveLength(1);
    expect(at(w, 0).cell).toBe(31);
    expect(at(w, 0).cells).toStrictEqual([30, 31]);
    for (const m of g.maps) {
      expect(wire(at(m, 2 * 30))).toBe(true);
      expect(wire(at(m, 2 * 31))).toBe(true);
    }
    expect(trench(at(g.maps[1], 2 * 32))).toBe(true);
    expect(liveOf(g, 1).hold).toBe(wireCharges - 1);
    expect(liveOf(g, 0).size).toBe(at(ExtendedTables.types, cInfantry).group);
    resolveOk(g);
    refusal(
      "wire already",
      check(g, { kind: cmdAttack, side: 1, unit: 1, cell: 31 }),
      msgNoTarget,
    );
    const u = liveOf(g, 1);
    u.hold = 0;
    g.units.set(1, u);
    refusal(
      "no charge",
      check(g, { kind: cmdAttack, side: 1, unit: 1, cell: 39 }),
      msgNoWire,
    );
    // A sapper clears wire as it fills a trench.
    expect(worked(g, terrainWire)).toBe(true);
    expect(worked(g, terrainCrater)).toBe(true);
  });
});

// tunnelScene has side 1's tunnellers at 40 under side 0's infantry 0 at 31 (size size) and side 0's infantry 2
// at 22.
function tunnelScene(size: number): Game {
  return workScene(cTunnel, 40, {
    0: { type: cInfantry, owner: 0, cell: 31, size },
    2: { type: cInfantry, owner: 0, cell: 22 },
  });
}

describe("TestTunnel", () => {
  it("blows the mine after two turns", () => {
    const g = tunnelScene(0);
    const full = at(ExtendedTables.types, cInfantry).group;
    order(g, 31, msgTunnel);
    let evs = resolveOk(g); // side 1
    const tn = kinds(evs, evTunnel);
    expect(tn).toHaveLength(1);
    expect(at(tn, 0).cell).toBe(31);
    expect(g.tunnels).toHaveLength(1);
    expect(kinds(evs, evBattle)).toHaveLength(0);
    resolveOk(g); // side 0: the first turn
    expect(at(g.tunnels ?? [], 0).dug).toBe(1);
    refusal(
      "busy",
      check(g, { kind: cmdAttack, side: 1, unit: 1, cell: 39 }),
      msgTunnelBusy,
    );
    resolveOk(g); // side 1
    evs = resolveOk(g); // side 0: the second turn blows it
    const mine = kinds(evs, evMine);
    expect(mine).toHaveLength(1);
    expect(at(mine, 0).cell).toBe(31);
    expect(at(mine, 0).cells).toStrictEqual([31]);
    expect(g.tunnels).toBeUndefined();
    const loss = Math.trunc((full * tunnelLoss + 99) / 100);
    expect(
      kinds(evs, evMined).map((e) => [e.unit, e.loss, e.side]),
    ).toStrictEqual([[0, loss, 0]]);
    expect(liveOf(g, 0).size).toBe(full - loss);
    expect(liveOf(g, 2).size).toBe(full);
    for (const m of g.maps) expect(at(m, 2 * 31)).toBe(terrainCrater);
    expect(
      check(g, { kind: cmdAttack, side: 1, unit: 1, cell: 39 }),
    ).toBeNull();
  });
});

describe("TestTunnelDestroys", () => {
  it("destroys a unit left with no strength", () => {
    const g = tunnelScene(2);
    order(g, 31, msgTunnel);
    for (let k = 0; k < 4; k++) resolveOk(g);
    expect(destroyed(liveOf(g, 0))).toBe(true);
    expect(occupant(g, 0, 31)).not.toBe(0);
  });
});

describe("TestTunnelLost", () => {
  it("closes the tunnel of lost tunnellers", () => {
    const g = tunnelScene(0);
    order(g, 31, msgTunnel);
    resolveOk(g);
    const u = liveOf(g, 1);
    u.size = 0;
    g.units.set(1, u);
    const evs = resolveOk(g);
    const l = kinds(evs, evTunnelLost);
    expect(l).toHaveLength(1);
    expect(at(l, 0).unit).toBe(1);
    expect(g.tunnels).toBeUndefined();
    expect(kinds(evs, evMine)).toHaveLength(0);

    const h = tunnelScene(0);
    for (const m of h.maps) m[2 * 39] = 3; // water
    refusal(
      "water",
      check(h, { kind: cmdAttack, side: 1, unit: 1, cell: 39 }),
      msgNoTarget,
    );
  });
});

describe("TestAmbulance", () => {
  it("heals the foot units next to it once a turn", () => {
    const g = workScene(cAmbulance, 40, {
      0: { type: cInfantry, owner: 1, cell: 31, size: 3 }, // next to it: heals
      2: { type: cInfantry, owner: 1, cell: 22, size: 3 }, // too far
      3: { type: cInfantry, owner: 0, cell: 39, size: 3 }, // the enemy
      4: { type: cArmCar, owner: 1, cell: 41, size: 3 }, // no foot unit
      5: { type: cInfantry, owner: 1, cell: 49 }, // at full strength
    });
    expect(kinds(resolveOk(g), evHealed)).toHaveLength(0);
    const h = kinds(resolveOk(g), evHealed);
    expect(h.map((e) => [e.unit, e.gain, e.side, e.cell])).toStrictEqual([
      [0, ambulanceHeal, 1, 31],
    ]);
    for (const [i, want] of [
      [0, 3 + ambulanceHeal],
      [2, 3],
      [3, 3],
      [4, 3],
    ] as const) {
      expect(liveOf(g, i).size, `unit ${String(i)}`).toBe(want);
    }
    expect(kinds(resolveOk(g), evBattle)).toHaveLength(0);
  });
});

describe("TestTunnelSave", () => {
  it("saves, loads and validates the tunnels", () => {
    const g = tunnelScene(0);
    order(g, 31, msgTunnel);
    resolveOk(g);
    const raw = save(g);
    const h = loadGame(raw, Builds);
    expect(h.tunnels).toStrictEqual(g.tunnels);
    expect(h.tunnels).toHaveLength(1);
    const b = new TextDecoder().decode(raw);
    expect(new TextDecoder().decode(save(h))).toBe(b);
    const bad = JSON.parse(b) as { tunnels: { cell: number }[] };
    at(bad.tunnels, 0).cell = -1;
    expect(() => loadGame(JSON.stringify(bad), Builds)).toThrow(
      /dosrules: load: tunnel \{Unit:1 Side:1 From:40 Cell:-1 Dug:0\}/,
    );
  });
});
