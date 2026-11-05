// Twin of internal/dosrules/victory_test.go: the end of a battle: the HQ held, annihilation, a draw, and the
// score.

import { describe, expect, it } from "vitest";
import {
  type Event,
  ErrOver,
  Refusal,
  check,
  cmdMove,
  cmdAttack,
  confirm,
  evEnd,
  evPhase,
  gameDestroyed,
  resolve,
} from "./cycle.ts";
import { at, attempt, errorsIs } from "./goruntime.ts";
import { tablesOf } from "./profile.ts";
import { neighbours } from "./support.ts";
import {
  type End,
  end,
  endAnnihilation,
  endHQ,
  endSurrender,
  finalScore,
  score,
  surrender,
} from "./victory.ts";
import {
  applyOk,
  armour,
  cycleH,
  cycleW,
  hq0,
  hq1,
  hqCycle,
  hqCycleHolds,
  reload,
  resolveOk,
  transportHold,
} from "./testing/scenes.ts";
import { Profiles, cInfantry, cPanzer, cSupply } from "./testing/tables.ts";

function endOf(evs: readonly Event[]): End {
  const e = at(evs, evs.length - 1);
  expect(e.kind, "no end event").toBe(evEnd);
  expect(e.end).not.toBeNull();
  if (e.end === null) {
    throw new Error("no end event");
  }
  expect(e.side, "end event side, winner").toBe(e.end.winner);
  return e.end;
}

describe("TestVictoryOpen", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const g = hqCycleHolds(
      p,
      {
        0: { type: cInfantry, owner: 0, cell: hq0 },
        1: { type: cPanzer, owner: 1, cell: hq1 },
        2: { type: cSupply, owner: 2, cell: 40 },
      },
      { 0: transportHold(40, 2, 2) },
    );
    expect(end(g, -1)).toStrictEqual({
      over: false,
      winner: -1,
      won: ["", ""],
      score: 0,
    });
    expect(g.ended).toBeNull();
    let want = 0;
    for (const ty of [cInfantry, cPanzer, cSupply]) {
      want += at(tablesOf(p).types, ty).armour;
    }
    expect(score(g)).toBe(want);
    expect(armour(g)).toBe(want);
    const evs = resolveOk(g);
    expect(at(evs, evs.length - 1).kind, "an open battle ends").toBe(evPhase);
  });
});

describe("TestVictoryAnnihilation", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const n = neighbours(40, cycleW, cycleH);
    for (let size = 1; size <= 6; size++) {
      let g = hqCycle(p, {
        0: { type: cInfantry, owner: 0, cell: 40, size },
        1: { type: cPanzer, owner: 1, cell: at(n, 0) },
        2: { type: cPanzer, owner: 1, cell: at(n, 3) },
      });
      applyOk(g, { kind: cmdAttack, side: 1, unit: 1, cell: 40 });
      applyOk(g, { kind: cmdAttack, side: 1, unit: 2, cell: 40 });
      const evs = resolveOk(g);
      const d = gameDestroyed(g);
      if (d.length !== 1 || d[0] !== 0) {
        expect(
          at(evs, evs.length - 1).kind,
          `size ${size}: ends with side 0 alive`,
        ).not.toBe(evEnd);
        continue;
      }
      const r = endOf(evs);
      expect([r.over, r.winner, r.won]).toStrictEqual([
        true,
        1,
        ["", endAnnihilation],
      ]);
      expect(r.score).toBe(armour(g) + 500);
      expect(score(g)).toBe(armour(g));
      expect(g.ended).toBe(r);
      const h = reload(g);
      expect(h, "an ended battle reloads equal").toStrictEqual(g);
      g = h;
      expect(
        check(g, { kind: cmdMove, side: 1, unit: 1, cell: 0 }),
        "a command after the end",
      ).toBeInstanceOf(Refusal);
      confirm(g, 0);
      confirm(g, 1);
      const [, err] = attempt(() => resolve(g));
      expect(errorsIs(err, ErrOver), "a resolution after the end").toBe(true);
      return;
    }
    expect.fail("no attack destroyed side 0's unit");
  });
});

// TestVictoryHQ captures the opposing HQ: the final score counts the concealed inventory and the move limit.
describe("TestVictoryHQ", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const n = neighbours(hq1, cycleW, cycleH);
    const g = hqCycle(p, {
      0: { type: cInfantry, owner: 0, cell: at(n, 0) },
      1: { type: cPanzer, owner: 1, cell: 40 },
    });
    g.limit = 8;
    g.concealed = true;
    applyOk(g, { kind: cmdMove, side: 0, unit: 0, cell: hq1 });
    const r = endOf(resolveOk(g));
    expect([r.over, r.winner, r.won]).toStrictEqual([true, 0, [endHQ, ""]]);
    expect(r.score).toBe((armour(g) + 100) * 3);
  });
});

