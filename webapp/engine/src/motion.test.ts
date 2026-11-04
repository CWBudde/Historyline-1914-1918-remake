// Twin of internal/dosrules/motion_test.go: the recorded movement queries, move orders, container cases and
// phase changes replay through the record-level functions on the profile's movement tables; each class of case
// must be present.

import { describe, expect, it } from "vitest";
import { type ExitRun, moveExit } from "./container.ts";
import { at } from "./goruntime.ts";
import { div, rem } from "./int.ts";
import {
  type MovementFixture,
  containerExpectedOf,
  containerMovement,
  fixtureMovement,
  mapOf,
  movementQuery,
  stepExpectedOf,
  visibilityExpectedOf,
  visibilityPhaseChange,
} from "./motionFixtures.ts";
import { endCode, moveDestinations, moveSeason, path, reach } from "./move.ts";
import type { Profile } from "./profile.ts";
import { execute } from "./route.ts";
import { load, replay } from "./testing/fixtures.ts";
import {
  decodeContainerFixture,
  decodeMovementFixture,
  decodeStepFixture,
  decodeVisibilityFixture,
} from "./testing/motionFixtures.ts";
import { Profiles, moveTables } from "./testing/tables.ts";
import { changePhase } from "./visibility.ts";

// The record counts of each build's files (the full build adds the winter cases).
const movements = { demo: 770, full: 913 } as const;
const steps = { demo: 743, full: 743 } as const;
const containers = { demo: 528, full: 528 } as const;
const winterCases = { demo: 0, full: 143 } as const;
const phaseChanges = { demo: 148, full: 148 } as const;

function replayMovement(p: Profile, f: MovementFixture): void {
  const tab = moveSeason(moveTables(p), f.winter ?? false);
  if (tab === null) {
    expect.fail(`${f.id}: a winter case without a winter table`);
  }
  const g = movementQuery(f);
  const x = f.expected;
  const [buf, bits, count] = reach(tab, g);
  const kept = x.remaining.map((_, cell) =>
    at(buf, div(cell, g.width) * 64 + rem(cell, g.width)),
  );
  expect(kept, `${f.id}: remaining`).toStrictEqual(x.remaining);
  expect(count, `${f.id}: cells reached`).toBe(x.count);
  const codes = x.codes.map(([cell]): [number, number] => [
    cell,
    endCode(tab, g, cell),
  ]);
  expect(codes, `${f.id}: end codes`).toStrictEqual(x.codes);
  if (x.destinations !== null) {
    expect(
      moveDestinations(tab, g, bits),
      `${f.id}: destinations`,
    ).toStrictEqual(x.destinations);
  }
  expect(x.paths, `${f.id}: a route per goal`).toHaveLength(f.goals.length);
  const routes = f.goals.map((goal) => path(tab, g, buf, bits, goal));
  expect(routes, `${f.id}: routes`).toStrictEqual(x.paths);
}

replay(
  "movements.json",
  decodeMovementFixture,
  movements,
  replayMovement,
  "terrain",
  "transition",
  "units",
  "unseen",
  "block",
  "budget",
  "edge",
  "random",
);

// TestWinterFixtures checks that the winter cases are there and that some of them reach differently on the
// summer table, so their replay shows that the season's table is read.
describe("winter fixtures", () => {
  it("some winter cases reach differently in summer", () => {
    let builds = 0;
    for (const [build, p] of Profiles) {
      const summer = moveTables(p);
      const all = load(build, "movements.json", decodeMovementFixture);
      const winter = all.filter((f) => f.winter === true);
      expect(winter, `${build}: winter cases`).toHaveLength(winterCases[build]);
      if (summer.winter === null) {
        continue;
      }
      builds++;
      let differs = 0;
      for (const f of winter) {
        const [buf] = reach(summer, movementQuery(f));
        if (
          f.expected.remaining.some(
            (want, cell) =>
              at(buf, div(cell, f.width) * 64 + rem(cell, f.width)) !== want,
          )
        ) {
          differs++;
        }
      }
      expect(differs, `${build}: winter cases reaching differently`).toBe(38);
    }
    expect(builds, "no profile has a winter table").toBeGreaterThan(0);
  });
});

replay(
  "steps.json",
  decodeStepFixture,
  steps,
  (p, f) => {
    const g = fixtureMovement(f);
    const r = execute(moveTables(p), g, f.order);
    expect(stepExpectedOf(g, r), f.id).toStrictEqual(f.expected);
  },
  "terrain",
  "transition",
  "units",
  "unseen",
  "block",
  "budget",
  "edge",
  "random",
  "order",
  "retarget",
  "pass",
);

replay(
  "containers.json",
  decodeContainerFixture,
  containers,
  (p, f) => {
    const tab = moveTables(p);
    const g = containerMovement(f);
    let r: ExitRun;
    if (f.order.hold < 0) {
      const o = { goal: f.order.goal, retarget: f.order.retarget };
      r = { unit: 0, remaining: new Int8Array(0), ...execute(tab, g, o) };
    } else {
      r = moveExit(tab, g, f.order);
    }
    expect(containerExpectedOf(g, f.order, r), f.id).toStrictEqual(f.expected);
  },
  "load",
  "slots",
  "weight",
  "moved",
  "nested",
  "building",
  "carry",
  "exit",
);

replay(
  "visibility.json",
  decodeVisibilityFixture,
  phaseChanges,
  (_, f) => {
    const g = visibilityPhaseChange(f);
    const r = changePhase(g);
    expect(visibilityExpectedOf(g, r), f.id).toStrictEqual(f.expected);
  },
  "phase",
  "sides",
  "life",
  "stale",
  "ships",
  "terrain",
  "holds",
  "random",
);

// MapOf is the fixture form of a movement query: it gives back the map of every recorded movement case.
describe.each(Profiles)("mapOf, %s build", (build) => {
  it("round-trips the movement fixtures' maps", () => {
    for (const f of load(build, "movements.json", decodeMovementFixture)) {
      const m = mapOf(movementQuery(f));
      expect(m, f.id).toStrictEqual({
        width: f.width,
        height: f.height,
        fill: f.fill,
        cells: f.cells,
        units: f.units,
        unseen: f.unseen,
        unit: f.unit,
        side: f.side,
        budget: f.budget,
        block: f.block,
      });
    }
  });
});
