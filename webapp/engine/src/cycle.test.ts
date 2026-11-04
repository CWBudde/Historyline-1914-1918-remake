// Twin of internal/dosrules/cycle_test.go: the command cycle on small scenes: validation, a whole round of moves,
// attacks and the phase change, destroyed targets, exits, and the attack step against the recorded order lists.
// BenchmarkDestinations is a Go benchmark and has no twin.

import { describe, expect, it } from "vitest";
import { destinations, destroyed, liveCell, zeroLive } from "./board.ts";
import { holdFlag40 } from "./container.ts";
import {
  type Event,
  type Game,
  ErrNotReady,
  attack,
  check,
  cmdAttack,
  cmdExit,
  cmdMove,
  cmdRepair,
  confirm,
  evAttack,
  evBattle,
  evMoved,
  evPhase,
  evSkipped,
  gameDestroyed,
  liveOf,
  movement,
  msgAttack,
  msgLimit,
  msgNoTarget,
  msgReadyUndone,
  newEvent,
  newPlayer,
  occupant,
  phaseAction,
  phaseMovement,
  resolve,
  unconfirm,
} from "./cycle.ts";
import { layoutDistance } from "./distance.ts";
import { terrainClass, zeroOutcome } from "./engage.ts";
import {
  type FixtureUnit,
  type OrdersFixture,
  engagementResult,
  ordersBattlefield,
  ordersList,
} from "./fixtures.ts";
import { at, attempt, errorsIs } from "./goruntime.ts";
import { type Profile, newProfile, setsUp, tablesOf } from "./profile.ts";
import { neighbours } from "./support.ts";
import { decodeOrdersFixture, load } from "./testing/fixtures.ts";
import {
  applyOk,
  cycleH,
  cycleW,
  energyCycle,
  hq0,
  hq1,
  inSlots,
  newCycle,
  newCycleHolds,
  refusal,
  resolveOk,
  transportHold,
} from "./testing/scenes.ts";
import {
  Profiles,
  cInfantry,
  cPanzer,
  cSupply,
  cTransportRec,
  moveTables,
} from "./testing/tables.ts";

describe("TestCycleValidation", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const n = neighbours(40, cycleW, cycleH);
    const g = newCycle(p, {
      0: { type: cInfantry, owner: 0, cell: 40 },
      1: { type: cInfantry, owner: 1, cell: at(n, 0) },
      2: { type: cInfantry, owner: 2, cell: 10 },
      3: { type: cInfantry, owner: 1, cell: at(n, 3) },
      4: { type: cInfantry, owner: 0, cell: 70 },
    });
    for (const [what, cmd, code] of [
      [
        "attack in the movement phase",
        { kind: cmdAttack, side: 0, unit: 0, cell: at(n, 0) },
        0,
      ],
      [
        "move in the action phase",
        { kind: cmdMove, side: 1, unit: 1, cell: 0 },
        0,
      ],
      [
        "the other side's unit",
        { kind: cmdMove, side: 0, unit: 1, cell: 0 },
        0,
      ],
      ["a neutral unit", { kind: cmdMove, side: 0, unit: 2, cell: 0 }, 0],
      ["no unit", { kind: cmdMove, side: 0, unit: 9, cell: 0 }, 0],
      [
        "attack on an empty cell",
        { kind: cmdAttack, side: 1, unit: 1, cell: 0 },
        msgNoTarget,
      ],
      [
        "attack on an own unit",
        { kind: cmdAttack, side: 1, unit: 1, cell: at(n, 3) },
        msgNoTarget,
      ],
      [
        "exit from an empty slot",
        { kind: cmdExit, side: 0, hold: cTransportRec, slot: 0, cell: 0 },
        0,
      ],
    ] as const) {
      refusal(what, check(g, cmd), code);
    }

    const atk = { kind: cmdAttack, side: 1, unit: 1, cell: 40 } as const;
    expect(check(g, atk)).toBeNull();
    const evs = applyOk(g, atk);
    const u = liveOf(g, 1);
    expect(evs).toStrictEqual([
      newEvent({
        kind: evAttack,
        side: 1,
        unit: 1,
        cell: 40,
        messages: [{ code: msgAttack, side: 1 }],
      }),
    ]);
    expect(u.flags4 & 0x200).not.toBe(0);
    expect([u.target, u.action[0], u.action[1]]).toStrictEqual([80, 9, 2]);
    expect(g.players[1].orders).toStrictEqual([1]);
    refusal("a second order", check(g, atk), 0);

    // The move limit counts the list; a computer side has none.
    g.limit = 1;
    const second = { kind: cmdAttack, side: 1, unit: 3, cell: 40 } as const;
    refusal("the move limit", check(g, second), msgLimit);
    g.players[1].computer = true;
    expect(check(g, second), "computer side at the limit").toBeNull();
    g.limit = 0;

    // A confirmed side gives no orders until it cancels.
    const mv = { kind: cmdMove, side: 0, unit: 0, cell: 41 } as const;
    confirm(g, 0);
    refusal("a confirmed side", check(g, mv), 0);
    const [, err] = attempt(() => resolve(g));
    expect(errorsIs(err, ErrNotReady)).toBe(true);
    const un = unconfirm(g, 0);
    expect(un).toHaveLength(1);
    expect(at(at(un, 0).messages, 0).code).toBe(msgReadyUndone);
    expect(check(g, mv), "move after the cancel").toBeNull();

    // A unit in a container is not on the map.
    const l = liveOf(g, 4);
    l.flags4 |= 0x4000;
    g.units.set(4, l);
    g.maps[0][2 * 70 + 1] = 0xff;
    refusal(
      "a held unit's move",
      check(g, { kind: cmdMove, side: 0, unit: 4, cell: 71 }),
      0,
    );
  });
});

