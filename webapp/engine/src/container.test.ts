// The port's own checks of container.ts and moveSeason beyond the Go tests: the spelled-out empty record and
// the season's tables leave the build's tables as they are.

import { describe, expect, it } from "vitest";
import {
  containerRecords,
  emptyHold,
  holdRecord,
  holdRecords,
} from "./container.ts";
import { moveSeason } from "./move.ts";
import { Demo, Full, moveTables } from "./testing/tables.ts";

describe("container records", () => {
  it("are emptyHold's bytes where none is given, each its own", () => {
    const want = holdRecord(emptyHold());
    const recs = containerRecords(
      new Map([[3, { ...emptyHold(), flags: 0x41 }]]),
    );
    expect(recs.size).toBe(holdRecords);
    for (const [i, r] of recs) {
      if (i !== 3) {
        expect(r, `record ${i}`).toStrictEqual(want);
      }
    }
    expect(recs.get(3)?.[0x19]).toBe(0x41);
    const r0 = recs.get(0);
    r0?.fill(1);
    expect(recs.get(1)).toStrictEqual(want);
    expect(containerRecords(new Map()).get(0)).toStrictEqual(want);
  });
});

describe("moveSeason", () => {
  it("gives winter terrains without changing the build's tables", () => {
    const t = moveTables(Full);
    const summer = t.terrains;
    expect(moveSeason(t, false)).toBe(t);
    const w = moveSeason(t, true);
    expect(w?.terrains).toBe(t.winter);
    expect(w?.types).toBe(t.types);
    expect(t.terrains).toBe(summer);
    expect(moveSeason(moveTables(Demo), true)).toBeNull();
  });
});
