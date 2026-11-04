// Twin of internal/dosrules/cycle_conflicts_test.go. Go's refusedMoveUnchanged compares the saves before and
// after the refused move; until the save layer is ported, the twin compares the whole game state, which every
// saved field is part of.

import { describe, expect, it } from "vitest";
import {
  type Command,
  type Game,
  cmdMove,
  evMoved,
  evRefused,
  occupant,
  phaseAction,
  phaseMovement,
} from "./cycle.ts";
import { at } from "./goruntime.ts";
import { applyOk, newCycle, resolveOk } from "./testing/scenes.ts";
import { Profiles, cInfantry } from "./testing/tables.ts";

function refusedMoveUnchanged(g: Game, c: Command): void {
  const before = structuredClone(g);
  const evs = applyOk(g, c);
  expect(
    evs.map((e) => e.kind),
    "occupied destination",
  ).toStrictEqual([evRefused]);
  expect(g, "a refused move changed the game state").toStrictEqual(before);
}

const cases = Profiles.flatMap(([build, p]) =>
  [0, 1].map((side) => [`${build}/side${side}`, p, side] as const),
);

// Moves happen immediately on the moving side's map. A second order to the same destination is refused without
// consuming that unit's order.
describe("TestCycleConflictingDestinations", () => {
  it.each(
    cases.flatMap(([name, p, side]) =>
      [0, 1].map((first) => [`${name}/first${first}`, p, side, first] as const),
    ),
  )("%s", (_name, p, side, first) => {
    const g = newCycle(p, {
      0: { type: cInfantry, owner: side, cell: 40 },
      1: { type: cInfantry, owner: side, cell: 42 },
    });
    at(g.players, side).phase = phaseMovement;
    at(g.players, 1 - side).phase = phaseAction;
    const second = 1 - first;
    const evs = applyOk(g, { kind: cmdMove, side, unit: first, cell: 41 });
    expect(evs.map((e) => e.kind)).toStrictEqual([evMoved]);
    expect(occupant(g, side, 41)).toBe(first);
    refusedMoveUnchanged(g, { kind: cmdMove, side, unit: second, cell: 41 });
    expect(at(g.players, side).orders, "orders after refusal").toStrictEqual([
      first,
    ]);
    // The waiting side still sees both units at their old cells.
    expect(
      [
        occupant(g, 1 - side, 40),
        occupant(g, 1 - side, 42),
        occupant(g, 1 - side, 41),
      ],
      "the moves changed the waiting side's map",
    ).toStrictEqual([0, 1, 0xff]);
    resolveOk(g);
    for (let viewer = 0; viewer < 2; viewer++) {
      expect(
        [
          occupant(g, viewer, 41),
          occupant(g, viewer, 40 + 2 * second),
          occupant(g, viewer, 40 + 2 * first),
        ],
        `side ${viewer}: occupants after resolution`,
      ).toStrictEqual([first, second, 0xff]);
    }
  });
});

// Two units cannot exchange occupied cells with queued moves. Once a unit moves elsewhere, its old cell is
// available to a later move in that phase.
describe("TestCycleSwapsAndVacatedCells", () => {
  it.each(cases)("%s", (_name, p, side) => {
    const g = newCycle(p, {
      0: { type: cInfantry, owner: side, cell: 40 },
      1: { type: cInfantry, owner: side, cell: 41 },
    });
    at(g.players, side).phase = phaseMovement;
    at(g.players, 1 - side).phase = phaseAction;
    refusedMoveUnchanged(g, { kind: cmdMove, side, unit: 0, cell: 41 });
    refusedMoveUnchanged(g, { kind: cmdMove, side, unit: 1, cell: 40 });
    for (const c of [
      { kind: cmdMove, side, unit: 0, cell: 31 },
      { kind: cmdMove, side, unit: 1, cell: 40 },
    ] as const) {
      expect(
        applyOk(g, c).map((e) => e.kind),
        "move after refusal",
      ).toStrictEqual([evMoved]);
    }
    expect(at(g.players, side).orders).toStrictEqual([0, 1]);
    resolveOk(g);
    for (let viewer = 0; viewer < 2; viewer++) {
      expect(
        [
          occupant(g, viewer, 31),
          occupant(g, viewer, 40),
          occupant(g, viewer, 41),
        ],
        `side ${viewer}: occupants after resolution`,
      ).toStrictEqual([0, 1, 0xff]);
    }
  });
});