// TestCycleTargets checks the assumed target rule: an attack needs an opposing unit of a class the attacker's
// target flags cover, within the range of the weapon for that class.
describe("TestCycleTargets", () => {
  const cAircraft = 2; // air range 1, attacks air, ground and water
  const cArtillery = 24; // surface range 7, attacks ground and water
  const cGround = 21; // surface range 1, attacks ground only
  const cBoat = 29; // a water unit
  it.each(Profiles)("%s", (_build, p) => {
    const n = neighbours(40, cycleW, cycleH);
    const far = 40 - 2 * cycleW; // two rows up: distance 2
    const g = newCycle(p, {
      0: { type: cInfantry, owner: 0, cell: 40 },
      1: { type: cAircraft, owner: 0, cell: at(n, 3) },
      2: { type: cInfantry, owner: 0, cell: 80 },
      3: { type: cBoat, owner: 0, cell: 70 },
      10: { type: cInfantry, owner: 1, cell: at(n, 0) },
      11: { type: cInfantry, owner: 1, cell: far },
      12: { type: cArtillery, owner: 1, cell: 0 },
      13: { type: cAircraft, owner: 1, cell: at(n, 2) },
      14: { type: cGround, owner: 1, cell: 71 },
      15: { type: cInfantry, owner: 1, cell: 62 },
    });
    expect([
      layoutDistance(p, cycleW, 70, 71),
      layoutDistance(p, cycleW, 71, 80),
      layoutDistance(p, cycleW, 71, 62),
    ]).toStrictEqual([1, 1, 1]);
    expect(layoutDistance(p, cycleW, 40, far)).toBe(2);
    expect(layoutDistance(p, cycleW, 0, 80)).toBeGreaterThan(7);
    for (const [what, unit, cell, ok] of [
      ["infantry next to its target", 10, 40, true],
      ["infantry two cells away", 11, 40, false],
      ["infantry against an aircraft", 10, at(n, 3), false],
      ["aircraft against an aircraft", 13, at(n, 3), true],
      ["aircraft against infantry", 13, 40, true],
      ["artillery in range", 12, 40, true],
      ["artillery out of range", 12, 80, false],
      ["artillery against an aircraft", 12, at(n, 3), false],
      ["a ground-only unit against a boat in range", 14, 70, false],
      ["a ground-only unit against infantry in range", 14, 80, true],
      ["a unit of its own side next to it", 14, 62, false],
    ] as const) {
      const err = check(g, { kind: cmdAttack, side: 1, unit, cell });
      if (ok) {
        expect(err, what).toBeNull();
      } else {
        refusal(what, err, msgNoTarget);
      }
    }
  });
});