// TestVictoryHQStanding: a unit placed on the opposing HQ has not entered it and does not win.
describe("TestVictoryHQStanding", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const g = hqCycle(p, {
      0: { type: cInfantry, owner: 0, cell: hq1 },
      1: { type: cPanzer, owner: 1, cell: 40 },
    });
    const evs = resolveOk(g);
    expect(at(evs, evs.length - 1).kind).not.toBe(evEnd);
    expect(g.ended).toBeNull();
  });
});

// TestVictoryHQMove reaches the opposing HQ by a move order.
describe("TestVictoryHQMove", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const n = neighbours(hq1, cycleW, cycleH);
    const g = hqCycle(p, {
      0: { type: cInfantry, owner: 0, cell: at(n, 0) },
      1: { type: cPanzer, owner: 1, cell: 40 },
    });
    applyOk(g, { kind: cmdMove, side: 0, unit: 0, cell: hq1 });
    expect(g.ended, "the battle ends before the resolution").toBeNull();
    const r = endOf(resolveOk(g));
    expect(r.winner).toBe(0);
    expect(r.won[0]).toBe(endHQ);
  });
});

describe("TestVictoryOwnHQ", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const g = hqCycle(p, {
      0: { type: cInfantry, owner: 0, cell: hq0 },
      1: { type: cPanzer, owner: 1, cell: hq1 },
    });
    const evs = resolveOk(g);
    expect(at(evs, evs.length - 1).kind).not.toBe(evEnd);
    expect(g.ended, "a side wins on its own HQ").toBeNull();
  });
});

// TestVictoryHQOverAnnihilation: the HQ captured wins as such, without the annihilation bonus, even where the
// other side has no unit left.
describe("TestVictoryHQOverAnnihilation", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const n = neighbours(hq1, cycleW, cycleH);
    const g = hqCycle(p, { 0: { type: cInfantry, owner: 0, cell: at(n, 0) } });
    applyOk(g, { kind: cmdMove, side: 0, unit: 0, cell: hq1 });
    const r = endOf(resolveOk(g));
    expect(r).toStrictEqual({
      over: true,
      winner: 0,
      won: [endHQ, ""],
      score: armour(g),
    });
  });
});

// TestVictorySurrender: a side gives up; the other wins by surrender at the current score, without the
// annihilation bonus, and the ended battle takes no further command or resolution.
describe("TestVictorySurrender", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const g = hqCycle(p, {
      0: { type: cInfantry, owner: 0, cell: 40 },
      1: { type: cPanzer, owner: 1, cell: 41 },
    });
    g.limit = 8;
    g.concealed = true;
    const evs = surrender(g, 1);
    expect(evs).toHaveLength(1);
    const r = endOf(evs);
    expect([r.over, r.winner, r.won]).toStrictEqual([
      true,
      0,
      [endSurrender, ""],
    ]);
    expect(r.score, "no annihilation bonus").toBe((armour(g) + 100) * 3);
    expect(g.ended).toBe(r);
    expect(surrender(g, 0), "a second surrender").toStrictEqual([]);
    expect(g.ended).toBe(r);
    expect(reload(g).ended, "an ended battle reloads").toStrictEqual(r);
    expect(
      check(g, { kind: cmdMove, side: 0, unit: 0, cell: 0 }),
      "a command after the surrender",
    ).toBeInstanceOf(Refusal);
    confirm(g, 0);
    confirm(g, 1);
    const [, err] = attempt(() => resolve(g));
    expect(errorsIs(err, ErrOver), "a resolution after the surrender").toBe(
      true,
    );
  });
});

describe("TestFinalScore", () => {
  it.each([
    [1000, false, false, 0, 1000],
    [1000, true, false, 0, 1500],
    [1000, false, true, 0, 1100],
    [1000, true, true, 4, 6400],
    [1000, false, false, 8, 3000],
    [1000, false, false, 16, 2000],
    [1000, false, false, 5, 1000],
    [20000, false, false, 4, 32500],
    [32600, false, false, 0, 32500],
  ] as const)(
    "finalScore(%d, %s, %s, %d) = %d",
    (s, annihilation, concealed, limit, want) => {
      expect(finalScore(s, annihilation, concealed, limit)).toBe(want);
    },
  );
});
