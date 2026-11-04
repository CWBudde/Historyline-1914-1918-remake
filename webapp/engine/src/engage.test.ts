// Twin of internal/dosrules/engage_test.go.

import { describe, expect, it } from "vitest";
import { classBits } from "./casualties.ts";
import { type Battlefield, engage } from "./engage.ts";
import { at } from "./goruntime.ts";
import { tablesOf } from "./profile.ts";
import { neighbours } from "./support.ts";
import { Profiles, cInfantry, cPanzer } from "./testing/tables.ts";

// Return fire: a defender next to its attacker fires back only when its type can; medium, heavy and train
// artillery (long range), the construction unit and the static balloon (never return fire) take the losses
// without a reply.
describe.each(Profiles)("engage returned, %s build", (_build, p) => {
  const cMedium = 11;
  const cHeavy = 12;
  const cConstruction = 16;
  const cBalloon = 23;
  it.each([
    [cInfantry, true],
    [cPanzer, true],
    [cMedium, false],
    [cHeavy, false],
    [cConstruction, false],
    [cBalloon, false],
  ] as const)("type %d fires back: %s", (defender, returned) => {
    const types = tablesOf(p).types;
    const n = neighbours(40, 9, 9);
    const td = at(types, defender);
    const b: Battlefield = {
      width: 9,
      height: 9,
      units: new Map([
        [
          0,
          {
            type: cInfantry,
            cell: 40,
            size: 6,
            alt: 0,
            exp: 0,
            flags4: 1 | (at(types, cInfantry).targetFlags & classBits),
            flags6: 0,
          },
        ],
        [
          1,
          {
            type: defender,
            cell: at(n, 0),
            size: td.group,
            alt: 0,
            exp: 0,
            flags4: 2 | (td.targetFlags & classBits),
            flags6: 0,
          },
        ],
      ]),
      terrain: new Map(),
      mapByte0: 0,
    };
    const o = engage(p, b, 0, at(n, 0));
    expect(o.returnFire, `infantry on an adjacent type ${defender}`).toBe(true);
    expect(o.returned, `type ${defender} fired back`).toBe(returned);
    if (!returned) {
      expect(
        o.attackerSize,
        `type ${defender} did not fire back, yet the attacker lost strength`,
      ).toBe(6);
    }
  });
});