// TestCycleRound plays a round: side 0 moves its unit away while side 1 attacks it where side 1's map still shows
// it; both confirm, the attack is fought at the old cell, and the phase change brings the maps into line.
describe("TestCycleRound", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const n = neighbours(40, cycleW, cycleH);
    const g = newCycle(p, {
      0: { type: cInfantry, owner: 0, cell: 40 },
      1: { type: cPanzer, owner: 1, cell: at(n, 0) },
    });
    const dests = destinations(p, movement(g, 0, 0));
    const dest = dests.find((c) => !n.includes(c) && c !== 40) ?? -1;
    expect(
      dest,
      `no destination away from cell 40 in ${dests.join(" ")}`,
    ).toBeGreaterThanOrEqual(0);
    let evs = applyOk(g, { kind: cmdMove, side: 0, unit: 0, cell: dest });
    expect(evs).toHaveLength(1);
    expect(at(evs, 0).kind).toBe(evMoved);
    expect(at(evs, 0).route.at(-1)).toBe(dest);
    expect([
      occupant(g, 0, dest),
      occupant(g, 0, 40),
      occupant(g, 1, 40),
      occupant(g, 1, dest),
    ]).toStrictEqual([0, 0xff, 0, 0xff]);
    expect(g.players[0].orders).toStrictEqual([0]);
    expect(liveOf(g, 0).flags4 & 0x200).not.toBe(0);
    refusal(
      "a second move",
      check(g, { kind: cmdMove, side: 0, unit: 0, cell: 40 }),
      0,
    );

    applyOk(g, { kind: cmdAttack, side: 1, unit: 1, cell: 40 });
    evs = resolveOk(g);
    expect(evs.map((e) => e.kind)).toStrictEqual([evBattle, evPhase]);
    const o = at(evs, 0).outcome;
    expect(o?.defender).toBe(0);
    expect(o?.returnFire).toBe(true);
    if (o === null) {
      return;
    }
    expect([liveOf(g, 0).size, liveOf(g, 1).size]).toStrictEqual([
      o.defenderSize,
      o.attackerSize,
    ]);
    expect(at(evs, 1).phase).toStrictEqual([phaseAction, phaseMovement]);
    expect(at(evs, 1).turn).toBe(0);
    expect(g.maps[0], "the maps differ after the phase change").toStrictEqual(
      g.maps[1],
    );
    if (o.defenderSize > 0) {
      expect(occupant(g, 1, dest)).toBe(0);
      expect(liveCell(liveOf(g, 0), 1)).toBe(dest);
    }
    for (const [i, u] of g.units) {
      expect(u.flags4 & 0x200, `unit ${i} keeps its moved flag`).toBe(0);
    }
    for (const pl of g.players) {
      expect(pl.ready).toBe(false);
      expect(pl.orders).toStrictEqual([]);
    }

    // The next change brings side 0 back to the movement phase and counts the turn.
    evs = resolveOk(g);
    expect(at(evs, -1 + evs.length).phase).toStrictEqual([
      phaseMovement,
      phaseAction,
    ]);
    expect(g.turn).toBe(1);
  });
});

// TestCycleDestroyedTarget gives two attacks on one target: where the first destroys it, the second is skipped,
// and the phase change marks it.
describe("TestCycleDestroyedTarget", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const n = neighbours(40, cycleW, cycleH);
    let skipped = 0;
    for (let size = 1; size <= 6; size++) {
      const g = newCycle(p, {
        0: { type: cInfantry, owner: 0, cell: 40, size },
        1: { type: cPanzer, owner: 1, cell: at(n, 0) },
        2: { type: cPanzer, owner: 1, cell: at(n, 3) },
      });
      applyOk(g, { kind: cmdAttack, side: 1, unit: 1, cell: 40 });
      applyOk(g, { kind: cmdAttack, side: 1, unit: 2, cell: 40 });
      const evs = resolveOk(g);
      expect(at(evs, 0).kind, `size ${size}`).toBe(evBattle);
      expect(at(evs, 2).kind, `size ${size}`).toBe(evPhase);
      if (!(at(evs, 0).outcome?.destroyed ?? []).includes(0)) {
        expect(
          at(evs, 1).kind,
          `size ${size}: the surviving target is not fought again`,
        ).toBe(evBattle);
        continue;
      }
      skipped++;
      expect(at(evs, 1).kind, `size ${size}`).toBe(evSkipped);
      // The phase change marks it 0x8000 and clears the action bytes, the battle's bit 4 of +0x14 too.
      const u = liveOf(g, 0);
      expect(u.flags4 & 0x8000).not.toBe(0);
      expect(u.action).toStrictEqual([0, 0, 0, 0, 0, 0, 0]);
      expect([occupant(g, 0, 40), occupant(g, 1, 40)]).toStrictEqual([
        0xff, 0xff,
      ]);
      expect(gameDestroyed(g)).toStrictEqual([0]);
    }
    expect(skipped, "no first attack destroyed the target").toBeGreaterThan(0);
  });
});

