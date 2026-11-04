// Twin of internal/dosrules/sapper_test.go: sappers dig and fill trenches.

import { describe, expect, it } from "vitest";
import {
  type Event,
  type Game,
  check,
  cmdAttack,
  confirm,
  evBattle,
  evDug,
  evFilled,
  evSkipped,
  evWork,
  liveOf,
  msgNoTarget,
} from "./cycle.ts";
import { at } from "./goruntime.ts";
import { msgTrench, terrainFilled, workCells } from "./sapper.ts";
import { tablesOf } from "./profile.ts";
import { neighbours } from "./support.ts";
import {
  applyOk,
  cycleH,
  cycleW,
  refusal,
  resolveOk,
  sapperScene,
  terrainOf,
  unworked,
} from "./testing/scenes.ts";
import { Profiles, cInfantry } from "./testing/tables.ts";

// work gives sapper 1 its order on cell, resolves the phase and returns the work's event.
function work(g: Game, cell: number): Event {
  const evs = applyOk(g, { kind: cmdAttack, side: 1, unit: 1, cell });
  expect(evs).toHaveLength(1);
  expect(at(evs, 0).kind).toBe(evWork);
  expect(at(at(evs, 0).messages, 0).code).toBe(msgTrench);
  expect(liveOf(g, 1).flags4 & 0x200).not.toBe(0);
  expect(g.players[1].orders).toStrictEqual([]);
  confirm(g, 0);
  confirm(g, 1);
  const done: Event[] = [];
  for (const ev of resolveOk(g)) {
    expect([evBattle, evSkipped], "the work fought").not.toContain(ev.kind);
    if (ev.kind === evDug || ev.kind === evFilled) {
      done.push(ev);
    }
  }
  expect(done).toHaveLength(1);
  return at(done, 0);
}

function expectTerrain(g: Game, want: Readonly<Record<number, number>>): void {
  for (const [c, rec] of Object.entries(want)) {
    expect(terrainOf(g, Number(c)), `cell ${c}`).toBe(rec);
  }
}

describe("TestSapperTargets", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const n = neighbours(40, cycleW, cycleH);
    const g = sapperScene(p, 40, at(n, 0), { [at(n, 3)]: unworked(p) });
    for (const [k, c] of n.entries()) {
      const err = check(g, { kind: cmdAttack, side: 1, unit: 1, cell: c });
      if (k === 3) {
        refusal("terrain the sapper leaves alone", err, msgNoTarget);
      } else {
        expect(err, `neighbour ${k} (${c})`).toBeNull();
      }
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
  });
});

describe("TestSapperDigs", () => {
  it.each(Profiles)("%s", (_build, p) => {
    // Even column: the target straight above, with an enemy on it, bends the trench round the sapper's front.
    let g = sapperScene(p, 40, 31, null);
    let ev = work(g, 31);
    expect([ev.kind, ev.cell, ev.cells]).toStrictEqual([
      evDug,
      31,
      [30, 31, 32],
    ]);
    expectTerrain(g, { 30: 0xaa, 31: 0xa7, 32: 0xab, 40: 0, 22: 0 });
    expect(liveOf(g, 0).size, "the unit on the target lost strength").toBe(
      at(tablesOf(p).types, cInfantry).group,
    );

    // Odd column, the target to the lower right: a straight line.
    g = sapperScene(p, 41, 0, null);
    ev = work(g, 51);
    expect(ev.cells).toStrictEqual([43, 51, 50]);
    expectTerrain(g, { 43: 0xaa, 51: 0xa7, 50: 0xaa });
  });
});

describe("TestSapperFills", () => {
  it.each(Profiles)("%s", (_build, p) => {
    // A trench target is filled; of its line only the cells sappers work change.
    const rock = unworked(p);
    const g = sapperScene(p, 40, 0, { 31: 0xa7, 32: rock });
    const ev = work(g, 31);
    expect([ev.kind, ev.cells]).toStrictEqual([evFilled, [30, 31]]);
    expectTerrain(g, { 30: terrainFilled, 31: terrainFilled, 32: rock });
  });
});

// workCells previews the work: the cells, the records they take and whether it fills, as the work then changes
// them, without changing the maps.
describe("TestSapperWorkCells", () => {
  it.each(Profiles)("%s", (_build, p) => {
    const rock = unworked(p);
    for (const [from, to, terrain] of [
      [40, 31, null],
      [41, 51, null],
      [40, 31, { 31: 0xa7, 32: rock }],
      [40, 31, { 30: 0xaa, 32: rock }],
    ] as const) {
      const g = sapperScene(p, from, 0, terrain);
      const before = [...g.maps[1]];
      const [cells, recs, fill] = workCells(g, 1, 1, to);
      expect(
        g.maps[1],
        `${from}→${to}: the preview changed the map`,
      ).toStrictEqual(before);
      const ev = work(g, to);
      expect(cells, `${from}→${to}`).toStrictEqual(ev.cells);
      expect(fill, `${from}→${to}`).toBe(ev.kind === evFilled);
      expect(recs).toHaveLength(cells.length);
      for (const [k, c] of cells.entries()) {
        expect(terrainOf(g, c), `${from}→${to}: cell ${c}`).toBe(at(recs, k));
      }
    }
    let g = sapperScene(p, 40, 0, null);
    expect(
      workCells(g, 1, 1, 42)[0],
      "a cell beyond the neighbours",
    ).toStrictEqual([]);
    expect(workCells(g, 1, 7, 31)[0], "no such unit").toStrictEqual([]);
    g = sapperScene(p, 40, 0, { 31: rock });
    expect(
      workCells(g, 1, 1, 31)[0],
      "terrain the sapper leaves alone",
    ).toStrictEqual([]);
  });
});
