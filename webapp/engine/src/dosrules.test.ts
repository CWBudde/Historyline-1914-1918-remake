// Twin of internal/dosrules/dosrules_test.go (the parts L1 ports: the fixture headers, TestTables,
// TestCasualtyFixtures, TestDistanceSwitch).

import { describe, expect, it } from "vitest";
import { casualties } from "./casualties.ts";
import { distance } from "./distance.ts";
import { casualtyInputs, casualtyResult, isFixtureHeader } from "./fixtures.ts";
import { attempt } from "./goruntime.ts";
import { type Tables, newProfile, tablesOf } from "./profile.ts";
import { neighbours } from "./support.ts";
import { listDir, readJSON } from "./testing/files.ts";
import {
  decodeCasualtyFixture,
  checkHeader,
  load,
} from "./testing/fixtures.ts";
import {
  Demo,
  DemoTables,
  Full,
  FullTables,
  Profiles,
} from "./testing/tables.ts";

const kinds = [
  "casualties",
  "containers",
  "engagements",
  "movements",
  "orders",
  "steps",
  "visibility",
];

// Every record of each build's fixture files carries the header checkHeader reads: schema 1, the profile
// name, an emulated routine of the tables' build, ids unique per file.
describe.each(Profiles)("dosrules fixture headers, %s build", (build, p) => {
  it("holds exactly the known kinds", () => {
    const files = listDir(`internal/dosrules/testdata/${build}`).filter((f) =>
      f.endsWith(".json"),
    );
    expect(files.map((f) => f.slice(0, -".json".length)).sort()).toEqual(kinds);
  });

  it.each(kinds)("%s records carry the fixture header", (kind) => {
    const records = readJSON(
      `internal/dosrules/testdata/${build}/${kind}.json`,
    );
    expect(Array.isArray(records)).toBe(true);
    const list = records as unknown[];
    expect(list.length).toBeGreaterThan(0);
    const seen = new Set<string>();
    for (const r of list) {
      if (!isFixtureHeader(r)) {
        expect.fail(`not a fixture record: ${JSON.stringify(r).slice(0, 120)}`);
      }
      checkHeader(p, r, seen);
    }
  });
});

// TestTables checks both builds' tables and that they differ only in the full build's winter table: any other
// difference between the builds would be a Profile field.
describe("tables", () => {
  it("both builds", () => {
    for (const [tab, build] of [
      [FullTables, "fe4fa446"],
      [DemoTables, "83fc4e37"],
    ] as const) {
      expect(tab.types).toHaveLength(51);
      expect(tab.build).toBe(build);
      expect(tab.move).not.toBeNull();
      expect(tab.move?.terrains).toHaveLength(175);
      expect(tab.terms).toHaveLength(9);
    }
    expect(FullTables.move?.winter).toHaveLength(175);
    expect(DemoTables.move?.winter).toBeNull();
    const full: Tables = {
      ...FullTables,
      build: "",
      move:
        FullTables.move === null ? null : { ...FullTables.move, winter: null },
    };
    const demo: Tables = { ...DemoTables, build: "" };
    expect(
      full,
      "the demo's tables differ from the full build's",
    ).toStrictEqual(demo);
    expect(Demo.transposedDistance).toBe(Full.transposedDistance);
  });
});

// The recorded casualty-routine runs cover the R1 input classes.
describe.each(Profiles)("casualty fixtures, %s build", (build, p) => {
  it("every record of casualties.json", () => {
    const fs = load(build, "casualties.json", decodeCasualtyFixture);
    expect(fs).toHaveLength(1386);
    const seen = new Set<string>();
    let divide = 0;
    let passed = 0;
    for (const f of fs) {
      checkHeader(p, f, seen);
      const [a, d, c] = casualtyInputs(f, tablesOf(p));
      const got = casualtyResult(...attempt(() => casualties(p, a, d, c)));
      expect(got, f.id).toStrictEqual(f.expected);
      passed++;
      if (f.expected.error !== undefined) {
        divide++;
      }
    }
    expect(passed).toBe(1386);
    expect(divide, "no fixture with a divide error").toBe(61);
  });
});

// TransposedDistance puts one of the six neighbours at distance 2; without it every neighbour is at distance
// 1, and every cell's distance is the number of steps through neighbours.
describe("distance switch", () => {
  it("layout and battle distance", () => {
    const w = 9;
    const h = 9;
    const layout = newProfile({ tables: FullTables });
    for (let cell = 0; cell < w * h; cell++) {
      let far = 0;
      for (const n of neighbours(cell, w, h)) {
        if (n < 0) {
          continue;
        }
        expect(
          distance(layout, w, cell, n),
          `cell ${cell}: neighbour ${n}`,
        ).toBe(1);
        if (distance(Full, w, cell, n) !== 1) {
          far++;
        }
      }
      const x = cell % w;
      const y = Math.trunc(cell / w);
      if (x > 0 && x < w - 1 && y > 0 && y < h - 1) {
        expect(far, `cell ${cell}: neighbours not at battle distance 1`).toBe(
          1,
        );
      }
    }
    // Breadth-first steps from the centre match the layout distance.
    const steps = new Map<number, number>([[40, 0]]);
    const queue = [40];
    for (let i = 0; i < queue.length; i++) {
      const c = queue[i] ?? 0;
      for (const n of neighbours(c, w, h)) {
        if (n >= 0 && !steps.has(n)) {
          steps.set(n, (steps.get(c) ?? 0) + 1);
          queue.push(n);
        }
      }
    }
    expect(steps.size).toBe(w * h);
    for (const [c, s] of steps) {
      expect(distance(layout, w, 40, c), `cell ${c}`).toBe(s);
    }
  });
});