// TestCycleExit takes a unit out of a transporter during the movement phase.
describe("TestCycleExit", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const n = neighbours(40, cycleW, cycleH);
    const g = newCycleHolds(
      p,
      {
        1: { type: cSupply, owner: 0, cell: 40 },
        10: { type: cInfantry, owner: 0 },
      },
      { [cTransportRec]: transportHold(40, 0, 1, 10) },
    );
    refusal(
      "a held unit's move",
      check(g, { kind: cmdMove, side: 0, unit: 10, cell: at(n, 0) }),
      0,
    );
    const evs = applyOk(g, {
      kind: cmdExit,
      side: 0,
      hold: cTransportRec,
      slot: 0,
      cell: at(n, 0),
    });
    expect(evs).toHaveLength(1);
    expect([at(evs, 0).kind, at(evs, 0).unit]).toStrictEqual([evMoved, 10]);
    const u = liveOf(g, 10);
    expect(g.holds.get(cTransportRec)?.slots[0]).toBe(0xff);
    expect([occupant(g, 0, at(n, 0)), occupant(g, 0, 40)]).toStrictEqual([
      10, 1,
    ]);
    expect(u.flags4 & 0x4000).toBe(0);
    expect(u.flags4 & 0x200).not.toBe(0);
    expect(g.players[0].orders).toStrictEqual([10]);
    resolveOk(g);
    expect([occupant(g, 1, at(n, 0)), occupant(g, 1, 40)]).toStrictEqual([
      10, 1,
    ]);
  });
});

// ordersGame is the state an order-list fixture gives the handler: the attacking side's map with the units and
// terrain classes, its list, and each attacker's target in +0x11.
function ordersGame(p: Profile, f: OrdersFixture): Game {
  const b = ordersBattlefield(f);
  const g: Game = {
    profile: p,
    width: f.width,
    height: f.height,
    maps: [[], []],
    units: new Map(),
    holds: new Map(),
    players: [newPlayer(), newPlayer()],
    turn: 0,
    limit: 0,
    concealed: false,
    ended: null,
    energy: [0, 0],
    buildings: new Map(),
  };
  for (const m of g.maps) {
    for (let c = 0; c < f.width * f.height; c++) {
      m.push(terrainClass(b, c), 0xff);
    }
  }
  for (const [i, u] of b.units) {
    const l = {
      ...zeroLive(),
      size: u.size,
      alt: u.alt,
      exp: u.exp,
      flags4: u.flags4,
      flags6: u.flags6,
      type: u.type,
      hold: 0xff,
    };
    l.pos[f.side] = 2 * u.cell + 1;
    g.units.set(i, l);
    at(g.maps, f.side)[2 * u.cell + 1] = i;
  }
  for (const o of ordersList(f)) {
    const l = liveOf(g, o.attacker);
    l.target = 2 * o.target;
    g.units.set(o.attacker, l);
    at(g.players, f.side).orders.push(o.attacker);
  }
  return g;
}

// TestCycleOrdersFixtures runs the cycle's attack step on every recorded order list (orders.json): the terrain
// records stand for their combat class, as in the recorded scenes, and the outcomes, strengths, destroyed units
// and the attacking side's map must be the recorded ones.
describe("TestCycleOrdersFixtures", () => {
  it.each(Profiles)("%s", (build, p0) => {
    const tables = tablesOf(p0);
    const mt = moveTables(p0);
    const p = newProfile({
      ...p0,
      tables: {
        ...tables,
        move: {
          ...mt,
          terrains: Array.from({ length: 9 }, (_, k) => ({
            flags: 0,
            mask: 0,
            cost: 0,
            airCost: 0,
            class: k,
          })),
        },
      },
    });
    const fs = load(build, "orders.json", decodeOrdersFixture);
    expect(fs).toHaveLength(183);
    for (const f of fs) {
      const g = ordersGame(p, f);
      const evs = attack(g, f.side);
      const units: FixtureUnit[] = [];
      const gone: number[] = [];
      for (const fu of f.units ?? []) {
        const u = liveOf(g, fu.index);
        units.push({
          ...fu,
          size: u.size,
          alternateSize: u.alt,
          experience: u.exp,
        });
        if ((at(u.action, 1) & 4) !== 0) {
          gone.push(fu.index);
        }
      }
      const occupants: [number, number][] = [];
      for (let c = 0; c < f.width * f.height; c++) {
        const occ = occupant(g, f.side, c);
        if (occ !== 0xff) {
          occupants.push([c, occ]);
        }
      }
      expect(f.expected.error ?? "", f.id).toBe("");
      expect(
        {
          orders: evs.map((ev: Event) =>
            engagementResult(ev.outcome ?? zeroOutcome(), null),
          ),
          units,
          gone,
          occupants,
        },
        f.id,
      ).toStrictEqual({
        orders: f.expected.orders,
        units: f.expected.units,
        gone: f.expected.gone,
        occupants: f.expected.occupants,
      });
    }
  });
});

// A two-cell ship's rear takes no orders: it moves and fights with its head.
describe("TestCycleShipRear", () => {
  const cShip = 32; // a head; its rear is type 33
  it.each(Profiles)("%s", (_build, p) => {
    const g = newCycle(p, {
      0: { type: cShip, owner: 0, cell: 20 },
      1: { type: cShip + 1, owner: 0, cell: 11 },
      2: { type: cShip, owner: 1, cell: 60 },
      3: { type: cShip + 1, owner: 1, cell: 69 },
    });
    refusal(
      "the rear's move",
      check(g, { kind: cmdMove, side: 0, unit: 1, cell: 12 }),
      0,
    );
    refusal(
      "the rear's attack",
      check(g, { kind: cmdAttack, side: 1, unit: 3, cell: 20 }),
      0,
    );
    expect(
      check(g, { kind: cmdMove, side: 0, unit: 0, cell: 21 }),
      "the head's move",
    ).toBeNull();
    // The head passes the rear check; the empty cell holds no target.
    refusal(
      "the head's attack",
      check(g, { kind: cmdAttack, side: 1, unit: 2, cell: 0 }),
      msgNoTarget,
    );
  });
});

describe("TestCycleSetUp", () => {
  const cLight = 10; // light artillery: fires after a move
  const cMedium = 11; // medium artillery: must set up
  const cHeavy = 12; // heavy artillery: must set up
  it.each(Profiles)("%s", (_build, p) => {
    const types = tablesOf(p).types;
    expect([
      setsUp(at(types, cLight)),
      setsUp(at(types, cMedium)),
      setsUp(at(types, cHeavy)),
    ]).toStrictEqual([false, true, true]);
    const g = newCycleHolds(
      p,
      {
        0: { type: cHeavy, owner: 0, cell: 20 },
        1: { type: cMedium, owner: 0, cell: 24 },
        2: { type: cLight, owner: 0, cell: 60 },
        3: { type: cSupply, owner: 0, cell: 64 },
        4: { type: cHeavy, owner: 0 },
        10: { type: cInfantry, owner: 1, cell: 78 },
      },
      { [cTransportRec]: transportHold(64, 0, 3, 4) },
    );
    // The heavy and the light artillery move, the medium one stays and the second heavy one leaves its
    // transporter.
    for (const u of [0, 2]) {
      const dests = destinations(p, movement(g, 0, u));
      const d = dests.find((c) => c !== liveCell(liveOf(g, u), 0));
      expect(d, `unit ${u} has no destination`).toBeDefined();
      applyOk(g, { kind: cmdMove, side: 0, unit: u, cell: d ?? 0 });
    }
    applyOk(g, {
      kind: cmdExit,
      side: 0,
      hold: cTransportRec,
      slot: 0,
      cell: 65,
    });
    resolveOk(g);
    expect(g.players[0].phase).toBe(phaseAction);
    for (const [u, want] of [
      [0, true],
      [1, false],
      [2, false],
      [3, false],
      [4, true],
    ] as const) {
      expect(
        (liveOf(g, u).flags4 & 0x200) !== 0,
        `unit ${u} (type ${liveOf(g, u).type}) holds its fire`,
      ).toBe(want);
    }
    refusal(
      "the moved heavy artillery's attack",
      check(g, { kind: cmdAttack, side: 0, unit: 0, cell: 78 }),
      0,
    );

    // It sets up for that one phase only.
    resolveOk(g);
    resolveOk(g);
    for (const [u, l] of g.units) {
      expect(
        l.flags4 & 0x200,
        `unit ${u} still holds its fire two changes later`,
      ).toBe(0);
    }
  });
});

// TestCycleDestroyedLeavesBuilding moves a unit into its HQ while the enemy's attacks hit it at the cell it
// left: destroyed there, it is in no slot after the phase change and cannot be repaired.
describe("TestCycleDestroyedLeavesBuilding", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const from =
      neighbours(hq0, cycleW, cycleH).find((c) => c >= 0 && c !== hq1) ?? -1;
    const units: Record<
      number,
      { type: number; owner: number; cell: number; size?: number }
    > = {
      0: { type: cInfantry, owner: 0, cell: from, size: 1 },
      9: { type: cInfantry, owner: 0, cell: 78 },
    };
    const shooters: number[] = [];
    for (const c of neighbours(from, cycleW, cycleH)) {
      if (c >= 0 && c !== hq0 && shooters.length < 3) {
        shooters.push(c);
        units[shooters.length] = { type: cPanzer, owner: 1, cell: c };
      }
    }
    const g = energyCycle(p, units);
    g.players[0].phase = phaseMovement;
    g.players[1].phase = phaseAction;
    applyOk(g, { kind: cmdMove, side: 0, unit: 0, cell: hq0 });
    expect(inSlots(g, 0), "the move did not take the unit into the HQ").toBe(
      true,
    );
    for (const k of shooters.keys()) {
      applyOk(g, { kind: cmdAttack, side: 1, unit: k + 1, cell: from });
    }
    resolveOk(g);
    expect(destroyed(liveOf(g, 0)), "the attacks left the unit").toBe(true);
    expect(inSlots(g, 0), "the destroyed unit is still in a slot").toBe(false);
    g.energy[0] = 100;
    for (let k = 0; k < 7; k++) {
      expect(
        check(g, {
          kind: cmdRepair,
          side: 0,
          hold: holdFlag40,
          slot: k,
          cell: 0,
        }),
        `slot ${k} offers a repair`,
      ).not.toBeNull();
    }
  });
});

// TestCycleDestroyedTransporterCargo destroys a loaded transporter: its cargo is destroyed with it.
describe("TestCycleDestroyedTransporterCargo", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const units: Record<
      number,
      { type: number; owner: number; cell?: number; size?: number }
    > = {
      1: { type: cSupply, owner: 0, cell: 40, size: 1 },
      10: { type: cInfantry, owner: 0 },
    };
    const n = neighbours(40, cycleW, cycleH);
    for (const [k, c] of n.slice(0, 3).entries()) {
      units[20 + k] = { type: cPanzer, owner: 1, cell: c };
    }
    const g = newCycleHolds(p, units, {
      [cTransportRec]: transportHold(40, 0, 1, 10),
    });
    for (let k = 0; k < 3; k++) {
      applyOk(g, { kind: cmdAttack, side: 1, unit: 20 + k, cell: 40 });
    }
    resolveOk(g);
    expect(destroyed(liveOf(g, 1)), "the attacks left the transporter").toBe(
      true,
    );
    expect(destroyed(liveOf(g, 10)), "the cargo survived").toBe(true);
    expect(inSlots(g, 10), "the cargo is still in a slot").toBe(false);
    expect(gameDestroyed(g)).toStrictEqual([1, 10]);
  });
});
